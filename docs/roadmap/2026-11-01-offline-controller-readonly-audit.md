# 冻结控制脚本：只读审查与候选修订范围

状态：READONLY_REVIEW_COMPLETE / PROPOSAL_PENDING_SELECTION / NOT_ACCEPTED。
基线：04948dc33b35a01205bed79a9b4dd9f416432c64。
用户本轮仅批准只读审查和制定最小修订建议，不是源码修复、测试或恢复运行批准。

## 已确认的原因

原始 stop.json 的 CannotIndex/调用栈和当前冻结源码一致：
driver.ps1:504 通过 ConvertFrom-Json（没有 AsHashtable）读取运行清单；
507 将 PSCustomObject 传给 Driver-FalseFlags；85 使用 Value[flag] 字典索引，
因此在 grant intent 写入和 Process 构造前失败。此前 exit1 是实际控制器缺陷，
不是测试通过、测试反例、平台限制或写函数本身的验证结果。

源码层面还存在同一边界的后续不兼容：driver.ps1:315 把该 Manifest 传给
New-ODRuntimeContext，而 source-model.ps1:99/64/33/3 的 Manifest 参数要求
Collections.IDictionary。该后续路径未运行；这里只读确认接口不一致，不声称
动态复现了第二个失败。只让 Driver-FalseFlags 接受对象，不能统一整个调用链。

因此算法层面的最小建议是：在 driver 的 RunOne 清单解码边界统一使用有序字典，
保留现有布尔精确类型、false 标记、来源/字节/生命周期及最终输出检查；不通过
放宽类型、跳过检查或 catch 后继续来绕过失败。Prepare 的对象式解码有其现有
字段更新用途，不能未核对就批量改全部 ConvertFrom-Json。

## 为什么不能直接改一行再运行

driver 本身是31个冻结来源之一。即使只改解码一行，原 writer-green 清单中的
driver 哈希也不再匹配，后续真实 profile 在 intent 前还会验证整份清单。
运行旧清单、去掉 driver pin、覆盖旧哈希，都会破坏来源合同。

此外旧清单名在多个消费者中固定：

| 位置 | 目前绑定 | 新纪元需要保持的关系 |
| --- | --- | --- |
| driver.ps1:404–410、463–464、502 | writer-green.frozen.json | writer 前置核验、writer final 和后续 repair 的 parent 一致 |
| record-boundary.tests.ps1:136 | writer-green.frozen.json | IntentFailureNoWork 使用当前实际 writer 清单，不能偷用旧清单 |
| source-model.ps1:110–112 | writer-green.frozen.json | intent 单元上下文的 prior SHA 必须等于实际 writer 纪元 |

repair-green.frozen.json 尚未创建，其后续固定入口可以保留。旧 seed 和
writer-green 两份清单、62份快照、两次 RED、失败工具结果、stop.json 都应原字节
原位置保留；不重命名失败、不把旧 writer 冻结改称新成功冻结。

## 候选路径与推荐

推荐一个显式、追加式的控制器修订纪元。另建固定名称的 writer 修订清单，
它记录旧 writer SHA、原 stop SHA、用户修订批准和精确改动来源；旧文件不覆盖。
沿用 writer-green grant/Phase 及其尚未出现的输出路径，不增加另一套测试语义。

来源候选窗口限三个现有 leaf 文件：

- driver.ps1：统一 RunOne 清单解析；新增唯一固定来源修订及严格恢复核验，
  将 writer/parent/final 绑定到新清单。原始 stop 仅在哈希精确、批准匹配且
  writer intent/输出/capture/unit 均不存在时作为已知前置失败被显式识别；
  新停线记录使用独立固定路径并立即阻止所有后续阶段。无通用 IgnoreStop、
  Force、任意清单/路径、回调或宽泛异常忽略参数。
- record-boundary.tests.ps1：只把 intent 控制的固定清单路径绑定到新 writer 纪元；
  不改任何组名、期望值、错误合同、真实引用断言、工厂行为或案例数量。
