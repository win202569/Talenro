# C12 Development Tool Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将五个开发工具与 C12 主模块依赖分离，保留完整校验并重新取得真实验收证据。

**Architecture:** 根模块保留业务、测试、Goose 库及 CLI。独立 `tools/devtools` 模块保存五个开发工具，公开脚本先验证工具模块，再以显式 -modfile 调用工具。本轮四个工具入口仅支持 Windows 独立 PowerShell，以共享私有 Job 帮助文件覆盖整个入口；对应 `.sh` 在所有平台只执行拒绝。既有 Bash smoke 和 C12 runner 不修改。

**Tech Stack:** Go 1.26.0 / go1.26.5、Windows 上的 PowerShell Core 7.6.5 正式版、既有 Bash smoke、Go 脚本契约测试、Docker Desktop（Linux 容器仅用于新增拒绝测试，不代表 Unix 工具流水支持）。

**Spec:** `docs/superpowers/specs/2026-09-25-c12-development-tool-isolation-design.md`（用户已批准）。

**Approved addendum:** `docs/superpowers/specs/2026-09-25-c12-devtools-windows-powershell-addendum-design.md`（用户已确认；与原设计冲突时以附录为准）。

**Latest approved revision:** `docs/superpowers/specs/2026-09-27-c12-devtools-windows-only-design.md`（用户已书面确认；平台支持与测试迁移冲突以此为准）。

**Approved lock-version revision:** `docs/superpowers/specs/2026-09-27-c12-devtools-version-lock-revision-design.md`（用户已书面确认；仅锁版本差异处理以此为准）。

**Approved four-tools revision:** `docs/superpowers/specs/2026-09-28-c12-devtools-four-graph-exceptions-design.md`（用户已书面确认；仅四项 tools 图级候选条款优先）。

**Approved root28 revision:** `docs/superpowers/specs/2026-09-28-c12-devtools-root28-graph-exceptions-design.md`（用户已书面确认；仅新增 root 28 条件候选和相应上限优先）。

**本次计划审批状态：已确认。** 用户随后以“是”同时确认修订计划和 R3c-2 的 27 个精确版本准备范围；2026-09-28 已完成该范围准备，校验记录见版本证据文档。原 root 三项/tools 四项准备已完成，不重复。保留 Native inline，安全和推送/合并门禁不变。

**Revision status:** 用户已确认本次 Windows-only 计划修订；保留 Native inline。任务 1 已完成专项与本地提交，Windows 收尾复验 49 PASS / 0 FAIL / 1 不适用 Skip，Linux 拒绝 10 PASS。经授权准备依赖后，任务 2 已取得迁移前基线，模块及消费者实现处于未提交工作树；路由专项和 Linux 四入口拒绝已通过，但完整模块图存在未接受的版本差异，任务 2 尚未完成，任务 3 未开始。历史 Linux 14 PASS / 4 FAIL 不变。最新见 `docs/roadmap/2026-09-27-devtools-task2-checkpoint.md`；原 207 项阻塞记录保留历史。

## Global Constraints

- 当前工作目录为既有 `codex/c12-b01-task9-coordinator` worktree，基线文档提交 `ff302750`；不创建第二个 worktree，不切换 main。
- 新模块 `talenro.local/devtools`；不建立 go.work，不添加两模块间 require/replace。
- Buf 1.72.0、protoc-gen-go 1.36.11、oapi-codegen 2.8.0、sqlc 1.31.1、golangci-lint 2.12.2；Goose 3.27.1 留在根模块。
- Required 保留原 32 个优先值和 retained 义务：root 28/tools 4 必须存在，只允许原值或各自精确条件候选。新增 root 28 清单完全按最新设计第 3 节；独立字面量锁定，禁止现场快照生成期望。
- 原 root 三候选仍限定 github.com/chzyer/readline v1.5.1、github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b、go.opentelemetry.io/otel/metric/x v0.66.0，自然消失不引入；新增 root 28 后总上限 root 31/tools 4。无 latest、范围、新 replace/exclude 或清单外候选。
- tools 四候选仅为 golang.org/x/time v0.11.0、modernc.org/mathutil v1.6.0、modernc.org/sortutil v1.2.0、modernc.org/strutil v1.2.0；对应原值 v0.14.0/v1.7.1/v1.2.1/v1.2.1 仍优先。四节点不可缺失，不能用于 root；无第五项 tools 例外。root/Goose mathutil v1.7.1 不变。
- 不关闭防护、不添加排除项、不删除共享模块缓存或第三方 testdata。
- 新工具验证独立默认 900 秒；既有 C11 外层超时保持，嵌套可更早失败，不扩展外层预算。
- C12 原准备预算及各项 5m 测试预算、候选树、模块图、进程、清理保护一律不改。
- 不提交 `.superpowers`、ETL、诊断二进制、共享缓存、原始敏感输出。
- 不把实现完成、编译通过、fake harness 通过或工具验证通过当作物理验收。
- I2/I3 仍开放；任何阶段失败均保留原因，停止同条件重试；不合并 main。
- Windows 的 verify-devtools、check-tools、generate、verify-c11 仅支持独立 `pwsh.exe -NoProfile -File`，精确要求 PowerShell Core 7.6.5 正式版；公开入口不支持 dot-source。
- 5.1、其他版本和预发布版本在加载帮助文件、Add-Type 或启动工作前退出 1，提示 `<入口名>: PowerShell 7.6.5 required.`；不自动安装、升级、转调，不固定 Codex 私有路径。
- 四个工具入口之间只用当前已验证宿主的绝对路径启动子入口，不重新查 PATH。旧 smoke 和 C12 runner 宿主合同保持，不全局替换 powershell.exe。
- 四个 `.sh` 在所有平台只用 shell 内建命令输出 stderr 一行 `<入口名>: this release requires Windows and PowerShell 7.6.5.`，stdout 空，退出 1。无平台探测、模块解析、临时目录、外部程序或绕过参数，不自动转调其他入口。
- `.ps1` 在非 Windows 平台也提前拒绝；Linux、macOS、其他 Unix、WSL 本轮均不支持四入口正向流程。撤回 Unix 支持不改变其他组件部署平台，也不要求 Docker 使用 Windows 容器。
- 保留历史 Unix 失败对应的源码/测试提交与脱敏报告，不把新拒绝测试记为旧清理缺陷修复。不全局 Skip Bash 测试；探针仍只在忽略目录，不晋升产品。
- PowerShell C11 原有 Bash smoke 子步骤保留；不是 Windows Bash 工具入口支持。无新原生监督程序、编码加载器或 C12 帮助函数复用。
- Defender 事件未解除：暂缓供应商提交不等于确认误报。含受阻 bootstrap 的全套普通测试、完整 C11 和 C12 物理 gate 暂不运行；仅继续不触发该路径的独立工作。
- Job 只保证本机自有进程边界，不承诺强杀后删除 Docker 服务端资源；既有资源协议不变。

## Review Focus

1. 缓存完全缺失时 Go mod verify 可能成功：Task 1 必须先证明依赖齐备且离线缺失失败。
2. 路径含空格、错误/预发布 PowerShell、任意平台 Bash 误调用及 PATH 伪解释器：Task 1/2 验证无工作拒绝、同宿主分派及 Buf 参数边界，不遗漏保留的 Bash smoke。
3. 工具构建使用备用模块，lint 子进程误用工具模块：Task 2 用根模块专属包验证分析对象。
4. 拆分触发间接版本降低、图裁剪掩盖新增/删除、例外跨模块误用、进入实际包、伪造证据或重复包：Task 2 R2/R6 用缺失/额外/错误归属/错误版本/错误记录、实际包用途和唯一证据绑定反例及完整图核对；Task 3 真实生成比较。不得用 Windows 包集合相同替代全图规则。
5. 子进程自然退出与超时竞争、嵌套 Job、外层提前终止及脱敏：Task 1/2 验证整个 PowerShell 入口的自有进程退出、外部 canary 存活及初始化失败关闭。

## 文件职责与顺序

| 文件 | 职责 / 任务 |
| --- | --- |
| scripts/verify-devtools.ps1、scripts/verify-devtools.sh（新增） | Task 1，独立离线验证、期限和自有进程清理 |
| scripts/private/devtools-process.ps1（新增） | Task 1，函数定义式加载，私有 Job 初始化；Task 2 被另三个 PowerShell 入口复用 |
| internal/e2e/devtools_verification_test.go（新增，e2e tag） | Task 1，验证脚本的进程、缓存、错误与隐私测试 |
| internal/e2e/devtools_runtime_test.go、devtools_ownership_test.go | Task 1 已有运行时和真实 Job 边界测试，保留 |
| internal/e2e/devtools_bash_rejection_test.go（新增，e2e tag） | Task 1 独立标准库拒绝夹具；Task 2 复用测四入口 |
| internal/e2e/devtools_unix_test.go、当前 verify-devtools.sh 和共享测试 | Task 1 先保存旧合同源码/测试检查点，再移除活动 Unix 正向夹具及旧脚本执行体 |
| docs/roadmap/2026-09-27-devtools-unix-boundary.md、2026-09-27-unix-supervision-feasibility.md、2026-09-25-progress-recovery-handoff.md | Task 1 保存历史事实和接续状态，不提交原始日志/探针 |
| go.mod、go.sum；tools/devtools/go.mod、go.sum（新增） | Task 2，依赖边界与固定版本 |
| scripts/generate.ps1/.sh、scripts/check-tools.ps1/.sh、scripts/verify-c11.ps1/.sh | Task 2，入口验证及显式工具分派 |
| buf.gen.yaml | Task 2，protoc-gen-go 嵌套分派 |
| internal/e2e/privacy_test.go；internal/e2e/devtools_routing_test.go（新增，e2e tag） | Task 2，脚本契约、模块边界、版本和调用对象 |
| internal/e2e/devtools_version_lock_test.go（新增，e2e tag） | Task 2 R2/R6，标准库实现的版本策略比较与真实完整模块查询；不新增生产入口 |
| internal/e2e/devtools_version_evidence_test.go（已有未提交，e2e tag） | Task 2 R2/R6，证据解析绑定、owner 隔离和包用途拒绝及纯数据测试 |
| internal/e2e/testdata/devtools-version-lock/baseline.json、policy.json（新增） | Task 2 R1/R2，可提交的脱敏不可变基线、精确目标/例外和经审核归属决策 |
| docs/roadmap/2026-09-27-devtools-version-lock-evidence.md（新增） | Task 2 R1–R8，逐项引入链、恢复/例外必要性、证据哈希、查询结果和未验收事项 |
| README.md、docs/runbooks/repository-recovery.md、docs/roadmap/current-status.md | Task 3，恢复步骤、兼容性、实际验收记录 |

