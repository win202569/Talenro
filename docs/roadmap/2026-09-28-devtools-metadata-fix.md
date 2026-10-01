# Development-tool cache metadata repair

The user approved repairing the empty `Dir` diagnosis after checkpoint
`959f6b5d`, with all four production lockfiles and integrity checks unchanged.
This is separate from the completed consumer cache-binding repair in
`aed4b10d`. No push, merge, new dependency preparation or security-gate change
is included.

## Implementation and targeted evidence

For an empty all-module `Dir` only, the Windows verifier now queries the
already-selected exact `module@version` offline. It requires one record with
the same identity, then applies the existing expected-cache-path, physical
directory, zip and ziphash checks. The final `go mod verify`, 900-second shared
deadline, 4 MiB output cap and sanitized failures remain unchanged. The query
does not add project checksum entries.

A real Go 1.26.5 test prepares an uppercase-path module in a disposable local
proxy/cache, removes only its test-owned full-module checksum, and confirms
the all-module query omits `Dir` despite an existing physical directory.
The original verifier failed this case at dependencies (RED, exit 1,
5.472s package time). After repair, this case and the existing valid,
missing-cache, missing-ziphash, tampered-directory and tampered-zip cases passed
(GREEN, exit 0, 26.251s package time). Every real-cache case also checks that
its project go.mod and go.sum bytes remain unchanged.

Additional cases combine absent project checksums with missing cache,
missing ziphash, directory tampering and archive tampering. All five new
cases passed in the combined run, as did the full
`TestDevtoolsVerificationPowerShell` group (76.36s).

## Covering regression: not accepted

The original scoped selector ran once with default/unset CGO, offline module
settings, timing instrumentation disabled, and the unchanged `-timeout=10m`.
It exited 1 at 600.717s. The log is
`.superpowers/sdd/2026-09-25-c12-development-tool-isolation/metadata-fix-combined.log`.
There were 991 PASS markers, seven FAIL markers including parents of two leaf
failures, and one existing inapplicable Skip. The timeout means later tests
are not verified by this run.

The two leaf failures were under `TestDevtoolsShellEntryRejection`:

- `generate.sh/C:/Program_Files/Git/bin/bash.exe/outside-cwd` (23.22s).
- `verify-c11.sh/C:/Program_Files/Git/bin/bash.exe/outside-cwd` (16.46s).

Both reported `context deadline exceeded` at
`devtools_bash_rejection_test.go:129`: the five-second marker-control call,
before the production entry invocation. At the overall timeout, the active
`TestScriptCleanupExitStatusContracts/Bash/verify` fixture was also in marker
control; its stack was in Windows `syscall.CreateProcess`. This identifies a
process-launch boundary for further investigation, not a proven OS, Defender,
or workload root cause. A read-only resource query was denied; no security
settings were changed. No identical rerun or increased budget was attempted.

Version-lock, consumer binding, environment restoration, ownership and the
fake C11 PowerShell contract passed, but do not override the failed covering
run. Changes remain local and uncommitted. Real public-cache verification,
check-tools, generation and lint were not advanced after this failure.

## Follow-up launch diagnostics

On the user's continuation, the Bash fixture gained optional test-only
`DEVTOOLS_TEST_TIMING=1` observations around Start and Wait. Arguments,
environment, assertions, cancellation and deadlines are unchanged. No
production code was changed for this diagnosis.

Two targeted comparisons exercised `outside-cwd` for generate.sh and
verify-c11.sh: the Git launcher passed (1.546s package time), and direct Bash
passed (0.940s). Starts took about 4–13ms; marker controls finished within
0.6s. These are observations, not proof of the previous timeout's cause.

One complete instrumented selector then finished in 459.228s, exit 1,
under the original 10-minute budget. Its log is
`metadata-fix-launch-timing-combined.log` beside the earlier failed log.
1039 PASS markers, two FAIL markers (one leaf plus parent), one existing
inapplicable Skip. All 360 Bash launch observations had no deadline error;
the maximum Start duration was 75.6395ms and maximum total was 871.4621ms.
Neither prior Bash failure nor the global timeout reproduced.

The remaining failure was
`TestDevtoolsVerificationPowerShell/environment-restore-success` (3.74s):
expected exit 0, received 124, `integrity failed`. Source inspection confirms
this disposable fixture replaces the production 900-second deadline with
two seconds even for successful environment-restoration cases. This run
therefore reached that short fixture deadline, not the production deadline.
The underlying delay remains unexplained; no claim is made that environment
restoration, Defender, or a specific OS component caused it. All five new
metadata cases passed again. Do not treat a different failure as acceptance.

The next design decision is whether normal-success fixtures should retain
the same two-second budget used to inject timeout failures. No such budget
change has been applied. No commit, push, merge, real generation or lint was
performed. Existing failed logs are retained.

## Approved success-fixture budget separation

The user approved separating successful process fixtures from timeout
injection. Normal exit-0 cases now use a ten-second deadline in their owned
script copies; the existing natural-exit-with-child thirty-second exception
and independent fifteen-second outer process guard are unchanged. Timeout
and timeout-with-child still inject two seconds and require exit 124.
Other failure-case budgets and all behavioral assertions are unchanged.
Production remains at 900 seconds; this is not a production performance fix.

The previous observed environment-restore-success failure is retained as
the pre-change evidence. A targeted run without timing instrumentation
passed all five selected cases (19.310s package time): timeout, natural-exit,
timeout-with-child, environment-restore-success, environment-restore-failure.
Log: `success-budget-targeted.log`. Four production locks remain unchanged.
The covering non-instrumented run (`success-budget-combined.log`) completed
in 465.502s, exit 1, within the unchanged ten-minute budget: 1037 PASS markers,
four FAIL markers (one leaf and its parents), and one existing inapplicable
Skip. The PowerShell verifier group passed in 75.18s, including environment
restoration success (1.98s), timeout (3.15s), and timeout-with-child (3.11s).
Both timeout cases still required exit 124. All five metadata cases and the
version-lock checks passed.

