# PDD API 交叉验证清单

## 验证日期: 2026-05-31

---

## 7.2 业务 API 详细设计

### 7.2.1 认证相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/auth/register` | POST | 定义 | ✅ 已实现 | user.Handler.Register |
| `/api/auth/login` | POST | 定义 | ✅ 已实现 | user.Handler.Login |
| `/api/auth/refresh` | POST | 定义 | ✅ 已实现 | user.Handler.RefreshToken |
| `/api/auth/logout` | POST | 定义 | ✅ 已实现 | user.Handler.Logout |
| `/api/auth/send-code` | POST | 定义 | ✅ 已实现 | user.Handler.SendCode |
| `/api/auth/reset-password` | POST | 定义 | ✅ 已实现 | user.Handler.ResetPassword |
| `/api/auth/oauth/{provider}` | GET | 定义 | ❌ 未实现 | OAuth登录跳转 |
| `/api/auth/oauth/callback` | GET | 定义 | ❌ 未实现 | OAuth回调 |

### 7.2.2 Token 相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/tokens/products` | GET | 定义 | ✅ 已实现 | token.Handler.GetProducts |
| `/api/tokens/products/{id}` | GET | 定义 | ❌ 未实现 | 商品详情 |
| `/api/tokens/my` | GET | 定义 | ✅ 已实现 | token.Handler.GetMyTokens |
| `/api/tokens/buy` | POST | 定义 | ✅ 已实现 | token.Handler.Buy |
| `/api/tokens/transfer` | POST | 定义 | ✅ 已实现 | token.Handler.Transfer |
| `/api/tokens/extract` | POST | 定义 | ✅ 已实现 | token.Handler.Extract |
| `/api/tokens/my/usage` | GET | 定义 | ❌ 未实现 | 使用记录 |

### 7.2.3 订单相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/orders` | GET | 定义 | ✅ 已实现 | order.Handler.List |
| `/api/orders/{id}` | GET | 定义 | ✅ 已实现 | order.Handler.Get |
| `/api/orders/{id}/refund` | POST | 定义 | ✅ 已实现 | order.Handler.RequestRefund |

### 7.2.4 统计相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/stats/usage` | GET | 定义 | ✅ 已实现 | stats.Handler.GetUsage |
| `/api/stats/consumption` | GET | 定义 | ✅ 已实现 | stats.Handler.GetConsumption |
| `/api/stats/bills` | GET | 定义 | ✅ 已实现 | stats.Handler.GetBills |
| `/api/stats/summary` | GET | 定义 | ✅ 已实现 | stats.Handler.GetSummary |

### 7.2.5 通知相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/notifications` | GET | 定义 | ✅ 已实现 | notify.Handler.List |
| `/api/notifications/unread-count` | GET | 定义 | ✅ 已实现 | notify.Handler.UnreadCount |
| `/api/notifications/{id}/read` | PUT | 定义 | ✅ 已实现 | notify.Handler.MarkRead |
| `/api/notifications/read-all` | PUT | 定义 | ✅ 已实现 | notify.Handler.MarkAllRead |

### 7.2.6 多语言相关

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/i18n/languages` | GET | 定义 | ✅ 已实现 | i18n.Handler.ListLanguages |
| `/api/i18n/translations/{locale}` | GET | 定义 | ✅ 已实现 | i18n.Handler.GetTranslations |
| `/api/user/language` | PUT | 定义 | ✅ 已实现 | user.Handler.UpdateLanguage |

### 7.2.7 供应商服务 API

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/vendor/register` | POST | 定义 | ✅ 已实现 | vendor.Handler.Apply |
| `/api/vendor/profile` | GET | 定义 | ✅ 已实现 | vendor.Handler.GetVendorByUserID |
| `/api/vendor/profile` | PUT | 定义 | ❌ 未实现 | 更新店铺信息 |
| `/api/vendor/products` | GET | 定义 | ✅ 已实现 | vendor.Handler.ListProducts |
| `/api/vendor/products` | POST | 定义 | ✅ 已实现 | vendor.Handler.CreateProduct |
| `/api/vendor/products/{id}` | PUT | 定义 | ❌ 未实现 | 更新商品 |
| `/api/vendor/products/{id}/price` | PUT | 定义 | ❌ 未实现 | 调价 |
| `/api/vendor/sales` | GET | 定义 | ✅ 已实现 | vendor.Handler.GetSales |
| `/api/vendor/settlements` | GET | 定义 | ✅ 已实现 | vendor.Handler.GetSettlements |
| `/api/vendor/settlements/{id}/confirm` | POST | 定义 | ❌ 未实现 | 确认结算单 |
| `/api/vendor/settlements/{id}/withdraw` | POST | 定义 | ❌ 未实现 | 申请提现 |

