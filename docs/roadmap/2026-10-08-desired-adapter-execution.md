# SDD ledger — plan: docs/superpowers/plans/2026-10-08-task10-desired-adapter.md

Execution date 2026-10-09 (Asia/Tbilisi). User approved Native execution.
Tasks 1–3 complete locally; uncommitted/unpushed. HEAD 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.
Existing linked worktree, codex/c12-b01-task9-coordinator, not a submodule.
Prior dirty files preserved. No new AGENTS.md or existing adapter implementation found.
Pre-flight: Task 1 queryDesiredProjection signature matches Task 2 consumer;
Task 3 consumes verification evidence. No shared-interface conflict.
Ruling: use portable ledger and PowerShell, retain local work rather than Bash
SDD scripts/commits/cleanup; approved no-commit/offline scope overrides skill
full-suite/download defaults — cost: manual progress record and unproven full acceptance.

## Task 1 evidence

- Full serving baseline PASS 4.220s. Missing-symbol RED, then fail-closed adapter behavior RED 4.285s.
- Cleanup/Panic/RowLimit RED 2.283s; implementation GREEN 4.075s.
- Late cancellation with converter domain error RED 4.464s: cancellation was checked only on successful conversion. Added the post-converter context check before returning domain errors; adapter GREEN 4.054s.
- Lifetime defensive Close-panic/cancel/success and repeated-cleanup cases GREEN 3.992s.
- Temporarily removing decoded byte clearing made Cleanup and all four Panic cases RED 3.462s (retained decoded material). Restored immediately; full serving GREEN 3.523s.
- Ruling: private desiredQueryLifetime.cleanup/finish own the undelivered projection and allow direct defensive lifecycle testing without production hooks. Main adapter signature unchanged. Cost: a small private owner type; tests do not pretend generated pgx ordinarily closes after conversion.
- Scope golden digests independently checked with .NET SHA256 and canonical UUID hex bytes, not mixed-endian Guid bytes.

## Task 2 evidence

- Missing reader symbols RED; fail-closed stub then behavior RED 4.220s for canceled/node preflight and successful BindingOrder.
- Private entry fixes BoundReadEpoch; seam resolves original callback DBTX before the adapter wrapper. No public capability added.
- Initial reader GREEN 4.474s; complete reader tests GREEN 4.593s.
- No-op discard mutation (test composition only) RED 3.998s on reject/cancel/panic, with retained staged projection. Immediately restored. Full serving GREEN 2.823s after formatting.
- Tests retain owned byte aliases, chain container and the real guard-owned projection pointer, verifying fixed signature/coordinates and all payloads disappear on failure; successful output remains intact.
- Evidence boundary: resolver seam proves guard/adapter/generated query/converter composition, not live Coordinator connection identity. Real fixed entry rejects unbound DBTX; authority accessor's separate tests cover exact private wrapper identity and expiry.

## Task 3 evidence and final review

- Fresh scoped regression PASS (all exit 0): serving 4.371s, contracts 4.840s, readiness 2.544s; authority `^(TestBoundRead|TestCoordinator)` 3.437s; store `-run '^$'` 3.196s, explicitly no tests run.
- All 30 internal/store filenames and SHA256 hashes match the initial inventory. SQL and four lock hashes also match. gofmt -l on five new Go files empty; Git tracked diff --check empty. Invocation search finds only the private entry/composition and test callers, no production entry.
- One fresh independent read-only reviewer completed (gpt-6-astra, high; review_desired_adapter_final): no Critical/Important or new Minor findings. No repair pass or second reviewer needed. Read all five new Go files, full spec/plan/ledger, guard, projection/converter, bound accessor; inspected generated method/types, DBTX, fixture/helpers and callers. This is not whole-branch review. Reviewer ran no tests; author supplies fresh verification evidence above.
- New-file whitespace/conflict checks PASS. Existing dirty files preserved. No commit/push, DB/Docker, full-suite/race, security-setting changes, dependency downloads or generation performed.

Task 1: complete (local only; final full serving test exit 0, 3.523s).
Task 2: complete (local only; final full serving test exit 0, 2.823s).
Task 3: complete (local only; scoped regressions, independent review, unchanged hashes and documentation checks).

### Final exclusion rulings (reviewer's complete list)

