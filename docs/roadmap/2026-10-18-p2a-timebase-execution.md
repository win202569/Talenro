# P2-A execution — stopped, unfinished development checkpoint

Readonly follow-up after cf878764: original source/records unchanged and775common
pins intact. Failure-site call graph and the confirmed group-only catch limitation
are documented in the [one-host diagnostic proposal](../superpowers/plans/2026-10-18-p2a-input-mutation-diagnostic.md).
No exact subassertion or product root cause established; no new run/fix/reviewer.
The new one-host proposal is NOT approved and does not unfreeze old allowance.

## Decision and actual result

After cc0663a590547e2203884303a9b75326482cff1a, user approved Native execution:
at most five ordinary PS7 targets, two exact temporary mutants with immediate
restoration, ONE readonly review conditional on all five precise results, fail-stop.
Approved plan SHA256 before status edits:
`0D442900A9B1F3BDB62BDA2AB1F27A06E386ACB92BC9A7E622FA105AE89852CD`.
Immutable spec SHA256:
`A8E8AD45081E73D526324AF15153C0DDB90E531A92A10C03BF6DAF40A1D52705`.

**STOPPED_AT_SLOT_2_INPUT_MUTATION; NOT ACCEPTED.** Consumed 2/5 targets;
remaining three slots frozen, not automatically reusable. Reviewer 0/1, gate not
satisfied. No mutant applied/restoration needed. No repair, retry, extra product
import/probe, global suite, PS5, Docker, live adapter or runtime/security query.
No main merge, promotion, P2-B/C/P3/C12 acceptance. P1 and old limits unchanged.

## Preparation and evidence

Reused linked worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator,
branch codex/c12-b01-task9-coordinator, origin github.com/win202569/Talenro.git.
Initially clean at cc0663a5. No new worktree or trust/ownership configuration.
Ordinary Git hit known ownership rejection; Git-only privileged readonly access
confirmed branch/origin and LF filter identity. Initial malformed filter-check
argument was corrected before freeze; no source-byte drift or target execution.

Six portable pre-product files came first. Static PS parse, JSON, physical LF,
filtered/unfiltered Git blob identity and 24 ordered groups checked. Product
absent through exact RED. Prepare froze 775 common pins: 762 inherited mapped/
current entries, six previous packet anchors, immutable design, six new files.
Historical mappings were preserved. Approved plan SHA is separate metadata,
allowing documentary status edits. Product was then written (eight function
definitions only), statically parsed and sealed without importing it.

Engine PS7.6.5 Core x64:
C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe,
SHA256 `362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139`.
Both hosts: -NoProfile -NonInteractive -File tests -Group all; ordinary inherited
token, reported admin=False. Limits 30000ms execution +2000ms owned exit +2000ms
separate asynchronous raw-byte stream drain. Consumption persisted before launch.
No process scanning/unrelated termination or test elevation.

| Slot | Exit | Elapsed ms | Lines | stderr bytes | Exact expected outcome |
| --- | ---: | ---: | ---: | ---: | --- |
| 1 red | 1 | 1685 | 3 | 0 | yes, precise ApiRequired RED |
| 2 full | 1 | 1599 | 21 | 0 | **no**, InputMutation failed |
| 3 phase-mutant | — | — | — | — | not started; frozen |
| 4 expiry-mutant | — | — | — | — | not started; frozen |
| 5 final | — | — | — | — | not started; frozen |

Both owned hosts exited, streams drained completely, neither timed out,
controller errors null, 775 common pins intact after each. Post-stop readonly
reconciliation again checked all 775 and both pairs of raw hashes. Source stayed
exactly sealed; "good" is a baseline label, not verified correctness.

## Portable stdout transcription

Ordered transcriptions of saved stdout below; not substitutes for raw-byte files
or new execution. Both stderr files empty.

Slot 1:

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
ASSERTION FAILED: P2AApiRequired
```

Slot 2:

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
PASS: P2ANullStop
PASS: P2ANormalStop
PASS: P2ANonzeroOrigin
PASS: P2AFractionalFrequency
PASS: P2AFractionalPhaseRounding
PASS: P2ADeadlineEquality
PASS: P2AStopCutoff
PASS: P2AStageWindows
PASS: P2ALateStop
PASS: P2AOriginLimit
PASS: P2AFrequencyLimit
PASS: P2AIntermediateOverflow
PASS: P2ARequestShape
PASS: P2AExactTypes
PASS: P2APassiveDescriptors
PASS: P2ADecoratedScalars
PASS: P2ARangeRelationships
ASSERTION FAILED: P2AInputMutation
```

