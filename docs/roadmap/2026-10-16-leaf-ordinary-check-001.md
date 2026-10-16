# Ordinary check001 — real single-package verification completed

## Latest outcome after renewed approval

After stopped-preparation checkpoint62395f01cbc48a4c46254c353412a1c67305ddd4,
user explicitly approved fixing the executor command, syntax-checking first,
and retaining the still-absent001 identity for the original single-run scope.
Exactly one Invoke-LeafOrdinaryCheck invocation then completed in a fresh64-bit
Windows PowerShell5.1.26100.9711 host. No controller/native retry occurred.
The previous parser failure remains recorded below;it was not a failed Go test.

**Scoped outcome:** PackagePassed=true from valid completed.json and its
15-file manifest;Acceptance=false. Default-vet build and original tests both
exit0,ordinary schema3 completion/query guards pass. This closes only this
single package's bounded ordinary evidence gap,not whole-C12/main acceptance.

### Measured phases

| Phase | Seconds | Meaning |
| --- | ---: | --- |
| Controller prepare |322.9042261|Full copied SDK/reference and product pre-audits,private cold directories |
| Native build receipt interval |73.9267187|One unchanged default-vet Go test -c build,exit0 |
| Build controller phase |74.7119646|Includes recording around native call;not an extra build |
| Native runtime receipt interval |7.2873656|One new original test binary execution,exit0 |
| Runtime controller phase |7.4462490|Includes recording around native call;not another test |
| Controller post-audit |45.4406632|Independent eligible product/SDK/artifact checks |
| Whole outer host |458.6760542|Includes host/preflight,phase gaps,evidence/finalization;summing native times alone is not total time |

Controller entered2026-10-16T18:11:28.2496950Z;outer host finished
2026-10-16T18:19:04.6385426Z. Native build18:16:52.6316271Z–18:18:06.5583458Z;
native runtime18:18:08.6802429Z–18:18:15.9676085Z. Durations above are receipt
wall-clock intervals/controller stopwatch values,not CPU measurements.

### Verified facts and limits

- Before either workload:input4 hashes and source71 aggregate matched the plan;
 13 helper pins,8 build/4 runtime arguments,20 environment fields,600s/150s
 budgets and typed policy flags matched. Full wrapper parsed without errors in
 PS7 and PS5.1 before invocation. No source/product/helper edits were made.
- One build and one runtime attempt,each durably recorded before its boundary.
 Default vet remained enabled;no custom analyzers,diagnostic modes,fuzz/race,
 cache warm-up,extra Go commands or historical binary reuse.
- Both receipts:ShutdownSchema3,RootExited/NaturalShutdownVerified/
 ShutdownEvidenceComplete/CleanupVerified true;TimedOut/TerminationRequested/
 MemberEvidenceTruncated false;ActiveAfterCleanup0/DescendantsRemaining0;
 initial accounting query succeeded with Error0,plus initial member-query
 evidence. LiveSampling/ShutdownDiagnosticOnly/LifetimeDiagnosticOnly false.
 Accounting reports546 build processes and2 runtime processes;these totals do
 not mean every process or analyzer identity was captured. No historical
 error87 root cause is established by this successful run.
- Full11536-file copied SDK/reference checks passed before/after;actual product
 inventories701 before and701 after matched. Eligible final audits all passed,
 no primary failure or audit errors. Source71 and four input pins matched again
 after completion. No full dependency-cache inventory or hostile-writer atomicity
 guarantee is claimed.
- Observed original tests:top-level RUN33/PASS33/SKIP0;subtest RUN135/PASS135/
 SKIP0. Final nonempty line exactly PASS;runtime stderr and outer stderr empty.
 No FAIL or instrumentation marker accepted. This is ordinary test evidence,
 not fuzz coverage or a claim that every possible body/path ran.
- Provisional result.json intentionally remains PackagePassed=false and
 FinalizationRequired=true. Only completed.json records PackagePassed=true and
 FinalizationRequired=false;both retain Acceptance=false. Saved-only reader
 accepted the final binding,and a later independent saved-only call accepted
 again. No stopped.json/completed.pending.json/launch-error contradiction.
