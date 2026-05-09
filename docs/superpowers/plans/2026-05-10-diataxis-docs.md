# Diátaxis 文档体系 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将人类长期文档重组为纯 Diátaxis 四象限，让新成员能从 `docs/README.md` 进入并快速理解项目、跑起系统、找到后续查阅位置。

**Architecture:** `docs/superpowers/` 保留为 AI 生成档案区，但不参与人类文档树。人类文档只由 `tutorials/`、`how-to/`、`reference/`、`explanation/` 四象限组成，每个目录用 `README.md` 作为 GitHub 自动渲染入口。现有 `README.md`、`docs/PRD.md` 和模块 README 中仍正确的事实迁移到四象限后，`docs/PRD.md` 从顶层移除。

**Tech Stack:** Markdown、GitHub README rendering、`rg`、`git diff --check`。

---

## 文件结构

Create:

- `docs/README.md` — 人类文档总入口，说明四象限和推荐阅读路径。
- `docs/tutorials/README.md` — 教程目录页，限定教程文档职责。
- `docs/tutorials/onboarding.md` — 新成员唯一必读教程。
- `docs/how-to/README.md` — 操作指南目录页。
- `docs/how-to/local-development.md` — 本地开发启动步骤。
- `docs/how-to/docker-deployment.md` — Docker 后端部署步骤。
- `docs/how-to/testing.md` — 测试、类型检查和构建验证步骤。
- `docs/how-to/troubleshooting.md` — 常见失败排查。
- `docs/reference/README.md` — 参考资料目录页。
- `docs/reference/repository-structure.md` — 仓库结构和关键入口文件。
- `docs/reference/backend-api.md` — 后端 HTTP API 契约。
- `docs/reference/backend-config.md` — 后端环境变量和运行时要求。
- `docs/reference/data-and-job-model.md` — Job、SubtitleAsset、状态流和扩展缓存模型。
- `docs/reference/whisper-cli.md` — `whisper-cli` 参数、输出和退出码契约。
- `docs/reference/extension-contract.md` — 扩展边界、消息协议、URL 限制和字幕模式。
- `docs/explanation/README.md` — 解释文档目录页。
- `docs/explanation/architecture-overview.md` — 当前整体架构解释。
- `docs/explanation/processing-pipeline.md` — 下载、转写、翻译、打包、渲染链路解释。
- `docs/explanation/module-boundaries.md` — 后端、Whisper CLI、扩展之间的边界解释。
- `docs/explanation/security-and-privacy.md` — 安全、隐私和部署边界解释。

Modify:

- `README.md` — 保持项目门面和快速开始，相关文档改为链接 Diátaxis 文档入口。
- `AGENTS.md` — 更新 `docs/` 描述，避免继续把 docs 描述成 PRD/spec/plan 集合。

Delete after migration:

- `docs/PRD.md` — 其中仍有效的事实迁移到四象限文档后删除，保持 `docs/` 顶层为纯 Diátaxis 入口。

Do not modify:

- `docs/superpowers/**` except this implementation plan. 该目录是 AI 生成档案区，不参与人类长期文档迁移。

---

### Task 1: 创建 Diátaxis 入口和目录页

**Files:**

- Create: `docs/README.md`
- Create: `docs/tutorials/README.md`
- Create: `docs/how-to/README.md`
- Create: `docs/reference/README.md`
- Create: `docs/explanation/README.md`

- [ ] **Step 1: 确认工作树干净**

Run:

```bash
git status --short
```

Expected: no output.

- [ ] **Step 2: 创建四象限目录**

Run:

```bash
mkdir -p docs/tutorials docs/how-to docs/reference docs/explanation
```

Expected: command exits 0.

- [ ] **Step 3: 新增 `docs/README.md`**

Use `apply_patch` to add this file:

