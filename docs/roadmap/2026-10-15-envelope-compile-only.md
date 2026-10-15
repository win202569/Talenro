# Envelope compile-only checkpoint — 2026-10-15

## Approved scope and implementation

The user approved a bounded compile-only option and, after helper checks/review,
one actual fresh envelope-observe-003 attempt. Revalidate version and origin,
then allow one existing compile with its600s bound. Never run the resulting test
binary or fuzz. Stop at any failure without retry. No install/download, policy
override, security change, WPR, Pair, broader product suite or main merge.
Preparation-only002 and failed001 remain consumed, not writable build targets.

`Invoke-EnvelopeObservation` adds trailing switch -CompileOnly. Combining it with
-PreparationOnly fails before planner/filesystem work. Neither switch supplied
still selects the existing build-and-observe path; preparation-only behavior is
retained. Compile-only performs the existing60s version/origin checks, offline
readonly environment,600s build, strict natural process/cleanup checks and
copied-context/fuzz/original-Envelope compiler-log checks. Original callback,
parser sources, product inventory and copied diagnostic source binding remain
required. The existing binary receipt binds source identity, binary hash and
compiler evidence. No native runner or completion policy is modified.

After recording the compiled binary binding, compile-only returns without a
success object and before the execution branch. Finally saves and compares the
last product inventory and rechecks binary hash/copied source identity. Only
then does it save compile-result.json and emit success. CompileOnly/BuildExecuted
are true; SamplingExecuted/Acceptance false. Success-file write errors terminate.
Stopped records include both mode flags and stage existence. Finalization errors
for either bounded mode preserve a stopped artifact and never emit success.
A binary receipt on a failed finalization is binding evidence, not compile success.

## Test-first verification and review

New envelope-compile.tests.ps1 initially failed at compile-only success: the old
driver reached version|origin|build|execute and the substituted process boundary
refused execution. The failure was inspected, confirming missing early-stop
behavior rather than a malformed compiler-log fixture. After implementation,
all19 assertions passed on PowerShell7 and WindowsPowerShell5.1.

Tests execute the real driver, planner, copied-source provenance, native receipt
completion predicates, compiler evidence parsers, inventory comparison, result
files and a streaming consumer. Only external Go launching and the slow full-tree
inventory are substituted, with constrained command arguments/bounds and a real
one-file hash fixture. A controlled non-executable file stands in for compiler
output; it is never launched. These results do not establish real compilation.

Coverage includes exactly version/origin/build and no execute; mode/acceptance
labels and binary/source receipt binding; build nonzero exit, failed natural
completion and missing log evidence; final binary drift, inventory drift and read
failure without success JSON/stream output; unchanged default execution boundary;
and conflicting modes before path planning. Existing19 preparation-only and21
planning/provenance checks also passed on each host:59 scoped checks per host.
The planning checks include a real product inventory. No full runner/product suite
was rerun; earlier31/137 results remain historical, not fresh acceptance evidence.

Independent read-only review found no Critical/Important defects and supported
one compile-only attempt, not sampling/main/acceptance. Minor coverage limits:
bad-log input fails the first copied-source parser rather than isolating the
Envelope parser; final copied-source drift and result-file write failure are not
directly injected. Actual compiler/dependency readiness and output functionality
were not judged; executing the binary is outside this approval.

## Single actual attempt

One invocation with -CompileOnly used fresh envelope-observe-003 on PowerShell7.6.5.
Wrapper exit0; CompileOnly/BuildPassed/BuildExecuted=true, SamplingExecuted and
Acceptance=false. No retry, no generated test executable launch and no fuzz.

| Stage | Native UTC interval | Saved observations |
| --- | --- | --- |
| version | 11:17:53.2262357–11:17:53.3476975 | Exit0, no timeout, natural complete shutdown, final0/empty, shutdown5ms |
| origin | 11:17:53.6566039–11:17:53.7615179 | Exit0, no timeout, natural complete shutdown, final0/empty, shutdown5ms |
| build | 11:17:54.1152042–11:19:46.0295662 | Exit0, no timeout,111914.362ms, natural complete shutdown, final0/empty, shutdown0ms |

Asia/Tbilisi build interval15:17:54–15:19:46. All three receipts have root exited,
NaturalShutdownVerified/ShutdownEvidenceComplete/CleanupVerified=true, no
termination requested, ActiveAfterCleanup0 and empty MembersAfterCleanup.
Build TotalProcesses517 is aggregate owned-job accounting, not517 simultaneous
processes or proof that any fuzz worker ran. Version/origin counts are2 each.
No historical PID was reopened; post-checks read saved evidence only.

Driver validation and a separate readonly saved-log check accepted the copied
context/fuzz compiler commands and original Envelope callback/DER/signature parser
command. Version/origin output checks passed the unchanged pinned runtime and
copied-toolchain/original-module requirements. Diagnostic source identity remains
E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306.
The test executable is15348224 bytes with SHA256
F422D2CBD6D9F2A6B34B5363CD45CC48E88AA49CDF0FF8DAB0D74DC765A798A6;
it matches binary-receipt.json and compile-result.json. Product inventories are
701 before/701 after, zero Path/SHA256 differences. No execute directory,
sample-result.json or stopped.json exists. Empty fuzz-cache scaffolding is not a
sampling result. This proves one compilation, not binary functionality or fuzz
acceptance, and does not repair historical001/origin87.

## Evidence and recovery

Ignored worktree-relative evidence root:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/envelope-observe-003/`.
003 is now consumed;001/002 stay consumed. No retry or replay is authorized.

| Local evidence | SHA256 |
| --- | --- |
| version/receipt.json | 12531914FC6E870D4862D60AAF1B6921AE624EC73DCA19DAC5B5D5D6666133CF |
| origin/receipt.json | 96B6C9AF24D5790564C943C5E599AF09CE7F94D300164C87645E63B90DC69DFD |
| build/receipt.json | 8273D743AFB25894C58122C91A932AF8D87270203049EAF7BCF22688DCD14BE6 |
| compile-result.json | D7C82ACA52D885E8C2EB420C4F0BB5D33A47B56AFB532ED64B157CF2202F0D39 |
| binary-receipt.json | 320E4DC6768976C4267C783F48AB1838EF10FD6F849963E2FFF0DA002D96F9E9 |
| build/stderr.txt | 9C0D19C31650A830DD168F5DFB8B325F76157AC95792B5FDCB9A7273886AFA60 |
| before-product.json and after-product.json | 5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |

Save reviewed source/tests and this report on the latest remotely verified
codex/c12-b01-task9-coordinator checkpoint, not automatically main. Read
current-status.md and docs/runbooks/repository-recovery.md when restoring on a
new computer. Raw logs/receipts, copied toolchains/caches and this machine's binary
remain ignored and uncommitted. Git restores source/progress, not an executable
ready for sampling; dependency preparation and a fresh authorized build would be
needed on a different machine. Do not reuse a report hash as executable proof.

Original Envelope/Pair sampling remains UNRUN, and C11/C12/Task10/B01/production/
main acceptance stays open. Next proposed work is a bounded single-observation
design defining whether and how to revalidate/reuse this local compiled artifact
under a fresh execution identity. No such reuse/observation entry is implemented
or authorized by this compile-only approval; the default driver would rebuild and
then execute and must not be invoked as a shortcut.
