# RTL and unknown-usage: source-only coverage review

Baseline: 795400763a7313e44a910e1e3947dda3ffe67a94. User approved read-only
source/test inspection, not implementation or execution. No app launch, tests,
build, settings change or acceptance promotion occurred. Only this review and
handoff documentation are changed. Shutdown remains unresolved under the
[consolidated stop point](2026-11-09-windows-shutdown-evidence-convergence.md).

## Source evidence

Paths below are relative to clients/windows/.

| Area | Implementation inspected | Existing assertions and limits |
| --- | --- | --- |
| RTL root and synthetic heading | MainWindow.xaml.cs sets Root direction from IsRtl; LayoutPresentation maps rtl-mixed to LoginHeading; LoginPage reapplies it on Reset | LayoutPresentationTests checks exact bidi fixture and IsRtl for all profiles. LayoutViewContractTests only checks presence of Root.FlowDirection, not its correct conditional assignment |
| Email isolation | LoginPage.xaml and AccountPage.xaml explicitly mark demo@example.com LeftToRight | Both declarations checked by LayoutViewContractTests; actual rendering unverified |
| Technical identifiers | NodesPage.AddNode assigns LeftToRight to non-null node IDs in RTL mode; ConnectionPage assigns it to LockedCountryValue | No direct assertions for these assignments found in the current test suite; Tag-based selection is separately checked |
| Unknown usage | AccountPage.Refresh first selects ordinary usage/device strings; Unknown/Stale choose UsageUnavailable; presentation replaces only UsageValue with the synthetic fixture | LayoutPresentationTests iterates every entitlement/profile for text mapping and unchanged model. ViewContractTests checks initial usage resource declaration, not live refresh |
| Profile activity | IsProfileActive requires Account plus Unknown/Stale; shell observer reads current model state | Existing activity test covers Account/Stale true and Account/Active false, but not Account/Unknown true or Unknown/Stale off Account false |
| Return to known usage | Each AccountPage.Refresh overwrites UsageValue/DevicesValue from current entitlement before optional fixture mapping; shell refreshes account after model notifications | No executable WinUI page-refresh transition test; pure mapping tests cannot prove event wiring or removal of stale on-screen text |

The test project links presentation logic but copies WinUI view source as content;
its declaration/string assertions do not execute native pages. Entitlement reducer
tests cover model state/admission, not displayed UsageValue/DevicesValue.
The text mapper need not itself check Page: the Account view owns UsageValue,
while IsProfileActive controls whether the current page is reported as active.
Off-page view refresh is not evidence of a visible fixture leak.

## Findings and next bounded work

No confirmed functional defect was established by this inspection. Three concrete
coverage gaps are candidates for a test-only follow-up, requiring separate scope:

1. Exercise unknown-usage activity across every DemoPage and EntitlementStatus,
   especially Account/Unknown and off-page Unknown/Stale. Reuse existing pure logic;
   do not launch the application.
2. Strengthen declaration checks for the root RTL/LTR conditional and technical-ID
   LTR guards. Negative controls should remove/change each expected declaration
   in memory and show the check detects it. Label these static contracts, not
   rendering or WinUI behavior tests.
3. Add a sequential Unknown -> Stale -> Active -> Unknown pure-presentation check
   on the same presentation object, preserving input model and ordinary fallback.
   This checks mapping state leakage only, not AccountPage.Refresh event wiring.

Do not refactor production code merely to make these checks possible. A real
failure found while adding tests should be reported and investigated before any
product fix. Test execution needs its own finite allowance and ordinary-token
preflight; no inherited UI launch allowance exists.

## Still requires native evidence

- Root mirroring, actual glyph/bidi order, node/email readability, unclipped actions.
- Unknown/Stale/Active visible usage and device counts through actual scenario controls.
- Profile persistence across logout/reentry and title activity on page changes.
- Normal shutdown, full accessibility and viewport/contrast coverage.

The historical C4 RTL route was stopped without profile validation; C5 usage was
not run. Neither source inspection nor future unit tests can mark those routes
accepted. No speculative service or shutdown changes are justified here.
