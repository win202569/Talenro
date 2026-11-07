# Native Observation Preflight Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline task-by-task with the established author-review boundary. Await plan approval before implementing. Steps use checkbox syntax;no subagent or independent-review claim.

**Goal:** Produce a validated minimal WPR profile and an evidence-backed feasibility
decision without starting any recording or native verification.

**Architecture:** Separate pure profile checks and decision logic from installed-tool
help/profile inspection. A fail-closed report records which runtime prerequisites
remain unproved. No real collector controller or Go launch path is built in this stage.

**Tech Stack:** Existing PowerShell7/.NET XML,installed WPR and WPT schema/help;
no new packages. Existing immutable receipt helpers remain unchanged.

**Spec:** docs/superpowers/specs/2026-11-07-c12-native-observation-design.md,
approved after1a145f52. This plan implements its non-recording feasibility gate;
live ownership/control/analysis execution requires a later separately approved plan.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator;
  branch codex/c12-b01-task9-coordinator;PowerShell only;44/34/6/217,NOT_ACCEPTED.
- Source directory(S):docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/native-observation.
  New ignored workspace(W):.superpowers/sdd/2026-11-07-c12-native-observation-preflight.
  Each invocation gets a fresh exclusive child root. Preserve old evidence.
- No WPR start/stop/cancel/merge,ETW session creation,Go invocation,cache inventory,
  exporter processing of ETL,GUI launch,install,privileged recording or security change.
  No imports that launch processes. No generic command execution interface.
- WPR pinned C:/WINDOWS/system32/wpr.exe SHA256
  A0D0139D85577E51A5AFC3B4C4C99E00EBC53C5972917AA81A4BE133C7C1F8A8.
  Use this executable,not the distinct WPT-directory wpr.exe.
- Installed reference directory(R):C:/Program Files (x86)/Windows Kits/10/Windows Performance Toolkit.
  R/WPRControlProfiles.xsd and R/SampleWPRControlProfiles.wprp are readable local
  references found2e658d/1e09a8. Record hashes;do not copy full vendor resources into Git.
- Custom candidate:one SystemCollector,one SystemProvider,one Memory profile;
  BufferSize1024KiB,Buffers128,fixed count,declared128MiB capacity. Only keywords
  ProcessThread,Loader,SampledProfile,FileIO,FileIOInit. No stacks,EventProvider,
  EventCollector,inheritance,per-CPU scaling or extra capture-state providers.
  Schema names verified in installed XSD by1e09a8;runtime efficacy is unproved.
- Existing4MiB receipt-root and64KiB individual JSON limits remain. No ETL exists
  in this stage. Do not confuse declared buffer capacity with total memory/disk cap.
- At most one RED and one GREEN batch per Task1/Task2,each60s;no child fixtures or
  real collector calls. Task3 permits at most eight read-only tool calls,each30s,
  through existing bounded owner. No automatic retry on unexpected failure.
  Allowed commands only:WPR help/profiledetails/profiles;installed exporter help.
  Do not query global active-session metadata in this stage or request elevation
  merely to make a help/profile check succeed.
- Existing owner/common tools may supervise allowed help/test processes with empty
  Session/raw mode;record current source hashes and actual executable identity.
  Do not modify owner/common,old profiles/recipes,Go code or any previous archive.

## File map and Review Focus

Create S/profile.ps1 (definition-only XML policy),S/native-observation.wprp,
S/profile.tests.ps1,S/readiness.ps1 (pure evidence decision),S/readiness.tests.ps1.
W holds bounded invocation receipts and packaging recipe;committed portable evidence
excludes binaries/vendor resources/private traces. Report is
docs/roadmap/2026-11-07-c12-native-observation-preflight-result.md.

1. Schema-valid but overbroad profile:reject extra providers/keywords,inheritance
   and count scaling (Task1 tests).
2. Profile validation mistaken for live readiness:never return runtime permission
   from static or synthetic evidence (Task2 tests).
