# New limited26-check matrix — 2026-10-15

## Approved scope and execution model

The user approved one new limited matrix, not a replay/rewrite of historical
matrix3 and not full project/production acceptance. It covers22 complete packages,
authority only `^(TestBoundRead|TestCoordinator)`, store compile only, and original
Envelope/Pair short fuzz checks. Original26 package/selector scopes are preserved
from the October13 execution record §3. No integration build tags, DB test run,
Docker/remote mutation, install, security change, retry or main merge.

New `matrix-check.ps1` adds a separate entry using the existing unmodified native
job boundary. It separates24 builds and25 runtime invocations (49 maximum native
attempts) into26 checks. Builds use600s outer bounds, ordinary tests retain120s
with150s outer bounds; fuzz retains5s/parallel1/test30s/native60s. Builds run at the
project root, binaries at their package directories, preserving relative testdata
and source-guard paths. Store is never executed. Nodeissuer is compiled with the
existing fuzz-instrumentation flags, then reused for its ordinary and two fuzz
checks; no change to Go runtime source or product code.

`matrix-check-001` uses the independently verified unpatched toolchain copy and
fresh offline cache/tmp plus separate fresh fuzz caches. No Go version/listing
query, copied-toolchain recreation or dependency download is part of the matrix.
The existing native wrapper's receipt-parent cwd was unsuitable for package tests,
so a matrix-only adapter passes the approved cwd directly to unchanged native Run
and serializes all public Result fields, including existing shutdown evidence.

Full copied-toolchain manifests and product inventories are checked before/after.
Every built binary, compile receipt and both compile streams are bound; validated
execution receipts and both output streams are also hashed and re-audited at the
end. Immediate binary/compile-artifact checks surround each execution. Ordinary
checks require actual matching PASS test records and terminal PASS; fuzz requires
worker/positive-execution evidence and terminal PASS without diagnostic hooks.
First failure stops the sequence; partial counts, progress and stopped records
remain. Final audits and result persistence precede any streamed success.
MatrixPassed is this new limited matrix only; Acceptance always remains false.

## Tests, defects caught before execution and review

Initial RED confirmed no new matrix entry. The first positive fixture then caught
a PowerShell array/concatenation precedence error that joined two accepted package
names into one string. Explicit per-element parentheses fixed that parser bug.
It was a new driver issue, not a product/Go test failure.

Independent read-only review identified missing final integrity binding of earlier
runtime evidence. A dedicated mutation of check01 output during check02 produced
a genuine RED: the old driver incorrectly accepted all checks. Runtime hashes and
their final audit were added, with compile stdout also bound. An intermediate
placement error referenced `$run` before execution; the fixture caught it and it
was moved after validation. The targeted case then passed. No real Go invocation
occurred during these fixture cycles. Re-review closed the Important finding;
no outstanding Critical/Important/Minor finding remained, conditional on tests.

Portable unit tests replace only native execution/full toolchain scanning and
use fixture-backed product enumeration. The real native C# Result type is populated
without launching a process, so native-property serialization is exercised along
with real wrapper/planner/files/hashes/parsers/completion/finalization logic.
They cover exact package/selector/cwd/budget contracts, no store execution,
existing/wrong identities, first-failure counts, compile/source evidence, timeout/
shutdown/exit, missing tests/PASS, fuzz zero/diagnostic records, binary/compile/
runtime/product/toolchain drift and result-write failure. No machine cache is
needed for these new unit fixtures. Existing suites keep their documented
independently prepared prerequisites.

Final60 assertions PASS on PowerShell7.6.5 and WindowsPowerShell5.1, both exit0;
existing154 scoped assertions also passed on both hosts. Total214 per host.
The narrow runtime-evidence selector passed after its recorded RED, and the
default60-assertion runs included it again. These tests did not execute Go/native
processes; the real limited matrix is a separate single authorized invocation.

## Actual one-shot result

One authorized `matrix-check-001` attempt stopped at check13/statecontract.
Final result: MatrixPassed=false, Acceptance=false, Passed12, Failed1, Unrun13,
Current13/execute, FinalizationFailed=false. There were26 top-level native
attempts:13 successful builds,12 successful runtimes and one failed runtime.
Checks14–26 were never started; no step14 directory exists. Authority, store and
the two dedicated short fuzz checks were therefore UNRUN in this matrix.
Nodeissuer's ordinary check05 did run its normal seed-mode fuzz tests, which is
not the dedicated duration-based fuzz validation scheduled at25/26.

