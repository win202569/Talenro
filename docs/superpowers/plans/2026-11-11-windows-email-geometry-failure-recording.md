# Email Geometry Failure Recording Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (preserved inline method) to implement this plan task-by-task. Steps use checkbox syntax.

**Goal:** Resolve Important1 by retaining sanitized operation and cleanup categories through the fixed entry's final failure surface without launching the app.

**Architecture:** Entry-owned bounded state is shared by ten backend callbacks, resource adapters and finalization. The shared observation engine remains unchanged; adapters forward identity and reads to the original Process/SafeHandle. No copied observation loop or new persistence channel.

**Tech Stack:** Existing PowerShell7/.NET memory fixtures and Git; no new dependencies.

**Spec:** docs/superpowers/specs/2026-11-11-windows-email-geometry-failure-recording-design.md (approved2026-11-11).

Status: prepared and self-reviewed; awaiting plan confirmation. No repair code executed.

## Global Constraints

- Existing codex/c12-b01-task9-coordinator linked worktree; PowerShell only. Record actual execution BASE before edits.
- Only observe-email-geometry.ps1, observe-email-geometry.tests.ps1 and necessary documentation change. Freeze old entries/tests, shared modules, App source, hashes, arguments and observation limits.
- No real launch/CIM/native probe, build, restore, install, main merge or acceptance promotion. Live allowance0. No retries, PID reopening or forced termination.
- Default invocation remains NotStarted before state allocation/imports/probes/writes. Public surface remains only RunApproved; no new public overrides.
- One run ID, existing exact manifest fields and separate app/observer paths remain unchanged. Shared cap360000ms/max sleep2000ms unchanged.
- Success remains GeometryEvidenceStatus=Unreviewed and Acceptance=NotAccepted. Failed cleanup must never permit entry success.
- No raw exceptions, text, stack, arbitrary paths or error objects in public failure summaries. No inner exceptions in the final sanitized exception.
- Preserve normal checkpoint push authority, not PR/main merge authority. Keep Important1 open until final-surface regression evidence passes.

## Review Focus

1. Adapter construction fails after acquiring a raw resource: no ownership gap, all already acquired resources have one cleanup owner.
2. Writer and stream cleanup both throw while observation already failed: retain all categories, attempt process cleanup too, never retry releases or report success.
3. State reuse/cross-run context and malicious exception payloads: reject mismatched correlation before acquisition and never leak payload or grow state with repeats.
4. Internal shared validation or final publication fails: mark inaccessible session errors unclassified, preserve known categories and throw even if information publication fails.
5. Process adapter hides identity failures or changes timing/evidence: original object/handle and native-first order remain demonstrable; original deadline/overrun tests stay unchanged.

## File map and exact interfaces

Modify clients/windows/tests/observe-email-geometry.ps1: all new helpers stay behind
the default guard in this existing entry. No module extraction/source gate expansion.
Modify clients/windows/tests/observe-email-geometry.tests.ps1: retain AST extraction,
fake OS functions and memory fixtures; enhance disposal counters/multiple faults.
Modify clients/windows/diagnostics/email-geometry.md and existing execution ledger
docs/roadmap/2026-11-11-windows-email-geometry-observer-entry.md for resulting contract.
Update current-status.md, repository-recovery.md, this plan and approved spec status.

