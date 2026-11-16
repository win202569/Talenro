# Exit focus observation interrupted — not accepted

Read-only reconciliation on 2026-11-16 (Asia/Tbilisi), for the single launch
approved on 2026-11-15 under the
[Exit focus plan](../superpowers/specs/2026-11-15-windows-exit-focus-design.md).
Source checkpoint b7aa8ca600a524d0e597d9c6e43b1243ea894fa2.
Run d2a50105f2ff463d856c166e0bc69068, PID14600, original session27326.

## Evidence and limits

Ordinary Login -> disconnected/lock-off Connection -> Account was observed.
Step0 titlebar prompt followed clock15:59:30 UTC on November15. User replied
“第0步完成”, but no post-reply capture or next prompt was made before interruption.
Delivery latency cannot be established from available timestamps. No Tab prompt
or Tab action is recorded. Steps1–6, including Exit focus, are NOT_RUN.

On resumption November16 at16:44:57 UTC, original terminal session was unavailable
(Unknown process id27326). This does not prove the application's exit. Preserved
receipt gives the authoritative observed deadline result:

- Last pair sequence181 elapsed358984.9467ms: identity verified; managed
  HasExited=false/WaitSignaled=false; native wait timeout and exit code259.
- Persisted outcome sequence182 elapsed360002.6553ms: timed-out, reason deadline,
  ExitVerified=false, Finalized=false, OverrunMs=3.
- Returned finalization and entry exit code are unavailable, not inferred.
- No normal-close request before300s is recorded. The planned closure deadline
  was not met; this run is NOT ACCEPTED. Do not claim normal exit or focus failure.

Read-only host CIM query on November16 found matching target/prototype count0.
This is current absence only, not retroactive proof of timely/normal exit; actual
exit time/method and interruption cause are unknown. No force kill, retry, new
launch, code/build/test or security-setting change during reconciliation.

## Preserved receipts

Local ignored candidate directory:
`ordinary-account-observations/d2a50105f2ff463d856c166e0bc69068/`.

| File | SHA256 |
| --- | --- |
| manifest.json | 820ACF3FB5C81BAA455359844775D764887D995A93FA03F5E9810E46FFA0F38D |
| samples.jsonl | BAB5F4E71A440321EB0509E8E2FEFBB9AF7D5B831FA0AEF843955B414175A34D |

Original records untouched. Post-run artifact hashes were not revalidated in this
reconciliation. Earlier Delete/Language/Devices->MFA passes remain unchanged.
Exit focus, remaining reverse paths/dialog/reset remain pending. Launch allowance
consumed,0 remaining. Stale acknowledgements must not resume this run.

Next safe step is a no-launch review of interruption handling and normal-close
ownership before proposing another live observation. Any implementation or new
launch requires its own scoped authorization; do not silently add a watchdog,
weaken safety gates, or assume the observer closes the application on deadline.
