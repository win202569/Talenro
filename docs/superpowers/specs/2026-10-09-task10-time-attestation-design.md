# Task 10：已签名时间证据隔离验证设计

日期：2026-10-09。

状态：用户已于本会话批准本文及实施计划；本地实现、修复后限定回归
及一次独立审查完成，两项轻微测试覆盖建议待补。实证见
../../roadmap/2026-10-09-time-attestation-execution.md。
未提交、推送；生产兼容性 UNKNOWN。

## 1. 目的和依据

补齐父设计 `2026-08-23-node-pop-control-plane-design.md` §7.3 的时间证据
格式、条件性验签和 freshness 判定。用户选择暂不接入真实平台环境。
成功标准是可重复的离线规则验证，不是生产连接可接受或可信时钟建立。

复用 `stateformat` 的唯一严格 JSON/JCS 解析器，遵守
`2026-10-09-node-state-format-contract-addendum.md` 的编码与大小约束。
该旧附录明确不涵盖 time-attestation；下述新增 schema、nonce 编码和
接口是本提案的新决定，不能宣称已由旧附录批准或已部署兼容。

现有 `stateverify.Verify` 要求 root、metadata 和 desired 全套输入；
时间证据可以在没有新 desired 的响应中出现。本阶段不伪造 desired，
不放宽旧验证入口，也不为本任务重构整个根链验证器。

## 2. 选择、范围及限制

采用独立 `internal/nodecontrol/statetime` 模块，接受原始 attestation、
原始 metadata 和显式的当前信任/时间/请求/检查点模型，输出条件性证据。
相比直接接入 TPM、真实可信时间与持久化恢复，这条路线可离线推进，
但不证明模型输入来源可信。相比将逻辑塞入 statepolicy，它不要求
desired 更新，也不改变已有节点状态转换语义。

范围内：新增纯格式 Kind、独立 Ed25519 验签、字段绑定、时间窗口与
节点检查点比较、重检、所有权清理和隔离边界测试。

范围外：根链安装/更新、metadata 信任建立、时间提供方、系统时钟、
TLS 实现、请求发送、随机数生成、真实单次消费、持久化写入/恢复、
硬件计数器、隔离操作、节点启动、生产调用方和签名服务。
不新增依赖、数据库字段、OpenAPI 或外部网络协议发布。

任何成功返回都不产生 production capability。调用方填写 Available、
摘要或来源 ID 不能证明平台认证；测试模型不能直接注册为运行时适配器。

## 3. 格式扩展

在 stateformat.Kind 末尾追加 TimeAttestation，保持已有值与接受语义。
解析器继续拒绝未知字段、重复字段、非 canonical JSON、尾随内容、
非规范整数/时间/base64url 及超限输入。payload 上限 64 KiB，envelope
上限 256 KiB；不扩宽共享限制。

内层 payload 字段恰为：

```text
schema_version = "node-time-attestation.v1"
control_plane_authority_epoch: U
node_authority_checkpoint_sequence: U
node_id: UUID
request_nonce: N
issued_at: T
valid_until: T
```

U、UUID、T 沿用旧附录。N 为 43 字符 raw unpadded base64url，严格解码
为 32 bytes，重编码必须逐字相等。nonce 不复用公钥非零限制：全零的
32 bytes 在格式上仍合法，格式校验不证明任何值具有随机性。
checkpoint 使用 1..MaxInt64；尚无可信本地 checkpoint 不能用零隐式初始化。

外层字段恰为：

```text
schema_version = "talenro-node-time-attestation/v1"
payload_b64url: 原始内层 JCS bytes 的 raw unpadded base64url
payload_sha256: D
key_id: D
signature: S
```

签名消息严格为普通 Ed25519 的：

```text
TALENRO-NODE-TIME-ATTESTATION-V1 || 单个 NUL byte || 原始 payload bytes
```

不对重编码后的内容验签，不采用 Ed25519ph，不用 payload hash 替代消息。
payload/envelope 分别计算身份摘要，沿用 P/E 区分；key_id 由签名选择并
绑定 metadata 内的同名公钥，不从未验证字段建立信任。