Task 1 不接入公开调用链，允许根目录尚无工具模块时独立入口明确失败。
Task 2 原子提交模块拆分和所有调用点，不留下根工具已删除但脚本仍调用旧路径的提交。
Task 3 执行真实验证与文档交接。每任务自带测试周期和独立可审查结果。

## Task 1: 新增独立开发工具验证入口

**Files:** 完成已有 Windows verify-devtools.ps1 和私有帮助文件；修改 verify-devtools.sh 为拒绝入口；保留 devtools_verification/runtime/ownership 测试，新增 devtools_bash_rejection_test.go；先归档后移除活动 devtools_unix_test.go。不覆盖既有诊断或忽略目录探针。

**Interfaces:**
- 唯一正向入口：Windows `pwsh.exe -NoProfile -File scripts/verify-devtools.ps1`（Core 7.6.5 正式版）。`bash scripts/verify-devtools.sh` 所有平台只拒绝；无公开跳过、模块路径、命令或超时覆盖参数。
- 私有帮助文件只定义 `Initialize-DevtoolsProcessOwnership`（无参数、无成功输出，失败抛异常）；入口捕获后按自身脱敏格式退出 1。Job 句柄不返回给消费者，保持到进程退出；不提供释放接口。
- 受支持验证路径成功输出仅 `verify-devtools: passed`，退出 0；失败仅输出 `verify-devtools: <stage> failed with exit code <n>.`，不回显底层输出。PowerShell 版本/平台及全平台 Bash 拒绝提示另按本文固定格式输出。
- 新测试接口 `testDevtoolsBashEntryRejection(t *testing.T, entry string)` 在 devtools_bash_rejection_test.go；仅依赖标准库，entry 是仓库固定 `.sh` 文件名。`TestDevtoolsVerificationBash` 在该文件调用它测试 verify-devtools.sh；Task 2 的四入口测试复用它，不依赖 PowerShell fixture 或其他测试文件。
- Go 非零退出原样传播；校验语义错误/清理失败退出 1；命令缺失 127；超时 124。
- 不启动生成、lint、Docker；不得 dot-source 或执行 C12 runner 来借用函数。

- [x] **0a. 写版本边界 RED。** 在 `TestDevtoolsVerificationPowerShell` 增加 `runtime-5.1-rejected`、
  `runtime-prerelease-rejected`、`runtime-other-version-rejected`、`runtime-7.6.5-accepted`。
  实际 5.1 使用系统 powershell.exe，成功路径用已解析的实际 7.6.5；两者在 fixture 中运行同一
  真实入口，缺少 7.6.5 则明确记录环境阻塞，绝不回退 5.1。版本拒绝断言完整输出和退出码：

```go
if exitCode != 1 || strings.TrimSpace(output) != "verify-devtools: PowerShell 7.6.5 required." {
    t.Fatalf("runtime rejection mismatch: exit=%d output=%q", exitCode, output)
}
```

  fixture 帮助文件只写入专属事件文件；错误版本必须在加载前拒绝，事件文件不得存在。
  7.6.5 正向分支使用真实帮助文件，不用空函数替代 Job。其他版本/预发布判定可在测试副本
  的运行时信息读取处注入明确值，必须断言恰好一次替换；标为判定逻辑测试，不算实际版本验收。
  拒绝分支配套自身进程启动事件观察，确认没有 csc/Go/工作子进程；观察器需独立正向对照。

- [x] **0b. 运行版本 RED，再实现入口最前部检查。**

```powershell
go test -tags=e2e ./internal/e2e -run '^TestDevtoolsVerificationPowerShell$/^runtime-' -count=1 -timeout=5m
```

  历史 Expected RED（此步已完成，不回退产品重新制造失败）：5.1 曾进入旧验证流程而非版本拒绝。
  恢复时核对已有证据并运行现有 GREEN；不能把测试初始化错误当 RED。
  在公开入口自身、加载任何帮助文件前写入下列 5.1 可解析语法，不使用需要子进程的版本命令：

```powershell
$devtoolsRuntime = $PSVersionTable
if ($devtoolsRuntime.PSEdition -ne 'Core' -or
    [string]$devtoolsRuntime.PSVersion -cne '7.6.5' -or
    $devtoolsRuntime.ContainsKey('PSVersionPreReleaseLabel') -or
    [Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
  [Console]::Error.WriteLine('verify-devtools: PowerShell 7.6.5 required.')
  exit 1
}
```

  重跑同一命令，Expected: 版本判定用例 PASS；正向完整性/进程测试尚需后续步骤，不能标记任务完成。

- [x] **0c. 冻结旧 Unix 合同证据。** 在改动旧执行体或删除活动夹具前，检查下列明确文件，
  将现有源码、测试与脱敏报告提交为带未验收说明的历史检查点；不得把红色专项称为通过。
  不重跑已知四项同原因失败、不提交原始日志或 `.superpowers` 探针。

```text
git add scripts/verify-devtools.sh internal/e2e/devtools_verification_test.go internal/e2e/devtools_unix_test.go docs/roadmap/2026-09-27-devtools-unix-boundary.md docs/roadmap/2026-09-27-unix-supervision-feasibility.md docs/roadmap/2026-09-25-progress-recovery-handoff.md
git diff --cached --check
git commit -m "wip(tooling): archive unaccepted Unix verification evidence"
```

  Expected: 提交仅包含列明文件，保存此前 14 PASS / 4 FAIL 的实现和测试；没有探针/二进制。
  记录该提交号及源码 SHA-256 到接续文档，作为后续移除活动 Unix 正向测试的可恢复来源。

- [x] **1. 补 Windows 合同缺口与全平台拒绝 RED。** 文件使用 `//go:build e2e` 和 `package e2e`。
  用 t.TempDir 创建带空格的镜像仓库，复制待测脚本，提供固定的 tools/devtools/go.mod，
  从现有 privacy_test.go 的假工具及进程测试模式建立本任务专属 fixture。
  已有脚本不得再用“入口不存在”作 RED 证据；新增断言必须失败于尚未实现的合同。测试表固定为：

```go
cases := []struct{name string; want int}{
    {"valid", 0}, {"missing-module", 1}, {"missing-cache", 1},
    {"tampered-directory", 1}, {"tampered-zip", 1},
    {"go-exit-7", 7}, {"wrong-version", 1}, {"timeout", 124},
    {"timeout-with-child", 124}, {"output-canary", 7},
    {"inherited-module-override", 1},
    {"missing-ziphash", 1}, {"output-limit", 1},
    {"ownership-init-failure", 1},
}
```

  测试名使用 `TestDevtoolsVerificationPowerShell`、`TestDevtoolsVerificationBash`。
  缺失和篡改测试用 fixture 独占缓存，绝不修改用户缓存。
  为真实完整性分支建立一个小型本地模块代理 fixture：go.mod、版本 .info、zip、list，
  准备阶段使用本地 file 代理下载到 fixture 缓存，再切换离线运行，逐个损坏 zip/源码。
  checksum 缺失、zip 和目录全部缺失必须失败；成功 fixture 运行真实 go mod verify。
  fake Go 只模拟进程/错误/环境合同，不作为真实哈希验证证据。

  补充 Windows `natural-exit-with-child`：fake 工具启动长存子进程，测试先取得双方句柄，
  再经专属放行文件允许工具正常退出；断言入口成功、子进程有界退出、外部 canary 存活。
  `combined-output-limit`：stdout 与 stderr 各输出 3 MiB 后保持运行，断言退出 1 而非 124，
  自有进程退出、无原始输出泄露；验证合计上限，不只测单流超过 4 MiB。
  已实现行为若首次测试即通过，记录为新增覆盖，不伪造 RED；发现缺陷才按真实 RED 修复。
  `runtime-non-windows-rejected` 在专属 `.ps1` 测试副本中恰好一次替换平台读取为 Unix 值，
  保留实际 7.6.5 宿主，断言版本/平台拒绝、帮助文件工作标记不存在。此为判定逻辑测试；
  没有实际 Unix PowerShell 时记未实测，不下载运行时或把副本当真实平台证据。

  将 `TestDevtoolsVerificationBash` 移到新的 devtools_bash_rejection_test.go，并用同文件的
  `testDevtoolsBashEntryRejection(t, "verify-devtools.sh")` 在 Windows 与 Linux 执行同一拒绝合同。
  不再在非 Windows 调用旧完整性/进程正向夹具；其源码已由 Step 0c 归档，移除活动
  devtools_unix_test.go 及仅服务于旧 Unix 正向路径的分派，保留 Windows 完整性和进程测试。
  此为合同迁移，不声称原取消或输出保护已修复。

  拒绝夹具使用实际解释器绝对路径：Windows 分别为 Git bin 包装器和 usr 实际 Bash；Linux
  使用已解析的 bash。复制未经插入工作标记的真实入口到带空格目录；stdout/stderr 分开捕获。
  专属 PATH 前置 Go、Git、Docker、pwsh/powershell、uname、mktemp 的事件写入器；临时私有
  帮助脚本同样写事件，以正向控制证明事件夹具有效，再清空本次事件记录后启动待测入口。
  不继续依赖旧 `stage=module` 字符串恰好出现一次的插桩。
  测试表：缺模块、完整模块、只读模块、仓库外 cwd、额外参数、GOFLAGS/-modfile 与
  GOCACHEPROG 污染、假的 OSTYPE 值、仅标记工具的 PATH、空 PATH。空 PATH 用实际解释器
  绝对路径启动；清除解释器启动文件变量 BASH_ENV/ENV，测试入口本身而非任意用户启动脚本。
  不伪造 uname 结果来声称另一个 OS 通过。下面的 exitCode/stdout/stderr 来自实际进程结果：

