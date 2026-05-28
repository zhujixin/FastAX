## 17. 附录：需求注册表

Flat lookup table for all 170+ requirement IDs. Each entry links to the source file for full details.


#### 用户模块 (F-REG, F-ACC)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-REG-01 | P0 | 06-features/01-user-module.md | 手机号/邮箱+密码+验证码注册 |
| F-REG-02 | P1 | 06-features/01-user-module.md | 企业用户注册需上传资质，管理员审核 |
| F-REG-03 | P0 | 06-features/01-user-module.md | 账号密码/验证码登录，失败锁定 |
| F-REG-04 | P1 | 06-features/01-user-module.md | 微信/支付宝/Google OAuth 登录 |
| F-REG-05 | P0 | 06-features/01-user-module.md | JWT Token 认证，24h 过期，刷新机制 |
| F-REG-06 | P1 | 06-features/01-user-module.md | 异地登录检测与提醒 |
| F-REG-07 | P0 | 06-features/01-user-module.md | 海外用户仅邮箱+密码注册 |
| F-REG-08 | P0 | 06-features/01-user-module.md | 国际验证码邮件发送 |
| F-ACC-01 | P1 | 06-features/01-user-module.md | 修改个人信息 |
| F-ACC-02 | P0 | 06-features/01-user-module.md | 重置密码 |
| F-ACC-03 | P2 | 06-features/01-user-module.md | 绑定/解绑第三方账号 |
| F-ACC-04 | P1 | 06-features/01-user-module.md | 企业子账号管理与限额 |
| F-ACC-05 | P2 | 06-features/01-user-module.md | 查看用户等级和权限 |
| F-ACC-06 | P0 | 06-features/01-user-module.md | 管理员进行用户等级管理 |
| F-ACC-07 | P0 | 06-features/01-user-module.md | 违规用户警告/冻结/注销 |

