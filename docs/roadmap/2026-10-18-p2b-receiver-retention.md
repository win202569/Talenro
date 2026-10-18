# Existing wrapped receiver and descriptor retention — verified within scope

Date:2026-10-18. Base:65bc59e9095091a9df54350a9b900efd21fedd16.
User approved new test files,one ordinary pinned PS7 host <=60000ms and one
in-process compilation of the unchanged CounterNote fixture. Product APIs/
imports/edits,full regression and old frozen slots were not authorized.

## Result

RETENTION_VERIFIED. PID49044,2760ms,exit0. No timeout or stderr;owned exit and
concurrent separate drains complete.1134 mapped pins intact after capture.
New1/1host SPENT,no retry/reviewer. Older latest repair1/3spent+2FROZEN and
review0/1ineligible remain unchanged,as do all earlier stopped scopes.

45/45cases passed:nine declared types(byte,uint16,uint32,uint64,int32,int64,
bool,string,byte-array) x clean,visible-note,hidden-note,visible-subclass,
hidden-subclass.9clean,36decorated,18hidden.486actual-reference IDs recorded.
All45sources were actual PSObject wrappers BEFORE capture;all45captures retain
that wrapper and all45relays retain the containing handle. Unlike the preceding
diagnostic,this does not replace these inputs with newly created clean wrappers.

Original,postcapture and postrelay snapshots retain wrapper,base,key,table and
ALL raw descriptor references.36decorated cases preserve the exact AttachedProbe
descriptor,class,name,hidden/reserved/instance flags across all3snapshots:
108attached-descriptor observations,including54hidden observations. Clean
cases contain no tag. All90capture/relay retention verdicts pass.

All45wrong-receiver controls are independently allocated equal-content
bases/wrappers with two generated SDK views. Their wrapper,base,key and table
IDs all differ from the original;all45replacement verdicts detect the mismatch.
They do not establish same-base wrong-wrapper behavior for this new helper
(the preceding separate diagnostic covers that controlled identity relation).

The unchanged C# fixture compiled ONCE in the target. A positive control
observed MemberType/Value/Copy/write counters before resetting them. All45
case counters are0before capture,after relay and after negative control.
This is instrumentation of that fixture,not universal arbitrary-getter proof.

## New test integration

New receiver-fixture.ps1 defines a test-only capture/relay/observation path.
A containing object[] carries an EXISTING actual PSObject,not the underlying
scalar through repeated member-view adaptation. The capture helper refuses
unwrapped inputs. SDK AsPSObject returns the original wrapper,then exact fields
supply base/table. The helper never strips/rebuilds descriptors.

Before capture,fixture-only SDK factories materialize psobject/psextended;
then exact raw Add inserts a literal note or unchanged CounterNote. Hidden
is set before insertion. No integrating Copy path is used. Fixed-field snapshots
capture all operands BEFORE retention assertions;no tag Value/MemberType/Copy
or coercion is used for observation or validation. Original SDK views plus
tag are compared by reference,not by content.

Reference registry uses fixed CLR ReferenceEqualityComparer and reflected
object[]argument arrays for actual PSObject keys. IDs are assigned ordinals,
not overridable hashes or serialized live references. The paired controlled
wrong-source cases make the check detect real receiver substitution.

This verifies the new test helper,not the old scalar proof suite or product.
No public scalar parameter/note-property retrieval path,local-storage/nesting,
all5optional SDK views,product-owned output copies or original33-group regression
ran. The new helper is ready for a separately bounded integration into those
product tests;do not extrapolate that public APIs preserve arbitrary wrappers.

## Product and acceptance remain unchanged

No product import/API/edit. B468B564 remains unchanged and the observed hidden
scalar metadata gap M1 remains OPEN. Failed18DA24E1 remains untouched and must
not be adopted. No main merge,formalP2-B,PS5,P2-C/P3/fullC12,live,identity,
capability,security or runtime acceptance. All authority flags remain false.
No additional target run is granted by a successful test-only result.

Next define the product-test bridge and exact baseline RED gate using the
verified actual-receiver helper,then continue the type-sensitive key/table
repair under a separately approved finite scope. Do not remove owner checks
or replace reference identity with primitive equality.

## Evidence and recovery

Leaf:docs/diagnostics/isolated-fuzz/p2b-receiver-retention.
Ignored packet:.superpowers/sdd/2026-10-18-p2b-receiver-retention.
Six immutable inputs plus a postcapture DERIVED portable-summary.json.
The compact summary retains45rows,486reference IDs,90positive verdicts and
45negative verdicts. Legends define zero-based CLRTypes indices and snapshot/
verdict column order;tagID0means clean. It omits duplicated raw owner/flag
detail,not part of the frozen input manifest. Raw remains authoritative.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 9049F39F4791ED297FD65EDFF8DE4995BBFEF330F690C71A4D19482E64EFF052 |
| receipt.json | 52FCF3BD967087A06043561E9949710CBC7C5B7D7A0656FF3620BEF2BC2D5159 |
| stdout.bin (563504bytes) | 0BADB5E4C4A3540582316E3DCF600F79BFB6F8A6CC2B67677C11C1FCD1153482 |
| stderr.bin (empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| retention.tests.ps1 | D8A94EC3DBB49FF26A42EA35A450D3FEA2D2EFD7E0C5F09BB8FD9BF2A80EE487 |
| receiver-fixture.ps1 | 5E27DDC682DE2A8C512B0ECA1D616555F7CCC762CEA18F793898C75B9F0E0217 |
| capture.ps1 | 1FB0A410592FF9D32F19F5EAFAC827745E1E551724DFDE95CAA6DCF56F81F17E |
| portable-summary.json (derived) | 8EEA4355C32A3E8D36F892FBF750CD5C982EC5345CD0D8210013BFCCDD1F668C |
| unchanged descriptor-fixture.cs | 87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE |

1134pins=inherited1121+priorpacket4+priorreport/summary2+sixinputs+livebaseline1.
Execution/owned-exit/drain60000/2000/2000ms,only owned target terminable. Static
parse passed before the only launch. A postcapture verbose JSON extraction
exceeded output budget;compact read-only extraction corrected that analysis,
with no target rerun or raw/receipt/source changes.

Necessary seven portable files,report,current-status and recovery preserve this
checkpoint on the existing devbranch,not main. Clone restores source/tests/
portable IDs/reports,not ignored raw,actual CLR references,SDK/runtime/cache/
sessions/mappings,and never resets budgets or grants replay/acceptance.