The remaining leaf failure was
`TestDevtoolsShellEntryRejection/verify-c11.sh/C:/Program_Files/Git/bin/bash.exe/outside-cwd`
(6.94s). Unlike the earlier marker-control failures, this report came from
the actual entry invocation (`devtools_bash_rejection_test.go:154`), after
the marker control. Its unchanged five-second context expired; timing was
disabled. Whether launch or subsequent execution consumed the time remains
unknown. The success-budget adjustment has targeted and covering evidence,
but overall acceptance remains blocked. No Bash budget increase, identical
rerun, real-tool advancement, commit, push or merge followed.

## Bounded Bash follow-up

On the next continuation, the existing Start/Wait timing observation was
enabled on any context expiry or missing process state even when normal
timing output is disabled. This test-only diagnostic changes no assertions,
launch arguments, environment, five-second context, or production behavior.

A predefined ten-sample, fail-fast run of the previously failing launcher
outside-cwd case completed all ten samples, exit 0 / 4.639s package time
(`bash-outside-cwd-bounded-sample.log`). This was a bounded reproduction
experiment, not retry-until-pass acceptance. A separate single complete Bash
entry group, with normal timing disabled and fail-fast enabled, also passed:
exit 0 / 62.844s (`bash-group-failure-timing.log`), under a two-minute outer
limit. The original five-second per-call limits were retained.

No timeout reproduced in these focused checks. They do not establish why
earlier runs failed or replace the last red combined result. Future normal
runs now retain Start and total durations on failure without requiring
continuous diagnostic output. No further whole-suite replay, tool download,
security change, commit, push, merge or real-tool advancement was performed.

## Fresh combined verification

The user approved the next full diagnostic regression. With ordinary timing
disabled and failure-only observations retained, the original complete
selector passed: exit 0 / 471.162s, 1041 PASS markers, zero FAIL markers,
one existing inapplicable Skip, within the unchanged ten-minute budget.
Log: `failure-timing-combined.log`. The source is the local uncommitted
metadata repair, success-budget separation and failure timing above, based
on `959f6b5d7cefcb9ccbd20f6d0dbac39c381ffb7e`.

This is a fresh covering pass, not proof of the earlier intermittent delays'
root cause or long-term stability. Earlier failed runs remain part of the
record. Real-cache verification then ran under its original 900-second
budget on the same uncommitted source state. It started at
2026-09-28T14:36:18.2648978Z and ended at 14:51:18.4646897Z: wrapper exit
124 / 900.2064723s. Log: `task3-verify-after-metadata-fix.log`.

Read-only process inspection confirmed the child had reached `go mod verify`,
so the previous metadata-directory rejection did not recur. The child's
cumulative reads grew from 245,689,647 to 801,620,408 and then 1,442,650,918
bytes, with increasing CPU time. This demonstrates ongoing work, not an
integrity pass, a completion percentage or a diagnosed performance cause.
The outer wrapper exhausted its unchanged budget and terminated the owned
process tree; no public success output was returned. Do not misreport this
as a checksum mismatch or claim the public verifier itself returned 124.

The 52 protected baseline files and four production lock hashes were checked
unchanged. No check-tools, generation or lint followed; no immediate rerun,
budget increase, commit, push or merge was performed. Task 3 acceptance
remains blocked on real-cache verification; the fresh regression pass stands
separately from that blocker.

## Bounded real-cache performance diagnosis

After the user's next continuation, no full verification was repeated.
Read-only archive central-directory inventory over the 507 exact selected
modules in the saved dependency query completed in 12.122s. Totals:
372,671,074 compressed bytes, 1,183,054,709 content bytes, 199,115 entries.
`github.com/sqlc-dev/doubleclick@v1.0.0` accounts for 127,449 entries
(about 64%), with 68,726,688 content bytes. Its entry count is therefore a
specific candidate for expensive small-file access, not evidence of damage.
Log: `cache-inventory.json`.

The fixed Go source verifies both archive content and the extracted tree,
using at most runtime GOMAXPROCS concurrent modules. The inventory process
reported 16 processors; this is not a measured Go runtime concurrency value.
The public verifier clears inherited GO-prefixed settings, including
GOMAXPROCS. No production concurrency was changed.

A .NET SHA-256 comparison of the first 1000 fixed archive entries read
1,253,742 bytes: archive 0.051s, directory 0.260s, all sample contents equal.
This small, ordered, potentially warm sample is not representative of the
whole tree and does not constitute Go integrity acceptance (`cache-sample.log`).

A separate disposable Go probe used the already-cached x/mod v0.38.0 dirhash
implementation. Its hash.go SHA-256 exactly matches the fixed Go toolchain's
vendored implementation: `24DE6A735C1605E39B6381E6962787FAA5C59ADAB896FB66DBC1AC8B6AFCB297`.
For the same doubleclick module, HashZip completed in 2.123s and matched the
cached ziphash; HashDir did not complete within the probe's shared 90-second
execution budget. The probe exited 124; `go run` reported exit 1 with
`exit status 124`. Compilation precedes that execution budget. The original
unquoted modfile command failed before execution and its log was retained;
the corrected invocation is in `cache-hash-probe-quoted.log`.

A final, different 45-second probe ran only DirFiles (no content hashing):
127,449 files enumerated in 10.152s, exit 0 (`cache-walk-probe.log`). Taken
together, these observations narrow the performance investigation to the
extracted tree's per-file access/hash work, rather than archive hashing alone.
They do not measure the split within one HashDir run, prove cold-cache
performance, or establish a disk, antivirus, concurrency, or OS root cause.
No claim is made that this one module explains all 900 seconds.

All probes only read the shared cache, without executing dependency business
code or downloading anything. They are private diagnostics, not replacement
verification or product features. Full integrity acceptance remains open;
do not drop modules, omit extracted-tree hashing, weaken checks, or extend
production budgets based on these results.

## Per-file latency diagnosis

The next user-approved read-only probe selected 2048 evenly spaced entries
from the sorted 127,449-file doubleclick archive, rather than reusing only
the first 1000 entries. It used os.Open, io.Copy into SHA-256, and Close,
measuring open, copy excluding time in the hash writer, hash writer, and
close separately. Its execution had a 90-second timer; compilation preceded
that timer. Log: `cache-file-latency.log`; source is the ignored private
`cache-file-latency.go`. No dependency business code was executed.

