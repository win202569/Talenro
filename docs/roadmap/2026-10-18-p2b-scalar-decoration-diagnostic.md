# P2-B scalar decoration diagnostic — INCOMPLETE_STOPPED at fixture identity

2026-10-18;base5313b008d8475926786b6bb36099d70aff5d12a6.
Branch codex/c12-b01-task9-coordinator. User asked to continue the recommended
limited verification of the saved Minor scalar-decoration coverage gap.
One NEW ordinary target host1/1spent;no product fix/full-suite replay/reviewer.
The original B468B564 candidate remains unchanged. This is an unfinished
diagnostic checkpoint,not regression failure proof or offline acceptance.

## Actual execution

PID20968,1585ms,ordinary Windows PS7.6.5 Core x64,admin=False,exit0,
no timeout,stderr0,owned exit/streams complete,pins intact.
Controller Outcome=INCOMPLETE,FormalAcceptance=false.
The diagnostic catches its assertion and emits saved JSON,so target exit0
does not mean that the35-vector matrix completed.

Stage=encode-scalar:visible-note. Seven clean controls completed:
encode scalar/array→encoded;binding Message scalar/array and Expected
scalar/array→binding-matched;decode array→decoded. All seven have zero
CounterNote counters before/after API/after inventory,retained base/root
descriptor identities,no ordinary extension,all three authority flagsfalse.

The eighth vector failed its fixture precondition at diagnostic.ps1:77,
through data-helpers.ps1:7. Product call for that vector is at line80,
so NO decorated input was submitted. Visible/hidden/subclass decorated
outcomes are UNDETERMINED. All28 decoration vectors remain unreached;
do not describe this as hidden-input acceptance or callback activation.

The composite precondition requires exactly one nonreserved entry,one tag,
same descriptor reference,exact class,false reserved flag and expected hidden
flag. The failed vector's local pre-snapshot is NOT serialized:rows are
appended only at85-86,after the API. Thus the saved result identifies the
failed composite assertion but not every predicate's observed value.
Do not invent missing raw counts/flags/reference observations.

## Read-only cause investigation, no rerun

The new driver inserts the original extension at line68,then compares the
raw retained entry against that SAME original reference in SP-Snapshot,
and requires SameDescriptor at77. That expectation conflicts with the SDK's
insertion-copy behavior.

[Official v7.6.5 source](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4272-L4289)
shows AddToTypesXmlCache creates memberToBeAdded via member.Copy().
The integrating Add overloads delegate to that path. Exact PSNoteProperty.Copy
constructs a new descriptor. Therefore,when the tagged entry is found,its
reference is not the original object passed to Add. This is a statically
confirmed erroneous fixture identity expectation,sufficient to explain the
stop;the unsaved pre-snapshot prevents claiming it was the only false
predicate in this specific vector.

Read-only local SDK method-body metadata corroborates the copy path. SDK
SHA2562D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B.
These method bodies were READ,not invoked;no extra target/fixture/product
run or managed compile followed the failed diagnostic.

| Method | Metadata token | IL SHA256 |
| --- | --- | --- |
| Void AddToReservedMemberSet(T, Boolean) | 06001E96 | A9013301AFE26DF63F0953F76D807904E4A109ED2332BBE7DEDFB4B51CB5EF15 |
| Void AddToTypesXmlCache(T, Boolean) | 06001E97 | 1BD2EE2E53F9FF3D0CD1F3EF5C2C77A541AD8FCBBB343DD945188D6AFB38D2F4 |
| System.Management.Automation.PSMemberInfo Copy() | 06001DAF | 2C1CC8CE42610CBDBBD2707CFF6606041F6E8542D1AB427EAD3814EBC2E17C08 |

AddToTypesXmlCache bytes contain the callvirt to PSMemberInfo.Copy token
06001D66;PSNoteProperty.Copy contains construction of a new PSNoteProperty.
Inspection of the outer Add overloads alone did not contain a direct Copy
call;the delegation path,not a naive outer-method byte search,is the evidence.

A further source constraint for any future fixture correction:unchanged
CounterNote.Copy returns a new CounterNote(Name,stored) and does not clone
the hidden flag. Setting hidden on the original before public Add is therefore
not enough to prove a retained hidden CounterNote. No hidden-subclass vector
ran,so this is a design limitation inferred from fixture source,not an observed
product result. A new fixture must establish actual retained class/flags/
identity before measuring product calls and save those preconditions before
asserting them. Do not relax them or mutate the old failed records to obtain
a green diagnosis.

