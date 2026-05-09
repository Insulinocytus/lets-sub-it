# Graph Report - .  (2026-05-10)

## Corpus Check
- 148 files · ~60,830 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 758 nodes · 1147 edges · 28 communities detected
- Extraction: 78% EXTRACTED · 22% INFERRED · 0% AMBIGUOUS · INFERRED: 254 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Backend Runtime Wiring|Backend Runtime Wiring]]
- [[_COMMUNITY_System Architecture Docs|System Architecture Docs]]
- [[_COMMUNITY_Whisper CLI Tests|Whisper CLI Tests]]
- [[_COMMUNITY_Extension Backend Client|Extension Backend Client]]
- [[_COMMUNITY_Subtitle Utility Contracts|Subtitle Utility Contracts]]
- [[_COMMUNITY_API Handler Tests|API Handler Tests]]
- [[_COMMUNITY_Runner VTT Files|Runner VTT Files]]
- [[_COMMUNITY_Reference Design Docs|Reference Design Docs]]
- [[_COMMUNITY_API Handler Types|API Handler Types]]
- [[_COMMUNITY_Whisper CLI Runtime|Whisper CLI Runtime]]
- [[_COMMUNITY_Backend Client Errors|Backend Client Errors]]
- [[_COMMUNITY_UI Component Variants|UI Component Variants]]
- [[_COMMUNITY_Translator Client|Translator Client]]
- [[_COMMUNITY_Runner Test Doubles|Runner Test Doubles]]
- [[_COMMUNITY_Backend Config|Backend Config]]
- [[_COMMUNITY_Job Monitor Runtime|Job Monitor Runtime]]
- [[_COMMUNITY_YouTube Overlay Flow|YouTube Overlay Flow]]
- [[_COMMUNITY_SQLite Store|SQLite Store]]
- [[_COMMUNITY_Extension Message Types|Extension Message Types]]
- [[_COMMUNITY_Reka UI Wrappers|Reka UI Wrappers]]
- [[_COMMUNITY_YouTube Video Detection|YouTube Video Detection]]
- [[_COMMUNITY_YouTube Watch Utilities|YouTube Watch Utilities]]
- [[_COMMUNITY_Extension VTT Parser|Extension VTT Parser]]
- [[_COMMUNITY_Runner Interfaces|Runner Interfaces]]
- [[_COMMUNITY_Agent Graphify Rules|Agent Graphify Rules]]
- [[_COMMUNITY_Local Storage Model|Local Storage Model]]
- [[_COMMUNITY_Extension Test Config|Extension Test Config]]
- [[_COMMUNITY_Input Component|Input Component]]

## God Nodes (most connected - your core abstractions)
1. `main()` - 18 edges
2. `LoadConfig()` - 17 edges
3. `openTestStore()` - 17 edges
4. `RealRunner.Start` - 16 edges
5. `NewJob()` - 14 edges
6. `handleExtensionMessage()` - 12 edges
7. `POST /jobs Handler` - 12 edges
8. `newTestServer()` - 11 edges
9. `Segment` - 11 edges
10. `transcribe_audio()` - 11 edges

## Surprising Connections (you probably didn't know these)
- `Docker Localhost Port Binding` --semantically_similar_to--> `Localhost Origin Policy`  [INFERRED] [semantically similar]
  docker-compose.yml → backend/internal/api/routes.go
- `No Sensitive LLM Request Logs` --semantically_similar_to--> `ChatTranslator.redactAPIKey`  [INFERRED] [semantically similar]
  docs/explanation/security-and-privacy.md → backend/internal/runner/translator.go
- `Job Reuse Key` --semantically_similar_to--> `Store.FindReusableJob`  [INFERRED] [semantically similar]
  docs/explanation/processing-pipeline.md → backend/internal/store/sqlite.go
- `Offline Test Boundary Rule` --semantically_similar_to--> `RealRunner Offline Pipeline Tests`  [INFERRED] [semantically similar]
  docs/explanation/module-boundaries.md → backend/internal/runner/real_runner_test.go
- `WXT Manifest 与 localhost host permissions` --semantically_similar_to--> `本机端口绑定和持久化运行数据`  [INFERRED] [semantically similar]
  extension/wxt.config.ts → docs/superpowers/specs/2026-05-02-docker-deployment-design.md