The run exited 0 and processed 1,123,493 bytes in 20.737s:

| Measured component | Aggregate duration |
| --- | ---: |
| File open | 19.852s |
| Read/copy excluding measured hash writes | 0.470s |
| SHA-256 writes | 0.001528s |
| Close | 0.302s |

Per-file total latency was 10.160ms at p50, 23.435ms at p95, and 98.338ms
maximum. The maximum included an 88.911ms close, while most sampled time
was in open (about 96% of wall time). Timer granularity and instrumentation
overhead limit individual near-zero measurements. Read/copy includes wrapper
and buffering overhead; it is not a kernel-only I/O measurement.

This is direct evidence that file opening, not SHA-256 computation, dominates
this distributed sample. The Go implementation hashes files serially within
each module, even though modules can be checked concurrently. It makes the
large file count a credible contributor to the observed verification timeout.
Do not extrapolate an exact full-run duration from one sample, or attribute
the delay to Defender, storage hardware, a filter driver, or Go itself without
system-level evidence. This diagnostic does not clear full integrity acceptance.

The next useful investigation is tracing the file-open boundary, not repeating
whole-cache hashes or increasing budgets. No production verifier optimization,
cache deletion, exclusion, security-setting change, commit, push or merge was
made during this diagnosis.

## System tracing readiness (read-only)

The next continuation inspected installed tools and volume state without
starting a trace. Windows Performance Recorder 10.0.26100 is present and
reported no active WPR recording. Its installed profiles include FileIO,
DiskIO and Minifilter. The cache is on a local C: NTFS volume with about
83.4 GB free at inspection; this does not establish disk health or latency.

The C: volume lists bindflt, UCPD, LnvMSRIO, WdFilter, storqosflt, CldFlt,
bfs, Wof and FileInfo instances. Presence does not establish fault or delay
attribution. No driver was disabled, detached or reconfigured.

Further discrimination requires a short system-level I/O/filter trace.
Such capture can include unrelated process names and file paths, exceeding
the prior test-process-only measurements. Explicit user approval is sought
before starting it; proposed capture is approximately 30 seconds, local-only,
with no upload or Git tracking of raw traces. No recording was started in
this step, and the real verification result remains blocked.

## Approved short system trace

The user explicitly approved a local-only, approximately 30-second system
I/O/filter recording, including the possibility of unrelated process names
and paths. The existing WPR state was checked before recording. A unique
instance used the installed Minifilter profile, which already includes file
and disk I/O; no duplicate FileIO profile or security-setting change was
needed. The read-only sample executable was built before capture.

Capture began at 2026-09-28T15:18:25.7216053Z. The wrapper requested stop after
30 seconds; Windows stop/rundown/merge finished at 15:19:47.3031564Z. Thus the
81.582s wrapper duration is not a claim of an 81.582s intended sampling window.
An intervening read-only status query returned control-library busy during
the stop; it was not repeatedly polled. Final WPR status confirmed not recording.

Raw trace `TalenroFileIO-3a4e7bdc586c4060aede178f9199fe80.etl` is 963,641,344
bytes and remains in the Git-ignored local diagnostic directory. It was not
uploaded or staged. The probe exited 0. During this recording the same 2048
files took only 0.338s, with 0.138s measured in open, compared with the prior
20.737s / 19.852s measurement. The slow condition did not reproduce. Cache
warmth, scheduling and recording effects are possible confounders, not proven
causes; near-zero individual timing values reflect measurement limits.

This trace cannot directly establish the culprit behind the prior slow opens.
No driver attribution, security change, additional capture, full verification
rerun, commit, push or merge followed. A local trace-summary conversion was
requested with full event output discarded, under a 90-second processing cap.
That conversion reached its cap and the owned converter process was stopped
(wrapper exit 124); no usable summary was produced. Raw ETL remains intact.
Event-loss statistics and detailed driver attribution are therefore unverified.

## Bounded targeted trace read

Windows Get-WinEvent could read the first 200 events in about 2.49s. This
establishes basic readability only, not capture completeness or causal
attribution. Inspection was limited to event metadata, without exporting
unrelated event payloads or file paths.

A separate read filtered event-header ProcessID to the sampled process
(35860), capped at 2000 events and 90 seconds. It reached the time cap with
no usable result (wrapper exit 124); the wrapper stopped only its own reader
process. Header-PID filtering also does not establish coverage of all classic
kernel events. No further full conversion or capture was attempted.

Detailed trace analysis remains blocked with the currently attempted built-in
readers. WPAExporter and xperf were not found on PATH or at the checked Windows
Performance Toolkit paths; this is not an exhaustive software inventory.
Installing another analyzer requires a separate decision. Even successful
analysis of this recording cannot establish the earlier slow condition,
which did not reproduce during capture. The ignored local ETL is retained,
with no upload, staging, security-setting change, commit, push or merge.

## Approved performance-tool installation

After the user approved installation, the Microsoft ADK download page supplied
the September 2026 installer (10.1.26100.9457). Its Authenticode signature was
Valid and signed by Microsoft Corporation; SHA-256 was
`AC6A930FDB5C2980BA5FEFE606D47EDAAFCF5F647B4337411500D158EA77300F`.
Only `OptionId.WindowsPerformanceToolkit` was selected, with quiet mode,
no restart and CEIP off. Setup returned 0 and its log reported no restart.

The standard Windows Performance Toolkit directory now contains xperf
10.0.26100.9457 and WPA/WPAExporter 11.7.395.48728. xperf help worked;
WPAExporter help also returned 0 but emitted a missing
`wpaexporter.deps.json` warning. This establishes installation and help-command
execution, not successful trace analysis. No new capture, trace upload,
security-setting change or verification rerun occurred. Raw ETL remains local.

## xperf analysis of the existing recording

The installed xperf successfully read `tracestats -timespan` (exit 0).
The ETL header reports UTC 15:18:25.5149616 through 15:19:18.1199034,
52.6049418 seconds, with zero lost events and zero lost buffers. This is
header-level evidence, not proof that the sampled workload is represented.
It is distinct from both the requested 30-second interval and the 81.582s
capture/merge wrapper duration.