格式层检查 `issued_at < valid_until` 且窗口不超过 5 分钟，不查询 now。
Bindings 新增独立 CheckpointSequence 和 RequestNonce[32] 字段，不能
借用 Sequence 混淆各对象 creation sequence。新增 transcript 分支，
不在格式层执行验签。Clear 清除新增字段，访问器仍返回独立值/副本。

## 4. 当前 metadata 的条件性信任边界

statetime 不接受一个裸公钥或“验签已通过”布尔值。它解析原始 metadata，
并要求其 P/E 身份、版本、epoch 与显式 CurrentMetadataObservation 精确
一致。观测还携带状态、来源 ID、观测 ID、节点 ID、已提交状态 ID。
非 Available、缺失/零身份、节点或 epoch 不一致均不产生结果。

这一精确绑定只证明“所用 bytes 正是调用方指定的 metadata 模型”，
不证明 metadata 的根签名已验证、已持久提交或没有被平台回滚。
statetime 不声称替代 stateverify/statepolicy。未来可信适配器必须把
本观测绑定到已认证、已提交的当前 metadata 及一致状态快照；不能
把 statepolicy.Candidate.State() 的 Unknown 拟议状态冒充此来源。

本阶段允许测试代码显式构造 Available 模型来检验规则；模型伪造不在
这一纯核心内可检测。该限制必须写入包文档、入口注释和执行记录。

## 5. 显式模型输入及检查顺序

建议入口为 EvaluateModel(ctx, input)，返回私有字段的 EvidenceModel。
input 包括两份 raw envelope 及以下固定值观测，不依赖 statepolicy
包，不增开该包的生产导入许可：

- 当前 metadata 观测：见 §4。
- 时间观测：Available 状态、ProviderID、ObservationID、Now、Floor。
- 请求观测：Available 状态、RequestID、ObservationID、NodeID、Epoch、
  ExpectedNonce[32]；表示待处理请求模型，而非真实 nonce 消费凭据。
- 本地 checkpoint 观测：Available 状态、SourceID、ObservationID、
  StateID、NodeID、Epoch、正 checkpoint，且 StateID 与 metadata 一致。

所有身份摘要均非零；节点、epoch 与签名 payload 一致。时间使用有限
有效范围内的显式值，Now >= Floor。不调用 time.Now 或读取环境变量。
零值、Unavailable、Rollback、未知状态、nil context 均明确拒绝；
预取消优先返回取消，不执行密码计算或返回部分结果。

顺序为：有界输入/观测检查 → 唯一 parser → 精确绑定 → 密钥授权与
签名 → 时间与 checkpoint 规则 → 返回完整独立结果。

密钥必须存在于当前 metadata，状态 active 或 retiring，且不在撤销
账本；revoked、缺失或未知 key_id 拒绝。metadata 当前有效；attestation
的 issued_at 和当前 Now 均不得早于 key.NotBefore，均早于 key.NotAfter；
valid_until 不得超过 metadata.ValidUntil 或 key.NotAfter。

## 6. freshness 与 checkpoint 规则

- 使用同一 Now 检查 metadata 和 attestation 的 `[issued_at, valid_until)`。
- attestation 的 `abs(issued_at - Now) <= 5m`，含恰好 5 分钟边界。
  即使通过偏差检查，issued_at > Now 仍因半开有效窗口拒绝。
- attestation checkpoint 小于本地值：回退错误；相等：无推进；
  大于：返回拟议 checkpoint 和 AdvanceRequired 标记。
- checkpoint 不与 root/metadata/desired 的 creation sequence 横向比较。
- 相同 checkpoint 的合法新请求可有新 nonce、新签名和新有效期，
  不视为同序不同内容冲突；相同 checkpoint 不建议计数器推进。
- 不写 checkpoint，不 reserve authority sequence，不更新 Now/Floor，
  不从连续的未来签名时间推进可信时间。

结果包含依赖的 P/E 身份、节点/epoch、key_id、请求 ID/nonce、观测身份、
旧/拟议 checkpoint 及 ValidUntil。仅表示在这些显式假设下规则通过，
不是“server response 已经认证”或“nonce 已消费”。

