import { useEffect, useState, useCallback } from "react";
import { Card, Table, Tag, Button, Space, Input, Select, message, Modal } from "antd";
import { SearchOutlined, ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";

interface OrderItem {
  id: string;
  order_no?: string;
  user_id?: number;
  product_name?: string;
  final_amount?: string;
  amount?: string;
  status: string;
  created_at: string;
  payment_method?: string;
}

const STATUS_MAP: Record<string, string> = {
  pending: "orange", paid: "blue", completed: "green",
  cancelled: "default", refunding: "orange", refunded: "red",
};

export default function OrderManagePage() {
  const [orders, setOrders] = useState<OrderItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<OrderItem | null>(null);

  const fetchOrders = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminService.listOrders({ page, size: 20 });
      const data = res.data.data;
      if (data) {
        setOrders((data?.items || []) as unknown as OrderItem[]);
        setTotal(data.total || 0);
      }
    } catch {
      message.error("加载订单列表失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  const handleRefund = async (id: string) => {
    try {
      await adminService.adminRefund(id);
      message.success("退款已处理");
      fetchOrders();
    } catch {
      message.error("退款处理失败");
    }
  };

  const handleDetail = async (orderId: string) => {
    try {
      const res = await adminService.getOrder(orderId);
      const detail = res.data.data as unknown as OrderItem;
      setSelectedOrder(detail);
      setDetailOpen(true);
    } catch {
      message.error("获取订单详情失败");
    }
  };

  const columns = [
    { title: "订单号", dataIndex: "order_no", key: "order_no", width: 160, render: (_: string, r: OrderItem) => r.order_no || r.id },
    { title: "用户 ID", dataIndex: "user_id", width: 80 },
    { title: "产品", dataIndex: "product_name", width: 150, render: (v: string) => v || "-" },
    {
      title: "金额", dataIndex: "final_amount", width: 100,
      render: (_: string, r: OrderItem) => `¥${parseFloat(r.final_amount || r.amount || "0").toFixed(2)}`,
    },
    {
      title: "状态", dataIndex: "status", width: 100,
      render: (s: string) => <Tag color={STATUS_MAP[s] || "default"}>{s}</Tag>,
    },
    { title: "日期", dataIndex: "created_at", width: 170, render: (v: string) => new Date(v).toLocaleString() },
    {
      title: "操作", width: 160,
      render: (_: unknown, record: OrderItem) => (
        <Space>
          <Button size="small" onClick={() => handleDetail(record.id)}>详情</Button>
          {record.status === "refunding" && (
            <Button size="small" danger onClick={() => handleRefund(record.id)}>退款处理</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">交易管理</h2>
        <Button icon={<ReloadOutlined />} onClick={fetchOrders}>刷新</Button>
      </div>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={orders}
          loading={loading}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage, showSizeChanger: true }}
        />
      </Card>

      <Modal
        title="订单详情"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={500}
      >
        {selectedOrder && (
          <div className="space-y-2">
            <p><strong>订单号:</strong> {selectedOrder.order_no || selectedOrder.id}</p>
            <p><strong>用户 ID:</strong> {selectedOrder.user_id}</p>
            <p><strong>产品:</strong> {selectedOrder.product_name || "-"}</p>
            <p><strong>金额:</strong> ¥{parseFloat(selectedOrder.final_amount || selectedOrder.amount || "0").toFixed(2)}</p>
            <p><strong>支付方式:</strong> {selectedOrder.payment_method || "-"}</p>
            <p><strong>状态:</strong> <Tag color={STATUS_MAP[selectedOrder.status]}>{selectedOrder.status}</Tag></p>
            <p><strong>创建时间:</strong> {new Date(selectedOrder.created_at).toLocaleString()}</p>
          </div>
        )}
      </Modal>
    </div>
  );
}
