# Task 10：隔离的 desired 签名关系验证设计

日期：2026-10-09。

状态：用户已确认本文书面设计；对应实施计划已编写，仍待用户审阅批准。
设计批准不等于实施许可。目前只更新文档，不改产品代码、不运行测试、
不暂存/提交/推送。计划获批后才实施，继续使用现有 coordinator worktree
和 Windows PowerShell。

## 1. 目标及证据边界

上一阶段已经证明字节格式与私有数据库投影一致，但非零假签名仍可通过
格式校验。本阶段增加独立的密码学检查，回答一个有限问题：

> 相对于本次明确指定的初始根，这组原始制品是否具有连续的根签名关系、
> 足够的 metadata 根签名，以及与所选 metadata 密钥相符的 desired 签名？

成功只返回 SignatureEvidence，不返回 Authorized、TrustedState 或可供
生产 reader 消费的授权能力。即使攻击者能构造一整套自签数据，结果也
只能说明它相对于攻击者提供的根成立，不能证明根来源可信。

本层不证明可信部署、当前时间有效、最高已接受版本、防回滚、事件真实性、
库存完整、fence-finalized provenance 或当前服务权限。没有生产入口、
root 安装、持久化高水位、signer 调用或私钥配置。

## 2. 依据、现状与方案取舍

依据为既有 C1.2 基础设计 §7.1、§7.4、§12，以及局部开发格式附录
`2026-10-09-node-state-format-contract-addendum.md`。后者的精确 wire
字段、字符串整数、原始 payload/envelope/transcript 身份和数量/字节
限制保持不变；本文不登记生产协议，也不改 c12-spec-set 或基础规范。

已核对现有 stateformat 的 artifact/root/metadata/envelope 实现：根公钥、
阈值、metadata key snapshot 和 signature shares 目前只存在于私有 wire
DTO，ParsedArtifact 尚未对外提供它们。新验签层不能另写 JSON schema
或绕过格式层抽取字段，因此需要有限、只读、可清理的材料副本接口。

采用独立 `internal/nodecontrol/stateverify`，消费 stateformat；不把验签
放进格式包，也不复制 C1.1 trust 工作流。相较一次接通可信时间/部署/生产
reader，此方案能先形成可测试的签名关系证据，代价是不能单独用于生产
放行。只修复上一轮小改进不能交付验签，因此不作为本阶段主路线。

依赖方向固定：stateverify → stateformat → contracts。stateverify 可用
标准库 crypto/ed25519，不依赖 serving、authority、store、trust、网络、
数据库、时间源或外部签名服务。不接受可替换 verifier 回调、trusted bool
或运行时算法协商。固定工具链实现普通 Ed25519 对完整 transcript 的
验证，不切换到 Ed25519ph/context 模式，不把 payload hash 当签名消息。

## 3. 输入契约

单一批次输入为原始 envelope bytes，而非调用方组装的解析材料：

| 输入 | 契约 |
| --- | --- |
| ExpectedNodeID | 独立提供的 canonical 非 nil RFC4122 variant UUID |
| ExpectedEpoch | 独立提供的正 bigint 范围 epoch |
| AnchorEnvelope | 独立提供的 InitialRoot envelope，固定 root_version=1 |
| Rotations | 按顺序提供的 0..32 个 RootRotation envelope |
| ActiveRootEnvelope | 无轮换时逐字等于 anchor；否则逐字等于最后一个 rotation |
| MetadataEnvelope | 一份完整 Metadata envelope |
| DesiredEnvelope | 一份完整 Desired envelope |

ExpectedNodeID/ExpectedEpoch 是本次比较目标，不是本层认证的身份来源。
AnchorEnvelope 不可由验证函数从 active root、metadata、TLS 或响应内
自动挑选；无 anchor 必须失败。初始根没有签名，不发明自签验证步骤。