```markdown
# Lets Sub It 文档

这里是面向人类协作者的长期文档入口。文档按 Diátaxis 四象限组织：教程、操作指南、参考资料和解释文档。

## 推荐阅读路径

| 你要做什么 | 从哪里开始 |
| --- | --- |
| 第一次加入项目 | [新成员入门](tutorials/onboarding.md) |
| 跑起本地开发环境 | [本地开发](how-to/local-development.md) |
| 用 Docker 跑后端 | [Docker 部署](how-to/docker-deployment.md) |
| 查 API、配置或数据模型 | [参考资料](reference/README.md) |
| 理解架构和边界 | [解释文档](explanation/README.md) |

## 四象限

- [Tutorials](tutorials/README.md)：按顺序学习的教程，优先服务新成员。
- [How-to guides](how-to/README.md)：完成具体任务的操作步骤。
- [Reference](reference/README.md)：稳定契约和查表信息。
- [Explanation](explanation/README.md)：系统为什么这样设计和运行。

`docs/superpowers/` 是 AI 协作过程中生成的规格和计划档案，不是人类长期文档入口。
```

- [ ] **Step 4: 新增四个目录 README**

Use `apply_patch` to add these files:

```markdown
# Tutorials

教程按顺序带读者完成学习路径。这里的文档应该让新成员逐步建立项目心智模型，而不是提供完整参考资料。

- [新成员入门](onboarding.md)
```

Save as `docs/tutorials/README.md`.

```markdown
# How-to Guides

操作指南回答“如何完成一个具体任务”。这里的文档应该给出可执行步骤、命令和检查方式，不承担完整背景解释。

- [本地开发](local-development.md)
- [Docker 部署](docker-deployment.md)
- [测试与构建验证](testing.md)
- [排障](troubleshooting.md)
```

Save as `docs/how-to/README.md`.

```markdown
# Reference

参考资料回答“当前契约是什么”。这里的文档应该准确、完整、便于查找，避免展开长篇设计解释。

- [仓库结构](repository-structure.md)
- [后端 API](backend-api.md)
- [后端配置](backend-config.md)
- [数据与任务模型](data-and-job-model.md)
- [Whisper CLI](whisper-cli.md)
- [扩展契约](extension-contract.md)
```

Save as `docs/reference/README.md`.

```markdown
# Explanation

解释文档回答“为什么系统这样工作”。这里放架构、处理链路、模块边界、安全和隐私等长期设计说明。

- [架构总览](architecture-overview.md)
- [处理链路](processing-pipeline.md)
- [模块边界](module-boundaries.md)
- [安全与隐私](security-and-privacy.md)
```

Save as `docs/explanation/README.md`.

- [ ] **Step 5: 验证入口文件存在**

Run:

```bash
test -f docs/README.md
test -f docs/tutorials/README.md
test -f docs/how-to/README.md
test -f docs/reference/README.md
test -f docs/explanation/README.md
git diff --check
```

Expected: all commands exit 0.

- [ ] **Step 6: Commit**

Run:

```bash
git add docs/README.md docs/tutorials/README.md docs/how-to/README.md docs/reference/README.md docs/explanation/README.md
git commit -m "docs: add diataxis documentation entrypoints"
```

Expected: commit succeeds.

---

### Task 2: 编写新成员入门教程

**Files:**

- Create: `docs/tutorials/onboarding.md`

- [ ] **Step 1: 新增 onboarding 文档**

Use `apply_patch` to create `docs/tutorials/onboarding.md` with these sections and content requirements:

