import { useState, useEffect } from "react";
import { Card, Table, Tabs, Spin, message } from "antd";
import { BarChartOutlined, LineChartOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";

interface ReportRow {
  period: string; orders: number; amount: string; paid_amount: string;
  refund_amount: string; new_users: number;
}

export default function ReportPage() {
  const [loading, setLoading] = useState(true);
  const [dailyData, setDailyData] = useState<ReportRow[]>([]);
  const [monthlyData, setMonthlyData] = useState<ReportRow[]>([]);

  useEffect(() => {
    Promise.all([
      api.get("/admin/reports/daily").then((r) => setDailyData(r.data?.data || [])),
      api.get("/admin/reports/monthly").then((r) => setMonthlyData(r.data?.data || [])),
    ]).catch(() => message.error("加载报表失败")).finally(() => setLoading(false));
  }, []);

  const columns = [
    { title: "周期", dataIndex: "period" },
    { title: "订单数", dataIndex: "orders" },
    { title: "销售额", dataIndex: "amount" },
    { title: "实收", dataIndex: "paid_amount" },
    { title: "退款", dataIndex: "refund_amount" },
    { title: "新用户", dataIndex: "new_users" },
  ];

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">经营报表</h2>
      <Tabs items={[
        { key: "daily", label: <span><BarChartOutlined /> 日报表</span>, children: <Card><Table dataSource={dailyData} rowKey="period" columns={columns} /></Card> },
        { key: "monthly", label: <span><LineChartOutlined /> 月报表</span>, children: <Card><Table dataSource={monthlyData} rowKey="period" columns={columns} /></Card> },
      ]} />
    </div>
  );
}
