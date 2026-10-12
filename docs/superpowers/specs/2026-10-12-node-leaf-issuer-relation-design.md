# Node CA—叶证书直接签发关系设计

日期：2026-10-12

状态：完整书面草案，待用户审阅批准；尚无实现或测试结论。

工作区：既有 `codex/c12-b01-task9-coordinator` worktree；起草基线为
`2877f89cc9c5a6ca0514d79c54826627a112d412`。该基线是已推送的开发检查点，
不是 main 或 C12 验收基线。

授权理解：用户“批准”允许按上一条建议制定本设计，不等于批准尚未存在的
实施计划、产品实现、运行接入或 main 合并。本轮只新增本文件；可以将本文件
作为待审草案做本地文档提交，不推送、不改其他进度文件。

## 1. 目标与已有边界

在已有部分材料检查之后，新增独立、离线的
`internal/nodecontrol/nodeissuer`，对同一组自有原始输入证明：

1. node leaf 仍满足旧材料形状、节点/信任域/subject key 声明与密钥标识规则。
2. 所提供的 CA 仍满足已批准的 CA 材料子集及 subject key 声明。
3. CA raw subject SPKI、CA SKI、leaf AKI 与本次声明公钥对应，leaf raw issuer
   Name 逐字等于该 CA raw subject Name。
4. 使用这张 CA 的实际 P-256 subject public key，验证这张 exact leaf 的
   ECDSA/SHA-256 签名。

这是 **direct issuer signature relation**，不是 CA 来源认证、完整证书链验证
或运行授权。“该公钥验证了签名”不升级为“受信任机构确实签发并授权了此身份”。
输入声明仍来自调用方；没有可信根、部署来源、当前时间或身份状态证据。

依据为 [Node PoP 父设计](2026-08-23-node-pop-control-plane-design.md) §6.1–6.3、
[leaf 材料设计](2026-10-11-node-leaf-material-design.md)、
[声明密钥标识设计](2026-10-11-node-leaf-key-identifiers-design.md) 和
[CA 材料设计](2026-10-11-node-ca-material-design.md)。

现有源码事实：`nodeleaf.Parse`、`nodeleafkeys.Evaluate`、`nodeca.Parse` 各自
拥有输入并输出部分事实；Summary 是可复制、可伪造的描述，不是认证能力。
`nodeleaf.Material.DER()` 已存在并返回克隆，另两包没有原始字节输出。
新层不调用这个 DER accessor，不添加任何旧 accessor，也不接受 Material 或
Summary 作为输入。旧三包的产品实现、公开 API、Proof、全部 Pending 不变。

真实 CA/provider 的生产兼容性继续为 UNKNOWN；本设计不自动迁移既有 20-byte
标识、证书、签发模板或 trust bundle。Task10/B01/C12/main 的阻塞继续保留。

## 2. 方案与组成

推荐独立组合层；另外两种方案仅用于明确取舍：

| 方案 | 优点与代价 |
| --- | --- |
| 独立 `nodeissuer`（推荐） | 复用真实 lower 检查，新增关系和验签责任；需三个明确的旧消费者守卫例外，但不修改旧产品能力 |
| 在 `nodeca` 中加入 leaf/验签 | 少一个包，但改变已冻结 CA API、职责和禁止验签边界，不采用 |
| 一次实现完整 chain/trust/runtime | 能给更完整结论，但依赖部署包、可信时间、状态及私钥来源，超出本轮范围 |

新产品包只依赖所需标准库、现有 UUID、`contracts` 及上述三个 lower 包。
没有 provider、算法注册器、验证 callback、可配置 profile 或 issuer discovery。

固定数据流：预算和纯值预检查 → 独立复制四项 byte 输入 → `nodeca.Parse` →
`nodeleafkeys.Evaluate`（其内部调用 `nodeleaf.Parse`）→ 对同一自有 DER 取得
临时字段 → raw/parser/lower 一致性及签名表示检查 → 公钥/名称/标识关系 →
实际验签 → 最终取消检查 → 移交结果。

