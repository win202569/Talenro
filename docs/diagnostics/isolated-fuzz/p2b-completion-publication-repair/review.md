## Strengths

Reviewed BASE `8b842a19a4d43c0b079c69d9d9e917c9d2cc1c2b` through HEAD `349947e4a57f23ca4f05cee673b46ef94345f265`: all eight frozen inputs, all 45 packet files, progress rulings, approved-design snapshot/execution brief, prior B1/B2 review and disposition, and relevant linked contracts.

This was read-only file/data and Git inspection. No repository code was executed or imported; no tests, probes, network requests, writes, trust-setting changes, or additional agents were used. The eight working-tree inputs match HEAD. Later author-generated handoff outputs are excluded.

Evidence identities are consistent:

- All **1,599 GREEN mappings match** their current files; no conflicting duplicate hashes.
- All **1,539 inherited mappings remain unchanged**.
- RED→GREEN changes exactly **one mapping: `publication.ps1`**; the other 1,598 match.
- Both publication snapshots match their corresponding manifest pins.
- Manifest hashes independently match:
  - RED: `CCECB43871CAB41AE40570F648AF47EDE33C01129DDAE745FBF8ACA2130F42EC`
  - GREEN: `3A763103C93625C522CF8BFF997F9280EA07F3278A6D3FCE6D221759368D4237`
- All five complete trace/capture comparisons match their literal oracles, preserving property order, array order and numeric tokens.
- Every host-recorded stream, trace and details hash matches its saved file.

| Saved run | Host PID | Elapsed | Exit | Exact output |
| --- | ---: | ---: | ---: | --- |
| RED | 34916 | 805 ms | 1 | 3 expected FAIL lines |
| GREEN | 21368 | 783 ms | 0 | 6 PASS groups |
| CONTAINMENT | 19960 | 1,655 ms | 0 | 7 PASS groups |
| FINAL | 36748 | 7,350 ms | 0 | Original 15 lines / 12 PASS |

The four child records identify PIDs 38744, 31824, 16884 and 2644, with recorded owned exits, expected exit codes and no rescue. Applicable drains, receipts and post-pin observations agree. These are verified saved observations, not retrospective live-process observations.

FINAL retains the original hashes:

- Stdout: `F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE`
- Seam trace: `4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E`

### B1/B2 and the five review-focus items

1. **B1’s specific integration defect is repaired.** The real finalizer still calls shared completion at [capture.ps1:157](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:157). Its subsequent publication now passes through the failure latch at [publication.ps1:3](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/publication.ps1:3), with immediate refusal handling at [capture.ps1:174](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:174). A previously failed completion cannot reach that normal later save. The healthy path rereads the existing ledger rather than rebuilding it from `$Rows`.

2. **B2’s specific output-order defect is repaired.** [publication.ps1:10](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/publication.ps1:10) renders, writes both output portions and flushes before attempting authoritative persistence. The actual caller at [capture.ps1:291](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:291) checks the result, then returns without further normal output. The saved disposed-writer case leaves the original `INCOMPLETE` ledger unchanged and next-slot eligibility false. This does not establish failure-closed behavior for the subsequent ledger write; see C1.

3. **First-cause retention and cleanup ordering remain sound for ordinary private inputs.** Actual errors are latched before final cleanup; cleanup precedes final bookkeeping. Duplicate error references are suppressed, secondaries are bounded at eight, overflow remains explicit, and preconsumption failure gates precede starts. However, the parent has the same-sink retry described in C2.

4. **The new tests exercise the real shared helpers.** Missing-file, malformed-JSON and disposed-`StreamWriter` failures are real operations, not fabricated error records or source-text assertions. Healthy output uses an independent literal expectation. These are shared-consumer tests plus static inspection of actual wiring—not dynamic injection into the actual integration finalizer or supervising ordinary parent.

5. **The helpers do not recursively log failures.** Fallback remains bounded and best-effort. Authority flags remain false. Nevertheless, C1 and C2 prevent a broader claim that all detected publication failures remain failure-closed and never retry their failed sink.

## Issues

### Critical — Must Fix

None found.

### Important — Should Fix

#### C1 — A detected final-ledger write failure can still leave a continuation marker that a later invocation trusts

Locations: [publication.ps1:15](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/publication.ps1:15), [bookkeeping.ps1:17](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/bookkeeping.ps1:17), and [capture.ps1:184](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:184).

The final publication serializes rows already carrying `EXPECTED_*` directly over `hosts.json`. If that write reports an I/O failure after affecting the destination, the catch marks only the live failure state incomplete. It does not establish that the destination still contains the preceding incomplete row.

