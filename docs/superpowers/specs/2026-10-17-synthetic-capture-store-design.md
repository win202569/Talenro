# 完整合成采集记录离线归档 — 正式设计草案

状态：**书面设计及后续计划已批准；实现检查点未验收，PS5边界验证超时后停止**。
实际结果见[执行报告](../../roadmap/2026-10-17-synthetic-capture-store.md)：已消费6/12，
7..12冻结，依赖回归与独立审查未启动；未修改测试或重试来绕过失败。
执行批准：用户在计划检查点 `aeb6fe363ac50b60e929f0123f762a286849a128` 后选1，
采用当前会话实现、最多12次新验证主机及一次独立只读审查；无修复/重跑预算。
批准发生于设计检查点 `b33e04bd0597d748ee5544ca563ca7654756240a` 之后。
基线：`e57db03bb7efa371aa091ae83e1619b9d2be697b`。
用户已同意独立、带版本的完整记录格式，授权将其写成正式设计。
本次仅写设计及恢复文档。书面设计批准后才能编写实施计划；计划批准、
执行方式和新的有限验证范围确定后才能实现。所有旧额度、冻结槽位及失败记录不变。

## 1. 目标、现状与非目标

目标：在另一台准备好依赖的 Windows 电脑上，从复制的离线归档恢复完整的
合成采集记录，原样保留 Unicode 声明、参数/流字节、事件顺序和失败事实，
重新验证其内部一致性，而非证明某条命令实际执行。

现有 `output-evidence-store.ps1` 的 TLROEV01/TLROEC01 只持久化输出包。
完整记录有20个顶层字段，见 `capture-adapter-contract.ps1` 的
New-ContractValidatedRecord/Test-SyntheticCaptureRecord。Lifecycle.Package
只在模拟生命周期 complete 时存在，不能靠旧存储覆盖 incomplete/timed-out/refused。

非目标：真实采集、进程启动/附加、OS时钟证明、watchdog、停止/清理进程、
安全设置或供应商处置、Stage B/I2/I3/full C12验收、main合并；不保存任意对象、
执行回调、迁移旧格式、自动修复损坏归档或承诺断电持久性。哈希不提供签名、
来源认证、加密、敌对管理员防护或协调篡改防护。

## 2. 选择与模块边界

采用独立固定布局二进制格式，旧存储及其读写API保持不变。
不采用旧输出包旁的附属记录：它需要跨文件协调，且没有Package的记录无法自然表达。
不采用通用JSON/CLIXML对象序列化：避免重复字段、隐式类型、Int64精度、
任意嵌套/反序列化行为成为新接口。代价是需要显式编解码和独立格式用例。

新增 `docs/diagnostics/isolated-fuzz/synthetic-capture-store.ps1` 及对应
`.tests.ps1`；此处是拟定文件名，不表示文件已实现。
新模块只单向导入既有 `capture-adapter-contract.ps1` 与 `isolation.ps1`，
使用公开 New-SyntheticCaptureRecord/Test-SyntheticCaptureRecord 和文件路径/
句柄检查，不依赖旧store的私有函数；不让合同、模拟器或旧store反向导入新模块。
新模块本地辅助函数使用独立前缀，不改现有合同、分类器和输出包语义。

公开接口：

- `Write-OfflineSyntheticCaptureStore -Directory <string> -Record <object>`。
  仅接受现有 Test-SyntheticCaptureRecord 返回 Verified=true 的完整20字段记录。
- `Read-OfflineSyntheticCaptureStore -Directory <string>`。
  返回精确7字段：Schema=1,Verified,Reason,Record,OriginVerified=false,
  Acceptance=false,RuntimeApproved=false。失败时 Verified=false,Record=null；
  成功时 Verified=true,Reason='consistent-synthetic-archive-only',Record为新建记录。
  失败Reason仅取固定代码：invalid-path,missing-archive,incomplete-archive,
  unsupported-format,invalid-payload,integrity-mismatch,invalid-record,io-error。
  不直接返回异常Message、完整路径、输入对象或输入内容；若多项条件同时失败，
  返回读取流程遇到的第一项，不承诺损坏文件的唯一故障归因。
- 写入前置失败抛出异常且不创建目标目录；写入中失败抛出异常并保留现场。
  写入成功返回与最终读取相同的7字段结果；读取验证失败不泄露半成品Record。

Directory 的字符串参数绑定不是敌对转换沙箱。Record在整个调用期间不得被并发
修改；不支持直接敌对调用内部辅助函数或进程级类型/方法替换。

