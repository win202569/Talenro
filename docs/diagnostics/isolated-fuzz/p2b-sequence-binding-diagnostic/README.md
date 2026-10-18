# P2-B sequence-binding diagnostic — one new evidence-only host

Approved by the user's continuation of the separately bounded diagnostic
recommendation after 5431f760. One ordinary pinned PS7 host; no retry, fixes,
reviewer, full-suite run, fixture compilation, install or acceptance.

Uses unchanged sealed candidate F76B2DB7 and literal fixtures. The helpers
are a definitions-only excerpt from the original full test through BC-Bind.
Runs only the original SequenceExact operation order: low valid binding,
two sequence mismatches, high UInt64 encode/decode equality, then captures
actual high binding result rather than aborting on its final equality.
Preceding full-suite groups are NOT replayed, so a different observation
does not erase or explain the historical failure by itself.

Raw Message/Expected inventories use fixed SDK name/type/flag fields and
set-reference identity, with no descriptor Value/MemberType/Copy getter
reads. Inventories run only AFTER the final actual binding call to avoid
priming its metadata history. API internals necessarily read trusted
application values. Captured outcome means evidence obtained, NOT tests green.

Controller refuses an existing packet, preconsumes one invocation, ordinary
parent/pinned engine, inherited VerificationPath mapping checks, only its owned
child, separate concurrent raw streams; 10000ms execution + 2000ms owned exit +
2000ms drain. 981 pins = 963 inherited + 13 stopped packet files + unchanged
live candidate + four new diagnostic inputs. No old allowance resets/replays.

Result: docs/roadmap/2026-10-18-p2b-sequence-binding-diagnostic.md.
Preserve all source/raw/oracles/reviews; source/report portable, ignored raw/
private caches/sessions/runtime not recoverable by cloning. No main/live grant.
