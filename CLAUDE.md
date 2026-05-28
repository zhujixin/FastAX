# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目定位

FastAX 是一个 **Token 代理与交易平台**（Go 单体优先），作为连接终端用户与 Token 源头（OpenAI、Claude、Gemini、DeepSeek、Qwen、GLM 等）的中间枢纽。

**当前阶段**：纯文档项目（产品设计阶段），尚无源代码。

---

## Commands (when code exists)

```bash
# 构建
go build -o bin/fastax ./cmd/fastax

# 测试全部
go test ./...

# 测试单个 domain
go test ./internal/domain/proxy/...

# 启动开发服务
go run ./cmd/fastax

# 数据库迁移 (AutoMigrate)
go run ./cmd/fastax -migrate

# 前端
cd web && npm run dev    # 开发
cd web && npm run build  # 构建
cd web && npm test       # 测试
```

---

## 文档导航

| 你需要什么 | 读哪个文件 |
|-----------|----------|
| 需求定义 + 优先级 | [doc/FastAX-PRD/00-index.md](doc/FastAX-PRD/00-index.md)（→ 文件索引定位具体 domain） |
| 设计 + 实现方案 | [doc/FastAX-PDD/00-index.md](doc/FastAX-PDD/00-index.md)（→ 文件索引定位具体 domain） |
| API 接口定义 | [doc/FastAX-PDD/03-api.md](doc/FastAX-PDD/03-api.md) |
| 数据库表定义 | [doc/FastAX-PDD/02-database.md](doc/FastAX-PDD/02-database.md) |
| 实现优先级 + 工時估算 | [doc/FastAX-development-plan.md](doc/FastAX-development-plan.md)（7 里程碑, 80-117 人日） |
| 生产级参考代码 | [ref/one-api/](ref/one-api/)（Go + Gin + GORM + Adaptor 模式） |
| one-api 架构剖析 | [ref/one-api-architecture-reference.md](ref/one-api-architecture-reference.md) |
| 术语表 | [doc/FastAX-PRD/19-glossary.md](doc/FastAX-PRD/19-glossary.md) |

**阅读顺序**：查需求 → PRD → 找设计 → PDD → 抄实现 → `ref/one-api/`

### 核心文件映射（PRD → PDD → Go package → one-api 参考）

| 功能域 | PRD 文件 | PDD 文件 | Go package | one-api 参考 |
|--------|---------|---------|-----------|-------------|
| 用户认证 | [01-user-auth.md](doc/FastAX-PRD/01-user-auth.md) | [05-user.md](doc/FastAX-PDD/05-user.md) | `domain/user` | `model/user.go` |
| Token 代理/路由 | [02-token-proxy-vendor.md](doc/FastAX-PRD/02-token-proxy-vendor.md) | [06-token-proxy-vendor.md](doc/FastAX-PDD/06-token-proxy-vendor.md) | `domain/proxy` | `relay/` (核心) |
| 订单/支付 | [03-order-payment.md](doc/FastAX-PRD/03-order-payment.md) | [07-order-payment.md](doc/FastAX-PDD/07-order-payment.md) | `domain/order` + `payment` | — |
| 风控 | [04-risk.md](doc/FastAX-PRD/04-risk.md) | [08-risk.md](doc/FastAX-PDD/08-risk.md) | `domain/risk` | `monitor/` |
| i18n | [06-i18n.md](doc/FastAX-PRD/06-i18n.md) | [11-i18n-overseas.md](doc/FastAX-PDD/11-i18n-overseas.md) | `shared/i18n` | `common/i18n/` |
| 多协议 | [07-multi-protocol.md](doc/FastAX-PRD/07-multi-protocol.md) | [12-multi-protocol.md](doc/FastAX-PDD/12-multi-protocol.md) | `domain/proxy/adaptor` | `relay/adaptor/` |
| 安全护栏 | [09-guardrails.md](doc/FastAX-PRD/09-guardrails.md) | [14-guardrails.md](doc/FastAX-PDD/14-guardrails.md) | `domain/guardrail` | — |
| BYOK | [10-byok.md](doc/FastAX-PRD/10-byok.md) | [15-byok.md](doc/FastAX-PDD/15-byok.md) | `domain/byok` | — |
| 供应商入驻 | [02-token-proxy-vendor.md](doc/FastAX-PRD/02-token-proxy-vendor.md) §2.6 | [06-token-proxy-vendor.md](doc/FastAX-PDD/06-token-proxy-vendor.md) §6 | `domain/vendor` | — |