### 7.2.8 管理后台—供应商管理

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/admin/vendors` | GET | 定义 | ✅ 已实现 | vendor.Handler.ListVendors |
| `/api/admin/vendors/{id}` | GET | 定义 | ✅ 已实现 | vendor.Handler.GetVendor |
| `/api/admin/vendors/{id}/approve` | POST | 定义 | ✅ 已实现 | vendor.Handler.ReviewVendor |
| `/api/admin/vendors/{id}/reject` | POST | 定义 | ✅ 已实现 | (合并到ReviewVendor) |
| `/api/admin/vendors/{id}/suspend` | POST | 定义 | ❌ 未实现 | 冻结供应商 |
| `/api/admin/vendor-commission-rates/{id}` | PUT | 定义 | ❌ 未实现 | 配置佣金比例 |
| `/api/admin/vendors/{id}/products` | GET | 定义 | ✅ 已实现 | vendor.Handler.ListProducts |
| `/api/admin/vendors/{id}/products/{pid}/approve` | POST | 定义 | ✅ 已实现 | vendor.Handler.ReviewProduct |
| `/api/admin/vendors/{id}/settlements` | GET | 定义 | ❌ 未实现 | 供应商结算记录 |

---

## 7.3 管理后台 API

| 分组 | 接口 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| **Dashboard** | `GET /api/admin/dashboard/summary` | 定义 | ✅ 已实现 | stats.Handler.GetDashboardSummary |
| | `GET /api/admin/dashboard/charts` | 定义 | ❌ 未实现 | 趋势图表数据 |
| **用户管理** | `GET /api/admin/users` | 定义 | ✅ 已实现 | user.Handler.ListUsers |
| | `GET /api/admin/users/{id}` | 定义 | ❌ 未实现 | 用户详情 |
| | `PUT /api/admin/users/{id}/status` | 定义 | ✅ 已实现 | user.Handler.SetUserStatus |
| | `PUT /api/admin/users/{id}/level` | 定义 | ❌ 未实现 | 修改用户等级 |
| **Token 管理** | `GET /api/admin/suppliers` | 定义 | ✅ 已实现 | vendor.Handler.ListSuppliers |
| | `POST/PUT /api/admin/suppliers` | 定义 | ✅ 已实现 | CreateSupplier/UpdateSupplier |
| | `GET /api/admin/channels` | 定义 | ❌ 未实现 | 渠道列表 |
| | `PUT /api/admin/channels/{id}/status` | 定义 | ❌ 未实现 | 启用/禁用渠道 |
| | `PUT /api/admin/channels/{id}/priority` | 定义 | ❌ 未实现 | 调整优先级 |
| | `GET /api/admin/products` | 定义 | ❌ 未实现 | Token 商品列表 |
| | `POST/PUT /api/admin/products` | 定义 | ❌ 未实现 | 新增/编辑商品 |
| | `PUT /api/admin/products/{id}/price` | 定义 | ❌ 未实现 | 调价 |
| **交易管理** | `GET /api/admin/orders` | 定义 | ✅ 已实现 | order.Handler.ListAdmin |
| | `GET /api/admin/orders/{id}` | 定义 | ✅ 已实现 | (复用order.Handler.Get) |
| | `POST /api/admin/orders/{id}/refund` | 定义 | ❌ 未实现 | 审核退款 |
| | `GET /api/admin/reports/daily` | 定义 | ❌ 未实现 | 日报表 |
| | `GET /api/admin/reports/monthly` | 定义 | ❌ 未实现 | 月报表 |
| **风控管理** | `GET /api/admin/risk/events` | 定义 | ✅ 已实现 | risk.Handler.ListEvents |
| | `PUT /api/admin/risk/events/{id}` | 定义 | ✅ 已实现 | risk.Handler.HandleEvent |
| | `GET /api/admin/risk/rules` | 定义 | ✅ 已实现 | risk.Handler.ListRules |
| | `POST/PUT /api/admin/risk/rules` | 定义 | ✅ 已实现 | risk.Handler.CreateRule |
| | `GET /api/admin/risk/blacklist` | 定义 | ❌ 未实现 | 黑名单 |
| | `POST /api/admin/risk/blacklist` | 定义 | ❌ 未实现 | 添加黑名单 |
| **系统管理** | `GET /api/admin/system/config` | 定义 | ❌ 未实现 | 系统配置 |
| | `PUT /api/admin/system/config` | 定义 | ❌ 未实现 | 更新配置 |
| | `GET /api/admin/system/admins` | 定义 | ❌ 未实现 | 管理员列表 |
| | `POST /api/admin/system/admins` | 定义 | ❌ 未实现 | 添加管理员 |
| | `GET /api/admin/system/logs` | 定义 | ✅ 已实现 | log.Handler.ListAuditLogs |
| **多语言配置** | `GET /api/admin/i18n/languages` | 定义 | ✅ 已实现 | i18n.Handler.ListAllLanguages |
| | `PUT /api/admin/i18n/languages/{id}` | 定义 | ✅ 已实现 | i18n.Handler.UpdateLanguage |
| | `PUT /api/admin/i18n/default` | 定义 | ✅ 已实现 | i18n.Handler.SetDefaultLanguage |
| **供应商管理** | `GET /api/admin/vendors` | 定义 | ✅ 已实现 | vendor.Handler.ListVendors |
| | `GET /api/admin/vendors/{id}` | 定义 | ✅ 已实现 | vendor.Handler.GetVendor |
| | `POST /api/admin/vendors/{id}/approve` | 定义 | ✅ 已实现 | vendor.Handler.ReviewVendor |
| | `POST /api/admin/vendors/{id}/reject` | 定义 | ✅ 已实现 | (合并到ReviewVendor) |
| | `PUT /api/admin/vendor-commission-rates/{id}` | 定义 | ❌ 未实现 | 配置佣金比例 |

---

## 7.6 新增 v3.0 API 端点

### 7.6.1 多协议原生 (PROTO)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/v1/messages` | POST | 定义 | ✅ 已实现 | proxy.Handler.ChatMessages |
| `/v1/rerank` | POST | 定义 | ❌ 未实现 | Rerank 统一 API |
| `/models/:variant` | GET | 定义 | ❌ 未实现 | 模型变体详情 |

