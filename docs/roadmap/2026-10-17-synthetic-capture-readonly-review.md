# Independent readonly readiness review — complete, acceptance blocked

Later dependency evidence: the user-approved [six-run regression packet](2026-10-17-synthetic-capture-dependency-regression.md)
completed81/106/110 on each engine, with administrator old-store and ordinary
simulation/contract contexts. This supplies the dependency component of
Important1, not closure of all acceptance gaps or either old conditional review.
Historical findings below remain as reviewed; no second reviewer dispatched.

Subsequent evidence: user approved the bounded test extension after fe77ddeb.
Its [four-run results](2026-10-17-synthetic-capture-supplemental.md) now directly
exercise the targeted Important2/3 cases on both engines. This historical review
is unchanged; the new test file has no second independent review. Important1
and other recorded limitations remain; no acceptance or gate promotion follows.

## Authority, identity and snapshot

After 562ec07e, the user approved the proposed NEW one-review readonly scope.
Exactly one fresh-context reviewer, `synthetic_store_readonly_review`, reviewed
the store, tests, approved design/plan, relevant dependencies and saved evidence.
No reviewed script/test was executed, no network used, no files changed by the
reviewer and no further reviewer delegated. Parent performed readonly checks of
the cited locations and preserved this report; no repair or retest occurred.

