# Host versus successful prefix: read-only comparison

Status: ROOT_CAUSE_UNCONFIRMED / NO_NEW_EXECUTION / NOT_ACCEPTED.
User approved continuation of read-only comparison after checkpoint07f2f155.
No operational edits, extracted function calls, diagnostic launches or test runs.
Counters remain27aux/6hosts/3compiles/217snapshots; old stops remain closed.

## Evidence checked

Rechecked2396bound source/history/helper hashes against timing input; all match.
Original host final still reports timeout, owned exit and completed streams,
exit-1, no primary/secondary error. Its journal contains only begin; stdout is
28bytes (PASS: ContractAPIAvailable), stderr0bytes. PostPinsIntact=false is not
proof of drift: driver.ps1:1465 skips post-verification after timeout.
Diagnostic27 independently completed the instrumented prefix in14458ms with
610paired events and no output. It did not run the full host.

## Differences and their evidentiary limits

| Boundary | Original host | Diagnostic27 | Conclusion |
| --- | --- | --- | --- |
| Launch | capture.ps1, Grant/epoch/freeze arguments | child helper, input/intent arguments | Same fixed7.6.5 non-admin runtime checks; different entry and scope |
| Setup | imports boundaries, publisher, helpers, guards, full suite | imports14allowlisted definitions, two instrumented | Successful availability output proves original setup reached the suite; no timestamp tells how much budget setup consumed |
| Guards | captures six function references and collector identity before tests | absent | recording.tests.ps1:2-18 observes references; no continuous interception or timed loop here. Difference alone is not evidence of cause |
| Prefix | manifest, three historical files/contexts, first copied baseline | same ordered data path with timing insertions | Work counts match99pin calls/42parses; one fast instrumented sample does not reproduce the original context |
| Assertion wrapper | Invoke-ODContractCase invokes Body through a pipeline, updates suite/details, calls LS-Result then LS-Need | directly collects assertion output | Wrapper and first result-emission path remain unmeasured; direct assertion success does not validate them |
| Parent streams | CopyToAsync starts on BOTH streams before WaitForExit | polls ReadAsync with output caps | Not the elementary wait-before-read deadlock pattern. Completion after termination does not prove normal-time flow was healthy |
| Error/publication | catch, Capture-Failure, Journal-ODFailure, flush, error summary, publication | separate structured result/event path | No original error event cannot exclude a stall before/during error recording; cannot infer successful publication |
| End scope | intended52groups plus recorder/publication | stops after first baseline assertion | Timing success is not full-suite success |

Source anchors (repository-relative):
- docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/capture.ps1:94-136.
- Same directory observation-contract.tests.ps1:14-27 and95-153.
- Same directory test-helpers.ps1:10-19, recording.tests.ps1:2-18.
- Same directory driver.ps1:1428-1468.
- Same directory runtime-prefix-timing-child.ps1:28 onward.
- Full measured timings:2026-11-04-runtime-prefix-timing-result.md and portable
  runtime-prefix-timing-v3-final archive.

## Narrowed interpretation

No evidence supports declaring a stream deadlock, guard defect, permanent slow
parser, or fixed timeout bug. The99pin lookups took5.937seconds in the diagnostic,
not120seconds; that measurement cannot be substituted for the old run.
Original LS-Result writes a line without an explicit per-result Flush; final
capture flush exists. Missing next line alone is not an instruction-level probe.
The time spent before availability, the full original scope/wrapper, transient
environment behavior and error/output paths remain unresolved.

## Recommended next decision, not execution authorization

If further runtime investigation is wanted, approve a separately designed,
versioned real-host stage-tracing attempt. It should preserve original source
identities and all stopped routes; mark setup/availability, source loading,
each context, wrapper entry/body exit, suite update, result write/flush and error
entry. Keep the original process ownership,120000ms ceiling and cleanup rules.
Specify a new finite computation/process allowance and namespace before any
implementation/run; no assumption that the exhausted27/6 allowance can be reused.
Do not optimize pin lookup, replace output handling, increase timeout or remove
guards based only on this comparison. This report makes no repair claim.
