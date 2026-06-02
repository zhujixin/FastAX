import { useEffect, useState } from "react";
import { Card, Table, Tag, Button, Space, Modal, Descriptions, Popconfirm, message } from "antd";
import { CheckOutlined, StopOutlined, ReloadOutlined } from "@ant-design/icons";
import { adminService } from "@/shared/api/admin";
import type { VendorInfo } from "@/shared/api/types";

export default function VendorListPage() {
  const [vendors, setVendors] = useState<VendorInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedVendor, setSelectedVendor] = useState<VendorInfo | null>(null);

  const fetchVendors = async () => {
    setLoading(true);
    try {
      const res = await adminService.listVendors();
      setVendors(res.data.data || []);
    } catch {
      message.error("加载供应商列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchVendors(); }, []);

  const handleReview = async (id: number, approved: boolean) => {
    try {
      await adminService.reviewVendor(id, approved);
      message.success(approved ? "已通过审核" : "已拒绝");
      fetchVendors();
    } catch {
      message.error("审核操作失败");
    }
  };

  const handleSuspend = async (id: number) => {
    try {
      await adminService.suspendVendor(id);
      message.success("已冻结");
      fetchVendors();
    } catch {
      message.error("操作失败");
    }
  };

  const handleDetail = async (id: number) => {
    try {
      const res = await adminService.getVendor(id);
      setSelectedVendor(res.data.data as VendorInfo);
      setDetailOpen(true);
    } catch {
      message.error("获取供应商详情失败");
    }
  };

  const statusMap: Record<string, { color: string; text: string }> = {
    approved: { color: "green", text: "已入驻" },
    pending: { color: "orange", text: "待审核" },
    rejected: { color: "red", text: "已拒绝" },
    suspended: { color: "default", text: "已冻结" },
  };

  const columns = [
    { title: "ID", dataIndex: "id", width: 60 },
    { title: "公司", dataIndex: "company_name", width: 180 },
    { title: "联系人", dataIndex: "contact_name", width: 100 },
    { title: "邮箱", dataIndex: "contact_email", width: 200 },
    {
      title: "状态", dataIndex: "status", width: 80,
      render: (s: string) => {
        const m = statusMap[s] || { color: "default", text: s };
        return <Tag color={m.color}>{m.text}</Tag>;
      },
    },
    {
      title: "商品", dataIndex: "products_count", width: 60,
    },
    { title: "入驻时间", dataIndex: "created_at", width: 130, render: (v: string) => v ? new Date(v).toLocaleDateString() : "-" },
    {
      title: "操作", width: 200,
      render: (_: unknown, r: VendorInfo) => (
        <Space>
          <Button size="small" onClick={() => handleDetail(r.id)}>详情</Button>
          {r.status === "pending" && (
            <>
              <Popconfirm title="确认通过?" onConfirm={() => handleReview(r.id, true)}>
                <Button size="small" type="primary" icon={<CheckOutlined />}>通过</Button>
              </Popconfirm>
              <Popconfirm title="确认拒绝?" onConfirm={() => handleReview(r.id, false)}>
                <Button size="small" danger>拒绝</Button>
              </Popconfirm>
            </>
          )}
          {r.status === "approved" && (
            <Popconfirm title="确认冻结?" onConfirm={() => handleSuspend(r.id)}>
              <Button size="small" danger icon={<StopOutlined />}>冻结</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">供应商管理</h2>
        <Button icon={<ReloadOutlined />} onClick={fetchVendors}>刷新</Button>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={vendors} loading={loading} pagination={{ pageSize: 20 }} />
      </Card>

      <Modal title="供应商详情" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={500}>
        {selectedVendor && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="ID">{selectedVendor.id}</Descriptions.Item>
            <Descriptions.Item label="公司">{selectedVendor.company_name}</Descriptions.Item>
            <Descriptions.Item label="联系人">{selectedVendor.contact_name}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{selectedVendor.contact_email}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={statusMap[selectedVendor.status]?.color}>{statusMap[selectedVendor.status]?.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="商品数">{selectedVendor.products_count}</Descriptions.Item>
            <Descriptions.Item label="入驻时间">{selectedVendor.created_at ? new Date(selectedVendor.created_at).toLocaleString() : "-"}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
