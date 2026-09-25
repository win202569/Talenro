# C12 Development Tool Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将五个开发工具与 C12 主模块依赖分离，保留完整校验并重新取得真实验收证据。

**Architecture:** 根模块保留业务、测试、Goose 库及 CLI。独立 `tools/devtools` 模块保存五个开发工具，公开脚本先验证工具模块，再以显式 -modfile 调用工具。C12 runner 不修改。

**Tech Stack:** Go 1.26.0 / go1.26.5、Windows PowerShell、Bash、现有 Go 脚本契约测试、Docker Desktop。

**Spec:** `docs/superpowers/specs/2026-09-25-c12-development-tool-isolation-design.md`（用户已批准）。

## Global Constraints

- 当前工作目录为既有 `codex/c12-b01-task9-coordinator` worktree，基线文档提交 `ff302750`；不创建第二个 worktree，不切换 main。
- 新模块 `talenro.local/devtools`；不建立 go.work，不添加两模块间 require/replace。
- Buf 1.72.0、protoc-gen-go 1.36.11、oapi-codegen 2.8.0、sqlc 1.31.1、golangci-lint 2.12.2；Goose 3.27.1 留在根模块。
- 不升级或隐式降级保留的选中版本；无法保留的差异先报告，不自行接受。
- 不关闭防护、不添加排除项、不删除共享模块缓存或第三方 testdata。
- 新工具验证独立默认 900 秒；既有 C11 外层超时保持，嵌套可更早失败，不扩展外层预算。
- C12 原准备预算及各项 5m 测试预算、候选树、模块图、进程、清理保护一律不改。
- 不提交 `.superpowers`、ETL、诊断二进制、共享缓存、原始敏感输出。
- 不把实现完成、编译通过、fake harness 通过或工具验证通过当作物理验收。
- I2/I3 仍开放；任何阶段失败均保留原因，停止同条件重试；不合并 main。

## Review Focus

1. 缓存完全缺失时 Go mod verify 可能成功：Task 1 必须先证明依赖齐备且离线缺失失败。
2. 仓库路径含空格、Bash/MSYS 转换及 Buf 子命令：Task 2 验证参数边界和根工作目录。
3. 工具构建使用备用模块，lint 子进程误用工具模块：Task 2 用根模块专属包验证分析对象。
4. 拆分触发间接版本降低或生成差异：Task 2 锁定迁移前选中版本，Task 3 真实生成比较。
5. 子进程自然退出与超时竞争、外层提前终止及脱敏：Task 1/2 验证只清理自有进程且返回失败。

## 文件职责与顺序

| 文件 | 职责 / 任务 |
| --- | --- |
| scripts/verify-devtools.ps1、scripts/verify-devtools.sh（新增） | Task 1，独立离线验证、期限和自有进程清理 |
| internal/e2e/devtools_verification_test.go（新增，e2e tag） | Task 1，验证脚本的进程、缓存、错误与隐私测试 |
| go.mod、go.sum；tools/devtools/go.mod、go.sum（新增） | Task 2，依赖边界与固定版本 |
| scripts/generate.ps1/.sh、scripts/check-tools.ps1/.sh、scripts/verify-c11.ps1/.sh | Task 2，入口验证及显式工具分派 |
| buf.gen.yaml | Task 2，protoc-gen-go 嵌套分派 |
| internal/e2e/privacy_test.go；internal/e2e/devtools_routing_test.go（新增，e2e tag） | Task 2，脚本契约、模块边界、版本和调用对象 |
| README.md、docs/runbooks/repository-recovery.md、docs/roadmap/current-status.md | Task 3，恢复步骤、兼容性、实际验收记录 |

Task 1 不接入公开调用链，允许根目录尚无工具模块时独立入口明确失败。
Task 2 原子提交模块拆分和所有调用点，不留下根工具已删除但脚本仍调用旧路径的提交。
Task 3 执行真实验证与文档交接。每任务自带测试周期和独立可审查结果。

## Task 1: 新增独立开发工具验证入口

**Files:** 创建两份 verify-devtools 脚本和 `internal/e2e/devtools_verification_test.go`。

