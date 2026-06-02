import api from "@/shared/utils/axios";
import type { ApiResponse, Notification } from "./types";

export const notifyService = {
  list: () =>
    api.get<ApiResponse<Notification[]>>("/notifications"),

  getUnreadCount: () =>
    api.get<ApiResponse<{ count: number }>>("/notifications/unread-count"),

  markRead: (id: number) =>
    api.put<ApiResponse<null>>(`/notifications/${id}/read`),

  markAllRead: () =>
    api.put<ApiResponse<null>>("/notifications/read-all"),
};
