# FastAX

<p align="center">
  <strong>Token 代理与交易平台</strong>
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> •
  <a href="#特性">特性</a> •
  <a href="#技术栈">技术栈</a> •
  <a href="#api-文档">API 文档</a> •
  <a href="#部署">部署</a>
</p>

---

## 项目简介

FastAX 是一个高性能的 **Token 代理与交易平台**，采用 Go 单体架构设计，作为连接终端用户与 Token 源头（OpenAI、Claude、Gemini、DeepSeek、Qwen、GLM 等）的中间枢纽。

### 核心价值

- **统一接入**：一个 API 访问多家 Token 供应商
- **智能路由**：自动选择最优供应商，支持故障转移
- **成本优化**：智能计费、用量监控、成本分析
- **企业就绪**：多租户、权限管理、审计日志

---

## 特性

### 核心功能（S0-S2）

- ✅ **用户认证**：JWT 认证、OAuth 2.0、SSO 单点登录
- ✅ **Token 管理**：产品管理、API Key 管理、用量追踪
- ✅ **订单支付**：订单生命周期、多支付网关（微信/支付宝/Stripe）
- ✅ **代理转发**：多协议支持（OpenAI/Anthropic/Gemini）、流式输出

### 增值功能（S3-S4）

- ✅ **供应商管理**：供应商入驻、渠道管理、健康监控
- ✅ **风险控制**：异常检测、限流策略、黑名单管理
- ✅ **通知系统**：多渠道通知（邮件/短信/Webhook）
- ✅ **统计分析**：用量统计、成本分析、收益报表

### 高级功能（S5-S6）

- ✅ **安全护栏**：内容过滤、合规检查、敏感词检测
- ✅ **BYOK 支持**：Bring Your Own Key，用户自带 API Key
- ✅ **成本优化**：智能路由、批量折扣、用量预测
- ✅ **企业功能**：多团队管理、权限控制、审计日志
- ✅ **模型市场**：模型发现、对比评测、推荐系统
- ✅ **插件系统**：自定义扩展、第三方集成

---

## 技术栈

### 后端

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.22+ | 主语言 |
| Gin | 1.10 | Web 框架 |
| GORM | 1.25 | ORM |
| SQLite | - | 主数据库（WAL 模式） |
| Redis | 6+ | 缓存（可选） |
| RabbitMQ | 3.x | 消息队列 |

### 前端

| 技术 | 版本 | 用途 |
|------|------|------|
| React | 18 | UI 框架 |
| TypeScript | 5.x | 类型安全 |
| Vite | 5.x | 构建工具 |
| Ant Design | 5.x | 组件库 |
| TailwindCSS | 3.x | 样式框架 |

### 基础设施

| 技术 | 用途 |
|------|------|
| Docker | 容器化 |
| Kubernetes | 编排 |
| Nginx | 反向代理 |
| Prometheus + Grafana | 监控 |
| ELK Stack | 日志 |

---

## 快速开始

### 前置要求

- Go 1.22+
- Node.js 18+（前端开发）
- Redis（可选，用于缓存）

### 安装

```bash
# 克隆仓库
git clone https://github.com/fastax/fastax-server.git
cd fastax-server

# 安装依赖
go mod tidy

# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置
vim config.yaml
```

### 配置

编辑 `config.yaml`：

```yaml
server:
  port: 8080
  mode: debug  # debug | release | test

database:
  path: data/fastax.db
  wal_mode: true

redis:
  addr: localhost:6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-here"
  access_expiry: 24h
  refresh_expiry: 168h

rate_limit:
  ip: 60
  user_default: 60
  user_enterprise: 300
```

### 运行

```bash
# 开发模式
go run ./cmd/fastax

# 或者编译后运行
go build -o bin/fastax ./cmd/fastax
./bin/fastax

# 数据库迁移
go run ./cmd/fastax -migrate
```

服务启动后访问：`http://localhost:8080`

### 前端开发

```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build
```

---

## 项目结构

