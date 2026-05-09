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
| `backend/cmd/server/main.go` | HTTP 服务入口 |
| `backend/internal/app/` | 环境配置、应用装配、真实模式工具检查 |
| `backend/internal/api/` | 路由、处理器、CORS、响应映射 |
| `backend/internal/store/` | GORM 模型和 SQLite 持久化 |
| `backend/internal/runner/` | 真实 runner、下载、翻译、VTT 打包 |
| `whisper/src/whisper_cli/cli.py` | 命令行契约和退出码 |
| `whisper/src/whisper_cli/transcribe.py` | faster-whisper 适配层 |
| `whisper/src/whisper_cli/vtt.py` | cue 校验和 WebVTT 渲染 |
| `extension/entrypoints/background.ts` | runtime 消息网关 |
| `extension/entrypoints/youtube.content.ts` | YouTube 页面集成 |
| `extension/entrypoints/popup/` | popup UI 入口 |
