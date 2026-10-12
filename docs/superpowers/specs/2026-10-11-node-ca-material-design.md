# Node CA 原始材料、声明公钥与密钥标识兼容性设计

日期：2026-10-11  
状态：书面设计及对应实施计划已获用户批准；五项 Native 任务已完成限定测试及 fresh 只读审查，按部分材料层本机交付（当地 2026-10-12）；不等于信任或整体验收。  
工作区：既有 `codex/c12-b01-task9-coordinator` worktree。  
本轮交付：只新增本文件；不写产品代码、不运行测试、不暂存、不提交、不推送。

审批留痕：用户于 2026-10-11 回答“是”批准设计，随后回答“1”批准对应实施计划并选择 Native。需求正文保持原样；正文中的“待书面批准”“本轮”和末尾草案自检为起草阶段的历史描述，以本段审批状态为准。执行记录见 docs/roadmap/2026-10-11-node-ca-material-execution.md；UTC 执行时间跨当地 2026-10-12，不改获批文件日期。仅既有 worktree 本机交付，不批准 Git 写入或解除原六项 Pending 与整体验收阻塞。

## 1. 目标、已有事实与新增取舍

用户希望继续 C12 架构，在已经完成的 node leaf 材料与声明密钥标识层之后，先设计 CA 材料检查，再推进实际叶证书验签。整体验收阻塞保留，Windows 操作使用 PowerShell，既有未提交进度不得被覆盖。

依据为 [Node PoP 父设计](2026-08-23-node-pop-control-plane-design.md) §6.1–6.2、[Node leaf 原始材料设计](2026-10-11-node-leaf-material-design.md)、[声明密钥标识附录](2026-10-11-node-leaf-key-identifiers-design.md)，以及 [Task10 主状态契约](2026-10-10-task10-main-state-contract-design.md)。父设计要求独立 CA profile，不允许把 leaf 约束直接套给 CA；原 Batch 03 计划仅给出 CA=true、certSign 和独立检查要求，没有替本增量冻结完整 CA 材料规则。

已知事实：

- `nodeleaf` 检查 node leaf 部分形状及调用方 expected SPKI，不认证签名或信任。
- `nodeleafkeys` 证明完整 exact SPKI SHA-256 的 32-byte SKI/AKI 派生关系；issuer SPKI 只是声明值。它不接收 CA DER，不能检查 CA SKI 兼容性。
- 两层的旧公开 API、结果强度、全部六项 Pending、原向量及消费者隔离不在本轮改写。
- 既有 `trust/trustclient` 是 C1.1 配置签名/分发边界，不是本设计的 X.509 CA 信任来源。

本文件推荐的新增取舍均待书面批准：独立包 `internal/nodecontrol/nodeca`、首版仅支持命名 P-256 CA subject key、封闭的四扩展材料子集、exact CA SPKI 与声明公钥逐字对应、CA SKI 使用既有完整 SPKI SHA-256 规则。没有已获批准的外部 issuer 或现有生产 CA 兼容承诺。

本增量的成功条件严格限定为：对同一份 owned CA DER，检查本文件规定的材料形状，确认其 subject SPKI 等于明确声明的 SPKI，并确认其 SKI 使用项目固定规则。它不是完整 CA 认证、证书链验证或可用于运行的信任能力。

## 2. 架构与替代方案

推荐独立 `nodeca`，只依赖标准库与现有 `contracts.Digest`。它不导入 `nodeleaf`、`nodeleafkeys`、主状态、trust provider、serving 或 store；测试也不通过导入旧材料包获取 fixture。因此首版无需放宽任何旧消费者守卫。

本次选择与另两个方案的边界：

| 方案 | 交付与代价 |
| --- | --- |
| 推荐：独立 CA 材料层 | 先识别 exact CA、公钥与 SKI 是否兼容；可单独验证，签名和受信来源仍未完成 |
| CA 材料与 leaf 验签组合 | 能更早证明实际签名关系，但需要新的组合接口及旧材料消费者例外；另行设计 |
| 完整信任包与运行接入 | 同时依赖部署 role、版本/high-water、状态、可信时间、chain 和 runtime；不在本次范围 |

本包不建立系统根池、不发现 issuer、不追踪证书链、不签发证书，也不提供可配置 profile、算法注册器、验证回调或 provider 接口。

## 3. 结果强度与未决责任

