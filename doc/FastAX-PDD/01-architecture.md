
#### 2.3.1 Token 调用主流程（含路由决策）

```
用户 Client                 网关 Gateway             代理服务 Proxy             路由引擎             供应商 Supplier
    │                        │                        │                      │                      │
    │── POST /v1/chat/completions ──→│                        │                      │                      │
    │  (stream: true)       │                        │                      │                      │
    │                        │── 鉴权 + 限流 + 费率扣除 ──→│                      │                      │
    │                        │                        │── 路由决策 (ROUTE-01) ──→│                      │
    │                        │                        │                      │── 加权轮询 ──────────→│
    │                        │                        │                      │── 健康检查 ──────────→│
    │                        │                        │                      │── 延迟排序 ──────────→│
    │                        │                        │←── 选定供应商 ────────│                      │
    │                        │                        │── POST /v1/chat/ ────────────────────────→│
    │                        │                        │   completions                              │
    │                        │                        │←── SSE 流式响应 ──────────────────────────│
    │                        │←── 转发流式响应 ─────────│                      │                      │
    │←── 流式 SSE ────────────│                        │                      │                      │
    │                        │                        │                      │                      │
    │                        │  ── 异步 ──→│                      │                      │
    │                        │        调用日志 → MQ → 日志服务 → ES         │                      │
    │                        │        用量扣减 → Redis → 异步持久化          │                      │
    │                        │        路由决策 → 统计服务                     │                      │
```

#### 2.3.2 购买支付流程

```
用户 Client                订单服务                支付服务             支付网关(微信/支付宝/Stripe)
    │                        │                      │                      │
    │── POST /api/tokens/buy ──→│                      │                      │
    │                        │── 创建订单(待支付) ──→│                      │
    │                        │                      │── 调用支付下单 ──→│
    │                        │                      │←── 支付链接 ───────│
    │                        │←── 返回支付链接 ───────│                      │
    │←── 支付 URL ────────────│                      │                      │
    │                                                                      │
    │── 用户完成支付 ──────────────────────────────────────────────────→│
    │                                                                      │
    │←── 支付回调 ───────────────────────────────────────────────────────│
    │                        │                      │                      │
    │                        │                      │←── 异步通知 ────────│
    │                        │                      │── 校验签名 ────────→│(查询)
    │                        │                      │── 更新支付状态 ─────│
    │                        │←── 支付成功通知 ──────│                      │
    │                        │── 更新订单状态        │                      │
    │                        │── 发放 Token         │                      │
    │                        │── 异步: 通知服务发送消息                     │
    │←── 购买成功 ────────────│                      │                      │
```

#### 2.3.3 供应商入驻与销售流程

```
供应商 Client          供应商服务(Vendor)          管理后台                 交易系统                支付系统
    │                        │                      │                      │                      │
    │── 提交入驻申请 ────────→│                      │                      │                      │
    │  (SUP-01)              │── 创建待审核记录 ────→│                      │                      │
    │                        │                      │── 管理员审核 ─────────│                      │
    │                        │                      │←── 审核通过 ─────────│                      │
    │←── 审核结果通知 ────────│                      │                      │                      │
    │                        │                      │                      │                      │
    │── 上架商品 ────────────→│                      │                      │                      │
    │  (SUP-04)              │── 合规审核 ─────────→│                      │                      │
    │                        │                      │←── 审核通过 ─────────│                      │
    │←── 商品上架成功 ────────│                      │                      │                      │
    │                        │                      │                      │                      │
    │  (用户购买供应商商品)                              │── 创建订单 ─────→│                      │
    │                        │                      │                      │── 支付处理 ──────────→│
    │                        │                      │                      │←── 支付成功 ──────────│
    │                        │                      │←── 订单完成 ────────│                      │
    │                        │                      │                      │                      │
    │── 查看销售看板 ────────→│                      │                      │                      │
    │                        │── 聚合订单/调用数据 ──→│                      │                      │
    │←── 销售数据 ────────────│                      │                      │                      │
    │                        │                      │                      │                      │
    │── 申请结算提现 ────────→│                      │                      │                      │
    │  (SUP-10)              │── 生成结算单 ────────→│                      │                      │
    │                        │                      │── 结算审批 ─────────│                      │
    │                        │                      │── 发起付款 ──────────────────────────────→│
    │                        │                      │                      │                      │
    │←── 结算完成通知 ────────│                      │                      │                      │
```

---

