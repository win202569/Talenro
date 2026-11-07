# Task12 static preparation — no implementation or runtime authorization

Baseline: `03c53d2e84d927d35807dbed54bd2a47409a06f6`, development branch
`codex/c12-b01-task9-coordinator`. User approved interface/dependency, specification
and pending-case inventory only. Working tree was clean before this preparation.
No product/test/fixture edits, imports, compilation, tests, system queries or
Go/Docker/native workloads. 44/34/6/217 and NOT_ACCEPTED unchanged.

This is a bounded preparation checklist, **not an executable implementation
plan**, a complete field-by-field protocol audit, or permission to skip Task11.
It uses writing-plans dependency/coverage checks without generating execution
commands or an implementation budget.

## Authorities and precedence

1. [Batch01 plan, Tasks8/11/12](../superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md): file ownership and interfaces.
2. [Approved v7 amendment](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md): §4.1 canonical rules; §4.2 inventory/anchors; §4.3 isolation/shutdown; §4.4 incarnation/runtime/lineage; §§4.5–4.7 absence/epoch/empty; §7.1 source sealing; §9.3 Down.
3. [Task11 current recovery boundary](../superpowers/plans/2026-11-07-task11-recovery-boundary.md) and [ledger disposition](2026-11-07-task11-ledger-evidence-gap-review.md): Task11 and runtime remain blocked.

The Batch01 header explicitly corrects its older digest formula. Use
`SHA-256(ASCII("talenro.c12." + schema) || 0x00 || JCS(body))` from §4.1;
do not copy the old missing-prefix formula. The complete §4.1 schema/role mapping
and subsequent approved Task11 clarifications remain authoritative; this note
does not create a second registry or redefine any signing role.

## Dependency and existing-interface map

|Dependency|Observed source / planned interface|Task12 constraint|
|---|---|---|
|Task11 canonical layer|Plan names CanonicalBodyDigest, ParseCanonicalBody, VerifySignatureEnvelope, VerifyCanonicalEvidenceBundle; declaration search in internal/nodecontrol found none of these exact exported functions|Unfulfilled interface, not permission to use an ad hoc parser or treat private helpers as the finished canonical layer|
|Task11 registry/evidence|Task12 consumes sole registry/envelope/evidence validation; name fixtures alone are not it|Do not begin typed verifier implementation before the upstream contract and its acceptance path are available|
|Task8 normalized DTOs|[authority_v7_boundary.go](../../internal/nodecontrol/contracts/authority_v7_boundary.go): AuthorityV7UpMigrationFactsV1, AuthorityV7DownMigrationFactsV1, AuthorityV7DownAuthorizationRequestV1 and FactsV1|Reuse these DTOs; preserve transaction nonce, actual transaction ID and exact evidence bindings; no duplicate public representations|
|Task8 sealed grant capabilities|[v7_migration_grant.go](../../internal/nodecontrol/authority/v7_migration_grant.go): VerifiedAuthorityV7UpGrant/DownGrant, private newVerified constructors, ConsumeVerified functions|Future VerifyAuthorityV7UpGrant/DownGrant are validation factories, not new public raw constructors. Copying a capability must not duplicate its single-use state|
|Task8 actual-transaction authorization|[v7_down_authorization.go](../../internal/nodecontrol/authority/v7_down_authorization.go): VerifiedAuthorityV7DownAuthorization; AuthorityProtocolDowngradeAuthorizer.AuthorizeAuthorityProtocolDowngrade(context.Context, contracts.AuthorityV7DownAuthorizationRequestV1) (VerifiedAuthorityV7DownAuthorization, error)|Future VerifyAuthorityProtocolDowngradeAuthorization must produce the existing sealed type. No guessed future txid or pre-transaction authorization|
|Task8 catalog|[nodecontrol.v1.yaml](../../db/schema/nodecontrol.v1.yaml); [catalog test](../../internal/store/nodecontrol_v7_catalog_test.go), TestNodeControlV7CatalogHasExact51Tables|Fixed 25+26 registry, not caller-provided table lists. Test definition exists; not rerun here|
|B03/B11 boundaries|Batch01 global constraints and v7 §9.3|No production attestor/provider implementation here. Preflight grants outside transaction; only the dedicated disposable Down authorization seam signs actual txid inside Goose transaction|

Exact exported Task12 factory declarations were also not found by the scoped
source search. This is a source observation, not a compiler result. Private
canonical/envelope helpers already exist in v7_migration_grant.go; their existence
does not establish Task11 completion or authorize exporting them unchanged.

## Planned files and present status

All nine exact Task12 paths below were checked with Test-Path and were absent:

- `internal/nodecontrol/contracts/authority_v7_source.go`
- `internal/nodecontrol/contracts/authority_v7_down.go`
- `internal/nodecontrol/authority/v7_verifier.go`
- `internal/nodecontrol/authority/v7_migration_grant_verifier.go`
- `internal/nodecontrol/authority/v7_environment_verifier_test.go`
- `internal/nodecontrol/authority/v7_empty_source_verifier_test.go`
- `internal/nodecontrol/authority/v7_down_verifier_test.go`
- `internal/nodecontrol/authority/v7_migration_grant_verifier_test.go`
- `testdata/c12/authority-v7/environment-source-down-golden.v1.json`

