# Capture adapter contract and synthetic evidence record — proposed design

Status: **WRITTEN DESIGN APPROVED FOR PLANNING** after9fbf8bf4. Original writing
authority followeda23115d5. Plan approval,execution method and fresh finite
verification allowance still required before implementation/tests. No host
diagnostic,real collector,watchdog or runtime allowance granted. All prior
spent/frozen budgets and failure history remain unchanged.

## Intent and staged delivery

Define the boundary needed to associate separate raw channel bytes with a
declared command,exit and time bounds,while retaining conservative failure
states. First deliverable is a pure synthetic contract/record,not a live adapter.
Its success criterion is deterministic internal consistency and rejection of
tampered/incomplete claims,never proof a particular command actually executed.

Selected architecture has three responsibilities:

|Part|First synthetic delivery|Later live work,not granted here|
|---|---|---|
|Byte source|Plain bounded packet of supplied chunk/EOF/exit/tick events;no callbacks|Owned redirected stdout/stderr streams,simultaneous raw-byte draining and EOF/exit observation|
|Evidence record|Command/time declarations,event snapshot,derived lifecycle/bytes and self-consistency hash|Bind independently observed executable/arguments/session/UTC facts to actual collection;separate durable record format|
|Outer control|Derive stop intent from virtual lifecycle;never execute it|Owned diagnostic-host watchdog and proven stop/cleanup disposition inside fixed outer limit|

Alternatives considered:record-only design leaves source/control boundary
implicit;live implementation now adds OS lifecycle/security obligations before
clearance. Selected interface-first design covers the three contracts while
keeping implementation authority synthetic-only. Existing simulator/package/
classifier/store and ordinary-ledger runtime schemas remain unchanged.

## Source boundary and input

Proposed new pure module `capture-adapter-contract.ps1` and corresponding tests
under `docs/diagnostics/isolated-fuzz`;it may import the corrected pure simulator
and its pure dependencies only. No isolation/store/native helper import. No
external stream,delegate,scriptblock,thread/task,timer,clock query,file access,
process/PID/handle operation,security query,network or acceptance activation.

Proposed constructor:
`New-SyntheticCaptureRecord -Packet <object>`.
Packet must be a plain exact NoteProperty object with fields:
Schema=1,SourceMode='synthetic',SessionLabel,ExecutableDeclaration,ArgumentBytes,
EncodingName,OriginDeclaration,StartedUTC,EndedUTC,Events.
Do not coerce types or inspect caller executable members via dispatch. Reuse
the corrected base-type strategy;check descriptors before accessing values.
Other SourceMode values are refused before any adapter action. There is no
hidden live implementation or RuntimeApproved switch that enables one.

- SessionLabel:ASCII `[A-Za-z0-9][A-Za-z0-9_-]{0,79}`;caller declaration,not a
  generated nonce,filesystem identity or actual ledger run/control name.
- ExecutableDeclaration:string1..1024 characters,strictUTF8<=4096 bytes;logical
  declaration only,no path resolution/signature/hash/file-open or execution.
- ArgumentBytes:actual byte[],0..65536 bytes,even length,strictUTF16LE text with
  no leading BOM. These are already serialized full argument-string bytes,not
  an argv array,script body or command to evaluate/reconstruct. Preserve supplied
  bytes without quoting changes. Never synthesize historical detected payloads.
- EncodingName:exact utf-8/utf-16le for output classification,not autodetection;
  OriginDeclaration:existing1..1024-character/strictUTF8<=4096-byte rule.
- StartedUTC/EndedUTC:exact invariant `yyyy-MM-ddTHH:mm:ss.fffffffZ` strings that
  parse as UTC DateTimeOffset. Caller-supplied synthetic observations,not actual
  OS clock queries. Store clock-order observation separately;never use UTC
  subtraction instead of virtual monotonic deadline checks.
- Events:actual one-dimensional System.Object[],0..4096,exact typed chunk/EOF/
  exit/tick shapes from the approved simulation spec,chunk<=65536 bytes. An
  additional packet-wide sum of ALL chunk bytes<=2097152 bounds trace copying;
  this includes chunks the simulation later rejects. This differs deliberately
  from the simulator's separate1MiB admitted-channel caps.

Validate structure/types/packet byte bounds before snapshot/admission. Invalid
packet returns a bounded refusal descriptor (Schema,SourceMode,Created=false,
Reason,OriginVerified=false,Acceptance=false,RuntimeApproved=false),no fabricated
record/hash. No promise to archive unadmitted malformed/oversized input. Caller
must not mutate input during evaluation;post-return copies must be independent.

## Evidence record and integrity

On structurally valid packet,return a separate plain record with exact fields:
Schema=1,Policy='synthetic-capture-contract-v1',SourceMode='synthetic',SessionLabel,
ExecutableDeclaration,Arguments,EncodingName,OriginDeclaration,StartedUTC,
EndedUTC,UtcOrderConsistent,Events,Lifecycle,StopIntent,CleanupStatus,
LiveActionsPerformed=false,BindingSHA256,OriginVerified=false,Acceptance=false,
RuntimeApproved=false. Constructor succeeds in creating a record even when the
derived lifecycle is incomplete/timed-out/refused;that is not collection success.

