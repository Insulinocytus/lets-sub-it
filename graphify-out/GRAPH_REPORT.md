# Graph Report - .  (2026-05-02)

## Corpus Check
- 123 files · ~55,354 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 562 nodes · 778 edges · 49 communities detected
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 218 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_.Start downloadAudio|.Start downloadAudio]]
- [[_COMMUNITY_NewJob openTestStore|NewJob openTestStore]]
- [[_COMMUNITY_handler test.go|handler test.go]]
- [[_COMMUNITY_handleExtensionMessage normalizeBackendBaseUrl|handleExtensionMessage normalizeBackendBaseUrl]]
- [[_COMMUNITY_Segment transcribe audio|Segment transcribe audio]]
- [[_COMMUNITY_response.go writeError|response.go writeError]]
- [[_COMMUNITY_YouTube subtitle translation PRD|YouTube subtitle translation PRD]]
- [[_COMMUNITY_job-monitor.ts pollJob|job-monitor.ts pollJob]]
- [[_COMMUNITY_cn shadcn-vue component pattern|cn shadcn-vue component pattern]]
- [[_COMMUNITY_main test cli.py|main test cli.py]]
- [[_COMMUNITY_LoadConfig config test.go|LoadConfig config test.go]]
- [[_COMMUNITY_handleExtensionMessage BackendClient Interface|handleExtensionMessage BackendClient Interface]]
- [[_COMMUNITY_YoutubeOverlay.vue changeMode|YoutubeOverlay.vue changeMode]]
- [[_COMMUNITY_Popup App Component|Popup App Component]]
- [[_COMMUNITY_translator.go .translateOne|translator.go .translateOne]]
- [[_COMMUNITY_Backend Mock MVP Design Spec|Backend Mock MVP Design Spec]]
- [[_COMMUNITY_NewHTTPHandler NewChatTranslator|NewHTTPHandler NewChatTranslator]]
- [[_COMMUNITY_Store Open|Store Open]]
- [[_COMMUNITY_ParseVideoID youtube test.go|ParseVideoID youtube test.go]]
- [[_COMMUNITY_toJobResponse formatTime|toJobResponse formatTime]]
- [[_COMMUNITY_page-watch.ts getCurrentVideoId|page-watch.ts getCurrentVideoId]]
- [[_COMMUNITY_Runner Interface|Runner Interface]]
- [[_COMMUNITY_BackendClient MessageHandler|BackendClient MessageHandler]]
- [[_COMMUNITY_vtt.ts parseTimestamp|vtt.ts parseTimestamp]]
- [[_COMMUNITY_runner.go Runner|runner.go Runner]]
- [[_COMMUNITY_cn Badge Component Barrel|cn Badge Component Barrel]]
- [[_COMMUNITY_runner.Store Interface|runner.Store Interface]]
- [[_COMMUNITY_whisper-cli Command|whisper-cli Command]]
- [[_COMMUNITY_VTT Generator|VTT Generator]]
- [[_COMMUNITY_Extension Manifest Configuration|Extension Manifest Configuration]]
- [[_COMMUNITY_handleSubtitleFile subtitleFilePathAllowed|handleSubtitleFile subtitleFilePathAllowed]]
- [[_COMMUNITY_writeError writeJSON|writeError writeJSON]]
- [[_COMMUNITY_Chat Completions Translator|Chat Completions Translator]]
- [[_COMMUNITY_Vitest Config|Vitest Config]]
- [[_COMMUNITY_Background Service Worker|Background Service Worker]]
- [[_COMMUNITY_CardAction UI Component|CardAction UI Component]]
- [[_COMMUNITY_Card Barrel Export|Card Barrel Export]]
- [[_COMMUNITY_Input Barrel Export|Input Barrel Export]]
- [[_COMMUNITY_Alert Barrel Export|Alert Barrel Export]]
- [[_COMMUNITY_Separator Barrel Export|Separator Barrel Export]]
- [[_COMMUNITY_Button Barrel Export|Button Barrel Export]]
- [[_COMMUNITY_SelectGroup|SelectGroup]]
- [[_COMMUNITY_Backend Server main|Backend Server main]]
- [[_COMMUNITY_Config|Config]]
- [[_COMMUNITY_LoadConfig|LoadConfig]]
- [[_COMMUNITY_createJobRequest|createJobRequest]]
- [[_COMMUNITY_jobResponse|jobResponse]]
- [[_COMMUNITY_assetResponse|assetResponse]]
- [[_COMMUNITY_Store|Store]]