```go
want := strings.TrimSuffix(entry, ".sh") + ": this release requires Windows and PowerShell 7.6.5."
if exitCode != 1 || stdout != "" || strings.TrimSpace(stderr) != want || strings.Count(strings.TrimSpace(stderr), "\n") != 0 {
    t.Fatalf("rejection mismatch: code=%d stdout=%q stderr=%q", exitCode, stdout, stderr)
}
if _, err := os.Stat(eventPath); !os.IsNotExist(err) {
    t.Fatalf("unsupported entry launched work: %v", err)
}
```

- [x] **2. 跑 RED。**

```powershell
$env:GOOS='windows'
go test -tags=e2e ./internal/e2e -run '^TestDevtoolsVerification(PowerShell|Bash)$' -count=1 -timeout=5m
```

  Expected RED: 旧 Windows 文案不符合统一提示，Linux 仍进入模块阶段/工作流程，因此拒绝
  断言失败；编译、环境初始化失败不是 RED。既有 Windows 正向测试预期继续通过。
  Windows 两个实际解释器必须覆盖；可用的 MSYS/Cygwin 也执行，缺失则记录未实测。
  Linux 只跑新的独立拒绝文件，不编译整套 e2e 或执行旧 Unix 正向流程：

```powershell
# 在 worktree 根目录运行；使用上一轮已准备的实际 Docker 路径，未找到则报告环境阻塞。
$dockerExe = Join-Path $env:LOCALAPPDATA 'Programs/DockerDesktop/resources/bin/docker.exe'
$repoPath = (Get-Location).Path
& $dockerExe run --rm --init --pull=never --network=none --read-only --cap-drop=ALL --security-opt=no-new-privileges --pids-limit=128 --memory=1g --cpus=2 --tmpfs /tmp:rw,exec,size=536870912 --mount "type=bind,source=$repoPath,target=/repo,readonly" --workdir /repo/internal/e2e --env GOCACHE=/tmp/go-build --env GOTOOLCHAIN=local --env GOENV=off --env GOWORK=off --env GOPROXY=off --env GOSUMDB=off --entrypoint /usr/local/go/bin/go golang@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd test -v -tags=e2e devtools_bash_rejection_test.go -run '^TestDevtoolsVerificationBash$' -count=1 -timeout=3m
```

  Expected RED: 同一拒绝断言失败；GREEN 时退出 0，实际 Linux 拒绝子用例 PASS、无 Skip。
  该命令仅是 Linux 拒绝测试，不替代 Windows 包级专项。无网络拉取、socket 或共享缓存挂载。
  Go 测试固定 GOTOOLCHAIN=local、GOENV=off、GOPROXY=off、GOSUMDB=off。本机固定工具链为
  `C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe`；
  仅在已确认的沙箱标准库访问问题出现时申请限定测试命令提权，不改全局配置。
  Go fixture 启动正式验证器时改用实际 7.6.5 的已解析绝对路径；不得在测试代码中硬编码
  `C:/Users/Lenovo/.cache/...`。缺少该版本记录未验收，不通过伪解释器模拟实际 Job 证据。

- [x] **3. 完成 Windows 验证命令链。** 完成平台/版本拒绝和 PowerShell 归属初始化后，以脚本位置解析根目录及工具模块，先检查两个锁文件存在。
  固定环境使用进程级设置且 finally 恢复：GOWORK=off、GOENV=off、GOTOOLCHAIN=local、
  GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly；
  继承的 GOFLAGS 中若含 -modfile/-overlay，先拒绝并退出 1；其余继承 GOFLAGS、
  GOEXPERIMENT、GOCACHEPROG 和模块/工具链覆盖项清除后使用上述固定设置。
  Windows 使用既有固定工具链目录，但不加载 C12 runner；必须精确匹配
  `go version go1.26.5 windows/amd64`。Bash 不再执行任何验证命令。
  所有验证命令在 tools/devtools 下运行：

```text
go version
go list -m -f '{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}' all
go mod verify
```

  依赖齐备规则：排除 go/toolchain 虚拟记录，其余选中模块必须有非空版本、存在的目录及
  对应缓存 zip、ziphash；路径须位于本次固定模块缓存下，拒绝本地替换和越界路径。
  缓存键按 Go module 路径大小写转义规则构造（大写 A -> !a），须测混合大小写路径。
  先检查齐备，再完整 verify；最终要求退出 0 且标准输出严格为 `all modules verified`。
  不运行 download；需要准备依赖时停止并提示阶段失败。

- [x] **4a. 抽取私有 Job 初始化并先测试边界。** 从现有验证器抽取
  `Initialize-DevtoolsProcessOwnership` 到 `scripts/private/devtools-process.ps1`；帮助文件加载
  不启动工作、不自动初始化。保留未命名 Job、非继承句柄、kill-on-close、当前进程先入 Job
  的顺序，初始化失败在 Go 启动前返回脱敏失败。入口在受控 catch 中调用：

```powershell
. (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
Initialize-DevtoolsProcessOwnership
# 初始化成功后才允许启动验证命令；不提前关闭 Job 句柄。
```

  在 fixture 私有帮助文件副本中把初始化替换成 `throw 'OWNERSHIP_PRIVATE_CANARY'`，断言
  退出 1、无工具事件、输出不含 canary；这只测试失败传播，不代替真实 Win32 初始化测试。
  用真实父 Job 启动验证器证明嵌套可用；测试保留自有进程句柄，强杀入口后有界确认子孙退出，
  独立外部 canary 存活。保留 WaitDelay 防止管道拖到子进程自然结束造成假通过。
  在 7.6.5 下测试明文 Add-Type 初始化，不能复制探针为正式实现。检查初始化互操作加载是否自身启动编译器：附录要求先有归属再有子进程，不能将 Add-Type
  的潜在编译子进程排除在外。若当前加载方案不能满足此顺序，停止并报告设计约束，不自行
  引入探针监督程序、编码加载器或降低合同。
  观察器订阅仅筛选被测进程的子进程启动事件，先订阅再初始化，初始化完成后启动专属正向
  对照，确认它被观察到；事件订阅在 finally 清理。记录真实运行时版本和初始化窗口事件，
  无观察权限应报告证据缺失，不能按“零事件”通过。本机探针的成功记录不能替代这一正式测试。

- [x] **4b. 实现进程预算与输出保护。** 单个入口从启动起只用一个 900 秒绝对截止时间，
  version、list、verify 共用剩余预算，不给每条命令另加 900 秒。
  PowerShell 以现有 verify-c11.ps1 的 `ConvertTo-C11CommandLineArgument`、
  `Get-C11CompletedProcessExitCode`、`Stop-C11OwnedProcessTree`、`Invoke-C11External`
  为已知模式，在新脚本内部建立私有实现；不改旧 C11 函数、不引入系统全局清理。
  沿用已实现的 .NET ProcessStartInfo 独立 ArgumentList、UseShellExecute=false、CreateNoWindow=true，
  保留进程句柄；超时停止自有进程树。若使用 Start-Process 必须 Hidden，不引入可执行字符串拼接。
  Windows 入口提前终止时不得留下自有 Go/子进程；不能证明则验收不通过。
  不新增 Unix 监督程序；四个 Bash 入口不启动工作，无需建立进程清理拓扑。

```powershell
# 每次启动前从同一个截止时间计算；耗尽即失败，不自动续期。
$remaining = [Math]::Floor(($deadline - [DateTime]::UtcNow).TotalSeconds)
if ($remaining -le 0) { throw 'verify-devtools deadline exhausted' }
```

  沿用已通过专项的内存有界捕获，不恢复旧落盘 sink。stdout/stderr 合计保留上限 4 MiB，
  固定读取缓冲区开销另计；只解析需要的数据，不打印原始错误，超限在命令仍运行时失败。
  以持续写超限输出后阻塞的 fixture 断言提前失败、进程退出及不泄露内容；清理失败不能被成功状态覆盖。
  保留私有测试副本的单次精确期限替换与独立外层保护，公开入口固定 900 秒，不增加参数。