lower 每次都是真实调用；不接受外部声明的 lower 成功结果，不增加重复的 leaf
六扩展或 CA 四扩展规则。新增 scanner 只负责新层所读取的签名 envelope 和字段
边界，不从旧包抽取共享 parser 或复制一套一般 ASN.1 框架。

## 3. 输入、预算与声明角色

公开入口固定为 `Evaluate(ctx context.Context, in Input) (*Material, error)`。
Input 恰好有两个字段：

```go
type Input struct {
    Leaf nodeleaf.Input
    CA   nodeca.Input
}
```

四项 byte 输入预算如下；任一超限先于空值、纯值检查、复制或解析：

| 字段 | 字节预算与角色 |
| --- | --- |
| `Leaf.DER` | 1–65,536；一张 exact node leaf certificate DER |
| `Leaf.ExpectedSPKI` | 1–512；调用方预置的 leaf subject SPKI |
| `CA.DER` | 1–65,536；一张调用方提供的 exact CA certificate DER |
| `CA.DeclaredSPKI` | 1–8,192；该 CA 自己的 subject SPKI，而非其 parent key |

四项 byte 输入合计最多 **139,776 bytes**。UUID 为固定值类型；TrustDomain 仍为
1–253-byte 小写 ASCII DNS 声明，每 label 1–63、首尾为 `[a-z0-9]`、内部仅
`[a-z0-9-]`；NodeID 非零且 RFC4122 variant。此合计不包含字符串/值类型开销，
也不是总堆内存、parser 临时分配或取消延迟上限。

新层只传递自有快照：`nodeca.Parse` 接收 owned CA 输入；
`nodeleafkeys.Evaluate` 的 `Leaf` 接收 owned leaf 输入，
`DeclaredIssuerSPKI` 恰为 owned `CA.DeclaredSPKI`。因此不新增第二份 caller
issuer 声明或第五项字节输入；随后仍比较 CA raw subject SPKI 与该声明的 exact
bytes，不能用“同一个变量”替代实际材料对应检查。

不接受 chain 数组、PEM、私钥、issuer ID、verified flag、摘要替代材料、trust
pool、时间参数、provenance callback 或预计算密钥标识。两种声明公钥首版都
必须是既有 profile 要求的 canonical named P-256 SPKI，实际有效长度为91 bytes。
不会放宽 `nodeleafkeys` 原有其他算法表示范围，也不会据此批准那些 CA 算法。

## 4. 输出与证明强度

`Proof uint8` 恰为 `None=0` 和 `NodeLeafDirectIssuerSignatureBound=1`。
Material 恰为含私有共享 owner 的包装，只公开 `Summary() Summary`、`Clear()`。
新层没有 DER/SPKI/signature 输出、Verify/Ready/Authorize/Recheck、序列化或安装。

Summary 恰有以下值字段：

```go
type Summary struct {
    Proof               Proof
    Leaf                nodeleafkeys.Summary
    CA                  nodeca.Summary
    LeafTBSDERDigest     contracts.Digest
    LeafIssuerDERDigest contracts.Digest
}
```

`Leaf` 是同一材料的 key-identifiers lower Summary，其 `Leaf` 内仍有完整的旧
`nodeleaf.Summary`；`CA` 是同一 CA 的 lower Summary。新层不再另设第三张 Pending
表，避免把复合未决责任误当成可逐项删减的验收清单。

每个有效结果必须同时满足：新 Proof、
`Leaf.Proof == nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound`、
`Leaf.Leaf.Proof == nodeleaf.NodeLeafMaterialShapeValid`、
`CA.Proof == nodeca.NodeCAMaterialShapeAndDeclaredKeyBound`，以及下面两张原表。

leaf 原 `[6]Obligation` 按原顺序完整保留：

1. `LeafSignatureAndIssuerTrust`
2. `KeyIdentifierDerivationAndIssuerBinding`
3. `TrustedCurrentTime`
4. `IdentityAndAuthorityProvenance`
5. `PrivateKeyPossession`
6. `MainStateAndRuntimeBinding`

