import { useEffect, useState, useCallback } from "react";
import { Card, Table, Tag, Button, Space, Input, message } from "antd";
import { ReloadOutlined, DownloadOutlined, SearchOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";

interface AuditLog {
  id: number;
  user?: string;
  username?: string;
  action: string;
  resource: string;
  ip: string;
  time?: string;
  created_at?: string;
}

export default function AuditLogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminService.listAuditLogs({ page, size: 20 });
      const data = res.data.data;
      if (data) {
        setLogs((data?.items || []) as unknown as AuditLog[]);
        setTotal(data.total || 0);
      }
    } catch {
      message.error("加载审计日志失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchLogs(); }, [fetchLogs]);

  const handleExport = async () => {
    try {
      const res = await adminService.exportAuditLogs();
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const a = document.createElement("a");
      a.href = url;
      a.download = `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
      message.success("导出成功");
    } catch {
      message.error("导出失败");
    }
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "用户", dataIndex: "user", width: 100, render: (_: string, r: AuditLog) => r.user || r.username || "-" },
    { title: "操作", dataIndex: "action", width: 120, render: (a: string) => <Tag color="blue">{a}</Tag> },
    { title: "资源", dataIndex: "resource", width: 200, ellipsis: true },
    { title: "IP", dataIndex: "ip", width: 130 },
    { title: "时间", dataIndex: "time", width: 170, render: (_: string, r: AuditLog) => r.time || (r.created_at ? new Date(r.created_at).toLocaleString() : "-") },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">审计日志</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchLogs}>刷新</Button>
          <Button type="primary" icon={<DownloadOutlined />} onClick={handleExport}>导出 CSV</Button>
        </Space>
      </div>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={logs}
          loading={loading}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage, showSizeChanger: true }}
        />
      </Card>
    </div>
  );
}
