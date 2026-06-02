import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import HttpBackend from "i18next-http-backend";

// Built-in translations as fallback (CDN /api/i18n/translations/:locale serves full files)
const resources = {
  "zh-CN": {
    common: {
      appName: "FastAX",
      login: "登录",
      register: "注册",
      logout: "退出登录",
      dashboard: "控制台",
      tokens: "Token 管理",
      orders: "订单管理",
      profile: "个人中心",
      admin: "管理后台",
      vendor: "供应商门户",
      save: "保存",
      cancel: "取消",
      delete: "删除",
      edit: "编辑",
      create: "新建",
      search: "搜索",
      export: "导出",
      loading: "加载中...",
      noData: "暂无数据",
      confirm: "确认",
      back: "返回",
    },
  },
  en: {
    common: {
      appName: "FastAX",
      login: "Login",
      register: "Register",
      logout: "Logout",
      dashboard: "Dashboard",
      tokens: "Tokens",
      orders: "Orders",
      profile: "Profile",
      admin: "Admin",
      vendor: "Vendor Portal",
      save: "Save",
      cancel: "Cancel",
      delete: "Delete",
      edit: "Edit",
      create: "Create",
      search: "Search",
      export: "Export",
      loading: "Loading...",
      noData: "No Data",
      confirm: "Confirm",
      back: "Back",
    },
  },
};

i18n
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: "zh-CN",
    supportedLngs: ["zh-CN", "en", "ja", "zh-TW"],
    defaultNS: "common",
    interpolation: {
      escapeValue: false, // React already safes from XSS
    },
    detection: {
      order: ["localStorage", "navigator"],
      caches: ["localStorage"],
    },
    backend: {
      loadPath: "/api/i18n/translations/{{lng}}",
    },
  });

export default i18n;
