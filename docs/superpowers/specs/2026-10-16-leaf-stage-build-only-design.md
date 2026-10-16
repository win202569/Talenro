# Stage-triggered build-only diagnostics — proposed written design

Date:2026-10-16. Base:9119d076f534bd712f25f3bc4c01f721df35392c.
Status:user selected option1,then approved this written spec after
b8e7956dafd54fe884e0ad2ac1554d97944942fe. The
[implementation plan](../plans/2026-10-16-leaf-stage-build-only.md) now awaits
review and execution-method selection. No implementation,fake verification or
real execution is authorized by written-spec approval alone;none is performed
in this documentation increment.

## 1. Intent, evidence and success criteria

Provide an explicit nodeleafbinding compile/link diagnostic whose fixed Go
invocation requests vet off,without disabling required vet elsewhere. Preserve
the32-attempt cumulative budget,existing stage trigger,exact-handle ownership,
cleanup limits and unconditional refusal of diagnostics as ordinary acceptance.
The purpose is a clearer diagnostic workload,not a way to make a test gate green.

The [saved-only analysis](../../roadmap/2026-10-16-compile-stage-budget-analysis.md)
found5 non-vet entries followed by27 vet admissions in about0.95s;225 vet
command/config records followed the target trigger. The fixed SDK schedules vet
even for go test -c. Delayed activation alone did not stop that later workload
from exhausting the budget. These observations do not predict that a changed
invocation will admit only5 entries or guarantee complete capture.

Success for a future implementation means:only the explicit new entry requests
-vet=off;scope/provenance survive serialization;current trigger/core/lifetime
requirements remain;ordinary/default/matrix/eager/full-stage behavior and vet
requirements are unchanged;portable fake tests prove those boundaries on PS7/5.1.
Any later live result stays scoped,possibly incomplete,and never grants acceptance.

## 2. Choice and excluded work

Selected:separate build-only diagnostic controller with a fixed invocation and
distinct owned run identity. Ordinary testing and vet acceptance stay outside
this diagnostic and continue to be required by their existing workflows.

Not selected:filtering retained members by image,which changes admission/ownership
and resource accounting;or increasing the cap,which has no justified complete-
capture bound. No eviction,reopening,second cohort,trigger shift,SDK/product patch,
default/matrix flag change,new native rights or interpretation of87 as success.

An independent vet-only diagnostic is not part of this increment. The existing
full-stage/eager diagnostics are not replaced or automatically chained to the new
entry. No automatic compile-then-vet pipeline,cache warm-up,retry,binary runtime,
Docker,installation,download,security-setting change,main merge or broad acceptance.

## 3. Separate entry and fixed plan

Add docs/diagnostics/isolated-fuzz/leaf-stage-build-only.ps1 with explicit functions:
Get-LeafStageBuildOnlyPlan,Get-LeafStageBuildOnlyEvidence,
Invoke-LeafStageBuildOnlyNative and Invoke-LeafStageBuildOnly. Importing/planning
must not invoke Go,the native runner,PID queries or full SDK/product inventories.
The four planner/controller inputs remain RunRoot,Toolchain,ReferenceReceipt,
ModuleCache;no optional vet switch,free-form argument/environment override or
arbitrary package is exposed by this entry.

RunRoot must be new and absent,beneath the existing
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator scope,with basename
leaf-stage-build-only-NNN (three digits). Do not reuse eager/stage compile001 or
route a synthetic old-name root through the old planner to bypass identity rules.
The planner validates its actual new identity,non-overlap,reparse/single-link
inputs,64-bit host,owned uninstrumented toolchain/reference and existing offline
module inputs using current helpers. Cache/tmp are fresh inside the new root.

Fixed argument sequence:test,-c,-x,-p=1,-vet=off,-count=1,-timeout=120s,
-o=<RunRoot>/package.test.exe,talenro.local/platform/internal/nodecontrol/nodeleafbinding.
Exactly one -vet=off and no alternate/duplicate vet flag. The120s flag is preserved
as a Go argument,not a test-runtime action or a replacement for the native600s limit.
Keep the current copied GOROOT,private GOCACHE/TEMP/TMP/GOTMPDIR,existing GOMODCACHE,
GOFLAGS=-mod=readonly,GOTOOLCHAIN=local,GOENV=off,GOPROXY=off,GOSUMDB=off,GOAUTH=off,
GOVCS=all:off,GO111MODULE=on,GOWORK=off,GOOS=windows,GOARCH=amd64 and empty
TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG. No extra inherited vet override
is introduced by the controller. These settings are not an OS network/filesystem sandbox.

Plan schema1 adds DiagnosticPolicy=leaf-stage-build-only-v1,VetMode=off,
VetEvidenceComplete=false,DiagnosticOnly=true,Acceptance=false,RuntimeAttempted=false,
NoRetry=true,Coverage=sampled. Bind actual go.exe hash and all16 production helpers:
the new file plus the existing full-stage planner's15 named dependencies.
Recheck those hashes and go.exe immediately before launch and in finalization.
Do not count tests as production helpers or omit transitive native sources.

