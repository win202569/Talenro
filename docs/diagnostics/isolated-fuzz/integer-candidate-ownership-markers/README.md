# Later ownership markers — PREPARED, NOT RUN

Preparation only under the approved bounded design. No controller is supplied.
Do not execute the target entry during preparation or assume a new run allowance.
Derives from ../integer-candidate-three-phase/: seven other PowerShell sources
are byte-identical; ownership.tests adds only four direct Console timestamps
at the first/last statements of SequentialIsolation and PlainControls bodies.
Removing those four lines recovers the original ownership source exactly.
Shared helpers,fixtures,API calls/arguments,mutations,assertions and order stay.
END is inside the body; the unchanged shared helper emits PASS after return.
No per-call wrappers,timers,finally ends,reloads,caches or guard changes.

Source path/body changes,compilation/history and four output calls can affect
timing. Runtime output ordering and actual group spans remain unverified.
Prior three-phase clone interference also remains; not unchanged API acceptance.
Closed body spans include fixture/decorations/API/equality/mutations,not dispatch
before START or PASS emission after END; not individual API or CPU shares.
Do not sum overlapping parent/child intervals as independent work.

expected-output.json specifies105 nonempty ordered records,64 PASS,22 markers,
11 pairs. OwnershipCase names SequentialIsolation/PlainControls have empty Parent
and are sequential top-level spans after prior scopes close.
Removing their four records restores the101-line inventory in order.

../candidate-ownership-capture-verifier.ps1 defines Test-CandidateOwnershipCapture.
It reads only the pinned105-line inventory and supplied text/receipt facts,
never imports or invokes targets. Retains closed pairs and OpenScopes.
OwnershipProgress states unobserved,entered,ended-awaiting-pass,passed distinguish
entry/body end/PASS; no inferred API/iteration/root cause. Termination distinguishes
actual receipt timeout from incomplete exit; malformed streams stay invalid.
Empty stderr,typed receipt facts,ordinary parent/child context,exact order/host,
Int64 monotonic ticks,nesting and full exit/budget gates remain required.
FormalProjectAcceptance always false. Not an authenticity or ownership proof:
future callers must independently verify provenance,engine/controller/source
identities,token,clock,owned-process lifecycle and saved stream hashes.

../candidate-ownership-capture-verifier.tests.ps1 exercises synthetic records only,
including all prefixes,ownership frontiers,errors and malformed spans. Its
optional verifier override permits RED checks against the unchanged101-line
classifier. No old91/95/101 classifier/inventory is modified or weakened.
PowerShell7 offline verification only;PS5 compatibility is unverified.

No real105-line run/timing is captured. All historical budgets and PS5 freeze
persist; no review/acceptance gate is consumed or cleared. A later run requires
separate approval for a NEW ordinary fresh PS7 single-host/30000ms/no warmup/
no retry/no fallback scope with source/controller pinning and bounded cleanup.
