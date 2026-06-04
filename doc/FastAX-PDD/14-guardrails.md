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
