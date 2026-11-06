# C12 shadow source-audit integration design

Status: DESIGN_PENDING_APPROVAL. User selected approach1: parallel recording,
not a new formal pass/fail gate. This document does not authorize implementation
or a live run. Baseline5b1dfdd2; existing counters44/34/6/217 unchanged.

## Intent and scope

Introduce opt-in evidence alongside the formal C12 runner to learn whether a
source-input audit can integrate with its lifecycle. Formal test assertions,
resource ownership, cleanup requirements, exit codes and capability flags remain
authoritative and unchanged. Shadow success is never formal acceptance.

Assistant-proposed first increment: audit an explicit allowlisted set of stable
repository source/configuration files only. Do not transplant diagnostic
observation44's historical manifest, fixed-root driver, compiled runtime or its
claim of whole-input coverage. Temporary executables and prepared bundles retain
their existing receipt/ownership verification; they are not shadow inputs because
normal cleanup removes them. This narrower coverage needs design approval.

## Approaches

1. **Selected by user: shadow recording first.** Keep formal decisions unchanged;
   report separate MATCHED/DRIFT/INCOMPLETE results. Lower promotion risk, but
   shadow drift does not fail an otherwise passing run.
2. **Deferred: mandatory audit gate.** Audit failure would fail formal execution;
   requires separately approved deadline/fault contracts and promotion evidence.

"Parallel" describes independent decisions, not background concurrency. Use
synchronous, owned calls; do not create fire-and-forget processes/tasks.

## Entry and configuration

Propose `-ShadowAudit` switch, default off, usable with existing Focused/Suite
parameter sets. With it off, no new directories, file reads, summaries or auditing
side effects. Existing package/profile/test allowlists remain unchanged.

Enabled mode uses a versioned manifest shipped in the repository, with explicit
relative paths and expected SHA256 values. No environment-variable activation,
user-supplied executable/script, recursive scan, wildcard, arbitrary absolute path
or automatic dependency discovery. Manifest and audit implementation identities
are bound in each record. Source-set membership must be reviewed in the later
plan against actual runner dependencies; omissions remain coverage limits.

Initial path policy: resolve under the bound source root, reject traversal,
reparse ancestors and duplicate/case-aliased paths. Validate before
hashing. Proposed fixed ceilings:128 files,32MiB aggregate bytes,1MiB manifest,
1MiB per summary. These are assistant proposals, not previously agreed quotas.
If limits or validation fail, record INCOMPLETE and do not silently sample.
An otherwise valid allowlisted path missing at either scan is DRIFT, not a
manifest validation error. Malformed/unsafe paths or an unreadable manifest are
INCOMPLETE. A missing file has a typed absent observation, never a fabricated hash.

## Ownership and lifecycle

- The runner remains sole owner of Go/Docker/resources/receipts and cleanup.
  Shadow code cannot start, adopt, release or remove those resources.
- Capture the invocation checkout root once before Suite switches repository
  roots. Create a separate exclusive evidence directory under that persistent
  checkout's repository-owned
  diagnostic output root, using a new random audit ID, never a reused group path.
  No caller-selected external destination. Reject occupied/reparse destinations.
  Retain records after runner cleanup; no new automated deletion policy.
- Bind a distinct source root per group: Focused audits its active checkout;
  Suite audits its actual groupSnapshot.Root and candidate.Tree, not the original
  checkout. Read the allowlisted manifest from that source root. Evidence remains
  outside the candidate snapshot, so it cannot dirty or vanish with the snapshot.
  Source stability here means stable during that group's scan window, not that a
  Suite snapshot survives its caller's cleanup.
- For each group, bind GroupID, profile, package, runSuffix, audit ID, manifest
  digest, implementation digest and candidate identity (when available; explicit
  null otherwise). Never claim a candidate identity was measured if unavailable.
- Capture before-map after group selection/identity is established but before
  its setup/workload starts. Freeze the selected list/digests for that group.
- Existing group execution and finally cleanup run unchanged. Capture after-map
  only after Invoke-C12Group cleanup attempts return and its failure state is fixed,
  while the Suite group snapshot still exists, before caller snapshot validation
  and removal. This is provisional evidence, not a final MATCHED publication.
  Hash only the stable repository inputs, never deleted prepared resources.
- If native/process ownership is uncertain or cleanup fails, emit INCOMPLETE
  without an after scan. Preserve primary and cleanup errors as authoritative.
  A stale before-map or another attempt's after-map cannot complete this record.
- Do not insert audit work inside cleanup loops or extend their deadlines. Hook
  implementation must preserve finally execution and outer environment restoration.
