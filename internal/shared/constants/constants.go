// 共享常量定义
//
// 本文件定义全局共享的命名常量，消除代码库中的魔术字符串和魔术数字：
//   - 角色常量 (RoleAdmin, RoleUser, ...)
//   - 用户等级常量 (LevelNormal, LevelVIP, ...)
//   - 订单/支付状态常量 (OrderStatusPending, ...)
//   - 分页默认值
//   - 通用业务常量
package constants

// ─── 用户角色 ───

const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
	RoleEnterprise = "enterprise"
)

// ValidRoles returns true if the given role string is a known role.
func ValidRoles() []string {
	return []string{RoleUser, RoleAdmin, RoleSuperAdmin, RoleEnterprise}
}

// ─── 用户等级 ───

const (
	LevelNormal     = "normal"
	LevelVIP        = "vip"
	LevelEnterprise = "enterprise"
)

// ─── 订单状态 ───

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunding = "refunding"
	OrderStatusRefunded  = "refunded"
)

// ─── 支付状态 ───

const (
	PaymentStatusPending  = "pending"
	PaymentStatusSuccess  = "success"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// ─── 供应商/产品状态 ───

const (
	StatusEnabled  = 1
	StatusDisabled = 0
	StatusPending  = "pending_review"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// ─── 风控等级 ───

const (
	RiskLevelL1 = "L1" // 低风险
	RiskLevelL2 = "L2" // 中低风险
	RiskLevelL3 = "L3" // 中高风险
	RiskLevelL4 = "L4" // 高风险
)

// ─── 清算状态 ───

const (
	SettlementPending   = "pending"
	SettlementSettled   = "settled"
	SettlementConfirmed = "confirmed"
)

// ─── 分页默认值 ───

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ─── 通用业务常量 ───

const (
	DefaultCurrency = "CNY"
	DefaultLocale   = "zh-CN"
)