```markdown
# 新成员入门

这篇教程帮助新成员在 30-60 分钟内理解 Lets Sub It 的目标、模块关系和本地运行路径。完成后，你应该能描述字幕处理链路，能启动相关模块，并知道后续按什么文档查契约。

## 1. 先理解项目目标

Lets Sub It 是一个本地优先的 YouTube 字幕生成与翻译工具。用户提交 YouTube 公开视频链接后，后端下载音频，调用本地 Whisper 转写，用 OpenAI-compatible LLM 翻译，再让 Chrome 扩展把字幕显示在 YouTube 播放页。

当前项目面向单用户本地自托管，不提供登录、多租户、计费或公网部署保护。

## 2. 认识三个模块

| 模块 | 作用 | 先看哪里 |
| --- | --- | --- |
| `backend/` | Go HTTP API、SQLite 持久化、job 复用、真实 runner、VTT 文件服务 | [后端 API](../reference/backend-api.md)、[后端配置](../reference/backend-config.md) |
| `whisper/` | Python `whisper-cli`，输入本地音频，输出合法 WebVTT | [Whisper CLI](../reference/whisper-cli.md) |
| `extension/` | Chrome MV3 扩展，负责 popup、background API 网关和 YouTube 字幕层 | [扩展契约](../reference/extension-contract.md) |

## 3. 跑起项目

如果你只是想尽快跑真实后端，走 [Docker 部署](../how-to/docker-deployment.md)。

如果你要开发代码，走 [本地开发](../how-to/local-development.md)。本地开发需要 `mise` 安装固定版本的 Go、Python、Node.js 和 `uv`。

## 4. 理解处理链路

核心链路是：

```text
YouTube URL -> backend job -> yt-dlp/ffmpeg -> whisper-cli -> LLM translation -> VTT packaging -> extension overlay
```

更完整的解释见 [处理链路](../explanation/processing-pipeline.md)。

## 5. 知道去哪改代码

| 你要改什么 | 先查 |
| --- | --- |
| HTTP 请求、响应、错误模型 | [后端 API](../reference/backend-api.md) |
| 环境变量、运行时工具要求 | [后端配置](../reference/backend-config.md) |
| Job 状态或字幕资产 | [数据与任务模型](../reference/data-and-job-model.md) |
| 转写 CLI 参数、退出码、VTT 契约 | [Whisper CLI](../reference/whisper-cli.md) |
| popup、background、content script 协作 | [扩展契约](../reference/extension-contract.md) |
| 模块之间该不该互相调用 | [模块边界](../explanation/module-boundaries.md) |

## 6. 完成标准

读完并完成对应运行步骤后，你应该能：

- 说明 `backend`、`whisper`、`extension` 如何协作。
- 选择 Docker 或本地开发路径启动项目。
- 说出 job 的主要状态流。
- 知道 API、配置、CLI、扩展协议分别在哪些 reference 文档里查。
```

- [ ] **Step 2: 验证 onboarding 链接目标存在**

Run:

```bash
test -f docs/tutorials/onboarding.md
test -f docs/reference/backend-api.md || true
test -f docs/reference/backend-config.md || true
test -f docs/reference/whisper-cli.md || true
test -f docs/reference/extension-contract.md || true
git diff --check
```

Expected: `onboarding.md` exists and `git diff --check` exits 0. Reference files are created in later tasks, so their checks may produce no lasting assertion in this task.

- [ ] **Step 3: Commit**

Run:

```bash
git add docs/tutorials/onboarding.md
git commit -m "docs: add onboarding tutorial"
```

Expected: commit succeeds.

---

### Task 3: 编写操作指南

**Files:**

- Create: `docs/how-to/local-development.md`
- Create: `docs/how-to/docker-deployment.md`
- Create: `docs/how-to/testing.md`
- Create: `docs/how-to/troubleshooting.md`

- [ ] **Step 1: 编写 `local-development.md`**

Use `apply_patch` to create `docs/how-to/local-development.md`. Include:

- `# 本地开发`
- 前置条件：`mise install`、`yt-dlp`、`ffmpeg`、LLM 配置。
- 依赖安装命令：

```bash
cd backend && mise exec -- go mod download
cd ../whisper && mise exec -- uv sync --dev
cd ../extension && mise exec -- npm install
```

- 后端启动命令，使用 root `README.md` 中现有 `PATH="$PWD/../whisper/.venv/bin:$PATH"` 示例，并包含 `LSI_LOG_LEVEL=info`。
- API smoke test `curl -X POST "http://127.0.0.1:8080/jobs"`，请求体包含 `youtubeUrl`、`sourceLanguage`、`targetLanguage`。
- 扩展 dev server 命令 `cd extension && mise exec -- npm run dev`。
- Chrome 加载路径 `extension/.output/chrome-mv3`。

- [ ] **Step 2: 编写 `docker-deployment.md`**

Use `apply_patch` to create `docs/how-to/docker-deployment.md`. Include:

- `# Docker 部署`
- 说明 Docker 后端镜像包含 Go 后端、Python `whisper-cli`、`yt-dlp` 和 `ffmpeg`。
- 命令：

```bash
cp .env.example .env
docker compose up -d
docker compose logs -f
docker compose down
```

