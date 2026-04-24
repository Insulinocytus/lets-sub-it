# AGENTS.md

## 项目概览

Lets Sub It 是一个自托管的 YouTube 字幕生成与翻译工具。目标链路是：提交公开视频链接，下载音频，本地转写，翻译，生成字幕，并由播放页加载字幕。

当前仓库仍处于 MVP 阶段，已落地且可运行的实现集中在 `whisper/`：一个 Python 3.12 + `faster-whisper` 的本地 CLI，用于把本地音频文件转成经过校验的 WebVTT 字幕。

主要目录：

- `whisper/`：Python 包 `whisper-cli`，包含 CLI、转写适配、VTT 渲染与 pytest 测试。
- `docs/`：PRD、规格说明与实施计划。该目录下另有 `docs/AGENTS.md`，要求文档正文使用中文。
- `backend/`：后续 Go 服务端预留目录，当前没有可运行实现。
- `extension/`：后续 Chrome extension 预留目录，当前没有项目脚本或源码实现。
- `mise.toml`：本地工具链版本，包含 Python 3.12 与 uv。

## 协作原则

- 始终使用简体中文与用户沟通。
- 优先做最小可行改动，不添加未被要求的功能、抽象或配置。
- 修改现有代码时保持手术式变更：只触碰与任务直接相关的文件和行。
- 如果需求存在多种解释，先说明假设；不确定且风险较高时先询问。
- 不要清理无关代码、重排无关格式或重构未被要求的模块。

## 环境与依赖

AI agent 的 shell 沙盒环境不会自动执行 `eval "$(mise activate zsh)"`。凡是使用 `mise.toml` 中锁定版本的工具，都必须通过 `mise exec --` 前缀调用，例如 `mise exec -- uv`、`mise exec -- python`。

从仓库根目录安装工具链：

```bash
mise install
```

同步 Python 依赖：

```bash
cd whisper
mise exec -- uv sync --dev
```

真实转写会调用 `faster-whisper`，可能需要模型下载能力以及本机可用的推理运行环境。单元测试使用 fake model，不会下载或运行真实 Whisper 模型。

## 开发工作流

运行本地 CLI 示例：

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

成功时 stdout 输出 JSON，`--output` 写入 WebVTT 文件。CLI 退出码契约：

- `0`：成功。
- `2`：输入校验失败，例如输入文件不存在、模型名或语言无效。
- `3`：转写失败。
- `4`：输出校验失败，例如无法生成合法 VTT。

## 测试说明

运行全部测试：

```bash
cd whisper
mise exec -- uv run pytest
```

运行单个测试文件：

```bash
cd whisper
mise exec -- uv run pytest tests/test_vtt.py
```

按测试名聚焦运行：

```bash
cd whisper
mise exec -- uv run pytest -k "vtt"
```

测试配置在 `whisper/pyproject.toml`：`pythonpath = ["src"]`，默认 `addopts = "-q"`。测试目录是 `whisper/tests/`，文件命名遵循 `test_*.py`。

改动 `whisper/src/whisper_cli/` 下行为时，应优先添加或更新相邻 pytest 测试，然后确认相关测试通过。

## 构建说明

验证 Python 包构建：

```bash
cd whisper
mise exec -- uv build
```

构建产物输出到 `whisper/dist/`，该目录由 `whisper/.gitignore` 忽略。不要手动提交构建产物。

## 代码风格

- Python 代码使用 3.12 语法，源码位于 `whisper/src/whisper_cli/`。
- 保持现有风格：类型标注、`from __future__ import annotations`、四空格缩进、简洁函数。
- CLI 入口在 `whisper/src/whisper_cli/cli.py`，脚本名由 `whisper/pyproject.toml` 的 `[project.scripts]` 暴露为 `whisper-cli`。
- 转写 SDK 适配在 `whisper/src/whisper_cli/transcribe.py`；WebVTT 时间轴和 cue 校验在 `whisper/src/whisper_cli/vtt.py`。
- 不要为单次使用创建额外抽象；不要引入新的依赖，除非任务明确需要。
- 当前没有配置统一 formatter 或 linter；不要仅为格式化而大范围改动文件。

## 文档规则

- 根目录 `README.md` 面向人类读者，`AGENTS.md` 面向 coding agent。
- `docs/` 下所有说明性正文必须使用中文；代码、命令、路径、配置键、API 名称和必须保留的引用可以使用英文。
- 更新行为契约、命令或目录结构时，同步检查 `README.md`、`whisper/README.md` 与相关 `docs/` 文件是否需要更新。

## Pull Request 指南

- 使用 `.github/pull_request_template.md` 填写资料、概要、Close Issue、按文件说明、动作确认和重点 Review 项目。
- 提交 PR 前至少运行与变更相关的测试；涉及 `whisper/` 包行为时运行 `cd whisper && mise exec -- uv run pytest`。
- 涉及打包、入口点或依赖配置时额外运行 `cd whisper && mise exec -- uv build`。
- PR 描述中列出实际执行过的验证命令和结果。

## 常见注意事项

- `backend/` 与 `extension/` 目前是预留目录，不要假设已有可运行服务或前端构建流程。
- 不要把真实音频样本、模型文件、下载缓存、`.env` 或构建产物提交到仓库。
- 真实转写可能访问网络下载模型；单元测试不应依赖网络、模型下载或本地 GPU。
- `uv.lock` 已提交，修改 Python 依赖时应通过 uv 更新锁文件，而不是手动编辑锁文件。
