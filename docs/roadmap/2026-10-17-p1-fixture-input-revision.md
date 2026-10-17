# Fixture input identity correction — bounded validation

## Scope and change

User approved the short in-chat design after `f7f7ee50`: correct only isolation
fixture input identity; retain original assertions and product behavior. Three
new fresh ordinary-token children, each <=30000ms: PS7 expected RED, then PS7
and PS5.1 candidate GREEN. Wrong RED, GREEN failure or infrastructure failure
stops remaining slots; no retry/fallback, extra tests, imports of product, review,
install or security-setting change. Generic skill full-suite/retry defaults do
not override this explicit cap/exclusion. No implementation-plan document was
needed for the approved bounded change.

Changed `docs/diagnostics/isolated-fuzz/offline-supervision-policy.ownership.tests.ps1`:

- Extracted the four existing isolation input factories into P1I1-IsolationVectors.
- Added FixtureInputIdentity: three owners per category must retain expected
  type/value and have pairwise distinct BaseObject references. This catches
  accidental reuse of cached scalar identities before decorating separate owners.
- After observing RED, changed only Int32/Int64/Boolean factory bodies: each
  invocation creates a fresh typed array and retrieves its boxed value via
  GetValue(0); Boolean is explicitly set true. The string construction remains
  the existing new-character-array string constructor for `clean`.
- FixtureIsolation consumes the same factories; its original decoration,
  returned-array distinction, mutation, metadata reference and third-owner
  assertions are unchanged. WrapPrimitive, Decorate, scalar/metadata checks and
  all product assertion/call paths are unchanged.

No shared-identity guard was weakened, no blanket change to scenario builders,
and no product repair was attempted. This validates four isolation categories
(Int32 zero, Int64 zero, Boolean true, String clean), not all possible scalars.

## Verification actually run

| Slot | Fresh host | Expected / actual | Elapsed ms | stderr |
| --- | --- | --- | ---: | --- |
| R1 | PS7.6.5 Core x64 | RED / exit1, FixtureInputIdentity assertion | 819 | empty |
| R2 | PS7.6.5 Core x64 | GREEN / exit0, all six fixture groups PASS | 1771 | empty |
| R3 | PS5.1.26100.9711 Desktop x64 | GREEN / exit0, all six fixture groups PASS | 2376 | empty |

R1 ran with the new guard but original literal primitive factories. Constructor,
PrimitiveMatrix and Descriptors passed before the expected identity assertion;
Isolation/Complete were not reached in R1. R2/R3 passed, in order: Constructor,
PrimitiveMatrix, Descriptors, InputIdentity, Isolation, Complete. Final marker:
fixture construction only; no product import/assertions, I1 closure or budget
renewal. ExpectedResult=true for R1 records the expected RED, not a passing test.
All children reported admin=False, TimedOut=false, owned-host-exited. Timing
covers launch/wait/drain; each is below30000ms. The only mode executed was fixture.

Before running, static parsing found no errors. AST/text comparison against
CF81 confirmed existing helpers outside FixtureChecks and the original isolation
assertion block unchanged. GREEN AST comparison against RED confirmed every
function except IsolationVectors unchanged; RED-to-GREEN diff is just the three
value-type factory bodies. Final Git diff confirms main product paths unchanged.
This is scope-preservation evidence, not a substitute for executed assertions.

Frozen controller used owned process handles, separate UTF8 stdout/stderr,
hidden no-profile noninteractive children, no shell and inherited ordinary token.
Slot consumed before launch; each prior expected result/raw hash checked before
continuation. Exact ordered six-line RED / nine-line GREEN output, exit code,
host line, empty stderr and unchanged pins are required. PS5 child received only
its standard Windows PowerShell module path to avoid inheriting PS7 module paths;
no machine setting was changed. Timeout/cleanup error paths were not exercised.

Final read-only verification checked3/3 receipts, all streams, current candidate,
RED/GREEN snapshots, frozen controller and49 protected historical entries. The
previous41 entries were inherited, except the current ownership-file CF81 pin was
mapped to the new byte-identical before snapshot in the NEW manifest only; eight
comparison-run anchors were added. All old manifests/raw remained untouched.
No test was replayed during reconciliation.

## Recovery and hashes

Necessary test/source and this report are portable in the development branch.
The ignored local packet `.superpowers/sdd/2026-10-17-p1-fixture-input-revision/`
retains scope, before/RED/GREEN snapshots, manifests, controller, index and streams;
no packet was deleted. Clone does not restore that raw packet/runtime and does
not authorize rerunning any slot. The CF81 predecessor is recoverable from
checkpoint f7f7ee50; current candidate is A476AB51, not historical CF81/9AA7.

| Artifact | SHA256 |
| --- | --- |
| Before snapshot | `CF81BD28AE69E7A49897FACFBFEA68022A384353BD155F9EC521B9B451FE6256` |
| RED suite snapshot | `CAB6E890438297F062B95AC46DEF8F115A5747181D3C2AC1493ADD84C55A8B80` |
| Current / GREEN suite | `A476AB519D5CA1E9032C2B27E06AB1D25CC4706DAA086C7F02ADD01557CEFABF` |
| frozen.json | `DF5CCC3FFFB9426CE4657C42A966DD0CB670CCD342EF2D31C6EC8D676E25C8B2` |
| green.json | `8018B07BE4B2D6A091C0F230DFB623BC807EF029980B938BFE36249D7B356A7C` |
| run-slot.ps1 | `057F9E1393890536AD58839ACAF04DBDD7686948E2318A6B11DE9DFC0298DC4B` |
| invocations.json | `8CDFCCD376B054F7A0BD503F6DC4D8DFF39F5EE78E08992811822CC55C7EF03E` |
| R1.stdout.txt | `1BEE9862288031587DA3679DC8B2ABFAC43F9738C773CBB2D127D5AA198BCF37` |
| R2.stdout.txt | `03A717EF04B7AB840041796938ACB347386E88D1812B4719A288E6B5B6C3BB51` |
| R3.stdout.txt | `3585FE0E020C7693085D2CCCA41B9F32A8B61339B136DBB7EC04E5C6951ED21F` |
| All stderr files (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

## Remaining gates

This bounded fixture correction passed its approved RED/GREEN sequence; new
allowance3/3 spent, no reviewer or remaining run. ProductFE23DF3B and original35
suite61A3BACF unchanged. No regression/all product mode, whole-project tests,
product fix or independent review executed; I1/P1 remains unfinished.

Historical fixture F1=false and F2 frozen remain historical; they were not rerun
or relabeled. Original repair's failed first slot remains false, six original
product slots frozen, review ineligible. Diagnostic budgets2/2, localization1/1,
identity comparison2/2 remain spent. Future product R1–R7/conditional review
remain separately gated proposals, not these fixture R1–R3. No main merge or
P2/P3/C12/release acceptance follows. Next work is to scope the product ownership
repair and its complete verification sequence, not to silently run regression.
