# Recover the project on a new computer

Clone `https://github.com/win202569/Talenro.git` and select the branch/commit named
in the latest handoff; do not assume `main` contains an unaccepted checkpoint.
For the latest development work, use `codex/c12-b01-task9-coordinator` and verify
its remote SHA against that handoff. The Task 10 checkpoint adds private read
boundaries and isolated tests only; it does not enable production serving or
clear PostgreSQL/PITR/C12 security gates. See [Task 10 scope and evidence](../roadmap/2026-10-08-task10-read-boundary.md).
The following hash identifies the earlier tooling checkpoint, not the latest tip.
The development checkpoint `fed278f3afae1abcd496caeef1ce855e73cc021d`
on `codex/c12-b01-task9-coordinator` was pushed and remotely confirmed on
2026-10-01, including its prerequisite commits. It is not merged into main and
is not full acceptance. Git carries
the source, generated contracts, tests, designs, implementation plans, and
runbooks. Start with [current status](../roadmap/current-status.md) to resume work.

## Standing preservation rule and latest development work

Latest:[pending runtime-only evidence probe](../roadmap/2026-10-17-task11-metadata-registry-implementation.md#probe-proposal-discussed-after68c2ef90--not-yet-approved).
Question/probe approval is still pending. Proposed0-build/one-runtime diagnostic
would start the same locally bound inventory-test binary in a new diagnostic job;
it cannot accept/rewrite metadata001. No probe control/root or new workload exists.
On a new computer the ignored binary is not restored by Git;missing inputs require
a new preparation decision,not an automatic build. Original budget stays1/1 used.

Latest:readonly investigation approved after45a14dbf is recorded in the
[failure report follow-up](../roadmap/2026-10-17-task11-metadata-registry-implementation.md#readonly-follow-up--approved-after45a14dbf).
The failed OpenProcess precedes ownership/exit/image checks;the natural-wait
loop is skipped by design,not timed out. Native defaults are not positive proof.
Underlying process-lifetime/Windows cause remains unknown;no source/receipt/control
change or additional tests/runs. Used budget still1 build/1 runtime. Resume with
a separate minimal acquisition-window evidence proposal,not metadata GREEN,
replay,elevation/security changes or a diagnostic-mode acceptance substitution.

Latest:[metadata implementation stopped checkpoint](../roadmap/2026-10-17-task11-metadata-registry-implementation.md).
Native/inline was approved afterc549a519. First RED used1 build/1 runtime;build
passed,but runtime's natural-shutdown gate failed (unknown job member/OpenProcess87,
forced cleanup). Final audits0 errors;not accepted TDD RED or metadata acceptance.
Only the initial inventory test is tracked;its two required metadata fixtures
and the product API intentionally remain absent. This is an unfinished development
checkpoint,not a green-test checkout. Do not create GREEN/code on stdout alone.
The saved-only predicate-correction permission cannot waive the native failure;
no correction/continuation occurred. Preserve consumed metadata001 and all old
identities;unused4-build/3-runtime slots are not automatic retry permission.
Resume with a separately scoped readonly observer/evidence investigation. Git
preserves portable failure facts and code/plan,not ignored receipts/SDK/cache.
Do not run old controls or assume their records exist on another computer.
All prior latest/pending entries below are historical.

Latest:the16-item written appendix was confirmed after0ec812cd,and the
[final literal146-row reconciliation](../roadmap/2026-10-17-task11-final-metadata-reconciliation.md)
records local146/0,not independent certification or implemented storage/auth.
The [metadata-only implementation plan](../superpowers/plans/2026-10-17-authority-v7-metadata-registry.md)
is proposed,not executed. Review it and select execution method/new bounded
verification budget before creating product files or native controls. Recommend
Native/inline;maximum5 builds/4 runtimes under fresh metadata-only identities.
No Go/native workload ran for this planning handoff. Existing name-only files
stay unchanged;all previous attempts remain consumed. Storage/consumer/codec,
canonical registration,DB/PITR and production/main gates remain open.
The following earlier latest/pending entries are historical.

Latest: [16-item written data-path appendix](../superpowers/specs/2026-10-17-authority-v7-data-path-clarification-design.md),
based on399f3d21. A/B/C direction approved;this written contract awaits review.
Current local metadata remains130/16. Proposed atomic attempt/activation companions
and per-source retirement-set publication precede dependent epoch signing;storage,
journal,guard and trusted-read integration are not implemented. No new commit
row kind or missing-row restore power. Resume at written review,then final146-row
reconciliation and scoped planning,not SQL changes or consumed-workload replay.

Latest: [16-item data-path review](../roadmap/2026-10-17-task11-data-path-review.md).
User approved the27-request written matrix aftera40c7b45. Current local review
is130/16,not implemented or independently certified. Reconstruct it from original
146-row audit + approved9 + source-backed#93 + approved27 request/profile mappings;
do not turn document deltas into a partial runtime registry. Remaining16 proposals
are grouped5 immutable preimages/3 provider envelopes/8 local trusted paths and
await direction approval. Resume there,then write the16-item clarification;#108
publication must precede epoch source-set signing,not depend on the later seal.
No product/SQL/manifest/key changes,workloads,canonical registration or main merge.

Latest: [27-request authentication matrix proposal](../superpowers/specs/2026-10-17-authority-v7-request-auth-matrix-design.md),
based onec2f55b5. Direction to write it was approved;the27 common request-envelope
choices,11 new roles,empty evidence sets,receiver/profile and archive requirements
are still awaiting written review. Keep current local metadata103/43;130/16 is
conditional on approving that proposal,not present acceptance. Resume at substantive
matrix review,then16 remaining data-path contracts;do not implement,provision keys,
alter canonical specs or replay completed verification identities from this document.

Latest: [current metadata reconciliation](../roadmap/2026-10-17-task11-metadata-current-reconciliation.md).
User approved the nine-item written appendix after8a8fc234,establishing local
102/44 at that approval point. One subsequent executor source review closes#93
using v7 L803/809/811,so current local review is103/43(27 request+16 other objects),
not independent certification or full acceptance. Reconstruct document review
from the original146-row audit plus the10 explicit deltas;do not load it as a
partial product registry. Resume at choosing the proposed mixed request-auth
direction,then writing/reviewing the27-row matrix;43 contracts still block full
metadata implementation. Preserve original93/53 and all consumed workload records.
Older "Latest" and pending-review statements below describe earlier handoffs.

Latest(2026-10-17): [nine-item written clarification](../superpowers/specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md),
based on4fa0e16b. User confirmed5 interpretations+4 known-role request-use
restrictions;the written appendix itself awaits review. Preserve original93/53
and5+48 historical records. Only after written approval may a new local audit
record102 classifications/44 unresolved(27 request+17 other objects);this is not
independent certification. Resume at written review,then grouped clarification
of44 remaining contracts,not metadata implementation or completed-test replay.
No new roles,wire/consumer changes,canonical-set registration or deployment.

Latest metadata work: [53-item consolidated reassessment](../roadmap/2026-10-16-task11-metadata-consolidated-reassessment.md),
based onbdebb846. All53 original exceptions are accounted for as27 request,4
known-role request and22 other-object questions. Five source-backed candidates
are proposed,not approved;48 questions remain after those candidates. Preserve
the original93/53 audit. Resume at grouped direction/candidate review before
writing a unified clarification or implementation plan. This document checkpoint
does not authorize product changes,new request roles,tests or old-workload replay.

Current resume point: [name-directory implementation result](../roadmap/2026-10-16-task11-name-directory-implementation.md).
Plan/Native approved after6c5ec277; fixtures90dc0e2a and read-only API990d6683.
Original five phases verified; approved [null-vector fix addendum](../superpowers/plans/2026-10-16-authority-v7-name-directory-null-vector-fix.md)
also completed its3 builds/3 runtimes once. Important1 CLOSED by observed
first-null RED→GREEN, then full contracts55 top-level/429 subtest PASS,0 SKIP/FAIL
and all final audits. Original RED parser-order stop was retained, not rewritten;
user-approved saved-only correction added0 native runs before two remaining
phases. Only private test decoding and six cases changed; production/fixtures
are byte-identical. No second reviewer or main merge. Resume at separately
scoped deferred metadata/codec work, not completed implementation or old workloads.
Source/tests/fixtures are portable; raw local SDK/cache/evidence is not cloned.
Do not redo completed implementation,
retarget/replay consumed leaf001/name-directory phases or equate cloning with
machine readiness. Metadata93/53 and broader gates stay open. Read current status;
older "Latest" paragraphs below are history, not new execution authority.

Latest:[name-only first-slice written revision](../superpowers/specs/2026-10-16-authority-v7-name-directory-design.md),
awaiting review after user approved scope reduction following7c6194bb. No policy
fields,authorization capability or product implementation. Metadata93/53 audit
remains later work. Resume at revision review,then a name-directory plan;do not
execute original full-metadata/Task11 commands or treat postponed gates as passed.

Latest:[146-row initial metadata audit](../roadmap/2026-10-16-task11-registry-metadata-audit.md).
C2 approved afterf6814efa;93 classifications source-backed by executor,53
reviews unclosed,not53 demonstrated defects. Full exceptions are portable.
No executable metadata plan;resume consolidated resolution or explicit re-scope.
Do not publish partial metadata,fill defaults or equate indexing with authority.

Latest:[C2 admission clarification proposal](../superpowers/specs/2026-10-16-authority-v7-metadata-admission-clarification.md),
awaiting written approval. Resume at review,then complete146-row source-backed
metadata decisions and a consolidated exception list. No executable plan or
implementation approval;do not treat proposed classification as deployed policy.

Latest:[146-row reconciliation/C2 report](../roadmap/2026-10-16-task11-registry-reconciliation.md).
C1 is locally approved after247f1f41,but canonical-set registration/deployment
compatibility stay open. Earlier147 count included the Markdown header;146 is
the corrected literal object set,no protocol object removed. All names/reference
locations are indexed,not all metadata certified. Resume at grouped C2 admission
clarification (or explicitly approved smaller name-only design),not coding/tests.
Older C1-open/147-count paragraphs are historical and superseded here.

Latest C1 work is a [written clarification proposal](../superpowers/specs/2026-10-16-authority-v7-request-auth-clarification.md),
not approval of its new request role or trust-policy rules. No canonical spec-set,
production/source/key/config/test changes. Resume at written review;C1 remains
OPEN and147-row registry planning incomplete. Do not treat the proposed role as
an installed provider capability or start old Task11 commands.

Task11 written design was approved after59bb7226;local plan preparation is now
stopped at [request-authentication/evidence-signing gap C1](../roadmap/2026-10-16-task11-registry-planning-blocker.md).
Do not infer a request role from its response or treat null evidence metadata
as permission to skip authentication. No complete147-row policy map or executable
plan exists. Resume at a narrowly scoped clarification proposal,then finish
reconciliation/planning;do not implement or execute old Task11 commands.

Latest handoff:[Task11 registry/vector written design](../superpowers/specs/2026-10-16-authority-v7-registry-design.md),
awaiting user review. The user approved its conversational boundary and retained
digest prefix,not implementation. Original plan's missing-prefix formula is
explicitly flagged;approved protocol/spec-set and product source are unchanged.
Resume at written-design review,then scoped planning with complete147-row
role/evidence reconciliation;do not run old Task11 test/fuzz commands or expose
existing private signer policies. No new environment or package acceptance.

Latest follow-up is the [post-ordinary remaining-gate triage](../roadmap/2026-10-16-post-ordinary-acceptance-triage.md),
source/report inspection only,no new workload. Leaf ordinary001 is consumed and
passed only its scoped gate. Resume with a proposed Task11 registry/literal-vector
design;do not replay matrices/leaf binaries or infer implementation authority.
Task10 private bound read and several private v7/canonical helpers already exist;
Task11 public layer/Task15 exact-epoch/Task18 manifest and physical gates remain
unfinished. Planned-path inventory is not a completion score or assertion that
all equivalent mechanisms are absent. Prior post-build-only priority1 is outdated.

Latest ordinary check001 completed after renewed approval following62395f01:
one default-vet build73.927s and one original runtime7.287s,exit0;33 top-level/
135 subtest PASS,no SKIP. Full SDK/product audits and15-file completion binding
verified,PackagePassed=true via final marker,Acceptance=false. Read the
[actual result](../roadmap/2026-10-16-leaf-ordinary-check-001.md) for timings,
bindings,prior parser failure and remaining gates. Do not replay001 or manually
run its binary. Raw output/control/SDK/cache stay local and are not cloned.
Resume from latest development tip at saved-only remaining-gate triage;older
stopped/awaiting-approval paragraphs below are history,not current instructions.

Ordinary check001 execution was approved after2734cad9 but stopped before
preflight command execution due to the executor's PowerShell syntax error.
Read the [stopped preparation record](../roadmap/2026-10-16-leaf-ordinary-check-001.md).
No controller/build/test call or root creation occurred;no retry was made.
Resume at a new corrected-preflight decision,not automatic controller invocation;
earlier awaiting-approval proposal paragraphs below are historical.

Latest handoff is the [ordinary check001 concrete proposal](../superpowers/plans/2026-10-16-leaf-ordinary-check-001.md).
Preparation only was approved;real execution still awaits approval. Proposed
roots were absent and four inputs/71-source aggregate matched at53dd9c2d.
No new real build/test or full SDK/product readiness evidence exists. Resume at
proposal review,not controller invocation. A different computer must prepare new
machine bindings/approval;never replay001 or use the historical build-only binary.

Ordinary leaf-check plan was approved for Native/inline with1 after8fa1363c.
Its [implementation record](../roadmap/2026-10-16-leaf-ordinary-check-implementation.md)
currently records planner/flow and initial saved/controller fake checks. Fresh
review found2 Important issues,reproduced RED;one repair pass has targeted GREEN
and final8-suite407-check results per host (1 static included each),all16 commands
exit0,71 source/test pins unchanged. Fixes were not independently rereviewed;
no deferred minors. Earlier pending-plan paragraphs are historical.
Resume from the actual record,not from an assumption that a real run is approved.
Completed checkpoint224780a3dc741984ec2384942f68b8e5a962e3a1 was pushed with
remote SHA=local and clean worktree. Use the latest tip,including follow-up
preservation metadata. The report records the106-file local workflow archive
and seal digest;only the copied/verified plan scratch was removed. Raw archives,
fake outputs,SDKs and caches are not cloned and must not be replayed as acceptance.

Latest stage:ordinary leaf-check written design is approved after9fdda6ff;
its [fake-only implementation plan](../superpowers/plans/2026-10-16-leaf-ordinary-check.md)
awaits plan approval/execution-method selection. Resume there,not at a real run.
The earlier pending-design paragraph below is historical;no implementation or
new test result exists yet.

Latest proposed work:[fixed ordinary leaf-check design](../superpowers/specs/2026-10-16-leaf-ordinary-check-design.md).
Written-spec review is pending;no controller or test implementation and no real
execution is authorized by that document. Preserve default vet/ordinary completion
gates and do not execute the old build-only binary. Resume at spec review,not a run.

Latest follow-up:[acceptance triage](../roadmap/2026-10-16-post-build-only-acceptance-triage.md)
is read-only. It identifies the next design decision and current source/path
checks;it grants no new run, matrix replay or acceptance waiver. Read it before
interpreting older chronological status paragraphs as current implementation gaps.

Latest real diagnostic:[build-only001](../roadmap/2026-10-16-leaf-stage-build-only-001.md),
separately approved under the [single-run plan](../superpowers/plans/2026-10-16-leaf-stage-build-only-001.md).
One compile exit0/331.219s,5 admissions below32,scope/trigger/core/lifetime and
final audits complete;completed marker and10 manifest artifacts verified.
Full11536-file SDK/701 product comparisons unchanged;66 source/test+16 helper+4
input pins verified. This supersedes the earlier in-flight65e54c54 checkpoint.
Result checkpoint `9ed03735ee9ac6f1a8755e223b8c16f5e9d1b999` was pushed with
remote SHA=local and clean state;select the latest development tip,including
follow-up preservation metadata. Seven workflow files were verified/copied into
the ignored control archive before removing only this plan's SDD scratch.
The report records its19-file seal digest;neither this archive nor raw diagnostic
output is cloned. The worktree and diagnostic history remain intact locally.
It is diagnostic-only:ordinary vet/test/production/main acceptance still open,
error87 unexplained,no binary runtime or retry. Read the report before resuming;
do not replay001,run its binary or silently rebuild its environment on a new PC.
These paths are machine-bound:cloning does not restore SDK/modules/raw receipts;
on another machine stop and prepare new bindings/approval rather than replay001.

Latest implementation checkpoint:user selected option1 for a separate build-only
diagnostic with fixed -vet=off;ordinary vet acceptance and32-attempt budget remain.
The [written spec](../superpowers/specs/2026-10-16-leaf-stage-build-only-design.md)
was approved after b8e7956d. Its [three-task plan](../superpowers/plans/2026-10-16-leaf-stage-build-only.md)
was approved for Native/inline fake-only execution after8a25712f. Tasks1/2 are
committed as1c13f5fb/4bd4054f;controller/final repair checkpoint is
a4bfe80ab898e0b6ada5a033a44534d0ff05a9e6. One fresh review/two Important repairs
are complete;final21 suites passed884 checks per host (11 static included each),
all42 commands exit0;66 source/test hashes matched. Read the [implementation record](../roadmap/2026-10-16-leaf-stage-build-only-implementation.md)
and current status before resuming. Full code/test/document checkpoint
ce484a0afc12f5ad4504f4550e43f4370cfbcd9d was pushed with remote SHA=local and
clean state;the latest tip also carries follow-up preservation metadata.
The98-file raw verification archive is deliberately ignored and is not restored
by cloning;the report preserves its manifest digest and all progress/limitations.
Only this plan's temporary SDD directory was removed after a verified archive
copy;the worktree,other plans and diagnostic histories remain. No real execution
or replay of001 is approved.

Latest saved-only [stage budget analysis](../roadmap/2026-10-16-compile-stage-budget-analysis.md)
explains27 vet admissions/32 cumulative slots and225 subsequent logged vet
command/config pairs. No new Run or policy change; no exact command/PID join.
Resume by agreeing a diagnostic design goal,not replaying compile001 or turning
off required vet acceptance. Raw analysis JSON is ignored and not cloned.

Latest real diagnostic:separately approved leaf-stage-retained-compile-001,
base d04baca3,compile exit0/66.574s but STOPPED at32 sampled admissions,27 recorded
vet.exe. Trigger/core evidence and final audits verified; lifetime sampling still
incomplete,no binary runtime,retry or acceptance. Recover the
[actual result and next boundary](../roadmap/2026-10-16-leaf-stage-retained-compile-001.md)
from the latest development tip. The used identity and ignored SDK/cache/raw
artifacts are not restored by cloning; never replay it automatically.

Latest [compile-stage trigger spec](../superpowers/specs/2026-10-16-compile-stage-retained-design.md)
is user-approved. Its [three-task implementation plan](../superpowers/plans/2026-10-16-compile-stage-retained.md)
was approved for Native/inline fake-only implementation, now complete with one
review fix pass. Final18-suite regressions passed698 checks per PS7/5.1 host
(10 static included each), all36 commands exit0;61 source/test hashes matched.
Implementation checkpoint:642363131d46906b49dd52ea84c5ee366b7b4e96 and its
prerequisites; select the latest branch tip for the subsequent evidence/recovery
documentation. The local56-file raw verification archive is deliberately ignored;
its manifest digest and limitations are preserved in the implementation record.
Read the [implementation record](../roadmap/2026-10-16-compile-stage-retained-implementation.md)
for actual results and remaining gates. Recover all three from this branch;
implementation approval is not automatic real execution on a fresh PC, and fake
verification cannot grant acceptance.

Latest saved-only [cap/chronology analysis](../roadmap/2026-10-16-leaf-retained-cap-analysis.md)
explains the cumulative32-attempt limit was reached at23.42s of a74.94s receipt.
No late new-member queries follow, although retained snapshots still occur.
No new execution/code change/acceptance; next is a separately agreed diagnostic
design, not cap expansion or a warm-cache rerun.

Latest separately approved leaf-retained-compile-001 returned compiler exit0 but
STOPPED on the32-attempt retained sampling cap; no runtime/retry/acceptance.
Read [actual result and next scope](../roadmap/2026-10-16-leaf-retained-compile-001.md).
Recover source/tests and portable progress from this branch, not ignored SDK/
cache/raw evidence. Do not reuse001, increase caps or start a full matrix.

Latest prepared diagnostic entry targets only nodeleafbinding compilation, not
test execution; see [scope and portable checks](../roadmap/2026-10-16-leaf-retained-compile-entry.md).
Real Go invocation is not approved by its fake-only implementation. On another
PC prepare/approve the concrete fixed SDK/reference/cache inputs separately; do
not infer live readiness from cloning this development checkpoint.

Latest controlled observation002 after the phase fix passes sampled diagnostic
checks only; see [evidence and limits](../roadmap/2026-10-16-retained-child-observation-002.md).
It does not clear the historical matrix/error87 or production/main gates.001
remains failed and must not be replayed or relabeled.

Latest2026-10-16 diagnostic preparation adds a dedicated retained-child wrapper
and portable fake-only tests. Start with
[its scope and continuation](../roadmap/2026-10-16-retained-child-entry.md) and
current status. The later separately approved single live observation STOPPED
on a wrapper phase-confusion defect; read
[its evidence and unresolved fix](../roadmap/2026-10-16-retained-child-observation-001.md).
Do not retry that identity or treat native cleanup facts as observation acceptance.
The later [observer phase fix](../roadmap/2026-10-16-retained-child-phase-fix.md)
corrects that wrapper check with portable fake-only tests, not a live rerun.
A provisional result alone is
not completion: later observations require completed.json, matching artifact/
manifest hashes and no stopped record. Recover source/tests from the development
branch; ignored receipts, caches and live process state are not cloned.

From2026-10-13 the user authorizes ordinary commits and pushes of necessary files
at completed-work/handoff checkpoints. The durable instruction is [AGENTS.md](../../AGENTS.md).
It does not authorize main merges, force pushes, branch deletion or clearing
acceptance/security gates. Only necessary reviewed files are saved; secrets,
caches, temporary files and ignored raw receipts are excluded.

The current handoff adds declared issuer mapping consistency plus the M1 test
hardening. Its outcome is declared consistency only, not authenticated registry,
production compatibility, database uniqueness or v7 acceptance. Read its
[execution record](../roadmap/2026-10-13-declared-issuer-mapping-consistency-execution.md)
and latest current-status appendices. The final commit SHA is verified remotely
and reported at the actual preservation handoff, not inferred from a pending push.
This checkpoint is unfinished: pre-save matrix3 stopped on FuzzIssuerEnvelope's
context deadline exceeded after24 successful commands; Pair was not run.
Earlier reviewed matrix2 passed26/26, but does not erase this latest blocker.
Recover the saved source/plans and continue the recorded diagnosis; do not treat
the saved development branch as main-ready or accepted production code.

On a fresh computer, clone into a new empty directory using PowerShell:

```powershell
git clone --branch codex/c12-b01-task9-coordinator --single-branch https://github.com/win202569/Talenro.git Talenro
if ($LASTEXITCODE -ne 0) { throw 'Clone failed.' }
Set-Location -LiteralPath Talenro
git rev-parse HEAD
git ls-remote origin refs/heads/codex/c12-b01-task9-coordinator
if ($LASTEXITCODE -ne 0) { throw 'Remote identity check failed.' }
```

Check the local SHA against the remote branch and the handoff SHA. Then read
AGENTS.md, current-status.md, the latest execution record, spec and plan to find
the exact completed work and open blockers. Do not reuse a nonempty checkout or
overwrite unrelated local changes through these example commands.
These are recovery instructions, not a claim that a fresh-machine clone or
physical acceptance was performed. Continue with the separately authorized
dependency preparation and verification gates below.

## Windows test environment

The current full repository suite includes Windows process, named-pipe, and
filesystem tests. Use Windows/amd64 with Git, Go 1.26.5 (selected by `go.mod`),
and Windows PowerShell 5.1 available for the existing legacy test/smoke contracts.
The four development-tool entries additionally require the exact stable
Microsoft PowerShell **Core 7.6.5** runtime; arrange its installation separately.
They reject 5.1, other 7.x versions and prerelease builds. An absolute path to
`pwsh.exe` is supported; no Codex-private installation path is a project prerequisite.
The separate live integration acceptance
also requires Docker and the pinned dependencies described in
[local development](local-development.md).

## Prepare dependencies separately from offline verification

From the selected checkout root, check the prerequisites without changing them:

```powershell
pwsh.exe -NoProfile -Command '$PSVersionTable.PSVersion.ToString(); $PSVersionTable.PSEdition'
$env:GOTOOLCHAIN='local'
go version
```

Expect `7.6.5`, `Core`, and `go1.26.5 windows/amd64`. An ordinary Go installation
alone is insufficient: the verifier requires the exact toolchain module under
`%USERPROFILE%/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64`.
Network preparation is a separate manual/authorized action, never an automatic
fallback from validation. The following preparation commands are instructions,
not a claim that fresh-machine recovery has been executed and accepted:

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='https://proxy.golang.org'; $env:GOSUMDB='sum.golang.org'
$env:GOAUTH='off'; $env:GOVCS='all:off'; $env:GOFLAGS=''
$env:GOPRIVATE=''; $env:GONOPROXY=''; $env:GONOSUMDB=''; $env:GOINSECURE=''
$env:GOPATH=Join-Path $env:USERPROFILE 'go'
$env:GOMODCACHE=Join-Path $env:GOPATH 'pkg/mod'
# Separately authorized exact toolchain-cache preparation, if absent:
go mod download golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64
if ($LASTEXITCODE -ne 0) { throw 'Pinned toolchain preparation failed.' }
$fixedGo=Join-Path $env:GOMODCACHE 'golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
if (-not (Test-Path -LiteralPath $fixedGo -PathType Leaf)) { throw 'Pinned toolchain is absent.' }
$fixedVersion=& $fixedGo version
if ($LASTEXITCODE -ne 0 -or $fixedVersion -cne 'go version go1.26.5 windows/amd64') { throw 'Pinned toolchain version mismatch.' }
$env:PATH=(Split-Path $fixedGo)+';'+$env:PATH
go mod download all
if ($LASTEXITCODE -ne 0) { throw 'Root dependency preparation failed.' }
go -C tools/devtools mod download all
if ($LASTEXITCODE -ne 0) { throw 'Tool dependency preparation failed.' }
```

Keep the committed lockfiles unchanged; do not substitute `latest`, add a
workspace/replace, or disable checksum verification. Missing dependencies during
offline checks remain failures requiring separate preparation approval.

Consumer execution now enforces the verifier's fixed cache/offline environment.
The empty-Dir metadata issue is repaired in the published checkpoint: an offline
exact-version query resolves metadata while preserving identity/cache binding
and complete integrity checks. The standalone verifier has one 3600s budget;
C11 outer verification/check-tools/generate caps are 3660/4200/4200s. These
approved limits do not clear security gates or prove fresh-machine recovery.
See [current verification evidence](../roadmap/2026-09-28-devtools-metadata-fix.md).
Do not change locks or repeat downloads to bypass a failure; inherited
cache/proxy/workspace overrides remain unsupported.

```powershell
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOFLAGS='-mod=readonly'
go mod verify
if ($LASTEXITCODE -ne 0) { throw 'Root integrity verification failed.' }
pwsh.exe -NoProfile -File scripts/verify-devtools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool integrity verification failed.' }
pwsh.exe -NoProfile -File scripts/check-tools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool versions failed.' }
pwsh.exe -NoProfile -File scripts/generate.ps1
if ($LASTEXITCODE -ne 0) { throw 'Generation failed.' }
git diff --exit-code -- api gen internal/store
if ($LASTEXITCODE -ne 0) { throw 'Generated content changed; investigate before committing.' }
```

Run one stage at a time using its original bounded execution policy: verification
and tool checks 900 seconds, generation and lint 600 seconds. Preserve failure
results rather than extending a deadline or repeatedly rerunning an unchanged
failure. Root integrity does not replace tool-module integrity, or vice versa.
Root `go mod verify` checks downloaded module content; by itself it is not proof
that a cache is complete. The public tool verifier also checks cache presence.

## Tool ownership and supported entrypoints

The root module retains Goose and runtime libraries. Buf, protoc-gen-go,
oapi-codegen, sqlc and golangci-lint live in `tools/devtools`. A bare root
`go tool buf` (or another moved tool) is no longer supported. For explicit tool
calls from the repository root, first pass the public verifier and use:

```powershell
pwsh.exe -NoProfile -File scripts/verify-devtools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool integrity verification failed.' }
go tool "-modfile=$((Join-Path (Get-Location).Path 'tools/devtools/go.mod'))" golangci-lint run ./...
if ($LASTEXITCODE -ne 0) { throw 'Root lint failed.' }
# Goose remains in the root module:
go tool goose -version
```

Do not put `-modfile` into `GOFLAGS`. Lint and generation target the root project,
not the tools directory. Nested devtool entries use the same verified PowerShell
executable, not a PATH fallback.

| Entry family | Windows | Linux/macOS/WSL |
| --- | --- | --- |
| verify-devtools/check-tools/generate/verify-c11 `.ps1` | Core 7.6.5 only; C11 full run remains blocked below | Unsupported |
| Same four `.sh` entries | Reject without tool work | Reject without tool work |
| Existing Bash smoke | Retained under its original prerequisites, including Git Bash | No new acceptance claim |

The withdrawn Unix positive implementation and its four historical failing
cases are preserved in commit `03947e31536a803aaca19c295edd8d9eb4d258f1`.
Rejection tests do not turn those failures into fixes. Do not switch Docker to
Windows-container mode merely because the host tool entrypoints require Windows.

## Ordinary and live tests: current execution gate

Full ordinary tests, full C11, and C12/bootstrap/13 physical gates remain blocked
by the recorded Defender event and open I2/I3 acceptance findings. Tool checks
do not lift that gate. Do not disable protection, add exclusions, restore a
quarantined file, or increase timeouts to obtain a pass. Microsoft-account-based
vendor review is deferred; no security clearance is implied.

Only the approved compile-only probe can run before separate gate restoration:

```powershell
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
```

This is not test execution or live acceptance. Race testing remains a separate
gate requiring the supported C toolchain and `CGO_ENABLED=1`.

No pre-existing `.superpowers` directory or caller `-overlay` is required.
The former local source replacement has been integrated into
`internal/nodecontrol/authority/postgres_repository.go`. Ordinary child Go
commands clear inherited `GOFLAGS`, `GOWORK`, and persisted Go environment
settings. The integration factory probe generates its own temporary overlay,
restricted to one test source file; it never replaces production sources.
Some semantic probes create temporary packages inside the checkout so that Go's
`internal` import boundary remains valid. Their input is generated by tracked
tests; it is not a file to copy from another computer.

## What Git does not restore

Go caches, local tool binaries, `.superpowers` session state, raw historical
test receipts, Docker volumes, passwords, signing keys, environment secrets,
and Codex conversation history are not part of the repository backup. Restore
any required operational credentials or database backups through their own
approved channels. The tracked plans and status document preserve project
context, but a new checkout does not reproduce old machine-bound evidence.

## Current acceptance boundary

The earlier full-test result used a local overlay and must not be treated as
proof that the old commit was independently reproducible. After the portability
repair, the then-approved full test command passed on 2026-09-08 at code commit
`16c4e8c4e43ae893aad8779639bec0f80ca096cb` in a new independent clone with a
space-containing path and no pre-existing `.superpowers` directory. The
PowerShell runner canonicalizes its pinned embedded Go source to LF before
verifying the digest, so normal CRLF checkout does not require a local patch.

That historical result does not authorize rerunning the currently blocked suite.
After separate security-gate restoration, follow the current stage plan and its
original budgets for verification on the new computer and live acceptance;
C1.2 Batch 01 and the full product remain incomplete, as recorded in
[current status](../roadmap/current-status.md).
