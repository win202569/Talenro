# Token Full API Bounded Regression Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this plan inline. No independent reviewer or subagent launch is authorized by this proposal.

**Goal:** Check the isolated Token candidate against existing full API literal-result, refusal-index and result-ownership suites.
**Architecture:** Prepare two isolated routing directories with byte-identical existing tests and the corresponding unchanged module. Run baseline full,baseline ownership,candidate full,candidate ownership in four separate fresh hosts;any failure stops all later hosts.
**Tech Stack:** Windows PowerShell orchestration;ordinary pinned PS7 7.6.5 Core x64;.NET owned-process APIs;no installs.
**Spec:** docs/roadmap/2026-10-18-p1-token-direct-fixed-members-result.md and source/constraints in 2026-10-18-p1-token-local-preparation.md.
**Status:** PROPOSED,NOT EXECUTED. Aftera4921533,user requested continuing;this turn prepares coverage analysis/plan only. NEW max4 target attempts require approval below.

## Global constraints

- Four proposed launch attempts total,order baseline-full,baseline-ownership,candidate-full,candidate-ownership. All currently UNAPPROVED/UNUSED;all baseline gates precede any candidate launch.
- Each host:30000ms child wall including startup/load;2000ms owned exit wait;2000ms stream drain;maximum34000ms supervision. No timeout increase or aggregate host reuse.
- Any attempted launch consumes its slot,including startup/fixture errors. Failure,timeout,missing output,source drift or unknown cleanup freezes ALL later slots;no retry,repair,extra host,warmup,reload,PS5 or engine substitution.
- New slots are separate from all historical spent/frozen slots. The two successful direct slots stay2/2SPENT;older failed/frozen direct slots and candidate-full1/2/PS5 freeze unchanged.
- No product/candidate/test assertion changes,skipped case,smaller observation bound,cache,timer instrumentation,known-type flag,coercion,metadata interception or security-setting changes.
- Full API suite retains4096 acceptance and4097 rejection;ownership fixture retains its complete primitive/descriptor/identity checks. These four hosts do NOT include reflection/shape extra suites,benchmarking or whole-project tests.
- No reviewer/live service/Docker/native supervision/main merge/promotion authority. Whole acceptance/I1/P1/C12/security/runtime/release gates remain withheld even if all pass.
- New source preparation,static fidelity,controller preparation and four ordered attempts are included IF this complete proposal is approved;no additional preparation-only confirmation is required.

## Sources and future files

All paths below are repository-relative. Work in existing isolated codex/c12-b01-task9-coordinator worktree.

Immutable sources:
- docs/diagnostics/isolated-fuzz/integer-type-reuse-full-regression/offline-supervision-policy.tests.ps1
  SHA256 61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3
- Same directory offline-supervision-policy.ownership.tests.ps1
  SHA256 C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7
- Baseline:docs/diagnostics/isolated-fuzz/integer-candidate-validation-copy-attribution/offline-supervision-policy.ps1
  SHA256 3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3
- Candidate:docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1
  SHA256 F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A
- Engine:C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe
  SHA256 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139

Future portable create:
- docs/diagnostics/isolated-fuzz/integer-token-api-regression/.gitattributes (LF scoped to new tree),README.md,source-pins.json,expected-output.json.
- Same tree baseline/ and candidate/,each with offline-supervision-policy.ps1 and the TWO existing test filenames,all copied byte-identically from specified sources. No wrapper/import guard change;existing relative module imports route to exactly one module per host.
- docs/roadmap/2026-10-18-p1-token-full-api-result.md;update current-status.md and repository-recovery.md.

Future ignored create:.superpowers/sdd/2026-10-18-p1-token-full-api-run/run.ps1,frozen.json,invocations.json,progress.md and per-slot stdout/stderr files. Never overwrite any existing packet. No executable source/controller is created by this planning turn.

## Review focus and existing coverage

1. Full output fields/type categories differ although summary labels match:retain recursive P1-Equal/P1I1-Equal and literal P1-Want/P1I1-Want objects;not merely baseline/candidate string equality.
2. Late invalid observations bypass validation due to earlier cutoff:InvalidObservation and ObservationScalars compare literal FailureObservationIndex,including later malformed/decorated input.
3. Caller/result aliasing across evaluations:CallerMutation checks input mutation and distinct Decision references;SequentialIsolation checks distinct root results,metadata mutation and edited-result isolation.
4. Fixture decorations/getters are lost or executed:FixtureChecks and PassiveDescriptors must run;fixture errors/callback differences are failures,not skipped cases. Existing full API tests lack CodeProperty coverage;prior direct suite supplies separate tested-case evidence,not universal API coverage.
5. Large4096-row validation exhausts30s:retain ArrayAndIntegerBounds exactly;timeout is incomplete failure,no size reduction or automatic retry. Independent hosts separate full API from ownership,not divide/drop assertions within them.