Only 18 PASS labels observed. InputMutation failed; OutputOwnership,
SequentialIsolation, RefusalShape, ConservativeFlags and Complete not reached.
No 24-PASS/27-line green, mutant sensitivity, final restored green or review.

## Failure investigation and next boundary

Systematic-debugging used only for readonly Phase 1 inspection. Outer test catch
retains group label, not exception/stack/subassertion. Evidence does not identify
the failing InputMutation operation or distinguish product ownership defect from
test-fixture defect. Root cause unresolved; no repair claimed.

That group builds input/A/B leaves, exercises retained descriptor/payload mutation
and cross-owner checks, changes input root/ObservedTick, compares saved results,
and makes a new normal call. Any can reach the same catch. No particular one is
claimed as the actual failure. No diagnostic instrumentation added or target
re-imported after failure. Frozen sources/tests/controller/oracle unchanged.

Next needs a separately scoped diagnosis/repair/validation decision, preserving
this packet and frozen files. Do not resume old Run, rewrite records, silently
reuse unused slots or dispatch conditional review. Recommend identifying failing
subassertion/fixture boundary before selecting a repair. Any instrumented run
needs separately agreed finite allowance, not expanded runtime authority.

Static-only limits: CodeProperty rejection and internal-error classification have
no injected dynamic coverage. Raw String held-view mutation is not fresh reread
persistence. No catastrophic recovery, concurrent/hostile type-data guarantee,
real clock/scheduling/identity/cleanup/storage evidence. Flags stay false.

## Immutable identities (SHA256)

Portable leaf: docs/diagnostics/isolated-fuzz/p2a-timebase-plan/.

| File | SHA256 |
| --- | --- |
| .gitattributes | E3083818059D2956756D32B1D4A7098E28F57DFB3B971F429A73E51672505B9D |
| capture.ps1 | 92B9DEF5001C2B136FE5195FBBA533B0C3DB8D694E254A98AE6ACEA7933EB3DA |
| expected-output.json | 6D29B86185081976681D9B41B53E2FBA996F60D4436995F48E5CB513659D585C |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| offline-timebase-plan.tests.ps1 | 11BA52E114FC8F14D2D2A020FC6E6C70C52A3F2105B0DADE548416A483B71EB9 |
| README.md | 1082F1424BA9BC3BDE06ED339239B6EE7D6626977B193202049A240D5EEC1E2F |
| source-pins.json | 480B6301C01BAAF63BDE7E855DAD85F889C7B53EE6CCDCA495EB487DBB71EE9E |

Ignored local packet: .superpowers/sdd/2026-10-18-p2a-timebase-plan/.

| File | Bytes | SHA256 |
| --- | ---: | --- |
| frozen.json | 340481 | 9AAD955334C597DFFB8DE2CF68D6007BCF3651434D1B044A8BAEFC7D926FE4E9 |
| invocations.json | 2711 | E0ADE0F19586A010ADC60354960869784905D43EBADBE7C9F8EE922C32940541 |
| product-seal.json | 546 | 5B2131F81C3468773374A77864344919F00832D783138D86914B46CB529C7693 |
| good-product.ps1 | 9607 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| red.stdout.txt | 106 | 31E8C781AB1E8428C2017279C3AEB143B47FDF9170D51CA52143FEAF7511DFCB |
| full.stdout.txt | 555 | 6501C845BA8E87A3507971494CA64995BB898207F7D0A56E50D54DB3EE805064 |
| red.stderr.txt | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| full.stderr.txt | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

## Preservation and skill disposition

Explicit save list: seven portable files, this failed report, plan status,
current-status, repository-recovery. No raw packets/caches/binaries/credentials
committed or deleted. Frozen README report pointer is advisory text; actual
location is docs/roadmap/2026-10-18-p2a-timebase-execution.md. No frozen-file repair.

Executing-plans/TDD/worktree skills guided isolation and tests-before-product.
Verification-before-completion requires actual failure reporting, not completion.
Approved finite route overrides generic extra runs/fix/retest/review/cleanup loops.
Native PowerShell controller/ledger replaces Bash helpers; no runner-completion
claim. Conditional review skill not triggered; no new reviewer spawned.

Git log/remote identity is authority for this checkpoint's eventual commit; this
report does not claim push before it occurs. Fresh machines recover latest dev
branch, not main/automatic replay. Source/tests/oracles and this stop state are
portable; ignored receipts, historical local pins, engine cache/sessions do not
clone. No standalone runtime success or permission is established by recovery.
