# Timeout / observation repair execution — 2026-10-19

Status: STOPPED_PRESTART_BUDGET_GATE_UNFINISHED. No acceptance.

Human approved [bounded request](2026-10-19-p2b-timeout-recording-budget.md)
after0f6b454481fae13b2f3bb742a5b6928ab503f47f: Native NEW6hosts+7children,
conditional1fresh readonly review. Request SHA256
E0C4F018B54E41A8A16D2354A767D627A351D5E66A9A4C7A5DF272842A9389A9.
Historical pending headers left unchanged; latest approval recorded here.
Old sources/evidence/failed results and frozen remainder are not changed.

All14inputs LF/static parse0errors before first Prepare;1795mapping freeze
inherits1670 and adds old96raw/13leaf/report1/approved snapshot1/new14inputs.
First mutation only sequence.ps1, second only capture.ps1; each leaves1794
other mappings unchanged. All tests/oracles/driver immutable throughout.

## Actual runs so far

| Grant | Host PID | Exit | Result | Children |
| --- | ---: | ---: | --- | ---: |
| prefix-red |47916|1|EmptyPrefix+ProperPrefix FAIL, other5PASS|0|
| prefix-green |4392|0|7byte groups+7core groups PASS/full ordered evidence|0|
| recording-red |28060|2|fixed real read-error/assertion abort; evidence absent|0|
| recording-green |44044|2|same fixed abort; required INCOMPLETE evidence exact|0|
| owned-green |15428|2|NONEXACT:4PASS then prestart assertion|3|

First4hosts exact, actually exited/drained, no transport errors/termination/
rescue. Fifth host also exited/drained without outer termination/rescue but
is NONEXACT, never accepted as recording's expected exit2.1795pins intact
before/after. RED absence not fabricated or confused with missing old originals.
sequence-green and conditional review were not run.

Frozen manifest identities in order:

- prefix-red95F4E916A8ECB9B50574DEED85FB64D950AF3654EE73D3EF840FBCAE9B751F19
- prefix-green1B2E33690D1AB268CEE733C50DC48F557A45D9A69167CE0DC9BF10D1D04B13A5
- recording-red6518A80ACEC70BF9B725E1E74869A440946F15678BE77D485531FF1FF33B6486
- recording-green9EA47190D0EF1360A3AB735FFB52EDD004234E487BE52955C28383257907BF6B

98literal byte checks: two streams'0…19prefixes, wrong bytes/extra suffixes,
null/type negatives; normal rejects38short cases. Pure validator is also
called by actual completion, but these tests alone prove no OS/body entry.
Core7groups again exercise disk/no-ack rejection, failed-sink no-reuse,
bad/missing/index handling, real error identity/8secondaries+overflow, actual
closed writer, preconsumed noStart, copied handles;0children/Start.

Recording probe used actual missing ledger ReadAllText/FileNotFoundException
and a subsequent actual assertion. Before assertion the real return/view
was collected; GREEN memory primary remained the same actual reference,
assertion is the sole secondary. Snapshot shows1logical read, failed ledger,
0Start/ack; group RecordingProbeAbort, AfterProbe NotRun. Required root/order/
error projections and whole signature checked independently by parent.
GREEN evidenceACDF55D7D2AF8E25D73A686BF1D0DF52CC275977DAE8D61724994449A5CD3E85.
No rewriting/retry/alternate output; old missing details remain missing.

## Actual owned run and mandatory stop

Source checkpoint15bd2e62f10516fb4391ee1922ffec5d4dd1aaeb pushed/remote equal/
clean tree before the target run. owned-green reused recording-green manifest
and input bytes, not a recovered Session or disk success marker.

Actual output: ReceiptFaultNoRetry, FirstCauseSurvivesCleanup, NormalExit,
Timeout PASS; then stderr ERROR: ASSERT: real prestart open failure, exit2.
Parent rejects stdout/stderr mismatch and does not positively parse the
candidate evidence. Its ChildrenObserved=0 is unevaluated after that rejection,
not a claim that0children ran. Independent post-stop readonly forensics below
uses the preserved candidate, never a positive continuation gate.

The new single evidence file exists, Status=INCOMPLETE, CurrentGroup=prestart,
NotRun=[PreStartFaultNoStart,UnassociatedNoKillWait]. Full actual step roots/
views/errors retained before assertions. Host PrimaryError is the assertion;
the earlier real budget-gate ErrorRecord remains in the step Result.Error and
its Session PrimaryError. Do not claim universal host-level first-cause
promotion was proven: only the dedicated recording probe established that
scope. Unexpected step-error promotion/classification remains a follow-up.

