## 7. 数据库详细设计

> **MVP 阶段建议**：参考 one-api 的单数据库模式，将 12 个微服务的 SQLite DB 合并为 **1 个 SQLite DB** (WAL 模式)。
> 理由：MVP 用户量 (<1000) 下 SQLite WAL 的并发读写完全够用，降低 80% 部署和运维复杂度，
> 后期用户增长时再按服务拆分。参考 one-api `model.InitDB()` + `model.InitLogDB()` 的双 DB 分离模式。

### 6.1 ER 图 (核心实体关系)

```
User ────1:N──── SubAccount
  │
  │──1:1──── UserProfile
  │
  │──1:N──── UserToken ──N:1── TokenProduct ──N:1── Supplier
  │              │
  │              │──1:N── TokenTransfer
  │
  │──1:N──── Order ──1:1── Payment
  │              │
  │              │──1:N── Refund
  │
  │──1:N──── CallLog
  │
  │──1:N──── RiskEvent
  │
  │──1:N──── Notification

  ── 新增实体 (PRD v2.4) ──

  SupplierVendor ──1:N── SupplierProduct ──── CallLog (调用)
       │                                        │
       │──1:N── Settlement                      │
       │                                        │
       user_id ──FK──→ User                     │
                                                │
  SupportedLanguage (配置表, 独立)
```

### 6.2 数据表详细设计

#### 6.2.1 用户相关表

##### `user` — 用户账号表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | 用户 ID |
| username | TEXT | UNIQUE, NOT NULL | 用户名 |
| password_hash | TEXT | NOT NULL | BCrypt 哈希 |
| email | TEXT | UNIQUE | 邮箱 |
| phone | TEXT | UNIQUE | 手机号 |
| role | TEXT | NOT NULL, DEFAULT 'user' | 角色 (含 vendor) |
| level | TEXT | NOT NULL, DEFAULT 'normal' | 用户等级 |
| status | INTEGER | NOT NULL, DEFAULT 1 | 0=冻结, 1=正常, 2=注销 |
| email_verified | INTEGER | DEFAULT 0 | 邮箱是否验证 |
| phone_verified | INTEGER | DEFAULT 0 | 手机号是否验证 |
| **preferred_language** | **TEXT** | **DEFAULT 'zh-CN'** | **用户首选语言 (PRD F-REG-07)** |
| last_login_ip | TEXT | — | 最后登录 IP |
| last_login_at | TEXT | — | 最后登录时间 |
| login_fail_count | INTEGER | DEFAULT 0 | 连续登录失败次数 |
| locked_until | TEXT | — | 锁定截止时间 |
| created_at | TEXT | NOT NULL | 创建时间 |
| updated_at | TEXT | NOT NULL | 更新时间 |

**索引**：`idx_email`, `idx_phone`, `idx_username`, `idx_status_created`, `idx_preferred_language`

##### `user_profile` — 用户扩展信息表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| user_id | INTEGER | UNIQUE, FK→user(id) | 用户 ID |
| avatar | TEXT | — | 头像 URL |
| real_name | TEXT | — | 真实姓名 |
| id_number | TEXT | — | 身份证号 (AES-256 加密) |
| company_name | TEXT | — | 企业名称 |
| business_license | TEXT | — | 营业执照 URL |
| company_address | TEXT | — | 企业地址 |
| invite_code | TEXT | UNIQUE | 邀请码 |
| invited_by | INTEGER | — | 邀请人 user_id |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

##### `sub_account` — 子账号表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| parent_id | INTEGER | FK→user(id), NOT NULL | 主账号 ID |
| email | TEXT | NOT NULL | 子账号邮箱 |
| password_hash | TEXT | NOT NULL | 密码 |
| token_quota | TEXT | DEFAULT 0 | Token 额度上限 |
| permissions | TEXT | — | 权限列表 ["api:chat", "api:embedding"] |
| status | INTEGER | DEFAULT 1 | 0=禁用, 1=启用 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

#### 6.2.2 Token 相关表

