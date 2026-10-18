# Isolated scalar-owned-return candidate — not promoted

Approved bounded design: roadmap/current-status.md at development commit
7bbaf3a683af924bcd7977f3533b2ccd49bc1bc1. Original Token candidate and old
first-case failure remain untouched. This leaf is not production replacement.

Three new helpers create fixed ownership metadata, rebox exact Int32/Int64/Boolean
values through typed storage, construct nonempty String values through fixed SDK
metadata, and rebuild exact13-field root/22-field Decision results with independently
stored scalar leaves. Both normal and catch/refusal returns use the builder.
Input validation/reduction logic and public signature are retained, not bypassed.
No caller metadata is copied, no serialization/global cache/type-data change.

The focused suite was prepared before the new candidate. It has15 scalar-copy
vectors and39 ownership cases/max80 product calls, exact CLR/full-literal oracles,
direct post-return attachments and payload-retention checks. It includes the old
18-case matrix,3 sticky Fault->Reason directions, and9 representative return leaves
in A/B directions. Null/prepared/refused outputs remain conservative. New unique
metadata names prevent historical selected metadata from confusing later cases.
The old first-case failure is retained RED evidence, not a fresh RED/mutant run
of the additional assertions. Partial representatives are not exhaustive proof.

Full API(all) and previous ownership(regression) files are exact pinned routing
copies. Their sibling module path selects this candidate without editing their
assertions. source-pins.json names the originals. expected-output.json independently
lists98 focused,38 full and22 ownership lines; Context is filled by the controller.

capture.ps1 prepares frozen new/prior identities then permits only three ordered
ordinary fresh PS7 hosts: focused, full, ownership. Each30s wall+2s owned exit+
2s drain. Any failure stops later slots; no retry, warmup, re-review or merge.
Use the latest result report, not these source files' existence, to decide whether
an allowance is unspent/authorized. Cloning does not renew execution authority.

Raw captures/frozen manifests/index live in an ignored local packet; Git carries
portable sources and the roadmap report, not runtimes, raw state or credentials.
No PS5, benchmark, whole-project/security/runtime/I1 acceptance is implied.