## 4. Reuse without changing the collector

The new file imports leaf-stage-retained-compile.ps1 for existing pure evidence
helpers and audit/persistence primitives. Its planner and explicit controller own
the new fixed scope. Do not turn the old planner/controller into a generic flag-
override dispatcher or pass a mutated old plan to the old controller. Preserve
existing entry signatures,arguments,source bindings and old run identities.

Invoke-LeafStageBuildOnlyNative calls the existing
IsolatedFuzzProcess.RunWithCompileStageRetainedDiagnostics once with the fixed
plan,streams and600s limit. No new native overload,parser,registry,API rights or
C# source changes are needed. The invocation still samples eligible job members,
not an image-filtered promise to retain only compiler/linker processes.

Keep trigger schema1/policy leaf-test-compile-log-v1 and its fixed test-inclusive
target compiler/package/source-directory grammar,16MiB total bytes,64KiB per
iteration/raw line and60000 polls. Keep the exact creation-handle unsignaled check,
candidate/activation/first-query clocks,and no registry before activation.
Use one unchanged registry:32 cumulative attempts,0x101000 access,inherit=false,
no eviction/reopening/retry. Keep600s native deadline,500ms natural observation
and shared5s cleanup. OS calls are not preempted by software deadlines.

## 5. Explicit scope evidence and acceptance isolation

After normal receipt conversion,the new controller alone attaches
BuildOnlyDiagnostic={Schema:1,Policy:leaf-stage-build-only-v1,VetMode:off,
VetEvidenceComplete:false}. This is serialized plain metadata,not a native handle
or proof of vet absence. Do not add a default-null field to all native reports or
invent the marker on historical/ordinary receipts. Both existing diagnostic flags
and CompileStageTrigger remain intact;never clear/strip them to reuse a success guard.

The new scope validator requires exact policy/schema/typed false vet-evidence,
the matching new plan/identity,receipt executable/working directory/stream paths,
ordered argument sequence and complete fixed environment binding. Missing,malformed,
duplicate/contradictory vet policy,wrong root/package or altered command/environment
makes ScopeEvidenceComplete=false,not an inferred default. The marker describes
requested invocation semantics. It does not establish authenticated host identity,
prove no vet process ever existed or replace the copied SDK/source binding.

Validate the saved plan against these fixed constants and all16 required helper
names/pins too;agreement between an equally altered plan and receipt is insufficient.
Saved-only validation permits the existing owned evidence root without executing
the planner's unused-root check. It must not delete/rename the root or fabricate
an unused identity to regenerate a plan. This is evidence reading,not replay.

Get-LeafStageBuildOnlyEvidence composes this scope result with the existing
Get-LeafStageRetainedCompileEvidence without altering the receipt or flags.
That helper supplies trigger/core/lifetime facts,not vet acceptance or an old
workflow identity. Keep CompileExitZero,CompileEvidenceComplete,
ScopeEvidenceComplete,TriggerEvidenceComplete,CoreNaturalComplete,
LifetimeEvidenceComplete and ObservationPassed distinct. The final controller
requires all relevant facts plus unchanged SDK/product/source/artifact audits.
At least one admitted entry and fewer than32 attempts remain necessary;cap/zero
samples/missing trigger cannot become completion. Retained data cannot fill
missing independent core evidence. Current compile-token checks remain required.

Every new plan/result/completion record identifies this policy and VetMode=off.
VetEvidenceComplete,Acceptance,PackagePassed,OriginalTestPassed,RuntimeAttempted
remain false even for completed diagnostic observations. No result can mark vet,
whole-workflow,production or historical acceptance satisfied. Old vet-enabled
diagnostics and ordinary workflows keep their existing semantics and saved results.

Strengthen Assert-ProcessCompletion in run.ps1 to reject any present
BuildOnlyDiagnostic property,including null/false/malformed values,independently
of trigger/mode flags. Preserve all existing refusal and shutdown gates. New
metadata is attached only by the new controller;ordinary converters/default/matrix
runs must not gain a placeholder marker. No other default/matrix dispatch change.

## 6. Execution, failure and durable evidence protocol

Use the existing explicit try/finally lifecycle in the new controller,without
refactoring old execution paths in this increment. Share established primitives,
not a hidden old-controller launch. Persist the plan,check sources,fully audit the
fixed copied SDK against its reference/copy receipts,and inventory product files.
Only after successful preflight create the fresh cache/tmp/build paths and the
single compile-only attempt record. NativeAttempts stays0 until that boundary,
then1;there is never a second call. A later live approval freezes actual inputs
and fresh64-bit host;this spec contains no runnable authorization or prepared identity.

