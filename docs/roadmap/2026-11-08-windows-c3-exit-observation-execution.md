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

## Screenshot-backed attempt — separately approved after6451e456

User accepted the recommended screenshot-first single reproduction. New budget
0/1 before launch; original1/1 stays consumed. Preflight: ordinary token, clean
worktree, both process counts0, no demo window and exact DLL/EXE hashes match.
Same route/deadlines and retained-handle observer; capture screenshot and text
before every click, refresh after each action. No additional implementation scope.

### Outcome of screenshot-backed attempt

New budget1/1 consumed. APP_START PID4124 at2026-11-08T21:56:55.5738305+04:00,
launcher10524, original handle acquired. Same executable/hashes/arguments as plan.
Combined screenshot/accessibility capture preceded Enter-demo click and Nodes
click. Both actions completed; refreshed Nodes showed exact long-country group
and jp-1 label, ordinary other rows and Auto selection, with offline notice intact.
Thus the prior click-geometry error did not recur in these two clicks.

After refreshing Nodes, accessibility reported SearchBox focused. One type_text
call sent jp-1, but both the immediate and a subsequent screenshot/text refresh
still showed the empty placeholder and unfiltered list. No tool exception was
returned, but the required input effect was not established. No explicit click
inside SearchBox preceded typing. Do not infer native keyboard focus from the
accessibility focus report alone, or classify this as an application search bug.
The executor stopped without a second text attempt or later route actions.

Normal close requested at17:57:48.415Z (21:57:48.415+04:00). Original launcher
session29899 returned APP_EXIT PID4124 ExitCode0 at21:57:48.5992495+04:00,
elapsed53.0222065s, approximately0.184s after request. At21:57:56.2332530,
independent queries returned CimDemoCount=0 and NameDemoCount=0, shell exit0.
Refreshed UI inventory found no demo window. No force-kill or additional launch.

Classification: scoped close from initial Nodes passed with actual exit0; full
C3 route remains UNVERIFIED because search, selection/application, connection,
dialog Cancel and final reopened Nodes state were not completed. Neither the
original shutdown cause nor an input-tool cause is established. No source/build/
test/settings/debugger changes; all broader gates unchanged.

Next recommendation: diagnose input targeting from existing records before another
app reproduction. A future approved route should explicitly click the observed
SearchBox, verify actual editable focus, then enter text and verify its value;
do not launch again merely because accessibility claimed focus. Alternatively,
agree on user-assisted entry. This is a recommendation, not a new launch allowance.