- [x] **4c. 实现独立 Bash 拒绝入口。** 将 verify-devtools.sh 替换为以下完整最小入口，
  不在它后面保留旧模块校验执行体，不调用 uname 或其他外部程序；旧源码在 Step 0c 提交中。

```bash
#!/usr/bin/env bash
printf '%s\n' 'verify-devtools: this release requires Windows and PowerShell 7.6.5.' >&2
exit 1
```

  其他三个 `.sh` 在 Task 2 与消费者迁移一起替换，当前不提前改变它们。

- [x] **5. GREEN 与独立负向确认。** 运行 Step 2，增加超时边界前自然退出、子进程存活 canary、
  父进程取消、stderr 敏感 canary、目录含空格和调用后环境不泄漏断言。
  移除齐备检查的测试副本必须使 missing-cache 负向测试失败，证明不会受 Go 的跳过语义欺骗。
  只在 fixture 中做变异，不修改真实脚本来求通过。
  重跑 Step 2 的 Windows 包级选择器与 Linux 单文件拒绝测试。
  Expected GREEN: 两条命令退出 0；Windows 正向、两个 Git Bash 拒绝路径及 Linux 拒绝通过。
  对 Windows Git 包装器专属不适用用例仍独立记录 Skip，不能把它或未知平台计为成功。
  接续文档写明历史 Linux 四项失败已从本轮支持合同撤回、并未修复，附 Step 0c 提交号。
  逐项核对任务合同后才记录 Task 1 完成；历史通过次数不替代本轮证据。
- [x] **6. 提交。**

```text
git add scripts/verify-devtools.ps1 scripts/verify-devtools.sh scripts/private/devtools-process.ps1 internal/e2e/devtools_verification_test.go internal/e2e/devtools_runtime_test.go internal/e2e/devtools_ownership_test.go internal/e2e/devtools_bash_rejection_test.go internal/e2e/devtools_unix_test.go docs/roadmap/2026-09-25-progress-recovery-handoff.md
git commit -m "feat(tooling): add fail-closed offline development tool verification"
```

## Task 2: 原子拆分模块并更新全部调用链

**Files:** 文件映射中的模块文件、六份入口脚本、buf.gen.yaml、privacy_test.go、devtools_routing_test.go、Task 1 的 devtools_bash_rejection_test.go。

**Interfaces:** 消费 Task 1 的无参公开入口；新工具模块锁文件和显式 -modfile 调用是唯一新增分派合同。
  Goose 裸 go tool 命令保持。子工具和 lint 仍以根工作目录运行。
  三个 PowerShell 消费者还调用 Task 1 的 `Initialize-DevtoolsProcessOwnership`；公开验证器必须
  通过当前已验证宿主的绝对路径启动独立 pwsh.exe 子进程，不使用 `& <verify-devtools.ps1>` 在同一进程运行第二次初始化。
  `.sh` 仅复用 Task 1 的拒绝合同，不消费工具模块或参与正向路由。

### 当前接续点与锁版本修订步骤

Task 1 已完成，不能为制造 RED 回退它。Task 2 的基线、脚本实现及原专项已存在：256 PASS / 0 FAIL / 1 不适用 Skip、577.969 秒；Linux 50 PASS、0.960 秒。这是历史证据，不勾选版本迁移完成。
R1 基线已归档；R2 新纯规则历史 76 PASS（含顶层）/0 FAIL，但真实 Lock 接线未完成。R3a/root 三项和 R3b/tools 四项已准备完成。tools 严格模块/图查询已成功；root cel pin 被 tidy 删除，未重复；root 28 why 成功，严格版本批次因缓存缺失失败。接续核对 R1 后执行新版 R2、获准的 R3c、R4–R8；不重建模块/基线，不重做 Task 1，不将历史绿色当验收。

**接口与文件（均为测试/证据，不供生产脚本调用）：**
- 在既有 `internal/e2e/devtools_version_lock_test.go` 中保留 `devtoolsVersionRecord`、`devtoolsVersionPolicy`、`compareDevtoolsVersions(snapshot devtoolsVersionSnapshot, policy devtoolsVersionPolicy) []string`、`readDevtoolsVersionSnapshot(t *testing.T, owner string) devtoolsVersionSnapshot`。Owner 仅 root/tools；Error 在过滤 main/go/toolchain 虚拟节点前检查。
- `Required` 仍保存原 32 优先值；root 28/tools 4 必须存在且 retained，取原值或满足证据的本归属候选。`Exceptions` 精确为 root 31/tools 4。`Membership` 人工审阅，`Evidence` 为 owner/path → ID。只修改 policy.json，不改 schema 1 或 baseline.json 的 666 模块、八组映射、54 原哈希。
- 新增 `devtoolsPackageRecord{ImportPath string; Module *devtoolsVersionRecord; Incomplete bool; Error json.RawMessage; DepsErrors []json.RawMessage}`，`devtoolsVersionSnapshot` 增加 `Packages map[string][]devtoolsPackageRecord`。tools 必须有 buf/protoc/oapi/sqlc/lint 五个键及每组完整非空查询结果，root 有 root/integration/goose 三组。缺键、空结果、重复 ImportPath、Error/Incomplete/DepsErrors 均失败；合成正向 fixture 使用至少一个无关标准库包，不能用空切片冒充成功查询。
- 在 `internal/e2e/devtools_version_evidence_test.go` 保留 `compareDevtoolsPackageUse(snapshot devtoolsVersionSnapshot, policy devtoolsVersionPolicy) []string` 及已有包/证据类型和函数；新增严格读取 `readDevtoolsPackageSnapshot(t *testing.T, phase string) []devtoolsPackageRecord` 到版本 Lock 测试文件。用途规则：root 三组禁入新增 28 路径，tools 五组禁入原四路径，均不论原值或候选。其他归属的合法包不误拒绝。R6 比较完整映射；查询固定环境/90 秒。
- 新增 `devtoolsEvidenceRecord{ID, Owner, Path, Version, Kind, Sum, GoModSum string}`；`parseDevtoolsEvidence(report []byte) ([]devtoolsEvidenceRecord, error)` 只解析证据文档逐行 `<!-- devtools-evidence {JSON} -->` 标记，JSON 使用上述字段名。拒绝格式错误、未知字段、重复 ID；Kind 仅 exception/removal。exception 绑定实际选中版本、非空 Sum/GoModSum；removal 绑定原基线版本。记录后的人类可读段落载明来源链/审核结论，机器标记不替代内容审查。
- 新增 `validateDevtoolsEvidence(snapshot devtoolsVersionSnapshot, policy devtoolsVersionPolicy, records []devtoolsEvidenceRecord) []string`：对实际使用例外和 removed 条目，检查 ID 非空、唯一、存在且 owner/path/version/kind 全匹配，错误则拒绝。`TestDevtoolsVersionLock` 必须同时通过版本比较、证据验证、包用途与原包基线比较，不能只调用版本比较器宣称条件例外合格。
- 测试内独立字面量保护原 Required 32 和 Exceptions 35（root 31/tools 4）；新增 `const devtoolsRootGraphCandidates` 逐行保存最新设计第 3 节的 `root path candidate` 28 三元组，再与原七项组成 `devtoolsAllowedVersions`。不得取 policy 或现场结果作期望。只用标准库，证据写既有 roadmap evidence 文件。

- [x] **R1. 核对已归档基线，不重建。** 核验 baseline.json 的 666 条模块、八组 516/523/656/858/137/279/575/1187 个包和 54 个原哈希；对照既有成功原始记录与恢复前四锁哈希，记录新的接续 HEAD 和四锁哈希，不覆盖原记录。
  Expected：基线身份/版本和原文件哈希不变，无本机绝对路径或凭据；历史归档不是本次真实验收。

- [x] **R2a. 扩展根 28 的 RED。** 在 `TestDevtoolsVersionPolicy/root-28-candidates` 用独立 28 字面清单逐项测试原值/精确候选正例、错误版本、移用于 tools、目标缺失/removed、未审核归属、缺证据反例；保留原七项及 35 回放。`TestDevtoolsVersionEvidence/root-28-bindings` 逐项覆盖正确绑定、错误 owner/path/version/kind、缺 ID/校验和；增加标记大小写错误字段、重复 JSON 字段、未知字段、尾随 JSON、重复 ID 拒绝，沿用精确字段名合同。`TestDevtoolsVersionPackageUse/root-28-phases` 对三组 × 28 路径 × 原/候选版本逐一拒绝；tools 使用本表路径、root 使用 mathutil v1.7.1 为正对照。第五 tools 或本表外第 29 新 root 路径拒绝。保留八组空/缺组/错误/重复反例。
  合成用例须逐项可判读，例如：
```go
// 正常 tools 组均含至少一个标准库包；添加目标包后必须失败。
s.Packages["sqlc"] = append(s.Packages["sqlc"],
    devtoolsPackageRecord{ImportPath: "modernc.org/mathutil",
        Module: &devtoolsVersionRecord{Path: "modernc.org/mathutil", Version: "v1.6.0"}})
if len(compareDevtoolsPackageUse(s, p)) == 0 { t.Fatal("graph-only exception became a used package") }
// 同一 ID 指向错误归属记录，不能只因 ID 非空而通过。
records[0].Owner = "root"
if len(validateDevtoolsEvidence(s, p, records)) == 0 { t.Fatal("wrong-owner evidence accepted") }
```
  新根边界用例示意（fixture 内其余阶段必须非空）：
