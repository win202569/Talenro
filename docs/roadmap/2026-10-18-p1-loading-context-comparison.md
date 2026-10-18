# Loading-context comparison — Local timeout, Inherited frozen

User approved the short in-chat design after the [loading audit](2026-10-18-p1-loading-scope-readonly-audit.md):
two NEW independent ordinary PS5 hosts, Local first and Inherited only after full
success, each30000ms, same original product/full order/assertions, boundary-only
timing and resolved-function provenance. Any failure/timeout stops the experiment.
No retry, warmup, product edit, internal product counter, reviewer or old-budget
reuse. New allowance is now1/2 spent, with the Inherited slot unrun/FROZEN.

## Portable diagnostic and static checks

Added six scripts and README under
`docs/diagnostics/isolated-fuzz/ps5-loading-context/`. No product copy exists in
this folder. Both variants use exactly these same files and the same dedicated
P1-LoadingCheck4096 helper; only child environment TALENRO_LOADING_VARIANT changes.
Local dot-sources the unchanged parent product inside that helper; Inherited
would retain the earlier load. Original caller arguments, expected result,
P1-Equal/P1-Need assertion and4097 refusal remain. No actor/deadline/security
checks were removed. Six source reversals restored original scripts after
removing diagnostic additions/relative-path adjustment and blank-line normalization.

Exact complete merged inventory was frozen before launch:83 original ordered
lines/64 PASS markers plus14 LOADTIME events and three LOADBIND records,100 lines
total. Expected timing scopes:InvalidInputs,Build4096,Check4096,Setup,
Evaluate4096,Assert4096,Check4097. Binding positions:caller-before,helper-active,
caller-after. Each records13 product command names/source paths/body hashes and
FunctionInfo/ScriptBlock reference comparisons. Provenance work is outside the
Evaluate4096 interval but inside broader Check4096; overhead is uncalibrated.

## Local result — failed/incomplete, NOT PASS

- Started UTC2026-10-18T08:51:43.2398746Z; ordinary PS5 5.1.26100.9711 Desktop x64.
- Configured30000ms; actual30081ms including81ms termination/collection overhead.
- CPU31546.875ms (all process threads); timeout=true,exit-1,owned-host-exited.
- Stderr empty;75 nonempty output lines:65 original lines, eight timing events,
  two provenance records;48 PASS markers, last P1PassiveDescriptors.
- CaptureValid=true means valid prefix only;ExpectedResult=false and
  FormalAcceptance=false. No assertion error was saved before termination.
- Closed pairs below;Check4096 and Evaluate4096 remain open. Next expected record
  is Evaluate4096 end. Assert4096,caller-after,Check4097 and subsequent groups were
  not observed. No tail reconstruction or repeated execution.

| Completed boundary | Wall ms | Process CPU ms |
|---|---:|---:|
| InvalidInputs | 465.213 | 468.750 |
| Build4096 | 980.147 | 937.500 |
| Setup (conditional local reload) | 27.982 | 31.250 |

Do not sum these with inclusive parent timings, equate process CPU with a CPU
percentage, infer total evaluation cost from an open scope, or treat the last
PASS as the exact termination instruction. Unlike the original-entry-only
prefix, this diagnostic DOES show Evaluate4096 started and did not finish.

## Provenance result and inference limits

All13 helper-active commands resolve to the original parent product file and
have the same definition hashes as caller-before. All13 FunctionInfo references
and all13 ScriptBlock references differ from caller-before; initial self-reference
comparisons were true. Thus fresh local command definitions were observed before
the ordinary named product call, rather than merely assuming reload from source.
These observations do not attest internal compiled-cache identity or every
nested runtime command dispatch. Caller-after was never emitted, so post-return
restoration/isolation is UNVERIFIED.

Local reload of the unchanged product did not suffice for completion in this
run. This strengthens the counterexample to a reload-alone guarantee, not proof
that scope has no effect. Inherited was NOT launched under the stop rule, so
there is no completed A/B comparison or measured relative speedup. Successful
historical internally instrumented copies remain confounded by modified bodies,
instrumentation and uncontrolled runtime conditions; none is a product fix.

Independent read-only audit matched every saved line to the frozen merged prefix,
recomputed all three timer pairs, checked two open scopes and26 binding entries,
13 changed reference pairs, hashes, engine/controller and all283 pins. Frozen
inventory inherited270 pins, added five preceding original-run anchors, seven
new portable files and the loading audit. Prior artifacts/verdicts unchanged.
There is exactly one invocation and no Inherited stream. No extra diagnostic
hosts were used for static verification or this audit.

## Integrity and recovery

Original product remains
`953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA`.
The copied top-level entry remains byte-identical to original entry
`12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422`;
its directory resolves copied child tests, so that hash alone does NOT imply
original-entry acceptance. Original source/tests were not edited.

Ignored local packet `.superpowers/sdd/2026-10-18-p1-loading-context/` retains
controller/scope/frozen/index/raw. It is not committed or required to clone.

| Artifact | SHA256 |
|---|---|
| frozen.json | 98074465253FFE629AFFB12EE969EC5A0F3685487BCBF3804293A79F93670C82 |
| run.ps1 | BF9E3360CE7CD2FCD8398EE3634A1C8EC93531840D6A012EBE7EF0DFF41F1C84 |
| invocations.json | AEC92CF5B70A093D47CB7EA7C3FAD99B80D8743D8C004403189E74559EBB980C |
| Local.stdout.txt | 0619453C87717C1D59D30CAEEE1E7094779F46EFB19FCE83ED973EE7A67F2758 |
| Local.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Original35 diagnostic tests | FE2A414974D133373FCC8996762637A02D65535173AF3ECD91A96238A643CD2D |

Engine remains60C9B298 (full hash in preceding original verification report).
Recover portable scripts, this result, current status and recovery runbook from
codex/c12-b01-task9-coordinator. Cloning restores neither ignored raw/cached
engines nor replay authority. The remaining Inherited slot stays frozen;
do not invoke scripts directly to bypass that stop.

PS5/I1/P1/candidate acceptance remains blocked;all older spent/frozen allowances,
review0/1 unused/ineligible and main/C12/security/runtime/release gates unchanged.
Next recommend a read-only hot-loop cost model of repeated reflection/member
inspection in the unchanged evaluator before proposing any further experiment
or optimization. No new run, cache/security bypass, larger timeout or product
change is authorized by this report.
