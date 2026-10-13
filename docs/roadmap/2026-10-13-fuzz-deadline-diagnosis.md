# Fuzz deadline diagnosis — bounded evidence, not a fix

日期：2026-10-13。基线：c75b66a900fe554e272b15737ab19f8b9c00c2c4，
分支codex/c12-b01-task9-coordinator。状态：发现并观察到框架退出竞态候选；
历史失败的精确归因未完全确定，未实施修复，验收阻塞保持开放。

## 1. 问题和本次范围

[上一检查点](2026-10-13-declared-issuer-mapping-consistency-execution.md)
提交前matrix3：前24项通过，第25项FuzzIssuerEnvelope报告context deadline exceeded，
12次执行、6.02s、包时间8.844s；Pair未运行。该失败已随c75b66a9保存，不抹掉。

本次仅源码追踪、真实context对照探针，以及一次加入框架诊断日志的原预算fuzz运行。
无产品、测试、SQL/module、fixture或固定Go工具链改动；未增加时限、改变并行度、
删除/复制旧私有缓存或调整安全设置。没有运行完整矩阵、Pair或真实环境验收。

独立证据scope：.superpowers/diagnostics/2026-10-13-fuzz-deadline/。
001/002保存程序、参数、环境、UTC、exit、完整stdout/stderr；
仅本机原始日志不进入Git。可恢复结论及必要探针在本报告和tracked文本中。

## 2. 固定源码追踪

Go实际版本：go1.26.5 windows/amd64；没有推断其他版本或上游修复状态。
下列路径相对该固定工具链src，SHA256为本机实际字节：

- internal/fuzz/fuzz.go：
  D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5
- context/context.go：
  765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7
- 项目internal/nodecontrol/nodeissuer/fuzz_test.go：
  318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E

源码关系：

1. CoordinateFuzzing在fuzz.go:107建立总超时ctx，在112建立fuzzCtx子上下文；
   228总ctx.Done分支调用stop(ctx.Err())。
2. stop在fuzz.go:130附近只按err==fuzzCtx.Err()或interrupt判断抑制；
   非nil错误可能进入fuzzErr，并在工作进程结束后返回。
3. context.go:549的取消先设置父err、关闭父Done（567），随后遍历并取消子
   上下文（569–571）。父/子观察并不保证同一时刻完成。
4. context.go:463的Err使用atomic load，只保证自己的Done已关闭；
   它不是“所有子上下文已经同步”的屏障。
5. 因此可存在parent.Err()==DeadlineExceeded且child.Err()==nil的短窗口。
   若stop在该窗口读取这两个值，已到期的正常总预算可能被记录为错误。
   这是从固定源码和下面观察得到的机制解释，不代表已经捕获历史原事件的全部状态。

Envelope回调只调用parseLeafEnvelope/checkSignatureDER并比较材料/错误身份；
回调未创建计时器或context，也未自行生成context deadline exceeded。
DER/signature路径经有限源码核对为有界解析，未据此宣称所有输入均无问题。
worker RPC/minimization/cleanup也阅读过，不能仅凭错误文字排除所有其他路径。

## 3. 真实context对照探针（001）

[可移植探针](../diagnostics/2026-10-13-fuzz-deadline-probe.go.txt)为标准库main程序。
实际执行的probe.go保留byte-identical归档probe.go.txt，
SHA256：87EADC3555905739423ACA090183C2E4287C4CA3097347454DBF27C922D1084F。
tracked探针仅去掉EOF多余空行，算法/语句未改；按LF/EOF规范化比较相同。
Git配置可能在新电脑checkout时转换CRLF，因此不能把原机器字节hash
当跨机器文件字节完全相同的证明；原运行档案不改写。

每个父context使用100µs超时；每组最多4096次，整个探针5s检查边界。
比较0/32个额外子上下文；本机GOMAXPROCS16，实际Go版本go1.26.5。

