# State recovery diagnostic test follow-up — 2026-10-10

User requested the next step; this follow-up closes the previously deferred
FutureSlot assertion coverage item only. No product behavior or interfaces change.

`TestRecoverySelect` now checks exact SlotB/SlotA for a future epoch8 in both
slot orders, and exact SlotUnknown for absent or older epoch6 inactive slots.
Expected values are derived directly in the test, not through product helpers.
This catches an omitted FutureSlot assignment, a wrong slot, or a spurious
diagnostic when no future exists.

This is coverage of existing correct behavior, not a product bug fix. No product
code was deliberately broken to manufacture RED, and no behavioral RED→GREEN
claim is made. The earlier review findings and fix evidence remain historical.
No additional independent review was performed for this small test-only change.

Verification uses the same pinned Go1.26.5/offline/readonly PowerShell settings
as the approved state-recovery plan. Commands:

```powershell
& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecoverySelect$' -count=1
& $recoveryGo test ./internal/nodecontrol/staterecovery -count=1
```

Results (session76475): targeted test exit0, 3.481s; full staterecovery package
exit0, 2.877s. No wider regressions, DB/Docker,
hardware/reboot, race or whole-repository tests were run for this test-only edit.

Baseline comparison: 387 Go/SQL/module files; only
`internal/nodecontrol/staterecovery/recover_test.go` changed, other386 unchanged.
gofmt applied; git diff --check exit0. HEAD remains
`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`. No staging, commits, pushes or cleanup.
Production compatibility and actual platform acceptance remain UNKNOWN/open.
