# Authority V7 Readonly Metadata Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Execution method must be confirmed for this plan;Native/inline is recommended below.

**Goal:** Add the complete146-entry readonly protocol metadata registry and independent registry/domain-framing fixtures without implementing authentication,storage or consumers.

**Architecture:** A fixed schema-sorted private literal table returns defensive descriptor copies. Two independently source-checked literal fixtures and the unchanged name-only oracle test exact identity/metadata correspondence;standard SHA256 tests frame synthetic canonical bytes without calling authority helpers.

**Tech Stack:** Existing Go1.26.0 module/toolchain Go1.26.5,standard library and unchanged internal/strictjson for test-only decoding;Windows PowerShell5.1 ordinary offline verification,no new dependency.

**Spec:** [Approved registry design](../specs/2026-10-16-authority-v7-registry-design.md),[C1](../specs/2026-10-16-authority-v7-request-auth-clarification.md),[C2](../specs/2026-10-16-authority-v7-metadata-admission-clarification.md),[nine-item appendix](../specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md),[27-request matrix](../specs/2026-10-17-authority-v7-request-auth-matrix-design.md),[16-item appendix](../specs/2026-10-17-authority-v7-data-path-clarification-design.md),and the complete literal checklist in [final146-row reconciliation](../../roadmap/2026-10-17-task11-final-metadata-reconciliation.md). Read all before implementation;do not use the superseded147 header count.

**Status:** PROPOSED,not executed. Written16-item appendix confirmed after0ec812cd;local reconciliation146/0,not independent certification. No product files,fixtures or native controls were created for this plan. Implementation,method and the new bounded verification budget await confirmation.

## Global Constraints

- W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator;branch codex/c12-b01-task9-coordinator;origin https://github.com/win202569/Talenro.git. Reuse the existing isolated worktree;no new branch/main merge,force push or branch deletion.
- Exactly146 canonical object/schema pairs from v7 §4.1. Exclude AuthorityEffectCommitmentV1,AuthorityEffectResolutionV1,ActivationDecisionEvidenceV1 from this registry while preserving their historical support.
- Body domain remains SHA-256(ASCII("talenro.c12."+schema) || 0x00 || JCS(body));signature domain remains talenro.c12.signature-envelope.v1. No wire/body/schema/algorithm/key/root/policy/canonical approved-set changes.
- Metadata describes protocol form only. A non-null role does not establish authentication;EvidenceKinds=[] does not waive request/channel checks. DB-form eligibility does not prove storage/commit;null is not a general trusted-object grant.
- Preserve all four name-only files byte-identical,including schema_names_test.go. Do not modify strictjson,authority helpers,diagnostic/C# helpers,SQL,manifest,generated/OpenAPI,provider/readiness/stateformat/stateverify or any consumer.
- No public digest/codec/parse/sign/verify/authorization capability,registration/override or runtime fixture/Markdown/config/environment reading. No new dependency,install,download,security-setting change,Docker/Bash/WSL,DB or live signing operation.
- Fresh fixture files must be UTF8/LF,each at most64KiB. Independent expected literals cannot be generated from the product table,private authority switches or the registry fixture.
- This is metadata-only scope. Every A owner/transaction/journal/publication/restore/Down guard obligation in the16-item appendix remains unimplemented here and needs separate planning. No acceptance of Task11/C12/Batch01/production/main.
- Verification is only the five newly proposed phases below,at most5 native builds/4 native runtimes. All earlier leaf/name-directory/null-fix identities and approvals remain consumed. Unexpected result or audit uncertainty stops,not an automatic retry/fix pass.
- Necessary files and progress/recovery documents travel with the existing development branch per AGENTS.md. Raw receipts,SDK/modules/cache/tmp/binaries remain ignored;cloning is not environment readiness.

## Review Focus

