# Offline output classifier — synthetic GREEN,one Minor deferred

User approved the bounded design and separate4-host synthetic allowance after
`62d0eacfadb8ed34d224b4221662e878b283461d`. New files:
`docs/diagnostics/isolated-fuzz/output-channel-classifier.ps1` and its `.tests.ps1`.
No existing runner/helper integration or real capture implementation is included.

## Contract and limits

`Get-OfflineOutputClassification` takes caller-supplied stdout/stderr byte arrays,
integer child exit and explicit encoding (`utf-8` or `utf-16le`). It returns
schema1,classification (`empty`,`progress-only`,`refused`),reason,progress count,
original byte lengths/SHA256 and the supplied exit/encoding. Acceptance and
RuntimeApproved are always false. Classification is not process/test success.

Caller arrays are unchanged. Metadata hashes the original supplied bytes,even
when classification refuses;null bytes are distinct from a supplied empty
stream. This does not copy/archive those bytes or certify how they were captured.
Caller input allocation and full-byte hashing cost are not bounded by the parser;
the1MiB-per-stream bound prevents decoding/XML parsing of oversized streams.

Strict decoder fallback rejects malformed input;stdout is validated as text but
not semantically interpreted. Nonzero/noninteger exits refuse. Nonempty stderr
requires a CLIXML header and one version1.1.0.1/default-namespace envelope. The
known subset requires progress Obj records,unique object/type references,exact
PSCustomObject/Object type names,SourceId and the fixed ProgressRecord fields.
Known type references must resolve to an earlier definition. Other channel,
type,field,attribute,element,comment,instruction or substantive trailing data
refuses. XML-permitted insignificant whitespace is treated as formatting.
Only the observed Nil current-operation form and Completed/Processing types are
supported;this is deliberately not a general CLIXML compatibility library.

XmlReader prohibits DTD,uses no external resolver and caps document characters;
no Import-Clixml/PSSerializer or generic object deserialization is used. One
leading decoded BOM may precede the required header;no encoding guessing or
replacement-character fallback. No file read,network request,source loader,
native API or child launcher is present in the classifier.

## Verification accounting

Tests exercise the real classifier with synthetic byte fixtures,no mocks/OS
adapters. RED on both hosts failed specifically at OfflineClassifierRequired
before implementation;individual later assertions were not separately RED.
Both scripts passed static parsing. Test-only oversized fixtures were adjusted
before GREEN so an empty array was not accidentally converted to null by an
if-expression;the size assertion therefore tests size refusal,not missing bytes.

| Call | Host/phase | Exit | ms | PASS | stderr bytes |
| --- | --- | --- | --- | --- | --- |
| 1 | PS7 RED | 1 | 1780 | 0 | 364,expected assertion |
| 2 | PS5.1 RED | 1 | 2089 | 0 | 549,expected assertion |
| 3 | PS7 GREEN | 0 | 963 | 44 | 0 |
| 4 | PS5.1 GREEN | 0 | 805 | 44 | 0 |

Each GREEN includes literal SHA256 known answers,empty/progress-only policy,
caller-byte preservation,referenced/multilingual progress in both encodings,
25 unsafe XML/channel cases,exit-type/code refusal,bad decoding,unknown encoding,
oversized stdout/stderr and null-input refusal.88 assertions total. No existing
probe/final allowlist or full-project suite rerun was authorized or performed.

All four hosts were fresh64-bit,sequential,NoLogo/NoProfile/NonInteractive/File,
120s each;PS5 used child-local inbox modules. No retries. Offline4/4 exhausted;
ledger Stage A70/70 is separate and remains exhausted,not replenished/transferred.
Raw local index/logs remain ignored under the existing preparation directory.
offline-invocations.json has4 completed entries;SHA256:
`DF4D38E57F6118F3FB3AD6589724EFB3B21B70F099B4E6B6E36505E7F58CA13F`.
Saved-file reads confirmed previous timing/policy/original/repair index hashes
unchanged,including the policy index's older pending-entry anomaly.

A first read-only summary attempt passed null from empty RED stdout to
Regex.Matches and printed two audit errors. No suite was rerun/no raw record
changed. A corrected read-only read cast empty stdout to a string;it then
reported all4 records,including0-byte RED stdout and both44-check GREEN logs.
This audit error is not hidden as a product failure or a consumed test call.

## Read-only review and disposition

One scoped reviewer inspected only these two new files against the base;no
edits,tests,helper launches or further delegation. No Critical/Important finding.
One Minor accepted/deferred:the numeric/reference regexes end with `$`,which can
match before final LF,and TryParse accepts surrounding whitespace. Thus a numeric
field such as AI containing `0` plus LF,or an encoded final LF in RefId,may pass
the intended canonical-number checks. This is static review evidence,not a new
dynamic reproduction. It does not grant execution/acceptance,but the subset is
slightly broader than intended. Suggested repair is absolute-end `\z` anchors
with explicit LF regression fixtures under a new bounded allowance.

Budget is exhausted:no post-GREEN source/test edits,repair or second review.
Do not call the strict subset fully closed or enable live integration on this
checkpoint. Reviewer declined universal real serializer compatibility,live raw
capture/archive completeness and prior diagnostic repair. Author retains all as
outside/unverified,not silently dropped. XML trailing formatting whitespace is
deliberately distinct from unparsed substantive trailing data.

## Recovery and next step

Tracked code/tests are self-contained and require no ignored fixtures,prior
cache or conversation for source recovery. The old byte-damaged PS5 transcript
was not replayed or retroactively accepted;its strict parent exit1 remains.
All actual ledger-native-root/child-001 run/control paths remain absent. No real
diagnostic/probe,Go/Docker workload,protection change,main merge or acceptance.
Safety disposition,raw capture,outer watchdog and exact runtime packet remain
pending as in the [readiness checklist](2026-10-17-ordinary-ledger-stageb-readiness.md).

Next:fix/test the Minor numeric-end strictness first,using synthetic fixtures
only and a separately authorized allowance. Proposed next allowance is4 fresh
hosts:PS7/5.1 RED on terminal-LF fixtures,then GREEN after replacing the three
numeric/reference end anchors with absolute-end anchors;120s,sequential,no retry.
This proposal is not approved and no such test has run. Do not treat this offline component
as approval to rerun a diagnostic host,integrate a launcher or begin Stage B.
