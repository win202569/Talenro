# One approved administrator-token hardlink diagnostic

## Authority and result

After checkpoint8fbe6049,the user selected option1:NEW at most one PS7 child
with administrator token,120000ms,one ordinary file and one link attempt,no
retry,ACL/security changes or original-suite replay. This is separate authority
from the closed non-admin context request;its guard and ledger were not edited.
The unchanged pinned diagnostic's historical no-elevation header is superseded
only for this explicitly approved invocation,not for future runs or full tests.

The new parent runner was parsed before launch,then run once through the tool's
approved elevated execution path. Expected username matched and parent admin
role=true;child setup independently recorded admin role=true. PS7.6.5 Core x64
exited0 in506ms,no timeout,empty stderr. Four ordered structured rows and the
diagnostic terminal marker were present. One four-byte source was created and
exactly one New-Item HardLink attempt returned created,LinkType=HardLink,error
null. Readonly follow-up confirmed source/link each4B with Base64 VExSMQ==.
No retry,alternate API,RunAs invocation,ACL/security change or deletion occurred.
This NEW1/1 diagnostic allowance is spent. All fixtures and raw receipts retained.

## Interpretation and next boundary

Earlier normal-context execution of the same pinned probe returned native5,
Win32Exception,HResult0x80004005;this administrator-context invocation created
the link. This supports execution-context dependence but does not isolate token
privileges from sandbox/effective-access/filter differences. It does NOT show
that Windows hardlinks universally require administrator rights,prove the exact
original denial cause,or fix/accept the archive implementation. These minimal
four-byte,same-directory fixtures differ from the original308B suite fixture;
the original failing suite case was not replayed.

Recommended next action,NOT authorized/executed here:agree a new bounded
reader-paths verification packet with explicit execution context,host count,
timeouts,expected assertions and stop conditions. Do not silently elevate the
suite,skip hardlink/junction/locking assertions or consume frozen slots. A
diagnostic link success is not a reader rejection assertion or module acceptance.

## Preservation and unchanged gates

New ignored ledger directory:
`.superpowers/sdd/2026-10-17-synthetic-capture-hardlink-admin`.
Index SHA256:
`61DE8914D3E6351A0B3085046ADA4C27EE867096E51F65E9D72504ABDE270E22`.
Retained fixture directchild:
`hardlink-probe-7323c84254e24e6a838d8b903aeeceaa`.
Raw receipts,parent runner and fixture are local/ignored,not transferred by Git.
The portable unchanged diagnostic and this result/limits are recoverable from
the existing development branch;cloning does not authorize another launch.

Readonly rechecks confirmed unchanged SHA256:

- Probe:56273A58C4E91ED7D1A41AFE8F97C198A1C829624C1C1521DF27FF5B2DB32407.
- Product:508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C.
- Tests:406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF.
- Normal probe index:36E9CC361787F65E589DAC1B332E70CB7BE53575D09261C856A23429477BEE1A.
- Closed context guard index:45E70AA65843D11E814E3E361C6A57924E429FF8EA22182F82422175CE53269E.
- Partition index:B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1.
- Original index:ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A.

Partition7/10 spent,8..10 frozen,new review unstarted. Original6/12 spent,
7..12 frozen,dependent regressions unrun,original review unstarted. No main
merge,full acceptance,live capture,security/vendor clearance,Stage B/I2/I3/full
C12 completion follows. Save only an unfinished development checkpoint.
