# Extent contract rebind — preflight blocked

User approved implementation plan at881bdd0ac965f34987334107cc4e7c6d2ac18cb2
with “是”. Read-only preflight found a plan identity collision before any
new approval/input/intent artifact or operational edit was created.

## Evidence and effect

The approved design reserves extent-contract-green for the single contract host.
The plan's provenance entries24–26 also assign extent-contract-green.intent.json,
extent-contract-green.final.json and extent-contract-green.tool-record.json to
auxiliary GREEN19. Task3 requires those host paths to be absent before launch.
Thus successful GREEN19 would consume the host namespace, and CreateNew/attempt
guards would correctly refuse the later host. Replacing the auxiliary receipts
would destroy evidence and is forbidden. This is a planning defect, not a test
failure or an observed runtime defect.

## Proposed minimal correction — pending approval

Keep the approved design and host slot unchanged. Use extent-contract-binding-red
and extent-contract-binding-green for auxiliary18/19 only, including their
intent/final/tool-record triples. Replace plan provenance entries21–26 with these
six distinct paths. Keep all six receipt roles and the30append/2306total count.
All runner/gate/candidate/map references to auxiliary receipts use these new
names; host paths remain extent-contract-green. Add an explicit disjointness
assertion across auxiliary/Prepare/host output sets inside the existing
CreateNewIdentity group, without adding a run or group.

Do not rewrite the approved plan in place. If this correction is approved, record
this addendum's SHA256 in the new approval/input and recovery map before RED18;
the already planned approval/input pins transitively bind it, so pin count stays
2306. Preserve original spec/plan bytes and both approval decisions.

## Actual state and resumption

- Task1: not executed; preflight only. Tasks2/3: not started.
- No source/test/helper edits; no new extent-contract P artifacts, no test call,
  freeze, process start, compilation or reviewer.
- Counters unchanged:17auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews.
- New authorized2aux/1host/31snapshots remain unconsumed. Old nine stops and all
  failed records remain untouched. No runtime stop was fabricated for this
  planning-only finding; do not treat absence of one as permission to run.
- Product NOT_ACCEPTED. Resume only after deciding the corrected auxiliary
  identities, then carry out preparation and the approved finite sequence.

Ruling: preserve the approved plan and pause before artifact creation, rather
than silently changing exact approval-bound receipt identities. This costs one
clarification and prevents an avoidable consumed-slot failure. No skill-driven
extra review, baseline test, cleanup or installation was performed.
