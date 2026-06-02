import { useState } from "react";
import { Card, Row, Col, Input, Button, Statistic, App } from "antd";
import { GiftOutlined, DollarOutlined } from "@ant-design/icons";
import api from "@/shared/utils/axios";
import { showSuccess, showError } from "@/shared/utils";

export default function TopUpPage() {
  const { message } = App.useApp();
  const [redemptionCode, setRedemptionCode] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [quota, setQuota] = useState(0);

  // Load current quota
  useState(() => {
    api.get("/tokens/my").then((res) => {
      const tokens = res.data.data || [];
      const total = tokens.reduce((s: number, t: { remaining?: number }) => s + (t.remaining || 0), 0);
      setQuota(total);
    }).catch(() => {});
  });

  const handleRedeem = async () => {
    if (!redemptionCode.trim()) {
      message.warning("请输入兑换码");
      return;
    }
    setSubmitting(true);
    try {
      const res = await api.post("/tokens/buy", { key: redemptionCode.trim() });
      const added = res.data.data?.quantity || 0;
      setQuota((prev) => prev + added);
      setRedemptionCode("");
      showSuccess(`兑换成功！获得 ${added.toLocaleString()} tokens`);
    } catch (err) {
      showError(err);
    } finally {
      setSubmitting(false);
    }
  };

  const handlePaste = async () => {
    try {
      const text = await navigator.clipboard.readText();
      setRedemptionCode(text);
    } catch {
      message.warning("无法读取剪贴板");
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-6">Token 充值</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card>
            <Statistic title="当前余额" value={quota} suffix="tokens" prefix={<DollarOutlined />} />
            <p className="text-gray-400 text-sm mt-2">前往商城购买更多 Token</p>
            <Button type="primary" className="mt-4" onClick={() => window.open("/tokens/buy", "_blank")}>
              前往购买
            </Button>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={<><GiftOutlined /> 兑换码充值</>}>
            <Input.TextArea
              value={redemptionCode}
              onChange={(e) => setRedemptionCode(e.target.value)}
              onPaste={(e) => { e.preventDefault(); setRedemptionCode(e.clipboardData.getData("text")); }}
              placeholder="请输入兑换码"
              rows={3}
              className="mb-3"
            />
            <div className="flex gap-2">
              <Button onClick={handlePaste}>粘贴</Button>
              <Button type="primary" onClick={handleRedeem} loading={submitting}>
                {submitting ? "兑换中..." : "立即兑换"}
              </Button>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
