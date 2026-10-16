# Retained-child observation002 — sampled observation PASS, not acceptance

## Approval and execution

User approved one new-number controlled observation after the observer phase fix
d0973dde52ea2a1880870afe944f4186dcafd094. Same fixed bounds as001:20s native runner
limit,10s marker waits,150ms requested child delay, unchanged500ms natural/shared5s
cleanup. Prelaunch verified that commit, clean worktree,12 final source pins and
unused retained-child-observe-002 identity. A fresh Windows PowerShell5.1 host
compiled current sources and called the explicit retained-only controller once.
No Go, Docker, SDK audit, matrix, download, security change or main merge.

Native receipt interval:2026-10-16T09:20:52.2833779Z through
2026-10-16T09:20:55.204901Z,2921.5231ms. Outer command exited0. No timeout,
termination request, native exception, automatic retry or second invocation.
The identity is consumed; do not reuse it even though this sample passed.

## Verified sample facts

- Root exited0. StageEvidenceComplete/CoreNaturalComplete/LifetimeEvidenceComplete
  and ObservationPassed true. Nonce/time/identity checks match the intended child
  to retained EntryId2; it is running at admission and root-wait, exited0 finally.
- Root-wait TotalProcesses4, active2/member-list2/DescendantsRemaining2. That
  historical count remains intact; the corrected observer no longer confuses it
  with final residuals. Final accounting active0 and final member list empty.
- Native NaturalShutdownVerified/ShutdownEvidenceComplete/CleanupVerified true;
  TimedOut/TerminationRequested false. NaturalObservationMs208, ShutdownElapsedMs214.
- Registry sampled68 member-list queries;3 attempts,3 owned admissions,3 release
  attempts,0 release failures. Incomplete/admission-cap false, Errors empty.
  All three final retained observations are unambiguous exited0, without forced
  termination. Other admission images are conhost.exe in this sample only; no
  inference is made about any unknown historical matrix member.
- Both original core member rows retain owned running reads and verified same-
  handle final exit0:the child's shared-window final wait starts15ms/ends205ms
  with485ms budget; the other row starts/ends205ms with295ms budget. Registry
  evidence does not replace or repair these independent core rows.
- All final retained image reads report31/null after exit; admission images are
  separately recorded. Image loss stays optional metadata, not a new identity
  exemption or explanation for older error87 cases.

Both diagnostic flags remain true, Coverage=sampled, Acceptance/OriginalTestPassed
remain false. The unchanged ordinary guard rejects the saved receipt with
`Lifetime diagnostics are not acceptance`. Read-only saved-evidence re-evaluation
reproduces ObservationPassed=true with no new process/native Run or PID lookup.

The local completed.json marker has Finalized/ObservationPassed true and binds
the exact manifest hash. All9 manifest-listed artifact hashes/lengths and single-
link safe paths were checked; no stopped.json exists. This is durable completion
of the sampled diagnostic observation, not product or original-test acceptance.
The root-wait-to-final phase distinction is exercised by a real sample after
the fix; historical observation001 remains stopped, not retrospectively passed.

## Limits

One controlled PowerShell lifecycle does not establish arbitrary Go/compiler/job
timing, error87's cause/fix, stress reliability or all handle-release paths. The
68 query calls are sampled coverage, not a complete lifecycle event trace. Handle
retention can change object lifetime. Release success does not prove every other
OS/user handle is gone. Ready/release/exit-intent markers are not exact OS exit
timestamps;150ms is requested delay, not a measured exit gap. No broader test batch
was rerun; earlier fake-only148-per-host/389-per-host results remain historical.

## Evidence and fresh-machine recovery

Ignored primary root:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/retained-child-observe-002.
Its manifest binds9 files:plan,four stages,receipt,streams,result. The manifest
does not include itself or its later completed.json marker. Fingerprints:

- Manifest:`F2A74BAE093C6972618196AF4ADDFAC078B2435B12324C77874A770A4C93D5C9`.
- Receipt:`4CC49EBD9375D17C3BDF125A10F414F7AFF34185FBF4D883BF44DF88DBD0F55D`.
- Completion marker:`F5505EF87A98D3DE64CD1D37E9875F243980407A47BE857EDA04BC2187921933`.
- Existing shell:`60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`.

Six plan-bound production source hashes and the shell hash matched after the
run. No source/test/guard code was edited this turn. Separate ignored read-only
audit:.superpowers/diagnostics/2026-10-16-retained-child/observation-002-audit/
saved-evidence-audit.json, SHA256
`4F3A97013197F0662CB8E83B5E14ECA5AB59FDE17B377B2AD3A7726690EE37D8`.
It records saved-only verification with LiveCallsInAudit=0. Raw receipts, markers,
streams, binaries, SDK, cache and process IDs are not recovery commits.

Recover the existing development branch, current-status.md,
[phase fix and portable tests](2026-10-16-retained-child-phase-fix.md) and
[repository recovery](../runbooks/repository-recovery.md). Git restores sources,
tests and this verified account, not ignored evidence or live process state.
Any new real execution needs its own scope/approval and fresh identity. Do not
automatically rerun a full matrix. Historical matrix003 remains2PASS/1FAIL/23UNRUN;
error87 and wider C11/C12/Task10/B01/production/main gates remain open.
