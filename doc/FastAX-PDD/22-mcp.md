> **Domain**: `domain/mcp` — MCP 网关 | **PRD**: FastAX-PRD/24-mcp-gateway.md
### 5.19 MCP 网关 (PRD §6.20 MCP)

#### 5.19.1 统一 MCP 端点架构

```
AI Agent (Client)
    │
    │ SSE / Streamable HTTP
    ▼
┌──────────────────────────────────────┐
│        FastAX MCP Gateway            │
│  POST /mcp  (Streamable HTTP)        │
│  GET  /mcp  (SSE)                    │
│                                      │
│  ┌──────────┐  ┌──────────────────┐  │
│  │ Auth     │→ │ Tool Router      │  │
│  │ (JWT/Key)│  │ (前缀匹配分发)    │  │
│  └──────────┘  └──────┬───────────┘  │
│                       │              │
│         ┌─────────────┼──────────┐   │
│         ▼             ▼          ▼   │
│  ┌──────────┐ ┌──────────┐ ┌──────┐  │
│  │github__* │ │slack__*  │ │db__* │  │
│  │→ GitHub  │ │→ Slack   │ │→ PG  │  │
│  │  MCP Svr │ │  MCP Svr │ │ MCP  │  │
│  └──────────┘ └──────────┘ └──────┘  │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ Tool Policy Engine (CEL/RBAC)  │  │
│  │ MCP Audit Logger               │  │
│  │ Connection Pool Manager        │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

#### 5.19.2 传输桥接

| Agent → Gateway | Gateway → MCP Server | 适用场景 |
|-----------------|---------------------|---------|
| Streamable HTTP (推荐) | Streamable HTTP | 远程 MCP Server |
| SSE | SSE | 远程 MCP Server (兼容) |
| — | stdio | 本地 MCP Server (同机部署) |

桥接逻辑：
1. Agent 通过 HTTP/SSE 连接 Gateway
2. Gateway 根据 MCP Server 注册的 transport 类型选择连接方式
3. stdio Server: Gateway 启动子进程，通过 stdin/stdout 通信
4. HTTP Server: Gateway 转发 HTTP 请求，管理 OAuth Token 刷新

#### 5.19.3 工具路由

```
工具命名规范: {namespace}__{tool_name}
  示例: github__issue_read, slack__send_message, postgres__query

路由流程:
  1. Agent 发送 tools/list → Gateway 聚合所有已注册 MCP Server 的工具列表
     (自动添加 namespace 前缀)
  2. Agent 发送 tools/call {name: "github__issue_read"}
  3. Gateway 解析 namespace → 查找对应的 MCP Server 连接
  4. Gateway 移除前缀，将 tools/call {name: "issue_read"} 转发到 GitHub MCP Server
  5. Gateway 返回结果给 Agent
```

#### 5.19.4 工具级授权 (CEL 策略)

```cel
// 示例策略
user.role == "admin" || tool.name.startsWith("read_")

// 策略配置
mcp_tool_policies 表:
  id, server_id, tool_pattern, policy_expr (CEL), action (allow/deny), priority
```

#### 5.19.5 API 端点

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/mcp` | POST | JWT/API Key | MCP Streamable HTTP 端点 (tools/list, tools/call) |
| `/mcp` | GET | JWT/API Key | MCP SSE 端点 |
| `/api/admin/mcp/servers` | GET | Admin | MCP Server 列表 |
| `/api/admin/mcp/servers` | POST | Admin | 注册 MCP Server |
| `/api/admin/mcp/servers/:id` | PUT | Admin | 更新 MCP Server 配置 |
| `/api/admin/mcp/servers/:id` | DELETE | Admin | 删除 MCP Server |
| `/api/admin/mcp/servers/:id/tools` | GET | Admin | 查看 Server 提供的工具列表 |
| `/api/admin/mcp/servers/:id/test` | POST | Admin | 测试 MCP Server 连通性 |
| `/api/admin/mcp/policies` | GET | Admin | 工具授权策略列表 |
| `/api/admin/mcp/policies` | POST | Admin | 创建授权策略 |
| `/api/admin/mcp/logs` | GET | Admin | MCP 事件审计日志 |

#### 5.19.6 Service 方法

```go
// MCPService — MCP 网关服务
type MCPService struct {
    servers   map[string]*MCPServerConn  // namespace → connection
    router    *ToolRouter                // 工具路由分发
    policy    *PolicyEngine              // CEL 策略引擎
    audit     *AuditLogger               // 审计日志
}

// MCP 协议处理
func (s *MCPService) ListTools(ctx context.Context, userInfo *UserInfo) ([]*Tool, error)
func (s *MCPService) CallTool(ctx context.Context, userInfo *UserInfo, toolName string, args map[string]interface{}) (*ToolResult, error)

// Server 管理
func (s *MCPService) RegisterServer(ctx context.Context, req *RegisterServerRequest) (*MCPServer, error)
func (s *MCPService) UpdateServer(ctx context.Context, id uint, req *UpdateServerRequest) (*MCPServer, error)
func (s *MCPService) DeleteServer(ctx context.Context, id uint) error
func (s *MCPService) ListServers(ctx context.Context) ([]*MCPServer, error)
func (s *MCPService) TestConnection(ctx context.Context, id uint) (*TestResult, error)
func (s *MCPService) GetServerTools(ctx context.Context, id uint) ([]*Tool, error)

// 授权策略管理
func (s *MCPService) CreatePolicy(ctx context.Context, req *PolicyRequest) (*MCPToolPolicy, error)
func (s *MCPService) ListPolicies(ctx context.Context) ([]*MCPToolPolicy, error)
func (s *MCPService) EvaluatePolicy(ctx context.Context, userInfo *UserInfo, toolName string) (bool, error)

// 审计日志
func (s *MCPService) ListAuditLogs(ctx context.Context, filter *AuditFilter) ([]*MCPEventLog, error)
func (s *MCPService) LogEvent(ctx context.Context, event *MCPEvent) error
```

#### 5.19.7 数据表

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `mcp_servers` | MCP Server 注册表 | namespace, name, transport (stdio/sse/http), endpoint_url, command (stdio), auth_type, auth_config (JSON encrypted), status, health_check_interval |
| `mcp_tool_policies` | 工具授权策略 | server_id, tool_pattern, policy_expr (CEL), action (allow/deny), priority, enabled |
| `mcp_event_logs` | MCP 事件审计日志 | trace_id, server_id, tool_name, user_id, action (list/call), request (JSON), response (JSON truncated), status, duration_ms, created_at |

#### 5.19.8 与现有模块复用关系

| 现有模块 | 复用方式 |
|---------|---------|
| `domain/proxy` | Adaptor 模式统一对接各家 MCP Server；路由引擎复用于工具分发 |
| `domain/guardrail` | 护栏流水线扩展到 MCP tools/call 的输入/输出检测 |
| `domain/enterprise` | ENT-02 团队隔离扩展到工具级授权；ENT-07 模型白名单模式参考 |
| `domain/user` | JWT 认证 + API Key 体系直接复用 |
| `domain/observability` | MCP 事件写入 OTel Span Event + trace_id 传播 |

---
## 相关模块

| 关系 | 模块 | 说明 |
|------|------|------|
| 依赖 | [代理模块](06-token-proxy-vendor.md) | 复用 Adaptor 模式 + 路由引擎 |
| 依赖 | [用户认证](05-user.md) | JWT + API Key 认证 |
| 依赖 | [安全护栏](14-guardrails.md) | 工具调用前后检测 |
| 关联 | [可观测性](21-otel.md) | MCP 事件 OTel Span Event |