1. Missing signer-role key versus explicit null,or null evidence array/element:TestAuthorityV7MetadataFixtureRejectsMalformed must reject missing/null misuse while accepting explicit role null and nonnil [].
2. Same-count metadata substitutions,request roles with [] and form swaps:TestAuthorityV7MetadataFixtureRejectsMalformed and TestAuthorityV7SchemaRegistryMetadata must compare all four fields,not just146 names.
3. Pointer/slice aliasing between calls,entries and lookup/list:TestAuthorityV7SchemaRegistryOwnership must prove each returned role pointer and evidence slice is independent.
4. Domain-prefix/NUL/schema/body confusion:TestAuthorityV7MetadataDomainFraming must match three literal preimages/digests and distinguish all five malformed framings without claiming canonical body validation.
5. Unknown/case/whitespace/historical names or accidental trust API:TestAuthorityV7SchemaRegistryRejectsUnknown and TestAuthorityV7SchemaRegistryPublicBoundary must preserve exact-match finite-error/zero-value behavior and metadata-only exports.

---

## File map and fixed interfaces

Create only these product/test/fixture files:

| File | Responsibility |
| --- | --- |
| internal/nodecontrol/contracts/schema_registry.go | Fixed complete metadata table and readonly API;no consumer. |
| internal/nodecontrol/contracts/schema_registry_test.go | Independent fixture decoding,semantic mutations,framing,API/ownership/public-boundary tests. |
| testdata/c12/authority-v7/schema-registry.v1.json |146 exact literal descriptors,coverage=metadata-registry. |
| testdata/c12/authority-v7/schema-registry-golden.v1.json | Independently transcribed146 expected descriptors and three framing-only vectors. |

Update this plan,docs/roadmap/2026-10-17-task11-metadata-registry-implementation.md,
docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md with actual
results. Do not create a success report before observing outcomes.

New production file has exactly these four exported declarations,no methods:

```go
type AuthorityV7SchemaDescriptor struct {
    Schema string
    ObjectName string
    EvidenceKinds []string
    SignerRoleOrNull *string
}
var ErrUnknownAuthorityV7Schema = errors.New("contracts: unknown authority v7 schema")
func LookupAuthorityV7Schema(schema string) (AuthorityV7SchemaDescriptor, error)
func AuthorityV7Schemas() []AuthorityV7SchemaDescriptor
```

Known empty kinds must be allocated nonnil empty slices;null role is nil.
Unknown returns the exact zero descriptor (nil kinds/role) plus the identical
sentinel,without input echo/wrapping. No normalization,aliases or fallback.
Use a fixed146-row schema-sorted private table and linear lookup. An internal
compact literal row may store empty role/kind strings as the representation of
the already-approved explicit null/[];it must not default missing/unknown rows.
Construct a fresh pointer for each non-null role and fresh evidence slice for
every descriptor returned,including each entry of each list.
The only production import needed is errors;no reverse authority import.
Document that ObjectName is a protocol label,not an implemented DTO or codec.

Frozen name-only SHA256:

| File | SHA256 |
| --- | --- |
| internal/nodecontrol/contracts/schema_names.go |2428B95DCF01560B36661799D8EDD354ABC25000DD82EF55107C7CCFC724348A |
| internal/nodecontrol/contracts/schema_names_test.go |554A8EBC3A494DD6943DA41F706DA0A1E6AA1A8C9D78279092B0024925EA3F23 |
| testdata/c12/authority-v7/schema-name-registry.v1.json | F505A358044D4F4292B97BE7003C54FC21E573276C3308B6E2055A755AD981E1 |
| testdata/c12/authority-v7/schema-name-registry-golden.v1.json | FA1586826F876991B2B4F142650F9D8970F7CA07A4015482E5713CF6CE3B3EE8 |

Recheck all four against this plan's starting Git commit before editing;
schema_names_test.go includes the completed null-negative fix and stays unchanged.

## Task1: Independent complete fixtures and framing contract

**Files:** Create the test file and both metadata JSON fixtures;no schema_registry.go yet.

**Interfaces:** Consumes final146-row checklist and approved source documents,
plus existing readNameGoldenFixture() ([]nameFixturePair,[]string,error).
Produces private test-only types metadataFixtureEntry (Schema,ObjectName string,
EvidenceKinds []*string,SignerRoleOrNull *string,exact JSON tags),metadataFramingVector
(Schema,CanonicalBody,PreimageHex,SHA256Hex strings),and helpers:

```go
func decodeMetadataRegistry(body []byte, expected []metadataFixtureEntry) ([]metadataFixtureEntry,error)
func decodeMetadataGolden(body []byte) ([]metadataFixtureEntry,[]metadataFramingVector,error)
func readMetadataGoldenFixture() ([]metadataFixtureEntry,[]metadataFramingVector,error)
```

