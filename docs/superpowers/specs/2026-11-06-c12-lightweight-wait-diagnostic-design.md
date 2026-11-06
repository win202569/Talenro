# C12 lightweight wait diagnosis — proposed design

Status: DESIGN_AWAITING_APPROVAL. User selected approach1 after3c1341b8;
selection authorizes this design, not another verification run.

## Intent and evidence

Two120s root verification attempts timed out. The sampled attempt accumulated
13.828s CPU with a roughly18–20MB working set, but had no I/O or waiting evidence.
Goal: collect enough process activity evidence to narrow hypotheses without
changing verification parameters or claiming a lock/disk cause prematurely.

Read-only self-process probes: b4906f could read thread states/wait reasons but
CIM I/O access failed in restricted execution. Normal-desktop514ed7 returned
read/write operation and transfer counts. These probes concern PowerShell itself,
not Go; they do not prove cross-process collection will work.

## Approaches

1. **Selected:** lightweight process I/O,CPU,memory and thread-state sampling.
   Small data scope; cannot identify a specific file/lock or establish physical
   disk throughput from process counters alone.
2. **Deferred:** detailed system event tracing. More attribution potential, but
   wider collection, privilege and privacy scope. No ETW/WPR/Procmon collection,
   tool installation or trace-session mutation is authorized by this design.

## Proposed execution boundary

- One new root-only go mod verify attempt, at most120s. Keep exact cached Go1.26.5,
  offline child environment and GOMAXPROCS=2 from the previous attempt. No devtools,
  compilation,C12,Docker changes,downloads,automatic retry or deadline extension.
- Preserve the verifier's prior restricted execution context. Use a separate
  normal-desktop **read-only collector** because CIM failed in the restricted
  context. Do not move Go into elevated execution to make collection convenient.
  If separation cannot be implemented safely, stop before starting Go.
- Before Go launch, verify cross-process collection on the live diagnostic parent
  only. This is capability checking, not a claim of Go behavior. Failure means no
  Go launch, no fallback to a different observation method without approval.
- Executor retains ownership of its Go process and120s timer, independent of
  collector calls. Identify the target by PID, creation time and expected executable
  identity, not process name alone. Refuse mismatches/PID reuse. Collector cannot
  terminate/adopt other processes or inspect their command lines/environment.
- Sample about every5s, at most24 scheduled samples. Allow at most5s per CIM query;
  never overlap/retry failed queries. Missing/unsupported fields are UNAVAILABLE,
  not zero. Collector failure does not extend verifier lifetime; stop its sampling
  and record the failure. Owned verifier is stopped at its original deadline.
- Record timestamp,CPU time,working/private memory, cumulative read/write operation
  and transfer counters, and thread-state/wait-reason counts when observable.
  Preserve raw numeric samples and deltas. Capture no file contents,environment
  values,network payloads or other processes. Counter interpretation is deliberately
  limited: process I/O is not necessarily physical disk activity.
- Preserve all claims,observations,outcomes and source/lockfile hashes in a fresh
  isolated root. Never reuse old roots. At timeout confirm owned-process exit;
  no follow-up test on exit uncertainty. Preserve old evidence unchanged.

## Interpretation and stopping rules

Increasing CPU/read counters supports ongoing work, not correctness or a completion
estimate. Low CPU plus changing I/O suggests an I/O-related activity pattern, not
a proven storage bottleneck. Low CPU and unchanged I/O plus repeated generic wait
states still does not prove a deadlock or a specific lock owner. No sampling result
alone justifies security exclusions,cache deletion,more concurrency or more time.
If evidence is inconclusive, report INCONCLUSIVE and propose the next scope; do not
repeat automatically. Successful go mod verify concerns the root graph/cache only,
not full C12 acceptance. Reparse and wider fault/containment gaps remain open.

## Handoff

Small diagnostic design; inline author review only, no security acceptance claim.
After document approval, write a bounded implementation/verification plan before
building the collector or launching Go. Include collector/executor coordination,
timeout cleanup tests and explicit evidence limits. Commit/push necessary files
under existing standing permission; no main merge.44/34/6/217 unchanged,
NOT_ACCEPTED. This design does not authorize execution by itself.
