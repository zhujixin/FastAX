import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, App } from "antd";
import { useNavigate } from "react-router-dom";
import { tokenService } from "@/shared/api";
import type { UserToken } from "@/shared/api/types";

export default function MyTokensPage() {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [tokens, setTokens] = useState<UserToken[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    tokenService.getMyTokens()
      .then((res) => setTokens(res.data.data || []))
      .catch(() => message.error("加载失败"))
      .finally(() => setLoading(false));
  }, []);

  const columns = [
    { title: "产品名称", dataIndex: "product_name", key: "name" },
    { title: "总量", dataIndex: "total", key: "total", render: (v: number) => v?.toLocaleString?.() ?? v },
    { title: "已用", dataIndex: "used", key: "used", render: (v: number) => v?.toLocaleString?.() ?? v },
    { title: "剩余", dataIndex: "remaining", key: "remaining", render: (v: number) => <span className="font-bold text-blue-600">{v?.toLocaleString?.() ?? v}</span> },
    { title: "状态", dataIndex: "status", key: "status", render: (s: string) => s === "active" ? <Tag color="green">活跃</Tag> : <Tag color="red">已用完</Tag> },
    { title: "到期", dataIndex: "expires_at", key: "expires", render: (v: string) => v ? new Date(v).toLocaleDateString() : "-" },
    { title: "操作", key: "action", render: (_: unknown, r: UserToken) => <Button size="small" onClick={() => { message.info("提取 API Key"); }}>提取</Button> },
  ];

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">我的 Token</h2>
        <Button type="primary" onClick={() => navigate("/tokens/buy")}>购买 Token</Button>
      </div>
      <Card>
        <Table dataSource={tokens} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </Card>
    </div>
  );
}
