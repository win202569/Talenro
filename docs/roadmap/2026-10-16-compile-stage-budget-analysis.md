# Stage compile001 — saved-only cumulative budget analysis

## Scope and conclusion

User continued from the proposed saved-only analysis after checkpoint
943008477eebf6d8572ed8ce2bfdc7b0ccbb38c3. This increment reads the existing
[stage compile001](2026-10-16-leaf-stage-retained-compile-001.md) receipt,log,
manifests and fixed source text. No native Run,Go/binary execution,live PID query,
cache/tmp traversal,full SDK/product inventory,policy change or retry occurred.

The observed cap is not a compiler failure or absent trigger. Delayed activation
changes when collection starts,not which later job members use its budget.
The original registry still counts32 cumulative attempts,without filtering vet,
recycling exited entries or starting another cohort. This run recorded5 non-vet
entries,then27 vet entries (84.375% of32). The latter admissions span about0.95s.
The policy correctly marks attempt32 capped and the consumer correctly refuses it.
The real diagnostic remains STOPPED; no claim that this analysis fixes coverage.

## Recorded timeline

Times below are saved UTC observations relative to native receipt start,not
process creation/exit times. Equal UTC readings can still have distinct ordered
monotonic counters. Entry numbers are local report IDs,not authenticated identities.

| Saved event | Seconds after native start | Recorded entries |
| --- | ---: | --- |
| Trigger activation | 49.846 | No earlier registry cohort in this mode |
| First admission query | 49.849 | Query1 |
| conhost/first compiler admissions | 49.853 | 1–2 |
| Second compiler admission | 50.169 | 3,query21 |
| Linker admissions | 51.230–51.244 | 4–5,queries89–90 |
| vet admissions | 54.038–54.988 | 6–32,queries270–332 |
| Native receipt end | 66.574 | Root-wait/final reads already retained entries |

First-to-last admission UTC span5.134318s; vet admission span0.9495314s.
Last member query ends at54.9876923s,leaving11.5865399s until receipt end,
including cleanup. That is a new-member-query gap,not proof that retained final
reads stopped or that an identified process was missed.

There are32 distinct numeric recorded PIDs and31 distinct source-query numbers
among the32 entry rows,from332 total queries. Maximum Assigned in this saved
entry-source subset is3 (root included by the query implementation),not32.
This supports the cumulative-versus-concurrent distinction; it does not reconstruct
maximum concurrency across all332 queries or all unobserved instants.
All saved admissions are running; root-wait/final states are exited0. No interim
exit time is recorded,so entries cannot be timed as live for the whole retained
handle interval. The capped lifetime validator still stops before full entry checks.

## Why vet follows the target compile

The existing build arguments are go test -c -x -p=1 with no -vet override.
The frozen Go1.26.5 SDK's src/cmd/go/internal/test/test.go explains vet as part
of building test binaries. Its testC branch constructs a non-running print action,
then still calls addTestVet; addTestVet returns early only when testVet.off is set.
src/cmd/go/internal/work/action.go builds vet actions recursively for imports.
Thus compile-only is not equivalent to vet-disabled,and -p=1 does not make the
number of successive job members small. This source reasoning is tied to the
fixed copied SDK,not an assertion about every Go version or a new Go invocation.

The saved18795-line stderr contains225 exact copied-vet command records and225
matching configuration blocks,with225 distinct object directories/import paths.
All command records follow trigger line3321:first at3638,last at18792.
The first configuration is internal/goarch; the last is the nodeleafbinding
target at b127 with33 Go files. The command/config pairing uses only saved
object-directory tokens,not cached files or process lookups.

Counting method:recognize the observed literal TERM='dumb' prefix,then the exact
quoted copied vet.exe path with a token boundary,normalizing doubled backslashes.
Parse only matching saved cat/EOF vet.cfg blocks as JSON. No shell evaluation,
WORK expansion or actual trigger grammar modification. An initial executable-at-
first-token count yielded0 because TERM was the first token; inspection exposed
the prefix and the corrected literal count above. Zero was not reported as absence.

These225 command emissions/configurations do not prove225 successful launches
or distinct OS identities. They have no per-command UTC/PID binding to registry
rows. In particular,225 minus27 is not a justified missed-process count,and the
last target vet command cannot be assigned to a specific retained entry.

## Policy source and limitations

The unchanged [registry](../diagnostics/isolated-fuzz/retained-lifetime.cs) Poll
returns when capped (line82),skips seen IDs,and increments Attempts before opening
(line98). Attempt32 sets AdmissionBudgetExhausted even without a33rd candidate.
Successful opens/ownership checks admit one retained entry; snapshots do not
decrement Attempts,and disposal releases handles without reusing budget.
The [session](../diagnostics/isolated-fuzz/compile-stage-session.cs) creates this
same registry once after trigger activation. No image-based eligibility filter
was introduced. The [consumer](../diagnostics/isolated-fuzz/leaf-retained-compile.ps1)
rejects capped diagnostics independently of Incomplete=false or raw final exits.

Six current source pins were checked:four repository policy/session/native/
consumer files and the two copied SDK files. The copied SDK files match the
frozen copy receipt hash DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10;
this is not a repeat of the full11536-file SDK audit. All11 original manifest
artifacts were hash/length checked before analysis and at preservation.
No implementation or tests changed; no regression suite was rerun this increment.
Image names are optional saved metadata,not an identity authentication boundary.

## Next decision and recovery

Recommended next step:agree the diagnostic objective before another implementation
or real run,and consider a written design separating compile/link observations from
the large vet phase while preserving required vet acceptance elsewhere. This is a
direction for discussion,not approval to disable vet,filter by image,change the
budget,evict/reopen handles,shift the trigger or rerun. Raising the cap alone has
neither a justified new bound nor a guarantee of complete capture. Removing vet
also cannot be assumed to leave exactly5 future admissions:timing and action
selection can change. Do not weaken default/matrix/security acceptance to make
this diagnostic appear green. Any changed behavior needs design approval,fake
tests and separately approved live execution.

Ignored analysis root:.superpowers/diagnostics/2026-10-16-compile-stage-budget/analysis-001.
analysis.json SHA256:FA03DA597301757843EF7A230693EB580FF1398F0D31FD3414254C817F086D09.
Its single-file analysis-manifest.json SHA256:
E4B137FD22E725C5793B9C971D96068D4218C63C8B98BD9351E86113B5DDA650.
The manifest excludes itself; source fingerprints,32 entry rows,225 command/config
pairings and caveats are in the local JSON. The original stage manifest remains
7BD6D63F921C393E155B7BF4B4F319B4A01D2262A90496955FDD8EA9E12D2FF8.

Git preserves this portable explanation,[current status](current-status.md) and
[recovery instructions](../runbooks/repository-recovery.md),not raw data/SDK/cache/
binaries or this analysis JSON. On a fresh computer,read the development branch
record;do not replay either used compile001 identity. Eager/stage compile001 remain
STOPPED,matrix003 remains2PASS/1FAIL/23UNRUN,error87 unresolved; all wider acceptance
gates and main remain unchanged. No further native attempt is scheduled.
