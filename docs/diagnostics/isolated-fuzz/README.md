# Isolated fuzz diagnosis

Use PowerShell on Windows. Read the approved2026-10-14 spec/plan and AGENTS.md.
This directory currently contains the verified isolation utility, not a complete
framework reproducer. Never run scripts against a project/cache root as target.

Run `powershell -NoProfile -File docs/diagnostics/isolated-fuzz/isolation.tests.ps1`
from the development worktree. Tests generate unique tiny fixture directories
inside the owned diagnostic root and retain them. Real link tests report UNRUN
if privileges prevent creation; UNRUN is not PASS. No external test library needed.

Dot-source isolation.ps1, then use New-IsolatedToolchain with an existing source
directory and a nonexistent unique RunRoot beneath
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/. The function checks
paths, links and independent source/copy manifests before returning. It does not
run or patch the copy. Do not substitute another toolchain or download dependencies.
The fixed version is Go1.26.5 windows/amd64, independently prepared on a fresh PC.

Generated copies, caches, fixtures and receipts are not recovery artifacts and
must not be committed. Rebuild from tracked scripts/templates; source hashes,
actual executable/GOROOT and runtime results still require verification. Original
matrix3 failure and Pair-unrun remain open; no product acceptance or main merge.