### PRD 需求 ID 前缀 → Domain

| 前缀 | Domain | 覆盖范围 |
|------|--------|---------|
| ROUTE- | `domain/proxy` | 路由/转发/熔断/重试 |
| SUP- | `domain/vendor` | 供应商入驻/适配器 |
| LANG- | `shared/i18n` | 多语言国际化 |
| PROTO- | `domain/proxy/adaptor` | 多协议原生支持 |
| MEDIA- | `domain/proxy/adaptor` | 多模态支持 |
| GRDL- | `domain/guardrail` | 安全护栏 |
| BYOK- | `domain/byok` | 自带 Key |
| PLUG- | `domain/plugin` | 插件系统 |
| COST- | `domain/cost` | 成本优化 |
| ENT- | `domain/enterprise` | 企业功能 |
| MKT- | `domain/market` | 模型市场 |

---

## 架构快照

```
7 层: Client → CDN → Nginx → fastax-server (Go 单体, 18 domain) → MQ/Redis → SQLite/ES → 外部供应商
关键: 单体优先 + domain package + Go interface 解耦 → 用户>5000 后按需拆 gRPC
```

### Domain 依赖层级（实现顺序）

```
shared/ (Layer 0 — 基础设施, 无依赖)
  ├── config/  ├── model/  ├── cache/  └── middleware/
       │
       ▼
Layer 1: domain/user, domain/token       (仅依赖 shared/)
       │
       ▼
Layer 2: domain/order, domain/payment    (依赖 user, token)
       │
       ▼
Layer 3: domain/proxy, domain/vendor     (依赖 token, order) ← 核心
       │
       ▼
Layer 4: domain/risk, notify, stats, commission, log   (依赖 proxy, order)
       │
       ▼
Layer 5: domain/guardrail, domain/byok   (依赖 proxy, user) ← P0 增强
       │
       ▼
Layer 6: domain/cost, enterprise, market, plugin       (P1/P2 扩展)
```

### 18 domains 各阶段一览

| 阶段 | 里程碑 | Domains | 估算人日 |
|------|--------|---------|---------|
| S0 | 项目骨架 | `shared/config`, `shared/model`, `shared/cache`, `shared/middleware` | 10-15d |
| S1 | 用户与 Token | `domain/user`, `domain/token` | 9-13d |
| S2 | 交易链路 | `domain/order`, `domain/payment` | 10-14d |
| S3 ⭐ | **代理转发** | **`domain/proxy`**, `domain/vendor` | **14-21d** |
| S4 | 增值 | `domain/risk`, `notify`, `stats`, `commission`, `log` | 13-20d |
| S5 | 安全增强 | `domain/guardrail`, `domain/byok` | 9-12d |
| S6 | 全功能 | `domain/cost`, `enterprise`, `market`, `plugin` | 15-22d |

### Proxy 模块内部结构（核心）

```
internal/domain/proxy/
├── service.go              # ProxyService interface
├── handler.go              # HTTP handler (流式/非流式)
│
├── relay/                  # 路由引擎 (直接参考 one-api relay/)
│   ├── controller/
│   │   └── relay.go        # 转发控制 + 重试循环
│   ├── adaptor.go          # GetAdaptor(apiType) 分发器
│   └── adaptor/            # 供应商适配器
│       ├── openai/
│       ├── anthropic/
│       └── gemini/
│
├── monitor/                # 健康检测 + 熔断 (参考 one-api monitor/)
│   ├── health.go           # 10s ping + 5min 周期性
│   └── circuit.go          # ShouldDisableChannel 决策
│
└── router.go               # /v1/* 路由注册
```

**转发流水线**：限流 → 鉴权 → 余额检查+预扣 → 路由决策(内存缓存) → 请求重写 → 转发 → 响应处理(后扣) → MQ

