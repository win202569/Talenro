# Node leaf partial-material increment — local execution

日期：2026-10-11  
工作区：`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`  
分支：`codex/c12-b01-task9-coordinator`

## 范围与实际实现

用户批准[书面设计](../superpowers/specs/2026-10-11-node-leaf-material-design.md)、
[五任务计划](../superpowers/plans/2026-10-11-node-leaf-material.md)，选择方案 1 Native：
作者逐项实现，最终一次 fresh-context 独立只读审查。未使用逐任务实现代理。

新增 `internal/nodecontrol/nodeleaf` 固定 20 个 Go 文件（10 产品、10 测试）。
没有修改既有 Go、隔离守卫、SQL、依赖或捕获向量，也没有产品消费者接入。
入口只检查 exact 单 DER、v3/empty subject/P-256/positive minimal serial、
canonical UTC exact 24h、六项固定扩展，以及输入节点/domain 的逐字 URI 和 exact SPKI 绑定。
DER 与 SPKI 分别拥有独立 buffer；共享 owner 的包装复制在 Clear 后共同失效，
Summary 与 DER 输出分离。产品没有签名验证、信任池、时钟、签发、I/O 或运行授权能力。

成功仅为 `NodeLeafMaterialShapeValid`，六项 Pending 固定保留：
LeafSignatureAndIssuerTrust、KeyIdentifierDerivationAndIssuerBinding、
TrustedCurrentTime、IdentityAndAuthorityProvenance、PrivateKeyPossession、
MainStateAndRuntimeBinding。Summary 是可伪造描述值，不是 capability；
零值／已清理结果不是 Ready。过期／未来、密码学签名错误及任意形状合规
SKI/AKI 的真实测试仍产生这些未完成责任；不选择未冻结的 key-ID 派生算法。

## 测试先行与诊断

四任务均先新增测试、实际运行缺声明 RED exit 1，再实现、同 selector GREEN
和完整新包 exit 0。RED 是计划明确允许的初始缺接口编译失败，
不是已复现行为缺陷的证明；GREEN 含实际正负断言。

- Task 1：预算/UUID/domain、固定错误、输入别名、共享包装/分离输出/nil 生命周期。
- Task 2：独立离线真实证书控制、有限 DER/length overflow、P-256 SPKI、
  serial sign-padding 内容长度及时间/年份边界。
- Task 3：六扩展全部 720 种排列、缺失/重复/unknown/noncritical、
  DEFAULT/criticality、闭合内值、1/64-byte key IDs、URI+otherName/registeredID。
- Task 4：完整入口重放拒绝表、两个独立 SHA-256 preimage、
  URI/SPKI exact 绑定、partial-proof 正例、负序列号默认/兼容模式、
  动态成功轨迹的阶段取消/到期/panic、实际 backing-store Clear 与有限 AST 边界。

Task 2 首次 GREEN exit 1：测试专属 synthetic issuer 漏填 PublicKeyAlgorithm，
本地 Go 的 CheckSignatureFrom 在处理实际密钥前拒绝 UnknownPublicKeyAlgorithm。
只补测试 issuer 的 ECDSA 标记，独立离线验签控制随后通过；产品未增加验签。
这一失败及其修正后的真实命令分别保留，不能将首次失败计为通过。

两次 JavaScript 编排器问题均发生在调用 Go 之前（eval 作用域 ReferenceError；
最终矩阵 eval 顶层 await SyntaxError），修正适配器后才运行实际命令，
不是 RED/GREEN 或额外测试结果。

审计限制：Task 1 首轮 RED 的 UTC、命令、输出及实际 exit 1 已保存，但四测试源
哈希是在 session pending 期间、实施前采集，不是独立启动前/退出后成对记录。
没有源修改发生于该 session，亦不据此补造缺失的前后凭据。
以后每条命令及最终矩阵都保存真实开始/结束 UTC、前后源哈希及退出码。
另一历史限制：后续逐任务和首轮矩阵收据的 Command 仅为测试/格式片段，
未逐进程保存完整环境 wrapper；这些进程使用规定设置是作者执行记录，
不能由当时收据独立确认。原收据保留、不回填。为最终交付重新执行相同
17 项，保存实际 FullCommand/Workdir/Login；新证据不消除早期留痕限制。
失败私有 owner 的清理是源码结构支持，未直接观察不可达内存；
AST 守卫是有限结构回归，不是形式化能力/控制流证明。

完整逐次记录在本机保留的
`.superpowers/sdd/2026-10-11-node-leaf-material/`；使用 PowerShell、
本地 Go 1.26.5、离线缓存，GOENV/GOWORK/GOPROXY/GOSUMDB/GOAUTH=off、
GOTOOLCHAIN=local、GOVCS=all:off、GOFLAGS=-mod=readonly。
没有下载依赖或修改永久环境。格式化仅指定上述 20 文件；
格式检查 exit 0 且无输出。

## 最终限定矩阵

每条独立进程，固定 `-count=1 -timeout 120s -p=1`。
以下为作者实际执行，独立 reviewer 不重跑。15 完整包、authority 指定测试、
store 仅编译，共 17 条 exit 0；不是 store 运行测试、全仓或生产验收。

