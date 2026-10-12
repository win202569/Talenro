# Task 10：期望状态读取设计——第一阶段暂存投影

日期：2026-10-08。状态：用户已确认方案和计划，第一阶段限定实现已完成；证据见执行记录。
这是跨组件工作的第一阶段设计，不是完整 DesiredStateFacts 或生产读取器设计。
本文件为设计依据；不授权暂存、提交或推送，后续阶段仍需单独设计批准。

## 目标和依据

最终目标沿用 Batch 01 Task 10：从同一 inventory 行的 active desired/root/
metadata 指针开始，在同一绑定连接内取得精确已提交制品；前后 authority
检查通过才发布不可变 facts。B03 响应构造阶段不得补查数据库。

依据：

- `docs/superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md`，Task 10。
- `docs/superpowers/specs/2026-08-23-node-pop-control-plane-design.md`，7.1–7.2。
- `db/migrations/00006_nodecontrol.sql` 中 inventory、state signing intents、root/metadata publish intents、desired states 的字段和约束。
- 已实现的 serving guardedRead、私有 certificate reader 与 authority BoundReadEpoch。

迁移存储 root/metadata 的 published_envelope；desired 存储 canonical_payload、
payload_digest、signature 和 signing_key_id，并没有等价的完整 signed-envelope 列。
因此不能把 desired payload 直接命名为完整签名 envelope，或自行编造序列化格式。
本次源码检索未找到 NodeStateRootRotation/DesiredStateSnapshot 的可复用业务解析器；
它们不是现有 authority-v7 外部证据 envelope 的同义类型。

## 阶段选择

推荐先做私有结构投影与纯校验，再做精确 SQL/行转换，最后做规范 envelope
处理与私有读取组合。优点是每阶段能离线验证，缺失密码学/来源证据不会被
混入“已验证 facts”。代价是第一阶段不能供 B03 直接消费。

另一个方案是一次实现查询、解析器、签名信任和 reader，涉及更多尚未核对的
业务契约；第三个方案仅返回原始 bytes，会丢失 active 指针与制品间一致性。
本方案采用第一种，不新增通用制品框架，也不改现有证书读取流程。

## 第一阶段范围

仅在 serving 内新增私有 desired 投影类型、纯校验构造函数、深拷贝与清理，
不导出 `DesiredStateFacts` 或 `Reader.GetActiveDesiredState`。
建议独立文件 `desired_projection.go` 和 `desired_projection_test.go`。
最终公开契约名留给完成规范 envelope、精确 fence 和绑定读取的后续阶段。

投影分三种专用结构，不以自由字符串选择制品种类：

| 结构 | 必须保留的字段 |
| --- | --- |
| 根/元数据发布制品 | PublishID、Version、BaseRootVersion、BaseMetadataVersion、AuthorityOperationID、AuthorityVersion、Payload、PayloadDigest、PublishedEnvelope、EnvelopeDigest |
| desired 制品 | NodeID、Generation、SigningID、AuthorityOperationID、AuthorityVersion、InventoryVersion、ResourceEnvelopeVersion/Digest、RootPublishID/Version、MetadataPublishID/Version、SigningKeyID、Payload/PayloadDigest、Signature、IssuedAt、EffectiveDeadline、ValidUntil |
| inventory 选择与组合 | NodeID、ActiveDesiredGeneration、ActiveRootPublishID/Version、ActiveMetadataPublishID/Version、RootRotations、ActiveRoot、ActiveMetadata、Desired |

ID 使用 uuid.UUID；版本/generation 使用 uint64；摘要与 key ID 使用固定
32 字节值；签名固定 64 字节；payload/envelope 使用独占 byte slice；
AuthorityVersion 复用 contracts 类型。原始 SQL signed/null 值转换不在本阶段。

构造输入另外包含 expected node UUID 和 expected authority epoch，均只供
私有结构校验。它们并非可信能力；后续 reader 必须从 BoundReadEpoch 取得
epoch，不能将本阶段构造函数导出为可由任意调用方授权的入口。

## 纯校验规则

1. expected node 非零且为 RFC4122 variant，epoch 位于 1..MaxInt64；输入错误
   为 ErrInvalidArgument。内部投影不一致为 ErrConflict；失败返回零投影。
2. 所有 ID 非零且 variant 有效；所有正版本/generation/authority sequence
   位于 1..MaxInt64。允许为零的 base version 仅限迁移允许的 base 字段。
   所有制品 authority epoch 精确等于 expected epoch，不推导“最新 epoch”。
3. inventory node、desired node 与 expected node 一致；desired generation、
   root/metadata publish ID 和版本分别与 active 指针精确相等。
