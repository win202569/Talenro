# Shadow audit runtime readiness — read-only review

## Approved sampled root attempt after1e1b75cb — timed out

Single c731c2/session52726→a1d7c1 exit1. Same pinned Go/offline/GOMAXPROCS=2
parameters and120s limit, with approximately5s CPU/memory samples. Root elapsed
120038.3429ms,TimedOut=true,ExitConfirmed=true,ExitCode=-1; no launcher exception,
empty stdout/stderr. Owned process terminated at deadline; devtools NOT_RUN.
Twenty-four samples: CPU increased from2875ms at~5s to13828.125ms at~120s;
working set18325504–20164608 bytes. Process was active, not continuously CPU-idle.
CPU time was much less than wall time; this alone does not distinguish I/O/lock
wait, scheduler contention, scanning overhead or progress through particular files.
No disk-I/O counters, stacks or per-module progress were captured. Root cause
remains unknown; no integrity success/corruption/deadlock conclusion.

Four go.mod/go.sum hashes unchanged. Child-only environment, no persistent
configuration/security changes, no compilation/C12/containers/downloads. No
automatic retry or extension. Cache-wide mutation was not measured.
Portable27files21821bytes decode/source-verified61c914:
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-c12-cache-sampled-stop-v1.json
SHA2560034DD28AFE0A3BF80CDB95649939F5FBB4226FEAD691F4C366E0A7C762A65C7.
Contains claim,24 samples,result and exact launcher. Earlier failure preserved.
Next should design a bounded observation that can separate I/O from lock waiting
before seeking another execution approval; do not simply repeat the same probe,
raise concurrency/deadline or alter security controls.44/34/6/217 unchanged,
NOT_ACCEPTED; no main merge.

## Read-only timeout investigation aftera78543a7

No verify rerun. Read saved result and installed Go1.26.5 source
src/cmd/go/internal/modcmd/verify.go (6b5c13). The command loads a module graph,
checks archive and extracted directory hashes, limits concurrent module checks by
GOMAXPROCS, and normally prints success only after the batch finishes. Empty
stdout is therefore not proof of a hang. Previous launcher used GOMAXPROCS=2;
16 logical processors observed now, but more parallelism is not a demonstrated fix.

Read-only ZIP central-directory inventory2a7152 (no extraction/content hashing):
99 declared root modules,22765 entries,608597910 expanded bytes,181059375 zip bytes.
Largest expanded sets: modernc.org/sqlite259376605 bytes; modernc.org/libc86622583;
github.com/klauspost/compress46238299; github.com/pierrec/lz4/v4 36988230;
golang.org/x/text29567911. Metadata read1805.283ms. This does not measure hashing
throughput or establish the resolved module graph used by the timed-out process.
Available C: space69346148352 bytes at observation, not historical free space.
No Go process observed. The command's final absent-process query produced exit1;
the metadata inventory itself completed and emitted the numbers above.

Evidence supports a substantial I/O/hash workload as a hypothesis, not a root
cause. Missing process CPU/I/O samples and stack/progress evidence prevent
distinguishing slow active verification from waiting on a filesystem or lock.
Do not infer cache corruption, disk-full, antivirus interference or deadlock.
No security exclusions, parallelism changes, timeout extensions or cache edits.

Recommended next scoped experiment, requiring approval: one new root-only attempt,
same120s/offline/toolchain/GOMAXPROCS=2 parameters, adding bounded CPU/memory and
exit-state samples. No devtools attempt, auto-retry, downloads, compile or C12.
Preserve old failure and new observations separately; stop at120s and confirm
owned-process exit. Sampling may distinguish activity from inactivity but cannot
by itself prove a specific filesystem/lock cause.44/34/6/217 unchanged,NOT_ACCEPTED.

## Approved offline content verification — stopped at deadline

After13a96123 user approved exact cached Go1.26.5 offline verification, one attempt
per root/devtools module,120 seconds each. Launcher47cf88/session24033 ended901b38
exit1: root go mod verify reached120044.7983ms, TimedOut=true, child ExitCode=-1,
stdout/stderr empty. The owned process tree was terminated and exit confirmed.
Devtools NOT_RUN: stopped on first unexpected outcome, no retry or budget increase.
This is neither successful integrity verification nor evidence of corrupt cache.
The specific cause of delay was not established; no active-process metrics captured.

Child-only environment cleared inherited GO*/GIT_* and selected fixed local
toolchain/cache, GOENV/GOWORK off, GOPROXY/GOSUMDB/GOAUTH off, GOVCS all:off,
GOFLAGS=-mod=readonly, GOMAXPROCS=2. User/machine environment unchanged. No build,
test, container operation or network-enabled download. All four root/devtools
go.mod/go.sum hashes unchanged. Cache-wide before/after inventory was not captured;
do not claim the Go command made no internal cache metadata changes.
Pinned Go executableSHA256:
D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.

Portable stop package4files6604bytes, decoded/source-verifiedf18cd8:
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-c12-cache-verify-stop-v1.json
SHA256F4D0EBB81961B79141A0E05F7D5A18715D56E3B0FF4A3D26D787B89E4CA48481.
Contains claim, root outcome, summary and exact launcher; original evidence kept.
Next proposed scope: read-only timeout investigation, without repeating verify,
changing security settings or extending the deadline. Future execution requires
separate scope.44/34/6/217 unchanged,NOT_ACCEPTED; no C12 run or main merge.

## Read-only cache/environment inventory after46a27e54

Filesystem-only inventorya68210/session45270→49050c inspected explicitly declared
requirements, not a resolved build graph:

