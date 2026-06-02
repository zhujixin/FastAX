import { Outlet } from "react-router-dom";
import { Layout, Menu, Button, Dropdown, Space, Select } from "antd";
import {
  HomeOutlined,
  DashboardOutlined,
  KeyOutlined,
  ShoppingCartOutlined,
  UserOutlined,
  BellOutlined,
  LogoutOutlined,
  ShopOutlined,
  MenuOutlined,
  SunOutlined,
  MoonOutlined,
  GlobalOutlined,
  GiftOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { useNavigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/shared/stores/authStore";
import { useThemeStore } from "@/shared/stores/themeStore";
import { useState } from "react";

const { Header } = Layout;

export default function UserLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { isAuthenticated, user, logout } = useAuthStore();
  const { mode, toggle } = useThemeStore();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const isAuthPage = ["/login", "/register", "/forgot-password"].includes(location.pathname);

  const menuItems = [
    { key: "/", icon: <HomeOutlined />, label: t("common.home", "首页") },
    ...(isAuthenticated
      ? [
          { key: "/dashboard", icon: <DashboardOutlined />, label: t("common.dashboard") },
          { key: "/tokens", icon: <KeyOutlined />, label: t("common.tokens") },
          { key: "/tokens/topup", icon: <GiftOutlined />, label: t("common.topup", "充值") },
          { key: "/orders", icon: <ShoppingCartOutlined />, label: t("common.orders") },
          { key: "/market", icon: <ThunderboltOutlined />, label: "模型市场" },
        ]
      : []),
    ...(user?.role === "vendor"
      ? [{ key: "/vendor/dashboard", icon: <ShopOutlined />, label: t("common.vendor") }]
      : []),
  ];

  const userMenu = {
    items: [
      { key: "profile", icon: <UserOutlined />, label: t("common.profile"), onClick: () => navigate("/profile") },
      { key: "byok", icon: <KeyOutlined />, label: "BYOK 密钥", onClick: () => navigate("/byok/keys") },
      { key: "notifications", icon: <BellOutlined />, label: t("common.notifications", "消息"), onClick: () => navigate("/notifications") },
      ...(user?.role === "admin" || user?.role === "super_admin"
        ? [{ key: "admin", icon: <DashboardOutlined />, label: t("common.admin"), onClick: () => navigate("/admin") }]
        : []),
      { type: "divider" as const },
      {
        key: "logout",
        icon: <LogoutOutlined />,
        label: t("common.logout"),
        onClick: () => {
          logout();
          navigate("/");
        },
      },
    ],
  };

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
  };

  if (isAuthPage) {
    return <Outlet />;
  }

  return (
    <Layout className="min-h-screen">
      <Header className="flex items-center justify-between px-4 bg-white border-b border-gray-200 sticky top-0 z-50">
        <div className="flex items-center gap-4">
          <Button
            type="text"
            icon={<MenuOutlined />}
            className="md:hidden"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          />
          <span
            className="text-lg font-bold text-blue-600 cursor-pointer"
            onClick={() => navigate("/")}
          >
            FastAX
          </span>
          <Menu
            mode="horizontal"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            className="hidden md:flex border-0 flex-1"
          />
        </div>

        <Space>
          <Select
            defaultValue={i18n.language}
            onChange={changeLanguage}
            size="small"
            className="w-24"
            suffixIcon={<GlobalOutlined />}
            options={[
              { value: "zh-CN", label: "中文" },
              { value: "en", label: "English" },
              { value: "ja", label: "日本語" },
              { value: "zh-TW", label: "繁體" },
            ]}
          />
          <Button
            type="text"
            icon={mode === "light" ? <MoonOutlined /> : <SunOutlined />}
            onClick={toggle}
          />
          {isAuthenticated ? (
            <Dropdown menu={userMenu} placement="bottomRight">
              <Button type="text" icon={<UserOutlined />}>
                {user?.username}
              </Button>
            </Dropdown>
          ) : (
            <Space>
              <Button type="link" onClick={() => navigate("/login")}>
                {t("common.login")}
              </Button>
              <Button type="primary" onClick={() => navigate("/register")}>
                {t("common.register")}
              </Button>
            </Space>
          )}
        </Space>
      </Header>

      {mobileMenuOpen && (
        <div className="md:hidden bg-white border-b p-2">
          {menuItems.map((item) => (
            <div
              key={item.key}
              className="px-4 py-2 hover:bg-gray-50 cursor-pointer rounded"
              onClick={() => {
                navigate(item.key!);
                setMobileMenuOpen(false);
              }}
            >
              {item.label}
            </div>
          ))}
        </div>
      )}

      <div className="flex-1 bg-gray-50">
        <Outlet />
      </div>
    </Layout>
  );
}
