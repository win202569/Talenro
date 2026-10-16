# Build-only001 — diagnostic completed, not ordinary acceptance

## Approved scope and actual execution

User approved the [specific single-run plan](../superpowers/plans/2026-10-16-leaf-stage-build-only-001.md)
after4c588bcf5efc3668c85600bdacbd0ae39b35c455. One existing controller/native
compile ran in the linked codex/c12-b01-task9-coordinator worktree,fresh64-bit
Windows PowerShell5.1.26100.9711. Frozen66 source/test rows,four input hashes,
16 helper bindings,fixed9 arguments/20 environment values and600s native budget
were checked. Source/helpers/tests/C# were not changed.

New owned identity:leaf-stage-build-only-001;fresh private cache/tmp,existing
copied uninstrumented SDK and offline UUID module. Only this diagnostic requests
-vet=off. No generated binary runtime,second Go command,retry,vet-only run,matrix,
Docker,download/install,security change or main merge occurred.

An in-flight checkpoint65e54c545a5f6cd56cbedcb15a27e62541207eb7 preserved the
then-unverified long pre-audit wait. Its “outstanding” state is historical;the
final actual outcome below supersedes it. It did not trigger another invocation.

## Observed result

- Native interval:2026-10-16T12:43:56.8914369Z through12:49:28.1107124Z;
  331.2192755s,exit0,no timeout or termination request. Outer exit0,one returned
  finalized diagnostic state. Binary exists and fixed target/copied compiler
  token checks passed;this is compilation,not execution of tests.
- CompileExitZero,CompileEvidenceComplete,ScopeEvidenceComplete,
  TriggerEvidenceComplete,CoreNaturalComplete,LifetimeEvidenceComplete and
  ObservationPassed are true. Failure/FinalizationError null;AuditErrors empty.
- Trigger Activated at12:49:21.609727Z,324.7182901s after native start;
  first registry query12:49:21.6399919Z.12463 polls,849356 consumed bytes;
  first eligible LF record3321 spans[844915,849356),4441 bytes. Saved raw-line/
  prefix and clock/first-query evidence revalidated without another native call.
- Registry361 queries,5 attempts/admissions/releases,0 release failures;
  AdmissionBudgetExhausted=false,Incomplete=false,no recorded registry errors.
  All five were running at admission and exited0 at root-wait/final reads.
  Optional admission images:1 conhost.exe,2 compile.exe,2 link.exe.
  This is sampled metadata,not an authenticated identity join,proof of vet absence
  or an exhaustive process trace. No counter32 cap was reached in this run.
- Last registry query ends12:49:28.0460979Z,64.6145ms before receipt end;
  retained root-wait/final reads still occur. This does not establish that every
  process,especially early/short-lived children,was observed.
- NaturalShutdownVerified/ShutdownEvidenceComplete/CleanupVerified true.
  Root-wait/pre-termination/final active counts and member lists are0;
  natural observation43ms,shared shutdown58ms,no forced termination.
  Core member rows are empty,so the historical OpenProcess87 branch was not
  exercised or explained. Optional exited-process image reads returned31 for
  all five root-wait/final entries;exit/ownership evidence remained valid under
  the existing image-optional policy. These are not hidden exit-query failures.
- Full11536-file copied-SDK pre/post checks passed with unchanged frozen
  reference/copy identity. Actual701 product files match before/after.
  ProductUnchanged/ToolchainUnchanged/ArtifactsUnchanged true. Saved-only
  verification also matched all66 source/test pins,16 helper pins and4 inputs.
  The module marker is not a full module-cache audit or OS sandbox guarantee.

Policy remains leaf-stage-build-only-v1,VetMode=off,DiagnosticOnly=true,
Coverage=sampled,NoRetry=true. VetEvidenceComplete,Acceptance,PackagePassed,
OriginalTestPassed and RuntimeAttempted remain typed false. Both native diagnostic
flags and trigger metadata remain intact. Ordinary Assert-ProcessCompletion
independently refuses the receipt:Build-only diagnostics are not acceptance.

