# Complete-flow timing read-only audit — remaining work is not removed by shape lookup

After8ae4f271, user requested continuation. This turn performs ZERO product/test
executions, no diagnostic child, no product/test/controller edits and no budget
spend. Only portable findings/status/recovery documentation is updated.
This is coordinator static analysis, not independent review or acceptance.

## Difference between the passing and blocked paths

Latest targeted PS7/PS5 receipts cover shape6, fixture8 and ownership10.
The complete shape wrapper instead calls shape6, reflection3, then ownership
Group=all: fixture8, original35, and ownership10, in one host process.
Thus targeted success does NOT include original35/ArrayAndIntegerBounds.
It also does not establish combined execution cost, cross-stage runtime behavior,
or an order-independent sum of separately observed runtimes.

The unchanged ArrayAndIntegerBounds group builds4096 six-field time records,
validates/evaluates the complete array, compares the returned result and checks
4097 refusal. Current product validates and copies all input records before
the separate decision loop. Nothing in shape-set reuse removes these loops.

## Source-derived operation accounting, NOT a new measurement

For one valid plain six-field time record on the current public path:

| Site | Type helper calls | Scalar extended-member checks |
| --- | ---: | ---: |
| Plain observation type | 1 | 0 |
| Kind token, including passive check | 2 | 1 |
| Shape object type plus six passive values | 7 | 6 |
| Sequence and AtMs integer checks | 4 | 2 |
| SessionLabel and ClockLabel type checks | 2 | 0 |
| Actor token, including passive check | 2 | 1 |
| Total per record | 18 | 10 |

For4096 valid records:73728 type-helper calls and40960 scalar extended-member
checks in observation validation, excluding root/array setup. This is static
path accounting with fixed literal input assumptions, not measured counters
on the current candidate. It does not apply unchanged to other kinds, malformed
values, early refusal, wrappers or the full regression suite.

Historical instrumented single-evaluation report recorded73746 type calls and
40969 extended-member calls INCLUDING root setup on an older product copy.
The totals are compatible with this accounting; they do not constitute a fresh
performance measurement of candidate953CC2A8.

The shape optimization replaces per-property scanning of fixed allowed names
with hash membership; it still performs the passive/type/member checks above.
For time's six admitted names, successful old sequential lookup makes21 name
comparisons versus six current set membership calls, plus current cache work.
This comparison is not a runtime cost estimate or proof of net speedup.
Both validation and decision loops remain linear for fixed schema sizes;
no newly established quadratic input-size behavior or permanent hang was found.

Removing repeated defensive checks is NOT recommended from these counts.
Counts do not identify which calls dominate elapsed/CPU time and current checks
also enforce descriptor/scalar ownership boundaries.

## Why existing timing cannot establish one cause

Older combined group-only diagnostic (different candidate6CF0D354) completed in
29783ms; ArrayAndIntegerBounds took19568.840ms as a WHOLE group, about65.7%
of process elapsed. This includes construction, invalid cases, evaluation and
comparison, not evaluation alone.

Its subsequent subphase diagnostic timed out at30080ms, but reached the boundary
at18181.547ms versus5823.606ms before. Slowdown already preceded the large case.
It saved an Evaluate4096 START only, not a completed evaluation duration.
A later ordered prefix pair did not reproduce a large cost from three extra
helper conditions. Separate runs changed instrumentation and runtime conditions;
none supplies a controlled unique explanation of the current full timeout.

Latest current-product full timeout saved original35 completion and ownership
through PassiveDescriptors, so the4096 case did finish in THAT run. The last
completed PASS is not the exact statement active at termination.
Latest targeted ownership completion independently confirms last-three semantics,
but neither completion locates current full-run costs.

Current source and retained evidence support an aggregate work/variability
investigation, NOT another speculative product optimization or timeout increase.
Historical aggregate CPU cannot distinguish host wait, runtime work, JIT/GC,
validation, assertion, logging or which group consumed CPU.

## Integrity and remaining state

Read-only recheck matched all183 pins of the last targeted packet, its controller,
two saved invocation rows and four raw stream digests. Index remains:
D5082A89243E16C33E578F8B8558F1E6A1480004A24A96FDFDF28D95CDB53F65.
Candidate953CC2A8 and original tests unchanged. Old records were not rewritten.

Targeted2/2 spent; shape2/3/one frozen; prior independent full1/2/one frozen;
reflection3/4/one frozen; extended-view3/7/four frozen; prior diagnostics spent;
review0/1 unused/ineligible. Full30s acceptance remains blocked; I1/P1 unresolved.
No main/C12/project/live/runtime/release acceptance.

A future bounded diagnostic should collect per-group monotonic wall and
cumulative process CPU at boundaries in the current COMPLETE order, preserving
product bytes and original helper/assertion bodies. Diagnostic copies and their
instrumentation overhead must be explicit; results cannot replace the original
acceptance gate. No such tooling or run has been created/authorized by this audit.
Any new child needs its own approved scope; no retries, warmups, old-slot use or
raised limit are implied. Clone restores this report/source, not ignored
evidence, cached engines or replay permission.
