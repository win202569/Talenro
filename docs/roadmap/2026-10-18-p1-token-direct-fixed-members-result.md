# Fixed-Members Token direct validation result — 2026-10-18

Status: DIRECT SUITES PASSED / FULL ACCEPTANCE WITHHELD.
Parent checkpoint:6bc5ffbc6ac577f793ee8039a9bdd2e911702452.
User approved the NEW bounded baseline-once/conditional-candidate-once run
described in [fixture revision](2026-10-18-p1-token-direct-fixture-revision.md).
The [original bounded plan](../superpowers/plans/2026-10-18-token-direct-bounded-validation.md)
provides unchanged time,oracle,owned-process and fail-stop terms;this approval
changes target and grants a new separate allowance,not reuse of old slots.

## Observed results

|Receipt|Baseline|Candidate|
|---|---|---|
|Attempt number|1|2|
|Start UTC|2026-10-18T10:29:29.5545281Z|2026-10-18T10:29:37.8378463Z|
|PID|44224|35864|
|Elapsed ms|2344|2158|
|CPU ms|2078.125|2031.25|
|Exit code|0|0|
|TimedOut|false|false|
|Owned cleanup|owned-host-exited|owned-host-exited|
|Streams complete|true|true|
|stdout ordered inventory|429/429 exact|429/429 exact|
|stderr bytes|0|0|
|PostPinsIntact|true|true|
|ControllerError|none|none|
|DirectSuitePassed|true|true|
|FormalAcceptance|false|false|

Both were ordinary inherited-token fresh pinned PS7 processes,no-profile and
noninteractive. Baseline completed all gates before candidate launch.
No co-loading,warmup,reload,retry,extra target,diagnostic or alternate engine.
Per host the fixed inventory is428 direct Check-Call case PASS lines plus one
terminal line,not429 distinct test cases. Each result matched independent
literal success/refusal outcomes,not agreement learned from the other module.

The corrected clean-owner precheck no longer blocks these runs. Cases exercise
both metadata paths,empty/case/unknown tokens,empty/null allowlists,wrong types,
fresh owned strings with all prepared descriptor/shadow combinations,invalid
token refusal order,pre/post descriptor/payload identity,mutation isolation and
zero callback counters. CodeProperty support compiled only within each approved
child;no support getter callback ran in these cases. No fixture assertion skipped.
This provides direct boundary evidence for the tested PS7 inputs,not proof
of every possible binding/descriptor behavior or full API safety.

## Exact source and capture identities

Target:
docs/diagnostics/isolated-fuzz/integer-token-local-candidate/token-direct.fixed-members.tests.ps1
SHA256 1EA8CDB1114EA6A7C68EF29E03F0029E277EFB64FF039EBCA8EB07B706E007A9.
Baseline3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3;
candidateF9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A.
SupportBF82CE30B9CE82449C90ADAA96C0B03F3892609FB903EF87B07C7FD68489409A;
engine362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Portable paths and remaining source pins are in the revision/preparation reports.

Ignored packet:.superpowers/sdd/2026-10-18-p1-token-direct-fixed-members-run/

|File|SHA256|
|---|---|
|run.ps1|5821D48BDBF8CA187B78B67A94A38FFF11CF566131C37848652FEF7C487B5457|
|expected-output.json|ADBFF638C3CB29057509261F3C67944C050F41262A43762D4C5AF60588392494|
|frozen.json|69168924641A4D85D77EE3C19ADD33C4CF023E10646E062AC88D520801440ECC|
|invocations.json|8D8A720591E5D09EA6CB179C50BC05D3290AFF26381A43DB9A6B763D55745366|
|baseline.stdout.txt|5B427B45727DCA607CB1BDC2905F62DFA0FA8C1291E29759D3B3E8D73D62C481|
|candidate.stdout.txt|BA7A57CD871986B4791E3EEFA63DF184F21ADDCB945F074A0B3802E260B1DBA6|
|both stderr files|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

All457 frozen pins verified before/after,including original449 identities and
the prior failed packet's manifest/index/streams. Both current captures/index
were re-read and hash/oracle-checked as data after completion,without re-running.
Original failed baseline remains unchanged and failed;new results supersede
only the revised fixture's pending status. No module/test/support edits this turn.
Controller suppresses only its own VoidTaskResult console values,not target bytes.

## Scope accounting and next gate

New allowance2/2SPENT:baseline1/1SPENT,candidate1/1SPENT. No rerun authorized.
Older direct allowance stays baseline1/1SPENT,candidate0/1UNUSED/FROZEN.
All historical diagnostics/full/review allowances remain spent/frozen as before.
Review0/1unused/ineligible;no independent reviewer launched. Author self-review
of controller/static safety only,not a whole-branch security review.

Ruling:approved bounded terms override general skill defaults for additional
suite/verification hosts,review,installs,Bash helpers,runs and scratch deletion.
Cost:weaker independent/full regression evidence;all wider acceptance withheld.
No RED/GREEN claim:both unchanged product variants passed the revised direct
suite;the earlier failure was in fixture preparation,not a product defect.

Elapsed/CPU values are receipts ONLY. These two non-benchmark hosts do not prove
a speedup or resolve the earlier4096-row timeout. No timing acceptance/model
confirmation,full API equality,refusal-index/result-ownership regression,PS5,
I1/P1/main/C12/security/runtime/release acceptance or main integration.
Candidate stays isolated/unpromoted.

Next recommend a separately scoped bounded full API equality,refusal-index and
result-ownership regression proposal for baseline/candidate on ordinary PS7,
keeping all existing guards and no retries. No such host is authorized or run
by this direct-only approval. Performance work follows only separate approval.

Recover from latest codex/c12-b01-task9-coordinator commit,not main. Git stores
source/plans,this portable evidence and progress/recovery,not ignored controller/
raw streams,caches or credentials. Missing session records on another computer
do not reset any consumed/frozen allowance or grant execution authority.
