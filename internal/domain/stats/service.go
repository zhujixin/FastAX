// 统计模块业务逻辑层
//
// 本文件实现了统计相关的业务逻辑：
//
// 用户统计：
//   - GetUsage: 用量统计（Token 消耗、请求数）
//   - GetConsumption: 消费统计（金额、订单数）
//   - GetBills: 账单明细（分页）
//   - GetSummary: 控制台总览（余额、用量、趋势）
//
// 管理员统计：
//   - GetDashboardSummary: 管理后台数据总览
//   - GetDailyReport: 日报表
//   - GetMonthlyReport: 月报表
package stats

import (
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// Service 统计服务结构体
type Service struct {
	db *gorm.DB
}

// NewService 创建统计服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------- 请求/响应类型 ----------

// UsageResponse 用量统计响应
type UsageResponse struct {
	TotalTokens     int    `json:"total_tokens"`
	PromptTokens    int    `json:"prompt_tokens"`
	CompletionTokens int   `json:"completion_tokens"`
	RequestCount    int    `json:"request_count"`
	Period          string `json:"period"`
}

// ConsumptionResponse 消费统计响应
type ConsumptionResponse struct {
	TotalAmount   string `json:"total_amount"`
	OrderCount    int    `json:"order_count"`
	PaymentCount  int    `json:"payment_count"`
	Period        string `json:"period"`
}

type BillResponse struct {
	ID        uint       `json:"id"`
	OrderID   uint       `json:"order_id"`
	PaymentNo string     `json:"payment_no"`
	Amount    string     `json:"amount"`
	Method    string     `json:"method"`
	Gateway   string     `json:"gateway"`
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type SummaryResponse struct {
	TotalTokens       int    `json:"total_tokens"`
	TotalAmount       string `json:"total_amount"`
	MonthTokens       int    `json:"month_tokens"`
	MonthAmount       string `json:"month_amount"`
	TotalRequests     int    `json:"total_requests"`
	TotalOrders       int    `json:"total_orders"`
}

// ---------- Service methods ----------

// GetUsage aggregates token usage from call_log for the given period.
// period: "day", "week", "month", "year" (default "month")
func (s *Service) GetUsage(userID uint, period string) (*UsageResponse, error) {
	start, err := periodStartTime(period)
	if err != nil {
		return nil, err
	}

	var result struct {
		TotalTokens      int
		PromptTokens     int
		CompletionTokens int
		RequestCount     int
	}

	err = s.db.Model(&model.CallLog{}).
		Select("COALESCE(SUM(tokens_total), 0) as total_tokens, COALESCE(SUM(tokens_prompt), 0) as prompt_tokens, COALESCE(SUM(tokens_completion), 0) as completion_tokens, COUNT(*) as request_count").
		Where("user_id = ? AND created_at >= ?", userID, start).
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("aggregate usage: %w", err)
	}

	return &UsageResponse{
		TotalTokens:      result.TotalTokens,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
		RequestCount:     result.RequestCount,
		Period:           period,
	}, nil
}

// GetConsumption aggregates payment amount from paid orders for the given period.
func (s *Service) GetConsumption(userID uint, period string) (*ConsumptionResponse, error) {
	start, err := periodStartTime(period)
	if err != nil {
		return nil, err
	}

	// Count paid/completed orders
	var orderCount int64
	if err := s.db.Model(&model.Order{}).
		Where("user_id = ? AND status IN ? AND created_at >= ?", userID, []string{"paid", "completed"}, start).
		Count(&orderCount).Error; err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	// Sum payment amounts from successful payments linked to user's orders
	var totalAmount string
	var paymentCount int64

	type payResult struct {
		TotalAmount string
		Count       int64
	}
	var pr payResult
	err = s.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(CAST(payments.amount AS REAL)), 0) as total_amount, COUNT(*) as count").
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.user_id = ? AND payments.status = ? AND payments.created_at >= ?", userID, "success", start).
		Scan(&pr).Error
	if err != nil {
		return nil, fmt.Errorf("aggregate payments: %w", err)
	}

	totalAmount = fmt.Sprintf("%.2f", parseFloat(pr.TotalAmount))
	paymentCount = pr.Count

	return &ConsumptionResponse{
		TotalAmount:  totalAmount,
		OrderCount:   int(orderCount),
		PaymentCount: int(paymentCount),
		Period:       period,
	}, nil
}

