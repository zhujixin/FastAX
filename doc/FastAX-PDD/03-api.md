## 8. 接口详细设计

### 8.1 用户端 API (OpenAI 兼容协议)

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

### 8.2 公开 API (无需认证)

#### 8.2.1 Token 商品

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/tokens/products` | GET | Token 商品列表 (含多语言字段) |
| `/api/tokens/products/:id` | GET | 商品详情 (含多语言字段) |

#### 8.2.2 模型市场 (公开)

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/models` | GET | 模型列表 |
| `/api/models/benchmarks` | GET | 基准测试数据 |
| `/api/models/variants/:variant` | GET | 模型变体详情 |
| `/api/providers/health` | GET | 供应商健康面板 |
| `/api/providers/:id/health` | GET | 供应商健康详情 |

#### 8.2.3 i18n

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/i18n/languages` | GET | 获取可用语言列表 |
| `/api/i18n/translations/:locale` | GET | 获取翻译文件 (CDN 回源) |

#### 8.2.4 支付回调 (公开 webhook)

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/payments/callback` | POST | 支付网关回调 (签名验证在 handler 内部进行) |

### 8.3 认证相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/auth/register` | POST | 无 | 注册 (手机/邮箱+验证码) |
| `/api/auth/login` | POST | 无 | 登录，返回 JWT (含语言偏好) |
| `/api/auth/refresh` | POST | 无 | 刷新 Access Token |
| `/api/auth/send-code` | POST | 无 | 发送验证码 (短信/邮件，按语言) |
| `/api/auth/reset-password` | POST | 无 | 重置密码 |
| `/api/auth/oauth/:provider` | GET | 无 | OAuth 登录跳转 |
| `/api/auth/oauth/callback` | GET | 无 | OAuth 回调 |
| `/api/auth/oauth/login` | POST | 无 | OAuth 直接登录 |
| `/api/auth/logout` | POST | JWT | 登出，销毁 Token |

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

### 8.4 Token 相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/tokens/products` | GET | 无 | Token 商品列表 (公开) |
| `/api/tokens/products/:id` | GET | 无 | 商品详情 (公开) |
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

### 8.5 订单相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/orders` | POST | JWT | 创建订单 |
| `/api/orders` | GET | JWT | 订单列表 |
| `/api/orders/:id` | GET | JWT | 订单详情 |
| `/api/orders/:id/cancel` | POST | JWT | 取消订单 |
| `/api/orders/:id/refund` | POST | JWT | 申请退款 |

### 8.6 支付相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/payments` | POST | JWT | 创建支付 |
| `/api/payments/callback` | POST | 无 | 支付回调 (公开 webhook) |
| `/api/payments/:order_id` | GET | JWT | 查询支付 |
| `/api/payments/refunds` | POST | JWT | 创建退款 |
| `/api/payments/refunds` | GET | JWT | 退款列表 |

### 8.7 用户相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/user/me` | GET | JWT | 获取当前用户信息 |
| `/api/user/language` | PUT | JWT | 更新用户语言偏好 |

### 8.8 统计相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/stats/usage` | GET | JWT | 用量统计 |
| `/api/stats/consumption` | GET | JWT | 消费统计 |
| `/api/stats/bills` | GET | JWT | 账单明细 |
| `/api/stats/summary` | GET | JWT | 控制台总览 |

### 8.9 通知相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/notifications` | GET | JWT | 通知列表 (当前语言) |
| `/api/notifications/unread-count` | GET | JWT | 未读数 |
| `/api/notifications/:id/read` | PUT | JWT | 标记已读 |
| `/api/notifications/read-all` | PUT | JWT | 全部已读 |

### 8.10 i18n 相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/i18n/languages` | GET | 无 | 获取可用语言列表 |
| `/api/i18n/translations/:locale` | GET | 无 | 获取翻译文件 |
| `/api/user/language` | PUT | JWT | 更新用户语言偏好 |

### 8.11 供应商自服务 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/vendor/apply` | POST | JWT | 提交入驻申请 |
| `/api/vendor/me` | GET | JWT | 查看店铺信息 |
| `/api/vendor/profile` | PUT | JWT | 更新店铺信息 |
| `/api/vendor/products` | GET | JWT | 供应商商品列表 |
| `/api/vendor/products` | POST | JWT | 创建商品 |
| `/api/vendor/products/:id` | PUT | JWT | 更新商品 |
| `/api/vendor/products/:id/price` | PUT | JWT | 调价 |
| `/api/vendor/sales` | GET | JWT | 销售看板 |
| `/api/vendor/settlements` | GET | JWT | 结算单列表 |
| `/api/vendor/settlements/:id/confirm` | POST | JWT | 确认结算单 |
| `/api/vendor/settlements/:id/withdraw` | POST | JWT | 申请提现 |

