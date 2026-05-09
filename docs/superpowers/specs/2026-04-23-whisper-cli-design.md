# Whisper SDK CLI 设计

## 背景

本文定义 [PRD.md](/Users/demo/.codex/worktrees/789b/lets-sub-it/PRD.md) 中本地转写 CLI 子任务的 MVP 设计。

当前实现范围保持很小：

- 在 monorepo 根目录用 `mise use` 生成并维护统一的 `mise.toml`
- 在仓库根目录新建独立 Python CLI 项目：`whisper/`
- 由根目录 `mise.toml` 统一管理 Python 和 `uv` 版本
- 用 `uv` 管理 Python 依赖和锁文件
- 自己实现一个 `whisper-cli` 命令，命令内部直接调用 `faster-whisper` Python SDK
- 接收本地音频文件路径作为输入
- 生成经过基础校验的 `source.vtt`

Go runner、API 集成、下载链路、翻译链路归入后续设计。

## 目标

- 提供一个稳定的本地命令，后续 Go runner 可以通过 `exec` 调用
- 让契约保持简单明确：输入音频路径，输出 `source.vtt`，退出码表达结果
- 让转写项目自包含，方便本地安装、调试和依赖管理
- 对输出文件做基础校验，让下游流程可以信任成功结果

## 后续阶段范围

- 常驻转写服务
- 队列和 worker 管理
- YouTube 下载逻辑
- `ffmpeg` 集成
- 生成 VTT 以外的结果持久化
- 翻译、双语字幕打包、字幕文件服务

## 约束

- 这是一个 monorepo，工具链版本由仓库根目录统一管理
- 项目目录固定为仓库根目录下的 `whisper/`
- 根目录 `mise.toml` 负责本地工具链版本
- `uv` 负责依赖和锁文件
- CLI 保持小接口，方便后续 Go `exec` 集成
- 实现方式对齐 PRD：runner 通过退出码和输出文件判断转写阶段结果

## 方案选择

当前设计固定采用一条路径：在 `whisper/` 里实现一个小型 Python package，并注册 console entrypoint `whisper-cli`。这个命令内部直接 import `faster_whisper` SDK，加载 `WhisperModel`，转写本地音频文件，把 SDK 返回的 segments 序列化成 WebVTT，校验输出文件，再用稳定退出码结束。

这条路径的理由：

- 命令行接口和输出格式由项目控制
- 错误处理和退出码映射由项目控制
- 进程边界只存在于 Go runner 调用 `whisper-cli` 这一层
- 后续 Go 集成只依赖一个稳定命令契约

## 项目结构

```text
repo/
  mise.toml
  whisper/
    pyproject.toml
    uv.lock
    README.md
    src/
      whisper_cli/
        __init__.py
        cli.py
        transcribe.py
        vtt.py
    tests/
      test_cli.py
      test_vtt.py
```

## 文件职责

- `/mise.toml`
  - 在 monorepo 根目录统一固定 `python` 和 `uv` 版本
  - 由仓库维护者在根目录执行 `mise use` 生成和更新
- `whisper/pyproject.toml`
  - 定义项目元数据、依赖和 `whisper-cli` 脚本入口
- `whisper/uv.lock`
  - 锁定依赖版本
- `whisper/src/whisper_cli/cli.py`
  - 解析参数、执行转写流程、输出机器可读的成功结果、把异常映射成退出码
- `whisper/src/whisper_cli/transcribe.py`
  - 封装 `faster-whisper` SDK 的模型加载和转写调用
- `whisper/src/whisper_cli/vtt.py`
  - 把 segments 序列化成标准 WebVTT，并做基础校验
- `whisper/tests/test_cli.py`
  - 覆盖 CLI 契约和失败映射
- `whisper/tests/test_vtt.py`
  - 覆盖 WebVTT 格式和校验规则
- `whisper/README.md`
  - 说明本地安装、运行和后续集成方式

## CLI 契约

### 命令

前提：仓库根目录已经通过 `mise use` 固定好 `mise.toml` 里的工具链版本。

```bash
mise install
cd whisper
uv sync
uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /path/to/source.vtt \
  --model small \
  --language ja
```

### 参数

- `--input`
  - 必填
  - 本地音频文件路径
- `--output`
  - 必填
  - 目标 VTT 文件路径
- `--model`
  - 必填
  - Whisper 模型名
- `--language`
  - 必填
  - 源语言代码

第一版只暴露这 4 个参数，保持接口稳定，减少预测式配置。

## 执行流程

1. 解析 CLI 参数
2. 校验必填参数和输入文件可读性
3. 创建输出文件的父目录
4. 通过 `faster-whisper` SDK 加载配置的模型
5. 把本地音频转写成有序 segments
6. 把 segments 序列化成 WebVTT
7. 校验生成的 VTT 结构
8. 写入 `source.vtt`
9. 向 stdout 输出一行 JSON 成功结果
10. 以退出码 `0` 结束

## 输出契约

### 成功 stdout

成功时，CLI 向 stdout 写入一行 JSON。

示例：

```json
{"output":"/tmp/source.vtt","language":"ja","duration_seconds":123.4,"segments":87}
```

字段：

- `output`
  - 解析后的输出路径
- `language`
  - 用户指定的源语言
- `duration_seconds`
  - 转写得到的总时长，数字类型
- `segments`
  - 生成的字幕 cue 数量

### 失败 stderr

失败时，CLI 向 stderr 写入一行简短错误摘要。

