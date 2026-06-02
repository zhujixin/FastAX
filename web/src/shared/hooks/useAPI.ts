import { useState, useCallback } from "react";
import type { AxiosResponse } from "axios";

export function useAPI<T = unknown>() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<T | null>(null);

  const call = useCallback(async (apiCall: () => Promise<AxiosResponse>) => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiCall();
      setData(res.data?.data ?? res.data);
      return res.data?.data ?? res.data;
    } catch (err: any) {
      const msg = err.response?.data?.message || err.message || "请求失败";
      setError(msg);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { loading, error, data, call, setData };
}
