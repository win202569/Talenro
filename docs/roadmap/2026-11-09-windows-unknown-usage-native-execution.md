# Unknown-usage native observation

User approved the one-launch plan at cc565557. Source baseline cc565557; budget
0/1 before launch. Ordinary token, empty target process/window inventory, exact
preserved EXE/DLL hashes and unchanged observer blobs confirmed before launch.
Plan: docs/superpowers/plans/2026-11-09-windows-unknown-usage-native-check.md.

Ruling: verification-only work uses this portable execution ledger; no source
implementation, TDD runs, scratch implementation workspace or code-review cycle.
Task1 identities feed Task2 and its receipts feed Task3. No interface conflict.
No extra launch/retry is permitted.

## Launch aborted at identity check — budget1/1 consumed

Run9635909714294fbdb4f31cb305ae84fb, PID23364, observer18120.
Original Process.StartTime UTC2026-11-09T18:39:47.4427247Z. Arguments exactly
--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage.
The original object and SafeHandle were obtained, but the command threw before
the first paired observation. No functional navigation/input was attempted.

Read-only CIM recheck returned the exact expected executable path and arguments,
PID23364 and CreationDate2026-11-09T18:39:47.4427240Z. The session command directly
compared formatted creation-time strings: these differ by7 FILETIME ticks (0.7us).
Thus the exact-string condition failed despite matching path/arguments and the
same microsecond timestamp prefix. This is an observation-command precision
mismatch, not evidence of an application defect or of the full identity gate
passing. No retrospective acceptance or permissive timestamp tolerance was applied.

The exception occurred before the command's try/finally sampling block, so it
also skipped that explicit writer/process disposal path. The observation host
ended with exit1; native receipts and a retained-handle exit verdict are absent.
Future runner preparation should place all post-launch validation inside guaranteed
resource disposal, without forced child termination. No helper/source repair here.

## Normal close and aftermath

The uniquely returned window matched the exact preserved executable and reported
zh-Hans/unknown-usage inactive. Screenshot/accessibility showed initial Login and
the offline notice. Used its freshly observed titlebar Close button once at
18:40:23.463Z. No Enter demo, scenario selection, text entry or connection action.
Immediate window inventory was still present; no repeated close was sent.

At18:40:31.6774635Z independent CIM inventory found zero demo/prototype targets;
at18:40:38.581Z fresh window inventory found none. This confirms no targets remained
at those observations, not actual exit code/time or original-handle completion.
No kill, restart, debugger, settings change or replacement launch.

Preserved EXE/DLL SHA256 still match the approved hashes after closure.
Ignored raw file under the old artifact directory:
process-observations/9635909714294fbdb4f31cb305ae84fb/samples.jsonl;
length0, SHA256E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855.
Do not rewrite the empty receipt or claim it contains samples.

## Results and next scope

Preflight: passed. Post-launch identity gate: failed as above. Initial/final Active,
Unknown, Stale, off-page activity and numeric restoration: NOT_RUN. Native exit
verdict: UNVERIFIED. Overall C5: NOT_ACCEPTED. Shutdown cause remains unresolved.

Next proposal is a test-only precision/cleanup check before any new launch: compare
timestamps at the explicitly declared CIM precision while retaining exact original
handle/native creation identity, and prove wrong PID/path/arguments and genuinely
different creation times still reject. Ensure all post-launch abort paths dispose
observer resources without killing the target. This needs a bounded repair scope;
no retry or remaining allowance is carried forward. No product change is indicated.
