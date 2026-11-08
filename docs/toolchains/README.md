# Windows WinUI tool recovery

`windows-winui.vsconfig` is reconstructed from the installed instance's
`state.json:selectedPackages` Component/Workload entries on2026-11-08, not a
successful native Installer export. It includes25 IDs and no extensions, secrets,
user identities or machine paths. Product entries are intentionally excluded.

Use a separately approved, official Visual Studio Community2026 installer (or a
properly licensed compatible edition) to import and review this configuration.
Importing installs components; do not import automatically. Select product/cache
paths and language for the new machine, review disk estimates, and preserve the
exclusion of optional IntelliCode, GitHub Copilot, .NET Copilot Skills and C++/full
Windows SDK options. Do not use includeRecommended/includeOptional blindly.

This file specifies component IDs, not exact package versions, paths, licensing,
SDK/package restore sources or a complete offline layout. It does not remove
components from an existing installation or reproduce another machine byte-for-byte.
Newer catalogs may change dependencies; review before installation. No import or
fresh-computer recovery test has been performed. Internal bundled packages can
contain Copilot names even with the optional AI components excluded.

Observed reference versions and evidence are in
[toolchain readiness](../roadmap/2026-11-07-windows-client-toolchain-readiness.md).
Project pins, Windows App SDK NuGet selection and build/test allowances remain
separate. Do not store tool binaries, caches or private installation logs here.

Format follows Microsoft's [configuration documentation](https://learn.microsoft.com/en-us/visualstudio/install/import-export-installation-configurations).
