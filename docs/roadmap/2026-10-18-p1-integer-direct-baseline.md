# Direct integer-helper baseline — PS7 and PS5 PASS

Afterf33920a4, user approved the recommended NEW independent two-host baseline:
ordinary fresh PS7 then PS5, prepared standalone suite once each,30000ms per
host, stop on any failure/timeout, no retry/warmup/product edit/old-slot reuse.
Both authorized hosts completed;new allowance2/2 spent. No other test launched.

## Observed results

| Engine | Started UTC,2026-10-18 | Wall ms | All-thread CPU ms | Exit | Lines / PASS |
|---|---|---:|---:|---:|---:|
| PS7 7.6.5 Core x64 | 09:03:01.0039500Z | 1047 | 765.625 | 0 | 17 / 14 |
| PS5 5.1.26100.9711 Desktop x64 | 09:03:06.8432474Z | 856 | 921.875 | 0 | 17 / 14 |

Both:parent/child admin=False,no timeout,empty stderr,owned-host-exited,
CaptureValid=true,ExpectedResult=true,FormalProjectAcceptance=false.
Full exact ordered output:CONTEXT,HOST,14 named PASS markers through
P1IntegerComplete,then the standalone LIMIT. No saved assertion or fixture error.
Different CPU/wall totals are not a matched benchmark or CPU percentages.

The [prepared suite](2026-10-18-p1-integer-direct-tests-prepared.md) now has actual
baseline evidence on both engines. Its direct calls bypass Shape; prepared
type/bounds/positional/default-and-owned-metadata,decoration,zero-callback,
error-order and clean/decorated transition assertions completed in these runs.
This verifies the tested fixtures/semantics on the two observed hosts, not every
possible input or future engine/runtime. No CodeProperty runtime fixture exists;
that stated coverage gap remains. No mutant was run and no type-query count or
performance improvement was measured. TDD red/green for a new candidate is NOT
complete merely because these existing-behavior baselines pass.

Product SHA remains
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
Test SHA remains
E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98.
No product/test changes in this turn. The source's PREPARED,NOT RUN header and
preceding preparation report identify their historical checkpoint; THIS result
supersedes that execution status without rewriting measured/frozen source bytes.
The existing original full-entry output inventory is still unchanged.

## Independent read-only checks and provenance

Rechecked290 frozen pins,manifest/controller,two engines,two receipts/four raw
streams and every ordered output line independently. Verified17 lines/14 PASS,
zero stderr,ordinary context,exit0,under-cap elapsed and owned exit for each.
Pins inherited283 from the loading diagnostic, added its five packet anchors,
prepared test and preparation report. Older records/verdicts were not rewritten.

Ignored local packet `.superpowers/sdd/2026-10-18-p1-integer-direct-baseline/`
retains scope/controller/frozen/index/raw,not committed or required to clone.

| Artifact | SHA256 |
|---|---|
| frozen.json | 3B006860FA358250D42AB5718821F197C051AE64193D79C38CA64E681C5AEB6F |
| run.ps1 | 9EBB7720898CD5564E93D615401F53DBA6BF48D64C30A65B5624913FBB7C9339 |
| invocations.json | 9732C568602393F7C5084F0E2AE598D13848B0EAABCAE1566F157D997BA35898 |
| PS7 stdout | CA28FBE43D7231D5AE3750C694081EC928BF3A89DF6B098F0EC041C3E0E4AF51 |
| PS5 stdout | 28B4AC8A2476FF1B86919B3469615743D8E7104D09BF6AA752F0C8597AA5037E |
| Both empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Engine hashes remain PS7 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139
and PS5 60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Gates and next recommendation

Standalone semantic PASS is NOT original complete-entry acceptance. The original
PS5 timeout remains;I1/P1/candidate and main/C12/security/runtime/release blockers
are not cleared. No type-reuse candidate implemented,independent review started,
or main merge. All older allowances/frozen slots remain,including loading1/2
spent/Inherited FROZEN and review0/1 unused/ineligible.

Next recommend a NEW isolated negative-control allowance,not automatic execution:
PS7 then PS5 once each,30000ms each;only a separate test product copy deliberately
omits integer passive inspection while keeping exact-type/bounds checks. Keep
the prepared suite byte-identical and original product untouched. Expected
negative behavior must be frozen before launch:exit1 at NoteDecorations with
missing-refusal assertion,after the first six groups. Any different failure,
unexpected success or timeout stops/freezes the later host. An expected refusal
here is control success,not permission to weaken tests or alter prior verdicts.
No retry/warmup or prior-budget reuse. This would verify that direct tests catch
guard loss before considering the minimal candidate;it would NOT prove its safety
or speed. Approval and preflight are required before any such new launch.

Recover this report,prepared test,current status and runbook from the existing
development branch;Git does not restore ignored raw/controllers,private runtimes,
credentials or new execution authority.
