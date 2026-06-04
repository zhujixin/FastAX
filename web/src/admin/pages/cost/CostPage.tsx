import { useState, useEffect } from "react";
import { Card, Tabs, Statistic, Row, Col, Slider, InputNumber, Button, Form, Switch, message, Spin } from "antd";
import { DollarOutlined, DatabaseOutlined, AlertOutlined, SettingOutlined } from "@ant-design/icons";
import { costService, type CacheStats } from "@/shared/api/cost";
import { statsService } from "@/shared/api";

export default function CostPage() {
  const [budgetForm] = Form.useForm();
  const [cacheForm] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [cacheStats, setCacheStats] = useState<CacheStats | null>(null);
  const [budget, setBudget] = useState<{ limit: number; spent: number; usage_pct: number }>({ limit: 0, spent: 0, usage_pct: 0 });

  useEffect(() => {
    Promise.all([
      costService.getCacheStats().then((r) => setCacheStats(r.data.data)),
      costService.getBudget().then((r) => {
        const d = r.data.data as { limit: number; spent: number; usage_pct: number } | null;
        if (d) setBudget(d);
      }).catch(() => {}), // budget may not be set
    ]).finally(() => setLoading(false));
  }, []);

  const handleSaveBudget = async () => {
    const values = await budgetForm.validateFields();
    try {
      await costService.setBudget("monthly", values.monthly_limit);
      await costService.setAlert([values.alert_threshold]);
      message.success("预算配置已保存");
    } catch { message.error("保存失败"); }
  };

  const handleSaveCache = async () => {
    const values = await cacheForm.validateFields();
    try {
      await costService.updateCacheConfig({
        enabled: values.enabled,
        similarity_threshold: values.similarity_threshold,
        ttl_seconds: values.ttl_minutes * 60,
        max_entries: values.max_entries,
      });
      message.success("缓存配置已保存");
      const stats = await costService.getCacheStats();
      setCacheStats(stats.data.data);
    } catch { message.error("保存失败"); }
  };

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  const budgetTab = (
    <div>
      <Row gutter={[16, 16]} className="mb-6">
        <Col span={8}>
          <Card><Statistic title="本月已消费" value={budget.spent} prefix={<DollarOutlined />} precision={2} suffix="元" /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="月度预算" value={budget.limit} prefix={<SettingOutlined />} precision={0} suffix="元" /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="使用率" value={budget.usage_pct} prefix={<AlertOutlined />} precision={1} suffix="%" /></Card>
        </Col>
      </Row>
      <Card title="预算封顶配置" className="mb-4">
        <Form form={budgetForm} layout="vertical" className="max-w-md">
          <Form.Item name="monthly_limit" label="月度预算上限 (元)" initialValue={budget.limit || 5000}>
            <InputNumber min={0} step={100} className="w-full" />
          </Form.Item>
          <Form.Item name="alert_threshold" label="告警阈值 (%)" initialValue={80}>
            <Slider min={50} max={100} marks={{ 50: "50%", 80: "80%", 90: "90%", 100: "100%" }} />
          </Form.Item>
          <Button type="primary" onClick={handleSaveBudget}>保存预算配置</Button>
        </Form>
      </Card>
    </div>
  );

  const cacheTab = (
    <div>
      <Row gutter={[16, 16]} className="mb-6">
        <Col span={6}><Card><Statistic title="缓存条目" value={cacheStats?.total_entries || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="活跃条目" value={cacheStats?.active_entries || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="总命中" value={cacheStats?.total_hits || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="命中率" value={((cacheStats?.hit_rate || 0) * 100).toFixed(1)} suffix="%" /></Card></Col>
      </Row>
      <Card title="语义缓存配置">
        <Form form={cacheForm} layout="vertical" className="max-w-md">
          <Form.Item name="enabled" label="启用语义缓存" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
          <Form.Item name="similarity_threshold" label="相似度阈值" initialValue={0.85}>
            <Slider min={0.7} max={0.99} step={0.01} marks={{ 0.7: "0.7", 0.85: "0.85", 0.92: "0.92", 0.99: "0.99" }} />
          </Form.Item>
          <Form.Item name="ttl_minutes" label="缓存 TTL (分钟)" initialValue={60}>
            <InputNumber min={1} max={1440} className="w-full" />
          </Form.Item>
          <Form.Item name="max_entries" label="最大缓存条目" initialValue={10000}>
            <InputNumber min={100} step={1000} className="w-full" />
          </Form.Item>
          <Button type="primary" onClick={handleSaveCache}>保存缓存配置</Button>
        </Form>
      </Card>
    </div>
  );

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">成本优化</h2>
      <Tabs
        items={[
          { key: "budget", label: <span><DollarOutlined /> 预算管理</span>, children: budgetTab },
          { key: "cache", label: <span><DatabaseOutlined /> 语义缓存</span>, children: cacheTab },
        ]}
      />
    </div>
  );
}
