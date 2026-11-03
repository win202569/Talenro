# History-repaired contract recovery design

Status: DRAFT / PENDING_APPROVAL. User requested the recovery proposal after
local repair e2d8ec8b3f673fdc0005c09afc570616c6d0834c. Documentation only;
no code, tests, Prepare or host authorized by drafting.

## Intent and actual baseline

Local history-path GREEN21 passed18groups, including real2307entry history and
the extracted production gate call. Earlier binding GREEN19 passed22groups plus
ten command invariants. These are distinct historical source epochs, not proof
that the current complete contract passes. Failed extent Prepare and all12stops
remain immutable. No extent manifest or host intent exists; absence does not
permit replay. Current counters21aux/4hosts/3compiles/186snapshots.

Current driver SHA0F46DF4AA5CF442535426A17FA65672AB21A34514D6ACE42CFDF71226472E10D.
Repair parent remains repair-green SHA
A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
There is no successful intervening freeze to use as a parent.

## Approaches and bounded choice

Recommended: add one enumerated recovery epoch to the existing four binding
consumers, retain original semantic suite and process owner. This follows the
requested new-identity recovery direction. Reusing stopped extent paths conflicts
with immutable attempts; a separate copied runner/suite risks divergent checks.
No generic ignore-stop switch, alternate runtime installation or dynamic filename
selection. This is a bounded recovery change, not an architecture redesign.

L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2;
P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner.

## Distinct identities and allowances

- Epoch history-contract; Prepare grant history-contract-rebind.
- P/history-contract.intent.json; P/history-contract.frozen.json;
  P/snapshots/history-contract.
- Semantic Grant remains contract-green; physical host slot history-contract-host,
  including intent/final/stdout/stderr/evidence/tool record and capture directory.
- Auxiliary slots history-contract-binding-red/green, never host or Prepare names.
- New local recovery stop history-contract.stop.json; all old stops stay required
  historical evidence and old routes stay blocked.

Propose TWO added auxiliary computations RED22/GREEN23, ceiling23; ONE new
Prepare attempt. Explicitly reactivate the previously unconsumed1host/31snapshots
for this route only, rather than silently adding a second host allowance.
MaxHosts1/MaxChildren0/MaxCompilations0/MaxReviewers0. If fully consumed, cumulative
counters23aux/5hosts/3compiles/217snapshots. Legacy downstream3hosts/7targets/
conditional1review remain frozen. No reviewer, child target, compilation or install.

## Four-file source boundary

Modify only driver.ps1, capture.ps1, observation-contract.tests.ps1 and
source-model.ps1 among31operational files. Reuse the tested Driver-VerifyHistory
unchanged, including its three external path/hash identities, containment and
alias checks. Keep the tested command extent branch unchanged. Other27operational
files and pure validators/oracles/recorder implementation remain byte-identical.

Extend exact epoch/slot selection and propagate explicitly through capture,
suite manifest selection, runtime/unit constructors and parent evidence/journal
consumers. Historical constructors retain historical paths; failed runtime
validation never falls back. Keep all49suite cases and3recorder groups identical
in order, mutation and assertion; semantic oracle remains contract-green.

New gate binds current candidate source hashes and actual current regression
receipts, while old candidate/receipts are checked against their own recorded
source identities, not relabelled as current. Verify original writer manifest,
final, evidence and streams; never replay accepted writer.

New manifest directly inherits original repair parent2276pins. Update exactly
four operational entries relative to that parent; append unique provenance for
new approval/spec/plan/input/test/receipts and actual intervening failed attempts,
local repairs and all12stops. Preserve original recovery maps. Store prior binding
and path-repair provenance in distinct maps; never invent a successful extent
manifest. The implementation plan must enumerate exact pin paths/count, detect
overlaps before execution and avoid circular hashes. Do not use old nominal2306
count without accounting for the new append set.

Unit context must validate repair-parent->accepted-writer lineage explicitly.
New recovery metadata must be consumed consistently by actual driver and source
model; names/signatures will be fixed in the plan before coding.

## Verification that exercises the actual boundary

Freeze complete fixture/input/test identities before RED22. Perform mechanical
syntax and fixture-key checks before writing preparation outputs; do not confuse
these checks with a diagnostic pass. Preserve prior malformed drafts untouched.

RED22 proves absent new binding with the actual selector/entry logic, not an
unrelated setup error. GREEN23 must exercise production read-only source/history/
candidate verification on real pinned files, including three external paths,
repository-relative identities and hashes. Do not repeat the prior fact-flag-only
coverage gap or mock Get-FileHash.

Structure the production validation into side-effect-free, callable input checks
and a separate attempt/receipt orchestration boundary. Reuse those exact input
checks from the gate and auxiliary tests. Validate returned values, negative
approval/source/parent/writer/hash/path/epoch cases, consumed-slot detection,
all disjoint artifact namespaces, four-source window and unchanged process/
recorder/suite boundaries. Preserve command extent and path verifier behavior.
The plan must define exact groups, real fixtures, extraction allowlist and
expected RED outcome; no extra runs or replay of18–21.

Do not manufacture a frozen manifest or successful GREEN receipt to test the
post-freeze path before those artifacts exist. Clearly distinguish pre-freeze
checks from the actual post-freeze/host validation. Fully authenticate actual
GREEN23 before accepting Prepare or host execution.

## Attempt ownership and freeze/host sequence

For each fixed action, acquire its exclusive CreateNew attempt record before
resource validation that can fail. Preliminary dispatch only recognizes the
fixed action/epoch and bounded destination. This records early failures instead
of leaving absence of an internal intent ambiguous. The acquired intent itself
is allowed only for that owned invocation, never a general consumed-slot bypass.
Other preexisting outputs/directories or any new-route stop reject. All failures
retain intent and actual error; no overwrite, completion-on-retry or cleanup.

After exact GREEN23: one Prepare validates provenance, writes exactly31snapshots,
verifies them and writes the final manifest only on success. Partial freeze ends
scope. Archive and decode-verify originals, commit/push and verify remote equality
before host start. One host uses the pinned executable/assembly/runtime/privilege
checks and existing owned-process tail,120000ms cap and2000ms exit/drain bounds.

Accept only actual parent Exact/EvidenceValid, original52ordered PASS groups,
exact original stdout/stderr/exit, post-source pins intact, zero children and
no timeout/forced termination/overflow/rescue. Preserve evidence/journal limits
and five false capability flags. A child exit0 or partial evidence is insufficient.

## Stops and recovery artifacts

First unexpected computational result, preparation/freeze failure, drift or
recording error stops all remaining work in this scope without patch/retry.
Even complete success ends after this single host and preservation; not product
acceptance or authority for downstream stages.

Save actual tool inputs/results and raw partial records in versioned portable
preparation/candidate/freeze/final archives; verify decoded bytes/lengths/hashes
and directories/empty files. Commit explicit necessary sources/tests/docs/archives
to current development branch, update status/recovery and verify remote SHA.
Do not stage secrets, binaries or ignored raw records. No main merge.

Written design approval precedes implementation planning and separate execution
approval. None of the proposed new allowances has been consumed.
