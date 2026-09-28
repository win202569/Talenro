# 换机接续说明 — 2026-09-25

本文件保存继续开发所需的决定、证据入口和未完成事项。它是未验收开发检查点，
不是全套测试通过、C12 完成或 main 合并记录。

2026-09-27 本机最新工作树状态见[任务 2 未验收检查点](2026-09-27-devtools-task2-checkpoint.md)：
原 207 项基线准备阻塞已解除；模块/入口迁移尚未提交，完整模块图版本差异仍未接受。
这些本机未提交内容尚不能从 GitHub 恢复；下述历史提交信息不代表本次实现已推送。

2026-09-28 接续：四个 tools 精确候选已按新批准范围准备，工具严格离线完整模块/图查询通过；新增纯规则测试通过，但真实验收未完成。根 cel.dev/expr v0.25.2 的单项约束被正常 tidy 删除，四锁回到实验前哈希，按 R4 暂停恢复，等待稳定来源调查或合同修订方向。详见[最新版本证据](2026-09-27-devtools-version-lock-evidence.md#2026-09-28-获批接续四候选准备与根恢复阻塞)。未提交、未推送，不能在另一台电脑获取本次新增工作。

## 从另一台电脑恢复

当前审批阶段：用户已确认根 28 书面设计；[原实施计划](../superpowers/plans/2026-09-25-c12-development-tool-isolation.md)已更新为 root31/tools4 条件候选流程，等待计划审阅。R3c-2 单列 27 个缺缓存精确版本，准备权限尚未授权；perfstat 四文件齐备仅作只读核验。保留 Native inline，不执行实现/下载，不推送。下方历史待审记录按阶段阅读。

最新阶段：[根 28 项精确条件例外正式设计](../superpowers/specs/2026-09-28-c12-devtools-root28-graph-exceptions-design.md)已撰写，待书面审阅。用户只批准方案方向及撰写设计；尚未批准该 spec、计划或新增准备清单。现行策略/锁文件未改，不开始实现；既有 Native inline 选择不重选。设计尚未提交推送。

2026-09-28 评估接续：[根 28 项逐项评估与修订建议](2026-09-28-devtools-root28-assessment-proposal.md)已保存，22 项直接父边/6 项历史图分类。28 项 why 成功且无实际包路径，严格版本元数据批次因离线缺缓存失败；候选未验收，规则未修改，下载未授权，等待方案方向审阅。报告未推送。

最新补充：[根图级版本稳定性调查](2026-09-28-devtools-root-version-stability-investigation.md)。已确认 cel 较高版本来源随 Buf 迁出，根当前无实际包引用路径；在现有固定依赖和隔离合同下未找到稳定恢复来源。未扩大例外、未修改四锁；后续需审阅合同冲突，不重复失败 pin。此报告同样尚未推送。

仓库： https://github.com/win202569/Talenro.git

进度分支：`codex/c12-b01-task9-coordinator`。最新 Windows 实现提交为 `e5ab7f8b`；
本说明及诊断报告在其后的文档提交中。应取得该分支最新提交，不只下载 main。

新目录中使用 PowerShell 执行：

```powershell
git clone --branch codex/c12-b01-task9-coordinator --single-branch https://github.com/win202569/Talenro.git Talenro
Set-Location Talenro
git status --short --branch
git log -3 --oneline
```

已有本地仓库时，先保存自己的未提交更改，再 fetch 并切换到上述分支；不要强制覆盖。
Git 恢复源码、测试、计划和文档，不恢复已安装软件、凭据、模块缓存或后台进程。

## 已批准的方向（无需重新选择方案）

- 按[实施计划](../superpowers/plans/2026-09-25-c12-development-tool-isolation.md)
  在当前任务内顺序执行（Native inline），保留验收阻塞。
- [原设计](../superpowers/specs/2026-09-25-c12-development-tool-isolation-design.md)
  及 [Windows 附录](../superpowers/specs/2026-09-25-c12-devtools-windows-powershell-addendum-design.md)
  已获批准；后续精确版本修订也已批准。
- 四个 Windows 开发工具入口最终只支持独立 PowerShell Core 7.6.5 稳定版，
  通过 pwsh.exe 调用；旧 5.1、其他版本、预发布版本提前拒绝。
  子入口使用当前已验证宿主的绝对路径，不自动安装、升级或静默切换。
- Windows 对应 Bash 入口提前拒绝；原生 Unix Bash 保留。既有 Bash smoke 和
  Windows C12 runner 不在此次替换范围内。WSL 测试不替代 Windows C12 物理验收。
- 不引入编码加载器、外层原生监督程序或新的预编译初始化程序集，不修改防护设置。
- 工具模块计划拆为 tools/devtools（talenro.local/devtools），业务模块保持
  talenro.local/platform；不引入 go.work、跨模块 require/replace 或工具版本变化。
  具体版本、原子迁移范围和验收条件以批准计划为准。

## 当前实现及证据

任务 1 仅部分完成，任务 2/3 尚未开始；详见
[Windows 实现检查点](2026-09-25-devtools-windows-implementation-checkpoint.md)。

- 新独立 verify-devtools.ps1：版本前置拒绝、离线依赖完整性检查、900 秒共享预算、
  环境恢复、私有非继承 kill-on-close Job、运行中 stdout/stderr 合计 4 MiB 保留上限。
- 私有帮助文件 scripts/private/devtools-process.ps1 仅加载不初始化。
- verify-devtools.sh 已实现 Windows 提前拒绝；原生 Unix 输出限制及取消行为仍需实现/验证。
- 三个 e2e 测试文件覆盖实际运行时、归属、完整性、取消、输出限制等边界；
  尚未接入其他消费者，尚未拆分模块。

前一轮已记录的专项回归（本次文档归档不重新运行）：

```text
go test -v -tags=e2e ./internal/e2e -run '^TestDevtoolsVerification(PowerShell|Bash)$' -count=1 -timeout=5m
```

历史结果：退出 0，63.767 秒，28 个子用例 PASS、0 FAIL、1 个不适用 Skip。
该 Skip 是 PowerShell 下不适用的 Git 包装器取消用例，不是 Unix 通过证据。
使用固定 Go 1.26.5，并设置 GOENV=off、GOTOOLCHAIN=local、GOPROXY=off、GOSUMDB=off。
新机器须准备运行时与合法缓存；不要假定随项目附带，也不要把旧电脑 Codex 私有路径写入产品。
当前 Windows 验证器查找 USERPROFILE 下对应 Go toolchain 模块缓存，具体位置见脚本。

实施裁定：输出捕获只在内存，合计上限比旧逐流检查严格；额外固定读取缓冲区不计作
无限保留。为区分初始化耗时与输出失败，超限测试仅在私有脚本副本设置 30 秒预算，
外层测试 15 秒；生产 900 秒不变。测试版本注入不冒充实际其他版本运行。

## 下一步及必须保留的阻塞

1. 定位可用的原生 Unix 环境，补齐任务 1 的 Unix 实现和真实夹具回归，再按计划核对剩余合同。
   旧电脑未找到可用 WSL 分发或 Docker 命令；用户说 Docker 已安装，尚需实际路径，
   不能断言未安装。此次未启动容器或下载依赖。
2. 任务 1 真正完成后，才进行任务 2 工具模块与四入口消费者的原子迁移。
3. 任务 3 的生成、回归、物理验收及最终独立审查仍待完成。

完整普通测试曾在 prepared-exit-pending 的 CreateProcessW 边界失败（Win32 5），
同一未修改子用例单独通过不构成修复。Defender 对 C12 风格 bootstrap 的检测尚未解决，
没有足够关联证据认定具体子用例或确认误报。见
[基线诊断](2026-09-25-prepared-worker-baseline-diagnostics.md)。
用户没有微软账号，供应商提交暂缓；这不是安全放行。
不要为了继续开发重跑受阻 C12/bootstrap、完整普通套件或完整 C11；恢复这些执行需要
单独处理原阻塞并获得明确继续授权。I2/I3、物理验收阻塞持续保留，不修改安全设置规避检测。

## 历史资料与不上传内容

- [原生外层 Job 探针](2026-09-25-devtools-native-job-probe.md)：历史替代方案，未选为产品入口。
- [5.1 初始化边界](2026-09-25-devtools-powershell-initialization-boundary.md)：为何停止旧方案。
- [7.6.5 可行性](2026-09-25-devtools-powershell7-feasibility.md)：后续版本决策的局部证据。
- [Docker 验收诊断](2026-09-24-docker-acceptance-diagnostics.md)：历史验收背景。

忽略目录 .superpowers 中的执行流水、临时探针、可执行文件和原始日志不上传；
也不上传凭据、bootstrap 密钥、原始事件转储或 ETL。上述文档和正式测试保存可接续的信息，
但原始本机证据不会通过 Git 恢复；如需原件，应另行决定安全备份范围。
换机可据本说明重建本地执行流水，不需重复已批准设计，也不能把历史测试视为新机器验收。

本次仅按用户要求提交并推送进度分支；不合并 main，不声称产品工作完成。

## 2026-09-27 环境续查

用户提供 Docker Desktop 的实际安装目录后，只读检查确认 CLI 和 Linux 引擎均可用：
Docker Desktop 4.85.0、Engine 29.6.2、desktop-linux context、linux/amd64。
CLI 位于用户提供安装目录下的 resources/bin/docker.exe；不把这台电脑的路径固化为项目要求。
因此上文“尚未定位 Docker”的旧环境阻塞已解除。

现有 Docker 镜像列表没有 Go 工具链镜像；本机已检查的 Go toolchain 缓存仅有
1.26.5 windows-amd64，没有 Linux 版本。尚未启动容器、下载镜像或执行 Unix 测试。
下一步需单独批准联网准备固定版本的 Linux Go 1.26.5 测试镜像，或提供已有离线镜像。
网络准备与离线验收仍分开；Docker 可用不代表任务 1 或 C12 验收通过。

后续用户已批准下载，官方 golang:1.26.5-bookworm 拉取成功。固定镜像摘要：
`sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd`。
禁网、只读、无宿主挂载的临时容器实际报告 `go version go1.26.5 linux/amd64`、
Bash 5.2.15；timeout、mktemp、wc、rm、rmdir、setsid 均存在。
另一个禁网、只读临时容器仅挂载项目 scripts 目录，确认 verify-devtools.sh 的 Bash
语法检查通过，且缺少工具模块时退出 1、仅输出预期脱敏错误。两个容器均配置 --rm。
这两项是环境/前置检查，不是依赖完整性、输出限制、取消清理或任务 1 整体通过。
未运行 C12、完整 C11 或完整普通测试；未修改生产代码、未提交或推送本次环境记录。

随后继续实现，新增 Linux 进程测试并将真实缓存完整性夹具移植到原生 Unix。
Windows Go 平台误接受已取得 RED→GREEN；Linux 专项当前 14 PASS、4 FAIL，
强杀/正常退出后的子进程清理及运行中输出上限仍未满足。详细结果与范围见
[2026-09-27 Unix 边界报告](2026-09-27-devtools-unix-boundary.md)。
任务 1 仍未完成；监督方式涉及既有设计边界，不能因 Docker 已就绪而直接进入模块拆分。

用户随后批准隔离的 Unix 监督可行性探针。Linux 原生候选取得八项局部正向结果，
但负向对照确认监督程序自身被强杀仍会遗留子进程，不能直接晋升产品。
详见[监督探针报告](2026-09-27-unix-supervision-feasibility.md)。
正式入口和测试未因此再修改，原四项失败及所有验收阻塞继续保留。

## Windows-only 修订与旧实现归档

用户已确认 Windows-only 设计及修订实施计划。旧 Unix 实现、测试和脱敏报告已保存在
本地提交 `03947e31536a803aaca19c295edd8d9eb4d258f1`，明确标记未验收，尚未推送。
其历史结果为 14 PASS / 4 FAIL，不因后续拒绝入口而变成修复。归档时工作文件 SHA-256：

| 文件 | SHA-256 |
| --- | --- |
| scripts/verify-devtools.sh | 676F8C9B79F3DF74547AB1FF52134BD362F6730E9CF9094EED6089454FA44B12 |
| internal/e2e/devtools_verification_test.go | DF3BDAC535DF31625DF4B41149F685F310FAACEF02157202FA2849D2DA3B8D38 |
| internal/e2e/devtools_unix_test.go | 91A52F10FF827135B7124808FD1CBC2095D02CC273F465A1D7554E3AF33C608A |

活动 Unix 正向夹具已移除（可从上述提交恢复），verify-devtools.sh 改为所有平台只拒绝。
新增标准库独立拒绝测试，覆盖两种实际 Windows Git Bash 和 Linux Bash，每个解释器十项
场景，使用工作标记正向对照验证无工具启动；没有全局跳过 Bash，也未修改 smoke。

本轮取得新合同 RED→GREEN：Linux 旧入口执行 uname、提示不符，10 项失败；修改后
10 项通过、0 Skip，退出 0 / 0.153 秒。Windows 旧入口的 20 项拒绝测试因旧提示失败；
修改后包级专项退出 0 / 81.259 秒：PowerShell 29 项通过、Bash 20 项通过、0 失败、
1 项 Git 包装器专属不适用 Skip。新增合计输出超限、自然退出遗留子进程及非 Windows
平台判定覆盖首次即通过，属于既有行为覆盖，不称为缺陷修复。

非 Windows PowerShell 只做私有副本平台判定测试，不是实际 Unix PowerShell 验收。
MSYS/Cygwin 的两个常见安装位置未找到解释器，macOS 及这些环境未实测。
PowerShell 生产实现、900 秒/4 MiB、Bash smoke、C12 runner 均未修改。
完整普通测试、完整 C11、C12 物理 gate 仍受原安全事件阻塞，不运行、不声称通过。
本地保存任务 1 专项实现后继续任务 2 版本基线检查；尚未推送或合并 main。

任务 1 已提交为 `3a43a724`，提交后再次专项通过（81.078 秒，同为 49 PASS / 0 FAIL /
1 不适用 Skip），执行记录已标记任务 1 完成，但不是项目验收完成。
任务 2 模块图读取成功，完整模块版本查询因离线缺失失败，诊断列出 207 项；未开始模块迁移。
详见[版本基线阻塞及缺失清单](2026-09-27-devtools-module-baseline-blocker.md)。
等待单独批准联网准备锁定版本依赖或提供离线缓存；不自动下载、不改变防护或版本。

## 2026-09-28 最新接续点：Task 2 已提交，Task 3 未验收

Task 2 已在最终完整专项通过后原子提交为 `a0ca18928f552752c6580929f9b6fff88cd8d58b`；task-done 已记录（541.597 秒，1031 PASS 标记、0 FAIL、1 既有不适用 Skip）。未推送、未合并 main。

Task 3 根完整性通过，工具模块验证因缓存不齐失败；147 项缺解压目录，其中 27 项还缺归档和哈希。仅编译通过，不是测试执行。真实版本、生成和 lint 未运行。独立审查的缓存验证/执行绑定问题保持开放；恢复说明已补固定工具链缓存前提。下一步需批准安全相关绑定修复和精确缓存准备范围，不得自动下载、扩大预算或解除原安全门禁。完整命令结果、审查范围及147项清单见[真实验证检查点](2026-09-28-devtools-real-verification-checkpoint.md)。本检查点仍仅本机可用。

## 2026-09-28 接续记录：版本规则与组合验收历史

以上缓存阻塞属于历史状态。后续获批精确准备及 root 28 条件例外已接入；32 Required、35 精确条件候选、八组原包映射与独立证据绑定专项通过。两个模块正常离线整理后四锁不变，52 个受保护文件不变。根专属 lint 样例和错误目录反例通过，Linux 四入口拒绝通过。

最新 Windows 完整开发工具专项退出 1 /571.336 秒，唯一失败子用例为 verify-c11 取消测试的自有工具 12 秒就绪等待。补诊断后单独重跑通过（就绪10.214秒），但未确定原因，不认定修复或整体通过。详细可提交事实见[版本锁证据](2026-09-27-devtools-version-lock-evidence.md)，完整日志在本计划忽略目录。

Task 2 尚未原子提交；这些新增工作目前只在本机工作树，GitHub 尚不能恢复本轮进度。下一步先排查就绪延迟，再在原预算内复验；不得因单次重跑成功跳过组合失败。Task 3 真实工具、生成、lint、独立审查未完成；原 Defender、完整普通/C11/C12、I2/I3 与 main 门禁不变。

最新补充：已完成测试副本阶段计时，两轮三个消费者对照均通过。关闭插桩后，原完整 Windows 专项退出 0 /399.344 秒，1031 PASS 标记、0 FAIL、1 既有不适用 Skip；verify-c11 就绪6.514秒且取消/无关进程检查通过。原时限不变，未并行执行其他测试或依赖查询。历史超时未确定根因，不称为修复或冷态稳定；本次只有诊断与复验证据更新，未原子提交/task-done、未推送，真实工具验收及原安全门禁仍待处理。