| 包 | 范围 | exit | 开始 UTC | 结束 UTC |
| --- | --- | --- | --- | --- |
| nodeleaf | 完整包测试 | 0 | 2026-10-11T13:14:48.4605586Z | 2026-10-11T13:14:52.6873085Z |
| statetimebinding | 完整包测试 | 0 | 2026-10-11T13:14:53.9855630Z | 2026-10-11T13:14:59.9732946Z |
| statepolicybinding | 完整包测试 | 0 | 2026-10-11T13:15:01.1783449Z | 2026-10-11T13:15:07.1775694Z |
| statebinding | 完整包测试 | 0 | 2026-10-11T13:15:08.3809192Z | 2026-10-11T13:15:13.4253481Z |
| statecontract | 完整包测试 | 0 | 2026-10-11T13:15:14.6253959Z | 2026-10-11T13:15:17.9036701Z |
| statematerials | 完整包测试 | 0 | 2026-10-11T13:15:19.0944584Z | 2026-10-11T13:15:22.6654794Z |
| statecomposition | 完整包测试 | 0 | 2026-10-11T13:15:23.8263396Z | 2026-10-11T13:15:27.7424204Z |
| stateformat | 完整包测试 | 0 | 2026-10-11T13:15:28.9745838Z | 2026-10-11T13:15:32.4026278Z |
| stateverify | 完整包测试 | 0 | 2026-10-11T13:15:33.7555546Z | 2026-10-11T13:15:37.6075905Z |
| statepolicy | 完整包测试 | 0 | 2026-10-11T13:15:38.6522499Z | 2026-10-11T13:15:42.8272641Z |
| staterecovery | 完整包测试 | 0 | 2026-10-11T13:15:44.0165059Z | 2026-10-11T13:15:47.5411072Z |
| statetime | 完整包测试 | 0 | 2026-10-11T13:15:48.7495369Z | 2026-10-11T13:15:52.2511666Z |
| serving | 完整包测试 | 0 | 2026-10-11T13:15:53.4028607Z | 2026-10-11T13:15:58.7461694Z |
| contracts | 完整包测试 | 0 | 2026-10-11T13:15:59.8679806Z | 2026-10-11T13:16:06.3657487Z |
| readiness | 完整包测试 | 0 | 2026-10-11T13:16:07.5761284Z | 2026-10-11T13:16:12.8601859Z |
| authority | 指定 BoundRead / Coordinator 测试 | 0 | 2026-10-11T13:16:13.9957653Z | 2026-10-11T13:16:19.8581732Z |
| store | 仅编译；无 store 测试 | 0 | 2026-10-11T13:16:21.1468326Z | 2026-10-11T13:16:25.0746808Z |

authority selector：`'^(TestBoundRead|TestCoordinator)'`；store selector：`'^$'`。
完整命令与实际输出见对应 `task-5-final-matrix-<package>.json`，
汇总见 `task-5-final-matrix-summary.json`。每条执行期间 20 新源文件均无漂移。
矩阵后 557 既有受保护文件加 20 新文件共 577 条核对：无漂移。

## 审查与裁定

一名 fresh-context gpt-6-astra/xhigh 独立只读 reviewer：0 Critical、0 Important、
2 Minor（上述首轮 paired-hash 与早期 full-wrapper 留痕限制）。作者逐项接受历史
限制并披露成本，不回填旧凭据。为最终验证做一轮新的相同 17 项并完整留痕，
同一 reviewer 检查新收据；没有代码修正或第二 reviewer。

审查完整覆盖 20 新文件与五项重点，独立在 2026-10-11T13:16:50.3007701Z
核对 577 条：无漂移，索引／HEAD／分支也与基线相同。reviewer 没有运行测试、
改文件、联网或再委派。裁定可做限定本机交付，不批准合并、证书使用或生产。
原始审查结论及每项排除判断的作者裁定保留在 independent-review.md 与
review-disposition.md。没有用旧审查替代本次审查；产品源码之后未修改。

## 冻结、文档与本机交付

实施基线 UTC：2026-10-11T12:50:13.9039572Z。
HEAD：`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`。
索引 SHA256：`1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`。
批准时 spec SHA256：
`2CAE43DE6CCEEB3416E68116E941EA0D7E6B7DAE62A94DA7ECC3ABEAB20CE2CC`；
plan SHA256：
`751887E3B1F7F9CDA42BE9AE143C9D017F372966F250BDE4E8C743C33C02CF1E`。
批准文档的需求正文保留，收尾仅更新状态与执行勾选。
current-status 只增加本增量摘要；旧执行记录不改。最终文档由作者核对，
不是额外独立文档审查。

20 新源码/测试保持已审查、已验证的同一冻结，没有审查后代码变化。
557 既有受保护条目中，仅获准的 current-status、本设计与本计划状态修改；
其余 554 条保持基线，全套旧 Go/guard/SQL/module/lock/捕获向量无改动。
原有 297 条未提交状态均保留；仅新增 20 Go 与本执行记录的 21 条状态。
最终只读文件/Git/文档核对收据见 final-closeout.json；没有用文档收尾代替源验证。

没有暂存、提交、推送、合并、重置、删分支、配置变更或 scratch/worktree 清理。
成果仍仅在本机；未形成跨电脑远程恢复检查点。
已有 Task10/B01/C12/main 验收阻塞不解除，生产兼容性仍 UNKNOWN。
未执行全仓/race、真实 DB/Docker、硬件、重启、C11/C12 PITR 或生产 conformance。
签名/issuer 信任、key-ID 派生/issuer 绑定、可信当前时间、身份/authority 来源、
私钥持有、主状态/运行绑定，需要后续独立设计与明确授权。
不承诺并发读写/Clear、immutable strings/不可达运行时副本/fatal OOM 的擦除。

