# Task11 metadata registry — stopped implementation checkpoint

Status: UNACCEPTED / TASK1 INCOMPLETE / NO METADATA PRODUCT API.
Plan: [approved metadata-only plan](../superpowers/plans/2026-10-17-authority-v7-metadata-registry.md).
Baseline:c549a5198523b9a8090fc75d4ab14fba936d7e1d.
The user confirmed Native/inline,maximum5 builds/4 runtimes and one final fresh
readonly review. The first frozen phase stopped;no later phase is authorized to
advance past this safety failure. No automatic retry or safety exemption.

## What actually changed

Only internal/nodecontrol/contracts/schema_registry_test.go was created in the
contracts package. It is the first inventory regression,compiled with the
existing package,and deliberately requires the not-yet-created metadata fixtures.
The registry product file and both metadata fixtures remain absent.
No decoder,domain-framing suite,metadata API or consumer was implemented.
Task1 and Task2 remain incomplete;no independent final review occurred.

Four existing name-only source/test/fixture files remain byte-identical to the
approved plan. Local metadata reconciliation remains146/0 as a design fact only.
No SQL,manifest,canonical approved-set,trust/key policy,authority/diagnostic helper,
SDK/dependency or security setting changed.

## Actual run and result

Run root:.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/metadata-registry-implementation-001.
Control root:the same path with -control suffix. Both were absent at preflight;
the run root and fixture-red identity are now consumed and must not be replayed.
Use the full W path from the plan;all raw evidence/control/cache/binaries stay
ignored on this machine and are not prerequisites for understanding this report.

| Stage | Observation |
| --- | --- |
| Preparation | Existing isolated branch/origin clean;four exact input pins and71-helper aggregate matched. New wrapper parsed in current host and fresh64-bit Windows PowerShell5.1 after a prelaunch generation fix. |
| Pre-audit | Full11536-file copied-SDK/reference and product checks completed before build;the long initial wait was audit,not a Go test hang. |
| Build | One default-vet build,exit0,100.6980032s;schema3,natural shutdown verified,638 total processes. UTC09:57:03.916803–09:58:44.6148062. |
| Runtime | One selected inventory test execution,exit1,1.3960109s. Output shows the expected missing schema-registry.v1.json failure,but the ordinary native gate failed. |
| Safety failure | NaturalShutdownVerified=false,TerminationRequested=true,ShutdownEvidenceComplete=false,ShutdownImageEvidenceComplete=false. Initial job accounting/list queries succeeded;one member remained at root wait. Its OpenProcess failed with Error87;identity/ownership/exit state remained unknown. |
| Cleanup | CleanupVerified=true,ActiveAfterCleanup=0,MembersAfterCleanup=[]. Forced cleanup is not natural completion and cannot establish TDD RED. No timeout was recorded. |
| Finalization | Verified=false,Acceptance=false,Failure=Natural shutdown not verified; forced cleanup is not success;AuditErrors=[]. Product before/after snapshot hashes identical. |
| Budget | Exactly1 build/1 runtime used. Remaining4 build/3 runtime slots are unused,not permission to bypass the stop or replay fixture-red. |

Observed test output,not accepted TDD evidence:

```text
=== RUN   TestAuthorityV7MetadataFixtureInventory
schema_registry_test.go:37: required metadata fixture schema-registry.v1.json:
open ../../../testdata/c12/authority-v7/schema-registry.v1.json:
The system cannot find the file specified.
--- FAIL: TestAuthorityV7MetadataFixtureInventory (0.00s)
FAIL
```

Only that selected test ran:one top-level RUN/FAIL,no PASS/subtest/SKIP observed.
The test body exited as expected,but the full approved RED contract did not pass.
Do not write product code or fixture GREEN on the strength of stdout alone.
The observer's DescendantsRemaining=1 is its shutdown observation,not a claim
that a process remains live after the verified forced cleanup.

## Readonly findings and correction approval boundary

An authoring error was found in the new ignored operational wrapper:its later
RED-result predicate retained old name-directory test/symbol names. The user
explicitly permitted retaining original records,readonly saved-result checking
and correcting subsequent predicates within the original budget,without replay
or relaxed process checks. No saved-only correction was published or accepted:
the actual failure is earlier,in the unmodified ordinary completion gate,not
that latent result predicate. Original verify.ps1/result/STOPPED/receipts remain
unchanged. No continuation/evaluation script or later phase was executed.

