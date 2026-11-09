# Fixed C4 observer entry — short design

Status: user selected the dedicated C4 entry approach and approved this written
design after35d3271b and inline implementation plan after5a13d26b. Implementation
and synthetic checks now exist; final review pending. No build or native launch.
Existing live allowance0. Scope is a small observation-tool addition, not UI repair.

## Intent and approaches

Prepare safe evidence for the outstanding C4 English/RTL profile while preserving
the tested C5 entry. User selected a separate fixed entry sharing observation logic.

- Selected: dedicated `observe-c4.ps1`, fixed C4 policy, existing orchestration and
  native/managed/identity/resource modules reused unchanged. No profile CLI choice.
- Alternative: add C4/C5 selection to the C5 entry. Less boundary duplication, but
  changes the already exercised entry and increases its public input surface.
- Rejected: ad-hoc launch or copied sampling loop; loses portable regression coverage
  or creates a second identity/deadline/cleanup implementation.

Implementation proposal: C4 owns its small live-boundary functions rather than
extracting private C5 functions or dot-sourcing a launch script. This deliberately
duplicates boundary wiring, not sampling/identity logic; future common-boundary
refactoring is outside scope. Synthetic tests must cover parity of safety gates.

## Fixed policy and compatibility

PowerShell7, same preserved executable:
`clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
EXE SHA256 `4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95`;
DLL SHA256 `7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB`.
Arguments exactly `--layout-test --layout-language=en-US --layout-profile=rtl-mixed`.
No executable, language, profile, deadline or callback CLI overrides. Missing or
changed artifact blocks; never rebuild or substitute automatically.

Default returns NotStarted before imports/CIM/native compilation/file writes.
Explicit RunApproved switch selects preparation, not authorization. Live requires
separate bounded approval, ordinary token, expected clean development branch,
matching runtime source blobs and artifact hashes, no existing demo/prototype
processes/windows and usable supported UI observation. Do not alter security settings.

Reuse `Invoke-TalenroC5Observation` despite its historical name: its ten-callback
contract accepts independent fixed expected identity and contains no profile logic.
Do not rename its API, change schemas or edit C5 runtime files in this work. C4
operator instructions must identify reused C5-prefixed diagnostic errors as shared
observer failures, not evidence that C5 ran. Bind C4 runId to exact arguments in a
new create-only manifest beside raw samples, written before launching; include only
fixed policy/hashes, no private environment inventory. Manifest/write failure blocks.

Retain original Process/SafeHandle, native-first identity and paired observations,
exact native creation identity and microsecond-truncated CIM comparison. Retain
failure evidence, provisional raw outcome, post-cleanup final result and overrun.
Unique create-new receipts under ignored artifact output; existing destinations and
reparse ancestors reject. Never kill, reopen PID or retry. Scope is a trusted workspace,
not hostile filesystem-race protection or hard real-time guarantees.

## Verification before any live request

Add C4 entry synthetic tests: default no effects, unknown overrides reject, fixed
English/RTL args and independent full command line, complete callbacks, source/hash/
token/branch/target rejection before acquisition, original resources, manifest and
writer acquisition failures, correlation and no overwrite. Use doubles/in-memory
fixtures; no live fallback, actual app/native helper/CIM or sleep. Retain the existing
five synthetic suites (historical310 checks, not a fresh result). Review the addition
before a separately authorized native session. No .NET build/test or dependencies.

## Future C4 evidence route (proposal, no launch allowance)

Proposed one launch, zero retries,360s observation cap, maximum2s sample interval;
startup30s, normal close requested by300s and normal-exit criterion30s. UI time
budget is checked before every action; leave unfinished checks NOT_RUN, never extend.

1. Login: English ordinary labels and permanent offline notice; root visibly
   mirrored and synthetic heading matches the committed rtl-mixed sample as far
   as available evidence permits. Inspect actual read-only email separately.
2. Enter demo, inspect Connection, Nodes and Account: visible layout direction,
   technical identifiers and actual email remain readable left-to-right. No search
   typing, connection request or node selection is needed; inaccessible identifiers
   remain UNVERIFIED rather than adding state-changing actions.
3. If time remains, logout/re-enter: chosen profile retained and visible ordinary
   reset state. This is selected visible-state evidence, not exhaustive reset proof.
4. Normal titlebar close; use retained-handle signal/code/time plus managed agreement.

Separate verdicts for visible mirroring, real-email LTR, technical-ID LTR, exact
fixture codepoints, clipping and profile retention. Screenshot appearance or title
cannot establish invisible bidi controls; accessible text that strips controls is
insufficient for exact codepoint claims. Report UNVERIFIED, not PASS by inference.
No natural-language shaping, screen-reader, contrast, target viewport, full keyboard,
fallback-language, C6, overall C5 or shutdown-root-cause acceptance is included.

## Deliverables and handoff

After design approval, write an implementation plan for the C4 entry, tests, guide
and portable execution record; specify manifest fields and acquisition cleanup
ownership before coding. No product source or existing C5 runtime changes. Preserve
necessary files on the existing development branch; no main merge. Raw/screenshots/
binaries stay out of Git. Approval of implementation never grants a native launch.
