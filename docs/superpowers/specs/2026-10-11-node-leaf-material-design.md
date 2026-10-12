# Node leaf 原始材料校验设计

日期：2026-10-11  
状态：书面设计已获用户批准；Native 五任务实现、限定验证与一次独立只读审查完成，仅本机交付。  
工作区：既有 `codex/c12-b01-task9-coordinator` worktree。  
阶段边界：本次只新增设计文档，不实现、不测试、不暂存、不提交、不推送。

实施收尾：以上及正文的“本轮”保留设计编写阶段的历史说明；后续实施按获准计划完成。
仅有限更新本状态，需求正文不变。17 项最终限定验证通过，六项 Pending 不消失，
两项历史留痕限制已披露，无 Git 写入或整体验收升级。
见 [执行记录](../../roadmap/2026-10-11-node-leaf-material-execution.md)。

## 1. 目的与证据强度

在已完成的时间证据绑定增量之后，增加一个独立、离线、纯材料校验包 `internal/nodecontrol/nodeleaf`。它接收 exact node leaf DER 和调用方预置的节点、信任域、公钥期望，检查本设计列明的结构与逐字对应关系。

成功只表示 `NodeLeafMaterialShapeValid`：单张材料满足本增量的封闭形状规则，并与显式期望对应。它不是完整的 node leaf profile 认证，更不是可信证书、当前可用身份或运行授权。解析证书与验证证书是不同操作。[Go 解析接口](https://pkg.go.dev/crypto/x509#ParseCertificate)、[Go 验证接口](https://pkg.go.dev/crypto/x509#Certificate.Verify)。

每个有效结果始终携带以下六项未完成责任，固定顺序、不可为空：

1. `LeafSignatureAndIssuerTrust`：签名、exact issuer、CA profile、信任链和部署授权的信任包。
2. `KeyIdentifierDerivationAndIssuerBinding`：SKI/AKI 的项目派生算法及 subject/issuer 公钥来源对应。
3. `TrustedCurrentTime`：可信当前时间、当前有效性、issuer backdate 与签发模板时间对应。
4. `IdentityAndAuthorityProvenance`：issuer/serial 唯一性、身份 epoch/lineage、行状态、历史、authority/fence 来源。
5. `PrivateKeyPossession`：本地 keystore/CSR 对应和私钥持有证明。
6. `MainStateAndRuntimeBinding`：主状态完整绑定、封存、持久化、恢复选择和实际运行授权。

即使所有本包测试通过，这六项也不会消失。过期、尚未生效、签名错误或密钥标识来源不正确的材料，仍可能满足本包的部分条件；不得据此安装或使用证书。

## 2. 已核对的项目边界与方案选择

依据父设计 `docs/superpowers/specs/2026-08-23-node-pop-control-plane-design.md` §6.1–6.3，以及主状态契约设计 `docs/superpowers/specs/2026-10-10-task10-main-state-contract-design.md`。

现有 `serving/certificate_facts.go` 只检查读取事实和 opaque DER/digest；未解析或认证 X.509。`serving/types.go` 的查询序列号是 unsigned magnitude，公钥摘要没有在该接口中冻结 SPKI preimage。主状态中的 issuer、serial、公钥等引用也不是完整证书事实。现有测试中充当 opaque DER 的字节不能充当本包正例。

三个方案中选择独立 `nodeleaf`：

- 扩展 `stateformat` 会混入已冻结的 JSON/格式证据边界。
- 直接放入 `serving` 会耦合数据库读取与材料检查。
- 独立包可以单独检查原始材料，保留上述来源和认证责任，且不改变既有调用路径。

本增量不导入 `serving`、`statecontract`、`staterecovery`、`statetimebinding` 或其他隔离证据包；允许沿用 `contracts.Digest` 和现有 UUID 类型。其他产品包不得导入或调用 `nodeleaf`。不修改旧源码、旧隔离守卫、现有测试向量、SQL、迁移、主状态模型或当前运行入口。

## 3. 输入与预算

`Input` 只包含：

| 字段 | 约束与含义 |
| --- | --- |
| `DER []byte` | 单张 exact DER，1–65,536 bytes；不接受 PEM、链数组、拼接证书或尾随字节 |
| `NodeID uuid.UUID` | 非零、RFC4122 variant；用于生成固定 lowercase、带连字符的 canonical UUID |
| `TrustDomain string` | 预置 exact 小写 ASCII DNS 名；总长 1–253，每 label 1–63，首尾为 `[a-z0-9]`，内部仅 `[a-z0-9-]` |
| `ExpectedSPKI []byte` | 预置 exact DER SubjectPublicKeyInfo，1–512 bytes；必须是命名曲线 P-256 的合法 ECDSA 公钥 |

不接受 caller-supplied `verified` 标志、私钥、CA、issuer ID、当前时间、证书状态、epoch/fence 或可变回调。输入值的可信来源不是本包所证明的内容。

所有外层长度预算在复制或解析前检查；不 trim、case-fold、IDNA 转换或 URL 规范化。期望 URI 固定由合法输入拼成：

`spiffe://<TrustDomain>/node/<NodeID.String()>`

最大长度为 304 ASCII bytes。expected SPKI 和 leaf SPKI 必须逐字相同；只比较点坐标相同不足以替代 exact DER 对应。

## 4. 封闭材料检查

### 4.1 单张原始 DER 与基本字段

先拥有输入，再调用 Go 1.26.5 的 `x509.ParseCertificate`。不得先解析调用方可变 buffer 再补复制。要求原始证书完全消费，`cert.Raw` 与 owned DER 逐字相同；不通过重新编码或“修正”材料获得接受结果。

检查 X.509 v3、empty subject DN 的 exact DER `30 00`，以及 P-256 ECDSA 的合法非零曲线点。expected SPKI 使用 `x509.ParsePKIXPublicKey` 检查，并与 leaf 的 `RawSubjectPublicKeyInfo` 逐字对应。叶证书签名算法只是解析所得字段：本增量不冻结 issuer 算法白名单、不执行签名运算、不用 leaf 公钥类型推断 issuer 公钥类型。

使用私有、有限的原始字段检查补充 Go 解析：最短 definite DER length、所检查值的正确 tag/primitive/constructed 形式、完整消费，以及 serial、subject、SPKI、extensions 的原始范围对应。拒绝本增量不支持的 issuer/subject unique identifiers 或额外 TBS 尾字段。此检查不是一般 ASN.1 框架，也不声称证明全部 RFC 5280 语义。

序列号必须数学值大于零；显式检查，不依赖 `GODEBUG` 对负数解析的兼容设置。DER INTEGER 内容必须是最小的正数表示，含必要 sign-padding 时总长仍不超过 20 bytes。导出的 serial 是去掉必要 sign-padding 的最小 unsigned magnitude，长度 1–20；不导出负数、零或多余前导零。20-byte magnitude 的最高位为 1 时会需要 21-byte INTEGER 内容，本增量明确拒绝；较短 magnitude 的必要 `00` padding 允许。此保守材料规则不改变现有 lookup 的输入规则，也不表示每个合法 lookup serial 都有本包接受的证书。[RFC 5280 序列号与编码说明](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.1.2.2)。

`NotBefore`、`NotAfter` 的原始时间值必须含秒、以 `Z` 结尾，无时区 offset 或小数秒；1950–2049 年采用 UTCTime，其余可表示年份采用 GeneralizedTime。解析字段为 UTC，年份 1–9999，前者严格早于后者，且 `NotAfter == NotBefore.Add(24*time.Hour)`。本包不读取系统时间，不证明 `NotBefore=issuer_now-5m`、签发模板时间或 `NotBefore <= trusted_now < NotAfter`。Summary 中两者是证书声明的时间，不是运行 deadline。

### 4.2 扩展集合与原始值

扩展恰好六项。先按原始顺序检查 OID inventory，拒绝缺失、重复、额外、未知 critical/noncritical 或错误 criticality；随后按下表固定顺序检查每项值，不依赖 map 遍历顺序。不要求证书中的六项采用特定排列。

| 顺序 / OID | critical | 本增量接受的 exact 内容 |
| --- | --- | --- |
| BasicConstraints / `2.5.29.19` | true | `30 00`，即默认 `CA=false` 且无 pathLen；解析字段也必须 `BasicConstraintsValid=true, IsCA=false` |
| KeyUsage / `2.5.29.15` | true | 最小 BIT STRING `03 02 07 80`，只含 digitalSignature |
| EKU / `2.5.29.37` | false | 完整 DER SEQUENCE，恰好一个 clientAuth OID `1.3.6.1.5.5.7.3.2` |
| SAN / `2.5.29.17` | true | 完整 DER SEQUENCE，恰好一个 context-specific primitive `[6]` URI；内容为 1–304 ASCII bytes，并逐字等于输入生成的期望 URI |
| SKI / `2.5.29.14` | false | 完整 DER OCTET STRING，非空、1–64 bytes |
| AKI / `2.5.29.35` | false | 完整 DER SEQUENCE，恰好一个 primitive `[0]` keyIdentifier，非空、1–64 bytes |

所有上述内层值必须最短 DER 编码并完整消费，不能有 trailing TLV、重复子项或非法 tag。false criticality 的 DEFAULT 不得显式编码为 false；已有 Go parser 未拒绝的这些形式仍由本包原始检查拒绝。Any EKU、serverAuth、未知 EKU、额外 KU、显式 default `CA=false`、pathLen、其他 SAN 类型或第二个 GeneralName 均拒绝。

仅检查 `cert.URIs` 长度不足够：本地 Go 1.26.5 parser 对某些 GeneralName arm 不形成公开 SAN 列表。因此必须检查原始 SAN，拒绝 URI 加 otherName/registeredID 等组合。AKI/SKI 也必须检查原始完整消费，不能只检查 Go 解析结果非空。实现依据是本地已核对的 Go 1.26.5 `crypto/x509/parser.go`，不是未固定版本的在线源码行为。

SAN 不接受大小写差异、percent encoding、额外 path、端口、userinfo、query、fragment 或末尾 `/`；本包没有 URL“等价”判定。ASCII 单 URI 不等于期望时返回绑定错误；类型、数量或 DER 不合规时返回相应材料/子集错误。

SKI/AKI 的 64-byte 上限及 AKI keyIdentifier-only 是本增量的本地保守子集，不是新生产 wire profile，也不宣称其他 RFC 合法 AKI 形式非法。RFC AKI 可以有 issuer/serial 信息，key identifier 也不存在本设计可以擅自采用的唯一派生方法。[RFC 5280 AKI/SKI](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.2.1.1)。

父设计要求 SKI/AKI 按项目 profile 算法从 exact subject/issuer 公钥导出，但现有项目尚未冻结该算法。因此本包只检查封闭形状，不能选择 SHA-1、截断 SHA-256 或 Go 默认生成行为来补齐该要求。后续独立设计必须冻结算法、issuer 来源和对应规则，再证明完整 profile；不能直接把本包结果改名为完整认证。

## 5. 输出、所有权与清理

对外接口限定为 `Parse(ctx, Input) (*Material, error)`、`Material.Summary() Summary`、`Material.DER() []byte` 和无返回值的 `Material.Clear()`。这些方法对 nil receiver 安全。本增量不增加 Recheck、链验证、序列化、安装、状态转换或运行入口。

`Material` 是无导出字段的薄包装，只持有私有 owner。成功 owner 保留 owned exact DER、owned expected SPKI 及固定 summary；临时 `*x509.Certificate`、`*ecdsa.PublicKey`、`big.Int` 或解析切片不进入结果。复制 Material 包装仍指向同一 owner，任一 Clear 后其他包装不能读到陈旧有效 summary。

Summary 以值返回，包含：

- `Proof=NodeLeafMaterialShapeValid`、固定六项 `Pending`。
- `NodeID`、exact `TrustDomain`。
- `LeafDERDigest=SHA256(exact DER)`、`SPKIDERDigest=SHA256(exact leaf RawSubjectPublicKeyInfo)`。
- `[20]byte` serial 与显式有效长度；未使用区域为零。
- `NotBefore`、`NotAfter` 的 UTC 声明值。

Summary 是可复制、可伪造的描述值，不是授权 capability。零值、nil owner 或已清空对象返回 `Proof=None` 的零 summary，不能把零 Pending 当作“责任已满足”。`DER()` 返回 detached copy；不导出可变解析对象、公钥指针或内部 buffer。

`SPKIDERDigest` 的名字冻结摘要 preimage；不得自动写入现有 `PublicKeyDigest`、主状态公钥引用或 serial/issuer digest，也不提供这些字段的隐式转换。Leaf DER 摘要不能代替 SPKI 摘要；本包没有足够事实生成 issuer ID。

取得 owner 后立即建立清理责任；失败、取消和正常 panic unwind 都清零仍可达的自有 mutable buffers 与 summary。只有最终成功后移交 owner。`Clear()` nil-safe、幂等；其后 `DER()` 返回 nil。清理不能影响输入和已返回的 detached copy。

不声称擦除 immutable strings、Go parser/allocator 的不可达分配、运行时复制、系统崩溃/fatal OOM 后内存，或调用方的 detached 数据。API 不支持并发 Parse 输入写入、并发 Clear/read 或跨 owner 的同步；不得用 race 测试替代声明的非并发边界。

## 6. 错误与取消

只返回固定、可用 `errors.Is` 判定的类别：

| 类别 | 例子 |
| --- | --- |
| `ErrInvalidInput` | nil context、无效 UUID/domain、expected SPKI 不是合法 P-256 |
| `ErrLimitExceeded` | DER 或 expected SPKI 超出外层预算 |
| `ErrInvalidMaterial` | 证书/所检查原始 DER 的语法、tag、最短编码或完整消费失败 |
| `ErrUnsupportedLeafProfile` | 可解析但超出本增量字段、扩展集合、用途、序列号或窗口子集 |
| `ErrBindingMismatch` | 单 URI 或合法 P-256 leaf SPKI 与固定期望不逐字相同 |
| `context.Canceled/DeadlineExceeded` | context 已取消或已到期 |

空 DER/SPKI 属于无效输入，不作长度超限；过长 domain 属于输入格式错误。错误不包装 Go parser 的原始错误、不携带 DER、URI、key ID、节点、serial、摘要或签名值；不记录日志。

检查顺序冻结为：context → 外层预算与空输入 → UUID/domain 纯值形状 → 创建 owner 并建立清理责任 → 分别拥有 DER/SPKI → owned expected SPKI 的完整解析与 P-256 形状 → owned leaf 的 Go 解析 → 原始基本字段/扩展 inventory → 六项扩展固定顺序 → 其余 leaf 字段子集 → URI 绑定 → SPKI 绑定 → 最终 context → owner 移交。expected SPKI 的原始范围也须为单一、最短 DER 且完全消费，不能解析调用方 buffer。每阶段第一项失败决定固定类别；取消检查使用阶段边界，不额外暴露失败字段。

Go parser 在 profile 检查前已拒绝的材料统一归为 `ErrInvalidMaterial`；例如某些重复扩展或 critical SKI/AKI，不保证进入后续 `ErrUnsupportedLeafProfile` 分支。对负数 serial 的最终接受集必须独立于 parser 兼容设置：无论 Go 先拒绝还是显式正数检查拒绝，都不能形成结果。

同步的标准解析调用不能被本包在内部中断；预算和阶段检查限制工作量，但不承诺硬实时取消延迟。解析之后若 context 已取消，清理并返回取消错误，不返回半成品。panic 保持向上传播；不恢复成成功，也不增加可注入的产品 callback/global hook。

## 7. 后续实施的测试与验收要求

本节定义待验证要求，不表示本轮已经运行或通过测试。测试与文件清单、逐条命令预算、冻结范围和执行方法由书面实施计划另行列明并获批准。

正例必须是离线真实 DER 证书：固定测试 issuer/leaf P-256 key、固定 node/domain/日期/serial、显式六项扩展及 criticality。可在 `_test.go` 使用真实签发构造，但不得依赖 Go 默认 SKI/AKI 或生产 parser 生成期望真值。fixture 的 exact 字段和摘要应有独立断言；不得改写旧捕获向量或用 opaque DER 假正例。

至少覆盖：

- 单 DER、PEM/拼接/尾随、outer/inner 非最小编码和非法 tag；所有预算边界先于复制与解析。
- v1/v2、非空 subject、错误 key type/curve、非法点；合法不同公钥与 SPKI preimage 对应。
- serial 零/负/额外前导零、必要 sign-padding、20-byte INTEGER 内容边界及需 21-byte 的拒绝。
- 字段窗口恰好 24h、前后倒置、非 24h、时间 tag/年份/offset/小数秒；expired/future 的合规窗口仍只产生部分条件结果。
- 六项逐一缺失、重复、额外、criticality 错误；未知 noncritical 也拒绝；KU/EKU/BC 的逐项偏差。
- 单 URI 的逐字差异、第二 URI、DNS/IP/email、以及 Go 公开 SAN 列表可能忽略的 otherName/registeredID。
- SKI/AKI 空/超限/尾随/重复/额外 arm；形状合规但派生关系错误的材料仍保留密钥标识责任。
- ASN.1 合法但密码学签名错误的材料不被误称为可信；有效 Pending 精确包含六项且没有 Ready/Authorize 结果。
- 输入别名与共享 subslice、失败后 owner 清理、包装复制后 Clear、detached summary/DER、预取消及阶段取消。
- panic 清理结构与内部 owner 的可达 buffers；不使用导出 test hook，也不宣称测到不可达内存擦除。

新增本包隔离守卫应检查产品 imports、公开 API、禁止调用和外部消费者。产品代码仅允许已批准的标准库解析/字段检查/摘要及既有类型依赖；禁止 `Verify`、`CheckSignature*`、CertPool/system roots、TLS、证书/CSR 签发、私钥读取、网络、DB、文件 I/O、系统时钟和随机源。测试用真实签发不计作产品调用。不得放宽任何旧隔离守卫。

后续限定回归面为新包及上一增量已批准的离线回归面，使用本地 Go 1.26.5 与逐条有界执行。全仓测试、race、真实 DB、Docker、硬件密钥、C11/C12 PITR 或生产 conformance 不在这个增量中；现有验收阻塞不会因本包结果消失。

## 8. 非目标与阶段交接

本设计不完成 CA profile/签名/信任包、deployment authority、证书状态与身份迁移、时间接入、Recovery/LKG、主状态 deterministic encoding/seal/counter/真实 I/O、私钥证明或运行授权。独立包成功后仍不能被转换成 staterecovery proposal 或进入 active slot。

本轮只产生本文件；不修改 current-status、旧执行记录或 `.superpowers` 台账，不创建包目录、代码、实施计划或测试产物。既有未提交内容属于此前进度，保持原样；不更改 Git 配置、HEAD、索引，不提交或推送。

按 `superpowers:brainstorming` 的架构路径，用户审阅并批准本文件后，下一阶段才是编写书面实施计划；实施计划还需审阅并选定执行方法，之后才可实现。当前对话确认不批准尚不存在的计划或代码。

## 9. 书面自检

- 范围：独立的部分材料检查，无既有产品接入、旧文件修改或额外运行路径。
- 一致性：严格匹配与有限 profile 子集没有升格为证书/身份认证；SKI/AKI 派生、issuer 与可信时间缺口显式保留。
- 明确性：预算、序列号 sign-padding、六项 OID/criticality、raw SAN/AKI 完整消费、摘要 preimage、错误顺序和 owner 生命周期均已定义。
- 阶段：原设计编写时待批准、未实施；现已按批准计划完成限定本机实现与审查，未进行 Git 写入。本节不代替执行证据。