```go
s := devtoolsCleanPackageFixture("root")
s.Packages["goose"] = append(s.Packages["goose"], devtoolsPackageRecord{
    ImportPath: "cel.dev/expr", Module: &devtoolsVersionRecord{Path: "cel.dev/expr", Version: "v0.25.1"},
})
if len(compareDevtoolsPackageUse(s, p)) == 0 { t.Fatal("root graph-only module became used") }
```
  TestDevtoolsVersionLock 保留两模块真实查询，补齐 module identity、Exclude、跨模块 require、Go/toolchain、工具表。历史 35 回放缺证据时仍拒绝，条件齐备才允许，不能整体删除回放。

- [x] **R2b. 运行 RED，完成测试辅助检查器，再跑 GREEN。**
```powershell
go test -v -tags=e2e ./internal/e2e -run '^TestDevtoolsVersion(Policy|Evidence|PackageUse)$' -count=1 -timeout=5m
```
  Expected RED：新增根候选正例和根用途/严格证据反例出现语义失败；不回退已实现功能来制造 RED，编译失败不算。随后在 `compareDevtoolsVersions` 的 Required 分支允许本归属精确候选，仍独立验证存在、retained 和证据；更新 policy 与独立字面量 35 例外。扩展上述证据/用途检查器并补严格读取/Lock 联合验证。重跑同选择器 Expected GREEN；真实 Lock 未补证前保持失败，Membership 保持 pending 直至 R5。
  旧 19 场景绿色不能覆盖新条件。真实 Lock 可在 R6 前保持阻塞；未完成证据的候选绝不记录验收通过。

- [x] **R3a. 保留原三个 root 准备记录。** readline v1.5.1、demangle v0.0.0-20250417193237-f615e6bd150b、metric/x v0.66.0 已获授权并准备成功；核对既有 Sum/GoModSum 与结果，不无故重复下载，也不视为已证明必要。

- [x] **R3b. 四个 tools 候选准备（历史已授权且已完成，不重复）。** 下列保留原准备范围与过程作为记录：
```text
golang.org/x/time@v0.11.0
modernc.org/mathutil@v1.6.0
modernc.org/sortutil@v1.2.0
modernc.org/strutil@v1.2.0
```
  固定 Go、独立最小备用模块、逐个显式 module@version 的 mod download -json；仅 https://proxy.golang.org 和 sum.golang.org。GOENV/GOWORK=off、GOTOOLCHAIN=local、GOAUTH=off、GOVCS=all:off，清除 GOPRIVATE/GONOSUMDB/GONOPROXY/GOINSECURE 及继承覆盖。整批总预算 900 秒，Hidden/无窗口，保留进程句柄，超时只停止自有进程。请求和结果写忽略目录，脱敏精确版本/校验和/来源/起止时间/退出码写证据文档。
  Expected：每个请求对应唯一精确记录，无 Error，Sum/GoModSum 齐备；不得执行下载内容或改生产锁文件。额外版本/传递准备需求停止报告，不回退 direct/VCS/auth。若只确认计划而未批准本段，缺缓存时停止询问。随后恢复禁网，不把 root 三候选或旧 207 项授权扩展到本段。

- [x] **R3c-1. 接续时只读复核缓存清单。** 逐项使用最新设计第 3 节的精确 path/version，按 Go 的大写字母 ! 小写转义定位缓存 .mod/.info/.zip/.ziphash，并核对已提取源码目录供后续审查；不得删除或改写缓存。2026-09-28 规划时四文件检查仅 perfstat@v0.0.0-20210106213030-5aafc221ea8c 齐备，其余 27 项存在缺失；文件存在不证明完整性。记录实际缺失文件类型到忽略目录，在可提交证据文档保存脱敏清单。
  Expected：28 项均有明确逐文件状态；复核只能缩小下列待授权下载集合，不能增加 path/version。perfstat 此轮仅离线读取既有校验和/源码并在完整性阶段验证，若发现仍需联网则另报，不把它隐含加入。

- [x] **R3c-2. 新增根候选精确准备（已单独授权并完成）。** 本段授权对象只包含以下 27 个精确版本的缺失模块元数据、归档/校验和及正常解压源码（逐项显式 mod download -json；不执行内容）。R3c-1 复核后已齐备项不下载：
```text
cel.dev/expr@v0.25.1
cloud.google.com/go@v0.34.0
github.com/Azure/go-ansiterm@v0.0.0-20210617225240-d185dfc1b5a1
github.com/BurntSushi/toml@v1.3.2
github.com/alecthomas/units@v0.0.0-20211218093645-b94a6e3cc137
github.com/ebitengine/purego@v0.8.4
github.com/gorilla/websocket@v1.4.2
github.com/hashicorp/go-version@v1.8.0
github.com/mattn/go-colorable@v0.1.14
github.com/moby/moby/api@v1.54.2
github.com/moby/moby/client@v0.4.1
github.com/moby/term@v0.5.0
github.com/pelletier/go-toml/v2@v2.2.2
github.com/shirou/gopsutil/v4@v4.25.6
github.com/sirupsen/logrus@v1.9.3
github.com/tklauser/go-sysconf@v0.3.12
github.com/tklauser/numcpus@v0.6.1
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.68.0
go.uber.org/zap@v1.27.1
golang.org/x/lint@v0.0.0-20190313153728-d0100b6bd8b3
golang.org/x/oauth2@v0.34.0
golang.org/x/xerrors@v0.0.0-20200804184101-5ec99f83aff1
google.golang.org/appengine@v1.4.0
google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013
google.golang.org/genproto/googleapis/api@v0.0.0-20260120221211-b8f7ae30c516
gopkg.in/yaml.v2@v2.2.3
honnef.co/go/tools@v0.0.0-20190523083050-ea95bdfd59fc
```
  固定 Go、独立最小备用模块、逐项显式 module@version；仅 https://proxy.golang.org 与 sum.golang.org。清除继承 GO* 覆盖后设 GOENV/GOWORK=off、GOTOOLCHAIN=local、GOAUTH=off、GOVCS=all:off，GOPRIVATE/GONOPROXY/GONOSUMDB/GOINSECURE 为空。整批共用 900 秒绝对期限，不分项重置；Hidden/保留自有进程句柄，超时仅清理本次自有进程。结果精确绑定 Path/Version、无 Error、Sum/GoModSum 齐全；原四锁哈希前后相同，恢复禁网。日志保留起止/退出码，脱敏记录写既有 evidence 文档。
  Expected：只准备批准且缺失的精确版本，校验来源可追溯；如 Go 请求清单外版本/传递准备，或权限未明确批准，停止报告，不扩范围。准备成功不证明源码无风险、例外必要或验收通过。计划确认和本段授权分别记录；仅确认计划不包含本段下载许可。

- [x] **R4. 核对实际精确候选及活动来源。** 在获准准备后读取两模块严格完整图/版本；32 Required 均存在且仅为原值或本归属精确候选。root 28 逐项核对较高版本来源、当前声明/选中值、TestImports/XTestImports、平台条件；22 直接边与 6 历史链按评估清单分别追踪，旧父版本不能冒充选中值。cloud/genproto/api/rpc 联合检查重复包。不重复已失败 pin、不为其余 27 项制造失败实验；若有具体稳定来源可保留原值，禁止引入新上游或工具回流。
  每组变更后正常离线 tidy、严格完整图查询，各命令仍为 90 秒。tidy 必须在所属模块目录运行：root 为仓库根，tools 为 tools/devtools，不从 root 用 tools modfile 整理根源码。记录 before/target/after、父边、cwd 和锁文件哈希。
  Expected：32 目标 retained、实际值在精确允许范围，活动链可解释；候选仍须 R5/R6 补证，不因严格查询成功就通过。未知差异、重复包、目标缺失、用途或平台影响无法解释即阻塞，不动 doubleclick/replace/exclude。

- [x] **R5. 审核最多 35 个条件例外及全部成员归属。** 原 root 三项/tools 四项沿用各自必要性合同；新增 root 28 按最新设计第 4 节逐项记录来源、用途、选择理由及实际做过的稳定性试验。没有做过的 pin 不写成失败。两个校验和必须与可信准备记录一致；逐项正文与机器标记交叉审查，不能只凭 ID 或缓存存在放行。
  为每个实际使用的例外及每个 removed 项写唯一绑定标记和可审阅正文；不要把所有删除指向一条不含对应模块身份的泛化说明。填充 Membership/Evidence 必须基于依赖路径审核，不能复制实际快照当期望。原值自然保持则不强行使用例外；tools 四节点缺失仍须重新报批。
  Expected：实际例外为 root 31/tools 4 固定集合的子集；每项版本、校验和、活动链、测试/构建条件审查可复核，未运行范围明确。非 Required 移出需原版本 removal 记录，32 Required 均不能移出。

