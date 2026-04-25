# AGENTS.md

## 项目概览

Lets Sub It 是一个自托管的 YouTube 字幕生成与翻译工具。目标链路是：提交 YouTube 公开视频链接，下载音频，本地转写，翻译，生成字幕，并由播放页加载字幕。

当前仓库处于 MVP 阶段，已经有两个可运行部分：

- `backend/`：Go 1.22 mock API server，提供真实 HTTP API、SQLite 持久化、job 复用、mock runner 状态推进和 VTT 字幕文件服务。
- `whisper/`：Python 3.12 包 `whisper-cli`，通过 `faster-whisper` 把本地音频文件转成经过校验的 WebVTT 字幕。

目标架构仍包括后续的 Chrome extension、真实 `yt-dlp` 下载、`ffmpeg` 音频处理、backend 调用 `whisper-cli`、OpenAI-compatible LLM 翻译，以及 `translated.vtt` / `bilingual.vtt` 打包。当前 backend 只模拟这些外部能力，不会访问 YouTube、调用 LLM 或执行真实转写。

主要目录：

- `backend/`：Go module `lets-sub-it-api`，入口在 `cmd/server/main.go`，核心代码位于 `internal/api/`、`internal/store/`、`internal/runner/`、`internal/app/`。
- `whisper/`：Python 包 `whisper-cli`，源码在 `src/whisper_cli/`，测试在 `tests/`。
- `docs/`：PRD、规格说明与实施计划。该目录下另有 `docs/AGENTS.md`，要求文档正文使用中文。
- `extension/`：Chrome extension 预留目录，目前只跟踪 `.gitignore`，不要假设已有可运行前端项目。
- `mise.toml`：本地工具链版本，包含 Go 1.22、Python 3.12 与 uv。

## 协作原则

- 始终使用简体中文与用户沟通。
- 优先做最小可行改动，不添加未被要求的功能、抽象或配置。
- 修改现有代码时保持手术式变更：只触碰与任务直接相关的文件和行。
- 如果需求存在多种解释，先说明假设；不确定且风险较高时先询问。
- 不要清理无关代码、重排无关格式或重构未被要求的模块。
- 不要覆盖用户或其他 agent 已做的未提交改动；开始前和结束前用 `git status --short` 核对工作树。

## 环境与依赖

AI agent 的 shell 环境不会自动执行 `eval "$(mise activate zsh)"`。凡是使用 `mise.toml` 中锁定版本的工具，都必须通过 `mise exec --` 前缀调用，例如 `mise exec -- go`、`mise exec -- uv`、`mise exec -- python`。

从仓库根目录安装工具链：

```bash
mise install
```

同步 backend 依赖：

```bash
cd backend
mise exec -- go mod download
```

同步 whisper 依赖：

```bash
cd whisper
mise exec -- uv sync --dev
```

真实转写会调用 `faster-whisper`，可能需要模型下载能力以及本机可用的推理运行环境。单元测试使用 fake model，不会下载或运行真实 Whisper 模型。

## 开发工作流

启动本地 mock API server：

```bash
cd backend
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

快速验证 `POST /jobs`：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "sourceLanguage": "ja",
    "targetLanguage": "zh-Hans"
  }'
```

backend 配置：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | job 工作目录根路径 |

运行本地 CLI 示例：

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

成功时 stdout 输出 JSON，`--output` 写入 WebVTT 文件。CLI 退出码契约：

- `0`：成功。
- `2`：输入校验失败，例如输入文件不存在、模型名或语言无效。
- `3`：转写失败。
- `4`：输出校验失败，例如无法生成合法 VTT。

## 测试说明

运行 backend 全部测试：

```bash
cd backend
mise exec -- go test ./...
```

运行 backend 单个包测试：

```bash
cd backend
mise exec -- go test ./internal/api
```

运行 whisper 全部测试：

```bash
cd whisper
mise exec -- uv run pytest
```

运行 whisper 单个测试文件：

```bash
cd whisper
mise exec -- uv run pytest tests/test_vtt.py
```

按测试名聚焦运行：

```bash
cd whisper
mise exec -- uv run pytest -k "vtt"
```

测试规则：

- backend 测试文件与被测包同目录，命名为 `*_test.go`。
- whisper 的 pytest 配置在 `whisper/pyproject.toml`：`pythonpath = ["src"]`，默认 `addopts = "-q"`。测试文件命名遵循 `test_*.py`。
- 改动 `backend/internal/` 行为时，优先添加或更新同包 Go 测试，然后确认相关 `go test` 通过。
- 改动 `whisper/src/whisper_cli/` 行为时，优先添加或更新相邻 pytest 测试，然后确认相关 pytest 通过。
- 不要让单元测试依赖网络、真实 YouTube、模型下载、本地 GPU、真实音频样本或外部 LLM API。