#### Token 代理模块 (F-TKN, F-PXY, F-TUS, OCN, OCN-SUP, SUP, ROUTE)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-TKN-01 | P0 | 06-features/02-token-proxy-module.md | 对接全球主流 Token 供应商 |
| F-TKN-02 | P2 | 06-features/02-token-proxy-module.md | 多种 Token 类型接入 |
| F-TKN-03 | P0 | 06-features/02-token-proxy-module.md | 管理员查看供应商库存 |
| F-TKN-04 | P0 | 06-features/02-token-proxy-module.md | 库存预警通知 |
| F-TKN-05 | P1 | 06-features/02-token-proxy-module.md | Token 批量导入/导出 |
| F-TKN-06 | P0 | 06-features/02-token-proxy-module.md | 多渠道配置与负载均衡 |
| F-TKN-07 | P0 | 06-features/02-token-proxy-module.md | 渠道健康检测（5分钟/次） |
| F-TKN-08 | P0 | 06-features/02-token-proxy-module.md | 管理员手动启用/禁用渠道 |
| F-TKN-09 | P0 | 06-features/02-token-proxy-module.md | 销售价格与折扣设置 |
| F-TKN-10 | P0 | 06-features/02-token-proxy-module.md | 价格透明展示 |
| F-PXY-01 | P0 | 06-features/02-token-proxy-module.md | 标准化 OpenAI 兼容接口 |
| F-PXY-02 | P0 | 06-features/02-token-proxy-module.md | 请求网关路由与加密传输 |
| F-PXY-03 | P0 | 06-features/02-token-proxy-module.md | 流式输出支持 (SSE/WebSocket) |
| F-PXY-04 | P1 | 06-features/02-token-proxy-module.md | 多协议兼容 (HTTP/WS/gRPC) |
| F-PXY-05 | P0 | 06-features/02-token-proxy-module.md | 双轨智能路由 |
| F-PXY-06 | P0 | 06-features/02-token-proxy-module.md | Token 实时验证 |
| F-PXY-07 | P0 | 06-features/02-token-proxy-module.md | Token 调用日志记录 |
| F-PXY-08 | P0 | 06-features/02-token-proxy-module.md | 请求速率限制 |
| F-PXY-09 | P0 | 06-features/02-token-proxy-module.md | 国内模型 API→OpenAI 兼容适配 |
| F-PXY-10 | P1 | 06-features/02-token-proxy-module.md | 国内模型文档与错误信息英文适配 |
| F-TUS-01 | P0 | 06-features/02-token-proxy-module.md | 查看持有 Token 与使用记录 |
| F-TUS-02 | P0 | 06-features/02-token-proxy-module.md | 购买/充值 Token 实时到账 |
| F-TUS-03 | P1 | 06-features/02-token-proxy-module.md | Token 提取 |
| F-TUS-04 | P1 | 06-features/02-token-proxy-module.md | Token 转让 |
| F-TUS-05 | P0 | 06-features/02-token-proxy-module.md | Token 到期提醒 |
| F-TUS-06 | P1 | 06-features/02-token-proxy-module.md | Token 续费 |
| F-TUS-07 | P1 | 06-features/02-token-proxy-module.md | Token 托管服务 (AES-256) |
| F-TUS-08 | P1 | 06-features/02-token-proxy-module.md | 企业子账号额度控制 |
| OCN-01 | P0 | 06-features/02-token-proxy-module.md | 国内节点部署优化海外访问链路 |
| OCN-02 | P0 | 06-features/02-token-proxy-module.md | 国内模型分类与展示（英文） |
| OCN-03 | P0 | 06-features/02-token-proxy-module.md | 国内模型美元价格展示 |
| OCN-04 | P1 | 06-features/02-token-proxy-module.md | 国内模型优劣势英文说明 |
| OCN-05 | P0 | 06-features/02-token-proxy-module.md | 跨境合规提示 |
| OCN-06 | P1 | 06-features/02-token-proxy-module.md | 国内模型调用量独立统计 |
| OCN-SUP-01 | P0 | 06-features/02-token-proxy-module.md | 供应商区域分类管理 |
| OCN-SUP-02 | P0 | 06-features/02-token-proxy-module.md | 国内供应商余额监控 |
| OCN-SUP-03 | P0 | 06-features/02-token-proxy-module.md | 国内供应商渠道冗余配置 |
| OCN-SUP-04 | P1 | 06-features/02-token-proxy-module.md | 国内模型兼容度报告 |
| SUP-01 | P0 | 06-features/02-token-proxy-module.md | 供应商注册申请 |
| SUP-02 | P0 | 06-features/02-token-proxy-module.md | 供应商资质审核 |
| SUP-03 | P0 | 06-features/02-token-proxy-module.md | 供应商店铺管理 |
| SUP-04 | P0 | 06-features/02-token-proxy-module.md | 商品上架与管理 |
| SUP-05 | P0 | 06-features/02-token-proxy-module.md | 自主定价 |
| SUP-06 | P1 | 06-features/02-token-proxy-module.md | 价格策略（折扣/促销） |
| SUP-07 | P0 | 06-features/02-token-proxy-module.md | 供应商 API 注册 |
| SUP-08 | P0 | 06-features/02-token-proxy-module.md | 供应商库存管理 |
| SUP-09 | P0 | 06-features/02-token-proxy-module.md | 销售看板 |
| SUP-10 | P0 | 06-features/02-token-proxy-module.md | 结算管理 |
| SUP-11 | P1 | 06-features/02-token-proxy-module.md | 供应商通知 |
| SUP-12 | P0 | 06-features/02-token-proxy-module.md | 供应商费率配置 |
| SUP-13 | P2 | 06-features/02-token-proxy-module.md | 供应商评级与评价 |
| SUP-14 | P0 | 06-features/02-token-proxy-module.md | 供应商违规处理 |
| SUP-15 | P0 | 06-features/02-token-proxy-module.md | 供应商 API 健康监控 |
| SUP-16 | P1 | 06-features/02-token-proxy-module.md | 供应商入驻协议在线签署 |
| SUP-17 | P0 | 06-features/02-token-proxy-module.md | Adaptor 适配器接口（9 方法，参考 one-api relay/adaptor/interface.go） |
| SUP-18 | P1 | 06-features/02-token-proxy-module.md | 渠道测试端点（参考 one-api TestChannel） |
| SUP-19 | P1 | 06-features/02-token-proxy-module.md | 渠道类型双枚举 channeltype+apitype |
| ROUTE-01 | P0 | 06-features/02-token-proxy-module.md | 渠道选择算法（Ability 索引 + 优先级分组+随机） |
| ROUTE-02 | P0 | 06-features/02-token-proxy-module.md | 加权随机选择（同优先级组内权重随机） |
| ROUTE-03 | P0 | 06-features/02-token-proxy-module.md | 自动故障转移 (Failover) |
| ROUTE-04 | P0 | 06-features/02-token-proxy-module.md | 轻量熔断（5xx/超时自动禁用，排除 401/403/429） |
| ROUTE-05 | P0 | 06-features/02-token-proxy-module.md | 实时健康检测（10秒/次） |
| ROUTE-06 | P0 | 06-features/02-token-proxy-module.md | 延迟敏感路由 |
| ROUTE-07 | P1 | 06-features/02-token-proxy-module.md | 成本优化路由 |
| ROUTE-08 | P1 | 06-features/02-token-proxy-module.md | 时段性调度 |
| ROUTE-09 | P0 | 06-features/02-token-proxy-module.md | 请求重写与透明转发 |
| ROUTE-10 | P1 | 06-features/02-token-proxy-module.md | 模型自动发现 |
| ROUTE-11 | P0 | 06-features/02-token-proxy-module.md | 请求级日志与全链路追踪 |
| ROUTE-12 | P0 | 06-features/02-token-proxy-module.md | 流式请求故障转移 |
| ROUTE-13 | P0 | 06-features/02-token-proxy-module.md | 限流与排队 |
| ROUTE-14 | P1 | 06-features/02-token-proxy-module.md | 灰度路由 |
| ROUTE-15 | P1 | 06-features/02-token-proxy-module.md | 供应商配额展示 |
| ROUTE-16 | P0 | 06-features/02-token-proxy-module.md | 路由规则热加载（SyncChannelCache 定时刷新） |
| ROUTE-17 | P0 | 06-features/02-token-proxy-module.md | Ability 索引表+内存缓存（group+model+channel 复合索引） |

