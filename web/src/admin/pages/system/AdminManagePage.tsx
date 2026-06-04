import { useState, useEffect } from "react";
import { Card, Table, Button, Modal, Form, Input, Select, Tag, App, Spin, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";

interface Admin {
  id: number; username: string; email: string; role: string; status: number; created_at: string;
}

export default function AdminManagePage() {
  const { message: msg } = App.useApp();
  const [loading, setLoading] = useState(true);
  const [admins, setAdmins] = useState<Admin[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchAdmins = () => {
    api.get("/admin/system/admins").then((r) => {
      setAdmins(r.data?.data || []);
    }).catch(() => msg.error("加载失败")).finally(() => setLoading(false));
  };

  useEffect(() => { fetchAdmins(); }, []);

  const handleCreate = async (values: Record<string, unknown>) => {
    try {
      await api.post("/admin/system/admins", values);
      msg.success("管理员已创建");
      setModalOpen(false); form.resetFields();
      fetchAdmins();
    } catch { message.error("创建失败"); }
  };

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">管理员管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setModalOpen(true); }}>添加管理员</Button>
      </div>
      <Card>
        <Table dataSource={admins} rowKey="id" columns={[
          { title: "用户名", dataIndex: "username" },
          { title: "邮箱", dataIndex: "email" },
          { title: "角色", dataIndex: "role", render: (r: string) => <Tag color={r === "super_admin" ? "red" : "blue"}>{r === "super_admin" ? "超级管理员" : "管理员"}</Tag> },
          { title: "状态", dataIndex: "status", render: (s: number) => <Tag color={s === 1 ? "green" : "red"}>{s === 1 ? "正常" : "冻结"}</Tag> },
          { title: "创建时间", dataIndex: "created_at", render: (v: string) => v?.slice(0, 10) },
        ]} />
      </Card>
      <Modal title="添加管理员" open={modalOpen} onCancel={() => setModalOpen(false)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}><Input.Password /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: "email" }]}><Input /></Form.Item>
          <Form.Item name="role" label="角色"><Select options={[{ value: "admin", label: "管理员" }, { value: "super_admin", label: "超级管理员" }]} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
