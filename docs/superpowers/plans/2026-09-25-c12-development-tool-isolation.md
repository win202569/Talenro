# C12 Development Tool Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将五个开发工具与 C12 主模块依赖分离，保留完整校验并重新取得真实验收证据。

**Architecture:** 根模块保留业务、测试、Goose 库及 CLI。独立 `tools/devtools` 模块保存五个开发工具，公开脚本先验证工具模块，再以显式 -modfile 调用工具。Windows 四个工具入口仅支持独立 PowerShell，以共享私有 Job 帮助文件覆盖整个入口；原生 Unix 保留 Bash。C12 runner 不修改。

**Tech Stack:** Go 1.26.0 / go1.26.5、Windows PowerShell、Bash、现有 Go 脚本契约测试、Docker Desktop。

**Spec:** `docs/superpowers/specs/2026-09-25-c12-development-tool-isolation-design.md`（用户已批准）。

**Approved addendum:** `docs/superpowers/specs/2026-09-25-c12-devtools-windows-powershell-addendum-design.md`（用户已确认；与原设计冲突时以附录为准）。

**Revision status:** 本修订待用户审阅。保留已选 Native inline 执行方式，不重新选择。原任务 1 未完成、任务 2/3 未开始；已有未提交实现只作为待核验工作，不能重记为完成。恢复时先读两份设计、当前 ledger 和本计划；本次修订不执行代码或测试。

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
- Windows 的 verify-devtools、check-tools、generate、verify-c11 仅支持独立 `powershell.exe -NoProfile -File`；公开入口不支持 dot-source。
- 四个 Windows `.sh` 入口在 Go/工具/验证器/Docker 启动前退出 1；不自动转调 PowerShell、不设绕过开关。Unix/WSL Bash 不回退到 Windows 工具，WSL 不替代 Windows C12 验收。
- PowerShell C11 原有 Bash smoke 子步骤保留；不是 Windows Bash 工具入口支持。无新原生监督程序、编码加载器或 C12 帮助函数复用。
- Defender 事件未解除：暂缓供应商提交不等于确认误报。含受阻 bootstrap 的全套普通测试、完整 C11 和 C12 物理 gate 暂不运行；仅继续不触发该路径的独立工作。
- Job 只保证本机自有进程边界，不承诺强杀后删除 Docker 服务端资源；既有资源协议不变。

## Review Focus

1. 缓存完全缺失时 Go mod verify 可能成功：Task 1 必须先证明依赖齐备且离线缺失失败。
2. 仓库路径含空格、Windows Bash 误调用及 Buf 子命令：Task 1/2 验证拒绝发生在工作前，受支持入口保留参数边界和根工作目录。
3. 工具构建使用备用模块，lint 子进程误用工具模块：Task 2 用根模块专属包验证分析对象。
4. 拆分触发间接版本降低或生成差异：Task 2 锁定迁移前选中版本，Task 3 真实生成比较。
5. 子进程自然退出与超时竞争、嵌套 Job、外层提前终止及脱敏：Task 1/2 验证整个 PowerShell 入口的自有进程退出、外部 canary 存活及初始化失败关闭。

## 文件职责与顺序

| 文件 | 职责 / 任务 |
| --- | --- |
| scripts/verify-devtools.ps1、scripts/verify-devtools.sh（新增） | Task 1，独立离线验证、期限和自有进程清理 |
| scripts/private/devtools-process.ps1（新增） | Task 1，函数定义式加载，私有 Job 初始化；Task 2 被另三个 PowerShell 入口复用 |
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

**Files:** 完成两份未提交 verify-devtools 脚本和 `internal/e2e/devtools_verification_test.go`；创建 `scripts/private/devtools-process.ps1`。不覆盖既有诊断报告或忽略目录探针。

