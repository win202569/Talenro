# P2-B sequence-binding diagnostic — invalid-input and extra metadata captured

2026-10-18; base 5431f76019c37dfce9b3dbdb31eb501e007e9fdd.
Branch codex/c12-b01-task9-coordinator. User asked to continue the recommended
separately bounded evidence-only diagnostic. One NEW ordinary host, 1/1 spent.
No product repair, old-slot replay, full-suite rerun or reviewer.

## Actual observation

PID46364,1203ms,ordinary Windows PS7.6.5 Core x64,admin=False,exit0,
no timeout,stderr0,owned exit/streams complete. Outcome=DIAGNOSTIC_CAPTURED,
Error=null,Stage=Captured,FormalAcceptance=false. Capture success is NOT
a passing regression or acceptance.

The four-file portable diagnostic leaf is
docs/diagnostics/isolated-fuzz/p2b-sequence-binding-diagnostic/.
It imports the unchanged sealed F76B2DB7 candidate and uses the original
literal fixtures. sequence-helpers.ps1 is a definitions-only excerpt from the
frozen full test through BC-Bind (plus a provenance comment). The target
executes only the original SequenceExact operation order; none of the earlier
nineteen full-suite groups is replayed. No counter-fixture compile, Add-Type,
dependency install, output DLL, nested agent or source mutation.

Observed checkpoints completed:
1. Low sequence binding matched.
2. Both low sequence mismatch controls matched.
3. UInt64 9223372036854775808 encoded equal to the literal patched frame.
4. The same frame decoded equal to the independently constructed Message.
5. Final binding was called once and its actual owned result retained.

Actual final result:
```text
Schema=1
Mode=p2b-codec-only
Accepted=false
Reason=invalid-input
Bytes=null
Message=null
NextSequence=null
SourceVerified=false
LiveActionsPerformed=false
RuntimeApproved=false
FinalEqual=false
```

The original final expectation is binding-matched and NextSequence
9223372036854775809. This reproduces the high-binding assertion mismatch
without replaying the full suite. It establishes an input-refusal symptom,
not an observed wrong computed successor or sequence-mismatch result.

## Raw inventory and bounded interpretation

Inventories are deliberately taken only AFTER the actual final API call:
they cannot prime that call's metadata history. They read fixed SDK
PSMemberInfo name/type/flags and object-reference identity, NOT descriptor
Value/MemberType/Copy getters. The APIs naturally read the trusted fixture's
application values; MemberValuesRead=false describes inventory only.

Message:17 members = twelve exact PSNoteProperty application descriptors +
five reserved/hidden/instance entries. Expected:15 members = thirteen exact
PSNoteProperty application descriptors + the known two SDK entries.

| Message metadata name | Exact CLR type (System.Management.Automation) | Reserved / Hidden / Instance |
| --- | --- | --- |
| psobject | PSInternalMemberSet | true / true / true |
| psextended | PSMemberSet | true / true / true |
| psbase | PSInternalMemberSet | true / true / true |
| psadapted | PSInternalMemberSet | true / true / true |
| pstypenames | PSCodeProperty | true / true / true |

psobject and psextended match the corresponding current SDK objects by
reference. The extra three do NOT match either of those two objects; this
diagnostic does not establish trusted-source identity for them. In particular,
pstypenames is a PSCodeProperty, not a permissible business PSNoteProperty.
Names/types/flags alone are NOT a safe rule to accept arbitrary descriptors.

Read-only source trace:
- control-frame.ps1:252 snapshots Message before Expected or sequence logic.
- Get-P2BCSnapshot:85 recognizes only psobject/psextended as SDK exclusions.
- Lines91-93 reject every other name/type/hidden/reserved application candidate.
- Line270 maps P2BC_INVALID to invalid-input.
- Sequence comparison begins at259 and successor calculation at268.

Thus the recorded seventeen-member Message is incompatible with the
two-exclusion rule by construction. This is a concrete metadata-history
compatibility gap, consistent with the actual invalid-input result.
The high-number encode/decode comparisons passed; changing sequence arithmetic
is not supported by the evidence.

