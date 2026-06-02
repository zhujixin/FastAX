// 统计模块 HTTP 处理器
//
// 本文件实现了统计相关的 HTTP 接口：
//
// 用户统计：
//   - GetUsage: 用量统计（按日/周/月）
//   - GetConsumption: 消费统计
//   - GetBills: 账单明细
//   - GetSummary: 控制台总览
//
// 管理员统计：
//   - GetDashboardSummary: 管理后台数据总览
//   - GetDailyReport: 日报表
//   - GetMonthlyReport: 月报表
package stats

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 统计模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建统计处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetUsage 获取用量统计
// GET /api/stats/usage?period=month
func (h *Handler) GetUsage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	resp, err := h.svc.GetUsage(userID.(uint), period)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

// GET /api/stats/consumption?period=month
func (h *Handler) GetConsumption(c *gin.Context) {
	userID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	resp, err := h.svc.GetConsumption(userID.(uint), period)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

// GET /api/stats/bills?page=1&page_size=20
func (h *Handler) GetBills(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, pageSize := 1, 20
	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	items, total, err := h.svc.GetBills(userID.(uint), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, int64(total), page, pageSize)
}

// GET /api/stats/summary
func (h *Handler) GetSummary(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.svc.GetSummary(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, resp)
}

// GET /api/admin/dashboard/summary
func (h *Handler) GetDashboardSummary(c *gin.Context) {
	resp, err := h.svc.GetDashboardSummary()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, resp)
}

// GET /api/admin/dashboard/charts?period=7d
func (h *Handler) GetDashboardCharts(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")
	charts, err := h.svc.GetDashboardCharts(period)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, charts)
}

// GET /api/admin/reports/daily?date=2026-05-31
func (h *Handler) GetDailyReport(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	report, err := h.svc.GetDailyReport(date)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, report)
}

// GET /api/admin/reports/monthly?year=2026&month=5
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		fmt.Sscanf(y, "%d", &year)
	}
	if m := c.Query("month"); m != "" {
		fmt.Sscanf(m, "%d", &month)
	}

	report, err := h.svc.GetMonthlyReport(year, month)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, report)
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