CA 原 `[6]Obligation` 按原顺序完整保留：

1. `CertificateSignatureAndIssuerChain`
2. `AuthorityIdentifierAndParentBinding`
3. `TrustedCurrentTime`
4. `DeploymentAuthorizationAndPurpose`
5. `BundleStatusAndRollback`
6. `LeafIssuerRelationAndRuntimeUse`

新 Proof 增加本设计明确检查的子事实；不表示这些复合责任整体已完成。
CA 自身签名、CA AKI 与 parent、可信根、完整链/path constraints、当前有效性、
可信来源/purpose、bundle status/high-water、issuer/serial 唯一性、身份/authority
epoch、私钥持有、主状态和 runtime 仍未证明。不得回写或自动升级旧结果。

`LeafTBSDERDigest` 是 exact 原始 TBS DER 的 SHA-256，包括 TBS tag/length；
`LeafIssuerDERDigest` 是 exact raw issuer Name DER 的 SHA-256。它们是描述与
审计关联值，不是输入能力、唯一身份或可信来源。不能仅比较摘要来代替 raw
关系；不能仅凭相同 TBS 摘要认定两张包含不同签名的证书为同一 exact DER。

## 5. 直接关系与首版签名规则

### 5.1 同一材料的对应

lower 成功后，对 owned CA DER 与 owned leaf DER 再调用固定 Go 1.26.5
`x509.ParseCertificate`。要求两者的 `Raw` 各自逐字等于 owned 输入。
临时字段与 lower 的全部描述对应：相关 raw DER/SPKI/Name 摘要、SKI/AKI、
serial/长度、NotBefore/NotAfter，以及 CA pathLen 的省略/零区分与 CRLSignAllowed。
leaf 的节点/域声明与 lower 值对应；不能拼接不同材料的 Summary。

leaf 专用 envelope scanner 严格消费单个外层 SEQUENCE，其三字段为 TBS、
outer AlgorithmIdentifier、signature BIT STRING；长度为最短 definite DER，
拒绝高 tag、indefinite length、越界/溢出、尾随和额外字段。TBS 按 v3 固定字段
顺序消费 version、serial、inner AlgorithmIdentifier、issuer、validity、subject、
SPKI、extensions，拒绝 unique identifiers/额外字段；旧 lower 仍负责各 profile。
取得的 TBS、issuer、subject、SPKI、signature 原始范围须分别与 parser 字段逐字
对应。新检查不重新编码 TBS/Name，不从 parser 的名称字符串重建原始材料。

### 5.2 公钥、名称及标识

固定关系顺序为：

1. CA `RawSubjectPublicKeyInfo == owned CA.DeclaredSPKI`。
2. leaf `RawIssuer == CA.RawSubject`，比较包括 tag/length 的完整 raw DER。
3. leaf AKI == CA SKI == SHA256(exact CA raw subject SPKI)，完整32 bytes；
   同时与两个真实 lower 结果对应。leaf SKI 仍按旧层绑定 leaf subject SPKI。

逐字 Name 对应是项目封闭子集的新增取舍，比一般 DN 等价更窄：显示文本相同
但字符串 tag、空白、大小写、RDN 编码/顺序等 bytes 不同，仍拒绝。不调用
`pkix.Name.String()`、文本归一化或仅比较 CN。RFC 5280 的一般 name chaining
涉及名称匹配；本设计不声称实现其完整名称等价算法。
[RFC 5280 issuer/name chaining](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.1.2.4)。

CA AKI 指向 CA parent，不要求它等于 leaf AKI 或 CA SKI。Subject==Issuer、
CA SKI==CA AKI 不证明 self-signed/root。相同 subject Name 和公钥可出现在不同
CA certificate 中；签名运算不能区分其不同 serial、parent、有效期或部署身份。
成功描述本次提供的 exact CA 材料和其公钥关系，不证明 issuer certificate 的
来源唯一性或可信 `issuer_id`。同公钥跨 leaf/CA 角色也不由本层判断部署隔离。

