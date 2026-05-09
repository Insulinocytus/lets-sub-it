# Docker 部署

Docker 后端镜像包含 Go 后端、Python `whisper-cli`、`yt-dlp` 和 `ffmpeg`，适合在本机或受控局域网环境中运行完整处理链路。

## 配置

先复制环境变量模板：

```bash
cp .env.example .env
```

`.env` 最少需要配置：

- `LSI_LLM_API_KEY`
- `LSI_LLM_MODEL`

默认情况下，Docker Compose 只绑定到 `127.0.0.1:8080`。只有确实需要局域网访问时，才将 `.env` 中的 `LSI_DOCKER_BIND_HOST` 改为 `0.0.0.0`。

不要把服务直接暴露到公网。该后端没有用户认证，设计目标是单用户本地自托管。

## 启动和停止

```bash
docker compose up -d
docker compose logs -f
docker compose down
```

## 数据卷

- `lsi-data` 保存 SQLite 数据库和 job 工作文件。
- `lsi-hf-cache` 保存 Hugging Face/Whisper 模型缓存，避免重复下载模型。