- source-model.ps1：只同步 intent 单元上下文的 writer 清单路径和 prior 绑定；
  不改变正常/历史上下文隔离、源码 AST 判定或完整 I2 合同。

这仍然是冻结测试来源的明确变更，不能称作“测试完全未改”。旧 RED 仅保留其
原始缺 API 结论，不重新执行，不追认它测试过本次新绑定。未来修订计划必须
列出逐行差异及绑定回归，得到明确批准后才能执行。其余28个当前来源不改；
后来原 Task3 的 observation-contract.ps1 独立窗口仍需保持原批准边界。

另一个可行方向是新建完全独立的下一版本并重新证明整条链，但这需要新的
RED/宿主/编译预算及更多来源副本，不符合本次最小修订和保留原运行额度的目标，
因此不推荐。直接覆盖旧清单、删除 stop 或仅放宽 helper 不是可接受替代方案。

## 尚需批准的计划内容与成本

这是候选范围，不是已批准的可运行脚本或完成的实施计划。选择推荐方向后，
再制定一个短计划，精确固定新清单/恢复记录路径、字段、创建顺序、哈希及单次
恢复条件；不在实施时临时猜测。计划需自查 writer→repair→intent→final 的类型
和 SHA 传递，并覆盖拒绝缺少/不匹配批准、原 stop 变化和已有尝试输出的情形。

拟议的增量成本须显式申请，不能隐含在原预算内：

- 增加一轮31份来源快照；现有62加修订31加原定 repair31，共124，而非原93。
- 拟增加两次现有工具上下文中的有限回归计算（一次 RED、一次 GREEN），不启动
  capture/目标进程、不编译夹具；用独立固定回归证明清单类型、false 布尔精确性
  和恢复拒绝条件。具体用例与期望在执行计划中先固定。本轮未授权也未执行。
- 原剩余5capture宿主/7目标/1夹具编译/条件1运行审查不增加、不退款；仍按原顺序，
  全部严格条件满足后才可使用。宿主启动前失败不等于可以自动重试。

不安装依赖、不改产品/CI/Docker/main、不修改全局错误去重或进程安全边界。
任何新来源、绑定、记录、回归或运行异常仍按批准的停线条件处理；不自动修复。
这是作者的有限源码审查，不是条件独立运行审查，也不是全链安全审查或验收。

## 本轮只读核验与恢复状态

读取完整 driver 和 source-model、相关调用/固定路径及原始 stop；没有导入运行
模块、调用其函数或做复现探针。只读比较2209个 writer pins 和31份 writer
快照，全部一致；intent/final/stdout/stderr/evidence/capture/unit/repair 均未出现。
本轮源码变更0，测试0，capture宿主0，目标0，编译0，独立运行审查0。

核验身份（SHA256）：

- driver：86978164BB4338AA7A9C27DBB7642FE055EE49DA3833DEC4E45C82D6FB4B14AC
- source-model：192AA46AC3DDDCE20F6AF9639EB79FC8D86C0AEA7D7DFAD2F68EE1A3433A30D9
- record-boundary.tests：346E6A21CA53521506196683F641D333D39AF0C6FB47FA485538528A553646EE
- stop：F19687B1628DC1C48CC771A07DA4308D48E89B8D51921DDACDB6C4CBFAFDE9D2
- seed：4E4342C389C507F6148AE029140814FBCB4046AD62D1EFE3BFD451863C9DC1BA
- writer：9AC3EAD775DB2BE7D21DE3B1CD3FE46342B1B176BA6B07C03F082D0F63B0C548

原185文件/49目录 portable-writer-prelaunch-stopped.json 保持不变，含实际失败、
工具记录和冻结来源。本轮仅新增本报告及状态/恢复文档，不制造新的运行证据。
项目继续 NOT_ACCEPTED；两次 RED 已消耗、62份快照已创建，所有剩余运行额度冻结。
