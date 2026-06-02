# FastAX 项目整体评估与最新开发计划

> 评估日期: 2026-06-01 (第三轮更新) | 基于 PRD v3.0 + PDD v3.0
> 原始估算: 80-117 人日, 7 里程碑 | 最新提交: `26f2932` feat: complete P2 advanced features

---

## 一、总体完成度

```
项目整体完成度: █████████░░ 82%

后端:  ██████████ 93%  (API 86/105 PDD 规范, 135 实际注册, 测试 506+ 全绿)
前端:  ████████░░ 76%  (25/33 页面接入 API, 33 页面全部存在)
文档:  ██████████ 95%  (PRD 23 文件 + PDD 21 文件, 完整)
```

| 维度 | 已完成 | 总量 | 完成率 |
|------|--------|------|--------|
| 后端 API (PDD 规范) | ~86 | ~105 | 82% |
| 后端 API (router 实际注册) | **135** | — | 超出 PDD 规范 |
| 后端测试函数 | 506+ | 506+ | 100% (25 包全绿) |
| 后端 Domain 全部实现 | 18 | 18 | **100%** |
| 前端页面(接入 API) | 25 | 33 | 76% |
| 前端页面(总数含硬编码) | 33 | 33 | 100% |
| 前端测试 | 0 | — | 0% |
| PRD 文档 | 23 | 23 | 100% |
| PDD 文档 | 21 | 21 | 100% |

---

## 二、后端评估

### 2.1 各 Domain 实现状态

| 阶段 | Domain | 状态 | 测试数 | Go文件 | 代码行数 | 说明 |
|------|--------|------|--------|--------|---------|------|
| S0 | shared/config | ✅ | 5 | — | — | 完整 |
| S0 | shared/model | ✅ | 2 | — | — | 完整, 25+ 表 |
| S0 | shared/cache | ✅ | 10 | — | — | 完整 |
| S0 | shared/middleware | ✅ | 26 | — | — | JWT/CORS/限流/语言 |
| S1 | domain/user | ✅ | 35 | 5 | 1644 | 含 OAuth/SSO/子账号 |
| S1 | domain/token | ✅ | 29 | 3 | 1625 | 含 usage history |
| S2 | domain/order | ✅ | 27 | 4 | 1248 | 含退款/管理员操作 |
| S2 | domain/payment | ✅ | 17 | 3 | 898 | 微信/Stripe |
| S3 | domain/proxy | ✅ | 77 | 10 | 4384 | ⭐ 核心模块, Rerank/Video/Image/Audio |
| S3 | domain/vendor | ✅ | 19 | 3 | 1637 | 含入驻/审核/结算 |
| S4 | domain/risk | ✅ | 22 | 4 | 1028 | 含黑名单 |
| S4 | domain/notify | ✅ | 31 | 4 | 1048 | 含模板管理 |
| S4 | domain/stats | ✅ | 26 | 5 | 1609 | 含 dashboard charts |
| S4 | domain/commission | ✅ | 25 | 4 | 822 | 完整 |
| S4 | domain/log | ✅ | 16 | 4 | 817 | 审计+调用日志 |
| S5 | domain/guardrail | ✅ | 22 | 4 | 895 | CRUD + 检测 + 启用/禁用 |
| S5 | domain/byok | ✅ | 23 | 4 | 763 | Key 管理完整 |
| S6 | domain/cost | ✅ | 27 | 4 | 924 | 预算+告警 (缺 cache stats/api) |
| S6 | domain/enterprise | ✅ | 24 | 4 | 1207 | SSO + Teams + 子账号 |
| S6 | domain/market | ✅ | 12 | 3 | 835 | 含 benchmarks + recommend |
| S6 | domain/plugin | ⚠️ | 14 | 2 | 434 | service 层有代码, router 未注册 |

**后端代码总计**: ~23,195 行 Go 代码 (17 domains + shared/)

### 2.2 后端 API 缺口 (~19 个，较上轮减少 3 个)

#### ✅ 已补全的缺口 (本轮新增，5+2=7 个)