Arguments is immutable Base64/ByteLength/SHA256 of supplied ArgumentBytes.
Events is an owned plain-object array;chunk data becomes immutable Base64 plus
ByteLength/SHA256 with exact fields Kind,ElapsedMs,Channel,Data. Other event shapes
are unchanged. Do not store executable objects/callbacks. On reread,validate
canonical bounded Base64/length/hash before reconstructing simulator events.
Lifecycle is the unchanged simulator result,including separate raw admitted
stdout/stderr snapshots,nullable exit,EOF/count/index,virtual time and Package.
Package exists only for complete lifecycle and may classify output as refused.

UtcOrderConsistent is EndedUTC>=StartedUTC;false is retained evidence of declared
clock reversal,not reordered/coerced dates. It prevents a positive consistency
verdict but must not erase admitted bytes. UTC values cannot authenticate origin
or extend the120000ms virtual deadline. SourceMode and authority flags are never
inferred from hashes or a successful verifier.

BindingSHA256 covers manifest AND full stored trace,not only the final aggregate
bytes:SHA256 of the following primitive binary encoding,all little-endian,no
generic JSON deserialization or culture-dependent string serialization:

1. ASCII domain8 `TLRCAP01`;Int32 schema1;Int32 output encoding enum1=utf8/2=utf16le.
2. SessionLabel,ExecutableDeclaration,OriginDeclaration,StartedUTC,EndedUTC in
   that order;each strictUTF8 encoded with Int32 byte length prefix,then bytes.
3. Int32 ArgumentBytes length then original bytes;Int32 event count.
4. Events in supplied order:one-byte kind(1chunk,2eof,3exit,4tick),Int64 elapsed;
   chunk:one-byte channel(1stdout,2stderr),Int32 length,raw bytes;
   eof:one-byte channel;exit:Int32 exit code;tick:no additional fields.

Hash and record are self-contained consistency checks,not signatures,immutable
attestation or hostile-admin protection. A caller can fabricate an entirely
consistent synthetic record. No live result may be relabeled synthetic to bypass
runtime gates;this API neither accepts nor authenticates live collection.

Proposed `Test-SyntheticCaptureRecord -Record <object>` returns separate bounded
fields Verified,Reason,OriginVerified=false,Acceptance=false,RuntimeApproved=false.
It checks exact shape/types/source/flags,all per-field/packet limits,canonical
snapshots,primitive binding hash,UTC-order consistency and rederived Lifecycle/
StopIntent/CleanupStatus. Reject altered stored labels or snapshots rather than
trusting them. Unknown versions refuse,no silent migration. Verified means
synthetic record internally consistent,not lifecycle complete or successful exit.

## Control and failure semantics

Delegate virtual event decisions to the existing simulator without changing its
120000ms deadline or EOF/exit/cap/first-failure/trailing behavior. No real-time
waiting or scheduling in this stage. Derive exactly:

|Lifecycle status|StopIntent|CleanupStatus|Real actions|
|---|---|---|---|
|complete|none|unknown|none|
|incomplete|none|unknown|none|
|timed-out|would-request-owned-stop|unknown|none|
|refused|would-refuse-and-stop-if-owned|unknown|none|

These are simulated intentions,not actual stop requests or cleanup receipts.
Never invent exit0,EOF,natural completion or known-clean state. First failure
preserves the simulator's admitted prefix;the separate bounded trace retains
otherwise structurally valid rejected/late events as supplied input,not proof
they occurred. Invalid packet preflight creates no trace record at all.

No automatic file persistence:the existing store serializes only Package and
would omit this record's manifest/trace/control evidence. Do not pass it off as
durable capture storage. A durable complete/failed record schema needs separate
design/review/verification;old schema/files are not overwritten or migrated here.

## Future live adapter prerequisites (not implementation authority)

Live design must bind exact preapproved executable/arguments/source identities
and one owned diagnostic-host session before collection;not arbitrary PID attach,
generic shell commands or caller flags as permission. Concurrent raw pipe reads
must not substitute decoded StreamReader text or drop unknown/error channels.
Overflow/read failure requires bounded owned-stop disposition,not blocking on
undrained pipes or silently truncating to success. Exact owner handles,clock
measurement and implementation remain a separate project,not placeholders to
be filled by this synthetic API.

The outer120s per case includes capture and cleanup;existing shared500ms natural/
5000ms cleanup,cap32,access0x101000,inherit=false stay unchanged. A clock deadline
does not prove processes stopped. On timeout,cleanup unknown unless separately
proven;no arbitrary PID kills,deadline extension,retry,identity reuse or forced
exit labeled natural. A surviving partial/complete file alone cannot certify
cleanup. Signature/source drift,alerts or unreadable prerequisites stop.

Historical safety-owner/vendor disposition,current relevant safety evidence,
concrete reviewed runtime packet and separate finite launch approval remain
required. No Go/Docker/services/install/network workload or security-setting
change follows. I2/I3/full C12 acceptance remain OPEN even after future samples.

## Verification requirements and review handoff

No test allowance granted by this spec. Future plan must enumerate fresh bounded
synthetic hosts and stop/review rules. Verify known command/trace hash fixtures,
separate channels/interleaving,exact UTF16 argument bytes,malformed/oversized
packet refusal,post-return mutation,clock reversal,all lifecycle outcomes,
stop-intent/cleanup refusal of false claims,manifest/trace/label tampering,
scripted-method/property nonexecution,and no live-mode path. Test counts are not
authority;all existing106-check results and earlier failures remain unchanged.

Self-review:separated synthetic declarations from observations,intentions from
actions,whole trace cap from admitted-stream caps,internal consistency from
completion/authenticity,and package persistence from full-record durability.
No live nonce/packet/script/launch hash issued. Ask user to review this written
spec;only after approval prepare a finite implementation plan/method decision.
