# Fixed C4 Observer Entry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for sequential implementation in this session if the user confirms the recommended method; otherwise preserve their chosen method. Steps use checkbox syntax. One final independent review, no live launch.

**Goal:** Add a fixed English/RTL observation entry without changing the tested C5 entry or shared observation logic.

**Architecture:** A no-start-by-default script owns C4 policy and live-boundary callbacks. It calls the existing ten-callback orchestration unchanged. A create-only manifest binds the run to C4 before process acquisition; synthetic tests exercise private functions through doubles.

**Tech Stack:** Windows PowerShell7, existing process/native/session-guard modules; no new packages or product build.

**Spec:** docs/superpowers/specs/2026-11-09-windows-c4-observer-entry-design.md, user approved after35d3271b.

Status: awaiting plan confirmation and execution-method choice. Code/test execution
has not started. Live allowance0; future runtime approval is separate.

## Global Constraints

- Use the existing isolated worktree and codex/c12-b01-task9-coordinator branch. PowerShell only.
- Fixed executable: clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe.
- EXE SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95; DLL SHA2567F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.
- Arguments exactly `--layout-test --layout-language=en-US --layout-profile=rtl-mixed`.
- Only public parameter: `[switch]$RunApproved`; default returns one NotStarted object before imports/probes/writes. No path/profile/language/time/backend overrides.
- Missing evidence blocks; ordinary token, clean expected branch, all consumed source blobs and artifact hashes, no existing targets. UI/window readiness is an operator prerequisite.
- No edits to observe-c5.ps1, its tests, c5-observation-session.psm1/tests, guard or native/managed modules. No product-code changes, installs, .NET build/test, real CIM/native compilation, actual app launch or sleep during implementation verification.
- Reuse fixed360000ms cap,2000ms maximum sleep, original Process/SafeHandle, exact native creation and microsecond CIM identity, raw/provisional/final receipt semantics. No kill/reopen/retry.
- Future operator limits: startup30s, normal close requested by300s, exit criterion30s. Synchronous overrun and hostile filesystem-race limitations remain explicit.
- Preserve necessary source/tests/docs on existing dev branch; no main merge, raw receipt/screenshot/binary commits or acceptance promotion.

## Review Focus

1. Default invocation or incomplete test doubles reach a live effect: reject before any probe or acquisition; never silently supply live fallback to synthetic tests.
2. Manifest creation, serialization, flush/disposal or sample-writer construction fails: no process launch and all already-owned resources receive cleanup attempts.
3. RunId/path collision or reparse ancestor redirects receipts: reject before writing; existing contents remain untouched, within the trusted-workspace threat model.
4. C4 evidence is mislabeled C5 or policy drifts: manifest/run correlation and exact English/RTL arguments remain independent of observed CIM text; shared diagnostic naming is documented.
5. Boundary wiring bypasses original-resource ownership or UTC validation: preserve original Process/SafeHandle and refuse absent/ambiguous CIM time; cleanup errors never become success.

## File map

Create:
- `clients/windows/tests/observe-c4.ps1`: fixed policy, preflight, create-only manifest/writer, ten live callbacks, shared orchestration invocation.
- `clients/windows/tests/observe-c4.tests.ps1`: synthetic behavioral suite; private-function AST extraction, recording doubles and memory-backed streams only.
- `clients/windows/diagnostics/c4-observation.md`: invocation boundaries, manifest/receipt meaning and future UI evidence route.
- `docs/roadmap/2026-11-09-windows-c4-observer-entry-execution.md`: portable RED/GREEN/review ledger and unresolved limits.

Update `docs/roadmap/current-status.md`, `docs/runbooks/repository-recovery.md`, and
this plan/spec status at completion. Read the existing C5 entry/tests for patterns,
not to dot-source them or copy their sampling loop. No independent shared interfaces
between new tasks: this is one integrated, testable deliverable.

## Task1 — Dedicated entry, synthetic coverage and recoverable handoff

