import { useEffect, useState, useRef, useCallback } from "react";
import { Spin, Tag } from "antd";
import { useNavigate } from "react-router-dom";
import {
  ThunderboltOutlined,
  SafetyOutlined,
  GlobalOutlined,
  BarChartOutlined,
  ApiOutlined,
  ArrowRightOutlined,
  RocketOutlined,
  CodeOutlined,
  DollarOutlined,
  MenuOutlined,
  CloseOutlined,
} from "@ant-design/icons";
import { tokenService } from "@/shared/api/tokens";
import type { TokenProduct } from "@/shared/api/types";

/* ─── 常量 ─── */
const NAV_ITEMS = [
  { label: "产品", href: "#products" },
  { label: "API 文档", href: "/docs" },
  { label: "定价", href: "/pricing" },
  { label: "控制台", href: "/dashboard" },
];

const MODELS = [
  { name: "DeepSeek-V3", provider: "DeepSeek", desc: "高性能推理", color: "#3b82f6" },
  { name: "GPT-4o", provider: "OpenAI", desc: "多模态旗舰", color: "#10b981" },
  { name: "Claude 3.5", provider: "Anthropic", desc: "安全可控", color: "#8b5cf6" },
  { name: "Gemini 2.0", provider: "Google", desc: "多模态原生", color: "#f59e0b" },
  { name: "GLM-4", provider: "智谱 AI", desc: "中文优化", color: "#ef4444" },
  { name: "Qwen-Max", provider: "阿里云", desc: "长上下文", color: "#06b6d4" },
];

const FEATURES = [
  {
    icon: <ThunderboltOutlined />,
    title: "智能路由",
    desc: "优先级分组 + 权重随机算法，故障自动切换，确保请求永不中断。",
    color: "#3b82f6",
    bg: "bg-blue-50",
  },
  {
    icon: <ApiOutlined />,
    title: "多协议兼容",
    desc: "原生支持 OpenAI、Anthropic、Gemini 协议，一套 API 对接所有模型。",
    color: "#10b981",
    bg: "bg-emerald-50",
  },
  {
    icon: <SafetyOutlined />,
    title: "安全护栏",
    desc: "输入/输出双阶段检测，PII 脱敏、注入防护、敏感内容过滤。",
    color: "#f59e0b",
    bg: "bg-amber-50",
  },
  {
    icon: <BarChartOutlined />,
    title: "用量分析",
    desc: "实时仪表盘 + 多维度报表，按模型/时间查看 Token 消耗趋势。",
    color: "#8b5cf6",
    bg: "bg-violet-50",
  },
];

const DEMO_MESSAGES = [
  { role: "user", content: "解释一下量子计算的基本原理" },
  {
    role: "assistant",
    content: "量子计算利用量子力学原理（叠加、纠缠）进行计算。与经典比特（0或1）不同，量子比特可同时处于0和1的叠加态，使量子计算机能并行处理大量计算...",
  },
];

