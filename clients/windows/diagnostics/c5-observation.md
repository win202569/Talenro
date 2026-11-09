# Fixed C5 observer

`../tests/observe-c5.ps1` is PowerShell7, no-start by default. From the repository:

```powershell
pwsh -NoProfile -File ./clients/windows/tests/observe-c5.ps1
```

This returns one `NotStarted` object, before environment queries, imports or file
creation. Unknown executable/profile/deadline/backend overrides are rejected.
`-RunApproved` selects live preparation but **does not grant approval**. Do not use
it until a separate bounded one-launch session is approved. Current live budget0.

## Future authorized session only

Operator first confirms available interactive desktop, no existing target windows,
ordinary token and a clean `codex/c12-b01-task9-coordinator` checkout. Script checks
branch/clean status, runtime observer source blobs, preserved EXE/DLL hashes and
process inventory. Any unavailable evidence blocks, without rebuilding or retry.
The broad prototype inventory can conservatively block unrelated prototypes.

The fixed old artifact and hashes are in `observe-c5.ps1` and the approved design.
Arguments are exactly `--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage`.
There is no automatic close, search, UI action, PID reopening or forced termination.
Keep the operator route: startup within30s, normal close requested by300s, exit
criterion30s. The observer cap is360000ms from immediately before the single launch
attempt, max sleep2000ms clipped to remaining time. OS calls, writes and disposal
are synchronous: this is not a hard real-time deadline. Do not start replacements.

Original Process and borrowed SafeHandle remain retained inside the resource scope.
Native creation must match exactly; CIM creation uses microsecond truncation only.
Expected path/full arguments are independently fixed, not copied from CIM evidence.
Native evidence precedes CIM, then each native/managed pair. On abort, stop functional
actions and arrange normal close; disposal of observation resources never kills.

## Receipts and failure

A fresh lowercase GUID identifies a CreateNew `samples.jsonl` under the ignored
artifact `process-observations/<runId>/` directory. Existing destinations and reparse
ancestors are refused. This is a trusted-workspace precaution, not protection
against an adversary racing filesystem operations. No raw receipt enters Git.

Schema1 records carry mode/runId/sequence/elapsed. Kinds: start, native-identity,
pair, outcome. Initial native identity is explicitly unverified; pair
`IdentityVerified` means the initial session guard passed, not that every raw field
in that pair has passed subsequent checks. Raw values remain available on conflict.
Native signal/code0 alone is not verified exit: valid time and managed agreement
are required. Outcomes: identity-failed, observation-failed, timed-out, exited.

Persisted outcome has `Finalized=false`: it cannot prove disposal succeeded.
The returned/information-stream completion has `Finalized=true` only after cleanup,
with elapsed/overrun measured then. Cleanup crossing the cap revokes timely success.
No finalized completion after a write/cleanup/publication failure; best-effort abort
has ExitVerified=false. Final progress transport itself can block; no hard deadline
is asserted. Preserve terminal output alongside raw receipts for a future live run.
Failures expose fixed reasons, not arbitrary exception text. Exit1/nonterminating
session failure never means acceptance; default NotStarted exits normally.
Failure receipts refresh elapsed time best-effort; post-cleanup abort includes actual
elapsed/overrun when its clock is valid. TimingVerified=false means the displayed
elapsed is only the last known valid reading. Managed acquisition failure preserves
the preceding native observation as an incomplete pair, never as verified exit.

## Synthetic verification and recovery

Run the five PowerShell test scripts: `observe-c5.tests.ps1`,
`c5-observation-session.tests.ps1`, `observation-session-guard.tests.ps1`,
`process-observation.tests.ps1`, `native-process-observation.tests.ps1`.
Entry tests AST-extract private definitions and execute command doubles: no public
callback override, native compilation, live CIM, real launch or sleeps. They test
policy/wiring, not OS behavior. Orchestration uses original-object memory fixtures.

Cloning restores source/tests/docs, not ignored preserved artifacts or raw evidence.
Missing binaries/hashes are a blocker: do not rebuild or substitute under this plan.
Obtain a separately approved artifact preparation path before any live attempt.
C5 remains NOT_ACCEPTED; RTL, real shutdown behavior and root cause remain unresolved.
