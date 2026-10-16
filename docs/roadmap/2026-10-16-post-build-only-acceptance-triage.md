# Post-build-only acceptance triage — read-only, not a new execution approval

Historical triage:its priority1 was subsequently addressed only by the scoped
[ordinary leaf001](2026-10-16-leaf-ordinary-check-001.md). For current ordering,
read the [post-ordinary triage](2026-10-16-post-ordinary-acceptance-triage.md).
Do not repeat this old recommendation as permission for another leaf/matrix run.

Baseline: `ecb05db3212fcd4af420aec78a7d4ab98bccc509`, existing development
branch `codex/c12-b01-task9-coordinator`. The user requested continuing after
build-only001 preservation. This increment reads tracked reports and source;
it runs no Go/native command, test, Docker operation or diagnostic replay.
No source, acceptance predicate, budget or security setting is changed.

## Current blockers, in dependency order

| Priority | Evidence and gap | Required next boundary |
| --- | --- | --- |
| 1 | [Build-only001](2026-10-16-leaf-stage-build-only-001.md) establishes compile/link diagnostic completion only; vet evidence and original-test/runtime acceptance remain false. | Specify an independently bounded ordinary nodeleafbinding verification scope, retaining ordinary vet and completion guards; do not execute its diagnostic binary. |
| 2 | [Matrix003](2026-10-15-limited-matrix-003.md) remains2PASS/1FAIL/23UNRUN; initial member acquisition failed with87. [Saved analysis](2026-10-15-matrix003-acquisition-analysis.md) establishes neither OS cause nor a code defect. | Do not replay/continue matrix003, waive87, or equate retained diagnostic sampling with ordinary shutdown evidence. New execution needs its own approval. |
| 3 | Historical ordinary timeout/cancellation findings and Envelope deadline failure remain open; later successes did not establish their causes. | Preserve original deadlines and failures. Diagnose one defined failure separately; no retry-until-green or whole-suite claim from isolated samples. |
| 4 | Physical PostgreSQL/PITR acceptance and the full Batch01 integration manifest remain incomplete. | Prepare environment/physical evidence separately; do not create a placeholder manifest to bypass registration. |
| 5 | Task10 and subsequent C12 units retain implementation/physical/security gates. | Resume from the latest scoped specifications and source, not historical prose claiming a component is absent; no public serving or main approval inferred. |

## Source checks made in this increment

- `matrix-check.ps1` exposes a fixed26-check `Invoke-LimitedMatrix` controller
  with four path inputs, not a public single-check selector. It loops all checks
  and stops on failure. Its ordinary build arguments do not request `-vet=off`.
  Calling low-level helpers or replacing its plan to skip directly to check03
  is not an already-approved single-package workflow.
- `scripts/run-c12-integration.ps1` still requires
  `testdata/c12/integration-contracts-schema-authority.v1.json` for Batch01;
  that literal file is absent in this checkout. This is a current path check,
  not a live runner result or a statement about Docker availability.
- Current source contains private
  `boundAuthorityReadSource.WithConsistentReadyRead` and its tests. Historical
  status paragraphs saying Task10 has no such mechanism are not current source
  inventory. The [Task10 execution history](2026-10-08-task10-read-boundary.md)
  records the later private prerequisite and its remaining physical/v7 limits.

## Recommended next decision

Prepare a narrowly scoped ordinary nodeleafbinding verification design first:
fresh identity, unchanged vet/completion policy, separately bounded build and
original-test execution, offline dependencies, explicit source/artifact binding,
stop at the first unverified outcome, no automatic retry. Determine whether a
dedicated entry is necessary before implementation. This is a recommendation,
not an approved design, executable plan or permission to launch new processes.
Do not alter the existing build-only entry to make it an acceptance entry.

No new failure cause or fix was established. No new tests were run. Recover this
triage and previous progress from the latest development branch; raw diagnostic
data, caches and SDKs remain local/ignored. Main and production remain gated.
