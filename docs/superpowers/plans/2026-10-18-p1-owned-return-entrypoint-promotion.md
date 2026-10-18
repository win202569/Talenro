# P1 Owned-return Entrypoint Promotion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task after explicit approval. Steps use checkbox syntax. Proposed Native execution only; no delegation or new reviewer allowance.

**Goal:** 将已获 PS7 限定离线验收的固定候选按原字节提升至开发分支的原 P1 离线入口，同时保存旧版本和可追溯的证据边界。

**Architecture:** 仅替换一个离线产品文件，不重写算法或切换所有历史诊断路由。候选、旧原产品和新入口身份分开记录。新入口验证使用原目录的两份字节相同测试，原始证据/旧预算冻结不动。

**Tech Stack:** Windows PowerShell；固定已有 x64 PS7 测试宿主；Git 开发分支保存。无安装、原生编译或新增依赖。

**Spec:** [PS7 限定验收批准记录](../../roadmap/2026-10-18-p1-owned-return-ps7-limited-acceptance-approval.md)，原 [P1 合同](2026-10-17-offline-supervision-policy.md) 的被动输入/输出与预算规则不变。

现状：用户在 cca8c0c6 后批准原计划，准备曾因备份换行错误停止；在 f436d139 后
明确批准只修正备份/必要准备文件、保留失败记录，在原最多两次预算内继续。
恢复已完成原入口字节替换及两次原目录 PS7 回归，2/2 已用；无新增审查。
见[恢复结果](../../roadmap/2026-10-18-p1-owned-return-entrypoint-promotion-recovery-result.md)及原[停止结果](../../roadmap/2026-10-18-p1-owned-return-entrypoint-promotion-result.md)。
原提案基线 eb7b3788，原执行 cca8c0c6，恢复执行 f436d139；提交保存按交接核实。

## Global Constraints

- 复用现有 worktree/分支 codex/c12-b01-task9-coordinator；不合并 main，不发布。
- PS7 限定离线范围及 O1–O23/N1–N9 全部保留。PS5/完整双引擎 P1 未验收；文档标记未验证，不新增引擎 guard 或宣称不兼容。
- 产品 public Get-OfflineSupervisionDecision([object]$Scenario)、13 根字段、22 Decision 字段不变。
- 120000ms 合成绝对预算、113500ms 等值截止、共享 500/5000/1000ms 阶段预算不变。
- 产品 Mode=policy-only、EvidenceClass=synthetic；OriginVerified/Acceptance/RuntimeApproved/LiveActionsPerformed=false，CleanupStatus=unknown。
- 无产品 Process/Job/pipe/clock/file/network/security API、回调或 imports；候选提升不是 P2/P3/live/C12 安全或运行放行。
- 候选/旧报告/旧控制器/旧测试/旧 pin JSON/ignored 原始包不修改；旧额度不重用。
- 不从 generic RED/GREEN、全项目检查、审查技能自动增加测试、变异或新 reviewer。此为固定已审源码的路径提升，不是新修复；代价是没有提升后独立复审。

## Review Focus

1. 路径错误：新入口必须与已验收候选字节一致，验证宿主必须导入原目录入口而非候选。
2. 历史身份丢失：旧入口字节须另存，旧清单不重写，不能声称提升后全部旧工作区 pin 仍匹配。
3. 合同漂移：完整 API 原 35 组、38 行及 4096/4097 容量/失败索引断言不变。
4. 装饰/返回别名：旧所有权 8 fixture+10 regression 组不变；新 plain-return 向量采用同字节候选既有证据，不冒称新入口重跑矩阵。
5. 运行范围扩大：原入口历史 PS5/probe 调用者不自动获准重跑，不授予 PS5 或实际监督能力。

## Fixed files and permitted future writes

所有产品路径以下以前缀 docs/diagnostics/isolated-fuzz/ 为准。

| 角色 | 路径 | 批准后所需物理 SHA256 |
| --- | --- | --- |
| 源候选（不改） | scalar-owned-return-candidate/offline-supervision-policy.ps1 | 1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8 |
| 原入口（唯一产品替换） | offline-supervision-policy.ps1 | 由 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA 变为上行候选身份 |
| 旧版本留存（新增） | owned-return-entrypoint-promotion-v1/baseline-policy.ps1.txt | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA；非可执行历史文本，不导入 |
| full 测试（不改） | offline-supervision-policy.tests.ps1 | 61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3 |
| ownership 测试（不改） | offline-supervision-policy.ownership.tests.ps1 | C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7 |

允许另增同一 promotion-v1 目录中的 capture.ps1、source-pins.json、expected-output.json、README.md、.gitattributes，供下述有界控制与恢复使用；创建前冻结各自实际摘要。
允许新增 docs/roadmap/2026-10-18-p1-owned-return-entrypoint-promotion-result.md，更新本计划执行勾选及 current-status/repository-recovery。
不修改 scalar-owned-return-candidate/README/source-pins 或任何旧包以假装其历史状态已变化。

## Task 1: 单次固定版本路径提升和两次新验证

**Interfaces:** 消费以上已批准候选字节、两份独立 literal 测试及已有 38/22 行 oracle；产出相同 public API 的原入口和一个独立 promotion 记录。无新 public 参数或 engine guard。

