# Offline I1/I2 single-owner repair — UNFINISHED, DO NOT RUN

Approved plan: docs/superpowers/plans/2026-11-01-offline-i1-i2-single-owner.md.
Design, Native execution and finite budget are approved; historical failures and
older budgets are unchanged. User also permitted pure authoring corrections
before the FIRST source freeze, not post-freeze retries/provenance relaxation.

All31source paths now exist. The49case I2 body and parent predicate calls are
authored, but driver preparation, writer-unit and recorder adaptation remain.
Writer/owner and validator APIs remain intentionally definition-free seeds.
No source freeze, availability RED, capture host, target, compilation or runtime
review has run. Syntax/JSON/LF checks alone do not establish functionality.

Planned: two tool-context availability REDs (1FAIL/25NotRun and1FAIL/51NotRun),
then5capture hosts/7targets/1in-memory fixture compilation/conditional1review.
Before RED, freeze all31sources and independent26/52group expectations; later
windows change only2writer/owner files, then1validator file, for93snapshots.

The actual capture publisher must have no collector, return live outcomes and
use the real source-bound default factory. The capture owner consumes once.
Writer tests inventory12private scopes: ten operation cases, owner controls and
one final-write fault scope; recorder stage bound stays32. No extra host/target.

Next: finish driver seed/source windows, writer/recorder evidence and fixed
unit-source contexts, then all prefreeze static checks BEFORE consuming REDs.
Do not use copied old driver grants or claim successful repair from this leaf.
