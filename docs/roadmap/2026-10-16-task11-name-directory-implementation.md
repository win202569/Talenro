# Task11 name-directory implementation — scoped evidence

Status: UNACCEPTED DEVELOPMENT CHECKPOINT. Implementation at90dc0e2a/990d6683
has verified RED/GREEN and full final audits; full contracts regression55
top-level/423 subtest PASS,0 SKIP/FAIL. Final independent review found Important1:
test-only negative-vector null handling remains OPEN. Existing test PASS does
not close this missing-case gap. No fix/additional workload has run. Original
five-build/four-runtime budget consumed; await approval of the bounded fix
addendum below. Not complete Task11/name-directory/C12 acceptance.

## Delivered scope and independent oracle

The [approved design](../superpowers/specs/2026-10-16-authority-v7-name-directory-design.md)
and [implementation plan](../superpowers/plans/2026-10-16-authority-v7-name-directory.md)
limit this slice to146 exact schema/object labels, case-sensitive lookup, a fresh
byte-sorted list and one finite unknown-name sentinel. Membership grants no
authentication, authorization, evidence eligibility or codec support.

The registry literal was extracted from the original v7 §4.1 table; the golden
literal was separately extracted from the146-row source index. Both were checked
against the original complete pair set before production implementation. All146
ordered pairs agree, with146 unique schemas/objects and none of the three excluded
historical objects. No expectation was generated from production lookup/list.
The production literal table was subsequently compared to the independent golden
set; all146 ordered pairs agree. Fixtures have18862/60364 bytes (UTF8/LF), each
below64KiB; golden includes881 literal rejected inputs. Runtime code never reads
these files or Markdown/config/environment.

Necessary implementation paths:

- internal/nodecontrol/contracts/schema_names.go
- internal/nodecontrol/contracts/schema_names_test.go
- testdata/c12/authority-v7/schema-name-registry.v1.json
- testdata/c12/authority-v7/schema-name-registry-golden.v1.json

No original metadata/canonical-envelope fixture, historical support, policy,
private signer/helper, SQL/generated schema, approved spec-set or consumer wiring
was changed. No dependencies, tools or live services were installed/downloaded.

## Observed ordinary verification

Owned local identity: name-directory-implementation-001 under the existing
isolated-fuzz diagnostic scope. These are new contracts-package attempts, not
replays/retargeting of consumed leaf001 or a matrix. Native/PS5.1 64-bit host;
default-vet builds, offline environment, unchanged ordinary collector and schema3
completion/query/type guards;500ms observation/5000ms cleanup budgets retained.
Each native build limit600s; original test timeout120s/native runtime150s.

| Phase | Build exit / seconds | Runtime exit / seconds | Scoped result |
| --- | --- | --- | --- |
| fixture-red | 0 / 97.345 | 1 / 1.688 | Verified expected missing-fixture assertion |
| fixture-green | 0 / 61.639 | 0 / 1.338 | Verified;2 top-level/92 subtest PASS,0 SKIP/FAIL |
| directory-red | 1 / 67.767 | not run | Verified expected undefined directory symbols |
| directory-green | 0 / 81.700 | 0 / 1.354 | Verified;4 named tests PASS,0 SKIP/FAIL |
| contracts-final | 0 / 65.646 | 0 / 1.408 | Verified;55 top-level/423 subtest PASS,0 SKIP/FAIL |

Verified rows require full pre/post copied-SDK/reference and product audits,
fixture/source/tool/argument/cwd bindings, artifact hashes and no ordinary-guard
failure. The original71 helpers and four input pins matched the approved plan.
RED means the documented behavior/undefined symbols, not native87, timeout,
forced cleanup or uncertain completion. No retry/extra native workload has run.
Exactly five native builds/four test-binary runtimes ran, each once. Full-package
ordinary execution includes existing fuzz seed cases, not a fuzz-engine campaign.
Native phase times above exclude full pre/post audits and must not be presented
as total elapsed time. Raw receipts/streams/manifests, wrapper, SDK/modules,
caches/tmp and binaries remain ignored locally, not part of a Git recovery claim.

No native launch/persistence exception occurred in these attempts. The existing
shared recorded-process helper's simultaneous native-error/error-write-failure
path was neither changed nor failure-injected; this report does not certify its
primary/secondary exception preservation in that hypothetical path. The ignored
operational wrapper is not a new shipped/reusable diagnostic controller, and
these once-only phase identities must not be replayed.