所有输入先检查个数及每个 envelope ≤262144 bytes，再按固定 Kind 交给
stateformat.ParseEnvelope；内部 payload ≤65536 等限制由格式层执行。
32 个轮换加四个单体共最多 36 个 envelope，不增加无界流读取入口。
这些是输入上限，不是对 JSON/JCS 库峰值分配量的承诺。

调用期间输入不得被并发修改。输出不能别名输入；任何失败不返回已通过
的局部链或 metadata。不得在错误后以较短链、其他 root 或其他 schema
重试并接受同一批次。

## 4. 格式材料接口的有限扩展

stateformat 仍是唯一 wire 解析者。为自有 ParsedArtifact 增加按 Kind
区分的材料读取接口，具体 Go 类型/方法签名由实施计划冻结，内容恰为：

- Root material：版本、threshold、按既有顺序的 key ID 与 32-byte public
  key；适用于 InitialRoot 或 RootRotation 中的 new_root_set。
- Metadata material：metadata/root 版本、key ID/public key、status、
  not_before/not_after、revocation 字段和累计 revoked IDs。
- Signature material：metadata shares，或轮换的 current/new 两组 shares；
  每项 key ID 与固定 64-byte signature。Desired 继续使用已有 Bindings。

这些接口返回深副本；对象/数组有清理方法，固定长度元素按值复制；不能
返回内部 slice/map、wire DTO 或“已信任”标记。错误 Kind、nil/zero/cleared
artifact 返回固定格式错误，不把空材料当成功。原有方法行为和 wire bytes
不变。ParseEnvelope 在既有校验成功后保留所需材料的私有、规范化自有副本；
访问器只复制，不重复解析 JSON。Clear 同时清理新增材料，构造失败清理
尚未交付的自有副本，不能产生第二套字段规则。

stateverify 只从本次自己 ParseEnvelope 得到的对象读取材料；不接受调用方
构造的 material struct 作为证据。Detached 副本即使可被修改，也不能改变
artifact 或已经返回的 SignatureEvidence。

上一轮两个新 Minor 纳入未来计划的针对性任务：清理 strictCanonical
直接持有的 canonical 临时缓冲；用仍为 canonical 的 limit+1 输入隔离
大小边界测试。两项历史 fixture/ownership Minor 不自动纳入或关闭。

## 5. 签名阈值算法

所有 shares 先通过格式检查，包括长度、非零、排序和唯一性。对一个
指定 root set 与一组 shares：

1. 按 key ID 匹配该 root 的 key，禁止跨组或跨 root 借用公钥。
2. 未知 key 的结构正确 share 不计数；不能用未知 key 填满 threshold。
3. 已知 key 必须对该对象的 exact transcript 通过普通 Ed25519 验签。
   任一已知 key 验签失败使整组失败，即使其他有效份额已够 threshold。
4. 每个不同 key 最多计一次；格式层的 key ID=SHA256(public key)及唯一性
   同时防止同一物理公钥以两个 ID 重复计数。
5. 完整检查该组后，要求有效不同 key 数量至少为 threshold；不是必须
   恰好等于 threshold。不得达到阈值便跳过剩余已知份额。

轮换 current/new 两组分别按各自 root set 计数。若公钥同时存在于两代
root，它在每组各可贡献一次；不要求两代 root 的公钥集合互斥，但不允许
在单组中重复计票。不合并两组份额来满足任意一方的阈值。

这是本文明确的局部验签策略；结构合法但密码学无效的未知 key share
仍可忽略，因为它不能成为计数候选。结构损坏的未知 share 仍先被格式层
拒绝。基础格式测试中的非零假签名不再是验签正向样本。

## 6. 验证顺序和对象间约束

### 6.1 初始根与逐次轮换

解析 anchor，要求其 epoch=ExpectedEpoch、版本 1、单 key/threshold 1。
记录其 exact payload/envelope digest，作为条件性证明的根身份，而不是
deployment-approved anchor 的证明。

