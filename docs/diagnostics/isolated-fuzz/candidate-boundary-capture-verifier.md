# Offline candidate boundary capture verifier

Dot-source candidate-boundary-capture-verifier.ps1 to define
Test-CandidateBoundaryCapture. Supply saved stdout/stderr strings and a parsed
receipt object. The function never launches/imports the policy or tests,never
writes files,and never revises the supplied receipt. It reads only the adjacent
integer-candidate-boundary-markers/expected-output.json inventory,requiring its
preapproved SHA256. No executable instructions are read from captured text.

The verifier uses that inventory's HostRule directly,avoiding the frozen
controller's overescaped copy. It validates ordered literal/context/host/marker
records,Int64 nonnegative/nondecreasing ticks,and only computes closed pairs.
Required receipt context fields must exist with boolean/numeric types;frequency
must be a positive integer,elapsed/CPU finite nonnegative,elapsed positive,
parent ordinary,child launched,streams saved,and owned-host-exited cleanup.

Results: invalid-capture rejects malformed text or missing/invalid receipt facts;
valid-prefix means valid observed output but missing tail or incomplete execution
gates;complete-diagnostic requires91 lines/64 PASS/eight markers/four pairs,
exit0,no timeout,and elapsed<=30000ms. A full stream with exit failure/timeout is
also classified valid-prefix,NOT complete. All results retain
FormalProjectAcceptance=false. Missing ends never get invented durations.

This is NOT an authenticity/security boundary:it consumes caller-supplied text
and facts,not authenticated evidence. Callers must separately verify saved file
hashes,source/controller/engine identities,receipt count,provenance and preservation
of original files. A forged receipt is not made trustworthy by valid syntax.
No process ownership,actual token,CPU measurement,frequency provenance or
product correctness is proved by this function. Existing frozen false verdicts
stay unchanged;return values are a separate offline assessment.

For portable synthetic tests only,use PowerShell7 to run
candidate-boundary-capture-verifier.tests.ps1. It builds a synthetic complete
stream with fixed integer ticks from the committed inventory and exercises the
real verifier. It does not need ignored session records,caches,credentials or
product executions. Saved-output integration is separately documented in the
2026-10-18 offline-verifier roadmap report. PS5 verifier compatibility is NOT
verified. Neither this tool nor tests grant replay/unfreeze/acceptance authority.
