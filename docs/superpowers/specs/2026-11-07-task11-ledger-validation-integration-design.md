# Task11 ledger validation and limited integration design

Date: 2026-11-07. Source baseline: af008736.
Status: approach selected; written design awaiting user review. No implementation,
test budget, native run or acceptance activation is authorized by this document.

## Intent and existing contracts

Resume the metadata-registry work through a separately verified process-evidence
path, without treating its old failed fixture-red as accepted or replacing every
ordinary runner. This is a rollout design for existing ledger code, not a new
identity algorithm or a claim to have fixed Error87.

Sources: [ledger design](2026-10-17-ordinary-member-identity-ledger-design.md),
[native proposal](../../roadmap/2026-10-17-ordinary-ledger-native-validation-proposal.md),
[Stage B prerequisites](../../roadmap/2026-10-17-ordinary-ledger-stageb-readiness.md),
[stopped metadata plan](../plans/2026-10-17-authority-v7-metadata-registry.md), and
[current recovery boundary](../plans/2026-11-07-task11-recovery-boundary.md).
Older readiness statements are historical; a new source/evidence comparison must
identify which findings later checkpoints closed, not repeat them all as defects.

Current source facts:

- `matrix-check.ps1:29` still invokes ordinary `IsolatedFuzzProcess.Run`.
- `process-native.cs:337` exposes the separate `RunWithOrdinaryIdentityLedger`.
- `ordinary-ledger-receipt.ps1` validates schema4 evidence and always returns
  `Acceptance=false`; internally consistent caller JSON is not execution provenance.
- `run.ps1:84` onward deliberately rejects ledger markers and diagnostic modes.
  That rejection is a rollout boundary, not a defect to delete.

## Approaches and chosen scope

The user selected staged verification followed by Task11-only integration.
Existing public completion gates and matrix callers remain unchanged. Compared
with a global switch, this limits compatibility risk and makes the enabled scope
auditable. A global migration may be considered later, not inherited here.
Directly removing the schema4 rejection or promoting diagnostic receipts is not
an alternative: neither supplies missing native evidence or workload authority.
Keeping Task11 blocked remains the required fallback if prerequisites cannot close.

## Stage 1 — finite evidence reconciliation

Produce one gap table for the existing implementation, not another diagnostic
framework. Each row records requirement, current source location, applicable
saved evidence/source binding, remaining gap and owner. Reuse historical evidence
only for its stated source and scope; fake checks are not native observations.

Cover: acquisition/ownership/retained-handle identity, sticky refusal, sampled
count reconciliation, clocks and teardown, schema isolation, serialized receipt
checks, source/executable binding, outer containment/raw output/finalization,
and security disposition. Distinguish an actual uncovered requirement from a
changed source requiring targeted regression. Do not repeat all old test budgets.

Deliver one bounded implementation/verification plan with named cases and fixed
maximum calls before execution approval. If a necessary security-owner/vendor
disposition or safe outer runner is unavailable, record the concrete blocker and
stop. No WPR, registry-access, administrator or cache-verification detour follows.
Existing historical detection is not relabelled false positive by this design.

## Stage 2 — native evidence, separately authorized

Reuse the existing root-only and controlled-parent/child probe questions, not the
old Go metadata binary. At most one native entry per case; child case follows only
after a complete successful root case. Proposed ceilings remain root 5s, child
20s, outer 120s per case, shared natural window 500ms and cleanup ceiling 5000ms.
These are ceilings for a later exact-input plan, not a granted allowance or a
promise to preempt an uninterruptible native call. No automatic retry or transfer.

The execution packet must bind a reviewed source revision, executable/source
hashes, actual arguments/environment/cwd, fresh owned output/control identities,
raw output capture, outer ownership/cleanup, and durable finalization. Do not
reuse historical identities merely because they once appeared absent. Refresh
only explicitly authorized readiness facts; no installs or security-setting edits.

Preserve limited rights, one acquisition per candidate, same-run owned handles,
sampled coverage and the original cap32 refusal semantics. Required results are
natural root/member completion, successful releases, consistent final accounting
and empty final membership, no timeout/forced termination/sticky failure, complete
source/output/receipt bindings and final saved validation. For the child case,
markers corroborate the controlled child but cannot replace handle identity or
prove child/root exit order. A missing observation is failure, not an exception.

Both probe results remain `Acceptance=false`. Two successful probes neither prove
complete sampling nor accept the old RED, prove Error87's cause, clear broader
security/physical gates, or authorize a metadata workload. Failed/unknown cleanup,
missing output, persistence failure, security alert or any gate failure stops.

## Stage 3 — explicit Task11-only rollout, separately reviewed

Only after Stage 2 evidence is reviewed may a later implementation add a tracked
Task11-specific adapter/controller. It invokes the ledger entry explicitly; it
does not modify `Invoke-MatrixNative`, ordinary Run overloads, or the public
`Assert-ProcessCompletion` rejection. No environment toggle or global fallback.

The adapter accepts only the exact metadata phase/package/selector allowlist and
source-bound execution packet in the revised metadata plan. It must combine:
(1) actual execution provenance and durable artifacts; (2) full schema4 evidence
evaluation; (3) phase-specific build/output/exit requirements and unchanged audits.
A caller-supplied boolean, complete synthetic receipt, diagnostic observation or
matching PID cannot satisfy these conditions. Preserve the pure evaluator's
`Acceptance=false`; report scoped phase status separately from overall acceptance.

Default behavior is refusal until an exact rollout decision and finite runtime
packet are approved. Test the adapter with controlled fake inputs before enabling
it: wrong scope/source/run, forged provenance, diagnostic contamination, schema
downgrade, incomplete/forced exit, persistence failure, and expected-RED versus
unrelated failure. Verify old callers still reject schema4. Exact interface and
test names belong in the implementation plan, not an unreviewed live script.

The first eventual metadata workload must re-establish the unsatisfied fixture-red
prerequisite under a newly approved identity and budget. This explicitly revises
the old runner selection; it does not replay metadata001 or retroactively accept
its output. Only accepted RED permits subsequent fixtures/API work. Old unused
four-build/three-runtime slots are not borrowed automatically.

## Non-goals and stop boundary

No full matrix, Docker/DB, product registry implementation, canonical policy
registration, general runner replacement or main merge. No new process rights,
weaker safety predicate, PID reopen, diagnostic-to-ordinary conversion, security
exclusion, quarantine restore, account change or retry-until-pass behavior.
No new observation count is claimed; 44/34/6/217 and NOT_ACCEPTED remain unchanged.

## Review and recovery

Design review checks that no stage's success grants the next stage's authority,
that small probes are not treated as general proof, and that the Task11 adapter
cannot become an alternate arbitrary-workload acceptance API. Necessary design,
plan, tests, adapters and recovery notes travel with the development branch;
private caches/binaries/consumed control roots do not.

Next after written approval: use writing-plans for Stage 1's single gap table and
bounded follow-up plan. Do not implement or run Stage 2/3 from this design alone.

Independent read-only builder check returned "nothing important left" for
planning questions. This is design-clarity feedback only, not a code/security
audit, runtime evidence or permission to activate acceptance. Author review also
confirmed unchanged public gates, distinct approvals and preserved failed history.
