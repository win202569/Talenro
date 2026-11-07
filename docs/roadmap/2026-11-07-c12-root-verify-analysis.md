# Root verifier: read-only boundary analysis

Status: analysis complete; cause unresolved; no new runtime authorized or run.
Source checkpoint d6003351;44/34/6/217,NOT_ACCEPTED unchanged.

## Evidence checked

Read-only e36508 decoded and length/SHA256-checked all13 entries of
portable-c12-root-verify-completion.json; outer SHA256 remains
70F1509BB843F46A53AF1C600D2C64AABADB696C47ADD157FD7BD03E2B8E162F.
The300.059s deadline receipt contains no phase attribution. Earlier120s process
counters show activity only for that earlier run,not progress in the300s run.
Selected two-module success proves neither full graph loading nor full cache health.

Installed Go1.26.5 source read in b1ccb7/3ab137/e36508:

- src/cmd/go/internal/modcmd/verify.go SHA256
  8441FFFD2705C4CEA48F7DE9C2445A6D108AEBB89353DEA4E935C1B79E478EFF.
- src/cmd/go/internal/modcmd/graph.go SHA256
  3B79CCEDB53E7033A1E3D8FC669C975A4206849CC62DF124DEAF2FB951943ABC.
- src/cmd/go/internal/modload/buildlist.go SHA256
  70E7D57F9A7308538335CB5C0957ABC8AEDA352159870CB86C526FE21455269D.

These are observed local source hashes,not independent binary/source attestation.

## Why silent output is inconclusive

runVerify initializes module state,then calls LoadModGraph with an empty version.
Only after graph loading does it obtain BuildList and dispatch verifyMod workers.
The semaphore limits concurrent workers to GOMAXPROCS. Each worker returns errors
through its buffered channel. The main goroutine finishes the dispatch loop
before collecting errors in module order. Thus already-detected errors may remain
unprinted while dispatch waits for other workers; even collection can wait for
an earlier module. Success is printed only after all results are collected.
No stdout/stderr therefore distinguishes neither graph loading,worker scheduling,
ZIP/directory hashing,nor queued errors. It is not evidence that no error occurred.

## Narrower diagnostic recommendation (not execution approval)

Question: can the shared dependency-graph-loading path complete in this environment
without hashing cached ZIP or extracted directory contents?

Candidate: one pinned native `go mod graph` with no flags/arguments beyond
`mod graph`,offline environment and GOMAXPROCS2 as before,existing bounded owner,
empty Session,new exclusive evidence root,maximum60000ms and existing4MiB limit.
No custom helper/build,collector,cache enumeration,repair,download or verify retry.
60s would be a new finite diagnostic budget,not a promised completion time.
Before running,prepare a separate approved execution recipe with exact pins,
module baselines,output handling and immutable claims. No such invocation ran here.

Rationale: graph and verify both initialize the same module state and call
LoadModGraph(state,ctx,""). With no -go override,graph does not switch that version.
After loading,graph walks requirement edges and prints them;it does not call
verifyMod or dirhash.HashZip/HashDir. LoadModGraph itself loads module metadata
and expands the graph. Offline/read-only flags are not a filesystem write barrier.

Interpretation gates for the proposed investigation:

- Normal exit with valid graph output,empty stderr and unchanged module files:
  this graph invocation completed in its observed time. It narrows the next
  question toward work omitted by graph,without proving the previous timeout's
  location (cache warmth/time differences remain). Graph edges are not the selected
  BuildList;do not present every edge endpoint as a verified module inventory.
- Deadline/empty output: still cannot isolate LoadModGraph from startup or graph
  traversal/output. No cause claim,no extension or automatic retry.
- Error/quota/changed files/unknown exit: preserve exact evidence and stop.
  A missing-metadata error is not permission to fetch dependencies.

Graph output may include module names/local replacement metadata. Review retained
output before portable publication;exclude sensitive paths/credentials and do not
blindly commit it. Preserve local original evidence. No main merge/full acceptance.

This recommendation tests one boundary rather than repairing an unproven cause.
Current turn uses systematic debugging's evidence-first phase;no implementation,
Go command,tests,Docker or new runtime monitoring. Author read-only review only.
