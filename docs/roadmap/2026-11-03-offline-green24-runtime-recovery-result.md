# GREEN24 runtime recovery execution progress

Status: STOPPED / PREPARATION_PARSER_FAILURE / NOT_ACCEPTED.

Final archive verified609files/135directories/31zero-byte, decoded bytes equal.
SHA256:F4CAFB362F3B8CA565228E81104AF6C58D9EB729455E318773D5DF2C6CAC6F39.

## Latest: STOPPED at preparation parser; no RED25

Preparation/run/preservation helpers were authored. An added outer foreach in
the preparation helper lacks its closing brace. Final syntax preflight failed;
assistant orchestration incorrectly still invoked the preparation command.
That command exited1 at parsing before executing its body. Both are assistant
errors, not a product/Docker failure. No preparation input/approval, auxiliary
intent, source edit, snapshot or host was produced. Counters24/4/3/186 unchanged.

Preserved the malformed helper and full preparation tool call/result. Created
history-contract-runtime.stop.json explicitly because a parse failure cannot
reach the helper's own catch. All13old stops remain;14stops total. No patch or
retry after the failed preparation, per the approved first-unexpected-result rule.
Next needs explicit authorization for a versioned authoring correction; do not
replay this stopped preparation or consume RED25 under the old stopped scope.

## Latest test-authoring checkpoint

Superseding authoring update:all20group bodies and Invoke-ODRuntimeBindingRegression
now exist; parser reports zero errors and zero top-level executable statements.
No regression invoked. Preparation/run/preservation helpers and immutable fixtures
remain unfinished, so the test is not yet frozen or ready to run.

Fixture contract for preparation:Binding.FixtureBindings contains Parent
(Path,SHA256),WriterSHA256,ExternalHistory and Stops(path/hash map of13old stops).
Binding.Green24Artifacts contains SHA256(path/hash map) and Records with actual
Amendment,Intent,Final,Tool objects read from preserved GREEN24 artifacts. Production
validation must authenticate these actual records, not trust input booleans.
Returned HistoricalGreen24 is the authenticated SHA256 map. Exact rejection
messages in the new tests distinguish approval,parent,writer,receipt,candidate
mismatches. New candidate filename uses the plan's dot separator.

The earlier partial-draft description below is historical and no longer describes
the current test's completeness; it still accurately records that no run occurred.

Created L/history-contract-runtime.tests.ps1 as an explicitly unfinished,
definitions-only draft. It contains the20ordered group names, local AST parsing,
scalar command extraction, read-only definition allowlist, precise rejection
and first-failure receipt helpers. No orchestration entry point exists yet,
so this draft cannot be mistaken for the complete regression suite.

Next: implement all20group bodies and immutable fixture/pre-post recording,
then author preparation/run/preservation helpers. Do not freeze or run this
partial draft. No production change or test invocation has occurred.

The user replied “批准” after implementation plan commit
1e7fd083a2e8cd4b2c94f50a72541216da53ac15. The plan is now approved for its
exact bounded sequence; original plan/design bytes are retained unchanged.
Do not ask for duplicate execution approval merely to resume this plan.

Read-only preflight checked GREEN24 amendment's31source hashes,2356preserved
files and helper hashes against actual files: no mismatch. Saved GREEN24 has
20PASS. The P/history-contract-runtime namespace remains absent. No runtime
test/input/intent, production edit, Prepare, new snapshot or host was created.
Counters remain24aux/4hosts/3compiles/186snapshots;13old stops retained.

Execution ledger:

- Task1 in progress: baseline preflight complete; complete20group test and
  preparation/run/preservation helpers still need authoring and parsing.
- Freeze test/helpers/input only after all fixtures and scalar AST consumers
  are checked. The explicit candidate path ends in runtime.candidate.input.json
  (dot separator), exactly as the68entry plan inventory specifies.
- Before any source edit: archive and remotely verify preparation, then run
  RED25 exactly once and require the plan's precise failure disposition.
- Task2 not started; Task3 not started. Do not relabel GREEN24 as new-source
  verification, replay24 or invoke a stopped historical driver route.

Interface preflight: Task1 produces immutable input/test and actual RED25 for
Task2; Task2 produces candidate and actual GREEN26 for Task3; completed freeze
and remote archive must precede the host. None of these outputs exists yet.
The plan's explicit native PowerShell, no reviewer, no cleanup and bounded
verification rules govern over generic skill wrappers/extra-suite suggestions.

Resume at Task1 test/helper authoring under the existing approval. This document
is a progress record, not a substitute for an actual test receipt or completion.
