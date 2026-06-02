import { useEffect, useState } from "react";
import { Card, Table, Button, Input, Space, Tag, Select, Popconfirm, App } from "antd";
import { SearchOutlined, PlusOutlined, ReloadOutlined, CopyOutlined, EyeOutlined, EyeInvisibleOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";
import { showSuccess, showError, copy } from "@/shared/utils";
import { renderStatus, renderResponseTime, renderQuota } from "@/shared/utils/render";
import { ITEMS_PER_PAGE, CHANNEL_TYPES } from "@/shared/constants";

interface Channel {
  id: number;
  name: string;
  type: string;
  key: string;
  base_url: string;
  models: string;
  status: number;
  priority: number;
  balance: number;
  response_time: number;
  test_model: string;
}

export default function ChannelManagePage() {
  const { message } = App.useApp();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [showDetail, setShowDetail] = useState(false);
  const [showKeys, setShowKeys] = useState<Set<number>>(new Set());

  const loadChannels = async () => {
    setLoading(true);
    try {
      const res = await api.get("/admin/suppliers");
      setChannels(res.data.data || []);
    } catch (err) {
      showError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadChannels(); }, []);

  const handleToggle = async (id: number, status: number) => {
    try {
      const newStatus = status === 1 ? 2 : 1;
      await api.put(`/admin/channels/${id}/status`, { status: newStatus });
      setChannels((prev) => prev.map((c) => (c.id === id ? { ...c, status: newStatus } : c)));
      showSuccess("操作成功");
    } catch (err) { showError(err); }
  };

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/suppliers/${id}`);
      setChannels((prev) => prev.filter((c) => c.id !== id));
      showSuccess("已删除");
    } catch (err) { showError(err); }
  };

  const handleCopyKey = async (key: string) => {
    const ok = await copy(key);
    ok ? showSuccess("已复制") : message.warning("复制失败");
  };

  const toggleShowKey = (id: number) => {
    const next = new Set(showKeys);
    next.has(id) ? next.delete(id) : next.add(id);
    setShowKeys(next);
  };

  const renderType = (type: string) => {
    const info = CHANNEL_TYPES[type.toLowerCase()];
    return info ? <Tag color={info.color}>{info.label}</Tag> : <Tag>{type}</Tag>;
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 50 },
    { title: "名称", dataIndex: "name" },
    { title: "类型", dataIndex: "type", render: renderType },
    { title: "状态", dataIndex: "status", render: (_: unknown, r: Channel) => renderStatus(r.status) },
    {
      title: "Key", dataIndex: "key",
      render: (_: unknown, r: Channel) => (
        <Space size={4}>
          <code className="text-xs bg-gray-100 px-1 rounded max-w-[120px] truncate inline-block">
            {showKeys.has(r.id) ? r.key : r.key.slice(0, 8) + "***"}
          </code>
          <Button size="small" type="text" icon={showKeys.has(r.id) ? <EyeInvisibleOutlined /> : <EyeOutlined />} onClick={() => toggleShowKey(r.id)} />
          <Button size="small" type="text" icon={<CopyOutlined />} onClick={() => handleCopyKey(r.key)} />
        </Space>
      ),
    },
    { title: "响应时间", dataIndex: "response_time", render: (_: unknown, r: Channel) => renderResponseTime(r.response_time || 0) },
    { title: "余额", dataIndex: "balance", render: (_: unknown, r: Channel) => renderQuota(r.balance) },
    ...(showDetail ? [
      { title: "优先级", dataIndex: "priority" },
      { title: "测试模型", dataIndex: "test_model" },
    ] : []),
    {
      title: "操作", key: "action",
      render: (_: unknown, r: Channel) => (
        <Space size="small">
          <Button size="small" onClick={() => handleToggle(r.id, r.status)}>
            {r.status === 1 ? "禁用" : "启用"}
          </Button>
          <Popconfirm title="确认删除此渠道？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">渠道管理</h2>
        <Space>
          <Input prefix={<SearchOutlined />} placeholder="搜索渠道" value={searchKeyword} onChange={(e) => setSearchKeyword(e.target.value)} onPressEnter={() => loadChannels()} className="w-48" />
          <Button icon={showDetail ? <EyeInvisibleOutlined /> : <EyeOutlined />} onClick={() => setShowDetail(!showDetail)}>
            {showDetail ? "隐藏" : "显示"}详情
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadChannels}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />}>新增渠道</Button>
        </Space>
      </div>
      <Card>
        <Table
          dataSource={channels.filter((c) => !searchKeyword || c.name.includes(searchKeyword))}
          columns={columns}
          rowKey="id"
          loading={loading}
          size="middle"
          pagination={{ pageSize: ITEMS_PER_PAGE, showTotal: (t: number) => `共 ${t} 条` }}
        />
      </Card>
    </div>
  );
}
