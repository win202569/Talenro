# Node leaf direct issuer relation — execution record

状态：五任务本地partial增量完成；源码、限定矩阵、独立审查及056正式文档保全检查均通过。
本文件记录本次局部增量，不是 Task10/B01/C12/main、合并或生产验收。

## 1. 审批、范围和真实基线

用户于2026-10-12批准[书面设计](../superpowers/specs/2026-10-12-node-leaf-issuer-relation-design.md)，
随后选择选项1 Native inline 执行[五任务计划](../superpowers/plans/2026-10-12-node-leaf-issuer-relation.md)。
审批消息没有提供秒级UTC，不以基线时间冒充审批时间。
五任务由作者在同会话顺序实现；仅在冻结矩阵完成后派一次全新上下文独立审查。
本次未获暂存、提交、推送、main合并、PR、强推、删除分支/worktree或provider迁移授权。

实际worktree：D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator。
BASE/未变HEAD：764dcb8b6b682ecce03fd6de02dbe339c66ff985。
分支：codex/c12-b01-task9-coordinator。
基线UTC：2026-10-12T11:53:53.9219067Z；766个非忽略路径；
当时仅本次获批计划未跟踪，其余非忽略文件与暂存区干净。
最新已推送检查点仍是2877f89cc9c5a6ca0514d79c54826627a112d412，
不能把本机设计提交或本次实现当作远端可恢复进度。

初始基线探测误含git write-tree --missing-ok，可能写已有tree对象；
不把该命令称为严格只读。它未改变逻辑HEAD/index，tree值为
cab171f25fee46cc149230507f28bd921c2102f0；随后不再使用write-tree。
首次确认及后续复核的index SHA256均为
D48CD341480E047AC816B9F189445A6F8EDBBF02018C3DCD97EF1165C3DF67BC。
没有暂存、提交、推送、配置写入或worktree清理。

## 2. 实际实现与保留责任

新增internal/nodecontrol/nodeissuer下exact24 Go文件：12产品、12测试。
仅改旧nodeleaf、nodeleafkeys、nodeca三个boundary_test.go的精确消费例外
及其新增正反控制；原API、守卫其余部分、测试案例和所有旧testdata保持原样。
除此之外只更新本计划checkbox、本执行文档和current-status顶部本增量。

固定输入为Leaf nodeleaf.Input + CA nodeca.Input，独立拥有四个byte buffer。
上限依次65,536 /512 /65,536 /8,192 bytes，合计139,776为输入预算，
不是总分配或峰值heap上限。Material只持有owner，Summary只含固定值；
无DER accessor、parser对象、key对象或更强Proof外泄。

真实调用顺序为CA Parse → nodeleafkeys Evaluate；
保留原lower error身份，返回前清理两层临时material。
同一owned字节再解析并核对所有摘要、Name/SPKI/ID、serial、UTC时间、
pathLen省略/显式零和CRLSign。之后最小leaf DER envelope和签名表示检查，
精确key → raw Name → full32 identifiers关系，
实际CA P-256公钥上的SHA256(full original TBS DER)/ECDSA VerifyASN1，
最终context检查后才valid/transfer。允许规范high-S，不重编码或重签TBS，
不添加x509信任链验证、私钥、随机数、I/O、nodeissuer产品消费者或provider。

成功Proof仅为NodeLeafDirectIssuerSignatureBound。
Leaf的NodeLeafDeclaredKeyIdentifiersBound及嵌套NodeLeafMaterialShapeValid、
CA的NodeCAMaterialShapeAndDeclaredKeyBound原级别保留，
下列两张原Pending表均原样保留，不因本组合层的局部关系而删除条目：

| lower | 原Pending（按原顺序） |
| --- | --- |
| leaf | LeafSignatureAndIssuerTrust；KeyIdentifierDerivationAndIssuerBinding；TrustedCurrentTime；IdentityAndAuthorityProvenance；PrivateKeyPossession；MainStateAndRuntimeBinding |
| CA | CertificateSignatureAndIssuerChain；AuthorityIdentifierAndParentBinding；TrustedCurrentTime；DeploymentAuthorizationAndPurpose；BundleStatusAndRollback；LeafIssuerRelationAndRuntimeUse |