## 4. 前端架构设计

### 3.1 前端整体架构 (含 i18n)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                             用户端 (User Portal)                                   │
│                                                                                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ 首页/商城  │ │ 控制台    │ │ Token管理 │ │ 订单中心  │ │ 个人中心  │ │ 供应商   │  │
│  │ (Home)    │ │(Dashboard)│ │ (Tokens) │ │ (Orders) │ │ (Profile)│ │ 门户    │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │(Vendor) │  │
│                                                                   └──────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                             管理后台 (Admin Panel)                                │
│                                                                                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ 运营看板   │ │ 用户管理  │ │ Token管控 │ │ 交易管理  │ │ 风控管理  │ │ 供应商   │  │
│  │ (Monitor) │ │ (Users)  │ │ (Tokens) │ │ (Orders) │ │ (Risk)   │ │ 管理    │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                            │
│  │ 系统管理   │ │ 运营工具  │ │ 审计日志  │ │ 多语言   │                            │
│  │ (System)  │ │ (Ops)    │ │ (Audit)  │ │ 配置    │                            │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘                            │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                             共享层 (Shared)                                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ 组件库     │ │ Hooks    │ │ Utils    │ │ API 客户端│ │ 类型定义   │ │ i18n     │  │
│  │ (UI Kit)  │ │          │ │          │ │ (Axios)  │ │ (Types)  │ │ 系统    │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 关键设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 框架 | React 18 + TypeScript | 生态成熟，团队熟悉度高 |
| 构建工具 | Vite | 开发体验好，构建速度快 |
| 路由 | React Router v6 | SPA 路由，支持嵌套布局和权限路由 |
| 状态管理 | Zustand | 轻量，TypeScript 支持好，无样板代码 |
| UI 组件库 | Ant Design 5 | 企业级组件丰富，后台管理场景成熟 |
| 样式方案 | TailwindCSS + CSS Modules | 原子化样式 + 组件级隔离 |
| HTTP 客户端 | Axios + 拦截器 | 统一错误处理、Token 刷新、请求取消、语言头注入 |
| **国际化** | **i18next + react-i18next** | **Lazy Load、插值、复数、命名空间、SSR 兼容** |
| 图表 | ECharts / AntV | 统计看板可视化 |
| 表单 | React Hook Form + Zod | 高性能表单 + Schema 校验 |
| 流式处理 | EventSource / fetch + ReadableStream | SSE 流式接收 Token 响应 |

### 3.3 i18n 架构设计 (PRD §6.9)

#### 3.3.1 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| 核心框架 | i18next | 支持命名空间、插值、复数、上下文、延迟加载 |
| React 绑定 | react-i18next | React hooks (useTranslation)、Trans 组件 |
| 语言检测 | i18next-browser-languageDetector | 自动检测浏览器语言 |
| 缓存 | i18next-localStorage-backend | 持久化用户语言选择 |
| 构建集成 | i18next-scanner / i18next-parser | 自动提取翻译 Key |
| SSR 兼容 | next-i18next / i18next-http-middleware | 根据技术栈选择 |

#### 3.3.2 翻译文件结构 (PRD R-i18n-MNT-02)

```
src/i18n/locales/
├── zh-CN/                     # 简体中文（源语言）
│   ├── common.json            # 全局通用文案
│   ├── home.json              # 首页
│   ├── auth.json              # 登录/注册
│   ├── token.json             # Token 商城
│   ├── order.json             # 订单相关
│   ├── payment.json           # 支付
│   ├── profile.json           # 个人中心
│   ├── admin.json             # 管理后台
│   ├── notification.json      # 通知模板
│   ├── error.json             # 错误信息
│   ├── docs.json              # 开发者文档
│   ├── vendor.json            # 供应商门户
│   └── compliance.json        # 合规文案
├── en/                        # 英文
│   ├── common.json
│   └── ...
├── ja/                        # 日语 (Phase 4)
├── ko/                        # 韩语 (Phase 4)
├── vi/                        # 越南语 (Phase 5)
└── th/                        # 泰语 (Phase 5)
```

#### 3.3.3 Key 命名规范 (PRD R-i18n-MNT-01)

```
<模块>:<子模块>.<路径>.<描述>
```

示例:
```json
{
  "token:product.buy_button": "立即购买",
  "token:product.price_label": "价格",
  "order:list.status.pending": "待支付",
  "common:nav.login": "登录",
  "common:error.required": "此项为必填"
}
```

