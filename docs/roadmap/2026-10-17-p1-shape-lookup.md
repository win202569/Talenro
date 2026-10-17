# Ordinal shape lookup candidate — contracts pass, complete regression still times out

Approved bounded design afterd292fc87: replace repeated field-name list scanning
with per-evaluation fixed-schema ordinal sets, retain all per-value checks.
User explicitly approved expected-RED exemption for this pure performance refactor.
Three ordinary-token children maximum, configured30000ms each: new-contract
baseline PS7, candidate full PS7, candidate full PS5. Any failure/timeout stops;
no retry,warmup,fallback,reviewer,old-slot use or main merge.

## Implementation and contracts

Candidate SHA256:
`953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA`.
Baseline6CF0D354 preserved in ignored packet. Metadata factory allocates a fresh
ordinal dictionary per public evaluation. Shape helper lazily creates an ordinal
HashSet from the existing fixed allowed-name list, keyed by literal root or
already exact-token-validated observation Kind. At most root+eight valid kinds
on this API path. No key/list is learned from input member names, no input/result
admission is cached, no dictionary is stored globally or returned. Helpers with
no key create a fresh set rather than sharing it.

Exact object type/property count checks remain before membership checks. Each
property must still have an admitted case-sensitive name and NoteProperty type;
every admitted value still passes the same passive/type/extended-member/
numeric/token/array checks. Count uses original Names.Count, not deduplicated set
count. Existing copy/decision/error behavior unchanged; only two API shape calls
receive keys. Existing scalar SDK reflection and descriptor defenses unchanged.
Repeated fixed observation-name array construction was NOT optimized in this step.

New six behavioral groups:
RootOrdinal (all six names),FieldBounds (missing/extra/same-count replacement),
ObservationOrdinal (all six base names),KindIsolation (interleaved time/control
and malformed later instances after earlier same-kind checks),RepeatedCalls
(result ownership and edited input reevaluation),DescriptorSafety (root/observation
script descriptors refused with zero callbacks). Literal expected results use
original35 comparison helpers, not mocked acceptance or source-text assertions.
No claim all possible shapes are covered by these six groups alone; existing
suites exercise remaining kinds where completed.

New test file was written and baseline run BEFORE product change. Its six groups
passed against old code, as expected for semantic-preserving refactoring; no RED
demonstration was claimed. New wrapper runs these contracts, then unchanged
existing complete reflection-reuse regression entry. Original35 and ownership
suites/old wrapper were not edited or replaced by diagnostic copies.

## Runs

| Slot | Engine/workload | Exit | Wall ms | CPU ms | Result |
| --- | --- | ---: | ---: | ---: | --- |
| 1 | PS7,new6 against baseline6CF0 | 0 | 1045 | 1140.625 | PASS |
| 2 | PS7,candidate,new6 + complete existing flow | -1 | 30071 | 38796.875 | TimedOut,ExpectedResult=false |
| 3 | PS5,same complete flow | Not launched | — | — | Frozen |

Both launched PS7.6.5 Core x64/admin=False,hidden/no-profile/no-shell,stderr empty,
owned-host-exited,CaptureValid=true,FormalProjectAcceptance=false. Configured
limit30000; actual slot2 total30071 includes71ms cleanup/collection overhead.
CPU is cumulative child/thread time,not per-group CPU or percentage; it can
exceed wall elapsed. No controlled before/after performance result was measured.

Baseline raw9 lines/6 PASS. Candidate raw77 lines/59 PASS: new6 +reflection3
+fixture8 +original35 +ownership first7 through PassiveDescriptors. Remaining
ownership SequentialIsolation,PlainControls and Complete,ownership final LIMIT,
old wrapper final PASS and new wrapper final PASS absent. Complete inventory
would be83 lines/64 PASS. No assertion failure/error was saved before termination;
missing groups cannot be assumed passing. Last PASS does not identify the exact
executing statement. PS5 has no row,child or streams. No retry followed timeout.

The additional completed ownership group relative to the previous candidate is
not proof of speedup: workload includes six added contract groups and host/runtime
conditions are uncontrolled. Candidate still fails to demonstrate full completion
within the retained30s gate. Do not call the optimization effective or accepted.

## Integrity and limits

173 immutable pins matched before/after:165 inherited excluding mutable product,
five prior packet anchors,new tests/wrapper and exact baseline snapshot. Product
candidate sealed after baseline and before slot2; candidate/seal/frozen hashes
checked around execution. Static syntax passed,explicit source diff reviewed;
no original test or prior record drift. Raw9/77-line prefixes independently
matched frozen inventories. Original35 SHA25661A3BACF,ownershipC4905925 unchanged.
Both engine digests pinned in inherited records.

No independent code/security review or PS5 proof. Full-success/PS5 parser paths
not exercised; no claims of those paths validated. Completed individual contracts
do not close I1/P1 or replace outstanding regression/mutant/review requirements.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `51EEC09AA3057D8580F0CA64BB03A6C10B6A38E1995DA2950F2F629FEBA17139` |
| run.ps1 | `52AF2E387FB4C433C944C122005C6876D755128DBF6700C01CE094EB217409BD` |
| candidate.json | `FB089A89EBC7C16CAE8F0052FE7B592413C12152BB50585C1BB589400A791DC7` |
| invocations.json | `498FA51C946B9F755D0AA2070160F13B9EA8D667ED59A918CFDE57A9047FE928` |
| Baseline.stdout.txt | `4DB7028F307E6DE249A68D4AE6F0F95C775371E8D9F08027E916492E52779FE1` |
| Baseline.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| PS7.stdout.txt | `630DC774EB58B6E3B38AF177DA457A315C79297C9E1640A98060E1EE5AD72AAC` |
| PS7.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| offline-supervision-policy.ps1 | `953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA` |
| offline-supervision-policy.shape-lookup.tests.ps1 | `9DAA38EF1F19CA2320D17FCA35CDA87A1B4DFBDFCE960AF40EFB223B9E245B6B` |
| offline-supervision-policy.shape-lookup-regression.ps1 | `12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422` |

Ignored packet `.superpowers/sdd/2026-10-17-p1-shape-lookup/` preserves scope,
baseline/controller/frozen/seal/index/raw. Older packets/verdicts retained.
Portable candidate/new tests/new wrapper/report/status/recovery saved on existing
development branch,not main. Clone restores inspectable unfinished work,not ignored
evidence/runtime or rerun authority. Digests identify local measured bytes,not
independent attestation; checkout line conversion may alter them.

## Handoff

New2/3 spent,remaining PS5 one frozen. Previous independent full1/2/one frozen,
reflection-reuse3/4/one frozen,extended-view3/7/four frozen stay separate; other
diagnostic budgets unchanged,review0/1 ineligible. Candidate953CC2A8 unaccepted,
I1/P1 unresolved; no main/C12/project/live/runtime/release acceptance.

Next review this change and residual performance/verification design before
stacking another optimization. Two attempted performance changes have not removed
the complete-regression timeout; this does not justify removing safety checks,
splitting a suite to declare PASS,raising the cap,automatic PS5/old-slot use or
another unchanged replay. No further edit/execution is authorized by this report.
