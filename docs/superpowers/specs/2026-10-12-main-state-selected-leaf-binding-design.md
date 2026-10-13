# 主状态选中叶证书投影与条件时间绑定设计

日期：2026-10-12。

状态：待书面审阅。用户已确认独立、有限的身份投影方向，并同意起草本文；
不表示本文或实施计划已获批准。本轮仅起草，不实施、不暂存、提交或推送。
起草基线为既有开发 worktree 的
`a820f550a4e1662143676dc3f58d4f385c14f5a1`，不是 main 验收基线。

## 1. 意图、依据与成功范围

关闭当前 nodecerttime 的一个明确组合缺口：节点相同且直接签名与条件时间
成立，不代表这份 leaf 是主状态 Proposed.Identity 实际引用的 leaf。
新增隔离层只绑定本次提出的主状态中的节点、exact leaf DER、公钥引用和
证书窗口；沿用同次条件时间，不接入身份切换状态机或运行授权。

依据：父设计 `2026-08-23-node-pop-control-plane-design.md` §6.1–6.3，
主状态覆盖契约 `2026-10-10-task10-main-state-contract-design.md`，
`2026-10-12-node-certificate-time-binding-design.md`，以及基线中的
nodecerttime、statecontract、statebinding、nodeleaf、nodeleafkeys 公共类型。

成功仅表示：完整真实 nodecerttime 判定通过，且本次 Proposed 主状态的
四类投影字段与其直接证书/条件时间证据一致。不是完整 Identity 域验收，
不验证旧状态选中证书、不证明 Proposed 已提交或被真实平台选中。
输入中的主状态及历史仍是显式条件声明，不是 authenticated sealed state。

## 2. 方案与分解

采用独立 `internal/nodecontrol/nodeleafbinding` 包，只调用一次真实
nodecerttime.Evaluate；不修改下层产品接口或职责。

备选一是扩展 nodecerttime，会把原时间证据与身份投影接口耦合；备选二是
同时实现 issuer/serial/row/history/key-handle/epoch 和授权，需新的材料与
来源契约，范围超出本增量。两者本轮均不采用。

本层继承 nodecerttime 及其真实时间下层的全部适用范围：Old 已存在、非
FreshInstall、旧新 Desired 已存在、完整 Root/Metadata 不变、floor 和
原历史/检查点关系均按原规则。不得通过新入口放宽旧限制或改错误类别。
只投影 Proposed.Identity；Previous 完整保留并仍接受下层全部检查。
不新增“已选中”布尔值、身份材料表、数据库行、provider 或回调。

## 3. 显式局部映射合同

为避免把数值相同的不同摘要类型自动互换，本规格明确提出局部隔离 profile
`main-selected-leaf.der-spki-window.v1`。它没有 wire 字段或运行时选择开关，
不是数据库或生产持久格式的兼容声明；只有书面批准后才能用于实现。

令 `main = owned.Time.Policy.Binding.Main.Proposed`，
`leaf = lower.Certificate.Leaf.Leaf`。投影关系为：

| 主状态字段 | 本层精确对应关系 |
| --- | --- |
| Scope.NodeID | 等于 leaf.NodeID 的 canonical UUID 字符串；并保留原完整跨节点检查 |
| Identity.Header | DomainIdentity / Present，不能用缺域或零值绕过 |
| Identity.Leaf | Kind=CertificateDER；Payload=leaf.LeafDERDigest；Envelope 为零 |
| Identity.PublicKey | 等于 leaf.SPKIDERDigest，即 exact leaf RawSubjectPublicKeyInfo DER 的 SHA-256 |
| Identity.Window.Issued | time.Unix(Issued, 0).UTC() 与 leaf.NotBefore 精确同 instant |
| Identity.Window.Until | time.Unix(Until, 0).UTC() 与 leaf.NotAfter 精确同 instant |

两个窗口整数先验证 UTC 年 1–9999 对应的既有秒范围
`[-62135596800, 253402300799]`，并要求 Issued < Until，再转换。
不以 leaf.Unix() 截断证书时间，不修复或替换字段，不容许偏差。
证书窗口由既有严格 parser 给出整秒，本次 Now 仍保留纳秒。
窗口合同新增映射的是证书 NotBefore/NotAfter，不是签发请求的 issuer_now。