#### 3.3.4 语言切换与持久化 (PRD LANG-01)

```
首次访问:
  1. 浏览器 Accept-Language → 语言检测器
  2. 匹配支持的语言列表 → 设置默认语言
  3. 加载对应翻译 TEXT (CDN)
  4. 渲染页面

用户主动切换:
  1. 点击语言选择器 → 切换语言
  2. 更新 i18next 语言
  3. 加载对应翻译 TEXT (按需)
  4. 持久化到 localStorage
  5. 已登录: 同步到用户设置

语言回退链 (PRD R-BIZ-42):
  zh-TW → zh-CN → en
  ja-JP → ja → en
  en-US → en
```

#### 3.3.5 性能优化 (PRD R-PERF-07, R-PERF-08)

| 策略 | 实现 |
|------|------|
| **按需加载** | 翻译文件按 page/chunk 拆分，路由级懒加载 |
| **CDN 缓存** | 翻译 TEXT 上传至 CDN，Cache-Control: max-age=31536000 |
| **首屏内联** | 首屏必要的翻译 Key 内联到 HTML，避免额外网络请求 |
| **SSR 同步** | SSR 时服务端注入当前语言翻译，客户端 hydration 无需额外加载 |
| **DOM 更新** | react-i18next Suspense + Trans 组件，只更新变更的 DOM 节点 |

### 3.4 页面路由设计

#### 3.4.1 用户端路由 (新增: 供应商门户)

| 路径 | 页面 | 权限 | 备注 |
|------|------|------|------|
| `/` | 首页/商城 | 公开 | Token 商品列表 & 推荐，随语言切换 |
| `/login` | 登录 | 公开 | Login 页面含语言选择器 |
| `/register` | 注册 | 公开 | Register 页面含语言选择器 |
| `/forgot-password` | 忘记密码 | 公开 | — |
| `/dashboard` | 用户控制台 | 登录 | 用量概览、快捷操作 |
| `/tokens` | 我的 Token | 登录 | — |
| `/tokens/:id` | Token 详情 | 登录 | — |
| `/tokens/buy` | 购买 Token | 登录 | — |
| `/orders` | 订单列表 | 登录 | — |
| `/orders/:id` | 订单详情 | 登录 | — |
| `/profile` | 个人中心 | 登录 | 含语言偏好设置 |
| `/profile/sub-accounts` | 子账号管理 | 企业用户 | — |
| `/bills` | 账单明细 | 登录 | — |
| `/notifications` | 消息中心 | 登录 | 按用户语言展示 |
| `/vendor/dashboard` | 供应商控制台 | 供应商 | 销售看板 |
| `/vendor/products` | 商品管理 | 供应商 | 上架/下架/定价 |
| `/vendor/orders` | 供应商订单 | 供应商 | — |
| `/vendor/settlements` | 结算管理 | 供应商 | 结算单/提现 |

#### 3.4.2 管理端路由 (新增: 供应商管理, i18n 配置)

| 路径 | 页面 | 权限 | 备注 |
|------|------|------|------|
| `/admin` | 运营看板 | 管理员 | 核心数据概览 |
| `/admin/users` | 用户管理 | 管理员 | CRUD、审核、冻结 |
| `/admin/tokens` | Token 管控 | 管理员 | 供应商、渠道、库存、价格 |
| `/admin/tokens/suppliers` | 供应商管理(平台) | 管理员 | 自有供应商管理 |
| `/admin/tokens/channels` | 渠道管理 | 管理员 | 优先级、健康状态 |
| `/admin/orders` | 交易管理 | 管理员 | — |
| `/admin/orders/reports` | 对账报表 | 管理员 | — |
| `/admin/risk` | 风控管理 | 管理员 | — |
| `/admin/risk/rules` | 风控规则 | 管理员 | — |
| `/admin/risk/blacklist` | 黑名单 | 管理员 | — |
| `/admin/vendors` | **入驻供应商管理** | 管理员 | 审核、佣金配置、违规处理 |
| `/admin/i18n` | **多语言配置** | 管理员 | 语种启用/禁用、默认语言设置 |
| `/admin/system` | 系统设置 | 超级管理员 | — |
| `/admin/system/admins` | 管理员账号 | 超级管理员 | — |
| `/admin/system/logs` | 审计日志 | 管理员 | — |
| `/admin/operations` | 运营工具 | 管理员 | 优惠券、活动配置 |

### 3.5 组件目录结构

