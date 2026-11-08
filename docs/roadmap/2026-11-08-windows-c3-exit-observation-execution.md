# C3 exit observation execution

Plan: docs/superpowers/plans/2026-11-08-windows-c3-exit-observation.md.
User approved afterfeb90856. Starting HEADfeb9085698e52a49c275ba7babfdac53d7753e72.

Task1 complete: clean expected development branch, AdministratorToken=False,
broader CIM/name checks found no demo process; supported UI inventory found no
demo window. DLL/EXE hashes match the plan exactly. Budget0/1 before launch.
Task1 identities feed Task2 observations, which feed Task3 portable result.
Ruling: use the approved verification-only sequential workflow and portable
ledger, without Bash helpers, implementation tests or a new review cycle; no
new product-code assurance is claimed.

## Outcome — UI input blocked; Login close observed, C3 route NOT_RUN

Budget1/1 consumed, no retry. No build/test/source/settings changes, force-kill,
debugger attachment or dump. All timestamps2026-11-08 +04:00 unless stated.

- APP_START PID19528, start21:54:02.2814461, launcher15376; original Process
  handle acquired successfully. Exact executable and arguments match the plan.
- Supported window inventory found one matching diagnostic window within startup
  allowance. Window title: zh-Hans/long-country inactive, root1136x620,
  page1136x534, render1/text1 (self-reported only). Login/offline notice observed.
- After a text-only accessibility observation, the first click on Enter demo
  returned `coordinate input geometry is unavailable`. The executor stopped the
  route under the approved error-stop rule, without retry. A subsequent combined
  screenshot/text observation still showed Login; no entry or model transition
  was confirmed. This is a tooling/input failure, not an app defect finding.
- Normal Alt+F4 close requested at17:54:35.260Z (21:54:35.260+04:00).
  Original launcher session67774 returned explicit APP_EXIT PID19528 ExitCode0
  at21:54:35.5927660, elapsed33.3091081s since launch, about0.333s after request.
  This is an application receipt, not merely a successful shell exit.
- Immediate UI inventory still listed the closing window; later refreshed
  inventory returned no demo window. At21:54:39 CIM/name queries returned no
  records, but the shell had exit1 from no-match Get-Process. A clean independent
  check at21:54:50.1121855 explicitly returned CimDemoCount=0, NameDemoCount=0,
  shell exit0. No residual demo process was observed.

Task2 completed only as an aborted observation attempt. C3 route/search/dialog/
final Nodes state remain NOT_RUN; the original C3 shutdown failure was neither
reproduced nor disproved. Scoped Login close passed with an actual exit0 receipt.
The retained-handle observation worked for this short Login-only session; do not
attribute the old failure to the old observer from this non-equivalent route.

Next: prepare an amended single-attempt route with screenshot-backed state before
the first click (and fresh state after every action), or explicitly agree on
keyboard navigation. The missing geometry followed a text-only capture, but that
association is not a proven tool root cause. No automatic relaunch is authorized.
All Task11/12 and44/34/6/217 NOT_ACCEPTED and C4-C6 gaps remain unchanged.
Task3: portable results/current-status/recovery updated; documentation-only diff
review and git diff --check required before preservation. No independent code
review or renewed automated-suite assurance is claimed by this runtime record.