#### 交易模块 (F-ORD, F-PAY)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-ORD-01 | P0 | 06-features/03-trade-module.md | 订单自动生成 |
| F-ORD-02 | P0 | 06-features/03-trade-module.md | 订单状态流转 |
| F-ORD-03 | P0 | 06-features/03-trade-module.md | 待支付订单超时取消 |
| F-ORD-04 | P1 | 06-features/03-trade-module.md | 用户申请退款 |
| F-ORD-05 | P0 | 06-features/03-trade-module.md | 订单查询 |
| F-ORD-06 | P0 | 06-features/03-trade-module.md | 管理员订单管理 |
| F-PAY-01 | P0 | 06-features/03-trade-module.md | 微信/支付宝/银行卡支付 |
| F-PAY-02 | P0 | 06-features/03-trade-module.md | Stripe 海外支付 |
| F-PAY-03 | P0 | 06-features/03-trade-module.md | 支付加密与状态同步 |
| F-PAY-04 | P1 | 06-features/03-trade-module.md | 自动对账报表 |
| F-PAY-05 | P1 | 06-features/03-trade-module.md | 对账报表导出 |
| F-PAY-06 | P2 | 06-features/03-trade-module.md | 分销佣金系统 |
| F-PAY-07 | P1 | 06-features/03-trade-module.md | 手续费规则配置 |
| F-PAY-08 | P0 | 06-features/03-trade-module.md | 供应商结算管理 |
| F-PAY-09 | P0 | 06-features/03-trade-module.md | 平台佣金计算 |

#### 风控模块 (F-RISK, F-SEC)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-RISK-01 | P1 | 06-features/04-risk-control.md | 设备指纹识别 |
| F-RISK-02 | P0 | 06-features/04-risk-control.md | AI 风控引擎实时分析 |
| F-RISK-03 | P1 | 06-features/04-risk-control.md | 自定义风控规则 |
| F-RISK-04 | P0 | 06-features/04-risk-control.md | Token 防刷保护 |
| F-RISK-05 | P0 | 06-features/04-risk-control.md | 异常事件预警与分级处置 |
| F-SEC-01 | P0 | 06-features/04-risk-control.md | 账号防盗与二次验证 |
| F-SEC-02 | P0 | 06-features/04-risk-control.md | 密码复杂度要求 |
| F-SEC-03 | P0 | 06-features/04-risk-control.md | 数据加密 (BCrypt/AES-256) |
| F-SEC-04 | P1 | 06-features/04-risk-control.md | 接口签名验证 |
| F-SEC-05 | P0 | 06-features/04-risk-control.md | 接口限流 |
| F-SEC-06 | P0 | 06-features/04-risk-control.md | 接口权限校验 (JWT) |