##### `supplier` — Token 供应商表 (平台自有)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| name | TEXT | NOT NULL | 供应商名称 (OpenAI, Claude...) |
| code | TEXT | UNIQUE, NOT NULL | 编码 (openai, claude) |
| description | TEXT | — | 描述 |
| api_base_url | TEXT | NOT NULL | API 基础地址 |
| api_key_encrypted | TEXT | NOT NULL | API Key (AES-256) |
| models | TEXT | — | 支持的模型列表 |
| **region** | **TEXT** | **DEFAULT 'overseas'** | **所属区域 (OCN-SUP-01)** |
| status | INTEGER | DEFAULT 1 | 0=禁用, 1=启用 |
| priority | INTEGER | DEFAULT 0 | 路由优先级 (数值越高越优先) |
| weight | INTEGER | DEFAULT 10 | 负载权重 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**: `idx_region`, `idx_status_priority`

##### `token_product` — Token 商品表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| supplier_id | INTEGER | FK→supplier(id) | 供应商 ID |
| **vendor_id** | **INTEGER** | **FK→supplier_vendor(id)** | **入驻供应商 ID (可选)** |
| name | TEXT | NOT NULL | 商品名称 |
| **name_i18n** | **TEXT** | — | **多语言名称 (LANG-02-02)** |
| type | TEXT | NOT NULL | Token 类型 (ai_token, digital_asset...) |
| model | TEXT | — | AI 模型名 (gpt-4, claude-3) |
| unit | TEXT | NOT NULL | 单位 (次/百万Token/个) |
| price | TEXT | NOT NULL | 销售单价 |
| original_price | TEXT | — | 原价 (划线价) |
| currency | TEXT | DEFAULT 'CNY' | 货币 |
| description | TEXT | — | 商品描述 |
| **description_i18n** | **TEXT** | — | **多语言描述** |
| validity_days | INTEGER | — | 有效期 (天) |
| usage_notes | TEXT | — | 使用限制说明 |
| sort_order | INTEGER | DEFAULT 0 | 排序 |
| status | INTEGER | DEFAULT 1 | 0=下架, 1=上架 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_supplier`, `idx_vendor`, `idx_type_model`, `idx_status_sort`

##### `token_inventory` — Token 库存表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| supplier_id | INTEGER | FK→supplier(id) | 供应商 ID |
| product_id | INTEGER | FK→token_product(id) | 商品 ID |
| total_amount | TEXT | NOT NULL | 总库存 |
| remaining_amount | TEXT | NOT NULL | 剩余库存 |
| alert_threshold | TEXT | DEFAULT 10.00 | 预警阈值 (%) |
| last_synced_at | TEXT | — | 最后同步时间 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

##### `user_token` — 用户持有 Token 表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| user_id | INTEGER | FK→user(id), NOT NULL | 用户 ID |
| product_id | INTEGER | FK→token_product(id) | 商品 ID |
| order_id | INTEGER | FK→order(id) | 来源订单 |
| total_amount | TEXT | NOT NULL | 总获得量 |
| used_amount | TEXT | DEFAULT 0 | 已使用量 |
| frozen_amount | TEXT | DEFAULT 0 | 冻结量 (转让中) |
| expires_at | TEXT | — | 过期时间 |
| status | INTEGER | DEFAULT 1 | 0=冻结, 1=可用, 2=过期 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_user_product`, `idx_expires_status`, `idx_user_status`

##### `token_transfer` — Token 转让记录表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| from_user_id | INTEGER | FK→user(id) | 转出方 |
| to_user_id | INTEGER | FK→user(id) | 转入方 |
| product_id | INTEGER | FK→token_product(id) | — |
| amount | TEXT | NOT NULL | 数量 |
| status | TEXT | DEFAULT 'pending' | 状态 |
| created_at | TEXT | NOT NULL | — |
| handled_at | TEXT | — | 处理时间 |

#### 6.2.3 订单相关表

