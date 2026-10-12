# Task 10：desired 规范格式与字段绑定设计

日期：2026-10-09。

配套已批准附录：`2026-10-09-node-state-format-contract-addendum.md`。
用户已批准两份书面设计作为局部开发依据；G1–G5 的精确选择以附录为准。

状态：用户已批准“先规范格式与字段绑定、后签名信任验证”的书面设计。
本文不是已批准实施计划。只保存本地文档，
不提交、不推送、不修改生产代码或既有已批准协议。发现的契约缺口列为
实施前置阻塞，不允许实现者自行补默认格式。

## 1. 目的与成功标准

既有 desiredProjection 已证明查询所得字段的结构一致性和自有内存生命周期，
但摘要正确的任意字节也可能通过它。下一阶段应进一步证明：

- root、metadata、desired 的字节符合各自唯一、闭合的 C1.2 格式；
- 存储的 canonical payload 已经是规范字节，不是读时修补得到的另一份数据；
- payload、发布封装、数据库投影中重复表达的字段精确一致；
- desired 完整封装可由已有 payload、key ID、signature 和摘要确定性构造，
  但必须先有明确的序列化契约，不得把 payload 冒称为完整封装；
- 成功值只证明格式与绑定，不证明签名可信、当前授权或可以发给节点。

保留现有 PowerShell、离线依赖、本地工作区与不发布约束。继续保留所有
既有未提交变更。完整 Task 10、C12 和安全验收不随本设计完成。

## 2. 依据与已确认事实

本次只读核对的来源：

- `docs/superpowers/specs/2026-08-23-node-pop-control-plane-design.md`：
  第 7、7.1、7.4、7.5、11.1 节。
- `docs/superpowers/specs/2026-08-24-nodecontrol-authority-abort-serving-design.md`：
  第 7.1–7.3 节；签名、有效期、可信时间及响应授权属于 B02/B03。
- `docs/superpowers/plans/2026-08-23-node-pop-control-plane-02-operator-state-signing.md`：
  canonical DTO/transcript 归属及类型约束。
- `docs/superpowers/specs/2026-10-08-task10-desired-facts-design.md`、
  `2026-10-08-task10-desired-adapter-design.md` 与当前 serving 实现。
- `internal/strictjson/decode.go`、`internal/trust/jcs.go`：现有解析工具。
- `api/openapi/node-agent-api.v1.yaml`、`node-bootstrap-api.v1.yaml`：
  外部 canonical_bytes 包装及整数范围。

当前 root/metadata 投影各有 Payload 与 PublishedEnvelope；desired 只有
Payload、PayloadDigest、SigningKeyID 和 Signature，没有等价完整封装列。
已有 reader、adapter、converter 和清理函数仍保持原职责。

本次检查时 `internal/nodecontrol/state` 尚不存在。B02 的书面类型定义
不是可直接调用的已实现 canonical codec；不得把它写成现成依赖。
C1.1 的 trust helper 是参考，不等于 C1.2 的 DTO、信任链或验签器。

## 3. 方案选择

采用“两阶段、单一契约所有者”：先由 C1.2 契约层冻结格式，再由本阶段
消费该格式做私有绑定；下一阶段单独设计验签和信任状态。

不采用 serving 内部另造一套长期 DTO 的方案，避免 B02 writer 与 reader
各自序列化而产生两个事实标准。也不一次实现规范格式、验签、可信时间、
生产授权全链路，以免把不同证据混为一个成功标志。

代价是必须先关闭第 4 节的契约缺口。该前置工作是协议设计，不是允许
修改既有协议、schema、OpenAPI 或批准集清单的隐含授权。

## 4. 实施前必须冻结的契约

以下保留发现时的五项缺口和关闭标准，不能留给实现时猜测。
配套附录已逐项提供获批的局部契约；具体包名、字段、编码和上限均以
附录为准。生产批准登记、部署兼容性和网络接线仍未关闭。

### G1：封装的精确 wire schema

现有规范描述 desired envelope 含 schema_version、key_id、JCS payload、
signature、content digest，但所核对资料未给出完整、无歧义的字段名与
payload 表示形式：嵌套对象、JSON 字符串、base64 字节不是同一种签名数据。
Root 初始版本 1 的发布封装也不能被当作 previous_root_version=0 的轮换。

关闭条件：明确 desired、初始 root、root rotation、metadata 的外层与
内层全部字段、schema 字面值、required/null/absent 规则、编码和字节上限；
明确 root 版本大于 1 时数据库 unsigned body 与 new_root_set 的关系。
禁止继承 C1.1 的 schema 或根据样例自行选 payload_jcs 等字段名。