#### 统计分析 (F-STAT)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-STAT-01 | P0 | 06-features/05-statistics.md | 核心数据统计看板 |
| F-STAT-02 | P2 | 06-features/05-statistics.md | 用户行为分析 |
| F-STAT-03 | P1 | 06-features/05-statistics.md | Token 用量统计与趋势 |
| F-STAT-04 | P1 | 06-features/05-statistics.md | 交易报表 |
| F-STAT-05 | P2 | 06-features/05-statistics.md | 用户分层与运营支持 |
| F-STAT-06 | P1 | 06-features/05-statistics.md | 系统性能监控报表 |
| F-STAT-07 | P1 | 06-features/05-statistics.md | 多语言翻译覆盖率统计 |

#### 通知模块 (F-NOT)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-NOT-01 | P0 | 06-features/06-notification.md | 站内信通知 |
| F-NOT-02 | P0 | 06-features/06-notification.md | 短信通知 |
| F-NOT-03 | P1 | 06-features/06-notification.md | 邮件通知 |
| F-NOT-04 | P2 | 06-features/06-notification.md | 通知模板管理与配置 |
| F-NOT-05 | P0 | 06-features/06-notification.md | 通知模板多语言 |

#### 运营工具 (F-OP)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-OP-01 | P2 | 06-features/07-operations.md | 优惠券发放与管理 |
| F-OP-02 | P2 | 06-features/07-operations.md | 限时活动配置 |
| F-OP-03 | P2 | 06-features/07-operations.md | 邀请有礼功能 |
| F-OP-04 | P2 | 06-features/07-operations.md | 活动数据统计与分析 |

#### 管理后台 (F-ADM)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| F-ADM-01 | P0 | 06-features/08-admin-panel.md | 控制台运营概览 |
| F-ADM-02 | P0 | 06-features/08-admin-panel.md | 用户管理列表 |
| F-ADM-03 | P0 | 06-features/08-admin-panel.md | Token 管理 |
| F-ADM-04 | P0 | 06-features/08-admin-panel.md | 交易管理 |
| F-ADM-05 | P0 | 06-features/08-admin-panel.md | 风控管理 |
| F-ADM-06 | P1 | 06-features/08-admin-panel.md | 系统管理 |
| F-ADM-07 | P1 | 06-features/08-admin-panel.md | 操作日志 |
| F-ADM-08 | P0 | 06-features/08-admin-panel.md | 多语言支持配置 |
| F-ADM-09 | P2 | 06-features/08-admin-panel.md | 多货币支持配置 |

