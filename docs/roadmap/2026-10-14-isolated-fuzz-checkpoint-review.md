# Isolated fuzz diagnostic checkpoint — independent review

2026-10-14; range96a70850..3df42326 plus new preflight files. One fresh-context
read-only reviewer: no tests/processes, edits, index/HEAD/config changes or extra
reviewers. Not five-arm completion or main approval. Author ran fix regressions;
no re-review was dispatched.

## Findings and disposition

Original:0 Critical,2 Important,1 Minor/recovery gap. Author re-grade:0 Critical,
3 Important,0 deferred Minor: broken fresh-computer recovery is Important.

I1 — shutdown completeness conflates missing image metadata with unknown exit.
The member was exact-handle-verified owned, wait-signaled/exited/code0, QueryError0,
but ImageError31. Signaled process handles establish termination; image lookup is
separate. See [Microsoft process termination](https://learn.microsoft.com/en-us/windows/win32/procthread/terminating-a-process).
This neither identifies the missing image nor explains error31.
Disposition: unresolved design-policy blocker. Current approved guard remains
binding; the receipt requested termination and cannot retrospectively PASS.
Proposed bounded revision separates ownership/liveness from image metadata while
keeping natural zero/empty, fixed bounds, unknown ownership/state, truncation,
timeout and forced-cleanup refusal. User approval required before implementation.

I2 — current portable templates and reused binaries were not source-bound.
Disposition: fixed in one scoped TDD pass. provenance.ps1 requires exactly6
installed files, receipt/actual hashes, all4 current portable templates and patch,
then computes canonical source identity. Binary receipts bind that identity;
missing/mismatched identity refuses reuse without automatic rebuild/arm retry.
Template-only drift and binary identity mismatch each RED→GREEN;7 assertions PASS.
Actual installed framework sources match current templates read-only. No old
receipt rewriting or actual fuzz process launch for this fix.

I3 — fresh copying did not save the receipt required by the runner; README led
with stale status. Promoted from Minor/recovery gap to Important by actual effect.
Disposition: fixed. After three manifests agree, New-IsolatedToolchain saves
copy-receipt.json using CreateNew, with returned/saved identity/manifests equal.
Missing-receipt RED→GREEN;18 real-link isolation assertions PASS. README current
state moved first and receipt handoff documented. No real full-copy/arm retry.

## Review Focus

Reviewer confirmed explicit prefix/reparse/hardlink guards/tests; copied GOROOT,
origin and compiler tokens; selected-context registry/unregister release and
release-before-worker-cancel; original short-circuit parent Err capture and unread
return fields; trace/watchdog/worker completion requirements rejecting nonzero
exit alone. No definite early-failure deadlock found by inspection, not execution.

## Declined-to-judge rulings

- Five remaining arm outcomes: UNRUN; missing integration evidence stays a blocker.
- Fresh test execution: reviewer was read-only; author reruns are not independent
  test reproduction.
- Coordinator early-failure integration beyond components/source: untested, no
  broader integration or timeout-exemption claim.
- Full source/product equality: still running during review; author subsequently
  verified11536 original rows and701 product hashes unchanged.
- Missing image, error31 cause, console-host creation/historical fuzz attribution:
  unknown; no archived PID queries or causal claims.
- Concurrent hostile filesystem mutation/arbitrary toolchains/unrelated behavior:
  outside bounded owned-job diagnosis; no general sandbox guarantee. Cost of wider
  claims would be unsafe transfer of this task's scope assumptions.
- Product/Envelope/Pair/matrix/toolchain promotion/main: excluded; existing gates
  remain open, diagnostic success cannot clear them.

## Outcome

Independent review completed for an unfinished checkpoint. Two issues fixed and
verified; I1 awaits approved policy revision. Context5 logic/completion passed;
trace3 logic/exit0 passed but process completion failed. All five fuzz arms remain
UNRUN; stop was correct. Not merge readiness or a finished coordinator study.

## Subsequent user-approved I1 revision (author verification, no re-review)

The user approved separating proven owned/signaled exit from missing image
metadata, retaining unknown-state/ownership/query failure and timeout/forced-cleanup
refusal plus500ms/shared5000ms bounds. A shared native predicate uses the exact
observed state and successful exit query, never code0 alone. Image completeness
is separately recorded.9 state tests RED→GREEN,90 total PS assertions PASS.
Fresh gate-framework-state-002 context5 + trace3 logic/completion PASS; trace run
captured owned/exited/code0/ImageError31 and naturally empty job in5ms without
termination. This verifies the bounded change, not a retrospective PASS for the
old forced-cleanup receipt or an independent re-review.
Continuation stopped at observe's source probe: owned member initial wait running,
later code0 and ImageError5. No later signaled-state read was captured on that
handle; no identity inferred, current guard remained failure. No fixture build or
actual arm executed. Retaining first-read facts and confirming final state on the
same verified live handle, plus a fresh preparation identity, needs an explicitly
bounded continuation; do not query archived PIDs or overwrite the failed directory.

## Subsequent author closeout — six actual controls, not a second review

Following separate human approvals, final schema3/v2 policy requires same retained
owned-handle signaled exit/query plus naturally empty job within shared500ms;
prior reads/images remain unchanged and diagnostic-only evidence cannot upgrade.
Full bounded-batch handles are gathered before waiting; unknown initial identity
still fails.183 scoped assertions passed (11 duplicate reused checks;172 distinct).

All six actual arms now passed their prescribed diagnostic checks, with completed
arms never repeated. Remaining synchronized-prep-002 exit0, injected crash exit1,
and early cancel exit1 have full worker/return traces, watchdogfalse, empty owned
job and no forced cleanup. Original11536/product701 hash inventories unchanged.
Earlier failed preparations remain failures; source/binary and recovery fixes stand.

This closes the missing five-arm evidence only at author-verification level. The
reviewer's original UNRUN statements/grades remain historical; it did not observe
these later executions or independently review their final receipts. No re-review
was dispatched. Early-failure behavior beyond the specific warmup-crasher/selected
cancel controls, hostile concurrent filesystem changes, arbitrary toolchains and
product acceptance remain outside the judgment. Unknown image-error/history causes
are not inferred. Full results do not establish main merge readiness or production
trust; see the execution report for the bounded evidence and recovery limitations.