3. Named session mistaken for exclusive ownership:require independent ownership
   evidence and separate Go/collector exit confirmation (Task2 synthetic tests).
4. Ring size mistaken for disk bound:separate capacity/storage enforcement and
   reject missing enforcement evidence (Task2 tests).
5. Missing/circular/unsafe XML or missing event data mistaken for success:reject
   DTD/external resolution/unknown input;mark observation gaps inconclusive (Tasks1–2).

## Task1 — Minimal profile and fail-closed validation

**Interfaces:** `Test-C12ObservationProfile([string]$Xml,[string]$SchemaPath)` returns
hashtable `{Valid:bool,DeclaredBufferBytes:long,Reasons:string[]}`. Missing schema
or invalid XML yields Valid=false,not an exception-based success. XML reader must
prohibit DTD and external resolution. XML schema validation plus explicit policy
checks;the schema being valid alone is insufficient. Reason codes stable per case.

- [ ] Read relevant installed XSD types and sample profile completely enough to
  resolve root/Profile/collector/provider references;record source hash and sections.
  Create W and initial ledger,check clean branch,existing source pins/module hashes.
  Expected:no process or recording created during setup.
- [ ] Write tests first with in-memory XML fixtures:exact candidate accepted at
  134217728 bytes;129 buffers rejected;File logging rejected;extra keyword/provider
  rejected;Stacks rejected;Base inheritance and per-CPU/percentage counts rejected;
  duplicate/unresolved IDs rejected;malformed XML/DTD/external entity rejected;
  missing schema rejected. Include zero/negative/overflow size and extra profile.
  Tests use no cache paths or process launch and do not expand external entities.
- [ ] Run `profile.tests.ps1 -Phase Red` once in a fresh supervised60s batch.
  Expected:nonzero from absent policy behavior/failed assertions,not timeout or
  schema/tool infrastructure failure. Preserve fixture and receipt;unexpected failure stops.
- [ ] Implement definition-only validator and candidate XML. Name profile
  C12NativeObservation.Light.Memory,Name C12NativeObservation,DetailLevel Light,
  LoggingMode Memory;collector/provider IDs C12SystemCollector/C12SystemProvider.
  No automatic fallback to built-in profiles;schema refusal is a finding.
- [ ] Run `profile.tests.ps1 -Phase Green` once,60s.
  Expected:all named assertions pass,confirmed exit0,no timeout,unchanged module files.
  Commit source/tests and reviewed synthetic evidence;do not claim live support.

## Task2 — Evidence readiness and classification contract

**Interfaces:** `Get-C12ObservationReadiness([hashtable]$Evidence)` returns
`{Status:string,Blockers:string[],RuntimeAuthorized:bool}`. Status is
BLOCKED or READY_FOR_RUNTIME_PLAN only. RuntimeAuthorized is always false.
Evidence contains named records with Status (Verified/Unverified/Failed),Source,
and Kind (Static/Synthetic/Observed);missing/unknown fields fail closed.
Required keys:Profile,Ownership,StartStopBounds,StorageEnforcement,OfflineAnalysis,
PrivacyAcknowledgement,PrivilegeSeparation. Synthetic records cannot satisfy real
ownership/storage/analysis requirements. No execution function is exported.

- [ ] Write RED tests for all keys verified versus each missing/failed/unknown;
  SyntheticOnly evidence blocked;InstanceNameOnly ownership blocked;128MiB buffer
  alone cannot satisfy StorageEnforcement;help-only cannot prove working ETL analysis;
  missing privacy approval blocked;collector timeout plus exited CLI cannot prove
  stopped collector;Go exit and collector stop are independent;PID without creation
  time insufficient;lost/unmapped events imply inconclusive attribution.
  The latter conditions are represented as invalid evidence records with explicit
  reason strings,not invented real observations. Use table-driven fixture objects.
