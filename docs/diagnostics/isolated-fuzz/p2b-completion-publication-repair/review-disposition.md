# Coordinator disposition — C1/C2 block acceptance

Complete sole independent report preserved in review.md. Author verified C1/C2
against frozen code. No extra execution, repair, retry or second review; all
4hosts+4children+1review consumed. No post-review frozen input/old raw changes.

## Findings

- B1 original mechanism narrowly resolved: failed State refuses later normal
  integration publication; healthy publication rereads existing ledger.
- B2 original mechanism narrowly resolved: output/render/flush occur before
  authoritative success save; ordinary output failure leaves incomplete ledger.
- C1 retained Important/unresolved: publication:15 passes EXPECTED_* rows to
  direct WriteAllText at bookkeeping:17. A locally detected late write failure
  can leave valid successful bytes while only live State becomes failed. Fresh
  capture:184–186 trusts persisted outcomes without a positive independent
  completion acknowledgement. No dynamic occurrence is claimed. Manual-stop/
  durability disclaimers do not repair that continuation gate.
- C2 retained Important/unresolved: post-start capture:211 can fail writing
  hosts.json, then cleanup reaches unconditional Complete at:228; bookkeeping:30
  can write the same readable failed sink again. First cause and cleanup remain
  protected, but no-retry/preserve-evidence contract is violated. No dynamic
  occurrence is claimed; helper's earlier expected receipt faults are distinct.
- Status REVIEW_BLOCKED_C1_C2_UNFINISHED. Narrow B1/B2 evidence valid, complete
  failure-closed publication/controller contract NOT accepted. R2/Minor2 deferred.

## All declined-to-judge items — rulings with effects/costs

Numbers exactly match review.md. These boundaries do not waive C1/C2.

1. Production provenance: offline fixtures only. Cost if wrong: untrusted
   traffic promoted to authority; no live consumer acceptance.
2. Stage legality/state ownership: Hello/Bound/Ready not implemented here.
   Cost: invalid real transitions; separate implementation required.
3. Capabilities/P2-C/clearance/permission: false flags remain false.
   Cost: unauthorized live actions; none allowed.
4. Tickets/double Commit/atomicity: no implemented release mechanism.
   Cost: duplicate release/races; no capability claim.
5. Stop/ACK/emptiness/sealing: owned fixture exit is not resource proof.
   Cost: false cleanup acceptance; no real release claim.
6. Watchdogs/deadlines: bounded waits are not hard-real-time guarantees.
   Cost: missed real deadlines; require separate runtime evidence.
7. Crash/power loss/transactions: no durability guarantee. Ordinary detected
   I/O remains in scope, C1 blocks. Cost: false safe-resume; no exclusion waiver.
8. Every sink unavailable: best-effort only, unknown/INCOMPLETE.
   Cost: unavailable diagnostics; never invent recovered success.
9. Real-finalizer injections: shared consumers tested, call sites static.
   Cost: uncovered caller paths; C1/C2 retained, no inflated dynamic claim.
10. Actual-parent output injection: closed writer observed only in unit host.
    Cost: overstated actual-parent coverage; source assessment distinguished.
11. Raw write/flush/final-save/fallback faults: first closed-output write only
    dynamically exercised. Cost: unobserved later faults; C1 not waived.
12. Preparation/preconsume faults: source checked, not dynamically induced.
    Cost: unverified error paths; successful preparation not universal proof.
13. Other serialization/index/path variants: bounded catches inspected.
    Cost: unknown variants; no complete input/failure acceptance.
14. Unknown Process/API/OS errors: static guards, not induced.
    Cost: OS-specific failures unobserved; no universal runtime claim.
15. Other platforms/versions/machines: pinned environment only.
    Cost: incompatible semantics; clone is not new-machine acceptance.
16. Concurrency/tamper/replay: serial private model. C1 ordinary errors still
    in scope. Cost: unsafe continuation; explicit gate blocker retained.
17. Symlinks/path aliases: lexical paths not hardened filesystem sandbox.
    Cost: hostile alias escape; no adversarial path-security guarantee.
18. Public hostile callers/state/getters: private ordinary controller data.
    Cost: arbitrary mutation effects; do not expose as hardened public API.
19. Fatal termination/OOM: no survival/reporting guarantee.
    Cost: lost evidence; unknown, never success.
20. Unowned/reused/disposed Process objects: exact retained owned objects only.
    Cost: wrong-process cleanup; no enumeration or tree-kill authority.
21. Historical suites/helper: identities/relevant boundaries inspected, not
    full rerun. Cost: inherited unknown defects; old results and quotas retained.
22. R2 collision: individual prestart description remains missing.
    Cost: diagnostic gap; deferred, not reconstructed.
23. Minor2 scalar ETS gap: unchanged/outside approval.
    Cost: incomplete metadata-isolation evidence; deferred, not claimed repaired.
24. CLR identity reconstruction: raw relationship assertions only.
    Cost: false reference recovery; serialization never recreates old references.
25. Chronology/current liveness: saved evidence, not live retrospective tracing.
    Cost: overclaimed independent observation; limit explicit.
26. Singleton JSON object: readers wrap @(...), private compatibility retained.
    Cost: future strict-array consumer mismatch; none introduced.
27. PASS/rendered EXPECTED/subordinate FINAL_EXACT: observations, not full
    authoritative completion. Cost: false next-slot eligibility; C1 retained.
28. Healthy reread/multirow concurrent changes: one ordinary row observed,
    concurrency excluded. Cost: overstated multirow/conflict protection; not claimed.
29. FinalLedgerWritten: local observation, not acceptance authority.
    Cost: treating data as permission; no capability assigned.
30. Scope labels: shared algorithm, not multiplied lifecycle fault injection.
    Cost: overstated coverage; dynamic limit retained.
31. Recovery/reconstruction/replay: no automatic rebuilding or consumed-slot
    reruns. Cost: fabricated evidence/extra execution; prohibited.
32. Remote SHA: author separately verifies successful push/equality.
    Cost: no remote recovery if mismatch; never claim equality before checking.
33. Later handoff outputs: author exact-byte/text/hash checks only, not an
    independent reviewed set. Cost: export mistakes; explicit comparison required.
34. Formal/main/live/full-C12: not accepted or authorized.
    Cost: premature deployment; no merge/release/live action.

## Next boundary

Preserve as unfinished development progress, with no more targets or automatic
fixes. Next work needs separately approved design for the whole completion/
continuation boundary and failed-sink lifecycle, including positive acknowledged
completion and no same-sink retry after failure. Reordering another isolated
save is insufficient for C1. No new design/implementation budget follows from
this report; old records and exhausted/frozen allowances remain unchanged.
