# Coordinator disposition — acceptance blocked

Complete sole fresh review is preserved in review.md. Author checked both
Important findings against the frozen source; neither was dynamically induced.
No frozen code, oracle or original evidence was changed after review. No
automatic implementation, target, retry or second review is authorized.

- B1 remains Important: capture:156 ignores failed completion;:172 saves the
  in-memory child rows before the failure gate at:174. This can overwrite a
  malformed ledger or retry a failed sink. First cause retained/no success does
  not excuse destruction of diagnostic evidence. Unresolved, blocks acceptance.
- B2 remains Important: capture:287–290 persists EXPECTED_* before final output;
  :295 catches output failure without invalidating that marker. The next run's
  :182–184 gates read it in a new State. Unresolved, blocks acceptance. No claim
  that B1/B2 occurred in the saved successful observations.
- R1 exception masking is narrowly repaired at the three shared boundaries;
  the complete failure-closed contract is NOT accepted. Status:
  REVIEW_BLOCKED_B1_B2_UNFINISHED. R2 and Minor2 remain explicitly deferred.

## Every declined-to-judge item — coordinator rulings

Numbers match review.md exactly. Each states the supported behavior and cost
if the boundary is misunderstood; none waives B1/B2 or the approved contract.

1. Production provenance: offline functions only; no trusted consumer claim.
   Cost if wrong: untrusted frames could be promoted to authority; prohibit use.
2. Stage legality/state ownership: no Hello/Bound/Ready implementation here.
   Cost: invalid live state transitions; require separate implementation.
3. Capabilities/P2-C/clearance/runtime: false flags remain false, no readiness.
   Cost: unauthorized live action; none permitted.
4. Tickets/double commit/atomicity: not implemented or validated.
   Cost: duplicate release/races; no live ticket use.
5. Stop/ACK/emptiness/sealing: fixture exit proves only that fixture observation.
   Cost: false cleanup proof; no resource-release acceptance.
6. Watchdogs/deadlines: bounded waits are not hard-real-time guarantees.
   Cost: missed real deadlines; separate runtime design required.
7. Crash/power-loss/durable transactions: ordinary writes are not crash atomic.
   Cost: partial records; no durability claim or recovery reconstruction.
8. All sinks failed: fallback best-effort, unknown/INCOMPLETE, not persistence.
   Cost: evidence unavailable; stop rather than claim recovered success.
9. Actual-finalizer injection: shared Save vectors only; wires statically read.
   Cost: unobserved call-site faults; B1/B2 remain blockers, no broad coverage.
10. Preconsume/output/fallback failure: source coverage only, not dynamic proof.
    Cost: missed continuation defects; B1/B2 retained, no extra probe authorized.
11. Other conversion/index/serialization failures: catches inspected, vectors
    limited. Cost: unknown behavior beyond observations; no universal guarantee.
12. Unknown Process/API/OS errors: static guards only, no induced evidence.
    Cost: unobserved OS failures; no all-failure runtime acceptance.
13. Other platforms/versions: pinned Windows/PS only.
    Cost: incompatible semantics elsewhere; cloning is not cross-host acceptance.
14. Concurrency/tampering/replay: serial private model only. Missing/bad ledger
    refusal remains mandatory. Cost: unsafe resume; B1 explicitly not waived.
15. Symlinks/path aliases: lexical boundary not a filesystem security sandbox.
    Cost: adversarial alias escape; no hostile filesystem claim.
16. Hostile public state/getters/exceptions: private ordinary State only.
    Cost: arbitrary caller behavior; do not expose as hardened public API.
17. Fatal shutdown/OOM: no survival/reporting guarantee.
    Cost: no evidence; remain unknown, not successful.
18. Unowned/reused/disposed process trees: held exact owned objects only.
    Cost: wrong process cleanup; no enumeration or broader kill authority.
19. Historical suites/helper: pinned and necessary source read, not full rerun.
    Cost: inherited undiscovered defects; preserve prior results/deferrals.
20. R2 collision: individual prestart exception detail remains lost; trace and
    separate unstarted chain preserved. Cost: diagnostic gap; explicitly deferred.
21. Minor2 ETS assertion gap: unchanged, not newly proved or repaired.
    Cost: incomplete scalar-metadata evidence; explicitly deferred.
22. CLR identity: raw JSON preserves assertions, cannot restore live references.
    Cost: false identity reproduction; no such recovery claim.
23. Chronology/liveness: saved observations only, not independent live tracing.
    Cost: overclaimed retrospective verification; receipts retain that limit.
24. Singleton object JSON: readers deliberately wrap @(...); private interface
    is compatible. Cost: future strict-array reader mismatch; none added here.
25. Scope labels: one shared algorithm, three named callers; not three injected
    lifecycle faults. Cost: inflated coverage; documentation explicitly limits it.
26. PASS observations: not authoritative completion. B2's persisted gate is
    stronger and remains blocked. Cost: false next-slot eligibility; no waiver.
27. Remote SHA: author separately verifies push equality; reviewer used no network.
    Cost: remote recovery absent if mismatch; never claim push before equality.
28. Handoff outputs: author verifies exact raw strings/bytes/hashes only; not
    independently reviewed. Cost: export errors; explicit byte comparisons required.
29. Formal/main/live/full-C12: none authorized or accepted.
    Cost: premature unsafe deployment; no merge/release or live action.

## Final boundary

4/4hosts+4/4children+1/1review consumed,0remaining. Old frozen budgets and all
old failures unchanged. Preserve this as an unfinished development checkpoint.
Next requires separately approved narrow B1/B2 design and finite validation;
this handoff does not itself approve more implementation, runs or review.