**Interfaces:**
- `powershell -NoProfile -File scripts/verify-devtools.ps1` / `bash scripts/verify-devtools.sh`，无公开跳过、模块路径、命令或超时覆盖参数。
- 成功输出仅 `verify-devtools: passed`，退出 0；失败仅输出 `verify-devtools: <stage> failed with exit code <n>.`，不回显底层输出。
- Go 非零退出原样传播；校验语义错误/清理失败退出 1；命令缺失 127；超时 124。
- 不启动生成、lint、Docker；不得 dot-source 或执行 C12 runner 来借用函数。

- [ ] **1. 写 RED 测试。** 新文件使用 `//go:build e2e` 和 `package e2e`。
  用 t.TempDir 创建带空格的镜像仓库，复制待测脚本，提供固定的 tools/devtools/go.mod，
  从现有 privacy_test.go 的假工具及进程测试模式建立本任务专属 fixture。
  初始无脚本时断言入口存在应失败，而不是 Skip。测试表固定为：

```go
cases := []struct{name string; want int}{
    {"valid", 0}, {"missing-module", 1}, {"missing-cache", 1},
    {"tampered-directory", 1}, {"tampered-zip", 1},
    {"go-exit-7", 7}, {"wrong-version", 1}, {"timeout", 124},
    {"timeout-with-child", 124}, {"output-canary", 7},
    {"inherited-module-override", 1},
}
```

  测试名使用 `TestDevtoolsVerificationPowerShell`、`TestDevtoolsVerificationBash`。
  缺失和篡改测试用 fixture 独占缓存，绝不修改用户缓存。
  为真实完整性分支建立一个小型本地模块代理 fixture：go.mod、版本 .info、zip、list，
  准备阶段使用本地 file 代理下载到 fixture 缓存，再切换离线运行，逐个损坏 zip/源码。
  checksum 缺失、zip 和目录全部缺失必须失败；成功 fixture 运行真实 go mod verify。
  fake Go 只模拟进程/错误/环境合同，不作为真实哈希验证证据。

- [ ] **2. 跑 RED。**

```powershell
$env:GOOS='windows'
go test -tags=e2e ./internal/e2e -run '^TestDevtoolsVerification(PowerShell|Bash)$' -count=1 -timeout=5m
```

  预期为新增入口未实现的断言失败。Bash 缺失须记录环境阻塞，不能称双平台通过。

- [ ] **3. 实现验证命令链。** 在脚本中以脚本位置解析根目录及工具模块，先检查两个锁文件存在。
  固定环境使用进程级设置且 finally/trap 恢复：GOWORK=off、GOENV=off、GOTOOLCHAIN=local、
  GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly；
  继承的 GOFLAGS 中若含 -modfile/-overlay，先拒绝并退出 1；其余继承 GOFLAGS、
  GOEXPERIMENT、GOCACHEPROG 和模块/工具链覆盖项清除后使用上述固定设置。
  Windows 使用既有 C12 固定工具链目录；Bash 原生使用已解析 Go，必须精确匹配 go1.26.5，
  Git Bash 则验证 Windows 本机 go1.26.5。所有命令在 tools/devtools 下运行：

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

- [ ] **4. 实现进程预算与输出保护。** 单个入口从启动起只用一个 900 秒绝对截止时间，
  version、list、verify 共用剩余预算，不给每条命令另加 900 秒。
  PowerShell 以现有 verify-c11.ps1 的 `ConvertTo-C11CommandLineArgument`、
  `Get-C11CompletedProcessExitCode`、`Stop-C11OwnedProcessTree`、`Invoke-C11External`
  为已知模式，在新脚本内部建立私有实现；不改旧 C11 函数、不引入系统全局清理。
  Start-Process 使用 Hidden，读取 Process.Handle 后等待，超时停止自有进程树。
  Bash 使用现有 `timeout --kill-after=5s` 模式和专属临时目录，失败必须清理自有 sink；
  Windows Bash 不能仅靠 POSIX PID 误杀 native PID，沿用已有 C11 native 进程识别测试要求。
  外层提前终止时也不得留下 native Go/子进程；无法证明这一点则 Task 1 不通过。

```powershell
# 每次启动前从同一个截止时间计算；耗尽即失败，不自动续期。
$remaining = [Math]::Floor(($deadline - [DateTime]::UtcNow).TotalSeconds)
if ($remaining -le 0) { throw 'verify-devtools deadline exhausted' }
```

  输出写入自有随机临时目录中的 sink，校验阶段只解析需要的数据，不打印原始错误。
  受控捕获上限 4 MiB，超限失败；清理失败不能被成功状态覆盖。
  私有执行函数接受 Deadline 供测试传入短期限，公开入口固定 900 秒；不增加生产跳过参数。