| API | 所属 | 说明 | 提交 |
|-----|------|------|------|
| `GET /api/admin/dashboard/charts` | admin | 7/30天趋势图 | `add9df7` |
| `GET /api/admin/orders/:id` | admin | 管理端订单详情 | `ee27c86` |
| `PUT /api/admin/guardrails/rules/:id` | guardrail | 更新护栏规则 | `501d5c8` |
| `DELETE /api/admin/guardrails/rules/:id` | guardrail | 删除护栏规则 | `501d5c8` |
| `PUT /api/admin/vendor-products/:id/price` | vendor | 管理员调价 | `e2b10a7` |
| `POST /v1/rerank` | proxy | Rerank 统一 API | `26f2932` |
| `POST /v1/video/generations` | proxy | 视频生成 | `26f2932` |
| `GET /api/models/benchmarks` | market | 基准测试数据 | `26f2932` |
| `GET /api/models/recommend` | market | 模型推荐 | `26f2932` |
| `GET /api/auth/oauth/:provider` | user | OAuth 跳转 | `3782348` |
| `POST /api/auth/oauth/login` | user | OAuth 登录 | `3782348` |
| `GET /api/admin/sso/config` | enterprise | SSO 配置查询 | `3782348` |
| `PUT /api/admin/sso/config` | enterprise | SSO 配置更新 | `3782348` |
| `GET/POST/PUT/DELETE /api/admin/teams` | enterprise | 团队 CRUD | `3782348` |
| `POST /api/enterprise/sub-accounts` 等 | enterprise | 子账号管理 | `3782348` |

#### P1 缺口 (6 个)

| API | 所属 | 说明 |
|-----|------|------|
| `GET /api/vendor/products` | vendor | 供应商自服务商品列表 |
| `POST /api/vendor/products` | vendor | 供应商自服务创建商品 |
| `GET /api/admin/channels` | admin | 渠道列表 (目前用 suppliers 代替) |
| `PUT /api/admin/channels/:id/status` | admin | 渠道启用/禁用 |
| `PUT /api/admin/channels/:id/priority` | admin | 渠道优先级调整 |
| `PUT /api/admin/guardrails/config` | guardrail | 护栏全局配置 |

#### P2 缺口 (13 个)

| API | 所属 | 说明 |
|-----|------|------|
| `GET /api/admin/system/config` | admin | 系统配置查询 |
| `PUT /api/admin/system/config` | admin | 系统配置更新 |
| `GET /api/admin/system/admins` | admin | 管理员列表 |
| `POST /api/admin/system/admins` | admin | 添加管理员 |
| `GET /api/byok/usage` | byok | BYOK 用量统计 |
| `PUT /api/byok/preference` | byok | BYOK 路由偏好 |
| `GET /api/cache/stats` | cost | 语义缓存统计 |
| `PUT /api/cache/config` | cost | 缓存策略配置 |
| `GET /models/:variant` | proxy | 模型变体详情 |
| `POST /v1/audio/transcriptions` | proxy | STT 语音转文本 |
| `PUT /api/admin/vendor-commission-rates/:id` | admin | 佣金比例配置 |
| `GET /api/admin/vendors/:id/settlements` | admin | 供应商结算记录 |
| `plugin/*` (全模块) | plugin | 插件系统完整 API + router 注册 |

### 2.3 测试评估

| 维度 | 数据 |
|------|------|
| 测试文件 | 42 个 |
| 测试函数 | 506 个 |
| 测试结果 | **全部通过** |
| 缺 handler 测试 | market, payment, plugin, token, vendor |
| 覆盖率报告 | 未配置 |
| CI/CD | 无 |

---

## 三、前端评估

### 3.1 页面状态总览

```
前端页面总数: 33 个
  ├── ✅ 已接入 API:  25 个 (76%) — 本轮从 11→25, +14 页面
  ├── ⚠️ 硬编码占位:   8 个 (24%)
  └── ❌ 完全缺失:     0 个 Domain 全部有页面
```

### 3.2 已接入 API 的页面 (25 个)

