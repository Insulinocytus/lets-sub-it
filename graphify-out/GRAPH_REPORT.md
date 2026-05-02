# Graph Report - .  (2026-05-02)

## Corpus Check
- 120 files · ~53,866 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 562 nodes · 778 edges · 49 communities detected
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 218 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Backend Runner Pipeline|Backend Runner Pipeline]]
- [[_COMMUNITY_Mock Runner & Testing|Mock Runner & Testing]]
- [[_COMMUNITY_API Handler Tests|API Handler Tests]]
- [[_COMMUNITY_Extension Backend Client|Extension Backend Client]]
- [[_COMMUNITY_Whisper Transcription Tests|Whisper Transcription Tests]]
- [[_COMMUNITY_Backend API Types & Handlers|Backend API Types & Handlers]]
- [[_COMMUNITY_CORS & Routing Middleware|CORS & Routing Middleware]]
- [[_COMMUNITY_Job Monitor & SSE|Job Monitor & SSE]]
- [[_COMMUNITY_UI Components (shadcn-vue)|UI Components (shadcn-vue)]]
- [[_COMMUNITY_Whisper CLI Tests|Whisper CLI Tests]]
- [[_COMMUNITY_App Configuration|App Configuration]]
- [[_COMMUNITY_Extension Runtime Integration|Extension Runtime Integration]]
- [[_COMMUNITY_YouTube Overlay Content Script|YouTube Overlay Content Script]]
- [[_COMMUNITY_Popup UI Logic|Popup UI Logic]]
- [[_COMMUNITY_LLM Chat Translator|LLM Chat Translator]]
- [[_COMMUNITY_Architecture Concepts & Specs|Architecture Concepts & Specs]]
- [[_COMMUNITY_Server App Initialization|Server App Initialization]]
- [[_COMMUNITY_SQLite Store Layer|SQLite Store Layer]]
- [[_COMMUNITY_YouTube URL Parsing Tests|YouTube URL Parsing Tests]]
- [[_COMMUNITY_API Route Handlers|API Route Handlers]]
- [[_COMMUNITY_YouTube Watch Detection|YouTube Watch Detection]]
- [[_COMMUNITY_Runner Interface & Mock|Runner Interface & Mock]]
- [[_COMMUNITY_Extension Architecture Concepts|Extension Architecture Concepts]]
- [[_COMMUNITY_WebVTT Parsing|WebVTT Parsing]]
- [[_COMMUNITY_Runner-Store Integration|Runner-Store Integration]]
- [[_COMMUNITY_Badge & Select UI|Badge & Select UI]]
- [[_COMMUNITY_Store Interface|Store Interface]]
- [[_COMMUNITY_Whisper CLI Design|Whisper CLI Design]]
- [[_COMMUNITY_VTT Format Concepts|VTT Format Concepts]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 89|Community 89]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 92|Community 92]]
- [[_COMMUNITY_Community 93|Community 93]]
- [[_COMMUNITY_Community 94|Community 94]]
- [[_COMMUNITY_Community 95|Community 95]]
- [[_COMMUNITY_Community 96|Community 96]]
- [[_COMMUNITY_Community 97|Community 97]]
- [[_COMMUNITY_Community 98|Community 98]]
- [[_COMMUNITY_Community 99|Community 99]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]

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

### Community 0 - "Backend Runner Pipeline"
Cohesion: 0.06
Nodes (38): Handler, NewHTTPHandler, checkTools, ChatTranslator, Cue, Translator Interface, Chat Completion Request, Cue (+30 more)

### Community 1 - "Mock Runner & Testing"
Cohesion: 0.09
Nodes (32): execCall, fakeTranslator, NewMockRunner(), openTestStore(), TestMockRunnerCompletesJobAndWritesAssets(), TestMockRunnerFailsJobWhenCompletionUpdateFails(), TestMockRunnerMarksCanceledJobAsFailed(), TestMockRunnerMarksFailedWhenTranscribingUpdateFails() (+24 more)

