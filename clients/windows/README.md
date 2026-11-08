# Windows offline demo — incomplete setup checkpoint

This directory contains pinned Core/Tests configuration, SessionTests and minimal
compiling model/event declarations with an identity reducer. No functional demo or
executable app exists. Behavioral assertions have NOT executed successfully.

Read [execution record](../../docs/roadmap/2026-11-08-windows-client-offline-demo-execution.md)
and [approved plan](../../docs/superpowers/plans/2026-11-08-windows-client-offline-demo-execution-amendment.md)
before running commands. The authorized network retry and locked restore passed;
the test dependency lock is committed. First test invocation compiled both projects
but aborted when VSTest could not connect to testhost within90seconds. This is not
behavioral RED; do not implement beyond the stub or retry automatically. Native
launch remains unauthorized. Cloning does not restore tool binaries or caches.

SDK/package caches and outputs are intentionally ignored. No secrets or live data
are required. This is not a production VPN client or an accepted build.
