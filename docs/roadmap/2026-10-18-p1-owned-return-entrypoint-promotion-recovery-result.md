# P1 owned-return：原离线入口提升完成（限定 PS7 开发分支）

日期：2026-10-18。恢复基线：f436d139b432722e0c4e2cb19dd91ada918acd70。
结果：PROMOTED_ON_DEVELOPMENT_BRANCH_PS7_LIMITED。
原预算两次目标宿主均完成，2/2 已用尽；没有增加次数、重跑目标或新增独立复审。
源码及本报告已保存于 5a9dc4dec616600239c2e569ca3f82bc2cbc4206，并推送原开发分支；
远端同 SHA、工作区干净已实际核实。本次仅补记该保存确认，最终文档提交在交接中报告。

## 批准、恢复方式与唯一产品变更

用户先在 cca8c0c6 后批准[原提升计划](../superpowers/plans/2026-10-18-p1-owned-return-entrypoint-promotion.md)。
准备因备份换行错误停止后，在 f436d139 已推送的明确问题
“是否允许仅修正备份及必要准备文件，保留失败记录，在原最多两次测试预算内继续？”
之后再次回复“是”。这是本次恢复授权，不是增加两个测试名额。

没有覆盖失败文件：原 owned-return-entrypoint-promotion-v1/ 六文件、原[停止报告](2026-10-18-p1-owned-return-entrypoint-promotion-result.md)
及其 ignored stop ledger 全部保留。另建 owned-return-entrypoint-promotion-recovery-v1/
六文件和新的本地 packet，正确 baseline 先核验为原 LF 字节后才覆盖产品入口。

唯一产品变更：docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1
从 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA
替换成已验收候选的原字节：
1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8。
没有算法重写、引擎 guard、接口/测试/独立预期改动；候选原文件不变。
产品与新控制器静态 parser 均 0 errors，不通过导入产品来进行准备检查。

## 实际原目录验证：同一原预算 2/2

| 顺序 | 原目录目标 / Group | StartUTC | 耗时 | 保存结果 |
| --- | --- | --- | ---: | --- |
| 1 | offline-supervision-policy.tests.ps1 / all | 2026-10-18T14:11:10.9299533Z | 21408ms | exit0，38 有序行、35 PASS 组 |
| 2 | offline-supervision-policy.ownership.tests.ps1 / regression | 2026-10-18T14:11:40.6832617Z | 13797ms | exit0，22 有序行、8 fixture+10 regression 组 |

两个目标均 Windows x64 PS7.6.5 Core、普通继承令牌 admin=False；每个执行上限
30000ms，加 2000ms owned exit 和 2000ms stream drain，未加时、无超时。
各次消费均先持久记录再 launch；两通道并发 raw-byte 捕获；stderr 均为空。
owned-host-exited、完整流、精确独立有序预期、762 个冻结检查均成立，ControllerError=null。
第二次仅在第一次完整通过后启动；测试真实路径在原目录，不是候选目录替代执行。

full 测试 SHA=61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3；
ownership SHA=C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7。
冻结使用既有独立 38/22 行 oracle，只预先解析普通 CONTEXT；未从产品输出生成期望。
父代理另只读逐行对照两份完整 stdout、stderr/回执/摘要及全体冻结路径，均一致。
这些是限定内部证据，不是执行认证、性能基准、覆盖率百分比或未来时限保证。

## 历史 pin 迁移：不是忽略 hash mismatch

恢复前 66+676 原条目仍全部匹配。提升后保留全部 742 条原路径/摘要，其中精确
等于原产品路径的 3 条标为 historical-destination，且只映射到正确 baseline
953CC2A8；另 739 条原样校验。新入口另钉选 1971F808，不借旧摘要覆盖新产品。
额外 9 anchors 保留旧准备六文件、停止报告、停止 ledger、限定验收批准记录。
加 5 项当前身份及新目录 6 文件，总共 762 个冻结检查，前后均核对完整。

没有修改旧 manifests、索引、回执、测试、报告或旧 pin JSON；不声称 3 条旧路径
在当前工作区仍对应旧字节。旧控制器若要求原 SHA，须使用对应历史快照，而不是
对新入口放宽检查或自动改路由。新控制器也没有通用排除/忽略不匹配的分支。