- [ ] Run `readiness.tests.ps1 -Phase Red` once,60s;expected assertion failure only.
- [ ] Implement pure decision logic:all required keys and required evidence sources
  must be present and verified;otherwise BLOCKED with all reasons. Even complete
  fixture evidence only tests the decision contract,not real readiness.
- [ ] Run `readiness.tests.ps1 -Phase Green` once,60s;expect full PASS and confirmed
  exit0 without timeout. No real collector,Go or exporter launch from tests.
  Commit source/tests and immutable results. A PASS is not runtime approval.

## Task3 — Installed-tool feasibility and honest handoff

**Consumes:**Task1 candidate/validation result and Task2 readiness contract.
**Produces:**readiness.json in W/Report plus portable reviewed evidence and report.
No live controller/export parser is built until its actual input/ownership contract
is understood;do not claim this plan implements runtime lifecycle handling.

- [ ] Recheck WPR/XSD hashes and candidate profile validation. Record candidate bytes
  and SHA. Run the following read-only calls once each with30s caps and raw receipts:
  `wpr.exe -profiles <absolute-S/native-observation.wprp>`;
  `wpr.exe -profiledetails <absolute-S/native-observation.wprp>!C12NativeObservation.Light`.
  Expected:exit0,exact single Memory profile/collector,128 buffers of1024KiB and
  exact five keywords,no extra providers. Unexpected expansion/locale ambiguity
  blocks readiness;do not parse partial text as proof or silently change limits.
- [ ] Use up to six additional non-recording help calls to establish named-session
  control semantics and installed export capabilities. Candidate exporter
  R/wpaexporter.exe exists (1e09a8);record its hash/version before querying help.
  Consult local help/schema or official vendor documentation for missing facts;
  do not invent switches or run exporter against traces. Any external documentation
  must be cited with supporting section. No downloads/installations.
- [ ] Write feasibility table with exact evidence and remaining gaps:existing
  collector conflict detection,unique-instance ownership binding,status ambiguity,
  failure during start/save,Go privilege separation,ring overwrite/event loss,
  save-time temporary storage/final ETL bound,offline file-event export identity
  and timestamp fields,raw-trace exclusion. State which facts cannot be demonstrated
  without an explicitly approved synthetic recording;do not perform that recording.
- [ ] Evaluate actual evidence with Get-C12ObservationReadiness. Expected:BLOCKED
  whenever ownership,save storage enforcement,analysis or privacy authority remains
  unproved;READY_FOR_RUNTIME_PLAN only when every requirement is evidenced. Never
  label untested configuration as runtime-ready. Report module hashes unchanged.
- [ ] If a gate is unmet,hand off its exact evidence,consequence and smallest proposed
  next scope. Do not compensate with broader providers,longer runtime,ETL quota
  relaxation,privilege elevation or implicit permission to discard recordings.
- [ ] If gates are evidenced,prepare a separate proposed live plan covering the
  design's300s native/30s start-status/60s save limits,owned cleanup and per-stage
  permission. Its exact commands,ETL storage budget and tests need their own review;
  this step is a written handoff requirement,not authority to implement or execute it.
- [ ] Package new source/test fixtures,tool help receipts,hashes and feasibility
  report after privacy review;decode every included entry and check length/SHA.
  Exclude vendor schema contents,binaries,private traces,caches and credentials.
  Update current-status/repository-recovery and this plan outcome. Review explicit
  file list,empty index,diff check;commit/push existing development branch and verify
  remote==HEAD plus clean worktree. Retain original local evidence.

## Self-review and approval boundary

Task1 profile contract feeds Task3;Task2 prevents a configuration-only success from
becoming runtime authorization. All five review risks have synthetic checks and
Task3 reality checks. No test can prove actual capture/cleanup/storage behavior
without observation;unproved requirements remain blockers,not waived gates.
Spec runtime lifecycle,ETL handling and analysis are deliberately deferred behind
its separate approval gate rather than falsely covered by this preflight plan.
Plan status:READY_FOR_REVIEW. Approval would authorize only the finite offline
implementation/tests/help inspections above,not real recording or native verify.