**Consumes:** `Invoke-TalenroC5Observation -Mode Live|Synthetic -Context <IDictionary> -Backend <IDictionary>`.
Context exactly RunId, ExpectedPath, ExpectedCommandLine. Backend exactly Preflight,
AcquireWriter, AcquireProcess, GetIdentity, ReadNative, ReadCim, ReadManaged,
ElapsedMs, SleepMs, Publish (scriptblocks). Reuse existing adapters and guard unchanged.

**Produces (private definitions in observe-c4.ps1):**
- `Get-C4Policy([string]$Root) -> hashtable`: Root, Executable, Dll, Arguments, ExpectedCommandLine, ExeHash, DllHash, ReceiptRoot, Sources. Root is test-only input internally; live root derives from script location.
- `Assert-C4OrdinaryToken() -> void`; `Invoke-C4Git([string]$Root,[string[]]$Arguments) -> string[]`; `Assert-C4Preflight($Policy) -> void`.
- `Assert-C4ReceiptDestination([string]$Directory) -> void`: reject existing destination and existing reparse ancestors.
- `New-C4ReceiptDirectory([string]$Directory) -> void`; `Open-C4NewStream([string]$Path) -> owned stream` (CreateNew/Write/Read share); `New-C4TextWriter($Stream) -> owned writer` (UTF8 without BOM, AutoFlush).
- `New-C4ReceiptWriter($Policy,$Context) -> one owned samples writer` (WriteLine/Dispose). Private filesystem/writer wrappers above are replaced by memory doubles in tests, never exposed on CLI.
- `New-C4Backend($Policy) -> hashtable` of exactly the ten callbacks. Capture CommandInfo for private dependencies before GetNewClosure, following the existing closure fix.

**Manifest contract:** `process-observations/<runId>/manifest.json`, beside samples.jsonl.
SchemaVersion=1, Observation='C4', RunId, Language='en-US', Profile='rtl-mixed',
ExecutableRelativePath (fixed repo-relative path above), Arguments, ExeSha256,
DllSha256, ObservationCapMs=360000, MaxSleepMs=2000. No timestamps, environment
inventory, arbitrary exception text or credentials needed. RunId must be lowercase
nonzero GUID N and match Context. Expected path/command line must match Policy.
Manifest is preparation metadata, never proof that a target launched or UI passed.

- [ ] **1. Write default/policy/preflight tests before entry implementation.**
  Define `Check($condition,$name)` and a failure assertion helper in the test script.
  Install fail-on-call Start-Process/Get-CimInstance/Add-Type/Start-Sleep/file wrappers
  before executing default entry or AST-extracted private definitions. Assert:

  ```powershell
  Check (@(& $entry).Count -eq 1) 'one default result'
  Check ((& $entry).Status -ceq 'NotStarted') 'no-start status'
  Check ($policy.Arguments -ceq '--layout-test --layout-language=en-US --layout-profile=rtl-mixed') 'fixed C4 arguments'
  Check ($policy.ExpectedCommandLine -ceq ('"'+$policy.Executable+'" '+$policy.Arguments)) 'independent full command'
  Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper absent'
  ```
  Include literal EXE/DLL hashes; unknown Executable/Profile/Language/Deadline/Backend
  switches reject; token/branch/dirty/source/missing/hash/target/inaccessible-CIM
  failures each produce zero writer/launch calls when composed with the shared scope.

- [ ] **2. Run the initial RED.**
  Run `pwsh -NoProfile -File clients/windows/tests/observe-c4.tests.ps1`.
  Expected: explicit missing-entry assertion, exit1, no live effects.

- [ ] **3. Implement default guard and private policy/preflight functions.**
  Use advanced script binding so unknown parameters reject. Source list includes
  observe-c4.ps1 plus all shared runtime .psm1/.cs files it consumes, not observe-c5.ps1.
  Validate Git exit codes and normalized working source blobs. Inventory as in C5;
  broad prototype matching is conservative and documented. No live script body yet.

