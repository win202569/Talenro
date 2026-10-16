# Authority V7 Name Directory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task after user review and execution-method selection. Steps use checkbox (`- [ ]`) syntax for tracking. Native/inline is recommended for these tightly coupled tasks.

**Goal:** Deliver the exact146 V1 schema/object-name pairs as a read-only directory, without authority or policy metadata.

**Architecture:** An internal fixed table backs exact schema lookup and a fresh sorted list. Two independently source-checked literal fixtures supply the test oracle; fixture decoding exists only in tests. Directory membership conveys no authentication, signing or evidence eligibility.

**Tech Stack:** Existing Go1.26.0/toolchain go1.26.5 module, standard library and internal/strictjson; Windows PowerShell, existing offline SDK/modules and unchanged ordinary process collector. No new dependencies.

**Spec:** [Approved name-only scope revision](../specs/2026-10-16-authority-v7-name-directory-design.md), implementing the exact names from [v7 §4.1](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md) and [146-pair source index](../../roadmap/2026-10-16-task11-registry-reconciliation.md).

**Status:** UNACCEPTED checkpoint after Native execution approved following6c5ec277. Implementation90dc0e2a/990d6683 and all five bounded phases verified; full contracts55 top-level/423 subtest PASS,0 SKIP/FAIL. Final review found Important1 (test-only negative-vector null acceptance), reopening strict fixture closure. Budget consumed; no fix/additional workload. Await [bounded fix addendum](2026-10-16-authority-v7-name-directory-null-vector-fix.md) approval; not complete slice/Task11/C12 acceptance.

**Execution ruling:** After the first phase, use fresh phase-specific children of R/go-cache for subsequent phases, keeping the same20 environment keys and all budgets/ordinary guards. A shared cache could otherwise suppress original compiler tokens in contracts-final when source is unchanged. This strengthens fresh compiler evidence; no additional build/runtime is authorized. Cost: more cold-build time/disk within the same600s limit. The first phase's original evidence remains unchanged.

## Global Constraints

- Use existing W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator and branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git. Preserve unrelated changes.
- Exactly146 schema/object pairs, case-sensitive, no trimming, aliases, version fallback or caller registration. Exclude AuthorityEffectCommitmentV1, AuthorityEffectResolutionV1 and ActivationDecisionEvidenceV1 from this new directory; preserve existing historical support.
- Only public type AuthorityV7SchemaName with Schema/ObjectName string fields, sentinel ErrUnknownAuthorityV7SchemaName, LookupAuthorityV7SchemaName(string) (AuthorityV7SchemaName,error), and AuthorityV7SchemaNames() []AuthorityV7SchemaName.
- Unknown returns the identical sentinel and zero struct, without input echo. Lists are ASCII/bytewise Schema-sorted independent slices; no shared mutable nested data.
- No SignerRoleOrNull/EvidenceKinds/placeholder, Verify/Sign/ParseBody, registry import, capability, keys, roots, policies, provider/readiness wiring, body codecs, cryptography or digest changes. No contracts-to-authority import.
- No runtime Markdown/JSON/config/environment reads. No SQL/generated/OpenAPI/approved-spec-set/private signer/helper/security-gate changes. Existing digest prefix remains unchanged.
- Metadata93 author classifications/53 unclosed reviews remain postponed, not certified. This slice cannot close full Task11/C12/security/production/main or historical error87/timeouts/fuzz/physical DB/PITR gates.
- PowerShell only; no installations/downloads, Docker/Bash/WSL, force push, branch deletion or main merge. Commit only explicit necessary files, never ignored SDK/cache/raw evidence/binaries/secrets.

## Review Focus

1. Count-preserving replacement or swapped pairing: Task1 validates full independent pair equality, not just146 entries.
2. Almost-valid input (case, whitespace, version, embedded NUL, historical name): Task2 rejects identically without reflecting attacker text.
3. Caller corruption of one returned list: Task2 changes fields/order and verifies later lookups/lists remain exact.
4. Plausible but malformed fixtures (missing/null/extra/duplicate/wrong-type/trailing members): Task1 explicitly rejects, including nested entries.
5. Directory mistaken for authority: Task2 pins the two-field/public-declaration boundary and documents no authority; Task3 reviews absence of consumer wiring.

## File map and interfaces

Create only these four product/test/fixture files:

- internal/nodecontrol/contracts/schema_names.go: fixed table, exact API and sentinel.
- internal/nodecontrol/contracts/schema_names_test.go: package contracts; independent test-only fixture types/validators, directory tests and AST/reflection boundary checks.
- testdata/c12/authority-v7/schema-name-registry.v1.json: literal146-pair inventory.
- testdata/c12/authority-v7/schema-name-registry-golden.v1.json: independent expected146 pairs and rejected-schema strings.