Limits: post-call inventory is not a pre-call branch trace. The exact moment
and SDK operation materializing the extra three were not instrumented, and
no general authenticity rule for them was proven. It is an inference from
the controlled operation order plus source that the encode/decode/comparison
history exposes additional SDK metadata before reuse. Do not assert the
precise creating statement, close the passivity finding, treat flags as
unforgeable, or claim every possible SDK metadata history is covered.
No callback-counter measurement is claimed in this run: no managed fixture
was compiled. Historical counter evidence remains in its own packets.

## Preservation and identities

Ignored packet:.superpowers/sdd/2026-10-18-p2b-sequence-binding-diagnostic/.
981 pins =963 inherited+13 previous stopped packet files+one unchanged live
candidate+four new diagnostic inputs. Rehashed after capture:981 checked,
zero mismatches; both saved stream hashes matched the receipt. Product stays
F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0.
Every old source/oracle/failed receipt/raw/trace/review remains unchanged.
No raw packet, private cache, binary or credential is committed.

Staged whitespace check passed. Git warned that a future checkout may convert
the three diagnostic PowerShell sources from LF to CRLF under local settings;
the hashes here identify the LF bytes observed during this capture, not a
promise about another machine's checkout line endings. Do not rewrite frozen
files/receipts or replay this consumed packet to reconcile that difference.

| Artifact | SHA256 |
| --- | --- |
| packet/frozen.json | 1AF2B26C9E979031D51AADA62C1F3DA1C8925B13F7539D2BB52B717C3C72A9F0 |
| packet/receipt.json | E2E6DDDD8322FF8110932154F615D206F749E5C516915DDE9D9946B94956ADF5 |
| packet/stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| packet/stdout.bin | 5A6EC404B0B171A550A112CADAEC40A12F719D64E3ECCD854623F4E04DCFBFD8 |
| source/capture.ps1 | 5F51B17F462CB442950ABC9738FB5FF39C4CBB8AAFAECE3AAD3C5BC3E5C1D097 |
| source/diagnostic.ps1 | 109E35D92D8CD0BF6C9943B1BAD343AD22EA170575547F192FEA1DD6F16A6313 |
| source/README.md | CE0383B93B298F0F12A1746E0F0079CBF073DE5E7AC47973AC69B809981A5BF5 |
| source/sequence-helpers.ps1 | 8FE8F3D323A3F9B5DE04AEAFC0A86CB900D92A26BD6B58F3B9DCC39C79314BB2 |

Controller rejects an existing packet, preconsumes one invocation, verifies
all inherited VerificationPath mappings, pins engine/live product, uses an
ordinary parent, only its owned child, separate concurrent raw streams.
Bounds10000ms execution+2000ms owned exit+2000ms drain; no relaxed process
checks. ProductImported/Called=true; CompilePerformed/FullSuiteRun=false.
Systematic-debugging drove evidence capture before any correction;
verification-before-completion limits the claim to the saved diagnostic.

## Disposition and next decision

Diagnostic1/1 SPENT. Prior repair scope3/3 spent,review0/1 INELIGIBLE.
All earlier consumed/frozen allowances unchanged; no reset/transfer/replay.
No new review, findings or declined-to-judge list; none was dispatched.
The original Important passivity blocker remains open pending a complete
repair/regression/review gate. Historical Minor comparator disposition unchanged.

Recommended next step: read-only SDK-origin/identity design investigation
for psbase,psadapted and especially the reserved pstypenames PSCodeProperty.
Before any new product rule, determine whether those descriptors can be
recognized through a fixed trusted SDK source without invoking their getters.
Keep exact12/13 application shapes, reject subclasses/masquerading members,
and fail closed when provenance is unproven. Do not simply expand a name/flag
whitelist or change raw count14 to17. A subsequent repair/test/review allowance
needs a separately agreed bounded design; none is executed by this diagnostic.

Necessary diagnostic/report/status/recovery preserved on the existing
development branch, not main. Fresh clone restores portable source/report,
not ignored raw/trace/cache/mappings/sessions/runtime dependencies or allowance
reset/replay rights. P2-A limited offline acceptance unchanged. No P2-B
acceptance/main/PS5/P2-C/P3/full C12/live lifecycle/identity/capability/process/
Job/clock/queue/release/security/runtime grant.