同key/name的另一张CA可具有相同直接数学关系，但CA DER digest不同；
不证明唯一签发证书、可信issuer_id或来源。CA自身错签名、任意形状有效parent AKI、
expired/future/cross-CA-window和same-key角色控制不升级为信任事实。

## 3. 环境与证据方法

Windows PowerShell only；仅使用已缓存Go1.26.5 windows/amd64及同目录gofmt：
C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/。
GOTOOLCHAIN=local、GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、
GOFLAGS=-mod=readonly；GOPATH=C:/Users/Lenovo/go；GOMODCACHE同目录pkg/mod；
GOCACHE只在本计划专属忽略scratch内。继承GODEBUG为空；
包装器遇非空值即拒绝，不输出其他可能带凭据的环境。

证据位置：.superpowers/sdd/2026-10-12-node-leaf-issuer-relation/。
基线、五份task brief、progress ledger、verify.ps1、fixture generator/manifest、
原始编号回执、源码freeze和review-1-verbatim.md均保留。
文本创建/编辑/回执保存只用apply_patch。
50次Go/gofmt子进程执行保留真实executable/arguments/workdir、UTC开始/结束、
完整stdout/stderr、真实exit、白名单环境、命令前后源码inventory和原包装器；
失败回执不覆盖。008、030、031是另行诊断/快照，不是完整子进程回执；
它们缺少进程开始/结束/child exit字段的事实保留，不补造。
后续054/055及最终保全记录也明确是只读快照，不计为测试命令。

13组独立合成fixture在007一次性离线生成后冻结，普通测试不签发或随机重建；
008 Windows独立SHA256核对234项，0 mismatch。
009独立标准库和真实lower有效控制通过。
RFC6979 A.2.5固定公开P-256原语向量及独立N-s值用于签名原语测试，
不把sample消息当完整certificate成功oracle，也不由被测函数计算预期。
审查者另用.NET独立验证13个leaf签名和234个摘要。

## 4. 全部早期attempt真实处置

全部时间为UTC；下表包括预期RED、非预期守卫失败、诊断和格式检查，
不能把文件名含green当成功。全套最终限定结果另列下一节。