- [ ] **4. Write receipt ownership tests, then run RED for missing receipt behavior.**
  Execute private functions with recording wrappers and memory streams. Success order:
  destination-check, directory-create, manifest-open/CreateNew, write, flush/dispose,
  samples-open/CreateNew, writer transfer; process acquisition only after all succeed.
  Parse captured manifest JSON and assert every field/value above and RunId equality.
  Fault matrix: invalid/colliding runId, context mismatch, existing directory/file,
  reparse ancestor, directory creation, manifest open/write/flush/dispose, samples
  open/writer construction. Assert zero launch, disposal attempts for every acquired
  resource, no overwrites/deletes, no extra success output and no private error text.
  Destination checker uses mockable path/attribute reads; no real reparse creation.
  Run same test command. Expected: new behavior assertion fails, not live invocation.

- [ ] **5. Implement receipt factory and ownership transfer.**
  Validate context first. Manifest stream/writer belongs to factory until fully
  disposed; failed cleanup blocks launch. Sample stream belongs to factory until
  writer construction/AutoFlush succeeds; then returned writer transfers to existing
  scope. On any partial factory failure attempt all owned cleanup without losing the
  primary error; do not delete partial artifacts or retry. A remaining incomplete
  directory is evidence, not a successful run. Use CreateNew for both files.

- [ ] **6. Write adapter/integration tests and observe RED, then implement live body.**
  Assert exactly ten callbacks; fake launch invoked once with fixed arguments, clock
  starts immediately before launch, original object/handle forwarded, native before
  managed, exact PID CIM filter and explicit UTC. Zero/multiple CIM rows, missing date
  or Unspecified kind reject; Local converts to UTC. Shared orchestration tested with
  C4 context and synthetic existing-adapter evidence must correlate with manifest,
  retain failures and dispose originals. Missing callback rejects without fallback.
  Live body: preflight before imports, fresh GUID, backend, unchanged orchestration;
  second preflight before writer matches C5 practice. Publish sanitized progress;
  explain C5-prefixed shared errors. Nonverified result/failure exits unsuccessfully.
  Never call RunApproved with real adapters during implementation. Expected GREEN:
  C4 suite exit0, native helper absent, recorded effect counts match expectations.

- [ ] **7. Write guide and execution ledger.**
  Include default versus later authorized invocation, manifest-only/incomplete files,
  raw provisional versus finalized terminal outcome, missing-artifact recovery and
  exact future C4 evidence route from spec. Codepoint/visual direction claims remain
  distinct; invisible controls stripped by tools stay UNVERIFIED. Capture actual
  RED/GREEN results and deviations. Update status/recovery without clearing gates.

- [ ] **8. Run scoped verification and commit checkpoint.**
  In separate PowerShell hosts run C4 tests and existing observe-c5,
  c5-observation-session, observation-session-guard, process-observation and
  native-process-observation `.tests.ps1` scripts. Expected: all exit0; existing
  counts32/120/48/32/78 unchanged. Parse new scripts with PowerShell parser; expected
  zero errors. `git diff --check` must pass; diff of C5/shared runtime against task
  baseline must be empty. Explicitly stage only new source/tests/guide and necessary
  docs; commit `feat: add fixed C4 observer entry with synthetic coverage`.

- [ ] **9. Obtain final independent read-only review and preserve result.**
  Review entire task against spec and all five Review Focus items. No live execution
  by reviewer. Fix Critical/Important with failing regression then GREEN full scoped
  suite; record deferred Minor findings and limitations. Commit necessary fixes/docs,
  push existing verified GitHub development branch, compare remote/local SHA and
  confirm clean worktree. Expected: remote equality, no raw/private files, no main merge.

## Self-review and handoff

One task owns policy, receipt acquisition and adapter wiring; shared APIs remain
unchanged. Steps1/4/6 cover the five Review Focus classes; Step7 carries all future
UI and evidence limits, without granting runtime execution. Manifest ownership and
cleanup are explicit, and no arbitrary public runner is introduced. Live Windows
behavior, races and RTL UI cannot be established by these synthetic checks.

Recommended method: implement sequentially in this session, with one fresh final
review. The entry is one tightly coupled deliverable; per-component implementers
would add coordination without independent interfaces. Confirm the plan/method
before implementation. This confirmation does not grant an application launch.