新 `Proof uint8` 只有 `None=0` 和 `NodeCAMaterialShapeAndDeclaredKeyBound=1`。成功同时描述三个部分事实：材料符合本文件子集；CA subject SPKI 与声明值逐字对应；CA SKI 等于该 exact SPKI 的完整 SHA-256。

有效 Summary 固定保留以下六项 CA 层责任，按下列顺序排列，不允许调用方填写、删减或标记完成：

1. `CertificateSignatureAndIssuerChain`：CA 自身签名、其 signer 算法政策、parent、链和实际 path constraints。
2. `AuthorityIdentifierAndParentBinding`：CA AKI 与真实 parent 公钥/SKI、issuer 命名及签发关系。
3. `TrustedCurrentTime`：用可信当前时间检查 CA、链和使用窗口。
4. `DeploymentAuthorizationAndPurpose`：部署授权的 exact CA 来源、trust domain、封闭用途与不同用途隔离。
5. `BundleStatusAndRollback`：bundle status、compromised/deauthorized、版本/sequence/high-water、撤销与状态变化。
6. `LeafIssuerRelationAndRuntimeUse`：同一 CA 与 exact leaf 的 AKI/issuer/签名关系及身份、私钥、主状态和运行授权。

`Obligation uint8` 的零值为 `UnknownObligation=0`，上述六项按声明顺序使用 1–6。这份新责任表只描述 CA 层，不能替换或缩短旧 `nodeleaf` 的六项 Pending。`nodeleafkeys` 的现有结果也不因新包存在而自动获得更强结论。

Summary 是可复制、可伪造的值描述，不是认证 capability，不能作为另一次检查的材料输入。不得将成功命名为 Trusted/Verified/Ready，不安装、不保存 active identity，不转为 recovery proposal。

## 4. 输入、声明与预算

建议公开入口为 `Parse(ctx context.Context, in Input) (*Material, error)`。`Input` 恰好只有两个 byte 字段：

| 字段 | 含义及外层预算 |
| --- | --- |
| `DER []byte` | 一张调用方提供的 exact CA certificate DER，1–65,536 bytes |
| `DeclaredSPKI []byte` | 调用方明确声明的该 CA **subject** SPKI，1–8,192 bytes |

这里的 DeclaredSPKI 对应“CA 用于签发 leaf 的公钥”，不是“签发这张 CA 证书的 parent 公钥”。本包不接受 leaf、链数组、parent SPKI、root pool、purpose、trusted flag、预计算标识、issuer ID 或 provenance callback。

两项总输入预算为 73,728 bytes；所有外层上限检查先于克隆和解析。上限是输入预算，不是 parser 总堆内存保证。8,192-byte 声明预算沿用已有 issuer SPKI 输入量级，但首版有效 canonical P-256 SPKI 必须恰好 91 bytes；填充到上限不能成为合法 SPKI。

输入完全来自调用方声明，不从 CA SKI/AKI、DN、系统 roots、自带证书或当前网络连接获得信任。声明值与材料一致，仅证明一致性。

## 5. 首版 CA 材料子集

文档 profile 名为 `node-ca-material.p256-spki-sha256.v1`，仅为固定包内规则的辨识，不是新 wire 字段或协商参数。它是独立 CA profile 的**材料子集**，不等同父设计要求的完整 CA 认证。

### 5.1 Subject key 算法与编码

首版仅接受命名曲线 ECDSA P-256：`id-ecPublicKey` 加 `prime256v1`、未压缩点、BIT STRING unused bits=0、合法且非零的曲线点。CA DER 中的 raw subject SPKI 和声明 SPKI 都须完整解析，满足这一相同规则，并与固定 Go 1.26.5 的 `MarshalPKIXPublicKey` 输出逐字一致；不一致即拒绝，不修复输入。

不接受 RSA、Ed25519、P-384、P-521、P-224、DSA、X25519、显式 EC 参数、压缩点、PEM、裸 key、尾随或未知算法。P-256-only 是本文件提出的首版支持范围，不是由 node leaf 的 P-256 推导出的 CA 必然要求，也不宣称其他算法天然不安全。

现有 `nodeleafkeys` 对 RSA、P-384/P-521、Ed25519 声明 SPKI 的表示检查继续保持原样；它们在此 CA 子集中不支持，不能据此回改旧产品或旧正例。未来扩展 CA 算法必须另行设计审阅，包括公钥强度、签名算法、编码及测试，不静默跟随工具链新增算法。

