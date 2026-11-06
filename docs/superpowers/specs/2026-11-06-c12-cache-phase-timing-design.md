# C12 selected-cache phase timing — short design

Status: DESIGN_READY_FOR_REVIEW. User selected approach1 after3173882f;
the concrete choices below are proposals, not execution authorization.

## Intent and evidence

The v2 root verifier timed out at120s despite22 samples with increasing read/CPU
counters. Existing evidence cannot attribute time to a module, ZIP hashing or
directory hashing. See docs/roadmap/2026-11-06-c12-wait-v2-analysis.md.
Goal: identify which selected module/phase consumes the bounded diagnostic time,
including when a phase never finishes. No claim to diagnose physical disk latency,
antivirus, lock ownership or the original root command from timing alone.

## Approaches and selection

1. **Selected by user:** isolated, limited module/phase timing. Less collection
   scope; explicit phase events. Requires a separately built helper and does not
   reproduce graph loading/concurrent scheduling of the authoritative command.
2. **Not selected:** OS file-I/O tracing. More direct original-workload attribution,
   but broader tooling/permission/privacy scope. No tracing is authorized here.
Repeating the unchanged root command adds no new attribution and is not proposed.

## Proposed sample and order (author choices for approval)

Two exact root go.mod entries, verified read-only in this design turn:

| Order | Module | Version | Reason for selecting |
|---|---|---|---|
| 1 | golang.org/x/text | v0.40.0 | Smaller contrasting candidate; earlier ZIP inventory ~29.6MB expanded |
| 2 | modernc.org/sqlite | v1.49.1 | Largest earlier inventoried candidate, ~259.4MB expanded |

These are illustrative size contrasts, not statistically representative samples
or evidence they caused the root timeout. Inventory sizes are historical metadata,
not newly remeasured. No automatic substitution if either candidate is absent.
Fixed sequence: text ZIP, text directory, sqlite ZIP, sqlite directory. Later
phases remain NOT_RUN if a phase errors or the global deadline expires.

## Helper, equivalence and storage boundaries

- Separate diagnostic Go helper plus PowerShell owner/watchdog, under existing
  diagnostic documents/tooling; no production runner,module file or pin changes.
- Use the exact vendored dirhash source shipped in pinned Go1.26.5, with its
  copyright/BSD license retained and source hashes recorded. Isolated local
  standard-library-only diagnostic module; no package downloads or dependency
  updates. Do not silently use another x/mod version or a PowerShell hash surrogate.
- Call HashZip with DefaultHash, then HashDir with canonical module@version prefix,
  matching the inspected verifier's content hashing semantics. Do not modify or
  rebuild the installed Go binary. This helper omits LoadModGraph and many verifier
  behaviors; even matching digests are not a root verification result.
- Offline build uses the exact pinned Go executable/SHA256 from v2, GOTOOLCHAIN=local,
  GOWORK=off, GOENV=off, GOPROXY/GOSUMDB/GOAUTH=off, GOVCS=all:off and explicit
  GOMAXPROCS=2. Isolated build cache/output and temp directory inside a fresh
  diagnostic workspace. Build writes are allowed only in that future approved
  workspace; selected module cache contents are read-only inputs.
- Record helper source and binary hashes. Build may warm OS/toolchain caches;
  that affects comparability and must be disclosed. No cache purge or security
  exclusion to manufacture a cold baseline. Missing build prerequisites stop work.
- No copying/extracting/modifying real module content. Only exact allowlisted
  cached ZIP,ziphash and extracted directory paths may be read. Validate canonical
  paths/containment, reject reparse ancestors and entries before content reads;
  reject traversal/malformed ZIP names and inconsistent module prefixes. A detected
  safety mismatch stops rather than broadening allowed paths. Cache mutation/races
  cannot be ruled out by metadata alone; record that limitation, not a snapshot claim.
- Compare each completed digest against the captured ziphash and matching root
  go.sum module-content entry. Missing/conflicting expected digests, read failures
  or mismatches stop; never repair/re-download cache contents automatically.

## Bounded run and evidence

- One restricted owner and one owned helper; no elevated collector/CIM/system trace.
  Sequential phases, not two concurrent module workers; do not compare duration
  directly with GOMAXPROCS2 scheduling in the root verifier.
- Proposed one live helper invocation, global120s measured by owner Stopwatch
  immediately before launch. Preparation and offline build/test budgets are separate
  and must be specified/approved in the implementation plan; no open-ended build.
- Deadline/failure stops only retained owned helper process/tree; wait at most5s
  for exit confirmation. Unknown exit blocks further work. No retry,second candidate
  substitution,timeout extension or next workload after a stop.
- Helper emits and flushes sequence-bound START/END events for each module/phase;
  parent drains asynchronously into bounded append-only evidence. Start event precedes
  all phase work. Measure monotonic duration, retain CPU if available, digest/status
  for completed phases, expected versus completed phase counts, and explicit end
  reason. No END means INCOMPLETE, never a successful zero-duration phase.
- Keep path safety/enumeration time separate from ZIP/directory hashing time, with
  preparation start/end markers under the same120s run limit. Otherwise slow directory
  prechecks could be mistaken for a fast hashing phase. Do not emit per-file paths,
  file contents, environment dumps, stack traces or unrelated process metadata.
- Bound each event/record to64KiB and total retained live diagnostic data to4MiB;
  malformed/partial events retained but not accepted as complete evidence. Stop on
  quota/protocol failure. Fresh exclusive root, immutable final receipts and hashes.
- Save exact source/build/run coordinates, original receipts and portable evidence;
  no binaries,caches,credentials or real module contents in Git. Another computer
  can rebuild after its own prerequisite checks; cloning does not restore caches.

## Required verification before a live run

The future plan must include synthetic fixtures with literal known expected digests
for ZIP/directory equivalence, changed content/mismatch, malformed/escaping paths,
reparse rejection, missing expected hashes, phase order, START without END, quota,
timeout/owned cleanup and read-only input preservation. No real cache workload in
unit tests. Distinguish fixture tests from live evidence. Define exact test/build/
run counts and fresh roots before implementation; stop unexpected results.

## Interpretation and approval boundary

If directory hashing is slower here, that identifies a selected-phase cost under
this helper, not the cause of the original timeout. ZIP hashing includes entry
decompression and hashing; directory hashing includes enumeration/open/read/hash.
Neither isolates raw storage latency. All four phases finishing validates only
selected content comparisons, not graph completeness or full C12 acceptance.

Small diagnostic design, inline author review only. This turn writes documents,
does not implement,build,test,launch Go/helper,trace,change security settings or run
Docker/C12/devtools. After written design approval, produce an implementation plan
with explicit budgets for separate approval.44/34/6/217 unchanged,NOT_ACCEPTED.
