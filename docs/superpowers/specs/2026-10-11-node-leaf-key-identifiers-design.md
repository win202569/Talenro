# Node leaf SKI/AKI 派生与声明签发者公钥绑定附录

日期：2026-10-11  
状态：书面附录及后续计划已获用户批准；Native 五任务实现、18 项限定验证和一次独立只读审查完成，仅本机交付。  
工作区：既有 `codex/c12-b01-task9-coordinator` worktree。  
本轮交付：只新增本文件；不改代码、不运行测试、不暂存、不提交、不推送。

审批留痕：用户于 2026-10-11 批准本书面附录，进入实施计划编写阶段。
正文中的“草案”“本轮”“尚未批准”保留起草阶段历史说明，需求正文不变。后续用户选择 1 批准计划并按 Native 完成；18 项最终命令通过，fresh review 0 Critical / 0 Important / 0 Minor，全部六项 Pending 保留、CA 兼容 UNKNOWN、整体验收不升级。见 [真实执行记录](../../roadmap/2026-10-11-node-leaf-key-identifiers-execution.md)。无暂存、提交、推送或远程恢复检查点。

## 1. 目的、依据与证据边界

本附录针对父设计 §6.1 尚未冻结的 SKI/AKI 派生算法，定义一个可单独实施的离线增量：证明 node leaf 中的两个密钥标识，与 exact subject SPKI 和调用方明确声明的 issuer SPKI 对应。

依据为 [Node PoP 父设计](2026-08-23-node-pop-control-plane-design.md) §6.1，以及已实现的 [Node leaf 原始材料设计](2026-10-11-node-leaf-material-design.md)。现有 `nodeleaf` 仅接受 SKI/AKI 的 1–64-byte 封闭形状，不证明它们从公钥派生；其 API、正反例与既有结果强度不在本轮改写。

本附录的成功描述为 `NodeLeafDeclaredKeyIdentifiersBound`，含义严格限定为：

- 同一份 owned leaf 满足既有 `NodeLeafMaterialShapeValid`。
- SKI 等于本附录算法对该 leaf exact SPKI 的结果；该 SPKI 已与预置 expected SPKI 逐字对应。
- AKI 等于同一算法对 owned `DeclaredIssuerSPKI` 的结果。

这里的“声明签发者”不等于真实、可信或部署授权的签发者。调用方完全可以提供攻击者公钥，并提供 AKI 与它对应、但签名错误的证书；该关系检查仍可能成功。不得把结果命名为 Verified/Trusted/Ready，不得安装证书或接入实际运行授权。

原六项 Pending 仍须按原顺序完整保留，尤其不删除 `KeyIdentifierDerivationAndIssuerBinding`：本增量只完成其派生与声明对应部分，没有证明 issuer 公钥的可信来源及其确实签发该 leaf。

## 2. 推荐方案与替代方案

### 2.1 算法比较

