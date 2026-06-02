import { useEffect, useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card, Form, Select, InputNumber, Button, App } from "antd";
import { useNavigate } from "react-router-dom";
import { tokenService } from "@/shared/api";
import { buyTokenSchema, type BuyTokenFormData } from "@/shared/utils/validations";
import type { TokenProduct } from "@/shared/api/types";

export default function BuyTokenPage() {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [products, setProducts] = useState<TokenProduct[]>([]);

  const { control, handleSubmit, formState: { errors, isSubmitting } } = useForm<BuyTokenFormData>({
    resolver: zodResolver(buyTokenSchema),
    defaultValues: { product_id: 0, quantity: 10000, payment_method: "wechat" },
  });

  useEffect(() => {
    tokenService.getProducts()
      .then((res) => setProducts(res.data.data || []))
      .catch(() => {});
  }, []);

  const onFinish = async (values: BuyTokenFormData) => {
    try {
      const res = await tokenService.buy(values);
      const data = res.data.data;
      message.success("订单已创建");
      if (data?.payment_url) {
        window.open(data.payment_url, "_blank");
      }
      navigate("/orders");
    } catch (err: any) {
      message.error(err.response?.data?.message || "下单失败");
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <Card title="购买 Token">
        <form onSubmit={handleSubmit(onFinish)}>
          <Form.Item label="产品" validateStatus={errors.product_id ? "error" : ""} help={errors.product_id?.message}>
            <Controller name="product_id" control={control} render={({ field }) => (
              <Select
                {...field}
                size="large"
                options={products.map((p) => ({
                  value: p.id,
                  label: `${p.name} — ¥${p.price}/1K tokens`,
                }))}
                placeholder="选择产品"
              />
            )} />
          </Form.Item>
          <Form.Item label="数量 (tokens)" validateStatus={errors.quantity ? "error" : ""} help={errors.quantity?.message}>
            <Controller name="quantity" control={control} render={({ field }) => (
              <InputNumber {...field} size="large" min={1000} max={10000000} step={1000} className="w-full" />
            )} />
          </Form.Item>
          <Form.Item label="支付方式" validateStatus={errors.payment_method ? "error" : ""} help={errors.payment_method?.message}>
            <Controller name="payment_method" control={control} render={({ field }) => (
              <Select {...field} size="large" options={[
                { value: "wechat", label: "微信支付" },
                { value: "alipay", label: "支付宝" },
                { value: "stripe", label: "Stripe (国际)" },
              ]} />
            )} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={isSubmitting}>
              确认购买
            </Button>
          </Form.Item>
        </form>
      </Card>
    </div>
  );
}
