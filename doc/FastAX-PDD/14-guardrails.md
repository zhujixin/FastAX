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
guardrail_rule 表:
  rule_id, name, stage(before/after), type(pii/injection/secret/content),
  action(block/redact/warn), conditions(JSON), enabled, priority

执行模式:
  Enforce (拦截阻断) — 默认
  Monitor (告警放行) — 灰度测试
  Log (仅记录)       — 合规审计

性能要求:
  PII 检测 ≤ 50ms, 注入检测 ≤ 100ms
  护栏故障不阻塞主请求 (bypass 熔断)
```