推荐完整 SPKI DER 的 SHA-256、无截断。RFC 7093 §2 方法 4 描述了哈希 DER SubjectPublicKeyInfo 的方式，§3 给出了 SHA-256、不截断的例子；该 RFC 是信息性文件，不意味着所有 PKI 都必须采用此方案。[RFC 7093 方法及例子](https://www.rfc-editor.org/rfc/rfc7093.html#section-2)。

| 方案 | 精确 preimage / 输出 | 项目取舍 |
| --- | --- | --- |
| 推荐：SPKI-SHA256-32 | 完整 DER SubjectPublicKeyInfo / 全部 32 bytes | 与项目的 exact SPKI 边界一致，算法标识与参数也参与摘要；可使用独立公开向量 |
| SHA256-BITSTRING-20 | subjectPublicKey BIT STRING 的有效内容，不含 tag、length、unused-bits count / 最左 20 bytes | 可与采用 RFC 7093 方法 1 的系统对应，但不是完整 SPKI 摘要 |
| SHA1-BITSTRING-20 | 相同 BIT STRING 内容 / 20 bytes | 可对应已有 RFC 5280 常用标识；本项目不选择它，也不因此宣称采用它的证书自动不可信 |

后二者作为互操作比较，不是可配置 fallback。密钥标识不是签名验证或授权机制；推荐方案的理由是精确编码契约、固定算法和可独立验证的输入，不是“换成 SHA-256 就完成信任认证”。[RFC 5280 密钥标识说明](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.2.1.2)。

### 2.2 架构比较

推荐新建独立 `internal/nodecontrol/nodeleafkeys`，组合已有 `nodeleaf.Parse`，只增加派生关系检查。

- 扩展旧 `nodeleaf.Input/Parse` 会把形状检查变成更强的关系检查，破坏已批准的接口和旧测试语义，不采用。
- 独立关系包保留旧结果强度与六项 Pending，能单独验证；代价是需要一个明确、狭窄的旧消费者守卫例外，见 §8。
- 一次性加入 CA profile、签名、链、trust-bundle、可信时间与运行接入，会同时依赖尚未冻结的来源接口，不纳入本增量。

不复制一套完整 leaf profile 检查，也不增加算法注册器、运行时配置或可注入的验证回调。

## 3. 推荐冻结的派生契约

草案中的 profile 名称为 `node-key-id.spki-sha256.v1`，仅用于文档和包常量的辨识，不是新 wire 字段，也不是调用方可选的协商参数。

定义 `KeyIdentifier` 为独立命名的 `[32]byte` 值。派生公式：

```text
SKI = SHA256(exact leaf RawSubjectPublicKeyInfo DER)
AKI = SHA256(exact DeclaredIssuerSPKI DER)
```

哈希输入包含 SPKI 的 SEQUENCE tag、DER length、AlgorithmIdentifier、参数以及完整 subjectPublicKey BIT STRING 编码。输入不含 PEM、证书外层、私钥、issuer/subject DN、serial、证书签名或任何前缀。不得先重新编码、截取 BIT STRING、只哈希 EC 点、拼接 domain separator 或转换为文本。

输出是全部 32 个二进制 octets，不是 64 字符 hex、base64 或前 20 bytes。leaf 中的 extension 内层值必须为：

| 字段 | extnValue 所包裹的 exact DER |
| --- | --- |
| SKI | `04 20` 后接 32-byte SKI |
| AKI | `30 22 80 20` 后接 32-byte AKI；只有 primitive `[0]` keyIdentifier |

OID、noncritical、DEFAULT false 的省略、六项扩展 inventory 与完整消费规则继续由旧 `nodeleaf` 保证。两个标识不新增不同用途的派生前缀；对同一个 exact SPKI，它们可以相同，不以 `SKI != AKI` 作为安全检查。

按该公式，SKI bytes 数值上等于旧 Summary 的 `SPKIDERDigest`，AKI bytes 数值上等于新 `DeclaredIssuerSPKIDERDigest`。这是同 preimage、同算法的数学对应，不批准把命名类型、证书扩展、数据库 `PublicKeyDigest`、issuer ID 或主状态引用隐式互换。

## 4. 签发者来源、互操作与后续认证责任

### 4.1 本增量只接受明确声明

新输入只有一份 `DeclaredIssuerSPKI`，由调用方显式提供。本包不通过 AKI 反查 key，不从 leaf 的 subject key、issuer DN、serial、未认证链、系统 roots 或任意自带 CA 自动选择 issuer。

本增量不接受 caller-supplied verified flag、issuer ID、预计算 AKI、受信任摘要、证书链或 provenance callback。输入 SPKI 能解析且 AKI 匹配，并不证明其来源。

未来要完成真实 issuer 绑定，必须用另行批准的组件，从部署授权、验签通过且受版本/high-water 约束的 trust-bundle 中选取有正确角色的 exact CA DER，并证明以下关系：

1. 从同一张 exact CA DER 取得 exact issuer SPKI，逐字等于本增量的声明输入。
2. CA 满足独立批准的 CA profile、用途、有效性、约束及信任链要求。
3. 使用该 CA 的实际公钥验证同一份 exact leaf 的签名，并检查所需 issuer 命名/签发关系。
4. CA 的 SKI、leaf 的 AKI 和本附录派生值满足同一套已批准 profile。
5. 继续完成可信时间、deauthorized/high-water、身份状态与运行授权责任。

这五项是后续责任，不是本包可通过读取声明值来标记完成的检查；本附录没有设计其 provider 或运行接口。

### 4.2 不自动迁移已有 CA 标识

RFC 5280 §4.2.1.2 要求，CA 证书中的 SKI 与该 CA 签发证书的 AKI keyIdentifier 对应；已有标识不能被本包自行重算替换后声称原链仍满足相同 profile。[RFC 5280 CA SKI/AKI 对应](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.2.1.2)。

因此，采用本附录的未来签发/部署流程必须事先确认 CA profile 和 issuer provider 使用同一 SPKI-SHA256-32 规则。已有 20-byte 或其他算法的 CA SKI 不会被视为本新 profile 的合规 issuer；不得截断、同时接受多算法、忽略 CA SKI 或改写现有证书补齐。

本关系包不接收 CA DER，不能检测 CA SKI 兼容性；此检查属于未来完整 issuer 认证层。仅因本包接受 leaf/声明 SPKI，不能宣称某外部 CA 或历史部署已经兼容。若不兼容，必须另行设计 provider/profile 迁移，不能隐式 fallback。

已核对本地固定 Go 1.26.5 `crypto/x509/x509.go`：自动生成 CA SKI 的路径可采用截断 SHA-256 的 BIT STRING 摘要，兼容设置可改用 SHA-1；签发 AKI 的路径还可取 parent SKI。本项目算法不能依赖这些自动行为或 `GODEBUG`。未来新测试/签发模板必须显式设置匹配的 CA SKI、leaf SKI/AKI；本轮不修改任何模板或 fixture。

## 5. 推荐接口、输入预算与声明 SPKI 子集

新包接口限定为 `Evaluate(ctx, Input) (*Material, error)`、`Material.Summary() Summary` 和无返回值的 `Material.Clear()`。不增加 DER 输出、Recheck、链验证、序列化、安装或 Ready/Authorize 接口。

`Input` 恰好包含：

| 字段 | 含义 |
| --- | --- |
| `Leaf nodeleaf.Input` | 同旧输入：exact leaf DER、NodeID、TrustDomain、ExpectedSPKI；不接受 Summary 替代 |
| `DeclaredIssuerSPKI []byte` | 一份调用方声明的 exact DER SubjectPublicKeyInfo，1–8,192 bytes；非 CA 认证材料 |

旧 Leaf 外层预算全部保留：DER 1–65,536 bytes、ExpectedSPKI 1–512 bytes，UUID/domain 的规则不变。三个 byte 输入总长度最多 74,240 bytes；所有外层预算在任何输入复制或解析之前检查。此值是输入预算，不是 Go parser 的总堆内存上限。空字段是无效输入，超长 byte 字段是预算超限。

声明 issuer SPKI 的首版编码支持范围建议为：RSA、命名曲线 ECDSA（P-256/P-384/P-521）、Ed25519。必须用固定 Go 1.26.5 完整解析，随后用其 `MarshalPKIXPublicKey` 输出进行逐字一致性检查；不相等即拒绝，绝不把重编码输出替换输入再接受。哈希始终使用 owned 原始输入。

这限定了首版可检查的 SPKI 表示，不批准任何 CA 签名算法、RSA 位数政策、CA 模板或生产安全等级。RSA 只要求该 parser 接受且 canonical 表示一致，不在本增量引入 CA 最小位数规则。ECDSA 点必须是指定命名曲线的合法非零点；Ed25519 只证明 parser 所检查的编码形状，不证明私钥持有或完整点/子群策略。

不接受 DSA、X25519、P-224、显式 EC 参数、压缩 EC 点、未知公钥类型、PEM、PKCS#1 裸 key、拼接或尾随材料。不能从 leaf 的 P-256 要求推导出 issuer 也必须是 P-256。未来 CA 若需要超出此表示子集的算法，须另行审阅扩展该子集；不能静默跟随工具链新增支持。

旧 leaf 的 ExpectedSPKI 仍要求 canonical named P-256，与该 leaf raw SPKI exact 相等；上述 issuer 子集不得放宽旧 subject key 规则。

## 6. 检查顺序、错误与所有权

### 6.1 有界、确定的检查顺序

推荐顺序固定为：

1. context 非 nil 且未取消；先检查三个 byte 字段是否有任一超出外层预算，再检查空值与旧 UUID/domain 纯值形状。预算超限优先于空值/纯值错误，保留旧 preflight 的优先级。
2. 创建私有 owner 并立即建立失败/panic 清理责任，然后分别复制 leaf DER、expected SPKI、declared issuer SPKI。
3. 对 owned Leaf 输入调用 `nodeleaf.Parse`；取得 lower Material 后立即安排 Clear，读取 detached Summary，要求其 Proof 与固定六项 Pending 正确。此调用保留旧错误和检查顺序。
4. 完整解析 owned declared issuer SPKI，检查 §5 表示子集与 canonical 逐字一致性。
5. 仅对同一份未改变的 owned leaf DER 重新取得临时证书字段；完整消费、raw DER 与 owned DER 逐字相等、raw SPKI 与 owned expected SPKI 逐字相等；两项 exact 摘要与 lower Summary 对应。读取 SKI/AKI 供后续比较：复用的是同一份材料的 lower 形状结论，不要求旧 Summary 导出它没有的标识字段。不得重新读取调用方 buffer，也不复制一套独立 leaf profile。
6. 按 SKI 后 AKI 的顺序，先检查两者均为 32 bytes，再比较派生值；计算 leaf/声明 issuer 的 exact SPKI 摘要。
7. 最终 context 检查，构造固定 Summary，移交成功 owner；所有 lower Material 和临时 marshal buffer 都清理。

各阶段边界检查取消；同步 Go 解析/重编码/哈希调用不能被内部即时中断。预算限制输入工作量，不承诺硬实时取消延迟。panic 继续向上传播，不恢复成成功。

这会短暂存在新包 owned 输入与 lower owned 输入的两份拷贝；它们都受外层预算限制，各自有明确清理责任。只解析调用方可变内存、解析后才复制或先克隆再建立清理，均不允许。

### 6.2 固定错误类别

沿用旧 `nodeleaf` 五个 sentinel 的 `errors.Is` 身份；新包可导出同身份的别名，不能包装 parser 的动态错误。仅新增固定 `ErrUnsupportedKeyIdentifierProfile`：

| 类别 | 含义 |
| --- | --- |
| `ErrInvalidInput` | context 为 nil、纯值输入无效；声明 issuer SPKI 无法完整解析、不在表示子集或非 canonical |
| `ErrLimitExceeded` | 三个 byte 输入之一超出外层预算 |
| `ErrInvalidMaterial` | lower leaf 原始材料失败，或对同一 owned leaf 的临时字段解析/对应失败 |
| `ErrUnsupportedLeafProfile` | lower leaf 超出已批准的形状子集 |
| `ErrUnsupportedKeyIdentifierProfile` | lower shape 已通过，但 SKI 或 AKI 不是恰好 32 bytes |
| `ErrBindingMismatch` | lower URI/SPKI 绑定失败，或两个 32-byte 标识之一与其派生值不相等 |
| `context.Canceled/DeadlineExceeded` | 阶段取消或到期 |

错误不包含 DER、URI、标识、公钥、节点、issuer、serial、摘要或签名值；不记录日志。lower 已在前面拒绝的 malformed SKI/AKI 仍返回 lower 类别，不人为改成新类别。

§6.1 第 6 步先按 SKI、AKI 顺序完成两项长度检查，再按 SKI、AKI 顺序比较；例如 SKI 的 32-byte 值错误且 AKI 长度 20 时，先返回标识 profile 错误。取消检查在阶段开始和结束执行，结束时若已取消，取消错误优先于该阶段记录的普通错误；否则按固定顺序返回首个错误，不随 map 遍历变化。

### 6.3 输出与生命周期

`Material` 无导出字段，只引用一个私有共享 owner。Summary 限定为值字段：

- 新 `Proof` 枚举：`None` 或 `NodeLeafDeclaredKeyIdentifiersBound`。
- `Leaf nodeleaf.Summary`：同一材料的 detached lower Summary，包含原形状 Proof 与原固定六项 Pending。
- `SubjectKeyIdentifier KeyIdentifier`、`AuthorityKeyIdentifier KeyIdentifier`。
- `DeclaredIssuerSPKIDERDigest contracts.Digest`：明确标注声明值来源，不能简称 TrustedIssuerDigest。

新包不导出第二份 Pending 以免与 nested Leaf 冲突。有效结果同时满足新 Proof、旧 Leaf Proof 和原六项 Pending；所有命名类型都只描述部分事实，不是可验证 capability。Summary 可复制、可伪造，不可作为新输入代替材料。

owner 保留自有三个 byte buffers 与固定 Summary，不保留 Go 证书、公钥、big.Int、parser 切片或 lower Material。方法 nil-safe；包装拷贝共享生命周期。任一 Clear 后所有包装返回零 Summary，包括两个 Proof 均为 None；零 Pending 不代表责任已满足。

失败、取消及正常 panic unwind 清零仍可达的自有 mutable buffers、临时 marshal buffer 与 Summary；成功后由 Clear 承担同样责任。Clear 幂等，不影响输入或先前 detached Summary。不承诺擦除 parser/allocator 的不可达分配、immutable strings、运行时副本、调用方值或 fatal OOM/系统崩溃后的内存。

不支持调用方在 Evaluate 中并发修改输入、不支持并发 Clear/read；不新增锁、callback/global hook 或并发安全承诺。

## 7. 后续实施的独立向量与拒绝测试

本节是未来验收要求，不表示本轮运行了任何测试。实施计划另列文件清单、逐条限定命令、记录要求、回归面与执行方法。

### 7.1 不依赖生产实现计算期望值

使用 RFC 7093 §3 的公开 P-256 SPKI 作为独立 subject 派生 oracle。91-byte exact DER 的 hex 为下列连续 bytes（展示换行不属于输入）：

```text
3059301306072a8648ce3d020106082a8648ce3d03010703420004
7f7f35a79794c950060b8029fc8f363a28f11159692d9d34e6ac948190434735
f833b1a66652dc514337aff7f5c9c75d670c019d95a5d639b72744c64a9128bb
```

其完整 SPKI SHA-256 期望为 `6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405`。应直接使用该固定值，不能调用产品派生函数建立期望值。[RFC 7093 公开例子](https://www.rfc-editor.org/rfc/rfc7093.html#section-3)。

该向量只检验派生函数/输入编码，不要求取得其私钥。端到端正例另用固定、仅测试的 subject/issuer key 真实构造 DER，显式设置 SKI/AKI，保留独立 golden bytes/digests。不得把旧 20-byte arbitrary ID fixture 改成新 profile 来制造旧回归通过：它们继续用于旧 shape 测试，新包对它们应拒绝。

### 7.2 必须覆盖的关系与边界

- 正例：P-256 subject；声明 issuer 分别为 P-256、P-384、P-521、RSA、Ed25519 的合规 SPKI 表示；这些只测试声明关系，不声称相应 CA 已批准。
- 同 key/different role：声明 issuer SPKI 恰等于 leaf SPKI，且两个标识相等时，关系本身可以通过，不能据此证明 self-signing 或信任。
- SKI/AKI 各自 1、20、31、32、33、64 bytes；形状合法的非 32-byte 值拒绝；1–64 之外继续由 lower 拒绝。
- 分别篡改 SKI、AKI、expected subject SPKI、declared issuer SPKI；对新 profile 的真值和身份来源不得混淆。
- SHA-1、截断 SHA-256、BIT STRING 全 SHA-256、EC-point-only SHA-256、hex/base64 文本、leaf-DER 摘要、加前缀、互换 subject/issuer preimage，均不能成为相同关系的替代算法。
- 声明 SPKI 的空值、8,192-byte 上限与超限；PEM、裸 PKCS#1、尾随、非最小 DER、算法参数表示差异、非法 EC 点与未支持 key 类型均拒绝。对共享 subslice/输入别名，复制之后的内部事实不能依赖调用方 buffer。
- lower 材料错误优先于声明 SPKI 解析；长度检查优先于标识比较；两项值错误采用 SKI 优先；阶段取消不得留下有效半成品。
- nil receiver、零值、包装拷贝、重复 Clear、detached Summary、失败与 panic 清理结构；不使用产品注入 hook，不宣称测得不可达内存擦除。
- 签名密码学错误、expired/future 但 lower 窗口合规、未经信任包声明的 issuer，仍可能通过关系检查；每个结果始终保留六项 Pending。
- 输入上限先于克隆/解析，类型/API/消费者隔离守卫以及禁止认证、I/O、时钟和运行接入的调用检查。

RSA/Ed25519 正例不必把它们用于真实 CA；必须至少另有 issuer/subject 不同的真实签发 DER 正例。测试可用离线签发操作，产品包不得签发证书、读取私钥或获取随机数。

## 8. 后续实施的唯一旧边界例外

旧 `nodeleaf` 设计与 `internal/nodecontrol/nodeleaf/boundary_test.go` 目前禁止任何外部产品包导入 `nodeleaf`。推荐组合方案不能假装这条限制不存在。

若本书面附录及后续实施计划获得批准，计划仅可提议以下一个旧源码改动：在该文件的 `leafConsumerAllowed` 中，增加 exact directory `internal/nodecontrol/nodeleafkeys` 且 package 恰为 `nodeleafkeys` 的导入例外（含同包测试）；其余原例外保持不变。不是前缀包通配，不允许任意 `_test.go`、兄弟包、外部测试 package 或运行入口成为额外消费者。

新包自身也须有消费者守卫，禁止其他产品包导入或调用它。除上述单一消费者例外，不改旧公开 API、旧产品代码、其他旧守卫、测试向量、SQL/迁移、serving/store/主状态结构、配置或实际运行路径。

本文件获准只意味着上述方案可进入计划审阅，不表示今天已经修改该守卫或放开运行接入。若用户不接受这一明确例外，应重新审阅架构选择，而不是悄悄复制旧检查或放宽全体隔离规则。

## 9. 非目标、阶段交接与书面自检

本附录不完成 CA profile、真实签名/信任链、部署授权 trust-bundle、issuer/serial 唯一性、可信当前时间、private-key possession、CSR/keystore、身份状态、主状态封存/恢复或实际运行授权。Task10/B01/C12/main 的既有验收阻塞保持不变。

本轮只新增本文件，不改 current-status、旧 spec/plan/execution、`.superpowers` 台账或任何源码；既有未提交进度保持原样，不改变 Git 配置、HEAD、索引，不提交/推送。只进行文档与基线的只读核对，不新增测试记录或“全部通过”结论。

采用 `superpowers:brainstorming` 的架构路径：本文件先交用户审阅；书面批准后才编写实施计划，计划须另行审阅并选定执行方法，之后才可实现。当前“要”只批准起草附录，不批准尚不存在的计划和代码。

书面自检：

- 算法：固定完整 DER SPKI、SHA-256 全 32 bytes，extension 内层编码与独立 oracle 明确；不依赖 Go 默认值，不协商 fallback。
- 来源：AKI 仅绑定声明 issuer；CA SKI 兼容、真实签名、trusted source 的责任分开，六项 Pending 不消失。
- 接口：原输入/结果强度保留，新包使用材料而非 Summary；预算、编码支持子集、错误优先级与共享 owner 生命周期明确。
- 范围：新包加一个明确的旧消费者例外是未来计划的推荐范围，不是本轮代码修改；没有主状态或运行接入。
- 阶段：完整草案已写；尚未获书面批准，未实施、测试、审查代码或进行 Git 写入。
