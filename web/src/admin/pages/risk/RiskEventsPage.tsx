import { useEffect, useState, useCallback } from "react";
import { Card, Table, Tag, Button, Space, message } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";
import type { RiskEvent } from "@/shared/api/types";

const LEVEL_MAP: Record<string, { color: string; text: string }> = {
  green: { color: "green", text: "记录" },
  yellow: { color: "gold", text: "预警" },
  orange: { color: "orange", text: "限权" },
  red: { color: "red", text: "冻结" },
};

export default function RiskEventsPage() {
  const [events, setEvents] = useState<RiskEvent[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchEvents = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminService.listRiskEvents();
      const data = res.data.data;
      setEvents(data?.items || (data as unknown as RiskEvent[]) || []);
    } catch {
      message.error("加载风控事件失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchEvents(); }, [fetchEvents]);

  const handleEvent = async (id: number, action: string) => {
    try {
      await adminService.handleRiskEvent(id, action);
      message.success("事件已处理");
      fetchEvents();
    } catch {
      message.error("处理失败");
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "类型", dataIndex: "type", width: 120 },
    { title: "关联", dataIndex: "user_identifier", width: 140 },
    {
      title: "等级", dataIndex: "level", width: 80,
      render: (l: string) => {
        const m = LEVEL_MAP[l] || { color: "default", text: l };
        return <Tag color={m.color}>{m.text}</Tag>;
      },
    },
    { title: "时间", dataIndex: "created_at", width: 170, render: (v: string) => new Date(v).toLocaleString() },
    {
      title: "处理状态", dataIndex: "handled", width: 90,
      render: (h: boolean) => h ? <Tag color="green">已处理</Tag> : <Tag color="red">未处理</Tag>,
    },
    {
      title: "操作", width: 120,
      render: (_: unknown, record: RiskEvent) => (
        record.handled ? null : (
          <Space>
            <Button size="small" onClick={() => handleEvent(record.id, "ignore")}>忽略</Button>
            <Button size="small" danger onClick={() => handleEvent(record.id, "freeze")}>冻结</Button>
          </Space>
        )
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">风控事件</h2>
        <Button icon={<ReloadOutlined />} onClick={fetchEvents}>刷新</Button>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={events} loading={loading} pagination={{ pageSize: 20 }} />
      </Card>
    </div>
  );
}
