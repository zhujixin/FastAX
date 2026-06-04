# FastAX 统一开发计划 v4.0

**创建日期**: 2026-06-04 | **基于**: PRD v3.1 + PDD v3.1 | **继承**: v1 (S0-S6) + v2 (S8-S13) + PRD Phase 8

## 背景

项目当前有 **三套独立计划体系**需要整合：
- v1 开发计划 (S0-S6)：80-117 人日，✅ 已完成
- v2 缺口闭合计划 (S8-S13)：16-26 人日，待开始
- PRD v3.1 Phase 8 (CACHE/OBSV/MCP)：38-61 人日，待开始

本计划将三套体系合并为 **Phase 1-10** 统一路线图。

## 当前基线

| 维度 | 状态 |
|------|------|
| 已实现 Domain | 18 个（user/token/order/payment/proxy/vendor/risk/notify/stats/commission/log/guardrail/byok/cost/enterprise/market/plugin/system） |
| API 端点 | 157 个已注册 |
| 数据表 | 33 张（16 个 model 文件） |
| 测试 | 506+ 函数，25 包全绿 |
| 代码规模 | ~23,200 行 Go |

---

## Phase 1：数据层修复 🔴 CRITICAL — 2-4 人日

**依赖**: 无

| # | 任务 | 文件 | 人日 |
|---|------|------|------|
| 1.1 | 补齐 6 个缺失字段（VendorID×3, RouteDecision, MinPrice, MaxPrice） | `shared/model/token.go`, `order.go`, `call_log.go`, `vendor.go` | 1d |
| 1.2 | 补齐/修正 13 个数据库索引 | 全量 model 文件 | 0.5d |
| 1.3 | 修正 2 个字段类型偏差（SubAccount.TokenQuota, token_inventory.AlertThreshold） | `shared/model/user.go`, `token.go` | 0.5d |
| 1.4 | AutoMigrate 验证 + 回归测试 | `go test ./...` | 0.5d |

**产出**: 所有 model 与 PDD 一致，`go test ./...` 全绿

---

## Phase 2：安全加固 🔴 CRITICAL — 2-3 人日

**依赖**: Phase 1

| # | 任务 | 文件 | 人日 |
|---|------|------|------|
| 2.1 | bcrypt cost 10 → 12 | `internal/domain/user/service.go` | 0.5d |
| 2.2 | JWT HS256 → RS256 | `internal/shared/middleware/auth.go` | 1d |
| 2.3 | AES-256-GCM 加密封装完善 | `internal/shared/crypto/` | 0.5d |
| 2.4 | 数据脱敏工具（手机号/邮箱/身份证/银行卡） | `internal/shared/mask/` | 0.5d |
| 2.5 | 安全回归测试 | `go test ./internal/domain/user/... shared/...` | 0.5d |

**产出**: bcrypt cost 12, JWT RS256, AES-256-GCM, 脱敏工具

---

## Phase 3：Proxy 增强 🟠 HIGH — 3-5 人日

**依赖**: Phase 1

| # | 任务 | 文件 | 人日 |
|---|------|------|------|
| 3.1 | Distributor 中间件（按用户等级分发流量） | `internal/shared/middleware/distributor.go` | 1d |
| 3.2 | 两阶段计费管线完善（PreConsume→PostConsume→BatchFlush） | `internal/domain/proxy/billing.go` | 1d |
| 3.3 | 补齐 3 个国内适配器（DeepSeek/Qwen/GLM） | `internal/domain/proxy/relay/adaptor_*.go` | 2d |
| 3.4 | Proxy 集成测试（流式/熔断/重试） | `internal/domain/proxy/proxy_test.go` | 1d |

**产出**: Distributor 上线，计费完整，3 个适配器，测试覆盖率 ≥ 80%

---

## Phase 4：i18n + 前端补全 🟠 HIGH — 4-6 人日

**依赖**: Phase 1

| # | 任务 | 人日 |
|---|------|------|
| 4.1 | 后端翻译文件补全（zh-CN/en/ja/ko/vi/th） | 1d |
| 4.2 | 多语言错误消息覆盖全部 API | 1d |
| 4.3 | 前端 i18n 12 个缺失命名空间补全 | 1d |
| 4.4 | 前端 4 个缺失页面完善 | 2d |
| 4.5 | 前端 3 个共享组件 + Store/Hook/Util | 1d |

**产出**: 6 语言覆盖率 ≥ 95%，4 页面完成

---

## Phase 5：条件路由增强（ROUTE-18~22）🟡 P1 — 5-8 人日

