# P2-B scalar reserved metadata: read-only comparison complete

Base1bd6f0af93fc669a3bf7e1106c4d0a442ad22b24. User approved ONLY the recommended
read-only structure comparison after the failed scalar repair. This report
records findings, not a new design approval or implementation/test allowance.
No product/test/controller/fixture import, invocation, modification or compile;
no new target host, reviewer, runtime packet, install or security change.
Only existing source, JSON/raw outcomes and official version-tagged source were
read. Report/status/recovery preservation is covered by standing instructions.

## Confirmed distinction

Saved first-run scalar matrix, strict CP936 read-only decoding:

| Variant | Cases | Raw entries before API | Nonreserved entries | Probe tags |
| --- | --- | --- | --- | --- |
| clean | 7 | 2 each | 0 each | 0 each |
| visible exact note | 7 | 3 each | 1 each | 1 each |
| hidden exact note | 7 | 3 each | 1 each | 1 each |
| visible subclass | 7 | 3 each | 1 each | 1 each |
| hidden subclass | 7 | 3 each | 1 each | 1 each |

The focused failed first clean input independently saved raw2/nonreserved0.
Thus a plain-value contract cannot be implemented by requiring the entire raw
collection to be empty. Conversely, these counts are observations, NOT an
allowlist: accepting any two entries or ignoring Reserved flags would not prove
legitimacy. Hidden application descriptors must still be rejected, as visible
ones already were in baselineB468B564. The known-failing18DA24E1 candidate rejects
all raw entries at control-frame.ps1:43 and masks copied-value failures at:56.

## SDK mechanism: static explanation, not recovered scalar instance identity

Official PowerShell v7.6.5 source shows that an extended member set shares its
owner's instance collection. Reserved-view lookup can materialize metadata;
generators retain existing same-name entries. Public enumeration suppresses
hidden entries and can materialize additional reserved views. This explains why
an empty raw collection is not the right invariant and why a generator return
does not authenticate origin. [Constructor](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L2939-L2954),
[generation and lookup](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4056-L4144),
[lookup dispatch](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4370-L4396),
[enumerator](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4649-L4694).

Test source reads the scalar PSObject and psextended views before saving counts
(scalar.tests.ps1:12-16,84-99). Inference: SDK materialization is consistent with
the two observed reserved entries. However SP-Snapshot stores only totals and
the injected tag descriptor, not the other entries' names/classes/owners. It
does NOT establish that these two actual scalar entries were the expected views
or met every owner/code-reference condition. No new instance was constructed
to fill that gap in this read-only task.

## Existing root validation provides a comparison, not a scalar proof

Get-P2BCSnapshot in the unchanged structural candidate already distinguishes
business descriptors from narrowly validated SDK metadata. Its relevant rules
are preserved in both historical and failed candidates; see the failed
candidate control-frame.ps1:90-134 and structural descriptor fixture:42-104.

| Category | Checks already present in root path |
| --- | --- |
| Receiver | Exact shim and member-set CLR types; exact raw collection type |
| Ownership | Exact PSObject owner; constructor owner and shim owner reference equality |
| Every SDK entry | Exact name/class; hidden, reserved and instance flags; owner reference |
| psobject / psextended | Mandatory; descriptor identity agrees with validated receiver |
| Optional base/adapted views | Exact internal-member-set class and internal owner |
| Optional type-name view | Exact code-property class; fixed SDK getter reference; null setter |
| Non-SDK entries | Only exact declared business descriptors; no hidden/extra/subclass entries |

Names here describe existing project requirements, not proof that scalar
descriptors possess them. Root fixture evidence previously verified owner/code
references on Message/Expected objects with five SDK entries and a separate
two-SDK-entry application-descriptor case. Focused trace SHA256:
1F9EE5357CF7DB4CE8D7C023BE6C017EDDB9E160D26F433AC5091C1329F43D12,
matched its saved invocation's TraceSHA256. Root evidence cannot be transferred
to boxed scalars, byte arrays or freshly copied outputs without their own proof.

The previous metadata evidence remains unchanged:
docs/diagnostics/isolated-fuzz/p2b-sdk-metadata-investigation/sdk-metadata.json,
SHA D6F18F947481CAE3C18B9870E3C6822114301254FF6382A60802441D21087E8B.
Its fixed field declarations and selected method IL were read as data only;
no SDK method was executed. Correct owner references must remain actual PSObject
references, not accidentally unwrapped base values (existing Test-P2BCReference).
Structural compatibility is not unforgeable factory provenance or an isolation
boundary against arbitrary same-process private reflection.

## What remains unknown and recommended next decision

The failed copy guard replaced its underlying exception. The precise failed
fresh-output predicate and complete member inventory were not serialized.
Post-API counts, result flags and retained identities in its first partial row
are null, not proven zero/false/true. No decorated input reached that API run;
34cases were unreached. Static reasoning cannot recover those missing facts.

The evidence supports revising the scalar rule to validate legitimate reserved
structure and refuse every additional descriptor, NOT to skip all hidden or
reserved entries. Before another implementation, present a bounded design that
explicitly covers scalar/array and fresh-output receiver/owner identities,
normal/reused SDK view histories, literal refusal expectations and preserved
underlying failure diagnostics. That design needs approval and a fresh finite
verification scope; this read-only comparison grants neither. Do not reuse the
last frozen full-test slot as a probe/retry, weaken guards, mutate sealed files
or promote old incomplete receipts. No new candidate or tests were written.

## Integrity, accounting and recovery

Read-only rehash: all1056 inherited mapped pins intact. Saved focused stdout
D5EE3E7EE933B29BE950F2894A094034016817659D1DF86E68417363BA5AC548 and empty stderr
match the unchanged receiptB847DB2BBFCAA831FE84DD63E85F73E47617DF4DEFD9DB1571795B8A84049CDE.
Failed candidate remains18DA24E131A63C5AFA6320D4066E971B117E2B57185D360AC8F07994144BF474;
historical structural candidate remainsB468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.
No new target/reviewer;original2/3spent,last1/3FROZEN,review0/1INELIGIBLE. Every
older consumed/frozen scope and M1 remain unchanged. Historical33PASS belongs
only to the historical tested scope, not to the failed scalar repair.

Recover latest codex/c12-b01-task9-coordinator,not main. This portable report
and source references explain the next decision without an old conversation.
Clone does not restore ignored raw/runtime/SDK/cache/mappings/sessions or reset
budget/replay rights. The failed candidate must not be adopted. No formal P2-B/
PS5/P2-C/P3/full C12/live/runtime/security acceptance;P2-A's separate scope stays.
