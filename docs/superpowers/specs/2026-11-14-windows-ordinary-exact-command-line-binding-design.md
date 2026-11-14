# Ordinary exact command-line binding amendment

Date2026-11-14. User approved this design; [implementation/test checklist](../plans/2026-11-14-windows-ordinary-exact-command-line-binding.md)
awaits confirmation. Documentation only: no entry/test/shared code changed, no tests,
restore/build or application launch. Live allowance remains0.

## Intent and approaches

Correct the expected serialization for the observed ordinary candidate without
making identity comparison tolerant. The prior run
29a966a471034fb5bb838febb4a2f981 showed exactly one terminal U+0020 after the quoted
executable. Its identity-failed result and ExitVerified=false remain unchanged.

Selected A: retain Start-Process and bind one exact command-line representation
plus the inspected launcher component. Small wrapper/test amendment; preserves
existing ownership and handle checks. Runtime updates deliberately block until
reviewed, rather than silently broadening the observed host contract.

Unselected B: change to an explicitly non-shell launch API and retain the old
quoted-only representation. This would need broader acquisition/lifecycle review
and new live evidence. Not justified for this narrowly observed mismatch.

Trim, alternate accepted strings, whitespace regexes and automatic fallback are
not alternatives: they conflict with the user's strict-comparison requirement.

## Exact contract change

This amendment, ON APPROVAL AND IMPLEMENTATION, supersedes only the old design's
quoted-only/no-terminal-separator expectation and corresponding test assertion.
The original code remains unchanged until separately authorized implementation.

Let P be the fixed policy's absolute executable path, derived from the repository
root and the existing pinned relative candidate path. The sole expected UTF-16
string is U+0022 + P + U+0022 + U+0020. For the prior checkout this is183 characters;
183 is evidence, not a universal hardcoded length for other checkout roots.
The serializer profile name is `shell-quoted-exe-one-u0020-v1`.

Arguments remains the empty string, and the Start-Process call MUST omit
ArgumentList. The final U+0020 is serialization, not an application argument.
Retain the shared guard's StringComparison.Ordinal unchanged. Do not normalize,
trim or case-fold either operand, accept the old form as a second option, derive
the expectation from CIM, or retry after mismatch. Changed path/case, no terminal
space, two spaces, tab, CR/LF, NBSP and any appended argument all fail.

## Launcher binding and acquisition boundary

Proposed fixed host evidence, copied from the read-only inspection:
- PowerShell version7.6.9, edition Core.
- Command: Microsoft.PowerShell.Management\Start-Process; command type Cmdlet.
- Implementing type: Microsoft.PowerShell.Commands.StartProcessCommand.
- Implementing assembly file: Microsoft.PowerShell.Commands.Management.dll,
  assembly version7.6.0.500, file version7.6.9.500.
- SHA256:2181516B3A55866A2263DDAB22A7547FC815240F474199088B8ABA030B92264A.

Future preflight resolves the module-qualified cmdlet, validates the command
type/implementing type and loaded assembly identity, versions, readable on-disk
hash and existing path/ancestor reparse checks. Reject ambiguity, missing values,
read failures or mismatches before writer/process acquisition; never silently
use a function/alias named Start-Process. Capture the validated CommandInfo for
the backend's launch callback, not a later unqualified lookup. Repeat validation
in the existing backend preflight immediately before acquisition. No public
launcher path/hash/version/profile override is added.

Absolute runtime installation paths are discovered from that loaded implementing
assembly, not hardcoded to the WindowsApps package directory. No probing by
launching a process. No switch to an administrator token, shell verb, alternate
runtime or alternative process creation API. Missing old runtime on a new machine
means blocked readiness; do not download/install it automatically.

These checks bind the selected component, not all .NET/Windows DLLs or an entire
host image. The inspected versions/hash do not guarantee every OS serializes
identically: any actual command-line discrepancy still aborts strictly. This is
trusted-workspace validation, not an adversarial concurrent replacement defense.

## Evidence record

Keep SchemaVersion1 and existing fields; additive manifest metadata:
`CommandLineProfile`, `ExpectedCommandLine`, and `Launcher` containing the fixed
PowerShell/edition, module-qualified command, implementing type, assembly/file
versions and SHA256. Populate from successfully verified evidence, not unchecked
machine state. The fixed expected values remain independently defined in policy.
Keep strict Context at exactly RunId,ExpectedPath,ExpectedCommandLine; no added
context keys. Do not rewrite old manifests or label the failed run as passing.

## Unchanged safety and scope

Six candidate hashes and build source5ee8a339b574171fc61f3cf0c0c636bd86421397 remain
as in the original design. Keep ordinary token, branch/clean-tree, source blob/
product diff, artifact/reparse checks, successful zero-target inventory, original
process/handle/native/CIM identity, CreateNew, cleanup/finalization and no fallback.
Default invocation remains inert BEFORE runtime discovery, module import, hash
query, receipts or acquisition. Only RunApproved is public.

No product/shared guard/session/old-entry changes, candidate rebuild, security or
service changes. Four existing deferred Minors remain separately recorded; this
amendment does not silently fix them. Runtime/C4/security/Release unaccepted.
Any later live approval still covers one launch, zero retries,360000ms maximum,
max sleep2000ms, normal close by300s and the original bounded manual route.

## Required synthetic evidence before any new live request

- Positive literal fixture with path spaces: exactly one U+0020; negative fixtures
  for every spelling above, all through the real unchanged strict identity guard.
- Old quoted-only form now rejected, not preserved as an alternate success.
- Policy/context/manifest exact agreement, Arguments empty, ArgumentList absent.
- Launcher command/type/version/hash/read/reparse failure blocks both writer and
  process acquisition; default invocation does not even discover the launcher.
- Prove captured validated command is used despite an unrelated unqualified
  Start-Process function/alias; no alternate command is invoked.
- Original retained-handle, six-file, receipt collision/combined cleanup and
  deadline/no-retry tests remain. Simulated launch only; no live probes.
- Fixed policy tests independently assert literals; failure tests must establish
  the expected rejection rather than pass because an unrelated stub threw.

After approval, prepare the implementation/test plan, then obtain scoped execution
authority. A passing synthetic suite and scoped review permit only a later live
proposal, not a live launch. Preserve necessary source/tests/sanitized records on
the existing development branch; never commit raw receipts or candidate binaries.

## Sources and confidence limits

See [read-only analysis](../../roadmap/2026-11-14-windows-ordinary-command-line-analysis.md)
for local metadata and official PowerShell/Windows references. Official Windows
documentation describes trailing whitespace; exactly U+0020 is established by
this run, not promised for every machine. PowerShell v7.6.0 source corroborates
the mechanism but is not exact source proof for the bundled7.6.9.500 assembly.
