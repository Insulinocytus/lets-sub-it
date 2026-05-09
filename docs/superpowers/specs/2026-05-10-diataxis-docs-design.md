# Diátaxis 文档体系设计

## 背景

当前 `docs/` 主要由 `PRD.md` 和 `docs/superpowers/` 组成。`docs/superpowers/` 下的 `specs/` 与 `plans/` 适合记录 AI 协作过程中产生的 ticket 级或 PR 级设计与实施过程，但不适合作为新成员理解项目的长期入口。

新的文档体系面向人类读者，目标是让新成员不需要阅读每一个历史 ticket 文档，也能理解项目、跑起系统，并知道后续应该到哪里查稳定契约和设计解释。

## 目标

- 采用纯 Diátaxis 四象限组织人类长期文档。
- 优先服务新成员 onboarding。
- 保持 GitHub-first 的阅读体验，使用 `README.md` 作为目录入口。
- 将当前事实、操作步骤、稳定契约和设计解释分离维护。
- 不引入 ADR；设计理由放在 `explanation/` 中维护。

## 非目标

- 不迁移或重组 `docs/superpowers/`。
- 不要求新成员阅读 `docs/superpowers/`。
- 不把历史 ticket 文档机械提炼成长期文档。
- 不在本次设计中编写完整内容，只确定结构、职责边界和迁移策略。

## 核心决策

人类长期文档采用 Diátaxis 四象限：

- `tutorials/`：按顺序学习的教程。
- `how-to/`：完成具体任务的操作指南。
- `reference/`：稳定、准确、可查表的契约。
- `explanation/`：解释系统为什么这样工作。

`docs/superpowers/` 是 AI 生成文档的存放处，不属于人类长期文档的信息架构。它可以保留在 `docs/` 下，但不出现在新人阅读路径中。

## 目录结构

```text
docs/
  README.md

  tutorials/
    README.md
    onboarding.md

  how-to/
    README.md
    local-development.md
    docker-deployment.md
    testing.md
    troubleshooting.md

  reference/
    README.md
    repository-structure.md
    backend-api.md
    backend-config.md
    data-and-job-model.md
    whisper-cli.md
    extension-contract.md

  explanation/
    README.md
    architecture-overview.md
    processing-pipeline.md
    module-boundaries.md
    security-and-privacy.md

  superpowers/
    specs/
    plans/
```

## 入口规则

`docs/README.md` 是人类文档的总入口，负责说明四象限的用途和推荐阅读路径。

推荐路径：

1. 新成员先读 `tutorials/onboarding.md`。
2. 需要跑项目时读 `how-to/local-development.md` 或 `how-to/docker-deployment.md`。
3. 需要查契约时读 `reference/` 下的对应文档。
4. 需要理解系统设计理由时读 `explanation/` 下的对应文档。

每个子目录都使用 `README.md` 作为目录页，以便 GitHub 在进入目录时自动渲染说明。

## 各象限职责

### `tutorials/`

面向新成员的学习路径。第一阶段只需要 `onboarding.md`。

`onboarding.md` 应覆盖：

- 项目解决的问题。
- `backend`、`whisper`、`extension` 的协作关系。
- 如何选择本地开发或 Docker 路径，并链接到对应 how-to 文档。
- 常见改动应先查看哪些模块和 reference 文档。
- 完成标准：能描述处理链路，能启动相关模块，知道按模块查文档。

### `how-to/`

面向具体任务，只写操作步骤，不承载完整背景解释。

- `local-development.md`：本地启动后端、Whisper CLI 和扩展。
- `docker-deployment.md`：Docker 后端部署流程。
- `testing.md`：按模块运行测试、类型检查和构建验证。
- `troubleshooting.md`：常见失败与处理，例如缺少工具、端口占用、LLM 配置、模型下载问题。

### `reference/`

面向查阅稳定契约，内容应准确、完整、少解释。

- `repository-structure.md`：目录职责和关键入口文件。
- `backend-api.md`：HTTP API 请求、响应和错误模型。
- `backend-config.md`：环境变量、默认值和运行时要求。
- `data-and-job-model.md`：Job 状态、数据模型和字幕资产。
- `whisper-cli.md`：CLI 参数、stdout/stderr、退出码和 WebVTT 契约。
- `extension-contract.md`：扩展存储、消息协议、backend URL 约束和 subtitle modes。

### `explanation/`

解释系统设计和边界，承接长期设计理由，不记录 ticket 执行过程。

- `architecture-overview.md`：整体架构和模块关系。
- `processing-pipeline.md`：下载、转写、翻译、打包、扩展渲染的完整链路。
- `module-boundaries.md`：后端、Whisper CLI、扩展之间的责任边界。
- `security-and-privacy.md`：公开 YouTube、本地单用户、localhost-only、密钥边界。

## 迁移策略

第一步只从现有 `README.md`、`docs/PRD.md`、`backend/README.md`、`whisper/README.md`、`extension/README.md` 提炼当前仍然正确的事实。

第二步保持根目录 `README.md` 作为项目门面和快速开始入口，将详细内容链接到 `docs/` 下对应象限。

第三步后续修改行为契约时，按文档类型更新对应象限：

- 命令流程变化更新 `how-to/`。
- API、配置、CLI、数据模型变化更新 `reference/`。
- 架构边界、安全模型、处理链路变化更新 `explanation/`。
- 新人路径变化更新 `tutorials/onboarding.md` 和 `docs/README.md`。

## 完成标准

- 新成员只读 `docs/README.md`、`tutorials/onboarding.md` 和必要的 how-to，就能理解项目并跑起系统。
- 日常开发者能在 `reference/` 中查到稳定契约，不需要从历史 ticket 文档中拼信息。
- 设计理由集中在 `explanation/`，不通过 ADR 单独维护。
- `docs/superpowers/` 保留为 AI 生成档案区，不进入人类文档主路径。