```
src/
├── user/                         # 用户端应用
│   ├── pages/
│   │   ├── home/                 # 首页/商城
│   │   ├── auth/                 # 登录/注册
│   │   ├── dashboard/            # 控制台
│   │   ├── tokens/               # Token 管理
│   │   ├── orders/               # 订单
│   │   ├── profile/              # 个人中心
│   │   └── vendor/               # 供应商门户 (新增)
│   │       ├── Dashboard.tsx
│   │       ├── Products.tsx
│   │       ├── Orders.tsx
│   │       └── Settlements.tsx
│   ├── layouts/
│   └── App.tsx
├── admin/                        # 管理端应用
│   ├── pages/
│   │   ├── dashboard/
│   │   ├── users/
│   │   ├── tokens/
│   │   ├── orders/
│   │   ├── risk/
│   │   ├── vendors/              # 入驻供应商管理 (新增)
│   │   ├── i18n-config/          # 多语言配置 (新增)
│   │   └── system/
│   ├── layouts/
│   └── App.tsx
├── shared/
│   ├── components/
│   │   ├── LanguageSelector/     # 语言选择器组件 (新增)
│   │   ├── Button/
│   │   ├── Modal/
│   │   ├── Table/
│   │   └── Form/
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── usePermissions.ts
│   │   ├── usePagination.ts
│   │   ├── useStreamResponse.ts
│   │   └── useLanguage.ts        # 语言切换 hook (新增)
│   ├── api/
│   │   ├── client.ts             # Axios + Accept-Language 拦截
│   │   ├── auth.ts
│   │   ├── tokens.ts
│   │   ├── orders.ts
│   │   ├── vendors.ts            # 供应商 API (新增)
│   │   └── i18n.ts               # 多语言 API (新增)
│   ├── stores/
│   │   ├── authStore.ts
│   │   ├── tokenStore.ts
│   │   ├── notificationStore.ts
│   │   └── i18nStore.ts          # 语言偏好 (新增)
│   ├── types/
│   │   ├── user.ts
│   │   ├── token.ts
│   │   ├── order.ts
│   │   ├── vendor.ts             # 供应商类型 (新增)
│   │   └── api.ts
│   ├── utils/
│   │   ├── format.ts             # Intl 日期/数字/货币格式化
│   │   ├── encrypt.ts
│   │   └── validators.ts
│   └── i18n/                     # 国际化 (新增)
│       ├── index.ts              # i18next 初始化
│       ├── config.ts             # 语言配置
│       ├── locales/              # 翻译文件
│       │   ├── zh-CN/
│       │   └── en/
│       └── detector.ts           # 自定义语言检测
├── main.tsx
└── routes.tsx                    # 路由配置 (含权限守卫)
```

---

## 5. 后端架构设计

### 4.1 单体优先架构 (可演进至微服务)

**设计决策**：采用 Go 单体应用（参考 one-api），代码按业务域拆分为独立 package，通过 Go interface 解耦，
预留微服务拆分点。MVP 阶段单进程部署，用户/业务增长后可将独立 domain 抽取为独立 gRPC 服务。

> **单体优先原则**：先 monolith 后 microservices，避免过早分布式带来的复杂度。
> 拆分条件：注册用户 >5000 或单模块 TPS >500 或团队 >5 人。

```
┌──────────────────────────────────────────────────────────────┐
│                    fastax-server (Go 单进程)                    │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ domain/  │ │ domain/  │ │ domain/  │ │ domain/  │       │
│  │ user     │ │ token    │ │ order    │ │ payment  │       │
│  │ (接口:    │ │ (接口:   │ │ (接口:   │ │ (接口:   │       │
│  │  UserSvc)│ │ TokenSvc)│ │ OrderSvc)│ │PaySvc)  │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ domain/  │ │ domain/  │ │ domain/  │ │ domain/  │       │
│  │ proxy    │ │ vendor   │ │ risk     │ │ notify   │       │
│  │ (核心路由)│ │          │ │          │ │          │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ domain/  │ │ domain/  │ │ domain/  │ │ shared/  │       │
│  │ stats    │ │commish   │ │ log      │ │ model/   │       │
│  │          │ │          │ │          │ │midware/  │       │
│  └──────────┘ └──────────┘ └──────────┘ │relay/    │       │
│                                         │config/   │       │
│  ┌────────────────────────────────┐     │cache/    │       │
│  │ 各 domain 通过 Go interface    │     └──────────┘       │
│  │ 互相调用，后期可抽取为 gRPC    │                          │
│  └────────────────────────────────┘                          │
└──────────────────────────────────────────────────────────────┘
         │                    │                     │
         ▼                    ▼                     ▼
    SQLite (单 DB)        Redis (缓存)          RabbitMQ (异步)
```

