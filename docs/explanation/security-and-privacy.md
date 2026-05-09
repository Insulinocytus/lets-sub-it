# 安全与隐私

Lets Sub It 面向单用户本地自托管，不是公开互联网多用户服务。安全边界的核心是：浏览器扩展不持有秘密，后端只暴露 localhost API，文件服务不能逃出 job 工作目录。

## 输入范围

系统只处理 `public YouTube` 视频。不要把私有视频、登录态、cookie 导入、权限绕过或受限内容访问纳入当前模型；这些能力会改变安全边界，也会让部署者承担额外风险。

## 后端边界

后端没有用户鉴权，预期运行在单用户本地自托管环境。部署者需要用主机和网络配置限制访问范围，不应把该服务当作可直接暴露到公网的生产级多租户 API。

SQLite 数据库、下载后的音频、中间转写文件和最终字幕文件都保存在本地磁盘。文件服务必须保持在 job 工作目录内，并继续保留路径穿越和符号链接逃逸防护；任何新文件访问逻辑都不能绕过这条边界。

## 密钥与 provider

Provider key 只属于 backend/server config，例如 `LSI_LLM_API_KEY`。Extension 不存储长期秘密，不直接调用翻译 provider，也不把 provider key 写入 storage、日志或消息载荷。

后端调用 OpenAI-compatible LLM 时，不应保存包含私有数据或 key 的 LLM request logs。日志可以记录阶段、job id 和错误摘要，但要避免完整请求体、敏感请求头和密钥。

## 浏览器扩展边界

Extension backend URL 保持 localhost-only，并且必须是带显式端口的 HTTP origin，例如 `http://127.0.0.1:8080` 或 `http://localhost:8080`。Manifest host permissions 也限制在 `http://127.0.0.1:*/*` 和 `http://localhost:*/*`。

Content script 只负责 YouTube 页面集成和字幕 overlay，不直接访问后端、不持有本地文件路径、不接触 provider key。所有网络请求都通过 background service worker 统一发出。
