import { useEffect, useState } from "react";
import { Card, Row, Col, Statistic, Spin } from "antd";
import { UserOutlined, DollarOutlined, ApiOutlined, ShoppingCartOutlined } from "@ant-design/icons";
import ReactECharts from "echarts-for-react";
import { adminService } from "@/shared/api/admin";
import type { DashboardCharts } from "@/shared/api/types";

export default function AdminDashboard() {
  const [loading, setLoading] = useState(true);
  const [summary, setSummary] = useState({ totalUsers: 0, todayRevenue: "0", totalOrders: 0, todayNewUsers: 0, todayNewOrders: 0, activeTokens: 0 });
  const [charts, setCharts] = useState<DashboardCharts | null>(null);

  useEffect(() => {
    async function fetch() {
      try {
        const [summaryRes, chartsRes] = await Promise.all([
          adminService.getDashboardSummary(),
          adminService.getDashboardCharts("7d"),
        ]);
        const s = summaryRes.data.data;
        if (s) {
          setSummary({
            totalUsers: s.total_users || 0,
            todayRevenue: s.today_revenue || "0",
            totalOrders: s.total_orders || 0,
            todayNewUsers: s.today_new_users || 0,
            todayNewOrders: s.today_new_orders || 0,
            activeTokens: s.active_tokens || 0,
          });
        }
        if (chartsRes.data.data) {
          setCharts(chartsRes.data.data);
        }
      } catch {
        // fallback to empty state
      } finally {
        setLoading(false);
      }
    }
    fetch();
  }, []);

  const revenueOption = {
    tooltip: { trigger: "axis" as const },
    xAxis: { type: "category" as const, data: charts?.revenue.map((d) => d.date.slice(5)) || [] },
    yAxis: { type: "value" as const },
    series: [{
      name: "收入",
      type: "line",
      data: charts?.revenue.map((d) => parseFloat(d.amount || "0")) || [],
      smooth: true,
      areaStyle: { opacity: 0.1 },
      itemStyle: { color: "#1677ff" },
    }],
  };

  const tokenOption = {
    tooltip: { trigger: "axis" as const },
    xAxis: { type: "category" as const, data: charts?.tokens.map((d) => d.date.slice(5)) || [] },
    yAxis: { type: "value" as const },
    series: [{
      name: "Token 消耗",
      type: "bar",
      data: charts?.tokens.map((d) => d.value || 0) || [],
      itemStyle: { color: "#52c41a" },
    }],
  };

  const ordersOption = {
    tooltip: { trigger: "axis" as const },
    xAxis: { type: "category" as const, data: charts?.orders.map((d) => d.date.slice(5)) || [] },
    yAxis: { type: "value" as const },
    series: [{
      name: "新订单",
      type: "bar",
      data: charts?.orders.map((d) => d.value || 0) || [],
      itemStyle: { color: "#fa8c16" },
    }],
  };

  const usersOption = {
    tooltip: { trigger: "item" as const },
    series: [{
      name: "用户增长",
      type: "line",
      data: charts?.users.map((d) => d.value || 0) || [],
      smooth: true,
      itemStyle: { color: "#722ed1" },
    }],
    xAxis: { type: "category" as const, data: charts?.users.map((d) => d.date.slice(5)) || [], show: false },
    yAxis: { type: "value" as const, show: false },
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64"><Spin size="large" /></div>;
  }

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">运营看板</h2>
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="总用户" value={summary.totalUsers} prefix={<UserOutlined />} suffix={<span className="text-xs text-green-500">+{summary.todayNewUsers} 今日</span>} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="今日收入 ¥" value={summary.todayRevenue} prefix={<DollarOutlined />} precision={2} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="活跃 Token" value={summary.activeTokens} prefix={<ApiOutlined />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="总订单" value={summary.totalOrders} prefix={<ShoppingCartOutlined />} suffix={<span className="text-xs text-blue-500">+{summary.todayNewOrders} 今日</span>} /></Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="近7日收入趋势"><ReactECharts option={revenueOption} style={{ height: 300 }} /></Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="近7日 Token 消耗"><ReactECharts option={tokenOption} style={{ height: 300 }} /></Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="近7日新订单"><ReactECharts option={ordersOption} style={{ height: 260 }} /></Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="近7日新用户"><ReactECharts option={usersOption} style={{ height: 260 }} /></Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="快捷入口" className="text-center">
            <div className="grid grid-cols-2 gap-3 pt-4">
              <a href="/admin/users" className="p-3 bg-blue-50 rounded text-blue-600 hover:bg-blue-100">用户管理</a>
              <a href="/admin/orders" className="p-3 bg-green-50 rounded text-green-600 hover:bg-green-100">交易管理</a>
              <a href="/admin/tokens" className="p-3 bg-orange-50 rounded text-orange-600 hover:bg-orange-100">Token 管理</a>
              <a href="/admin/guardrails/rules" className="p-3 bg-purple-50 rounded text-purple-600 hover:bg-purple-100">安全护栏</a>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