**Interfaces:**
- Windows：`powershell.exe -NoProfile -File scripts/verify-devtools.ps1`；Unix：`bash scripts/verify-devtools.sh`。Windows Bash 只执行拒绝路径；无公开跳过、模块路径、命令或超时覆盖参数。
- 私有帮助文件只定义 `Initialize-DevtoolsProcessOwnership`（无参数、无成功输出，失败抛异常）；入口捕获后按自身脱敏格式退出 1。Job 句柄不返回给消费者，保持到进程退出；不提供释放接口。
- 受支持验证路径成功输出仅 `verify-devtools: passed`，退出 0；失败仅输出 `verify-devtools: <stage> failed with exit code <n>.`，不回显底层输出。Windows Bash 的不支持提示另按下文固定格式输出。
- Go 非零退出原样传播；校验语义错误/清理失败退出 1；命令缺失 127；超时 124。
- 不启动生成、lint、Docker；不得 dot-source 或执行 C12 runner 来借用函数。

- [ ] **1. 先保存历史失败，再补 RED 测试。** 文件使用 `//go:build e2e` 和 `package e2e`。
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

  Windows 下 `TestDevtoolsVerificationBash` 改测入口拒绝，原 parent-cancel / launcher-cancel
  的失败证据保留在 ledger，不称为修复。原生 Unix 分支保留完整性及进程测试，不能在函数
  开头无条件以非 Windows 为由 Skip；为 Unix 构造真实本地 Go/cache fixture，不使用 Windows 路径。
  拒绝测试先创建 fake Go/Docker 事件文件写入器，运行真实入口，断言如下（结果由 fixture 捕获）：

```go
if exitCode != 1 || !strings.Contains(output, "PowerShell") {
    t.Fatalf("expected unsupported-platform rejection: code=%d output=%q", exitCode, output)
}
if _, err := os.Stat(eventPath); !os.IsNotExist(err) {
    t.Fatalf("unsupported entry launched work: %v", err)
}
```

- [ ] **2. 跑 RED。**

```powershell
$env:GOOS='windows'
go test -tags=e2e ./internal/e2e -run '^TestDevtoolsVerification(PowerShell|Bash)$' -count=1 -timeout=5m
```

  预期为新平台拒绝/进程归属断言失败。Windows 分别测试 Git bin 包装器和 usr 实际解释器；
  可用的 MSYS/Cygwin 实际入口也执行同一拒绝断言，缺失环境列未实测。Unix 用本机 Go 1.26.5
  执行相同选择器（不设置 GOOS=windows）；缺少环境不能称跨平台通过。
  Go 测试固定 GOTOOLCHAIN=local、GOENV=off、GOPROXY=off、GOSUMDB=off。本机固定工具链为
  `C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe`；
  仅在已确认的沙箱标准库访问问题出现时申请限定测试命令提权，不改全局配置。

- [ ] **3. 实现验证命令链。** 完成入口平台拒绝和 PowerShell 归属初始化后，以脚本位置解析根目录及工具模块，先检查两个锁文件存在。
  固定环境使用进程级设置且 finally/trap 恢复：GOWORK=off、GOENV=off、GOTOOLCHAIN=local、
  GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly；
  继承的 GOFLAGS 中若含 -modfile/-overlay，先拒绝并退出 1；其余继承 GOFLAGS、
  GOEXPERIMENT、GOCACHEPROG 和模块/工具链覆盖项清除后使用上述固定设置。
  Windows 使用既有固定工具链目录，但不加载 C12 runner；Bash 原生使用已解析本地 Go，
  必须精确匹配 go1.26.5，拒绝 Windows 平台工具。Windows Bash 在此之前已经拒绝。
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

- [ ] **4a. 抽取私有 Job 初始化并先测试边界。** 从现有验证器抽取
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
  检查初始化互操作加载是否自身启动编译器：附录要求先有归属再有子进程，不能将 Add-Type
  的潜在编译子进程排除在外。若当前加载方案不能满足此顺序，停止并报告设计约束，不自行
  引入探针监督程序、编码加载器或降低合同。

