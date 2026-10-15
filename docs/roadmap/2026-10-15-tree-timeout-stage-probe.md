# One tree-timeout stage probe — 2026-10-15

Subsequent approved minimal-entry probe observed flushed stdout plus file marker
and completed naturally within2s. Its payload explicitly launches no child and is
not this longer-script fixture; it does not explain or retroactively pass this
timeout. See [new single-entry evidence and limits](2026-10-15-entry-startup-probe.md).

## Approved scope

The user approved exactly one diagnostic with parent-entry, before-child-start,
after-Start-Process-return and child-entry markers. The original two-second
whole-process budget, native runner, safety/cleanup rules and original test remain
unchanged. No Go, fuzz, product build, production policy change or retry was allowed.

New `tree-timeout-probe.ps1` provides readonly planning, evidence collection and an
explicit one-shot entry point. It requires a new owned `tree-timeout-probe-NNN`
identity, rejects existing/original-sampling identities before launch, and uses the
same existing Windows PowerShell5.1 executable/hidden Start-Process structure.
Stage writes add observation overhead within, not outside, the existing budget.
The after-return marker records the cmdlet's actual success flag immediately;
missing markers do not imply successful or failed startup.

The probe always labels its summary/result Acceptance=false, OriginalTestPassed=false,
DiagnosticOnly=true. It requires verified empty-job cleanup before returning. A
cleanup failure saves stopped.json and throws, without retry. A diagnostic command
returning0 means only its capture/cleanup workflow returned, not that the original
test or a fuzz/sample passed. Raw process receipts retain their native timeout facts.

## Verification and single execution

The initial safe stub failed the two-second-budget assertion (RED), before any
observed process launch. After implementation, nine non-launching checks passed
under PowerShell7 and Windows PowerShell5.1: budget, readonly plan, outside/wrong/
duplicate identity rejection, generated-parent parsing, missing markers and
partial-parent markers without invented child success. The tests never call a
valid execution path; they only invoke execution against an existing rejected root.

Actual probe001 ran ONCE at UTC09:46:36.481–09:46:38.538
(Asia/Tbilisi13:46:36–13:46:38). Its receipt records:

- TimedOut=true, TotalProcesses=1, RootExited=false, termination requested.
- CleanupVerified=true, final active accounting0 and member list empty.
- All four stage files absent; stdout/stderr both zero bytes.
- The recorded root was owned/running at shutdown; both acquisition APIs succeeded
  with error0. The historical origin87 was not reproduced.
- Result Acceptance=false and OriginalTestPassed=false. No second probe was run.

Thus this sample did not observe even the first parent-script marker within the
budget, let alone establish the intended descendant scenario. It narrows the next
investigation to reaching/completing that first observable statement. It does NOT
prove PowerShell startup latency, marker-write failure, scheduling or a native
cleanup defect. Empty stderr under forced termination does not exclude an
unflushed/delayed error. Extra instrumentation/command size may affect timing;
this is a diagnostic observation, not reproduction equivalence or a fixture fix.

Readonly readback rechecked absent markers, false acceptance and verified empty-job
cleanup, without another observed process. Original Envelope callback SHA256 still
equals318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E.

Ignored worktree-relative evidence:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/tree-timeout-probe-001/`.
SHA256: process/receipt.json
`D1F286EAF51DA69F1BC7E78B7E2B01A12A8A5492B9E7003CE441B37F2B2AC6ED`;
result.json `8151351359D8395963529C7D0439FAA20E0FE1B69D420985F64A95B8F8DA956D`.
Do not query recorded PIDs or reuse this consumed probe identity.

## Independent review and remaining limits

Independent read-only review found no blocking issue in the two new script/test
files and supported an unaccepted development checkpoint, not main/product
acceptance. It did not run processes/tests. Nonblocking gaps remain explicit:
unit parsing covers the parent, not a decoded-child/argument roundtrip; positive
after-return/child markers and the cleanup-refusal branch are not exercised.
stopped.json contains Acceptance=false but does not duplicate the success
envelope's other diagnostic labels; absence never asserts acceptance.

No full project/native timing suite was rerun in this one-probe scope. The prior
run.tests.ps1 descendant failure, its16 unrun checks, original origin87 and product
gates remain open. Nine passing helper checks and one safely cleaned timeout cannot
erase them. No main merge or acceptance upgrade.

## Next step and recovery

Further work needs a distinct approved diagnostic or fixture design that separates
reaching the first parent statement from descendant creation without weakening the
original timeout/cleanup tests. Do not simply enlarge2s, repeat until markers
appear, retry preparation001 or infer acceptance from an empty group after cleanup.

Save necessary code/tests/reports on the latest
codex/c12-b01-task9-coordinator; remote SHA equality is checked at handoff. Recover
that branch and read current status/repository recovery. The scripts use existing
PowerShell and generate new local fixtures, not old caches/toolchains. Cloning
does not restore raw/live state or authorize replaying this completed diagnostic.
