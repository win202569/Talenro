# Task 10 reader-side boundary — partial, not accepted

Base checkpoint: `92651a31` on `codex/c12-b01-task9-coordinator`.

## Approved scope and contract

The user approved interfaces and isolated unit tests without Docker/C12 startup,
database integration, PITR or final acceptance. The existing approved Batch 01
Task 10 and Abort/serving amendment section 7.1 supply the contract; the v7
amendment section 8.3 remains mandatory for eventual production serving.

This increment adds `BoundAuthorityReadSource` with the frozen callback signature
and a private reader-side staging primitive. It holds both facts and domain
errors until the bound source returns, uses cancellation > authority > domain
error precedence, returns only finite sentinels, and clears temporary facts on
failure. It rejects absent/nil/duplicate callback capabilities. Typed readers
must eventually supply exclusive defensive facts and complete cleanup functions.
No public constructor can turn an arbitrary source into a production reader.

## Pre-flight and rulings

- Task 9 produces Coordinator readiness and an owned PostgresRepository, but its
  current `CheckReady` is not the Task 10 physical-connection capability. Task 15
  v7 readiness is also unfinished. Neither may be relabeled production serving.
- Ruling: implement only the consumer boundary now. A recording source models
  external pre/post checks, while tests exercise the real staging/error/cleanup
  code. This does not prove physical same-connection, head equality, pool
  mismatch rejection or failover handling. Cost of confusing the two: unsafe
  authorization; those source tests remain explicitly pending.
- Ruling: retain the approved interface, add no exported factory, SQL, generated
  store, schema, HTTP route or listener. The private generic helper will support
  both planned typed readers; complete certificate/trust/root/metadata/desired
  projections remain pending. This is a prerequisite, not a usable reader.
- Ruling: use existing value-free authority error sentinels for the internal
  boundary. Unknown query/driver errors fail closed; wrapped finite errors are
  reduced to their sentinel without retaining private messages.
- Ruling: full repository/C11/C12 tests remain prohibited by the unresolved
  historical Defender event and the user's narrowed scope. Only isolated unit
  and inspected dependency tests may run. No dependency downloads or security
  configuration changes. Bash-only skill bookkeeping is not run on this
  PowerShell-only machine; this portable note is the partial execution ledger.

## Evidence

- Initial RED: the test package failed to build because both the frozen source
  interface and `guardedRead` were missing.
- Behavioral RED: a fail-closed stub compiled, then failed order, error-priority
  and cancellation assertions (1.586s package duration, exit 1).
- Initial GREEN: the real staging implementation passed all three groups and
  29 subcases (2.328s package duration, exit 0).
- Independent read-only review found one P2: joined query errors could let a
  domain classification outrank authority unavailable. Two added assertions
  failed before the fix (2.416s, exit 1); authority classification now precedes
  domain classification. A third joined cancellation case also protects the
  highest priority. Final serving suite: three groups, 32 subcases, PASS
  (2.691s, exit 0). Contracts and readiness package suites also PASS (3.670s and
  2.494s); the inspected Coordinator-only regression PASSes (3.197s, exit 0).
- Final review Ruling: the callback context belongs to the bound source and may
  be canceled during normal cleanup. The consumer checks cancellation during
  query and the caller context after source return; the source must report real
  post-check cancellation through its error. Rechecking a scoped context after
  cleanup could reject a valid read. Cost if the source violates this contract:
  missed cancellation; source-level tests are required before production use.
- Final review Ruling: physical connection/head equality, PostgreSQL/PITR and
  production authorization were excluded from this review because they are not
  implemented in this increment. They stay open, never inferred from fixture
  results. The reviewer ran no tests; RED/GREEN evidence above is from the parent
  session. No deferred minor findings, no commit/push and no main merge.

## Follow-up: private same-connection mechanism

