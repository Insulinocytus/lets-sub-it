# AGENTS.md

## 项目概览

Lets Sub It 是一个自托管的 YouTube 字幕生成与翻译工具。目标链路是：提交 YouTube 公开视频链接，下载音频，本地转写，翻译，生成字幕，并由播放页加载字幕。

当前仓库处于 MVP 阶段，是一个多模块仓库：

- `backend/`：Go 1.22 mock API server。提供真实 HTTP API、SQLite 持久化、job 复用、mock runner 状态推进和 VTT 字幕文件服务。
- `whisper/`：Python 3.12 包 `whisper-cli`。通过 `faster-whisper` 把本地音频文件转成经过校验的 WebVTT 字幕。
- `extension/`：Chrome MV3 extension。使用 WXT、Vue、TypeScript 和 Vitest，支持 popup 提交任务、background 统一访问 backend、storage 缓存和 YouTube watch 页面字幕层。
- `docs/`：PRD、规格说明与实施计划。该目录下另有 `docs/AGENTS.md`，要求文档正文使用中文。

当前 backend 是“真实 API + 真实持久化 + mock runner”，不会访问 YouTube、调用 LLM 或执行真实转写。真实 `yt-dlp` 下载、`ffmpeg` 音频处理、backend 调用 `whisper-cli`、OpenAI-compatible LLM 翻译，以及基于真实字幕生成 `translated.vtt` / `bilingual.vtt` 仍是后续迭代。

主要入口：

- backend server：`backend/cmd/server/main.go`
- backend API：`backend/internal/api/`
- backend store：`backend/internal/store/`
- backend runner：`backend/internal/runner/`
- whisper CLI：`whisper/src/whisper_cli/cli.py`
- whisper 转写适配：`whisper/src/whisper_cli/transcribe.py`
- whisper VTT 处理：`whisper/src/whisper_cli/vtt.py`
- extension entrypoints：`extension/entrypoints/`
- extension API/message/storage/subtitle/youtube 逻辑：`extension/src/`

## 协作原则

- 始终使用简体中文与用户沟通。
- 优先做最小可行改动，不添加未被要求的功能、抽象或配置。
- 修改现有代码时保持手术式变更：只触碰与任务直接相关的文件和行。
- 如果需求存在多种解释，先说明假设；不确定且风险较高时先询问。
- 不要清理无关代码、重排无关格式或重构未被要求的模块。
- 不要覆盖用户或其他 agent 已做的未提交改动；开始前和结束前用 `git status --short` 核对工作树。
- 不要提交构建产物、数据库、模型文件、下载缓存、真实音频样本、`.env`、API key 或 LLM 请求日志。

## 环境与依赖

AI agent 的 shell 环境不会自动执行 `eval "$(mise activate zsh)"`。凡是使用 `mise.toml` 中锁定版本的工具，都必须通过 `mise exec --` 前缀调用，例如 `mise exec -- go`、`mise exec -- uv`、`mise exec -- python`、`mise exec -- npm`。

工具链由根目录 `mise.toml` 管理：

- Go `1.22`
- Python `3.12`
- Node.js `22`
- `uv`

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

同步 extension 依赖：

```bash
cd extension
mise exec -- npm install
```

根目录没有统一 package manager 或 monorepo runner。请进入具体子目录执行对应命令。

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
    "sourceLanguage": "en",
    "targetLanguage": "zh-CN"
  }'
```

启动 extension 开发服务器：

```bash
cd extension
mise exec -- npm run dev
```

在 Chrome extension developer mode 中加载 WXT 输出目录 `.output/chrome-mv3`。popup 默认连接 `http://127.0.0.1:8080`，当前只支持 `http://localhost:<port>` 或 `http://127.0.0.1:<port>`。

运行本地 Whisper CLI 示例：

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

真实转写会调用 `faster-whisper`，可能需要模型下载能力以及本机可用的推理运行环境。单元测试使用 fake model，不会下载或运行真实 Whisper 模型。

## Backend 契约

主要配置：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | job 工作目录根路径 |

HTTP API：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕生成 job |
| `GET` | `/jobs/:id` | 查询 job 状态 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询已完成字幕资产 |
| `GET` | `/subtitle-files/:jobId/:mode` | 读取 VTT 文件，`mode` 为 `source`、`translated` 或 `bilingual` |

状态流转：

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

失败时状态为 `failed`，响应中的 `errorMessage` 记录错误摘要。

## Extension 契约

