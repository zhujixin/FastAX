import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, Space, Modal, Form, Input, InputNumber, Select, message } from "antd";
import { PlusOutlined, EditOutlined, ReloadOutlined } from "@ant-design/icons";
import { tokenService } from "@/shared/api/tokens";
import { adminService } from "@/shared/api/admin";

interface ProductItem {
  id: number;
  name: string;
  model: string;
  type: string;
  unit: string;
  price: string;
  stock: number;
  status: string;
  supplier?: string;
}

export default function TokenManagePage() {
  const [products, setProducts] = useState<ProductItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchProducts = async () => {
    setLoading(true);
    try {
      const res = await tokenService.getProducts();
      setProducts((res.data.data as unknown as ProductItem[]) || []);
    } catch {
      message.error("加载商品列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchProducts(); }, []);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await adminService.createProduct(values);
      message.success("商品已创建");
      setModalOpen(false);
      form.resetFields();
      fetchProducts();
    } catch {
      // validation error
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "名称", dataIndex: "name", width: 180 },
    { title: "模型", dataIndex: "model", width: 140, render: (v: string) => v || "-" },
    { title: "类型", dataIndex: "type", width: 80 },
    { title: "单位", dataIndex: "unit", width: 80 },
    { title: "价格", dataIndex: "price", width: 100, render: (v: string) => `¥${parseFloat(v || "0").toFixed(4)}` },
    { title: "库存", dataIndex: "stock", width: 100, render: (v: number) => v?.toLocaleString() || "0" },
    {
      title: "状态", dataIndex: "status", width: 80,
      render: (s: string) => <Tag color={s === "active" ? "green" : "red"}>{s || "active"}</Tag>,
    },
    {
      title: "操作", width: 100,
      render: (_: unknown, record: ProductItem) => (
        <Button size="small" icon={<EditOutlined />} onClick={() => message.info("编辑功能开发中")}>编辑</Button>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">Token 商品管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchProducts}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>新增商品</Button>
        </Space>
      </div>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={products}
          loading={loading}
          pagination={{ pageSize: 20 }}
        />
      </Card>

      <Modal title="新增商品" open={modalOpen} onOk={handleCreate} onCancel={() => setModalOpen(false)} width={480} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="商品名称" rules={[{ required: true }]}>
            <Input placeholder="GPT-4 Token 包" />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ label: "Token", value: "token" }, { label: "Subscription", value: "subscription" }]} />
          </Form.Item>
          <Form.Item name="model" label="模型标识">
            <Input placeholder="gpt-4" />
          </Form.Item>
          <Form.Item name="unit" label="计费单位" rules={[{ required: true }]}>
            <Input placeholder="1K tokens" />
          </Form.Item>
          <Form.Item name="price" label="单价" rules={[{ required: true }]}>
            <InputNumber min={0} step={0.01} className="w-full" placeholder="0.15" />
          </Form.Item>
          <Form.Item name="stock" label="库存 (Token 数量)">
            <InputNumber min={0} className="w-full" placeholder="10000000" />
          </Form.Item>
          <Form.Item name="currency" label="币种" initialValue="CNY">
            <Select options={[{ label: "CNY", value: "CNY" }, { label: "USD", value: "USD" }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