- One supplemental reporting command passed a relative path to an absolute-only
 safe-file helper and stopped before checking pins. It changed nothing and
 launched no workload. A corrected absolute-path saved-only check then verified
 all71/four pins and completion. This is not another preflight/controller/native
 attempt and did not alter any original evidence.

Top-level observed PASS names (all also RUN;none SKIP):

```text
TestLeafBindingProductBoundary
TestLeafBindingCapabilityBoundary
TestLeafBindingPreflightPriority
TestLeafBindingBudget
TestLeafBindingConsumers
TestLeafBindingIssuerBindingBoundary
TestLeafBindingEqualSKIValue
TestLeafBindingUncoveredReferences
TestLeafBindingTemporalInheritance
TestLeafBindingErrorPriority
TestLeafBindingCancellationAndPanic
TestLeafBindingFixtureControls
TestLeafBindingAPI
TestLeafBindingLowerErrors
TestLeafBindingLowerOwnership
TestLeafBindingProjectionControls
TestLeafBindingOwnedComplete
TestLeafBindingOwnedAliasIsolation
TestLeafBindingNormalizeTimes
TestLeafBindingModel
TestLeafBindingOwnedBinding
TestLeafBindingOwnedCertificate
TestLeafBindingOwnedContent
TestLeafBindingOwnedPolicy
TestLeafBindingOwnedSummary
TestLeafBindingOwnedTable
TestLeafBindingOwnedTableOrdering
TestLeafBindingOwnedTime
TestLeafBindingOwnershipBoundary
TestLeafBindingProjection
TestLeafBindingSameKeyDifferentDER
TestLeafBindingWindowEdges
TestLeafBindingYearOneWindow
```

All135 observed subtest RUN/PASS names are separately retained in TestOutput
within the bound result.json and execute/stdout.txt,not conflated with33 top-level
tests. Raw names/streams remain ignored local evidence,not restored by cloning.

### Durable bindings and recovery

Root:.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/leaf-ordinary-check-001.
Adjacent leaf-ordinary-check-001-control retains wrapper,71 source/four input
snapshots,host status,ledger,outer streams and saved evaluation. No raw/cache/tmp/
SDK/binary material is committed or removed. This identity/approval is consumed;
never replay001 or run its binary manually.

| Evidence | SHA256 |
| --- | --- |
| completed.json |3853AD693E2A67EEBF3A4960E3CEC8140B504D88A70497A9ACFD918E917C7A3D |
| evidence-manifest.json |4B0C67A7B4588382B03A9B156C74D9B06B10EB0AD5C65EA17EAF991887A70310 |
| result.json |132D2A54EFDF5F0392B25C48E7F8AAFDEE2BF8C36BB5DE5AE783BCA9B3B15DB5 |
| build/receipt.json |A783DC98D446DAB0804C6A7ABDF32961723C268CC47E3E2F908BEFF727A8D2C3 |
| execute/receipt.json |3F89C7E8145BA43349669C838DBFA036CF92AC02D2883D4F5F309C7ED78C08F6 |
| control/run.ps1 |ED4E06A1DFC9BF68022E2E8CB656B032C06CFB74152138E3162D877FF605D6DD |

Use latest codex/c12-b01-task9-coordinator tip for portable progress,not main.
Cloning restores code/tests/plans/report,not this machine's SDK/cache/raw evidence.
Historical matrix003 remains2 PASS/1 FAIL/23 UNRUN;old error87/timeouts/fuzz,
physical PostgreSQL/PITR/Batch01 and broader C11/C12/Task9/10/15/18/security/
production/main gates stay open. No automatic matrix continuation or merge.
Next is a saved-only remaining-gate triage,not another real run by implication.

Operational rulings:retain existing worktree and PowerShell control bookkeeping,
no Bash/install/whole-Go baseline (explicit scope;cost:no broader regression);
run existing controller without implementation/live RED-GREEN (no new source;
cost:no new implementation evidence);supplemental absolute-path correction is
saved-only reporting,not workload retry (no native seam called;cost if wrong:
an incorrectly labelled attempt count). Keep branch/worktree/raw controls intact
under AGENTS preservation instead of merge/PR/cleanup (no integration authority;
cost:no main acceptance).

### Independent final evidence review