CA **证书自身**的签名算法属于其 parent，而不必属于 CA subject key。这里不根据 subject P-256 推断签名算法、parent key 或 self-signing；只要求固定 parser 接受字段编码、内外 AlgorithmIdentifier 逐字相等、签名 BIT STRING 非空且 unused bits=0，并完整消费。即使 parser 将签名算法映射为 Unknown，本材料层也不把它改写或称为受支持的签名政策；签名算法许可及实际验签保留为第一项 Pending。

### 5.2 Certificate、DN、serial 与时间形状

- 一张 X.509 v3 certificate、单一完整外层 SEQUENCE，无 PEM、拼接、尾随或未消费的 TBS 字段；禁止 issuer/subject unique identifiers。
- 被本层读取的 DER tag/length 使用确定、最小的长度编码；禁止 indefinite length、高 tag number、越界与整数溢出。嵌套字段和 extension 内层值均必须完整消费，不以 Go parser 成功代替本层检查。
- issuer 与 subject 都为非空 DN：raw Name SEQUENCE 中至少一个非空 RDN SET，每个 AttributeTypeAndValue 完整消费；交由固定 Go parser 检查其字段语法。每个 raw Name 最多 4,096 bytes。不规定 CN、大小写、文本等价或命名授权；不对整个 DN 重新编码后替换原始字节。
- serial 使用最小 DER INTEGER，正数、非零；**编码 content**最多 20 octets，包含必要的正数 sign padding。Summary 只保留去掉必要 padding 后的 unsigned bytes 和有效长度。不把相同 serial 解释为同一 CA，也不证明 serial 全局唯一。
- NotBefore/NotAfter 使用 UTC、整秒、canonical X.509 time 编码；1950–2049 使用 UTCTime，其他 1–9999 年使用 GeneralizedTime；要求 NotBefore < NotAfter。原始 time bytes 与解析结果对应。
- CA 不使用 node leaf 的 24 小时寿命或 issuer_now-5m 规则。本增量不冻结 CA 最大寿命、backdate、当前有效性或根证书豁免；这些是后续完整 CA/time 政策。已过期或尚未生效的材料仍可能通过本层。

RFC 5280 要求 CA subject 为非空 DN；下列封闭子集及数值预算是项目新增取舍，不是声称所有合法 PKI 都必须采用。[RFC 5280 subject](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.1.2.6)。

### 5.3 精确扩展集合

首版恰好四个不同 OID，允许它们以任意顺序出现，但不能重复、缺失或加入其他扩展。critical 的 DEFAULT false 必须省略，true 必须使用 DER `ff`；不接受显式 false。OID 和内层值必须完整消费。

| Extension | critical | 值要求 |
| --- | --- | --- |
| BasicConstraints `2.5.29.19` | true | `cA=true`；pathLen 省略或为最小非负 INTEGER 0–255 |
| KeyUsage `2.5.29.15` | true | 恰好 keyCertSign，或恰好 keyCertSign+cRLSign；其他 bit 全部禁止 |
| SKI `2.5.29.14` | false/省略 | 内层 `04 20` 加 32 bytes，须等于 exact subject SPKI SHA-256 |
| AKI `2.5.29.35` | false/省略 | 内层 `30 22 80 20` 加 32 bytes，仅 primitive `[0]` keyIdentifier；本层只检查形状 |

KeyUsage 内层分别恰为 `03 02 02 04` 或 `03 02 01 06`，unused bits、尾零和命名 bit string 的最短表示都必须正确；不能把 leaf 的 digitalSignature 用途当作 CA certSign。

BasicConstraints 只有 cA 与可选 pathLen；raw 值与 Go 的 `BasicConstraintsValid/IsCA/MaxPathLen/MaxPathLenZero` 必须一致。HasPathLen 明确区分“省略”与“显式零”；0–255 是材料表示上限，不是批准的链深度或忽略更大约束的许可。实际链是否满足 pathLen 仍未检查。[RFC 5280 BasicConstraints](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.2.1.9)。

首版一律要求 AKI，即使自签 CA 按通用标准可省略；这是更窄的项目子集。不能从 DN 相等或 SKI==AKI 判定 self-signed/root。AKI 与 CA 自身 SKI 可以相同或不同，均不证明其派生或 parent 对应。[RFC 5280 AKI](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.2.1.1)。

