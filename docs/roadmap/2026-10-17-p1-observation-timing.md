# ObservationScalars isolated timing — diagnostic complete, not acceptance

User approved one fresh ordinary-token PS7 child <=30000ms after64e6872a. Only
the equivalent ObservationScalars path, with per-item start/end timing; no full
suite, fixture preflight, warmup, old R3 replay, candidate/test edits, retries,
fallback, reviewer or old-budget use. Probe is retained as a throwaway diagnostic,
not production or a replacement for complete-module verification.

## Observed result

Sole host PS7.6.5 Core x64/admin=False: exit0,1691ms total launch/wait/drain,
TimedOut=false, owned-host-exited, stderr empty. CaptureValid=true,
Disposition=Complete. Exactly65 start/end pairs, final P1I1ObservationScalars PASS,
summary Started=Completed=Expected=65 and final LIMIT. No active item or truncated
tail. Probe-local elapsed1033ms; first start at54ms, last end at1032ms.

| Per-item core elapsed | ms |
| --- | ---: |
| Sum of65 items | 890.762 |
| Minimum | 6.244 |
| Median (33rd sorted observation) | 10.887 |
| Mean | 13.704 |
| Maximum | 103.598 |

Slowest completed item: id1, arm/index0/Sequence,103.598ms. Next slowest id2
arm/SessionLabel28.990ms and id10 arm/HandleInheritanceBlocked24.912ms. These are
one-run observations, not stable benchmarks or a proved cause for first-item cost.
Times cover decoration, product decision and original result assertions inside
the renamed core. Scenario construction outside that call, start/end JSON logging,
module import and process startup are not all included in item times; sum is not
total process time. Instrumentation adds overhead and changes call context.

65 checks retain original path: normal arm11/release7/progress13/cleanup9/persist8,
then control9/fault7 and final cutoff/Classification1. Same literal refusal
expectations/type/value/metadata guards, no weakened assertion. Hash-table payload
field ordering can vary within a phase; frozen validation permits each expected
field once in that phase, with ordered phases and consecutive ids.

**Conclusion:** this independently executed path completed quickly in a fresh
host; no long per-item stall observed. It does not recreate the old mixed-suite
state, retroactively identify the old timeout statement, or prove cumulative
overhead as the unique explanation. The old full R3 still timed out/false. Four
other uncompleted ownership groups remain unverified; this diagnostic is not
a continuation/completion of the frozen R3 or a GREEN acceptance receipt.

Read-only source inspection notes original ArrayAndIntegerBounds constructs and
evaluates4096 observations. That is a reasonable future timing candidate, not
measured evidence of a bottleneck here. No original-suite group was rerun for
that hypothesis. Next scope phase timing/localization before choosing any
optimization or verification partition; do not simply increase the30s limit.

## Fidelity, transport and recovery

Static parse passed for probe/controller. AST/text checks matched eleven copied
helpers to ownershipC490 and exact ObservationScalars command. Original
CheckDecorated definition renamed to ObservationTiming-Core only; all its body
and assertions unchanged. A wrapper forwards identical arguments and emits
timing, capped65 calls. Other helper definitions unchanged. Original test script
was not imported/executed; productCF4C was imported for these65 diagnostic calls.
No other case/preflight was executed. Source and original test files unchanged.

Frozen manifest inherits86 historical pins plus14 extended-view packet anchors
and two current product/test pins:102 protected entries. Controller/manifest/probe
hashes, phase/field/event/summary inventory, one-host cap and timeout policy fixed
before launch. Slot consumed before Process.Start. Ordinary token, hidden no-
profile child, no shell, separate UTF8 streams and owned-handle cleanup preserved.
The planned BudgetExhausted/partial-tail parsing paths were not exercised. No
timeout/cleanup guarantees can be inferred from this fast normal exit.

Local ignored packet `.superpowers/sdd/2026-10-17-p1-observation-timing/` retains
scope, frozen controller/manifest, invocation index and raw stream pair; no prior
packet deleted or rewritten. Portable probe/report/status/recovery are in Git;
clone does not restore raw/runtime state or authorize a repeat. Hashes anchor
saved bytes, not independent historical authenticity.

| Artifact | SHA256 |
| --- | --- |
| Portable probe | `9099B9C9A68C467AE952DF409038F65FD00A1035CC353D9CD1F7883B39903396` |
| frozen.json | `7F2CC9E645FB1A33775CF984B1BC52605682F2BEDE48CBAF8FD085F399D96418` |
| run-once.ps1 | `BA1FE1D993D2427DE38C78B4F3DDC63B86081FA6B3C5F94857141D5091CF01B8` |
| invocations.json | `5E1E69218DEC7007C0C01D927A4752A0AAF744ACD17CC2BE035E73E9934D506F` |
| D1.stdout | `38E33410E2AE2BA9505126C8868880D5CD2EF1F48AE48F27C520037C6CDD95CF` |
| D1.stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

New diagnostic1/1 spent. Extended-view repair remains3/7/four frozen/review0/1
ineligible, old R3 false preserved. ProductCF4CB4B5/ownershipC4905925/original35
suite unchanged, I1/P1 unfinished. No main/project-wide/P2/P3/live-supervision or
C12/release acceptance; no subsequent profiling/repair/retest/reviewer authorized.
