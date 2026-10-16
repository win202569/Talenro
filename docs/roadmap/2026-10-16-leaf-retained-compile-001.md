# Leaf retained compile001 — compile exit0, diagnostic sample STOPPED

## Approved execution

After preparing the fixed entry, the user approved continuation to one real
compile-only observation:leaf-retained-compile-001, fixed copied SDK/reference,
existing offline module cache,600s native deadline and unchanged32-attempt cap.
Base commit:0453f31f27dca23bca6c06b1a14d05ca97467c08. Prelaunch checked a clean
worktree,19 source/test pins, frozen reference/copy/go hashes and unused identity.
A fresh Windows PowerShell5.1 host loaded the current collector and invoked the
controller exactly once. No generated test binary was executed, no matrix,
Docker, download, installation, security change, retry or main merge was performed.

The complete11536-file fixed SDK audits before/after execution are separate from
the600s native deadline. The pre-audit produced no incremental progress and was
the lengthy initial wait; that wait must not be reported as compiler runtime.
Native receipt interval:2026-10-16T09:51:52.7608369Z through
2026-10-16T09:53:07.7057170Z,74944.8801ms. The outer command exited1 because the
controller refused a capped diagnostic sample, not because the compiler exited1.

## Verified outcome and limits

- Native root exited0; no timeout or termination request. CompileExitZero and
  CompileEvidenceComplete true:binary present, copied compiler/original leaf test
  source tokens and copied context/internal-fuzz compile tokens verified.
- CoreNaturalComplete/NaturalShutdownVerified/ShutdownEvidenceComplete/
  CleanupVerified true. Root-wait descendants/active0 and final active0, empty
  member lists; natural observation20ms/shared cleanup30ms. Original core member
  evidence has0 rows, so this run does not exercise the historical root-wait
  OpenProcess87 branch or establish its cause/fix.
- Initial accounting reports545 total processes, not545 captured identities.
  Retained diagnostics report1362 sampled queries,32 attempts/admissions/releases,
  0 release failures,0 recorded errors, Incomplete=false BUT
  AdmissionBudgetExhausted=true. Reaching32 is independently capped by policy;
  Incomplete=false does not override it. All32 recorded final states are exited0.
  These raw counters/final states are not full lifecycle validation:the dedicated
  validator stops at the cap before checking the remaining per-entry phases.
- The saved first/last query metadata spans09:51:54.0378239Z to09:52:16.1802961Z,
  while compilation continues until09:53:07.7057170Z. Sampling is neither a full
  trace nor evidence for processes outside those32 admissions. No later PID
  lookup, identity inference, handle reopening or warm-cache retry was performed.
- ProductUnchanged/ToolchainUnchanged/ArtifactsUnchanged true;701 product files
  compared, full fixed SDK pre/post audit succeeded,12 helper/native pins and
  go.exe unchanged. AuditErrors empty, FinalizationError null.

LifetimeEvidenceComplete/ObservationPassed remain false with
`Lifetime sample incomplete or capped`; primary failure is
`Diagnostic compile sample incomplete; no retry`. Both diagnostic flags remain
true. Acceptance/PackagePassed/OriginalTestPassed/RuntimeAttempted remain false.
The unchanged ordinary guard refuses the saved receipt with
`Lifetime diagnostics are not acceptance`. Compiling tests is not running them.

stopped.json exists; completed.json is absent. result.json remains provisional
(FinalizationRequired=true), not durable successful completion. The stopped
outcome and its11 manifest-listed artifact hashes/lengths were verified using
safe single-link paths. Saved-only re-evaluation reproduced core-complete/
lifetime-incomplete and compiler-token evidence without another native Run.
No source/test/collector/guard changes or broader test batch occurred this turn;
the earlier fake283-per-host results remain historical, not rerun here.

## Evidence and recovery

Ignored primary root:.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/
leaf-retained-compile-001. Fingerprints:

- Manifest:`6572BB08617612DD5065109345926715C3288589D274B635D07A41FE7821DC0D`.
- Receipt:`6EDD316680499D16D29D4FD6DFDF3A62B66CC206887D7DD6456DB904BAE47702`.
- Stopped:`A7A29777985509C4C64BC43BB2F208BD1B0DEDD06B4382BC72285B62E0918BA9`.
- Saved-only audit:.superpowers/diagnostics/2026-10-16-leaf-retained/
  compile-001-audit/saved-evidence-audit.json, SHA256
  `09A621238DB9866798EF7DA2CB43432F3A579EC520EB8171980000EF27FE3E95`.

The persisted SDK reference/copy pair is
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55` /
`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`;
go.exe SHA256 is
`D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F`.

Recover the existing development branch, [entry and portable tests](2026-10-16-leaf-retained-compile-entry.md),
current-status.md and [repository recovery](../runbooks/repository-recovery.md).
Git carries the sources/tests and this account, not raw receipts, generated
binary, SDK, caches, credentials or live state. Do not reuse001 or infer live
readiness on another computer from cloning alone.

Recommended next scope:read-only analysis of saved admission/query chronology
and the cap policy's diagnostic usefulness; propose a bounded design only if
needed. Do not increase the cap, reuse warm caches, retry or run the binary/full
matrix without a separately agreed scope. Historical matrix003 remains
2PASS/1FAIL/23UNRUN; error87 and wider C11/C12/Task10/B01/production/main gates stay open.
