# Isolated integer type-reuse candidate — direct semantics PASS on PS7/PS5

After031e5f24, user approved the recommended NEW candidate-only two-host
allowance:ordinary fresh PS7 then PS5 once each,30000ms per host,stop/freeze
later host on any failure/timeout,no retry/warmup/source edit/original promotion/
reviewer/old-slot reuse. Both2/2 slots completed and are spent. No other tests ran.

## Observed results

| Engine | Started UTC,2026-10-18 | Wall ms | Process CPU ms | Exit | Lines / PASS |
|---|---|---:|---:|---:|---:|
| PS7 7.6.5 Core x64 | 09:11:06.3209711Z | 881 | 750.000 | 0 | 17 / 14 |
| PS5 5.1.26100.9711 Desktop x64 | 09:11:11.6008843Z | 819 | 750.000 | 0 | 17 / 14 |

Both:parent/child admin=False,no timeout,empty stderr,owned-host-exited,
CaptureValid=true,ExpectedResult=true,CandidateOnly=true,
FormalProjectAcceptance=false. All14 groups through P1IntegerComplete and final
LIMIT match the frozen17-line inventory. No saved assertion or fixture error.

The byte-identical direct suite loaded the adjacent isolated candidate. Tested
exact types,bounds/default/positional/owned-metadata behavior,decorations,
zero callbacks,error precedence and clean/decorated transitions passed on both
observed hosts. Per-engine stdout is also byte-identical to the corresponding
unmodified [baseline](2026-10-18-p1-integer-direct-baseline.md). Together with the
prior exact [guard-loss rejection](2026-10-18-p1-integer-guard-loss-negative-control.md),
this supplies tested sensitivity to that missing guard and candidate behavior
for these vectors. It does not prove every safety contract or complete an
optimization/full-project acceptance cycle.

CPU/wall figures describe these two standalone runs only. Do NOT interpret
their difference from earlier baselines as calibrated savings,stable latency,
a matched performance experiment or proof of fewer dynamic type queries.
No4096 complete-flow timing,original full regression,CodeProperty runtime fixture,
independent review or candidate promotion was performed. The original PS5
complete-entry timeout remains unresolved.

## Identity and verification

- Original product unchanged:
  953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
- Candidate unchanged:
  3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3.
- Direct suite unchanged:
  E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98.

Candidate/test/README preparation-time NOT RUN/UNVERIFIED wording was not edited
after measurement. THIS report updates their execution status for the direct
suite only;full regression,performance and promotion remain unverified.

Independently rechecked312 frozen pins,manifest/controller,two engines,two
receipts/four streams,each exact ordered line,17/14 counts,exit0,ordinary context,
under-cap elapsed,empty stderr and owned exit. Per-engine output hashes equal
the prior baseline hashes.312 pins inherit301,add seven negative-control packet
anchors,three candidate files and preparation report. Old records unchanged.

Ignored local packet `.superpowers/sdd/2026-10-18-p1-integer-candidate-semantics/`
retains scope/controller/frozen/index/raw,not committed or clone prerequisites.

| Artifact | SHA256 |
|---|---|
| frozen.json | 254661AFC930C2CD8E74041BA0723C51797D431FF9DFAB7A23EEFF757D6569DF |
| run.ps1 | FCB31A7F8A99BDF4559A6BC24E7A81BB1BB08A050A8AFE0FF7C5F73C6E993DEF |
| invocations.json | 0B5AD4264CBE1E470F378455A7E098FABB31221B3F3054C09837D487A83A0245 |
| PS7 stdout | CA28FBE43D7231D5AE3750C694081EC928BF3A89DF6B098F0EC041C3E0E4AF51 |
| PS5 stdout | 28B4AC8A2476FF1B86919B3469615743D8E7104D09BF6AA752F0C8597AA5037E |
| Both empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

## Gates and next recommendation

All prior spent/frozen budgets persist,including loading1/2 spent/Inherited
FROZEN,baseline2/2 and negative2/2. Review0/1 unused/ineligible. Original product
not replaced;PS5/I1/P1/candidate/main/C12/security/runtime/release blockers remain.

Next recommend a separately approved NEW complete candidate-copy regression:
prepare an isolated folder with the byte-identical3537E9F5 candidate plus six
unchanged original entry/wrapper/test scripts. No product-internal markers,
special4096 reload helper,assertion changes or test weakening. Preserve original
order shape6/reflection3/fixture8/original35/ownership10,83 ordered lines/64 PASS.
Run ordinary fresh PS7 then PS5 once each,30000ms each,stop/freeze later host on
any failure/timeout,no retry/warmup/old-budget reuse. This would test full loading
and boundary behavior on the candidate copy,NOT silently replace the original,
clear its historical verdicts or automatically authorize main/review. Do not
prepare/launch additional full regression within this already spent allowance.

Recover candidate/source,this result,current status and runbook from the existing
development branch. Git restores neither ignored raw/controllers,runtime caches,
credentials nor authority for new tests.
