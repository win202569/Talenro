# 节点证书与条件时间证据绑定设计

日期：2026-10-12。

状态：待书面审阅。用户已选择时间证据绑定方向，并确认组件边界及时间、
错误顺序和测试规则；这些确认允许起草本文，不表示本文或实施计划已获批准。

工作区为既有 `codex/c12-b01-task9-coordinator` worktree，起草基线为
`f5d731bef9d767d57ca32f0f4c3a0f532a4bc322`，不是 main/C12 验收基线。
本轮仅新增本文，不暂存、提交、推送、创建 PR、合并或清理。

## 1. 意图与证明范围

新增独立离线组合包 `internal/nodecontrol/nodecerttime`，将真实
`nodeissuer.Evaluate` 的直接 CA–leaf 关系与真实
`statetimebinding.Evaluate` 的条件时间证据绑定到同一份自有输入。
成功只表示：本次直接签名关系成立；两组输入属于同一节点；在本次条件
时间观测 Now 下，所提供的 CA 和叶证书都处于本设计的有效窗口内。

依据为父设计 [节点身份与信任域](2026-08-23-node-pop-control-plane-design.md)
§6.1–6.3、[直接签名关系设计](2026-10-12-node-leaf-issuer-relation-design.md)
及 [条件时间证据绑定设计](2026-10-11-main-state-time-evidence-binding-design.md)。
源代码依据为上述两包，以及它们现有输入、摘要、所有权和隔离守卫。

当前时钟的真实性、平台观测来源、TLS 验证、nonce 原子消费、持久检查点
推进并未由 statetimebinding 认证。nodeissuer 也没有建立 CA 来源、parent/
root/chain/path、部署 purpose、bundle 状态或回滚历史的信任。本层不会
把这些条件假设升级成事实，也不会删除任何下层 Pending/Obligations。

本次不绑定主状态 Identity.Leaf、issuer/serial/public-key/key-handle 引用、
身份 epoch/lineage、证书行状态或当前选中身份。即使节点 UUID 相同，本层
仍不证明该 leaf 是主状态实际选中的 leaf，或 CA 属于部署授权信任包。
证书没有本层可使用的 authority epoch 字段，不把时间摘要 epoch 声称为
已签入证书；TrustDomain 仍只由旧 leaf 规则绑定调用方声明。

不产生 State/Commit/Ready/Authorize、连接许可、恢复放行、私钥持有或
启动能力。生产/provider 兼容保持 UNKNOWN；Task10/B01/C12/main 既有阻塞不变。

## 2. 方案与组件

用户已选择继续时间绑定，而非先做主状态身份映射或完整 CA 信任包。
对该方向采用独立组合层，不修改已冻结的两个下层产品职责：

| 方案 | 取舍 |
| --- | --- |
| 独立 nodecerttime（采用） | 两个真实判定器复用，新增跨输入关系与时间窗；需明确的旧消费守卫例外 |
| 扩展 nodeissuer | 改变旧签名关系接口及其禁止时间/主状态能力的边界，不采用 |
| 扩展 statetimebinding | 把证书材料加入现有时间/策略模型，改变旧输入和证明范围，不采用 |

本层只调用 nodeissuer.Evaluate 和 statetimebinding.Evaluate 两个完整
判定入口，再读取独立 Summary 并清理下层 owner。其他依赖仅用于完整值
复制、预算和自有值清理，不直接调用它们的 parser/Evaluate/Verify/State。
无新 wire、持久格式、provider 接口、外部回调、依赖安装或运行消费者。

复用 statetimebinding 的原范围：已存在 Old；非 FreshInstall；旧/新
Desired 已存在；旧/新完整 Root/Metadata 不变；既有 root 历史和检查点、
时间 floor 关联仍按原规则核对。本层不放宽这些限制，也不提供替代路径。

## 3. 输入与公开接口

公开接口冻结为以下形状；这里只是规格声明，不是本轮实现：

