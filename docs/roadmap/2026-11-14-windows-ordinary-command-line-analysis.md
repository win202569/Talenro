# Ordinary command-line serialization — read-only analysis and proposal

Date2026-11-14. No application/process launch experiment, build, code/test change,
or identity relaxation. Prior run29a966a471034fb5bb838febb4a2f981 is the only live
evidence: actual183 characters, expected182, exactly one extra terminal U+0020.

## Evidence and confidence

Entry omits ArgumentList and passes FilePath/WorkingDirectory/PassThru. The shared
identity guard uses StringComparison.Ordinal. It correctly rejected the observed
string; the expected serialization, not the comparator, needs design review.

Local Start-Process implementation metadata inspected without invocation:
Microsoft.PowerShell.Commands.StartProcessCommand in Management.dll,
assembly7.6.0.500/file7.6.9.500, SHA256
2181516B3A55866A2263DDAB22A7547FC815240F474199088B8ABA030B92264A.
Methods include StartWithShellExecute and BuildCommandLine. Metadata alone does
not prove which branch ran; no local method was invoked or target relaunched.

Official PowerShell v7.6.0 source defaults UseShellExecute to Windows desktop,
sets Arguments only when ArgumentList is non-null, and selects its Shell path.
This is version-family corroboration, NOT byte-for-byte source proof for the
bundled7.6.9.500 assembly. Microsoft's ShellExecuteW documentation explicitly
describes a trailing whitespace character for executable open/runas command
lines. It does not promise exactly U+0020 for every host. Combining that with
the actual single-space measurement strongly supports Shell serialization as
the cause; no claim of a universal Windows rule or full local call trace.

Sources:
- https://raw.githubusercontent.com/PowerShell/PowerShell/v7.6.0/src/Microsoft.PowerShell.Commands.Management/commands/management/Process.cs (UseShellExecute selection and BuildCommandLine)
- https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shellexecutew (lpFile/lpParameters trailing whitespace)

## Proposed approaches — user choice pending

A (recommended): keep launcher unchanged; replace the sole expected string with
quoted absolute EXE plus exactly one U+0020 for this observed host contract.
The accepted set remains a singleton, not old-or-new. Keep Arguments empty and
ArgumentList absent, all six artifacts/source/process/path/original-handle guards
unchanged. No Trim/TrimEnd, regex whitespace, prefix matching, alternative suffix,
dynamic expectation copied from CIM, or retry. Host differences fail closed.
Record serialization profile and bound launcher version/hash in manifest;
propose a preflight launcher version/hash match to avoid silently extending this
one-host observation to changed runtime. Runtime upgrades require new review.

B: replace launcher with an explicitly configured non-shell process start path,
retaining quoted-only expectation. Larger change to ownership/acquisition and
host behavior; needs separate design, synthetic lifecycle coverage and live
validation. Do not simply add unrelated Start-Process flags to change its branch.

A requires explicit amendment of the former no-trailing-space design, not an
unannounced tolerance adjustment. User has not yet selected/approved either.

## Minimum future verification for A

Independent literal expectation: quote + fixed path + quote + char32, including
path with spaces. Accept only that string; reject no trailing character, two
spaces, tab, CR/LF, NBSP, diagnostic/other arguments, changed path/case and
copied-from-actual expectations. Test policy/context/manifest agreement,
launcher binding mismatch before acquisition, and absent ArgumentList. Preserve
all identity and cleanup regressions, fixed hashes, clean source and zero-target
checks. No product change or candidate rebuild. Scope tests/review separately;
subsequent live attempt still needs new explicit one-run approval.

This turn prepares choices only; no design is marked approved, no entry changed,
no synthetic tests or live checks run. Prior failed outcome remains unaccepted.
