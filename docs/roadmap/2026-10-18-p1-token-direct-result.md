# Token direct validation result — 2026-10-18

Status: FAILED BASELINE / CANDIDATE NOT RUN / UNACCEPTED.
Parent checkpoint:f7e43a2618f2bfeefd1b566da535f910db22f188.
The user approved the bounded two-attempt plan. Only its first attempt ran.
[Approved plan](../superpowers/plans/2026-10-18-token-direct-bounded-validation.md)
and [prepared immutable source identities](2026-10-18-p1-token-local-preparation.md).

## Actual attempt

|Field|Observed receipt|
|---|---|
|Variant/number|baseline /1|
|Start UTC|2026-10-18T10:25:24.3016388Z|
|PID|27956|
|Parent admin role|false;ordinary inherited child token|
|Engine|pinned PS7 7.6.5;no alternate engine|
|Wall/CPU|3060ms /2437.5ms|
|TimedOut/exit|false /1|
|Owned cleanup|owned-host-exited|
|StreamsComplete|true|
|Target stdout|empty;0/429 expected nonempty lines|
|Target stderr|113 bytes|
|ControllerError|none|
|PostPinsIntact|true;449 pins including434 historical|
|DirectSuitePassed/FormalAcceptance|false /false|
|Candidate|never launched;slot UNUSED/FROZEN|

Target stderr, retained unchanged:

> token-direct.tests.ps1: The property 'Members' cannot be found on this object. Verify that the property exists.

There is no passing direct case or candidate execution. This is not evidence
that the Token candidate is correct or incorrect;its behavior remains untested.
Support compilation/module-load code precedes fixture checks in the script,
but no explicit per-stage receipt exists;do not turn that into an independent
support compilation/engine-binding acceptance claim.

## Read-only failure trace; no repair/replay

Source inspection points to the test fixture's clean-owner path:
token-direct.tests.ps1:47 directly reads field.psextended.Members;Check-Call
invokes Check-Owner at51 before Token at53;the first case starts at65.
The strict-mode missing-property error is consistent with this pre-call
access,not a literal Token refusal. Raw stderr lacks a runtime source-line
stack,so exact execution location is a source-trace inference,not an observed
entry marker. No target re-execution,helper probe or diagnostic was used.
Production Passive already uses the fixed SDK PSMemberSet.Members getter
rather than assuming that property is exposed by an adapted PowerShell view.
A later source-only fixture correction must preserve owned type/value,
descriptor/payload identity and callback-zero duties,not skip clean checks.

The outer controller printed two System.Threading.Tasks.VoidTaskResult values
while collecting completed CopyToAsync tasks. These are controller console
artifacts,NOT target stdout and NOT PASS lines. The separate original target
byte streams and strict oracle still reject this attempt. Controller source
and captures were not edited after freezing.

## Frozen packet identities (ignored; not committed)

Packet:.superpowers/sdd/2026-10-18-p1-token-direct-run/

|File|SHA256|
|---|---|
|run.ps1|847272896AAA8F88639DEA4F1670B4244FD69FE67C81A6F34117D1414C293DDA|
|expected-output.json|ADBFF638C3CB29057509261F3C67944C050F41262A43762D4C5AF60588392494|
|frozen.json|958F079178EDB2312F96D3DCA8CAD5877F72E170ED5FAEC4E573A5BD3C32EEEB|
|invocations.json|5DB0616D83327D0FBC6FF66B221942F169EE8B4E3E73C165409A43647173D256|
|baseline.stdout.txt|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
|baseline.stderr.txt|B47B3ABD6A1A53DC8AA22CC46A6AF1FA8EA8557CE69516370DBE78817C07F9EC|

Controller/source/inventory and all449 pins verified before and after.
Original434 pins and prior attribution manifest/controller/index/stdout/stderr
anchors remain unchanged. Target/support/candidate/baseline source untouched.
No candidate stream files exist;the durable index has exactly one failed row.
No attempt to invoke the candidate controller path was made after failure.

## Plan accounting and recovery

Steps1–3:expected inventory prepared first,controller static AST/source review,
ordinary-token and frozen-pin checks performed. No synthetic child tests.
Steps4–6:one baseline launch consumed,streams/receipt retained,strict failure.
Step7:BLOCKED/FROZEN by baseline failure. Step8:portable report/status/recovery
saved with normal dev-branch checkpoint. Task1 is FAILED,not test-complete.

Ruling:the approved bounded plan overrides generic skill defaults for extra
tests/review/install/Bash helpers/reruns/scratch deletion. Cost:weaker independent
validation;all acceptance remains withheld. Author self-review only;no reviewer.
No RED/GREEN/full API/performance claim. No source repairs,oracle relaxation,
retry,warmup,PS5 fallback,timeout increase,main integration or new diagnostics.

NEW direct allowance:max2;baseline1/1SPENT,candidate0/1UNUSED/FROZEN,not available
for a repaired baseline or free candidate run. All old spent/frozen allowances
unchanged,review0/1unused/ineligible. I1/P1/main/C12/security/runtime/release and
candidate acceptance remain blocked.

Resume from latest codex/c12-b01-task9-coordinator commit,not main.
Git preserves source,plan,portable failure summary and recovery state,not ignored
raw/controller/session caches. Missing those on a new computer DOES NOT reset
the consumed/frozen slots. Next recommend separately authorized fixture-only
source correction/static inspection,then a NEW bounded verification proposal;
do not silently rerun or reinterpret this failed attempt.
