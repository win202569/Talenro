# 节点证书与条件时间证据绑定 — 执行记录

计划及规格日期：2026-10-12。用户批准计划并选择 Native 顺序执行。
本记录只覆盖 nodecerttime 条件增量，不是全分支、main、生产或 C12 验收。
正式执行/状态文档由作者在最终审查后整理；不能将它们称作另一次独立审查。

## 范围与本机状态

- 已批准规格：[设计](../superpowers/specs/2026-10-12-node-certificate-time-binding-design.md)。
- 已执行计划：[五任务计划](../superpowers/plans/2026-10-12-node-certificate-time-binding.md)。
- 工作树：`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`。
- 分支：`codex/c12-b01-task9-coordinator`。
- BASE/HEAD 保持 `f5d731bef9d767d57ca32f0f4c3a0f532a4bc322`。
- 新增 exact35 Go 文件：17 产品、18 测试，全部在 `internal/nodecontrol/nodecerttime/`。
- 旧源码只修改八个准许的 `boundary_test.go`，增加精确消费者例外及控制；旧产品、旧 API、testdata、模块/配置/SQL 未改。
- 本轮未暂存、提交、推送、合并、建 PR、改 Git 配置或清理工作树/证据。

八个旧测试包为 nodeissuer、statetimebinding、statepolicybinding、
statebinding、statecontract、statematerials、statepolicy、stateverify。
stateverify 文件的新增许可目标仅为 statepolicy，仍不允许直接消费 stateverify。
本增量未添加产品运行消费者。

产品保留完整下层摘要、原 Proof、Pending/Obligations 和原错误。
在真实 nodeissuer 与 statetimebinding 成功后，绑定节点及观测 tuple，
按 CA 先于 leaf 的半开区间检查本次条件 Now，并发布：

```text
NodeCertificateConditionalTimeBound
ObservedAt = Time.Policy.Time.Now
ValidUntil = min(Time.ValidUntil, CA.NotAfter, Leaf.NotAfter)
```

纳秒精度保留；合法 year-one 证书起点不被零值规则误拒绝。
这不是可信当前时间、CA 链信任或运行授权；ValidUntil 不是缓存/复用许可。
新的 Now/观测身份需重新 Evaluate。浅拷贝 Model 共享 Clear 生命周期；
独立 Evaluate 及每次 Summary 的可变集合相互隔离。并发使用不支持。

## 逐任务 RED / GREEN 与失败保留

所有编号对应本计划独有忽略目录中的原始 JSON、stdout/stderr；
命名包含 green 不等于实际通过，以下以真实 exit 和输出为准。

| 任务 | 实际 RED/诊断 | 实际 GREEN / task-done |
| --- | --- | --- |
| 1 合同/预算/八消费者例外 | 001/002 缺失合同接口的编译失败；001 还修正测试辅助声明；003–010 八旧测试分别拒绝新精确正例或错误放行子路径 | 011 合同、012–019 八旧测试、020 新包完整测试均 exit0 |
| 2 所有权/规范化/生命周期 | 021 缺失接口编译失败；023/024 故意反转 Kind/P 排序的控制均真正失败，随后恢复正确实现 | 022 局部；025 全部命名控制；026 新包完整测试均 exit0 |
| 3 固定材料/两次 lower/同快照绑定 | 033 缺失 lower/binding 接口编译失败；027 构造器字段、029/030 checker schema/type 错误分别保留，不算行为 RED | 028 生成；031 独立 checker；032 普通固定材料控制；034 命名测试；035 新包完整测试均 exit0 |
| 4 窗口/最小上限/Evaluate | 036/037 缺失接口编译失败；036/038 测试字段修正、039 把 leaf 的512上限误用于 CA 的8192上限，均如实保留为测试修正而非产品缺陷 | 040 命名测试及041新包完整测试均 exit0 |
| 5 守卫/矩阵/审查/收尾 | 042 缺失 guard 接口编译失败；043 guard 合法 defer FuncLit 误拒绝与 schema 初始化问题；046 扩展 manifest 控制对原 manifest 真正失败；067 错误 readiness 路径 setup failed | 044 guards；045 独立扩展 checker；047 固定 manifest/守卫/取消控制；23 条最终有效命令；073 task-done 均 exit0 |