- [ ] **4b. 实现进程预算与输出保护。** 单个入口从启动起只用一个 900 秒绝对截止时间，
  version、list、verify 共用剩余预算，不给每条命令另加 900 秒。
  PowerShell 以现有 verify-c11.ps1 的 `ConvertTo-C11CommandLineArgument`、
  `Get-C11CompletedProcessExitCode`、`Stop-C11OwnedProcessTree`、`Invoke-C11External`
  为已知模式，在新脚本内部建立私有实现；不改旧 C11 函数、不引入系统全局清理。
  Start-Process 使用 Hidden，读取 Process.Handle 后等待，超时停止自有进程树。
  原生 Unix Bash 使用有界执行和专属临时目录，失败必须清理自有 sink；不能把 timeout 或 trap
  的存在当作强制取消成功证据，必须实测入口终止后子孙退出。Windows 不再实现 Bash 监督。
  受支持入口提前终止时不得留下自有 Go/子进程；无法证明这一点则相应平台验收不通过。

```powershell
# 每次启动前从同一个截止时间计算；耗尽即失败，不自动续期。
$remaining = [Math]::Floor(($deadline - [DateTime]::UtcNow).TotalSeconds)
if ($remaining -le 0) { throw 'verify-devtools deadline exhausted' }
```

  输出写入自有随机临时目录中的 sink，校验阶段只解析需要的数据，不打印原始错误。
  受控捕获上限 4 MiB，超限失败；必须在命令仍运行时限制捕获，不能只在结束后检查文件大小。
  以持续写超限输出后阻塞的 fixture 断言提前失败、进程退出及不泄露内容；清理失败不能被成功状态覆盖。
  私有执行函数接受 Deadline 供测试传入短期限，公开入口固定 900 秒；不增加生产跳过参数。

- [ ] **5. GREEN 与独立负向确认。** 运行 Step 2，增加超时边界前自然退出、子进程存活 canary、
  父进程取消、stderr 敏感 canary、目录含空格和调用后环境不泄漏断言。
  移除齐备检查的测试副本必须使 missing-cache 负向测试失败，证明不会受 Go 的跳过语义欺骗。
  只在 fixture 中做变异，不修改真实脚本来求通过。
  Windows Bash 的最前部拒绝代码固定提示（其余三个入口由 Task 2 同步）：

```bash
case "${OSTYPE-}" in
  msys*|cygwin*|win32*) printf '%s\n' 'verify-devtools: Windows requires PowerShell.' >&2; exit 1 ;;
esac
devtools_platform=$(uname -s 2>/dev/null) || {
  printf '%s\n' 'verify-devtools: platform failed with exit code 1.' >&2
  exit 1
}
case "${devtools_platform}" in
  MINGW*|MSYS*|CYGWIN*) printf '%s\n' 'verify-devtools: Windows requires PowerShell.' >&2; exit 1 ;;
esac
```

  此检查放在 shebang 后、依赖解析/临时目录创建/验证前；平台识别失败也必须失败关闭。
  拒绝路径不包含绕过变量，不能伪造 uname 测试结果就声称实际 Cygwin 已验证。
- [ ] **6. 提交。**

```text
git add scripts/verify-devtools.ps1 scripts/verify-devtools.sh scripts/private/devtools-process.ps1 internal/e2e/devtools_verification_test.go
git commit -m "feat(tooling): add fail-closed offline development tool verification"
```

## Task 2: 原子拆分模块并更新全部调用链

**Files:** 文件映射中的模块文件、六份入口脚本、buf.gen.yaml、privacy_test.go、devtools_routing_test.go。

**Interfaces:** 消费 Task 1 的无参公开入口；新工具模块锁文件和显式 -modfile 调用是唯一新增分派合同。
  Goose 裸 go tool 命令保持。子工具和 lint 仍以根工作目录运行。
  三个 PowerShell 消费者还调用 Task 1 的 `Initialize-DevtoolsProcessOwnership`；公开验证器必须
  通过独立 powershell.exe 子进程调用，不使用 `& <verify-devtools.ps1>` 在同一进程运行第二次初始化。

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
  新增 `TestDevtoolsWindowsEntryRejection`、`TestDevtoolsConsumerOwnership`。前者遍历四个 `.sh`
  入口，复用 Task 1 的无工作事件/退出 1/PowerShell 提示断言；后者遍历三个 `.ps1` 消费者，
  让验证成功后 fake 工具创建长存子进程，强杀外层入口并检查全部自有句柄退出及外部 canary 存活。
  在 C11 fixture 中假造 Docker 响应，不创建真实容器；Job 测试不宣称 Docker 服务端清理成功。

