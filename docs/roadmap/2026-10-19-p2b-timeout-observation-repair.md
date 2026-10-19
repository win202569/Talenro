# Timeout / observation repair execution — 2026-10-19

Status: PREFIX_AND_RECORDING_VERIFIED_OVERALL_UNFINISHED.

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

| Grant | Host PID | Exit | Exact expected result | Children |
| --- | ---: | ---: | --- | ---: |
| prefix-red |47916|1|EmptyPrefix+ProperPrefix FAIL, other5PASS|0|
| prefix-green |4392|0|7byte groups+7core groups PASS/full ordered evidence|0|
| recording-red |28060|2|fixed real read-error/assertion abort; evidence absent|0|
| recording-green |44044|2|same fixed abort; required INCOMPLETE evidence exact|0|

All4hosts actually exited/drained, no transport errors/termination/rescue,
1795pins intact before/after. Recording exit2 is only the frozen negative
signature, not acceptance of generic exceptions. RED absence not fabricated
or confused with missing old originals. Last two target runs not started yet.

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

## Remaining scope and limitations

Spent4of6hosts/0of7children/0of1review. Remaining2hosts/7children: owned-green
3targets, sequence-green4targets; reuse recording-green bytes, not its Session
or files as authority. First nonexact/unknown/record failure freezes remainder
and review, no automatic repair/retry. Two planned fix windows consumed.

Source checkpoint includes14inputs/tests and recovery docs; owned/sequence
still unverified in this stage. Raw ignored on original machine, final portable
export pending. Losing original context/raw does not grant clone replay;
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