## 7. 重检、消费和生产集成边界

RecheckModel(ctx, evidence, input) 重新解析及验签原始输入，再执行全部
规则；除时间观测可刷新外，原始 P/E、metadata/request/checkpoint
观测及依赖身份必须与初次结果精确相同。时间提供方不能更换，Now 和
Floor 不能下降，新 Now 必须仍在有效范围。新时间观测 ID 可变化。
无效、零值、Clear 后证据拒绝；不把摘要一致当作无需重新验签的理由。

纯函数重复接收同一个未消费请求模型仍可能成功，这不是重放防护。
未来运行时必须：使用同一真实 TrustedTimeSource 验证 TLS；保证
metadata/checkpoint 一致快照；从加密随机源生成 nonce；将真实请求
与该连接/响应绑定；在最终重检和接受时原子消费请求，并处理并发重复
响应；推进 checkpoint 时执行受保护持久提交。所有这些是独立未完成门禁。
本阶段不增加 TLSValidated/NonceConsumed 布尔值来假装满足它们。

## 8. 错误、所有权与依赖边界

固定错误类别区分：格式/输入错误、绑定错误、签名错误、未授权密钥、
缺少保障观测、无效时间/过期、checkpoint 回退、证据陈旧、取消。
错误字符串不包含 raw payload、nonce、签名或完整观测。

调用者不得并发修改输入或执行 Clear。结果不借用输入可变切片；
失败/取消清除可达的自有临时缓冲并返回 nil 结果。Clear 可重复且 nil
安全；不声称清除运行时内部副本或支持并发清理。

statetime 仅依赖 stateformat、contracts 及明确允许的纯标准库/既有
UUID 库。格式层仍禁止密码验证；新包禁止网络、存储、环境、真实时钟
和进程入口。AST 测试禁止任何非测试生产包导入 statetime；无运行时
注册或默认适配器。旧 stateverify/statepolicy 的产品语义不改变。

## 9. 测试与验收范围

测试采用离线固定密钥、独立签名消息向量和显式模型，不访问真实服务。
必须覆盖：

1. 严格字段集、编码、精确大小、时间格式、nonce 长度/填充/非规范
   编码/全零合法值、Kind 错配及 transcript 的域和单 NUL 边界。
2. 正确签名、错误域/密钥/签名、payload 篡改、metadata 身份替换、
   错误节点/epoch、未知/撤销/缺失密钥及 active/retiring 正例。
3. 有效期 5 分钟及越界，Now 起止边界，时间偏差边界，metadata/key
   截断约束，时间不可用/回退/低于 Floor，远端时间不能改变输入时钟。
4. checkpoint 较低/相同/较高、不同节点隔离、不与 creation sequence
   混用、同 checkpoint 新 nonce 可成功但不要求推进。
5. nonce 不匹配拒绝；同一未消费模型重复调用可成功的显式限制测试；
   改变 RequestID/旧状态/metadata 后重检拒绝，刷新时间后重检及过期。
6. nil/零值/取消、错误无数据泄漏、输入和输出所有权、重复 Clear、
   无结果残留、依赖/无生产消费者的正反边界测试。

后续实施计划沿用 Native 主代理 TDD。最终限定回归覆盖 statetime、
stateformat、stateverify、statepolicy、serving、contracts、readiness；
authority 仅 TestBoundRead/TestCoordinator，store 仅编译不执行测试。
禁止自动扩大至真实 DB/Docker、整仓、C11/C12、bootstrap、PITR 或 race。
测试和最终独立审查尚未执行；不得将本文当作通过记录。

## 10. 后续门禁

本文待用户审阅批准后，才编写实施计划；计划获批后才开始代码实现。
沿用用户此前选择的 Native 执行方式，不自动开始提交/推送。
由于当前只批准设计工作，不执行技能模板中的自动提交步骤。
旧签名/信任策略成果与所有无关修改保持不变。

本增量完成后仍不能宣布完整 Task 10/B01/C12、生产安全或 main 验收。
真实平台来源、TLS、单次消费、持久 checkpoint 与部署兼容性必须在
后续独立工作中取得实证，不能以本阶段模型测试替代。
