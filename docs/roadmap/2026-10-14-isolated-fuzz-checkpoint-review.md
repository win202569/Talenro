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
