## 8. 非功能需求

### 7.1 性能需求

| 需求 | 指标 | 说明 |
|------|------|------|
| R-PERF-01 | 接口响应时间 P95 ≤ 500ms | 核心业务接口，不含 Token 中转代理 |
| R-PERF-02 | Token 中转延迟 P95 ≤ 1000ms | 用户调用 Token 端到端延迟 |
| R-PERF-03 | 流式输出首包延迟 ≤ 500ms | 首个 token 返回时间 |
| R-PERF-04 | 并发用户数 ≥ 5000 | 系统同时在线用户 |
| R-PERF-05 | 每秒事务数 (TPS) ≥ 1000 | 核心交易接口 |
| R-PERF-06 | 页面加载时间 ≤ 3s | 首屏加载 |
| R-PERF-07 | **翻译文件首屏加载 ≤ 30KB** | **按页面/模块拆分翻译，按需加载** |
| R-PERF-08 | **语言切换 DOM 更新 ≤ 200ms** | **不触发整页刷新** |

### 7.2 可用性需求

| 需求 | 指标 | 说明 |
|------|------|------|
| R-AVAIL-01 | 系统可用率 ≥ 99.5% | 全年计划外停机 ≤ 43 小时 |
| R-AVAIL-02 | Token 中转成功率 ≥ 99% | — |
| R-AVAIL-03 | 故障恢复时间 ≤ 30 分钟 | 从故障发现到恢复 |
| R-AVAIL-04 | 渠道切换时间 ≤ 30 秒 | 自动检测故障并切换备用渠道 |

### 7.3 安全需求

| 需求 | 指标 | 说明 |
|------|------|------|
| R-SEC-01 | 传输加密：HTTPS（TLS 1.3） | 全站强制 HTTPS |
| R-SEC-02 | 密码存储：BCrypt 加密 | 不存储明文密码 |
| R-SEC-03 | 敏感数据存储：AES-256 | Token 信息、交易数据、身份证号等 |
| R-SEC-04 | 数据脱敏：手机号隐藏中间 4 位，邮箱部分隐藏 | 默认展示脱敏数据 |
| R-SEC-05 | 审计日志覆盖率 100% | 所有关键操作记录日志 |

### 7.4 可扩展性需求

| 需求 | 说明 |
|------|------|
| R-EXT-01 | 支持新增 Token 类型，无需修改核心代码 |
| R-EXT-02 | 支持接入新供应商，通过插件化或配置化方式扩展 |
| R-EXT-03 | 业务模块支持水平扩展（微服务架构） |
| R-EXT-04 | 预留标准 API 接口，便于与 ERP、CRM 等系统对接 |
| R-EXT-05 | **翻译文件按模块独立拆分，新增页面不影响现有翻译** |

### 7.5 兼容性需求

| 需求 | 说明 |
|------|------|
| R-COMP-01 | PC 端：Chrome / Firefox / Safari / Edge 最新两个主版本 |
| R-COMP-02 | 移动端：iOS Safari / Android Chrome 最新两个主版本 |
| R-COMP-03 | API 兼容 OpenAI 标准接口协议 |
| R-COMP-04 | **SSR 兼容：翻译文件需在服务端渲染时同步加载，避免客户端闪烁** |
| R-COMP-05 | **SEO 兼容：多语言页面需正确配置 `hreflang` 标签** |

### 7.6 数据备份与恢复

| 需求 | 说明 |
|------|------|
| R-BK-01 | 每日全量备份，每小时增量备份 |
| R-BK-02 | 异地备份存储 |
| R-BK-03 | 备份恢复测试每季度至少一次 |
| R-BK-04 | 数据丢失恢复目标：RPO ≤ 1 小时，RTO ≤ 4 小时 |

### 7.7 多语言可维护性需求

| 需求 | 说明 |
|------|------|
| R-i18n-MNT-01 | 翻译 Key 命名规范：`<模块>:<子模块>.<路径>.<描述>`，如 `token:product.buy_button` |
| R-i18n-MNT-02 | 每个翻译文件按页面模块独立拆分，不出现一个超大 JSON |

### 7.8 安全护栏需求

| 需求 | 说明 |
|------|------|
| R-GRDL-01 | PII 检测响应时间 ≤ 50ms（不显著增加调用延迟） |
| R-GRDL-02 | Prompt 注入检测响应时间 ≤ 100ms |
| R-GRDL-03 | 护栏规则支持热加载，修改后 ≤ 30s 生效 |
| R-GRDL-04 | 审计日志不可篡改，加密存储，保留 ≥ 12 个月 |
| R-GRDL-05 | 护栏系统具备 bypass 熔断：自身故障时不阻塞主请求 |

