import api from "@/shared/utils/axios";
import type { ApiResponse, UsageStats } from "./types";

export const statsService = {
  getUsage: (period?: string) =>
    api.get<ApiResponse<UsageStats>>("/stats/usage", { params: { period } }),

  getConsumption: (period?: string) =>
    api.get<ApiResponse<Record<string, unknown>>>("/stats/consumption", { params: { period } }),

  getBills: (params?: { page?: number; page_size?: number }) =>
    api.get<ApiResponse<{ items: Record<string, unknown>[]; total: number }>>("/stats/bills", { params }),

  getSummary: () =>
    api.get<ApiResponse<{
      total_tokens: number;
      total_amount: string;
      month_tokens: number;
      month_amount: string;
      total_requests: number;
      total_orders: number;
    }>>("/stats/summary"),
};
