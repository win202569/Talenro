# Windows offline demo — incomplete setup checkpoint

This directory contains pinned Core/Tests configuration, SessionTests and initial
session reducer branches. User-run tests observed expected RED (12 failed,1 passed);
the subsequent minimal implementation is not yet compiled/tested. No app exists.

Read [execution record](../../docs/roadmap/2026-11-08-windows-client-offline-demo-execution.md)
and [approved plan](../../docs/superpowers/plans/2026-11-08-windows-client-offline-demo-execution-amendment.md)
before running commands. The authorized network retry and locked restore passed;
the test dependency lock is committed. First test invocation compiled both projects
but aborted when VSTest could not connect to testhost within90seconds. Later user-run
comparison reached assertions; GREEN now awaits ordinary-shell validation. Native
launch remains unauthorized. Cloning does not restore tool binaries or caches.

SDK/package caches and outputs are intentionally ignored. No secrets or live data
are required. This is not a production VPN client or an accepted build.
