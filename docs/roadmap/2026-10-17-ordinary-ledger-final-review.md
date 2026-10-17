# Ordinary ledger — independent readonly final review: fixes required

Author closeout found an additional failed-initial-snapshot provenance gap in B;
see linked repair report. Implementation remains incomplete despite scoped green
regressions. No second independent review or native validation is claimed.

Follow-up:[single corrective pass verified in fake/static scope](2026-10-17-ordinary-ledger-review-repair.md).
All six findings have grouped regression coverage;final14 runs/350 assertions
green.29/29 repair calls used. This is not a second independent review,native
acceptance or merge readiness. Original findings/verdict below are historical.

Date:2026-10-17. Reviewed89cb1ad09bf831216c5999ddcb4409e4cea4d18d through
36367122d5e7254238efc7f81baaf9ddcd77a5eb. One fresh-context independent reviewer,
as authorized for Native/inline execution. No reviewer edits,tests,native calls,
network,Git mutations or further delegation. Author checked cited source paths
and accepts all six findings;no corrective code or verification has run yet.

Verdict:NOT IMPLEMENTATION COMPLETE;NOT READY FOR MERGE/ROLLOUT.
The14 final fake/static runs/350 assertions remain valid scoped observations,
but did not cover these defects. Budget34/34 consumed;no retry allowance remains.
No native target/metadata continuation occurred. Do not run existing ordinary
drivers from this checkpoint before resolving finding1:they can execute a
workload then reject its receipt due to the introduced compatibility defect.

## Findings accepted after source verification

### 1. P1 — default-null new fields break existing ordinary receipt workflows

run.ps1:65 refuses new property presence,including null. matrix-check.ps1:41
and leaf-ordinary-check.ps1:191 copy all native Result fields except the existing
null CompileStageTrigger exception. New null MemberIdentityLedger,
LedgerFinalOutcome and RunIdentity therefore appear on legacy ordinary receipts
and trigger refusal. leaf-ordinary-check.ps1:89 additionally derives required
receipt shape from current Result.GetFields,so old receipts acquire new mandatory
fields. The new common serializer test did not exercise these other writers.

Required correction:omit absent new-mode fields consistently from ordinary
writers and keep legacy schema-specific shape contracts stable. Preserve refusal
of genuine new-mode/downgrade markers. Add pure regression for all affected
writers and old saved-shape fixtures,not a real workload replay.

### 2. P2 — ordinary initial shutdown queries bypass sticky ledger failure tracking

process-native.cs:453 calls ReadInitialJobQueries before the new session. Its
MembersWithEvidence starts capacity16 and retries incomplete/ERROR_MORE_DATA
results up to8 times without checking the ledger deadline. With17–31 remaining
members,a failed/truncated attempt can be followed by success without marking
the ledger incomplete. This violates the proposed single bounded query authority.

Required correction:route the new mode's initial shutdown queries through the
deadline-aware ledger contract. Existing legacy modes retain their own behavior.
Test truncation/error and overrun so no subsequent empty query can erase them.

### 3. P2 — root-only final query timestamps lack shutdown lower bound

ordinary-ledger-receipt.ps1:81 validates final accounting/list upper bounds and
orders them against entry finals inside a loop. For zero entries,that loop does
not run;pre-shutdown final queries with internally ordered timestamps can pass.

Required correction:require both final queries to start within the shutdown
interval independently of entry count. Add a valid root-only fixture and stale
accounting/list mutations,retaining monotonic and shared-window checks.

### 4. P2 — new policy-only marker can reach no-schema legacy acceptance

run.ps1:64 recognizes three new properties but not the new policy value
ordinary-owned-ledger-final-exit-v1. If schema/object markers are missing while
that policy remains,the legacy branch can return success when its old cleanup
fields pass. This is an explicit-marker gap,not arbitrary-JSON authentication.

Required correction:reject that policy marker before schema dispatch;test it
with missing schema and with older schemas,without weakening genuine legacy paths.

### 5. P2 — evaluator ignores legacy diagnostic-observation containers

ordinary-ledger-receipt.ps1:48 rejects several diagnostic properties but never
examines MemberEvidence/ShutdownMemberEvidence. Injecting a DiagnosticShutdown
row there can leave the otherwise valid schema4 envelope EvidenceComplete=true.
The ordinary acceptance gate is still closed,but pure evidence isolation is wrong.

Required correction:require schema4 legacy observation containers to be empty
when present,or reject diagnostic content explicitly. Add both container mutation
cases and retain valid new-mode serialization coverage.

### 6. P2 — existing source bindings omit newly loaded dependencies

run.ps1:4 now loads ordinary-ledger.cs,ordinary-ledger-session.cs and
ordinary-ledger-receipt.ps1. Active inventories including
leaf-ordinary-check.ps1:32 and retained-child-observation.ps1:58 omit those files.
Their drift/source-binding claims no longer cover everything the runner loads.

Required correction:inventory all active source-binding producers/consumers and
include the new loaded dependencies consistently. Do not rewrite historical
receipts or reuse old bindings to authorize runs. Add source-drift rejection
tests using portable fake/filesystem fixtures only.

## Strengths and limits

Reviewer confirmed exact owned-handle continuity,Error87 preservation,no candidate
reopen,and release marking before close (no retry after close failure). Count
mismatch refuses even for an empty final list;checked deadlines,backward clocks,
strict Boolean validation and explicit new entry address important core risks.
The evaluator always returns Acceptance=false. These strengths do not waive the
six findings,which are all blocking for this unfinished development slice.
No Critical finding or deferred Minor was reported.

## Declined areas and author rulings

- Ruling:actual Windows scheduling,handle/accounting effects and Error87 recurrence
  remain unverified because no native run was authorized. Cost:the native adapter
  or observer behavior may still fail;fake tests cannot establish a live fix.
- Ruling:workload acceptance,metadata continuation and rollout stay unopened per
  user scope. Cost:this checkpoint is not deployable or proof of project readiness.
- Ruling:cryptographic authenticity after arbitrary JSON rewriting is outside the
  receipt contract. Cost:the validator does not authenticate hostile rewritten
  records;explicit-marker validation must still be corrected under finding4.
- Ruling:the allowlist is not full-project regression;no broader test success is
  claimed. Cost:unlisted callers may have additional regressions,as finding1 shows.

All earlier execution rulings remain in the implementation report/ignored ledger.
No review recommendation is silently discarded and no second review was dispatched.

## Proposed bounded repair round — awaiting user approval

One corrective pass covering these six findings,with26 additional fake/static
suite invocations maximum:three new grouped regression suites ×RED/GREEN ×two
hosts=12,then the existing7-suite final allowlist ×two hosts=14. New proposed
groups:ordinary serialization/shape/source-binding compatibility (findings1/6),
deadline-aware query/final interval boundaries (2/3),and receipt mode isolation
(4/5). Preserve all rights,32-member cap,500ms/5000ms budgets and schema4 refusal.
No native target,Go build,old controller replay,SDK changes or rollout.

This is a proposal only,not an approved new test allowance or a completed repair
plan. Before executing,write exact fixtures/commands and scope any extra active
binding files against these findings. Any further unexpected failure stops;
no automatic retry or new review cycle. Push portable repair progress as usual.
