# FastAX 开发计划 v2 — 全面 PDD 对照审计

> **审计日期**：2026-06-03  
> **审计范围**：PDD v3.0 全部 19 个文件 vs 实际代码（后端 + 前端 + 数据库）  
> **审计维度**：API 端点 × 数据库表 × 中间件 × 安全 × i18n × 适配器 × 前端页面  
> **当前状态**：S0-S7 基本完成，API 覆盖率 100%，前端页面接入率 97%

---

## 1. 审计总览

| 维度 | 检查项 | 完全合规 | 部分合规 | 缺失 | 合规率 |
|------|--------|---------|---------|------|--------|
| API 端点 | 118 | 118 | 0 | 0 | **100%** |
| 数据库表 | 28 张 | 13 张 | 10 张 | 5 处缺字段 | 46% 表完全合规 |
| 数据库索引 | 29 个 | 14 个 | 5 个 | 10 个缺失/错误 | 48% |
| 中间件 | 5 个 | 1 (cors) | 3 (auth,ratelimit,lang) | 1 (distributor) | 20% 完全合规 |
| Relay 模式 | 8 项 | 4 | 2 | 2 (billing,distributor) | 50% |
| 适配器 | 6 个 | 3 | 0 | 3 (DeepSeek/Qwen/GLM) | 50% |
| i18n | 5 项 | 1 (model) | 1 (name_i18n) | 3 (翻译文件/错误消息/回退链) | 20% |
| 安全 | 6 项 | 0 | 2 (ratelimit,oauth) | 4 (bcrypt/jwt/aes/masking) | 0% |
| 前端架构 | 11 项 | 10 | 1 (CSS Modules) | 0 | 91% |
| 前端页面 | 34 个 | 30 个 | 0 | 4 个 | 88% |
| 前端 i18n | 13 命名空间 | 1 | 0 | 12 | 8% |
| 前端组件/Store/Hook | 18 项 | 5 | 0 | 13 | 28% |

---

## 2. 新里程碑规划

### S8：数据层修复（2-4 人日）🔴 CRITICAL

> 目标：修复数据库表结构偏差，补齐缺失字段和索引

#### Task S8.1 — 补齐缺失字段（1-1.5d）

| 表 | 缺失字段 | 说明 |
|---|---------|------|
| `token_product` | `VendorID uint` | FK→supplier_vendor，供应商入驻必需 |
| `order` | `VendorID uint` | FK→supplier_vendor，供应商订单追踪 |
| `call_log` | `VendorID uint` | FK→supplier_vendor，供应商调用日志 |
| `call_log` | `RouteDecision string` | 路由决策记录 (ROUTE-11) |
| `supplier_product` | `MinPrice string` | 平台最低限价 (R-BIZ-50) |
| `supplier_product` | `MaxPrice string` | 平台最高限价 (R-BIZ-50) |

**文件**：`internal/shared/model/token.go`, `order.go`, `call_log.go`, `vendor.go`

#### Task S8.2 — 补齐缺失/错误索引（1-1.5d）

| 表 | 索引 | 说明 |
|---|------|------|
| `user` | `idx_status_created` | 复合索引 (status, created_at) |
| `user` | `idx_preferred_language` | 单列索引 |
| `supplier` | `idx_status_priority` | 复合索引 (status, priority) |
| `token_product` | `idx_type_model` | 复合索引 (type, model) |
| `token_product` | `idx_status_sort` | 复合索引 (status, sort_order) |
| `user_token` | `idx_expires_status` | 复合索引 (expires_at, status) |
| `order` | `idx_created` | 单列索引 |
| `notification` | `idx_user_created` | 重命名为 idx_notification_created 对齐 |
| `supplier_vendor` | `idx_created` | 单列索引 |
| `settlement` | `idx_period` | 复合索引 (period_start, period_end) |
| `byok_key` | `idx_byok_user` | 修正为复合 (user_id, status) |
| `provider_health` | `idx_health_provider` | 修正为复合 (provider_id, period_start) |
| `model_variant` | `idx_variant` (UNIQUE) | 唯一复合索引 (base_model, suffix, provider_id) |