```go
type Input struct {
    Certificate nodeissuer.Input
    Time        statetimebinding.Input
}

type Proof uint8 // None=0; NodeCertificateConditionalTimeBound=1

type Summary struct {
    Proof                  Proof
    Certificate            nodeissuer.Summary
    Time                   statetimebinding.Summary
    ObservedAt, ValidUntil time.Time
}

type Model struct { owned *owner } // owner 为私有类型

func Evaluate(context.Context, Input) (*Model, error)
func (*Model) Summary() Summary
func (*Model) Clear()
```

Proof 恰为 None 和 NodeCertificateConditionalTimeBound，不增加可信/可运行
等级。无 Recheck、原始字节访问器、可变 parser 对象或第三张责任表。
每次新观测须用完整输入重新 Evaluate；旧摘要不是新调用的输入凭据。

不接受 caller-built Summary/Model、verified bool、额外 Now、系统时间、
时间 provider 句柄、签名回调、私钥、容差、连接 deadline 或租约参数。
现有 Policy.Old 仍是下层所需的显式不可信历史模型，不是可信 sealed state。

Time 和 Certificate 嵌套摘要完整保留原 Proof、Pending 和全部层级的
Obligations。Time.Policy.Binding.Proposed.Target.Links 和 Coverage 在捕获、
每次 Summary 返回时独立复制，其他字段按实际类型逐值保留。
ObservedAt 等于已核对的 Time.Policy.Time.Now，不另外生成时间。
ValidUntil 是第 7 节的条件上限，不是运行授权或当前状态仍未变化的证明。

Model 包装拷贝共享同一 owner 生命周期；任一包装 Clear 后其他包装读到
完整零值 Summary。nil/零值/已 Clear 的 Model 均返回完整零值摘要；Clear
nil-safe、幂等。独立 Evaluate 不共享可变 backing store。不得并发修改
输入、并发 Evaluate/Clear/读摘要；不提供并发安全或线性化保证。

## 4. 完整预复制预算

预算检查覆盖两个完整输入，先于任何自有切片、索引或 Old 指针分配。
所有容量检查完成后才检查本节列明的必需非空 raw；其余值形状和域语义
留给真实下层，不在本层复制 UUID/domain、DER 或签名规则。

| 范围 | 上限 |
| --- | --- |
| Certificate.Leaf.DER / ExpectedSPKI | 65,536 / 512 bytes |
| Certificate.CA.DER / DeclaredSPKI | 65,536 / 8,192 bytes |
| Certificate 四份 raw 合计 | 139,776 bytes |
| Certificate.Leaf.TrustDomain | 253 bytes（域语法仍由 nodeissuer 检查） |
| Time.Policy.Binding 两个表的 Materials / Expected | 每表各 128 项 |
| 每表 Materials / Expected 的 RootRotation；Target.Links / Main.RootTransition | 各 32 项 |
| 表中 RootSet raw / 其他 raw；独立 Anchor.Raw | 65,536 / 262,144；262,144 bytes |
| 每表 raw；完整 Binding raw 含独立锚 | 16 MiB；32 MiB + 262,144 bytes |
| 每个 Expectation 的 root keys / metadata keys / revoked / desired slots | 5 / 64 / 4,096 / 8 项 |
| 两份 Main 各自 root keys / metadata keys / revoked | 5 / 64 / 4,096 项 |
| 两份 Main 各自 bootstrap / agent deauthorized | 各 128 项 |
| 两份 Main 各自 desired slots / LKG / local faults / supervisor faults | 8 / 8 / 64 / 16 项 |
| Time.Policy.Policy.Batch.Rotations；Batch 每次出现的 raw | 32 项；262,144 bytes |
| Batch raw 合计；完整 Time.Policy raw | 36 × 262,144；32 MiB + 37 × 262,144 bytes |
| Policy.Old.Roots / 每 root public keys | 33 / 5 项 |
| Policy.Old.Keys / Old.RevokedIDs / History | 各 4,096 项 |
| 每个旧 key 的 Status / RevocationReason；每个 History.IncidentKind | 8 / 32；64 bytes |
| Time.Time.AttestationEnvelope / MetadataEnvelope；两者合计 | 各 262,144；524,288 bytes |
| 完整 Time raw | 32 MiB + 39 × 262,144 = 43,778,048 bytes |
| 完整 Input raw | 43,917,824 bytes |