SAN、EKU、NameConstraints、Policies、AIA、CRLDistributionPoints、其他已知/未知 critical 或 noncritical 扩展全部拒绝。这里是拒绝超出子集的材料，不是忽略约束或认可其缺省授权。cRLSign 只是声明用途，不触发 CRL 下载、解析或撤销认证。

## 6. SKI、AKI 与现有 leaf 层的兼容关系

固定派生沿用已批准的 `node-key-id.spki-sha256.v1` 规则：

```text
CA_SKI = SHA256(exact owned CA RawSubjectPublicKeyInfo DER)
```

使用完整 SPKI 的 tag、length、AlgorithmIdentifier/参数与 BIT STRING，输出全部 32 bytes；无截断、前缀、文本转换、BIT STRING-only、EC-point-only 或重新编码 preimage。重编码仅用于拒绝非 canonical 表示，不能成为摘要输入。

CA SKI 的 20-byte、31/33-byte、SHA-1、截断 SHA-256 或其他表示不兼容本子集；不得在内存中替换 SKI 后接受旧证书。需要改用新 CA/provider profile 时，必须另行批准迁移与信任包部署，不能自动生成新 CA、修改历史 fixture 或扩大 trust roots。

未来要把本层与 leaf 层组合，另行批准的组件必须对**同一组 exact 原始材料**证明：CA raw SPKI 等于 leafkeys 的 DeclaredIssuerSPKI；CA SKI 等于 leaf AKI；实际 leaf 签名与 issuer 命名关系成立；CA 来源、purpose、chain、可信时间、状态和 rollback 约束成立。不能只拼接两份可伪造 Summary 或仅比较摘要来证明这些关系。

本包不输入 leaf，故不直接检验 CA SKI==leaf AKI，也不更新旧层的 CA compatibility UNKNOWN。它最多给出“这张调用方提供的 CA 材料符合固定 SKI 规则”的局部事实；真实外部 CA/provider 的兼容性与部署接受仍未验证。

CA AKI 不对 DeclaredSPKI 派生：DeclaredSPKI 是 CA 自己的 subject key，而 AKI 指向其 parent。没有 parent 材料时不能把二者角色混淆。CA AKI 是否正确、缺失规则如何适用于完整链、parent SKI 对应及自签真实性都由后续认证层负责。

## 7. 检查顺序、错误与生命周期

### 7.1 固定顺序

1. context 非 nil；若已取消/到期，先返回对应 context 错误。检查两项是否有任一超出外层预算，再检查空输入；预算超限优先于空输入。
2. 创建私有 owner，立即安排失败及正常 panic unwind 的清理，再分别独立克隆 DER 与 DeclaredSPKI。不能在首次克隆后才建立清理责任。
3. 完整解析 owned DeclaredSPKI，检查首版 P-256 表示及重编码逐字对应。
4. 对 owned DER 使用固定 Go 1.26.5 ParseCertificate，再进行 CA 专用 raw scanner 检查；raw certificate、TBS、SPKI、DN、serial、time 和扩展与 parser 字段对应。未通过即返回固定错误，不读取调用方 buffer。
5. 按 certificate 基本字段、BC、KU、SKI 形状、AKI 形状的固定顺序检查 §5；不随 extension/map 顺序决定错误优先级。
6. 比较 CA raw SPKI 与 owned DeclaredSPKI，然后比较 CA SKI 与 exact raw SPKI 的 SHA-256；相同阶段出现两项 mismatch 时，SPKI mismatch 优先。AKI 不做派生比较。
7. 计算 exact CA DER、SPKI、raw subject Name 与 raw issuer Name 摘要，构造固定 Summary 与完整 Pending；最后检查 context，才移交成功 owner。

每个解析、重编码、scanner、profile 或绑定阶段在开始和结束检查取消。阶段末已取消时，context 错误优先于该阶段的普通错误；否则返回固定顺序中的首个错误。同步 Go 操作不能被内部即时打断，不承诺硬实时取消延迟。panic 不恢复为成功。

这里新建的是 CA 专用检查，不复制一套 node leaf 身份/profile 逻辑；可沿用旧包的设计模式，但不改其代码或抽取共享 parser 来扩大旧依赖。

### 7.2 五类固定错误

