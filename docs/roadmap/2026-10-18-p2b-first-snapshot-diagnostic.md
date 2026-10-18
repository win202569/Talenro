# First snapshot diagnostic — captured, product NOT repaired

Date2026-10-18;base27e6dc45aeb853a989aa919c223b9420e8510153.
User explicitly approved one extra diagnostic host,maximum120000ms,only actual
first-snapshot return/internal exception,no product fix or full-suite rerun.

## Execution and scope

New leaf p2b-first-snapshot-diagnostic has four frozen inputs;1289mapped pins =
inherited1262 +11previous packet files +11previous portable files +1prior report
+4diagnostic inputs. All pins and fixed candidate43067AD0/codec9A532E17 unchanged.

One ordinaryPS7.6.5Corex64 host:PID33584,1060ms,exit0,DIAGNOSTIC_CAPTURED.
No timeout,owned exit/streams complete,empty stderr,all1289pins intact.
This means diagnostic capture completed,NOT that the inbox passed its tests.

Target imported the fixed unchanged module,created one Inbox from the same
six-field Binding,then applied original BindingIsolation mutations:
digest[0]=99,session[0]=99,frequency=1001. Exactly one public snapshot call.
An AST-selected temporary line breakpoint at the snapshot catch(line258)
recorded the existing ErrorRecord/exception chain without replacing code,
return values or private helpers. CatchHits1;observation errornull;breakpoint
removed in finally. No second snapshot/helper retry,compilation,full suite,
product change,system/security query or independent reviewer.

## Observed failure and source localization

Creation returned Accepted=true/Reason=created with a token.
The first snapshot returned the complete conservative result:
Accepted=false,Reason=internal-error,Inbox/EndpointId/Frame/Snapshot=null,
AcceptedBytes Int32=0,and allthree authorityflagsfalse.

Captured inner exception:
System.Management.Automation.RuntimeException: Unexpected scalar.
Captured original stack:
Copy-P2BCScalar(control-frame.ps1:146) <-
Copy-IBTree(inbox.psm1:21) <- New-IBResult(inbox.psm1:29) <-
Get-P2BInboxSnapshot(inbox.psm1:256).

The inbox tree-copy dispatcher reaches its scalar fallback when preparing the
snapshot result. The codec scalar helper correctly rejects objects outside its
specified scalar types. This establishes an inbox integration/dispatch failure,
not a reason to relax the accepted codec scalar guard or a proven Binding aliasing
defect. No recursive tree-copy frame appears between the result builder and the
scalar fallback in this captured stack.

Source review identifies the relevant boundary:Copy-IBTree compares the
reference-preserving raw CLR type with PSCustomObject/object[] while its input
can be a PowerShell object wrapper representing the constructed snapshot.
The concrete runtime CLR type and wrapper/base identities were NOT recorded by
this diagnostic. Wrapper classification is a source-supported repair direction,
not an additional measured fact or blanket resolution of all snapshot cases.
No claim is made about later members,arrays,ownership isolation or all32groups.

## Fixed identities

| Artifact | SHA256 |
| --- | --- |
| Frozen manifest | 4A398AE49E88206C4D0CD9D24C27EA42999DC9D64CC84A66E1D773085CFD391B |
| Invocation receipt | 6183737559F63F52D6A51755D643473DF3B17323A55C548543812CD8FC236492 |
| Diagnostic target | C108CD3F8622280D8B5FA0AB762E615527B2EE1F648A419103340984652C06B0 |
| Capture controller | EF9EB30BA9D1F32C2AA78FF9BD6E0662B5D61C7E662283E878A0424DAE94F5E9 |
| stdout JSON | 36FCF7711AAC0969971410241A638C6999CF4DF301CE351859DD3839F4C7F379 |
| stderr empty | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Unchanged inbox candidate | 43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8 |
| Unchanged codec | 9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C |

Portable-summary.json preserves full actual result,exception/stack,receipt and
source hashes. Ignored packet remains
.superpowers/sdd/2026-10-18-p2b-first-snapshot-diagnostic/.
No raw records or previous failure verdicts were overwritten.

## Budget and next boundary

Separate diagnostic1/1SPENT. Original implementation2/5SPENT+3FROZEN,
review0/1ineligible;all older budgets unchanged. No extra run or automatic repair.
Current product is still UNFINISHED and not accepted. This record is not main,
runtime,security,completeP2-B/fullC12 acceptance.

Recommended repair scope:correct only the inbox's own tree-copy type dispatch
and independently owned recursive snapshots,retain scalar rejection/passivity/
array ownership,do not modify codec9A532E17. First write an explicit bounded
repair/verification plan covering this reproduced failure and later existing
suite/mutations;the three old slots cannot silently become a new repair cycle.
This recommendation authorizes no implementation or execution.

Save four diagnostic inputs,portable result,this report and current plan/status/
recovery via ordinary devbranch commit/push;confirm remote equality and clean.
Clone restores portable evidence,not tokens/CLR references,raw packets,SDK/cache/
credentials/debugger state or execution rights.
