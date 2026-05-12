# Lets Sub It Extension

Chrome Manifest V3 extension，用于提交 YouTube 字幕任务，并在 YouTube watch 页面渲染已完成的字幕。

## Setup

```bash
mise install
cd extension
mise exec -- npm install
```

## Development

```bash
cd extension
mise exec -- npm run dev
```

在 Chrome extension developer mode 中加载 WXT 生成的 Chrome extension 输出目录。

## Test

```bash
cd extension
mise exec -- npm run test
```

## Build

```bash
cd extension
mise exec -- npm run build
```

CI 会在 extension 相关代码更新后构建 Chrome MV3 extension，并上传 GitHub Actions artifact。进入对应 workflow run，下载 `lets-sub-it-extension-chrome-<run-number>`，解压后在 `chrome://extensions` 中开启开发者模式，选择包含 `manifest.json` 的解压目录加载。

## Backend

popup 默认连接 `http://127.0.0.1:8080`。第一版 MVP 只支持 `localhost` 和 `127.0.0.1` backend URLs。
popup 是临时页面，关闭后会在下次打开时通过 background 查询 `GET /jobs/active?videoId=...&targetLanguage=...` 恢复当前视频的任务进度。
