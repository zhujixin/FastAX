import { Card, Form, Input, Button, App } from "antd";

export default function SystemPage() {
  const { message } = App.useApp();

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">系统设置</h2>
      <Card title="基本配置">
        <Form layout="vertical" onFinish={() => message.success("配置已保存")}>
          <Form.Item label="站点名称" name="site_name" initialValue="FastAX">
            <Input />
          </Form.Item>
          <Form.Item label="限流 (次/分钟)" name="rate_limit" initialValue={60}>
            <Input type="number" />
          </Form.Item>
          <Form.Item label="JWT 过期时间 (小时)" name="jwt_expire" initialValue={24}>
            <Input type="number" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">保存配置</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