The preparatory syntax error was traced to JavaScript replacement-string suffix
expansion while adapting the wrapper;only its generated selector region was
repaired before either parser accepted it and before any native boundary.
A subsequent source read found the latent old-name predicates noted above.
This authoring mistake is recorded,not attributed to the native Error87.

Saved runtime member evidence:job list reported member16616;OpenProcess
QUERY_LIMITED_INFORMATION|SYNCHRONIZE returned87,State=unknown,Owned=false,
ExitCodeRead=false,ImagePath=null. No historical PID lookup or attempt to
reopen that PID was made. These facts locate the rejected boundary;they do not
identify the process or prove the underlying Windows/race/security cause.
No TestMain/init/exec.Command/os.StartProcess match was found in this package
during the scoped readonly source check. That does not identify an OS-created
member or prove equivalence to any older Error87 incident.

Immediate commentary initially called this a build-stage failure;the detailed
saved review corrected it:build succeeded and runtime shutdown failed.

## Portable evidence fingerprints

SHA256 values refer to preserved ignored records,not files to commit.

| Relative artifact | SHA256 |
| --- | --- |
| fixture-red/result.json and STOPPED.json | ABDCD463C187D4C4F10F0C3487717476F082BFB218E1BAA4A6997A4B55DF97FE |
| fixture-red/build/receipt.json | C0AA66A6E59BEE1F3FCB62D73C5182590E1795DCE2E69553F1ECFEEC948BFE48 |
| fixture-red/build/stderr.txt | A737B93EC23F6208B647C518BD21EB84EA953C4D249F82BEDF886FA7D7D72E20 |
| fixture-red/runtime/receipt.json |83F47E354197E9A7CB4FBC5365398FF5904ED3CF0A54386D2184627A1AC057D1 |
| fixture-red/runtime/stdout.txt | C8DAEBE01AAF4681EA3ACE98CF0230BEFBA935BFC453C45A00E535FC7324C36B |
| fixture-red/artifact-binding.json |3D25F3E177FA52E5180863D207351698405B2C3029FF8B4FAA7009406A7E4033 |
| fixture-red/before-product.json and after-product.json |5EA0844391C6EBA4B26E6EF0E71C4531ADEDB4A7B52F308D6CD52C3A11FA9C78 |
| fixture-red/before-fixtures.json |56D84A3E48D2F565C25AAA2FCBA8BCA0CC877F05A927D8080F13A539109F814A |
| control/verify.ps1 |5BABFEDA6293F202C6BF85C4CFB9FD3F9FD8C339C8DA10F846E97460EFC9454F |

## Rulings and resumption

- Ruling:adapt Bash-only skill bookkeeping using PowerShell/apply_patch,not Bash or extra task-done tests;the project explicitly requires PowerShell and a fixed native budget. Cost if wrong:less automated bookkeeping;this report and the plan preserve task state.
- Ruling:the approved contracts-only protocol overrides generic skill baseline downloads,whole-repository testing and automatic review-fix replay. Cost if wrong:broader regressions remain unverified;no broader acceptance is claimed.

No deferred review minors:the final review stage was never reached.
All important native gates remain mandatory. Do not change Error87 into a pass,
drop process ownership/image/exit checks,extend/retry this consumed phase,or
continue on cleanup=true alone. The limited saved-only predicate-correction
approval does not cover any such safety-policy change.

Next decision:approve a separately scoped readonly investigation of this saved
runtime completion failure and the relevant ordinary observer source/tests.
No new native workload,helper repair,revised policy or new run is implied.
Any eventual repair/test/run must receive its own bounded decision. On a new
computer resume from this unaccepted development checkpoint and these portable
facts,not a replay of ignored controls or an assumption that main is current.

## Readonly follow-up — approved after45a14dbf

Scope:read the saved metadata001 records,ordinary observer/guard source and
existing test definitions only. No Go/native/helper tests were executed,no
historical PID reopened,no source/fixture/control/receipt modified. The code
references below are from the unchanged files at45a14dbf.

### Confirmed rejection chain

