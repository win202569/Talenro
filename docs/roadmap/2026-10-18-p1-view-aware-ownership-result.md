# P1 view-aware ownership attempt — incomplete, not accepted

Date: 2026-10-18. Approved design: d10d39ee61fd4053866d8f87984b846b00ba9d22.
Branch: codex/c12-b01-task9-coordinator. This is an unfinished development
checkpoint, not a main merge or a completed ownership/security review.

## Outcome

NEW focused slot 1/3 consumed. At the 30-second wall limit the controller
terminated only its owned host (PID 46000); exit confirmed, both streams drained.
Elapsed 30093ms, exit -1, TimedOut=true, StreamsComplete=true,
PostPinsIntact=true, ControllerError=null, stderr=0 bytes.
Outcome EXECUTION_INCOMPLETE; SuitePassed=false; FormalAcceptance=false.

No retry, repair, extra probe/host or reviewer followed. New slots 2/3 (full API)
and 3/3 (existing ownership) are FROZEN_UNRUN. Their capture files do not exist.
The separate OLD repair remaining 2/3 and all other frozen allowances remain
frozen. Nothing about this new attempt renews them.

## Verified prefix, not a full pass

The final saved stdout has 103 of the required 153 ordered nonempty lines.
A read-only comparison against the PRE-FROZEN inventory confirms the whole
103-line prefix and all 37 of its 53 expected fixed-schema facts. This comparison
did not launch a host, import the module or alter the oracle/controller.
There is no terminal RESULT, no control-after PASS and no exact final call count.

41 PASS labels comprise:

- Two fixture-only positive sharing controls: shared pre-stored String wrapper
  and shared boxed Int32; independent counterparts remained clean.
- All 15 direct scalar vectors: null; Int32 zero/min/max; Int64 zero/min/max;
  Boolean true/false; String clean/refused/unknown/policy-only/supervisor-lost/
  internal-error. These account for 29 completed DIRECT scalar calls and one
  direct metadata helper call.
- Initial normal decision control.
- 23/39 behavioral cases: all seven original rows in input/A/B directions
  (21), plus root-schema A/B (2).

The next line is BEGIN: root-mode-A, without a fact or PASS for that case.
The 23 completed cases account for 46 completed decision evaluations plus the
initial normal control = at least 47 completed decision calls. The next case
may have attempted up to two more; the exact interrupted count is NOT recorded.
The declared maximum remains 80, not an observed completion count. Calls to
helpers inside decision evaluations are not included in the direct counts.

All 37 facts report effective held descriptor/payload mutation, unchanged field/
base, clean other retained views AND fresh rereads, and fresh selected bases.
Observed storage/capability aggregates:

| Actual stored type | Reread metadata retained | Facts |
| --- | --- | ---: |
| System.Int32 | true | 8 |
| System.Int64 | true | 6 |
| System.Boolean | true | 8 |
| System.Management.Automation.PSObject | true | 3 |
| System.String | false | 12 |

The three stored wrappers are the selected String input cases. The 12 raw String
facts have positive held-view mutation evidence but no retained field-reread
metadata; absence alone was NOT their pass condition. This is public-view proxy
evidence, not a full CLR binder trace. Positive fixture controls demonstrate
known sharing visibility, NOT candidate-mutant sensitivity. Historical Int32 RED
remains preserved; no fresh product-negative-control proof exists for revised
assertions. No ownership gap was reported in the completed prefix; later cases
and the full run remain unknown.

## Immutable sources and portable additions

Candidate unchanged SHA256:
1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8.

[Portable test/controller/oracles](../diagnostics/isolated-fuzz/view-aware-ownership/README.md)
are NEW standalone files. The old focused test, full API test, ownership test,
module and diagnostic leaves were not edited. Complete independent literal
13-root/22-Decision expectations and the original 39-case matrix are preserved.
String input wrapping occurs before evaluation; no post-evaluation input/output
replacement. The controller freezes ordered labels, fact schemas and sources
before launch and preconsumes each slot. It does not relax process safety.

Portable file hashes (SHA256):

| File in view-aware-ownership | SHA256 |
| --- | --- |
| .gitattributes | A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9 |
| capture.ps1 | DFFF04AFCC10B8ECB66D3D8208374422A7ADE81B4D98B0B6626666CC6DF54A20 |
| expected-output.json | 8E1F82BD788437149A1842152BE41228E23510BE874CDE02D07CC7C1BDA2ADA2 |
| owned-return.tests.ps1 | F06D199BE4E592E1677B5654DBE4D93EA98A82316E0A3E8B66E0A481BA271A12 |
| README.md | 6BB3FE3A6879CB170A83009A6AEFD36FEA33B8DC542B79FC7B2119146B9EF133 |
| source-pins.json | 05F55154BEEF32069A1CB69B83065148DA8AA714C64E7355A3C93454DBCF34C6 |

Both new scripts passed parse-only checks before preparation. A source-generation
replacement-string expansion error in the NEW controller was caught by static
inspection and corrected BEFORE freezing or any host launch; no frozen bytes
were repaired afterward.

## Preserved raw anchors (not portable executable prerequisites)

Ignored original packet:
.superpowers/sdd/2026-10-18-p1-view-aware-ownership/

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 31A3934BB45726B0DFE0422A18E624B12D65421E1EB6FDC779186F292A3825EF |
| invocations.json | 693CD668355EF3685BB49AA0D6164122C347ECBAB936C516E5C3D53F0E16CC18 |
| focused.stdout.txt | B68032819D9F00BBDB8F630687A34A62569FA4B283696259BCD96AED29E7B265 |
| focused.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

582 frozen pin entries = 566 inherited + four latest diagnostic anchors + six
unchanged source anchors + six new portable files. All checked intact before/
after the target run and again during read-only reconciliation. Earlier frozen
records are preserved, not overwritten or reconstructed.

## Remaining work and boundaries

The timeout's cause is not established by line ordering. Static source inspection
shows repeated walks of accumulated prior views, with reflection-based field/base
invariants before and after each mutation: a plausible growing observer cost,
NOT a measured attribution and NOT evidence that the candidate itself is slow.
No per-case timing was captured. Do not weaken cross-owner checks or increase the
time limit solely to make the run pass.

Recommended next scope: read-only cost/coverage analysis of the frozen observer
and its saved prefix, then a separately approved bounded redesign if justified.
Do not run the two frozen later slots, edit this frozen leaf, or retry automatically.
A new computer can inspect these portable sources and this report after cloning
the development branch, but cannot recreate missing ignored raw packets or renew
the spent allowance. A future execution needs its own explicit scope.

Approved bounded execution overrides generic extra-RED/full-project/reviewer/
rerun/cleanup workflows. No new product code was changed and no extra checks
outside that budget were launched. Historical raw records are retained.
No I1 closure, main merge, C12/security/runtime/release/PS5/performance acceptance.
