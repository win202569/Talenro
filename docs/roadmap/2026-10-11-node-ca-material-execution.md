# Node CA Material — local execution record (2026-10-11 approval; local closeout 2026-10-12)

## Status and authorization

User approved design (“是”) and the five-task plan (“1”, Native). Implemented locally in `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`; branch `codex/c12-b01-task9-coordinator`. This is a partial-material increment, not production trust or whole-project acceptance. Approval/increment filenames retain 2026-10-11; actual UTC receipts map to local Oct 12 in UTC+4. No Git configuration/index/HEAD writes, staging, commits, pushes, merges, dependency downloads or cleanup authorized or performed.

24 new Go files:12 product/12 tests under `internal/nodecontrol/nodeca/`; no product consumer or old source/test/testdata/config/interface change. The only allowed old document edits are spec/plan state and checklists, and a new top current-status section preserving the prior body. One new execution document is allowed.

Spec: [approved design](../superpowers/specs/2026-10-11-node-ca-material-design.md).
Plan: [approved implementation plan](../superpowers/plans/2026-10-11-node-ca-material.md).
Canonical local evidence: `.superpowers/sdd/2026-10-11-node-ca-material/` (existing ignore rule verified; preserved, not a remote recovery checkpoint).

## Delivered contract and retained obligations

`Parse(context.Context, Input)` accepts bounded owned DER and caller-declared CA subject SPKI; exact canonical P-256 representation, minimal, completely consumed known X.509 v3 raw structure, positive serial, canonical UTC time interval, bounded nonempty names, exact four distinct BC/KU/SKI/AKI extensions, full32-byte SHA256 raw-SPKI SKI binding. Successful detached Summary preserves pathLen absent versus zero, CRL-sign allowance and four raw-byte digests. Shared-owner Clear invalidates all wrappers and erases both owned buffers; no raw-byte output, Verify/Ready/Authorize/serialization/install API.

`NodeCAMaterialShapeAndDeclaredKeyBound` always retains the fixed six Pending items in order:
1. CertificateSignatureAndIssuerChain
2. AuthorityIdentifierAndParentBinding
3. TrustedCurrentTime
4. DeploymentAuthorizationAndPurpose
5. BundleStatusAndRollback
6. LeafIssuerRelationAndRuntimeUse

Unknown signature algorithm, bad signature bytes, arbitrary shape-valid AKI, expired/future CA or equal subject/issuer do not create trust. Declared SPKI is subject key, not parent key. Old nodeleaf/nodeleafkeys proof grades/Pending remain unchanged; old CA/provider compatibility UNKNOWN.

Input limits DER65,536/SPKI8,192/aggregate73,728 bytes are input budgets, not heap bounds. Failed private-owner erasure is source-supported; successful alias cleanup is observed. No concurrent input mutation/Clear/read, fatal-OOM/crash erasure or hard-real-time cancellation assurance.

## Environment

Pinned cached native Windows Go1.26.5, PowerShell only. No dependency/network downloads; modules readonly. The original persisted Go target was Linux on a Windows host: attempt0002 failed before tests started, attempt0003 diagnosed it. Corrected process-only target below, no persistent Go configuration change. Inherited GODEBUG was empty/null and is recorded rather than overwritten.

103 retained original wrapper receipts include the complete wrapper, exact command/workdir, actual start/end UTC, full output, exit and before/after source/spec/plan hashes. Attempt0094 is the sole exception: an author collection-helper failure lost its completed output; unknown UTC/exit/hashes remain explicitly unknown, not reconstructed or counted as RED/GREEN. Attempt0095 independently retains genuine regression RED. The common wrapper is:

```powershell
$env:GOOS='windows'
$env:GOARCH='amd64'
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$env:GOSUMDB='off'
$env:GOAUTH='off'
$env:GOVCS='all:off'
$env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'
$env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-node-ca-material/go-cache'
$caGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$caFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

The original receipts preserve original failures; none replaced by a passing retry or reconstructed time. A read-only metadata summarization initially exceeded output budget; reduced-field rerun succeeded. This was not a test attempt or passing verification and did not alter the receipts.

## Task and review repair evidence

- Task1:0001 genuine missing-contract RED →0004 native GREEN,0005 full suite. 0002 platform-executable failure is not GREEN;0003 target diagnostic.
- Task2:0006 one-time temporary test-only signed golden generation with fixed test keys, then replaced with immutable inline bytes. 0007 independent PowerShell SHA256 oracles. 0008 syntax error is not RED;0009 missing-helper RED →0010 GREEN,0011 full suite.
- Task3:0012 RED →0013 GREEN,0014 full suite. Field/name/serial/time/pathLen and all extension-permutation cases ran.
- Task4:0015 RED →0016 GREEN,0017 full suite. Full15-field Summary, subject/parent roles, cancellation and fixed error priorities ran.
- Task5:0018 missing-guard RED;0019–0021 genuine guard false positives diagnosed;0022 GREEN. Supplemental0023 RED showed an unapproved depth32 rejection and premature ownership-handoff guard miss;0024 GREEN with iterative bounded-input DER traversal and last-clone handoff guard;0025 guards GREEN. Both fuzz entries were present for the file-inventory RED.
- First format0026/check0027 and final0028–0032 passed on revision1, but are superseded historical evidence, not completion evidence for revision8.
- First fresh read-only gpt-6-astra/xhigh review:0 Critical/3 Important/1 Minor. Three confirmed Important defects: DN RDN SET ordering, invalid universal tag forms in unknown signer parameters, BC malformed tail misclassified as unsupported. Attempt0033 reproduces all three;0034 targeted GREEN after fixes. New RFC7093 full32-byte Summary SPKI-digest/SKI literal check already passed0033: missing coverage closed, not a product RED.
- Revision2 exact24 format0035/check0036 and final0037–0041 passed, but second fresh review found0 Critical/2 Important/0 Minor: valid opaque SET incorrectly treated as SET OF, and malformed nested AKI misclassified as Unsupported.0042 reproduced both.0043 unknown-SET fix GREEN while known DN SET OF order stays checked.0044 failed on an early new-package malformed-AKI test oracle (not GREEN); original malformed constructed[0] EOCzero input retained as InvalidMaterial, well-formed nested OCTET STRING Unsupported control added.0045 targeted GREEN. Revision3 format0046/check0047 and final0048–0052 passed, but third fresh review found0 Critical/3 Important/0 Minor: incomplete nonP256 declared-SPKI consumption, UID early rejection hides malformed UID/TBS, and known implicit AKI serial minimality omitted.0053 genuine RED;0054 SPKI/0055 UID/0056 serial GREEN, with all valid unsupported controls and prior regressions retained. Revision4 exact24 format0057/check0058 and final0059–0063 all exit0 on frozen-sources-4.json. Those revision4 passes are superseded historical evidence. Fourth fresh review found 0 Critical/2 Important/0 Minor: known RSA key-payload tails and AKI issuer registeredID implicit OID minimality. Attempt0064 genuine RED also covers DSA payload/parameter tails and malformed OID controls;0065 focused GREEN. Exact Go re-encoding now precedes unsupported-key return; explicit DSA grammar consumption preserves valid unsupported DSA without adding algorithm support. RegisteredID content checks minimally encoded, terminated base-128 groups without numeric-size policy. Added shared CA-wire malformed-sentinel controls, then revision5 format0066/check0067 and all five commands0068–0072 pass on frozen-sources-5.json. Those revision5 passes are historical. Fifth fresh review found0 Critical/2 Important/0 Minor: incomplete known AKI GeneralNames grammar and a bounded-Go-int OID cap on valid opaque parameters.0073 reproduced all supplied GeneralNames/large-OID cases;0074 added known X400 attribute-family encoding/container REDs; valid excluded controls already passed.0075 focused GREEN after IA5/Name/otherName/EDI/X400 grammar and representation-only opaque OID repairs. Only der.go, extensions.go, raw_certificate.go and parse_test.go changed from freeze5. Exact24-source formatting0076/check0077 and final0078–0082 passed on freeze6. Those revision6 passes are now historical. Sixth fresh whole-increment review found0 Critical/2 Important/0 Minor: rejected directoryName AttributeValue bypassed pinned Name interpretation, and extension ANY primitive/required-container grammar was insufficient.0083 reproduced these and related REAL/PDV/CHARACTER STRING grammar omissions but one new valid EXTERNAL-reference control used length0d for11bytes: that fixture failure is not a product defect or GREEN. Corrected only the new literal to0b;0084 obtained clean genuine assertion RED with valid controls already passing.0085/0086 focused GREEN after pinned six Name string types and extension-only VisibleString/RelativeOID/time/REAL/EXTERNAL/PDV/CHARACTER STRING grammar repairs. Added all six identification-choice and unbounded REAL exponent controls. Only raw_certificate.go/extensions.go/parse_test.go changed from freeze6; no new accepted algorithm/profile/Pending/consumer/API/dependency.0087 exact24format,0088 emptyformatcheck,0089–0093 finalfive commands pass onfreeze7. Those revision7 passes were superseded after its fresh review found the two defects described next.


- Review7:0 Critical/2 Important/0 Minor. Universal TIME14 invalid UTF8/notation was not checked, and a valid generic GeneralizedTime leap second inherited Go's false-invalid limitation.0094 collection diagnostic has unavailable outcome;0095 genuine assertion RED →0096 GREEN and0097 expanded/calendar/precision/duration/interval/boundary/caller-byte-preservation controls GREEN. Canonical TIME uses the ASCII tstring subset and selected envelope/calendar/DER transformations without a numeric year/fraction/duration/recurrence cap. Generic second60 is calendar-checked using59 only in a bounded cleared temporary; input stays exact; UTC/certificate validity front-parser rejection is retained. No actual leap-event scheduling, chronology or vendor subtype conformance is established.

## Fresh final verification (revision8)

```powershell
& $caGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
& $caGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
& $caGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
& $caGo test ./internal/nodecontrol/nodeca -run '^$' -fuzz '^FuzzCADER$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
& $caGo test ./internal/nodecontrol/nodeca -run '^$' -fuzz '^FuzzCASPKI$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
```

| Receipt | Command scope | Result |
| --- | --- | --- |
| 0099 | Exact24-file gofmt check | exit0,empty output |
| 0100 | nodeca full package | exit0,3.583s package time |
| 0101 | unchanged nodeleaf regression | exit0,3.664s |
| 0102 | unchanged nodeleafkeys regression | exit0,2.737s |
| 0103 | FuzzCADER5s/one worker | exit0;5 seeds,36 executions,0newinteresting,no reported crash |
| 0104 | FuzzCASPKI5s/one worker | exit0;65 cached seeds,4963 executions,1newinteresting,total66,no reported crash |

Ordinary tests use `-count=1 -p=1 -timeout=120s`; fuzz uses fixed5s/30s parameters. No timeout increases or failure exclusions. Process-start/overall wall-clock delays were observed from0053 onward; cause not established. Test bodies finished without timeout. A late read-only process snapshot returned no matching process/exit1, not passing evidence; missing UTC for that separate diagnostic was not reconstructed. Very low DER fuzz execution count is limited exploration, not exhaustive correctness evidence.

## All retained attempts

Local `attempt-NNNN.md` files retain the complete original receipts except0094, whose separate author-constructed collection-failure diagnostic explicitly records unavailable output/UTC/exit/hash. Numbering0001–0104 is continuous:103 original wrapper receipts plus that one diagnostic, NOT104 complete original receipts. Exit1 rows preserve original failures; platform/syntax/oracle/guard distinctions are documented above. Historical labels such as “GREEN” at0002 describe an intended invocation, not actual acceptance.0094 has no classified outcome;0095 is the retained genuine RED.

| Attempt | Recorded purpose | Actual start UTC | Actual end UTC | Exit |
| --- | --- | --- | --- | --- |
| 0001 | # Attempt 1 — Task 1 RED | 2026-10-11T18:05:55.8821474Z | 2026-10-11T18:06:59.7154164Z | 1 |
| 0002 | # Attempt 2 — Task 1 GREEN | 2026-10-11T18:07:26.3131788Z | 2026-10-11T18:07:54.1549828Z | 1 |
| 0003 | # Attempt 3 — Task 1 diagnostic: effective Go target | 2026-10-11T18:08:11.1169382Z | 2026-10-11T18:08:11.1775894Z | 0 |
| 0004 | # Attempt 4 — Task 1 GREEN native target | 2026-10-11T18:08:23.4965688Z | 2026-10-11T18:09:03.1906429Z | 0 |
| 0005 | # Attempt 5 — Task 1 complete suite | 2026-10-11T18:09:07.8034212Z | 2026-10-11T18:09:10.5460052Z | 0 |
| 0006 | # Attempt 6 — Task 2 one-time fixture generation (not RED/GREEN) | 2026-10-11T18:09:42.5808323Z | 2026-10-11T18:09:55.119212Z | 0 |
| 0007 | # Attempt 7 — Task 2 independent PowerShell SHA256 oracle | 2026-10-11T18:10:28.789633Z | 2026-10-11T18:10:28.8473414Z | 0 |
| 0008 | # Attempt 8 — Task 2 RED | 2026-10-11T18:11:37.4807846Z | 2026-10-11T18:11:37.7289993Z | 1 |
| 0009 | # Attempt 9 — Task 2 RED after test syntax repair | 2026-10-11T18:11:46.9421263Z | 2026-10-11T18:11:47.7387837Z | 1 |
| 0010 | # Attempt 10 — Task 2 GREEN | 2026-10-11T18:12:42.7781118Z | 2026-10-11T18:12:46.370349Z | 0 |
| 0011 | # Attempt 11 — Task 2 complete suite | 2026-10-11T18:13:18.9260237Z | 2026-10-11T18:13:22.2384034Z | 0 |
| 0012 | # Attempt 12 — Task 3 RED | 2026-10-11T18:14:31.6776623Z | 2026-10-11T18:14:32.4612016Z | 1 |
| 0013 | # Attempt 13 — Task 3 GREEN | 2026-10-11T18:15:14.5817663Z | 2026-10-11T18:15:18.2971833Z | 0 |
| 0014 | # Attempt 14 — Task 3 complete suite | 2026-10-11T18:15:41.5698445Z | 2026-10-11T18:15:45.0197994Z | 0 |
| 0015 | # Attempt 15 — Task 4 RED | 2026-10-11T18:16:46.3183301Z | 2026-10-11T18:16:47.1313789Z | 1 |
| 0016 | # Attempt 16 — Task 4 GREEN | 2026-10-11T18:17:10.6175407Z | 2026-10-11T18:17:14.2335495Z | 0 |
| 0017 | # Attempt 17 — Task 4 complete suite | 2026-10-11T18:17:37.857689Z | 2026-10-11T18:17:41.2532957Z | 0 |
| 0018 | # Attempt 18 — Task 5 guard RED | 2026-10-11T18:19:22.2715992Z | 2026-10-11T18:19:23.3175269Z | 1 |
| 0019 | # Attempt 19 — Task 5 guard GREEN | 2026-10-11T18:21:07.4577619Z | 2026-10-11T18:21:12.5223214Z | 1 |
| 0020 | # Attempt 20 — Task 5 guard GREEN after false-positive repair | 2026-10-11T18:21:28.5174046Z | 2026-10-11T18:21:32.4406991Z | 1 |
| 0021 | # Attempt 21 — Task 5 precise time-chain RED | 2026-10-11T18:21:46.9707684Z | 2026-10-11T18:21:50.9272792Z | 1 |
| 0022 | # Attempt 22 — Task 5 guard GREEN time-chain regression | 2026-10-11T18:21:57.9147704Z | 2026-10-11T18:22:01.9209669Z | 0 |
| 0023 | # Attempt 23 — Task 5 supplemental RED: legal deep unknown parameters and premature clone handoff | 2026-10-11T18:23:02.737034Z | 2026-10-11T18:23:06.6200164Z | 1 |
| 0024 | # Attempt 24 — Task 5 supplemental GREEN | 2026-10-11T18:23:39.7039463Z | 2026-10-11T18:23:43.5415049Z | 0 |
| 0025 | # Attempt 25 — Task 5 guard final GREEN before format | 2026-10-11T18:23:56.474435Z | 2026-10-11T18:24:00.1769501Z | 0 |
| 0026 | # Attempt 26 — Task 5 format exactly 24 new Go files | 2026-10-11T18:24:01.7121258Z | 2026-10-11T18:24:01.8333184Z | 0 |
| 0027 | # Attempt 27 — Task 5 format check: empty output required | 2026-10-11T18:24:03.4553402Z | 2026-10-11T18:24:03.5108336Z | 0 |
| 0028 | # Attempt 28 — Final 1/5 nodeca full suite | 2026-10-11T18:24:04.8044596Z | 2026-10-11T18:24:08.6816249Z | 0 |
| 0029 | # Attempt 29 — Final 2/5 nodeleaf regression | 2026-10-11T18:24:12.3725151Z | 2026-10-11T18:24:16.5692946Z | 0 |
| 0030 | # Attempt 30 — Final 3/5 nodeleafkeys regression | 2026-10-11T18:24:20.5212924Z | 2026-10-11T18:24:24.7365788Z | 0 |
| 0031 | # Attempt 31 — Final 4/5 FuzzCADER | 2026-10-11T18:24:26.2932494Z | 2026-10-11T18:25:31.6167563Z | 0 |
| 0032 | # Attempt 32 — Final 5/5 FuzzCASPKI | 2026-10-11T18:25:39.9230925Z | 2026-10-11T18:25:49.1383665Z | 0 |
| 0033 | # Attempt 33 — Review 1 counterexample RED and external-oracle coverage characterization | 2026-10-11T18:33:17.9709998Z | 2026-10-11T18:33:23.0743965Z | 1 |
| 0034 | # Attempt 34 — Review 1 GREEN: three fixes, RFC oracle and valid deep Unknown retained | 2026-10-11T18:34:08.1149934Z | 2026-10-11T18:34:13.292339Z | 0 |
| 0035 | # Attempt 35 — Revision 2 exact 24-file formatting | 2026-10-11T18:36:31.0187049Z | 2026-10-11T18:36:31.0688886Z | 0 |
| 0036 | # Attempt 36 — Revision 2 exact 24-file format check | 2026-10-11T18:36:36.4687048Z | 2026-10-11T18:36:36.512249Z | 0 |
| 0037 | # Attempt 37 — Revision 2 final nodeca suite | 2026-10-11T18:36:38.0563904Z | 2026-10-11T18:36:43.1897947Z | 0 |
| 0038 | # Attempt 38 — Revision 2 final nodeleaf suite | 2026-10-11T18:36:44.8421643Z | 2026-10-11T18:36:49.7260895Z | 0 |
| 0039 | # Attempt 39 — Revision 2 final nodeleafkeys suite | 2026-10-11T18:36:51.333816Z | 2026-10-11T18:36:56.1056095Z | 0 |
| 0040 | # Attempt 40 — Revision 2 final CA DER fuzz | 2026-10-11T18:37:00.9754438Z | 2026-10-11T18:37:12.9876426Z | 0 |
| 0041 | # Attempt 41 — Revision 2 final CA SPKI fuzz | 2026-10-11T18:37:20.2745695Z | 2026-10-11T18:37:31.0536579Z | 0 |
| 0042 | # Attempt 42 — Review 2 behavioral reproduction RED | 2026-10-11T18:46:48.9976443Z | 2026-10-11T18:46:53.8924022Z | 1 |
| 0043 | # Attempt 43 — Review 2 opaque SET fix GREEN with RDN/form regression | 2026-10-11T18:47:08.1683922Z | 2026-10-11T18:47:13.2055761Z | 0 |
| 0044 | # Attempt 44 — Review 2 AKI nested consumption fix GREEN | 2026-10-11T18:47:15.2349931Z | 2026-10-11T18:47:20.204398Z | 1 |
| 0045 | # Attempt 45 — Review 2 AKI fix GREEN after malformed/profile oracle correction | 2026-10-11T18:47:50.9581728Z | 2026-10-11T18:47:55.6805375Z | 0 |
| 0046 | # Attempt 46 — Revision 3 exact 24-file formatting | 2026-10-11T18:47:57.0442226Z | 2026-10-11T18:47:57.0949766Z | 0 |
| 0047 | # Attempt 47 — Revision 3 exact 24-file format check | 2026-10-11T18:47:58.5377444Z | 2026-10-11T18:47:58.5745844Z | 0 |
| 0048 | # Attempt 48 — Revision 3 final nodeca suite | 2026-10-11T18:48:03.9460973Z | 2026-10-11T18:48:08.6882355Z | 0 |
| 0049 | # Attempt 49 — Revision 3 final nodeleaf suite | 2026-10-11T18:48:10.0154737Z | 2026-10-11T18:48:14.426388Z | 0 |
| 0050 | # Attempt 50 — Revision 3 final nodeleafkeys suite | 2026-10-11T18:48:15.7137641Z | 2026-10-11T18:48:19.7649789Z | 0 |
| 0051 | # Attempt 51 — Revision 3 final CA DER fuzz | 2026-10-11T18:48:21.1101981Z | 2026-10-11T18:48:33.173557Z | 0 |
| 0052 | # Attempt 52 — Revision 3 final CA SPKI fuzz | 2026-10-11T18:48:34.6420728Z | 2026-10-11T18:48:45.3389063Z | 0 |
| 0053 | # Attempt 53 — Review 3 three malformed/profile reproductions RED | 2026-10-11T18:59:43.0677715Z | 2026-10-11T19:00:51.1660023Z | 1 |
| 0054 | # Attempt 54 — Review 3 SPKI consumption fix GREEN | 2026-10-11T19:01:21.9445523Z | 2026-10-11T19:02:14.0179185Z | 0 |
| 0055 | # Attempt 55 — Review 3 UID complete consumption fix GREEN | 2026-10-11T19:02:38.1728618Z | 2026-10-11T19:03:08.7475388Z | 0 |
| 0056 | # Attempt 56 — Review 3 known AKI implicit serial fix GREEN | 2026-10-11T19:03:24.2496357Z | 2026-10-11T19:03:54.6127574Z | 0 |
| 0057 | # Attempt 57 — Revision 4 exact 24-file formatting | 2026-10-11T19:04:08.2626637Z | 2026-10-11T19:04:08.7191792Z | 0 |
| 0058 | # Attempt 58 — Revision 4 exact 24-file format check | 2026-10-11T19:04:20.7578181Z | 2026-10-11T19:04:21.1542457Z | 0 |
| 0059 | # Attempt 59 — Revision 4 final nodeca suite | 2026-10-11T19:04:33.7430207Z | 2026-10-11T19:05:22.4744639Z | 0 |
| 0060 | # Attempt 60 — Revision 4 final nodeleaf suite | 2026-10-11T19:05:34.1271102Z | 2026-10-11T19:06:02.5249929Z | 0 |
| 0061 | # Attempt 61 — Revision 4 final nodeleafkeys suite | 2026-10-11T19:06:16.3751248Z | 2026-10-11T19:06:49.9132169Z | 0 |
| 0062 | # Attempt 62 — Revision 4 final CA DER fuzz | 2026-10-11T19:07:04.9136986Z | 2026-10-11T19:08:05.2613742Z | 0 |
| 0063 | # Attempt 63 — Revision 4 final CA SPKI fuzz | 2026-10-11T19:08:18.8172325Z | 2026-10-11T19:08:59.8473537Z | 0 |
| 0064 | Revision-4 review findings RED and known unsupported controls | 2026-10-11T19:29:17.9832407Z | 2026-10-11T19:30:05.7158042Z | 1 |
| 0065 | Revision-4 findings focused GREEN after known payload fixes | 2026-10-11T19:30:51.226603Z | 2026-10-11T19:31:35.3080461Z | 0 |
| 0066 | Revision-5 exact 24-source formatting | 2026-10-11T19:32:27.3040086Z | 2026-10-11T19:32:31.6754388Z | 0 |
| 0067 | Revision-5 exact formatting check (must be empty) | 2026-10-11T19:33:36.4938388Z | 2026-10-11T19:33:37.6449712Z | 0 |
| 0068 | Revision-5 final full nodeca package | 2026-10-11T19:33:54.1634756Z | 2026-10-11T19:35:07.800151Z | 0 |
| 0069 | Revision-5 final existing nodeleaf regression | 2026-10-11T19:35:27.5633796Z | 2026-10-11T19:36:29.0091823Z | 0 |
| 0070 | Revision-5 final existing nodeleafkeys regression | 2026-10-11T19:36:56.7998737Z | 2026-10-11T19:37:34.4210559Z | 0 |
| 0071 | Revision-5 final FuzzCADER bounded smoke | 2026-10-11T19:37:52.844546Z | 2026-10-11T19:38:55.642815Z | 0 |
| 0072 | Revision-5 final FuzzCASPKI bounded smoke | 2026-10-11T19:39:12.8056184Z | 2026-10-11T19:40:10.4086889Z | 0 |
| 0073 | Revision-5 full known GeneralNames and opaque OID findings RED | 2026-10-11T20:00:54.2963165Z | 2026-10-11T20:01:38.3325108Z | 1 |
| 0074 | GeneralNames grouped repair supplemental known X400 payload boundaries RED | 2026-10-11T20:03:15.7947636Z | 2026-10-11T20:03:47.4885177Z | 1 |
| 0075 | Revision-6 grouped GeneralName syntax and unbounded opaque OID focused GREEN | 2026-10-11T20:05:34.4991802Z | 2026-10-11T20:06:09.5173133Z | 0 |
| 0076 | Revision-6 exact 24-source formatting | 2026-10-11T20:07:50.5388897Z | 2026-10-11T20:07:51.418611Z | 0 |
| 0077 | Revision-6 exact formatting check (must be empty) | 2026-10-11T20:08:54.7887198Z | 2026-10-11T20:08:55.7425808Z | 0 |
| 0078 | Revision-6 final full nodeca package | 2026-10-11T20:09:42.2684186Z | 2026-10-11T20:10:31.3582068Z | 0 |
| 0079 | Revision-6 final full nodeleaf package | 2026-10-11T20:11:58.8100772Z | 2026-10-11T20:12:32.4654096Z | 0 |
| 0080 | Revision-6 final full nodeleafkeys package | 2026-10-11T20:12:44.9524403Z | 2026-10-11T20:13:19.8924237Z | 0 |
| 0081 | Revision-6 final FuzzCADER smoke | 2026-10-11T20:13:39.0234333Z | 2026-10-11T20:14:35.5159508Z | 0 |
| 0082 | Revision-6 final FuzzCASPKI smoke | 2026-10-11T20:14:47.4306325Z | 2026-10-11T20:15:40.1853802Z | 0 |
| 0083 | Review-6 Name and universal ANY grammar assertion RED | 2026-10-11T20:36:11.5421973Z | 2026-10-11T20:36:49.7982451Z | 1 |
| 0084 | Review-6 corrected controls assertion RED | 2026-10-11T20:37:15.5188701Z | 2026-10-11T20:37:49.6314747Z | 1 |
| 0085 | Review-6 focused Name and extension ANY fixes GREEN candidate | 2026-10-11T20:40:43.9177599Z | 2026-10-11T20:41:18.6617676Z | 0 |
| 0086 | Review-6 family-complete controls and unchanged isolation guards | 2026-10-11T20:42:11.3804208Z | 2026-10-11T20:42:44.6144934Z | 0 |
| 0087 | Revision-7 exact 24-source formatting | 2026-10-11T20:43:28.2973217Z | 2026-10-11T20:43:29.1343972Z | 0 |
| 0088 | Revision-7 exact formatting check (must be empty) | 2026-10-11T20:43:54.5168333Z | 2026-10-11T20:43:55.5343788Z | 0 |
| 0089 | Revision-7 final full nodeca package | 2026-10-11T20:44:27.2113144Z | 2026-10-11T20:45:08.1327647Z | 0 |
| 0090 | Revision-7 final existing nodeleaf regression | 2026-10-11T20:45:23.6803726Z | 2026-10-11T20:46:18.0832091Z | 0 |
| 0091 | Revision-7 final existing nodeleafkeys regression | 2026-10-11T20:46:34.4667668Z | 2026-10-11T20:47:35.3474985Z | 0 |
| 0092 | Revision-7 final FuzzCADER bounded smoke | 2026-10-11T20:47:53.9691643Z | 2026-10-11T20:49:02.4105809Z | 0 |
| 0093 | Revision-7 final FuzzCASPKI bounded smoke | 2026-10-11T20:52:02.9859738Z | 2026-10-11T20:52:44.1034848Z | 0 |
| 0094 | collection failure; outcome unavailable | unavailable | unavailable | unknown (collection loss) |
| 0095 | review-7 TIME/leap fresh regression RED before production implementation; 0094 unavailable | 2026-10-11T21:17:59.7745506Z | 2026-10-11T21:18:51.4819247Z | 1 |
| 0096 | review-7 TIME/leap repair focused GREEN and boundaries | 2026-10-11T21:20:34.9416696Z | 2026-10-11T21:21:35.919981Z | 0 |
| 0097 | review-7 expanded TIME edge controls, certificate validity front-parser boundary and caller-byte preservation GREEN | 2026-10-11T21:23:09.4035487Z | 2026-10-11T21:23:51.8927396Z | 0 |
| 0098 | revision-8 format exactly 24 increment files | 2026-10-11T21:24:30.234823Z | 2026-10-11T21:24:30.8256009Z | 0 |
| 0099 | revision-8 exact 24-file formatting check | 2026-10-11T21:24:53.8541091Z | 2026-10-11T21:24:54.243036Z | 0 |
| 0100 | revision-8 final approved command 1/5 | 2026-10-11T21:25:07.1815197Z | 2026-10-11T21:25:38.0574084Z | 0 |
| 0101 | revision-8 final approved command 2/5 | 2026-10-11T21:25:48.7160388Z | 2026-10-11T21:26:18.5140571Z | 0 |
| 0102 | revision-8 final approved command 3/5 | 2026-10-11T21:26:29.5233236Z | 2026-10-11T21:26:53.067873Z | 0 |
| 0103 | revision-8 final approved command 4/5 | 2026-10-11T21:27:05.1867799Z | 2026-10-11T21:27:59.1132963Z | 0 |
| 0104 | revision-8 final approved command 5/5 | 2026-10-11T21:28:10.6628778Z | 2026-10-11T21:28:41.1446069Z | 0 |

## Frozen current source hashes

Revision8 manifest actualUTC: 2026-10-11T21:29:15.5858393Z. Exact24 current files match0099–0104 before/after hashes and0098 posthashes; candidate-8-consistency.json verified currentSourceDrift0 at 2026-10-11T21:30:42.8928597Z. All five final processes ended before this freeze read. From freeze7 only extensions.go and parse_test.go differ. Any later source edit invalidates this acceptance/review evidence.

| File | SHA256 |
| --- | --- |
| boundary_test.go | F644256710F4BE7302017A1D375C4BB5E5926E0102A2CE7C2FDAE23893FBD05B |
| budget.go | 8E3B9F45C326BCD2BD0174F50FDBE6BC3437905CCF11932A8B3EE5B5599DAF57 |
| budget_test.go | 2B84D68C9D56E9CD6167239EC20AE9E929642BA73A6736B260BED7AC6E4FCFEB |
| ca_fields.go | C1A91AC97C109C48154485A505929789AEB22DC9BD69EB0666377009CDFF0AD2 |
| ca_fields_test.go | 4A1EC88F190E91D7998C167CBBCF4D2790F5BC918C96272A028C6D950CF68725 |
| der.go | AEEA9839770D3922501A91C67C6E85F8531D3B64DB36E76B917F5BF1350CBA8D |
| der_test.go | 37C4113558FA4F62677AC20AEA1ED13314EE018537D0ECB7C808479B8FF671E0 |
| doc.go | F6790FD2DBEDD44538370308E00C3E2E0D7F3EB33789689FF608142C4881A163 |
| errors.go | 893FEC3E470AF2F9C2025740A3A6E5DD64D85572564F03E634740C6E95847A62 |
| extensions.go | E1218192300A0A3FF6792926938346B29A8A3F9CEBA897E344413A2094E6A200 |
| extensions_test.go | 0A2D3B87D307986D1969DCC23B4A349ECBAFF7F74F91B8E986F81729CF85D0CC |
| fixtures_test.go | E29AA04ACAC3EACF9524426F274E8C4F93A002BA74A9598D6EE582BD543B123C |
| fuzz_test.go | EB827AD3EE693F4CE0DBE03EBE0FC6340B16284AE499743EA7BDD245B91B0C14 |
| input.go | 5A6E979C7B1E84357CEBCC26F443D4144F946750C5095976392A94F968C7EB52 |
| material.go | 92E2A18760E9C1C5C90256CA5442A1432A9F42A6E2647A55BCC093A58D68EF47 |
| material_test.go | 981B6735574D28EA53A61A68CFA7A72DA2B761B6E8D90FABD1EB1344A9D7123A |
| owned.go | 787211003166E96697199D16DD5055DA8FF4BDBA9D1A731540EB0187C7B3A07A |
| owned_test.go | A7511F813D4ADB678A2FA23E1EC05FB96E9A7A6948BEE448811DD1EC1F947793 |
| parse.go | AA4BE70D7531000FCB8CB1834FB9CF88AC2633FF51383E81574CABD51A3C5833 |
| parse_test.go | 13F3EB5547CA846F89C6DF3E84C6DE77D79AF9F42FA1C96B31A1C28087158508 |
| raw_certificate.go | 6C4546A21C40178D031D11B6D9BEE6D1D09F607F71275F05594B4AB227F1BEE0 |
| raw_certificate_test.go | 79AE777980EAF1485A962766AD9CE2D44A2C398AB247655750D8D93B63A691A7 |
| spki.go | 95AAA8CE6398AC585836656BFDBA943DA6F6C41D36DF4F67D258F4DBAAA8067F |
| spki_test.go | 62DB5B59B1E11111E33E68E84943401747C5912BE8B89D6FAB3C8E20FBC5251D |

## Final independent review and closeout

Fresh approved gpt-6-astra/xhigh read-only review8 returned **0 Critical / 0 Important / 0 Minor** on the exact freeze8 sources. No deferred Minor or source edit after the clean review. Full report is retained at `.superpowers/sdd/2026-10-11-node-ca-material/final-review-8.md`, exactly as returned. Author save UTC was 2026-10-11T21:44:24Z; this is not a reconstructed reviewer-completion timestamp.

Review independently inspected all24 sources,735 baseline files,44 old testdata,all five normalized original briefs,the unchanged requirement body,all eight freezes and104 attempt paths. Only103 are complete original wrapper receipts;0094 remains unavailable diagnostic metadata. Current final receipts0099–0104 match freeze8 and record formatting/five-command exit0. Reviewer ran no tests,formatter,network or writes; recorded outcomes are author evidence, not independent reruns.

All22 declined-to-judge domains retain the existing explicit Pending obligations and rulings below. The expanded-year leading-zero question and normative/informative offset-range discrepancy did not establish a further defect; no independently settled exhaustive ISO compatibility conclusion is asserted.

The approved documentary closeout adds this execution record and only spec/plan state/checklist and top current-status edits. A final read-only preservation audit is required after those edits; its actual result is retained in the local evidence directory. No source change, new consumer, dependency or Git write is part of that closeout. All broader acceptance blockers remain.

Author final documentary/source preservation check at actual UTC 2026-10-11T21:45:49.9513215Z: **passed**. Inventory760 =735 baseline +24 frozen Go files +1 execution document; dirty entries365 =340 baseline +25 new entries, with no old entry missing. Only the three permitted old document paths differ;732 other baseline files and all44 old testdata remain unchanged. All24 source hashes still match freeze8; no new testdata/consumer; requirement-body SHA256 unchanged; all five checkbox-normalized briefs unchanged; no unchecked task steps. Prior current-status heading and exact raw body remain preserved (the whole file differs because the new top section was inserted). HEAD/branch/index remain unchanged. Exact read-only commands/results are retained in `final-closeout-baseline-check.json` and `final-closeout-check-8.json`; a final confirmation is made after inserting this recorded result. These are author documentary checks, not additional test attempts or an extension of the independent source review.

## Decisions and limits

1. Ruling: Windows PowerShell-only and no-Git limits override Bash bookkeeping, script .gitignore writes, commits, and scratch deletion — apply_patch equivalents retain briefs, BASE, full receipts and canonical progress.md — cost if wrong: less automated bookkeeping, evidence remains auditable.
2. Ruling: approved limited regression commands override generic full-repository baseline/suite — execute per-task nodeca suites plus the five final commands — cost if wrong: unrelated repository failures remain unassessed.
3. Ruling: approved plan Task 5 overrides generic deferred Minor and no re-review template — fix in-scope findings and fresh-review any source changes, preserve evidence — cost if wrong: additional review time.
4. Ruling: inherited Go target is Linux on Windows — add process-only GOOS=windows/GOARCH=amd64 to the approved wrapper, do not write Go configuration — cost if wrong: this validates native Windows, not Linux cross-builds.
5. Ruling: freezing new independently signed golden fixtures requires a one-time test-only generation step — use a temporary helper in fixtures_test.go, retain output receipt then replace it with immutable inline bytes, no persistent generator command or runtime re-signing — cost if wrong: fixture provenance is constrained to the recorded fixed keys/toolchain and independent PowerShell hashes.
6. Ruling: nested AlgorithmIdentifier/Name values can be adversarially deep — bound the private DER parameter/value walker to 32 levels as malformed rather than risking stack exhaustion — cost if wrong: unusual nested out-of-purpose Name/algorithm values are rejected.
7. Ruling: supersede Task 2's 32-depth limit — 0023 proves fixed Go accepts deeply nested Unknown signer parameters and approved profile does not cap depth; use iterative bounded-input traversal instead — cost if wrong: traversal temporary heap remains input-bounded, not a total heap guarantee.
8. Final: Ruling: declined CA signature/signer policy — approved first Pending remains, not added here — cost if wrong: signature authenticity remains unproved.
9. Final: Ruling: declined parent/chain/root/AKI-parent/self-sign/path constraints — no parent input or approved trust — cost if wrong: parent/chain authenticity remains unproved.
10. Final: Ruling: declined current time/max lifetime/backdating — approved time Pending remains — cost if wrong: current validity remains unproved.
11. Final: Ruling: declined deployment provenance/domain/purpose/authorization — explicit future authority layer — cost if wrong: deployment authorization remains unproved.
12. Final: Ruling: declined bundle/revocation/compromise/version/high-water/rollback — explicit future status layer — cost if wrong: current issuer status remains unproved.
13. Final: Ruling: declined CA-leaf/private-key/installation/runtime relations — no leaf/key/runtime input — cost if wrong: actual leaf/runtime use remains unproved.
14. Final: Ruling: declined provider compatibility/migration — no provider or old-profile change authorized — cost if wrong: production compatibility remains unknown.
15. Final: Ruling: declined concurrent mutation/Clear/read — expressly unsupported — cost if wrong: concurrent callers require a separate design.
16. Final: Ruling: declined fatal OOM/crash/unreachable copies/hard real-time — guarantees expressly excluded — cost if wrong: no erasure or latency assurance for those cases.
17. Final: Ruling: declined formal completeness of AST guards — finite guard plus actual-source inspection, not proof — cost if wrong: untested bypasses may remain.
18. Final: Ruling: declined exhaustive fuzz/resource/Linux/race/DB/Docker/full-repo behavior — only approved native limited commands — cost if wrong: these environments/behaviors remain unassessed.
19. Final: Ruling: declined unrelated dirty history/Task10/B01/C12/main readiness — preserved baseline, increment-only review — cost if wrong: whole-project acceptance remains blocked.
20. Final: Ruling: full opaque unknown-parameter universal primitive-content semantics — stand outside narrow fixed-Go field acceptance plus consumed/minimal TLV contract, not a complete universal-value validator or signer policy — cost if wrong: schema-specific illegal primitive content may pass this partial layer.
21. Final: Ruling: opaque SET versus SET OF schema-dependent ordering — no schema-dependent ordering claim for unknown parameters; remove unapproved SET OF assumption while retaining known DN RDN SET OF checks — cost if wrong: unknown-parameter canonical ordering is not certified, requires later signer/schema policy.
22. Ruling: new-package constructed AKI test oracle must distinguish malformed nested bytes from structurally valid excluded representation — retain the original zero-content input as InvalidMaterial and add explicit well-formed nested OCTET STRING as Unsupported control, rather than weaken validation — cost if wrong: the malformed-versus-profile contract could still be misstated; final independent review will inspect both cases.
23. Final: Ruling: global serial uniqueness — exact minimal positive serial projection does not establish uniqueness and spec expressly excludes it — cost if wrong: duplicate serials remain undetected by this layer.
24. Final: Ruling: DN text normalization/semantic equivalence — preserve exactraw names and byte digests without name authorization or normalization — cost if wrong: semantically equivalent differently encoded names are not identified as equivalent.
25. Ruling: rejected GeneralName value-policy boundary — check representation, known field/type/ordering/consumption, but do not infer additional naming/routing, length/range or full RFC/provider-conformance policies from excluded issuer forms — cost if wrong: well-encoded excluded values are not certified as conforming all external subtype/policy constraints; this does not waive malformed encodings or the approved fixed error contract.
26. Ruling: schemas behind unrecognized otherName type OIDs and unknown X400 attribute types — consume/minimally validate their ANY values, including known universal string encodings, without claiming resolution of a missing vendor schema — cost if wrong: vendor-specific meaning and schema-dependent constraints are unproved; every such issuer-bearing AKI still returns no material.
27. Ruling: full legacy character-set repertoire conformance beyond pinned compatibility interpretation — not certified without corresponding repertoire/codec validation; T61/GeneralString compatibility and broader Videotex/GraphicString meanings remain outside this partial guarantee — cost if wrong: legacy-character semantics/compatibility remain unproved; known universal representation/required-container checks are not waived.
28. Ruling: Name AttributeValue interpretation — use pinned Go six string types/compatibility for rejected directoryNames, with representation-only OIDs, not CN/OID-specific RFC subtypes — cost if wrong: representations the pinned interpreter cannot handle are not certified; naming authorization/normalization remain unproved.
29. Ruling: extension universal time precision — canonical UTC/GeneralizedTime lexical/calendar checks use bounded cleared temporary while arbitrary canonical fraction precision is preserved, not Go nanosecond-capped — cost if wrong: finite checks are not complete generic time-policy assurance or trusted current validity.
30. Ruling: attempt 0094 output was lost by an author receipt-collection helper error — keep its command/diagnostic with unknown timestamps, hashes and exit; do not infer an outcome; use independently retained genuine RED 0095 and later commands — cost if wrong: the unavailable attempt remains unauditable and an extra command was required.
31. Ruling: generic TIME/GeneralizedTime checks validate selected representation/calendar rules, not actual leap-second scheduling, chronological interval ordering, or a vendor-defined TIME subtype — no trusted time or scheduling capability added, no original bytes rewritten — cost if wrong: semantic time-event/property conformance remains unproved; known malformed UTF-8/canonical notation is not waived.

32. Ruling: selected TIME source interpretation — follow X.680(2021) normative Tables6/7 for required expanded-year digit count and the -15..+15 local-UTC hour range; the author observed a differing informative AnnexJ +16 endpoint and considered leading-zero expanded-year notation, but review8 established no defect and did not independently settle exhaustive ISO compatibility — cost if wrong: another standards edition or interoperability interpretation may require a separately evidenced correction; no complete compatibility certification is claimed.

The original depth32 ruling is historical and superseded by the following iterative-walker ruling, not a current limit. Prior reviews’ repeated declined domains reaffirm the same rulings rather than dropping exclusions. Opaque unknown-parameter universal primitive-content semantics and schema-dependent SET/SET OF order are expressly not certified. For example unknown UTF8/IA5 parameter content may pass the pinned parser and this narrow layer; this is not complete universal DER-value validation or signer-policy approval. Known RDN SET OF, universal tag forms and recognized CA/SPKI/extension encoding/container grammar remain checked. Rejected GeneralNames now check known string encodings and typed boundaries; rulings25–31 do not waive malformed encoding or fixed error priority. No full RFC5280 subtype value-policy/provider-conformance or unknown vendor-schema certification is claimed. Grammar consulted directly in [RFC5280 §4.2.1.6](https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.6) and [Appendix A.1](https://www.rfc-editor.org/rfc/rfc5280#appendix-A.1).

Review8 found no Critical/Important/Minor finding; no deferred Minor. A source-only audit, finite AST guards and finite fuzz do not prove complete parser correctness. Final status/execution documents are author-checked, not independently source-reviewed. Original historical dirty progress remains preserved; Task10/B01/C12/main gates remain blocked. Current changes and ignored receipts are local only, uncommitted/unpushed.

Review6 supplemental representation/container oracle consulted official [X.680(2021)](https://www.itu.int/rec/dologin_pub.asp?id=T-REC-X.680-202102-I!!PDF-E&lang=e&type=items) automatic tagging/character forms and accessible [X.690(2015)](https://www.itu.int/rec/dologin_pub.asp?id=T-REC-X.690-201508-S!!PDF-E&lang=e&type=items) canonical time/REAL/EXTERNAL/PDV clauses, with matching indexed2021 clauses. Full2021X.690 fetch failed, not represented as read. This is not a full standards-conformance/legacy-repertoire/vendor-schema certification.

Review7/revision8 time evidence uses selected [X.680(2021)](https://www.itu.int/rec/dologin_pub.asp?id=T-REC-X.680-202102-I!!PDF-E&lang=e&type=items) TIME notation and [X.690(2015)](https://www.itu.int/rec/dologin_pub.asp?id=T-REC-X.690-201508-S!!PDF-E&lang=e&type=items) canonical representation, plus the explicit leap-second alternative in [RFC4517 §3.3.13](https://www.rfc-editor.org/rfc/rfc4517#section-3.3.13). Generic GeneralizedTime leap-second validity is an inference from this basis, not a quoted ITU explicit60 rule, and is not proof of an actual leap event. Reviewer7 received the author-retrieved evidence and did not independently browse full standards. No full standards-conformance claim.


Additional review8 source-provenance boundary: selected [X.680(2021)](https://www.itu.int/rec/dologin_pub.asp?id=T-REC-X.680-202102-I!!PDF-E&lang=e&type=items) normative Table6 defines large-year representations by required decimal digit count; Table7 supplies its notation. Rejection of leading-zero signed expanded years is the author's inference from those definitions, not an independently established blanket ISO8601 rule. The same normative Table6 lists local-UTC hours -15..+15, whereas informative AnnexJ.2.11 lists a +16 endpoint. The implementation follows the normative table. The official full 2021 PDF contained no LFn/LBn notation or the hypothesized +001985 example in the author's text search; this is not proof of their invalidity across other standards editions. The limited ISO2004 preview discusses agreed expanded widths and was not the full time section. Reviewer8 considered these questions on supplied evidence and established no source-supported defect, but did not independently retrieve the full standards. Complete standards/edition/interoperability compatibility remains unproved; known malformed encodings and fixed sentinels are not waived.
