import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, Switch, Space, Modal, Form, Input, Select, InputNumber, Popconfirm, message } from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";
import type { RiskRule } from "@/shared/api/types";

export default function RuleListPage() {
  const [rules, setRules] = useState<RiskRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchRules = async () => {
    setLoading(true);
    try {
      const res = await adminService.listRiskRules();
      setRules(res.data.data || []);
    } catch {
      message.error("加载规则失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchRules(); }, []);

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await adminService.setRuleEnabled(id, enabled);
      message.success(enabled ? "已启用" : "已禁用");
      fetchRules();
    } catch {
      message.error("操作失败");
    }
  };

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await adminService.createRiskRule(values);
      message.success("规则已创建");
      setModalOpen(false);
      form.resetFields();
      fetchRules();
    } catch {
      // validation error
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "名称", dataIndex: "name", width: 180 },
    { title: "条件", dataIndex: "condition", width: 260, ellipsis: true },
    {
      title: "动作", dataIndex: "action", width: 80,
      render: (a: string) => <Tag color={a === "red" ? "red" : a === "orange" ? "orange" : "gold"}>{a}</Tag>,
    },
    {
      title: "状态", dataIndex: "enabled", width: 80,
      render: (e: boolean, record: RiskRule) => <Switch checked={e} onChange={(v) => handleToggle(record.id, v)} />,
    },
    {
      title: "操作", width: 80,
      render: (_: unknown, record: RiskRule) => (
        <Popconfirm title="确定删除?" onConfirm={() => message.info("删除功能开发中")}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">风控规则</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchRules}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>新建规则</Button>
        </Space>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={rules} loading={loading} pagination={{ pageSize: 20 }} />
      </Card>

      <Modal title="新建风控规则" open={modalOpen} onOk={handleCreate} onCancel={() => setModalOpen(false)} width={480} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="规则名称" rules={[{ required: true }]}>
            <Input placeholder="批量注册检测" />
          </Form.Item>
          <Form.Item name="condition" label="触发条件" rules={[{ required: true }]}>
            <Input.TextArea rows={2} placeholder="同IP 1h 内注册 >5" />
          </Form.Item>
          <Form.Item name="action" label="处理动作" rules={[{ required: true }]}>
            <Select options={[
              { label: "记录 (Green)", value: "green" },
              { label: "预警 (Yellow)", value: "yellow" },
              { label: "限权 (Orange)", value: "orange" },
              { label: "冻结 (Red)", value: "red" },
            ]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
