> **Domain**: `domain/proxy` + `domain/token` + `domain/vendor` — Token 代理/路由/供应商 | **PRD**: FastAX-PRD/02-token-proxy-vendor.md | **参考**: one-api relay/ + model/channel.go
### 5.2 Token 代理模块 (Proxy Service — 核心模块)

#### 5.2.1 智能路由策略 (PRD ROUTE-01, ROUTE-02, ROUTE-17, 参考 one-api Ability 索引)

**设计选择：优先级分组 + 同优先级权重随机**（替代实时加权评分公式）

参考 one-api (30.6K Stars) 生产验证的 `CacheGetRandomSatisfiedChannel` 策略，路由决策 O(1)，不实时计算多维分数。

```
                     ┌───────────────────────────────────────┐
                     │            Proxy Service               │
                     │                                       │
    User Request ──────→│  ┌─────────────────────────────┐  │
                     │  │  路由决策引擎 (内存缓存命中)      │  │
                     │  │                                │  │
                     │  │  1. 查内存缓存:                  │  │
                     │  │     group+model → []Channel     │  │
                     │  │     ( InitChannelCache 定时加载 ) │  │
                     │  │  2. 按 priority 分组             │  │
                     │  │  3. 最高优先级组内 rand 随机      │  │
                     │  │     (weight 影响被选中概率)       │  │
                     │  │  4. 健康检查过滤                  │  │
                     │  │  5. 返回最优渠道                  │  │
                     │  └─────────────────────────────┘  │
                     │                                       │
                     │  ┌─────────────────────────────┐  │
                     │  │  内存缓存同步                  │  │
                     │  │  · 启动全量加载                │  │
                     │  │  · 定时刷新 (60s)             │  │
                     │  │  · 渠道变更即时触发             │  │
                     │  │  · 数据源: ability_index 表    │  │
                     │  └─────────────────────────────┘  │
                     └───────────────────────────────────────┘
```

渠道选择算法（参考 one-api `CacheGetRandomSatisfiedChannel`）:

```
算法: 选择最优渠道

输入: group, model
输出: Channel | nil

1.  从内存缓存获取候选列表:
    channels = group2model2channels[group][model]
    if channels 为空 → return nil

2.  取最高优先级分组:
    firstPriority = channels[0].priority
    endIdx = 首个 priority != firstPriority 的位置
    // channels[0:endIdx] 即为最高优先级组

3.  权重随机选择:
    if 失败重试场景:
        idx = rand(endIdx, len(channels))  // 跳过最高优先级
    else:
        idx = rand(0, endIdx)              // 最高优先级组内随机

4.  return channels[idx]

注:
- 不实时计算 priority×0.4 + health×0.3 + ... 等公式
- health/latency 等指标通过健康检测影响渠道的启用/禁用状态
- 参考: one-api model/ability.go + model/cache.go
```

#### 5.2.2 熔断器机制 (PRD ROUTE-04, 参考 one-api monitor)

**设计选择：轻量级自动禁用**（替代完整 gobreaker 状态机）

参考 one-api 的 `monitor.ShouldDisableChannel` 实现，熔断逻辑更务实：

- 401/403/400 等客户端错误不触发熔断（认证问题不影响渠道健康）
- 429 限流不触发熔断（渠道本身正常）
- 5xx 连续错误触发渠道自动禁用
- 自动禁用后通过周期性健康检测恢复

```
熔断决策逻辑 (Relay 错误处理):

请求失败
    │
    ▼
┌──────────────────────────┐
│ shouldRetry(statusCode)  │
│                          │
│ if 401/403/400 → 不重试  │──→ 直接返回错误给用户
│ if 2xx          → 不重试  │
│ if 429/5xx      → 重试   │──→ 自动切换渠道重试
└──────────────────────────┘

┌──────────────────────────────────┐
│ processChannelRelayError          │
│                                  │
│ if ShouldDisableChannel(err):    │──→ DB 更新 channel.status=3 (自动禁用)
│     monitor.DisableChannel(id)   │     发送 event.channel.health-changed
│ else:                            │
│     monitor.Emit(id, false)      │──→ 更新错误率统计
└──────────────────────────────────┘

ShouldDisableChannel 判断逻辑:
  - statusCode == 401 || 403 → false (认证错，非渠道问题)
  - statusCode == 429       → false (限流，渠道正常)
  - statusCode == 400       → false (请求参数问题)
  - statusCode / 100 == 5   → true  (服务端错误，触发禁用)
  - statusCode == 0 (超时)  → 根据连续超时次数决定

配置参数:
  - retryTimes: 3            (失败重试次数)
  - autoDisableThreshold: 5  (5 分钟内连续失败 N 次则自动禁用)
  - autoDisableRate: 0.5     (错误率 > 50% 触发禁用)

恢复机制:
  - 自动禁用的渠道通过健康检测恢复 (§5.2.4)
  - 健康检测成功 → status=1 (启用) + 更新 ability_index
```