Do not occupy schema-registry.v1.json or canonical-envelope-golden.v1.json. Update this plan's checkboxes, docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md, and create docs/roadmap/2026-10-16-task11-name-directory-implementation.md for factual results. No production test-helper export.

Fixture contracts (exact member names, no optional members): registry has schema="talenro.c12.authority-v7.schema-name-registry.v1", coverage="name-directory", entries; golden has schema="talenro.c12.authority-v7.schema-name-registry-golden.v1", coverage="name-directory", expected_entries and rejected_schemas. Each pair has schema and object_name, both nonempty strings. Pair arrays have146 unique schemas and146 unique object names, strictly schema-sorted. Both fixtures are at most64KiB each.

Private test interfaces: nameFixturePair{Schema string `json:"schema"`; ObjectName string `json:"object_name"`}; readNameRegistryFixture() ([]nameFixturePair,error); readNameGoldenFixture() ([]nameFixturePair,[]string,error). Keep decoder/semantic validation callable on in-memory mutated bytes for negative tests. These interfaces must not depend on the future public production type, so Task1 compiles alone.

### Task1: Independent portable name fixtures

**Files:** Create schema_names_test.go and both JSON fixtures from the file map.

**Interfaces:** Consume strictjson.Decode(io.Reader,int64,any) error with limit64<<10; produce the two private readers and nameFixturePair defined above for Task2. Read paths relative to contracts package cwd: ../../../testdata/c12/authority-v7/.

- [x] Write TestAuthorityV7NameFixtureInventory requiring both files, exact schema/coverage, all required non-null members,146 pairs, strict sorting, unique schema/object and exact equality between the independent inventories. It initially fails on missing files, not undefined production APIs.
- [x] Run Appendix A phase fixture-red. Verified build exit0, ordinary runtime exit1 with the named missing-fixture assertion; ordinary completion and final audits passed. Missing compiler/module/native evidence would not count as expected RED.
- [x] Author both literal pair sets from §4.1 (exclude the table header). Check each independently against the source index before production implementation; record pair/count/exclusion review in the result report. Never generate either expected set from production lookup/list, nor populate golden by invoking the registry reader. Their equality test catches divergence, not a jointly copied mistake; source review is mandatory.
- [x] Fill golden rejected_schemas with empty string; every three excluded historical object-name strings; and, for every valid schema, uppercase, leading/trailing ASCII space, appended NUL, and terminal .v1 replaced with .v0/.v2. Deduplicate and byte-sort this literal array. Explicitly include an unrelated "talenro.c12.unknown.v1". Assert none belongs to expected_entries.
- [ ] Implement private test decoding via strictjson plus semantic checks; do not modify strictjson. Write TestAuthorityV7NameFixtureRejectsMalformed with missing/null/extra/duplicate top-level and pair fields, case-altered member names, number/object/array in string fields, null pair, wrong markers, trailing JSON, invalid UTF8, >64KiB, duplicate schema/object,146-entry missing+extra replacement and mismatched pairs. Mutate in-memory copies only; compare variants against the untouched independent expected set, never rewrite tracked fixtures during tests. **Reopened by final review Important1:** element-level null in rejected_schemas is not rejected; follow-up addendum pending.
- [x] Run Appendix A phase fixture-green, selector ^TestAuthorityV7NameFixture. Require build/runtime exit0, each named test PASS, no SKIP/FAIL and complete evidence. Review literal sources and JSON diff.
- [x] Stage exactly these three paths; commit "test(contracts): add independent authority v7 name fixtures". Do not commit expected-red artifacts. Save factual partial status if interrupted; no implementation acceptance yet.

### Task2: Read-only directory and boundary tests

**Files:** Create schema_names.go; extend schema_names_test.go. Do not edit fixtures to match production failures.

**Interfaces:** Consume Task1 readers as oracle. Produce precisely the public type/functions/sentinel in Global Constraints; no other exported declarations or methods in schema_names.go.

