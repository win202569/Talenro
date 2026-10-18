# P2-A InputMutation — one diagnostic captured the failing site

## Approval, consumption and outcome

After f5549619d34eadcdfc580d5587d01bf9c55145c3, user explicitly approved ONE
NEW ordinary diagnostic run, unchanged product/assertions, failure-site/stack
instrumentation only. Approved diagnostic plan physical SHA256 before status edit:
C8C79512687A0082C293A5E038DA0AED209D1984CFF400C327F559AEDBC51BAF.

**REPRODUCED_SITE_CAPTURED. Diagnostic allowance 1/1 spent; no retry/fix/reviewer.**
This is diagnosis, NOT green tests, a repair, acceptance or authority to resume
the old route. Old P2-A allowance stays2/5consumed with3frozen,review0/1ineligible.
P1/old records/limits unchanged. No main/B/C/P3/PS5/Docker/live/safety queries.

Ordinary inherited PS7.6.5 Core x64,admin=False; child exit1,1183ms,22ordered lines,
stderr0,no timeout. Owned host exited, separate asynchronous raw-byte streams
drained,controller error null;790checks intact during capture and readonly
recheck. Preconsumption was persisted before the single launch. No extra product
import, probe, host, mutant, fix or review after the capture.

## What the evidence establishes

Assertion ordinal126 failed in TB-Need (diagnostic line30, original10), called
from TB-Leaf diagnostic78/original55:

```powershell
foreach($p in $members){TB-Need ($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty)}
```

Diagnostic stage=leaf-initially-plain,Field=Schema,OtherField='',OtherIndex=0.
RuntimeException,ErrorId=Assertion,Message=Assertion. Stack then reaches TB-Leaves
diagnostic85/original61, and diagnostic208/original160:

```powershell
$inputs=@(TB-Leaves $r)
```

It is the INPUT request's Schema view, not output A/B enumeration or an
input-mutation result comparison. The test's initial-plain-view requirement found
a NoteProperty before the group's mutation loop. The loop, root mutation,
ObservedTick replacement, saved-result comparisons and next-call comparison
were not reached. Two product calls did occur before leaf collection; their
returned Created/Reason/Plan were not logged, so no success/refusal claim about
those calls is made. The earlier18PASS prefix matches the failed original run.

## Immediate cause versus unresolved origin

Confirmed immediate failure: the fixture's Schema view was already decorated
when TB-Leaf required an initially plain view. This identifies a fixture-
precondition failure site, not evidence that the post-mutation product-output
comparison failed. It does NOT globally clear the product or prove who decorated
that particular scalar.

Source-only trace identifies a focused hypothesis: preceding DecoratedScalars
intentionally uses TB-Mutate to attach P2AMetadata descriptors; New-TBRequest
directly assigns numeric parameters/default Schema to properties rather than
constructing explicitly independent typed-array scalar storage. A retained/shared
numeric fixture view or default is therefore a candidate pollution boundary.
Product source has no explicit write to Request properties; its note-property
writes target local snapshot/owned containers. Neither observation proves exact
runtime reference identity or the original descriptor's name/payload.

No metadata-name/backing-reference snapshot was captured at this site, and no
additional characterization probe was authorized. Instrumentation may itself
affect observation. Therefore the precise propagation mechanism remains a
hypothesis, not a verified root cause or completed repair.

Recommended next decision: establish independent fresh numeric-fixture controls
and trace cross-case descriptor ownership before deciding whether to change
New-TBRequest only. Preserve wrong-type/decorated-input tests, full literal/type
checks, held-view/reread distinctions and callback-zero assertions. Do not weaken
the initially-plain check, clear metadata in-place, reorder/skip preceding tests,
or change product arithmetic/ownership to hide this result. Any source change
and bounded validation need separate approval; no extra allowance inferred here.

## Preparation and scope rulings

Reused expected linked non-submodule worktree/development branch/origin at
f5549619,initially clean. Git-only privileged readonly operations checked metadata
and LF filtering; diagnostic target/controller remained ordinary. No trust,
ownership, dependencies or security settings changed.

Exact product copy retained SHA256
350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B.
Mechanically preserved original lines1–167; static inverse reconstruction matched
all167lines after removing only approved instrumentation. Ordered19case names
unchanged. Original-prefix command ASTs identical:35TB-Need,18product,36request,
13want,5normal,1refusal,18reject. No assertion arguments/literals weakened.
Original later groups omitted by design; only diagnostic termination/catch added.

Preparation checker initially compared whole TB-Case extents, which necessarily
include the new markers. Refined that check to ordered case names plus the above
unchanged assertion/call ASTs; no source change or target launch resulted from
the checker mismatch. Physical LF/JSON/PS parse and physical/filtered Git blobs
checked for all six portable files before freeze.