Use nameFixtureBase for paths. No new export or modification of the existing
name helper. Both fixture row arrays are sorted by exact schema and checked
against all146 unchanged name pairs. Reuse the existing881 name-golden negative
strings in Task2 rather than duplicating them in the new <=64KiB golden.

Registry root exact keys:schema,coverage,entries.
Markers:talenro.c12.authority-v7.schema-registry.v1 / metadata-registry.
Golden root exact keys:schema,coverage,expected_entries,framing_vectors.
Markers:talenro.c12.authority-v7.schema-registry-golden.v1 / registry-and-domain-framing.
Each row has exactly schema,object_name,signer_role_or_null,evidence_kinds.
Role key is required even when null;non-null roles must be nonempty exact strings.
Kinds key is required and nonnull;[] or a single approved nonnull string only.
Golden vector exact keys:schema,canonical_body,preimage_hex,sha256_hex.

- [ ] Write TestAuthorityV7MetadataFixtureInventory first. It opens both absent new paths,requires the exact markers and146 sorted unique schema/object pairs,all four fields matching each other and unchanged name golden;assert71 EXT/non-null,27 DB/null,32 []/non-null,16 []/null,37 distinct roles and three framing vectors. At this point use no new product API.

The inventory test's core size/required-shape assertions are:

```go
if len(entries) != 146 { t.Fatalf("entries=%d", len(entries)) }
if len(vectors) != 3 { t.Fatalf("vectors=%d", len(vectors)) }
for _, entry := range entries {
    if entry.EvidenceKinds == nil { t.Fatal("missing/null evidence_kinds") }
    if entry.SignerRoleOrNull != nil && *entry.SignerRoleOrNull == "" {
        t.Fatal("empty signer role")
    }
}
```

- [ ] Run fixture-red once through the protocol below. Require verified build0/runtime1,only the inventory test failing because the two new fixture files do not exist. A compiler/native/decoder failure is not the expected RED. No placeholder fixture or product code before this observation.
- [ ] Implement strict test-only decoders and independently transcribe the two literal rowsets from the final table and each approved source. Registry and golden must be separately reviewed against source,never copied/generated from each other or product code. List each original ID/source in the implementation report's146-row checklist;retain the portable reconciliation as the review oracle.
- [ ] Use unchanged strictjson.Decode(bytes.NewReader(body),64<<10,&typedWire) first to reject unknown/case/duplicate fields,wrong types,invalid UTF8/trailing bytes/oversize. Preserve null elements with []*string;reject nil evidence array/elements. Then use standard encoding/json map[string]json.RawMessage solely for required-key presence at root/row/vector levels;explicit signer null is allowed but omitted signer key is not. Do not pass RawMessage through strictjson's typed validator or edit strictjson to make this work. Reject null rows,empty names/roles and all missing keys.
- [ ] Add TestAuthorityV7MetadataFixtureRejectsMalformed with named in-memory mutations for both fixture shapes:root/row/vector unknown,case-changed,missing and duplicate keys;null root/row/array;missing signer versus allowed null signer;empty/number/bool/object/array role;null evidence element at first/last position;number/bool/object/array kind;duplicate kind,unknown kind,both kinds;wrong markers,unsorted/duplicate object/schema,145/147 entries;146-row missing+extra replacement;wrong approved role and DB/EXT/[] swaps;role swaps preserving counts;oversize,invalid UTF8 and trailing content. Compare each semantic mutation with the untouched opposite fixture so a valid but wrong role does not become its own expected value. Valid explicit role-null and [] controls must pass. Do not mutate tracked fixtures on disk for negative tests.
- [ ] Add TestAuthorityV7MetadataDomainFraming with these literal vectors. Assert decoded preimage bytes equal prefix+schema+NUL+body and independent crypto/sha256 equals the literal digest. Assert digest inequality for omitted prefix,double prefix,omitted NUL,changed schema and changed body. Synthetic bodies are already-canonical framing bytes only,not legal Receipt/Record bodies;no JCS/body/crypto-signature validity claim.