// GetBills returns paginated payment records for the user.
func (s *Service) GetBills(userID uint, page, pageSize int) ([]BillResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query: payments linked to user's orders
	baseQuery := s.db.Model(&model.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.user_id = ?", userID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count bills: %w", err)
	}

	var payments []model.Payment
	if err := baseQuery.
		Select("payments.*").
		Order("payments.created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("query bills: %w", err)
	}

	bills := make([]BillResponse, len(payments))
	for i, p := range payments {
		bills[i] = BillResponse{
			ID:        p.ID,
			OrderID:   p.OrderID,
			PaymentNo: p.PaymentNo,
			Amount:    p.Amount,
			Method:    p.Method,
			Gateway:   p.Gateway,
			Status:    p.Status,
			PaidAt:    p.PaidAt,
			CreatedAt: p.CreatedAt,
		}
	}

	return bills, total, nil
}

// GetSummary returns an overview: total and current-month usage & consumption.
func (s *Service) GetSummary(userID uint) (*SummaryResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Total usage
	var totalUsage struct {
		TotalTokens  int
		RequestCount int
	}
	if err := s.db.Model(&model.CallLog{}).
		Select("COALESCE(SUM(tokens_total), 0) as total_tokens, COUNT(*) as request_count").
		Where("user_id = ?", userID).
		Scan(&totalUsage).Error; err != nil {
		return nil, fmt.Errorf("total usage: %w", err)
	}

	// Current month usage
	var monthUsage struct {
		TotalTokens  int
		RequestCount int
	}
	if err := s.db.Model(&model.CallLog{}).
		Select("COALESCE(SUM(tokens_total), 0) as total_tokens, COUNT(*) as request_count").
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Scan(&monthUsage).Error; err != nil {
		return nil, fmt.Errorf("month usage: %w", err)
	}

	// Total consumption
	type amountResult struct {
		TotalAmount string
		Count       int64
	}

	var totalPay amountResult
	if err := s.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(CAST(payments.amount AS REAL)), 0) as total_amount, COUNT(*) as count").
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.user_id = ? AND payments.status = ?", userID, "success").
		Scan(&totalPay).Error; err != nil {
		return nil, fmt.Errorf("total consumption: %w", err)
	}

	// Current month consumption
	var monthPay amountResult
	if err := s.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(CAST(payments.amount AS REAL)), 0) as total_amount, COUNT(*) as count").
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.user_id = ? AND payments.status = ? AND payments.created_at >= ?", userID, "success", monthStart).
		Scan(&monthPay).Error; err != nil {
		return nil, fmt.Errorf("month consumption: %w", err)
	}

	return &SummaryResponse{
		TotalTokens:   totalUsage.TotalTokens,
		TotalAmount:   fmt.Sprintf("%.2f", parseFloat(totalPay.TotalAmount)),
		MonthTokens:   monthUsage.TotalTokens,
		MonthAmount:   fmt.Sprintf("%.2f", parseFloat(monthPay.TotalAmount)),
		TotalRequests: totalUsage.RequestCount,
		TotalOrders:   int(totalPay.Count),
	}, nil
}

// ---------- Admin Dashboard ----------

type DashboardSummary struct {
	TotalUsers      int    `json:"total_users"`
	TodayNewUsers   int    `json:"today_new_users"`
	TotalOrders     int    `json:"total_orders"`
	TodayNewOrders  int    `json:"today_new_orders"`
	TotalRevenue    string `json:"total_revenue"`
	TodayRevenue    string `json:"today_revenue"`
	ActiveTokens    int    `json:"active_tokens"`
}