累计用 `len > limit-total` 等溢出安全比较；按位置出现次数计数，不按
backing pointer 去重。跨输入或同表别名均计数并独立复制。raw 总量不
包括模型元素/字符串头等对象的 heap 开销；不是峰值内存或耗时上限。

Certificate 四份 raw、Time.Policy.Binding.Proposed.Anchor.Raw 必须非空，
否则在全部容量检查之后返回 ErrInvalidInput。其他缺失/无效材料、Old、
观测状态和身份、未知 MaterialKind 等保持真实下层的错误判定与先后顺序。
预算等长通过只证明容量合规，不证明材料可解析或语义可接受。

## 5. 所有权与规范化

完整复制 Certificate 四份字节；Time 的全部十二域、两张完整材料/期望
表、目标/轮换/锚、Batch 全 raw 及二维 rotations、Old 指针和所有 roots/
public keys/keys/revoked、完整 History、两份时间信封和全部观测值。
不删未选材料、不截断历史、不以少数摘要字段代替整个输入。

仅沿用旧规范化：空集合为 nil；两张材料表的顶层 Materials/Expected 按
Kind/P 排序，其他集合顺序保留；已有所有 time.Time 字段用 Round(0).UTC()
保留完整纳秒、消除进程本地 monotonic/时区表示差异，零值保留真正零值。
这不是向整秒舍入；主状态 UnixSeconds、raw、签名和所有摘要 preimage 不变。
不补缺失时间，不提高 floor，不从 attestation 设置时钟，也不改写证书窗口。

每份可变字段在克隆/分配前建立失败及正常 Go panic 清理责任；先解除
调用者切片别名再登记清理，不能因半途失败清零调用者原数据。取得完整
owned 输入后，在下一次 context 回调或下层调用之前保护本层 owner。
下层返回的 owner 立即登记 Clear，先于 context/错误检查；有可变集合的
Time.Summary 直接捕获到已保护的本层摘要字段，不能留未受保护的临时副本。

仅最终 context 检查成功后设置 valid 并转移 owner。失败/取消返回 nil
模型；panic 在清理后传播，不 recover 成成功。Clear 清零可达自有 raw、
集合、Old 和摘要，不能损坏调用者输入、已交出的独立摘要或其他模型。
不承诺字符串、不可达/runtime 副本擦除、致命 OOM/crash 恢复或硬实时。

## 6. 同快照流程及关联

固定顺序：

1. nil context → ErrInvalidInput；检测到预取消/超时则原样返回。检查完整
   容量、必需非空 raw，再取得并保护完整规范化的 owned Input。
2. 调用 nodeissuer.Evaluate(ctx, owned.Certificate)，立即 defer Clear。
   失败保持原下层错误；成功捕获摘要，检查非 nil owner 和恰为
   NodeLeafDirectIssuerSignatureBound，不重跑 parser/验签或重建字节。
3. 调用 statetimebinding.Evaluate(ctx, owned.Time)，立即 defer Clear。
   失败保持原下层错误；成功直接捕获受保护摘要，检查非 nil owner、
   恰为 ConditionalTimeEvidenceBound 及用于以下关系的必要有效值。
4. 绑定节点和已检查的时间快照。证书摘要 Certificate.Leaf.Leaf.NodeID
   必须等于 owned.Certificate.Leaf.NodeID 和 Time.Time.NodeID；后者必须
   对应 owned.Time.Time.Metadata/Request/Checkpoint 的 NodeID 及两份
   Main.Scope.NodeID 的 canonical UUID 字符串。
5. 输出 Time.Policy.Time 的 ProviderID、ObservationID、Now、ObservedFloor
   与 owned.Time.Time.Time 四个字段及 owned.Time.Policy.Policy.Time 的
   ProviderID、ObservationID、Now、Floor 一致；不引入第二份独立 Now。
   Old.TimeProviderID、Old.TimeFloor、两份 Main.Scope 和全部其余关联仍
   由真实 statetimebinding 完整检查，不接管其业务判定。
