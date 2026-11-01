# Saved reference data readonly recheck — actual stopped result

Status: FAILED / STOPPED_READONLY_COMPUTATION. No original verdict changed.

Executed the [approved zero-host plan](../superpowers/plans/2026-11-01-offline-error-reference-readonly-recheck.md)
after explicit user approval “是”. Baseline
5baa34513d151a2231d3e2a6f8952b4cb3671ebc, existing development branch.
Plan SHA256 BAB169EA984DF804F06BB4607E249CF04F4686CE3373CA70DF4653B60A71D9BA.
No new diagnostic/capture host, target, compilation, reviewer, runtime import
or execution of saved parent/probe/capture/snapshot. No original source edits.

## Completed preflight and static work

Original16files/9directories matched decoded original archive bytes/hashes,
and8original bound inputs matched the original binding. Original archive SHA256
remains the preflight-verified
A12F16F6F42A8365B0D7D34BE8D70771BBDF9FAD96F5FC43EBFDC0B8B37E74D6.
This was scoped diagnostic-data provenance, not completion of the earlier
failed integration full-pin audit.

Authored a new inline readonly computation; no executable source file added.
Static AST/text comparison confirmed both original validation helpers unchanged
and only the expected-label singleton-array assignment changed inside the
4505→4509character saved-data predicate segment. Other changes were explicit
readonly input/result plumbing. No Process/import/ScriptBlock invocation or
ReferenceEquals on JSON; only a CreateNew write to the new result was permitted.
Tool-input SHA256:
2481EF7A889793239B80D690705569990325ABE4CFAD069179085DECF4E7EDBC.

## Actual sole computation failure

Created one consumed intent and performed exactly one readonly computation.
It exited1 with System.Management.Automation.SessionStateUnauthorizedAccessException,
ErrorId VariableNotWritable:
`Cannot overwrite variable Error because it is read-only or constant.`

The failure occurred at the preserved original predicate assignment
`$error=$v.ErrorProjections[$j]` (new computation line50). PowerShell variable
names are case-insensitive; this conflicts with the automatic `$Error` variable.
The original saved parent checker contained this additional naming defect.
The original run had stopped earlier at its label comparison, so it had not
reached this assignment. The new array correction let processing advance here,
but did not establish that every remaining original predicate was valid.

New result.json explicitly records Scope=SavedDiagnosticDataOnly,
Outcome=FAILED, CasesChecked=0 and RecordedMeasurements=[]; original parent
Exact=false/exit1 and original probe exit0 remain unchanged. No complete case
was accepted by this recheck. This is an author-written checker defect, not an
environment/quota/platform or display issue. No variable rename, further
predicate correction, rerun, regrade or fallback success file followed.

OriginalsUnchanged=false in this NEW result means its post-computation original
byte comparison was NOT REACHED. It is not a detected original-file mutation.
The computation had no write to originals; the successful preflight comparison
is separate evidence. Do not claim a successful post-check or complete data
contract validation. This report also does not replace the incomplete older
integration audit.

## Preserved outcome and limits

NEW diagnostic/capture hosts0/targets0/compiles0/reviews0. Original diagnostic
single-host allowance remains exhausted; integration6/7/1/conditional1 FROZEN.
I1/I2/C1/C2 not closed, five capability flags remain false; no runtime acceptance,
production change, main merge or release. Original source/command/results were
not written by the recheck; their failure records must remain authoritative.

Ruling: approved Native/PowerShell readonly scope replaces skill-default
setup/Bash/TDD/full-suite/task-done reruns, agents/review/fix/deletion. Cost:
only saved-data assessment, not fresh runtime or live-reference proof. Actual
computation failure invoked the stop rule, not an automatic correction pass.

New portable-readonly-recheck.json preserves8actual recording-sidecar files/
1root directory: approval, ledger, preflight/static/tool input, consumed intent,
actual computation return and failed result. Decoded lengths/hashes/bytes
verified against these new originals. Archive SHA256:
844D976FD867DDCD9899595E7720E07061E703328FD11358A333D516DEF447F1.
It references the original16file archive; it does not replace or relabel it.

Further work needs separate written scope and approval. Before another attempt,
review the entire saved-data checker for reserved automatic-variable names and
singleton/array/type handling, rather than assuming one label correction proves
the rest. Any permission to rename this variable must be explicit alongside
the already-defined label correction; this stopped result grants none. Do not
execute a diagnostic host, change frozen production/reference contracts or
release integration budgets as a consequence of this data-only failure.
