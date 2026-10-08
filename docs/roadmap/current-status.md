# Project status — updated 2026-10-08

Approved test-only follow-up: fallback metadata now has a valid positive control
and rejection cases for short/extra fields, multiline replies, mismatched path
or version, empty directory and an existing but incorrect directory. The fixture
checks exact offline query arguments and that rejected metadata never reaches
integrity verification. Owned-copy identity-check removal controls demonstrate
that the mismatch tests depend on the intended guards. Frozen-clock cases cover
3599s success and 3600s expiry; 3599s/3601s budget mutations reverse the respective
outcomes. The 16 selected metadata/budget cases PASS (31.444s package duration,
exit 0). Production scripts and all four dependency lock hashes are unchanged.
The user authorized saving this follow-up as an unaccepted development checkpoint
on codex/c12-b01-task9-coordinator, without merging main. Confirm publication
against the remote branch before recovery; this does not clear security/acceptance
gates or authorize a full C11/C12 run.

The complete scoped TestDevtoolsVerificationPowerShell regression also PASSes
with -count=1 and a 4m cap (112.063s package duration, exit 0); its Bash-only
launcher case is intentionally skipped. Independent read-only review found no
blocking issue. Nonblocking coverage note: the multiline case also violates the
field-count guard, so it does not isolate newline-guard sensitivity. This result
is not a full repository, C11 or C12 acceptance run.
Fresh pre-commit repeat of the same scoped group PASSes (102.445s package
duration, exit 0), with the same Bash-only skip and no production-script changes.

Read-only resumption check: local HEAD is published checkpoint fed278f3 and
the worktree was clean before this documentation correction. Defender service,
antivirus and real-time protection report enabled; intelligence version is
1.459.601.0. The recorded Commando.A!ml threat reports IsActive=False and
DidThreatExecute=False. This is current product status, not a false-positive
determination or clearance of the historical bootstrap event. No vendor/security
owner disposition has been supplied; full ordinary/C11/C12 and I2/I3 remain
blocked. No blocked executable/test, security change, upload or rerun occurred.
Recovery instructions now reflect the published metadata/budget fixes rather
than their obsolete local-only state. These doc updates accompany the test-only
checkpoint; earlier local-only statements describe their respective work steps.

Checkpoint publication scope: the user authorized committing and pushing the
eight reviewed source/test/document files to codex/c12-b01-task9-coordinator.
This is an UNACCEPTED DEVELOPMENT CHECKPOINT, not a release or main merge.
Ignored machine-local diagnostics are excluded. Earlier "no commit/push"
statements below describe those individual work steps, not this authorization.
Use the Git remote branch state to confirm publication completion.

Latest approved change: C11 devtools verification now has a 3660s outer cap;
check-tools and generate each have a 4200s total stage cap, including their
own repeated 3600s verifier. All other stage limits, repeated checks, cleanup
and failure propagation remain unchanged. The previous outer/inner timeout
mismatch is addressed in code, not yet accepted by a real full C11 run.
Four scoped regression groups PASS (91.175s, exit 0): C11 stage-budget
forwarding, failure-stops-consumers, same-host dispatch and PowerShell C11
fake-tool contract. The new test failed against the old 900/900/600 values
before implementation and passed afterward; ordinary unit-stage default is
still 600s. No full suite/long verification, commit or push. Security gates
remain blocked; historical shorter-budget statements below are superseded.

Independent read-only checkpoint review (2026-10-01): no blocking production
defect identified; suitable as an unaccepted development checkpoint, not for
overall/security acceptance. Nonblocking follow-ups: direct negative coverage
for malformed/mismatched fallback metadata, and tighter behavioral coverage
of the exact 3600s boundary. No additional runtime tests were run during review.
Eight pending source/test/document files are in scope; local diagnostics remain
ignored. No staging, commit, push or merge performed.

## Historical progress (superseded where noted above)

Earlier scoped regression PASSED: the entire TestDevtoolsVerificationPowerShell
group, fresh run with -count=1 and a 4m cap, package duration 104.482s, exit 0.
Both observer cases, aged-clock budget cases, integrity/cache checks, timeout
and descendant cleanup, cancellation, output caps and environment restoration
passed. Only the inapplicable Git-for-Windows launcher case skipped. This closes
the observed failure for this scoped group, not the full project suite or
real shared-cache acceptance. Four lock hashes remain unchanged; gofmt and
diff checks pass (existing Git LF/CRLF notice only). No implementation change,
long verification, C11/C12 run, commit or push in this step. Next unresolved
integration issue is the unchanged shorter C11 outer-stage budget; security
gates remain in force. Entries below record the earlier sequence.