| 备份 | 实际 SHA256 | 状态 |
| --- | --- | --- |
| 新 recovery-v1/baseline-policy.ps1.txt | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA | 正确 LF，先验证后覆盖入口 |
| 旧 promotion-v1/baseline-policy.ps1.txt | 54A07EB214250AF4EED0C03496B67D259AA68DA1BA4AC3665CBB1FDCADEFEE81 | 错误 CRLF，原样保留，不改成通过 |

## 保存证据及换机行尾边界

新 ignored packet：.superpowers/sdd/2026-10-18-p1-owned-return-entrypoint-promotion-recovery-v1/。

| 文件 | SHA256 |
| --- | --- |
| frozen.json | 9F47C38196A3866FCE4918420A29CA6DAFF3E08300C0EA1D46573CE9069BCDD9 |
| invocations.json | 93C121C6D8E2F6727F7C9047FD46683F8707D47EF35022D6A5B04602C23E27E5 |
| full.stdout.txt | 3FE31C43E6947167D06A1C03C94A382856B249C36A1FAA0DBF2BBD6C99342D76 |
| ownership.stdout.txt | 0571AD4FEE3BA556B3D5A29D0E36A4C7AA5C3CAA95E97F2FF0556DA8D53464B8 |
| 两份 stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| capture.ps1 | 69622265C81F4E35F5D123F0AA2313DCA493FF60325F1AE2C0E52D4D430BF165 |
| source-pins.json | 8DFE8E9EA4F34675A5434AED3964B49A7DCB862A53C9B59C34D6BDBA08025496 |
| expected-output.json | 07358AE6468C8E7AF2B348ABBCDA5F2557B126CF3C35FFE54CDD19D96D83CD10 |
| README.md | 268652FC500874C5943E5CAA7978BD5D4374B664E6D55D1B9FF679AF8FD8B0D6 |
| .gitattributes | 4BB217115CC58733891AB567A92E208F6607D1910126F367FAE83A2994579427 |

已存在的原入口属性是 text/eol=crlf，本轮未改该属性。测试现场 LF 物理摘要为
1971F808；同一内容按 Git 既有属性签出为 CRLF 的摘要为
C82EE1D8846961567227EFA602B2B065C8F858ED764B21CF3E033CCBE712E1BA。
这延续已披露 N8 行尾限制，不把新签出文件冒称执行过的 LF 字节；提交时已核对
规范化 Git 内容与现场 LF 一致，并核对签出过滤内容对应上述 CRLF 身份。
正确旧版本文本通过 -text 保留 LF 原字节。没有为使验证通过而改产品行尾/摘要。

换机从本开发分支恢复本报告、源码/测试、两种备份、原批准及恢复准备文件。
原始 packet、引擎缓存、凭据和对话不随 Git 恢复；固定原执行摘要不自动适用于
任何换机文件。不要用新 checkout 冒充旧原始记录复核，也不要因 clone 重跑已用预算。

## 保留限制与流程处置

[限定验收批准](2026-10-18-p1-owned-return-ps7-limited-acceptance-approval.md)及其 O1–O23/N1–N9 全部保留。
I1/R1 仅 AddressedWithinStatedPS7Vectors，原历史问题/失败不全局改判。
18 条 raw String 仍只支持 held-view，不支持 fresh-field reread 元数据保留。
此次未重新跑 focused39-case 或 observer45-case；其证据只用于同字节候选范围。
PS5/完整双引擎 P1、新产品变异、未覆盖状态/叶子、并发、灾难恢复、真实时钟/
身份/采集/取消/清理/持久化仍未证明，P2/P3/main/live/C12/安全/发布不获批准。
所有实际产品授权字段仍 false，CleanupStatus=unknown；Would* 不是真实行为。

Ruling：按用户明确恢复批准另建目录，不修旧失败记录；代价是新增恢复材料。
Ruling：固定已审字节采用且预算仅两次，覆盖通用技能的额外 RED/全项目/task-done/
fix/review/cleanup 循环；代价是无新增产品变异、提升后独立复审或全项目保证。
Ruling：PowerShell-native 记账，不用 Bash skill helpers；代价是无 generic runner 完成声明。
Ruling：旧目的路径仅映射正确旧备份；代价是历史消费者需其原版本，不兼容任意旧控制器直接重跑。
没有新增审查分级或偷偷修复/延期 Minor 问题；没有 scratch 删除。

开发分支原入口已按获准范围采用候选，不等于完整 P1 或项目完成。
下一项实质工作需另定范围；可优先整理 P2 的前提/接口设计，但本轮不自动启动
设计、真实适配器实现、运行或 main 合并。若目标是完整 P1，应先处理同候选 PS5 门槛。
