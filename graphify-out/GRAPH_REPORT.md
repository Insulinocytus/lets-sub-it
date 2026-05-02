# Graph Report - .  (2026-05-02)

## Corpus Check
- 125 files · ~57,359 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 703 nodes · 992 edges · 40 communities detected
- Extraction: 77% EXTRACTED · 23% INFERRED · 0% AMBIGUOUS · INFERRED: 231 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Backend API|Backend API]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Backend API|Backend API]]
- [[_COMMUNITY_Docs And Plans|Docs And Plans]]
- [[_COMMUNITY_Backend Runner|Backend Runner]]
- [[_COMMUNITY_Extension Storage|Extension Storage]]
- [[_COMMUNITY_Extension Popup|Extension Popup]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Docs And Plans|Docs And Plans]]
- [[_COMMUNITY_Backend API|Backend API]]
- [[_COMMUNITY_Docker Deployment|Docker Deployment]]
- [[_COMMUNITY_Backend Runner|Backend Runner]]
- [[_COMMUNITY_Extension Popup|Extension Popup]]
- [[_COMMUNITY_Backend Runner|Backend Runner]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Extension UI Components|Extension UI Components]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Extension Storage|Extension Storage]]
- [[_COMMUNITY_YouTube Overlay|YouTube Overlay]]
- [[_COMMUNITY_Backend Store|Backend Store]]
- [[_COMMUNITY_Backend API|Backend API]]
- [[_COMMUNITY_Backend Store|Backend Store]]
- [[_COMMUNITY_Extension UI Components|Extension UI Components]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_YouTube Overlay|YouTube Overlay]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Extension Storage|Extension Storage]]
- [[_COMMUNITY_Testing Contracts|Testing Contracts]]
- [[_COMMUNITY_Agent Operating Rules|Agent Operating Rules]]
- [[_COMMUNITY_vtt ts parseTimestamp|vtt ts parseTimestamp]]
- [[_COMMUNITY_Backend Runner|Backend Runner]]
- [[_COMMUNITY_Docs And Plans|Docs And Plans]]
- [[_COMMUNITY_Extension Popup|Extension Popup]]
- [[_COMMUNITY_Backend Runner|Backend Runner]]
- [[_COMMUNITY_Backend Store|Backend Store]]
- [[_COMMUNITY_Docs And Plans|Docs And Plans]]
- [[_COMMUNITY_Extension UI Components|Extension UI Components]]
- [[_COMMUNITY_Extension UI Components|Extension UI Components]]
- [[_COMMUNITY_Extension UI Components|Extension UI Components]]
- [[_COMMUNITY_Extension Storage|Extension Storage]]

