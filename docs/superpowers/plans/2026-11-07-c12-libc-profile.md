# Fixed Libc Profile Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline task-by-task,with the existing approved author-review boundary. Await this plan's approval before implementation or runtime.

**Goal:** Add a backward-compatible fixed libc-v1 profile and obtain one bounded selected-module diagnostic result if offline gates pass.

**Architecture:** Keep the paired-v1 default;explicitly resolve fixed identities in Go and PowerShell. Thread the profile through live/final parsing and bind fresh sources,binary and receipts. No generalized module input.

**Tech Stack:** Windows PowerShell/.NET,pinned Go1.26.5,existing standard-library helper and vendored dirhash.

**Spec:** docs/superpowers/specs/2026-11-07-c12-libc-profile-design.md (approved after d26780f4).

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator;
  branch codex/c12-b01-task9-coordinator;PowerShell only;inline author review.
- Source directory (S): docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/cache-phase.
  New workspace (W): .superpowers/sdd/2026-11-07-c12-libc-profile.
  Fresh exclusive stage roots only;never reuse old test/live roots or delete evidence.
- Profiles exact/case-sensitive:paired-v1(default)=[x/text v0.40.0,sqlite v1.49.1];
  libc-v1=[modernc.org/libc v1.72.1]. Fully qualified first two paths are
  golang.org/x/text and modernc.org/sqlite. Empty/unknown/case variants reject.
- Preserve Schema1,three phases per candidate,START/END ordering,session/sequence,
  output caps,path/reparse checks,zero-duration allowance and owned exit confirmation.
- Do not modify phase.go,dirhash,common helpers,old invoke/test scripts or archives.
  No arbitrary module/version argument,cache download/repair/delete,security changes,
  root verify,project/devtools suite,Docker,CIM,trace or main merge. This is a local
  diagnostic test suite,not broader acceptance.44/34/6/217,NOT_ACCEPTED unchanged.
- Pinned Go executable C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe;
  SHA256D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.
- Starting owner SHA25633782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E;
  common SHA25677824603B0817D32B447C50CA3818F5E34E1924D8403A800073165C65081566F.
  Preserve initial owner snapshot with sibling common file under W/Bootstrap for
  the unchanged outer bounded launcher;record hashes before import. New owner is
  tested as the inner subject,not used to supervise its own RED/GREEN outer batch.
- Child environment GOENV=off,GOWORK=off,GOTOOLCHAIN=local,GOPROXY=off,GOSUMDB=off,
  GOAUTH=off,GOVCS=all:off,GOFLAGS=-mod=readonly,GOMAXPROCS=2,CGO_ENABLED=0.
  Offline/build GOPATH,GOMODCACHE,GOCACHE,GOTMPDIR,TEMP,TMP under W/isolated/<name>.
  Live uses GOPATH=C:/Users/Lenovo/go,GOMODCACHE=C:/Users/Lenovo/go/pkg/mod;
  remaining scratch stays isolated. No persistent settings or tool installation.
- Budget:OwnerRed60s,OwnerGreen60s,GoRed180s,GoGreen180s,Build180s,Live120s;
  exactly one invocation per stage at most. Go tests have30s internal timeout.
  Four synthetic owned children only in OwnerGreen:normal5s,protocol-stop5s,
  partial-cutoff3s,native-raw5s;none in OwnerRed. Existing bounded cleanup applies.
  Expected RED is the gate to implement;unexpected RED/infrastructure failure or
  any GREEN/build/live failure stops subsequent stages. No repair/retry budget.

## Review Focus

1. Default profile regression:omitted and explicit paired produce identical lists/streams.
2. Invalid profile causes side effects:reject before filesystem access/child launch.
3. Parser selection lost on final drain:libc owned completion must pass both reads.
4. Partial/extra/mismatched stream falsely completes:explicit negative fixture cases.
5. Stale binary/source set accepted:complete manifest key-set and hash equality required.
Tasks1–2 cover first four with synthetic tests;Task3 checks source binding before
any real run. Real cache contents remain uninspected until Live prerequisites.

## Task1 — Fixed Go selection and diagnostic-local test suite

**Files:** Create S/profiles.go,S/profiles_test.go;modify S/main.go.
Create S/invoke-libc.ps1 for all stage recipes and fresh immutable evidence.
**Interfaces:** ProfileIDs(name string)([][2]string,error);
ParseProfileArgs(args []string)(profile,session,repo string,err error),using a private
flag.FlagSet with default paired-v1;no filesystem/process operations. main calls
this parser before constructing cache candidates or invoking Run.

- [ ] Verify clean branch/pins/paths/approved remote checkpoint;decode portable
  root receipt baseline and hash-check it. Snapshot bootstrap owner/common. Record
  W/progress.md and source hashes. Confirm no Go/fixture launched during setup.
