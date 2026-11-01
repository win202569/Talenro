# Offline I1/I2 continuation — actual stopped checkpoint

Status: STOPPED_ACTION01_UNEXPECTED_ASSERTION; UNACCEPTED.

Executed the [approved continuation plan](../superpowers/plans/2026-11-01-offline-i1-i2-continuation.md)
after explicit user execution approval “是”. Baseline
92bece45c74bf7a052bd4310cc6ce592f5206f2b, existing development branch
codex/c12-b01-task9-coordinator. Release applied only to the unused v3
7hosts/7targets/2compilations/conditional1review, not a new grant or v2 release.

## Controller and successful preflight

Read-only preflight matched actual2482seed pins/28source snapshots and all44
prior archived originals/12directories. Original driver, runtime seed and old
stop were preserved. Created only a separate continuation-driver.ps1 control
source and its own sidecar. Static AST/LF/JSON and exact original-to-new allowed
region reconstruction passed;19original validation helpers and all28runtime
inputs stayed unchanged. No dynamic control-negative proof was performed.

Approved plan SHA256:
60EF64212A282D1A9D1091FB8717F21DB51BEA1A5545772D60DC340C37B355E9.
Controller SHA256:
E51F86E6CA864A98C0188CB941D3B36331C07F2AEA17E237CB5A80F907C40565.
Control binding SHA256:
CF8ADF3DDBE3EC2E9859FC424985235AAF1DD7D44F46A3F3CD78B9759923011B.
Approval SHA256:
77A8F944AE8A52918B23C169BFD4D7719F5A0E734CECA497EAE3DEC5234E9581.
Actual seed identity remains
9411278BDB6636F38C1D643C647C754194B0D67249958342BF26FC49D6DFDDC1.

## Actual first action and observed failure

Ordinal01 writer-red used one capture PID20520. Capture exit2, not the
required complete4FAIL/9PASS/exit1. Parent command exit1. Exact process owned
exit and stream completion were true; no timeout, termination or rescue.
Parent PostPinsIntact=true, but Exact=false/EvidenceValid=false and
EvidenceState=Unknown. Original stdout228bytes/stderr47bytes remain preserved.

Partial evidence34262bytes exists and is explicitly INCOMPLETE. It contains
7completed groups:4FAIL/3PASS and7RecordCase observations. During the eighth
case IntentDisposeOnly, record-boundary.tests.ps1:53 threw
`ASSERT: first real error retained once`. Three remaining writer groups and
all three final recorder groups are NotRun. This is a genuine unexpected
test-harness assertion, not intended RED, platform filtering or a display error.
Parent first rejects raw output mismatch; the retained capture error identifies
the inner assertion. It does not authorize repairing a frozen assertion.

Journal contains begin/unexpected-error/published, not the required healthy
begin/tests-complete/published sequence.25recorded guard stages through
Finally(IntentDisposeOnly) show all four outer references unchanged. This
supports those recorded observations only: final recorder/private-state/
production-probe groups never ran, so there is no full recorder acceptance.

The preceding two Intent compound cases each saved two FileNotFoundException
secondary projections, another deviation from the frozen one-secondary RED
expectation. The source records inside the writer then records the outer
rethrow, while reference deduplication assumes the same ErrorRecord object.
This supports a rethrow/reference-identity duplication hypothesis. The aborted
case's actual first.Count and private object identities were not serialized;
their exact values/mechanism remain UNKNOWN. No diagnostic experiment, source
patch, test change, retry or extra host was performed to confirm that mechanism.
Historical v2 first internal cause remains separately UNKNOWN.

## Stop, consumption and remaining limitations

New continuation sidecar stop.json records the actual parent failure; original
historical stop remains intact. Only actions/01 exists with consumed intent and
complete actual tool object; no terminal-success record. All8later actions,
both production fixes, later freezes and conditional reviewer were not attempted.

Actual consumption:1capture host/0targets/1in-memory fixture compilation/0reviews.
Remaining v3 6hosts/7targets/1compilation/conditional1review are FROZEN. Older
v2 limits remain separately frozen/exhausted; none are refunded. Runtime source
snapshots remain28, not84. I1/I2 unresolved/unfixed, five capability flagsfalse;
C1/C2 open, R2/Minor2 deferred. No repository/full-C12/production acceptance,
main merge, release or review eligibility.

Additional handoff limitation: the author's independent post-stop integrity
command had an extra closing brace and failed parsing before any check ran.
Actual ParserError/exit1 is preserved, not relabelled success. No corrected
retry was performed. Earlier successful full preflight and the parent's actual
PostPinsIntact=true are separate evidence; do not claim a successful independent
post-stop full-pin audit. Mechanical export byte checks and staged-byte checks
are not substitutes for that missing audit.

Ruling: approved finite Native/PowerShell workflow replaces skill-default
setup/Bash/baseline/full-suite/task-done reruns, implementation agents/per-task
review, automatic fixes and deletion. Cost: only scoped/static-control assurance,
not dynamic negative-control or broad runtime acceptance. Stop rule took
precedence over any skill-default reproduction/fix experiments.

## Portable handoff

portable-continuation.json preserves69actual files/28directories including root:
12control-sidecar files plus57runtime-packet files, including original28runtime
snapshots, partial error evidence, journal/raw streams, consent/control source
snapshots, full action tool object and actual failed integrity command result.
Decoded lengths/hashes/bytes verified against originals. Portable SHA256:
F43A27BD75FD03CDBE9CC5AB60836E87815CE980AFFCE426E964658D912A7485.
Earlier archives were not replaced. Ignored raw directories remain local;
ordinary control source/docs/archive provide the development checkpoint.

Do not rerun01, invoke02..09, remove either stop, edit frozen source/assertions
or create a new runtime namespace. Next requires separate written scope and
approval for any further diagnosis/repair/verification. Prefer examining the
already-saved error-reference assumptions before proposing another runtime
attempt; no new attempt is implied. Clone restores progress/evidence, not live
Session/Ack, execution permission, dependencies or remaining-grant release.
