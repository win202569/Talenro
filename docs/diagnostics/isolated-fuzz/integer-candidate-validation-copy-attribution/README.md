# Validation/copy attribution — PREPARED, NOT RUN

Derives from ../integer-candidate-ownership-markers/. Six PS1 sources unchanged.
Only phase-evaluator.ps1 adds instrumentation and a fresh unique entry name;
original35 tests change only that single caller command. Remove eight vcDiag
lines and restore entry/caller names to recover prior files exactly.
All guards,fixtures,assertions,copy/list operations and order remain in place.
No shared helper/global state,per-row output,wrapper,finally end or second pass.

Six typed local longs initialize after RootSetup END,before ValidationCopy START.
Each row samples t0 before original failure-index/input assignment,t1 before
ordered snapshot/copy/list Add,t2 immediately after Add. Accumulate t1-t0 and
t2-t1,then completed rows,all after t2. One invariant ATTRIBUTION summary follows
normal ValidationCopy END,before Decision START. Exceptions do not fabricate it.
No partial row count is available if this region times out before its summary.

expected-output.json has106 ordered nonempty records/64 PASS/22 timing markers/
11 pairs;summary is not a timing pair. The distinct offline-only
../candidate-ownership-attribution-capture-verifier.ps1 validates exact order,
ordinary context/PS7/typed receipt facts,all original tick/nesting/exit/budget
gates,summary row4096,Int64 values and exact sums <= closed ValidationCopy span.
Missing summary is null/unavailable,not zero. Valid later partial prefixes may
retain available totals without a whole-run hypothesis. Invalid captures publish
no totals. FormalProjectAcceptance always false;this is not authenticity,
process provenance,security proof or runtime acceptance.

H-Attribution on a complete diagnostic:validation sum > copy sum supports only
this instrumented ordering;copy >= validation falsifies it. No classification of
partial/invalid captures as a completed experiment. No individual helper/pure CPU
or original-run cost attribution;timer/assignment work lies asymmetrically in
spans. Three samples per complete row means12288 extra reads at4096 rows.
Totals omit loop/accounting work;parent-minus-sums is merely unattributed time,
not calibrated overhead. New name/source/body/history and GC effects can alter
behavior/timing,including a timeout beyond the unchanged30000ms cap.

Synthetic tests:
../candidate-ownership-attribution-capture-verifier.tests.ps1
Static read/AST fidelity only:
../candidate-ownership-attribution-preparation.tests.ps1
Neither executes/imports target sources. Old47/63/76/122 offline verifier suites
and20 ownership static checks remain distinct and unchanged.
This folder has11 files including its own LF .gitattributes;the14 old artifacts
and their pinned attribute rules are not edited. New utility names inherit the
existing candidate-ownership-*.ps1 LF rule. No ignored raw/controller dependency
is required to reproduce offline preparation checks on another computer.

No controller or target run here. Any later execution requires separately
approved NEW ordinary fresh PS7 single-host/30000ms/no warmup/retry/fallback
scope,prelaunch identity freezing and bounded owned-host cleanup.
All prior budgets/PS5 freezes/review gates remain unchanged;no product promotion.