1. Initial job accounting and member-list calls succeeded,each reporting one
   active/listed non-root member. Buffer capacity16/bytes136,Assigned=Returned=1,
   no evidence truncation. The listed identifier was16616,not0 and within UInt32.
2. process-native.cs:200 calls OpenProcess(0x101000,false,(uint)row.Id).
   Line202 saves GetLastWin32Error immediately,prior to timing/serialization;
   saved ProcessOpen.Succeeded=false/Error87 matches that branch. This is NOT a
   failure of IsProcessInJob,WaitForSingleObject,GetExitCodeProcess or image lookup:
   none of those member operations was reached after the failed open.
3. The catch leaves the observation unknown. JobMembership/CoreObservedUTC/
   FinalShutdown are null;there is no retained,job-verified member handle.
   IsFinalShutdownStateVerified rejects it. Lines424–427 then set incomplete
   shutdown evidence and skip the natural-observation loop altogether.
4. Lines449 onward request forced job cleanup. The final empty group is verified,
   but run.ps1:73 deliberately refuses incomplete/non-natural/forced results.
   Assert-MetaReceipt reaches this unchanged guard before any old-name RED-output
   predicate. Thus fixing the latent wrapper predicate cannot unblock this run.

### What the saved fields do and do not establish

| Saved field | Correct interpretation |
| --- | --- |
| Owned=false,JobMembership=null | Ownership was never established;not proof of a foreign process. |
| ExitCode=0,ExitCodeRead=false | Default unread value;not a successful member exit. |
| ImageError=0,ImagePath=null | Image API was not reached;not a successful image read or an image-loss-only case. |
| NaturalObservationMs=6,TimedOut=false | Fast rejection of unknown state;not exhaustion of the500ms observation cap or120s Go timeout.500ms is a maximum,not permission to wait/requery an unverified PID. |
| CleanupVerified=true,TerminationRequested=true | Forced empty-group cleanup succeeded;natural shutdown remains unverified. |
| TotalProcesses=2 | Job accounting counted two processes over the run;no image,parent or creation trace identifies the additional member. |

Using the recorded monotonic ticks/Frequency=10000000,the list-query end to
OpenProcess start gap is0.2196ms,and the failed open bracket is0.3658ms.
These measurements locate the acquisition window;they do not prove the process
exited during it. Wall-clock stamps are not substituted for monotonic intervals.

Microsoft documents [87 as ERROR_INVALID_PARAMETER](https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-),
distinct from5/ERROR_ACCESS_DENIED. Its [OpenProcess documentation](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess)
describes opening an existing local process and a specific PID0 invalid-parameter
case;this receipt's nonzero identifier does not match that case. The [job member
list structure](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_process_id_list)
provides identifiers,not retained process handles or this member's image/exit
history. These documents do not establish the exact cause of this nonzero-ID87.
Do not prescribe elevation or disabling security software from this error alone.

### Existing tests read,not run

run.tests.ps1:78–88 explicitly rejects forced cleanup,incomplete member evidence,
non-natural shutdown and residual members. shutdown-state.tests.ps1 rejects
unknown ownership/unread exit state/unknown wait state. member-query.tests.ps1
distinguishes a successful open from a failing membership API and rejects foreign
job membership;it retains a live child handle and does not reproduce this failed
open window. These are source-reviewed contracts,not fresh passing test results.

### Diagnostic limit and next decision

Direct rejection cause is established;the underlying Windows/process-lifetime
cause is not. A listed-member disappearance/acquisition race is a hypothesis,
not a verified explanation. No evidence identifies a Go-created child,console
helper,security product,privilege issue or common cause with prior Error87 runs.
With LiveSampling=false and LifetimeDiagnostics=null,this saved ordinary receipt
has no earlier identity/handle observation capable of closing that gap.

Safe next step is to design a minimal acquisition-window evidence proposal before
another workload:not another metadata-test replay. It must specify what earlier
identity/lifetime evidence is missing,how any prospective handle is bound to the
owned job,diagnostic-only versus acceptance use,and a separate bounded approval
for any source changes/tests/run. Existing retained-handle diagnostic modes are
not automatically ordinary acceptance and are not invoked under this follow-up.
No implementation fix or revised safety policy is justified by this readonly
finding alone. Metadata Task1/2 remain stopped;used budget stays1 build/1 runtime.
