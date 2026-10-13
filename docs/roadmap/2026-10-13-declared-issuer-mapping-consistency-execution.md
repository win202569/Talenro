# Declared issuer mapping consistency — 本机执行记录

日期：2026-10-13。状态：首次五任务交付记录；后续M1处理状态见§7。初次审查无阻断项，M1初次待办。
工作区：`.worktrees/c12-b01-task9-coordinator`。
分支：`codex/c12-b01-task9-coordinator`。
HEAD基线及交付值：`d2829bfaa93ae2a119162a5ce6057fbe223d2cb7`。
本次未暂存、提交、推送、PR、合并、清理或安装；不是远端恢复点。

## 1. 实际范围与结果

按[获批计划](../superpowers/plans/2026-10-13-declared-issuer-mapping-consistency.md) Native顺序执行；
[获批规格](../superpowers/specs/2026-10-13-production-issuer-serial-mapping-compatibility-design.md)
SHA256保持4AEB4D41F72066CDE1B129D3A42ECC81D33BFD4CD6D413E7849F67DD0F45857C。

新增纯组件`internal/nodecontrol/nodeissuermapping`，恰25个Go文件：12产品、13测试。
唯一旧Go差异为`internal/nodecontrol/nodeissuerbinding/consumer_boundary_test.go`：
36新增/2删除，只准许exact新目录/package消费exact真实lower；旧限制保持。
没有旧产品、旧fixture、SQL、module或runtime入口变更。

唯一新公开入口`CheckDeclared(context.Context,Input)(*Model,error)`。
它检查调用方声明的三个head、完整登记表、显式IssuerID/CA版本、CA DER、
NodeID、leaf DER、SPKI及canonical serial与一次真实`nodeissuerbinding.Evaluate`结果一致。
Issuer逐byte匹配；全表重复同键即使相同/非选中也拒绝；撤回不fallback/latest。
成功profile为`declared-issuer-serial.registry-version-link.v1`，Proof仅为
`DeclaredIssuerSerialCorrespondenceConditionalTimeBound`。没有生产成功位、来源工厂、
真实发布/登记wire/hash、可信读取、DB访问、迁移、回填、lease或provider能力。

声明预算/形状先于lower；材料原错误先于声明对应错误；可观察ctx错误优先。
上层owner只存自有Claims、完整Summary、valid。Material全部预算/复制由真实lower负责，
不能把上层入口描述为已复制全部Material raw。Claims两集合及返回Evidence两slice脱离别名；
失败和普通panic清理可达自有数据，panic继续传播。浅Model复制共享Clear，独立Model互不影响。
并发、隐藏runtime副本、字符串或物理内存擦除、硬实时取消/OOM/crash不在保证中。

所有继承Evidence/责任/原观察时间保留，新五责任全部保留：

- AuthenticatedRegistryPublicationAndRollbackProtection
- AuthenticatedIssuanceLinkAndRowStatus
- GuardedV7SnapshotAndLease
- CrossRenewalDatabaseUniqueness
- FullIdentityCompositionCommitAndRuntimeAuthorization

FabricatedClaimsAreNotTrusted、CoherentRollbackClaim、CoherentChangedCatalogClaim明确：
完整伪造、自洽旧head或同自报digest下改变非选中声明可以通过，只得声明一致；
这不证明部署/发布来源、当前头、防回退、登记内容摘要、真实行状态或原签发版本。

## 2. 本轮自有证据与TDD

证据仅在本轮忽略scope
`.superpowers/sdd/2026-10-13-declared-issuer-mapping-consistency/`。
它包含plan-path、原完整task briefs/BASE、progress、676-file baseline、原始stdout/stderr、
程序/完整参数/固定离线环境/UTC/exit、701-file矩阵冻结清单、生成器档案、review记录。
没有读取/复用/改写旧增量的私有证据或缓存。

