## 8. 接口详细设计

### 7.1 用户端 API (OpenAI 兼容协议)

#### `POST /v1/chat/completions`

**请求体**:
```json
{
  "model": "gpt-4",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": true,
  "temperature": 0.7,
  "max_tokens": 2048
}
```

**认证**: `Authorization: Bearer <user_api_key>`

**流式响应**:
```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant"},"index":0}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"},"index":0}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"!"},"index":0}]}

data: [DONE]
```

**非流式响应**:
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1680000000,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 8,
    "total_tokens": 33
  }
}
```

### 7.2 业务 API 详细设计

#### 7.2.1 认证相关

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/auth/register` | POST | 无 | 注册 (手机/邮箱+验证码) |
| `/api/auth/login` | POST | 无 | 登录，返回 JWT (含语言偏好) |
| `/api/auth/refresh` | POST | Refresh Token | 刷新 Access Token |
| `/api/auth/logout` | POST | JWT | 登出，销毁 Token |
| `/api/auth/send-code` | POST | 无 | 发送验证码 (短信/邮件，按语言) |
| `/api/auth/reset-password` | POST | 无 | 重置密码 |
| `/api/auth/oauth/{provider}` | GET | 无 | OAuth 登录跳转 |
| `/api/auth/oauth/callback` | GET | 无 | OAuth 回调 |

**`POST /api/auth/login`**:
```json
// Request
{
  "account": "13800138000",
  "password": "abc123!@#",
  "device_info": {"fingerprint": "...", "user_agent": "..."}
}

// Response
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rft_xxxxxxxxx",
  "expires_in": 86400,
  "user": {
    "id": 1,
    "username": "zhangsan",
    "role": "user",
    "level": "normal",
    "preferred_language": "zh-CN"
  }
}
```

#### 7.2.2 Token 相关

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/tokens/products` | GET | 可选 | Token 商品列表 (含多语言字段) |
| `/api/tokens/products/{id}` | GET | 可选 | 商品详情 (含多语言字段) |
| `/api/tokens/my` | GET | JWT | 我持有的 Token |
| `/api/tokens/buy` | POST | JWT | 购买 Token |
| `/api/tokens/transfer` | POST | JWT | 转让 Token |
| `/api/tokens/extract` | POST | JWT | 提取 Token |
| `/api/tokens/my/usage` | GET | JWT | 我的使用记录 |

**`POST /api/tokens/buy`**:
```json
// Request
{
  "product_id": 1,
  "quantity": 100,
  "payment_method": "wechat"
}

// Response
{
  "order_no": "ORD202605270001",
  "amount": 199.00,
  "status": "pending",
  "payment_url": "https://pay.wechat.com/..."
}
```

**`GET /api/tokens/my`**:
```json
// Response
{
  "items": [
    {
      "product_id": 1,
      "product_name": "GPT-4 Token 包",
      "total": 1000000,
      "used": 350000,
      "remaining": 650000,
      "expires_at": "2026-08-27T00:00:00Z",
      "status": "active"
    }
  ],
  "total": 1
}
```

#### 7.2.3 订单相关

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/orders` | GET | JWT | 订单列表 |
| `/api/orders/{id}` | GET | JWT | 订单详情 |
| `/api/orders/{id}/refund` | POST | JWT | 申请退款 |

#### 7.2.4 统计相关

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/stats/usage` | GET | JWT | 用量统计 |
| `/api/stats/consumption` | GET | JWT | 消费统计 |
| `/api/stats/bills` | GET | JWT | 账单明细 |
| `/api/stats/summary` | GET | JWT | 控制台总览 |

#### 7.2.5 通知相关

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/notifications` | GET | JWT | 通知列表 (当前语言) |
| `/api/notifications/unread-count` | GET | JWT | 未读数 |
| `/api/notifications/{id}/read` | PUT | JWT | 标记已读 |
| `/api/notifications/read-all` | PUT | JWT | 全部已读 |

#### 7.2.6 新增：多语言相关 (PRD §6.9)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/i18n/languages` | GET | 无 | 获取可用语言列表 (LANG-04) |
| `/api/i18n/translations/{locale}` | GET | 无 | 获取翻译文件 (CDN 回源) |
| `/api/user/language` | PUT | JWT | 更新用户语言偏好 |

**`GET /api/i18n/languages`**:
```json
// Response
{
  "languages": [
    {"locale": "zh-CN", "name": "中文", "is_default": true},
    {"locale": "en", "name": "English", "is_default": false}
  ],
  "default_locale": "zh-CN"
}
```

