> **Domain**: `domain/user` — 用户注册/登录/账号管理 | **PRD**: FastAX-PRD/01-user-auth.md | **参考**: one-api model/user.go + controller/user.go
## 6. 模块详细设计

### 5.1 用户模块 (User Service)

#### 5.1.1 注册流程设计 (含海外用户邮箱注册)

```
┌────────┐     ┌────────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐
│ Client │     │ Gateway    │     │ User-Svc │     │  SMS Svc │     │  DB    │
└───┬────┘     └─────┬──────┘     └────┬─────┘     └────┬─────┘     └───┬────┘
    │                │                  │                │               │
    │  POST /register│                  │                │               │
    │  (支持邮箱/手机) │                  │                │               │
    │────────────────→│                  │                │               │
    │                │── 转发 ──────────→│                │               │
    │                │                  │── 校验验证码 ──→│(或邮件服务)    │
    │                │                  │←── 验证通过 ────│               │
    │                │                  │── 检查重复 ────→│────           │
    │                │                  │                │   查 email    │
    │                │                  │                │   / phone     │
    │                │                  │                │←───           │
    │                │                  │── BCrypt 哈希 ──│               │
    │                │                  │── preferred_language ← Accept-Language │
    │                │                  │── INSERT user ─→│────           │
    │                │                  │                │   写入 DB     │
    │                │                  │←── 成功 ───────│               │
    │                │                  │── 发布事件      │               │
    │                │                  │  registered    │               │
    │                │                  │── 生成 JWT ────│               │
    │                │←── 返回 JWT ─────│                │               │
    │←── 注册成功 ────│                  │                │               │
    │   {token}      │                  │                │               │
```

海外用户邮箱注册 (PRD F-REG-07): 不要求手机号，仅邮箱+密码，通过邮件验证码代替短信验证码。

#### 5.1.2 认证鉴权方案

```
┌─────────────────────────────────────────────────────────────────────┐
│                        JWT 认证方案                                   │
│                                                                     │
│  Access Token:  JWT, 有效期 24h                                     │
│    payload: { user_id, role, level, preferred_language, exp, jti }  │
│                                                                     │
│  Refresh Token: 随机字符串, 有效期 7天, Redis 存储                     │
│                                                                     │
│  刷新流程:                                                           │
│  1. Access Token 过期 → 客户端用 Refresh Token 调用 /api/refresh     │
│  2. 服务端校验 Refresh Token 有效性 + Redis 中存在                     │
│  3. 颁发新的 Access Token + Refresh Token (Rotation)                 │
│  4. 旧 Refresh Token 失效                                            │
│                                                                     │
│  权限模型: RBAC (Role-Based Access Control)                          │
│  Role: guest / user / enterprise / agent / vendor / admin / super_admin │
│                                                                     │
│  语言偏好传递:                                                        │
│  - 请求头 Accept-Language → 网关解析 → 注入 JWT payload              │
│  - 服务端通过 JWT 获取用户语言偏好，用于通知模板选择                    │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.1.3 OAuth 第三方登录

| Provider | 流程 |
|----------|------|
| Google / GitHub / 微信 | OAuth 2.0 Authorization Code Flow |
| 跳转 | `GET /api/auth/oauth/:provider` → 第三方授权页 |
| 回调 | `GET /api/auth/oauth/callback?code=xxx` → JWT |
| 直接登录 | `POST /api/auth/oauth/login` (传入 provider + access_token) |

#### 5.1.4 子账号管理体系

- **企业主账号** → 创建子账号 → 分配额度/权限
- **额度模型**：总量配额制，子账号消耗从企业池扣除
- **权限模型**：可调用的 API 列表、可访问的页面、可操作的按钮
- **隔离级别**：子账号间数据隔离，不可查看对方调用记录

#### 5.1.5 API 端点

**认证相关 (8)**：

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/auth/register` | POST | 无 | 注册 |
| `/api/auth/login` | POST | 无 | 登录 → JWT |
| `/api/auth/refresh` | POST | 无 | 刷新 Token |
| `/api/auth/logout` | POST | JWT | 登出 |
| `/api/auth/send-code` | POST | 无 | 发送验证码 |
| `/api/auth/reset-password` | POST | 无 | 重置密码 |
| `/api/auth/oauth/:provider` | GET | 无 | OAuth 登录跳转 |
| `/api/auth/oauth/callback` | GET | 无 | OAuth 回调 |
| `/api/auth/oauth/login` | POST | 无 | OAuth 直接登录 |

**用户相关 (3)**：

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/user/me` | GET | JWT | 获取当前用户信息 |
| `/api/user/language` | PUT | JWT | 更新语言偏好 |

**管理后台 (6)**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/users` | GET | 用户列表 |
| `/api/admin/users/:id` | GET | 用户详情 |
| `/api/admin/users/:id/status` | PUT | 冻结/解冻 |
| `/api/admin/users/:id/level` | PUT | 修改等级 |
| `/api/admin/system/admins` | GET | 管理员列表 |
| `/api/admin/system/admins` | POST | 添加管理员 |

#### 5.1.6 Service 方法

```go
// 认证
Register(req *RegisterRequest) (*LoginResponse, error)
Login(req *LoginRequest) (*LoginResponse, error)
RefreshToken(refreshToken string) (*LoginResponse, error)
Logout(userID uint, refreshToken string) error
SendCode(phoneOrEmail string) error
ResetPassword(req *ResetPasswordRequest) error

// OAuth
GetOAuthRedirectURL(provider, callbackURL string) (string, error)
OAuthLogin(req *OAuthLoginRequest) (*LoginResponse, error)

// 用户信息
GetUser(userID uint) (*UserResponse, error)
UpdateLanguage(userID uint, language string) error

// 管理端
ListUsers(page, pageSize int, keyword string) (*UserListResponse, error)
GetUserDetail(id uint) (*UserDetailResponse, error)
SetUserStatus(id uint, status int) error
SetUserLevel(id uint, level string) error
ListAdmins() ([]UserResponse, error)
CreateAdmin(username, password, email, role string) (*UserResponse, error)
```
