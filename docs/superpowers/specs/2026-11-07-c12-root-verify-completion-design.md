# Root verifier completion attempt — evidence and scope

Status: APPROVED AND EXECUTED once;300s deadline reached,STOP/NOT_PASSED.
Result: docs/roadmap/2026-11-07-c12-root-verify-result.md. Remaining proposal
language below preserves the original rationale;it does not authorize another run.
This is an execution-budget proposal using existing tooling,not a new verifier.

## Evidence synthesis

1. Prior authoritative root run ended at120.048s without output;22 samples showed
   increasing process read/CPU counters in all21 observed intervals. Neither
   corruption nor success follows from that timeout; partial stalls remain possible.
2. Selected-cache v3 completed in56.717s with4 matching hashes and12 complete
   events. text ZIP0.120s/directory12.378s;sqlite ZIP0.761s/directory42.456s.
   Directory phases are96.86% of that sample's phase time,not of the original
   root verifier. Graph loading,other modules,concurrency and cache warmth differ.
3. Read-onlyf16505 revalidated all25 phase archive entries and matched the prior
   root archive's saved SHA256995D2D1AF7A3BA34EF7DFF801024F08727EFE6AD6D678166C979935BE0267519.
4. Pinned Go verify.go inspected99a511: LoadModGraph,at mostGOMAXPROCS workers,
   ZIP then directory verification,success printed only after collection. Modules
   absent from both ZIP/directory cache are skipped; root/toolchain entries also
   skipped. Success is native cached-dependency integrity,not full cache availability,
   application tests,buildability,or the helper's explicit root go.sum comparison.

Sources: docs/roadmap/2026-11-06-c12-wait-v2-analysis.md and
docs/roadmap/2026-11-07-c12-phase-v3-result.md;their portable archives remain canonical.

## Recommendation and tradeoff

One native root go mod verify attempt with a300000ms cap,not repeated120s runs.
300s is a proposed finite resource budget (2.5 times the old cap),NOT an estimated
completion time or an extrapolation from two modules. User approval is required.
Compared with more surrogate profiling,this directly asks whether the actual
verification can complete. It may still time out; that outcome ends this attempt.

Use unchanged corrected Invoke-C12PhaseOwned for bounded raw stdout/stderr and
owned cleanup,without a phase Session: native Go output is not helper JSONL.
No collector,CIM,trace,process inventory or extra progress worker. Cost: no
within-run phase/CPU/I/O attribution; a second timeout does not establish cause.
Keep GOMAXPROCS2 and original offline/root/cache coordinates. Reuse tested owner,
not older collector/handshake code. No production changes or diagnostic rebuild.

## Boundaries and interpretation

- One metadata-only prerequisite check;at most one Go child/process tree.
-300s measured before launch;stop only owned tree at cutoff,confirm exit within5s.
  Existing bounded drain/cleanup may add up to5s each;no extra Go runtime allowance.
-4MiB retained records/streams under existing owner caps;immutable exclusive root.
-No root/devtools go.mod/go.sum edits,cache clearing/repair,downloads,security/Git
  configuration changes,concurrency changes or automatic retries.
-Native Go receives the existing cache to inspect,not a copied/empty substitute.
  Offline settings are not a filesystem read-only mount; do not claim atomic
  immutability of all external cache files. Capture four project module hashes.
-Only exit0 + no timeout/failure + confirmed exit + exact native success output
  + empty stderr + unchanged module files earns ROOT_CACHE_VERIFY_PASSED.
  Missing dependency/network-disabled errors are failure,not permission to fetch.
-A success does not increment44/34/6/217,run devtools/C12 or merge main;overall
  NOT_ACCEPTED remains until separate wider acceptance work is completed.
-A timeout,nonzero exit,quota error,unknown exit or altered module files stops.
  Preserve evidence;no second attempt,timeout increase or repair under this budget.

Inline author review retained. No root command or new tests run while planning.
