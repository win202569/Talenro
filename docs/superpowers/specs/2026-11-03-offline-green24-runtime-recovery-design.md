# GREEN24-backed runtime recovery design

Status: DRAFT / PENDING_DESIGN_APPROVAL. Documentation only.
Baseline:1a9b160c71f6a397e7ed1779ce24381abcc0efa7.
The user selected a new versioned recovery entry and asked for this design.
Names and the finite allowance below are proposals, not execution approval.

## Intent and approaches

Recover the path to one original52group contract host after the local test
correction passed. Preserve the distinction between historical evidence and
verification of newly edited source.

Selected: add one explicit recovery epoch to the existing driver/capture/model
binding path. Keep the original suite and process owner. This needs coordinated
binding edits but preserves the existing safety and evidence checks.

Rejected alternatives: replacing GREEN23 with GREEN24 or deleting its stop would
rewrite history and reopen a consumed route; an independent copied runner would
duplicate the safety-critical execution path. Neither fits the chosen direction.
This is a bounded binding extension, not a new runner architecture.

## Actual baseline

GREEN24:20PASS/0FAIL/0NotRun/exit0; all errors null. Receipt SHA256:
AB86CCC7042089B0D705E9FB5933B4ADD92045514B5B976068315693F7CCFC8A.
Its immutable amendment SHA256:
FA9A9A82DFC135A2284B62C341F1D1B16B47ABFC3ADCDDC42DF39D765F1A99E2.
The31live source hashes still match that amendment. All13old stops remain.
No history-contract Prepare, frozen manifest, snapshots or host intent exists.
Cumulative consumption:24auxiliary/4hosts/3compilations/186snapshots.

GREEN24 proves the unchanged candidate's local20group checks, including the
corrected AST collection handling. It does NOT verify source edited after it,
the new runtime receipt validator, a completed freeze, or a live host.

## Proposed fixed identities and scope

L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner.

- Epoch:history-contract-runtime; Prepare grant:history-contract-runtime-rebind.
- Semantic host Grant:contract-green; physical slot:history-contract-runtime-host.
- P/history-contract-runtime.intent.json, .frozen.json and .stop.json.
- P/snapshots/history-contract-runtime; capture/unit use the new host slot.
- Auxiliary slots:history-contract-runtime-binding-red/green.
- Separate runtime approval/input/candidate files; never reinterpret the existing
  history-contract-v2.input.json amendment as a runtime grant.
- No free-form epoch/path selection, generic ignore-stop flag or old-route retry.

Only the existing four operational binding files may change:driver.ps1,
capture.ps1,source-model.ps1,observation-contract.tests.ps1. Other27operational
files remain byte-identical. Preserve Driver-VerifyHistory, command-error extent
logic, pure validators,49suite cases and3recorder groups, their order/assertions,
and the owned-process tail. Existing historical constructors retain their paths.

## Evidence and source-epoch separation

New read-only validation must authenticate the actual GREEN24 amendment, test,
intent, final and complete tool/session records:computation24/ceiling24,
20ordered PASS, exit0, null errors and exact source/history/pre-post identities.
Retain GREEN23 as failed historical evidence and RED22 as original failing proof;
neither is a successful new-runtime receipt.

The old validator is pinned to original input/candidate hashes and compares live
source. Calling it unchanged after editing the driver would reject the new
candidate. Do not bypass that check or re-label old hashes: preserve the31source
inventory and four baseline file bytes from GREEN24 in immutable new inputs.
Authenticate old receipts against their own saved epoch; authenticate the new
candidate against the current31live source hashes.

Keep historical GREEN24 provenance and new RED25/GREEN26 provenance in distinct
named maps. Both are consumed by the real gate, not merely recorded. Shared
read-only validation must be directly exercised by new tests and by the actual
gate call; no acceptance based on caller-supplied validity booleans.

The last successful parent remains repair-green.frozen.json, SHA256:
A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
Its writer parent remains:
182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461.
Inherit the original2276pins and recovery maps, replace only allowed operational
hashes and append unique immutable provenance. The implementation plan MUST
enumerate every appended path and calculate the exact final count;2319is the old
unrealized proposal, not a reusable new total. No self/future-result pin cycles.

## Proposed finite verification and execution

Propose two additional auxiliaries RED25/GREEN26, ceiling26; one new Prepare.
Explicitly reactivate only the previously unused1host/31snapshots for this new
route, not another duplicate allowance. No child targets, compilations or
reviewers; legacy downstream3hosts/7targets/conditional1review remain frozen.
Maximum cumulative consumption if successful:26/5/3/217.
These permissions require an approved implementation plan before execution.

Freeze complete test/helper/input identities before RED25. Test the actual missing
epoch/slot behavior first. GREEN26 must exercise actual GREEN24 authentication,
negative receipt/source/approval mutations, real relative/external history checks,
parent/writer lineage, namespace separation, non-adoption of consumed attempts,
scalar AST call selection, actual gate-call wiring and unchanged safety/suite
boundaries. Reuse the corrected node-enumeration convention, not the failed
comma-wrapped helper. The plan fixes exact groups, signatures and dispositions;
no replay of22/23/24 and no fake successful receipt or manifest.

For each new fixed action acquire an exclusive CreateNew intent, with a locally
owned random token, before engine/resource checks. Caller-supplied ownership or
an existing intent is not reusable. Preserve new stop/partial records on failure.
Host entry must authenticate the completed Prepare intent/token, manifest,
31snapshots and exact new candidate/receipt provenance before starting.

After exact GREEN26, perform one Prepare and verify31snapshots plus all pins.
Archive and push the completed freeze, verify remote equality, then run at most
one original contract host. Keep pinned runtime/assembly/privilege checks,
120000ms host cap,2000ms exit/drain bounds, evidence/journal bounds, zero child
targets, five false capability flags and cleanup limited to the owned process.

Accept this bounded host result only with original52ordered PASS, exact original
stdout/stderr/exit, parent Exact/EvidenceValid, post-pins intact and no timeout,
termination, overflow or rescue. It is still not overall product acceptance.

## Stop and recovery contract

First unexpected preparation, execution, identity drift, freeze or recording
result stops all remaining work; no in-place repair/retry. Even full success ends
after this one host and preservation. Keep all13old stops and any new one.

Preserve actual calls/results/partials in versioned portable preparation,
candidate, freeze and final archives; decode-check bytes/hashes/lengths/directories
and zero-byte files. Commit explicit necessary sources/tests/docs/archives to
codex/c12-b01-task9-coordinator, ordinary push to the verified origin, and confirm
remote SHA equals HEAD. No main merge, force push, deletion, install or secret
publication. Update current-status and repository-recovery at each handoff.

## Approval boundary

This short design records the selected bounded extension. No new tests, source
edits, freeze or host are performed by drafting it. Approval of this document
permits implementation planning next; it does not itself authorize execution.
