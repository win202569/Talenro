# Main-state selected issuer/serial local-reference binding

日期：2026-10-13。状态：三部分会话设计已确认；本文等待书面审阅。
分类：架构增量。本文只定义未来合同，不表示实现、测试或验收已经完成。
基线：649cc0891011334dd032bcc4a9618110b0a9c5ce，
分支 codex/c12-b01-task9-coordinator，工作区 .worktrees/c12-b01-task9-coordinator。

## 1. 用户意图、事实与范围

用户选择：定义新的、明确版本化的本地 issuer/serial 映射，先设计，
再批准实施计划；不猜测旧算法，不直接改生产存储或运行授权链。
核心规则、组件/数据流、测试/验收三部分均已分别获得会话确认。
会话确认仅允许撰写本文；本文及之后的实施计划仍须各自审阅批准。

现有 nodeleafbinding 通过一次真实 nodecerttime 判定，已把 Proposed
主状态的节点、exact leaf DER 摘要、exact leaf SPKI DER 摘要及完整
证书窗口对应起来；issuer/serial 未取得该层的材料对应结论。
既有所有 Pending/Obligations 和整体阻塞继续保留。

只读核对的事实：

- statecontract/input.go：Identity.IssuerID/Serial 是 contracts.Digest，
  后者为固定 [32]byte；原 shape/transition 只检查结构及声明间关系。
- serving/types.go：CertificateLookupInput.IssuerID 是字符串，
  SerialBytes 是 1–20 bytes 的非零首字节序列，形状通过不等于授权。
- db/migrations/00006_nodecontrol.sql：issuer_id 是 1–128 bytes、
  [A-Za-z0-9:_-] 格式的 text，serial_bytes 是 1–20 bytes 的 bytea。
  数据库的长度约束本身不能证明序列号来自已验证证书。
- db/queries/nodecontrol_serving.sql 的 ListExactCertificateProjection
  用字符串和原始字节等值筛选；它没有提供本次主状态摘要的转换合同。
- nodeissuer/nodeca/nodeleaf 提供实际解析、关联和签名证据；
  leaf 的 Serial 是左对齐 [20]byte，加 SerialLength，
  CA 的 CADERDigest 对应 exact CA DER 的 SHA-256。

在这些已检查路径中，未找到把存储 issuer 字符串/序列号字节变为
主状态两个摘要字段的既有合同。这个有限搜索结论不是“全系统绝无
任何其他规则”的证明。本设计因此明确选择新规则，而非兼容性猜测。

唯一未来新产品目录：internal/nodecontrol/nodeissuerbinding。
本增量只把 Proposed.Identity.IssuerID/Serial 与真实下层材料按§3
精确对应，并保留既有叶证书投影。成功不是完整 identity、数据库
row、唯一性、来源、私钥、历史、提交、Ready 或 runtime 授权证明。

不绑定 Previous 的证书材料；不迁移旧摘要；不接受生产 issuer 字符串、
provider/registry/数据库句柄、私钥、额外 raw、替代真实材料的预计算摘要、
verified flag、用户提供的 Summary、回调或替换下层判定器。

## 2. 选定方法与代价

比较过三种方向：

1. 新本地映射合同（已选择）：在真实叶证书绑定之上增加一个纯组件，
   编码和作用域明确，便于离线验证。代价：生产兼容性仍 UNKNOWN，
   需未来单独建立与部署 issuer 字符串/登记记录之间的可信映射。
2. 先接生产 issuer 登记/provider/store：能探查生产来源，但扩大到
   尚未验收的运行边界，不属于本次有限增量。
3. 仅写核对报告：不引入能力，但也不完成新的材料引用对应。

新合同标识固定为 main-selected-issuer-serial.ca-der-bound.v1。
其语义由新组件的 Evaluate/API 和返回 Profile 明确指定，不通过
“字段看起来像某类摘要”自动识别，也不提供旧格式 fallback。
Input 不增加格式选择器；调用此组件即选择该固定合同。

issuer 在本合同中表示 exact 签发 CA 证书的本地引用，不是部署 issuer
字符串、CA 公钥身份、Subject Name、SKI/AKI 或父 CA 身份。
CA 续签即使公钥和名字不变，exact DER 变化也改变引用的 preimage。
因此 serial 引用也依赖这张 exact CA 证书，不跨续签证书沿用。
这些是明确设计代价，不是现有数据库唯一键的替代物。

