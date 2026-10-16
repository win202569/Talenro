# Retained-child observation001 — STOPPED, not accepted

Later separately approved [observer phase fix](2026-10-16-retained-child-phase-fix.md)
corrects the code with fake-only tests. This observation's result, bytes and
stopped status remain historical and unchanged; no live rerun or relabeling.

## Approval and execution

User separately approved exactly one real retained-child-observe-001 observation
after the dedicated entry checkpoint4d5f6e1eeb9a9c98d56bdd7fccf7d16dbaacc349.
Prelaunch checks confirmed that commit, a clean existing development worktree,
12 recorded source pins and an unused identity. A fresh Windows PowerShell5.1
host compiled the on-disk C# definitions and invoked the dedicated controller
once. No SDK, Go, Docker, matrix, download, security change or main merge.

Native receipt interval:2026-10-16T09:11:55.6795308Z through
2026-10-16T09:11:58.1997878Z,2520.257ms. Fixed runner limit20s; child requested
150ms after nonce-checked release; marker waits10s. Existing500ms natural/shared5s
cleanup limits unchanged. Outer command exited1 because the observation checker
refused the receipt. There was no native exception, timeout or automatic retry.
The identity is consumed and MUST NOT be invoked again.

## Saved facts, not an observation PASS

- Parent root exited0. All four stage markers are complete and nonce/time/identity
  checked. Child marker identity matches retained EntryId2.
- TotalProcesses4; root-wait active2/list2. These are point-in-time facts, not final
  residuals. The legacy-named DescendantsRemaining field is2 at that same phase.
- Retained registry:67 sampled queries,3 attempts/3 admitted entries,3 release
  attempts,0 release failures. No incomplete/cap/error/ambiguity flag. Every
  entry has owned admission, root-wait and final observations on its retained
  entry. Intended PowerShell child is running at admission/root-wait and exited0
  finally. The other two admission images are conhost.exe in this run only;
  this does not identify any unknown member in older failed matrices.
- Initial accounting/list queries succeeded. Both original root-wait member
  rows satisfy the existing pure final same-handle predicate; intended child's
  final wait records start19ms/end207ms, budget481ms on the shared natural clock.
- Native NaturalShutdownVerified/ShutdownEvidenceComplete/CleanupVerified true;
  TerminationRequested/TimedOut false; final active0 and empty final member list.
  NaturalObservationMs216, ShutdownElapsedMs219. These satisfy the saved limits.
- Final retained image reads report31/null after signaled exit; images are
  optional metadata. Admission images and verified owned handle evidence remain
  separately recorded. No missing image is used to invent an unknown identity.

Saved result:StageEvidenceComplete=true,LifetimeEvidenceComplete=true,
ChildEntryId=2,CoreNaturalComplete=false,ObservationPassed=false. Error is
`Core cleanup bounds incomplete`. Both diagnostic receipt flags remain true;
Acceptance/OriginalTestPassed remain false. stopped.json exists, completed.json
does NOT exist. No provisional or raw result was upgraded to success.
Readiness/release/exit-intent timestamps do not prove exact OS exit times or an
exact150ms exit gap. Successful release calls do not prove every OS/user handle
has vanished. Retention observer effects remain possible.

## Read-only localization

Saved-evidence re-evaluation reproduces the same refusal without another native
Run, PID query or modified receipt. The only true clause in the cleanup-bounds
disjunction is DescendantsRemaining!=0. Source trace:

- process-native.cs assigns this field from CountDescendants(root,MembersAtRootWait)
  immediately after the root wait, before final cleanup.
- retained-child-observation.ps1 incorrectly requires that historical count to be
  zero alongside final cleanup facts, although this controlled fixture explicitly
  allows a child to outlive its parent.
- The unchanged ordinary guard uses this count only for receipts without a
  shutdown schema; schema3 relies on final accounting/list/member evidence.
  The actual diagnostic receipt still gets its unconditional lifetime-mode refusal.

This establishes a wrapper phase-confusion defect, not an error87 cause or fix.
No code, fixture, gate, historical receipt or outcome was changed this turn.
Earlier389-per-host synthetic verification is historical, NOT rerun here. Its
positive synthetic fixture left DescendantsRemaining at0 despite a root-wait
member; that failed to cover this valid lifecycle shape. No independent review
or live rerun is claimed for this analysis.

Recommended next bounded change, NOT implemented/approved here: align the new
observer with schema3 phase semantics, retain the root-wait count unchanged as
diagnostic data, and add portable realistic root-wait-active/final-empty tests
plus negative final-residual controls. Preserve both diagnostic flags, core row
coverage and ordinary acceptance refusal. Obtain that short design approval
before implementation. Any later real observation requires a new separately
approved identity; do not relabel001 or retry it.

## Evidence and recovery

Ignored primary root:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/retained-child-observe-001.
Its manifest binds10 files (plan,four stages,receipt,streams,result,stopped), all
lengths/SHA256 rechecked. Manifest SHA256:
`175C2189F427AFD9673C3E6D2D4E3CA54CDD0B0AE2403DBCDD42C3BE7D87A5F4`.
Receipt SHA256:
`B2331B5C1BF8191FE3566581BBBE18E9E6F107BFF15E34CA384B8CA60A6AEF18`.
Six saved production source hashes and existing shell hash rechecked after run;
shell SHA256:
`60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`.
Stdout0bytes; stderr596bytes contains module-preparation CLIXML progress, not a
native launch exception. No raw streams, IDs, SDK or receipt files are committed.

Separate ignored saved-only audit:
.superpowers/diagnostics/2026-10-16-retained-child/observation-001-audit/saved-evidence-audit.json,
SHA256 `780E1AA37F82025D0EC6C441D703D828C9170DF185E447CDB50CC7851C4E794C`.
It records the unique failing bound, two passing pure final predicates and
unchanged diagnostic guard refusal, without any counterfactual receipt or live
call. Saved IDs were read only from receipts, never queried against live processes.

On another computer recover the existing development branch, read current status,
[entry implementation](2026-10-16-retained-child-entry.md) and
[repository recovery](../runbooks/repository-recovery.md). Git restores code/tests
and this account of the unresolved defect, not ignored evidence or process state.
No product source changes/full SDK audit are claimed. Historical matrix003 stays
2PASS/1FAIL/23UNRUN; error87 and wider C11/C12/Task10/B01/production/main gates remain open.
