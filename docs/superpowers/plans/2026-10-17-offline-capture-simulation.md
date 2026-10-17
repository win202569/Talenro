# Offline Capture Simulation Implementation Plan

> **For agentic workers:** User approved plan/inline execution and4 hosts after88d623d3,then separate repair/diagnostic allowances. Latest106PASS/host verifies identified Important fix and wrapper regression;scoped delivery COMPLETE. All allowances spent/frozen,no real runtime authority. Historical states below retained.

**Goal:** Implement and verify a deterministic supplied-event capture model without real capture or runtime authority.
**Architecture:** A pure evaluator validates a bounded event array,accumulates owned channel bytes,and derives lifecycle status from virtual times,EOFs and exit. Reuse unchanged pure package/classifier helpers only;do not connect the file store or native diagnostics.
**Tech Stack:** Windows PowerShell5.1 and PowerShell7,.NET in-memory byte arrays/MemoryStream,existing pure evidence helpers,no new dependencies.
**Spec:** `docs/superpowers/specs/2026-10-17-offline-capture-simulation-design.md`,approved by user after3e9808c2.

## Global Constraints

- No component process start/handle/PID access,stream I/O,OS clock,sleep,thread/task,timer,file I/O,network,security query,nonce generation,or real probe identity. MemoryStream is only an in-memory byte accumulator,not external stream I/O.
- Test harness hosts are separate explicitly counted verification processes,not live collector/probe workloads;no process launched by the evaluator or fixture bodies.
- Events actual one-dimensional System.Object[],0..4096;plain exact PSCustomObject NoteProperties;no coercion/lazy enumeration/executable properties.
- Chunk0..65536 bytes,channel total<=1048576 bytes;strict Int32/Int64 fields,virtual time0..Int64.MaxValue,exit Int32 range. Equal times retain array order.
- Fixed120000ms virtual deadline;event at boundary timed-out before application. At most one exit and EOF per channel;exit may precede buffered chunks.
- First failure stops;retain exactly admitted prefix,no clipping/retry/late recovery. Trailing input after complete refuses;partial/failed result Package=null.
- Encoding utf-8/utf-16le;origin1..1024 characters/strictUTF8<=4096 bytes. Header rejection admits nothing.
- OriginVerified=false,Acceptance=false,RuntimeApproved=false;complete lifecycle is not successful output. No provenance/cleanup/durability/Stage B acceptance.
- Proposed fresh allowance4 synthetic-only test hosts:PS7 RED,PS5 RED,PS7 GREEN,PS5 GREEN,sequential,120s wall-clock outer limit each,no retry. Not authorized until plan approval explicitly includes it.
- All earlier budgets and unused frozen slots stay unchanged. No baseline/dependency installation/full-project replay,real security query or elevated fixture permission needed for this in-memory suite.

## File Structure

- Create `docs/diagnostics/isolated-fuzz/output-capture-simulation.ps1`:pure evaluator and private validation/byte-finalization helpers.
- Create `docs/diagnostics/isolated-fuzz/output-capture-simulation.tests.ps1`:literal synthetic event fixtures/assertions,including invalid types and bounds.
- Create `docs/roadmap/2026-10-17-offline-capture-simulation.md`:authority,results,hashes,review/rulings and limitations.
- Update current-status/repository-recovery at handoff;do not modify approved package/classifier/store/isolation/runtime files.
- Ignored audit runner/index may be prepared for a separate Simulation4 cap after approval;source/test imports must not depend on those ignored artifacts. Preserve receipts;no deletion.

## Review Focus

1. Caller executable properties or implicit conversions must be refused without invoking them;test ScriptProperty with a counter that must remain0.
2. Completion followed by malformed or late input must not hide trailing input;test final Package=null and preserved prior bytes/lifecycle.
3. EOF/exit state must survive timeout/refusal without synthesized success;test exact counters,index,time,nullable exit.
4. Multibyte codepoints split across chunks must be reassembled before strict decoding;test UTF8 and UTF16 splits with literal expected bytes/classification.
5. Boundary-sized inputs must reject whole offending chunks before copies/admission;test independent channel caps and4096/4097 event boundaries.