The next user-approved increment adds private Coordinator.newBoundReadSource
and boundAuthorityReadSource. The factory accepts no second pool or repository;
it requires the Coordinator's exact PostgresRepository to also own transactions,
and that repository's database must be an exact pgxpool.Pool. It acquires once,
uses the acquired DBTX for pre-check/query/post-check, and releases once. The
ordinary readiness adapter binds even pending-fence discovery to that DBTX,
without an explicit transaction around provider calls.

The algorithm rejects closed/busy/in-transaction connections, backend-handle
replacement, non-ready results, provider/database head changes, database system
or timeline changes, malformed database points and WAL regression. It freezes
pointer-bearing heads before the callback and permits unrelated WAL advance.
Domain errors still receive a post-check; cancellation wins and errors stay
value-free. This is a private prerequisite, not a production source factory.

- RED: missing types initially prevented compilation; a fail-closed stub then
  compiled and failed behavioral order/release/error/factory assertions
  (2.829s, exit 1).
- GREEN: initial four groups passed (2.836s, exit 0); additional actual-readiness
  DBTX binding and WAL/alias cases passed with the full six-group selection
  (2.814s, exit 0).
- Ruling: use the existing ordinary readiness projection only inside a private
  implementation. Task 15's exact-epoch v7 catalog/runtime/activation/incarnation
  proofs, database OID/name and lease gates are still absent. There is no public
  factory or production reader. Cost if this prerequisite is exposed prematurely:
  stale authority could authorize data. It must not be exposed as-is.
- Ruling: connection lifecycle is tested with controlled low-level doubles,
  and actual readiness SQL dispatch with an owned recording DBTX. These prove
  algorithm/wiring behavior, not PostgreSQL execution, failover, transaction
  semantics, or different-database rejection on a live cluster. Those live
  checks remain gated. No full-suite, Docker, C12, commit or push in this step.
- Final isolated regression PASSes: serving 2.892s, contracts 3.402s, readiness
  2.807s, and the BoundRead/Coordinator selection 3.005s; all exit 0. Formatting
  checks are clean and all four dependency lock hashes match the prior baseline.
- Independent read-only review found no P1/P2 in this increment. Deferred minors:
  explicitly assert release on precheck failure/panic paths; cover provider-head
  pointer alias mutation in addition to the existing database-head case.
- Final review Ruling: live connection/failover, PostgreSQL/PITR, v7 proofs/leases
  and production authorization were not judged. They are absent or gated, not
  waived. The private factory must remain unexported until those requirements
  are met; treating this review as production approval would risk stale access.

## Coverage closure and checkpoint authorization

Both previously deferred minor test gaps are covered. Precheck tests now assert
zero releases when no connection was acquired, exactly one release otherwise,
including acquire returning a connection plus error/cancellation. Separate panic
cases require the original panic to propagate after one release from precheck,
query and postcheck. Provider snapshot tests use valid complete heads, independent
database/provider point storage, an unchanged positive control and an in-place
provider-point change that must fail closed.

Behavioral mutation proof (temporary changes to the new uncommitted implementation,
restored before final verification): removing deferred release failed six error
cases and all three panic cases (3.497s, exit 1). Removing only the provider clone
made the changed-head case return nil and fail its assertion (3.103s, exit 1).
Final scoped regression PASSes: serving 6.108s, contracts 7.349s, readiness 4.490s,
and the BoundRead/Coordinator selection 2.439s; all exit 0. There are now eight
isolated source test groups. No permanent production-code change was needed for
this test strengthening.

The user authorized saving the accumulated Task 10 work to the existing
development branch as an unaccepted checkpoint. Earlier no-commit/no-push notes
record those individual steps; verify the remote SHA before recovery. No main
merge, public serving factory, security change or blocked integration execution.

## Remaining work

### Lookup input follow-up (local, not yet committed)

