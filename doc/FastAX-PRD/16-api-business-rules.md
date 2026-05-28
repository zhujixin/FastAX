## 10. 接口需求

### 9.1 用户端 API（对外）

| 接口 | 说明 | 兼容协议 |
|------|------|----------|
| `/v1/chat/completions` | AI 模型对话补全 | OpenAI 兼容 |
| `/v1/models` | 获取可用模型列表 | OpenAI 兼容 |
| `/v1/completions` | 文本补全 | OpenAI 兼容 |
| `/v1/embeddings` | 向量嵌入 | OpenAI 兼容 |
| `/v1/audio/transcriptions` | 语音转文本 | OpenAI 兼容 |
| `/v1/audio/translations` | 语音翻译 | OpenAI 兼容 |
| `/v1/images/generations` | 图片生成 | OpenAI 兼容 |

### 9.2 业务 API

| 接口分组 | 主要接口 |
|----------|----------|
| **用户服务** | POST /api/register, POST /api/login, POST /api/refresh, GET/PUT /api/profile, POST /api/password/reset |
| **Token 服务** | GET /api/tokens/products, GET /api/tokens/my, POST /api/tokens/buy, POST /api/tokens/transfer, POST /api/tokens/extract |
| **订单服务** | GET /api/orders, GET /api/orders/:id, POST /api/orders/:id/refund |
| **通知服务** | GET /api/notifications, PUT /api/notifications/:id/read |
| **统计服务** | GET /api/stats/usage, GET /api/stats/consumption, GET /api/stats/bills |
| **管理后台** | GET/POST/PUT/DELETE 各管理端接口（用户管理/Token 管控/交易管理/风控/系统） |
| **多语言服务** | GET /api/i18n/languages（获取可用语言列表）, GET /api/i18n/translations/:locale（获取翻译文件） |
| **供应商服务** | POST /api/vendor/register（入驻申请）, GET/PUT /api/vendor/profile（店铺管理）, GET/POST/PUT /api/vendor/products（商品管理）, GET /api/vendor/sales（销售看板）, GET /api/vendor/settlements（结算管理） |
| **管理后台-供应商** | GET/POST/PUT /api/admin/vendors（供应商审核与管理）, GET/PUT /api/admin/vendor-commission-rates（佣金配置） |
| **多协议原生** | POST /v1/messages（Anthropic 兼容）, POST /v1/chat/completions（OpenAI 兼容）, GET /v1/models（模型列表+变体）, GET /models/:variant（变体详情） |
| **BYOK 服务** | GET/POST/PUT /api/byok/keys（管理自有 Key）, GET /api/byok/usage（BYOK 用量统计）, PUT /api/byok/preference（路由优先级设置） |
| **安全护栏** | GET/POST/PUT /api/admin/guardrails/rules（护栏规则管理）, GET /api/admin/guardrails/logs（检测日志）, PUT /api/admin/guardrails/config（全局配置） |
| **成本优化** | GET /api/admin/cache/stats（语义缓存统计）, PUT /api/admin/cache/config（缓存策略配置）, PUT /api/user/budget（用户预算设置）, GET /api/user/cost-alerts（告警配置） |
| **模型市场** | GET /api/models/compare（模型对比）, GET /api/providers/health（供应商健康面板）, GET /api/models/benchmarks（基准数据） |
| **企业功能** | POST /api/admin/sso/config（SSO 配置）, GET/POST /api/admin/teams（团队管理）, GET /api/admin/audit/export（审计导出） |
| **多模态** | POST /v1/images/generations（图片）, POST /v1/audio/speech（TTS）, POST /v1/audio/transcriptions（STT）, POST /v1/video/generations（视频） |

### 9.3 第三方对接接口

| 对接方 | 接口说明 |
|--------|----------|
| Token 供应商（海外） | 批量采购、库存同步、价格同步、调用转发（OpenAI、Claude、Gemini 等） |
| Token 供应商（国内） | 批量采购、库存同步、价格同步、API 兼容适配（DeepSeek、Qwen、GLM、文心一言等） |
| 微信支付 | 支付下单、支付回调、退款 |
| 支付宝 | 支付下单、支付回调、退款 |
| Stripe | 海外支付下单、支付回调（美元结算，海外用户必需） |
| 阿里云/腾讯云短信 | 短信发送 |
| SendCloud | 邮件发送（含海外邮箱验证码） |
| 公安联网核查 | 用户身份核验 |
| 反洗钱数据库 | 交易合规检测 |
| **Anthropic API** | **Anthropic 原生协议 `/v1/messages` 转发** |
| **Google Gemini API** | **Google Gemini 原生协议转发** |
| **Midjourney / Flux** | **图片生成 API 对接** |
| **ElevenLabs** | **语音合成（TTS）API 对接** |
| **SAML/OIDC IdP** | **企业 SSO 身份提供商对接（Okta、Azure AD、Google Workspace）** |
| **Lakera Guard / Pangea** | **第三方安全护栏 API 对接（可选增强）** |

---

## 11. 业务规则

### 10.1 用户规则

| 规则 | 说明 |
|------|------|
| R-BIZ-01 | 未登录用户仅可浏览 Token 商城首页，不可下单和调用 |
| R-BIZ-02 | 普通用户注册无需审核，企业用户注册需资质审核 |
| R-BIZ-03 | 企业资质审核超时 48 小时自动通知管理员 |
| R-BIZ-04 | 连续 5 次登录失败，账号锁定 15 分钟 |
| R-BIZ-05 | 用户等级分为：普通、企业、代理，等级影响价格折扣和功能权限 |

