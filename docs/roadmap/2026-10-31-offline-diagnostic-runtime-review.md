# Independent readonly review — verbatim report

Reviewer: offline_runtime_final_readonly, gpt-6-astra/xhigh, fresh context.
Reviewed source checkpointa9e76e51; one authorized review, no reruns/fix/re-review.
The report below is preserved verbatim. Author dispositions are in the execution
report and portable ledger; both Important findings accepted, still unresolved.

## Assessment

Not ready for narrow acceptance: **0 Critical, 2 Important, 0 Minor findings**.

The recorded runs support substantial progress, including the non-confounded missing-ack control. However, first-cause preservation and the promised complete evidence validation have defects. Preserve the unaccepted checkpoint; this report does not authorize fixes, further tests, another review, or a main merge.

## Strengths

- The public entry defaults to read-only inspection, rejects unsupported requests before runtime import, and hard-codes its three runnable profiles independently of catalog data.
- Session and acknowledgement authority uses private reference identity, fixed step/epoch bindings, irreversible failure latching, and consumption before startup I/O. Disk outcomes do not participate in the continuation gate.
- Managed I/O rejects access after failure; owned-process cleanup retains the exact process object and avoids PID/name lookup, tree killing, and receipt retries.
- Preparation handles optional `Role`, required `ExpectedState`, consistent duplicate paths, and the two exact addition windows. The explicit source list excludes later portable exports.
- Stored tool records distinguish running sessions from completed commands. Sessions `59922`, `98389`, and `53127` have retained start/running/completion records; the earlier collector stop and explicit subsequent resume remain documented.

## Important findings

### I1. Disposal failures can replace the original record-write failure

Locations:

- [profiles.ps1:42](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/scripts/private/offline-diagnostics/profiles.ps1:42)
- [capture.ps1:57](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/capture.ps1:57)

Two related paths violate first-cause preservation:

1. During `profile.intent.json` creation, a write/flush failure is captured and rethrown. If `Dispose()` also throws, the `finally` block throws that later exception instead. The original remains only in the module’s private failure state; no `ProfileReport` is returned, so the outer capture receives the disposal exception as the invocation failure.
2. In `Capture-Once`, the original write/flush exception is initially stored only in local `$primary`. A subsequent disposal exception is immediately passed to `Capture-Failure`. If the host previously had no failure, disposal becomes `HostPrimary`; the original is added later as secondary.

This matters precisely when recording fails: the diagnostic handoff can report cleanup as the cause and obscure the failure that initiated it. The approved design explicitly requires cleanup and secondary failures not to replace the real first cause.

Correction direction: retain the first exception across the entire write/flush/dispose operation, record disposal exceptions only afterward as bounded secondary errors, and propagate the original across the public boundary. Preserve single-attempt/no-retry behavior.

This is a static finding. Neither compound failure was induced in the recorded runs. Both problematic branches are also present in the frozen snapshots.

### I2. “Exact” validation skips non-Step observation contents and error signatures

Primary location: [driver.ps1:193](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/driver.ps1:193).

Supporting locations:

- [core.tests.ps1:40](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/core.tests.ps1:40)
- [original normative contract:98](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/superpowers/plans/2026-10-19-offline-diagnostic-entry-migration.md:98)

For each private observation, the parent validates only the outer `Name`/`Value` keys. It examines the value only when `Name == 'Step'`. Consequently, `Read`, `Write`, `Complete`, `Output`, and their case-summary observations have no complete parent-side schema/value/error-signature validation.

The core tests do not close that gap. For example, `missing`, `bad`, and `index` require an unsuccessful operation and a non-null error, then report only failure/ack/start facts. An unrelated read/parse failure can satisfy those assertions. The frozen oracle does not bind these errors’ deterministic type/id/message and complete observation structure.

The actual stored records contain the expected missing-file, malformed-JSON, invalid-index, and disposed-writer errors. The problem is that the acceptance predicate does not establish that specificity: changed or missing non-Step details can still receive `EvidenceValid=true` and `Exact=true`. This contradicts the normative requirement to freeze and validate deterministic fields, order, types, and actual error signatures.

Correction direction: in a separately authorized revision, freeze the complete ordered observation inventory and per-case I/O/view/error contracts, reject unknown/missing observations, and limit variability to the expressly permitted fields. Do not retrofit the frozen oracle to these results.