#### 用户端 (10 个)
| 页面 | 路由 | 覆盖 Domain |
|------|------|------------|
| 登录页 | `/login` | user |
| 注册页 | `/register` | user |
| 忘记密码 | `/forgot-password` | user |
| 首页产品展示 | `/` | token/market |
| 用户 Dashboard | `/dashboard` | stats |
| 我的 Token | `/tokens` | token |
| 购买 Token | `/tokens/buy` | token/order/payment |
| 订单列表 | `/orders` | order |
| 通知中心 | `/notifications` | notify |
| 账单页面 | `/bills` | stats/token |

#### 管理端 (12 个)
| 页面 | 路由 | 覆盖 Domain |
|------|------|------------|
| Admin Dashboard | `/admin` | stats |
| 用户管理 | `/admin/users` | user |
| Token 管理 | `/admin/tokens` | token/proxy |
| 订单管理 | `/admin/orders` | order |
| 风控事件 | `/admin/risk/events` | risk |
| 风控规则 | `/admin/risk/rules` | risk |
| 黑名单 | `/admin/risk/blacklist` | risk |
| 供应商管理 | `/admin/vendors` | vendor |
| 审计日志 | `/admin/audit` | log |
| i18n 配置 | `/admin/i18n` | i18n |
| 护栏规则管理 | `/admin/guardrails/rules` | guardrail |
| 护栏日志 | `/admin/guardrails/logs` | guardrail |

#### 新增页面 (本轮新建 7 个，3 个)
| 页面 | 路由 | 覆盖 Domain | 状态 |
|------|------|------------|------|
| BYOK Key 管理 | `/byok/keys` | byok | ✅ 已接入 |
| 模型市场 | `/market` | market | ✅ 已接入 |
| SSO + 团队管理 | `/admin/enterprise` | enterprise | ✅ 已接入 |
| 预算 + 缓存配置 | `/admin/cost` | cost | ✅ 已接入 |

### 3.3 页面存在但硬编码 (8 个)

| 优先级 | 区域 | 页面 | 路由 | 说明 |
|--------|------|------|------|------|
| 🟡 中 | 用户端 | 子账号管理 | `/profile/sub-accounts` | enterprise API 已就绪 |
| 🟡 中 | 用户端 | 个人信息编辑 | `/profile` | user API 已就绪 |
| 🟡 中 | 用户端 | 充值 | `/tokens/topup` | token API 已就绪 |
| 🟡 中 | 供应商门户 | Dashboard | `/vendor` | vendor API 已就绪 |
| 🟡 中 | 供应商门户 | 商品管理 | `/vendor/products` | 后端缺少 GET/POST |
| 🟡 中 | 供应商门户 | 订单 | `/vendor/orders` | order API 已有 |
| 🟡 中 | 供应商门户 | 结算 | `/vendor/settlements` | commission API 已有 |
| 🟢 低 | 管理端 | 渠道管理 | `/admin/channels` | 后端缺少 channels API |

### 3.4 完全缺失的前端页面 — 无!

> 所有 Domain (除 plugin) 都有前端页面覆盖。plugin 为 P2 模块，后端/router 均未集成，暂不纳入前端范围。

### 3.5 前端基础设施

| 项目 | 状态 |
|------|------|
| 路由 | 完整 (react-router v7), 33 页面路由全部注册 |
| API 层 | **8 个 service 文件** (auth/tokens/orders/admin/stats/notifications/enterprise/byok) |
| 状态管理 | 基础 (authStore + themeStore, 2 个 zustand store) |
| 国际化 | i18next 配置完成, 翻译文件存在 |
| 表单验证 | zod + react-hook-form (auth 页面使用) |
| UI 组件 | Ant Design 5 + TailwindCSS |
| 测试 | **零覆盖** (无测试框架/用例/CI) |
| 构建 | Vite + TypeScript |
| 流式请求 | useSSE hook 已实现 |
| 导航菜单 | AdminLayout +4 项 (护栏/企业/成本含子菜单), UserLayout +2 项 (BYOK/模型市场) |

---

## 四、文档评估

| 文档类型 | 数量 | 状态 |
|---------|------|------|
| PRD (需求) | 23 文件 | 完整 |
| PDD (设计) | 21 文件 | 完整 |
| 缺 PDD 的 Domain | 3 个 | stats, commission, log (在合并文件中提及但无独立设计) |
| API 文档 | PDD 03-api.md | 完整 |
| 数据库文档 | PDD 02-database.md | 完整 |
| 开发计划 | 本文件 | 更新版 |

