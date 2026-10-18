# P2-B wrapper reference diagnostic: controlled sharing confirmed

## Scope and result

User approved ONE independent ordinary PS7 process after f3890d52 to observe
wrapper/base/key/table/descriptor-owner references. This is an evidence-only
spike,not a product repair or acceptance test. New diagnostic-only audit leaf:
`docs/diagnostics/isolated-fuzz/p2b-wrapper-reference-diagnostic/`.

Outcome DIAGNOSTIC_CAPTURED: all12cases and36phases complete,Error=null,
87distinct references registered. PID48084,2960ms,exit0;no timeout/stderr,
owned exit and concurrent separate streams complete. Allowance1/1 SPENT;
no retry,extra host,reviewer,product import/API,fixture compile or product edit.
All older frozen scopes stay frozen. M1 remains OPEN;no main/acceptance grant.

## Construction and observation

Four kinds: byte7,byte[2]{7,9},fresh string probe,and normal empty custom objects.
Three constructions each: same wrapper,different wrappers requesting the same
base,and independently constructed equal-content values. Custom objects use
normal empty PSObject constructors;their custom singleton base can be shared,
so that row is an independent-WRAPPER/key control,not an independent-base claim.
Independent byte,array and string bases were proven different by reference.

Setup uses fixed normal SDK constructors and reserved-view factories. Phases:
generate psobject via left wrapper and psextended via right;complete missing
views;repeat both views on both wrappers. Factories are used ONLY to construct
controlled SDK histories,not as origin/authenticity checks. The fixed SDK key
routine observes actual key references. Snapshot reads existing fixed fields;
no descriptor Value,MemberType or Copy invocation. No malicious descriptors.

All references are retained in object[] and compared through fixed reflection
argument arrays. Registry IDs identify actual reference equality,not hash codes,
primitive equality or guessed owners. All partial rows/phases are saved before
construction assertions. There is NO expected-sharing oracle;different owners
are recorded,not automatically failed. Both scripts parsed before freezing.
Before launch,base/wrapper branch assignments were made directly inside branch
bodies to avoid conditional-output pipeline enumeration/unwrapping. No frozen
source changed after launch. One initial read-only summary query had a shell
pipeline syntax error;only that non-mutating analysis query was corrected.

## Observed relationships

Relations below hold in all three saved phases for each construction:

| Construction | Cases | Wrapper same | Base same | Key same | Table same |
| --- | --- | --- | --- | --- | --- |
| Same wrapper,all four kinds | 4 | yes | yes | yes | yes |
| Distinct wrappers,same byte/array base | 2 | no | yes | yes | yes |
| Distinct wrappers,same string/custom base | 2 | no | yes | no | no |
| Independent byte/array/string instances | 3 | no | no | no | no |
| Independent empty custom wrappers | 1 | no | yes(singleton) | no | no |

Concrete byte case2: wrapperIDs7 and9,base/keyID6,tableID11. psobject descriptor8
retains owner7; psextended descriptor10 retains owner9/constructor-owner9 and
its extended raw table11. BOTH wrapper snapshots see the same two descriptors.
These identities remain unchanged after completion/reuse. Array case5 similarly
has wrappers28/30,base/key27,table32 and descriptor owners28/30.

This directly establishes that legitimate controlled SDK metadata can have
different wrapper owners while sharing one actual base,key and table. The
same-base string/custom controls also directly show why BASE equality alone
is insufficient. Equal-content byte/array/string independent controls stay
distinct;no primitive/content equality was used to merge identity.

Across132descriptor observations (repeated observations,NOT132unique cases):
all132 have hidden/reserved/instance flags true and owner key/table equal to
their container's key/table. All66psobject entries have internal-owner equal
to their own descriptor owner;all66psextended entries have constructor-owner
equal to their own descriptor owner and extended raw table equal to container.
These are after-the-fact read-only observations,not product acceptance rules.

## Portable evidence and limits

`portable-summary.json` is a DERIVED summary,not the raw capture. It preserves
all12construction records,36phase relationships,descriptor flags/owner/base/
key/table IDs and87reference CLR types. Column legends are embedded;null means
not applicable/unrecorded,not a successful equality. SourceRawSHA256 binds it
to the original stdout. It intentionally omits raw formatting,wrapper flag
values,immediate-base IDs and method-call lists;raw retains those details.
Read-only reconciliation matched every saved reference type/state and all132
descriptor observations to raw. Summary SHA256:
681DA52B4F28EBD9723C73DFFE110E6E54EF88D5F537BEA8346F26E2310DA583.

This controlled history is not the actual old failed invocation. The earlier
literal.Byte failure remains INCOMPLETE/unvalidated;its missing reference
relations cannot be retroactively filled. Product behavior,hidden-decoration
refusal,all nine scalar types,fresh-owned outputs,optional SDK descriptors,
local-storage mode and adversarial foreign-owner cases were NOT tested here.
No universal ownership/provenance/passivity/security claim follows. Reflection
diagnostic methods are not a proposed product implementation.

Next recommend a revised scalar ownership contract separating wrapper identity
from metadata key/table identity,while retaining exact descriptor structure and
descriptor-specific internal-owner/constructor-owner consistency. String/custom
branches need their distinct treatment;do not globally substitute base equality
or ignore owners. Present a bounded repair design and new finite verification
scope for approval before writing another product candidate. This completed
spike grants no repair/retest/reviewer or frozen-slot reuse rights.

## Integrity, budget and recovery

1090mapped pins match post-run and read-only rehash:1078inherited+5priorpacket+
2priorreports+4newinputs+livebaseline. Engine362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139;
SDK2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B.
Bounds60000ms execution+2000ms owned exit+2000ms separate concurrent drain;
only the owned target can be terminated. No install/network/security changes.

Ignored packet: `.superpowers/sdd/2026-10-18-p2b-wrapper-reference-diagnostic/`.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | F65B39A606ADA511CCB34E8BF48C366640A01241A276CD548193F12E0706CF34 |
| receipt.json | CD0871ACB0605188AD6BA5B31DAF34EE6A2A915EDA11C2BA6B96DE685A4C8537 |
| stdout.bin | 71877998E5A55A14EF9C5C0EFBE9F88BE84D488F702A45693E4E18A66F17F56A |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Sealed diagnostic.ps1 SHA1A62A204463B8B4BDD7562CD7ED7CA17810E5E294708D8D992C257BEE7461625.
ProductB468B564 unchanged;failed18DA24E1 unchanged and must not be adopted.
Independent diagnostic1/1spent;latest repair1/3spent+2FROZEN/review0/1ineligible;
earlier repair2/3spent+1FROZEN and every previous scope unchanged. No main or
formal P2-B/PS5/P2-C/P3/full C12/live/runtime/security acceptance.

Recover latest codex/c12-b01-task9-coordinator,not main. Portable audit scripts,
derived summary/report/status/recovery are committed. Clone does not restore
ignored raw,actual CLR object references,runtime/SDK,caches,mappings/sessions
or reset/replay rights. IDs preserve THIS run's recorded relations only.
