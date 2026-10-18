# P2-A fixture isolation repair — focused regression green, full suite stopped

## Approval and actual status

After45a32ea732ca1172580ab81c8a882464a18e2f13,user approved the bounded short
design: Native,six NEW ordered ordinary hosts and ONE NEW readonly reviewer only
if all six precise expected outcomes hold. Authorizing design physical SHA256
before status edits:0F76FE39A86F529A8B388DDD650382C2E82DF724C8E0F95EB55787ECD3CA4C0D.
No new architectural spec/implementation-plan required by the bounded path.

**STOPPED_AT_SLOT_3_FULL_INPUT_MUTATION; unfinished development checkpoint.**
Consumed3/6hosts;remaining3frozen,not automatically reusable. Review0/1ineligible.
No mutants applied,so restoration unnecessary. No fixes,retries,extra imports/
probes,reviewer or further target after this unexpected full-suite result.
Old allowances remain spent/frozen. No main/acceptance/PS5/P1/P2-B/C/P3/live grant.

| Slot | Exit | Elapsed ms | stdout lines | stderr bytes | Exact outcome |
| --- | ---: | ---: | ---: | ---: | --- |
| 1 red,old constructor | 1 | 788 | 3 | 0 | expected SchemaIndependent failure |
| 2 focused,repaired constructor | 0 | 946 | 12 | 0 | expected9groups green |
| 3 full,repaired constructor | 1 | 1183 | 21 | 0 | unexpected InputMutation failure |
| 4 phase-mutant | — | — | — | — | not started,frozen |
| 5 expiry-mutant | — | — | — | — | not started,frozen |
| 6 final full | — | — | — | — | not started,frozen |

All three targets ordinary PS7.6.5 Core x64,admin=False,no timeout,owned-host-exited,
complete separate concurrent raw-byte drain,controller errors null. All811common
checks intact after each and readonly reconciliation;raw hashes verified. Max
30000ms execution+2000ms owned exit+2000ms drain. Preconsumption before each launch.

## What changed and what is verified

Original frozen product/tests/controller/oracle and prior packets unchanged.
New isolated leaf contains exact product350CAB1A and full-suite copy whose ONLY
change is original New-TBRequest definition -> fixture-builder.ps1 load. Original
assertions,literals,group ordering and expected full24PASS/27line contract preserved;
static inverse transformation verified before freeze.

First wrote/froze focused tests with exact original3line constructor,watched RED at
P2AFixtureSchemaIndependent,then wrote the helper repair,statically parsed/sealed
it without extra execution,and watched SAME focused tests pass9groups:
SchemaIndependent,NumericIndependent,MutationIsolation,Defaults,WrongTypes,
DecoratedInputs,Ranges,NoCallbacks,Complete.

Repaired helper creates independent typed-array storage only for numeric fixture
values. Explicit numeric decorations are forwarded rather than stripped;wrong
types/nulls/ranges keep exact characteristics. Fixed SDK type/descriptor access,
no caller GetType/ToString/getter invocation. Focused tests confirm all five numeric
leaves independently owned,retained A metadata/payload with clean B/C,default
independence,explicit decorated preservation,wrong types/ranges and observable
callback-zero controls. Product is not imported/called by focused tests.

This is evidence for the fixture-isolation change WITHIN those focused cases,not
evidence that the original full InputMutation behavior is fixed or product cleared.

## Remaining failure and diagnostic limitation

Full suite emitted18PASS labels through RangeRelationships,then failed InputMutation.
OutputOwnership,SequentialIsolation,RefusalShape,ConservativeFlags and Complete
were not reached. No full green,mutant sensitivity,restored final green or review.

The full-suite catch deliberately remains unchanged and emits only a group label.
Its555byte stdout is byte-identical to the previous full failure's stdout,NOT
proof of the same subassertion/cause. The OLD diagnostic stack (old helper,input
Schema initial-plain check) cannot be attributed to this repaired-helper run.
The current failure might occur at a later operation in the same group;no saved
site/stack/leaf state distinguishes those alternatives. No reproduction was
performed to reconstruct missing data.

Next recommendation: a separately scoped site diagnostic against this exact
repaired helper/full copy,recording InputMutation boundaries and the failing
subassertion without changing implementation or weakening assertions. Preserve
these records;do not replay frozen slots or assume prior trace still applies.
No additional repair proposed before that evidence. Any new diagnostic/repair/
validation requires an explicit new scope/allowance,not the unused3slots.

CodeProperty/internal-error static-only coverage,rawString held-view/reread
distinction,PS5,concurrency/type-data,catastrophic recovery and real clocks/
identity/supervision/streams/storage/cleanup limits remain unchanged.