`minifilterdelay -apfp` completed in 33.483s with exit 0 and empty stderr,
producing 223 aggregate rows. No sampled-process name (including the `cache`
prefix) was found. A separate `process` action completed in 1.446s with exit 0
and empty stderr; its 440-line report had no match for PID 35860 or the
`cache-file` name. Neither report permits attribution of the sample's I/O
to a filter. Absence from these reports is not evidence of zero filter delay.
Why the sample is absent remains undetermined; header loss counters alone
do not establish workload coverage.

All reports remain in the ignored local diagnostic directory. No unrelated
process-level findings are included here. No new recording, upload, security
change or full-cache rerun was performed. The existing recording is not
sufficient to diagnose the earlier slow sample: it neither reproduced the
delay nor yielded a matching process in these analyses. Any further capture
needs an explicitly scoped plan and approval, not repeated conversion of
the same ETL. Integrity acceptance remains blocked.

## Proposed coverage-first recapture (not yet approved or executed)

Read-only inspection of installed WPR profile details confirms that
Minifilter.Verbose.Memory includes ProcessThread, Loader, FileIO and FilterIO
events. Missing process output is not explained by absence of those keywords.
Early-event retention in memory mode is a hypothesis, not an established cause.
Installed WPR help supports file mode and an explicit temporary-recording path.

Proposed scope is one capture, not an automatic retry loop:

- Prepare a uniquely named, short-name diagnostic executable before capture.
  Reuse the read-only 2048-file sample; record PID, UTC interval and timings.
  Keep that same process alive after the sample while WPR saves the trace.
- Check for existing recordings and adequate free space first. Use a unique
  WPR instance, the existing Minifilter profile, file mode and a private local
  temporary-recording directory. File-mode disk traffic can affect timing;
  this run primarily validates coverage, not an unbiased latency comparison.
- Request stop when sampling finishes or at 30 seconds, whichever comes first;
  allow rundown/merge to take additional time. Interrupt only the owned sample
  if it is still reading at the deadline; preserve partial-result status.
  Stop early if monitored temporary data exceeds 2 GiB or free space drops
  below 10 GiB. These are monitored triggers, not hard storage-size guarantees.
- Preserve the prior ETL. Keep new raw traces and reports Git-ignored and
  local-only; they may include unrelated system process names and file paths.
- Require matching PID/name/time plus attributable file-operation evidence,
  and check loss counters. A matching process alone does not prove I/O coverage.
  If coverage still fails, stop and report; do not start another capture.
- If coverage succeeds but slow opens do not recur, report only validated
  capture coverage. Do not infer a cause, relax verification, or change security.

This continuation only proposes the capture; no recorder was started and no
probe modification or execution occurred. Explicit approval is required before
implementation and collection. Existing acceptance blockers remain unchanged.

## Approved coverage recapture result

The user approved the preceding one-shot proposal. The standard-library-only
probe was compiled offline before recording. Its existing read-only sample
was retained, with an optional stdin wait after reporting results so the same
process remained alive through save. The wrapper passed PowerShell parsing.

The unique file-mode recording was
`TalenroCoverage-fcd726018c574c18a9dd4091c20a69d1`, with probe
`tca0a383cf.exe`, PID 16480. The probe started at UTC 15:41:19.9752847;
stop was requested at 15:41:20.9424225 because sampling finished, not because
of a deadline or space trigger. Save completed at 15:42:09.3954836 and WPR
confirmed not recording. Probe exit was 0; raw ETL size is 663,748,608 bytes.
The 2048-file / 1,123,493-byte sample took 0.403426s, including 158.314ms open,
170.108ms read/copy, 1.542ms measured hashing and 71.839ms close. The earlier
slow condition again did not reproduce. File-mode recording is a timing
confounder, not proof of improvement or a root-cause fix.

A combined header/process/minifilter analysis reached its 60-second cap
(wrapper 124) without output and only its owned analyzer was stopped.
Separate lightweight header and process queries then returned 0. The header
spans UTC 15:41:19.8022138–15:41:43.4438010 (23.6415872s), reporting zero lost
events and buffers. The process report explicitly contains tca0a383cf.exe
(16480), starting at relative 459891us and present through trace end. Thus
process presence is verified this time. Attributable file-operation evidence
has not yet been extracted, so the proposed full coverage criterion remains
unmet; do not equate process presence with verified I/O coverage.

No second capture or further full driver aggregation was attempted. New and
old ETLs remain local and ignored by Git. The four production lock hashes
remain unchanged. No security setting, production code, acceptance budget,
commit, push or merge changed during this capture. Any further work should
use this existing trace and target the sample interval, not repeat capture.

## Target-window I/O activity confirmed

A subsequent local-only `minifilterdelay -range 0 1500000 -apfp` analysis
of the new ETL completed in 23.043s, exit 0, with empty stderr. This window
covers the recorded process start and sample completion. The aggregate report
contains the uniquely named tca0a383cf.exe with filter totals (microseconds):
bindflt 30251, WdFilter 88849, storqosflt 22501, cldflt 28230, bfs 13520,
Wof 23989 and fileinfo 97322.

Together with the prior PID/name/time match, this establishes attributable
file-filter I/O activity for the probe, beyond process presence alone. It
does not establish a one-to-one accounting of every sampled file/read, nor
exclude missing events solely from header counters. These aggregate values
must not be summed as independent wall-clock costs or treated as causal
blame: the recorded sample was fast and did not reproduce the original delay.
The earlier analysis timeouts remain preserved; narrowing the requested
report succeeded in this run, without proving why prior analysis was slower.

The capture-coverage investigation now has positive process and I/O evidence.
No more capture or trace conversion is needed merely to show target activity.
The original cache-verification timeout remains unresolved and acceptance
remains blocked. No new recording, security change, dependency change,
integrity rerun, commit or push occurred in this continuation. The report
`coverage-filter-1500ms.txt` remains in the ignored local diagnostic directory.

## Full-directory progress diagnostic reproduced slow opens

