# Fixed type-method lookup comparison — exploratory result,not a fix

## Authority and method

After73acd8ca,user approved the [separate proposed spike](2026-10-17-synthetic-capture-type-lookup-audit.md):
NEW maximum2 hosts PS7 thenPS5,120000ms each,no retry,stop on abnormal/timeout.
Both slots are now consumed. No product/original-test edit,public verifier,
archive/store I/O,original suite replay or use of frozen slots occurred.

The [standalone diagnostic](../diagnostics/isolated-fuzz/synthetic-capture-type-lookup-diagnostic.ps1)
defines two differently named exploratory helpers with the same null handling,
object parameter and safe Object.GetType.Invoke on Value.PSObject.BaseObject.
Baseline resolves the method each time;candidate resolves it once in diagnostic
script scope. No product helper is loaded,redefined or wrapped. No caller input,
type result or validation verdict is cached. This does not design a production
cache lifetime/initialization boundary or certify its safety.

Before timing,both helpers check eight literal fixture cases:null,string,int32,
int64,bool,Object[],Byte[],PSCustomObject with a forged GetType ScriptMethod.
The forged method increments a counter and would return the wrong string type;
observed count remains zero. Each timed group makes36864 calls with five object,
three string and one int64 inputs per nine calls,matching the audited static
type-query mix. The same forged object is reused;this is not4096 distinct record
events or a full verifier. Each result is checked against a literal expected Type.
Loop/modulo,dynamic helper invocation and type-result comparison overhead is
included equally in both group timers. Setup and reporting are outside timers.

PS7 executes baseline then candidate;PS5 executes candidate then baseline. This
is one sample per variant per host,not randomized repeated benchmarking. The
untimed fixture checks exercise both helpers beforehand;no dedicated controlled
warmup exists. Opposite orders on different engines cannot isolate order/JIT/
cache/system-load effects. Do not attribute differences to one of those causes.

## Results

Both children exit0,empty stderr,fixtures row,two ordered group start/end pairs
and terminal diagnostic message. PS7 observed7.6.5Core x64,total2374ms;
PS5 observed5.1.26100.9711Desktop x64,total2165ms. Parent total includes launch
and bookkeeping and is not the sum of timed helper work alone.

|Host/order|Baseline ms|Candidate ms|Candidate minus baseline|
|---|---:|---:|---:|
|PS7 baseline first|1086.954|597.417|-489.537ms (-45.04%)|
|PS5 candidate first|776.307|808.549|+32.242ms (+4.15%)|

Per host,all16 untimed fixture results and73728 timed results matched expected
types;callback counters remained zero. This establishes those fixture outcomes,
not passive-property admission,arbitrary hostile input or public record validity.
The copied helper bodies were statically parsed before launch and unchanged
between both hosts. Raw stdout and saved index rows were compared readonly after
execution;both match. No retry or extra host was launched.

## Interpretation and next work

The target PS5 observation does NOT show an improvement from metadata reuse.
PS7 shows a lower candidate duration,but the opposite PS5 result and one sample
per order/engine do not establish a stable cache benefit. Neither result measures
the full verifier or proves how much method lookup contributes to its two slow
phases. Do not extrapolate group deltas to an end-to-end speedup or claim method
lookup is the main root cause. The static36864 lookup count alone was insufficient.

Recommendation:do NOT implement metadata caching as the timeout repair on this
evidence. Product code remains unchanged. Return to readonly review of per-object
pipeline/name extraction and passive-property enumeration in the two slow loops;
any next controlled probe or product optimization needs its own bounded design
and execution scope. No automatic additional diagnostic or regression is granted.
The original public-suite timeout remains unresolved and must not be hidden by
removing checks,raising caps or relabelling this probe as GREEN.

## Preservation and gates

Probe SHA256 at both launches and after execution:
`05757E44C71C6D848C272E0AB7AC939A200DF06597606FA30E02C3DD05F43C5E`.
New two-entry index:
`AD6D8361078A968DBF8B3431E2E0F0341F2FD0C82B67D02C1D1555066B54CF56`.
Unchanged contract verifier:
`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2`.
Unchanged simulator/safe type helper:
`9DA8A74B984A3111F00B54AA160490DF327BE17A5A07A044B510946EB26A2A13`.
Unchanged archive module/original tests:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C` /
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Prior archive/coarse/internal indexes rechecked unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`,
`DA2CEFA720F55A00423872E54CED6C3369F5506D6F5B6E7AFE706E835246716F`,
`80784F02FE395FE6A57BA4813332224172AAB514708245BB2D83B6716E4BF7F8`.

Raw stdout/stderr,index,runner/progress retained locally under ignored
`.superpowers/sdd/2026-10-17-synthetic-capture-type-lookup`;no receipts deleted
or rewritten to erase failures. Git preserves the portable probe definition and
report,not local sessions/caches or permission to rerun. This exploratory script
is not a supported test/runtime entry or product feature.

Original archive6/12 spent,7..12 frozen,dependent regression and review unstarted;
coarse2/2 spent with PS5 timeout;internal2/2 spent;new lookup2/2 spent. No repair,
acceptance,main merge,real capture/security/vendor clearance,Stage B/I2/I3/full
C12 or expanded authorization follows. Save an unfinished development checkpoint.
