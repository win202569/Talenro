# Inline observer actual ownership timing — 39/39 within 30 seconds

Date: 2026-10-18. Base b90d28711a30aa7489c702259c05dc4a161b8b34.
Branch: codex/c12-b01-task9-coordinator.

## Result

The user continued the separately proposed candidate-executing timing diagnostic.
Exactly ONE ordinary-user PS7 target host completed in19954ms under the unchanged
30,000ms execution limit. Exit0, no timeout, stderr0, owned exit/streams complete,
ControllerError/CheckerError null, all666frozen input pins intact.
Outcome PASSED_WITHIN_MATRIX, FormalAcceptance false.

RESULT: ISOLATED_WITHIN_MATRIX cases=39 calls=80 scalarCalls=29 metadataCalls=1

All39behavioral matrix cases and control-after passed. Final timing counters
confirm80/80decision calls,29/29direct scalar calls and1/1direct metadata call
attempted/completed. Both known-sharing controls,15direct scalar vectors and the
normal before/after controls also passed. No ownership-gap result was emitted.

Limits:30s execution +2s owned exit +2s stream drain. New allowance consumed before
launch. No retry, warmup, later suite, repair, timeout increase, assertion/history
removal, product-source change, promotion or main merge. Old slots remain frozen.
StartUTC2026-10-18T13:33:24.1364897Z, owned targetPID46380.

## Output and independent read-only checks

The same ordinary controller invoked the unchanged strict timing checker on SAVED
TEXT after target exit/drain. No second child process or candidate import by the
checker. Postprocessing lies outside the target timer.

Checker outcome COMPLETE_TEXT_ONLY:1340TIME events (670complete stage pairs),
153base lines,53ownership FACT records, no open announcement. Stdout1493lines.
Last event1340:control-after/oracle-after LEAVE, all declared call counts complete.
An independent read-only rehash reconfirmed all666pins after execution.

The checker itself still reports ProcessReceiptValidated=false, RuntimeVerified=false
and FormalAcceptance=false because it validates text only. Parent receipt separately
confirms the bounded target's exit, streams, pins and stderr; neither implies
deployment/security/runtime acceptance. PASSED_WITHIN_MATRIX means precisely the
stated synthetic matrix on the fixed PS7 engine, not a general ownership proof.

## Completed-stage timing

Frequency10,000,000ticks/sec; completed stage body total18.6676386seconds.

| Completed stage group | Seconds |
| --- | ---: |
| Prior-history scans, before+after | 11.1499656 |
| Literal oracles, before+after | 2.5720527 |
| Candidate decisions,80completed | 2.0320015 |
| Capture/select | 1.7960091 |
| Current-owner scans, before+after | 0.6874915 |
| All other completed stages | 0.4301182 |

Prior scans remain59.73% of completed body, but the complete matrix fit30s.
Prior-before5.4358639s; prior-after5.7141017s.
Recorded TIME emission before the last event0.2967023s; unclassified remainder
before that event0.4678546s. Child last timestamp19.4321955s.
The final emission's own duration, parent launch/parse and unrecorded tail are not
individually attributed. Do not assign the parent/child clock difference to scans.

## Matched-prefix comparison, NOT a repeated benchmark

Read-only comparison with the earlier original-observer timed-out capture matched
the first935event structures:sequence,case,stage,boundary,pair and every attempted/
completed call counter. There are467complete intervals plus the matching next
entry announcement. No interrupted work is inferred from that entry.

| Same completed prefix | Original observer | Inlined observer |
| --- | ---: | ---: |
| Completed stage body, seconds | 28.5570738 | 11.5535808 |
| Completed prior scans, seconds | 19.6150681 | 5.5616888 |

This is a useful observed reduction over matched event coverage. It is not a
controlled/repeated benchmark or an isolated causal estimate; process state,
system load and instrumentation can affect timings. The original full matrix
never completed, so no original full-duration speedup ratio is claimed.
No further optimization is justified solely by this one successful run.

## Frozen source boundary

Only V-Others had previously been expanded in the optimized test; the original
helper definitions, SDK invocation/argument arrays, all checks/scans/history,
39cases, literal oracles and timing schema are unchanged. Offline v2 equivalence
45/45 was completed before this candidate execution. No source changed this turn.

Controller inherits657frozen inputs +5previous raw anchors +4new portable files:
666pins. Twelve fixed hash literals matched actual source/engine files before
launch; controller parse errors0; ordinary token was independently checked.
No candidate/test/logger/checker/inventory edits, installs or settings changes.

## Evidence

Ignored packet: .superpowers/sdd/2026-10-18-p1-inline-ownership-timing.

| File | SHA256 |
| --- | --- |
| frozen.json | 803585E86355FD1664C36A29495BFA8BFF9C1FCC3596E8ACDE403ED8C699FD62 |
| invocation.json | C012FE83EB47D9FB7736A12A78BC45FD61BEF4EA68FB9AD6114B1CB4A2D268C9 |
| stdout.txt | 442DE40B2FE3137AC14C62C5C3EB79B427B138A2B6022B819B7B2E16777B925B |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| analysis.json | 65110E040EE6A0083CB64BB8B1EFE3BDD8CFFC64E878A1B6750C051727F3EBD6 |
| capture.ps1 | 9603B184A6A7B5F2E334EF6CFC0BA20F2DB1BB5BBDD27E69BA49ED0625C2AD88 |
| README.md | B18C115DBC77A7D844990A627913067197631EAE683A09C811094ED19F3545DD |
| source-pins.json | 426E4E7DF454577AE9E0C79D4ED86CC1258250B166E284F3C45281127FE43751 |
| .gitattributes | 0CD209979B0AD7DB7DDC9CB658156B0A9C18859999F71E541861F3AB58E7AB11 |

Unchanged candidate1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8;
optimized target physical8F41A9850474720F5AE2525E3E905018884DA2CB55A92E8C669F339965173E04;
strict checker7E140A1BA753E1CBA836B36753FB3B741F615CB5941E9C1FD4990F9AF476B4D4.
The target's historical prepared-only comment is not broad execution authority;
only this separately approved one-shot controller authorized this run.

All historical failures, ignored raw receipts and successful diagnostics remain
untouched. Portable controller/source pins and this result/timing table preserve
progress on a fresh PC; cloning does not restore raw receipts or engine cache.
As documented earlier, Git normalizes the optimized target's final LF to CRLF,
so its physical run hash is not a universal fresh-checkout hash.

## Remaining boundary and next recommendation

This is the first complete result for this optimized isolated matrix, not project
completion. It does not close I1, prove sensitivity to every product mutation,
promote the isolated candidate into production, replace independent review, pass
all project tests or establish security/runtime/C12 acceptance.

Next: perform a read-only reconciliation of the remaining acceptance checklist
and candidate-promotion conditions against the saved evidence. Identify required
versus already satisfied gates before proposing further implementation/tests.
Do not repeatedly rerun successful diagnostics, reuse old allowances or merge main
merely because this matrix now passes. No further run/promotion is authorized here.
