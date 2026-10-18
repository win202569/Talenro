# P2-B codec negative controls/restored full — review complete, BLOCKED

Date2026-10-18;base07269905d7ba84ca4891be08246c3573aad99bf6.
Branch codex/c12-b01-task9-coordinator. User approved continuation of the
recommended two negative controls,restored full and one conditional readonly
review. Exactly3NEW ordinary hosts plus1NEW conditional review;old frozen
route not reopened. No fix/retry/replacement review or broader implementation.

## Subject and preconditions

Candidate docs/diagnostics/isolated-fuzz/p2b-payloadlength-repair/:
moduleFA6047A4 unchanged;test239E80EF contains only previously approved
PayloadLength bracket correction. Fixtures/oracle unchanged. Saved preceding
full32PASS/35exact lines(13200ms) is prerequisite,not rerun as warmup.
New controller/README/.gitattributes are in p2b-codec-negative-controls/.
Original PayloadLength failure and expression diagnostic remain preserved.

## Three ordered exact outcomes

| Slot | Target PID / time | Exit / output | Result |
| --- | --- | --- | --- |
| reserved-tail | 48116 /920ms | exit1,4lines,API PASS then ReservedZero assertion | Exact expected |
| sequence-equality | 39264 /1088ms | exit1,4lines,API PASS then SequenceExact assertion | Exact expected |
| restored final | 24552 /12924ms | exit0,32PASS/35lines | Exact expected |

All ordinary Windows PS7.6.5Corex64,admin=False,no timeout,stderr0,owned exit,
complete streams,pins intact. FormalAcceptance=false for all. Exact source
mutations were the unique tail check Count32→0 and Sequence -ne→-lt;
not arbitrary error injection. Test groups select only API and the relevant
assertion. Each mutant was immediately restored with apply_patch and hash
verified before the next operation. Final uses original good module hash.
Final stdout matches the prerequisite full byte-for-byte,including LIMIT line.
No tests/oracle/controller edits after Prepare,imports outside the three
targets,extra target run or change to old allowances.

## Evidence identities

| Artifact | SHA256 |
| --- | --- |
| Good module / two immediate restorations | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| Fixed tests | 239E80EFDE155F2950A8C8AE9633680D549CF7B825EBCCB5D65D54327375B597 |
| New capture controller | 14DCAA31EE6AD22BE7636226626C3840C17DCDB1AAD5472BA08AD5844CCF2FFD |
| reserved-tail mutant | 50D01FF9135D68C87C9AB680B6DBDB18BF72FF57C111D2BF72C23AA6F0B65D96 |
| sequence-equality mutant | A0ACD4AC30B6BBE392354705E8FF96B2B9226E48DA8D443C084F972A6459AE94 |
| frozen.json | AE6F1A5BFCF6FAF32A88847E0B216169791C81A1D1CD774183E15A1965BA3235 |
| invocations.json | 9AF0350499BB00B18DE0E8B2D68006944E23B06AD4CDF8E6D89A1EF9C2E52CB9 |
| Condensed readonly review.md | F521A007534FA1B99AA2E72466A03B5944A774F9327A5E44E7B99E8827535F00 |
| reserved-tail.stdout.bin | 7732625FCB49BC6ACE18006261D43EEDE9C251BAAE99B5D9841651A2FEC8281C |
| sequence-equality.stdout.bin | A98C7D4FF9B0534FADEEF43E4F12CF22635ADFD91D16D330CA4B98CED6A3BA4B |
| final.stdout.bin / preceding full stdout | 3D5AC08C75B978114BEAA8D857B0D522866B1E60CB2D6896EC5FEC540E515EE3 |
| All empty stderr files | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Ignored packet:.superpowers/sdd/2026-10-18-p2b-codec-negative-controls/.
895mapped pins=888previous+4previous full packetfiles+3new controller files.
One mutable candidate-module pin maps explicitly to new good-product.ps1;
actual subject independently checked exact mutant/good hashes before/after
each host. No historical manifest rewriting or source drift allowance.
All895pins/sixrawstreams/index and restored module rechecked before review.
Controller preconsumes persistently,enforces ordered max3 and previous exact
outcomes,ordinary parent/pinned engine,separate concurrent raw drains and
only owned child. Bounds60000ms execution+2000ms owned exit+2000ms drain.
No raw packet deletion/rewriting or Git inclusion.

## Conditional independent review

Gate3/3 exact outcomes satisfied. Fresh reviewer
p2b_codec_negative_controls_readonly_review dispatched once with a fresh
context;seat1/1 preconsumed inignored review-seat.md. Readonly scope is the
pure codec/source/tests/fixtures/oracle/controllers and four saved good/mutant
outcomes;inherited sources hash-only. No imports/runs/edits/security/network/
compile/nested agents or replacement review. Review complete:0Critical,
1Important,1Minor. Important finding blocks a no-blocker/acceptance assessment.
No source fix,retest,additional host or replacement review followed.

