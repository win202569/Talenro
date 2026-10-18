# P2-B member inventory — two reserved entries identified, NOT repaired

2026-10-18;basef30aaf1c6ff3f6ffb5947442c4cb00257b4b6ecf.
Branch codex/c12-b01-task9-coordinator. User approved one member-identity
inventory diagnostic,not product repair or continuation of frozen slots.
One NEW ordinary host1/1spent. Product and all historical sources unchanged.

## Actual result

PID20860,ordinary Windows PS7.6.5Corex64,admin=False,1136ms,exit0,
no timeout,stderr0,owned exit/complete streams,pins intact.
Outcome=DIAGNOSTIC_CAPTURED;FormalAcceptance=false. No product import/API
call or descriptor Value/MemberType/Copy getter call during the inventory.
The unchanged in-memory counter fixture's direct observable control ran first,
then all counters were reset before raw access/inventory. Both recorded counter
sets afterward are MemberReads/ValueReads/Copies/Writes/Total=0/0/0/0/0.

Driver uses the frozen p2b-member-set-ingress helper/fixture. It calls only
DG-NewMessage,DG-Install,DG-RawMembers,not the complete stopped target or
product. Names and flags are read via fixed PSMemberInfo SDK fields:
name,<IsReservedMember>k__BackingField,<IsHidden>k__BackingField,
<IsInstance>k__BackingField. CLR identity uses fixed Object.GetType metadata.
No overridable descriptor getter or caller field value is needed to inventory.
Record Stage=Captured,Error=null,MemberValuesRead=false,ProductImported=false,
ProductCalled=false,RuntimeApproved=false. Actual collection contains14members.

| Index | Name | Actual CLR type | Reserved / Hidden / Instance |
| --- | --- | --- | --- |
| 1 | Kind | PSNoteProperty | false / false / true |
| 2 | Sender | PSNoteProperty | false / false / true |
| 3 | Receiver | PSNoteProperty | false / false / true |
| 4 | BindingDigest | PSNoteProperty | false / false / true |
| 5 | SessionId | PSNoteProperty | false / false / true |
| 6 | Sequence | PSNoteProperty | false / false / true |
| 7 | OriginTick | PSNoteProperty | false / false / true |
| 8 | Frequency | PSNoteProperty | false / false / true |
| 9 | DeadlineTick | PSNoteProperty | false / false / true |
| 10 | ObservedTick | PSNoteProperty | false / false / true |
| 11 | Payload | PSNoteProperty | false / false / true |
| 12 | psobject | PSInternalMemberSet | true / true / true |
| 13 | Version | P2BDescriptorFixture.CounterNote | false / false / true |
| 14 | psextended | PSMemberSet | true / true / true |

SDK type namespace is System.Management.Automation. Eleven ordinary fields
plus substituted Version make12application descriptors;psobject/psextended
are the two additional SDK-reserved entries observed in this controlled
construction. This explains count14,not missing/duplicated protocol fields.
It does NOT establish a general whitelist for hidden/reserved descriptors,
prove that metadata flags are unforgeable,or dynamically reproduce the old
product callback. A future rule must not skip arbitrary hidden members or
replace the12/13application-field contract with a14member threshold.

## Identities and raw retention

| Artifact | SHA256 |
| --- | --- |
| Diagnostic target | 6F7DC524014D553BD2585BCD9EE0883798DC8D01F6B802FF5E331351B3D44F25 |
| Diagnostic controller | 838ABF5FD0CE8828E0871660851F6254FCF3A4FEF5B93C151CF113D6CEB41274 |
| Frozen helper,unchanged | E0688991C897EDE6FA7F17D6B47502FC2584BF56E6800120CE0D2EA024F10FC3 |
| Product baseline,unchanged | FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6 |
| frozen.json | D51117F9D28B470A017BFDD39879EBEC4FF006822FD7C8C6B4A51B64441796A9 |
| receipt.json | 09EBD5D567767820C7781D1FE97431A2AE0ACB6E724C212885885BE1EE21135B |
| stdout.bin | 8CC8D8E5587E269B750FBD75EB5375F078A08E4DEA4CE30CA413D635ECBFA2FD |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Ignored packet:.superpowers/sdd/2026-10-18-p2b-member-inventory-diagnostic/.
948mapped pins=939inherited+6previous packet files+3new diagnostic inputs.
All948 rehashed after capture with zero mismatch. Source/oracle/failed index/
raw/trace/review records retained unchanged,no deletions or raw Git inclusion.
Controller preconsumes,refuses replay,ordinary parent/pinned engine,only owned
child,separate concurrent raw streams. Bounds10000ms execution+2000ms owned
exit+2000ms drain. No external compile host,dependency install or security change.

## Proposed bounded follow-up — NOT approved or implemented

Systematic-debugging established member identities before recommending a rule.
Design-clarification treats the follow-up as an existing ingress's bounded
change,no architectural spec/plan document. Short in-chat proposal:

- New candidate only. Recognize only the observed psobject/psextended metadata
  by exact names,exact SDK CLR types,all three observed flags AND same-object
  reference to the current object's corresponding SDK member sets. Do not
  trust names/flags alone or any other hidden/extra member.
- Business input remains exactly12Message/13Expected fields. Reject unsupported,
  duplicate,masquerading or extra members. Require exact PSNoteProperty CLR
  identity before reading application descriptor members or values;subclasses
  rejected with zero callbacks. Public APIs/wire/payload/reason contracts unchanged.
- New allowance proposed3ordinary hosts:updated-helper RED/probe proving raw
  inventory separation and old callback,focused GREEN after conditional product
  patch,full33group GREEN. Use unchanged in-memory counter fixture;include
  targeted rejection of hidden extras/masquerading metadata. Preserve traces.
- One fresh readonly reviewer only if all3exact. Any unexpected result stops;
  no fix/retry loop or old frozen-slot reuse. Approval required before code/run.

This proposal is not proof of correctness or support for other SDK-reserved
metadata histories. Unexpected metadata must fail closed rather than silently
expand the exception list. Product remains baseline and Important passivity
finding unresolved;historical Minor comparator finding still deferred.

Current diagnostic1/1spent;member-set route1/3used+2frozen/review0/1ineligible;
previous descriptor route and all older consumed/frozen scopes unchanged.
Proposed new3+1=0/3,0/1,unapproved/unstarted. No reviewer or product patch.
Necessary diagnostic/report/status/recovery preserved on existing devbranch,
not main. Clone restores portable files/report,not ignored raw/cache/mappings/
sessions or replay/reset rights. No acceptance/PS5/P2-C/P3/C12/live/security/
runtime/real identity/capability/process/Job/clock/queue/release grant.
