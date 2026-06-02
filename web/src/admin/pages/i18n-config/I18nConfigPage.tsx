import { useEffect, useState } from "react";
import { Card, Table, Switch, Button, Modal, Form, Input, Select, Space, Popconfirm, message, Tag } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";
import type { LanguageInfo } from "@/shared/api/types";

export default function I18nConfigPage() {
  const [languages, setLanguages] = useState<LanguageInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchLanguages = async () => {
    setLoading(true);
    try {
      const res = await adminService.listAllLanguages();
      setLanguages(res.data.data || []);
    } catch {
      message.error("加载语言列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchLanguages(); }, []);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await adminService.createLanguage(values);
      message.success("语言已添加");
      setModalOpen(false);
      form.resetFields();
      fetchLanguages();
    } catch {
      // validation error
    }
  };

  const handleSetDefault = async (locale: string) => {
    try {
      await adminService.setDefaultLanguage(locale);
      message.success(`已将 ${locale} 设为默认语言`);
      fetchLanguages();
    } catch {
      message.error("操作失败");
    }
  };

  const columns = [
    { title: "语种代码", dataIndex: "locale", width: 100 },
    { title: "名称", dataIndex: "name", width: 150 },
    {
      title: "状态", dataIndex: "enabled", width: 80,
      render: (e: boolean) => e === false ? <Tag color="default">禁用</Tag> : <Tag color="green">启用</Tag>,
    },
    {
      title: "默认", dataIndex: "is_default", width: 80,
      render: (d: boolean) => d ? <Tag color="blue">默认</Tag> : null,
    },
    {
      title: "操作", width: 160,
      render: (_: unknown, r: LanguageInfo) => (
        <Space>
          {!r.is_default && (
            <Popconfirm title={`设 ${r.locale} 为默认语言?`} onConfirm={() => handleSetDefault(r.locale)}>
              <Button size="small">设为默认</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">多语言配置</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchLanguages}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加语言</Button>
        </Space>
      </div>
      <Card>
        <Table rowKey="locale" columns={columns} dataSource={languages} loading={loading} pagination={false} />
      </Card>

      <Modal title="添加语言" open={modalOpen} onOk={handleCreate} onCancel={() => setModalOpen(false)} width={400} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="locale" label="语种代码" rules={[{ required: true }]}>
            <Input placeholder="fr" />
          </Form.Item>
          <Form.Item name="name" label="显示名称" rules={[{ required: true }]}>
            <Input placeholder="Français" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
