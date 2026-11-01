# Error-reference assertion — readonly analysis

Status: READONLY_ANALYSIS_COMPLETE; ROOT_MECHANISM_NOT_YET_PROVEN.
Baseline6f9ebae425bc18a8b0fb7b82e9a5f3d107bc9237. The user approved only
reading the saved failed result and related source and proposing a revision.
No code changes, imports, test/Prepare/diagnostic execution, compilation,
targets, agents or reviewer. Runtime remainder6/7/1/conditional1 stays FROZEN.

## Confirmed facts

The original evidence is writer-red.evidence.json in the retained integration
packet, also preserved in portable-continuation.json. The actual parent final
is Exact=false/EvidenceValid=false, not a successful intended RED.

| Saved completed case | Actual primary | Actual thrown | Secondary count / expected RED count |
| --- | --- | --- | --- |
| CaptureWriteDisposeFirstCause | FileNotFoundException | ObjectDisposedException | 1 / 1 |
| CaptureFlushDisposeFirstCause | FileNotFoundException | ObjectDisposedException | 1 / 1 |
| IntentWriteDisposeFirstCause | ObjectDisposedException | FileNotFoundException | 2 / 1 |
| IntentFlushDisposeFirstCause | ObjectDisposedException | FileNotFoundException | 2 / 1 |

In BOTH completed Intent cases, the two secondary projections are equal in
all five recorded fields: Type, Message, ErrorId, Position and Stack. Their
actual Open/Write/Flush/Dispose counts are1/1/0/1 and1/1/1/1 respectively.
Two error records therefore do not imply two Dispose calls. The fixture's
DisposeCalls tracks its operation; serialized equality is NOT reference identity.
The other three completed Capture cases pass and have actual secondary count0.

The seventh completed group is IntentFlushDisposeFirstCause. The next active
case is IntentDisposeOnly, identified by the last journal guard stage Finally.
The assertion at record-boundary.tests.ps1:53 requires exactly one entry with
the wanted FullyQualifiedErrorId in primary+secondaries. Its actual throw proves
the count was not1, but does not serialize whether it was0,2or another value.
No RecordCase observation for the aborted case exists. Never fill that gap
with a number inferred from earlier cases.

## Source-level causal path

1. Write-ODProfileIntentRecord in private/write-boundary.ps1:19–20 calls
   Capture-Failure inside catch/finally and then uses a bare rethrow.
2. The test's outer catch in record-boundary.tests.ps1:48 captures the arriving
   ErrorRecord and calls Capture-Failure again.
3. Capture-Failure in capture.ps1:21–26 deduplicates only ReferenceEquals against
   the current primary/secondaries. It does not compare fields or exception text.
4. The test gathers primary+secondaries, selects every matching ErrorId and
   throws before recording the case if its count differs from1.

These sources and the two saved duplicate projections establish a mismatch
between the frozen RED expectation and actual collection behavior. Under the
unchanged collector path, both entries survived reference-based deduplication;
this supports the inference that distinct ErrorRecord references reached it.
It does not establish whether their underlying Exception references were the
same, nor which engine boundary created a new wrapper. Those identities were
not serialized and cannot be reconstructed from JSON.

CaptureDisposeOnly is a useful existing contrast: its writer retains a local
primary and later uses `throw $primary`, rather than the Intent path's immediate
collection plus bare rethrow. That completed case passes with no secondary.
The differing collection/rethrow routes support a focused hypothesis; they are
not a controlled proof that changing just one throw form fixes the whole chain.

Known I1 first-cause ordering defects are still visible in the table. This
analysis does not dismiss them as a harmless harness problem or close I1/I2.
The partial run still lacks three completed writer groups and all three final
recorder groups. The25saved stages show no outer function-reference escape,
but do not constitute full publisher/private-state acceptance.

## Evidence limits and rejected shortcuts

Do not call two equal serialized errors the same object. Do not fix this by
deduplicating only ErrorId/message: two independent failures can share text,
and original-reference preservation is an explicit requirement. Do not drop
the count assertion, accept arbitrary extra secondaries or retrofit the frozen
oracle to make this failed run pass. Do not assume the production defect is
fixed merely because the Capture-only controls succeeded.

The historical failed independent integrity command is still NOT_COMPLETED;
this analysis is not a replacement full-pin audit. Reading the original writer,
test and fixture yielded their unchanged expected SHA256 identities:

- Writer: F3AC2F3ED96B3C99EA2C137478D15DA5FA1BDE940BBCD5D426683926A081812C.
- Writer tests:5516D9D2CE597364C4FB59053E3216D9C8DEC10E80DC71D2E89FBECC1BE9B8CF.
- Fixture:1626D2803E32F03FCD6B60F3F03C13E0A876F4C2439A5CCA23A087F65761B9C5.

## Recommended next decision — not execution authority

Stop adding runtime generations or launch controllers. Before any integration
retry, separately design a tiny error-reference propagation diagnostic, using
real existing PowerShell/.NET I/O errors and explicit inner/outer observation
points. It should distinguish bare rethrow from throwing a retained record,
compare ErrorRecord AND Exception references, preserve event order and every
original object until the comparison completes, and report which collection
boundary adds the extra entry. No targets or fixture compilation is needed
for that proposed narrow question; exact host allowance and implementation
must still be written and explicitly approved before any experiment.

Only after that mechanism is established should a separately approved change
select a single collection owner or retain/rethrow a verified original record.
Preserve actual first-vs-cleanup order and distinct-error reference semantics,
then define a new legitimate source/contract transition; never rewrite the
existing frozen tests/manifests or re-run consumed writer-red implicitly.
The actual old failures/archives remain unchanged. This recommendation releases
no budget and does not authorize a new source, diagnostic or runtime attempt.
