# P2-B controller exception containment — unfinished execution

Date:2026-10-19. Execution base19e3880d5133ecdb75bed9981ceff25ceb0547ce.
Branch codex/c12-b01-task9-coordinator. Status:GREEN_FAILED_FROZEN_UNFINISHED.
User approved [Native plan](../superpowers/plans/2026-10-19-p2b-controller-exception-containment.md)
andNEW3hosts+5children+conditional1review after19e3880d.
This document records actual progress,not acceptance or authority to rerun.

## Scope and immutable inputs

New controller-only [leaf](../diagnostics/isolated-fuzz/p2b-controller-exception-containment/README.md).
Seven text inputs prepared before RED. New helper calls real Process/File/Task;
directory-as-receipt path causes real post-start write failure,no ACL change,
product callback/setter/import,installation or native compilation.
Old five seam inputs/summary/fullreview/dispositions andall historicalraw unchanged.

1398pins=1378inherited+oldraw8+oldhandoff4+newinputs7+approvedplanSnapshot1.
REDmanifestF9C76ECB6C78FDAE6D492A8B71626088746946D6C9E78A8BFE234263DB9C280C.
GREENmanifest9AAF1DA823DC427F0989DBAEFD802FAFE70B08653177FE93289D038ADA377248.
ApprovedPlan3680755E62407264EB4AF4204FF9EC96D06A3FAAC6E1BF352FCAC039DE9E18DA.
Only owned-run changed in the single authorized repair,DB762EE2...→7A425C66...;
other1397mapped identities identical. Baseline and GREEN source snapshots retained.
Existing ordinary parent bookkeeping sourceB2D91DB0... saved separately,no extra target.

## Actual runs

| Slot | Host | Result | Owned child |
| --- | --- | --- | --- |
| RED |PID30464,813ms|exit1,expected failure;exact receipt|PID2004,still alive before rescue;exact-owned rescue confirmed exit|
| GREEN |PID20428,824ms|exit2,INCOMPLETE;1of5groups PASS,then assertion failure|PID47688,terminated by repaired helper;exit observed,no outer rescue|
| FINAL |not launched|frozen|not launched|

Both started hosts exited with complete captured streams/no parent error/timeout
and1398postpins intact. REDstdoutED7862BF...;empty stderr.
GREENstdout9C0FFE18...;stderr2EE208A1... contains exactly
ASSERTION FAILED: PrimarySurvivesFinalReceiptFault.
The GREEN host failed;no successful GREEN or final acceptance is claimed.

GREEN first check observed Started/TerminationRequested/OwnedExit/StreamsComplete/
ActualHasExited true,TimedOutfalse,Rescues0,ExitCode-1 after termination.
Second check observed FirstRecordIdentity=false,DistinctFinalError=true,
CleanupCount1. The first ErrorRecord identity assertion failed.
Test compares returned PrimaryError with the oldest record in automatic Error
collection (containment.tests.ps1:65,81–84);helper assigns PrimaryError in catch
(owned-run.ps1:28). Source alone and saved boolean traces do NOT establish whether
the returned record was replaced or the comparator selected a different wrapper/
reference. Live CLR ErrorRecord objects were not persisted. Cause unresolved;
no correction/relabel/fault injection/probe or replacement test has been run.

Unexecuted:NormalExit,Timeout,PreStartOutputFault,FINALoriginal12groupseam,
andconditionalreview. The earlier two successful seam runs remain unchanged.

## Budget, stop and handoff

NEW2of3hosts+2of5children consumed=4of8newPS7processes;remaining1host+3children
FROZEN. Review0of1spent,conditionalseat FROZEN,not dispatched. No retry/secondfix/
automatic continuation;generic extra suite/task-done/review/deletion is overridden
by approved finite plan. This is limited source/observation evidence only.
Task1/Task2 NOT complete. Existing I1 remains open/unaccepted.
Old seam2+1/tree5+1spent;original2spent+3FROZEN/diagnostic1spent unchanged.

[Portable summary](../diagnostics/isolated-fuzz/p2b-controller-exception-containment/portable-summary.json)
preserves complete hosts/children/RED/GREEN/before-rescue JSON subtrees as raw JSON
text,stream bytes inBase64 plushashes and19artifact identities. REDbaseline/
parent bookkeeping text snapshots are handoff outputs,not extra frozen inputs.
The two readable .ps1.txt snapshots may receive CRLF on a Windows checkout;
SourceSnapshots in the LF-pinned JSON also preserves their exact original bytes
asBase64 withlength/hash. Use those bytes for identity,not checkout line endings.
Fresh readonly handoff validation:5JSON subtrees andraw numeric tokens exact;
1398GREEN pins intact,only1planned RED→GREEN source delta,all19raw hashes,
8decodedstream byte hashes and2source snapshot hashes match. FINAL absent.
This bookkeeping validation is not an additional test run or GREEN reclassification.
The current seven sources and old fixed dependencies restore from Git;raw packets,
engines,credentials,live ErrorRecords/Process objects and execution rights do not.
Do not automatically rerun Prepare or reconstruct absent raw on a new computer.

Next recommended step:separately authorize readonly source/evidence analysis of
the failed identity assertion first. Such analysis must preserve original failure
and spend no targets;any proposed comparator/helper change or new verification
needs explicit revised scope/budget before implementation. No main/live/formal/
real handshake/capability approval follows from this unfinished checkpoint.