The user approved a bounded CertificateLookup increment: strict node/issuer/
serial/digest validation, constructor-fixed agent trust selection and defensive
copying. Added types.go and types_test.go in internal/nodecontrol/serving.
The node policy follows existing nonzero RFC4122 validation, without adding an
unsupported UUID version restriction. Serials use canonical unsigned magnitude:
high-bit bytes are valid; zero, empty, leading-zero and >20-byte values are not.
Domains are exact lowercase ASCII, 1–253 bytes with 1–63-byte labels; there is no
case folding or trimming. Constructors validate shape only, not certificate
signature, caller provenance or authorization. Epoch/lineage remain DB facts.

Five test groups cover malformed identities, valid boundaries, agent-domain
selection, defensive ownership and fail-closed zero values. Initial skeleton
failed the valid-input/domain/ownership tests (3.904s, exit 1); implementation
passed the full serving package (2.718s). A temporary constructor-validation
bypass caused all 14 malformed-input cases plus invalid-domain coverage to fail
(2.715s, exit 1), then was restored. Neighbor regression passed serving 3.694s,
contracts 3.446s, readiness 3.253s and BoundRead/Coordinator 7.039s (exit 0).
Four dependency lock hashes are unchanged. No Reader, DB queries, public source
factory, security settings, full suite, C12 bootstrap or PITR was touched.

Final restored serving regression passed (2.574s, exit 0); gofmt and tracked
diff whitespace checks are clean. Independent read-only review approved this
isolated increment with no Critical/Important defects. Optional positive tests
for one-character issuers, digit-edge domains and non-v4 RFC4122 UUIDs are
deferred coverage suggestions, not implementation failures. Reader/SQL/facts,
readiness/connection integration, authorization provenance, production/main and
the Defender alert were explicitly outside that review and remain open. The
reviewer did not run tests; timings above are the executor's observed results.
This increment consists of two new Go files and two modified progress documents;
it has not been staged, committed or pushed. Next bounded step: defensive typed
certificate facts and their validation, before SQL/production integration.

### Certificate facts follow-up (local, not yet committed)

User approved the bounded immutable result-type design. Added certificate_facts.go
and certificate_facts_test.go, without changing the prior lookup increment.
Private newCertificateFacts validates the five exact lookup fields; nonzero
RFC4122 certificate/lineage/operation IDs; positive bigint-range identity,
inventory and authority coordinates; matching inventory epoch/lineage; finite
active/recovery_pending/recovery_limited statuses; nonzero ordered timestamps;
1–65536-byte DER with exact SHA-256; and, for agent queries only, the complete
fixed-selector server CA high-water group with matching authority epoch. Trust
operation/sequence are independent of certificate activation, not forced equal.
Non-agent queries reject unexpected trust data. CertificateFacts exposes only
detached snapshots; serial/DER slices and the trust pointer are copied both ways.

Ruling: this is pure private projection validation, NOT committed-fence proof or
authorization. X.509 parsing/signature/profile, SPKI extraction, trusted-time
expiry, full inventory state/endpoint policy and same-read provenance remain
future reader/B03 obligations. Tests intentionally use small opaque DER bytes
to exercise integrity rather than pretending to authenticate an X.509 fixture.
Cost if misused: an unproven projection could be mistaken for authority; no
public constructor, Reader or production consumer is exposed by this increment.

Observed RED: fail-closed stub rejected valid facts and misclassified invalid
lookup (2.971s, exit 1). GREEN: whole serving package 2.541s. Mutation check:
temporarily bypassing validation caused all 37 invalid projection cases to fail
(2.587s, exit 1); bypass removed. Restored regression: serving 4.476s, contracts
5.437s, readiness 3.785s, all passing. No full suite or blocked security-related
bootstrap/integration was run. The last committed checkpoint remains 6752b961;
both lookup and facts increments are local/uncommitted/unpushed.

