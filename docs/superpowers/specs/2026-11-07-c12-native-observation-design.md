# Native verifier observation — proposed design

Status: DESIGN_APPROVED by user after1a145f52;not runtime approval.
First-stage plan:docs/superpowers/plans/2026-11-07-c12-native-observation-preflight.md.
User selected external observation of the unchanged native command;this document
proposes the concrete boundaries. Source checkpoint f248366b. PowerShell only,
existing codex/c12-b01-task9-coordinator worktree,inline author review.
Global44/34/6/217,NOT_ACCEPTED unchanged.

## Purpose and alternatives

Observe one future native verification invocation to identify visible work near
its deadline. Existing three-module samples are successful but cannot attribute
the previous300s native timeout. Do not claim a future trace recreates that run.

- Selected: Windows Performance Recorder (WPR) externally observes the pinned,
  unchanged Go executable. Preserves native code,but introduces measurement cost,
  system-wide metadata exposure and possibly incomplete symbol resolution.
- Rejected for this scope: an instrumented verifier copy offers explicit internal
  stages,but is another binary and cannot directly establish native behavior.
- Not useful for this question: more selected-module samples or a longer silent
  native timeout. Neither provides same-invocation attribution.

Success is defensible attribution to observed work,or an explicit inconclusive
result with preserved evidence. Native verification success and project acceptance
remain separate decisions. No repair,cache deletion/download,installation,security
setting changes,Go distribution modification,main merge or automatic retry.

## Local feasibility evidence (help queries only)

Tools1a8bbc/5c89f5/8360b1/a27c3a queried help,versions,hashes and profile details;
no start/stop/cancel,collector session or Go invocation occurred.

- WPR C:/WINDOWS/system32/wpr.exe,version10.0.26100.9711,
  SHA256 A0D0139D85577E51A5AFC3B4C4C99E00EBC53C5972917AA81A4BE133C7C1F8A8.
- WPA C:/Program Files (x86)/Windows Kits/10/Windows Performance Toolkit/wpa.exe,
  version11.7.395.48728,SHA256
  3CFAB105EE080AC3EDC1FA22183C793033E6B64D5D6839E97E037EC3FBF23BD8.
- Local help supports memory mode by default,profile details,custom profile files,
  and a unique instance name applied consistently to session commands. These
  capabilities are not proof that concurrent kernel collectors are isolated.
- CPU.light+FileIO.light reports1592+179 buffers of1024KiB (1771MiB,about1.73GiB)
  and numerous extra providers,including antimalware and networking-related ones.
  This is configured buffer capacity,not measured live allocation. Do not use
  this built-in combination or GeneralProfile as an automatic fallback.
- Previous local source inspection found mod verify does not register the
  build-only debug-runtime-trace flag;do not pass it or assume internal tracing.

## Two gates before any real observation

1. After design approval,write an implementation plan for a minimal custom memory
   profile,session ownership handling,limits and local evidence/export path.
   Validate the profile with non-recording tools first. Candidate system events:
   process/thread identity,image mapping,sampled CPU and file access/filename
   mapping;omit unrelated event providers and optional stacks unless their need,
   symbol usefulness and resource cost are justified. Exact schema/keywords and
   declared buffers must be verified,not guessed. Proposed buffer-capacity ceiling
   is128MiB;this is not a promise of total WPR/WPA memory or output-file size.
2. Before runtime,demonstrate from configuration/tool behavior how session ownership,
   bounded start/stop,output limits and analysis are enforced. Obtain explicit
   approval covering privileged collection,possible unrelated-process metadata,
   chosen resource limits and one finite native run. Design approval alone does
   not grant that authority. Failure of feasibility stops without recording.

No assumption that WPR can capture only the future Go PID. Filtering after capture
does not remove collection-time privacy exposure. If capture-time PID restriction
is unavailable,the runtime approval must explicitly accept that limitation.

## Proposed execution envelope (not yet authorized)

- One native `go mod verify`,maximum300s,not an increase over the prior native cap.
  Same pinned Go1.26.5,worktree/offline environment,GOMAXPROCS2 and four module-file
  baseline hashes. No simultaneous tests,builds,other diagnostics or helper hashing.
- Collector start/status allowance30s;native owned cleanup/drain uses existing
  bounded exit-confirmation rules;collector stop/save allowance60s. A stop command
  timing out does not prove the underlying collector stopped. No total-time promise
  or automatic new run if cleanup remains unconfirmed.
