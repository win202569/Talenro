# Synthetic Capture Adapter Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended inline execution,or superpowers:subagent-driven-development only if selected by the user. Approved after30aa84f9 for inline execution and4 hosts;executed but review-blocked. Original planning checkboxes below are not acceptance evidence. See execution disposition.

## Execution disposition — 2026-10-17

### Approved Important-finding correction after75a6cabd

Disposition:call1 PS7RED exited1/2280ms;6 API regressions failed,then test helper
errored Shadow fixture not attached before aggregate assertion. ExpectedResult
false;STOPPED. Source unchanged,revised tests NOT GREEN. Repair1/4 used,remaining
3 frozen. Proposed fixture/bookkeeping correction and remaining3 continuation
requires new explicit approval;no call1 replay. See execution report latest.

User approved length-shadow correction and four NEW synthetic verification
hosts,not renewal of original allowance:PS7 RED,PS5 RED,PS7 GREEN,PS5 GREEN,
sequential120s each,no retry. Tests first exercise public constructor/verifier
with passive Length/Rank shadows on arguments,chunks,event arrays,metadata and
nested snapshots. RED must fail the named aggregate assertion after collecting
individual outcomes,not fixture errors. Then reject shadowing descriptors before
caller-visible intrinsic member reads;keep all original78 assertions. GREEN
must finish full updated contract suite with exit0/empty stderr. No unrelated
culture-token change;that candidate stays unverified/deferred. Stop on any
unexpected result;no second review or extra fix pass. Save separate repair index
under this plan's ignored workspace,retain original index/receipts unchanged.
Update report/status/recovery and commit/push necessary files at handoff.
Pre-flight:one cohesive correction,no shared task interfaces. No dependency or
live/runtime/security scope change. Original execution record below is historical.

Implementation/tests written. RED7/5 missing API as expected;GREEN7/5 each78PASS,
exit0,empty stderr. Approved cap4/4 spent,no source/test edits after GREEN.
One readonly review found Important passive ETS Length-shadow handling defect;
parent confirmed static path,not reproduced/fixed. Overall UNFINISHED,not task
acceptance. Report/status/recovery preserve this development checkpoint and
propose separately approved regression/fix/four-host verification. No extra run,
second review or allowance reuse. Full evidence and limitations:
[execution report](../../roadmap/2026-10-17-capture-adapter-contract.md).

**Goal:** Build a pure synthetic record binding declared command/time metadata to a bounded event trace and rederived lifecycle,without live collection.
**Architecture:** Validate/snapshot a synthetic packet,derive the existing simulator result,and hash canonical manifest/trace bytes. A separate verifier reconstructs from validated snapshots and rejects altered claims;no runtime action or persistence adapter.
**Tech Stack:** Windows PowerShell5.1 and PowerShell7,.NET bounded in-memory binary encoding/SHA256,unchanged corrected simulator and pure evidence dependencies.
**Spec:** `docs/superpowers/specs/2026-10-17-capture-adapter-contract-design.md`,written spec approved after9fbf8bf4.

## Global Constraints

- Module imports only corrected simulator/pure dependencies;no isolation/store/native helper. Do not modify existing simulator,package,classifier,store or ordinary-ledger code/schemas.
- No component process/PID/handle,external stream,file,clock query,sleep,timer/thread/task,network,security query,nonce generation or real adapter action. MemoryStream only for local primitive encoding.
- SourceMode exactly synthetic;no flag enables live mode. Metadata declarations/hashes are not origin authentication. LiveActionsPerformed,OriginVerified,Acceptance,RuntimeApproved allfalse;CleanupStatus alwaysunknown.
- Packet has exact10 fields from spec;valid record exact20 fields;refusal exact7 fields;verifier exact5 fields. All containers plain exact NoteProperties with strict uncoerced types and non-dispatching validation.
- Events actual Object[]<=4096;chunk byte[]<=65536;sum ALL packet chunks<=2097152,including rejected/late data. Simulator still separately caps admitted streams at1MiB each.
- ArgumentBytes<=65536,even,strictUTF16LE,no leadingBOM. SessionLabel ASCII1..80. Executable/origin strings1..1024 chars/strictUTF8<=4096 bytes;output encoding utf-8/utf-16le.
- UTC strings exactly invariant yyyy-MM-ddTHH:mm:ss.fffffffZ. Reversal retained,UtcOrderConsistent=false and verifierfalse. UTC never controls virtual120000ms deadline.
- Invalid packet preflight creates refusal descriptor,no record/hash/admitted trace. Valid packets can yield records with incomplete/timed-out/refused lifecycle. Verify consistency,not successful output.
- No automatic persistence/schema migration;Package-only storage omits manifest/trace/stop evidence. No main merge,install,security changes,Stage B or wider acceptance.
- Proposed fresh4 synthetic-only hosts:PS7 RED,PS5 RED,PS7 GREEN,PS5 GREEN,sequential120s each,no retry/extra fix reserve. Test harness hosts are counted verification processes,not collector workloads. All old allowances/frozen slots unchanged. No full-project/baseline or old-suite replay.