**依赖**: Phase 3

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 5.1 | Token 长度路由（ROUTE-18）：按 prompt token count 选模型 | §5.2.8 | 1.5d |
| 5.2 | 内容类型路由（ROUTE-19）：图片/音频/视频自动分发 | §5.2.8 | 1.5d |
| 5.3 | 最低成本路由（ROUTE-20）：同模型多渠道比价 | §5.2.8 | 1.5d |
| 5.4 | P2C 负载均衡（ROUTE-21）：Power of Two Choices + PeakEWMA | §5.2.8 | 1d |
| 5.5 | 路由即时热更新（ROUTE-22）：Redis Pub/Sub <1s | §5.2.8 | 1d |
| 5.6 | 路由引擎集成测试 | — | 1d |

**产出**: 5 种条件路由策略，路由 API 配置端点

---

## Phase 6：语义缓存引擎（CACHE-01~07）🔴 P0 — 13-21 人日

**依赖**: Phase 3 | **ROI**: 降本 30-80%

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 6.1 | 新建 `domain/cache` 模块骨架（service+handler+model） | §5.17.1 | 1d |
| 6.2 | **L1 精确匹配缓存**（CACHE-01）：SHA256 + Redis + TTL | §5.17.2 | 2d |
| 6.3 | 缓存控制 Header（CACHE-04）：x-fastax-cache-* | §5.17.2 | 1d |
| 6.4 | 缓存计费（CACHE-06）：命中按原价 10-30% 计费 | §5.17.1 | 1d |
| 6.5 | Proxy 集成（CacheMiddleware 注入 Relay 流水线） | §5.17.5 | 2d |
| 6.6 | 验收：命中率 ≥ 15%，延迟 ≤ 1ms | — | 1d |
| 6.7 | **L2 语义向量缓存**（CACHE-02）：chromem-go + ONNX | §5.17.1 | 3d |
| 6.8 | 命名空间隔离（CACHE-07）：user/team/global | §5.17.3 | 1.5d |
| 6.9 | 验收：总命中率 ≥ 50%，P50 ≤ 15ms | — | 1d |
| 6.10 | L3 灰度区验证（CACHE-03）：廉价 LLM 判定 | §5.17.1 | 2d |
| 6.11 | 流式 SSE 缓存（CACHE-05） | §5.17.1 | 3d |
| 6.12 | 全量集成测试 + 性能基准 | — | 2d |

**产出**: 三层缓存架构，命中率 ≥ 50%，3 张缓存表

---

## Phase 7：流式护栏 + DLP（GRDL-10~14）🟡 P1 — 5-8 人日

**依赖**: Phase 3

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 7.1 | 流式护栏（GRDL-10）：SSE chunk-by-chunk 实时评估 | §5.11.5 | 2d |
| 7.2 | DLP 双阶段扫描（GRDL-11）：输入+输出 | §5.11.5 | 1.5d |
| 7.3 | 自定义正则替换（GRDL-12） | §5.11.5 | 1d |
| 7.4 | 工具级拦截（GRDL-13）：function_call/code_interpreter/MCP | §5.11.5 | 1d |
| 7.5 | 第三方护栏适配器接口（GRDL-14） | §5.11.5 | 1d |
| 7.6 | 护栏集成测试（含流式场景） | — | 1d |

**产出**: 流式护栏 + DLP + 第三方集成接口

---

## Phase 8：OpenTelemetry 可观测性（OBSV-01~06）🔴 P0 — 7-12 人日

**依赖**: Phase 3 | **企业选型硬指标**

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 8.1 | 新建 `domain/observability` 模块骨架 | §5.18.1 | 1d |
| 8.2 | OTel SDK 集成（TracerProvider + OTLP Exporter） | §5.18.1 | 2d |
| 8.3 | Span 树：Auth→Guardrail→Proxy→Upstream 全链路 | §5.18.2 | 2d |
| 8.4 | GenAI 语义约定（OBSV-02）：gen_ai.* 标准属性 | §5.18.3 | 1d |
| 8.5 | 结构化日志关联（OBSV-03）：trace_id + span_id | §5.18.3 | 1d |
| 8.6 | Prometheus 指标（OBSV-05）：/metrics + 40+ 指标 | §5.18.5 | 1.5d |
| 8.7 | W3C Traceparent 传播（OBSV-04） | §5.18.4 | 1d |
| 8.8 | Grafana 仪表板（OBSV-06）：4 面板 JSON | §5.18.6 | 1d |
| 8.9 | 集成测试 | — | 1d |

**产出**: 全链路追踪，/metrics 端点，4 个 Grafana 面板

---