每条原始失败保留；005和012是测试移植/字段路径错误，**不是目标RED**。
后续006/013修正测试后才分别确认缺失生命周期/公开入口的目标RED。
交付021的源码/格式/回执审计成功；末尾rg因无未完成checkbox返回1，导致外层
检查命令exit1。该预期no-match不是Go失败或目标RED；023显式核对全部30步骤。

- Task1：001缺API/预算RED；002精确消费例外RED；003/004各GREEN。
- Task2：005移植错误保留；006缺生命周期RED；007反射完整字段、
  alias/共享Clear/自有backing清理及动态取消/截止/panic GREEN。
- Task3：008一次性生成成功；009独立39份材料controls GREEN；
  010缺声明/lower/投影helper RED；011目标选择器GREEN。
  21-byte DER负例确名由旧跟踪源码核对后补强，015/最终矩阵覆盖。
- Task4：012测试路径错误保留；013缺CheckDeclared RED；014选择器GREEN，
  动态枚举730个真实组合ctx检查点；015新包完整GREEN（20.020s）。
- Task5：016缺guard函数RED；017空schema闭合集合RED，product/member/consumer已通过；
  逐源审计12产品后固定ASTschema；018三守卫/真实负mutation及格式、
  import alias/local rename正控制GREEN；019格式26路径空输出/exit0；
  matrix1全部26项exit0，之后唯一fresh只读审查。

声明Golden未调用新CheckDeclared、validateDeclarations或bindProjection。
39 Material packets及原manifest/raw完全继承既有固定控制；新增synthetic声明，
用真实lower和独立X.509/DER/SPKI/serial核对。普通测试不签发、不读旧源。
运行过的生成器byte-identical归档`freeze.go.txt`：
91B6C84EE3EAE8AB7F701720B7CC6F43CCB42F7354CF28BDE95F4657425938D2。
最终前没有活动生成器.go躲在ignored scope中。

## 3. 最终26独立命令与冻结

固定Go1.26.5 Windows/amd64。每项count1/p1；普通项timeout120s。
GOTOOLCHAIN=local、GOPROXY/GOSUMDB/GOAUTH=off、GOVCS=all:off、
GOFLAGS=-mod=readonly；缓存仅为本轮scope/go-cache，模块来自已有只读缓存。

matrix1-01..21分别完整运行以下nodecontrol包：
nodeissuermapping、nodeissuerbinding、nodeleafbinding、nodecerttime、nodeissuer、
nodeca、nodeleaf、nodeleafkeys、contracts、serving、statebinding、statecomposition、
statecontract、stateformat、statematerials、statepolicy、statepolicybinding、
staterecovery、statetime、statetimebinding、stateverify。
matrix1-22完整internal/readiness。
matrix1-23 authority仅`^(TestBoundRead|TestCoordinator)`。
matrix1-24 store仅`^$`编译；明确没有数据库测试执行。
matrix1-25/26 nodeissuer分别FuzzIssuerEnvelope/FuzzIssuerPair：
run`^$`、fuzztime5s、parallel1/p1、timeout30s。

26/26 exit0；22完整包+authority选择+store编译+2短fuzz。
累计命令wall时间411.56s（含构建），不是测试CPU时间。
最终fuzz执行次数24/6915、测试包实际6.999s/7.843s；只有有限抽样，
不能当覆盖充分、无缺陷或生产安全证明。
首次fuzz含冷构建，命令wall80.73s；未放宽测试时限或跳过失败。

676旧源码盘点仅一个授权旧测试改变；新增25后恰701 Go/SQL/module文件。
matrix1-before.json与matrix1-after.json逐路径/内容一致，文件本身SHA256均为：
77EBAA87C995BD21E114E1A24DBFCEEA2FF7F9AC7D99F09D658EC849E92B2D8D。
26回执摘要matrix1-summary.json：
FEBED9FC0828AF2C3787F7448852BF720655C25D80855AF16369738D5BD27A8F。
审查者也核对了交付workspace与这701项一致。
019格式为25新Go+1旧测试，不与26命令计数混淆；末尾文档编辑后再检查格式与同一冻结源码。
020交付格式检查26路径空输出/exit0；021核对701源hash、唯一旧差异及26成功回执。
022只读Git审计确认必要文件30、HEAD/index/spec不变、暂存为空、diff检查通过。

