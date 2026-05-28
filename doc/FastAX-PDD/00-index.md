# FastAX Token 代理平台 — 产品设计文档 (PDD)

**文档状态**：定稿  
**创建日期**：2026-05-28  
**版本号**：v3.0  
**基于 PRD 版本**：v3.0  

> **本文档已按 domain 拆分为多文件，详见 [FastAX-PDD/](.) 目录。** 推荐使用下方文件索引定位。

---

## 文件索引

| # | 文件 | Go package | 内容 | 行数 |
|---|------|-----------|------|------|
| 00 | [00-index.md](00-index.md) | — | 文档控制、Claude 指南、架构总览 | 291 |
| 01 | [01-architecture.md](01-architecture.md) | — | 系统/前端/后端架构设计 | 454 |
| 02 | [02-database.md](02-database.md) | `shared/model/` | 数据库表 + Redis 缓存 | 608 |
| 03 | [03-api.md](03-api.md) | `shared/` | API 接口设计 | 414 |
| 04 | [04-security-deploy-stack.md](04-security-deploy-stack.md) | `shared/` | 安全/部署/技术栈 | 248 |
| 05 | [05-user.md](05-user.md) | `domain/user` | 用户模块 | 69 |
| 06 | [06-token-proxy-vendor.md](06-token-proxy-vendor.md) | `domain/proxy`+`token`+`vendor` | Token 代理/路由/供应商 | 300 |
| 07 | [07-order-payment.md](07-order-payment.md) | `domain/order`+`payment` | 订单/支付 | 102 |
| 08 | [08-risk.md](08-risk.md) | `domain/risk` | 风控 | 32 |
| 09 | [09-notify.md](09-notify.md) | `domain/notify` | 通知 | 43 |
| 10 | [10-channel-adapter.md](10-channel-adapter.md) | `shared/relay` | 渠道适配器 | 51 |
| 11 | [11-i18n-overseas.md](11-i18n-overseas.md) | `shared/i18n` | 多语言+海外优化 | 101 |
| 12 | [12-multi-protocol.md](12-multi-protocol.md) | `domain/proxy/adaptor` | 多协议 | 53 |
| 13 | [13-multimedia.md](13-multimedia.md) | `domain/proxy/adaptor` | 多模态 | 36 |
| 14 | [14-guardrails.md](14-guardrails.md) | `domain/guardrail` | 安全护栏 | 39 |
| 15 | [15-byok.md](15-byok.md) | `domain/byok` | BYOK | 30 |
| 16 | [16-plugin.md](16-plugin.md) | `domain/plugin` | 插件 | 27 |
| 17 | [17-cost-optimization.md](17-cost-optimization.md) | `domain/cost` | 成本优化 | 31 |
| 18 | [18-enterprise.md](18-enterprise.md) | `domain/enterprise` | 企业功能 | 11 |
| 19 | [19-model-marketplace.md](19-model-marketplace.md) | `domain/market` | 模型市场 | 12 |

---

## 目录

