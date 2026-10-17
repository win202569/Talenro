# Task11：27项请求认证与封装矩阵提案

状态：用户在ec2f55b5后允许按混合方向继续编写矩阵，并在a40c7b45后确认
本书面矩阵；**27项及11个请求角色已获本地设计批准，未授权实施或部署**。
基线：ec2f55b5ad2b0f4abae046c9e54dfb5fdcbc20ac。
本文新增11个请求角色与27项明确用途限制，不声称原规格已有这些规则。
原方向批准仅允许写矩阵；随后书面确认才批准这些局部设计，不批准部署或其余16项。

## 1. 目的、已知事实与选择

目的：让每个请求有明确的调用方、接收方、body认证与独立证据边界，防止把
response signer、caller传来的proof或目录查找当作请求授权。
范围只含[更新对照](../../roadmap/2026-10-17-task11-metadata-current-reconciliation.md)
中R组27项，不修改其余119项。矩阵批准前为103/43，书面批准后当前局部审查为
130/16；不是27项代码已实现、独立认证或canonical集合批准。

依据：[原v7规格](2026-08-24-nodecontrol-authority-v7-upgrade-design.md)（下文L编号
以本基线为准）、[已批C1](2026-10-16-authority-v7-request-auth-clarification.md)、
[已批C2方法](2026-10-16-authority-v7-metadata-admission-clarification.md)和
[已批九项澄清](2026-10-17-authority-v7-metadata-nine-item-clarification-design.md)。

已知事实：L325/327规定credential隔离和accepted key set，L278规定exact request/
bundle/response保存，L639规定特殊archive展开，L659要求attestor与live candidate
相互认证。这些约束不能单独定义本27项的request envelope角色；尤其candidate
相互认证不是对编排方request封装的定义。27项中未找到可直接登记的完整既有
请求封装契约，不能为凑出“无签名通道”比例而填null。

因此本矩阵**建议这27项均补充通用请求envelope**，按责任再细分11个角色，而非
共用一个角色。原有明确的channel、credential、attestor可信连接与业务检查全部
保留并叠加，不被请求签名替代；例如原初审#16的mutually-authenticated local
holder续租通道不在本矩阵内，继续保持原分类。混合方向是保留已明确通道、逐项
补足缺失契约，不承诺在这27项内必有null。上述27项签名选择是此次本地批准的
新增设计，不是原规格已有事实；后续若需改变须显式修订，不以未签名/响应角色
作为自动备选。

替代方案：只使用既有proof/连接身份而不冻结请求body绑定，不能关闭当前缺口；
为所有请求新建另一种认证wire协议则必须另定义transcript/传输/profile，会扩大
范围。本提案选择复用§4.1的严格编码与签名域，但不自动扩展C1的admin权限。

## 2. 共同封装及不变式（新增请求契约）

27项均采用原§4.1九字段envelope：schema、body、body_digest、signer_role、
signer_key_id、signature_policy_version、trust_root_digest、signature_algorithm、signature。
body仍为各对象原有字段，不能新增audience、caller ID、timestamp或envelope digest。
不新造schema，不改body digest或原response的request_digest：

```text
body_digest = SHA-256(ASCII("talenro.c12." + schema) || 0x00 || JCS(body))
signature_input = ASCII("talenro.c12.signature-envelope.v1") || 0x00 || JCS(metadata)
```

metadata仍为原七字段：schema、body_digest、signer_role、signer_key_id、
signature_policy_version、trust_root_digest、signature_algorithm。
Ed25519/P-256、low-S、base64url及strict JSON规则均复用原§4.1。算法必须由trusted
policy为精确role/key固定；不能由请求自选。矩阵profile不是wire字段或运行时fallback。

认证顺序是：精确schema→本矩阵唯一role→接收方预置的trusted policy/key/root/
algorithm→canonical bytes和签名→调用方身份及scope→原有proof/bundle/state检查。
请求签名只证明这个请求的认证条件，不能替代mutable state、可信时间、provider
phase、single-use nonce、commit attestation、held状态或业务authorization。
任何认证不确定结果在provider mutation或archive Fence前拒绝，不扩权重试。

## 3. 十一个精确角色与调用方责任（新增本地设计，非原规格已有角色）

调用方是未来可信policy绑定的主体，不是可从body或role字符串自报的身份。
同一主体可以承担多个明确授权的请求角色，但不能因某角色被信任而取得另一个
角色；每个role/key的schema允许集必须是下表精确集合的受限子集，不允许通配。

