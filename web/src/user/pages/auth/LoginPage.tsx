import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, Input, Button, App } from "antd";
import { WechatOutlined, GoogleOutlined, GithubOutlined, RocketOutlined } from "@ant-design/icons";
import { useNavigate, Link } from "react-router-dom";
import { useAuthStore } from "@/shared/stores/authStore";
import { authService } from "@/shared/api";
import { loginSchema, type LoginFormData } from "@/shared/utils/validations";

type LoginMode = "phone" | "email";

export default function LoginPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { setAuth } = useAuthStore();
  const [mode, setMode] = useState<LoginMode>("phone");
  const [sending, setSending] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const { control, handleSubmit, getValues, formState: { errors, isSubmitting } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { account: "", password: "" },
  });

  const sendCode = async () => {
    const account = getValues("account");
    if (!account) {
      message.warning(mode === "phone" ? "请先输入手机号" : "请先输入邮箱");
      return;
    }
    setSending(true);
    try {
      await authService.sendCode(account);
      message.success("验证码已发送");
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

  const oauthLogin = (provider: string) => {
    window.location.href = authService.getOAuthUrl(provider);
  };

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden">
      {/* ── 背景渐变 ── */}
      <div className="absolute inset-0 bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50" />

      {/* ── 装饰元素 ── */}
      <div className="absolute top-0 left-0 w-full h-full">
        {/* 大圆 */}
        <div className="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] rounded-full bg-gradient-to-br from-blue-100/60 to-indigo-100/40 blur-3xl" />
        <div className="absolute bottom-[-15%] left-[-10%] w-[600px] h-[600px] rounded-full bg-gradient-to-tr from-violet-100/50 to-sky-100/30 blur-3xl" />

        {/* 小圆点装饰 */}
        <div className="absolute top-[20%] left-[15%] w-2 h-2 rounded-full bg-blue-300/40" />
        <div className="absolute top-[30%] right-[20%] w-3 h-3 rounded-full bg-indigo-300/30" />
        <div className="absolute bottom-[25%] left-[25%] w-2.5 h-2.5 rounded-full bg-violet-300/35" />
        <div className="absolute top-[60%] right-[15%] w-2 h-2 rounded-full bg-sky-300/40" />

        {/* 网格线 */}
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

      {/* ── 登录卡片 ── */}
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
          <h2 className="text-xl font-bold text-gray-800 mb-1">欢迎回来</h2>
          <p className="text-gray-400 text-sm mb-6">登录您的账户以继续</p>

          {/* ── 登录方式切换 ── */}
          <div className="flex bg-gray-100/80 rounded-xl p-1 mb-6">
            {([
              { key: "phone", label: "手机号登录" },
              { key: "email", label: "邮箱登录" },
            ] as const).map((tab) => (
              <button
                key={tab.key}
                type="button"
                onClick={() => setMode(tab.key)}
                className={`flex-1 py-2.5 text-sm font-medium rounded-[10px] transition-all duration-200 ${
                  mode === tab.key
                    ? "bg-white text-blue-600 shadow-sm"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit(onFinish)} className="space-y-4">
            {/* 账号输入 */}
            <Form.Item className="!mb-0" validateStatus={errors.account ? "error" : ""} help={errors.account?.message}>
              <Controller
                name="account"
                control={control}
                render={({ field }) => (
                  <Input
                    {...field}
                    size="large"
                    placeholder={mode === "phone" ? "请输入手机号" : "请输入邮箱"}
                    className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                  />
                )}
              />
            </Form.Item>

            {/* 手机号模式：验证码登录 */}
            {mode === "phone" ? (
              <div className="flex gap-3">
                <Controller
                  name="password"
                  control={control}
                  render={({ field }) => (
                    <Input
                      {...field}
                      size="large"
                      placeholder="验证码"
                      className="h-12 rounded-xl text-base flex-1 !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                    />
                  )}
                />
                <Button
                  type="default"
                  size="large"
                  loading={sending}
                  disabled={countdown > 0}
                  onClick={sendCode}
                  className="h-12 rounded-xl px-5 text-blue-600 border-blue-200 hover:border-blue-400 font-medium shrink-0"
                >
                  {countdown > 0 ? `${countdown}s` : "获取验证码"}
                </Button>
              </div>
            ) : (
              /* 邮箱模式：密码登录 */
              <Form.Item className="!mb-0" validateStatus={errors.password ? "error" : ""} help={errors.password?.message}>
                <Controller
                  name="password"
                  control={control}
                  render={({ field }) => (
                    <Input.Password
                      {...field}
                      size="large"
                      placeholder="请输入密码"
                      className="h-12 rounded-xl text-base !bg-gray-50/50 !border-gray-200 hover:!border-blue-400 focus:!border-blue-500"
                    />
                  )}
                />
              </Form.Item>
            )}

            <Button
              type="primary"
              htmlType="submit"
              block
              size="large"
              loading={isSubmitting}
              className="h-12 !rounded-xl !text-base !font-medium !bg-gradient-to-r !from-blue-500 !to-indigo-600 hover:!from-blue-600 hover:!to-indigo-700 !border-none shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 transition-all"
            >
              登录
            </Button>
          </form>

          {/* ── 分隔线 ── */}
          <div className="flex items-center gap-4 my-6">
            <div className="flex-1 h-px bg-gray-200/80" />
            <span className="text-xs text-gray-300 shrink-0">其他登录方式</span>
            <div className="flex-1 h-px bg-gray-200/80" />
          </div>

          {/* ── SSO 按钮 ── */}
          <div className="flex justify-center gap-4">
            <button
              type="button"
              onClick={() => oauthLogin("wechat")}
              className="w-12 h-12 rounded-xl bg-white border border-gray-200 flex items-center justify-center text-[#2aa745] hover:bg-[#f0fdf4] hover:border-[#86efac] transition-all shadow-sm hover:shadow-md"
              title="微信登录"
            >
              <WechatOutlined className="text-xl" />
            </button>
            <button
              type="button"
              onClick={() => oauthLogin("google")}
              className="w-12 h-12 rounded-xl bg-white border border-gray-200 flex items-center justify-center text-[#ea4335] hover:bg-[#fef2f2] hover:border-[#fca5a5] transition-all shadow-sm hover:shadow-md"
              title="Google 登录"
            >
              <GoogleOutlined className="text-xl" />
            </button>
            <button
              type="button"
              onClick={() => oauthLogin("github")}
              className="w-12 h-12 rounded-xl bg-white border border-gray-200 flex items-center justify-center text-[#24292f] hover:bg-[#f6f8fa] hover:border-[#d1d5db] transition-all shadow-sm hover:shadow-md"
              title="GitHub 登录"
            >
              <GithubOutlined className="text-xl" />
            </button>
          </div>

          {/* ── 底部链接 ── */}
          <div className="text-center mt-6 text-sm text-gray-400 space-x-4">
            <Link to="/register" className="text-blue-500 hover:text-blue-600 font-medium transition-colors">注册账号</Link>
            <span className="text-gray-300">·</span>
            <Link to="/forgot-password" className="text-gray-400 hover:text-blue-500 transition-colors">忘记密码</Link>
          </div>
        </div>

        {/* 协议 */}
        <p className="text-center mt-6 text-xs text-gray-400">
          登录即表示同意{' '}
          <a href="#" className="text-blue-500 hover:text-blue-600 transition-colors">服务协议</a>
          {' '}和{' '}
          <a href="#" className="text-blue-500 hover:text-blue-600 transition-colors">隐私政策</a>
        </p>
      </div>
    </div>
  );
}