#### 多语言模块 (LANG)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| LANG-01-01 | P0 | 06-features/09-i18n-module.md | 浏览器语言自动检测 |
| LANG-01-02 | P0 | 06-features/09-i18n-module.md | 语言选择器 |
| LANG-01-03 | P0 | 06-features/09-i18n-module.md | 语言持久化 |
| LANG-01-04 | P1 | 06-features/09-i18n-module.md | 登录态多设备同步 |
| LANG-01-05 | P2 | 06-features/09-i18n-module.md | URL 路径语言标识 |
| LANG-01-06 | P0 | 06-features/09-i18n-module.md | 语言回退策略 |
| LANG-01-07 | P1 | 06-features/09-i18n-module.md | 语言选择器推荐语言 |
| LANG-02-01 | P0 | 06-features/09-i18n-module.md | 静态文案翻译 (t()) |
| LANG-02-02 | P0 | 06-features/09-i18n-module.md | 动态文案本地化 |
| LANG-02-03 | P0 | 06-features/09-i18n-module.md | 占位符与插值 |
| LANG-02-04 | P1 | 06-features/09-i18n-module.md | 复数形式支持 |
| LANG-02-05 | P0 | 06-features/09-i18n-module.md | 日期/时间本地化 |
| LANG-02-06 | P1 | 06-features/09-i18n-module.md | 数字本地化 |
| LANG-02-07 | P1 | 06-features/09-i18n-module.md | 货币格式本地化 |
| LANG-02-08 | P2 | 06-features/09-i18n-module.md | RTL 布局预留 |
| LANG-03-01 | P0 | 06-features/09-i18n-module.md | 邮件模板多语言 |
| LANG-03-02 | P0 | 06-features/09-i18n-module.md | 短信模板多语言 |
| LANG-03-03 | P0 | 06-features/09-i18n-module.md | 站内信多语言 |
| LANG-03-04 | P1 | 06-features/09-i18n-module.md | 模板语言选择策略 |
| LANG-04-01 | P0 | 06-features/09-i18n-module.md | 语言列表配置 |
| LANG-04-02 | P0 | 06-features/09-i18n-module.md | 语言切换展示 |
| LANG-05-01 | P0 | 06-features/09-i18n-module.md | 开发者文档多语言 |
| LANG-05-02 | P1 | 06-features/09-i18n-module.md | API 错误消息多语言 |
| LANG-05-03 | P1 | 06-features/09-i18n-module.md | API 响应多语言字段 |
| LANG-05-04 | P1 | 06-features/09-i18n-module.md | 代码示例多语言 |
| LANG-06-01 | P1 | 06-features/09-i18n-module.md | FAQ 多语言 |
| LANG-06-02 | P1 | 06-features/09-i18n-module.md | 工单系统语言标识 |
| LANG-06-03 | P2 | 06-features/09-i18n-module.md | 预置回复多语言 |

#### 非功能需求 (R-*)

| ID | Source | Description |
|----|--------|-------------|
| R-PERF-01 | 07-non-functional.md | 接口响应 P95 ≤ 500ms |
| R-PERF-02 | 07-non-functional.md | Token 中转延迟 P95 ≤ 1000ms |
| R-PERF-03 | 07-non-functional.md | 流式输出首包 ≤ 500ms |
| R-PERF-04 | 07-non-functional.md | 并发用户 ≥ 5000 |
| R-PERF-05 | 07-non-functional.md | TPS ≥ 1000 |
| R-PERF-06 | 07-non-functional.md | 页面加载 ≤ 3s |
| R-PERF-07 | 07-non-functional.md | 翻译文件首屏 ≤ 30KB |
| R-PERF-08 | 07-non-functional.md | 语言切换 DOM 更新 ≤ 200ms |
| R-AVAIL-01 | 07-non-functional.md | 系统可用率 ≥ 99.5% |
| R-AVAIL-02 | 07-non-functional.md | Token 中转成功率 ≥ 99% |
| R-AVAIL-03 | 07-non-functional.md | 故障恢复 ≤ 30 分钟 |
| R-AVAIL-04 | 07-non-functional.md | 渠道切换 ≤ 30 秒 |
| R-SEC-01 | 07-non-functional.md | HTTPS TLS 1.3 |
| R-SEC-02 | 07-non-functional.md | BCrypt 密码存储 |
| R-SEC-03 | 07-non-functional.md | AES-256 敏感数据 |
| R-SEC-04 | 07-non-functional.md | 数据脱敏 |
| R-SEC-05 | 07-non-functional.md | 审计日志 100% |
| R-EXT-01 | 07-non-functional.md | 支持新增 Token 类型 |
| R-EXT-02 | 07-non-functional.md | 插件化供应商扩展 |
| R-EXT-03 | 07-non-functional.md | 水平扩展 (微服务) |
| R-EXT-04 | 07-non-functional.md | 标准 API 对接 |
| R-EXT-05 | 07-non-functional.md | 翻译文件独立拆分 |
| R-COMP-01 | 07-non-functional.md | PC 浏览器兼容 |
| R-COMP-02 | 07-non-functional.md | 移动端浏览器兼容 |
| R-COMP-03 | 07-non-functional.md | OpenAI 接口兼容 |
| R-COMP-04 | 07-non-functional.md | SSR 翻译同步加载 |
| R-COMP-05 | 07-non-functional.md | SEO hreflang 标签 |
| R-BK-01 | 07-non-functional.md | 每日全量/小时增量备份 |
| R-BK-02 | 07-non-functional.md | 异地备份 |
| R-BK-03 | 07-non-functional.md | 季度恢复测试 |
| R-BK-04 | 07-non-functional.md | RPO ≤ 1h, RTO ≤ 4h |
| R-i18n-MNT-01 | 07-non-functional.md | 翻译 Key 命名规范 |
| R-i18n-MNT-02 | 07-non-functional.md | 翻译文件按模块拆分 |