| Profile | 本地已批唯一signer_role | 原编号允许集 | 主体与接收方；授权来源 |
| --- | --- | --- | --- |
| U | claim_v1_upgrade_request | 12,20,52,66,68,69 | exact activation/namespace的upgrade coordinator→provider；既有credential policy中的coordinator key＋独立预置的请求role/schema ACL，不能用待发布policy自授权。 |
| E | claim_v1_epoch_request | 34,38,45 | 已授权epoch coordinator→provider；current credential policy的coordinator身份＋请求ACL，仍验证各自intent、lease、gap/commit proof。 |
| L | claim_v1_serving_lease_request | 78,80 | 已授权lease coordinator→provider；current credential policy的coordinator身份＋请求ACL，attestor记账和live-holder证明仍独立。revoke不要求已失效旧holder重新签名。 |
| B | claim_v1_rebind_request | 83 | 已授权rebind coordinator→provider；current coordinator身份＋recovery-scope请求ACL，仍必须有独立operator authorization及candidate证据。 |
| O | claim_v1_ordinary_request | 85,87,88 | exact current runtime holder经non-exportable incarnation signer→provider；current credential policy的exact incarnation key＋逐schema ordinary ACL和fresh context。 |
| R | claim_v1_read_request | 42,71,74 | scope受限的Inspect/proof主体→provider；独立预置的只读请求ACL，proof purpose再受单独允许集限制，不授予mutation能力。 |
| C | claim_v1_commit_challenge_request | 57,126 | 已授权commit-proof coordinator→provider；current coordinator身份＋精确row-kind/purpose/staging请求ACL。 |
| F | claim_v1_archive_fence_request | 61 | 已授权恢复编排方→exact commit-attestor/Fence signer；独立预置的archive mutation请求ACL，不从DB或普通accepted key set自动取得。 |
| I | claim_v1_archive_inspect_request | 63 | 已授权archive reader→exact commit-attestor；独立只读archive ACL，不具有Fence能力，且保留live candidate相互认证。 |
| S | claim_v1_staging_request | 120,123,129,131 | exact staging activation的coordinator→provider；current coordinator身份＋staging请求ACL，capability/intent/terminal proof仍必需。 |
| D | provider_downgrade_retirement_request | 139 | disposable downgrade coordinator→exact provider；独立预置的downgrade请求ACL，不能从普通key set或B11 DB role推导；另须原L993独立admin authorization。 |

U/E/L/B/C/S必须是现有accepted key set中exact coordinator，不额外向该集合加入
“新请求key”。O只用exact incarnation key，不能导出private key或让attestor对caller
任意body签名；attestor必须重验本schema许可、live holder与body中的runtime/context。
R/F/I/D使用独立、受管理的窄scope ACL，不因为属于普通accepted key set而获得权限。
R/I的只读ACL不因source accepted set为空而自动消失，也不能因此复活普通凭据：
必须有另外显式登记且未撤销的只读授权，retired namespace仍只允许原契约的读取。

请求role与provider response、security-admin、downgrade authorization、rebind
authorization等独立签发能力分离。provider response签名key不能用于请求认证；
普通请求授权不能提交C1 policy管理、签业务response、archive attestation或独立
admin authorization。同一协调者身份不意味着可以复用这些签发能力；不同trust
realm不能互相兜底。O使用既有incarnation key，不取消该key原有的独立数据库证明
签名能力；两类用途必须各自经过attestor的role/schema/可信读或live-holder检查，
数学上同一公钥能验签不构成用途授权，不为满足角色分离另造第三个accepted key。

## 4. 完整27项矩阵

每行共同取值：通用envelope=采用；SignerRoleOrNull=对应profile唯一role；
EvidenceKinds=[]。这是对本27个精确schema逐项提出的显式规则，不是对其他Request
的默认。下表来源列仅证明原body/状态约束，**不证明拟议role已在原规格中存在**。

