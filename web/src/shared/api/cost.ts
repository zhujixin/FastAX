import api from "@/shared/utils/axios";
import type { ApiResponse } from "./types";

export interface BudgetInfo {
  user_id: number;
  period: string;
  limit: number;
  spent: number;
  exceeded: boolean;
  usage_pct: number;
  period_start: string;
  period_end: string;
}

export interface AlertInfo {
  user_id: number;
  thresholds: number[];
}

export interface CacheStats {
  total_entries: number;
  active_entries: number;
  total_hits: number;
  hit_rate: number;
  estimate_savings: string;
}

export const costService = {
  // User budget
  getBudget: () =>
    api.get<ApiResponse<BudgetInfo>>("/user/budget"),

  setBudget: (period: string, limit: number) =>
    api.put<ApiResponse<BudgetInfo>>("/user/budget", { period, limit }),

  // User cost alerts
  getAlerts: () =>
    api.get<ApiResponse<AlertInfo>>("/user/cost-alerts"),

  setAlert: (thresholds: number[]) =>
    api.put<ApiResponse<AlertInfo>>("/user/cost-alerts", { thresholds }),

  // Admin cache management
  getCacheStats: () =>
    api.get<ApiResponse<CacheStats>>("/admin/cache/stats"),

  updateCacheConfig: (config: { enabled: boolean; similarity_threshold: number; ttl_seconds: number; max_entries: number }) =>
    api.put<ApiResponse<Record<string, unknown>>>("/admin/cache/config", config),
};