## 3. 必须恢复的字段与含义

|组|顶层字段|恢复要求|
|---|---|---|
|清单，9项|Schema,Policy,SourceMode,SessionLabel,ExecutableDeclaration,EncodingName,OriginDeclaration,StartedUTC,EndedUTC|按现有合同精确匹配；文本不规范化，UTC仍是声明|
|原始输入，2项|Arguments,Events|恢复原字节、规范Base64/长度/hash；全量事件含晚到/被模拟器拒绝的结构合法事件，保持顺序|
|派生事实，4项|UtcOrderConsistent,Lifecycle,StopIntent,CleanupStatus|保存并重算比对；不将意图当作实际操作，CleanupStatus仍unknown|
|绑定，1项|BindingSHA256|保留原始摘要并与现有合同重算结果精确比较|
|行动/权限，4项|LiveActionsPerformed,OriginVerified,Acceptance,RuntimeApproved|全部必须是false；既不提升也不从文件标记推断|

Lifecycle的17字段及可选Package的13字段全部恢复；精确字段集合保持现有定义。
byte快照仍表现为Base64/ByteLength/SHA256三个字段。盘上只存快照原字节，读取后
重新生成规范快照：对已经验证的输入，其Base64、长度和SHA256值与原快照一致。
不保存PS对象装饰、方法或Int32/Int64等价值的装箱差别；完整恢复指合同字段值、
字节和顺序一致，不指对象身份或ETS装饰一致。记录不因归档而新增字段。

允许 complete/incomplete/timed-out/refused 的**内部一致**记录，包含非零退出码、
拒绝分类、缺失EOF/exit、保留前缀和未知清理状态。文件发布 complete 标记只表示
归档结构完整，不表示Lifecycle.Status=complete或运行成功。
既有校验拒绝的时钟倒序记录、7字段Created=false描述符及非法/超限输入不进入此
格式；不宣称保全所有未被合同接纳的数据。不得借此将真实运行重标为synthetic。

## 4. 类型、边界与内存输入

按固定字段布局检查plain对象/精确object[]和原始CLR标量类型，再读字段值。
检查属性描述符时不执行Getter/GetType/ToString；拒绝ScriptProperty/CodeProperty、
非NoteProperty对象字段及非CLR Length/Rank覆盖。精确标记用Ordinal比较，不使用
文化等价、大小写折叠或字符串强制转换。合法、未调用的GetType ScriptMethod不需
作为数据保存，不能被派发。所有层级必须有本地边界检查，不能只信任一次上游校验。

继续现有上限：SessionLabel ASCII1..80；ExecutableDeclaration/OriginDeclaration
各1..1024字符且严格UTF8<=4096字节；UTC精确28字符；encoding仅utf-8/utf-16le；
参数<=65536原字节，严格UTF16LE、偶数长度、无前导BOM；事件<=4096，chunk<=65536，
全量chunk之和<=2097152；模拟器 admitted stdout/stderr 各<=1048576。
事件ElapsedMs为0..Int64.MaxValue，exit为Int32；不放宽120000ms虚拟deadline。

新磁盘payload上限8388608字节（8MiB），含头；先检查文件总长再分配/读取。
该界限覆盖全量trace2MiB、Lifecycle快照2MiB、可选Package重复快照2MiB、
参数64KiB、最多4096项事件标量及有界清单，不提升任何单字段限制。
读取每个长度/计数前检查剩余字节，用有溢出检查的Int64做累计；负值、超上限、
未知枚举、非法标量、非法UTF8、截断或尾随字节均拒绝，不能靠总长合规跳过细项。

## 5. 固定线格式 v1

所有整数小端；Bool是精确一个字节0或1，其他值拒绝，不直接接受宽松ReadBoolean。
可空Int32用Bool存在位，0无后续数据，1后跟Int32。对象及数组无任意字段名或类型名。
Text是Int32 UTF8字节长度后跟严格UTF8；限额同时检查字节与合同字符数。
Snapshot是Int32原字节长度后跟原字节；读取后产生规范Base64/ByteLength/SHA256。
Digest是32个SHA256原始字节，恢复为64位大写十六进制，无文本替代写法。

单独目录下精确使用 `record.bin`、`complete.bin` 和可选失败哨兵 `stopped`。
只读本规范点名的文件，不递归加载/执行其他内容；存在额外文件不作为记录字段读取。
payload为ASCII8 `TLRSYN01`，Int32存储版本1，然后以下顺序的完整Record：

