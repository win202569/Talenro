# Token-local source preparation — 2026-10-18

Status: PREPARED ONLY / UNACCEPTED. Parent checkpoint:6b4b4a8a.
User approved source preparation and static checks, NOT execution.

## Saved artifacts

Directory:docs/diagnostics/isolated-fuzz/integer-token-local-candidate/

|File|SHA256 (physical LF bytes)|
|---|---|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|
|offline-supervision-policy.ps1|F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A|
|token-direct.tests.ps1|04B72510A4FA5D915A3A2E5AF024FA387E92C0FD83A25811AAAB0573FE1A3692|
|TokenDescriptorSupport.cs|BF82CE30B9CE82449C90ADAA96C0B03F3892609FB903EF87B07C7FD68489409A|
|source-pins.json|F72A570701A448394FEC5F4C427788C102E9DB712D640CE5EBF8E174D943698B|
|README.md|FAE6313F9E20FF87EB7CAA8B6CCAE3A821D31FFF74217443524013669FA26D14|

Baseline:integer-candidate-validation-copy-attribution/offline-supervision-policy.ps1,
SHA2563537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3.
The unchanged source header still says integer-local;README explains the
Token-only extension rather than violating outside-body identity.

## Static evidence only

Tests and support source prepared before candidate. PowerShell Parser.ParseFile
read both modules and new test as data:zero syntax errors. AST comparison:
13 functions and identical signatures;only Assert-SupervisionPolicyToken differs;
12 other exact function texts and all text outside Token body identical.
The original Passive primitive/member/descriptor guard block is copied intact.
Exact string refusal remains before guards, ordinal membership after guards;
metadata fallback stays first. No new caller trust/type parameter, coercion,
cache, global state, output or timers in candidate. C# reviewed as text only;
neither C# syntax nor compilation is verified.

Author self-review only, not independent review. Direct test literal outcomes,
owned strings/descriptors/payload checks, zero-callback counters, both metadata
paths, decorated-invalid ordering and mutation/clean isolation are prepared.
An unexpected baseline result or lost fixture decoration blocks comparison;
caller checks do not prove callee binding equivalence. CodeProperty support
failure blocks execution, never skips a case. Future runs select one pinned
baseline/candidate module only, without warmup/reload/retry/PS5.

No candidate/baseline/test/helper import or invocation, C# compilation, controller
creation, product/full API execution, diagnostic, benchmark or new target host
occurred. Runtime RED/GREEN, baseline oracle satisfaction, decoration retention,
engine binding equivalence, regressions and speed are UNVERIFIED.

Historical434 pins verified unchanged before preparation. Final preservation
checks also cover pinned manifest/controller/invocation/stdout/stderr identities
and staged checkout-filter bytes;no historical source/record is edited.

## Resume and gates

Recover from latest commit on codex/c12-b01-task9-coordinator, not main.
Portable artifacts need no ignored session path. No executable controller here.
Next step is a NEW bounded execution proposal/approval for separate fresh PS7
baseline/candidate hosts with explicit fail-stop limits;do not execute on this
preparation approval. Then resolve any baseline/fixture/binding discrepancy
without relaxing oracle or guards. Full API equality/refusal-index/result
ownership gates remain mandatory later. Passing direct tests alone would NOT
prove performance or acceptance.

All historical spent/frozen allowances unchanged, including attribution1/1,
candidate-full1/2 with PS5 frozen,review0/1unused/ineligible. No independent
review launched. Candidate/I1/P1/main/C12/security/runtime/release remain
UNACCEPTED. Routine Git checkpoint is preservation, not execution authority.