## Preparation event and preservation

An earlier controller start stopped BEFORE packet creation/target allocation:
a mechanical source adaptation had corrupted the expected old invocation
hash suffix from ...3198A1981 to ...3198A11020. Read-only Get-FileHash matched
the original76066E6C...1981 report identity and Test-Path showed no new packet.
Only that unfrozen new controller literal was corrected before freezing.
No old record was changed,no target launched at that preflight,no spent
allowance reset. Then exactly one target was allocated/compiled/run.
Both events are recorded here;do not describe two target hosts or hide the
preflight correction as genuine historical hash drift.

Five-file portable leaf:
docs/diagnostics/isolated-fuzz/p2b-scalar-decoration-diagnostic/.
Unchanged counter C# was compiled once in memory in this one target;direct
observable control passed. Definitions-only data helpers are byte-identical
to the previous diagnostic helper excerpt. No mock,product-generated oracle,
installed dependency or output DLL. Product imported/called only for seven
clean controls,not decorated cases. No new independent reviewer.

Ignored packet:.superpowers/sdd/2026-10-18-p2b-scalar-decoration-diagnostic/.
1020 mapped pins=998 inherited+15 previous packet files+unchanged live
B468B564 module+unchanged review report+five new inputs.
Read-only reconciliation after stop:1020 checked,zero mismatch;saved raw
stream hashes match receipt. All historical source/oracles/raw/trace/review/
budget records unchanged;no frozen input edited after target launch.

| Artifact | SHA256 |
| --- | --- |
| packet/frozen.json | 7EDB4EFB27D168B544268904645B5C7736DF64C91B7D1B780C87DF03B05C2D76 |
| packet/receipt.json | F9CA78FA3F78BEFFCB7B446613386AE454F99C3B9F226819CEEC08FD8F226024 |
| packet/stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/stdout.bin | 938C077784C33A2D9B56CA234292E809669E599E25CF1093A5DF1638CD2FE5E5 |
| source/.gitattributes | C1D75387C71316ADD449DCD42424BD5D6076CA37442786E4A9206EC60245F425 |
| source/capture.ps1 | 03C539D60C5A418BE4568F9E75BCF66903F62E7FFDA44F9FF53CB8C71E016FB7 |
| source/data-helpers.ps1 | 8FE8F3D323A3F9B5DE04AEAFC0A86CB900D92A26BD6B58F3B9DCC39C79314BB2 |
| source/diagnostic.ps1 | 03F9E8B80F8D8ED9566F4B2BC7C4936176E08BB272C535FCC575D838CFE77955 |
| source/README.md | 6C2A30B3AD9CA0FFBB23121785617B674141FE41867267C3E615032E303E47D2 |

Controller preconsumes/refuses existing packet,pins engine AND SDK,checks
prior exact3host/review receipts,only its owned child,separate concurrent
raw streams;10000ms execution+2000ms owned exit+2000ms drain.
The parent incomplete exception is not child stderr. No wider process scan
or weakened safety gate. Packet/raw/cache/binary/credentials are not in Git.

## Disposition and next decision

Systematic-debugging kept the investigation evidence-first and stopped before
any product correction. Verification-before-completion forbids claiming
coverage closed from seven clean controls. Diagnostic1/1SPENT,no retry.
Previous33-group GREEN and readonly review retain their original scoped
evidence;this fixture failure neither overturns them nor closes M1.
The earlier Minor remains OPEN;no proven new product bypass/callback finding.

Recommendation:prepare a NEW diagnostic candidate that tracks the actual
retained descriptor,proves hidden/subclass state without assuming public Add
preserves the original object/flags,and serializes every precondition before
checking it. Only a separately bounded new host allowance may then investigate
the decorated inputs. No such correction/probe is executed or approved here.
Keep product fixed until actual decorated-input evidence supports a change.

All old budgets remain spent/frozen,including structural3/3hosts+1/1review.
No reset/reuse/new review/full-suite rerun/acceptance/main or real capability.
Necessary failed diagnostic source/report/status/recovery preserved on existing
devbranch under standing authorization. Clone restores portable files/report,
not ignored raw/caches/SDK/sessions/mappings/dependencies or reset/replay rights.
P2-A limited acceptance unchanged;no formal P2-B/PS5/P2-C/P3/full C12/live/
identity/capability/process/Job/clock/queue/release/security/runtime grant.