## Durable completion and saved-only verification

completed.json exists with the fixed policy/false flags and correct manifest
binding;stopped.json and completed.pending.json are absent. result.json remains
provisional with FinalizationRequired=true;it is not the completion authority.
All10 explicit manifest artifacts were checked for safe single-link paths,hashes
and lengths;artifact binding also passed. Cache/tmp were not inventoried.

Evidence-manifest SHA256:
`E16951E7D0EC9E1E8AB51CD7636361EAE388AFC1EC49977B2B5607471819C628`.
Receipt SHA256:
`7CE25F0448AFC691745CA91EBCCAC7D7A27C2F3D7AB25C11092F9FB8C026201C`.
The separate saved reader exited0,reproducing all complete scope/trigger/core/
lifetime facts,marker binding and ordinary refusal. It called no native Run,
unused-root planner,live PID query or full SDK/product inventory. No saved
receipt/log/marker was repaired or reclassified.

## Timing and interpretation

Outer interval12:28:12.2911792Z..12:52:29.4403475Z,total1457.1491683s
(24m17.149s). Pre-native944.6002577s includes host/import/planning/full SDK/product
checks;native331.2192755s;post-native181.3296351s includes validation/audits/
finalization. The long pre-audit wait was not compiler time. The cause of that
long wait was not diagnosed;this is not a controlled timing comparison and must
not be attributed solely to-vet=off.600s is a native deadline,not a global audit
time/memory bound or OS-call preemption guarantee.

## Recovery and remaining gates

Local ignored output:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/leaf-stage-build-only-001`.
Adjacent `leaf-stage-build-only-001-control` preserves wrapper,status,input/source
snapshots,outer log and saved validation. Raw artifacts,SDK/modules/cache/binary
are not in Git and are not restored by cloning. Necessary report/status/plan/
recovery documents travel on the latest development tip,not automatically main.
Never replay/delete/reuse001 or run its generated binary to “finish” this record.

Prior fake884-per-host checks remain earlier implementation evidence;no new fake
or whole-product test batch ran here. Historical eager/stage compile001 remain
STOPPED,matrix003 remains2PASS/1FAIL/23UNRUN,error87 unresolved. Required ordinary
vet/test and broader C11/C12/Task10/B01/security/production/main gates stay open.
Next:return to the outstanding ordinary acceptance checklist/choose the next
separately scoped task,not another automatic diagnostic or a main merge.

## Review and execution rulings

One fresh,read-only saved-evidence/report review independently matched all10
manifest artifacts/completion binding,66 source pins,four inputs and701 identical
product rows. All five Review Focus items were addressed;no Critical,Important
or Minor issues and no deferred minors. It ran neither wrapper nor validator.
Final preservation/remote verification is executor work,not inferred from review.

1. Use PowerShell bookkeeping instead of Bash workflow scripts:Windows-only
   constraint;cost is manual ledger/brief checking.
2. Execute existing frozen software without a new TDD/setup/whole-Go baseline or
   repeated fake batch,as explicitly approved;cost is that prior fake results are
   prior evidence,not a fresh regression claim.
3. Leave general unchanged-implementation correctness/unexercised failure paths
   to earlier implementation evidence or separate work;cost is no new general
   correctness proof from this saved-result review.
4. Do not infer exhaustive capture,vet absence,authenticated images or error87
   resolution from this sample;cost is that those questions remain unanswered.
5. Do not infer why pre-audit was slow or comparative speed benefit;cost is no
   causal/performance conclusion from these timings.
6. Leave fresh-PC readiness and ordinary/security/production/main acceptance to
   their separate gates;cost is no deployment/main-ready conclusion here.
7. Executor verifies final commits,remote SHA and cleanliness after review;cost
   if checked incorrectly is a false preservation/recovery claim.
