# Fixed libc profile — implementation progress

Source plan approved after e7e2a9b9. Global44/34/6/217,NOT_ACCEPTED.

## Go selector checkpoint

GoRed08a212/891406:29.693s,exit1,only TestProfileLibc assertion failed with PROFILE
against the declared legacy-only stub. GoGreen5075e6:2.428s,exit0,diagnostic-local
package suite passed (existing16 and new5 top-level tests). Both owned exits
confirmed,no timeout/failure,four module hashes unchanged. Raw stdout/stderr read.
New fixed resolver/CLI parser preserves paired-v1 default and rejects unapproved
names without filesystem access. main resolves options before cache candidates.
No phase hashing or safety code changed. No real libc/build/owner tests yet.
portable-c12-libc-profile-go.json preserves exact RED/GREEN sources and receipts.
This is a partial implementation checkpoint,not final profile or C12 acceptance.
