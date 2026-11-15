# Devices-focus observation stopped before Tab — 2026-11-15

User authorized one launch/zero retries/360s maximum, at most three Tabs with60s
phone acknowledgements and no button activation. Starting checkpoint
3ceeec4442dbf3950fb53dd91a53b431c9e76f6d. No code/test/build changes.
Ordinary entry retained all preflight/exact-binding/safety checks.

Run028c06c63f5046a5a99194150baa78f3, PID5520, retained session22046.
Agent observed ordinary Login, navigated Enter demo, observed disconnected/lock-off
Connection then Account at default1422x815 screenshot bounds. No resize.

Step0 numbered phone question requested one titlebar click, then completion.
Prompt clock15:19:31 UTC (19:19:31 Asia/Tbilisi), deadline60s later. Matching
"第0步完成" arrived while monitoring; first following clock15:20:27 UTC, within
deadline. However, both the read-only post-ack capture and one additional capture
displayed chat content rather than the selected demo window. No reliable target
focus evidence existed. Agent stopped before issuing step1, announced cancellation,
then freshly enumerated/selectively activated the ordinary target only for normal
close. Fresh screenshot showed Account; agent clicked its normal titlebar X.
No inputs were sent to chat, no Tab requested/sent, no continuation after recovery.

Cause unresolved: target capture/occlusion/focus handling or user interaction may
be involved, but screenshots alone do not establish any one cause or show which
device sent the reply. Do not blame user or infer a product keyboard defect. The
previously verified Settings -> MFA route remains valid for its recorded run.

Persisted outcome sequence72:137804.7464ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned:137825.2404ms, ExitVerified/Finalized=true,
OverrunMs0. Entry exit0. Post-close matching CIM count0; six artifact hashes match
manifest. Launch/exit PASS; all three Tab steps NOT_RUN in this run; Devices focus
remains unverified. No full UI/C4/security acceptance or retry. Allowance0.

Ignored receipts under ordinary candidate:
ordinary-account-observations/028c06c63f5046a5a99194150baa78f3/.
Manifest SHA256 AA52D6F311CF32988E632DB4BC9C400B5C448C8FD4BEFFFEF0BA034E43775917.
Samples SHA256 4DB12DF4FB8D975E9C29E7768B7F402BF8EA712545AFF35063B8CE8D28D6E8FC.

Next: read-only clarification of whether desktop focus changed when responding,
and review target-capture evidence. Do not automatically relaunch, reactivate during
the keyboard sequence to mask uncertainty, or relax the stopping condition.