| Schema | canonical_body | preimage_hex | sha256_hex |
| --- | --- | --- | --- |
| claim-v1-receipt.v1 | {} |74616c656e726f2e6331322e636c61696d2d76312d726563656970742e7631007b7d |42b3b66ec6d4150843820279b2db2d399e6c963d2060f38163404b41f0406191 |
| claim-v1-record.v1 | {} |74616c656e726f2e6331322e636c61696d2d76312d7265636f72642e7631007b7d |c51d55e1da084dde8b73181e3e1efe0b05abeff0f8d941ff3ba119ca78bc17ff |
| claim-v1-receipt.v1 | {"k":"v"} |74616c656e726f2e6331322e636c61696d2d76312d726563656970742e7631007b226b223a2276227d |110a41e1df0d9c82d1718b9de7d1f94a48abe7c77119e40c355a04f6aa94ab7d |

Planning calculation used .NET standard SHA256 over explicitly assembled UTF8 bytes;
future Go tests independently recompute it. Prefix/schema are ASCII;body is UTF8.
No private production helper supplied these expected values.

- [ ] Format only the new Go test with the existing copied gofmt. Run fixture-green once;all fixture/framing tests PASS with accepted ordinary evidence/full audits.
- [ ] Review explicit new test/two-fixture/doc list,commit as test: freeze authority v7 metadata and framing oracles. This commit is test-only;no metadata API yet,no overall acceptance. Preserve progress under AGENTS.md without adding raw controls.

## Task2: Readonly API,ownership and complete scoped regression

**Files:** Create schema_registry.go;extend schema_registry_test.go;update this plan
and the implementation/current-status/recovery documents.

**Interfaces:** Consumes Task1 helpers and frozen literal oracles.
Produces the four-declaration production API above,with no API change to names.

- [ ] Write TestAuthorityV7SchemaRegistryMetadata:for every independent golden row,lookup must succeed with exact Schema,ObjectName,role value/null and nonnil exact kinds;list must have146 entries in strict bytewise schema order and match every field. Compare name-only lookup/list identity for all146. Explicitly assert#7 admin request role/[],#16 local null/[],#47 DB/null not commit proof,#60 journal null/[],all27 profile roles/[],and D16's exact5/3/8 partition. These are metadata assertions,not actual auth/storage tests.
- [ ] Write TestAuthorityV7SchemaRegistryRejectsUnknown:all881 unchanged name-golden negative strings,including genuine empty string,case/whitespace/version/alias/historical inputs must return descriptor==zero (reflect.DeepEqual) and err==ErrUnknownAuthorityV7Schema;error string exactly contracts: unknown authority v7 schema,never input echo. Retain original nameAPI behavior.

Pin the independent role/empty-set and unknown boundary with assertions such as:

```go
got, err := LookupAuthorityV7Schema("publish-claim-v1-credential-policy-request.v1")
if err != nil || got.SignerRoleOrNull == nil || *got.SignerRoleOrNull != "claim_v1_security_admin_request" {
    t.Fatalf("C1 request role: %+v, %v", got, err)
}
if got.EvidenceKinds == nil || len(got.EvidenceKinds) != 0 { t.Fatal("C1 evidence kinds") }
bad, err := LookupAuthorityV7Schema("")
if err != ErrUnknownAuthorityV7Schema || !reflect.DeepEqual(bad, AuthorityV7SchemaDescriptor{}) {
    t.Fatalf("unknown result: %+v, %v", bad, err)
}
```

