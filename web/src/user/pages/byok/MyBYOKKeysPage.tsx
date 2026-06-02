import { useEffect, useState, useCallback } from "react";
import { Table, Button, Modal, Form, Input, Select, Space, Tag, Popconfirm, message, Tooltip } from "antd";
import { PlusOutlined, DeleteOutlined, KeyOutlined, CopyOutlined, EyeOutlined, EyeInvisibleOutlined } from "@ant-design/icons";
import { tokenService } from "@/shared/api/tokens";
import type { BYOKKey } from "@/shared/api/types";

const PROVIDER_OPTIONS = [
  { label: "OpenAI", value: "openai" },
  { label: "Anthropic (Claude)", value: "anthropic" },
  { label: "Google (Gemini)", value: "gemini" },
  { label: "DeepSeek", value: "deepseek" },
  { label: "Qwen (通义千问)", value: "qwen" },
  { label: "GLM (智谱)", value: "glm" },
  { label: "Moonshot (月之暗面)", value: "moonshot" },
  { label: "其他", value: "other" },
];

export default function MyBYOKKeysPage() {
  const [keys, setKeys] = useState<BYOKKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set());

  const fetchKeys = useCallback(async () => {
    setLoading(true);
    try {
      const res = await tokenService.listBYOKKeys();
      setKeys(res.data.data || []);
    } catch {
      message.error("加载 BYOK 密钥失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchKeys(); }, [fetchKeys]);

  const handleAdd = async () => {
    try {
      const values = await form.validateFields();
      await tokenService.addBYOKKey(values);
      message.success("密钥已添加");
      setModalOpen(false);
      form.resetFields();
      fetchKeys();
    } catch {
      // validation error
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await tokenService.deleteBYOKKey(id);
      message.success("密钥已删除");
      fetchKeys();
    } catch {
      message.error("删除失败");
    }
  };

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await tokenService.setBYOKKeyStatus(id, enabled);
      message.success(enabled ? "密钥已启用" : "密钥已禁用");
      fetchKeys();
    } catch {
      message.error("操作失败");
    }
  };

  const providerLabel = (v: string) => {
    const found = PROVIDER_OPTIONS.find((o) => o.value === v);
    return found ? found.label : v;
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    {
      title: "供应商", dataIndex: "provider", width: 160,
      render: (v: string) => <Tag color="blue">{providerLabel(v)}</Tag>,
    },
    {
      title: "Key", dataIndex: "key_prefix", width: 280,
      render: (_: string, record: BYOKKey) => (
        <Space>
          <code className="text-xs bg-gray-100 px-2 py-0.5 rounded">
            {record.key_prefix}********
          </code>
        </Space>
      ),
    },
    {
      title: "状态", dataIndex: "status", width: 100,
      render: (v: number, record: BYOKKey) => (
        <Tag color={v === 1 ? "green" : "red"}>
          {v === 1 ? "启用" : "禁用"}
        </Tag>
      ),
    },
    {
      title: "已用 Token", dataIndex: "total_used", width: 120,
      render: (v: number) => v?.toLocaleString() || "0",
    },
    {
      title: "添加时间", dataIndex: "created_at", width: 170,
      render: (v: number) => new Date(v * 1000).toLocaleString(),
    },
    {
      title: "操作", width: 140,
      render: (_: unknown, record: BYOKKey) => (
        <Space>
          <Button
            type="link"
            size="small"
            danger={record.status === 1}
            onClick={() => handleToggle(record.id, record.status !== 1)}
          >
            {record.status === 1 ? "禁用" : "启用"}
          </Button>
          <Popconfirm title="确定删除此密钥?" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="max-w-5xl mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-xl font-bold">我的 API Key (BYOK)</h2>
          <p className="text-gray-500 text-sm mt-1">
            添加您自己的供应商 API Key，优先使用自带 Key 消费，不足时自动回退平台 Token
          </p>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          添加 Key
        </Button>
      </div>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={keys}
        loading={loading}
        pagination={{ pageSize: 15 }}
        locale={{ emptyText: "暂无 Key，点击上方按钮添加" }}
      />

      <Modal
        title="添加 API Key"
        open={modalOpen}
        onOk={handleAdd}
        onCancel={() => { setModalOpen(false); form.resetFields(); }}
        width={480}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="provider" label="供应商" rules={[{ required: true, message: "请选择供应商" }]}>
            <Select options={PROVIDER_OPTIONS} placeholder="选择 AI 供应商" showSearch />
          </Form.Item>
          <Form.Item
            name="key"
            label="API Key"
            rules={[{ required: true, message: "请输入 API Key" }]}
            extra="Key 将使用 AES-256-GCM 加密存储，确保安全"
          >
            <Input.Password
              placeholder="sk-..."
              iconRender={(visible) => (visible ? <EyeOutlined /> : <EyeInvisibleOutlined />)}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
