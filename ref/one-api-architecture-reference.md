# one-api 项目架构参考

> 来源: [songquanpeng/one-api](https://github.com/songquanpeng/one-api) (30.6K+ Stars, MIT License)
> 参考版本: v0.6.10
> 衍生项目: [QuantumNous/new-api](https://github.com/QuantumNous/new-api) (改进版, 19K+ Stars)
> 本文档基于网络搜索结果分析整理，2026-05-28

---

## 1. 项目定位

one-api 是一个 **LLM API 管理 & 分发系统**，核心能力：
- 统一 OpenAI 兼容接口，对接 30+ 模型供应商
- API Key 管理与二次分发
- 渠道负载均衡与故障转移
- 用户配额与计费管理
- 单可执行文件，Docker 一键部署

---

## 2. 目录结构

```
one-api/
├── common/                    # 共享工具函数
│   ├── config.go              # 配置管理
│   ├── env.go                 # 环境变量读取
│   ├── gin.go                 # Gin 工具函数
│   ├── random.go              # 随机数工具
│   ├── encrypt.go             # 加密工具
│   └── ...
├── constant/                  # 系统常量
│   ├── channel_type.go        # 渠道类型定义
│   ├── api_type.go            # API 类型定义
│   └── ...
├── model/                     # 数据模型 (GORM)
│   ├── channel.go             # 渠道模型 (Channel)
│   ├── ability.go             # 能力模型 (Ability) - 反范式设计
│   ├── user.go                # 用户模型
│   ├── token.go               # 令牌模型 (Token/API Key)
│   ├── log.go                 # 日志模型
│   ├── quota.go               # 配额相关
│   ├── options.go             # 系统配置
│   └── ...
├── controller/                # HTTP 请求处理器
│   ├── relay.go               # 中继控制器 (核心)
│   ├── channel.go             # 渠道管理
│   ├── token.go               # 令牌管理
│   ├── user.go                # 用户管理
│   ├── log.go                 # 日志查询
│   └── ...
├── router/                    # HTTP 路由注册
│   ├── router.go              # 主路由定义
│   └── web/                   # 前端静态文件
├── middleware/                # Gin 中间件
│   ├── auth.go                # 认证鉴权
│   ├── admin_auth.go          # 管理员鉴权
│   ├── rate_limit.go          # 限流
│   ├── cors.go                # CORS
│   ├── recovery.go            # 异常恢复
│   └── ...
├── relay/                     # 中继引擎 (核心模块)
│   ├── adaptor.go             # 适配器分发器 (GetAdaptor)
│   ├── adaptor/               # 各厂商适配器实现
│   │   ├── openai/            # OpenAI 适配器
│   │   ├── anthropic/         # Anthropic Claude 适配器
│   │   ├── gemini/            # Google Gemini 适配器
│   │   ├── aws/               # AWS Bedrock 适配器
│   │   ├── azure/             # Azure OpenAI 适配器
│   │   ├── ollama/            # Ollama 本地模型
│   │   ├── groq/              # Groq 适配器
│   │   ├── palm/              # Google PaLM
│   │   ├── proxy/             # 通用代理
│   │   ├── replicate/         # Replicate 适配器
│   │   ├── stepfun/           # 阶跃星辰
│   │   ├── tencent/           # 腾讯混元
│   │   ├── togetherai/        # Together AI
│   │   ├── vertexai/          # Vertex AI
│   │   ├── xai/               # xAI (Grok)
│   │   ├── xunfei/            # 讯飞星火
│   │   ├── zhipu/             # 智谱 GLM
│   │   ├── ai360/             # 360 智脑
│   │   ├── ali/               # 阿里通义千问
│   │   └── ...
│   ├── apitype/               # API 类型常量
│   ├── billing/               # 计费逻辑
│   ├── channeltype/           # 渠道类型
│   ├── meta/                  # 中继元数据
│   ├── relaymode/             # 中继模式
│   └── constant/              # 中继常量
├── service/                   # 业务逻辑层 (new-api 新增)
├── dto/                       # 数据传输对象 (new-api 新增)
├── i18n/                      # 国际化 (new-api 新增)
├── monitor/                   # 监控 (one-api 原有)
├── bin/                       # 构建脚本
├── docs/                      # 文档
├── web/                       # React 前端
├── main.go                    # 入口文件
├── go.mod / go.sum            # Go 模块
└── Dockerfile                 # Docker 构建
```

---

## 3. 核心架构分层

```
请求进入
    │
    ▼
┌──────────────┐
│  router/      │  Gin 路由分组
│  router.go    │  · /api/* → 管理后台
│               │  · /v1/*  → OpenAI 兼容中继
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ middleware/   │  中间件管道
│  auth.go     │  · Token 鉴权
│  rate_limit  │  · 限流
│  cors.go     │  · CORS
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ relay/        │  中继引擎 (核心)
│  adaptor.go  │  · 适配器分发
│  relay.go    │  · 请求转发
│  billing/    │  · 计费
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ relay/adaptor/│  供应商适配器
│  openai/     │  · OpenAI 格式
│  anthropic/  │  · Claude 格式
│  gemini/     │  · Gemini 格式
│  ...         │  · 其他供应商
└──────────────┘
```

---

## 4. 核心数据模型

### 4.1 Channel (渠道)

```go
// model/channel.go
type Channel struct {
    Id               int    `json:"id" gorm:"primaryKey;autoIncrement"`
    Name             string `json:"name" gorm:"unique"`
    Type             int    `json:"type" gorm:"default:0"`     // 渠道类型 (OpenAI/Anthropic/...)
    Key              string `json:"key" gorm:"type:text"`      // API Key (加密存储)
    BaseURL          string `json:"base_url" gorm:"type:text"` // 自定义 Base URL
    Models           string `json:"models" gorm:"type:text"`   // 支持的模型列表 (逗号分隔)
    ModelMapping     string `json:"model_mapping" gorm:"type:text"` // 模型映射 JSON
    Priority         int    `json:"priority" gorm:"default:0"` // 优先级
    Weight           int    `json:"weight" gorm:"default:0"`   // 权重 (负载均衡)
    Status           int    `json:"status" gorm:"default:1"`   // 0=禁用, 1=启用
    Group            string `json:"group" gorm:"type:text;default:''"` // 用户组限制
    CreatedTime      int64  `json:"created_time" gorm:"autoCreateTime"`
}
```

### 4.2 Ability (能力 - 反范式设计)

```go
// model/ability.go
// 每条记录 = 用户分组 + 模型 + 渠道 的组合
// 用于快速查询匹配的渠道
type Ability struct {
    Group     string `json:"group" gorm:"type:varchar(64);uniqueIndex:idx_group_model"`
    Model     string `json:"model" gorm:"type:varchar(255);uniqueIndex:idx_group_model"`
    ChannelId int    `json:"channel_id" gorm:"uniqueIndex:idx_group_model"`
    Enabled   bool   `json:"enabled"`
    Priority  int    `json:"priority"`
    Weight    int    `json:"weight"`
}
```

### 4.3 Token (用户令牌/API Key)

```go
// model/token.go
type Token struct {
    Id         int    `json:"id" gorm:"primaryKey;autoIncrement"`
    Key        string `json:"key" gorm:"uniqueIndex;type:char(48)"` // 令牌值
    Name       string `json:"name" gorm:"type:varchar(255)"`        // 令牌名称
    UserId     int    `json:"user_id" gorm:"index"`                 // 所属用户
    Models     string `json:"models"`                               // 可调用的模型列表
    UnlimitedQuota bool `json:"unlimited_quota"`                    // 是否无限配额
    Quota      float64 `json:"quota" gorm:"default:0"`              // 配额额度
    UsedQuota  float64 `json:"used_quota" gorm:"default:0"`         // 已用配额
    Status     int    `json:"status" gorm:"default:1"`              // 0=禁用, 1=启用
    ExpiredAt  int64  `json:"expired_at"`                           // 过期时间
    CreatedTime int64 `json:"created_time" gorm:"autoCreateTime"`
}
```

### 4.4 Log (调用日志)

```go
// model/log.go
type Log struct {
    Id            int     `json:"id" gorm:"primaryKey;autoIncrement"`
    UserId        int     `json:"user_id" gorm:"index"`
    TokenId       int     `json:"token_id"`      // 使用的令牌 ID
    TokenName     string  `json:"token_name"`    // 令牌名称 (冗余)
    Model         string  `json:"model"`         // 请求模型
    PromptTokens  int     `json:"prompt_tokens"` // 提示 Token 数
    CompletionTokens int  `json:"completion_tokens"` // 补全 Token 数
    Quota         float64 `json:"quota"`          // 消耗配额
    ChannelId     int     `json:"channel_id"`     // 使用的渠道
    ChannelName   string  `json:"channel_name"`   // 渠道名称 (冗余)
    IP            string  `json:"ip"`
    RequestTime   int64   `json:"request_time"`
    CreatedTime   int64   `json:"created_time" gorm:"autoCreateTime"`
}
```

---

## 5. 适配器模式 (Adapter Pattern)

### 5.1 适配器接口

```go
// relay/adaptor.go (核心接口定义)
type Adaptor interface {
    // 初始化适配器
    Init(meta *Meta)
    
    // 获取请求体处理函数
    GetRequestURL(meta *Meta) (string, error)
    SetupRequestHeader(c *gin.Context, req *http.Request, meta *Meta) error
    ConvertRequest(c *gin.Context, relayMode int, request *Request) error
    
    // 处理响应
    DoResponse(c *gin.Context, resp *http.Response, meta *Meta) (usage *Usage, err error)
    
    // 模型列表
    GetModelList() []string
    
    // 渠道类型
    GetChannelType() int
}

// 适配器分发器
func GetAdaptor(apiType int) Adaptor {
    switch apiType {
    case apitype.Anthropic:
        return &anthropic.Adaptor{}
    case apitype.OpenAI:
        return &openai.Adaptor{}
    case apitype.Gemini:
        return &gemini.Adaptor{}
    case apitype.AWS:
        return &aws.Adaptor{}
    // ... 其他供应商
    }
    return nil
}
```

### 5.2 请求处理流程 (`relay/relay.go`)

```go
// 简化的中继处理流程
func Relay(c *gin.Context) {
    // 1. 获取用户和令牌信息
    userId := c.GetInt("id")
    tokenId := c.GetInt("token_id")
    
    // 2. 从 Ability 表选择匹配渠道
    channel := CacheGetRandomSatisfiedChannel(userGroup, modelName)
    
    // 3. 获取对应适配器
    adaptor := GetAdaptor(channel.Type)
    
    // 4. 预扣配额 (估算)
    preConsumeQuota(tokenId, estimatedTokens)
    
    // 5. 适配器转换请求并转发
    err := adaptor.ConvertRequest(c, relayMode, request)
    
    // 6. 处理响应 (流式/非流式)
    usage, err := adaptor.DoResponse(c, resp, meta)
    
    // 7. 后扣配额 (按实际用量校正)
    postConsumeQuota(tokenId, preConsumed, actualQuota)
    
    // 8. 记录调用日志
    recordLog(userId, tokenId, channel, model, usage)
}
```

### 5.3 渠道选择算法

```go
// 渠道选择流程
func CacheGetRandomSatisfiedChannel(group, model string) *Channel {
    // 1. 构建查询条件
    query := getChannelQuery(group, model)
    
    // 2. 从缓存或 DB 获取匹配的 Ability 列表
    abilities := getCachedAbilities(query)
    
    // 3. 按优先级分组
    groups := groupByPriority(abilities)
    
    // 4. 从最高优先级组中按权重随机选择
    for _, group := range groups {
        selected := weightedRandomSelect(group)
        if selected != nil && isChannelHealthy(selected) {
            return selected
        }
    }
    
    // 5. 所有渠道都不可用 ⇒ 返回错误
    return nil
}
```

---

## 6. 配额计费系统

### 6.1 配额计算公式

```
额度 = 分组倍率 × 模型倍率 × (PromptToken数 + CompletionToken数 × 补全倍率)
```

- **分组倍率**: 不同用户组可设置不同倍率 (如 VIP 组 0.8x)
- **模型倍率**: 每种模型有固定倍率 (GPT-4 = 30x, GPT-3.5 = 1x)
- **补全倍率**: Completion Token 额外系数 (GPT-3.5 = 1.33x)

### 6.2 两阶段配额

| 阶段 | 时机 | 动作 |
|------|------|------|
| **预消费** | 请求转发前 | 根据输入 token 估算预扣额度 |
| **后消费** | 响应完成后 | 按实际用量计算差值，多退少补 |

### 6.3 缓存优化策略

| 策略 | 说明 |
|------|------|
| **Redis 缓存** | 用户配额、分组信息缓存在 Redis |
| **内存缓存** | `MEMORY_CACHE_ENABLED=true` 时启用 |
| **批量更新** | `BATCH_UPDATE_ENABLED=true` 时小额更新先缓存再批量刷入 DB |

---

## 7. 路由注册

```go
// router/router.go
func SetRouter(router *gin.Engine) {
    // 管理后台 API (需管理员 Token)
    adminGroup := router.Group("/api")
    adminGroup.Use(middleware.AdminAuth())
    {
        // 渠道管理
        adminGroup.GET("/channels", controller.GetChannels)
        adminGroup.POST("/channel", controller.AddChannel)
        adminGroup.PUT("/channel/:id", controller.UpdateChannel)
        adminGroup.DELETE("/channel/:id", controller.DeleteChannel)
        adminGroup.POST("/channel/:id/test", controller.TestChannel)
        
        // 令牌管理
        adminGroup.GET("/tokens", controller.GetTokens)
        adminGroup.POST("/token", controller.AddToken)
        adminGroup.PUT("/token/:id", controller.UpdateToken)
        adminGroup.DELETE("/token/:id", controller.DeleteToken)
        
        // 用户管理
        adminGroup.GET("/users", controller.GetUsers)
        adminGroup.POST("/user", controller.AddUser)
        adminGroup.PUT("/user/:id", controller.UpdateUser)
        adminGroup.DELETE("/user/:id", controller.DeleteUser)
        
        // 系统设置
        adminGroup.GET("/options", controller.GetOptions)
        adminGroup.PUT("/options", controller.UpdateOptions)
        
        // 日志
        adminGroup.GET("/logs", controller.GetLogs)
    }
    
    // 中继 API (OpenAI 兼容)
    relayGroup := router.Group("/v1")
    relayGroup.Use(middleware.TokenAuth())
    {
        relayGroup.Any("/chat/completions", controller.Relay)
        relayGroup.Any("/completions", controller.Relay)
        relayGroup.Any("/embeddings", controller.Relay)
        relayGroup.Any("/moderations", controller.Relay)
        relayGroup.Any("/images/generations", controller.Relay)
        relayGroup.Any("/audio/transcriptions", controller.Relay)
        relayGroup.Any("/audio/speech", controller.Relay)
        relayGroup.Any("/models", controller.Relay)
        relayGroup.Any("/models/:model", controller.Relay)
    }
}
```

---

## 8. 关键设计亮点

### 8.1 值得借鉴的设计

| 设计 | 说明 | 对 FastAX 的参考价值 |
|------|------|---------------------|
| **Ability 反范式模型** | group+model+channel 组合索引，快速查询匹配渠道 | FastAX 路由引擎可参考此设计实现快速渠道筛选 |
| **Adapter 模式** | 每个供应商独立适配器，新增只需实现接口 | FastAX proxy-service 的 VendorAdapter 设计一致 |
| **两阶段配额** | 预扣 → 实扣，防止超用 | FastAX Token 计费可参考实现 |
| **批量更新** | 小额写入先缓存再批量刷 DB | FastAX call_log 高写入场景可借鉴 |
| **主从节点** | Master 管理后台 + Slave 纯转发，横向扩展 | FastAX proxy-service 独立扩缩容 |
| **Key 加密存储** | 数据库中 API Key 加密存储 | FastAX 供应商 Key 存储 |
| **Redis 缓存链** | 渠道信息、配额、限流全部可 Redis 缓存 | FastAX Redis 缓存设计参考 |

### 8.2 one-api vs FastAX 差异

| 维度 | one-api | FastAX |
|------|---------|--------|
| 定位 | API Key 管理与分发 | Token 代理与交易平台 |
| 供应商 | 聚合已有 API | 自有 + 第三方入驻 |
| 商业模式 | 开源免费 | 商业平台 + 开放市场 |
| 多协议 | OpenAI 兼容为主 | OpenAI + Anthropic + Gemini 原生 |
| 多语言 | 有 i18n | 完整 i18n + 多语言通知 |
| 部署 | 单机关怀 | 微服务 + K8s |
| 数据库 | SQLite/MySQL/PostgreSQL | SQLite + MongoDB + ES |
| 前端 | React + Semi-UI | React + Ant Design 5 |

---

## 9. 参考资源

| 资源 | 地址 |
|------|------|
| one-api GitHub | https://github.com/songquanpeng/one-api |
| new-api GitHub | https://github.com/QuantumNous/new-api |
| new-api DeepWiki | https://deepwiki.com/QuantumNous/new-api/1.1-system-architecture |
| 最新 Release | https://github.com/songquanpeng/one-api/releases |
