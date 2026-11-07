# Task11 ledger evidence reconciliation — Stage1 complete, runtime BLOCKED

Date: 2026-11-07. Reviewed baseline `9b09e3290ce899db688733fb48931c0d3141211c`,
branch `codex/c12-b01-task9-coordinator`, clean worktree at entry. Origin:
`https://github.com/win202569/Talenro.git`.

User approved the [Stage1 plan](../superpowers/plans/2026-11-07-task11-ledger-evidence-reconciliation.md).
This is a text/hash/Git review, not execution or independent security certification.
No test, helper import, native/Go/Docker workload or system/security query ran.
44/34/6/217 and NOT_ACCEPTED remain unchanged.

## Current source bindings

All paths below are under `docs/diagnostics/isolated-fuzz/`. S1–S7 designate
these exact working-file SHA256 values, not a loaded-runtime attestation.

|Ref|File|SHA256|
|---|---|---|
|S1|ordinary-ledger.cs|6D9E59010E9A301CBF9C21A466A07DB8894382AD7501DB512957862BA77D438C|
|S2|ordinary-ledger-session.cs|06E4E5C164792AFA17A63F34F2C34A523F4F468CE1F36AC87F0B1CF28B9469E5|
|S3|ordinary-ledger-receipt.ps1|D0D751C567F8925ADFD3E86129F05F7CDDCF776AB7C975CD1811EE4889670F9A|
|S4|ordinary-ledger-native-probe.ps1|B93AD786C47180EB22A0E6F7B2F339712E8D6B7FA1D1FA53BC256C5058BD59B3|
|S5|process-native.cs|370B084306025D50F5233743DDCFB5322671C673FF2D75D76A2977F05BDF2B3C|
|S6|run.ps1|CFD2016F15D00BFD6E196169381C1EB05818BF964431B80626E9FA70895A8CC0|
|S7|matrix-check.ps1|6A85689FA60985959574DC435F3C75E7822E5264806ECE7937C9E340974D0EAB|

Git comparison of 4b40f034815863cbf132360b734a8059d149bfda..HEAD shows no
committed delta for these seven files or ordinary-ledger test files. That binds
the source lineage, not the entire old host or captured result. The adapter
report explicitly records S4's same hash. Other reports generally bind commits
and invocation-index hashes, not a portable full source/host/run manifest;
their reported PASS counts remain historical, not fresh verified runs.

## Saved evidence key and supersession

- R1: [implementation](2026-10-17-ordinary-ledger-implementation.md), original
  final checkpoint36367122 and 14 final fake calls/350 assertions.
- R2: [final review](2026-10-17-ordinary-ledger-final-review.md), range89cb1ad..36367122;
  six findings superseded by [review repair](2026-10-17-ordinary-ledger-review-repair.md)
  and [initial snapshot correction](2026-10-17-ordinary-ledger-initial-snapshot-fix.md).
  Snapshot base776f2d3, fixc6d6bf8a; 21 focused assertions per host recorded.
- R3: [probe preparation](2026-10-17-ordinary-ledger-native-probe-preparation.md),
  [review](2026-10-17-ordinary-ledger-native-probe-review.md),
  [repair](2026-10-17-ordinary-ledger-native-probe-repair.md): old636e4890 binding
  and exit-order claims were corrected at a02baace; repair reports47 per host.
- R4: [inbox policy](2026-10-17-ordinary-ledger-inbox-policy.md), base5ea1cef7,
  fixfa3353c7: dedicated held system-shell identity/signature policy replaces
  the old single-link incompatibility.64 per host reported; original pending
  index entry remains disclosed, not rewritten as a clean index.
- R5: [timing coverage](2026-10-17-ordinary-ledger-timing-coverage.md), base3bba6b7c,
  21 per host and two caught mutants. It closes R4's two missing fake timing
  regressions, not real OS scheduling or crash durability.
- R6: [read-only adapters](2026-10-17-ordinary-ledger-readonly-adapters.md),
  checkpoint4b40f034, historical real file/signature reads only. Shell hash,
  identity and source S4 hash recorded; parent wrapper exit1 preserved. Original
  raw stream bytes were not archived; this is not a clean native workload pass.
- R7: [Stage B checklist](2026-10-17-ordinary-ledger-stageb-readiness.md),
  historical checkpoints9e107335/726e51a6. Later
  [staged integrity evidence](2026-11-05-staged-integrity-integration-gap-review.md)
  documents observation44, repaired saved-evidence checks, parent persistence and
  default-off offline hooks. Its original missing-tests/Boolean findings are
  historical, not asserted as current. Latest combined33 matched with reparse
  blocked; wider lifecycle/fault/review gaps remain. This is not a qualified
  ledger-probe outer runner or approval of the new Task11 path.

## Single requirement table

`bound` means the cited source or explicit historical binding matches its stated
scope; it never means present-machine acceptance. `unknown` is lack of usable
current execution binding. Mixed classes/applicability are stated separately.