These are pinned below. Hostile same-process mutation,resource exhaustion,real pipe concurrency/OS deadlines and provenance are explicitly untested/non-goals,not guarantees.

## Task 1: Pure capture evaluator and bounded verification

**Interfaces:** consumes existing New-/Test-OfflineOutputEvidencePackage and New-OfflineEvidenceStream from the pure package helper. Produces `Get-OfflineCaptureSimulation([object]$Events,[object]$EncodingName,[object]$OriginDeclaration)` with the exact result fields/types in the spec. No entry-point side effects.

### Step1 — write synthetic tests before source

- [ ] Add guard `Check-Capture ($null -ne (Get-Command Get-OfflineCaptureSimulation -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineCaptureSimulationRequired'`,dot-source helper only if present.
- [ ] Use test-local plain-object constructors for chunk/eof/exit/tick;return exact literal Kind/Channel,integer time/exit and byte[] without casting invalid-fixture inputs. Use real evaluator,not execution mocks.
- [ ] Add literal checks of the following shape (test-local Check-Capture throws on false and prints one PASS):

```powershell
$events=[object[]]@(
    [pscustomobject]@{Kind='chunk';ElapsedMs=0;Channel='stdout';Bytes=[byte[]]@(97,98,99)}
    [pscustomobject]@{Kind='exit';ElapsedMs=1;ExitCode=0}
    [pscustomobject]@{Kind='eof';ElapsedMs=2;Channel='stdout'}
    [pscustomobject]@{Kind='eof';ElapsedMs=3;Channel='stderr'}
)
$r=Get-OfflineCaptureSimulation -Events $events -EncodingName 'utf-8' -OriginDeclaration 'synthetic-only'
Check-Capture ($r.Status -ceq 'complete' -and $r.AcceptedEventCount -eq 4 -and $null -eq $r.FailureEventIndex -and $r.LastObservedElapsedMs -eq 3) 'CompleteLifecycleFacts'
Check-Capture ($r.Stdout.Base64 -ceq 'YWJj' -and $r.Stdout.ByteLength -eq 3 -and $r.Stdout.SHA256 -ceq 'BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD') 'LiteralAbcSnapshot'
Check-Capture ($null -ne $r.Package -and (Test-OfflineOutputEvidencePackage $r.Package).Verified -and -not $r.OriginVerified -and -not $r.Acceptance -and -not $r.RuntimeApproved) 'PackageIntegrityNotAuthority'
```

- [ ] Add assertions per table;helper expectations must be literals or derived independently from fixture inputs,never evaluator output reused as expected:

