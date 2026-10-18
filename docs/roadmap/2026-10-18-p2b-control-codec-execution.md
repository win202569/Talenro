# P2-B pure control codec — INCOMPLETE checkpoint

Date: 2026-10-18. Base: fab8466e0ce15f4a1d98e0274852433ac9e435f6.
Branch: codex/c12-b01-task9-coordinator. No main merge or acceptance.

User approved Native implementation and exactly five new ordinary PS7 hosts
plus one readonly reviewer conditional on five exact outcomes. Implemented
three pure codec/binding APIs, tests, twelve independent literal frames, exact
output oracle and bounded controller in
docs/diagnostics/isolated-fuzz/p2b-control-codec/ (eight portable files).
No real lifecycle, identity, capability, queue, clock or release implementation.
All product authority flags remain false. P2-A scoped acceptance unchanged.

## Actual execution and stop

| Slot | Actual result | Disposition |
| --- | --- | --- |
| red | PID39128,1759ms,exit1,3lines,API assertion only | Exact expected RED |
| full | PID24072,2597ms,exit1,16lines,13PASS then PayloadLength failure | Unexpected; STOP |
| reserved-tail | Not launched | FROZEN |
| sequence-equality | Not launched | FROZEN |
| restored final | Not launched | FROZEN |

Consumed2/5; remaining3 frozen. Reviewer0/1, ineligible; none dispatched.
Both hosts ordinary Windows PS7.6.5 Core x64, ParentAdministrator=false;
no timeout, owned-host-exited, complete streams, empty stderr, pins intact,
ControllerError=null, FormalAcceptance=false. Full FirstMismatch=15 (zero-based).
Controller's parent stop exception is not target stderr. No mutants applied,
no restoration necessary. No repair, retry or further product import followed.

Saved full stdout after the two context/host lines:

```text
PASS: P2BCodecApiRequired
PASS: P2BCodecHelloLiteral
PASS: P2BCodecEncodeAllKinds
PASS: P2BCodecDecodeAllKinds
PASS: P2BCodecRoundTripSecondary
PASS: P2BCodecExactMessageShape
PASS: P2BCodecExactTypes
PASS: P2BCodecPassiveDescriptors
PASS: P2BCodecDecoratedScalars
PASS: P2BCodecByteArrayShape
PASS: P2BCodecFrameLength
PASS: P2BCodecHeaderMagicVersion
PASS: P2BCodecReservedZero
ASSERTION FAILED: P2BCodecPayloadLength
```

Root cause NOT established. The frozen catch records only the group, not the
underlying exception or assertion. Readonly source locates the failing group
at test lines195–201; it tests malformed per-kind lengths and nonzero padding.
This evidence cannot distinguish a fixture/expression failure from a codec
defect. Do not relabel as expected, weaken the oracle or claim later groups pass.
Groups14–32 not proven; no mutation sensitivity, restored final or review proof.

## Frozen identities and readonly reconciliation

All868 mapped entries were rehashed through VerificationPath after the stop:
zero mismatches. Original plan snapshot, spec, inherited records, seven new
premodule inputs and sealed good module preserved without edits.

| Artifact | SHA256 |
| --- | --- |
| Approved plan snapshot | 15E7F4D33673D8DF2C9C2CC577BDACADA08797B4C34599199BABE2E8DD80D094 |
| Spec | 41BD5F7225D65D7BAE903E693946A2595C5036A2408B16BECE37965C622E36CE |
| Tests | DE23854DF2EB1F2F084625C0CD84312A4024FEA73BD7D51C8316C7409B769E4A |
| Module / good-product | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| Engine | 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 |
| frozen.json | C240C986D107AF7C6E15C9543936499D43E547CC123159BFCB0351407021A77A |
| product-seal.json | B18E9B6B14BB80FE0010A45216A1631A759EA10329896D32D86766526328676D |
| invocations.json | 6674962585BF6A4FFF80399283D765930FC17659D21A6E5FB7AE6A7DF591B14E |
| red.stdout.bin | DA2707E247C413974EDA6FB424DA4223ED7FC8B66EF4E32E622DEE125C66C575 |
| full.stdout.bin | DC8BA4DCA9714C34204DF612344D422A85CAF836FFF7F6583B90A0D63B58DE62 |
| Both empty stderr files | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Raw packet remains local/ignored at
.superpowers/sdd/2026-10-18-p2b-control-codec/; no raw file was rewritten,
deleted or committed. The ledger and live plan status are updated separately
from the immutable approved-plan snapshot. Report is a portable transcription,
not a substitute for raw receipts on a fresh machine.

## Scope rulings and recovery

Windows-only PowerShell-native ledger/controller replaced Bash skill scripts;
no generic script-completion evidence claimed. Explicit finite scope overrides
generic extra baseline/global test/fix/retest/reviewer/cleanup loops; cost is
narrower, unfinished verification. Systematic debugging required evidence-first
readonly reconciliation; it did not authorize instrumentation or another run.
Task1 sources and five proposed receipts matched Task2 review inputs at preflight.

Necessary files are saved as an unfinished ordinary development checkpoint.
Recover latest development branch, not main; read current-status and recovery.
Cloning does not restore raw packets, historical mappings, engine caches or
sessions, and does not reset frozen allowances. No automatic Prepare/Run/replay.
Next proposal: separately bound diagnosis of PayloadLength, preserving this
source/oracle and failed packet. No new diagnostic or repair is authorized by
this report. No PS5/P2-C/P3/C12/live/security/runtime grant or acceptance.
