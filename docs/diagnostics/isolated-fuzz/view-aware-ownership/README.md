# View-aware ownership regression (not formal acceptance)

Approved after design d10d39ee. Candidate 1971F808 stays byte-identical.
No product changes, old baseline rerun, new product mutant, review, promotion or main merge.

The test catches scalar base sharing and later descriptor/payload contamination.
Independent complete 13 root / 22 Decision literal oracles are retained from the
prior test. Its 15 scalar vectors and 39 behavioral cases are unchanged.
String inputs selected by the matrix are explicitly wrapped BEFORE evaluations;
no selected input or output is replaced after evaluation or attachment.
Strong public views, member collections, field descriptors, storage types and
base references are retained before mutation. Non-String storage requires held
AND reread retention; actual raw String storage requires effective held mutation,
fresh bases and cross-owner held/reread checks. Reread loss alone proves nothing.
Public-view identity is a proxy, not a full CLR binder/storage trace.

Two fixture-only shared-String-wrapper / shared-boxed-Int32 positive controls run
before candidate calls, with independent clean counterparts. They validate known
sharing visibility, NOT candidate-mutant sensitivity. Historical Int32 RED is
retained; revised assertions have no fresh product-negative-control proof.
53 fixed-schema facts (14 non-null vectors + 39 behavioral cases) accompany
153 ordered output lines. Scalar null + two normal controls have no mutation fact.

## Budget and recovery

New allowance: max three fresh ordinary PS7 7.6.5 x64 hosts in order:
focused (80 decision, 29 direct scalar, 1 direct metadata helper calls),
unchanged full API -Group all (35 groups/38 lines), unchanged ownership
-Group regression (18 groups/22 lines). Internal helper calls inside decisions
are not part of the direct helper counts. Each host: 30s + 2s owned exit + 2s
stream drain. Preconsume before launch. Any failure freezes later slots: no
retry, repair, extra probe/host/reviewer, or old frozen slot reuse.
All 566 prior pins, four latest raw anchors, six immutable source anchors and
these six portable files are frozen before first launch (582 pin entries).

capture.ps1 Prepare/Run is a historical bounded controller, NOT an automatic
fresh-computer launch instruction. It requires the original ignored pin chain,
so a clone cannot recreate or renew consumed allowances. Portable sources and
oracles suffice for code inspection and a separately approved future run.
Raw receipts remain ignored; the roadmap report records verified outcome/hashes.
Even all passes do not close I1 or grant security/runtime/main/C12/release/PS5/
performance acceptance. No whole-project run is authorized in this allowance.