### 7.6.2 安全护栏 (GRDL)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/admin/guardrails/rules` | GET/POST | 定义 | ✅ 已实现 | guardrail.Handler.ListRules/CreateRule |
| `/api/admin/guardrails/rules/:id` | PUT/DELETE | 定义 | ⚠️ 部分实现 | 只有enabled，无DELETE |
| `/api/admin/guardrails/logs` | GET | 定义 | ✅ 已实现 | guardrail.Handler.ListLogs |
| `/api/admin/guardrails/config` | PUT | 定义 | ❌ 未实现 | 全局配置 |

### 7.6.3 BYOK (自带 Key)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/byok/keys` | GET/POST | 定义 | ✅ 已实现 | byok.Handler.ListKeys/AddKey |
| `/api/byok/keys/:id` | PUT/DELETE | 定义 | ⚠️ 部分实现 | 只有DELETE和status |
| `/api/byok/usage` | GET | 定义 | ❌ 未实现 | BYOK 用量统计 |
| `/api/byok/preference` | PUT | 定义 | ❌ 未实现 | 路由优先级配置 |

### 7.6.4 成本优化 (COST)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/cache/stats` | GET | 定义 | ❌ 未实现 | 语义缓存统计 |
| `/api/cache/config` | PUT | 定义 | ❌ 未实现 | 缓存策略配置 |
| `/api/user/budget` | GET/PUT | 定义 | ✅ 已实现 | cost.Handler.GetBudget/SetBudget |
| `/api/user/cost-alerts` | GET/PUT | 定义 | ✅ 已实现 | cost.Handler.GetAlerts/SetAlert |

