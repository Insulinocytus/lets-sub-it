# Logical deletion of Subtitle Results

Deleting a Subtitle Result keeps its job and subtitle asset rows and only marks the jobs deleted (`jobs.deleted_at`), while its working directory (audio and VTT files) is removed. The rows are a record of what was generated, when, and with what outcome, much like a log; they are not kept to support restore, and no restore exists. Files are the bulk of the disk usage and carry nothing the record needs.

## Considered Options

- Keep rows and files: allows restore, but disk use only grows (every `task api:smoke` run leaves another audio file).
- Keep rows and VTT, drop audio: smaller, but still keeps content nobody reads.
- Delete rows physically: loses the record.

## Consequences

- Only jobs carry the deletion mark; assets are hidden through their job. An asset written by a runner after its job was deleted therefore never becomes visible.
- Deleted jobs cannot be restored, because their files are gone.
