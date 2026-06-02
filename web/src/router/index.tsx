import { lazy, Suspense } from "react";
import { useRoutes } from "react-router-dom";
import { Spin } from "antd";
import UserLayout from "@/user/layouts/UserLayout";
import AdminLayout from "@/admin/layouts/AdminLayout";

// Eagerly load user-facing pages for fast FCP
import HomePage from "@/user/pages/home/HomePage";
import LoginPage from "@/user/pages/auth/LoginPage";
import RegisterPage from "@/user/pages/auth/RegisterPage";
import ForgotPasswordPage from "@/user/pages/auth/ForgotPasswordPage";
import DashboardPage from "@/user/pages/dashboard/DashboardPage";
import MyTokensPage from "@/user/pages/tokens/MyTokensPage";
import BuyTokenPage from "@/user/pages/tokens/BuyTokenPage";
import TopUpPage from "@/user/pages/tokens/TopUpPage";
import OrderListPage from "@/user/pages/orders/OrderListPage";
import OrderDetailPage from "@/user/pages/orders/OrderDetailPage";
import ProfilePage from "@/user/pages/profile/ProfilePage";
import SubAccountsPage from "@/user/pages/profile/SubAccountsPage";
import BillsPage from "@/user/pages/profile/BillsPage";
import NotificationsPage from "@/user/pages/profile/NotificationsPage";
import VendorDashboard from "@/user/pages/vendor/Dashboard/Dashboard";
import VendorProducts from "@/user/pages/vendor/Products/Products";
import VendorOrders from "@/user/pages/vendor/Orders/Orders";
import VendorSettlements from "@/user/pages/vendor/Settlements/Settlements";
import MyBYOKKeysPage from "@/user/pages/byok/MyBYOKKeysPage";
import MarketPage from "@/user/pages/market/MarketPage";

// Lazy load admin pages for smaller initial bundle
const AdminDashboard = lazy(() => import("@/admin/pages/dashboard/DashboardPage"));
const AdminUsers = lazy(() => import("@/admin/pages/users/UserListPage"));
const AdminTokens = lazy(() => import("@/admin/pages/tokens/TokenManagePage"));
const AdminChannels = lazy(() => import("@/admin/pages/tokens/ChannelManagePage"));
const AdminOrders = lazy(() => import("@/admin/pages/orders/OrderManagePage"));
const AdminRiskEvents = lazy(() => import("@/admin/pages/risk/RiskEventsPage"));
const AdminRiskRules = lazy(() => import("@/admin/pages/risk/Rules/RuleListPage"));
const AdminBlacklist = lazy(() => import("@/admin/pages/risk/Blacklist/BlacklistPage"));
const AdminVendors = lazy(() => import("@/admin/pages/vendors/VendorListPage"));
const AdminI18n = lazy(() => import("@/admin/pages/i18n-config/I18nConfigPage"));
const AdminLogs = lazy(() => import("@/admin/pages/system/LogViewerPage"));
const AdminSettings = lazy(() => import("@/admin/pages/system/SettingsPage"));
const AdminAuditLogs = lazy(() => import("@/admin/pages/system/AuditLogPage"));
const AdminGuardrailRules = lazy(() => import("@/admin/pages/guardrails/GuardrailRulesPage"));
const AdminGuardrailLogs = lazy(() => import("@/admin/pages/guardrails/GuardrailLogsPage"));
const AdminEnterprise = lazy(() => import("@/admin/pages/enterprise/EnterprisePage"));
const AdminCost = lazy(() => import("@/admin/pages/cost/CostPage"));

function Lazy({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={<div className="flex items-center justify-center h-64"><Spin size="large" /></div>}>
      {children}
    </Suspense>
  );
}

const routes = [
  {
    element: <UserLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/forgot-password", element: <ForgotPasswordPage /> },
      { path: "/dashboard", element: <DashboardPage /> },
      { path: "/tokens", element: <MyTokensPage /> },
      { path: "/tokens/:id", element: <MyTokensPage /> },
      { path: "/tokens/buy", element: <BuyTokenPage /> },
      { path: "/tokens/topup", element: <TopUpPage /> },
      { path: "/orders", element: <OrderListPage /> },
      { path: "/orders/:id", element: <OrderDetailPage /> },
      { path: "/profile", element: <ProfilePage /> },
      { path: "/profile/sub-accounts", element: <SubAccountsPage /> },
      { path: "/bills", element: <BillsPage /> },
      { path: "/notifications", element: <NotificationsPage /> },
      { path: "/vendor/dashboard", element: <VendorDashboard /> },
      { path: "/vendor/products", element: <VendorProducts /> },
      { path: "/vendor/orders", element: <VendorOrders /> },
      { path: "/vendor/settlements", element: <VendorSettlements /> },
      { path: "/byok/keys", element: <MyBYOKKeysPage /> },
      { path: "/market", element: <MarketPage /> },
    ],
  },
  {
    path: "/admin",
    element: <AdminLayout />,
    children: [
      { index: true, element: <Lazy><AdminDashboard /></Lazy> },
      { path: "users", element: <Lazy><AdminUsers /></Lazy> },
      { path: "tokens", element: <Lazy><AdminTokens /></Lazy> },
      { path: "channels", element: <Lazy><AdminChannels /></Lazy> },
      { path: "orders", element: <Lazy><AdminOrders /></Lazy> },
      { path: "risk", element: <Lazy><AdminRiskEvents /></Lazy> },
      { path: "risk/rules", element: <Lazy><AdminRiskRules /></Lazy> },
      { path: "risk/blacklist", element: <Lazy><AdminBlacklist /></Lazy> },
      { path: "vendors", element: <Lazy><AdminVendors /></Lazy> },
      { path: "i18n", element: <Lazy><AdminI18n /></Lazy> },
      { path: "logs", element: <Lazy><AdminLogs /></Lazy> },
      { path: "settings", element: <Lazy><AdminSettings /></Lazy> },
      { path: "audit", element: <Lazy><AdminAuditLogs /></Lazy> },
      { path: "guardrails/rules", element: <Lazy><AdminGuardrailRules /></Lazy> },
      { path: "guardrails/logs", element: <Lazy><AdminGuardrailLogs /></Lazy> },
      { path: "enterprise", element: <Lazy><AdminEnterprise /></Lazy> },
      { path: "cost", element: <Lazy><AdminCost /></Lazy> },
    ],
  },
];

export default function AppRouter() {
  return useRoutes(routes);
}
