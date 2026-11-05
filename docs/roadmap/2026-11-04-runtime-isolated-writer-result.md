# Isolated writer implementation — preparation stopped

## 2026-11-05 read-only pin inventory follow-up

No hash benchmark, runtime replay, compilation or failed-output access. Counters
remain43/32/6/217. Inventory uses computation43 input.json plus file metadata.

|Input map|Hash comparisons|Sum of current file lengths, bytes|
|---|---:|---:|
|SourceSHA256|31|712273|
|PreservedSHA256|2400|175910433|
|HistoricalEvidenceSHA256|265|50346034|
|HelperSHA256|7|44638|
|Total|2703|227013378|

2669unique literal paths contain225795692bytes.34extra comparisons account for
1217686bytes (about0.54percent of total). Deduplication alone does not remove most
of the byte workload; its timing effect is unmeasured and repeated checks are not
automatically redundant from an integrity/observation-window perspective.
No absent-path pins are present in this particular manifest.

Largest individual inputs include portable-history-contract-runtime-completion-final.json
(38061125bytes), pinned System.Management.Automation.dll (19646264bytes), and
portable-runtime-isolated-writer-scenarios-final.json (17147719bytes). Most compared
bytes belong to retained/historical maps, not the31active source pins. File lengths
are logical bytes, not physical disk reads; caching, command overhead and scheduling
are not measured here. No particular large file is proven slow.

Call path: writer finally -> Map-IW -> Assert-HTInputs -> four sequential map loops
-> Get-HTHash -> Get-FileHash SHA256 for every entry. The observed output dictionary
overwrites duplicate literal-path keys only AFTER each hash, so it does not avoid
the duplicate work. Map serialization/write/hash occurs only after the full scan.
Observation43's last marker pins.begin precedes that scan; no pins.end was captured.
The helper neither caches hashes nor partitions historical versus live inputs.

Read-only inventory reporting initially used ordered dictionaries with property
sorting/Measure-Object, yielding null unique bytes and invalid ordering. Those
derived fields were discarded and recomputed with PSCustomObject metadata rows;
corrected figures above are from tool chunkf10ef4. This did not modify evidence or
rerun workload. Map subtotals were accumulated explicitly and were unaffected.

Conclusion: the2second exit-success window includes a full2703-comparison audit
whose logical byte workload grows with retained history. This is a structural
budget risk, not proof of the exact cause of timeout42/43 or the original full-host
timeout. Next design discussion should examine integrity-check scope/scheduling
and its guarantees, not merely extend the deadline, remove checks or repeat runs.
Moving checks to another phase changes the observation window and requires explicit
design approval and equivalent integrity reasoning; no such change is made here.

## 2026-11-05 latest — phase observation43 failed with two valid markers

Plan919122d0 approved;12memory cases/63assertions and static RED/GREEN passed.
Draft preparation historical references were corrected before its FIRST invocation;
no prior evidence changed. First preparation2657pins/37archive files/10script parses
passed. Prepared checkpointdd773432b6c07b2014781cddf67c6da1b46f361e remote-equal
before session38489, chunks604b24/daabec exit1. No runtime/test/preparation retry.

Supervisor FAILED/DrainTimeout, elapsed16236ms, DrainStart14205ms,
ExitStart16220ms. Producer5448 exit0; writer4216 was terminated and exit-1 confirmed.
Both owned exits confirmed; no pending resources/unknown exits. Counters43/32/6/217.
READY/COMPLETE and relay48/0 real EOF were recorded, but writer normal exit absent;
therefore outputs remain unavailable/null. No failure data inspection or salvage.

Timing parser: PARTIAL,481bytes,2valid identity-bound records, no tail/truncation:

|Seq|Phase|ElapsedTicks|Frequency|
|---|---|---:|---:|
|1|task.returned|126054191|10000000|
|2|pins.begin|126815166|10000000|