|ID|Requirement|Current source / hash reference|Saved report and binding|Evidence class; applicability|Remaining gap|Responsible role|
|---|---|---|---|---|---|---|---|
|G1|Acquire once, verify ownership, retain same-run identity|S1 Poll/read/release; S5 native adapter|R1/R2,36367122→c6d6bf8a; AcquisitionOwnsOnce, ForeignReleased, CrossRunIdentity definitions|source-only: bound; historical-fake: unknown runtime binding|No source regression identified; real owned-handle lifecycle for the proposed exact run unobserved. Sampling is not exhaustive tracing.|Ledger maintainer; separately authorized runtime verifier|
|G2|Sticky failures, cap32 refusal, sampled count reconciliation|S1 Poll/CurrentPollQuery; S2 Finish|R2 snapshot correction; Open87Sticky, Cap32, MissedMemberCount, CurrentFailedQueryRecorded definitions|source-only: bound; historical-fake: unknown runtime binding; old pre-fix evidence: changed|Initial snapshot repair is present; do not reopen that old defect. Native count/query behavior still unverified for current packet.|Ledger maintainer/runtime verifier|
|G3|Shared time budget, final reads, exactly-once release, cleanup|S1 Dispose; S2 Finish; S5 ordinary entry/teardown|R1/R2; Shared500ms, DeadlineOverflow, BackwardClock, CloseFailureSticky, SessionFakeCleanup definitions|source-only: bound; historical-fake: unknown runtime binding|500ms shared window and teardown source exist; no current native timeout/cleanup proof. Internal deadline checks cannot alone establish outer preemption.|Runtime/containment maintainer|
|G4|Schema4, diagnostics and legacy isolation|S3 evaluator; S6 Assert-ProcessCompletion; S7 Invoke-MatrixNative|R2 compatibility/mode repair; GenuineLegacy, PolicyOnlyDowngrade, DiagnosticRowRefused, SerializationRefusal definitions|source-only: bound; historical-fake: unknown runtime binding|Public refusal and legacy Run are intentional rollout prohibition, not defects. Task11-specific adapter remains future separately reviewed work.|Task11 rollout maintainer/approver|
|G5|Serialization and strict pure evidence evaluation|S3 Test-OrdinaryLedgerEvidence; S6 receipt fields|R1/R2; NestedRoundtrip, StrictTypes, CompleteButNotAccepted, RootOnlyValid definitions|source-only: bound; historical-fake: unknown runtime binding|Existing fake assertions do not prove execution provenance. EvidenceComplete must never promote Acceptance=false or accept the historical metadata RED.|Evidence-schema maintainer|
|G6|Loaded sources, executable, exact run/arguments binding|S4 initializer, held shell bindings and entry recheck; S6 dependency list|R3→R4→R5; R6 explicit S4 hash equals current; Git source lineage unchanged since4b40f034|source-only: bound; historical-fake: unknown runtime binding; historical-native: bound to R6 historical adapter only, current host unknown|Old binding defects and known missing timing tests repaired. Missing current complete eleven-source/host/executable/run packet; stale host identity cannot be reused. No new packet generated.|Runtime packet preparer and reviewer|
|G7|Raw capture, owned outer containment and durable finalization|S4 consumption/manifest/completed/saved validation; S5 output handling|R3/R5 fake finalization; R6 non-clean wrapper/no raw archive; R7 later partial integration evidence|source-only: bound; historical-fake: unknown current integration; historical-native: bound to historical scope only; missing: unknown ledger-specific outer qualification|Not an absence of all capture tooling: later progress is retained. No demonstrated current ledger-specific end-to-end outer120s capture/cleanup/finalization binding. Disk-hang/power-loss guarantees not established; no new test invented or run.|Outer-runner/evidence owner and independent reviewer|
|G8|Legitimate safety disposition and exact runtime/rollout authority|S4 RuntimeApproved/pin checks are caller gates only; S6 public refusal|R7 historical Defender1116/17273 and1117/17280 association; approved Stage1 plan grants documentation only|external-disposition: missing, unknown; source-only: bound; rollout authority: missing, not-applicable to Stage1|No recorded legitimate owner/vendor decision covering the proposed small experiment and historical finding. False positive and exact attribution unproved. No approved exact Stage2 packet or Stage3 rollout decision.|Legitimate security owner/vendor (person not established); user for subsequent exact scope|

## Finite disposition / execution ledger

Task1 complete: eight rows distinguish source facts, historical fake checks,
historical native adapter observations and unavailable external decisions.
No newly missing fake test is asserted: the two specifically documented timing
gaps were covered by R5. Source-changed historical evidence is superseded rather
than counted as current. G1–G5 cannot close G6–G8.

Task2 disposition: **BLOCKED for Stage2/Stage3**. G8 alone is sufficient;
G6/G7 additionally lack a qualified current exact-input integration. A legitimate
security-owner/vendor disposition must identify decision maker, evidence, scope
and conditions; generic continuation or caller switches cannot supply it. Later
runtime consideration also needs source-bound outer ownership, capture, timeout,
cleanup and finalization evidence. No person is assigned without evidence.

No targeted-verification plan, runnable packet, nonce, additional diagnostic or
repair scope is created. No old allowance is reused. No WPR/registry/admin/cache
detour follows. Stop this route after preserving the report; do not ask for
another generic confirmation as if that resolves the missing external evidence.

Ruling: use this tracked report as the documentation execution ledger and use
read-only document checks rather than generic Bash/TDD helpers — the approved
plan explicitly prohibits test/import/runtime steps. Cost: no new runtime
confidence; that is recorded as unknown rather than silently promoted.

Document checks: eight unique G rows, referenced local documents present, only
the report/plan/status/recovery documentation changed, and whitespace diff check.
Final review and Git preservation are recorded by the checkpoint handoff; no
claim of workload acceptance or main merge is made.

Final independent read-only document review: no Critical/Important/Minor findings;
ready for a documentation checkpoint only. Author disposition of declined areas:
native behavior, security clearance, private raw records, fresh-machine readiness
and runtime acceptance remain unverified/outside this review; cost of promoting
them would be false readiness. Seven hashes were separately recomputed by the
coordinator and matched this report; Git preservation is verified separately at
handoff, not inferred from the review. No second review or runtime action follows.
