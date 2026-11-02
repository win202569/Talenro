# Call-model amendment — actual execution and stop

Plan/design approved by user “是” after ae594fef; native execution.
NOT_ACCEPTED. No runtime implementation was performed in this attempt.

## Completed preparation

- Historical2251pins validated with writer-projection snapshots for31operational
  sources;155snapshots and seven historical stops preserved; new slots absent.
- Twenty-group regression and immutable input committed/pushed at b879c60c.
  TestSHA C5263B96182D3B450AC670A8CBE4415ECC43159A1479AB897B5F7D6AA4C5F6C0.
  InputSHA 3D07B813B64210CCF6880A7DB18D1D8C34F98F6A059471F1F5411B989A07CB6C.
- Preparation portable archive453files/125directories/31zero-byte files,
  decoded equality verified. No new runtime source epoch or runtime attempt.

## Actual first auxiliary result

contract-call-model-red.intent.json consumed computation10of11.
Actual exit2,0PASS/0FAIL/20NotRun, not expected RED0PASS/1FAIL/19NotRun/exit1.
Setup failed before ExactDescriptor. Actual error:
PropertyNotFoundException / PropertyNotFoundStrict:
“The property 'SHA256' cannot be found on this object. Verify that the property exists.”
Stack: extracted Get-ODPinnedSHA256 line5 → Read-ODSourceModel line5 → frozen
contract-call-model.tests.ps1 line60. No blanket catch reclassified this as
valid expected rejection. RecordingError null; PostSourceSHA256 empty in original
receipt, retained without reconstruction. Full original tool result preserved.

New eighth contract-call-model.stop.json blocks subsequent work under the
approved first-unexpected-result rule. No GREEN, contract freeze/host, other
runtime stage, fixture compilation, target, or reviewer ran.

## Readonly diagnosis and boundary

The selected record-boundary source pin exists uniquely and includes its valid
SHA256. The actual source function uses $matches as the pin-array variable,
then performs -cnotmatch before reading $matches[0].SHA256. Static reading and
the actual error trace point to collision with the automatic regex $Matches
variable, whose name is case insensitive. This is an inferred root cause, not
an additional executed reproduction or verified fix. Repair should use a
non-automatic local name and preserve all pin state/hash/conflict checks.

All three runtime source hashes were independently read after the stop and
still match the approved baseline; this is separate evidence, not backfilled
into the failed receipt. The model's pre-existing setup defect was not covered
by the frozen plan's anticipated RED signature. Do not bypass it, patch/retest
the frozen regression, or call this the expected descriptor failure.

Consumption:10auxiliary/3hosts/3compiles/155snapshots/0targets/0reviews.
Remaining1auxiliary GREEN plus4hosts/7targets/0compiles/31snapshots/conditional1review
are frozen. Accepted writer remains accepted at its original boundary, not
rerun or regraded. Whole I1/I2/product remains NOT_ACCEPTED, no main merge.

## Proposed next decision — not execution authority

Authorize a bounded recovery revision covering the pin-array variable collision
within the existing source-model file, a separately identified regression for
this newly observed failure, and explicit new attempt/allowance accounting.
Preserve this failed attempt, frozen tests, all stops and archives. Do not
silently reuse the spent RED slot or the original expected failure signature.
No recovery design/plan, extra attempt or stop override is approved by this note.

## Portable handoff

NEW portable-contract-call-model.json:457actual P/Q files/125directories/
31zero-byte files; decoded original bytes, lengths, hashes and directories
verified. SHA256 01B4748BB6E672DB9D8170B65ACEC0D800DDEF48058DA7AE5392DA9E5FFB37E7.
Earlier archives unchanged. Ordinary Git includes the frozen regression and this
report; ignored raw files are preserved only through the portable archive.
