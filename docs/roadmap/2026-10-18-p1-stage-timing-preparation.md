# P1 stage timing — preparation only, no execution

Date:2026-10-18. Base:fb38f5b8ab0bfeaf56fd18a765173bbab2b5551e.
User approved the bounded in-chat instrumentation design, expressly limited to
preparing code/checklists;no candidate execution or old-slot unfreeze.

[Prepared portable files](../diagnostics/isolated-fuzz/stage-timing-ownership/README.md)
include the independent instrumented test, fixed stage/output inventory,
offline saved-text checker, source anchors, LF attributes and future checklist.
No new launch controller, frozen raw packet or execution allowance was created.
This is unexecuted development code, NOT a verified diagnostic or a timeout fix.

## Preserved contract

Existing candidate1971F808, frozen view-aware test F06D199B, original controller,
oracle and raw evidence unchanged. All582inherited entries remain intact.
Old NEW1/3consumed+2/3FROZEN_UNRUN and older frozen budgets are unchanged.
No target host, product/helper invocation, benchmark, warmup, extra probe,
reviewer or runtime RED/GREEN/full-project test was run. Preparation-only user
scope takes priority over generic extra-verification workflows.

Original15scalar vectors,39behavioral cases,independent13root/22Decision literals,
freshness,metadata mutation,all current and prior-unselected held/reread checks
are retained. Historical selected contexts remain excluded,as before.
No scan optimization,timeout increase,post-evaluation rewrap or result replacement.

## Instrumentation and limits

Flat monotonic stage boundaries separate fixtures,decision calls,oracles,
view capture/selection,selected mutation,current-owner scans,prior-owner scans,
and record/retain. Helpers receive no subject/result/wrapper/payload references.
Original candidate calls remain in their existing scope;no scriptblock wrapper.
Attempted/completed counters remain separate,with80decision/29direct scalar/
1direct metadata bounds.Internal helper calls inside decisions are not included.

Fixed inventory:670ordered stage pairs,1340TIMErecords plus original153base
lines/53fact schemas,maximum1493nonempty success lines.No per-leaf logs.
TIMErecords have19fixed fields and are independently ordered from base output;
both share stdout.Do not assume pipeline/direct-console cross-format ordering.

Body timing starts after entry output and ends before leave output.TIMEs'
serialization/write/flush is accumulated separately.EmitTicksBefore deliberately
lags the current emission;the last emission cost is unknown.Bookkeeping gaps
and other unclassified costs are retained,not relabeled as candidate time.
Parent launch/parse time precedes the child clock origin.This is not a baseline
benchmark;instrumentation itself may make a later diagnostic time out earlier.

An open entry proves only the next stage was ANNOUNCED.It does not prove that
its body started or reveal exact work completed since the last record.
Malformed/truncated/missing records and inconsistent clocks/counters are invalid/
incomplete,never ignored or promoted to an ownership finding.

inspect-capture.ps1 only reads stored text/inventory;it does not launch or import.
It validates fixed stage order/pairing,monotonic/cumulative tick accounting,
bounded counters,and the original ownership facts/prefix.Complete text yields
COMPLETE_TEXT_ONLY,not receipt/runtime/acceptance success.Prefix mode remains
PREFIX_ONLY.RuntimeVerified and FormalAcceptance remain false.
The checker itself has NOT been behaviorally tested.

## Static verification actually performed

- Both new scripts parsed cleanly using the existing host's parser,without
  executing their top-level code,functions or the candidate.
- Removing added timing-only code and normalizing comments/newlines/semicolons
  reproduces all4910original significant PowerShell tokens,in order.This checks
  static preservation only;runtime identity/pipeline equivalence is unverified.
- New JSON loaded;original complete base output/fact oracle compares equal.
  Inventory counts:670pairs,1340TIMErecords,19fields,153base lines.
- Rechecked582frozen entries,candidate and original receipt/stream anchors.
- Diff/checkout-byte checks precede preservation;no ignored raw artifacts staged.

Prepared source hashes(SHA256):

| File in stage-timing-ownership | SHA256 |
| --- | --- |
| .gitattributes | A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9 |
| CHECKLIST.md | F4133CFA82B83457FA1D72980D5573BAA099EB6A5153F27FAC230FEF2AEC2B02 |
| inspect-capture.ps1 | B3C6FE888A6303204B47FC23B9D22F790A207EFCA6956A71021EE738CC19C36C |
| output-inventory.json | 7514F6C3348CF02EE59769F1EA895883E3A92F4C35373F7F7D963686156561D3 |
| owned-return.tests.ps1 | 986CDE7605B88500C35B70D5C4152B1D63370BAC3301B5BA514AC3B2C34AE5BA |
| README.md | E0B6749484B4B6B96AA26232002D6B4A118946AB4F762FA3DFA08FCF35D3F510 |
| source-pins.json | 61E2C2FCA4C24FEBE061A24B6602F6FE5CAA5B5CCFF6E2CF1D3AFD330790DF80 |

## Remaining gates

[Checklist](../diagnostics/isolated-fuzz/stage-timing-ownership/CHECKLIST.md)
explicitly leaves behavior tests,process controller,freeze and execution unchecked.
Suggested next scope is separately approved OFFLINE fixture validation of the
logger/checker contract(no candidate import),before any diagnostic host proposal.
It must cover malformed/reordered/open records,accounting,counter bounds and
prefix-versus-complete classification.No such execution is authorized here.

Before a later candidate diagnostic,prepare an approved bounded controller,
pin the engine,all582old entries plus four latest packet anchors and new files/
schemas,preconsume a NEW allowance and retain30s wall+2s owned exit+2s drain.
No old remaining slot reuse,no automatic retry/repair/reviewer or safety change.
Portable source/progress can be recovered by cloning the development branch;
ignored raw packets and consumed allowances cannot be recreated by cloning.

No I1/main/security/runtime/C12/release/PS5/performance acceptance.
