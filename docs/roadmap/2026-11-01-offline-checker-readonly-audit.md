# Complete saved-data checker — readonly audit

Assessment: NOT_READY_FOR_ANOTHER_ONE_LINE_RETRY. Two actual execution
blockers and four additional static contract/recording findings below.
No source edits, checker/probe invocation, data mutations, diagnostic hosts,
targets, compilation, agents or independent reviewer. This is author readonly
analysis, not the still-frozen conditional independent runtime review.

Baseline4573a3a881ef3f03ff58c5e66b58422742f6296b. Scope: read the COMPLETE
original parent tool-input command (A), COMPLETE readonly recheck command (B),
probe source, actual failed results, and original diagnostic/recheck contracts.
AST parsing and assignment/location inventory only; no mutated-data evaluation.
Existing failed results remain failed, and post-byte checks remain uncompleted.

Locations A/B below are logical1-based lines INSIDE the JSON cmd strings, not
physical JSON-file lines or an attempted rewrite of actual error stack numbers:

- A: original diagnostic packet/parent-invocation.tool-input.json.
- B: readonly-recheck packet/recheck.tool-input.json.
- Both are preserved in the committed portable archives; neither was executed.

## A1 — Actual blocker: reserved automatic variable assignment

Important; actual failure reproduced by the already-saved failed recheck,
not by this audit. A82/B49 assign `$error=$v.ErrorProjections[$j]`, conflicting
case-insensitively with automatic `$Error`. The actual exception is
SessionStateUnauthorizedAccessException / VariableNotWritable. AST parsing
still reports0syntax errors, demonstrating that syntax-valid is not safe to run.
The assignment scan identifies the same conflicting direct assignment in both
commands. No additional direct assignment to Error was found outside these.

Revision: use a task-specific descriptive variable such as recordProjection,
update its uses together, and lint direct assignments/parameters/loop variables
against reserved automatic/read-only names. Keep writable ErrorActionPreference
distinct; do not bypass read-only protection or clear/replace the automatic Error.

## A2 — Actual blocker: singleton expected-label unrolling

Important; original parent failed at its event comparison. A74 assigns the
conditional's single emitted Outer label to a scalar string; A80 indexes it
as an array. B41 now preserves an explicit string array, but B still failed
later at A1. Thus the label correction alone never established a complete pass.

Revision: preserve zero/one/many sequence shape explicitly at its construction,
and declare expected labels once as fixed data. Inspect similar assignments
without mechanically treating every singleton as a bug: the probe's join-based
label comparison does not use the same scalar-string indexing path.

## A3 — Static finding: complete shape/type contract is not explicit

Important; A75–100/B42–67. Some fields have proper numeric/boolean checks,
but Events,ErrorProjections,matrix outer arrays/matrix rows and SecondaryIndices
lack explicit array-type requirements. ObservationIndex, PrimaryIndex,
secondary index values, and COMPLETE Schema/Cases/FaultCalls are compared by
value without the same numeric-type assertions used on case Schema/FaultCalls.
Count/index operations and PowerShell value comparisons are not explicit JSON
shape/type contracts. The two matrices are individually checked but the
same-record-implies-same-captured-Exception relationship is not checked.

Revision: independently specify exact array/map/null/string/integer/boolean
types, ordered keys and dimensions; validate indices before use and final-row
types as strictly as case rows. Keep permitted off-diagonal measurements
unpredicted, while checking universal observer invariants. This does not justify
forcing the four rethrow outcomes to previously observed values.

No scalar/string/bad-shape variant was evaluated in this audit. Missing guards
are source facts; the full set of variants currently accepted/rejected is NOT
dynamically established. Do not describe proposed counterexamples as observed
false passes, especially while A1 prevents later predicates from running.

## A4 — Static finding: error-source attribution is substring-only

Important; A85/B52 accepts Position containing probe-path-plus-colon and Stack
containing Invoke-ODReferenceFault. It does not parse the actual file/line,
require a numeric valid line, or bind it to the frozen ReadAllText statement.
The fixed Type/Message/ErrorId checks are useful but do not fill this gap.

Revision: validate source location against the frozen source as data, actual
helper AST extent and bounded valid line/call-chain rules. Keep explicitly
permitted formatting variability documented. Do not require a new I/O fault
or execute the probe to parse source metadata. No forged-position data test
was run; this is an attribution predicate gap, not proof of original forgery.

## A5 — Static finding: trailing blank JSONL records are normalized away

Important for the stated exact7row contract; A63–65/B30–32 first permits any
LF-terminated text, then TrimEnd(LF) removes ALL trailing LFs before splitting.
Consequently the predicate has no exact-one-terminal-delimiter check and can
discard extra empty trailing records before counting. This follows from the
string operations; no original bytes were appended or mutated to demonstrate it.

Revision: validate raw framing before parsing/normalizing: exactly one final
LF, no CR/BOM/blank/extra row, strict UTF8 and byte bound. Remove only the one
required delimiter. Preserve original raw bytes and do not rewrite an input
to make it conform.

## A6 — Static finding: result-write cleanup can replace write/flush cause

Important; A6–9 Parent-JSON and B88–90 write/flush in try, then call Dispose
in an unprotected finally. If operation and cleanup both fail, cleanup can
replace the original operation error. This repeats the first-cause hazard being
investigated, now in control-record persistence. Probe cleanup, in contrast,
retains its first failure and only adopts Dispose failure if none exists.

Revision: keep a first actual ErrorRecord locally, dispose the held stream once,
retain bounded cleanup detail separately and propagate the original failure.
No retry/reopen/alternate success file. This path was NOT fault-induced here;
actual archived writes succeeded, and source inspection is not dynamic proof.

## Cross-cutting reporting and preserved strengths

The new result's OriginalsUnchanged=false conflates “comparison not reached”
with “comparison found mismatch” for machine readers. Existing reports correctly
explain the ambiguity. A future result contract should separate NOT_CHECKED,
MATCH and MISMATCH; never retroactively change this recorded false value.

Preserve working boundaries: fixed six-case/fault inventory, strict UTF8/byte
bound, complete ordered dictionary-key checks, explicit event-count integers,
boolean matrices/diagonals/symmetry, reference-based collector-count consistency,
independent fault control, false capability flags, original raw retention and
new-result versus original-failed-parent separation. Parent retains the exact
Process object, does not search PID/name or kill a tree, and uses finite waits.
No new dynamic lifecycle guarantee was tested by this audit.

## Unified revision recommendation — not approval to implement

Do not copy the current predicate block verbatim again with one rename. First
write one consolidated data-contract revision covering A1–A6 and tri-state
provenance reporting. Separate a pure saved-data validator from process launch
and record persistence so source linting and future in-memory data tests can
exercise it without a diagnostic host. Define valid originals and targeted
zero/one/many/type/framing/location/cross-matrix negative cases BEFORE running
the checker. Any such tests/implementation are future work requiring approval,
not experiments performed in this turn.

Keep all existing sources, originals and failures unchanged. A future corrected
data assessment remains a separate limited verdict, not original Exact=true,
live-reference reconstruction, I1/I2 production repair or budget restoration.
Original integration6/7/1/conditional1 is FROZEN, diagnostic allowance exhausted;
earlier incomplete full-pin audit remains incomplete. No production/main scope.

Limits: no full8input/16file byte audit repeated here, no checker/probe rerun,
no mutations/fault induction, no semantic audit of unrelated project paths and
no independent reviewer. AST/source review cannot guarantee the absence of
additional runtime defects. The committed actual failures remain authoritative.