#### 5.2.3 流式代理机制 (PRD ROUTE-12 — 流式请求故障转移)

```
用户 Client              Proxy Service               上游 Supplier A           上游 Supplier B
    │                         │                          │                        │
    │── POST /v1/chat/completions ──→│                          │                        │
    │   (stream: true)    │                          │                        │
    │                         │── 路由选择 → Supplier A ──→│                        │
    │                         │                          │                        │
    │                         │←── SSE (非流式) ──────────│                        │
    │                         │    (开始流式输出)          │                        │
    │                         │                          │                        │
    │←── 转发 chunk ──────────│                          │                        │
    │←── 转发 chunk ──────────│                          │                        │
    │                         │                          │                        │
    │                         │  ↑↑↑ 上游 A 断开 ↑↑↑     │                        │
    │                         │                          │                        │
    │                         │── 自动切换 → Supplier B ────────────────────────→│
    │                         │   (携带已产生的上下文)      │                        │
    │                         │                          │                        │
    │←── 转发 chunk ──────────│←── SSE (继续) ────────────│                        │
    │                         │                          │                        │
    │ 对用户侧: 连接不中断，仅在切换瞬间可能略有延迟                │                        │
```

#### 5.2.4 渠道健康检测 (PRD F-TKN-07, ROUTE-05)

| 检测项 | 频率 | 动作 |
|--------|------|------|
| 接口可用性 (HTTP 200) | 每 5 分钟/每 10 秒 | 失败则标记 unhealthy |
| 余额/库存 | 每 5 分钟 | 低于阈值触发预警 (F-TKN-04) |
| 响应延迟 P95 | 实时监测 | 延迟持续 > 2s 降权 |
| 错误率 | 实时监测 | 错误率 > 5% 自动切换 |
| **供应商 API 健康** (SUP-15) | 每 30 秒 | 供应商 API 不可用则自动下架 |

#### 5.2.5 转发逻辑详细设计 (PRD §6.2.7 ROUTE, 参考 one-api controller/relay.go)

```
请求处理流水线 (含失败重试):

  请求进入
     │
     ▼
  ┌─────────────┐
  │ 1. 限流检查  │──→ 超限 → HTTP 429 (ROUTE-13)
  └──────┬──────┘
         ▼
  ┌─────────────┐
  │ 2. 鉴权     │──→ 无效 → HTTP 401
  └──────┬──────┘
         ▼
  ┌─────────────┐
  │ 3. Token    │──→ 余额不足 → 422
  │  余额检查   │
  │  预扣配额   │──→ 估算预扣用户配额 (PreConsumedQuota)
  └──────┬──────┘
         ▼
  ┌──────────────────────┐
  │ 4. 路由决策 (缓存)     │──→ 查内存缓存 group+model → []Channel
  │    CacheGetRandom     │     按 priority 分组 → 随机选
  │    SatisfiedChannel   │     熔断器过滤
  └──────┬──────┘
         ▼
  ┌─────────────┐
  │ 5. 请求重写  │──→ 模型映射 + Header 改写 (ROUTE-09)
  └──────┬──────┘
         ▼
  ┌─────────────┐
  │ 6. 转发请求  │──→ 流式 / 非流式
  └──────┬──────┘
         │
         ├── 成功 ──────→ 7. 响应处理 (计费+日志)
         │                  ↓
         │               8. 异步后处理 (MQ→统计/风控)
         │
         └── 失败 ──────→ 重试循环 (参考 one-api Relay 函数):
                             for i = retryTimes; i > 0; i-- {
                                 // 跳过刚失败的渠道
                                 channel = CacheGetRandomSatisfiedChannel(
                                     group, model, ignoreFirstPriority=true)
                                 if channel.id == lastFailedId { continue }
                                 // 重新设置渠道上下文并重试
                                 SetupContextForSelectedChannel(c, channel, model)
                                 bizErr = relayHelper(c, relayMode)
                                 if bizErr == nil { return }
                             }
                             全部失败 → 返回错误给用户

重试条件 (参考 one-api shouldRetry):
  - 429 (Too Many Requests)      → 重试
  - 5xx (服务端错误)              → 重试
  - 网络超时/连接断开             → 重试
  - 400/401/403/2xx              → 不重试 (直接返回)

灰度路由 (ROUTE-14):
  - 支持按百分比分配流量到新供应商
  - 通过配置中心动态调整: canary.supplier_id = X, canary.weight = 10%

时段性调度 (ROUTE-08):
  - 非工作时段 (22:00-08:00): 切至低成本渠道
  - 通过 CRON 表达式配置调度策略

延迟敏感路由 (ROUTE-06):
  - 根据用户 IP 地域选择最近节点
  - 海外用户 → 海外节点 → 国内供应商 (通过海外→国内专线)
```

