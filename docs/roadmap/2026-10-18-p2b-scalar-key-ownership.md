# P2-B scalar key/table ownership — stopped precondition checkpoint

Date: 2026-10-18. Base: 3e395fd511fce1619e3da4de90bf2f5efc9942ed.
User approved the bounded type-sensitive repair design and NEW 3 ordered
ordinary pinned PS7 hosts plus 1 conditional independent read-only review.
Historical source, observations, raw receipts and all old frozen budgets remain
unchanged. This report is an unfinished checkpoint, NOT acceptance.

## Outcome

First host stopped unexpectedly at SDKProof:literal.String, history 0, before
any product API. New scope: 1/3 hosts spent; remaining 2 FROZEN; review 0/1
ineligible. No product edit, seal, focused/full target or reviewer. Do not replay,
repair frozen test inputs, consume old slots or silently resume after this stop.

PID 45752, 1570ms, exit 1, INCOMPLETE, ExactExpected=false. No timeout;
owned process exited and separate concurrent streams drained. stderr empty.
All 1109 mapped pins intact after execution and read-only reconciliation.
Execution/owned-exit/drain bounds remain 60000/2000/2000ms; no security changes.

## What the saved data actually establishes

- All 54 controlled key/table identities validated: nine declared scalar/array
  types x default/local storage x same-wrapper/same-base-distinct/
  equal-content-independent. Non-string default same-base wrappers share key/
  table; local-mode and string distinct wrappers do not. Independent equal-content
  instances stay separate. These use fixed SDK constructors/factories and SDK
  key-method oracle ONLY in fixtures; ProductChecked=false for all 54.
- 21 normal SDK proofs validated: Byte, UInt16, UInt32, UInt64, Int32, Int64,
  Boolean, each at histories 0/1/2. Fixed descriptor classes/names/flags and
  owner-key/table plus own internal/constructor consistency passed. The recorded
  counter totals are 0 for those completed proofs only.
- Proof 22, literal.String history 0, is incomplete. Raw inventory contains only
  psextended (exact PSMemberSet, hidden/reserved/instance true). Its own constructor
  owner matches its own instance owner, but OwnerSame, OwnerKeySame, OwnerTableSame
  and ExtendedTableSame are false relative to the separately acquired shim owner.
- The exact failed assertion is SC-ValidateEntries line 42:
  reference equality of the supplied raw table to the shim owner's table.
  It precedes the descriptor loop. Neither raw-table/owner-table IDs nor that
  predicate's operands were saved, so do not fabricate those reference IDs.
- Proof 22 SDKCount, ApplicationCount and Counters are null, NOT zero. The final
  descriptor's constructor match is not validation of the entire proof.
- Initial proofs incomplete: 22/146 partial records, 21 validated; output-leaf
  proofs unreached, string histories 1/2 and byte-array literal proofs unreached.
  Additional API histories, 35-row hidden matrix, 182 malformed-metadata controls,
  19 direct guard key/table controls and original full regression all unreached.
- Fixture compiled once in the approved ordinary target and observable controls
  passed. Product definitions imported, ProductCalled=false. No new product
  behavior or fresh-output-copy compatibility was verified.

## Read-only interpretation, not a further execution

The test acquires psextended through DG-RawMembers in one helper call, then
psobject through SC-Owner in another. For strings, unlike ordinary numeric/array
bases, independently created wrappers need not share storage. Thus the fixture
still assumes a stable actual wrapper across separate helper operations. The
saved mismatches and controlled identity cases support that explanation; they
do NOT reconstruct a complete string wrapper/reference trace or prove precisely
where a new wrapper arose.

Official SDK source uses wrapper-sensitive keys for strings and local-storage
mode; it does not justify a blanket same-base or value-equality exception.
[SDK key rule](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L968-L986).
No product weakening follows from this fixture failure.

Next bounded design must keep the ACTUAL wrapper and its raw collection together
across fixture observations (fixed-field reads, containing reference arrays,
snapshot operands before assertion), and distinguish public input types from
string/bool output checks. It must establish that an ownership rule can be
applied without changing receivers merely by observing them. No new target,
fixture repair/retry or product edit is authorized by this report. Preserve this
failure as evidence, not a product refusal or expected hidden-decoration RED.

## Frozen evidence

New leaf: docs/diagnostics/isolated-fuzz/p2b-scalar-key-ownership.
New ignored packet: .superpowers/sdd/2026-10-18-p2b-scalar-key-ownership.
portable-summary.json is a DERIVED repository artifact bound to the raw hash;
it was written after capture and is not one of the 12 immutable inputs.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 1DED7EEAA0EF5B1B22C84B62232D42408219E894D5627621C61185B790CDEE99 |
| invocations.json | 67D0D621493A58F006744486AD3EBF909EA6610F051D89F069E2B858B37369D6 |
| red.stdout.bin (66126 bytes) | 7CF77B39A07A2BD990B5CDE3503752A14F295F87A116C4004A9ACB3A1FAFD4D1 |
| red.stderr.bin (empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| baseline and unedited new product | B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039 |
| target.ps1 | 58E045412FBA91B190EB5C96A6E12EBF1291D79FF739651D106BB33A8B94759F |
| pinned pwsh.exe | 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 |
| pinned SDK DLL | 2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B |

1109 pins = inherited wrapper diagnostic 1090 + its ignored packet 4 + report/
portable summary 2 + new immutable inputs 12 + new baseline snapshot 1.
The historical failed candidate 18DA24E1 remains unchanged and must not be adopted.
The public hidden-scalar metadata gap M1 remains OPEN in B468B564.

## Preparation and handoff

All new PowerShell files passed STATIC parsing before the only target launch.
An initial host-side JSON file inventory exceeded output budget and could not
parse; reread at adequate output budget before preparing any target. No target
was launched by those inventory reads. No frozen file changed after Prepare.

Preserve 14 new portable files (13 prepared source/test/fixture/controller files
plus derived summary), this report and current-status/recovery updates on the
existing development branch. No main merge, broader test claim, formal P2-B,
PS5, P2-C/P3/full C12, live, identity, capability, security or runtime approval.
All authority flags remain false. Clone restores portable material, not ignored
raw, actual references, runtime/SDK/cache/sessions/mappings or replay rights.