In particular, a late/ambiguous write failure can leave a complete, parseable successful document even though the operation returned failure. The current invocation then exits 2, but a later invocation constructs a fresh failure state and authorizes continuation from the persisted `Outcome` fields. It has no independent evidence that the preceding final write returned success.

This is narrower than demanding crash-proof or power-loss durability. It concerns an **ordinary error detected by the live controller** whose persistence outcome is uncertain. The documented requirement that ambiguous I/O remain unknown/incomplete is not enforced by a gate that treats whatever valid success marker remains on disk as sufficient. A manual-stop instruction does not change that gate.

This failure was **not dynamically induced** in the saved runs. The finding is the statically visible absence of a successful-completion acknowledgement distinct from the potentially affected file.

**Remedy:** Under separately approved work, make continuation depend on positively acknowledged completion—not merely a parseable `EXPECTED_*` row following an ambiguous write. Unknown acknowledgements must refuse continuation without retrying or reconstructing the failed sink. This need not promise recovery when every sink fails.

#### C2 — The parent retries its failed host-ledger sink during finalization

Locations: [capture.ps1:211](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:211), [capture.ps1:228](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/capture.ps1:228), and [bookkeeping.ps1:30](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair/bookkeeping.ps1:30).

The post-start `Save-Json $index $hosts` can fail and latch the first error. The parent correctly notices that state, skips normal waiting and performs cleanup. However, it then unconditionally invokes `Complete-P2BLedger` against the same `hosts.json`.

`Complete-P2BLedger` does not reject an already failed sink or failure state: after reading a valid existing row, it calls `Save-P2BLedger` again.

For an ordinary write refusal that leaves the preconsumption document readable—for example, a sharing/access failure—the finalizer therefore attempts another write to that same failed sink. If the transient restriction clears, it overwrites the prior document; if it persists, another write error is generated. The new publication helper’s latch is not involved in this path.

Cleanup still precedes the retry, the first cause remains retained, and the later failure gate prevents normal success. This is **not** a demonstrated process leak or renewed exception masking. It is a caller-level violation of the no-retry/preserve-failed-sink boundary, analogous to the mechanism that made B1 important.

**Remedy:** Track failed-ledger publication separately from necessary resource cleanup. After a host-ledger write fails, finish cleanup but do not attempt final read/replace/write publication to that sink; retain the first cause and use only the approved bounded fallback.

### Minor — Nice to Have

No new minor finding.

The existing **R2 `PreStart`/`prestart` diagnostic-key collision** remains present and explicitly deferred. Its missing individual pre-start exception description was not reconstructed. **Minor2** remains deferred and unchanged.

## Recommendations

- Record **B1 and B2’s original mechanisms as narrowly repaired**, while leaving complete failure-closed publication acceptance unfinished because of C1 and C2.
- Preserve all successful observations, original RED evidence, frozen inputs, snapshots and historical findings unchanged.
- Disposition C1 and C2 explicitly. No automatic fix, target, retry or second review follows from this report.
- Preserve the exhausted allowance: **4/4 hosts, 4/4 children and 1/1 fresh review**. Historical frozen allowances remain frozen.
- Keep `SourceVerified`, `LiveActionsPerformed`, `RuntimeApproved`, `FormalAcceptance` and `RealGateImplemented` false.
- Label later portable exports, reports, dispositions, status and recovery changes **author-checked, not independently reviewed by this reviewer**.

## Declined to judge

The following is the complete list of considered boundaries or behaviors not positively accepted by this review. None waives C1, C2 or an approved requirement.

