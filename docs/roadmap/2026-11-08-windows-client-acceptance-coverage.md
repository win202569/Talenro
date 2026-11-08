# Windows demo acceptance coverage — evidence reconciliation only

Latest addendum: after f349823c approval, PID8700 observed K1–K6 keyboard gaps
and closed normally; see latest native-verification record. Priority1 below is
completed only for those explicit paths. No full keyboard/accessibility acceptance;
expanded dropdown and disabled-state permutations remain outside those checks.
Next safe step is read-only preparation of priorities2–4 methods/constraints,
not a repeated core-flow session. No launch allowance remains.

Baseline: d64de426. User approved splitting compound checklist rows and mapping
existing evidence. This update changes documentation only: no app launch, tests,
build, dependency restore, instrumentation or settings change. No acceptance gate
is waived; launch/test budgets remain consumed. No main merge.

## Confirmation: separate visible outcome from internal atomicity

| Subcheck | Existing evidence | Limit |
| --- | --- | --- |
| Before Confirm, old preference/lock remain; Cancel preserves them | Native PID17108, gap-only session item4; NodeSelectionTests ConnectedApplyRequiresConfirmationBeforeAnyCommit and CancelConfirmationPreservesConnectionPreferenceAndCountryLock | Selected scenarios, not every event interleaving |
| Confirm visibly starts Connecting, DE lock, de-1 subsequently selected | PID17108 item4 in native-verification record | UI snapshots establish outcomes, not every intermediate state |
| Reducer returns changed preference and lock together | NodeSelectionTests.CrossCountryConfirmationNamesBothCountriesBeforeAtomicCommit; included in prior full160/160 run recorded by focus-fix | Model-level assertion; no new test run here |
| UI handler submits Confirm/Cancel and publishes returned state | Read-only NodesPage.Apply_Click and DemoViewModel.Dispatch: handler dispatches chosen event; Dispatch assigns reducer result before PropertyChanged | Source trace, not measured live event count or concurrency proof |
| No intermediate partial state in native runtime | No independent runtime trace in existing records | Remains unverified; repeated clicking/screenshots alone would not establish it |

## Reset: preserve demonstrated sequences without overclaiming

| Subcheck | Existing evidence | Limit |
| --- | --- | --- |
| Logout clears model state | SessionTests.LogoutClearsAllSessionState compares a dirty model to new DemoState; prior full160/160 | Fixture covers populated fields shown in the test, not arbitrary UI/persistent storage |
| Reentry disconnected, lock off, active account, Auto, blank notice | Earlier native session item6 and PID17108 item7 | Observed values, not exhaustive hidden-state inspection |
| Changed Nodes search/list reset | PID17108 item7: JP + Loading → logout → empty search/normal/Auto | Explicitly exercised sequence; already checked separately |
| Login selector returns normal | PID10372 Login reset sequence: invalid → normal → enter → logout | Does not claim entry from invalid scenario; it was restored to normal first |
| Close/reopen starts fresh after saved node/lock | PID10208→PID14520: saved jp-1/JP, normal close, reopen Login, reentry disconnected/lock off/Auto/empty search | Historical source baseline0d8a436f; not a fresh full reset rerun after PageHost-only change |
| Every internal/UI field and storage behavior resets | No exhaustive runtime evidence | Remains unverified; do not equate the selected checks with persistence/security acceptance |

Evidence sources: [native sessions](2026-11-08-windows-client-native-verification.md),
[latest test/build and focus verification](2026-11-08-windows-client-focus-fix.md),
and clients/windows/Talenro.Demo.Tests/{NodeSelectionTests,SessionTests}.cs.
Latest code change affects PageHost tab navigation only; historical evidence is
retained with its baseline rather than silently described as a current-build rerun.

## Remaining verification priorities — proposal, not execution authorization

1. Prepare an explicit missing-path keyboard matrix from existing recordings:
   remaining reverse traversal and modal variants only. Do not repeat the fixed
   Settings↔Connect/Search/MFA checks or core connect/reset flows without a reason.
2. Separately agree screen-reader testing and a reliable observation method;
   stale focus metadata is not an accessibility verdict.
3. Establish feasible logical viewport dimensions and text/contrast checks.
   The recorded larger target does not fit the previously reported desktop;
   that is historical information, not a fresh measurement. Do not alter settings.
4. Long/RTL samples and English fallback need an agreed way to display them;
   presence of ui-copy-cases.json is not proof they have been rendered.
5. Internal atomicity and exhaustive storage/reset evidence need a scoped method,
   not another general UI session. Keep those boundaries explicitly unresolved.

The original missing-path matrix/session proposal above is now historical and
superseded by the latest addendum. This record itself grants no new runs.
Overall native checklist remains PARTIAL; Task11/12 and44/34/6/217 NOT_ACCEPTED.
