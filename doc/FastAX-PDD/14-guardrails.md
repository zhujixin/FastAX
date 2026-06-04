> **Domain**: `domain/guardrail` — 安全护栏 | **PRD**: FastAX-PRD/09-guardrails.md
### 5.11 安全护栏 (PRD §6.12 GRDL)

#### 5.11.1 护栏流水线架构

```
请求进入 → [Before 护栏] → 转发到供应商 → [After 护栏] → 返回给用户
               │                              │
               ├── PII 检测                   ├── 内容审核
               ├── Prompt 注入检测             ├── PII 扫描 (响应)
               ├── 密钥扫描                    └── 合规检测
               └── 内容审核 (输入)
```

| 阶段 | 检测项 | 处理动作 | 参考实现 |
|------|--------|---------|---------|
| Before | PII (邮箱/手机/SSN/信用卡) | block / redact / warn | 正则 + NLP 命名实体识别 |
| Before | Prompt 注入 | block | 模型分类器 + 规则 |
| Before | 密钥扫描 (API Key/Token) | redact | 正则匹配 known patterns |
| Before | 内容审核 (涉政/暴恐/色情) | block | 第三方 API / 模型分类器 |
| After | PII 泄露 (响应中) | redact | 同 Before PII 检测 |
| After | 内容合规 (响应) | block | 同 Before 内容审核 |

#### 5.11.2 护栏规则配置

```
guardrail_rules 表:
  id, name, stage(before/after), type(pii/injection/secret/content),
  action(block/redact/warn), conditions(JSON), enabled, priority

执行模式:
  Enforce (拦截阻断) — 默认
  Monitor (告警放行) — 灰度测试
  Log (仅记录)       — 合规审计

性能要求:
  PII 检测 ≤ 50ms, 注入检测 ≤ 100ms
  护栏故障不阻塞主请求 (bypass 熔断)
```

#### 5.11.3 API 端点

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/guardrails/rules` | GET | Admin | 规则列表 |
| `/api/admin/guardrails/rules` | POST | Admin | 创建规则 |
| `/api/admin/guardrails/rules/:id` | PUT | Admin | 更新规则 |
| `/api/admin/guardrails/rules/:id` | DELETE | Admin | 删除规则 |
| `/api/admin/guardrails/rules/:id/enabled` | PUT | Admin | 启用/禁用规则 |
| `/api/admin/guardrails/logs` | GET | Admin | 检测日志查询 (按 trace_id/user_id/stage 筛选) |
| `/api/admin/guardrails/detect` | POST | Admin | 实时检测测试 (提交文本，返回检测结果) |
| `/api/admin/guardrails/config` | PUT | Admin | 全局配置 (模式切换/启用开关) |

#### 5.11.4 Service 方法

```go
// 规则管理
CreateRule(req *RuleRequest) (*GuardrailRule, error)
ListRules(stage string) ([]GuardrailRule, error)
SetRuleEnabled(id uint, enabled bool) error
UpdateRule(id uint, req *RuleRequest) (*GuardrailRule, error)
DeleteRule(id uint) error

// 检测引擎
Detect(req *DetectRequest) (*DetectResult, error)
  // DetectRequest: { text, stage, types[] }
  // DetectResult: { passed, findings[{type, entity, action}] }

// 日志查询
ListLogs(traceID string, userID uint, stage string) ([]GuardrailLog, error)

// 配置管理
UpdateGlobalConfig(mode string, enabled bool) error
GetConfig() map[string]interface{}
```

#### 5.11.5 流式护栏 + DLP (PRD §6.12 GRDL-10~14) 🆕 v3.1

| 需求ID | 功能 | 设计要点 | 优先级 |
|--------|------|---------|--------|
| GRDL-10 | 流式护栏 | 在 SSE 流处理循环中逐 chunk 累积文本，每 N chunks（可配，默认 5）触发一次增量检测。检测到违规时：`block`=立即 `close(stream)` + 发送 `[DONE]`；`redact`=将敏感片段替换为 `***` 后继续发送 | P1 |
| GRDL-11 | DLP 数据防泄露 | 双阶段扫描：Before 护栏扫描 Prompt 中的敏感文档内容（身份证号、银行卡号、内部项目代号）；After 护栏扫描 LLM 输出中的敏感信息泄露。使用正则+关键词库+NER 模型三层检测 | P1 |
| GRDL-12 | 自定义正则替换 | `guardrail_rules` 表新增 `rule_type=regex_replace`，`conditions` JSON 存储 `{pattern, replacement}`。在护栏流水线中优先执行 regex_replace 规则再执行检测规则 | P1 |
| GRDL-13 | 工具级拦截 | 扩展护栏引擎支持 `target=function_call` / `target=code_interpreter` / `target=mcp_tool` 三种新检测目标。在 OpenAI function_call 和 MCP tools/call 请求中提取工具名，匹配 `guardrail_rules` 中的 `tool_pattern` 字段进行阻断 | P2 |
| GRDL-14 | 第三方护栏集成 | 定义 `GuardrailAdapter` 接口：`Detect(ctx, text, config) (*DetectResult, error)`。管理员通过 API 注册外部护栏引擎 URL+API Key，平台在护栏流水线中按优先级依次调用内部规则→第三方引擎 | P2 |

**数据表变更**：
```sql
ALTER TABLE guardrail_rules ADD COLUMN target VARCHAR(50) DEFAULT 'prompt';
  -- prompt / response / function_call / code_interpreter / mcp_tool
ALTER TABLE guardrail_rules ADD COLUMN rule_type VARCHAR(50) DEFAULT 'detect';
  -- detect / regex_replace
```

**API 端点新增**：
| `/api/admin/guardrails/adaptors` | GET | Admin | 第三方护栏适配器列表 |
| `/api/admin/guardrails/adaptors` | POST | Admin | 注册第三方护栏引擎 |