1. Final: Ruling: live Coordinator/physical identity/PostgreSQL decoding and complete SQL/schema semantics remain unverified — transport/resolver tests prove only isolated composition; generated ordering was inspected — cost if mistaken: integration defects could escape; require real DB evidence before production.
2. Final: Ruling: PITR, writer history, cross-epoch recovery, bootstrap, C11/C12/B01/main, v7 proof/incarnation/lease and security acceptance remain open — this increment neither implements nor waives them — cost: broader project acceptance remains blocked.
3. Final: Ruling: canonical parsing, crypto, trusted time/current authorization, full signed envelope and network-size enforcement are not delivered — private structural staging must not be promoted to authenticated serving facts — cost if mistaken: unverified data could be trusted; production stays unwired.
4. Final: Ruling: cleanup covers accessible owned decode/staging storage and Scan count only — no driver-prefetch/allocation or unreachable copy/string/network-buffer erasure claim — cost: this is not a total-memory or secure-erasure guarantee.
5. Final: Ruling: async callbacks, retained/reused capabilities, concurrent mutation/wrapper use remain unsupported — synchronous private-source contract governs this increment — cost if violated: races/lifetime misuse; no race or concurrent-use acceptance.
6. Final: Ruling: wrapper is single-query generated-code-specific, not a general DB decorator — valid field pointers and exactly one Query are fixed by the real invocation — cost: future reuse requires redesign and tests.
7. Final: Ruling: process termination and multiple simultaneous panics are excluded — defer proves ordinary cleanup/single-panic propagation, not process survival — cost: no universal cleanup/panic-preservation guarantee.
8. Final: Ruling: post-conversion Close panic is tested through the private lifetime method as a defensive branch — generated code normally closes before conversion — cost: helper evidence is not a natural live-query scenario.
9. Final: Ruling: reviewer did not rerun tests/mutations/hash checks — author observed and recorded those runs; final docs are author-maintained — cost: verification is independently code-reviewed but not independently re-executed.
10. Final: Ruling: review covers this increment and named dependencies, not historical branch/prior unrelated edits — only scoped local completion is justified — cost: cannot infer main/whole-branch acceptance.

Final: minor (deferred): older certificate fixture reuses UUIDs, weakening future role-swap detection (present mapping was previously verified).
Final: minor (deferred): older desired projection ownership test double-XOR/shared-buffer case and explicit envelope-alias checks need strengthening; current adapter assertions do not count as editing/closing it.

### Handoff

Keep this worktree and ledger locally as explicitly approved; no staging, commit,
push, merge or worktree cleanup. Next development gate is canonical-envelope/
signature design, not a public reader or production enablement. Full Task 10 and
security/integration gates remain open. The execution skill drove TDD, the
portable ledger and one final independent review; the approved no-publication
scope overrides default branch integration and workspace deletion.


## Baseline lock hashes
go.mod A0883514AF23F6947A8CF90F52B20D56AF5FC51BCECEEC2DD7C98082DF59E9E0
go.sum 642D51BD341B270B35EC976F6196152A33057ED38E54FBCEBF143A50271733D7
tools/devtools/go.mod 0A8312541FEC1A9FDC6555DB05D1F1D1342EE523B3273A800347714C146A71B9
tools/devtools/go.sum 85A30247BB0F2A8568D8081A31F029CFA8FEA89C9F2CFE3FF17EEDD2A0340164
SQL 53E781075556FCCE10385E868369FF7F2C52107B304955D0E8855CBDCB7333EB
generated query 396E72D42F67B1F19A98D4E67ED0FD4857AB4C8315283DDB5CED5663F286B040
Querier 2BE2E4D8702A23D2EFFFA9DD816F10EB188996A3C6B27D9FD3FFCF1D2FBCE8D2

