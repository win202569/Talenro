# Isolated writer implementation — preparation stopped

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