6. 依第 7 节先检查 CA 时间窗，再检查叶证书时间窗，建立 ObservedAt/
   ValidUntil 和完整嵌套摘要。最后检测 context，再发布结果。

每个阶段在判定前及处理阶段错误时检查 context；检测到取消/超时优先
于该阶段错误。只有一次同步调用内的检测优先级，不承诺取消即时中断
crypto、排序或分配，亦不证明所有输入由平台原子共捕获。

本层不增加 TimeProviderID == CA digest、信任域 == metadata 域等无依据
等式。时间签名只认证原 attestation transcript；未包含其中的证书摘要
不是已被时间 signer 签名的字段。新增关联只是本次 owned 输入的组合关系。

## 7. 时间窗与条件上限

Now 唯一取受核对的 Time.Policy.Time.Now。必要有效值要求：Now 与
Time.ValidUntil 非零，UTC 年在 1–9999，且 Now 严格早于 Time.ValidUntil。
CA/leaf 窗口端点的 UTC 年也在 1–9999，起点严格早于终点；沿用旧证书
规则，不因合法端点恰为 0001-01-01T00:00:00Z、与 Go 零时刻同值而另加
IsZero 拒绝。真实下层应已保证这些条件；矛盾摘要属于 ErrInvalidEvidence，
不做补救，也不把已解析的完整证书时间误当成缺失的观测字段。

CA 和叶证书分别使用项目本地半开规则 `NotBefore <= Now < NotAfter`：
Now.Before(NotBefore) → ErrNotYetValid；否则 !Now.Before(NotAfter) →
ErrExpired。先 CA 后 leaf；不强制转换跨包状态枚举或使用 Unix 秒截断。
生效时刻恰等允许；到期时刻恰等拒绝。前后 1 ns 有真实区别。

```text
ObservedAt = Time.Policy.Time.Now
ValidUntil = min(Time.ValidUntil,
                 Certificate.CA.NotAfter,
                 Certificate.Leaf.Leaf.NotAfter)
```

同值最小项不另分优先级，返回同一 instant。成功要求 ObservedAt <
ValidUntil；并不意味着该结果可缓存到该时刻、可跳过新的观测/状态检查、
可延长连接或可运行。新的 Now/观测身份必须重新 Evaluate。

不增加 clock-skew allowance，不借 attestation 的旧 freshness 规则给
证书多加 5 分钟；不验证 issuer_now 或 NotBefore=issuer_now-5m。
不要求 leaf 的整个 24 小时窗口包含在 CA 窗口内。只要在本次 Now 两者
各自有效，即使 leaf 早于 CA 生效或晚于 CA 到期仍可得到该部分结果；
ValidUntil 受较早到期限制，完整签发模板/窗口策略责任仍未完成。

## 8. 错误合同

新增恰六个固定 sentinel，不包含材料字节、外部 error 文本或敏感参数：

| 名称 | 固定 Error() 文字与范围 |
| --- | --- |
| ErrInvalidInput | `nodecerttime: invalid input`；nil context 或预复制必需非空字段缺失 |
| ErrLimitExceeded | `nodecerttime: limit exceeded`；任一容量或累计预算超限 |
| ErrInvalidEvidence | `nodecerttime: invalid evidence`；下层成功却 nil/错误 proof/status，或第 7 节矛盾必要值 |
| ErrBindingMismatch | `nodecerttime: binding mismatch`；第 6 节的跨输入节点/时间快照关系不一致 |
| ErrNotYetValid | `nodecerttime: certificate not yet valid`；按 CA→leaf 顺序命中的尚未生效 |
| ErrExpired | `nodecerttime: certificate expired`；按 CA→leaf 顺序命中的已到期 |

两个真实下层的非 nil 错误直接保留，不包装、换文字、合并成上表 sentinel，
不丢失 statetimebinding.PolicyFailure 的原 Code/Reason/错误链。
不接受可替换判定器，不新增未知错误回调或人为生成下层错误的产品 hook。
context 的 Canceled/DeadlineExceeded 检测仍优先，成功路径最终复查。

