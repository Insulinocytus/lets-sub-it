# Graph Report - lets-sub-it  (2026-05-10)

## Corpus Check
- 102 files · ~32,625 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 409 nodes · 632 edges · 20 communities detected
- Extraction: 71% EXTRACTED · 29% INFERRED · 0% AMBIGUOUS · INFERRED: 186 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]

## God Nodes (most connected - your core abstractions)
1. `main()` - 18 edges
2. `LoadConfig()` - 17 edges
3. `openTestStore()` - 17 edges
4. `NewJob()` - 14 edges
5. `handleExtensionMessage()` - 12 edges
6. `newTestServer()` - 11 edges
7. `Segment` - 11 edges
8. `transcribe_audio()` - 11 edges
9. `Store` - 10 edges
10. `render_vtt()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `cacheAndNotify()` --calls--> `getSettings()`  [INFERRED]
  extension/src/api/job-monitor.ts → extension/src/storage/settings.ts
- `getMonitorBackendBaseUrl()` --calls--> `getSettings()`  [INFERRED]
  extension/src/api/job-monitor.ts → extension/src/storage/settings.ts
- `getMonitorClient()` --calls--> `createBackendClient()`  [INFERRED]
  extension/src/api/job-monitor.ts → extension/src/api/backend-client.ts
- `getMonitorBackendBaseUrl()` --calls--> `normalizeBackendBaseUrl()`  [INFERRED]
  extension/src/api/job-monitor.ts → extension/src/api/backend-client.ts
- `createMonitorKey()` --calls--> `normalizeBackendBaseUrl()`  [INFERRED]
  extension/src/api/job-monitor.ts → extension/src/api/backend-client.ts

## Communities

### Community 0 - "Community 0"
Cohesion: 0.09
Nodes (25): completingRunner, NewHandler(), newTestServer(), TestActiveJobReturnsLatestJobForVideoAndLanguage(), TestActiveJobReturnsNullWhenNoJobExists(), TestPostJobsCreatesJobAndCompletes(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion() (+17 more)

### Community 1 - "Community 1"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 2 - "Community 2"
Cohesion: 0.16
Nodes (25): execCall, fakeTranslator, NewRealRunner(), argValue(), assertArg(), findExecCall(), TestRealRunnerCompletesJob(), TestRealRunnerDownloadFailed() (+17 more)

### Community 3 - "Community 3"
Cohesion: 0.14
Nodes (18): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+10 more)

### Community 4 - "Community 4"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 5 - "Community 5"
Cohesion: 0.14
Nodes (14): downloadAudio(), TestDownloadAudioCreatesJobDir(), TestDownloadAudioNetworkError(), TestDownloadAudioSuccess(), TestDownloadAudioTimeout(), TestDownloadAudioVideoUnavailable(), TestDownloadAudioYtDlpMissing(), logJobCompleted() (+6 more)

### Community 6 - "Community 6"
Cohesion: 0.19
Nodes (17): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+9 more)

### Community 7 - "Community 7"
Cohesion: 0.15
Nodes (14): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, translationCue, translationPrompt, translationResponse (+6 more)

### Community 8 - "Community 8"
Cohesion: 0.2
Nodes (17): fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_passes_compute_type_to_transcriber(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails(), test_cli_returns_code_2_when_input_file_is_not_readable() (+9 more)

### Community 9 - "Community 9"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 10 - "Community 10"
Cohesion: 0.13
Nodes (6): openTestStore(), foreignKeyDSN(), Open(), openWithLogWriter(), TestStoreDoesNotLogExpectedRecordNotFound(), Store

### Community 11 - "Community 11"
Cohesion: 0.21
Nodes (12): Exception, FakeModel, FakeSegment, test_transcribe_audio_allows_empty_segment_output(), test_transcribe_audio_passes_compute_type_to_sdk(), test_transcribe_audio_rejects_english_only_model_with_non_english_language(), test_transcribe_audio_uses_sdk_and_builds_result(), test_transcribe_audio_uses_sdk_reported_duration() (+4 more)

### Community 12 - "Community 12"
Cohesion: 0.25
Nodes (13): checkTools(), NewHTTPHandler(), TestCheckToolsRequiresWhisperCLI(), TestNewHTTPHandlerRequiresToolsByDefault(), NewChatTranslator(), makeTranslatorTestCues(), TestChatTranslatorFailsOnNon2xx(), TestChatTranslatorFailsWhenTranslationMissing() (+5 more)

### Community 13 - "Community 13"
Cohesion: 0.22
Nodes (14): Cue, newTranslationPrompt(), cueText(), nonEmptyLines(), parseWebVTTCues(), renderBilingualVTT(), renderTranslatedVTT(), TestParseWebVTTCuesParsesMultipleCuesAndMultilineText() (+6 more)

### Community 14 - "Community 14"
Cohesion: 0.32
Nodes (10): test_vtt_rejects_blank_lines_in_cue_text(), test_vtt_rejects_empty_segments(), test_vtt_rejects_near_zero_negative_timestamps(), test_vtt_rejects_negative_timestamps(), test_vtt_rejects_timestamp_separator_in_cue_text(), test_vtt_rejects_zero_duration_after_millisecond_rounding(), test_vtt_writes_header_and_cues(), format_timestamp_from_milliseconds() (+2 more)

### Community 15 - "Community 15"
Cohesion: 0.32
Nodes (5): ConfigureLogger(), slogLevel(), TestConfigureLoggerDefaultsInvalidLevelToInfo(), TestConfigureLoggerFiltersByLevel(), main()

### Community 16 - "Community 16"
Cohesion: 0.39
Nodes (6): ParseVideoID(), TestParseVideoIDRejectsNonWatchYouTubeURL(), TestParseVideoIDRejectsUnsupportedScheme(), TestParseVideoIDRejectsUnsupportedURL(), TestParseVideoIDSupportsShortURL(), TestParseVideoIDSupportsWatchURL()

### Community 17 - "Community 17"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 19 - "Community 19"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 20 - "Community 20"
Cohesion: 0.67
Nodes (2): Runner, Store

## Knowledge Gaps
- **23 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+18 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 19`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 20`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewHTTPHandler()` connect `Community 12` to `Community 0`, `Community 10`, `Community 2`, `Community 15`?**
  _High betweenness centrality (0.152) - this node is a cross-community bridge._
- **Why does `NewRealRunner()` connect `Community 2` to `Community 12`, `Community 5`?**
  _High betweenness centrality (0.088) - this node is a cross-community bridge._
- **Why does `NewChatTranslator()` connect `Community 12` to `Community 7`?**
  _High betweenness centrality (0.067) - this node is a cross-community bridge._
- **Are the 15 inferred relationships involving `main()` (e.g. with `test_cli_requires_all_required_arguments()` and `test_cli_creates_parent_directory_for_output()`) actually correct?**
  _`main()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 14 inferred relationships involving `LoadConfig()` (e.g. with `main()` and `TestLoadConfigUsesDefaults()`) actually correct?**
  _`LoadConfig()` has 14 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `openTestStore()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerLogsJobLifecycle()`) actually correct?**
  _`openTestStore()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `NewJob()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerLogsJobLifecycle()`) actually correct?**
  _`NewJob()` has 13 INFERRED edges - model-reasoned connections that need verification._