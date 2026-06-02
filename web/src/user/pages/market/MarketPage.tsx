import { useEffect, useState } from "react";
import { Card, Table, Tag, Row, Col, Statistic, Progress, Spin, Tabs, Button } from "antd";
import { ThunderboltOutlined, ClockCircleOutlined, CheckCircleOutlined, DollarOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";
import type { ApiResponse } from "@/shared/api/types";

interface ModelInfo {
  id: number;
  name: string;
  provider: string;
  type: string;
  price_per_1k?: number;
  context_window?: number;
  latency_ms?: number;
  capabilities?: string[];
}

interface ProviderHealth {
  id: number;
  provider: string;
  availability: number;
  avg_latency: number;
  p95_latency: number;
  error_rate: number;
  status: "healthy" | "degraded" | "down";
}

export default function MarketPage() {
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [providers, setProviders] = useState<ProviderHealth[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetch() {
      try {
        const [modelsRes, providersRes] = await Promise.all([
          api.get<ApiResponse<ModelInfo[]>>("/models"),
          api.get<ApiResponse<ProviderHealth[]>>("/providers/health"),
        ]);
        setModels(modelsRes.data.data || []);
        setProviders(providersRes.data.data || []);
      } catch { /* fallback */ }
      finally { setLoading(false); }
    }
    fetch();
  }, []);

  const statusColor = (s: string) => s === "healthy" ? "green" : s === "degraded" ? "orange" : "red";

  const providerCols = [
    { title: "供应商", dataIndex: "provider", width: 120 },
    { title: "可用率", dataIndex: "availability", width: 100, render: (v: number) => <Progress percent={Math.round(v)} size="small" /> },
    { title: "平均延迟", dataIndex: "avg_latency", width: 100, render: (v: number) => `${v}ms` },
    { title: "P95 延迟", dataIndex: "p95_latency", width: 100, render: (v: number) => `${v}ms` },
    { title: "错误率", dataIndex: "error_rate", width: 80, render: (v: number) => <Tag color={v > 5 ? "red" : "green"}>{v}%</Tag> },
    { title: "状态", dataIndex: "status", width: 80, render: (s: string) => <Tag color={statusColor(s)}>{s}</Tag> },
  ];

  const modelCols = [
    { title: "模型", dataIndex: "name", width: 150 },
    { title: "供应商", dataIndex: "provider", width: 100 },
    { title: "类型", dataIndex: "type", width: 80 },
    { title: "价格", dataIndex: "price_per_1k", width: 100, render: (v: number) => v ? `¥${v.toFixed(4)}/1K` : "-" },
    { title: "上下文", dataIndex: "context_window", width: 100, render: (v: number) => v ? `${(v / 1000).toFixed(0)}K` : "-" },
    { title: "延迟", dataIndex: "latency_ms", width: 80, render: (v: number) => v ? `${v}ms` : "-" },
    { title: "能力", dataIndex: "capabilities", width: 250, render: (v: string[]) => v?.map((c) => <Tag key={c} className="text-xs">{c}</Tag>) },
  ];

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-6">模型市场</h2>

      <Row gutter={[16, 16]} className="mb-6">
        <Col span={6}><Card><Statistic title="可用模型" value={models.length} prefix={<ThunderboltOutlined />} /></Card></Col>
        <Col span={6}><Card><Statistic title="活跃供应商" value={providers.length} prefix={<CheckCircleOutlined />} /></Card></Col>
        <Col span={6}>
          <Card><Statistic
            title="平均可用率"
            value={providers.length > 0 ? Math.round(providers.reduce((s, p) => s + p.availability, 0) / providers.length) : 0}
            suffix="%" prefix={<ClockCircleOutlined />}
          /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic
            title="最低价格"
            value={models.length > 0 ? Math.min(...models.map((m) => m.price_per_1k || 999)) : 0}
            prefix={<DollarOutlined />} suffix="/1K tokens"
            precision={4}
          /></Card>
        </Col>
      </Row>

      <Tabs
        items={[
          {
            key: "models",
            label: "模型列表",
            children: <Table rowKey="id" columns={modelCols} dataSource={models} pagination={{ pageSize: 20 }} />,
          },
          {
            key: "providers",
            label: "供应商健康",
            children: <Table rowKey="id" columns={providerCols} dataSource={providers} pagination={false} />,
          },
        ]}
      />
    </div>
  );
}
