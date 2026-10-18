# P2-B cold-cache repair — bounded verification complete

Date: 2026-10-18. Base commit: f3405822f444666269980a219e6df30a2418edf9.
Branch: codex/c12-b01-task9-coordinator. No main merge or formal acceptance.

## Outcome and implementation

The preceding retained-scalar candidate's P1 is reproduced and resolved within
the approved structural-compatibility contract. A null wrapper _instanceMembers
cache no longer proves absent metadata: the candidate passively queries the fixed
SDK ConditionalWeakTable by the existing type/local-aware key, then applies the
unchanged exact raw-descriptor validation. A genuine absent table remains valid.
There is no cache assignment, insertion, SDK factory/key-method invocation or
descriptor getter/Value/MemberType/Copy execution in the added product path.
Fixed field/type/method signatures are checked. Actual PSObject keys travel in
object[] reflection arguments. Copy-P2BCScalar through EOF is unchanged.

The prior 327E8F33 candidate remains historically unaccepted; its review/receipts
are not rewritten. The old B468B564 baseline and failed 18DA24E1 are unchanged.
The bounded M1/P1 repair is now verified; this is not a universal provenance,
passivity, isolation or whole-project claim.

## Approved execution and recorded results

New allowance: exactly three ordered ordinary pinned PS7 targets, maximum
60000ms execution plus existing 2000ms owned-exit and 2000ms stream-drain bounds.
All three are spent; one conditional independent read-only reviewer is spent.
No additional target/probe, postreview product edit, or second review occurred.
Old latest 1/3 spent + 2 FROZEN and every older budget remain unchanged.

| Gate | PID | Elapsed ms | Exit | Recorded result |
| --- | --- | --- | --- | --- |
| RED baseline | 19908 | 14339 | 1 | Exact expected 36 shared-dirty misses; other 48 controls match |
| Focused candidate | 40332 | 28028 | 0 | All 84 match: 38 accepts, 46 refusals |
| Original full suite | 836 | 46366 | 0 | Original 33 groups; exact 36-line oracle |

All receipts: expected within scope, no timeout, empty stderr, completed owned
exit and separate drains, all 1185 mapped pins intact. Full trace: zero callbacks,
null error, RuntimeApproved=false. Saved results are execution evidence, not
independently authenticated historical-process attestation.

The new 84 cases contain 18 absent-clean, 8 shared-clean, 32 shared-dirty,
10 isolated dirty-peer, 10 own-key dirty, and 6 actual-wrapper public-decode cases.
74 subjects are cold; 10 own-key controls already have materialized caches.
78 direct calls preserve cache identity. Peer descriptors retain exact references,
classes/name/flags and all controlled counters stay zero. Public clean copying
may materialize its cache. Before invocation, fixed-field/oracle snapshots prove
cold-subject cache state without generating its SDK member views.
The baseline misses are 32 direct plus 4 public shared-dirty cases; all 36 become
refusals. The retained suites also cover 27 proofs, 14 API histories, 35 matrix
rows, 182 malformed descriptors, 45 direct guards and 54 materialized key controls.

## Independent review and parent disposition

[Complete review](../diagnostics/isolated-fuzz/p2b-cold-cache-repair/review.md):
Critical0 / Important0 / Minor0; no required item incomplete. The reviewer
independently reconciled 1185 pins, stream/seal hashes, 168 cold-case rows across
RED/focused, retained results and exact full output without running product/tests.

Parent accepts the narrow P1-resolution conclusion after source/evidence
reconciliation. Every declined behavior remains excluded, not silently waived:

1. Fresh reviewer reproduction: prohibited; saved evidence only.
2. Other SDK/PS5/platform/architecture/elevated contexts: unverified.
3. Arbitrary private reflection, forged helpers or SDK corruption: outside contract.
4. Concurrent mutation/TOCTOU/process isolation/adversarial replay: not established;
   workflow replay prevention is not a tamper-proof authority ledger.
5. Unforgeable factory provenance: not claimed.
6. Nested-wrapper and empty-existing-table cold combinations: no dedicated dynamic
   vectors. Inherited nested-key logic is unchanged and zero-descriptor traversal
   is statically supported, but no executed coverage is claimed for these cases.
7. Historical implementations/stopped scopes: hash-only, not reopened.
8. Performance, arbitrary depth/resource exhaustion, live/security/identity/
   capability/runtime approval, formal P2-B/PS5/P2-C/P3/full C12 or main integration:
   not established. No authority flag is promoted.

These exclusions do not block this bounded repair handoff. They cannot be used
as acceptance evidence for a wider contract. Any further execution needs a new
explicit finite scope; do not automatically replay capture.ps1 on another machine.

## Durable provenance

| Artifact | SHA256 |
| --- | --- |
| New-scope baseline | 327E8F3362D92EF30A2BA71A9B0E8ECCEB896778F927560192483FA37FA2AC29 |
| Candidate | 9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C |
| Frozen manifest | 3B6923094EFCF4C118E1293C9E66058E3E787D2A6001D749D8850BD8A422C0B2 |
| Product seal | 14ED69F9C788C545A19964B9D3C4F90D288482E424B82628AD1A4824B5B91708 |
| Invocation index | C4611B23B9E70BF24F32168A571DDBB4E706E35350C7DD64857AFC46F09BA0CE |
| RED stdout | F00EB6C2D502FFCB2B9DB330BD3BFA9D866575DDF0AE3BFDA270D6E634E4D8B7 |
| Focused stdout | 21AFEF0E7F30AA0F552890B940C39AAFFA83334CF16CDEAD5E620FC25F2AC02F |
| Full stdout | B7C48E01A63F3E338CFC7E3D8C17A3BD2880CFAED63E166AEAC43A20D106AA62 |
| Full trace | 1F9EE5357CF7DB4CE8D7C023BE6C017EDDB9E160D26F433AC5091C1329F43D12 |
| Every empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Portable summary | 362808EA91F100312B6720C72B68BBC26A74017BC0E603D60DCCF622964F2595 |
| Pinned engine | 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 |
| Pinned SDK | 2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B |

Engine: ordinary PS7.6.5 Core x64, SDK assembly7.6.0.500. The local frozen packet
is .superpowers/sdd/2026-10-18-p2b-cold-cache-repair/ and remains ignored.
The 1185 pins comprise inherited1154 + prior packet12 + prior report/summary/
review3 + prior sealed product1 + new immutable inputs14 + new baseline1.

## Recovery and next decision

Commit/push necessary source, fixtures/tests/controller, derived portable summary,
complete review, this report and updated status/recovery to the existing devbranch.
No main merge, force push, PR, branch deletion, tool install or security change.
The summary contains ordinal reference IDs and recorded CLR type names, not live
objects or actual retained CLR references. Clone restores portable work/history,
not raw receipts, private mappings, runtime/SDK binaries, caches, sessions,
credentials or replay authority. Read repository-recovery.md before resuming.

Next is a separately scoped limited offline acceptance/remaining-roadmap decision,
not automatic live implementation, full C12 acceptance or another test round.
