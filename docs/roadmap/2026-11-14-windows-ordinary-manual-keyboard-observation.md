# Ordinary manual-keyboard observation — 2026-11-14

User confirmed uninterrupted prompt visibility and authorized one launch, zero
retries, maximum six minutes. Protocol approved; exact binding and all preflight
checks unchanged. Starting commit06986e953b92746e873a59b0b5b61d1ce4a0fa86.
No source/test/build changes. One launch only; remaining live allowance0.

## Observed route and limitations

- Run cba6b07ba6da43faa802369cf19d30eb, PID24256, retained session90772.
- Ordinary Chinese login displayed; agent navigated Enter demo, observed
  disconnected/lock-off Connection, then Account. Default screenshot1422x815;
  lower account controls already clipped, so optional resize was omitted.
- Agent ended automated input and prompted one titlebar click followed by one
  physical Tab, release and pause, without chat reply. Two read-only captures,
  each preceded by a2-second settle, showed no clear button outline; first
  accessibility focus was pane2. No agent input was interleaved in that block.
- Physical key completion/time was NOT independently confirmed. These captures
  cannot distinguish an uncompleted handoff from a focus/rendering/input issue.
  The short observation interval is a protocol limitation, not evidence that
  manual Tab failed. No reproduced application defect is established.
- On uncertainty agent ended the block, requested no further input, and clicked
  normal titlebar X. No additional keys, retry, force-kill, dump or settings change.

## Results

| Item | Result |
| --- | --- |
| Exact-bound launch and retained-process exit | PASS |
| Visible Account keyboard start/traversal/auto-scroll | UNVERIFIED |
| Placeholder keyboard close/focus return | NOT_RUN |
| Exit-demo reset/re-entry | NOT_RUN |

Persisted outcome sequence27:47094.5603ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned outcome:47105.8991ms, Finalized=true,
ExitVerified=true, OverrunMs0. Entry exit0. These distinct lifecycle records are
preserved, not rewritten. Post-close host CIM target count0; six candidate hashes
match the manifest and prior accepted binding. No broader UI/C4/security acceptance.

Ignored raw receipt directory beneath the ordinary candidate:
ordinary-account-observations/cba6b07ba6da43faa802369cf19d30eb/.
Manifest SHA256 E4FF9AE414CDCA44DAB2AC5128A663999266A4E549C79D16769291670BA0E137.
Samples SHA256 780E150F6A8EFC193E41ED65F74B1F11956EE4C1D0E5CE0DE59C3267E1CC6733.
Raw receipts/binaries are not portable repository content.

Next: clarify whether the user completed that titlebar/Tab handoff before close;
review a reliable non-focus-stealing completion signal before proposing another
run. Do not infer key delivery from elapsed time, alter code, or relaunch under
this exhausted authorization. Existing remaining-acceptance list still applies.