## 4. 唯一fresh审查与待办

一位gpt-6-astra/xhigh、隔离上下文的只读reviewer：
`declared_mapping_final_review`。无实现子代理、第二审查者、新chat或审查写入。
完整阅读25新Go、唯一旧diff、spec/plan/rulings、生成器及回执；
独立交叉核对39材料/原manifest/生成常量、701源hash、26命令、格式、
730点回执、HEAD和空staged diff。五Review Focus均实质覆盖。

结论：bounded local checkpoint可交付；Critical0/Important0/Minor1。
M1：lower_test.go/check_test.go目前反射比较错误类型/字段/unwrap链，但DeepEqual可接受
不同identity的同文errors.New；把所有pointer错误计为typed不足以明确要求结构化控制。
建议追加terminal sentinel直接identity检查，显式要求已存在policy-failure/unwrap控制。

作者核对：当前lower/check直接保留原错误，固定policy-failure确有结构化字段/unwrap。
按使用者实际影响，M1维持非阻断Minor，记录待办而非伪称已修复。
本轮无产品修复，因此没有虚构回归RED–GREEN、第二矩阵或修复复审。
将来修复须按获批计划做对应回归、完整26重矩阵和同reviewer复查。

## 5. 作者裁定（理由与代价）

按作出顺序记录实施裁定；不是新增生产权限。

1. PowerShell替代Bash工作区/task脚本：遵守Windows限制；代价为人工identity/hash审计风险。
2. 获批26限定矩阵替代通用full-suite/install/commit/cleanup默认：遵守明确边界；
   代价为完整/物理/远端验收仍缺，保留本机scope而非清理。
3. Native顺序实现+末尾唯一fresh审查：用户选定方式；代价为中途没有独立任务审查。
4. 若修复则完整重矩阵+同reviewer复查：遵守获批计划而非技能默认不复审；
   代价为出现修复时增加验证/审查成本，本轮未触发。
5. bindProjection在slice前重复私有validateEvidence：避免私有无效Summary导致越界；
   公开路径仍仅一次真实lower，代价为少量有界重复检查。
6. 审查当前工作区库存/hash+授权旧diff而非空committed range：没有Git写入授权；
   代价为结果仅本机、未远端保存。

审查者每个“Declined to judge”均逐项裁定，没有暗中接受：

1. **Authentic registry publication and deployment authority**：No authenticated source exists; success remains caller-declared consistency。若误判：Treating it as authority could accept forged deployment/publication facts。

2. **Persistent current-head and rollback protection**：Coherent old coordinates intentionally remain acceptable declarations。若误判：Stale eligible records could be mistaken for current authority。

3. **Registry content-digest verification**：No approved registry wire encoding/hash is implemented。若误判：Changed catalog contents can retain a caller-reported digest。

4. **Withdrawal safety across historical snapshots**：Authentic publication/current-state evidence is absent。若误判：A withdrawn registration could be represented by an old eligible declaration。

5. **Real issuance provenance, row status and historical original CA selection**：Certificate/issuance UUIDs and version association are declarations, not authenticated rows。若误判：Invented or historically incorrect associations remain possible。

6. **Actual database uniqueness, transactions, legacy backfill and migration**：No SQL/database changes or physical tests are authorized; existing namespace is preserved。若误判：Real writes may conflict or legacy evidence may be insufficient。

7. **Real v7 protected reads, connection identity, leases, incarnation and pre/post stability**：No production reader or private protected-read capability is added。若误判：Combining this component as a runtime gate would omit required source/lease checks。

8. **Trusted observation time and caching authority**：Inherited times remain conditional; each new observation requires the proper lower/production checks。若误判：Reusing results beyond real freshness boundaries could accept expired evidence。

