# Ownership marker diagnostic — complete diagnostic, NOT project acceptance

Afterd7e7eaa2,user said continue to the proposed ONE ordinary fresh PS7 diagnostic,
30000ms,no warmup/retry/PS5. New independent allowance1/1 now SPENT.
All old allowances and frozen slots remain unchanged.

## Observed result and independent read-only audit

StartedUTC2026-10-18T09:57:20.0665760Z. Parent ordinary;all saved child context
records ordinary. Existing pinned PS7.6.5 Core x64 engine. ONE child launched.
Wall28983ms,CPU39093.750ms,exit0,TimedOut=false,owned-host-exited,streams saved.
105 ordered nonempty records,64 PASS,22 markers,11 closed pairs,OpenScopes=[].
Last PASS:ShapeLookupRegression. Empty stderr;no controller/fixture/assertion
error observed. Both OwnershipProgress states passed;Termination=completed.
Classifier CaptureValid=true,ExpectedResult=true,complete-diagnostic;
FormalProjectAcceptance=false.

Independent read-only audit did NOT invoke target or reuse classifier timing
calculations:raw records compared to exact105 rules;ordinary contexts/PS7 hosts,
all literal records and marker order/nesting checked;nonnegative Int64 ticks
globally monotonic;11 pairs reconstructed with BigInt difference and nearest
ties-to-even rounding. All results matched receipt exactly. Saved stream hashes,
412 frozen pins,manifest/controller and single invocation verified unchanged.
New markers appeared in planned order:Sequential START96/END97/PASS98 and
Plain START99/END100/PASS101 (one-based nonempty records).
No raw tail repaired,no controller amended,no retry.

## Closed intervals

All values are milliseconds from frequency10000000 ticks/second. Source absolute
ticks retained here for independently checking rounding;parents overlap children.

| Scope | Start ticks | End ticks | Elapsed ms |
|---|---:|---:|---:|
| InvalidInputs | 3825152743540 | 3825155271661 | 252.812 |
| Build4096 | 3825155272401 | 3825161086124 | 581.372 |
| RootSetup | 3825161418594 | 3825161534571 | 11.598 |
| ValidationCopy | 3825161536229 | 3825336688060 | 17515.183 |
| Decision | 3825336691586 | 3825342415386 | 572.380 |
| Evaluate4096 | 3825161122792 | 3825342417294 | 18129.450 |
| Assert4096 | 3825342418168 | 3825342724534 | 30.637 |
| Check4096 | 3825161087988 | 3825342726173 | 18163.818 |
| Check4097 | 3825342726958 | 3825342872521 | 14.556 |
| SequentialIsolation | 3825370843517 | 3825375502698 | 465.918 |
| PlainControls | 3825375764100 | 3825379062757 | 329.866 |

## What this establishes — and does not

Both later ownership bodies completed and emitted PASS in this instrumented
run:SequentialIsolation465.918ms,PlainControls329.866ms. That does NOT prove a
previous timeout entered either body,nor demonstrate a deterministic hang there.
Body spans include fixture/decorations/API/equality/mutations,exclude helper
dispatch before START and PASS emission after END;not individual API time.

ValidationCopy is the largest closed measured phase here,17515.183ms,inside
Evaluate4096(18129.450ms). It combines validation and snapshot/copy;not proof of
one pure validation/helper operation as root cause. RootSetup11.598ms,
Decision572.380ms,assertions30.637ms. Do not sum parent/child times or assign CPU
shares from these overlapping wall intervals.

Only1017ms remained under the30000ms cap. A single completed diagnostic with
different source path/body/output/history does not prove stable performance or
fix the prior timeout. No timeout-limit increase,cache,guard/type/descriptor/
ownership/assertion weakening,product replacement or main merge.
The original candidate full-regression/PS5 and wider C12 acceptance gates remain.
No new independent reviewer;review0/1 unused/ineligible.

## Identities and retained records

Ignored local packet:
`.superpowers/sdd/2026-10-18-p1-ownership-marker-run/`.
Retain frozen.json,run.ps1,scope.md,invocations.json,PS7.stdout.txt andPS7.stderr.txt.
These raw/controller files are NOT committed;Git restores this report and the
already-pushed portable preparation files,not the old machine's raw records or
permission to replay. Prior packet/controller/index/streams remain unchanged.

| Artifact | SHA256 |
|---|---|
| Frozen412-pin manifest | `2200649219C275E676BD99AF81604DB3428712D11AB12F2B61FC0ADE0CEAB61D` |
| Controller | `DF34F13A86FF990A066C9C867C4AC160157094FFF99A8C232EDB01FA7F8F3D54` |
| Final single-invocation index | `7F8D8368B2F2263A85FF352ECB8756952DD4F151A13D5F344EF66993ECB9A292` |
| PS7 stdout | `9F89E5F5FFD523C893B55CE74EEEDF9D82C136500896342CF9C2F1F48D371E04` |
| PS7 stderr(empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| PS7 engine | `362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139` |
| Top-level target | `12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422` |
| Unmodified original product module | `953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA` |
| Isolated candidate module | `3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3` |

New controller derives from frozen three-phase controller. Changes only route
the new source/inventory/verifier/412-pin set,limit argument to PS7,and retain
OwnershipProgress/Termination. Process start/wait/drain/owned-host kill and
stream capture path unchanged. Manifest frozen before launch,one slot recorded
before launch. No target sources edited this turn.

Before launch:122 ownership-classifier synthetic cases and20 static preparation
cases passed again;no target imported by those scripts. No new process-safety
or security acceptance claimed from supplied-text classification.

## Next bounded step / remaining gates

Recommend read-only source-cost analysis of combined ValidationCopy and its
plain-field/type/descriptor checks,using saved evidence;no immediate rerun and
no speculative fix. Any new probe/design/implementation/run requires its own
scope;do not spend frozen PS5 or retry this1/1 slot.

Ownership-marker allowance1/1spent;three-phase/split/boundary each1/1spent;
candidate-full1/2spent/PS5FROZEN,other historical budgets unchanged.
Review0/1unused/ineligible. Candidate/I1/P1/main/C12/security/runtime/release
UNACCEPTED. Only this isolated instrumented PS7 diagnostic completed.