- Finalize shadow outcome only after the caller's snapshot validation/removal,
  suite candidate cleanup, outer prepared-artifact cleanup and environment
  restoration have returned. Use a separate run-summary referencing immutable
  provisional group records; never rewrite them. Any cleanup/ownership failure
  makes final shadow state INCOMPLETE even if maps matched. A process that never
  reaches this boundary has no final summary and is therefore INCOMPLETE. This
  preserves original group output/throw behavior and script exit selection.

## Evidence and decision contract

Each record has schema/version, mode=SHADOW, Synthetic=false for real runs,
attempt bindings, covered paths, before/after references (length/hash), elapsed
time, original test/cleanup outcomes and shadow state. Use strict JSON duplicate
key/type checking, including real Boolean values, informed by review-v4 tests.
Records are immutable exclusive writes; a final summary requires both persisted
maps and identity agreement. Validate exact bytes after publication before
reporting persisted MATCHED. Do not expose an ObservationPassed field that callers
could confuse with the existing verifier's decision.

|Shadow state|Meaning|Effect on formal result|
|---|---|---|
|DISABLED|Switch absent; no evidence produced|None|
|MATCHED|Complete bound maps equal approved expected pins|None|
|DRIFT|Measured source missing/changed against expected pins|None|
|INCOMPLETE|Invalid manifest, unsafe path, I/O failure, uncertain ownership or incomplete evidence|None|

Mismatch in the before-map is DRIFT too; do not recalculate expected values to
hide it. Preserve first error and completed partial records. If final persistence
fails, do not fabricate a file or retry/overwrite; emit a redacted best-effort
warning naming only audit ID and a closed reason code. Missing summary means
INCOMPLETE, never MATCHED. Audit exceptions must not mask formal primary/cleanup
errors or change the chosen exit status. Do not suppress existing runner errors.

Do not archive environment dumps, connection strings, passwords, nonce/HMAC keys,
database data or raw Go/Docker output. Evidence contains allowlisted metadata and
file digests, not file contents. Human console output uses a distinct
`C12 shadow audit` prefix and never prints a second `C12 group ... passed` line.

## Timing and limits

Synchronous hashing/publication can add wall time and filesystem calls can hang.
Byte/count bounds are not hard timeouts. Check the existing remaining deadline
before starting each audit stage; if expired, skip with INCOMPLETE. Never reset or
extend group/suite/native/cleanup deadlines. A call begun before expiry can still
finish late. Formal verdict logic is unaffected, but timing/scheduling may differ
when opted in; no claim of zero execution impact or hard audit cancellation.

A true no-delay guarantee would require a separately designed ownership-aware
worker/deadline mechanism; it is not authorized here. Default off preserves the
current path. Existing cleanup always takes priority over shadow after-scan.

## Components and verification before live use

Later plan should separate a definition-only shadow audit module, pure manifest
and outcome logic, fixture tests, and minimal runner lifecycle hooks. Use existing
PowerShell/.NET only; no imports that run workloads. Builder chooses internal
names, schema field layout and exact hook locations within these constraints.

Required offline evidence: off-path equivalence; valid before/after; drift/missing
file; malformed/duplicate/type-invalid manifest; unsafe/reparse/occupied paths;
wrong attempt binding; exclusive-write and summary failure; early deadline skip;
unknown exit/cleanup failure skip; primary-error preservation; no secret leakage;
every formal pass/fail outcome identical with shadow success/drift/error. Tests
must exercise real small-file I/O and actual extracted hook logic, not only mocks.
No live Go/Docker or full integration run follows from approving this design:
the implementation plan will state finite test budgets and require approval.

## Non-goals and remaining acceptance

No main merge, capability promotion, mandatory gate, complete dependency audit,
atomic filesystem snapshot, loaded-module attestation, hostile-local-admin trust,
hard disk timeout or original host-timeout root-cause claim. Prior16 verifier and
10 parent persistence cases support patterns, not proof of this new integration.
Disk-full/mid-write/flush/stop-marker persistence, clean-machine preparation and
independent security acceptance remain open. Promotion to a mandatory gate needs
its own reviewed design and actual integration evidence.

## Builder check

One independent read-only builder check identified source-root ambiguity in Suite,
the outer snapshot cleanup boundary and missing-file state inconsistency. Resolved
above: actual per-group candidate source, persistent checkout evidence root,
provisional maps plus post-cleanup run summary, and valid missing paths=DRIFT.
These are assistant design decisions for user review, not previously approved
implementation. The check is not a security or runtime acceptance review.

## Approval checkpoint

User selected only approach1. Assistant proposals needing review here: stable
source-only first increment, default-off switch, explicit manifest/size ceilings,
retained metadata-only evidence and synchronous timing caveat. After document
approval, write an implementation plan; do not implement in this design turn.
