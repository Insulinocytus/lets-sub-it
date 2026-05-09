# 后端 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕生成任务 |
| `GET` | `/jobs/:id` | 查询任务状态 |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | 查询指定视频和目标语言的最新任务 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询已完成字幕产物 |
| `GET` | `/subtitle-files/:jobId/:mode` | 获取指定任务的 VTT 文件内容 |

## `POST /jobs`

请求 JSON：

```json
{
  "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "sourceLanguage": "en",
  "targetLanguage": "zh"
}
```

## 字幕文件模式

`GET /subtitle-files/:jobId/:mode` 的 `mode` 可用值：

- `source`
- `translated`
- `bilingual`

## 安全边界

API 响应不得暴露服务端本地绝对文件路径。前端只能通过 `/subtitle-files/:jobId/:mode` 获取字幕文件内容。