示例：

```text
transcription failed: model download error
```

stderr 使用纯文本，便于直接进入 Go runner 日志和人工排障。

## 退出码

- `0`
  - 成功
  - `source.vtt` 已存在并通过校验
- `2`
  - 参数或输入校验失败
  - 示例：缺少必填参数、输入路径缺失、输出路径无效
- `3`
  - 转写执行失败
  - 示例：模型加载失败、音频解码失败、转写层运行时异常
- `4`
  - 输出校验失败
  - 示例：segment 列表为空、VTT 结构无效、cue 数量为 0

这些退出码数量少、语义稳定，后续 Go runner 可以直接映射成阶段失败原因。

## WebVTT 规则

生成文件需要满足以下规则：

- 文件首行是 `WEBVTT`
- cue 数量大于 0
- 每个 cue 保留原始 segment 时间
- cue 时间轴单调不降
- 每个 cue 包含字幕文本
- 输出文件存在时直接覆盖
- 输出父目录缺失时自动创建

第一版 VTT writer 保持简单：

- 省略 cue identifier
- 每个 segment 对应一个 cue block
- 文本内容来自 segment text，并去掉首尾空白

## 架构

### `transcribe.py`

这个模块负责 SDK 边界。它把 CLI 参数转换成 `faster-whisper` SDK 调用，并返回最小内存结构：

- 用户指定的语言
- 有序 segments
- 聚合时长元数据

这个模块只返回转写数据，退出码映射交给 `cli.py`。

### `vtt.py`

这个模块负责 VTT 格式化和校验。它接收标准化 segments，返回序列化后的 WebVTT 内容，或者抛出输出校验错误。

这层拆分让格式逻辑可以在加载模型之前单独测试。

### `cli.py`

这个模块负责公开命令行为：

- 解析参数
- 调用转写层
- 调用 VTT 层
- 写入输出文件
- 打印成功 JSON
- 把已知失败类型转换成稳定退出码

## 错误处理

第一版使用少量显式错误类型：

- 输入校验错误
- 转写错误
- 输出校验错误

每种错误类型映射一个退出码。CLI 打印错误摘要后退出，让人读起来清楚，也让后续 Go runner 容易消费。

## 测试与 TDD

实现阶段采用 TDD，先锁住 CLI 契约和 VTT 规则，再补最小实现。

### TDD 执行约束

每个实现点都按这个顺序推进：

1. 先写一个失败测试，明确描述目标行为
2. 运行该测试，确认失败原因符合预期
3. 写最小实现让测试通过
4. 重新运行相关测试，确认结果为绿色
5. 保持重构范围只服务当前测试覆盖的行为

### 必需测试

1. `test_vtt_writes_header_and_cues`
   - 验证 `WEBVTT`、时间轴和 cue 文本
2. `test_vtt_rejects_empty_segments`
   - 验证空 segment 输出会被拒绝
3. `test_cli_requires_all_required_arguments`
   - 验证缺少任一必填参数时退出码为 `2`
4. `test_cli_creates_parent_directory_for_output`
   - 验证嵌套输出路径会自动创建
5. `test_cli_prints_json_on_success`
   - 验证成功 stdout 的 JSON 结构稳定
6. `test_cli_returns_code_3_when_transcriber_fails`
   - 验证转写失败会映射到退出码 `3`

### 首批失败测试顺序

实现从以下失败测试开始：

1. `test_cli_requires_all_required_arguments`
2. `test_vtt_writes_header_and_cues`
3. `test_vtt_rejects_empty_segments`
4. `test_cli_creates_parent_directory_for_output`
5. `test_cli_prints_json_on_success`
6. `test_cli_returns_code_3_when_transcriber_fails`

### 后续测试

真实音频集成测试放到后续阶段。它需要模型可用、运行时间更长，并且对本地环境更敏感。第一阶段先覆盖确定性的单元测试和 CLI 契约测试。

## 假设

- 调用方提供可读的本地音频文件
- 主机环境可以在仓库根目录通过 `mise` 安装 Python 和 `uv`
- 主机环境可以安装并运行 `faster-whisper` Python SDK 所需依赖
- 第一版沿用库默认的模型下载和缓存行为

## 风险

- 首次本地运行可能需要下载模型资产
- 主机原生依赖差异可能影响安装
- 长音频会显著增加本地执行时间
- Whisper 输出质量受模型大小和源音频质量影响

## MVP 固定决策

为了降低实现歧义，第一版固定以下选择：

- package 作为仓库根目录下的独立项目：`whisper/`
- monorepo 根目录统一持有 `mise.toml`
- 依赖管理使用 `uv`
- 工具链版本管理通过根目录 `mise` 完成
- 命令只暴露 4 个必填用户参数
- 实现过程采用 TDD
- 成功运行会覆盖目标输出文件
- 成功输出使用 stdout 单行 JSON
- 失败输出使用 stderr 单行纯文本
- 空转写结果视为失败

## 完成标准

该子项目满足以下条件时视为完成：

- 在仓库根目录执行 `mise install` 可以准备统一工具链环境
- 在 `whisper/` 内执行 `uv sync` 可以准备 Python 依赖
- 首批失败测试按既定顺序落地，并在实现完成后通过
- `uv run whisper-cli --input ... --output ...` 可以生成有效 `source.vtt`
- CLI 成功和失败路径返回预期退出码
- 必需测试在本地通过
- 后续 Go runner 可以把 CLI 当成黑盒命令调用
