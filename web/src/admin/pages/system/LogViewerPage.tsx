import { useEffect, useState } from "react";
import { Card, Table, Input, Select, Space, Button, DatePicker, Tag, App } from "antd";
import { SearchOutlined, ReloadOutlined, CopyOutlined, BarChartOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";
import { showError, showSuccess, copy, timestamp2string, isAdmin } from "@/shared/utils";
import { renderNumber, renderQuota } from "@/shared/utils/render";
import { ITEMS_PER_PAGE, LOG_TYPES } from "@/shared/constants";

interface LogEntry {
  id: number;
  user_id: number;
  username: string;
  token_name: string;
  model_name: string;
  channel_id: number;
  prompt_tokens: number;
  completion_tokens: number;
  quota: number;
  type: number;
  content: string;
  request_id: string;
  created_at: number;
}

export default function LogViewerPage() {
  const { message } = App.useApp();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [logType, setLogType] = useState(0);
  const [showStats, setShowStats] = useState(false);
  const [totalQuota, setTotalQuota] = useState(0);
  const [filters, setFilters] = useState({
    token_name: "",
    model_name: "",
    username: "",
    start_time: "",
    end_time: "",
  });

  const admin = isAdmin();

  const loadLogs = async () => {
    setLoading(true);
    try {
      const endpoint = admin ? "/admin/call-logs" : "/tokens/my/usage";
      const params: Record<string, unknown> = { type: logType || undefined };
      if (filters.token_name) params.token_name = filters.token_name;
      if (filters.model_name) params.model_name = filters.model_name;
      if (filters.username) params.username = filters.username;
      if (filters.start_time) params.start_time = filters.start_time;
      if (filters.end_time) params.end_time = filters.end_time;
      const res = await api.get(endpoint, { params });
      const data = res.data.data?.items || res.data.data || [];
      setLogs(data);
      setTotalQuota(data.reduce((s: number, l: LogEntry) => s + (l.quota || 0), 0));
    } catch (err) {
      showError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadLogs(); }, [logType]);

  const handleCopyRequestId = async (requestId: string) => {
    const ok = await copy(requestId);
    ok ? showSuccess("已复制 Request ID") : message.warning("复制失败");
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 50 },
    { title: "时间", dataIndex: "created_at", width: 150,
      render: (_: unknown, r: LogEntry) => (
        <code className="cursor-pointer text-xs" onClick={() => handleCopyRequestId(r.request_id)} title="点击复制 Request ID">
          {timestamp2string(r.created_at)}
        </code>
      ),
    },
    { title: "类型", dataIndex: "type", width: 60,
      render: (t: number) => {
        const info = LOG_TYPES.find((lt) => lt.value === t);
        return info ? <Tag>{info.label}</Tag> : <Tag>{t}</Tag>;
      },
    },
    ...(admin ? [{ title: "用户", dataIndex: "username", width: 80 }] : []),
    { title: "Token", dataIndex: "token_name", width: 100, ellipsis: true },
    { title: "模型", dataIndex: "model_name", width: 100, ellipsis: true },
    ...(admin && logType !== 5 ? [{ title: "渠道", dataIndex: "channel_id", width: 60 }] : []),
    ...(logType !== 5 ? [
      { title: "Prompt", dataIndex: "prompt_tokens", width: 80, render: renderNumber },
      { title: "Completion", dataIndex: "completion_tokens", width: 90, render: renderNumber },
      { title: "额度", dataIndex: "quota", width: 80, render: renderQuota },
    ] : []),
    { title: "内容", dataIndex: "content", ellipsis: true },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">日志查看</h2>
        <Space>
          <Button icon={<BarChartOutlined />} onClick={() => setShowStats(!showStats)}>
            {showStats ? "隐藏" : "显示"}统计
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadLogs}>刷新</Button>
        </Space>
      </div>

      {showStats && (
        <Card className="mb-4" size="small">
          <span className="text-gray-500">当前筛选总消费: </span>
          <span className="font-bold text-lg">{renderQuota(totalQuota)}</span>
          <span className="text-gray-400 ml-2">(共 {logs.length} 条记录)</span>
        </Card>
      )}

      <Card className="mb-4" size="small">
        <Space wrap>
          <Select value={logType} onChange={setLogType} style={{ width: 100 }}
            options={LOG_TYPES}
          />
          <Input placeholder="Token 名称" value={filters.token_name} onChange={(e) => setFilters({ ...filters, token_name: e.target.value })} className="w-32" />
          <Input placeholder="模型名称" value={filters.model_name} onChange={(e) => setFilters({ ...filters, model_name: e.target.value })} className="w-32" />
          {admin && <Input placeholder="用户名" value={filters.username} onChange={(e) => setFilters({ ...filters, username: e.target.value })} className="w-24" />}
          <DatePicker placeholder="开始时间" onChange={(_, s) => setFilters({ ...filters, start_time: typeof s === "string" ? s : "" })} />
          <DatePicker placeholder="结束时间" onChange={(_, s) => setFilters({ ...filters, end_time: typeof s === "string" ? s : "" })} />
          <Button type="primary" icon={<SearchOutlined />} onClick={loadLogs}>查询</Button>
        </Space>
      </Card>

      <Card>
        <Table
          dataSource={logs}
          columns={columns}
          rowKey="id"
          loading={loading}
          size="small"
          scroll={{ x: 1000 }}
          pagination={{ pageSize: ITEMS_PER_PAGE, showTotal: (t: number) => `共 ${t} 条` }}
        />
      </Card>
    </div>
  );
}