---

## 实施原则

1. **单体优先** — 所有 domain 在同一进程内, 通过 Go interface 调用; 用户 >5000 后按需抽取 gRPC 服务
2. **参考代码驱动** — proxy 模块直接参考 `ref/one-api/` 代码, 不重新发明轮子
3. **先协议后扩展** — 先实现 OpenAI 兼容协议 (P0), Anthropic/Gemini 原生协议 (P0) 紧随其后, 模型变体/多模态 (P1)
4. **测试即文档** — `go test ./internal/domain/...` 覆盖每个 domain 的核心路径; proxy 模块必须含流式/熔断/重试测试
5. **每个 milestone 可独立验证** — 各阶段结束时 `go test ./...` 全绿 + 手动冒烟测试通过

---

## 关键设计（实现前必读）

| 决策 | 要点 | 参考 |
|------|------|------|
| **路由** | 优先级分组 + 同优先级权重随机（非加权评分公式） | one-api `CacheGetRandomSatisfiedChannel` |
| **熔断** | 5xx/超时自动禁用渠道(DB status=3)，排除 401/403/429，健康检测恢复 | one-api `monitor.ShouldDisableChannel` |
| **重试** | 失败后跨渠道重试（跳过刚失败渠道），仅重试 429/5xx/超时 | one-api `Relay()` 函数 |
| **计费** | 两阶段：预扣(估算) → 转发 → 后扣(按实际多退少补) + 批量更新 | one-api `BatchUpdateEnabled` |
| **Adaptor** | 9 方法接口: Init/GetRequestURL/SetupRequestHeader/ConvertRequest/ConvertImageRequest/DoRequest/DoResponse/GetModelList/GetChannelName | one-api `relay/adaptor/interface.go` |
| **渠道选择** | `ability_index` 表 (group+model+channel 复合索引) + 全量内存缓存 `InitChannelCache`, `SyncChannelCache` 60s 定时刷新 | one-api `model/ability.go` + `model/cache.go` |
| **流式转发** | SSE 流式输出, io.Copy 零拷贝, 流中断自动切换备用供应商 | one-api relay 核心函数 |

### 关键风险

| 风险 | 缓解措施 |
|------|---------|
| proxy 路由引擎复杂度高 | 直接参考 one-api `relay/` 代码, 不做过度设计; 先实现 OpenAI 协议, Anthropic/Gemini 后续追加 |
| 支付网关对接调试 | MVP 仅对接微信支付 + Stripe, 支付宝放 P1; 沙箱环境提前联调 |
| SQLite 并发瓶颈 | MVP 用户量 <1000 完全够用; WAL 模式 + 读写分离; 达到瓶颈后分库 |
| 安全护栏误杀率 | 默认 Monitor 模式上线, 观察一周后切换 Enforce; 配置白名单机制 |

---

## 技术栈

| 层 | 选型 |
|----|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite (WAL) / Redis (go-redis) / RabbitMQ (amqp091-go) |
| 前端 | React 18 / TypeScript / Vite / Ant Design 5 / TailwindCSS / i18next |
| 部署 | Nginx / Docker / K8s / Prometheus+Grafana / ELK |

## 目录结构（实现时）

```
cmd/fastax/main.go          # 入口
internal/
├── shared/                  # 共享层
│   ├── model/               #   GORM 模型 (18+ 表)
│   ├── config/              #   Viper 配置
│   ├── cache/               #   Redis 缓存
│   ├── middleware/           #   Gin 中间件 (auth/ratelimit/language/distributor)
│   ├── relay/               #   路由引擎 (adaptor 分发 + 转发控制)
│   └── i18n/                #   国际化
├── domain/                  # 业务域 (各含 service.go + handler.go)
│   ├── user/ ├── token/ ├── order/ ├── payment/ ├── proxy/ (核心)
│   ├── vendor/ ├── risk/ ├── notify/ ├── stats/ ├── commission/ ├── log/
│   ├── guardrail/ ├── byok/ ├── cost/ ├── enterprise/ ├── market/ ├── plugin/
└── router/                  # Gin 路由 (api.go + relay.go)
```
