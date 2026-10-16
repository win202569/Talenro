# Fixed leaf retained compile entry — 2026-10-16

Later separately approved real use:compile001 returned compiler exit0 but STOPPED
at the unchanged32-attempt diagnostic cap. No generated binary execution or retry;
read [actual outcome](2026-10-16-leaf-retained-compile-001.md). The fake-only scope
and historical test counts below describe preparation, not that later live run.

## Approved scope

User approved a bounded independent diagnostic compile entry for the package at
the historical matrix003 failure, plus fake-only tests/review and preservation.
Base:3c8c1a463aa52b229a82561e74b0d353c976d523. This approval does NOT include actual
Go invocation, a generated binary runtime, SDK audit, matrix replay, Docker,
download, security changes, main merge or product acceptance.

New files under docs/diagnostics/isolated-fuzz:

- leaf-retained-compile.ps1: fixed planner, retained-only native boundary,
  separate diagnostic receipt checks, controller/final audits/persistence.
- leaf-retained-compile.tests.ps1: portable tiny fixtures and fake inventories/
  native boundary; actual wrapper, parsers, source pins, files and hashes execute.

Existing mapping/matrix/child helpers, native collector/registry, ordinary
acceptance guard and all Go/product sources are unchanged.

## Locked execution and interpretation

Importing/planning never runs Go. Get-LeafRetainedCompilePlan requires a new owned
leaf-retained-compile-NNN identity, a64-bit Windows host, owned uninstrumented
toolchain/reference and an existing offline module cache. It rejects reparse/
linked inputs and output overlap with the SDK/module inputs. No input deletion,
SDK copying/patching, installation or dependency download is performed.

Invoke-LeafRetainedCompile, NOT invoked for real here, audits the fixed SDK using
the existing full11536-file/pinned manifest checker before and after execution.
The returned reference/copy-receipt digest pair is persisted as ToolchainReference.
Twelve helper/native source fingerprints and the actual go.exe hash are also
bound to the saved plan and checked pre/post. These are on-disk provenance, not
attestation of an already-loaded Add-Type assembly; later live use needs a fresh
host and separately approved concrete SDK/reference identity.

Exactly one command:the copied go.exe runs `test -c -x -p=1 -count=1 -timeout=120s`
with a fixed output binary and `talenro.local/platform/internal/nodecontrol/nodeleafbinding`.
Native deadline600s; unchanged500ms natural/shared5s cleanup. Arguments, package,
environment and budget are not caller-customizable. The project is the fixed cwd.
The new root has separate go-cache/tmp and uses the existing module cache with
`-mod=readonly`; that flag is not filesystem read-only enforcement for the cache.
The existing offline environment disables Go proxy/sumdb/toolchain fallbacks,
but does not sandbox all OS network access or all inherited tool settings.
Only compile:the generated binary is NEVER executed and no runtime follow-up is
present. The600s native limit does not bound pre/post audits or preempt OS calls.

Native dispatch uses RunWithRetainedHandleDiagnostics, preserving both diagnostic
flags/schema3. The dedicated observer requires unforced natural core completion,
original query metadata/core member coverage, unique owned retained entries,
ordered same-entry phase reads, final exit0 and release evidence. Root-wait
DescendantsRemaining is preserved, not mistaken for final residuals. Registry
records never fill missing original core rows. No ordinary guard is called with
cleared flags. Its actual diagnostic refusal remains exercised in tests.

At least one admission is needed; zero samples, missing/ambiguous/error87/forced/
budget/close failures stay unverified.32 attempts is capped even if its flag is
inconsistent. No bound expansion, retry, handle reopening or matrix enablement.
A cold Go build may reach this cap; that is an explicit incomplete outcome, NOT
permission to mark compilation accepted or start another attempt.

CompileExitZero, CompileEvidenceComplete, CoreNaturalComplete and lifetime facts
remain separate. Copied compiler tokens/original package test source, context and
internal/fuzz source tokens and a binary are needed for compile evidence. A
diagnostic ObservationPassed never sets Acceptance, PackagePassed, OriginalTestPassed
or RuntimeAttempted:all stay false, even on a complete sample.

