# Hardlink context comparison — stopped by non-admin guard before child launch

## Approved boundary and actual result

After8f85c90f,user approved NEW maximum1 PS7 child through the tool-approved
outside-sandbox path,120000ms,no retry,using the unchanged minimal probe.
Approval explicitly required the expected current user and a non-administrator
parent token BEFORE launching any child or creating any probe fixture. No RunAs,
ACL/ownership/security changes,alternate API,cleanup or original test replay.

The ignored orchestration script was parsed before execution. One tool approval
request ran that parent script;the recorded username matched,but the supplied
parent token had AdministratorRole=true. GuardPassed=false,ChildLaunched=false.
The parent wrote the separate metadata receipt then threw the guard-stop error.
The tool command exited1;that is NOT a diagnostic child exit. Child ExitCode,
ElapsedMs,TimedOut and stderr fields remain null,Rows empty. No child stdout/
stderr receipts exist because no child was launched. No new fixture,file or
hardlink attempt occurred. The earlier normal-context fixture/receipts remain.

Thus one context request/guard check occurred,but0 of1 proposed diagnostic
children launched. The request is closed under the no-retry constraint;do not
reuse the proposed child cap as authority for another context request. No guard
removal,RunAs request,alternate route or elevated diagnostic fallback occurred.
The tool's approval path itself supplied the administrator token;the agent did
not start the actual diagnostic in that token or alter the user's permissions.

## What this does and does not establish

The available approval route does not satisfy the approved non-admin comparison
on this invocation. We still have only the earlier normal-context hardlink
denial(native5),not a second link outcome. This cannot isolate sandbox restrictions
from token/effective-access differences or explain the original slot7 failure.
No link creation success,repair,archive assertion or test acceptance follows.

Continuing needs a meaningful user choice:guidance for one manual normal-user
terminal execution,or NEW explicit permission for at most one administrator-token
minimal diagnostic(120s,one file/one link attempt,no retry,ACL/security change or
suite replay). Neither is automatically authorized. No diagnostic source change
is needed to record token role,but the current parent guard must NOT be bypassed
under old authorization. Keep a separate record for any future approved action.

## Preservation and gates

Probe hash unchanged:
`56273A58C4E91ED7D1A41AFE8F97C198A1C829624C1C1521DF27FF5B2DB32407`.
New one-entry request/guard index:
`45E70AA65843D11E814E3E361C6A57924E429FF8EA22182F82422175CE53269E`.
Earlier normal probe index unchanged:
`36E9CC361787F65E589DAC1B332E70CB7BE53575D09261C856A23429477BEE1A`.
Partitions index unchanged:
`B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1`.
The metadata,guard script/progress stay ignored/local under
`.superpowers/sdd/2026-10-17-synthetic-capture-hardlink-context`.
Portable report/status/runbook preserved on development branch,not main.

Product/test/probe unchanged. Partition7/10 spent,8..10 frozen,new review
unstarted;original6/12 spent,7..12 frozen,original review unstarted;all earlier
diagnostic budgets spent. No main merge/full acceptance/live/security/vendor
clearance,Stage B/I2/I3/full C12 or expanded permission. Cloning is not launch
authority. Stop current turn for execution-context direction.
