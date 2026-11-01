# Owner scope verifier: typed comparison and retained failure evidence

Status: PROPOSED; implementation and three-computation allowance pending approval.
Small auxiliary-verifier revision; original runtime architecture remains unchanged.

## Evidence and decision

Baseline0e17eba9703846725b1a8b8eca0c286bec887f6d; see
[actual stopped GREEN](../../roadmap/2026-11-01-offline-owner-scope-signature.md).
Readonly tracing found that DefinitionsExact compares source FunctionDefinitionAst.Body
with loaded ScriptBlock.Ast.Extent, stripping braces without checking node kind.
[PowerShell7.6.5 function compilation](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/parser/Compiler.cs#L3135),
[wrapper](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/runtime/Operations/MiscOps.cs#L1250)
and [Ast accessor](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/runtime/CompiledScriptBlock.cs#L1968)
retain the function definition node. The right-hand extent includes its declaration,
not just its body. This explains an algorithmic false-rejection path; it does not
recover the missing comparison operands from the old failed run or regrade it.

ODS-Case also throws before producing a FAIL entry or preserving completed groups;
the outer catch saves only the exception. Fixing only the initial lookup catch did
not cover these later cases. Another one-line assertion edit leaves this failure
mechanism intact. Recommended: one typed comparison boundary plus one case-result
boundary for all nine groups. Reject relaxing equality or directly executing the
profile to test visibility. This recommendation awaits approval.

## Typed comparison

New auxiliary-only Assert-ODImportedDefinition(Expected,Loaded,Comparison) accepts
two FunctionDefinitionAst nodes and a private diagnostic holder. Reject other
node kinds rather than guessing. Expected names are the four original allowlisted
unqualified names; loaded names must be exactly script:<expected name>.
Compare Body.Extent.Text with Body.Extent.Text exactly, including braces. Compare
inline ParameterAst.Extent.Text arrays separately, preserving null versus empty,
order, types and defaults. Body param blocks are already included in body text.
No TrimStart/TrimEnd character-set surgery, broad whitespace normalization or
ignoring name/parameter differences. Pass emits nothing; mismatch throws exact
RuntimeException Message/FQID 'OD: imported definition mismatch'.

Capture BEFORE comparison: expected/loaded node kinds, names, body text, inline
parameter texts and source identities. A diagnostic is not a runtime capability.
Unknown nodes must retain their kinds and available text, never fabricated body.

## Retained case boundary

Every original nine-group body runs through one case runner. Preserve the first
actual ErrorRecord, current group, completed results, comparison diagnostics and
before/after source identities when a case fails. Mark only actually completed
groups PASS, actual failing group FAIL, later bodies NotRun without executing them.
The diagnostic remains Unexpected=true/exit2; never becomes a successful final.
Move setup inside the protected phase boundary; no unguarded helper extraction
after a failed case. Setup or diagnostic-projection errors retain the original
cause separately from the recording error. Controlled runner unit faults are
explicit test fixtures, not claimed profile/writer events.

## Scope, provenance and finite budget

Create separately versioned full auxiliary owner-scope-verifier.tests.ps1 and
regression owner-scope-verifier-regression.tests.ps1, both outside31runtime sources.
Preserve older tests/inputs/results/archives and all five current stops unchanged.
Only driver.ps1 requires further runtime-source binding changes; the other four
candidate files stay at baseline bytes. This remains within the original five-
source window relative to parent D8D31B8B…FAC57C, other26 unchanged.

Request3additional existing-tool computations: verifier RED, verifier GREEN, full
nine-group scope GREEN. Old3computations remain consumed; total scope/verifier
ceiling6. Each new computation max2private modules; no diagnostic host/target/
compile/profile-body/record-writer invocation, installs or extra agents/reviews.
Literal AST mutations and controlled runner throws are permitted within these
computations. No extra runtime/snapshot budget: remainder5hosts/7targets/1compile/
conditional1review, snapshot ceiling155,93already created, all still gated.

Keep exact historical scope RED307D8BC0…79C1A, bound to its old test/input. Do NOT
rerun availability/scope REDs or treat old failed GREEN as valid. New verifier RED
demonstrates the OLD comparison error, not a fresh runtime scope failure. A strict
new stop owner-scope-verifier.stop.json blocks all stages. Historical signature
stop1D8844C2…18F38 is recoverable only through exact supplement approval/hashes.
Any unexpected result stops without patch/retry/refreeze. Native method retained.
