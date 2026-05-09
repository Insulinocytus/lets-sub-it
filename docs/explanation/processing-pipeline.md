# 处理链路

从用户提交 URL 到 YouTube 页面渲染字幕，系统按同一条 job 链路推进。这个链路把浏览器交互、后端持久化、外部工具执行和页面 overlay 解耦，便于定位失败发生在哪个阶段。

1. 用户在 extension popup 中提交 YouTube 公开视频 URL，选择源语言和目标语言。
2. Popup 通过 background service worker 发送 `job:create` 消息。
3. Background 调用 Go API 的 `POST /jobs`，后端解析 `videoId` 并按 `videoId + targetLanguage` 查找可复用的完成结果或进行中任务。
4. 需要新任务时，后端创建 `Job` 记录，状态从 `queued` 开始，并把工作目录记录到 SQLite。
5. Embedded runner 在后台执行任务，依次进入 `downloading`、`transcribing`、`translating` 和 `packaging`。
6. Popup 通过 background 轮询任务状态，向用户展示当前阶段和错误摘要。
7. 任务完成后，background 查询字幕资产并写入扩展本地缓存，缓存键包含 `backendBaseUrl + videoId + targetLanguage`。
8. Content script 在 YouTube watch 页面识别当前 `videoId`，通过 background 解析可用字幕资产。
9. Content script 获取 `translated` 或 `bilingual` VTT 内容，并用 Shadow DOM overlay 在播放页渲染字幕。

## 阶段

| 阶段 | 执行者 | 完成条件 |
| --- | --- | --- |
| `downloading` | `yt-dlp` + `ffmpeg` | 下载子进程成功退出；工作目录中存在目标媒体文件；后端可以读取最终文件路径。 |
| `transcribing` | `whisper-cli` | `whisper-cli` 子进程退出码为 `0`；`source.vtt` 文件存在；`source.vtt` 可解析且 cue 数量大于 0。 |
| `translating` | OpenAI-compatible LLM | 所有 source cues 都产出对应目标语言文本；每个 cue 的翻译请求携带最多前后各 10 条上下文 cue；LLM 返回 JSON 格式 `{"translation": "..."}`，翻译数量与 source cues 一致；`translated.vtt` 写入成功。 |
| `packaging` | 后端打包字幕资产 | 生成 `bilingual.vtt`；创建 `SubtitleAsset` 记录；`source.vtt`、`translated.vtt`、`bilingual.vtt` 均可通过文件服务按 mode 访问。 |

## 当前限制

当前 MVP 不会在服务重启后自动恢复进行中的 runner。当前 `POST /jobs` 会复用非 `failed` 的旧 job，所以服务重启后的卡住任务通常需要先清理或标记失败，才能用同一 `videoId + targetLanguage` 创建新任务。

状态、字段和复用键见 [数据与任务模型](../reference/data-and-job-model.md)。
