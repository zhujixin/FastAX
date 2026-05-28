## 18. 附录：市场竞争对标分析

### 17.1 对标维度总览

| 功能维度 | OpenRouter | New-API/One API | LiteLLM | Portkey | VoAPI | **FastAX 目标** |
|---------|-----------|-----------------|---------|--------|-------|----------------|
| 模型数 | 300+ | 25+ | 100+ | 1600+(BYOK) | 可配置 | 200+ (平台)+BYOK |
| 多模型路由 | ✅ 加权 | ✅ 权重随机 | ✅ 多策略 | ✅ LB+Fallback | ✅ 高级 | ✅ 加权+成本+延迟 |
| 多协议原生 | ✅ 全协议 | ❌ 仅 OpenAI | ✅ 多协议 | ✅ 多协议 | ❌ 仅 OpenAI | ✅ OpenAI+Anthropic+Gemini |
| 多模态 | ❌ | ✅ Midjourney/Suno | ❌ | ✅ Vision/Audio | ❌ | ✅ 图片/视频/音频/音乐 |
| 安全护栏 | ❌ | ❌ | ✅+Akto | ✅ 插件市场 | ❌ | ✅ PII/注入/内容审核 |
| BYOK | ✅ 5%费 | ❌ | ❌ | ✅ BYOK | ❌ | ✅ BYOK ≤ 5% |
| 语义缓存 | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ Phase 7 |
| 企业 SSO | ❌ | ❌ | ✅ | ✅ SOC 2 | ❌ | ✅ SAML/OIDC |
| 供应商入驻 | ✅ 市场 | ❌ | ❌ | ❌ | ❌ | ✅ 入驻+结算 |
| 插件系统 | ❌ | ❌ | ✅ 回调 | ✅ 市场 | ❌ | ✅ 中间件流水线 |
| 开源/自部署 | ❌ | ✅ 开源 | ✅ MIT | ❌ | ❌ | ❌ SaaS 优先 |
| 国内模型 | ⚠️ 高延迟 | ✅ 优秀 | ⚠️ 有限 | ⚠️ 有限 | ✅ 优秀 | ✅ 双轨专线 |
| 国内直连 | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |

### 17.2 核心差异化策略

| 差异化方向 | FastAX 优势 | 对标平台差距 |
|-----------|-------------|-------------|
| **双轨专线+国内模型** | 海外节点→国内专线→国内供应商，P95 ≤ 1500ms | OpenRouter 访问国内延迟 300ms+ |
| **供应商开放市场** | 第三方厂家入驻+自主定价+结算提现 | OpenRouter 仅聚合，供应商不可入驻 |
| **多协议原生+多模态** | 原生支持 Anthropic/Gemini/OpenAI+图片/视频/音频/音乐 | New-API 仅 OpenAI 协议；OpenRouter 仅文本 |
| **安全护栏全栈** | PII+注入+内容审核+密钥扫描+不可篡改审计+合规导出 | OpenRouter/New-API 无护栏功能 |
| **成本优化多引擎** | Prompt Caching+语义缓存+成本感知路由+模型回退链+预算封顶 | Portkey 支持缓存但无语义缓存；LiteLLM 无模型回退链 |
| **企业功能完善** | SSO+团队隔离+BYOK+审计导出+数据驻留 | VoAPI 有计费但无 SSO；One API 无企业功能 |

### 17.3 竞品参照列表

| 平台 | 定位 | 参考价值 | 网址 |
|------|------|----------|------|
| **OpenRouter** | 全球最大模型聚合市场 | 商业模式、路由策略、模型变体 | openrouter.ai |
| **New-API** | 开源 API 中转站 (19K+ Stars) | 渠道管理、模型映射、计费体系 | github.com/Calcium-Ion/new-api |
| **LiteLLM** | 开源企业代理网关 (41K+ Stars) | 多策略路由、护栏系统、企业功能 | github.com/BerriAI/litellm |
| **Portkey** | 企业级 AI 网关 | 护栏插件市场、语义缓存、BYOK | portkey.ai |
| **One API** | 开源 API 聚合 (30K+ Stars) | 轻量部署、国内模型支持 | github.com/songquanpeng/one-api |
| **VoAPI** | 商业分发平台 | 多级计费、分层用户管理 | voapi.cn |
| **LLMProxy** | 安全优先代理 (开源) | 18+ 插件流水线、6 层防御 | github.com/fabriziosalmi/llmproxy |
| **SignalVault** | 安全审计代理 | 加密审计日志、PII 检测、预算控制 | peerpush.net/p/signalvault |
| **CrazyRouter** | 最低价聚合 | 627+ 模型、多模态、~45% 官方价 | crazyrouter.com |
| **35Gateway** | 开源多模态网关 | 图片/视频/音频/音乐统一 API | github.com/guo2001china/35gateway |

---

*文档结束*