Native window13:09:36.2730028Z–13:16:37.3870516Z (UTC), excluding preflight and
postflight audits. All26 receipts show no timeout, natural verified shutdown and
clean final active0;25 exits0, one exit1. Aggregate counts do not identify individual
processes. Thirteen binary/compile bindings plus12 validated runtime bindings
contain88 files; all were independently rechecked unchanged. The failed runtime's
raw receipt/streams are retained and pinned below, not labelled successful or
included among those12 validated runtime records. Product inventories701/701
Path/SHA256 pairs match exactly. Final copied-toolchain/binary/evidence audit passed.
No retry, source/test change, directory deletion or later Go command followed.

Additional post-run read-only original-toolchain audit passed11536 files/
218649457 bytes,252878ms, completed13:24:47.6712535Z. Reference/copy receipt hashes
were rechecked unchanged:
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55` and
`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.
Original toolchain is unchanged; this read-only audit launched no Go program.

### Failure diagnosis — diagnostic SDK data included by source scan

`TestContractBoundary` failed at `internal/nodecontrol/statecontract/boundary_test.go:345`:
the whole-repository WalkDir excludes `.git`, `.worktrees` and `vendor`, but not
root `.superpowers`. It therefore parses the copied SDK fixture
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/framework-001/toolchain/src/cmd/compile/internal/syntax/testdata/sample.go`.
Line25 intentionally contains an invalid import and an expected-error annotation;
the test reports `expected ';', found ','`. The file is a stock compiler syntax
negative fixture, not project production Go source. Its current hash matches the
original frozen toolchain manifest:
`319E857B9F45FB14457E49356301E827F4426DCEF4DCEB679DFC087735FB5639`.
This was established by saved output and read-only source/hash inspection, not a
second test run. No old PID was queried.

Check13 runtime UTC13:16:17.8762336Z–13:16:37.3870516Z,19510.818ms; exit1,
TimedOut=false, root exited naturally, cleanup verified, no termination requested,
final active0, shutdown0ms, TotalProcesses2 aggregate. The named failing test took
18.70s; stderr is empty. This is a source-scanner scope failure, not the historical
fuzz deadline, a120s test timeout or evidence of a new business-logic defect.
It does not prove business correctness either. The earlier SDK copies remain
intact; they were not removed to make the guard pass.

The current execution approval requires stopping, so no old Go test/guard was
modified and no rerun occurred. Next proposed work requires a bounded new approval:
precisely separate root-owned diagnostic artifacts from production source scanning,
retain unauthorized-consumer negative controls, then separately validate that fix.
Do not silence all parse errors, broadly skip arbitrary production directories,
delete retained SDK evidence or treat the matrix as accepted.

| Retained local evidence (under matrix-check-001) | SHA256 |
| --- | --- |
| plan.json |3A3CA06393327D1BA1E8371D10048290C372B6E4F90DFD2B1E2103B8AF9DA725 |
| matrix-result.json |C33DD9F096DB318BCE5C7491062D6130240A2D78A12AB5C93FAEA979C08653D7 |
| stopped.json |9CE707D60059F8C3B606D2F1A3200C4C35BE30DF2624BB83B0D525C8BD9C312D |
|13-statecontract/build/receipt.json |7E9A77A625354ADFC7F96F5874879A787C381448AADDE45BF0A5AE0817939E09 |
|13-statecontract/execute/receipt.json |DA0E0942DDDCC695B546819F71533805ECA3B5BFAF445461BAC790BD3A72676E |
|13-statecontract/execute/stdout.txt |D666A6D0D81A79847C20A02E90D9FA2AD7FACC4E36DA91D9218D8C1E17B1CBFF |
|13-statecontract/execute/stderr.txt (empty) |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| before/after-product.json |5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |
| original-source-postcheck.json |0F776E86294FBDEFAF3723DF23B42CC1C7260A50732C702DD463D8761E28FDD2 |

Consumed identity001 must not be replayed. The final result above is authoritative;
progress.json reflects the last pre-failure stage and is not a final count summary.

## Recovery and remaining boundaries

Necessary source/tests and this status record are saved to the existing development
branch, not main. Read current-status.md and docs/runbooks/repository-recovery.md.
Raw receipts, caches, toolchains and binaries remain ignored, not restored by Git.
Another computer must independently prepare the fixed toolchain/offline modules
and an approved fresh matrix identity. Do not replay consumed001 or infer another
run/longer fuzz from this report.

Even26/26 would not rewrite historical matrix3 (24pass/Envelopefail/Pairunrun),
prove the old deadline cause, establish adequate fuzz coverage or close original
origin001/87, C11/C12/Task10/B01/production/main acceptance. Native aggregate process
counts are not individual identity evidence; fixed short fuzz counts are limited
samples, not a performance or absence-of-defects proof. Helpers are not transactional
protection against hostile concurrent filesystem changes. Review itself ran no
native/Go/tests/PID queries and did not judge actual environment outcome.
