# 2026-10-12 开发进度恢复检查点

## 性质与授权

用户在收到“下一步建议建立可迁移开发检查点；提交、推送需另行批准”后回答“按你的建议继续”。本阶段按此建立当前开发分支的提交和普通推送检查点，仅用于跨电脑恢复进度；不合并 main、不创建 PR、不强推、不删除分支或 worktree、不改变 Git 配置，也不解除 Task10/B01/C12/main 及生产验收阻塞。

保存分支：`codex/c12-b01-task9-coordinator`。
远端：[win202569/Talenro](https://github.com/win202569/Talenro)。
开始 HEAD：`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`，远端开发分支相同；开始远端 main 为 `e1571f9fb1940f341d8e40f55907bec5e4f30dcb`。提交 SHA 和推送是否完成，以最终 Git 回执及本次聊天交付为准；此文档不预先证明远端写入成功。

此前 spec/plan/execution 中的“无 Git 授权、未提交、仅本机交付”是各实施阶段的真实历史记录，本阶段的新授权只改变保存方式，不回写或冒充当时已提交。历史独立审查的基线、索引和 HEAD 仍是历史值，不被新检查点替换。

## 保存和保留内容

原工作区365项待保存内容均为本项目的源码、测试、测试向量、SQL/生成代码、设计、计划及执行记录。本检查点保存这些相互依赖的增量，而不只保存最近25项 CA 文件；没有修改产品实现或旧测试材料来整理检查点。恢复开发进度不表示这些历史增量已获得一次整分支独立审查或生产验收。

最新 CA 包24源码/测试仍匹配 freeze8。CA审查8为0 Critical/0 Important/0 Minor，证明范围仅是部分材料形状和声明subject公钥绑定，原六项Pending不变。先读[当前项目状态](current-status.md)与[CA执行记录](2026-10-11-node-ca-material-execution.md)。

最新 CA 原始证据另归档于：

- [162份原始文本证据归档](checkpoints/2026-10-12-node-ca-material-evidence.zip)
- [逐文件字节长度与SHA256清单](checkpoints/2026-10-12-node-ca-material-evidence-manifest.json)
- [本次提交前重新验证完整回执](checkpoints/2026-10-12-verification.json)

归档大小436021 bytes，SHA256=`A9A3B547422F397951D2EFE23494AC5D6ED2A6BEC95574820FF62F5859805CEA`；逐条解码内容核对与原文件一致。保留104条attempt路径，其中103为完整原始回执，0094仍是输出丢失诊断，退出码/时间/哈希/结果未知，不能重构为RED或GREEN。保留八版冻结、八份审查记录、原始brief、基线、进度和最终保全记录；报告摘要/作者转录身份不改。排除了五份过时执行草稿及所有Go缓存，原本机文件不删除。

更早各增量的源码、向量、设计和正式执行记录已随工作区保存；它们的全部Git忽略原始scratch记录没有一并迁移，仍在原电脑。缓存、依赖安装、凭据、数据库状态、Docker环境和聊天历史不属于Git检查点，也不保证冷机器离线开箱运行。内联测试密钥来自固定测试种子，绝不是部署密钥；高置信度凭据模式检查未发现命中，不等于形式化秘密检测证明。

## 本次新验证和已知欠账

使用缓存的原生Windows Go1.26.5、PowerShell、模块只读及禁下载环境。继承GODEBUG为空，没有写全局Go或Git配置。完整wrapper、实际UTC、命令、输出、退出码均见验证回执。

- 17个完整包（16个nodecontrol包加readiness）全部exit0。
- authority仅`^(TestBoundRead|TestCoordinator)`指定测试exit0，未运行integration构建标签。
- store仅`-run '^$'`编译exit0，无store测试执行。
- 两项CA短时fuzz均exit0；DER仅6次执行，SPKI53,788次执行、11项新增覆盖样本，未报告crash。不代表穷尽测试。
- 762个当时非忽略文件的前后清单SHA256相同，CA24文件与freeze8前后均无漂移。
- 更广的待保存Go文件格式检查**未全通过**：唯一输出为`internal/nodecontrol/stateverify/testdata/generate_vectors.go`。该旧测试专用生成脚本带`//go:build ignore`，位于CA实施前基线中且哈希未变；本次只读差异检查确认格式未规范，保留原样。这不是CA新文件格式失败，也不把整体格式状态改称通过。
- 暂存后`git diff --cached --check`退出2：历史SQL/文档/测试README末尾空行，以及部分历史Markdown/原错误回执行尾空格（其中Markdown双空格用于换行）仍在。完整提示留于原电脑检查点准备记录；不声称该检查通过，也不为了备份改写旧需求/证据。369项暂存清单没有漏项或额外文件，CA24个暂存blob与原始源文件逐一相同。此为未完成整体验收的开发快照，不据此合并main。
- 未运行全仓库/race/Linux、真实DB/Docker、硬件/重启、供应商生产兼容或C12/PITR验收。测试命令通过不解除那些门槛。

## 在另一台电脑恢复

用PowerShell在一个新的空目录位置克隆开发分支，不覆盖现有工作区：

```powershell
git -c core.autocrlf=false clone --branch codex/c12-b01-task9-coordinator https://github.com/win202569/Talenro.git Talenro-checkpoint
Set-Location -LiteralPath './Talenro-checkpoint'
git log -1 --oneline
git status --short
```

将显示的提交SHA与本次聊天最终交付的SHA核对。分支后续可能推进；如果需要精确历史检查点，在没有未保存修改的新克隆中按交付SHA检出。不要将本开发分支当作main或直接进入生产部署。

clone前的`-c core.autocrlf=false`只是该命令的进程选项，不修改原电脑配置；它使首次Windows检出尽量保留Git内LF。原电脑core.autocrlf=true，而CA源文件原始为LF；后续电脑自己的Git设置仍可能在检出时转换换行。历史SHA256记录的是当时原始字节，换行改变会导致哈希不同；不要改写证据来掩盖差异。

如需继续核对CA原始回执，可解压归档到原相对位置。存在旧证据目录时先停止，避免覆盖：

```powershell
$restoreEvidencePath='.superpowers/sdd/2026-10-11-node-ca-material'
if (Test-Path -LiteralPath $restoreEvidencePath) {
    throw '证据目录已存在，请先核对归属，不覆盖。'
}
Expand-Archive -LiteralPath 'docs/roadmap/checkpoints/2026-10-12-node-ca-material-evidence.zip' -DestinationPath $restoreEvidencePath
```

JSON清单记录每个解压文件的SHA256和字节长度；无需运行其中的历史命令。历史命令包含原电脑绝对路径，换电脑运行前需按项目环境重新配置Go1.26.5、模块依赖及缓存路径。离线缓存没有上传，历史回执也不能被视为新电脑已跑过测试。

安全的下一开发步骤是从当前状态中的未决门槛制定范围明确的计划，而不是自动合并main或降低Pending。