#### 业务规则 (R-BIZ)

| ID | Source | Description |
|----|--------|-------------|
| R-BIZ-01 | 10-business-rules.md | 未登录仅可浏览 |
| R-BIZ-02 | 10-business-rules.md | 普通用户免审，企业需审 |
| R-BIZ-03 | 10-business-rules.md | 审核超时 48h 通知 |
| R-BIZ-04 | 10-business-rules.md | 5 次失败锁定 15 分钟 |
| R-BIZ-05 | 10-business-rules.md | 用户等级体系 |
| R-BIZ-10 | 10-business-rules.md | 支付成功实时到账 |
| R-BIZ-11 | 10-business-rules.md | Token 转让同平台限制 |
| R-BIZ-12 | 10-business-rules.md | 到期前 7/3/1 天提醒 |
| R-BIZ-13 | 10-business-rules.md | 过期 Token 不可用/退款 |
| R-BIZ-14 | 10-business-rules.md | Token 提取需身份验证 |
| R-BIZ-20 | 10-business-rules.md | 待支付订单 30 分钟取消 |
| R-BIZ-21 | 10-business-rules.md | 未使用 Token 可退款 |
| R-BIZ-22 | 10-business-rules.md | 已消耗 Token 不退 |
| R-BIZ-23 | 10-business-rules.md | 退款分级审批 |
| R-BIZ-24 | 10-business-rules.md | 交易记录保留 ≥ 5 年 |
| R-BIZ-30 | 10-business-rules.md | 同 IP 日注册 ≤ 3 |
| R-BIZ-31 | 10-business-rules.md | 日购买金额等级限制 |
| R-BIZ-32 | 10-business-rules.md | API 频率限制 |
| R-BIZ-33 | 10-business-rules.md | 异常交易触发条件 |
| R-BIZ-34 | 10-business-rules.md | 三级风控预警 |
| R-BIZ-40 | 10-business-rules.md | 未登录语言 = 浏览器 |
| R-BIZ-41 | 10-business-rules.md | 已登录语言 = 个人设置 |
| R-BIZ-42 | 10-business-rules.md | 语言回退链 |
| R-BIZ-43 | 10-business-rules.md | 翻译缺失回退不显示 Key |
| R-BIZ-44 | 10-business-rules.md | 通知按接收者语言发送 |
| R-BIZ-45 | 10-business-rules.md | 海外用户仅邮箱注册 |
| R-BIZ-46 | 10-business-rules.md | 合规说明预先展示 |
| R-BIZ-47 | 10-business-rules.md | 渠道故障同模型切换 |
| R-BIZ-48 | 10-business-rules.md | 供应商入驻资质审核 |
| R-BIZ-49 | 10-business-rules.md | 商品上架合规审核 |
| R-BIZ-50 | 10-business-rules.md | 供应商价格上下限 |
| R-BIZ-51 | 10-business-rules.md | 结算周期 T+7 |
| R-BIZ-52 | 10-business-rules.md | API 连续 5 分钟不可用自动下架 |
| R-BIZ-53 | 10-business-rules.md | 平台佣金按百分比 |
| R-BIZ-54 | 10-business-rules.md | 供应商实时提供 Token |

#### 合规需求 (R-COMPL)