### 5.3 叶证书签名编码与算法

profile 名仅为包内文档辨识：`node-leaf-direct-issuer.p256-sha256.v1`，不是 wire
字段或协商参数。首版仅接受 leaf `ecdsa-with-SHA256`，OID
`1.2.840.10045.4.3.2`，inner 与 outer AlgorithmIdentifier 都须逐字为：

```text
30 0a 06 08 2a 86 48 ce 3d 04 03 02
```

参数必须省略，不接受 NULL；不接受 SHA-1/SHA-224/SHA-384/SHA-512、RSA、
Ed25519、DSA 或 Unknown。固定 parser 的 SignatureAlgorithm 同时必须是
`x509.ECDSAWithSHA256`。这些是新层算法政策，不能倒推出旧 leaf 材料层原来
已经限制算法。[RFC 5758 ECDSA/SHA2](https://www.rfc-editor.org/rfc/rfc5758.html#section-3.2)。

进入本层 AlgorithmIdentifier 检查后，先要求 SEQUENCE 恰含一个完整 OID 与
至多一个参数 TLV：OID 为非空、每个 base128 分量最短且终止，不以 Go int
大小额外限制 arc；参数的外层 tag/最短 length/边界须完整消费，NULL 若出现
须恰为 `05 00`，多参数或已读取编码错误先返回 malformed。其他非许可参数
只检查上述外层表示，拒绝为 unsupported；不承诺解析 opaque 参数内部 schema
或所有未知 ASN.1 primitive 语义。目标算法本身只在 exact 无参数 bytes 时接受。

signature BIT STRING 的 unused-bits byte 恰为0，其内容完整消费且非空。签名
payload 是一个 exact DER SEQUENCE，恰含两个最小 DER INTEGER `r,s`，没有
尾随、额外字段或 P1363 裸拼接；二者均满足 `1 <= value < P-256 order N`：

```text
N = ffffffff00000000ffffffffffffffffbce6faada7179e84f3b9cac2fc632551
```

payload 长度8–72 bytes；INTEGER content 最多33 bytes，只允许正数所需的单个
sign-padding，不能接受负数、零、冗余零或截断。新层显式检查最小编码和范围，
不把布尔验签失败当作所有表示错误的替代。r/s 比较使用固定32-byte unsigned
magnitude 与固定 N，不把外部 key/回调作为曲线参数来源。
[RFC 3279 ECDSA signature representation](https://www.rfc-editor.org/rfc/rfc3279.html#section-2.2.3)。

不新增 low-S 政策：所有合法范围内且验签通过的 high-S/low-S 表示均可接受，
不将 s 替换为 N-s、不重签、不修改 exact DER。此选择不保证签名表示唯一；
证书关联继续以 exact leaf DER 摘要为准。

### 5.4 实际运算和刻意不检查的事实

仅从同一临时 CA Certificate 取得实际 `*ecdsa.PublicKey`；要求非 nil、
Curve 为 `elliptic.P256()`、坐标非 nil 且在曲线上，不能使用 leaf key 或另一
声明 key 代替。计算 `SHA256(exact owned leaf RawTBSCertificate DER)`，调用
`ecdsa.VerifyASN1`。false 返回固定签名错误，没有算法 fallback。
不得 hash 整张 leaf、仅 TBS content、重新 marshal TBS、signature 或 CA DER。
这是 Go 的 hash/ASN.1 签名验签原语，不自行实现椭圆曲线运算。
[Go VerifyASN1](https://pkg.go.dev/crypto/ecdsa#VerifyASN1)、
[RFC 5280 signed input](https://www.rfc-editor.org/rfc/rfc5280.html#section-4.1.1.3)。

CA 证书自身的签名算法属于其 parent，本层不验证它，也不根据 CA subject key
强行限定它。CA 自身形状合法但密码学错误的签名仍可能通过本组合层。
不调用 x509 Verify/CheckSignature*、CertPool/SystemCertPool，不加载系统 roots，
不自动寻找 parent；keyCertSign/CA 字段由真实 nodeca lower 检查。

不读取当前时钟；不增加 leaf 窗口包含于 CA 窗口、两者窗口重叠、issuer_now
backdate 或可信当前有效性政策。旧各自窗口形状仍须通过，过期/未来或跨 CA
窗口的材料仍可能验签成功。pathLen 省略/零保留，但本层不检查完整路径。
以上不是给运行入口的豁免，而是仍待独立时间/信任认证完成的责任。

## 6. 检查顺序、错误和生命周期

### 6.1 固定顺序与取消

1. nil context 返回本层 ErrInvalidInput；非 nil context 已取消/到期时先返回
   原 context 错误。检查四项上限，再检查空值和 UUID/domain；超限优先。
2. 创建空 owner，在首次克隆之前登记失败与正常 panic unwind 的清理，分别
   独立克隆四项 byte 输入；重叠 subslice 也不能共用 backing storage。
3. 对 owned CA 调用 nodeca.Parse；立即 defer Clear，检查固定 Proof/Pending。
   本阶段原 CA 错误优先于尚未执行的 leaf 错误。
4. 调用 nodeleafkeys.Evaluate，使用 §3 的 owned 参数；立即 defer Clear，
   检查两级 leaf Proof 和原六项 Pending，保留该 lower 原检查顺序与错误。
5. 按 CA 后 leaf 的顺序重新取得临时 parser 字段，检查 §5.1 的快照对应。
   leaf scanner 先检查完整 envelope/inner-outer 一致性，再检查 §5.3 的算法和
   signature 表示；这些表示/profile 错误优先于尚未执行的关系或密码学错误。
6. 按 §5.2 公钥、名称、标识顺序检查；不能随 extension/map 顺序变化。
7. 进行 §5.4 的 hash/验签，构造 detached Summary，最终检查 context 后才设置
   valid 并移交 owner；所有 lower Material 在返回前清理。

每个阶段在开始与结束检查 ctx.Err；结束时观察到取消，context 错误优先于该
阶段的普通错误。同步 Go 操作无法即时中断，不承诺硬实时取消；panic 向上
传播，不恢复为成功、不返回部分结果。不增加 goroutine、锁、callback 或全局 hook。

### 6.2 固定错误

本层以下六个 sentinel 独立命名，错误文本固定，无动态材料：

| Sentinel | 固定文本和分类 |
| --- | --- |
| `ErrInvalidInput` | `node issuer: invalid input`；nil context、空值或纯值声明无效 |
| `ErrLimitExceeded` | `node issuer: limit exceeded`；任一外层 byte 输入超限 |
| `ErrInvalidMaterial` | `node issuer: invalid material`；lower 描述/快照对应异常、raw envelope/inner-outer 不一致、签名 malformed/nonminimal/非零 unused bits |
| `ErrUnsupportedSignatureProfile` | `node issuer: unsupported leaf signature profile`；结构完整、inner/outer 相同但非 exact 许可 AlgorithmIdentifier（含 NULL 参数），或 canonical r/s 为零、负值、>=N/超范围 |
| `ErrIssuerRelationMismatch` | `node issuer: issuer relation mismatch`；本层公钥、raw Name 或标识关系不同 |
| `ErrInvalidSignature` | `node issuer: invalid leaf signature`；前述检查均通过，实际验签 false |

nodeca/nodeleafkeys 的原 sentinel 直接返回，不包装、不映射成本层同名 error；
context.Canceled/DeadlineExceeded 亦原样返回。errors.Is 身份保留，调用方不能
假定来自不同包的同名 sentinel 相等。固定 Go parser/lower 已提前拒绝的输入
仍返回其原错误，不为凑新表而改写前置分类。CA/leaf 双错误采用上面阶段顺序。

本层已进入签名 scanner 时，缺字段/错误 tag/不最短编码/尾随属 malformed；
单个 canonical 负数或零整数有可解释的值但不在签名政策内，属 unsupported。
非最小负数仍是 malformed。NULL AlgorithmIdentifier 须是完整正确 NULL；
参数自身 malformed 仍属 malformed，不以“算法不支持”遮盖已读取的编码错误。
其他结构完整但非许可 AlgorithmIdentifier 的参数不宣称做完整未知 schema 验证；
实施计划须冻结这一错误分类的正反例，不扩大旧 CA 的未知 signer 语义承诺。

错误不包含 certificate、DN、URI、NodeID、serial、公钥、摘要、标识、签名、路径
或 parser 动态文本；不日志记录、不持久化。

### 6.3 自有状态与清理

owner 只保留 owned leaf DER、expected leaf SPKI、owned CA DER、declared CA
SPKI、固定 Summary 和 valid 状态。不保留 Certificate、公钥/big.Int、raw scanner
指针或 lower Material。parser 临时引用只能落在本层 owned buffer 上；lower
临时副本由其自身 Clear 承担。四个 buffer 独立；不以 raw accessor 取得新材料。

nil/零值 Material 返回零 Summary；包装拷贝共享 owner。任一包装 Clear 后所有
包装结果都为零，所有 Proof 为 None；零 Pending 不表示责任完成。Clear 幂等，
不改调用方输入、先前 detached Summary 或其他独立 Evaluate 的结果。

失败、取消、正常 panic unwind 和 Clear 清零仍可达的自有 mutable buffers 与
Summary。若产生临时 mutable buffer，创建后立即登记清零。没有不可达 parser
分配、immutable string、runtime 副本、调用方数据、fatal OOM/崩溃擦除保证。
不支持 Evaluate 期间并发修改输入、不支持并发 Summary/Clear；不宣称常数时间
或 secret-memory 保证。139,776-byte 外层输入预算不等于峰值堆内存限制。

## 7. 旧消费者例外与新层隔离

只有本设计及后续实施计划都批准后，才允许修改这三个旧测试守卫文件：

| 文件 | 唯一新增消费范围 |
| --- | --- |
| `internal/nodecontrol/nodeleaf/boundary_test.go` | leafConsumerAllowed 仅增 exact nodeissuer 目录/package，原 nodeleafkeys 例外和所有旧案例保留 |
| `internal/nodecontrol/nodeleafkeys/boundary_test.go` | keysConsumerAllowed 对其所管的 nodeleaf/nodeleafkeys 仅增同一 exact nodeissuer 例外 |
| `internal/nodecontrol/nodeca/boundary_test.go` | guardConsumer 仅为 exact nodeca import 增同一 exact nodeissuer 例外，nodeca 子路径不放开 |

exact 指仓库根下 `internal/nodecontrol/nodeissuer` 直接目录且 package 恰为
`nodeissuer`，含该同包测试；不接受目录前缀、nested、兄弟、外部路径、伪造同名
package 或 `nodeissuer_test`。不删除原正反例，新增相对/绝对规范路径正例与上述
拒绝控制。旧包的产品 import/member、禁止 Verify/CheckSignature*、API、文件
inventory、清理守卫及其他消费者拒绝规则完全不变。

新 nodeissuer 自身禁止任何额外产品消费者；同包测试例外也须绑定 exact 目录，
不能宽泛放行所有 `_test.go`。新固定 API、文件清单和产品 import/member 守卫
由实施计划冻结；仅在新层准许精确 `ecdsa.VerifyASN1` 原语，不泛开 `.Verify`。
拒绝点/空白 alias、重复 import/alias、alias 局部遮蔽及伪装 selector/method-value
绕过；准许正常显式命名 alias 的对照例。静态守卫是有限检查，不是形式证明。

产品禁止私钥/随机数/签发、time.Now、文件/网络/数据库/provider、root pool、
x509 Verify/CheckSignature*、安装/主状态/serving/runtime、序列化、goroutine、
callback、全局可变注入点及泛算法协商。测试专用离线向量构造不是产品许可。
除这三个明确守卫例外，不改旧源码或原测试数据、SQL、配置、依赖与生成代码。

## 8. 后续验收要求

以下都是未来检查，不是本轮已执行或通过的测试。

### 8.1 独立向量与正例

- 新包自己的冻结 test-only CA/leaf exact DER、声明 SPKI、TBS、签名字节及
  golden digests；至少一个不同 subject/issuer key 的真实 ECDSA/SHA-256 正例。
  期望值不由待测产品 parser、hash helper 或 VerifyASN1 产生。
- 公开原语向量固定采用下面的 RFC 6979 A.2.5，而不是在测试中生成期望签名；
  项目 CA/leaf pair 另行冻结 exact DER 和独立摘要，检验完整组合。公开消息不是
  certificate TBS，也不能代替端到端正例。没有真实 provider 私钥，不改旧 fixture，
  不在普通测试中随机重签 golden。
- 显式设置32-byte项目 SKI/AKI；不依赖 Go 自动标识或 GODEBUG。覆盖 CA KU
  两种值、pathLen 省略/0、扩展顺序变化、不同合法 DN、canonical r/s sign-padding。
- 对一份固定合法签名构造并冻结 s 与 N-s 两个表示，均可验签时均接受；两张
  exact DER 摘要不同、相同 TBS 摘要不变，不宣称唯一表示或可信 issuer 身份。

公开原语向量只取公钥和签名事实，不需要导入来源给出的测试私钥或 nonce，
也不将 RFC 的 deterministic signing 实现带入产品。下列 hex 均为完整值：

```text
message ASCII = sample
SHA256(message) = af2bdbe1aa9b6ec1e2ade1d694f41fc71a831d0268e9891562113d8a62add1bf
Qx = 60fed4ba255a9d31c961eb74c6356d68c049b8923b61fa6ce669622e60f29fb6
Qy = 7903fe1008b8bc99a41ae9e95628bc64f2f1b20c2d7e9f5177a3c294d4462299
r  = efd48b2aacb6a8fd1140dd9cd45e81d69d2c877b56aaf991c34d0ea84eaf3716
s  = f7cb1c942d657c41d436c7a1b6e29f65f3e900dbb9aff4064dc4ab2f843acda8
```

公钥与 r/s 来源为 [RFC 6979 A.2.5](https://www.rfc-editor.org/rfc/rfc6979.html#appendix-A.2.5)；
消息摘要在起草时以 Windows SHA-256 独立核对，不来自产品代码。r/s 对应的
DER INTEGER 都需要正数 sign-padding。后续计划给出冻结 DER 签名和完整 SPKI
字节；测试期望“有效”来自公开向量，不能由待测 VerifyASN1 反推建立。

### 8.2 拒绝与优先级

- nil/取消/到期 context、四字段空/上限/超限、UUID/domain 无效、overlapping
  subslice；预算超限优先于空/纯值，CA lower 失败优先于 leaf lower 失败。
- lower 原拒绝类别和原 Proof/Pending 完整回归；声明 issuer 的其他算法不会
  因旧 keys 层接受而绕过 nodeca 的 P-256-only 子集。
- raw/parser/lower 错配；单个 certificate 尾随/拼接、TBS/outer 不同算法、
  多字段/缺字段、非最短 length、高 tag、unused bits !=0、empty signature。
- ECDSA/SHA-256 无参数正例、NULL 正确表示但不支持、malformed NULL、其他
  hash/key OID、Unknown；r/s 负数、零、1、N-1、N、N+1、冗余零、必要 padding、
  截断、超长、额外 INTEGER/尾随、P1363；固定 malformed/unsupported 优先级。
- 正确 AKI/name 但错 key 签名；正确 key 签名但不同 issuer Name；相同显示 DN
  的不同 raw 编码；不同 CA subject key/声明公钥、错32-byte CA SKI/leaf AKI。
  某些错标识由 lower 更早拒绝，应保留 lower error，不能硬测新 sentinel。
- 签名表示错误加 Name mismatch，先返回表示错误；表示合法但数学签名错误
  加 Name mismatch，先返回关系错误。关系 helper 单元控制中的 Name/标识双错
  按 Name-first；公开 Evaluate 若标识错误已被 lower 先拒绝，仍保留 lower error，
  不能声称所有对外双错都采用 Name-first。
- 只改 exact TBS 的形状合法字段但保留旧签名，必须密码学拒绝；另冻结“签名
  错误 preimage（整张 leaf 或仅 TBS content）”对照，证明使用原 TBS。产品禁止
  marshal/重建 TBS，由静态/源码检查约束；若重编码恰与原 bytes 相同，验签本身
  不能区分两种实现，不能把这种测试声称为重新编码路径的动态证明。
- 错签名不会返回 Material；阶段取消也没有有效半成品。不能用测试 hook 改
  产品流程；阶段取消仅证明可观测边界，不承诺同步运算硬实时中断。

### 8.3 不升级、生命周期与隔离

- 未部署 CA、CA 自身错签名、CA parent AKI 任意形状合法值、过期/未来、leaf
  窗口超出 CA 窗口、same subject/key 的另一 CA，以及同 key 跨角色材料仍可能
  通过直接关系；每个正例都断言全部旧 Pending，不把通过误标为可信或可运行。
- 无 parent/root/可信时钟/private-key/身份状态输入，不可能声称完成相应责任。
  不新增 provider/provider compatibility 或 issuer/serial uniqueness 结论。
- nil/零值/包装拷贝/重复 Clear/detached Summary、独立 Evaluate 生命周期隔离、
  调用方输入不变、成功可达 owned buffers 清零、失败/panic 清理注册结构；
  不声称测得不可达内存擦除或并发安全。
- 三个旧消费者例外逐一做正反例、原全部案例保留；新层 API/import/member/
  inventory/消费者/清理守卫及伪装绕过控制。不能通过删除旧 guard 来造通过。
- 对有界 leaf envelope/signature 和 CA/leaf pair 突变运行限定 fuzz，检查不崩溃、
  不产生有效半成品、结果保持所有 Pending。有限 fuzz 不等于所有 DER/签名覆盖。

后续计划至少冻结：nodeissuer 全包、旧 nodeca/nodeleaf/nodeleafkeys 全回归和
已有受影响包限定回归、exact 格式/差异检查，以及上述短 fuzz 的命令与证据。
Windows 操作仍仅使用 PowerShell，工具链为缓存的 Go1.26.5，离线只读模块模式，
不安装依赖。原始 RED/GREEN、取消/超时、source pre/post hashes 与审查修订都
如实留痕；不存在的失败输出不能补造成通过。旧格式/空白欠账保留其既有阻塞。

实施后独立质量/安全审查仍需执行；本次作者设计自检不是代码独立审查。
全仓库、race/Linux/真实 DB/Docker、硬件/reboot/PITR/provider 及 C12 权威验收
不因本限定增量测试通过而被标记完成。

## 9. 阶段交接与作者自检

本轮只交付这份待审设计，不运行产品测试、不写产品代码、不修改旧 spec/plan/
execution/current-status，不安装工具、不写 Git 配置、不创建分支/worktree，
不推送/合并/清理。单文件本地文档提交只保存待审草案，不代表书面设计已批准。

采用 superpowers:brainstorming 的架构流程。用户批准的是起草阶段；审阅并批准
本文件后，才使用 writing-plans 编写实施计划。该计划还须再审阅并选定执行方法，
之后才可实现。没有把一次“批准”扩展为跨过所有后续门槛。

作者自检应核对：四 byte 输入预算准确、CA subject/parent 角色分开；真实 lower
调用而非 Summary 输入；exact Name/标识及 TBS hash/signature 关系；签名表示/
算法/数学错误分开且确定；所有旧 Proof/Pending/API 与生产 UNKNOWN 不改变；
仅三个旧消费者例外；自有输入及清理限制明确；没有完整链/时间/运行承诺，
没有预先声称实现、测试、独立审查或 C12/main 验收通过。