- [ ] Write tests first. A minimal pre-implementation compatibility stub may expose
  the new API while retaining only legacy behavior so RED is an assertion failure,
  not an undefined-symbol compile failure. Mark and retain this RED snapshot.
  TestProfileDefaults:omitted equals explicit paired IDs/order;TestProfileLibc:
  exact one libc pair;TestProfileReject:empty,unknown,PAIRED-V1,LIBC-V1 fail;
  TestProfileCopies:mutating one result cannot affect a later result;
  TestProfileCLI:parser rejects unknown/positional args and invalid profile with
  deliberately nonexistent repo path without reading it. Existing repo validation
  remains in main;main must invoke selection before any cache path query.
- [ ] Run GoRed once through bootstrap owner in S with
  `go test -count=1 -timeout=30s ./...`,180s cap,isolated caches. Expected:owned exit
  confirmed,nonzero with profile assertion failures,not timeout/compile/environment
  error;retain raw output and source snapshot. Existing phase tests remain present.
- [ ] Implement exact fixed resolvers and parse-before-access;refactor main only
  for option plumbing,retain fixed approved repo/cache and generic Run interface.
  Format touched Go files. No changes to hashing or safety validation semantics.
- [ ] Run GoGreen once with same command/caps. Expected:all diagnostic package
  tests including existing16 top-level phase tests and new profile tests pass,
  exit0,no timeout/owner failure,four module hashes unchanged. Save named results.
- [ ] Commit necessary Go tests/source and RED/GREEN evidence checkpoint;do not
  claim libc runtime or whole project acceptance. Source set still incomplete for
  final build until Task2 and launcher/manifest implementation are finished.

## Task2 — Profile-bound PowerShell parsing and owned integration

**Files:** Modify S/owner.ps1;create S/profile.tests.ps1,S/profile-fixture.ps1.
**Interfaces:** Get-C12PhaseProfile([string]$Profile='paired-v1') returns fixed
ordered Module/Version records;Read-C12PhaseEvents adds trailing Profile default;
Invoke-C12PhaseOwned adds trailing Profile default. Unknown profile throws PROFILE
before reading/launch;nondefault with empty Session throws PROFILE_SESSION.

- [ ] Write profile.tests.ps1 first with Phase OwnerRed/OwnerGreen and fresh root
  supplied by invoke-libc.ps1. Build fixed synthetic paired12/libc6 event arrays.
  First RED case calls existing parser on libc fixture with -Profile libc-v1;
  require either unsupported profile parameter or old identity rejection to prove
  missing behavior. Do not treat arbitrary fixture/import errors as expected RED.
  No fixture children in RED;retain original owner snapshot.
- [ ] Run OwnerRed once,60s outer bootstrap owner. Expected:documented missing-
  profile behavior,confirmed exit and no infrastructure failure. Then implement
  resolver,early rejection and Profile propagation to every live/final parser read.
  Derive terminal phase count from selected list;do not infer profile from events.
- [ ] OwnerGreen once,60s,executes all static cases:default/explicit paired12,
  libc6,exact resolver IDs and fresh returned data;cross-profile both directions;
  wrong version/session/sequence/phase;extra terminal event;START without END;
  partial line then completed line;ERROR/MISMATCH stop;missing digest;event/total
  caps;open writer read/append and denied second writer;occupied record unchanged.
  Unknown/empty/case variant profile rejects against nonexistent stream path and
  nonexistent executable,proving profile rejection precedes I/O/launch.
  Nondefault+empty Session must reject before starting any child.
- [ ] In that same GREEN batch run exactly four fixture children with pinned
  current PowerShell executable:libc complete stream,wrong-profile stream then wait,
  libc START then wait for3s cutoff,native raw text with Session/Profile omitted.
  Assert exit confirmed in all;completion6 events/3 phases and identical raw bytes;
  protocol error stops;cutoff remains Incomplete/NotRun;raw mode remains unparsed.
  Fixture scripts touch only supplied synthetic input/output,never module cache.
- [ ] Store fixture input and original receipts;quota filler is synthetic only
  and kept outside the bounded receipt root. Preserve equivalent old owner/stream
  regression cases within this batch without rerunning historical occupied roots.
  Any failure stops Build/Live. Commit reviewed source/tests and offline evidence.

## Task3 — Fresh artifact binding and single libc run

**Files:** Complete S/invoke-libc.ps1;W/Build and W/Live artifacts (ignored).
**Interfaces:** invoke-libc.ps1 -Stage {GoRed,GoGreen,OwnerRed,OwnerGreen,Build,Live};
all stages create immutable claim/result and check before/after module hashes.
Build/Live require recorded GREEN evidence bound to tested Go/owner source.