## Durable evidence and failure handling

The controller saves plan/attempt, product before/after, diagnostic receipt and
streams, optional binary, initial artifact hash binding and provisional result.
Applicable product/SDK/helper-source/artifact audits are attempted independently;
one failure cannot suppress later audits. Primary failure and every audit error
are retained; no streamed success is emitted before finalization.

The manifest is an explicit bounded file list, never a traversal of generated
go-cache/tmp. Writes use CreateNew/path-link checks. result.json is provisional
(FinalizationRequired=true); completed.json is written only after successful
final audits/manifest and binds its exact hash. Any persistence failure stops
without that marker and attempts stopped.json. A successful-looking provisional
result without valid completion/hash bindings is NOT durable completion. Partial
native DTOs on native throw are unavailable, and no cleanup proof is invented.
Raw compiler logs/SDK/cache/binary/receipts stay ignored, never recovery commits.

## TDD and review

Initial test failed at the named missing-entry assertion, then the first fake
implementation passed its initial cases on both hosts. One independent read-only
review identified three Important issues:counter32 with a cleared cap flag,
short-circuited final audits, and missing persisted SDK receipt identity. All three
were reproduced as named focused RED assertions and corrected in one fix pass.
The focused checks then passed; tests cover product mismatch/persistence failure,
SDK failure and simultaneous product/SDK/artifact errors without skipping audits
or losing the initial compile error. The counter fixture is composite test DTO
data, not a real32-process chronology or native ownership proof.

Reviewer did not run tests or re-review the fix pass. Coordinator verifies the
fixes/final tests; no second independent review is claimed. Review exclusions
(live native/Go/SDK/product inventories/timing) are accepted as outside this
preparation. The native shim's method selection/arguments are source-reviewed,
not dynamically exercised. File link metadata is used, never fake PID queries.

Final scoped verification:7 suites/283 assertions perPowerShell7 and Windows
PowerShell5.1,566 primary records, all14 commands exit0. Breakdown:leaf entry99,
mapping36,retained child70,registry27,integration12,receipt16,short-child fixture23.
Four integration checks per host are static. Three focused review RED/GREEN pairs
are supplemental, not added to those totals. Six adjacent suites ran before the
review fix pass but import no new leaf helper and their dependencies are unchanged;
the complete leaf suite ran after all fixes on both hosts. No claim that every
command restarted after each edit. Initial missing-entry RED and first two leaf
GREEN runs were observed directly, not invented as local log files.

Ignored evidence:.superpowers/diagnostics/2026-10-16-leaf-retained/validation-001
retains14 final suite logs, six focused review RED/GREEN logs and context/source/
summary bindings. Its23-file evidence-manifest.json SHA256 is
`99332A4EF940885668577FB341EC4171D1B602980A6F74211571A7C96F58D2F6`.
All23 lengths/hashes and19 final source/test fingerprints were rechecked.
PowerShell parsing, whitespace and local document links were also checked.
Broader matrix/full-project/real-process tests are excluded, not counted as PASS;
older148/389-per-host totals must not be relabeled as current full verification.

## Recovery and next decision

Clone the existing development branch and read current-status.md,
[repository recovery](../runbooks/repository-recovery.md) and this scope. The new
suite uses only repository files, Windows PowerShell/.NET and tiny local fixtures;
it requires no private SDK cache or saved historical receipts. Existing adjacent
tests run in their existing fake/default modes. Git does not restore raw evidence,
SDKs, caches, credentials, or live process state.

Later real compile execution requires separate approval of its exact identity,
SDK/reference/module paths, current bindings,600s deadline and cap interpretation.
Do not reuse an identity or automatically execute the binary/full matrix.
Observation001 stays failed and002 stays a single sampled PowerShell observation;
neither is changed or replayed here. Error87's cause/fix is still unestablished.
Historical matrix003 remains2PASS/1FAIL/23UNRUN and wider C11/C12/Task10/B01/
production/main gates remain open.
