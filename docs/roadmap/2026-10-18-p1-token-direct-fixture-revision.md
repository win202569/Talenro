# Token direct fixture fixed-Members revision — 2026-10-18

Status: SOURCE REVISION PREPARED / RUNTIME UNVERIFIED / UNACCEPTED.
Parent checkpoint:4c5edbd58f688b97de29079eb14accdab7aba0ae.
User requested following the recommendation:fix only the test precheck first,
with any NEW verification requiring separate approval.

## Source change and hypothesis

[Previous baseline failure](2026-10-18-p1-token-direct-result.md):
3060ms,exit1,no timeout,empty stdout,113-byte missing-Members error.
Read-only source trace points to Check-Owner's first clean-owner branch before
Token. Raw stderr lacks runtime line-stack;exact site remains an inference.
No new reproduction or claim of runtime-confirmed root cause.

New sibling:
docs/diagnostics/isolated-fuzz/integer-token-local-candidate/token-direct.fixed-members.tests.ps1
SHA256 1EA8CDB1114EA6A7C68EF29E03F0029E277EFB64FF039EBCA8EB07B706E007A9.

Original token-direct.tests.ps1 remains byte-identical,including its faulty
precheck,so the original attempt/frozen record remains reproducible as source.
Only new Check-Owner clean branch replaces adapted field.psextended.Members
with the SDK PSMemberSet.GetProperty('Members') getter against the existing
view's BaseObject. Enumerate the returned collection and refuse ANY member;
do not skip the clean check. This follows the existing Passive getter pattern.
No helper call to production Passive/Shape replaces the fixture check.
No proxy,custom metadata,type coercion,cache,output or exception suppression.

The new file shares the original directory's unchanged source-pins.json,
baseline/candidate/support files and leaf LF rules. It does NOT require ignored
session files. Existing README still describes the original prepared test;
use this report to identify the revised future target explicitly.

## Static evidence only

Parser.ParseFile:zero errors in revised and original scripts;no execution.
Four function signatures identical;Direct-Type,Fresh-Owner and Check-Call
function texts identical. All text outside Check-Owner body identical.
Type/base-value and stored-descriptor/member-type/payload identity code before
the replaced branch identical. Existing error literal remains the same.
Clean getter now uses fixed SDK metadata rather than an adapted property.

The unchanged Check-Call still directly invokes Token and checks literal
refusals/no output and zero callback counters,with pre/post owner checks.
Cases/metadata paths/decorations/mutation checks unchanged:the planned future
ordered inventory remains428 case PASS lines plus one terminal line per variant.
This is a source inventory statement,NOT429 passing tests.

No revised/original test,module/helper/support load/compile/invocation,new
controller,target process,synthetic child or diagnostic this turn. No runtime
RED/GREEN or repaired-behavior claim. Original failed test remains failed;
new revision is pending execution. Author self-review only.

All449 prior frozen pins and original packet manifest/controller/inventory/
index/stdout/stderr hashes held before/after. Candidate and product unchanged.
Old direct baseline1/1SPENT,candidate0/1UNUSED/FROZEN;no slots replenished.

## Next separate approval

If approved next,prepare a NEW controller/frozen packet targeting ONLY this
revised script,with new allowance max2 attempts:baseline ONCE first,candidate
ONCE only after complete baseline success. Same ordinary fresh pinned PS7,
30000ms child wall,2000ms owned exit wait,2000ms drain,exact429-line independent
ordered oracle,empty stderr,exit0,complete streams and intact pins.
Compilation of unchanged test-only CodeProperty support permitted only inside
each approved child;no precompile/warmup/helper probe/reload/PS5.
Every launch attempt consumes its slot;any baseline error/fixture discrepancy/
timeout blocks candidate. No source repairs,retry,extra host,oracle relaxation,
full API/performance/reviewer/main authority. Preserve all prior records.
The controller may suppress its own VoidTaskResult console artifacts without
touching captured target bytes or success criteria. None is prepared now.

That next execution allowance is NOT granted by this source-only repair turn.
I1/P1/main/C12/security/runtime/release/full API/candidate acceptance stay blocked.
Save necessary source/report/progress/recovery to existing dev branch;cloning
does not restore ignored receipts or reset any spent/frozen allowance.
