# Talenro

Talenro is a Go control-plane foundation. Its specifications live in
[docs/superpowers/specs/](docs/superpowers/specs/).

The module path `talenro.local/platform` is internal-only and is not intended
for public consumption.

## Getting started

For a new computer, follow [repository recovery](docs/runbooks/repository-recovery.md).
Development-tool entrypoints require Windows and **PowerShell Core 7.6.5**.
Windows PowerShell 5.1 and other PowerShell 7 versions are not supported by those
entrypoints. The separate C12/smoke host requirements are unchanged.

```powershell
pwsh.exe -NoProfile -Command '$PSVersionTable.PSVersion.ToString(); $PSVersionTable.PSEdition'
$env:GOTOOLCHAIN='local'
go version
```

Expect `7.6.5`, `Core`, and Go `1.26.5`. Follow the recovery runbook to prepare
both modules separately, then run the offline tool checks. The four development
`.sh` entrypoints reject execution on every platform; Bash smoke is separate.
Full ordinary/C11/C12 acceptance is currently blocked by the recorded security
event. Do not run the full suite or weaken protection to bypass that gate.

See [project status](docs/roadmap/current-status.md) for delivered features and
remaining work. The C12 secure execution controller is part of the test
infrastructure; it does not close the whole C1.2 node/POP implementation.