- 技术栈是 `WXT + Vue + TypeScript + Vite + Vitest + shadcn-vue`。
- 包管理器是 npm，锁文件是 `extension/package-lock.json`。
- 入口文件在 `extension/entrypoints/`：`background.ts`、`youtube.content.ts`、`popup/`。
- background service worker 是唯一 HTTP API 网关；popup 和 content script 通过 runtime message 访问它。
- 支持语言固定为 `en` 和 `zh-CN`，`sourceLanguage` 与 `targetLanguage` 不能相同。
- 播放页字幕模式只支持 `translated` 和 `bilingual`；backend 仍保留 `source` 文件服务。
- backend URL 只能是带端口的本机 HTTP origin，例如 `http://127.0.0.1:8080`。
- Manifest host permissions 只允许 `http://127.0.0.1:*/*` 和 `http://localhost:*/*`。

## Whisper CLI 契约

`whisper-cli` 输入本地音频文件，输出合法 WebVTT。成功时 stdout 输出 JSON，`--output` 写入 `.vtt` 文件。

参数：

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `--input` | 是 | 本地音频文件路径 |
| `--output` | 是 | 输出 `.vtt` 路径，不能与输入路径相同 |
| `--model` | 是 | `faster-whisper` 模型名，例如 `small` |
| `--language` | 是 | 转写语言代码，例如 `ja`、`en` |

退出码：

| 退出码 | 含义 |
| --- | --- |
| `0` | 成功 |
| `2` | 输入校验失败，例如输入文件不存在、模型名或语言无效 |
| `3` | 转写失败 |
| `4` | 输出校验失败，例如无法生成合法 VTT |

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

运行 extension 全部测试：

```bash
cd extension
mise exec -- npm run test
```

运行 extension 类型检查：

```bash
cd extension
mise exec -- npm run typecheck
```

运行 extension 聚焦测试：

```bash
cd extension
mise exec -- npx vitest run src/api/backend-client.test.ts
```

测试规则：

- backend 测试文件与被测包同目录，命名为 `*_test.go`。
- whisper 的 pytest 配置在 `whisper/pyproject.toml`：`pythonpath = ["src"]`，默认 `addopts = "-q"`。测试文件命名遵循 `test_*.py`。
- extension 使用 Vitest + jsdom + WXT Vitest plugin，测试文件位于 `extension/src/**/*.test.ts`。
- 改动 `backend/internal/` 行为时，优先添加或更新同包 Go 测试，然后确认相关 `go test` 通过。
- 改动 `whisper/src/whisper_cli/` 行为时，优先添加或更新相邻 pytest 测试，然后确认相关 pytest 通过。
- 改动 `extension/src/` 或 `extension/entrypoints/` 行为时，优先添加或更新相邻 Vitest 测试，并运行相关 `npm run test` 或聚焦测试。
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

验证 extension 可构建：

```bash
cd extension
mise exec -- npm run build
```

构建产物：

- backend 默认 Go build 产物不要提交。
- `whisper/dist/` 由 `whisper/.gitignore` 忽略，不要提交。
- `extension/.output/` 和 `extension/.wxt/` 由 `extension/.gitignore` 忽略，不要提交。

## 代码风格

### Go backend

- 使用 Go 1.22，保持标准 `gofmt` 风格。
- 不要引入新框架、队列系统或后台任务系统，除非任务明确要求。
- HTTP 层位于 `backend/internal/api/`；请求解析、响应结构、路由和 CORS 逻辑应留在该包内。
- SQLite/GORM 持久化位于 `backend/internal/store/`；schema 当前通过 GORM `AutoMigrate` 初始化。
- mock runner 位于 `backend/internal/runner/`；当前阶段不要把 mock 描述成真实下载、转写或翻译。
- API 响应不要暴露本地字幕文件绝对路径；前端应通过 `/subtitle-files/:jobId/:mode` 获取 VTT。
- 文件服务必须限制在 job 工作目录内，避免路径穿越和符号链接逃逸。

### Python whisper-cli

- 使用 Python 3.12 语法，源码位于 `whisper/src/whisper_cli/`。
- 保持现有风格：类型标注、`from __future__ import annotations`、四空格缩进、简洁函数。
- CLI 入口在 `whisper/src/whisper_cli/cli.py`，脚本名由 `whisper/pyproject.toml` 的 `[project.scripts]` 暴露为 `whisper-cli`。
- 转写 SDK 适配在 `whisper/src/whisper_cli/transcribe.py`；WebVTT 时间轴和 cue 校验在 `whisper/src/whisper_cli/vtt.py`。
- 不要为单次使用创建额外抽象；不要引入新依赖，除非任务明确需要。
- 当前没有配置统一 formatter 或 linter；不要仅为格式化而大范围改动文件。
- `uv.lock` 已提交，修改 Python 依赖时应通过 uv 更新锁文件，而不是手动编辑锁文件。

