import { useState, useEffect } from "react";
import { Card, Form, Input, InputNumber, Button, App, Spin, message } from "antd";
import api from "@/shared/utils/axios";

export default function SystemPage() {
  const { message: msg } = App.useApp();
  const [loading, setLoading] = useState(true);
  const [form] = Form.useForm();

  useEffect(() => {
    api.get("/admin/system/config")
      .then((res) => {
        const items = res.data?.data?.items;
        if (Array.isArray(items)) {
          const config: Record<string, string> = {};
          items.forEach((item: { config_key: string; config_value: string }) => {
            config[item.config_key] = item.config_value;
          });
          form.setFieldsValue({
            site_name: config.site_name || "FastAX",
            max_retries: parseInt(config.max_retries || "3", 10),
            log_level: config.log_level || "info",
            enable_signup: config.enable_signup === "true",
          });
        }
      })
      .catch(() => msg.error("加载系统配置失败"))
      .finally(() => setLoading(false));
  }, [form]);

  const handleSave = async (values: Record<string, unknown>) => {
    try {
      await api.put("/admin/system/config", {
        site_name: String(values.site_name || "FastAX"),
        max_retries: String(values.max_retries || 3),
        log_level: String(values.log_level || "info"),
        enable_signup: String(values.enable_signup ?? true),
      });
      msg.success("配置已保存");
    } catch { message.error("保存失败"); }
  };

  if (loading) return <div className="flex justify-center py-20"><Spin size="large" /></div>;

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">系统设置</h2>
      <Card title="基本配置">
        <Form form={form} layout="vertical" onFinish={handleSave} className="max-w-lg">
          <Form.Item label="站点名称" name="site_name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="最大重试次数" name="max_retries">
            <InputNumber min={1} max={10} className="w-full" />
          </Form.Item>
          <Form.Item label="日志级别" name="log_level">
            <Input placeholder="info / debug / warn / error" />
          </Form.Item>
          <Form.Item label="启用注册" name="enable_signup" valuePropName="checked">
            <Input type="checkbox" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">保存配置</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