---

## 五、关键风险矩阵

| 风险 | 概率 | 影响 | 当前状态 | 缓解措施 |
|------|------|------|---------|---------|
| 前端页面量大, 人手不足 | 高 | 工期延后 | 22 页硬编码 + 6 domain 缺失 | 优先核心流程页, 非核心用简单列表代替 |
| Plugin 后端未完成 | 中 | S7 阻塞 | router 未注册, 无 handler | P2 优先级, 可降级或后移 |
| 前端无测试体系 | 高 | 质量退化 | 0 测试 | 接入 vitest + testing-library, 核心页面先覆盖 |
| API 文档与实现有偏差 | 中 | 对接返工 | 部分 API 未标注 | 用 Swagger/OpenAPI 自动生成并验证 |
| SQLite 并发瓶颈 | 低 | 生产性能 | WAL 模式已启用 | 用户量 <1000 够用, 超阈值分库 |
| 无 CI/CD | 中 | 回归风险 | 手动测试 | 引入 GitHub Actions: lint → test → build |

---

## 六、更新后的开发计划

### 原计划 vs 实际 vs 剩余

| 里程碑 | 原估算 | 实际状态 | 剩余工作 | 重估算 |
|--------|--------|---------|---------|--------|
| S0 项目骨架 | 10-15d | ✅ 完成 | — | 0d |
| S1 用户与 Token | 9-13d | ✅ 完成 | — | 0d |
| S2 交易链路 | 10-14d | ✅ 完成 | — | 0d |
| S3 代理转发 | 14-21d | ✅ 完成 (含 Rerank/Video) | — | 0d |
| S4 增值模块 | 13-20d | ✅ 完成 | — | 0d |
| S5 安全增强 | 9-12d | ✅ 完成 (含 guardrail CRUD) | — | 0d |
| S6 全功能 | 15-22d | ✅ 基本完成 | plugin + 部分 P1/P2 API | 见下方 |
| **S7 前端补全** | — | ⚠️ 进行中 (76%) | 8 页面 API 接入 + 体验优化 | **8-16d** |
| **S8 后端补全** | — | ⚠️ 19 个 API | P1/P2 缺口 + plugin | **3-6d** |
| **S9 测试与 CI** | — | ❌ 未开始 | 前端测试体系 + CI/CD | **8-12d** |

### 当前进度与评估差异

相比于上轮评估 (2026-06-01 第二轮)，本轮新增:
- 13 个后端 API (P0 全部补全 + P2 Rerank/Video/Benchmarks + OAuth/SSO/Teams)
- 14 个前端页面接入真实 API (11→25)
- 7 个前端全新页面 (guardrail/byok/enterprise/cost/market)
- 4 个新 API service 文件 (stats/notifications/enterprise/byok)
- 后端 135 个注册端点, 所有 25 测试包全绿

### 新里程碑简化

```
S7: 前端收尾 (8-16 人日) ─── 8 页面 + 体验优化
S8: 后端收尾 (3-6 人日) ─── P1/P2 API 补全 + plugin
S9: 测试/CI/上线 (8-12 人日)
```

---

### S7A: 剩余页面 API 接入 (3-5d) 🟡 P1

**目标**: 将剩余 8 个硬编码页面接入真实 API

| 优先级 | 页面 | 人日 | 说明 |
|--------|------|------|------|
| 1 | 子账号管理 (`/profile/sub-accounts`) | 0.5d | enterprise 子账号 API 已就绪 |
| 2 | 个人信息编辑 (`/profile`) | 0.5d | user profile API 已就绪 |
| 3 | 充值页 (`/tokens/topup`) | 0.5d | token API 已就绪 |
| 4 | 供应商门户 Dashboard (`/vendor`) | 0.5d | vendor stats API 已就绪 |
| 5 | 供应商门户 订单/结算/商品 (3页) | 1.5d | 3 个 vendor 页面 API 接入 |

**可交付**: 前端 33/33 页面 100% API 接入, 供应商可自助管理

### S7B: 体验优化 (5-11d) 🟢 P2

