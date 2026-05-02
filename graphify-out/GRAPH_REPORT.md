# Graph Report - .  (2026-05-02)

## Corpus Check
- 124 files · ~56,617 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 482 nodes · 718 edges · 21 communities detected
- Extraction: 74% EXTRACTED · 26% INFERRED · 0% AMBIGUOUS · INFERRED: 189 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Backend API Tests|Backend API Tests]]
- [[_COMMUNITY_Runtime Contracts|Runtime Contracts]]
- [[_COMMUNITY_MVP Design Docs|MVP Design Docs]]
- [[_COMMUNITY_Extension Backend Client|Extension Backend Client]]
- [[_COMMUNITY_Whisper Transcription Tests|Whisper Transcription Tests]]
- [[_COMMUNITY_Backend Runner Store|Backend Runner Store]]
- [[_COMMUNITY_Backend API Types|Backend API Types]]
- [[_COMMUNITY_Extension Job Monitor|Extension Job Monitor]]
- [[_COMMUNITY_Real Runner VTT|Real Runner VTT]]
- [[_COMMUNITY_Chat Translation|Chat Translation]]
- [[_COMMUNITY_Backend Config|Backend Config]]
- [[_COMMUNITY_YouTube Overlay|YouTube Overlay]]
- [[_COMMUNITY_Code When Output|Code When Output]]
- [[_COMMUNITY_Mock Runner Newmockrunner|Mock Runner Newmockrunner]]
- [[_COMMUNITY_Sqlite Foreignkeydsn Open|Sqlite Foreignkeydsn Open]]
- [[_COMMUNITY_Downloader Downloadaudio Testdownloadaudiocreatesjobdir|Downloader Downloadaudio Testdownloadaudiocreatesjobdir]]
- [[_COMMUNITY_Youtube Parsevideoid Testparsevideoidrejectsnonwatchyoutubeurl|Youtube Parsevideoid Testparsevideoidrejectsnonwatchyoutubeurl]]
- [[_COMMUNITY_Page Watch Video|Page Watch Video]]
- [[_COMMUNITY_Agent Operating Rules|Agent Operating Rules]]
- [[_COMMUNITY_Vtt Parsetimestamp Parsevtt|Vtt Parsetimestamp Parsevtt]]
- [[_COMMUNITY_Runner Store|Runner Store]]