| ID | Source | Description |
|----|--------|-------------|
| R-COMPL-01 | 11-compliance.md | 用户隐私告知同意 |
| R-COMPL-02 | 11-compliance.md | 账号注销与数据删除 |
| R-COMPL-03 | 11-compliance.md | 不收集无关信息 |
| R-COMPL-04 | 11-compliance.md | 公安联网身份核验 |
| R-COMPL-05 | 11-compliance.md | 反洗钱数据库对接 |
| R-COMPL-06 | 11-compliance.md | 审计日志 ≥ 6 个月 |
| R-COMPL-07 | 11-compliance.md | 合规数据库实时适配 |
| R-COMPL-08 | 11-compliance.md | 多语言多货币支持 |
| R-COMPL-09 | 11-compliance.md | 数据出境安全评估 |
| R-COMPL-10 | 11-compliance.md | 生成式 AI 合规 |
| R-COMPL-11 | 11-compliance.md | 英文用户协议/隐私政策 |
| R-COMPL-12 | 11-compliance.md | 国内模型使用限制标注 |

#### 多协议原生支持 (PROTO)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| PROTO-01 | P0 | 06-features/10-multi-protocol.md | Anthropic Messages API 原生支持 |
| PROTO-02 | P0 | 06-features/10-multi-protocol.md | Gemini API 原生支持 |
| PROTO-03 | P1 | 06-features/10-multi-protocol.md | OpenAI Realtime API 代理 |
| PROTO-04 | P0 | 06-features/10-multi-protocol.md | 协议自动检测与转换 |
| PROTO-05 | P1 | 06-features/10-multi-protocol.md | 模型后缀变体（Model Suffix） |
| PROTO-06 | P1 | 06-features/10-multi-protocol.md | 模型自动发现 |
| PROTO-07 | P1 | 06-features/10-multi-protocol.md | 模型重命名/别名 |
| PROTO-08 | P2 | 06-features/10-multi-protocol.md | Rerank 模型支持 |
| PROTO-09 | P1 | 06-features/10-multi-protocol.md | Embeddings 多供应商路由 |
| PROTO-10 | P1 | 06-features/10-multi-protocol.md | MCP 协议支持 |

#### 多模态支持 (MEDIA)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| MEDIA-01 | P1 | 06-features/11-multimodal.md | 图片生成统一 API |
| MEDIA-02 | P2 | 06-features/11-multimodal.md | 视频生成支持 |
| MEDIA-03 | P1 | 06-features/11-multimodal.md | 语音合成/识别统一 API |
| MEDIA-04 | P2 | 06-features/11-multimodal.md | 音乐生成支持 |
| MEDIA-05 | P1 | 06-features/11-multimodal.md | 多模态路由与故障转移 |
| MEDIA-06 | P1 | 06-features/11-multimodal.md | 多模态成本追踪 |
| MEDIA-07 | P1 | 06-features/11-multimodal.md | 媒体内容审核 |

#### 安全护栏 (GRDL)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| GRDL-01 | P0 | 06-features/12-guardrails.md | PII 检测与脱敏 |
| GRDL-02 | P0 | 06-features/12-guardrails.md | Prompt 注入检测 |
| GRDL-03 | P1 | 06-features/12-guardrails.md | 内容审核 |
| GRDL-04 | P0 | 06-features/12-guardrails.md | 密钥扫描 |
| GRDL-05 | P0 | 06-features/12-guardrails.md | 护栏执行模式 |
| GRDL-06 | P1 | 06-features/12-guardrails.md | 自定义护栏规则 |
| GRDL-07 | P0 | 06-features/12-guardrails.md | 不可篡改审计日志 |
| GRDL-08 | P0 | 06-features/12-guardrails.md | 护栏流水线 |
| GRDL-09 | P1 | 06-features/12-guardrails.md | 合规报告导出 |

#### BYOK (自带 Key)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| BYOK-01 | P1 | 06-features/13-byok.md | 用户管理自有 API Key |
| BYOK-02 | P1 | 06-features/13-byok.md | BYOK + 平台 Token 混合路由 |
| BYOK-03 | P1 | 06-features/13-byok.md | BYOK 用量看板 |
| BYOK-04 | P1 | 06-features/13-byok.md | Key 轮换与过期管理 |
| BYOK-05 | P2 | 06-features/13-byok.md | 团队 Key 共享 |
| BYOK-06 | P1 | 06-features/13-byok.md | BYOK 模型限制 |
| BYOK-07 | P1 | 06-features/13-byok.md | BYOK 平台费用 |