The next continuation returned to the integrity timeout without a new system
capture or public-verifier rerun. A private, offline-built probe enumerated
doubleclick's full tree and used dirhash.Hash1 with timed open/read/close
wrappers and ten-second progress counters. The cached library hash.go and
the fixed Go toolchain's vendored hash.go have identical SHA-256
`24DE6A735C1605E39B6381E6962787FAA5C59ADAB896FB66DBC1AC8B6AFCB297`.
This is diagnostic instrumentation, not a substitute acceptance verifier.

The first measurement was invalid for read counts: embedding os.File promoted
WriteTo, bypassing the timed Read method. Its owned process was stopped and
its log retained as invalid read-count evidence. The corrected wrapper uses
a named file field; a self-test copied its own executable and asserted byte
and close counters before the next run. Preventing the fast-copy interface
and adding timing/atomic counters can change performance, so exact timings
must not be equated with the uninstrumented verifier.

The corrected run enumerated 127449 files in 11.231s. At its 90.002s deadline
it had opened and closed 7405 files and read 5,216,229 bytes. Completed-call
counters reported 75.421s in open, 1.633s in read and 1.307s in close. The
process exited 124 with no final directory digest or integrity conclusion.
Counters at the deadline are a concurrent snapshot and may exclude an
in-flight call. Slow opens therefore recur when traversing the full directory,
even though the earlier dispersed 2048-file sample was fast. Cache warmth,
access ordering and other environment effects remain hypotheses, not causes.
Do not extrapolate an exact full verification duration from this partial run.

Logs and diagnostic source remain ignored and local. Four production lock
hashes were unchanged. No shared-cache modification, security change, new
trace, full test suite, commit, push or acceptance clearance occurred.
Further OS attribution, if pursued, should target this demonstrated slow
full-directory workload, not repeat the fast dispersed sample; another
system-wide recording requires explicit approval.

## Approved full-directory trace: enumeration only

The user approved one approximately 30-second local-only trace of the
full-directory diagnostic. The probe was rebuilt offline with unique name
th086c97b6.exe, and its measurement self-test passed before recording. Progress
and enumeration messages use stderr; only a final hash result uses stdout,
so enumeration completion cannot prematurely stop the capture wrapper.

Instance `TalenroCoverage-c5d2da9e9ab34164a562a375aabdaeed` recorded PID 43044.
Probe start was UTC 15:54:21.4927152; the deadline triggered stop at
15:54:51.6333893. The owned probe was terminated unfinished (exit -1), and
save completed at 15:55:57.4289578. WPR confirmed not recording. The raw ETL
is 1,720,713,216 bytes, preserved locally and Git-ignored. No retry occurred.

Progress at 10, 20 and 30 seconds remained in enumeration; zero content
opens/reads had begun. Consequently this recording did not capture the
previously demonstrated slow content-open phase. Enumeration exceeded its
prior 11.231s measurement; recording overhead and environment changes are
confounders, not an established cause. No directory digest was produced.

Separate xperf header/process queries returned 0. Header duration is
53.4603449s with zero reported lost events/buffers; process lifetime was
185047–30298173us relative to trace start, matching th086c97b6.exe (43044).
A narrowly scoped filter aggregation for seconds 10–11 completed in 32.198s,
exit 0 with empty stderr. Target-process totals in microseconds were fileinfo
351595, Wof 85838, bindflt 74994, storqosflt 53985, WdFilter 45363, cldflt
62875 and bfs 40483. These confirm filter activity during enumeration only.
Do not sum them as independent costs or label the largest value the root cause.

Four production lock hashes remained unchanged; no shared-cache or security
settings were changed. No public verification, full tests, commit or push
followed. This approved trace is complete; original integrity acceptance
remains blocked. Another capture is not automatically authorized. If future
phase-specific capture is considered, workload readiness must be established
before recording so enumeration does not consume the content-read window.

## Phase-gated trace reproduced content-open delays

The user authorized continuing with enumeration before recording. A gated
probe (tgca27bcf9.exe, PID 33272) completed enumeration of 127449 files in
19.239s and waited for an explicit stdin signal. Only after checking its
ready message and recorder state did the wrapper start one file-mode trace
and release content hashing. Measurement self-test and wrapper parsing passed.

Instance `TalenroCoverage-46ad7b9fb2514fea8e8a080f406d59b3` started at UTC
15:59:58.7297552; hash-start was 15:59:58.7529135. The wrapper requested stop
at 16:00:28.9295613 and terminated its unfinished probe (exit -1). Save
completed at 16:01:26.6560834; WPR confirmed not recording. The ETL is
1,075,838,976 bytes and remains local and ignored by Git.

The last progress snapshot, approximately 30s into hashing, reported 2070
files closed, 1,809,536 bytes read, 28.345s in completed opens, 0.953s in read
and 0.464s in close. Thus the slow content-open workload was present during
this recording. There was no final digest or integrity pass. Instrumentation
and recording remain performance confounders.

xperf header/process queries returned 0. Header duration was 52.9052083s,
with zero reported lost buffers/events; the matching process was present from
before trace start to relative 30.412125s. Analysis of completion events in
seconds 10–11, with a minimum reported delay of 1000us, completed in 31.553s
(exit 0, empty stderr). Exact process-column filtering found 62 target rows,
all WdFilter.sys CREATE callbacks, with durations 1313–28315us and cumulative
971226us. Their calls/returns span 9.982509–11.014946s: boundary-straddling
calls mean the sum must not be described as 97.1% of a precise one-second
wall interval. The threshold excludes shorter events and these rows do not
prove that other filters had no cost.

This is direct evidence localizing long recorded open callbacks to the
WdFilter path for this workload, not proof of an intrinsic driver defect,
exclusive cause, or a reason to disable protection. Underlying scan/wait
mechanisms remain unverified. Existing trace analysis is now preferable to
further capture. Four lock hashes remained unchanged; no exclusions,
security changes, cache changes, public verification rerun, commit or push
occurred. Original acceptance blockers remain.

## 2026-09-30: OnOpen scan requests correlated with delayed opens