All initial store hashes:
```json
[{"Name":"db.go","Hash":"770DFABB639E487BD985C9942ADC0D854E84ED06A6FD30348AB1EB33081BC403"},{"Name":"deviceauth_integration_test.go","Hash":"A7DC6B0A7B8A2320C133BA260CA6C8F7EC69D3385C4509E2A188146F4A6DDC3B"},{"Name":"deviceauth_task13_support_test.go","Hash":"C05219427C3B0AB34D49BF156C8EC3A32B2BAEEC96A7DB9A854EC303161DBBC1"},{"Name":"deviceauth.sql.go","Hash":"8C78115299569AD56041E24B31BEB153E674994578EFF3CD7BCFF786A97CA899"},{"Name":"idempotency.sql.go","Hash":"7F68EB7F171ED18B36C8BC53048D083ADBD3508BD175F328CF8F47D872906FC9"},{"Name":"identity_integration_test.go","Hash":"F2327D71E1C89F7EB238A204D9E9CFFA023256FF89EB720E7A0C97F397D1198F"},{"Name":"identity_session_integration_test.go","Hash":"72F1D2BB7E9ECA331AE1B67CBD9CF4F53F2C40FD5CB26261B7B2DAFF3B278A1A"},{"Name":"identity_session_support_test.go","Hash":"3BE8DBA9D25C01ED9788869B1D27CFCA413130CE0F7AC2D2C837CA2E6B6C5D42"},{"Name":"identity_task9_support_integration_test.go","Hash":"D547DB0A5FC09E886B2756BF4875C7ABF959A9D4B4A61DCE654535E2EDC6AF63"},{"Name":"identity_task9_support_test.go","Hash":"0224DA3DA74295C3B12F875B113073DF7A09794389A8C78AD8998D30EC28DC49"},{"Name":"identity.sql.go","Hash":"37AD835F18BA6072A8B66CD082C0D1D568F21DDB81A9156920278B3F1E765121"},{"Name":"json_equivalence_test.go","Hash":"5AE9635A623F9F897261EC9D402729A7AE20DDF3217D37C32069B36D77B1EBAF"},{"Name":"models.go","Hash":"8C3A4578C6F1624488A3A273BC192AD6D7A3B7A8CDB509196E7775F453F549F4"},{"Name":"nodecontrol_authority.sql.go","Hash":"4EF133D799C703DC5D86383E3697A460CEF03ABCB40EEB48169D88499C9E3935"},{"Name":"nodecontrol_schema_integration_test.go","Hash":"6E6145096DA96A32ED9981476AB98C32ECECC852182BF67BCD28986D1222EC70"},{"Name":"nodecontrol_schema_manifest_test.go","Hash":"67A532BC2B3C78A563DADDCABFD3D59BE93F1DB8599900F73924EDC6D875AF4F"},{"Name":"nodecontrol_schema_test.go","Hash":"D8037813CED83841CCA65A1B2EA71C788993E85B585600CBA9C9DAEB987774BD"},{"Name":"nodecontrol_serving.sql.go","Hash":"396E72D42F67B1F19A98D4E67ED0FD4857AB4C8315283DDB5CED5663F286B040"},{"Name":"nodecontrol_v7_acl_integration_test.go","Hash":"911B6311297D87083D54B7D7A4D64C780E879642BC969F12B8989EE8CEC2C53A"},{"Name":"nodecontrol_v7_catalog_test.go","Hash":"D5BBCDDA94D6B2998893EF4B87D7C47B82E9A69D39D6237295E42EF183C2E47B"},{"Name":"nodecontrol_v7_down_integration_test.go","Hash":"EE1D136F4D788A5C45423E3FDED9F32ED2ACE6C0D9A7836FEA74F716BAB30B5E"},{"Name":"nodecontrol_v7_migration_integration_test.go","Hash":"7C9AF47421EA0DEC95115001A66A16FAC25E8EB8155A02A3DC50AC53C78BC6D7"},{"Name":"outbox.sql.go","Hash":"956CB162D647F9226F71E341376A15C7FC66A9760DBE37D7CB98F7B95746E38D"},{"Name":"querier.go","Hash":"2BE2E4D8702A23D2EFFFA9DD816F10EB188996A3C6B27D9FD3FFCF1D2FBCE8D2"},{"Name":"system_metadata_integration_test.go","Hash":"6C4073DBDD928E64E7A553B4E6714FEB61F44CECDD30B7071F922C1EEE53023A"},{"Name":"system_metadata.sql.go","Hash":"3E598D1DF5591785293A5DECA9B4A7A1B0668A4BE2C9FB79EF3D4E3D26E29699"},{"Name":"task18_delivery_support_test.go","Hash":"94483A7844B34B8CDDE18F081F79479D16484C6A09E85F695DF0DB5CC45EA070"},{"Name":"trust_delivery_integration_test.go","Hash":"FA60E5B25E761C7A610FC891C135B3D75C9A6C2ACC17F35DC8DCDD15756CDAD6"},{"Name":"trust_task16_support_test.go","Hash":"4F8E33B77C3C34C86BFD1402FEA170CC5DB007AA19C7928ED658BF0697CD96C8"},{"Name":"trust.sql.go","Hash":"E04477F935C400C31DF3DEF8F163BED5B4A8A4FA8FEF65A9B393B82285C056CD"}]
```