## God Nodes (most connected - your core abstractions)
1. `NewJob()` - 17 edges
2. `main()` - 17 edges
3. `openTestStore()` - 16 edges
4. `LoadConfig()` - 15 edges
5. `handleExtensionMessage()` - 12 edges
6. `newTestServer()` - 12 edges
7. `Segment` - 11 edges
8. `Store` - 10 edges
9. `render_vtt()` - 10 edges
10. `transcribe_audio()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `Claude Agent Operating Rules` --semantically_similar_to--> `Agent Operating Rules`  [INFERRED] [semantically similar]
  CLAUDE.md → AGENTS.md
- `Backend Contract` --conceptually_related_to--> `Backend Module`  [EXTRACTED]
  AGENTS.md → README.md
- `Backend API Server` --implements--> `Backend Module`  [EXTRACTED]
  backend/README.md → README.md
- `Docker Compose Backend Service` --implements--> `Real Runner`  [EXTRACTED]
  docker-compose.yml → README.md
- `MVP Scope` --references--> `Real Runner`  [EXTRACTED]
  docs/PRD.md → README.md

## Hyperedges (group relationships)
- **End To End Subtitle Flow** — readme_extension_module, docs_prd_background_service_worker, backend_readme_api_server, docs_prd_embedded_runner, docs_prd_whisper_cli_contract, docs_prd_openai_compatible_llm, readme_webvtt_artifacts, docs_prd_content_script_overlay [EXTRACTED 1.00]
- **Job Reuse And Cache Contract** — docs_prd_job, docs_prd_subtitle_asset, docs_prd_local_cache_entry, docs_prd_video_preference, extension_readme_popup_restore [EXTRACTED 1.00]
- **Real Translation Runner Pattern** — openai_plan_llm_config, openai_plan_vtt_cue, openai_plan_translator_interface, openai_plan_chat_translator, openai_plan_realrunner_translation_integration [EXTRACTED 1.00]
- **Real Runner End-to-End Pipeline** — 2026_04_27_real_audio_download_design_yt_dlp_download, 2026_04_23_whisper_cli_design_whisper_cli, 2026_04_29_openai_compatible_translation_design_real_translation_pipeline, 2026_04_24_backend_mock_mvp_design_http_api_contract [INFERRED 0.86]
- **Extension Subtitle Runtime Flow** — 2026_04_25_extension_mvp_plan_background_gateway, 2026_04_25_extension_mvp_plan_storage_cache, 2026_04_25_extension_mvp_plan_youtube_overlay, 2026_04_25_extension_mvp_design_message_protocol, 2026_04_25_extension_mvp_design_webvtt_parser [EXTRACTED 1.00]
- **Docker Real Runner Runtime Bundle** — 2026_05_02_docker_deployment_design_single_image, 2026_05_02_docker_deployment_design_lsi_data_volume, 2026_05_02_docker_deployment_design_localhost_binding, 2026_05_02_docker_deployment_plan_backend_dockerfile, 2026_05_02_docker_deployment_plan_docker_compose [EXTRACTED 1.00]

## Communities

### Community 0 - "Backend API Tests"
Cohesion: 0.08
Nodes (28): NewHandler(), newTestServer(), TestActiveJobReturnsLatestJobForVideoAndLanguage(), TestActiveJobReturnsNullWhenNoJobExists(), TestPostJobsCreatesJobAndCompletesWithMockRunner(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion(), TestSubtitleFileRejectsInvalidMode() (+20 more)

### Community 1 - "Runtime Contracts"
Cohesion: 0.07
Nodes (40): Backend Contract, Extension Contract, Backend API Server, Backend Runner Boundary, Docker Compose Backend Service, lsi-data Docker Volume, Background Service Worker, YouTube Content Script Overlay (+32 more)

### Community 2 - "MVP Design Docs"
Cohesion: 0.07
Nodes (39): faster-whisper SDK, WebVTT 输出校验, whisper-cli, Backend Mock MVP 设计, Backend HTTP API 契约, Job 数据模型, SubtitleAsset 数据模型, Backend Mock MVP 实施计划 (+31 more)

### Community 3 - "Extension Backend Client"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 4 - "Whisper Transcription Tests"
Cohesion: 0.14
Nodes (21): Exception, FakeModel, FakeSegment, test_transcribe_audio_allows_empty_segment_output(), test_transcribe_audio_rejects_english_only_model_with_non_english_language(), test_transcribe_audio_uses_sdk_and_builds_result(), test_transcribe_audio_uses_sdk_reported_duration(), test_vtt_rejects_blank_lines_in_cue_text() (+13 more)

### Community 5 - "Backend Runner Store"
Cohesion: 0.16
Nodes (24): execCall, fakeTranslator, NewRealRunner(), argValue(), assertArg(), findExecCall(), TestRealRunnerCompletesJob(), TestRealRunnerDownloadFailed() (+16 more)

### Community 6 - "Backend API Types"
Cohesion: 0.14
Nodes (18): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+10 more)

### Community 7 - "Extension Job Monitor"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 8 - "Real Runner VTT"
Cohesion: 0.15
Nodes (16): Cue, RealRunner, transcribeAudio(), newTranslationPrompt(), cueText(), nonEmptyLines(), parseWebVTTCues(), renderBilingualVTT() (+8 more)

### Community 9 - "Chat Translation"
Cohesion: 0.15
Nodes (17): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, translationCue, translationPrompt, translationResponse (+9 more)

### Community 10 - "Backend Config"
Cohesion: 0.18
Nodes (16): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+8 more)

### Community 11 - "YouTube Overlay"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 12 - "Code When Output"
Cohesion: 0.22
Nodes (16): fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails(), test_cli_returns_code_2_when_input_file_is_not_readable(), test_cli_returns_code_2_when_language_code_is_invalid() (+8 more)

### Community 13 - "Mock Runner Newmockrunner"
Cohesion: 0.18
Nodes (9): NewMockRunner(), openTestStore(), TestMockRunnerCompletesJobAndWritesAssets(), TestMockRunnerFailsJobWhenCompletionUpdateFails(), TestMockRunnerMarksCanceledJobAsFailed(), TestMockRunnerMarksFailedWhenTranscribingUpdateFails(), MockRunner, recordingStore (+1 more)

### Community 14 - "Sqlite Foreignkeydsn Open"
Cohesion: 0.15
Nodes (5): foreignKeyDSN(), Open(), openWithLogWriter(), TestStoreDoesNotLogExpectedRecordNotFound(), Store

### Community 15 - "Downloader Downloadaudio Testdownloadaudiocreatesjobdir"
Cohesion: 0.36
Nodes (7): downloadAudio(), TestDownloadAudioCreatesJobDir(), TestDownloadAudioNetworkError(), TestDownloadAudioSuccess(), TestDownloadAudioTimeout(), TestDownloadAudioVideoUnavailable(), TestDownloadAudioYtDlpMissing()

### Community 16 - "Youtube Parsevideoid Testparsevideoidrejectsnonwatchyoutubeurl"
Cohesion: 0.39
Nodes (6): ParseVideoID(), TestParseVideoIDRejectsNonWatchYouTubeURL(), TestParseVideoIDRejectsUnsupportedScheme(), TestParseVideoIDRejectsUnsupportedURL(), TestParseVideoIDSupportsShortURL(), TestParseVideoIDSupportsWatchURL()

### Community 17 - "Page Watch Video"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 19 - "Agent Operating Rules"
Cohesion: 0.5
Nodes (4): Agent Operating Rules, Claude Agent Operating Rules, Docs Chinese Language Policy, Docs Claude Chinese Language Policy

### Community 20 - "Vtt Parsetimestamp Parsevtt"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 21 - "Runner Store"
Cohesion: 0.67
Nodes (2): Runner, Store

## Knowledge Gaps
- **45 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+40 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Vtt Parsetimestamp Parsevtt`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Runner Store`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewHTTPHandler()` connect `Backend API Tests` to `Backend Runner Store`, `Chat Translation`, `Backend Config`, `Mock Runner Newmockrunner`, `Sqlite Foreignkeydsn Open`?**
  _High betweenness centrality (0.090) - this node is a cross-community bridge._
- **Why does `NewRealRunner()` connect `Backend Runner Store` to `Real Runner VTT`, `Backend API Tests`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **Why does `NewJob()` connect `Backend Runner Store` to `Mock Runner Newmockrunner`, `Backend API Types`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 16 inferred relationships involving `NewJob()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`NewJob()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **Are the 14 inferred relationships involving `main()` (e.g. with `test_cli_requires_all_required_arguments()` and `test_cli_creates_parent_directory_for_output()`) actually correct?**
  _`main()` has 14 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `openTestStore()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`openTestStore()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 12 inferred relationships involving `LoadConfig()` (e.g. with `main()` and `TestLoadConfigUsesDefaults()`) actually correct?**
  _`LoadConfig()` has 12 INFERRED edges - model-reasoned connections that need verification._