- `.env` 最少配置：`LSI_LLM_API_KEY`、`LSI_LLM_MODEL`。
- 默认绑定 `127.0.0.1:8080`，局域网访问才改 `LSI_DOCKER_BIND_HOST=0.0.0.0`。
- 数据卷：`lsi-data` 保存 SQLite 和 job 文件，`lsi-hf-cache` 保存 Hugging Face/Whisper 模型缓存。
- 明确不要把服务直接暴露到公网。

- [ ] **Step 3: 编写 `testing.md`**

Use `apply_patch` to create `docs/how-to/testing.md`. Include:

- `# 测试与构建验证`
- 全量测试命令：

```bash
cd backend && mise exec -- go test ./...
cd ../whisper && mise exec -- uv run pytest
cd ../extension && mise exec -- npm run test
```

- 聚焦测试命令：

```bash
cd backend && mise exec -- go test ./internal/api
cd backend && mise exec -- go test ./internal/runner -run TestRealRunnerCompletesJob
cd whisper && mise exec -- uv run pytest tests/test_vtt.py
cd whisper && mise exec -- uv run pytest -k "vtt"
cd extension && mise exec -- npx vitest run src/api/backend-client.test.ts
cd extension && mise exec -- npm run typecheck
```

- 构建验证命令：

```bash
cd backend && mise exec -- go build ./...
cd ../whisper && mise exec -- uv build
cd ../extension && mise exec -- npm run build
```

- 说明测试必须离线友好，不使用真实 YouTube、模型下载、GPU、外部 LLM 或 provider key。

- [ ] **Step 4: 编写 `troubleshooting.md`**

Use `apply_patch` to create `docs/how-to/troubleshooting.md`. Include a table with these rows:

| 问题 | 检查项 |
| --- | --- |
| 后端启动失败并提示工具缺失 | 确认 `yt-dlp`、`ffmpeg` 和 `whisper-cli` 在 `PATH` 中 |
| job 在 `translating` 阶段失败 | 确认 `LSI_LLM_BASE_URL`、`LSI_LLM_MODEL` 已配置；OpenAI 默认 endpoint 还需要 `LSI_LLM_API_KEY` |
| 扩展无法连接后端 | 确认 backend URL 是带端口的本机 HTTP origin，且不包含路径或查询参数 |
| Whisper 首次运行很慢 | 真实转写可能触发模型下载；单元测试不会下载模型或依赖 GPU |
| 字幕文件返回 404 | 确认任务已完成，且 `mode` 是 `source`、`translated` 或 `bilingual` |

- [ ] **Step 5: 验证操作指南**

Run:

```bash
test -f docs/how-to/local-development.md
test -f docs/how-to/docker-deployment.md
test -f docs/how-to/testing.md
test -f docs/how-to/troubleshooting.md
rg -n "LSI_LLM_API_KEY|mise exec --|docker compose|yt-dlp|ffmpeg" docs/how-to
git diff --check
```

Expected: `rg` prints matching lines from the new how-to files, and `git diff --check` exits 0.

- [ ] **Step 6: Commit**

Run:

```bash
git add docs/how-to/local-development.md docs/how-to/docker-deployment.md docs/how-to/testing.md docs/how-to/troubleshooting.md
git commit -m "docs: add operational guides"
```

Expected: commit succeeds.

---

### Task 4: 编写参考资料

**Files:**

- Create: `docs/reference/repository-structure.md`
- Create: `docs/reference/backend-api.md`
- Create: `docs/reference/backend-config.md`
- Create: `docs/reference/data-and-job-model.md`
- Create: `docs/reference/whisper-cli.md`
- Create: `docs/reference/extension-contract.md`

- [ ] **Step 1: 编写 `repository-structure.md`**

Create a concise reference containing:

- `# 仓库结构`
- top-level tree for `backend/`、`whisper/`、`extension/`、`docs/`、`docker-compose.yml`、`mise.toml`
- key entry point table copied from `AGENTS.md`:
  - `backend/cmd/server/main.go`
  - `backend/internal/app/`
  - `backend/internal/api/`
  - `backend/internal/store/`
  - `backend/internal/runner/`
  - `whisper/src/whisper_cli/cli.py`
  - `whisper/src/whisper_cli/transcribe.py`
  - `whisper/src/whisper_cli/vtt.py`
  - `extension/entrypoints/background.ts`
  - `extension/entrypoints/youtube.content.ts`
  - `extension/entrypoints/popup/`

