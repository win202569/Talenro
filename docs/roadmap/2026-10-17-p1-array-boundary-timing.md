# Isolated4096/4097 timing — large valid evaluation dominates this run

User approved one fresh ordinary-token PS7 child after19ef2fe1: at most two
product calls, total<=30000ms, separately time build/evaluate/assert for4096 and
4097 observations. No full original suite, fixture preflight, warmup, prior R3
replay, product/original-test edits, retry, fallback, reviewer or old-budget use.
Retained probe is throwaway diagnostic tooling, not production or a regression
suite replacement. Original boundary expectations/equality assertion retained.

## Observed timing and assertions

Sole PS7.6.5 Core x64/admin=False host: exit0, ElapsedMs22447 including launch,
wait and stream collection; TimedOut=false, owned-host-exited, stderr empty.
CaptureValid=true, Disposition=Complete. All six start/end pairs present, exactly
two product calls, final diagnostic PASS/summary/LIMIT;17 ordered stdout lines.
No active phase or truncated tail. Probe-local elapsed21782ms.

| Phase | Elapsed ms |
| --- | ---: |
| Build4096 | 161.528 |
| Evaluate4096 | 21253.172 |
| Assert4096 | 158.405 |
| Build4097 | 54.802 |
| Evaluate4097 | 47.721 |
| Assert4097 | 42.582 |

Evaluate4096 consumes about95% of total process elapsed in this one observation.
Build phases also construct literal expectations and check input lengths;
assert phases compare full expected results. Phase elapsed excludes marker JSON
logging before/after the timed body. Setup/import/process startup and logging
are not all in phase sums. This is one instrumented fresh-host observation,
not a stable benchmark or cross-host timing guarantee.

4096 input is the original sequence0..4095/time/AtMs0 path. Full literal prepared
result expectation retains LastAtMs0/LastSequence4095 and original defaults.
4097 input appends the original sequence4096/time observation; full literal
invalid-scenario Refused result expected, Evaluated=false/Decision null and
unchanged false authority flags. Both comparisons passed. Input-length checks
are added diagnostic preconditions, not substitutes for original assertions.

**Conclusion:** valid4096 evaluation is a measured major cost contributor in
this isolated boundary path; array building, final assertions and4097 rejection
are much smaller here. Together with the previous65-check isolated1.691s path,
this supports investigating cumulative full-suite work. It does not identify
the exact old R3 executing statement, prove a unique old timeout cause, or
attribute cost to any particular function/member getter inside evaluation.
No optimization, extra timing run or larger timeout was attempted.

## Fidelity, integrity and limits

Static parsing passed. AST/text checks match five helpers to the unchanged
original35-case source: Need, New-P1Scenario, Observation, Want and Equal. The
exact original P1-Check equality assertion is inside BoundaryTiming-Assert.
Evaluation/assertion were separated only for timing; same original boundary
input/expected values and comparison logic. Other ArrayAndIntegerBounds invalid
type/array/shadow-length cases were deliberately not run, nor the full named
group or suite. No original test script was imported/executed. ProductCF4C was
imported and called twice. Source/tests remain unchanged, not edited for timing.

Controller froze phase/row/call-counter/event/summary inventories, byte hashes,
one-host/two-call caps and30s budget before launch; slot consumed before start.
Source/evidence integrity inherited102 pins plus five prior packet anchors
and portable ObservationScalars timing probe:108 protected entries. Original
false R3 and every older raw/manifest preserved. Ordinary hidden no-profile
process, no shell, separate UTF8 streams and owned-handle cleanup preserved.
Timeout/partial-tail/error-cleanup paths were not exercised by this completed
run and are not claimed validated. CaptureValid is diagnostic transport status,
not whole-module acceptance.

Ignored local packet `.superpowers/sdd/2026-10-17-p1-array-boundary-timing/`
retains scope/controller/manifest/index and raw stream pair; no packet deletion.
Portable probe/report/status/recovery in Git enable inspection on another
computer, not restoration of ignored raw/runtime or permission to rerun. Byte
hashes show saved consistency, not independent historical attestation.

| Artifact | SHA256 |
| --- | --- |
| Portable probe | `288F74EF7E64A72CD9C625D123444FAC11BB7528F10DF2DABAD2EDDA24C782DE` |
| frozen.json | `B7DC4A65C56929F39AC5047BE3C5C9B99D2970B991B9C8CBB258E2644F9E0D4F` |
| run-once.ps1 | `01BD2872C38E2E2C22D3AF1250B0FF90744F9297CC7EA89AEED622386185C48A` |
| invocations.json | `35E1A8EE6278F0502CFE88E075E908CD65A398F41979B04653E427663187D45B` |
| D1.stdout | `4EE6A2F0B831A6C0ADE6751B4D70E7452C9646DB5346C7D1B9587AE4FDAC46E5` |
| D1.stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

New diagnostic1/1 spent. ProductCF4CB4B5/ownershipC4905925/original35 unchanged;
I1/P1 still unfinished. Old full R3=false and extended-view3/7/four frozen/review
0/1 ineligible preserved; no current diagnostic substitutes for those missing
whole-module/mutant/final runs. All other earlier budgets remain unchanged.
Next scope performance attribution or a semantic-preserving optimization, with
explicit verification allowance and original assertions/30s gate retained. Do
not silently resume R4 or raise limits. No main/project-wide/P2/P3/live/runtime
or C12/release acceptance follows from these two passing boundary calls.