**文件**：`internal/shared/model/user.go`, `channel.go`, `token.go`, `order.go`, `notify.go`, `vendor.go`, `byok.go`, `i18n.go`, `health.go`

#### Task S8.3 — 修正类型偏差（0.5d）

| 表 | 字段 | 当前类型 | PDD 要求 | 修正 |
|---|------|---------|---------|------|
| `sub_account` | `TokenQuota` | `int64` | TEXT | 改为 `string` |
| `token_inventory` | `AlertThreshold` | `float64` | TEXT(DEFAULT 10.00) | 改为 `string` |

**文件**：`internal/shared/model/user.go`, `token.go`

---

### S9：安全加固（2-3 人日）🔴 CRITICAL

#### Task S9.1 — BCrypt 加密强度修正（0.5d）

- 所有 `bcrypt.DefaultCost`(10) → `bcrypt.DefaultCost`(12)
- 文件：`internal/domain/user/service.go`

#### Task S9.2 — JWT 算法升级 HS256→RS256（1d）

- 生成 RSA 2048 密钥对
- 修改 `GenerateAccessToken`/验证逻辑使用 RS256
- 文件：`internal/shared/middleware/auth.go`, `internal/shared/config/config.go`

#### Task S9.3 — 数据脱敏工具（0.5d）

- 新增 `internal/shared/mask/mask.go`
- 实现：手机号(中间4位)、邮箱(@前部分)、身份证(首尾各4位)、银行卡(后4位)、API Key(首尾各4字符)

#### Task S9.4 — AES-256-GCM 服务端加密封装（0.5d）

- 新增 `internal/shared/crypto/aes.go`
- 提供 `Encrypt(plaintext, key)` / `Decrypt(ciphertext, key)` 
- 用于 BYOK Key、身份证号等敏感字段

---

### S10：Proxy 核心增强（3-5 人日）🟡 HIGH

#### Task S10.1 — Distributor 中间件（1-1.5d）

参考 `ref/one-api/middleware/distributor.go`：
- 从 JWT 提取 user group
- 调用 `CacheGetRandomSatisfiedChannel` 预选渠道
- 注入 channel 到 gin.Context
- 文件：`internal/shared/middleware/distributor.go`

#### Task S10.2 — 计费管线（预扣+后扣）（1.5-2d）

- `PreConsume(userID, model)` — 估算 Token 用量并冻结额度
- `PostConsume(userID, actualTokens)` — 按实际用量多退少补
- `BatchFlush()` — 批量刷入 SQLite（每 10s 或 100 条）
- 集成到 `proxy/service.go` 的 `Relay()` 方法中
- 文件：`internal/domain/proxy/service.go`

#### Task S10.3 — 国内适配器补齐（1-1.5d）

参考 `ref/one-api/relay/adaptor/`：
- `relay/adaptor/deepseek/` — DeepSeek（OpenAI 兼容，可直接复用 OpenAIAdaptor + 特定 header）
- `relay/adaptor/qwen/` — 通义千问（messages↔input/history 转换）
- `relay/adaptor/glm/` — 智谱 GLM（prompt↔messages 转换）

---

### S11：i18n 国际化补全（2-3 人日）🟡 HIGH

#### Task S11.1 — 后端翻译文件（1d）

- 创建 `internal/shared/i18n/locales/zh-CN.json`, `en.json`, `ja.json`
- 补齐 13 个命名空间：common, home, auth, token, order, payment, profile, admin, notification, error, docs, vendor, compliance
- 实现 `GetTranslations(locale)` 返回完整翻译 JSON

#### Task S11.2 — 多语言错误消息（0.5-1d）

- 修改 `response.Error()` 支持语言参数
- 为 13 个错误码提供中/英/日翻译
- 中间件注入的 `Accept-Language` 用于选择错误消息语言

#### Task S11.3 — 前端 i18n 补全（1d）

- 创建 `web/src/shared/i18n/locales/zh-CN/` 和 `en/` 目录
- 补齐 12 个缺失命名空间的翻译文件
- 替换 AdminLayout 硬编码 labelMap 为真实 i18n 翻译
- 修复 UserLayout 硬编码字符串

---

### S12：前端补全（2-3 人日）🟡 HIGH

#### Task S12.1 — 缺失页面开发（1-1.5d）