// GetDashboardSummary returns platform-wide metrics for admin dashboard.
func (s *Service) GetDashboardSummary() (*DashboardSummary, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Total users
	var totalUsers int64
	if err := s.db.Model(&model.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	// Today new users
	var todayUsers int64
	if err := s.db.Model(&model.User{}).Where("created_at >= ?", todayStart).Count(&todayUsers).Error; err != nil {
		return nil, fmt.Errorf("count today users: %w", err)
	}

	// Total orders
	var totalOrders int64
	if err := s.db.Model(&model.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	// Today new orders
	var todayOrders int64
	if err := s.db.Model(&model.Order{}).Where("created_at >= ?", todayStart).Count(&todayOrders).Error; err != nil {
		return nil, fmt.Errorf("count today orders: %w", err)
	}

	// Total revenue (from successful payments)
	var totalRevenue struct {
		Total string
	}
	if err := s.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(CAST(amount AS REAL)), 0) as total").
		Where("status = ?", "success").
		Scan(&totalRevenue).Error; err != nil {
		return nil, fmt.Errorf("sum revenue: %w", err)
	}

	// Today revenue
	var todayRevenue struct {
		Total string
	}
	if err := s.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(CAST(amount AS REAL)), 0) as total").
		Where("status = ? AND created_at >= ?", "success", todayStart).
		Scan(&todayRevenue).Error; err != nil {
		return nil, fmt.Errorf("sum today revenue: %w", err)
	}

	// Active tokens (user_tokens with status=1)
	var activeTokens int64
	if err := s.db.Model(&model.UserToken{}).Where("status = ?", 1).Count(&activeTokens).Error; err != nil {
		return nil, fmt.Errorf("count active tokens: %w", err)
	}

	return &DashboardSummary{
		TotalUsers:     int(totalUsers),
		TodayNewUsers:  int(todayUsers),
		TotalOrders:    int(totalOrders),
		TodayNewOrders: int(todayOrders),
		TotalRevenue:   fmt.Sprintf("%.2f", parseFloat(totalRevenue.Total)),
		TodayRevenue:   fmt.Sprintf("%.2f", parseFloat(todayRevenue.Total)),
		ActiveTokens:   int(activeTokens),
	}, nil
}

// ---------- helpers ----------

func periodStartTime(period string) (time.Time, error) {
	now := time.Now()
	switch period {
	case "day":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()), nil
	case "month", "":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "year":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), nil
	default:
		return time.Time{}, fmt.Errorf("invalid period: %s (use day/week/month/year)", period)
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// ---------- Dashboard Charts ----------

// ChartDataPoint 单个数据点
type ChartDataPoint struct {
	Date   string `json:"date"`
	Value  int64  `json:"value"`
	Amount string `json:"amount,omitempty"`
}

// DashboardCharts 管理后台图表数据
type DashboardCharts struct {
	Revenue []ChartDataPoint `json:"revenue"`
	Users   []ChartDataPoint `json:"users"`
	Orders  []ChartDataPoint `json:"orders"`
	Tokens  []ChartDataPoint `json:"tokens"`
}

