# P2-B SDK reserved metadata — read-only origin investigation

2026-10-18; base cd6a891942a52b06282b77e933f1cb420c027b65.
Branch codex/c12-b01-task9-coordinator. User approved continuing the recommended
read-only SDK-origin/identity investigation. This is research and a proposed
bounded follow-up, NOT implementation approval, a test run or acceptance.

## Evidence and scope

Inspected only reflection metadata, method-body bytes and resolved metadata
tokens in the already available local PowerShell SDK. No product import/API,
descriptor getter invocation, SDK reserved-member factory invocation, fixture
construction/compilation, full test, standalone diagnostic target, reviewer,
install or security-setting change.

Local System.Management.Automation.dll:
- PowerShell host7.6.5; assembly version7.6.0.500.
- SHA256 2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B.
- Module MVID42bb522b-1082-4dc3-92f1-94f47317bb85.
- Portable metadata evidence:
  docs/diagnostics/isolated-fuzz/p2b-sdk-metadata-investigation/sdk-metadata.json.
- Evidence JSON SHA256:
  D6F18F947481CAE3C18B9870E3C6822114301254FF6382A60802441D21087E8B.

The evidence retains selected method signatures/tokens/IL hashes and bytes,
generic-token resolutions and fixed field declarations, not the SDK binary.
It describes implementation metadata, not a dynamic descriptor-instance test.
A leaf-local .gitattributes preserves the evidence JSON's LF line endings on
fresh checkouts; no historical attributes or frozen files were changed.

Official version-tagged source corroboration:
[PowerShell v7.6.5 MshMemberInfo.cs](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4056-L4144).
The SDK creates five reserved members. The integrating enumerator constructor
calls the generation routine, while generation reuses a pre-existing entry
rather than checking its origin. PSTypeNames is created with a fixed SDK
getter MethodInfo; its value getter is executable and must not be used to
validate provenance. Source lines4649-4658 explain public enumeration's
materialization side effect.

Local bytecode/tokens independently corroborate the relevant factory calls,
owner-field assignments and fixed getter-reference backing fields. The
constructor token0A000AED resolves with a PSMemberInfo generic context;
0A000AEA resolves to GenerateAllReservedMembers. Local enumeration-constructor
IL SHA256 B5E740A3E6488FF2231FFC8C5474BC94A792E924188A26DB44446F1D361331B4.
This supports the earlier operation-history explanation, not a recorded
pre-call branch trace of that past invocation.

Exploratory metadata-only queries initially used the wrong namespace for
CachedReflectionInfo, then omitted a generic context when resolving a token;
a generic disassembly helper also failed. Corrected direct metadata queries
succeeded and only their outputs are in the evidence JSON. Those query errors
were not product/fixture tests, did not execute target methods, and did not
consume/reopen any historical test allowance. No complete decompiler claim.

## Findings

| Descriptor | SDK construction / validation-relevant fixed metadata |
| --- | --- |
| psbase, psadapted | Exact PSInternalMemberSet; private _psObject field identifies its PSObject owner; inherited PSMemberInfo.instance records the owner |
| pstypenames | Exact PSCodeProperty; inherited instance owner; fixed GetterCodeReference and SetterCodeReference backing fields |
| expected getter | SDK Language.CachedReflectionInfo.ReservedNameMembers_PSTypeNames is a static readonly MethodInfo field used by the factory |

The precise backing fields are recorded as declarations, with exact
DeclaringType and FieldType. PSCodeProperty.Value is virtual/non-final and the
class is not sealed. A validation path must first require exact descriptor CLR
type, then use fixed FieldInfo reads, not MemberType, Value, Copy,
TypeNameOfValue or virtual descriptor properties. Reading the stored getter
reference is not invoking that getter.

Important distinction: factory origin is not an unforgeable identity.
GenerateMemberSet and GeneratePSTypeNames return/keep existing same-name
entries. Calling a generator and comparing against what it returns can merely
compare the attacker-supplied entry to itself. That is not an independent
canonical reference. Name/type/flags or a repeated lookup alone are inadequate.

