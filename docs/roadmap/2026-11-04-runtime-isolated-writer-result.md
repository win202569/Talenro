# Isolated writer implementation — preparation stopped

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
