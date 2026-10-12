# C1.2 node-state 格式契约附录（局部开发设计已批准）

日期：2026-10-09。

状态：用户已于本会话批准本文及配套设计，允许编写局部实施计划；计划尚待审阅。
本文给出 canonical-binding 设计 G1–G5 的明确建议，不修改既有批准集、
源码、数据库、OpenAPI 或依赖。不暂存、提交或推送。

## 1. 范围、权威和兼容性

配套设计：`2026-10-09-task10-desired-canonical-binding-design.md`。
本附录只定义 desired、初始 root、root rotation、metadata 的纯格式及其
与现有投影的绑定；不定义 recovery/time-attestation 的新格式，不实现
签名验证、信任锚认证、可信时间、数据库写入或生产读取入口。

以下新增 schema、字段名、排序和编码均是本提案的设计决定，不声称早期
规范已经逐字规定它们。获批后才可作为本局部格式阶段的实施依据。
生产规范登记及 B02/B03/OpenAPI 消费方同步仍需单独纳入受审查变更；
不能直接改动带 hash 的旧规范而保持批准集 hash 不变。

兼容性硬门禁：实施前只读查明是否存在本组制品的持久/共享部署或外部
消费方。仓库中没有 state 包不等于没有部署。若无法取得权威发布记录，
只允许隔离的私有 codec/测试，禁止写入持久库、网络发布或迁移旧签名。
若已发布不同 wire bytes，必须另行版本化/迁移设计，不复用本提案的 v1
标识覆盖旧数据，不接受双格式 fallback。不能重编码已签名 payload 后
沿用旧签名或摘要。

## 2. 五项选择及替代方案

| 缺口 | 推荐决定 | 未选方案与代价 |
| --- | --- | --- |
| G1 封装 | 外层 JCS JSON，payload_b64url 携带原始内层 JCS bytes | 内嵌对象需重复抽取规范化；JSON 字符串转义有额外膨胀 |
| G2 整数 | 本文格式中所有整数均为无前导零十进制字符串 | 限制到安全整数会缩窄现有 bigint 范围；自制大整数 JCS 不再是标准 JCS |
| G3 大小 | payload 64 KiB、envelope 256 KiB；专用受限 codec | 不提高共享解码器默认限制；不静默收缩已允许的存储上限 |
| G4 身份 | payload、完整封装和签名 transcript 分别定义 | 不使用同一个含糊的 content digest 名称指代三者 |
| G5 归属 | 新纯包 internal/nodecontrol/stateformat 独占本组格式 | 不让 serving 依赖 state workflow，不在 writer/reader 各复制一套 |

