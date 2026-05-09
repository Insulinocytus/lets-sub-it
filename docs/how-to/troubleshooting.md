# 排障

| 问题 | 检查项 |
| --- | --- |
| 后端启动失败并提示工具缺失 | 确认 `yt-dlp`、`ffmpeg` 和 `whisper-cli` 在 `PATH` 中 |
| job 在 `translating` 阶段失败 | 确认 `LSI_LLM_BASE_URL`、`LSI_LLM_MODEL` 已配置；OpenAI 默认 endpoint 还需要 `LSI_LLM_API_KEY` |
| 扩展无法连接后端 | 确认 backend URL 是带端口的本机 HTTP origin，且不包含路径或查询参数 |
| Whisper 首次运行很慢 | 真实转写可能触发模型下载；单元测试不会下载模型或依赖 GPU |
| 字幕文件返回 404 | 确认任务已完成，且 `mode` 是 `source`、`translated` 或 `bilingual` |