| 原编号 / object | schema | Profile | 原来源与必须保留的认证后约束 |
| --- | --- | --- | --- |
| 12 AcquireDatabaseIncarnationRegistrationChallengeRequestV1 | acquire-database-incarnation-registration-challenge-request.v1 | U | L355/325/327；先有可信credential policy，匹配provider/endpoint/namespace/activation，challenge single-use。 |
| 20 RegisterDatabaseIncarnationRequestV1 | register-database-incarnation-request.v1 | U | L367/369；challenge与三份attestor envelope exact，production ID来自intent，success仅建立closed precursor。 |
| 34 RequestAuthorityEpochTransitionV1 | request-authority-epoch-transition.v1 | E | L432/434；committed intent、fresh proof/Head/current lease、zero pending与预分配ID，不能凭签名直接推进epoch。 |
| 38 ResolveAuthorityEpochTransitionRequestV1 | resolve-authority-epoch-transition-request.v1 | E | L438；trusted-read commit proof、exact pending CAS，response loss走Inspect，不重建pending。 |
| 42 InspectAuthorityEpochTransitionRequestV1 | inspect-authority-epoch-transition-request.v1 | R | L448；scope exact，caller_nonce绑定request/response，stage和stored envelope精确对应。 |
| 45 RecoverAuthorityEpochTransitionRequestV1 | recover-authority-epoch-transition-request.v1 | E | L454/610/683；intent proof、exact-one ordinary/gap、suffix与transcript ID闭合，只导出已决定outcome。 |
| 52 PrepareClaimV1GenesisRequestV1 | prepare-claim-v1-genesis-request.v1 | U | L496；expected Head=registered_pending_genesis，current DB attempt/runtime链与absence proof仍必需。 |
| 57 AcquireDatabaseCommitChallengeRequestV1 | acquire-database-commit-challenge-request.v1 | C | L604/608；row-kind、semantic ID、related set、purpose、generation及candidate exact，challenge不是提交证明。 |
| 61 FenceDatabaseCommitAttestationArchiveRequestV1 | fence-database-commit-attestation-archive-request.v1 | F | L637/639/641；独立reason授权、fresh Inspect、trusted semantic-ID journal和exact key set，不接受caller-only evidence。 |
| 63 InspectDatabaseCommitAttestationArchiveRequestV1 | inspect-database-commit-attestation-archive-request.v1 | I | L659；fresh provider Head/lineage与live candidate相互认证，caller字段只能做equality check，不能选择历史parent。 |
| 66 CompleteClaimV1GenesisRequestV1 | complete-claim-v1-genesis-request.v1 | U | L675；activation commit attestation/current proof/runtime/epoch exact，不继承completion响应角色。 |
| 68 PrepareClaimV1GenesisReleaseRequestV1 | prepare-claim-v1-genesis-release-request.v1 | U | L677；completion attestation、current proof及open ID预分配，响应的“同一角色”不授权请求。 |
| 69 OpenClaimV1GenesisRequestV1 | open-claim-v1-genesis-request.v1 | U | L679；open nonce/release attestation、mode matrix、current holder与epoch exact，不能提前允许ordinary。 |
| 71 InspectClaimV1GenesisRequestV1 | inspect-claim-v1-genesis-request.v1 | R | L681/683；按scope只读导出exact archived request/bundle/response，recovery使用transcript ID。 |
| 74 AcquireDatabaseIncarnationProofRequestV1 | acquire-database-incarnation-proof-request.v1 | R | L709/711；只允许该主体ACL中的purpose，fresh verifier nonce、single-use、current holder及30秒/lease窗口全部保持。 |
| 78 RenewClaimV1ServingLeaseRequestV1 | renew-claim-v1-serving-lease-request.v1 | L | L730/734/738/740；attestor记录exact request digest，三种renewal矩阵、ordinary prefix和history exact，不延长过期证明。 |
| 80 RevokeClaimV1ServingLeaseRequestV1 | revoke-claim-v1-serving-lease-request.v1 | L | L744；attestor先保存ID/digest/nonce，reason封闭、current lease/runtime exact，失效holder不可被要求复活来签请求。 |
| 83 RebindClaimV1RuntimeInstanceRequestV1 | rebind-claim-v1-runtime-instance-request.v1 | B | L746/752；独立operator authorization、termination、new candidate/lineage、scope与完整evidence set，签名不替代CAS。 |
| 85 ClaimV1ReserveRequestV1 | claim-v1-reserve-request.v1 | O | L713/715/722；exact current holder、fresh proof/context、effect/scope及provider phase，accepted key本身不足以Reserve。 |
| 87 ClaimV1FinalizeRequestV1 | claim-v1-finalize-request.v1 | O | L713/717/722；reservation、effect、DB point与允许的holder/context descendant匹配，不接受只有合法签名的伪commit。 |
| 88 ClaimV1AbortRequestV1 | claim-v1-abort-request.v1 | O | L713/718/722；reservation/abort reason封闭，原terminalization及Abort安全规则不变。 |
| 120 RecoverStagingImportCapabilityRequestV1 | recover-staging-import-capability-request.v1 | S | L851/853；committed intent、archive acquire/proof、same held exclusion，恢复只供revocation，不允许重import。 |
| 123 AcquireFreshV7StagingExclusionRequestV1 | acquire-fresh-v7-staging-exclusion-request.v1 | S | L841；capability/registration commit proof、fresh current Head与single-use challenge，零ordinary/no pending。 |
| 126 AcquireStagingDatabaseCommitChallengeRequestV1 | acquire-staging-database-commit-challenge-request.v1 | C | L871/873；fixed outcome kind、purpose、exact generation和held-state matrix，candidate preimage不能冒充已提交row。 |
| 129 ReleaseFreshV7StagingExclusionRequestV1 | release-fresh-v7-staging-exclusion-request.v1 | S | L877/879；held-only、无recovery、exact application attestation、current Head/lineage及terminal互斥。 |
| 131 AbortFreshV7StagingExclusionRequestV1 | abort-fresh-v7-staging-exclusion-request.v1 | S | L879；仅held或held_recovery_abort_only合法分支，exact revocation/application/proof；required不能直接Abort。 |
| 139 RetireProviderProtocolForDowngradeRequestV1 | retire-provider-protocol-for-downgrade-request.v1 | D | L993/995/1001；disposable-only、独立admin authorization、membership exact-cover与zero projections，不能靠请求role永久retire。 |