整数选择依据 RFC 8785 §3.1 对超过 IEEE-754 精度数据使用字符串的建议：
[RFC 8785](https://www.rfc-editor.org/rfc/rfc8785.html#section-3.1)。
这是本协议的编码选择，不是声称 RFC 要求所有整数都必须使用字符串。

## 3. 通用严格规则

所有本文对象字段集闭合：下面列出的字段全部 required；未明确允许 null
的字段不允许 null；不允许 unknown、重复（含转义后同名）、大小写别名、
尾随值、非法 UTF-8 或孤立 surrogate。数组不可为 null。对象/数组嵌套
深度最多 16，根容器计第 1 层；payload 独立解码重新计深度。

输入必须逐字等于其 RFC 8785 JCS 序列化结果；禁止 trim、Unicode
normalize、补字段、自动排序后接受、宽松 base64 或数值类型转换。
先限制输入长度，再解析；不能先把无限流读入内存。

基础类型：

- `U`：JSON string，解码字符匹配 `[1-9][0-9]{0,18}`，精确整数范围
  1..9223372036854775807；不接受 JSON number、符号、小数、指数、空白。
  解析/比较使用精确整数，不经 float64。
- `D`：64 个 lowercase hex 字符，解码为非全零 32 bytes。
- `K`：raw unpadded base64url，43 字符，精确 32 bytes；重编码逐字相等，
  拒绝全零 public key。这里只检查结构，不证明有效密钥或信任。
- `S`：raw unpadded base64url，86 字符，精确 64 bytes；重编码逐字相等，
  拒绝全零值以保持现有投影约束；非零假签名仍不等于验签成功。
- `T`：20 字符 `YYYY-MM-DDTHH:MM:SSZ`，有效公历日期、年份 0001..9999，
  秒 00..59；不接受偏移、闰秒、小数秒或零时间哨兵。
- `UUID`：lowercase hyphenated RFC4122 variant UUID，非 nil；不额外缩窄
  UUID version nibble。API 若有更窄限制，由独立入口契约执行。
- `ID64`：1..64 ASCII 字符，`[A-Za-z0-9][A-Za-z0-9._-]*`。
- `ID128`：同一字符规则，长度 1..128。
- `Build`：1..64 字符，精确三个以点分隔的非负十进制整数；各分量
  0..2147483647，无前导零（单个 0 除外），不含 v 前缀、预发布或 build 后缀。
  这是新增收窄提案，必须与 B02/B03 支持的发布版本格式同步，不能自动生效。

所有 U 字段在 Go/数据库中仍可保留整数类型；wire DTO 使用显式严格编码，
不得使用默认 uint64 JSON number 序列化。该选择仅适用于本文签名制品，
不修改外部普通 API 的 PositiveInt64 或既有 authority-v7 transcript。

## 4. 内层 payload 的精确字段集

以下 `{field:type,...}` 是字段清单，不是可被直接签名的示例 JSON。
schema_version 是固定字符串，不是数值 U。除 revocation 分支明确说明
外，不存在可选字段。

### 4.1 NodeDesiredStateV1

字段恰为：

```text
schema_version = "node-desired-state.v1"
control_plane_authority_epoch: U
authority_sequence: U
node_id: UUID
generation: U
inventory_version: U
node_resource_envelope_version: U
node_resource_envelope_digest: D
issued_at: T
valid_until: T
minimum_agent_version: Build
minimum_supervisor_version: Build
reason: initial | operator_update | drain | resume | lease_refresh |
        clear_slot_quarantine | restore_reauthorize
processes: ProcessSpec[0..8]
```

`issued_at < valid_until <= issued_at+24h`，checked time arithmetic。
processes 按 slot_id ASCII bytewise 升序且唯一；允许空数组只表达格式，
不证明 enabled node 的库存完整性。reason 不可因旧 JSON 示例遗漏而省略。

ProcessSpec 恰为 `{slot_id:ID64,adapter,lifecycle,release_id:ID128,required:bool,
test_profile:ID128,capacity_profile_id:ID128,capacity_profile_version:U,
required_metrics:Metric[1..64],capacity_limits:CapacityLimits,
probe_profile:ID128,restart_profile:ID128}`。
adapter=`fixture|xray|sing_box`；lifecycle=`running|stopped|draining`。
Metric 为 1..64 字符 `[a-z][a-z0-9_]*`，排序唯一；具体 supported 集合
必须由 B02/B03 已批准的 release capability 判断，格式通过不表示支持。
以上 profile/metric 长度与字符选择是本附录显式提案。

CapacityLimits 恰含下列十个 U 字段，其数值范围保持基础规范：

| 字段 | 最小值 | 最大值 |
| --- | --- | --- |
| egress_limit_bps | 1000000 | 1000000000000 |
| connection_limit | 1 | 10000000 |
| handshake_limit_per_second | 1 | 1000000 |
| cpu_quota_millicores | 100 | 64000 |
| cpu_limit_basis_points | 1 | 10000 |
| memory_limit_bytes | 67108864 | 1099511627776 |
| task_limit | 32 | 4096 |
| file_descriptor_limit | 64 | 1000000 |
| queue_limit | 1 | 1000000 |
| packet_loss_limit_basis_points | 1 | 10000 |

没有路径、命令、任意 argv、环境、URL、原始配置或动态 map 字段。
库存逐项相等、资源总和、profile/release 可用性不属于纯格式证明。

### 4.2 NodeStateRootSetV1

恰为 `{schema_version:"node-state-root-set.v1",control_plane_authority_epoch:U,
authority_sequence:U,root_version:U,threshold:U,keys:RootKey[1..5]}`。
RootKey 恰为 `{key_id:D,algorithm:"ed25519",public_key:K}`。
keys 按 key_id bytewise 升序唯一；每项 key_id 必须等于 SHA-256(raw public key)
的 lowercase hex，因而相同物理 public key 不能以不同 ID 重复。
`1 <= threshold <= len(keys)`。无 expiry 字段。
初始 root_version=1 时额外要求一把 key、threshold=1；这仍非可信锚认证。

### 4.3 NodeStateRootRotationBodyV1

恰为 `{schema_version:"node-state-root-rotation.v1",
control_plane_authority_epoch:U,authority_sequence:U,previous_root_version:U,
new_root_set:NodeStateRootSetV1,issued_at:T,valid_until:T}`。
body 与 new_root_set 的 epoch/sequence 精确相等；新版本是旧版本 checked +1；
`issued_at < valid_until <= issued_at+7d`。不允许 0→1 轮换。

### 4.4 NodeStateTrustMetadataV1

恰为 `{schema_version:"node-state-trust-metadata.v1",
control_plane_authority_epoch:U,authority_sequence:U,root_version:U,
metadata_version:U,issued_at:T,valid_until:T,role:"node_state",next_refresh_at:T,
revoked_key_ids:D[0..4096],keys:MetadataKey[1..64]}`。
`issued_at < valid_until <= issued_at+7d`，
`issued_at <= next_refresh_at <= valid_until-48h`；因此本提案的 metadata window
至少 48h（比仅有 refresh 上界的旧文字更明确，属于待批准约束）。
两数组各按 key_id bytewise 排序唯一，不允许截断以适配大小上限。

MetadataKey 恰为 `{key_id:D,algorithm:"ed25519",public_key:K,
status:active|retiring|revoked,not_before:T,not_after:T,
revoked_at:T|null,revocation_reason:scheduled|security_incident|null,
incident_id:UUID|null}`。key ID 必须等于 public key SHA-256。
所有 key 要求 not_before < not_after。

- active/retiring：后三项必须显式 null，ID 不在 revoked_key_ids；key window
  与 metadata window 相交。
- revoked：revoked_at 非 null 且不晚于 issued_at；scheduled 的 incident_id
  必须 null，security_incident 必须非 null；ID 必须出现在 revoked_key_ids。
- active windows 的并集需覆盖完整 metadata window，半开区间，无空洞。

上述是单份对象内的确定性一致性，不证明跨版本状态合法、累计集合只增、
incident 真实开放或 tombstone 满足保留期限；这些仍属 B02/B03。

## 5. 完整封装字段集与初始根

所有 outer 均为独立 JCS bytes。`payload_b64url` 是其 inner JCS 原始 bytes
的 canonical raw base64url，不能装入另一封装或重新构造的等价 payload。
`payload_sha256:D` 恒等于 SHA-256(decoded payload_b64url)。
这些名称是本附录正式提出的统一选择，不兼容 payload_jcs 等别名。

| outer schema_version | 其他 required 字段（恰好） |
| --- | --- |
| talenro-node-desired-state/v1 | key_id:D, payload_b64url, payload_sha256:D, signature:S |
| talenro-node-state-root-initial/v1 | control_plane_authority_epoch:U, authority_sequence:U, root_version:U, payload_b64url, payload_sha256:D |
| talenro-node-state-root-rotation/v1 | control_plane_authority_epoch:U, authority_sequence:U, previous_root_version:U, root_version:U, payload_b64url, payload_sha256:D, current_signatures:Share[0..64], new_signatures:Share[0..64] |
| talenro-node-state-metadata/v1 | control_plane_authority_epoch:U, authority_sequence:U, root_version:U, payload_b64url, payload_sha256:D, signatures:Share[0..64] |

Share 恰为 `{key_id:D,signature:S}`；各数组排序唯一。格式允许空数组及
结构正确的未知 key share；不补齐/过滤它们，不据数量声称满足阈值。
所有结构损坏或重复 share 拒绝；crypto 层只能忽略结构正确的未知 share，
且永不计入 threshold。64 项是本提案的每组解析上限，不扩大 root key 数。

Outer 与 inner 中每个同名 authority/root 字段必须精确相等。Rotation 的
outer root_version 对应 embedded new_root_set.root_version；metadata 的
root_version 对应其 body.root_version。

初始 root 的 payload 是完整 NodeStateRootSetV1，root_version=1；初始
outer 没有签名字段，不伪造自签阈值，也不是 network rotation。
它是数据库中部署初始化制品的格式，信任仅可来自独立受信部署；未来 writer
必须真实写入对应发布行/指针/fence，reader 不得合成该行或从 HTTP 信任它。
如果 B02 初始化路径尚未具备这些记录，读取继续失败关闭。

根版本大于 1 时，发布行 canonical_payload 是整个 rotation body，不能
只存 new_root_set；published_envelope 是完整 rotation outer。
active root 必须与链尾对应发布投影一致。不增加版本 1 数据库行作为
额外网络信任证据，不修改既有固定 1 起点、最多 32 次轮换限制。

## 6. 摘要、签名 transcript 与投影映射

令 P 为原始 canonical payload bytes，E 为完整 canonical outer bytes：

- payload digest = SHA-256(P)，对应现有 PayloadDigest/payload_digest。
- envelope digest = SHA-256(E)，对应 PublishedEnvelopeDigest；未来 API
  `canonical_bytes=base64url(E)` 时，其 digest 必须是 envelope digest。
- signature transcript = 对象的固定 domain（末尾单个 NUL）拼接 P；
  不签 base64 文本、不签 E，也不把 SHA-256(P) 当作签名消息。

保留三个既有 domain：`TALENRO-NODE-DESIRED-STATE-V1`、
`TALENRO-NODE-STATE-METADATA-V1`、`TALENRO-NODE-STATE-ROOT-ROTATION-V1`。
初始根不在本阶段定义自签 transcript。本文数字编码改变 P，因而会改变
transcript、摘要和签名；即使 domain 保持不变，也不能复用旧签名。

Desired 从当前行的 P、PayloadDigest、SigningKeyID、Signature 构造唯一 E。
先验证 P 格式及与投影 node、epoch/sequence、generation、inventory、resource
version/digest、issued_at、valid_until 的精确相等，再构造并严格回读 E。
E 的摘要是新派生事实，不冒称为数据库原有列，也不覆盖 PayloadDigest。

Metadata 的 P/E/root_version/metadata_version/authority 对应 active metadata
投影；各轮换 P/E/previous/new version/authority 对应各自发布投影；初始 root
只走第 5 节分支。ID、root/metadata 版本等未出现在 desired P 的数据库绑定
继续由 exact SQL 保证，不注入 P。EffectiveDeadline 不转成签名字段，不在
缺少可信 key/metadata 证据时声明已重算其授权效力。

## 7. 长度、嵌套与 HTTP 隔离

| 层次 | 上限（包含全部字节/字符） |
| --- | --- |
| 每个 inner P | 65536 bytes |
| payload_b64url | 87382 ASCII characters，且解码 P 不超过上限 |
| 每个 outer E | 262144 bytes |
| base64url(E) 网络字符串 | 349526 ASCII characters，且解码 E 不超过上限 |
| 单次私有投影 | 既有至多 32 rotations + active root + metadata + desired，不放宽 |

raw base64url 长度采用 ceil(4*n/3)，decoder 同时验证真实解码长度，不只
信任字符上限。本文只限制可接受字节与容器，不保证库内部总分配上界。
64 keys/4096 revoked IDs 是数量上限，不保证所有上限同时能放入 64 KiB；
达到任一限制必须失败，不能截断撤销历史或增大 payload 上限。

现有 API canonical_bytes maxLength=65536 不随本文自动变化。后续生产接线
必须更新 agent/bootstrap 等实际承载本组制品的 schema、生成客户端及
response body/long-poll 聚合限制，并验证至少 32 段最坏响应策略。
本阶段不选择静默删链/分页、不提升 HTTP body 上限、不宣称最大制品可上网。
这里定义的 349526 是单制品编码上限，不是整个响应上限。

纯 stateformat 包内提供固定大小的严格 inner/outer 解析，不直接用
64 KiB 上限的 trust.CanonicalizeJSON 解析 outer，不修改 strictjson.Decode
默认行为。复用已锁定的 JCS 引擎；专用 token 检查须逐层验证重复/required/
类型/null/Unicode/depth，再规范化比较。不得复制 C1.1 信任策略。

## 8. 归属与接口边界

`internal/nodecontrol/stateformat` 独占本文 body/outer wire DTO、严格解析、
格式构造、局部语义和 transcript 字节构造；可以依赖标准库、已锁定 JCS
以及底层纯 contracts 类型，不能依赖 state、serving、authority、store、
trust 的业务层、时间源、signer、网络或数据库。

未来 B02 state workflow 和 serving 私有绑定层各自消费该纯包，不互相
反向导入。ProcessSpec/CapacityLimits 的逻辑 IR 仍按 B02 归属 contracts；
wire DTO 仅在 stateformat 编解码边界存在，不成为第二套业务 IR。
B02 原计划中 state/contracts.go、state/canonical.go 的相应条目需正式
调整为消费纯包；本轮不修改该计划。

结果名称用 Parsed/CanonicalFormat，禁止 Verified/Authorized。成功对象
持有自有 bytes/数组，访问返回副本；可清理的内部临时值在失败时 discard。
不接受 caller-built trusted bool 或外部可替换验签回调。纯格式有限错误
由 serving 映射为 Conflict；尚不改变现有 reader 接线或错误优先级。

## 9. 必须冻结的向量与验收

实施计划必须先列出独立字面向量，再写先失败后通过的测试；期望摘要不能
从被测 encoder 当场生成。设计文档本身不伪造尚未运行的 crypto 向量。

数字正向字面值：`"1"`、`"9007199254740991"`、`"9007199254740992"`、
`"9007199254740993"`、`"9223372036854775807"`，对支持完整 U 范围的字段
逐个保持原整数；负向包括 JSON number、`"01"`、`"+1"`、`"1e0"`、
`"0"`、`"9223372036854775808"`、null 和 missing。

其他必测组：四类完整封装；所有 required/null/extra/type 分支；root/metadata
嵌套逐字段错配；byte-length/depth/count 边界；排序唯一；public-key digest；
Unicode/base64 非规范别名；metadata 时间与 active-window coverage；desired
reason/required=false 的真实存在；C1.1 schema 拒绝；初始根不可作为轮换；
两个合法但不同对象的 payload/key/signature/outer/projection 交换；摘要身份
不能互换；假签名格式通过但不能获得授权；成功副本隔离和失败暂存清理。

现有批准的测试范围未自动扩展。stateformat 的离线纯单元测试和任何
依赖工具测试必须在实施计划明确列出；不运行 DB/Docker/C11/C12/bootstrap/
PITR/race/全仓测试，不改变安全设置，不下载依赖。

## 10. 与既有文件的同步清单

| 文件/边界 | 后续必须同步的内容 | 本轮动作 |
| --- | --- | --- |
| 基础 C1.2 规范 §7/§11 | 本文 schema、字符串整数、reason、完整外层和新增边界 | 不改冻结文件 |
| B02 实施计划 | 唯一 codec 归属、wire DTO/业务 IR 区分、向量和初始化制品 | 不改 |
| B03/agent 消费方 | 使用相同 wire bytes，格式结果不等于验签 | 不改 |
| agent/bootstrap OpenAPI 与生成客户端 | 单制品尺寸、digest 身份；不把外部普通整数一并改字符串 | 不改 |
| strictjson/trust | 保持现有调用的默认 64 KiB 与业务语义 | 不改 |
| DB/schema/generated SQL | bigint 范围和原摘要列身份保持；持久旧制品不能重签冒充原文 | 不改 |
| c12-spec-set.v1.json/验收证据 | 正式协议批准登记需独立受审查变更，旧证据不得沿用 | 不改、不生成批准摘要 |

## 11. G1–G5 对照及审阅结论

G1 由第 4–5 节给出完整字段与初始根；G2 由第 3、9 节给出精确整数编码；
G3 由第 7 节分开存储/解码/网络；G4 由第 4、6 节给出 required/null 与摘要；
G5 由第 8 节确定单一纯包及依赖方向。五项均有具体提案，不再留给实现猜测。

自查：本文的新增限制、旧示例字节不兼容和签名变化均已明示；初始根不
自举信任；格式成功不代表验签；网络与部署门禁未被大小表误关闭。
G1–G5 已获本会话批准作为隔离局部开发契约，不表示生产规范登记或实现完成。
配套计划：`../plans/2026-10-09-task10-desired-canonical-binding.md`，尚待审阅；
仍不自动批准代码执行、生产规范登记、提交或推送。
