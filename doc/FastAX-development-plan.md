# FastAX 生产开发计划

> 基于 PRD v3.0 和 PDD v3.0 制定
> 估算总工时: 80-117 人日 | 18 domains | 7 个里程碑

---

## 1. 全局依赖图

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

---

## 2. Domain 卡片

### Layer 0 — 基础设施 (shared/)

#### shared/config — 配置管理 (2-3d)

| 项目 | 内容 |
|------|------|
| 职责 | 加载配置文件 + 环境变量，全局配置结构体 |
| 技术 | Viper |
| 产出 | `internal/shared/config/config.go` — Server/DB/Redis/MQ 配置 |
| 参考 | one-api `common/config/config.go` |

#### shared/model — 数据模型 (3-5d)

| 项目 | 内容 |
|------|------|
| 职责 | GORM 模型定义 + AutoMigrate + 索引 |
| 技术 | GORM + SQLite WAL |
| 产出 | `internal/shared/model/*.go` — 全部 18+ 表模型 |
| 设计参考 | PDD §7.2 数据表详细设计 |

模型清单: user, user_profile, sub_account, token_product, user_token, token_transfer, channel, ability, order, refund, payment, settlement, call_log, supplier_vendor, supplier_product, commission_record, notification, risk_event, supported_language, byok_key, guardrail_rule, guardrail_log, semantic_cache, model_variant, provider_health

#### shared/cache — Redis 缓存 (2-3d)

| 项目 | 内容 |
|------|------|
| 职责 | Redis 客户端封装 + 缓存 Key 规范 + TTL 策略 |
| 产出 | `CacheGet`/`CacheSet`/`CacheDecrease` 通用接口 |
| Key 规范 | `user:quota:{id}` (60s), `token:{key}` (60s), `route:health:{id}` (10s) |
| 参考 | one-api `common/redis.go` + `model/cache.go` |

#### shared/middleware — Gin 中间件 (3-4d)

| 项目 | 内容 |
|------|------|
| 产出 | `auth.go` — JWT 鉴权 + `Accept-Language` 解析 |
| | `rate_limit.go` — IP/用户/接口三维限流 |
| | `language.go` — 浏览器语言检测 + 回退链 |
| | `distributor.go` — 渠道分发（注入 channel 到 context）|
| 参考 | one-api `middleware/` 全部 |

---

### Layer 1 — 核心业务

#### domain/user — 用户模块 (5-7d)

| 项目 | 内容 |
|------|------|
| 接口 | `UserService { Register, Login, RefreshToken, GetUser, UpdateLanguage, AddSubAccount }` |
| 数据模型 | user (含 preferred_language), user_profile, sub_account |
| 关键逻辑 | BCrypt(cost=12)、JWT Access(24h) + Refresh(7d Rotation)、RBAC 角色 |
| 注册 | 手机号+邮箱+验证码、海外用户仅邮箱注册 |
| 登录 | 账号密码 / 验证码 / OAuth、连续 5 次失败锁定 15min |
| 测试重点 | 注册→验证→登录→Token 刷新→权限校验→子账号管理 |
| 参考 | one-api `model/user.go` + `controller/user.go` |

#### domain/token — Token 商品 (4-6d)

| 项目 | 内容 |
|------|------|
| 接口 | `TokenService { GetProducts, GetUserTokens, Buy, Transfer, Extract }` |
| 数据模型 | token_product, user_token, token_transfer |
| 关键逻辑 | 库存管理、到期前 7/3/1 天提醒、AES-256-GCM 加密存储 |
| Redis | `token:{key}` TTL 60s |
| 事件 | `event.token.low-stock` → notify |
| 测试重点 | 购买→到账→查询→转让→提取→到期提醒 |

---

### Layer 2 — 交易链路

#### domain/order — 订单管理 (5-7d)