No new capture was taken. The existing phase-gated ETL was exported locally
only for seconds 10–11 and the registered Microsoft-Antimalware-AMFilter and
Microsoft-Antimalware-Engine providers. xperf returned 0 in 9.479s with empty
stderr; the 377903-byte export remains Git-ignored. Unrelated payloads are
not included in this report.

For the target process and doubleclick paths, the window contains 61 FileScan
events, 61 FileScanResult events, 61 StreamScanRequestTask starts and 61 stops.
Pairing engine start/stop by ActivityId and path yields 60 complete requests;
the window boundaries leave one leading stop and one trailing start unpaired.
All 60 pairs match an exact full file path (after the observed volume-path
normalization) and fit inside exactly one target WdFilter CREATE interval.
Request durations range from 913 to 26496us, average 14374.567us. Their sum
862474us is elapsed request duration, not independently measured CPU time.

For example, explain_397.txt has WdFilter CREATE at 10002176–10030491us
(28315us); AMFilter reports FileScan reason OnOpen at 10002358us; the engine
request runs 10003546–10030042us (26496us), followed by FileScanResult at
10030367us. Matching ActivityId, target process, full path and nested times
provide direct evidence that these delayed opens involve on-open scan requests,
not merely the presence of Defender on the machine.

This narrows the mechanism to opening-triggered scanning for the observed
workload. Request elapsed time still does not distinguish active scanning,
queueing, storage, cloud checks or other internal waits. No undocumented status
codes are interpreted as threat verdicts; no existing security finding is
cleared. The trace does not establish an intrinsic Defender defect or prove
that every module/time period has the same cause. Protection, caches, lockfiles,
production budgets and acceptance gates were not changed; no tests, commit,
push or external log upload occurred in this continuation.

## Defender performance analyzer corroboration

On 2026-09-30, installed Get-MpPerformanceReport successfully parsed the
existing phase-gated ETL (child exit 0). The request used explicit UTC start/end
bounds corresponding to trace seconds 10–11, TopScans 1000 and Raw output.
Its local-only CLIXML report contains 65 scans overall; filtering by exact
target PID 33272, image tgca27bcf9.exe and the doubleclick cache directory
selects 60. All 60 are classified by the official analyzer as RealTimeScan,
Reason OnOpen, SkipReason Not skipped. These fields are analyzer output,
not interpretations of undocumented event codes.

Installed help defines raw durations in 100ns units. Converted target scan
durations are 0.9138–26.4957ms, average 14.374502ms, sum 862.4701ms, consistent
with the earlier xperf event pairing at its coarser displayed precision.
The returned scan schema has classification, timestamps, duration and identity
fields, but no CPU/queue/storage/cloud-wait breakdown. Consequently actual
on-open real-time scanning is corroborated; its internal delay mechanism
remains unresolved. This is not a malware verdict or acceptance clearance.

Only existing local trace data was read. No New-MpPerformanceRecording,
active scan, protection/exclusion change, dependency operation, production
verification rerun, upload, commit or push occurred. Report and diagnostics
remain Git-ignored. Repeating the same capture is not the recommended next
step; further investigation needs a specific new observable or a separately
approved verification/remediation plan that preserves protection and gates.

## Proposed protected full-verification diagnostic (pending approval)

Read-only source review confirms two distinct 900s limits: the public
verify-devtools.ps1 shared deadline (metadata plus integrity), and the private
task3 wrapper's outer deadline. Increasing only the outer deadline cannot
extend the public script's inner integrity budget. Removing metadata checks,
trusting only the zip, omitting doubleclick testdata or accepting a partial
hash would change coverage and is not proposed.

Recommended next experiment is one separately labelled, uninstrumented
full `go mod verify` diagnostic in tools/devtools, with an explicitly approved
maximum of 3600s. This is a resource cap, not a predicted completion time or
a change to the public 900s contract. Use the same fixed Go executable,
offline/sanitized environment, selected module graph and shared cache. Before
launch, recheck all selected directory/archive/ziphash presence and exact
cache binding, because Go verify alone may skip wholly missing cached modules.
Abort if those preconditions fail; do not download or repair dependencies.

Record exact invocation, module count, lock hashes, elapsed time and owned
process liveness/CPU/I/O counters locally. Avoid ETW/profiling instrumentation
and concurrent verification jobs. Retain the 4MiB output cap; terminate only
the owned process tree at the deadline or user cancellation. Check all four
lock hashes again afterward. Do not change Defender, exclusions, shared-cache
contents, concurrency settings, or production/test budgets.

A zero exit with the exact success output proves only completion of this
diagnostic under its recorded conditions; it does not pass the public 900s
entry point, establish cold-cache performance, clear the existing threat
finding, or authorize downstream full tests. A mismatch/error or timeout
remains a failure/incomplete result, with no automatic retry or remediation.
Any production budget/design revision would need separate approval and tests.
This continuation only evaluated and documented the proposal: no extended
verification, process launch, code change, commit or push was performed.

## Approved extended integrity diagnostic passed (2026-09-30)

After explicit approval, a private diagnostic copy was mechanically derived
from the current public verifier. Changes were limited to its 3600s diagnostic
deadline, fixed source-root/helper paths, a selected-module milestone, and a
distinct diagnostic output label. The original public file was never written.
All module identity/cache-binding/directory/archive/ziphash checks, offline
environment, fixed Go 1.26.5 and PowerShell 7.6.5 requirements, process ownership,
4MiB command-output limit, and exact `all modules verified` comparison remained.
Source-shape assertions and generated-script parsing passed before execution.

The single run started at 2026-09-30T10:58:39.4809451Z and reached integrity
at 10:59:11.6321667Z after checking 507 selected modules. It completed at
11:33:54.3518125Z, wrapper duration 2114.9064061s, exit 0, with
`extended-integrity-diagnostic: passed`. This implies the retained exact Go
success-output check passed; the diagnostic does not print a substituted hash
or accept partial verification. No concurrent module-verification process was
found at launch. Owned Go CPU/read counters progressed throughout; no second
run or system-wide trace was started.

Four expected production lock hashes matched before/after. Public verifier
SHA-256 remained
`8991BE556E94328FFF11107E57F09DE5B7F21D1ADD0AC22AFB1DB3809FF578D1`.
The original 900s deadline remained in place. The wrapper and generated copy,
along with extended-integrity-2026-09-30.log, are ignored local diagnostics.

