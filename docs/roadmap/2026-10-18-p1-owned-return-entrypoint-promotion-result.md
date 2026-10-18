# P1 原入口提升：准备阶段停止，尚未替换源码

日期：2026-10-18。执行基线 cca8c0c619720c678b53b31ebc7949a336943537。
状态：UNFINISHED_STOPPED_BEFORE_PROMOTION；PromotionCompleted=false。
新目标宿主 0/2 已用，两个未用名额冻结；没有产品或测试运行失败的结论。

## 批准与实际步骤

用户对“按计划保存旧版本、替换原离线入口、最多两次 PS7 回归、失败即停、
沿用既有审查且不新增复审、不合并 main”明确回答“是”。
批准对象为[提升计划](../superpowers/plans/2026-10-18-p1-owned-return-entrypoint-promotion.md)
在 cca8c0c6 的文本，执行前物理 SHA256：
DB37257895A1215C50F3BF5265F10C64E16D2169AB42D95E2BC5AE175FA4B51D。

Step1 完成：隔离 worktree、干净 HEAD/分支/GitHub origin、固定引擎、66 审查
输入+676 继承条目一致。Step2 新建 6 份准备文件后，在覆盖产品之前核对 baseline
原字节失败，故停止。Step3–7 未执行；没有产品导入或调用控制器 Prepare/full/ownership。

## 已确认原因：父代理准备错误，不是环境阻断

我在整理新 PowerShell/JSON 文件格式为 CRLF 时，错误地把要求原字节留存的
baseline-policy.ps1.txt 一并转换了。原产品为 LF；因此备份摘要发生变化。

| 对象 | SHA256 | 字节/CR/LF |
| --- | --- | --- |
| 未变原入口 | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA | 18734 / 0 / 227 |
| 错误备份（保留） | 54A07EB214250AF4EED0C03496B67D259AA68DA1BA4AC3665CBB1FDCADEFEE81 | 18961 / 227 / 227 |

两者仅在内存归一化换行后相等，这不是原字节一致，不能通过。失败命令在此终止，
后续 oracle/mapping 检查当时未执行，不能记成通过。停止后只读确认上述事实，
静态解析新控制器 0 errors，并重新核对全部 742 个原证据条目仍匹配；没有修正文件。

## 保存对象、状态与边界

新目录 docs/diagnostics/isolated-fuzz/owned-return-entrypoint-promotion-v1/ 中：
baseline-policy.ps1.txt 保留错误字节，capture.ps1/source-pins.json/
expected-output.json/.gitattributes 保留未完成准备版本，README 记明停止。
source-pins.json 的 baseline SHA 是预期 953CC2A8，不是错误备份的实际身份；
不能冒称已 frozen 或可执行通过。742 项中拟议映射的 3 项尚无有效旧字节备份支持。

新 ignored packet 仅有停止 ledger：
.superpowers/sdd/2026-10-18-p1-owned-return-entrypoint-promotion-v1/progress.md。
它在失败后记录状态，故 capture 的 New packet only 检查会拒绝复用此路径。
没有 frozen.json、invocations.json 或 stdout/stderr；禁止删除停止记录来重试。
没有把未运行预算记为消耗；0/2 已用但两项冻结，不等于自动可继续。

保存时普通 Git whitespace 检查也把错误备份的 227 个 CR 行尾报告为 trailing
whitespace。为保留失败原字节，不清洗此文件：其余 9 份文件仍按默认检查，备份
单独识别 CR 行尾并核对实际摘要；不据此宣称备份修复或原字节校验通过。

原入口仍 953CC2A8；候选仍 1971F808；所有旧测试/报告/pin JSON/packets 不变。
[候选 PS7 限定验收](2026-10-18-p1-owned-return-ps7-limited-acceptance-approval.md)
不因准备失败被撤销，也不扩展到原入口提升或完整双引擎 P1。
全部 O1–O23/N1–N9、PS5、产品变异缺口、main/live/P2/P3/C12 门槛保留。

## 流程处置与下一步

执行计划/工作区/TDD 通用流程不新增 RED、全项目验证、修复或复审；PowerShell
控制器替代 Bash helpers，保留全部记录。代价是无 generic runner 完成声明、无
新变异或提升后独立复审。系统化排查技能仅用于只读确认原因。
依 fail-stop 条款，换行错误也不自行修正；代价是提升在准备阶段停止。

推荐另行批准：仅修正备份为原 LF 字节及必要准备路径/元数据，另建不覆盖本次
停止记录的准备包，在原未消耗的最多两个名额内继续；不增额、不重跑已运行目标、
不放宽摘要或进程安全检查。该建议尚未执行或获准。

本轮保存为明确未完成的开发检查点，不能称提升完成。换机先读本报告；Git 交付
必要准备文件与停止说明，不恢复 ignored 原始目录、引擎缓存、凭据或自动运行许可。