| 编号 | child exit | 实际UTC | 状态/处置 |
| --- | --- | --- | --- |
| 001 | 0 | 2026-10-12T11:55:35.1112934Z → 2026-10-12T11:57:44.6182712Z | 旧三包基线，exit0 |
| 002 | 1 | 2026-10-12T11:58:24.5746554Z → 2026-10-12T11:58:43.3724178Z | 消费者新允许案例真实 RED |
| 003 | 0 | 2026-10-12T11:58:59.5869361Z → 2026-10-12T11:59:10.0785503Z | 三个精确例外 GREEN |
| 004 | 1 | 2026-10-12T12:00:21.7526385Z → 2026-10-12T12:00:22.8459734Z | 合同缺声明，预期编译 RED |
| 005 | 0 | 2026-10-12T12:01:31.9937547Z → 2026-10-12T12:01:35.4327805Z | 合同 GREEN |
| 006 | 0 | 2026-10-12T12:01:54.957013Z → 2026-10-12T12:02:06.7171292Z | Task1 新包及旧三包 GREEN |
| 007 | 0 | 2026-10-12T12:04:07.8782049Z → 2026-10-12T12:04:09.9463808Z | 一次性标准库合成材料生成，exit0 |
| 008 | n/a | 2026-10-12T12:05:16.5514672Z（单次诊断UTC） | Windows 234 个摘要诊断；13 variants，0 mismatch；不是进程回执 |
| 009 | 0 | 2026-10-12T12:05:18.2395766Z → 2026-10-12T12:05:21.9331897Z | 独立 fixture/stdlib/lower 控制 GREEN |
| 010 | 1 | 2026-10-12T12:06:30.3644302Z → 2026-10-12T12:06:32.5370354Z | 真实 lower/snapshot 缺声明 RED |
| 011 | 0 | 2026-10-12T12:07:26.1912441Z → 2026-10-12T12:07:30.5859347Z | lower/snapshot GREEN |
| 012 | 0 | 2026-10-12T12:08:05.0693539Z → 2026-10-12T12:08:28.311189Z | Task2 四包 GREEN |
| 013 | 1 | 2026-10-12T12:10:28.1754486Z → 2026-10-12T12:10:29.1399726Z | scanner/signature 缺声明 RED |
| 014 | 0 | 2026-10-12T12:11:27.7184194Z → 2026-10-12T12:11:35.8515868Z | scanner/signature GREEN |
| 015 | 0 | 2026-10-12T12:12:10.842987Z → 2026-10-12T12:12:26.5380039Z | Task3 四包 GREEN |
| 016 | 1 | 2026-10-12T12:14:25.4063831Z → 2026-10-12T12:14:26.175645Z | relation/Evaluate 缺声明 RED |
| 017 | 0 | 2026-10-12T12:15:08.0765957Z → 2026-10-12T12:15:11.8363982Z | relation/Evaluate GREEN |
| 018 | 0 | 2026-10-12T12:15:43.1935054Z → 2026-10-12T12:15:56.3453049Z | Task4 四包 GREEN |
| 019 | 0 | 2026-10-12T12:16:32.2101454Z → 2026-10-12T12:16:35.5071142Z | 取消观察 GREEN：10 observed /22 completed first |
| 020 | 1 | 2026-10-12T12:17:37.140769Z → 2026-10-12T12:17:37.9712766Z | guard 缺 checker RED |
| 021 | 1 | 2026-10-12T12:22:08.399301Z → 2026-10-12T12:22:12.3817803Z | 失败：正常 lower.go 被误拒绝；文件名 green 不是通过结论 |
| 022 | 0 | 2026-10-12T12:23:14.5795972Z → 2026-10-12T12:23:18.3700646Z | 临时只读 AST 分解诊断 exit0；临时测试随后移除 |
| 023 | 1 | 2026-10-12T12:23:48.6174249Z → 2026-10-12T12:23:52.3553697Z | nodeca.Parse 默认/命名 alias 真实断言 RED |
| 024 | 1 | 2026-10-12T12:24:50.4474241Z → 2026-10-12T12:24:54.2493046Z | 失败：Parse 修正后 capability 把 ImportSpec 当异常 |
| 025 | 0 | 2026-10-12T12:25:53.6604971Z → 2026-10-12T12:25:57.7231104Z | guard GREEN：只跳过已验证 import 声明 |
| 026 | 0 | 2026-10-12T12:26:52.332694Z → 2026-10-12T12:26:56.4728282Z | 完整新包及两个普通 fuzz seed corpus GREEN |
| 027 | 0 | 2026-10-12T12:27:43.160869Z → 2026-10-12T12:27:43.2325024Z | 仅 exact27 Go 文件格式化，exit0 |
| 028 | 0 | 2026-10-12T12:27:44.9865161Z → 2026-10-12T12:27:45.0556453Z | exact27 格式检查，exit0、空输出 |
| 029 | 0 | 2026-10-12T12:27:46.8153971Z → 2026-10-12T12:27:51.5481361Z | 格式后完整新包 GREEN，所有内存拒绝变体 active |
| 030 | n/a | 未捕获进程开始/结束 UTC | 旧三 guard 净差异/Git 只读输出；未留完整进程时间回执 |
| 031 | n/a | 2026-10-12T12:29:06.1647257Z（快照UTC） | 保全快照，766 基线中只计划/三 guard 变化；不是进程回执 |

021根因是test-only import member名单漏了计划明确允许的nodeca.Parse，
非产品行为缺陷。023先用默认/命名alias真实复现，再只补该成员；
024仍失败，根因capability遍历把已通过Product/Canonical校验的import
GenDecl中ImportSpec当默认异常；只跳过已验证import声明，025 GREEN。
临时诊断测试已移除，原输出保留，没有放宽禁用能力或改产品源码。

四个test-only checker分别约束产品imports/members、exact能力/API/文件闭包、
消费者目录/package和拥有/清理/验签阶段流。所有内存拒绝变体必须active，
无故意破坏真实源码、skip或panic recover。有限AST检查不是任意Go程序的形式证明。

