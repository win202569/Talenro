# P1 offline timing fixtures — 36/37, checker defect, not accepted

Date:2026-10-18. Base:a90555f6cc10c1c49dc5b4790e1cc32ecb119de0.
User approved the recommended OFFLINE logger/checker validation,not a candidate
diagnostic. One fixed37case offline batch ran on the pinned ordinary PS7 engine.
No candidate import,decision/scalar helper call,full ownership test,old frozen
slot use,reviewer,repair or retry. No main merge or wider acceptance.

## Actual outcome

37cases completed:36PASS,1FAIL. Exit1,elapsed16246ms,no timeout,owned host
exited,streams complete,stderr0bytes,pins intact,controller error null.
The bounded controller correctly stops on the non-green batch.
No source was changed after launch and no case was re-executed.

Failure:duplicate-json-property — "invalid capture was accepted".
The standalone checker accepted an ambiguous duplicate-field TIME record as
a valid PREFIX_ONLY input. This is NOT evidence that it granted a full-suite
pass or runtime/security acceptance;its receipt/runtime/acceptance flags remain
false. It nevertheless violates the intended strict fixed-record contract.

Failing synthetic input begins:

    TIME: {"Schema":0,"Schema":1,"Sequence":1,"Case":"startup", ...}

The rest is the otherwise valid first module-load entry with zero call counts.
offline.tests.ps1 reconstructs the full portable witness without raw packets.

## Cause and bounded next step

The checker at B3C6FE88 calls ConvertFrom-Json BEFORE comparing the deserialized
property-name inventory(line20then21). The captured duplicate-key witness shows
that this path accepts the final valid Schema value while the earlier duplicate
is no longer represented in that subsequent inventory check. Post-deserialization
property validation alone therefore does not establish raw-key uniqueness.

The FACT path uses the same parse-then-inventory pattern(lines56/57). Its duplicate
key behavior was NOT separately exercised here;it is a source-visible related
risk,not a second confirmed failure. A future fix should cover both TIME and FACT
raw-field uniqueness,including escaped equivalent names/case rules,without
weakening existing numeric/order/accounting/property requirements.

Recommend separately approving a minimal checker-only fix plus a new offline
regression batch. Preserve this RED witness/current checker bytes;do not patch
them simply to rerun this consumed controller. No candidate timing run should
be proposed as accepted while this checker defect remains unresolved.

## Passed coverage and limits

The36successful cases cover:

- Valid synthetic complete text classified COMPLETE_TEXT_ONLY,with false
  receipt/runtime/acceptance flags;closed/open/empty prefixes remain PREFIX_ONLY.
- Complete-mode rejection of incomplete text,including lowercase mode spelling.
- Missing/duplicate/reordered entry/leave records;negative elapsed/duration,
  changed/zero frequency,wrong stage/body totals,emission overlap.
- Completed>attempted,decision/scalar/metadata count bounds,float/decreasing
  counters,wrong stage counts,malformed/truncated/extra-field/wrong-stage records.
- Missing/false ownership facts,reordered base output,timing-only text and a
  missing final base line not admitted as complete text.
- Actual logger helper output preserves a single raw sentinel reference with
  no extra pipeline objects;its emitted first pair is accepted as a prefix,
  with positive previous-emission cost and last-emission duration unknown.
- Logger rejects nested stages and mismatched stage ends.

Only AST-selected definitions T-Emit,T-Begin,T-End were executed from the prepared
source inside an isolated temporary module scope. The instrumented test was
parsed as DATA,never dot-sourced. No product module or fixture/oracle functions
were loaded. Synthetic full text is protocol data,NOT candidate results.

The sentinel check does not prove every PSObject/String wrapper binding path.
Negative-case success means the malformed fixture was rejected,not a proof of
all possible malformed encodings or exact error classification. No baseline
performance,real owned-return identity,timeout attribution or product-mutant
sensitivity proof follows. Candidate imports/product calls remain0by construction.

## Portable source and preserved raw anchors

[Offline harness/controller](../diagnostics/isolated-fuzz/offline-timing-validation/README.md)
is NEW,not a modification of prepared logger/checker or old frozen leaves.
Four portable files were pinned before launch:

| File | SHA256 |
| --- | --- |
| .gitattributes | A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9 |
| capture.ps1 | 1E4EB1E7C7C4C63F13CD9E123F1111A38F1DBFCE39EC9AA771380EDE863710F6 |
| offline.tests.ps1 | EB01C6CBDF1F4C6ABEE94DE0F422C516C13E225C8B9F2E72962097D747434FED |
| README.md | 32C316BF6B4E0B82464EEAA9A50D536FB84D3E2C46D754C9798FB124E0DABB9F |

597frozen entries=582old+4previous packet anchors+7prepared files+4offline files.
All rechecked intact after the batch. Candidate1971F808 and checkerB3C6FE88
unchanged. New source creation had a tool-string quoting error caught before any
file write/launch;static parser checks passed before the single offline run.

Ignored packet:.superpowers/sdd/2026-10-18-p1-offline-timing-validation/

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 468E8FC26CEB6FBC2694E67BE9C32CB882DC9A18E7F78C732840C353AA4A5474 |
| invocation.json | DDBB274A6276829EF225A84A3FED55CED4699C8CB66E7BB47854B21120322186 |
| stdout.txt | A80672787C902D285414569DF545BAA9DA4D4461EF59DAB822C00FB4635E44BB |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| dup-property.synthetic.txt | 459400F661D6EA2907330D6C991F9FB5580A3E7E677CD18A46617C4B06887038 |

42stdout lines:36PASS+1FAIL+4actual logger TIME records+1terminal summary.
These scratch-stage TIME records validate logger guards,not the ownership
StagePlan;the separate captured first pair is inspected inside its fixture.
The original streams and synthetic inputs remain unedited and ignored.

Offline batch1/1consumed. Existing candidate-focused1/3spent+2/3FROZEN_UNRUN
and all older frozen budgets unchanged. A clone restores harness/progress,
not ignored packets or any right to rerun. Current preparation checklist is a
historical record;do not turn its offline gate green with36/37.

No I1/main/security/runtime/C12/release/PS5/performance acceptance.