本层增加完整有界复制、真实下层判定、两个固定用途的 SHA-256 派生。
不优化成摘要拼接后绕过下层，不扩展为通用 hashing/parser API。

## 3. exact 编码与投影对应

定义 H 为 SHA-256，输出原始 32 bytes；“||”为原始字节拼接，
没有十六进制、base64、JSON、DER INTEGER 编码或字符串数字转换。

两个域常量为 ASCII，均恰 36 bytes；下述 \x00 表示一个末尾 NUL byte，
不是四个可见字符。点、连字符、大小写、版本和 NUL 都必须逐字相同：

~~~text
DI = "talenro.main-selected-issuer-ref.v1\x00"
DS = "talenro.main-selected-serial-ref.v1\x00"
~~~

设 E 为此次真实 nodeleafbinding.Evaluate 返回的完整 Summary：

~~~text
C = E.Evidence.Certificate.CA.CADERDigest                 // 原始32 bytes
S = E.Evidence.Certificate.Leaf.Leaf.Serial               // 左对齐20-byte数组
L = E.Evidence.Certificate.Leaf.Leaf.SerialLength         // uint8
I = H(DI || C)                                           // 68-byte preimage
R = H(DS || I || U8(L) || S[0:L])                         // 69+L bytes
~~~

U8(L) 是恰一个原始 byte，不是 decimal、varint、uint32 或长度前缀字符串。
Serial preimage 长度范围 70–89 bytes；只 hash 实际长度，不 hash 20-byte
数组未使用尾部或 89-byte scratch 的全部容量。Issuer preimage 固定 68 bytes。
固定 [68]byte 和 [89]byte scratch 在自有清理登记后填充，用后清零；
不通过动态 append 扩容，不使用 hash.Hash/流式接口/通用编码器。

派生前要求 C 非零，1 <= L <= 20，S[0] 非零，S[L:20] 全零。
S[0:L] 表示严格正整数的 canonical、最短、无符号大端 magnitude：
不含 DER INTEGER 为防符号位而添加的前导 00；首字节 >= 0x80 是允许的。
合法低层表示不能被“规范化”为另一种编码；矛盾摘要按§6拒绝。
使用 leaf 的 serial，不使用 CA 自身的 serial。Serial 的取得来自
真实下层解析证书并验证签名后的结果，不从调用方的新字段读取。

引用关系：

| Proposed 主状态字段 | 本层对应 |
| --- | --- |
| Identity.IssuerID | 逐 byte 等于 I |
| Identity.Serial | 逐 byte 等于 R |
| 节点、Leaf、PublicKey、Window | 完整沿用真实下层已绑定的 Projection，不替换来源 |

主状态字段是本合同下的 digest 值，不能把 bare CADERDigest、
CA SPKIDERDigest、Name digest、SKI/AKI、原始 serial 的 SHA-256、
数据库 issuer 字符串 hash、字符串数字或 hex bytes 冒充 I/R。
只有当另一路径产生的值恰相同，byte 比较才无法区分其赋值来源；
本层不声称从相同 bytes 认证调用方的字段命名、来源或算法执行过程。
域分隔/长度编码定义不同 preimage，不承诺数学无碰撞或发行唯一性。

此映射不写回 Input，不修复 main、不填充缺失字段、不加宽下层时间。
除 Proposed 声明对应外，不新增对 Previous 的材料认证结论。
原 statecontract 对同 Leaf 声明间 issuer/serial 一致性的冲突规则不变。

## 4. 未来接口、Evidence 与责任

以下仅冻结未来 API 合同，不是本轮产品代码：

~~~go
type Input = nodeleafbinding.Input
type Proof uint8
const (
    None Proof = iota
    SelectedIssuerSerialReferencesConditionalTimeBound
)
type Obligation uint8
const (
    UnknownObligation Obligation = iota
    StoreIssuerMappingAndDeploymentProvenance
    KeyHandleAndPrivateKeyPossession
    IdentityEpochLineageHistoryAndRowStatus
    MainStateAndObservationProvenance
    FullIdentityCompositionCommitAndRuntimeAuthorization
)
type Projection struct {
    Leaf nodeleafbinding.Projection
    IssuerID, Serial contracts.Digest
}
type Summary struct {
    Proof Proof
    Profile string
    Evidence nodeleafbinding.Summary
    Proposed Projection
    ObservedAt, ValidUntil time.Time
    Obligations [5]Obligation
}
type Model struct { owned *owner }
func Evaluate(context.Context, Input) (*Model, error)
func (*Model) Summary() Summary
func (*Model) Clear()
~~~