| 项目 | 内容 |
|------|------|
| 接口 | `OrderService { Create, Get, List, Cancel, ApplyRefund }` |
| 数据模型 | order (6 状态), refund |
| 状态机 | DRAFT → PENDING → PAID → COMPLETED / CANCELLED; PAID → REFUNDING → REFUNDED |
| 超时 | PENDING 30 分钟自动取消 |
| 审批 | 退款: ≤1000 元运营审批, >1000 元上级审批 |
| 事件 | `event.order.created` → payment; `event.order.paid` → token+notify+stats+commission |

#### domain/payment — 支付对接 (5-7d)

| 项目 | 内容 |
|------|------|
| 接口 | `PaymentService { Pay, Callback, Refund, Reconciliation }` |
| 数据模型 | payment, settlement |
| 网关 | 微信支付 + Stripe (MVP), 支付宝 (P1) |
| 对账 | T+1 自动对账, 逐笔匹配 order_id/amount/status |
| 结算 | T+7 自动生成结算单, 银行转账/PayPal 提现 |
| 事件 | `event.settlement.created` / `event.settlement.paid` → vendor+notify |

---

### Layer 3 — 代理转发核心 ⭐

#### domain/proxy — 转发引擎 (10-15d)

**这是 FastAX 最核心的模块**，直接参考 one-api `relay/` 包的代码实现。