## 5. 冻结后的22条最终限定矩阵

freeze1 UTC：2026-10-12T12:27:45.0556453Z；
027格式化exit0；028格式检查exit0且无输出；029格式后新包通过。
032–053全部实际exit0，每条before/after27项长度/hash均匹配同一freeze1。
18个完整包、1条指定authority、1条store仅编译、2条短fuzz；
不是整仓库。普通包统一-count=1 -p=1 -timeout=120s；
authority只-run '^(TestBoundRead|TestCoordinator)'，不启integration tag；
store只-run '^$'，没有运行store测试。
fuzz各-run '^$'、精确目标、-count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s，
环境与上述包装器一致。

| 矩阵项 / 回执 | 包 | 范围 | exit | 实际UTC |
| --- | --- | --- | --- | --- |
| 1 / 032 | internal/nodecontrol/nodeissuer | 完整包 | 0 | 2026-10-12T12:29:25.2100083Z → 2026-10-12T12:29:30.6568283Z |
| 2 / 033 | internal/nodecontrol/nodeca | 完整包 | 0 | 2026-10-12T12:29:32.4431567Z → 2026-10-12T12:29:38.1468492Z |
| 3 / 034 | internal/nodecontrol/nodeleaf | 完整包 | 0 | 2026-10-12T12:29:39.9149344Z → 2026-10-12T12:29:45.0515999Z |
| 4 / 035 | internal/nodecontrol/nodeleafkeys | 完整包 | 0 | 2026-10-12T12:29:46.8529928Z → 2026-10-12T12:29:51.9823996Z |
| 5 / 036 | internal/nodecontrol/contracts | 完整包 | 0 | 2026-10-12T12:29:53.9299541Z → 2026-10-12T12:30:18.3740784Z |
| 6 / 037 | internal/nodecontrol/serving | 完整包 | 0 | 2026-10-12T12:30:20.0759844Z → 2026-10-12T12:30:37.6736515Z |
| 7 / 038 | internal/nodecontrol/statebinding | 完整包 | 0 | 2026-10-12T12:30:39.3288506Z → 2026-10-12T12:30:46.1118657Z |
| 8 / 039 | internal/nodecontrol/statecomposition | 完整包 | 0 | 2026-10-12T12:30:47.9005097Z → 2026-10-12T12:30:52.9839548Z |
| 9 / 040 | internal/nodecontrol/statecontract | 完整包 | 0 | 2026-10-12T12:30:54.7052491Z → 2026-10-12T12:30:59.985832Z |
| 10 / 041 | internal/nodecontrol/stateformat | 完整包 | 0 | 2026-10-12T12:31:01.7501505Z → 2026-10-12T12:31:06.4077304Z |
| 11 / 042 | internal/nodecontrol/statematerials | 完整包 | 0 | 2026-10-12T12:31:08.3902104Z → 2026-10-12T12:31:13.5953518Z |
| 12 / 043 | internal/nodecontrol/statepolicy | 完整包 | 0 | 2026-10-12T12:31:15.5326281Z → 2026-10-12T12:31:21.32674Z |
| 13 / 044 | internal/nodecontrol/statepolicybinding | 完整包 | 0 | 2026-10-12T12:31:23.4287816Z → 2026-10-12T12:31:34.1235197Z |
| 14 / 045 | internal/nodecontrol/staterecovery | 完整包 | 0 | 2026-10-12T12:31:35.9292907Z → 2026-10-12T12:31:39.9294971Z |
| 15 / 046 | internal/nodecontrol/statetime | 完整包 | 0 | 2026-10-12T12:31:41.8672633Z → 2026-10-12T12:31:46.8056487Z |
| 16 / 047 | internal/nodecontrol/statetimebinding | 完整包 | 0 | 2026-10-12T12:31:48.6331889Z → 2026-10-12T12:31:56.5645806Z |
| 17 / 048 | internal/nodecontrol/stateverify | 完整包 | 0 | 2026-10-12T12:31:58.4241763Z → 2026-10-12T12:32:02.9939963Z |
| 18 / 049 | internal/readiness | 完整包 | 0 | 2026-10-12T12:32:04.954003Z → 2026-10-12T12:32:11.6271018Z |
| 19 / 050 | internal/nodecontrol/authority | 指定 authority 测试 | 0 | 2026-10-12T12:32:13.6590661Z → 2026-10-12T12:32:24.602228Z |
| 20 / 051 | internal/store | store 仅编译、无测试 | 0 | 2026-10-12T12:32:26.4011747Z → 2026-10-12T12:32:32.9274676Z |
| 21 / 052 | internal/nodecontrol/nodeissuer | FuzzIssuerEnvelope 5s 配置 | 0 | 2026-10-12T12:32:34.7722354Z → 2026-10-12T12:33:50.7799958Z |
| 22 / 053 | internal/nodecontrol/nodeissuer | FuzzIssuerPair 5s 配置 | 0 | 2026-10-12T12:33:52.6460494Z → 2026-10-12T12:34:04.7543354Z |