#### 5.2.6 供应商入驻流程 (PRD §6.2.6 SUP)

```
供应商入驻全流程:

  1. 提交申请 (SUP-01)
     - 填写企业信息 (公司名、联系人、邮箱)
     - 上传资质文件 (营业执照)
     - 填写 API 接入信息 (API 端点、认证方式)
     - 签署入驻协议 (SUP-16)

  2. 平台审核 (SUP-02)
     - 管理员审核资质
     - 配置佣金比例 (SUP-12)
     - 审核通过 → 开通供应商权限
     - 审核超时 48h 自动提醒

  3. 商品上架 (SUP-04)
     - 供应商创建 Token 商品
     - 设定模型类型、API 端点、计费单位
     - 自主定价 (SUP-05)，在平台限价范围内
     - 设置库存总量 (SUP-08)
     - 提交平台合规审核 (R-BIZ-49)

  4. 销售运营
     - 商品对用户展示
     - 用户下单 → 平台转发请求到供应商 API
     - 供应商 API 健康监控 (SUP-15)
     - 销售看板实时查看 (SUP-09)

  5. 结算提现 (SUP-10)
     - 平台 T+7 生成结算单
     - 供应商确认结算单
     - 申请提现 (银行转账/PayPal)
     - 平台审核并付款

供应商 API 适配器模式 (参考 one-api relay/adaptor/interface.go):

  ┌──────────────────────────────────────────────┐
  │          Adaptor (接口)                        │
  │                                                │
  │  + Init(meta)                                 │
  │  + GetRequestURL(meta) string                 │
  │  + SetupRequestHeader(c, req, meta)           │
  │  + ConvertRequest(c, mode, request) any       │
  │  + ConvertImageRequest(request) any           │
  │  + DoRequest(c, meta, body) Response          │
  │  + DoResponse(c, resp, meta) (Usage, Error)   │
  │  + GetModelList() []string                    │
  │  + GetChannelName() string                    │
  └──────────────────┬───────────────────────────┘
                     │
           ┌─────────┴──────────┐
           │   Adaptor 通用实现   │
           │                     │
           │ 1. Init: 渠道元数据  │
           │ 2. GetRequestURL:   │
           │    构建上游请求 URL  │
           │ 3. SetupRequestHeader│
           │    设置认证 Header  │
           │ 4. ConvertRequest:  │
           │    OpenAI → 供应商格式│
           │ 5. DoRequest: 发送   │
           │    HTTP 请求到上游   │
           │ 6. DoResponse:      │
           │    解析响应 + 计费   │
           │ 7. GetModelList:    │
           │    可用模型列表      │
           └────────────────────┘

渠道类型与 API 类型枚举 (参考 one-api channeltype + apitype):
  - channeltype: 区分"供应商平台"(OpenAI/Anthropic/Azure/AWS/阿里云等)
  - apitype:     区分"API 协议"(OpenAI 兼容/Anthropic Messages/Gemini/AWS 等)
  - 两者组合实现: Azure 使用 OpenAI 协议但不同 channeltype
```