## File Structure

- Create `docs/diagnostics/isolated-fuzz/capture-adapter-contract.ps1`:New-/Test-SyntheticCaptureRecord plus focused private shape/snapshot/canonical encoding helpers.
- Create `docs/diagnostics/isolated-fuzz/capture-adapter-contract.tests.ps1`:synthetic literal fixtures,byte-level expectations,invalid/tampered/callback cases.
- Create `docs/roadmap/2026-10-17-capture-adapter-contract.md`:authority/results/hashes/review/rulings and limitations;update status/recovery at handoff.
- After approval only,prepare this plan's ignored audit ledger/runner/index with separate cap4. No source/test dependency on ignored files,raw fixture persistence or another machine's cache.

## Review Focus

1. Nested untrusted Lifecycle/Arguments/Events objects must not reach dependency helpers that dispatch caller GetType;counter fixtures cover outer/nested ScriptMethod and ScriptProperty paths.
2. Different interleavings with identical aggregate bytes/lifecycle must bind different traces;hash cannot cover only output snapshots.
3. Preflight limits include bytes after a simulated failure;failed lifecycle must not bypass total packet cap or cause partial constructor success for invalid structure.
4. UTC reversal and equal UTC with monotonic timeout must not fabricate success/cleanup or change deadline behavior.
5. Verifier must reject truthful hashes paired with false derived flags/status/stop/cleanup,not merely detect a changed hash.

Each is exercised below. Real byte acquisition,clock/command provenance,concurrent mutation,resource exhaustion,hostile process-wide tampering and durability are excluded/unmeasured,not certified.

## Task1: Constructor,verifier and bounded synthetic verification

**Interfaces:** `New-SyntheticCaptureRecord([object]$Packet)` returns exact record or refusal descriptor;`Test-SyntheticCaptureRecord([object]$Record)` returns exact verifier result. Consume unchanged Get-OfflineCaptureSimulation and locally constructed pure package/stream helpers. No live or persistence interface.

### Step1 — write tests first

- [ ] Add guard for both APIs,with first assertion named SyntheticCaptureContractRequired;dot-source helper only if present. Print current PS version/edition/x64 from the same host. Do not spawn diagnostic hosts.
- [ ] Use literal minimal packet and real constructor/verifier;test-local Check-Contract throws/prints PASS. Core assertions:

```powershell
$packet=[pscustomobject]@{
    Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x'
    ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o'
    StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z'
    Events=[object[]]@()
}
$record=New-SyntheticCaptureRecord -Packet $packet
Check-Contract ((Test-SyntheticCaptureRecord $record).Verified -and $record.Lifecycle.Status -ceq 'incomplete' -and $record.Arguments.Base64 -ceq '' -and $record.CleanupStatus -ceq 'unknown' -and $record.StopIntent -ceq 'none') 'ConsistentIncompleteIsNotRuntimeSuccess'
Check-Contract (-not $record.LiveActionsPerformed -and -not $record.OriginVerified -and -not $record.Acceptance -and -not $record.RuntimeApproved) 'SyntheticRecordNeverGrantsAuthority'
```

