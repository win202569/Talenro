# Independent read-only review — cold-cache repair

Reviewer: /root/cold_cache_final_review. One authorized seat, completed 2026-10-18.
The complete returned findings, qualifications and disposition follow.

## Strengths

- The repair directly addresses the prior P1: Get-P2BCStorage consults the authoritative SDK weak table only when the wrapper's local cache is null. A genuine miss remains clean; a found table enters the existing exact descriptor checks. See control-frame.ps1:70–91.
- Binding is appropriately narrow: fixed SDK field name, exact declaring type, static/readonly field, exact constructed ConditionalWeakTable type, and instance TryGetValue(object, rawCollection&) -> bool signature checks. Actual PSObject keys are placed into object[] arguments without a coercing direct CLR key call. See control-frame.ps1:25–26,75–84.
- The added product path contains no cache assignment, table insertion, SDK factory/key-method call, or descriptor Value/MemberType/Copy/getter execution. Found descriptors retain the existing type, name, flags, owner, key, raw-table, internal-owner, constructor-owner, and SDK-code-reference checks. See control-frame.ps1:115–134.
- The change is tightly contained. Comparison against the new packet's baseline-product.ps1 shows only the two metadata fields, null-cache lookup, and comments changed. The text from Copy-P2BCScalar to EOF is ordinal-identical.
- The regression construction is independent of product expectations. Peer A receives views/decorations; B is constructed over the same base and observed using fixed fields, SDK key oracle, and readonly lookup. Preconditions are saved before product invocation and explicitly prove receiver type, cache state, key/table relationship, tag identity/class/flags, and zero callbacks. See cache-fixture.ps1:43–97, receiver-fixture.ps1:49–84.
- The public regression passes the actual wrapper slot to decode, not a reconstructed underlying byte array (cache-fixture.ps1:100). The saved four dirty public subjects retain their wrapper/base/key identity and null cache while returning invalid-input; the clean public cases may materialize during copying, as documented.
- The literal frame oracle is not produced by the candidate encoder: data-helpers.ps1:46 reads frozen FrameHex, and the full suite compares encoding/decoding against independently typed literals.
- Controller gating is adequate for the stated non-adversarial workflow: ordered/preconsumed slots, no replay after an incomplete/unexpected receipt, seal before focused/full, pinned engine/SDK/input inventory, separate concurrent stream drains, bounded execution/exit/drain waits, and termination limited to its owned target process. See capture.ps1:60–99,133–153.

Source references above are in this p2b-cold-cache-repair directory.

## Findings

- Critical: None identified.
- Important: None identified.
- Minor: None identified requiring repair within this bounded contract.

No required read-only review item remains incomplete.

## Evidence independently checked

I analyzed saved text/JSON and calculated hashes only; I did not execute/import/compile any product, tests, fixture, or controller.

- All 1,185 mapped pins currently match.
- Frozen manifest, product seal, baseline snapshot, candidate, all saved streams, and full trace hashes match the stated seals/receipts.
- RED: 84 completed, 48 matched, exactly 36 shared-dirty accepts, with the expected UnmaterializedSharedDecorationRefusal.
- Focused: 84/84 matched, 38 accepts / 46 rejects, all 36 prior misses refused, zero callback counters, all direct cache-reference retention checks true.
- Retained evidence includes 27 proofs, 14 API histories, 35 matrix rows, 45 guards, 54 materialized key controls, and 182 malformed SDK cases.
- Full stdout matches the frozen 36-line oracle exactly; saved trace has zero total callbacks and null error.
- Receipts record three ordinary pinned PS7 hosts, expected exits, no timeout, completed owned exit/drains, and empty stderr.

## Behaviors considered but declined to judge

1. Fresh runtime reproduction: prohibited by this review authorization; conclusions use saved execution evidence plus independent static/hash/JSON checks.
2. Other SDK/PowerShell versions, PS5, platforms, architectures, or elevated hosts: only the pinned ordinary PS7.6.5/Core/x64 configuration is evidenced.
3. Arbitrary private-reflection mutation, forged helper metadata/functions, or replacement/corruption of SDK internals: explicitly outside the structural-compatibility contract.
4. Concurrent mutation, TOCTOU resistance, process isolation, and adversarial anti-replay: explicitly outside scope. The controller prevents ordinary workflow reuse; it is not a tamper-proof authority ledger.
5. Unforgeable SDK factory provenance: explicitly excluded. The implementation validates structure and references, not historical provenance.
6. New dynamic proof of nested-wrapper or empty-existing-table cold cases: no dedicated such cases occur in the 84-case cold matrix, and further execution is forbidden. Static review finds the inherited nested-key calculation unchanged; an exact empty raw table naturally traverses zero descriptors without cache assignment/insertion. I do not claim independently executed coverage for those combinations.
7. Historical source correctness or stopped investigations: historical material was hash-checked only; those scopes were not reopened.
8. Performance, arbitrary-depth/resource-exhaustion guarantees, live actions, identity/capability/security/runtime approval, formal P2-B/PS5/P2-C/P3/full-C12 acceptance, or main-branch integration: not established by this repair/review and not authorized here.

These are explicit boundaries on conclusions, not concealed product findings.

## Recommendations

Preserve this review, exact receipts/seals, and the bounded exclusions in the durable report. Do not describe the saved suite as general isolation/security assurance or broaden acceptance. No repair or additional run is recommended under the exhausted allowance.

## Assessment

The prior P1 is resolved within the stated bounded contract. The implementation now distinguishes a null wrapper cache from absent authoritative shared metadata, preserves key isolation and inherited validation, and the saved regression changes exactly the 36 intended baseline misses to refusals without observed control regressions.

This is a read-only review conclusion—not formal acceptance, main authorization, or permission for further execution.