- [ ] **Step 2: 编写 `backend-api.md`**

Create a reference containing:

- `# 后端 API`
- API table with exactly these endpoints:
  - `POST /jobs`
  - `GET /jobs/:id`
  - `GET /jobs/active?videoId=...&targetLanguage=...`
  - `GET /subtitle-assets?videoId=...&targetLanguage=...`
  - `GET /subtitle-files/:jobId/:mode`
- `POST /jobs` request JSON:

```json
{
  "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "sourceLanguage": "en",
  "targetLanguage": "zh"
}
```

- `mode` values: `source`、`translated`、`bilingual`
- security note: API responses must not expose local absolute file paths.

- [ ] **Step 3: 编写 `backend-config.md`**

Create a reference table with these environment variables and defaults:

- `LSI_ADDR`: `127.0.0.1:8080`
- `LSI_DB_PATH`: `./data/backend.sqlite3`
- `LSI_WORK_DIR`: `./data/jobs`
- `LSI_LOG_LEVEL`: `info`
- `LSI_DOWNLOAD_TIMEOUT`: `10m`
- `LSI_WHISPER_MODEL`: `small`
- `LSI_WHISPER_COMPUTE_TYPE`: `default`
- `HF_TOKEN`: empty
- `LSI_LLM_BASE_URL`: `https://api.openai.com`
- `LSI_LLM_API_KEY`: empty
- `LSI_LLM_MODEL`: empty
- `LSI_LLM_TIMEOUT`: `2m`

Include runtime tool requirements: `yt-dlp`、`ffmpeg`、`whisper-cli` on `PATH` for local backend; Docker image already includes them.

- [ ] **Step 4: 编写 `data-and-job-model.md`**

Create a reference containing:

- `# 数据与任务模型`
- Job state flow:

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

- failed state rule: `failed` records `errorMessage`.
- `Job` field list from `docs/PRD.md`: `id`、`videoId`、`youtubeUrl`、`sourceLanguage`、`targetLanguage`、`status`、`stage`、`progressText`、`errorMessage`、`attempt`、`workingDir`、`createdAt`、`updatedAt`
- `SubtitleAsset` field list from `docs/PRD.md`
- extension cache entries: `SubtitleAssetCacheEntry` and `VideoPreference`
- reuse key: `videoId + targetLanguage`

- [ ] **Step 5: 编写 `whisper-cli.md`**

Create a reference containing:

- `# Whisper CLI`
- direct command:

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --compute-type int8 \
  --language ja
```

- required params: `--input`、`--output`、`--model`、`--language`
- optional currently used param: `--compute-type`
- exit codes: `0`、`2`、`3`、`4`
- stdout success JSON shape:

```json
{
  "output": "/tmp/source.vtt",
  "language": "ja",
  "duration_seconds": 123.45,
  "segments": 42
}
```

- WebVTT output rule: output must be valid WebVTT; tests use fake models and stay offline.

- [ ] **Step 6: 编写 `extension-contract.md`**

Create a reference containing:

- `# 扩展契约`
- technology stack: WXT、Vue、TypeScript、Vitest、Tailwind/shadcn-vue
- boundaries:
  - background service worker is the only HTTP API gateway
  - content scripts must not call backend directly
  - extension never stores provider keys
- backend URL rule: only `http://localhost:<port>` and `http://127.0.0.1:<port>`
- manifest host permissions:
  - `http://127.0.0.1:*/*`
  - `http://localhost:*/*`
- supported languages: `en` and `zh`; `sourceLanguage` must not equal `targetLanguage`
- watch page subtitle modes: `translated` and `bilingual`; backend also serves `source`
- message protocol names from `docs/PRD.md`: `settings:get`、`settings:update`、`job:create`、`job:get`、`job:active`、`subtitle:resolve`、`subtitle:fetch-file`、`subtitle:update-mode`

- [ ] **Step 7: 验证参考资料**

Run:

```bash
test -f docs/reference/repository-structure.md
test -f docs/reference/backend-api.md
test -f docs/reference/backend-config.md
test -f docs/reference/data-and-job-model.md
test -f docs/reference/whisper-cli.md
test -f docs/reference/extension-contract.md
rg -n "POST /jobs|LSI_ADDR|queued -> downloading|whisper-cli|subtitle:fetch-file" docs/reference
git diff --check
```

Expected: `rg` prints matching contract lines, and `git diff --check` exits 0.

- [ ] **Step 8: Commit**

Run:

```bash
git add docs/reference/repository-structure.md docs/reference/backend-api.md docs/reference/backend-config.md docs/reference/data-and-job-model.md docs/reference/whisper-cli.md docs/reference/extension-contract.md
git commit -m "docs: add reference contracts"
```

Expected: commit succeeds.

---

### Task 5: 编写解释文档并迁移 PRD 长期事实

**Files:**

- Create: `docs/explanation/architecture-overview.md`
- Create: `docs/explanation/processing-pipeline.md`
- Create: `docs/explanation/module-boundaries.md`
- Create: `docs/explanation/security-and-privacy.md`
- Delete: `docs/PRD.md`

- [ ] **Step 1: 编写 `architecture-overview.md`**

Create an explanation containing:

- `# 架构总览`
- system split into three runtime modules: `extension`、Go API server with embedded runner、`whisper-cli`
- mermaid diagram copied from root `README.md` and adjusted only if needed for links
- explanation that SQLite and job files are local, extension communicates through background service worker, backend calls external tools and LLM
- link to `../reference/repository-structure.md`

- [ ] **Step 2: 编写 `processing-pipeline.md`**

Create an explanation containing:

- `# 处理链路`
- numbered flow from user submitting URL to content script rendering subtitles
- stage table:
  - `downloading`: `yt-dlp` + `ffmpeg`
  - `transcribing`: `whisper-cli`
  - `translating`: OpenAI-compatible LLM
  - `packaging`: backend writes `translated.vtt` and `bilingual.vtt`
- completion conditions from `docs/PRD.md` for download, transcribe, translate, package
- link to `../reference/data-and-job-model.md`

- [ ] **Step 3: 编写 `module-boundaries.md`**

Create an explanation containing:

- `# 模块边界`
- backend owns HTTP parsing, CORS, SQLite, job lifecycle, runner, file serving
- whisper owns local audio-to-WebVTT transcription and CLI exit code contract
- extension owns popup, storage, background message gateway, YouTube page integration
- boundary rules:
  - content scripts do not call backend directly
  - frontend uses `/subtitle-files/:jobId/:mode`, not local paths
  - tests use fakes/stubs for external tools and LLM
  - no extra queue framework unless explicitly requested

- [ ] **Step 4: 编写 `security-and-privacy.md`**

Create an explanation containing:

- `# 安全与隐私`
- public YouTube videos only
- backend has no user auth and is for single-user local self-hosting
- provider keys are backend/server config only
- extension does not store secrets and does not call translation providers directly
- localhost-only extension backend URLs and manifest host permissions
- file serving stays within job work directory and must preserve traversal/symlink protections
- no LLM request logs containing private data or keys

- [ ] **Step 5: 删除旧 `docs/PRD.md`**

Run:

```bash
git rm docs/PRD.md
```

Expected: file is staged for deletion. Current product and architecture facts are represented in the new tutorial, reference, and explanation documents.

- [ ] **Step 6: 验证解释文档**

Run:

```bash
test -f docs/explanation/architecture-overview.md
test -f docs/explanation/processing-pipeline.md
test -f docs/explanation/module-boundaries.md
test -f docs/explanation/security-and-privacy.md
test ! -f docs/PRD.md
rg -n "content scripts|localhost|yt-dlp|whisper-cli|OpenAI-compatible|public YouTube" docs/explanation
git diff --check
```

Expected: `docs/PRD.md` no longer exists, `rg` prints matching explanation lines, and `git diff --check` exits 0.

- [ ] **Step 7: Commit**

Run:

```bash
git add docs/explanation/architecture-overview.md docs/explanation/processing-pipeline.md docs/explanation/module-boundaries.md docs/explanation/security-and-privacy.md
git commit -m "docs: add architecture explanations"
```

