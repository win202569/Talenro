# Bounded copy transport observation — 2026-11-04

## Final outcome

COPY_PREFIX_COMPLETED / NOT_ACCEPTED. Exactly computation29 and host8 consumed;
total29/8/3/217. Preparation checkpoint ee763f897b5be21f876012b2d86a78ef8c8b9901
was pushed and remote equality confirmed before the single invocation. No retry.

Both copy tasks reached RanToCompletion, without copy/capacity errors or deferred
release. Owned child exited0; stdout is exactly two CRLF-terminated PASS lines
for ContractAPIAvailable and BaselineCore (48bytes), stderr is actually empty.
Both real groups and schema rows passed; cutoff retained, no later groups ran.
Independent verification checked2439 hashes/2405unique paths and all four parent/
child before/after maps.31 operational sources and old evidence remain unchanged.
Journal8920bytes contains62events/31closed intervals, exact prior structural
skeleton, ordered timestamps and valid UTF8/LF. No timeout/termination/overflow,
child/recording/identity/cleanup errors. Five acceptance flags remain false.

## Timing observations, not causal A/B

Milliseconds from separate runs; both transport mechanism and destination changed.

| Measurement | Saved28 | This29 |
| --- | ---: | ---: |
| Parent preflight | 942 | 1358 |
| Parent execution | 13234 | 16239 |
| Parent cleanup | 37 | 204 |
| Setup | 237.4735 | 343.7183 |
| First context | 864.0919 | 3696.3082 |
| Second context | 2904.4079 | 3294.6649 |
| Third context | 3542.1268 | 2953.3766 |
| Wrapper | 1192.4395 | 942.4180 |
| Terminal flush | 3.7371 | 3.2103 |
| Child input before | 876.1469 | 1317.5187 |
| Child input after | 2349.5256 | 2165.5757 |

This29 shared drain measured52ms. No speedup or root-cause claim follows from
these observations. CopyToAsync into fixed memory completed on this prefix;
the original direct-to-disk CopyToAsync route, historical transient conditions,
later52groups and full publication remain unvalidated. Failure branches were
statically inspected but not behaviorally exercised. No business-source fix.

## Recovery and next boundary

Final portable-runtime-copy-transport-final.json:26files/2directories, no unavailable
outputs; all decoded lengths, hashes, containment and bytes checked.
SHA256 4887A84DD04E8D9D6FA2E1B40803C0E7EE9296C6C90AE0741F81A323C52B10D7.
Raw preparation/run/verification records and scripts are included in this portable
archive; ignored local session files are not staged individually. Do not rerun.
Next useful step: read-only synthesis of saved observations to decide whether
another bounded diagnostic is justified. New execution requires a new scope;
this result alone is not acceptance and must not be merged as a product fix.

## Preparation

Plan825b98a7 approved by user. Four child-side files mechanically derived from
diagnostic28; a new parent uses two nonexpandable65536-byte buffers and asynchronous
stream copies.14 static checks passed; independent2439 hash comparisons and4
byte-exact reverse transformations passed. Diagnostic not invoked at preparation.

Preparation archive:15files/1directory, decoded bytes verified;
SHA256 A581469AE62D66A670ED9EE37ACE57905A887B2DB64A5DECE5FBFB0E111B988B.
InputSHA256 E45AAD367F9AF3C30012FF870186682552FD4AB30DE344A885A9A8FE2CE1A845.
Counters28/7/3/217. Exactly one approved computation29/at most host8 remains.

Git whitespace check reports one inherited blank EOF line in common.ps1:102.
Preserved intentionally for exact transformation reversal; raw/index blobs match.

NOT_ACCEPTED. Static checks do not exercise failure branches. The original disk
transport, later52groups and full publication are outside this observation.
