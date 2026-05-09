# 扩展契约

## 技术栈

- WXT
- Vue
- TypeScript
- Vitest
- Tailwind/shadcn-vue

## 边界

- background service worker 是唯一 HTTP API gateway。
- content scripts 不得直接调用后端。
- 扩展永不存储 provider keys。

## 后端地址规则

扩展只接受以下带端口的本地 HTTP origin：

- `http://localhost:<port>`
- `http://127.0.0.1:<port>`

## Manifest host permissions

- `http://127.0.0.1:*/*`
- `http://localhost:*/*`

## 语言和字幕模式

支持语言：

- `en`
- `zh`

`sourceLanguage` 必须不等于 `targetLanguage`。

YouTube watch page 字幕模式：

- `translated`
- `bilingual`

后端同时提供 `source` 文件模式。

## 消息协议

- `settings:get`
- `settings:update`
- `job:create`
- `job:get`
- `job:active`
- `subtitle:resolve`
- `subtitle:fetch-file`
- `subtitle:update-mode`