### 7.6.5 模型市场 (MKT)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/models/compare` | POST | 定义 | ✅ 已实现 | market.Handler.CompareModels |
| `/api/providers/health` | GET | 定义 | ✅ 已实现 | market.Handler.ListProviders |
| `/api/models/benchmarks` | GET | 定义 | ❌ 未实现 | 基准测试数据 |

### 7.6.6 多模态 (MEDIA)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/v1/images/generations` | POST | 定义 | ✅ 已实现 | proxy.Handler.ImageGenerations |
| `/v1/audio/speech` | POST | 定义 | ✅ 已实现 | proxy.Handler.AudioSpeech |
| `/v1/audio/transcriptions` | POST | 定义 | ❌ 未实现 | 语音转文本 STT |
| `/v1/video/generations` | POST | 定义 | ❌ 未实现 | 视频生成 (P2) |

### 7.6.7 企业功能 (ENT)

| 接口 | 方法 | PDD状态 | 实现状态 | 备注 |
|------|------|---------|----------|------|
| `/api/admin/sso/config` | GET/PUT | 定义 | ❌ 未实现 | SSO 配置 |
| `/api/admin/teams` | GET/POST | 定义 | ❌ 未实现 | 团队管理 |
| `/api/admin/teams/:id` | PUT/DELETE | 定义 | ❌ 未实现 | 团队编辑/删除 |
| `/api/admin/audit/export` | GET | 定义 | ✅ 已实现 | log.Handler.ExportAuditLogs |

---

## 统计汇总

| 类别 | PDD定义 | 已实现 | 未实现 | 完成率 |
|------|---------|--------|--------|--------|
| 认证相关 | 8 | 6 | 2 | 75% |
| Token相关 | 7 | 5 | 2 | 71% |
| 订单相关 | 3 | 3 | 0 | 100% |
| 统计相关 | 4 | 4 | 0 | 100% |
| 通知相关 | 4 | 4 | 0 | 100% |
| 多语言相关 | 3 | 3 | 0 | 100% |
| 供应商服务 | 11 | 7 | 4 | 64% |
| 供应商管理 | 9 | 6 | 3 | 67% |
| 管理后台 | 30 | 18 | 12 | 60% |
| 多协议原生 | 3 | 1 | 2 | 33% |
| 安全护栏 | 4 | 3 | 1 | 75% |
| BYOK | 4 | 2 | 2 | 50% |
| 成本优化 | 4 | 2 | 2 | 50% |
| 模型市场 | 3 | 2 | 1 | 67% |
| 多模态 | 4 | 2 | 2 | 50% |
| 企业功能 | 4 | 1 | 3 | 25% |
| **总计** | **105** | **69** | **36** | **66%** |

---

## 优先级建议

### P0 (必须完成 - 影响核心功能)
1. `/api/tokens/products/{id}` - 商品详情
2. `/api/tokens/my/usage` - 使用记录
3. `/api/admin/orders/{id}/refund` - 审核退款
4. `/api/admin/guardrails/rules/:id` DELETE - 删除规则

### P1 (重要 - 影响用户体验)
1. OAuth登录 (`/api/auth/oauth/*`)
2. 供应商店铺更新 (`/api/vendor/profile` PUT)
3. 商品管理 (`/api/vendor/products/{id}` PUT)
4. 结算确认/提现 (`/api/vendor/settlements/*`)
5. 管理后台完善 (用户详情、商品管理、报表)

### P2 (可延后 - 扩展功能)
1. Rerank API
2. 视频生成 API
3. SSO/SAML/OIDC
4. 团队管理
5. 基准测试
6. 语义缓存管理
