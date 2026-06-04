> **Domain**: `domain/mcp` — MCP 网关 | **PDD**: §5.19 | **新增于**: PRD v3.1

### 6.20 MCP 网关（MCP）

参考 Envoy AI Gateway MCP、Portkey Agent Gateway、Docker MCP Gateway、Gravitee MCP Proxy、Zuplo MCP Gateway 等行业方案。

Model Context Protocol (MCP) 已成为 AI Agent 与外部工具之间的事实通信标准。2024 年 11 月 Anthropic 首创规范，2025-2026 年所有主流 AI 基础设施厂商均推出 MCP Gateway。MCP 网关是 FastAX 从「API 代理」扩展到「Agent 基础设施」的战略性新业务方向。

```
AI Agent → MCP Gateway (FastAX) → MCP Server A (GitHub API)
                                  → MCP Server B (PostgreSQL DB)
                                  → MCP Server C (Slack)
                                  → ...
```

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| MCP-01 | **统一 MCP 端点**：提供单一 MCP 入口端点（`/mcp`），聚合多个 MCP Server 的工具列表，Agent 通过一个端点发现和调用所有工具 | P1 | 支持 SSE 和 Streamable HTTP 两种传输协议；工具名称自动添加前缀防冲突（如 `github__issue_read`） |
| MCP-02 | **传输桥接**：支持 stdio ↔ SSE ↔ Streamable HTTP 三种传输协议的统一桥接，Agent 使用 HTTP/SSE 连接网关，网关通过 stdio 连接本地 MCP Server | P1 | 参考 Envoy AI Gateway v0.5 MCP 传输桥接 |
| MCP-03 | **工具路由**：按 tool name 前缀将请求路由到对应的 MCP Server，支持负载均衡（多实例部署时）和故障转移 | P1 | 与现有 ROUTE 引擎复用路由基础设施 |
| MCP-04 | **工具级授权 (RBAC)**：基于 CEL 表达式或 RBAC 策略，控制每个用户/Agent 可调用的工具范围。示例策略：`user.role == "admin" \|\| tool.name.startsWith("read_")` | P2 | 参考 Envoy AI Gateway CEL 策略引擎；可与 ENT-02 团队隔离联动 |
| MCP-05 | **MCP 连接池管理**：管理到上游 MCP Server 的连接池（OAuth Token 自动刷新、会话超时清理、连接健康检测、空闲回收） | P2 | 降低连接管理复杂度，提高稳定性 |
| MCP-06 | **MCP 审计追踪**：完整记录所有 MCP 事件（工具列表获取、工具调用、调用者身份、授权决策、响应状态），关联 trace_id，支持导出审计报告 | P2 | 与 OBSV-01 全链路追踪联动 |

**关键设计决策**：

| 决策 | 要点 | 参考 |
|------|------|------|
| **协议版本** | 优先支持 MCP 规范最新稳定版（2025+ Streamable HTTP 传输），SSE 作为兼容备选 | Anthropic MCP Spec |
| **工具前缀策略** | 采用 `namespace__tool_name` 双下划线防冲突（参考 Envoy AI Gateway），也可配置自定义前缀 | Envoy AI Gateway tool prefix |
| **认证集成** | 复用现有 JWT 认证中间件 + API Key 体系；MCP Server 凭证加密存储（AES-256-GCM） | 与 BYOK-01 Key 管理复用 |
| **传输选择** | **Agent→网关**：Streamable HTTP（推荐）或 SSE；**网关→MCP Server**：stdio（本地）、SSE（远程）、Streamable HTTP（远程） | — |
| **安全模型** | 最小权限原则：Agent 默认无工具访问权限，需显式授权；MCP Server 凭证加密隔离存储 | — |

**与现有模块的关系**：

| 现有模块 | 复用方式 |
|---------|---------|
| `domain/proxy` | 复用 Adaptor 模式统一对接各家 MCP Server |
| `domain/guardrail` | 护栏流水线可扩展到 MCP 工具调用的输入/输出检测 |
| `domain/enterprise` | 团队隔离 + RBAC 扩展到工具级授权 |
| `domain/log` | 调用日志扩展到 mcp_event 记录 |
| `domain/user` | JWT 认证 + API Key 体系复用 |

**实施路径**：

| 阶段 | 内容 | 预计人日 |
|------|------|---------|
| Phase 1 | 统一 MCP 端点 + 工具聚合 + 基础路由（MCP-01/02/03） | 8-12d |
| Phase 2 | 工具级授权 + 连接池管理（MCP-04/05） | 5-8d |
| Phase 3 | MCP 审计 + 第三方 MCP Server 市场（MCP-06 + 扩展） | 5-8d |

