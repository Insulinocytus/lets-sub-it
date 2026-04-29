# OpenAI-compatible 翻译链路设计

## 背景

当前 `backend/` 已经支持两种 runner：

- `LSI_RUNNER_MODE=mock`：完整状态流和三种字幕文件都由 mock runner 生成。
- `LSI_RUNNER_MODE=real`：真实调用 `yt-dlp` 下载音频，并调用 `whisper-cli` 生成真实 `source.vtt`；`translated.vtt` 和 `bilingual.vtt` 仍使用 mock 内容。

本设计补齐 real runner 的翻译阶段：解析真实 `source.vtt`，调用 OpenAI-compatible Chat Completions API 生成目标语言翻译，再基于真实转写和翻译结果生成 `translated.vtt` 与 `bilingual.vtt`。

## 目标

- 在 `LSI_RUNNER_MODE=real` 下，`translating` 阶段调用 Chat Completions 兼容接口。
- 生成的 `translated.vtt` 保持与 `source.vtt` 相同的 cue 数量和时间轴。
- 翻译单条 cue 时携带邻近上下文：目标 cue 前后各 10 条作为上下文输入，但只要求模型返回目标 cue 的翻译。
- 生成的 `bilingual.vtt` 使用同一时间轴，每个 cue 先写 source 文本，再写 translated 文本。
- 保持现有 HTTP API 响应结构不变，extension 不需要改动。
- 单元测试不依赖真实 OpenAI、真实 YouTube、真实 Whisper 模型或外部网络。

## 非目标

- 不实现 `mock` runner 的真实翻译；mock 模式继续完全离线。
- 不新增 provider registry、插件系统或多 provider 抽象。
- 不实现重试、分块并发、流式响应、成本统计、可配置上下文窗口或阶段级日志。
- 不在 extension 中保存或读取 LLM API key。
- 不放宽 backend 的本机自托管边界，也不把服务描述为公网生产服务。

## 外部 API 选择

本阶段使用 Chat Completions 兼容接口：

```text
POST {LSI_LLM_BASE_URL}/v1/chat/completions
```

请求体包含 `model`、`messages` 和 `response_format`。OpenAI 官方 Chat Completions API 参考说明该 endpoint 接收聊天消息并生成模型响应，同时支持 `response_format`；`json_object` 是兼容性较好的 JSON mode，但仍需要在 system 或 user message 中明确要求输出 JSON。

本阶段优先使用 `response_format: {"type":"json_object"}`，而不是 `json_schema`。原因是 OpenAI-compatible 服务对 `json_schema` 的支持差异更大；`json_object` 对本项目的“返回同长度字符串数组”已足够，服务端仍会严格校验返回 JSON。

## 配置

新增 backend 配置：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API origin，不含 `/v1/chat/completions` 路径 |
| `LSI_LLM_API_KEY` | 空字符串 | 可选；非空时发送 `Authorization: Bearer <key>` |
| `LSI_LLM_MODEL` | 空字符串 | real 模式翻译必填 |
| `LSI_LLM_TIMEOUT` | `2m` | 单次翻译 HTTP 请求超时 |

`LSI_RUNNER_MODE=real` 启动时仍检查 `yt-dlp`、`ffmpeg` 和 `whisper-cli`。`LSI_LLM_MODEL` 不在启动时 fatal；它在 real runner 进入翻译阶段时校验。这样本地启动错误和 job 失败语义保持简单：工具缺失阻止 real server 启动，翻译配置缺失让具体 job 失败在 `translating` 阶段。

## 组件设计

### VTT cue 处理

在 `backend/internal/runner/` 下新增轻量 VTT 处理函数：

- 解析 `source.vtt` 为 cue 列表。
- 每个 cue 保留原始时间行和文本行。
- 渲染 translated VTT：复用 source cue 时间行，文本替换为翻译结果。
- 渲染 bilingual VTT：复用 source cue 时间行，文本为 source 文本行后接 translated 文本行。

解析器只覆盖本项目 `whisper-cli` 输出的简单 WebVTT 子集：

- 文件必须以 `WEBVTT` 开头。
- cue 之间使用空行分隔。
- cue 时间行包含 `-->`。
- 支持多行 cue 文本。
- 不支持复杂 WebVTT 注释、STYLE、REGION 或 cue settings 的完整语义；如果遇到无法识别的结构，返回错误并让 job 失败。

### Translator 接口

在 runner 包内新增最小接口，便于单元测试注入：

```go
type Translator interface {
    Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error)
}
```

real runner 构造时接收 translator。正常 app 启动使用 Chat Completions translator，测试使用 fake translator。mock runner 不使用该接口。