### Chrome extension

- 使用 TypeScript、Vue SFC、WXT 和 npm。
- 路径别名 `@/*` 指向 `extension/src/*`。
- popup UI 位于 `extension/entrypoints/popup/`；业务校验优先放在 `extension/src/popup/`。
- background 消息协议和 HTTP client 位于 `extension/src/api/`。
- extension storage 逻辑位于 `extension/src/storage/`。
- WebVTT 解析和当前 cue 命中逻辑位于 `extension/src/subtitles/`，保持可独立测试。
- YouTube watch URL 和 SPA 页面变化逻辑位于 `extension/src/youtube/`。
- shadcn-vue 本地组件位于 `extension/src/components/ui/`。只添加实际使用的组件，不提前引入组件库范围外代码。
- 不要在 content script 中直接访问 Go backend；网络请求统一通过 background service worker。
- 不要把翻译 provider 密钥或任何长期敏感凭据放进 extension。
- `package-lock.json` 已提交，修改 npm 依赖时通过 npm 更新锁文件，而不是手动编辑锁文件。

## 文档规则

- 根目录 `README.md` 面向人类读者，`AGENTS.md` 面向 coding agent。
- `docs/` 下所有说明性正文必须使用中文；代码、命令、路径、配置键、API 名称和必须保留的引用可以使用英文。
- 遵循 AGENTS.md 约定：子目录中更近的 `AGENTS.md` 优先于根目录说明。
- 更新行为契约、命令、API、退出码、目录结构、支持语言、mock/真实边界或安全边界时，同步检查 `README.md`、`backend/README.md`、`whisper/README.md`、`extension/README.md` 与相关 `docs/` 文件是否需要更新。

## 安全与数据

- 只处理 YouTube 公开视频，不要加入私有视频、登录态、cookie 导入或鉴权绕过相关能力。
- backend 当前没有用户鉴权，默认面向单用户本地自托管；不要把它描述成公网生产服务。
- 不要把真实音频样本、模型文件、下载缓存、SQLite 数据库、`.env`、API key、LLM 请求日志或构建产物提交到仓库。
- extension 只允许本机 backend URL；不要放宽到任意远程主机，除非任务明确要求并同步更新权限说明。
- 后续真实翻译密钥应只保存在服务端配置中，extension 不应持有 provider 密钥。

## Pull Request 指南

- 仓库当前未发现 `.github/` workflow 或 PR template；不要假设存在 CI 或模板。
- PR 或变更说明建议按影响范围标注，例如 `[backend]`、`[whisper]`、`[extension]`、`[docs]`。
- 提交 PR 前至少运行与变更相关的测试，并在说明中列出实际执行过的验证命令和结果。
- 涉及 `backend/` 行为时运行 `cd backend && mise exec -- go test ./...`。
- 涉及 `whisper/` 包行为时运行 `cd whisper && mise exec -- uv run pytest`。
- 涉及 `extension/` 行为时运行 `cd extension && mise exec -- npm run test`，类型相关改动还应运行 `cd extension && mise exec -- npm run typecheck`。
- 涉及打包、入口点或依赖配置时额外运行对应构建命令：backend 用 `cd backend && mise exec -- go build ./...`，whisper 用 `cd whisper && mise exec -- uv build`，extension 用 `cd extension && mise exec -- npm run build`。

## 常见注意事项

- backend 当前是 mock runner，不要把 mock VTT 或 mock 进度描述成真实下载、真实转写或真实翻译结果。
- extension 第一版只支持 `en` 和 `zh-CN`，不要在 UI 或文档中承诺完整语言列表。
- backend CORS 只允许本机 HTTP origin 且必须显式端口。
- 单元测试必须离线、可重复，不依赖真实 YouTube、模型下载、GPU、外部 LLM 或本地私有数据。
- 修改 Go 依赖时通过 Go 工具更新 `backend/go.sum`，不要手动编辑。
- 修改 Python 依赖时通过 uv 更新 `whisper/uv.lock`，不要手动编辑。
- 修改 extension 依赖时通过 npm 更新 `extension/package-lock.json`，不要手动编辑。