Envelope：5个初始seeds，75,755 executions，8 new interesting、total13。
Pair：4个初始seeds，6,917 executions，1 new interesting、total5。
没有报告崩溃或失败种子、没有新testdata。
5s是fuzz配置，不是外部总耗时；日志fuzz elapsed约6s。
052外部开始/结束跨约76s，而Go报告包耗时8.435s；
不把外部延迟原因假定为已确定，也不改timeout掩盖它。
001基线也有较长首次启动/构建外部耗时，无失败/超时结论。

019的32次取消控制中10次观察到取消，22次在观察前完成；
允许后一种竞争结果，不能据此声称硬实时取消。
只对观察到的ctx取消给予普通错误/成功前的优先级。
reachable owner擦除、detached Summary和独立owner行为有动态控制；
失败未返回私有owner及parser/runtime不可达副本不被声称动态观察。

## 6. 一次全新上下文独立审查与逐项处置

/root/nodeissuer_final_review；gpt-6-astra/xhigh；fork_turns=none；
只读，未派实施子代理或新开替代reviewer。报告原文保存review-1-verbatim.md。
返回发生在054的2026-10-12T12:41:42.1357026Z快照之前；
消息未暴露秒级返回UTC，不回填不存在的精确时间。

结论：0 Critical /0 Important /0 Minor；
Ready to accept this frozen partial increment: Yes。
没有源代码修订、未决项或Deferred Minor，也没有第二轮/第二位审查。
作者采纳建议，只做正式文档与保全收尾，再向用户请求独立Git保存授权。
最终文档是作者核对，不扩称为已经独立复审的源码。

审查者读完整spec/plan、all24新文件、全部Ruling、旧guard净差异、相关lower；
独立检查27冻结hash、766基线无缺失且仅计划/三guard变化、HEAD/index，
234摘要、13个.NET数学签名验证；检查22个真实Go回执，没有重跑Go测试。

五个强制focus分别通过：same-key/name CA歧义及不同DER/Pending；
原lower error与仅已到达阶段的新优先级；
原full TBS而非重编码/内容片段；重叠buffer和独立Evaluate owner；
已观察取消优先与最终移交。没有把同sentinel动态断言夸大为独立证明
key/Name/ID顺序，相关次序还由源码、guard和审查支持。

审查者以下10类declined-to-judge全部保留，不视为验收：

1. CA自身signature、parent binding、root/完整chain/path：原Pending职责。
2. 可信issuer身份、来源或唯一签发证书：没有可信来源输入，允许same-key/name歧义。
3. 可信当前有效性和leaf/CA窗口包含：无trusted-time输入或窗口策略。
4. 部署purpose、bundle/rollback、identity epochs、issuer/serial唯一性、
   私钥持有和runtime/main-state授权：独立职责，非直接关系层。
5. 生产/provider兼容及旧identifier迁移：合成fixture不足，仍UNKNOWN。
6. 通用DN等价和unknown algorithm parameters完整语义：
   固定策略只比较exact Name bytes、参数外层有界表示。
