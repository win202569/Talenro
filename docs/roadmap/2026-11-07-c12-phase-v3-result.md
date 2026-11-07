# Phase v3 — verified binary and corrected owner

User continued afterbbf0ea4a to verify source/build correspondence and perform one
bounded real diagnostic. One helper invocation120s maximum,5s owned exit wait,
new exclusive .superpowers/sdd/2026-11-07-c12-cache-phase-v3/Live. No new build,
tests,retry,security changes,root verifier/devtools/Docker. Failures stop.

Read-only84d556/da832c verified all Go/module/vendor/provenance files match v2
successful build receipt and helper SHA25649943D88D223F41405173E0D9A22F0AA00FC972A26A7BF3DEC9F38343A22332C.
Corrected owner hash33782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E
matches stream GREEN's exact snapshot;10/10 result receipt checked. Common helper
hash77824603B0817D32B447C50CA3818F5E34E1924D8403A800073165C65081566F pinned too.

Ruling: reuse the unchanged already-built Go binary; separately bind corrected
PowerShell owner to its GREEN proof. Do not rewrite original build manifest or
bypass its SOURCE_CHANGED gate. V3 invocation records both provenances and checks
the complete compiled-source set before launch. Cost: two explicit evidence chains
instead of a single build manifest; no additional compilation/cache warming here.
Invocation recipe is retained in the portable archive; another computer must
prepare/rebuild under its own approved prerequisites,not execute this receipt
against missing private cache paths. Source checkpointbbf0ea4a already pushed.

Same fixed text/sqlite order,preparation/ZIP/directory events,read-only allowlist,
output quota and cleanup. Approved inline review/no full acceptance remains.
44/34/6/217,NOT_ACCEPTED. Initial state: real outcome not yet available.

## Result — selected phases complete

Invocation5163da/session33302 ->3ee142:exit0,56717.0329ms,no timeout,
owned exit confirmed,Failure=null. stdout3213bytes,stderr empty.
Session93f33bd88f9549f7b7549414be721888. All12 START/END events complete,
all6 phases OK,all4 content digests match expected root go.sum/ziphash values.
Four root/devtools module hashes unchanged; no additional workload followed.

| Module | Preparation seconds | ZIP seconds | Directory seconds |
|---|---:|---:|---:|
| golang.org/x/text v0.40.0 |0.2833952|0.1201420|12.3784326|
| modernc.org/sqlite v1.49.1 |0.6119459|0.7610604|42.4556856|

text digest h1:Ub2Z6/xjgF1WrYQz2nuITOEegKFtiIy+rieRJ5lHZKs=
sqlite digest h1:dYGHTKcX1sJ+EQDnUzvz4TJ5GbuvhNJa8Fg6ElGx73U=

Read-only verification8dcd14 independently compared raw events with receipt,
sequence/session and root go.sum,checked all current source hashes against claim.
Sum of measured phases56.6106617s;directory phases54.8341182s (96.86%).
Owner total includes launch/drain/cleanup and is not identical to phase sum.

Interpretation: directory hashing dominates this two-module sequential sample.
It includes enumeration/open/read/hash,so these timings do not isolate disk I/O,
antivirus,lock waiting or a unique root cause. Preparation does separate safety
enumeration from hashing,but directory hashing enumerates again. Cache state may
be warm from previous activity; this is not a cold or statistically representative
benchmark. The helper omits LoadModGraph and root verifier concurrency/other modules.
Four matches do NOT mean authoritative root verification or C12 acceptance passed.
No physical-disk/lock/antivirus attribution is supported. Overall NOT_ACCEPTED.

Recovery: portable-c12-cache-phase-v3.json contains exact invocation recipe,
current source,old successful binary build receipt,corrected owner GREEN proof,
original claim/process/stage records and raw events. No binary,module contents
or caches. All earlier failure evidence retained. This120s live slot is consumed.
Next recommended work is read-only synthesis of these timings and existing root
timeout evidence into a scoped root-verification plan; no automatic root rerun,
timeout extension,cache/security modification or new real diagnostic.
