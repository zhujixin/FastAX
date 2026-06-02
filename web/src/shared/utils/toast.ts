import { message, notification } from "antd";

// Unified response envelope from backend
export interface APIEnvelope<T = unknown> {
  code: number;
  message: string;
  data: T;
  trace_id?: string;
}

// Toast durations (ms)
export const TOAST_DURATION = {
  SUCCESS: 3000,
  INFO: 5000,
  WARNING: 6000,
  ERROR: 8000,
};

export function showSuccess(msg: string) {
  message.success(msg, TOAST_DURATION.SUCCESS / 1000);
}

export function showError(err: unknown) {
  if (err && typeof err === "object" && "response" in err) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } }; message?: string };
    const status = axiosErr.response?.status;
    switch (status) {
      case 401:
        if (!window.location.pathname.includes("/login")) {
          window.location.href = "/login?expired=true";
        }
        return;
      case 429:
        message.error("请求过于频繁，请稍后再试", TOAST_DURATION.ERROR / 1000);
        return;
      case 500:
        message.error("服务器内部错误", TOAST_DURATION.ERROR / 1000);
        return;
    }
    message.error(axiosErr.response?.data?.message || axiosErr.message || "请求失败", TOAST_DURATION.ERROR / 1000);
    return;
  }
  if (err instanceof Error) {
    message.error(err.message, TOAST_DURATION.ERROR / 1000);
    return;
  }
  message.error(String(err), TOAST_DURATION.ERROR / 1000);
}

export function showWarning(msg: string) {
  message.warning(msg, TOAST_DURATION.WARNING / 1000);
}

export function showInfo(msg: string) {
  message.info(msg, TOAST_DURATION.INFO / 1000);
}

export function showNotice(msg: string) {
  notification.info({ message: "系统通知", description: msg, duration: 0 });
}
