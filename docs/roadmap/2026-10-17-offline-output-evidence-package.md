# Offline output evidence objects — synthetic verification complete

User approved the short design and separate4-host allowance after development
checkpoint `7e311f6480e3ae74f54955d789d1466e1839b3a0`. New standalone helper/tests:
`docs/diagnostics/isolated-fuzz/output-evidence-package.ps1` and `.tests.ps1`.
The pure classifier is imported unchanged;runtime/probe helpers are not imported.

## Contract and evidence meaning

New-OfflineOutputEvidencePackage accepts two supplied raw byte arrays,explicit
utf-8/utf-16le,integer exit and a1–1024-character origin declaration. It returns
a plain schema1/offline-output-evidence-v1 object containing both streams as
canonical Base64,length and SHA256,plus encoding/exit/origin,classification and
progress count. OriginVerified,Acceptance and RuntimeApproved remain false.
Each stream is bounded to1MiB before encoding;oversized/missing bytes or invalid
metadata refuse construction. No supplied byte array is modified.

Test-OfflineOutputEvidencePackage requires exact plain NoteProperty shapes and
metadata types,known schema/policy and false authority flags. It rejects extra/
missing fields,invalid or noncanonical Base64,invalid lengths/hashes and
inconsistent stored labels. Encoded length is bounded before decoding;decoded
length/hash are checked,then classification is recomputed from the decoded
bytes. Verification returns an explicit refusal rather than trusting the label.

Verified means **internal byte/metadata consistency only**. Nonzero child exit
or invalid text can be retained in a Verified package with Classification=refused.
This preserves failure evidence,not a successful child result. A consistent
alteration of bytes and hashes is not cryptographically detectable without an
external trust anchor. Origin and encoding are caller declarations,not attested
capture facts. Changing a claimed origin does not authenticate it.

Base64 strings remain independent of later caller-array mutation. Concurrent
mutation during construction or verification is not supported;callers must not
change input arrays/objects during those calls. The two streams are not promised
as one atomic cross-stream snapshot or a hostile same-process security boundary.

The API returns objects,not persisted files. No input-file reads,archive writer,
process launcher,OS adapter,network,watchdog,source-authentication layer or live
collector is implemented. Loading the source file/dependency is ordinary module
loading,not input-file ingestion. In-memory JSON object roundtripping was tested;
arbitrary JSON text parsing,duplicate-key detection,durable serialization,
atomicity/fsync and archival completeness are not supplied by these functions.

## TDD and captured results

Both RED hosts failed at OfflineEvidencePackageRequired before implementation.
The remaining assertions were not individually exercised in RED. Both new files
passed static parsing;both GREEN hosts then passed the entire49-check suite.

| Call | Host/phase | Exit | ms | PASS | stderr bytes |
| --- | --- | --- | --- | --- | --- |
| 1 | PS7 RED | 1 | 1633 | 0 | 367,expected assertion |
| 2 | PS5.1 RED | 1 | 2030 | 0 | 557,expected assertion |
| 3 | PS7 GREEN | 0 | 1098 | 49 | 0 |
| 4 | PS5.1 GREEN | 0 | 866 | 49 | 0 |

98 GREEN assertions cover independent abc/Base64/SHA256 known answers,empty
streams,caller-array mutation after return,JSON object roundtrip,25 package
mutation cases,noncanonical pad bits,unverified origin,encoding retention,
nonzero exits,invalid raw text,real classification recomputation,count tampering,
exact1MiB limit,invalid construction and non-object refusal. Fixtures are inline
synthetic data;no fake substitutes replace the package/classifier implementation.

All4 fresh64-bit hosts ran sequentially,NoLogo/NoProfile/NonInteractive/File,
120s each;PS5 used child-local inbox modules. No retry,extra suite,real diagnostic
or full-project test. Package allowance4/4 exhausted,separate from classifier8/8
and ledger Stage A70/70. Those previous allowances remain exhausted/frozen.

Raw index/logs remain ignored beneath the existing preparation directory.
evidence-invocations.json has4 completed entries;SHA256:
`A7B3630C849FDE89D931F027ADE49D2BBD0AC483FA21F23F4CB9600ACE70E743`.
Saved-file reads verified earlier offline-repair/original/timing/policy indices
match their recorded hashes. No raw records were changed or committed.

## Independent read-only review and disposition

One scoped reviewer inspected the two new files against the base checkpoint;
no edits,tests,helper/host launches or additional delegation. No Critical,
Important or Minor findings in the approved synthetic scope. Author checked
this against actual source/test changes and captured results. No source/test
edits after GREEN;documentation only. This is not live integration readiness.

Reviewer declined concurrent mutation,origin authenticity,durable archive/JSON
text ingestion/live capture,and recertifying unchanged classifier internals.
Author retains every limitation above;none is silently treated as delivered.
Execution totals are coordinator-verified,not inferred from the source review.

## Recovery and next safe step

Recover these tracked files with the existing classifier from the development
branch;no previous ignored fixture/cache/raw archive is needed to run the source
under a separately authorized allowance. Raw output/provenance may be sensitive:
do not commit real payloads,credentials or private event dumps. Only synthetic
fixtures and portable reports are in this checkpoint.

No actual probe identity consumed,real diagnostic/probe,Go/Docker workload,
security-setting change,main merge or acceptance activation. The prior PS5
strict-parent exit1 is unchanged;this cannot reconstruct its missing raw bytes.
Safety disposition,live raw capture,outer watchdog and exact Stage B packet
remain pending in the [readiness checklist](2026-10-17-ordinary-ledger-stageb-readiness.md).

Next:design bounded persistence for these offline objects,including failure
preservation and independent reread verification,without launching a process.
Any file writer/tests need their own design/allowance;do not infer permission
to collect a real host or activate Stage B from this in-memory feature.