LeafDERDigest 与 SPKIDERDigest 的 exact preimage 使用真实下层已有结论；
本层不重算摘要、不读取 parser 对象、不用 SKI/AKI 代替主状态引用。
既有 leaf SKI 使用同一 SPKI preimage，数值可能恰等于 SPKIDERDigest；
本层比较的是显式合同下的摘要值，不可能从相同 bytes 推断其赋值来源，
不声称能够拒绝恰好相等的 SKI 数值或认证调用方的字段命名。
这条局部 PublicKey 合同不得自动写入 store.PublicKeyDigest、数据库行、
其他引用或旧设计。生产兼容仍为 UNKNOWN，未来接入须单独确认。

Identity.IssuerID 和 Identity.Serial 在既有模型中是 contracts.Digest，
而证书序列号是 `[20]byte` 加长度；本层不定义其 hash preimage，
不把 CA DER/SPKI/AKI 当成 issuer ID，不比较摘要与原始整数。
它们及 Identity.KeyHandle、IdentityEpoch、LineageID、LineageUntil、History
保留原下层结构/转换检查，但不取得本层材料认证结论。
证书没有可用于核对 authority/identity epoch 或 lineage 的对应字段。
不把“未覆盖”字段默认为匹配，也不把有限成功称为完整身份有效。

## 4. 接口、摘要与责任

以下为未来实现接口，不是本轮代码：

```go
type Input = nodecerttime.Input
type Proof uint8 // None=0; SelectedLeafProjectionConditionalTimeBound=1

type Projection struct {
    NodeID string
    Leaf, PublicKey contracts.Digest
    Window statecontract.TimeWindow
}
type Summary struct {
    Proof Proof
    Evidence nodecerttime.Summary
    Proposed Projection
    ObservedAt, ValidUntil time.Time
    Obligations [5]Obligation
}
type Model struct { owned *owner }
func Evaluate(context.Context, Input) (*Model, error)
func (*Model) Summary() Summary
func (*Model) Clear()
```

Obligation 为 uint8，零值 UnknownObligation；有效结果固定按下列顺序返回：
IssuerAndSerialReferenceMapping、KeyHandleAndPrivateKeyPossession、
IdentityEpochLineageHistoryAndRowStatus、MainStateAndObservationProvenance、
FullIdentityCompositionCommitAndRuntimeAuthorization。
Projection 不带未认证 issuer/serial/epoch/key-handle 值，避免读者把列出的
字段误认为本层已验证；完整原声明保留在自有输入中。

Evidence 完整保留所有层级 Proof、Pending、Obligations、Links 和 Coverage。
不删除下层的较宽未完成责任，即使本层完成其一个有限子关系。
ObservedAt 与 ValidUntil 精确沿用 Evidence，不另加宽或收紧期限；主状态
窗口已经与 leaf 相同，无独立期限来源。每个新观测完整重新 Evaluate。
结果不是缓存到期前可复用的许可，不提供 Recheck/State/Commit/Ready/
Authorize、原始字节访问器、下层 owner 或可信等级。

## 5. 完整快照、预算与生命周期

本层必须先取得自己的完整 Input，然后把该份快照送给下层并读取投影。
不能先把 caller Input 给下层，再从 caller Input 读取 Identity；不能只
复制本层投影四个字段或依赖下层私有输入访问器。

预算逐项继承基线 nodecerttime 的书面规格§4及 budget.go 的完整限制，
作为本层预复制合同：全部四份证书 raw、两张 Materials/Expected、完整
Main 十二域和集合、Anchor、Batch rotations/raw、Old roots/keys/revoked、
History 和两份时间信封均检查；完整 raw 上限为 43,917,824 bytes。
各位置按出现次数累计，溢出安全；包括未选材料，预算超限先于必需非空。
本层 Input 不增加 raw 或集合，不修改各项限额；实施计划须展开逐项表和
常量对应，不能只检查总量。raw 限额不是对象 heap 或峰值内存上限；
本层与下层各自复制意味着同调用存在额外内存开销，不能声称零复制。

完整所有权、失败/panic 清理和规范化继承原合同：解除 caller 别名后
登记自有清理；所有集合/raw/Old 深复制；仅表顶层排序、nil-empty 规范化
及时间 Round(0).UTC()，其他有序字段不改；不补缺材料。
nodecerttime 返回的 owner 立即登记 Clear，先于检查 context/error；
完整下层 Summary 直接捕获到已保护的本层 owner，不能留下未受保护的
可变临时副本。Summary 每次独立复制完整可变 Links/Coverage。