073 为审查后的新包完整复核：
`test ./internal/nodecontrol/nodecerttime -count=1 -p=1 -timeout=120s`，
真实 exit0，Go 报告 21.708s；不将此值称为整个进程耗时。
审查后没有 Go 源码修改，只有本计划状态、账本和正式执行/项目状态文档。

完整预复制预算先于 required-empty；包括 late History 超限同时空 anchor 的优先级。
只声明输入 raw 上限 43,917,824 bytes，不声明总 heap 上限。
部分复制、lower 返回后立即登记 Clear、Time 可变摘要直接写入受保护 owner、
最终 context 之后才发 Proof/valid/包装的次序均受源码控制。
602 个可观察 Err 回调逐点覆盖 cancel/deadline/原 panic 传播及调用者输入不变；
不将此控制称为硬实时取消或直接观察失败私有 owner 擦除。

## 固定签名材料与独立核对

- 仅本计划 scratch 中一次性构造 34 个完整 Input；没有永久生成器进入35文件库存。
- 生成器采用独立确定的 CA/leaf P-256 密钥及 root/metadata/time Ed25519 seed；
  ECDSA 签名生成后被冻结，普通测试不随机重签。
- 普通测试源内嵌固定 gzip/base64 JSON 及 manifest，每次解码为独立值；
  不写旧 fixture/testdata、不借用其他计划证据。
- 031/045 的独立 checker 重新核对 exact DER/SPKI/TBS 摘要、leaf 数学签名、
  JCS raw envelope/payload 的 Ed25519 签名、material reference payload/envelope 摘要及两真实 lower 结果。
- 控制含正常/跨节点、history0/32、CA/leaf 起止 ±1ns/精确边界、
  CA/leaf/时间先到期与全部 ties、leaf 不包含于 CA 窗口、
  year-one 证书端点、错误签名/nonce/metadata/rollback/缺失观测/typed PolicyFailure。
- 下层有效正控制先建立，才测跨节点拒绝；私有人工 Summary mutation 明确不是公开输入信任路径。
- frozen 两常量被 reviewer 独立解码并与本计划 JSON 逐字节比较；34 packets 一致。

七个被禁止的包 stateformat/statecomposition/stateverify/statetime/nodeca/nodeleaf/
nodeleafkeys 不由新产品或普通新测试直接 import。
AST 显式能力检查加17产品实现版本 schema，安全局部/import 改名和格式仍接受；
生命周期和消费者反控制都是有限回归屏障，不是形式证明。

## 最终格式与23条限定命令

048 对35新 Go+8旧测试共43路径做 `gofmt -l`：
空输出、真实 exit0，前后摘要不变。没有整仓格式重写。

下表每行独立命令，均真实 exit0。
完整包统一 `test <path> -count=1 -p=1 -timeout=120s`；
authority/store/fuzz 的选择器如下明确区分。

| 回执 | 路径 | 范围 / 结果 |
| --- | --- | --- |
| 049 | ./internal/nodecontrol/nodecerttime | 完整包 PASS |
| 050 | ./internal/nodecontrol/nodeissuer | 完整包 PASS |
| 051 | ./internal/nodecontrol/nodeca | 完整包 PASS |
| 052 | ./internal/nodecontrol/nodeleaf | 完整包 PASS |
| 053 | ./internal/nodecontrol/nodeleafkeys | 完整包 PASS |
| 054 | ./internal/nodecontrol/contracts | 完整包 PASS |
| 055 | ./internal/nodecontrol/serving | 完整包 PASS |
| 056 | ./internal/nodecontrol/statebinding | 完整包 PASS |
| 057 | ./internal/nodecontrol/statecomposition | 完整包 PASS |
| 058 | ./internal/nodecontrol/statecontract | 完整包 PASS |
| 059 | ./internal/nodecontrol/stateformat | 完整包 PASS |
| 060 | ./internal/nodecontrol/statematerials | 完整包 PASS |
| 061 | ./internal/nodecontrol/statepolicy | 完整包 PASS |
| 062 | ./internal/nodecontrol/statepolicybinding | 完整包 PASS |
| 063 | ./internal/nodecontrol/staterecovery | 完整包 PASS |
| 064 | ./internal/nodecontrol/statetime | 完整包 PASS |
| 065 | ./internal/nodecontrol/statetimebinding | 完整包 PASS |
| 066 | ./internal/nodecontrol/stateverify | 完整包 PASS |
| 072 | ./internal/readiness | 完整包 PASS，纠正计划路径笔误 |
| 068 | ./internal/nodecontrol/authority | 指定 BoundRead/Coordinator 测试 PASS |
| 069 | ./internal/store | 仅编译，明确 no tests to run |
| 070 | ./internal/nodecontrol/nodeissuer | FuzzIssuerEnvelope 短 fuzz PASS |
| 071 | ./internal/nodecontrol/nodeissuer | FuzzIssuerPair 短 fuzz PASS |

