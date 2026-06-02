import api from "@/shared/utils/axios";
import type { ApiResponse } from "./types";

export const enterpriseService = {
  listSubAccounts: () =>
    api.get<ApiResponse<Array<{ id: number; name: string; quota: number; used: number; status: string }>>>("/enterprise/sub-accounts"),

  createSubAccount: (data: { name: string; quota: number }) =>
    api.post<ApiResponse<Record<string, unknown>>>("/enterprise/sub-accounts", data),

  setSubAccountStatus: (id: number, status: string) =>
    api.put<ApiResponse<null>>(`/enterprise/sub-accounts/${id}/status`, { status }),

  updateQuota: (id: number, quota: number) =>
    api.put<ApiResponse<null>>(`/enterprise/sub-accounts/${id}/quota`, { quota }),

  getUsage: () =>
    api.get<ApiResponse<Record<string, unknown>>>("/enterprise/usage"),
};