Post-restart targeted verification PASSED: runtime-5.1-rejected (5.08s) and
ownership-boundary (8.03s), package duration 15.337s, exit 0. Windows reports
a new boot on 2026-10-01. No observer code, assertions, deadlines or security
settings were changed to obtain this result. The observed failure cleared
after restart; this does not establish its underlying cause or prove stability.
Only these two tests were rerun, not the entire verifier regression group or
the long shared-cache verification. Next: consolidate the short verifier
regression group; C11 outer budgets and existing security gates remain blocked.
No commit or push. Earlier failed runs below are retained as historical evidence.

Approved bounded timeout revision implemented: the standalone PowerShell
devtools verifier now shares one fixed 3600s deadline across all commands.
Offline/cache-completeness/integrity checks, output cap and owned-process
cleanup are unchanged. C11 stage limits and the private task3 wrapper remain
unchanged; neither is ready for full acceptance. No long verification rerun,
commit or push is included.

Short regression: both new aged-clock cases PASS, along with integrity,
timeout/descendant cleanup, output-cap and environment-restoration cases.
The `TestDevtoolsVerificationPowerShell` group still FAILS (100.130s, exit 1):
`runtime-5.1-rejected` reports `PositiveControl:false`, and `ownership-boundary`
reports `process-event positive control missing`. Both reproduce in an isolated
rerun (15.925s, exit 1). They fail at CIM process-event observer self-checks;
the underlying reason is not established. No assertion was relaxed. Four lock
hashes remain unchanged; formatting and diff checks pass (existing Git line-
ending warning only). Next: diagnose these observer failures before claiming
regression acceptance; full tests and the existing security gates stay blocked.

Observer diagnosis (2026-09-30): independent probes reproduce zero
Win32_ProcessStartTrace events in both PowerShell 7.6.5 and Windows PowerShell
5.1, even with a diagnostic-only 10s observation window. The effective token
is administrator, the child parent PID matches the query, a local event queue
control passes, and Winmgmt is running. A simultaneous filtered intrinsic WMI
process-creation subscription observes the exact same owned child (1 match),
while the trace subscription observes none. The fault boundary is narrowed to
the process-start trace event path, not general WMI delivery or the revised
verifier deadline; the underlying provider/trace cause is not yet established.
No polling substitute, test relaxation, service restart or security change.

Provider/session checks: exact COM registration resolves to signed-valid
krnlprov.dll; WMI logs report provider startup result 0x0. During the final
bounded probe, Admin_PS_Provider is running with Microsoft-Windows-Kernel-Process
process/thread/image keywords (0x70), yet the matching trace event still does
not arrive; the intrinsic control still does. This does not identify a repairable
repository defect. Stop repeating unchanged probes. Suggested next recovery
boundary: user-controlled Windows restart, then rerun only the two failed
observer tests before any larger acceptance. Restart has not been performed
and is not guaranteed to resolve the fault. No session/service/security changes.

Historical evidence before this revision:

Approved extended diagnostic PASSED at 2026-09-30T11:33:54Z: 507 selected
modules, exit 0, 2114.906s total (about 35m15s), within the separate 3600s cap.
All cache-binding/completeness preconditions and the exact Go success-output
check were retained. Four locks and public verifier source remained unchanged.
This result exceeds 900s and does not satisfy or change the public acceptance
gate, establish cold-cache performance, or clear the security finding.
Log: ignored extended-integrity-2026-09-30.log. No automatic rerun followed.

Historical next-step assessment (superseded by the diagnostic and revision above): both public verifier and diagnostic
wrapper have 900s deadlines; extending the wrapper alone cannot fix the inner
limit. Proposed one-off full cache-verification diagnostic has a separate
3600s cap, preserves Defender/cache-binding/completeness preconditions and
offline fixed-toolchain behavior, and requires approval. Even diagnostic
success would not satisfy or change the public 900s acceptance gate.

Official Defender analyzer corroboration: existing-trace analysis succeeded
and classifies all 60 target scans in the selected window as RealTimeScan /
OnOpen / Not skipped (average 14.375ms). Its returned fields do not separate
CPU from internal waits. Actual opening-triggered real-time scanning is now
confirmed, but internal delay cause and original integrity acceptance remain
unresolved. No recapture, protection change, upload or verification rerun.

