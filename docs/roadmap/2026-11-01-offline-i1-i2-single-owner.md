# Single-owner I1/I2 implementation — stopped during preflight

Status: STOPPED_PREFLIGHT_AUTHOR_EXPECTED_HASH_ERROR / NOT_ACCEPTED.
Executed preparation after explicit approval “是” of the
[finite implementation plan](../superpowers/plans/2026-11-01-offline-i1-i2-single-owner.md).
Baseline7d4570bd0389b00a8296b2e3d3a9ce0090a7359a; existing linked dev branch.

## Actual progress and error

Clean branch/origin, absent new paths and historical seed identity verified.
Full2482Present pin comparison passed; running tool session83168 was waited to
terminal exit0, with original initial/wait chunks preserved. Recursive28seed
source snapshots then matched current original bytes.

The next referenced-archive comparison used an author-mistyped expected SHA:
`F43A27BD75FD03CDBE9CCF3FCD5334F3D107BC9237` for portable-continuation.json.
Actual SHA returned:
`F43A27BD75FD03CDBE9CC5AB60836E87815CE980AFFCE426E964658D912A7485`.
The comparison exited1. The expected constant is malformed; this failure does
not establish archive corruption or changed original data. It is not an
availability RED, platform fault, display-only error or successful preflight.
Later referenced-archive checks in that command were not reached.

No correction, retry, new suffix, source/test authoring or historical replay
followed. Full preflight remains incomplete even though pin/snapshot checks
passed. The prior failed audit is not retrospectively regraded by this check.

## Consumption and limits

Task1 incomplete; no31source seed freeze or availability RED. No v4 runtime
directory, producer/owner/validator implementation, profile run or independent
runtime review. New diagnostic leaf contains only LF attributes and stopped
portable data. These are handoff records, not the planned runtime/test tree.

Consumed0availabilityREDs/0capture hosts/0targets/0compiles/0reviews.
All NEW2REDs/5hosts/7targets/1compile/conditional1review remain unused but
FROZEN by the stop. Old v3 and earlier limits unchanged/frozen. Five capability
flagsfalse, I1/I2 unresolved, C1/C2 open; no runtime/product/main acceptance.

Ruling: follow the approved unexpected-error stop instead of changing the
author's expected value and continuing. Cost: implementation remains unfinished
before code or budget consumption. A separately approved recovery scope is
needed; unused allowances do not automatically authorize a corrected attempt.

The recording packet .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner
preserves actual consent/ledger, setup tool record, complete pin-check transport,
snapshot/archive-check input/result and exact approved plan/spec snapshots.
New leaf/portable-stopped.json preserves these records byte-exactly; it does not
claim a complete source freeze, tests or missing evidence that never existed.

Portable archive7actual files/1root directory; decoded lengths/SHA256/bytes
matched actual records. Archive SHA256:
8CA37E0E633991DF2B6933D9A23B962F9869094E62E6484F68CC4D2F810337E6.