##### `order` — 订单表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| order_no | TEXT | UNIQUE, NOT NULL | 订单号 (业务号) |
| user_id | INTEGER | FK→user(id), NOT NULL | 用户 ID |
| product_id | INTEGER | FK→token_product(id) | — |
| **vendor_id** | **INTEGER** | **FK→supplier_vendor(id)** | **入驻供应商 (可选)** |
| quantity | TEXT | NOT NULL | 数量 |
| unit_price | TEXT | NOT NULL | 单价 |
| amount | TEXT | NOT NULL | 总金额 |
| discount_amount | TEXT | DEFAULT 0 | 优惠金额 |
| final_amount | TEXT | NOT NULL | 实付金额 |
| currency | TEXT | DEFAULT 'CNY' | 货币 |
| payment_method | TEXT | — | 支付方式 |
| status | TEXT | NOT NULL | 状态 |
| remark | TEXT | — | 备注 |
| paid_at | TEXT | — | 支付时间 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_order_no`, `idx_user`, `idx_vendor`, `idx_status`, `idx_created`

##### `payment` — 支付记录表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| order_id | INTEGER | FK→order(id) | 订单 ID |
| payment_no | TEXT | UNIQUE | 支付流水号 |
| amount | TEXT | NOT NULL | 支付金额 |
| method | TEXT | NOT NULL | 支付方式 |
| gateway | TEXT | NOT NULL | 支付网关 (wechat/alipay/stripe) |
| gateway_trade_no | TEXT | — | 网关交易号 |
| gateway_status | TEXT | — | 网关支付状态 |
| status | TEXT | NOT NULL | 状态 |
| raw_response | TEXT | — | 网关原始响应 (TEXT) |
| paid_at | TEXT | — | 支付时间 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

##### `refund` — 退款记录表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| order_id | INTEGER | FK→order(id) | 订单 ID |
| payment_id | INTEGER | FK→payment(id) | — |
| refund_no | TEXT | UNIQUE | 退款单号 |
| amount | TEXT | NOT NULL | 退款金额 |
| reason | TEXT | — | 退款原因 |
| status | TEXT | NOT NULL | 状态 |
| operator_id | INTEGER | FK→user(id) | 审核人 |
| remark | TEXT | — | 审核备注 |
| created_at | TEXT | NOT NULL | — |
| handled_at | TEXT | — | 处理时间 |

#### 6.2.4 调用日志表

##### `call_log` — Token 调用日志表 (高写入，分区表)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| trace_id | TEXT | NOT NULL | 全链路追踪 ID |
| user_id | INTEGER | FK→user(id) | 用户 ID |
| sub_account_id | INTEGER | — | 子账号 ID (如有) |
| product_id | INTEGER | FK→token_product(id) | — |
| supplier_id | INTEGER | FK→supplier(id) | 实际路由到的供应商 |
| **vendor_id** | **INTEGER** | **FK→supplier_vendor(id)** | **入驻供应商 (可选)** |
| **route_decision** | **TEXT** | — | **路由决策信息 (ROUTE-11)** |
| request_path | TEXT | NOT NULL | 请求路径 (/v1/chat/completions) |
| request_model | TEXT | — | 请求模型 |
| tokens_prompt | INTEGER | DEFAULT 0 | Prompt Token 数 |
| tokens_completion | INTEGER | DEFAULT 0 | Completion Token 数 |
| tokens_total | INTEGER | DEFAULT 0 | 总 Token 消耗 |
| response_time_ms | INTEGER | — | 响应耗时 |
| is_stream | INTEGER | DEFAULT 0 | 是否流式 |
| status_code | SMALLINT | — | HTTP 状态码 |
| status | TEXT | NOT NULL | 调用状态 |
| error_message | TEXT | — | 错误信息 |
| client_ip | TEXT | — | 客户端 IP |
| user_agent | TEXT | — | User-Agent |
| created_at | TEXT | NOT NULL | — |

**索引**：`idx_trace_id`, `idx_user_created`, `idx_supplier_created`, `idx_vendor_created`, `idx_created`
**归档**：按 `created_at` 按月归档 (独立 DB 文件, 保留 6 个月)

#### 6.2.5 风控相关表

##### `risk_event` — 风控事件表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| user_id | INTEGER | FK→user(id) | — |
| event_type | TEXT | NOT NULL | 事件类型 (abnormal_login, large_trade, rapid_api...) |
| risk_level | TEXT | NOT NULL | 风险等级 |
| description | TEXT | — | 事件描述 |
| rule_id | INTEGER | — | 触发的风控规则 ID |
| related_info | TEXT | — | 相关信息 (IP、金额、设备指纹等) |
| action_taken | TEXT | — | 处置动作 |
| status | TEXT | DEFAULT 'pending' | 状态 |
| handler_id | INTEGER | FK→user(id) | 处理人 |
| handled_at | TEXT | — | 处理时间 |
| created_at | TEXT | NOT NULL | — |

##### `risk_rule` — 风控规则表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| name | TEXT | NOT NULL | 规则名称 |
| category | TEXT | NOT NULL | 分类 (register, trade, api) |
| conditions | TEXT | NOT NULL | 规则条件表达式 |
| action | TEXT | NOT NULL | 处置动作 (alert, rate_limit, freeze) |
| risk_level | TEXT | NOT NULL | 风险等级 |
| priority | INTEGER | DEFAULT 0 | 优先级 |
| enabled | INTEGER | DEFAULT 1 | 是否启用 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

#### 6.2.6 通知相关表

##### `notification` — 通知表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| user_id | INTEGER | FK→user(id) | 接收人 |
| type | TEXT | NOT NULL | 类型 (order, security, expiry, system) |
| channel | TEXT | NOT NULL | 通道 (in_app, sms, email) |
| title | TEXT | NOT NULL | 标题 (已本地化) |
| content | TEXT | NOT NULL | 内容 (已本地化) |
| **language** | **TEXT** | — | **发送时使用的语言** |
| is_read | INTEGER | DEFAULT 0 | 是否已读 |
| read_at | TEXT | — | 读取时间 |
| created_at | TEXT | NOT NULL | — |

**索引**：`idx_user_read`, `idx_user_created`

##### `notification_template` — 通知模板表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| code | TEXT | UNIQUE, NOT NULL | 模板编码 |
| name | TEXT | — | 模板名称 |
| channel | TEXT | NOT NULL | 通道 |
| content | TEXT | NOT NULL | 模板内容 (FreeMarker) |
| **language** | **TEXT** | **DEFAULT 'zh-CN'** | **语言 (LANG-03)** |
| status | INTEGER | DEFAULT 1 | 0=禁用, 1=启用 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

#### 6.2.7 佣金相关表

##### `commission` — 佣金记录表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| agent_id | INTEGER | FK→user(id) | 代理用户 ID |
| customer_id | INTEGER | FK→user(id) | 客户用户 ID |
| order_id | INTEGER | FK→order(id) | 关联订单 |
| order_amount | TEXT | NOT NULL | 订单金额 |
| commission_rate | TEXT | NOT NULL | 佣金率 |
| commission_amount | TEXT | NOT NULL | 佣金金额 |
| status | TEXT | DEFAULT 'pending' | 状态 |
| settled_at | TEXT | — | 结算时间 |
| created_at | TEXT | NOT NULL | — |

#### 6.2.8 审计日志表

##### `audit_log` — 操作审计日志表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| trace_id | TEXT | NOT NULL | 追踪 ID |
| operator_id | INTEGER | — | 操作人 ID |
| operator_name | TEXT | — | 操作人用户名 |
| operator_ip | TEXT | — | 操作人 IP |
| action | TEXT | NOT NULL | 操作动作 (user.create, order.refund...) |
| resource_type | TEXT | — | 资源类型 |
| resource_id | TEXT | — | 资源 ID |
| detail | TEXT | — | 操作详情 (变更前后对比) |
| result | TEXT | NOT NULL | 结果 |
| fail_reason | TEXT | — | 失败原因 |
| created_at | TEXT | NOT NULL | — |

**索引**：`idx_trace_id`, `idx_operator`, `idx_action`, `idx_created`

#### 6.2.9 新增：供应商入驻相关表 (PRD §6.2.6)

##### `supplier_vendor` — 入驻供应商表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| user_id | INTEGER | UNIQUE, FK→user(id) | 关联用户 ID |
| company_name | TEXT | NOT NULL | 企业名称 |
| contact_name | TEXT | — | 联系人姓名 |
| contact_email | TEXT | NOT NULL | 联系邮箱 |
| contact_phone | TEXT | — | 联系电话 |
| business_license | TEXT | — | 营业执照 URL |
| api_base_url | TEXT | NOT NULL | API 基础地址 |
| api_auth_type | TEXT | DEFAULT 'api_key' | API 认证方式 |
| api_key_encrypted | TEXT | — | API Key (AES-256) |
| commission_rate | TEXT | NOT NULL | 平台佣金比例 (SUP-12) |
| settlement_cycle | TEXT | DEFAULT 't+7' | 结算周期 (R-BIZ-51) |
| status | TEXT | NOT NULL, DEFAULT 'pending' | 状态 |
| reject_reason | TEXT | — | 审核驳回原因 |
| approved_at | TEXT | — | 审核通过时间 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_user`, `idx_status`, `idx_created`

