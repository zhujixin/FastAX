import { useState, useRef, useCallback } from "react";

interface SSEMessage {
  id?: string;
  content: string;
  role?: string;
  finish_reason?: string;
}

interface UseSSEOptions {
  onMessage?: (msg: SSEMessage) => void;
  onDone?: () => void;
  onError?: (err: Error) => void;
}

export function useSSE({ onMessage, onDone, onError }: UseSSEOptions = {}) {
  const [streaming, setStreaming] = useState(false);
  const [content, setContent] = useState("");
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const startStream = useCallback(
    async (url: string, body: Record<string, unknown>, apiKey?: string) => {
      setStreaming(true);
      setContent("");
      setError(null);

      const controller = new AbortController();
      abortRef.current = controller;

      try {
        const token = apiKey || localStorage.getItem("fastax-auth")
          ? JSON.parse(localStorage.getItem("fastax-auth") || "{}")?.state?.accessToken
          : null;

        const res = await fetch(url, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify({ ...body, stream: true }),
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }

        const reader = res.body?.getReader();
        if (!reader) throw new Error("No readable stream");

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || ""; // Keep incomplete line in buffer

          for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed || !trimmed.startsWith("data: ")) continue;

            const data = trimmed.slice(6); // Remove "data: " prefix
            if (data === "[DONE]") {
              setStreaming(false);
              onDone?.();
              return;
            }

            try {
              const parsed = JSON.parse(data);
              const delta = parsed.choices?.[0]?.delta?.content || "";
              setContent((prev) => prev + delta);
              onMessage?.({
                id: parsed.id,
                content: delta,
                role: parsed.choices?.[0]?.delta?.role,
                finish_reason: parsed.choices?.[0]?.finish_reason,
              });
            } catch {
              // Skip unparseable lines
            }
          }
        }
        setStreaming(false);
        onDone?.();
      } catch (err: any) {
        if (err.name !== "AbortError") {
          const msg = err.message || "Stream error";
          setError(msg);
          onError?.(err);
        }
        setStreaming(false);
      }
    },
    [onMessage, onDone, onError]
  );

  const stopStream = useCallback(() => {
    abortRef.current?.abort();
    setStreaming(false);
  }, []);

  return { streaming, content, error, startStream, stopStream };
}
