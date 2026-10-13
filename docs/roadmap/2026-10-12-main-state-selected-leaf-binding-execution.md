# Main-state selected-leaf binding — local execution record

Date: 2026-10-12. Workspace: `.worktrees/c12-b01-task9-coordinator`.
Branch: `codex/c12-b01-task9-coordinator`; unchanged base/HEAD:
`a820f550a4e1662143676dc3f58d4f385c14f5a1`.

Authority: the user approved the [design](../superpowers/specs/2026-10-12-main-state-selected-leaf-binding-design.md)
and [five-task Native plan](../superpowers/plans/2026-10-12-main-state-selected-leaf-binding.md).
This plan explicitly excludes Git writes, push/merge, cleanup, installations,
whole-repository tests and production acceptance. Earlier approvals do not
override this checkpoint's separate Git gate.

## Delivery and current state

All five approved tasks are complete as a bounded local checkpoint. Task 5's
first complete matrix passed, but fresh review found one Minor deviation. It
was corrected with regression RED/GREEN and schema RED/GREEN; the complete
24-command repeat passed and same-reviewer reinspection closed M1 with no new
Critical, Important or Minor findings. This does not close overall acceptance.

The new pure `internal/nodecontrol/nodeleafbinding` package comprises exactly
16 product files and 17 tests. Nine authorized old consumer-test files receive
only the exact new-package exception and positive/negative controls (507 added
lines). No old product, fixture, SQL or module changes are authorized here.

The local profile `main-selected-leaf.der-spki-window.v1` binds only Proposed
Identity's node, selected certificate DER digest, SPKI DER digest and complete
certificate window. The layer owns the complete bounded input, calls real
`nodecerttime.Evaluate` once, immediately protects lower cleanup, preserves its
full Evidence and original errors, inherits its observation/expiry, retains
all lower obligations plus five explicit unresolved responsibilities, and
publishes only after the final context check. It adds no provider or runtime
authorization path.

## Task evidence

1. Contract, budgets and consumer boundaries: RED 001–010; GREEN 011–021.
2. Complete ownership, deep copies and shared clearing: RED 022; GREEN 023–024.
   Reflection checks cover nested mutable fields; 39 ownership cancellation
   points. Clearing assertions concern reachable owned backing only.
3. Frozen material and real lower: generator 025; RED 026; GREEN 027.
   Original 34 inputs and manifests are AST-extracted, byte-frozen and checked;
   only unsigned Previous/Proposed identity projections change. Independent
   checks retain raw bytes and parse actual DER/SPKI/windows. Ordinary tests
   decode constants, not old fixture files or freshly generated signatures.
4. Projection/publication: RED 028; 029 failed a reference mutation that zeroed
   an issuer digest and was correctly rejected by old shape validation. 030
   reproduced it; changing that control to a nonzero mutation yielded 031 and
   full 032 GREEN. The full context probe observes 655 cancellation/deadline/
   panic points with no returned model on failure and unchanged caller input.
5. Boundaries: RED 033–034; 035 found a stale package-name comparison in the
   new scanner; its scoped correction yielded 036 GREEN. 049 demonstrated an
   inadequate different-node same-key control; 050 uses the frozen same-node,
   same-SPKI, different-signed-DER leaf instead. 051 independently checks
   CA-SPKI and raw EC-point alternatives; 052 passes the refined full new suite.

First frozen final matrix: 052, 061–070, 074–086 select 24 exit-zero commands,
including ten reruns 077–086 to align all 42 Go hashes after control refinement.
It includes 20 full packages, specified authority tests, compile-only store
and two configured five-second nodeissuer fuzz invocations. Initial Envelope
count 36,917; Pair count 6. Actual whole-process times 65.199/10.244 seconds;
Go reports 6.881/8.037 seconds. Five seconds is not a whole-process deadline.
The 42-path formatting listing was empty. These are scoped results, not full
repository or physical acceptance.