- [x] **R6. 严格离线全图、八组包、证据绑定和稳定性复验。** 固定 Go；GOENV/GOWORK=off、GOTOOLCHAIN=local、GOPROXY/GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly。两模块 list -m -json all、go mod graph 各有 90 秒上限；tools 元数据用明确 modfile 或所属目录，记录真实 cwd/argv。禁止 -e。
  八组命令沿用下方原 Step 1 的完整字段：root/integration 为 -deps -test ./...（后者加 -tags=integration），Goose 为 -deps 加其 CLI 路径；buf/protoc/oapi/sqlc/lint 分别使用其既有固定完整工具包路径、-deps、显式 tools modfile。每条 90 秒，不执行第三方测试。TestDevtoolsVersionLock 通过 readDevtoolsPackageSnapshot 读取这八组并逐组与 baseline.json 比较 ImportPath/Module.Path/Module.Version，同时调用用途及证据检查器，不读取忽略日志当实时成功证据。
  再在各所属模块目录正常离线 tidy，四锁字节和新模块选择必须不变。重跑：
```powershell
go test -v -tags=e2e ./internal/e2e -run '^TestDevtoolsVersion(Policy|Evidence|PackageUse|Lock)$' -count=1 -timeout=5m
```
  Expected：所有适用规则通过；root 28/tools 4 存在且值与条件合规，root 三组无新增 28 包，tools 五组无原四包，root/Goose mathutil v1.7.1 不变。八组映射及 52 保护哈希不变。四锁与整理前字节比较，不要求拆分后根锁等于迁移前历史哈希。专项 5m/R7 10m 总预算不扩展，超时保留阻塞，不替代 mod verify。

- [x] **R7. 补 lint fixture 并跑原专项。** 在 newDevtoolsRouteFixture 增加只存在于根模块的 rootonly/fixture.go（无外部依赖）；TestDevtoolsLintTargetsRoot 保留真实 cwd/argv/GOFLAGS 断言，增加解析该文件及其包归属的检查。另用仅测试副本将 lint cwd 改为 tools/devtools，要求现有断言失败，不能靠假工具自报成功。先见针对错误对象的 RED，再保持正常对象 GREEN；真实 lint 在 Task 3 单独验证。
  核对现有 Step 5a/5b 不回退，执行下方 Step 6 完整选择器（TestDevtools.* 自动包含新版本测试）及 Linux 独立拒绝文件选择器，仍为原 10m/3m 上限。
  Expected：新增版本测试实际出现在 RUN 清单，适用测试全部通过；Linux 四入口拒绝不作 Unix 正向成功；无额外预算、隐私断言删减或忽略历史失败。

- [x] **R8. 原子提交及接续。** 只有 R1–R7 和原 Task 2 的全部适用合同成立，才执行 Step 7 的原子提交，额外明确加入新版本测试、两份 testdata 和脱敏证据文档；检查暂存清单不包含其他未批准修改或 .superpowers。使用 task-done 重跑原完整选择器记录最终结果，不因历史绿色直接完成。然后继续 Task 3；有阻塞则保存事实，不提交部分模块/消费者破链状态。
  Expected：提交可在新电脑恢复基线/策略/源码；Task 2 完成记录有实际最终证据。推送、合并和安全验收恢复均仍不授权。

### 原任务步骤（保留历史与未完成合同）

- [x] **1. 记录不可变版本基线。** 在修改前用固定 Go 读取 `go mod graph`、模块选中版本及普通、
  integration、Goose CLI 和五个工具的包闭包。结果仅写忽略的证据目录。
  每条有界只读命令出现离线缺失就停止，不能把不完整 -e 输出当成功基线。
  记录根 go.mod/go.sum、runner、生成目录的哈希。环境不允许下载时报告准备阻塞。

```powershell
go list -m -json all
$packageFields = 'Dir,ImportPath,Name,Standard,ForTest,Module,Match,DepOnly,GoFiles,CgoFiles,TestGoFiles,XTestGoFiles,Imports,ImportMap,Deps,TestImports,XTestImports,Incomplete,Error,DepsErrors,EmbedFiles,TestEmbedFiles,XTestEmbedFiles'
go list -deps "-json=$packageFields" -test -mod=readonly ./...
go list -deps "-json=$packageFields" -test -tags=integration -mod=readonly ./...
go list -deps "-json=$packageFields" -mod=readonly github.com/pressly/goose/v3/cmd/goose
```

  执行裁定：默认全字段 JSON 还会计算 Stale 构建状态；本地 Go 源码确认这触发额外构建器
  工作，非版本基线要求。明确保留上述依赖/模块/错误字段，仍解析完整包集合、不使用 -e。
  Goose 原全字段查询及五工具合并查询均曾触及 90 秒诊断上限，失败事实保留；不声称性能
  缺陷已修复。五工具按各自完整路径逐项使用相同字段和 90 秒上限，最终合并其依赖集合。
  这不是完整性校验、构建或测试通过证据，不改变任何产品及验收预算。

- [x] **2. 写 RED 边界与调用测试。** 新增 `TestDevtoolsModuleBoundary`、`TestDevtoolsRouting`、
  `TestDevtoolsNestedProtoc`、`TestDevtoolsLintTargetsRoot`、`TestDevtoolsFailureStopsConsumers`。
  下面断言同时用于源合同检查和 fake 调用事件检查，不只匹配任意源码子串：

```text
root.tool == {github.com/pressly/goose/v3/cmd/goose}
devtools.tool == {buf, protoc-gen-go, oapi-codegen, sqlc, golangci-lint 的完整包路径}
Goose library remains root; protobuf and oapi runtime remain if used
generate first event = successful verify-devtools
all moved tools use exactly the repository tools/devtools/go.mod
all tool consumer working directories = repository root
verification failure => zero version/generation/lint/Docker events
```

  模块合同用 `go mod edit -json` 解析，避免为测试把新解析依赖重新引入主模块。
  根模块专属测试包放进 fixture；假 lint 验证 cwd/argv/环境，后续真实 lint 验证根源码确被分析。
  为 Buf 建立会真正执行本地插件命令的 fake buf，不只把 `buf generate` 当成功文本输出。
  更新既有 privacy fake 的 tool 分派以解析 -modfile 参数，但不放宽 canary 或顺序断言。
  在 Task 1 的独立拒绝文件新增 `TestDevtoolsShellEntryRejection`，以及路由文件中的
  `TestDevtoolsConsumerOwnership`。前者遍历四个 `.sh` 入口调用
  `testDevtoolsBashEntryRejection(t, entry)`，在 Windows 与 Linux 均断言零工作、统一提示、
  stdout 空、退出 1；后者遍历三个 `.ps1` 消费者，
  让验证成功后 fake 工具创建长存子进程，强杀外层入口并检查全部自有句柄退出及外部 canary 存活。
  在 C11 fixture 中假造 Docker 响应，不创建真实容器；Job 测试不宣称 Docker 服务端清理成功。

```go
entries := []string{"verify-devtools.sh", "check-tools.sh", "generate.sh", "verify-c11.sh"}
consumers := []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"}
```

  旧 Bash 工具入口隐私正向测试不迁移到 Unix：四入口在所有平台均不支持。按入口逐项改为
  拒绝合同，原测试版本由历史提交保留。特别核对 privacy_test.go 中
  `TestVerifyC11BashContract`、`TestVerifyC11BashWindowsConformancePathReachesNativeGo`、
  `TestVerifyC11BashWindowsConformancePathRejectsMalformedConversion`；后两项分别改名为
  `TestVerifyC11BashRejectsBeforeNativeGo`、`TestVerifyC11BashRejectsBeforePathConversion`，
  测试拒绝发生在工具/路径转换之前，不保留旧成功路径或改成无条件 Skip。
  `TestScriptCleanupExitStatusContracts` 的 Bash/verify 项改验拒绝，不删 Bash/smoke 项。
  `TestVerifyFakeToolsExerciseEveryPrivacyCanary` 和旧 round2 gate 契约对 PowerShell 的断言
  全保留；仅移除四个已撤回 `.sh` 的正向工具预期，并加入其拒绝行为覆盖。
  `TestSmokeBashWindowsScriptPathLoadsRepositoryEnvironment` 与其他 smoke 用例继续运行。
  不移除 Bash smoke 隐私用例，不全局 Skip Bash 测试。fixture 复制私有帮助文件，保留真实 Job
  初始化，不能替换为空操作以使取消测试通过。
  增加 `TestDevtoolsSameHostDispatch` 和 `TestDevtoolsConsumerRuntime`：三个消费者均测试真实
  5.1 早期拒绝、7.6.5 接受；在专属 PATH 最前放置写入事件的伪 pwsh/powershell，真正父入口用
  已验证绝对路径启动。ConsumerRuntime 同时复用 Task 1 的非 Windows 平台判定逻辑测试，
  标为测试副本证据，不宣称实际 Unix PowerShell 运行验证。
  仅在同宿主分派子脚本测试副本中加入采集版本和当前进程路径的标记，不替换实际
  子解释器或初始化。子入口断言：

```go
if childRuntime.Version != "7.6.5" || !strings.EqualFold(childRuntime.Executable, parentExecutable) {
    t.Fatalf("nested runtime changed: %+v", childRuntime)
}
if _, err := os.Stat(fakeInterpreterEvent); !os.IsNotExist(err) {
    t.Fatalf("PATH interpreter was invoked: %v", err)
}
```

  测试内的 childRuntime 为从专属 JSON 标记读取的 `{Version string; Executable string}`；
  parentExecutable 为测试启动父进程所用的规范绝对路径。标记文件不属于产品输出。
  对 C11 的 check-tools、generate、verify-devtools 三条分派分别断言；保留 smoke 原宿主，
  不把 smoke 的合法 powershell 调用误判为这四个入口的分派失败。

- [x] **3. 跑 RED。**