## God Nodes (most connected - your core abstractions)
1. `main()` - 18 edges
2. `LoadConfig()` - 17 edges
3. `NewJob()` - 17 edges
4. `openTestStore()` - 16 edges
5. `handleExtensionMessage()` - 12 edges
6. `newTestServer()` - 12 edges
7. `Segment` - 11 edges
8. `transcribe_audio()` - 11 edges
9. `Store` - 10 edges
10. `render_vtt()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `Backend Request Error Mapping` --implements--> `Backend Mock MVP Design`  [INFERRED]
  extension/src/api/backend-client.ts → docs/superpowers/specs/2026-04-24-backend-mock-mvp-design.md
- `Claude Agent Operating Rules` --semantically_similar_to--> `Agent Operating Rules`  [INFERRED] [semantically similar]
  CLAUDE.md → AGENTS.md
- `runner.RealRunner.Start` --implements--> `Real Runner Download Transcribe Translate Pipeline`  [INFERRED]
  backend/internal/runner/real_runner.go → README.md
- `api.Handler.handleJobs` --implements--> `Job Reuse by videoId and targetLanguage`  [INFERRED]
  backend/internal/api/handler.go → README.md
- `api.Handler.handleActiveJob` --conceptually_related_to--> `Job Reuse by videoId and targetLanguage`  [INFERRED]
  backend/internal/api/handler.go → README.md

## Hyperedges (group relationships)
- **End-to-end Subtitle Pipeline** — concept_chrome_extension_module, concept_backend_module, concept_whisper_cli_module, concept_real_runner_pipeline, runner_real_runner_start [EXTRACTED 1.00]
- **Backend HTTP API Surface** — api_routes, api_handle_jobs, api_handle_active_job, api_handle_subtitle_assets, api_handle_subtitle_file [EXTRACTED 1.00]
- **Local Security Boundaries** — concept_localhost_only_backend_origin, api_is_allowed_local_origin, concept_safe_subtitle_file_serving, api_subtitle_file_path_allowed, concept_single_user_security_boundary [INFERRED 0.88]
- **Real Runner Subtitle Generation Pipeline** — downloader_download_audio, transcriber_transcribe_audio, translator_chattranslator, vtt_cue_parse_webvtt_cues, vtt_cue_render_translated_vtt, vtt_cue_render_bilingual_vtt, sqlite_create_subtitle_asset [INFERRED 0.87]
- **Job And Asset Persistence Contract** — store_job, store_subtitle_asset, sqlite_store, sqlite_find_reusable_job, sqlite_find_latest_job, sqlite_update_job_status, sqlite_find_subtitle_asset [EXTRACTED 0.95]
- **MVP End To End Contract** — PRD_self_hosted_subtitle_pipeline, PRD_extension_backend_contract, PRD_security_boundaries, 2026_04_25_extension_mvp_plan, 2026_04_29_openai_compatible_translation_plan, 2026_04_23_whisper_sdk_cli_plan, 2026_05_02_docker_deployment_plan [EXTRACTED 0.88]
- **Extension Job Lifecycle** — 2026_04_25_extension_mvp_design_extension_mvp_design, messages_extension_message_protocol, App_submit_job, App_poll_job, backend_client_create_backend_client, job_monitor_start_job_monitor [EXTRACTED 0.89]
- **Real Runner Pipeline** — 2026_04_27_real_audio_download_design_real_audio_download, 2026_04_23_whisper_cli_design_whisper_cli_contract, 2026_04_29_openai_compatible_translation_design_chat_completions_translation, 2026_05_02_docker_deployment_design_backend_docker_deployment, 2026_04_24_backend_mock_mvp_design_backend_mock_mvp_design [INFERRED 0.84]
- **Popup And Background Recovery Monitoring** — README_extension_runtime_docs, App_restore_job_for_current_tab, background_background_service_worker, job_monitor_persisted_job_monitoring, job_monitor_test_persisted_monitor_tests, App_test_popup_active_job_restore_test [EXTRACTED 0.87]
- **Extension Message Dispatch Flows** — message_handler_handleExtensionMessage, message_handler_JobCreateFlow, message_handler_SubtitleResolveFlow, message_handler_clientFromSettings, message_handler_updateSubtitleMode, message_handler_errorToMessage [EXTRACTED 0.92]
- **Handler Behavior Test Coverage** — message_handler_test_HandleExtensionMessageSuite, message_handler_handleExtensionMessage, message_handler_test_InvalidLanguagePairBehavior, message_handler_test_JobMonitorBehavior, message_handler_test_BackendOriginCacheIsolation [EXTRACTED 0.90]
- **shadcn-vue UI Primitive Set** — alert_Alert, badge_Badge, button_Button, card_Card, input_Input, select_SelectItem, select_SelectTrigger, ui_SharedClassPassthroughPattern, ui_RekaPrimitiveWrapperPattern [INFERRED 0.86]
- **Select Wrappers Form Reka/Shadcn Select API** — Select_select_root_wrapper, SelectContent_select_content_wrapper, SelectGroup_select_group_wrapper, SelectSeparator_select_separator_wrapper, SelectValue_select_value_wrapper, SelectLabel_select_label_wrapper, SelectItemText_select_item_text_wrapper, index_select_component_barrel [EXTRACTED 0.93]
- **Subtitle Cache Scopes Assets By Backend, Video, And Language** — subtitle_cache_subtitleAssetKey, subtitle_cache_videoPreferenceKey, subtitle_cache_setCachedSubtitleAsset, subtitle_cache_getCachedSubtitleAsset, subtitle_cache_updateCachedSubtitleMode, subtitle_cache_test_storage_isolation_contract [EXTRACTED 0.92]
- **YouTube Overlay Mode Switch Rollback Workflow** — YoutubeOverlay_overlay_component, YoutubeOverlay_changeMode, YoutubeOverlay_loadVtt, YoutubeOverlay_restoreDisplayedSubtitles, YoutubeOverlay_test_mode_rollback_contract [EXTRACTED 0.91]
- **Extension Subtitle Cue Flow** — extension_subtitles_vtt_parsevtt, extension_subtitles_vtt_parsetimestamp, extension_subtitles_vtt_vttcue, extension_subtitles_activecue_findactivecue, extension_subtitles_vtt_tests [INFERRED 0.86]
- **Extension YouTube Watch Detection Flow** — extension_youtube_videoid_parseyoutubewatchvideoid, extension_youtube_pagewatch_getvideoidfromlocationhref, extension_youtube_pagewatch_getcurrentvideoid, extension_youtube_pagewatch_watchvideoidchanges, extension_youtube_videoid_tests, extension_youtube_pagewatch_tests [INFERRED 0.88]
- **Whisper CLI Transcription To VTT Pipeline** — whisper_cli_build_parser, whisper_cli_main, whisper_transcribe_transcribe_audio, whisper_transcribe_transcriptionresult, whisper_vtt_segment, whisper_vtt_render_vtt, whisper_readme_cli_contract, whisper_cli_tests [INFERRED 0.90]

## Communities

### Community 0 - "Backend API"
Cohesion: 0.07
Nodes (32): NewHandler(), newTestServer(), TestActiveJobReturnsLatestJobForVideoAndLanguage(), TestActiveJobReturnsNullWhenNoJobExists(), TestPostJobsCreatesJobAndCompletesWithMockRunner(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion(), TestSubtitleFileRejectsInvalidMode() (+24 more)

### Community 1 - "Testing Contracts"
Cohesion: 0.1
Nodes (31): execCall, fakeTranslator, NewMockRunner(), openTestStore(), TestMockRunnerCompletesJobAndWritesAssets(), TestMockRunnerFailsJobWhenCompletionUpdateFails(), TestMockRunnerMarksCanceledJobAsFailed(), TestMockRunnerMarksFailedWhenTranscribingUpdateFails() (+23 more)

### Community 2 - "Backend API"
Cohesion: 0.07
Nodes (40): Extension Contract, api.Handler.handleActiveJob, api.Handler.handleJobs, api.Handler.handleSubtitleAssets, api.Handler.handleSubtitleFile, api.isAllowedLocalOrigin, api.ParseVideoID, api.Routes (+32 more)

### Community 3 - "Docs And Plans"
Cohesion: 0.07
Nodes (38): Backend Contract, Offline Repeatable Test Rule, Backend API Server, Backend README Runner Boundary, CLAUDE Agent Instruction Mirror, Mock Runner Default, Docker Compose Backend Service, lsi-data Docker Volume (+30 more)

### Community 4 - "Backend Runner"
Cohesion: 0.1
Nodes (23): Cue, downloadAudio(), TestDownloadAudioCreatesJobDir(), TestDownloadAudioNetworkError(), TestDownloadAudioSuccess(), TestDownloadAudioTimeout(), TestDownloadAudioVideoUnavailable(), TestDownloadAudioYtDlpMissing() (+15 more)

### Community 5 - "Extension Storage"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 6 - "Extension Popup"
Cohesion: 0.08
Nodes (31): Backend Mock MVP Design, Backend Mock MVP Plan, Extension MVP Design, Real Audio Download Design, pollJob, Popup App, restoreJobForCurrentTab, submitJob (+23 more)

### Community 7 - "Testing Contracts"
Cohesion: 0.13
Nodes (22): Exception, FakeModel, FakeSegment, test_transcribe_audio_allows_empty_segment_output(), test_transcribe_audio_passes_compute_type_to_sdk(), test_transcribe_audio_rejects_english_only_model_with_non_english_language(), test_transcribe_audio_uses_sdk_and_builds_result(), test_transcribe_audio_uses_sdk_reported_duration() (+14 more)

### Community 8 - "Docs And Plans"
Cohesion: 0.1
Nodes (27): Backend Mock MVP 设计, Backend HTTP API 契约, Job 数据模型, SubtitleAsset 数据模型, Backend Mock MVP 实施计划, Go Mock Backend, MockRunner 状态机, SQLite/GORM Store (+19 more)

### Community 9 - "Backend API"
Cohesion: 0.14
Nodes (18): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+10 more)

### Community 10 - "Docker Deployment"
Cohesion: 0.12
Nodes (26): faster-whisper SDK, WebVTT 输出校验, whisper-cli, Docker 默认本机端口绑定, lsi-data Named Volume, 多阶段构建单镜像, backend/Dockerfile, docker-compose.yml (+18 more)

### Community 11 - "Backend Runner"
Cohesion: 0.1
Nodes (25): Whisper SDK CLI Plan, Real Audio Download Plan, OpenAI-Compatible Translation Plan, Docs Chinese Language Policy, Security and Privacy Boundaries, Download Audio, Fake Translator, RealRunner Pipeline Tests (+17 more)

### Community 12 - "Extension Popup"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 13 - "Backend Runner"
Cohesion: 0.15
Nodes (17): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, translationCue, translationPrompt, translationResponse (+9 more)

### Community 14 - "Testing Contracts"
Cohesion: 0.16
Nodes (18): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+10 more)

### Community 15 - "Extension UI Components"
Cohesion: 0.13
Nodes (20): Alert Component, AlertDescription Component, AlertTitle Component, alertVariants, Badge Component, badgeVariants, Button Component, buttonVariants (+12 more)

### Community 16 - "Testing Contracts"
Cohesion: 0.2
Nodes (17): fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_passes_compute_type_to_transcriber(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails(), test_cli_returns_code_2_when_input_file_is_not_readable() (+9 more)

### Community 17 - "Extension Storage"
Cohesion: 0.16
Nodes (19): Bind Overlay To Video Time, Change Subtitle Mode, Subtitle Updated Message Guard, Load Subtitles For Video, Load Selected VTT File, YouTube Subtitle Overlay Component, Reset Loaded Subtitles, Restore Displayed Subtitles (+11 more)

### Community 18 - "YouTube Overlay"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 19 - "Backend Store"
Cohesion: 0.14
Nodes (17): Docker Deployment Plan, Job State Flow, Self-Hosted YouTube Subtitle Pipeline, Store AutoMigrate, MockRunner Failure Handling, MockRunner Start, Create Subtitle Asset, Find Latest Job (+9 more)

### Community 20 - "Backend API"
Cohesion: 0.23
Nodes (13): Job Create Flow, MessageHandlerDeps, Subtitle Resolve Flow, clientFromSettings, errorToMessage, handleExtensionMessage, ok Result Wrapper, Backend Origin Cache Isolation (+5 more)

### Community 21 - "Backend Store"
Cohesion: 0.2
Nodes (1): Store

### Community 22 - "Extension UI Components"
Cohesion: 0.24
Nodes (10): Select Content Wrapper, Select Label Wrapper, Select Scroll Down Button Wrapper, Select Scroll Up Button Wrapper, Select Separator Wrapper, Select Root Wrapper, Separator Wrapper, Select Component Barrel (+2 more)

### Community 23 - "Testing Contracts"
Cohesion: 0.39
Nodes (6): ParseVideoID(), TestParseVideoIDRejectsNonWatchYouTubeURL(), TestParseVideoIDRejectsUnsupportedScheme(), TestParseVideoIDRejectsUnsupportedURL(), TestParseVideoIDSupportsShortURL(), TestParseVideoIDSupportsWatchURL()

### Community 24 - "YouTube Overlay"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 25 - "Testing Contracts"
Cohesion: 0.33
Nodes (6): getCurrentVideoId, getVideoIdFromLocationHref, YouTube Page Watch Tests, watchVideoIdChanges, parseYouTubeWatchVideoId, YouTube Video Id Tests

### Community 26 - "Extension Storage"
Cohesion: 0.5
Nodes (5): Default Extension Settings, Create Language Pair, Read Extension Settings, Settings Storage Tests, Update Extension Settings

### Community 27 - "Testing Contracts"
Cohesion: 0.4
Nodes (5): findActiveCue, parseTimestamp, parseVtt, Extension VTT Parser Tests, Extension VTT Cue

### Community 29 - "Agent Operating Rules"
Cohesion: 0.5
Nodes (4): Agent Operating Rules, Claude Agent Operating Rules, Docs Chinese Language Policy, Docs Claude Chinese Language Policy

### Community 30 - "vtt ts parseTimestamp"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 31 - "Backend Runner"
Cohesion: 0.67
Nodes (2): Runner, Store

### Community 32 - "Docs And Plans"
Cohesion: 0.67
Nodes (3): Whisper CLI Contract, Chat Completions Translation Design, Backend Docker Deployment Design

### Community 33 - "Extension Popup"
Cohesion: 0.67
Nodes (3): Create Job Form Shape, Create Job Form Validation Tests, Create Job Form Validator

### Community 38 - "Backend Runner"
Cohesion: 1.0
Nodes (2): MockRunner, MockRunner Failure Path Tests

### Community 39 - "Backend Store"
Cohesion: 1.0
Nodes (2): SQLite Foreign Key DSN, SQLite Store

### Community 40 - "Docs And Plans"
Cohesion: 1.0
Nodes (2): Extension MVP Plan, Extension Backend Contract

### Community 89 - "Extension UI Components"
Cohesion: 1.0
Nodes (1): Select Group Wrapper

### Community 90 - "Extension UI Components"
Cohesion: 1.0
Nodes (1): Select Value Wrapper

### Community 91 - "Extension UI Components"
Cohesion: 1.0
Nodes (1): Select Item Text Wrapper

### Community 92 - "Extension Storage"
Cohesion: 1.0
Nodes (1): Video Preference Shape

## Knowledge Gaps
- **120 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+115 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Backend Store`** (10 nodes): `Store`, `.CreateJob()`, `.CreateSubtitleAsset()`, `.FindJob()`, `.FindLatestJob()`, `.FindReusableJob()`, `.FindSubtitleAsset()`, `.FindSubtitleAssetByJobID()`, `.Migrate()`, `.UpdateJobStatus()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `vtt ts parseTimestamp`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Backend Runner`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Backend Runner`** (2 nodes): `MockRunner`, `MockRunner Failure Path Tests`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Backend Store`** (2 nodes): `SQLite Foreign Key DSN`, `SQLite Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Docs And Plans`** (2 nodes): `Extension MVP Plan`, `Extension Backend Contract`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension UI Components`** (1 nodes): `Select Group Wrapper`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension UI Components`** (1 nodes): `Select Value Wrapper`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension UI Components`** (1 nodes): `Select Item Text Wrapper`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension Storage`** (1 nodes): `Video Preference Shape`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewHTTPHandler()` connect `Backend API` to `Testing Contracts`, `Backend Runner`, `Testing Contracts`?**
  _High betweenness centrality (0.053) - this node is a cross-community bridge._
- **Why does `NewJob()` connect `Testing Contracts` to `Backend API`, `Backend Store`?**
  _High betweenness centrality (0.053) - this node is a cross-community bridge._
- **Are the 15 inferred relationships involving `main()` (e.g. with `test_cli_requires_all_required_arguments()` and `test_cli_creates_parent_directory_for_output()`) actually correct?**
  _`main()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 14 inferred relationships involving `LoadConfig()` (e.g. with `main()` and `TestLoadConfigUsesDefaults()`) actually correct?**
  _`LoadConfig()` has 14 INFERRED edges - model-reasoned connections that need verification._
- **Are the 16 inferred relationships involving `NewJob()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`NewJob()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `openTestStore()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`openTestStore()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `handleExtensionMessage()` (e.g. with `getSettings()` and `updateSettings()`) actually correct?**
  _`handleExtensionMessage()` has 7 INFERRED edges - model-reasoned connections that need verification._