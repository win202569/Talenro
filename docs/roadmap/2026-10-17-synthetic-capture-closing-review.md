# Independent supplemental/evidence closing review — complete

## Scope and method

After 9952d414, the user approved the recommended NEW one-seat independent
readonly closing review. Reviewer `synthetic_store_closing_review` received
fresh context: supplemental source, approved requirements, relevant unchanged
production functions, portable evidence/status/recovery and saved receipts.
Base: `fe77ddebec235efa0732a775b4efa389d0c6a77a`.
Reviewed head: `9952d414c51a395b7d446b77620f56766472302f`.

No reviewed script was executed/dot-sourced, no tests/diagnostic children,
network, edits, Git writes, fixes, retries or further delegation. Parent
performed readonly source-reference/hash checks and preserves this report.
The one NEW closing-review seat is spent, not an unused second-review budget.
Neither old conditional acceptance review is resumed/completed by this review.

## Strengths and evidence examined

### Independent event fixtures

[Supplemental source](../diagnostics/isolated-fuzz/synthetic-capture-store.supplemental.tests.ps1),
line17 onward, constructs headers/lifecycle/binding input independently of the
production encoder, with test-local hashing. Positive control at line51 checks
both private decoder and public reader, event retention and incomplete lifecycle.
Negative cases at lines56..73 cover kind0/5, EOF channel0/3, negative time and
33 individually legal65536-byte chunks exceeding the aggregate limit. Each
requires the targeted private error before public invalid-payload/false/null
refusal. This prevents later semantic rejection from standing in for guard
execution. Corresponding unchanged product guards are at store126 and147..151.

### Real bounded writer failures

Supplemental line78 onward saves the real new-file helper, injects only local
file collisions and forwards to native CreateNew/write/flush. It fabricates
neither exceptions nor success. The production catch at store223 is exercised.
Checks distinguish primary/secondary errors, compare primary message/FQID/base
HResult, read retained collision/payload/sentinel bytes and require public
refusal. The stopped-directory case makes sentinel publication fail without
replacing the marker error. Scope/reference and normal writer/reader control
at lines112..121 support the limited claim that injection did not escape scope.

### Final evidence boundary

Reviewer checked saved supplemental9/9/6/6 PASS sequences and timings4140/
2602/2175/2871ms, plus dependency81/81/106/106/110/110 sequences and recorded
timings8897/2836/2625/2089/34628/55175ms. Ten child exits0, no recorded timeout.
Raw sequences match indexes; dependency sequences match frozen inventories.
Old-store dependency runs are administrator-context; simulation/contract and
supplemental runs ordinary-context. PS5 stderr retains two Completed progress
records/584 decoded characters per child, classified under prior policy;
PS7 stderr empty. No tests were rerun to obtain this review's conclusion.

Supplemental totals22 content occurrences plus8 sentinels across both engines,
not30 independent requirements. Dependency594 PASS likewise counts occurrences,
not unique requirements or a coverage percentage. Source and index hashes match.

## Findings

- Critical: none confirmed.
- Important defects in new supplemental tests/evidence: none confirmed.
- Minor defects: none confirmed.

Historical proposals in older documents are not additional unused authority.
No defect-free conclusion is asserted beyond the inspected closing-review scope.

## Prior finding disposition

|Earlier finding|Closing-review disposition|Limit|
|---|---|---|
|Important2 malformed-event coverage|Addressed for approved kind/channel/time/aggregate cases on both engines|Not exhaustive decoder or populated-package oracle coverage|
|Important3 writer-error preservation|Addressed for the three approved collision scenarios plus normal control|Not every writer fault stage or interruption|
|Important1 dependency component|Addressed for named unchanged suites in recorded contexts|Overall acceptance finding remains open|

Trigger for overclaim remains treating separate successful packets as unified
acceptance or completion of either old conditional review. Consequence would
be unsupported compatibility/readiness/main claims. Keep unfinished disposition
and obtain separate scope for any remaining execution; do not reuse frozen slots.

## Accepted known gaps and unknowns

- Original PS5 monolithic boundary timeout remains failed.
- Original ordinary-context reader hardlink fixture failure remains failed.
- Administrator PS5 reader receipt remains ExpectedResult=false under its
  original strict empty-stderr rule. Later progress-aware policies do not rewrite it.
- Exact timeout/denial causes remain unestablished.
- Original7..12/partition8..10 stay frozen; both old conditional acceptance
  reviews remain unstarted, with their original prerequisites unmet.
- PS5 missing-root dynamic evidence, file-level reparse, independent fully
  populated/package oracle and maximum duplicated-package coverage remain limited.
- Ignored raw receipts/runners/inventories/fixtures are not recovered by cloning.
- Characterization of unchanged behavior is not a newly observed product RED/GREEN fix.

## Declined to judge

- Whole-branch product correctness: outside this supplemental closing scope.
- Unexecuted environment cases and historical causes: insufficient causal
  evidence; no diagnostic execution authorized.
- Hostile filesystem/admin races and concurrent caller mutation: excluded models.
- Provenance/signing/encryption/coordinated-tampering resistance: not promised
  by consistency hashes.
- Power-loss durability/transactionality/hardware persistence: not promised/tested.
- Live capture/actual execution/OS timing/watchdog/stop/cleanup: outside offline scope.
- Automatic migration/future compatibility: separate version design.
- Vendor/security/Stage B/I2/I3/full C12/production acceptance: outside authority.
Parent retains these exclusions; none is silently dropped from the handoff.

## Assessment and parent disposition

Supplemental source is coherent; saved evidence supports the approved limited
claims. Targeted Important2/3 and dependency portion of Important1 are supported
by adequate bounded evidence. **Ready to merge main: No.** The overall acceptance
finding and listed limitations remain. This review grants no wider clearance.

Parent checked the cited fixture/forwarding/error/control locations readonly;
no disagreement with the limited finding dispositions. No source/test change,
new execution, fix or additional review is needed within this completed request.
Preserve a reviewed development checkpoint, not an accepted release. Future
work should explicitly settle remaining acceptance scope, not repeat already
passing packets or quietly promote historical failures. No further action is
automatically authorized by this recommendation.

## Integrity and recovery

Postreview SHA256 values match pre-review:

- Supplemental:243B4673BF22614718D61FF92BB421F7EE09D993691D0803BB88EC8D11C4B87A.
- Product:508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C.
- Original test:406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF.
- Supplemental index:33D8DFC4A53C94176B4FD0EBF60B970D4FAD3F500859A0692994257A9A6E0AD1.
- Dependency index:8867CB8CB41AA6450468F3754A784B663DBF2DDE1D698C6C0FB12AD31D50CFAD.

Reviewer also matched all12 dependency pins. Raw state and local closing-review
ledger remain ignored/retained. Git preserves this report and current recovery
notes on the existing development branch, not main; cloning grants no new
execution authority. All previously spent/closed/frozen allowances stay so.