Final source regression: BoundRead/Coordinator selection 6.328s, exit 0. All four
dependency lock hashes are unchanged and gofmt reports no changes needed.
Independent read-only review approved this bounded increment, with no actionable
Critical/Important/Minor findings. It did not rerun tests. Ruling: its deferred
SQL/fence/readiness/connection/factory and B03 signature/time/policy assessments
remain open acceptance requirements, not waived requirements; production and
main readiness cannot be inferred. Next work is to design the exact combined
certificate/inventory/trust query and committed-fence projection before wiring
these staged facts into a reader; the v7 and integration gates still apply.

### Remaining acceptance gates

### Combined certificate query work in progress

The user approved exact combined SQL, generated code and isolated tests, with
no live DB/production entry or commit/push. Added nodecontrol_serving.sql and
private certificate_projection.go/tests. SQL matches node/issuer/serial/DER/key,
current inventory epoch/lineage, expected authority epoch and allowed statuses.
Separate certificate_activate/node and trust_bundle_publish/global_node_trust
fences match operation/epoch/sequence/scope digest and committed/active states.
Agent trust purpose/listener are fixed; non-agent mode joins no trust row. LIMIT
2 preserves ambiguity for the private selector, which accepts exactly one
validated candidate. Missing/duplicate/corrupt return zero facts with finite
classification. No latest/MAX fallback, schema change or production reader.

Private arguments derive the node scope from canonical 16-byte UUID input and
the global node-trust scope from the fixed protocol transcript. The literal
node golden digest was independently computed with .NET SHA-256; global golden
comes from Task 5's table. The expected epoch must eventually be supplied by
the same bound ready snapshot; range validation does not prove provenance.

Observed cardinality RED 6.152s; full serving GREEN 11.176s. Argument-binding RED
10.824s; full serving GREEN 9.291s. Four dependency locks remain unchanged.
An independent read-only review found no defects in initial SQL/cardinality;
it excluded the concurrently added argument mapping and did not execute tests.
Generated-row conversion, actual SQL behavior, v7/connection/readiness/lease,
PITR and downstream authorization remain open, not review waivers.

The required public verify-devtools.ps1 was launched once, offline, with fixed
PowerShell 7.6.5 (exec session 52922). Its final result was retrieved on the user's
status request: `verify-devtools: passed`, exit 0. On the next continuation the
fixed Go 1.26.5 explicit tools-module sqlc entry was started (session 77003), with
the same fixed GOPATH/GOMODCACHE and offline/read-only Go settings. Generation
completed with exit 0. Generated sqlc version is 1.31.1. The only generated
changes are nodecontrol_serving.sql.go and the matching Querier method; models
are unchanged. A second successful generation compared the entire internal/store
file set and SHA-256 hashes and found no drift (6.367s command time). This uses
a content baseline rather than staging, preserving the no-index-write scope.

Fresh verification after generation: serving whole package PASS 2.977s, and
store compile-only PASS 1.837s (`-run '^$'`, no tests run), both exit 0. gofmt
reported no files; git diff --check was clean. Four lock hashes remain unchanged.
No whole generate.ps1/C11 run or actual PostgreSQL query execution is claimed.
The current private helpers are not yet wired to generated rows. In particular,
InventoryLineageID is nullable and validity timestamps are pgtype.Timestamptz:
the future adapter must reject NULL/infinity and malformed digest lengths before
constructing staged facts. Authority epoch must come from the bound ready source,
never an externally supplied scalar. These are remaining implementation gates.
No dependency preparation, shared-cache repair or security-setting change is
authorized. Query generation is complete; reader integration and Task 10 remain
incomplete. All accumulated work is local/uncommitted/unpushed. No launched
generation/test processes remain running at this checkpoint.

### Still-open acceptance

### Private generated-row conversion follow-up

