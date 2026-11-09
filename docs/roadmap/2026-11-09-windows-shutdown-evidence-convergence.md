# Windows shutdown: consolidated evidence and investigation stop point

Status: historical shutdown stall remains UNRESOLVED; recent runs did not
reproduce it. No repair or broader acceptance is claimed. This document consolidates
existing receipts only: no new launch, build, test, service change or runtime budget.
It supersedes earlier proposed next actions, not the historical observations.

## Established observations

| Evidence | Outcome | Limit |
| --- | --- | --- |
| Historical automatic input, old artifact | Two observations reached the 360s cap without native process-exit signal; wait chains pointed to the same svchost PID/thread | Window disappearance and managed HasExited/code0 did not establish exit; no blocking API identified |
| Historical physical manual input, old artifact | Two scoped exits succeeded | Fixed order, timing, focus/IME and service state were uncontrolled |
| Three recent instrumented routes | baseline/titlebar close, long-country/titlebar close, long-country/Alt+F4 all exited; each had 24 healthy complete stage markers | Success is per-run; complete handler stages alone do not prove process exit |
| Preserved old artifact, current procedure | Unchanged EXE/DLL, automatic jp-1, long-country/Alt+F4: 42 identity-valid pairs, native/managed exit0, no leftovers | No stage instrumentation; internal handler boundary unknown |
| Deployed-file comparison | 527 files per tree; 522 identical; only App/Core binaries and symbols differ | Current disk comparison, not historical loaded-module evidence |

The last old-artifact run (86b9cadd4d22486d877313f16aa30ad7) establishes that
the recorder and new platform build are NOT necessary for that successful exit.
It does not erase historical failures or prove that either the environment or
observer caused them. Few, nonrandomized runs provide no reliability estimate.

Sources, identities, timestamps, raw-record hashes and limitations remain in:

- [Historical wait chain](2026-11-08-windows-wait-chain-observation-execution.md).
- [Adjacent automatic/manual comparison](2026-11-09-windows-adjacent-input-comparison.md).
- [Earlier manual observation](2026-11-09-windows-manual-input-observation-execution.md).
- [First recorded route](2026-11-09-windows-close-stage-first-observation.md).
- [Long-country route](2026-11-09-windows-close-stage-long-country-observation.md).
- [Alt+F4 route](2026-11-09-windows-close-stage-altf4-observation.md).
- [Artifact comparison](2026-11-09-windows-close-stage-artifact-comparison.md).
- [Preserved old-artifact run](2026-11-09-windows-preserved-artifact-observation.md).

## Hypotheses, not conclusions

| Hypothesis | Current evidence | Useful discriminator still missing |
| --- | --- | --- |
| Old artifact always stalls; recorder fixed it | Contradicted by unchanged old-artifact success | A reproducible failure tied to a specific code boundary; intermittent app defects remain possible |
| Automatic input/service interaction | Historical association and wait-chain edge remain relevant | Contemporaneous exact identities and evidence of the actual blocking operation; tool transport is unknown |
| Observer procedure affects lifetime | Historical roughly5s sampling/outer waits differs from current roughly2s sampling | Approved controlled comparison changing only observer procedure, with a discriminating outcome defined in advance |
| Focus, IME, timing or service state | Plausible uncontrolled differences, not measured causes | Relevant state captured contemporaneously without collecting private content |
| Dependency upgrade explains outcome | No changed dependency files within the compared deployment sets | Evidence of changed external loaded components; file comparison cannot supply that history |

Later service enumeration associated the wait-chain host with cbdhsvc_8b240.
That was not an exact historical membership snapshot; thread lifetime was not
verified. Do not call this a clipboard-service defect, proven deadlock, or proof
that automatic input used the clipboard. Do not inspect clipboard contents.
Core method IL/signature equality is not full semantic equivalence; platform
metadata differs. Neither platform nor instrumentation causation is established.

## Conditions for any future investigation

These are decision criteria, NOT authorization to launch or collect additional data.
All existing launch budgets are exhausted.

1. State one specific hypothesis, one changed variable, the outcomes that would
   distinguish it, a bounded launch count and explicit stopping rules before work.
   If another success would not change the conclusion, do not repeat that run.
2. Preserve exact artifact hashes, path/arguments, ordinary-token preflight,
   run identity and original retained process handle. Correlate native creation
   time, PID and managed start time; retain raw records without overwriting them.
3. Record actual input value/focus, input method, route, close gesture/timestamps
   and the complete observer procedure. If relevant state/version metadata is
   needed, define its narrow scope first; no secrets or unrelated user data.
4. Determine exit from the original native handle and valid exit evidence, not
   window disappearance, exit code alone or managed HasExited alone. Preserve
   disagreements. A fresh wait-chain observation must identify the current target;
   never reuse historical PIDs as live identities.
5. If a future instrumented failure has a healthy complete close sequence but no
   native exit, investigate beyond that recorded handler boundary. A valid partial
   prefix narrows the last observed boundary only: missing tail is inconclusive,
   not proof that the next operation blocked. Missing/unhealthy records cannot
   establish handler completion. Old artifacts have no such stage evidence.
6. Keep any approved absolute deadline. Stop on identity/safety failure, unexpected
   extra target, budget exhaustion or completion. No automatic retry, forced exit,
   dump/debugger attachment, service restart, settings change or broadened acceptance.
   Those actions are not authorized by this report.

## Disposition and recovery

Stop undirected repetitions and speculative shutdown changes. Preserve the
unresolved blocker and all existing NOT_ACCEPTED gates. Do not merge main or
declare the application repaired. A new bounded discriminator requires separate
scope; an external support submission also requires separate authorization.

This checkpoint contains the portable evidence ledger and links, not the ignored
raw logs, screenshots, binaries or live machine state. Recover source/tests/docs
from the latest codex/c12-b01-task9-coordinator commit using
[repository recovery](../runbooks/repository-recovery.md). A rebuild is not the
preserved historical artifact unless its exact identity is independently verified.
Cloning does not restore runtime state or grant permission to resume observations.
