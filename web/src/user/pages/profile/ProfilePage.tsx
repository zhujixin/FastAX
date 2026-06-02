import { Card, Descriptions, Select, App } from "antd";
import { useAuthStore } from "@/shared/stores/authStore";

export default function ProfilePage() {
  const { user, setUser } = useAuthStore();
  const { message } = App.useApp();

  const handleLanguageChange = (lang: string) => {
    if (user) {
      setUser({ ...user, preferred_language: lang });
      localStorage.setItem("i18nextLng", lang);
      message.success("语言偏好已更新");
    }
  };

  const roleMap: Record<string, string> = {
    admin: "管理员",
    super_admin: "超级管理员",
    vendor: "供应商",
    user: "用户",
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-6">个人中心</h2>
      <Card>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="用户 ID">{user?.id}</Descriptions.Item>
          <Descriptions.Item label="用户名">{user?.username}</Descriptions.Item>
          <Descriptions.Item label="角色">{roleMap[user?.role || ""] || user?.role}</Descriptions.Item>
          <Descriptions.Item label="等级">{user?.level}</Descriptions.Item>
          <Descriptions.Item label="语言偏好">
            <Select
              value={user?.preferred_language || "zh-CN"}
              onChange={handleLanguageChange}
              style={{ width: 120 }}
              options={[
                { value: "zh-CN", label: "中文" },
                { value: "en", label: "English" },
                { value: "ja", label: "日本語" },
              ]}
            />
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
}