##### `supplier_product` — 供应商商品表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| vendor_id | INTEGER | FK→supplier_vendor(id), NOT NULL | 供应商 ID |
| name | TEXT | NOT NULL | 商品名称 |
| **name_i18n** | **TEXT** | — | **多语言商品名 (SUP-04)** |
| type | TEXT | NOT NULL | Token 类型 |
| model | TEXT | NOT NULL | 模型名称 |
| api_endpoint | TEXT | NOT NULL | API 调用端点 |
| auth_type | TEXT | DEFAULT 'api_key' | 认证方式 |
| unit | TEXT | NOT NULL | 计费单位 |
| price | TEXT | NOT NULL | 单价 |
| currency | TEXT | DEFAULT 'USD' | 货币 |
| min_price | TEXT | — | 平台最低限价 (R-BIZ-50) |
| max_price | TEXT | — | 平台最高限价 (R-BIZ-50) |
| stock_total | TEXT | — | 总库存 (SUP-08) |
| stock_remaining | TEXT | — | 剩余库存 |
| status | TEXT | DEFAULT 'pending_review' | 状态 |
| health_status | TEXT | DEFAULT 'unknown' | API 健康状态 (SUP-15) |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_vendor`, `idx_status`, `idx_health`

##### `settlement` — 供应商结算表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INTEGER | PK | — |
| vendor_id | INTEGER | FK→supplier_vendor(id) | 供应商 ID |
| settlement_no | TEXT | UNIQUE, NOT NULL | 结算单号 |
| period_start | TEXT | NOT NULL | 结算周期开始 |
| period_end | TEXT | NOT NULL | 结算周期结束 |
| total_sales | TEXT | NOT NULL | 总销售额 |
| commission_amount | TEXT | NOT NULL | 平台佣金 |
| net_amount | TEXT | NOT NULL | 应结金额 |
| currency | TEXT | DEFAULT 'USD' | 结算货币 |
| status | TEXT | DEFAULT 'pending' | 状态 |
| payment_method | TEXT | — | 支付方式 (bank_transfer, paypal) |
| paid_at | TEXT | — | 付款时间 |
| remark | TEXT | — | 备注 |
| created_at | TEXT | NOT NULL | — |
| updated_at | TEXT | NOT NULL | — |

**索引**：`idx_vendor`, `idx_status`, `idx_period`

#### 6.2.10 新增：多语言配置表 (PRD §6.9.4)

##### `supported_language` — 支持的语言配置表

```sql
CREATE TABLE `supported_language` (
  `id` INTEGER PRIMARY KEY,
  `locale` TEXT NOT NULL UNIQUE COMMENT '语言代码',
  `name` TEXT NOT NULL COMMENT '语言名称（英文）',
  `native_name` TEXT NOT NULL COMMENT '语言名称（母语）',
  `is_enabled` INTEGER DEFAULT 1 COMMENT '是否启用',
  `sort_order` INTEGER DEFAULT 0 COMMENT '排序',
  `is_default` INTEGER DEFAULT 0 COMMENT '是否默认语言',
  `fallback_locale` TEXT DEFAULT 'en' COMMENT '回退语言'
);
```

#### 6.2.11 BYOK 相关表 (PRD §6.13)

##### `byok_key` — 用户自有 API Key

```sql
CREATE TABLE `byok_key` (
  `id` INTEGER PRIMARY KEY,
  `user_id` INTEGER NOT NULL REFERENCES user(id),
  `provider` TEXT NOT NULL,
  `key_encrypted` TEXT NOT NULL,
  `key_iv` TEXT NOT NULL,
  `alias` TEXT DEFAULT '',
  `model_whitelist` TEXT DEFAULT '',
  `status` INTEGER DEFAULT 1,
  `last_used_at` INTEGER DEFAULT 0,
  `expires_at` INTEGER DEFAULT 0,
  `created_at` INTEGER NOT NULL
);
CREATE INDEX idx_byok_user ON byok_key(user_id, status);
```

##### `guardrail_rule` — 护栏规则

```sql
CREATE TABLE `guardrail_rule` (
  `id` INTEGER PRIMARY KEY,
  `name` TEXT NOT NULL,
  `stage` TEXT NOT NULL,
  `type` TEXT NOT NULL,
  `action` TEXT NOT NULL,
  `conditions` TEXT,
  `priority` INTEGER DEFAULT 0,
  `enabled` INTEGER DEFAULT 1,
  `created_at` INTEGER NOT NULL
);
```

##### `guardrail_log` — 护栏检测日志

```sql
CREATE TABLE `guardrail_log` (
  `id` INTEGER PRIMARY KEY,
  `trace_id` TEXT NOT NULL,
  `user_id` INTEGER REFERENCES user(id),
  `rule_id` INTEGER REFERENCES guardrail_rule(id),
  `stage` TEXT NOT NULL,
  `detected_entities` TEXT,
  `action_taken` TEXT NOT NULL,
  `created_at` INTEGER NOT NULL
);
CREATE INDEX idx_guardrail_trace ON guardrail_log(trace_id);
CREATE INDEX idx_guardrail_time ON guardrail_log(created_at);
```

##### `semantic_cache` — 语义缓存条目

```sql
CREATE TABLE `semantic_cache` (
  `id` INTEGER PRIMARY KEY,
  `prompt_hash` TEXT NOT NULL,
  `prompt_vector` BLOB,
  `response_encrypted` TEXT NOT NULL,
  `model` TEXT NOT NULL,
  `hit_count` INTEGER DEFAULT 0,
  `created_at` INTEGER NOT NULL,
  `expires_at` INTEGER NOT NULL
);
CREATE INDEX idx_cache_hash ON semantic_cache(prompt_hash);
CREATE INDEX idx_cache_expires ON semantic_cache(expires_at);
```

##### `model_variant` — 模型后缀变体

```sql
CREATE TABLE `model_variant` (
  `id` INTEGER PRIMARY KEY,
  `base_model` TEXT NOT NULL,
  `suffix` TEXT NOT NULL,
  `provider_id` INTEGER REFERENCES supplier(id),
  `price_coefficient` REAL DEFAULT 1.0,
  `priority` INTEGER DEFAULT 0
);
CREATE UNIQUE INDEX idx_variant ON model_variant(base_model, suffix, provider_id);
```

##### `provider_health` — 供应商健康历史

```sql
CREATE TABLE `provider_health` (
  `id` INTEGER PRIMARY KEY,
  `provider_id` INTEGER REFERENCES supplier(id),
  `status` INTEGER NOT NULL,
  `avg_latency_ms` INTEGER DEFAULT 0,
  `error_rate` REAL DEFAULT 0,
  `check_count` INTEGER DEFAULT 0,
  `period_start` INTEGER NOT NULL,
  `period_end` INTEGER NOT NULL
);
CREATE INDEX idx_health_provider ON provider_health(provider_id, period_start);
```

### 6.3 Redis 缓存设计

| Key 模式 | Value 类型 | 说明 | TTL |
|----------|-----------|------|-----|
| `user:session:{user_id}` | Hash | 用户会话 | 7d |
| `user:quota:{user_id}` | String (INT64) | 用户配额 | 60s |
| `token:product:{id}` | TEXT | Token 商品信息 | 30min |
| `token:supplier:{id}` | TEXT | 供应商信息 | 5min |
| `route:health:{id}` | String | 渠道健康状态 | 10s |
| `rate:limit:{key}` | Counter | 限流计数 | 1min |
| `verify:code:{phone_or_email}` | String | 验证码 | 5min |
| `refresh:token:{token}` | String | Refresh Token | 7d |
| `i18n:translations:{locale}:{ns}` | Hash | 翻译缓存 | 1h |
| `config:system:*` | String | 系统配置 | 10min |

### 6.4 数据库分片策略 (SQLite)

| 表 | 策略 | 说明 |
|---|------|------|
| call_log | 按月归档 (独立 DB 文件) | 高频写入，保留 6 个月 |
| audit_log | 按月归档 (独立 DB 文件) | 写入频繁，保留 12 个月 |
| notification | 按 user_id hash 分库 (16 DB) | 用户级数据，按用户分散 |
| commission | 按月归档 (独立 DB 文件) | 财务数据，按时间归档 |
| settlement | 按月归档 (独立 DB 文件) | 结算数据，按时间归档 |
| 其余核心表 | 单 DB 文件 + WAL 模式 | 数据量可控，WAL 支持并发读写 |

> **SQLite 分片说明**：SQLite 不支持 MySQL 表分区语法。可通过多 DB 文件实现水平拆分：每个分片/分区生成独立 `.db` 文件，应用层通过 `ATTACH DATABASE` 或连接路由跨库查询。WAL (Write-Ahead Logging) 模式支持并发读取+单写入器，满足中小规模并发需求。

---