7. concurrent mutation/Summary/Clear、硬实时、constant-time、
   peak heap、不可达/崩溃时内存擦除：没有这些保证。
8. 穷尽DER或AST guard形式正确性：有限fuzz/回归只是证据。
9. 既有lower较宽test-consumer/subpath行为：未改基线，
   本次只授权三个新exact例外。
10. 整仓/race/Linux/真实DB/Docker/provider/hardware/reboot/PITR及C12/main：
    不在批准限定矩阵中。

## 7. Native执行Ruling与代价（全部）

1. PowerShell端口替代技能Bash workspace/task-start/task-done脚本；
   Windows-only和apply_patch要求优先。代价：簿记可能较弱；
   用brief、BASE、manifest和真实回执保留门禁。
2. 批准的限定矩阵优先于安装模块/整仓验证通用指导；
   离线范围不扩大。代价：无关/全系统回归未验证。
3. 保留本计划scratch、延后Git写入/清理；
   完整原始证据和独立授权要求优先。代价：本地占磁盘，
   未保存前不是可迁移远端检查点。
4. 只额外允许exact in.Leaf.NodeID.Variant()纯值receiver；
   plan普通receiver名单漏列该UUID方法，spec要求RFC4122预检。
   代价：有限guard多一个纯值方法；无新crypto/I/O/输入/运行能力。
5. 审查冻结未提交source及旧guard净差异，而非empty HEAD..HEAD或
   非空commit-range脚本；计划禁止预先提交。代价：没有增量提交历史；
   用完整源码/spec/真实基线/原始回执覆盖净变化。

6. CA自身signature、parent/root/chain/path仍不验收；实际leaf数学验签不认证CA来源，原Pending保留。
   代价：把partial当信任链可能接受不可信issuer。
7. 可信issuer身份/来源/唯一证书仍不认证；same-key/name控制保留不同CA DER，无可信来源输入。
   代价：误读为唯一issuer可能选错来源。
8. 可信当前有效性和leaf/CA窗口包含不评估；没有可信时间/策略输入，lower时间Pending保留。
   代价：过期/未来/越窗材料仍可能满足数学关系。
9. 部署purpose、bundle/rollback、identity epochs、issuer/serial唯一性、私钥持有及runtime/main-state仍独立；无对应输入/能力，原义务保留。
   代价：误用为授权/状态/持有证明会遗漏必要控制。
10. 生产/provider兼容和identifier迁移仍UNKNOWN；只有合成fixture，没有生产验收。
    代价：真实部署材料可能不兼容。
11. 通用DN等价/unknown参数完整语义不认证；批准策略只作raw Name比较和有界外层表示，未发现新具体缺陷。
    代价：语义相同Name可被拒绝，opaque排除schema没有完整解释。
12. 并发/硬实时/constant-time/峰值heap/不可达或crash-time擦除不保证；合同明确排除，只有观察到的取消与可达Clear证据。
    代价：依赖更强调度、资源或擦除承诺的调用方可能不安全。
13. 穷尽DER/AST形式正确性不证明；短fuzz、active负例和源码审查只是有限证据。
    代价：未探索encoding或未来guard绕过仍可能存在。
14. 既有lower较宽测试消费/subpath例外不重审或扩宽；净差异只有三个新exact例外。
    代价：基线原有guard弱点可能仍在。
15. 整仓/race/Linux/DB/Docker/provider/hardware/reboot/PITR和C12/main不验收；批准离线矩阵只覆盖本增量，原高层门禁保留。
    代价：范围外集成、环境或物理故障仍未验证。

finishing-a-development-branch按本计划只保留现有分支/worktree，
不执行其通用全仓命令、合并/PR菜单或清理；这是上述2/3的范围约束，
不是新增合并授权。TDD、独立审查与完成前核对实际限制了实现和交付声明。

## 8. exact27源码冻结（SHA256）

所有最终测试与源码审查使用以下同一freeze1。后续仅文档/checkbox变动；
任何源码改动须重新format/freeze、全部22条及同reviewer复核。

