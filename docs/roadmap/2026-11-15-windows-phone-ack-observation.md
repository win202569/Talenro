# Phone acknowledgement observation — 2026-11-15

User approved revised protocol, sent phone readiness message in this conversation,
then authorized one launch, zero retries, maximum six minutes. Starting source
checkpoint716a19582b66b54d5f8b9d9de74f736d5495c8b8. No source/test/build changes.

Run b2d9a6029bb243f8bdd5bd9f73d1d609, PID25972, retained session24246.
Existing ordinary entry passed unchanged preflight/exact binding. Agent observed
Chinese ordinary Login, navigated Enter demo, observed disconnected/lock-off
Connection, then Account. Default screenshot1422x815; resize omitted.

Agent ended automated input and issued step0: one titlebar click, then phone
reply "第0步完成", maximum30s. No matching reply arrived during the monitored
deadline; agent cancelled the step and requested no more input. Actual titlebar
operation UNKNOWN. No Tab or other numbered key was requested or sent. Keyboard,
auto-scroll, dialog and reset checks NOT_RUN; no product failure conclusion.
Readiness delivery beforehand does not establish timely in-run delivery.

Post-run clarification: user sent late "第0步完成" and explicitly confirmed the
titlebar had been clicked before the window closed; only the phone reply was late.
Final interpretation: step0 performed per user report, acknowledgement timed out.
This supersedes the in-run UNKNOWN above without changing raw records or advancing
the cancelled step. Keyboard checks remain NOT_RUN; no product failure inferred.

After cancellation, first read-only close-preparation capture unexpectedly showed
chat content, so it was not used for target input. Agent enumerated windows again,
selected the unique ordinary candidate, activated it solely for normal close,
captured fresh target screenshot and clicked its titlebar X. No input was sent
to chat or settings. No relaunch, key retry, force-kill or security change.

Persisted outcome sequence45:83623.6084ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned:83645.9452ms, ExitVerified/Finalized=true,
OverrunMs0. Entry exit0. Host post-close CIM matching targets0; all six artifact
hashes match original manifest. Exact launch/exit PASS, UI coverage not advanced.

Raw receipts remain ignored beneath ordinary candidate directory at
ordinary-account-observations/b2d9a6029bb243f8bdd5bd9f73d1d609/.
Manifest SHA256:9D536CAF4D3BC848D0E18EDE71263056265A10D5457FFE82C031C8BEE8C8E59F.
Samples SHA256:7544B5EC049D09C83F55B47EFE2EB3422000F0ED5E288F71D45559CB9A20FD4E.
No broader UI/C4/security acceptance. One-run authorization consumed, allowance0.
Subsequent no-app roundtrip sample:07:43:12 to07:43:42 UTC, approximately30s,
including human response and agent scheduling, not pure transport latency.
User then requested60s waits and reduced per-run coverage with unchanged total
cap/safety gates; see revised manual-keyboard protocol. No new launch occurred.