- Preserve Go's original execution privilege. Elevate only narrowly scoped collector
  control if required;never launch Go elevated to overcome execution restrictions.
- Before starting,check existing recordings/collector conflicts read-only. Any
  conflict or ambiguous ownership blocks startup;never stop someone else's session.
  New random instance name,fresh exclusive evidence directory,tool hashes,exact
  profile bytes/hash and launch claim precede collection. Instance name is necessary
  but insufficient:verify actual active session/profile identity before launching Go.
- Bind Go PID plus creation time,executable hash and owned launch receipt;avoid PID
  reuse mistakes. Preserve absolute timestamps and the trace timebase relationship.
- Collect from before owned launch until normal exit or deadline,then save only the
  owned recording. Memory mode may overwrite early events;record retained interval,
  event losses and gaps. Never infer absence of work from missing events.
- If collector startup/identity fails,do not launch Go. If collection fails mid-run,
  stop the owned Go process using existing safe ownership rules and preserve partial
  evidence;no silent unobserved continuation. Confirm Go exit and collector stop
  independently. Unconfirmed cleanup is a blocker requiring user action,not success.
- No generic cancel/stop,force stop,service changes or reboot. Any abandonment of
  an owned unsaved recording requires an explicit recovery decision because evidence
  can be lost;do not automatically discard it to satisfy a timeout.

Resource feasibility is a hard gate:the plan must distinguish ring-buffer capacity,
recorder overhead,save-time temporary space and final ETL size. A memory limit does
not cap save-time disk use. Preserve the existing4MiB JSON/stdout receipt limit;
ETL must have a separately approved local-only storage budget and enforced stop
policy,not be smuggled into or allowed to relax that limit. If safe bounds cannot
be established with installed tools,this approach is blocked,not silently broadened.

## Analysis and interpretation contract

Use installed WPA or a verified installed exporter offline;no symbol downloads,
new tools or server uploads. The implementation plan must identify a reproducible
export/read path before recording. Do not assume native Go stacks resolve to Go
function names in WPA. A supported file-event view may still be useful without them.

| Same-invocation evidence | Allowed conclusion | Not established |
| --- | --- | --- |
| Owned Go accesses module ZIP paths | ZIP-related file activity at those times | Hash completion,success,or exact function without stronger evidence |
| Owned Go accesses extracted module files | Directory-related file activity at those times | Full module coverage,why reads cost time,or integrity |
| CPU/threads active without attributable paths | Activity observed,phase unresolved | Deadlock absence or a particular verifier stage |
| Resolved stacks with trustworthy mapping | Sampled function/path at sample time | Continuous occupancy or whole-phase duration |
| No usable events/lost identity/missing interval | Inconclusive | Inactivity or a completed stage |

Graph loading,dispatch and error collection may remain indistinguishable from
external file/CPU events. Do not promise all internal stages can be reconstructed.
Record phase labels as inferred activity categories,not the helper's START/END
protocol. Record concurrent module activity separately;do not invent serial phases.
Completion requires native exit0 and its normal result plus existing module-file
checks;deadline remains NOT_PASSED even when the trace is informative. Trace validity,
native result and project acceptance must be reported in separate fields.

## Privacy,portability and handoff

Raw ETL may contain unrelated process names,paths,command lines and other system
metadata even with a reduced profile. Store only in an approved ignored local
directory;never commit,upload or inline/base64-package raw ETL. Do not promise it
contains no sensitive content or that a tool's PII option fully sanitizes it.

Publish reviewed derived evidence only:target identity binding,event counts,time
intervals,module-relative path categories,loss/coverage indicators and conclusions.
Exclude unrelated processes and user paths;record every omission and export rule.
Keep original evidence locally with length/SHA and explicit nonportable classification.
Another computer can restore code/profile/recipe/report,not the private trace or
machine state. Do not claim full raw-evidence recoverability from Git alone.

At handoff save approved specs/plans,necessary profile/controller/tests and reviewed
results to the existing development branch,with remote SHA equality and clean-tree
checks. This design-only checkpoint contains no implementation or recording.
Next step after document approval:write the bounded implementation/validation plan.
Inline author review:ownership,privacy,non-atomic cleanup,symbol limitations and
resource-bound feasibility are explicit gates;no independent reviewer claimed.
