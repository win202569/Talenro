# Verifier property traversal — readonly audit and pending diagnostic

## Findings

After8b31282c,user requested continued readonly review of per-object pipeline
and property enumeration cost. No product/test change or new host occurred.
The previous method-metadata comparison did not establish PS5 benefit and is
not being implemented or repeated.

Assert-ContractEquivalent's object branch collects expected field names with
`@($Expected.PSObject.Properties|ForEach-Object{$_.Name})`,checks actual shape,
then recursively compares the values in that name order. For4096 stored tick
objects,this creates at least4096 pipeline invocations,besides metadata and
nested lifecycle objects. This is a static count,not measured pipeline cost.
The previous9888.525ms PS5 recursive phase includes many other operations;
none of that duration can yet be assigned specifically to this pipeline.

Get-ContractPacket first inspects all event descriptors for NoteProperty,then
reads each value and applies passive-value checks. Assert-ContractShape and
Assert-OfflineCaptureEvent also materialize descriptors and check exact shape,
types/tokens/bounds. These traversals have distinct safety responsibilities:
the first pass must finish before any value may be read;passive-value checks
reject executable members and Length/Rank shadows;shape/event validation
preserves exact fields and supported values. Repeated traversal is not evidence
that checks can be dropped or combined. Snapshotting descriptors once might
need separate mutation/order/ownership analysis;it is not proposed now.

Existing tests contain ordinal marker/field cases,ETS bound-shadow refusals and
executable-property callback guards. They were read,not rerun. Historical pass
counts are not current verification. Production and original tests stay unchanged.

## Next question/probe — proposal only,awaiting user approval

Treat this as an exploratory spike,not a product optimization. Question:does
ordinary foreach extraction of descriptor names improve on the per-object
ForEach-Object pipeline while preserving exact ordered Object[] results?

Propose NEW maximum2 hosts PS7 thenPS5,120000ms each,no retry,stop on unexpected
result/timeout and retain partial output. Two differently named standalone
exploratory helpers,both returning the same explicit array container:baseline
uses the existing pipeline expression;candidate uses array-captured foreach
over the same Properties collection and reads only descriptor.Name. No product
helper is loaded,redefined or wrapped;no descriptor/value/result caching.

Before timing,compare results with literal ordered names for empty,single-field,
multi-field,Unicode and getter-bearing controlled fixtures. A controlled getter
counter must remain zero:neither variant reads property.Value. This only tests
name extraction on fixtures,not admission of executable properties as records.
Keep product passive/type/shape/ordinal/depth/bound/ownership checks unchanged.

Time extraction once for each of4096 distinct literal two-field tick objects
per variant. Both groups use identical invocation and per-result type/count/
ordinal/order checks;fixture creation remains outside timers. PS7 baseline first,
PS5 candidate first;one sample per variant/host,no stable speedup or causal
order/JIT claim. This measures name extraction with harness overhead,not the
whole recursion,packet validation or public verifier. No archive/store APIs,
full verifier,original test replay or metadata-cache change.

If useful,this supports only a later narrowly scoped product design and semantic
verification. If no clear PS5 benefit,do not implement based on invocation count.
No script or new runtime result exists yet;approval is required before probe
construction/execution. Old allowances are not reused or silently expanded.

## Recovery and boundaries

Original archive6/12 spent,7..12 frozen,dependent regression/review unstarted;
coarse timing2/2 spent with PS5 timeout;internal timing2/2 spent;method lookup
comparison2/2 spent. Product contract/simulator/store/original suite hashes
remain those in the previous comparison report. No root-cause/fix/GREEN or
acceptance claim. No main merge,real capture/security/vendor clearance,Stage B/
I2/I3/full C12 or extra permission follows. Save documents-only progress.
