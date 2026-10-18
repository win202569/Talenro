# P2-A test type receiver repair — five exact outcomes andreadonly review complete

## Scope and execution

After8627f71d130d497c7800c8f2382da71fced69398,user approved the short bounded
design andexact five NEW ordinary hosts plus ONE conditional readonly review.
Purpose:correct test traversal/type comparison,not product behavior. Existing
isolated linked worktree/non-submodule,branch codex/c12-b01-task9-coordinator
andGitHub origin reused. Original sources,diagnostic records andfrozen routes
remain untouched;old repair3/6spent,unused3frozen,old review0/1ineligible.

New allowance5/5spent,all exact expected. New review1/1dispatched readonly after
the five-outcome gate;completed without scoped findings below. No extra target,
probe,retry or fix.

| Ordered slot | Child exit | Elapsed ms | stdout lines | Exact expected outcome |
| --- | ---: | ---: | ---: | --- |
| 1 red | 1 | 1193 | 21 | true |
| 2 full | 0 | 1782 | 27 | true |
| 3 phase-mutant | 1 | 814 | 4 | true |
| 4 expiry-mutant | 1 | 819 | 4 | true |
| 5 final | 0 | 1829 | 27 | true |

All five ordinary PS7.6.5 Core x64,admin=False,stderr0,no timeout,
owned-host-exited,complete separate concurrent raw-byte drain,post-pins intact,
ControllerError=null. Max30000ms execution+2000ms owned exit+2000ms drain.
Durable per-slot consumption before each launch;839mapped checks intact after
each run andreadonly reconciliation. Raw byte hashes rechecked.

The controller returns success for EXPECTED mutant/RED child exit1. That means
the negative control failed exactly as intended,not that a failing full suite
was called green.

## Minimal change and what it establishes

New leaf:docs/diagnostics/isolated-fuzz/p2a-type-receiver-repair/.
The ONLY full-suite text difference from the previous repaired-fixture suite is:

```powershell
# Before: TB-Type fixed reflection arguments
$arguments=[object[]]@($Value,$null)
# After: unwrap the reflection receiver, not the asserted type/value
$arguments=[object[]]@($Value.PSObject.BaseObject,$null)
```

No assertion expression,literal input,ordering,descriptor checks,metadata checks,
callback rule or expected result weakened. The old suite was copied/frozen and
executed RED before writing the change. SealRepair enforces exact whole-file
one-expression transformation. Same suite with only this receiver correction
then passes24groups/27lines,including InputMutation,OutputOwnership and
SequentialIsolation. TB-Equal also uses TB-Type,so corrected container recognition
enables its existing recursive field checks rather than relying on a scalar
fallback.

Canonical product andfixture helper remain byte-identical350CAB1A andC720E247.
Combined with the prior current-version stack identifying Plan sent to TB-Leaf,
the RED/GREEN result supports the test receiver handling as the cause of that
traversal failure within this reproducible route. It is not proof about every
ETS/type-data environment or a retroactive complete trace of older runs.

Two individually sealed product-copy mutants were then applied:
phase500 floor(false)->ceiling(true) failed FractionalPhaseRounding exactly;
expiry observed>=deadline -> observed>deadline failed DeadlineEquality exactly.
Both baseline expressions were restored. Final module hash350CAB1A andsame
repaired full test hashB5C4E0CE then passed24groups again. No mutant retained.

## Readonly review

Completed one independent readonly review by p2a_type_receiver_readonly_review.
Critical:none;Important:none;Minor:none identified within the bounded scope.
Reviewer independently checked exact one-expression transformation,retained
assertions/type/descriptor/callback rules,unchanged fixture/restored product,
all839mapped pins,engine/seal/source/raw identities,all five precise ordered
outcomes andfour script parses. No imports/target/controller executions,writes,
Git mutations,new hosts or further reviewers.

Disposition:suitable for the narrow developer checkpoint;NOT global acceptance
or main-merge authorization. New review1/1spent;no replacement review implied.

Executor accepts andretains every declined judgment:

