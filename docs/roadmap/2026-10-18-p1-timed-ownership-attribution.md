# P1 timed ownership attribution — timeout, observer cost measured

Date:2026-10-18. Base:935d4215efa63dfe465a8ec8b7c0918c18d7e55f.
User approved the recommended ONE new30s candidate stage-attribution diagnostic.
No candidate/test/logger/checker/assertion changes were authorized or made.
No retry,repair,warmup,extra probe/reviewer or later full/ownership suite ran.

## Receipt and validation

NEW stage-attribution1/1consumed. TimedOut=true,exit-1,elapsed30095ms.
Only the owned target host(PID44380) was terminated;owned exit confirmed,
both original streams drained,stderr0bytes,615pins intact,controller error null.
Outcome EXECUTION_INCOMPLETE,FormalAcceptance=false.

After target exit,the same ordinary PS7 controller invoked the existing strict
checker on SAVED TEXT only. No second host or product import by the checker.
CheckerError=null;TextOutcome=PREFIX_ONLY. Controller/inspector postprocessing
is outside the30s target timer. A valid prefix is NOT a full matrix pass.

Saved1038nonempty stdout lines:935TIMErecords+103original base lines.
Strict inventory validation confirmed467complete stage pairs and one open
announcement,plus the complete base-line prefix and37valid ownership facts.
All615pins and raw/analysis hashes rechecked intact at handoff.

## What completed

41base PASSlabels comprise two known-sharing controls,15direct scalar vectors,
initial normal control,and23/39behavioral cases(all21input/A/B rows plus
root-schema A/B). No control-after or terminal RESULT was emitted.
No ownership-gap label appears in this completed prefix;later cases remain
unverified. This does not close ownership acceptance.

Last saved event:sequence935,root-mode-A,prior-before,ENTER,pair935.
The preceding current-before interval completed. The open entry only proves
that historical-view scanning was ANNOUNCED;it does not prove its body began,
nor quantify work completed after that boundary.

At that boundary,the recorded attempted/completed counts are49/49decision calls,
29/29direct scalar-copy calls and1/1direct metadata call. These confirm49completed
decisions,not completion of the declared80call budget. Internal helper calls
inside decisions are not counted by the direct helper bounds. No final
interrupted-stage count is inferred.

## Measured completed-body intervals

Frequency10000000ticks/second. Sum of completed stage-body intervals:
28.5570738seconds. These include completed intervals within root-mode-A even
though that case has not earned a fact/PASS. This is an instrumented PREFIX,
not the complete matrix or an uninstrumented benchmark.

| Completed interval group | Seconds | Share of completed body |
| --- | ---: | ---: |
| Prior-view scans(before+after) | 19.6150681 | 68.69% |
| Literal oracles(before+after) | 2.6968846 | 9.44% |
| Candidate decision calls(49completed) | 2.1274615 | 7.45% |
| View capture/selection | 1.8978478 | 6.65% |
| Current-owner scans(before+after) | 1.7579091 | 6.16% |
| Other completed stages | 0.4619027 | 1.62% |

Prior-before:10.1826304seconds over37completed intervals.
Prior-after:9.4324377seconds over37completed intervals.
Root-mode-A's open prior-before interval is NOT included.

Last event elapsed29.2712532seconds from the CHILD clock origin.
Completed prior TIME emission cost recorded at that event:0.2882443seconds.
Unclassified remainder before that timestamp:0.4259351seconds.
28.5570738+0.2882443+0.4259351=29.2712532.

The current/final TIME emission's own duration is unreported. Parent launch/
parse time and unrecorded work after the final announcement cannot be inferred
by subtracting child timestamps from controller wall time. No such difference
is assigned to prior scans or candidate calls. Original FACT output lies inside
record-retain;other unclassified bookkeeping/output stays unclassified.

## Conclusion and limits

In THIS instrumented completed prefix,historical-view scans dominate measured
body time. Candidate decisions are a much smaller measured component. This
supports prioritizing observer implementation cost,not changing candidate
behavior or weakening ownership assertions. It does not identify the exact
PowerShell function/CLR call taking that time,prove a deadlock,or quantify the
older uninstrumented run's causal split. Instrumentation itself perturbs timing.

Recommended NEXT scope:design an observer-only implementation optimization,
with explicit equivalence/sensitivity checks preserving every current/prior-
unselected held/reread assertion and the30s safety policy. Do not clear history,
filter owner coverage,cache away required post-mutation checks,rewrap outputs
or raise the timeout to make it pass. No optimization or further run is
authorized by this report. Old full/ownership slots stay frozen.

## Immutable subject and new portable controller

Candidate1971F808,timing test986CDE76,strict checker7E140A1B and inventory7514F6C3
remain exact. All607prior entries plus four latest strict-checker anchors and
four new controller/pin/README/attribute files were frozen before launch.

| File in timed-ownership-attribution | SHA256 |
| --- | --- |
| .gitattributes | A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9 |
| capture.ps1 | E782F7B628F0E9F0C10790EAB16955E7361E6D760CBA9735E0F2F708FA1E4201 |
| README.md | C0948BB6B17BC92F56802DD6EDD57CF9E9EDE7CC2FD379F826E42CBE88F23CD4 |
| source-pins.json | 901BC7FBB2D5A5D0E5DF390F27D017928A8E039924DC4FCF2092CFC7488D6730 |

Controller parse/static source checks preceded the single fresh ordinary PS7
host. Old offline batches were not rerun. Preparation authorization is recorded
in [portable controller scope](../diagnostics/isolated-fuzz/timed-ownership-attribution/README.md).

## Preserved raw anchors

Ignored packet:.superpowers/sdd/2026-10-18-p1-timed-ownership-attribution/

| Artifact | SHA256 |
| --- | --- |
| frozen.json | E0622E5FE6994998E45E503E9DD51E05F52A5E8A3FA33A6FD4158DF3F6A55317 |
| invocation.json | 461456CF7E66A6A8D2F4BB4AA6232844897A43A5AC0C15E3541BFA0EB2D457FA |
| stdout.txt | EC63374C80CCB78765CCB78268919C02032DE0FE3B0EE0EEEA00852008A3A34D |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| analysis.json | A1A950609234B95F66364C7BD738FCF435B08B2B89FFCB656E3943AFCB94141C |

Original records remain ignored and unchanged;portable sources/this report/
current-status/recovery are saved on the existing development branch. A clone
restores progress,not ignored packets or a fresh execution allowance.
No main merge,I1/security/runtime/C12/release/PS5/performance acceptance.
