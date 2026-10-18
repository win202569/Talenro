# P2-B strict passive descriptor repair — INCOMPLETE_STOPPED

2026-10-18; base eb2b94d91ae75194d5e745bac0f452b35507d3f1.
Branch codex/c12-b01-task9-coordinator. User explicitly approved the NEW
bounded three-host plus one conditional readonly-review route. Three hosts
consumed; first two matched, full suite failed. Review gate NOT met.
This is an unfinished development checkpoint, NOT acceptance or a merge.

## Implemented scope

New portable ten-file leaf: docs/diagnostics/isolated-fuzz/p2b-passive-descriptor-repair/.
Historical source, oracles, packets and allowances remain unchanged.

The helper distinguishes precisely the observed psobject and psextended
metadata by ordinal name, exact CLR PSInternalMemberSet / PSMemberSet type,
reserved+hidden+instance flags and same-object reference to this object's
corresponding SDK sets. Application inventory stays twelve Message fields
(thirteen Expected fields), not fourteen business fields.

After the exact RED/probe, the new product module was patched once and sealed.
Get-P2BCSnapshot now reads the fixed PSMemberSet.internalMembers SDK field,
proves receiver/collection identity, checks every application's exact
PSNoteProperty CLR identity, name and flags, and retains verified descriptors
before reading values. It does not use public Properties/Members enumeration
for incoming root descriptors. Hidden/reserved/noninstance, extra, duplicate,
masquerading and subclass application members are refused.
Wire layout, payloads, reason order, public APIs, result ownership and false
authority flags were not intentionally changed. These statements describe
implementation, not proof of complete correctness or all SDK metadata histories.

The counter-only C# fixture, literal frames, full 33-group tests and output
oracle are byte-identical to the preceding candidate. New helper adds strict
metadata classification and real hidden/reserved/noninstance/hidden-extra/
masquerading negative cases across encoder, binding Message and Expected.
No mocks, external actions, installed dependency, output DLL, product compile
or product test instrumentation. C# fixture compiled in memory in each of the
three already-approved target hosts only.

## Saved outcomes

| Slot | PID / elapsed / exit | Actual result |
| --- | --- | --- |
| RED/probe | 46208 / 1451ms / 1 | Six exact expected lines; EXPECTED_WITHIN_SCOPE |
| Focused GREEN | 17528 / 1802ms / 0 | Six exact lines; EXPECTED_WITHIN_SCOPE |
| Full 33-group | 47800 / 10592ms / 1 | 19 PASS groups, then SequenceExact assertion; INCOMPLETE |
| Conditional fresh readonly review | Not dispatched | INELIGIBLE |

All three: ordinary Windows PS7.6.5 Core x64, admin=False, no timeout,
stderr zero bytes, owned process exited, streams fully drained, post-pins
intact. Full expected 36 lines but actual 22 (context + host + 19 PASS +
assertion). The controller's parent exception is not child stderr.
FormalAcceptance=false throughout.

RED saved evidence:
- Raw exact PSMemberSet / PSMemberInfoInternalCollection<PSMemberInfo>.
- Fourteen entries = twelve application + two strictly identified SDK entries.
- One retained CounterNote; raw enumeration/classification callbacks zero.
- Baseline old encoder callbacks four: MemberReads=2, ValueReads=2,
  Copies=0, Writes=0; intentional final Passivity assertion confirms RED.

Focused saved evidence: the same 14/12/2 inventory, subtype count one,
RawCallbacks=0, final MemberReads/ValueReads/Copies/Writes/Total all zero,
Error=null. Every substituted Message/Expected field, built-in
PSVariableProperty, and flag/extra/masquerade cases reached passive refusal.
Trusted valid encode/binding controls also passed. This demonstrates these
controlled focused cases, not global acceptance.

Full suite passed ApiRequired, DescriptorPassivity, HelloLiteral, EncodeAllKinds,
DecodeAllKinds, RoundTripSecondary, ExactMessageShape, ExactTypes,
PassiveDescriptors, DecoratedScalars, ByteArrayShape, FrameLength,
HeaderMagicVersion, ReservedZero, PayloadLength, PayloadConstants,
MalformedPayloadRange, DirectionRoles and ObservationSemantics.

It failed in SequenceExact at control-frame.tests.ps1:261, through BC-Bind:80
and BC-Need:12. That assertion expects binding-matched with NextSequence
9223372036854775809 for the high UInt64 sequence 9223372036854775808.
Earlier encode/decode comparisons in that same group had completed.
The trace contains only the assertion/error stack, NOT the actual binding
result, reason or then-current member inventory. Its inventory/counters are
the earlier DescriptorPassivity observation, NOT evidence of what happened
inside the failing binding. Root cause is UNDETERMINED. Do not claim a sequence
arithmetic bug, metadata-history bug or false test based on this trace alone.

No product/helper/oracle correction, rerun, extra probe, reviewer, acceptance
or cleanup was performed after this unexpected outcome. SequenceExhaustion
and all later groups were not reached; Complete/LIMIT not emitted.

## Evidence identities

Ignored packet: .superpowers/sdd/2026-10-18-p2b-passive-descriptor-repair/.
Source means the new portable leaf above; packet means that ignored directory.