| 路径 | bytes | SHA256 |
| --- | --- | --- |
| internal/nodecontrol/nodeca/boundary_test.go | 24855 | 45E482F6BE43C9C12B25966580A8ACF91F639FA6DAB9E27728A9F1F8C328A988 |
| internal/nodecontrol/nodeissuer/boundary_test.go | 27867 | 4C37561955097A254D46CF2AFE47E14677C650FEFFE376289712E99F83A31713 |
| internal/nodecontrol/nodeissuer/budget.go | 1189 | 55106EE1B0909BA364B779A155AB936E4C690D1BE2B6F3192BC161B7F6FD5278 |
| internal/nodecontrol/nodeissuer/budget_test.go | 4059 | 44A378518EAFB7375B964410DAF9D713DEB37811018E921EE3921E277F403CCE |
| internal/nodecontrol/nodeissuer/der.go | 2868 | 5C544F71102B167651CF75177F1011A07BC32FFA952C988E401E882FFB8C0105 |
| internal/nodecontrol/nodeissuer/der_test.go | 2904 | BF1D5F024E966A9C8E53322FE60A0E0A2B3B48EA99CD1A40DA0FED470C166CEC |
| internal/nodecontrol/nodeissuer/doc.go | 744 | 6FD85524DAD85DA876FC9BA0AC66921706D276E1032DC61078D26CFCF20E4F49 |
| internal/nodecontrol/nodeissuer/errors.go | 540 | 615D0C058DF29A15CAD9180367025F5A50C03E5E80E716A0D4FECC1820B216EB |
| internal/nodecontrol/nodeissuer/evaluate.go | 1885 | 9B49537BEE505EF5989A375ECC12FBC20440907F426CB2F5DAD6FD0F04D5957E |
| internal/nodecontrol/nodeissuer/evaluate_test.go | 10399 | 83F2A8BA5ED86F8E9D3B12A375E48F751087F8A2AE09022CF9D012B36811A169 |
| internal/nodecontrol/nodeissuer/fixtures_test.go | 92561 | E9CDC40C98D07898764EBC3472597364A72294AD2DC5F3063F091AC17F312B30 |
| internal/nodecontrol/nodeissuer/fuzz_test.go | 3185 | 318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E |
| internal/nodecontrol/nodeissuer/input.go | 716 | 3C576692FB70E0052230AD1C46BFC9D0B9CFBE634DA85DF93280A26033E7DDB9 |
| internal/nodecontrol/nodeissuer/lower.go | 2444 | 7FDCD5EA67443387A428F91B8EF48AC203ADA2F9A2A5A116A4AA28724BF90553 |
| internal/nodecontrol/nodeissuer/lower_test.go | 3254 | 99570CD4BCC820A9C7CD1D556E139152F49128F24CCE19ACCB13180762590301 |
| internal/nodecontrol/nodeissuer/material.go | 263 | C3DBA340F8DC76BE46DEC489A957DAC7D713B2F4BD47AFC987988036F36D27E4 |
| internal/nodecontrol/nodeissuer/material_test.go | 1664 | 80356ADB06BEF095E71953A2591BBD17B6A74EA8EB56867EC995102E42874A95 |
| internal/nodecontrol/nodeissuer/owned.go | 1193 | BBC0A5B0659E79F9BE43020C2035CCDE077480DF4CDF457BE28FF56FF2FD8496 |
| internal/nodecontrol/nodeissuer/owned_test.go | 1947 | E5991C0E75B9218FF63FDDD853F82C7FB1283B67BB2D13A86023DBE11997D7A9 |
| internal/nodecontrol/nodeissuer/relation.go | 860 | B04B5178FFB1994C2F77D71BF6B2AD971C324C5AD87255B7141790696E44F8DB |
| internal/nodecontrol/nodeissuer/relation_test.go | 1650 | 055AE53AB8299025DF944A9F54260B4E0F25E23B1238BD6E5021BB130672FFAB |
| internal/nodecontrol/nodeissuer/signature.go | 3002 | 7D7947D4BB5F6FDECE5F897B59FB64631E79A4DF872A852B2DCA3A8BD04EAC5B |
| internal/nodecontrol/nodeissuer/signature_test.go | 8953 | 5BFA9A8B701D11026A0967772EFEA8997204A444B676E94FD0187021F0B6BF21 |
| internal/nodecontrol/nodeissuer/snapshot.go | 2833 | 2C22E03022FEF90F98DEA482C7193A00DEBA941D4DD206046791FD8CF31E703D |
| internal/nodecontrol/nodeissuer/snapshot_test.go | 3221 | 984557C4735D4FCB300316BBF4A609C29736AE4A5A420DE7EDFE454ACF19B7DF |
| internal/nodecontrol/nodeleaf/boundary_test.go | 16057 | B359D686C26A53FB739FE449B8857D536AB4CBAB0727E65AA1D5584B19B886FB |
| internal/nodecontrol/nodeleafkeys/boundary_test.go | 21258 | 27DDB38691235EACD913C29B954C9FD02294075367DBEF9137FAD71A4806DFA9 |

