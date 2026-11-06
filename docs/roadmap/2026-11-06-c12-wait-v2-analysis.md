# C12 v2 saved-observation analysis

Status: READ_ONLY_ANALYSIS_COMPLETE; next approach not selected. No new execution.
Source checkpoint9d779458; canonical portable-c12-cache-wait-v2.json contains22
real samples and exact source/receipts. Sample payloads were decoded in memory,
length/SHA checked, then analyzed (tool9da630). No old records modified.

## Findings and limits

- Sample1:5427.6505ms,ReadBytes641048365,ReadOperations220527,CPUms7578.125.
- Sample22:119655.7048ms,ReadBytes798191417,ReadOperations289181,CPUms18531.25.
- Across114.2280543s between these samples:157143052 read bytes,68654 read
  operations and10953.125ms CPU growth. Mean read-counter growth approximately
  1.31MiB/s; CPU approximately0.096 core-equivalents, not9.6% of the whole machine.
- All21 intervals have positive read-byte and CPU deltas. Interval read growth
  ranges0.506–2.986MiB/s,CPU0.054–0.158 core-equivalents (rounded). No stationary
  process-wide counters in the observed intervals; this is not proof every worker
  advances, nor proof no partial stall exists.
- Generic waits UserRequest/Executive/EventPairLow cannot identify a file, lock
  owner or stack. Snapshot wait states are not the fraction of wall time waiting.
- The first sample already has substantial cumulative reads. No interval before
  it was observed; do not infer an early fast phase or graph/ZIP completion from it.
- Process I/O counters can include cached/nonphysical I/O; they are not physical
  device throughput. No disk latency/queue/security-filter evidence exists here.
- About5.34–5.52s between later samples includes scheduling/query/observer overhead;
  do not present this as an exact5s cadence or an overhead-free baseline.
- Root verify timed out120048.1025ms,with empty output and confirmed exit. Neither
  cache corruption nor successful integrity verification follows from timeout.

## Installed source cross-check (read only)

Pinned toolchain source inspected via f99ebf/df857b:
`C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/src/cmd/go/internal/modcmd/verify.go`.
runVerify rejects module arguments(line52),loads the graph(line61),limits concurrent
module workers by GOMAXPROCS(line59),and emits success only after result collection
(line89). verifyMod hashes ZIP(line122),then extracted directory(line133).

Its vendored `src/cmd/vendor/golang.org/x/mod/sumdb/dirhash/hash.go`:
HashZip opens/decompresses entries and hashes file contents; HashDir walks directory
entries then hashes their contents through Hash1. Therefore ZIP bytes on disk,
expanded directory bytes and process read counters are different quantities.
The current evidence cannot map any sample to a module or one of these phases.
The prior99-declared-module inventory is not the fully resolved graph used by
LoadModGraph; do not divide observed bytes by its size to claim percent completion.

## Next approaches for user choice (not an approved design)

1. **Recommended: limited per-module ZIP/directory phase timing.** Prepare a
   separate bounded diagnostic helper, with a reviewed module allowlist and the
   pinned toolchain's hashing semantics. Emit module/phase start/end before and
   after each operation, so a cutoff still leaves useful attribution. No cache
   edits,security changes or system-wide capture. This needs separately approved
   offline build/test/run budgets; it is not `go mod verify <module>` (unsupported),
   not a replacement verifier and not production acceptance. Selection and
   sequential timing differ from the concurrent root run and can alter cache
   warmth,so comparison cannot alone establish root cause.
2. **Alternative: narrowly scoped OS file-I/O tracing of the original command.**
   May attribute operations/latency without a surrogate workload, but needs a
   separate feasibility/privacy/tooling review, explicit tracing authorization,
   size/time caps and cleanup. Some capture may initially include wider system
   metadata even if filtered later; no tracing/install/security changes authorized.

Do not repeat the unchanged root command, enlarge timeout/concurrency, delete
caches or add security exclusions on the basis of this analysis. Choose an
approach first, then write/approve its design and execution plan. This turn has
no Go,compilation,child fixture,CIM/live collector,Docker orC12 execution.
44/34/6/217 unchanged,NOT_ACCEPTED. No independent security acceptance.