Private interfaces (internal only, names fixed for tests):
- New-EGFailureState([AllowNull()][string]$RunId) -> hashtable with RunId, ordered lists OperationFailures/CleanupFailures. Null allowed before ID generation; otherwise canonical nonzero lowercase32 GUID.
- Add-EGFailure($State,[string]$Category) -> no output; validates against lists below, deduplicates ordinally, preserves insertion order within each list.
- Get-EGFailureSummary($State) -> detached PSCustomObject with exactly SchemaVersion=1,RunId,Outcome=Failed,OperationFailures[string[]],CleanupFailures[string[]],GeometryEvidenceStatus=Unreviewed,Acceptance=NotAccepted.
- Stop-EGFailure($State) -> never returns; publish summary with Write-Information, then throw sanitized InvalidOperationException with Data['EGFailureSummary'] containing detached summary and no inner exception. If publication throws, record Publication and rebuild exception data without retrying publication.
- New-EGProcessOwner($Process,$State) -> adapter with Original, forwarding Id property, Dispose(). GetIdentity/ReadManaged explicitly unwrap Original.
- New-EGWriterOwner($Writer,$Stream,[string]$Kind,$State) -> adapter with WriteLine(string),Flush(),Dispose(); Kind exactly Manifest or Samples.
- New-EGTextWriter($Stream,[string]$Kind,$State) -> leaveOpen StreamWriter; construction/configuration failures record category and clean any constructed writer; caller still owns stream.
- New-EGReceiptWriter($Policy,$Context,$State) -> owned sample adapter; all manifest cleanup complete before return.
- New-EGBackend($Policy,$State) -> same exact ten callback keys; validates State.RunId matches Policy/Context before acquisition.
- Invoke-EGSession($Policy,$Context,$State,$Backend,[ValidateSet('Synthetic','Live')][string]$Mode) -> invokes existing engine once; on throw adds SessionUnclassified and calls Stop-EGFailure after shared cleanup. On return, any retained failure forces Stop-EGFailure; otherwise existing ConvertTo-EGResult. Top-level only passes Live; tests only pass Synthetic.

Allowlisted OperationFailures tokens (27 maximum):
Setup,Preflight,Import,ReceiptDirectory,ManifestOpen,ManifestConstruct,ManifestWrite,
ManifestFlush,ManifestAdapter,SamplesOpen,SamplesConstruct,SamplesWrite,SamplesFlush,
SamplesAdapter,ProcessAcquire,ProcessAdapter,Identity,NativeRead,CimRead,ManagedRead,
Clock,Sleep,Publication,SessionUnclassified,Context,Result,State.
CleanupFailures tokens (5 maximum):ManifestWriterDispose,ManifestStreamDispose,
SamplesWriterDispose,SamplesStreamDispose,ProcessDispose.
Unknown tokens throw without appending; never use exception messages for classification.
SessionUnclassified means the shared error is opaque, not that recorded callbacks
all succeeded; add it even alongside known categories on any shared-session throw.

## Task1 — State, ownership and final-surface repair

**Files:** new entry/tests above. **Consumes:** existing policy/context, shared ten-callback engine and original resource fixtures. **Produces:** all interfaces above and end-to-end failure evidence.

- [ ] Add state tests before implementation: null/canonical run ID, invalid IDs, unknown tokens, insertion order/deduplication, detached snapshots, exact eight-field summary, no cross-run mutation. Assert operation/cleanup counts never exceed27/5 even after1000 repeats. Run `pwsh -NoProfile -File clients/windows/tests/observe-email-geometry.tests.ps1`; expected RED missing New-EGFailureState, no OS effect.
- [ ] Implement state/summary helpers, run same command; expected GREEN for state tests and retained existing coverage.
- [ ] Add adapter fault tests with independent counters: writer cleanup throws then stream cleanup still runs; both throw; process cleanup throws; repeated Dispose does not retry; constructor/adapter creation fails after raw acquisition. Assert each owned release attempted once and every distinct token retained. Run same command; expected RED absent adapters or incorrect ownership.
- [ ] Implement adapters and integrate factory ownership. Construct writers using UTF8(no BOM), buffer1024, leaveOpen=true and existing AutoFlush=true. Mark release attempted before calling Dispose; attempt writer then stream in finally-equivalent logic, rethrow failure after both attempts. Use explicit factory-to-adapter transfer flags; failed construction leaves factory ownership. Record constructor-path cleanup failure rather than swallowing it. Update fixture streams to count attempts rather than equating a thrown Dispose with a proven leak. Run same command; expected GREEN.
- [ ] Add real backend/shared Synthetic integration regression using simultaneous managed-read, samples writer/stream disposal and process disposal failures. Exercise Invoke-EGSession and Stop-EGFailure, capture final safe exception and information output. Pin core assertions:

