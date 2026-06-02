// Navigation shorthand
let _navigate: ((path: string) => void) | null = null;

export function setNavigate(nav: (path: string) => void) {
  _navigate = nav;
}

export function navigateTo(path: string) {
  _navigate?.(path);
}

// Timestamp conversion (Unix seconds → YYYY-MM-DD HH:mm:ss)
export function timestamp2string(ts: number): string {
  const d = new Date(ts * 1000);
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

// Clipboard
export async function copy(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    // Fallback
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    try {
      document.execCommand("copy");
      return true;
    } catch {
      return false;
    } finally {
      document.body.removeChild(textarea);
    }
  }
}

// JSON validation
export function verifyJSON(str: string): boolean {
  try { JSON.parse(str); return true; } catch { return false; }
}

// URL helpers
export function removeTrailingSlash(url: string): string {
  return url.replace(/\/+$/, "");
}

// Role checks
export function isAdmin(): boolean {
  try {
    const auth = JSON.parse(localStorage.getItem("fastax-auth") || "{}");
    return auth?.state?.user?.role === "admin" || auth?.state?.user?.role === "super_admin";
  } catch { return false; }
}

export function isRoot(): boolean {
  try {
    const auth = JSON.parse(localStorage.getItem("fastax-auth") || "{}");
    return auth?.state?.user?.role === "super_admin";
  } catch { return false; }
}

// Mobile detection
export function isMobile(): boolean {
  return window.innerWidth <= 768;
}

// Download file
export function downloadTextAsFile(text: string, filename: string) {
  const blob = new Blob([text], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
