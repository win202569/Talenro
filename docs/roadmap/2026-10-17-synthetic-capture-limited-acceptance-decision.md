# 完整合成记录归档 — 有限离线验收结论（用户已批准）

状态：**用户已批准本最终有限结论及其限制；本模块该版本完成有限离线功能验收。**
见[最终批准记录](2026-10-17-synthetic-capture-limited-acceptance-approval.md)。不等于发布验收，不授权main合并、安全放行或额外执行。

## 1. 已批准判定

正式结论：**此版本在所列环境和已批准限制下，完成完整合成记录归档的有限离线功能验收。**
用户已对提交d5b547e4中的本结论及限制明确回答“是”；批准范围不扩展至其他事项。

依据：[已批准验收附录](../superpowers/specs/2026-10-17-synthetic-capture-store-acceptance-addendum.md)
第3..7节。它已批准多批次固定版本证据的采用方式和第5节限制；这不是事后修改旧记录。
本次交付以包含下列固定文件的开发提交为准，代码基线274b152a未修改产品。

|固定文件|SHA256|
|---|---|
|synthetic-capture-store.ps1|508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C|
|原测试|406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF|
|补充测试|243B4673BF22614718D61FF92BB421F7EE09D993691D0803BB88EC8D11C4B87A|
|文件reparse测试|991D75D7926CF63CE70B954E76919568B4DB79541D3B5106D2AAD2AE26DEFECB|

## 2. 协调者核对结果

- 附录采用的[五组证据](2026-10-17-synthetic-capture-coverage-summary.md)及
  [补充证据](2026-10-17-synthetic-capture-supplemental.md)均保留各自版本、环境和原判定。
- [六次依赖回归](2026-10-17-synthetic-capture-dependency-regression.md)各引擎81/106/110项完整；
  所列12项源码/测试hash在最终核对时均一致，不以旧结果覆盖未知版本。
- [文件级重解析点补证及审查](2026-10-17-synthetic-capture-file-reparse.md)已按单独批准执行，
  两引擎各7项完整、无超时/重试；payload/marker真实文件链接、正常对照及拒绝均有证据。
- [实现审查](2026-10-17-synthetic-capture-readonly-review.md)、
  [补充收尾审查](2026-10-17-synthetic-capture-closing-review.md)和本轮专项审查各自限定范围；
  未确认未处理Critical/Important产品或新增测试缺陷。先前指定验证缺口已取得约定证据。
- 未发现影响上述映射的源码漂移；没有新增代码修复、伪造结果、跳过必要检查或隐去冲突记录。

据此，附录第6节原保留的文件级reparse必要补证及约定审查现已具备，
可以提交本有限判定；不是声称原停止路线已全部通过。

## 3. 环境及限制随结论保留

- Windows x64，观测引擎PS7.6.5 Core、PS5.1.26100.9711 Desktop。
- admission/byte/event/格式/事件及写入补充/模拟器/合同证据为普通环境；
  完整reader、旧存储含link回归、文件reparse证据为管理员环境。
- 不宣称普通用户完成现有全部fixture/reader套件，不推断普遍需要管理员；
  文件链接命中真实路径防护，不单独证明所有安全句柄防护分支或所有reparse类型。
- PS5缺根、完整已填充Package独立oracle、最大重复Package组合等，按附录第5节
  已批准的披露边界保留，不宣称动态覆盖，也不默认为产品规范删除。
- PS5原boundary超时、普通reader fixture失败、PS5管理员reader严格空stderr失败
  仍是失败记录，具体原因未证实修复。PS5后续progress政策不追溯改写旧结果。
- 原计划/分组条件审查路径继续记为停止；新路径采用已经获准的替代证据，
  不勾选旧槽位或将新审查改名为旧审查。所有旧冻结/已耗预算保持不变。

## 4. 不随本结论放行的事项

不包含真实采集、命令实际运行/来源认证、签名/加密、敌对管理员与目录竞态、
事务/断电持久性、迁移、全项目测试、Stage B/I2/I3/full C12、供应商/安全处置、
发布或main合并。Verified仅为内部一致，文件标记不授予Acceptance/RuntimeApproved。

## 5. 用户决定与换机恢复

用户已明确批准第1节有限结论及第3/4节随附限制，详见独立批准记录。
批准只变更本模块该版本的有限离线验收状态，不自动启动任何下一项动作。
若拒绝或要求修改，保持未验收检查点，说明差异后再定范围，不能继续加测或改绿旧记录。

必要测试、证据报告、附录和恢复说明提交现有开发分支；实际fixture、raw receipt、
runner和会话保持ignored/local。新电脑可按Git恢复交付进度，不能冒称已重新核验原机器
原始记录或获得重跑权限。最终用户批准记录已另行保存，不以原提案代替批准记录。
