# Direct integer-helper regressions prepared — NOT RUN

Aftercb932195, the user approved preparing direct tests with NO execution.
Added only `docs/diagnostics/isolated-fuzz/offline-supervision-policy.integer-validation.tests.ps1`
plus this checkpoint/status/recovery documentation. Product, original test
files, full-entry inventory and diagnostic controllers remain unchanged.
No target host, product dot-source, fixture evaluation or test body was executed.

## Prepared behavior coverage

Fourteen groups, each stopping on first assertion/fixture failure:

| Group | Break it is intended to catch |
|---|---|
| ApiRequired | Missing direct integer helper or owned metadata factory |
| PlainDefaults | Lost zero/default bounds/int-long admission or success output leakage |
| OwnedMetadataAndPositional | Broken explicit metadata or existing positional argument order |
| SignedBounds | Missing bound, exclusive endpoints, signed extremes rejected |
| InvalidTypes | Coercion/admission of null,bool,string,float,decimal,unsigned,short,byte,char,enum,object |
| WrongTypeNoCallbacks | User-shadowable GetType/ToString evaluation on a noninteger |
| NoteDecorations | Skipped direct rejection of owned NoteProperty decorations |
| AliasDecorations | Skipped direct alias rejection or alias target evaluation |
| ExecutableDecorations | ScriptProperty/ScriptMethod execution or admission on owned integers |
| DecorationBeforeBounds | Bounds failure incorrectly preceding decorated-scalar refusal |
| TypeBeforeDecoration | Decoration work preceding the exact-integer type gate |
| DecorationAfterClean | Cached acceptance or retained refusal across calls on the same owned wrapper |
| CleanAfterRefusal | Cross-value/metadata-context contamination |
| Complete | Terminal marker emitted only after preceding groups finish |

Assertions invoke the REAL Assert-SupervisionPolicyInteger, not the scenario
evaluator/Shape checker or a mocked product. Expected errors are literal contract
categories, not results computed by the product. Helpers capture success/error
output and assert no pipeline leakage. Explicit/default metadata paths are
prepared across type/bounds/decoration cases; positional calls also exist.

Fixtures use fresh typed-array boxed int/long values and fresh string storage,
wrapped via the established SDK PSObject object constructor. No cached literal
is decorated. SDK type/view getters and member inventories check that decoration
survives; fixture errors cannot masquerade as expected product refusal. Extension
names include Probe,Members,Length,Rank, with GetType also exercised for executable
members. Callback counters must stay zero. The only member removal is from an
owned test wrapper during the clean/decorated/clean sequence; it is not filesystem
or project data deletion. No installs,Add-Type,process launch or file writes occur
in the test code.

## Explicit verification limits

Only AST parsing with the current PowerShell7.6.5 parser and static inspection
were performed:zero syntax errors,14 unique case names,four direct product-call
sites, no scenario/Shape indirection or product overrides. This is NOT a PS5
runtime check, fixture validation, semantic PASS, mutation-test result or completed
TDD red/green cycle. Both engine execution and fixture preservation remain
UNVERIFIED. A baseline semantic PASS would still not prove fewer type queries.

No CodeProperty runtime fixture was constructed. Its safety guard remains in
the unchanged product, but dynamic CodeProperty callback refusal is not claimed.
Existing scenario suites and original complete-entry checks remain necessary;
this new file is standalone and is NOT silently inserted into their83-line
expected inventory. No product type-reuse candidate has been implemented.

Prepared complete output, if it succeeds in a future authorized run:two original
CONTEXT/HOST lines,14 PASS markers and one LIMIT line (17 lines). This inventory
is a proposal for a future freeze, not an observed output receipt. Any runtime
failure must be recorded accurately,including fixture failures,without weakening
checks to force a pass.

## Integrity, budgets and recovery

Prepared test SHA256:
`E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98`.
Unchanged product SHA256:
`953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA`.
All283 previous frozen pins held. Loading-context allowance stays1/2 spent;
Inherited slot unrun/FROZEN. All older spent/frozen budgets and review0/1
unused/ineligible remain unchanged. No new execution allowance was used.
PS5/I1/P1/candidate/main/C12/security/runtime/release gates are not cleared.

Next recommendation is a separately approved NEW two-host baseline allowance:
ordinary fresh PS7 then PS5, this standalone suite once each,30000ms per host,
exact frozen output and source/engine checks,stop on any failure/timeout,
no retry/warmup/product edit/old-slot reuse. Do not launch until approved.
The user authorized preparation only in this turn. A product change still needs
its design/verification boundaries and honest test-first evidence.

Recover this file/report,current status and runbook from the existing development
branch. Raw/controller packets,runtimes,caches or replay authority are not
restored by Git. This is an explicitly UNVERIFIED preparation checkpoint.
