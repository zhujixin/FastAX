import api from "@/shared/utils/axios";
import type { ApiResponse, PaginatedData, Order, OrderDetail } from "./types";

export const orderService = {
  create: (data: { product_id: number; quantity: number; payment_method: string }) =>
    api.post<ApiResponse<{ order_no: string }>>("/orders", data),

  list: (params?: { page?: number; size?: number }) =>
    api.get<ApiResponse<PaginatedData<Order>>>("/orders", { params }),

  get: (id: string) =>
    api.get<ApiResponse<OrderDetail>>(`/orders/${id}`),

  cancel: (id: string) =>
    api.post<ApiResponse<null>>(`/orders/${id}/cancel`),

  requestRefund: (id: string, reason?: string) =>
    api.post<ApiResponse<null>>(`/orders/${id}/refund`, { reason }),
};
