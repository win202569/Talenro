# Account alignment experiment execution checkpoint

## 2026-11-12: candidate implemented, offline verification complete

Resumed from1f195f0f on the existing clean coordinator worktree. Sole application
delta: AccountGeometryPanel HorizontalAlignment="Center". MaxWidth, padding,
spacing, email direction, Login, diagnostic adapters and process gates unchanged.
The original Account canonical hash is retained: validator first asserts the exact
named StackPanel and Center value, then removes only this attribute from a clone
before the existing diagnostic-name normalization and baseline comparison.

Actual ordinary-token runs (no retries, restore, native build or application launch):
- RED:19 total,18 passed,1 expected failure in Account OnlyApprovedPageChangesArePresent,
  expected Center, actual null; exit1. Ten independent mutation cases and Login passed.
- GREEN after the single XAML attribute:19/19 passed, exit0,0.6068s VSTest time.
- Geometry regression:97/97 passed, exit0,5.7886s VSTest time, including existing
  build-configuration evaluation/isolation checks. These do not build the native app.
- RED produced xUnit2031; changed to Assert.Single's predicate overload, preserving
  the same cardinality check. GREEN/regression emitted no compiler/analyzer warnings.

Exact commands remain in the approved plan; full-project suite not run/claimed.
Tests are source-scope guards, not proof that native layout clipping is fixed.
Source and copied test fixture SHA256 matched:
- Account: DFBF6C94A68E25573AED467013A9A8723B39045478E684FE010AFA283E23EE1C
- Login: 89BE7A354A3041CA119FF95DB5B4B72A8CFADDDD92EFE3A9B7EAE5F42731DBBD
Previously fixed EXE/DLL and run geometry hashes matched the preserved inventory.
Git diff --check passed (Git reports normal LF-to-CRLF checkout warnings).

Candidate source/tests commit: e6c77ddf2bdce4db6033f6402ca3070eae9d2176.
Independent read-only review account_alignment_candidate_review:0 Critical,
0 Important,0 Minor; suitable as an experimental checkpoint only. Reviewer also
independently checked source/copied fixture hashes and all five focus items.
Native RTL/LTR/narrow-window rendering,
new candidate artifact build/binding and C4 remain NOT_ACCEPTED/unverified.
No main merge, no raw evidence deletion or commit. Keep branch and worktree.

Review dispositions/rulings: native layout/width/wrapping effects remain unverified
and block a repair claim; full-suite health remains unverified under the approved
targeted scope. Reviewer did not rerun tests or revalidate historical artifacts:
parent-observed fresh runs/hashes above provide those limited checks, not broader
independent runtime evidence. No deferred minor findings. PowerShell/manual ledger
substitutes for Bash-only wrappers (bookkeeping requires explicit checks); retained
ignored scratch preserves evidence but is not cloneable. Branch integration is not
being completed: keep this checkpoint, without full-suite/visual acceptance or main merge.

## Ordinary-token comparison: baseline passes

User approved one ordinary-PowerShell diagnostic comparison. Executed through the
approved external tool context, not an administrator shell: token check returned
Administrator=False, related process count0, unique log absent, and timeout/debug
override variables absent. An initial command had a PowerShell interpolation parse
error before any statements ran; corrected syntax did not repeat a test launch.

One actual test invocation used the same project/filter/--no-restore/normal console
logging as diagnostic01, with unique baseline-ordinary-01.log. UTC start
19:07:14.2113895, end19:07:19.0369125: 8 passed,0 failed,0 skipped, exit0;
VSTest total0.7378s, whole command~4.8s. Subsequent external CIM query matched0
dotnet/testhost/vstest.console processes. No clients/windows tracked changes.

This supports an execution-context-related local communication restriction in the
failed run, not broken baseline assertions. Exact deny policy/component remains
unidentified; no claim that all environment differences were controlled. No timeout,
firewall/UAC/security-setting or dependency changes. No app launch or native build.

Closed local diagnostic log hashes:
- host baseline-ordinary-01.host.26-11-11_23-07-18_25196_5.log:
  1580083E6B3CB4615C99316545027C153A26A2AE445D6FB05B57217861418AF5
- runner baseline-ordinary-01.log:
  6DD3D2DB8CC70EE17B55DABE8B9E71F32CDBD6B4F966124A5828FB2A56D659CC

Baseline is now verified for these8 contract tests only. Next resume the approved
alignment plan at RED test creation, not by repeating baseline diagnosis. New test
commands must use the approved execution mechanism and ordinary-token check, with
no automatic retries. Candidate still unimplemented; C4 NOT_ACCEPTED. Raw logs
remain local/ignored, not part of remote recovery.

## Diagnostic follow-up: localhost socket access denied

