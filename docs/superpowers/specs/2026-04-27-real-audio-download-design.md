# 真实音频下载 — 设计说明

## 概述

将 backend mock runner 的 `downloading` 阶段替换为真实 `yt-dlp` 调用，下载 YouTube 音频并转码为 MP3 128kbps。

这是真实管线四阶段中的第一步。后续 transcribing/translating/packaging 仍走 mock，逐迭代替换。

## 模式切换

通过环境变量 `LSI_RUNNER_MODE` 控制：

| 值 | downloading | 后续阶段 |
|---|------------|---------|
| `mock`（默认） | mock（sleep+空文件） | mock |
| `real` | 真实 yt-dlp 下载 | mock（本次不改动） |

## 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LSI_RUNNER_MODE` | `mock` | `mock` 或 `real` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | 单次下载超时 |

## 新增文件

### `backend/internal/runner/downloader.go`

```go
func downloadAudio(ctx context.Context, jobID, youtubeURL string) (string, error)
```

- 单次 `exec.CommandContext` 调用 `yt-dlp`
- 参数：`-x --audio-format mp3 --audio-quality 128K -o "<workDir>/<jobID>/audio.%(ext)s" <url>`
- 返回产物路径 `<workDir>/<jobID>/audio.mp3`
- 支持通过变量替换 exec 函数以进行单元测试

### `backend/internal/runner/downloader_test.go`

使用 Go FakeExec 模式（`TestMain` + 伪进程），覆盖：

| 场景 | 模拟 | 断言 |
|------|------|------|
| 正常下载 | 伪 yt-dlp 生成有效 mp3 | audio.mp3 存在且非空 |
| 视频不可用 | 退出 1 + "Video unavailable" | error 含完整 stderr |
| 下载超时 | 伪 yt-dlp sleep | context deadline exceeded |
| 网络错误 | 退出 1 + "network error" | error 含完整 stderr |
| yt-dlp 缺失 | LookPath 失败 | error 含工具缺失提示 |

## 修改文件

### `backend/internal/runner/runner.go`

`processDownloading()` 按模式分叉：

- `LSI_RUNNER_MODE=real` → 调用 `downloadAudio()`，失败则 job 置 `failed` + 完整 stderr
- `LSI_RUNNER_MODE=mock` → 现有逻辑不变

## 启动检查

Server 启动时执行 `which yt-dlp` 和 `which ffmpeg`：

| 模式 | 缺失时行为 |
|------|-----------|
| `real` | **fatal**，server 拒绝启动，打印缺失工具名并退出 |
| `mock` | **warning**，不影响启动 |

## Job 工作目录

```
<LSI_WORK_DIR>/<jobId>/
├── audio.mp3          # yt-dlp 产物（本次产出）
├── source.vtt         # 后续转写（本次不碰）
├── translated.vtt     # 后续翻译（本次不碰）
└── bilingual.vtt      # 后续双语（本次不碰）
```

## 状态流转

```
queued -> [downloading] -> transcribing -> translating -> packaging -> completed
               ↑
          LSIRUNNERMODE=real:
            yt-dlp -x --audio-format mp3 --audio-quality 128K
            context.WithTimeout → 10min
```