- PS5 compatibility:no authorized/evidenced PS5 execution.
- Other OS/engine/elevated contexts:only ordinary Windows PS7.6.5 Core x64 covered.
- Real clocks/scheduling/hard elapsed guarantees/live runtime:declared integers
  andsaved finite host timings do not prove these.
- Actual process supervision/cancellation/cleanup/publication/integration:
  this product returns plans only.
- Whole-project acceptance/beyond stated assertions andtwo mutants:no wider run.
- Malicious ledger edits/concurrent controllers/descendants/OS failures:controller
  is not a security boundary;five ordinary runs do not validate these scenarios.
- Persistent metadata on re-read strings:unchanged tests permit non-retention,
  while held-view mutation andcross-object isolation remain checked.
- Fresh-machine replay of ignored packets:historical mappings/raw records not
  cloned;portable report andsources establish recovery context,not replay grant.
- Draft result report/Git integration/remote preservation/merge readiness:not
  independently reviewed;parent checks docs andGit separately. No merge grant.

These are scope limitations,not claims that an observed blocking failure was
dismissed. Review found no scoped blocker;no repair/retest after review.

Review is bounded to this repair/controller/evidence anddoes not authorize
main merge or global/live-runtime acceptance. Any blocking finding is reported
without automatic fixes or reruns under this spent allowance.

## Verified versus out of scope

Verified: this local PS7 pure declared-integer suite,strict type/shape/passive
descriptor/decorated scalar/range handling,input andoutput ownership/sequential
isolation,conservative flags,two named rounding/equality mutation controls.

Not established: all-project tests,PS5,CodeProperty/internal-error dynamic paths,
concurrency/global type-data variation,catastrophic recovery,real clocks,
process supervision,stream/storage/cleanup adapters,live runtime,P2-B/C/P3,
C12-wide acceptance or main integration. Three output authority flags remain
false by contract. Prior frozen budgets remain spent/frozen,not unlocked.

Skills used: bounded design approved in chat (no new architectural spec/plan),
TDD old-suite RED before the sole repair,verification against all five exact
receipts,existing worktree detection andconditional ONE independent review.
User's finite scope overrides generic extra global tests/install/retest/fix/
review loops;PowerShell native capture/ledger replaces Bash helper scripts.
No security/trust/ownership/dependency changes;Git-only escalation for metadata.

## Portable ordered output

Transcriptions below preserve outcomes;they are not new executions.

### Red

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
PASS: P2ANullStop
PASS: P2ANormalStop
PASS: P2ANonzeroOrigin
PASS: P2AFractionalFrequency
PASS: P2AFractionalPhaseRounding
PASS: P2ADeadlineEquality
PASS: P2AStopCutoff
PASS: P2AStageWindows
PASS: P2ALateStop
PASS: P2AOriginLimit
PASS: P2AFrequencyLimit
PASS: P2AIntermediateOverflow
PASS: P2ARequestShape
PASS: P2AExactTypes
PASS: P2APassiveDescriptors
PASS: P2ADecoratedScalars
PASS: P2ARangeRelationships
ASSERTION FAILED: P2AInputMutation
```

### Repaired full andrestored final (identical stdout)

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
PASS: P2ANullStop
PASS: P2ANormalStop
PASS: P2ANonzeroOrigin
PASS: P2AFractionalFrequency
PASS: P2AFractionalPhaseRounding
PASS: P2ADeadlineEquality
PASS: P2AStopCutoff
PASS: P2AStageWindows
PASS: P2ALateStop
PASS: P2AOriginLimit
PASS: P2AFrequencyLimit
PASS: P2AIntermediateOverflow
PASS: P2ARequestShape
PASS: P2AExactTypes
PASS: P2APassiveDescriptors
PASS: P2ADecoratedScalars
PASS: P2ARangeRelationships
PASS: P2AInputMutation
PASS: P2AOutputOwnership
PASS: P2ASequentialIsolation
PASS: P2ARefusalShape
PASS: P2AConservativeFlags
PASS: P2AComplete
LIMIT: declared integers only; no real clock or runtime evidence
```

