import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, Input, Button, Card, App } from "antd";
import { UserOutlined, LockOutlined } from "@ant-design/icons";
import { useNavigate, Link } from "react-router-dom";
import { useAuthStore } from "@/shared/stores/authStore";
import { authService } from "@/shared/api";
import { loginSchema, type LoginFormData } from "@/shared/utils/validations";

export default function LoginPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { setAuth } = useAuthStore();

  const { control, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { account: "", password: "" },
  });

  const onFinish = async (values: LoginFormData) => {
    try {
      const res = await authService.login(values);
      const { access_token, refresh_token, user } = res.data.data;
      setAuth(access_token, refresh_token, user);
      message.success("登录成功");
      navigate("/dashboard");
    } catch (err: any) {
      message.error(err.response?.data?.message || "登录失败");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <Card className="w-96 shadow-lg">
        <h2 className="text-center text-2xl font-bold mb-6">FastAX 登录</h2>
        <form onSubmit={handleSubmit(onFinish)}>
          <Form.Item
            validateStatus={errors.account ? "error" : ""}
            help={errors.account?.message}
          >
            <Controller
              name="account"
              control={control}
              render={({ field }) => (
                <Input {...field} size="large" prefix={<UserOutlined />} placeholder="手机号/邮箱" />
              )}
            />
          </Form.Item>
          <Form.Item
            validateStatus={errors.password ? "error" : ""}
            help={errors.password?.message}
          >
            <Controller
              name="password"
              control={control}
              render={({ field }) => (
                <Input.Password {...field} size="large" prefix={<LockOutlined />} placeholder="密码" />
              )}
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={isSubmitting}>
              登录
            </Button>
          </Form.Item>
        </form>
        <div className="text-center space-x-4">
          <Link to="/register">注册账号</Link>
          <Link to="/forgot-password">忘记密码</Link>
        </div>
      </Card>
    </div>
  );
}