790common checks=775original inherited mapped/current entries+8original packet
anchors+original product+6new diagnostic files. Original mapping unchanged.
Approved plan SHA is metadata, not a mutable live pin. Portable source-pins.json
includes full line map and exact allowlisted transformation.

Ruling: native PowerShell capture/ledger replaces Bash skill helpers, as the
approved Windows-only route requires; no generic runner-completion evidence.
Ruling: exactly one diagnostic overrides generic extra TDD/baseline/task-done/
full-project/fix/retest/review/cleanup loops; cost is no regression or acceptance
evidence. All ignored raw packets retained. No deferred review findings because
no reviewer was authorized or dispatched.

## Ordered stdout transcription

This portable transcription does not replace raw bytes or constitute a new run.
Both ordinary CONTEXT/HOST and18PASS labels match the original prefix exactly.

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
PASS: P2ANullStop
PASS: P2ANormalStop
PASS: P2ANonzeroOrigin
PASS: P2AFractionalFrequency
PASS: P2AFractionalPhaseRounding
PASS: P2ADeadlineEquality
PASS: P2AStopCutoff
PASS: P2AStageWindows
PASS: P2ALateStop
PASS: P2AOriginLimit
PASS: P2AFrequencyLimit
PASS: P2AIntermediateOverflow
PASS: P2ARequestShape
PASS: P2AExactTypes
PASS: P2APassiveDescriptors
PASS: P2ADecoratedScalars
PASS: P2ARangeRelationships
DIAG: {"Schema":1,"Outcome":"INPUT_MUTATION_EXCEPTION","Case":"InputMutation","Stage":"leaf-initially-plain","Field":"Schema","OtherField":"","OtherIndex":0,"AssertionOrdinal":126,"OriginalOperationLine":55,"ExceptionType":"System.Management.Automation.RuntimeException","ErrorId":"Assertion","Message":"Assertion","DiagnosticLine":30,"Stack":"at TB-Need, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 30\r\nat TB-Leaf, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 78\r\nat TB-Leaves, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 85\r\nat <ScriptBlock>, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 208\r\nat TB-Case, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 35\r\nat <ScriptBlock>, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-input-mutation-diagnostic\\diagnostic.tests.ps1: line 200","ProductChanged":false,"RepairPerformed":false,"Acceptance":false}
ASSERTION FAILED: P2AInputMutation
```

## Identities (SHA256)

Portable leaf: docs/diagnostics/isolated-fuzz/p2a-input-mutation-diagnostic/.

| File | SHA256 |
| --- | --- |
| .gitattributes | E3083818059D2956756D32B1D4A7098E28F57DFB3B971F429A73E51672505B9D |
| capture.ps1 | 4214DA64052E7E0D81E27349F249E96286233270CBEFC8974326CABED252C1F5 |
| diagnostic.tests.ps1 | F097FA916CBAC294AB88696FAF088236F72B51ABA359BF1DFC4580C78905B859 |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| README.md | 8ECC6FFB53C622741D61C93B22612B110D217B101B42C0C6320B2C8BCBEE9D4E |
| source-pins.json | F44EAD9948EAEE977181D4FA71898451722CDB0908C50F442C4BD5036B3D1E5C |

Ignored packet: .superpowers/sdd/2026-10-18-p2a-input-mutation-diagnostic/.

| File | Bytes | SHA256 |
| --- | ---: | --- |
| frozen.json | 345060 | AF3D93BB4C1AEA2BD788AA08E1DF2DED6798C2C063906B717C05255459C1BE99 |
| invocation.json | 2760 | 7893E8BE695854CE6DFEB6FEEEB0F8CB6726FED40F76AAEEFB4AC4B82BABBA9E |
| stdout.bin | 2013 | 1E0FC080EC3BB7D4665F949534582773429C5BFB27A14B832E508E2C3905D082 |
| stderr.bin | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Engine SHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Original product/tests/controller/oracle and all original raw records unchanged.
No diagnostic repair or independent review exists. CodeProperty/internal-error,
PS5,concurrency/type-data and real-adapter limits remain open.

## Recovery

Save six new portable files,this report,diagnostic plan status,current-status and
repository-recovery on existing dev branch. No ignored raw packets or binaries
committed/deleted. Git log/remote SHA verification establishes the eventual
checkpoint identity; this report makes no push claim before it occurs.
Fresh-machine recovery restores source/map/stop result, not ignored receipts,
historical local pins,engine cache or sessions; do not replay the spent allowance.
