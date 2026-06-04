import { useState, useEffect } from "react";
import { Card, Table, Button, Modal, Form, Input, Select, InputNumber, Tag, App, Spin, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";

interface Supplier {
  id: number; name: string; code: string; description: string;
  api_base_url: string; region: string; status: number;
  priority: number; weight: number;
}

export default function SupplierManagePage() {
  const { message: msg } = App.useApp();
  const [loading, setLoading] = useState(true);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Supplier | null>(null);
  const [form] = Form.useForm();

  const fetchSuppliers = () => {
    api.get("/admin/suppliers").then((r) => {
      setSuppliers(r.data?.data || []);
    }).catch(() => msg.error("加载失败")).finally(() => setLoading(false));
  };

  useEffect(() => { fetchSuppliers(); }, []);

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      if (editing) {
        await api.put(`/admin/suppliers/${editing.id}`, values);
        msg.success("已更新");
      } else {
        await api.post("/admin/suppliers", values);
        msg.success("已创建");
      }
      setModalOpen(false); setEditing(null); form.resetFields();
      fetchSuppliers();
    } catch { msg.error("操作失败"); }
  };

  const setStatus = async (id: number, status: number) => {
    await api.put(`/admin/suppliers/${id}/status`, { status });
    msg.success("状态已更新");
    fetchSuppliers();
  };

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">供应商管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true); }}>新建供应商</Button>
      </div>
      <Card>
        <Table dataSource={suppliers} rowKey="id" columns={[
          { title: "名称", dataIndex: "name" },
          { title: "代码", dataIndex: "code" },
          { title: "区域", dataIndex: "region", render: (r: string) => <Tag color={r === "domestic" ? "blue" : "green"}>{r === "domestic" ? "国内" : "海外"}</Tag> },
          { title: "优先级", dataIndex: "priority" },
          { title: "权重", dataIndex: "weight" },
          { title: "状态", dataIndex: "status", render: (s: number) => <Tag color={s === 1 ? "green" : "red"}>{s === 1 ? "启用" : "禁用"}</Tag> },
          { title: "操作", render: (_: unknown, r: Supplier) => (<>
            <Button size="small" onClick={() => { setEditing(r); form.setFieldsValue(r); setModalOpen(true); }}>编辑</Button>
            <Button size="small" className="ml-2" onClick={() => setStatus(r.id, r.status === 1 ? 0 : 1)}>{r.status === 1 ? "禁用" : "启用"}</Button>
          </>)}],
        ]} />
      </Card>
      <Modal title={editing ? "编辑供应商" : "新建供应商"} open={modalOpen} onCancel={() => { setModalOpen(false); setEditing(null); }} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="code" label="代码" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="api_base_url" label="API URL" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="api_key_encrypted" label="API Key" rules={[{ required: !editing }]}><Input.Password /></Form.Item>
          <Form.Item name="region" label="区域"><Select options={[{ value: "overseas", label: "海外" }, { value: "domestic", label: "国内" }]} /></Form.Item>
          <Form.Item name="priority" label="优先级"><InputNumber min={0} className="w-full" /></Form.Item>
          <Form.Item name="weight" label="权重"><InputNumber min={1} max={100} className="w-full" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
