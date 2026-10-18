# Preparation ledger — approved ownership marker design

Spec: [bounded proposal](2026-10-18-p1-ownership-group-marker-proposal.md).
Base:3e58ecfcd5278d8a48b4eed76dabc08c26caf031. User confirmed preparation only.

Tasks: (1) independent synthetic RED; (2) isolated sources/inventory/classifier;
(3) offline/static/regression verification; (4) portable status/recovery/checkpoint.
Pre-flight shared interface: new ownership source emits four OWNERSHIP records;
inventory consumes them at one-based96/97/99/100; classifier consumes105 lines;
tests independently insert literal ownership records into the prior101 fixture.
No interface conflict found.

Ruling: execute inline in the already isolated development worktree and keep this
portable ledger,not Bash skill helper scripts — Windows/PowerShell-only project
rule and fresh-machine recovery requirements. Cost if wrong: manual bookkeeping
is less automated;source/test identities and commits remain independently visible.

Ruling: verification is offline/static only;no project target suite,controller,
new reviewer/subagent,install,cleanup of historical packets or frozen-slot reuse —
explicit approved scope and unused/ineligible review gate override generic skill
full-suite/review/cleanup steps. Cost: no runtime/group-output-order proof or fresh
independent review;formal acceptance remains blocked.

Task1 complete: new test first executed against unchanged101-line classifier;
exit1 ASSERTION FAILED: NormalPS7,as expected because it refuses105-line input.
No missing-module/syntax failure. This exercises the admission behavior,not mocks.
Task2 implemented: eight source copy files (seven identical),four direct marker
lines in ownership body edges,distinct105 inventory and offline-only classifier.
Initial GREEN:110 synthetic cases passed;388 prior pins/manifest/index/controller
verified unchanged. Full preparation checks follow;no target execution occurred.

Task2 complete:distinct105 inventory/hash-pinned classifier,body-edge ownership
copy and seven identical companion sources. No target-source import/invocation.
Task3 complete:122 synthetic cases (including103 supported incomplete prefixes),
20 static fidelity cases,and unchanged47/63/76 offline regressions all exit0.
Total328. Existing pinned PowerShell7 engine hash confirmed. Static helper label
parameter initially shadowed the file-loop name,causing nonexistent Parse/file
read;actual source parse0 errors. Scope reproduction proved it;single parameter
rename fixed it,then20/20 passed. No target source or assertion was weakened.

Final scope check:self-review by author under the no-new-review approval boundary;
not an independent reviewer or acceptance. Checked exact source restoration/body
edges,immutable old tools,receipt/order/nesting/ticks,distinct no-entry/open/end/
PASS/error handling and false formal acceptance. No unresolved in-scope finding
or deferred minor. Runtime ordering/performance/owned-host provenance cannot be
proved by synthetic records;kept blocked,not silently judged complete.

Task4 prepared:portable report,hash identities,current status and recovery updated.
Explicit14 artifact+four documentation checkpoint files only;commit/push follows
final verification. No cleanup of old packets,no target/controller execution,
no frozen slot/review0/1 release. Full-project/runtime acceptance remains blocked.

Final fixed:staging revealed root PS1 CRLF checkout policy would invalidate the
physical pins on a fresh computer. check-attr RED=eol crlf;new scoped diagnostic
.gitattributes GREEN=eol lf. No old bytes/root/global settings changed. Git's
actual staged-blob checkout filter with core.autocrlf=true reproduced25 file
SHA256 identities in memory:14 artifacts+8 prior source+3 inventory dependencies.
An initial read-only helper did not launch because Get-Command returned two Git
paths;selecting the first installed path corrected only that helper.
Ruling:scoped LF attributes protect recovery byte identities — normal dependency
preservation,not a target/security change. Cost if wrong:checkout-byte policy
differs on these named diagnostics;unrelated old files remain out of scope.
Task4 verification complete:328 cases plus25 checkout-byte checks;necessary
18-file checkpoint only. Git staging/diff/remote verification completes handoff.
