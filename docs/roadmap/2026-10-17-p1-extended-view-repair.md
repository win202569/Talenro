# Extended-member view repair — STOPPED BEFORE R3 / UNVERIFIED

User approved the bounded design after3f09208e: use the extended-member view,
inspect descriptor collections without reading their values/calling callbacks,
preserve ordinary controls and original assertions. Seven new <=30000ms fresh
ordinary-token child slots; unexpected results/errors stop without retry. One
read-only reviewer only after seven expected receipts. Prior budgets stay frozen.

## What was changed and verified

Before product edits, ownership suite added FixtureExtendedView. For existing
Int32-zero/Int64-zero/Boolean-true/String-clean vectors, four fresh wrappers each
characterize the public PSMemberSet.Members getter: plain extended view has zero
members; after adding one member named Members, exactly one matching descriptor
is visible. NoteProperty, ScriptProperty, AliasProperty and ScriptMethod covered;
callback count stays zero. The fixed SDK getter avoids dynamic lookup of the
possibly hostile Members name. These are literal bounded fixture expectations,
not a proof covering every scalar/member kind or hostile runspace.

The suite still runs fixture preflight then, in all mode, the unchanged original
35-case suite before ten ownership cases. If the original suite fails, wrapper
now propagates exit1 instead of throwing a second Prior P1 suite failed error.
Original assertions/exit result/printed assertion remain intact; unexpected
stderr emitted by the original script is still captured, not filtered. This
transport change explains empty stderr on these REDs versus the old R3 wrapper
error. No original suite assertion was changed.

| Slot | Fresh host | Result | ms | stderr | ExpectedResult |
| --- | --- | --- | ---: | --- | --- |
| R1 | PS7.6.5 Core x64 | exit1/P1EmptyPrepared | 1738 | empty | true, expected RED |
| R2 | PS5.1.26100.9711 Desktop x64 | exit1/P1EmptyPrepared | 2129 | empty | true, expected RED |

Both ran all mode, passed eight fixture groups including ExtendedView, then
original P1ApiRequired, and failed the original P1EmptyPrepared assertion. Exactly
15 ordered lines each, admin=False, TimedOut=false, owned-host-exited. Other33
original cases and ten ownership cases not reached. No passing product claim.

Only after both expected REDs, product Passive scalar block replaced IsInstance
enumeration with psextended plus the fixed public PSMemberSet.Members getter;
any extended descriptor rejects the scalar. No descriptor.Value is read. All
other product functions unchanged by AST/text comparison. Static parse passed.
Candidate CF4CB4B5 was snapshotted/sealed, but **has not been executed**.

## Why execution stopped

The coordinator erroneously supplied GreenSHA256=PLACEHOLDER in the attempted
R3 controller call instead of the saved seal digest. The controller rejected it
with GREEN seal drift at line50 **before row creation, slot consumption or child
launch**. This is an agent invocation error, not a product test failure or bad
candidate evidence. Correct seal and candidate files were never altered.

The error was not silently corrected/retried. Exactly two invocation rows and
two stdout/stderr pairs remain; R3 streams and mutant.json do not exist. New
budget2/7 consumed, R3–R7 (five slots) frozen, reviewer0/1 ineligible. An explicit
resume may correct only this argument and continue those same five frozen slots
with unchanged suite/controller/seals; it must not replay R1/R2 or revive older
allowances. Without that permission no further run/repair/reviewer is authorized.

## Integrity and recovery

Before snapshots retain productA38C1EE2 and suite86FD5C12. New frozen inventory
inherits78 pins, mapping live source/test pins to these byte-identical snapshots
in the NEW manifest only, plus seven prior diagnosis packet anchors and the
portable empty probe:86 protected historical entries. Prior packets/manifests/
false verdicts unchanged. Final read-only reconciliation verified all86 pins,
two expected receipts and streams, candidate/suite/controller/seal bytes, and
absence of R3/mutant artifacts. No candidate/test edit followed the rejected call.

Controller retains preconsumed slots, prior-row/hash gates, ordinary tokens,
hidden noninteractive no-profile children, separate UTF8 streams, exact inventory/
exit checks and empty-only stderr. PS5 receives only its standard module path in
its child environment, no system setting change. Timeout/cleanup failure paths
were not exercised. No review dispatched. No whole-project/runtime/P2/P3/main
or C12/release acceptance; I1/P1 still unfinished.

Local ignored packet `.superpowers/sdd/2026-10-17-p1-extended-view-repair/` retains
scope, rejected-invocation note, before/frozen/candidate snapshots, controller,
seals/index and streams. No deletion. Git carries necessary source/tests and
portable reports/status/recovery, not raw/runtime state or automatic run authority.
Recover current development branch, not main; candidate CF4C is not an accepted
working module. This report and its exact hashes support a narrowly scoped resume
without a new design/RED sequence if the user explicitly permits that correction.

| Artifact | SHA256 |
| --- | --- |
| Current / candidate product | `CF4CB4B52430C1F7A31EC3DC80EF71704A7B550A4055DB630991EC7674551147` |
| Current / frozen ownership suite | `C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7` |
| frozen.json | `E5F07BFD9258C508D31B8B4B1035BCB9E332D88A0E95D9F1CF86E0D85DBCBC33` |
| green.json (candidate seal, not test success) | `86FA9A37E5AD8AFA303A2B3687658AB13770B9283F0BD4DE7143941BA8DF0098` |
| run-slot.ps1 | `84D0EDE9A3FB0F4780865FAF17BBBF76003AAE51DFCDBD3CBAD7FA3ACA12FEED` |
| invocations.json | `B7A4B9248367D99FEEF3AA7DAFA85ADD3DE2A5AC88CC89482F6E08F34122FF8B` |
| R1.stdout | `358A73E69214403547532A270E15BE55AB4616D7EA94C43E5E6CF169391D9A53` |
| R2.stdout | `41A8D18ED51E57699BEDC278804163DB7011C7C93D606DA3C8273E7B341E85DF` |
| Both stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

Earlier IsInstance repair remains3/7 spent/four frozen/review ineligible; this
task does not consume or unfreeze those four. Other historical allowances and
the old R3=false are retained. Next decision is permission to correct the one
pre-launch argument and resume current R3–R7, not another implementation plan.
