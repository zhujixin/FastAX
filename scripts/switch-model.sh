#!/bin/bash
# Claude Code 多模型切换脚本
# 用法: source scripts/switch-model.sh [模型名称]

CONFIG_FILE="$HOME/.claude/settings.json"

# 定义模型配置
declare -A MODELS=(
  # 小米 MIMO 模型
  ["mimo"]="mimo-v2.5-pro"
  ["mimo-pro"]="mimo-v2.5-pro"

  # DeepSeek 模型
  ["deepseek"]="deepseek-chat"
  ["deepseek-coder"]="deepseek-coder"
  ["deepseek-reasoner"]="deepseek-reasoner"

  # Claude 模型（需要原生 API）
  ["claude-sonnet"]="claude-sonnet-4-6"
  ["claude-opus"]="claude-opus-4-8"
  ["claude-haiku"]="claude-haiku-4-5-20251001"
)

# 定义 API 端点
declare -A ENDPOINTS=(
  ["mimo"]="https://api.xiaomimimo.com/anthropic"
  ["mimo-pro"]="https://api.xiaomimimo.com/anthropic"
  ["deepseek"]="https://api.deepseek.com/anthropic"
  ["deepseek-coder"]="https://api.deepseek.com/anthropic"
  ["deepseek-reasoner"]="https://api.deepseek.com/anthropic"
  ["claude-sonnet"]="https://api.anthropic.com"
  ["claude-opus"]="https://api.anthropic.com"
  ["claude-haiku"]="https://api.anthropic.com"
)

# 显示帮助
show_help() {
  echo "Claude Code 模型切换工具"
  echo ""
  echo "用法: source scripts/switch-model.sh [模型名称]"
  echo ""
  echo "可用模型:"
  echo "  mimo           - 小米 MIMO v2.5 Pro"
  echo "  mimo-pro       - 小米 MIMO v2.5 Pro"
  echo "  deepseek       - DeepSeek Chat"
  echo "  deepseek-coder - DeepSeek Coder"
  echo "  deepseek-reasoner - DeepSeek Reasoner"
  echo "  claude-sonnet  - Claude Sonnet 4.6"
  echo "  claude-opus    - Claude Opus 4.8"
  echo "  claude-haiku   - Claude Haiku 4.5"
  echo ""
  echo "示例:"
  echo "  source scripts/switch-model.sh deepseek"
  echo "  source scripts/switch-model.sh mimo"
}

# 切换模型
switch_model() {
  local model_name="$1"

  if [[ -z "$model_name" ]]; then
    show_help
    return 1
  fi

  if [[ -z "${MODELS[$model_name]}" ]]; then
    echo "❌ 未知模型: $model_name"
    show_help
    return 1
  fi

  local model="${MODELS[$model_name]}"
  local endpoint="${ENDPOINTS[$model_name]}"

  # 更新配置文件
  cat > "$CONFIG_FILE" << EOF
{
  "env": {
    "ANTHROPIC_BASE_URL": "$endpoint",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "$model",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "$model",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "$model",
    "ANTHROPIC_MODEL": "$model"
  },
  "permissions": {
    "defaultMode": "bypassPermissions"
  },
  "model": "sonnet",
  "enabledPlugins": {
    "gopls-lsp@claude-plugins-official": true
  },
  "skipDangerousModePermissionPrompt": true
}
EOF

  echo "✅ 已切换到: $model_name ($model)"
  echo "📡 API 端点: $endpoint"
  echo ""
  echo "请重新启动 Claude Code 会话以生效"
}

# 执行切换
switch_model "$1"
