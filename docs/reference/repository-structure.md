# 仓库结构

```text
lets-sub-it/
├── backend/              # Go 后端 HTTP API、SQLite/GORM 持久化和真实任务 runner
├── whisper/              # Python whisper-cli，封装 faster-whisper 并输出 WebVTT
├── extension/            # Chrome MV3 扩展，负责弹窗、后台消息网关和 YouTube 字幕叠加
├── docs/                 # PRD、规格、计划和 reference 文档
├── docker-compose.yml    # 本地 Docker 后端编排
└── mise.toml             # Go、Python、Node.js 和 uv 工具链版本
```

## 关键入口

| 路径 | 用途 |
| --- | --- |
| `backend/cmd/server/main.go` | HTTP server entry point |
| `backend/internal/app/` | env config, app wiring, real-mode tool checks |
| `backend/internal/api/` | routes, handlers, CORS, response mapping |
| `backend/internal/store/` | GORM models and SQLite persistence |
| `backend/internal/runner/` | real runner, download, translation, VTT packaging |
| `whisper/src/whisper_cli/cli.py` | command-line contract and exit codes |
| `whisper/src/whisper_cli/transcribe.py` | faster-whisper adapter |
| `whisper/src/whisper_cli/vtt.py` | cue validation and WebVTT rendering |
| `extension/entrypoints/background.ts` | runtime message gateway |
| `extension/entrypoints/youtube.content.ts` | YouTube page integration |
| `extension/entrypoints/popup/` | popup UI entry |
