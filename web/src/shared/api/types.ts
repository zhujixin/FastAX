// Unified API response wrapper (matches backend shared/response)
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
  trace_id?: string;
}

export interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}

// Auth
export interface LoginRequest {
  account: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: UserInfo;
}

export interface RegisterRequest {
  username: string;
  email?: string;
  phone?: string;
  password: string;
}

export interface UserInfo {
  id: number;
  username: string;
  role: "user" | "admin" | "super_admin" | "vendor";
  level: string;
  preferred_language: string;
}

// Token Products
export interface TokenProduct {
  id: number;
  name: string;
  supplier: string;
  price: number;
  stock: number;
  model?: string;
  status: string;
}

export interface UserToken {
  id: number;
  product_name: string;
  total: number;
  used: number;
  remaining: number;
  expires_at: string;
  status: string;
}

// Orders
export interface Order {
  id: string;
  product_name: string;
  amount: number;
  status: string;
  created_at: string;
}

export interface OrderDetail extends Order {
  quantity: number;
  payment_method: string;
  user_id: number;
  updated_at: string;
}

// Vendor
export interface VendorInfo {
  id: number;
  company_name: string;
  contact_name: string;
  contact_email: string;
  status: string;
  products_count: number;
  created_at: string;
}

export interface VendorProduct {
  id: number;
  name: string;
  price: number;
  status: string;
  sales: number;
}

// Risk
export interface RiskEvent {
  id: number;
  type: string;
  level: string;
  user_identifier: string;
  handled: boolean;
  created_at: string;
}

export interface RiskRule {
  id: number;
  name: string;
  condition: string;
  action: string;
  enabled: boolean;
}

// Stats
export interface DashboardSummary {
  total_users: number;
  today_new_users: number;
  total_orders: number;
  today_new_orders: number;
  total_revenue: string;
  today_revenue: string;
  active_tokens: number;
}

export interface ChartDataPoint {
  date: string;
  value?: number;
  amount?: string;
}

export interface DashboardCharts {
  revenue: ChartDataPoint[];
  users: ChartDataPoint[];
  orders: ChartDataPoint[];
  tokens: ChartDataPoint[];
}

export interface UsageStats {
  total_tokens: number;
  used_tokens: number;
  remaining: number;
  daily_usage: Array<{ date: string; tokens: number }>;
}

// Notifications
export interface Notification {
  id: number;
  title: string;
  content: string;
  read: boolean;
  created_at: string;
}

// i18n
export interface LanguageInfo {
  locale: string;
  name: string;
  is_default: boolean;
  enabled?: boolean;
}

// Model Market
export interface ModelVariant {
  id: number;
  name: string;
  provider: string;
  price_per_1k: number;
  context_window: number;
  latency_ms: number;
  capabilities: string[];
}

export interface ProviderHealth {
  id: number;
  provider: string;
  availability: number;
  avg_latency: number;
  p95_latency: number;
  error_rate: number;
  status: "healthy" | "degraded" | "down";
}

// Guardrail
export interface GuardrailRule {
  id: number;
  name: string;
  stage: "before" | "after";
  type: "pii" | "injection" | "secret" | "content";
  action: "enforce" | "monitor" | "log";
  conditions: string;
  priority: number;
  enabled: number;
}

export interface GuardrailLog {
  id: number;
  trace_id: string;
  user_id: number;
  rule_id: number;
  stage: string;
  detected_entities: string;
  action_taken: string;
  created_at: number;
}

// BYOK
export interface BYOKKey {
  id: number;
  user_id: number;
  provider: string;
  key_prefix: string;
  status: number;
  total_used: number;
  created_at: number;
}
