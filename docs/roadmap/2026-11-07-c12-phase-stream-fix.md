# Live log sharing regression — offline repair

User continued after5e2aec77; scope is offline regression/minimal repair only.
One RED then one GREEN;10 cases,RED first open-writer case should fail without
launching a child. GREEN uses3 harmless PowerShell fixtures: complete protocol,
invalid-session stop (each5s bound),START then3s cutoff;5s owned exit wait.
New exclusive root .superpowers/sdd/2026-11-07-c12-phase-stream-fix.
No Go/build/live module reads,security changes or extra retries. Preserve old
records. Unexpected result stops. No full-suite/independent acceptance claim.

Cause: parser File.ReadAllBytes opens incompatibly with an existing Write handle.
Production writer remains FileShare.Read; proposed reader uses Read access with
ReadWrite sharing to acknowledge that writer,not to enable additional writers.
Read a bounded snapshot length,dispose reader in finally; protocol/quotas unchanged.
Tests keep the writer open and append across parse calls,check partial completion,
identity/order/digest/quota rejection,then actual owner/session/child integration.
The source exercised by each stage is snapshotted,including RED,for recovery.

RED15cf43 reproduced the exact ReadAllBytes sharing violation in first case;
no child launched. Minimal fix replaces only that read with a read-only handle
sharing ReadWrite,bounded initial-length read and guaranteed reader disposal.

GREEN10e257:10/10 cases passed. Actual session-enabled owner fixtures:
complete EXITED738.2088ms/exit0,invalid-session ERROR BINDING305.6011ms,
cutoff DEADLINE3032.806ms. All3 exits confirmed; complete output exactly matches
12-record input; cutoff keeps one START as incomplete. No real module reads,
Go/build/live invocation or full-suite run. Production writer still denies a
second writer (regression assertion passed). Byte/event limits unchanged.

Author review: initial length is capped before allocation; finite read loop fails
on truncation; reader disposed on every path; protocol/identity/cleanup unchanged.
Scope is this ASCII live protocol/file-sharing integration; no wider parser or
concurrent adversarial cache snapshot guarantee. Portable archive
portable-c12-phase-stream-fix.json includes exact RED/GREEN owner source,receipts,
synthetic inputs and raw outputs. Its4MiB quota fixture is synthetic zero fill.
Old v2 build source binding no longer matches owner.ps1; do not reuse its Live
root or bypass SOURCE_CHANGED. Future real run needs fresh scoped approval and
reviewed source/binary binding.44/34/6/217,NOT_ACCEPTED.