## Hyperedges (group relationships)
- **End-to-end Subtitle Generation Flow** — readme_end_to_end_pipeline, handler_handle_jobs, downloader_download_audio, backend_readme_runner_boundary, response_to_asset_response [INFERRED 0.85]
- **Localhost Backend Access Boundary** — agents_security_boundaries, readme_limitations, docker_compose_localhost_binding, routes_localhost_origin_policy [INFERRED 0.80]
- **Offline Backend Test Pattern** — agents_offline_test_contract, handler_test_fake_completion_runner, handler_test_job_lifecycle_coverage, app_test_tool_prerequisite_coverage, config_test_environment_coverage [INFERRED 0.85]
- **RealRunner Pipeline Flow** — real_runner_realrunner_start, downloader_downloadaudio, transcriber_transcribeaudio, vtt_cue_parsewebvttcues, translator_chattranslator_translate, vtt_cue_rendertranslatedvtt, vtt_cue_renderbilingualvtt, sqlite_createsubtitleasset [EXTRACTED 1.00]
- **SQLite Job and Subtitle Asset Persistence Contract** — models_job, models_subtitleasset, migrations_store_migrate, sqlite_findreusablejob, sqlite_updatejobstatus, sqlite_createsubtitleasset, sqlite_findsubtitleasset [EXTRACTED 1.00]
- **Documented Extension to Backend Subtitle Flow** — architecture_background_gateway, architecture_go_api_embedded_runner, architecture_whisper_cli_boundary, architecture_local_vtt_files, security_content_script_boundary [EXTRACTED 1.00]
- **Subtitle Generation Flow** — backend_api_contract, job_state_flow, real_runner, whisper_cli_contract, chat_completions_translator, subtitle_asset_contract, extension_background_gateway [INFERRED 0.86]
- **Diataxis Documentation System** — diataxis_docs_plan, diataxis_human_docs_tree, how_to_readme_operational_guides, reference_readme_contracts_index [EXTRACTED 1.00]
- **Real Runner Processing Pipeline** — real_audio_yt_dlp_download, translation_chat_translator, translation_vtt_packaging, onboarding_pipeline [EXTRACTED 1.00]
- **Extension Message Gateway Flow** — popup_app, background_gateway, message_handler, backend_client, job_monitor, youtube_content_overlay_mount [INFERRED 0.86]
- **Diátaxis Documentation Onboarding Path** — diataxis_human_docs_quadrants, tutorials_index, onboarding_tutorial, diataxis_superpowers_archive [EXTRACTED 1.00]
- **Extension Message Protocol Flow** — messages_extension_message, messages_settings, messages_create_job_input, messages_job, messages_subtitle_asset, messages_send_extension_message, messages_message_result [EXTRACTED 1.00]
- **CVA Variant Class Contract** — ui_variant_class_pattern, alert_alert_variants, alert_alert_component, badge_badge_variants, badge_badge_component, button_button_variants, button_button_component [EXTRACTED 1.00]
- **Data Slot Layout Composition** — alert_data_slot_contract, alert_alert_component, alert_alert_description_component, alert_alert_title_component, card_slot_layout_contract, card_card_component, card_card_header_component, card_card_action_component, card_card_content_component, card_card_description_component, card_card_footer_component, card_card_title_component [INFERRED 0.84]
- **Reka Select Wrapper Pattern** — select_RekaSelectWrapperSet, select_SelectItemWrapper, select_SelectTriggerWrapper, select_SelectScrollButtonWrappers, select_SelectTextAndSeparatorWrappers, utils_cn, ui_RekaPropDelegation [EXTRACTED 1.00]
- **Backend Origin Scoped Storage Pattern** — settings_updateSettings, subtitle_subtitleAssetKey, subtitle_videoPreferenceKey, external_normalizeBackendBaseUrl, settings_storageTests, subtitle_cacheTests [INFERRED 0.85]
- **Subtitle Mode Rollback Pattern** — overlay_changeMode, overlay_loadVtt, overlay_restoreDisplayedSubtitles, overlay_ModeRollbackFlow, overlay_modeRollbackTests, external_sendExtensionMessage [EXTRACTED 1.00]
- **YouTube Watch Video ID Flow** — video_id_parseYouTubeWatchVideoId, page_watch_getVideoIdFromLocationHref, page_watch_getCurrentVideoId, page_watch_watchVideoIdChanges [EXTRACTED 1.00]
- **whisper-cli Transcription To VTT Flow** — cli_main, transcribe_transcribeAudio, transcribe_TranscriptionResult, vtt_py_renderVtt, cli_JsonSummaryStdout [EXTRACTED 1.00]
- **WebVTT Production And Consumption Contract** — vtt_py_Segment, vtt_py_renderVtt, vtt_py_WebVttRenderingContract, ext_vtt_parseVtt, ext_vtt_VttCue [INFERRED 0.85]

