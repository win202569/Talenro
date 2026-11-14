# Ordinary exact-binding live observation — partial UI evidence, verified exit

Date2026-11-14 Asia/Tbilisi; checkpoint12066030. User authorized one host-context
run, all exact binding/preflight checks retained, zero retries, maximum6min.
Run f3c2f05a843b42abac087ae32517791f, PID272. One launch only.

## Process result

All entry preflight gates passed and session reached regular sampling (not the
earlier identity abort). Start63.2675ms. Titlebar normal Close before5min.
Persisted outcome sequence90 at173641.0056ms: exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned post-cleanup outcome at173658.9987ms:
ExitVerified=true, Finalized=true, OverrunMs0. Entry exit0. These are distinct
original records; no rewrite of provisional finalization or earlier failed runs.
Successful host CIM inventory after close:0 target/prototype processes; final
window inventory empty. Six candidate artifact hashes still match manifest.

## UI evidence and limitations

| Check | Result and scope |
| --- | --- |
| Ordinary Login | Observed Chinese offline notice, read-only demo@example.com, masked demo password; title Talenro, no LAYOUT TEST marker. |
| Initial Connection | Observed simulated disconnected, country lock off; never connected or opened Nodes. |
| Default Account | Complete email, Plus mock plan,12.4/100 GB usage and2/5 activated-device text visible. Lower actions extend below viewport. |
| Qualitatively narrower Account | User resized horizontally and confirmed done; email remains complete, usage wraps legibly, device count readable. No exact client/root dimensions claimed. |
| Keyboard focus/scroll route | UNVERIFIED. First Tab call reported user input; later refreshed observations and Tab attempts reported window/pane focus without a visible button focus outline. Cannot attribute conclusively to app defect versus input automation/focus behavior. |
| Placeholder dialog/focus return | NOT_RUN; stopped after keyboard check could not be established. |
| Exit-demo/reset/re-entry | NOT_RUN. |
| Normal process close | VERIFIED by original retained observation session and final exit0, not merely later absence. |

No pointer scrolling performed by agent. User was asked not to scroll or click;
cannot independently reconstruct all user input. Clicked account heading once to
establish content focus after window-focus report; no button focus progression
established. Did not activate MFA, Delete, connection or settings actions. Did not
prolong interaction or retry launch. Initially announced narrow check would be
skipped while awaiting reply; subsequent observed resizing/user confirmation
provided qualitative evidence, not exact viewport acceptance.

Captured window images were approximately1422x815 initially and761x815 after
resizing; these are screenshot bounds, NOT root/client DIP. No1024x768,
1440x1024,620–660 root-width, DPI, RTL or geometry-receipt claim. Ordinary build
does not emit diagnostic geometry. No new configuration/code/build changes.

## Retained evidence

Ignored candidate ordinary-account-observations/f3c2f05a843b42abac087ae32517791f/:
- manifest.json SHA25622CEF76036560F1C6B2CF4F75C6633460E55841012E7AA756A118A7D123F3636
- samples.jsonl SHA2567A76218D78E8CF5B00CBA50E504CFC4B20526A7F544988457A0DC5E38506FFEC

This run supports exact-binding operation and normal exit on this candidate plus
the observed default/narrow qualitative display only. It is not full C4/UI/
security/Release acceptance. Prior failures and deferred Minors remain intact.
Live allowance spent. Suggested next step: read-only consolidate remaining
keyboard/dialog/reset coverage and investigate focus evidence without launching.
Any further live attempt requires new explicit approval. Clone restores sanitized
record/source, not screenshots/raw receipts/binaries.
