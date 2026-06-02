import { z } from "zod";

export const loginSchema = z.object({
  account: z
    .string()
    .min(1, "请输入手机号或邮箱")
    .refine(
      (v) => /^1\d{10}$/.test(v) || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
      "请输入有效的手机号或邮箱"
    ),
  password: z.string().min(6, "密码至少 6 位").max(64, "密码过长"),
});

export const registerSchema = z.object({
  username: z.string().min(2, "用户名至少 2 个字符").max(32, "用户名过长"),
  email: z.string().email("请输入有效邮箱").optional().or(z.literal("")),
  phone: z.string().regex(/^1\d{10}$/, "请输入有效手机号").optional().or(z.literal("")),
  password: z.string().min(6, "密码至少 6 位").max(64, "密码过长"),
});

export const buyTokenSchema = z.object({
  product_id: z.number().positive("请选择产品"),
  quantity: z.number().min(1000, "最少 1000 tokens").max(10000000, "超出限额"),
  payment_method: z.enum(["wechat", "alipay", "stripe"], { message: "请选择支付方式" }),
});

export const vendorRegisterSchema = z.object({
  company_name: z.string().min(1, "请输入公司名称"),
  contact_name: z.string().min(1, "请输入联系人"),
  contact_email: z.string().email("请输入有效邮箱"),
  contact_phone: z.string().min(1, "请输入联系电话"),
  api_base_url: z.string().url("请输入有效 URL"),
  api_auth_type: z.string().min(1, "请选择认证方式"),
});

export type LoginFormData = z.infer<typeof loginSchema>;
export type RegisterFormData = z.infer<typeof registerSchema>;
export type BuyTokenFormData = z.infer<typeof buyTokenSchema>;
