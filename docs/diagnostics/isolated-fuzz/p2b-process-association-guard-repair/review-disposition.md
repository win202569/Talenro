# Association guard repair — coordinator review disposition

Status: REVIEW_BLOCKED_R1_UNFINISHED. The unique fresh read-only review is saved
in full in review.md. Author checked R1 at containment.tests.ps1:119 and
capture.ps1:151/216, and R2 at containment.tests.ps1:114/194. No code, frozen
input, oracle, raw receipt or saved result was corrected after review.

## Findings

- R1 Important retained: final ledger reads/writes escape secondary-error handling
  and can mask an earlier failure. Resource cleanup precedes those writes and
  saved runs stay valid, but first-failure preservation is incomplete. Blocks
  completion/acceptance; no automatic fix, test, rerun or second review.
- R2 Minor deferred: PreStart/prestart case-insensitive key collision loses the
  individual pre-start observer exception description. Saved trace still checks
  real error presence/Unassociated; the separate unstarted chain remains. Do not
  reconstruct missing details or weaken this qualification.
- Original I1 post-start cleanup and association repair have limited observed
  supporting evidence; this is not acceptance of the entire controller contract.
  No formal/main/live/fullC12 approval is inferred.

## Every declined-to-judge item dispositioned

Numbers correspond exactly to review.md; no exclusion waives a requirement.

1. Production handshake/provenance: deferred to actual trusted-consumer design/
   implementation; offline fixtures give no delivery capability.
2. Hello/Bound/Ready/state ownership: still unimplemented, not accepted by seam.
3. Capabilities/P2-C/safety/runtime permission: prerequisites remain blocked;
   no query or rights were supplied by this execution.
4. Tickets/Arm/Commit/atomicity: no new implementation or positive acceptance.
5. Native/crash recovery: untested and unapproved, not inferred from managed exits.
6. Real clocks/watchdogs/deadlines/hard realtime: saved bounded waits are not proof;
   unknown/overrun remains unknown and no new clock authority is supplied.
7. Stop/ACK/emptiness/sealing/release: fixture exits do not stand in for these.
8. Unknown Process/API/exception paths: static-only limitation retained; no dynamic
   injection claim, and R1 remains a blocking finding rather than a waived path.
9. Other environments: limit evidence to pinned ordinary Windows PS7.6.5 x64.
10. Concurrency/adversarial ledger/replay: not a supported security claim; maintain
    serial private workflow and no automatic replay on recovery.
11. Disposed/reused/public objects: keep new-owned-undisposed/private contract;
    no arbitrary-caller or lifecycle-reuse guarantee.
12. History/Minor2: historical hashes checked, no old suite re-review or execution;
    original limited approvals and scalar ETS deferral unchanged.
13. CLR identities: store raw relationship observations only, never claim restored
    object references from JSON or Base64.
14. Live state/chronology: saved source/receipts only, no after-the-fact process
    queries or reconstruction of unrecorded execution observations.
15. Remote equality: parent performs ordinary authorized final push and compares
    exact local/remote SHA; that check is separate from independent code review.
16. Final portable/recovery outputs: author-generated and author-verified against
    immutable raw after review; explicitly not independently reviewed. No second
    reviewer seat or implied review coverage for these later outputs.

## Budget and next gate

Spent3/3hosts+4/4children+1/1review; zero remaining. Old containment and causality
reverification remaining allowances stay FROZEN. Save necessary files only on
existing development branch; preserve ignored raw. No main merge or deletion.
Next needs a separately approved narrow R1 repair plan and explicit new budget;
R2 may be proposed there but is not silently fixed now. No acceptance waiver.

## Execution rulings retained

- PowerShell/manual ledger instead of Bash wrappers; no extra baseline, suite,
  task-done rerun, fix pass or raw deletion. Cost: missing wrapper automation,
  not broader execution authority.
- Prepare -Slot green implements observer-only refreeze without a new host.
  Cost: preparation can fail closed, not permission to repeat a consumed slot.
