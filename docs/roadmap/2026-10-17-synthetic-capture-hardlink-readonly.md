# Hardlink fixture denial — readonly findings and pending minimal probe

## Readonly scope

Afterf99ec699,user requested continued investigation. This turn read existing
slot7 receipts,source,the candidate failed fixture metadata/ACLs,drive metadata
and current orchestration identity. No child test/diagnostic,link creation,
fixture write,ACL/ownership change,privileged rerun or cleanup occurred.

The failed reader-paths run started2026-10-17T14:53:06Z;latest fixture directories
share prefix synthetic-0bbfdbcb53dc4f07875ca41f62b20ac1 and creation14:53:09Z.
The final directory numbered28 matches source-path progression:18 malformed
archives+8 marker archives+existing-target archive+first hardlink-source archive.
This correlates the retained fixture with failure;the old child did not log
its exact resolved link arguments or native error object before termination.

## Observations at inspection time

- The correlated directory's record.bin exists,308 bytes;complete.bin exists,
  52 bytes. Their attributes are Archive,not ReadOnly/ReparsePoint. The fixture
  directory and its archive root show ordinary Directory attributes.
- Expected destination synthetic-0bbfdbcb53dc4f07875ca41f62b20ac1-link-record.bin
  under that same fixture root does not exist at inspection time.
- D is reported Fixed/NTFS/ready by DriveInfo. Source and intended destination
  are both under the same archive fixture root in the source command.
- Root,directory,record and marker ACLs are inherited/unprotected,administrator
  owned,with no explicit Deny entries in the inspected rules. Modify grants
  exist. No ACL/owner was changed. No full ACL/SID dump is committed.
- The current orchestration token is not in the Administrators role. That is
  a current observation,NOT a reconstruction of every restriction in the failed
  child token. Administrator ownership does not itself establish causality.
- Failed operation is New-Item -ItemType HardLink at current test line140,
  before the first hardlink refusal assertion. Saved localized stderr already
  has damaged text decoding;exit1/location are retained,not a reliable native
  code/exception subtype. The receipt remains unmodified.

Missing source/preexisting destination/read-only attributes are not supported
by the retained state. These post-failure observations do not prove what was
true at every instant or establish an effective access check. Absence of a DACL
Deny rule does NOT rule out restricted-token,filesystem or execution-sandbox
restrictions. No verified attribution to ACLs,sandbox,PowerShell or security
software is made;no rationale for disabling security or taking ownership follows.

## Pending diagnostic spike — awaiting approval,not constructed/executed

Question:what structured exception/native-code information does the SAME
New-Item HardLink operation return in one fresh controlled fixture under the
current normal execution context?

Propose NEW maximum1 PS7 child,120000ms total,no retry. In a new uniquely named
dedicated diagnostic child directory under the existing ignored fixture root,
create one literal ordinary file and attempt exactly one hardlink using the
same New-Item operation. Capture structured exception type,HResult,available
inner/native error code and FullyQualifiedErrorId with explicit UTF8 logging,
not inference from localized console text. Keep original receipts untouched.
No alternate API/link mechanism to bypass a denial. No elevation,ACL change,
ownership change,security-setting change,delete or original-group replay.

Normal-file setup failure or unrelated unexpected error stops the probe. The
single link attempt may succeed OR be refused;either is a diagnostic outcome,
not an accepted hardlink safety test. Preserve fixture and raw evidence in both
cases. If refused and code remains generic,the conclusion stays unresolved;no
automatic outside-sandbox/admin comparison or retry is granted. Any further
action requires its own scoped approval. The frozen partitions are not resumed.

## Preservation and gates

No new diagnostic code or allowance consumption exists this turn. Product/test
and new seven-entry index unchanged:
test `406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`;
index `B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1`.
New partitions7/10 spent,8..10 frozen,new review unstarted. Original archive
6/12 spent,7..12 frozen,original review unstarted;prior diagnostic budgets spent.
No repair,GREEN,full acceptance,main merge,real capture/security/vendor clearance,
Stage B/I2/I3/full C12 or expanded permission. Git restores this report,not local
fixture state or permission to run the pending probe.
