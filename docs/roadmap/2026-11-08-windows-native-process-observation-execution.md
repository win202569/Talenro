# Native recorder execution ledger

Plan: docs/superpowers/plans/2026-11-08-windows-native-process-observation.md.
Approved after85169c27; baseline85169c27, clean linked worktree, no submodule.
PowerShell7.6.9; original recorder baseline32/32 synthetic PASS.
Task1 in progress; Task2/3 pending. No live/native execution allowance.
Pre-flight: Task1 raw payload normalization feeds Task2 DTO adapter; Task3 uses
both test receipts. DTO time/exception metadata must not be replaced by adapter
timestamps or rethrown wrappers. Preserve native helper timing/type verbatim.
Ruling: use this portable ledger instead of Bash-only workspace/task scripts —
Windows PowerShell-only requirement and recoverability take precedence; manual
task completion/commit evidence replaces script bookkeeping.
Ruling: run only approved synthetic/static/compile checks, not the full app suite
or dependency setup — explicit scope forbids live calls/builds; broader behavior
remains unverified. Static safety checks supplement manual review, not runtime proof.

Task1: complete. Missing-module RED (exit1) then48/48 synthetic GREEN (exit0).
Commands: & ./clients/windows/tests/native-process-observation.tests.ps1.
Only normalization/synthetic export exists at this checkpoint; native helper not loaded.
Task2 in progress; Task3 pending. No native API/application calls.

Task2: complete. RED missing live export and missing helper source (both exit1),
then66/66 synthetic/static and20/20 compile/reflection checks passed (exit0).
Original synthetic regression32/32 passed; original module/test diff empty.
Commands: & ./clients/windows/tests/native-process-observation.tests.ps1;
& ./clients/windows/tests/native-process-observation.compile.tests.ps1;
& ./clients/windows/tests/process-observation.tests.ps1 (separate fresh hosts).
Compile check invoked only the private stale-type loader rejection branch after
explicit compilation; it did not invoke live export, Capture or any native API.
Interop manually inspected for immediate errors and reference finally; ABI and
concurrent handle behavior remain runtime-unverified. Task3 review pending.

## Final independent review and one fix pass

Fresh reviewer reviewed85169c27..ccbf9ac0, independently reran66/20/32 checks,
and found two Important/P2 issues, no Critical or separate Minor findings:
1. Get-RawMember pipeline enumeration allowed nested singleton arrays to bypass
   exact scalar checks. Added six malformed nested-field regression assertions;
   observed RED at reject nested array Wait, then preserved member values without
   enumeration in both dictionary and object-property paths.
2. New-NativeCallRecord string binding coerced absent not-run times into empty
   strings. Added in-memory DTO conversion fixtures (no helper or native calls),
   observed RED at not-run Wait has absent timestamps, then retained nullable
   timestamps. Fixtures cover acquisition errors, original exception type/timing,
   later high-bit code/time data and no accidental conversion of unavailable data.
Final: fixed both findings — separate regression RED observed for each; combined
GREEN78/78 synthetic/static assertions. No second review or native invocation.

Final: minor (deferred): reviewer recommended executable successful loader-cache
reuse/forced-reimport coverage. Preloaded-type rejection is tested; reuse/reimport
control flow is currently inspected only. Defer this additional coverage until
before live integration; no known loader defect or claim of that runtime coverage.

Reviewer declined-to-judge rulings:
- Final: Ruling: native ABI/Windows results remain NOT_RUN — explicitly excluded;
  cost is possible native defects despite compilation.
- Final: Ruling: disposal races/reference acquisition remain static-only — no
  runtime probing allowed; cost is unverified concurrent lifetime behavior.
- Final: Ruling: launch ownership/creation identity stays caller responsibility —
  recorder accepts a retained handle; integration must enforce identity or evidence
  could be attached to the wrong launch.
- Final: Ruling: application shutdown/root cause/acceptance unchanged — diagnostic
  component is not a fix; cost is existing unresolved shutdown blockage.
- Final: Ruling: full suites/builds/installs/historical changes excluded — preserve
  approved scope; no broader project regression coverage is claimed.
- Final: Ruling: preserve development branch/worktree and ordinary checkpoint push
  under AGENTS standing authorization, no merge/PR/cleanup — broader readiness
  remains unassessed. No disposable workspace was created or deleted.

Primary references checked for implementation boundary:
- https://learn.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.marshal.getlastpinvokeerror?view=net-10.0
- https://learn.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.safehandle.dangerousaddref?view=net-10.0

Task3: complete for approved development-checkpoint scope. Final fresh-host runs:
native synthetic/static78/78, compile/reflection20/20, original synthetic32/32;
all exit0 after fixes. Native adapter/Capture/API invocation NOT_RUN throughout.
No app launch/build/restore/full suite/settings changes. Original recorder/test
unchanged. Recovery records include deferred loader-cache/reimport coverage and
all runtime limitations. Only ordinary development checkpoint push is intended.

## Loader coverage follow-up

User requested the next step after3b5e38dc. Added only
clients/windows/tests/native-process-observation.loader.tests.ps1, exercising the
existing private loader in a fresh PowerShell7 host. First load compiles the helper;
second load returns the identical cached Type. Forced reimport loses the private
cache and rejects the already loaded helper on both attempts without adopting or
replacing it. Import itself still does not load the helper. All11 assertions pass.
This is characterization coverage of existing behavior, not a new production fix;
first run passed, no RED/GREEN implementation claim. Production files unchanged.
No handles were constructed, no live export/Capture/native APIs were invoked.
Regression reruns:78 synthetic/static,20 compile/reflection,32 original synthetic,
all pass. Deferred loader coverage item is now closed; earlier review remains
historical, this small test-only follow-up was self-reviewed. Actual ABI, handle
lifetime, launch identity and application shutdown still require scoped runtime
evidence. Next prepare same-handle paired observation scope, not automatic launch.
