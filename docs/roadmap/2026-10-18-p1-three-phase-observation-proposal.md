# Three-phase timing — proposed bounded observation design, NOT PREPARED

Aftere3a1a8ee,user approved design only for root/setup,validation-copy and
decision timing. ZERO new probe/controller/verifier/source implementations,
target/offline-suite launches,retries,warmups or slot releases. This preserves
a proposal for confirmation,not approval to prepare or execute it.

## Purpose and alternatives

Latest candidate PS7 API-call region17567.757ms lacks internal phase records.
H-VC remains untested:the first whole ValidationCopy loop is larger than both
RootSetup and Decision. Retain full input checks/assertions while separating
these coarse regions;do not attempt per-helper optimization yet.

Options:
- Modify the shared evaluator for every prior call or add a4096-selection flag:
  changes earlier call bodies/branches/history. Avoid this wider intervention.
- Reload an instrumented module inside the4096 helper:reintroduces the known
  local-definition loading confound. Not recommended.
- RECOMMENDED:define a uniquely named,phase-marked evaluator clone ONCE at the
  original35 test-script startup;only4096 calls that name. Original13-function
  candidate module remains byte-identical,loaded as before. The clone calls the
  same existing helper definitions;no module reload or instrumentation flag.

There is no zero-interference option here. A separate entry clone has a fresh
name/FunctionInfo/ScriptBlock/source path/body/history. Its command resolution,
compilation and runtime behavior are not proven identical to the old entry.

## Proposed source edits in a NEW isolated copy only

Derive from integer-candidate-api-equality,not any measured-file modification.

1. Preserve the existing candidate module and five other test/wrapper sources
   byte-identically. No changes to their helper bodies,safety checks or loading.
2. Add one definition-only file containing Get-OfflineSupervisionDecisionPhaseProbe,
   a clone of candidate Get-OfflineSupervisionDecision. Same object Scenario
   parameter,Set-StrictMode,result/metadata/validation/copy/decision/catch/return
   operations and order;only rename entry and insert six direct Console timestamp
   statements. No new catch/finally/retry/telemetry branch,per-row counter or cache.
3. Original35 test startup keeps its candidate dot-source at script level,then
   dot-sources the clone definition file once at that same level. That additional
   load defines ONLY the new entry;it does not redefine the13 original functions.
4. Change ONE API command name inside existing P1-SplitCheck4096 to the clone.
   Parameters,local r,full P1-Equal/P1-Need expression,return,outer/inner markers,
   caller arguments and complete test order stay unchanged. Other evaluations,
   including4097,keep using the original candidate entry.

Unlike the prior API/equality split preparation,this step would alter one
diagnostic helper's API target and add one startup definition load. Do not
claim all helper texts or the entire loading context are unchanged. Original
product module/helpers remain unchanged;diagnostic entry identity/history do not.

## Exact proposed phase boundaries

References to3537E9F5 candidate module lines,not implemented new-file numbers:

| Phase | START placement | END placement |
|---|---|---|
| RootSetup |First statement in cloned function,before Set-StrictMode(:115)|After owned-list creation(:131),before validation loop|
| ValidationCopy |Immediately before first for loop(:132)|Immediately after that loop closes(:163)|
| Decision |Before result/decision/context setup(:164)|After second loop closes(:232),immediately before ordinary return(:233)|

Six markers,three sibling elapsed-wall regions inside existing Evaluate4096.
Root includes result initialization,root/type/member/array checks and metadata;
it is not merely metadata lookup. ValidationCopy includes BOTH per-row guards
and snapshot/list append because they are interleaved. Decision includes its
setup and state loop,not only the time case. No attribution of CPU or runtime
cost to individual SDK calls. Parameter binding/entry dispatch/return transfer
are outside or between these internal scopes and remain unassigned residual.

Do not subtract these hypothetical phases from earlier17567.757ms or transfer
old original-PS5 percentages. All values in this proposal are unmeasured.

## Prospective output/parser/fidelity contract

Full inventory would be101 ordered nonempty lines,64 PASS,18 timing markers,
nine closed pairs. Existing95-line sequence remains after stripping the six
new records. Fixed nesting:
Check4096 contains Evaluate4096 and Assert4096 siblings;
Evaluate4096 contains RootSetup,ValidationCopy,Decision siblings.
No overlapping siblings,no duplicated/missing/name-swapped marker or inferred
END. Positive frequency and nonnegative monotonic invariant Int64 ticks required.

The new101-line inventory and depth-three classifier must be DISTINCT from old
91/95-line inventories and their47/63-case tools. Do not weaken/rewrite them.
Future preparation uses test-first offline synthetic positive/nested-prefix/
host/context/order/tick/containment/receipt/failure cases,without product imports.
Compute only closed pairs. Parent and child durations overlap;never sum them as
independent work or call parent-minus-children calibrated overhead.

Failure semantics:if source catches a refusal/error and returns before a phase
END,outer Evaluate END may appear with a phase still open. Preserve that actual
sequence as incomplete/invalid diagnostic with observed early return,NOT a
timeout/permanent hang. Do not fabricate phase ends with finally blocks.
A real timeout exposes the last valid unfinished stack,not its completed duration
or exact row count. Unexpected/malformed output never earns complete status.

Static preparation checks would require AST-only parsing,six exact source hashes,
clone parameter/operation/guard/catch identity after removing markers/renaming
entry,exact two-edit reversal of original35 (API command plus one dot-source),
and exactly one clone call site. No prior source/controller/raw edits.
The new entry clone need not and must not be imported during static fidelity
checks. No autorunning controller in preparation.

## Interpretation and next user decision

Complete ValidationCopy>RootSetup AND >Decision supports H-VC only for that
instrumented run. Root/decision larger falsifies that ordering. Open ValidationCopy
after root closes localizes unfinished work without proving a specific helper,
stable latency,deadlock or original cause. A complete fast run is not untouched
entry acceptance or a matched performance result.

Confirm this bounded design before PREPARATION ONLY:new isolated clone/output,
strict offline classifier/tests and static fidelity. No target launch in that
step. A later proposed execution would need a NEW ordinary PS7 one-host/30000ms/
no retry/no warmup/no PS5 fallback allowance,pinned sources/controller/raw
preservation and owned-process limits. No such authority is created here.

Read-only source/placement checks held367 prior pins and latest invocation
index5090D283DE61BBAE35ABF07521BCEBD3AB512007F4A704ED2CD0C4E0693DEE0D.
Split1/1 and boundary1/1 spent,candidate-full1/2 spent/PS5 FROZEN,older budgets/
review0/1 unused/ineligible unchanged. Candidate/I1/P1/main/C12/security/runtime/
release gates unresolved. Preserve proposal/status/recovery on development
branch;Git restores neither ignored raw,cached engines,credentials nor replay
authority. No product fix or probe implementation delivered in this turn.
