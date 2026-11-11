# C4 email geometry execution ledger

## 2026-11-11 — Task3 source implementation, review pending

BASE:ca176210. Debug/layout-only gate, isolated App paths, real-page adapter and
guarded App/MainWindow/page-target hooks implemented; XAML changes are names only.
Isolation RED: missing gates/probe exclusion, plus one test path-separator mismatch;
normalize only separator in assertion. After implementation8/8 passed. Adapter
contracts RED missing cleanup/invalidation helpers→5/5 passed; normalized XAML hashes
match original Login/Account declarations after removing only approved names.
Full portable suite367/367 passed, exit0. Six separate PowerShell hosts passed
152/32/120/48/32/78 assertions (462 total); existing tests/modules/LayoutWindowObserver
diff empty againste4c42a33. Old EXE/DLL SHA256 unchanged:
4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 /
7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.

Ruling: add small pure GeometryCleanup.Run and session Invalidate/RunId integration
members — attempt every detach and retain observer failure while using shared state;
risk: portable tests cannot exercise actual WinUI event registration. Review source,
then separately approved native build/live tests remain mandatory.
Ruling: always mark ClipChainComplete=false because implicit compositor/template clips
cannot be proven from this bounded parent walk — avoids false visibility claims;
cost: geometry may remain insufficient even when every measured box fits.
Ruling: overall adapter read/subscription exceptions stop collection and invalidate
the receipt; per-property failures preserve individual statuses. No retry/refund;
cost: an exceptional attempt may produce no sample, explicitly unsuccessful transport.

Independent review pending. No WinUI adapter compilation, native build/launch or
new output hash binding. Source/evaluation tests do not verify native output metadata.

## 2026-11-11 — Task2 complete

BASE:5c168a35. Fixed activation parser, validated/frozen bounded snapshot buffer,
create-only owned stream and idempotent Finish/Dispose implemented. Activation RED
(missing parser)→7 passed; receipt RED (missing session)→25 passed; serializer-seam
RED→27 passed. Additional element checks exposed swapped RootDIP/local-slot labels:
31 passed/2 failed before enforcing field-specific spaces, then33/33 passed.
Final full portable suite:354/354 passed,0 failed,0 skipped, exit0,113ms, no warnings.
No native build/launch, page changes or existing observer changes. Task3 and final
independent review pending; no diagnostic samples gathered and C4 remains NOT_ACCEPTED.

Ruling: use an internal serializer dependency and direct UTF8/no-BOM buffer write,
not StreamWriter — permits serialization failure injection and removes a separate
writer-owned lifecycle; stream capability/construction failure remains tested. Risk:
bounded buffered output uses memory before writing, acceptable for4 bounded samples.
Ruling: factory is internal with injectable path metadata, not public CLI/private-only;
an exclusive persistent `<run>.claim` file reserves each run before directory creation.
This prevents cooperating writers reusing the same ID; collisions/reparse ancestors
reject and existing content is untouched. Claim remains on failure (no deletion/retry).
Risk: extra tiny claim file; trusted parent directories must not be concurrently
replaced by a hostile actor. Metadata checks are not a handle-based hostile-race defense.
Ruling: freeze nested collections on Add; field statuses are validated using exact
property names, allowlisted roles/enums, finite values and field-specific spaces.
Risk if incomplete: rejected evidence rather than silent publication; Task3 must use
these exact names. Loaded/Visible retain false with non-Ok status when unavailable.

Finish means transport cleanup succeeded, not diagnosis success. Zero samples are
explicit; footer is provisional until external normal-exit0 evidence. IO is synchronous
outside layout callbacks and can block; no hard timeout guarantee, retry or worker.
All task-owned test temp directories were cleaned by scoped test cleanup; no user
receipts removed. Original restricted-run timeout record remains below.

## 2026-11-11 — Task1 implementation

Task1 BASE:8945c205. Added pure evidence/evaluator and budget/coordinator, linked
explicitly into portable tests. Geometry filter first failed with missing new types,
then passed20 cases; capture filter first failed with missing budget/coordinator,
then passed6. Full portable suite passed312/312 before the final two evidence cases.
The added extreme-coordinate fixture initially expected overflow of an arithmetic
operation that actually rounded to a finite value; corrected its left coordinate
to MaxValue/2 so subtracting -MaxValue genuinely overflows. Production unchanged.
Final Task1 full portable suite:314/314 passed,0 failed,0 skipped, exit0 (95ms).
Task1 complete; Tasks2/3 and final independent review pending. Evidence tests22,
capture tests6; original286 retained. Source-only diagnostic addition, not a clipping fix.
No native App build/launch. No changes to existing observers or page layout.
Ruling: reserve attempts before reads; failed acquisition does not refund a slot.
This follows the approved contract; incomplete samples remain evidence gaps.

## 2026-11-11 — ordinary-host baseline comparison passed

User approved one ordinary-host comparison using the identical command below.
Precheck: Windows administrator-role check False; no dotnet/testhost/vstest.console
processes observed; worktree clean. No timeout or system settings changed.
`dotnet test clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore`
completed with exit0: total286, passed286, failed0, skipped0, test duration403ms.
No native App build or launch and no dependency restore. This establishes a passing
portable baseline in ordinary-host context; it does not establish the exact cause
of the earlier restricted-context connection timeout or a product repair.

Tasks1/2/3 remain pending; next is Task1 evaluator tests, then RED→GREEN under the
approved plan. Preserve original failed-run evidence. This comparison approval was
for one run, not unlimited host-context retries. Final review remains pending.

## 2026-11-11 — approved, baseline runner blocked

Plan: ../superpowers/plans/2026-11-10-windows-c4-email-geometry.md.
Implementation baseline: e4c42a33a9929152affb221529b67e873c21d042.
User approved sequential inline implementation, portable tests and a final independent
review. No native App build/launch, restore/install or system changes authorized.

Verified existing linked worktree and clean development branch. Shared interfaces
between Tasks1/2/3 were checked: evidence, coordinator and receipt session agree.
Ruling: use PowerShell/manual task ledger rather than bundled Bash helpers because
the repository requires PowerShell; bookkeeping drift is the risk, mitigated by
explicit baseline, task state and test results recorded here.

Command (worktree root):
`dotnet test clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore`

Core and Tests compilation completed. VSTest found one test assembly, then reported
that vstest.console failed to connect to testhost within 90 seconds; run aborted,
exit1. No pass/fail test count was produced. This is not RED evidence for new code
and not evidence of any test assertion failure. No geometry source/tests written.

One initial tool call was aborted without usable output; a read-only process check
found no dotnet/testhost before the recorded run. The recorded run was not retried.
Read-only ordinary-host observation identified dotnet5384, dotnet17768, testhost11224;
after the runner aborted, a second check found all three absent. No processes killed.
Restricted-context process enumeration did not show them while ordinary-host did;
this observation alone does not prove the connection timeout's cause.

Tasks1/2/3: pending. Final review: pending. C4: NOT_ACCEPTED, launch allowance0.
Next recommendation: one ordinary-host baseline comparison after user direction,
same command/no restore, unchanged timeout, no security-setting changes. Do not
implement against an unverified baseline or silently classify timeout as a passing test.
