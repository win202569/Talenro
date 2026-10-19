# Controller exception containment — offline repair candidate

This leaf is not a handshake/state/capability implementation. Old frozen seam
files, results and I1 review are unchanged. Plan:
../../../superpowers/plans/2026-10-19-p2b-controller-exception-containment.md

Seven inputs are frozen before RED. Only owned-run.ps1 may change once for GREEN;
both source revisions and manifests are retained. One new capture references
the original seam script/oracle, not a copied product. No product test hook,
setter, callback, native compilation, dependency installation or live action.

Budget: RED host+Hold child; GREEN host+3children (fault result shared by2checks,
normal exit, timeout); FINAL capture host+one original seam child. Total8newPS7
processes, conditional1readonlyreview, no retries. Pre-start file failure has no
Start attempt. Slot attempts are persisted before launch. Ordinary fixedPS7 only.

Fixture children emit fixed ASCII; Hold self-limits at45000ms. Callers retain
the exact new Process object until independent exit observation/rescue; no PID
or process-name lookup. A termination request is not proof of exit. Real Kill
failure/OS stalls/concurrency/hostile tampering are not dynamically guaranteed.

Portable source and fixture data survive a clone; ignored packets, old raw,
engine/SDK, credentials and runtime authority do not. Do not automatically rerun
Prepare/tests/capture on recovery. Read plan/current status and remaining budget.
