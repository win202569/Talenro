# Preparation V2 — original RED recovered without replay

Status: ORIGINAL_RED_RECOVERED_REMAINDER_STILL_FROZEN.

Readonly investigation recovered the original completed CommandExecution
event from this chat's own local rollout log. Event identity:
exec-e8181097-b0bd-4194-862c-2155a509cb0d, completed at
2026-10-31T17:10:18.22Z, duration27.8405118seconds. The exact submitted command
and matching source/test SHA256 bind it to the previously UNKNOWN invocation.
No new test invocation, interpreter, process target or review was started.

## Actual recovered result

The wrapper exited0 because it caught and serialized the test error, not
because the test passed. Original captured JSON records PhaseRED,Passedfalse,
System.Management.Automation.PropertyNotFoundException,
PropertyNotFoundStrict and the exact message:

> The property 'Role' cannot be found on this object. Verify that the property exists.

Serialized original ErrorRecord, stack and converter source are present in
that original event output. Both recorded source/test hashes match the
unchanged files. This confirms the expected optional-Role RED from the
original invocation; it does not claim all later assertions ran.

Saved source-event JSON text without JSON reserialization, SHA256
8A9F21354290DE3A527E8D8CBA664BB752749F86BE992A56131FF3CFA266D787,
11852bytes; saved captured-output text, SHA256
8954D1E6978E42B3F446CA425FBD1359A812452C410C37F6C9CB1DE7E77CEF3A,
8440bytes. These are UTF8 text extractions from original stored app data,
not newly run output or invented raw process-stream files. Earlier UNKNOWN
reports, collection-status and portable exports remain unchanged as history.
portable-recovery.json preserves actual packet records with exact bytes.

## Collector cause and correction rule

The collector parsed empty output after10seconds, while the command completed
after27.84seconds. It failed to retain the complete tool result before parsing.
Future orchestration must retain that result first, distinguish returned
session_id from an observed exit_code, and wait only on the actual returned
session. Preserve each output chunk before parsing; empty/pending/truncated
output is not a JSON result or success. No PID/session guessing or rerun is
needed to recover this completed invocation. This is a confirmed workflow
diagnosis, not a claim that a new collector implementation was runtime-tested.

## Next boundary — proposal, not execution

One regression remains consumed; no refund or replacement RED requested.
Remaining1GREEN regression and1Prepare remain frozen until explicitly resumed.
Role guard/preparation body are still not corrected/implemented. Original
runtime6hosts/7children/conditional1review remain frozen; all8runtime paths
remain absent. C1/C2 open,R2/Minor2 deferred; five acceptance flagsfalse.

The packet now legitimately contains the ledger plus3immutable audit records:
collection-status.json and the2recovered files. The original ledger-only
pristine assumption can no longer be applied literally. Request approval to
allow only these3explicitly hashed records, add their3pins (2075=>2078), keep
17source snapshots/8Absent, and consume only the remaining1GREEN/1Prepare.
Any preparation manifest/snapshot/unknown extra content still blocks execution;
no deletion, replacement namespace, RED rerun, broader budget or review.

This report is the proposed narrow adjustment; no runtime authority derives
from its presence. Actual compatibility fix and successful preparation still
require implementation and verification. No merge or acceptance occurred.
