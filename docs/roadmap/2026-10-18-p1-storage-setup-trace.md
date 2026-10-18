# Synthetic storage setup trace — assignment/reread loses wrapper storage

Date: 2026-10-18. Base f1779db788eb1b3a4a7f1efd4b34e1e4f35e9dce.
Branch: codex/c12-b01-task9-coordinator.

## Outcome and scope

The user continued the proposed synthetic-only setup diagnosis. One ordinary-user
PS7 child completed two fixed Int32 diagnostic paths and 14 trace records, exit0
in836ms. Limit10,000ms plus2,000ms owned exit and2,000ms stream drain. No timeout,
owned exit/streams complete, stderr0, no controller error. All637 frozen file
pins remain intact on an independent read-only rehash. Candidate imports/calls0.
No observer scan, actual ownership suite, failed-batch rerun, fixture fix or
old-slot reuse occurred. FormalAcceptance remains false.

The result identifies an observable loss boundary: wrapper construction and
R-Wrapper return produce native System.Management.Automation.PSObject, but reading
the existing note property's Value after assignment reports System.Int32 again.
The exact setup path independently reports unchanged Int32 storage, unchanged
descriptor identity and unchanged base identity. Thus the earlier fixture's
"Storage fixture did not change type" guard correctly rejected a non-mutation.

This is not proof of which internal PowerShell binder/setter/getter step performs
the conversion. The split path adds locals and type observations, so it cannot
prove the precise behavior of every invisible temporary in the original RHS.
The evidence does rule out a failure to create/return a wrapper in the observed
split path, and does not support blaming Fresh-Value or Owner for an earlier
storage-type change: their observed values all remain Int32.

## Recorded trace

NativeType below uses the SDK reflection/trampoline R-Type helper, not ordinary
PowerShell GetType member adaptation. PSObject abbreviates
System.Management.Automation.PSObject; Int32 abbreviates System.Int32.

| Sequence | Path/stage | NativeType |
| --- | --- | --- |
| 1 | exact / before-assignment | Int32 |
| 2 | exact / after-assignment | Int32 |
| 3 | split / raw-array-box | Int32 |
| 4 | split / fresh-helper-return | Int32 |
| 5 | split / nested-fresh-return | Int32 |
| 6 | split / owner-parameter | Int32 |
| 7 | split / owner-slot-inside | Int32 |
| 8 | split / owner-slot-returned | Int32 |
| 9 | split / after-leaf-capture | Int32 |
| 10 | split / constructor-argument | Int32 |
| 11 | split / direct-constructor-result | PSObject |
| 12 | split / original-wrapper-return | PSObject |
| 13 | split / slot-after-local-assignment | Int32 |
| 14 | split / direct-slot-control | Int32 |

Exact path: captured stored/base types Int32; after-assignment StorageChanged=false,
DescriptorSame=true, BaseSame=true, BaseType=Int32.
Split path: CapturedStored=Int32, constructor ArgumentCount=1, final StorageChanged=false.

The saved text was independently checked read-only: 15 lines, 14 JSON trace
records, sequences1..14, exactly2 exact and12 split records, and exactly the final
RESULT: STORAGE_SETUP_TRACE paths=2 records=14 productCalls=0.
That RESULT means only trace completion, not equivalence, fixture validity,
performance improvement, product correctness, security/runtime acceptance or C12 completion.
The previous45-case batch still has only3 completed paired cases.

## Implementation boundary

New isolated storage-setup-diagnostic leaf contains diagnostic.ps1, capture.ps1,
README, source-pins JSON and attributes. The diagnostic parses the original test
as DATA, extracts exactly eight allowlisted helper definitions into a fresh module,
and never runs test top-level code, V-Others, V-Invariant or candidate functions.
Two synthetic Int32 paths only; trace payloads are type names/booleans/integers.
The original fixture, both old controllers, optimized test copy, candidate,
inventories and all prior receipts/reports are unchanged.

Controller inherits627 pins and five anchors from the failed equivalence batch,
then pins all5 new files (637 total), and consumes its one new slot before launch.
Ten fixed hash literals were checked against actual files before launch. Both
new scripts have zero parser errors. No product implementation or full-suite
verification was attempted within this narrowly scoped diagnosis.

## Evidence hashes

Ignored local packet: .superpowers/sdd/2026-10-18-p1-storage-setup-diagnostic
StartUTC2026-10-18T13:25:53.2537246Z, owned childPID45852.

| File | SHA256 |
| --- | --- |
| frozen.json | B036DA584F00F9A3105771595A61F2333A61F1B9C8528098B78A3D04D6B2A68C |
| invocation.json | 46E0D238E795A738C6115C3490264C4330C952A8E5218AB6D1C0B3DB27DB4F25 |
| stdout.txt | D1019FB4721E7516E38B7DAA786A78D759C3B9FFD3392D7DE6823011A63988AF |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| capture.ps1 | FF4A7D3836FDCE81F4FC480F5225BF3F083F1F1BBFD487D1A7C03BA5341B7F60 |
| diagnostic.ps1 | DC93A10D32A339F43A8698C0B3E45D6A374045619F0D79078D96A051C0F218D8 |
| README.md | CC4BE4F2FA0413DAF1ACB1F5D4568B10681D989CB5C80C222195C8B3805F1A3C |
| source-pins.json | 4517E49DFFA5623511D4FAAE6178CC4B1B466560444B707D5E0EAC02CB15ABB7 |
| .gitattributes | 0CD209979B0AD7DB7DDC9CB658156B0A9C18859999F71E541861F3AB58E7AB11 |

Raw packets stay ignored and untouched. Portable source plus this complete trace
table preserve the result for a fresh PC; cloning cannot restore local raw receipts,
private cache, engine path or execution authority.

## Next bounded proposal

Preserve all existing failed/successful diagnostic inputs. Explore a separate
synthetic-only fixture construction that genuinely changes stored native type
without replacing the descriptor, base identity or weakening any existing checks.
A possible mechanism to evaluate is the public SDK property setter via the same
reflection trampoline, with fresh object-array arguments rather than PowerShell's
ordinary property assignment. This is an untested proposal, not a proven fix.
First require direct native stored-type and unchanged-descriptor/base evidence,
then build a new equivalence harness copy without removing storage-type coverage.
All five storage kinds and both directions need valid setup, not assumed symmetry.

No further execution, fixture repair, candidate matrix, timeout increase or
acceptance is included in this completed diagnostic allowance. A future focused
construction check/new equivalence batch needs its own agreed bound. Continue
from this development branch and latest status, not main.
