# Email geometry observation — attempt history and completed live run

## Latest: separately approved ordinary-terminal run completed

Source checkpoint82104a772b68b39fd57126405cae8c62852e4560. User separately approved
one new launch, zero retries and presence for about6 minutes. Read-only diagnosis
had reproduced CIM AccessDenied in the agent command environment while token,
branch/clean/source/hash/destination checks passed. User's ordinary PowerShell CIM
comparison succeeded with0 matching targets. This supports an execution-environment
access difference, not a proven specific Windows ACL cause; no permissions changed.

Initial user command could not resolve pwsh: command-not-found, empty LASTEXITCODE,
no script invocation/app launch. User verified the existing Codex bundled PowerShell
by absolute path (version7.6.9), then invoked the unchanged protected entry once from
the ordinary terminal. No install, PATH edit, elevation or guard bypass.

RunId55550244cd6449eab7d870733015eb23, TargetPid6820. User supplied terminal output:
start elapsed172.2566ms; final sequence71,Outcome=exited,Reason=agreement,
ExitVerified=true,Finalized=true,ElapsedMs=135667.3054,OverrunMs=0; entry exit code0.
GeometryEvidenceStatus=Unreviewed,Acceptance=NotAccepted. This is user-supplied final
terminal evidence, distinct from the independently read persisted provisional outcome
(135627.5072ms,ExitVerified=true,Finalized=false). Do not rewrite the provisional file.

### UI observations (UTC,2026-11-11)

- User reported window appeared; no precise user startup timestamp was supplied.
- 18:44:39.837: unique fixed-artifact window returned by supported window inventory.
- 18:44:46.409: Login accessibility state obtained. First screenshot showed the
  occluding Codex window, so it is NOT Login visual evidence. After activating the
  selected demo window, a fresh screenshot showed Login and its visible demo email.
- 18:45:02.178: clicked Enter demo once (explicit offline/no-account-submission UI).
  Refresh initially lagged, then confirmed Connection landing page. No connection
  control was touched. Account navigation was selected once at18:45:12.285.
- 18:45:17.158: Account screenshot shows email left-side clipping at initial top;
  accessibility exposes AccountGeometryEmail and vertical scroll value0. Root1136x620
  DIP,page916x534 DIP,render1,text1. This reproduces a visual symptom, not its cause.
- 18:45:26.730: one normal titlebar close click. Immediate inventory still showed
  window;18:45:42.506 inventory no longer showed it. No repeated close/force termination.

Observer directory creation18:43:12.646Z and final elapsed~135.7s place normal close
within300s and final observation within360s; last samples write18:45:28.371Z follows
close by~1.64s. Directory time is not an exact launch timestamp. Startup-within30s
was NOT independently established: agent's first window observation was later.
Do not claim every operator timing requirement passed or full scenario acceptance.
No credential entry, search, scrolling, connection, node selection or settings action.

### Preserved record inventory

Relative to the fixed email-geometry artifact directory; run ID as above. Metadata
and hashes rechecked after user confirmed finalization; values unchanged from the
first post-close inventory. No schema/completeness or geometry interpretation yet.

| Record | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/<run>/manifest.json |715|BE56A8778568679C2A2545147A44F62C9DE51691476F55F31B28C2186212008C|
| email-geometry-observations/<run>/samples.jsonl |132108|A60840C9A21A469F9E41A3AC29504652586E684C1939BF0BBA7132C4B81118FE|
| email-geometry-records/<run>/geometry.jsonl |36114|AE5ABE46357E8E5AB2DF100A09843CB6C62BB4E9E8B865D226FE5619AB97DB1A|
| email-geometry-records/<run>.claim |0|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

Manifest binds the fixed new EXE/DLL hashes and en-US/rtl-mixed arguments as designed.
Raw files and tool screenshots remain local/ignored or conversation evidence, not
committed. A fresh clone restores this record, NOT the raw evidence or screenshots.
Further evidence transfer needs its own scoped decision; do not fabricate or rerun.

New run budget consumed:launches1,retries0,remaining0. Next proposed work is read-only
schema/run correlation/completeness and geometry/visual evidence review of this run,
without relaunch or code changes. C4 remains NOT_ACCEPTED; clipping is not repaired.

## Historical: first attempt blocked before launch

User approved the single attempt and presence after reading the bounded plan.
Source checkpoint17778db8cd021c463e78a42a7b02648c625693f3, coordinator branch,
clean worktree. Immediately preceding ordinary-token check at
2026-11-11T18:36:05.0062491Z returned Administrator=False.

Invoked the approved entry exactly once via pwsh -NoProfile -File
clients/windows/tests/observe-email-geometry.ps1 -RunApproved.
Entry exited1 after about1.85 seconds, with final sanitized summary:

- RunId:1362f967af8440d9b0cbefb6847caebd
- Outcome:Failed
- OperationFailures:[Preflight]
- CleanupFailures:[]
- GeometryEvidenceStatus:Unreviewed
- Acceptance:NotAccepted

The top-level preflight failed before module import/resource acquisition/launch.
No application launch, UI input, retry, forced termination or settings change.
No Login/Account screenshots or timing/exit evidence acquired. Window inventory
was used only to establish UI-tool readiness before invocation; unrelated window
titles are intentionally omitted from this record.

Read-only post-attempt checks found no observer run directory, app geometry run
directory or persistent claim for this run ID. Worktree remained clean and on the
expected branch before documenting the result. No raw record was created/modified
or deleted by hand. The reported category does not distinguish the exact failing
preflight gate: do not infer administrator failure, hash mismatch or existing process.

Single approved attempt consumed; launches0, retries0, remaining run allowance0.
This is not a process-exit/geometry result and not acceptance. Next proposed work:
separate read-only gate/environment diagnosis, without invoking RunApproved or
relaxing checks. A later launch requires new bounded approval after the cause is known.
