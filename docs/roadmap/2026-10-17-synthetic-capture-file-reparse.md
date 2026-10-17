# File-level reparse evidence and independent review — complete

## Approved scope and immutable snapshot

After274b152a, the user approved one new test file, two administrator-context
children (PS7 thenPS5.1),120000ms each, exactly7 labels, fail-stop/no retry or
repair/security change. Exactly one fresh readonly review was authorized only
after both succeeded. Both prerequisites were met; no extra tests/review seats.
This implements the separate packet required by approved acceptance addendum6.

New portable test: [synthetic-capture-store.reparse.tests.ps1](../diagnostics/isolated-fuzz/synthetic-capture-store.reparse.tests.ps1).
SHA256: `991D75D7926CF63CE70B954E76919568B4DB79541D3B5106D2AAD2AE26DEFECB`.
At review, committed base/head were both274b152afe926ac05b8c0dad0878f758e3a9ee4f;
the frozen new working-file addition was independently read/hashed, not yet in
Git diff. Parent verified actual Git state before/after review; final handoff
commit includes the exact tested file. Product/original/supplemental unchanged.

## Test design and execution

Public writer/reader creates a preserved normal source. Separate new GUID
direct-child directories each contain one CreateNew byte-identical ordinary
file and one real SymbolicLink, targeting the corresponding preserved source
file. Tested leaves are record.bin and complete.bin. No original file is
removed/replaced. Before refusal, each fixture checks non-directory status,
ReparsePoint attribute, SymbolicLink type, exact target and readable original
bytes, including the ordinary companion. Missing or corrupted fixtures cannot
substitute for this precondition.

The real reader must return exact7-field result, Verified=false,Record=null,
authority flags false,Reason=io-error. Source bytes remain unchanged, source
reader still verifies and neither archive acquires stopped. No path/handle/
codec/reader mock, file cleanup, root relocation or permission/security change.
This characterizes unchanged behavior, not a new product RED/GREEN repair.

Seven ordered labels: API required, normal control, payload fixture, payload
refusal, marker fixture, marker refusal, completion. These include preconditions
and sentinels, not seven independent requirements or all OS reparse types.

|New slot|Administrator engine|Exit|Elapsed ms|PASS|Stderr chars|Progress objects|Result|
|---|---|---:|---:|---:|---:|---:|---|
|1|7.6.5 Core x64|0|1781|7|0|0|complete|
|2|5.1.26100.9711 Desktop x64|0|1424|7|584|2|complete|

No timeout/retry. Parent and child administrator guards passed; exact frozen
labels, host/context/LIMIT and source pins matched. PS5 output contains only
two structured Completed progress objects under the previously approved policy,
not ignored error output. XML classification disables DTD/external resolution
and does not deserialize objects. All decoded output remains retained.

## Independent review

One fresh reviewer `synthetic_file_reparse_review` inspected the frozen new
test, relevant production functions, addendum/design and saved receipts after
both runs succeeded. No scripts/tests/dynamic diagnostic executions, edits,
network, fixes, Git writes or further delegation by reviewer.

Findings: **Critical0,Important0,Minor0 confirmed new defects.**

Strengths: normal public control and retained ordinary source at test21..30;
real, readable file links plus matching companion at31..42; exact refusal,
unchanged source and successful reread at44..49. Static path tracing through
store159 onward and isolation33 onward supports actual linked-leaf rejection.
Missing files would yield incomplete-archive; bad directory scope invalid-path.
No validator substitution was found. Raw stdout/index/hash checks agree with
the two observed results and frozen7-label inventory.

Reviewer declined to judge:

- Independent Git provenance: ownership protection blocked reviewer Git reads;
  reviewer made no escalation/trust-setting change. Parent separately verified
  the supplied HEAD/untracked status; reviewer independently verified file hashes.
- New execution/current fixture state: outside readonly authority; runtime facts
  derive from saved receipts, not review reruns.
- Ordinary-user success: both runs are administrator-context evidence.
- Independent effectiveness of handle-level reparse defense: these links hit
  the earlier real path guard; public refusal does not isolate every guard branch.
- Other reparse tags, hostile races, security assurance, whole-branch correctness
  or historical-failure fixes: outside the approved finite packet.

Assessment: approved addendum6 requirement has adequate bounded evidence for
real payload/marker file symbolic links in the two recorded admin environments.
Ready for a separately submitted limited acceptance decision: **Yes**, subject
to parent fixed-version/evidence reconciliation and explicit user approval.
Main merge, release, Stage B/I2/I3/full C12/security/final acceptance: **not granted**.
Parent checked cited source locations and all12 prior source/test pins, with
no disagreement or automatic repair. No second review was requested.

## Integrity, recovery and final gate

New invocation index:
`4BCEBA102E47AA37F80F0E8E8020249F1E67A9B3C623A3F2A73E87517618850D`.
Product/original/supplemental hashes unchanged508F853F/406C7C66/243B4673;
full values in the [approved addendum](../superpowers/specs/2026-10-17-synthetic-capture-store-acceptance-addendum.md).
Prior supplemental/dependency indexes remain33D8DFC4/8867CB8C, with full hashes
in their reports. All12 dependency manifest pins still match.

NEW2/2 children and1/1 review spent. All previous failed records, budgets and
frozen slots remain unchanged. Approved addendum allows a separate new finite
acceptance route; it does not turn old stopped routes into successful runs.
This packet supplies its missing file-reparse evidence and agreed review, not
final user acceptance. See [proposed limited verdict](2026-10-17-synthetic-capture-limited-acceptance-decision.md).

Raw runner/index/stdout/stderr/local progress stay ignored under
`.superpowers/sdd/2026-10-17-synthetic-capture-file-reparse`; every created
fixture is retained in the existing root. Git transfers new test/report and
recovery notes, not raw state or execution authority. No main/security clearance.