Local gap76.0975ms includes instrumentation and scheduling. There is no paired
pins.end, so no full pin-check duration is available. Writer after-map/role-result
absent. This narrows observed progress to after task return and recorded pins.begin,
before receipt of pins.end; it does not prove a particular file or permanent hash
deadlock. Begin-record write/flush and scheduling are also inside that uncertainty.
Do not align writer elapsed ticks with supervisor clocks or infer cross-pipe order.

Verifier d2f63d ran once, exit1 TIMED_PREFIX_FAILED;2703pins/2669unique paths matched.
Read-only b27b77 confirmed5available maps and62events matching computation29; child
HOST_PREFIX_COMPLETED, two business groups passed, no reported errors/open intervals.
Transport success, business checks and timing completeness remain separate conditions.

Final40files/3520045bytes,2unavailable data files excluded, decoded/source-byte
verification passed; SHA25698278C9EBA3AE834220D81E4D880B51F5C4AA52610091B007B90C95E6CEB2E4C.
Author scope review only; no independent reviewer or broad suite per finite plan.
Original protocol/DLLs/checks/deadlines unchanged; five flags false, NOT_ACCEPTED.
Gate ended. Next recommend read-only input-set and Assert-HTInputs call-path analysis
before proposing revised instrumentation/check architecture or a new finite run.

## 2026-11-05 read-only follow-up — post-COMPLETE exit path

No new process, compilation, benchmark, replay or failed-output inspection.
Counters remain42/30/6/217. Original runtime files/evidence stay frozen.

Source trace: Writer.RunAsync closes both sinks, hashes closed files, emits
COMPLETE and returns0. The PowerShell writer role then performs Map-IW in finally
(Assert-HTInputs over the complete manifest, serialize/CreateNew/flush the map,
hash that map), writes role-result and exits. The supervisor requires normal
owned exits AND terminal control/diagnostic readers; receipt alone is insufficient.
Thus the post-receipt work shares the same2000ms drain/exit-success window.

Existing evidence comparison (not a new timing run):

|Observation|Recorded duration / inventory|
|---|---|
|Synthetic S01 producer-exit observation to supervision finish|785ms|
|Synthetic S02 producer-exit observation to supervision finish|786ms|
|Business42 producer input.before interval|953.2089ms|
|Business42 producer input.after interval|2579.3967ms|
|Business42 drain start to failure-exit start|2015ms|
|Synthetic pinned inventory|2558comparisons/2524unique paths/194661307unique file bytes|
|Business pinned inventory|2657comparisons/2623unique paths/219763790unique file bytes|

Inventory byte counts use file metadata only; they are not measured I/O, and
duplicate comparisons may read a file repeatedly. Failed data paths are not inputs.
Producer input.after uses the same Assert-HTInputs implementation as writer
Map-IW but runs in another process/time; its2579ms does NOT establish writer timing.
Pin growth (+99unique paths,+25102483unique bytes) is a plausible contributor,
not a demonstrated cause. Cache, scheduling and filesystem variation remain open.

Confirmed: no successful writer exit within budget, COMPLETE captured, absent
writer.after-map/role-result, both exits confirmed after termination. The missing
map cannot distinguish an unfinished hash scan, pre-file serialization, a pending
post-emission continuation, or a process scheduling delay. No timestamp marks
COMPLETE emission/return/Map-IW entry, so a precise stalled instruction cannot be
reconstructed. Do not present this as the original full-host timeout root cause.

Recommended next scope: design a new version with bounded phase timing around
the existing writer completion and after-map steps, then request a finite new
observation budget. Preserve checks and deadlines initially; do not simply remove
the writer map, accept COMPLETE alone, or extend the timeout on this evidence.
This is an investigation recommendation, not approval to instrument or rerun.

## 2026-11-05 latest — business prefix42 stopped at DrainTimeout

Preparation37785d3f was pushed/remote-equal before the single invocation.
Runtime session63686 exited1; supervisor elapsed17029ms, normal-drain start14999ms,
failure-exit start17014ms. Producer PID9668 exit0; writer PID27388 was terminated
and owned exit-1 confirmed. Both starts consumed; counters42/30/6/217, no retry,
compilation, deadline change or safety-check removal. No pending or unknown exits.