### 8.12 佣金相关 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/commissions` | GET | JWT | 佣金列表 |
| `/api/commissions/total` | GET | JWT | 佣金总额 |
| `/api/commissions/withdraw` | POST | JWT | 申请提现 |

### 8.13 BYOK API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/byok/keys` | GET | JWT | 用户 Key 列表 |
| `/api/byok/keys` | POST | JWT | 添加 Key |
| `/api/byok/keys/:id` | DELETE | JWT | 删除 Key |
| `/api/byok/keys/:id/status` | PUT | JWT | 更新 Key 状态 |
| `/api/byok/usage` | GET | JWT | BYOK 用量统计 |
| `/api/byok/preference` | PUT | JWT | 路由优先级配置 |

### 8.14 成本优化 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/user/budget` | GET | JWT | 用户预算设置 |
| `/api/user/budget` | PUT | JWT | 设置用户预算 |
| `/api/user/cost-alerts` | GET | JWT | 成本告警配置 |
| `/api/user/cost-alerts` | PUT | JWT | 设置成本告警 |

### 8.15 模型市场 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/models` | GET | 无 | 模型列表 (公开) |
| `/api/models/compare` | POST | JWT | 模型多维度对比 |
| `/api/models/recommend` | GET | JWT | 模型推荐 |
| `/api/models/benchmarks` | GET | 无 | 基准测试数据 |
| `/api/models/variants/:variant` | GET | 无 | 模型变体详情 |
| `/api/providers/health` | GET | 无 | 健康面板 |
| `/api/providers/:id/health` | GET | 无 | 供应商健康详情 |

### 8.16 企业功能 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/enterprise/sub-accounts` | POST | JWT | 创建子账号 |
| `/api/enterprise/sub-accounts` | GET | JWT | 子账号列表 |
| `/api/enterprise/sub-accounts/:id/status` | PUT | JWT | 子账号状态 |
| `/api/enterprise/sub-accounts/:id/quota` | PUT | JWT | 子账号额度 |
| `/api/enterprise/usage` | GET | JWT | 企业用量 |
| `/api/enterprise/sub-accounts/:id/usage` | GET | JWT | 子账号用量详情 |

---

### 8.17 管理后台 API

#### Dashboard

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/dashboard/summary` | GET | 核心数据总览 |
| `/api/admin/dashboard/charts` | GET | 趋势图表数据 |

#### 用户管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/users` | GET | 用户列表 (搜索/筛选) |
| `/api/admin/users/:id` | GET | 用户详情 |
| `/api/admin/users/:id/status` | PUT | 冻结/解冻用户 |
| `/api/admin/users/:id/level` | PUT | 修改用户等级 |

#### Token/商品管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/products` | POST | 新增商品 |
| `/api/admin/products/:id` | PUT | 编辑商品 |

#### 供应商 (Supplier) 管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/suppliers` | POST | 新增供应商 |
| `/api/admin/suppliers` | GET | 供应商列表 |
| `/api/admin/suppliers/:id` | GET | 供应商详情 |
| `/api/admin/suppliers/:id` | PUT | 编辑供应商 |
| `/api/admin/suppliers/:id/status` | PUT | 启用/禁用供应商 |

#### 渠道 (Channel) 管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/channels` | GET | 渠道列表 |
| `/api/admin/channels/:id/status` | PUT | 启用/禁用渠道 |
| `/api/admin/channels/:id/priority` | PUT | 调整优先级 |

#### 供应商入驻 (Vendor) 管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/vendors` | GET | 入驻供应商列表 |
| `/api/admin/vendors/:id` | GET | 供应商详情 |
| `/api/admin/vendors/:id/review` | POST | 审核供应商 (approve/reject 合并) |
| `/api/admin/vendors/:id/suspend` | POST | 冻结供应商 |
| `/api/admin/vendors/:id/settlements` | GET | 供应商结算记录 |
| `/api/admin/vendor-commission-rates/:id` | PUT | 配置佣金比例 |
| `/api/admin/vendors/:vendor_id/products` | POST | 创建供应商商品 |
| `/api/admin/vendors/:vendor_id/products` | GET | 供应商商品列表 |
| `/api/admin/vendor-products/:id/review` | POST | 商品审核 |
| `/api/admin/vendor-products/:id/price` | PUT | 商品调价 |

#### 订单管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/orders` | GET | 订单列表 |
| `/api/admin/orders/:id` | GET | 订单详情 |
| `/api/admin/orders/:id/refund` | POST | 审核退款 |