This establishes successful full-cache verification under this run's conditions
with protection unchanged and a larger diagnostic budget. It does not pass the
900s public entry point, prove cold-cache latency, clear the existing Defender
finding/I2/I3, or authorize downstream full tests. It also does not alone prove
that every part of the 35-minute run was caused by real-time scanning. No
production budget, security setting, cache contents, commit, push or merge was
changed. Any revision to production verification policy needs separate approval
and regression validation; do not repeat this run merely to warm the cache.

## Approved standalone deadline revision (2026-09-30)

The user approved a bounded revision of the standalone PowerShell verifier
from 900s to a fixed shared 3600s budget. Only this production deadline and
its explanatory comments change in this step. C11 outer stage deadlines,
Bash rejection-only entry, private diagnostic wrappers, security settings,
and the four dependency lock files are not widened or changed.

Two owned-copy behavioral fixtures age the initial deadline clock without
changing the production budget or command deadline checks: verification must
still succeed after a simulated 1800s elapsed, and must exit 124 after 3601s.
The red run failed the first case with the old 900s budget (exit 124 instead
of 0) and passed the expired-budget case. Existing timeout injection budgets
and independent short outer guards remain unchanged; only their PowerShell
source anchor is updated to the approved 3600s deadline.

Validation used pinned Go 1.26.5 and PowerShell 7.6.5 with offline readonly
module settings: `go test -v -tags=e2e ./internal/e2e -run
'^TestDevtoolsVerificationPowerShell$' -count=1 -timeout=4m`.
The group returned exit 1 in 100.130s. Both new budget cases passed, as did
the integrity/cache-corruption, timeout/child-cleanup, cancellation, output-cap
and environment-restoration cases. The inapplicable Bash launcher case skipped.
Two failures must remain explicit:

- `runtime-5.1-rejected`: `ChildCount:0 PositiveControl:false`.
- `ownership-boundary`: `process-event positive control missing`.

An isolated rerun selecting just those two cases reproduced both failures
(15.925s, exit 1, one-minute test cap). Source inspection places the failures
in CIM process-start observer positive controls. Legacy runtime rejection
occurs before the changed deadline; the ownership fixture loads the unchanged
helper, not the verifier. This narrows the failure boundary but does not
establish why the observer misses events. No test assertion, observer timeout,
system service or protection setting was altered. The entire regression group
is NOT green. The initial sandbox attempt also failed to access the Go build
cache; tests ran only after approved elevated execution.

Formatting check produced no filenames; `git diff --check` passed with only
the existing LF/CRLF notice. Four production lock hashes matched the pre-run
values. C11, check-tools, generate, process helper and Bash entry have no diff.

This is not acceptance of a new real full-cache run. No long verification,
full ordinary tests, C11/C12/bootstrap, commit, push or merge is performed.
The earlier diagnostic remains historical evidence under its own conditions.

## Process observer failure boundary (2026-09-30)

Read-only diagnosis used the ignored local `probe-process-observer.ps1`, which
creates only owned PowerShell children and temporary parent-PID-filtered event
subscriptions, then removes subscriptions and disposes/cleans up those children.
The initial probe mistakenly expected Register-CimIndicationEvent to emit a
subscriber object; it failed before starting a child. Diagnostic code was
corrected to inspect Get-EventSubscriber. No product/test code was changed.

Both PowerShell 7.6.5 and Windows PowerShell 5.1.26100.9444 registered a
subscription but observed zero process-start trace events at 3s and 10s;
control children exited 0 and Winmgmt was Running. Follow-up checks found:

- Effective Windows token is administrator (the documented prerequisite).
- Local New-Event/Get-Event queue control returns exactly one event.
- CIM process lookup confirms the owned child's actual parent matches the
  PID in the trace query; a longer-lived child still produces no trace event.
- With two simultaneous parent-filtered subscriptions, the intrinsic
  `__InstanceCreationEvent WITHIN 1` control observes the exact owned child
  once, while `Win32_ProcessStartTrace` observes zero events through 10s.

This isolates a failure of the process-start trace event path in this
environment, not general PowerShell event queuing, general WMI delivery, the
parent filter identity, or the 3600s verifier deadline. It does not yet identify
the provider/trace subsystem root cause. Recent WMI operation log entries for
both diagnostic subscriptions carried 0x80041032 at teardown, including the
working intrinsic control; they do not by themselves explain the asymmetry.

Microsoft's documentation specifies administrator PowerShell for this trace
example: https://learn.microsoft.com/en-us/powershell/module/cimcmdlets/register-cimindicationevent

The intrinsic subscription is only a diagnostic comparison, not an acceptable
silent replacement for trace-based proof that no short-lived child launched.
Production assertions and 3s observation budgets remain unchanged. No service
restart, event-provider repair, trace-session stop, security change, full test
run, commit or push was performed. Further work should inspect the specific
trace-provider/session state before proposing any system-level intervention.

## Provider/session checks and bounded recovery handoff (2026-09-30)

Read-only inspection confirms exact COM registration for
{9877D8A7-FDA1-43F9-AEEA-F90747EA66B0} resolves to
Windows System32/wbem/krnlprov.dll, with a Valid Authenticode signature.
The WMI event registration lists process/thread start/stop and module load
queries; operational events report the provider started with result 0x0.
An initial registry-provider path lookup failed due to path spelling; a direct
Registry.ClassesRoot lookup confirmed the entry exists. Do not interpret that
initial lookup or formatting artifacts as missing registration or invalid signing.

The final short probe queried active ETW sessions without recording or changing
them. Admin_PS_Provider was running with Microsoft-Windows-Kernel-Process,
keywords 0x70 (process/thread/image), real-time mode, and zero logman buffers
lost. Its xperf snapshot reported 2 events lost; this alone does not attribute
the missing specific child event. The trace subscription again delivered zero
events through 10s; intrinsic comparison delivered the matching owned child.
The additional intrinsic event was expected because xperf was itself launched
under the probe's parent-PID filter. All probe-owned processes/subscriptions
were cleaned up. Existing system trace sessions were not stopped or changed.

