# 后端 Docker 一键部署设计

> 日期: 2026-05-02
> 状态: Draft

## 目标

让用户通过 `docker compose up` 一条命令就能在本地部署带 real runner 的后端服务，无需手动安装 Go、Python、yt-dlp、ffmpeg 等依赖。

## 方案选择

**方案 A：多阶段构建单镜像**（选中）

一个 Dockerfile 通过多阶段构建打包 Go 后端 + Python whisper-cli + yt-dlp + ffmpeg，compose 一键启动。不需要改后端代码（仍通过 `exec.Command` 调 `whisper-cli` CLI）。

排除的方案：
- 方案 B（双容器 HTTP）：需改后端代码架构，违背 YAGNI
- 方案 C（双容器共享卷）：复杂度高，无实际收益

## 文件清单

| 文件 | 位置 | 说明 |
|---|---|---|
| `Dockerfile` | `backend/Dockerfile` | 多阶段构建，产物单一镜像 |
| `docker-compose.yml` | 仓库根目录 | compose 编排，一键启动 |
| `.env.example` | 仓库根目录 | 环境变量模板 |
| `.dockerignore` | 仓库根目录 | 排除不相关文件，减小构建 context |

## Dockerfile 详细设计

三阶段构建：

### 阶段 1 — Go 构建（`golang:1.22-bookworm`）

1. 复制 `backend/go.mod` + `backend/go.sum`
2. `go mod download`（利用 Docker 层缓存）
3. 复制 `backend/` 全部源码
4. `CGO_ENABLED=1 go build -o /build/server ./cmd/server`（`go-sqlite3` 需要 CGO）

### 阶段 2 — Python whisper-cli（`python:3.12-slim-bookworm`）

1. 安装 `uv`
2. 复制 `whisper/pyproject.toml` + `whisper/uv.lock`
3. `uv sync --no-dev --no-editable`（安装依赖和包本身到 venv）
4. 输出 venv 在 `/opt/whisper-venv/`

### 阶段 3 — 运行时（`python:3.12-slim-bookworm`）

1. `apt-get install ffmpeg`（系统包）
2. `pip install yt-dlp`（安装到系统 Python，避免加 PPA）
3. 从阶段 1 复制 `/build/server` 到 `/usr/local/bin/server`
4. 从阶段 2 复制 `/opt/whisper-venv/`
5. 设置 `PATH` 和运行时环境变量默认值：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `PATH` | `/opt/whisper-venv/bin:$PATH` | 让后端找到 `whisper-cli` |
| `HF_HOME` | `/data/huggingface` | 让 Whisper 模型缓存落在持久卷 |
| `LSI_ADDR` | `0.0.0.0:8080` | 容器内监听所有接口 |
| `LSI_DB_PATH` | `/data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `/data/jobs` | Job 工作目录 |
| `LSI_RUNNER_MODE` | `real` | 默认使用 real runner |

6. 创建数据目录 `/data/jobs` 和 `/data/huggingface`
7. 暴露端口 `8080`
8. 入口 `["/usr/local/bin/server"]`
9. `whisper-cli` 命令通过 PATH 可用（venv bin 目录已加入 PATH）
10. Whisper 模型在首次处理时自动下载到 `/data/huggingface/`，持久化在 volume 中

## docker-compose.yml 详细设计

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

设计决策：
- 端口映射默认绑定 `127.0.0.1`，不对外暴露；如需局域网访问，可将 `LSI_DOCKER_BIND_HOST` 设为 `0.0.0.0`
- named volume `lsi-data` 持久化 SQLite、job 文件、Whisper 模型缓存
- LLM 密钥通过 `.env` 文件或环境变量注入
- `restart: unless-stopped` 保证异常退出后自动重启

## .env.example

```
# Docker 端口绑定主机（默认仅本机访问；如需局域网访问可改为 0.0.0.0）
LSI_DOCKER_BIND_HOST=127.0.0.1

# Whisper 模型（首次处理视频时自动下载到持久卷）
LSI_WHISPER_MODEL=small

# LLM 配置（real runner 必填 LLM_API_KEY 和 LLM_MODEL）
LSI_LLM_BASE_URL=https://api.openai.com
LSI_LLM_API_KEY=sk-your-key-here
LSI_LLM_MODEL=gpt-4.1-mini

# 超时设置
LSI_DOWNLOAD_TIMEOUT=10m
LSI_LLM_TIMEOUT=2m
```

## .dockerignore

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

## 使用方式

```bash
# 1. 复制环境变量模板
cp .env.example .env
# 2. 编辑 .env，至少填写 LLM_API_KEY 和 LLM_MODEL
# 3. 启动
docker compose up -d
# 4. 查看日志
docker compose logs -f
# 5. 停止
docker compose down
```

## 不包含

- Docker Hub 发布或 CI/CD 流水线
- 健康检查端点（当前 API 没有此功能）
- 多容器编排（whisper sidecar 等）
- GPU 支持（faster-whisper 默认用 CPU，用户可自行添加 GPU 配置）
- 扩展（Chrome 扩展独立安装，不属于后端 Docker 范畴）
