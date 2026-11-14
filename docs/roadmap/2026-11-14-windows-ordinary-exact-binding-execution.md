# Exact launcher binding implementation — synthetic evidence

Date2026-11-14; base c77b1b426aca148439d193f355357876997feb35.
User authorized implementation/synthetic tests only; no live allowance.

## Implemented scope

Only ordinary observer entry and focused test changed. Sole expected command
line is quoted fixed absolute EXE plus one U+0020, profile
shell-quoted-exe-one-u0020-v1. Shared Ordinal identity logic unchanged; no Trim,
alternate accepted form or retry. Arguments empty, ArgumentList absent.

Read-only launcher collector resolves module-qualified cmdlet, inspects loaded
implementing assembly metadata, checks file/ancestor reparse paths and hashes it.
Validator compares fixed PS/edition/command/type/assembly/file/hash fields and
actual command identity. Preflight returns binding only after all existing gates.
Backend clears prior binding on each preflight, captures validated command and
rejects writer/process acquisition without it. Manifest adds profile, exact
expected command and scalar launcher evidence; Context keys unchanged.

## Verification performed

Fresh PowerShell7.6.9 hosts; no App/native probe/build/restore/live CIM/sleep.
RED: exit1, explicit `FAIL: exact one U+0020 command line` against old entry.
GREEN: one focused invocation exit0,751 assertions PASS.
Five existing suites, once each, all exit0:
- observe-c4.tests.ps1:152 assertions.
- c5-observation-session.tests.ps1:120 assertions.
- observation-session-guard.tests.ps1:48 assertions.
- process-observation.tests.ps1:32 cases.
- native-process-observation.tests.ps1:78 assertions.

Test additions cover exact/near-match guard cases, independent launcher constants,
real collector with boundary mocks, command cardinality/type, metadata fields,
missing/read/hash/reparse failures, binding scalar copy, no acquisition on failed
binding, shadowed unqualified launch, stale binding invalidation and manifest.
Prior actual CreateNew collision tests retain two small ignored disk fixtures;
combined primary/cleanup tests remain. Tests read actual cmdlet metadata without
invoking that cmdlet; all process launch callbacks are synthetic functions.

No new app output or original receipt modified. Existing product/shared modules/
old entries remain untouched. Historical failed live run remains failed; this
synthetic result does not establish actual window, command line or verified exit.

## Review and handoff

Independent exact_binding_review completed on base-to-working diff:0 Critical,
0 Important production findings,3 new Minor coverage deviations below. Reviewer
considers implementation ready for a later separately approved live proposal,
not actual runtime acceptance. No extra test/fix budget inferred. Four previously
recorded Minors remain deferred.

New deferred Minors (accepted as coverage limitations, no repairs this turn):
1. tests92: collector zero/multiple/function/missing-file/read/hash errors use
   catch-any Reject and are not routed through acquisition counters. Explicit
   alias, missing Command and missing AssemblyPath fixtures absent. Production
   inspected as fail-closed, but not every named failure/acquisition claim proven.
2. tests127: direct identity negatives catch any exception; tab/CR/LF/NBSP append
   after existing space rather than independently replacing it. Shared Ordinal
   rejects both, but the planned proof matrix is incomplete.
3. tests65: source snapshot mutation is verified then restored before manifest;
   no literal mutate-then-record test. Default sentinels do not explicitly trap
   every metadata read. Copy/inert behavior inspected, these specific proofs absent.

Reviewer exclusions retained: live host serialization/UI/process exit and broad
runtime/C4/security/Release are unverified; component binding is not entire-host
integrity or hostile replacement protection. Reviewer did not rerun tests; counts
above come from main-agent executions. Checkpoint/recovery verification belongs
to handoff. Prior4 Minors remain deferred (AutoFlush disposal diagnostic loss,
independent fixed paths, real Git exit/doc-only HEAD, ordinary loop/disagreement).

Scope ruling: accept these3 findings as nonblocking test-coverage deviations,
not new acceptance evidence — production guards remain exact/fail-closed on
review, and authorized test allowance is spent — cost is less direct proof for
those edge cases. Carry them forward without silently broadening verification.

Execution rulings: preserve PowerShell-only bookkeeping with apply_patch ledger
instead of Bash helpers (manual bookkeeping risk); do not run broad project/build
suites outside synthetic approval (broader regressions unverified). No other
scope change authorized. Plan-owned ignored evidence retained; nothing deleted.
