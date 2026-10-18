# Bounded readonly review — handshake seam verification

Base: `c1b87ccc604b488dae3f48e8325b1238cd3e1df3`  
Head: `ca897038fc3b6fee6fa12cbacf034283f056b0aa`

Review scope: the ten checkpoint files identified in the review package and coordinator’s exact changed-file list, approved plan/specification, ledger rulings, new evidence packet, and necessary fixed codec/inbox/helper boundaries. Historical implementation and evidence were otherwise treated as hash-only dependencies.

## Strengths

- The sequence distinction is exercised correctly. `OrdinaryGap` checks complete typed snapshots and exact FIFO delivery of ordinary frames 1/4 around heartbeats 2/3. `StaleExpectedRejected` subsequently checks the actually dequeued frame against independent literal cursor 2. `DeclaredExpectedIsNotAuthority` uses literal cursor 4 without promoting checker success into production authority. See [handshake-seam.tests.ps1:57](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/handshake-seam.tests.ps1:57).
- Codec results receive complete ten-field root checks, including types, accepted/rejected status, null fields, next sequence, and all three false authority flags. Inbox result checks use the fixed twelve-field helper contract. See [handshake-seam.tests.ps1:23](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/handshake-seam.tests.ps1:23).
- ArmAck reference 3 and StopAck reference 2 are independently asserted against the fixture payloads, distinct from frame sequence 7. Positive and negative expected-reference cases exercise the fixed checker. Heartbeat references and the declared 113499/113500 expiry boundary are also explicit. See [handshake-seam.tests.ps1:100](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/handshake-seam.tests.ps1:100).
- Duplicate admission remains rejected after Take. The stopped-inbox case retains the original token, queued frame, sequence state, and first failure while both Feed and Take reject further activity. It does not create another token to receive an ACK. See [handshake-seam.tests.ps1:77](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/handshake-seam.tests.ps1:77) and [handshake-seam.tests.ps1:151](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/handshake-seam.tests.ps1:151).
- The controller consumes a slot before launch, enforces full/final ordering, rejects prior incomplete/nonexact outcomes, checks frozen dependencies, compares complete JSON structure and raw numeric tokens, and requires matching stdout/trace hashes across the two saved runs.
- Scope statements are appropriately conservative: these are test-only declaration/seam checks, not a real handshake implementation, state machine, ticket system, provenance proof, or capability validation. The approved first-green exception is disclosed rather than presented as product RED→GREEN evidence.

## Evidence inspected

Read-only inspection independently confirmed:

- All **1,378 frozen mappings** match their current files, with zero mismatches.
- Frozen manifest hash: `D7476546E4F2491E31D9A4B5A76B2A1458DCF3172CA07EF6D874E5E6970ACD86`.
- Invocation-index hash: `F631E13950558B50F2B3A7FA3404FC2F462132B2B743D34796826456163DB21E`.
- Both stdout files: `F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE`.
- Both trace files: `4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E`.
- Both stderr files have the empty-content SHA-256.
- Approved specification and plan snapshot hashes match the package.
- Portable `Invocations`, `FullTrace`, and `FinalTrace` match their raw JSON text after normalizing only CRLF/LF and outer whitespace. No numeric conversion was used for this comparison.
- Full trace and frozen oracle match across **271 ordered structural/value tokens**, including property order and raw numeric tokens. The identical final-trace hash establishes the same saved final trace.
- Saved receipts record PID26732/5331ms and PID29608/4044ms, each exit 0, with no timeout, completed owned exit/drains, intact post-run pins, and exact expected output.

These are validations of saved evidence, not new executions.

## Issues

### Critical — none identified

### Important — 1

**I1. A post-launch controller exception can abandon the owned target without enforcing its execution bound.**