### G2：整数与规范化的精确一致性

当前数据库/投影允许正数直到 MaxInt64，API PositiveInt64 也使用 int64；
基础 desired 示例使用 JSON number。不能默认把这些数全部经过浮点
规范化仍然保留原整数，也不能暗自改为十进制字符串或降低允许范围。

关闭条件：C1.2 所有相关计数/版本/authority 字段统一选择并审批无损 wire
表示，与现有 API、B02 类型和规范化算法相容；固定包含 2^53 边界及
MaxInt64 的逐字向量。若需要协议修订，显式列出受影响文档和兼容性，
不能通过修改 expected bytes 把数值变化掩盖成测试通过。

### G3：解析、存储与网络大小是不同层次

投影允许 payload 至多 65,536 字节、published envelope 至多 262,144 字节。
现有 strictjson.Decode 和 trust.CanonicalizeJSON 都最多接受 64 KiB。
API canonical_bytes 字符串的 maxLength 65,536 也不是相同数量的原始字节。

关闭条件：明确每种 inner/outer 的解码后上限、嵌套深度与编码膨胀，以及
网络包装的独立上限。若批准较大的严格解析入口，必须是明确受限的新
入口或独立 codec，保持现有调用的默认限制；本阶段不得静默放宽全局
strictjson/trust 限制，也不得默默把原 256 KiB 可存储对象缩成 64 KiB。
网络传输和 HTTP 接线仍不属于本阶段。

### G4：字段清单与摘要身份

Metadata 规范同时要求 outer/root_version 与 payload 对应值一致，但其
后续概述并未完整重复字段清单；B02 desired 类型含 reason，基础 JSON
示例未展示它。完整 codec 必须以获批的精确字段清单为准。

关闭条件：逐字段列明 required 与可空分支，区分 SHA-256(payload)、
SHA-256(完整 envelope) 与签名 transcript 的身份；明确 API digest 指向
哪一种。数据库 PayloadDigest 继续按原始 payload 计算，不改名偷换。

### G5：单一所有者和无环依赖

B02 计划把 state canonical DTO 放在 state，并让 state workflow 消费
serving；因此直接让 serving 反向依赖未来整个 state 包存在环依赖风险。

关闭条件：通过获批附录指定一个无 DB/authority/serving 依赖的纯格式
所有者，B02 writer 与 serving 共用它。优先评估放入既有 contracts 的
专用格式文件，或提取独立纯 codec 包；不得同时保留两套类型/helper。
这是待审批的包归属调整，不在本轮实施或更改 B02 计划。

## 5. 契约冻结后的组件与数据流

1. 纯格式层检查长度、严格 JSON 结构、闭合字段、原始值类型、规范表示
   和局部语义。拒绝重复/未知字段、尾随对象、非法 UTF-8、非规范编码、
   非法 null/缺字段，不能只依赖 Go 零值区分 required 字段。
2. 生成规范字节并与输入逐字比较；不将非规范输入重写后当作原文接受。
   比较数字时不容忍有损转换。返回独立持有的数据和固定有限错误。
3. 私有绑定层消费 desiredProjection，逐项对照已解析字段、投影字段和
   封装中的 payload；检查全部 root rotations，不能仅检查链尾。
4. 在精确 wire schema 已冻结后，desired envelope 从所选行的原始 payload、
   原始签名、key ID、payload 摘要按唯一规则构造；不重新签名、不选新 key，
   不把数据库没有存储的字段当作从数据库读取。构造结果须严格回读验证。
5. 仅在所有检查完成后交付私有格式结果，命名需表达 format/binding，
   不能命名 Verified、Authorized 或冒用最终 DesiredStateFacts。

本阶段先做纯组件和隔离组合测试，不改现有 readDesiredProjection 的
成功语义或将格式结果接入生产。未来把格式检查放入 guardedRead callback
时必须保持 domain error 后检、取消优先和失败 discard；不能先交付再校验。
不允许可配置 verifier、空操作 validator 或调用方传入“已验证”布尔值。

## 6. 精确绑定要求

