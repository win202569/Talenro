# Single-Tab focus observation — 2026-11-15

User clarified titlebar location, then authorized one launch, zero retries,360s
maximum under the60s phone acknowledgement protocol. Only titlebar click and one
Tab allowed. Starting checkpoint e9bf2845e7d3cd02255baeb00ffff0eeafc142c7.
No source/test/build changes; exact binding and all preflight gates unchanged.

Run b3aac484bee7448d8db59c6bb3b29de8, PID10276, retained session51434.
Ordinary Chinese Login observed; agent navigated Enter demo, observed disconnected
and lock-off Connection, then Account at default1422x815 screenshot bounds.
No resize. Automated input then stopped until normal close.

## Acknowledged sequence

Numbered prompts used the asynchronous user-question channel; each had one
outstanding action and60s maximum. User answered through that channel during the
active observation. Clock reads bracket processing, not exact transport timestamps.

- Step0 prompt/clock approximately07:51:15 UTC, deadline07:52:15 UTC. User replied
  "第0步完成"; first following clock07:51:33 UTC. Fresh read-only screenshot showed
  Account and pointer in titlebar. No extra activation/click by agent.
- Step1 prompt/clock approximately07:51:46 UTC, deadline07:52:46 UTC. User replied
  "第1步完成"; first following clock07:52:01 UTC. Fresh screenshot visibly showed
  a clear black focus outline around left-sidebar Settings (设置). Account remained
  displayed; no Settings activation, second Tab, pointer scroll or content click.
- Accessibility focused_element still reported pane2. Visible outline is the
  evidence for focus; pane metadata is insufficient to reject it. This agrees
  with the earlier caution about relying on pane/window focus metadata alone.
- Agent announced completion and clicked normal titlebar X. No further tests.

## Results and limits

Titlebar operation confirmed by user; single acknowledged Tab to visible Settings
focus PASS for this route. This does NOT prove account-content button traversal,
reverse navigation, automatic scrolling, placeholder close/focus-return, reset,
screen-reader behavior or broader UI/C4/security acceptance. Those remain NOT_RUN
in this observation. Do not transfer prior diagnostic-binary acceptance.

Persisted outcome sequence46:85277.2109ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned:85298.1559ms, ExitVerified/Finalized=true,
OverrunMs0. Entry exit0. Post-close matching CIM count0; six artifact hashes match
manifest. Exact launch/exit PASS. One launch consumed, remaining allowance0.

Ignored receipts beneath ordinary candidate:
ordinary-account-observations/b3aac484bee7448d8db59c6bb3b29de8/.
Manifest SHA256 6E6DA423EEA65EA543019C09AB3085D42FC4B8772A5A5F6C82D2031691FC6439.
Samples SHA256 0D57DEC73E28A447490C59D1618A090F5BD7204954CA9373F5BC4F8B59412D85.

This run demonstrates timely acknowledgements through the question channel.
It does not prove why previous plain-message replies were late: channel choice,
titlebar clarification and user timing changed together. No platform defect claim.
Next proposed work is read-only remaining-coverage review and a separately scoped
follow-up plan, not automatic relaunch, added keys or a code fix.