## 5. 接收方、subject、scope与bootstrap

以下为新增请求policy约束，不新增body字段。每个接收方必须在请求外预置受管理
的trusted profile：允许schema/role、subject、public key、root、policy version、
单一算法和exact receiver identity/endpoint、namespace或archive域、activation范围。
请求中的key/root只能选择已固定的可信项，不能自行建立信任。缺任一必要绑定则
不配置该调用，不允许“首次请求自动注册”“从body导入key”或从未验response反推。

同一已签名tuple(schema,role,key,root,policy version,algorithm)的信任配置必须
唯一绑定一个接收服务及其协议scope；不得在另一provider/namespace/attestor上
复制该tuple接受请求。body已有provider/endpoint/namespace/activation字段时还须
逐字相等；没有某字段的schema从该预置profile及受信现有activation/registration
关系机械绑定，不能从网络路由或caller hint补信任。无歧义绑定无法建立时阻塞
部署，不能让同一签名在多个recipient之间自行选择。

同一physical key可在其显式允许的多个请求role下使用，避免扩大L327既有accepted
key集合；但每个精确role/policy tuple仍必须有唯一recipient/scope，不能以角色复用
省略受限schema ACL。request与provider response/独立admin的key identity及签名
能力必须分离；既有incarnation key的O请求和DB attestation复用仅按§3的明确边界
保留，各用途的独立检查不得互相替代。本轮不生成key，不指定真实key ID、root或
policy version，也不为本矩阵重配incarnation key。

U首次registration前的policy必须是已由C1独立admin流程建立且可信的policy，
不能以registration request批准它自己。recovery/rebind/lease使用当前coordinator
身份与窄scope policy，不依赖已丢失旧holder的请求私钥；但不绕过原有candidate、
termination、intent或commit proof。O仅由当前incarnation signer在live holder满足
条件时签名，不能为了恢复让旧holder key重新可用。D/F/I的独立policy不能从普通
claim-v1 credential policy或DB role自动导出，D的request签名不替代L993授权。

R的read/proof权限逐schema/purpose批准：普通Inspect读权不自动包含proof请求权，
ordinary_mutation purpose的proof也不授予O角色；verifier仍须绑定本次proof ID、
caller nonce、expected Head与own scope。任何reader都不能用别的主体的proof
请求/response或缓存nonce满足本次检查。签名请求不取消现有请求限速、保密和
最小访问范围；retired state的读取不开放genesis、lease、Fence或ordinary mutation。

## 6. 证据用途、archive与重试

拟明确本27项EvidenceKinds=[]：即便请求签名有效，也不能作为独立
external_signed_envelope或database_immutable_body bundle项，不因被DB/archive
保存而升级。这是新增明确用途限制，不是原规格已逐项禁止的声明。
已有response资格不变，request body digest仍是原body digest，不能改成新envelope摘要。

原L278的exact body/bundle/response保存继续；本提案额外要求按原操作ID/body
归档第一次获接受请求的exact envelope及认证subject/profile引用，后续合法retry
另记认证事实，不改写初始证据。归档新增内容不写回业务body、response、DB既有
digest字段或CanonicalEvidenceBundleV1结构。具体archive物理布局属后续consumer
实施，不在本轮创建或迁移。