// GetDashboardCharts returns chart time-series data for the admin dashboard.
// period: "7d" (default) or "30d"
func (s *Service) GetDashboardCharts(period string) (*DashboardCharts, error) {
	days := 7
	if period == "30d" {
		days = 30
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	charts := &DashboardCharts{
		Revenue: make([]ChartDataPoint, 0, days),
		Users:   make([]ChartDataPoint, 0, days),
		Orders:  make([]ChartDataPoint, 0, days),
		Tokens:  make([]ChartDataPoint, 0, days),
	}

	for d := startDate; d.Before(now) || d.Equal(now); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayStart := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		// Revenue
		var revenue float64
		s.db.Model(&model.Payment{}).
			Where("status = ? AND created_at >= ? AND created_at < ?", "success", dayStart, dayEnd).
			Select("COALESCE(SUM(CAST(amount AS REAL)), 0)").
			Row().Scan(&revenue)
		charts.Revenue = append(charts.Revenue, ChartDataPoint{
			Date:   dateStr,
			Amount: fmt.Sprintf("%.2f", revenue),
		})

		// New users
		var newUsers int64
		s.db.Model(&model.User{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&newUsers)
		charts.Users = append(charts.Users, ChartDataPoint{
			Date:  dateStr,
			Value: newUsers,
		})

		// New orders
		var newOrders int64
		s.db.Model(&model.Order{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&newOrders)
		charts.Orders = append(charts.Orders, ChartDataPoint{
			Date:  dateStr,
			Value: newOrders,
		})

		// Tokens used
		var tokensUsed int64
		s.db.Model(&model.CallLog{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Select("COALESCE(SUM(tokens_total), 0)").
			Row().Scan(&tokensUsed)
		charts.Tokens = append(charts.Tokens, ChartDataPoint{
			Date:  dateStr,
			Value: tokensUsed,
		})
	}

	return charts, nil
}

// ---------- Reports ----------

type DailyReport struct {
	Date          string `json:"date"`
	NewUsers      int64  `json:"new_users"`
	NewOrders     int64  `json:"new_orders"`
	Revenue       string `json:"revenue"`
	TokensUsed    int64  `json:"tokens_used"`
	ActiveUsers   int64  `json:"active_users"`
}

type MonthlyReport struct {
	Year          int            `json:"year"`
	Month         int            `json:"month"`
	TotalUsers    int64          `json:"total_users"`
	NewUsers      int64          `json:"new_users"`
	TotalOrders   int64          `json:"total_orders"`
	TotalRevenue  string         `json:"total_revenue"`
	TokensUsed    int64          `json:"tokens_used"`
	DailyBreakdown []DailyReport `json:"daily_breakdown"`
}

// GetDailyReport returns daily statistics for a given date.
func (s *Service) GetDailyReport(date string) (*DailyReport, error) {
	// Parse date
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %s, use YYYY-MM-DD", date)
	}

	dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	report := &DailyReport{Date: date}

	// New users
	s.db.Model(&model.User{}).
		Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
		Count(&report.NewUsers)

	// New orders
	s.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
		Count(&report.NewOrders)

	// Revenue
	var revenue float64
	s.db.Model(&model.Payment{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "success", dayStart, dayEnd).
		Select("COALESCE(SUM(CAST(amount AS REAL)), 0)").
		Row().Scan(&revenue)
	report.Revenue = fmt.Sprintf("%.2f", revenue)

	// Tokens used
	s.db.Model(&model.CallLog{}).
		Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
		Select("COALESCE(SUM(tokens_total), 0)").
		Row().Scan(&report.TokensUsed)

	// Active users (users who made API calls)
	s.db.Model(&model.CallLog{}).
		Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
		Select("COUNT(DISTINCT user_id)").
		Row().Scan(&report.ActiveUsers)

	return report, nil
}

// GetMonthlyReport returns monthly statistics with daily breakdown.
func (s *Service) GetMonthlyReport(year, month int) (*MonthlyReport, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d", month)
	}

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0)

	report := &MonthlyReport{
		Year:  year,
		Month: month,
	}

	// Total users at end of month
	s.db.Model(&model.User{}).
		Where("created_at < ?", monthEnd).
		Count(&report.TotalUsers)

	// New users this month
	s.db.Model(&model.User{}).
		Where("created_at >= ? AND created_at < ?", monthStart, monthEnd).
		Count(&report.NewUsers)

	// Total orders this month
	s.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", monthStart, monthEnd).
		Count(&report.TotalOrders)

	// Total revenue this month
	var revenue float64
	s.db.Model(&model.Payment{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "success", monthStart, monthEnd).
		Select("COALESCE(SUM(CAST(amount AS REAL)), 0)").
		Row().Scan(&revenue)
	report.TotalRevenue = fmt.Sprintf("%.2f", revenue)

	// Tokens used this month
	s.db.Model(&model.CallLog{}).
		Where("created_at >= ? AND created_at < ?", monthStart, monthEnd).
		Select("COALESCE(SUM(tokens_total), 0)").
		Row().Scan(&report.TokensUsed)

	// Daily breakdown
	report.DailyBreakdown = make([]DailyReport, 0)
	for d := monthStart; d.Before(monthEnd); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		daily, err := s.GetDailyReport(dateStr)
		if err != nil {
			continue
		}
		report.DailyBreakdown = append(report.DailyBreakdown, *daily)
	}

	return report, nil
}
