# C12 Selected Cache Phase Timing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task, with the approved inline author-review boundary. Await plan approval before implementing or running anything.

**Goal:** Time selected ZIP and directory hashing independently, retaining incomplete-phase evidence at a fixed cutoff.

**Architecture:** A standard-library-only Go helper uses a byte-exact local copy of pinned dirhash. A restricted PowerShell owner controls build/test/live process lifetimes, drains bounded output and preserves exclusive evidence. No live collector or authoritative root verifier is launched.

**Tech Stack:** Existing PowerShell7.6.5,Go1.26.5,.NET process APIs,Windows file attributes.

**Spec:** docs/superpowers/specs/2026-11-06-c12-cache-phase-timing-design.md, approved by user afterf8eb10d7. This plan needs separate approval.

## Global constraints and budgets

- Existing codex/c12-b01-task9-coordinator worktree; no production code/root or
  devtools go.mod/go.sum changes.44/34/6/217 unchanged,NOT_ACCEPTED. No main merge.
- New evidence root `.superpowers/sdd/2026-11-06-c12-cache-phase-v1/` with exclusive
  OwnerRed,OwnerGreen,GoRed,GoGreen,Build,Live subdirectories. Never reuse/delete
  old evidence. No security changes,downloads,system tracing,CIM,Docker orC12.
- One PowerShell owner RED and one GREEN invocation (at most10 cases each).
  GREEN may launch exactly two harmless owned PowerShell fixtures: normal exit,
  and sleeping child cut off after1000ms; exit wait5000ms. No fixtures in RED.
- Two Go test invocations: RED and GREEN, each one compiler/test process tree,
 180s outer bound including compilation and `go test -timeout 30s` inside. At most
 16 test groups each, synthetic inputs only. Missing prerequisite/compile error is
 not expected RED: provide minimal compiling unimplemented stubs for new APIs.
- After all GREEN gates: one release helper build,180s outer bound. No probe build,
  retries or fallback compiler. At most one live helper run,120s including its
  safety preparation. Each owned process/tree gets at most5s exit confirmation.
- All commands restricted,using pinned
  `C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe`,
  SHA256D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.
- Child-only environment removes inherited GO*/GIT_* then sets GOENV=off,GOWORK=off,
  GOTOOLCHAIN=local,GOPROXY=off,GOSUMDB=off,GOAUTH=off,GOVCS=all:off,
  GOFLAGS=-mod=readonly,GOMAXPROCS=2,CGO_ENABLED=0. GOPATH,GOMODCACHE,GOCACHE,
  GOTMPDIR,TEMP,TMP point to new workspace subdirectories; GOCACHE may be shared
  between this plan's approved Go invocations only. Live input cache root is an
  explicit allowlisted read path, not the build environment's GOMODCACHE.
- Live selection/order exactly text v0.40.0 ZIP/directory,then sqlite v1.49.1
  ZIP/directory. Missing input,hash mismatch,safety rejection or timeout stops all
  later phases,which are NOT_RUN. No substitute module or automatic root verify.
-64KiB/event,4MiB total retained live stream/records; bound build/test output too.
  No contents,per-file path lists,environment dumps or unrelated process data in receipts.
- Expected RED assertion failures permit implementation; any other failure/timeout
  stops further tests/build/live work. Preserve evidence and hand off,never silently
  consume extra slots. Full-suite instructions do not override this scoped budget.

## File map and interfaces

Under `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/cache-phase/` create:
- `go.mod`: module `talenro.local/c12-cache-phase`,go1.26.0,no requirements.
- `internal/dirhash/hash.go`,LICENSE,PATENTS: exact pinned vendored files.
- `provenance.json`: toolchain/source hashes,origin and license metadata.
- `phase.go`,`phase_test.go`: path validation,expected digest parsing,phase runner.
- `main.go`: fixed live allowlist/CLI and JSONL event output; no test-only live mode.
- `owner.ps1`: definition-only bounded process/stream/receipt helpers.
- `owner.tests.ps1`: `-Phase OwnerRed|OwnerGreen`,exclusive evidence.
- `invoke.ps1`: `-Stage GoRed|GoGreen|Build|Live`,one claimed stage per invocation.

Go interfaces: `Candidate{Module,Version,Zip,Dir,Expected string}`;
`Event{Schema,Sequence int; Session,Module,Version,Phase,Kind,Status,Digest string;
ElapsedNs int64}`. Phases PREPARE,ZIP,DIRECTORY; Kind START/END; timing uses Go
monotonic time. `ValidateCandidate(c Candidate) error`,
`ReadExpected(sumText,ziphashText,module,version string)(string,error)`,
`Run(candidates []Candidate,session string,emit func(Event) error) error`.
Live CLI accepts only `-session <32hex> -repo <verified worktree>`; fixed cache
root `C:/Users/Lenovo/go/pkg/mod`. No arbitrary candidate/file/command arguments.

