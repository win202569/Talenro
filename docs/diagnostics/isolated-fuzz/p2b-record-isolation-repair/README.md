# Narrow record-isolation harness

Authorized plan: docs/superpowers/plans/2026-10-31-offline-record-isolation-repair.md.
The only repair is a no-members export declaration in case-module.ps1 between
the frozen seed and green phases. No frozen production/test source is changed.
The separate publisher is recording instrumentation; its successful checks do
not close I1/I2, grant runtime acceptance, or restore historical allowances.

Fixed sequence, only under the explicit two-host authorization:
Prepare seed; RunOne isolation-red with actual seed SHA; edit builder only;
Prepare isolation-green with actual previous SHA; RunOne isolation-green with
its actual SHA. Exactly two capture hosts, no target, compilation or review.
All8sources/expected values freeze before RED; only builder changes for GREEN.
Any unexpected failure freezes the remainder. No retry/namespace switch.

Publisher roles HostEvidence, Healthy and ExistingFailure are fixed in advance,
not fallback paths. First failed publisher is latched before subsequent IO.
Actual errors retain their in-process reference identity until serialization.
Parent verifies raw streams/exit/evidence/journal/known case files and all pins.
No historical driver/module import, fixture compilation, Session/Ack or profile.
Portable records are preservation data, not execution rights or original errors
recovered from the failed historical writer-red. No main merge is included.
