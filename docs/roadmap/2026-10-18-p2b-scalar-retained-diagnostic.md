# P2-B retained scalar diagnostic: evidence complete, hidden decoration gap confirmed

## Scope and outcome

User approved a corrected NEW fixture and ONE bounded evidence-only host after
e66fe5059f326402e1f6e3032d476f58adfda1c6. Outcome DIAGNOSTIC_CAPTURED means
complete observations, NOT a regression/acceptance pass. All 35 rows completed;
14 hidden-decorated inputs were accepted contrary to the plain-value contract.
M1 is now an observed controlled input-validation gap, not just missing coverage.
Product source remains B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.
No product repair, full-suite rerun, new reviewer, main merge or formal acceptance.

## Corrected fixture and limits

Portable source: `docs/diagnostics/isolated-fuzz/p2b-scalar-retained-diagnostic/`.
The prior failed fixture used public integrating Add, which copies descriptors;
its original-object identity assertion was invalid. The unchanged CounterNote
Copy also does not preserve Hidden. Old source/raw/receipts were not edited.

The new test-only construction uses Add on the exact pinned SDK internal raw
collection, with Hidden set before insertion to keep hidden bookkeeping coherent.
It discovers the actual retained descriptor and records its class, fixed-name,
hidden/reserved flags and reference identity; original-versus-retained identity
is observed, not assumed. Each partial row and precondition is saved before
asserting. Inventories read fixed metadata fields, not descriptor values.
This deliberate private-SDK construction is not product behavior, an origin
guarantee, or a boundary against arbitrary same-process private reflection.

The definitions-only data helper is byte-identical to the previous probe:
8FE8F3D323A3F9B5DE04AEAFC0A86CB900D92A26BD6B58F3B9DCC39C79314BB2.
Existing literal frames and managed fixture remain unchanged. Each vector gets
fresh independently boxed scalar/array data. CounterNote compiles once in this
ordinary target only; its observable positive control precedes resetting counters.
Counters cover this instrumented subclass only, not arbitrary getters or all
possible callbacks. No output DLL, install or security-setting changes.

## Exact matrix

Each row below covers the same seven modes in order: encode Message OriginTick,
encode Message Payload, bind Message OriginTick, bind Message Payload, bind
Expected Schema, bind Expected BindingDigest, decode input byte array. Scalar
types are long for OriginTick and int for Schema; array inputs are byte[].
Accepted reasons are encoded for encode, binding-matched for bind and decoded
for decode. Every refused result is invalid-input.

| Raw rows | Variant | Accepted | Refused | Counter total per API |
| --- | --- | --- | --- | --- |
| 1–7 | clean | 7 | 0 | 0 |
| 8–14 | visible exact PSNoteProperty | 0 | 7 | 0 |
| 15–21 | hidden exact PSNoteProperty | 7 | 0 | 0 |
| 22–28 | visible CounterNote subclass | 0 | 7 | 0 |
| 29–35 | hidden CounterNote subclass | 7 | 0 | 0 |

All seven precondition booleans were true for every row. All 28 inserted
descriptors retained their observed original reference, exact expected class,
requested Hidden flag and nonreserved status; tag count and ordinary count were
one. All 35 base values and root field descriptors retained identity across
the API. Post-inventory counters equal immediate post-API counters (all zero).
SourceVerified, LiveActionsPerformed and RuntimeApproved stayed false throughout.
No fixture exception, timeout or stderr; Error=null, Stage=Captured.

Source inspection explains the result: Assert-P2BCPlain in the unchanged
structural candidate enumerates public PSMemberSet.Members for scalar/byte-array
decoration detection. Unlike its raw root snapshot, that view omits hidden
descriptors. Controlled hidden entries were therefore not rejected. These
observations do NOT demonstrate code execution, callback activation, production
exploitability, universal passivity or broader security acceptance.

## Capture and integrity

One ordinary PowerShell 7.6.5 Core x64 target, PID43056, 1982ms, exit0. Owned
exit and separate concurrent stdout/stderr drain completed. Bounds were 10000ms
execution + 2000ms owned exit + 2000ms drain; only the owned target was in scope.
Engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139;
SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B.
All 1030 mapped pins rehashed with zero drift after capture and again read-only.
Inventory: inherited1020 + four prior packet files + five new source files +
prior failure report. Live product, engine and SDK are additionally checked.

Ignored local packet: `.superpowers/sdd/2026-10-18-p2b-scalar-retained-diagnostic/`.

| File | SHA256 |
| --- | --- |
| frozen.json | 6221862E99D0504124AF68ED75B4B3B357A38663ED1B95199FE7FE8C2E879D84 |
| receipt.json | F729BC02402677F9415DAEC4610E08097FA02E3AA2E681AC8F103A20D92B93B8 |
| stdout.bin | 5268B70A26139A2C03C27F706012142B3FDCBBC7B6D81963636E82FDAE9E3867 |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Sealed new diagnostic.ps1 C9FF554B1AED7AF99A62FF2C389E18A51676FD24060CBA047B967A611F5CD818;
capture.ps1 037628C125ABDDF7F77D354F1840214CA594BD315777337C7A0C6942AE2E9AC3.
Before freezing, all three new PowerShell files parsed without errors. No new
source changes, fixes, retries or extra target hosts after freezing.

## Remaining decision and recovery

Current new allowance1/1 SPENT. Previous failed scalar diagnostic1/1 remains
INCOMPLETE with only seven clean controls; structural3/3+review1/1 and every
earlier spent/frozen scope remain unchanged. Historical full33PASS describes
that suite only and does not cover this newly demonstrated gap. No new review
was requested or performed; the prior review transcript is preserved verbatim.

Recommend a separately approved repair design: use passive raw scalar/byte-array
inventory rather than public hidden-filtering enumeration, preserving SDK
structural compatibility and exact types. Create a NEW candidate/test leaf and
agree finite RED/focused/full verification and conditional review before any
execution. Do not mutate or replay sealed leaves or reuse consumed host slots.
No implementation or new execution allowance is implied by this recommendation.

Recover from the latest codex/c12-b01-task9-coordinator commit, not main. Source,
this complete matrix summary and recovery docs are portable. Clone does not
restore ignored raw packets, runtime/SDK, mappings, caches or conversations and
does not reset budgets/replay rights. Formal P2-B/PS5/P2-C/P3/full C12/live or
security acceptance remains ungranted; P2-A's separate limited scope is unchanged.