## 9. 保全、历史文本与交付门

054预收尾核对：790非忽略路径=766基线+24新Go；766无缺失；
仅计划/三guard不同；27 freeze无漂移；44旧testdata均原字节不变；
旧guard diff whitespace exit0/空输出；HEAD/branch/index未变。
055把所有[x]还原[ ]后，完整计划原字节SHA256与基线均为
DE05378A79CC8E66AB1E3CFE9E3A7EC13F57F9E3D49A5DD651FCD6277F15A186，
证明只改checkbox。设计原hash仍
B7380C6C4633A339BA28D9E25C884F63BE2D54D26DF6DE85BFA3662F711262B7。

current-status只添加顶部本增量，完整旧正文保留为历史；
去掉新增段并统一CRLF/LF后的历史hash应为
2610EC7F995E088258A16348C319D1B854B717E3D16BCD287D1BEB1E700A2E3F。
最终文档后再次检查应为791非忽略路径、30项本次dirty：
24新Go、三guard、计划、status、执行文档。
无新增产品消费者、旧testdata/config/go.mod/go.sum/源码漂移。
精确增量的空白检查不被扩称为历史全分支空白/格式欠账已经清偿。

local Git config在054只读audit hash为
D6886A65FF6AD1EF560A526C23D99203DBC652BDA14A19E1CCD3EC603983285D。
本次没有配置写调用；此hash用于文档收尾前后核对，
不冒充初始基线未记录的config hash。
最终056保全及交付前057复核的实际结果保留在本计划scratch；
若不通过，不勾Task5完整完成。

本地进度和被忽略的原始证据尚未作为新远端检查点保存。
下一门需用户明确批准暂存/提交/普通开发分支推送；
完整证据如需跨电脑迁移，也须在保存阶段审核并归档，不自动加入Go cache。
不合并main、不建PR、不强推、不删任何分支/worktree。
所有既有Task10/B01/C12/main阻塞与历史验收限制保留。

## 10. 后续批准的Git保存阶段（不回写实施历史）

用户在收到明确归档/暂存/提交/普通开发分支推送建议后回复“批准”。
这是新的保存授权；以上无Git授权、原HEAD/index/本机状态仍为当时事实。
保存前HEAD仍764dcb8b，远端开发分支2877f89c、main e1571f9f。
058只读格式检查exit0/空输出，059–080再次运行同22条限定命令全部exit0，
所有前后27 source hashes仍同原freeze1，未修改任何源码或重新审查。
本次fuzz真实Envelope44次/13初始缓存项，Pair6911次/5初始项、1新增；
不混用先前较大执行量，不作整仓或生产验收声明。

103份本计划根目录原始文本证据逐entry长度/hash核对归档；
保留历史失败/原reviewer报告/所有Ruling和本次真实复测回执，排除Go cache。
归档排序首试失败发生在ZIP创建前；复现及string[]/Ordinal修正对照保存，
未影响源码或测试结果。原始证据不删除、不回填缺失时间。
暂存/提交/推送结果以实际Git回执和聊天交付为准，不能预先声称成功；
不合并main、不建PR、不强推、不清理或扩大原Pending。
见[本次检查点恢复说明与归档清单](2026-10-12-nodeissuer-checkpoint.md)。