Latest existing-trace analysis correlates delayed opens with Defender OnOpen
scan requests: all 60 complete engine request pairs in seconds 10–11 match
the target process, exact cache-file paths and enclosing WdFilter CREATE
intervals. Request duration averaged 14.375ms (0.913–26.496ms). This is direct
scan-path evidence, not proof of active CPU cost, internal wait cause or a
Defender defect. No new capture or protection change; integrity acceptance
remains blocked and the existing security finding is not cleared.

Latest phase-gated trace reproduced content-open slowness after enumeration:
about 30s of hashing completed 2070 files, with 28.345s measured in open.
The recorder is stopped. Target process coverage and zero header loss counters
were verified. In seconds 10–11, all 62 target callbacks exceeding 1ms were
WdFilter.sys CREATE (1.313–28.315ms each). This localizes recorded delays to
that filter path, not an intrinsic driver defect or exclusive cause. No
security setting changed and no integrity pass was obtained; acceptance stays
blocked. Further analysis should use the existing trace, not another capture.

Latest approved trace: one 30-second full-directory run stopped at its deadline
and saved locally (1.72 GB); WPR confirmed stopped. The probe remained in
enumeration throughout, so content-open slowness was not captured. PID/time
coverage and target filter activity during seconds 10–11 were verified with
zero header-reported losses, but no causal driver attribution is established.
No automatic recapture, security change or acceptance clearance followed.

Newest integrity diagnostic: a progress-instrumented full-directory hash
reproduced slow opens. Enumeration found 127449 files in 11.231s; at the 90s
diagnostic deadline only 7405 files were completed, with 75.421s measured in
open versus 1.633s read and 1.307s close. Exit 124, no digest/acceptance pass.
An initial read-counter defect was corrected and self-tested; its invalid
log is retained. Instrumentation affects timing and the OS cause is unproven.
No new system trace, public-verifier rerun or security change occurred.

Newest trace result: analysis limited to the first 1.5s succeeded in 23.043s
(exit 0), identifying file-filter I/O activity for tca0a383cf.exe. Together
with the PID/time match this confirms target process and I/O presence, not
exhaustive accounting of all 2048 reads. The sample was fast; filter totals
are not a root-cause attribution and must not be summed as wall-clock costs.
No additional capture is needed just to establish target activity. Original
cache-verification timeout and acceptance blockers remain unresolved.

Latest diagnostic: the approved one-shot file-mode recapture saved locally
and WPR stopped. The 2048-file sample took 0.403s (slow condition absent).
Independent xperf queries confirm PID 16480 / tca0a383cf.exe is in this new
trace, whose header reports zero lost events/buffers. Combined filter analysis
timed out at 60s with no output; attributable file-operation coverage remains
unverified. No second capture, security change or acceptance clearance followed.

Newest result: the complete development-tool selector passed in 471.162s
(1041 PASS markers, zero FAIL, one existing inapplicable Skip), with normal
timing disabled and failure diagnostics retained. Earlier intermittent
failures below remain historical evidence, not a proven root-cause fix.
Real-cache integrity verification advanced to `go mod verify`, but its
unchanged 900-second wrapper budget expired (exit 124 / 900.206s). No integrity
pass was obtained; check-tools, generation and lint did not follow. Changes
remain local and uncommitted, with Task 3 acceptance blocked on this result.

Bounded follow-up localized a performance candidate: doubleclick v1.0.0
contains 127,449 of the selected archives' 199,115 entries. Its Go archive
hash matched in 2.123s, while directory hashing exceeded a 90-second probe;
directory enumeration alone took 10.152s in a separate run. Per-file tree
access warrants further tracing; no underlying OS cause or full integrity
pass is established, and no verification requirement was relaxed.

A distributed 2048-file sample now pinpoints the dominant measured operation:
19.852s in file open out of 20.737s wall time (about 96%), versus 0.470s
read/copy and 0.001528s hashing. The underlying Windows/storage/filter cause
remains unproven. Next investigation should target file opens; full integrity
acceptance remains blocked and no security setting or production budget changed.