## Recorded evidence examined

The six final records report the following, consistently with their retained stdout and tool-completion output:

| Grant | Capture exit | Direct targets | Recorded result |
| --- | ---: | ---: | --- |
| entry-red | 1 | 0 | Only `EntryPresent` fails; five NotRun |
| entry-green | 0 | 0 | Six PASS; fourteen rejected requests |
| runtime-red | 1 | 0 | Exact missing-backend error; fifteen NotRun |
| runtime-green | 0 | 0 | Sixteen PASS; 98 prefix conditions |
| lifecycle | 0 | 3 | Fourteen PASS |
| continuation | 0 | 4 | Thirteen PASS plus twelve original seam PASS |

All six finals record owned exit, completed streams, intact post-pins, no rescue, and empty stderr. “Exact” here means **reported by the current checker**, subject to I2.

The canonical target observations identify seven distinct targets: lifecycle `20468`, `22132`, `6724`; continuation `19392`, `17684`, `17104`, `18336`. Their recorded exits/drains and cleanup tuples match the intended receipt-fault, normal, timeout, and continuation cases.

The continuation evidence genuinely avoids the zero-budget confound: A is `Active`, `NextOrdinal=1`, `Reserved=1`, with its real acknowledgement available; its frozen case has two steps/budget two. Missing acknowledgement rejects without increasing starts or changing the ledger, and the genuine acknowledgement subsequently advances A1 once. The output-fault target exits/drains before the actual disposed-writer failure and receives no acknowledgement.

The three actual manifests contain 2,129 mappings each, with states `2121/8`, `2123/6`, and `2129/0` Present/Absent. Their eight runtime-path transitions match the approved windows. Snapshot inventories contain **12 + 14 + 20 = 46** files.

## Completeness and limitations

I read the complete approved design, original normative contract, replacement plan, review brief, progress ledger/rulings, execution report, all twenty designated inputs, and the fixed fixture/association observer. I inspected all six intents/finals/stdout records, retained Prepare/Run tool records, relevant actual evidence branches, manifest transitions, snapshot inventories, and frozen copies containing the findings.

This was a whole-new-chain static review, not merely a last-commit review. No Git commands, imports, tests, project scripts, network calls, writes, new worktrees, or subagents were used.

I did **not** independently recompute every historical pin hash or byte-compare all 46 snapshots. Those full integrity checks remain author-recorded evidence, not a second verification performed here. Large evidence/manifests were inspected selectively rather than certifying every repeated archival field. Supplied baseline/HEAD identities were not independently checked with Git.

## Declined to judge

- Dynamic late write/flush/dispose failures, asynchronous/concurrent cleanup failures, and simultaneous outlet failures: not induced by the approved runs; static first-cause defects are nevertheless reported in I1.
- Dynamic classifier exceptions: the preservation branch was inspected, but no actual classifier exception was demonstrated.
- Host death, machine crash, and guaranteed descendant cleanup after capture-host loss: no such experiment occurred; the parent owns only its direct capture process.
- Hard completion guarantees during indefinitely blocked synchronous filesystem/console operations: the recorded finite process waits do not establish that broader guarantee.
- Hostile same-process reflection, forced module reload, or deliberate source/manifest replacement: the design expressly does not provide a security sandbox.
- Cross-process/global crash-recoverable authorization or restoring Session/Ack from portable files: explicitly outside the design.
- Runtime behavior on other machines, runtimes, identities, or privilege levels: the evidence is restricted to the pinned environment.
- OS-level prohibition of historical scripts and correctness of unrelated historical/product paths: preserving them is not a claim that they are disabled or repaired.
- Independent semantic revalidation of the unchanged seam/codec/inbox implementations: this review assessed their fixed binding and retained output/trace, not a new audit of those components.
- Operator documentation and the final portable handoff: explicitly later mechanical work, not runtime inputs or completed review deliverables.
- Git/remote checkpoint identity and a complete archival byte audit: not independently revalidated under this review’s restrictions.
- Production, formal, whole-C12 acceptance, or closure of historical C1/C2: outside this narrow assessment; R2/Minor2 remain deferred and all five capability flags remain false.

The code-reviewer template determined this report’s structure only; it did not expand the approved review scope.
