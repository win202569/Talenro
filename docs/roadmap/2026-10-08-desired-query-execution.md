# SDD ledger — plan: docs/superpowers/plans/2026-10-08-task10-desired-query.md

Approved spec and plan, Native execution. Tasks 1–3 complete locally; uncommitted/unpushed.
BASE/HEAD 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee; merge-base main
e1571f9fb1940f341d8e40f55907bec5e4f30dcb. Existing linked worktree, branch
codex/c12-b01-task9-coordinator; no submodule. All prior dirty files preserved.
Baseline serving PASS 2.573s. Four lock hashes unchanged from prior phase.

Pre-flight: Task 1 generated row aliases feed Task 2 conversion; compatible.
Task 3 consumes Tasks 1–2 evidence; no live DB evidence is available or claimed.
Ruling: use this portable ledger instead of Bash SDD scripts; retain it rather
than delete it because no commits are authorized — Windows PowerShell only and
cross-session recovery — cost: manual checkbox/evidence tracking.
Ruling: limit suite/generation and retain local changes per approved scope,
overriding skill full-suite/commit/setup-download defaults — existing safety
boundary — cost: DB/PITR/production acceptance remains unproven.

## Final review and fix pass

Fresh reviewer found no present SQL/converter correctness defect and no Critical
or new Minor finding. One Important finding: static GuardShape did not detect
removal of exact fence-presence/scope/coordinate predicates and false defaults.
Re-grade: Important stands, because the approved fail-closed regression boundary
was incompletely covered even though current SQL contains the protections.
Final: fixed I1 — TestDesiredQueryGuardSensitivity RED 4.672s, then strengthened
all four fence-presence/coordinate/scope requirements, three false-default flags,
current pointer/version/epoch relationships and terminal/status requirements;
query GREEN 3.381s. No production SQL change was needed. This is static checker
sensitivity, not real database mutation/execution evidence. No second review.

Final scoped suite PASS: serving 4.833s, contracts 6.156s, readiness 4.114s,
authority selector 5.269s; store compile-only 2.491s (no tests). All exit 0.
Generated files identical to stable post-generation baseline; four locks
unchanged; gofmt -l and git diff --check clean. No running tests remain.

Review coverage: spec/plan/ledger, desired SQL source, row converter/new tests,
existing desired projection/tests read completely; generated desired types/scan,
Querier, supporting schema/uniqueness, helpers and caller references inspected.
Historical merge-base..HEAD was not audited as a whole; unrelated devtools,
testinfra/coordinator history is not covered by this review.

Final: Ruling: retain all reviewer exclusions as explicit unproven gates, not
waivers — real PostgreSQL/planner/decoding/snapshot/migration/ACL/corruption/PITR
needs permitted DB evidence; epoch/scope derivation, cancellation/error ordering,
panic cleanup and staging erasure need the absent adapter/reader; canonical,
signatures, trusted time/current authorization/envelope/network limits are later
layers; initial-root writer availability, history retention, cross-epoch recovery
and limit deployment need writer/runbook validation; v7/production/B01/C12/main
acceptance needs its gates; race/full suite/security remain prohibited; prior
certificate sameUUID and desired M1 remediation remain deferred; reviewer did
not independently rerun generation/tests (parent did); broad historical branch
remains unaudited — cost if these limitations are overlooked: unsupported
production/whole-project acceptance. None is a claim of correctness outside the
private increment. No new deferred minors; old minors remain as documented.

Task 3: complete — scope-limited tests and one final-review fix pass complete.
Keep worktree and portable ledger; no index write, commit, merge or push.
Next: design private desired query adapter lifecycle and bound composition,
then canonical-envelope gates before any public serving interface.

Task 3 regression PASS: serving 3.542s, contracts 4.248s, readiness 3.008s;
authority selected groups 4.268s; store compile-only 1.948s (no tests).
gofmt -l and whitespace checks clean; four dependency lock hashes unchanged.
Fresh read-only reviewer /root/review_desired_query (gpt-6-astra, high) dispatched
with full spec/plan, ledger, branch bases, local file inventory and Review Focus.
At dispatch there was no review conclusion. Historical branch is much broader than this increment;
actual review coverage must be reported explicitly, not inferred from dispatch.

Task 2: complete — missing function RED; fail-closed stub behavior RED 3.165s;
row GREEN 2.929s; added ownership and single-row leaf rejection GREEN 2.949s.
Removing publish-envelope cloning caused Ownership RED 2.954s (input alias
RootPublishedEnvelope); restored immediately. Full serving GREEN 3.190s.
The new test checks each mutation before restoration and explicitly checks
owned tail/active-root envelope isolation. Prior phase's test M1 is not rewritten.
No product adapter, exported Reader, canonical parsing or signature checks added.

Task 1: complete — query missing-interface RED observed, query GREEN 2.807s;
store compile-only PASS 1.763s. sqlc 1.31.1 offline generation repeated without
drift across all store files; only serving query and Querier changed.
Guard removal sensitivity detects root-fence and generator-bound omissions.
Initial sensitivity failure selected a different CASE; narrowed to the actual
generator CASE, without relaxing the required guard.
Ruling: quote '-modfile=tools/devtools/go.mod' in PowerShell — unquoted argument
arrived truncated at Go and generation never started — cost: none to contract.
Ruling: select nullable source timestamps directly instead of CASE + explicit
cast — sqlc inferred time.Time for the cast and pgtype.Timestamptz for the direct
source; this preserves NULL/infinity evidence — cost: invalid diagnostic rows
can carry timestamps, but no artifact bytes and no facts are returned.

