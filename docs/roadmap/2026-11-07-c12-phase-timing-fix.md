# Phase timing assertion correction — 2026-11-07

User continued after c6b55f03 and the recommendation for bounded timing-only repair.
Scope: one RED/GREEN cycle (two test processes maximum), each180s outer/30s tests,
16 existing groups, synthetic data only; no release build/live input/full suite.
New exclusive evidence root .superpowers/sdd/2026-11-07-c12-phase-timing-fix.
Fixed Go/offline environment/owned cleanup unchanged. Any unexpected failure stops.

Read-only cause analysis: pinned time/time.go Since subtracts monotonic readings;
runtime/time_windows_amd64.s time.now and sys_windows_amd64.s nanotime1 read the
same interrupt-time counter. Equal readings legitimately produce zero. Existing
owner parser already permits nonnegative duration and requires START/END status.
The original positive-only assertion conflicts with that measured-time contract.
Original failure lacked exact phase/value; do not claim its precise clock history.

Reproduce deterministically with testing/synctest's frozen clock, using the actual
phase runner and synthetic files. No production clock injection, artificial sleep,
clamping to1ns, or change to hashing/path/process checks. After expected RED,
assert nonnegative timing AND complete ordered START/END evidence, matching digest
and expected value. Frozen-clock case must retain actual zero measurements.
This corrects a test assumption, not permission to accept missing END events.
Author inline review only;44/34/6/217,NOT_ACCEPTED. Build/Live remain blocked.

RED8fa46f/session18617 ->a3becf:24.37s,exit1 confirmed,no timeout.
Real-clock ZIP and DIRECTORY both emitted END/OK with elapsed_ns=0; PREPARE
3066700ns. Frozen-clock PREPARE/ZIP/DIRECTORY all0. Exactly the old positive-only
assertion failed; other15 groups passed. This reproduces the actual zero-duration
case as well as the deterministic boundary. Four module files unchanged.

GREEN2a7228/session76930 ->e1849c:19.08s,exit0 confirmed,no timeout,empty stderr.
All16 groups passed,including real-clock and same-tick subtests. Four module
files unchanged; phase.go/main.go/owner.ps1/vendor hashes unchanged from c6b55f03.
Only test assumptions and additional completion assertions changed. No runtime
sleep/clamping/security changes. Two allowed test slots consumed; no extra run.
Portable timing-fix evidence preserves both receipts/streams and final sources.
Original v1 evidence remains untouched. Author review: negative durations, missing
END,wrong sequence/status/digest now fail this test; fixed-clock nonzero result
also fails. This is scoped diagnostic testing,not whole-project acceptance.
Next: separately resume the bounded build/live plan after pre-build review;
neither build nor live invocation is authorized/executed by this repair turn.
