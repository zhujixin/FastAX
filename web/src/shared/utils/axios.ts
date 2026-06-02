import axios from "axios";
import { useAuthStore } from "@/shared/stores/authStore";

const api = axios.create({
  baseURL: "/api",
  timeout: 30000,
  headers: { "Content-Type": "application/json" },
});

// Request interceptor: attach JWT + language header
api.interceptors.request.use((config) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  const lang = localStorage.getItem("i18nextLng") || "zh-CN";
  config.headers["Accept-Language"] = lang;

  return config;
});

// Response interceptor: handle 401 refresh, unified errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;

    // Token expired → try refresh
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;
      const { refreshToken } = useAuthStore.getState();
      if (refreshToken) {
        try {
          const res = await axios.post("/api/auth/refresh", {
            refresh_token: refreshToken,
          });
          const { access_token, refresh_token, user } = res.data.data;
          useAuthStore.getState().setAuth(access_token, refresh_token, user);
          original.headers.Authorization = `Bearer ${access_token}`;
          return api(original);
        } catch {
          useAuthStore.getState().logout();
          window.location.href = "/login";
        }
      } else {
        useAuthStore.getState().logout();
        window.location.href = "/login";
      }
    }

    return Promise.reject(error);
  }
);

export default api;
