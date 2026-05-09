# 测试与构建验证

测试必须保持离线友好、可重复执行。单元测试不能使用真实 YouTube、模型下载、GPU、外部 LLM 或 provider key。

## 全量测试

从仓库根目录执行：

```bash
(cd backend && mise exec -- go test ./...)
(cd whisper && mise exec -- uv run pytest)
(cd extension && mise exec -- npm run test)
```

## 聚焦测试

每条命令都从仓库根目录单独执行：

```bash
(cd backend && mise exec -- go test ./internal/api)
(cd backend && mise exec -- go test ./internal/runner -run TestRealRunnerCompletesJob)
(cd whisper && mise exec -- uv run pytest tests/test_vtt.py)
(cd whisper && mise exec -- uv run pytest -k "vtt")
(cd extension && mise exec -- npx vitest run src/api/backend-client.test.ts)
(cd extension && mise exec -- npm run typecheck)
```

## 构建验证

从仓库根目录执行：

```bash
(cd backend && mise exec -- go build ./...)
(cd whisper && mise exec -- uv build)
(cd extension && mise exec -- npm run build)
```
