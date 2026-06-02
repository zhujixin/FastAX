import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, Input, Space, App } from "antd";
import { SearchOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api";
import type { UserInfo } from "@/shared/api/types";

export default function UserListPage() {
  const { message } = App.useApp();
  const [users, setUsers] = useState<UserInfo[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    adminService.listUsers()
      .then((res) => setUsers(res.data.data?.items || []))
      .catch(() => message.error("加载用户列表失败"))
      .finally(() => setLoading(false));
  }, []);

  const handleFreeze = async (id: number, currentStatus: string) => {
    try {
      const newStatus = currentStatus === "active" ? "frozen" : "active";
      await adminService.setUserStatus(id, newStatus);
      message.success("状态更新成功");
      setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, level: newStatus } : u)));
    } catch {
      message.error("操作失败");
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "用户名", dataIndex: "username" },
    { title: "角色", dataIndex: "role", render: (r: string) => <Tag color={r === "admin" ? "red" : "blue"}>{r}</Tag> },
    { title: "等级", dataIndex: "level" },
    { title: "语言", dataIndex: "preferred_language" },
    {
      title: "操作", render: (_: unknown, r: UserInfo) => (
        <Space>
          <Button size="small" onClick={() => message.info(`用户详情 ID:${r.id}`)}>详情</Button>
          <Button size="small" onClick={() => adminService.setUserLevel(r.id, r.level === "vip" ? "normal" : "vip").then(() => message.success("已更新")).catch(() => message.error("失败"))}>升级</Button>
          <Button size="small" danger onClick={() => handleFreeze(r.id, r.level)}>冻结</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">用户管理</h2>
        <Input prefix={<SearchOutlined />} placeholder="搜索用户" className="w-64" />
      </div>
      <Card>
        <Table dataSource={users} columns={columns} rowKey="id" loading={loading} />
      </Card>
    </div>
  );
}
