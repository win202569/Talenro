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

Independent final review pending at this checkpoint. Phase2 NOT AUTHORIZED.
Before phase2, expand native adapter contracts: ordinary token, connected instance,
retained handles, source/path/reparse/six hashes, receipt/transport cancellation,
actual elapsed checks at effect execution, authenticated handshake and original
window lifecycle. Pure effect enums do not perform or certify any OS operation.
Later isolated build, native verification and live approval remain separate gates.
Old product binaries, pins and UI acceptance remain historical and unchanged.
