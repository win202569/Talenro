# Cooperative guardian phase1 — pure implementation

Scope: user approved inline pure protocol/identity/time state machines, simulated
managed tests, final independent review and checkpoint push. No App restore or
build, native helper/guardian run, application launch, installs or security changes.
Base:009c7ea936574a4ad69a545d58af8fa2f2517e29.

## Evidence

Preserve the [original failed baseline](2026-11-16-guardian-phase1-baseline-blocked.md).
The explicitly approved one ordinary-token host rerun of the SAME command passed
409/409, exit0 (session41958); timeout was not increased. This resolves the baseline
gate, not the historical infrastructure cause.

`dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore`

Task1 activation/protocol RED20->GREEN20; exact identity RED2->full GREEN433.
Task2 state machines RED16->GREEN40; verdict RED4->full GREEN453; late target
assignment RED1->GREEN45 scoped tests. Task3 phone budget RED1->full GREEN464
(session50235, exit0, 464 passed/0 failed/0 skipped). Existing managed tests include
MSBuild property/validation evaluation, not App Build/Restore targets.

Shared code is linked explicitly only by Tests, outside App. No package changes.
Activation parser, bounded canonical HMAC frames, immutable transcript strings,
strict command line, native creation/CIM microsecond identity, serialized one-use
decisions, receipt/send/queue completions, deadline rechecks and separate verdicts
are implemented. In-memory tests cover a complete successful flow and failures.
No native pipe, handle, timer, dispatcher or transport assurance follows from this.

## Contract decisions

- PowerShell/apply_patch ledger equivalents replace Bash bookkeeping; no shell
  installation. Risk if wrong: less automation in ledger consistency.
- Digests/challenges are immutable lowercase hex strings, canonical raw32 on wire;
  bootstrap owns and clears key bytes. Future adapters must match encoding exactly.
- Real target is assigned only at the verified launch event; READY uses prelaunch
  elapsed time, launch switches once to T0-relative elapsed. No timer renewal.
  Adapters must capture actual creation, not predict PID or weaken identity checks.
- UI callback may run before enqueue acknowledgement, but only after persisted
  acceptance and one enqueue intent. Adapters must preserve original-object proof.
- Receive can precede sender-completion receipt; verdict uses causal constraints,
  not artificial cross-process log ordering. Correlation/finalization/hashes remain
  explicit trusted adapter evidence, not inferred from an ACK.

## Recovery and next gate

Read the [approved plan](../superpowers/plans/2026-11-16-windows-cooperative-guardian-implementation.md)
and [design](../superpowers/specs/2026-11-16-windows-cooperative-close-guardian-design.md).
Tests use committed deterministic fixture material, not real secrets or old receipts.
Restore dependencies only under separate preparation on a fresh computer; the
verified command here deliberately uses existing dependencies with --no-restore.

Independent final review completed on009c7ea9..8e45f02f. Three Important findings
were reproduced separately RED1 each and fixed in one pass; no Critical findings.
Final full managed suite467/467, exit0, session92504; git diff --check clean.
No second review was commissioned; fixes are verified by regression tests.
Phase1 pure-layer gate complete; phase2 NOT AUTHORIZED.
Before phase2, expand native adapter contracts: ordinary token, connected instance,
retained handles, source/path/reparse/six hashes, receipt/transport cancellation,
actual elapsed checks at effect execution, authenticated handshake and original
window lifecycle. Pure effect enums do not perform or certify any OS operation.
Later isolated build, native verification and live approval remain separate gates.
Old product binaries, pins and UI acceptance remain historical and unchanged.

## Independent review disposition

1. Prelaunch wire contract: fixed explicit stage-specific absence. READY carries
   canonical App(0,0,empty,empty), T0=0 and zero challenges. Assignment carries the
   real target/T0. HELLO carries app challenge only; CHALLENGE/PROOF/BOUND/close
   require both. Control frames reserve zero challenges; generated application
   challenges must be nonzero. Field order/MAC/size stay unchanged. Invalid stage
   combinations reject. Wire lifecycle regression now exercises all stages.
2. Clock handoff: Launch accepts actual elapsed since immutable T0, not synthetic
   zero, and rejects assignment past10000ms. Regression covers500ms, late binding,
   duplicate assignment and backward clocks. No new READY-to-launch timeout added.
3. Phone cancellation: explicit irreversible Cancel under lock; binding loss also
   cancels. Synthetic close claim preempts outstanding confirmation; duplicate/
   late replies cannot restore authority. Native coordinator must call cancellation
   on close claim, cancellation, binding loss and session termination.

Deferred minor: observer-loss integration fixture still uses Deadline rather than
propagating ParentLost into its encoded request, and permutations run at290s rather
than covering genuine pre290 early/loss reason classification. Production reason
selection is present; extend this fixture before phase2 integration. Not a claim of
autonomous/timer coverage for an early/manual exit.

### Review boundaries retained (executor rulings)

All nine declined-review areas remain explicit, not silently accepted:

1. Native connected-instance/retained-handle identity, PID reuse, pipe replacement,
   token/ACL/remote checks: phase2/native gates. Risk: mock identity is not OS proof.
2. Randomness, key delivery/erasure/inheritance, actual listener dormancy: future
   native checks. Risk: parser success alone cannot establish secret isolation.
3. Physical I/O/cancellation/receipt bounds and effect-time deadline checks: future
   adapters. Risk: a timely decision can still execute late; late effects must fail.
4. Original Window, dispatcher, modality, normal-close cancellation and exit:
   future UI/native evidence. Risk: simulated generation is not retained-object proof.
5. QPC/suspension/scheduling/independent timer: future native gate. Risk: injected
   clock tests do not guarantee OS scheduling. Pure handoff defect was fixed now.
6. Truth of verification/correlation/finalization/hash/exit/classification flags:
   trusted producer contract, tested structurally, not externally proven here.
   Risk: fabricated evidence can misclassify; adapters must obtain real evidence.
7. Separate READY-to-launch timeout: not invented; READY response<=10s and actual
   T0-to-BOUND<=10s retained. Risk: prelaunch wait is not bounded by this model.
8. New artifacts/dormant App/UI regression/historical acceptance transfer: no new
   artifacts built; all require later approval. Risk: old passes cannot accept new code.
9. Legacy defects and baseline infrastructure root cause: outside phase1 diff;
   successful host rerun resolves only test gate. Risk: old infrastructure can recur.

Review range limited to this phase1 change set; existing branch and worktree are
retained. User explicitly chose necessary-file commit/push, not merge/PR/deletion.
Plan scratch ledger is migrated to this portable record; no raw receipts removed.
