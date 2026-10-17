# In-context public verifier timing — exploratory observations,not a repair

## Authority and preservation of source

User approved the [four-child proposal](2026-10-17-synthetic-capture-context-profile-proposal.md)
after3c2c1c57:PS7 ordinary/instrumented,thenPS5 ordinary/instrumented,120000ms
each,no retry,stop on abnormal/timeout. All4/4 completed;no additional child.
One4096tick record construction and exactly one public verification per child.
No archive/store API,codec roundtrip,original suite replay,system trace,install,
product/original-test edit or unfreezing of original slots7..12.

The [mechanical preparer](../diagnostics/isolated-fuzz/synthetic-capture-context-profile-prepare.ps1)
pins the four pure-chain source hashes,parses originals and inserts entry/finally
instrumentation only inside the seven specified functions. Removing ONLY those
designated additions exactly restores the complete original text,not merely a
rough token comparison. Copies parse and Test-SyntheticCaptureRecord's complete
definition is identical. No source statement is replaced. Copies are written
as UTF8 with BOM for PS5;therefore byte hashes can differ even in classifier,
which has zero function insertions. Saved text is checked against the validated
transform. No existing copy directory may be overwritten.

This generation/static checking ran in the existing orchestration host before
child launches;it did not import/execute product code. Copies were reverified
before every child and after all four;launch/postrun manifests match. Ordinary
children import originals;instrumented children import only the isolated copy
chain,not original definitions followed by replacements. Generated copies and
raw receipts stay ignored;the portable preparer/probe are exploratory artifacts,
not supported runtime/test entries or product features.

The [probe](../diagnostics/isolated-fuzz/synthetic-capture-context-profile-diagnostic.ps1)
uses literal synthetic s/x/o metadata,empty arguments,fixed UTC and4096 ticks at
ElapsedMs0. Profiling activates only for the single public verifier call.
Function-local frame state and a diagnostic stack aggregate calls/timestamps
in memory,without per-call file/console logging or caller-value logging/caching.
Entry/finally checks nesting;stack inconsistencies also fail final diagnostic
guards even if caught by the unchanged public verifier. Original return/throw
paths remain in the copied body. This is intrusive instrumentation,not sampling
or universal proof of semantic equivalence under arbitrary inputs.

## Complete child observations

All children exit0,empty stderr,ordered setup/verify markers,checks row and
terminal message. PS7 observed7.6.5Core x64;PS5 observed5.1.26100.9711Desktop x64.

|Slot/mode|Parent total ms|Setup ms|Public verification ms|
|---|---:|---:|---:|
|1 PS7 ordinary|18850|2611.458|15576.302|
|2 PS7 instrumented|37134|2960.655|33534.851|
|3 PS5 ordinary|33580|4984.260|28022.083|
|4 PS5 instrumented|47871|4884.635|42398.108|

Each public verifier returned Verified=true,Reason=consistent-synthetic-only and
false authority flags. Fixture facts matched4096 events/accepted events,
incomplete lifecycle,null exit/package,false action/authority flags. Every child
returned the same binding:
`C70058134925BDEB45A94027C750FF45CB7C5DD2F5B8C34812F88D5D12A47E94`.
These are positive diagnostic fixture outcomes,NOT the full archive suite.

Both instrumented hosts recorded exactly the same function counts,110811 entries
and110811 exits,balanced empty stack,max measured nesting5. Ordinary hosts
recorded no instrumentation entries/exits. The53353 safe capture-type calls
include other verifier paths beyond the36864 static subset previously audited.

|Function|Calls per instrumented host|PS7 inclusive ms|PS7 subtracted ms|PS5 inclusive ms|PS5 subtracted ms|
|---|---:|---:|---:|---:|---:|
|Get-OfflineCaptureValueType|53353|5864.770|5864.770|6249.774|6249.774|
|Get-OfflineEvidenceValueType|5|0.683|0.683|1.001|1.001|
|Assert-ContractPassiveProperties|20517|2309.392|2309.392|3580.137|3580.137|
|Assert-OfflineEvidencePassiveValue|12304|1014.889|1014.889|1198.484|1198.484|
|Assert-ContractShape|12296|5192.324|4096.070|5796.669|5470.178|
|Get-ContractPacket|1|10817.989|6293.765|14117.016|8630.531|
|Assert-ContractEquivalent|12335|43500.745|8243.048|54486.342|10167.143|

