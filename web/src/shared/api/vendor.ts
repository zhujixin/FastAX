import api from "@/shared/utils/axios";
import type { ApiResponse } from "./types";

export interface VendorProduct {
  id: number;
  vendor_id: number;
  name: string;
  type: string;
  model: string;
  api_endpoint: string;
  unit: string;
  price: string;
  currency: string;
  stock_total?: string;
  stock_remaining?: string;
  status: string;
  health_status: string;
}

export interface VendorSales {
  vendor_id: number;
  period: string;
  total_orders: number;
  total_amount: string;
  paid_orders: number;
  paid_amount: string;
  refunded_orders: number;
  refunded_amount: string;
  by_status: { status: string; count: number; amount: string }[];
}

export interface VendorSettlement {
  id: number;
  vendor_id: number;
  settlement_no: string;
  period_start: string;
  period_end: string;
  total_sales: string;
  commission_amount: string;
  net_amount: string;
  currency: string;
  status: string;
  payment_method?: string;
  paid_at?: string;
  remark?: string;
}

export const vendorService = {
  // Profile
  getProfile: () =>
    api.get<ApiResponse<{ id: number; company_name: string; contact_email: string; status: string }>>("/vendor/me"),

  updateProfile: (data: { contact_name?: string; contact_email?: string; contact_phone?: string; api_base_url?: string }) =>
    api.put<ApiResponse<null>>("/vendor/profile", data),

  apply: (data: {
    company_name: string;
    contact_name: string;
    contact_email: string;
    contact_phone: string;
    business_license: string;
    api_base_url: string;
    api_auth_type: string;
    commission_rate: string;
    settlement_cycle: string;
  }) => api.post<ApiResponse<{ vendor_id: number; status: string }>>("/vendor/apply", data),

  // Products
  listProducts: (params?: { status?: string }) =>
    api.get<ApiResponse<VendorProduct[]>>("/vendor/products", { params }),

  createProduct: (data: {
    name: string;
    type: string;
    model: string;
    api_endpoint: string;
    auth_type?: string;
    unit: string;
    price: string;
    currency?: string;
    stock_total?: string;
  }) => api.post<ApiResponse<VendorProduct>>("/vendor/products", data),

  updateProduct: (id: number, data: { name?: string; price?: string; currency?: string; status?: string }) =>
    api.put<ApiResponse<null>>(`/vendor/products/${id}`, data),

  updateProductPrice: (id: number, price: string) =>
    api.put<ApiResponse<null>>(`/vendor/products/${id}/price`, { price }),

  // Sales
  getSales: (period?: string) =>
    api.get<ApiResponse<VendorSales>>("/vendor/sales", { params: { period } }),

  // Settlements
  getSettlements: () =>
    api.get<ApiResponse<VendorSettlement[]>>("/vendor/settlements"),

  confirmSettlement: (id: number) =>
    api.post<ApiResponse<null>>(`/vendor/settlements/${id}/confirm`),

  requestWithdrawal: (id: number) =>
    api.post<ApiResponse<null>>(`/vendor/settlements/${id}/withdraw`),
};