Proof/Obligation 的底层均为 uint8，按上述顺序从零开始。
成功 Profile 恰为§2固定标识；nil/零值/清理后 Summary 为完整零值，
包括 Profile=""、None、零 Projection/时间/Evidence/Obligations。
私有 owner 仅有 input Input、summary Summary、valid bool 三字段。
无额外导出 API、owner/raw/parser 访问器、全局 hook、State/Ready/
Authorize/Commit/Recheck、序列化、缓存许可或通用引用派生工具。

Evidence 是完整真实下层摘要，包括原 Proof、Pending、Obligations、
叶投影及全部嵌套证据。不得删除下层 IssuerAndSerialReferenceMapping，
也不改写任何旧 enum：它记录该下层未完成的较宽责任；本层新增 Proof
仅完成新本地合同的有限材料关系，不解决生产映射及来源。
本层固定五责任依上列顺序填充，第一项显式保留未完成的存储/部署映射。

Proposed.Leaf 直接沿用 E.Proposed；两个新 digest 只在比较匹配后列出。
摘要不新增 CA raw、serial 原始字节输出或生产 issuer 字符串；
原 Evidence 已包含的完整证书摘要和 serial 保持原结构，不删减。
ObservedAt/ValidUntil 精确沿用 E；任何新观测必须完整重评估，
成功摘要不允许缓存到期限内后用于其他观测或运行行动。

## 5. 完整快照、预算与生命周期

先检查 context 和完整预复制预算，取得本层完整自有 Input；所有
原始证书、签名/材料表、Main 十二域、Policy Old/History、时间输入
都来自这份快照。不能先用 caller Input 调下层，再从 caller 读取两个
引用；不能只复制 issuer/serial，也不能暴露或调用旧私有复制 helper。

逐项继承基线649cc089的 nodeleafbinding/budget.go 和
[已批准叶投影规格§5](2026-10-12-main-state-selected-leaf-binding-design.md)。
本层不增加调用方 raw/集合，不放大任何限制。必须在新包本地预检查，
不得只依赖发生在下层复制阶段的检查。主要限制如下，全部边界须覆盖：

| 项目 | 原限制 |
| --- | --- |
| leaf DER / expected SPKI / CA DER / declared CA SPKI | 65,536 / 512 / 65,536 / 8,192 bytes；合计139,776 |
| leaf TrustDomain | 253 bytes |
| 两个材料表 | 每表 Materials/Expected 各128；raw累计每表16MiB |
| 每个材料 raw | RootSet 65,536；其余262,144 bytes；RootRotation 各最多32 |
| Expected 嵌套 | Root keys5；Metadata keys64/revoked4096；Desired slots8 |
| Main Previous/Proposed 各自集合 | Root5；Metadata keys64/revoked4096；两个 trust deauthorized 各128；Desired8；LKG8；Recovery local64/supervisor16 |
| Main.RootTransition / Proposed.Target.Links | 各32 |
| Proposed.Anchor.Raw | 262,144 bytes |
| Binding raw合计 | 32MiB + 262,144 |
| Policy.Batch | Rotations32；四固定及各rotation raw各262,144；累计36×262,144 |
| Binding+Batch raw合计 | 32MiB + 37×262,144 |
| Policy.Old | Roots33，每root keys5；Keys4096；RevokedIDs4096 |
| Old key字符串 | Status8 / RevocationReason32 bytes |
| Policy.History | 4096；每项IncidentKind64 bytes |
| 两个 time raw信封 | 各262,144；合计524,288 |
| Time路径全部raw | 43,778,048 bytes |
| 完整Input raw总量 | 43,917,824 bytes |

按每个出现位置累计，别名及重复出现不能去重预算；加法/减法须溢出安全。
全部预算先于必要非空检查；原四份 certificate raw 及 Proposed.Anchor.Raw
必需非空。其他格式/语义由完整真实下层拒绝，不能用预算通过替代成功。
实施计划将把完整限制逐项展开为新 helper/测试/常量对应，不只测总量。