| Sentinel | 用途 |
| --- | --- |
| `ErrInvalidInput` | nil context、空 byte 输入；声明 SPKI malformed、尾随或非 canonical |
| `ErrLimitExceeded` | 任一外层 byte 输入超限 |
| `ErrInvalidMaterial` | CA DER/parser/raw scanner malformed、不完整、非最小编码或 raw/parser 字段不一致 |
| `ErrUnsupportedCAProfile` | 能解释的材料超出首版子集：版本/key 算法、名字预算、serial 范围、时间形状、扩展 inventory/criticality/值、pathLen 或标识长度不符；声明 SPKI 为合法但非 P-256 key 也归此类 |
| `ErrBindingMismatch` | 两份合规 P-256 SPKI 不同，或 32-byte CA SKI 与 exact SPKI 派生值不同 |

另直接返回 `context.Canceled/DeadlineExceeded`。普通错误不携带 parser 动态文本、证书、DN、serial、摘要、公钥、标识、签名或路径；不记录日志。固定 Go parser 在前置阶段已经拒绝的材料仍返回 ErrInvalidMaterial，不人为改成后置 profile 错误：例如重复扩展若由 Go parser 先拒绝，不能要求之后的 inventory 检查再返回 ErrUnsupportedCAProfile。具体 malformed 与 unsupported 的分支及组合优先级必须在实施计划中按上述语义列成测试表，不能靠错误字符串判定。

### 7.3 Material 与 Summary

公开 Material 只有私有共享 owner 和两个方法：`Summary() Summary`、`Clear()`；另有前述包函数 Parse。没有 DER/SPKI 输出、Recheck、Verify、Ready、Authorize、序列化或安装接口。

Summary 只包含值字段：Proof、固定 `[6]Obligation` Pending、`CADERDigest/SPKIDERDigest/SubjectDERDigest/IssuerDERDigest contracts.Digest`、`SubjectKeyIdentifier/AuthorityKeyIdentifier KeyIdentifier`、`Serial [20]byte` 与 `SerialLength uint8`、UTC 整秒的 `NotBefore/NotAfter time.Time`、`HasPathLen bool/PathLen uint8` 和 `CRLSignAllowed bool`。`KeyIdentifier` 是本包独立命名的 `[32]byte` 类型，不是旧包类型或 contracts.Digest 的别名。省略 pathLen 时 HasPathLen=false、PathLen=0，不能误解为实际零约束。

owner 只保留 owned DER、声明 SPKI、固定 Summary 与 valid 状态，不保留 Certificate/public key/big.Int、parser 指针、DN 字符串或临时 marshal buffer。临时 marshal buffer 创建后立即安排清零。Summary 不导出签名算法字段供调用方误当许可政策。

方法 nil-safe，零值返回零 Summary；包装拷贝共享生命周期。任一包装 Clear 后全部包装返回零 Summary，Proof=None；零 Pending 不表示责任已满足。Clear 幂等，不改调用方输入或已经 detached 的 Summary。

失败、取消、正常 panic unwind 与 Clear 清零仍可达的自有 mutable buffers、临时 marshal buffer 和 Summary。保证不涵盖 parser/allocator 不可达分配、immutable strings、运行时副本、调用方值或 fatal OOM/系统崩溃。输入并发修改、并发 Clear/read 均不支持，不引入锁、goroutine、全局 hook 或并发安全承诺。

## 8. 后续实施的验收要求

本节定义未来检查，不表示本轮运行过测试或已有实现。后续实施计划另列精确文件清单、限定命令、执行方法与证据记录。

### 8.1 独立期望值与正例

- 使用已批准附录列出的 RFC 7093 P-256 exact SPKI 和固定完整 SHA-256 值，检查 SPKI preimage 与完整 32-byte 输出；产品函数不能计算自己的期望值。仅复用文档事实，不导入旧产品包或改变旧 fixture。
- 实施阶段在新包自身建立冻结的测试专用 CA golden DER、SPKI、扩展 bytes 与独立计算的摘要。golden 不在测试运行中随机重签；expected 值不由待测 parser/派生函数产生。
- 包内新正例覆盖同 subject 的自签及不同 parent 签发的 CA 材料、KU 两种封闭值、pathLen 省略/0/1/255、扩展不同顺序、不同合法非空 DN。材料通过不称为真实认证。
- 新 fixture 显式设置 32-byte SKI/AKI，不依赖 Go 自动生成标识或 GODEBUG。既有所有 testdata/向量保持原样。

### 8.2 拒绝与不升级结论