- [x] **Step 1 — preflight/freeze before any product write.** 核对当时 HEAD/干净工作区/分支/origin、批准实际回复、上表旧/新/测试身份及验收报告。完整核对现有 66 审查输入和 676 继承条目；任一异常停止，不用旧额度诊断执行。
- [x] **Step 2 — preserve baseline and prepare portable controller.** 通过可审阅补丁保存旧入口非可执行文本，校验旧 SHA 后才允许覆盖入口。新 packet 使用 .superpowers/sdd/2026-10-18-p1-owned-return-entrypoint-promotion-v1/，不覆盖旧文件。控制器仅允许 Prepare、full、ownership 三动作，只有后两者启动目标宿主。预期 oracle 从已冻结独立 38/22 行清单采用，仅预先解析 ordinary CONTEXT 占位符，不由产品输出生成。
- [x] **Step 3 — bind intentional identity transition.** 新 source-pins 分开记录旧 HEAD/旧原入口身份、baseline 文本、候选、新入口/测试/控制器/oracle/引擎身份，以及所有继承条目的原始路径/摘要。继承条目若且仅若路径精确等于被替换入口，标记 historical-destination，映射 baseline 文本校验旧 SHA；所有其他继承路径仍按原值校验。记录具体条目数，绝不全局忽略 hash mismatch、重写旧 manifest 或称全部旧 live-path pin 仍完整。
- [x] **Step 4 — exact source promotion.** 用 apply_patch 将原入口替换成候选同字节内容，不复制测试或改算法；按原换行策略核验新入口 SHA=1971F808…、baseline SHA=953CC2A8…、候选/测试不变。静态 parser 检查产品与新增控制器，无 dot-source/import/执行探测。差异只能是固定候选累计 Integer/Token/owned-return 改动和本计划列出的新记录文件。失败按下文停止。
- [x] **Step 5 — NEW slot 1: original-path full API.** 仅普通继承令牌 PS7，-NoProfile -NonInteractive -File <原目录 offline-supervision-policy.tests.ps1> -Group all。成功要求 exit0、精确 38 行/35 PASS 组、stderr 空、owned exit、流完整、全部新身份与历史映射一致。不能调用候选目录测试冒充原路径验证。
- [x] **Step 6 — NEW slot 2: original-path ownership.** 仅 Step 5 全部成功后启动同引擎原目录 ownership.tests.ps1 -Group regression。成功要求 exit0、精确 22 行/8 fixture+10 regression 组，以及与 Step 5 相同的严格证据检查。不能放宽断言或排除慢项。
- [x] **Step 7 — read-only reconcile and handoff.** 逐项读取新回执、流和身份，记录实际路径、原/新 SHA、两个结果与所有保留门槛。满足才记 PROMOTED_ON_DEVELOPMENT_BRANCH_PS7_LIMITED，仍非 main/release/完整 P1；未满足记录 UNFINISHED_STOPPED。仅提交明确文件并普通推送开发分支，确认远端 SHA/clean 状态。

## Approved finite allowance — original 2/2 consumed after explicit recovery approval

最多两个顺序 fresh ordinary PS7 目标宿主，不启动 baseline/red/旧 focused/mutant/PS5/benchmark/全项目宿主。
固定已有 PS7：C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，
SHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139。
每次 30000ms 执行 +2000ms owned exit +2000ms stream drain，不加时。
异步并发 raw-byte stdout/stderr 捕获，slot 消费必须先持久记录再启动；禁止为测试提权。
stderr 必须空，不沿用更宽 CLIXML 例外。任何未知上下文、非预期输出、异常、超时、缺失/漂移或 drain 失败立即停止余项。
超时只能处理精确 owned 目标句柄，不扫描/终止其他 PID；清理不明如实记录 unknown。
Git-only 提权若需要，单独受平台权限审核，不作为普通测试宿主提权或安全许可。
不申请新独立 reviewer：沿用已审同字节产品的限定结论和本轮父代理路径/回执核对，明确没有提升后独立复审。用户可拒绝该证据采用方式；不得自动补派。

## Failure and recovery

任何失败停止，不修产品/测试/清单、不重跑、不续额度、不自动回退源文件。保留完整失败记录，入口若已替换则明确标为 staged promotion / 未完成，不宣称已提升成功。
旧源码可从保留的 baseline 文本及 eb7b3788 取回；需要回退时单独报告确切目标/当前差异并取得方向，再通过普通补丁恢复该文件，不用 reset/checkout 或抹去用户改动。
原候选已批准的有限验收不因路径验证失败被改写；路径提升是否成功是另一个事实。
旧 pin 消费者仍需原历史版本/独立旧快照，不自动改路由、迁移旧控制器或启动它们。

## Self-review and approval handoff

本计划只处理已批准候选的原路径采用，不实现新功能。五类 Review Focus 由 Step 1–7 的身份/历史映射/API/ownership/范围检查分别约束；没有追加缺失 PS5、产品 mutation 或真实运行证明。
当前只读检索发现原入口被既有诊断脚本/历史 pin 引用；这是潜在导入路由影响，不是这些脚本重新获准执行。候选未发现 PSScriptRoot/PSCommandPath/Join-Path/import 路径依赖；两对测试物理摘要相同。这些是文本/身份检查，不是新运行证据。
恢复审批明确允许必要准备路径/元数据修正，故使用独立 owned-return-entrypoint-promotion-recovery-v1/ leaf/packet 和正确 LF 备份；旧失败 leaf/packet/report 原样保留。
Step1–7 完成；源码/证据提交 5a9dc4dec616600239c2e569ca3f82bc2cbc4206 已推送，远端同 SHA、工作区干净经实际核实。本次仅补记保存确认。2/2 用尽，不再重跑或自动增加审查。所有原范围与门槛不变。
