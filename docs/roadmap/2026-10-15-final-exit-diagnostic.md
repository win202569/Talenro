# Final-exit diagnostic continuation — 2026-10-15

Subsequent actual preparation001 stopped at origin before build/fuzz: initial
member ownership/state could not be proved (QueryError87). This later result is
not fixed by the passing tests below. See [current preparation-stop evidence](2026-10-15-envelope-preparation-stop.md).
The former unconsumed-identity/readiness statements below retain the earlier
checkpoint's state; preparation001 is now consumed and must not be repeated.

## Scope and result

The user requested continuation after checkpoint8d6d95f1. This step adds diagnostic
records to the existing final-exit test fixture and runs one instrumented diagnostic
suite, followed by the other scoped runner tests. No production runner, timeout,
acceptance predicate, toolchain, product source or original fuzz callback changed.
No Go executable, original Envelope build/fuzz, Pair or full product matrix ran.

The historical failure was **not reproduced**, not fixed or retrospectively passed.
Its final row was not saved, so its failed conjunct/root cause remains unknown.
The current healthy observation alone cannot prove host load, scheduling, shell
shutdown or a native defect caused the old failure. Do not repeat until passing or
increase the500ms final-exit window. All production/main acceptance gates stay open.

## Diagnostic-only fixture change

`final-exit.tests.ps1` prints its new unique owned fixture root and saves
`short-final.json` / `long-final.json` immediately after the native call, before
the existing assertions. Each contains the unchanged returned member row, runtime,
case, requested child sleep, shared window, after-call clock and UTC. Serialization
is outside the native wait and cannot add to its accepted budget. The fixture
uses a real retained job-verified process handle with intentionally constructed
running/core/image-error input; that synthetic image error is not a new API failure.

Finally, `short-cleanup.json` / `long-cleanup.json` record owned-job termination API
results. `EmptyJobVerified=false` explicitly means this fixture does not measure
whole-job emptiness; it is not a completion receipt. Handles are still closed in
finally even if writing the cleanup diagnostic fails. Every new record is tagged
`Acceptance=false`; raw fixtures remain ignored and are not uploaded.

## One diagnostic observation

PowerShell7.6.5, UTC09:21:40–09:21:44 (Asia/Tbilisi13:21:40–13:21:44):

| Case | Native evidence | Interpretation |
| --- | --- | --- |
| Short child, requested sleep150ms | Start10ms, budget490ms, end213ms; exited, exit-code read succeeded, code7 | Current positive case passed inside the original shared500ms window; core running snapshot preserved |
| Long child, requested sleep30000ms | Start1ms, budget499ms, end500ms; still running, no exit-code read | Correctly rejected; a later member received no fresh window |
| Native producer / real short descendant | Root exit0, no timeout, final member proof by157ms, natural job empty, no forced termination | Saved schema3/v2 receipt passed serialization validation; accounting0/member list empty |

The two simple fixture cleanup API calls reported success; empty-job verification
was not performed for those fixtures. The separate native-producer receipt did
verify empty-job completion. Do not confuse these different evidence levels.

Raw diagnostic root, relative to the development worktree:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/final-exit-tests-ec9f8e56af284d349d54abfa5e4e5b1c/`.
Do not query these historical PIDs or reuse the run identity.

## Fresh scoped verification

The single instrumented final-exit suite returned0 with52 PASS. The remaining eight
scripts returned0: shutdown-observation33, shutdown-state9, run31,
runner-regression21, materialize8, provenance8, preflight4 and isolation15.
The first isolation run reported two groups UNRUN because sandbox link creation was
unavailable. After inspecting the exact two-file fixture/link targets, an elevated
isolation-only run passed all18 checks, including actual junction/ancestor-junction
and multiple-hardlink rejection. No install, security-setting change, broad write,
native timing-test retry or product test was used to obtain this link coverage.

Thus the full scoped set has184 PASS checks (173 distinct because11 same-handle
assertions are reused). Actual assertion executions this turn total199, including
the15 overlapping isolation checks repeated solely for the permission-bound link
coverage. No scoped check remains UNRUN after that supplement. This is author
verification, not an independent review or product acceptance; the full product
suite was intentionally outside this diagnostic scope.

## Next step and recovery

The old failure is unresolved but now the fixture can retain evidence on recurrence.
The current scoped runner checks are green; this is not evidence that the original
Envelope failure has disappeared. Original Envelope sampling remains UNRUN and
`envelope-observe-001` has not been consumed. A subsequent bounded continuation may
consider the previously approved single observation with its original provenance,
offline, fresh-cache and safety checks; failure still stops without automatic retry.
Do not merge main or declare matrix3/Pair/C11/C12/Task10/B01 accepted.

Necessary test/report/recovery changes are preserved on the existing development
branch; actual commit/push and remote SHA verification are reported at handoff.
Recover the latest `codex/c12-b01-task9-coordinator`, not automatically main, and
read current status plus `docs/runbooks/repository-recovery.md`. This fixture suite
does not need an old Go copy/cache: it creates unique tiny local fixtures and uses
the current PowerShell runtime. Old ignored receipts, toolchains and live process
state are not recovered by cloning. Do not rerun completed diagnostic arms just
to recover progress; any future diagnostic execution needs an explicit purpose.