### 4.2 业务域与微服务映射

| 业务域 (package) | 职责 | Go interface | 未来微服务名称 | 参考 one-api |
|------------------|------|-------------|---------------|-------------|
| **domain/user** | 注册/登录/用户管理/子账号/语言偏好 | UserService | user-service | model/user.go |
| **domain/token** | Token 商品/库存/持有/转让 | TokenService | token-service | model/token.go |
| **domain/proxy** | Token 中转/流式代理/智能路由/熔断器/健康检测/多协议适配 | ProxyService | proxy-service | relay/ (核心) |
| **domain/order** | 订单创建/流转/退款 | OrderService | order-service | — |
| **domain/payment** | 支付对接/对账/退款/供应商结算 | PaymentService | payment-service | — |
| **domain/vendor** | 供应商入驻/店铺/商品/定价/结算 | VendorService | vendor-service | — |
| **domain/risk** | 风控规则引擎/异常检测/预警 | RiskService | risk-service | monitor/ |
| **domain/stats** | 统计看板/报表/趋势/供应商销售 | StatsService | stats-service | — |
| **domain/notify** | 站内信/短信/邮件/多语言模板 | NotifyService | notify-service | common/message/ |
| **domain/commission** | 分销佣金/结算/提现 | CommissionService | commission-service | — |
| **domain/log** | 操作日志/审计日志采集 | LogService | log-service | — |

### 4.3 域间通信与微服务演进策略

**单体阶段**（MVP）：

| 通信方式 | 场景 | 实现 |
|----------|------|------|
| **Go interface 调用** | 域间实时查询 | 每个 domain 定义 Service interface，单体时直接调用实现 |
| **异步消息** | 非实时解耦 | RabbitMQ (amqp091-go) |
| **事件总线** | 领域事件发布/订阅 | RabbitMQ + 死信队列 |

**微服务阶段**（按需拆分后）：

| 通信方式 | 场景 | 技术选型 |
|----------|------|----------|
| **gRPC** | 域间实时高吞吐 | gRPC + Protocol Buffers |
| **异步消息** | 非实时解耦 | RabbitMQ |
| **事件总线** | 领域事件 | RabbitMQ + 死信队列 + 重试 |

> **演进策略**：每个 domain 的 Service interface 就是未来 gRPC 服务的 proto 定义。
> 拆分时只需将 interface 实现改为 gRPC client 调用，业务代码无需修改。
> 参考 one-api：虽为单体，但 package 边界清晰（controller/model/relay/middleware），易于抽取。

### 4.5 事件/消息 Topic 定义

| Topic | 生产者 | 消费者 | 说明 |
|-------|--------|--------|------|
| `event.user.registered` | domain/user | domain/notify | 新用户注册成功 |
| `event.user.login.abnormal` | domain/user | domain/risk | 异地登录检测 |
| `event.order.created` | domain/order | domain/payment, domain/stats | 订单创建 |
| `event.order.paid` | domain/order | domain/token, domain/notify, domain/stats, domain/commission | 支付完成 |
| `event.order.refunded` | domain/order | domain/payment, domain/token | 退款完成 |
| `event.token.consumed` | domain/proxy | domain/token, domain/stats, domain/risk | Token 消耗 |
| `event.token.low-stock` | domain/token | domain/notify | 库存预警 |
| `event.risk.triggered` | domain/risk | domain/notify, domain/order | 风控事件触发 |
| `event.channel.health-changed` | domain/proxy | domain/notify | 渠道状态变更 |
| `event.vendor.registered` | domain/vendor | admin, domain/notify | 供应商入驻申请 |
| `event.vendor.approved` | admin | domain/vendor, domain/notify | 审核通过 |
| `event.vendor.product.created` | domain/vendor | domain/token, domain/stats | 商品上架 |
| `event.settlement.created` | domain/payment | domain/vendor, domain/notify | 结算单生成 |
| `event.settlement.paid` | domain/payment | domain/vendor, domain/notify | 结算完成 |
| `event.user.language.changed` | domain/user | domain/notify | 用户语言偏好变更 |

---