Child reports HOST_PREFIX_COMPLETED, ContractAPIAvailable/BaselineCore passed,
cutoff true, no errors, and62events/31intervals structurally match computation29.
Relay reports48/0bytes and real EOF; control contains matching READY and COMPLETE.
Nevertheless writer did not normally exit within the shared2000ms drain window.
Transport FAILED/DrainTimeout is authoritative; outputs remain unavailable/null.
No physical failure output was inspected, hashed or salvaged.

Acceptance verifier baca92 exited1 at its transport-success gate, correctly refusing
to accept the business result alone. Read-only review68bb61 checked2657pins,
2623unique paths and5available maps. Writer after-map and role-result are absent.
Writer role runs full Map-IW in finally after Writer.RunAsync emits COMPLETE;
this is the next investigation target, not proof of the exact blocking operation.
Original full-host timeout remains unresolved; no integration/full acceptance claim.

Final archive37files/3429629bytes,2unavailable outputs excluded, decode/source checks
passed: SHA25693F36D09D25F18D4ED46D01B1B809822F5D5A7357AC4C67AC690FD4701927DAD.
Sources, scripts and original records now frozen. Author scope review only, under
the approved no-agent/no-broad-test exception; no fresh independent reviewer.
Separate business verification is authoritative, never result.json BusinessVerified=false
or the presence of COMPLETE alone. Any repair/new observation needs a new version
and separately approved finite budget. All five flags remain false; NOT_ACCEPTED.

## 2026-11-05 latest — finite synthetic scenario gate verified

Preparation checkpointb142444b84bec91c4d841aca02dd5244ca229443 was pushed and
remote-equal before one scenario-driver invocation. Session16311 exited0.
No retry, new compilation, C# edit, unit rerun or business host was performed.

|Case|Computation|Observed outcome|Elapsed ms|
|---|---:|---|---:|
|S01|33|SUCCESS,8193stdout/48stderr exact pattern,two exit0|3131|
|S02|34|SUCCESS,0/0real finalized empty output,two exit0|3024|
|S03|35|ReadyTimeout,producer never started,writer terminated/exit confirmed|5017|
|S04|36|DrainTimeout,held write,producer exit0,writer terminated|4046|
|S05|37|DrainTimeout,held finalization,producer exit0,writer terminated|4130|
|S06|38|WriterExited,writer exit23,producer exit0|2904|
|S07|39|ProducerFailed,producer exit23,writer terminated|1997|
|S08|40|CapacityExceeded,65537read/65536sent,both owned exits|2080|
|S09|41|MissingReceipt,both exit0,outputs unavailable|2890|

All9fixed assertions passed, including7intentional failure observations. These
seven are NOT successful output captures. All17started processes have confirmed
owned exits; no pending resources/unknown exits. Programmed deadlines were retained;
observed elapsed values are not OS wall-clock guarantees.

Independent read-only verification checked2558pins/2524unique paths, raw bounded
UTF8/LF control order/identities,18balanced journal records/1638bytes,31available
identity maps and5allowed absent writer after-maps (S03,S04,S05,S07,S08 termination).
Available maps each393451bytes/SHA256
2F74262F4DE4A3F12B6833915D89CABBF4C54D876FF68B7BEB999134FA64623A.
S01/S02 four closed success files total8241bytes matched independently generated
pattern/empty expectations and COMPLETE hashes/lengths. Fault-case file contents,
lengths and hashes were not read to manufacture evidence; fields remain null.

Final archive124files/17147719bytes,12unavailable data exclusions, decoded and
source-byte checked; SHA256
71B8CFA5EBFF2519FEA1EE29E9EF9777671E844D75120871BBCDB950DBF98FB5.
Both existing DLLs are referenced by pinned identity, not copied into the archive.
Original sources, compiler failures, claims and prior33-file compiled archive
remain unchanged. Current41/28/6/217; finite scenario task complete.