1. [文档控制](#1-文档控制)
2. [Claude Code 实施指南](#2-claude-code-实施指南)
3. [系统总体架构](#3-系统总体架构)
4. [前端架构设计](#4-前端架构设计)
5. [后端架构设计](#5-后端架构设计)
6. [模块详细设计](#6-模块详细设计)
7. [数据库详细设计](#7-数据库详细设计)
8. [接口详细设计](#8-接口详细设计)
9. [安全架构设计](#9-安全架构设计)
10. [部署架构设计](#10-部署架构设计)
11. [技术栈选型](#11-技术栈选型)
12. [附录：术语对照](#12-附录术语对照)

---

## 1. 文档控制

| 版本 | 日期 | 修订人 | 修订内容 |
|------|------|--------|----------|
| v1.0 | 2026-05-27 | — | 初稿，基于 PRD v1.0 编写 |
| v2.0 | 2026-05-27 | — | 重写：新增多语言(i18n)、供应商入驻与销售平台、海外→国内模型优化、转发逻辑设计(参考 CC Switch) |
| v3.0 | 2026-05-28 | — | Go 单体架构重构；路由/熔断/计费优化(参考 one-api)；新增 PROTO/MEDIA/GRDL/BYOK/PLUG/COST/ENT/MKT 8 大模块设计；新增 6 张数据表 |

**新增设计覆盖 PRD 需求**:
- 多语言模块 (LANG-01 ~ LANG-06)
- Token 厂家入驻与销售平台 (SUP-01 ~ SUP-16, SUP-19)
- 海外→国内模型专项优化 (OCN-01 ~ OCN-06, OCN-SUP-01 ~ OCN-SUP-04)
- 转发逻辑设计 (ROUTE-01 ~ ROUTE-17)
- 合规需求 (R-COMPL-09 ~ R-COMPL-18)
- **多协议原生支持 (PROTO-01 ~ PROTO-10)**
- **多模态支持 (MEDIA-01 ~ MEDIA-07)**
- **安全护栏 (GRDL-01 ~ GRDL-09)**
- **BYOK (BYOK-01 ~ BYOK-07)**
- **插件系统 (PLUG-01 ~ PLUG-07)**
- **成本优化 (COST-01 ~ COST-08)**
- **企业功能 (ENT-01 ~ ENT-08)**
- **模型市场 (MKT-01 ~ MKT-06)**

---

## 2. Claude Code 实施指南

### 2.1 文档使用方式

| 文档 | 作用 | 使用时机 |
|------|------|---------|
| `doc/FastAX-PRD/00-index.md` | 需求总索引 + 文件导航 | 查"要做什么" → 按 domain 定位到对应文件 |
| `doc/FastAX-PRD/01-user-auth.md` ~ `21-market-analysis.md` | 按 domain 拆分的需求详情 | 每个文件对应一个 Go package |
| `doc/FastAX-PDD.md` (本文) | 设计 + 实现参考 | 查"怎么做" |
| `ref/one-api/` | 生产级参考实现 | 抄代码模式 + 接口定义 |
| `CLAUDE.md` | 项目约束 + 技术栈 | 持久上下文 |

### 2.2 实施阶段与依赖顺序

```
阶段1 ── 基础设施与数据层
  ├── shared/model/ (GORM 模型 + SQLite 表)
  ├── shared/config/ (Viper 配置管理)
  ├── shared/cache/ (Redis 缓存层)
  └── shared/middleware/ (Gin 中间件: 鉴权/限流/语言)
       ↓
阶段2 ── 核心业务域
  ├── domain/user    (注册/登录/用户管理/JWT)
  ├── domain/token   (Token 商品/库存/持有)
  ├── domain/order   (订单创建/状态机)
  └── domain/payment (支付对接/对账)
       ↓
阶段3 ── 代理转发核心
  ├── domain/proxy/relay/    (路由引擎 + Adaptor 接口)
  ├── domain/proxy/adaptor/  (OpenAI/Anthropic/Gemini 适配器)
  ├── domain/proxy/monitor/  (健康检测 + 熔断)
  └── domain/vendor          (供应商入驻/商品)
       ↓
阶段4 ── 增值模块
  ├── domain/risk       (风控规则)
  ├── domain/notify     (通知: 站内信/短信/邮件)
  ├── domain/stats      (统计看板)
  └── domain/commission (佣金结算)
       ↓
阶段5 ── P0 增强模块
  ├── domain/guardrail  (安全护栏: PII/注入检测)
  └── domain/byok       (自带 Key)
       ↓
阶段6 ── P1/P2 模块
  ├── domain/proxy/adaptor/  (多模态: 图片/语音/视频)
  ├── domain/cost      (语义缓存/预算)
  ├── domain/market    (模型市场/对比)
  ├── domain/enterprise (SSO/团队)
  └── domain/plugin    (插件系统)
```

### 2.3 Go package 目录结构

```
cmd/
└── fastax/
    └── main.go              # 入口: 初始化 DB/Redis/路由 → Gin.Run

internal/
├── shared/
│   ├── model/               # GORM 模型 (所有表)
│   │   ├── user.go
│   │   ├── token.go
│   │   ├── channel.go
│   │   ├── ability.go
│   │   ├── order.go
│   │   ├── call_log.go
│   │   └── ...              # + byok_key, guardrail_rule 等 v3.0 表
│   ├── config/              # Viper 配置
│   ├── cache/               # Redis 缓存 (参考 one-api common/redis.go)
│   ├── middleware/          # Gin 中间件 (参考 one-api middleware/)
│   │   ├── auth.go          # JWT 鉴权
│   │   ├── rate_limit.go    # 限流
│   │   ├── language.go      # Accept-Language 解析
│   │   └── distributor.go   # 渠道分发 (参考 one-api middleware/distributor.go)
│   ├── relay/               # 路由引擎 (参考 one-api relay/)
│   │   ├── adaptor.go       # Adaptor 分发器 (GetAdaptor)
│   │   └── controller/      # 转发控制 (RelayTextHelper / RelayImageHelper)
│   ├── monitor/             # 健康检测 + 熔断 (参考 one-api monitor/)
│   └── i18n/                # 国际化 (参考 one-api common/i18n/)

├── domain/                   # 业务域 (每个 domain 一个 Service interface)
│   ├── user/
│   │   ├── service.go        # UserService interface
│   │   ├── service_impl.go   # 单体实现
│   │   ├── handler.go        # HTTP handler
│   │   └── model.go          # domain 内 DTO
│   ├── token/
│   ├── order/
│   ├── payment/
│   ├── proxy/               # 核心: 路由转发
│   │   ├── service.go        # ProxyService interface
│   │   ├── handler.go        # OpenAI 兼容端点
│   │   └── router.go         # 路由注册
│   ├── vendor/
│   ├── risk/
│   ├── notify/
│   ├── stats/
│   ├── commission/
│   ├── log/
│   ├── guardrail/            # 安全护栏 (P0)
│   ├── byok/                 # BYOK (P0)
│   ├── cost/                 # 成本优化 (P1)
│   ├── enterprise/           # 企业功能 (P2)
│   ├── market/               # 模型市场 (P1)
│   └── plugin/               # 插件系统 (P2)
│
└── router/
    ├── api.go              # /api/* 业务路由
    └── relay.go            # /v1/* OpenAI 兼容路由
```

### 2.4 核心接口定义

每个 domain 的 Service interface 即未来 gRPC proto 定义。单体阶段直接调用实现，微服务阶段替换为 gRPC client。

```go
// domain/user/service.go
type UserService interface {
    Register(ctx, email, password) (user, error)
    Login(ctx, account, password) (token, error)
    RefreshToken(ctx, refreshToken) (accessToken, error)
    GetUser(ctx, userId) (user, error)
    UpdateLanguage(ctx, userId, locale) error
}

// domain/proxy/service.go  — 核心路由
type ProxyService interface {
    ChatCompletion(ctx, request) (response, error)     // 流式/非流式
    ChatCompletionStream(ctx, request) (chan Chunk, error)
    GetModels(ctx, group) ([]Model, error)
}

// domain/vendor/service.go  — 供应商适配器
type VendorService interface {
    Register(ctx, vendorInfo) (vendor, error)
    CreateProduct(ctx, vendorId, product) (product, error)
    GetSales(ctx, vendorId, period) (sales, error)
}
```

### 2.5 构建与测试命令

```bash
# 开发
go run ./cmd/fastax                  # 启动服务
go test ./internal/domain/...        # 测试业务域
go test ./internal/shared/...        # 测试共享层

# 构建
go build -o bin/fastax ./cmd/fastax  # 编译单二进制

# 数据库迁移 (GORM AutoMigrate)
go run ./cmd/fastax -migrate         # 自动建表
```

---

## 3. 系统总体架构

### 2.1 架构分层

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              客户端层 (Client)                                       │
│    Web 端 (React + i18n)   │   移动端 (H5/App)   │  第三方 Client (OpenAI API)      │
└─────────────────────────────────────┬───────────────────────────────────────────────┘
                                      │
┌─────────────────────────────────────▼───────────────────────────────────────────────┐
│                            CDN / 负载均衡层                                           │
│                     CloudFlare / Nginx / ALB (HTTPS 终止)                            │
│                     翻译静态 TEXT 文件通过 CDN 分发                                     │
└─────────────────────────────────────┬───────────────────────────────────────────────┘
                                      │
┌─────────────────────────────────────▼───────────────────────────────────────────────┐
│                              网关层 (Gateway)                                         │
│      Nginx / Gin 内置路由                                                              │
│      • 路由转发  • 限流  • 鉴权  • 日志记录  • 请求/响应转换  • Accept-Language 解析   │
└──────┬──────────────────┬──────────────────┬────────────────────────────────────────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  用户端API    │  │  管理端API    │  │  开放API      │
│  (公网)       │  │  (内网/VPN)   │  │  (公网)       │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           业务服务层 (Go 微服务)                                 │
│                                                                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │ user-    │ │ token-   │ │ order-   │ │ payment- │ │ notify-  │ │ vendor-  │    │
│  │ service  │ │ service  │ │ service  │ │ service  │ │ service  │ │ service  │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│  │ risk-    │ │ stats-   │ │ proxy-   │ │ commish- │ │ log-     │                │
│  │ service  │ │ service  │ │ service  │ │ service  │ │ service  │                │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘                │
│                                                                                     │
│  ┌────────────────────────────────────────────────────────────────────────────┐    │
│  │     中间件层: 消息队列(RabbitMQ) · 缓存(Redis) · 调度器(gocron)              │    │
│  └────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                             数据存储层 (Data Store)                                   │
│                                                                                     │
│  SQLite (WAL)          Redis                   Elasticsearch          │
│  • 业务核心数据          • 缓存/会话              • 审计/日志分析        │
│  • 事务强一致(ACID)       • Token 临时             • 全文检索            │
│  • 供应商商品/结算        • 翻译文件缓存                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              外部集成层 (External)                                    │
│  Token供应商(海外)  │  Token供应商(国内)  │  微信支付  │  支付宝  │  Stripe  │  短信  │  邮件  │
│  OpenAI/Claude/Gemini  │  DeepSeek/Qwen/GLM │                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                             前端静态资源层                                            │
│  CDN:  locales/ (翻译 TEXT 文件) · 静态构建产物 (JS/CSS)                             │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 架构设计原则

| 原则 | 说明 |
|------|------|
| **微服务架构** | 独立业务域拆分为独立 Go 服务，独立部署扩缩容，每服务独立 SQLite DB |
| **API 网关统一入口** | 所有外部请求经网关统一处理，网关层解析 `Accept-Language` 注入请求头 |
| **无状态设计** | 服务实例无状态，Session 数据外置到 Redis，语言偏好通过 JWT 或 header 传递 |
| **异步解耦** | 非实时操作通过消息队列异步处理（通知、日志、风控、结算） |
| **可观测性** | 全链路追踪 + 指标监控 + 日志聚合 |
| **防御性编程** | 所有外部输入校验，服务间调用超时 + 熔断 + 重试 |
| **多语言优先** | 前端框架层内置 i18n，API 设计考虑多语言字段 (i18n TEXT) |
| **供应商隔离** | 每个供应商独立适配器，新增供应商不影响核心路由 |

### 2.3 核心数据流