| 项目 | 人日 | 说明 |
|------|------|------|
| 加载骨架屏 | 1d | 列表/详情页统一 Skeleton loading |
| 错误处理 | 1d | 全局 ErrorBoundary + API 错误统一提示 |
| 响应式适配 | 1d | 手机端适配关键页面（购买/登录/Dashboard） |
| 暗色模式 | 1d | Ant Design 5 ConfigProvider 主题切换 |
| 流式对话测试页 (Playground) | 1d | 代理转发测试 (开发用) |
| PWA 离线支持 | 1d | Service Worker + 离线缓存 |
| E2E 测试 (可选) | 3d | Playwright: 注册→登录→购买→调用 全流程 |

### ~~S7C: S5-S6 新模块页面~~ — ✅ 已完成!

> guardrail、byok、enterprise、cost、market 全部 7 个新页面已完成并接入 API

---

### S8: 后端收尾 (3-6d)

#### S8A: API 补全 (2-4d)

| 优先级 | API 数量 | 人日 | 说明 |
|--------|---------|------|------|
| P1 缺口 6 个 | 6 | 1.5d | vendor 自服务商品 CRUD, channels 管理, guardrail 全局配置 |
| P2 缺口 13 个 | 13 | 2d | system config, admins, byok usage/preference, cache stats/config, STT, variant, commission rates, settlements |
| Plugin 完整实现 | 3-5 API | 1.5d | router 注册 + handler 实现 |

#### S8B: handler 测试补全 (0.5-1d)

| 项目 | 人日 | 说明 |
|------|------|------|
| handler 层测试 | 0.5d | market/payment/plugin/token/vendor handler 测试 |
| 覆盖率配置 | 0.5d | `go test -coverprofile` + 基准线 |

---

### S9: 测试体系 + CI/CD + 上线准备 (8-12d)

#### S9A: 前端测试 (4-6d)

| 项目 | 人日 | 说明 |
|------|------|------|
| 接入 vitest | 0.5d | 测试框架配置 |
| 接入 testing-library | 0.5d | 组件测试工具 |
| API service 单元测试 | 1d | 8 个 API service (auth/tokens/orders/admin/stats/notifications/enterprise/byok) |
| 核心组件测试 | 1d | Auth pages + Token buy + Order list |
| E2E 冒烟测试 | 1.5d | Playwright: 注册→登录→购买→下单 全流程 |
| 测试 CI 集成 | 0.5d | GitHub Actions test job |

#### S9B: CI/CD + DevOps (2-3d)

| 项目 | 人日 | 说明 |
|------|------|------|
| GitHub Actions | 1d | lint → test → build 流水线 |
| 测试覆盖率报告 | 0.5d | go test -coverprofile + Codecov |
| Docker 构建优化 | 0.5d | 多阶段构建, 镜像瘦身 |
| 部署文档 | 0.5d | Nginx/Docker/K8s 部署指南 |

---

## 七、总工时汇总

### 已完成
| 阶段 | 人日 |
|------|------|
| S0-S6 (后端 + 基础前端) | 80-117d ✅ |
| S7 前端补全 (部分) | ~20d (14 页面 API 接入 + 7 新页面) |

### 剩余工作
| 阶段 | 人日 | 优先级 |
|------|------|--------|
| S7A 剩余页面 API 接入 | 3-5d | 🟡 P1 |
| S7B 体验优化 | 5-11d | 🟢 P2 |
| S8A 后端 API 补全 | 2-4d | 🟡 P1 |
| S8B handler 测试补全 | 0.5-1d | 🟡 P1 |
| S9A 前端测试 | 4-6d | 🟡 P1 |
| S9B CI/CD | 2-3d | 🟡 P1 |
| **剩余总计** | **16.5-30d** | |

### 总体
| 指标 | 数值 |
|------|------|
| 已完成 | ~100-137 人日 |
| 剩余 | 16.5-30 人日 |
| **全项目总计** | **~117-167 人日** |
| 当前完成度 | **~82%** |
| 预计可交付 MVP | S7A+S8 = **5.5-10d** (含 P1 后端补全) |

---

## 八、建议执行顺序

