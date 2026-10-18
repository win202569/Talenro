# P1 strict timing JSON checker — offline44/44

Date:2026-10-18. Base:2d020e508908133df8d728395ea29d4ff6b9a2c1.
User approved the bounded checker-only design plus one new offline regression
batch. This checkpoint does NOT authorize an actual candidate timing diagnostic.

## Verified result

44/44offline cases passed,exit0,elapsed22208ms,no timeout. Owned host exited,
both streams complete,stderr0bytes,607pins intact,controller error null.
Summary:RESULT: OFFLINE_FIXTURES cases=44 passed=44 failed=0 productCalls=0.
49stdout lines=44PASS+4logger TIME records+1summary. No target-case failure,
post-launch repair,retry,warmup,extra host or reviewer.

The old exact duplicate-field RED witness and new passing regression input have
the SAME SHA256:
459400F661D6EA2907330D6C991F9FB5580A3E7E677CD18A46617C4B06887038.
Previously the old checker admitted that ambiguous input as PREFIX_ONLY;the
new checker rejects it. Historical36/37RED and old sources remain preserved,
not rerun or overwritten. This is historical RED -> new GREEN on identical input.
New variant assertions have no fresh old-checker negative-control run.

## Minimal revised path

[Independent revised checker](../diagnostics/isolated-fuzz/strict-timing-json/inspect-capture.ps1)
first parses each TIME/FACT record with the already installed runtime's JSON
document API and enumerates DECODED ROOT property names,using OrdinalIgnoreCase
for uniqueness in the PowerShell property namespace. Duplicate exact/case/escaped
aliases are rejected BEFORE ConvertFrom-Json can collapse them. JSON document
resources are disposed on success/failure. No regex JSON-name extraction or
new dependency/install is used.

After that guard,the original converter and all schema/type/order/counter/tick/
ownership-fact/acceptance checks remain. Schemas are flat root fields,not a new
general nested JSON API. The revised checker reads the exact old output inventory;
the old checker B3C6FE88 and instrumented test986CDE76 stay byte-identical.
Candidate1971F808 is unchanged and NEVER imported or invoked in this batch.

## Regression coverage

All37prior offline cases remain,including the original duplicate-json-property
case. Seven added cases passed:

- TIMEduplicate case aliases and Unicode-decoded aliases.
- FACTduplicate exact,case and Unicode-decoded aliases.
- Valid single escaped TIME and FACT field names are NOT falsely rejected.

The five new duplicate variants require the explicit duplicate-json-property
rejection reason,not any unrelated exception. Existing malformed/truncated/
reordered/clock/counter/fact/prefix-versus-complete rules remain covered.
The original complete synthetic text remains COMPLETE_TEXT_ONLY with receipt/
runtime/acceptance flags false. The actual logger helper sentinel/nesting/
wrong-end checks remain passed;only three AST-allowlisted logger definitions
execute,never the ownership test top-level code or product helpers.

Static checks:all three new scripts parsed,case count44. Removing the new raw-key
guard and restoring the two parse call sites/routing reproduces all1200original
checker significant tokens. This is static preservation,not a proof of all parser
inputs. Runtime evidence is limited to the44offline cases.

## Execution and preservation notes

The controller initially found a singular/plural prior receipt filename typo
during pin preflight BEFORE packet creation,manifest freezing,attempt consumption
or child launch. Read-only path inspection confirmed that no packet existed.
That preparation routing typo was corrected before the ONE actual batch.
No frozen or old record was edited. A live read of exclusively held output
was refused during execution;final streams were read only after host completion.
Neither event created another offline host or reran a fixture.

The new offline batch is1/1consumed. Old candidate-focused1/3spent+2/3FROZEN_UNRUN,
old offline36/37batch and older frozen budgets remain unchanged.
607frozen entries=597inherited+4prior packet anchors+6new portable files.
All checked before/after launch and again at handoff. Raw records stay ignored.

| New strict-timing-json file | SHA256 |
| --- | --- |
| .gitattributes | A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9 |
| capture.ps1 | A20C5FD6667A5DCC47327E09B5AE0F626419DD23FBEFC51EE0B8B6D48EB7C2D1 |
| inspect-capture.ps1 | 7E140A1BA753E1CBA836B36753FB3B741F615CB5941E9C1FD4990F9AF476B4D4 |
| offline.tests.ps1 | 023C7215DCE083E76013D8B7E0BEDD1FAE5E0D9BA0FE1653F5F67BA241F2C632 |
| README.md | 40C1BF731F17E9FBE1BE4D2926AF3F167E13E5F410D6373667044F31CB8E8156 |
| source-pins.json | 1976C94C2CFBE4E071F8168C550E7A0BFC08294E0878F68560AE96AA8622E5B8 |

Ignored packet:.superpowers/sdd/2026-10-18-p1-strict-timing-json/

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 90F5E19FE954E37C81AB60FB9B224F3DB0462A737FAAD3790A9F057A3858FDF7 |
| invocation.json | D37361FE9BF3C73B2290848AB94AF456655810B0F13A0446D429EC68C5CF5A3C |
| stdout.txt | C6F3D25AF65AAC206F018FB0A67155B746D11A5916E486F39114245BCBF8830F |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

## Remaining boundary

The duplicate-root-key checker defect is fixed and verified WITHIN these offline
vectors. This is not actual candidate output,ownership/runtime verification or
a general malformed-JSON proof. Logger wrapper/pipeline behavior beyond the raw
sentinel,real stage timings and the original30s timeout cause remain unknown.
No product-mutant sensitivity proof,performance measurement or acceptance gained.

Recommended next scope:separately approve one new bounded instrumented candidate
diagnostic/controller,using this revised checker and unchanged timing test,with
30s wall+2s owned exit+2s drain,pinned files/raw inventory and stop-on-failure.
Do not run it automatically or reuse frozen full/ownership slots. Its purpose
would be stage attribution,NOT ownership/main/I1/security/runtime/C12/release/
PS5/performance acceptance. Portable code/report/status are preserved on the
existing development branch;cloning cannot restore raw packets or allowances.
