# Authority V7 Null Negative-Vector Fix Implementation Plan Addendum

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans with the user's preserved Native/inline method. This is the original final review's ONE fix pass, not a second implementation/review cycle. Steps use checkbox (`- [ ]`) syntax. Await user approval of this additional bounded verification before tests or fixes.

**Goal:** Reject JSON null negative-vector elements while preserving legitimate empty-string rejection and all existing directory behavior.

**Architecture:** Preserve null information in the private test-only golden wire shape, reject nil elements before extracting strings, and leave the public API/production table/fixture literals unchanged. Verify the final review's Important1 through an observed RED, fixture GREEN and complete contracts-package regression; no second reviewer.

**Tech Stack:** Existing Go1.26.5 copied SDK, standard encoding/json via unchanged internal/strictjson, Windows PowerShell5.1 ordinary native collection, existing offline module cache; no new dependencies.

**Spec:** [Approved name-only design §4](../specs/2026-10-16-authority-v7-name-directory-design.md), [original plan](2026-10-16-authority-v7-name-directory.md) Task3 stop rule and [Important1/source trace](../../roadmap/2026-10-16-task11-name-directory-implementation.md#final-independent-review--important1-open).

**Status:** Proposed after the original five-build/four-runtime budget was consumed. Source-traced defect, not yet runtime-reproduced or fixed. Native/inline execution method is preserved; only this additional fix/verification scope needs approval.

## Global Constraints

- Keep W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, codex/c12-b01-task9-coordinator and origin https://github.com/win202569/Talenro.git. Start from latest saved development tip, not main.
- Modify only internal/nodecontrol/contracts/schema_names_test.go and necessary plan/report/current-status/recovery documents. Freeze schema_names.go and both tracked JSON fixture bytes; keep146 exact pairs,881 literal rejection strings and the four-declaration public API unchanged.
- Do not change shared strictjson/diagnostic helpers, Go SDK, schemas, dependencies, canonical spec-set, roles/evidence/authorization/codec/provider wiring, process/query/type/completion predicates or security settings.
- No replay of the original five phase identities, extra native retries, installs/downloads, Docker/Bash/WSL, fuzz engine/race/benchmark/DB, PR/main merge, force push or branch deletion.
- Follow TDD; no fix before observed new regression RED. Scope remains contracts-package ordinary verification, not whole-repository or production acceptance. No second independent reviewer: resolve the existing review's one Important finding with covering RED/GREEN plus full package regression.

## Review Focus

1. First negative element null is currently indistinguishable from valid empty string: new null-rejected-schema-first mutation must fail before the fix and reject afterward.
2. Null elsewhere in the array must also reject: add null-rejected-schema-last, even where existing sorting happens to reject it already.
3. A genuine empty string must remain valid negative-vector data: existing complete inventory/881-string equality and unknown-input tests must remain green.
4. Non-string numbers/booleans/objects/arrays must not be stringified: add negative-element-number/bool/object/array mutations.
5. Fix must not alter146-pair matching, ownership, public boundary or other contracts: unchanged literals and complete contracts regression verify the boundary.

### Task 1: Repair test-only null preservation and verify once

**Files:** Modify internal/nodecontrol/contracts/schema_names_test.go; update original plan, this addendum, docs/roadmap/2026-10-16-task11-name-directory-implementation.md, docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md.

**Interfaces:** Keep decodeNameGolden([]byte) ([]nameFixturePair,[]string,error), readNameGoldenFixture() ([]nameFixturePair,[]string,error), all public APIs and marker/64KiB contracts unchanged. Only private nameGoldenFixture.RejectedSchemas changes from []string to []*string to retain primitive null information. No exported helper or runtime JSON import API.

- [ ] Add six in-memory golden-only mutations to TestAuthorityV7NameFixtureRejectsMalformed: null-rejected-schema-first (replace element0 with nil), null-rejected-schema-last (replace final element with nil), and negative-element-number/bool/object/array (replace element0 with42/true/empty map/empty array). Keep tracked fixture files unchanged. Existing TestAuthorityV7NameFixtureInventory pins the valid empty string and entire881-vector set.
- [ ] Run new null-red phase below with unchanged decoder. Expected verified build0/runtime1, specifically the null-rejected-schema-first subtest's "malformed fixture accepted" failure. Other failures, native errors, timeout or audit uncertainty stop; this is not evidence of the production directory failing.
- [ ] Change only nameGoldenFixture.RejectedSchemas to []*string. In decodeNameGolden, reject every nil element, then construct []string preserving exact values/order; perform existing membership/strict ordering checks on that slice and return it. Preserve nonnil pointer-to-empty-string as ""; no default, trim, coercion or global strictjson change.
- [ ] Format only the modified test file with the existing copied SDK gofmt tool. Run null-green phase, selector ^TestAuthorityV7NameFixture; require all existing/new fixture tests PASS, exit0 and all ordinary/full audit guards.
- [ ] Run contracts-final phase with selector . on the complete contracts package; require default-vet build0/runtime0, final PASS, observed top-level tests, no SKIP/FAIL/instrumentation and all independent audits/bindings. Record actual names/counts/times, not a guessed new count. Confirm production file and both fixtures are byte-identical to the pre-fix commit.
- [ ] Mark Important1 resolved only after the named regression's observed RED→GREEN and full package verification. Record the existing reviewer and its two scoped exclusions/rulings; do not request another review or expand into metadata/codec work. If any phase is unexpected, save an explicitly unfinished checkpoint and stop; no second fix pass/retry under this addendum.
- [ ] Review explicit file list, commit the test-only fix plus necessary result/recovery documents, push normally to the existing development branch and verify remote SHA=local/clean. Keep raw evidence/control/SDK/cache/binary ignored. No main merge or overall acceptance.

## Additional bounded verification — proposed, not yet authorized

Let S=W/.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator and new R=S/name-directory-null-vector-fix-001; control R-control. Both were absent when this proposal was written; recheck safe paths/absence after approval. Do not create either while awaiting approval. The old original phase records remain untouched and consumed.

After approval only, create ignored control/verify.ps1 via apply_patch with frozen entry Invoke-NameDirectoryNullFixPhase([string]Phase). Adapt the prior operational wrapper to these three fixed phases, not its old identity; parse in PS7 and fresh64-bit Windows PowerShell5.1 before any workload. Keep the unchanged shared ordinary collector/guards and independently revalidate fixed arguments/cwd/environment/path/source/tool bindings. No fixed leaf controller or saved reader/15-file marker reuse.

Reuse the original plan's four exact input SHA256 pins and71-helper aggregate AF83F1057EA2B71E31BCB38BD59F6518C2C68424BE77F1B1FD62A04D17340869; full copied-SDK/reference and product inventories before/after each phase, independent fixture/tool/pin/artifact final audits even after failure. Bind actual attempts durably before each native boundary. No native/persistence error preservation beyond actual available evidence may be claimed; an incomplete/failed record is unverified, never PASS. Existing shared-helper simultaneous native/error-write failure limitations remain unfixed/outside this test-only repair.

Use existing SDK=S/uninstrumented-001/toolchain; modules C:/Users/Lenovo/go/pkg/mod; fresh64-bit C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -NoProfile -NonInteractive, avoiding inherited PS7 module paths. Child environment exact20 keys: GOROOT=<SDK>;GOCACHE=R/go-cache/<phase>;GOMODCACHE=<modules>;GOTOOLCHAIN=local;GOENV=off;GOPROXY=off;GOSUMDB=off;GOAUTH=off;GOVCS=all:off;GO111MODULE=on;GOWORK=off;GOFLAGS=-mod=readonly;GOOS=windows;GOARCH=amd64;TEMP/TMP/GOTMPDIR=R/tmp;TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG empty. No extra Go env/version/list/vet/warm-up commands.

Exact build arguments from W, each native limit600s/default vet:

`test -c -x -p=1 -count=1 -timeout=120s -o=R/<phase>/contracts.test.exe talenro.local/platform/internal/nodecontrol/contracts`

Runtime only after accepted build/compiler/new-binary bindings, from W/internal/nodecontrol/contracts, native limit150s:

`-test.count=1 -test.timeout=120s -test.v=true -test.run=<selector>`

| Phase | Selector | Build/runtime exits | Required result |
| --- | --- | --- | --- |
| null-red | ^TestAuthorityV7NameFixture | 0/1 | Only newly added first-null acceptance regression fails |
| null-green | ^TestAuthorityV7NameFixture | 0/0 | Existing/new fixture cases PASS |
| contracts-final | . | 0/0 | Complete contracts package PASS |

At most three builds/three runtimes, each once, in new phase receipt directories. Always unchanged IsolatedFuzzProcess.Run(...,false,false), typed schema3/query/completion checks,500ms observation and5000ms cleanup. Any unexpected result, native87, timeout, forced termination, identity/truncation/query/cleanup/audit/persistence uncertainty stops before another phase. Preserve primary/secondary information actually available; never rewrite old evidence or relax a predicate. A different computer/pin mismatch needs separate preparation, not automatic downloads or old-root replay.

## Self-review and handoff

All five focus cases belong to Task1; private wire pointer representation preserves null without changing outward signatures or genuine empty-string behavior. Literal sets/public production file remain frozen. Three phases supply regression RED, fixture GREEN and complete package GREEN within explicit budgets; no hidden retry or second reviewer. This addendum closes only Important1 if verified, not other deferred C12 gates. Await user approval; no new test/fix/native workload has run under it.