authority：`-run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s`。
store：`-run '^$' -count=1 -p=1 -timeout=120s`，未跑 store 测试。
两 fuzz：`-run '^$' -fuzz '^FuzzIssuerEnvelope$'` 或
`'^FuzzIssuerPair$'`，各 `-fuzztime=5s -parallel=1 -p=1 -timeout=30s`。

| 短 fuzz | 初始 seeds | 实际 executions / new interesting | Go 报告耗时 | 原始进程 UTC 差 |
| --- | --- | --- | --- | --- |
| 070 Envelope | 5 | 110,263 / 4（total9） | 7.116s | 56.6840272s |
| 071 Pair | 4 | 8 / 0（total4） | 8.747s | 11.1220882s |

5s 是配置的短 fuzz 时段，既不是进程总时长，也不是穷尽验证。
包括首次 instrumentation 编译/启动在内的时间不能武断归因为产品失败；
Pair 只有8次实际 executions，尤其有限。

067 实际执行 `./internal/nodecontrol/readiness`，exit1、directory not found，
没有测试开始。此失败保留，不改写为通过。
包声明与目录检查证实 readiness/store 位于 `internal/`；
按账本 Ruling 只纠正路径，不改选择器/超时/依赖/源码。
store 错误路径未再无意义尝试。因此最终矩阵是24次实际执行、
其中23条有效选择通过，而不是“所有实际尝试均通过”。

## 源码冻结、证据与 Git 边界

PowerShell runner 使用缓存 Go1.26.5/windows/amd64、本计划专用 GOCACHE；
GOTOOLCHAIN=local、GOPROXY/GOSUMDB/GOAUTH=off、GOVCS=all:off、
GOFLAGS=-mod=readonly。未下载依赖或扩大验收。
回执保留实际 UTC、选择器、exit、完整 stdout/stderr、允许环境字段、
每条命令前后的当时源码摘要。最终矩阵及073的范围是 `internal/nodecontrol`
下377个非 testdata Go 文件；早期按阶段保留当时的实际文件集合，不追补为377。
前20条已有完整最终回执；021起额外保存 started 标记以识别未完成运行。

所有23条有效最终命令的 before/after 与共同377源码集相同；
048 的43路径与相应当前源一致。reviewer 独立重算确认零漂移。
073 再次完整包复核；pre-closeout-audit.json 与最终 closeout-audit.json
记录作者收尾核对的摘要、exact47 Git状态范围、空暂存差异和 Git 保持。
这不是整仓摘要冻结，尤其不代表包含所有 readiness/store 源码。

批准规格 SHA256 保持：
`595A4471C642727AE60D8C816BDEDAA65FD3B9F21EF812B503CDD12F91CE694B`。

本轮 index SHA256 保持：
`15ADFA44B16BB2020BE7623FC44CC334F2B50D88AA2D9341152A7421C2312DE1`。

本轮 Git config SHA256 保持：
`D6886A65FF6AD1EF560A526C23D99203DBC652BDA14A19E1CCD3EC603983285D`。

原始证据全部在
`.superpowers/sdd/2026-10-12-node-certificate-time-binding/`，
含 progress/五 briefs/review-package/final-review、固定 Input/manifest、
一次性生成/独立核对器、001–073编号回执及完整输出。
其中048是格式回执，其余72是 Go 回执；所有实际失败均保留。
该目录按既有规则被忽略，未进入 Git。
不删除、清理或把其他计划 scratch 算作本轮归档。

## 一次 fresh 审查与作者处置