## Initial generated file hashes

```json
[{"Name":"db.go","Hash":"770DFABB639E487BD985C9942ADC0D854E84ED06A6FD30348AB1EB33081BC403"},{"Name":"deviceauth_integration_test.go","Hash":"A7DC6B0A7B8A2320C133BA260CA6C8F7EC69D3385C4509E2A188146F4A6DDC3B"},{"Name":"deviceauth_task13_support_test.go","Hash":"C05219427C3B0AB34D49BF156C8EC3A32B2BAEEC96A7DB9A854EC303161DBBC1"},{"Name":"deviceauth.sql.go","Hash":"8C78115299569AD56041E24B31BEB153E674994578EFF3CD7BCFF786A97CA899"},{"Name":"idempotency.sql.go","Hash":"7F68EB7F171ED18B36C8BC53048D083ADBD3508BD175F328CF8F47D872906FC9"},{"Name":"identity_integration_test.go","Hash":"F2327D71E1C89F7EB238A204D9E9CFFA023256FF89EB720E7A0C97F397D1198F"},{"Name":"identity_session_integration_test.go","Hash":"72F1D2BB7E9ECA331AE1B67CBD9CF4F53F2C40FD5CB26261B7B2DAFF3B278A1A"},{"Name":"identity_session_support_test.go","Hash":"3BE8DBA9D25C01ED9788869B1D27CFCA413130CE0F7AC2D2C837CA2E6B6C5D42"},{"Name":"identity_task9_support_integration_test.go","Hash":"D547DB0A5FC09E886B2756BF4875C7ABF959A9D4B4A61DCE654535E2EDC6AF63"},{"Name":"identity_task9_support_test.go","Hash":"0224DA3DA74295C3B12F875B113073DF7A09794389A8C78AD8998D30EC28DC49"},{"Name":"identity.sql.go","Hash":"37AD835F18BA6072A8B66CD082C0D1D568F21DDB81A9156920278B3F1E765121"},{"Name":"json_equivalence_test.go","Hash":"5AE9635A623F9F897261EC9D402729A7AE20DDF3217D37C32069B36D77B1EBAF"},{"Name":"models.go","Hash":"8C3A4578C6F1624488A3A273BC192AD6D7A3B7A8CDB509196E7775F453F549F4"},{"Name":"nodecontrol_authority.sql.go","Hash":"4EF133D799C703DC5D86383E3697A460CEF03ABCB40EEB48169D88499C9E3935"},{"Name":"nodecontrol_schema_integration_test.go","Hash":"6E6145096DA96A32ED9981476AB98C32ECECC852182BF67BCD28986D1222EC70"},{"Name":"nodecontrol_schema_manifest_test.go","Hash":"67A532BC2B3C78A563DADDCABFD3D59BE93F1DB8599900F73924EDC6D875AF4F"},{"Name":"nodecontrol_schema_test.go","Hash":"D8037813CED83841CCA65A1B2EA71C788993E85B585600CBA9C9DAEB987774BD"},{"Name":"nodecontrol_serving.sql.go","Hash":"AF5D164980E9847C55561BDD713497A3F8D9BF4FF8073342A024CE6B7C9F31BC"},{"Name":"nodecontrol_v7_acl_integration_test.go","Hash":"911B6311297D87083D54B7D7A4D64C780E879642BC969F12B8989EE8CEC2C53A"},{"Name":"nodecontrol_v7_catalog_test.go","Hash":"D5BBCDDA94D6B2998893EF4B87D7C47B82E9A69D39D6237295E42EF183C2E47B"},{"Name":"nodecontrol_v7_down_integration_test.go","Hash":"EE1D136F4D788A5C45423E3FDED9F32ED2ACE6C0D9A7836FEA74F716BAB30B5E"},{"Name":"nodecontrol_v7_migration_integration_test.go","Hash":"7C9AF47421EA0DEC95115001A66A16FAC25E8EB8155A02A3DC50AC53C78BC6D7"},{"Name":"outbox.sql.go","Hash":"956CB162D647F9226F71E341376A15C7FC66A9760DBE37D7CB98F7B95746E38D"},{"Name":"querier.go","Hash":"966AA32E417EC67D52B7D985A3CF4EF1EBD25A226182839116336EA7A0BB63EF"},{"Name":"system_metadata_integration_test.go","Hash":"6C4073DBDD928E64E7A553B4E6714FEB61F44CECDD30B7071F922C1EEE53023A"},{"Name":"system_metadata.sql.go","Hash":"3E598D1DF5591785293A5DECA9B4A7A1B0668A4BE2C9FB79EF3D4E3D26E29699"},{"Name":"task18_delivery_support_test.go","Hash":"94483A7844B34B8CDDE18F081F79479D16484C6A09E85F695DF0DB5CC45EA070"},{"Name":"trust_delivery_integration_test.go","Hash":"FA60E5B25E761C7A610FC891C135B3D75C9A6C2ACC17F35DC8DCDD15756CDAD6"},{"Name":"trust_task16_support_test.go","Hash":"4F8E33B77C3C34C86BFD1402FEA170CC5DB007AA19C7928ED658BF0697CD96C8"},{"Name":"trust.sql.go","Hash":"E04477F935C400C31DF3DEF8F163BED5B4A8A4FA8FEF65A9B393B82285C056CD"}]
```
