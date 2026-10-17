# Native probe Stage A — independent review blocks completion

Latest:[approved covering repair verified in fake scope](2026-10-17-ordinary-ledger-native-probe-repair.md).
Binding/test-protection fixes pass47 per host and final350;child observation
narrowed with user approval. Stage A36/36 used,Stage B unapproved. Findings below
are retained original evidence,not a second review of the repair.

User decision after2feca69b:approved one covering repair pass with18 extra fake
calls,fixing bindings/test protection and narrowing the child observation to
same-run owned-handle natural completion. No OS exit-ordering claim,Stage B or
new reviewer authorized. Repair execution will be recorded separately;original
findings/range and18-call records below remain historical evidence.

Date:2026-10-17. Range6a9ade2e5fbb961e826233430df9485107f40f3c..
636e489092051d9f8523215da55cce978adbc79e;two added files. One fresh-context
independent reviewer,readonly;no edits,index/branch mutations,tests,native calls,
network or further delegation. Author source-checked all findings after review.
Verdict:Stage A unfinished;Stage B not ready and not approved.18/18 fake calls
consumed;no repair or verification after review.39 checks per host/final350 are
valid scoped observations,not proof that the helper's binding requirements hold.

## Findings and author disposition

No Critical findings. Three Important findings accepted:

1. **Loaded-source identity not bound to disk hashes.** Helper line2 imports
   run.ps1 before plan preparation;lines31–34 hash current files later,87–89
   reconstruct from disk. Editing a dependency after import can produce a plan
   bound to new files while old functions/compiled types execute. run.ps1 also
   reuses preloaded IsolatedFuzzProcess. A fresh runtime host alone is not proof
   unless loading identity is verified. Bind an import-time verified snapshot,
   reject unverified preloaded types and refuse drift. Do not claim current
   SHA256 identifies already-loaded implementation without that evidence.
2. **Executable not held invariant through launch.** Helper lines197–210 hash
   the executable during plan assertion,then perform other filesystem/saved-root
   work before launching by path. No deny-write/delete binding spans that window.
   A normal system update or replacement could change bytes before entry;later
   saved rejection would not undo the consumed call. Preserve a verified readonly
   executable binding through launch and apply matching load-snapshot discipline.
3. **Controlled child ordering claim is stronger than evidence.** Lines138–140
   combine parent exit-intent/child markers,running admission and final exited
   read. The child may exit before the parent actually terminates. The fake
   likewise sets Exited before Finish. Natural completion can be valid,while
   the proposal's stronger outliving-root scenario is not established. Either
   add suitable OS-backed post-root liveness evidence or explicitly narrow the
   proposed observation with user agreement. Do not infer OS ordering from UTC
   intent markers. Author recommends narrowing to same-run owned-handle natural
   completion instead of expanding native instrumentation in this preparation.

One Minor finding,currently deferred:

4. **Fake boundary permissive and installed after import.** Tests line4 imports
   helper before replacement;lines27–30 normally return fake success rather than
   fail closed when unarmed. Current helper has no import-time native invocation,
   and no real boundary was called in these runs. Future regressions are not
   guarded sufficiently. Prefer explicitly armed fail-closed fake and import-time
   protection when separately authorized. Deferred is not a claim of compliance
   with that proposal requirement;include it in the next user-selected scope.

Source paths:docs/diagnostics/isolated-fuzz/ordinary-ledger-native-probe.ps1 and
ordinary-ledger-native-probe.tests.ps1. The reviewed source is unchanged.

## Reviewer strengths

Fixed identities/create-new consumption,exact argument regeneration,plan pins,
no retry and child-after-completed-root preserve scope. Saved validation checks
manifests,recomputes evidence and refuses ordinary failed/provisional results.
Schema4 acceptance refusal remains unchanged. UTF16LE encoded command bodies and
ASCII scripts are appropriate. Test mocks are disclosed rather than represented
as actual native behavior. These strengths do not close the findings above.

## Declined-to-judge rulings

1. Actual native scheduling/cleanup/file-link behavior:unexecuted because fake
   boundaries replace them. Remain blocked,not inferred from mock results;cost
   if wrong would be unwarranted OS/cleanup confidence.
2. Current security/signature/host freshness,inbox-module environment and outer
   120s enforcement:Stage B prerequisites,not demonstrated by helper. No current
   clearance or runtime authority;cost if missed is unsafe/unbounded execution.
3. Broader unchanged evaluator/serializer:only relevant semantics inspected,
   not a new full audit. Existing evidence retained,not certification;cost if
   wrong is an inherited validation gap. No guard is relaxed.
4. Hostile caller redefining functions,direct native-wrapper invocation or
   rewriting all artifacts:local caller gates are not a security boundary or
   cryptographic attestation. Do not expose as untrusted API;cost if misused
   would be false authorization/authenticity assumptions.
5. Concurrent saved-artifact mutation:hashing and parsing are separate reads,
   not an atomic snapshot. Set aside as requiring concurrent mutation rather
   than ordinary provisional-failure promotion. Author retains that limited
   threat model,not an atomicity guarantee;cost is audit inconsistency under a
   concurrent writer. Stage B must not assume adversarial-write protection.
6. Power-loss durability:successful close/completion markers are not fsync or
   crash-test evidence. Recovery must revalidate and refuse missing/corrupt
   artifacts;cost if overstated is losing evidence after power failure.
7. Supplied pass counts:reviewer did not rerun. Author read all18 captured
   outputs/index and verified39 per GREEN host/final350/empty successful stderr;
   no independent test certification claimed. Additional tests unauthorized.

## Proposed next bounded decision

Recommended scope:fix findings1/2;obtain agreement to narrow the child observation
for3;explicitly include the deferred test protection improvement4. Extend the
single new probe suite with source-load drift/preloaded rejection,executable
binding and fail-closed seam regressions. Proposed18 extra fake calls:RED/GREEN
on both hosts (4),then original seven-suite final allowlist on both hosts (14).
Stage A total36,not a replay or extension of older47 review-repair calls.
Exact fixtures/expected RED and locking/load interfaces must be recorded before
running. No new reviewer loop or Stage B/native execution implied. This scope,
claim narrowing and budget are proposals only,not yet approved.

Preserve current source/tests,raw logs,index and this disposition on the existing
development branch. No main merge,workflow replay,security-setting change,Go/
Docker workload,metadata continuation or acceptance activation.