一次 fresh-context gpt-6-astra/xhigh 审查，本增量 untracked 源也包含在范围内。
审查者只读、无测试/构造器/格式/构建/网络/Git写入，
独立检查35新文件、八实际旧测试差异、批准要求、回执和当前源码摘要/固定常量。
测试通过均为作者运行的原始证据，不是 reviewer 重跑。

实际结论：0 Critical、0 Important、1 Minor。
作者以使用影响重新分级：保持 Minor——当前是隔离无运行消费者的条件增量，
完整规格已有合同，未发现对应实现缺陷，但仅看 API 的读者仍可能误解。
没有 Critical/Important 修复 pass、没有第二 reviewer、没有审查后 Go 修改。

### Deferred minors

M1：补充 exported API 注释，明确新的观测须重新评估、
ValidUntil 不允许缓存或授权、Model 浅拷贝共享 Clear 失效，
以及不支持 concurrent Clear/read。本轮按 Native 执行规则延期，不秘密修入。
此项必须在未来集成前继续可见，不能写“审查零问题”。

### Rulings I made（全部、按发生顺序）

以下25条为账本原文：前3条执行裁定，后22条逐项处置 reviewer 的
Declined-to-judge。各项都保留误判成本；未知/范围外不转化为通过。

- Ruling: Use native PowerShell equivalents for Bash-only workspace/task-start/task-done/review packaging scripts — user's PowerShell-only constraint and evidence-preserving no-Git plan override defaults — cost if wrong: bookkeeping must be audited against actual command receipts.

- Ruling: No dependency download/full-repo baseline, no commits or scratch deletion — approved plan explicitly freezes cached offline dependencies, limited matrix and evidence preservation — cost if wrong: broader integration/acceptance remains unverified.

- Ruling: Use actual ./internal/readiness and ./internal/store for the named readiness/full and store/compile-only validations — rg package declarations and directory checks prove both are at internal/, while spec names these packages without the mistaken nodecontrol prefix — cost if wrong: recorded command differs from plan typo; 067 failure stays preserved and does not count as passing. No source, timeout, dependency or test-selection expansion; store bad path is not needlessly attempted.

- Final: Ruling: R04 current-time authenticity and observation provenance — a user gets only a conditional caller-supplied observation, not a trusted clock/provider; the proof remains explicitly conditional — cost if wrong: untrusted time could be mistaken for real certificate validity.

- Final: Ruling: R05 atomic external capture — independent owned copying protects the supplied values but does not make an external read atomic; no capture API is added — cost if wrong: mixed platform observations could be mistaken for one physical snapshot.

- Final: Ruling: R06 TLS/transport authentication — no transport verifier is implemented or invoked; the result grants no channel trust — cost if wrong: unauthenticated transport could be accepted by a later consumer.

- Final: Ruling: R07 atomic nonce consumption — comparison/evaluation does not consume a nonce; no transactional replay guarantee is supplied — cost if wrong: a later consumer could permit replay.

- Final: Ruling: R08 durable checkpoint/physical commit — this pure evaluation performs no persistent advancement or commit; caller persistence remains separate — cost if wrong: restart could lose or reuse an observation.

- Final: Ruling: R09 CA provenance/self-signature/parent-root-chain/path/deployment-purpose trust — direct issuer mathematics and inherited Pending are retained, not promoted to deployment trust — cost if wrong: an untrusted CA could be interpreted as authorized.

- Final: Ruling: R10 trust-bundle authorization and rollback history — retain lower conditional policy checks without adding trusted bundle integration or durable complete-history evidence — cost if wrong: unauthorized bundles or rollback could be mistaken for trusted state.

- Final: Ruling: R11 main-state selected leaf/issuer/serial/public-key/key-handle/epoch/lineage/row correspondence — those relations are not bound by this increment; existing Pending remain visible — cost if wrong: unrelated certificate material could be used for a main-state identity.

- Final: Ruling: R12 certificate authority-epoch/metadata-domain authentication — no equality is invented without a signed transcript contract; these extra relations remain unproved — cost if wrong: mismatched authority/domain could be mistaken for authenticated association.

