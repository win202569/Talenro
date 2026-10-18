# P2-B raw scalar repair: RED captured, controller decode blocked

## Status: unfinished, no product repair

User approved the bounded in-chat design after 340c08d9: a NEW candidate/test
leaf, three ordered ordinary targets (RED, focused, full), then one conditional
independent readonly review. Old sealed sources/records and budgets stay intact.
New leaf: `docs/diagnostics/isolated-fuzz/p2b-scalar-raw-repair/`.

First target consumed1/3. The expected test failure occurred, but the controller
failed strict UTF-8 decoding and retained Outcome INCOMPLETE. Remaining2/3 are
FROZEN; review0/1 is INELIGIBLE. No retry, corrective source edit, new probe,
reviewer or product change after that stop. No product seal exists. Both the
new unchanged baseline copy and historical structural candidate remain
B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.
M1 remains an observed, UNFIXED controlled input-validation gap.

## Preparation and regression contract

Existing clean linked worktree/branch reused; no new checkout or install.
New scalar.tests.ps1 catches the actual bug: restoring hidden-filtered public
Members validation must fail the hidden-decoration refusal assertion. It invokes
the real public codec/binding APIs with literal data, not mocks or product-built
expected results. Each of the same35 independent vectors saves its preconditions,
actual retained descriptor identity/class/flags and before/after counters. All
rows are saved before a final HiddenDecorationRefusal contract assertion.
Seven clean cases must accept; all28 decorated cases must refuse invalid-input
with no output payload/sequence and zero controlled counters. Authority flags
must stay false. Fixture construction retains the prior corrected exact-SDK raw
Add approach and saves partial rows before assertions.

The unchanged original full33-group tests, descriptor fixture PS/C#, literal
frames and definitions-only data helper were copied byte-for-byte into this
new leaf. Their SHA256 respectively:

- Full tests: 7CBC263600948F9806D8749B43D0B9DB55C100F48757B1C5CC49CF94A92C60B1
- Descriptor PS: 402DFB2C2CAF4EFE6F696EB5AAA7FBAE5C983488A2BF7B3FFFF2D39258D63D0E
- Descriptor C#: 87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE
- Literal frames: 5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D
- Data helper: 8FE8F3D323A3F9B5DE04AEAFC0A86CB900D92A26BD6B58F3B9DCC39C79314BB2

All six new PowerShell files parsed without errors before freezing. The two
test comments were clarified before Prepare; no frozen file was subsequently
changed. The new product file was only an unchanged baseline, not a repair.

## First target and read-only reconciliation

PID17588,1970ms,exit1;ordinary PowerShell7.6.5 Core x64. No timeout/stderr,
owned exit and concurrent separate streams complete. Bounds60000ms execution,
2000ms owned exit,2000ms drain. Only the owned target was in scope. Counter C#
compiled once in that approved target only; observable positive control passed.

Controller saved raw stdout/stderr hashes and post-run pin integrity before
attempting strict UTF-8 decoding. GetString then failed on byte A1 at zero-based
index76226, so ExactExpected=false and Outcome=INCOMPLETE were never promoted.
Readonly inspection found A1 AD at indices76226–76227 in Error.Position; strict
CP936 decoding resolves this pair to U+2026 (ellipsis) in the PowerShell-rendered
source excerpt. CP936-decoded JSON parsed completely and reported:
Stage HiddenDecorationRefusal, Passed=false, Captured=true, Error chain exactly
System.Management.Automation.RuntimeException / HiddenDecorationRefusal at
scalar.tests.ps1:149. No fixture failure preceded this contract assertion.
This is a separately documented read-only interpretation, NOT a rewrite or
promotion of the controller receipt. No raw file was transcoded or replaced.

Same seven modes per variant, in order: encode Message OriginTick/Payload,
bind Message OriginTick/Payload, bind Expected Schema/BindingDigest, decode
input array. Scalars long/int; arrays byte[]. All35 row phases Completed.

| Rows | Variant | Accepted | Refused | Post-API counter total |
| --- | --- | --- | --- | --- |
| 1–7 | clean | 7 | 0 | 0 each |
| 8–14 | visible exact PSNoteProperty | 0 | 7 | 0 each |
| 15–21 | hidden exact PSNoteProperty | 7 | 0 | 0 each |
| 22–28 | visible CounterNote subclass | 0 | 7 | 0 each |
| 29–35 | hidden CounterNote subclass | 7 | 0 | 0 each |

Accepted reasons encoded/binding-matched/decoded match the respective modes;
all14 refusals are invalid-input. All recorded precondition booleans true,
all base/root-descriptor identities retained, authority flags false. Counters
cover CounterNote only, not arbitrary getter execution/universal passivity.
The test therefore observed the expected bug on the unchanged baseline, but
the approved controller's exact-RED gate remains unfulfilled. No GREEN or full
run occurred. Historical full33PASS is not a new full regression result.

## Sealed integrity evidence

1046 mapped pins=1030 inherited+four previous packet files+previous diagnostic
report+ten immutable new inputs+baseline snapshot. All rehashed after target
and again read-only: zero drift. Engine and SDK separately pinned to
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 and
2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B.

Ignored packet: `.superpowers/sdd/2026-10-18-p2b-scalar-raw-repair/`.

| File | SHA256 |
| --- | --- |
| frozen.json | 6229E52C1AE3D07B2FDF4AE08C58C5381CF40350C7CF8D903695152A2EB3AEA0 |
| invocations.json | 7E7C5FA1D8A543D6262C734CEB465E2320DD320D87E05551423FA27E3B752751 |
| red.stdout.bin | 78B6E38372B86EE436A1AF42331651AF3DBDD91734331D589A574CF57EDA7767 |
| red.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Sealed controller: 0EC619E9E371E2A3989FF3BE87BE47B91D39906A5DA5BFA0AB23249CA4877133.
Sealed scalar tests: 26DFF070F4E1BE524492F72D9BA3E3E906D2F78C58C07D2FA0D6DFB9CEFB3E2A.

## Next decision, not yet authorized

Recommend preserving this entire packet and all sealed source. Separately
approve a NEW continuation controller that validates the saved RED via explicit
strict CP936 read-only interpretation plus its original hashes/preconditions;
never edit/promote the old invocation or rerun RED. If approved, use ONLY the
original remaining two hosts for one sealed minimal product repair, focused35
and unchanged full33 regression, then the original one conditional readonly
review. Specify deterministic child output decoding for subsequent captures;
do not loosen process/identity/pin/stream/timeout/oracle checks. This proposal
does not itself unfreeze those slots or authorize implementation. Any mismatch
in reconciliation must stop. No new target budget is requested here.

Current and all previous consumed/frozen budgets remain unchanged. Preserve
all old findings including the incomplete earlier scalar fixture and completed
retained-fixture diagnostic. No main merge, formal P2-B/PS5/P2-C/P3/full C12,
live/runtime/security acceptance. Recover latest development branch, not main.
Clone restores portable source/tests/report/matrix, not ignored raw/SDK/runtime/
caches/mappings/conversations, and never resets budgets or grants replay rights.
