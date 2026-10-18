# P2-B SDK structural compatibility — three exact hosts; review complete with Minor limit

2026-10-18; base98f21970a41288ca2acbd008f46cde8ae2657135.
Branch codex/c12-b01-task9-coordinator. User explicitly approved the short
bounded design: structural metadata compatibility, new three ordered ordinary
hosts and one conditional fresh readonly review. This is not main/acceptance
or a grant to execute real session capabilities.

## Implemented change and preserved contract

New portable ten-file leaf:
docs/diagnostics/isolated-fuzz/p2b-sdk-structural-compatibility/.
Historical F76B2DB7 module and every earlier source/oracle/packet remain intact.
New sealed module:
B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.

Mandatory psobject/psextended identity checks remain. Optional psbase/psadapted
require exact PSInternalMemberSet/name/flags and matching PSObject owner
references. Optional pstypenames requires exact PSCodeProperty/name/flags/
owner, its stored getter reference identical to the fixed SDK MethodInfo,
and a null setter. All reads use fixed SDK fields. No metadata Value,
MemberType, Copy or TypeNameOfValue getter/factory is called by product ingress.

PSObject owners are retained in reflection argument arrays; direct CLR
argument binding could unwrap them to shared PSCustomObject bases and lose
their distinct identities. Tests substitute actual foreign owners to exercise
these checks. Metadata validation admits a structural SDK descriptor shape,
NOT an unforgeable factory-origin assertion or sandbox against arbitrary
same-process private-field reflection. This limitation was explicitly approved.

All business descriptors still require exact PSNoteProperty before any field
value; exact Message12/Expected13 types, wire256, all literal payloads/directions/
sequence arithmetic/reason priority, independent owned Result10 and three
false authority flags remain the contract. Product remains function definitions
only at load, no C# dependency, process/Job/clock/file/network/live functions.

Only the new module changed after the expected RED, then was sealed once.
Full33-group test file, literal12-frame fixtures, counter C# and focused target
are byte-identical to the previous candidate. New helper extends the existing
DescriptorPassivity group without adding/removing full-suite groups.
No mock/product-generated oracle.

## Actual bounded results

| Slot | PID / elapsed / exit | Exact result |
| --- | --- | --- |
| RED/probe | 47272 / 1337ms / 1 | Six expected lines; EXPECTED_WITHIN_SCOPE |
| Focused GREEN | 28864 / 2050ms / 0 | Seven expected lines; EXPECTED_WITHIN_SCOPE |
| Full33-group GREEN | 46664 / 18352ms / 0 | Thirty-six expected lines,33PASS; EXPECTED_WITHIN_SCOPE |
| Conditional readonly review | One fresh readonly seat completed after all three exact | No verified Critical/Important; one Minor coverage limit |

All three targets ordinary Windows PS7.6.5 Core x64,admin=False,no timeout,
stderr0,owned process exit,complete concurrent stream drain,post-pins intact.
FormalAcceptance=false throughout. Each target compiled the unchanged
counter-only fixture in memory in that same approved host; no output DLL,
installed dependency or separate compiler/test host.

RED first proved both valid input shapes with all five SDK metadata entries,
correct exact PSObject owner and fixed code references,zero CounterNote
callbacks. The baseline binding then returned invalid-input/null successor,
followed by the intentional MetadataCompatibility assertion. Thus product
implementation was gated on an observed behavioral failure,not a text test.

Focused GREEN covered ordinary/reused public SDK enumeration,valid encode and
binding with literal next72623859790382857;30 metadata negative cases across
encode/binding-Message/binding-Expected (foreign instance/internal owners,
wrong getter/setter or hidden flags);37 per-business-field CounterNote
substitutions;15 earlier flag/extra/masquerade cases;built-in PSVariableProperty
refusal;and fresh valid controls. No business schema expansion.

Focused trace and full trace show Error=null,OwnerAndCodeReferencesVerified=true,
MessageApplicationCount12,ExpectedApplicationCount13,SDKCount5 for the controlled
proof,zero proof counters. Later fresh raw ingress retained14entries =12
application+2mandatory SDK,one CounterNote subtype,raw callbacks0,exact owner
typePSObject. Final MemberReads/ValueReads/Copies/Writes/Total all0.

Those counters measure only the compiled CounterNote fixture. They do not
instrument arbitrary PSCodeProperty getters; absence of product metadata-getter
calls also requires source review. MetadataGettersInvoked=false is a test
design assertion,not an independent spy for every SDK getter.

