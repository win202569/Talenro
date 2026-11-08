# Windows accessibility/layout — read-only feasibility review

Baseline: 6be52b54, App/Core last changed at d64de426. This review reads source,
existing requirements and local tool documentation only. No app/Narrator launch,
test, build, install, instrumentation or system setting change. No new acceptance
or runtime allowance. Earlier six keyboard groups are not repeated.

## Findings and evidence limits

| Area | Evidence available now | Missing evidence / feasible next method |
| --- | --- | --- |
| Screen reader | C:/Windows/System32/Narrator.exe exists; XAML has Polite status regions and PowerGlyph AccessibilityView=Raw | File presence does not establish working speech. Current sky API provides screenshots/accessibility trees, not a speech-output recording. Use an agreed human listener for actual announcements; tree inspection is a separate semantic check |
| Exact viewport | Requirements specify 1024x768 minimum and 1440x1024 target logical layout | Screenshot bounds are not measured client content size. No in-app size readout found. Require measured root/client dimensions and scale; ordinary window resizing alone is insufficient evidence |
| Long/RTL | Four synthetic cases in tests/ui-copy-cases.json; email blocks explicitly LeftToRight | No fixture consumer found in App/Core. Text in a JSON file is not rendered coverage; needs an approved test-only loading/display path |
| English | en-US and zh-Hans resources; MainWindow.Copy uses ResourceLoader | Account Language button only opens a placeholder. App.OnLaunched has no language selector. Need app-scoped test language selection or a separately agreed environment; forcing English and proving fallback are different cases |
| Contrast/enlarged text | Theme declares light background and accent/white primary buttons | Source colors do not cover rendered hover/pressed/disabled/focus/system styles. Need agreed criteria and measured rendered pairs. Enlarged text must be tested separately from display scaling |

Source anchors: Talenro.Demo.App/{MainWindow.xaml,MainWindow.xaml.cs,App.xaml.cs,
Resources/Theme.xaml,Views/LoginPage.xaml,Views/ConnectionPage.xaml,
Views/NodesPage.xaml,Views/AccountPage.xaml.cs}; paths relative to clients/windows.
Requirements: docs/superpowers/specs/2026-11-07-windows-client-four-screen-design.md
and docs/superpowers/plans/2026-11-07-windows-client-offline-demo.md.
Tool capability inspected from local computer-use docs/api.md; this does not claim
that all possible external accessibility tools have been inventoried or ruled out.

## Proposed evidence collection, not permission to execute

### A. Readout and semantic names

Before any session, agree an available listener and a finite launch allowance.
Record reader/version, language, test action, actual spoken words and result.
Cover Login labels/read-only demo fields; primary connection name/state; search,
node selection/disabled semantics; account usage/unavailable state; one modal
title/action and focus return. Check whether status changes are announced and
whether decorative power glyph is omitted from normal reader navigation.
Do not infer these outcomes from Raw/Polite declarations or old focus metadata.
If speech cannot be independently observed, record UNVERIFIED, not PASS.
No real credentials or unrelated desktop audio should enter evidence.

### B. Viewports and text

Record actual client/root logical width and height, scale, text scale and monitor
work area before checking four pages, bottom actions and representative dialogs.
Outer window dimensions, screenshots and desktop resolution are not substitutes.
Historical1920x1200 at125% implies1536x960 logical full display before chrome;
it cannot contain a1440x1024 client area. This is historical arithmetic, not a
fresh display measurement. A sufficiently large environment must be available
or separately approved; do not silently lower the target/change system settings.
At each attainable measured size, inspect clipped actions, wrapped status text,
scroll access and dialog actions. Record unavailable targets explicitly.

### C. Synthetic language/layout rendering

Recommended next design decision: a test-only app-scoped fixture/diagnostic entry
that can select long labels, RTL layout with isolated LTR email, and English, and
show actual logical dimensions. This is NOT implemented or approved by this note.
Design must keep default demo behavior unchanged, avoid OS locale/display changes,
use synthetic data only, and separate test presentation from simulated state.
Prefer shared production views so a separate gallery cannot mask actual layout
issues. Specify resource-selection timing and how test mode is excluded/guarded;
add bounded regression/build/runtime checks only after design/allowance approval.
English forced rendering does not by itself prove missing-resource fallback.
The RTL sample is a synthetic stress probe, not an approved translation or proof
of natural Arabic/Persian shaping. All four JSON cases need observed rendering.

### D. Contrast

Agree the applicable contrast criteria before scoring. Inventory actual text/
background and focus/control states, capture measured values with the method and
state, then compare. A source-palette calculation may be a preliminary estimate
only, not full native contrast acceptance. No pass/fail calculation was run here.

## Handoff

Immediate recommendation: design the test-only presentation/measurement entry
before requesting more runtime sessions. Separately arrange a human listener for
screen-reader checks; if unavailable, retain that blocker without repeated launches.
No source change is authorized by this review. Existing main-merge prohibition,
PARTIAL native checklist and Task11/12 plus44/34/6/217 NOT_ACCEPTED remain.

Documentation checks: explicit changed-file review and git diff --check; no claim
of fresh160/160 tests. Necessary notes are preserved on the development branch.