- [ ] Independently compare BindingSHA256 against SHA256 of this hand-encoded fixture hex,not a second call to production encoder. Decode hex in test utility using Convert.ToByte pairs (PS5 compatible):

```text
544C5243415030310100000001000000
01000000730100000078010000006F
1C000000323032362D31302D31375430303A30303A30302E303030303030305A
1C000000323032362D31302D31375430303A30303A30302E303030303030305A
0000000000000000
```

- [ ] Add literal/assertion cases in this table;expectations independent of implementation helpers:

|Fixture / break caught|Required assertions|
|---|---|
|Valid complete trace abc stdout,raw00FF0D0A stderr,UTF16 argument bytes41002D4E|Created record;stdoutYWJj/stderrAP8NCg==/argumentsQQAtTg==;lifecycle complete but output classification refused;verifiertrue,flagsfalse|
|UTF8/UTF16 split progress stderr|Exact original bytes and progress-only/count1;no per-chunk decoding|
|Exit before buffered chunks,missing EOF/exit,timeout tick120000,duplicate EOF/backward time/trailing event|Existing simulator semantics retained,nullable exit/prefix/count/index;stop intent matches four-row spec table,cleanupunknown;valid structural records may verify even if lifecycle fails|
|Two traces differing only interleaving of stdout/stderr chunks with same per-stream output/count/time|Final outputs equal,BindingSHA256 different;trace mutation without updated binding refused|
|StartedUTC after EndedUTC|Record retained with bytes and UtcOrderConsistent=false;verifierfalse,no metadata reorder|
|Equal UTC bounds plus virtual120000 tick|Timed-out despite equalUTC;would-request-owned-stop;no actual actions or known-clean claim|
|UTC malformed,offset instead ofZ,missing7digits,non-string/impossible date|Refusal descriptor only;no fabricated record/hash|
|Schema/source mode/unknown/missing/miscased fields and non-plain containers|Strict refusal;live source rejected,Createdfalse;descriptor SourceMode is constant synthetic refusal context,not caller mode echo or successful relabeling|
|Arguments empty/exact65536/even strictUTF16,65537/odd/BOM/unpaired surrogate/wrong types|Valid exact bytes preserved;invalid packet rejected before record;literal expected UTF16 bytes remain unchanged|
|Strings and session exact boundaries/overflow,invalid encoding and numeric strings/bools|Strict noncoercing types;session ASCII regex and UTF8/surrogate checks|
|4096/4097 events,65536/65537 chunk,2MiB exact total and2MiB+1 including chunks after timeout|Accept exact structural cap;over cap refuses entire packet even if simulator would stop earlier;no clipping|
|One channel1MiB+1 within2MiB total|Record exists,simulator stream-limit refusal;admitted prefix differs from retained full trace and both verified correctly|
|Mutate packet byte arrays/events/arguments after construction|Record snapshots/hash/derived lifecycle unchanged;no concurrent mutation promise|
|Mutate record metadata,UTC,trace,arguments Base64/length/hash,binding;extra/missing fields|Verifierfalse;canonical Base64 padbits/whitespace/oversize/type checks before decode|
|Alter Lifecycle,Package,snapshots,StopIntent,CleanupStatus,UtcOrderConsistent or any authority flag while leaving genuine BindingSHA256 unchanged|Verifierfalse by recomputation/exact shape,not only hash test;known-clean and performed-action claims refused|
|Counter-only ScriptMethod GetType / ScriptProperty on packet,event,nested arguments and lifecycle|No caller executable member runs;bad shapes refuse safely,valid unused GetType method does not intercept type query;no forwarding raw untrusted containers to dependency shape helpers|
|Change version/policy/source or use a refusal descriptor as a record|Verifierfalse,no migration or mode activation|

- [ ] Assert exact field names/types/conservative flags on every return and canonical stream metadata. Verifierfalse must preserve false authority flags;constructor refusal has Created=false and no record/hash fields. Do not assert mock call counts as functionality.

### Step2 — verify RED in two hosts

