# Talenro

Talenro is a Go control-plane foundation. Its specifications live in
[docs/superpowers/specs/](docs/superpowers/specs/).

The module path `talenro.local/platform` is internal-only and is not intended
for public consumption.

## Getting started

For a new computer, follow [repository recovery](docs/runbooks/repository-recovery.md).
The current full repository test suite includes Windows-native C12 tests.

```powershell
go version
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
$env:GOFLAGS=''; $env:GOWORK='off'; $env:GOENV='off'
go mod download
go test ./... -count=1 -p=1 -timeout=60m
```

See [project status](docs/roadmap/current-status.md) for delivered features and
remaining work. The C12 secure execution controller is part of the test
infrastructure; it does not close the whole C1.2 node/POP implementation.
