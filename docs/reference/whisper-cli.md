# Whisper CLI

## 直接运行

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --compute-type int8 \
  --language ja
```

## 参数

必填参数：

- `--input`
- `--output`
- `--model`
- `--language`

当前使用的可选参数：

- `--compute-type`

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 成功 |
| `2` | 输入校验失败 |
| `3` | 转写失败 |
| `4` | 输出校验失败 |

## 成功输出

成功时 stdout JSON 形状：

```json
{
  "output": "/tmp/source.vtt",
  "language": "ja",
  "duration_seconds": 123.45,
  "segments": 42
}
```

## WebVTT 输出规则

`--output` 必须写出合法 WebVTT。测试使用 fake models，并且必须保持离线可重复。
