# Fixed libc diagnostic profile — short design

Status: DESIGN_APPROVED by user after d26780f4. Execution plan is
docs/superpowers/plans/2026-11-07-c12-libc-profile.md;await plan approval before
implementation/runtime. Proposal wording below preserves the original design.

## Intent and evidence

Measure PREPARE,ZIP and DIRECTORY for modernc.org/libc v1.72.1 without repeating
the whole root verifier. The saved graph/root declarations identify this uncovered
dependency near the already-measured sqlite sample. There is no evidence yet that
libc is slow,corrupt or even fully cached. See2026-11-07-c12-diagnostic-coverage.md.
Success remains selected-module evidence,not root-cause proof or C12 acceptance.

## Approaches and selected scope

- Selected by user: fixed profiles in the existing Go helper and PowerShell
  validator. Retain the existing default and share hashing/owned-process logic.
  Cost: profile selection must agree across the two languages and be tested.
- Alternative discussed: separate copied helper/validator. Stronger separation,
  but duplicates safety logic and future fixes;not selected.
- Arbitrary module/version arguments are out of scope;no general cache scanner.

Small change: named allowlists and matching protocol selection,not a new hashing
engine or process launcher. Existing inline author-review boundary is retained.

## Exact profile contract (proposed implementation details)

| Profile | Ordered candidates | Completion |
| --- | --- | --- |
| paired-v1 (default) | golang.org/x/text v0.40.0,modernc.org/sqlite v1.49.1 | 6 phases,12 events |
| libc-v1 | modernc.org/libc v1.72.1 | 3 phases,6 events |

Go adds `-profile`,default paired-v1. PowerShell Read-C12PhaseEvents and
Invoke-C12PhaseOwned add a trailing optional `Profile`,default paired-v1.
Only these exact case-sensitive names are accepted;empty/unknown/case variants
are rejected,not normalized/fallen back. Reject before cache path queries in Go,
and before stream reading/process launch in PowerShell. Existing calls omitting
Profile retain their old semantics;native no-Session calls still skip phase parsing.
Explicit nondefault profile without Session is rejected as an invalid combination.

Use fixed profile resolvers in each language;tests assert exact identities/order.
Do not expose a caller-provided candidate list through the CLI/owner API. The
existing internal Run(candidates,session,emit) remains usable for synthetic tests.
Compute expected phase count from the selected fixed list,not global constants.
Thread Profile through every live and final event-parser invocation in the owner.

Event Schema1 and existing fields stay unchanged for backward compatibility.
Bind the selected profile in immutable claim/result records and CLI arguments.
The parser validates exact candidate identity/version,sequence,session and phase
order against that selected profile;cross-profile streams must fail. No automatic
profile detection from output. START/END and failure/partial/extra-event semantics,
digest format checks and zero-duration allowance stay unchanged.

For libc,expected digest is loaded via existing root go.sum/ziphash agreement,
not taken from an arbitrary CLI input. Current root go.sum records
h1:db1xwJ6u1kE3KHTFTTbe2GCrczHPKzlURP0aDC4NGD0= for v1.72.1.
The final live gate separately checks both ZIP/DIRECTORY digests against that
expected root digest;parser structural completion alone is not integrity success.

## Files and compatibility

Under docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/cache-phase:
main.go adds profile selection;a small profiles.go may hold Go fixed identities;
owner.ps1 adds fixed profile resolution/propagation. Add profile-focused Go and
PowerShell tests/fixtures in new files. Leave phase.go hashing/safety semantics,
vendored dirhash,common helpers and historical invocation recipes unchanged.
Existing paired fixture cases must still pass with Profile omitted. Historical
test scripts use occupied output roots:do not rerun those paths or edit old receipts.
New scoped test entry points must use fresh roots and preserve equivalent cases.

Any source edit invalidates old build bindings. Do not reuse old helper.exe as
libc-capable or relax source matching. A fresh build manifest binds exact Go source,
owner/common,tested profile contract,toolchain and binary hashes;execution verifies
the complete source file set (including additions/removals),not only old keys.

## Verification requirements and proposed budgets

Future implementation plan must enumerate tests before source edits:

- Go RED/GREEN:omitted/explicit paired equality,exact libc list,unknown/empty/case
  profile rejection before cache access;profile candidate data does not leak across
  calls. Preserve existing synthetic phase/hash/path/digest tests.
- PowerShell RED/GREEN:default12-event stream and explicit paired stream pass;
  libc6-event stream passes;both cross-profile directions,wrong version/session/
  sequence/order,extra terminal event,missing END and partial line are rejected or
  classified incomplete as appropriate. Recheck caps,open writer and unchanged bytes.
- Synthetic owned-child integration:libc normal completion,wrong-profile protocol
  stop,partial cutoff;exit must be confirmed. Native empty-Session regression must
  remain raw-text-safe. Unknown profile must launch no fixture child.
- Build and live only after offline gates pass. Live success requires exact libc
  profile/session,6 complete events,3 OK phases,2 matching digests,exit0,no timeout/
  owner failure,empty stderr,and unchanged four root/devtools module hashes.

Suggested finite plan budget (not permission to run now):one PowerShell RED and
one GREEN batch,each60s;one Go RED and one GREEN,each180s outer/30s Go test timeout;
one build180s;one libc live120s. Synthetic children use at most5s each,cutoff3s,
with at most4 such children in GREEN and none in expected RED. Budget includes
regression checks in these batches,not extra historical-suite invocations.
Existing bounded owned cleanup/drain applies. No automatic retries or repair loops;
failed gates stop later stages. Exact recipes,fixture counts and source pins belong
in the implementation plan.120s is a proposed cap,not a duration estimate.

## Safety,acceptance and recovery

Keep fixed worktree/cache coordinates,pinned Go1.26.5,offline environment,
GOMAXPROCS2,path/reparse checks,exclusive roots,immutable receipts and4MiB caps.
Offline fixtures use isolated temporary data,not real module contents. Real libc
input inspection happens only in the approved live stage;missing input stops,never
downloads/repairs. No root verify,Docker,CIM,trace,security changes or main merge.
No broader project suite/acceptance is claimed by this diagnostic-local plan.

Archive exact source/recipe/claims/raw bounded output and observed hashes,after
privacy review and entry length/SHA verification. Do not commit binary/cache/module
contents or credentials. Preserve old evidence. Update status/recovery and push
necessary files to existing codex/c12-b01-task9-coordinator;verify remote equality.
44/34/6/217 and NOT_ACCEPTED remain unchanged whatever libc's result.

Design self-review:default compatibility and native raw mode are explicit;both
producer/parser selection and artifact rebinding are covered. No claim that these
requirements are implemented or tested. After user approval,write the execution
plan before changing source. The names/API details and suggested budgets above
are the author's proposals for approval,not previously authorized runtime.
