import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, App } from "antd";
import { useNavigate } from "react-router-dom";
import { orderService } from "@/shared/api";
import type { Order } from "@/shared/api/types";

export default function OrderListPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    orderService.list()
      .then((res) => setOrders(res.data.data?.items || []))
      .catch(() => message.error("加载失败"))
      .finally(() => setLoading(false));
  }, []);

  const statusMap: Record<string, { color: string; text: string }> = {
    paid: { color: "blue", text: "已支付" },
    completed: { color: "green", text: "已完成" },
    refunding: { color: "orange", text: "退款中" },
    refunded: { color: "red", text: "已退款" },
    cancelled: { color: "default", text: "已取消" },
  };

  const columns = [
    { title: "订单号", dataIndex: "id", key: "id" },
    { title: "产品", dataIndex: "product_name", key: "product" },
    { title: "金额", dataIndex: "amount", key: "amount", render: (v: number) => `¥${Number(v).toFixed(2)}` },
    { title: "状态", dataIndex: "status", key: "status", render: (s: string) => {
      const m = statusMap[s] || { color: "default", text: s };
      return <Tag color={m.color}>{m.text}</Tag>;
    }},
    { title: "创建时间", dataIndex: "created_at", key: "created", render: (v: string) => v ? new Date(v).toLocaleString() : "-" },
    { title: "操作", key: "action", render: (_: unknown, r: Order) => <Button size="small" onClick={() => navigate(`/orders/${r.id}`)}>详情</Button> },
  ];

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-6">我的订单</h2>
      <Card>
        <Table dataSource={orders} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </Card>
    </div>
  );
}
