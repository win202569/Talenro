# Windows client tool installation result

Completion observed:2026-11-08. Scope: the explicitly approved Visual Studio
Community2026 WinUI tool installation, not client implementation or project runs.

## Approval and preflight

User confirmed Community licensing applies to their usage. User reported1920x1200
at150% scaling, then reported manually changing to125% after discussion. Agent did
not change display/security settings. User explicitly approved installation of
the reviewed14.39GB selection with product/cache on D. Source and dependency
evidence remain in [installation proposal](2026-11-07-windows-client-tool-installation-proposal.md).

Immediately before Install, fresh Installer screenshots confirmed:
- Visual Studio Community2026 October2026 feature update18.11.3.
- WinUI required group retained; IntelliCode, GitHub Copilot, .NET Copilot Skills,
  optional C++/UWP/DirectX/full Windows SDK selections unchecked.
- Product `D:\DeveloperTools\VisualStudio\2026` (2.97GB estimate).
- Cache `D:\DeveloperTools\VisualStudio\Packages` (3.44GB estimate), retain selected.
- Shared path unchanged `C:\Program Files (x86)\Microsoft Visual Studio\Shared`
  (66MB); fixed system tools/cache/SDK estimate7.92GB.
- Total14.39GB, matching approval. Required internal native dependencies retained.

## Observed execution and outcome

Clicked Install once. Download and installation advanced without retry, repair,
hard process termination or changes to selection. Download total adjusted from
about3.44GB to3.47GB during processing;562 packages were reported. Download speeds
varied; Microsoft.Net.CoreUWP spent several observations at52% before progress
continued. Installation reached88%,99% assembly optimization, then100% cleanup.

Final screenshot/accessibility state: installed tab lists Community2026,
18.11.3, all installations up to date, with Modify/Launch/More controls. Progress
bars disappeared. No installation error or reboot request appeared. Generic hidden
accessibility strings (including no-products/news-network placeholders) were not
treated as errors in contradiction to the visible installed product card.

Installer version4.11.24.2075. Launch-after-install was visually unchecked; no
Launch button was clicked. No user credentials/sign-in, automatic reboot, Developer
Mode, security/privacy setting changes, or project restore/build/test/run occurred.

This establishes Installer-reported completion, NOT independent command-level
toolchain verification or native app acceptance. No dotnet/MSBuild process was
started by the agent; catalog versions remain proposed until checked on disk or
with the approved metadata-only commands. No configuration export yet.

## Next bounded step

Propose a metadata-only verification and configuration-export pass: confirm the
installed instance, .NET SDK/MSBuild versions and required component inventory;
export a secret-free portable workload configuration, then update the offline
demo readiness plan. Do not create/build/run a project as part of that pass.
Keep binaries, downloaded packages, installer caches and private logs out of Git.
Another computer must install its own approved dependencies; Git recovery does
not reproduce installed tools. Task11/12 runtime blockers and44/34/6/217,
NOT_ACCEPTED remain unchanged.