```powershell
go test -tags=e2e ./internal/e2e -run '^TestDevtools(ModuleBoundary|Routing|NestedProtoc|LintTargetsRoot|FailureStopsConsumers|ShellEntryRejection|ConsumerOwnership|SameHostDispatch|ConsumerRuntime)$' -count=1 -timeout=5m
```

  Expected RED: 模块边界、消费者前置验证/归属/同宿主分派和后三个 `.sh` 拒绝尚未实现，
  各自失败于对应断言；不能以环境错误冒充 RED。Linux 复用 Task 1 Step 2 的固定禁网容器命令，
  将 -run 改为 `^TestDevtools(VerificationBash|ShellEntryRejection)$`，保持只编译拒绝测试文件。

- [x] **4. 建立新模块并整理根模块。** 新模块基本内容固定如下，require/sum 必须从 Step 1
  的实际选中闭包生成并逐项对照，不在计划中臆造完整间接版本表。

```go
module talenro.local/devtools
go 1.26.0
toolchain go1.26.5
tool (
    github.com/bufbuild/buf/cmd/buf
    github.com/golangci/golangci-lint/v2/cmd/golangci-lint
    github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
    github.com/sqlc-dev/sqlc/cmd/sqlc
    google.golang.org/protobuf/cmd/protoc-gen-go
)
require (
    github.com/bufbuild/buf v1.72.0
    github.com/golangci/golangci-lint/v2 v2.12.2
    github.com/oapi-codegen/oapi-codegen/v2 v2.8.0
    github.com/sqlc-dev/sqlc v1.31.1
    google.golang.org/protobuf v1.36.11
)
```

  使用 apply_patch 创建初始文件/编辑 tool 表，go mod tidy 属机械整理，可在明确模块目录运行。
  接续时不重建已有模块；版本处理以 R1–R6 和已批准锁版本修订为准。仅添加具有保留来源、
  经稳定性验证的约束，不把全图全部直接 require。不保留无实际来源的工具依赖来掩盖拆分失效。
  不借 tidy 升级、不 use latest、不生成 workspace、不引入跨模块 replace。
  根模块移除五个 tool 指令，保留 Goose；doubleclick 若仍出现，先报告真实引入链，不强删。

- [x] **5a. 同步入口支持与归属。** 三个 `.sh` 消费者替换为 Task 1 Step 4c 同型最小拒绝入口，
  提示分别为 `check-tools: this release requires Windows and PowerShell 7.6.5.`、
  `generate: this release requires Windows and PowerShell 7.6.5.`、
  `verify-c11: this release requires Windows and PowerShell 7.6.5.`。移除旧执行体，不留回退分支。
  三个 `.ps1` 最前部使用 Task 1 Step 0b 的进程内检查，拒绝提示分别为
  `check-tools: PowerShell 7.6.5 required.`、`generate: PowerShell 7.6.5 required.`、
  `verify-c11: PowerShell 7.6.5 required.`。检查通过后才加载帮助文件并初始化；初始化 catch 按各入口现有
  internal 阶段脱敏输出、退出 1。C11 的 git 清单命令也必须在初始化之后。
  PowerShell C11 中 Get-C11BashExecutable 和 Bash smoke 调用保持，不删除或转换它们。

- [x] **5b. 同步 Windows 脚本调用。** generate/check-tools 的 `.ps1` 在任何工具前调用对应 verify-devtools；
  C11 在工具阶段前也调用验证入口，保持原 check-tools 900 秒、generate 600 秒外层预算。
  外层剩余时间不足时允许明确失败，不用改预算或跳过标记规避重复验证。
  Windows 消费者独立调用验证器的参数固定为：

```powershell
$devtoolsPowerShell = [Environment]::ProcessPath
if ([string]::IsNullOrWhiteSpace($devtoolsPowerShell) -or
    -not [IO.Path]::IsPathFullyQualified($devtoolsPowerShell) -or
    [IO.Path]::GetFileName($devtoolsPowerShell) -ine 'pwsh.exe' -or
    -not [IO.File]::Exists($devtoolsPowerShell)) {
  throw 'devtools host resolution failed'
}
$verifyArguments = @('-NoProfile', '-NonInteractive', '-File', (Join-Path $PSScriptRoot 'verify-devtools.ps1'))
```

  该路径读取只在 7.6.5 检查通过后执行，错误走当前入口脱敏 internal/host 阶段退出 1；不
  输出路径、不回退 PATH 或系统 powershell.exe。C11 的 check-tools/generate 同样使用
  `$devtoolsPowerShell`；不要覆盖用于原有 smoke 的 `$powerShell`，避免扩大迁移范围。
  用各入口现有受控执行函数传递这组独立参数，检查非零退出立即停止。check-tools 没有通用
  执行器时新增本地 `Invoke-DevtoolsVerification`（无参、无成功输出、失败调用 Stop-Script）；
  保留脱敏和退出码，不采用执行字符串、不新增 ExecutionPolicy 绕过。

```powershell
$devtoolsMod = Join-Path $repoRoot 'tools/devtools/go.mod'
# 保留 Invoke-External 的原脱敏与错误传播语义。
Invoke-External -Stage 'generate: SQL' -FilePath 'go' -ArgumentList @('tool', "-modfile=$devtoolsMod", 'sqlc', 'generate')
```

  check-tools 仅为那五个工具增加 -modfile；Goose 仍是 `go tool goose -version`。
  C11 的 lint 是 `go tool -modfile=<absolute tools module> golangci-lint run ./...`，cwd 根。
  不把 -modfile 放进全局 GOFLAGS；脚本不接受用户提供任意备用模块路径。

```yaml
# buf.gen.yaml 中保留现有 out 和 opt，只替换 local 命令。
local: ["go", "tool", "-modfile=tools/devtools/go.mod", "protoc-gen-go"]
```

- [x] **6. GREEN 与兼容性确认。** 重跑 Step 3 及 Task 1；运行下列真实现有脚本契约：

```powershell
go test -tags=e2e ./internal/e2e -run '^(TestDevtools.*|TestVerifyC11.*Contract|TestVerifyC11Bash.*Reject.*|TestSmokeBashWindowsScriptPathLoadsRepositoryEnvironment|TestVerifyFakeToolsExerciseEveryPrivacyCanary|TestScriptCleanupExitStatusContracts|TestVerifyCommandContractRejectsLegacyMissingRound2Gates)$' -count=1 -timeout=10m
```

  同时运行 Step 3 的 Linux 拒绝文件选择器。Expected GREEN: Windows 上全部适用专项通过；
  Linux 四入口实际拒绝通过；不记录 Unix 正向工具流水成功。新增/重命名的拒绝测试须全部
  命中上述选择器，执行后检查实际 RUN 清单，不能因漏选得到虚假的绿色结果。
  测根路径含空格、继承 GOFLAGS/-modfile 污染、缺工具锁文件、父任务取消和 Buf 子命令失败。
  对实际 argv/cwd/退出码做断言，保留既有隐私测试。确认 runner、smoke 和生产源码无差异。
- [x] **7. 原子提交所有模块及路由文件。**

```text
git add go.mod go.sum tools/devtools/go.mod tools/devtools/go.sum scripts/generate.ps1 scripts/generate.sh scripts/check-tools.ps1 scripts/check-tools.sh scripts/verify-c11.ps1 scripts/verify-c11.sh buf.gen.yaml internal/e2e/privacy_test.go internal/e2e/devtools_routing_test.go internal/e2e/devtools_bash_rejection_test.go
git commit -m "refactor(tooling): isolate development tools from C12 module verification"
```

## Task 3: 真实验证、恢复手册与最终审查交接

2026-09-28 执行证据：Task 2 已本地提交 a0ca189；根完整性退出0/875.905秒，公开工具验证在依赖阶段退出1/7.296秒，147项缓存缺目录（27项同时缺归档/哈希）。仅编译退出0/59.767秒，未执行测试。后续工具/生成/lint未运行。一次独立审查已完成，缓存验证/执行绑定 Important 尚待授权修复；恢复文档补齐固定工具链缓存前提，Minor及审查排除项留存。Task 3 不标完成、不运行 task-done；原安全、main及推送门禁不变。见[真实验证检查点](../../roadmap/2026-09-28-devtools-real-verification-checkpoint.md)。

**Files:** README.md、docs/runbooks/repository-recovery.md、docs/roadmap/current-status.md，本计划证据区。
**Interfaces:** 消费 Task 1/2 的公开入口；输出逐阶段真实结果，不修改代码来覆盖失败。

- [x] **1. 更新新电脑恢复说明。** 两个模块分别准备依赖（网络准备与离线验收明确分开），
  说明显式工具命令和裸 go tool 兼容性变化；固定版本，不加入防护排除项建议。
  先由用户准备 Microsoft 正式发布的 PowerShell 7.6.5。安装/下载属于单独环境准备，不在
  此计划离线执行步骤内自动完成；文档写清 5.1 不受支持、其他 7.x 也未纳入本次精确版本范围。
  允许用户用完整路径启动 pwsh.exe，禁止把本机 Codex 运行时位置写进通用恢复命令。
  手册明确仅 Windows 四个 `.ps1` 为正向入口；四个 `.sh` 在 Linux、macOS、WSL 和 Windows
  都只拒绝，Unix 暂不支持。既有 Bash smoke 仍保留，按其原前提准备 Git Bash；这不等于支持
  Bash 工具流水，也不要求把 Linux Docker 引擎改成 Windows 容器。
  人工检查命令（不改变系统）：

