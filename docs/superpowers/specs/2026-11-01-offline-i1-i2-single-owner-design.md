# I1/I2 离线运行实现修复：单一收集归属设计

状态：DESIGN_PENDING_USER_REVIEW。仅设计，不是代码、测试或运行批准。
依据：用户已选择“统一错误收集归属、保留原始首因及现有去重规则”。
本稿提出跨层合同的明确修订，仍须审阅批准。基线为正式数据评估检查点
68e9b2b45203495d63f4b44d4cc1364416796c3c。

## 1. 目的与证据边界

“运行实现/生产写边界”仅指离线诊断实际调用的实现，不接入产品、CI、Docker
或默认验证链。修复 I1 首因顺序与重复收集、I2 完整观察判定；不关闭历史发现。

已知证据：旧 writer-red 在 IntentDisposeOnly 意外中断；后续简化诊断记录了
同一底层 Exception 经重抛形成不同 ErrorRecord 的现象。新的数据校验器通过
44组测试，并对原始六案例作出 DATA_CONTRACT_MATCH。这些结论不能恢复历史
中断案例的私有对象，也没有动态证明真实磁盘晚写/Dispose 复合失败。

旧写入源码、测试、oracle、manifest、raw、portable、stop 与审查保持不变。
本稿不是让旧控制器继续跑。v3 剩余6宿主/7目标/1编译/条件1审查继续冻结；
更早额度不变。Session/Ack 不序列化、不从磁盘恢复；五能力标记仍 false。

## 2. 选择及代价

采用用户选择的单一收集归属：每次记录操作只产生一个私有活对象结果，调用者
按原始发生顺序登记它；预期记录操作失败不再“内部登记后 throw、外层再登记”。

未采用按 Exception 引用合并 ErrorRecord：会改变全局去重语义，也不能仅凭六
案例推断所有未来独立错误。更不采用 Message/ErrorId/字段相同即合并。

保留 Capture-Failure 规则：同一个 ErrorRecord 引用只留一次；不同记录不因
相同 Exception/文字合并；已有 HostPrimary 不替换；次因最多8个，第9个只置
Overflow。代价是必须修改新版本的真实调用点和测试合同，不能只改单个 writer。

**明确合同变化：**旧测试要求外层 catch 收到的抛出 ErrorRecord 与原始首因是
同一个对象。本设计不作此承诺，也不以“同一 Exception”冒充同一 ErrorRecord。
新版本使用不经重抛的私有结果保留原始记录引用；旧断言/结果原字节保留，新测试
独立版本化。批准本设计意味着批准这项跨层传递变更，不是暗改旧测试使其变绿。

## 3. 私有记录结果与唯一所有者

新私有写函数仍只接收内部固定 Path/Bytes，无调用者工厂、回调或任意输出根。
正常入口使用真实 CreateNew 工厂；隔离测试工厂不能进入运行模块/最终记录器。

函数返回恰一个私有结果对象，字段有序：

| 字段 | 类型与规则 |
| --- | --- |
| Succeeded | bool；Open/Write/Flush/Dispose 全部成功才 true |
| PrimaryError | 实际捕获的 ErrorRecord 或 null；首个实际失败；不得构造替代记录 |
| SecondaryErrors | 明确 ErrorRecord[]；只含首因之后发生的清理错误；正常为空数组 |

这不是公开授权对象、磁盘报告或通用异常通道。仅同宿主直接传递，结果中的记录
不能克隆、序列化往返、转换成投影后再用于引用断言。无错误时不能输出其他值。
Succeeded=true 必须伴随 PrimaryError=null、SecondaryErrors=空数组；false 必须
有真实 PrimaryError。当前一次流操作最多一个 Dispose 次因，不接受任意错误列表。

写函数不调用 Capture-Failure、不发布成功台账、不 throw 已捕获的预期记录错误。
只在局部留存首次 Open/Write/Flush 异常；finally 仅 Dispose 已持有的确切流一次，
捕获其错误。若此前无首因，Dispose 错误成为唯一首因；否则成为清理次因。
Open 失败后不 Write/Flush/Dispose；Write 失败后不 Flush；任何失败都不重开/重试。
参数/结果构造或实现缺陷等未分类异常不是“可处理写失败”，继续走 unknown 停线。
捕获边界仅包围真实 Open/Write/Flush/Dispose 调用；不能用整个函数的大 catch
把结果构造或实现错误伪装为已处理的 I/O 失败。实际 I/O 失败不靠错误文字猜测分类。

