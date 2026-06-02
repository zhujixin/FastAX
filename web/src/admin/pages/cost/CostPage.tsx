import { useState } from "react";
import { Card, Tabs, Statistic, Row, Col, Slider, InputNumber, Button, Form, Switch, message, Table, Tag } from "antd";
import { DollarOutlined, DatabaseOutlined, AlertOutlined, SettingOutlined } from "@ant-design/icons";

export default function CostPage() {
  const [budgetForm] = Form.useForm();
  const [cacheForm] = Form.useForm();

  const handleSaveBudget = () => {
    message.success("预算配置已保存 (API 开发中)");
  };

  const handleSaveCache = () => {
    message.success("缓存配置已保存 (API 开发中)");
  };

  const budgetTab = (
    <div>
      <Row gutter={[16, 16]} className="mb-6">
        <Col span={8}>
          <Card><Statistic title="本月已消费" value={1280.50} prefix={<DollarOutlined />} precision={2} suffix="元" /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="月度预算" value={5000} prefix={<SettingOutlined />} precision={0} suffix="元" /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="缓存命中率" value={34.5} prefix={<DatabaseOutlined />} precision={1} suffix="%" /></Card>
        </Col>
      </Row>
      <Card title="预算封顶配置" className="mb-4">
        <Form form={budgetForm} layout="vertical" className="max-w-md">
          <Form.Item name="monthly_limit" label="月度预算上限 (元)" initialValue={5000}>
            <InputNumber min={0} step={100} className="w-full" />
          </Form.Item>
          <Form.Item name="daily_limit" label="日预算上限 (元)" initialValue={500}>
            <InputNumber min={0} step={50} className="w-full" />
          </Form.Item>
          <Form.Item name="per_user_limit" label="单用户预算上限 (元)" initialValue={1000}>
            <InputNumber min={0} step={100} className="w-full" />
          </Form.Item>
          <Form.Item name="alert_threshold" label="告警阈值 (%)" initialValue={80}>
            <Slider min={50} max={100} marks={{ 50: "50%", 80: "80%", 90: "90%", 100: "100%" }} />
          </Form.Item>
          <Button type="primary" onClick={handleSaveBudget}>保存预算配置</Button>
        </Form>
      </Card>
      <Card title="成本告警规则">
        <Table
          dataSource={[
            { id: 1, type: "月度预算", threshold: "80%", action: "通知管理员", status: "active" },
            { id: 2, type: "日预算", threshold: "90%", action: "限流", status: "active" },
          ]}
          rowKey="id"
          pagination={false}
          columns={[
            { title: "类型", dataIndex: "type" },
            { title: "阈值", dataIndex: "threshold" },
            { title: "动作", dataIndex: "action" },
            { title: "状态", dataIndex: "status", render: (s: string) => <Tag color={s === "active" ? "green" : "default"}>{s}</Tag> },
          ]}
        />
      </Card>
    </div>
  );

  const cacheTab = (
    <div>
      <Row gutter={[16, 16]} className="mb-6">
        <Col span={6}><Card><Statistic title="缓存条目" value={12580} /></Card></Col>
        <Col span={6}><Card><Statistic title="命中率" value={34.5} suffix="%" /></Card></Col>
        <Col span={6}><Card><Statistic title="节省 Token" value={425000} /></Card></Col>
        <Col span={6}><Card><Statistic title="节省费用" value={63.75} prefix="¥" precision={2} /></Card></Col>
      </Row>
      <Card title="语义缓存配置">
        <Form form={cacheForm} layout="vertical" className="max-w-md">
          <Form.Item name="enabled" label="启用语义缓存" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
          <Form.Item name="similarity_threshold" label="相似度阈值" initialValue={0.92}>
            <Slider min={0.7} max={0.99} step={0.01} marks={{ 0.7: "0.7", 0.85: "0.85", 0.92: "0.92", 0.99: "0.99" }} />
          </Form.Item>
          <Form.Item name="ttl_minutes" label="缓存 TTL (分钟)" initialValue={60}>
            <InputNumber min={1} max={1440} className="w-full" />
          </Form.Item>
          <Form.Item name="max_entries" label="最大缓存条目" initialValue={100000}>
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
