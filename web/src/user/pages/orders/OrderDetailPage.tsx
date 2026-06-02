import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, Descriptions, Tag, Button, App, Spin } from "antd";
import { orderService } from "@/shared/api";
import type { OrderDetail } from "@/shared/api/types";

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    orderService.get(id)
      .then((res) => setOrder(res.data.data))
      .catch(() => message.error("加载失败"))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;
  if (!order) return <div className="text-center py-20 text-gray-500">订单不存在</div>;

  const handleRefund = async () => {
    try {
      await orderService.requestRefund(order.id, "用户申请退款");
      message.success("退款申请已提交");
      setOrder({ ...order, status: "refunding" });
    } catch (err: any) {
      message.error(err.response?.data?.message || "退款失败");
    }
  };

  const statusMap: Record<string, { color: string; text: string }> = {
    paid: { color: "blue", text: "已支付" },
    completed: { color: "green", text: "已完成" },
    refunding: { color: "orange", text: "退款中" },
    refunded: { color: "red", text: "已退款" },
    cancelled: { color: "default", text: "已取消" },
  };

  const sm = statusMap[order.status] || { color: "default", text: order.status };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <Button className="mb-4" onClick={() => navigate("/orders")}>← 返回列表</Button>
      <Card title="订单详情">
        <Descriptions column={1} bordered>
          <Descriptions.Item label="订单号">{order.id}</Descriptions.Item>
          <Descriptions.Item label="产品">{order.product_name}</Descriptions.Item>
          <Descriptions.Item label="数量">{order.quantity?.toLocaleString()} tokens</Descriptions.Item>
          <Descriptions.Item label="金额">¥{Number(order.amount).toFixed(2)}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={sm.color}>{sm.text}</Tag></Descriptions.Item>
          <Descriptions.Item label="支付方式">{order.payment_method || "-"}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{order.created_at ? new Date(order.created_at).toLocaleString() : "-"}</Descriptions.Item>
        </Descriptions>
        {order.status === "paid" && (
          <div className="mt-6">
            <Button danger onClick={handleRefund}>申请退款</Button>
          </div>
        )}
      </Card>
    </div>
  );
}