## Scope/skill ledger

Reused linked non-submodule worktree/dev branch/origin,initially clean at45a32ea7.
Git-only privileged access for ownership restrictions;no trust/ownership/security/
dependency changes. Test/controller hosts remained ordinary.
Nine portable files parse/JSON/LF/filter checked;old builder/product phase hashes
separate from immutable811checks=800inherited+4latestanchors+7newfiles.
SealRepair written only after exact RED;RED records were not rewritten.

Ruling: native PowerShell capture/ledger replaces Bash helpers under Windows-only
scope;no generic runner-completion evidence.
Ruling: approved exact6+1route overrides generic extra baseline/full-project/task-
done/fix/retest/review/cleanup loops;cost is no broader completion evidence.
TDD and verification skills require reporting focused green AND full failure.
Conditional requesting-code-review gate not met;no reviewer dispatched.
All original and new raw packets retained;no frozen source/oracle correction.

## Portable ordered stdout

Transcriptions below do not replace raw bytes or constitute additional runs.

### Slot1 red

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
ASSERTION FAILED: P2AFixtureSchemaIndependent
```

### Slot2 focused

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AFixtureSchemaIndependent
PASS: P2AFixtureNumericIndependent
PASS: P2AFixtureMutationIsolation
PASS: P2AFixtureDefaults
PASS: P2AFixtureWrongTypes
PASS: P2AFixtureDecoratedInputs
PASS: P2AFixtureRanges
PASS: P2AFixtureNoCallbacks
PASS: P2AFixtureComplete
LIMIT: fixture regression only; no product import or call
```

### Slot3 full

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
ASSERTION FAILED: P2AInputMutation
```

## Source identities (SHA256)

Leaf:docs/diagnostics/isolated-fuzz/p2a-fixture-isolation-repair/.

| File | SHA256 |
| --- | --- |
| .gitattributes | E3083818059D2956756D32B1D4A7098E28F57DFB3B971F429A73E51672505B9D |
| capture.ps1 | DECA64A7966FDF3EF7B9389A2E2D50614567C5D4F165207BDB8F6366559D02F0 |
| expected-output.json | A6EF7F586CEFE679DDD804A3F475C8BFF035022DB8396085D523DEA2032405A9 |
| fixture-builder.ps1 | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| fixture-isolation.tests.ps1 | AA47F130563D3B12F888D9BC8341AFFFBDBBC293F5F286E0C691CCA3027B9693 |
| full.tests.ps1 | 4BF2F3C9B4FB18809FD07187CEC598B0DA8B4CB9F42836DA5520F3814D18BBB8 |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| README.md | 1F3A918D3E6750B87BF078BF1A88D4337C99223E3A9C13ADD42EDABEBE8C501C |
| source-pins.json | 2394205C36D071319C1E46C2C98CF7100F755EEC48ECBB6983F5FFC2935339C2 |

Ignored packet:.superpowers/sdd/2026-10-18-p2a-fixture-isolation-repair/.

| File | Bytes | SHA256 |
| --- | ---: | --- |
| frozen.json | 355507 | C7920AD9E89E4D78718E08B59A2CEEB7ECBB71AF3A40851E281E4310D5F506F3 |
| repair-seal.json | 411 | 0EA5506C92229C7F394592DA1E573C34C30D471CF4FD069CB4C67961C1DDC90B |
| invocations.json | 4439 | E48A6239B67258C2D36E58E321D3712BF25B793150F0BD8752A68B75B8DBBC52 |
| old-builder.ps1 | 290 | D3093A8DD92C9BD3ACD757CA3D9216F8D6482451203728ABA05D2F5A625C71E0 |
| repaired-builder.ps1 | 1723 | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| product-baseline.ps1 | 9607 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| red.stdout.bin | 119 | 5042CC72DB3027BEF2A3ACDE9751A3BD6CB43F4BFC5AA36B8E2904FC0C449642 |
| focused.stdout.bin | 403 | E50AFB71D3D8EEE951B01EBEFC3AF6CCD800156B66D20D40F0E20ABBAA4446C7 |
| full.stdout.bin | 555 | 6501C845BA8E87A3507971494CA64995BB898207F7D0A56E50D54DB3EE805064 |
| each of3stderr.bin | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

EngineSHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.

## Recovery

Save nine necessary portable files,this report,status/recovery and approved-design
actual stop status on existing dev branch. No ignored raw packets,caches,binaries
or secrets committed/deleted. Git log/remote SHA check establishes the eventual
checkpoint;no push claim before it occurs. Fresh machines recover source/oracles/
focused result/full stop state,not local raw receipts,historical pins,engine cache
or sessions. Do not automatically replay spent/frozen allowance or merge main.