- [ ] Write TestAuthorityV7SchemaRegistryOwnership:mutate returned Schema/ObjectName,role pointer contents,each nonempty kind element,and append to empty kinds;then new lookup/list must equal untouched golden. Check lookup→lookup,list→list,list→lookup,lookup→list and two entries sharing a role. Distinct non-null role pointers must not alias even within one list;[] remains nonnil after fresh calls. Mutating order/entry replacement must not change subsequent sorted lists.
- [ ] Write TestAuthorityV7SchemaRegistryPublicBoundary using go/ast plus reflection:exact four exported declarations from the new file,no exported methods,exact signatures/four descriptor fields,only errors import,private fixed literal146 rows,no runtime I/O/register/override/decode/verify/exported extra capability. Check doc comments explicitly distinguish protocol label/metadata from DTO/verification/storage/admission. Tests parse source but production does not.
- [ ] Run api-red once. Expected verified build exit1 with undefined new API symbols in this new test file;no runtime. Stop on unrelated build error. Preserve full compile diagnostics and guards rather than treating any nonzero exit as TDD proof.
- [ ] Implement exactly the frozen production API and146 literal metadata rows using the final checklist. Return fresh pointers/slices as specified;do not read fixtures at runtime or infer role from names. Do not alter consumers to make the new table used. Format only the two new Go files.
- [ ] Run api-green once,requiring all named registry tests and all existing name tests PASS with ordinary evidence/audits. Then run contracts-final once with selector .;require complete contracts-package default-vet build/runtime PASS,no SKIP/FAIL/instrumentation and all compiler/source/tool/fixture/binary bindings. Record actual top-level/subtest names/counts and durations,not old55/429 as a new result.
- [ ] If the selected method is Native/inline,request one fresh readonly whole-slice review after scoped regression,using the current session's model unless the user explicitly selects another. If Subagent-driven is selected,follow that skill's task/review gates instead. Any important finding stops acceptance;no new implementation fix or native retry budget is silently granted. No additional review of consumed old attempts.
- [ ] Write actual scoped results and remaining exclusions,verify four name-only files unchanged,review explicit file list and commit as feat: add readonly authority v7 schema metadata registry. Push necessary completed/explicitly unfinished checkpoint normally to current branch and verify remote SHA=local/clean. Do not merge main or equate metadata-only verification with full Task11/C12 acceptance.

## New bounded verification protocol — proposed,not authorized yet

S=W/.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator.
R=S/metadata-registry-implementation-001;control=R-control.
Both were absent in this planning read-only check. After approval,recheck absent,
safe,non-reparse,nonoverlapping paths and current branch/origin/baseline.
If occupied,drifted or unsafe,stop;do not delete/reuse or invent a fresh-looking
root to retry. New-computer preparation requires revised environment confirmation.

After approval only,create ignored control/verify.ps1 with
Invoke-MetadataRegistryPhase([string]Phase),accepting only the five fixed rows.
Adapt the operational structure of the completed name-directory/null-fix wrappers,
not their old identities/controllers/saved readers. Parse the new wrapper in the
available host and fresh64-bit Windows PowerShell5.1 before any workload.
Use unchanged Invoke-MatrixRecordedProcess(Executable,Arguments,Environment,
WorkingDirectory,ReceiptDir,LimitSeconds) from matrix-check.ps1 and its ordinary
helpers;do not invoke Invoke-LimitedMatrix or old fixed leaf controllers.