Together with the earlier18unit cases/81assertions, this verifies the selected
synthetic protocol/containment gate only. No claim of original timeout root cause,
exact original transport replay, business correctness, full-suite pass or publication
acceptance. All five capability flags remain false; overall NOT_ACCEPTED.
Stop here. A future bounded business-prefix observation requires its own approved
plan/budget, unchanged business guards/cutoff and original48/0-byte success checks.
No main merge, cleanup, fault-file salvage or additional process is authorized.

## 2026-11-05 remaining scenario gate prepared

User continued after4b935dae. New role/driver accept only writer/producer and
Scenarios respectively; no compiler entry or C# change. Existing v3 DLL reused.
S01..S09 are now computations33..41 with at most17starts and0compilations.
Current32/11/6/217, ceiling41/28/6/217. Original expected outcomes/deadlines unchanged.
Per-case handoff additionally records available writer-map identities and checks
successful S01/S02 stable files against fixed pattern/empty bytes before next row.
Failed files stay unavailable and are never read/hash-salvaged.

Preparation ran once; four scripts parsed. Independent2558hashes/2524unique pins
passed; prior33-file compiled evidence and both DLLs unchanged.
InputSHA256B8EA519BD519671C7E2E10FED2FF379441CC22E38BA7F6A649822FE3A2CED049.
19-file751268-byte preparation archive decoded/source-byte checked:
SHA256100487A3C7E850F32C283BC5F29F708EB9C98FC264DD87F33454B9399662CC57.
No scenario executed yet; checkpoint must be pushed before the one invocation.

## 2026-11-05 latest — v3 compiler and unit gate verified

Prepared checkpoint6dfbd38dcdbc36dcbff27d4da4491cb297bc439e was pushed and
remote-equal before the one driver-v3 Compile invocation. Computation32,
worker11,compilation6 succeeded. U01..U18 all passed,81assertions, no case errors.
Owned process19324 exited0 in4371ms; no kill or pending resources. Both captured
streams completed:32stdout bytes (ISOLATED_WRITER_UNITS_VERIFIED plus CRLF),stderr0.
Journal2balanced records/190strict-UTF8 LF bytes.

Independent verification checked2531hashes/2495unique pins and all four388920-byte
maps, SHA256A54FFA60581E8063921F87409A68C863433B7A653F4724A6AD2E9A36B4DB1FF8.
The80896-byte DLL matches the compiler receipt/reference identity:
SHA2562AEF2A39B5D499981378D8FD1F749A3073AD92C03641CEBBA28E9060B3FFCD7B.
It is retained locally and excluded from Git/portable archive.

Compiled archive33files/2817369bytes,1excluded binary, decoded/source-byte checked:
SHA256755D55B0EB738031266D4B58CAAA096A8916EE4881B7A28243B410994AE87FC6.
All old failures and original source/evidence retained. The only C# correction
remains the test-helper field-to-property declaration; no assertion relaxation.

Post-run recording caveat: after a new user turn, session-store paths/metadata were
unavailable. A read-only verification accidentally defaulted to main checkout and
failed before reading input; an incomplete controller record was also written
there. It was retained under compile/routing-incomplete.tool-record.json, not
discarded or presented as a full receipt. Explicit absolute-worktree read-only
verification then passed. compile/tool-record.json records the original successful
tool response's summary and distinguishes it from a byte-identical raw dump.
No compiler/unit/process replay or original evidence mutation occurred.

Active correction complete; counters32/11/6/217. Nine real-process scenarios remain
NOT_RUN. No business host, old-timeout repair claim, full-suite acceptance or main
merge. All five capability flags remain false; overall NOT_ACCEPTED.
Next is a separately scoped versioned scenario entry and renumbered33..41ledger,
reusing the verified DLL without recompilation. Compile-only v3 entry cannot launch
those scenarios. The next run's updated script/identity binding must be frozen
and pushed before any scenario starts.