- [ ] **5. GREEN 与独立负向确认。** 运行 Step 2，增加超时边界前自然退出、子进程存活 canary、
  父进程取消、stderr 敏感 canary、目录含空格和调用后环境不泄漏断言。
  移除齐备检查的测试副本必须使 missing-cache 负向测试失败，证明不会受 Go 的跳过语义欺骗。
  只在 fixture 中做变异，不修改真实脚本来求通过。
- [ ] **6. 提交。**

```text
git add scripts/verify-devtools.ps1 scripts/verify-devtools.sh internal/e2e/devtools_verification_test.go
git commit -m "feat(tooling): add fail-closed offline development tool verification"
```

## Task 2: 原子拆分模块并更新全部调用链

**Files:** 文件映射中的模块文件、六份入口脚本、buf.gen.yaml、privacy_test.go、devtools_routing_test.go。

**Interfaces:** 消费 Task 1 的无参公开入口；新工具模块锁文件和显式 -modfile 调用是唯一新增分派合同。
  Goose 裸 go tool 命令保持。子工具和 lint 仍以根工作目录运行。

- [ ] **1. 记录不可变版本基线。** 在修改前用固定 Go 读取 `go mod graph`、模块选中版本及普通、
  integration、Goose CLI 和五个工具的包闭包。结果仅写忽略的证据目录。
  每条有界只读命令出现离线缺失就停止，不能把不完整 -e 输出当成功基线。
  记录根 go.mod/go.sum、runner、生成目录的哈希。环境不允许下载时报告准备阻塞。

```text
go list -m -json all
go list -deps -json -test -mod=readonly ./...
go list -deps -json -test -tags=integration -mod=readonly ./...
go list -deps -json -mod=readonly github.com/pressly/goose/v3/cmd/goose
```

- [ ] **2. 写 RED 边界与调用测试。** 新增 `TestDevtoolsModuleBoundary`、`TestDevtoolsRouting`、
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

- [ ] **3. 跑 RED。**

```powershell
go test -tags=e2e ./internal/e2e -run '^TestDevtools(ModuleBoundary|Routing|NestedProtoc|LintTargetsRoot|FailureStopsConsumers)$' -count=1 -timeout=5m
```

- [ ] **4. 建立新模块并整理根模块。** 新模块基本内容固定如下，require/sum 必须从 Step 1
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
  先加迁移前闭包版本约束，再 tidy，然后比较选中版本；如有保留模块降级，添加明确 require
  保持基线再检查。不保留无实际来源的工具依赖来掩盖拆分失效。
  不借 tidy 升级、不 use latest、不生成 workspace、不引入跨模块 replace。
  根模块移除五个 tool 指令，保留 Goose；doubleclick 若仍出现，先报告真实引入链，不强删。

- [ ] **5. 同步脚本调用。** generate/check-tools 两端在任何工具前调用对应 verify-devtools；
  C11 在工具阶段前也调用验证入口，保持原 check-tools 900 秒、generate 600 秒外层预算。
  外层剩余时间不足时允许明确失败，不用改预算或跳过标记规避重复验证。

```powershell
$devtoolsMod = Join-Path $repoRoot 'tools/devtools/go.mod'
# 保留 Invoke-External 的原脱敏与错误传播语义。
Invoke-External -Stage 'generate: SQL' -FilePath 'go' -ArgumentList @('tool', "-modfile=$devtoolsMod", 'sqlc', 'generate')
```

```bash
run_quiet 'generate: SQL' go tool "-modfile=${repo_root}/tools/devtools/go.mod" sqlc generate
```

  check-tools 仅为那五个工具增加 -modfile；Goose 仍是 `go tool goose -version`。
  C11 的 lint 是 `go tool -modfile=<absolute tools module> golangci-lint run ./...`，cwd 根。
  不把 -modfile 放进全局 GOFLAGS；脚本不接受用户提供任意备用模块路径。

```yaml
# buf.gen.yaml 中保留现有 out 和 opt，只替换 local 命令。
local: ["go", "tool", "-modfile=tools/devtools/go.mod", "protoc-gen-go"]
```

