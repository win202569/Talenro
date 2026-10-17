# One hardlink attempt — structured access-denial evidence

## Authority and method

After6fb486f1,user approved the [minimal probe](2026-10-17-synthetic-capture-hardlink-readonly.md):
NEW maximum1 PS7 child,120000ms total,no retry. One unique diagnostic directory
under the existing ignored fixture root,one ordinary literal file,one attempt
using New-Item -ItemType HardLink. No elevation,alternate API,ACL/ownership or
security change,cleanup,product import,archive assertion or original-suite replay.
The new1/1 allowance is now consumed;all frozen test/review states remain.

The [exploratory script](../diagnostics/isolated-fuzz/synthetic-capture-hardlink-diagnostic.ps1)
sets explicit UTF8 console logging;the parent also uses UTF8 stream decoding.
The file is CreateNew/four literal bytes TLR1 and closed before the link attempt.
Source content/preexisting destination guards run before and after. Error logging
records exception chain(type,HResult,available native code,message),category and
FullyQualifiedErrorId. Permission denial is a diagnostic outcome,not a passed
archive/hardlink-rejection test;unexpected setup/error/timeout would stop.
Script and parent were statically parsed before the single launch.

## Observations

PS7 observed7.6.5Core x64,administrator role=false. Child exit0/745ms,no timeout,
empty stderr,four ordered structured rows and terminal diagnostic message.
Ordinary file creation succeeded;source remained4 bytes/Base64 VExSMQ==.
Exactly one link attempt returned Outcome=refused and no destination exists.

|Captured field|Observed value|
|---|---|
|ExceptionType|System.ComponentModel.Win32Exception|
|NativeErrorCode|5|
|HResult|-2147467259|
|HResultHex|0x80004005|
|Message|拒绝访问。|
|Category|NotSpecified|
|FullyQualifiedErrorId|System.ComponentModel.Win32Exception,Microsoft.PowerShell.Commands.NewItemCommand|
|Inner chain|one exception,no inner exception,not truncated|

The native code and HResult are DIFFERENT fields;do not relabel0x80004005 as
0x80070005 or infer an unobserved inner error. The explicit text handling retained
the localized message correctly,unlike the earlier damaged slot7 stderr. That
old receipt was not rewritten. Parent exit0 means diagnostic capture completed,
NOT that a hardlink was created or the safety test passed.

Retained fixture directory:
`.superpowers/diagnostics/offline-synthetic-capture-store/hardlink-probe-6523cd92b8ff4c7d90d8113e04aefa16`.
It contains source.bin only;link.bin is absent. No fixture/receipt deleted.
This controlled fixture uses the same command/root but a new directory and
four-byte source,with link/source in that directory. It is NOT a byte-for-byte
replay of the old308-byte source/root-level destination or its historical token.

## Interpretation and next boundary

We now have direct structured evidence of access denial specifically at link
creation while ordinary file creation succeeds in this controlled context.
The archive reader's hardlink-rejection assertion was not involved. This narrows
the failing operation,but generic native5 alone does not identify whether
effective access,sandbox policy,filesystem filtering or another layer denies it.
It does not prove that the historical slot7 has the identical native code/cause.
No product fix or test skip is justified by this probe.

Further progress requires a separately approved execution-context investigation
or controlled comparison,with scope/permission boundary stated first. No
outside-sandbox/admin comparison,extra link attempt or original rerun is
automatically authorized. Do not change ACLs,take ownership,disable security,
switch to an alternate link API or mark coverage accepted to get past this gate.
Frozen partition8..10/original7..12 and both unstarted reviews remain untouched.

## Preservation and gates

### Pending bounded execution-context comparison

User requested continuation;next short spike is presented in chat,NOT approved
or executed yet. Propose NEW maximum1 PS7 child via the tool's separately approved
outside-sandbox execution path,120000ms,no retry. Use the unchanged pinned probe
above,new dedicated fixture and exactly one ordinary-file/one link attempt.
Compare with the already retained normal-context native5 result;do not rerun it.
Before any child/fixture,check the approved parent token is the expected current
user and not in the Administrators role;otherwise stop without trying. No RunAs,
Windows administrator launch,ACL/ownership/security change,alternate API,cleanup
or original-test replay. If tool approval is denied/unavailable,stop:do not switch
routes or retry. Record context/request outcome and keep a separate index.

Created/refused outcomes are diagnostic only. A difference would support a
context-dependent restriction,not isolate every token/policy difference or grant
permission to run the full suite outside sandbox. If both refuse,cause remains
unresolved. Product/tests and all frozen slots/reviews stay unchanged. This
proposed1 slot is separate from the spent1/1 normal-context diagnostic;no new
probe source or child has been created this turn.

Probe SHA256 at launch/postrun:
`56273A58C4E91ED7D1A41AFE8F97C198A1C829624C1C1521DF27FF5B2DB32407`.
New one-entry index:
`36E9CC361787F65E589DAC1B332E70CB7BE53575D09261C856A23429477BEE1A`.
Unchanged store/test:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C` /
`406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`.
Partition/original indexes unchanged:
`B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1` /
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Postrun raw/index rows,probe hash and retained source/absent link were checked
readonly without a new child or link attempt. Raw/index/runner/progress stay
ignored under `.superpowers/sdd/2026-10-17-synthetic-capture-hardlink`.
Git restores portable probe/report,not local fixtures/caches/launch authority.
The script is exploratory,not a supported runtime/test entry or product feature.

No repair,GREEN,full acceptance,main merge,real capture/security/vendor clearance,
Stage B/I2/I3/full C12 or expanded permission. Save unfinished development progress.
