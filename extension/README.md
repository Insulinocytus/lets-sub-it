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

## Backend

popup 默认连接 `http://127.0.0.1:8080`。第一版 MVP 只支持 `localhost` 和 `127.0.0.1` backend URLs。
