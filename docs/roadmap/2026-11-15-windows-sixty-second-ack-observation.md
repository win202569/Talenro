# Sixty-second phone acknowledgement observation — 2026-11-15

User authorized one launch, zero retries, maximum360s; reduced route only titlebar
confirmation and one Tab. Source1144c0c2e8f8830ec31ffbf62d860ae82dfccc9c.
Prior same-conversation phone readiness and no-app roundtrip existed. No new
source/test/build changes or weaker checks. Run dfc23e08440247ceb4e8c01a06ce41de,
PID17696, retained session10186. Existing entry passed exact binding/preflight.

Agent observed ordinary Chinese Login, navigated Enter demo, observed disconnected
and lock-off Connection, then Account. Default1422x815 screenshot, no resize.
Agent ended automated input and requested step0 titlebar click plus phone reply.
Prompt clock approximately07:46:06 UTC (11:46:06 Asia/Tbilisi); deadline60s later.
Bounded retained-session polls continued; no key/activation/click during waiting.
At07:47:08 UTC no matching acknowledgement had arrived. Agent cancelled step0,
announced no further input and normal close. Actual titlebar operation UNKNOWN;
absence of acknowledgement is not failed input. Step1 never issued, no Tab sent.

During normal-close target activation, tool reported user input detected. Agent
did not reuse earlier coordinates: obtained fresh target screenshot, then normal
titlebar X. This tool notice alone cannot identify the user's action or prove
step0 completion. No restart, repeated key, forced termination or settings change.

Persisted outcome sequence56:105430.4441ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned:105449.5374ms, ExitVerified/Finalized=true,
OverrunMs0. Entry exit0. Post-close matching CIM count0. All six artifact hashes
match the manifest. Launch/exit PASS; step0 confirmation timed out; single-Tab
focus test NOT_RUN. Full traversal/scroll/dialog/reset NOT_RUN; no product defect
or broader UI/C4/security acceptance inferred.

Ignored receipts under ordinary candidate:
ordinary-account-observations/dfc23e08440247ceb4e8c01a06ce41de/.
Manifest SHA256 D9ACFB4DF1F393BF86DE0ABC590821B5D9424C8F65708A087839995783B03667.
Samples SHA256 EF04A41CCDC999968670B3024DA1C547253AB60F728B34C820495882E1A2AC8F.

Allowance consumed, remaining0. Next proposed work: no-app check of message
delivery while the agent is actively observing, plus user clarification of actual
operation/send timing. Prior between-turn30s sample does not establish delivery
during an active turn. Do not assert platform queuing as a proven cause, extend
waits, or relaunch without separate scope/approval. Preserve later clarifications.
