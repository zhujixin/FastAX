import { useEffect, useState, useCallback } from "react";
import { Table, Button, Modal, Form, Input, Select, InputNumber, Switch, Space, Popconfirm, Tag, message } from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";
import type { GuardrailRule } from "@/shared/api/types";

const STAGE_OPTIONS = [
  { label: "输入检测 (Before)", value: "before" },
  { label: "输出检测 (After)", value: "after" },
];
const TYPE_OPTIONS = [
  { label: "PII (个人隐私)", value: "pii" },
  { label: "注入攻击", value: "injection" },
  { label: "密钥扫描", value: "secret" },
  { label: "内容审核", value: "content" },
];
const ACTION_OPTIONS = [
  { label: "阻断 (Enforce)", value: "enforce" },
  { label: "监控 (Monitor)", value: "monitor" },
  { label: "仅记录 (Log)", value: "log" },
];

export default function GuardrailRulesPage() {
  const [rules, setRules] = useState<GuardrailRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<GuardrailRule | null>(null);
  const [form] = Form.useForm();

  const fetchRules = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminService.listGuardrailRules();
      setRules(res.data.data || []);
    } catch {
      message.error("加载护栏规则失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchRules(); }, [fetchRules]);

  const handleCreate = () => {
    setEditingRule(null);
    form.resetFields();
    form.setFieldsValue({ priority: 0 });
    setModalOpen(true);
  };

  const handleEdit = (rule: GuardrailRule) => {
    setEditingRule(rule);
    form.setFieldsValue(rule);
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await adminService.deleteGuardrailRule(id);
      message.success("规则已删除");
      fetchRules();
    } catch {
      message.error("删除失败");
    }
  };

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await adminService.setGuardrailRuleEnabled(id, enabled);
      message.success(enabled ? "规则已启用" : "规则已禁用");
      fetchRules();
    } catch {
      message.error("操作失败");
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingRule) {
        await adminService.updateGuardrailRule(editingRule.id, values);
        message.success("规则已更新");
      } else {
        await adminService.createGuardrailRule(values);
        message.success("规则已创建");
      }
      setModalOpen(false);
      fetchRules();
    } catch {
      // validation error
    }
  };

  const stageColor = (stage: string) => stage === "before" ? "blue" : "green";
  const actionColor = (action: string) => {
    switch (action) {
      case "enforce": return "red";
      case "monitor": return "orange";
      default: return "default";
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "规则名称", dataIndex: "name" },
    {
      title: "阶段", dataIndex: "stage", width: 120,
      render: (v: string) => <Tag color={stageColor(v)}>{v === "before" ? "输入检测" : "输出检测"}</Tag>,
    },
    {
      title: "类型", dataIndex: "type", width: 120,
      render: (v: string) => {
        const m: Record<string, string> = { pii: "PII", injection: "注入", secret: "密钥", content: "内容" };
        return m[v] || v;
      },
    },
    {
      title: "动作", dataIndex: "action", width: 100,
      render: (v: string) => <Tag color={actionColor(v)}>{v}</Tag>,
    },
    { title: "优先级", dataIndex: "priority", width: 80 },
    {
      title: "状态", dataIndex: "enabled", width: 80,
      render: (v: number, record: GuardrailRule) => (
        <Switch
          checked={v === 1}
          onChange={(checked) => handleToggle(record.id, checked)}
        />
      ),
    },
    { title: "条件", dataIndex: "conditions", ellipsis: true },
    {
      title: "操作", width: 140,
      render: (_: unknown, record: GuardrailRule) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除此规则?" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-bold">安全护栏规则</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新建规则</Button>
      </div>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={rules}
        loading={loading}
        pagination={{ pageSize: 20 }}
      />

      <Modal
        title={editingRule ? "编辑规则" : "新建规则"}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        width={560}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="规则名称" rules={[{ required: true, message: "请输入名称" }]}>
            <Input placeholder="例如: 邮箱地址检测" />
          </Form.Item>
          <Form.Item name="stage" label="检测阶段" rules={[{ required: true }]}>
            <Select options={STAGE_OPTIONS} />
          </Form.Item>
          <Form.Item name="type" label="检测类型" rules={[{ required: true }]}>
            <Select options={TYPE_OPTIONS} />
          </Form.Item>
          <Form.Item name="action" label="处理动作" rules={[{ required: true }]}>
            <Select options={ACTION_OPTIONS} />
          </Form.Item>
          <Form.Item name="priority" label="优先级">
            <InputNumber min={0} max={100} className="w-full" />
          </Form.Item>
          <Form.Item name="conditions" label="条件表达式 (JSON)">
            <Input.TextArea rows={3} placeholder='{"pattern": "regex"}' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
