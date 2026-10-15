# Original Envelope preparation001 — stopped before build

Later separately approved work introduced -PreparationOnly and executed a new002
identity once. Version/origin and final inventories passed, with no build/fuzz;
see [bounded preparation002 evidence](2026-10-15-envelope-preparation-only.md).
This does not repair or reattribute001's saved unknown-ownership failure below.

Subsequent approved work adds per-API acquisition diagnostics, not a preparation
retry or fix for this87. New tests pass, but a related descendant-fixture regression
remains blocked. See [current diagnostic increment](2026-10-15-member-query-diagnostics.md).
No new fields were retroactively added to the failed receipt below.

## Authorized execution and outcome

The user requested continuation of the previously approved single original-project
Envelope observation after checkpoint329394f3. The unchanged tracked
`Invoke-EnvelopeObservation` driver was invoked once with the verified instrumented
copy, existing module cache, offline readonly settings and new
`envelope-observe-001` run identity. No automatic retry or wider validation ran.

Preparation stopped at **origin**, before any product build or fuzz. The copied
Go version command passed; the copied Go environment command exited0 and printed
the expected copied GOROOT/GOTOOLDIR and original project go.mod, but its process
completion gate failed. Correct output/exit0 does not override missing member
evidence or turn forced cleanup into natural completion. The driver saved
`stopped.json` with `SamplingExecuted=false` / `Acceptance=false`, then its finally
saved and compared the original product inventories. No driver/native code was
changed to bypass this failure. This is an unfinished diagnostic checkpoint.

## Retained evidence

Times below are UTC, with Asia/Tbilisi local13:27 corresponding to UTC09:27.

| Stage | Time | Recorded facts |
| --- | --- | --- |
| version | 09:27:32.825–09:27:32.960 | Go1.26.5 windows/amd64, exit0, no timeout; natural completion, accounting0/member list empty, no termination requested |
| origin | 09:27:33.189–09:27:33.263 | Root exited0, no timeout; one listed non-root member had unknown state/ownership, QueryError87, no exit-code read or final-handle proof |
| origin cleanup | shared shutdown elapsed1ms | Natural completion/evidence false; termination requested only for the owned job, cleanup verified, accounting0/member list empty |
| stopped | 09:27:33.325 | Stage origin; `Natural shutdown not verified; forced cleanup is not success`; sampling/acceptance false |

The origin receipt uses schema3 / `owned-final-signaled-exit-image-optional-v2`.
Its unknown member was not eligible for the final wait, so the runner failed
closed instead of waiting out the remaining500ms. This was not a60-second timeout
or a500ms final-wait expiry. Image path was absent; its identity must not be guessed
from other runs or classified as conhost. No archived PID was reopened or queried.

Static inspection narrows the recorded failure to the pre-ownership part of
`InspectMembers`: opening the listed process handle or verifying that handle's
job membership. The saved row does not identify which API returned87. A member
disappearing between enumeration and handle acquisition is a hypothesis, not an
established cause; ownership-query failure is not excluded. This is a different
evidence stage from the earlier short-fixture final-exit assertion failure, which
already had an owned retained handle. A common root cause has not been proved.

Original product Go/SQL/module inventories contain701 files before and after, with
zero Path/SHA256 differences. Installed six-source/current-template diagnostic
identity still equals
`E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306`.
No build/execute directory, product test binary, binary receipt or sample result
was created. No dependency download/install, official-toolchain patch, Pair,
product matrix, completed-arm replay or acceptance upgrade occurred.

Raw evidence remains in ignored worktree-relative
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/envelope-observe-001/`.
Receipt SHA256 pins:

- version/receipt.json: `FCB6DFB83C06F8B414E826B0BE11F6CA7DA5CC1286D52C407E12B412CBE93140`
- origin/receipt.json: `C39D4015B8C5695B993DB8A646134DE210C9EBAE4D2CC7D38AB1875EA397E73F`
- stopped.json: `6C534B41CF76651555D20D5C34E41CB739323587ED8C3A4E1128AC6AC65DD91B`

Readonly post-checks accepted the saved version receipt, rejected the saved origin
receipt, rechecked the701-file equality and stopped scope, and confirmed all five
post-origin artifacts absent. They launched no processes under observation and
did not modify the saved receipts. Earlier184 scoped test checks remain historical
passing test evidence; they do not erase this new actual-preparation failure.

## Next work and fresh-computer recovery

Preparation001 has been consumed and must not be reused or silently replaced.
Original Envelope sampling is still UNRUN, not failed or passed; no fuzz outcome
exists. A new preparation or sampling attempt requires a distinct purpose and
authorization, not retry-until-pass behavior. Keep Pair/matrix3 and
C11/C12/Task10/B01/production/main gates open.

The next bounded diagnostic should first distinguish handle-open failure from
job-membership-query failure and preserve the operation/phase plus timing. Do not
change unknown-member policy, exempt an image name, infer process exit from later
empty accounting, lengthen the safety window, or reopen these historical PIDs.
Earlier verified-handle final-wait work does not repair missing initial ownership
proof. Any change to acquisition/retention policy needs its own reviewed design
and regression evidence before another actual preparation is considered.

Save necessary reports to the latest existing development branch checkpoint and
verify the remote SHA at handoff. On a new computer, recover
`codex/c12-b01-task9-coordinator`, read current status and the repository recovery
runbook. Tracked driver/templates/reports preserve workflow and the blocked stage;
ignored caches, copied toolchains, raw receipts and live state are not uploaded.
Prepare dependencies separately; cloning does not authorize replaying preparation001
or clearing its failure. This turn adds reports only, not a production fix or new
independent code review.