### Task 1: Prepare and execute the existing suites unchanged

**Interfaces:** Consume exact source copies and per-suite -Group argument below;produce append-only-per-attempt receipts with SuitePassed and FormalAcceptance=false. No module/helper pre-import or mixed-module process.

- [ ] Step1: Copy the six module/test files into the new two routing directories;prepare LF rules/pins/README and independent expected output inventory as data BEFORE controller. Verify physical hashes/PS syntax and exact-copy AST/text. If sources cannot be copied faithfully or source pins differ,stop without launching. No imports,test calls or support compilation during static preparation.
- [ ] Step2: Construct literal ordered output inventory from source,not a learning run. Full suite38 nonempty lines:CONTEXT,ApiRequired PASS,HOST,34 further P1 PASS lines through Complete,then existing synthetic LIMIT. Ownership22 lines:CONTEXT,HOST,eight P1I1Fixture PASS,fixture LIMIT,ten P1I1 PASS through Complete,ownership LIMIT. Fix the ordered case labels from the source inventory;preserve both limits. Require current ordinary Windows identity CONTEXT with admin=False and literal 'HOST: 7.6.5 Core x64=True' framing. No dynamic timing markers. Partial pass counts alone never pass.
- [ ] Step3: Prepare and statically self-review controller:one owned Process.Start site,max4 order gate,durable attempted-slot receipt BEFORE launch,prior-failure freeze,no overwrite,concurrent original-byte capture,owned Process.Kill only and bounded exit/drain. Verify inventory mismatch/missing/extra/reordered lines,stderr/nonzero exit/timeout/unknown cleanup/drift all reject by inspection;no synthetic target hosts.
- [ ] Step4: Freeze all457 existing pins plus latest direct packet manifest/controller/index/both stdout/stderr anchors,new source copies/oracle/controller/engine/plan identities. Assert ordinary parent token;missing engine/records/elevated token/drift/unreadable receipts/unresolved safety check means no launch. Preserve original attempts and their consumption.
- [ ] Step5: Launch baseline-full ONCE:fresh no-profile noninteractive pinned PS7,-File <absolute baseline/offline-supervision-policy.tests.ps1> -Group all. Do not pre-import;script imports its one local module. Require exact38-line inventory,exit0,empty stderr,no timeout/controller error,complete streams,known owned exit,intact pre/post pins.
- [ ] Step6: Only after Step5 success,launch baseline-ownership ONCE:fresh same engine,-File <absolute baseline/offline-supervision-policy.ownership.tests.ps1> -Group regression. Its fixture checks precede its single module import. Do NOT use -Group all,which would rerun full tests and load the module again. Require exact22-line inventory and all Step5 receipt gates.
- [ ] Step7: Only after BOTH baseline successes,repeat Step5 ONCE targeting candidate directory,then after success Step6 ONCE targeting candidate directory. Independent literal object/index/ownership assertions must pass for each;agreement with baseline alone is insufficient. Any earlier failure freezes remaining hosts.
- [ ] Step8: Save original streams/receipt including UTC/PID/source/engine hashes,elapsed/CPU/exit/timeout/cleanup/drain/pin/oracle status. CPU/wall are receipts ONLY,not performance evidence. Re-read completed captures/index as data,never rerun. Record exact launch consumption,failure or all four results and source-copy fidelity in portable report/current-status/recovery.
- [ ] Step9: Explicitly stage necessary new portable sources/pins/README/report/docs only;verify format and Windows checkout-filter physical hashes;commit/push existing dev branch and verify remote SHA/clean tree. No ignored raw/controller/cache,secrets,main merge or scope promotion. Keep failed or successful packet;missing raw on another computer never resets slots.

## Approval requested

Approve preparing the isolated exact-source copies/controller and executing up to FOUR new ordered ordinary PS7 hosts under these limits. This authorizes no runtime source repair,extra test/helper/compile probe,retry,warmup,PS5,performance run,independent review or main integration. If any attempt fails,stop and retain evidence;later actions need new authority.
