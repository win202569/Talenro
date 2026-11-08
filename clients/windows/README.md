# Windows offline demo — incomplete setup checkpoint

This directory currently contains only pinned Core/Tests project configuration.
There is no implemented demo, executable app or behavioral test suite yet.

Read [execution record](../../docs/roadmap/2026-11-08-windows-client-offline-demo-execution.md)
and [approved plan](../../docs/superpowers/plans/2026-11-08-windows-client-offline-demo-execution-amendment.md)
before running commands. The first restore failed at NuGet DNS resolution; no
test dependency lock is available. Do not assume cloning restores dependencies
or that an empty dependency lock proves success. Next restore needs the bounded
retry decision described in the record. Native launch remains unauthorized.

SDK/package caches and outputs are intentionally ignored. No secrets or live data
are required. This is not a production VPN client or an accepted build.
