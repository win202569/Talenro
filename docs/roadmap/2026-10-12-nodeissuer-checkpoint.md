# 2026-10-12 Nodeissuer 开发恢复检查点

## 授权与性质

用户在收到“归档必要证据，暂存、提交并普通推送当前开发分支；不合并main、不创建PR”后回复“批准”。
本保存阶段以此新授权执行，不回写此前实施阶段的无Git授权历史，
不改变原独立审查基线/HEAD/index，也不授权强推、配置修改、清理或provider迁移。

保存分支：codex/c12-b01-task9-coordinator。
远端：[win202569/Talenro](https://github.com/win202569/Talenro)。
保存前HEAD：764dcb8b6b682ecce03fd6de02dbe339c66ff985；
远端开发分支：2877f89cc9c5a6ca0514d79c54826627a112d412；
远端main：e1571f9fb1940f341d8e40f55907bec5e4f30dcb。
远端读取UTC：2026-10-12T12:53:14.4379383Z。
本提交也将通过普通fast-forward推送保存此前尚未推送的直接issuer设计提交。
实际新提交SHA及推送成功与否以Git回执和本次聊天最终交付为准；
本文件不预先声称远端写入成功，也无法在自身提交内容中自引用提交SHA。

本次保存33个文件：24新nodeissuer Go源码/测试、三个lower最小消费guard、
计划、执行记录、顶部状态更新，以及本恢复说明/ZIP归档/JSON清单。
不修改生产代码来整理备份，不扩大消费者或删除Pending。
先读[当前状态](current-status.md)和[完整执行记录](2026-10-12-node-leaf-issuer-relation-execution.md)。
直接数学验签Proof不是CA来源、信任链、当前有效性、部署授权、私钥持有或运行证明；
Task10/B01/C12/main及既有整体验收阻塞保持原样，生产/provider兼容仍UNKNOWN。

## 新的提交前验证

058只读exact27文件格式检查exit0、空输出。
059–080重新运行批准的22条限定命令，全部exit0，
每条命令前后27个长度/hash均匹配原reviewed freeze1：
18个完整包、指定authority、store仅编译（无store测试）、两个短fuzz。
authority未启用integration tag；未运行整仓/race/Linux/真实DB/Docker/hardware/reboot/PITR。
离线Windows Go1.26.5、PowerShell、模块只读和禁下载环境不变。

本次Envelope fuzz实际44 executions，13个初始缓存corpus entries、无新增；
Pair实际6,911 executions，5个初始缓存entries，1 new interesting、total6。
配置fuzztime=5s不等于总elapsed；有限数据不能声称穷尽。
原实施阶段75,755 /6,917次结果仍以原032–053回执保留，没有替换成此次结果。
没有产品源码修改或第二轮独立审查；原fresh审查仍为0 Critical/0 Important/0 Minor。

## 原始证据归档

- [103份本增量原始记录](checkpoints/2026-10-12-nodeissuer-evidence.zip)
- [逐文件字节长度与SHA256清单](checkpoints/2026-10-12-nodeissuer-evidence-manifest.json)

归档UTC：2026-10-12T13:03:19.1438410Z。
ZIP大小：339823 bytes。
ZIP SHA256：8EEC11CA2E21B9CF2D363C040F094382FB3B6815590E25721F327A1D5B3D828E。
103个entry都与源记录逐字节长度/hash一致；不依赖仅ZIP整体hash判定内容。
每个entry是根目录文件名，没有路径穿越或嵌套目录。

包括：766路径原基线、五task brief、完整progress/Ruling、原始RED/GREEN/诊断、
全部编号001–080路径、freeze1、fixture manifest/generator、23个本次验证回执
（1格式+22 Go）、真实reviewer原文、旧guard净diff、保存前Git/远端只读快照、
校验包装器和归档脚本。80个编号路径中有诊断/快照，不把它们一律称为完整进程回执。
008/030/031缺少完整进程时间/exit的限制和初始write-tree对象探测误用仍保留。
第一次归档排序因Object[]与generic comparer参数不兼容失败，
发生在ZIP创建之前；只读复现、typed string[]/Ordinal对照和真实失败说明均保存。
这一归档步骤问题未改变源码、测试回执或独立审查。

排除全部go-cache目录、依赖安装、凭据、数据库/Docker状态与聊天历史。
较早增量的全部原始scratch未新增打包；此前CA证据归档已在父提交中。
原电脑记录不删除。归档是提交前快照，不含随后暂存/提交/推送回执，
这些后续操作以Git实际结果和聊天交付为准，不能虚构在早先快照内。

高置信度凭据模式预检未发现命中，不是形式化秘密检测证明。
测试用scalar1/scalar2及RFC向量不是部署私钥。恢复证据不需要运行其中脚本；
原命令含原电脑绝对路径，在新电脑运行前须重新配置工具与依赖。

## 在另一台电脑恢复

使用PowerShell，在新的空目录位置克隆；不覆盖现有工作区：

```powershell
git -c core.autocrlf=false clone --branch codex/c12-b01-task9-coordinator https://github.com/win202569/Talenro.git Talenro-nodeissuer-checkpoint
Set-Location -LiteralPath './Talenro-nodeissuer-checkpoint'
git log -1 --oneline
git status --short
```

将提交SHA与本次最终聊天交付核对。分支可能继续推进，
精确历史恢复应在没有未保存修改的新克隆中按交付SHA检出，
不能把开发分支当main或生产验收结果。

clone进程的core.autocrlf=false不修改原电脑全局配置；
用于首次检出保持Git blob换行。原电脑core.autocrlf=true，
暂存文档可能规范CRLF为LF。证据ZIP保留原始回执字节；
SHA256记录不能随换行变化回填或改写。
缓存与依赖没有上传，不保证新电脑离线开箱运行。

需要恢复原始证据时，先确认目标目录不存在：

```powershell
$nodeissuerEvidenceTarget='.superpowers/sdd/2026-10-12-node-leaf-issuer-relation'
if (Test-Path -LiteralPath $nodeissuerEvidenceTarget) {
    throw '证据目录已存在，请先核对归属，不覆盖。'
}
Expand-Archive -LiteralPath 'docs/roadmap/checkpoints/2026-10-12-nodeissuer-evidence.zip' -DestinationPath $nodeissuerEvidenceTarget
```

随后按JSON清单核对每个entry字节长度和SHA256。
历史通过回执不代表新电脑已跑过测试；继续开发仍需针对新环境重新验证。