|Module|Declared entries|Cache directory/mod/info/zip/ziphash present and ziphash metadata matches go.sum|
|---|---:|---:|
|Root|99|99|
|tools/devtools|318|318|

No missing items or ziphash metadata mismatches among these declarations. This
does not hash extracted source/archives or prove transitive graph completeness;
go mod verify, compilation and workload execution were not performed. The read
command emitted a nonterminating Join-Path empty-parent error for root go.mod,
then explicitly assigned go.sum before reading it; root and devtools sum paths
were confirmed present5eac6c. Counts are inventory observations, not a clean
integrity-verification result. No downloads/cache writes/settings changes made.

Presence-only environment checks show GOPATH set in user/process scopes and
matching the expected user go directory; GOTOOLCHAIN/GOWORK/GOENV/GOMODCACHE/
GOFLAGS/GOPROXY/GOSUMDB unset in checked scopes. GIT_PAGER remains present even
with shell profile loading disabled (5eac6c), absent from user/machine environment.
This narrows the cause to inherited/tool-process environment rather than proving
a profile setting; the precise injector remains unidentified. Preserve the gate.
No full environment values or credentials recorded. Worktree clean after checks.

Next proposed scope: exact cached Go1.26.5, offline dependency-content verification
for root and devtools with explicit finite deadlines, no compilation/downloads/
container operations. A dedicated command process may need an explicitly approved
clean environment; do not change user/machine variables or weaken runner checks.
Only after integrity checks should a bounded live C12 plan be proposed. Existing
native containment/reparse/fault acceptance gaps remain.44/34/6/217 unchanged.

## Approved Docker startup follow-up

After514ecd19 user approved starting the installed Docker Desktop and read-only
engine/image queries only. Initial restricted launch40f255/09de14 left no observed
Desktop/backend process or pipe. Normal-desktop launch1bce07 started successfully;
restricted querya358d6 then reported pipe access denied, not engine absence.
Normal-desktop read-only query1fec5a succeeded:

- Engine29.6.2,OSType linux; running containers0 at query time.
- postgres:18.4-alpine3.23,linux/amd64,
  sha256:996d0920e4ff9df1fc19dacb904492f3c1ec0ec1cc338f0ad7123be7731c5f5e.
- redis:8.8.1-alpine3.23,linux/amd64,
  sha256:8096655e437712b07503796fb64d81359256cfcff0ab29d95a7da72863786efb.
- nats:2.14.3-alpine3.22,linux/amd64,
  sha256:c11af972c99ae542de8925e6a7d9c533aa1eb039660420d2074beed6089b3bf0.

No image pull, container creation, installation, setting change or C12 execution.
Desktop remains running as authorized. Engine/image-availability blocker below
is superseded by this observation, not by a full runtime attestation. Image tags
can change later; future runner still must obtain its own identity receipts.
Restricted pipe access remains a separate execution-context limitation. Pinned
toolchain selection, inherited GIT_PAGER gate, dependency-cache completeness and
native containment readiness remain unresolved/unverified. Next: scoped read-only
environment/dependency preparation review, then a separately approved finite live
plan.44/34/6/217 unchanged;NOT_ACCEPTED; no main merge.

Baseline:6ff362e623cbdd8752bfda26ce33df60b6236822. This review does not authorize
or execute a live runner, dependency download, container creation or settings change.

## Observations

|Check|Observed result|Evidence|
|---|---|---|
|Worktree|Clean at review start|b13af4|
|PowerShell|7.6.5|8b8d0f|
|PATH Go|go1.25.5 windows/amd64, not required1.26.5|8b8d0f|
|Pinned Go cache|Exact go1.26.5 windows/amd64 exists and version query succeeds|2d3ef4|
|Docker CLI|Installed in user-provided DockerDesktop resources/bin path|b13af4|
|Docker Linux engine|Version query exit1: named pipe missing; server unavailable|8b8d0f|
|Docker Desktop/backend processes|Neither named process observed|2d3ef4|
|Forbidden inherited names|GIT_PAGER present in this tool process; absent from user/machine scopes|8b8d0f,2d3ef4|
|Source pins|All6 current file hashes match frozen manifest|2d3ef4|

Pinned executable:
`C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe`.
Version queries temporarily selected GOTOOLCHAIN=local in the command process to
prevent automatic toolchain download, then restored its original value. No persistent
configuration changed. No environment values/credentials were recorded.

## Blockers and unverified prerequisites

- Docker engine unavailable; no image/daemon/resource inventory could be established.
  Process absence corroborates unavailability but is not proof of installation damage.
- Current tool-process GIT_PAGER would be rejected by Assert-C12NoInheritedDependencies.
  Its injection source is not determined. Do not remove the runner's environment gate.
- A future dedicated execution process must select the existing pinned Go and satisfy
  the inherited-environment gate; PATH alone currently selects an older installation.
- Dependency cache completeness, required PostgreSQL/Redis/NATS image availability,
  native containment and full environment readiness remain unverified.
- Reparse coverage remains BLOCKED. Existing fault/timeout/independent-review gaps
  remain; offline success is not live acceptance.

## Next scoped action

Seek approval to start the existing Docker Desktop application, then perform only
read-only daemon/image checks. Starting Docker may resume existing workloads under
their configured restart policies; do not treat it as a read-only action. Do not
pull images, install/upgrade tools, create containers or run C12 without separate
scope. Then propose a finite Focused run and explicit cleanup/deadline budget.

No live run budget proposed as ready while these prerequisites remain unknown.
Original offline7/7 plan consumed; no test reruns.44/34/6/217 unchanged,
NOT_ACCEPTED; no main merge. Review documentation is committed for recovery only.
