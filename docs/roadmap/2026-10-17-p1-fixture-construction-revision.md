# P1 fixture constructor revision — stopped checkpoint

Status: **UNFINISHED / F1 FIXTURE ISOLATION ASSERTION FAILED.** User approved
the [fixture-only revision plan](../superpowers/plans/2026-10-17-p1-fixture-construction-revision.md)
after b98251f4bcb3cf5a3069dcc3704705706eae9e3d. F1 is consumed and false;
F2 was not launched and is frozen. No retry, alternative construction, post-stop
fixture correction, product import/edit/test, original-slot replay or reviewer.
The partial candidate is preserved as an explicitly unaccepted test checkpoint.

## Authorized change and static preservation checks

The ownership test gained private P1I1-WrapPrimitive, explicitly resolving the
public one-System.Object constructor and invoking it with a single argument.
Only the two wrapper-construction expressions in Decorate and PassiveDescriptors
changed. All original helpers and ten product-case bodies were statically AST
compared against the pre-edit snapshot, allowing only those two substitutions;
the literal product expectations and original combined fidelity guard match.
The five new fixture checks were written before implementing the candidate.
Static parsing passed. Group=fixture returns before the old-test/product-import
path; when an assertion fails, its catch exits instead of reaching that path.

Baseline ownership suite9AA7C29C was preserved byte-for-byte in this plan's new
ignored packet and remains in Git history. Current suiteCF81BD28 is an authorized
new version, NOT a current match for the old9AA7 pin. Historical manifests, raw
streams, verdicts and diagnostic probe are unchanged. ProductFE23DF3B, original
35-case suite61A3BACF and prior protected15 remain fixed.

## Actual F1 evidence

PS7.6.5 Core x64, DESKTOP-H4SVCLU\Lenovo admin=False, inherited ordinary token.
Exit1 after973ms, no timeout, empty stderr; controller ExpectedResult=false.
Saved stdout has exactly context, host and these lines in order:

```text
PASS: P1I1FixtureConstructor
PASS: P1I1FixturePrimitiveMatrix
PASS: P1I1FixtureDescriptors
ASSERTION FAILED: P1I1FixtureIsolation
```

The constructor/signature check passed. The scalar matrix passed for Int32
0/1/-1/Min/Max, Int64 0/Min/Max, bool true/false and strings empty/clean/refused/
unknown across its construction/metadata/assignment/retrieval/helper-return
checks. ScriptProperty/AliasProperty/ScriptMethod descriptor cases passed with
callback count0. These are bounded named-case results, not universal guarantees.

The isolation case failed. It creates separate Int32/Int64/bool/string fixtures,
requires distinct metadata arrays, mutates one and checks that others are
unaffected. The saved group-level assertion does NOT identify the failing
primitive category or subcondition. Reference inequality, metadata-field identity,
second/third fixture contents and later fresh-fixture checks remain possible
failure points; none is established as the cause. No interactive helper call or
extra diagnostic was made to distinguish them. The original product I1 was not
evaluated or dynamically reproduced.

FixtureComplete was not reached, so its dynamic product-function-absence check
and final LIMIT marker did not run. Static branch/catch tracing supports no
product import; this is not independent execution attestation. F2/PS5.1 outcomes
are UNRUN, not inferred from prior successful diagnostic hosts or F1's partial
passes. Do not promote three PASS labels into overall fixture acceptance.

## Frozen local execution-byte hashes

|Item|SHA256|
|---|---|
|Pre-edit ownership snapshot|9AA7C29CCB640270EBCFD562F75AA3A47C95324C5A4CE786E0E354BDBC0D694F|
|Revised, unaccepted ownership suite|CF81BD28AE69E7A49897FACFBFEA68022A384353BD155F9EC521B9B451FE6256|
|New controller|E33AE83A7D02E6E24D7F92788A4F7BDAE5D83337289B94F5081F3A29991006B2|
|New manifest/inventory/pins|274170F8A989FA52D26721C15D498D275792F879698E4995F58DD8A9FC705B8E|
|Stopped one-row index|12EE36F1D550580A3C86BB661FA7C1B189053030DF4FF0BDB315C9F9A14FFD2C|
|F1 stdout|6FAF1B590B3DE76DE9BAC60D1CD69CC5110FC25EED13C80A6EF91FAC2149750D|
|F1 empty stderr|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

New manifest protects29 entries: the previous diagnostic packet's21 pins, with
the historical ownership-test entry redirected to the identical pre-edit snapshot,
plus eight unchanged diagnostic probe/controller/manifest/index/stream anchors.
All matched before/afterF1 and parent handoff. This version mapping does not edit
old manifests or falsely claim the revised test matches9AA7. Git may normalize
line endings separately from these execution hashes.

## Rulings, budgets and recovery

- Approved fixture-only allowance overrides generic extra RED/GREEN/baseline/
  task-done/whole-project runs, historical review, automatic fix and packet
  deletion. Cost: partial candidate remains unverified and no independent review.
- PowerShell-native bookkeeping reused the verified linked worktree, no new
  environment/setup/security changes. Cost: narrower bookkeeping, not more scope.
- The test's version change was expressly authorized; pre-edit bytes/old pins
  are retained. Cost: recovery must distinguish historical from current test
  versions, never promote an old verdict for the revised file.
- Isolation group failure is not renamed a successful constructor repair or
  a product finding; its specific subcondition is unknown. No alternate method,
  weakened assertion or second host is permitted after the stop.

Retain `.superpowers/sdd/2026-10-17-p1-fixture-construction-revision/` with ledger,
pre-edit snapshot, controller/manifest/index and originalF1 stream pair. All older
packets remain retained. Git contains revised test/approved plan/report/recovery,
not ignored raw or machine/runtime state. Recover latest development branch
codex/c12-b01-task9-coordinator, not main, and do not run this candidate implicitly.

Fixture allowance1/2 spent, remainingF2 frozen. Original repair slot1 remains
ExpectedResult=false, its six remaining slots frozen/review ineligible; diagnostic
allowance2/2 remains spent. Proposed future R1–R7/product repair/conditional scoped
review have NOT been approved or executed. I1/security/Stage B/P2/P3/main/C12/
release remain open. Next requires a separately bounded investigation of the
isolation assertion or an explicit decision about the unfinished candidate;
this handoff grants no new host, fix or budget transfer.