完整自有复制/清理保持原合同：所有 raw、集合、Old 解除 caller 别名，
仅表顶层原约定排序、nil-empty 规范化及时间 Round(0).UTC()；
其他有序字段不重排，不补缺材料，不改身份或引用值。
每段复制在发生可能失败/取消/panic前登记只清理自有内容的保护。

真实下层返回 owner 后立即登记 Clear，再检查 context/error；
完整 lower.Summary() 直接捕获到已保护的本层 owner，不留未保护的可变
临时摘要。本层 Summary 每次深复制以下两个可变 slice：

~~~text
Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links
Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage
~~~

其余完整 Evidence 值不省略。复制失败有返回摘要保护；owner清理清零两
slice自有backing后清完整Summary/Input；固定hash scratch同样有清理保护。
Model浅拷贝共享owner，任一Clear使全部包装失效；Clear nil-safe且幂等。
独立owner、已返回的detached摘要和caller内容不能被另一owner清理破坏。

并发访问（含Summary/Clear及同时修改Input）不支持。只保证可达自有
内容清理，不保证字符串/runtime隐藏副本或下层从未返回的owner擦除、
致命OOM/crash恢复或硬实时中断。raw预算不是总heap/峰值/延迟上限；
新增整层复制的成本保留，不能称为零复制。

## 6. 判定顺序、摘要一致性与错误

固定流程：

1. nil context返回ErrInvalidInput；可观察预取消/超时优先。检查全部
   预算/必需非空，建立规范化完整自有输入和失败/panic清理保护。
2. 直接 nodeleafbinding.Evaluate(ctx, owned.input) 恰一次，立即保护
   其owner；在检查下层错误前检查当前context。非nil错误原样返回。
   不改为第二次Evaluate、直接旧Evaluate/parser或caller-Summary比较。
3. 成功须非nil owner，完整摘要捕获后要求E.Proof恰为
   nodeleafbinding.SelectedLeafProjectionConditionalTimeBound。
   E.ObservedAt/ValidUntil非零，UTC年1–9999，ObservedAt<ValidUntil；
   E.ObservedAt==E.Evidence.ObservedAt==
   E.Evidence.Time.Policy.Time.Now，E.ValidUntil==E.Evidence.ValidUntil。
   时间比较为规范化完整值，不使用仅instant的Equal。
   不要求E.ValidUntil==E.Evidence.Time.ValidUntil，证书可更早截止。
   E.Proposed恰与自有Proposed的Scope.NodeID、Leaf.Payload、PublicKey、
   Window四项选定值组成的下层Projection相同；原header/kind/envelope
   仍由真实下层检查。本层不制造另一个摘要替代下层。
4. 检查§3的C及leaf serial固定表示，先检查范围再索引/切片；
   在自有保护下构造exact preimage、计算I/R、清scratch。
   先比Proposed.Identity.IssuerID，再比Identity.Serial，不修复输入。
5. 建立本层Projection、Proof、Profile、五责任及原时间；
   最终context检查成功后才设置valid并移交owner；失败返回nil模型。

新增恰四个固定sentinel，其文字不含调用方内容、证书、serial、digest、
路径或外部错误动态文本：

| 名称 | 固定文字及语义 |
| --- | --- |
| ErrInvalidInput | nodeissuerbinding: invalid input；nil context/原必需非空 |
| ErrLimitExceeded | nodeissuerbinding: limit exceeded；完整预复制预算 |
| ErrInvalidEvidence | nodeissuerbinding: invalid evidence；成功却nil/错Proof/矛盾摘要或非法C/serial表示 |
| ErrBindingMismatch | nodeissuerbinding: binding mismatch；两个新引用对应失败 |

下层bad signature、过期、原leaf投影错配、格式/关联、typed PolicyFailure
等原错误不包装、不重分类。下层判定先于新映射；下层过期加新引用
错配仍返回原过期错误。每阶段在错误/发布前检查context；当前观察到的
Canceled/DeadlineExceeded优先，正常panic清理后传播，不recover为成功。
不承诺中断同步crypto、分配、排序或从未观察到的取消瞬间。

