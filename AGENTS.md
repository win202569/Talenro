# Project instructions

## Preserve progress for a fresh computer

The user authorizes committing necessary project files and pushing ordinary
checkpoints to the existing development branch after completed work or handoff.
Do not repeatedly ask for commit/push permission for that routine preservation.
This standing instruction does not authorize main merges, force-pushes, branch
deletion, history rewrites, installs, security-setting changes or wider acceptance.

- Include necessary source/tests, approved specs/plans, progress and recovery
  documentation. Record remaining tasks, blockers and verified versus unverified
  results so another computer can resume from the repository alone.
- Review an explicit file list before staging; do not blindly stage unrelated
  edits. Never commit secrets, credentials, private/signing keys, caches, tool
  binaries, Docker data, temporary files or ignored session/raw receipts.
- Verify in the approved scope before claiming completion. A failed verification
  may be saved only as a clearly labelled unfinished development checkpoint,
  without claiming acceptance or merging main.
- Push to the verified GitHub origin and the existing development branch, then
  confirm the remote branch SHA equals the local commit. If a push is blocked,
  report that only the local commit exists; never claim remote recovery.
- Update docs/roadmap/current-status.md and the relevant execution/recovery
  documents at handoff. Ordinary source/tests and portable fixture data must not
  depend on ignored session directories or a previous machine's private cache.
- Fresh-computer recovery starts from the latest development-branch commit, not
  automatically main. Read docs/runbooks/repository-recovery.md and current status.
  Dependencies/credentials and real environment acceptance need their own
  preparation; cloning does not restore caches, conversations or live data.

## Execution environment

Use PowerShell on Windows. Preserve existing test/security gates and scoped
acceptance boundaries; do not install tools or change security settings merely
to make a validation pass.