操作结果只交付给一个拥有该边界的调用者，由它消费并调用 Capture-Failure：先
PrimaryError，再 SecondaryErrors。写函数本身没有 Host 状态。一次调用点只消费
一次，消费标志在登记前锁存；发现重复消费或 malformed 结果即停止，不补做 I/O。
这不是可从外部构造的运行权限；具体私有实现可由实施者选择最小显式对象结构。

### 3.1 Profile intent 的所有者

Invoke-ODProfile 在任何 intent I/O 前锁存调用尝试。intent 写结果由 profile
作用域独占消费，写失败不进入测试/profile 工作体、不建立 Ack、不增加目标启动。
收束已持有资源后返回 INCOMPLETE ProfileReport，包含实际首因/次因活引用；
不再让 intent 错误先 throw 穿过模块，再在捕获宿主补收一个 wrapper。

捕获宿主只在 ProfileReturn 边界导入该 profile 的错误序列一次；不是把整个
报告当异常 throw。profile 与 capture 是不同台账，但同一边界内各有唯一所有者。
正常失败报告必须被调用者显式判失败，绝不因调用正常返回就当作运行成功。

新 ProfileReport 统一 Schema=2，保留原报告有序根字段：Schema,Profile,EntryPath,
EntrySHA256,KernelPath,KernelSHA256,FrozenSHA256,Status,Groups,Observations,
PrimaryError,SecondaryErrors,Overflow,Closed 及原五能力标记。intent 失败时 Groups
和 Observations 为真实空数组；Closed 反映实际收束，不能固定为 true。此分支要
在所有元数据已获得后进入；元数据/摘要失败则保持 unknown，不伪造完整报告。

### 3.2 最终 capture 记录的所有者

publisher 私有模块只执行返回结果的写函数，不另留独立 HostPrimary。将结果活
引用直接交给 capture 所有者；由 capture 先登记首因再登记次因。记录失败后该
出口不可再发布；保留实际部分字节/工具错误/已有独立原始流，不寻找替代成功文件。

首因/摘要或报告转换本身再失败，作为新的真实次因有界保留；不能丢弃首因、
清空台账、构造成功报告或重试失败摘要。未知异常的外层 catch 仍收集并停止，
没有“此错误应该已处理”的宽泛忽略分支。没有全局 wrapper 匹配/抑制表。

## 4. I2 完整观察合同

保留同一纯入口形状 Assert-ODProfileObservations(Rows,Profile,Context)：成功无
输出；合同不匹配仅抛 RuntimeException，Message/FQID 均为
`OD: observation contract mismatch`。意外实现异常不冒充成功拒绝。

判定器对**完整**观察序列而非 Step 子集负责。现有正常 profile inventory 仍为
core22、owned8、live15；intent 失败是 Schema2 INCOMPLETE 的独立报告分支，
不得要求它具有正常完整观察数，也不得把空观察判为成功。三种正常报告中根字段、
Name/Value、View/Sinks、IORoot/Result、Error、summary 和 sequence 全部递归核对。

每层都明确有序键、真实数组/映射/空值及数值/布尔/字符串类型；拒绝缺失、重复、
未知、错序和隐式类型强转。固定错误 Type/ErrorId/Message 来自独立合同及真实
测试行为，不能从待评估输出“学习”期望。仅路径模板、PID、ElapsedMs 和明确
源位置允许受限变化；源位置由冻结源码 AST 独立推导，不能只查 Contains/任意行。

解析、源码模型读取、投影等分离在纯判定器之外；Context 是已验证输入数据，
不是 I/O 或执行许可。沿用已验证 BOM 精确比较经验，不使用文化相关比较判定
原始字节/标识；保留原始字节，不 Normalize/TrimEnd 多个终止符来掩盖坏记录。

同一新判定器必须实际被最终父消费者调用，覆盖全部 ProfileReturn 和 Schema2
失败分支。纯数据单元测试通过不能替代这条真实接线的有限集成证据。

## 5. 新旧归属与实施边界

