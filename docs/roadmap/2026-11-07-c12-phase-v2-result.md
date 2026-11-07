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

## Observed outcome

Source checkpoint0b73dedddcc847b295b4df98627e62f2c38ee88e pushed/remote verified.
Build09177d/session51632 ->2f7cc3:exit0,23488.7821ms,empty stdout/stderr,
no timeout,owned exit confirmed,4 module hashes unchanged.
Binary SHA25649943D88D223F41405173E0D9A22F0AA00FC972A26A7BF3DEC9F38343A22332C.
Embedded VCS stamping failure resolved for this build; no Git security change.

Live64f41f:ERROR after487.077ms,no timeout,owned exit confirmed,exit-1.
Failure: ReadAllBytes cannot read stdout.raw while its writer handle is open.
Raw stdout212bytes,stderr0bytes. One valid text PREPARE START event retained,
session8e96122fe99c454d8faed1549344f7dc; PREPARE INCOMPLETE. Text ZIP/DIRECTORY
and all sqlite phases NOT_RUN. No completed phase duration or digest comparison.
4 module files unchanged. Both approved slots consumed; no retry/fix after failure.

Read-only investigationd5f32c: owner.ps1 opens stdout.raw for Write with share Read,
then calls Read-C12PhaseEvents before releasing that writer. Parser uses
File.ReadAllBytes,whose open conflicts with the existing writer. After finally
closes the writer,the identical saved START record parses successfully. This is
an owner/parser integration defect,not evidence of a module cache/Go timeout.
Existing parser fixtures close files before parsing; both owned-process tests
omit Session,so neither exercises live parsing with an open writer.

Next recommended scope: offline owner/parser integration regression and minimal
compatible read-handle correction; test actual simultaneous writer/reader,partial
records,protocol failure and owned exit. No new real cache workload until that
gap is verified. Any future live attempt needs fresh approval/root; no implicit
reuse of this consumed slot. Do not loosen quotas,identity/path checks or deadline.

Archive portable-c12-cache-phase-v2.json retains exact source,claims,receipts,
bounded original streams and ledger; excludes binary,caches and module contents.
Earlier archives remain untouched. This diagnostic omits root graph loading,
uses sequential phases and potentially warmed caches; no original root-cause,
physical-disk,full-suite or independent-security acceptance.44/34/6/217,NOT_ACCEPTED.
