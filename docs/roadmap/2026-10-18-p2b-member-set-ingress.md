# P2-B member-set ingress — INCOMPLETE_STOPPED at member inventory gate

2026-10-18;base90e4dae78af5072bd841f141124472855e6a8ee4.
Branch codex/c12-b01-task9-coordinator. User explicitly approved the revised
bounded3NEW-host+1conditional-review route,unchanged counter-only in-memory
managed fixture,old scopes untouched. First host unexpectedly failed its
RawIngress assertion;remaining two hosts frozen,review ineligible. No product
repair was implemented. This is an unfinished development checkpoint.

## Actual work and outcome

New portable leaf:docs/diagnostics/isolated-fuzz/p2b-member-set-ingress/,
ten files. Old product copied unchanged. New test helper gets psextended,
requires exact PSMemberSet identity,uses fixed SDK internalMembers field,
requires exact field collection type before enumeration. Original counter
fixture/literal frames/oracle retained byte-identically. Focused/full test
drivers save a rich sidecar trace (types/stage/counts/exceptions) without
changing the expected stdout. No frozen historical helper/source was edited.

| Slot | Actual | Disposition |
| --- | --- | --- |
| red/probe | PID33684,1394ms,exit1,4lines,RawIngress assertion | Unexpected;STOP |
| focused GREEN | Not launched | FROZEN |
| full33group GREEN | Not launched | FROZEN |
| conditional readonly review | Not dispatched | INELIGIBLE |

All ordinary Windows PS7.6.5Corex64,admin=False,no timeout,stderr0,owned
exit/complete streams,pins intact. ExactLines=false,Outcome=INCOMPLETE,
FormalAcceptance=false. Expected red had6lines ending in Passivity assertion,
not4lines ending in RawIngress. Controller parent exception is not child stderr.
No subsequent test/import/alternative route/repair/seal/reviewer or retry.

Actual stdout:

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2BDescriptorFixtureObservable
ASSERTION FAILED: P2BDescriptorRawIngress
```

Saved red.trace.json establishes why this differs from the earlier receiver
TargetException,even though the group-only stdout bytes happen to match:

```text
Stage: RawReturned
SetType: System.Management.Automation.PSMemberSet
RawType: System.Management.Automation.PSMemberInfoInternalCollection<PSMemberInfo>
RetainedCount: 14
SubtypeCount: 1
RawCallbacks: 0
OriginalCallbacks: null
MemberReads / ValueReads / Copies / Writes / Total: 0 / 0 / 0 / 0 / 0
Error: Descriptor assertion
Location: Invoke-P2BDescriptorRegression,descriptor-fixture.ps1:63
```

The corrected metadata route returned a collection with the expected CLR
identity. Enumeration retained14members,including one actual CounterNote,
with zero measured fixture callbacks. The assertion requires12members,one
subtype and zero callbacks,so the observed count14 explains this stop.
This is NOT a reflection target mismatch or observed callback during this
raw access. The trace does not name the other members;the identities/reasons
for the additional two members are UNDETERMINED. Do not guess they are safe
SDK members,ignore them,or simply change12→14 to obtain a green result.

Original encoder callback reproduction did NOT execute (OriginalCallbacks
null). Product remains FA6047A4 baseline;Important descriptor-passivity finding
is unresolved. Zero counters for this controlled raw-path probe do not prove
complete ingress semantics for all inputs. The full33group test file has not
run;no new full-suite or acceptance claim.

## Frozen identities

| Artifact | SHA256 |
| --- | --- |
| Baseline module/current candidate copy | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| descriptor.tests.ps1 | DEE0B50D6FF9FB46FCFBFB55C3858163B1C42C3F68785C07010D9DD34C5A2884 |
| descriptor-fixture.ps1 | E0688991C897EDE6FA7F17D6B47502FC2584BF56E6800120CE0D2EA024F10FC3 |
| Full tests,NOT run | 7CBC263600948F9806D8749B43D0B9DB55C100F48757B1C5CC49CF94A92C60B1 |
| capture.ps1 | 5974FA239082A91F3D38A711F12D9F482E9DC044552DF16B35651BCC2A4133A2 |
| frozen.json | E7ACC64C066CC2C20682563FAD7725B14FB45058A965B910877565F1C3801949 |
| invocations.json | 219DB3E05EC3CB856FB841C2C67891283D96FAD3FA7844CBE018211E2EDBB117 |
| red.stdout.bin | B966DD333E8CB67729822924CE89D7B3FC3D75418157C709AD58B50B1F0CC074 |
| red.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| red.trace.json | AB274746C056D3E1A6B5F60EA799BC2062967DE8C5C7E279C1FC15B39A9867A6 |

Ignored packet:.superpowers/sdd/2026-10-18-p2b-member-set-ingress/.
939mapped pins=924previous+4previous diagnostic packet files+1previous live
module+9new immutable inputs+1baseline snapshot. All939 rehashed after stop:
zero drift. Raw/index/trace and every original failure/source/oracle/review
remain unchanged;none deleted or committed. New target compiles unchanged
managed fixture in memory only,no installed dependency or output DLL.
Controller preconsumes,strict ordered max3/previous exact-result gates,
ordinary parent/pinned engine,60000ms execution+2000ms owned exit+2000ms
concurrent separate raw drain,only owned child,raw and trace hashes retained.

## Scope rulings and recovery

TDD/systematic-debugging required the expected RED and complete safe-ingress
precondition before product implementation. Raw return and zero callbacks
alone do not satisfy the descriptor inventory gate. No arbitrary expectation
change or next-slot investigation is authorized. Existing linked worktree
reused;explicit finite scope overrides generic extra baseline/global tests/
fix/retest/review/cleanup loops;PowerShell receipts replace Bash scripts.
Rich sidecar evidence prevents having to infer this failure from group-only
stdout;it does not retrospectively alter the old records or their causes.

New route1/3consumed+2FROZEN,review0/1INELIGIBLE. Previous descriptor route
1/3+2frozen/0review remains;RawIngress diagnostic1/1spent. All older scopes,
including original2/5+3frozen,other diagnostics,full and negative-control/review
allowances,retain their recorded consumed/frozen states. No reset/substitution.

Next recommendation:one separately bounded identity-inventory diagnostic
recording each actual member's fixed-SDK name/type/flags without reading
caller values or invoking overridable getters. Establish the extra members'
identities before deciding a candidate rule. No such diagnostic or product
patch/reviewer is authorized or executed by this report.

Necessary candidate/tests/trace-capable controller/report/status/recovery
preserved on existing devbranch under standing authorization,not main. Clone
restores portable files/report,not ignored raw/trace/cache/mappings/sessions
or replay rights. No acceptance,claimed product fix or main merge. P2-A
limited acceptance unchanged;no PS5/P2-C/P3/C12/live/security/runtime/real
identity/capability/process/Job/clock/queue/release grant.