## 7. 能力隔离与未来精确例外

产品只从nodeleafbinding取得真实判定和Summary/Clear。其他下层导入仅为
Input的完整预算/复制/清理类型：contracts、nodecerttime、nodeissuer、
statetimebinding、statepolicybinding、statebinding、statecontract、
statematerials、statepolicy；禁止调用其Evaluate/Verify/parser。
statepolicy自有Old.Clear只清已取得所有权的模型。
不直接导入stateverify、stateformat、statecomposition、statetime、
nodeca、nodeleaf或nodeleafkeys来替代完整下层。

标准库限context/errors/bytes/sort/time和crypto/sha256：
前五者仅原完整复制、排序、规范化、阶段context和固定error必要能力；
sha256只准Sum256，固定用于§3两个preimage，正常成功路径各计算一次；
失败/取消可在计算前或两者之间退出。本层不直接解析DER/JCS/JSON，
不直接做签名/chain验证；真实下层仍执行原验证。本层不用encoding、hex/base64、
动态反射、I/O、进程、环境、系统时钟、随机、网络、provider/数据库、
unsafe、goroutine、recover、hash.Hash、流式hash或额外crypto。
禁止函数/方法值逃逸、dot/blank/重复import、shadow能力及额外导出。
实施计划固定精确文件库存、成员白名单、AST/schema和实际mutation控制。

未来旧源码仅准以下十个测试文件新增精确消费者例外及正反控制：

1. nodeleafbinding/consumer_boundary_test.go
2. nodecerttime/consumer_boundary_test.go
3. nodeissuer/boundary_test.go
4. statetimebinding/boundary_test.go
5. statepolicybinding/boundary_test.go
6. statebinding/boundary_test.go
7. statecontract/boundary_test.go
8. statematerials/boundary_test.go
9. statepolicy/boundary_test.go
10. stateverify/boundary_test.go（仅其statepolicy导入例外，不授权直接stateverify）

所有相对目录位于internal/nodecontrol。例外仅exact新目录/package
nodeissuerbinding的直接Go文件、exact目标导入；拒绝外部测试package、
子目录、sibling、伪同名package、其他根和subpath；原规则完整保留。
新包自己的消费者仍只准exact本目录/package的_test.go；
不授予serving/provider/store/runtime消费者或安装/保存证书能力。

旧产品、fixtures、SQL、模块依赖、原接口/enum/限额一律不变；不通过
导出旧私有helper省掉本层所有权，也不修改旧guard使新代码无条件通过。
本轮未实施任何上述例外。

## 8. 独立材料、测试与有限验收

未来使用离线固定签名材料、测试key、显式时间，不访问真实凭据/时钟。
可在专属忽略目录一次性AST提取现有固定leaf-binding输入和manifest，
独立解码/解析certificate DER、serial和CA摘要，按§3创建匹配声明；
旧raw和旧fixtures不变。普通测试从本层固定常量解码，不读旧fixture
源码，不动态重签，不用本层derive/Evaluate输出反推期望。
一次性生成器最终保留为非Go源码归档，不能借忽略目录绕过旧consumer
扫描；原始失败、归档hash及重放代价如实记录。

需要新CA续签/同serial不同issuer样例时，仅在隔离一次性生成器用固定
测试key生成真实有效签名材料，再独立解析核对后冻结。不得只改DER
而绕过签名，不改生产材料，不能在普通测试随机生成正例。

测试覆盖至少包括：

1. 真实下层及新层完整成功；exact Profile/Proof/Projection/全Evidence/
   五责任、纳秒时间和证书提前截止不变。
2. 独立固定preimage+digest向量，覆盖serial长度1/20、首字节>=0x80及
   区分用途/版本/NUL/长度/顺序/错误全容量尾padding/hex等编码。
3. 错issuer、错serial及两者同时错；bare CA DER/SPKI/Name/SKI/AKI、
   原始serial hash、CA自身serial、不同CA下相同serial等替代输入。
   数值恰相同不能被声称可辨认来源；实际负例先证明数值确有区别。
4. 同公钥/同名字、不同真实CA DER（续签），在相同合法观测下真实
   下层分别成功；旧CA引用不得套新材料。明确这是exact证书身份选择，
   不是生产issuer字符串或公钥身份认证。
