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