## Execution rulings and costs

1. PowerShell/apply_patch equivalents replace Bash-only skill bookkeeping.
   Plan identity, BASE and observed results are preserved; task-done does not
   replay workloads. Cost if wrong: bookkeeping omissions, mitigated by this
   tracked report, explicit source history and original local evidence.
2. The approved offline five-build/four-runtime budget and contracts-package
   scope override generic skill dependency-download/whole-repository setup runs.
   Cost: other packages/environments are unmeasured, not claimed passing.
3. After fixture-red, each remaining phase uses a fresh child of approved
   R/go-cache, rather than a shared warm cache that could omit original compiler
   tokens for unchanged final source. Same20 environment keys, all boundaries,
   ordinary guards and native budgets; no extra builds or weaker predicates.
   Cost: additional cold-build time/disk within the same600s limit. First-phase
   evidence/environment was not rewritten.
4. Final independent review covered this entire approved slice from6c5ec277,
   not re-review older unrelated C12 development commits as if part of this task.
   Cost: earlier branch work is not newly certified; broader gates remain open.

5. Final: deferred metadata, codecs and authorization/evidence behavior remain
   excluded exactly as the approved name-only design requires; name lookup is
   still not authorization. Cost: these capabilities are not delivered/certified.
6. Final: older C12 branch changes/security/production/physical gates are not
   recertified by this review; preserve their independent open status. Cost:
   earlier issues may remain, and no broader/main acceptance is implied.

## Final independent review — Important1 OPEN

One fresh read-only reviewer, /root/name_directory_final_review, inspected the
whole slice6c5ec277..990d6683, original146-pair sources, current result/rulings
and saved evidence. No edits/workloads were performed by the reviewer. It
independently confirmed all146 pairs across five representations, ordering,
exclusions, exact lookup, copied ownership, four declarations, no consumer wiring
and observed55/423 PASS counts. Critical:0; Important:1; Minor:0.

Important1: nameGoldenFixture.RejectedSchemas is []string. Replacing its first
valid empty string with JSON null loses the distinction during typed decoding.
strictjson's primitive-token pass-through plus encoding/json's documented-in-code
primitive-null ignore leaves the element as ""; membership/order checks and the
inventory's expected empty string consequently still accept the malformed vector.
The committed fixture itself contains a genuine string, not null; production
lookup is not the defective component. The missing element-level negative case
prevents claiming the strict fixture contract complete.

Executor independently traced schema_names_test.go's wire type/decode/validation,
strictjson/decode.go walkValue, and copied SDK encoding/json/decode.go null branch
(primitive/string null is ignored, pointers can preserve nil). Finding remains
Important. This is source confirmation, not an extra runtime reproduction.
Do not alter shared strictjson or remove the legitimate empty rejection string.

The [bounded fix addendum](../superpowers/plans/2026-10-16-authority-v7-name-directory-null-vector-fix.md)
proposes test-only nil-preserving decoding plus first/last null and other typed
element mutations, then at most three additional builds/three runtimes for RED,
fixture GREEN and whole-package GREEN. Same Native/inline method, one fix pass,
no second reviewer. Await user approval before modifying tests/decoder or running
those workloads. The old five phases are neither retried nor invalidated as
observations; their coverage was insufficient for the new finding.

Reviewer exclusions were deferred metadata/codec/authority behavior and older
C12/security/production/physical acceptance. Rulings5–6 above explicitly preserve
those approved exclusions and their costs rather than treating them as passed.

## Remaining gates and recovery

Metadata93 author source-backed classifications/53 unclosed reviews remain
postponed, not independently certified or filled with defaults. Full Task11,
canonical-set/deployment compatibility, message exact-cover, physical DB/PITR,
historical matrix003/error87/timeouts/fuzz, security/production/main remain open.
No merge/PR/main update is authorized by scoped directory tests.

Resume from the latest codex/c12-b01-task9-coordinator development tip, not main.
Read current status and this result before acting on historical plan checkboxes.
Do not replay attempted phase identities, manually run their binaries or treat
Git cloning as restoration of private toolchain/cache/raw evidence/live data.
Current resume point is Important1/addendum approval, not another review or old
phase replay. Both proposed fix root/control paths were absent when written;
recheck them only after approval. If verification stops, preserve its exact stage
and request the bounded next decision; do not silently reset pins, retry or relax
safety checks. All necessary source/tests/fixtures and this open finding are
included in the development checkpoint; final handoff verifies actual remote SHA.
