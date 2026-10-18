# Bounded inbox — approved tick-vector correction and remaining4+1

User approved after cf347c4c: preserve original stop/RED, correct only origin
mismatch vector ObservedTick4 to11 when OriginTick10, refreeze and use remaining
four original slots full/capacity/sequence/final plus conditional readonly review.
NO new RED, no extra allowance. Original source/tests/controller/records untouched.
Product43067AD0 byte-identical, unverified until exact full results. Fixedcodec
9A532E17 unchanged. Other tests/helpers/fixtures/oracle/target byte-identical.
New controller only binds preserved RED and new input seal, four-slot order.
120000ms execution+2000ms ownedexit+2000ms separate concurrent drains, ordinaryPS7.
Any unexpected result freezes remaining slots/review; no repair/retest/reviewer loop.
Mutants restore immediately and retain exact source identities.32groups/35lines.
Same finite inbox design, no pipes/processes/clock or runtime authority. Allflagsfalse.
Raw packet .superpowers/sdd/2026-10-18-p2b-bounded-inbox-tick-correction ignored.
Portable source/results/reports saved on existing devbranch, no main/acceptance.
Clone restores no raw/actualreferences/tokens/cache/SDK/credentials or replay rights.