| Case/fixture | Assertions |
|---|---|
|Empty object[]|incomplete,count0,indexnull,time0,both EOFfalse,ExitObservedfalse,ExitCodenull,empty snapshots,Packagenull|
|All6 permutations of exit/eof stdout/eof stderr at equal time3|complete,count3,time3,flags true;no artificial event reordering|
|exit at1,chunk abc at2,EOFs at3|complete,abc preserved;early exit does not forbid buffered data|
|Only exit / only both EOF / one missing EOF|incomplete;exact admitted flags/count,missing ExitCode remainsnull|
|Complete at119999 / final required event at120000 /120001|complete vs timed-out;late event not applied,correct failure index/last time,Packagenull on timeout|
|abc,tick120000,then later valid EOF/exit|timed-out,abc retained,count1,index1;no recovery from later input|
|Complete,then tick4 / malformed event / late tick120000|refused,trailing-after-complete,all prior facts/bytes retained,Packagenull;diagnostic time updated only for valid nondecreasing event|
|abc then backward timestamp / negative time / string,bool,double time|refused at index1,count1,abc retained,last time unchanged on invalid time|
|Duplicate EOF / duplicate exit / chunk after channel EOF|refused before applying event;exact index/count/old flags/exit preserved|
|Unknown Kind/channel,missing/extra/miscased fields,null event,hashtable,ScriptProperty|refused with no executing property getter;counter0 and no byte admission|
|Header null/scalar/hashtable/typed byte[]/multidimensional array/>4096 array|refused,indexnull,count0,empty snapshots;do not enumerate invalid outer arrays|
|Exactly4096 ticks at0|incomplete,count4096;4097 ticks header-refused,count0|
|Bad encoding/non-string encoding,empty/>1024 origin,unpaired surrogate/non-string origin|header-refused,no event applied|
|ExitCode -1/intMin/intMax / long outsideInt32 / string,bool,double|valid integers retained on complete even if classification refused;other values refused|
|Chunk empty /65536 /65537 /null/string/object[]|valid empty/max admitted;others refused whole chunk,no coerced bytes|
|16 chunks65536 per channel|each exactly1MiB admitted;next1-byte chunk refuses whole chunk and preserves length/hash/prefix,other channel independent|
|UTF8 bytes E4,B8,AD split across chunks|aggregate Base64 5Lit,length3;no per-chunk decoding|
|UTF16 bytes41,00,2D,4E split at odd boundaries|aggregate QQAtTg==,length4,explicit UTF16 retained|
|Progress CLIXML stderr split across chunks,UTF8 andUTF16|complete package progress-only/count1,exact input bytes retained|
|Invalid stderr00 FF0D0A /nonzero exit with valid lifecycle|complete simulation with refused classification;stderr AP8NCg==,no authority granted|
|Caller arrays mutated AFTER result|both result snapshots/package unchanged;no concurrency guarantee asserted|

- [ ] Assert every result's exact schema/policy/property set and conservative booleans;Stdout/Stderr metadata verifies admitted bytes. Assert failure/index/counters/time/EOF/nullable exit on each refusal class,not merely Status.

### Step2 — observe RED once per host

- [ ] After approval only,prepare separate ignored Simulation index with prelaunch consumption,budget4,only this suite,NoLogo/NoProfile/NonInteractive,120s per host,sequential,no automatic retry. PS5 child-only inbox module path. Save stdout/stderr/exit/elapsed;parent checks child record,not wrapper exit alone.
- [ ] Run PS7 RED thenPS5 RED. Expected both childexit1 at OfflineCaptureSimulationRequired,0PASS. This expected missing-API assertion alone permits proceeding to implementation;syntax/environment/unexpected failure stops. It does not prove each later assertion individually RED.

### Step3 — implement the approved pure interface

- [ ] Dot-source only output-evidence-package.ps1. Validate metadata/outer array before events. Inspect plain NoteProperty shape before accessing event data. Keep two bounded in-memory byte accumulators,EOF flags,nullable exit,count,index and virtual time.
- [ ] Implement the spec's precedence and result fields exactly;never use elapsed-time subtraction that could overflow Int64. Compare nondecreasing times and fixeddeadline directly. Reject an offending whole chunk before append.
- [ ] Define stable Reason codes for tests:complete,incomplete,invalid-input,invalid-event,time-went-backward,deadline-reached,chunk-after-eof,duplicate-eof,duplicate-exit,stream-limit,trailing-after-complete,package-integrity-failed. Shape/type/value violations use invalid-event;bad outer/metadata use invalid-input. Package refused classification is not package-integrity-failed.
- [ ] Trailing-after-complete wins even for malformed/late input;validate only that one event's shape/types/time to optionally update LastObservedElapsedMs,no lifecycle application or further events. Structural invalidity/backward time leaves last time unchanged. No caller executable properties accessed.
- [ ] Finalize owned immutable byte snapshots for all outcomes. Only final complete builds/verifies Package;on internal integrity failure retain admitted lifecycle/bytes,Statusrefused,Reasonpackage-integrity-failed,FailureEventIndexnull (not an event failure). Dispose in-memory resources;no files or OS resources acquired.

### Step4 — GREEN once per host,then one scoped review