| 页面 | 路由 | 后端 API |
|------|------|---------|
| Admin 供应商管理 (平台自有) | `/admin/suppliers` | 已有 `GET/POST/PUT /api/admin/suppliers` |
| Admin 报表 | `/admin/reports` | 已有 `GET /api/admin/reports/daily`, `/monthly` |
| Admin 管理员管理 | `/admin/system/admins` | 已有 `GET/POST /api/admin/system/admins` |

#### Task S12.2 — 共享组件库（1-1.5d）

PDD 要求但缺失的组件：
- `shared/components/LanguageSelector.tsx`
- `shared/components/DataTable.tsx` (Ant Design Table 封装)
- `shared/components/StatusTag.tsx` (状态标签)

#### Task S12.3 — 缺失 Store/Hook/Util（0.5d）

- `shared/stores/tokenStore.ts` — Token 余额缓存
- `shared/stores/notificationStore.ts` — 未读计数
- `shared/hooks/usePagination.ts` — 分页封装
- `shared/utils/format.ts` — Intl 日期/货币格式化

---

### S13：体验优化与测试（5-8 人日）🟢 MEDIUM

#### Task S13.1 — 统一加载/错误/空状态（1-2d）

- 全站 Loading → `<Spin>` 组件标准化
- 全站 Error → 统一错误提示+重试按钮
- 全站 Empty → Ant Design `<Empty>` + 引导文案

#### Task S13.2 — 前端组件测试（3-4d）

- 关键页面 snapshot 测试
- API service mock 测试

#### Task S13.3 — CI/CD 流水线（1-2d）

- Dockerfile (Go 多阶段构建 + Nginx 静态文件)
- GitHub Actions: lint → test → build → docker push

---

## 3. 工时汇总

| 阶段 | 优先级 | 内容 | 人日 |
|------|--------|------|------|
| **S8** | 🔴 CRITICAL | 数据层修复（字段+索引+类型） | 2-4d |
| **S9** | 🔴 CRITICAL | 安全加固（bcrypt+jwt+脱敏+加密） | 2-3d |
| **S10** | 🟡 HIGH | Proxy 增强（distributor+计费+适配器） | 3-5d |
| **S11** | 🟡 HIGH | i18n 补全（翻译文件+错误消息+前端） | 2-3d |
| **S12** | 🟡 HIGH | 前端补全（页面+组件+Store/Hook） | 2-3d |
| **S13** | 🟢 MEDIUM | 体验优化+测试+CI/CD | 5-8d |
| | | **总计** | **16-26d** |

---

## 4. 依赖关系

```
S8 (数据层) ──── 无依赖，优先执行
   │
   ▼
S9 (安全) ────── 依赖 S8 模型修正
   │
   ▼
S10 (Proxy) ──── 依赖 S8 模型修正
   │
   ├── S11 (i18n) ─── 可并行 S12
   │
   └── S12 (前端) ─── 可并行 S11
          │
          ▼
       S13 (测试/CI/CD)
```

**S8 和 S9 为阻塞项，必须先完成。S10-S12 可并行执行。**

---

## 5. 与 v1 计划的差异

| 项目 | v1 计划 (80-117d) | v2 计划 (16-26d) |
|------|-------------------|-------------------|
| 阶段 | 从零构建 7 个里程碑 | 基于已完成代码的修补 |
| 新增 API | 118 个 | 0 个（API 已 100%） |
| 重点 | 功能实现 | 质量加固 + 补齐遗漏 |
| 数据库 | 28 张表新建 | 6 个字段 + 13 个索引修正 |
| 安全 | 基础实现 | 加密强度/算法/脱敏合规 |
| 前端 | 38 页从零开发 | 4 页缺失 + 组件库 |

---

## 6. 实施原则

1. **S8 优先** — 数据层修正是所有后续工作的基础，必须先完成
2. **安全不可妥协** — bcrypt cost=12、AES-256-GCM 必须合规
3. **参考代码复用** — 国内适配器直接从 `ref/one-api/relay/adaptor/` 移植
4. **保持全绿** — 每个 task 结束后 `go test ./...` 必须通过
5. **增量提交** — 每个 task 独立 commit
