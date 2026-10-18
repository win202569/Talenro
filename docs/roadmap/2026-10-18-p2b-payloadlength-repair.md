# P2-B PayloadLength test-only candidate — one full pass, NOT accepted

Date2026-10-18; base3b6b39fa9b750b0a096ef820f970af2f5e6e3325.
Branch codex/c12-b01-task9-coordinator. User answered “按你的建议继续” to
new-candidate one-expression correction and separately bounded verification.
Native maintenance in the existing isolated linked worktree, not a new design
or execution of the frozen original plan. Smallest new allowance:one ordinary
PS7.6.5 full host, no retry/mutant/reviewer. This allowance is now1/1 consumed.

## Change and prior evidence

New portable leaf:docs/diagnostics/isolated-fuzz/p2b-payloadlength-repair/.
Seven files:module,test,fixtures,expected-output,capture,README,.gitattributes.
Product/fixtures/oracle byte-identical to the original; only test content change:

```diff
-foreach($n in @($len-1,$len+1,113))
+foreach($n in @(($len-1),($len+1),113))
```

Readonly comparison enforces the complete test equals the original with this
single replacement; no deleted assertion or product/helper/catch modification.
Prior expression-only diagnostic reproduced MethodNotFound/op_Subtraction
with zero loop-body entries, while parenthesized control yielded15,17,113.
That original failure and diagnostic were observed before this test correction.
No production code changed. PayloadLength now exercises malformed per-kind
declared lengths and nonzero padding against the real decoder, not a mock.

## Exact single-host result

PID23896,ordinary admin=False,Windows PS7.6.5 Core x64,elapsed13200ms.
Exit0,32PASS/35ordered lines exactly match original frozen full oracle;
no timeout,stderr0,owned exit=true,streams complete=true,post pins intact=true.
Outcome=EXPECTED_WITHIN_SCOPE;FormalAcceptance=false. No extra import/probe
or host preceded/followed this approved target run.

Ordered PASS groups (all prefixed P2BCodec):ApiRequired,HelloLiteral,
EncodeAllKinds,DecodeAllKinds,RoundTripSecondary,ExactMessageShape,ExactTypes,
PassiveDescriptors,DecoratedScalars,ByteArrayShape,FrameLength,HeaderMagicVersion,
ReservedZero,PayloadLength,PayloadConstants,MalformedPayloadRange,DirectionRoles,
ObservationSemantics,SequenceExact,SequenceExhaustion,ArmReferences,
StopReferences,HeartbeatReference,Timebase,ArmEquality,BindingMismatch,
FragmentRefusal,InputMutation,OutputOwnership,SequentialIsolation,
ConservativeFlags,Complete. Two leading context/host lines; final line:
`LIMIT: protocol declarations only; no identity, capability or runtime approval`.

## Identity and preservation

| Artifact | SHA256 |
| --- | --- |
| Original test, unchanged | DE23854DF2EB1F2F084625C0CD84312A4024FEA73BD7D51C8316C7409B769E4A |
| New test | 239E80EFDE155F2950A8C8AE9633680D549CF7B825EBCCB5D65D54327375B597 |
| Both module copies | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| Both fixtures | 5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D |
| Both expected-output files | 6D805E7D31F010153FA14ECA22D2E60AEF06560D2DFEB14EED419200A7A680EF |
| New capture | CEFD96D5A7C9C653BCAF8C931BC764A1BCBB997D9B1064CB22568BF2E94FDD0A |
| Engine | 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 |
| frozen.json | 6BF854E076D81112130688BF68200CE9F7F7CB4CE09393F04454440FB7FCD048 |
| receipt.json | D9FA8125F2F2991C6BB04AB11B36CE6511B8046E4F019E92B2C828C4B3E39D70 |
| stdout.bin | 3D5AC08C75B978114BEAA8D857B0D522866B1E60CB2D6896EC5FEC540E515EE3 |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

New ignored packet:.superpowers/sdd/2026-10-18-p2b-payloadlength-repair/.
888mapped pins=868inherited+10original packet files+3diagnostic packet files+
7new candidate files. Rehashed before/after target and again readonly at handoff:
zero drift. Original failed raw/index,source,oracle and all history remain intact.
No packet was removed or committed. Report is portable transcription, not raw.
Controller preconsumes persistently, refuses replay, bounds execution60000ms+
2000ms owned exit+2000ms concurrent separate raw stream drain, owns only its
child, and compares all output against original oracle, never learned output.

## Rulings, limits and next step

TDD/systematic-debugging kept this a one-expression test repair backed by the
prior minimal reproduction; verification-before-completion required the actual
single-host result before any pass claim. Existing worktree reused. No product
implementation/dependency install or architecture extension. Scope ruling:
one separately bounded host overrides generic baseline/global-test/fix/retest/
review/cleanup loops;cost:negative-control/restored-final/review proof absent.
Windows-only PowerShell-native receipt/controller replaces Bash task scripts;
no generic task-done evidence or new agent claimed. No original plan reapproval
or reviewer dispatch is inferred from this test-only maintenance authorization.

Original route2/5consumed,3frozen,old conditional review0/1ineligible;
expression diagnostic1/1spent;this repair validation1/1spent. No resets or reuse.
No mutation or independent review was performed, so not accepted or complete.
Recommended next scope:two original specified negative controls on this new
candidate,immediate restoration after each,then restored full;one conditional
readonly reviewer only after all three exact outcomes. Requires separately
bounded approval;not launched here. No auto-fix/retry on unexpected results.

Necessary files saved/pushed to existing dev branch under standing authorization;
recover there,not main. Clone does not restore ignored raw/cache/mappings or
sessions,reset consumed/frozen allowances or authorize automatic replay.
P2-A limited acceptance unchanged. SourceVerified/LiveActionsPerformed/
RuntimeApproved remainfalse. No real identity/capability/process/Job/clock/
queue/release,PS5/P2-C/P3/C12/live/security/main/acceptance grant.
