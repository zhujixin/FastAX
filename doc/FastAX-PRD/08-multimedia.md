> **Domain**: `domain/proxy` (adaptor/) — 多模态支持 | **PDD**: §5.10

### 6.11 多模态支持（MEDIA）

参考 CrazyRouter、35Gateway、Portkey 等平台的多模态聚合能力。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| MEDIA-01 | **图片生成统一 API**：聚合 DALL-E、Midjourney、Flux、Stable Diffusion 等，统一 `POST /v1/images/generations` 端点 | P1 | 支持 failover 切换（DALL-E → Midjourney） |
| MEDIA-02 | **视频生成支持**：聚合 Sora、Kling、Runway、Veo 等视频生成 API，统一请求格式 | P2 | — |
| MEDIA-03 | **语音合成/识别统一 API**：聚合 OpenAI TTS/STT、ElevenLabs、Azure Speech 等，统一 `POST /v1/audio/speech`、`/v1/audio/transcriptions` 端点 | P1 | — |
| MEDIA-04 | **音乐生成支持**：聚合 Suno、Udio 等音乐生成 API | P2 | — |
| MEDIA-05 | **多模态路由与故障转移**：同模态多供应商间的自动重试和切换（如图片生成 A 失败→B） | P1 | — |
| MEDIA-06 | **多模态成本追踪**：按模态类型（文本/图片/视频/音频）独立统计消耗和费用 | P1 | — |
| MEDIA-07 | **媒体内容审核**：生成的图片/视频/音频自动过安全审核，违规内容标记/拦截 | P1 | — |