## 2026-11-05 v3 correction prepared — compiler not invoked

User approved one additional compilation/worker after6c167190. New test source
changes only Pending from a public ReadOnlyMemory<byte> field to a property with
a private setter. No runtime/protocol/assertion changes. SavedCS1690 is the prior
failing evidence; no extra RED build or suppression. Active scope is computation32,
worker11,compilation6 with18unchanged unit cases; no scenario starts this turn.
Role-v3 and driver-v3 accept Compile only and use a fresh directory and claims.

One preparation invocation passed. Independent2531hashes/2495unique pins verified
all original inputs and the37-file stopped archive. Four scripts parsed; only the
approved test declaration differs.17-file725144-byte preparation archive decoded
and source-byte checked, SHA256
E64FB67C725C078EB68C6940B5B1CF9E2891C9515E418BF013EA934A7515A5E0.
InputSHA256E0886A836D7DB319EEC90BD6F31AE7661551C295E0C1126199891E2BE77D2E88.
Current31/10/5/217 unchanged; runtime not yet verified. Push before invoking once.
Original failures below are retained historical evidence.

## 2026-11-05 latest — compilation failed, gate stopped

After preparation checkpoint a6d4329196ba004eb5d0d9152f91f86d9a542d24 was pushed
and remote equality confirmed, driver-v2 Compile was invoked exactly once.
Computation31 launched one owned worker and attempted compilation5 once.
The compiler rejected tests.cs174:56 with CS1690: accessing a member on
HeldWriteStream.Pending (ReadOnlyMemory<byte>, a value-type field of the
MemoryStream-derived marshal-by-reference test helper) may cause a runtime exception.
The diagnostic and full caught error text are preserved in both compile receipts.

CompileSucceeded=false and Report=null:18unit cases did not execute. No scenario
gate claim or scenario directory was created. Worker exited1 after3306ms; owned
exit confirmed, no termination requested, no pending resources. Both captured
streams completed: stdout30bytes ISOLATED_WRITER_UNITS_FAILED plus CRLF,stderr0.
No source repair, retry, recompile, new unit run or process lookup followed.

Read-only independent verification checked2503hashes/2465unique pins and all four
384285-byte identity maps (same SHA256
0B44B04F203B0B6A2AF20BBF4BB5BAA5DAEE7D81E4D1ED5EE821E1A230C95363).
Journal:2balanced records,190strict-UTF8/LF bytes. Original11drafts/stop and prior
adapter evidence remain unchanged. A0byte failed DLL artifact exists locally;
it was not loaded or deleted and is explicitly excluded from the portable archive.

Final archive:37files,2773373bytes,1excluded binary; decoded/source bytes verified.
SHA2565F59F55C10BFF526BD2634B11F5A611621FEADC2845351A1A73ED2D1E6E1E59B.
Counters31/10/5/217. Task1 prepared; Task2 STOPPED; scenarios32..40 NOT_RUN.
There is no valid compiled checkpoint and no basis to claim protocol/runtime pass.
All five flags remain false and overall status remains NOT_ACCEPTED.

Proposed next action, NOT authorized or implemented: preserve original test source,
make a versioned correction to the test helper's struct-field access (a property
or an explicit local copy), and request one additional compilation/worker attempt
with updated computation numbering. Existing unused scenario budget cannot be
repurposed to retry compilation. No business host or main merge is included.

## 2026-11-05 approved v2 correction — preparation verified, runtime not invoked

User approved preserving originals, correcting dictionary-value summation in a
new prepare-v2 and continuing in the original budget. Driver-v2 authenticates
the original stop SHA and rejects separate new preparation/execution-v2 stops.
C# sources, tests and role entry are unchanged. Preserve-v2 writes new archives.

