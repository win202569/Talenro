# Historical path repair — stopped before implementation

Planed3cead3 approved by user “是”. Preparation checkpoint91cbcf9a was pushed
and verified. RED20 ran exactly once:0PASS/1FAIL/17NotRun/exit1,Expected=true.
Actual System.IO.IOException matched saved invalid-path message and exact
FileReadError,Microsoft.PowerShell.Commands.GetFileHashCommand; stack preserved
Driver-Hash line17 and Driver-CheckHash line18. Setup/Recording errors null.
31sources and2314immutable inputs were unchanged before/after RED20.

Before implementing production code or invoking GREEN21, read-only fixture
inspection found a frozen test authoring defect. history-path.tests.ps1 obtains
the live relative expected-output.json hash from binding.OriginalHistory. That
map deliberately pins operational files through old snapshot paths and does not
contain this live key. The key is present in binding.BaselineSources, with SHA
0FA169E9C7D69EB4D53AFF6C20378ADE4DB43E6767880F2991F26A7B87C69BBF.
Therefore the current GREEN fixture would supply null rather than the intended
valid expected hash. This is an assistant-authored regression-fixture defect,
not a newly observed production failure. GREEN was not run to rediscover it.

Saved history-path.stop.json records this finding without fabricating a runtime
error or a GREEN receipt. No production edit, candidate input/archive, GREEN21,
Prepare, host, snapshot or compilation occurred. Original frozen test/input and
RED20 remain unchanged. No deletion, retry or extra test invocation.

Cumulative20auxiliary/4hosts/3compiles/186snapshots; remaining GREEN21 is
unconsumed but frozen. All ten previous stops remain plus the new local stop.
Runtime remains stopped; product NOT_ACCEPTED.

Proposed next step, not executed: preserve original test/input and RED20; create
a separately approved versioned test/input that takes this existing fixture hash
from BaselineSources, with an explicit preflight presence check. Bind the original
RED20 and amendment as provenance, without replaying RED20. Reopening the one
remaining GREEN allowance requires explicit approval; no added host/freeze scope.
Do not silently edit the frozen test or ignore the new stop.

Preparation archive562files/135directories/31zero-byte, decoded equal, SHA
D4C4F5120F49BF4DD9823347758C40E8B16588D21FE860728D6DB98FD81A8FD8.
Final portable-history-path-final.json preserves original P/Q records and required
H inputs, including actual RED tool call and local stop. Latest status records
its verified inventory/hash. Resume from latest development branch, not main.

Ruling: honor frozen input/test identities and first-unexpected stop over generic
skill fix/retry loops; no extra reviewer/suite/cleanup. This prevents spending
GREEN21 on a known invalid fixture but leaves the production repair unfinished.
