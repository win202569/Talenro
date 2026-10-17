# Empty-control diagnosis — IsInstance predicate false positive

## Approved question and answer

User approved two fresh PS7 diagnostic hosts after8364f1a0: clean versus the
same fixture preflight, each <=30000ms. Each host calls the candidate decision
function once on an owned ordinary empty scenario, then directly invokes root
shape validation once in StrictMode Latest to expose its exception. Record the
actual decision and five primitive fields' member descriptors. No original test
script import/run, R3 replay, candidate/test edit, retry, fallback or reviewer.
This is a diagnostic spike; the retained probe is throwaway inspection tooling,
not production or a regression-suite replacement.

**Finding: the candidate's IsInstance-only rejection misclassifies ordinary
primitive members.** Both modes returned the same invalid-scenario refusal.
Direct root validation threw Decorated scalar at candidate module line8. Even
the clean host's Int32 members include CompareTo/Equals/etc with IsInstance=true;
the predicate therefore rejects an ordinary input without any fixture preflight.
The prior choice to treat that flag alone as reliable added-member provenance
was incorrect for the observed runtime. No product repair was made in this task.

This explains a concrete ordinary-control rejection path consistent with the
stopped P1EmptyPrepared failure. It does not retroactively add an internal trace
to the old R3 receipt, prove every possible R3 influence, or establish that
preflight has no effects on any other values. It establishes that preflight is
not necessary for this failure, and the five saved descriptor lists match across
these two controlled executions. PS5.1 was not diagnosed here.

## Execution and observations

| Host | Preflight | Exit | ms | stderr | Transport/integrity |
| --- | --- | ---: | ---: | --- | --- |
| D1 PS7.6.5 Core x64 | none | 0 | 932 | empty | accepted |
| D2 PS7.6.5 Core x64 | copied seven fixture groups PASS | 0 | 1882 | empty | accepted |

Both ordinary-token/admin=False, TimedOut=false, owned-host-exited. Elapsed time
includes launch/wait/stream collection. D1 has four exact ordered output lines;
D2 has12 including seven PASS lines and preflight LIMIT before diagnostic data.
ExpectedResult=true means diagnostic transport, not a passing product control.

Each actual result is Schema1, Mode policy-only, EvidenceClass synthetic,
Evaluated=false, Reason invalid-scenario, State Refused, FailureObservationIndex
null, Decision null; OriginVerified/Acceptance/RuntimeApproved/LiveActionsPerformed
false and CleanupStatus unknown. DirectShape.Passed=false in both, with Message
and Id Decorated scalar. Captured stack passes through Passive line8 and Shape
line22. Raw Position retains the host's rendering, including replacement glyphs;
no raw text was corrected or rewritten.

The original decision function catches its internal error; the separate direct
validation reveals the matching rejection path without altering that function.
One decision call and one **additional direct** shape call per host were made;
ShapeCalls=1 in the diagnostic means direct calls, not internal product calls.

Saved descriptor totals are identical in both modes:

| Field | Base type | Members | IsInstance=true | Member types |
| --- | --- | ---: | ---: | --- |
| Schema | Int32 | 28 | 28 | Method |
| Mode | String | 56 | 56 | Method, Property, ParameterizedProperty |
| SessionLabel | String | 56 | 56 | Method, Property, ParameterizedProperty |
| ClockLabel | String | 56 | 56 | Method, Property, ParameterizedProperty |
| StartMs | Int32 | 28 | 28 | Method |

No NoteProperty/ScriptProperty/AliasProperty/ScriptMethod appears in those saved
primitive-member lists. Descriptor collection does not read member.Value or
invoke its callable content. Result and field-descriptor JSON compare equal
between modes. We did not instrument which descriptor the original R3 visited
first, so the report does not claim an old first-member trace.

## Integrity, scope and recovery

Static parsing passed for probe/controller. AST text comparison confirmed the
nine fixture helper definitions copied into the probe are identical to frozen
ownership86FD5C12. No original test script was imported; D2 executed only those
copied fixture helpers before the authorized product import. Both modes imported
the same unchanged productA38C1EE2. The synthetic input has the same empty-scenario
field/value shape as the original control; the original control test was not run.

Frozen manifest retained62 inherited pins plus14 prior stopped-packet anchors and
two current product/test pins:78 entries. Old source/tests/manifests/raw retained
unchanged. Each slot was consumed before launch; D1 expected receipt and stream
hashes checked before D2. Exact schema/member shapes and preflight inventory
were frozen before D1; neither refusal nor validation exception was treated as
transport failure. Separate UTF8 streams, hidden no-profile process, no shell,
inherited ordinary token and owned-handle timeout/cleanup rules were preserved.
Timeout/error-cleanup branches were not exercised by these fast successful hosts.

Portable probe: `docs/diagnostics/isolated-fuzz/offline-supervision-policy.empty-probe.ps1`.
Local ignored packet `.superpowers/sdd/2026-10-17-p1-empty-control-diagnosis/`
retains scope, frozen controller/manifest, index and four raw streams; no packet
was deleted. Git recovers this probe/report/status/recovery and candidate from
the development branch, not raw/runtime state or permission to run it again.
Staged whitespace checking reported one extra blank line at the probe's EOF
(line152). The executed probe bytes are deliberately retained unchanged to match
the frozen hash; this harmless formatting warning is recorded, not claimed as a
clean whitespace check. No behavior or gate was changed to address that warning.

| Artifact | SHA256 |
| --- | --- |
| Probe | `374D79485D999ECCF7226702DBA197FC935E7644438D68337B389D322AD621FE` |
| frozen.json | `F7DEEFE55BC6B70AD44020EDEF5C0448C9EB2C4401C187FE8C0F46A07F6B5182` |
| run-slot.ps1 | `33CC1F384D1FADE0D14EFB3F28BE958A909DEF4AA5BDD599E9287CB89CDA67EA` |
| invocations.json | `957DA9B2D18CD22401DACA2BF42AB315DEE909060B7B14459E401691A3DB3100` |
| D1.stdout | `6F7E0EAED505FFC58E2115AABEF43625EFD4AB135D18B9236CFCE5C564625003` |
| D2.stdout | `B15EC26922B821F85063DB30CC94318757C9595BB6CD61781D1560EC575DB9E2` |
| Both stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

## Remaining work

New diagnostic2/2 spent. ProductA38C1EE2 remains an unaccepted candidate; ownership
suite86FD5C12 and original35 suite61A3BACF unchanged. Prior R3 stays false; repair
3/7 spent/four frozen, conditional review0/1 ineligible. All other historical
spent/frozen budgets unchanged. No main, P2/P3/runtime, project-wide or C12/release
acceptance. Next scope a replacement added-member discriminator with explicit
ordinary-value controls and decorated-value refusal tests, not a blanket removal
of scalar protection. Fresh repair/verification authority is needed; no fix,
mutant, re-review or further diagnostic was started here.