1. **Production handshake implementation and provenance:** this is an offline controller repair, not a trusted production consumer.
2. **Hello/Bound/Ready legality and production state ownership:** the retained seam does not implement those stages.
3. **Real capabilities, P2-C readiness, owner/vendor clearance and runtime permission:** fixtures and source review establish none of these.
4. **One-use release tickets, double Arm/Commit and cross-process atomicity:** not implemented or validated by this repair.
5. **Actual Stop/ACK semantics, Job emptiness, resource-release proof and cleanup sealing:** fixture exits are not substitutes for those guarantees.
6. **Real watchdogs, StopAt/D enforcement and hard-real-time bounds:** bounded waits and recorded elapsed times do not prove those properties.
7. **Crash recovery, power loss and durable transactional storage:** outside the approved execution. This exclusion does **not** waive C1’s handling of locally detected ordinary I/O errors.
8. **Guaranteed persistence/reporting when every sink fails:** fallback is explicitly best-effort; unavailable evidence must remain unknown, not recovered success.
9. **Dynamic failure injection into actual integration/child/parent finalizers:** shared helpers and actual call sites were inspected, but those lifecycle injections were not performed.
10. **Dynamic output failure in the actual supervising ordinary parent:** the disposed writer was exercised in the publication unit host. Actual-parent wiring was reviewed statically.
11. **Separate dynamic failures during raw-output writing, flushing, final-ledger writing or fallback output:** the saved closed-writer case fails at the first write. Other boundaries were inspected statically; C1 is not waived by their lack of dynamic coverage.
12. **Dynamic preparation/preconsumption persistence failures:** snapshot/manifest and start-gate paths were inspected, not newly faulted. Successful saved preparation does not prove every failure path.
13. **All serialization, conversion, index and path-error variants:** exception boundaries were inspected; the listed real vectors do not cover every possible input or failure.
14. **Unknown Process/API/OS failures and unusual exception chains:** relevant guards were read, but no additional OS failures were induced.
15. **Other PowerShell/.NET versions, architectures, operating systems or machines:** observations remain limited to the pinned Windows/PowerShell environment.
16. **Concurrent controllers, adversarial modification and replay resistance:** the approved workflow is serial and private. This does not excuse ordinary detected errors or the current continuation gate in C1.
17. **Junctions, symlinks and adversarial path aliases:** lexical packet-path checks are not a separately accepted filesystem security boundary.
18. **Arbitrary public callers, mutated state, hostile getters/setters and custom exception implementations:** these functions consume ordinary controller-owned private data.
19. **Fatal termination or memory exhaustion during error retention:** no survival or successful reporting guarantee was established.
20. **Disposed, reused or unowned Process objects and unrelated process trees:** the contract requires newly created, retained exact objects. No process enumeration or live-state query was performed.
21. **A fresh correctness review or rerun of every historical suite and the entire frozen lifecycle helper:** relevant boundaries and identities were inspected; historical acceptance was not reopened. The helper’s existing receipt-failure vectors are distinct from C2’s host-ledger retry.
22. **Deferred R2:** the individual pre-start diagnostic loss remains acknowledged and unrepaired.
23. **Deferred Minor2/scalar ETS coverage:** unchanged and outside this approval.
24. **Reconstruction of CLR identity from JSON:** recorded relationship assertions are evidence; serialized files cannot restore the original `ErrorRecord` references.
25. **Independent historical execution chronology or current process liveness:** the review inspected saved evidence and source, not the original live executions.
26. **Singleton-row JSON represented as an object rather than an array:** private readers deliberately wrap parsed results in `@(...)`; no incompatible fixed-array consumer was found.
27. **Individual PASS lines, rendered `EXPECTED_*` output or subordinate `FINAL_EXACT` as full completion:** these can be observations before remaining recording finishes. The authoritative host continuation gate is a separate concern addressed in C1.
28. **Healthy reread preserving arbitrary additional rows under concurrent edits:** source performs an existing-ledger reread; the saved healthy case covers one ordinary row, not concurrent multirow mutation.
29. **`FinalLedgerWritten` as an independent acceptance capability:** it is a local bookkeeping observation, not the continuation authority. No broader guarantee is inferred from that field.
30. **Scope labels as separate algorithms or lifecycle injections:** the implementation is intentionally shared; labels do not multiply dynamic coverage.
31. **Recovery that reconstructs missing historical evidence or automatically reruns exhausted slots:** neither is authorized; clone recovery does not restore raw packets, dependencies, identities or runtime rights.
32. **Remote checkpoint equality:** network access was prohibited. Local HEAD and frozen source identity were verified; remote equality remains author-checked.
33. **Later portable summary, report/addendum, dispositions, status and recovery handoff:** outside the independent frozen review set.
34. **Formal, main, live or full-C12 acceptance:** no such authority is supplied by this review or the limited successful observations.

## Assessment

**B1:** Resolved narrowly. Failed integration completion now prevents the subsequent normal ledger publication; the missing/malformed evidence and retained first cause are preserved.

**B2:** Resolved narrowly. Actual normal parent output and flush precede authoritative success persistence, and no normal output follows that persistence.

**Ready to accept the complete failure-closed publication/controller contract? No.** C1 leaves uncertain final-write results consumable as success, and C2 retries a failed parent ledger during finalization. The saved runs remain valid evidence for their exact observed paths; they do not close these remaining error paths.

The sole review is complete. No further implementation, execution or review is automatically authorized.