- Location: [capture.ps1:102](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/capture.ps1:102), particularly the launch/cleanup block at lines 101–107.
- After `$p.Start()` succeeds, the controller persists the launched row before attaching stream-copy tasks or reaching `WaitForExit`. If that `Save-Json` throws—for example, because the invocation index becomes unwritable—the controller jumps directly to `finally`.
- The `finally` block disposes streams and the `Process` object but neither terminates a still-running child nor performs a bounded owned-exit observation. Disposing a `Process` wrapper is not child termination. Similar exceptional exits before completion can bypass the normal timeout path.
- Effect: the controller stops bookkeeping, but its already-launched target can remain running without the promised controller-enforced bound or confirmed owned exit. Slot preconsumption prevents an ordinary retry, but does not resolve this child-lifetime problem. This violates the reasonable expectation behind the explicit bounded-target and failure-stop controls.
- Severity rationale: this is an operational failure-containment defect in the new controller, not a production handshake defect. The two saved successful invocations do not exhibit it and their observed results remain valid; nevertheless, normal-path evidence does not establish the advertised controller guarantee under a post-launch I/O failure.
- Suggested handling: record this finding and stop completion under the approved Important-finding rule. Any separately authorized repair should ensure that every post-launch exit path retains ownership, conditionally terminates only that exact target when needed, performs bounded exit/drain handling, and preserves the original failure even if receipt persistence also fails. Do **not** modify the frozen controller, inject a failure, rerun, or request another reviewer under the exhausted allowance.

### Minor — none newly identified

The pre-existing inbox scalar `Count.Probe` assertion gap remains explicitly deferred. It is not a newly discovered issue or a repaired item in this increment.

## Recommendations

- Preserve both exact saved runs and this complete review without relabelling their observed outcomes.
- Mark the checkpoint **unfinished/review-blocked by I1**, with NEW2/2 targets and the sole reviewer seat consumed.
- Preserve the five frozen inputs unchanged. Do not use the remaining documentation/preservation work as authority for a controller fix, test, retry, or re-review.
- Keep the distinction between successful fixture observations and unimplemented real handshake/capability behavior prominent in the final status and recovery notes.

## Declined to judge

Each item below was considered and deliberately left outside this bounded review:

1. **Real owned-delivery provenance and a production handshake consumer:** the increment supplies fixture calls, not an implementation of `ConsumeOwnedControl` or an equivalent trusted consumption path. Lookalike-token rejection does not establish that future provenance property.
2. **Role-local handshake state, Hello/Bound/Ready ordering, duplicate stages, cross-direction ownership, and sender bookkeeping:** these remain expressly unimplemented. The ordinary-gap vector establishes inbox transport behavior, not legality of those messages in a live handshake phase.
3. **Actual resources, endpoint/process identity, P2-C readiness, owner/vendor clearance, and precise runtime approval:** none is established by these offline declarations or false authority flags; querying them was prohibited.
4. **Ticket preparation/consumption, double Commit, cancellation/stop races, native release attempts, and crash atomicity:** no such mechanism exists in this increment. No conclusion about those future behaviors follows from reference checks.
5. **Real clocks, StopAt/D enforcement, heartbeat loss/watchdog behavior, independent stopping, and hard-real-time guarantees:** the tested expiry values are declaration integers. Controller host limits are not proof of those runtime properties.
6. **Actual Stop/ACK transmission, Job/process termination, observed-empty evidence, cleanup, or successful sealing:** stopped-token behavior concerns the fixed inbox only; it proves none of those resource outcomes.
7. **Broader historical codec/inbox correctness, old suite results, and the deferred scalar ETS coverage gap:** only necessary dependency boundaries were inspected; historical artifacts otherwise received hash checks, not re-review or execution.
8. **Unapproved hosts, platforms, concurrent controller launches, hostile packet tampering, installation, and recovery replay:** the approved scope is a fixed ordinary PS7.6.5 serial workflow. These additional environments/adversarial models were not evaluated or authorized.
9. **Independent repository-history/remote-state verification:** the optional Git diff was unavailable in this review tool context. Direct-file review used the coordinator-supplied ten-file list and verified frozen identities. I did not independently establish remote push state or inspect unrelated branch changes.
10. **Dynamic reproduction of I1 or other controller fault injection:** prohibited by the exhausted execution budget. I1 is a source-established failure path, not an observed failure in either saved invocation.

## Assessment

**Bounded review completeness:** Complete for the supplied increment and permitted read-only evidence scope. No target/controller/test execution, module import, compilation, edits, index/HEAD changes, network/security queries, temporary worktrees, or nested agents were performed.

**Can this be preserved as a test-only development checkpoint?** Yes, **only as an unfinished/review-blocked checkpoint**, preserving I1, all limitations, and the exact saved observations.

**Ready to call this bounded task complete?** No. One Important controller finding requires the approved stop disposition; no automatic remediation allowance remains.

**Ready to merge to main, formally accept, or use live?** No. This review authorizes none of those actions.

The twelve groups and portable evidence support their narrow offline declaration/seam claims. They do not implement or verify a real handshake, state/capability system, and the new controller’s exceptional post-launch containment is incomplete.
