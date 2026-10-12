# 2026-10-12 Nodecerttime 开发恢复检查点

## 授权与性质

用户在收到“归档必要证据、提交并普通推送现有开发分支，不合并main”的建议后回复“批准”。
这是本次保存阶段的新授权，不回写此前实施/审查阶段的无Git授权历史、
BASE/HEAD/index冻结记录或未验收状态。没有PR、强推、配置修改或清理授权。

保存分支：`codex/c12-b01-task9-coordinator`。
远端：[win202569/Talenro](https://github.com/win202569/Talenro)。
保存前HEAD和远端开发分支均为
`f5d731bef9d767d57ca32f0f4c3a0f532a4bc322`；
读取到的远端main为`e1571f9fb1940f341d8e40f55907bec5e4f30dcb`。
远端实际读取UTC见归档中的checkpoint-save-before.json。
实际提交SHA和推送结果以最终Git回执及聊天交付为准，
本文件不预填远端写入成功，也不能在自身提交中自引用提交SHA。

保存exact50必要文件：35新Go源码/测试、八处最小旧consumer guard、
批准规格/计划、执行记录/顶部项目状态（这四份既有增量文档），
加本恢复说明、ZIP和逐文件JSON清单。没有为备份修改产品或扩大消费者。
先读[当前状态](current-status.md)与[完整实施记录](2026-10-12-node-certificate-time-binding-execution.md)。

NodeCertificateConditionalTimeBound只是本次证书与调用者条件观测的组合。
全部原Proof/Pending/Obligations保留，不认证可信当前时间、原子平台采集、
CA来源/链/部署授权、selected identity、私钥持有、nonce原子消费、
持久检查点或运行权限。ValidUntil不是缓存/授权期限；
新观测需要重新Evaluate，浅拷贝Model共享Clear生命周期，不支持并发使用。
生产/provider兼容UNKNOWN，Task10/B01/C12/main阻塞仍保留。

## 新的提交前验证

074格式检查exact43路径：exit0/空输出，无整仓重写。
075–097原范围23条限定命令全部实际exit0：
19完整包、指定authority测试、store仅编译（no tests to run）、两短fuzz。
选定命令、timeout、缓存离线Windows Go1.26.5/PowerShell环境不变；
每条before/after377个nodecontrol非testdata Go源码摘要与已审查版本一致。
完整验证与每条实际UTC/耗时见checkpoint-verification.json和原始075–097回执。
377集合不是整仓，也不包含全部readiness/store源码。

此次Envelope实际40 executions、9初始缓存entries、无新增；
Pair5,037、4初始、1new interesting(total5)。
Go报告分别8.261s/7.996s；配置fuzztime=5s不等于整个进程elapsed，
有限短fuzz不证明穷尽。原实施110,263/8次结果与失败回执原样保留，
不替换为此次结果。不扩大为整仓/race/Linux/真实DB-Docker/provider/
hardware/reboot/PITR或C12权威验收。

没有产品源码修改或第二轮独立审查。
原fresh只读审查：0 Critical/0 Important/1 Minor。
M1仍延期：exported API注释需补充新观测重评估、非缓存/授权期限、
共享Clear失效及不支持并发Clear/read；实现记录已经明确这些使用边界。
测试是作者运行证据，不是reviewer动态重跑。

首次暂存差异检查在两份作者新文档发现末尾多余空行（实际exit2），
只进行末尾空行规范化，不修改Go源码、批准规格或原审查记录。
原失败保留为归档之后的checkpoint-stage-failure.json，未虚构在此前ZIP内；
暂存内容与最终检查结果仍以实际Git回执为准。

## 原始证据归档

- [395份本计划证据](checkpoints/2026-10-12-nodecerttime-evidence.zip)
- [逐文件字节长度和SHA256清单](checkpoints/2026-10-12-nodecerttime-evidence-manifest.json)

归档UTC：`2026-10-12T15:48:11.1295322Z`。
ZIP大小：5132156 bytes。
ZIP SHA256：`161EB51FAEBE06867FADDD61B6A825CE3586E38ED9C120CBB1A265FEBC99CB66`。
JSON清单SHA256：`8F5DCD83316CFB5ACE887532C57FB3C2151D0FCB19DBF6826BCD73E2F6091BA2`。
每个entry的实际长度/hash均与原始记录相符，并再检查源文件未在归档中改变；
不只依赖ZIP整体hash。条目按Ordinal排序，路径相对根，
含generator/的嵌套文件，但没有绝对路径、..或路径穿越/reparse文件。

包括：五task brief、批准计划归属、完整ledger/25项Ruling/M1、
真实reviewer原文/review package、所有原始001–073记录（含048格式）、
提交前074–097记录（含074格式）、started/完成回执、
固定Input/gzip/manifest及一次性构造器/独立checker、
原源码hash/审查和收尾核对、checkpoint授权/保存前Git快照及验证/归档/暂存脚本。
全部原始失败、测试/构造器/checker错误及路径笔误按真实结果保留，
不把命名带green的失败记录写成通过。

排除全部go-cache、依赖安装、本机凭据、数据库/Docker状态、
其他计划scratch及聊天历史。原电脑证据不删除。
高置信度凭据模式预检394份文本未发现命中，
不是形式化秘密检测证明；固定scalar/seed是测试材料，不是部署密钥。
归档是提交前快照，不含随后归档-result、暂存/提交/推送回执，
这些以实际Git结果和最终聊天交付为准，不能虚构在早先快照内。

证据脚本含原电脑绝对路径/历史固定BASE和index检查，
用于理解原始运行，不应在新电脑直接盲跑。缓存与依赖没有上传，
不承诺新电脑离线开箱运行；换行变化不能回填历史原字节hash。
Git提交包含必要项目状态，证据ZIP保留原始字节。

## 在另一台电脑恢复

在新的空目录位置使用PowerShell克隆，不覆盖已有工作区：

```powershell
git -c core.autocrlf=false clone --branch codex/c12-b01-task9-coordinator https://github.com/win202569/Talenro.git Talenro-nodecerttime-checkpoint
Set-Location -LiteralPath './Talenro-nodecerttime-checkpoint'
git log -1 --oneline
git status --short
```

将提交SHA与本次聊天的实际交付核对。
分支可能随后继续推进；精确历史恢复应在没有未保存修改的新克隆中
按交付SHA检出，不能把开发分支当main/生产版本。
进程参数core.autocrlf=false不修改原电脑全局配置。
原电脑core.autocrlf=true，暂存可能规范文本换行；
Git blob和原电脑raw文件摘要用途不同，历史证据不因换行重新生成。

如需恢复原始证据，先核对ZIP并确认目标不存在：

```powershell
$certTimeArchive='docs/roadmap/checkpoints/2026-10-12-nodecerttime-evidence.zip'
$certTimeExpectedArchiveSHA='161EB51FAEBE06867FADDD61B6A825CE3586E38ED9C120CBB1A265FEBC99CB66'
if ((Get-FileHash -LiteralPath $certTimeArchive -Algorithm SHA256).Hash -ne $certTimeExpectedArchiveSHA) {
    throw '证据ZIP校验不符，停止。'
}
$certTimeEvidenceTarget='.superpowers/sdd/2026-10-12-node-certificate-time-binding'
if (Test-Path -LiteralPath $certTimeEvidenceTarget) {
    throw '目标证据目录已存在，不覆盖。'
}
Expand-Archive -LiteralPath $certTimeArchive -DestinationPath $certTimeEvidenceTarget
```

然后按JSON清单核对全部entry的字节长度与SHA256。
恢复历史通过回执不代表新电脑已跑过测试，继续开发需按新环境重新验证。
保存阶段不合并main、不清理原分支/工作树，不改变任何既有验收阻塞。
