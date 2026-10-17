# Reflection metadata reuse — partial improvement, full regression still times out

Unaccepted development checkpoint after8b7294b1. User approved the bounded design:
reuse fixed reflection information once per evaluation, preserve every input
check, no shared input/result cache. Four fresh ordinary-token verification
children maximum, configured30000ms each, stop on failure/timeout, no retry.
The user separately exempted expected-RED demonstration for this pure performance
refactor. No structural/text test was used to stand in for runtime assertions.
Old R4–R7 and all historical budgets remain separate and frozen.

## Change

Product creates a local context containing only fixed SDK GetType MethodInfo
and PSMemberSet.Members PropertyInfo. It passes that context explicitly through
type/passive/shape/integer/token/bool helpers. No context is stored in script or
global state, accepted input metadata is not cached, and the context is not
returned. Internal helpers called directly without a context construct their own;
the public decision API constructs once per evaluation.

Reflection Invoke/GetValue still run for every relevant value. Extended scalar
member enumeration, executable descriptor/Length/Rank rejection, exact types,
ordinal tokens, shape/numeric/array limits, copy-before-decision and all authority
flags are unchanged. Public API body comparison, after removing only context
creation/arguments, equals the original source. Original35 and ownership suites
were not edited. This source comparison supports review, not execution proof.
Candidate SHA256:
`6CF0D354196E37D07DD07F70868B741BFC39E18E89238C563BCF161AAF51B93F`.

New functional regression checks:
- Same input object edited between calls must be reevaluated; prior result stays
  unchanged and later result/Decision are independently owned.
- A clean input decorated later with a scalar member named Members must be
  refused; restoring a plain value must be reevaluated successfully.
- A later root ScriptProperty must be refused without executing its callback;
  a subsequent fresh plain scenario must still pass.

The three tests use full literal result expectations and original35 equality
helpers, not mocks, source-text expectations or a performance threshold disguised
as a functional assertion. They passed in S3 before existing suites ran.

## Executed slots and evidence

| Slot | Purpose | Exit / total ms | Result |
| --- | --- | --- | --- |
| S1 PS7 | Unmodified4096/4097 boundary baseline | 0 /14810 | Both assertions, six phases PASS |
| S2 PS7 | Candidate same boundary probe | 0 /13315 | Both assertions, six phases PASS |
| S3 PS7 | New isolation + ownership all (fixtures/original35/ownership) | -1 /30068 | TimedOut; incomplete, ExpectedResult=false |
| S4 PS5 | Same regression wrapper | Not launched | Frozen after S3 |

All launched children PS7.6.5 Core x64,admin=False,hidden/no-profile/no-shell,
separate UTF8 streams. Stderr empty; owned-host-exited each. S3 configured limit
was30000ms; saved total30068ms includes termination/collection overhead, not a
claim of strict observed total<=30000. No elevated probe, warmup, fallback, retry,
mutant or reviewer was run.

S1 Evaluate4096=13769.489ms,Evaluate4097=22.192ms.
S2 Evaluate4096=12314.290ms,Evaluate4097=31.417ms.
One observed4096 reduction1455.199ms (10.568%). Both hosts used the unchanged
portable boundary probe with original expectations/assertions and two API calls.
Its ProductChanged=false flag describes no source writes *inside that probe*;
the controller separately pins baseline/candidate module bytes across hosts.
This one-pair result is not a stable benchmark, proof of a uniform improvement,
or resolution of the old/full regression timeout. Older timings varied widely.

S3 saved62 complete stdout lines: new3 checks, fixture8, original35, then
ownership ApiRequired/DecoratedInteger/DecoratedBoolean/DecoratedString/RootScalars
(5/10) PASS. ObservationScalars,PassiveDescriptors,SequentialIsolation,PlainControls
and Complete have no completed PASS. Final wrapper PASS absent. Last completed
group does not identify the exact executing statement at timeout. No assertion
failure was emitted before termination; missing groups must not be marked PASS.

## Controller inventory issue found during saved-record inspection

The frozen success-path inventory mistakenly puts original35 HOST before
P1ApiRequired; that existing suite emits ApiRequired before HOST (lines21–22 in
this combined stream). S3's genuine timeout already makes ExpectedResult=false,
so this does not explain or remove the timeout. A separate read-only comparison
accounting for that two-line order confirms the saved62-line prefix; original
raw/controller/index were not edited and no verdict was upgraded. A completed
future run would also need that parser inventory corrected and re-frozen under
an explicitly approved continuation. No retry or S4 followed this finding.

## Integrity and preservation

Before launch, controller froze123 immutable pins:113 inherited entries excluding
the intentionally mutable product path, five previous packet anchors, profile
probe, exact baseline snapshot, boundary probe, and new two regression scripts.
All123 still match after S3. Product baseline hashCF4CB4B5 was preserved separately;
candidate hash was sealed after S1 and before S2. Source only changed once as a
candidate stage (formatting finalized before seal), not between S2/S3.
Original35 SHA25661A3BACF and ownershipC4905925 unchanged. Static parsing passed.
Expected test/phase inventories, four-slot order and stop-on-failure were frozen
before S1. No claim that unexercised PS5/success/timeout-error paths are validated.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `E0916C3FCDE2AE5617D885B0AC3913DCCEA26B1451E92298D882B94D92AB2E8A` |
| run.ps1 | `A60FFB73CB8B4DCF76E0AF7B1ABFDBF7C1B3E5145C21ADC88E7BE757E09348D9` |
| candidate.json | `8DA8CCD87911D7291EFCDE57E2C663917721D48131E655A626C452D46C7D4AE6` |
| invocations.json | `ED00AAF6A760049CB25A35D373F9F549547741AC783F07BE1CD2806826AAA212` |
| S1.stdout | `AA48E00EE5E7D8153C0BEDCAD0FDEAF49EEDB8B49067725A3B4B7CE9C3B63A8A` |
| S2.stdout | `12A8962E5E0530F50045E0AEAA66645456301742858CB794172F961456C966C2` |
| S3.stdout | `38C6DBFDD2529D0323C75EDB108FF3DF94BDB6E63452352097ACC9996A8A2AEF` |
| Each stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| New functional tests | `09BBAFD33ECED7DE1D98E00290A6B0D5E357E80F74A8766DBCC8409131C9B14F` |
| Regression wrapper | `0B7692FE426407AA9431D54BDB1AA964216FB695CD5A202045ABDDCEB5C61069` |

Ignored local packet
`.superpowers/sdd/2026-10-17-p1-reflection-reuse/`
retains scope/baseline/controller/manifest/seal/index/raw; all older records kept.
Portable candidate/tests/wrapper/report/status/recovery enable another computer
to inspect this unfinished work. Clone does not restore ignored evidence/runtime
or authorize reruns; local byte digests may differ after checkout line conversion.

## Handoff

New3/4 spent; remaining S4 frozen. Old extended-view3/7/four frozen, old R3=false,
earlier budgets and review0/1 ineligible remain. Candidate6CF0D354 is NOT accepted;
I1/P1 still unresolved. No main merge, full regression success, C12/project/live/
runtime/release acceptance.

Next recommended scope is cumulative phase timing in the combined regression
path, with unchanged candidate/assertions/30s gate, not another speculative
optimization or automatic S4. Any future execution needs an explicit bounded
allowance and correction of the documented success-path inventory. Performance
improvement alone is insufficient to remove the full regression blocker.
