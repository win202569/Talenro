# Shadow audit runtime readiness — read-only review

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