Model 浅拷贝共享私有 owner，任一 Clear 使全部副本无效；nil/零值/清理后
Summary 为完整零值；Clear nil-safe、幂等。禁止并发访问，包括 Summary/
Clear、输入修改。只清理自有可达内容，不损坏 caller 或已返回摘要；不
声称字符串/runtime 隐藏副本擦除、致命 OOM/crash 恢复或硬实时取消。

## 6. 流程、时间与错误

固定流程：

1. nil context 返回 ErrInvalidInput；预取消/超时优先。检查全部预算及
   原必需非空材料，取得并保护完整规范化自有快照。
2. 调用 nodecerttime.Evaluate(ctx, owned.input) 恰一次，立即保护其 owner。
   下层错误原样返回；成功捕获完整摘要，要求非 nil owner、Proof 恰为
   NodeCertificateConditionalTimeBound、非零且有效的 ObservedAt/ValidUntil，
   UTC 年均在 1–9999，ObservedAt < ValidUntil，且
   ObservedAt == Evidence.Time.Policy.Time.Now（规范化完整值）。
   不要求 ValidUntil == Evidence.Time.ValidUntil；CA/leaf 可能更早到期。
3. 按§3表顺序核对 Proposed 节点、header、leaf 引用、公钥及完整窗口。
   已有下层结构检查不能替代本层精确投影比较。
4. 保留全部责任并建立 Projection；ObservedAt/ValidUntil 沿用下层。
   最终 context 检查后才设置有效标记并发布模型。

新增恰四个固定 sentinel，文字不含输入、UUID、摘要、raw、路径或外部文本：
ErrInvalidInput=`nodeleafbinding: invalid input`；
ErrLimitExceeded=`nodeleafbinding: limit exceeded`；
ErrInvalidEvidence=`nodeleafbinding: invalid evidence`；
ErrBindingMismatch=`nodeleafbinding: binding mismatch`。
前两者用于预复制合同；第三者用于下层成功却 nil/错误 Proof 或矛盾时间
摘要；第四者用于本层投影字段关系失败（含窗口范围/精确对应失败）。

下层证书未生效/已到期、格式/签名/关联、PolicyFailure 等非 nil 错误完整
保留，不包装或转为本层错误。下层判定先于本层投影，所以同时出现下层
过期与投影错配时仍返回原过期错误；不为识别投影错配跳过真实下层。
每阶段在检查错误及发布前检查 context，当前检测到的 context 错误优先；
不承诺同步 crypto/分配/排序实时中断。panic 清理后传播，不 recover 成成功。

## 7. 隔离与准许未来修改

唯一新产品目录为 internal/nodecontrol/nodeleafbinding。产品只调用
nodecerttime.Evaluate/Summary/Clear；其他下层包仅为完整值复制、预算、
摘要和自有清理的类型依赖，不直接调用其他 Evaluate/Verify/parser。
标准库限定 context/errors/bytes/sort/time 的必要成员；time.Unix 仅以
已检查窗口整数和零纳秒调用并转 UTC，Equal 仅用于窗口 instant 对应；
保留 Round(0)/UTC/Before/Year/IsZero 及原复制所需能力。禁止系统时间、
hash/crypto、JSON/DER/JCS 重解析、动态反射、I/O/环境/进程/随机、网络、
provider、数据库、unsafe、goroutine、recover、dot/blank/重复 import。
UUID.String 仅用于本层 canonical 节点对应，不引入其他 UUID 能力。

未来旧源码例外仅以下九个测试文件：nodecerttime/consumer_boundary_test.go，
nodeissuer/statetimebinding/statepolicybinding/statebinding/statecontract/
statematerials/statepolicy/stateverify 各自的 boundary_test.go。
前者仅增加 exact nodecerttime import 的新包消费者；中间七个包只开放各自
类型复制依赖；stateverify 仅其 statepolicy 消费扫描的新例外，仍不允许
新包导入 stateverify。nodeissuer 不准直接 Evaluate；能力白名单另行守卫。