### Community 2 - "API Handler Tests"
Cohesion: 0.1
Nodes (23): NewHandler(), newTestServer(), TestPostJobsCreatesJobAndCompletesWithMockRunner(), TestPostJobsRejectsMissingSourceLanguage(), TestSubtitleAssetReturnsAssetAfterCompletion(), TestSubtitleFileRejectsInvalidMode(), TestSubtitleFileRejectsNonRegularFile(), TestSubtitleFileRejectsPathOutsideJobDir() (+15 more)

### Community 3 - "Extension Backend Client"
Cohesion: 0.13
Nodes (23): BackendClientError, createBackendClient(), errorFromResponse(), invalidBackendBaseUrlError(), normalizeBackendBaseUrl(), request(), requestJson(), clientFromSettings() (+15 more)

### Community 4 - "Whisper Transcription Tests"
Cohesion: 0.14
Nodes (21): Exception, FakeModel, FakeSegment, test_transcribe_audio_allows_empty_segment_output(), test_transcribe_audio_rejects_english_only_model_with_non_english_language(), test_transcribe_audio_uses_sdk_and_builds_result(), test_transcribe_audio_uses_sdk_reported_duration(), test_vtt_rejects_blank_lines_in_cue_text() (+13 more)

### Community 5 - "Backend API Types & Handlers"
Cohesion: 0.13
Nodes (18): apiError, assetResponse, createJobRequest, errorBody, Handler, subtitleFileNameForMode(), subtitleFilePathAllowed(), jobResponse (+10 more)

### Community 6 - "CORS & Routing Middleware"
Cohesion: 0.09
Nodes (24): isAllowedLocalOrigin, routeHandler (interface), Routes (API route definition), withCORS (CORS middleware), ChatTranslator (OpenAI-compatible), CORS localhost-only policy, Job deduplication (videoId + targetLanguage), Job state machine (+16 more)

### Community 7 - "Job Monitor & SSE"
Cohesion: 0.18
Nodes (21): addPersistedJobMonitor(), cacheAndNotify(), createMonitorKey(), ensureJobMonitorAlarm(), ensurePersistedJobMonitors(), getJobMonitorAlarmName(), getMonitorBackendBaseUrl(), getMonitorClient() (+13 more)

### Community 8 - "UI Components (shadcn-vue)"
Cohesion: 0.16
Nodes (24): Alert, AlertDescription, AlertTitle, alertVariants (cva), Button, buttonVariants (cva), Card, CardContent (+16 more)

### Community 9 - "Whisper CLI Tests"
Cohesion: 0.16
Nodes (20): fake_result(), test_cli_creates_parent_directory_for_output(), test_cli_prints_json_on_success(), test_cli_rejects_output_path_matching_input_path(), test_cli_requires_all_required_arguments(), test_cli_returns_code_2_when_creating_output_directory_fails(), test_cli_returns_code_2_when_input_file_is_not_readable(), test_cli_returns_code_2_when_language_code_is_invalid() (+12 more)

### Community 10 - "App Configuration"
Cohesion: 0.18
Nodes (16): Config, envDurationOrDefault(), envOrDefault(), LoadConfig(), TestLoadConfigDownloadTimeoutCustom(), TestLoadConfigDownloadTimeoutDefault(), TestLoadConfigDownloadTimeoutInvalid(), TestLoadConfigLLMCustomValues() (+8 more)

### Community 11 - "Extension Runtime Integration"
Cohesion: 0.13
Nodes (19): findActiveCue, BackendClient Interface, BackendClientError, createBackendClient Factory, normalizeBackendBaseUrl Validator, PersistedJobMonitor Type, cacheAndNotify, handleJobMonitorAlarm (+11 more)

### Community 12 - "YouTube Overlay Content Script"
Cohesion: 0.22
Nodes (13): bindVideo(), canUpdate(), changeMode(), cleanupVideoListeners(), handleModeClick(), handleRuntimeMessage(), isSubtitleUpdatedMessage(), loadForVideo() (+5 more)

### Community 13 - "Popup UI Logic"
Cohesion: 0.14
Nodes (16): CreateJobForm Type, DEFAULT_SETTINGS Constant, Form Validation (validateCreateJobForm), Job Status Polling (popup), Subtitle Mode Switching with Rollback, YouTube Page Watch (getCurrentVideoId, watchVideoIdChanges), Popup App Component, Popup Entry Mount (+8 more)

