# Record-isolation harness — scoped RED→GREEN complete

Status: SCOPED_HARNESS_VERIFIED_NOT_RUNTIME_ACCEPTED. Execution base
`e75af9d17ab57771f752a4d61f9ffa7717163710`; user approved the
[narrow Native plan](../superpowers/plans/2026-10-31-offline-record-isolation-repair.md),
SHA `790571C1F49DB0ACD903605AD1F9B4CAC5F31A59B02C232B62925324B35E3826`.
No production/public-entry/validator source or historical test was changed.

## Actual execution

Created the eight planned sources in the independent p2b-record-isolation-repair
leaf. Assertions and independent literal expectations were written before the
seed implementation. Preliminary static AST/LF/JSON checks passed; all2365
inherited dependency entries matched. No baseline/suite/module probe ran.

| Phase | Actual evidence |
| --- | --- |
| Prepare seed | Once,2394Present pins/8snapshots/ChangedInputs[]; exit0 |
| isolation-red | Once,PID14852,5FAIL/4PASS exactly, capture exit1; parent exit0/Exact=true |
| Prepare isolation-green | Once,2394Present/8more snapshots; exactly case-module.ps1 changed; exit0 |
| isolation-green | Once,PID1496,9PASS exactly, capture exit0; parent exit0/Exact=true |

Seed manifest SHA:
`013E6F5BBCEF61850363EA2B8599BBE05E5630B1D32D85AA5B9A15B20127B8AE`.
GREEN manifest SHA:
`93B1D18EE6C4FF944FEB45AA8CE8E91C7011BDD3C0374FFA9A6587C2AAF1D447`.

The sole between-phase source edit adds `Export-ModuleMember` without arguments
inside the new case-module construction. Other7new sources and all inherited
inputs remained byte-identical. All16source snapshots are preserved.

RED failed precisely OuterFactorySame/OuterCaptureWriterSame/
OuterIntentWriterSame/OuterCollectorSame/FactoryReplacementPrivate. The other
four groups passed. GREEN passed all nine. Actual module exports dropped from
four functions to zero members; all four outer ScriptBlock references remained
original during case creation and after inner factory replacement. Inner state
and sentinel remained separate from outer state. No global restoration masked
an import leak, and no historical driver/capture/runtime module was executed.

Actual existing-file CreateNew errors retained the same in-process ErrorRecord
reference in the module collector. Independent healthy publication produced the
exact declared JSON bytes once, with open/write/flush/dispose1. ExistingFailure
publication attempted open1/write0/flush0/dispose0; its second logical call was
rejected before IO, with actual IOdelta0 and the same first error retained.
Parent independently confirmed both existing files still contained ASCII PRIOR.

Both hosts had empty stderr, exact owned exit, completed shared double-stream
drain, no timeout/termination/rescue/parent error and intact2394post-pins.
Both final evidence records were COMPLETE, all9groups present, no host primary/
secondary error or overflow, NotRun=[], all five capability flagsfalse. Parent
checked exact schemas, values, types, command/module identities, real error
signatures and frozen source/stack call sites, actual files, complete journal
and raw output/exit. PASS text alone was not the gate. Full original tool
objects were saved before interpretation; no early session/restart occurred.

## Budget, limitations and rulings

Consumed exactly the NEW2capture hosts/0targets/0compilations/0reviews;
allowance exhausted. Original repair6hosts/7targets/1compile/conditional1review
stays FROZEN; scope-diagnostic1host remains exhausted. No third host, retry,
extra Prepare, independent review, install, deletion or main merge occurred.

Ruling1: the explicitly approved finite native sequence replaced skill-default
Bash/setup/baseline/suite/task-done reruns/agents/review/deletion. Cost: only this
named harness regression is verified; the new publisher received instrumentation
checks, not its own missing-function RED cycle or independent review. This is
not full-repository/security/production/full-C12 acceptance.

Ruling2: before any Prepare, allow precisely ledger and the saved preliminary
static-tool record in the new seed directory guard. The first guard expected
only ledger despite preserving that record. No command had failed; no source
was frozen. Cost: final driver hash differs from the preliminary static record;
that original record is retained, and actual Prepare re-parsed/hashed/snapshotted
the final bytes. No frozen source or runtime allowance was changed by this ruling.

The publisher's healthy IO and actual open failure do not dynamically prove
compound write/flush/dispose failures, host death or all exits failing. The
new isolation seam has not been integrated into frozen v2 or a public profile.
The original writer-red's first internal ErrorRecord is still UNKNOWN; neither
this result nor its errors replace the original empty evidence. I1/I2 remain
unresolved, C1/C2 open and R2/Minor2 deferred; five flagsfalse.

## Preservation and next step

Sources, both original result chains, intent/final/raw/evidence/journals,
approved plan/spec snapshots, all16source snapshots and complete tool returns
are preserved by the new leaf's portable-final.json. Original raw stays ignored
on this machine; portable base64 retains exact bytes including zero-byte files.
Archive/source byte comparisons are author checks, not another runtime run or
an independent review. Actual export:47files/6directories/16source snapshots;
all decoded bytes were compared against the originals, including both empty
stderr files. LF portable SHA:
`83FE6E002281837B4E841F334A591B6D936EE2D90D107441529AF4D123DFFD2D`.
The archived ledger captures the pre-export/pre-push handoff point; this report
and latest current-status provide the subsequent preservation disposition.
Clone the latest existing development branch and read
current status/recovery; archives restore progress, not identity/cache/Session/
Ack or execution allowance. Do not replay the now-exhausted driver.

Next requires a separately written/approved actual I1/I2 repair and integration
revision with an exact source list and new finite verification budget. It may
use this tested recording boundary in a new approved integration, not silently
modify or resume the frozen old runtime/test namespace. This checkpoint grants
no extra implementation, tests, review, acceptance or main-merge permission.