The user-approved approximately 30-second system trace was saved locally and
WPR stopped; the raw 963.6 MB ETL is Git-ignored and not uploaded. The sampled
opens were fast in this recording (0.338s total versus the earlier 20.737s),
so the slow condition was not captured. Local summary conversion hit its
90-second cap; driver attribution and event-loss statistics remain unverified.

Targeted follow-up could read the first 200 ETL events, but the sample-PID
query also hit its 90-second cap without usable results. No new capture,
tool installation or security change followed. Detailed trace analysis is
blocked with the attempted built-in readers; the original local-only ETL is
preserved. Another analyzer would require a separate decision and cannot
retroactively reproduce the slow condition absent from this capture.

The user subsequently approved installing only Windows Performance Toolkit.
Microsoft-signed ADK 10.1.26100.9457 setup returned 0 without restart;
xperf and WPAExporter help commands ran. WPAExporter warned of a missing
deps.json file, so actual analysis remains unverified. No new recording,
upload or security change occurred. Next step is a local xperf trace summary;
this does not clear the existing integrity-verification acceptance blocker.

xperf analysis subsequently succeeded: the ETL header spans 52.605s and
reports zero lost events/buffers. Filter/process aggregation completed in
33.483s, but neither it nor a separate process report identified the sampled
program (PID 35860). Workload coverage is therefore unverified despite zero
header loss counters. This recording cannot attribute the earlier slow opens;
no new capture or security change followed, and acceptance remains blocked.

Latest local follow-up: the approved empty-cache-metadata repair and its real
cache regression cases pass, with all four lockfiles unchanged. The covering
development-tool run failed at its unchanged 10-minute deadline and had two
Bash marker-control timeouts; acceptance remains open. These changes are not
committed or pushed. See [metadata repair evidence](2026-09-28-devtools-metadata-fix.md).

Follow-up instrumented regression completed in 459.228s but still failed:
the original Bash timeouts did not reproduce; environment-restore-success
instead hit its test-only two-second deadline. Production budgets were not
changed. The failure cause and overall acceptance remain open.

The subsequently approved success-fixture budget adjustment passed targeted
checks and the PowerShell verifier group in a fresh non-instrumented run.
Production and timeout-injection limits are unchanged. The covering run
still exited 1 in 465.502s: one Git Bash outside-cwd entry timed out. Overall
acceptance remains blocked; changes remain local, uncommitted and unpushed.

Latest focused diagnosis: ten bounded samples of the failing Bash case and
one complete Bash entry group passed (4.639s / 62.844s). Failure-only launch
timing is now retained even with normal diagnostics off. The intermittent
timeout was not reproduced; the last combined failure is not cleared.

The original snapshot was checked against the implementation and plans at
`2668b901`, followed by the repository-portability repair. This branch also
records the Task 9 core checkpoint below. Implemented components are distinct
from the acceptance criteria of their containing stage.

| Stage | Status |
| --- | --- |
| Foundation | Completed development acceptance; see the implementation sequence. |
| C1.1 account, device identity, configuration trust | Completed development acceptance. The delivered bundle is a test configuration, not a production tunnel configuration. |
| C1.2 node and POP control plane | Partially implemented within Batch 01. Contracts, schema, authority coordination, migration guards, and test infrastructure exist; full Batch 01 acceptance has not closed. |
| C1.3 entitlements, traffic ledger, quota | Planned; business implementation pending. |
| C1.4 scheduling, tunnel credentials, production distribution | Planned; implementation pending. |
| C1.5 five-POP composed fault acceptance | Pending the preceding stages. |
| Four-platform Tunnel Engine | Designed; native client implementation pending. |
| Payments, subscriptions, refunds | Planned; implementation pending. |
| Referral, affiliate, rewarded ads | Designed; implementation pending. |
| Commercial UI, localization, packaging, stores | Designed; implementation and release acceptance pending. |

The merged **C12 Secure Execution Architecture** concerns the Windows test
execution controller. Its completion does not mean C1.2 or the VPN product is
complete. The repository currently has no node-agent or node-core-supervisor
command, production node API implementation, or Xray/sing-box runtime adapter.

## Repository portability repair

Commit `91e47661944357f5dd5c6499036b0fc549629ca5` integrates the former local
repository compilation fix and removes pre-existing overlay dependencies from
tests. Authority, store, and migration package tests pass without the overlay.
The sealed factory probe also passes with race instrumentation enabled. An
independent review found no Critical or Important issues in this repair.

