# 数据与任务模型

## 任务状态

正常状态流：

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

失败时任务状态为 `failed`，并在 `errorMessage` 记录错误摘要。

## `Job` 字段

- `id`
- `videoId`
- `youtubeUrl`
- `sourceLanguage`
- `targetLanguage`
- `status`
- `stage`
- `progressText`
- `errorMessage`
- `attempt`
- `workingDir`
- `createdAt`
- `updatedAt`

## `SubtitleAsset` 字段

- `jobId`
- `videoId`
- `targetLanguage`
- `sourceLanguage`
- `sourceVttPath`
- `translatedVttPath`
- `bilingualVttPath`
- `createdAt`

`sourceVttPath`、`translatedVttPath`、`bilingualVttPath` 是后端本地路径，不直接暴露给扩展。

## 扩展缓存条目

`SubtitleAssetCacheEntry` 字段：

- `jobId`
- `videoId`
- `targetLanguage`
- `sourceLanguage`
- `files`
- `createdAt`
- `selectedMode`
- `lastSyncedAt`

`files` 包含：

- `source`
- `translated`
- `bilingual`

`VideoPreference` 字段：

- `videoId`
- `targetLanguage`
- `selectedMode`

缓存键包含 `backendBaseUrl`。字幕产物缓存按 `backendBaseUrl + videoId + targetLanguage` 区分；视频偏好按 `backendBaseUrl + videoId` 区分。

## 复用键

后端任务和产物复用键为 `videoId + targetLanguage`。