The user approved a bounded conversion from generated SQL rows to private facts,
without a database connection, production reader, commit or push. Added
certificate_row.go and certificate_row_test.go. Cardinality is checked before
accessing a row. NULL lineage/times, non-finite or unknown timestamp modifiers,
nonpositive signed coordinates, wrong expected epoch and nonexact digest lengths
return zero facts. Agent presence must match the lookup mode, and non-agent rows
must not smuggle any trust fields behind a false presence flag. Exact status,
identity, DER integrity and domain are rechecked by newCertificateFacts.

Driver rows remain caller-owned; successful facts own copies. Ruling: this pure
converter does not clear caller buffers or establish ready-epoch provenance.
The future query adapter must own/clear temporary rows, sanitize query errors,
and stage facts until the same bound source's post-check succeeds. Cost if
misused: stale/unproven data could be treated as authority or retained on failure;
no public reader or call path has been added, and those integration gates remain.

Observed RED 2.528s (normal conversion and invalid-input classification), initial
serving GREEN 2.603s. Digest tests preserve a valid 32-byte prefix when adding an
extra byte. Temporarily relaxing !=32 to <32 made all three oversized digest
cases fail (2.792s); all exact guards were restored immediately. Final regression
PASS: serving 3.119s, contracts 3.506s, readiness 1.977s, store compile-only 2.237s
with no store tests run. Four dependency locks unchanged; gofmt and diff whitespace
checks clean. No full suite/C12/bootstrap/PITR or live DB was run.

Independent read-only review approved this bounded increment with no Critical or
Important issues and confirmed restored length guards. Minor (deferred): use
distinct certificate/lineage/operation UUIDs in the row fixture to better detect
future field swaps; present mappings were verified correct. Reviewer ran no
tests. Its exclusions (database semantics, readiness/fence provenance, production
integration, Task 10/main acceptance, unresolved security alert) remain open,
not waived. No commits, pushes or background processes remain from this step.

### Private query adapter follow-up

Approved scope: fixed generated-query invocation, finite error handling and
adapter-owned temporary-material cleanup; isolated tests only, no live database,
production reader or Git publication. Added certificate_adapter.go and its tests.
The adapter validates parameters before querying, captures generated Scan
destination pointers before each scan, closes rows once and clears decoded byte
buffers and destinations even on partial scan errors or panic. Query/scan/iteration
failures map to AuthorityUnavailable, cancellation wins, and only successful
transport reaches the row validator's NotFound/Conflict classification.

The pinned pgx bytea scanner for ordinary []byte allocates/copies caller-owned
decoded bytes. Cleanup does not target network buffers and is not secure erasure
of immutable strings or runtime copies. Successful facts are detached before row
cleanup. A guardedRead composition test retains aliases to staged serial/DER/trust
and verifies actual postcheck rejection clears them; its success control verifies
published facts survive cleanup. This is test-only composition, not a production
source/factory or proof of epoch provenance.

Observed RED 2.576s and initial full-serving GREEN 3.745s. Final scoped regression:
serving 2.591s, contracts 3.591s, readiness 2.617s, authority selected TestBoundRead
and TestCoordinator groups 2.923s, store compile-only 1.873s (no tests run), all
exit 0. Four dependency locks unchanged; formatting and whitespace clean.
Independent read-only review of the adapter and original tests, plus supplemental
review of the composition test, found no findings. Review excludes live
PostgreSQL semantics, bound epoch provenance, production authorization, full Task
10/main acceptance and security resolution. The prior row-fixture distinct-UUID
minor remains deferred. No commit/push, dependency download or security change.

### Private bound-epoch follow-up

Approved design and Native plan: docs/superpowers/specs/2026-10-08-task10-bound-epoch-design.md
and docs/superpowers/plans/2026-10-08-task10-bound-epoch.md. The private authority
source now checks equal positive bounded provider/database epochs before issuing
its private DBTX wrapper. All DBTX methods forward to the acquired connection;
epoch is a frozen atomic value, cleared before postcheck and during panic unwind.
BoundReadEpoch recognizes only the exact private type while active. This is not
a SQL permission revocation or a substitute for the mandatory v7 serving gate.