1. Schema:Int32；Policy:Text<=64字节；SourceMode:Text<=16字节；
   SessionLabel:Text<=80字节；ExecutableDeclaration:Text<=4096字节。
2. Arguments:Snapshot<=65536；EncodingName:Text<=8字节；
   OriginDeclaration:Text<=4096字节；StartedUTC/EndedUTC:各Text精确28字节；
   UtcOrderConsistent:Bool。
3. Events:Int32计数，然后有序事件。kind Byte枚举1chunk/2eof/3exit/4tick；
   ElapsedMs:Int64；chunk后跟channel Byte枚举1stdout/2stderr及Snapshot<=65536；
   eof后跟channel；exit后跟Int32；tick无附加数据。累计所有chunk原字节。
4. Lifecycle，按下一段固定顺序。
5. StopIntent:Text<=64字节；CleanupStatus:Text<=32字节；LiveActionsPerformed:Bool；
   BindingSHA256:Digest；OriginVerified/Acceptance/RuntimeApproved:各Bool。

Lifecycle顺序：Schema:Int32,Policy:Text<=64,Status:Text<=16,Reason:Text<=64,
AcceptedEventCount:Int32,FailureEventIndex:可空Int32,LastObservedElapsedMs:Int64,
StdoutEOF:Bool,StderrEOF:Bool,ExitObserved:Bool,ExitCode:可空Int32,
Stdout:Snapshot<=1048576,Stderr:Snapshot<=1048576,Package:Bool存在位及存在时的
固定Package布局,OriginVerified:Bool,Acceptance:Bool,RuntimeApproved:Bool。
计数/索引与当前事件总数一致性和所有派生值由重算验证，不仅做整数范围校验。

Package顺序：Schema:Int32,Policy:Text<=64,EncodingName:Text<=8,ExitCode:Int32,
OriginDeclaration:Text<=4096,OriginVerified:Bool,Stdout:Snapshot<=1048576,
Stderr:Snapshot<=1048576,Classification:Text<=16,ProgressCount:Int32,
Acceptance:Bool,RuntimeApproved:Bool。所有Text标记仍须符合现有合同的精确取值，
Text字节限额不是允许任意文本通过的枚举检查。

complete.bin精确52字节：ASCII8 `TLRSYC01`，Int32存储版本1，Int64 payload长度，
32字节SHA256(record.bin全部字节)。长度须为非负且不超过8MiB，并与打开的payload
一致。未识别版本拒绝，不尝试旧格式、JSON或通用反序列化回退。
TLRCAP01仅是现有记录绑定哈希的域分离前缀，不是完整存储格式；不得混用三种magic。

## 6. 校验及重建

写入先检查本地被动对象形状/标量边界，再调用现有公开记录校验；拒绝任意不一致值。
在创建目录前将已校验数据编码为自有有界payload，后续写盘不再读调用方对象。
此操作不宣称抵御并发输入修改；任何发现的前后不一致均拒绝，不裁剪或修正输入。

读入先取得文件安全绑定和总长，通过完成标记hash检查后按固定布局解码。
从原字节生成快照，但保留盘上的BindingSHA256、Lifecycle、意图和权限字段。
调用现有 Test-SyntheticCaptureRecord 检查重建的完整记录：它从清单/事件重算绑定
和Lifecycle/StopIntent等，并比较已存派生事实。不能直接用重建结果覆盖磁盘上的
派生事实来“修复”篡改。仅当完整校验通过才公开Record；返回Record不得共享可变
调用方/解析缓存，包括空事件数组。

payload与其完成hash可以被协调修改；绑定hash也可以被伪造一致。因此Verified仍仅是
内部一致，不是不可篡改、来源证明或真实执行认证。将来合同或模拟语义改变时，旧归档
若不能按v1验证就拒绝，不能静默升级；兼容解码/迁移需另行版本设计和批准。

## 7. 本地文件边界与发布状态

独立根为仓库的 `.superpowers/diagnostics/offline-synthetic-capture-store`，保持ignored。
读操作不创建根；写操作仅在Record校验/编码成功且父路径无reparse后创建缺失的
专用根，再重新检查。前置校验失败连专用根也不创建。
目标仅允许其一个直接子目录，leaf为ASCII字母数字开头，后接字母数字/下划线/连字符，
长度1..80。拒绝UNC、相对路径、ADS、目录穿越、尾点/空格及不规范路径。
Windows路径比较OrdinalIgnoreCase；协议/文本标记比较Ordinal，不能混为一谈。