```go
entries := []string{"verify-devtools.sh", "check-tools.sh", "generate.sh", "verify-c11.sh"}
consumers := []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"}
```

  旧 Windows Bash 工具入口隐私正向测试迁移到 Unix 分支；Windows 改为明确拒绝合同。
  不移除 Bash smoke 隐私用例，不全局 Skip Bash 测试。fixture 复制私有帮助文件，保留真实 Job
  初始化，不能替换为空操作以使取消测试通过。

- [ ] **3. 跑 RED。**

```powershell
go test -tags=e2e ./internal/e2e -run '^TestDevtools(ModuleBoundary|Routing|NestedProtoc|LintTargetsRoot|FailureStopsConsumers|WindowsEntryRejection|ConsumerOwnership)$' -count=1 -timeout=5m
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

- [ ] **5a. 同步入口支持与归属。** 三个 `.sh` 消费者复制 Task 1 最前部平台检查，提示分别为
  `check-tools: Windows requires PowerShell.`、`generate: Windows requires PowerShell.`、
  `verify-c11: Windows requires PowerShell.`；平台识别失败提示同样替换入口前缀。Unix 原有分派保留。
  三个 `.ps1` 在主流程第一个子进程前加载帮助文件并初始化；初始化 catch 按各入口现有
  internal 阶段脱敏输出、退出 1。C11 的 git 清单命令也必须在初始化之后。
  PowerShell C11 中 Get-C11BashExecutable 和 Bash smoke 调用保持，不删除或转换它们。

- [ ] **5b. 同步脚本调用。** generate/check-tools 的受支持平台在任何工具前调用对应 verify-devtools；
  C11 在工具阶段前也调用验证入口，保持原 check-tools 900 秒、generate 600 秒外层预算。
  外层剩余时间不足时允许明确失败，不用改预算或跳过标记规避重复验证。
  Windows 消费者独立调用验证器的参数固定为：

```powershell
$powerShell = Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe'
$verifyArguments = @('-NoProfile', '-NonInteractive', '-File', (Join-Path $PSScriptRoot 'verify-devtools.ps1'))
```

  用各入口现有受控执行函数传递这组独立参数，检查非零退出立即停止。check-tools 没有通用
  执行器时新增本地 `Invoke-DevtoolsVerification`（无参、无成功输出、失败调用 Stop-Script）；
  保留脱敏和退出码，不采用执行字符串、不新增 ExecutionPolicy 绕过。

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
# 仅在原生 Unix/WSL 环境运行；Windows 不执行这组正向验证。
bash scripts/verify-devtools.sh
bash scripts/check-tools.sh
bash scripts/generate.sh
git diff --exit-code -- api gen internal/store
```

  每个平台独立报告；生成变更不允许直接提交，先调查版本/路径差异。
  任一工具验证超时则整体工具链验收开放，不标记通过、不重复相同失败运行。
  当前安全事件阻塞仍在，不运行 C12 专项测试来补足工具链证据。Unix 无可用环境时记录未验收，
  不把 Windows Bash 的拒绝测试计为 Unix 成功。恢复手册列出四个入口支持矩阵和保留的 Bash smoke。
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
  子进程/外层取消、全平台参数和未降低现有保护。审查缺陷修复需独立 RED/GREEN。
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

本修订待用户审阅确认，继续保留 Native inline：本会话逐任务执行，最后进行一次独立整体审查。
不重新选择执行方式，不把本计划批准当作安全事件解除、测试通过或推送/合并授权。
本次编写计划不启动实现、测试或代理。
