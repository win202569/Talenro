# 超时与失败记录修订：新验证预算申请

日期：2026-10-19。基础提交：e0710824da6f853ffb0a57a8a45b111bbd40d1ef。
状态：DIRECTION_APPROVED_BUDGET_PENDING；不是执行批准，也不是测试结果。

用户已批准聊天中三项受限设计：200ms不变；超时允许空流/固定文本的
字节前缀、正常退出仍严格全文；断言前收集观察；结果单次写入，失败停线。
本文件仅细化新运行授权范围，不另起产品架构或修改旧书面规范/计划。
本轮只读核对及文档保存，0实现/0测试/0目标/0编译/0新增代理或审查。

## 1. 已知事实及不能推断的内容

[旧执行报告](2026-10-19-p2b-live-completion-supervisor.md)记录：核心7组通过；
owned-red精确复现回执重试；owned-green仅3条PASS后exit2，不能验收。
timeout子进程14092实际退出-1、排空、约210ms，stdout/stderr均0bytes；
normal子进程23384正常退出0、两流完整、约376ms。旧capture只在全部测试
正常返回后保存详情，故断言中断时旧trace/details没有写出。

源码及原件支持“超时的流合同错误地要求必有完整输出”和“失败记录边界
晚于断言”两项修订。调度/启动延迟是解释候选，不是已测根因；旧运行丢失
的内存ErrorRecord及故障子进程PID不重建。新合同不把旧失败改为成功。

旧额度已用4宿主/4子进程/0审查，余下1宿主/4子进程/条件审查继续冻结。
新额度与其相互独立，不退款、不迁移Session/确认，不凭本文件自动启动。

## 2. 修订边界

### 字节合同

- stdout固定字节为ASCII `P2B-CONTAINMENT-OUT`，stderr为
  `P2B-CONTAINMENT-ERR`，各19bytes；独立比较，不假定两流同步。
- 仅固定timeout用例允许长度0至19的逐字节前缀，包括空数组及完整文本。
  非前缀、超长、null或非byte[]都拒绝；不得把空数组通过管道折叠成null。
- normal、sequence的EmitAndExit仍逐字节完整匹配；原seam全部输出/trace/
  数字及哈希合同不变。禁止统一放宽为Contains、任意空流或忽略stderr。
- timeout仍须实际Start/关联、TimedOut/TerminationRequested、exit-1、
  OwnedExit/StreamsComplete及无控制器错误。没有输出不证明fixture正文
  已执行；本向量只证明确切拥有对象的超时收束，正常向量另证正文输出。
- 纯字节判定助手只返回数据校验结果，不接受“观察成功”setter、不造
  Process/确认，也不替代私有生命周期门。时间/进程关联/失败出口门不放宽。

### 断言前观察与单次证据

每次真实Invoke/受管IO返回后，测试先收集实际返回、Session快照、实际
ErrorRecord引用及原始流路径，再断言；不能只在断言成功后记详情。
内存首因保留原引用；断言/后续失败不能覆盖它，次因最多8个，重复引用
去重，第9个overflow。预期负例仍归属各自被测Session，不污染正常宿主。

新capture使用每Grant唯一固定的evidence.json封套，取代新路径中的
trace/details二文件；旧文件保持原样。封套包括已完成组、当前失败组、
未运行组、实际观察及首/次因投影，不导出确认身份。最多32项观察、每项
至多32个出口、总UTF8序列化结果不超过1MiB；超限不截断成成功。

收集/断言异常也进入同一个最终保存边界：先停止会话并撤销确认，完成
允许的确切对象清理和输出，再至多一次序列化/CreateNew/Write/Flush/Dispose。
没有异常时也只用此边界，不能正常路径写完再在finally补写。保存/关闭失败
锁存错误、非零退出、停线，不重开/覆盖、不寻找替代台账；错误流仅一次
尽力摘要。已有封套、宿主死亡、文件IO阻塞或保存失败不保证有完整详情。

封套只有证据用途。即使失败写留下完整字节，也不能授予下一Grant/确认；
父运输器仍以实际持有进程的退出/排空、返回结果及完整冻结校验判断精确性。
封套不是成功事务、持久化权限或恶意同进程代码的安全边界。

## 3. 拟申请的完整新预算

推荐Native内联实施：无实现子代理、无逐项review。NEW6个PS7宿主+
7个受控PS7子进程=13新进程；仅全部精确后1次fresh只读整体审查。
已有普通父上下文负责Prepare/单次运输/文件核对，不另起PS7做这些工作。

| 顺序/Grant | 宿主 | 子进程 | 固定内容及精确预期 |
| --- | ---: | ---: | --- |
| prefix-red |1|0|生产调用的字节助手基线仍要求全文；空/适当前缀两组恰FAIL，其余5组PASS，exit1 |
| prefix-green |1|0|同一固定向量7组PASS；新模块核心回归7组PASS，共14PASS，exit0 |
| recording-red |1|0|真实缺文件读取错误后固定断言中断；内存先收集，基线未保存封套；精确预期exit2、封套缺失，不把其他exit2算RED |
| recording-green |1|0|相同只读错误/同一断言中断，exit2；唯一封套完整记录INCOMPLETE及真实首因、断言次因、未运行标记；这只是预定负例，不是正常成功 |
| owned-green |1|3|回执目录故障Hold、正常EmitAndExit、200ms超时Hold；另有预启动故障0Start，六组PASS/exit0 |
| sequence-green |1|4|A0正常、B0正常、C0正常后输出失败、A1原seam；四组及原15行/12PASS/exit0 |
| 合计 |6|7|条件性1审查；不加FINAL、性能轮、编译轮、故障重跑或第二审查 |

