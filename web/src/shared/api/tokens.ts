import api from "@/shared/utils/axios";
import type { ApiResponse, PaginatedData, TokenProduct, UserToken, BYOKKey } from "./types";

export const tokenService = {
  getProducts: () =>
    api.get<ApiResponse<TokenProduct[]>>("/tokens/products"),

  getProduct: (id: number) =>
    api.get<ApiResponse<TokenProduct>>(`/tokens/products/${id}`),

  getMyTokens: () =>
    api.get<ApiResponse<UserToken[]>>("/tokens/my"),

  getUsageHistory: (params?: { page?: number; size?: number }) =>
    api.get<ApiResponse<PaginatedData<Record<string, unknown>>>>("/tokens/my/usage", { params }),

  buy: (data: { product_id: number; quantity: number; payment_method: string }) =>
    api.post<ApiResponse<{ order_no: string; amount: number; payment_url?: string }>>("/tokens/buy", data),

  transfer: (data: { token_id: number; target_user_id: number; amount: number }) =>
    api.post<ApiResponse<null>>("/tokens/transfer", data),

  extract: (data: { token_id: number }) =>
    api.post<ApiResponse<{ api_key: string }>>("/tokens/extract", data),

  // BYOK (Bring Your Own Key)
  listBYOKKeys: () =>
    api.get<ApiResponse<BYOKKey[]>>("/byok/keys"),

  addBYOKKey: (data: { provider: string; key: string }) =>
    api.post<ApiResponse<BYOKKey>>("/byok/keys", data),

  deleteBYOKKey: (id: number) =>
    api.delete<ApiResponse<null>>(`/byok/keys/${id}`),

  setBYOKKeyStatus: (id: number, enabled: boolean) =>
    api.put<ApiResponse<null>>(`/byok/keys/${id}/status`, { enabled }),
};