- [ ] Before Build,freeze full file-set manifest under S (all source/test/recipe,
  module,provenance and license files),plus common and approved spec/plan hashes.
  Key-set equality detects additions/removals;verify each SHA. Bind separate RED/
  GREEN snapshots accurately;never rewrite an earlier manifest to match later code.
  Inspect manifest comparison with in-memory synthetic equal/changed/added/removed
  dictionaries (no child/cache work);require only equal case to pass.
- [ ] Commit/push finalized necessary sources before Build;check remote==HEAD.
  Build once with pinned Go in S: `go build -buildvcs=false -trimpath -o <W/Build/helper.exe> .`,
  180s cap. Expected:exit0,no failure/timeout,unchanged module hashes. Record binary
  SHA and exact source set;do not use old helper binary. Build failure ends runtime.
- [ ] Live metadata prerequisites:exact libc directory/ZIP/ziphash exist under
  fixed cache;check ancestors/reparse and pinned source/binary/tool hashes. Read
  bounded root go.sum and ziphash with existing safety semantics;require matching
  libc v1.72.1 digest. No recursive cache inventory before helper PREPARE. Missing
  input stops without download/repair;do not substitute empty cache.
- [ ] Claim Live with Profile=libc-v1,session GUID,all hashes and environment.
  Invoke new owner once with -Session <session> -Profile libc-v1,120000ms;
  helper args `-profile libc-v1 -session <session> -repo <approvedRepo>`.
  Expected:finite result,not guaranteed success. Poll under60s updates,no workload
  alongside it. Capture raw output and partial records even on failure/cutoff.
- [ ] Pass only Started/ExitConfirmed,true;exit0,no timeout/failure;6 complete
  events/3 OK phases,correct session/profile/identity,ZIP and directory digests both
  equal root go.sum/ziphash expected digest,empty stderr,four module hashes unchanged.
  Otherwise STOP/NOT_PASSED with exact failed gates;no retries. Report PREPARE/ZIP/
  DIRECTORY durations only when their END exists;missing END is incomplete,not0.

## Task4 — Portable checkpoint and bounded handoff

**Files:** Create docs/roadmap/2026-11-07-c12-libc-profile-result.md and
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-c12-libc-profile.json;
update current-status.md,repository-recovery.md,this plan outcome and source provenance.

- [ ] Package exact source/test snapshots,recipes,claims,original bounded receipts,
  raw output and manifests;review privacy first. Exclude binaries,caches,credentials
  and real module contents. Any omitted private evidence gets explicit length/hash
  and omission reason;do not silently redact originals or claim full portability.
- [ ] Decode every entry and compare length/SHA to originals;record archive hash.
  Distinguish diagnostic-local test success,build success and libc integrity from
  C12 acceptance/prior timeout attribution. Keep44/34/6/217,NOT_ACCEPTED unchanged.
- [ ] Inline author-review checks all five focus items and spec coverage. No new
  independent review claim. Review explicit changed files,require empty index,
  diff --check,stage exact list,commit/push existing branch;confirm remote==HEAD
  and clean worktree. Preserve all local evidence;no automatic follow-on runtime.

## Self-review / handoff

Task1 Go profile contracts and Task2 owner contracts converge only at Task3 live;
profile names/order/default match,source snapshots distinguish stages,and failures
stop later work. No occupied historical roots reused. Task4 preserves every outcome.
Plan approval authorizes these finite stages only. PLAN_READY_FOR_REVIEW;
no source changes or validation runs while planning.

## Execution outcome — stopped before Build

Approved aftere7e2a9b9. Task1 GoRed08a212 expected TestProfileLibc failure;
GoGreen5075e6 diagnostic-local suite passed. Task2 OwnerRed7e8499 expected ORDER;
OwnerGreen374e41 passed18 cases/four synthetic owned exits. Four module files unchanged.
Task3 manifest comparison reached receipt publication d16513 then TOTAL_LIMIT:
receipt destination incorrectly includes isolated caches under whole plan root.
No durable manifest success;Build/Live NOT_RUN confirmed854810. No retries or cap
changes. Task4 preserves partial implementation and failure for recovery,not acceptance.
See2026-11-07-c12-libc-profile-result.md. New correction/resumption scope required.

## Scoped publication correction

User subsequently approved receipt-directory fix/offline recheck only.
RED0871e6 reproduced TOTAL_LIMIT;GREENe50889 saved/read5 manifest cases in a fresh
ManifestCheck directory and verified duplicate rejection. Original4MiB cap and
old evidence preserved. See2026-11-07-c12-manifest-publication-fix.md.
Build/Live still NOT_RUN;resumption requires approval,no Go/owner test retries.
