import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, Input, Button, Card, App } from "antd";
import { UserOutlined, LockOutlined, MailOutlined, PhoneOutlined } from "@ant-design/icons";
import { useNavigate, Link } from "react-router-dom";
import { authService } from "@/shared/api";
import { registerSchema, type RegisterFormData } from "@/shared/utils/validations";

export default function RegisterPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();

  const { control, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: { username: "", email: "", phone: "", password: "" },
  });

  const onFinish = async (values: RegisterFormData) => {
    try {
      await authService.register({
        username: values.username,
        email: values.email || undefined,
        phone: values.phone || undefined,
        password: values.password,
      });
      message.success("注册成功，请登录");
      navigate("/login");
    } catch (err: any) {
      message.error(err.response?.data?.message || "注册失败");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <Card className="w-96 shadow-lg">
        <h2 className="text-center text-2xl font-bold mb-6">FastAX 注册</h2>
        <form onSubmit={handleSubmit(onFinish)}>
          <Form.Item validateStatus={errors.username ? "error" : ""} help={errors.username?.message}>
            <Controller name="username" control={control} render={({ field }) => (
              <Input {...field} size="large" prefix={<UserOutlined />} placeholder="用户名" />
            )} />
          </Form.Item>
          <Form.Item validateStatus={errors.email ? "error" : ""} help={errors.email?.message}>
            <Controller name="email" control={control} render={({ field }) => (
              <Input {...field} size="large" prefix={<MailOutlined />} placeholder="邮箱（选填）" />
            )} />
          </Form.Item>
          <Form.Item validateStatus={errors.phone ? "error" : ""} help={errors.phone?.message}>
            <Controller name="phone" control={control} render={({ field }) => (
              <Input {...field} size="large" prefix={<PhoneOutlined />} placeholder="手机号（选填）" />
            )} />
          </Form.Item>
          <Form.Item validateStatus={errors.password ? "error" : ""} help={errors.password?.message}>
            <Controller name="password" control={control} render={({ field }) => (
              <Input.Password {...field} size="large" prefix={<LockOutlined />} placeholder="密码（至少6位）" />
            )} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={isSubmitting}>
              注册
            </Button>
          </Form.Item>
        </form>
        <div className="text-center">
          <Link to="/login">已有账号？去登录</Link>
        </div>
      </Card>
    </div>
  );
}