- Final: Ruling: R13 time-signer authentication of certificate digests — the time transcript omits them; composition does not claim that they were signed — cost if wrong: a mathematical composition could be mistaken for a jointly signed certificate/time attestation.

- Final: Ruling: R14 full issuance template/issuer-time/skew/whole-leaf-window containment — only the approved half-open observed-time windows and exact minimum are enforced; wider leaf windows are deliberately permitted — cost if wrong: a future deployment's additional certificate policy could be bypassed.

- Final: Ruling: R15 private-key possession/hardware protection — no key operation or hardware evidence exists; public certificate evidence proves neither — cost if wrong: an unavailable or unprotected key could be treated as possessed/protected.

- Final: Ruling: R16 State/Commit/Ready/Authorize/startup/recovery/connection permission — no runtime consumer or authorization route is added; the model is evidence only — cost if wrong: conditional evidence could trigger unsafe runtime authorization.

- Final: Ruling: R17 fresh-install/missing-Desired/changed-root-metadata support — preserve the existing time-binding evaluator's supported shapes and rejection/error behavior — cost if wrong: unsupported installation or transition cases could be incorrectly advertised as supported.

- Final: Ruling: R18 concurrent mutation/Model-Summary-Clear use — shared lifecycle is sequential and concurrency is unsupported; owned copies do not supply linearization — cost if wrong: a caller could race reads, mutation or Clear.

- Final: Ruling: R19 hard-real-time cancellation/peak memory — only observable callback cancellation and declared raw caps are evidenced, not latency deadlines or total heap ceilings — cost if wrong: deployment resource or latency limits could be exceeded.

- Final: Ruling: R20 strings/unreachable-runtime copies/fatal OOM-crash erasure — cleanup covers reachable owned mutable values under ordinary unwind, not full-process or crash-time erasure; inaccessible failure owners are source-supported, not directly observed — cost if wrong: secrets or evidence copies could remain outside the stated boundary.

- Final: Ruling: R21 formal capability/heap-safety proof — finite AST version locks and mutation controls are regression barriers, not a formal sandbox — cost if wrong: future bypasses could evade the finite controls.

- Final: Ruling: R22 exhaustive crypto-fixture coverage/fresh audit of every lower — 34 fixed controls and short fuzzing support this composition; inherited algorithms are not comprehensively re-audited — cost if wrong: untested cryptographic or lower-layer defects could remain.

- Final: Ruling: R23 independent dynamic reproduction — reviewer source/hash/constant checks are independent, while every test outcome remains author-run evidence — cost if wrong: the reader could overestimate independent dynamic verification.

- Final: Ruling: R24 whole-repo/race-Linux/DB-Docker/real TLS-provider/hardware-reboot-PITR — the finite native matrix supplies none of these broader acceptance claims; production/provider compatibility stays UNKNOWN — cost if wrong: environment or integration failures could remain undiscovered.

- Final: Ruling: R25 remote recovery/merge-release/Task10-B01-C12-main blockers — local evidence and modifications are retained without Git writes or acceptance promotion; a separate save/push approval remains necessary — cost if wrong: progress could be unavailable on another computer or mistaken for release-ready.

## 仍未认证与下一授权

原下层全部 Pending/Obligations、生产/provider 兼容 UNKNOWN 和
Task10/B01/C12/main 阻塞保留。证书与条件时间可组合不等于可信 current-time、
平台原子采集、TLS、nonce 消耗、持久 checkpoint/physical commit、
部署 CA/链/路径/目的、selected leaf/private key/main-state identity/runtime authority。
不支持未被原 time-binding 支持的安装/transition 形态。

未运行整仓/race/Linux、真实 DB/Docker、TLS/provider、硬件/reboot/PITR 验收。
既有失败/取消/未验收历史不被本轮覆盖或回写。
普通 unwind 下可达 owned 缓冲清理、raw byte cap 和有限 AST 控制不保证
并发、硬实时、全 heap、不可达/字符串/runtime/crash-time 擦除或形式安全。

本轮新增实现与忽略证据尚只在本机，不能靠远端检查点在新电脑恢复这些新增内容。
建议下一步单独批准：归档本计划必要原始证据（不含缓存）、
暂存/提交必要文件并普通推送现有开发分支。
这不包括合并 main、PR、强推、配置修改或删除分支/工作树/证据。
