# Independent guardian implementation planning — blocked before implementation

Date2026-11-16. User approved formal design24b3d957. That approval is recorded,
but platform research found an unresolved safety requirement before an executable
implementation plan could be completed. This document is a blocker report, NOT
an approved implementation plan. No program edits, build, test, helper or app run.

## Evidence

Microsoft's [IsWindow documentation](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-iswindow)
warns that a window can be destroyed after a check and its handle reused for a
different window. [PostMessageW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-postmessagew)
addresses a window handle and returns without waiting for message processing;
its signature has no expected process-creation-identity parameter.

Inference: retained original process identity plus a separate window-owner check
does not by itself establish atomic check-and-close of that same window. Repeated
checks, lifecycle notifications and a very short gap have not been demonstrated
to remove the race. A post-dispatch check cannot undo a misdirected request.
This is not a reproduced incident; it is a missing safety proof in the proposed
adapter. Do not substitute CloseMainWindow/another messaging API without verifying
its binding semantics. No claim that every possible external approach is impossible.

Microsoft's [QPC guidance](https://learn.microsoft.com/en-us/windows/win32/sysinfo/acquiring-high-resolution-time-stamps)
supports the planned machine-local monotonic timing basis; it does not resolve
the window dispatch race or guarantee scheduler deadlines.

## Consequence

The approved design requires exact process/window binding, not merely best-effort
checks. Its earlier builder check left concrete API/lifecycle validation to
planning; this research identifies the unresolved part. Design approval does not
authorize silently weakening that requirement. All290/300/360s deadlines, no-kill,
no-retry, existing safety gates and210s phone cutoff remain unchanged. Allowance0.

## Recommended next decision

Seek permission for a design revision exploring a cooperating application-side
close endpoint: private per-run capability-bound channel; the verified application
validates the request and schedules normal close on its own UI thread. Retain
independent guardian timing and original-process exit verification. No public
command endpoint, injected code or external HWND fallback. This is a candidate
architecture, not an established safety proof; channel authentication, endpoint
lifetime, replay resistance and UI shutdown semantics require review.

This would change the candidate application, requiring separately authorized code
changes, isolated build, six new artifact hashes and impact-based revalidation.
Preserve old binaries and their acceptance records; do not transfer old passes to
new hashes. It exceeds the approved no-candidate-change design and must be chosen
explicitly before implementation planning continues.

Alternative: continue read-only research for an external atomic binding mechanism
while keeping the candidate unchanged. Until documented evidence is found, native
close dispatch remains blocked. Accepting an HWND race is not the recommendation
and has not been authorized. No implementation work is permitted by this report.