## Communities

### Community 0 - "Backend Runtime Wiring"
Cohesion: 0.05
Nodes (64): Backend Runtime Pipeline, Multi-module MVP Architecture, Offline Unit Test Contract, Security and Data Boundaries, NewHTTPHandler Real Backend Wiring, Tool Prerequisite Test Coverage, External Tool Prerequisite Check, Backend README API Usage (+56 more)

### Community 1 - "System Architecture Docs"
Cohesion: 0.05
Nodes (59): Background Service Worker Gateway, Go API Server with Embedded Runner, Local Subtitle VTT Files, Three Module Architecture, Whisper CLI Boundary, Chinese Documentation Policy, Diataxis Documentation Structure, downloadAudio (+51 more)

### Community 2 - "Whisper CLI Tests"
Cohesion: 0.08
Nodes (39): Exception, fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_passes_compute_type_to_transcriber(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails() (+31 more)

### Community 3 - "Extension Backend Client"
Cohesion: 0.07
Nodes (44): Extension BackendClient, BackendClient 网络和 JSON 错误映射, BackendClient 单元测试, Job 与 Subtitle 后端 API 方法, backend URL localhost origin 校验, Background runtime message 和 alarm 网关, Diátaxis 文档体系设计, tutorials how-to reference explanation 四象限 (+36 more)

### Community 4 - "Subtitle Utility Contracts"
Cohesion: 0.08
Nodes (43): Half-Open Cue Interval, VttCue, findActiveCue Tests, findActiveCue, assertDifferentLanguages, Browser Runtime Message Listener, normalizeBackendBaseUrl, YouTube Page Watch Utilities (+35 more)

### Community 5 - "API Handler Tests"
Cohesion: 0.08
Nodes (29): completingRunner, NewHandler(), newTestServer(), TestActiveJobReturnsLatestJobForVideoAndLanguage(), TestActiveJobReturnsNullWhenNoJobExists(), TestPostJobsCreatesJobAndCompletes(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion() (+21 more)

### Community 6 - "Runner VTT Files"
Cohesion: 0.09
Nodes (27): Cue, downloadAudio(), TestDownloadAudioCreatesJobDir(), TestDownloadAudioNetworkError(), TestDownloadAudioSuccess(), TestDownloadAudioTimeout(), TestDownloadAudioVideoUnavailable(), TestDownloadAudioYtDlpMissing() (+19 more)

### Community 7 - "Reference Design Docs"
Cohesion: 0.11
Nodes (35): Backend HTTP API Contract, Backend Configuration Contract, Backend Mock MVP Design, Backend Mock MVP Implementation Plan, Backend Runner Interface, Chat Completions Translator, Data and Job Model Contract, Diataxis Documentation Plan (+27 more)

### Community 8 - "API Handler Types"
Cohesion: 0.1
Nodes (24): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+16 more)

### Community 9 - "Whisper CLI Runtime"
Cohesion: 0.1
Nodes (32): whisper-cli Arguments, whisper-cli Exit Code Mapping, whisper-cli JSON Summary Stdout, build_parser, whisper-cli main, main_entry, CLI Behavior Test Coverage, Fake Transcription Result (+24 more)

### Community 10 - "Backend Client Errors"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 11 - "UI Component Variants"
Cohesion: 0.08
Nodes (31): Alert Component, AlertDescription Component, AlertTitle Component, Alert Variant Classes, ARIA Alert Role, Alert Data Slot Contract, Badge Component, Badge Variant Classes (+23 more)

### Community 12 - "Translator Client"
Cohesion: 0.12
Nodes (24): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, translationCue, translationPrompt, translationResponse (+16 more)

### Community 13 - "Runner Test Doubles"
Cohesion: 0.16
Nodes (25): execCall, fakeTranslator, NewRealRunner(), argValue(), assertArg(), findExecCall(), TestRealRunnerCompletesJob(), TestRealRunnerDownloadFailed() (+17 more)

### Community 14 - "Backend Config"
Cohesion: 0.12
Nodes (22): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+14 more)