### Community 14 - "LLM Chat Translator"
Cohesion: 0.16
Nodes (11): chatCompletionRequest, chatCompletionResponse, chatMessage, chatResponseFormat, ChatTranslator, Translation Prompt Structure, translationCue, translationPrompt (+3 more)

### Community 15 - "Architecture Concepts & Specs"
Cohesion: 0.17
Nodes (15): CORS Middleware, downloadAudio Function, Job State Machine, LSI_RUNNER_MODE Configuration, MockRunner, RealRunner, Runner Interface, YouTube Video ID Parser (+7 more)

### Community 16 - "Server App Initialization"
Cohesion: 0.27
Nodes (11): checkTools(), NewHTTPHandler(), TestCheckToolsRequiresWhisperCLIInRealMode(), NewChatTranslator(), makeTranslatorTestCues(), TestChatTranslatorFailsOnNon2xx(), TestChatTranslatorFailsWhenTranslationMissing(), TestChatTranslatorPromptRequiresTranslationField() (+3 more)

### Community 17 - "SQLite Store Layer"
Cohesion: 0.2
Nodes (4): ErrNotFound, foreignKeyDSN(), Open(), Store

### Community 18 - "YouTube URL Parsing Tests"
Cohesion: 0.39
Nodes (6): ParseVideoID(), TestParseVideoIDRejectsNonWatchYouTubeURL(), TestParseVideoIDRejectsUnsupportedScheme(), TestParseVideoIDRejectsUnsupportedURL(), TestParseVideoIDSupportsShortURL(), TestParseVideoIDSupportsWatchURL()

### Community 19 - "API Route Handlers"
Cohesion: 0.29
Nodes (7): ParseVideoID, formatTime, handleJobByID, handleJobs, handleSubtitleAssets, toAssetResponse, toJobResponse

### Community 20 - "YouTube Watch Detection"
Cohesion: 0.47
Nodes (4): getCurrentVideoId(), getVideoIdFromLocationHref(), watchVideoIdChanges(), parseYouTubeWatchVideoId()

### Community 21 - "Runner Interface & Mock"
Cohesion: 0.4
Nodes (5): api.Runner Interface, MockRunner, RealRunner, Runner Interface, Mock VTT Constants

### Community 22 - "Extension Architecture Concepts"
Cohesion: 0.5
Nodes (4): BackendClient, Backend HTTP API, Extension Storage and Cache, MessageHandler

### Community 24 - "WebVTT Parsing"
Cohesion: 1.0
Nodes (2): parseTimestamp(), parseVtt()

### Community 25 - "Runner-Store Integration"
Cohesion: 0.67
Nodes (2): Runner, Store

### Community 26 - "Badge & Select UI"
Cohesion: 0.67
Nodes (3): Badge Component Barrel, cn (class merge utility), Select Component Barrel

### Community 27 - "Store Interface"
Cohesion: 0.67
Nodes (3): api.Store Interface, runner.Store Interface, recordingStore (test fake)

### Community 28 - "Whisper CLI Design"
Cohesion: 1.0
Nodes (3): whisper-cli Command, Whisper CLI Implementation Plan, Whisper CLI Design Spec

### Community 29 - "VTT Format Concepts"
Cohesion: 1.0
Nodes (3): VTT Generator (Whisper), VTT Parser (Extension), WebVTT Format

### Community 34 - "Community 34"
Cohesion: 1.0
Nodes (2): Extension Manifest Configuration, WXT Extension Config

### Community 35 - "Community 35"
Cohesion: 1.0
Nodes (2): handleSubtitleFile, subtitleFilePathAllowed

### Community 36 - "Community 36"
Cohesion: 1.0
Nodes (2): writeError, writeJSON

### Community 37 - "Community 37"
Cohesion: 1.0
Nodes (2): Chat Completions Translator, Translator Interface

### Community 86 - "Community 86"
Cohesion: 1.0
Nodes (1): Vitest Config

### Community 87 - "Community 87"
Cohesion: 1.0
Nodes (1): Background Service Worker

