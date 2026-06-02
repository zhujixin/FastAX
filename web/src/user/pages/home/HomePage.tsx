import { useEffect, useState } from "react";
import { Spin, Tag } from "antd";
import { useNavigate } from "react-router-dom";
import { tokenService } from "@/shared/api/tokens";
import type { TokenProduct } from "@/shared/api/types";

export default function HomePage() {
  const navigate = useNavigate();
  const [products, setProducts] = useState<TokenProduct[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetch() {
      try {
        const res = await tokenService.getProducts();
        setProducts(res.data.data || []);
      } catch {
        // fallback to empty
      } finally {
        setLoading(false);
      }
    }
    fetch();
  }, []);

  return (
    <div className="max-w-6xl mx-auto px-4 py-12">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">FastAX Token 平台</h1>
        <p className="text-lg text-gray-500">一站式 AI Token 代理与交易，连接全球大模型</p>
      </div>

      {loading ? (
        <div className="flex justify-center py-20"><Spin size="large" /></div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {products.map((p) => (
            <div
              key={p.id}
              className="p-6 border rounded-lg hover:shadow-lg transition-shadow cursor-pointer bg-white"
              onClick={() => navigate("/tokens/buy")}
            >
              <div className="flex justify-between items-start mb-2">
                <h3 className="text-xl font-semibold">{p.name}</h3>
                <Tag color={p.status === "active" ? "green" : "red"}>{p.status}</Tag>
              </div>
              <p className="text-gray-500 mb-4">{p.supplier || p.model || "AI 大模型"}</p>
              <div className="flex justify-between items-end">
                <span className="text-2xl font-bold text-blue-600">
                  ¥{parseFloat(p.price as unknown as string || "0").toFixed(4)}/token
                </span>
                <span className="text-sm text-gray-400">库存: {p.stock?.toLocaleString() || "0"}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
