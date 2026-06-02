import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, Modal, Form, Input, Space, Popconfirm, message } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";

interface BlacklistItem {
  id: number;
  ip: string;
  reason: string;
  created_at?: string;
}

export default function BlacklistPage() {
  const [items, setItems] = useState<BlacklistItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await adminService.listBlacklist();
      setItems((res.data.data as BlacklistItem[]) || []);
    } catch {
      message.error("加载黑名单失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchList(); }, []);

  const handleAdd = async () => {
    try {
      const values = await form.validateFields();
      await adminService.addBlacklist(values);
      message.success("已添加");
      setModalOpen(false);
      form.resetFields();
      fetchList();
    } catch {
      // validation error
    }
  };

  const handleRemove = async (id: number) => {
    try {
      await adminService.removeBlacklist(id);
      message.success("已移除");
      fetchList();
    } catch {
      message.error("移除失败");
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "IP 地址", dataIndex: "ip", width: 180 },
    { title: "原因", dataIndex: "reason", width: 200 },
    { title: "添加时间", dataIndex: "created_at", width: 170, render: (v: string) => v ? new Date(v).toLocaleString() : "-" },
    {
      title: "操作", width: 80,
      render: (_: unknown, record: BlacklistItem) => (
        <Popconfirm title="确定移除此 IP?" onConfirm={() => handleRemove(record.id)}>
          <Button size="small" danger>移除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">黑名单管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchList}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加 IP</Button>
        </Space>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ pageSize: 20 }} />
      </Card>

      <Modal title="添加黑名单" open={modalOpen} onOk={handleAdd} onCancel={() => setModalOpen(false)} width={400} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="ip" label="IP 地址" rules={[{ required: true, message: "请输入 IP" }]}>
            <Input placeholder="192.168.1.100" />
          </Form.Item>
          <Form.Item name="reason" label="加入原因" rules={[{ required: true }]}>
            <Input placeholder="批量注册" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
