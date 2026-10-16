# Compile-stage-triggered retained diagnostics — proposed written design

Date:2026-10-16. Base:5b6f6d1084b2bff076e71aa1c9eab0fd6f15a74f.
Status:the user approved this written spec after54dc50188c23a885fd5f1ba4a7d2ba90b17f0bcb.
The [implementation plan](../plans/2026-10-16-compile-stage-retained.md) is pending
user review and execution-method selection. No code changes, fake-test execution,
real compile or acceptance is authorized by this design approval alone.

## 1. Intent and non-goals

Improve the diagnostic usefulness of the existing bounded retained collector for
one fixed nodeleafbinding compile. Start new-member sampling only after observing
the target package's test-inclusive compiler command in this run's stderr log.
Keep32 cumulative admission attempts, exact owned handles,600s native deadline,
500ms natural/shared5s cleanup and unconditional ordinary diagnostic refusal.

The [saved-only cap analysis](../../roadmap/2026-10-16-leaf-retained-cap-analysis.md)
shows compile001 exhausted32 admissions at23.42s of a74.94s receipt. Root-wait
had no original core members; early retained samples do not explain the previous
OpenProcess87. This design changes the sampling window, not that historical
result. It does not promise to capture a short-lived compiler, a root-wait member,
every descendant, or a particular shutdown race. Retention and log polling can
change timing/object lifetime; every later live report must disclose that.

No SDK/product/test callback patching, warm-cache preparation/retry, runtime of
the generated binary, matrix enablement, extra privileges, installation, network
fallback, handle eviction/reopening or interpretation of87 as success is included.

## 2. Alternatives and choice

- Selected:fixed target-package compiler-log trigger. It is workload-stage-based
  rather than calibrated to one machine's elapsed time. Logs may be delayed,
  incomplete or absent on cache hits; missing evidence must stop the observation.
- Not selected:fixed elapsed delay. Existing74.94s timing does not predict another
  build and cannot supply a reliable end-phase boundary.
- Existing controlled-child fixture remains a complementary test of lifecycle
  mechanics, not a substitute for real compilation or a reproduction of error87.
- Not selected:larger cap or rotating samples. Those alter the agreed resource/
  ownership policy without guaranteeing evidence at the problematic boundary.

This is a target-package-stage sample, NOT an asserted shutdown-phase sample.
Even after activation, more than32 new identities may appear. Capped outcomes
stay incomplete, including exactly32 successful admissions with no33rd candidate.

## 3. Component boundaries and compatibility

Add a small compile-stage-trigger.cs unit for bounded incremental log parsing,
clocked trigger state and report DTOs, with fake byte-source/clock seams. It owns
only its reader and parser buffers, never process handles or a background thread.
Production reads the existing native-created stderr file through one read-only
handle; no shell, process enumeration or execution of text from that file.

Add explicit RunWithCompileStageRetainedDiagnostics beside the existing native
entries, with the same executable/args/environment/cwd/streams/limit inputs.
It selects a new internal mode; existing public signatures and default behavior
remain unchanged. Its policy is fixed to the leaf package. Derive the compiler
path from the copied go.exe toolchain root and the source directory from cwd;
require the fixed controller's validated input bindings. No ambient environment
flag, arbitrary regex, caller-selected delay or externally supplied trigger file.

This new mode must not combine with legacy live sampling/shutdown diagnostics.
Both LifetimeDiagnosticOnly and ShutdownDiagnosticOnly remain true; shutdown
schema3/policy are unchanged. An optional CompileStageTrigger report distinguishes
this mode. run.ps1 preserves that optional object, without inventing it for old
receipts. Assert-ProcessCompletion continues refusing diagnostics before success
checks; it must additionally refuse a present CompileStageTrigger even if its
diagnostic flags are malformed/cleared. Existing ordinary receipt behavior is
unchanged. Missing trigger metadata cannot be accepted by the new controller.

Add an independent leaf-stage-retained-compile.ps1 controller/tests, not a changed
meaning for leaf-retained-compile.ps1 or reused compile001 identity. Its fresh
owned root basename is leaf-stage-retained-compile-NNN under the existing scope.
It reuses fixed compile-only arguments/offline environment, SDK reference checks,
source/product/artifact audits, bounded manifests and stopped/completed protocol.
Extend required source pins to the new helper/native parser and serializer.

## 4. Fixed trigger grammar and provenance