L639的特殊provider request展开仍须同组signed response的request_digest及
immutable Inspect共同认证，不能因现在多了一份请求签名就接受caller-only bytes。
只有契约明确的archive/展开位置可携带原文，不增添第三种evidence_kind。

每次在线调用（含exact retry）必须通过当前可信请求认证和scope检查。业务幂等
仍以原semantic ID＋body及原状态规则为准，不按envelope signature字节重新建立
operation；不改nonce/expiry/terminal CAS规则。换签名key不改变body digest，
但必须是同一可信subject或已由独立管理流程显式授予该原operation恢复scope的
主体，不能仅凭新key也有同role就接管别人operation。认证后的body变化仍conflict。
历史archive回验按原recorded transition规则，不要求把历史body重签成当前事实。
过期proof不授权新effect；合法exact retry可返回已保存结果但不能重跑mutation。

## 7. 兼容性、状态与停止规则

本提案是27项请求wire/role及archive认证记账的补充，不是无行为变更的文档更名。
11个role、27个[]限制、recipient/scope绑定、主体恢复规则和额外请求envelope保存
已作为新增内容获本地书面确认。现有provider/attestor是否使用其他请求认证方式未核实，
兼容性UNKNOWN。不能上线新旧格式双路径、legacy unsigned fallback或响应role别名。
即使本地设计获批，接线前仍须冻结各接收方实际profile/版本、compatibility vectors、
caller/consumer清单及拒绝旧格式边界；无法唯一绑定recipient则继续阻塞。

本矩阵现已获书面确认，新的本地metadata对照记为130/16；剩余P组16项仍须各自
明确数据通道，不能因为请求规则清楚就默认通过。这个局部分类计数不是27项
已实现或独立认证；canonical登记、实际profile/兼容性及runtime验收仍开放。

不改canonical approved spec-set及摘要，不改变C1和已批九项角色，不创建产品
registry/fixture/codec/provider，不运行Go/build/test/fuzz/Docker/DB，不合并main。
全146项未闭合前，不生成完整metadata交付的实施就绪声明或执行计划。

## 8. 后续验收要求（本轮未运行）

- 独立字面量覆盖27项精确schema/object/profile→role和显式[]；R组missing+extra
  抵消计数也失败。11角色只允许表中schema子集，read→Fence、read→ordinary、
  upgrade→policy-admin、D→DB downgrade authorizer等跨用全部拒绝。
- request/response role和key互换、错误recipient/namespace/activation、复制可信
  tuple到另一receiver、caller key/root、旧policy/撤销key及算法替换在mutation前拒绝。
- 保留每行原proof/lease/Head/intent/commit/held/authorization拒绝；合法签名不能
  取消local candidate相互认证，O签名不能在死holder上获得，L revoke不要求复活旧holder。
  O与DB证明即便使用同一既有incarnation key，也必须拒绝跨schema/role/用途替换；
  请求签名不得成为让attestor略过独立trusted-read后签DB证明的入口。
- 首次policy bootstrap不能自授权；source retired的显式只读ACL仅可读取，不复活
  普通credential。合法请求签名尝试两种独立evidence包装均拒绝，archive展开仍需
  provider response＋Inspect绑定，request_digest仍只绑定原body。
- 相同ID/body不同合法envelope保持同一effect；wrong subject、无恢复scope、不同body、
  nonce跨请求与过期proof的新effect均拒绝；archive不能被retry覆盖或重写历史。
- 目录层变异测试只证明元数据；真实consumer的认证、副作用0和重试验证须在后续
  相应计划另行冻结。不得把名称目录测试当作上述运行时认证已实现。

## 9. 自检、审阅与恢复

本轮只检查27项身份一一对应、11角色/profile引用与精确允许集、来源链接、[]用途
声明、原43项计数和变更边界；没有实际认证、签名或业务负载执行，没有独立复审。
设计自检须确认新增规则标记清晰、没有null兜底、没有请求/响应能力混用、没有把
candidate互认证当作编排方身份、没有新增body字段或改变摘要/accepted key集合。

整份27项矩阵及新增信任/用途限制现已获本地书面确认；继续剩余16项数据通道，
不是直接写产品代码。最新来源核对见[16项数据通道审查](../../roadmap/2026-10-17-task11-data-path-review.md)。
必要设计/恢复文件随现有开发分支保存；新电脑从
[恢复记录](../../runbooks/repository-recovery.md)继续，不重跑已消耗的旧验证阶段。