Inclusive recursive totals overlap and exceed verifier wall time:PS5 recursive
total54.486s versus full call42.398s. Do NOT sum inclusive values or treat them
as percentages of the ordinary run. Subtracted values remove only immediately
nested measured intervals. Unmeasured operations and parts of entry/exit
bookkeeping remain;they are not exact CPU/self time of original functions.
Timestamp/frame overhead is particularly relevant for53353 short type queries.

## Interpretation and next recommendation

The public path executes a high volume of guarded helper calls for this large
record. Count consistency and fixture outcomes are useful. However instrumented
durations are materially higher than ordinary observations:PS7 33.535s versus
15.576s;PS5 42.398s versus28.022s. Separate fresh-host samples do not allow a
precise overhead subtraction,stable ratio or attribution of all differences to
one cause. The timing distribution is significantly perturbed and is inconclusive
for choosing an individual micro-optimization or declaring an exact root cause.
It is not evidence to bypass safe type dispatch,passive-property/ordinal checks,
defensive copying or repeated public verification. No product repair proposed
from these aggregates alone;the two micro-probes also did not show PS5 benefit.

Original boundary-group failure is not reproduced or cleared by timing one
public call. Readonly tracing confirms the writer verifies input and then calls
the public reader,which verifies the reconstructed record. Thus the4096-event
archive case does more than this probe,besides preceding cases,codec/hash and
filesystem work. Do not infer an archive duration by multiplying this one sample
or conclude that disk cost is zero. The original timeout could reflect cumulative
work and variable latency;that is a hypothesis,not an established cause.

Recommend a readonly workload/budget inventory of the original boundary group
and its repeated full-verification paths before more instrumented or isolated
micro-tests. Separate correctness requirements from total-host workload;any
optimization or regrouping of verification requires a new explicit design and
finite execution scope,not dropping tests,raising the cap or reusing frozen
allowances. No new profiler/diagnostic host is authorized by this report.

## Hashes,recovery and acceptance boundaries

Preparer SHA256:
`5D271ED86B7F41D60D89C812A22E41EA615C701DADB65E2D9442ABBFAAFCDCE3`.
Probe SHA256 at all launches/postrun:
`98A7623919F95AF3FE9A8A523DABB561F5A7A14E5C33E1ACBDAAFDA3A3841E6B`.
New four-entry index:
`07C14EA6AF3A562881DDB54B8322CACCE2DE56C7CA0951755C2E46740E501DD9`.
The four original dependency hashes are pinned in the preparer;all unchanged.
Archive/original tests remain:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C` /
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Prior archive/coarse/internal/lookup/names indexes rechecked unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`,
`DA2CEFA720F55A00423872E54CED6C3369F5506D6F5B6E7AFE706E835246716F`,
`80784F02FE395FE6A57BA4813332224172AAB514708245BB2D83B6716E4BF7F8`,
`AD6D8361078A968DBF8B3431E2E0F0341F2FD0C82B67D02C1D1555066B54CF56`,
`3872A1006995E12DEA30FD37E55489F6D7BC3363A524AC652453B941517EDCB1`.

Raw/index rows compared matching readonly. All generated copies,manifests,
stdout/stderr,index/runner/progress retained locally under ignored
`.superpowers/sdd/2026-10-17-synthetic-capture-context-profile`. No prior record
deleted/rewritten to erase failure. Git restores portable source/recipe/report,
not ignored raw sessions,caches or execution authority. Pinned-byte mismatches
must stop preparation,not be silently accepted on a new computer.

Original archive6/12 spent,7..12 frozen,dependent regression/review unstarted;
coarse2/2 spent with PS5 timeout;internal2/2,lookup2/2,names2/2,context4/4 spent.
No timeout resolution,GREEN,acceptance,main merge,real capture/security/vendor
clearance,Stage B/I2/I3/full C12 or expanded permission. Save unfinished progress.
