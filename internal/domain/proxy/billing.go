// 计费管线
//
// 本文件实现 Token 代理的计费逻辑：
//   - PreConsume: 转发前估算 Token 用量并冻结额度
//   - PostConsume: 转发后按实际用量扣减（多退少补）
//   - BatchFlush: 定时批量刷入 SQLite
//
// 计费流程：
//   请求进入 → PreConsume(冻结估算量) → 转发 → PostConsume(解冻+实际扣减) → 返回
package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/domain/proxy/relay"
	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// BillingManager 计费管理器
type BillingManager struct {
	db          *gorm.DB
	mu          sync.Mutex
	userLocks   map[uint]*sync.Mutex // per-user mutex to serialize billing ops
	preConsumed map[uint]string      // userID → pre-consumed amount (frozen)
	batch       []usageRecord
	batchMu     sync.Mutex
}

type usageRecord struct {
	UserID       uint
	ProductID    uint
	Amount       string // actual tokens used
	PreAmount    string // pre-consumed estimate
	Model        string
	SupplierID   uint
}

// NewBillingManager 创建计费管理器
// PRAGMA settings (WAL mode, busy_timeout) are configured globally in model.InitDB.
func NewBillingManager(db *gorm.DB) *BillingManager {
	bm := &BillingManager{
		db:          db,
		userLocks:   make(map[uint]*sync.Mutex),
		preConsumed: make(map[uint]string),
		batch:       make([]usageRecord, 0, DefaultBillingBatchSize),
	}
	// 启动批量刷入协程：每 10 秒或累积 100 条
	go bm.batchFlushLoop(DefaultBillingFlushInterval)
	return bm
}

// getUserLock returns a per-user mutex to serialize billing operations.
func (bm *BillingManager) getUserLock(userID uint) *sync.Mutex {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if mu, ok := bm.userLocks[userID]; ok {
		return mu
	}
	mu := &sync.Mutex{}
	bm.userLocks[userID] = mu
	return mu
}

// PreConsume 预扣：估算 Token 用量，冻结用户额度
// 使用 per-user mutex 序列化同一用户的并发计费操作，
// 替代 SQLite 的 FOR UPDATE（在 WAL 模式下不可靠）。
// 返回预扣量（字符串），失败时返回错误但不应阻断转发（记日志即可）
func (bm *BillingManager) PreConsume(userID uint, modelName string, messages []relay.Message) (string, error) {
	// Serialize billing ops per user to prevent concurrent balance corruption
	userMu := bm.getUserLock(userID)
	userMu.Lock()
	defer userMu.Unlock()

	// 估算 Token 用量：每消息 50 tokens + 内容长度/4
	estimated := TokensPerMessage * len(messages)
	for _, m := range messages {
		estimated += len(m.Content) / 4
	}
	if estimated < MinTokenEstimate {
		estimated = MinTokenEstimate
	}

	amount := fmt.Sprintf("%d", estimated)

	// 查找用户持有的 Token 记录
	var userTokens []model.UserToken
	if err := bm.db.Where("user_id = ? AND status = 1 AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Find(&userTokens).Error; err != nil {
		return amount, fmt.Errorf("query user tokens: %w", err)
	}

	if len(userTokens) == 0 {
		return amount, nil // 无 Token 记录，跳过扣减（免费/试用模式）
	}

	// Re-read the first token record inside the per-user lock
	var locked model.UserToken
	if err := bm.db.Where("id = ? AND status = 1", userTokens[0].ID).First(&locked).Error; err != nil {
		return amount, fmt.Errorf("read token record: %w", err)
	}

	remaining := parseAmount(locked.TotalAmount) - parseAmount(locked.UsedAmount) - parseAmount(locked.FrozenAmount)
	if remaining < float64(estimated) {
		return amount, fmt.Errorf("insufficient token balance: need %d, have %.0f", estimated, remaining)
	}

	bm.mu.Lock()
	bm.preConsumed[userID] = amount
	bm.mu.Unlock()

	// 冻结额度（per-user lock 保证串行，无需 FOR UPDATE）
	newFrozen := parseAmount(locked.FrozenAmount) + float64(estimated)
	bm.db.Model(&locked).Update("frozen_amount", fmt.Sprintf("%.0f", newFrozen))

	return amount, nil
}

// PostConsume 后扣：按实际用量扣减（解冻估算量，扣减实际量）
// Uses the same per-user mutex as PreConsume to prevent interleaving.
func (bm *BillingManager) PostConsume(userID uint, preAmount string, actualTokens int) error {
	userMu := bm.getUserLock(userID)
	userMu.Lock()
	defer userMu.Unlock()

	bm.mu.Lock()
	delete(bm.preConsumed, userID)
	bm.mu.Unlock()

	var userTokens []model.UserToken
	if err := bm.db.Where("user_id = ? AND status = 1 AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Find(&userTokens).Error; err != nil {
		return fmt.Errorf("query user tokens: %w", err)
	}

	if len(userTokens) == 0 {
		return nil
	}

	t := &userTokens[0]
	// 解冻预扣量
	preAmt := parseAmount(preAmount)
	currentFrozen := parseAmount(t.FrozenAmount)
	if currentFrozen >= preAmt {
		currentFrozen -= preAmt
	}
	// 增加实际用量
	currentUsed := parseAmount(t.UsedAmount) + float64(actualTokens)

	updates := map[string]interface{}{
		"frozen_amount": fmt.Sprintf("%.0f", currentFrozen),
		"used_amount":   fmt.Sprintf("%.0f", currentUsed),
	}
	return bm.db.Model(t).Updates(updates).Error
}

// RecordUsage 记录用量到批量缓冲区
func (bm *BillingManager) RecordUsage(userID, productID, supplierID uint, modelName string, preAmount string, actualTokens int) {
	bm.batchMu.Lock()
	defer bm.batchMu.Unlock()

	bm.batch = append(bm.batch, usageRecord{
		UserID:     userID,
		ProductID:  productID,
		Amount:     fmt.Sprintf("%d", actualTokens),
		PreAmount:  preAmount,
		Model:      modelName,
		SupplierID: supplierID,
	})

	// 达到 100 条立即刷入
	if len(bm.batch) >= DefaultBillingBatchSize {
		bm.flushBatch()
	}
}

func (bm *BillingManager) batchFlushLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		bm.batchMu.Lock()
		if len(bm.batch) > 0 {
			bm.flushBatch()
		}
		bm.batchMu.Unlock()
	}
}

func (bm *BillingManager) flushBatch() {
	for _, r := range bm.batch {
		bm.db.Create(&model.CallLog{
			UserID:     r.UserID,
			ProductID:  r.ProductID,
			SupplierID: r.SupplierID,
			RequestModel: r.Model,
			TokensTotal:  parseAmountAsInt(r.Amount),
			Status:     "success",
			CreatedAt:  time.Now(),
		})
	}
	bm.batch = bm.batch[:0]
}

// Flush 手动强制刷入（用于服务关闭时）
func (bm *BillingManager) Flush() {
	bm.batchMu.Lock()
	defer bm.batchMu.Unlock()
	bm.flushBatch()
}

func parseAmount(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func parseAmountAsInt(s string) int {
	var v int
	fmt.Sscanf(s, "%d", &v)
	return v
}