## 构建说明

验证 backend 可构建：

```bash
cd backend
mise exec -- go build ./...
```

验证 Python 包构建：

```bash
cd whisper
mise exec -- uv build
```

构建产物输出到 `whisper/dist/`，该目录由 `whisper/.gitignore` 忽略。不要手动提交构建产物。

## 代码风格

### Go backend

- 使用 Go 1.22，保持标准 `gofmt` 风格。
- 不要引入新框架或队列系统，除非任务明确要求。
- HTTP 层位于 `backend/internal/api/`；请求解析、响应结构、路由和 CORS 逻辑应留在该包内。
- SQLite/GORM 持久化位于 `backend/internal/store/`；schema 当前通过 GORM `AutoMigrate` 初始化。
- mock runner 位于 `backend/internal/runner/`；当前阶段不要把 mock 描述成真实下载、转写或翻译。
- API 响应不要暴露本地字幕文件绝对路径；前端应通过 `/subtitle-files/:jobId/:mode` 获取 VTT。

### Python whisper-cli

- 使用 Python 3.12 语法，源码位于 `whisper/src/whisper_cli/`。
- 保持现有风格：类型标注、`from __future__ import annotations`、四空格缩进、简洁函数。
- CLI 入口在 `whisper/src/whisper_cli/cli.py`，脚本名由 `whisper/pyproject.toml` 的 `[project.scripts]` 暴露为 `whisper-cli`。
- 转写 SDK 适配在 `whisper/src/whisper_cli/transcribe.py`；WebVTT 时间轴和 cue 校验在 `whisper/src/whisper_cli/vtt.py`。
- 不要为单次使用创建额外抽象；不要引入新的依赖，除非任务明确需要。
- 当前没有配置统一 formatter 或 linter；不要仅为格式化而大范围改动文件。

## 文档规则

- 根目录 `README.md` 面向人类读者，`AGENTS.md` 面向 coding agent。
- `docs/` 下所有说明性正文必须使用中文；代码、命令、路径、配置键、API 名称和必须保留的引用可以使用英文。
- 更新行为契约、命令、API、退出码、目录结构或 mock/真实边界时，同步检查 `README.md`、`backend/README.md`、`whisper/README.md` 与相关 `docs/` 文件是否需要更新。
- 遵循 AGENTS.md 约定：子目录中更近的 `AGENTS.md` 优先于根目录说明。

## 安全与数据

- 只处理 YouTube 公开视频，不要加入私有视频、登录态或鉴权绕过相关能力。
- 不要把真实音频样本、模型文件、下载缓存、SQLite 数据库、`.env`、API key、LLM 请求日志或构建产物提交到仓库。
- backend 文件服务必须限制在 job 工作目录内，避免路径穿越和符号链接逃逸。
- extension 不应持有翻译 provider 密钥；后续真实翻译密钥应只保存在服务端配置中。

## Pull Request 指南

- 使用 `.github/pull_request_template.md` 填写资料、概要、Close Issue、按文件说明、动作确认和重点 Review 项目。
- 提交 PR 前至少运行与变更相关的测试。
- 涉及 `backend/` 行为时运行 `cd backend && mise exec -- go test ./...`。
- 涉及 `whisper/` 包行为时运行 `cd whisper && mise exec -- uv run pytest`。
- 涉及打包、入口点或依赖配置时额外运行对应构建命令：backend 用 `cd backend && mise exec -- go build ./...`，whisper 用 `cd whisper && mise exec -- uv build`。
- PR 描述中列出实际执行过的验证命令和结果。

## 常见注意事项

- `extension/` 目前不是可运行前端项目；不要新增 npm/pnpm/yarn 流程，除非任务明确进入 extension 实现。
- backend 当前是“真实 API + 真实持久化 + mock runner”，不是完整生产链路。
- 真实转写可能访问网络下载模型；单元测试不应依赖网络、模型下载或本地 GPU。
- `uv.lock` 已提交，修改 Python 依赖时应通过 uv 更新锁文件，而不是手动编辑锁文件。
- `backend/go.sum` 已提交，修改 Go 依赖时应通过 Go 工具更新锁定信息，而不是手动编辑。
