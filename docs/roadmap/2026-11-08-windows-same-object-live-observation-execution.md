# Same-object live observation execution

Plan: docs/superpowers/plans/2026-11-08-windows-same-object-live-observation.md.
Approved after4c5a323d. Preflight: clean expected worktree, AdministratorToken=False,
CIM/name demo counts0 and no demo window; app DLL/EXE hashes exactly match plan.
Module blob92ee0f0c23fd1521fbe103a781ab1ae2c9a58a7c matches68ec4a7c; import exposes
only the two planned functions. PowerShell7.6.9/runtime10.0.14. Budget0/1 before start.
Ruling: execute verification-only scope using session launcher and portable ledger;
no Bash helpers, implementation/full-suite run or new code review. Task1 identity
feeds Task2 receipts and Task3 classification; no app/source/settings changes.

## Runtime observations (UTC unless stated)

Budget1/1 consumed. APP_START PID5136 at2026-11-08T22:23:47.0837765+04:00,
launcher16772, original handle acquired. Initial module snapshots Mode=live:
identity fields ok; WaitSignaled=false and HasExited=false; ExitCode/ExitTime
each error/null with System.Management.Automation.MethodInvocationException.
These running-process errors were preserved, not converted to zero.

Short route completed within2minutes: Enter demo -> Nodes -> screenshot-backed
explicit field click -> visible caret -> one jp-1 input -> stable screenshot and
accessibility value jp-1; only long-country JP group and jp-1 row remained. No
selection/apply, simulated connection or dialog. Offline notice retained. Several
immediate captures lagged; fresh state was inspected before subsequent input.

Normal close requested2026-11-08T18:24:41.186Z. First subsequent snapshot at
18:24:42.7659373Z (ended42.7662795Z) preserved the same PID/start and reported:
- WaitSignaled: ok/false.
- HasExited: ok/true.
- ExitCode: ok/numeric0.
- ExitTime: ok/1601-01-01T00:00:00.0000000Z.

Status=ok denotes getter/type success only, not plausible exit time or acceptance.
The1601 value predates process creation and cannot establish actual exit time.
No outer wait signal arrived. Thus the disagreement is now observed on the same
retained Process object, not merely two objects from different execution contexts.
This still does not identify whether app teardown, OS state, runtime caching or
the execution environment caused it. The original-object HasExited may cache its
value; repeated True snapshots are not independent native exit measurements.

At22:24:50.5300127+04:00 external CIM/name counts both1; refreshed UI inventory
showed no demo window. At22:25:16.2508972+04:00 CIM listed5136/start22:23:47/
one thread and launcher16772. The30s close deadline failed. Continue only the
approved bounded read-only observation until360s; no further UI input or kill.

## Final result

Final snapshot18:29:47.1032795Z..18:29:47.1035462Z retained exactly the same
post-close disagreement, including numeric0 and1601 timestamp. Observer returned
OBSERVATION_TIMEOUT at22:29:47.1073414+04:00, elapsed360.015069s, Killed=False;
no OUTER_WAIT_SIGNAL event occurred. Shell exit0 means observer command completed,
not that application normal shutdown passed. At22:29:54.3651342+04:00 CIM found
neither demo nor launcher16772, and name-based demo count was0. No force-kill.

Classification: focused input/filter PASS; same-object evidence collection worked
in this one session; shutdown remains FAIL/UNRESOLVED under the30s criterion.
The live adapter has now been exercised for running-state getter errors and this
contradictory post-close state, but its signaled-exit path was not observed here.
The module's ok/type label does not validate chronology. Do not promote0 plus
HasExited to acceptance when the wait signal is absent and ExitTime is implausible.
No app/module/source/settings changes, builds or automated tests were performed.
Budget1/1 consumed; no unused runtime allowance. Raw receipts/screenshots remain
session-only; this portable summary preserves findings without private payloads.

Next: inspect the getter/native-error boundary and process teardown evidence,
starting with read-only source/reference analysis of the0/1601/unsignaled tuple.
Do not repeat the UI route without a new discriminator. A native wait/exit-code
probe, altered launcher environment or stack/dump capture would need a separately
reviewed scope/approval. Current evidence does not identify app vs runtime vs
OS/execution-environment root cause; all broader NOT_ACCEPTED gates remain.