```
fastax-server/
├── cmd/
│   └── fastax/
│       └── main.go              # 应用入口
├── internal/
│   ├── shared/                  # 共享层
│   │   ├── config/              # 配置管理
│   │   ├── model/               # 数据模型
│   │   ├── cache/               # 缓存（Redis）
│   │   ├── middleware/          # 中间件
│   │   ├── response/            # 响应封装
│   │   └── i18n/                # 国际化
│   ├── domain/                  # 业务域
│   │   ├── user/                # 用户管理
│   │   ├── token/               # Token 管理
│   │   ├── order/               # 订单管理
│   │   ├── payment/             # 支付管理
│   │   ├── proxy/               # 代理转发（核心）
│   │   │   ├── relay/           # 路由引擎
│   │   │   ├── adaptor/         # 适配器
│   │   │   └── monitor/         # 健康监控
│   │   ├── vendor/              # 供应商管理
│   │   ├── risk/                # 风险控制
│   │   ├── notify/              # 通知系统
│   │   ├── stats/               # 统计分析
│   │   ├── commission/          # 佣金管理
│   │   ├── log/                 # 日志审计
│   │   ├── guardrail/           # 安全护栏
│   │   ├── byok/                # BYOK 支持
│   │   ├── cost/                # 成本优化
│   │   ├── enterprise/          # 企业功能
│   │   ├── market/              # 模型市场
│   │   └── plugin/              # 插件系统
│   └── router/                  # 路由注册
├── web/                         # 前端项目
├── doc/                         # 项目文档
├── ref/                         # 参考代码
├── config.example.yaml          # 配置示例
└── README.md
```

---

## API 文档

### 认证

所有需要认证的 API 需要在请求头中携带 JWT Token：

```
Authorization: Bearer <your-jwt-token>
```

### 主要端点

| 模块 | 端点 | 说明 |
|------|------|------|
| **认证** | `POST /api/v1/auth/login` | 用户登录 |
| | `POST /api/v1/auth/register` | 用户注册 |
| | `POST /api/v1/auth/refresh` | 刷新 Token |
| **用户** | `GET /api/v1/user/profile` | 获取用户信息 |
| | `PUT /api/v1/user/profile` | 更新用户信息 |
| **Token** | `GET /api/v1/tokens` | 获取 Token 列表 |
| | `POST /api/v1/tokens` | 创建 Token |
| **订单** | `GET /api/v1/orders` | 获取订单列表 |
| | `POST /api/v1/orders` | 创建订单 |
| **代理** | `POST /v1/chat/completions` | OpenAI 兼容接口 |
| | `POST /v1/messages` | Anthropic 兼容接口 |
| **管理** | `GET /api/v1/admin/dashboard` | 管理后台 |

完整 API 文档请参考：[doc/FastAX-PDD/03-api.md](doc/FastAX-PDD/03-api.md)

---

## 开发指南

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./internal/domain/user/...
go test ./internal/domain/proxy/...

# 带覆盖率
go test -cover ./...
```

### 代码规范

- 遵循 Go 标准项目布局
- 使用 `gofmt` 和 `golangci-lint` 格式化代码
- 编写单元测试覆盖核心逻辑
- 提交前确保所有测试通过

### 提交规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型（type）：
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

---

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t fastax-server .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  fastax-server
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fastax-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fastax-server
  template:
    metadata:
      labels:
        app: fastax-server
    spec:
      containers:
      - name: fastax-server
        image: fastax-server:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        - name: data
          mountPath: /app/data
      volumes:
      - name: config
        configMap:
          name: fastax-config
      - name: data
        persistentVolumeClaim:
          claimName: fastax-data
```

---

## 性能指标

- **并发支持**：单机 1000+ QPS
- **响应延迟**：< 100ms（P99）
- **可用性**：99.9%（配合健康检查和故障转移）

---

## 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 提交更改：`git commit -m 'feat: add your feature'`
4. 推送分支：`git push origin feature/your-feature`
5. 提交 Pull Request

### 开发环境

```bash
# 安装开发工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行 linter
golangci-lint run

# 运行测试
go test -race ./...
```

---

## 许可证

本项目采用 [MIT License](LICENSE) 开源许可证。

---

## 联系方式

- **问题反馈**：[GitHub Issues](https://github.com/fastax/fastax-server/issues)
- **功能建议**：[GitHub Discussions](https://github.com/fastax/fastax-server/discussions)
- **安全漏洞**：请通过邮件私下报告

---

## 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM
- [go-redis](https://github.com/redis/go-redis) - Redis 客户端
- [one-api](https://github.com/songquanpeng/one-api) - 参考架构
