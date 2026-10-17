# Ordered property name extraction — exploratory result,not a fix

## Authority and method

After0c5cc555,user approved the [separate name-extraction spike](2026-10-17-synthetic-capture-property-audit.md):
NEW maximum2 hosts PS7 thenPS5,120000ms each,no retry,stop on abnormal/timeout.
Both slots are now consumed. No product/original-test edit,product helper import
or replacement,public verifier,archive/store I/O or original suite replay.

The [diagnostic](../diagnostics/isolated-fuzz/synthetic-capture-property-names-diagnostic.ps1)
uses two differently named standalone helpers. Pipeline baseline extracts
descriptor.Name through ForEach-Object;candidate captures ordinary foreach output
into an array. Both use unary-comma return to preserve the same explicit Object[]
container,including zero/one element cases. Neither reads descriptor.Value or
caches descriptors,values or results. Product checks are not bypassed or changed;
this diagnostic does not call product admission code at all.

Before timing,both helpers check five literal fixtures:empty,single-field,
ordered multiple fields,Unicode field name and a controlled ScriptProperty.
Expected names are literal and compared ordinally in exact order;results must
have exact Object[] type,rank1 and expected count. The ScriptProperty getter
increments a counter and throws if evaluated. Counter remains zero. Extracting
its name is NOT permission to admit that object as a capture record.

Each timed group extracts names once for each of4096 distinct literal tick
objects with ordered Kind/ElapsedMs fields. Setup is outside timers;identical
helper invocation and per-result shape/order/text assertion overhead is included
in both group times. PS7 runs pipeline then foreach;PS5 runs foreach then pipeline.
Untimed fixture checks exercise both helpers beforehand,but there is no dedicated
controlled warmup. One sample per variant/host,opposite orders on different
engines:no stable benchmark or causal order/JIT/cache/load attribution.

## Results

Both children exit0,empty stderr,fixture row,two ordered group start/end pairs
and terminal diagnostic message. PS7 observed7.6.5Core x64,total1781ms;
PS5 observed5.1.26100.9711Desktop x64,total1450ms. Parent total includes launch
and bookkeeping and is not just name-extraction work.

|Host/order|Pipeline ms|Foreach ms|Foreach minus pipeline|
|---|---:|---:|---:|
|PS7 pipeline first|650.010|324.852|-325.158ms (-50.02%)|
|PS5 foreach first|392.122|423.076|+30.954ms (+7.89%)|

Per host,all10 untimed fixture results and8192 timed name arrays matched exact
expected shape/order/text;getter counters remained zero. Static parsing completed
before launch;script hash unchanged across both launches and after execution.
Raw stdout rows were compared with saved index rows readonly and match. No retry
or additional host was launched;all original indexes remain unchanged.

## Interpretation and recommendation

PS5 does not show a candidate improvement in this observation. PS7 shows lower
foreach time,but this does not establish a stable benefit across hosts or prove
how much pipeline setup contributes to recursive verification. The diagnostic
includes common assertion/invocation overhead and differs from the recursive
function context. It does not time full records,packet checks or the original
public boundary suite. Do not subtract these timings from prior9888.525ms or
predict end-to-end savings. Opposite orders are not a within-engine order control.

Do NOT implement the pipeline replacement as the timeout repair on this evidence.
No product change or GREEN follows. The metadata-cache probe also did not show
PS5 improvement. Rather than selecting another isolated micro-optimization from
static call counts,recommend returning to in-context full-verifier profiling
design:preserve the call path and identify expensive actual operations in those
loops before proposing repairs. Tool availability,measurement overhead,finite
host scope and all safety guards must be specified separately. This report does
not authorize a profiler install,new host,repeated suite,check removal or cap rise.

## Preservation and gates

Diagnostic launch/post-run SHA256:
`A3DAF77A8FE096413541350A7EA199DBF65E0BBB640E486E0C12CA3AEFB8BCD7`.
New two-entry index:
`3872A1006995E12DEA30FD37E55489F6D7BC3363A524AC652453B941517EDCB1`.
Unchanged contract/simulator:
`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2` /
`9DA8A74B984A3111F00B54AA160490DF327BE17A5A07A044B510946EB26A2A13`.
Unchanged archive module/original tests:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C` /
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Prior archive/coarse/internal/lookup indexes rechecked unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`,
`DA2CEFA720F55A00423872E54CED6C3369F5506D6F5B6E7AFE706E835246716F`,
`80784F02FE395FE6A57BA4813332224172AAB514708245BB2D83B6716E4BF7F8`,
`AD6D8361078A968DBF8B3431E2E0F0341F2FD0C82B67D02C1D1555066B54CF56`.

Raw stdout/stderr,index,runner/progress retained locally under ignored
`.superpowers/sdd/2026-10-17-synthetic-capture-property-names`;all prior receipts
preserved. Git restores probe definition/report,not raw sessions,caches or launch
authority. Exploratory script is not a supported test/runtime entry or feature.

Original archive6/12 spent,7..12 frozen,dependent regression/review unstarted;
coarse2/2 spent with PS5 timeout;internal2/2,lookup2/2,new names2/2 spent. Original
timeout remains unresolved. No repair,acceptance,main merge,real capture/security/
vendor clearance,Stage B/I2/I3/full C12 or additional permission follows.
