# Historical path verifier — local repair passed

Status: LOCAL_REPAIR_VERIFIED; product NOT_ACCEPTED. User approved correcting
the unexecuted draft and using only remaining GREEN21 after bd0273f7. Version3
preserves original/v2 test bytes and stops. No RED20 replay or added allowance.

## Actual evidence

Original RED20 remains0PASS/1FAIL/17NotRun/exit1,Expected=true, actual saved
invalid absolute-path IOException; no setup/recording error. Original final SHA
5C038EA43525119695ABF6BD131FA76D75A63F5FD7F1EDA05DAF5C40C3D3409E.

Version3 preparation first parsed all four new scripts, then fixed31source and
2326immutable file identities. Corrected test uses BaselineSources for the live
relative fixture and explicitly checks presence/valid hash. Original2307entry
history remains unchanged and is exercised directly, not reconstructed.

Only operational edit is driver.ps1: added Driver-VerifyHistory and replaced the
faulty historical-file loop with its call. Verifier binds the three exact external
paths/hashes, resolves repository-relative paths inside the root, rejects invalid
path forms and conflicting aliases before hashing, and uses unchanged actual
Driver-Hash/Driver-CheckHash. No process start, file write or fallback in verifier.
Candidate checkpoint0113bff4375ff3f4468b8bbbe48e4003286a3467 was pushed/verified
before GREEN21.

GREEN21 actually ran once:18PASS/0FAIL/0NotRun/exit0,Expected=true; RootError,
SetupError,RecordingError all null. Coverage: saved absolute path, full original
history, other external identities, relative file, wrong/missing hash/file,
unauthorized absolute path, drive/root-relative paths, UNC/device paths, traversal,
alias conflicts/identity, malformed inputs, driver-only source window, actual
extracted gate-call execution, and reverse restoration of exact original driver.
Both full-history and actual gate-call tests invoke the production verifier with
real file hashes.31sources/2326immutable inputs unchanged pre/post computation.

- Version3 input SHA8C0352FDFFC6004C266DF2A84BBF3121D377AB24373027DE20F3B2EBDC0281F0.
- Test SHA244A76A226B8404D7DAA75B1B6F3178BBE0B2578F84B613E546D660F435BAE06.
- Actual GREEN final SHA49D575369E2CCCDBF2C432B180C6729309C8FCC4CD23FCEA00DFC287386B09FF.
- Repaired driver SHA0F46DF4AA5CF442535426A17FA65672AB21A34514D6ACE42CFDF71226472E10D.
- Actual input/result saved in P/history-path-v3-green.tool-record.json.
- Candidate portable archive577files/135directories/31zero-byte, decoded equal,
  SHA9CB80872F25C35BEB7DB8B279A3345C1D29EBD510F13A6B8C57CAD5C6F95BF79.
- Final portable-history-path-v3-final.json contains all actual P/Q originals and
  two required H inputs; latest status gives verified inventory and hash.

## Boundary and resumption

Cumulative21auxiliary/4hosts/3compiles/186snapshots. No new host/Prepare/snapshot,
compilation/reviewer, installation or cleanup. All12previous stops remain; no v3
stop was generated. extent-contract.frozen.json and extent-contract-green intent
remain absent. Old candidate/freeze identities intentionally do not bind repaired
driver bytes; do not rewrite their pins or run their stopped routes.

Local repair complete, full contract/product acceptance unproven. A new approved
recovery route and source freeze are required before any full contract host; the
previous unconsumed1host/31snapshots and downstream allowances remain frozen.
No main merge. Restore from latest development branch and final portable archive.

Execution ruling: version3 retains the failed v2 draft rather than overwriting it;
only approved fixture correction/namespace/approval plumbing changed in the test,
same18groups and original remaining21 allowance. Native finite execution omitted
generic extra suites/reviewer/cleanup as approved. No independent review claimed.
