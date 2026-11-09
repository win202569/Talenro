# Close-stage implementation — build-only checkpoint

User approved the plan and native sequential execution plus one final review.
Range starts 41ce1d92; Task1 f96d8eed, Task2 f16c03a5, Task3 9ba5488c.
No WinUI launch, forced exit, service/security change or main merge.

## Evidence

- Baseline 226 tests; restricted testhost could not connect in 90 seconds.
  Same timeout under ordinary external execution passed; no product-cause claim.
- Task1 RED 7 then GREEN 7, full 233. Task2 positive RED 4 and rejection RED 14,
  GREEN 18, full 251. Task3 activation RED 11, GREEN 20, full 271; four additional
  pure flow examples bring current full suite to 275/275 (0 skipped).
- Build gates RED missing target then GREEN. Verification-path regression RED
  before App-only path property, then GREEN.
- Trace, ordinary verification copy and layout verification copy each build with
  0 warnings/errors. PE metadata verifier passes: recorder types absent from
  ordinary/layout and present in trace. DLLs inspected, never executed.
- Original demo and two prototype EXE/DLL SHA256 values unchanged after builds:
  demo EXE 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95;
  DLL 7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB;
  TextBox EXE D19AFA9B50F68A8340CF25F3F8FB967717C68BA426061BEB1A7286C864793612;
  DLL D31308E544BC72C3671195C78A6A7D072B776E27768C9D1CF7DD8304D5C5D6A7;
  list EXE 32E1E7392B46D89DC297DA03C097A1661894B61A919DE697BB087738A76C16B5;
  DLL 9A390F3F19ED94F4E33A0E1905A29A705655A1E324E85F9F933D601525E86BD1.

## Decisions / failures preserved

1. Ordinary external test execution, no timeout change: restricted host failure;
   cost is environment difference, not a diagnosis of the application.
2. PowerShell/apply_patch bookkeeping instead of Bash-only helpers: repository
   constraint; cost is manual ledger maintenance.
3. Default-off exclusion moved from Task4 to Task1: protect partial commits;
   cost is changed sequencing.
4. Startup-only bounded readiness wait and internal worker state: deterministic
   startup/tests; cost is additional internal API, never a close-time wait.
5. MainWindow explicitly passes buffer: one extra source file; cost is conditional
   wiring, verified by trace build but not live execution.
6. Initial relative global verification paths avoided shared absolute collisions,
   but exposed Core's stale obj sources (16 duplicate-attribute errors).
7. Attempted escaped DefaultItemExcludes list failed; single pattern fixed Core
   source selection but App failed with 52 missing generated-symbol errors.
8. Superseded 6/7: App-only CloseStageVerificationMode, absolute NuGet extensions
   path. Evaluation proved relative global path left NuGetPackageRoot empty and
   ApplicationDefinition absent. Cost: another opt-in build property; no Core
   source change, old-output deletion or change to preserved artifact paths.

## Independent review and bounded fix pass

Fresh read-only review of 41ce1d92..82ec5c49 identified one Important issue:
unknown/mismatched verification mode silently selected canonical output paths.
No Critical or Minor findings. Four negative cases reproduced RED (0/4 rejected),
then the initial validation target was extended with CLOSESTAGE002; GREEN4/4 and
the whole gate script passed. Full .NET suite again275/275. No second reviewer.
Post-fix ordinary/layout/trace builds passed0warnings0errors in14.51/13.14/13.37s;
metadata isolation passed again and all six original hashes matched again.

Review boundary rulings (accepted, not acceptance waivers):

9. Live callbacks/native removal/shutdown cause remain unjudged: no launch
   authority. Cost: build success cannot establish shutdown correctness.
10. Real overhead, storage latency/tail-loss and retained-process evidence remain
    unmeasured. Cost: live recording may perturb timing or lose useful evidence.
11. Hostile concurrent replacement and arbitrary deployment aliases are outside
    the trusted-workspace model. Cost: path checks are not an adversarial sandbox.
12. Deliberate global MSBuild output/intermediate overrides are outside supported
    commands; documented rather than prohibited. Cost: callers can defeat isolation.
13. Reviewer accepted validation ledger, did not rerun tests/builds/hashes under
    its read-only mandate. Cost: review is not independent runtime reproduction;
    executor's post-fix verification is the evidence.

No deferred minor findings. Necessary source/tests/docs remain on the existing
development branch; no PR, main merge, branch or worktree deletion authorized.

## Remaining gate

Original shutdown issue remains unresolved;
runtime stage callbacks, overhead and retained-handle process observations not
verified by these tests. All broader acceptance blockers remain. Recover from the
development branch, read diagnostics/close-stage-recording.md for exact commands;
do not assume main, caches, original binaries or conversations are restored.