Full stdout contains all33 ordered PASS groups including previously failing
SequenceExact,SequenceExhaustion,all kind literal encode/decode,Reason/flags,
reserved bytes,ownership/mutations,sequential isolation and Complete,then the
exact protocol-only LIMIT. Focused/full rich trace hashes coincide because
their helper observations are deterministic; separate PIDs/raw output streams
and invocation rows document distinct target executions.

## Frozen source and receipt identities

Ignored packet:.superpowers/sdd/2026-10-18-p2b-sdk-structural-compatibility/.
source means the new ten-file leaf;packet means that ignored directory.

| Artifact | SHA256 |
| --- | --- |
| packet/baseline-product.ps1 | F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0 |
| packet/focused.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/focused.stdout.bin | AF213029B09BBBBAC80F5D8173CC1263E4047D1E255EE451FF6E361AC31DF493 |
| packet/focused.trace.json | 1F9EE5357CF7DB4CE8D7C023BE6C017EDDB9E160D26F433AC5091C1329F43D12 |
| packet/frozen.json | BFE501435C9DC055FA0D98E12E4EDBD0F3F609716D15EC79B71C1A74844EC07C |
| packet/full.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/full.stdout.bin | B7C48E01A63F3E338CFC7E3D8C17A3BD2880CFAED63E166AEAC43A20D106AA62 |
| packet/full.trace.json | 1F9EE5357CF7DB4CE8D7C023BE6C017EDDB9E160D26F433AC5091C1329F43D12 |
| packet/invocations.json | 76066E6C997803A8509EAD6A5139C1CA94DCDC7686F8A8D855BCB4B3198A1981 |
| packet/product-seal.json | 9FE06DD1A6F6E166668D4F273C5434A9D4C3D59730D235F0AAED26286C60BCD7 |
| packet/red.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/red.stdout.bin | B66365276607D60F2B676E699121619F536E3CC6567FBF1079854837A01A5A95 |
| packet/red.trace.json | 4EABBA0B8EA2EE078B194166673F7EDBBFBE0279391335F01E4D74BB7EEE8F23 |
| packet/review-dispatch.json | 682DD760DBC72232294ADC368F4E141DA8C194A7DDC5C098F2664A2B89E02521 |
| source/.gitattributes | AE886B465C556A0DC309D704AD3DF90678C7AD856FD288CB11BDD9F697DF060C |
| source/capture.ps1 | 0BEA917A3275507A4A63C4E979CDC72AAC55FDD15F3458F497068A92E5B9A1F8 |
| source/control-frame.ps1 | B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039 |
| source/control-frame.tests.ps1 | 7CBC263600948F9806D8749B43D0B9DB55C100F48757B1C5CC49CF94A92C60B1 |
| source/descriptor-fixture.cs | 87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE |
| source/descriptor-fixture.ps1 | 402DFB2C2CAF4EFE6F696EB5AAA7FBAE5C983488A2BF7B3FFFF2D39258D63D0E |
| source/descriptor.tests.ps1 | DEE0B50D6FF9FB46FCFBFB55C3858163B1C42C3F68785C07010D9DD34C5A2884 |
| source/expected-output.json | 61FDEDEEB012597727CFEEB103FB2137BCA554801F300653E284C9AD069F66D5 |
| source/fixtures.json | 5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D |
| source/README.md | F6AE08D3B1AD3DC6AB621C177517178D64BB060D60814F4AB9D9BDFB40AEC3FE |
| packet/review-completion.json | 89D0942C61496E5AA4FC449B88037B4EBE8B5F6BC6D9A686BAD313FF557DC5D8 |
| roadmap/2026-10-18-p2b-sdk-structural-compatibility-review.md | 4D5CB5969441791C4292A87AA8966103D068E27EF2CCCF0BA0128B40ABEC87FD |

998 mapped pins =981 inherited+four prior diagnostic packet files+one redundant
previous live module+two readonly research artifacts+nine new immutable inputs+
one pre-repair baseline snapshot. Engine and SDK also pinned separately:
engine362A356C...,SDK2D8AF026.... All998 rehashed after the three targets,
zero mismatch;all nine raw/trace hashes matched the invocation index.
The frozen source/oracle/controller/helper were not changed after Prepare,
and module identity remained unchanged after SealProduct.