- [x] Write TestAuthorityV7NameDirectoryExactPairs: for each independent expected pair, lookup returns exactly both strings and nil; entire list equals expected order/set and length146. Write TestAuthorityV7NameDirectoryRejectsUnknown: all golden negatives return zero struct and err==ErrUnknownAuthorityV7SchemaName; errors.Is succeeds; error text equals the sentinel text for all inputs.
- [x] Write TestAuthorityV7NameDirectoryOwnsResults: obtain two lists, change first entry's strings, swap elements, append; obtain a third list and repeat lookups for all146; second/third and lookups must still equal oracle. Mutate a returned lookup value too. All fields are value strings, no shared backing exposure.
- [x] Write TestAuthorityV7NameDirectoryPublicBoundary: reflection requires exactly Schema/ObjectName exported string fields and no methods; AST of schema_names.go requires exactly the four named exported declarations, exact signatures and no exported mutable table. Require no import of authority, os, encoding/json, reflect or runtime configuration/parser packages in the production file. Its public comments explicitly say membership is not authentication/authorization/evidence eligibility. Review private table/manual code for no callback/registration or consumer wiring; AST is a regression check, not a security proof.
- [x] Run Appendix A phase directory-red. Expect naturally completed build exit1 with undefined new API/type/sentinel diagnostics; no test binary runtime. Other failures stop, not RED credit.
- [x] Implement the exact public declarations in schema_names.go: private fixed146-pair array, exact schema search, unknown zero+sentinel; list allocates/copies the schema-sorted array. Define sentinel once with errors.New("contracts: unknown authority v7 schema name"). No initialization from fixtures, body types or environment. A linear scan is adequate; no additional index/registration abstraction.
- [x] Run gofmt only on the two new Go files using the existing copied SDK tool (no Go package workload). Run Appendix A phase directory-green, selector ^TestAuthorityV7NameDirectory. Require all named tests PASS and complete evidence. Do not lower gates for compile or test failures.
- [x] Review explicit two-file diff and commit "feat(contracts): add read-only authority v7 name directory". Preserve fixtures unchanged and progress accurately labelled pending final regression/review.

### Task3: Full contracts regression, review and recoverable handoff

**Files:** Review all four new files; update plan/status/recovery and the factual result report.

**Interfaces:** Consume Task1 fixtures and Task2 API. Appendix A produces saved ordinary build/runtime receipts, streams, audit bindings and actual phase outcomes; no old leaf completed marker applies to this package.

- [x] Run Appendix A phase contracts-final with selector . against the entire existing contracts package. Default-vet build and all ordinary tests exit0; final line PASS,55 top-level/423 subtest PASS,0 SKIP/FAIL/instrumentation; ordinary guards and independent final audits verified. This is not the entire repository or a fuzz-engine run.
- [x] One fresh read-only final reviewer completed the full approved slice review;146 pairs/ownership/API/no-wiring verified, Critical0/Important1/Minor0. Important1 is test-only null negative-element acceptance. This is not execution authority or acceptance.
- [x] Preserve Important1 and stop for the bounded fix/verification addendum; original five builds/four runtimes consumed, no fix/retry or second reviewer. Other gates unchanged.
- [x] Write factual report with actual scope, phase results/audits, open review finding, source trace, rulings/costs and proposed next decision. Full metadata93/53 and wider gates remain open.
- [ ] Stage only four product/fixture paths plus these necessary plan/report/status/recovery paths as applicable. Commit "docs: record authority v7 name-directory implementation evidence"; verify origin/branch, push normally, compare remote branch SHA to local HEAD and require clean worktree. Report a blocked push as local-only. Another computer starts from this dev tip, not main; raw local evidence/SDK/cache must be independently prepared there.

## Appendix A: Proposed bounded offline verification protocol

This appendix was approved with the plan after6c5ec277, not authority inherited from the consumed leaf001 run. Its five named phases are now consumed and verified; see the factual result. Do not call Invoke-LeafOrdinaryCheck, retarget its fixed plan, replay a matrix, or manually run a historical binary. No additional workload is authorized by saving/resuming this plan.

Use S=W/.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator; new absent root R=S/name-directory-implementation-001 and control R-control. Existing copied SDK is S/uninstrumented-001/toolchain; modules C:/Users/Lenovo/go/pkg/mod. If absent/drifted on another computer, stop for new environment preparation, not download/rewrite pins.

Before workloads, create only an ignored operational wrapper R-control/verify.ps1 with interface Invoke-NameDirectoryPhase([string]Phase). Freeze this plan's package, phase table, paths and argument arrays; no caller overrides. Parse wrapper in PS7 and fresh64-bit Windows PowerShell5.1 without running workloads; review its phase guards/attempt/finalization paths against this appendix. Import unchanged matrix-check.ps1 only to use Invoke-MatrixRecordedProcess(Executable,Arguments,Environment,WorkingDirectory,ReceiptDir,LimitSeconds) and existing ordinary evidence/audit helpers. Do not call Invoke-LimitedMatrix or change any tracked diagnostic helper. If existing guards cannot safely validate this package without changing policy/helpers, stop and report the specific gap before a native boundary.

