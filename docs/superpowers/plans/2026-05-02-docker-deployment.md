# 后端 Docker 一键部署 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 通过 `docker compose up` 一条命令部署带 real runner 的后端服务，无需手动安装依赖。

**Architecture:** 多阶段构建单镜像 — Go 二进制 + Python whisper-cli venv + yt-dlp + ffmpeg 打包在同一个容器中。后端代码不做任何修改，仍通过 `exec.Command("whisper-cli", ...)` 调用 CLI。Docker compose 编排一个服务 + named volume 持久化数据。

**Tech Stack:** Docker multi-stage build, Go 1.22, Python 3.12, uv, ffmpeg, yt-dlp

---

## 文件结构

| 文件 | 操作 | 职责 |
|---|---|---|
| `backend/Dockerfile` | 创建 | 三阶段构建：Go 编译 → Python whisper-cli → 运行时镜像 |
| `docker-compose.yml` | 创建 | 单服务编排，named volume，环境变量映射 |
| `.env.example` | 创建 | Docker 部署环境变量模板 |
| `.dockerignore` | 创建 | 排除不相关文件，减小构建 context |

**不修改任何现有代码文件。** 后端代码没有任何 Docker 相关改动。

---

### Task 1: 创建 backend/Dockerfile

**Files:**
- Create: `backend/Dockerfile`

- [ ] **Step 1: 创建 Dockerfile**

```dockerfile
# ---- Stage 1: Go build ----
FROM golang:1.22-bookworm AS go-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 go build -o /build/server ./cmd/server

# ---- Stage 2: Python whisper-cli ----
FROM python:3.12-slim-bookworm AS py-builder

WORKDIR /build

RUN pip install uv

COPY whisper/ ./
RUN uv venv /opt/whisper-venv && \
    UV_PROJECT_ENVIRONMENT=/opt/whisper-venv uv sync --no-dev --no-editable

# ---- Stage 3: Runtime ----
FROM python:3.12-slim-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir yt-dlp

COPY --from=go-builder /build/server /usr/local/bin/server
COPY --from=py-builder /opt/whisper-venv /opt/whisper-venv

ENV PATH="/opt/whisper-venv/bin:$PATH" \
    HF_HOME=/data/huggingface \
    LSI_ADDR=0.0.0.0:8080 \
    LSI_DB_PATH=/data/backend.sqlite3 \
    LSI_WORK_DIR=/data/jobs \
    LSI_RUNNER_MODE=real

RUN mkdir -p /data/jobs /data/huggingface

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
```

- [ ] **Step 2: 验证 Dockerfile 语法**

Run: `docker build --check -f backend/Dockerfile . 2>&1 || echo "BUILDKit check not available, proceeding"`

Docker `--check` 可能不普遍可用。此步为可选验证，失败不阻塞后续步骤。

- [ ] **Step 3: 构建镜像**

Run: `docker build -t lsi-backend-test -f backend/Dockerfile .`

Expected: 构建成功，无错误输出。可能需要几分钟（下载 Go 模块、Python 依赖、ffmpeg）。

- [ ] **Step 4: 验证镜像内二进制和工具可用**

Run: `docker run --rm --entrypoint="" lsi-backend-test which whisper-cli`
Run: `docker run --rm --entrypoint="" lsi-backend-test which yt-dlp`
Run: `docker run --rm --entrypoint="" lsi-backend-test which ffmpeg`

Expected: 以上三个命令分别输出 `/opt/whisper-venv/bin/whisper-cli`、`/usr/local/bin/yt-dlp`、`/usr/bin/ffmpeg`。

- [ ] **Step 5: 提交**

```bash
git add backend/Dockerfile
git commit -m "feat(backend): add multi-stage Dockerfile for one-command deployment"
```

---

### Task 2: 创建 docker-compose.yml

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: 创建 docker-compose.yml**

```yaml
services:
  backend:
    build:
      context: .
      dockerfile: backend/Dockerfile
    ports:
      - "${LSI_DOCKER_BIND_HOST:-127.0.0.1}:8080:8080"
    volumes:
      - lsi-data:/data
    environment:
      - LSI_RUNNER_MODE=real
      - LSI_WHISPER_MODEL=${LSI_WHISPER_MODEL:-small}
      - LSI_LLM_BASE_URL=${LSI_LLM_BASE_URL:-https://api.openai.com}
      - LSI_LLM_API_KEY=${LSI_LLM_API_KEY:-}
      - LSI_LLM_MODEL=${LSI_LLM_MODEL:-}
      - LSI_LLM_TIMEOUT=${LSI_LLM_TIMEOUT:-2m}
      - LSI_DOWNLOAD_TIMEOUT=${LSI_DOWNLOAD_TIMEOUT:-10m}
    restart: unless-stopped

volumes:
  lsi-data:
```

注意：端口映射默认绑定 `127.0.0.1`，避免把无鉴权 backend 暴露到局域网。用户如需局域网访问，可在 `.env` 中将 `LSI_DOCKER_BIND_HOST` 设为 `0.0.0.0`。`LSI_ADDR` 在容器内始终为 `0.0.0.0:8080`（Dockerfile ENV 设置），只控制容器内监听地址。

- [ ] **Step 2: 验证 compose 配置语法**

Run: `docker compose config`

Expected: 输出解析后的 YAML，无错误。

- [ ] **Step 3: 提交**

```bash
git add docker-compose.yml
git commit -m "feat: add docker-compose.yml for one-command deployment"
```

---

### Task 3: 创建 .env.example

**Files:**
- Create: `.env.example`

