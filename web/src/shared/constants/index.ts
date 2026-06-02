// Must match backend pagination size
export const ITEMS_PER_PAGE = 10;

// Channel type options (provider → { label, color })
export const CHANNEL_TYPES: Record<string, { label: string; color: string }> = {
  openai: { label: "OpenAI", color: "green" },
  anthropic: { label: "Anthropic", color: "orange" },
  gemini: { label: "Gemini", color: "blue" },
  deepseek: { label: "DeepSeek", color: "purple" },
  qwen: { label: "Qwen", color: "cyan" },
  glm: { label: "GLM", color: "geekblue" },
  moonshot: { label: "Moonshot", color: "magenta" },
  baichuan: { label: "Baichuan", color: "red" },
  zhipu: { label: "Zhipu", color: "volcano" },
  cohere: { label: "Cohere", color: "lime" },
  ollama: { label: "Ollama", color: "gold" },
  custom: { label: "自定义", color: "default" },
};

// Log type options
export const LOG_TYPES = [
  { value: 0, label: "全部" },
  { value: 1, label: "充值" },
  { value: 2, label: "消费" },
  { value: 3, label: "管理" },
  { value: 4, label: "系统" },
  { value: 5, label: "测试" },
];

// Order status
export const ORDER_STATUS: Record<string, { label: string; color: string }> = {
  pending: { label: "待支付", color: "default" },
  paid: { label: "已支付", color: "blue" },
  completed: { label: "已完成", color: "green" },
  cancelled: { label: "已取消", color: "default" },
  refunding: { label: "退款中", color: "orange" },
  refunded: { label: "已退款", color: "red" },
};

// Risk levels
export const RISK_LEVELS: Record<string, { label: string; color: string }> = {
  green: { label: "记录", color: "green" },
  yellow: { label: "预警", color: "gold" },
  orange: { label: "限权", color: "orange" },
  red: { label: "冻结", color: "red" },
};

// User roles
export const USER_ROLES: Record<string, { label: string; color: string }> = {
  user: { label: "用户", color: "blue" },
  admin: { label: "管理员", color: "red" },
  super_admin: { label: "超级管理员", color: "purple" },
  vendor: { label: "供应商", color: "green" },
};
