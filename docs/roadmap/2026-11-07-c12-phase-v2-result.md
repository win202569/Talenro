# Selected cache phase v2 — approved execution

User approved afterf97a7cf4: one180s build with -buildvcs=false,then at most one
120s live helper. This disables embedded Git metadata only; explicit source
checkpoint/source hashes and binary SHA remain the provenance. No Git configuration,
ownership or security settings change. No additional test invocations/retries.

Only invocation changes: new exclusive workspace
.superpowers/sdd/2026-11-07-c12-cache-phase-v2 and the build flag. All Go/runtime/
owner/vendor sources unchanged from the16-group GREEN and owner10-group checks.
Both old archives/roots remain untouched. No binary/cache contents enter Git.
Original approved design/plan governs module order,read-only input allowlist,
PREPARE/ZIP/DIRECTORY events,120s global cap,5s owned exit wait and failure stop.
Source checkpoint must be committed/pushed before live. No root verifier/C12/
devtools/Docker/trace or full-suite acceptance.44/34/6/217,NOT_ACCEPTED.

Status at source checkpoint: build and live not yet run. Outcomes appended below.
