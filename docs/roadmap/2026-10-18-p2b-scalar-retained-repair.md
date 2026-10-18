# Retained scalar repair — tests exact; independent review blocks acceptance

Date:2026-10-18. Base:34bdafe520c8b42879aad964c66a593d66d65ba6.
User approved new bounded3ordered ordinary PS7hosts+1conditional independent
readonly review. New leaf:p2b-scalar-retained-repair. Historical source/raw/
receipts/frozen budgets unchanged. This is an UNFINISHED development checkpoint;
do not adopt the candidate as a verified complete repair or merge main.

## Actual outcome

Three test gates matched their exact approved expectations. One independent
review completed:Critical0,Important1(P1),Minor0. Parent accepts the Important
finding as an unresolved correctness/coverage blocker. M1 is NOT closed.
All3/3new hosts and1/1review spent. No postreview product edit,target,probe,
re-review or additional fixture compile. Old latest1/3spent+2FROZEN and all
older stopped allowances remain unchanged. Do not replay them on this or a
new computer.

| Gate | PID | Elapsed ms | Exit | Exact result |
| --- | --- | --- | --- | --- |
| RED | 41112 | 4626 | 1 | HiddenDecorationRefusal;7clean accept,14visible refuse,14hidden wrongly accept |
| focused | 42524 | 23933 | 0 | 7clean accept,all28decorations refuse;182malformed cases,45guards,54key controls |
| full | 44864 | 47949 | 0 | original33groups,exact36line oracle,tracecounter0/errornull |

Every gate also passed27 retained SDK proofs(nine types x two/all-five/reused
views)and14normal public API histories where applicable (RED/focused). Both
35-row matrices retained caller-side bridge wrapper/base/raw descriptors and
original exact tag descriptors. No callee-internal binding identity claim.
All ordinary pinned PS7,execution/owned-exit/drain60000/2000/2000ms,no timeout,
owned exit and concurrent separate streams complete,stderr empty.1154mapped
pins intact after all runs and review. Successful tests do not override review.

## Product changes retained,not accepted

Only new candidate control-frame.ps1 changed:fixed-field receiver/type/storage
helpers,type-sensitive SDK key rule,raw scalar/byte-array descriptor validation,
and shared fresh-copy guard retaining InnerException on failure. Exact
names/classes/flags,owner key/table,own internal/constructor owner and fixed
pstypenames getter/null setter are checked for materialized tables. Extra
visible/hidden note/subclass descriptors refuse in tested histories.
No descriptor Value/MemberType/Copy/stored getter execution used for validation.
The original root/common-owner validator and ALL product text from
Copy-P2BCMessage through public protocol APIs remain byte-identical to baseline.

Candidate327E8F33 is sealed. Baseline B468B564 and failed18DA24E1 remain unchanged.
The candidate is partial progress,NOT a new accepted baseline.

## Independent review blocker and parent source check

See [complete review](../diagnostics/isolated-fuzz/p2b-scalar-retained-repair/review.md).
Get-P2BCStorage line68reads only the wrapper's _instanceMembers cache;
Assert-P2BCPlain line92accepts a null table. A null local cache does not prove
that no shared-key metadata exists. Another default non-string wrapper over
the same base can already have decorated that key. Existing tests materialize
both wrappers before checking sharing and miss this state.

Parent checked the code and official SDK source:InstanceMembers fills an empty
cache by looking up the resurrection key. Thus the early acceptance decision
does not establish absence of shared metadata.
[PowerShell v7.6.5 InstanceMembers](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L606-L623).
This is a source-supported correctness finding,NOT a newly executed bypass.
The historical failing instance was not recreated and no new target ran.

Next bounded repair needs passive key-correct authoritative-table lookup when
a local cache is uninitialized,plus dirty shared-key/unmaterialized and clean
null-cache controls,including string/local non-sharing cases. Preserve no
product factory/key-method or descriptor-getter execution. That change needs
a new approved finite verification scope;current sealed product stays intact.

## Rulings on every declined-to-judge item

