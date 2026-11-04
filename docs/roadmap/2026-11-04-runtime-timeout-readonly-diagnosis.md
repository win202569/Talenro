# Runtime host timeout: read-only investigation

Status: ROOT_CAUSE_UNCONFIRMED / NO_EXECUTION / NOT_ACCEPTED.

User requested continuation after the proposed read-only diagnosis. No test,
driver, host, extracted source function, replay, profiling probe or code change
was executed. This report distinguishes static work counts from timing evidence.

## Observed boundary

- Host final: TimedOut/TerminationRequested/OwnedExit/StreamsComplete true,
  ExitCode -1, Exact/EvidenceValid/PostPinsIntact false. Driver exited1.
- Stdout28bytes contains only PASS: ContractAPIAvailable; stderr empty.
- capture/history-contract-runtime-host/journal.jsonl contains only begin.
  No intermediate stage timing, exception entry, tests-complete or published entry.
- LS-Result prints after a group completes. Absence of BaselineCore output does
  not distinguish context preparation, the first assertion, output delay, or a
  stalled error path. It does not establish that the second test started.
- The120000ms timer begins before stream creation and process startup, not after
  ContractAPIAvailable. The record does not timestamp that first result.

## Source trace and candidate hotspot

observation-contract.tests.ps1:95 prints availability. Lines99-119 then import
source-model, hash/read the manifest, and load three historical evidence files
and build three contexts. Line153 invokes the first BaselineCore assertion.
source-model.ps1:106 New-ODFixedContext always rebuilds the source model; its
Historical branch chooses the same14files for each of the three profiles.

Static normal-path work before the first assertion includes42AST parses of the
same14files (68900bytes total per set),99Get-ODPinnedSHA256 calls, each scanning
2344manifest rows:232056row predicate evaluations. Each pin call also hashes its
target. Repeated parsing/pin scans are a candidate hotspot, NOT a measured cause.
Filesystem/runtime scheduling and the first assertion remain unexcluded.

All three current historical evidence hashes match the suite's literal hashes:

- runtime-green:167842bytes,58B2D81047526EFB2BEE3FD8155DFAF493FC3FD6ECAAE32B0E7ED19C7DB66EA0
- lifecycle:125513bytes,3F7BDEE2B22F3C91EB5957259F18E1DF96E0CFD18A3F5E692D54C26460A51911
- continuation:212058bytes,D6729F1EBB54435B426013704B1AFD754E82F2EE7F6C725F520C02A56C532726

The older contract-green attempt reached FAIL: BaselineCore and recorded an
assertion error, not a timeout. This is not a successful timing control: it used
an older source/manifest and records no comparable phase durations.

## Evidence limit and proposed next decision

Existing records cannot locate the exact instruction or attribute120seconds to
any component. Do not optimize, cache checks, extend timeouts or retry based on
the static counts alone. A separately approved bounded diagnostic would need
flushed timestamps around manifest loading, each historical context, source
parse/pin work and first assertion, with an explicit new budget and namespace.
Such instrumentation/run is NOT authorized or performed by this report.

All stopped scopes remain closed. Counters26aux/5hosts/3compiles/217snapshots;
15stops retained. Existing completion-final archive remains the original evidence.
