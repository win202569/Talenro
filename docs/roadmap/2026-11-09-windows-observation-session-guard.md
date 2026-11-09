# Observation guard repair preparation

Baseline c15753a1. User requested the next repair step; scope is diagnostic helper
code plus synthetic tests, not product behavior or another application launch.
The previous inline command is not a tracked runner: added a small reusable guard
and resource scope rather than silently rewriting the historical command/receipt.
Implementation instructions: clients/windows/diagnostics/observation-session-guard.md.

## Verification

- Initial test-first run failed with the explicit assertion that the guard module
  was missing; this is an implementation-absence RED, not a live reproduction.
- After implementation: new suite41 assertions passed, exit0. It reproduces the
  historical unequal formatted timestamps, accepts only same-microsecond CIM
  truncation with exact native identity, rejects adjacent microseconds/native
  one-tick mismatch/wrong identity/unknown evidence, and checks cleanup on success,
  identity failure, sample failure and acquisition failure. Both primary and
  disposal errors survive simultaneous failure.
- Existing process-observation suite32 cases and native-observation synthetic/static
  suite78 assertions passed in separate PowerShell hosts, exit0. Native probe was
  not loaded/executed. No .NET suite/build or WinUI launch was performed.

No millisecond tolerance, process-name exemption, kill, PID reopening or historical
receipt reinterpretation. Exact command-line comparison also avoids accepting an
unknown-usage-extra suffix through the earlier substring check. These are helper
guarantees over supplied evidence, not proof of its live provenance.

Resource factories must transfer exactly one owned resource; partial factory failure
remains the factory's cleanup responsibility. The scope cannot bound a blocking
Dispose. It does not replace the caller's deadline or detect live process completion.

## Handoff

The old C5 attempt remains NOT_ACCEPTED with its empty receipt untouched. The new
helpers have NOT been exercised against an application process. Future preparation
must explicitly import/use the guard, obtain successful native creation evidence
on the original handle before validating CIM, and put all post-launch work inside
Observe. No remaining launch allowance or automatic retry; shutdown remains unresolved.
No main merge. Source/tests/runbook/this ledger are portable; cached binaries/raw
receipts remain local. Review disposition is recorded below before checkpointing.

## Independent review and final verification

Read-only reviewer found0 Critical/0 Important/1 Minor: add process-disposal failure,
both-disposal failure and cleanup failure after successful observation. Added those
cases without changing the module; final new suite48 assertions passed, exit0,
including suppression of success output when cleanup fails. No second review.
Reviewer did not independently run tests. No deferred minor remains from this review.

Reviewer declined live CIM behavior, retained-handle authenticity/races, original
observer integration, application behavior, launch/termination policy and main
acceptance. Disposition: these remain expressly unverified/outside this pure-helper
checkpoint; future callers must supply authenticated same-handle evidence and
separate bounded authorization. Do not treat the guard as a completed live runner
repair or a replacement for those checks. No broader acceptance is granted.
