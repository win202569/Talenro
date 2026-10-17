# P1 ownership fixture — two-host diagnostic observations

Status: **APPROVED DIAGNOSTIC CAPTURE COMPLETE; ROOT-CAUSE VERDICT INCONCLUSIVE.**
User approved the [fixture-only plan](../superpowers/plans/2026-10-17-p1-ownership-fixture-diagnosis.md)
after aa026ba054912e0b402ece2c231945cc17e51311. Both separate diagnostic hosts
completed their exact transport/schema/integrity expectations. Budget2/2 spent,
no reviewer proposed or dispatched. No product/original test import or execution,
fixture correction, alternative construction, original-slot replay or further
probe occurred. This is not product verification or Important I1 closure.

## Recorded host results

|Slot|Context/version|Exit / elapsed|Streams and receipt verdict|
|---|---|---|---|
|D1|Lenovo ordinary inherited token, PS7.6.5 Core x64|0 / 687ms|12 exact lines,9 schema1 rows, stderr empty, ExpectedResult=true|
|D2|same identity/token, PS5.1.26100.9711 Desktop x64|0 / 686ms|same inventory/schema, stderr empty, ExpectedResult=true|

Full identity DESKTOP-H4SVCLU\Lenovo, admin=False. Both were below30000ms,
without timeout/retry/elevation. Each host evaluated the copied original owned
helper path once and its instrumented counterpart once on separate fresh fixtures.
No original test script was invoked. API/product decisions were not evaluated.

## Every saved row reconciled

The [portable probe](../diagnostics/isolated-fuzz/offline-supervision-policy.fixture-probe.ps1)
copies only fixture constructors/type/decoration helpers; it imports nothing.

|Row|PS7 D1|PS5.1 D2|
|---|---|---|
|original-result|FixtureFailure; combined guard=false|Returned; combined guard=true|
|original|Int32, value matches0, Metadata absent|same|
|wrapper-created|PSCustomObject, value does not match0, Metadata absent|Int32, value matches0, Metadata absent|
|metadata-added|PSCustomObject/value false; Metadata NoteProperty, same array=true|Int32/value true; Metadata NoteProperty, same array=true|
|field-assigned|same as metadata-added|same as metadata-added|
|field-retrieved|same as metadata-added|same as metadata-added|
|guard|TypeMatches=false, ValueMatches=false; descriptor present/kind/reference all true; original combined guard=false|all five predicates true; original combined guard=true|
|helper-return|NotReached; all measurement fields null|Observed: Int32/value true, Metadata NoteProperty/same returned array=true|
|conclusion|original/instrumented agree=true; Inconclusive; all four execution/repair flags false|same agreement/verdict/flags|

Type names above abbreviate System.Int32 and System.Management.Automation.PSCustomObject.
Missing metadata produces null kind/reference, not a manufactured false primitive
comparison. D1's ordered false names are exactly TypeMatches,ValueMatches; D2's
array is empty. Both conclusion rows keep ProductImported/ProductEvaluated/
OriginalTestsRun/FixtureRepaired=false. Context/host, all9 rows and final diagnostic
LIMIT marker reconcile against the frozen inventory and each stream hash.

## Supported inference and what remains unknown

For D1, the earliest observed divergence is wrapper construction, before metadata
attachment or caller-field assignment. The copied original helper also failed,
and the instrumented guard shows why that path cannot satisfy its type/value
checks. Metadata presence and nested-array identity did NOT fail in either run.
For D2, the same frozen construction retained the intended primitive and all guard
predicates held, including the returned metadata reference observation.

These observations support a runtime-specific construction-boundary hypothesis,
not a universal PowerShell or CLR-constructor explanation. No overload-resolution
metadata, alternative constructor, conversion correction, integer/bool/string
matrix or additional run was inspected dynamically. Under the approved rule,
two failing D1 predicates and D2's all-true path retain **Inconclusive** rather than
declaring a unique root cause. Agreement of the two copied paths is useful but
does not prove absence of every observer effect. No original stopped run was
replayed; its record alone still cannot name a failed subcondition.

Potential next decision, NOT authority: write a narrow fixture-construction
correction/revised-sequence plan with explicit cross-engine type/value/metadata/
identity requirements. Do not remove/relax the guard or claim that a replacement
construction works before its own approved evidence. Product I1 remains a separate
open finding, not reproduced or repaired by this fixture-only probe.

## Frozen local-byte hashes

|Item|SHA256|
|---|---|
|Probe|8097F9DB4A2F5E18D24AD7133AD7B1BBA54FBBD8B5A14C95509E8FF8B0CCB84F|
|Controller|080A0D34F5A50E467C70C5A17BE90E5F685FB7BD9BB42E26D7552EA1EC238759|
|Manifest/schema/pins|BC7F3E4383E8CE1CC0328B1CC1FB5660455DC7A8E681C0DEB3DE6A6B6863D74D|
|Final two-row index|54D1DE1732675B5D733F7A40776819F00373BB71CF5A6FAED552FBC0854FC7E1|
|D1 stdout|5B74E4756524244EE65D8CC6791D5EC1417BACCBF934DF96654401F7BFCF6946|
|D2 stdout|FDB3F6ED598B46DACFBCED100055894E2EEEEE9047B3513547757EF77C22580F|
|Both empty stderr|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

Protected21 entries = prior protected15 plus unchanged P1 source, stopped new
ownership test, stopped manifest/index/stdout/stderr. Their exact full hashes are
in the [stopped record](2026-10-17-p1-scalar-ownership-repair-execution.md) and the
prior pinned P1 records. Parent matched these before/after both hosts and final
readonly reconciliation; hashes prove consistency, not authenticated execution.
Source FE23DF3B, original suite61A3BACF and stopped new suite9AA7C29C unchanged.

## Rulings, limits and recovery

- Approved fixture-only diagnostics replace generic TDD product cycles, baseline/
  task-done/whole-project tests and review defaults. Cost: no fix confidence or
  independent review; only the named owned construction paths were observed.
- PowerShell-native ledger/controller replaces Bash skill scripts; verified
  existing linked worktree was reused. Cost: narrower bookkeeping, no authority
  expansion. No installations, trust/protection/ownership changes or new worktree.
- All old/new packets retained; prior records read only for pins/receipt integrity.
  Cost: ignored raw is machine-local, not cloned or authenticated by Git.
- D1 has two false predicates and D2 none: honor the frozen Inconclusive rule,
  not a unique-cause claim. Helper-return inspects the caller field against the
  returned metadata reference, or reports NotReached; it does not force execution
  past the guard. No adaptive probes or automatic corrections were made.

New raw packet `.superpowers/sdd/2026-10-17-p1-ownership-fixture-diagnosis/` retains
ledger, manifest, controller, index and both stdout/stderr pairs; nothing deleted.
Recover portable probe/plan/report/status from latest codex/c12-b01-task9-coordinator,
not main. Cloning does not restore raw, conversations or runtime state. The original
repair slot1 stays ExpectedResult=false;6 remaining repair slots stay frozen,
review0/1 remains ineligible. Diagnostic2/2 is now spent. P1 Important I1, all18
prior review limits, security/Stage B/P2/P3/runtime/main/C12/release gates remain
open. A follow-up requires a separately approved fixture correction/sequence;
this completed diagnostic record does not authorize it.