- [ ] **Step 1: 创建 .env.example**

```
# Docker 端口绑定主机（默认仅本机访问；如需局域网访问可改为 0.0.0.0）
LSI_DOCKER_BIND_HOST=127.0.0.1

# LLM 配置（real runner 必填 API_KEY 和 MODEL）
LSI_LLM_BASE_URL=https://api.openai.com
LSI_LLM_API_KEY=sk-your-key-here
LSI_LLM_MODEL=gpt-4.1-mini

# Whisper 模型（首次处理视频时自动下载到持久卷）
LSI_WHISPER_MODEL=small

# 超时设置
LSI_DOWNLOAD_TIMEOUT=10m
LSI_LLM_TIMEOUT=2m
```

注意：没有包含 `LSI_ADDR`，因为容器内地址固定为 `0.0.0.0:8080`，端口发布的绑定主机由 `LSI_DOCKER_BIND_HOST` 控制。

- [ ] **Step 2: 提交**

```bash
git add .env.example
git commit -m "feat: add .env.example for Docker deployment configuration"
```

---

### Task 4: 创建 .dockerignore

**Files:**
- Create: `.dockerignore`

- [ ] **Step 1: 创建 .dockerignore**

```
# Not needed in Docker build
extension/
docs/
graphify-out/

# Git and IDE
.git/
.github/
.vscode/
.idea/

# Python caches (whisper source is needed for build)
whisper/.venv/
whisper/__pycache__/
whisper/dist/

# Data and env
data/
.env
*.sqlite3

# Build artifacts (rebuilt in Docker)
backend/server

# Config for other tools
.codex/
.opencode/
mise.toml
```

- [ ] **Step 2: 提交**

```bash
git add .dockerignore
git commit -m "feat: add .dockerignore to minimize Docker build context"
```

---

### Task 5: 更新 .gitignore

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: 在 .gitignore 末尾追加 .env 忽略项**

在现有 `.gitignore` 末尾追加（如果尚未包含 `.env`）：

```
# Docker
.env
```

先检查 `.gitignore` 和 `backend/.gitignore` 是否已有 `.env` 条目。`backend/.gitignore` 已包含 `# env file` 和 `.env`，但根目录 `.gitignore` 没有。在根 `.gitignore` 追加。

- [ ] **Step 2: 提交**

```bash
git add .gitignore
git commit -m "chore: add .env to root .gitignore for Docker deployment"
```

---

### Task 6: 端到端构建验证

**Files:** 无新文件

- [ ] **Step 1: 完整构建**

Run: `docker compose build`

Expected: 构建成功，无错误。

- [ ] **Step 2: 启动服务验证容器正常**

```bash
cp .env.example .env
docker compose up -d
```

Expected: `docker compose ps` 显示 backend 服务 `Up`。

- [ ] **Step 3: 检查启动日志**

Run: `docker compose logs backend`

Expected: 日志包含 `starting lets-sub-it-api on 0.0.0.0:8080`。

- [ ] **Step 4: 用 curl 测试 mock job 创建**

需要临时切换为 mock 模式来测试。在 `docker-compose.yml` 中将 `LSI_RUNNER_MODE=real` 改为 `LSI_RUNNER_MODE=mock`，然后：

```bash
docker compose up -d
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

Expected: 返回 JSON 包含 `"status":"queued"` 或后续状态。

- [ ] **Step 5: 清理和恢复**

```bash
docker compose down
```

恢复 `docker-compose.yml` 中 `LSI_RUNNER_MODE=real`。确保 `git status` 无残留修改。

---

### Task 7: 更新 README 文档

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 在README 中添加 Docker 部署说明**

在 README.md 的"开发环境"或"快速开始"部分之后，添加 Docker 部署部分。先阅读现有 README 确定插入位置。

添加以下内容：

```markdown
## Docker 部署

一键部署后端（含 real runner）：

```bash
cp .env.example .env
# 编辑 .env，至少填写 LSI_LLM_API_KEY 和 LSI_LLM_MODEL
docker compose up -d
```

查看日志：`docker compose logs -f`
停止服务：`docker compose down`

Docker 默认只绑定 `127.0.0.1:8080`。如需让局域网设备访问，可在 `.env` 中将 `LSI_DOCKER_BIND_HOST` 改为 `0.0.0.0`。

数据（SQLite 数据库、Job 文件、Whisper 模型缓存）持久化在 Docker named volume `lsi-data` 中。
```

- [ ] **Step 2: 提交**

```bash
git add README.md
git commit -m "docs: add Docker deployment instructions to README"
```

---

### Task 8: 更新 AGENTS.md

**Files:**
- Modify: `AGENTS.md`

- [ ] **Step 1: 在 AGENTS.md 的 "Dev Environment Tips" 末尾添加 Docker 部署信息**

在 `"## Dev Environment Tips"` 部分末尾追加：

```markdown
- Docker 部署通过 `docker compose up -d` 一键启动，详见 README "Docker 部署" 部分。构建镜像不需要本地安装 Go、Python、yt-dlp、ffmpeg。
```

- [ ] **Step 2: 更新 Backend Configuration 表格，添加 Docker 默认值说明**

在 `"## Backend Configuration"` 表格下方或 `LSI_ADDR` 行备注中，添加说明：

在 `LSI_ADDR` 行的描述列追加：`Docker 内默认 0.0.0.0:8080（监听所有接口）`

- [ ] **Step 3: 提交**

```bash
git add AGENTS.md
git commit -m "docs: mention Docker deployment in AGENTS.md"
```