A clean clone exposed two further portability issues: CRLF checkout changed the
pinned embedded Go-source digest, and whole-second watchdog rounding discarded
almost one second of the native operation budget. The follow-up normalizes only
CRLF to LF before source verification and uses remaining milliseconds without
extending production deadlines. Regression coverage retains source-tamper
rejection, native descendant termination, and exact resource cleanup. This
follow-up is committed as `16c4e8c4e43ae893aad8779639bec0f80ca096cb`; its
independent review is closed with no remaining blockers.

On 2026-09-08, a new independent clone of that commit, in a path containing
spaces and with no pre-existing `.superpowers` directory, passed:

```text
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOFLAGS="" GOENV=off GOWORK=off
go test ./... -count=1 -p=1 -timeout=60m
```

The command exited 0; `internal/testinfra` completed in 663.533 seconds. This is
the Windows untagged repository suite, not live Docker/PITR or production
acceptance. The code checked here is reproducible without the old local source
overlay; dependencies and tools must still be installed on a new computer.

## Next implementation work

1. Close the remaining [Batch 01](../superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md)
   implementation and acceptance gaps in dependency order: Task 9 Coordinator
   integration, Task 10 guarded serving, Tasks 11–14 complete protocol contracts,
   then Tasks 15–18 exact-epoch recovery and final acceptance.
2. Proceed through the remaining [C1.2 batches](../superpowers/plans/2026-08-23-node-pop-control-plane.md),
   including inventory/operator state, node identity and mTLS, agent, supervisor,
   observations, core adapters, and final operational acceptance.

Clean-checkout reproducibility is now verified as described above; follow
[repository recovery](../runbooks/repository-recovery.md) on another computer.

The [implementation sequence](implementation-sequence.md) preserves historical
acceptance records. Its C1.2 `current design` label predates the current partial
implementation; it is not a declaration that C1.2 has passed final acceptance.

## First Batch 01 delivery unit

Task 9 is in progress on `codex/c12-b01-task9-coordinator`. Core checkpoints
`e0a387f3` and `1ace2bf7` replace the old arbitrary resolver constructor with the
exact sealed dispatcher, commit a durable claim before provider Abort, and connect
activation and stored-outcome recovery to the canonical evidence APIs. This is
not complete Task 9 or Batch 01 acceptance, and is not a production serving gate.

The checkpoint includes three independently reviewed repository corrections:
`1bbe83a9` fixes the checkpoint/time material branches, `e1e9c8a1` enforces their
reason-specific matrix, and `8e8a34d2` requires domain absence for an aborted
fence while retaining the domain-outcome requirement for a committed fence.

An independent clean linked worktree at `1ace2bf7`, containing the final ordinary
Coordinator/readiness code, passed:

```text
go test ./internal/nodecontrol/authority ./internal/readiness -count=1 -timeout=10m
go test -race ./internal/nodecontrol/authority -run 'TestCoordinator|TestActivationAdmission' -count=1 -timeout=5m
```

The complete authority/readiness packages finished in 12.725s / 2.954s; the
specified race check finished in 2.549s. Both commands exited 0. Ordinary commands
used Windows/amd64, CGO=0, empty GOFLAGS and GOENV/GOWORK off; the race check used
CGO=1 and the existing MinGW GCC. Independent core protocol review is closed
with no remaining Critical, Important, or Minor issues. Recovery mutation tests
now compare independent domain inputs and terminal context, and distinguish a
failed Commit before persistence from a lost response after persistence.

One full ordinary repository run also passed without changing its timeout:

```text
go test ./... -count=1 -timeout=10m
```

It exited 0, including authority 13.695s, readiness 3.522s, store 5.162s, and
testinfra 541.302s. Subsequent fixture-only commits do not change this untagged
code. This is not execution evidence for PostgreSQL or physical PITR tests.

Claim-v1 PostgreSQL crash/atomicity and two-connection fixtures are committed
at `cdb9697d`, with a timestamp-precision correction at `cc88e916`. They are
statically reviewed, not live database accepted. An independent
integration-tag compilation check at `cc88e916` exited 0 for authority/readiness
(0.338s / 1.890s, no tests executed), using:

```text
go test -tags=integration ./internal/nodecontrol/authority ./internal/readiness -run '^$' -count=1 -timeout=5m
```