组合失败次序：所有预算超限先于本层必需非空错误；Certificate 判定先于
Time 判定；两者成功才核对本层关联；关联先于证书时间窗；CA 窗口先于
leaf 窗口。例：CA 已到期、leaf 尚未生效时返回 ErrExpired；CA 尚未
生效、leaf 已到期时返回 ErrNotYetValid。原下层内部的错误顺序不改。

## 9. 隔离与准许修改范围

未来实现唯一新源码目录为 `internal/nodecontrol/nodecerttime/`。唯一
旧源码修改是下列八份 boundary_test.go 的新消费例外及相应正反例：

| 旧文件（internal/nodecontrol 下） | 仅新增的消费关系 |
| --- | --- |
| nodeissuer/boundary_test.go | exact nodeissuer import → 新包；保留旧本包测试例外 |
| statetimebinding/boundary_test.go | exact statetimebinding import → 新包 |
| statepolicybinding/boundary_test.go | 仅为完整输入/摘要值复制使用 exact statepolicybinding |
| statebinding/boundary_test.go | 仅为完整值复制使用 exact statebinding |
| statecontract/boundary_test.go | 仅为完整十二域值复制使用 exact statecontract |
| statematerials/boundary_test.go | 仅为表/期望复制及预算使用 exact statematerials |
| statepolicy/boundary_test.go | 仅为完整 Old/History 值复制/清理使用 exact statepolicy |
| stateverify/boundary_test.go | 仅其对 statepolicy 的消费判断；不允许新包 import stateverify |

每项新例外同时要求：目标 import 完全相等；路径规范化后在当前仓库根
下的直接目录 internal/nodecontrol/nodecerttime；package 恰为 nodecerttime。
允许其直接产品文件及同包测试，不允许 nested、相似前缀、另一个根的
相同后缀、外部路径、正确目录错误 package 或 nodecerttime_test。
原例外及测试保持原样；不通过一般“同名包”或宽泛前缀开放能力。旧扫描
对其他消费者/测试的既有范围不作为新包自行放宽的理由。

新包 capability/消费者守卫须与这些消费者例外配套：标准库仅
context.Context、errors.New、bytes.Clone/Compare、sort.Slice、time.Time；
时间方法仅 Round(0)/UTC/IsZero/Year/Before，UUID.String 仅用于第 6 节
节点规范字符串比较，Clear 仅作用于已登记的自有 owner/值；
两下层仅 Input/Summary/Model 或 Material 类型、规定 proof/status、Evaluate/
Summary/Clear；其余上表依赖仅限定值类型/预算常量与自有值清理。嵌套
未知类型用有界泛型复制，不直接 import stateformat/statecomposition/
stateverify/statetime/nodeca/nodeleaf/nodeleafkeys；不新增该七包的旧例外。
实施计划须给出实际使用成员的精确白名单和文件/API 库存，不能默认开放
整包能力。只有本层拥有的 Old 等值可清理，不能经类型依赖调用判定器。

禁止 crypto/digest、DER/JSON/JCS parser 或重编码、动态反射、系统时钟、
timer/后台任务、随机、provider、数据库、文件/环境/进程、网络、unsafe、
goroutine、recover、dot/blank/重复 import、借别名或函数值绕过判定入口。
新包自身仅允许 exact 同目录/package 的测试消费者；任何产品消费者、
外部测试包、子路径 import 或 sibling 测试均拒绝。

八份旧测试以外，旧产品、API、Proof/责任表、现有测试/向量、SQL、生成
文件、模块/配置及已完成进度记录不变。若实现发现需要超出该范围，须
停止并提出书面修订，不把这一规格解释为自动批准扩大改动。

## 10. 实施后的测试与有限验证

实施计划须先测试再实现。组合负例应有真实下层可分别成立的正控制；
不能因无效证书签名或过期 time attestation 提前拒绝而误称覆盖本层边界。
使用离线固定签名材料、固定测试 key 和显式 Now，不使用系统时钟或
产品测试 hook。新增固定材料/预期摘要在专属测试/证据范围内产生和核对；
不修改旧 fixture，不把一次性生成器放进产品包，也不在普通测试随机重签。