Expected: commit includes the four new explanation files and deletion of `docs/PRD.md`.

---

### Task 6: 对齐根 README、AGENTS 和链接完整性

**Files:**

- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: 更新 `README.md` 仓库结构**

Modify the `docs/` line under `## 仓库结构` from PRD/spec/plan wording to:

```text
├── docs/                    # Diátaxis 文档：tutorials、how-to、reference、explanation
```

- [ ] **Step 2: 更新 `README.md` 相关文档**

Replace the current `## 相关文档` list with:

```markdown
## 相关文档

- [文档入口](docs/README.md)
- [新成员入门](docs/tutorials/onboarding.md)
- [本地开发](docs/how-to/local-development.md)
- [Docker 部署](docs/how-to/docker-deployment.md)
- [后端 API](docs/reference/backend-api.md)
- [架构总览](docs/explanation/architecture-overview.md)
- [Backend README](backend/README.md)
- [Whisper README](whisper/README.md)
- [Extension README](extension/README.md)
```

- [ ] **Step 3: 更新 `AGENTS.md` 项目概览**

Modify the `docs/` bullet in `AGENTS.md` to:

```markdown
- `docs/` — Chinese human documentation organized with Diátaxis (`tutorials/`, `how-to/`, `reference/`, `explanation/`) plus AI-generated archives under `docs/superpowers/`. The nested `docs/AGENTS.md` takes precedence for all files under `docs/`.
```

- [ ] **Step 4: 最终链接和结构验证**

Run:

```bash
test -f docs/README.md
test -f docs/tutorials/onboarding.md
test -f docs/how-to/local-development.md
test -f docs/how-to/docker-deployment.md
test -f docs/how-to/testing.md
test -f docs/how-to/troubleshooting.md
test -f docs/reference/repository-structure.md
test -f docs/reference/backend-api.md
test -f docs/reference/backend-config.md
test -f docs/reference/data-and-job-model.md
test -f docs/reference/whisper-cli.md
test -f docs/reference/extension-contract.md
test -f docs/explanation/architecture-overview.md
test -f docs/explanation/processing-pipeline.md
test -f docs/explanation/module-boundaries.md
test -f docs/explanation/security-and-privacy.md
test ! -f docs/PRD.md
rg -n "docs/PRD.md|docs/superpowers/specs/2026-04|docs/superpowers/specs/2026-05-02" README.md docs/README.md docs/tutorials docs/how-to docs/reference docs/explanation
git diff --check
```

Expected:

- all `test` commands exit 0
- `rg` exits 1 with no matches, because human-facing entrypoints should not link to old PRD or ticket-level superpowers specs
- `git diff --check` exits 0

- [ ] **Step 5: 运行 docs-only 总验证**

Run:

```bash
find docs -maxdepth 2 -type f | sort
git status --short
```

Expected:

- output includes `docs/README.md` and files under all four Diátaxis directories
- `docs/superpowers/**` remains present
- status only shows intended documentation and `AGENTS.md` changes before commit

- [ ] **Step 6: Commit**

Run:

```bash
git add README.md AGENTS.md
git commit -m "docs: link root readme to diataxis docs"
```

Expected: commit succeeds.

---

## Self-Review

Spec coverage:

- Pure Diátaxis四象限: Task 1 creates the structure and Task 6 verifies all target files.
- GitHub-first入口: Task 1 uses `README.md` for root and all quadrants.
- New member priority: Task 2 creates `tutorials/onboarding.md`.
- No ADR: no ADR directory or decision log is created.
- `docs/superpowers/` outside human docs: Task 1 and Task 6 document and verify that human entrypoints do not link to old ticket specs.
- Migration from existing facts: Tasks 3-5 explicitly use root `README.md`、`docs/PRD.md` and module README content.
- Root README and AGENTS alignment: Task 6 updates both.

Placeholder scan:

- The plan contains concrete files, commands, headings, source facts, and expected outputs.
- The plan does not require future workers to invent the information architecture.

Type and path consistency:

- All paths use the approved Diátaxis directories.
- All internal links are relative to their source document.
- `docs/PRD.md` is deleted only after its stable facts are represented in the new docs.