The earlier whole-branch review found no Critical issues and one Important
legacy PITR integration regression, described below. The controlled-interface
correction and final decoder fix now have independent source-review approval and
fresh ordinary verification, as recorded below. The branch remains an explicitly
unaccepted development checkpoint, not ready to merge into main: physical PITR
acceptance and the historical validation finding remain open.
Docker is unavailable on this host, and the repository has no existing GitHub
workflow to provide a replacement live run. The user approved the
[controlled PITR interface addendum](../superpowers/specs/2026-09-19-c12-controlled-pitr-test-interface-design.md)
on 2026-09-19. Its [six-task implementation plan](../superpowers/plans/2026-09-19-c12-controlled-pitr-test-interface.md)
was subsequently approved for subagent-driven execution with “是” on 2026-09-19.
Tasks1–5 are implemented through `e3b5ed93726dd0b7d02d8a609ae24963e48087fb`;
Task6 registration and portable evidence are included in the commit introducing
the execution ledger below. Cleanup remains runner-owned. The plan now records
actual commands/results, source hashes, ordered root rulings R1–22 and explicit
user exception U1; its original planned gates are not execution evidence.

The preceding review's Important legacy integration finding was that
`TestPITRBeforeRevocationFailsClosed` supplied the in-memory
`coordinatorEffectResolver` to the new sealed test dispatcher while using real
PostgreSQL transactions. That source implemented neither the registered
transactional resolver nor activator; the handler rejected this DBTX path.
The legacy schema could also fail before that point. Task5 replaced that harness
with real PostgresRepository/registered-handler/controller P/A/B transactions,
preserved the top-level name, and received independent source review approval
for a development checkpoint. Task6 closes exact PITR-only runner registration.
This is source remediation now covered by final whole-branch review and its scoped
fix rereview, not a claim that the unavailable physical gate passed or that the
checkpoint has received acceptance approval.

The existing authority-v7 runner was attempted without changing its guards or
budgets. Initial launches rejected inherited `GIT_*` variables. A separate child
process with only those inherited keys removed advanced past that guard, then
exited 1 with `Go graph go mod verify timed out after its bounded native wait`.
No database test ran. Docker's absence is a separate known environment limitation,
not the reported termination point of this attempt.

Existing-fence/lost-claim recovery is in scope. Reconstructing an entirely lost
claim-v1 fence remains fail-closed because the current contracts do not provide
its exact authoritative ProtocolActivationID. The activation/runtime/lease-aware
v7 readiness gate is owned by Task 15 and its staging-state acceptance by Task 18;
ordinary Task 9 readiness must not be used as a substitute.

Task 10 has no `WithConsistentReadyRead`, authorized-certificate projection, or
active desired-state projection yet. The migrated test-local PITR read checks readiness
and reads a fixed test snapshot; it does not provide the required same-connection
readiness/query/readiness contract for those real projections.

Later tasks should reuse the existing private canonical/envelope/evidence
helpers, while completing the public schema/role registry and exact-epoch
projection. The current ordinary checkpoint query still selects
`MAX(authority_epoch)`; that is not the planned activation/terminal-chain
projection. Finally, `-Suite batch01` requires the tracked
`testdata/c12/integration-contracts-schema-authority.v1.json`, which is absent.
The runner currently rejects that missing manifest before starting test groups;
the manifest must be derived from the completed required integration cases.

## Controlled PITR implementation checkpoint — 2026-09-19

Branch: `codex/c12-b01-task9-coordinator`, isolated worktree
`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`. Plan implementation base:
`d8b8e01483ae5e7fffa5dfb76b4f847ab95740de`. Task source checkpoints are Task1
`29e03bc8`, Task2 `816fa7fa`, Task3 `374a8098`, Task4 `0cb6c720`, Task5
`e3b5ed93726dd0b7d02d8a609ae24963e48087fb`, and the Task6 commit containing this
ledger (`test(c12): close controlled PITR gates and record evidence`). The Task6
handoff originally awaited independent review; the subsequent whole-branch
findings and bounded fix verification are recorded below. Scoped final rereview
and the authorized development-checkpoint push are complete. GitHub confirmed
`adc3494d9fd6391b2520358d346bc2b7e7de05bd` on the development branch; this concluding
documentation update records that result. No PR, main merge or worktree removal
was performed. On another computer, fetch and switch to
`codex/c12-b01-task9-coordinator`; main does not contain this checkpoint.

