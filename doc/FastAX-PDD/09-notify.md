> **Domain**: `domain/notify` — 通知模块 | **PRD**: FastAX-PRD/05-notify-stats-admin.md
### 5.5 通知模块 (Notify Service)

#### 5.5.1 通知通道策略 (含多语言)

| 通知类型 | 优先级 | 通道 | 多语言支持 |
|----------|--------|------|-----------|
| 验证码 | 高 | 短信/邮件 | 按用户语言发送 (LANG-03-02) |
| 安全提醒 (异地登录) | 高 | 短信 + 邮件 | 按用户语言发送 |
| 订单通知 | 中 | 站内信 | 按用户当前语言 (LANG-03-03) |
| Token 到期提醒 | 中 | 站内信 + 邮件 | 按用户设置语言 (LANG-03-01) |
| 系统公告 | 低 | 站内信 | 多语言模板 |
| 对账/账单 | 低 | 邮件 | 按用户语言 |

#### 5.5.2 通知模板设计 (含多语言)

```
数据库存储:
  - template_code, name
  - channel (sms/email/in-app)
  - language (zh-CN/en-US/ja...)
  - content (模板内容, 含占位符, 如 FreeMarker)
  - status

模板示例 (多语言):
┌──────────────────────────────────────────────┐
│ 模板: order.paid                              │
│                                               │
│ zh-CN: 尊敬的 ${userName}，您的订单            │
│        (${orderNo}) 已支付成功。               │
│                                               │
│ en: Dear ${userName}, your order              │
│     (${orderNo}) has been paid successfully.   │
│                                               │
│ ja: ${userName} 様、ご注文 (${orderNo})       │
│     が正常に支払われました。                    │
└──────────────────────────────────────────────┘

语言选择策略 (LANG-03-04):
  1. 优先使用用户设置的语言
  2. 通知类消息可附带语言参数覆盖
  3. 未找到对应语言模板 → 回退到默认语言
```