Read only this invocation's fresh stderr stream from byte offset0, incrementally.
Never scan historical logs, stdout, a caller-selected path or compiler output
from a prior run. Import/planning and parser construction do not launch anything.

An eligible complete LF-terminated command line must tokenize as follows:

1. First token equals this copied SDK's pkg/tool/windows_amd64/compile.exe path.
2. Exactly one separate -p argument has the exact following token
   talenro.local/platform/internal/nodecontrol/nodeleafbinding.
3. Exactly one standalone -pack token begins the source suffix. Every subsequent
   token is an absolute .go path with immediate parent equal to
   cwd/internal/nodecontrol/nodeleafbinding; at least one ends in _test.go.
   An empty suffix, a later flag or a non-source token is not eligible.

Use a narrow parser for the observed Go -x format:ASCII whitespace separates
tokens; double quotes group tokens and support escaped backslash/quote. Reject
unclosed quotes, unsupported escapes and invalid UTF-8 in a candidate line;
do not use shell parsing/evaluation. Normalize path separators and compare fixed
Windows paths case-insensitively without expanding variables or resolving links
from text; package comparison is ordinal/case-sensitive. Substring matches,
vet JSON, link commands, package-prefix collisions, option values containing
test filenames and same-basename compilers outside the pinned SDK do not trigger.
Ignore complete non-command lines; malformed lines that purport to start the
expected compiler command fail the diagnostic rather than accepting a partial
match. Nonmatching commands never allocate registry handles.

Activate once on the first eligible command; subsequent text is not rescanned.
Retain raw matching-line bytes as base64 bounded as below, byte start/end offsets (exclusive
end including LF),1-based line number, raw line SHA256, and SHA256 of the complete
prefix through that line. Parsing may strip CR before LF, but hashes cover raw
bytes. Record expected compiler/package/source directory in the saved plan.
Final audit must verify the same prefix and matching line against saved stderr,
not merely find a similar later command. Log text is evidence of emission only,
not authenticated OS process identity or proof of execution order.

## 5. Bounded reading and state machine

Fixed limits:16MiB total bytes scanned,64KiB per runner iteration,64KiB per raw
line including terminator, and at most60000 reader polls. Buffer incomplete lines
only up to the line limit. Bound both parser storage and work per iteration; no
whole-file ReadAllText/rescanning from0. A limit reached without a trigger fails
the diagnostic; do not silently truncate, increase limits or restart the scan.

Open the runner-created file once after stream creation using read access and
sharing compatible with the existing writer, without delete sharing. The
controller retains its current fresh-path/reparse/single-link checks. Detect
length regression below already consumed offset, read errors and invalid clock
ordering; latch a diagnostic error. Do not assert that sharing/hash checks provide
a general filesystem security boundary or detect every hostile concurrent edit.
Final prefix/hash binding must detect changed consumed bytes that affect evidence.

States:Waiting -> Activated OR Failed OR RootExitedWithoutActivation. Initial
state has no registry and no new-member queries/admissions. In the explicit new
mode's existing runtime loop, read a bounded chunk, examine root using its exact
creation handle with zero wait, then activate only if a candidate is complete,
the root is not signaled at that check, and the original runtime deadline remains.
Capture that root-check result and timestamps. This ordering does not guarantee
the root remains alive between check, activation and first poll.

On activation record monotonic/UTC observation interval, then construct exactly
one registry and call its unchanged bounded Poll in that iteration. Never perform
catch-up queries for the waiting period. Continue normal registry polling until
the existing cap/error/deadline/root exit conditions. Stop reading/scanning after
activation. Root-wait and post-cleanup snapshots and finally releases remain on
the original shared cleanup clock; neither trigger nor cap grants more time.

A reader/parser failure disables further triggering/sampling but does not skip
root supervision or ordinary cleanup; retain its error and let the one bounded
native invocation finish. No throw may bypass reader disposal or handle cleanup.
Waiting at root exit becomes RootExitedWithoutActivation; a matching buffered
line processed after that exit cannot retroactively activate. No fallback to
eager collection, elapsed delay or replay. OS calls are not preempted by these
software limits, just as with the existing native deadline.

## 6. Evidence and interpretation