- [ ] Prepare independent ignored cap4 runner/ledger with prelaunch consumption,strict host/phase order,120s each,NoLogo/NoProfile/NonInteractive,PS5 child-only inbox modules;preserve decoded output/child exit/timing. No old runner index edits.
- [ ] Run PS7 RED thenPS5 RED. Expected childexit1 at missing SyntheticCaptureContractRequired and0PASS. Parent must check recorded child result,not stale LASTEXITCODE. Unexpected result stops without retry.

### Step3 — implement only the approved pure contract

- [ ] Validate exact packet fields/mode/metadata and whole typed trace before clone/admission. Use corrected PSObject-base/direct CLR type strategy and descriptor-first checks;never call input.GetType or an input ScriptProperty/ScriptMethod. Metadata parsed with invariant strict UTC/UTF16/UTF8,not evaluated as commands.
- [ ] Snapshot arguments/raw event bytes to immutable Base64/length/hash and reconstruct fresh trusted typed events for simulator. Recheck bounds before Base64 allocation/decode. Constructor returns exact record even when clock-orderfalse or lifecycle fails;only invalid packet yields refusal descriptor.
- [ ] Encode BindingSHA256 using the exact TLRCAP01 primitive layout in spec;length/count and exit Int32,time Int64,kind/channel one-byte. Preserve event order and bothUTC declarations. Dispose in-memory resources;no files/threads/clocks/OS calls.
- [ ] Derive simulator Lifecycle and StopIntent per spec,cleanupunknown and all authority/actionsfalse. Do not invoke verifiertrue as a constructor requirement for UTC-reversed records;that would discard required failed-clock evidence.
- [ ] Implement verifier shape/bounds/version/source/snapshots/hash checks,decode only bounded canonical inputs,rederive lifecycle/stop/UTC flags and compare full typed fields recursively. Reconstruct locally trusted containers before using unchanged package helpers;do not send caller-supplied Lifecycle/Arguments directly to GetType-based dependency validators. Culture-dependent/generic JSON text is not a canonical comparer.
- [ ] Use stable bounded reason codes:constructor invalid-packet;verifier consistent-synthetic-only,invalid-record,binding-mismatch,clock-order-inconsistent,derived-state-mismatch. More specific internal diagnostics may be logged only as bounded static codes,not caller command dumps. No secrets/raw real data in tracked fixtures.

### Step4 — GREEN and one scoped review

- [ ] Run PS7 GREEN thenPS5 GREEN through the same cap4 index. Expected exit0,all assertions PASS,empty stderr. Stop before second host on any unexpected failure. No retries/extra baseline/full-project suite/old-suite replay.
- [ ] ONE fresh readonly review of new module/tests/spec and consumed simulator contract;review focus above. No reviewer execution/edits/subdelegation. Required findings mean unfinished checkpoint/new separately authorized fix allowance;no hidden RED/GREEN reserve or second review loop. Record declined scopes and any deferred minors explicitly.

### Step5 — preservation and recovery

- [ ] Save authority/counters,observed host versions,expected/actual failures,source/test/index hashes,review/rulings and excluded guarantees in roadmap report. Retain ignored raw receipts,no automatic deletion.
- [ ] Explicit stage new source/tests/report plus status/recovery;commit/push existing development branch and confirm remote/local SHA and clean tree. No main merge or new runtime/security authority. Fresh machine uses tracked source/tests,not old caches/ignored audit files;do not auto-renew test budgets.

## Self-review and approval handoff

One cohesive constructor/verifier task;no inter-task dependency conflict. The
three architectural roles are represented by one synthetic record contract,
not three independently deployed subsystems. Exact signature/shape/primitive
types,clock reversal,first-failure versus full bounded trace,and dependency
callback hazards are mapped to tests. Literal canonical bytes give an independent
hash oracle. No written-spec behavior change needed;refusal SourceMode context
and bounded reason strings refine unspecified diagnostic representation only.

Recommend Native/inline implementation with one scoped reviewer at the end,
because constructor/verifier share one trace schema and canonical encoder.
Subagent-driven is an alternative only if selected. Ask user to approve plan,
method and proposed4 synthetic host calls before source/test work. Spec approval
alone authorizes this planning document,not implementation or verification.
