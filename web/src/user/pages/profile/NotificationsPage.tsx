import { useEffect, useState } from "react";
import { Card, List, Tag, Button, Space, message, Empty, Badge } from "antd";
import { MailOutlined, ReloadOutlined } from "@ant-design/icons";
import { notifyService } from "@/shared/api/notifications";
import type { Notification } from "@/shared/api/types";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchNotifications = async () => {
    setLoading(true);
    try {
      const res = await notifyService.list();
      setNotifications(res.data.data || []);
    } catch {
      message.error("加载通知失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchNotifications(); }, []);

  const handleMarkRead = async (id: number) => {
    try {
      await notifyService.markRead(id);
      setNotifications((prev) => prev.map((n) => n.id === id ? { ...n, read: true } : n));
    } catch {
      message.error("操作失败");
    }
  };

  const handleMarkAll = async () => {
    try {
      await notifyService.markAllRead();
      setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
      message.success("已全部标为已读");
    } catch {
      message.error("操作失败");
    }
  };

  return (
    <div className="max-w-3xl mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-bold">消息中心</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchNotifications}>刷新</Button>
          <Button onClick={handleMarkAll}>全部已读</Button>
        </Space>
      </div>
      <Card>
        {notifications.length === 0 && !loading ? (
          <Empty description="暂无消息" />
        ) : (
          <List
            loading={loading}
            dataSource={notifications}
            renderItem={(item) => (
              <List.Item
                extra={<Tag color={item.read ? "default" : "blue"}>{item.read ? "已读" : "未读"}</Tag>}
                actions={!item.read ? [
                  <Button key="read" type="link" size="small" onClick={() => handleMarkRead(item.id)}>标为已读</Button>
                ] : undefined}
              >
                <List.Item.Meta
                  avatar={!item.read ? <Badge dot><MailOutlined className="text-lg" /></Badge> : <MailOutlined className="text-lg text-gray-400" />}
                  title={<span className={item.read ? "text-gray-500" : "font-bold"}>{item.title}</span>}
                  description={`${item.content} — ${new Date(item.created_at).toLocaleString()}`}
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  );
}