```powershell
pwsh.exe -NoProfile -Command '$PSVersionTable.PSVersion.ToString(); $PSVersionTable.PSEdition'
```

  Expected: `7.6.5` 和 `Core`；正式入口仍自行检查，不能靠这次检查生成跨运行跳过标记。

```text
go mod download all
go -C tools/devtools mod download all
```

  上述是人工/受批准准备步骤，不在离线验证器中执行。下载缺包必须保留原因，不能验收中偷偷联网。
- [ ] **2. 固定实现提交，依次真实验证。** 记录每条命令、提交、起止时间、退出码和输出摘要。
  根完整 mod verify 成功仍不能替代工具模块验证，反之亦然。

```powershell
pwsh.exe -NoProfile -File scripts/verify-devtools.ps1
pwsh.exe -NoProfile -File scripts/check-tools.ps1
pwsh.exe -NoProfile -File scripts/generate.ps1
git diff --exit-code -- api gen internal/store
```

  上述流程成功后，单独验证真实根 lint：先再次通过公开 verify-devtools.ps1，再在根目录运行
  `go tool "-modfile=$((Join-Path (Get-Location).Path 'tools/devtools/go.mod'))" golangci-lint run ./...`。
  使用固定 Go 和独立受控进程，沿用 C11 lint 的 600 秒阶段上限，保存实际 argv/cwd/退出码；
  不通过完整 verify-c11.ps1 间接启动，以免触发受阻普通/C12 路径。不将 -modfile 加进 GOFLAGS。
  Expected：真实 lint 退出 0，分析根源码而非工具模块；超时/失败如实保留，不调整规则求绿。

  Expected: 三个 Windows 入口退出 0，生成目录 git diff 退出 0。任一失败保留原输出摘要，
  不能继续当作工具链成功。生成变更不允许直接提交，先调查版本/路径差异。
  任一工具验证超时则整体工具链验收开放，不标记通过、不重复相同失败运行。
  当前安全事件阻塞仍在，不运行 C12 专项测试来补足工具链证据。Unix 正向生成已撤回本轮
  交付范围，不再运行或记为成功；四入口的实际 Windows/Linux 拒绝结果仍独立列出。
  恢复手册列出支持矩阵、保留的 Bash smoke 及历史 Unix 失败对应的归档提交。
- [ ] **3. 普通与编译回归（执行门禁）。** 下列完整普通测试及完整 C11 会触发当前受阻路径，
  在安全事件得到单独处理并明确允许恢复前保持未运行。仅 compile-only 命令可先执行；它不运行
  bootstrap，也不记为普通测试通过。不得通过排除失败子测试声称“全套通过”。

```powershell
# 可先执行：仅编译，不运行测试函数。
$env:GOOS='windows'
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
```

```powershell
# 当前禁止执行：安全事件处理并明确允许恢复后才运行。
$env:GOOS='windows'
go test ./... -count=1 -timeout=15m
```

  编译成功明确标为 compile-only。完整 e2e/C11 运行沿用当前仓库入口和外层原时限，
  使用已有受控本地测试环境，不给任何既有测试增加预算。缺环境就记录未运行。
- [ ] **4. 独立执行 13 项物理 gate（当前阻塞，不执行）。** 仅在安全事件处理、明确恢复授权及
  原有环境前提均满足后执行；任务 1/2 成功不自动打开此门禁。使用原批准计划
  `2026-09-19-c12-controlled-pitr-test-interface.md` Task 6 Step 4 原命令；以下集合逐条调用，
  不能合并正则，也不能直接运行 integration 包全体测试：

| Profile / Package | Run / seam |
| --- | --- |
| authority-v7-pitr / ./internal/testinfra | ^TestC12AuthorityPITRProfile$ |
| authority-v7-pitr / ./internal/nodecontrol/authority | ^TestPITRBeforeRevocationFailsClosed$ |
| authority-v7-pitr / ./internal/testinfra | ^TestC12AuthorityPITROwnershipWALFailureSeam$，各自独立运行以下五个 seam |
| authority-v7 / ./internal/nodecontrol/authority | 以下六个独立测试 |

```text
after-intent-before-create
after-create-before-actual
after-actual-before-return
after-clean-intent-before-remove
after-remove-before-clean-result

TestCoordinatorPostgresCrashRecoveryMatrix
TestCoordinatorPostgresAbortDomainSafetyAndIdempotence
TestCoordinatorPostgresCrashResolverFailsClosedOnSQLDeletionOrCorruption
TestCoordinatorPostgresAtomicActivationCrashMatrix
TestCoordinatorPostgresFenceFirstAbortRace
TestCoordinatorPostgresFinalizeCapturesAfterEffectWriterCommit
```

```powershell
# 每次填入上表的一个精确测试；所有原脚本检查保持。
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITRProfile$' -Timeout 5m
```

  私有 seam 在原精确 selector 后添加 `-PITRFailureSeam '<上列精确值>'`。
  要求 go-json PASS 而非 Skip，并检查原 cleanup/foreign-canary 及 A/B 恢复断言。
  共享准备再次超时就停止其余同原因调用；不把扩大缓存预算当修复，不改 runner。
- [ ] **5. 更新事实与提交。** 当前状态列实现、工具验证、普通测试、13 gate、I2/I3、
  缓存/Defender 性能限制各自状态。真实工具验证还失败时也如实记录，不删除原诊断。
  不以暖缓存结果声称冷态稳定，不清缓存造冷态。

```text
git add README.md docs/runbooks/repository-recovery.md docs/roadmap/current-status.md docs/superpowers/plans/2026-09-25-c12-development-tool-isolation.md
git commit -m "docs(tooling): record isolated toolchain recovery and verification evidence"
```

- [ ] **6. 整体审查与交接。** 按用户选择的执行方式完成独立审查，重点检查双模块完整性、
  Windows 子进程/外层取消、全平台无工作拒绝、保留的 Bash smoke 及未降低现有保护。
  审查缺陷修复需独立 RED/GREEN。
  无条件保持 main 不变；推送/合并遵循届时明确授权，不能把计划批准当合并批准。

## 自查及审批

设计第 1–3 节对应 Task 2 的模块/版本基线，第 4 节对应 Task 2 的分派和嵌套测试，
第 5 节对应 Task 1 的验证和 Task 2 的接入，第 6–8 节对应文件映射和 Task 3 验收交接。
五项 Review Focus 均已分配测试。Go verify 跳过完全未下载模块的实际行为已纳入 Task 1。
新 900 秒入口不扩展 C11 900/600 秒外层预算，这一潜在阻塞明确保留。

附录第 2 节的平台边界对应 Task 1/2 的拒绝测试与 Task 3 文档；第 3 节 Job 生命周期对应
Task 1 的私有帮助文件和 Task 2 的全入口取消测试；第 4 节预算及完整性对应 Task 1/2；
第 5–6 节对应测试迁移、阻塞门禁与交接。Add-Type 初始化自身的子进程风险已明确列为停止条件，
不因当前 PowerShell focused 测试曾通过而跳过。Docker 服务端清理不在新增 Job 保证内。
7.6.5 附录修订对应 Task 1 Step 0a/0b 的版本 RED/GREEN、Step 4a 的初始化观察，Task 2
SameHostDispatch/ConsumerRuntime 及消费者调用，Task 3 的新电脑前提与独立证据。
保留文中的 C12 powershell 调用作为原合同的受阻命令，不做全局替换；当前不得执行。

2026-09-27 Windows-only 书面设计和本次对应计划修订均已获用户确认。
修订第 1–2 节由 Task 1 Step 1/2/4c/5 和 Task 2 Step 2/3/5a/6 覆盖；第 3 节保留
Windows 既有完整性、Job、版本及路由测试；第 4 节由 Task 1 Step 0c 历史提交和 Task 2
privacy/smoke 逐项迁移覆盖；第 5–6 节由三任务顺序、Task 3 支持矩阵及原安全门禁覆盖。
已对齐统一拒绝文案、标准库独立拒绝夹具、生产 900 秒/4 MiB 与既有 900/600 秒外层预算；
Windows 内存捕获裁定已反映到计划，未恢复旧落盘捕获。没有通过支持范围调整勾选任务完成。

继续保留 Native inline：本会话逐任务执行，最后进行一次独立整体审查。
2026-09-27 的历史锁版本计划及三个 root 候选准备已获批准。2026-09-28 四项 tools 书面设计、对应计划修订与 R3b 精确准备范围均已获确认；既有 Native inline 不变。
历史锁版本与四工具修订的基线/权限/用途合同保留；与最新根 28 设计冲突处以上方新版 R2–R6 为准。Required32 不变，候选上限 root31/tools4；历史绿色不替代新回归。
四工具设计原合同的对应关系保留。最新 root28 设计第 1–3 节对应全局约束/R2 的独立 32/35 字面集合；第 4 节对应 R4/R5 逐项链、校验来源、测试/平台及历史父边审查；第 5 节对应 R1/R2/R6 的不可变基线和联合 Lock；第 6 节对应 R3c-1/2 的只读核查、独立准备授权和总期限；第 7 节对应 R2/R6/R7 反例、包/保护哈希、预算与 Task 3 真实验证；第 8 节对应 R8、安全门禁及审批交接。未做实验不写成失败，文件齐备不写成完整性通过。
本次计划修订与 R3c 27 候选精确准备已由用户同时确认。授权仍限上述清单，任何新增准备范围另行处理。任何批准不等于安全事件解除、测试通过、推送或合并。