Absence of these paths does not mean all related behavior is absent: existing
grant, authorization, catalog and boundary tests must be preserved. In particular,
`TestAuthorityV7DownAuthorizationBurnsBeforeEveryFallibleCheck` already defines
window-boundary, exact retirement coverage and one-use checks. This turn inspected
definitions only and makes no fresh claim that those tests pass.

## Pending coverage inventory

IDs below are preparation labels, not newly implemented test names or an approved
test-call count. The future focused plan must map each row to exact fixtures,
full field sets/roles/windows from the specification, and the accepted Task11 API.
Positive fixtures must be independently derived, not generated by the verifier.

|ID / family|Normative source|Required valid case and refusal cases|Intended owner file|
|---|---|---|---|
|P01 canonical ingress|§4.1 / Task11|Exact canonical typed body and envelope; refuse missing/extra/duplicate fields, noncanonical values, wrong schema/role/key/root/algorithm and evidence-arm splice. Reuse Task11 tests rather than implementing a second canonicalizer|Task11 dependency; Task12 consumer vectors|
|P02 inventory/anchors|§4.2|Complete latest ordered inventory and exact anchor set; refuse truncation, aliases, duplicates/extras, stale anchor and wrong identity tuple|v7_environment_verifier_test.go|
|P03 isolation/shutdown|§4.3|Local isolation window ≤15min and exact complete shutdown set; refuse missing session/route/capability isolation, missing set member and mismatched credential/provider. Local-only isolation cannot establish empty/full source|v7_environment_verifier_test.go|
|P04 incarnation/runtime|§4.4|Exact system/OID/name/incarnation, current holder binding/generation and non-superseded lease; refuse cloned OID/name mismatch, wrong process/mount identity, expired lease and rebind without permanent old-holder termination. Runtime lease ≤5min; registration challenge ≤2min|v7_environment_verifier_test.go|
|P05 lineage|§4.4 and Task12|Exact root/next preimages and permitted continuous history; boundary vectors 1/64 links, reject gap/sibling/65th link/fork before anchor and cross-identity chain. Do not invent or shorten the specification formulas|v7_environment_verifier_test.go|
|P06 empty/absence/epoch|§§4.5–4.7|Exact singular inventory with matching database/provider absence and epoch evidence; refuse environment_count>1, wrong nonce/provider, nonzero history/counts or splice. Provider absence window ≤5min; genesis admission is not pre-attempt classification|v7_empty_source_verifier_test.go|
|P07 source seal/archive|§7.1 / Task12|Complete membership retirement, per-member retirement, post-seal exact cover and archive; minimal indeterminate seal cannot export/archive/activate. Reject missing/extra members and kind/schema substitution|v7_empty_source_verifier_test.go|
|P08 Up factory|Task12 / §§4.3,5.1|Bind installation kind, nonce, local isolation and exact preallocated intent IDs; reject missing/expired/mismatched inputs; return existing sealed grant only|v7_migration_grant_verifier_test.go|
|P09 Down preflight|§9.3|Disposable fixture, latest retired inventory, all providers permanently retired, fixed catalog and forbidden-row zeros; reject production, partial retirement, unknown profiles/records, cross-provider replay, stale evidence and consumed latch|v7_down_verifier_test.go; v7_migration_grant_verifier_test.go|
|P10 actual-txid authorization|§9.3|Exact stored pristine inventory, nonce, actual Goose txid, catalog/anchor/retirement, dedicated role and ≤2min window; reject guessed/wrong txid, wrong scope, future issuance, expiry-at-now and replay/copy reuse; preserve existing burn-before-fallible-check semantics|v7_down_verifier_test.go; existing v7_down_authorization_test.go|
|P11 three-stage inventory|§9.3|Distinct pristine 1/0, authorized 1/1, final 0/0 control counts; reject stage substitution and reconstructing pre-auth digest from auth=1. Pure vectors cannot establish actual DB locking/rollback behavior|v7_down_verifier_test.go; later Task16 physical conformance|
|P12 temporal/evidence ownership|Task12 / §9.3|Historical admission evaluated at recorded transition time; new mutation requires fresh evidence. Reject caller booleans/opaque digest without preimages; distinguish stored exact retry from new authorization. Preserve defensive byte copies and one-use rights|All Task12 verifier test files as applicable|

## What is ready, and what is not

- Complete now: static dependency map, specification routing, planned-file presence
  checks and coverage families. No new algorithm, DTO, signature or golden data
  has been designed or frozen by this note.
- Before an implementation plan: resolve exact factory parameters from accepted
  Task11 types, expand each family's complete literal fields/roles/nullability and
  maximum windows, and distinguish existing coverage from genuinely new tests.
  The old plan's combined selectors are not a newly approved execution budget.
- Before implementation/testing: Task11 dependency and runtime prerequisites must
  be resolved through their existing explicit gates. No fake pass, source review
  or document approval accepts its failed RED or creates an alternate runner.
- Preserve source ownership: Task12 validates evidence; B03 supplies production
  attestation/provider behavior; B11 orchestrates; Task16 owns physical conformance.
  Do not claim pure fixtures prove permanent retirement, real cleanup or DB locks.

Self-review: every original Task12 interface family is represented above; existing
sealed types and the digest-prefix correction are preserved. This is author
static preparation, not an independent code/security audit. Document checks verify
local links, the explicit documentation-only diff and whitespace. End this scoped
preparation after committing/pushing; no automatic implementation or test follows.
