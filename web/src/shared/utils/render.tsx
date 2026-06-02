import { Tag } from "antd";

// Human-readable number: 1000000 → "1M"
export function renderNumber(num: number): string {
  if (num >= 1e9) return (num / 1e9).toFixed(1) + "B";
  if (num >= 1e6) return (num / 1e6).toFixed(1) + "M";
  if (num >= 1e3) return (num / 1e3).toFixed(1) + "k";
  return num.toString();
}

// Truncate text
export function renderText(text: string, limit = 30): string {
  if (!text) return "";
  return text.length > limit ? text.slice(0, limit) + "..." : text;
}

// Deterministic color label
const LABEL_COLORS = ["magenta", "red", "volcano", "orange", "gold", "lime", "green", "cyan", "blue", "geekblue", "purple"];

export function renderColorLabel(text: string) {
  let hash = 0;
  for (let i = 0; i < text.length; i++) {
    hash = (hash * 31 + text.charCodeAt(i)) & 0xffffffff;
  }
  return <Tag color={LABEL_COLORS[Math.abs(hash) % LABEL_COLORS.length]}>{text}</Tag>;
}

// Group labels (vip/pro → gold, svip/premium → red)
export function renderGroup(groups: string) {
  if (!groups) return null;
  return groups.split(",").map((g) => {
    const trimmed = g.trim();
    const isVip = ["vip", "pro"].includes(trimmed.toLowerCase());
    const isSvip = ["svip", "premium"].includes(trimmed.toLowerCase());
    return <Tag key={trimmed} color={isSvip ? "red" : isVip ? "gold" : "default"}>{trimmed}</Tag>;
  });
}

// Channel status
export function renderStatus(status: number | string) {
  const s = Number(status);
  if (s === 1) return <Tag color="green">启用</Tag>;
  if (s === 2) return <Tag color="red">禁用</Tag>;
  if (s === 3) return <Tag color="orange">熔断</Tag>;
  return <Tag color="default">未知</Tag>;
}

// Response time color
export function renderResponseTime(ms: number) {
  if (ms < 1000) return <Tag color="green">{ms}ms</Tag>;
  if (ms < 3000) return <Tag color="gold">{ms}ms</Tag>;
  return <Tag color="red">{ms}ms</Tag>;
}

// Quota display
export function renderQuota(quota: number): string {
  if (quota === undefined || quota === null) return "-";
  if (quota >= 1e9) return (quota / 1e9).toFixed(2) + "B";
  if (quota >= 1e7) return (quota / 1e7).toFixed(0) + "0M";
  if (quota >= 1e6) return (quota / 1e6).toFixed(2) + "M";
  return quota.toLocaleString();
}