Preserve the receipt/streams,binary if present and artifact binding before
evaluating evidence. Native launch failure records launch-error with cleanup
unverified rather than fabricating a partial receipt. Preserve the primary error
and attempt every applicable independent final product,SDK,source and artifact
audit even after another failure. Audit/persistence failures cannot become success.

Reuse the explicit bounded manifest protocol,never recursively inventory cache/tmp:
plan.json,attempt.json,before-product.json,after-product.json,build/receipt.json,
build/stdout.txt,build/stderr.txt,build/launch-error.json,package.test.exe,
artifact-binding.json,result.json,stopped.json,each only if present. Bind hashes/
lengths and safe single-link paths. result.json is provisional;only successful
diagnostic finalization writes completed.json bound to the manifest and this
scope,with vet/acceptance/runtime flags false. Failure writes STOPPED,no completed
marker,no automatic retry. Do not rewrite source receipts for post-hoc reclassification.

Full SDK/product inventories and inherited whole-log compile-token checks remain
separate from trigger/prefix scan bounds and the600s native deadline. Do not claim
a global memory/time bound. Log visibility,cache effects,short-lived member capture,
timing perturbation and error87 remain unproven until separately scoped evidence;
even a later complete sample is not an exhaustive process trace.

## 7. Required fake verification before live consideration

After written-spec and implementation-plan approvals,use named RED/GREEN tests
with tiny portable SDK/module/log fixtures and fake native boundaries only.
No Go/native Run,binary/PID query,full SDK audit,installation or real vet in these tests.
Use fresh PS7/Windows PS5.1 hosts and actual parsers,hashes,JSON and persistence code.

- Planner imports/creation do not launch;new identity,owned paths/non-overlap,
  exact single vet-off argument,fixed environment,16 named helper pins/go hash;
  old full-stage/eager/default/matrix arguments and invocation remain unchanged.
- Valid scope roundtrip and separate false vet/acceptance/runtime flags;missing/
  null/wrong marker,wrong policy/schema/types,wrong root/package/executable/paths,
  duplicate/missing/changed vet arguments and environment drift all refuse scope.
  A plan and receipt altered in the same way still fail the fixed policy;missing
  or duplicate helper pins cannot pass merely because both records agree.
  Check flat source/argument arrays on PS5.1;do not silently drop pins or rely on
  an extended JSON value/Count wrapper as a substitute for the specified arrays.
- Exactly one fake native call,no test-binary call,no extra vet call/retry;
  both diagnostic flags,raw trigger bytes/hashes and clocks remain unchanged.
  Completed diagnostic remains false for vet/package/ordinary acceptance.
- Ordinary guard refuses a BuildOnlyDiagnostic marker even when null/false or
  trigger metadata is removed and both mode flags cleared;legacy/default receipts
  remain compatible and no new marker appears on ordinary runs.
- Compose trigger/core/lifetime failures independently,including prior provenance/
  clock regressions,zero samples,attempt32 cap,ambiguity,87,forced cleanup/release
  failure;metadata alone cannot repair them. Old receipts cannot become new-scope
  observations by guessing absent policy or reusing their completed/stopped markers.
- Nonzero compile,missing binary/tokens,source/executable drift,independent audit
  failures,multiple errors,result/manifest/completion persistence failures retain
  stopped/provisional semantics;all applicable final audits still run.
- Run existing relevant fake suites,including stage controller,trigger/session,
  receipt/guard,matrix,pair,envelopes,mapping,retained/eager/child paths. Record
  actual host/suite/assertion counts with static checks separate,not old totals.
  Independent whole-increment code review precedes an implementation-complete
  claim. Clearly labelled unfinished checkpoints may still be preserved under
  AGENTS.md;the documentation-only spec checkpoint needs no implementation review.

Fake tests prove implementation contracts,not real vet absence,capture completeness,
new-machine readiness or acceptance. A later real attempt needs separate approval,
frozen current SDK/reference/module/source bindings,new identity,one compile only,
full pre/post audits and no generated-binary runtime.

## 8. Review and preservation boundary

Save this spec and progress/recovery links on the existing development branch;
commit/push only necessary documents,no raw logs/receipts/SDK/cache/binaries.
Self-review for placeholders,scope,interfaces,failure ordering,compatibility and
contradictory claims. Ask the user to review this written spec before invoking
writing-plans;the plan and execution method need their own subsequent review.
No implementation has been made or test result claimed by this spec increment.

Historical eager/stage compile001 remain STOPPED,matrix003 remains
2PASS/1FAIL/23UNRUN,error87 unresolved. C11/C12/Task10/B01/security/production/main
gates stay open. Cloning restores source/design/progress,not caches,credentials,
raw evidence,toolchains or a live environment;recover via the latest development
tip and [repository runbook](../../runbooks/repository-recovery.md),not automatic replay.
