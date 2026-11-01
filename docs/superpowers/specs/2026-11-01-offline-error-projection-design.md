# Error-record projection: truthful missing metadata, strict acceptance

Status: PROPOSED_DESIGN; approach1 selected by user reply "是". This selects
an approach, not approval of this written design, a plan or new execution budget.
Small boundary amendment; native execution retained, no implementation agents
or extra review. Implementation remains blocked pending design and later plan.

## Evidence and intent

Baseline a4bccf782133176fafe147a7200d74dbeaca00de; see the
[actual stopped writer](../../roadmap/2026-11-01-offline-owner-scope-verifier.md).
The host printed26PASS and journaled tests-complete, then failed in
Capture-Value/capture.ps1:32 while reading InvocationInfo.PositionMessage.
Childexit2/parentexit1; no evidence file; NOT_ACCEPTED. Synthetic ContractOnly
ErrorRecords retained by SharedExceptionDifferentRecordsKept are a plausible
source of absent invocation metadata, not a recovered actual failed object.

Goal: make diagnostic projection faithfully represent a valid ErrorRecord with
missing metadata, without making incomplete real I/O evidence acceptable.
Non-goals: remove the synthetic identity control, fabricate locations/stacks,
accept stdout as proof, reconstruct missing evidence, change writer ownership,
profile behavior, process safety, time bounds or wider product acceptance.

## Approaches and selected decision

1. Selected: null-safe generic ErrorRecord projection, followed by unchanged
   strict real-error consumers. One consistent diagnostic representation and
   minimal format change; requires tests of both projection and rejection.
2. Alternative: special synthetic-control DTO before generic projection. Narrow
   change at one call site, but leaves the generic missing-metadata fault alive
   elsewhere and creates a second projection convention. Not selected.

## Projection contract

Capture-Value keeps its existing signature, recursive behavior and ordered
ErrorRecord keys: Type,Message,ErrorId,Position,Stack. Only inspect the existing
InvocationInfo before accessing PositionMessage. A null InvocationInfo yields
Position=null; otherwise preserve its actual PositionMessage exactly. Preserve
the actual ScriptStackTrace, including null; do not invent a stack or convert
missing information to an empty string. Type/message/id remain actual values.
No catch-all that converts arbitrary projection failures into successful data;
unsupported values still reject. No mutation of the original ErrorRecord or
exception; reference-identity assertions remain on live objects before projection.

Null represents unavailable information, never permission to omit mandatory
evidence. Driver-UnitError/Driver-ErrorLocation keep requiring actual location,
stack and exact writer operation binding. Real I/O errors with null/missing
location or stack remain rejected. The deliberately synthetic owner control
retains ContractOnly=true, its two separate records/shared exception assertion,
and truthful null projection; it must not be presented as an observed I/O event.

## Verification boundary

Propose three new finite auxiliary computations, each once: old-source projection
RED; repaired projection GREEN; new-epoch full scope GREEN. Freeze new separately
versioned auxiliaries/inputs before execution; preserve all old auxiliaries and
receipts. No runtime profile body, record writer, child host or compilation inside
auxiliaries; max2private modules per computation, no exploratory probes.

Tests must use the actual extracted Capture-Value definition, real caught errors
and explicitly synthetic ErrorRecords, not an imitation converter. Cover absent
invocation metadata, present real metadata preserved byte-for-byte, nullable
stack, two records sharing one exception, nested ordered dictionaries/arrays,
and JSON round-trip without mutation or stray output. Independently exercise
the actual strict consumer with null/missing Position/Stack negatives so a
successful projection cannot silently weaken acceptance. Reject unsupported
values and retain the actual first failure/partial groups if projection fails.
Synthetic unit evidence is labelled test data, not reconstructed writer evidence.

The integration computation rechecks actual four imported definitions and all
nine existing scope assertions against the NEW sources, plus the new projection
bindings. Earlier8PASS/9PASS are historical proofs, not proofs of edited bytes.
Only a subsequently successful real writer host with the original26groups,
complete exact evidence, recorder journal, streams and source pins can release
the original contract/later chain. First unexpected result stops; no patch/retry.

## Source epochs, attempts and proposed allowance

Keep all six existing stops,124snapshots, original sources/receipts/archives.
Old writer-scope-green and writer-scope-repair remain consumed immutable history.
A new fixed writer attempt/manifest/snapshot namespace is required. A newest
projection stop always blocks; historical verifier stop is recoverable only via
explicit approval and exact recorded hashes. No public bypass/path/callback.

Semantic repair is limited to capture.ps1. Fixed rebindings/provenance may touch
the same five-source window: driver.ps1,record-boundary.tests.ps1,capture.ps1,
source-model.ps1,observation-contract.tests.ps1. The other26runtime sources stay
unchanged in the new writer epoch; later observation-contract.ps1 implementation
remains conditional on actual writer acceptance. Tests stay outside31runtime
sources. Exact names, receipt schemas, provenance pins and source count will be
specified in the implementation plan before any execution, without circular hashes.

Proposed additions, NOT yet authorized:3auxiliary computations,1replacement
writer host,1fixture compile,31replacement writer snapshots. Cumulative ceilings
would be9auxiliary computations,7hosts,3compiles,186snapshots. Arithmetic:
6auxiliary/2hosts/2compiles/124snapshots already consumed; new writer adds31,
then the original conditional contract epoch retains its31 (total186).
Original remaining4later hosts/7targets/conditional1runtime review are unchanged;
no added targets/reviewers, old failed attempts are never refunded. These are
separate budgets: approving this design alone does not start or authorize them.

Preserve actual tool results and original error records; publish a NEW portable
archive on success or stop, without overwriting prior archives. Save source/tests,
design/plan and clear verified/unverified recovery status on the existing
development branch. No main merge, install, security change or history rewrite.