输出文件只用CreateNew、独占写，已有目录/文件一律拒绝，不覆盖/复用。
所有路径组件拒绝reparse；打开文件后按同一SafeFileHandle检查非reparse且硬链接数1。
读持有payload/marker句柄，FileShare.Read阻止写/删；验证前后都检查stopped存在性。
这些措施不提供敌对并发改目录、管理员或文件系统层竞态的完备防护。

写入顺序：校验/编码 -> 创建新目录 -> 独占写record.bin并Flush/关闭 ->
以安全句柄重读payload并核对长度/hash -> 最后CreateNew写complete.bin并关闭 ->
调用公开读取进行完整最终验证 -> 返回成功。文件内容完成与Lifecycle完成相互独立。

创建目录后的错误尝试CreateNew写stopped，随后抛出原错误；stopped写入失败不能
覆盖原错误，也不能把错误转换为成功。保留所有文件，不自动删目录或重试。
缺payload、缺/部分marker、stopped存在、hash/语义不符均返回失败和Record=null。
进程崩溃可能来不及写stopped；只按现存文件和完整校验判定，不虚构故障阶段。
如果文件已全部发布且确实通过校验，后续读取可通过，不代表先前写调用已返回成功。
不承诺事务原子性、fsync或断电落盘；Flush/关闭/重读不等价于硬件持久性。

## 8. 换机恢复与敏感信息

Git保存必要源码、固定合成测试数据、设计、计划和恢复文档；实际归档、测试输出、
缓存和运行现场默认不提交。复制归档需另获数据传输授权，保留整个目录及stopped，
不能只挑complete.bin或删除失败标记。归档不加密，参数/文本可能敏感；仅使用批准
的合成非敏感fixture，不将真实命令或凭据导入测试。
新电脑准备PowerShell及项目依赖后，在新专用根下放置归档的完整副本，由新读取API
只读验证；文件中的exe/参数/路径/文本一律不执行，也不用于定位外部数据。
复制不会恢复旧机器的会话/缓存、自动续跑进程或扩大任何旧验证额度。
旧输出包仍由旧API读取，不自动转换为缺少历史清单/trace的“完整记录”。

## 9. 未来验收要求（本设计不授权执行）

后续实施计划须规定新的有限主机次数、120秒单次上限、失败停止/保留规则、
独立审查边界与是否存在修复预算；不得借用任何旧/冻结槽位。本轮没有测试调用。
需要基于真实公开接口和本地合成文件验证，而不是grep源码或mock校验器：

- 独立手编码二进制/marker oracle、全部20字段往返及完整四种生命周期；
  Package=null、exit缺失、非零exit/拒绝分类、晚到或被拒绝trace字节仍保留。
- 参数/Unicode/UTF8与UTF16声明/交错顺序/空数组独立所有权/全部精确界限；
  metadata、快照、绑定、派生状态和权限篡改拒绝；getter/方法不调用。
- 非法UTF8、长度/计数负值及溢出、超过总长/字段/trace上限、未知magic/版本/
  枚举、非0/1布尔和存在位、截断、尾随字节及完整hash但语义错误的payload拒绝。
- 缺文件/坏marker/stopped、已有目标不覆盖、目录/文件reparse或多硬链接、
  独占写句柄冲突拒绝。受控中断只构造有界文件状态，不做真实断电/进程采集。
- 旧格式继续由旧API读取，新旧交叉读取拒绝；当前依赖链相关回归与有限主机
  耗时单独记录，不将文件测试通过当作真实运行、安全放行或main合并依据。

## 10. 自审与下一门槛

自审：20字段组数9+2+4+1+4=20；Lifecycle17/Package13与当前源码一致。
完成marker8+4+8+32=52；8MiB总长保留细项限制；原字节快照恢复规范值，不归档
可执行对象；输入一致性、归档完整、生命周期完成和真实性是四个不同判断。
旧格式与TLRCAP01不复用；未验证时钟倒序记录不被悄悄放行；无自动迁移/修复/
断电保证。未写产品代码、未新增测试执行、未获得实施权限。

用户已审阅批准本书面设计。此批准仅进入编写实施计划；实施须另行审阅计划、选择执行方式
并批准有限验证范围。设计保存到现有开发分支，不合并main或删除任何旧工作现场。