### 10.2 Token 规则

| 规则 | 说明 |
|------|------|
| R-BIZ-10 | Token 购买支付成功后实时到账 |
| R-BIZ-11 | Token 转让时，转入方需与转出方为同一平台用户 |
| R-BIZ-12 | Token 到期前 7/3/1 天发送提醒 |
| R-BIZ-13 | 已过期的 Token 不可使用，不可退款 |
| R-BIZ-14 | Token 提取需验证用户身份（短信/Google 验证码） |
| R-BIZ-15 | **两阶段配额消费**：请求转发前先按输入 Token 数估算预扣（预消费），响应完成后按实际用量计算差值多退少补（后消费），防止超用 | 
| R-BIZ-16 | **配额计算公式**：`实际扣费 = 分组倍率 × 模型倍率 × (PromptToken数 + CompletionToken数 × 补全倍率)`，其中分组倍率按用户组设定，模型倍率按模型固定，补全倍率针对 Completion Token 额外加权 |
| R-BIZ-17 | **批量更新策略**：小额配额更新（单次 < 阈值）先缓存到 Redis，达到批量阈值（如 100 条或 10s 间隔）后批量刷入数据库，减少高频写入压力 |

### 10.3 交易规则

| 规则 | 说明 |
|------|------|
| R-BIZ-20 | 待支付订单超时 30 分钟自动取消 |
| R-BIZ-21 | 已支付订单在 Token 未使用的情况下可申请退款 |
| R-BIZ-22 | 已消耗的 Token 不支持退款 |
| R-BIZ-23 | 退款审批权限：单笔 ≤ 1000 元运营主管审批，> 1000 元需上级审批 |
| R-BIZ-24 | 交易记录保留 ≥ 5 年（财务审计要求） |

### 10.4 风控规则

| 规则 | 说明 |
|------|------|
| R-BIZ-30 | 同一 IP 每日注册账号数 ≤ 3 |
| R-BIZ-31 | 同一账号每日最大购买金额依据等级设定 |
| R-BIZ-32 | API 调用频率：普通用户 ≤ 60 次/分钟，企业用户 ≤ 300 次/分钟 |
| R-BIZ-33 | 异常交易触发条件：单笔 ≥ 5000 元 / 日交易 ≥ 10 笔 / 异地登录后立即交易 |
| R-BIZ-34 | 风控预警三级分级：黄色（提醒）/ 橙色（限频）/ 红色（冻结） |

### 10.5 多语言规则

| 规则 | 说明 |
|------|------|
| R-BIZ-40 | 未登录用户的语言由浏览器 `Accept-Language` 和 IP 地域决定 |
| R-BIZ-41 | 已登录用户的语言以其个人设置优先，忽略浏览器语言 |
| R-BIZ-42 | 语言回退链依次为：目标语言 → 回退语言 → 默认语言（zh-CN） |
| R-BIZ-43 | 翻译缺失时在前端不显示 Key 名称，应回退显示默认语言文案 |
| R-BIZ-44 | 通知类消息始终以接收者设置的接收语言发送 |
| R-BIZ-45 | 海外用户可仅使用邮箱注册，无需绑定手机号 |
| R-BIZ-46 | 国内大模型对海外用户开放时，需在购买前展示合规说明和免责条款 |
| R-BIZ-47 | 国内供应商渠道故障时，优先切换至同模型备用渠道，而非跨模型切换 |
| R-BIZ-48 | 供应商入驻需提交有效资质文件，经管理员审核后方可开通销售权限 |
| R-BIZ-49 | 供应商商品上架后需经过平台合规审核方可对用户展示 |
| R-BIZ-50 | 供应商价格不得低于平台设定的最低限价，不得高于最高限价 |
| R-BIZ-51 | 供应商结算周期默认为 T+7，平台可根据供应商信誉调整结算周期 |
| R-BIZ-52 | 供应商 API 连续 5 分钟不可用时，平台自动下架该商品并通知供应商 |
| R-BIZ-53 | 平台佣金默认按交易额百分比计算，供应商入驻协议中明确约定 |
| R-BIZ-54 | 供应商商品售出后，用户调用产生的 Token 消耗由供应商 API 实时提供，平台不垫付 |
| R-BIZ-55 | 多协议原生接入时，协议转换不得改变请求语义和响应完整性 |
| R-BIZ-56 | 模型后缀变体（:nitro/:floor）选择由路由引擎根据实时数据动态决定 |
| R-BIZ-57 | 安全护栏默认对所有用户生效，企业用户可配置自定义护栏规则 |
| R-BIZ-58 | 护栏处于 Monitor（告警）模式时不阻断请求但不豁免违规责任 |
| R-BIZ-59 | BYOK 模式下，用户的 API Key 仅用于转发调用，平台不得用于其他目的 |
| R-BIZ-60 | BYOK 调用失败时（Key 无效/过期/配额耗尽），可回退到平台 Token 但需用户授权 |
| R-BIZ-61 | 语义缓存命中仅返回缓存内容，不触发上游计费，但平台可收取缓存服务费 |
| R-BIZ-62 | 预算封顶到达时自动熔断全部非紧急调用，仅管理员可解除 |
| R-BIZ-63 | 企业 SSO 用户首次登录时自动创建平台账号并关联企业身份 |
| R-BIZ-64 | 模型市场的供应商健康数据每小时更新一次，延迟 ≤ 5 分钟 |
| R-BIZ-65 | 用户可随时导出个人数据（含调用记录、消费明细、护栏日志） |

---

