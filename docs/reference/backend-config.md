# 后端配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | 任务工作目录根路径 |
| `LSI_LOG_LEVEL` | `info` | 结构化日志级别 |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | 下载阶段超时时间 |
| `LSI_WHISPER_MODEL` | `small` | 传给 `whisper-cli --model` 的模型名或本地模型目录 |
| `LSI_WHISPER_COMPUTE_TYPE` | `default` | 传给 `whisper-cli --compute-type` 的计算类型 |
| `HF_TOKEN` | 空 | Hugging Face 可选 token，用于提高模型下载限额 |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API origin |
| `LSI_LLM_API_KEY` | 空 | 翻译 provider API key |
| `LSI_LLM_MODEL` | 空 | 翻译模型名 |
| `LSI_LLM_TIMEOUT` | `2m` | 单段翻译超时时间 |

## 运行时工具要求

本地启动真实后端时，`yt-dlp`、`ffmpeg`、`whisper-cli` 必须在 `PATH` 中。Docker 镜像已包含这些运行时工具。
