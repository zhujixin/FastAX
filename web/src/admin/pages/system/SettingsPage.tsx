import { useEffect, useState } from "react";
import { Card, Tabs, Form, Input, InputNumber, Button, Switch, Select, Divider, App } from "antd";
import { SaveOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";
import { showSuccess, showError, isRoot } from "@/shared/utils";

interface OptionItem {
  key: string;
  value: string;
}

export default function SettingsPage() {
  const { message } = App.useApp();
  const [options, setOptions] = useState<Record<string, string>>({});
  const [originOptions, setOriginOptions] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const admin = isRoot();

  useEffect(() => {
    loadOptions();
  }, []);

  const loadOptions = async () => {
    setLoading(true);
    try {
      // Load system status/config from backend
      const res = await api.get("/admin/system/config");
      const data = res.data.data || {};
      setOptions(data);
      setOriginOptions(data);
    } catch {
      // Defaults for development
      const defaults: Record<string, string> = {
        SystemName: "FastAX",
        Logo: "",
        FooterHTML: "",
        ServerAddress: window.location.origin,
        PasswordLoginEnabled: "true",
        GitHubOAuthEnabled: "false",
        WeChatAuthEnabled: "false",
        RegisterEnabled: "true",
        EmailVerificationEnabled: "false",
        RateLimitEnabled: "true",
        RateLimitCount: "60",
        QuotaPerUnit: "500000",
        DisplayInCurrency: "false",
        ChannelDisableThreshold: "0.9",
      };
      setOptions(defaults);
      setOriginOptions(defaults);
    } finally {
      setLoading(false);
    }
  };

  const updateOption = async (key: string, value: string) => {
    try {
      await api.put("/admin/system/config", { key, value });
      setOriginOptions({ ...originOptions, [key]: value });
    } catch (err) { showError(err); }
  };

  const handleToggle = async (key: string, checked: boolean) => {
    const value = checked ? "true" : "false";
    setOptions({ ...options, [key]: value });
    await updateOption(key, value);
    showSuccess("已更新");
  };

  const handleSave = async () => {
    const changed: Record<string, string> = {};
    for (const key of Object.keys(options)) {
      if (options[key] !== originOptions[key] && !key.endsWith("Enabled")) {
        changed[key] = options[key];
      }
    }
    if (Object.keys(changed).length === 0) { message.info("无更改"); return; }
    try {
      await api.put("/admin/system/config", changed);
      setOriginOptions({ ...options });
      showSuccess("配置已保存");
    } catch (err) { showError(err); }
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">系统设置</h2>
      <Tabs defaultActiveKey="general" items={[
        {
          key: "general",
          label: "通用",
          children: (
            <Card loading={loading}>
              <Form layout="vertical">
                <Form.Item label="站点名称">
                  <Input value={options.SystemName} onChange={(e) => setOptions({ ...options, SystemName: e.target.value })} />
                </Form.Item>
                <Form.Item label="服务器地址">
                  <Input value={options.ServerAddress} onChange={(e) => setOptions({ ...options, ServerAddress: e.target.value })} />
                </Form.Item>
                <Form.Item label="页脚 HTML">
                  <Input.TextArea value={options.FooterHTML} onChange={(e) => setOptions({ ...options, FooterHTML: e.target.value })} rows={3} />
                </Form.Item>
                <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>保存通用设置</Button>
              </Form>
            </Card>
          ),
        },
        ...(admin ? [{
          key: "auth",
          label: "认证",
          children: (
            <Card loading={loading}>
              <Form layout="vertical">
                <Form.Item label="密码登录">
                  <Switch checked={options.PasswordLoginEnabled === "true"} onChange={(v) => handleToggle("PasswordLoginEnabled", v)} />
                </Form.Item>
                <Form.Item label="注册开放">
                  <Switch checked={options.RegisterEnabled === "true"} onChange={(v) => handleToggle("RegisterEnabled", v)} />
                </Form.Item>
                <Form.Item label="GitHub OAuth">
                  <Switch checked={options.GitHubOAuthEnabled === "true"} onChange={(v) => handleToggle("GitHubOAuthEnabled", v)} />
                </Form.Item>
                <Form.Item label="邮箱验证">
                  <Switch checked={options.EmailVerificationEnabled === "true"} onChange={(v) => handleToggle("EmailVerificationEnabled", v)} />
                </Form.Item>
              </Form>
            </Card>
          ),
        }] : []),
        ...(admin ? [{
          key: "rate",
          label: "限流",
          children: (
            <Card loading={loading}>
              <Form layout="vertical">
                <Form.Item label="启用限流">
                  <Switch checked={options.RateLimitEnabled === "true"} onChange={(v) => handleToggle("RateLimitEnabled", v)} />
                </Form.Item>
                <Form.Item label="限流次数 (每分钟)">
                  <InputNumber value={Number(options.RateLimitCount) || 60} onChange={(v) => setOptions({ ...options, RateLimitCount: String(v) })} min={1} max={10000} className="w-32" />
                </Form.Item>
                <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>保存限流设置</Button>
              </Form>
            </Card>
          ),
        }] : []),
        {
          key: "display",
          label: "显示",
          children: (
            <Card loading={loading}>
              <Form layout="vertical">
                <Form.Item label="显示为货币">
                  <Switch checked={options.DisplayInCurrency === "true"} onChange={(v) => handleToggle("DisplayInCurrency", v)} />
                  <span className="text-gray-400 ml-2">启用后将以货币形式显示额度</span>
                </Form.Item>
                <Form.Item label="每单位额度对应的 Token 数">
                  <InputNumber value={Number(options.QuotaPerUnit) || 500000} onChange={(v) => setOptions({ ...options, QuotaPerUnit: String(v) })} min={1} className="w-48" />
                </Form.Item>
                <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>保存显示设置</Button>
              </Form>
            </Card>
          ),
        },
      ]} />
    </div>
  );
}
