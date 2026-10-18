# P1 owned-return：PS7 限定离线功能验收批准记录

日期：2026-10-18。批准前开发提交：fbcd2409d6361b416ae0b2ec468746ef18900085。
状态：USER_APPROVED_PS7_LIMITED_OFFLINE_ACCEPTANCE。
LimitedOfflineAcceptance=true；FullDualEngineP1Acceptance=false；PromotionApproved=false。

## 1. 实际批准来源与固定提案

父代理在上述提案提交已推送且远端一致后明确询问：

> 是否批准提案第 1、4、5、6 节，并仅记录这项限定验收？

该问题此前明确说明推荐 A：仅验收固定候选的指定 PS7 离线范围，明确接受尚无
新增产品变异测试的证据缺口；保留全部限制及 PS5 门槛，不替换源码、不合并
main、不授权真实运行。用户紧接着回复：**是**。

因此记录为用户明确批准该限定结论、A 路线两项范围决定及全部随附限制，
不是从历史泛称“继续”、测试 PASS 或提交动作推定批准。

批准对象为[原提案](2026-10-18-p1-owned-return-ps7-limited-acceptance-proposal.md)
在 fbcd2409 中的固定文本：
SHA256=3FA358474D96B0AD54E7B70C8479A9139997C5C9ECC1F760747155C68FF62C28。
原提案保留其起草时 PROPOSED_NOT_APPROVED 状态，不回写历史；现行批准见本记录。

## 2. 已批准的精确结论和版本

固定 owned-return 候选版本，在已有 Windows 普通用户 PS7 保存证据所覆盖的
合成输入和顺序调用向量内，完成 **PS7 限定离线功能验收**。
采用同版本 API、旧所有权回归、后返回值矩阵及限定独立审查的组合证据。
I1/R1 仅记 AddressedWithinStatedPS7Vectors，不全局关闭历史问题。
不宣称原 PS7/PS5.1 双引擎 P1 合同已完成。

| 固定文件（docs/diagnostics/isolated-fuzz/scalar-owned-return-candidate 下） | 本机已验证物理字节 SHA256 |
| --- | --- |
| offline-supervision-policy.ps1 | 1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8 |
| offline-supervision-policy.tests.ps1 | 61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3 |
| offline-supervision-policy.ownership.tests.ps1 | C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7 |

引擎身份及其余矩阵/observer/审查身份由固定提案和所引报告界定，不推广至所有
PS7 版本、其他平台、后续修改、当前原产品 953CC2A8 或 Token F9B4E408。
不是整个分支、项目、C12 或既有 store/raw coordinator 的验收。

## 3. 明确获准的两项范围决定及保留限制

1. 选择提案 A 的 PS7 限定路线：PS5.1 兼容性仍未验证、未验收；不等于已证
   不兼容。完整双引擎 P1 门槛继续开放，没有删除原合同或改写旧测试。
2. 明确接受没有新 owned-return omission/product mutant 的证据缺口，仅不把它
   作为本次狭窄文档验收的先决条件。采用正向 backing/descriptor/payload 检查
   和限定独立审查，不冒称产品变异已通过，不永久豁免未来推广的证据义务。

提案第 5 节 O1–O23、N1–N9 及[独立审查全部处置](2026-10-18-p1-owned-return-readonly-review.md)
完整随批准保留，没有一项转成“已测试通过”。特别包括：保存记录不是执行认证、
有限向量不是穷尽、合成时钟/actor/cleanup/readback 不是真实事实、并发/灾难恢复/
未覆盖 binding/descriptor 不保证、单次 timing 不是基准、换机行尾影响物理身份。

String 事实不变：32 条原生标量与 3 条 stored PSObject String 的 held/reread
均保留；18 条 raw String 只支持 held-view 保留，fresh-field reread 不保留。
原 I1/Token R1 审查、污染 witness、旧 fixture/timing 失败、停止和额度记录不改写。
最新独立审查仍只是源码与保存证据的限定判断，不追溯称其作过正式验收或见证执行。

## 4. 不获准的动作和不变机器字段

批准只更新文档层面的固定版本限定验收状态。产品仍保持 Mode=policy-only、
EvidenceClass=synthetic、OriginVerified/Acceptance/RuntimeApproved/
LiveActionsPerformed=false、CleanupStatus=unknown；Would* 仍仅为假设决定。
不回写任何历史回执的 FormalAcceptance=false。

不授权源码替换、候选提升、main 合并、发布、PS5 兼容声明、P2/P3 实现/实验、
真实采集/历史载荷/取消/清理/持久化、安全设置/信任变更、新测试/变异/修复/
重跑/审查/安装或新额度。旧测试/审查已用、冻结或不合格额度全部保持不变。
有限验收不是项目完成或运行、安全 owner/vendor、Stage B/I2/I3 放行。

## 5. 本轮实际核对与保存

本轮仅核对批准对象、Git 基线/分支/origin、固定提案及审查身份，并将 66 个
审查输入条目与 676 个继承条目逐个核哈希，均未变；不是重新运行测试或复审。
仅新增本批准记录，更新 current-status 和 repository-recovery，无源码、测试、
原提案、原报告、ignored 原始记录或权限设置改动。
保存至既有 codex/c12-b01-task9-coordinator 开发分支，不到 main。
实际提交与远端一致性在提交推送后另向用户报告，不在本文件预写成功结果。

换机可从开发分支恢复必要源码、测试、提案、报告和本批准记录；原始回执、
引擎缓存、凭据不随 clone 恢复，不得冒称新电脑已复核旧机器原始记录。
本次批准不自动启动下一项工作。后续若需要提升候选/替换原产品，应先单独
提出精确变更与证据要求；若要完整 P1，先单独处理同候选 PS5/双引擎门槛。