Original session46350 subsequently completed with exit1: VSTest failed to connect
to testhost within90 seconds and aborted. Read-only external process query then
found0 matching dotnet/testhost/vstest.console processes. Application event IDs
1000/1001/1026 in 18:55–19:01 UTC had no matches; this is not proof against every
possible failure.

User then approved one separately scoped diagnostic baseline run, no retries.
Same project/filter/--no-restore, console normal verbosity, plus --diag targeting
ignored plan-local baseline-diagnostic-01.log. No source/settings change.
Session24920 started after19:02:43 UTC. Managed Core/Tests compiled. At19:03:33 UTC:

- Runner PID16132 listened on127.0.0.1:49289 and launched testhost PID9800.
- Host trace emitted preliminary connected messages, but then reported
  SocketException10013 (access denied) from TcpClient.CompleteConnectAsync in its
  communication initialization callback. Preliminary messages alone are NOT a
  successful handshake; runner continued its unchanged90000ms connection wait.
- This local socket error is direct evidence for the diagnostic run, and a
  plausible explanation of the prior timeout, not proof of the prior run's cause.
  Which execution restriction/firewall/security component denies access is unknown.
- Runner also logged loader warnings for Microsoft.Bcl.AsyncInterfaces and
  Microsoft.VisualStudio.TestPlatform.Fakes; their causal relevance is unproven.
  Do not install packages or change dependencies based on warnings alone.

At19:04:27 UTC no final summary/exit had yet been returned. Logs were open; hashing
failed due to sharing restrictions, so no final log hashes are claimed. Preserve
both runner and host logs in the ignored plan directory; do not commit raw logs.
Final collection: session24920 returned exit1, test run aborted after the unchanged
90-second connection timeout; runner logged exit1 at19:05:03.803 UTC. This exceeded
the overall120-second observation budget including compilation; no replacement run
or forced stop was issued. The error reported PID9800 still running at error time;
current residual-process state has not been independently confirmed for this run.
After completion, log hashes were readable:
- host: 23894A5F8926C94825EC8D0471ADF734EDDBEB8C5E849D2C24E22353E99ADF8A
- runner: BBE961AE5C08A79CE441ED08CD643228B8220170950F22BE01DBF3816C2DDDA3
This is a failed baseline, not a test assertion failure or proof of final cleanup.
No automatic retry, force termination, timeout increase, app launch or native build.
Next confirm residual process state read-only; any new test execution outside the
current tool restrictions requires separate approval, not a security-setting change.

The historical checkpoint below describes the earlier state, superseded above.

Plan and inline execution approved. Base: 23b211783a2541b852d0485c999be3686534337c.
Worktree: c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator.
Initial tracked worktree clean; existing linked worktree confirmed, not main.

## Blocked before candidate changes

One baseline command was invoked from the worktree root:

```powershell
dotnet test clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometryAdapterContractTests' --logger 'console;verbosity=minimal'
```

Tool session 46350 produced Core/Tests compilation output and VSTest's one matched
test-file message. No test summary or exit code was returned during the bounded
observation. Session polling continued to report running. No passing baseline,
RED failure or test count can be claimed. A read-only Get-Process dotnet,testhost
query returned no rows; absence there does not prove tool-session completion or
identify the cause. Last timestamp check was 2026-11-11T18:58:55Z.

No test/application source edits, native build, app launch, restore, repeated test
command, forced termination or guard relaxation. Independent candidate review has
not been dispatched because no candidate exists. Test host versus output transport
failure remains unknown; do not diagnose this as application close failure.

## Preserved evidence

Fixed email-geometry artifact hashes rechecked unchanged:
- EXE: CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3
- DLL: 154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9
- Run 55550244cd6449eab7d870733015eb23 geometry.jsonl:
  AE5ABE46357E8E5AB2DF100A09843CB6C62BB4E9E8B865D226FE5619AB97DB1A

## Execution decisions and recovery

- PowerShell/apply_patch bookkeeping replaces Bash-only skill wrappers, respecting
  the project shell constraint. Cost: manual ledger consistency must be checked.
- Approved offline scope is retained; source contracts cannot prove native visual
  behavior. Cost: full-project and native acceptance remain unverified.
- Ignored scratch evidence is retained for diagnosis, not deleted. Cost: local
  scratch remains; it is not restored by cloning and is not a required dependency.

Next: read-only investigation of the existing baseline session/test environment;
do not start a replacement test or application automatically. Resolve the unknown
run state before obtaining any needed authority to resume. Then resume plan at
baseline/RED preparation, not GREEN. No implementation task is complete; C4 remains
NOT_ACCEPTED and live allowance remains 0. Remote recovery restores this checkpoint
and approved plan, not the local tool session or binary/raw records.