### Important — descriptor identity not established before virtual access

Candidate control-frame.ps1:64,67,70,73 accepts descriptors by MemberType,
then reads Value without first proving exact PSNoteProperty CLR identity.
An exact PSCustomObject root does not prove its descriptors are passive.
Readonly fixed SDK reflection (also confirmed by coordinator):
PSNoteProperty IsSealed=false;MemberType/Value getters and Copy are virtual,
IsFinal=false;SDK assembly version7.6.0.500. Descriptor subclass can report
NoteProperty and execute code through those accesses,before value type checks.
Tests153–166 cover normal ScriptProperty/AliasProperty/ScriptMethod but not
this subclass boundary. No subclass compilation or execution was authorized
or performed;this is a static contract finding,not an observed exploit run.

Disposition:ACCEPTED as Important blocker after source/SDK verification.
Do not claim callback-free ingress or limited offline acceptance. A future
repair needs fixed metadata descriptor identity before any overridable access,
with independently bounded regression coverage;whether framework collection
access itself invokes such members must also be checked. No repair here.

### Minor — arbitrary trailing blank lines accepted by comparison

Negative-controls capture.ps1:79 and repair capture.ps1:57 use TrimEnd on all
trailing LF,so could hide extra blank output lines despite exact-line gates.
Disposition:ACCEPTED,DEFERRED to future authorized controller correction;
not changed after freeze. Independent reviewer checked actual physical newline
counts35(prerequisite full),4,4,35;no such defect in these saved streams.
Do not invalidate actual evidence or portray the comparator as fully exact.

### Declined-to-judge items and coordinator dispositions

1.Real authentication/capabilities/handles/process/Job/release/stop/cleanup/
security approval:accepted exclusion,pure unit has no such implementation.
2.Live clocks/scheduling/heartbeat loss/real deadlines:accepted exclusion,
only declared integer calculations reviewed,not timing guarantees.
3.Lifecycle/cross-call replay/atomic single-use tickets:accepted exclusion,
intentionally absent,no implied replay protection.
4.Transport/fragment assembly/async cancellation/queues/capacity:accepted
exclusion,no transmission or bounded receiver implementation claimed.
5.P2-C/P3/full C12/broader production readiness:accepted exclusion,no grant.
6.PS5/other engines/platforms/concurrent mutation/hostile global type-data:
accepted untested boundary,no compatibility or concurrency claim.
7.Dynamic CodeProperty rejection/injected internal failures/catastrophic
recovery:accepted static-only/no-compilation boundary,not proven dynamically.
8.Historical source correctness:accepted hash-only protection,not re-review.
9.Historical immediate-restoration timing:reviewer did not witness timing;
coordinator performed restore immediately after each mutant,then checked hash
before next operation. Current identities verified,not independent live proof.
10.Target execution and timeout/kill/drain failure behavior:accepted readonly
review boundary;saved actual successful cleanup is not failure-path testing.
11.Commit/push/main/formal acceptance:not reviewer responsibility;coordinator
verifies routine devbranch preservation separately,no main or acceptance grant.

Review considered all above boundaries without silently discarding findings.
Strengths include independent literal fixtures,strict integer/layout/binding
logic and credible actual evidence;these do not negate the Important defect.
Ignored review.md stores condensed review result;review-seat.md retains seat.

## Rulings and recovery

Verification-before-completion required raw/identity/output reconciliation
before review. Requesting-code-review used one fresh-context independent
reviewer rather than a coordinator-only verdict. Explicit finite scope
overrides generic extra baseline/global tests/fixes/retests/review/cleanup
loops;PowerShell-native receipts replace Bash task scripts;old packets retained.
No acceptance,main merge or production-readiness claim follows from tests alone.
Receiving-code-review required verifying the Important source/SDK evidence,
then accepting the blocker without an unauthorized fix or new target run.

New allowance3/3hosts used,review1/1seat consumed. Original2/5used+3frozen,
original review0/1ineligible,diagnostic1/1spent,repair full1/1spent unchanged.
Necessary controller/report/status/recovery saved on existing devbranch under
standing authorization. Fresh clone recovers source/portable report,not ignored
raw/cache/historical mappings/sessions or replay rights;no budget reset.
P2-A limited acceptance unchanged. No PS5/real identity/capability/Job/clock/
queue/release/P2-C/P3/C12/live/security/runtime grant;all product authority
flags remainfalse. Any limited offline acceptance is a separate user decision.
Current next step is a separately bounded descriptor-passivity investigation/
repair proposal,not acceptance. The Minor comparator hardening is also deferred.