| Artifact | SHA256 |
| --- | --- |
| packet/baseline-product.ps1 | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| packet/focused.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/focused.stdout.bin | C3C59A189658D3962092B11048D3097D1B5690A95355B21529A4932815574A6F |
| packet/focused.trace.json | 6EAC905F3F16FEE03965CC78DD08C245F0468E677E0E9A9141B1ABC6BAA13B68 |
| packet/frozen.json | 89A67BB759A52B121CBCB7493E8C570FD116A433FD6432181B900CDC96AE301C |
| packet/full.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/full.stdout.bin | 089E06EEDAD06F37EFBB23CE735CE5D5A3B61204171D1D44B9596DA0E717BBBD |
| packet/full.trace.json | 4043E086E282A6608F07B1689C5C246F1A213472AB35EBE281E2946DB18916C6 |
| packet/invocations.json | 6D35BDE10C3559668564C06830F4C034947AF20CA30B191D6FA6FF9C12253ED6 |
| packet/product-seal.json | 0917957D98DBC455E7FA17C7719EAC3987FFC1B3BF551785A9D12C044BFD23D8 |
| packet/red.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/red.stdout.bin | 4B192E805667B1C6E11684684E68ACF06BF926602D0E4BBA2F7988C34F17DB34 |
| packet/red.trace.json | D3F363A30929361F0317AA65EBB05F99D1699916BBD2434B1D39BB0761EF8C9B |
| source/.gitattributes | AE886B465C556A0DC309D704AD3DF90678C7AD856FD288CB11BDD9F697DF060C |
| source/capture.ps1 | 307795A890996D5455A789475DDA1D4D91FCEB18BB9CB0565D77A6A3009B729F |
| source/control-frame.ps1 | F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0 |
| source/control-frame.tests.ps1 | 7CBC263600948F9806D8749B43D0B9DB55C100F48757B1C5CC49CF94A92C60B1 |
| source/descriptor-fixture.cs | 87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE |
| source/descriptor-fixture.ps1 | 93F317B7A51082B8E5BAE1C571E26D5E7CA622CDB6F9C0598CBBCD8E64B17FDE |
| source/descriptor.tests.ps1 | DEE0B50D6FF9FB46FCFBFB55C3858163B1C42C3F68785C07010D9DD34C5A2884 |
| source/expected-output.json | 43740D0D345B3E2896A44D1439CA3ED8294093C410B0C698EA69FE990BDDDEE6 |
| source/fixtures.json | 5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D |
| source/README.md | DF041FB55F1CE9003F38F146467B8E6166515831C814735F3CB1B33481BC752A |

The baseline module remains FA6047A4... in every historical source and the new
ignored baseline snapshot; the new sealed candidate is F76B2DB7....
963 mapped pins = 948 inherited + four prior diagnostic packet files +
one prior live module + nine immutable new inputs + one baseline snapshot.
Read-only reconciliation after STOP: 963 checked, zero mismatches; all nine
saved raw/trace hashes also matched the invocation index. No raw/trace/index/
cache/binary included in Git. Source after sealing remained unchanged.

The controller preconsumes each invocation, enforces strict order/max-three,
checks all historical VerificationPath mappings rather than reinterpreting
historical paths, pins the engine and mutable-product seal, and compares
exact stdout lines removing at most one final terminator. Bounds remain
60000ms execution + 2000ms owned exit + 2000ms concurrent stream drain.
Only its owned child is managed; no process scanning or relaxed safety gate.

## Scope, findings and next decision

Staged whitespace check reports one extra final blank line in README.md:38
and capture.ps1:95. These are formatting-only warnings, not a passing check.
Both files were frozen before execution; they remain byte-identical rather
than being tidied after STOP. No broader whitespace-clean claim is made.

Using-git-worktrees reused the existing isolated branch. TDD required and
obtained the expected baseline RED before the only product patch.
Verification-before-completion and the explicitly approved finite stop rule
prohibit a success claim, automatic repair/retest or reviewer after full failure.
No fresh review findings or declined-to-judge list exists: no reviewer ran.
The historical Important passivity finding has a candidate repair with focused
evidence, but it is NOT closed/accepted because full regression and review gates
failed/not met. Historical Minor comparator finding remains deferred on its old
controller; the new controller's comparator does not retroactively repair it.

New scope: 3/3 hosts SPENT, 0/1 reviewer INELIGIBLE; no remaining host slots.
All previous consumed/frozen scopes retain their recorded state, including
member-set route 1/3+2 frozen, prior descriptor route 1/3+2 frozen,
separate diagnostics 1/1 spent, original 2/5+3 frozen, and previous full/
negative-control/review allowances. None reset, transferred or reused.

Recommended next step, NOT yet approved: one separately bounded evidence-only
diagnostic for this exact SequenceExact path, recording actual binding result
and fixed-SDK raw member identities/flags at relevant stages, while preserving
this candidate and all originals. Establish the failing condition before any
new rule or correction. Do not rerun this exhausted capture controller or
weaken strict descriptor checks merely to obtain green.

Necessary source/tests/report/status/recovery are saved to the existing
development branch under standing authorization. Recover that branch, not
main; clone restores portable files and this report, NOT ignored raw packets,
private caches, sessions, runtime dependencies or allowance reset/replay rights.
No main merge, formal P2-B acceptance, PS5/P2-C/P3/full C12/live lifecycle/
identity/capability/process/Job/clock/queue/release/security/runtime grant.
P2-A's separately scoped limited offline acceptance remains unchanged.
