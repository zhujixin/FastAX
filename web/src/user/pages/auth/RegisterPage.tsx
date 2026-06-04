import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, Input, Button, App } from "antd";
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  SafetyCertificateOutlined,
  RocketOutlined,
} from "@ant-design/icons";
import { useNavigate, Link } from "react-router-dom";
import { authService } from "@/shared/api";
import { registerSchema, type RegisterFormData } from "@/shared/utils/validations";

export default function RegisterPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [sending, setSending] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const { control, handleSubmit, getValues, formState: { errors, isSubmitting } } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: { username: "", email: "", phone: "", password: "", verify_code: "" },
  });

  const sendVerifyCode = async () => {
    const { email } = getValues();
    if (!email) {
      message.warning("请先输入邮箱");
      return;
    }
    setSending(true);
    try {
      await authService.sendCode(email);
      message.success("验证码已发送至邮箱");
      let count = 60;
      setCountdown(count);
      const timer = setInterval(() => {
        count--;
        setCountdown(count);
        if (count <= 0) clearInterval(timer);
      }, 1000);
    } catch (err: any) {
      message.error(err.response?.data?.message || "发送失败");
    } finally {
      setSending(false);
    }
  };

  const onFinish = async (values: RegisterFormData) => {
    try {
      await authService.register({
        username: values.username,
        email: values.email || undefined,
        phone: values.phone || undefined,
        password: values.password,
        verify_code: values.verify_code,
      });
      message.success("注册成功，请登录");
      navigate("/login");
    } catch (err: any) {
      message.error(err.response?.data?.message || "注册失败");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden">
      {/* ── 背景渐变 ── */}
      <div className="absolute inset-0 bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50" />

      {/* ── 装饰元素 ── */}
      <div className="absolute top-0 left-0 w-full h-full">
        <div className="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] rounded-full bg-gradient-to-br from-blue-100/60 to-indigo-100/40 blur-3xl" />
        <div className="absolute bottom-[-15%] left-[-10%] w-[600px] h-[600px] rounded-full bg-gradient-to-tr from-violet-100/50 to-sky-100/30 blur-3xl" />

        <div className="absolute top-[20%] left-[15%] w-2 h-2 rounded-full bg-blue-300/40" />
        <div className="absolute top-[30%] right-[20%] w-3 h-3 rounded-full bg-indigo-300/30" />
        <div className="absolute bottom-[25%] left-[25%] w-2.5 h-2.5 rounded-full bg-violet-300/35" />
        <div className="absolute top-[60%] right-[15%] w-2 h-2 rounded-full bg-sky-300/40" />

        <div
          className="absolute inset-0 opacity-[0.015]"
          style={{
            backgroundImage: `
              linear-gradient(rgba(99, 102, 241, 0.5) 1px, transparent 1px),
              linear-gradient(90deg, rgba(99, 102, 241, 0.5) 1px, transparent 1px)
            `,
            backgroundSize: '60px 60px',
          }}
        />
      </div>

      {/* ── 注册卡片 ── */}
      <div className="relative z-10 w-full max-w-[440px] mx-4">
        {/* Logo */}
        <div className="text-center mb-8">
          <div
            className="w-14 h-14 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg shadow-blue-500/20 cursor-pointer hover:scale-105 transition-transform"
            onClick={() => navigate("/")}
          >
            <RocketOutlined className="text-white text-2xl" />
          </div>
          <h1 className="text-2xl font-bold text-gray-800">FastAX</h1>
          <p className="text-gray-400 text-sm mt-1">Token 代理与 AI 模型调度平台</p>
        </div>

        {/* 表单卡片 */}
        <div className="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl shadow-gray-200/50 border border-white/60 p-8">
          <h2 className="text-xl font-bold text-gray-800 mb-1">创建账号</h2>
          <p className="text-gray-400 text-sm mb-6">注册即享 100 万免费 Token</p>

          <form onSubmit={handleSubmit(onFinish)} className="space-y-4">
            {/* 用户名 */}
            <Form.Item className="!mb-0" validateStatus={errors.username ? "error" : ""} help={errors.username?.message}>
              <Controller
                name="username"
                control={control}
                render={({ field }) => (
                  <Input
                    {...field}
                    size="large"
                    prefix={<UserOutlined className="text-gray-400" />}
                    placeholder="用户名"
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                  />
                )}
              />
            </Form.Item>

            {/* 邮箱 */}
            <Form.Item className="!mb-0" validateStatus={errors.email ? "error" : ""} help={errors.email?.message}>
              <Controller
                name="email"
                control={control}
                render={({ field }) => (
                  <Input
                    {...field}
                    size="large"
                    prefix={<MailOutlined className="text-gray-400" />}
                    placeholder="邮箱（必填）"
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                  />
                )}
              />
            </Form.Item>

            {/* 手机号 */}
            <Form.Item className="!mb-0" validateStatus={errors.phone ? "error" : ""} help={errors.phone?.message}>
              <Controller
                name="phone"
                control={control}
                render={({ field }) => (
                  <Input
                    {...field}
                    size="large"
                    prefix={<PhoneOutlined className="text-gray-400" />}
                    placeholder="手机号（选填）"
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                  />
                )}
              />
            </Form.Item>

            {/* 密码 */}
            <Form.Item className="!mb-0" validateStatus={errors.password ? "error" : ""} help={errors.password?.message}>
              <Controller
                name="password"
                control={control}
                render={({ field }) => (
                  <Input.Password
                    {...field}
                    size="large"
                    prefix={<LockOutlined className="text-gray-400" />}
                    placeholder="密码（至少6位）"
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                  />
                )}
              />
            </Form.Item>

            {/* 验证码 */}
            <Form.Item className="!mb-0" validateStatus={errors.verify_code ? "error" : ""} help={errors.verify_code?.message}>
              <Controller
                name="verify_code"
                control={control}
                render={({ field }) => (
                  <Input
                    {...field}
                    size="large"
                    prefix={<SafetyCertificateOutlined className="text-gray-400" />}
                    placeholder="验证码"
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                    suffix={
                      <Button
                        type="link"
                        size="small"
                        loading={sending}
                        disabled={countdown > 0}
                        onClick={sendVerifyCode}
                        className="!px-0 !text-blue-500 hover:!text-blue-600"
                      >
                        {countdown > 0 ? `${countdown}s` : "获取验证码"}
                      </Button>
                    }
                  />
                )}
              />
            </Form.Item>

            <Button
              type="primary"
              htmlType="submit"
              block
              size="large"
              loading={isSubmitting}
              className="h-12 !rounded-xl !text-base !font-medium !bg-gradient-to-r !from-blue-500 !to-indigo-600 hover:!from-blue-600 hover:!to-indigo-700 !border-none shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 transition-all"
            >
              注册
            </Button>
          </form>

          {/* ── 底部链接 ── */}
          <div className="text-center mt-6 text-sm">
            <span className="text-gray-400">已有账号？</span>
            <Link to="/login" className="text-blue-500 hover:text-blue-600 font-medium ml-1 transition-colors">
              去登录
            </Link>
          </div>
        </div>

        {/* 协议 */}
        <p className="text-center mt-6 text-xs text-gray-400">
          注册即表示同意{' '}
          <a href="#" className="text-blue-500 hover:text-blue-600 transition-colors">服务协议</a>
          {' '}和{' '}
          <a href="#" className="text-blue-500 hover:text-blue-600 transition-colors">隐私政策</a>
        </p>
      </div>
    </div>
  );
}
