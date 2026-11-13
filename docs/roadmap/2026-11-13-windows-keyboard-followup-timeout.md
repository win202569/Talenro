# Keyboard-only follow-up — timed out, checks not completed

Recorded on2026-11-13 Asia/Tbilisi after resuming the prior approved observation.
Source/document baseline6fd63e52. User approved one launch, zero retries, at most
six minutes for keyboard-only dialog return focus and automatic scrolling.
Fixed account-center entry launched once: runa9fb7f631688418ca06806aaaaeb0527,
PID17380. The user confirmed window visibility and later reported resize complete.
The agent did not verify the resized condition or execute the keyboard checks
before the observation deadline. This is an incomplete observation, not a user
fault or a demonstrated application focus/scroll defect.

## Verified on resume

- Original execution session43324 was unavailable: Unknown process id. No terminal
  exit code or finalized post-cleanup result could be recovered through that session.
- Saved outcome sequence182: elapsed360000.835ms, Outcome=timed-out,
  Reason=deadline, ExitVerified=false, Finalized=false, OverrunMs1.
- Geometry contains Login snapshots1/2 only, Account0; health Samples2,
  Provisional=true, Failed=false, ErrorCodes empty. Capture health does not override
  the process timeout or establish completion of either requested UI check.
- Current authoritative CIM query succeeded, matching target/prototype count0.
  This proves current absence only, not exact exit time, exit code or graceful exit.
- No new launch, forced termination, UI action, build/test or source change on resume.
  All raw records remain unchanged. Live allowance consumed; C4 NOT_ACCEPTED.

## Local inventory

Relative to
`clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/`;
RUN meansa9fb7f631688418ca06806aaaaeb0527.

| Path | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/RUN/manifest.json | 730 | 90F7BE9CE475679791E06F04505CC529A344ABD20802BCCAF63FD204F46B55A3 |
| email-geometry-observations/RUN/samples.jsonl | 345801 | 9C3CE52A5360A415EF8A5ABEB41F6D0DD51B850D558529B09CD36C618F8F5892 |
| email-geometry-records/RUN/geometry.jsonl | 18150 | C94B427189EFFD42CE5551FCDBE1FD4D67F7645C3857EF5B0334911AFEC30915 |
| email-geometry-records/RUN.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Both keyboard-only items remain UNVERIFIED. Prior lower-controls evidence is
unchanged. Before any further run, agree an uninterrupted interaction window and
a fresh bounded approval; never resume UI actions from this expired allowance.
Do not repeatedly relaunch to compensate for delayed handoffs. Cloning restores
this sanitized record, not ignored receipts or binaries.