每项例外要求当前仓库根内直接目录、package 恰为 nodeleafbinding、
import 完全相等；允许直接产品及同包测试，拒绝 nested、类似前缀、
外部根、错误 package、外部测试包及 sibling。原例外/反例保持不变。
新包不开放任何运行消费者；不修改 nodecerttime 原产品/API/Proof/AST
schema，不修改其他产品、fixtures、SQL/module/lock、历史规格或完成记录。
实施计划冻结精确文件/API/成员库存和全部旧源码摘要；发现越界必须修订
书面范围，不能用本设计默认扩权。

## 8. 验证设计与有限验收

采用离线固定签名样例、测试 key 和显式时间，不访问真实凭据/时钟/网络。
新增匹配的主状态声明可在本层测试中构造；不改旧 fixture，也不在普通
测试中随机重签。组合负例必须有真实 nodecerttime 单独成功的控制，
同时按原 statecontract 规则一致调整 Previous/Proposed，避免旧同 leaf
字段冲突遮蔽本层待测关系。不能通过替换下层的产品 hook 伪造证据。

覆盖要求：

1. 真实下层和本层分别成功；完整 Evidence/所有责任、纳秒 Now、期限不变。
2. node/header/Kind/Payload/Envelope/PublicKey/窗口逐项错配；所有先决
   负例与可达投影负例分开标注；同节点另一有效签名 leaf 不得套用原引用。
3. exact DER 不同但 SPKI 相同、证书引用相同但公钥引用不同均拒绝；
   数值不相同的 CA digest、原始 EC 点 hash 等引用拒绝；数值恰与
   SPKIDERDigest 相等的 leaf SKI 正控制允许，并仍保留来源未认证责任。
4. 秒窗口恰等、各端点前后 1 秒、合法 year-1 端点、范围外/反向窗口；
   原 Now 前后 1 ns 半开规则保持。下层失败先于本层错配的双故障控制。
5. 完整预算等长/+1、所有集合、累计 raw、重复位置计数；容量通过不
   扩称语义通过。完整输入规范化/别名、摘要深复制和共享 Clear 生命周期。
6. 新观测完整重评估，不能沿用旧摘要；失败/取消返回 nil 模型、输入
   不变；取消/超时/正常 panic 的清理登记及可达自有内存检查分别报告。
7. 未覆盖字段在下层允许的范围内一致改变仍可产生有限成功，且五项责任
   不消失；这仅证明本层未宣称覆盖，不表示发行/历史/row 状态有效。
8. 九个精确消费例外及反例，产品成员白名单、完整库存、所有权登记和
   无运行消费者守卫；AST/schema 控制不是形式或穷尽安全证明。

未来最小矩阵为新 nodeleafbinding 完整包，加上一阶段 19 个完整包、
指定 authority TestBoundRead/TestCoordinator、store 仅编译不运行测试、
两项既有 nodeissuer 短 fuzz，共 24 个独立命令。格式另记。
使用缓存 Go 1.26.5、PowerShell、离线只读模块；具体命令、时间限制、
证据路径、RED/GREEN、源码冻结及 fresh 增量质量/安全审查写入实施计划。
本轮不运行这些测试、不声称新增测试通过。

不做整仓/race/Linux、真实 DB/Docker、TLS/provider、硬件/reboot/PITR。
有限成功不解除生产兼容 UNKNOWN、Task10/B01/C12/main 的既有阻塞。

## 9. 作者自检与下一门槛

已核对：Proposed 与 Previous 责任区分；DER/SPKI 摘要角色显式而非隐式
类型互换；issuer/serial 不猜 hash；秒窗口不截断证书时间；完整快照和
有界深复制；原错误顺序及全部责任保留；九个仅测试消费者例外；无运行
能力。局部 profile 是待批准的新模型合同，不预告生产格式兼容。
本文无待填占位项，实施文件分工和精确测试命令属于后续计划。
作者自检不是独立审查、测试结果或用户对本文的书面批准。

本轮仅本机保存本文，未修改路线图或其他文件，未提交/推送；不按技能
默认文档提交步骤扩大本次“起草书面设计”授权。用户审阅并批准本文后，
下一步仅使用 writing-plans 编写实施计划；计划再获批准并确定执行方法
后才可实施。Git 保存、推送、合并及整体验收仍受各自授权和门槛约束。