- [ ] **6. GREEN 与兼容性确认。** 重跑 Step 3 及 Task 1；运行下列真实现有脚本契约：

```powershell
go test -tags=e2e ./internal/e2e -run '^(TestDevtools.*|TestVerifyC11.*Contract|TestVerifyFakeToolsExerciseEveryPrivacyCanary|TestScriptCleanupExitStatusContracts|TestVerifyCommandContractRejectsLegacyMissingRound2Gates)$' -count=1 -timeout=10m
```

  测根路径含空格、继承 GOFLAGS/-modfile 污染、缺工具锁文件、父任务取消和 Buf 子命令失败。
  对实际 argv/cwd/退出码做断言，保留既有隐私测试。确认 runner、smoke 和生产源码无差异。
- [ ] **7. 原子提交所有模块及路由文件。**

```text
git add go.mod go.sum tools/devtools/go.mod tools/devtools/go.sum scripts/generate.ps1 scripts/generate.sh scripts/check-tools.ps1 scripts/check-tools.sh scripts/verify-c11.ps1 scripts/verify-c11.sh buf.gen.yaml internal/e2e/privacy_test.go internal/e2e/devtools_routing_test.go
git commit -m "refactor(tooling): isolate development tools from C12 module verification"
```

## Task 3: 真实验证、恢复手册与最终审查交接

**Files:** README.md、docs/runbooks/repository-recovery.md、docs/roadmap/current-status.md，本计划证据区。
**Interfaces:** 消费 Task 1/2 的公开入口；输出逐阶段真实结果，不修改代码来覆盖失败。

- [ ] **1. 更新新电脑恢复说明。** 两个模块分别准备依赖（网络准备与离线验收明确分开），
  说明显式工具命令和裸 go tool 兼容性变化；固定版本，不加入防护排除项建议。

```text
go mod download all
go -C tools/devtools mod download all
```

  上述是人工/受批准准备步骤，不在离线验证器中执行。下载缺包必须保留原因，不能验收中偷偷联网。
- [ ] **2. 固定实现提交，依次真实验证。** 记录每条命令、提交、起止时间、退出码和输出摘要。
  根完整 mod verify 成功仍不能替代工具模块验证，反之亦然。

```powershell
powershell -NoProfile -File scripts/verify-devtools.ps1
powershell -NoProfile -File scripts/check-tools.ps1
powershell -NoProfile -File scripts/generate.ps1
git diff --exit-code -- api gen internal/store
```

```bash
bash scripts/verify-devtools.sh
bash scripts/check-tools.sh
bash scripts/generate.sh
git diff --exit-code -- api gen internal/store
```

  每个平台独立报告；生成变更不允许直接提交，先调查版本/路径差异。
  任一工具验证超时则整体工具链验收开放，不标记通过、不重复相同失败运行。
  可继续不依赖工具生成的 C12 专项测试，但只能单独报告其结果，不得宣布整体完成。
- [ ] **3. 普通与编译回归。**

```powershell
$env:GOOS='windows'
go test ./... -count=1 -timeout=15m
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
```

  编译成功明确标为 compile-only。完整 e2e/C11 运行沿用当前仓库入口和外层原时限，
  使用已有受控本地测试环境，不给任何既有测试增加预算。缺环境就记录未运行。
- [ ] **4. 独立执行 13 项物理 gate。** 使用原批准计划
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
  子进程/外层取消、全平台参数和未降低现有保护。审查缺陷修复需独立 RED/GREEN。
  无条件保持 main 不变；推送/合并遵循届时明确授权，不能把计划批准当合并批准。

## 自查及审批

设计第 1–3 节对应 Task 2 的模块/版本基线，第 4 节对应 Task 2 的分派和嵌套测试，
第 5 节对应 Task 1 的验证和 Task 2 的接入，第 6–8 节对应文件映射和 Task 3 验收交接。
五项 Review Focus 均已分配测试。Go verify 跳过完全未下载模块的实际行为已纳入 Task 1。
新 900 秒入口不扩展 C11 900/600 秒外层预算，这一潜在阻塞明确保留。

本计划待用户审阅并选择执行方式。推荐 Native：三个任务接口紧密，本会话顺序实现可减少
交接成本，完成后由新审查者做整体审查。也可选逐任务子代理实现/审查，成本较高但中间
审查更细。本次编写计划不启动实现、测试或代理。
