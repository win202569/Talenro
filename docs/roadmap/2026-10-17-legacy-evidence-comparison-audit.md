# Legacy evidence comparison audit — readonly source findings

Follow-up:[classifier-only correction](2026-10-17-classifier-exact-markers.md)
reproduced classifier XML/encoding concerns and passed87 classifier checks per
host;readonly review0Critical/0Important/0Minor accepts bounded checkpoint only.
Sources/hash/static status below describe the
original audit checkpoint,not the new classifier. Package/store remain unchanged;
package exact-label/callerGetType concerns still unresolved. No integration or
runtime acceptance follows from classifier-only results.

## Scope and evidence level

User requested continuation of readonly comparison inventory after17c52edf282ded32fc366a36247e6db4283f9d1c.
Read complete classifier/package/store sources,their tests,relevant prior
reports and current contract/simulator call paths. No module imported/sourced,
test or diagnostic host executed,XML parsed,input constructed,store opened or
security state queried. No implementation/test edits. Only this report and
status/recovery are preserved under the standing cross-computer documentation
rule. Author source inspection only,no new independent review.

Current source SHA256:

- Classifier:A3C8A01BD62ADB68EBE2F3ED8777D8258CE8BC0EF75CED55B33053FDE7A13590.
- Package:B2CA5135E9089AC6EE2A34A0F95F178BAFF98A333BEB78C6ACAB85E915C57E2D.
- Store:7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81.

Prior contract correction dynamically established linguistic/ordinal divergence
onPS7/PS5;see [110-check evidence](2026-10-17-capture-contract-ordinal.md).
That does NOT dynamically reproduce the legacy paths below. All legacy effects
remain static reachable risks/candidate fixtures until separately tested.

## Comparison-path inventory

|Area and source lines|Reachability and concern|Existing restrictions / disposition|
|---|---|---|
|Classifier9–11,29–39,63–81:XML names,namespace,attribute values,type/progress text|Caller stderr becomes XmlNode strings;channel progress,SourceId/Record,Completed/Processing,type names and envelope values use linguistic comparisons|XML grammar rejects some name variants and NUL;it does not require exact approved attribute/text values. Prioritize legal XML text/attribute ignorable-character fixtures,not invalid NUL XML. Static risk,not fresh reproduction|
|Classifier104–105:EncodingName branches|Standalone caller reaches linguistic utf-8/utf-16le checks|The new contract already validates exact encoding before delegating;standalone API/package admission is not covered by that preflight|
|Package8,16,53,58,62:encoding,shape,policy,classification|Standalone New-/Test-OfflineOutputEvidencePackage sees caller strings without new contract guards;non-exact accepted labels may survive recomputation if equivalence is linguistic|Exact primitive types/false flags remain. Policy/encoding/classification checks should be investigated under separate fixtures;shape lookup may independently reject malformed names,so no universal name-bypass claim|
|Package26–31:Base64/hash comparisons|Untrusted strings reach comparisons only after strict bounded ASCII regexes and reconstruction|The canonical alphabets exclude the Unicode equivalents motivating this audit. No analogous bypass established here;ordinal consistency could be defense in depth,not automatically a reproduced defect|
|Store9–12:path checks|Canonical Windows path equality uses OrdinalIgnoreCase explicitly|Intentional path semantics,not the same linguistic-token issue. Do not blindly make Windows paths case-sensitive or weaken isolation checks|
|Store52,59:magic;57,111:hashes|Magic is exactly8 ASCII-decoded bytes;hash comparisons use Base64 generated from raw hashes|Fixed length/ASCII or generated alphabet materially narrows input. No demonstrated same insertion/equivalence bypass. Preserve strict byte format,not an automatic schema migration|
|Store60–68:read encoding|Stored Int32 must be1/2,then maps to internal exact literal|No caller text encoding admitted on this read path|
|Store88–93:write package preflight/encoding mapping|Writer relies on standalone package validator,then linguistic encoding-to-enum branch|Inherits package admission risk;could canonicalize a non-exact declaration into a valid enum. Static concern,not a newly executed roundtrip. Fix upstream before deciding a store-format rewrite is needed|

The new contract's exact metadata/derived comparisons do not inspect semantic
XML markers inside raw stream bytes. Its simulator still calls the classifier
for complete output. Therefore XML semantic comparison risk is reachable through
the current complete-record path,not only unsupported direct calls to internals.
Hashes retain supplied bytes;they cannot make over-broad classification strict.
False authority flags still prevent a classifier label from granting acceptance.

## Coverage and adjacent input-boundary concern

Current classifier tests cover ordinary wrong channels/types/version/namespace,
malformed/trailing XML,DTD/comments,instructions,canonical numeric endings,
Unicode display text,strict decoding and bounds. Package tests cover ordinary
metadata/shape mutation,hash/Base64 consistency,known encodings and non-object
input. Store tests cover local binary roundtrip,valid-hash malformed fields,
encoding enums,bounds,links,reparse paths and locks. No ignorable-character
protocol-equivalence fixture was found in these three suites. Reading tests is
not rerunning or certifying them;all previous PASS/failure history stays intact.

Separate static observation:package12 invokes caller Value.GetType() before it
establishes a plain NoteProperty shape. This can dispatch a caller ETS method,
unlike the safe reflected-base strategy in the corrected contract. The store
writer reaches this package preflight too. No callback was executed or counter
fixture run in this audit. Treat standalone non-dispatching shape validation as
a separately scoped boundary concern,not silently fixed by ordinal string work.
No new claim is made about every other standalone hostile-object/length boundary.

## Recommended order — not implementation authority

1. First bounded classifier correction:legal XML ignorable-character values and
   explicit encoding tokens must be refused;ordinary Unicode progress display
   text and original byte hashes stay unchanged. No package/store/schema/path
   refactor. Propose four NEW hosts PS7RED/PS5RED/PS7GREEN/PS5GREEN,120s each,
   no retries,stop without code change if reproduction absent,then one readonly
   review. Unexpected outcomes stop;no hidden fix reserve. Await approval.
2. Separately design/package-boundary tests for exact labels and nondispatching
   shape validation. Do not combine an untested callback fix into classifier
   work or claim new contract guards protect standalone callers.
3. Reassess writer inheritance after upstream repair;only then design complete
   capture-record persistence. Existing Package store omits manifest/event trace/
   control evidence. No reason established to rewrite its binary format now.

No new verification allowance follows from this document. Old4/4,repair3/4 with
frozen fourth,normal-fix3/3,ordinal4/4 and all earlier budgets remain unchanged.
No real collector/watchdog,actual command/clock provenance,persistence migration,
security/vendor clearance,Stage B,I2/I3,full C12 acceptance or main merge. These
source findings do not invalidate recorded110-check results;they delimit what
those synthetic contract tests never claimed to cover.
