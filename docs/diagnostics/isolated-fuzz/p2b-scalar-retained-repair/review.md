# Independent read-only review — complete output

Reviewer: /root/retained_scalar_final_review. One authorized review, no nested
agents or execution. This postcapture review is not a frozen test input.

## Strengths

- The repair is narrowly scoped. Compared with the new-scope baseline snapshot, changes are confined to scalar storage/validation helpers and the fresh-copy guard. The root/common-owner validator, public APIs, protocol encoding, reason ordering, and conservative authority flags remain unchanged.
- For a materialized member table, validation checks exact descriptor classes, case-sensitive names, hidden/reserved/instance flags, ownership key and table identity, internal/constructor-owner consistency, and the fixed SDK pstypenames getter with null setter. Unexpected descriptors—including hidden notes and subclasses—are rejected without reading their Value, MemberType, Copy, or stored getter.
- Reflection arguments deliberately preserve actual PSObject references. Scalar adaptation acquires one shim rather than independently reacquiring string views.
- Fresh-copy validation preserves its caught exception as InnerException at control-frame.ps1:127–128.
- The caller-side bridge is accurately described as caller-side evidence, not a callee-internal identity observation.
- Saved evidence reconciles:
  - All 1,154 frozen pins match when resolved from the review worktree.
  - Baseline, product, frozen manifest, product seal, and saved stream hashes match the review package.
  - RED: exact HiddenDecorationRefusal, seven clean accepts, fourteen visible refusals, fourteen hidden accepts.
  - Focused: seven clean accepts and twenty-eight decoration refusals; 182 malformed-descriptor cases, 45 direct guards, 54 storage-key comparisons.
  - Full: exact 36-line oracle, zero recorded counter total, null error.
  - All three receipts record ordinary pinned PS7 execution, completed streams, no timeout, and empty stderr.

An initial hash reconciliation used the parent directory for relative manifest entries and consequently reported missing paths. Repeating that read-only hash check from the manifest's intended worktree produced zero mismatches; this was not a test rerun.

## Findings

### Critical

None identified.

### Important

**[P1] A null wrapper cache is treated as proof that no decorations exist**

File: control-frame.ps1:68 and :92 in this leaf.

Get-P2BCStorage calculates the SDK storage key but obtains the descriptor table solely from the passed wrapper's _instanceMembers field. Assert-P2BCPlain then immediately accepts when that field is null.

These are different propositions:

1. This particular wrapper has not materialized its local member-table cache.
2. No member table exists for its SDK resurrection key.

The code establishes only the first. For default-storage non-string scalar/byte-array wrappers, another wrapper around the same base can share the resurrection key and its attached metadata. An unmaterialized wrapper can therefore have a null cache while decorations already exist for that key. The current early return skips those decorations entirely.

This is reachable within the intended input model, not merely a forged-metadata scenario: the validator explicitly accepts actual PSObject receivers at line 77. A direct PSObject-valued byte-array argument can take that branch through ConvertFrom-P2BControlFrame at line 339; an exact root note property holding an actual wrapper likewise reaches the scalar validator at line 209. Neither path requires materializing that wrapper's SDK views before the null-cache decision.

The tests do not close this gap. scalar-sdk-fixture.ps1:229–230 generates views on both wrappers before comparing storage, and direct guards generate views before validation at :91. Those tests establish sharing after materialization, not absence of shared metadata before it.

Impact: Decorated input can pass the plain-input guard and be copied into an otherwise valid accepted result, violating the explicit requirement to refuse every extra descriptor. A clean null-cache wrapper must remain acceptable, but null cache alone cannot justify acceptance.

Required fix direction: Distinguish a genuinely absent authoritative table from an uninitialized wrapper cache, using a passive, key-correct inspection that preserves the no-factory/no-key-method-execution and descriptor-passivity requirements. Add coverage for a decorated shared-key base presented through an unmaterialized wrapper, alongside genuinely clean null-cache controls and local/string non-sharing controls. This requires new authorization; no repair or additional execution was performed.

Evidence qualification: This is a source-based correctness finding. I did not execute a new exploit or probe, and the saved runs do not demonstrate this specific history.

### Minor

None separately identified. The coverage omission above is part of the Important finding, not a second finding.

## Behaviors considered but declined to judge

- Observed runtime outcome of the null-cache/shared-table history: Not executed; all three host slots are spent. The finding concerns the unsupported acceptance decision visible in source.
- Callee-internal wrapper identity across public API binding: Caller-side snapshots do not observe that. They establish caller retention only.
- Exhaustive SDK key equivalence: The saved oracle covers 54 specified materialized cases. It does not cover every nested-wrapper, cache-initialization, or storage-transition history.
- Compatibility beyond the eight-wrapper depth bound: The loop fails closed after its finite bound. No evidence establishes acceptance of every deeper legitimate history; no such compatibility claim is made.
- Exhaustive coherent reflection forgeries, concurrent mutation, or same-process isolation: Helpers are conventions, not a security boundary. Structural checks do not establish unforgeable provenance or atomicity.
- Fresh-copy guard failure at runtime and exact resulting exception chain: The source retains the caught exception, but saved guard cases exercise clean copies, not an injected fresh-copy failure.
- Historical implementation correctness: Historical files were hash-only inputs. I did not inspect their source or reopen their stopped scopes.
- Speed, PS5 compatibility, other SDK versions, or live operation: Only finite pinned PS7 execution is evidenced; elapsed times are not performance or portability guarantees.
- Formal P2-B/P2-C/P3/full-C12 acceptance, identity/capability approval, deployment, or merge readiness: Outside this bounded review. Authority flags remain false.
- Independently proving the historical execution environment: Receipt/stream/seal consistency was checked; historical processes were not recreated.

## Recommendation and assessment

Do not hand off this candidate as a verified complete repair. Preserve the sealed candidate and its successful bounded receipts, but record one unresolved Important correctness issue concerning null-cache versus authoritative shared metadata.

The materialized-table repair and saved gates are strong evidence for the tested histories. They do not establish the required refusal behavior for unmaterialized wrappers sharing decorated SDK storage. No merge or broader acceptance is authorized.
