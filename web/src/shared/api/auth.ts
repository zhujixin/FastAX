import api from "@/shared/utils/axios";
import type { ApiResponse, LoginRequest, LoginResponse, RegisterRequest, UserInfo } from "./types";

export const authService = {
  login: (data: LoginRequest) =>
    api.post<ApiResponse<LoginResponse>>("/auth/login", data),

  register: (data: RegisterRequest) =>
    api.post<ApiResponse<null>>("/auth/register", data),

  refresh: (refreshToken: string) =>
    api.post<ApiResponse<LoginResponse>>("/auth/refresh", { refresh_token: refreshToken }),

  logout: () =>
    api.post<ApiResponse<null>>("/auth/logout"),

  sendCode: (account: string) =>
    api.post<ApiResponse<null>>("/auth/send-code", { account }),

  resetPassword: (account: string) =>
    api.post<ApiResponse<null>>("/auth/reset-password", { account }),

  getOAuthUrl: (provider: string) =>
    `/api/auth/oauth/${provider}`,

  oauthLogin: (provider: string, code: string) =>
    api.post<ApiResponse<LoginResponse>>("/auth/oauth/login", { provider, code }),
};
