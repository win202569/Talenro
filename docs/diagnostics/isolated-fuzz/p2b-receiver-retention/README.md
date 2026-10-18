# Existing receiver/descriptor retention — approved one host

After 65bc59e9 the user approved a local test integration, new files only,
one ordinary pinned PS7 host at most60000ms, and one in-process compilation of
the unchanged CounterNote fixture. No product import/API/edit/full suite/
reviewer. All old source/raw/receipts/frozen allowances remain unchanged.

Break caught: capture/relay that replaces an actual wrapper,changes its base/
key/table,loses/replaces attached descriptors,changes hidden/reserved/instance
flags,class or name,or invokes CounterNote callbacks must fail this harness.
This is test infrastructure,not a production validation implementation.

45 cases=nine types(byte,uint16,uint32,uint64,int32,int64,bool,string,byte-array)
x clean/visible-note/hidden-note/visible-subclass/hidden-subclass. Create actual
PSObject in containing array,materialize two SDK views BEFORE attaching exact
descriptors with exact raw Add (no integrating Copy),then snapshot original
already-decorated wrapper. The revised receiver helper refuses an unwrapped
source,uses AsPSObject once on the actual source,keeps wrapper/base/table in a
containing array and passes that handle across helpers without scalar adaptation.
Compare original,postcapture and postrelay actual wrapper/base/key/table and all
raw descriptor references. Require exact tag class/name/flags and same descriptor.
No tag Value/MemberType/Copy/coercion used by observation/retention validation.

Positive observable fixture control touches MemberType/Value/Copy/write once,
then resets counters. All45 cases require counters0 before/after capture/relay.
CounterNote instrumentation does not cover arbitrary getter execution elsewhere.
For EACH case a separately constructed equal-content base/wrapper serves as a
wrong-receiver negative control: retention verdict MUST detect replacement.
IDs are ordinal actual-reference identities;fixed .NET ReferenceEqualityComparer
dictionary lookup uses object[] reflection arguments to preserve PSObject keys,
not overridable equality/hash methods or hash codes as IDs.

Scope is retained wrapper/metadata in the new containing-array helper flow.
No claim that public scalar parameters,ordinary note-property retrieval,
local-storage/nesting histories,all optional SDK metadata or product output
copies preserve the same wrapper;those are not verified by this test alone.

Controller preconsumes one receipt,refuses replay,pins engine+SDK and inherited
1121 mappings;adds old packet4+old report/summary2+six new immutable inputs+
live baseline1=1134 mapped pins. UTF-8 in target. Execution60000ms,owned
exit2000ms,concurrent separate stream drain2000ms;only owned target terminable.
Unexpected result stops,no fix/retry/extra target/reviewer. Derived summary/
report/status/recovery after capture do not change frozen inputs or old verdicts.
Necessary portable checkpoint commit/push to existing devbranch,not main.
Clone excludes raw/runtime/SDK/cache/session/mappings,never resets budgets.
M1 OPEN;no formalP2-B/PS5/P2-C/P3/fullC12/live/identity/capability/security/
runtime acceptance;all authority flags false.
