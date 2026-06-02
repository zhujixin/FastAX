import api from "@/shared/utils/axios";
import type { ApiResponse, PaginatedData, UserInfo, Order, RiskEvent, RiskRule, GuardrailRule, VendorInfo, DashboardSummary, DashboardCharts, LanguageInfo } from "./types";

export const adminService = {
  // Dashboard
  getDashboardSummary: () =>
    api.get<ApiResponse<DashboardSummary>>("/admin/dashboard/summary"),

  getDashboardCharts: (period?: string) =>
    api.get<ApiResponse<DashboardCharts>>("/admin/dashboard/charts", { params: { period } }),

  // Users
  listUsers: (params?: { page?: number; size?: number }) =>
    api.get<ApiResponse<PaginatedData<UserInfo>>>("/admin/users", { params }),

  getUser: (id: number) =>
    api.get<ApiResponse<UserInfo>>(`/admin/users/${id}`),

  setUserStatus: (id: number, status: string) =>
    api.put<ApiResponse<null>>(`/admin/users/${id}/status`, { status }),

  setUserLevel: (id: number, level: string) =>
    api.put<ApiResponse<null>>(`/admin/users/${id}/level`, { level }),

  // Orders
  listOrders: (params?: { page?: number; size?: number }) =>
    api.get<ApiResponse<PaginatedData<Order>>>("/admin/orders", { params }),

  getOrder: (id: string) =>
    api.get<ApiResponse<Order>>(`/admin/orders/${id}`),

  adminRefund: (id: string) =>
    api.post<ApiResponse<null>>(`/admin/orders/${id}/refund`),

  // Reports
  getDailyReport: (date?: string) =>
    api.get<ApiResponse<Record<string, unknown>>>("/admin/reports/daily", { params: { date } }),

  getMonthlyReport: (month?: string) =>
    api.get<ApiResponse<Record<string, unknown>>>("/admin/reports/monthly", { params: { month } }),

  // Risk
  listRiskEvents: (params?: object) =>
    api.get<ApiResponse<PaginatedData<RiskEvent>>>("/admin/risk/events", { params }),

  handleRiskEvent: (id: number, action: string) =>
    api.put<ApiResponse<null>>(`/admin/risk/events/${id}/handle`, { action }),

  listRiskRules: () =>
    api.get<ApiResponse<RiskRule[]>>("/admin/risk/rules"),

  createRiskRule: (data: Partial<RiskRule>) =>
    api.post<ApiResponse<RiskRule>>("/admin/risk/rules", data),

  setRuleEnabled: (id: number, enabled: boolean) =>
    api.put<ApiResponse<null>>(`/admin/risk/rules/${id}/enabled`, { enabled }),

  // Blacklist
  listBlacklist: () =>
    api.get<ApiResponse<Array<{ id: number; ip: string; reason: string }>>>("/admin/risk/blacklist"),

  addBlacklist: (data: { ip: string; reason: string }) =>
    api.post<ApiResponse<null>>("/admin/risk/blacklist", data),

  removeBlacklist: (id: number) =>
    api.delete<ApiResponse<null>>(`/admin/risk/blacklist/${id}`),

  // Vendors
  listVendors: () =>
    api.get<ApiResponse<VendorInfo[]>>("/admin/vendors"),

  getVendor: (id: number) =>
    api.get<ApiResponse<VendorInfo>>(`/admin/vendors/${id}`),

  reviewVendor: (id: number, approved: boolean, reason?: string) =>
    api.post<ApiResponse<null>>(`/admin/vendors/${id}/review`, { approved, reason }),

  suspendVendor: (id: number) =>
    api.post<ApiResponse<null>>(`/admin/vendors/${id}/suspend`),

  // Guardrails
  listGuardrailRules: () =>
    api.get<ApiResponse<GuardrailRule[]>>("/admin/guardrails/rules"),

  createGuardrailRule: (data: Partial<GuardrailRule>) =>
    api.post<ApiResponse<GuardrailRule>>("/admin/guardrails/rules", data),

  updateGuardrailRule: (id: number, data: Partial<GuardrailRule>) =>
    api.put<ApiResponse<GuardrailRule>>(`/admin/guardrails/rules/${id}`, data),

  deleteGuardrailRule: (id: number) =>
    api.delete<ApiResponse<null>>(`/admin/guardrails/rules/${id}`),

  setGuardrailRuleEnabled: (id: number, enabled: boolean) =>
    api.put<ApiResponse<null>>(`/admin/guardrails/rules/${id}/enabled`, { enabled }),

  listGuardrailLogs: (params?: { trace_id?: string; stage?: string; user_id?: number }) =>
    api.get<ApiResponse<Array<Record<string, unknown>>>>("/admin/guardrails/logs", { params }),

  // Product management
  createProduct: (data: Record<string, unknown>) =>
    api.post<ApiResponse<Record<string, unknown>>>("/admin/products", data),

  updateProduct: (id: number, data: Record<string, unknown>) =>
    api.put<ApiResponse<null>>(`/admin/products/${id}`, data),

  // BYOK management
  listBYOKKeys: (params?: object) =>
    api.get<ApiResponse<PaginatedData<Record<string, unknown>>>>("/admin/byok/keys", { params }),

  // SSO
  getSSOConfig: () =>
    api.get<ApiResponse<Record<string, unknown>>>("/admin/sso/config"),

  updateSSOConfig: (data: Record<string, unknown>) =>
    api.put<ApiResponse<null>>("/admin/sso/config", data),

  // Teams
  listTeams: () =>
    api.get<ApiResponse<Array<Record<string, unknown>>>>("/admin/teams"),

  createTeam: (data: Record<string, unknown>) =>
    api.post<ApiResponse<Record<string, unknown>>>("/admin/teams", data),

  updateTeam: (id: number, data: Record<string, unknown>) =>
    api.put<ApiResponse<null>>(`/admin/teams/${id}`, data),

  deleteTeam: (id: number) =>
    api.delete<ApiResponse<null>>(`/admin/teams/${id}`),

  // i18n
  listAllLanguages: () =>
    api.get<ApiResponse<LanguageInfo[]>>("/admin/i18n/languages"),

  createLanguage: (data: LanguageInfo) =>
    api.post<ApiResponse<null>>("/admin/i18n/languages", data),

  updateLanguage: (locale: string, data: Partial<LanguageInfo>) =>
    api.put<ApiResponse<null>>(`/admin/i18n/languages/${locale}`, data),

  setDefaultLanguage: (locale: string) =>
    api.put<ApiResponse<null>>("/admin/i18n/default", { locale }),

  // Audit
  listAuditLogs: (params?: object) =>
    api.get<ApiResponse<PaginatedData<Record<string, unknown>>>>("/admin/audit/logs", { params }),

  exportAuditLogs: () =>
    api.get("/admin/audit/export", { responseType: "blob" }),
};
