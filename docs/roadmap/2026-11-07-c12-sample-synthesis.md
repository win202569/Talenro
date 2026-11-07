# Selected-module evidence synthesis

Status: READ_ONLY_SYNTHESIS_COMPLETE; source5c237a04;44/34/6/217,NOT_ACCEPTED.
No Go/build/test/Docker/cache scan or new runtime invocation in this analysis.
Systematic debugging used to separate observations from root-cause hypotheses.

## Verified saved inputs

Read-only14ffd6 decoded every entry and checked recorded length/SHA256 in four
archives under docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2:

| Archive | Entries | SHA256 |
| --- | ---: | --- |
| portable-c12-cache-phase-v3.json | 25 | D7CD9347E29C0BE19109274301B93DCD9E7C5A288BCA5A8BF23DA8880081CF4E |
| portable-c12-libc-profile-live.json | 60 | 9A60F7E4A74C87BDB61329C6301D06DFC96A88208B4216AB726141C9FAFDF9E4 |
| portable-c12-root-verify-completion.json | 13 | 70F1509BB843F46A53AF1C600D2C64AABADB696C47ADD157FD7BD03E2B8E162F |
| portable-c12-graph-boundary.json | 13 | 85516ED24C7A2F842EC6B2839F3C0F15EC3FBA172506517E1BD714D857F78121 |

These111 entry checks establish archive consistency,not a new real-cache check.
Read-onlydb8952 checked saved successful live gates,all END statuses and ZIP/DIR
digest agreement,and calculated the table from original event nanoseconds.

## Observed timings (seconds)

| Module/version | PREPARE | ZIP | DIRECTORY | Phase sum | Directory share |
| --- | ---: | ---: | ---: | ---: | ---: |
| golang.org/x/text v0.40.0 | 0.2833952 | 0.1201420 | 12.3784326 | 12.7819698 | 96.84% |
| modernc.org/sqlite v1.49.1 | 0.6119459 | 0.7610604 | 42.4556856 | 43.8286919 | 96.87% |
| modernc.org/libc v1.72.1 | 2.0218729 | 0.3651346 | 64.0042556 | 66.3912631 | 96.40% |

Each module has three complete OK phases and two matching integrity digests.
Both selected runs have confirmed exit0,empty stderr,no timeout,and unchanged
four module files. Owner wall times are56.7170329s for the paired run and66.8872084s
for libc;phase sums exclude process startup/owner overhead.

The first two samples share a run;libc was measured later with another binary
and profile. This is not a controlled benchmark. Directory phases dominate these
samples,but neither the cause of that cost nor its behavior in native verify was
measured. Do not infer antivirus interference,disk fault,corruption or a hang.
Do not add/average/extrapolate these samples to predict the300s native run:
different coverage,parallel scheduling,cache warmth and observation times matter.

## Coverage and unresolved boundaries

Saved graph has1059 edges and101 root edges. Excluding go/toolchain leaves99 root
module declarations. All three exact measured module/version pairs occur there;
96 root-declared pairs lack evidence from these selected runs. This is a coverage
inventory only,not a3/99 acceptance score or a reconstructed native BuildList.
Graph endpoints can contain historical versions;uncached content is not inventoried.
The helper does not recursively verify dependencies. The96 are not96 known failures.

Saved native graph completed in0.3069959s. Saved native verify reached its300.0592263s
deadline with empty streams and confirmed termination,not success. These are
different invocations:fast graph does not retroactively locate the earlier timeout.
The existing root-verify-analysis explains delayed error collection;silence cannot
exclude errors already queued behind other modules. No per-module/phase progress
was retained for that timed-out native invocation,so its cause remains unresolved.

## Recommended decision

Stop expanding single-module samples solely to explain the native timeout.
The next diagnostic question should be:which native verification boundary remains
active at the deadline (graph/build-list,dispatch,ZIP/directory work,or collection)?
First design a bounded attribution method against the pinned local verifier flow,
including what it can and cannot observe. Prefer evidence from the same invocation;
a custom helper is not automatically equivalent to native verify.

Any proposal must specify tool/source binding,process ownership,privacy/output
limits,finite runtime/cleanup budget,and criteria for a useful or inconclusive
result before approval. Do not install tools,modify the Go distribution,add a
general cache scanner,raise deadlines or run another native verify from this report.
No additional runtime budget is authorized;previous stage slots remain consumed.
No repair is justified yet. Retain all evidence and the current acceptance blocker.

Author read-only review only. Source/test files and old archives remain unchanged.
Recovery:read this report plus the four committed archives on the latest existing
development branch;no previous machine's cache or ignored receipts are needed to
reproduce this arithmetic and coverage comparison.