Preparation ran once and exited0. Independent read-only verification checked2503
hashes/2465unique pins,18unit names/9scenarios,17scenario+1compiler starts.
InputSHA256493FCEFCE38D71415497F7AE5D068483242F1DDBEB4ED6038CC60935EF28D6B3.
The20-file701088-byte v2 preparation archive was decoded and source-byte checked:
SHA2563F8C652EE92F2FC071EF96D0F314D6B1CBA745EB3EF94E602CDE586678F23B68.
Counters remain30/9/4/217. No new compiler, process, unit or scenario execution.
This checkpoint must be pushed before invoking the single compiler31/compile5.
No new design ruling: only approved count correction and authenticated versioned
entry/archive names. Original stopped preparation below remains historical evidence.

Status: PREPARATION_STOPPED / NOT_ACCEPTED. No compiler or synthetic child ran.
Approved plan: e45073471e25ea8d3b016b931d29776040ae5564, user reply 是.
Current counters remain30/9/4/217; approved ceiling40/27/5/217 was not consumed.

## Work authored, not verified

Five source drafts implement the proposed frame/receipt protocol, sibling-process
supervisor, relay, writer/producer roles and one-shot driver. Eighteen unit cases
were authored before implementation. They have not been compiled or executed;
none is claimed passing. Nine process scenarios have not been started.
No business-source changes, business host, full-suite run or merge occurred.

## Observed stop

The one data-only preparation invocation failed at prepare.ps1:45:
`Measure-Object -Property MaxStarts -Sum` rejected the MaxStarts argument.
The preceding source constructs the scenario rows as ordered dictionaries.
Read-only analysis points to using a dictionary key as an object-property
aggregation input. Proposed correction: explicitly sum the numeric dictionary
values (or deliberately construct typed rows), preserving the exact18-start cap.
This correction has NOT been applied or tested, and preparation was not replayed.

The stop happened after historical hash checks, four PowerShell syntax parses,
the single compilation-site check and the18-name check, but before writing the
static-checks report, approval binding or input manifest. Reaching that line is
not an independent behavioral validation of the source.

The original prepare/preserve scripts, all five drafts, actual initial tool
result and both polls, and preparation.stop.json are preserved. No input.json,
compile/scenarios directory, new DLL or gate claim exists. Original evidence and
old adapter remain unchanged. No cleanup or source repair followed the stop.

Independent read-only verification completed2470historical hash comparisons,
confirmed all previously absent paths and all seven uncreated new outputs, and
recorded frozen source hashes. The11-file archive is128030bytes, with no binary
or unavailable output included. Decode/length/hash/source-byte verification passed.
ArchiveSHA256:011ABB2C0503963BDB605A8316189DD5F1D3B260A5A94E5BBDEBC1360F06EA4A.
StopSHA256:894DF77E82EFEEC69F233FBF85A0E439F0590C502529D49A2DBE07D1A3F5185B.

## Recovery and next authorized decision

Read portable-runtime-isolated-writer-preparation.json under
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2. It is a STOPPED source/evidence
checkpoint, not a successful preparation receipt. Decode-check paths, lengths
and SHA256 before restoring ignored artifacts; never overwrite existing evidence.
The preceding adapter archives remain prerequisites. No runtime/DLL/cache binary
is committed; a fresh machine cannot silently rebuild the old adapter.

Task1 remains incomplete and Task2 has not begun. Original drafts and stop are
frozen. A versioned preparation correction and authenticated continuation entry
require renewed approval under the stop rule; unused runtime allowance is not
permission to retry automatically. Keep the same18unit/9scenario gate, no extra
compile, no broad tests and no business-host authority. Corrected preparation
must be archived and pushed before the sole compile; its compiled checkpoint
must be independently verified and pushed before scenarios.

## Scope decisions

The approved single-compile exception omits separately executed RED testing;
there is no executed-TDD claim. No reviewer agent/process was used.
The definition-only historical hash helper is reused only after its identity is
checked; the driver captures its hash-function definitions for script scope.
No historical operational entry or business module is invoked. If this scope
decision is wrong, identity/preparation must stop rather than expand execution.

All five capability flags remain false. This topology changes the original
transport path and cannot establish the old timeout's root cause or its repair.
