# Compile-stage retained diagnostics — implementation record

## Approved scope and execution

User approved the [written design](../superpowers/specs/2026-10-16-compile-stage-retained-design.md)
after54dc5018, then the [three-task plan](../superpowers/plans/2026-10-16-compile-stage-retained.md)
and Native/inline execution after5b679ed5. Feature base is
5b679ed543a6c27b3f192561e1cad696af0c7ba0, on the existing isolated development
worktree/branch. No real native Run, Go compile, generated binary execution,
SDK audit, PID lookup, Docker, download, security change or main merge is included.

Implementation source/tests and one independent review fix pass are complete.
Final post-fix fake-only verification passed on PowerShell7 and Windows PowerShell5.1:
18 suites/698 PASS records per host, including10 STATIC records per host;
all36 commands exit0 (1396 total records, static included). All36 logs were
recounted against summaries and all61 post-fix source/test SHA256 pins matched.
This is a development checkpoint, not native/product acceptance.

The staged whitespace check then found a trailing blank line in the new controller.
Only that blank line was removed; its source pin and both123-check controller
logs were refreshed. Other source/test files and their final logs were unchanged.

## Implemented boundaries

- compile-stage-trigger.cs:bounded incremental file reader, exact compiler/package/
  test-source grammar, first candidate, prefix/raw-line hashes/base64, clocks and
  root-gated activation.16MiB/64KiB iteration/64KiB line/60000 polls. Same grammar
  is reused for saved evidence; no shell evaluation, arbitrary trigger regex,
  background thread or process identity inference from text.
- compile-stage-session.cs:zero registry queries/opens before activation; exactly
  one retained registry afterward. Existing32 cumulative attempt/ownership/rights/
  snapshot/release policy unchanged. Root can exit between check and first poll;
  zero samples/cap remain incomplete, not retrospectively accepted.
- process-native.cs:new explicit RunWithCompileStageRetainedDiagnostics, using
  the actual creation handle for a zero-wait root check. Existing entry signatures
  and eager/default mode selection remain. Root supervision/cleanup continue
  after trigger failures. Existing600s runtime/500ms natural/shared5s cleanup
  limits stay fixed; software deadlines do not preempt OS calls.
- run.ps1:four-source loader, optional non-null trigger conversion and unconditional
  refusal when a receipt contains trigger metadata. A default-null native field
  is omitted on ordinary runs; missing metadata is never inferred for old receipts.
  Source inventories and retained child/eager leaf bindings include both new C#
  transitive dependencies. Existing registry code is unchanged.
- leaf-stage-retained-compile.ps1:independent fixed compile-only planner/controller
  under fresh leaf-stage-retained-compile-NNN.15 production source pins and actual
  go.exe bind the plan. Independent trigger/core/lifetime results and product/SDK/
  source/artifact audits, bounded manifest and stopped/completed protocol persist.
  The new controller is not enabled in default or matrix invocation.

Trigger text means observed target-package compiler command emission, not a
shutdown timestamp, live compiler identity or guaranteed short-process capture.
Cache hits/log latency/format changes may prevent useful sampling. Retention and
reading can affect timing/object lifetime. All four acceptance/runtime/package
flags stay false, even if a diagnostic sample later completes.

The existing offline settings/-mod=readonly are retained, not a filesystem/network
sandbox claim. Trigger scanning and prefix verification are bounded; the reused
whole compile-log token checks and full SDK/product inventories remain separate
work, not part of a claimed global memory/time cap.

## TDD and compatibility findings

Task1 observed the named missing-trigger RED, then41 PASS on both hosts; fixture
Windows escaping was corrected without weakening the production grammar.
Task1 commit:e0e971b21cee810f25bafce70d44ce6e301745ba.
Task2 observed named missing-session/optional-report REDs, then eight scoped
suites296 PASS perPS7/5.1,10 static checks each, all16 commands exit0.
Task2 commit:b3c42aa26dcbad05fb4d5bde47d458f21d47363a.
These are scoped historical per-task runs, not final-suite totals.

Task3 missing-controller RED preceded implementation. Initial generated helper
syntax mistakes were corrected; those parser errors are not counted as successful
TDD cases. A preliminary full stage run was deliberately interrupted after focused
tests exposed two provenance gaps. Named FirstMatch and malformed-prefix REDs
then GREEN prevent replacing an earlier eligible/malformed compiler candidate
with a later command. The complete current suites were subsequently rerun after
all review fixes; neither interrupted runs nor earlier successes are final evidence.

Final legacy matrix regression exposed a second reflection-based receipt copier:
it copied the new typed default-null field and triggered ordinary diagnostic
refusal. Only that null-field omission was added to matrix-check.ps1; dispatch,
arguments, shutdown logic and success gates are unchanged. Non-null metadata is
still preserved and refused. Existing retained-child fixed source-count tests
were updated6->8 with explicit new dependency names, not just a looser count.

## Execution rulings and remaining gates