translator 接收完整 cue 列表，而不是只接收文本数组。这样 Chat Completions translator 可以为每个目标 cue 构建上下文窗口，同时对外仍返回与 cue 数量一致的翻译数组。

### Chat Completions translator

Chat translator 负责：

- 规范化 `LSI_LLM_BASE_URL`，拼接 `/v1/chat/completions`。
- 逐条 cue 构造 JSON 请求。
- 对第 `i` 条 cue，携带 `[max(0, i-10), min(len(cues)-1, i+10)]` 范围内的上下文 cue。
- 在请求中明确标记目标 cue，要求模型只翻译目标 cue。
- 非空 `LSI_LLM_API_KEY` 时设置 `Authorization` header。
- 设置 `Content-Type: application/json`。
- 校验 HTTP 2xx。
- 解析 `choices[0].message.content`。
- 将 content 解析为 JSON 对象，并读取 `translation` 字段。
- 收集每条 cue 的翻译，最终校验翻译数量与输入 cue 数量一致。

请求的 message 语义固定为：

- system：要求模型只做字幕翻译，结合上下文理解语义，输出 JSON，不添加解释。
- user：提供 `sourceLanguage`、`targetLanguage`、`target` 和 `context`。`context` 中包含目标 cue 前后各 10 条以内的文本，目标 cue 会通过 `isTarget: true` 标记。

期望 assistant content 为：

```json
{
  "translation": "目标 cue 的翻译"
}
```

示例：翻译第 100 条 cue 时，请求会携带第 90 到第 110 条 cue 作为上下文，但 `target.index` 仍是 `100`，响应只允许包含第 100 条的翻译。

## 状态流转

real runner 的状态保持现有顺序：

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

变化点：

1. `transcribing` 成功后，`source.vtt` 已存在。
2. `translating` 解析 `source.vtt` 并调用 translator。
3. translator 成功后写入 `translated.vtt`。
4. `packaging` 使用 source cue 和 translated cue 写入 `bilingual.vtt`。
5. 创建 `SubtitleAsset`，现有 `/subtitle-assets` 与 `/subtitle-files/:jobId/:mode` 不变。

## 错误处理

以下错误会让 job 进入 `failed`，`stage` 记录为对应阶段：

| 场景 | stage |
| --- | --- |
| `source.vtt` 不存在、为空或无法解析 | `translating` |
| `LSI_LLM_MODEL` 为空 | `translating` |
| LLM HTTP 请求失败或超时 | `translating` |
| LLM 返回非 2xx | `translating` |
| LLM response JSON 无法解析 | `translating` |
| `choices[0].message.content` 为空 | `translating` |
| content 不是合法 JSON 或缺少 `translation` | `translating` |
| 收集到的翻译数量与 source cue 数量不一致 | `translating` |
| 写入 `translated.vtt` 失败 | `translating` |
| 写入 `bilingual.vtt` 失败 | `packaging` |

错误消息可以包含 HTTP status 和截断后的响应摘要，但不能包含 `LSI_LLM_API_KEY`。

## 测试策略

backend 单元测试覆盖：

- config：新增 LLM 配置默认值和环境变量读取。
- VTT：解析多个 cue、多行 cue、渲染 translated、渲染 bilingual、解析错误。
- translator：使用 `httptest.Server` 断言请求路径、headers、payload，并覆盖成功、非 2xx、无效 JSON、缺少 `translation`、上下文窗口和目标 cue 标记。
- translator 上下文窗口：覆盖中间 cue 使用前后各 10 条上下文，开头和结尾 cue 自动截断窗口。
- real runner：fake `yt-dlp` 与 `whisper-cli` 生成真实 `source.vtt`，fake translator 返回确定翻译，断言 `translated.vtt` 和 `bilingual.vtt` 不再包含 mock 翻译文本。
- real runner 失败：translator 返回错误时 job 进入 `failed`，`stage=translating`。

验证命令：

```bash
cd backend
mise exec -- go test ./...
```

## 文档更新

实现完成后同步更新：

- `README.md`：说明 real runner 已真实翻译，新增 LLM 环境变量和运行示例。
- `backend/README.md`：说明 runner boundary 和 LLM 配置。

如果实现过程中发现 API 契约、环境变量或 mock/real 边界发生变化，需要同时更新相关 `docs/` 文件。

## 已确认决策

- 用户确认优先实现 Chat Completions 兼容接口。
- 选择方案：`LSI_RUNNER_MODE=real` 直接使用真实翻译链路，`mock` 模式保持不变。
- 不为本次引入独立 `LSI_TRANSLATOR_MODE`。
- 不实现 provider registry。