CompileStageTrigger schema1 stores policy name leaf-test-compile-log-v1, final
state, fixed limits/consumed bytes/poll count, matching-line/prefix bindings when
activated, clock frequency and observation intervals, exact-root zero-wait
result, activation time, first registry query timing, and a bounded list of
errors (at most16 entries, each at most512 characters). No native handles or
mutable parser/API internals are serializable. All metadata has explicit absent
values before activation; zero or omitted timestamps are not invented evidence.

The new observer requires complete trigger provenance, root-not-signaled check,
ordered activation/first-query/receipt clocks and final prefix binding, AND the
existing independent core and retained-lifetime validators. At least one admitted
entry, fewer than32 attempts, unambiguous same-owned-handle reads, release evidence,
natural core completion and unchanged source/SDK/product/artifact audits remain
required. Registry data cannot fill missing original core rows or convert log
paths/PIDs into ownership evidence. Image paths remain optional metadata.

Keep CompileExitZero, CompileEvidenceComplete, TriggerEvidenceComplete,
CoreNaturalComplete, LifetimeEvidenceComplete and ObservationPassed separate.
Acceptance/PackagePassed/OriginalTestPassed/RuntimeAttempted always false, even
for a complete diagnostic sample. Missing trigger, partial line, scanner failure,
root already signaled, zero admissions,32 attempts, ambiguity,87 or cleanup/
release/audit/persistence failure saves STOPPED, never completed.json.
Do not clear flags or alter a saved receipt to reuse an ordinary success guard.

Persist plan/attempt/receipt/streams/product inventories/artifact binding/result,
plus explicit bounded manifest and stopped/completed protocol. result.json alone
is provisional. A completed diagnostic requires all final audits and a completion
marker binding the exact manifest; never recursively inventory cache/tmp. The
manifest includes raw trigger evidence within the saved receipt, not new copies
of SDK/cache files. Independent final audits still run after an earlier failure.

## 7. Required fake-only verification before later live consideration

Use portable synthetic logs and fake clock/byte-source/native boundaries. No Go,
native Run, process/PID lookup or full SDK audit in the implementation test phase.
Exercise production parser/state/controller serialization and persistence code.

- Fragment eligible input across reads/UTF-8 boundaries; complete line triggers
  once, partial line never does; CRLF, quoted spaces/backslashes, no test source,
  wrong package/prefix/compiler directory, vet/link lines and repeated tokens.
- Require exact eligible standalone source tokens, not text inside another
  argument. Include fixture text shaped like current saved command without local
  absolute paths, ignored receipts or machine caches as dependencies.
- Test pre-trigger0 registry queries/opens; activation -> first query ordering;
  triggered registry remains unchanged; no reset/eviction/extra32-attempt cohort.
- Test byte/line/poll limits, line boundary exactly at limit, malformed candidate,
  short reads/read exception, file length regression, prefix alteration, backwards
  clock, deadline exhausted and reader disposal on every terminal/exception path.
- Root exits before line completion or between candidate read/root check;
  root exits immediately after activation; zero admissions/cap still incomplete.
- Same-handle lifecycle/release failures, original core missing evidence, compile
  nonzero, copied compiler-token failure, product/SDK/source/artifact drift,
  multiple final-audit failures and manifest/result persistence failures.
- Populated optional-report roundtrip including raw-line bytes/hashes and nulls;
  missing/corrupted trigger clocks/offsets/hash refusal, ordinary guard refusal
  with both flags true and with trigger metadata present but flags cleared.
- Run relevant legacy fake suites to prove eager retained/default/matrix paths
  unchanged. Separate static checks from executable assertions and publish exact
  suite/host counts only after observed results, not historical totals.

Use TDD with named RED/GREEN failures and independent review after an approved
implementation plan. Tests do not prove real file visibility/capture latency,
compiler format compatibility, short-lived process capture or error87's cause.
Any later real attempt needs separately approved frozen SDK/reference/module
bindings and a new identity; it must compile once only and never run the binary.

## 8. Review, preservation and next gate

Spec self-review must check placeholders, scope, ordering, concrete resource
bounds, compatibility and fail-closed semantics. This document does not declare
implementation verified or alter historical001/002/matrix outcomes. C11/C12/
Task10/B01/production/main gates remain open.

Save this spec with progress/recovery links on the existing development branch;
do not include raw logs/receipts/SDK/cache/binaries. Git restores the proposed
design and source/tests, not the previous computer's environment or live state.
Written-spec review is complete; the next gate is implementation-plan review
and execution-method choice. No implementation or live rerun has started.
