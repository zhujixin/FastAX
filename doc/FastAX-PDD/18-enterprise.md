> **Domain**: `domain/enterprise` — 企业功能 | **PRD**: FastAX-PRD/13-enterprise.md
### 5.15 企业功能 (PRD §6.16 ENT)

| 功能 | 设计要点 | 优先级 |
|------|---------|--------|
| SSO/SAML/OIDC | 内嵌 SAML2 服务提供方 + OIDC 客户端, 支持 Okta/Azure AD/Google Workspace | P2 |
| 团队/项目隔离 | team 管理, 独立计费、独立配额、数据隔离 | P1 |
| 子账号管理 | 企业主账号创建子账号、分配额度/权限、用量监控 | P0 |
| 审计导出 | SOC2/GDPR/HIPAA 格式, 带合规头信息, 一键导出 JSON/CSV | P1 |
| 预付费套餐 | 预购额度 + 用量承诺折扣, 支持自动续费 | P1 |
| 模型白名单 | 企业管理员限制团队可用模型列表 | P1 |
| 数据驻留控制 | 指定区域偏好 (仅国内/仅海外/指定区域) | P2 |

#### 5.15.1 API 端点

**用户端 (JWT)**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/enterprise/sub-accounts` | POST | 创建子账号 (email+password+quota+permissions) |
| `/api/enterprise/sub-accounts` | GET | 子账号列表 |
| `/api/enterprise/sub-accounts/:id/status` | PUT | 启用/禁用子账号 |
| `/api/enterprise/sub-accounts/:id/quota` | PUT | 更新子账号配额 |
| `/api/enterprise/usage` | GET | 企业总体用量 |
| `/api/enterprise/sub-accounts/:id/usage` | GET | 子账号用量详情 |

**管理后台 (Admin)**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/sso/config` | GET | SSO 配置 (SAML/OIDC 端点) |
| `/api/admin/sso/config` | PUT | 更新 SSO 配置 |
| `/api/admin/teams` | GET | 团队列表 |
| `/api/admin/teams` | POST | 创建团队 |
| `/api/admin/teams/:id` | PUT | 编辑团队 |
| `/api/admin/teams/:id` | DELETE | 删除团队 |
| `/api/admin/audit/export` | GET | 审计日志导出 |

#### 5.15.2 Service 方法

```go
// 子账号管理
CreateSubAccount(parentID, req *SubAccountRequest) (*SubAccountResponse, error)
ListSubAccounts(parentID) ([]SubAccountResponse, error)
SetSubAccountStatus(id, parentID, status) error
UpdateQuota(id, parentID, quota) error

// 用量统计
GetEnterpriseUsage(parentID, period) (*UsageStats, error)
GetSubAccountUsage(subAccountID, period) (*UsageStats, error)

// SSO 配置
GetSSOConfig() (*SSOConfig, error)
UpdateSSOConfig(req *SSOConfig) error

// 团队管理
CreateTeam(parentID, req *CreateTeamRequest) (*Team, error)
ListTeams(parentID) ([]Team, error)
GetTeam(id) (*Team, error)
UpdateTeam(id, req *UpdateTeamRequest) error
DeleteTeam(id) error
```

> **注意**: SSO 配置和 Teams 当前使用内存存储 (包级变量)，未来计划迁移到数据库。
