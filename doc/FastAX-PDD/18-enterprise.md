> **Domain**: `domain/enterprise` — 企业功能 | **PRD**: FastAX-PRD/13-enterprise.md
### 5.15 企业功能 (PRD §6.16 ENT)

| 功能 | 设计要点 | 优先级 |
|------|---------|--------|
| SSO/SAML/OIDC | 内嵌 SAML2 服务提供方 + OIDC 客户端, 支持 Okta/Azure AD/Google Workspace | P2 |
| 团队/项目隔离 | team + project 两级, 独立计费、独立配额、数据隔离 | P1 |
| 审计导出 | SOC2/GDPR/HIPAA 格式, 带合规头信息, 一键导出 JSON/CSV | P1 |
| 预付费套餐 | 预购额度 + 用量承诺折扣, 支持自动续费 | P1 |
| 模型白名单 | 企业管理员限制团队可用模型列表 | P1 |
| 数据驻留控制 | 指定区域偏好 (仅国内/仅海外/指定区域) | P2 |

