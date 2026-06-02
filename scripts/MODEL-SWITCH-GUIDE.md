# Claude Code 多模型切换指南

## 快速开始

### 1. 添加快捷命令到 Shell 配置

将以下内容添加到你的 `~/.bashrc` 或 `~/.zshrc`：

```bash
# Claude Code 模型切换快捷命令
alias ccm='source ~/path/to/fastax-server/scripts/switch-model.sh'
alias ccc='bash ~/path/to/fastax-server/scripts/current-model.sh'

# 快速切换函数
ccs() {
  source ~/path/to/fastax-server/scripts/switch-model.sh "$1"
}
```

然后重新加载配置：
```bash
source ~/.bashrc  # 或 source ~/.zshrc
```

### 2. 使用方法

```bash
# 查看当前模型
ccc

# 切换到 DeepSeek
ccs deepseek

# 切换到小米 MIMO
ccs mimo

# 切换到 Claude Sonnet
ccs claude-sonnet
```

## 支持的模型

| 命令 | 模型 | API 端点 | 说明 |
|------|------|----------|------|
| `ccs mimo` | mimo-v2.5-pro | api.xiaomimimo.com | 小米 MIMO 代理 |
| `ccs deepseek` | deepseek-chat | api.deepseek.com | DeepSeek Chat |
| `ccs deepseek-coder` | deepseek-coder | api.deepseek.com | DeepSeek Coder |
| `ccs deepseek-reasoner` | deepseek-reasoner | api.deepseek.com | DeepSeek Reasoner |
| `ccs claude-sonnet` | claude-sonnet-4-6 | api.anthropic.com | Claude Sonnet 4.6 |
| `ccs claude-opus` | claude-opus-4-8 | api.anthropic.com | Claude Opus 4.8 |
| `ccs claude-haiku` | claude-haiku-4-5 | api.anthropic.com | Claude Haiku 4.5 |

## 前置要求

### DeepSeek 模型

1. 注册 DeepSeek 账号: https://platform.deepseek.com
2. 获取 API Key
3. 修改脚本中的 `ANTHROPIC_AUTH_TOKEN`

### Claude 原生模型

1. 注册 Anthropic 账号: https://console.anthropic.com
2. 获取 API Key
3. 修改脚本中的 `ANTHROPIC_AUTH_TOKEN`

## 高级配置

### 自定义 API Key

编辑 `scripts/switch-model.sh`，在对应模型的配置中添加：

```bash
# 在 switch_model() 函数中修改
cat > "$CONFIG_FILE" << EOF
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-api-key-here",
    "ANTHROPIC_BASE_URL": "$endpoint",
    ...
  }
}
EOF
```

### 环境变量方式（临时切换）

```bash
# 临时使用 DeepSeek
ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic \
ANTHROPIC_MODEL=deepseek-chat \
claude

# 临时使用 Claude 原生
ANTHROPIC_BASE_URL=https://api.anthropic.com \
ANTHROPIC_MODEL=claude-sonnet-4-6 \
claude
```

### 使用 --model 参数

```bash
# 启动时指定模型
claude --model sonnet
claude --model opus
claude --model haiku

# 使用完整模型名
claude --model claude-opus-4-8
```

## 故障排除

### 问题：切换后模型未生效

**解决方案**：
1. 退出当前 Claude Code 会话
2. 重新启动 `claude`

### 问题：API 认证失败

**解决方案**：
1. 检查 API Key 是否正确
2. 确认 API 端点是否支持该模型
3. 检查账户余额/配额

### 问题：脚本权限错误

**解决方案**：
```bash
chmod +x scripts/switch-model.sh
chmod +x scripts/current-model.sh
```

## 工作原理

Claude Code 的模型配置存储在 `~/.claude/settings.json` 文件中：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "API 端点",
    "ANTHROPIC_MODEL": "模型名称",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "Haiku 槽位模型",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "Opus 槽位模型",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "Sonnet 槽位模型"
  },
  "model": "当前槽位 (sonnet/opus/haiku)"
}
```

切换脚本通过修改这个文件来实现模型切换，需要重启 Claude Code 会话才能生效。