```
优先级排序 (更新):
  第 1 周: S8A 后端 P1 API 补全 → S7A 剩余页面 API 接入
  第 2 周: S9B CI/CD 搭建 → S8B handler 测试补充
  第 3 周: S9A 前端测试体系 → S7B 体验优化
  第 4 周: P2 后端 API + S8A plugin → 代码冻结, 上线准备

当前实际可并行:
  - S7A (前端) + S8A (后端) 可同时进行
  - S9B (CI/CD) 可立即开始, 不依赖其他任务
  - S7B (体验优化) 在 S7A 完成后进行

建议用人:
  后端 1 人 × 1 周 (补全 P1 API + 测试)
  前端 1 人 × 2-3 周 (页面收尾 + 测试 + 体验)
  或 全栈 1 人 × 3-4 周
```

---

## 九、附录: 全 18 Domain 完成度明细

| # | Domain | 后端 API | 前端页面 | 测试 | 文档 | 综合 | 说明 |
|---|--------|---------|---------|------|------|------|------|
| 1 | shared/config | ✅ | N/A | ✅ | ✅ | 100% | |
| 2 | shared/model | ✅ | N/A | ✅ | ✅ | 100% | 25+ 表 |
| 3 | shared/cache | ✅ | N/A | ✅ | ✅ | 100% | |
| 4 | shared/middleware | ✅ | N/A | ✅ | ✅ | 100% | JWT/CORS/RateLimit/Language |
| 5 | domain/user | ✅ | ✅ | ✅ | ✅ | 98% | 含 OAuth/SSO, 35 测试 |
| 6 | domain/token | ✅ | ✅ | ✅ | ✅ | 95% | 含 usage history, 29 测试 |
| 7 | domain/order | ✅ | ✅ | ✅ | ✅ | 95% | 含退款/管理员操作, 27 测试 |
| 8 | domain/payment | ✅ | ✅ | ⚠️ | ✅ | 90% | 17 测试, 缺 handler 测试 |
| 9 | domain/proxy | ✅ | ✅ | ✅ | ✅ | 98% | ⭐ 最完整, Rerank/Video, 77 测试 |
| 10 | domain/vendor | ✅ | ⚠️ | ⚠️ | ✅ | 85% | 供应商门户 4 页待接入, 19 测试 |
| 11 | domain/risk | ✅ | ✅ | ✅ | ⚠️ | 90% | Admin 3 页已接入, 22 测试 |
| 12 | domain/notify | ✅ | ✅ | ✅ | ⚠️ | 90% | 通知/模板已接入, 31 测试 |
| 13 | domain/stats | ✅ | ✅ | ✅ | ❌ | 85% | Dashboard/charts 已接入, 26 测试 |
| 14 | domain/commission | ✅ | ⚠️ | ✅ | ❌ | 80% | vendor 结算页待接入, 25 测试 |
| 15 | domain/log | ✅ | ✅ | ✅ | ❌ | 90% | 审计/调用日志已接入, 16 测试 |
| 16 | domain/guardrail | ✅ | ✅ | ✅ | ✅ | **95%** | 规则CRUD+日志, 22 测试, **从60%↑** |
| 17 | domain/byok | ✅ | ✅ | ✅ | ✅ | **95%** | Key管理页+API, 23 测试, **从60%↑** |
| 18 | domain/cost | ⚠️ | ✅ | ✅ | ✅ | **85%** | 预算+告警已接入, 27 测试, **从50%↑** |
| 19 | domain/enterprise | ⚠️ | ✅ | ✅ | ✅ | **90%** | SSO+Teams+子账号已接入, 24 测试, **从55%↑** |
| 20 | domain/market | ⚠️ | ✅ | ✅ | ✅ | **90%** | Benchmarks+Recommend, 12 测试, **从50%↑** |
| 21 | domain/plugin | ❌ | ❌ | ⚠️ | ✅ | 20% | service 有代码, router 未注册 |

> 图例: ✅ 完成 | ⚠️ 部分 | ❌ 缺失 | N/A 不适用
> 
> **本轮提升**: guardrail +35%, byok +35%, cost +35%, enterprise +35%, market +40%, vendor +15%, payment +10%
