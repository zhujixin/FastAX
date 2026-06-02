import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { ConfigProvider, App as AntApp, theme } from "antd";
import zhCN from "antd/locale/zh_CN";
import enUS from "antd/locale/en_US";
import "./index.css";
import "./shared/i18n";
import { useThemeStore } from "@/shared/stores/themeStore";
import AppRouter from "./router";

const antdLocales: Record<string, typeof zhCN> = {
  "zh-CN": zhCN,
  en: enUS,
};

function Root() {
  const lang = localStorage.getItem("i18nextLng") || "zh-CN";
  const antdLocale = antdLocales[lang] || zhCN;
  const mode = useThemeStore((s) => s.mode);

  return (
    <ConfigProvider
      locale={antdLocale}
      theme={{
        algorithm: mode === "dark" ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: "#1677ff",
          borderRadius: 6,
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <AppRouter />
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Root />
  </StrictMode>
);
