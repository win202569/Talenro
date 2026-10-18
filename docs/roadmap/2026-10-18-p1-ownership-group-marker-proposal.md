# Later ownership group markers — proposed bounded design, NOT PREPARED

After5e6dc245,user approved design only for start/end observation of the two
later ownership groups. ZERO new probe/controller/verifier implementations,
target/offline-suite runs,retries,warmups or frozen-slot releases.
This proposal still needs confirmation before preparation.

## Question and minimal intervention

Latest timeout ended after P1I1PassiveDescriptors PASS with no SequentialIsolation
entry/completion marker. Existing source has nine SequentialIsolation API calls
and five PlainControls calls;both groups passed in two earlier complete candidate
diagnostics. Need distinguish no observed entry,an unfinished body,and completed
body with later timeout. No semantic defect or exact active call is established.

Alternatives:
- Change common P1I1-Case to time all groups:changes shared helper/history and
  broadens capture beyond the two selected groups. Not recommended.
- Wrap/replace per-group API calls or shorten fixtures:changes call structure or
  weakens coverage,not justified.
- RECOMMENDED:four standalone Console timestamp statements directly at the first
  and last statements of the two existing bodies. No common-helper branch,
  extra wrapper,per-call timer,product/function reload or skipped assertion.

## Proposed new isolated copy

Derive from integer-candidate-three-phase in a NEW folder;never edit measured
files. Its candidate module,phase clone,original35 and remaining wrappers/tests
stay byte-identical. Only ownership.tests.ps1 changes,four added lines:
- SequentialIsolation START before its existing three-case foreach.
- SequentialIsolation END after that foreach,still inside the case body.
- PlainControls START before its existing first scenario construction.
- PlainControls END after its existing last complete assertion,inside that body.

Keep P1I1-Case/P1I1-Equal/fixture helper texts,all APIs/arguments/decorations/
metadata/result mutations,full assertions and test order byte-identical.
The shared P1I1-Case still emits its original PASS AFTER the body returns.
No finally-generated end record. Direct Console output bypasses the PowerShell
success pipeline as existing probes do;runtime ordering remains unverified.

The changed ownership script body/source path and four timestamp/output calls
may affect execution/compilation and earlier history,despite unchanged function
texts. Not zero-interference or unchanged-entry acceptance. Existing three-phase
clone interference remains;no additional product-load intervention proposed.

## Prospective inventory and classifier

Complete inventory would be105 ordered nonempty records,64 PASS,22 timing
markers and11 closed pairs (existing18/nine plus four/two).
New timing kind OwnershipCase,Parent empty,strict names SequentialIsolation and
PlainControls. Prefix OWNERSHIP with invariant nonnegative Int64 ticks.
Both are nonoverlapping top-level spans after prior4096 scopes close.

Existing101-line inventory SequentialIsolation PASS is at one-based96 and
PlainControls PASS at97. Insert each group's START/END immediately before its
PASS. Expected new positions:Sequential START96,END97,PASS98;Plain START99,
END100,PASS101. Original Complete PASS/limit/wrappers follow unchanged.
Removing the four new records must reproduce the101-line inventory exactly.

Use a DISTINCT105-line pinned inventory and offline classifier;do not loosen
or overwrite old91/95/101 tools. Fixed global order,ordinary host/context,typed
execution facts,nonnegative monotonic ticks and exact phase/ownership nesting.
Both START and END plus all original PASS/limit/wrapper records and execution
gates are required for complete-diagnostic. FormalProjectAcceptance stays false.

Closed body elapsed includes fixture creation,decorations,API calls,equality/
assertion and mutations inside that group. It excludes shared case dispatch/
parameter setup before START and PASS emission after END. It is NOT individual
API or equality time. Cases are sequential,not nested in each other;do not
combine them with unrelated parent/child intervals as independent CPU work.

## Interpretation without inventing progress

- No Sequential START:entry unobserved;do not say it is hanging in that body.
- START without END at actual timeout:body entered but unfinished;cannot identify
  which iteration/API/equality/fixture statement or a permanent hang.
- END without following PASS at timeout:body reached its final marker;the
  expected completion record is still missing. Preserve that distinct frontier,
  not full-suite acceptance or an assertion failure.
- Sequential PASS then Plain open:later group becomes observed active span.
- Both groups end/PASS but later timeout:retain the later frontier;do not blame
  one completed group.
- No-timeout exit/error or malformed/missing/crossed marker:invalid/incomplete
  capture;do not reclassify as timeout,repair the tail or rerun automatically.
- Complete fast capture:use closed spans only;not guaranteed reproduction,
  root cause,stable speedup or authority to merge/raise the limit.

Future preparation requires test-first offline synthetic cases for all these
frontiers,wrong host/context/receipt/order/name/edge/ticks/overflow/nesting,
plus regression of old tools. Static AST parse only;seven other source hashes
unchanged;remove precisely four markers to recover ownership source C4905925
exactly,with every original helper/body operation/assertion intact.
Do not import/invoke target sources for fidelity checks.

## Scope and next decision

Confirm design before PREPARATION ONLY:new isolated copy,strict105-line
inventory/offline verifier/tests and static fidelity,no autorunning controller.
Any later target run needs separately approved NEW ordinary fresh PS7 single
host,30000ms,no warmup/retry/PS5 fallback,source/controller pinning and bounded
owned-process handling. No timeout increase or frozen-slot reuse.

Read-only placement/inventory checks held388 prior pins and latest index
02A8907BE3E1D12B3013EEF77ECABE4B6BAE314D0EBF034633D4EF07526A56D9.
Three-phase/split/boundary allowances each1/1 spent,candidate-full1/2 spent/
PS5 FROZEN;all old budgets/review0/1 unused/ineligible remain. Candidate/I1/P1/
main/C12/security/runtime/release acceptance unresolved.
Save proposal/status/recovery on development branch;Git restores neither
ignored raw/runtime caches,credentials nor replay authority.
