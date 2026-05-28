> **Domain**: `domain/guardrail` — 安全护栏 | **PDD**: §5.11

### 6.12 安全护栏（GRDL）

参考 Portkey Guardrails、LiteLLM Guardrails、SignalVault、@llm-guardrails/core 等行业方案。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| GRDL-01 | **PII 检测与脱敏**：自动检测用户 Prompt/Response 中的 PII（邮箱、手机号、SSN、信用卡、IP），支持 block/redact/warn 三种处理动作 | P0 | 可配置检测粒度 |
| GRDL-02 | **Prompt 注入检测**：检测恶意 Prompt 注入、越狱攻击、提示泄露等安全威胁 | P0 | 对接前/后置检测 |
| GRDL-03 | **内容审核**：对输入/输出内容进行毒性、仇恨言论、暴力、色情等内容安全检测 | P1 | 支持自定义敏感词库 |
| GRDL-04 | **密钥扫描**：自动检测请求中泄露的 API Key、Bearer Token、AWS 凭证、数据库连接串等 | P0 | 防止密钥通过 Prompt 泄露 |
| GRDL-05 | **护栏执行模式**：支持三种模式 —— Enforce（拦截/阻断）、Monitor（告警但放行）、Log（仅记录） | P0 | 可逐步灰度上线 |
| GRDL-06 | **自定义护栏规则**：支持管理员自定义正则、关键词、NLP 分类规则，实现业务级内容管控 | P1 | — |
| GRDL-07 | **不可篡改审计日志**：安全事件日志使用 AES-256-GCM 加密存储，API Key 哈希使用 HMAC-SHA256，支持合规导出 | P0 | — |
| GRDL-08 | **护栏流水线**：支持 Before 护栏（输入检测）和 After 护栏（输出检测）两阶段，每个阶段允许多条规则依次执行 | P0 | — |
| GRDL-09 | **合规报告导出**：一键导出 SOC2、GDPR、HIPAA 格式的安全审计报告（JSON/CSV） | P1 | — |

