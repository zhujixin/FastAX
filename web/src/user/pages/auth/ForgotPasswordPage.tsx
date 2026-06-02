import { Form, Input, Button, Card, App } from "antd";
import { MailOutlined } from "@ant-design/icons";
import { Link } from "react-router-dom";
import api from "@/shared/utils/axios";

export default function ForgotPasswordPage() {
  const [form] = Form.useForm();
  const { message } = App.useApp();

  const onFinish = async (values: { account: string }) => {
    try {
      await api.post("/auth/reset-password", values);
      message.success("重置邮件已发送");
    } catch (err: any) {
      message.error(err.response?.data?.message || "发送失败");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-96 shadow-lg">
        <h2 className="text-center text-2xl font-bold mb-6">忘记密码</h2>
        <Form form={form} onFinish={onFinish} size="large">
          <Form.Item name="account" rules={[{ required: true, message: "请输入手机号或邮箱" }]}>
            <Input prefix={<MailOutlined />} placeholder="手机号/邮箱" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>发送重置链接</Button>
          </Form.Item>
        </Form>
        <div className="text-center">
          <Link to="/login">返回登录</Link>
        </div>
      </Card>
    </div>
  );
}