The private certificate reader has no caller-supplied epoch; it fixes resolution
to the real authority accessor and stages through guardedRead/adapter/discard.
A private resolver seam permits isolated composition tests, not production
configuration or an exported test/source factory. Alias cleanup evidence uses
the actual guard/adapter/discard composition; the reader itself keeps staged
material encapsulated. See the execution ledger for this explicit test ruling.

Authority baseline PASS 2.946s; behavioral RED 3.053s then scoped GREEN 3.100s.
Reader behavioral RED 2.859s then whole-serving GREEN 2.647s. Final scoped PASS:
serving 3.493s, contracts 4.292s, readiness 3.198s, authority selected groups
3.468s; store compile-only 1.894s (no tests run). Locks unchanged; no formatting
or whitespace findings. Independent read-boundary review found no Critical or
Important issues; prior distinct-UUID fixture minor remains deferred. Historical
branch changes outside the read boundary were not reviewed in full. No commits, push, real DB,
full suite, race-detector, dependency downloads or security changes.

### Desired projection phase 1

User approved the private structural-projection design and Native plan. Added
desired_projection.go and its tests only; no SQL/store, existing reader, schema
or exported facts changes. Positive inputs are defensively cloned, snapshots
are detached, and owned byte material/root containers are cleared on discard.
Exact active pointer/ID/version matches, same expected epoch, positive bigint
coordinates, payload/envelope bounds and SHA-256, desired time-window order and
at most 32 structurally continuous root rotations are checked. Empty intervals
are valid only at the active version; neither interval base nor opaque bytes
prove canonical encoding, signatures, trust, active fences or authorization.

Observed RED 3.518s and initial serving GREEN 1.410s. No-op discard RED 2.382s;
removing desired payload cloning made Ownership fail 3.134s; restored immediately,
serving GREEN 2.605s. Final isolated PASS: serving 2.527s, contracts 3.188s,
readiness 2.537s, selected authority 4.240s, store compile-only 2.501s (no tests).
Four locks plus serving SQL/generated query/Querier hashes match pre-task values.
Formatting and whitespace clean. Independent review found no Critical/Important
issues. M1 deferred: shared tail/active-root input buffers are XOR-mutated twice
in the ownership test; strengthen single-mutation assertions and explicit
envelope isolation. Implementation copies both correctly. Prior certificate
same-UUID fixture minor remains deferred. This is not whole-branch approval.
No commits, push, dependency download, security change, real DB or full-suite run.

### Remaining project gates after desired projection

Task 10 Step 1 is not fully complete: physical connection loss/replacement and
pool mismatch on a real database still require integration proof. Eight isolated
source groups now exercise the private algorithm and its ordinary DBTX binding.
Steps 2–5 remain open: v7-gated Coordinator-only production factory,
exact certificate/trust and desired/root/metadata SQL projections, generated
store, defensive typed readers, real PostgreSQL/PITR and full sub-gate evidence.
Task 15/18 v7 readiness and existing I2/I3/security gates stay open. Do not merge
main or report Task 10, Batch 01, fresh-machine recovery or C12 acceptance.
### Exact desired SQL and private row conversion — 2026-10-08

Approved fixed-base query design/plan implemented locally. Root chain starts at
version 1; active versions 1..33 only, with 0..32 exact continuous rotations.
SQL follows the same inventory row's exact desired/root/metadata pointers,
requires active intents and exact committed-active fences, retains invalid
diagnostic rows and missing history slots, and bounds the generator arguments.
Private conversion checks row counts/headers, nullable finite timestamps,
signed-coordinate and byte-length boundaries before constructing owned staging.
No SQL adapter, bound desired reader, canonical envelopes or production entry.