PowerShell interfaces: `Invoke-C12PhaseOwned -Executable -Arguments -WorkingDirectory
-Environment -Root -DeadlineMs` returns Started,TimedOut,ExitConfirmed,ExitCode,
ElapsedMs,EndReason and bounded stream locations/hashes. `Read-C12PhaseEvents
-Path -Session` returns validated ordered events plus Complete/Incomplete/NotRun
phase state. Pure parsing never starts a process. Wrapper validates exact stage
command and executable hash before passing them to the generic owned-process helper.

## Review focus

1. Metadata inspection follows a reparse point: validate ancestors and Windows
   FILE_ATTRIBUTE_REPARSE_POINT on every visited entry before recursing/opening.
2. ZIP traversal/duplicate names or wrong prefix: reject,never extract; synthetic tests.
3. Same hashing name but different implementation: pinned byte hash and literal digest fixture.
4. Parent output read blocks cutoff or partial line becomes success: chunked async
   drains with byte caps and independent monotonic deadline; protocol tests.
5. Surrogate timing overclaimed: PREPARE markers,INCOMPLETE state,graph/concurrency/
   cache-warmth caveats mandatory in final report.

## Task1 — Owned process and event contracts

- [ ] Write10 cases before implementation: sequence/session mismatch,partial final
  line,START without END,illegal phase order,64KiB event cap,4MiB total cap,
  occupied output preservation,deadline decision,normal exit,1000ms owned cutoff.
  RED records missing owner implementation before any child launches.
- [ ] Run `owner.tests.ps1 -Phase OwnerRed` once; expected missing implementation.
- [ ] Implement owner.ps1 and invoke.ps1: exclusive paths and atomic final JSON;
  retain raw partial stdout/stderr; async bounded chunk reads,never wait for a whole
  line synchronously. Drain time cannot extend process deadline. Protocol/quota
  failure kills only retained owned process/tree. Unknown exit blocks next stage.
- [ ] Run `owner.tests.ps1 -Phase OwnerGreen` once; expect10 matched cases and
  both permitted fixture exits confirmed. Review and commit explicit files.

## Task2 — Hashing semantics and safety under synthetic fixtures

- [ ] Copy pinned vendored source/license files via reviewed edits,retain notices.
  Source hash.go SHA25624DE6A735C1605E39B6381E6962787FAA5C59ADAB896FB66DBC1AC8B6AFCB297
  (read-onlyff5222). Preserve LF so repository bytes match. Verify LICENSE/PATENTS
  against installed originals and record hashes; do not invent missing license text.
- [ ] Write phase_test.go and minimal compiling API stubs returning an explicit
  unimplemented error.16 groups: empty fixture,one-file ZIP/directory equivalence,
  changed content,missing/conflicting sum,read error,wrong candidate identity,
  traversal ZIP,duplicate ZIP,wrong ZIP prefix,escaped filesystem root,reparse
  metadata refusal,read-only fixture preservation,event order,emitter failure,
  digest mismatch stopping later phases,monotonic duration/nonempty START evidence.
  Reparse bit classifier uses injected synthetic metadata; mark that limitation,
  do not create junctions or change privileges to make a fixture possible.
- [ ] Literal expected empty digest `h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=`.
  File `fixture.example/mod@v1.0.0/a.txt` containing exact UTF-8 `abc`,no newline,
  has expected `h1:xYudg4RCFWzdnRPnaadGTWOBYxaNjPFPGOgJZ3DiUvY=`; independently
  derived from SHA256(abc) and specified Hash1 summary format inff5222,not the helper.
- [ ] Run `invoke.ps1 -Stage GoRed` once: exact pinned `go test -count=1 -timeout=30s ./...`
  in isolated diagnostic module,180s owner bound. Expect assertion failures for
  unimplemented APIs,not compile/permission/toolchain errors. Preserve result.
- [ ] Implement validation and Run using unchanged dirhash functions. Live metadata
  adapters inspect Windows reparse attributes; don't rely solely on Go symlink mode.
  Check canonical containment and safe ZIP names,prefix/duplicates; bound expected
  hash metadata reads. PREPARE START precedes validation and enumeration,END follows;
  then ZIP START/END and DIRECTORY START/END; compare expected h1 before moving on.
  Reject malformed expected digests and duplicate conflicting go.sum entries.
- [ ] Run `invoke.ps1 -Stage GoGreen` once with same command/bounds; expect16 groups
  passing,no real cache reads. Build `main.go` allowlist/stream integration within
  these tests using Run (no production binary launched). Review/commit checkpoint.

## Task3 — One build and one bounded selected-cache run

- [ ] Run `invoke.ps1 -Stage Build` once: pinned `go build -trimpath -o <Build/helper.exe> .`,
 180s bound,offline isolated caches. Record source/binary/module-file hashes and
  command/environment allowlist. Nonzero/timeout/missing prerequisites stop.
- [ ] Commit/push necessary source/tests/provenance before any live read workload;
  never track binary or build cache. Confirm root/devtools module hashes unchanged.
