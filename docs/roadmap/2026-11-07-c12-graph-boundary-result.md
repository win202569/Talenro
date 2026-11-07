# Offline graph boundary — completed

Status GRAPH_COMMAND_COMPLETED;overall NOT_ACCEPTED,counters44/34/6/217 unchanged.
Approved source1197cbd2cacfeecae5d322600d99f57938e92e20;remote83c14e matched.
Recipe syntax6e25e2 passed;invocation7f9b8f passed prerequisites and ran once.

## Observed result

- Pinned Go1.26.5 `mod graph`,offline real cache,GOMAXPROCS2,unchanged owner,
  empty Session,60000ms cap. No other Go/test/build/collector workload.
- Elapsed306.9959ms,Started=true,ExitConfirmed=true,ExitCode0,EndReason=EXITED,
  TimedOut=false,Failure=null. All nine classification gates passed.
- stdout80553 bytes,1059 structurally valid UTF-8 graph edges;stderr0 bytes.
- stdout SHA25680D2ACE98205028DEE609598E4C9E43778636BFD09C355F476584E1EEC712A2E;
  stderr SHA256E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855.
- Four root/devtools module-file hashes unchanged. Baseline was decoded/hash-checked
  from the committed portable root receipt,not previous ignored local files.

The shared dependency-graph-loading entry and this command's graph traversal/output
completed quickly in this run. This narrows the next investigation toward work
performed only by verify (selected build list,worker scheduling,cache ZIP/directory
checks and result collection). It does not locate the earlier300s timeout:time,
cache warmth and environment may differ.1059 is an edge count,not verified modules
or BuildList size. No dependency content integrity or wider acceptance is proven.

## Review and preservation

Read-only9064b5 checked every graph token's character structure,all unique module
paths and sensitive/path markers. The only project-local module identity is the
existing talenro.local/platform;other observed paths are dependency module names.
No unusual tokens,credential/path markers or stderr/error content were found.
Original streams may be included unchanged;no privacy omission/redaction needed.
Receipts/recipe contain expected already-documented workspace/tool paths only.

Portable archive portable-c12-graph-boundary.json preserves exact invocation,
source,preflight/claim,raw streams and receipts. Every included entry is decoded
and compared against original length/SHA before commit. No binaries,caches,module
contents or credentials. All old and new local evidence retained. Author inline
review only;no independent review or whole-project validation claimed.

## Next boundary

This plan's single attempt is consumed;no automatic verify retry or new runtime.
Recommended next step: read-only comparison of saved graph,root go.mod and existing
phase helper's fixed allowlist to define a bounded selected-module diagnostic scope.
Do not equate all graph endpoints with selected versions,hash arbitrary caches,
modify the helper,compile it or launch another check without separate scope.