/* ─── 粒子背景 ─── */
function ParticleBg() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const resize = () => {
      canvas.width = canvas.offsetWidth * devicePixelRatio;
      canvas.height = canvas.offsetHeight * devicePixelRatio;
      ctx.scale(devicePixelRatio, devicePixelRatio);
    };
    resize();
    window.addEventListener("resize", resize);

    const particles = Array.from({ length: 30 }, () => ({
      x: Math.random() * canvas.offsetWidth,
      y: Math.random() * canvas.offsetHeight,
      r: Math.random() * 1.2 + 0.3,
      vx: (Math.random() - 0.5) * 0.15,
      vy: (Math.random() - 0.5) * 0.15,
      opacity: Math.random() * 0.15 + 0.05,
    }));

    let raf: number;
    const animate = () => {
      ctx.clearRect(0, 0, canvas.offsetWidth, canvas.offsetHeight);
      particles.forEach((p) => {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0) p.x = canvas.offsetWidth;
        if (p.x > canvas.offsetWidth) p.x = 0;
        if (p.y < 0) p.y = canvas.offsetHeight;
        if (p.y > canvas.offsetHeight) p.y = 0;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(99,102,241,${p.opacity})`;
        ctx.fill();
      });

      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 100) {
            ctx.beginPath();
            ctx.moveTo(particles[i].x, particles[i].y);
            ctx.lineTo(particles[j].x, particles[j].y);
            ctx.strokeStyle = `rgba(99,102,241,${0.02 * (1 - dist / 100)})`;
            ctx.stroke();
          }
        }
      }
      raf = requestAnimationFrame(animate);
    };
    animate();

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("resize", resize);
    };
  }, []);

  return <canvas ref={canvasRef} className="absolute inset-0 w-full h-full" />;
}

/* ─── 统计数字动画 ─── */
function AnimatedNumber({ target, suffix = "" }: { target: number; suffix?: string }) {
  const [count, setCount] = useState(0);
  useEffect(() => {
    if (target === 0) return;
    const duration = 1500;
    const step = Math.ceil(target / (duration / 16));
    let current = 0;
    const timer = setInterval(() => {
      current = Math.min(current + step, target);
      setCount(current);
      if (current >= target) clearInterval(timer);
    }, 16);
    return () => clearInterval(timer);
  }, [target]);

  return (
    <span>
      {count.toLocaleString()}
      {suffix}
    </span>
  );
}

/* ─── 模型卡片 ─── */
function ModelCard({ model, onClick }: { model: (typeof MODELS)[0]; onClick: () => void }) {
  return (
    <div
      className="flex-shrink-0 w-48 p-4 rounded-xl bg-white border border-gray-100 hover:border-blue-200 hover:shadow-lg hover:shadow-blue-100/50 transition-all duration-300 cursor-pointer group"
      onClick={onClick}
    >
      <div className="flex items-center gap-3 mb-3">
        <div
          className="w-10 h-10 rounded-lg flex items-center justify-center font-bold text-sm"
          style={{ backgroundColor: model.color + "15", color: model.color }}
        >
          {model.name.slice(0, 2)}
        </div>
        <div>
          <div className="text-gray-900 font-medium text-sm">{model.name}</div>
          <div className="text-gray-400 text-xs">{model.provider}</div>
        </div>
      </div>
      <div className="text-gray-500 text-xs">{model.desc}</div>
      <div className="mt-3 flex items-center gap-1.5 text-blue-500 text-xs opacity-0 group-hover:opacity-100 transition-opacity">
        <span>查看详情</span>
        <ArrowRightOutlined className="text-[10px]" />
      </div>
    </div>
  );
}

/* ─── AI 对话预览 ─── */
function ChatPreview() {
  const [displayedText, setDisplayedText] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const textRef = useRef("");
  const indexRef = useRef(0);
  const timerRef = useRef<number | undefined>(undefined);

  const startTyping = useCallback(() => {
    textRef.current = DEMO_MESSAGES[1].content;
    indexRef.current = 0;
    setDisplayedText("");
    setIsTyping(true);

    const type = () => {
      if (indexRef.current < textRef.current.length) {
        const nextIndex = Math.min(indexRef.current + 3, textRef.current.length);
        setDisplayedText(textRef.current.slice(0, nextIndex));
        indexRef.current = nextIndex;
        timerRef.current = window.setTimeout(type, 30);
      } else {
        setIsTyping(false);
        timerRef.current = window.setTimeout(startTyping, 3000);
      }
    };
    type();
  }, []);

  useEffect(() => {
    const delay = window.setTimeout(startTyping, 1000);
    return () => {
      clearTimeout(delay);
      clearTimeout(timerRef.current);
    };
  }, [startTyping]);

  return (
    <div className="w-full max-w-md bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-xl shadow-gray-200/50">
      {/* 顶部栏 */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100 bg-gray-50/50">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-green-500" />
          <span className="text-gray-600 text-xs font-medium">FastAX Chat</span>
        </div>
        <Tag color="blue" className="!rounded-full !text-[10px] !px-2 !py-0">
          DeepSeek-V3
        </Tag>
      </div>

      {/* 消息区 */}
      <div className="p-4 space-y-4 min-h-[200px] bg-gradient-to-b from-gray-50/30 to-white">
        {/* 用户消息 */}
        <div className="flex justify-end">
          <div className="bg-blue-500 text-white text-sm px-4 py-2 rounded-2xl rounded-br-md max-w-[80%]">
            {DEMO_MESSAGES[0].content}
          </div>
        </div>

        {/* AI 回复 */}
        <div className="flex gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center flex-shrink-0">
            <span className="text-white text-xs font-bold">AI</span>
          </div>
          <div className="bg-gray-100 text-gray-700 text-sm px-4 py-3 rounded-2xl rounded-tl-md max-w-[80%] leading-relaxed">
            {displayedText}
            {isTyping && <span className="inline-block w-1.5 h-4 bg-gray-400 ml-1 animate-pulse" />}
          </div>
        </div>
      </div>

      {/* 输入框 */}
      <div className="px-4 pb-4">
        <div className="flex items-center gap-2 bg-gray-50 rounded-xl px-4 py-3 border border-gray-100">
          <span className="text-gray-400 text-sm flex-1">输入消息...</span>
          <div className="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center cursor-pointer hover:bg-blue-600 transition-colors">
            <ArrowRightOutlined className="text-white text-xs" />
          </div>
        </div>
      </div>
    </div>
  );
}

/* ─── 移动端菜单 ─── */
function MobileMenu({ isOpen, onClose, onNavigate }: {
  isOpen: boolean;
  onClose: () => void;
  onNavigate: (path: string) => void;
}) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-white/95 backdrop-blur-lg flex flex-col items-center justify-center">
      <button
        type="button"
        onClick={onClose}
        className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
      >
        <CloseOutlined className="text-2xl" />
      </button>
      <div className="space-y-6 text-center">
        {NAV_ITEMS.map((item) => (
          <a
            key={item.label}
            href={item.href}
            className="block text-2xl text-gray-600 hover:text-blue-600 transition-colors"
            onClick={(e) => {
              if (item.href.startsWith("/")) {
                e.preventDefault();
                onNavigate(item.href);
                onClose();
              }
            }}
          >
            {item.label}
          </a>
        ))}
        <div className="pt-6 space-y-4">
          <button
            type="button"
            onClick={() => { onNavigate("/login"); onClose(); }}
            className="w-48 py-3 border border-gray-200 text-gray-600 hover:text-blue-600 hover:border-blue-300 rounded-xl transition-colors"
          >
            登录
          </button>
          <button
            type="button"
            onClick={() => { onNavigate("/register"); onClose(); }}
            className="w-48 py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-xl transition-colors"
          >
            注册
          </button>
        </div>
      </div>
    </div>
  );
}

/* ═══════ 主页面 ═══════ */
export default function HomePage() {
  const navigate = useNavigate();
  const [products, setProducts] = useState<TokenProduct[]>([]);
  const [loading, setLoading] = useState(true);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  useEffect(() => {
    async function fetchProducts() {
      try {
        const res = await tokenService.getProducts();
        setProducts(res.data.data || []);
      } catch {
        // fallback
      } finally {
        setLoading(false);
      }
    }
    fetchProducts();
  }, []);

  const handleNavigate = (path: string) => {
    navigate(path);
  };

  return (
    <div className="bg-white text-gray-900">
      {/* ═══════ 导航栏 ═══════ */}
      <nav className="fixed top-0 left-0 right-0 z-40 bg-white/80 backdrop-blur-xl border-b border-gray-100">
        <div className="max-w-[1200px] mx-auto px-6 h-16 flex items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2.5 cursor-pointer" onClick={() => handleNavigate("/")}>
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center">
              <RocketOutlined className="text-white text-sm" />
            </div>
            <span className="text-gray-900 font-bold text-lg">FastAX</span>
          </div>

          {/* 桌面端导航 */}
          <div className="hidden md:flex items-center gap-8">
            {NAV_ITEMS.map((item) => (
              <a
                key={item.label}
                href={item.href}
                className="text-gray-500 hover:text-blue-600 text-sm transition-colors"
                onClick={(e) => {
                  if (item.href.startsWith("/")) {
                    e.preventDefault();
                    handleNavigate(item.href);
                  }
                }}
              >
                {item.label}
              </a>
            ))}
          </div>

          {/* 操作按钮 */}
          <div className="hidden md:flex items-center gap-3">
            <button
              type="button"
              onClick={() => handleNavigate("/login")}
              className="px-4 py-2 text-gray-500 hover:text-blue-600 text-sm transition-colors"
            >
              登录
            </button>
            <button
              type="button"
              onClick={() => handleNavigate("/register")}
              className="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium rounded-lg transition-colors"
            >
              免费试用
            </button>
          </div>

          {/* 移动端菜单按钮 */}
          <button
            type="button"
            className="md:hidden text-gray-500 hover:text-gray-700"
            onClick={() => setMobileMenuOpen(true)}
          >
            <MenuOutlined className="text-xl" />
          </button>
        </div>
      </nav>

      {/* 移动端菜单 */}
      <MobileMenu
        isOpen={mobileMenuOpen}
        onClose={() => setMobileMenuOpen(false)}
        onNavigate={handleNavigate}
      />

      {/* ═══════ Hero 区 ═══════ */}
      <section className="relative min-h-screen flex items-center overflow-hidden bg-gradient-to-b from-blue-50/50 via-white to-white">
        {/* 背景效果 */}
        <div className="absolute inset-0">
          <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] rounded-full bg-gradient-to-br from-blue-100/40 to-transparent blur-3xl" />
          <div className="absolute bottom-[-10%] right-[-5%] w-[50%] h-[50%] rounded-full bg-gradient-to-tl from-indigo-100/30 to-transparent blur-3xl" />
          <ParticleBg />
        </div>

        <div className="relative z-10 max-w-[1200px] mx-auto px-6 w-full py-20">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            {/* 左侧：文字内容 */}
            <div>
              {/* 标签 */}
              <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-blue-100 bg-blue-50 text-blue-600 text-sm mb-8">
                <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                <span>支持 30+ AI 模型</span>
              </div>

              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold leading-tight mb-6">
                <span className="text-gray-900">新一代</span>
                <br />
                <span className="bg-gradient-to-r from-blue-600 via-indigo-600 to-violet-600 bg-clip-text text-transparent">
                  AI Token 平台
                </span>
              </h1>

              <p className="text-lg text-gray-500 max-w-lg mb-8 leading-relaxed">
                一站式接入 OpenAI、Claude、DeepSeek 等顶级大模型。
                <br />
                智能路由、实时监控、安全合规，让每个 Token 物超所值。
              </p>

              <div className="flex items-center gap-4 mb-12">
                <button
                  type="button"
                  onClick={() => handleNavigate("/register")}
                  className="px-8 py-3.5 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-all duration-200 hover:scale-105 hover:shadow-lg hover:shadow-blue-500/25 flex items-center gap-2"
                >
                  开始使用
                  <ArrowRightOutlined />
                </button>
                <button
                  type="button"
                  onClick={() => handleNavigate("/docs")}
                  className="px-8 py-3.5 border border-gray-200 hover:border-blue-300 text-gray-600 hover:text-blue-600 font-medium rounded-xl transition-all duration-200 flex items-center gap-2"
                >
                  <CodeOutlined />
                  API 文档
                </button>
              </div>

              {/* 统计数据 */}
              <div className="grid grid-cols-3 gap-8 max-w-md">
                {[
                  { label: "接入模型", value: 30, suffix: "+" },
                  { label: "供应商", value: 15, suffix: "+" },
                  { label: "累计调用", value: 12500, suffix: "万+" },
                ].map((stat) => (
                  <div key={stat.label}>
                    <div className="text-2xl font-bold text-gray-900">
                      <AnimatedNumber target={stat.value} suffix={stat.suffix} />
                    </div>
                    <div className="text-gray-400 text-sm mt-1">{stat.label}</div>
                  </div>
                ))}
              </div>
            </div>

            {/* 右侧：AI 对话预览 */}
            <div className="hidden lg:flex justify-center">
              <ChatPreview />
            </div>
          </div>
        </div>
      </section>

      {/* ═══════ 模型展示区 ═══════ */}
      <section id="products" className="py-20 bg-gray-50">
        <div className="max-w-[1200px] mx-auto px-6">
          <div className="text-center mb-12">
            <Tag color="blue" className="!rounded-full !mb-4">支持模型</Tag>
            <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
              一站式接入全球顶级模型
            </h2>
            <p className="text-gray-400 max-w-lg mx-auto">
              覆盖对话、编程、推理、多模态等场景，价格透明，按量付费
            </p>
          </div>

          {/* 模型横向滚动 */}
          <div className="overflow-x-auto pb-4 scrollbar-hide">
            <div className="flex gap-4 min-w-max px-4">
              {MODELS.map((model) => (
                <ModelCard
                  key={model.name}
                  model={model}
                  onClick={() => handleNavigate("/market")}
                />
              ))}
            </div>
          </div>

          <div className="text-center mt-8">
            <button
              type="button"
              onClick={() => handleNavigate("/market")}
              className="text-blue-500 hover:text-blue-600 text-sm font-medium transition-colors flex items-center gap-1.5 mx-auto"
            >
              查看全部模型
              <ArrowRightOutlined className="text-xs" />
            </button>
          </div>
        </div>
      </section>

      {/* ═══════ 特性区 ═══════ */}
      <section className="py-20 bg-white">
        <div className="max-w-[1200px] mx-auto px-6">
          <div className="text-center mb-16">
            <Tag color="blue" className="!rounded-full !mb-4">核心优势</Tag>
            <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
              为什么选择 FastAX
            </h2>
            <p className="text-gray-400 max-w-lg mx-auto">
              不仅仅是代理转发——从路由、安全到成本优化，覆盖 AI Token 全生命周期
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {FEATURES.map((feature) => (
              <div
                key={feature.title}
                className="group p-8 rounded-2xl bg-white border border-gray-100 hover:border-blue-100 hover:shadow-xl hover:shadow-blue-50 transition-all duration-300"
              >
                <div
                  className={`w-12 h-12 rounded-xl flex items-center justify-center text-xl mb-5 ${feature.bg} transition-transform group-hover:scale-110`}
                  style={{ color: feature.color }}
                >
                  {feature.icon}
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-3">{feature.title}</h3>
                <p className="text-gray-400 text-sm leading-relaxed">{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══════ 产品区 ═══════ */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-[1200px] mx-auto px-6">
          <div className="text-center mb-16">
            <Tag color="blue" className="!rounded-full !mb-4">产品服务</Tag>
            <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
              灵活的定价方案
            </h2>
            <p className="text-gray-400 max-w-lg mx-auto">
              按量付费，无隐藏费用，注册即享免费额度
            </p>
          </div>

          {loading ? (
            <div className="flex justify-center py-20">
              <Spin size="large" />
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* 免费版 */}
              <div className="p-8 rounded-2xl bg-white border border-gray-100 hover:border-blue-100 hover:shadow-lg transition-all duration-300">
                <div className="text-gray-500 text-sm font-medium mb-2">入门版</div>
                <div className="text-4xl font-bold text-gray-900 mb-2">免费</div>
                <div className="text-gray-400 text-sm mb-6">注册即享</div>
                <ul className="space-y-3 mb-8">
                  {["100 万 Token/月", "基础模型访问", "社区支持", "标准路由"].map((item) => (
                    <li key={item} className="flex items-center gap-2 text-gray-500 text-sm">
                      <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                      {item}
                    </li>
                  ))}
                </ul>
                <button
                  type="button"
                  onClick={() => handleNavigate("/register")}
                  className="w-full py-3 border border-gray-200 text-gray-600 hover:text-blue-600 hover:border-blue-300 rounded-xl transition-colors"
                >
                  开始使用
                </button>
              </div>

              {/* 专业版 */}
              <div className="p-8 rounded-2xl bg-gradient-to-b from-blue-50 to-white border border-blue-200 relative shadow-lg shadow-blue-100/50">
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <Tag color="blue" className="!rounded-full !px-3">推荐</Tag>
                </div>
                <div className="text-blue-600 text-sm font-medium mb-2">专业版</div>
                <div className="text-4xl font-bold text-gray-900 mb-2">
                  ¥0.001
                  <span className="text-lg font-normal text-gray-400">/千 Token</span>
                </div>
                <div className="text-gray-400 text-sm mb-6">按量付费</div>
                <ul className="space-y-3 mb-8">
                  {["无限 Token 额度", "全部模型访问", "智能路由", "优先支持", "用量分析"].map((item) => (
                    <li key={item} className="flex items-center gap-2 text-gray-500 text-sm">
                      <span className="w-1.5 h-1.5 rounded-full bg-blue-500" />
                      {item}
                    </li>
                  ))}
                </ul>
                <button
                  type="button"
                  onClick={() => handleNavigate("/register")}
                  className="w-full py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-xl transition-colors"
                >
                  立即升级
                </button>
              </div>

              {/* 企业版 */}
              <div className="p-8 rounded-2xl bg-white border border-gray-100 hover:border-blue-100 hover:shadow-lg transition-all duration-300">
                <div className="text-gray-500 text-sm font-medium mb-2">企业版</div>
                <div className="text-4xl font-bold text-gray-900 mb-2">定制</div>
                <div className="text-gray-400 text-sm mb-6">联系我们</div>
                <ul className="space-y-3 mb-8">
                  {["专属部署", "SLA 保障", "安全合规", "专属客服", "定制开发"].map((item) => (
                    <li key={item} className="flex items-center gap-2 text-gray-500 text-sm">
                      <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
                      {item}
                    </li>
                  ))}
                </ul>
                <button
                  type="button"
                  onClick={() => handleNavigate("/contact")}
                  className="w-full py-3 border border-gray-200 text-gray-600 hover:text-blue-600 hover:border-blue-300 rounded-xl transition-colors"
                >
                  联系销售
                </button>
              </div>
            </div>
          )}
        </div>
      </section>

      {/* ═══════ CTA 区 ═══════ */}
      <section className="py-24 bg-gradient-to-b from-white to-blue-50">
        <div className="max-w-[1200px] mx-auto px-6 text-center">
          <div className="max-w-2xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
              准备好开始了吗？
            </h2>
            <p className="text-gray-400 text-lg mb-8">
              注册即享 100 万免费 Token，无需绑定信用卡
            </p>
            <div className="flex items-center justify-center gap-4 flex-wrap">
              <button
                type="button"
                onClick={() => handleNavigate("/register")}
                className="px-10 py-4 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-all duration-200 hover:scale-105 hover:shadow-lg hover:shadow-blue-500/25 flex items-center gap-2"
              >
                免费注册
                <ArrowRightOutlined />
              </button>
              <button
                type="button"
                onClick={() => handleNavigate("/docs")}
                className="px-10 py-4 border border-gray-200 hover:border-blue-300 text-gray-600 hover:text-blue-600 font-medium rounded-xl transition-all duration-200 flex items-center gap-2"
              >
                <CodeOutlined />
                查看文档
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* ═══════ 页脚 ═══════ */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="max-w-[1200px] mx-auto px-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
            {[
              {
                title: "产品",
                links: [
                  { label: "模型市场", path: "/market" },
                  { label: "Token 购买", path: "/tokens/buy" },
                  { label: "API 文档", path: "/docs" },
                  { label: "定价", path: "/pricing" },
                ],
              },
              {
                title: "资源",
                links: [
                  { label: "开发文档", path: "/docs" },
                  { label: "SDK 下载", path: "/sdk" },
                  { label: "状态监控", path: "/status" },
                  { label: "更新日志", path: "/changelog" },
                ],
              },
              {
                title: "公司",
                links: [
                  { label: "关于我们", path: "/about" },
                  { label: "联系我们", path: "/contact" },
                  { label: "加入我们", path: "/careers" },
                ],
              },
              {
                title: "法律",
                links: [
                  { label: "服务协议", path: "/terms" },
                  { label: "隐私政策", path: "/privacy" },
                  { label: "数据保护", path: "/data-protection" },
                ],
              },
            ].map((col) => (
              <div key={col.title}>
                <h4 className="text-white font-medium mb-4 text-sm">{col.title}</h4>
                <ul className="space-y-3">
                  {col.links.map((link) => (
                    <li key={link.label}>
                      <a
                        href={link.path}
                        className="text-gray-400 hover:text-white text-sm transition-colors"
                        onClick={(e) => {
                          e.preventDefault();
                          handleNavigate(link.path);
                        }}
                      >
                        {link.label}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>

          <div className="flex flex-col md:flex-row items-center justify-between pt-8 border-t border-gray-800">
            <div className="flex items-center gap-2 text-gray-400 text-sm mb-4 md:mb-0">
              <div className="w-6 h-6 rounded bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center">
                <RocketOutlined className="text-white text-xs" />
              </div>
              FastAX © {new Date().getFullYear()}
            </div>
            <div className="flex items-center gap-6">
              <a href="#" className="text-gray-500 hover:text-white transition-colors">
                <GlobalOutlined className="text-lg" />
              </a>
              <a href="#" className="text-gray-500 hover:text-white transition-colors">
                <CodeOutlined className="text-lg" />
              </a>
              <a href="#" className="text-gray-500 hover:text-white transition-colors">
                <DollarOutlined className="text-lg" />
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