| Case | Actual PID | Exit | Exited/drained | Other actual observations |
| --- | ---: | ---: | --- | --- |
| receipt-fault |48940|-1|true/true|80ms, termination,1receipt attempt/0secondaries/primary same; streams empty|
| normal |41736|0|true/true|472ms, no timeout/termination,exact19bytes each stream|
| timeout |44848|-1|true/true|218ms,TimedOut/termination,both streams0bytes,valid literal prefixes|
| prestart |none|not started|not applicable|Reserved1,StartCalls0,StartIO1,OutIO0,Observation=null|

The actual failing step error is Host target budget exhausted before Process
creation (sequence.ps1:86). Test order launches all3authorized children before
the zero-Start prestart case. The cumulative host gate then rejects3>=3 before
Invoke-LSOwnedStep/stream-open/association observation. Therefore the intended
directory-open error was never exercised; the subsequent test assertion finds
Observation=null. No fourth child or process rescue was attempted. This is a
source-and-record-confirmed ordering/coverage conflict, not another timeout or
an OS association failure. Do not weaken the gate, raise quota, treat the
rejection as the intended IO fault, reorder or rerun within this frozen stage.

owned-green evidence SHA256
BE0B533396A68F5425A93C1EF4F57B8D6054521FE23D45B212DC6750EE5649AC.
Readonly postfailure full1795mapping check:0drift. Actual3children all have
fixed executor exit/drain observations; parent directly observed only the host.
No new independent review, runtime import, query-by-PID or further target run.

Next candidate: a separately authorized ordering correction could exercise the
zero-Start case before the3actual targets, keeping the budget gate unchanged;
unexpected step errors also need explicit host-first-cause classification.
Neither change nor any new test budget is approved by this failed run.

## Portable handoff author check

[Portable data](../diagnostics/isolated-fuzz/p2b-timeout-observation-repair/portable-summary.json)
preserves all103packet originals:33JSON artifacts as raw UTF8 text,70exact
byte artifacts as base64,including all56source snapshots;29directory entries.
All103including6empty files checked against original bytes/lengths/hashes,
0differences. Wrapper physicalLF,3851361bytes,SHA256
AC44BCF608C577835AE7D12C578C549FA758C452191CEA5A9FA6A1E49B187F28.
Raw JSON numeric text is never parsed/reserialized into JS numbers; embedded
line endings and binary bytes stay intact regardless of outer formatting.

recording-red.evidence.json marked expected absent. Two intended prestart
groups,sequence and review marked unrun; current host/global versus per-step
error distinction and unevaluated parent child count explicitly preserved.
No missing originals were reconstructed. Raw kept ignored, no deletion.
Export/byte verification are author-only mechanical handoff, not another
runtime test, independent review or acceptance of the nonexact owned grant.
Captured progress.md is the pre-export ledger snapshot; later handoff commit
SHA and Git transport checks are recorded in Git/tool evidence/final reply.

## Remaining scope and limitations

Spent5of6hosts/3of7children/0of1review. Remaining1host/4children and sole
conditional review FROZEN. Both planned source fix windows consumed. Owned
six-group run incomplete, two prestart groups NotRun; sequence not started.
No automatic repair/reorder/retry/substitute grant/re-review or quota refund.

Source checkpoint includes14inputs/tests and recovery docs; owned incomplete/
sequence unrun, no complete acceptance. Raw ignored on original machine;
portable preservation is evidence only. Losing original context/raw grants no replay;
never reconstruct originals from portable summaries. Existing development
branch codex/c12-b01-task9-coordinator, not main, is recovery starting point.

Bootstrap routing/pre-assert observation/single-envelope transport are the
approved first-assembly boundary, not every error path RED-tested. Late write,
flush/close/host-death not dynamically induced. No process-tree/watchdog/Job/
hard-fileIO-deadline guarantee. Parent directly sees capture hosts; child
lifecycle comes from fixed executor observations, not independent OS relookup.

Author report only; no independent review yet. C1/C2 still unclosed;
R2/Minor2 deferred, old observations/quotas unchanged. No install/compile/
extra suite/production/main/live/formal/full-C12 acceptance. All capability
flagsfalse. Manual finite PowerShell execution replaces skill default extra
Bash/setup/full-suite/task-done/deletion/fix rounds (limited broader coverage);
approved document bytes kept despite historical pending labels (read this
approval record). These are the only process rulings so far.