#### 7.2.7 新增：供应商服务 API (PRD §6.2.6)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/vendor/register` | POST | JWT | 提交入驻申请 (SUP-01) |
| `/api/vendor/profile` | GET | JWT(供应商) | 查看店铺信息 (SUP-03) |
| `/api/vendor/profile` | PUT | JWT(供应商) | 更新店铺信息 |
| `/api/vendor/products` | GET | JWT(供应商) | 商品列表 |
| `/api/vendor/products` | POST | JWT(供应商) | 创建商品 (SUP-04) |
| `/api/vendor/products/{id}` | PUT | JWT(供应商) | 更新商品 |
| `/api/vendor/products/{id}/price` | PUT | JWT(供应商) | 调价 (SUP-05) |
| `/api/vendor/sales` | GET | JWT(供应商) | 销售看板 (SUP-09) |
| `/api/vendor/settlements` | GET | JWT(供应商) | 结算单列表 (SUP-10) |
| `/api/vendor/settlements/{id}/confirm` | POST | JWT(供应商) | 确认结算单 |
| `/api/vendor/settlements/{id}/withdraw` | POST | JWT(供应商) | 申请提现 |

**`POST /api/vendor/register`**:
```json
// Request
{
  "company_name": "AI Model Inc.",
  "contact_name": "John Doe",
  "contact_email": "john@example.com",
  "contact_phone": "+1-555-0123",
  "business_license": "https://oss.example.com/license.pdf",
  "api_base_url": "https://api.example.com/v1",
  "api_auth_type": "api_key",
  "agreed_to_terms": true
}

// Response
{
  "vendor_id": 1,
  "status": "pending",
  "message": "入驻申请已提交，等待平台审核"
}
```

#### 7.2.8 新增：管理后台—供应商管理

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/vendors` | GET | 管理员 | 供应商列表 (SUP-02) |
| `/api/admin/vendors/{id}` | GET | 管理员 | 供应商详情 |
| `/api/admin/vendors/{id}/approve` | POST | 管理员 | 审核通过 |
| `/api/admin/vendors/{id}/reject` | POST | 管理员 | 审核驳回 |
| `/api/admin/vendors/{id}/suspend` | POST | 管理员 | 冻结供应商 (SUP-14) |
| `/api/admin/vendor-commission-rates/{id}` | PUT | 管理员 | 配置佣金比例 (SUP-12) |
| `/api/admin/vendors/{id}/products` | GET | 管理员 | 供应商商品列表 |
| `/api/admin/vendors/{id}/products/{pid}/approve` | POST | 管理员 | 商品合规审核通过 |
| `/api/admin/vendors/{id}/settlements` | GET | 管理员 | 供应商结算记录 |

### 7.3 管理后台 API

| 分组 | 接口 | 说明 |
|------|------|------|
| **Dashboard** | `GET /api/admin/dashboard/summary` | 核心数据总览 |
| | `GET /api/admin/dashboard/charts` | 趋势图表数据 |
| **用户管理** | `GET /api/admin/users` | 用户列表 (搜索/筛选) |
| | `GET /api/admin/users/{id}` | 用户详情 |
| | `PUT /api/admin/users/{id}/status` | 冻结/解冻用户 |
| | `PUT /api/admin/users/{id}/level` | 修改用户等级 |
| **Token 管理** | `GET /api/admin/suppliers` | 供应商列表 |
| | `POST/PUT /api/admin/suppliers` | 新增/编辑供应商 |
| | `GET /api/admin/channels` | 渠道列表 |
| | `PUT /api/admin/channels/{id}/status` | 启用/禁用渠道 |
| | `PUT /api/admin/channels/{id}/priority` | 调整优先级 |
| | `GET /api/admin/products` | Token 商品列表 |
| | `POST/PUT /api/admin/products` | 新增/编辑商品 |
| | `PUT /api/admin/products/{id}/price` | 调价 |
| **交易管理** | `GET /api/admin/orders` | 订单列表 |
| | `GET /api/admin/orders/{id}` | 订单详情 |
| | `POST /api/admin/orders/{id}/refund` | 审核退款 |
| | `GET /api/admin/reports/daily` | 日报表 |
| | `GET /api/admin/reports/monthly` | 月报表 |
| **风控管理** | `GET /api/admin/risk/events` | 风控事件列表 |
| | `PUT /api/admin/risk/events/{id}` | 处理风控事件 |
| | `GET /api/admin/risk/rules` | 规则列表 |
| | `POST/PUT /api/admin/risk/rules` | 新增/编辑规则 |
| | `GET /api/admin/risk/blacklist` | 黑名单 |
| | `POST /api/admin/risk/blacklist` | 添加黑名单 |
| **系统管理** | `GET /api/admin/system/config` | 系统配置 |
| | `PUT /api/admin/system/config` | 更新配置 |
| | `GET /api/admin/system/admins` | 管理员列表 |
| | `POST /api/admin/system/admins` | 添加管理员 |
| | `GET /api/admin/system/logs` | 操作日志 |
| **多语言配置** | `GET /api/admin/i18n/languages` | 语言列表 (F-ADM-08) |
| | `PUT /api/admin/i18n/languages/{id}` | 启用/禁用语种 |
| | `PUT /api/admin/i18n/default` | 设置默认语言 |
| **供应商管理** | `GET /api/admin/vendors` | 入驻供应商列表 |
| | `GET /api/admin/vendors/{id}` | 供应商详情 |
| | `POST /api/admin/vendors/{id}/approve` | 审核通过 |
| | `POST /api/admin/vendors/{id}/reject` | 审核驳回 |
| | `PUT /api/admin/vendor-commission-rates/{id}` | 配置佣金比例 |

### 7.4 错误码设计

| HTTP 状态码 | 业务码 | 说明 |
|-------------|--------|------|
| 200 | 0 | 成功 |
| 400 | 1001 | 请求参数错误 |
| 400 | 1002 | 验证码错误或过期 |
| 401 | 2001 | Token 过期或无效 |
| 401 | 2002 | 账号被冻结 |
| 403 | 3001 | 权限不足 |
| 403 | 3002 | 接口限流 |
| 404 | 4001 | 资源不存在 |
| 409 | 5001 | 重复操作 (如重复订单) |
| 422 | 6001 | Token 余额不足 |
| 422 | 6002 | Token 已过期 |
| 422 | 6003 | 超出购买限额 |
| 429 | 7001 | 请求过于频繁 |
| 500 | 9001 | 系统内部错误 |
| 503 | 9002 | 服务暂不可用 |

**统一响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "trace_id": "txn_xxxxxxxxxxxx"
}
```

