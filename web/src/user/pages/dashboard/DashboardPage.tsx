import { useEffect, useState } from "react";
import { Card, Statistic, Row, Col, Spin } from "antd";
import { KeyOutlined, ApiOutlined, DollarOutlined, ArrowUpOutlined, ShoppingCartOutlined } from "@ant-design/icons";
import ReactECharts from "echarts-for-react";
import { statsService } from "@/shared/api/stats";

export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [summary, setSummary] = useState({ totalTokens: 0, totalAmount: "0", monthTokens: 0, monthAmount: "0", totalRequests: 0, totalOrders: 0 });
  const [usageData, setUsageData] = useState<number[]>([]);
  const [usageLabels, setUsageLabels] = useState<string[]>([]);

  useEffect(() => {
    async function fetch() {
      try {
        const [summaryRes, usageRes] = await Promise.all([
          statsService.getSummary(),
          statsService.getUsage("week"),
        ]);
        const s = summaryRes.data.data;
        if (s) {
          setSummary({
            totalTokens: s.total_tokens || 0,
            totalAmount: s.total_amount || "0",
            monthTokens: s.month_tokens || 0,
            monthAmount: s.month_amount || "0",
            totalRequests: s.total_requests || 0,
            totalOrders: s.total_orders || 0,
          });
        }
        // Usage data - generate placeholder if API returns aggregated values
        if (usageRes.data.data) {
          const d = usageRes.data.data as any;
          if (d.daily) {
            setUsageLabels(d.daily.map((item: any) => item.date?.slice(5)));
            setUsageData(d.daily.map((item: any) => item.tokens));
          } else {
            setUsageLabels(["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]);
            setUsageData([0, 0, 0, 0, 0, 0, 0]);
          }
        }
      } catch {
        // use defaults
      } finally {
        setLoading(false);
      }
    }
    fetch();
  }, []);

  const usageOption = {
    tooltip: { trigger: "axis" as const },
    grid: { top: 20, right: 20, bottom: 30, left: 50 },
    xAxis: { type: "category" as const, data: usageLabels },
    yAxis: { type: "value" as const },
    series: [{
      name: "Token 用量",
      type: "line",
      data: usageData,
      smooth: true,
      areaStyle: { opacity: 0.15 },
      itemStyle: { color: "#1677ff" },
    }],
  };

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-6">控制台</h2>
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="累计 Token" value={summary.totalTokens.toLocaleString()} prefix={<KeyOutlined />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="API 调用次数" value={summary.totalRequests.toLocaleString()} prefix={<ApiOutlined />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="累计消费" value={summary.totalAmount} prefix={<DollarOutlined />} suffix="元" /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="订单数量" value={summary.totalOrders} prefix={<ShoppingCartOutlined />} /></Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="近 7 天 Token 用量"><ReactECharts option={usageOption} style={{ height: 320 }} /></Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="本月概览" className="space-y-4">
            <div className="flex justify-between"><span className="text-gray-500">本月 Token</span><span className="font-bold">{summary.monthTokens.toLocaleString()}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">本月消费</span><span className="font-bold">¥{summary.monthAmount}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">API 调用</span><span className="font-bold">{summary.totalRequests.toLocaleString()}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">订单数</span><span className="font-bold">{summary.totalOrders}</span></div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
