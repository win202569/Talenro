# P2-B RawIngress diagnostic — receiver mismatch identified, NOT repaired

2026-10-18;base1d183dce200ec995c19df25183e3b9bf603b8c62.
Branch codex/c12-b01-task9-coordinator. User approved the proposed single
minimal diagnostic preserving specific exception and intermediate values.
Separate allowance1NEW ordinary host,now1/1spent;no repair,reviewer,old slot
reuse or unfreezing. Original descriptor route remains1/3used+2frozen,
review0/1ineligible. Earlier consumed/frozen allowances unchanged.

## Single-host evidence

PID21488,ordinary Windows PS7.6.5Corex64,admin=False,elapsed1304ms,exit0,
no timeout,owned exit=true,streams complete=true,stderr0,pins intact.
Outcome=DIAGNOSTIC_CAPTURED;FormalAcceptance=false. Exit0 means evidence
capture completed,NOT that ingress/product validation passed.

The new driver imports only the frozen test-helper definitions,compiles the
same counter-only managed fixture in memory,checks its observable control,
constructs the same message/subtype and calls preserved DG-RawMembers once.
No product module import or public API call,no alternative raw route tested.
The original failed controller/target is not replayed or edited;this is a
separately approved evidence-only driver. Exceptions are retained,not hidden
behind a group-only assertion. No product code or test contract changed.

Key saved fields:

```text
Stage: RawFunction
FixtureObservable: true
RootBaseType: System.Management.Automation.PSCustomObject
ReceiverStorageType: System.Management.Automation.PSInternalMemberSet
RawResultType: null
DescriptorCount / CounterSubtypeCount / OriginalGateWouldPass: null
MemberReads / ValueReads / Copies / Writes / Total: 0 / 0 / 0 / 0 / 0
FullyQualifiedErrorId: TargetException
Outer exception: System.Management.Automation.MethodInvocationException
Inner exception: System.Reflection.TargetException
Message: Object type System.Management.Automation.PSObject does not match target type System.Management.Automation.PSInternalMemberSet.
Location: DG-RawMembers,descriptor-fixture.ps1:21
ProductImported: false
ProductCalled: false
RuntimeApproved: false
```

Frozen helper line20 stores `$Value.PSObject` as the receiver;line21 invokes
the fixed InstanceMembers getter declared on PSObject. The stored receiver
is actually PSInternalMemberSet,so reflection rejects the target before a
raw collection is returned. This explains the observed RawIngress failure
as a test-helper receiver mismatch,not a failed zero-callback/count assertion.
No collection enumeration occurred. Counters remain zero at this failure;
that does NOT demonstrate a usable callback-free descriptor access path.
Original Important static product passivity finding remains unresolved,and
this diagnostic still does not dynamically reproduce the product callback.

## Identity and preservation

| Artifact | SHA256 |
| --- | --- |
| Diagnostic target | 0107E80ACE1D1C0CE80A0F987C7C87D47F72A8FFFB2443AC30412600C6CE2094 |
| Diagnostic controller | 5C5690580CBFC162483E963C779DABE32A673ACBD2664D7ACB0ACA42060636C8 |
| Original helper,unchanged | E42C7FBAC8048999D3963D6215C9BA4B04323424DA99E5CED47DB8BDE0F9FFE2 |
| Counter fixture,unchanged | 87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE |
| Product baseline,unchanged | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| frozen.json | A9AE3DF83EEEA6626A4A495F722E7C2091F798B10E64BC5342D6D7667275B1F0 |
| receipt.json | FEA1184766FE8CE83F9ACC0AEF53542C23F7381619F456AE103A4CF62DB3343E |
| stdout.bin | 99AF5684A556AB6A9EE3FB1B086D30DB6B15D3AE972B3BF61DFA8A094E65FBA1 |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Ignored packet:.superpowers/sdd/2026-10-18-p2b-rawingress-diagnostic/.
924mapped pins=916previous+5previous packet files+3new diagnostic inputs.
All924 rehashed after capture with zero mismatch. Prior raw/index/source/
oracle/review records retained without rewrite or deletion. New raw stdout
contains the complete exception chain,source position and script stack;
portable report above transcribes key evidence,not a substitute for raw.
Controller preconsumes persistently,refuses replay,uses ordinary parent and
pinned engine,concurrent separate raw streams and only owned child. Bounds
10000ms execution+2000ms owned exit+2000ms drain. No new dependency install,
native compile,network/security query,product import or reviewer.

## Handoff

Systematic-debugging required explicit boundary evidence before proposing a
repair;the single diagnostic isolated target-type mismatch without changing
the original helper,testing alternate code or claiming a safe ingress route.
No further run or implementation followed this observation.

Next recommendation:correct the reflection receiver only in a new test
candidate,then re-establish observable descriptor retention and zero-callback
ingress before any product patch. Actual SDK-object acquisition must be
validated,not replaced by another guessed receiver expression. That correction
and its finite verification scope need their own approval;none executed here.
Do not spend the two frozen slots or substitute a new reviewer automatically.

Necessary driver/controller/report/status/recovery preserved on existing
development branch,not main. Fresh clone restores portable files/report,not
ignored raw/cache/mappings/sessions,run rights or allowance resets. Product
remains baseline;no acceptance,claimed fix or main merge. P2-A limited
acceptance unchanged;no real identity/capability/process/Job/clock/queue/
release,PS5/P2-C/P3/C12/live/security/runtime grant.