Missing-interface and behavior RED; row GREEN 2.929s; ownership mutation RED
2.954s restored, full serving GREEN 3.190s. Independent review found an Important
static guard-test gap (not an existing SQL defect): removing fence-presence,
scope/coordinate or false-default checks could escape the static checker.
Sensitivity RED 4.672s then query GREEN 3.381s; final scoped PASS serving 4.833s,
contracts 6.156s, readiness 4.114s, selected authority 5.269s, store compile-only
2.491s. sqlc 1.31.1 generated twice stably; only serving query/Querier changed;
locks unchanged. Review excludes whole historical branch and real DB/production
acceptance. Existing minors deferred, no new minors. Uncommitted/unpushed.
Full evidence and rulings: 2026-10-08-desired-query-execution.md.

### Desired adapter and bound composition — 2026-10-09

Approved Native plan complete locally, with final independent read-only review
finding no Critical/Important or new Minor issues in this increment.
The private adapter calls the real generated exact query once with fixed scope
digests and the bound epoch. It captures Scan destinations before decoding,
clears partial/full decoded material even on panic, closes explicitly once,
accepts at most 33 scans (33 rows Conflict; a 34th row Unavailable), and owns
undelivered projections until cleanup and cancellation checks finish.
The private reader obtains the epoch from the original callback DBTX before
wrapping it. guardedRead controls postcheck/error precedence and discards owned
projections on rejection/cancellation/panic. Successful scoped-context cleanup
does not invalidate the outer request. The resolver seam is private test
composition, not a new source capability or evidence of live connection binding.

Adapter and reader behavioral RED/GREEN recorded in the execution ledger.
Removing byte clearing made Cleanup/Panic tests fail; no-op discard made
reject/cancel/panic composition fail. Both mutations immediately restored.
Fresh scoped PASS: serving 4.371s, contracts 4.840s, readiness 2.544s, selected
authority 3.437s, store compile-only 3.196s (no tests run). No SQL/generated,
certificate or authority implementation edits in this increment. No public
reader, commit, push, database/Docker/full-suite/race or security-setting change.

Remaining gates are unchanged: canonical envelopes and signature verification,
trusted time/current authorization, real physical connection/DB/PITR evidence,
writer/history/runbook and cross-epoch recovery, v7 proof/incarnation/lease,
production factory, full Task 10/B01/C12/main and existing security acceptance.
Prior certificate-fixture UUID and desired-projection ownership-test minors
remain deferred; new adapter tests do not silently close those historical items.
Full evidence: 2026-10-08-desired-adapter-execution.md.

### Isolated canonical format and private binding — 2026-10-09

Approved local format addendum implemented in new stateformat package. Strict
presence/type/Unicode/depth checks precede exact JCS comparison; string integers
retain bigint precision. Desired, initial root, rotations and metadata enforce
local closed schemas, budgets, sorting, public-key IDs and windows/coverage.
Independent literal envelopes and .NET digest/transcript vectors are frozen.
Private binder revalidates structural staging, then exact payload/envelope/row
coordinates and fixed-base chain; failure/cancel/panic clears owned artifacts.
The existing reader/adapter/authority/certificate/SQL/store files are unchanged.

Scoped post-review regression passed: stateformat 4.598s, serving 5.966s,
contracts 6.635s, readiness 3.713s, selected authority 3.139s, store compile-only
3.626s (no tests). Independent review found no Critical/Important issues;
two new Minors deferred: mutable canonicalization temporary clearing, and an
otherwise-canonical over-budget test input. Reviewer checked protected/vector
hashes independently but did not rerun tests. Detailed
RED/GREEN, mutation evidence and 156 protected hashes are in
2026-10-09-desired-canonical-execution.md. No signature/trust/current-time proof,
production wiring or protocol registration; deployment compatibility UNKNOWN.
No commit/push/merge, DB/Docker/full-suite/race or security changes. Full Task 10,
B02/B03/API alignment, v7/DB/PITR/security/B01/C12/main gates remain open, along
with the two historical fixture/ownership-test minors.
