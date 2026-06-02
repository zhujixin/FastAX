import { useEffect, useState, useCallback } from "react";
import { Card, Table, Tag, message } from "antd";
import { statsService } from "@/shared/api/stats";

interface BillItem {
  id: number;
  order_id: number;
  payment_no: string;
  amount: string;
  method: string;
  gateway: string;
  status: string;
  created_at: string;
}

export default function BillsPage() {
  const [bills, setBills] = useState<BillItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const fetchBills = useCallback(async () => {
    setLoading(true);
    try {
      const res = await statsService.getBills({ page, page_size: 20 });
      const data = res.data.data;
      if (data) {
        setBills((data.items as unknown as BillItem[]) || []);
        setTotal(data.total || 0);
      }
    } catch {
      message.error("加载账单失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchBills(); }, [fetchBills]);

  const statusMap: Record<string, string> = {
    success: "green", paid: "green", pending: "orange", refunded: "red",
  };

  const columns = [
    { title: "账单 ID", dataIndex: "id", width: 80 },
    { title: "订单 ID", dataIndex: "order_id", width: 80 },
    { title: "支付单号", dataIndex: "payment_no", width: 180, ellipsis: true },
    { title: "金额", dataIndex: "amount", width: 100, render: (v: string) => `¥${parseFloat(v || "0").toFixed(2)}` },
    { title: "方式", dataIndex: "gateway", width: 80, render: (v: string, r: BillItem) => (v || r.method || "-") as string },
    {
      title: "状态", dataIndex: "status", width: 80,
      render: (s: string) => <Tag color={statusMap[s] || "default"}>{s}</Tag>,
    },
    { title: "时间", dataIndex: "created_at", width: 170, render: (v: string) => v ? new Date(v).toLocaleString() : "-" },
  ];

  return (
    <div className="max-w-4xl mx-auto p-4">
      <h2 className="text-xl font-bold mb-4">账单明细</h2>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={bills}
          loading={loading}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        />
      </Card>
    </div>
  );
}