```
internal/domain/proxy/
├── service.go              # ProxyService interface
├── handler.go              # HTTP handler (流式/非流式)
│
├── relay/                  # 路由引擎 (参考 one-api relay/)
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

| 项目 | 内容 |
|------|------|
| **路由算法** | `CacheGetRandomSatisfiedChannel` — 优先级分组→同优先级权重随机 |
| **路由缓存** | `InitChannelCache` 全量内存加载, `SyncChannelCache` 60s 定时刷新 |
| **Ability 索引** | group+model+channel_id 复合索引, O(1) 查询候选渠道 |
| **熔断** | 5xx/超时自动禁用(DB status=3), 排除 401/403/429, 健康检测恢复 |
| **重试** | `Relay()` 函数: 失败 → `shouldRetry()` 判断 → 跨渠道重试(跳过刚失败渠道) |
| **流式转发** | SSE 流式输出, io.Copy 零拷贝, 流中断自动切换备用供应商 |
| **转发流水线** | 限流→鉴权→余额检查+预扣→路由决策(内存缓存)→请求重写→转发→响应处理(后扣)→MQ |
| **计费** | 预扣(估算)→转发→后扣(多退少补); `BatchUpdateEnabled` 批量刷入 |
| **Adaptor** | `Init/GetRequestURL/SetupRequestHeader/ConvertRequest/ConvertImageRequest/DoRequest/DoResponse/GetModelList/GetChannelName` (9 方法) |
| **协议** | OpenAI `/v1/chat/completions` + Anthropic `/v1/messages` + Gemini 原生 |
| **模型变体** | `:nitro`(最快)、`:floor`(最便宜)、`:thinking`(推理) 语法解析 |
| **参考** | one-api `relay/` 全套 |
| **测试重点** | 流式转发正确性、熔断恢复、跨渠道重试、配额扣减一致性、多协议兼容 |

#### domain/vendor — 供应商入驻 (4-6d)

| 项目 | 内容 |
|------|------|
| 接口 | `VendorService { Register, GetProfile, CreateProduct, GetSales, GetSettlements }` |
| 数据模型 | supplier_vendor, supplier_product |
| 入驻流程 | 提交申请→管理员审核(佣金配置)→商品上架(合规审核)→销售运营→T+7 结算 |
| 状态机 | pending → approved \| rejected → suspended → terminated |
| 定价 | 供应商自主定价, 平台限价范围内; 支持批量/会员/限时折扣 |
| 适配器 | 入驻供应商实现 Adaptor 接口即可接入平台 |

---

### Layer 4 — 增值模块

#### domain/risk — 风控引擎 (3-5d)

| 项目 | 内容 |
|------|------|
| 接口 | `RiskService { Evaluate, GetRules, CreateRule, GetEvents }` |
| 数据模型 | risk_event |
| 规则引擎 | IF condition THEN action (level) — 4 级: 绿(记录)/黄(预警)/橙(限权)/红(冻结) |
| 检测项 | 批量注册、大额交易、异地登录、API 调用频率异常 |
| 事件 | `event.risk.triggered` → notify+order |

#### domain/notify — 通知模块 (3-5d)

| 项目 | 内容 |
|------|------|
| 接口 | `NotifyService { SendInApp, SendEmail, SendSMS, GetNotifications, MarkRead }` |
| 数据模型 | notification (template_code+language 唯一索引) |
| 通道 | 站内信 / 邮件 (SendCloud) / 短信 (阿里云/腾讯云) |
| 多语言 | 每个 template_code+language 一条记录, 按用户语言发送 |
| 回退链 | zh-TW→zh-CN→en, ja→en |

#### domain/stats — 统计看板 (3-4d)

| 项目 | 内容 |
|------|------|
| 接口 | `StatsService { GetDashboard, GetUsage, GetBills, GetSalesReport }` |
| 数据 | 聚合查询 + ES 全文检索 |
| 看板 | 核心数据概览、Token 用量趋势、供应商销售看板、翻译覆盖率统计 |

#### domain/commission — 佣金结算 (2-3d)

| 项目 | 内容 |
|------|------|
| 接口 | `CommissionService { Calculate, GetSettlements, Withdraw }` |
| 数据模型 | commission_record |
| 逻辑 | 分销佣金计算、T+7 结算周期、提现审核 |

#### domain/log — 审计日志 (2-3d)

| 项目 | 内容 |
|------|------|
| 接口 | `LogService { WriteAuditLog, QueryLogs, Export }` |
| 存储 | ES, 保留 12 个月 |
| 逻辑 | 异步写入(队列)、操作日志全覆盖、合规导出 SOC2/GDPR |

---

### Layer 5 — P0 增强模块

#### domain/guardrail — 安全护栏 (5-7d)

| 项目 | 内容 |
|------|------|
| 接口 | `GuardrailService { CheckBefore, CheckAfter, GetRules, UpdateRules, GetLogs }` |
| 数据模型 | guardrail_rule, guardrail_log |
| 两阶段 | Before(输入检测) → 转发 → After(输出检测) |
| 4 类检测 | PII(邮箱/手机/SSN/信用卡)、Prompt 注入(越狱/泄露)、密钥扫描(AKSK/Token)、内容审核 |
| 3 模式 | Enforce(拦截阻断)、Monitor(告警放行)、Log(仅记录) |
| 性能 | PII ≤ 50ms, 注入 ≤ 100ms; 护栏故障 bypass 不阻塞主请求 |
| 事件 | `event.guardrail.triggered` → notify+risk |

#### domain/byok — 自带 Key (4-5d)

| 项目 | 内容 |
|------|------|
| 接口 | `BYOKService { AddKey, GetKeys, DeleteKey, UpdateKey, GetUsage, SetPreference }` |
| 数据模型 | byok_key (AES-256-GCM + 独立 IV 加密) |
| 混合路由 | 优先 BYOK → 不足时自动回退平台 Token (需用户授权) |
| Key 轮换 | 定时检测过期/失效 Key, 自动通知 |
| 平台费用 | ≤ 5% 服务费或零加成 |

---

### Layer 6 — P1/P2 扩展模块

#### domain/cost — 成本优化 (4-6d)

| 项目 | 内容 |
|------|------|
| 语义缓存 | Prompt 向量化(ONNX/sentence-transformers) → Redis Stack 相似度检索 → 缓存命中按 10% 计费 |
| 数据模型 | semantic_cache (向量+加密响应+命中计数) |
| 预算封顶 | 月/日/用户维度成本上限, 超限自动熔断(仅管理员可解除) |
| 模型回退链 | gpt-4 → claude-3-haiku → deepseek-chat |

#### domain/enterprise — 企业功能 (4-6d)

| 项目 | 内容 |
|------|------|
| SSO | SAML2/OIDC, 支持 Okta/Azure AD/Google Workspace |
| 团队 | team + project 两级隔离, 独立计费配额 |
| 审计导出 | SOC2/GDPR/HIPAA 格式 JSON/CSV |
| 数据驻留 | 指定区域偏好(仅国内/仅海外) |

#### domain/market — 模型市场 (3-4d)

| 项目 | 内容 |
|------|------|
| 模型对比 | 价格/延迟/上下文/能力评分 多维度对比 |
| 健康面板 | 供应商历史可用率/平均延迟/P95/错误率 (公开) |
| 数据模型 | model_variant, provider_health |

#### domain/plugin — 插件系统 (4-6d)

| 项目 | 内容 |
|------|------|
| 3 钩子 | beforeRequest / afterRequest / route 扩展点 |
| 沙箱 | 独立 panic recover + 超时 500ms + 故障不影响核心链路 |
| 插件市场 | 第三方开发者提交, 管理员安装启用 (P3) |

---

## 3. 里程碑与工时汇总

| 里程碑 | 包含 domains | 人日 | 可交付物 |
|--------|-------------|------|---------|
| **S0** 项目骨架 | shared/config, model, cache, middleware | 10-15d | `go run ./cmd/fastax` 可启动, DB 自动建表 |
| **S1** 用户与 Token | user, token | 9-13d | 注册/登录/JWT/Token 购买全流程 |
| **S2** 交易链路 | order, payment | 10-14d | 下单→支付→回调→结算完整闭环 |
| **S3** 代理转发 ⭐ | **proxy**, vendor | 14-21d | **可转发 Token 调用** — MVP 核心里程碑 |
| **S4** 增值 | risk, notify, stats, commission, log | 13-20d | 通知可达、风控可用、统计可见 |
| **S5** 安全增强 | guardrail, byok | 9-12d | PII/注入防护上线、BYOK 可用 |
| **S6** 全功能 | cost, enterprise, market, plugin | 15-22d | 语义缓存、SSO、模型市场 |
| | **总计** | **80-117d** | **完整平台** |

---

## 4. 关键风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| proxy 路由引擎复杂度高 | 高 | S3 延期 | 直接参考 one-api `relay/` 代码, 不做过度设计; 先实现 OpenAI 协议, Anthropic/Gemini 后续追加 |
| 支付网关对接调试 | 中 | S2 延期 | MVP 仅对接微信支付 + Stripe, 支付宝放 P1; 沙箱环境提前联调 |
| SQLite 并发瓶颈 | 低 | S3 性能 | MVP 用户量 <1000 完全够用; WAL 模式 + 读写分离; 达到瓶颈后分库 |
| 语义缓存向量检索精度 | 中 | S6 质量 | 先用 Redis Stack 内置向量能力, 不引入独立向量数据库; 缓存穿透时回退到直接转发 |
| 安全护栏误杀率 | 中 | S5 体验 | 默认 Monitor 模式上线, 观察一周后切换 Enforce; 配置白名单机制 |

---

## 5. 实施原则

1. **单体优先** — 所有 domain 在同一进程内, 通过 Go interface 调用; 用户 >5000 后按需抽取 gRPC 服务
2. **参考代码驱动** — proxy 模块直接参考 `ref/one-api/` 代码, 不重新发明轮子
3. **先协议后扩展** — 先实现 OpenAI 兼容协议 (P0), Anthropic/Gemini 原生协议 (P0) 紧随其后, 模型变体/多模态 (P1)
4. **测试即文档** — `go test ./internal/domain/...` 覆盖每个 domain 的核心路径; proxy 模块必须含流式/熔断/重试测试
5. **每个 milestone 可独立验证** — 每个阶段结束时 `go test ./...` 全绿 + 手动冒烟测试通过
