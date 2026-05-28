> **Domain**: `domain/proxy` (adaptor) — 多模态支持 | **PRD**: FastAX-PRD/08-multimedia.md
### 5.10 多模态支持 (PRD §6.11 MEDIA)

#### 5.10.1 多模态统一 API 架构

```
POST /v1/images/generations      ──→  Adaptor 分发
POST /v1/audio/speech            ──→    │
POST /v1/audio/transcriptions    ──→    │
POST /v1/video/generations       ──→    │
                                        ▼
                              ┌──────────────────┐
                              │  模态路由引擎      │
                              │                   │
                              │  1. 按模态筛选     │
                              │  2. 按优先级分组   │
                              │  3. 权重随机选择   │
                              │  4. 故障转移       │
                              └──────────────────┘
```

| 模态 | 统一端点 | 供应商适配器 |
|------|---------|-------------|
| 图片生成 | `POST /v1/images/generations` | DALL-E, Midjourney, Flux, Stable Diffusion |
| 语音合成 | `POST /v1/audio/speech` | OpenAI TTS, ElevenLabs, Azure Speech |
| 语音识别 | `POST /v1/audio/transcriptions` | Whisper, Azure STT |
| 视频生成 | `POST /v1/video/generations` | Sora, Kling, Runway (P2) |

#### 5.10.2 多模态成本追踪

```
call_log 扩展:
  media_type TEXT      -- text/image/audio/video
  media_cost REAL      -- 按媒体类型独立计费
  media_meta TEXT      -- 媒体元数据 JSON (分辨率/时长/格式)
```