1. Null-cache runtime result:retain source-based Important blocker,do not claim
   executed reproduction. Cost:cannot close the repair without new evidence.
2. Callee-internal identity:caller snapshots prove caller retention only.
   Cost:no claim that internal API binding preserves every wrapper.
3. Exhaustive keys/histories:54materialized oracle cases are bounded evidence,
   not all cache/nesting/storage transitions. Cost:other histories unverified.
4. Depth beyond8:keep finite fail-closed behavior,not broad compatibility.
   Cost:deep legitimate wrappers may be rejected.
5. Reflection forgeries/concurrency/isolation:outside pure helper contract,
   no unforgeable provenance or atomicity claim. Cost:not safe as an arbitrary
   same-process attacker isolation boundary.
6. Injected fresh-copy failure:source preserves InnerException,but runtime
   chain not injected/tested. Cost:do not call that error branch runtime-verified.
7. Historical correctness:old files hash-only,old scopes not reopened.
   Cost:no new conclusion closes old failures or changes old acceptance.
8. Speed/PS5/other SDK/live:only finite pinned PS7 execution evidenced.
   Cost:no performance,portability or live-operation guarantee.
9. Formal acceptance/deployment/merge:excluded;authority flags remain false.
   Cost:this checkpoint is not deployable/merge-authorized acceptance.
10. Historical environment attestation:review checks receipt/stream/seal
    coherence,not recreation/independent attestation. Cost:no stronger claim.

There are no deferred Minor findings. The Important finding is not downgraded
merely because the original tests omitted the triggering state.

## Evidence and portable recovery

New packet:.superpowers/sdd/2026-10-18-p2b-scalar-retained-repair (ignored).
13immutable inputs plus baseline snapshot,newproduct seal,three receipts,
raw streams,full trace and one review-request record retained. New derived
portable-summary.json records per-process IDs,all70matrix results,54proofs,
28histories,182malformed cases,45guards,54key controls and full output/trace.
It is postcapture output,NOT a frozen input. Complete review is postcapture too.
No old raw/receipt rewritten. Reviewer initially resolved relative pin paths
from parent directory,then corrected the readonly workdir check:zero mismatches,
no test replay. Parent reconciled summary hashes and matrix/negative data to raw.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 4DB2EB76F700D8570AF68E3B8F33C7CE684CD34949C88478F8580C983F705602 |
| product-seal.json | D27FA3E646D5B9E41801ACFD56A51FAC0D74FD2E57DEE692DE43889D8C3EA014 |
| invocations.json | 870AE643DC372AC57EAA1ABC5FB188A36314780C5F5DC1101436E416179C36F5 |
| new candidate | 327E8F3362D92EF30A2BA71A9B0E8ECCEB896778F927560192483FA37FA2AC29 |
| red.stdout.bin (785334bytes) | 23A20030006F02452199EA327E1B83627FBA8A7FAF708EF21083E55DD5377995 |
| focused.stdout.bin (1541581bytes) | 9C9989880DC8D7998F70501FE58F30E4FE302594ECC9895DE943342C132B23FA |
| full.stdout.bin (1165bytes) | B7C48E01A63F3E338CFC7E3D8C17A3BD2880CFAED63E166AEAC43A20D106AA62 |
| all stderr(empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| full.trace.json | 1F9EE5357CF7DB4CE8D7C023BE6C017EDDB9E160D26F433AC5091C1329F43D12 |
| portable-summary.json(derived) | E414933C8B086B184719B839C40157CF01B69B4B39FC711B3B50070BBA2A55F0 |
| review-request.md | B781EE6F1C8D1FAEACBD2E3762E02624CFA766BD3A7F34D3233DCE84D301CC3C |

1154pins=inherited1134+priorpacket4+priorreport/summary2+13inputs+baseline1.
Necessary sources/tests/fixtures/summary/review/report/status/recovery preserve
this unfinished checkpoint on existing devbranch,not main. Clone excludes ignored
raw,actual references,SDK/runtime/cache/sessions/mappings and resets no budget.
No formalP2-B/PS5/P2-C/P3/fullC12/live/identity/capability/security/runtime approval.
