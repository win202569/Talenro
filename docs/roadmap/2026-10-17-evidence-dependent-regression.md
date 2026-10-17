# Evidence dependent-chain regression after package boundary correction

## Scope and execution rule

User accepted the recommended compatibility regression after development commit
`362075604b905ffc5662b1c2e3f4a989c37ccfd3`. Six NEW maximum sequential hosts:
store PS7/PS5,simulation PS7/PS5,contract PS7/PS5;120000ms per host,no retry,
stop at first unexpected result. No repair reserve or source/test modification.
All previous spent/frozen allowances remain unchanged. Existing development
branch only;no main merge or wider acceptance.

Store suite uses local synthetic files,hardlinks and a junction under its
dedicated ignored fixture root. Fixtures and all text receipts are retained.
Simulation and contract use supplied synthetic events/declarations,not live
collection,OS watchdog or child cleanup. Each test suite has a fresh host;
this does not exercise cross-suite shared-session pollution.

## Results

All6/6 completed successfully without retry. Total540 PASS,zero FAIL and empty
stderr in every host. Observed simulation/contract host banners7.6.5Core x64
and5.1.26100.9711Desktop x64. Store uses the same resolved executable paths
but does not print its own host banner;do not claim a missing independent banner.

|Call|Suite|Host|Exit|Elapsed ms|PASS|FAIL|Stderr characters|
|---|---|---|---|---|---|---|---|
|1|store|PS7|0|13592|54|0|0|
|2|store|PS5|0|2626|54|0|0|
|3|simulation|PS7|0|2760|106|0|0|
|4|simulation|PS5|0|2420|106|0|0|
|5|contract|PS7|0|37850|110|0|0|
|6|contract|PS5|0|102530|110|0|0|

Expected complete-suite counts are54 store,106 simulation,110 contract per host.
Only exit0,empty stderr,zero FAIL,exact count and terminal LIMIT together pass.
The PS5 contract run completed within its120000ms cap,not a performance target
or guaranteed future timing. No timeout,retry or source/test edit occurred.
Source/test hashes were rechecked unchanged after all6 runs. Classifier/package
standalone suites were not rerun in this allowance;their prior results remain
separate scoped evidence,while current consumer suites exercised dependencies.

Store's independent binary-layout literal,UTF16/raw-byte roundtrips,reconstructed
classification,missing/malformed marker/payload rejection,exact bounds,invalid
origin refusal,hardlink/junction/exclusive-writer refusals all passed on each
host. Simulation and contract tested event lifecycle/admitted prefix,byte and
trace limits,Unicode/raw progress,exact contract tokens,passive callbacks and
tampered-record refusal as already present in their suites. This is existing
coverage,not newly written store-side adversarial regression cases.

## Evidence and limitations

Source/tests come from the base commit above with no implementation changes.
Dependency SHA256 at execution:

- Classifier:`F323C9C46190DE06F25BFA55A2B7983707B1FF5877548F9703870F9BB771A0BF`.
- Package:`857543A9DDA9B9C849B2507C6EE419C1F73AA0D0C8599CD9D88738841D021198`.
- Store:`7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81`.
- Store tests:`B7A49A0F0FB2D91B7968115F8FEEAFA523564C0582AA0DD905A762BE7CAA0159`.
- Simulation:`9DA8A74B984A3111F00B54AA160490DF327BE17A5A07A044B510946EB26A2A13`.
- Simulation tests:`B6E7E342A76FB6BD2AEE6641BAF65305846071A2562F0CD97AC4ECD1806CE7FE`.
- Contract:`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2`.
- Contract tests:`5A6975AAF9998D62304BA3DDB7483A6CE1A1382FBE9D05EBDA209BD7AE71566F`.

Ignored local run index/stdout/stderr and bounded runner live at
`.superpowers/sdd/2026-10-17-evidence-dependent-regression`;these are text
receipts,not raw live collection/provenance artifacts. Portable tracked tests
do not depend on that directory or private cache contents. A new computer needs
its own PowerShell hosts and explicit bounded execution scope,not automatic
replay of this consumed allowance.
Six-entry invocation index SHA256:
`CDED008245202FB573C14834A714D2723901DEB6EC14DAD8EC8C98C7200FAA55`.

This run checks existing consumer behavior against the current package and
classifier dependency versions. It does not add new adversarial cases or prove
every writer/caller boundary. No new independent review or full-project suite;
previous scoped review conclusions are not expanded by consumer test results.
No full capture-record persistence format,atomicity/fsync/power-loss promise,
command/clock/origin authenticity,live collector/watchdog/OS cleanup guarantee,
security/vendor clearance,Stage B,I2/I3/full C12 acceptance or main merge.
OriginVerified/Acceptance/RuntimeApproved remainfalse.

## Next safe work

The previously unrun consumer-regression gap after36207560 is now closed for
these three existing suites/two hosts only. Next examine the store writer's
inherited admission using focused synthetic boundary cases before deciding
whether any local store correction is necessary. No store format rewrite is
justified by these passing results;full capture-record persistence is separate
design work. New implementation/tests or further counted hosts require their
own bounded scope;this6/6 allowance supplies no repair reserve or replay.