### Community 15 - "Job Monitor Runtime"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 16 - "YouTube Overlay Flow"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 17 - "SQLite Store"
Cohesion: 0.13
Nodes (6): openTestStore(), foreignKeyDSN(), Open(), openWithLogWriter(), TestStoreDoesNotLogExpectedRecordNotFound(), Store

### Community 18 - "Extension Message Types"
Cohesion: 0.13
Nodes (17): assertDifferentLanguages, browser.runtime.sendMessage, CreateJobInput DTO, Different Language Constraint, ExtensionMessage Protocol, isSupportedLanguage Guard, Job API DTO, JobStatus State Machine (+9 more)

### Community 19 - "Reka UI Wrappers"
Cohesion: 0.43
Nodes (8): Reka Select Wrapper Set, SelectItem Wrapper, Select Scroll Button Wrappers, Select Text Label Value Separator Wrappers, SelectTrigger Wrapper, Separator Wrapper, Reka Prop Delegation Pattern, cn Class Merge Helper

### Community 20 - "YouTube Video Detection"
Cohesion: 0.36
Nodes (8): YouTube Navigation Listeners, getCurrentVideoId, getVideoIdFromLocationHref, Page Watch Test Coverage, watchVideoIdChanges, YouTube Watch URL Contract, parseYouTubeWatchVideoId, Video ID Parsing Test Coverage

### Community 21 - "YouTube Watch Utilities"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 23 - "Extension VTT Parser"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 24 - "Runner Interfaces"
Cohesion: 0.67
Nodes (2): Runner, Store

### Community 29 - "Agent Graphify Rules"
Cohesion: 1.0
Nodes (2): Graphify Knowledge Graph Workflow, Agent Operating Guidelines

### Community 30 - "Local Storage Model"
Cohesion: 1.0
Nodes (2): Docker Persistent Runtime Volumes, Local-first Storage Model

### Community 31 - "Extension Test Config"
Cohesion: 1.0
Nodes (2): Vitest jsdom WXT 测试配置, 扩展测试脚手架冒烟测试

### Community 32 - "Input Component"
Cohesion: 1.0
Nodes (2): Input Component, Input Model Value Bridge

## Ambiguous Edges - Review These
- `Extension Runtime Contract` → `Extension Language Code Ambiguity`  [AMBIGUOUS]
  docs/reference/extension-contract.md · relation: conceptually_related_to
- `Extension MVP Design` → `Extension Language Code Ambiguity`  [AMBIGUOUS]
  docs/superpowers/specs/2026-04-25-extension-mvp-design.md · relation: conceptually_related_to
- `getCachedSubtitleAsset` → `loadForVideo`  [AMBIGUOUS]
  extension/src/content/YoutubeOverlay.vue · relation: conceptually_related_to
- `updateCachedSubtitleMode` → `changeMode`  [AMBIGUOUS]
  extension/src/content/YoutubeOverlay.vue · relation: conceptually_related_to

## Knowledge Gaps
- **99 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+94 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Extension VTT Parser`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Runner Interfaces`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Agent Graphify Rules`** (2 nodes): `Graphify Knowledge Graph Workflow`, `Agent Operating Guidelines`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Local Storage Model`** (2 nodes): `Docker Persistent Runtime Volumes`, `Local-first Storage Model`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension Test Config`** (2 nodes): `Vitest jsdom WXT 测试配置`, `扩展测试脚手架冒烟测试`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Input Component`** (2 nodes): `Input Component`, `Input Model Value Bridge`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Extension Runtime Contract` and `Extension Language Code Ambiguity`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Extension MVP Design` and `Extension Language Code Ambiguity`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `getCachedSubtitleAsset` and `loadForVideo`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `updateCachedSubtitleMode` and `changeMode`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `NewHTTPHandler()` connect `API Handler Tests` to `SQLite Store`, `Translator Client`, `Runner Test Doubles`, `Backend Config`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `NewRealRunner()` connect `Runner Test Doubles` to `API Handler Tests`, `Runner VTT Files`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `NewChatTranslator()` connect `Translator Client` to `API Handler Tests`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._