#### 报表

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/reports/daily` | GET | 日报表 |
| `/api/admin/reports/monthly` | GET | 月报表 |

#### 风控管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/risk/rules` | GET | 规则列表 |
| `/api/admin/risk/rules` | POST | 新增规则 |
| `/api/admin/risk/rules/:id/enabled` | PUT | 启用/禁用规则 |
| `/api/admin/risk/events` | GET | 风控事件列表 |
| `/api/admin/risk/events/:id/handle` | PUT | 处理风控事件 |
| `/api/admin/risk/blacklist` | GET | 黑名单列表 |
| `/api/admin/risk/blacklist` | POST | 添加黑名单 |
| `/api/admin/risk/blacklist/:id` | DELETE | 删除黑名单 |

#### 佣金管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/commissions/:id/settle` | POST | 结算佣金 |

#### 系统管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/system/config` | GET | 系统配置 |
| `/api/admin/system/config` | PUT | 更新配置 |
| `/api/admin/system/admins` | GET | 管理员列表 |
| `/api/admin/system/admins` | POST | 添加管理员 |

#### 审计日志

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/audit/logs` | GET | 操作日志 |
| `/api/admin/audit/export` | GET | 审计日志导出 |
| `/api/admin/call-logs` | GET | 调用日志查询 |
| `/api/admin/system/logs` | GET | 系统日志 (→ 审计日志) |

#### 通知模板管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/notifications/templates` | GET | 模板列表 |
| `/api/admin/notifications/templates` | POST | 创建模板 |
| `/api/admin/notifications/templates/:id` | PUT | 更新模板 |

#### i18n 管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/i18n/languages` | GET | 语言列表 |
| `/api/admin/i18n/languages` | POST | 新增语言 |
| `/api/admin/i18n/languages/:locale` | PUT | 更新语言 |
| `/api/admin/i18n/default` | PUT | 设置默认语言 |

#### 安全护栏管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/guardrails/rules` | GET | 规则列表 |
| `/api/admin/guardrails/rules` | POST | 创建规则 |
| `/api/admin/guardrails/rules/:id` | PUT | 更新规则 |
| `/api/admin/guardrails/rules/:id` | DELETE | 删除规则 |
| `/api/admin/guardrails/rules/:id/enabled` | PUT | 启用/禁用规则 |
| `/api/admin/guardrails/logs` | GET | 检测日志查询 |
| `/api/admin/guardrails/detect` | POST | 护栏检测测试 |
| `/api/admin/guardrails/config` | PUT | 全局配置 |

#### SSO 管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/sso/config` | GET | SSO 配置 |
| `/api/admin/sso/config` | PUT | 更新 SSO 配置 |

#### 团队管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/teams` | GET | 团队列表 |
| `/api/admin/teams` | POST | 创建团队 |
| `/api/admin/teams/:id` | PUT | 编辑团队 |
| `/api/admin/teams/:id` | DELETE | 删除团队 |

#### 缓存管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/admin/cache/stats` | GET | 语义缓存统计 |
| `/api/admin/cache/config` | PUT | 缓存策略配置 |

---

### 8.18 OpenAI 兼容 Relay 路由

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/chat/completions` | POST | API Key | Chat Completion (流式/非流式) |
| `/v1/messages` | POST | API Key | Anthropic Messages API |
| `/v1/images/generations` | POST | API Key | 图片生成 |
| `/v1/audio/speech` | POST | API Key | 语音合成 TTS |
| `/v1/audio/transcriptions` | POST | API Key | 语音转文本 STT |
| `/v1/video/generations` | POST | API Key | 视频生成 |
| `/v1/rerank` | POST | API Key | Rerank 统一 API |
| `/v1/models` | GET | API Key | 模型列表 |

---

### 8.19 错误码设计

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

### 8.20 接口通用规范

| 规范 | 内容 |
|------|------|
| 协议 | HTTPS (TLS 1.3) |
| 请求头 | `Content-Type: application/json` |
| 认证 | `Authorization: Bearer <token>` |
| **语言** | **`Accept-Language: zh-CN` (多语言内容/错误消息)** |
| 追踪 | `X-Trace-Id: <uuid>` (全链路追踪) |
| 版本 | URL 路径版本 (`/api/...`, `/v1/...`) |
| 分页 | `?page=1&size=20`, 返回 `{items:[], total, page, size}` |
| 时间格式 | ISO 8601 (`2026-05-27T10:30:00Z`) |
| 货币 | 最小单位整数 (分) 或 TEXT |

### 8.21 API 统计

| 分类 | 端点数 |
|------|--------|
| 公开 (无需认证) | 12 |
| 认证相关 | 9 |
| 用户端 (JWT) | 48 |
| 管理后台 (JWT + Admin) | 80 |
| Relay /v1 (API Key) | 8 |
| **总计** | **157** |