Base: `aeb6fe363ac50b60e929f0123f762a286849a128`.
Reviewed head: `562ec07eeeafbb6014bf66400b0753d0b6d3906d`.
Product SHA256: `508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Test SHA256: `406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`.

This is the newly approved diagnostic/readiness review, NOT completion of the
original implementation review or conditional all10-success partition review.
One review allowance is spent; no automatic fixes, second review or extra hosts.

## Reviewer findings

### Strengths

- Fixed20/17/13-field layouts match the design. Lengths/counts, strict UTF8,
  canonical snapshots, integer ranges and boolean bytes have bounded validation.
- Descriptors precede field access; scalar checks avoid caller GetType/ToString.
- Reader retains stored lifecycle/binding/authority facts, then contract-validates
  them rather than silently replacing them with recomputed values.
- Minimal wire payload and marker expectations are independently hand-encoded.
- CreateNew, same-handle link/attribute checks, read-sharing restrictions,
  marker-last publication and stopped-reader rejection are present.
- Documentation distinguishes consistency, provenance, execution and acceptance.

### Critical

None confirmed by static inspection.

### Important 1 — required acceptance evidence remains incomplete

Classification: validation/readiness gap, not a confirmed code defect.
Reference: [coverage summary](2026-10-17-synthetic-capture-coverage-summary.md),
remaining-gaps section (line44 at reviewed snapshot).

Treating separate assertion groups as accepted implementation would overclaim
compatibility and unmet execution gates. Dependency targets81/106/110 per engine
remain unrun in this implementation packet. Original PS5 boundary timeout and
ordinary-context reader fixture failure remain failed. Administrator PS5 reader
child exited0 with45 assertions but its strict empty-stderr gate failed on584
characters of progress; ExpectedResult=false remains correct. Both old reviews
are unstarted. PS5 missing-root behavior lacks a dynamic result.

Recommendation: preserve the unfinished disposition; any further verification
needs new scope, not frozen-slot reuse or retrospective receipt promotion.

### Important 2 — direct malformed-event rejection cases are absent

Classification: test coverage gap, not demonstrated decoder failure.
Reference: [tests](../diagnostics/isolated-fuzz/synthetic-capture-store.tests.ps1),
line121; production checks in [store](../diagnostics/isolated-fuzz/synthetic-capture-store.ps1),
lines126 and147..151 at reviewed snapshot.

The malformed-payload loop clones a zero-event minimal literal. It does not
directly construct archives with unknown event-kind/channel enums, negative
event elapsed time, or cumulative on-disk chunk bytes above the trace bound.
Those inputs are explicitly required rejection branches. Production checks
exist; no static bypass was found, but the current inventory does not dynamically
demonstrate them. Admission-side oversized objects do not replace decoder tests.

Recommendation: record the gaps and use independently encoded event-bearing
fixtures only in a newly approved bounded test scope. No tests added here.

### Important 3 — writer failure preservation lacks direct behavioral evidence

Classification: validation gap, not an observed error-handling defect.
Reference: [store](../diagnostics/isolated-fuzz/synthetic-capture-store.ps1), line223.

The post-directory-creation catch attempts a stopped sentinel and bare rethrow;
the nested catch suppresses sentinel-write failure. This is statically appropriate,
but tests manually construct stopped/partial archives and the existing-target
test fails before directory creation. They do not directly trigger this catch
or nested sentinel-write failure. Original-error preservation and best-effort
sentinel publication therefore lack direct behavioral evidence.

Recommendation: retain as unverified handling. A controlled failure-state test
requires a new approved scope; do not change permissions or terminate real
capture processes to manufacture it. No fault injection or test execution here.

### Minor and additional limitations

No confirmed minor code defects. The reviewer additionally records:

- Independent literal covers a minimal incomplete record, not a separately
  encoded populated event trace or optional package.
- Junction and payload/marker hardlinks are covered, not file-level reparse.
- No independent recursive comparison of every populated field; some assurance
  comes from the shared production contract verifier.
- Old-reader cross-format assertion is path rejection, not new-payload parsing.
- Successful maximum-byte fixture is incomplete, not maximum duplicated package
  snapshots. These limits are evidence gaps, not confirmed product failures.

### Evidence disposition

97 PASS occurrences per engine comprise87 content assertions plus10 repeated
API/completion sentinels, not87 unique requirements or a unified suite pass.
Admission/byte/event are ordinary-context; readers are administrator-context
with the PS5 qualification above; later format is normal-user and uses its own
prior-approved progress-aware policy. Duplicate labels/partial runs add no new
coverage. Exact timeout/denial causes are not established.

### Declined to judge

- Hostile administrator/filesystem races and concurrent caller mutation:
  explicitly excluded threat models.
- Signing/provenance/encryption/coordinated tampering: consistency hashes do
  not promise those properties.
- Power-loss durability, transactions and hardware persistence: not promised.
- Live capture, OS timing, watchdog/stop/cleanup: outside this offline module.
- Exact timeout/hardlink-denial causes or universal admin requirement: insufficient
  causal evidence; no new diagnostics authorized.
- Automatic migration/future contract compatibility: separate version design.
- Vendor/security clearance, Stage B/I2/I3/full C12 and production acceptance:
  outside this review authority. Parent retains these exclusions and gates.

## Assessment and parent disposition

Reviewer: source is statically coherent with the approved offline design; no
confirmed product defect within inspected scope. Readiness incomplete. Ready
to merge: **No**. No tests were run by this review, so this is not a new dynamic
success result or proof of absence of defects.

Parent readonly checks confirmed the cited event guards, zero-event mutation
loop and writer catch/test structure. All three Important findings are retained
as unresolved validation gaps; none is silently downgraded into acceptance or
treated as an implemented fix. No product/test changes are justified solely by
this static report. Next recommend a bounded test-extension design addressing
the direct event-decoder and writer-failure gaps, with explicit failure injection
limits and future host budget before implementation. Dependency regressions
and remaining environmental gaps still require their own closure.

## Recovery and unchanged gates

Postreview product/test hashes above match pre-review values. Partition index
B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1,
reader index080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211,
format index5FA944DF33467C084ED50260CAA2201C77EF4341F482CFBD3E12163D71D52600
are unchanged. Original6/12 and partition7/10 remain spent with remaining slots
frozen; later2/2 reader and2/2 format allowances are spent. Prior raw receipts,
fixtures and local review ledger remain retained/ignored, not transferred by
Git. Portable report/status/runbook are preserved on the development branch.
No main merge or security/full acceptance follows; cloning grants no new
execution authority. Overall implementation remains unaccepted.
