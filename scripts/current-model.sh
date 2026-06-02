#!/bin/bash
# 查看当前 Claude Code 模型配置

CONFIG_FILE="$HOME/.claude/settings.json"

if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "❌ 配置文件不存在: $CONFIG_FILE"
  exit 1
fi

echo "📋 当前 Claude Code 模型配置:"
echo ""

# 提取配置信息
BASE_URL=$(grep -o '"ANTHROPIC_BASE_URL": *"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
MODEL=$(grep -o '"ANTHROPIC_MODEL": *"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
SLOT=$(grep -o '"model": *"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)

echo "  模型名称: $MODEL"
echo "  API 端点: $BASE_URL"
echo "  模型槽位: $SLOT"
echo ""

# 判断模型类型
if [[ "$BASE_URL" == *"xiaomimimo"* ]]; then
  echo "  类型: 小米 MIMO 代理"
elif [[ "$BASE_URL" == *"deepseek"* ]]; then
  echo "  类型: DeepSeek 代理"
elif [[ "$BASE_URL" == *"anthropic"* ]]; then
  echo "  类型: Anthropic 原生 API"
else
  echo "  类型: 自定义代理"
fi
