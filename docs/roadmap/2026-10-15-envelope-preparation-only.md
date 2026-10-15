# Envelope version/origin preparation only — 2026-10-15

Later separately approved work added compile-only mode and ran fresh003 once:
version/origin and compilation passed, no binary/fuzz execution. See the
[compile-only record](2026-10-15-envelope-compile-only.md). The statements below
describe002's preparation-only scope and do not authorize sampling or replay.

## Approved bounded change

After the diagnostic runner's31/31 regression checkpoint0c03b373, the user
approved an optional preparation-only mode, helper tests/review and exactly one
fresh actual preparation. Only Go version and GOROOT/GOTOOLDIR/GOMOD queries may
execute. No build, fuzz, Pair, retry, downloads/installs, security changes, WPR,
budget relaxation or main merge. Historical preparation001/origin87 remains
consumed and unresolved, not reinterpreted because runner tests now pass.

`Invoke-EnvelopeObservation` adds trailing switch `-PreparationOnly`. Default
execution remains the original observation path. Explicit preparation mode
uses the unchanged planner, copied-executable/source provenance, offline readonly
environment,60s query bounds, strict process completion and expected output gates.
After origin it rechecks product/source identity and exits before build. The
existing final product inventory still runs. Only after final inventory saving
and equality checking does it save preparation-result.json and emit success.
PreparationPassed is version/origin preparation only; PreparationOnly=true,
BuildExecuted=false, SamplingExecuted=false and Acceptance=false remain explicit.

Normal stopped records now include PreparationOnly and BuildExecuted, based on
mode and actual build-directory existence. A preparation-only finalization error
saves stopped.json with Stage=finalize/PriorStage and false execution/acceptance
labels, then throws. Default finalization errors still rethrow. Root freshness,
job ownership, unknown-member refusal, natural exit and cleanup predicates are
not modified. Creating the existing empty cache/tmp directories is not build or
sampling execution. The optional mode does not authorize resuming into build.

## Test-first checks and review

New envelope-preparation.tests.ps1 first failed at the real driver failing to
return preparation success before its build boundary. The external process
boundary is substituted to avoid unauthorized Go runs; version/origin commands,
arguments, offline/readonly settings and60s bounds are constrained. A controlled
single-file inventory avoids repeated full701-file scans in branch tests. Real
planning, copied-source provenance, completion gates, inventory equality, driver
branches, disk writes and streaming result consumption execute. Synthetic receipt
results do not establish actual process ownership or machine readiness.

Final19 assertions passed on PowerShell7 and WindowsPowerShell5.1, covering the
successful early exit, exactly two query stages, saved scope labels, absence of
build/binary/sampling artifacts, final inventory execution, version/origin natural
completion rejection, wrong origin output, default build-boundary preservation,
and final-inventory drift/read errors with no success file or streaming output.
Existing21 envelope-observe planning/provenance checks also passed on each host,
including actual product inventory enumeration:40 scoped assertions per host.
No full runner/product suite was rerun here; prior31 runner/137 helper results
are historical evidence, not freshly rerun tests or project acceptance.

Independent read-only review found two Important issues during implementation:
success JSON was initially saved before final inventory, and moving that write
alone still allowed return to stream a success object before finally. Each was
reproduced with a failing behavioral test before repair. Final review confirmed
both fixed, no Critical/Important findings remaining, and approved the one actual
preparation only. No reviewer ran processes/tests or edited files. Direct copied-
source drift branch coverage remains limited; existing real provenance checks
and source identity revalidation are retained. Actual readiness, build/fuzz and
product acceptance were not judged by static review.

## Actual execution and recovery

One actual invocation of -PreparationOnly used fresh envelope-observe-002,
distinct from001, on the PowerShell7 host. Wrapper exit0; PreparationPassed=true,
BuildExecuted=false, SamplingExecuted=false and Acceptance=false. No retry.

| Query | Native UTC interval | Observed facts |
| --- | --- | --- |
| version | 11:11:41.3996928–11:11:42.9888451 | Go1.26.5 windows/amd64, exit0, no timeout; natural completion, complete evidence, no requested termination, final accounting0/list empty; shutdown3ms |
| origin | 11:11:43.2990095–11:11:43.3824085 | Exit0; copied framework-001/toolchain GOROOT and its pkg/tool/windows_amd64, original project go.mod; natural completion, complete evidence, no requested termination, final0/empty; shutdown4ms |

Both native process counts are2, which is accounting rather than intended-child
proof. No archived PID was queried. Asia/Tbilisi time was15:11. The original
701-file Go/SQL/module inventories match before/after with zero Path/SHA256
differences. Copied executable and diagnostic identity validation passed; source
identity remains E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306.
There is no build/execute directory, nodeissuer.test.exe, binary-receipt.json,
sample-result.json or stopped.json under002. Cache/tmp scaffolding alone exists.

Ignored worktree-relative evidence root:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/envelope-observe-002/`.

| Local evidence | SHA256 |
| --- | --- |
| version/receipt.json | 6B0877FD67F28D57E52CE9E720BD84DD04DD3089004E50D27BA31717F5D73E8F |
| origin/receipt.json | DE6D580E82C93E6A6222D5CEB70C512817BDD0C7EF3911EAF38B7EE540C00F0F |
| preparation-result.json | 83DE16EC8A84585DC32D0C38A8F783A6503CDF2F6C87B6A1A31AC7B4D6B1181A |
| before-product.json and after-product.json (identical) | 5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |

This one preparation passes locally; it does not explain or repair001's unknown
ownership QueryError87, prove that failure cannot recur, or authorize replay.
No product binary was compiled and original Envelope/Pair sampling remains UNRUN.
C11/C12/Task10/B01/production/main gates remain open. Preparing successfully is
not product acceptance or permission to invoke the default auto-sampling driver.

Save necessary source/tests and public summaries on the existing
codex/c12-b01-task9-coordinator branch, verify remote SHA at handoff, and restore
that checkpoint on a new computer. Read current-status.md and
docs/runbooks/repository-recovery.md. Raw receipts/caches/toolchain copies are
ignored local artifacts, not uploaded. Recreate prerequisites separately; cloning
does not restore the local environment or authorize replaying001/002.

Next proposed step is a bounded compile-only mode with tests/review, because the
default driver still proceeds directly from build into fuzz. Any actual compile
needs a separately approved fresh identity and current provenance/environment
checks; do not reuse preparation-only002 as a writable build destination or
start fuzz under this preparation approval.
