# I1 scalar rejection repair — STOPPED / UNACCEPTED

## Approved scope and actual disposition

User approved the bounded in-chat design after589a1d04: reject added members on
integer/bool/string values without executing callbacks; preserve normal values;
isolate decorated test inputs and retain original assertions. Seven new fresh
ordinary-token children, each<=30000ms: two engine REDs, two complete module
GREENs, one PS7 guard-removal mutant, then two restored final GREENs. One fresh
read-only independent review only if all seven receipts meet expectations.
Unexpected results stop all remaining execution with no retry/fallback.

**Stopped at R3.** Both REDs matched the expected decorated-integer refusal
assertion failure. Candidate PS7 complete-module run failed at the original
P1EmptyPrepared assertion. R4–R7 are unrun/frozen; reviewer0/1 is ineligible.
No mutation created, no candidate repair/retest after stopping. Current source
and tests are saved only as an unfinished development checkpoint, not accepted
or safe-to-merge work. I1 and P1 remain open.

## Changes saved for inspection

Product `docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1` adds one
block in Assert-SupervisionPolicyPassive: for Int32/Int64/Boolean/String, inspect
member descriptors and reject IsInstance=true. No member Value is read there.
All other product functions unchanged by AST/text comparison. The intended
distinction follows [Microsoft's IsInstance documentation](https://learn.microsoft.com/en-us/dotnet/api/system.management.automation.psmemberinfo.isinstance?view=powershellsdk-7.6.0),
but documentation is not evidence that this candidate preserves normal behavior.
R3 explicitly failed that preservation gate; no unique cause is established.

Ownership tests isolate values in WrapPrimitive before decoration: value types
use a fresh typed array/GetValue, strings use String.Copy. Exact one-object
constructor and existing type/value/metadata/reference guards remain. Added
FixtureWrappedInputIdentity checks four existing isolation categories for
distinct wrapper/input bases and lack of Metadata on the original input after
decorating the copy. This is bounded coverage, not every string/value case.
Existing empty-string matrix still runs, but the new identity check is not an
empty-string independence proof. No tests were relaxed to pass the candidate.

Each approved child now runs fixture self-checks before product assertions;
complete-module children also run the unchanged original35-case suite before
the ten ownership cases. This scheduling is included in the frozen inventory.
The fixture-only LIMIT marker describes that preflight phase only; R1–R3 do
subsequently import/call product. No claim that these were fixture-only hosts.
Full project/runtime tests were not run; approved scope is this module only.

## Saved observations

| Slot | Host / mode | Exit | ms | stderr chars | ExpectedResult |
| --- | --- | ---: | ---: | ---: | --- |
| R1 | PS7.6.5 / regression | 1 | 1958 | 0 | true: expected RED |
| R2 | PS5.1.26100.9711 / regression | 1 | 2643 | 0 | true: expected RED |
| R3 | PS7.6.5 / all | 1 | 1796 | 252 | false: stopped |

All three x64/admin=False; TimedOut=false, owned-host-exited. Each preflight
passed Constructor, PrimitiveMatrix, Descriptors, InputIdentity, Isolation,
WrappedInputIdentity and Complete. R1/R2 then passed P1I1ApiRequired and failed
P1I1DecoratedInteger exactly, 12 ordered lines each. This reproduces refusal
expectation mismatch, not all downstream metadata mutation manifestations.

R3 had14 lines instead of the frozen59: preflight passed, original P1ApiRequired
passed, then ASSERTION FAILED: P1EmptyPrepared. Stderr reports Prior P1 suite
failed from the ownership-suite catch, not a timeout/transport warning. Other33
original cases and all ten ownership cases were not reached in R3. Actual
decision fields and internal caught exception were not printed. No extra call
was made to recover them. Consequently source guard versus preflight state
interaction cannot be distinguished from saved output alone. No automatic
preflight removal, alternate host or diagnostic replay is authorized.

## Integrity and recovery

New manifest inherits49 prior historical pins, maps the live FE23 product pin to
the new byte-identical before snapshot in this manifest only, and adds13 prior
fixture-packet anchors:62 protected entries. Old raw/manifests unchanged. Before
each child, slot was consumed and prior receipt/stream hashes checked. Controller
used exact ordered output/exit expectations, empty-only stderr, inherited
ordinary token, owned handles, separate UTF8 streams, hidden no-profile children,
no shell. PS5 child-only standard module path avoids PS7 module inheritance;
no system setting change. Timeout/error cleanup paths were not exercised.

After stopping, read-only reconciliation verified62 pins, current candidate,
suite/controller/snapshot hashes, exactly three receipts true/true/false and all
six streams. No import, test, edit to candidate, retry or reviewer followed R3.
Hashes anchor saved bytes; they are not independent authenticity attestation.

Ignored local packet `.superpowers/sdd/2026-10-17-p1-scalar-rejection-repair/`
retains scope, before snapshots, frozen suite, candidate snapshot, controller,
manifests/index and streams. Git carries source/tests/report/status/recovery,
not raw/controller/runtime. Recover latest development branch, not main; prior
working fixture-only checkpoint is589a1d04. Do not run this candidate by cloning
or treat candidate bytes as passing. All old packets retained, none deleted.

| Artifact | SHA256 |
| --- | --- |
| Product before | `FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB` |
| Product current / candidate snapshot | `A38C1EE2F6EF75C09F0729AFEF5A091C7317D5AC005BD2AF2FB3AC8BD410E95E` |
| Ownership before | `A476AB519D5CA1E9032C2B27E06AB1D25CC4706DAA086C7F02ADD01557CEFABF` |
| Ownership current / frozen | `86FD5C129C18181CFC3106BC5657EEE23FED1F254CD4E495A8C59C1EA7A61B38` |
| Original35 unchanged | `61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3` |
| frozen.json | `3C801AA3ACCB49C4A184EA0FCFF33213CD3A2046C94D5CA801F26129CBD47B4F` |
| green.json (candidate seal, not a PASS verdict) | `47DFB2ED6455FAE5B77A6BFD73A8CE1136BB5B5B3CC3B2D5D79A1B84473914CD` |
| run-slot.ps1 | `EB832D2A53810EC991BB0CA1396C81CE6AC71052A1999127F57EFF0105F0338D` |
| invocations.json | `8ECECA15E3870BB924E5F3941DC1C182FB01D1D2A88543F63100F1DC621E10FF` |
| R1.stdout | `C7588F6BFBAA9879D3B38D7E9192D4E55B3C46BB8D31CE4D8892A7E124F48E38` |
| R2.stdout | `BFE03FC296CBA400CF75CF31C665585BFB95C7C8C95D05D7A8CB800B6ACCB50D` |
| R3.stdout | `BC8F2F1038F15F24807F66403878008294691ACFF8CC057F69C3F4BF1726A7AE` |
| R1/R2.stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| R3.stderr | `240F72B8715B34111A8BA66FAF7D41A1BCEE95E4801973985A912BEF97699CDA` |

New budget3/7 consumed; four remaining slots frozen, review0/1 ineligible. All
earlier spent/frozen allowances and false verdicts remain unchanged. No main,
P2/P3, live supervision, whole-project or C12/release acceptance. Next requires
a separately bounded diagnostic question for P1EmptyPrepared, not continuing R4
or reclassifying R3 as a successful GREEN. No new implementation plan needed
merely to save this stopped bounded task.
