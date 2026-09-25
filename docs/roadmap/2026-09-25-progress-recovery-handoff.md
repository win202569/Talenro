# 换机接续说明 — 2026-09-25

本文件保存继续开发所需的决定、证据入口和未完成事项。它是未验收开发检查点，
不是全套测试通过、C12 完成或 main 合并记录。

## 从另一台电脑恢复

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