4. active root/metadata 制品 ID、版本与对应指针精确相等；元数据的
   BaseRootVersion 精确等于 active root version。root 的 BaseRootVersion
   等于 Version-1；metadata 的 BaseMetadataVersion 等于 Version-1。
5. 每个 payload 长 1..65536，每个已发布 root/metadata envelope 长
   1..262144；摘要非零并精确匹配 SHA-256。desired 签名必须非全零且
   为 64 字节，key ID 为非零 32 字节；这仅检查形状，不验证签名。
6. desired 时间非零，IssuedAt < EffectiveDeadline <= ValidUntil，且
   ValidUntil 不晚于 IssuedAt+24h。只检查已记录窗口，不以本机当前时间
   判断授权。ResourceEnvelopeVersion 和 InventoryVersion 为正值，resource
   digest 非零；不把 payload 内字段一致性视为已经检查。
7. 不接受仓储/DBTX 句柄，不修改输入；成功值及每次快照都深拷贝所有 slice，
   包括根链容器及各制品 bytes。清理仅作用于自有暂存 bytes，然后清零结构。
   不承诺不可变字符串或运行时副本的安全擦除。

## 根链：第一阶段仅检查显式区间的结构

初始可信根版本固定为 1，来自受信 host deployment，不从数据库自动信任。
RootRotations 最多 32 段，输入显式保留 ChainBaseRootVersion（1..MaxInt64）。
该值只是本次暂存链的区间起点，不是用户 high-water，也不是已认证 trust anchor。

逐段要求 BaseRootVersion 等于上一版本、Version 精确加 1、同一 epoch、
authority sequence 严格增加、PublishID 不重复。验证加法前检查 MaxInt64，
不得溢出。链尾精确等于 ActiveRoot 的 ID/版本/authority tuple/摘要/bytes。
空链仅允许 ChainBaseRootVersion 等于 active root version；它表示区间内没有
轮换，不能据此宣称节点已信任该根。缺段、重复、乱序、超过 32 均失败，
不能排序、截断或用 MAX/latest 填补。

这里刻意不决定数据库应返回从版本 1 开始的全部历史还是有明确保留边界的
窗口。冻结的最终 reader 只接 node ID，而 B03 以 client high-water 选择
连续后缀；窗口起点的可信来源和缺失历史的响应规则必须在 SQL 阶段设计中
明确批准后才能接入。第一阶段的显式区间不得绕过这个门禁。

## 测试与第一阶段完成标准

采用字面期望值及彼此不同的合法 UUID，保留有意要求相等的字段；避免已知
证书 fixture 同 ID 掩盖字段交换的问题。测试至少覆盖：

- 合法空区间、一段、32 段；33 段、重复 ID、断链、逆序、错误末端与溢出。
- node/generation/root/metadata 指针或 authority epoch 任一不匹配返回零值。
- 缺失 bytes、错误摘要、1/65536/65537 payload 和 1/262144/262145 envelope 边界。
- 无效时间顺序、24h 上限、零签名/key ID/资源摘要、版本越界。
- 构造后修改输入、修改一次快照均不改变内部值；清理自有副本不破坏输入。
- 一份摘要正确但不是合法业务 envelope 的测试数据可通过此结构层；测试名称
  与注释必须明确这不是规范解析或验签证据。

先失败后实现，记录实际 RED/GREEN；只允许运行既定 serving/contracts/
readiness 测试、限定 authority 组与 store 仅编译。不下载依赖，不改锁文件。
第一阶段的完成只表示私有结构与所有权边界可用，不表示期望状态可服务。

## 后续阶段门禁（本附录不授权实现）

SQL 阶段必须另行设计一条从 active 指针出发的组合查询，核对 exact active
signing/publish intents、各自 committed-active fence、operation kind/scope，
并区分不存在的状态与损坏/部分状态。根链窗口、历史行状态、初始根发布行的
表示须与原规范和真实 schema 一致，不得臆造缺失制品。禁止 latest/MAX/orphan fallback。

规范处理阶段必须核对 root/metadata/desired 的 canonical body 与外层绑定，
明确 desired 完整 envelope 的构造/验证；摘要相等不等于规范格式或验签。
最终组合才能复用同连接 epoch、guardedRead 与清理函数，并形成完整
DesiredStateFacts；不得用本阶段类型替代它。

生产 v7 授权、B03 时间/签名/端点策略、真实 PostgreSQL/PITR、完整 C12、
安全告警、main 合并和发布保持未验收。第一阶段方案获书面批准后编写其
实施计划；后续阶段不随此次批准自动开放。
