# Original Envelope observation — bounded diagnostic continuation

Latest actual continuation2026-10-15: preparation001 was invoked once and stopped
at the origin command's natural-completion gate, before build/fuzz. Version passed;
origin root exited0 but a listed member lacked initial ownership/state proof
(QueryError87). Owned cleanup was verified but does not pass the gate.701 product
hashes unchanged; sampling remains UNRUN, preparation001 is consumed. See [new
actual-stop report](2026-10-15-envelope-preparation-stop.md). Earlier readiness
observations below are historical, not permission to repeat this preparation.

Continuation2026-10-15: the user requested the bounded final-exit diagnosis. One
instrumented observation did not reproduce the old failure; fresh scoped coverage
is184 PASS after permission-bound link coverage was supplemented. The historical
failure's cause remains unknown, and original Envelope build/sample are still
UNRUN. See [retained evidence and limits](2026-10-15-final-exit-diagnostic.md).
The stopped2026-10-14 run below is preserved, not retrospectively passed.

## Approved scope

After the six minimal actual-framework diagnostic arms completed, the user approved
one original-project `FuzzIssuerEnvelope` observation sample using the verified
instrumented Go1.26.5 copy. Product sources remain unchanged. Build once, bounded
600 seconds; fuzz5s/parallel1/count1/internal timeout30s/external limit60s, with new
build/fuzz/temp directories. No Pair, full matrix, downloads, installation, official
toolchain changes, automatic retry, main integration or acceptance upgrade.

## Preparation and source-binding diagnosis

Tracked `envelope-observe.ps1` selects the original package/test/parser sources,
requires the pinned copied executable and six installed diagnostic-source hashes,
checks the complete copy receipt, enforces offline readonly modules, and records
the original product inventory, compiler evidence, binary hash, trace and owned-job
completion. Existing run identities are refused. Observed results explicitly set
`Acceptance=false`; stopped preparation is retained without another attempt.

No product build/fuzz was launched while preparing this driver. A real source-binding
test failed before launch: the same six source files were sorted differently by
Windows PowerShell5.1.26100.9711 and PowerShell7.6.5 (both zh-CN). The former computed
`18067D017BAE895F36E06E71BB42A4C04AF2ECA0A914A3285AC419FC807B45A3`, the latter the
recorded identity `E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306`.
This is an identity-serialization defect, not evidence for the historical fuzz
deadline's cause or product source drift.

An independent fixture test reproduced the failure under Windows PowerShell5.1.
The repair uses the historical literal six-path order instead of runtime collation;
it preserves the recorded identity, existing hashes, source validation and binary
mismatch rejection. Historical receipts and instrumented toolchain sources were
not rewritten. The eight provenance fixture assertions subsequently passed under
both runtimes. All21 driver assertions also completed successfully under each
runtime, without launching Go (29 checks per runtime,58 executions/29 distinct
assertions). Both scripts returned0. These do not cover the failed native suite.

## Current execution state

The fresh scoped suite stopped in `final-exit.tests.ps1`:13 assertions passed
(11 inherited same-handle checks and2 final-exit checks), then
`later signaled exit refreshes code only in final evidence` failed. This aborts the
suite; remaining suites were not executed and the earlier183-PASS checkpoint is
historical evidence only, not a fresh passing result. No automatic test retry.

The short real owned-child fixture requests150ms sleep followed by exit7, while
the final proof is bounded by the shared500ms window. Its compound assertion
requires exited state, successful exit-code read and code7. The failed run did not
persist the final row, so the precise failed conjunct and timing are not known;
host scheduling or timeout must not be asserted as the cause. The fixture's finally
requests termination only of its owned job and closes handles; termination success
was not independently persisted/verified by this failed fixture. No old-PID/name-based cleanup.
Readiness/release files remain under the unique ignored final-exit test root.

Safety-check failure stops this continuation **before Go launch**. Original
Envelope build and sample are **UNRUN**, and `envelope-observe-001` does not exist
(fresh read at UTC18:51:46). The readonly plan inspection covered701 product Go/SQL/module
files and selected only the original Envelope package, not Pair. A final no-match
`rg` inspection returned1; that was not a Go build or sampling failure.

Do not lengthen the final-exit window, suppress the failure, rerun until passing,
or consume the authorized sample before resolving this safety gate. Next work is
a bounded diagnosis that preserves the final row/clock/result and distinguishes
fixture timing from native-runner behavior. This requires direction beyond the
stopped sampling step. No production fix or completed observation is claimed.

Original matrix3 failure/Pair-unrun, real environment, C11/C12/Task10/B01 and main
acceptance remain open. Even an observed exit0 under this instrumented copy cannot
clear those gates. No new independent review has been performed.

## Fresh-computer recovery

Recover the latest `codex/c12-b01-task9-coordinator` commit, read current status and
`docs/runbooks/repository-recovery.md`. Tracked scripts/templates/report preserve
the workflow; ignored copied toolchains, caches and raw receipts are not uploaded
and cannot be assumed to exist on another computer. Independently prepare and
verify the pinned Go1.26.5 copy and existing module cache without downloads in this
sampling scope. Do not use the installed official toolchain as the patch target.

Dot-source `docs/diagnostics/isolated-fuzz/envelope-observe.ps1`. Planning with
`Get-EnvelopeObservationPlan <approved-new-absolute-run-root>
<verified-copied-toolchain> <existing-module-cache>` does not create directories or
launch Go. Actual `Invoke-EnvelopeObservation` needs explicit sampling authority
and a fresh identity beneath the owned2026-10-14 diagnostic root. Never repeat a
completed sample just to restore progress, reuse an existing run root, or treat a
stopped/missing receipt as permission for a retry. Preserve failures and request
direction if a new sampling attempt or wider validation is needed.
