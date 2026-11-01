# Saved-data checker — confirmed framing root cause

Status: CAUSE_IDENTIFIED; IMPLEMENTATION_UNCHANGED; NOT_ACCEPTED.
Scope approved by user “是”: locate cause only, not fix or retry the suite.
Baseline90043c58d8506302815010527122e4a05d68106d.

## Corrected source location

The failed suite's saved stack names saved-data-contract.ps1:44. Numbered
source inspection confirms this is the raw framing guard, NOT the JSON parser
catch (line53). The previous report/ledger's parse-boundary description was an
author interpretation error. Original failed outputs and historical reports
are retained; the two failures are not regraded.

## Minimal evidence

Two bounded in-memory diagnostic commands, no helper import/invocation:

1. Read only the original first row and call JsonDocument.Parse directly,
   retaining any actual exception; also select its exact string overload by
   reflection as a control. Both parsed Object successfully in PS7.6.5 /
   .NET10.0.11. The row was System.String,1436characters. Source/data hashes
   before/after matched. This does not validate any later checker predicates.
2. Inspect only the raw framing conditions and a one-character ASCII control.
   Original15228bytes start with hex7B22536368656D61 / codepoint123 (`{`), not
   UTF8 BOM. Raw CR absent, one terminal LF present, double terminal LF absent.
   Under zh-CN, StartsWith(U+FEFF) using default comparison returns true on BOTH
   the original text and the literal `{`. Ordinal comparison returns false on
   both. The original combined framing guard evaluates false.

The default culture-sensitive prefix comparison treats this BOM character as
ignorable in the observed runtime. Negating that result rejects BOM-free valid
input before parsing. This accounts for both positive-control failures and
means the negative tests' PASS results did not establish downstream coverage.
This is checker code behavior, not evidence of original data corruption.

Actual command inputs and complete tool returns are preserved in
[root-cause-evidence.json](../diagnostics/isolated-fuzz/p2b-saved-data-contract-revision/root-cause-evidence.json).
No saved command/probe/capture, validator, receipt helper or42group suite was
executed. No source/test/oracle/old archive changed; no diagnostic process,
target, compilation or reviewer. Formal assessment remains unattempted.

## Recommended repair boundary — not implemented

Replace culture-sensitive BOM detection with an exact byte/codepoint or ordinal
check; keep rejecting genuine UTF8 BOM. Add an independently literal regression
control for BOM-free ASCII under the affected culture and a genuine BOM control.
Keep existing framing/type/duplicate/source checks and failed evidence intact.
Any future test amendment must be versioned separately from frozen originals;
do not silently edit the frozen42group model or retrospectively claim coverage.

Future execution should have a separately agreed finite RED/GREEN/assessment
scope. This diagnosis does not promise later predicates pass after the framing
fix. Production I1/I2 and compound receipt I/O remain unverified; integration
6/7/1/conditional1 remains FROZEN. No main merge or release.
