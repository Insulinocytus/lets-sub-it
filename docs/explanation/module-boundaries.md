# 模块边界

Lets Sub It 的模块边界按运行环境和责任划分。每个模块只拥有自己能可靠控制的事情，跨边界协作通过稳定接口完成。

## 后端

`backend` 拥有 HTTP 请求解析、响应映射、CORS、SQLite 持久化、job 生命周期、runner 调度、外部工具调用编排和字幕文件服务。它是唯一知道本地数据库路径、job 工作目录和 provider key 的模块。

后端同时负责把本地文件安全映射成 HTTP 字幕响应。前端使用 `/subtitle-files/:jobId/:mode` 获取字幕内容，而不是读取或缓存本地绝对路径。

## Whisper CLI

`whisper` 拥有本地音频到 WebVTT 的转写能力，以及 `whisper-cli` 的命令行参数、stdout JSON 摘要和退出码契约。Go runner 只把本地音频路径、输出路径、模型和语言传给 CLI，不直接耦合 `faster-whisper` SDK。

这种边界让转写测试可以使用 fake model，也让后端测试可以用 stub 命令覆盖外部进程行为。

## 扩展

`extension` 拥有 popup、设置与字幕缓存、background message gateway、任务轮询、YouTube watch 页面识别和字幕 overlay。Popup 和 content script 只通过 runtime message 与 background service worker 协作，background 是扩展内唯一 HTTP API 网关。

## 协作规则

- `content scripts` 不直接调用后端；content script 必须通过 background service worker 获取任务状态、字幕资产和 VTT 内容。
- 前端使用 `/subtitle-files/:jobId/:mode`，不使用后端本地路径。
- 测试必须用 fakes 或 stubs 隔离外部工具、模型下载、GPU、真实 YouTube 和 LLM。
- 除非明确要求，不引入额外 queue framework；当前长任务由 Go 服务内的 embedded runner 和 goroutine 执行。