## Phase 9：上下文压缩 + 碳感知路由（COST-09~11）🟢 P2 — 3-5 人日

**依赖**: Phase 6

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 9.1 | 上下文压缩网关（COST-09）：超阈值自动压缩 | §5.14.6 | 1.5d |
| 9.2 | 分层上下文策略（COST-10）：T0-T3 四层 | §5.14.6 | 1d |
| 9.3 | 碳感知路由（COST-11）：Electricity Maps API | §5.14.6 | 1d |
| 9.4 | 集成测试 | — | 0.5d |

**产出**: 上下文压缩引擎，碳感知路由

---

## Phase 10：MCP 网关 + 收尾 🟡 P1 — 23-36 人日

**依赖**: Phase 7, 8

| # | 任务 | PDD | 人日 |
|---|------|-----|------|
| 10.1 | 新建 `domain/mcp` 模块骨架 | §5.19.1 | 1d |
| 10.2 | 统一 MCP 端点（MCP-01）：SSE + Streamable HTTP | §5.19.1 | 3d |
| 10.3 | 传输桥接（MCP-02）：stdio↔SSE↔HTTP | §5.19.2 | 2d |
| 10.4 | 工具路由（MCP-03）：前缀匹配 + 工具聚合 | §5.19.3 | 2d |
| 10.5 | 工具级授权（MCP-04）：CEL 策略引擎 | §5.19.4 | 2d |
| 10.6 | 连接池管理（MCP-05）：OAuth + 健康检测 | §5.19.1 | 1.5d |
| 10.7 | MCP 审计（MCP-06）：mcp_event + trace_id | §5.19.1 | 1.5d |
| 10.8 | MCP 集成测试（接入 3+ MCP Server） | — | 2d |
| 10.9 | Plugin router 注册 + 测试补全 | — | 1d |
| 10.10 | CI/CD 流水线 | — | 3d |
| 10.11 | 全量回归 + 性能压测 | — | 3d |

**产出**: MCP 网关 + CI/CD + 全量回归

---

## 总览

| Phase | 名称 | 人日 | 优先级 | 依赖 |
|-------|------|------|--------|------|
| 1 | 数据层修复 | 2-4d | 🔴 CRITICAL | — |
| 2 | 安全加固 | 2-3d | 🔴 CRITICAL | Phase 1 |
| 3 | Proxy 增强 | 3-5d | 🟠 HIGH | Phase 1 |
| 4 | i18n + 前端 | 4-6d | 🟠 HIGH | Phase 1 |
| 5 | 条件路由 | 5-8d | 🟡 P1 | Phase 3 |
| 6 | 语义缓存 | 13-21d | 🔴 P0 | Phase 3 |
| 7 | 流式护栏+DLP | 5-8d | 🟡 P1 | Phase 3 |
| 8 | OTel 可观测性 | 7-12d | 🔴 P0 | Phase 3 |
| 9 | 上下文压缩 | 3-5d | 🟢 P2 | Phase 6 |
| 10 | MCP 网关+收尾 | 23-36d | 🟡 P1 | Phase 7,8 |
| **总计** | | **67-108d** | | |

## 执行时间线

```
Week 1-2   → Phase 1 + Phase 2 [并行]
Week 2-4   → Phase 3 + Phase 4 [并行]
Week 5-7   → Phase 6 L1 + Phase 8 [并行，最高ROI]
Week 8-10  → Phase 5 + Phase 7 [并行]
Week 8-12  → Phase 6 L2+L3
Week 11-13 → Phase 9
Week 12-17 → Phase 10
```

## 里程碑

| M | 名称 | 完成标志 | Week |
|---|------|---------|------|
| M1 | 基础加固 | `go test ./...` 全绿 | 2 |
| M2 | 体验完整 | 前端 4 页 + i18n 全覆盖 | 4 |
| M3 | 核心降本 | 精确缓存命中率 ≥ 15% | 6 |
| M4 | 企业可观测 | OTel Span + /metrics 端点 | 7 |
| M5 | 智能路由 | 缓存命中率 ≥ 50% | 10 |
| M6 | 全功能 | MCP 网关可用，CI/CD 就绪 | 17 |

## 验证标准

1. 每个 Phase 完成后 `go test ./...` 全绿
2. Phase 6: 缓存命中率 ≥ 15%（L1），≥ 50%（L1+L2），P50 ≤ 15ms
3. Phase 8: `/metrics` 端点暴露 40+ Prometheus 指标
4. Phase 10: MCP 网关接入 ≥ 3 个 MCP Server
5. 所有新增 Domain 有 handler_test.go 覆盖
6. 数据库 Migration 无报错