- 0额外子：4096次，捕获父已截止/直接子Err仍nil共3次。
- 32额外子：3106次，捕获1次，然后达到总检查边界。
- 两组捕获值套入现有stop错误比较分支，分别会保留3/1个DeadlineExceeded。

观察使用真实context，不是替代项目lower；该分类步骤是对源码表达式的重演，
**没有直接调用实际CoordinateFuzzing.stop**，所以不伪称完整fuzzer错误已复现。
实验数量不是固定预期：调度和平台不同可能出现0窗口；程序exit0不等于捕获成功。
5s是迭代检查预算，不承诺硬实时退出/OOM/crash保证。不存在工具链补丁。

## 4. 框架日志采集（002）

固定原命令：
`go test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$'
-fuzztime=5s -parallel=1 -p=1 -timeout=30s`。

只增加本次进程GODEBUG=fuzzdebug=1；这是固定源码提供的诊断日志，
不是永久环境/安全设置变更。GOOSwindows/GOARCHamd64、GOTOOLCHAINlocal、
GOPROXY/GOSUMDB/GOAUTHoff、GOVCSall:off、GOFLAGS-mod=readonly；
使用本诊断独立go-cache和既有只读模块缓存。未读取或复制上一增量私有缓存。

实际：

- 新独立cache初始5个种子，全部warmup完成；原失败cache曾有10项，状态不同。
- 71486次执行，新增6个interesting条目，最终11项；exit0，包时间7.840s。
- 诊断日志先记录fuzz.go:228的stop，stopping:false（总预算截止）；
  之后记录232的stop，stopping:true（worker结束）；没有crasher记录。
- 原命令实际仍只有5s fuzz预算和30s测试上限；
  命令wall约148.49s包含新缓存冷构建，不是放宽fuzz预算。
- 日志时钟来自固定Go time.Now格式，不附时区；receipt UTC独立记录。
- 日志会扰动调度且cache/输入状态不同；这次通过不能证明原竞态已修复，
  更不能补成原matrix3第25通过/第26已运行或整个矩阵已验收。

## 5. 判断、未验收项与下一步

已证实：在当前真实Go context中可以观察到父/子截止传播窗口；
固定fuzz stop判断在该观察值下会保留截止错误。诊断fuzz确实走总截止退出路径。

较强候选：历史裸DeadlineExceeded可能来自这个正常退出被误分类的窗口。
未证实：原失败那次stop读取的实际父/子状态；本次日志没有重现同样失败。
没有证据允许声称“材料解析缺陷不存在”“系统/安全事件已清除”或“错误已修复”。

下一步应先明确并批准限界处理方案：怎样在固定原验收门槛下建立完整复现和
验证，而不把正常退出误判，也不吞掉真正材料/worker错误。不得直接修改固定Go
模块缓存、安装新工具链、改成按次数验收、增加超时或只反复重跑直到偶然通过。
任何工具链/验收合同调整都需要单独设计、授权及可恢复的对应回归证据。

全部701 Go/SQL/module文件前后未变；
before-source.json与after-source.json摘要均为
9DB4FC99B644493C45EB7C0F30D59840DBA97F445EDA0C08BEAAE2A090D41B58。
所有旧spec/产品/fixture和未验收范围保持；仅必要诊断文本与状态文档保存。
新电脑可获取探针和结论，但不能重建旧私有cache、历史输入序列或机器调度。

## 6. 新电脑复核入口（不是自动执行授权）

先读AGENTS.md和repository-recovery.md，独立准备已批准的固定Go1.26.5工具链。
不安装替代版本。将tracked .go.txt复制到新的临时目录作为probe.go，
为本次复核指定独立GOCACHE，在原模块之外也可运行该标准库程序。
保留完整输出和实际GOMAXPROCS/版本；禁止用观察计数不同来推断修复或退化。

需要fuzz日志时，仍用上面原5s/30s预算，单独记录GODEBUG=fuzzdebug=1和
新cache种子数/输入差异。不把诊断日志模式的通过当正常验收；保留历史失败。