应覆盖：

1. 真实 nodeissuer 和 statetimebinding 单独通过及联合通过；下层失败原
   错误/PolicyFailure、无观测、rollback、nonce/metadata/签名错误保持。
2. 两张证书 NotBefore/NotAfter 的恰等及前后 1 ns；对应时间证据在该
   instant 有效的控制；合法 year-1 证书端点与缺失 Now/期限的区别；
   CA→leaf 双重故障顺序和所有阶段错误优先级。
3. CA/leaf/time 三种最早到期及并列；Now 有效但 leaf 窗口不包含于 CA
   的正例；CA 自身签名/来源未认证等既有限制不被悄悄补齐。
4. 两下层分别有效但 node 不同的组合拒绝；provider/ObservationID/Now/
   Floor 不一致由真实时间下层拒绝；同 instant 的既有 UTC 规范化接受，
   非零纳秒保留；1 ns 的新观测须新 Evaluate，不承诺旧结果新鲜。
5. 每项预算等长/+1、总量等长/+1、重复 raw 按出现次数计数；不把容量
   正控制扩称下层语义接受。未知 kind、缺失字段保持规定分类及顺序。
6. 全嵌套输入克隆、原输入不变、别名隔离；nil/零值/包装拷贝/重复 Clear、
   独立 Evaluate、独立 Summary 集合；可达 backing store 清零。
7. 复制/下层返回/捕获可变摘要/最终发布处的取消与正常 panic 清理；源码
   清理登记守卫与可观察运行测试分开。不伪称测得不可返回 owner 的内存。
8. 原 Proof/Pending/Obligations 全部保留；条件证据不绑定 Identity 引用、
   不证明 CA trust/domain/epoch 来源、无运行消费者的正反控制。
9. 八份旧消费者例外的相对/绝对路径正例及全部拒绝控制；新包 import/
   API/成员/库存/所有权/消费者守卫；stateverify import 显式仍拒绝。

未来最小局部矩阵为 23 个独立命令：nodecerttime 加当前 18 个完整包
nodeissuer/nodeca/nodeleaf/nodeleafkeys/contracts/serving/statebinding/
statecomposition/statecontract/stateformat/statematerials/statepolicy/
statepolicybinding/staterecovery/statetime/statetimebinding/stateverify/readiness；
authority 仅 TestBoundRead/TestCoordinator；store 仅编译、不运行测试；
两项既有 nodeissuer envelope/pair 短 fuzz。格式检查另记。
沿用已缓存 Go 1.26.5、离线只读模块、PowerShell；完整包 -count=1、
-p=1、-timeout=120s；既有 fuzz 配置 5s、单并行、timeout=30s。
实施计划再冻结完整命令和证据规则，实际 executions/失败/取消/源码
前后 hash 如实记录；本轮没有执行这些命令，也没有通过声明。

未来 fresh 只读增量质量/安全审查仍需完成。上述限定矩阵、短 fuzz、AST
控制和单次审查均不是形式/穷尽证明，不代替整仓、race/Linux、真实 DB/
Docker、TLS/provider、硬件/reboot/PITR 或 C12 权威验收。

## 11. 本轮交付与下一门槛

本轮仅新增本文，作者自检应核对：完整预算算术、两次真实判定、唯一
Now 与纳秒规范化、跨节点与观察绑定、CA→leaf 半开窗、三者最小上限、
原错误及责任保留、可变摘要/共享 owner 清理、八处旧测试例外、无运行
能力、23 项未来有限矩阵及全部未认证责任。文档自检不是独立源码审查。

brainstorming 的默认文档提交步骤不在本轮明确的仅起草范围内；本文
尚未暂存/提交/推送，仅本机保存，不是换电脑可恢复的远端检查点。
不修改旧设计/计划/路线图以预告尚未发生的实现、验证或保存成功。

用户审阅并批准本文后，下一步仅用 writing-plans 编写实施计划。计划须
再审阅并选择执行方法，之后才实施；本轮确认不能替代任何后续批准。
Git 保存、推送、main 合并及整体验收仍各受其原授权和门槛约束。