9. **Private-key possession, full identity/history, commit, endpoint/provider and runtime authority**：All inherited/new obligations remain; no runtime consumer or factory is added。若误判：Material consistency could otherwise be mistaken for authorization。

10. **Concurrent caller/model mutation**：Documented unsupported; caller must synchronize access。若误判：Races could corrupt input/snapshots or invalidate clearing assumptions。

11. **Hard-real-time cancellation, OOM/crash cleanup, hidden copies, strings and physical erasure**：Only observed cooperative boundaries and reachable owned mutable storage are covered。若误判：Cancellation latency or surviving copies can exceed these limited guarantees。

12. **Exhaustive security proof, unsampled inputs, performance/cross-platform/full-suite/production acceptance**：Only scoped Windows receipts and finite schema/mutation/fuzz controls were run。若误判：Unexamined platforms, inputs and physical paths may still fail。

13. **Final execution/status-document accuracy**：Author checks handoff documents against actual receipts and reviewed state; no independent document review is claimed。若误判：An author transcription error remains possible。

## 6. 交付及仍开放门槛

路线图仅追加本轮实际状态；获批规格/旧产品/旧fixture/SQL/module、HEAD/index不改。
索引SHA256保持1E845329E5310FED0BD00F8A18167678B9BCAE15EF7570F9B09654B8EA11A769。
所有失败/原回执留在本轮本机scope，未列为Git保存文件。

本轮必要文件共30：25新Go、1旧测试、获批spec/plan、此执行记录、current-status追加。
它们尚未暂存/提交/推送；不能从GitHub获取本轮进度。Git保存需要独立授权。
此前d2829bfa远端恢复点只包含上一增量，不自动覆盖本轮未保存内容。

没有真实发布、签发来源、DB唯一性/事务、v7读取/租约/前后稳定性、可信观察、
生产兼容或runtime权限验收。历史C12/B01/Task10/Task15/main、I2/I3、
完整ordinary/C11/C12、物理测试和其他既有阻塞保持开放。
安装Docker与本轮纯组件/限定矩阵成功不关闭这些门槛。

## 7. M1测试加固后续（2026-10-13）

用户“继续下一步”后处理初次交付留下的M1；此节更新当前状态，上文初次
deferred/未重矩阵描述保持为历史事实。只改本组件lower_test.go/check_test.go；
12产品文件、其schema、25-Go库存、全部fixture、唯一旧test差异不变。

先把原反射错误链比较及普通pointer分类提取到test-only helper，024确认
同文字/不同identity错误被错误接受；025分别确认direct、wrapped及普通
sentinel被误计为structured的真实RED。失败原回执未删除/覆盖。
随后比较每一层具体类型和全部字段，terminal sentinel额外按identity比较；
合法独立分配的结构化wrapper仍允许，且保持相同unwrap终端。
lower和public双故障测试必须各执行exact policy-failure控制一次：
具体PolicyFailure类型/来源包、Code/Reason字段及non-nil unwrap；
不再把任意pointer当typed。正控制使用真实lower产生的结构化错误，
字段改变、普通pointer及direct/wrapped不同终端身份负控制均实际拒绝。
没有引入更低层产品import、测试hook或产品行为变更。

026针对性GREEN；027格式26路径空/exit0；matrix2完整26/26 exit0。
累计命令wall229.37s；fuzz Envelope/Pair实际72498/9次、
测试包8.339s/8.503s，仍只有限证据。
701文件Path/SHA256逐项前后一致。matrix2-before原JSON因scope盘点工具
Compare-Object -PassThru为两条测试记录添加了SideIndicator元数据，
所以JSON文件字节摘要不同，但文件Path/SHA256未变；保留原始记录，不覆盖：