按输入顺序处理每次轮换：epoch 保持相等、previous 精确等于当前版本、
新版本 checked +1、rotation sequence 严格高于当前 root sequence；body
与 embedded root 的绑定已由格式层验证。以同一个原始 rotation transcript
分别满足 current 与 new threshold 后，才能推进本次内存中的 current root。
缺段、重排、同版本分叉、跨 epoch 和只传 new_root_set 均失败。

最终 ActiveRootEnvelope 必须与链尾逐字相等；无轮换时与 anchor 逐字相等。
不能只比较其中的 key set 而忽略原始发布制品身份。

### 6.2 Metadata

要求 metadata epoch/root version 等于当前 root，并用该 root 的公钥和
threshold 验证 metadata 的 exact transcript。保留完整 snapshot 与 revoked
ledger；不从旧 metadata 补入缺失 key。本批次只有一份 metadata，没有
历史高水位，因此不能宣称 metadata version/sequence 相比本地历史已增加。

不人为比较不同对象流的 creation sequence 大小来代替其各自历史检查。
轮换流内部的严格 sequence 递增仍必须执行。

### 6.3 Desired

要求 desired epoch=ExpectedEpoch，node=ExpectedNodeID。SigningKeyID 必须
精确匹配所选 metadata 中存在且不在 revoked ledger 的 key；status 只允许
active 或 retiring，revoked/缺失 key 不接受。

以匹配公钥验证 desired 的 exact transcript。检查 key.not_before ≤
desired.issued_at < key.not_after，并要求 desired.valid_until 不晚于
metadata.valid_until 和 key.not_after；24 小时本体上限仍由格式层保证。
这是静态窗口关系，不是当前时间有效性，也不重新认证存储的 EffectiveDeadline。

不要求 desired.issued_at ≥ 当前 metadata.issued_at：较新的完整 metadata
可保留一把 retiring key 来验证旧 snapshot。是否在实际签发时使用 active
key、历史窗口不可变、撤销事件真实及当前仍可使用，属于后续工作流/信任层。

Desired payload 没有签入 root/metadata version 或 envelope digest，因此
不能凭空证明它最初引用的唯一 metadata 制品。若另一份签名正确、同 epoch
且仍包含适用 key 的 metadata 也满足这些关系，本层可形成相对于该批次的
证据；不能把这种兼容替换一概叫作可被密码学检测的拼接。已有 serving 的
exact SQL/projection 绑定仍承担发布行选择；本轮不改其行为。

## 7. 可信时间和历史的隔离

本层不取 time.Now，不接受 caller-supplied now 冒充可信时间，不安装根、
推进高水位、清除 incident 或持久化信任状态。纯签名关系证据可能由目前
已过期的轮换/metadata/desired 产生；必须显式记录它不是当前可接受状态。

这不是放宽基础设计的网络根恢复规则：真正的 consumer 仍必须在可信
当前时间内检查每个 rotation 的有效期；过期链只能走受信部署恢复，不能
借本层的成功结果补链。未来生产接线须独立审阅可信时间、部署来源、
回滚保护和最终授权，不能直接把 SignatureEvidence 当作入口凭据。

## 8. 输出、错误与生命周期

成功结果为私有字段的 SignatureEvidence，记录 ExpectedNodeID/Epoch、
anchor 的 P/E digests、按顺序的 rotation P/E digests、最终 root 版本、
metadata 版本/P/E digests、desired generation/sequence/P/E digests/key ID。
这些身份固定本次实际检查的批次，不新增 wire schema 或总签名摘要。
结果不携带 private key、trusted flag、当前授权期限或可调用服务能力。

访问器返回固定字段值或 slice 副本；nil/zero/cleared 的行为由计划固定为
零值而非有效证据。成功结果支持幂等 Clear，清掉自有数组/固定字段。
不支持复制所有者结构或与 Clear 并发使用，不承诺擦除 immutable strings、
JCS/运行时内部副本或调用方已经获得的 detached 副本。