- byte 输入空/边界/超限、别名 subslice、克隆前预算、nil/取消/到期 context、阶段失败无半成品。
- PEM、拼接/尾随、BER/非最小长度、截断/越界/溢出、不同 inner/outer AlgorithmIdentifier、unique identifiers、TBS 尾字段及 extension 内层尾随。
- 非 v3、空/超预算 DN、非最小/负/零/21-octet serial、sign padding 边界；CA 时间不使用 leaf 的 24-hour 规则，canonical time、窗口倒置和相等边界。
- 非 P-256 key、非法点、显式曲线/压缩点/参数差异、声明与 raw key 各自 malformed，以及合法不同 P-256 key 的绑定失败。
- 四扩展每项缺失、重复、criticality 或值错误、任意额外已知/未知扩展、显式 false；BC false/非法 pathLen/256、KU 未设置 certSign/额外 bit/非最短 BIT STRING。
- SKI/AKI 各自 1/20/31/32/33/64 bytes、malformed 内层、错误 32-byte SKI；AKI 的 issuer/serial 字段、constructed `[0]`、额外字段或尾随均拒绝。
- 32-byte AKI 任意改变但保持形状时仍可能通过；不能把它误测为 parent 绑定失败。改变 CA 签名字节但保持本层 BIT STRING 形状、parser 映射为 Unknown 的签名算法、过期/未来时间、未部署材料仍可能通过，六项 CA Pending 始终完整；空签名或非零 unused bits 拒绝。
- Subject==Issuer 或 SKI==AKI 不提高结论；没有 leaf 的输入不能声称 leaf AKI、签名或 issuer 命名已验证。
- nil/零值/包装拷贝、重复 Clear、detached Summary、成功 owned 清零、失败及 panic 清理结构；不声称测得不可达内存擦除。
- 对有界 DER/SPKI 及嵌套长度突变运行限定 fuzz；检查不会产生有效半成品。输入预算与正常 unwind 清理仍不等于总堆内存、硬实时或 fatal OOM 保证。

### 8.3 API 与隔离守卫

新包必须有自身的固定 API、产品 import/member、文件 inventory 和消费者守卫。禁止任何其他产品包消费它，包内测试例外须精确到目录及 package；不能放开任意 `_test.go`、兄弟包或运行入口。

产品禁止 x509 Verify/CheckSignature*、CertPool/SystemCertPool、CreateCertificate/私钥/随机数、当前时间、文件/网络/数据库/provider、callback、goroutine、安装/封存/主状态接入。AST 守卫的别名、方法值、局部遮蔽及伪装 selector 拒绝测试须明确；其结果是有限静态检查，不是形式证明。

新包不导入或修改旧 `nodeleaf/nodeleafkeys`。旧隔离守卫无需新增例外；未来组合层不能复用本文件批准来偷开消费者。后续至少运行本包限定测试与旧两层回归，既有整体验收门槛不降低。

## 9. 非目标、写入边界与阶段交接

本设计不实现真实 CA/leaf 验签、root/parent chain、部署签名 trust-bundle、purpose/domain 授权、bundle 状态/high-water、可信当前时间、issuer/serial 唯一性、CSR/签发/provider 迁移、私钥持有证明、身份状态、主状态封存/恢复、TLS listener 或运行授权。Task10/B01/C12/main 的现有阻塞不解除。

本轮只新增这份书面草案；不改 current-status、旧 spec/plan/execution、`.superpowers` 台账、源码、测试、SQL/迁移或配置。不安装依赖，不运行测试，不写 Git 配置/HEAD/索引，不暂存、提交、推送、合并或清理。历史推送授权不扩展成本轮授权，现有本地进度保持原样。

采用 `superpowers:brainstorming` 的架构流程：用户此次选择 1 批准起草；先审阅本文件。书面设计批准后才使用 writing-plans 编写实施计划；计划须再审阅并选定执行方法，之后才允许产品实现。技能中的设计文档提交模板不适用本轮的明确无 Git 写入边界。

书面自检：

- 首版 P-256、四扩展、DN/serial/time/pathLen 与输入上限均明示为材料子集，CA 规则没有沿用 leaf 寿命或用途。
- CA subject key 与 parent signer 分开；CA SKI 派生已定义，CA AKI 仅形状，真实信任及 leaf 关系保留未决。
- 旧两层强度、全部六项 Pending、旧测试向量和消费者规则保持原样；不自动迁移已有 20-byte CA。
- owned 输入、阶段顺序、固定错误、共享生命周期及清理保证的限制明确；没有预先声称测试通过。
- 本文件是完整待审草案，不是已批准 spec、实施计划、完整 CA profile 或远程恢复检查点。