- [ ] Run `invoke.ps1 -Stage Live` once: one verified helper process,120000ms cap,
  fresh session and Live root; exact text/sqlite paths and expected digest sources.
  Parent retains partial events and confirms owned exit within5000ms on cutoff.
  No extra module inventory/hash workload outside this bound.
- [ ] Validate events and report PREPARE/ZIP/DIRECTORY times separately. A missing
  END is INCOMPLETE; subsequent phases NOT_RUN. If all four hashing phases match,
  say selected digests matched,not root verification passed. Report cache warming,
  sequential scheduling and non-atomic cache inspection limitations.

## Task4 — Evidence and recovery

- [ ] Archive exact source/provenance,claims,RED/GREEN/build/live receipts and bounded
  event streams as `portable-c12-cache-phase-v1.json` alongside diagnostics. Label
  synthetic versus live data. No binaries,cache/module contents or credentials.
- [ ] Verify decode,length,SHA and source bytes; outer JSON LF. Record every used
  slot and any blocked stage,never fill missing results with success.
- [ ] Update current-status.md,repository-recovery.md,shadow-runtime-readiness.md
  and plan outcome. Explicit staging,commit/push existing branch,verify remote SHA
  equals HEAD and clean worktree. Preserve all roots; stop without new experiment.

## Self-review and handoff

Task1 covers deadline/output protocol/owned cleanup;Task2 covers hashing and path
safety;Task3 selected attribution and bounded execution;Task4 portable recovery.
Synthetic reparse tests are not a real-junction acceptance claim; live adapter must
fail closed on unreadable attributes. Build costs/cache warmth and omitted graph
loading remain explicit. No new runtime action this planning turn.
PLAN_READY_FOR_REVIEW; await approval,then execute inline within these budgets.

## Execution outcome — 2026-11-06

Approved after72bae466. Task1 complete (commit7a8ef1a3): OwnerRed expected missing
implementation; OwnerGreen10/10 with2 harmless owned children,exits confirmed.
Task2 partial: vendor/source/license byte hashes verified; GoRed03130d compiled
and failed16 unimplemented assertions as expected (46991.1073ms).
GoGreen72f676 failed TestMonotonicPhaseEvidence after2228.9569ms,exit confirmed,
no timeout. Other15 groups did not report failure. The assertion does not identify
which END had nonpositive duration; timer resolution remains unproven.
STOP applied: no repair/retry,Build/Live NOT_RUN,all owner/Go test slots consumed.
Four module files unchanged. Task4 unfinished-checkpoint archive/recovery updates
prepared. Archive final source only; GoRed original stub hashes retained in claim.
Additive Candidate.SumFile/Event.Expected support expected digest reads inside
PREPARE/comparison evidence; no broader CLI/cache allowlist. Author-only review,
synthetic reparse test limitation;44/34/6/217,NOT_ACCEPTED. Fresh approval required
for further runtime work. Historical checkboxes above are the original plan.

## Resumed build outcome — 2026-11-07

Timing repair e307084c passed16 groups; user continued to Task3. Current sources
matched GREEN receipt; source checkpoint verified remote. Original Build slot
consumed:56c195 exit1 in24015.731ms,VCS status exit128,no timeout,exit confirmed.
Read-only Git ownership rejection explains stamping failure; no helper binary.
Live NOT_RUN. Stop applied,no retry/security changes. See
docs/roadmap/2026-11-07-c12-phase-build-result.md. Separate archive preserves
build receipt without overwriting previous test evidence. Next command amendment
(-buildvcs=false plus new exclusive build root/budget) is proposed,not executed.

## Approved v2 amendment outcome — 2026-11-07

User approved afterf97a7cf4; source0b73dedd uses new v2 workspace and
-buildvcs=false only. Build2f7cc3 passed23488.7821ms. Live64f41f stopped487.077ms
with log-file sharing conflict in owner ReadAllBytes; exit confirmed,no timeout.
Text PREPARE incomplete; all remaining phases NotRun. Both slots exhausted,
no correction/retry. See2026-11-07-c12-phase-v2-result.md and separate portable
v2 archive. Prior checklists/outcomes retained historically;NOT_ACCEPTED.

## Corrected-owner v3 completion — 2026-11-07

Offline sharing fixbbf0ea4a passed10 checks. User continued to source/binary
correspondence review and one bounded live diagnostic. Go sources/binary unchanged
from v2 build; corrected owner independently matched GREEN snapshot/receipt.
Separate v3 claim preserves both provenances without rewriting old manifest.
3ee142:exit0,56717.0329ms,12 complete events,all6 phases OK,4 digest matches.
4 module hashes unchanged. Read-only8dcd14 verified raw/receipt/source/go.sum.
Task3 selected phase evidence and Task4 portable handoff now obtained; this does
not complete authoritative root/C12 acceptance. Detailed timings/caveats in
docs/roadmap/2026-11-07-c12-phase-v3-result.md;archive separate v3. No extra
build/test/retry,old failures retained,44/34/6/217,NOT_ACCEPTED.