Revalidate the four exact inputs and71-helper aggregate from [ordinary001 frozen-input table](2026-10-16-leaf-ordinary-check-001.md#frozen-inputs-and-preparation-observations), not its controller/root/package approval. Expected71-file SHA256 is AF83F1057EA2B71E31BCB38BD59F6518C2C68424BE77F1B1FD62A04D17340869. Recompute full copied-SDK/reference audit before/after each phase and inventory product sources. Record permitted new-file bytes per phase; unrelated source changes stop. Old inventory counts are not expectations for the intentionally added files.

Use fresh64-bit C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -NoProfile -NonInteractive directly; record host identity/bitness and avoid PS7-specific module-path inheritance. Exact20-key child environment: GOROOT=<copied SDK>;GOCACHE=R/go-cache;GOMODCACHE=<existing modules>;GOTOOLCHAIN=local;GOENV=off;GOPROXY=off;GOSUMDB=off;GOAUTH=off;GOVCS=all:off;GO111MODULE=on;GOWORK=off;GOFLAGS=-mod=readonly;GOOS=windows;GOARCH=amd64;TEMP/TMP/GOTMPDIR=R/tmp;TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG empty. No extra Go env/version/list/vet/warm-up command.

Each phase uses new R/<phase>/build and R/<phase>/runtime receipt directories; output binary R/<phase>/contracts.test.exe. Persist planned attempt and source/tool/argument/cwd bindings before each boundary; never reuse or delete an attempted phase. Exact build argument array from W:

`test -c -x -p=1 -count=1 -timeout=120s -o=R/<phase>/contracts.test.exe talenro.local/platform/internal/nodecontrol/contracts`

Build native limit600s, default vet (never -vet=off). Runtime, only after accepted exit0 build and compiler/new-binary hash bindings, from W/internal/nodecontrol/contracts:

`-test.count=1 -test.timeout=120s -test.v=true -test.run=<selector>`

Runtime native limit150s. Always unchanged IsolatedFuzzProcess.Run(...,false,false), schema3 completion/query gates,500ms natural observation and shared5000ms cleanup; no diagnostic mode, exception for error87 or relaxed predicate. Full SDK/product/artifact final audits run independently after any earlier failure; preserve primary and secondary errors even if persistence fails. No raw cache/tmp tree walking as substitute for scoped audits.

| Phase | Selector | Expected build/runtime | Meaning |
| --- | --- | --- | --- |
| fixture-red | ^TestAuthorityV7NameFixture | 0/1 | Named missing-file assertion only |
| fixture-green | ^TestAuthorityV7NameFixture | 0/0 | Fixture and malformed-input tests PASS |
| directory-red | not run | 1/not run | Exact undefined new API diagnostics only |
| directory-green | ^TestAuthorityV7NameDirectory | 0/0 | Exact names/ownership/boundary PASS |
| contracts-final | . | 0/0 | Entire contracts ordinary package PASS |

At most five builds/four runtimes, each once. Expected behavioral/undefined-symbol RED still requires ordinary natural completion and audit integrity; it is not a native failure retry exemption. Any unexpected output, timeout, native87, forced termination, identity/query/cleanup/truncation/audit/persistence uncertainty stops forward execution. Save actual evidence; no retry, new-looking root, debugger/PID workaround, privilege/security change or additional workload without a new scoped decision.

Per-phase manifests bind actual receipt/streams/binary when present and source/tool pins. Final factual report must distinguish compile/test exit from ordinary evidence acceptance and post-audit success. Do not fabricate leaf's15-file completed binding or call its saved reader for this new protocol. Keep Acceptance=false throughout; only fully supported contracts-package result may be labelled passed. This operational wrapper is ignored, not a shipped generic runner; portable source/tests/fixtures must not depend on it.

## Plan self-review and approval handoff

Self-review: design §§2–4 map to Tasks1–2; §5 and exclusions map to Task3/global constraints. Five Review Focus cases have named tests/review steps. Fixture types remain independent of production; both limits are64KiB; no new metadata/codec. Expected RED versus unexpected native failure is explicit; historical controllers/approvals are not generalized. Three deliverables, one small API and bounded verification; no program bodies transcribed.

User approved plan/Appendix A and Native/inline after6c5ec277. All five bounded phases ran once and verified, but final review Important1 prevents strict fixture/slice completion. Preserve the implementation as an unaccepted development checkpoint and review the bounded fix addendum before additional workloads. No second reviewer or old-phase replay; broader gates remain open. The final preservation checkbox alone is not remote evidence; actual handoff verifies the pushed SHA.