```powershell
Check ($summary.OperationFailures -contains 'ManagedRead') 'managed category retained'
Check ($summary.OperationFailures -contains 'SessionUnclassified') 'opaque session labelled'
Check (($summary.CleanupFailures -join ',') -ceq 'SamplesWriterDispose,SamplesStreamDispose,ProcessDispose') 'all cleanup categories retained'
Check ($summary.Acceptance -ceq 'NotAccepted') 'failure never accepts'
Check ($caught.InnerException -eq $null) 'no raw inner exception'
Check (($summary | ConvertTo-Json -Depth 5) -notmatch 'private-') 'payload absent'
```

  Run same command; expected RED missing session wrapper or lost categories, not real acquisition.
- [ ] Add final-surface cases: inaccessible identity validation failure yields SessionUnclassified; information publication throws once and terminal Data contains Publication; state contains a failure despite synthetic successful result so wrapper rejects; context/run mismatch has zero acquisition. Preserve original Process reference checks on adapter.Original, managed argument and SafeHandle; assert native-first order and no PID reopening. Run same command; expected RED on each missing contract before implementation.
- [ ] Wire all ten backend boundaries to the shared state, preserving callback output cardinality. Process acquisition catches adapter-construction errors and disposes acquired original exactly once. Id-forwarding/identity failures record Identity; callback failures record named category then rethrow. Writer methods record their Kind-specific category before rethrowing; Dispose records cleanup categories. Do not modify shared engine or catch based on raw message text.
- [ ] Wire top-level initialization after inert guard: initialize state with null RunId, generate ID once and bind state/policy/context; tag setup/preflight/import failures where they occur. Invoke-EGSession once. Ensure Stop-EGFailure is not recaught and published twice by top-level catch (separate setup catch from session execution). Existing nonverified/unfinished result still exits unsuccessfully. Run same suite; expected GREEN, record actual assertion count.
- [ ] Run seven suites separately using `pwsh -NoProfile -File clients/windows/tests/<name>.tests.ps1`: observe-email-geometry,observe-c4,observe-c5,c5-observation-session,observation-session-guard,process-observation,native-process-observation. Expected all exit0; existing counts152/32/120/48/32/78 unchanged. Previous new180 count is a baseline, not the new required total.
- [ ] Run `git diff --check`; compare frozen files against execution BASE; read-only recheck old/new EXE/DLL hashes from2026-11-11-windows-c4-email-geometry-build.md. Expected clean whitespace, no frozen source edits and same hashes (missing artifact is reported, not rebuilt). Commit explicit entry/tests as `fix: retain geometry observation and cleanup failure categories`.

## Task2 — Review, documentation and recovery

**Files:** documentation listed above. **Consumes:** Task1 tested final summary and ownership contracts. **Produces:** review disposition, current/recovery status and remote checkpoint.

- [ ] Document allowlist/unclassified limitations, no durable delivery guarantee, original resource ownership and no-live boundary. Record RED→GREEN evidence and exact actual tests; do not label Important1 fixed merely because code exists.
- [ ] One fresh read-only final review for this repair plan, all five focus items, no live actions. This is the bounded addendum implementation review, not a repeat review of untouched prior code. Use preserved inline execution method; no per-task implementer agents. Address Critical/Important in one regression-first pass, rerun seven suites, record every declined boundary and deferred Minor; no re-review loop.
- [ ] Update existing entry plan/ledger to close Important1 only when covered final-surface regressions and review disposition support it. Remaining real-run/geometry evidence stays NOT_RUN/NotAccepted. Update status/recovery/spec and this plan; retain historical test evidence as historical.
- [ ] Check explicit diff list and `git diff --check`, commit documentation/fixes, push existing branch only. Verify `git rev-parse HEAD` equals `git ls-remote origin refs/heads/codex/c12-b01-task9-coordinator` and `git status --short` is empty. Expected remote recovery checkpoint; no PR/main merge or launch.

## Self-review and handoff

Spec coverage: state/privacy/final summary in Task1; partial ownership and constructor
cleanup in Task1; original evidence and shared timing preserved by integration plus
unchanged six suites; documentation/review/recovery in Task2. All five focus cases
have explicit tests. Signatures consistently pass one shared State and retain exact
backend keys; default guard stays ahead of all effects. Author choices: finite27/5
tokens, detached exception-data summary, conservative SessionUnclassified on shared
throws. No tests or implementation performed during planning. Confirm this plan
before preserved inline execution; no separate live approval is implied.