One fresh read-only whole-increment reviewer checked all16 source/test snapshot
hashes twice and returned3 Important findings,0 Critical/Minor. All were accepted
by effect and reproduced as five named producer/controller RED assertions:
complete nonmatching64KiB record, malformed first compiler token, and internally
ordered entry evidence predating deferred activation. One fix pass made each
focused command GREEN. The shared producer/consumer grammar now fails exact-bound
nonmatches and malformed expected-compiler starts while keeping exact-bound
eligible commands/path boundaries. Padding fixtures use shorter nonmatching
records. The stage consumer binds every entry's deferred timing and query1/last
query metadata to the report-level queries, not only the report-level first
query to the trigger. Extra serialized metadata/last-query negative controls
remain in the full suite. No second independent review was performed; coordinator
post-fix regression provides the repair evidence, not a claimed re-review.

- Use PowerShell bookkeeping equivalents of Bash-only skill scripts, preserving
  plan-specific ledger/brief/base/test evidence. Risk:manual bookkeeping needs
  explicit verification.
- Verification stays within approved fake diagnostic suites, not Go/live/full
  product tests or installs. Risk:whole-product/native behavior remains unverified.
- Update the old child source-binding test to assert8 named dependencies. Risk:
  incomplete provenance could otherwise be missed.
- Omit only default-null trigger metadata in the matrix reflection adapter to
  preserve legacy semantics. Risk:incorrect filtering could drop real metadata;
  non-null values must remain preserved and refused.

Reviewer exclusions were explicitly resolved as follows:

- Final suite outcomes:coordinator must verify post-fix logs/source bindings;
  reviewer did not execute tests. Risk:unchecked reporting would misstate coverage.
- Live Win32/error87/capture/latency:excluded by approved fake-only scope; later
  controlled execution needs approval. Risk:native behavior remains unknown.
- Actual Go format/cache/log visibility:synthetic grammar tests only, no promise
  of capture on real workloads. Risk:a later observation may remain incomplete.
- Actual SDK/product/artifact integrity:fake audit composition only this turn.
  Risk:machine inputs still need pre/post verification before later live use.
- Hostile mutation/identity authentication:explicitly not a security boundary.
  Risk:log text and hash checks do not establish a trusted host or process identity.
- Main/product/wider gates and historical outcomes:unchanged/unaccepted.
  Risk:promoting this diagnostic checkpoint would bypass outstanding acceptance.
- Commit/push/recovery:coordinator verifies remote identity/clean state at handoff,
  not the reviewer. Risk:cloning alone does not restore environment/raw evidence.

On-disk source pins do not attest an already-loaded Add-Type assembly. Any later
live proposal must use a fresh64-bit PowerShell host with frozen current bindings.

Historical compile001 remains STOPPED and matrix003 remains2PASS/1FAIL/23UNRUN.
No error87 cause/fix, production/main readiness or wider C11/C12/Task10/B01
acceptance is established. Any later real attempt needs its own approved frozen
SDK/reference/module bindings and fresh identity; never reuse001 or run the
generated binary automatically.

## Recovery

Implementation commits:e0e971b21cee810f25bafce70d44ce6e301745ba (trigger),
b3c42aa26dcbad05fb4d5bde47d458f21d47363a (session/native wiring),
642363131d46906b49dd52ea84c5ee366b7b4e96 (controller/review fixes).
The subsequent documentation-only preservation commit records this evidence.

Local ignored archive:.superpowers/diagnostics/2026-10-16-compile-stage/validation-001.
Its explicit manifest binds56 files:36 final logs,10 focused RED/GREEN logs,
2 host summaries,2 source-pin snapshots,review,ledger,3 task briefs and owner marker.
Manifest SHA256:534AEC5FE3433464C22DA1D166AA017A6827667E6DD272C0952A073D7F23755E.
All56 archived lengths/hashes were verified; the manifest excludes itself.
This archive is not committed or restored by cloning. The plan-specific temporary
execution workspace can be removed after preserving these records; the development
worktree and other diagnostic histories must remain. Portable source/tests and this
report carry the recovery state, not a promise to recover local raw evidence.

Final suite totals per host:trigger47, session15, receipt16 (6 static), stage123,
matrix60, pair45, uninstrumented28, sample22, compile19, preparation19, observe21,
mapping36, lifetime27, integration12 (4 static), lifetime receipt16, child70,
eager leaf99, short child23. No deferred minor review findings remain.

Start from the latest handoff commit on codex/c12-b01-task9-coordinator, the design,
plan, current-status.md and [repository recovery](../runbooks/repository-recovery.md).
New tests use repository code and tiny local synthetic fixtures, not saved
historical receipts or a private SDK. Some adjacent legacy default fake suites
still reference local SDK/module paths for planning; passing them here does not
prove those inputs are prepared on a new computer. No dependency install is implied.
Git does not carry raw receipts/logs, toolchains, generated binaries/caches,
credentials or live process state.
