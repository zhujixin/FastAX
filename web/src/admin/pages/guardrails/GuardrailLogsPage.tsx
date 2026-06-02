import { useEffect, useState, useCallback } from "react";
import { Table, Input, Select, Space, Tag, Card, Row, Col, Statistic } from "antd";
import { SearchOutlined, SafetyOutlined, WarningOutlined, StopOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";

interface GuardrailLog {
  id: number;
  trace_id: string;
  user_id: number;
  rule_id: number;
  stage: string;
  detected_entities: string;
  action_taken: string;
  created_at: number;
}

export default function GuardrailLogsPage() {
  const [logs, setLogs] = useState<GuardrailLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({ trace_id: "", stage: "", user_id: "" });

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = {};
      if (filters.trace_id) params.trace_id = filters.trace_id;
      if (filters.stage) params.stage = filters.stage;
      if (filters.user_id) params.user_id = Number(filters.user_id);
      const res = await adminService.listGuardrailLogs(params);
      setLogs((res.data.data as unknown as GuardrailLog[]) || []);
    } catch {
      // error handled by axios interceptor
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => { fetchLogs(); }, [fetchLogs]);

  const stageColor = (stage: string) => stage === "before" ? "blue" : "green";
  const actionColor = (action: string) => {
    switch (action) {
      case "enforce": return "red";
      case "monitor": return "orange";
      default: return "default";
    }
  };
  const actionIcon = (action: string) => {
    switch (action) {
      case "enforce": return <StopOutlined />;
      case "monitor": return <WarningOutlined />;
      default: return <SafetyOutlined />;
    }
  };

  const enforcedCount = logs.filter((l) => l.action_taken === "enforce").length;
  const monitoredCount = logs.filter((l) => l.action_taken === "monitor").length;

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    {
      title: "Trace ID", dataIndex: "trace_id", width: 180, ellipsis: true,
      render: (v: string) => <code className="text-xs">{v?.slice(0, 16)}...</code>,
    },
    { title: "用户 ID", dataIndex: "user_id", width: 80 },
    { title: "规则 ID", dataIndex: "rule_id", width: 80 },
    {
      title: "阶段", dataIndex: "stage", width: 90,
      render: (v: string) => <Tag color={stageColor(v)}>{v === "before" ? "输入" : "输出"}</Tag>,
    },
    {
      title: "检测实体", dataIndex: "detected_entities", width: 120,
      render: (v: string) => <Tag>{v}</Tag>,
    },
    {
      title: "处理动作", dataIndex: "action_taken", width: 100,
      render: (v: string) => <Tag color={actionColor(v)} icon={actionIcon(v)}>{v}</Tag>,
    },
    {
      title: "时间", dataIndex: "created_at", width: 170,
      render: (v: number) => new Date(v * 1000).toLocaleString(),
    },
  ];

  return (
    <div>
      <h2 className="text-lg font-bold mb-4">护栏检测日志</h2>

      <Row gutter={16} className="mb-4">
        <Col span={8}>
          <Card size="small">
            <Statistic title="总检测次数" value={logs.length} prefix={<SafetyOutlined />} />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic title="阻断次数" value={enforcedCount} prefix={<StopOutlined />} valueStyle={{ color: "#cf1322" }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic title="监控告警" value={monitoredCount} prefix={<WarningOutlined />} valueStyle={{ color: "#fa8c16" }} />
          </Card>
        </Col>
      </Row>

      <Space className="mb-4">
        <Input
          placeholder="Trace ID 搜索"
          prefix={<SearchOutlined />}
          value={filters.trace_id}
          onChange={(e) => setFilters((f) => ({ ...f, trace_id: e.target.value }))}
          allowClear
          style={{ width: 240 }}
        />
        <Select
          placeholder="阶段筛选"
          value={filters.stage || undefined}
          onChange={(v) => setFilters((f) => ({ ...f, stage: v || "" }))}
          allowClear
          style={{ width: 140 }}
          options={[
            { label: "输入检测", value: "before" },
            { label: "输出检测", value: "after" },
          ]}
        />
        <Input
          placeholder="用户 ID"
          value={filters.user_id}
          onChange={(e) => setFilters((f) => ({ ...f, user_id: e.target.value }))}
          allowClear
          style={{ width: 120 }}
        />
      </Space>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={logs}
        loading={loading}
        pagination={{ pageSize: 20, showSizeChanger: true }}
        size="middle"
      />
    </div>
  );
}