## God Nodes (most connected - your core abstractions)
1. `main()` - 21 edges
2. `NewJob()` - 16 edges
3. `cn (class merge utility)` - 16 edges
4. `LoadConfig()` - 15 edges
5. `openTestStore()` - 15 edges
6. `handleExtensionMessage()` - 12 edges
7. `Segment` - 11 edges
8. `newTestServer()` - 10 edges
9. `render_vtt()` - 10 edges
10. `transcribe_audio()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `VTT cue parsing and rendering (backend)` --semantically_similar_to--> `render_vtt (VTT rendering)`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-04-29-openai-compatible-translation.md → whisper/src/whisper_cli/vtt.py
- `render_vtt (VTT rendering)` --semantically_similar_to--> `Job (GORM model)`  [INFERRED] [semantically similar]
  whisper/src/whisper_cli/vtt.py → backend/internal/store/models.go
- `OpenAI-compatible translation implementation plan` --references--> `Job (GORM model)`  [INFERRED]
  docs/superpowers/plans/2026-04-29-openai-compatible-translation.md → backend/internal/store/models.go
- `YouTube subtitle translation PRD` --references--> `Whisper CLI exit codes`  [EXTRACTED]
  docs/PRD.md → whisper/src/whisper_cli/cli.py
- `YouTube subtitle translation PRD` --references--> `WebVTT format`  [EXTRACTED]
  docs/PRD.md → whisper/src/whisper_cli/vtt.py

## Hyperedges (group relationships)
- **Popup Job Submission and Polling Flow** — popup_app, form_validation, background_service_worker, job_polling [INFERRED 0.85]
- **Content Script Subtitle Display and Mode Switching Flow** — youtube_content_script, youtube_overlay, page_watch, mode_switching_rollback [INFERRED 0.80]
- **Extension Storage Layer (Settings + Cache)** — settings_storage, subtitle_cache_storage, normalizeBackendBaseUrl [INFERRED 0.80]
- **Card Compound Component Pattern** — card_Card, card_CardHeader, card_CardTitle, card_CardDescription, card_CardContent, card_CardFooter [EXTRACTED 1.00]
- **Select Compound Component Pattern** — select_SelectItem, select_SelectTrigger, select_SelectContent, select_SelectGroup, select_SelectSeparator [EXTRACTED 1.00]
- **UI Component Shared Dependency Layer (cn, reka-ui, vueuse, lucide)** — cn_utility, reka_ui, vueuse_core, lucide_vue_next [INFERRED 0.80]
- **Job Completion Polling and Notification Flow** — job-monitor_startJobMonitor, job-monitor_cacheAndNotify, job-monitor_notifyYoutubeWatchTabs, backend-client_BackendClient [INFERRED 0.85]
- **Extension Message Dispatch Architecture** — message-handler_handleExtensionMessage, messages_ExtensionMessage, backend-client_createBackendClient, job-monitor_startJobMonitor [INFERRED 0.90]
- **VTT Parsing and Active Cue Rendering** — vtt_parseVtt, vtt_VttCue, active-cue_findActiveCue [INFERRED 0.90]
- **Real Runner Job Pipeline** — runner_downloadAudio, runner_transcribeAudio, runner_parseWebVTTCues, runner_Translator, runner_renderTranslatedVTT, runner_renderBilingualVTT [EXTRACTED 1.00]
- **Mock Runner Job Pipeline (writes mock VTT constants)** — runner_MockRunner, runner_mockVTTConstants, runner_Store [EXTRACTED 1.00]
- **API Handler Routes** — api_handleJobs, api_handleJobByID, api_handleSubtitleAssets, api_handleSubtitleFile [EXTRACTED 1.00]
- **Whisper CLI pipeline: input → transcribe → validate → render → output** — whisper_cli_cli_main, whisper_cli_transcribe_audio, whisper_cli_render_vtt, whisper_cli_exit_codes [EXTRACTED 1.00]
- **Job lifecycle: create → state transitions → subtitle asset** — store_Job, store_statusConstants, store_Store, store_SubtitleAsset [INFERRED 0.85]
- **RealRunner translation pipeline: VTT parse → cue extract → translate → render translated/bilingual** — vtt_cue_parsing_concept, chat_translator_concept, translator_concept, translation_context_window [INFERRED 0.80]
- **Job Lifecycle Pipeline: queued -> downloading -> transcribing -> translating -> packaging -> completed** — concept_job-state-machine, concept_mock-runner, concept_real-runner, concept_download-audio, concept_translator-interface [EXTRACTED 1.00]
- **Extension Job Submission and Subtitle Display Flow** — concept_backend-client, concept_message-handler, concept_extension-storage-cache, concept_vtt-parser-extension, concept_youtube-video-id-parser [EXTRACTED 0.90]
- **Whisper CLI produces source.vtt which backend feeds to LLM translator** — concept_whisper-cli, concept_vtt-generator-whisper, concept_chat-completions-translator, concept_webvtt-format [INFERRED 0.85]

## Communities

### Community 0 - ".Start downloadAudio"
Cohesion: 0.06
Nodes (38): Handler, NewHTTPHandler, checkTools, ChatTranslator, Cue, Translator Interface, Chat Completion Request, Cue (+30 more)

### Community 1 - "NewJob openTestStore"
Cohesion: 0.09
Nodes (32): execCall, fakeTranslator, NewMockRunner(), openTestStore(), TestMockRunnerCompletesJobAndWritesAssets(), TestMockRunnerFailsJobWhenCompletionUpdateFails(), TestMockRunnerMarksCanceledJobAsFailed(), TestMockRunnerMarksFailedWhenTranscribingUpdateFails() (+24 more)

### Community 2 - "handler test.go"
Cohesion: 0.1
Nodes (23): NewHandler(), newTestServer(), TestPostJobsCreatesJobAndCompletesWithMockRunner(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion(), TestSubtitleFileRejectsInvalidMode(), TestSubtitleFileRejectsNonRegularFile(), TestSubtitleFileRejectsPathOutsideJobDir() (+15 more)

### Community 3 - "handleExtensionMessage normalizeBackendBaseUrl"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 4 - "Segment transcribe audio"
Cohesion: 0.14
Nodes (21): Exception, FakeModel, FakeSegment, test_transcribe_audio_allows_empty_segment_output(), test_transcribe_audio_rejects_english_only_model_with_non_english_language(), test_transcribe_audio_uses_sdk_and_builds_result(), test_transcribe_audio_uses_sdk_reported_duration(), test_vtt_rejects_blank_lines_in_cue_text() (+13 more)

### Community 5 - "response.go writeError"
Cohesion: 0.13
Nodes (18): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+10 more)

### Community 6 - "YouTube subtitle translation PRD"
Cohesion: 0.09
Nodes (24): isAllowedLocalOrigin, routeHandler (interface), Routes (API route definition), withCORS (CORS middleware), ChatTranslator (OpenAI-compatible), CORS localhost-only policy, Job deduplication (videoId + targetLanguage), Job state machine (+16 more)

### Community 7 - "job-monitor.ts pollJob"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 8 - "cn shadcn-vue component pattern"
Cohesion: 0.16
Nodes (24): Alert, AlertDescription, AlertTitle, alertVariants (cva), Button, buttonVariants (cva), Card, CardContent (+16 more)

### Community 9 - "main test cli.py"
Cohesion: 0.16
Nodes (20): fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails(), test_cli_returns_code_2_when_input_file_is_not_readable(), test_cli_returns_code_2_when_language_code_is_invalid() (+12 more)

### Community 10 - "LoadConfig config test.go"
Cohesion: 0.18
Nodes (16): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+8 more)

### Community 11 - "handleExtensionMessage BackendClient Interface"
Cohesion: 0.13
Nodes (19): findActiveCue, BackendClient Interface, BackendClientError, createBackendClient Factory, normalizeBackendBaseUrl Validator, PersistedJobMonitor Type, cacheAndNotify, handleJobMonitorAlarm (+11 more)

### Community 12 - "YoutubeOverlay.vue changeMode"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 13 - "Popup App Component"
Cohesion: 0.14
Nodes (16): CreateJobForm Type, DEFAULT_SETTINGS Constant, Form Validation (validateCreateJobForm), Job Status Polling (popup), Subtitle Mode Switching with Rollback, YouTube Page Watch (getCurrentVideoId, watchVideoIdChanges), Popup App Component, Popup Entry Mount (+8 more)

### Community 14 - "translator.go .translateOne"
Cohesion: 0.16
Nodes (11): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, Translation Prompt Structure, translationCue, translationPrompt (+3 more)

### Community 15 - "Backend Mock MVP Design Spec"
Cohesion: 0.17
Nodes (15): CORS Middleware, downloadAudio Function, Job State Machine, LSI_RUNNER_MODE Configuration, MockRunner, RealRunner, Runner Interface, YouTube Video ID Parser (+7 more)

### Community 16 - "NewHTTPHandler NewChatTranslator"
Cohesion: 0.27
Nodes (11): checkTools(), NewHTTPHandler(), TestCheckToolsRequiresWhisperCLIInRealMode(), NewChatTranslator(), makeTranslatorTestCues(), TestChatTranslatorFailsOnNon2xx(), TestChatTranslatorFailsWhenTranslationMissing(), TestChatTranslatorPromptRequiresTranslationField() (+3 more)

### Community 17 - "Store Open"
Cohesion: 0.2
Nodes (4): ErrNotFound, foreignKeyDSN(), Open(), Store

### Community 18 - "ParseVideoID youtube test.go"
Cohesion: 0.39
Nodes (6): ParseVideoID(), TestParseVideoIDRejectsNonWatchYouTubeURL(), TestParseVideoIDRejectsUnsupportedScheme(), TestParseVideoIDRejectsUnsupportedURL(), TestParseVideoIDSupportsShortURL(), TestParseVideoIDSupportsWatchURL()

### Community 19 - "toJobResponse formatTime"
Cohesion: 0.29
Nodes (7): ParseVideoID, formatTime, handleJobByID, handleJobs, handleSubtitleAssets, toAssetResponse, toJobResponse

### Community 20 - "page-watch.ts getCurrentVideoId"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 21 - "Runner Interface"
Cohesion: 0.4
Nodes (5): api.Runner Interface, MockRunner, RealRunner, Runner Interface, Mock VTT Constants

### Community 22 - "BackendClient MessageHandler"
Cohesion: 0.5
Nodes (4): BackendClient, Backend HTTP API, Extension Storage and Cache, MessageHandler

### Community 24 - "vtt.ts parseTimestamp"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 25 - "runner.go Runner"
Cohesion: 0.67
Nodes (2): Runner, Store

### Community 26 - "cn Badge Component Barrel"
Cohesion: 0.67
Nodes (3): Badge Component Barrel, cn (class merge utility), Select Component Barrel

### Community 27 - "runner.Store Interface"
Cohesion: 0.67
Nodes (3): api.Store Interface, runner.Store Interface, recordingStore (test fake)

### Community 28 - "whisper-cli Command"
Cohesion: 1.0
Nodes (3): whisper-cli Command, Whisper CLI Implementation Plan, Whisper CLI Design Spec

### Community 29 - "VTT Generator"
Cohesion: 1.0
Nodes (3): VTT Generator (Whisper), VTT Parser (Extension), WebVTT Format

### Community 34 - "Extension Manifest Configuration"
Cohesion: 1.0
Nodes (2): Extension Manifest Configuration, WXT Extension Config

### Community 35 - "handleSubtitleFile subtitleFilePathAllowed"
Cohesion: 1.0
Nodes (2): handleSubtitleFile, subtitleFilePathAllowed

### Community 36 - "writeError writeJSON"
Cohesion: 1.0
Nodes (2): writeError, writeJSON

### Community 37 - "Chat Completions Translator"
Cohesion: 1.0
Nodes (2): Chat Completions Translator, Translator Interface

### Community 86 - "Vitest Config"
Cohesion: 1.0
Nodes (1): Vitest Config

### Community 87 - "Background Service Worker"
Cohesion: 1.0
Nodes (1): Background Service Worker

### Community 88 - "CardAction UI Component"
Cohesion: 1.0
Nodes (1): CardAction UI Component

### Community 89 - "Card Barrel Export"
Cohesion: 1.0
Nodes (1): Card Barrel Export

### Community 90 - "Input Barrel Export"
Cohesion: 1.0
Nodes (1): Input Barrel Export

### Community 91 - "Alert Barrel Export"
Cohesion: 1.0
Nodes (1): Alert Barrel Export (with cva variants)

### Community 92 - "Separator Barrel Export"
Cohesion: 1.0
Nodes (1): Separator Barrel Export

### Community 93 - "Button Barrel Export"
Cohesion: 1.0
Nodes (1): Button Barrel Export

### Community 94 - "SelectGroup"
Cohesion: 1.0
Nodes (1): SelectGroup

### Community 95 - "Backend Server main"
Cohesion: 1.0
Nodes (1): Backend Server main

### Community 96 - "Config"
Cohesion: 1.0
Nodes (1): Config

### Community 97 - "LoadConfig"
Cohesion: 1.0
Nodes (1): LoadConfig

### Community 98 - "createJobRequest"
Cohesion: 1.0
Nodes (1): createJobRequest

### Community 99 - "jobResponse"
Cohesion: 1.0
Nodes (1): jobResponse

### Community 100 - "assetResponse"
Cohesion: 1.0
Nodes (1): assetResponse

### Community 101 - "Store"
Cohesion: 1.0
Nodes (1): Store (SQLite persistence)

## Knowledge Gaps
- **102 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+97 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `vtt.ts parseTimestamp`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `runner.go Runner`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Extension Manifest Configuration`** (2 nodes): `Extension Manifest Configuration`, `WXT Extension Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `handleSubtitleFile subtitleFilePathAllowed`** (2 nodes): `handleSubtitleFile`, `subtitleFilePathAllowed`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `writeError writeJSON`** (2 nodes): `writeError`, `writeJSON`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Chat Completions Translator`** (2 nodes): `Chat Completions Translator`, `Translator Interface`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Vitest Config`** (1 nodes): `Vitest Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Background Service Worker`** (1 nodes): `Background Service Worker`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `CardAction UI Component`** (1 nodes): `CardAction UI Component`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Card Barrel Export`** (1 nodes): `Card Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Input Barrel Export`** (1 nodes): `Input Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Alert Barrel Export`** (1 nodes): `Alert Barrel Export (with cva variants)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Separator Barrel Export`** (1 nodes): `Separator Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Button Barrel Export`** (1 nodes): `Button Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `SelectGroup`** (1 nodes): `SelectGroup`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Backend Server main`** (1 nodes): `Backend Server main`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Config`** (1 nodes): `Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `LoadConfig`** (1 nodes): `LoadConfig`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `createJobRequest`** (1 nodes): `createJobRequest`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `jobResponse`** (1 nodes): `jobResponse`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `assetResponse`** (1 nodes): `assetResponse`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Store`** (1 nodes): `Store (SQLite persistence)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewHTTPHandler()` connect `NewHTTPHandler NewChatTranslator` to `Store Open`, `LoadConfig config test.go`, `handler test.go`, `NewJob openTestStore`?**
  _High betweenness centrality (0.126) - this node is a cross-community bridge._
- **Why does `Open()` connect `Store Open` to `NewHTTPHandler NewChatTranslator`, `NewJob openTestStore`, `handler test.go`?**
  _High betweenness centrality (0.123) - this node is a cross-community bridge._
- **Why does `Store` connect `Store Open` to `YouTube subtitle translation PRD`?**
  _High betweenness centrality (0.115) - this node is a cross-community bridge._
- **Are the 14 inferred relationships involving `main()` (e.g. with `test_cli_requires_all_required_arguments()` and `test_cli_creates_parent_directory_for_output()`) actually correct?**
  _`main()` has 14 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `NewJob()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`NewJob()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 12 inferred relationships involving `LoadConfig()` (e.g. with `main()` and `TestLoadConfigUsesDefaults()`) actually correct?**
  _`LoadConfig()` has 12 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `openTestStore()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`openTestStore()` has 6 INFERRED edges - model-reasoned connections that need verification._