### 7.9 多协议兼容性需求

| 需求 | 说明 |
|------|------|
| R-PROTO-01 | API 需兼容 OpenAI、Anthropic、Gemini 三大原生协议 |
| R-PROTO-02 | 流式协议（SSE）在协议转换中完整保持，不丢失数据 |
| R-PROTO-03 | 协议检测基于请求路径自动识别，无需用户指定 |
| R-PROTO-04 | 模型后缀变体解析延迟 ≤ 5ms |

---

## 9. 数据需求

### 8.1 核心数据实体

| 实体 | 说明 | 主要字段 |
|------|------|----------|
| **User** | 用户账号 | user_id, username, password_hash, email, phone, role, status, level, created_at, **preferred_language** |
| **UserProfile** | 用户扩展信息 | user_id, avatar, company_name, business_license, real_name, id_number |
| **SubAccount** | 企业子账号 | sub_id, parent_id, email, token_quota, permissions, status |
| **TokenProduct** | Token 商品 | product_id, name, type, supplier_id, price, unit, description, status, **name_i18n (JSON)** |
| **TokenInventory** | Token 库存 | inventory_id, supplier_id, product_id, total_amount, remaining, alert_threshold |
| **Supplier** | Token 供应商 | supplier_id, name, api_endpoint, api_key_encrypted (AES-256-GCM), status, priority, weight, **region (domestic/overseas)**, group (用户组限制), model_mapping (JSON 模型映射) |
| **Ability** | 渠道能力索引（反范式） | group, model, channel_id, enabled, priority, weight — group+model+channel_id 复合唯一索引，快速查询用户组可用的模型渠道 |
| **UserToken** | 用户持有的 Token / API Key | ut_id, user_id, product_id, name, key (API Key 值, 唯一索引), amount, used_amount, frozen_amount, unlimited_quota, expires_at, status, created_at |
| **TokenTransfer** | Token 转让记录 | transfer_id, from_user, to_user, product_id, amount, status, created_at |
| **CallLog** | Token 调用日志 | log_id, user_id, token_id, token_name (冗余), channel_id, channel_name (冗余), model, request_path, prompt_tokens, completion_tokens, quota, route_decision (JSON 路由决策), response_time, status, ip, created_at |
| **Order** | 订单 | order_id, user_id, product_id, quantity, amount, currency, payment_method, status, created_at |
| **Payment** | 支付记录 | payment_id, order_id, amount, method, transaction_id, status, paid_at |
| **Refund** | 退款记录 | refund_id, order_id, amount, reason, status, operator_id, created_at |
| **RiskEvent** | 风控事件 | event_id, user_id, event_type, risk_level, description, status, handler_id, handled_at |
| **Commission** | 分销佣金 | commission_id, agent_id, customer_id, order_id, amount, rate, status, settled_at |
| **SupportedLanguage** | 支持的语言 | id, locale, name, native_name, is_enabled, sort_order, is_default |
| **SupplierVendor** | 入驻供应商 | vendor_id, user_id, company_name, contact_email, business_license, api_base_url, api_key_encrypted (AES-256-GCM), commission_rate, status, created_at |
| **SupplierProduct** | 供应商商品 | product_id, vendor_id, name, type, api_endpoint, auth_type, unit, price, stock_total, stock_remaining, status, **name_i18n (JSON)** |
| **Settlement** | 供应商结算 | settlement_id, vendor_id, period_start, period_end, total_sales, commission_amount, net_amount, status, paid_at |
| **BYOKKey** | 用户自有 API Key | key_id, user_id, provider (openai/anthropic/...), key_encrypted (AES-256-GCM), status, last_used_at, expires_at |
| **GuardrailRule** | 安全护栏规则 | rule_id, name, stage (before/after), type (pii/injection/secret/content), action (block/redact/warn), conditions (JSON), enabled |
| **GuardrailLog** | 护栏检测日志 | log_id, trace_id, user_id, rule_id, stage, detected_entities (JSON), action_taken, created_at |
| **SemanticCache** | 语义缓存条目 | cache_id, prompt_hash, prompt_vector, response (encrypted), model, hit_count, created_at, expires_at |
| **ModelVariant** | 模型后缀变体 | variant_id, base_model, suffix (nitro/floor/thinking), provider_id, price_coefficient, priority |
| **ProviderHealth** | 供应商健康历史 | health_id, provider_id, status, avg_latency_ms, error_rate, check_count, period_start, period_end |