- [ ] Run PS7 GREEN thenPS5 GREEN through the same finite index. Expected each childexit0,all fixture assertions PASS,empty stderr. Any unexpected failure stops before second host;no correction/retry under this allowance. Do not claim whole-project regression coverage.
- [ ] Dispatch ONE fresh readonly reviewer scoped to new helper/tests/spec and unchanged pure dependency contract. Review focus above;no reviewer test runs/edits or delegation. No second independent review loop. Record Critical/Important/Minor findings and declined scopes.
- [ ] If review finds a required fix,save clearly unfinished checkpoint and request separate bounded fix/RED-GREEN authority;the four-call cap has no hidden fix reserve. Minor findings recorded explicitly,not silently expanded into implementation.

### Step5 — preserve portable handoff

- [ ] Record observed host versions/exit/count/elapsed/stderr,index/source/test SHA256,expected vs unexpected failures and review limits in roadmap report. Test scripts should print host version from the running host,not spawn another query host. All raw receipts remain ignored.
- [ ] Explicitly stage new helper/tests/report plus current status/recovery. Commit/push ordinary checkpoint on existing dev branch,confirm remote/local SHA equality and clean tree. No main merge/force-push,install,security changes,Stage B or refreshed runtime facts.

## Plan self-review and execution handoff

One cohesive task;no inter-task interfaces or implementation delegation needed.
Spec input/result fields,strict bounds/time precedence and evidence semantics
covered. Reason codes refine the unconstrained diagnostic string,not behavior;
package-integrity failure has no event index. All five review focus classes
mapped to concrete fixtures. No newly identified spec conflict requiring
design revision. No code/test/host execution during plan preparation.

Recommendation:Native/inline implementation by current agent,one fresh scoped
review at the end. Alternative:subagent-driven implementation if explicitly
selected. Ask user to approve this written plan,select execution method,and
grant the proposed4-host synthetic allowance before beginning. Approval of
the earlier written spec alone does not grant execution or tests.

## Execution checkpoint (authoritative continuation state)

Approved inline/4 hosts after88d623d3. RED PS7exit1/1730ms andPS5exit1/2323ms
expected missingAPI;GREEN PS7exit0/2394ms/100PASS andPS5exit0/1936ms/100PASS,
both empty stderr. Initial stale parent LASTEXITCODE error preserved;readonly
saved-child check confirmed correct RED,no replay. PowerShell-native ledger/
runner honored cap;no extra task-done/full-suite/install/scratch deletion.

One fresh readonly review0Critical/1Important/0Minor:input GetType() ScriptMethod
dispatch before validation. Accepted static finding,not reproduced/fixed. No
post-GREEN source/test edits. Overall Task1 INCOMPLETE,4/4 spent;follow Step4
required-fix stop,not a hidden fix pass. Full report/rulings/recovery:
`docs/roadmap/2026-10-17-offline-capture-simulation.md`. Separate authority needed
for counter-only RED fixtures,non-dispatching type/shape fix,and4 new hosts.

Latest fix authority(after71b3358c):RED7/5 counter assertion failed as expected
(961/804ms). Reflection/local shape correction applied;PS7GREENexit1/981ms
after3PASS at normal tick assertion. No retry/PS5GREEN/second review;repair3/4
spent,unused fourth frozen. Candidate regressed,Task1 stillINCOMPLETE. Actual
wrapper/base mismatch unconfirmed. New cap3 diagnostic/correction/two suites
proposed in report,not authorized by earlier approvals. Old GREEN not transferable.

Latest cap3 approved after3d8fa50f:PS7diagnostic exit0/775ms confirmed reflection
PSObject/basePSCustomObject difference,counter0. BaseObject unwrap correction,
unchanged tests thenPS7exit0/2307ms/106PASS andPS5exit0/1793ms/106PASS,empty stderr.
Task1 scoped COMPLETE;Important/compatibility regression verified by previous
REDs/new GREEN,no second review or whole-project claim. No source/test edit
afterward;cap3/3 spent,old unused slot frozen. Prior rulings/exclusions preserved.
