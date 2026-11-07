# Selected phase helper build — stopped 2026-11-07

User continued aftere307084c. Source hashes matched TimingGreen receipt;
remote source checkpoint verified21f74b. Build/Live roots were absent.
One approved restricted build used original command:
go build -trimpath -o <exclusive Build/helper.exe> .

Result0ef0aa/session96775 ->56c195:exit1 after24015.731ms,no timeout,
owned exit confirmed. stdout0bytes,stderr90bytes:
error obtaining VCS status: exit status128; Go suggests -buildvcs=false.
Four root/devtools module hashes unchanged; helper.exe absent; Live root absent.
No retry,release binary execution,real cache hashing,Docker,C12 or security change.

Read-only cause analysisbfffd6: plain git status --porcelain rejects repository
ownership (dubious ownership). Pinned Go cmd/go/internal/vcs/vcs.go gitStatus
calls that command; load/pkg.go requests VCS stamping for a non-test main package.
Default test builds omit this stamp, explaining why synthetic tests could pass
while this build failed. No global safe.directory or ownership changes made.

Next proposed amendment: disable only embedded VCS stamping with -buildvcs=false,
retain explicit source commit/source hashes/binary hash in receipts; use a fresh
exclusive build attempt and preserve failed Build root. One180s build,then at most
one120s live run only after build/source gates pass; no extra test/retry. This is
a proposal,not implemented or run here. No cache/security checks may be weakened.

Archive portable-c12-phase-build-v1.json preserves final sources,original build
claim/result/streams and ledger,not caches or binaries. Earlier v1 and timing-fix
archives untouched. Inline author review only;44/34/6/217,NOT_ACCEPTED.
Original build slot exhausted; new scoped approval required to proceed.