prefix七组固定为EmptyPrefix、ProperPrefix、CompletePrefix、InvalidPrefix、
NormalExact、NormalShortRejected、NullAndTypeRejected。两流各逐一覆盖长度
0…19，含错误字节、超长、null/错误类型；期望来自固定字面字节，不由
被测助手计算。纯字节用例不伪造OS观察；owned-green才验证真实超时对象。

recording两个Grant都只用固定零目标用例：在独立Session对未创建的固定
台账路径实际读取，保留实际错误，关闭Session，再触发指定断言中断。
其他组标NotRun。不是任意脚本块、callback、伪造ErrorRecord或未计划错误。
父入口须逐项核对固定用例/输出/exit2/文件存在性及GREEN封套的错误类型、
首因同引用的内存断言、0Start/0确认、组顺序和未运行状态。任何其他错误、
首因缺失、宽泛匹配“有异常就通过”均非精确，停线。

所有输出全文、根字段及顺序、完整向量和错误摘要模板在首次RED之前冻结。
动态PID/耗时/原始字节另存观察，按本合同逐字节判定，不从实测回填oracle。
recording-red的预定缺失只属于新负例，不能补造旧owned-green丢失的记录。

## 4. 输入隔离、变更窗口和等待

新leaf拟为docs/diagnostics/isolated-fuzz/p2b-timeout-observation-repair/；
新packet拟为.superpowers/sdd/2026-10-19-p2b-timeout-observation-repair/。
本轮不创建这两个运行目录。旧leaf13文件、旧packet96原件及旧报告只读。

新14输入：原12文件职责保持（.gitattributes、README.md、expected-output.json、
session.psm1、sinks.ps1、owned-step.ps1、sequence.ps1、core.tests.ps1、
owned.tests.ps1、sequence.tests.ps1、capture.ps1、driver.ps1），另加固定
prefix.tests.ps1和recording.tests.ps1。所有测试/字面oracle先写，全输入LF/
静态解析后才冻结/运行；正常收束代码沿用已修版本，不恢复回执重试。

首次冻结前的新路径/Grant路由、观察先收集、单封套运输接线是明示的首次
装配边界。它不是每条异常路径均有独立RED证明；recording-red/green覆盖
指定断言中断，不冒称覆盖死亡/所有flush/close失败。独立审查须专查此边界。

继承旧owned-green清单1670映射，追加旧packet96、旧leaf13、旧报告1、
本申请批准快照1、新输入14，拟共1795映射。同路径同哈希历史角色可保留，
同路径冲突拒绝；原件缺失立即停线，不从portable重建。当前只读数量核对
为96/13；正式Prepare仍须逐项核SHA256和角色，不以计数代替完整校验。

只允许两次事先声明的修复窗口：

1. 精确prefix-red后，只修改sequence.ps1的字节判定/实际消费分支。
2. 精确recording-red后，只修改capture.ps1的失败最终保存边界。

其他输入、测试、oracle、驱动始终不变。每次新宿主对应独立批准序列，
不热换活Session；前后清单/源码快照分开留存。若需改其他文件即停止，
不能临时扩大窗口。新文档/便携输出不作执行输入或继续权限。

外层宿主非sequence最长120000ms、sequence150000ms；内部整轮原上限
140000ms不增加，预留退出/双流4000ms，等待取各步上限与剩余预算较小值。
退出2000ms、双流共用2000ms；Hold45000ms、故障/正常2000ms、超时200ms、
seam110000ms均不加长。不增加ready协议或为等输出另开等待窗口。
只处置仍持有的确切Process，不按PID/名称重找、不杀树、不查询无关进程。
文件IO及宿主失联不具硬实时保证；外层杀宿主不证明所有后代已退出。

## 5. 停线、审查与交接

任何非精确RED/GREEN、未知、漂移、未退出/未排空、未计划异常或记录失败，
在有限清理后冻结全部剩余新Grant和审查。只有上述两项精确RED容许各一次
对应修复；没有自动再修/退款/重跑/替代目标/二审，不删除raw。

全部6宿主/7子进程及封套/输出/清单精确后，才消费唯一fresh整体只读审查：
独立上下文、最强可用模型，读取全部新输入/原始证据/两次快照及旧合同；
无执行/导入/网络/改写/其他代理。重点检查前缀仅用于timeout、空数组身份、
断言前收集、保存失败首因/无重试、确认不早于全部完成动作、父运输不信磁盘。
Critical/Important或审查不完整均停线，不自动修复/再测/二审。

必要源码/测试/报告/状态/恢复文档和便携证据提交推送现有开发分支，并核
远端SHA相等；便携JSON保留原始数字文本和流/快照字节，同时核对Git的LF
容器与内嵌payload。缺失和未运行如实标注，数据不恢复Session/确认/权限。

不安装、不改安全设置、不合并main、不发布；无真实/正式/full-C12验收。
SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/
RealGateImplemented仍false，C1/C2未关闭，R2/Minor2继续延期，旧证据/额度不变。

## 6. 文档自检与待批准项

作者已核对：6+7=13；两项RED各独立宿主/零目标；实际孩子3+4=7；记录
负例exit2与正常成功明确区分；原合同与新字节合同不改判混用；只有两个
单文件变更窗口；旧冻结额度不解冻；恢复资料不恢复权限。本稿非独立审查。

请批准：按已同意的受限设计，Native实施本范围，使用NEW6宿主+7子进程
及条件1次只读审查。批准前不得实现/Prepare/启动目标/运行测试或派审查。
