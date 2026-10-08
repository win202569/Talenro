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

Task 10 Step 1 is not fully complete: physical connection loss/replacement and
pool mismatch on a real database still require integration proof. Eight isolated
source groups now exercise the private algorithm and its ordinary DBTX binding.
Steps 2–5 remain open: v7-gated Coordinator-only production factory,
exact certificate/trust and desired/root/metadata SQL projections, generated
store, defensive typed readers, real PostgreSQL/PITR and full sub-gate evidence.
Task 15/18 v7 readiness and existing I2/I3/security gates stay open. Do not merge
main or report Task 10, Batch 01, fresh-machine recovery or C12 acceptance.