纯包使用有限 sentinel 分类：InvalidInput（含格式/参数）、BindingMismatch、
SignatureMismatch（含阈值不足）和 Canceled。名称由计划固定，无输入值、
公钥字节、JSON/库动态错误外泄。nil ctx 为 InvalidInput；已取消优先。
每个对象、每个签名验证前后和最终交付前检查取消。单次标准验签不能被
中途抢占，不虚报实时取消 SLA；份额数有界且不会启动后台验签任务。

内部 owner 在错误、取消或 panic 时清理已解析的 artifact、材料副本、
transcript 和未交付输出；错误返回 nil、无部分证据。Panic 不变成成功，
按既有惯例在清理后传播。不得加入测试观察 hook 或可关闭校验的配置。

## 9. 测试和实施计划要求

未来计划必须明确请求新 stateverify 包的离线纯单元测试许可，本轮不运行。
沿用固定工具链、缓存依赖、PowerShell，不下载新库。测试包含：

- 独立冻结的真实签名向量和原始 transcript bytes；期望不能由被测 encoder
  或 verifier 生成。额外属性用测试专用固定 seed 签名，不接触生产密钥。
- 0/1/32 条合法链；33 条、缺段、乱序、跨 epoch、不能验证原链签名的
  替换 anchor，以及不等于链尾的 active root。完整另一套自签数据相对于
  其自带独立输入根可形成条件性证据，但不获得部署信任。
- current/new 任一侧不足；两组不能互借；1..5 key/不同阈值；有效票数超过
  threshold；达到阈值后仍有已知坏签名必须失败；未知份额不计票。
- 正确格式假签名、错误 key、错误 domain、签 P hash 而非 transcript、
  未重新签名的 P/E 内容替换、不同对象类型和 C1.1 代用不能通过。
  另测 §6.3 所述兼容 metadata 替换，避免测试声称签名绑定了不存在的字段。
- metadata root/epoch 与 desired node/epoch/key 绑定；active/retiring 正向，
  revoked/缺失 key 负向；签发时间半开边界和 valid_until 上限。
- 明确正向证明：只要签名和静态关系成立，已过当前日期的制品也只能形成
  条件性 SignatureEvidence；此结果不能改变 reader 或任何信任状态。
- 格式材料/输出副本隔离、nil/zero/cleared、最终失败/取消/panic 清理，
  保留 owned aliases 并逐次变异立即断言，避免双重变异抵消。
- 依赖/调用边界：stateverify 无 DB/网络/时间源/签名服务和生产调用者；
  stateformat 不反向导入 verifier，也不新增第二套 JSON 规则。
- 上一轮两项新 Minor 的先失败后通过测试及原有格式/serving 限定回归。

测试和审查只能证明本地增量；不得运行真实 DB/Docker、C11/C12/bootstrap/
PITR/race/全仓测试，不改变安全配置，不绕过先前安全拒绝。

## 10. 变更范围、未关闭门禁与自查

未来实施允许新增 stateverify 及测试，有限扩展 stateformat 材料接口/清理
与相关测试，并更新进度文档。现有 wire bytes/独立格式向量不变；serving
reader/adapter/binder、authority、SQL/schema/generated store、API、四锁、
基础规范、B02 原计划和批准集均保持不变。两项历史 Minor 继续开放。

部署兼容性仍 UNKNOWN；本文及未来实现均仅是局部开发成果，不能发布
旧签名的新格式替代物，不能声称 Task 10/B01/C12/main/security 验收完成。

自查：唯一 wire 所有者、独立初始根参数、精确域/字节、双阈值、无部分
交付、条件性证据与时间/历史隔离均已明确。特别消除了“数据库初始根等于
可信根”“验签通过等于当前有效”“desired 唯一签入 metadata 身份”三种
错误推论。具体 Go 签名、文件拆分、测试命令与执行方法交由下一份实施
计划逐项冻结；本文审阅通过不自动授权实现或提交推送。
