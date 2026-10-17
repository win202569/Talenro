# Targeted shape / ownership semantics — PS7 and PS5 pass, full gate still blocked

## Approved scope and unchanged candidate

After checkpoint57a58651, user approved a NEW independent budget of two fresh
ordinary-token hosts: PS7 then PS5, each capped at30000ms, stopping on any
failure/timeout. No retries, warmups, reviewer, product/original-test edits or
use of any previously frozen slot. Both slots are now spent (2/2).

Portable entry:
[Targeted wrapper](../diagnostics/isolated-fuzz/offline-supervision-policy.targeted-shape-ownership.ps1).
It invokes unchanged shape6 then unchanged ownership tests with Group=regression:
fixture8 plus all ownership10. It does NOT invoke original35 or reflection3.
The wrapper emits an explicit targeted-only completion marker and limit.

Candidate SHA256 remains
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
Original35 remains
61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3.
Ownership tests remain
C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7.
Shape tests remain
9DAA38EF1F19CA2320D17FCA35CDA87A1B4DFBDFCE960AF40EFB223B9E245B6B.
New portable wrapper SHA256:
816FD00DE9C332FA5D07CA6AF447E741E4FEC6C67FFB95430376092CBDBF885E.

## Fresh results

| Host | Wall ms | Cumulative CPU ms | Exit | Lines / PASS markers | Result |
| --- | ---: | ---: | ---: | --- | --- |
| PS7 7.6.5 Core x64 | 5492 | 8109.375 | 0 | 37 / 25 | Targeted PASS |
| PS5 5.1.26100.9711 Desktop x64 | 9419 | 11468.750 | 0 | 37 / 25 | Targeted PASS |

Each run completed six shape groups, eight fixture groups and ten ownership
groups (24 test groups); the25th PASS marker is wrapper completion, not a test.
Both emitted the exact frozen ordered output inventory including all headers,
stages and limits. Both had zero stderr characters, no timeout,
CaptureValid=true, ExpectedResult=true, FormalProjectAcceptance=false and
owned-host-exited cleanup. Parent and child context recorded admin=False.
CPU is cumulative across process threads and may exceed wall time; it is not
CPU percent or a per-group profile. This is not a speedup comparison.

The previously missing ownership groups SequentialIsolation, PlainControls and
Complete are now observed passing on this candidate in BOTH hosts. In particular,
the unchanged decorated-scalar repeated-refusal/result-independence matrix ran,
rather than substituting the new repeated-call shape test for it.

This closes the specific missing targeted semantic evidence. It does not claim
that the original35 full flow passes on PS5, that a combined full run completes,
or that candidate performance improved.

## Capture integrity and local receipts

Ignored local packet:
.superpowers/sdd/2026-10-17-p1-targeted-semantic-verification/
contains scope, immutable controller/frozen inventory, two invocation rows and
four raw streams. It is deliberately NOT committed.

Before launching, froze183 source/receipt pins: inherited173 plus eight previous
shape-packet anchors plus current candidate and new wrapper. Engine hashes were
checked before their respective launches. Slots were recorded before child Start.
Owned process handles only; bounded wait/drain; no elevated test execution.
After both runs, all183 pins, controller, two receipts and four streams were
read-only rechecked. Prior failed controllers/raw/verdicts were not rewritten.

| Artifact | SHA256 |
| --- | --- |
| Frozen inventory | 72F2654468827F04810711B37198F5FC652FBC6A4EFA5F7E3EE9690B699DE6D2 |
| Controller | 87EC6E4AD9BB3C906861A80B186D4C27435C26B6FC74C1C8530BED8AE17BB6B9 |
| Invocation index | D5082A89243E16C33E578F8B8558F1E6A1480004A24A96FDFDF28D95CDB53F65 |
| PS7 stdout | 6EB5B88D8BBF3982D52A1C4834512D9F36EADF0BE312CBF6A83BE86368A5225A |
| PS5 stdout | AADBD8190E16CC32D5CB4CB008655E0C4B46620DA7CC1B70D9A8064810EC9AB0 |
| Both empty stderr files | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

PS7 engine hash:
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
PS5 engine hash:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Remaining blockers and recovery

The earlier combined full flow still FAILED to finish within30000ms; saved
30071ms timeout is unchanged. Targeted subsets cannot be assembled into a full
PASS. Original full PS5 flow and independent review remain uncompleted.
Candidate remains UNACCEPTED; I1/P1 unresolved. No main/C12/project/live/runtime/
release acceptance, no automatic reviewer eligibility and no security-gate waiver.

Budgets unchanged except this new targeted2/2:
shape2/3 with one frozen; prior independent full1/2 with one frozen;
reflection3/4 with one frozen; extended-view3/7 with four frozen.
Review0/1 remains unused/ineligible. Earlier diagnostic budgets remain spent.

Commit the new portable wrapper plus this report/status/recovery notes to the
existing development branch. Clone restores these portable artifacts and
original tests, NOT ignored raw evidence, cached engines, credentials or
permission to replay. Preserve earlier reports as historical results.

Next focus is the unresolved complete-flow timing, not another semantic rewrite.
Use saved evidence first to define a controlled complete-flow verification/
performance investigation plan. Any new execution requires separately scoped
authorization; do not consume frozen slots, retry unchanged runs, increase the
30s limit, or infer a unique bottleneck from the last completed group.