A structural compatibility contract can instead require exact descriptor
class/name/flags, correct owner references, and—for pstypenames—the exact fixed
SDK getter MethodInfo object plus a null setter, without executing the getter.
This validates a narrowly allowed descriptor structure/code target; it does
NOT prove historical factory creation. Same-process code with private-field
reflection/unsafe access can recreate or alter such structures. No claim of
isolation from arbitrary in-process CLR reflection is appropriate. The product
already states its helpers are conventions, not a security boundary.

None of these proposed instance comparisons has been dynamically demonstrated
in this investigation. In particular, the previous raw diagnostic did not save
owner or getter/setter references for its actual three extra descriptors.
SDK metadata supports a candidate mechanism, not its completed correctness.

## Proposed bounded follow-up — NOT approved or started

Intent: legitimate SDK property-enumeration histories must not make a valid
Message fail binding, while input descriptors remain passive. No wire/API/
ownership/sequence policy change; business shapes stay12Message/13Expected.
This adjusts an existing ingress, so a short in-chat bounded design is used,
not a new architectural spec or implementation plan.

New candidate only, preserve every historical module, helper, oracle and packet.
Retain mandatory strict psobject/psextended checks. Permit only optional
psbase/psadapted/pstypenames with exact name/class/flags and fixed owner checks.
For pstypenames also require its stored getter reference to be the fixed SDK
MethodInfo and its setter to be null. Do not call factories to manufacture a
supposed canonical identity, invoke descriptor getters, or skip arbitrary
hidden members. Missing SDK metadata or unsupported structures fail closed.
All application descriptors still require exact PSNoteProperty before values.

Acceptance of this approach explicitly means structural metadata compatibility,
NOT an unforgeable factory-provenance guarantee or same-process security sandbox.
If actual factory provenance is required, this proposal is insufficient;
a different external trust/isolation design must be agreed instead.

Proposed NEW allowance, not borrowed from spent/frozen scopes:
1. One ordinary PS7 host: controlled SDK-history/owner/code-reference helper
   proof, plus compatibility RED against unchanged F76B2DB7. No product patch
   before the exact preconditions and expected failure are observed.
2. After that gate only, patch the new candidate and run focused GREEN:
   ordinary/reused SDK-history cases, malformed/extra/hidden/duplicate business
   entries, wrong owner/getter/setter and masquerading descriptors, plus prior
   passive subclass refusals. Product never invokes metadata getters.
3. One full33-group GREEN against unchanged literal frames/reason expectations.
4. One fresh readonly reviewer only if all three ordered hosts are exact.

Use the unchanged counter-only managed fixture in memory only in those approved
hosts; no install/output DLL or extra compiler host. It measures its own
CounterNote callbacks, not arbitrary PSCodeProperty getter execution; source/
metadata-target checks and review must also cover the latter. Any unexpected
result stops; no repair/retry or replacement reviewer. No acceptance/main/live
grant. Proposed0/3hosts,0/1reviewer,UNAPPROVED/UNSTARTED.

## Preservation and recovery

Read-only reconciliation after investigation: all981 prior mapped pins match;
saved diagnostic stdout/stderr hashes match their receipt. Sealed product
unchanged at F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0.
No old records changed, no new target host or reviewer launched. Historical
diagnostic1/1spent and repair3/3spent/review0/1ineligible remain unchanged,
as do every earlier consumed/frozen allowance and open finding.

Save this metadata evidence/report/current status/recovery on the existing
development branch under standing authorization, not main. Fresh clone restores
portable material, not SDK caches, ignored raw packets, mappings, sessions or
replay/reset rights. No P2-B formal acceptance, PS5/P2-C/P3/full C12/live/
identity/capability/process/Job/clock/queue/release/runtime/security grant.