**多语言错误消息 (PRD LANG-05-02)**:
```json
// Accept-Language: en
{
  "code": 6001,
  "message": "Insufficient token balance",
  "data": null,
  "trace_id": "txn_xxxxxxxxxxxx"
}

// Accept-Language: ja
{
  "code": 6001,
  "message": "トークン残高が不足しています",
  "data": null,
  "trace_id": "txn_xxxxxxxxxxxx"
}
```

### 7.6 新增 v3.0 API 端点

#### 7.6.1 多协议原生 (PROTO)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/messages` | POST | JWT | Anthropic Messages API |
| `/v1/rerank` | POST | JWT | Rerank 统一 API |
| `/models/:variant` | GET | JWT | 模型变体详情 |

#### 7.6.2 安全护栏 (GRDL)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/guardrails/rules` | GET/POST | Admin | 规则列表/创建 |
| `/api/admin/guardrails/rules/:id` | PUT/DELETE | Admin | 更新/删除规则 |
| `/api/admin/guardrails/logs` | GET | Admin | 检测日志查询 |
| `/api/admin/guardrails/config` | PUT | Admin | 全局配置 |

#### 7.6.3 BYOK (自带 Key)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/byok/keys` | GET/POST | JWT | 用户 Key 列表/添加 |
| `/api/byok/keys/:id` | PUT/DELETE | JWT | 更新/删除 Key |
| `/api/byok/usage` | GET | JWT | BYOK 用量统计 |
| `/api/byok/preference` | PUT | JWT | 路由优先级配置 |

#### 7.6.4 成本优化 (COST)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/cache/stats` | GET | Admin | 语义缓存统计 |
| `/api/cache/config` | PUT | Admin | 缓存策略配置 |
| `/api/user/budget` | GET/PUT | JWT | 用户预算设置 |
| `/api/user/cost-alerts` | GET/PUT | JWT | 成本告警配置 |

#### 7.6.5 模型市场 (MKT)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/models/compare` | POST | JWT | 模型多维度对比 |
| `/api/providers/health` | GET | 无 | 健康面板 (公开) |
| `/api/models/benchmarks` | GET | JWT | 基准测试数据 |

#### 7.6.6 多模态 (MEDIA)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/images/generations` | POST | JWT | 图片生成 |
| `/v1/audio/speech` | POST | JWT | 语音合成 TTS |
| `/v1/audio/transcriptions` | POST | JWT | 语音转文本 STT |
| `/v1/video/generations` | POST | JWT | 视频生成 (P2) |

#### 7.6.7 企业功能 (ENT)

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/sso/config` | GET/PUT | Admin | SSO 配置 |
| `/api/admin/teams` | GET/POST | Admin | 团队管理 |
| `/api/admin/teams/:id` | PUT/DELETE | Admin | 团队编辑/删除 |
| `/api/admin/audit/export` | GET | Admin | 审计日志导出 |

### 7.5 接口通用规范

| 规范 | 内容 |
|------|------|
| 协议 | HTTPS (TLS 1.3) |
| 请求头 | `Content-Type: application/json` |
| 认证 | `Authorization: Bearer <token>` |
| **语言** | **`Accept-Language: zh-CN` (多语言内容/错误消息)** |
| 追踪 | `X-Trace-Id: <uuid>` (全链路追踪) |
| 版本 | URL 路径版本 (`/api/v1/...`, `/v1/...`) |
| 分页 | `?page=1&size=20`, 返回 `{items:[], total, page, size}` |
| 时间格式 | ISO 8601 (`2026-05-27T10:30:00Z`) |
| 货币 | 最小单位整数 (分) 或 TEXT |

---
