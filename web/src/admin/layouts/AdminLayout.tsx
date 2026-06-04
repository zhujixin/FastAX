import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { Layout, Menu, Button, Dropdown, Space, Select } from "antd";
import {
  DashboardOutlined,
  UserOutlined,
  KeyOutlined,
  ShoppingCartOutlined,
  AlertOutlined,
  ShopOutlined,
  TranslationOutlined,
  SettingOutlined,
  AuditOutlined,
  FileTextOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SunOutlined,
  MoonOutlined,
  GlobalOutlined,
  SafetyOutlined,
  TeamOutlined,
  DollarCircleOutlined,
} from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/shared/stores/authStore";
import { useThemeStore } from "@/shared/stores/themeStore";
import { useState } from "react";

const { Header, Sider, Content } = Layout;

const adminMenuItems = [
  { key: "/admin", icon: <DashboardOutlined />, label: "nav.dashboard" },
  { key: "/admin/users", icon: <UserOutlined />, label: "nav.users" },
  {
    key: "tokens",
    icon: <KeyOutlined />,
    label: "nav.tokenMgmt",
    children: [
      { key: "/admin/tokens", label: "nav.products" },
      { key: "/admin/channels", label: "nav.channels" },
{ key: "/admin/suppliers", label: "nav.suppliers" },
    ],
  },
  { key: "/admin/orders", icon: <ShoppingCartOutlined />, label: "nav.orders" },
{ key: "/admin/reports", icon: <BarChartOutlined />, label: "nav.reports" },
  {
    key: "risk",
    icon: <AlertOutlined />,
    label: "nav.risk",
    children: [
      { key: "/admin/risk", label: "nav.riskEvents" },
      { key: "/admin/risk/rules", label: "nav.riskRules" },
      { key: "/admin/risk/blacklist", label: "nav.blacklist" },
    ],
  },
  { key: "/admin/vendors", icon: <ShopOutlined />, label: "nav.vendors" },
  {
    key: "guardrails",
    icon: <SafetyOutlined />,
    label: "nav.guardrails",
    children: [
      { key: "/admin/guardrails/rules", label: "nav.guardrailRules" },
      { key: "/admin/guardrails/logs", label: "nav.guardrailLogs" },
    ],
  },
  { key: "/admin/i18n", icon: <TranslationOutlined />, label: "nav.i18n" },
  { key: "/admin/logs", icon: <FileTextOutlined />, label: "nav.logs" },
  { key: "/admin/settings", icon: <SettingOutlined />, label: "nav.settings" },
  { key: "/admin/audit", icon: <AuditOutlined />, label: "nav.auditLog" },
  { key: "/admin/enterprise", icon: <TeamOutlined />, label: "nav.enterprise" },
  { key: "/admin/system", icon: <SettingOutlined />, label: "nav.system" },
{ key: "/admin/system/admins", label: "nav.admins" },
  { key: "/admin/cost", icon: <DollarCircleOutlined />, label: "nav.cost" },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { user, logout } = useAuthStore();
  const { mode, toggle } = useThemeStore();
  const [collapsed, setCollapsed] = useState(false);

  // Map i18n keys to Chinese fallback for now
  const labelMap: Record<string, string> = {
    "nav.dashboard": "运营看板",
    "nav.users": "用户管理",
    "nav.tokenMgmt": "Token 管理",
    "nav.products": "商品管理",
    "nav.channels": "渠道管理",
    "nav.risk": "风控管理",
    "nav.riskEvents": "风控事件",
    "nav.riskRules": "风控规则",
    "nav.blacklist": "黑名单",
    "nav.orders": "交易管理",
    "nav.vendors": "供应商管理",
    "nav.i18n": "多语言配置",
    "nav.logs": "日志查看",
    "nav.settings": "系统设置",
    "nav.auditLog": "审计日志",
    "nav.guardrails": "安全护栏",
    "nav.guardrailRules": "护栏规则",
    "nav.guardrailLogs": "检测日志",
    "nav.enterprise": "企业管理",
    "nav.system": "系统配置",
    "nav.cost": "成本优化",
"nav.suppliers": "平台供应商",
"nav.reports": "经营报表",
"nav.admins": "管理员管理",
  };

  const getSelectedKeys = () => {
    const path = location.pathname;
    if (path === "/admin" || path === "/admin/") return ["/admin"];
    return [path];
  };

  const getOpenKeys = () => {
    const path = location.pathname;
    if (path.startsWith("/admin/risk")) return ["risk"];
    if (path.startsWith("/admin/tokens")) return ["tokens"];
    if (path.startsWith("/admin/guardrails")) return ["guardrails"];
    return [];
  };

  const userMenu = {
    items: [
      { key: "back", icon: <DashboardOutlined />, label: "返回用户端", onClick: () => navigate("/dashboard") },
      { type: "divider" as const },
      {
        key: "logout",
        icon: <LogoutOutlined />,
        label: t("common.logout"),
        onClick: () => {
          logout();
          navigate("/login");
        },
      },
    ],
  };

  return (
    <Layout className="min-h-screen">
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        theme="dark"
        className="min-h-screen"
      >
        <div className="h-14 flex items-center justify-center text-white text-lg font-bold border-b border-gray-700">
          {collapsed ? "FX" : "FastAX Admin"}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={adminMenuItems.map((item) => ({
            ...item,
            label: labelMap[item.label] || item.label,
          }))}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>

      <Layout>
        <Header className="flex items-center justify-between px-4 bg-white border-b border-gray-200">
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Space>
            <Select
              defaultValue={i18n.language}
              onChange={(lang: string) => i18n.changeLanguage(lang)}
              size="small"
              className="w-24"
              suffixIcon={<GlobalOutlined />}
              options={[
                { value: "zh-CN", label: "中文" },
                { value: "en", label: "EN" },
              ]}
            />
            <Button
              type="text"
              icon={mode === "light" ? <MoonOutlined /> : <SunOutlined />}
              onClick={toggle}
            />
            <Dropdown menu={userMenu} placement="bottomRight">
              <Button type="text" icon={<UserOutlined />}>
                {user?.username}
              </Button>
            </Dropdown>
          </Space>
        </Header>

        <Content className="m-6 p-6 bg-white rounded-lg min-h-[calc(100vh-64px)]">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