作者建议新版本归属（尚未创建）：scripts/offline-diagnostics-v4/ 和
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/。旧 v3/历史 leaf 不改；
新版本不重定向旧命令、不复制/恢复旧 grant。精确文件清单及来源哈希由实施计划
冻结；不是在本稿中授权创建新版本、导入模块或恢复旧 v3 控制器。

新版本须同时接线 public entry、session/profiles、write-boundary、publisher、
capture、最终 parent/observation-contract 和隔离测试，不能只另造演示 helper。
保持固定三个 profile、默认 Inspect 无运行准备及未知请求提前拒绝；无任意脚本、
路径、重试、恢复或 Force 参数。Session/Ack/受管出口及确切 Process 原约束不变。

不动 api/cmd/internal、原 seam/codec/inbox、CI、Docker、默认验证脚本和 main。
必要依赖缺失须报告，不安装工具或放宽安全设置来换取通过。新机器恢复从本开发
分支/恢复文档开始；不得从 portable 恢复运行引用或把磁盘成功当授权。

## 6. 后续验收义务（本轮不执行）

实施计划必须先冻结新测试/独立 oracle，再列有限 RED→GREEN、允许改动窗口、
宿主/目标/编译/条件审查精确预算和真实输入完整性检查。此设计不申请任何数额，
也不复用旧冻结额度；任何实际失败/unknown 后停止，不自动增加轮次。

I1 至少验证两实际写边界各5向量：Write+Dispose、Flush+Dispose、DisposeOnly、
SuccessOnce、OpenFailureOnce。引用断言在活对象范围进行，PrimaryError 是首个
实际捕获记录，同一引用只登记一次；复合次因恰1，DisposeOnly 无次因；原有首因
不替换；正常字节一致；失败顺序计数/一次Dispose/现有文件不变均须核对。

另验证：intent 失败返回 INCOMPLETE 且工作体/目标/Ack0；publisher 结果回传后
只有 capture 登记；两条独立同文字错误保留两份；重复相同引用不增加次因；不同
记录即使共享 Exception 仍按原规则保留（仅作为显式合同控制，不能伪装实测）。
收集/投影/最终记录故障必须保留真正首因及 unknown；无重复消费或重开失败出口。

I2 保留历史3正例+20单变异作为资料基线，另加新 Schema2 的正常/intent 失败/
未知异常分支和严格数组/数值/完整字段变异。每个反例必须得到固定拒绝签名；
有效控制必须通过，防止“全部拒绝”却报告反例全绿。新控制先冻结，不改旧 oracle。

受控 Stream 故障夹具只在私有测试作用域替换工厂，不泄漏正常/最终记录器。
真实默认 CreateNew 成功/已存在文件检查另外保留。夹具真实异常不等于 OS 晚写
或真实 FileStream.Dispose 故障；未诱发情形明确写未动态验证。

真实入口验证和获准独立运行审查仍是后续窄验收条件。保留原有限退出/排空期限、
确切 Process 所有权，不增 PID/名称重查、杀树、Job/watchdog、无限等待或预热。
单元/数据评估/设计审阅均不等于这些集成证据；历史 C1/C2 保持 open。

## 7. 审阅与下一步

本轮只读源码/证据并写本稿；无代码、测试、Prepare、编译、目标或运行审查。
技能要求的一次只读文档可实施性核对单独记录，不替代条件独立运行审查或放行。
本稿待用户批准后，才编写具有精确文件、接口、测试和独立预算的实施计划。
设计批准本身仍不允许运行；旧冻结输入/失败不回填，I1/I2 尚未修复或关闭。

来源：完整独立运行审查、2026-11-01 I1/I2 continuation 停线报告、原离线入口
迁移设计、原 record-boundary/observation-contract 源码，以及最新正式数据评估。

文档可实施性核对：一次独立上下文只读设计检查，任务 i1_i2_design_builder_check，
只读本稿、无代码/测试/Git/网络/写入，返回“nothing important left”。作者另补明
成功/失败结果互斥关系、单个清理次因及真实操作的窄 catch 边界；未进行第二次核对。
这是方案梳理技能的文档检查，不是运行源码/证据的独立质量审查，也不消耗或释放
任何已冻结的条件运行审查额度。最终设计批准仍由用户决定。
