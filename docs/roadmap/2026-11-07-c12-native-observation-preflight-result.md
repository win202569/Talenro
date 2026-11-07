# Native observation preflight — runtime blocked

Approved plan after f27ae030. Offline source/tests completed;real observation
readiness BLOCKED. No recording,Go invocation,ETL processing,installation,registry
change or elevated tool retry. Global44/34/6/217,NOT_ACCEPTED unchanged.

## Verified work

- Profile RED7bde83 failed the explicit missing-policy assertion;GREEN86fee7
  passed21 synthetic cases plus the shipped candidate,exit0 in767.107ms.
  Tests reject extra providers/keywords,over-budget buffers,non-Memory profiles,
  unsafe XML,inheritance/scaling and malformed identities. Commit54548848.
- Readiness RED993964 failed the explicit missing-policy assertion;GREEN3db6d3
  passed59 decision-contract cases,exit0 in713.131ms. Commit8c5b3471.
  Decisions never grant runtime authorization. This function evaluates supplied
  claims;it neither authenticates evidence nor independently observes the machine.
- WPR Profiles2236b5 and Details152ede exited0;one Memory collector,128 buffers
  of1024KiB,exactly FileIO/FileIOInit/Loader/ProcessThread/SampledProfile. No stacks
  or extra event providers shown. This establishes declared profile shape/capacity,
  not live resource consumption or completeness of captured events.
- HelpAdvanced/Stop/Status87f81a exited0. Named-instance syntax and status/stop help
  are available;no real global status was queried and no session was created.
- All ten supervised processes confirmed exit with no owner timeout/failure and
  unchanged four module files. Two expected RED exits and the exporter crash are
  not passing executions. Owner/common hashes unchanged87cd0b.

## Unexpected exporter failure — no retry

ExporterHelp826c2f:arguments `-help`,elapsed7312.1314ms,exit-532462766,confirmed.
Raw output first reports absent wpaexporter.deps.json,then a fatal type-initializer
exception whose inner SecurityException says requested registry access is not
allowed. Stack includes WpaEventSource.GetValue/KitsCeipOptedIn during initialization.
It did not return usable help. Do not interpret wrapper exit0 as tool success.

Read-only40b9c0 confirms the named deps.json path is absent. The missing-file message
does not prove the cause of the fatal exception or that the installation is broken.
Registry restrictions are observed in this execution context;no registry/security
settings were changed. No elevation retry,repair,reinstall or substitute tool run.
Six of eight allowed help/profile slots used;remaining two unused after failure.

Exporter path:C:/Program Files (x86)/Windows Kits/10/Windows Performance Toolkit/wpaexporter.exe;
SHA256 FC603F0316027C024132C9571A02959AD90C89CBC246B49F08958A74E6CEED62;
file version11.7.395.48728. Hash was claimed before launch;version was read after
failure,not before as the plan requested. This sequencing deviation is retained.

## Feasibility disposition

| Gate | Evidence | Disposition |
| --- | --- | --- |
| Minimal profile | XML tests and actual WPR profile details | Static validation passed |
| Session identity/conflicts | Named-instance/status help only | Unproved;no runtime ownership claim |
| Bounded start/stop and failures | Existing child owner plus WPR help | Collector lifecycle unproved;CLI exit cannot prove session stopped |
| Storage enforcement |128MiB declared ring capacity | Temporary/save/final ETL hard bound unproved |
| Offline analysis | Exporter help crashed | Blocked;actual export fields/timebase/identity unresolved |
| Ring retention/loss | No trace created | Retained interval and event loss unknown |
| Privilege separation | No controller or privileged process | Unproved |
| Privacy approval | Non-recording plan only | Real unrelated metadata collection not authorized |
| Publication | Only synthetic data and non-recording receipts | No ETL to publish;future raw ETL remains excluded |

Actual assessment87cd0b saved BLOCKED and RuntimeAuthorized=false. Resolving the
exporter alone would not clear ownership/storage/privacy gates. Do not start a
synthetic recording or native verify merely to work around these prerequisites.

## Preservation and next scope

Final portable archive:portable-c12-observation-preflight-final.json under
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2;109 entries,166344 decoded bytes,
all entry lengths/SHA compared with originals. SHA256
A5E81FC3D7F97E1BA0A2B557417F225998DFA536BC6201E93DB4452CA93EBF64.
Earlier profile/readiness checkpoints retained unchanged. Archive includes exact
stage snapshots,recipes,raw outputs,assessment and ledger. Expected documented
local paths occur in receipts;no credentials or private trace payload observed.
Vendor schema/sample contents,binaries,caches and ETL excluded. Local reference
hashes (not independent vendor attestation):XSD
3FB7D9A091E16C8BF402E4E4E0EE18BAA1B0BFA5C3397924976EE8034D800B4D;
sample6AF24FB65EBE1B2D8ED2771A5B3A1688DE182F0DD6F26C08DF0AE4C53D0C7B23.

Smallest recommended next scope:read-only investigation of the installed exporter's
packaging/configuration and documented offline help/export path,without relaunch,
elevation,installation,registry reads/changes or trace capture. Determine whether
a supported path exists before proposing any new runtime budget. Do not assume
permissions should be relaxed. Original native300s timeout remains unattributed.

Inline author review only. Approved finite test scope overrides generic full-suite
and cleanup steps;no whole-project suite or independent review claimed. Exact XML
allowlist deliberately rejects benign unapproved extensions (future changes need
policy/test review). All evidence preserved;no main merge or acceptance promotion.