- 原before JSON：2CB3A57EC3B028489B2B201F27E568E162198411AA26D24731A17F5B67044F1E
- after JSON：9DB4FC99B644493C45EB7C0F30D59840DBA97F445EDA0C08BEAAE2A090D41B58
- 仅选Path/SHA256的before规范视图与after的文件摘要相同。
- matrix2-summary：2D51FA92DE4E556FED073985F03D84A69CA9DC523B41F158CAFB0BCDA03CDFDA

原fresh审查者已复查：M1 ADDRESSED；新增Critical0/Important0/Minor0，
bounded local handoff可交付。审查没有改文件或重跑测试；核对了024–027、
完整matrix2/701源hash和当前workspace，仅两个授权测试改变；产品/schema/
fixture/旧consumer/spec/HEAD/index不变。也检查了所提交§7和status附录，
其历史deferred与后续处理、元数据差异和有限fuzz归因准确；此处最终闭合文字
为作者根据实际复查结果追加，不宣称审查者已检查未来文字。

新增审查边界D14及作者裁定：test-only matcher只验证既有有限下层错误合同，
不扩展为任意错误图、循环/multi-error unwrap或其他错误实现的通用验证器。
理由是无产品暴露、无任意error输入验证能力或替换lower hook；
代价是未来下层错误合同变化时，此辅助测试可能无法充分比较或因循环超时，
必须重新审查和补强，不能当通用错误安全证明。原D1–D13风险仍保留；
D13本次只对已提供文本得到独立检查，最终闭合追加仍为作者核对。

必要文件仍30个、仅本机。全部§5未验收边界与风险继续保留，
不因错误测试加固升级生产信任或关闭历史门槛。
## 8. 必要文件保存与长期授权

用户进一步要求“以后都要提交必要文件，以便在全新电脑上恢复进度”。
按上文询问的commit/push交接语境，这授权本次和后续正常开发分支保存；
不包含main合并、force push、删除分支、安装、安全设置或更宽验收。
root AGENTS.md记录长期规则，repository-recovery.md补充分支克隆和恢复次序。
必要文件32个：原30个加AGENTS.md及runbook更新；源码和fixture与M1复查版本一致。
密钥、凭据、缓存、工具二进制、Docker数据、conversation/session和ignored
原始回执不进入Git；当前状态、spec/plan、执行记录及未验收项必须可从仓库恢复。

上述“仅本机”描述为此前交付历史；本次实际提交/push及远端SHA在保存动作
成功后确认，不在动作前声称成功。新电脑的真实克隆、依赖准备及物理验收
仍未执行，Git可获取不等于生产验收。M1已解决，D1–D14范围保持。
### 保存前matrix3失败记录及检查点分类

fresh提交前matrix3前24项exit0，第25项FuzzIssuerEnvelope exit1：
context deadline exceeded；12次执行，报告6.02s，包时间8.844s。
命令严格保持run'^$'/fuzz'^FuzzIssuerEnvelope$'/fuzztime5s/parallel1/p1/timeout30s。
第26项FuzzIssuerPair未运行，矩阵停止；不是26/26通过。
已有matrix1/matrix2的成功及M1复查是此前实际证据，不删除，但不覆盖这次失败。

全部701源Path/SHA256未变；matrix3-before/stopped-source JSON字节摘要同为
9DB4FC99B644493C45EB7C0F30D59840DBA97F445EDA0C08BEAAE2A090D41B58。
完整参数/环境/UTC/exit/stdout/stderr保留于matrix3-25-extra原回执，
汇总matrix3-stopped-summary明确24成功/1失败/1未运行。
回调源码只解析/比较材料，没有自行设置deadline；未见失败输入或testdata生成。
Go框架协调/worker存在时间预算，但当前证据不足以确定具体根因。
未重试、加时、删cache、放宽守卫或改任何旧产品/测试；不能称此失败已修复。

用户长期授权包含保存待完成进度，故本次按“未验收开发检查点”提交推送32个
必要文件，保留最新验证阻塞，不合并main。换机后从current-status此条恢复，
先继续调查失败，再依据获批门槛验证；单纯Git可获取不意味着验收完成。