No repository fix follows from these observations. To avoid further unchanged
sampling, the proposed next recovery boundary is a user-controlled Windows
restart followed only by runtime-5.1-rejected and ownership-boundary tests.
Last boot is 2026-09-17; uptime is context, not proof of cause. A restart is a
recovery experiment, not a guaranteed fix, and has not been executed. No WMI
rebuild, re-registration, provider-host termination, trace-session stop or
security change is authorized or performed. Preserve the regression/security
blocks until fresh evidence passes the required checks.

## Post-restart targeted verification passed (2026-10-01)

After the user reported a Windows restart, the boot date was confirmed and
only the two previously failing observer cases were rerun, with fixed Go
1.26.5 and PowerShell 7.6.5, offline readonly module settings:
`go test -v -tags=e2e ./internal/e2e -run
'^TestDevtoolsVerificationPowerShell$/^(runtime-5.1-rejected|ownership-boundary)$'
-count=1 -timeout=1m`.

Both passed: runtime-5.1-rejected 5.08s, ownership-boundary 8.03s;
package duration 15.337s, exit 0. Build/startup wall time preceded the reported
test duration. No observer code, assertion, timeout or security setting changed.
Restart cleared the observed failure in this run; it does not prove the
underlying cause, permanent recovery or complete regression acceptance.
No long cache verification, full ordinary tests, C11/C12/bootstrap, commit or
push was performed. The next scoped verification is the short verifier group;
all remaining security and outer-budget restrictions continue to apply.

## Post-restart complete verifier regression passed (2026-10-01)

Following user authorization to continue to the next step, ran the entire
scoped verifier group using the pinned offline toolchain:
`go test -v -tags=e2e ./internal/e2e -run
'^TestDevtoolsVerificationPowerShell$' -count=1 -timeout=4m`.
The group passed in 104.05s, package duration 104.482s, exit 0. All applicable
runtime, ownership, cache/integrity, aged-clock, timeout/descendant cleanup,
cancellation, output-cap and environment-restoration cases passed. Only
launcher-cancel skipped because it applies to the Git-for-Windows launcher,
not this PowerShell entry. No product or test code changed for this run.

Four production lock hashes still match the checkpoint. gofmt reports no
unformatted test file; git diff --check passes with the existing LF/CRLF notice.
This establishes a fresh green scoped verifier regression, not a full project
suite result, cold-cache performance, long shared-cache verification or C11/C12
acceptance. The shorter C11 outer budgets and all security restrictions remain.
No long verification, full ordinary tests, commit, push or merge was performed.

## Approved C11 outer-budget alignment (2026-10-01)

After explicit approval of the 61/70/70-minute design, changed only the three
corresponding stage limits in verify-c11.ps1: devtools verification 3660s,
check-tools 4200s and generate 4200s. The latter two are whole-stage caps,
including their independent repeated 3600s verification and subsequent work;
they are not an additional 70 minutes after verification. These are resource
ceilings, not guarantees about cold-cache completion. Existing ValidateRange
maximum 4200, all other stage limits, failure propagation and cleanup are
unchanged. No verification is removed or reused to skip a later check.

New TestDevtoolsC11StageBudgets executes the real first four stage invocations
and Invoke-C11Stage forwarding function, replacing only the external execution
boundary with a recorder. It checks the three approved timeout arguments and
the unchanged 600s unit-test stage default without launching actual C11 tools.
An initial harness run lacked PSScriptRoot context; after correcting the owned
test harness, the red run reported the intended 900/900/600 versus
3660/4200/4200 mismatches (3.649s, exit 1). Production limits were then changed.

Fresh offline fixed-toolchain run:
`go test -v -tags=e2e ./internal/e2e -run
'^(TestDevtoolsC11StageBudgets|TestVerifyC11PowerShellContract|TestDevtoolsFailureStopsConsumers|TestDevtoolsSameHostDispatch)$'
-count=1 -timeout=4m` passed, package duration 91.175s, exit 0. All four selected
groups passed; no skipped cases were reported. Existing failure injection,
same-host routing and fake-tool C11 stage-order/error contracts remain green.
gofmt reports no unformatted new test file; git diff --check passes with only
LF/CRLF notices. Four production lock hashes remain unchanged.

This step does not execute real full C11/C12, ordinary full tests, long cache
verification or generated production tools. No security setting, private
diagnostic wrapper budget, dependency lock, commit, push or merge changed.
The earlier outer-budget blocker is superseded in source; real acceptance
and security review remain outstanding. Do not use historical private 900s
wrappers as evidence that the revised public limits were exercised.

## Independent checkpoint review (2026-10-01)

A read-only reviewer inspected the working-tree changes against
959f6b5d7cefcb9ccbd20f6d0dbac39c381ffb7e, including the new stage-budget test
and this evidence ledger. No blocking production defect was identified.
The reviewer assessed this as suitable for an unaccepted development checkpoint,
not as approval to merge or clear security gates. No tests were independently
rerun by the reviewer; recorded execution evidence remains the preceding runs.

Nonblocking follow-ups retained: directly test malformed/multiple-record,
identity/version-mismatched and wrong-cache-path exact-version metadata replies;
strengthen exact 3600s boundary behavior (current aging cases test 1800s success
and 3601s expiry). These are coverage observations, not demonstrated defects.
Prefer behavioral coverage rather than a source-literal-only assertion.

Clarified the historical section of current-status and refreshed the older
cache-binding evidence cross-reference. Review did not judge actual long-cache
completion, full C11/C12/ordinary/bootstrap acceptance, security finding/I2/I3,
or reboot root cause; these exclusions remain open, not silently waived.
Four lock hashes still match. Local probe/log paths remain Git-ignored.
No staging, commit, push or merge occurred.

## Safety gates retained

The four production lock SHA-256 values match the prior checkpoint. No shared
cache content is intentionally corrupted or removed; corruption is restricted
to test-owned temporary caches. Full ordinary tests, real full C11/C12 and
bootstrap, 13 physical gates, Defender review and I2/I3 remain blocked.
