# Ordinary Devices -> MFA reverse-focus observation

Date 2026-11-15. Approved one launch, zero retries, maximum 360s;
[approved scope](../superpowers/specs/2026-11-15-windows-reverse-focus-design.md).
Source checkpoint 084e448b716ad8d03f4d05a9149cff631af4cd50.
Run 2602b503efa6498e947ba55082254ce6, PID 17784; original session 98376 completed.

## Observed result

PASS for this single reverse pair: three individually acknowledged Tabs showed
Settings, MFA, Manage devices; one acknowledged Shift+Tab returned a clear visible
outline to MFA. No activation, extra key, dialog, resize or pointer scrolling.
This does not accept other reverse edges, Delete/Exit, dialog return or reset.

Preparation observed ordinary Chinese offline/read-only Login, Enter demo,
disconnected Connection with country lock off, then Account at default size.
Screenshots were 1422x815, not a measured root/client DIP size. After preparation
the agent only read the window during the manual block. User remained on demo
and acknowledged each step by phone, all within its 60s response allowance.

| Step | Clock sampled after reply, UTC | Fresh visual observation |
| --- | --- | --- |
| 0 titlebar single click | 15:43:38 | Reliable Account view, pointer at titlebar |
| 1 one Tab | 15:44:12 | Settings outline |
| 2 one Tab | 15:44:39 | MFA outline |
| 3 one Tab | 15:45:12 | Manage devices outline |
| 4 one Shift+Tab | 15:45:54 | MFA outline returned; Devices outline removed |

Clock before step0 prompt was 15:43:05. Other prompts followed the preceding
capture in the tool transcript; no fabricated exact prompt timestamps. Step1's
reply used Chinese numeral wording (第一步完成), an unambiguous matching step.
Step4 was issued after Devices evidence and before the 220s ceiling. No shortened
wait, repeated/batched key, focus hunting or deadline extension. Accessibility
focus still reported pane2, not reliable as the sole focus verdict. No viewport
movement was needed on the reverse pair; reverse scrolling NOT_EXERCISED.

## Exit and artifact evidence

Immediately after final capture, normal titlebar X requested; original observer:

- Persisted sequence111: elapsed215938.1875ms, exited/agreement, ExitVerified=true,
  Finalized=false, OverrunMs=0.
- Returned sequence111: elapsed215960.4329ms, exited/agreement,
  ExitVerified=true, Finalized=true, OverrunMs=0.
- Entry exit0; post-close host CIM matching target/prototype count0.
- All six candidate SHA256 values matched the unchanged entry policy after close.

Preflight exact binding, launcher pin, ordinary token, source/clean-branch,
artifact/path/reparse and zero-target checks were retained. No build, restore,
code change, synthetic test, security-setting change, force kill or retry.

Ignored receipts under candidate directory:
`ordinary-account-observations/2602b503efa6498e947ba55082254ce6/`.

| Receipt | SHA256 |
| --- | --- |
| manifest.json | 5C51E3B2B261AF6AA54247B11E2C483F0624BC5288EA30C6880CE1AE8ADDF530 |
| samples.jsonl | 5511653FF66D536B8A38EBCF7E4422913695B2DB9DA23E72FEB3830DDEC852B8 |

Raw receipts remain local/ignored; sanitized record is portable, not a substitute
for replaying environment acceptance on another computer. Allowance consumed, 0
remaining. Delete/Exit focus-only, remaining reverse edges, dialog keyboard close
and returned focus, reset, exact geometry/DPI/RTL and broader acceptance remain
separate. No automatic next launch.
