# Lets Sub It

Self-hosted pipeline that turns a YouTube video's audio into source-language, translated and bilingual subtitles.

## Language

**Job**:
One run of the pipeline (download, transcribe, translate, package) for a video and target language.
_Avoid_: Task, request

**Subtitle Result**:
Everything the service holds for one video and target language: its jobs, the subtitle asset and the job working files. It is the unit that is reused and the unit that is deleted.
_Avoid_: Cache, subtitles (when meaning the whole result)

**Reuse**:
Answering a new submission with an existing non-failed job for the same video and target language instead of running the pipeline again.
_Avoid_: Cache hit

**Deleted Subtitle Result**:
A Subtitle Result that was deleted: hidden from reuse and every lookup, its files gone, its records kept as a history of what was generated. It cannot be restored.
_Avoid_: Purged, removed
