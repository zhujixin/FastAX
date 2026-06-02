import { useEffect, useState } from "react";
import { Card, Tabs, Table, Button, Modal, Form, Input, Switch, Space, Popconfirm, message, Descriptions } from "antd";
import { PlusOutlined, ReloadOutlined, TeamOutlined, SafetyOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";

interface TeamItem {
  id: number;
  name: string;
  description?: string;
  member_count?: number;
  created_at?: string;
}

export default function EnterprisePage() {
  // SSO state
  const [ssoConfig, setSsoConfig] = useState<Record<string, unknown>>({});
  const [ssoLoading, setSsoLoading] = useState(false);
  const [ssoForm] = Form.useForm();

  // Teams state
  const [teams, setTeams] = useState<TeamItem[]>([]);
  const [teamsLoading, setTeamsLoading] = useState(false);
  const [teamModalOpen, setTeamModalOpen] = useState(false);
  const [teamForm] = Form.useForm();

  const fetchSSO = async () => {
    setSsoLoading(true);
    try {
      const res = await adminService.getSSOConfig();
      const cfg = res.data.data || {};
      setSsoConfig(cfg);
      ssoForm.setFieldsValue(cfg);
    } catch { message.error("加载 SSO 配置失败"); }
    finally { setSsoLoading(false); }
  };

  const saveSSO = async () => {
    try {
      const values = await ssoForm.validateFields();
      await adminService.updateSSOConfig(values);
      message.success("SSO 配置已保存");
      fetchSSO();
    } catch { /* validation */ }
  };

  const fetchTeams = async () => {
    setTeamsLoading(true);
    try {
      const res = await adminService.listTeams();
      setTeams((res.data.data as unknown as TeamItem[]) || []);
    } catch { message.error("加载团队列表失败"); }
    finally { setTeamsLoading(false); }
  };

  const handleCreateTeam = async () => {
    try {
      const values = await teamForm.validateFields();
      await adminService.createTeam(values);
      message.success("团队已创建");
      setTeamModalOpen(false);
      teamForm.resetFields();
      fetchTeams();
    } catch { /* validation */ }
  };

  const handleDeleteTeam = async (id: number) => {
    try {
      await adminService.deleteTeam(id);
      message.success("团队已删除");
      fetchTeams();
    } catch { message.error("删除失败"); }
  };

  useEffect(() => { fetchSSO(); fetchTeams(); }, []);

  const ssoTab = (
    <Card loading={ssoLoading}>
      <Descriptions title="SSO 单点登录配置" bordered size="small" column={1} className="mb-4">
        <Descriptions.Item label="状态">{ssoConfig.enabled ? "已启用" : "未启用"}</Descriptions.Item>
        <Descriptions.Item label="提供商">{ssoConfig.provider as string || "-"}</Descriptions.Item>
      </Descriptions>
      <Form form={ssoForm} layout="vertical" className="max-w-lg">
        <Form.Item name="enabled" label="启用 SSO" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="provider" label="SSO 提供商">
          <Form.Item name="provider" noStyle>
            <select className="border rounded px-2 py-1 w-full">
              <option value="">选择...</option>
              <option value="saml">SAML2</option>
              <option value="oidc">OpenID Connect</option>
              <option value="okta">Okta</option>
              <option value="azure">Azure AD</option>
              <option value="google">Google Workspace</option>
            </select>
          </Form.Item>
        </Form.Item>
        <Form.Item name="client_id" label="Client ID"><Input placeholder="SSO Client ID" /></Form.Item>
        <Form.Item name="client_secret" label="Client Secret"><Input.Password placeholder="SSO Client Secret" /></Form.Item>
        <Form.Item name="idp_url" label="IdP URL"><Input placeholder="https://..." /></Form.Item>
        <Button type="primary" onClick={saveSSO}>保存配置</Button>
      </Form>
    </Card>
  );

  const teamColumns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "团队名称", dataIndex: "name", width: 180 },
    { title: "描述", dataIndex: "description", ellipsis: true },
    { title: "成员数", dataIndex: "member_count", width: 80 },
    { title: "创建时间", dataIndex: "created_at", width: 170, render: (v: string) => v ? new Date(v).toLocaleString() : "-" },
    {
      title: "操作", width: 80,
      render: (_: unknown, r: TeamItem) => (
        <Popconfirm title="确定删除此团队?" onConfirm={() => handleDeleteTeam(r.id)}>
          <Button size="small" danger>删除</Button>
        </Popconfirm>
      ),
    },
  ];

  const teamsTab = (
    <div>
      <div className="flex justify-between items-center mb-4">
        <span className="font-bold">团队列表</span>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchTeams}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setTeamModalOpen(true)}>创建团队</Button>
        </Space>
      </div>
      <Table rowKey="id" columns={teamColumns} dataSource={teams} loading={teamsLoading} pagination={{ pageSize: 20 }} />
      <Modal title="创建团队" open={teamModalOpen} onOk={handleCreateTeam} onCancel={() => setTeamModalOpen(false)} width={400} destroyOnClose>
        <Form form={teamForm} layout="vertical">
          <Form.Item name="name" label="团队名称" rules={[{ required: true }]}><Input placeholder="研发团队" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">企业管理</h2>
      <Tabs
        items={[
          { key: "sso", label: <span><SafetyOutlined /> SSO 配置</span>, children: ssoTab },
          { key: "teams", label: <span><TeamOutlined /> 团队管理</span>, children: teamsTab },
        ]}
      />
    </div>
  );
}