Raw execution receipts preserve the actual executable, offline environment,
arguments, UTC timestamps, exit code, stdout/stderr and source hashes before
and after each command in ignored local scope
`.superpowers/sdd/2026-10-12-main-state-selected-leaf-binding/`.
Numbering gaps are not attempted commands. Failed receipts remain failures;
derived summaries do not overwrite raw evidence. Local ignored receipts are
not claimed portable or remotely saved.

## Independent review and correction

Fresh reviewer `/root/selected_leaf_binding_final_review` used the strongest
available review model, `gpt-6-astra`, at supported `xhigh` reasoning, read-only.
First result: 0 Critical / 0 Important / 1 Minor. Review did not rerun tests.
It independently checked 42 frozen source hashes and selected before/after
receipts, the 380-entry protected baseline, unchanged HEAD/index/design,
34 input pairs, 898 raw positions, 68 TBS hashes, 152 declaration-change paths,
and 19 successful projection manifests against certificate parsing.

M1: full normalized observation equality was required by design sections 6–7,
but lower.go used instant-only `time.Equal`. The actual normalized lower did
not demonstrate a reachable bypass; the defensive contract was nonetheless
too broad. 087 RED proves a same-instant/different-location value was accepted;
088 GREEN follows the full-value comparison fix. 089 RED proves the pinned
source guard rejects that real code change; only lower.go's AST hash is renewed
to `8c870d3524f71dfd4c44c6d8457022f74b9d5b6acdcc8ea297cf5d9cf21b5ae9`.
090 passes the lower regression and all four boundaries. Exactly three Go
files differ from the first review freeze: lower.go, lower_test.go and
boundary_test.go. No other boundary schema or consumer permission is widened.

The corrected final matrix 091–114 all exited zero. Each command's before/after
hashes match all 42 files in the correction freeze. A second 42-path formatting
receipt has an empty listing. Final fuzz Envelope/Pair executed 9/5 cases, with
Go durations 10.345/8.967 seconds and whole-process durations 39.59/30.91 seconds
(rounded). These small finite runs cannot establish broad coverage. All 104 raw
receipts remain separate: 81 exit-zero and 23 exit-one, including actual earlier
RED and failed attempts. Exit-one receipts are not acceptance evidence.

Same-reviewer final result: M1 ADDRESSED; new Critical 0 / Important 0 / Minor 0,
no scope drift. The reviewer independently reversed the three changes in memory
and matched the first freeze, checked all 42 current hashes, the protected
baseline, design, exact33-file inventory and each raw 091–114 receipt/output.
It also checked the original index hash without invoking Git; HEAD/branch rely
on author receipts and the first review record. It did not rerun tests or the
formatter. All 17 boundaries below remain unjudged, not waived. This final
documentation/checkbox closeout is the author's handoff, not a new independent
document endorsement; no Go source changes follow reinspection.

## Explicit rulings and their costs

R01: use PowerShell instead of skill-provided Bash bookkeeping scripts, per
the user's platform requirement and explicit plan workspace. Cost: adapted
bookkeeping needs auditing; plan-path, task briefs and receipts provide it.

R02: keep the approved limited matrix/no-Git/no-install/no-cleanup scope over
general skill defaults. Cost: wider integration and remote recovery remain
unverified, not silently promoted to acceptance.

R03: 047 correctly found the ignored one-off fixture generator's `.go` file
as an old product consumer. Preserve it byte-identically as `freeze.go.txt`,
record the archive hash, then retry the unchanged guard (048 PASS). Cost:
replay needs deliberate isolated restore and re-archive. No deletion or guard
relaxation is justified by its ignored status.

R04: approved store command had a path typo; 071 failed on the nonexistent
`./internal/nodecontrol/store`. Use actual `./internal/store` with the same
compile-only `-run '^$'`, timeout and offline environment (074 PASS). Cost:
literal command differs from the drafted plan; the plan preserves the original
and records the erratum. No store tests or product edits are inferred.

The reviewer declined the following domains. Author dispositions retain each
boundary; none is a waived acceptance gate:

| Domain | Disposition and cost |
| --- | --- |
| Issuer/serial authentication | Explicit unresolved obligation; no preimage/reference authentication claim. |
| Key handle/private-key possession | Retained obligation; no provider or possession proof. |
| Epoch/lineage/history/row status | Structural declarations only; material authenticity remains unresolved. |
| Main/observation/deployment CA provenance | Conditional inputs; platform provenance remains unproved. |
| Previous certificate | Only Proposed is bound; Previous material must be separately checked. |
| Full composition/commit/runtime/cached authorization | No consumer/API added; no durable or reusable authorization claim. |
| Provider/store fingerprint compatibility | Local profile only; production compatibility UNKNOWN. |
| End-to-end year-one observation | Unit window endpoint and CA-year-one controls only; frozen year-one leaf is expired at its real 2020 observation. |
| Concurrency/race | Unsupported; no race-safety claim. |
| Hidden runtime copies/forensic erasure | Reachable owned backing only; no hidden-copy or physical-erasure guarantee. |
| Real-time cancellation/OOM/crash recovery | Observed context points only; synchronous work and fatal failures are not guaranteed interruptible/recoverable. |
| Peak heap/performance | Raw budgets bound input, not total layered heap or latency. |
| Exhaustive crypto/fuzz/formal AST proof | Finite frozen controls and pinned syntax, not exhaustive mathematical proof. |
| Whole repository/race/Linux | Outside approved matrix; broader validation remains open. |
| DB/Docker/TLS/provider/hardware/reboot/PITR | Not exercised; installed Docker does not satisfy physical gates. |
| C12/B01/Task10/main/Git/production acceptance | Separate gates; no merge, push or overall completion claim. |
| Future edits | First freeze was followed by correction, full repeat matrix and same-reviewer reinspection; subsequent source/document changes require their own review and do not inherit endorsement. |

## Preservation and handoff

Approved design SHA256:
`3FCC3A2977B55C077660EA0F6D17AB2E2D42EA52A0ABAC45EC1BCC35EE621267`.
Pre-review Git index SHA256:
`391994201A90478BFDB9DED09DAE34810A537A2E9A7609A7BF7F0E900FC36BBE`.
Post-correction checks confirm all 380 protected baseline entries have only the
nine approved test differences, with unchanged design/HEAD/index. All 25 plan
steps are complete after same-reviewer reinspection and author documentation
closeout. No commit, push, main merge, cleanup or remote recovery checkpoint is
claimed. Any next Git preservation requires separate approval and must keep
historical acceptance blockers and these limited results explicit.

## 2026-10-13 Git preservation handoff

After the local delivery, the user requested the stated next step: commit and
push the necessary checkpoint files. This authorizes preservation on the
existing `codex/c12-b01-task9-coordinator` branch, not main merge, PR creation,
cleanup or broader acceptance. The original 46-file handoff hashes and empty
index were checked before any documentation update; origin is the approved
`github.com/win202569/Talenro.git`. Remote branch at the preflight read was
`bfb01bad53f465142cdaa4c9d1aabde4af206f2a`, verified as an ancestor of local HEAD.
Only ordinary fast-forward push is permitted; no force-push or remote overwrite.

Fresh pre-commit verification 115–138 reran the same24 approved commands, all
exit0. Every command's before/after42 source hashes matched the correction
freeze; a fresh42-path formatting check was empty. Envelope/Pair counts were
37,513/5; Go durations6.929/8.423s, whole recorded process durations30.432/10.808s.
Neither configured5s fuzz nor these finite counts prove exhaustive coverage.
All128 local raw receipts remain preserved (105 exit0,23 exit1); they are ignored
local evidence, not uploaded caches. The portable source/tests, design, plan,
this execution record and current-status comprise the46 necessary files.

The two documentation updates here are author Git handoff records, not changes
to reviewed Go sources or a new independent review endorsement. The finishing
skill's broader suite/menu/PR defaults remain bounded by the approved limited
matrix and this preservation-only next step. Cost: full repository/physical
acceptance still remains open; no PR, main integration or worktree deletion.

The prior section describes the original local delivery, before this separate
Git authorization. To restore progress, use the development branch above;
actual save completion is established by the resulting Git commit and matching
GitHub branch tip, not by interpreting this pre-commit record as a push receipt.
