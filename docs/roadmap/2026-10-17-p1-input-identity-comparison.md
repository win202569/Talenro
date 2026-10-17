# Int32 input-identity comparison — diagnostic only

## Approved scope and disposition

User approved the in-chat two-host diagnostic spike after checkpoint
`69cc76882c9a5e24fff2f61758bfbdce4ba64210`. No further implementation plan was
required for this bounded question: does the observed metadata mismatch differ
between independent and shared Int32 base identities, and at which saved stage?

Two fresh ordinary-token PS7 children, independent then shared, each capped at
30000ms. One fixed strategy: obtain boxed zero via typed Int32 array GetValue(0);
two separate arrays/retrievals for independent, one retrieval assigned to both
fields for shared. Measure actual identity before decoration; stop on failed
identity precondition, helper fidelity or infrastructure without retry/fallback.
No extra types, original F1 replay, warmups, product/test import/execution/change,
reviewer or old-budget use. The diagnostic probe is a throwaway experiment,
retained for inspection/recovery, not production or an accepted fixture change.

**Disposition: both identity preconditions held; all six snapshots reached.**
The independent case retained separate field metadata. The shared case first
showed A.Metadata matching B's returned array at AfterB, before A-array mutation.
These observations support shared input identity as a relevant fixture-isolation
factor in this controlled path; they do not establish a unique runtime mechanism,
generalize to other types/hosts, fix the fixture, or close product review I1.

## Fresh run evidence

Both hosts reported PS7.6.5 Core x64, admin=False, exit0, stderr empty,
TimedOut=false, owned-host-exited, exact four-line transport/schema accepted.
Runtime files were not installed or changed. Timing includes launch, wait and
stream collection. ExpectedResult=true means diagnostic transport/integrity,
not acceptance of product tests or any favored reference outcome.

| Slot | Mode | Elapsed ms | Measured base identity same | Stages reached |
| --- | --- | ---: | --- | --- |
| D1 | independent | 688 | false | 3/3 |
| D2 | shared | 686 | true | 3/3 |

Inputs remain Int32 zero at all snapshots. `am` and `bm` denote arrays returned by
the copied decoration helper; comparisons below use ReferenceEquals against
NoteProperty.Value, not content equality. A dash means no B array yet or no B
descriptor, not a false comparison. Both descriptors, when present, are notes.

| Mode / snapshot | A.Metadata=am | A.Metadata=bm | B.Metadata=am | B.Metadata=bm | am / bm contents |
| --- | --- | --- | --- | --- | --- |
| independent / AfterA | true | — | — | — | initial / — |
| independent / AfterB | true | false | false | true | initial / initial |
| independent / AfterAMutation | true | false | false | true | changed / initial |
| shared / AfterA | true | — | true | — | initial / — |
| shared / AfterB | false | true | false | true | initial / initial |
| shared / AfterAMutation | false | true | false | true | changed / initial |

AfterA, B has no Metadata descriptor in independent mode; in shared mode B already
has a NoteProperty whose Value matches am, although B has not been decorated.
AfterB, am and bm are reference-distinct in both modes and remain distinct after
mutation. The shared case's A.Metadata=bm is directly observed here; it was **not**
observed by the earlier localization run. Do not retroactively rewrite that run.

## Static and read-only verification

Before any child launched, static parsing found no errors in probe/controller.
AST helper-text comparison confirmed Type, WrapPrimitive and Decorate are exact
prefix-only copies from the previous localization probe; no helper logic fix.
Frozen manifest stored exact result/precondition/stage shapes and stage order,
probe/controller hashes, 41 inherited protected pins, host order and limits.
Each slot was consumed before launch; prior successful receipt and stream hashes
were checked before D2. Final read-only reconciliation found exactly two rows,
matching saved raw hashes, successful preconditions, unchanged frozen tooling and
all41 protected pins. No rerun occurred during reconciliation.
An initial read-only report-anchor check had a PowerShell path-list grouping
error (attempted to hash the packet directory). Correcting parentheses in that
read-only check verified the saved file anchors/table; no probe/controller,
manifest, raw data or invocation row was changed and no child was relaunched.

The controller used owned process handles only, separate stdout/stderr, a hidden
noninteractive no-profile child, no shell execution and inherited ordinary token.
Stderr policy was frozen as empty-only; no CLIXML deserialization or exemption.
Timeout/error cleanup branches were not exercised, so these two fast runs do not
validate those branches. Schema checks are bounded receipt checks, not a general
security proof. New hashes anchor saved bytes, not independent historical truth.

## Portable and local recovery

Portable diagnostic: `docs/diagnostics/isolated-fuzz/offline-supervision-policy.identity-probe.ps1`.
Running it again is **not** authorized by this report or by cloning. Existing
local ignored packet `.superpowers/sdd/2026-10-17-p1-input-identity-comparison/`
contains scope, frozen manifest, controller, invocation index and D1/D2 streams.
Raw/controller remain local and uncommitted; no old or new packet was deleted.
The portable table and hashes preserve the findings, not the raw process state.

| Artifact | SHA256 |
| --- | --- |
| Portable probe | `57D63924F72EB048F4E19FBBA3FB5D2E2BEC4416A421B9416F56AD296313DE01` |
| frozen.json | `3130BEA9BDF75C0F5DBEEE39067C6BB413FD19D7E0C51D906183D6D08DEE0F9B` |
| run-slot.ps1 | `6B55D8AC9986EE81DDD2F754CB3851F7A48889C95FF3055B189CDAA26D1BADE9` |
| invocations.json | `8B9D829CD945B0940BC16124F10D33C1228D7509E8DFB8B81E60A6A1360C5F24` |
| D1.stdout.txt | `6BC29DEA1242D6B33F906EDF971324E440934A35E9462A2CB91A5BEA40FBF5A9` |
| D2.stdout.txt | `0363495AED548FF246744B11075198A88DE3DAC07551C36F444D2637FE75819E` |
| Both stderr files (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

Protected product remains FE23DF3B, original35 suite61A3BACF, current unaccepted
ownership candidateCF81BD28. Historical9AA7 snapshot/old manifests/raw unchanged.
The prior35 pins and six latest-localization anchors formed41 protected entries.

## Remaining work / unchanged gates

New diagnostic allowance **2/2 spent**, no remaining host or review. Prior
localization1/1 and fixture diagnostics2/2 remain spent; fixture revision F1=false,
F2 unrun/frozen. Original repair's failed first slot remains false, its six slots
remain frozen and review ineligible. P1/I1 unfinished; R1–R7 product repair and
conditional review remain unapproved. No P2/P3, C12/release or main acceptance.

Recommendation: use these observations when scoping a narrow fixture-input
identity correction with explicit guards, rather than weakening the original
array-reference assertion. Other primitive categories and shared-identity cases
still need a deliberately specified treatment; the present Int32 experiment does
not authorize a blanket fresh-boxing fix. No further change/run is made here.