#### 插件/扩展系统 (PLUG)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| PLUG-01 | P2 | 06-features/14-plugin-system.md | 中间件流水线 |
| PLUG-02 | P2 | 06-features/14-plugin-system.md | 自定义路由策略插件 |
| PLUG-03 | P2 | 06-features/14-plugin-system.md | Webhook 转换 |
| PLUG-04 | P2 | 06-features/14-plugin-system.md | 限流策略插件 |
| PLUG-05 | P2 | 06-features/14-plugin-system.md | 监控插件接口 |
| PLUG-06 | P3 | 06-features/14-plugin-system.md | 插件市场 |
| PLUG-07 | P2 | 06-features/14-plugin-system.md | 插件沙箱隔离 |

#### 成本优化引擎 (COST)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| COST-01 | P1 | 06-features/15-cost-optimization.md | 上游 Prompt Caching |
| COST-02 | P2 | 06-features/15-cost-optimization.md | 语义缓存 |
| COST-03 | P1 | 06-features/15-cost-optimization.md | 缓存计费比率配置 |
| COST-04 | P0 | 06-features/15-cost-optimization.md | 预算封顶 |
| COST-05 | P0 | 06-features/15-cost-optimization.md | 成本告警 Webhook |
| COST-06 | P1 | 06-features/15-cost-optimization.md | 成本感知路由 |
| COST-07 | P1 | 06-features/15-cost-optimization.md | 模型回退链 |
| COST-08 | P2 | 06-features/15-cost-optimization.md | Token 压缩 |

#### 企业功能 (ENT)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| ENT-01 | P2 | 06-features/16-enterprise.md | SSO/SAML/OIDC 集成 |
| ENT-02 | P1 | 06-features/16-enterprise.md | 团队/项目隔离 |
| ENT-03 | P1 | 06-features/16-enterprise.md | 审计导出 |
| ENT-04 | P2 | 06-features/16-enterprise.md | 角色级预算控制 |
| ENT-05 | P1 | 06-features/16-enterprise.md | 预付费套餐 |
| ENT-06 | P1 | 06-features/16-enterprise.md | API 限速定制 |
| ENT-07 | P1 | 06-features/16-enterprise.md | 模型白名单/黑名单 |
| ENT-08 | P2 | 06-features/16-enterprise.md | 数据驻留控制 |

#### 模型市场与发现 (MKT)

| ID | Pri | Source | Description |
|----|-----|--------|-------------|
| MKT-01 | P1 | 06-features/17-model-marketplace.md | 模型对比工具 |
| MKT-02 | P1 | 06-features/17-model-marketplace.md | 供应商稳定性指标 |
| MKT-03 | P2 | 06-features/17-model-marketplace.md | 模型基准测试 |
| MKT-04 | P2 | 06-features/17-model-marketplace.md | 模型推荐引擎 |
| MKT-05 | P0 | 06-features/17-model-marketplace.md | 供应商健康公开面板 |
| MKT-06 | P1 | 06-features/17-model-marketplace.md | 模型变更日志 |

#### 非功能需求新增

| ID | Source | Description |
|----|--------|-------------|
| R-COMP-06 | 07-non-functional.md | Anthropic/Gemini/OpenAI 多协议兼容 |
| R-COMP-07 | 07-non-functional.md | MCP 协议兼容 |
| R-GRDL-01 | 07-non-functional.md | PII 检测 ≤ 50ms |
| R-GRDL-02 | 07-non-functional.md | Prompt 注入检测 ≤ 100ms |
| R-GRDL-03 | 07-non-functional.md | 护栏规则热加载 ≤ 30s |
| R-GRDL-04 | 07-non-functional.md | 审计日志加密保留 ≥ 12 月 |
| R-GRDL-05 | 07-non-functional.md | 护栏系统自身故障不阻塞主请求 |
| R-PROTO-01 | 07-non-functional.md | 三大原生协议兼容 |
| R-PROTO-02 | 07-non-functional.md | 流式协议完整保持 |
| R-PROTO-03 | 07-non-functional.md | 协议自动路径检测 |
| R-PROTO-04 | 07-non-functional.md | 模型后缀变体解析 ≤ 5ms |

---