5. 各可达主状态错配负例先assert完整nodeleafbinding成功，再要求
   本层nil模型/ErrBindingMismatch；若同Leaf触发旧transition冲突，
   同时一致调整Previous/Proposed未签名声明以满足原合同。
   lower已拒绝的前置负例独立标记，不能算到达本层。
6. 私有摘要一致性负例：C零、L0/21、首零、尾部非零、错Proof/时间/
   下层投影；仅私有helper测试可构造，产品无替换lower hook。
7. 下层错误身份/typed链、双故障优先级、完整预算等长/+1、重复位置/
   未选材料/字符串/所有集合与Old。全部raw原始hash和声明变化路径核对。
8. 深复制、caller/共享backing隔离、两Evidence slices、独立owner/
   Model浅拷贝共享Clear、nil/zero/幂等、hash scratch及可达自有清零。
9. 动态枚举本层和同步下层可观察context.Err点，注入Canceled、
   DeadlineExceeded及正常panic；caller不变，失败nil或清理后panic。
   点数只记实际观察，不固定套用上一轮655或称硬实时保证。
10. 未认证key-handle/epoch/lineage/history/来源在原下层合法范围内
    一致变化仍有限成功，全部责任不消失；只证明未覆盖而非有效。
11. exact消费/成员/库存/预算/所有权/一次lower/最后发布/两次固定hash
    mutation；均实际改变源码后被拒绝。良性格式、局部变量及命名
    import别名不应误报；有限AST/schema不是形式化或穷尽安全证明。

未来最小矩阵：新nodeissuerbinding完整包，加基线的20个完整包
nodeleafbinding、nodecerttime、nodeissuer、nodeca、nodeleaf、nodeleafkeys、
contracts、serving、statebinding、statecomposition、statecontract、
stateformat、statematerials、statepolicy、statepolicybinding、staterecovery、
statetime、statetimebinding、stateverify及internal/readiness；再加authority
指定TestBoundRead/TestCoordinator、实际internal/store仅编译（不运行测试）、
两项既有nodeissuer短fuzz，共21+1+1+2=25独立命令，格式另记。
PowerShell、缓存Go1.26.5、离线只读模块；完整普通包/指定测试仍用
原120s限制，两个fuzz配置5s并保留原30s命令timeout。5s不是整个进程
时长保证，次数/基线/整体耗时及真实失败必须记录，不改时限掩盖失败。
具体命令、RED/GREEN步骤、文件分工及精确库存由获批后的实施计划固定。

未来实现完成需fresh只读增量质量/安全审查；修复真实RED/GREEN，
重跑完整25矩阵并同审查者复查，不能预写零问题。本次规格撰写不派
实现代理、不另建用户chat、不执行Go命令或独立实现审查。
上述是未来验收要求，旧阶段测试成功不能代替本增量测试结果。

不做整仓/race/Linux、真实DB/Docker/TLS/provider、硬件/reboot/PITR；
不关闭C12/B01/Task10/main、生产兼容UNKNOWN、物理验收、历史失败原因、
I2/I3或其他既有阻塞。安装Docker不等于真实验收。

## 9. 自检、阶段门槛与本轮交付

作者自检核对：两个36-byte域/68-byte issuer及70–89-byte serial编码；
exact CA证书而非公钥/存储字符串身份；叶serial而非CAserial；Proposed/
Previous责任区分；不修正声明；完整预算/复制/两slices；一次真实下层、
原时间/错误/责任保留；新局部Proof不删除旧较宽责任；十个仅测试例外；
25有限矩阵；无运行消费者。无算法、范围或接口待填项。
本文中的文件分工/命令展开属于之后的实施计划，不是隐含实施授权。
作者自检不是独立审查、测试通过或用户对本文的书面批准。

本轮只在本机新增本文，不修改路线图、产品、测试、SQL或模块，不运行
Go、不暂存/提交/推送、不合并或清理。按用户阶段约束，不采用技能的
默认文档提交步骤；代价是这份待审规格尚不是远端恢复点。

请用户审阅本文。书面规格批准后，仅进入writing-plans编写实施计划；
计划再获批准并确定执行方法后才可实现。Git保存、推送、合并和整体
验收继续各自受独立授权及门槛约束，之前保存649cc089不覆盖本增量。