Pin the four exact inputs and71-helper aggregate from
[ordinary001 frozen inputs](2026-10-16-leaf-ordinary-check-001.md#frozen-inputs-and-preparation-observations);
values are preserved here to avoid inventing new expectations:

| Input | SHA256 |
| --- | --- |
| S/framework-001/copy-receipt.json | AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55 |
| S/uninstrumented-001/copy-receipt.json | DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10 |
| S/uninstrumented-001/toolchain/bin/go.exe | D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F |
| C:/Users/Lenovo/go/pkg/mod/github.com/google/uuid@v1.6.0/go.mod |73DA47B6338B00A082FD451AA35A3273D3ADC09B8E9BBA98DAB01091E402AF6E |

All71 immediate .ps1/.cs helpers under W/docs/diagnostics/isolated-fuzz:
sort Name|UPPERCASE_SHA256 strings with StringComparer.Ordinal,join LF with finalLF,
hash UTF8/no BOM;expected AF83F1057EA2B71E31BCB38BD59F6518C2C68424BE77F1B1FD62A04D17340869.
Recheck all pins,full copied-SDK/reference and product inventories before/after
each phase. Freeze actual source/test/fixture bytes for each phase;pre→post may
not change within a phase. Eligible independent final source/SDK/fixture/tool/
artifact audits must run even after failure;preserve primary and secondary errors.
An input mismatch stops;do not silently refresh expected hashes.

Toolchain=S/uninstrumented-001/toolchain;modules=C:/Users/Lenovo/go/pkg/mod.
Fresh64-bit host:
C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -NoProfile -NonInteractive.
Record host version/bitness and avoid inherited PS7-specific module paths.
Exact20 child environment keys:

```text
GOROOT=<Toolchain>
GOCACHE=<R>/go-cache/<phase>
GOMODCACHE=C:/Users/Lenovo/go/pkg/mod
GOTOOLCHAIN=local
GOENV=off
GOPROXY=off
GOSUMDB=off
GOAUTH=off
GOVCS=all:off
GO111MODULE=on
GOWORK=off
GOFLAGS=-mod=readonly
GOOS=windows
GOARCH=amd64
TEMP=<R>/tmp
TMP=<R>/tmp
GOTMPDIR=<R>/tmp
TALENRO_FUZZ_DIAG_MODE=
TALENRO_FUZZ_FIXTURE=
GODEBUG=
```

Each phase gets a fresh cold GOCACHE. Exact build arguments from W,default vet,
native600s:

`test -c -x -p=1 -count=1 -timeout=120s -o=<R>/<phase>/contracts.test.exe talenro.local/platform/internal/nodecontrol/contracts`

Runtime only after accepted build/compiler/new-binary binding,from
W/internal/nodecontrol/contracts,native150s:

`-test.count=1 -test.timeout=120s -test.v=true -test.run=<selector>`

| Phase | Selector | Expected build/runtime | Required observation |
| --- | --- | --- | --- |
| fixture-red | ^TestAuthorityV7MetadataFixtureInventory$ |0/1 | Only absent new fixture paths cause inventory failure. |
| fixture-green | ^TestAuthorityV7Metadata(Fixture\|DomainFraming) |0/0 | Inventory,malformed rejection and framing pass. |
| api-red | not executed |1/none | Exact undefined new registry API diagnostics,no runtime. |
| api-green | ^TestAuthorityV7(SchemaRegistry\|Name) |0/0 | Registry exactness/ownership/boundary and unchanged name suite pass. |
| contracts-final | . |0/0 | Entire contracts package passes. |

Maximum5 builds and4 runtimes,each phase once. RED nonzero exits still require
natural completion,exact ordinary guards and accepted audit bindings.
Use unchanged IsolatedFuzzProcess.Run(...,false,false),typed schema3 completion/
query evidence,500ms natural observation and shared5000ms cleanup. Do not exempt
error87,PID/identity/exit/query failures,truncation,forced termination or timeout.
No extra go env/version/list/vet/warmup,race,fuzz engine,benchmark or whole-repo run.
Ordinary complete-package execution may run existing fuzz seed tests,not fuzzing.

Persist phase/attempt identity immediately before each native boundary;entering
the collector is not proof the child launched. Bind actual executable/arguments/
environment/cwd/source/compiler/tool and stdout/stderr/receipt hashes.
Accept TDD RED only from the exact named reason,not generic exit1;parse Go
parent-first/subtest output semantically,not an assumed leaf-first order.
No saved record rewriting after evaluation failure. An output-order parser issue
is a stop requiring a saved-only decision,not permission to replay RED.

Before declaring scoped verification,require final nonempty stdout line PASS,
at least one original top-level Test PASS,expected named selectors observed,
no unexplained FAIL/SKIP/instrumentation,all final audits and durable complete
bindings. Record actual counts per phase and distinguish expected RED from
GREEN/full-package results. Acceptance=false throughout;do not reuse the old
leaf15-artifact manifest or report a package pass from provisional output alone.
Any unexpected result,native/persistence uncertainty or audit failure stops forward
execution. Preserve an unfinished checkpoint and request a separate decision;
this plan includes no automatic repair/retry/reviewer-fix allowance.

## Plan self-review and handoff

Checked inline:full146-source precedence and exact four-field oracle;C1/9/27/16
exceptions;explicit null/[] and presence handling;defensive pointer/slice ownership;
three literal framing vectors and five framing negatives;all five Review Focus
items mapped to named tests;four-file implementation boundary;five-build/four-
runtime ceiling and no consumed replay. No storage/auth/codec scope is implied.
Fixture/body canonical validity and actual source/transaction enforcement remain
deliberate exclusions,not missing tasks inside this metadata-only deliverable.

Please review this plan and confirm the execution approach before any product
or native-control edits. Recommend Native/inline:the two tasks share one small
test/oracle contract,and I can implement them directly with one final fresh
readonly review. Subagent-driven remains an alternative with per-task gates.
No implementation,Go command or native workload ran while writing this plan.