Controller preconsumes,orders exactly red/focused/full,refuses replay/retry,
checks prior exact outcomes/raw/trace hashes and all VerificationPath mappings,
ordinary parent,engine+SDK hashes,only owned child,60000ms execution+2000ms
owned exit+2000ms concurrent raw drain,separate stdout/stderr. Comparison
normalizes CRLF and removes at most one final terminator,not arbitrary trailing
blank lines. No old capture/controller/failure was retroactively repaired.

## Independent review and disposition

The single user-approved fresh reviewer was dispatched only after all three
exact outcomes and a saved preconsumed review-dispatch.json. Review is limited
to current source/tests/oracle/fixture/controller,saved evidence,spec/contract
and readonly SDK metadata;inherited executable source hash-only,no imports/
runs/edits/compile/network/security or nested agents.

[Verbatim reviewer response](2026-10-18-p2b-sdk-structural-compatibility-review.md)
preserved. Verdict: LIMITED OFFLINE NEXT DECISION with Minor limitation;
Critical0,Important0,Minor1. Review completed without execution or writes.
The earlier root-descriptor Important finding is addressed in this candidate
for the demonstrated vectors and reviewed path,not retroactively in frozen
historical modules and not as exhaustive passivity or formal acceptance.

M1 accepted and OPEN: control-frame.ps1:33-36 still checks scalar/byte-array
decorations via the public PSMemberSet.Members view. Full tests:171-188 cover
retained visible decorations,not hidden or subclass decorations on scalars/
byte arrays. Main agent verified these exact source lines. This is a coverage
gap,NOT a verified accepted-input bypass or callback activation. No code or
frozen test altered,no additional probe. Do not broaden counters/source review
into an exhaustive scalar-passivity guarantee. Historical Minor comparator
finding remains deferred on old controllers;this controller uses the stricter
one-final-terminator comparison and does not change old records.

Executor rulings on every declined-to-judge boundary:
- Hidden/subclass scalar decorations: retain as the explicit M1 gap;recommend
  a separately bounded targeted diagnostic before broader offline acceptance.
- Universal metadata-getter nonexecution: accept limitation;CounterNote-only
  measurement plus reviewed source is not universal dynamic instrumentation.
- Arbitrary private reflection/unsafe/concurrent/hostile type-data: excluded
  by the approved structural contract;no security-isolation claim.
- Other versions/PS5/platforms: excluded;only pinned ordinaryPS7.6.5Corex64.
- Catastrophic recovery/unsupported-SDK setup/internal-error fault injection:
  unproven;no claim or extra execution to close them.
- Factory provenance,real identity/authentication/capability/clock/process/Job/
  release/queue/capture/cleanup/P2-C/P3/security/live operation: outside scope;
  separate designs and permissions remain necessary.
- Historical executable correctness: hash preservation only,no reacceptance.
- Independent runtime recreation: prohibited;runtime claims stay tied to saved
  receipts,not a new reviewer test run.

Handoff formatting check flags the review transcript's intentional Markdown
two-space hard break at review.md:24 as trailing whitespace. Preserve that
verbatim reviewer text and its recorded hash;do not silently normalize it.
Other staged files pass the scoped whitespace check. This is a documentation
formatting warning,not a failed target run,and no whole-diff whitespace-clean
claim is made.

No replacement reviewer,automatic fix,retest,acceptance or main merge follows
this verdict. The recommended targeted gap investigation is NOT approved or
run here;current3+1 allowance is fully spent and must not be replayed.

## Skills, budgets and handoff boundary

Existing linked worktree reused. Systematic debugging grounded this change
in earlier invalid-input/extra-metadata observations and SDK metadata.
TDD required the exact compatibility RED before implementation.
Verification-before-completion limits claims to saved exact target outputs;
requesting-code-review supplies one fresh seat only after the approved gate.
Explicit finite scope overrides generic extra baseline/global-suite/retest/
cleanup loops;all historical allowances stay spent/frozen as recorded.

New hosts3/3SPENT;review1/1SPENT/completed. Prior repair3/3spent/review0/1ineligible,
sequence diagnostic1/1spent and every earlier consumed/frozen scope unchanged.
No resets/transfers/replays. No formal P2-B acceptance or main merge. P2-A's
limited offline acceptance remains unchanged. No PS5/P2-C/P3/full C12/real
identity/capability/session lifecycle/security/runtime/release grant.

Necessary portable source/tests/fixtures/report/review/status/recovery are
preserved on the existing devbranch under standing authorization. Fresh recovery
uses that branch,not main. Clone does not restore ignored raw packets,SDK/cache/
mapping/session state/dependencies or grant replay/reset rights. Read the final
review disposition before selecting a subsequent bounded decision.