### 8.2 数据存储方案

| 数据类型 | 存储方案 | 说明 |
|----------|----------|------|
| 核心业务数据（用户、订单、Token 等） | SQLite (WAL 模式) | 每微服务独立 DB 文件，WAL 支持并发读写 |
| 非结构化数据（日志、行为记录） | MongoDB | — |
| 高频缓存数据（Token 信息、热门商品） | Redis | 降低 DB 压力，支持分布式 |
| Token 临时存储 | Redis | 减少本地存储泄露风险 |
| 操作日志/审计日志 | Elasticsearch | 配合 ELK 进行日志分析 |
| **API Key 加密存储** | **AES-256-GCM + 每 Key 独立 IV** | 供应商 API Key 和用户 BYOK Key 在数据库中加密存储，应用层解密后使用，加密密钥由密钥管理服务 (KMS) 管理 |
| **翻译文件分发** | **CDN 静态资源** | **翻译文件编译为 JSON 静态资源部署至 CDN** |
| **语义缓存向量存储** | **pgvector / Redis Stack** | **用于语义缓存 Prompt 向量化存储和相似度检索** |
| **安全护栏规则存储** | Redis | 热加载护栏规则，支持 30s 内生效 |
| **供应商健康历史存储** | SQLite (每服务独立) | 可配置保留期（默认 90 天） |

### 8.3 语言配置数据设计

```sql
-- 平台支持的语言列表
CREATE TABLE `supported_languages` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `locale` VARCHAR(10) NOT NULL UNIQUE COMMENT '语言代码',
  `name` VARCHAR(50) NOT NULL COMMENT '语言名称（英文）',
  `native_name` VARCHAR(50) NOT NULL COMMENT '语言名称（母语）',
  `is_enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `sort_order` INT DEFAULT 0 COMMENT '排序',
  `is_default` TINYINT(1) DEFAULT 0 COMMENT '是否默认语言',
  `fallback_locale` VARCHAR(10) DEFAULT 'en' COMMENT '回退语言'
);

-- 渠道能力索引表（反范式设计，用于高速路由查询）
CREATE TABLE `ability_index` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `group_name` VARCHAR(64) NOT NULL COMMENT '用户分组',
  `model` VARCHAR(255) NOT NULL COMMENT '模型名称',
  `channel_id` INT NOT NULL COMMENT '渠道 ID',
  `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `priority` INT DEFAULT 0 COMMENT '优先级（数值大优先）',
  `weight` INT DEFAULT 0 COMMENT '权重（同优先级内随机选择）',
  UNIQUE INDEX `idx_group_model_channel` (`group_name`, `model`, `channel_id`),
  INDEX `idx_channel_id` (`channel_id`)
) COMMENT='Ability 反范式索引：group+model+channel 复合索引，路由引擎免计算直接查询匹配渠道';

-- Token 调用日志（高写入表，按月归档）
CREATE TABLE `call_log` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `user_id` INT NOT NULL COMMENT '用户 ID',
  `token_id` INT DEFAULT NULL COMMENT '令牌 ID',
  `token_name` VARCHAR(255) DEFAULT NULL COMMENT '令牌名称（冗余）',
  `channel_id` INT DEFAULT NULL COMMENT '渠道 ID',
  `channel_name` VARCHAR(255) DEFAULT NULL COMMENT '渠道名称（冗余）',
  `model` VARCHAR(255) NOT NULL COMMENT '调用的模型',
  `prompt_tokens` INT DEFAULT 0 COMMENT '输入 Token 数',
  `completion_tokens` INT DEFAULT 0 COMMENT '输出 Token 数',
  `quota` DECIMAL(20,4) DEFAULT 0 COMMENT '消耗配额',
  `route_decision` TEXT COMMENT '路由决策 JSON（含选中渠道、优先级、权重）',
  `request_time` INT DEFAULT 0 COMMENT '请求耗时 (ms)',
  `ip` VARCHAR(64) DEFAULT NULL COMMENT '请求 IP',
  `status` TINYINT(1) DEFAULT 1 COMMENT '请求状态',
  `created_time` BIGINT NOT NULL COMMENT '创建时间',
  INDEX `idx_user_created` (`user_id`, `created_time`),
  INDEX `idx_channel_id` (`channel_id`),
  INDEX `idx_trace_id` (`id`)
) COMMENT='Token 调用日志（高写入，按月分区归档）';
```

---