### Phase mutant

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
ASSERTION FAILED: P2AFractionalPhaseRounding
```

### Expiry mutant

```text
CONTEXT: DESKTOP-H4SVCLU\Lenovo admin=False
HOST: 7.6.5 Core x64=True
PASS: P2AApiRequired
ASSERTION FAILED: P2ADeadlineEquality
```

## Source identities

| Portable file | SHA256 |
| --- | --- |
| .gitattributes | E3083818059D2956756D32B1D4A7098E28F57DFB3B971F429A73E51672505B9D |
| capture.ps1 | 718BD87AE191ABFC2845395D5EC375CB9FF88B5207E45A6A5182F984717F5387 |
| expected-output.json | DA3CF1837213F9451A2CB2941DC248B663116CD52EEC199D1D01646A90CA0225 |
| fixture-builder.ps1 | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| full.tests.ps1 | B5C4E0CE14F0B08C647A966884BE6496AA9C05A5D5BEA827B6F86F87EAFCAA65 |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| README.md | E01FB70E7458683AF6583AF96B732C8F088A1EBA025037992CA2603E71BEE4B4 |
| source-pins.json | 8F8EAAA9BF4B9A2FD2FF759144B79637E0D4917F816425CC5B33CC07C54DFB19 |

Local ignored packet:.superpowers/sdd/2026-10-18-p2a-type-receiver-repair/.
Common839=829inherited mapped checks+4prior anchors+6immutable local files.
Full suite andphase product have separate old/new/mutant identities. Historical
Builder-named controller fields refer to full.tests.ps1 here;fixture helper
itself is immutable andpinned among the six common local inputs.

| Receipt | Bytes | SHA256 |
| --- | ---: | --- |
| frozen.json | 367077 | D0725F51180648141ED8858EC747A8AB65BDF08D229D5948A4505D18566F1642 |
| repair-seal.json | 411 | 4009560CB93A267EBB050DEB7A4887516A69B0AB31789B226CB2EB4B8CF3015F |
| invocations.json | 7668 | 9C3C0A4CEA02566512BBA79E0EFB75CDC34FC85B721A829AC78B6539E797C9D3 |
| old-tests.ps1 | 15739 | 4BF2F3C9B4FB18809FD07187CEC598B0DA8B4CB9F42836DA5520F3814D18BBB8 |
| repaired-tests.ps1 | 15759 | B5C4E0CE14F0B08C647A966884BE6496AA9C05A5D5BEA827B6F86F87EAFCAA65 |
| product-baseline.ps1 | 9607 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| red.stdout.bin | 555 | 6501C845BA8E87A3507971494CA64995BB898207F7D0A56E50D54DB3EE805064 |
| full.stdout.bin andfinal.stdout.bin (each) | 735 | 0F165F63428FB1B3825D28A5D9BD9812E125C596153D585D2F16C8AB78ED102B |
| phase-mutant.stdout.bin | 140 | 461D7670C9137F61677A557109C9514A9D81B5AF281E687B60E5C28CCD8A2053 |
| expiry-mutant.stdout.bin | 133 | 74654A6B484E3AFC377DD928FA513C6BBE12AEE64F9B2ABFF33C308D3B1B7E28 |
| each of5stderr.bin | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Mutant identities:phaseFD22144180841E83E62CCDA16A638C9E516DCAEF225E2637231A94F6C900E422;
expiryF0D58CBD1F9A99E3A977C87EA938E674D397FF2A10311D14EF32D34E5EA81662.
EngineSHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.

## Recovery

Save eight portable files,this report,current status/recovery on existing dev
branch under standing preservation authorization;verify local/remote commit
equality after push. No raw ignored packet,cache,binary or secrets committed.
Fresh computers recover source/oracle/portable results andspent allowance state,
not local raw receipts,historical mappings,engine cache or sessions. Use this
new corrected full-suite copy for separately authorized future P2-A work;older
leaves remain immutable historical failures. Do not automatically replay these
five consumed slots,unfreeze older routes,merge main or grant live acceptance.
