# Account-center single live observation

User explicitly confirmed one launch, zero retries,360000ms cap, normal close,
no forced termination and no automatic acceptance. Source ac205585c2b6c537410195538665c8d34bbc9438;
ordinary token, clean branch, no matching target/prototype process, candidate hashes
and six committed source blobs checked before invocation. Used the committed
observe-email-geometry-account-center.ps1 -RunApproved once in ordinary-token
external execution context. No build/restore/source/security changes.

RunId66db34b22cc748bbb8fdf61f656b879a, PID5232. Started258.6562ms.
UI observed Login, entered offline demo, then Account and plan. One coordinate
navigation attempt left Connection visible; refreshed accessible Account navigation
was then used. No connection simulation/account submission or settings action.
Account email demo@example.com was fully visible in this observed window. Root
1136x620 DIP, Account page916x534 DIP, reported render1/text1, en-US/rtl-mixed.
No window resizing, narrow-layout/LTR/scaling matrix or text entry performed.
Clicked the title-bar close button normally; no process kill, dump or relaunch.

Observer command exit0. Final output: ExitVerified=true, Finalized=true,
Outcome=exited, Reason=agreement, ElapsedMs41337.2835, OverrunMs0. Post-run
target/prototype process query succeeded:0. This run consumed its one-launch
allowance; further launches need new approval.

Important evidence distinction: samples.jsonl has24 rows. Its persisted final
outcome is provisional, Finalized=false, ElapsedMs41313.8014; console final result
after cleanup is Finalized=true at41337.2835ms. Do not rewrite the saved record or
mistake its pre-cleanup outcome for finalization. Geometry file has header,4 samples
and health record: Samples4, Provisional=true, Failed=false, ErrorCodes empty.
Both Login and Account have two samples; ClipChainComplete=false is retained.
GeometryEvidenceStatus remains Unreviewed; Acceptance remains NotAccepted.

Candidate artifact directory:
clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/.
Manifest agrees with fixed candidate path, EXE/DLL hashes, run and360000ms/2000ms
limits. EXE/DLL hashes unchanged after run. Raw records remain ignored/local:

| Relative record under candidate directory | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/66db34b22cc748bbb8fdf61f656b879a/manifest.json | 730 | 1F3776C17516547CF2623568A49238E9FDE6E7D57B977EF3B14FEDCCBA7D130A |
| email-geometry-observations/66db34b22cc748bbb8fdf61f656b879a/samples.jsonl | 41833 | 4C9536622B34DCACC2BA6A24CA12BC4B0EC8701BB6CEB9C818B8A65240E8B5C6 |
| email-geometry-records/66db34b22cc748bbb8fdf61f656b879a/geometry.jsonl | 36103 | FA4E3F6DE3C155549B324FF3AE7717F9988F4F96F769C07825CEC6C2D26F370E |
| email-geometry-records/66db34b22cc748bbb8fdf61f656b879a.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Next: read-only compare saved candidate geometry with preserved baseline run
55550244cd6449eab7d870733015eb23, validate correlation/measurement comparability and
boundary evidence before any scoped layout conclusion. Do not infer full clipping
repair from one visible screenshot or successful exit. No C4 acceptance/main merge.
Fresh clone restores sanitized provenance only, not raw data or screenshots.