The [portable execution record](../superpowers/plans/2026-09-19-c12-controlled-pitr-test-interface.md#execution-record--2026-09-19-development-checkpoint-not-acceptance)
contains the actual approval/execution checklist, exact commands/results, source
hashes, complete ordered root decisions R1–22 with reasons/costs, explicit user U1,
and all 13 separate physical gate commands. It does not require local scratch to
identify what was implemented, tested or left unaccepted.

Task6 actual PowerShell registration RED failed before implementation at the
first resource boundary under the wrong profile; the final exact registration/API
pair passes (2.473s). Exact no-Docker behavior passes (testinfra2.799s,
authority bridge0.318s); integration compile-only passes (0.324s/0.367s, no tests
executed); direct native race passes (6.337s/1.384s, no race report); ordinary
authority/readiness passes (10.336s/2.822s). `-Race` in the physical runner is still
not forwarded and provides no race evidence.

All source/tests were frozen before the one final ordinary command:

```powershell
$env:GOOS='windows'; go test ./... -count=1 -timeout=15m
```

Session81642 exited0; all packages passed, testinfra587.277s. Three changed
source/test hashes matched after completion. Documentation was appended afterward,
not included retroactively in that source freeze. Task5's preceding frozen full
run26567 also passed (testinfra563.456s). Neither success erases prior failures or
establishes their cause: Task4 frozen80961 failed its transient-validator150s
helper and testinfra15m limit (903.139s); unchanged R21 run26986 failed platform
cancellation1.2828222s>1s and prepared oversized-frame CreateProcessW Win325
(testinfra583.287s). The unchanged exact transient-validator rerun passed130.82s,
which is not a full-suite pass or cause diagnosis. The earlier baseline/Task1/Task3
failure and focused-rerun history remains in the linked record.

User U1 explicitly chose “继续第 5–6 项，保留验收阻塞”: continue development and
push only an unaccepted branch checkpoint, without merging main. Task4's Important
validation finding remains OPEN after final review; this is not a general waiver of
new source findings, physical gates or final reporting.

Physical precheck again found no Docker executable; its daemon, endpoint/image
identity, restored candidate behavior and live cleanup could not be checked.
Go/PowerShell/Git are available; optional Go-tool goose lookup was interrupted
without a result, so dependency availability is not claimed. Inherited GIT_* keys
also require the existing legal clean-child launch. No physical runner was invoked,
guard removed, Docker installed or deadline expanded. The seven PITR calls
(two public consumers and five private seam values) plus six preserved authority-v7
crash calls are **unavailable/not executed**, not failed/skipped/passed tests. Each
retains a separate5m budget. Required explicit go-json PASS, foreign-canary/exact
cleanup, real P Ready/A/B terminals, A-present/B-absent restore, timeline advance,
precise B-provider denial and same-candidate A-provider Ready remain unverified.

The legacy authority-owned Docker/raw-pool/migration/cleanup harness was replaced,
not retained as fallback. Its unrelated fresh-cluster physical branch was removed;
the retained `TestAuthorityReadiness` identity-mismatch unit test is **not equivalent
physical coverage**. R17 transparently corrects the original plan's inclusive-on
record-end defect to off/false with PostgreSQL18.4 source citations. R20 retains a
private recovery-safe identity query, without new consumer SQL/grants. Task9/B01
is still unaccepted; Task10 certificate/desired-state serving and reconcile
recovery remain undelivered.

### Whole-branch review and bounded final source fixes

The independent whole-branch review of `e1571f9f..3998d5c9` identified Important
I1 (materialized results retained a shared mutable driver decoder), Important I2
(all13 physical gates unexecuted), Important I3 (the historical Task4 ordinary
validation finding), and Minor M1/M2 (an ineffective uncertain-Commit read
assertion and a raw-driver guard that missed function-value aliases). No Critical
finding or further production Coordinator/repository/readiness defect was found.

The single final fix wave based on `3998d5c9` addresses I1 and both Minors only:
each result keeps its independent fixed pgx decoder and reserves4KiB of the
existing shared64MiB budget before allocation; no Scan acquires backend ownership.
The stable-map regression exercises two legal materialized GetAuthorityFenceHead
rows separately through Access and Transaction, then barrier-starts their first
pg_lsn scans. Pre-fix assertions proved shared retained state and missing decoder
charging; an actual pre-fix race was **not** reproduced. The corrected concurrent
tests pass under the direct native race detector. M1 now requires a non-error
materialized read and exactly one driver-query delta after uncertain Commit.
M2 checks imported selector references against the existing small symbol inventory,
with real-guard alias-negative and approved-symbol-positive controls.

Focused final-fix verification: new regression RED0.382s, guard-alias RED1.693s;
GREEN new regressions0.400s, ordinary API/selector guards1.421s, approved controller
behavior/registry2.793s, direct new-regression race3.830s, broader scoped race4.385s,
pure bridge0.339s/race1.381s, and integration compile-only authority0.343s /
testinfra0.371s. All GREEN commands exited0. The existing codec, NULL/empty-byte,
expiry, eager-close and Commit-boundary tests are included in the covering gates.
Exact commands and decoder-cost rationale are in the linked portable record.

The fix is committed as `b1e1d7c209bd63acd01f3ee00ba06eb765044990`. Independent
scoped rereview verified I1/M1/M2 **ADDRESSED**, with no new Critical, Important,
Minor or out-of-scope source finding. It independently checked the pinned decoder
allocation bound, both Access/Transaction regressions and unchanged Commit ordering.
Source review is closed; acceptance is not approved.

Root then completed the fresh frozen ordinary command shown above on that exact
clean commit: session5472 exited0, all packages passed, including testinfra469.316s,
authority17.181s, readiness4.484s, trust4.228s and trustclient1.304s. The run began
at2026-09-19T18:20:11+04; post-run verification at18:28:56+04 found all seven source/
test/document SHA256s unchanged and the same clean HEAD. No source/test/document
edits or other test runs occurred during it. Subsequent commits update only the
portable evidence. Ordinary81642 remains prior-tree evidence, while5472 covers
the final source fix. Neither run diagnoses the historical failures.

I2 and I3 remain OPEN acceptance
blockers under U1. All13 physical calls remain unavailable/not executed, and the
historical failure causes remain unproved. The review's excluded Task10/15/18,
B03 provider/decoder, lost-fence identity, cross-process restart, hostile same-process
sandboxing, removed fresh-cluster coverage, runner race forwarding and wholesale
runner audit remain outside this delivery; none is claimed implemented or accepted.
## 2026-09-28 development-tool isolation checkpoint

Task 2 is committed locally as `a0ca18928f552752c6580929f9b6fff88cd8d58b` on
`codex/c12-b01-task9-coordinator`; no push or merge is included. The root retains
Goose/runtime dependencies, and five development tools use `tools/devtools`.
Windows development entries require stable PowerShell Core 7.6.5; the four Bash
tool entries reject all platforms. Existing Bash smoke requirements are retained.

The final task-level Windows selector passed in 541.597s under the original10m
budget (1031 PASS markers,0 FAIL,1 existing inapplicable Skip). Earlier Linux
four-entry rejection passed0.916s. These are scoped implementation tests, not
full ordinary/C11/C12 or production acceptance. Historical readiness timeout
571.336s remains recorded; later successes do not establish its cause or cold/load
stability. Eight package maps and52 protected files were unchanged.

Task 3 root integrity passed (875.905s), but public tool verification failed at
dependency presence. Consumer cache binding is now repaired in local commit
`aed4b10d13295fade38dd1f5c7ccd089aff46706`; the full original-env devtool selector
passed379.127s,1036 PASS markers/0 FAIL/1 inapplicable Skip. Exact147 cache preparation
completed62.288s without changing locks. A subsequent real verifier still failed
(1.352s): all147 physical directories exist, but Go omits their Dir metadata when
the tools go.sum lacks their full-module checksums. The previous missing-directory
inference was incorrect; no verifier guard or lockfile was changed to bypass it.
Tool versions, generation and real root lint remain unrun. Two-package compile-only
passed59.767s earlier, with no tests executed. Fresh-machine recovery is unverified.
See the [latest evidence and remaining verifier blocker](2026-09-28-devtools-cache-binding-fix.md).
No success is inferred from fake-tool routing tests. The13 physical
gates, full ordinary/C11/C12 and I2/I3 remain OPEN/blocked by the existing security
event. Docker is installed locally, but availability does not constitute physical
acceptance. No Defender exception, quarantine restoration or bootstrap change was
made. See the [current recovery runbook](../runbooks/repository-recovery.md) and
[version-lock evidence](2026-09-27-devtools-version-lock-evidence.md).