### Community 88 - "Community 88"
Cohesion: 1.0
Nodes (1): CardAction UI Component

### Community 89 - "Community 89"
Cohesion: 1.0
Nodes (1): Card Barrel Export

### Community 90 - "Community 90"
Cohesion: 1.0
Nodes (1): Input Barrel Export

### Community 91 - "Community 91"
Cohesion: 1.0
Nodes (1): Alert Barrel Export (with cva variants)

### Community 92 - "Community 92"
Cohesion: 1.0
Nodes (1): Separator Barrel Export

### Community 93 - "Community 93"
Cohesion: 1.0
Nodes (1): Button Barrel Export

### Community 94 - "Community 94"
Cohesion: 1.0
Nodes (1): SelectGroup

### Community 95 - "Community 95"
Cohesion: 1.0
Nodes (1): Backend Server main

### Community 96 - "Community 96"
Cohesion: 1.0
Nodes (1): Config

### Community 97 - "Community 97"
Cohesion: 1.0
Nodes (1): LoadConfig

### Community 98 - "Community 98"
Cohesion: 1.0
Nodes (1): createJobRequest

### Community 99 - "Community 99"
Cohesion: 1.0
Nodes (1): jobResponse

### Community 100 - "Community 100"
Cohesion: 1.0
Nodes (1): assetResponse

### Community 101 - "Community 101"
Cohesion: 1.0
Nodes (1): Store (SQLite persistence)

## Knowledge Gaps
- **102 isolated node(s):** `Config`, `Runner`, `Store`, `execCall`, `Cue` (+97 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `WebVTT Parsing`** (3 nodes): `vtt.ts`, `parseTimestamp()`, `parseVtt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Runner-Store Integration`** (3 nodes): `runner.go`, `Runner`, `Store`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 34`** (2 nodes): `Extension Manifest Configuration`, `WXT Extension Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 35`** (2 nodes): `handleSubtitleFile`, `subtitleFilePathAllowed`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 36`** (2 nodes): `writeError`, `writeJSON`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 37`** (2 nodes): `Chat Completions Translator`, `Translator Interface`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 86`** (1 nodes): `Vitest Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 87`** (1 nodes): `Background Service Worker`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 88`** (1 nodes): `CardAction UI Component`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 89`** (1 nodes): `Card Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 90`** (1 nodes): `Input Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 91`** (1 nodes): `Alert Barrel Export (with cva variants)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 92`** (1 nodes): `Separator Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 93`** (1 nodes): `Button Barrel Export`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 94`** (1 nodes): `SelectGroup`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 95`** (1 nodes): `Backend Server main`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 96`** (1 nodes): `Config`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 97`** (1 nodes): `LoadConfig`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 98`** (1 nodes): `createJobRequest`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 99`** (1 nodes): `jobResponse`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 100`** (1 nodes): `assetResponse`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 101`** (1 nodes): `Store (SQLite persistence)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewHTTPHandler()` connect `Server App Initialization` to `SQLite Store Layer`, `App Configuration`, `API Handler Tests`, `Mock Runner & Testing`?**
  _High betweenness centrality (0.126) - this node is a cross-community bridge._
- **Why does `Open()` connect `SQLite Store Layer` to `Server App Initialization`, `Mock Runner & Testing`, `API Handler Tests`?**
  _High betweenness centrality (0.123) - this node is a cross-community bridge._
- **Why does `Store` connect `SQLite Store Layer` to `CORS & Routing Middleware`?**
  _High betweenness centrality (0.115) - this node is a cross-community bridge._
- **Are the 14 inferred relationships involving `main()` (e.g. with `test_cli_requires_all_required_arguments()` and `test_cli_creates_parent_directory_for_output()`) actually correct?**
  _`main()` has 14 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `NewJob()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`NewJob()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 12 inferred relationships involving `LoadConfig()` (e.g. with `main()` and `TestLoadConfigUsesDefaults()`) actually correct?**
  _`LoadConfig()` has 12 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `openTestStore()` (e.g. with `TestRealRunnerCompletesJob()` and `TestRealRunnerDownloadFailed()`) actually correct?**
  _`openTestStore()` has 6 INFERRED edges - model-reasoned connections that need verification._