| 对象 | 必须绑定 | 不能据此推断 |
| --- | --- | --- |
| desired payload | node、generation、inventory version、resource envelope version/digest、authority epoch/sequence、issued_at、valid_until 对应所选行 | 当前库存逐 slot 相等、已验签、当前有效 |
| desired 完整封装 | 原始 payload、payload digest、SigningKeyID、Signature | 密钥可信或签名有效 |
| metadata | outer/body 的 authority 与 root version；metadata version 对应 active metadata；payload 原文和摘要对应发布投影 | 阈值满足、撤销历史连续、当前授权 |
| 初始 root | 获批初始发布格式与版本 1；投影中对应 authority/payload/digest | deployment trust anchor 已认证 |
| 每一 root rotation | body/embedded new_root_set 的 authority 相等；previous/new version 与投影及相邻链相符；外层绑定原始 body | current/new threshold 已验证、可跳过过期链 |

仅绑定实际存在于 canonical body 的字段；数据库 publish/signing/operation ID
未被定义为 payload 字段时，不向签名 body 注入它们。现有 exact SQL 对这些
ID 的检查仍独立保留。EffectiveDeadline 是存储的派生字段，不冒充原始
签名字段，不在缺少可信 metadata/key 证据时宣称已重新验证它。

局部格式检查可检查排序、唯一性、固定算法、字节长度、key ID 与 public key
摘要对应等由格式契约规定的关系；不得据签名长度或数组数量声称阈值通过。
未知签名的结构容许规则必须与最终 threshold verifier 契约一致，不能由
本层私自替代它的信任判定。

## 7. 所有权、错误和隐私

输入投影只读，不修改或清零调用方的数据。成功值拥有自身 bytes/数组，
快照返回副本；失败、取消或异常时丢弃本组件已分配的可访问临时材料。
不宣称擦除不可变字符串、JSON 库内部或不可达运行时副本。

格式不符、同体异值或层间不匹配，在 serving 绑定边界归为固定 Conflict；
取消为 Canceled；来源/后检失败为 AuthorityUnavailable，优先级沿用现有
guard。纯格式组件不导入 authority，错误映射由绑定层执行。不返还部分
对象，不把字节、key 材料、SQL、解析器动态报错写入公开错误或日志。

## 8. 验证设计

先冻结独立的字面向量，再先 RED 后实现。期望值不能由被测 encoder 生成。

- 正向：初始 root、连续轮换、metadata、desired 完整封装，各自 body 与
  envelope 原始 bytes、payload/envelope digest 和字段映射。
- 字段攻击：逐字段删除/重复/改类型/置 null/增加未知字段；嵌套层同样覆盖。
- 编码：键顺序/空白/转义、非规范 base64/hex、非法 Unicode、数字精度边界、
  大小和深度上限；语义相同但非规范原文仍拒绝。
- 绑定：逐项改 outer/body/projection 中一层，或交换不同合法对象，均失败；
  特别覆盖 metadata root_version、rotation embedded authority、desired key
  和 signature 来自另一行、内外摘要身份混淆。
- 格式不是验签：构造格式正确但密码学签名无效的向量，本层可通过；测试
  名称明确它未被授权，未来 crypto 层必须拒绝，不能把假签名当信任证据。
- 所有权：单次变异立即断言；失败清理与 success 对照，不采用双重变异
  抵消的测试方式。纯组件不连接数据库、时间源、signer 或网络。
- 跨 schema：C1.1 合法制品不能被接受为 C1.2 对象。

实施计划需列出批准的 codec 所属包及其测试命令。当前只批准的测试范围
仍是 serving/contracts/readiness、限定 authority selector 和 store 编译；
新增包或修改共享 strictjson/trust 时须在计划中明确验证范围，不能默认为
已有许可。真实 DB/Docker、全仓、C11/C12/bootstrap/PITR/race 仍不运行。

## 9. 本阶段不交付的能力

不验 Ed25519 签名、不判断 root threshold/可信 anchor、撤销历史、可信当前
时间、响应提交授权，不创建 production Reader 或 Coordinator 工厂，不改
SQL/schema/generated store、证书路径、security 设置、依赖锁和批准集。
不生成可当作生产批准凭据的 spec digest 或验收结论。

## 10. 审阅与下一步

范围、数据流与拒绝规则及 G1–G5 的配套局部契约已获书面批准。
实施计划已编写为 `../plans/2026-10-09-task10-desired-canonical-binding.md`，
须再审阅后才执行。不能用既有 adapter 测试代替格式验证，也不能把
局部设计批准解释为部署兼容性、生产协议登记或实际验收。

自查结论：缺口均有附录中的明确选择，早期方案描述以附录细化结果为准。
实施须等计划确认；生产接线、批准集更新、提交和推送未获此次授权。