One fresh read-only reviewer found0 Critical/0 Important/0 Minor in the five
documentation files. It independently accepted the saved reader and verified
all six published hashes,71 source pins/aggregate,four inputs,701 matching saved
product rows,33/135 separate test counts and both native receipt intervals/
ordinary facts. No fixes or deferred minors. No workload,full inventory or file
mutation was performed by the reviewer;this is not an implementation rereview.

Executor rulings on every declined area:

1. Re-execution/mutation failure cases remain untested in this operational
 review;no such workload was approved — cost:no new failure-injection evidence.
2. Full fresh SDK/product/module-cache inventories are not repeated by reviewer;
 existing controller audits/bindings support the report — cost:no independent
 fresh machine-state certification or whole-cache readiness claim.
3. Historical errors/matrix/fuzz/unrelated commits/prior fake completeness remain
 outside this evidence window — cost:those existing gates remain unresolved.
4. Exhaustive analyzer/process identities/test-path coverage remain unclaimed;
 finite receipts/output cannot establish them — cost:coverage remains bounded.
5. Hostile-writer atomicity/arbitrary wrapper persistence-failure correctness is
 not certified by this observed successful run — cost:external-threat and
 unexercised implementation guarantees remain unproved.
6. Past root absence/clean state/approval cannot be recreated by saved-only
 review;executor observed checks and direct user messages are the authority,
 not retrospective hash attestation — cost:no independent historical attestation.
7. Remote preservation is the executor's final check;main/production/broader
 acceptance remains excluded — cost:review alone establishes neither remote
 synchronization nor integration readiness. Final handoff supplies actual
 remote/local equality rather than inferring it from this review.

## Historical first preparation — stopped before preflight execution

User explicitly approved the [concrete plan](../superpowers/plans/2026-10-16-leaf-ordinary-check-001.md)
with “是” after proposal checkpoint2734cad9da2a82670be5cdd7d1c261850526d72c.
No real ordinary package result was obtained.

## Observed failure

The executor's first preflight command contained an unquoted Bash-style path
brace expression: `docs/diagnostics/isolated-fuzz/{isolation,envelope-sample,provenance}.ps1`.
PowerShell reported ParserError / “Missing argument in parameter list” at the
comma and exited1 before executing the command body. This was an executor
command-construction error, not a product/Go/collector/test failure. The preceding
read-only skill/plan/spec reads succeeded; they did not invoke the controller.

No preflight Git/helper command in that failed command body ran. No control
wrapper or output root was created, no controller was called, no native build
or runtime was attempted, and no SDK/product audit or package acceptance was
obtained. Post-stop read-only inspection confirmed both proposed directories
leaf-ordinary-check-001 and leaf-ordinary-check-001-control remain absent and the
development worktree was clean before this documentation handoff.

There is no controller stopped.json/receipt to claim: the failure occurred before
the controller/control directory existed. This portable report records the
actual outer-tool failure, rather than inventing native evidence.

## Scope and next decision

The approved plan explicitly says even a failed preflight returns for a new
decision. No corrected preflight or controller call was attempted. Task1 is
incomplete;Task2 never started;Task3 records this stopped preparation rather than
a completed native attempt. Checkboxes are not marked complete for unperformed
work. Historical error87, real matrix003 failure, timeouts/fuzz and broader
C12/main/production gates remain open;PackagePassed is unestablished and overall
Acceptance remains false.

Next recommendation: authorize correcting the executor's path list to explicit
PowerShell-compatible file arguments, parse-checking the complete wrapper before
any workload boundary, then restarting preflight under a revised single-run
decision. Because001 never created a root or crossed the call boundary, its name
can only be reused if that new decision explicitly permits it;the previous
approval must not silently be replayed. Do not change product/helper source or
weaken evidence gates. No correction or live retry is claimed here.

Ruling: retain the existing worktree and PowerShell-only bookkeeping;no Bash
skill scripts, installation or whole-project baseline — the approved operational
scope permits only this bounded run — cost: no broader regression evidence.
Ruling: stop on the outer parser failure, with no implementation/reviewer/fix
cycle — no implementation changed and the operational stop rule governs — cost:
the intended real validation remains undone pending a new decision.

Recovery: use latest codex/c12-b01-task9-coordinator tip, not main. This report,
plan and status/recovery documentation are portable;machine inputs still need
fresh validation. No raw SDK/cache/binary data is committed.
