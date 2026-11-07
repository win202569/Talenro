# Saved graph versus diagnostic coverage

Status: READ_ONLY_COMPARISON_COMPLETE;source807de05a;44/34/6/217,NOT_ACCEPTED.
No Go,build,test,cache inventory/content reads,tool changes or new runtime.

## Verified inputs and comparison

Tool8901f5 decoded all13 graph archive entries and checked length/SHA256.
Archive SHA25685516ED24C7A2F842EC6B2839F3C0F15EC3FBA172506517E1BD714D857F78121.
Root go.mod SHA256AF6A9F41F70B5ED37A39598917BB20BBBEFB14F368D8F17EB7AD3B45B46DA9F5.
Saved output contains1059 edges;101 originate from talenro.local/platform.
All99 root require entries exactly occur among those root edges;the other two
are go@1.26.0 and toolchain@go1.26.5. No missing declared pair was observed.
This is textual declaration/edge agreement,not a reconstructed selected BuildList,
an inventory of cached modules,or a claim that every dependency is present.

Current helper main.go SHA256
68D41F0FA265CB123C5554A98A47BF79F80DF5C656620D586E123956272BE421;
phase.go SHA256E397B428BB8D34BBE88B3970067A71CD7C602663BAE797D28866EC2C2F37D010.
Its only real candidates are x/text v0.40.0 and sqlite v1.49.1. Run iterates these
explicit candidates and hashes each ZIP/directory;it does not traverse dependency
edges. Therefore sqlite success did not validate libc or sqlite's other dependencies.

## Evidence-backed candidate selection

| Object | Saved evidence | Coverage / recommendation |
| --- | --- | --- |
| modernc.org/libc v1.72.1 | Root edge and goose v3.27.1 edge;sqlite v1.49.1 requires v1.72.0 | Not covered;first proposed next diagnostic object |
| github.com/klauspost/compress v1.19.0 | Root edge;goose requires older v1.18.5 | Not covered;defer to keep next scope to one object |
| golang.org/x/text v0.40.0 | Root edge;prior selected-cache run | Already measured;no repeat proposed |
| modernc.org/sqlite v1.49.1 | Root and goose edges;prior selected-cache run | Already measured;does not include libc content |

The libc choice follows an adjacent uncovered dependency of the measured sqlite
sample. It is NOT ranked by file count,size,cache presence or observed slowness:
none of those were measured here. Its ZIP/directory could be missing;do not fetch
or repair if a future prerequisite check finds that. No root-cause claim follows.

Read-only c0bcc0 finds libc v1.72.0/v1.72.1 and compress v1.17.7/v1.18.0/v1.18.3/
v1.18.5/v1.19.0 among graph endpoints. Edges retain historical requirements;
do not run all versions or interpret the list as native verify's selected list.
v1.72.1 is an explicitly root-declared candidate,not a claimed BuildList extraction.
Root go.sum contains libc v1.72.1 h1:db1xwJ6u1kE3KHTFTTbe2GCrczHPKzlURP0aDC4NGD0=;
compress v1.19.0 h1:sXLILfc9jV2QYWkzFOPWStmcUVH2RHEB1JCdY2oVvCQ=.
These are recorded expectations,not newly computed cache digests.

## Tooling boundary before another real run

The Go CLI fixes two candidates in main.go. owner.ps1 also fixes those identities,
six phases and the terminal slot count at lines25,30,49. Changing only Go candidates
would cause the existing protocol validator to reject different identities/counts.
Generic phase.go accepts lowercase module paths and plain vN.N.N versions;libc's
identity fits these syntactic constraints,but real input prerequisites remain unknown.

Recommended next work is a separately approved design for a fixed single-libc
diagnostic profile,with matching producer/parser contracts and offline fixtures.
Do not silently broaden the current profile or make it accept arbitrary modules;
preserve the old two-module protocol/evidence and its safety checks. Design must
define source changes,RED/GREEN cases (identity/order/partial/terminal/count),
finite build/live budgets and new artifact binding before implementation/runtime.
No such changes or budgets are authorized by this comparison report.

Success would describe this module's PREPARE/ZIP/DIRECTORY only,not root integrity
or the prior timeout cause. If broader root attribution is required instead of
sample coverage,choose a separate instrumentation approach rather than claiming
that accumulating successful samples proves the full command healthy.

Systematic debugging's evidence-first comparison used;author review only.
All prior evidence retained. Next decision: approve designing the single-libc
profile;do not launch additional root verification or cache work automatically.
