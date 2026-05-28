> **Domain**: `shared/i18n` — 多语言国际化 + 海外→国内优化 | **PRD**: FastAX-PRD/06-i18n.md | **参考**: one-api common/i18n/
### 5.7 多语言模块设计 (PRD §6.9)

#### 5.7.1 语言配置管理 (PRD LANG-04)

```
管理后台 → 多语言配置:

  supported_languages 表驱动:
  ┌──────────┬──────────┬──────────────┬───────────┬───────┐
  │ locale   │ name     │ native_name  │ is_enabled │ is_default │
  ├──────────┼──────────┼──────────────┼───────────┼───────┤
  │ zh-CN    │ Chinese  │ 中文         │ 1         │ 1      │
  │ en       │ English  │ English      │ 1         │ 0      │
  │ ja       │ Japanese │ 日本語       │ 0         │ 0      │
  │ ko       │ Korean   │ 한국어       │ 0         │ 0      │
  └──────────┴──────────┴──────────────┴───────────┴───────┘

  管理员操作:
  - 启用/禁用语种 (F-ADM-08)
  - 设置默认语言
  - 配置回退语言链
  - 前端语言选择器仅显示已启用语种 (LANG-04-02)
```

#### 5.7.2 多语言响应处理 (PRD LANG-05-03)

```
API 多语言内容返回策略:

  场景: 商品详情、分类名称等动态内容

  方案 1 — 多字段存储 (推荐):
    token_product 表:
    - name (默认语言, zh-CN)
    - name_i18n TEXT: {"en":"GPT-4 Token Pack", "ja":"GPT-4 トークンパック"}

    API 响应:
    GET /api/tokens/products/1
    {
      "id": 1,
      "name": "GPT-4 Token 包",             // 默认语言
      "name_i18n": {                         // 多语言字段
        "en": "GPT-4 Token Pack",
        "ja": "GPT-4 トークンパック"
      },
      ...
    }

    前端按当前语言渲染: data.name_i18n[language] || data.name

  方案 2 — Accept-Language 驱动:
    GET /api/tokens/products/1
    Header: Accept-Language: en
    Response: { "name": "GPT-4 Token Pack", ... }
```

### 5.8 海外→国内模型优化设计 (PRD §6.2.4-6.2.5)

#### 5.8.1 双轨架构

```
                      ┌────────────────────────┐
                      │    FastAX Proxy         │
                      │    (Global Deployment)  │
                      └──────┬──────────┬───────┘
                             │          │
                    ┌────────▼──┐  ┌───▼────────┐
                    │ 海外节点    │  │ 国内节点     │
                    │ (US/EU)   │  │ (CN)       │
                    │           │  │            │
                    │ 海外→海外  │  │ 海外→国内   │
                    │ OpenAI    │  │ DeepSeek   │
                    │ Claude    │  │ Qwen       │
                    │ Gemini    │  │ GLM        │
                    └───────────┘  └────────────┘

海外用户访问国内模型路径:
  用户 → 海外节点 → 专线/优化链路 → 国内节点 → 国内供应商
  (OCN-01: 目标延迟 ≤ 1500ms)
```

#### 5.8.2 跨境合规展示 (PRD OCN-05, R-COMPL-09~12)

```
海外用户购买国内模型 Token 时的合规流程:

  1. 用户选择国内模型商品 → 加入购物车
  2. 系统检测用户为海外用户 (IP/注册地)
  3. 弹出合规提示:
     - 数据跨境说明: "本产品由国内供应商提供，您的数据将被传输至中国境内处理"
     - 适用法律法规: 《生成式人工智能服务管理暂行办法》
     - 用户须知: 使用限制和免责条款
  4. 用户确认后 → 继续购买流程
  5. 购买记录标记 "跨境交易" 标识

  合规文案多语言 (LANG-05):
  - 英文版用户协议 (R-COMPL-11)
  - 英文版隐私政策
  - 英文版合规说明
```

