# Later ownership frontier — read-only sequential-isolation audit

After93761c95,user approved read-only review of later sequential-isolation
source/workload and saved timeout versus completion evidence.
ZERO product/test/controller/source edits,probe preparation,offline-suite/target
launches,retries,warmups or frozen-slot releases. No independent reviewer.

## Observed termination versus unproven semantic failure

Latest three-phase target was killed at its configured30000ms controller limit,
wall30081ms including cleanup;TimedOut=true,exit-1,empty stderr.
Raw is the valid101-line inventory PREFIX:95 lines/59 PASS,last
P1I1PassiveDescriptors. No assertion/fixture error recorded.
Next expected SequentialIsolation PASS is absent. No case-start/API-start marker
exists for it,so neither entry,progress,active instruction nor elapsed time of
that case is established. This is a failed full-suite timeout verdict,not an
observed SequentialIsolation assertion failure or proof of a permanent stall.

P1I1-Case(:5) executes the body then writes PASS. Last PASS confirms that
previous body returned,but does not time the gap,next dispatch or next case.
Empty prior OpenScopes means all instrumented4096 regions closed,not that the
whole suite or later tests completed.

## Source loading and static workload

Source:
integer-candidate-three-phase/offline-supervision-policy.ownership.tests.ps1,
SHA C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7.

Runner:186 invokes original35 child tests under Group all,then:187 dot-sources
the ordinary adjacent candidate module before ownership regression. Ownership
calls name Get-OfflineSupervisionDecision,not the phase clone. Source hashes
match original ownership tests and both earlier candidate diagnostic copies.
This is static command/loading intent,not dynamic binding/cache attestation.

SequentialIsolation(:225–233) loops exactly three literal cases:
ExitCode at observation2,NaturalCompletion at3,Classification at4.
For each case it:
- Builds a fresh five-observation normal scenario and decorates the selected
  primitive with owned mutable Metadata through the guarded fixture constructor.
- Evaluates the same decorated input twice,checking full refusal results and
  distinct result references.
- Mutates the input metadata array and checks both existing refusal results.
- Mutates only result a's Reason,checks result b remains unchanged,then evaluates
  a fresh clean five-observation normal scenario and checks its full result.

On successful completion that is3 API call sites times3 cases =9 API calls
(six decorated refusals,three clean controls),six top-level equality call sites
times3 =18 full comparisons,plus three explicit a/b reference checks. Counts
exclude recursive equality calls and fixture helper work. They are conditional
source arithmetic,not observed counts from the timeout. Each input declares
five observations,not4096;do not equate declared input size with rows actually
admitted before a decorated refusal.

The body contains no explicit file/network/process/wait call or unbounded
source loop;fixture helpers use SDK reflection/PSObject construction and checks,
and full equality is depth-bounded. This does NOT prove runtime engine calls
cannot be slow or stall. There is no demonstrated external dependency blocker.

PlainControls(:234–240) follows:five API calls/five top-level full comparisons
for normal,long-valued control,empty prepared,unknown-exit and sticky-failure
fixtures. Their declared observation lengths are5,5,0,3,5. Complete is a final
trivial assertion. These semantic/control/ownership checks must not be deleted,
cached,skipped or weakened to fit the deadline.

## Existing results, compared without replay

| Saved candidate diagnostic | Wall ms | Outcome | Sequential / Plain PASS saved |
|---|---:|---|---|
| Outer boundary markers |28437|exit0,complete text via independent audit;old host-classifier defect retained|yes / yes|
| API/equality split |28208|exit0,complete-diagnostic|yes / yes|
| Three-phase clone |30081|timeout/exit-1,valid-prefix only|no / no|

Checked saved stream hashes and literal completion/error lines for all three.
No saved ASSERTION FAILED/FIXTURE ERROR/ERROR output in their inspected streams.
Earlier two PASSes supply real prior coverage of these unchanged vectors,not
current acceptance,stable runtime or proof the current case would pass eventually.

Module/body/loading/history/instrumentation and ambient runtime conditions
differ. Earlier28-second complete runs and this30-second timeout support
investigating end-to-end timing margin,but do NOT prove that cumulative work
alone explains it or identify a new product defect. Do not relabel failure as
a harmless timeout,raise the cap or merge main.

Latest4096 ValidationCopy16881.149ms is a completed measured combined phase,
not a measured duration of ownership tests. Timestamps are absolute monotonic
ticks;no saved same-clock parent launch baseline or later-case timing permits
reconstructing remaining budget at SequentialIsolation entry. Do not infer
per-case time from deadline,last PASS position,CPU total or missing output.

## Verification and next recommendation

AST parse/command extraction found three SequentialIsolation API sites/six
equality sites,five PlainControls API/equality sites,with finite literal fixture
loop. Source hashes equal across original/outer/split/three-phase copies.
All388 prior pins and latest index
02A8907BE3E1D12B3013EEF77ECABE4B6BAE314D0EBF034633D4EF07526A56D9
held. Saved stream hashes/tails compared without importing/running target code.
No root cause or semantic repair established.

Next recommend a bounded DESIGN for sparse group-level start/end observations
on SequentialIsolation and PlainControls in a new isolated copy,keeping full
suite order,existing APIs/fixtures/assertions and common case helper unchanged.
Goal:distinguish a real open case from timeout before entry/after return,and
measure only completed cases. Explicitly address script-body/history/output
perturbation and missing-end/early-error interpretation BEFORE preparation.
Do not start another full run or continue finer4096 timing automatically.

Three-phase1/1,split1/1,boundary1/1 spent;candidate-full1/2 spent/PS5 FROZEN;
all older allowances/review0/1 unused/ineligible remain. Candidate/I1/P1/main/
C12/security/runtime/release gates unresolved. Save analysis/status/recovery on
development branch. Git restores neither ignored raw,cached engines,credentials
nor renewed test/replay authority.
