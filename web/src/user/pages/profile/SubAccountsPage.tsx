import { useEffect, useState } from "react";
import { Card, Table, Button, Tag, Modal, Form, Input, InputNumber, Space, Popconfirm, message } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { enterpriseService } from "@/shared/api/enterprise";

interface SubAccount {
  id: number;
  name: string;
  quota: number;
  used: number;
  status: string;
}

export default function SubAccountsPage() {
  const [accounts, setAccounts] = useState<SubAccount[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchAccounts = async () => {
    setLoading(true);
    try {
      const res = await enterpriseService.listSubAccounts();
      setAccounts(res.data.data || []);
    } catch {
      message.error("加载子账号失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchAccounts(); }, []);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await enterpriseService.createSubAccount(values);
      message.success("子账号已创建");
      setModalOpen(false);
      form.resetFields();
      fetchAccounts();
    } catch {
      // validation error
    }
  };

  const handleSetStatus = async (id: number, status: string) => {
    try {
      await enterpriseService.setSubAccountStatus(id, status);
      message.success("状态已更新");
      fetchAccounts();
    } catch {
      message.error("操作失败");
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "名称", dataIndex: "name", width: 150 },
    { title: "配额", dataIndex: "quota", width: 120, render: (v: number) => v?.toLocaleString() || "0" },
    { title: "已用", dataIndex: "used", width: 120, render: (v: number) => v?.toLocaleString() || "0" },
    {
      title: "使用率", width: 100,
      render: (_: unknown, r: SubAccount) => {
        const pct = r.quota > 0 ? Math.round((r.used / r.quota) * 100) : 0;
        return <Tag color={pct > 90 ? "red" : pct > 70 ? "orange" : "green"}>{pct}%</Tag>;
      },
    },
    {
      title: "状态", dataIndex: "status", width: 80,
      render: (s: string) => <Tag color={s === "active" ? "green" : "red"}>{s}</Tag>,
    },
    {
      title: "操作", width: 120,
      render: (_: unknown, r: SubAccount) => (
        <Space>
          {r.status === "active" ? (
            <Popconfirm title="确定冻结?" onConfirm={() => handleSetStatus(r.id, "frozen")}>
              <Button size="small" danger>冻结</Button>
            </Popconfirm>
          ) : (
            <Button size="small" type="primary" onClick={() => handleSetStatus(r.id, "active")}>启用</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="max-w-4xl mx-auto p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">子账号管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchAccounts}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>创建子账号</Button>
        </Space>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={accounts} loading={loading} pagination={{ pageSize: 20 }} />
      </Card>

      <Modal title="创建子账号" open={modalOpen} onOk={handleCreate} onCancel={() => setModalOpen(false)} width={400} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="dev-team" />
          </Form.Item>
          <Form.Item name="quota" label="Token 配额" rules={[{ required: true, type: "number", min: 1 }]}>
            <InputNumber min={1} className="w-full" placeholder="100000" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
