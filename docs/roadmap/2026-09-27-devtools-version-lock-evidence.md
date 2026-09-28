# 开发工具锁版本修订证据

状态：实施中，未验收。计划 R1–R8 与 R3 精确候选准备已获批准；不授权推送、合并或恢复安全事件阻塞的测试。

## 不可变基线

可提交数据：`internal/e2e/testdata/devtools-version-lock/baseline.json`，schemaVersion 1。
来源为拆分前成功的严格离线查询，不使用带 -e 的诊断作为基线。
保存 666 条模块记录；八组 Windows 包映射分别为 root 516、integration 523、Goose 656、Buf 858、protoc 137、oapi 279、sqlc 575、lint 1187。原记录 Error/Incomplete/DepsErrors 均为空。
包只保留 ImportPath、模块路径与版本；模块只保留 Path/Version/Main。54 个原文件哈希中 2 个为根 go.mod/go.sum，其他 52 个为受保护源码、runner 或生成文件。没有缓存目录或本机绝对路径。
此为历史静态基线，不证明本次代码或第三方测试通过。

## 本次恢复前检查点

工作分支 codex/c12-b01-task9-coordinator，HEAD a5b76f30a1d2819d03caea9b03b3953e0f3e14d9；原 Task 2 BASE 3a43a724de801bf24abdede71d806559016cb602。锁文件包含前轮未提交实现，以下哈希不替代不可变基线。

| 文件 | 恢复前 SHA-256 |
| --- | --- |
| go.mod | A0883514AF23F6947A8CF90F52B20D56AF5FC51BCECEEC2DD7C98082DF59E9E0 |
| go.sum | 642D51BD341B270B35EC976F6196152A33057ED38E54FBCEBF143A50271733D7 |
| tools/devtools/go.mod | 0A8312541FEC1A9FDC6555DB05D1F1D1342EE523B3273A800347714C146A71B9 |
| tools/devtools/go.sum | 85A30247BB0F2A8568D8081A31F029CFA8FEA89C9F2CFE3FF17EEDD2A0340164 |

## 当前门槛

2026-09-28 最新：32 Required、35 个条件候选及归属证据已经接入实际检查，版本专项退出 0（7.032 秒）。两模块严格完整查询、八组实时包比较和证据绑定通过；各所属目录离线 tidy 后四份锁文件字节不变，52 个受保护文件哈希不变。
根模块专属 lint 样例先因缺少文件失败（17.058 秒），补入后正常目录与错误目录反例一起通过（32.049 秒）。元数据由真实 Go 读取，外部 lint 本身仍为测试替身，不据此声称真实 lint 通过。
Linux 固定本地镜像禁网/只读测试退出 0（测试包 0.916 秒），四个入口拒绝行为通过；这不是 Unix 正向支持。
Windows 完整专项退出 1（571.336 秒，原 10 分钟预算未超限）：1029 个 PASS 标记、2 个 FAIL 标记（同一失败子用例及其父测试）、1 个既有不适用 Skip。唯一失败子用例为 TestDevtoolsConsumerOwnership/verify-c11.ps1：12 秒内未观察到自有工具就绪，尚未走到本次取消清理断言。新增版本测试与真实 Lock 在该组合运行均通过，Lock 耗时 18.77 秒。不能用这些局部绿色抵消整体失败。
随后仅增加测试诊断：超时先停止自有父进程、等待输出写入结束再读取日志，避免数据竞争；就绪时记录耗时。未改 12 秒就绪、20 秒外层或清理时限。单独诊断运行退出 0（19.972 秒），就绪耗时 10.2137481 秒。单次重跑未复现，不能据此认定已修复，也不能把原因归咎于 Defender 或并行负载；组合失败仍开放，诊断编辑尚未经过完整组合复验。
上述为前轮停止点。最新关闭插桩的完整组合复验退出 0（399.344 秒，1031 个 PASS 标记、0 FAIL、1 个既有不适用 Skip）；三个消费者就绪分别为 3.384、3.453、6.514 秒，取消与无关进程断言全部通过。原 10 分钟总预算、12 秒就绪及其他限制未调整，本代理未并行启动其他测试/依赖查询。
本次完整复验通过，不等于历史失败根因已确定或已修复，也不证明冷态/负载下稳定。当时 Task 2 尚未原子提交及 task-done。

最新完成记录：Task 2 已在原选择器、关闭插桩及原 10 分钟上限下再次通过（541.597 秒，1031 PASS 标记、0 FAIL、1 既有不适用 Skip），随后原子提交为 `a0ca18928f552752c6580929f9b6fff88cd8d58b`，task-done 已记录。三个消费者就绪为 4.761、5.682、6.459 秒。本地提交不是推送或合并；Task 3 真实工具/生成/lint 与独立审查仍须各自记录结果。
历史 256 PASS / 0 FAIL / 1 Skip 及 Linux 50 PASS 不替代新规则验证。
完整普通测试、完整 C11、C12/bootstrap 与 I2/I3 保持原阻塞；未验证平台和第三方测试不声称通过。

## R2：比较器的阶段性证据

先运行无条件放行的比较器，14 个拒绝场景全部按预期失败；3 个合法场景通过。沙箱初次无法读取 archive/zip 标准库，只记为环境失败。经受控授权的同一离线选择器实际运行，RED 包耗时 18.977 秒；两模块严格查询分别退出 1，不能作为版本 RED 或完整图成功。
实现比较器后基础 17 场景通过，包耗时 4.275 秒。补入 35 项历史诊断差异逐项拒绝回放和三个精确候选/自然消失场景后，19 个子场景通过，包耗时 6.243 秒。诊断记录仅用于拒绝回放，绝不充当完整查询证据。
policy.json 的模块归属仍是 pending，不从实际图自动生成“已审核”结论。真实 Lock 仍不通过，证据记录存在性校验等剩余合同尚未收尾；R2 不标记完整完成。

## R3：获准的精确准备

三个精确版本均缺少完整缓存，已通过 proxy.golang.org 和 sum.golang.org 准备，退出 0；没有执行下载内容，没有修改四份生产锁文件。下载并不证明例外必要或已验收。

| 版本 | Sum | GoModSum |
| --- | --- | --- |
| github.com/chzyer/readline v1.5.1 | h1:upd/6fQk4src78LMRzh5vItIt361/o4uq553V8B5sGI= | h1:Eh+b79XXUwfKfcPLepksvw2tcLE/Ct21YObkaSkeBlk= |
| github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b | h1:ogbOPx86mIhFy764gGkqnkFC8m5PJA7sPzlk9ppLVQA= | h1:gx7rwoVhcfuVKG5uya9Hs3Sxj7EIvldVofAWIUtGouw= |
| go.opentelemetry.io/otel/metric/x v0.66.0 | h1:YkCrx1zLOChi9ZcZ6euupOcsgzbVlec7D/xoEU1+cTA= | h1:d1+BDj9t96do0/1LoU1ayfCv79ZgNE41qbhBvnMOBZk= |

## R4：四项工具约束不能经 tidy 稳定保留

已核对原图及被保留父模块的缓存清单：

| 工具模块目标 | 拆分后父边声明 | 原较高版本的来源 | 本次精确约束整理后 |
| --- | --- | --- | --- |
| golang.org/x/time v0.14.0 | moby/moby/client v0.5.0 → v0.11.0 | oapi-codegen/runtime v1.6.0 → v0.14.0 | require 被删除 |
| modernc.org/mathutil v1.7.1 | pingcap/tidb/pkg/parser 固定版本 → v1.6.0 | Goose/libc/memory/sqlite → v1.7.1 | require 被删除 |
| modernc.org/sortutil v1.2.1 | 同一 parser → v1.2.0 | libc v1.72.1 → v1.2.1 | require 被删除 |
| modernc.org/strutil v1.2.1 | 同一 parser → v1.2.0 | libc v1.72.1 → v1.2.1 | require 被删除 |

parser 的精确版本为 v0.0.0-20260418072757-ce92298d1124。四者在原五工具各自 Windows 实际包映射中均为零包；这只说明当前包基线，不证明第三方测试或其他平台无影响。
先尝试的临时诊断脚本错误地在根 cwd 使用工具 modfile 做 tidy，导致根源码被解释为 talenro.local/devtools 并查询本项目路径，退出 1 / 30.626 秒；这是执行脚本错误，不是依赖方案证据。输出保留，不作通过。
修正为在 tools/devtools 实际目录运行固定 Go 的禁网 tidy，退出 0 / 11.276 秒。四个新 require 全部被移除。四份锁文件 SHA-256 与上面的恢复前检查点逐项完全相同，证明恢复未生效，没有留下本轮的模块变化。

按设计第 4/8 节停止这条恢复路径：没有再次补写约束、跳过 tidy、加入虚构 import/tool、回挂根专属依赖、增加 replace/exclude 或扩大例外。根 28 项恢复、R5–R8 与 Task 3 未继续。当前证据只排除了本次显式 require 方案，不声称所有可能方案已被证明不可行。
下一步需要决定是否另行评估四个图级依赖的规则修订，或继续寻找符合原合同的稳定来源；未经新的明确决定，不接受这四项降级、不改变既有 32/3 策略。

## 2026-09-28 获批接续：四候选准备与根恢复阻塞

以上是历史停止点。用户随后批准四项 tools 条件例外设计、对应计划修订及 R3b 精确准备范围；32 个目标保留，固定例外上限变为 root 3 / tools 4，不扩展到其他路径或版本。

新增策略、证据绑定、包用途单元测试先见 RED（退出 1，0.513 秒），实现后本次离线选择器退出 0，0.525 秒，76 个 PASS 标记、0 FAIL。该计数包含顶层测试；仅证明已覆盖的纯数据场景，不证明真实图、证据正文或八组实际包已通过。真实 Lock 接线、成员归属审核、包映射复验和其他专项仍未完成。

四个版本均缺缓存，准备时间 UTC 2026-09-28 11:14:32.453–11:14:44.731，整批约 12.278 秒；固定 Go，仅官方 proxy.golang.org 与 sum.golang.org，无 direct/VCS/auth 回退。四项退出码均 0，未执行下载内容，四份生产锁文件哈希未变。

| 精确版本 | Sum | GoModSum |
| --- | --- | --- |
| golang.org/x/time v0.11.0 | h1:/bpjEDfN9tkoN/ryeYHnv5hcMlc8ncjMcM4XBk5NWV0= | h1:CDIdPxbZBQxdj6cxyCIdrNogrJKMJ7pr37NYpMcMDSg= |
| modernc.org/mathutil v1.6.0 | h1:fRe9+AmYlaej+64JsEEhoWuAYBkOtQiMEU7n/XgfYi4= | h1:Ui5Q9q1TR2gFm0AQRqQUaBWFLAhQpCwNcuhBOSedWPo= |
| modernc.org/sortutil v1.2.0 | h1:jQiD3PfS2REGJNzNCMMaLSp/wdMNieTbKX920Cqdgqc= | h1:TKU2s7kJMf1AE84OoiGppNHJwvB753OYfNl2WRb++Ss= |
| modernc.org/strutil v1.2.0 | h1:agBi9dp1I+eOnxXeiZawM8F4LawKv4NzGWSaLfyeNZA= | h1:/mdcBmfOibveCTBxUl5B5l6W+TTH1FXPLHZE6bTosX0= |

准备后恢复禁网，根 cwd 配合显式 `-modfile=tools/devtools/go.mod` 的严格 `list -m -json all` 退出 0 / 4.144 秒，`mod graph` 退出 0 / 0.095 秒，无 `-e`。活动图确认 moby/client v0.5.0 → x/time v0.11.0，以及固定 parser → mathutil v1.6.0 / sortutil v1.2.0 / strutil v1.2.0。仍不据此标记候选完整验收。

根严格 `mod graph` 退出 0 / 0.504 秒，保留的 grpc v1.80.0 声明 cel.dev/expr v0.25.1。先对这一普通目标做单项最小约束实验：仅新增 root 间接 require cel.dev/expr v0.25.2，未改变 grpc、其他上游或工具归属。在真实根目录正常禁网 tidy 退出 0 / 3.280 秒后，该 require 被删除。重新严格 graph 退出 0 / 0.079 秒，仍是 grpc v1.80.0 → cel.dev/expr v0.25.1。

实验后四锁 SHA-256 与实验前逐项相同：

- go.mod：A0883514AF23F6947A8CF90F52B20D56AF5FC51BCECEEC2DD7C98082DF59E9E0
- go.sum：642D51BD341B270B35EC976F6196152A33057ED38E54FBCEBF143A50271733D7
- tools/devtools/go.mod：0A8312541FEC1A9FDC6555DB05D1F1D1342EE523B3273A800347714C146A71B9
- tools/devtools/go.sum：85A30247BB0F2A8568D8081A31F029CFA8FEA89C9F2CFE3FF17EEDD2A0340164

按 R4 和原设计第 4/8 节停止后续恢复。证据仅排除该单项显式约束方案，不证明所有稳定方案均不可能。不反复补写 require、不造假 import、不改缓存、不新增 replace/exclude，不将 cel 自动扩展为第八项例外。其余 root 目标、R5–R8、Task 3 尚未完成；没有暂存、提交、推送、合并或运行受阻的完整普通/C11/C12 测试。下一步需选择继续限定范围的稳定来源调查，或另行评估根图级节点合同修订；当前不放宽验收。

## 2026-09-28 根候选准备与路径归属审核

用户确认本次计划及 R3c-2 27 项准备范围。27 项官方代理/校验准备全部退出 0，UTC 11:37:43.336–11:39:29.113（约 105.777 秒），四生产锁未变。perfstat 仅离线缓存元数据读取，退出 0 / 0.088 秒。准备不代表安全等价。

新纯规则 RED 退出 1 / 4.646 秒，新增精确候选、根用途与严格 JSON 场景均观察到语义失败；GREEN 退出 0 / 3.156 秒。联合真实 Lock 退出 1 / 34.049 秒，仅证据绑定与归属未审核失败，八组读取、原包映射、用途与模块边界均无失败。两模块新严格 graph 分别退出 0 / 0.131、0.101 秒。根完整模块查询退出 0 / 0.811 秒，共 308 条，31 差异均匹配批准的精确候选。

### 归属审查方法与边界

[逐项路径证据](2026-09-28-devtools-membership-path-audit.json)保存严格图及每项判定的入口链：从两个固定主模块入口逐边追踪，不读取 list -m 快照来生成期望。原 665 非主模块分别审核两种归属，共 1330 项。root 保留 306、移出 359；tools 保留 507、移出 158；没有节点同时从两侧丢失，Required 32 均保留。每个 removed 项缺席于本侧入口可达图，同时有另一侧可达链；不是从全仓删除依赖。下列唯一记录逐项绑定身份和原值，不共享空泛删除证据。快照与期望仍由独立检查器比较。

### 就绪延迟的补充诊断（2026-09-28）

仅在测试副本启用阶段 UTC 计时，记录脚本入口、归属初始化前后与校验命令开始；生产脚本不变，12 秒就绪/20 秒外层及清理时限不变。首次 verify-c11 诊断退出 0（12.405 秒），就绪 6.518 秒；可见四个 PowerShell 入口和两轮依赖验证，初始化分别约 0.154–0.407 秒。随后三个消费者各两次的对照退出 0（60.612 秒）：check-tools 就绪 3.644/3.506 秒，generate 3.463/3.491 秒，verify-c11 6.729/8.825 秒。均验证自有进程取消及无关进程存活。

这些样本未复现历史失败，只能说明本次执行路径和耗时，不能定位历史根因、证明冷态稳定或归因 Defender。插桩可能扰动时序，不作为未修改入口验收。随后关闭插桩、原完整 Windows 选择器以 10 分钟上限独立运行，本代理不并行运行其他测试或依赖查询；最终退出 0 /399.344 秒，详见上方最新门槛记录。

### 源码用途审查补充

后续递归源码核对补充：prometheus/common 的 config/http_config.go 和 grpc 的 credentials/oauth/oauth.go 直接使用 oauth2，但这两个子包不在当前三组根侧包集合中。此前“直接 Go 文件搜索”仅为限定文件范围，不能解读为整个 common/grpc 不使用 oauth2。上述文件未带平台构建标签；ch-go/client.go、Goose/pkg/dockermanage/manager.go 同样未带平台构建标签，未使用原因是子包边界，不是 Windows 自动排除。grpc 中 genproto/googleapis/rpc 引用由独立 rpc 模块提供，不能按包名前缀误算成旧聚合 genproto 提供包；真实包映射按 Go 的 Module.Path 判定。

补充平台元数据检查：固定 Windows Go 1.26.5 以 GOOS=linux/darwin、GOARCH=amd64、CGO_ENABLED=0 各读取 root/integration/goose，共六次严格离线 list -deps（root/integration 含 -test）；均退出 0，28 项候选均没有提供包，也未报告重复包。每条沿用 90 秒上限，无 -e、下载、编译或第三方测试执行。它仅说明这两种构建配置的解析结果，不等于在 Linux/macOS 上运行成功，不覆盖 CGO=1、其他架构/标签或所有第三方测试。

root 28 的原高值来源和当前声明父边逐项见[评估表](2026-09-28-devtools-root28-assessment-proposal.md)。新鲜八组包映射不变，根三组未使用这 28 项；不会因此宣称整个上游模块没有引用。具体正对照：ch-go v0.71.0 的 client.go/handshake.go 使用 zap，internal/version 使用 go-version，但 Goose 当前只经 ClickHouse 使用 ch-go/compress、proto；这与模块级图中存在 zap 而当前包集合无 zap 一致。Goose pkg/dockermanage 使用 moby API/client，不在 CLI 当前闭包。YDB tests/integration 使用 xerrors/zap，未执行这些第三方测试。当前 CH/runtime/common 的直接 Go 文件搜索未发现其他相应目标 import；只说明此次源码范围，不能覆盖所有传递包或未来平台行为。

root 三旧候选继续以 pprof 的 readline/demangle 声明和 sqlite 的测试导入、grpc 测试导入 sdk/metric 所引出的 metric/x 为来源；不升级上游或反复 pin。tools 四项仍以 moby/client pkg/progress 与 parser goyacc 的用途解释图节点；五工具当前无相关包。原值源、精确候选及必要性限于获批隔离边界；不声称全面漏洞审查、第三方动态测试或其他平台通过。B 类历史图完整链随逐项路径证据保存，不将旧父节点版本当实际选中值。cloud/genproto 未加直接 require，无包映射变化或重复包错误。

以下绑定记录仅供机器检查已审阅的身份、来源校验和和移出链，尚不代表完整 Task 2/Task 3 验收；还须稳定性、完整性、脚本及独立审查。

<!-- devtools-evidence {"ID":"exception-root-cel-dev-expr","Owner":"root","Path":"cel.dev/expr","Version":"v0.25.1","Kind":"exception","Sum":"h1:1KrZg61W6TWSxuNZ37Xy49ps13NUovb66QLprthtwi4=","GoModSum":"h1:hrXvqGP6G6gyx8UAHSHJ5RGk//1Oj5nXQ2NI02Nrsg4="} -->

root cel.dev/expr v0.25.1：活动链 `talenro.local/platform -> google.golang.org/grpc@v1.80.0 ; google.golang.org/grpc@v1.80.0 -> cel.dev/expr@v0.25.1`。原值 v0.25.2；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-cloud-google-com-go","Owner":"root","Path":"cloud.google.com/go","Version":"v0.34.0","Kind":"exception","Sum":"h1:eOI3/cP2VTU6uZLDYAoic+eyzzB9YyGmJ7eIjl8rOPg=","GoModSum":"h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw="} -->

root cloud.google.com/go v0.34.0：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 ; github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 -> google.golang.org/grpc@v1.47.0 ; google.golang.org/grpc@v1.47.0 -> golang.org/x/oauth2@v0.0.0-20200107190931-bf48bf16ab8d ; golang.org/x/oauth2@v0.0.0-20200107190931-bf48bf16ab8d -> cloud.google.com/go@v0.34.0`。原值 v0.121.2；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-Azure-go-ansiterm","Owner":"root","Path":"github.com/Azure/go-ansiterm","Version":"v0.0.0-20210617225240-d185dfc1b5a1","Kind":"exception","Sum":"h1:UQHMgLO+TxOElx5B5HZ4hJQsoJ/PvUvKRhJHDQXO8P8=","GoModSum":"h1:xomTg63KZ2rFqZQzSB4Vz2SUXa1BpHTVz9L5PTmPC4E="} -->

root github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/Azure/go-ansiterm@v0.0.0-20210617225240-d185dfc1b5a1`。原值 v0.0.0-20250102033503-faa5f7b0171c；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-BurntSushi-toml","Owner":"root","Path":"github.com/BurntSushi/toml","Version":"v1.3.2","Kind":"exception","Sum":"h1:o7IhLm0Msx3BaB+n3Ag7L8EVlByGnpq14C4YWiu/gL8=","GoModSum":"h1:CxXYINrC8qIiEnFrOxCa7Jy5BFHlXnUU2pbicEuybxQ="} -->

root github.com/BurntSushi/toml v1.3.2：活动链 `talenro.local/platform -> github.com/oapi-codegen/runtime@v1.6.0 ; github.com/oapi-codegen/runtime@v1.6.0 -> github.com/BurntSushi/toml@v1.3.2`。原值 v1.6.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-alecthomas-units","Owner":"root","Path":"github.com/alecthomas/units","Version":"v0.0.0-20211218093645-b94a6e3cc137","Kind":"exception","Sum":"h1:s6gZFSlWYmbqAuRjVTiNNhvNRfY2Wxp9nhfyel4rklc=","GoModSum":"h1:OMCwj8VM1Kc9e19TLln2VL61YJF0x1XFtfdL4JdbSyE="} -->

root github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137：活动链 `talenro.local/platform -> github.com/prometheus/common@v0.66.1 ; github.com/prometheus/common@v0.66.1 -> github.com/alecthomas/units@v0.0.0-20211218093645-b94a6e3cc137`。原值 v0.0.0-20240927000941-0f3dac36c52b；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-ebitengine-purego","Owner":"root","Path":"github.com/ebitengine/purego","Version":"v0.8.4","Kind":"exception","Sum":"h1:CF7LEKg5FFOsASUj0+QwaXf8Ht6TlFxg09+S9wz0omw=","GoModSum":"h1:iIjxzd6CiRiOG0UyXP+V1+jWqUXVjPKLAI0mRfJZTmQ="} -->

root github.com/ebitengine/purego v0.8.4：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/ebitengine/purego@v0.8.4`。原值 v0.10.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-gorilla-websocket","Owner":"root","Path":"github.com/gorilla/websocket","Version":"v1.4.2","Kind":"exception","Sum":"h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=","GoModSum":"h1:YR8l580nyteQvAITg2hZ9XVh4b55+EU/adAjf1fMHhE="} -->

root github.com/gorilla/websocket v1.4.2：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/gorilla/websocket@v1.4.2`。原值 v1.5.3；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-hashicorp-go-version","Owner":"root","Path":"github.com/hashicorp/go-version","Version":"v1.8.0","Kind":"exception","Sum":"h1:KAkNb1HAiZd1ukkxDFGmokVZe1Xy9HG6NUp+bPle2i4=","GoModSum":"h1:fltr4n8CU8Ke44wwGCBoEymUuxUHl09ZGVZPK5anwXA="} -->

root github.com/hashicorp/go-version v1.8.0：活动链 `talenro.local/platform -> github.com/ClickHouse/ch-go@v0.71.0 ; github.com/ClickHouse/ch-go@v0.71.0 -> github.com/hashicorp/go-version@v1.8.0`。原值 v1.9.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-mattn-go-colorable","Owner":"root","Path":"github.com/mattn/go-colorable","Version":"v0.1.14","Kind":"exception","Sum":"h1:9A9LHSqF/7dyVVX6g0U9cwm9pG3kP9gSzcuIPHPsaIE=","GoModSum":"h1:6LmQG8QLFO4G5z1gPvYEzlUgJ2wF+stgPZH1UqBm1s8="} -->

root github.com/mattn/go-colorable v0.1.14：活动链 `talenro.local/platform -> github.com/oapi-codegen/runtime@v1.6.0 ; github.com/oapi-codegen/runtime@v1.6.0 -> github.com/mattn/go-colorable@v0.1.14`。原值 v0.1.15；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-moby-moby-api","Owner":"root","Path":"github.com/moby/moby/api","Version":"v1.54.2","Kind":"exception","Sum":"h1:wiat9QAhnDQjA7wk1kh/TqHz2I1uUA7M7t9SAl/JNXg=","GoModSum":"h1:+RQ6wluLwtYaTd1WnPLykIDPekkuyD/ROWQClE83pzs="} -->

root github.com/moby/moby/api v1.54.2：活动链 `talenro.local/platform -> github.com/pressly/goose/v3@v3.27.1 ; github.com/pressly/goose/v3@v3.27.1 -> github.com/moby/moby/api@v1.54.2`。原值 v1.55.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-moby-moby-client","Owner":"root","Path":"github.com/moby/moby/client","Version":"v0.4.1","Kind":"exception","Sum":"h1:DMQgisVoMkmMs7fp3ROSdiBnoAu8+vo3GggFl06M/wY=","GoModSum":"h1:z52C9O2POPOsnxZAy//WtKcQ32P+jT/NGeXu/7nfjGQ="} -->

root github.com/moby/moby/client v0.4.1：活动链 `talenro.local/platform -> github.com/pressly/goose/v3@v3.27.1 ; github.com/pressly/goose/v3@v3.27.1 -> github.com/moby/moby/client@v0.4.1`。原值 v0.5.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-moby-term","Owner":"root","Path":"github.com/moby/term","Version":"v0.5.0","Kind":"exception","Sum":"h1:xt8Q1nalod/v7BqbG21f8mQPqH+xAaC9C3N3wfWbVP0=","GoModSum":"h1:8FzsFHVUBGZdbDsJw/ot+X+d5HLUbvklYLJ9uGfcI3Y="} -->

root github.com/moby/term v0.5.0：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/term@v0.5.0`。原值 v0.5.2；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-pelletier-go-toml-v2","Owner":"root","Path":"github.com/pelletier/go-toml/v2","Version":"v2.2.2","Kind":"exception","Sum":"h1:aYUidT7k73Pcl9nb2gScu7NSrKCSHIDE89b3+6Wq+LM=","GoModSum":"h1:1t835xjRzz80PqgE6HHgN2JOsmgYu/h4qDAS4n929Rs="} -->

root github.com/pelletier/go-toml/v2 v2.2.2：活动链 `talenro.local/platform -> github.com/oapi-codegen/runtime@v1.6.0 ; github.com/oapi-codegen/runtime@v1.6.0 -> github.com/pelletier/go-toml/v2@v2.2.2`。原值 v2.3.1；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-power-devops-perfstat","Owner":"root","Path":"github.com/power-devops/perfstat","Version":"v0.0.0-20210106213030-5aafc221ea8c","Kind":"exception","Sum":"h1:ncq/mPwQF4JjgDlrVEn3C11VoGHZN7m8qihwgMEtzYw=","GoModSum":"h1:OmDBASR4679mdNQnz2pUhc2G8CO2JrUAVFDRBDP/hJE="} -->

root github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/power-devops/perfstat@v0.0.0-20210106213030-5aafc221ea8c`。原值 v0.0.0-20240221224432-82ca36839d55；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-shirou-gopsutil-v4","Owner":"root","Path":"github.com/shirou/gopsutil/v4","Version":"v4.25.6","Kind":"exception","Sum":"h1:kLysI2JsKorfaFPcYmcJqbzROzsBWEOAtw6A7dIfqXs=","GoModSum":"h1:PfybzyydfZcN+JMMjkF6Zb8Mq1A/VcogFFg7hj50W9c="} -->

root github.com/shirou/gopsutil/v4 v4.25.6：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/shirou/gopsutil/v4@v4.25.6`。原值 v4.26.4；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-sirupsen-logrus","Owner":"root","Path":"github.com/sirupsen/logrus","Version":"v1.9.3","Kind":"exception","Sum":"h1:dueUQJ1C2q9oE3F7wvmSGAaVtTmUizReu6fjN8uqzbQ=","GoModSum":"h1:naHLuLoDiP4jHNo9R0sCBMtWGeIprob74mVsIT4qYEQ="} -->

root github.com/sirupsen/logrus v1.9.3：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/sirupsen/logrus@v1.9.3`。原值 v1.9.4；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-tklauser-go-sysconf","Owner":"root","Path":"github.com/tklauser/go-sysconf","Version":"v0.3.12","Kind":"exception","Sum":"h1:0QaGUFOdQaIVdPgfITYzaTegZvdCjmYO52cSFAEVmqU=","GoModSum":"h1:Ho14jnntGE1fpdOqQEEaiKRpvIavV0hSfmBq8nJbHYI="} -->

root github.com/tklauser/go-sysconf v0.3.12：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/tklauser/go-sysconf@v0.3.12`。原值 v0.3.16；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-tklauser-numcpus","Owner":"root","Path":"github.com/tklauser/numcpus","Version":"v0.6.1","Kind":"exception","Sum":"h1:ng9scYS7az0Bk4OZLvrNXNSAO2Pxr1XXRAPyjhIx+Fk=","GoModSum":"h1:1XfjsgE2zo8GVw7POkMbHENHzVg3GzmoZ9fESEdAacY="} -->

root github.com/tklauser/numcpus v0.6.1：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/tklauser/numcpus@v0.6.1`。原值 v0.11.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-go-opentelemetry-io-contrib-instrumentation-net-http-otelhttp","Owner":"root","Path":"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp","Version":"v0.68.0","Kind":"exception","Sum":"h1:CqXxU8VOmDefoh0+ztfGaymYbhdB/tT3zs79QaZTNGY=","GoModSum":"h1:BuhAPThV8PBHBvg8ZzZ/Ok3idOdhWIodywz2xEcRbJo="} -->

root go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0：活动链 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0 ; github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.49.0`。原值 v0.69.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-go-uber-org-zap","Owner":"root","Path":"go.uber.org/zap","Version":"v1.27.1","Kind":"exception","Sum":"h1:08RqriUEv8+ArZRYSTXy1LeBScaMpVSTBhCeaZYfMYc=","GoModSum":"h1:GB2qFLM7cTU87MWRP2mPIjqfIDnGu+VIO4V/SdhGo2E="} -->

root go.uber.org/zap v1.27.1：活动链 `talenro.local/platform -> github.com/ClickHouse/ch-go@v0.71.0 ; github.com/ClickHouse/ch-go@v0.71.0 -> go.uber.org/zap@v1.27.1`。原值 v1.28.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-golang-org-x-lint","Owner":"root","Path":"golang.org/x/lint","Version":"v0.0.0-20190313153728-d0100b6bd8b3","Kind":"exception","Sum":"h1:XQyxROzUlZH+WIQwySDgnISgOivlhjIEwaQaJEJrrN0=","GoModSum":"h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc="} -->

root golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 ; github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 -> google.golang.org/grpc@v1.47.0 ; google.golang.org/grpc@v1.47.0 -> google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013 ; google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013 -> golang.org/x/lint@v0.0.0-20190313153728-d0100b6bd8b3`。原值 v0.0.0-20190930215403-16217165b5de；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-golang-org-x-oauth2","Owner":"root","Path":"golang.org/x/oauth2","Version":"v0.34.0","Kind":"exception","Sum":"h1:hqK/t4AKgbqWkdkcAeI8XLmbK+4m4G5YeQRrmiotGlw=","GoModSum":"h1:lzm5WQJQwKZ3nwavOZ3IS5Aulzxi68dUSgRHujetwEA="} -->

root golang.org/x/oauth2 v0.34.0：活动链 `talenro.local/platform -> github.com/prometheus/client_golang@v1.23.2 ; github.com/prometheus/client_golang@v1.23.2 -> golang.org/x/oauth2@v0.30.0`。原值 v0.36.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-golang-org-x-xerrors","Owner":"root","Path":"golang.org/x/xerrors","Version":"v0.0.0-20200804184101-5ec99f83aff1","Kind":"exception","Sum":"h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=","GoModSum":"h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0="} -->

root golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-sdk/v3@v3.135.0 ; github.com/ydb-platform/ydb-go-sdk/v3@v3.135.0 -> golang.org/x/xerrors@v0.0.0-20200804184101-5ec99f83aff1`。原值 v0.0.0-20220517211312-f3a8303e98df；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-google-golang-org-appengine","Owner":"root","Path":"google.golang.org/appengine","Version":"v1.4.0","Kind":"exception","Sum":"h1:/wp5JvzpHIxhs/dumFmF7BXTf3Z+dd4uXta4kVyO508=","GoModSum":"h1:xpcJRLb0r/rnEns0DIKYYv+WjYCduHsrkT7/EB5XEv4="} -->

root google.golang.org/appengine v1.4.0：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 ; github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 -> google.golang.org/grpc@v1.47.0 ; google.golang.org/grpc@v1.47.0 -> golang.org/x/oauth2@v0.0.0-20200107190931-bf48bf16ab8d ; golang.org/x/oauth2@v0.0.0-20200107190931-bf48bf16ab8d -> google.golang.org/appengine@v1.4.0`。原值 v1.6.7；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-google-golang-org-genproto","Owner":"root","Path":"google.golang.org/genproto","Version":"v0.0.0-20200526211855-cb27e3aa2013","Kind":"exception","Sum":"h1:+kGHl1aib/qcwaRi1CbqBZ1rk19r85MNUf8HaBghugY=","GoModSum":"h1:NbSheEEYHJ7i3ixzK3sjbqSGDJWnxyFXZblF3eUsNvo="} -->

root google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 ; github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 -> google.golang.org/grpc@v1.47.0 ; google.golang.org/grpc@v1.47.0 -> google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013`。原值 v0.0.0-20220519153652-3a47de7e79bd；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-google-golang-org-genproto-googleapis-api","Owner":"root","Path":"google.golang.org/genproto/googleapis/api","Version":"v0.0.0-20260120221211-b8f7ae30c516","Kind":"exception","Sum":"h1:vmC/ws+pLzWjj/gzApyoZuSVrDtF1aod4u/+bbj8hgM=","GoModSum":"h1:p3MLuOwURrGBRoEyFHBT3GjUwaCQVKeNqqWxlcISGdw="} -->

root google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516：活动链 `talenro.local/platform -> google.golang.org/grpc@v1.80.0 ; google.golang.org/grpc@v1.80.0 -> google.golang.org/genproto/googleapis/api@v0.0.0-20260120221211-b8f7ae30c516`。原值 v0.0.0-20260715232425-e75dac1f907d；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-gopkg-in-yaml-v2","Owner":"root","Path":"gopkg.in/yaml.v2","Version":"v2.2.3","Kind":"exception","Sum":"h1:fvjTMHxHEw/mxHbtzPi3JCcKXQRAnQTBRo6YCJSVHKI=","GoModSum":"h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI="} -->

root gopkg.in/yaml.v2 v2.2.3：活动链 `talenro.local/platform -> github.com/elastic/go-sysinfo@v1.15.4 ; github.com/elastic/go-sysinfo@v1.15.4 -> howett.net/plist@v0.0.0-20181124034731-591f970eefbb ; howett.net/plist@v0.0.0-20181124034731-591f970eefbb -> gopkg.in/yaml.v2@v2.2.1`。原值 v2.4.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-honnef-co-go-tools","Owner":"root","Path":"honnef.co/go/tools","Version":"v0.0.0-20190523083050-ea95bdfd59fc","Kind":"exception","Sum":"h1:/hemPrYIhOhy8zYrNj+069zDB68us2sMGsfkFJO0iZs=","GoModSum":"h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4="} -->

root honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc：活动链 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 ; github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180 -> google.golang.org/grpc@v1.47.0 ; google.golang.org/grpc@v1.47.0 -> google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013 ; google.golang.org/genproto@v0.0.0-20200526211855-cb27e3aa2013 -> honnef.co/go/tools@v0.0.0-20190523083050-ea95bdfd59fc`。原值 v0.7.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-chzyer-readline","Owner":"root","Path":"github.com/chzyer/readline","Version":"v1.5.1","Kind":"exception","Sum":"h1:upd/6fQk4src78LMRzh5vItIt361/o4uq553V8B5sGI=","GoModSum":"h1:Eh+b79XXUwfKfcPLepksvw2tcLE/Ct21YObkaSkeBlk="} -->

root github.com/chzyer/readline v1.5.1：活动链 `talenro.local/platform -> github.com/google/pprof@v0.0.0-20260115054156-294ebfa9ad83 ; github.com/google/pprof@v0.0.0-20260115054156-294ebfa9ad83 -> github.com/chzyer/readline@v1.5.1`。原值 v0.0.0-20180603132655-2972be24d48e；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-github-com-ianlancetaylor-demangle","Owner":"root","Path":"github.com/ianlancetaylor/demangle","Version":"v0.0.0-20250417193237-f615e6bd150b","Kind":"exception","Sum":"h1:ogbOPx86mIhFy764gGkqnkFC8m5PJA7sPzlk9ppLVQA=","GoModSum":"h1:gx7rwoVhcfuVKG5uya9Hs3Sxj7EIvldVofAWIUtGouw="} -->

root github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b：活动链 `talenro.local/platform -> github.com/google/pprof@v0.0.0-20260115054156-294ebfa9ad83 ; github.com/google/pprof@v0.0.0-20260115054156-294ebfa9ad83 -> github.com/ianlancetaylor/demangle@v0.0.0-20250417193237-f615e6bd150b`。原值 v0.0.0-20200824232613-28f6c0f3b639；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-root-go-opentelemetry-io-otel-metric-x","Owner":"root","Path":"go.opentelemetry.io/otel/metric/x","Version":"v0.66.0","Kind":"exception","Sum":"h1:YkCrx1zLOChi9ZcZ6euupOcsgzbVlec7D/xoEU1+cTA=","GoModSum":"h1:d1+BDj9t96do0/1LoU1ayfCv79ZgNE41qbhBvnMOBZk="} -->

root go.opentelemetry.io/otel/metric/x v0.66.0：活动链 `talenro.local/platform -> go.opentelemetry.io/otel/sdk/metric@v1.44.0 ; go.opentelemetry.io/otel/sdk/metric@v1.44.0 -> go.opentelemetry.io/otel/metric/x@v0.66.0`。原值 无此节点；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-tools-golang-org-x-time","Owner":"tools","Path":"golang.org/x/time","Version":"v0.11.0","Kind":"exception","Sum":"h1:/bpjEDfN9tkoN/ryeYHnv5hcMlc8ncjMcM4XBk5NWV0=","GoModSum":"h1:CDIdPxbZBQxdj6cxyCIdrNogrJKMJ7pr37NYpMcMDSg="} -->

tools golang.org/x/time v0.11.0：活动链 `talenro.local/devtools -> github.com/moby/moby/client@v0.5.0 ; github.com/moby/moby/client@v0.5.0 -> golang.org/x/time@v0.11.0`。原值 v0.14.0；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-tools-modernc-org-mathutil","Owner":"tools","Path":"modernc.org/mathutil","Version":"v1.6.0","Kind":"exception","Sum":"h1:fRe9+AmYlaej+64JsEEhoWuAYBkOtQiMEU7n/XgfYi4=","GoModSum":"h1:Ui5Q9q1TR2gFm0AQRqQUaBWFLAhQpCwNcuhBOSedWPo="} -->

tools modernc.org/mathutil v1.6.0：活动链 `talenro.local/devtools -> github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 ; github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/mathutil@v1.6.0`。原值 v1.7.1；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-tools-modernc-org-sortutil","Owner":"tools","Path":"modernc.org/sortutil","Version":"v1.2.0","Kind":"exception","Sum":"h1:jQiD3PfS2REGJNzNCMMaLSp/wdMNieTbKX920Cqdgqc=","GoModSum":"h1:TKU2s7kJMf1AE84OoiGppNHJwvB753OYfNl2WRb++Ss="} -->

tools modernc.org/sortutil v1.2.0：活动链 `talenro.local/devtools -> github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 ; github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/sortutil@v1.2.0`。原值 v1.2.1；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"exception-tools-modernc-org-strutil","Owner":"tools","Path":"modernc.org/strutil","Version":"v1.2.0","Kind":"exception","Sum":"h1:agBi9dp1I+eOnxXeiZawM8F4LawKv4NzGWSaLfyeNZA=","GoModSum":"h1:/mdcBmfOibveCTBxUl5B5l6W+TTH1FXPLHZE6bTosX0="} -->

tools modernc.org/strutil v1.2.0：活动链 `talenro.local/devtools -> github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 ; github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/strutil@v1.2.0`。原值 v1.2.1；用途与必要性按上述对应审查段/评估表，校验和来自精确准备结果（perfstat 为既有缓存离线读取）。

<!-- devtools-evidence {"ID":"membership-root-4d63-com-gocheckcompilerdirectives","Owner":"root","Path":"4d63.com/gocheckcompilerdirectives","Version":"v1.3.0","Kind":"removal"} -->

root 4d63.com/gocheckcompilerdirectives：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> 4d63.com/gocheckcompilerdirectives@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-4d63-com-gochecknoglobals","Owner":"root","Path":"4d63.com/gochecknoglobals","Version":"v0.2.2","Kind":"removal"} -->

root 4d63.com/gochecknoglobals：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> 4d63.com/gochecknoglobals@v0.2.2`。基线 v0.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-bufplugin-connectrpc-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/bufplugin/connectrpc/go","Version":"v1.20.0-20250718181942-e35f9b667443.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/bufplugin/connectrpc/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/gen/go/bufbuild/registry/connectrpc/go@v1.20.0-20260713175918-10d915f5b43b.1 -> buf.build/gen/go/bufbuild/bufplugin/connectrpc/go@v1.20.0-20250718181942-e35f9b667443.1`。基线 v1.20.0-20250718181942-e35f9b667443.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-bufplugin-protocolbuffers-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go","Version":"v1.36.11-20260626152828-968bf0468096.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go@v1.36.11-20260626152828-968bf0468096.1`。基线 v1.36.11-20260626152828-968bf0468096.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-protodescriptor-protocolbuffers-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/protodescriptor/protocolbuffers/go","Version":"v1.36.11-20250109164928-1da0de137947.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/protodescriptor/protocolbuffers/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/bufbuild/protodescriptor/protocolbuffers/go@v1.36.11-20250109164928-1da0de137947.1`。基线 v1.36.11-20250109164928-1da0de137947.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-protovalidate-connectrpc-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/protovalidate/connectrpc/go","Version":"v1.20.0-20250717185734-6c6e0d3c608e.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/protovalidate/connectrpc/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/gen/go/bufbuild/registry/connectrpc/go@v1.20.0-20260713175918-10d915f5b43b.1 -> buf.build/gen/go/bufbuild/protovalidate/connectrpc/go@v1.20.0-20250717185734-6c6e0d3c608e.1`。基线 v1.20.0-20250717185734-6c6e0d3c608e.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-protovalidate-protocolbuffers-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go","Version":"v1.36.11-20260709200747-435963d16310.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go@v1.36.11-20260709200747-435963d16310.1`。基线 v1.36.11-20260709200747-435963d16310.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-registry-connectrpc-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/registry/connectrpc/go","Version":"v1.20.0-20260713175918-10d915f5b43b.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/registry/connectrpc/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/bufbuild/registry/connectrpc/go@v1.20.0-20260713175918-10d915f5b43b.1`。基线 v1.20.0-20260713175918-10d915f5b43b.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-bufbuild-registry-protocolbuffers-go","Owner":"root","Path":"buf.build/gen/go/bufbuild/registry/protocolbuffers/go","Version":"v1.36.11-20260713175918-10d915f5b43b.1","Kind":"removal"} -->

root buf.build/gen/go/bufbuild/registry/protocolbuffers/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/bufbuild/registry/protocolbuffers/go@v1.36.11-20260713175918-10d915f5b43b.1`。基线 v1.36.11-20260713175918-10d915f5b43b.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-gen-go-pluginrpc-pluginrpc-protocolbuffers-go","Owner":"root","Path":"buf.build/gen/go/pluginrpc/pluginrpc/protocolbuffers/go","Version":"v1.36.11-20241007202033-cf42259fcbfc.1","Kind":"removal"} -->

root buf.build/gen/go/pluginrpc/pluginrpc/protocolbuffers/go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/gen/go/pluginrpc/pluginrpc/protocolbuffers/go@v1.36.11-20241007202033-cf42259fcbfc.1`。基线 v1.36.11-20241007202033-cf42259fcbfc.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-app","Owner":"root","Path":"buf.build/go/app","Version":"v0.2.1-0.20260626143626-be153867abea","Kind":"removal"} -->

root buf.build/go/app：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/app@v0.2.1-0.20260626143626-be153867abea`。基线 v0.2.1-0.20260626143626-be153867abea，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-bufplugin","Owner":"root","Path":"buf.build/go/bufplugin","Version":"v0.10.0","Kind":"removal"} -->

root buf.build/go/bufplugin：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/bufplugin@v0.10.0`。基线 v0.10.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-bufprivateusage","Owner":"root","Path":"buf.build/go/bufprivateusage","Version":"v0.1.0","Kind":"removal"} -->

root buf.build/go/bufprivateusage：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/bufprivateusage@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-hyperpb","Owner":"root","Path":"buf.build/go/hyperpb","Version":"v0.1.3","Kind":"removal"} -->

root buf.build/go/hyperpb：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/go/protovalidate@v1.2.0 -> buf.build/go/hyperpb@v0.1.3`。基线 v0.1.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-interrupt","Owner":"root","Path":"buf.build/go/interrupt","Version":"v1.1.0","Kind":"removal"} -->

root buf.build/go/interrupt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/interrupt@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-protovalidate","Owner":"root","Path":"buf.build/go/protovalidate","Version":"v1.2.0","Kind":"removal"} -->

root buf.build/go/protovalidate：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/protovalidate@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-protoyaml","Owner":"root","Path":"buf.build/go/protoyaml","Version":"v0.7.0","Kind":"removal"} -->

root buf.build/go/protoyaml：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/protoyaml@v0.7.0`。基线 v0.7.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-spdx","Owner":"root","Path":"buf.build/go/spdx","Version":"v0.2.0","Kind":"removal"} -->

root buf.build/go/spdx：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/spdx@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-buf-build-go-standard","Owner":"root","Path":"buf.build/go/standard","Version":"v0.1.1-0.20260325175353-2b287e071df5","Kind":"removal"} -->

root buf.build/go/standard：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> buf.build/go/standard@v0.1.1-0.20260325175353-2b287e071df5`。基线 v0.1.1-0.20260325175353-2b287e071df5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-charm-land-lipgloss-v2","Owner":"root","Path":"charm.land/lipgloss/v2","Version":"v2.0.3","Kind":"removal"} -->

root charm.land/lipgloss/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> charm.land/lipgloss/v2@v2.0.3`。基线 v2.0.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-cloud-google-com-go-auth","Owner":"root","Path":"cloud.google.com/go/auth","Version":"v0.16.5","Kind":"removal"} -->

root cloud.google.com/go/auth：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> cloud.google.com/go/auth@v0.16.5`。基线 v0.16.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-cloud-google-com-go-compute","Owner":"root","Path":"cloud.google.com/go/compute","Version":"v1.6.1","Kind":"removal"} -->

root cloud.google.com/go/compute：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> cloud.google.com/go/compute@v1.6.1`。基线 v1.6.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-cloud-google-com-go-firestore","Owner":"root","Path":"cloud.google.com/go/firestore","Version":"v1.6.1","Kind":"removal"} -->

root cloud.google.com/go/firestore：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> cloud.google.com/go/firestore@v1.6.1`。基线 v1.6.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-codeberg-org-chavacava-garif","Owner":"root","Path":"codeberg.org/chavacava/garif","Version":"v0.2.0","Kind":"removal"} -->

root codeberg.org/chavacava/garif：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> codeberg.org/chavacava/garif@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-codeberg-org-polyfloyd-go-errorlint","Owner":"root","Path":"codeberg.org/polyfloyd/go-errorlint","Version":"v1.9.0","Kind":"removal"} -->

root codeberg.org/polyfloyd/go-errorlint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> codeberg.org/polyfloyd/go-errorlint@v1.9.0`。基线 v1.9.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-connectrpc-com-connect","Owner":"root","Path":"connectrpc.com/connect","Version":"v1.20.0","Kind":"removal"} -->

root connectrpc.com/connect：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> connectrpc.com/connect@v1.20.0`。基线 v1.20.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-connectrpc-com-grpcreflect","Owner":"root","Path":"connectrpc.com/grpcreflect","Version":"v1.3.0","Kind":"removal"} -->

root connectrpc.com/grpcreflect：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/bufbuild/buf@v1.72.0 -> connectrpc.com/grpcreflect@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-connectrpc-com-otelconnect","Owner":"root","Path":"connectrpc.com/otelconnect","Version":"v0.9.0","Kind":"removal"} -->

root connectrpc.com/otelconnect：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> connectrpc.com/otelconnect@v0.9.0`。基线 v0.9.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-dev-gaijin-team-go-exhaustruct-v4","Owner":"root","Path":"dev.gaijin.team/go/exhaustruct/v4","Version":"v4.0.0","Kind":"removal"} -->

root dev.gaijin.team/go/exhaustruct/v4：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> dev.gaijin.team/go/exhaustruct/v4@v4.0.0`。基线 v4.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-dev-gaijin-team-go-golib","Owner":"root","Path":"dev.gaijin.team/go/golib","Version":"v0.6.0","Kind":"removal"} -->

root dev.gaijin.team/go/golib：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> dev.gaijin.team/go/golib@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-4meepo-tagalign","Owner":"root","Path":"github.com/4meepo/tagalign","Version":"v1.4.3","Kind":"removal"} -->

root github.com/4meepo/tagalign：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/4meepo/tagalign@v1.4.3`。基线 v1.4.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Abirdcfly-dupword","Owner":"root","Path":"github.com/Abirdcfly/dupword","Version":"v0.1.7","Kind":"removal"} -->

root github.com/Abirdcfly/dupword：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Abirdcfly/dupword@v0.1.7`。基线 v0.1.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-AdminBenni-iota-mixing","Owner":"root","Path":"github.com/AdminBenni/iota-mixing","Version":"v1.0.0","Kind":"removal"} -->

root github.com/AdminBenni/iota-mixing：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/AdminBenni/iota-mixing@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-AlwxSin-noinlineerr","Owner":"root","Path":"github.com/AlwxSin/noinlineerr","Version":"v1.0.5","Kind":"removal"} -->

root github.com/AlwxSin/noinlineerr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/AlwxSin/noinlineerr@v1.0.5`。基线 v1.0.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Antonboom-errname","Owner":"root","Path":"github.com/Antonboom/errname","Version":"v1.1.1","Kind":"removal"} -->

root github.com/Antonboom/errname：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Antonboom/errname@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Antonboom-nilnil","Owner":"root","Path":"github.com/Antonboom/nilnil","Version":"v1.1.1","Kind":"removal"} -->

root github.com/Antonboom/nilnil：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Antonboom/nilnil@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Antonboom-testifylint","Owner":"root","Path":"github.com/Antonboom/testifylint","Version":"v1.6.4","Kind":"removal"} -->

root github.com/Antonboom/testifylint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Antonboom/testifylint@v1.6.4`。基线 v1.6.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ClickHouse-clickhouse-go-linter","Owner":"root","Path":"github.com/ClickHouse/clickhouse-go-linter","Version":"v1.2.0","Kind":"removal"} -->

root github.com/ClickHouse/clickhouse-go-linter：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ClickHouse/clickhouse-go-linter@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Djarvur-go-err113","Owner":"root","Path":"github.com/Djarvur/go-err113","Version":"v0.1.1","Kind":"removal"} -->

root github.com/Djarvur/go-err113：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Djarvur/go-err113@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-Masterminds-semver-v3","Owner":"root","Path":"github.com/Masterminds/semver/v3","Version":"v3.5.0","Kind":"removal"} -->

root github.com/Masterminds/semver/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/Masterminds/semver/v3@v3.5.0`。基线 v3.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-MirrexOne-unqueryvet","Owner":"root","Path":"github.com/MirrexOne/unqueryvet","Version":"v1.5.4","Kind":"removal"} -->

root github.com/MirrexOne/unqueryvet：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/MirrexOne/unqueryvet@v1.5.4`。基线 v1.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-OpenPeeDeeP-depguard-v2","Owner":"root","Path":"github.com/OpenPeeDeeP/depguard/v2","Version":"v2.2.1","Kind":"removal"} -->

root github.com/OpenPeeDeeP/depguard/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/OpenPeeDeeP/depguard/v2@v2.2.1`。基线 v2.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alecthomas-assert-v2","Owner":"root","Path":"github.com/alecthomas/assert/v2","Version":"v2.11.0","Kind":"removal"} -->

root github.com/alecthomas/assert/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/alecthomas/chroma/v2@v2.24.1 -> github.com/alecthomas/assert/v2@v2.11.0`。基线 v2.11.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alecthomas-chroma-v2","Owner":"root","Path":"github.com/alecthomas/chroma/v2","Version":"v2.24.1","Kind":"removal"} -->

root github.com/alecthomas/chroma/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alecthomas/chroma/v2@v2.24.1`。基线 v2.24.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alecthomas-go-check-sumtype","Owner":"root","Path":"github.com/alecthomas/go-check-sumtype","Version":"v0.3.1","Kind":"removal"} -->

root github.com/alecthomas/go-check-sumtype：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alecthomas/go-check-sumtype@v0.3.1`。基线 v0.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alecthomas-repr","Owner":"root","Path":"github.com/alecthomas/repr","Version":"v0.5.2","Kind":"removal"} -->

root github.com/alecthomas/repr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/alecthomas/chroma/v2@v2.24.1 -> github.com/alecthomas/repr@v0.5.2`。基线 v0.5.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alecthomas-template","Owner":"root","Path":"github.com/alecthomas/template","Version":"v0.0.0-20190718012654-fb15b899a751","Kind":"removal"} -->

root github.com/alecthomas/template：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/yeya24/promlinter@v0.3.0 -> github.com/alecthomas/template@v0.0.0-20190718012654-fb15b899a751`。基线 v0.0.0-20190718012654-fb15b899a751，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alexkohler-nakedret-v2","Owner":"root","Path":"github.com/alexkohler/nakedret/v2","Version":"v2.0.6","Kind":"removal"} -->

root github.com/alexkohler/nakedret/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alexkohler/nakedret/v2@v2.0.6`。基线 v2.0.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alexkohler-prealloc","Owner":"root","Path":"github.com/alexkohler/prealloc","Version":"v1.1.0","Kind":"removal"} -->

root github.com/alexkohler/prealloc：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alexkohler/prealloc@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alfatraining-structtag","Owner":"root","Path":"github.com/alfatraining/structtag","Version":"v1.0.0","Kind":"removal"} -->

root github.com/alfatraining/structtag：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alfatraining/structtag@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alingse-asasalint","Owner":"root","Path":"github.com/alingse/asasalint","Version":"v0.0.11","Kind":"removal"} -->

root github.com/alingse/asasalint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alingse/asasalint@v0.0.11`。基线 v0.0.11，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-alingse-nilnesserr","Owner":"root","Path":"github.com/alingse/nilnesserr","Version":"v0.2.0","Kind":"removal"} -->

root github.com/alingse/nilnesserr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/alingse/nilnesserr@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-anthropics-anthropic-sdk-go","Owner":"root","Path":"github.com/anthropics/anthropic-sdk-go","Version":"v1.38.0","Kind":"removal"} -->

root github.com/anthropics/anthropic-sdk-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/anthropics/anthropic-sdk-go@v1.38.0`。基线 v1.38.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-armon-go-metrics","Owner":"root","Path":"github.com/armon/go-metrics","Version":"v0.3.10","Kind":"removal"} -->

root github.com/armon/go-metrics：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/armon/go-metrics@v0.3.10`。基线 v0.3.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ashanbrown-forbidigo-v2","Owner":"root","Path":"github.com/ashanbrown/forbidigo/v2","Version":"v2.3.1","Kind":"removal"} -->

root github.com/ashanbrown/forbidigo/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ashanbrown/forbidigo/v2@v2.3.1`。基线 v2.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ashanbrown-makezero-v2","Owner":"root","Path":"github.com/ashanbrown/makezero/v2","Version":"v2.2.1","Kind":"removal"} -->

root github.com/ashanbrown/makezero/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ashanbrown/makezero/v2@v2.2.1`。基线 v2.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-aymanbagabas-go-osc52-v2","Owner":"root","Path":"github.com/aymanbagabas/go-osc52/v2","Version":"v2.0.1","Kind":"removal"} -->

root github.com/aymanbagabas/go-osc52/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/aymanbagabas/go-osc52/v2@v2.0.1`。基线 v2.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-aymanbagabas-go-udiff","Owner":"root","Path":"github.com/aymanbagabas/go-udiff","Version":"v0.4.1","Kind":"removal"} -->

root github.com/aymanbagabas/go-udiff：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `charm.land/lipgloss/v2@v2.0.3 -> github.com/aymanbagabas/go-udiff@v0.4.1`。基线 v0.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bahlo-generic-list-go","Owner":"root","Path":"github.com/bahlo/generic-list-go","Version":"v0.2.0","Kind":"removal"} -->

root github.com/bahlo/generic-list-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/bahlo/generic-list-go@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-benbjohnson-clock","Owner":"root","Path":"github.com/benbjohnson/clock","Version":"v1.1.0","Kind":"removal"} -->

root github.com/benbjohnson/clock：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `go.uber.org/zap@v1.19.0 -> github.com/benbjohnson/clock@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bits-and-blooms-bitset","Owner":"root","Path":"github.com/bits-and-blooms/bitset","Version":"v1.24.4","Kind":"removal"} -->

root github.com/bits-and-blooms/bitset：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/charmbracelet/x/ansi@v0.11.7 -> github.com/bits-and-blooms/bitset@v1.24.4`。基线 v1.24.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bkielbasa-cyclop","Owner":"root","Path":"github.com/bkielbasa/cyclop","Version":"v1.2.3","Kind":"removal"} -->

root github.com/bkielbasa/cyclop：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bkielbasa/cyclop@v1.2.3`。基线 v1.2.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-blackwell-systems-gcf-go","Owner":"root","Path":"github.com/blackwell-systems/gcf-go","Version":"v1.2.2","Kind":"removal"} -->

root github.com/blackwell-systems/gcf-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/speakeasy-api/openapi@v1.24.0 -> github.com/blackwell-systems/gcf-go@v1.2.2`。基线 v1.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-blizzy78-varnamelen","Owner":"root","Path":"github.com/blizzy78/varnamelen","Version":"v0.8.0","Kind":"removal"} -->

root github.com/blizzy78/varnamelen：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/blizzy78/varnamelen@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bmatcuk-doublestar-v4","Owner":"root","Path":"github.com/bmatcuk/doublestar/v4","Version":"v4.10.0","Kind":"removal"} -->

root github.com/bmatcuk/doublestar/v4：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/bufbuild/protocompile@v0.14.2-0.20260716165721-bb5762d29672 -> github.com/bmatcuk/doublestar/v4@v4.10.0`。基线 v4.10.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bombsimon-wsl-v4","Owner":"root","Path":"github.com/bombsimon/wsl/v4","Version":"v4.7.0","Kind":"removal"} -->

root github.com/bombsimon/wsl/v4：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bombsimon/wsl/v4@v4.7.0`。基线 v4.7.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bombsimon-wsl-v5","Owner":"root","Path":"github.com/bombsimon/wsl/v5","Version":"v5.8.0","Kind":"removal"} -->

root github.com/bombsimon/wsl/v5：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bombsimon/wsl/v5@v5.8.0`。基线 v5.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-breml-bidichk","Owner":"root","Path":"github.com/breml/bidichk","Version":"v0.3.3","Kind":"removal"} -->

root github.com/breml/bidichk：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/breml/bidichk@v0.3.3`。基线 v0.3.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-breml-errchkjson","Owner":"root","Path":"github.com/breml/errchkjson","Version":"v0.4.1","Kind":"removal"} -->

root github.com/breml/errchkjson：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/breml/errchkjson@v0.4.1`。基线 v0.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-brianvoe-gofakeit-v6","Owner":"root","Path":"github.com/brianvoe/gofakeit/v6","Version":"v6.28.0","Kind":"removal"} -->

root github.com/brianvoe/gofakeit/v6：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/go/protovalidate@v1.2.0 -> github.com/brianvoe/gofakeit/v6@v6.28.0`。基线 v6.28.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bufbuild-buf","Owner":"root","Path":"github.com/bufbuild/buf","Version":"v1.72.0","Kind":"removal"} -->

root github.com/bufbuild/buf：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bufbuild/buf@v1.72.0`。基线 v1.72.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bufbuild-protocompile","Owner":"root","Path":"github.com/bufbuild/protocompile","Version":"v0.14.2-0.20260716165721-bb5762d29672","Kind":"removal"} -->

root github.com/bufbuild/protocompile：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bufbuild/protocompile@v0.14.2-0.20260716165721-bb5762d29672`。基线 v0.14.2-0.20260716165721-bb5762d29672，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-bufbuild-protoplugin","Owner":"root","Path":"github.com/bufbuild/protoplugin","Version":"v0.0.0-20260414125817-25d1d281b46b","Kind":"removal"} -->

root github.com/bufbuild/protoplugin：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/bufbuild/protoplugin@v0.0.0-20260414125817-25d1d281b46b`。基线 v0.0.0-20260414125817-25d1d281b46b，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-buger-jsonparser","Owner":"root","Path":"github.com/buger/jsonparser","Version":"v1.1.2","Kind":"removal"} -->

root github.com/buger/jsonparser：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/buger/jsonparser@v1.1.2`。基线 v1.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-butuzov-ireturn","Owner":"root","Path":"github.com/butuzov/ireturn","Version":"v0.4.1","Kind":"removal"} -->

root github.com/butuzov/ireturn：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/butuzov/ireturn@v0.4.1`。基线 v0.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-butuzov-mirror","Owner":"root","Path":"github.com/butuzov/mirror","Version":"v1.3.0","Kind":"removal"} -->

root github.com/butuzov/mirror：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/butuzov/mirror@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-catenacyber-perfsprint","Owner":"root","Path":"github.com/catenacyber/perfsprint","Version":"v0.10.1","Kind":"removal"} -->

root github.com/catenacyber/perfsprint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/catenacyber/perfsprint@v0.10.1`。基线 v0.10.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ccojocar-zxcvbn-go","Owner":"root","Path":"github.com/ccojocar/zxcvbn-go","Version":"v1.0.4","Kind":"removal"} -->

root github.com/ccojocar/zxcvbn-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ccojocar/zxcvbn-go@v1.0.4`。基线 v1.0.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charithe-durationcheck","Owner":"root","Path":"github.com/charithe/durationcheck","Version":"v0.0.11","Kind":"removal"} -->

root github.com/charithe/durationcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charithe/durationcheck@v0.0.11`。基线 v0.0.11，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-bubbles","Owner":"root","Path":"github.com/charmbracelet/bubbles","Version":"v0.21.0","Kind":"removal"} -->

root github.com/charmbracelet/bubbles：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/charmbracelet/bubbles@v0.21.0`。基线 v0.21.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-bubbletea","Owner":"root","Path":"github.com/charmbracelet/bubbletea","Version":"v1.3.10","Kind":"removal"} -->

root github.com/charmbracelet/bubbletea：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/charmbracelet/bubbletea@v1.3.10`。基线 v1.3.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-colorprofile","Owner":"root","Path":"github.com/charmbracelet/colorprofile","Version":"v0.4.3","Kind":"removal"} -->

root github.com/charmbracelet/colorprofile：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/colorprofile@v0.4.3`。基线 v0.4.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-lipgloss","Owner":"root","Path":"github.com/charmbracelet/lipgloss","Version":"v1.1.0","Kind":"removal"} -->

root github.com/charmbracelet/lipgloss：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/charmbracelet/lipgloss@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-ultraviolet","Owner":"root","Path":"github.com/charmbracelet/ultraviolet","Version":"v0.0.0-20251205161215-1948445e3318","Kind":"removal"} -->

root github.com/charmbracelet/ultraviolet：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/ultraviolet@v0.0.0-20251205161215-1948445e3318`。基线 v0.0.0-20251205161215-1948445e3318，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-ansi","Owner":"root","Path":"github.com/charmbracelet/x/ansi","Version":"v0.11.7","Kind":"removal"} -->

root github.com/charmbracelet/x/ansi：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/x/ansi@v0.11.7`。基线 v0.11.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-cellbuf","Owner":"root","Path":"github.com/charmbracelet/x/cellbuf","Version":"v0.0.13-0.20250311204145-2c3ea96c31dd","Kind":"removal"} -->

root github.com/charmbracelet/x/cellbuf：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/charmbracelet/x/cellbuf@v0.0.13-0.20250311204145-2c3ea96c31dd`。基线 v0.0.13-0.20250311204145-2c3ea96c31dd，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-exp-golden","Owner":"root","Path":"github.com/charmbracelet/x/exp/golden","Version":"v0.0.0-20250806222409-83e3a29d542f","Kind":"removal"} -->

root github.com/charmbracelet/x/exp/golden：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `charm.land/lipgloss/v2@v2.0.3 -> github.com/charmbracelet/x/exp/golden@v0.0.0-20250806222409-83e3a29d542f`。基线 v0.0.0-20250806222409-83e3a29d542f，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-term","Owner":"root","Path":"github.com/charmbracelet/x/term","Version":"v0.2.2","Kind":"removal"} -->

root github.com/charmbracelet/x/term：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/x/term@v0.2.2`。基线 v0.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-termios","Owner":"root","Path":"github.com/charmbracelet/x/termios","Version":"v0.1.1","Kind":"removal"} -->

root github.com/charmbracelet/x/termios：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/x/termios@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-charmbracelet-x-windows","Owner":"root","Path":"github.com/charmbracelet/x/windows","Version":"v0.2.2","Kind":"removal"} -->

root github.com/charmbracelet/x/windows：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/charmbracelet/x/windows@v0.2.2`。基线 v0.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-chzyer-logex","Owner":"root","Path":"github.com/chzyer/logex","Version":"v1.1.10","Kind":"removal"} -->

root github.com/chzyer/logex：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/google/pprof@v0.0.0-20210407192527-94a9f03dee38 -> github.com/chzyer/logex@v1.1.10`。基线 v1.1.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-chzyer-test","Owner":"root","Path":"github.com/chzyer/test","Version":"v0.0.0-20180213035817-a1ea475d72b1","Kind":"removal"} -->

root github.com/chzyer/test：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/google/pprof@v0.0.0-20210407192527-94a9f03dee38 -> github.com/chzyer/test@v0.0.0-20180213035817-a1ea475d72b1`。基线 v0.0.0-20180213035817-a1ea475d72b1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ckaznocha-intrange","Owner":"root","Path":"github.com/ckaznocha/intrange","Version":"v0.3.1","Kind":"removal"} -->

root github.com/ckaznocha/intrange：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ckaznocha/intrange@v0.3.1`。基线 v0.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-cli-browser","Owner":"root","Path":"github.com/cli/browser","Version":"v1.3.0","Kind":"removal"} -->

root github.com/cli/browser：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/cli/browser@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-clipperhouse-displaywidth","Owner":"root","Path":"github.com/clipperhouse/displaywidth","Version":"v0.11.0","Kind":"removal"} -->

root github.com/clipperhouse/displaywidth：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/clipperhouse/displaywidth@v0.11.0`。基线 v0.11.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-clipperhouse-stringish","Owner":"root","Path":"github.com/clipperhouse/stringish","Version":"v0.1.1","Kind":"removal"} -->

root github.com/clipperhouse/stringish：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/charmbracelet/colorprofile@v0.4.3 -> github.com/clipperhouse/stringish@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-clipperhouse-uax29-v2","Owner":"root","Path":"github.com/clipperhouse/uax29/v2","Version":"v2.7.0","Kind":"removal"} -->

root github.com/clipperhouse/uax29/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/clipperhouse/uax29/v2@v2.7.0`。基线 v2.7.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-containerd-typeurl-v2","Owner":"root","Path":"github.com/containerd/typeurl/v2","Version":"v2.2.0","Kind":"removal"} -->

root github.com/containerd/typeurl/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/containerd/errdefs/pkg@v0.3.0 -> github.com/containerd/typeurl/v2@v2.2.0`。基线 v2.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-coreos-go-semver","Owner":"root","Path":"github.com/coreos/go-semver","Version":"v0.3.1","Kind":"removal"} -->

root github.com/coreos/go-semver：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/coreos/go-semver@v0.3.1`。基线 v0.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-coreos-go-systemd-v22","Owner":"root","Path":"github.com/coreos/go-systemd/v22","Version":"v22.3.2","Kind":"removal"} -->

root github.com/coreos/go-systemd/v22：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/coreos/go-systemd/v22@v22.3.2`。基线 v22.3.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-cpuguy83-go-md2man-v2","Owner":"root","Path":"github.com/cpuguy83/go-md2man/v2","Version":"v2.0.7","Kind":"removal"} -->

root github.com/cpuguy83/go-md2man/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/cpuguy83/go-md2man/v2@v2.0.7`。基线 v2.0.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-creack-pty","Owner":"root","Path":"github.com/creack/pty","Version":"v1.1.24","Kind":"removal"} -->

root github.com/creack/pty：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/moby/moby/client@v0.5.0 -> github.com/creack/pty@v1.1.24`。基线 v1.1.24，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-cristalhq-acmd","Owner":"root","Path":"github.com/cristalhq/acmd","Version":"v0.12.0","Kind":"removal"} -->

root github.com/cristalhq/acmd：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/go-critic/go-critic@v0.14.3 -> github.com/cristalhq/acmd@v0.12.0`。基线 v0.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-cubicdaiya-gonp","Owner":"root","Path":"github.com/cubicdaiya/gonp","Version":"v1.0.4","Kind":"removal"} -->

root github.com/cubicdaiya/gonp：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/cubicdaiya/gonp@v1.0.4`。基线 v1.0.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-curioswitch-go-reassign","Owner":"root","Path":"github.com/curioswitch/go-reassign","Version":"v0.3.0","Kind":"removal"} -->

root github.com/curioswitch/go-reassign：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/curioswitch/go-reassign@v0.3.0`。基线 v0.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-daixiang0-gci","Owner":"root","Path":"github.com/daixiang0/gci","Version":"v0.13.7","Kind":"removal"} -->

root github.com/daixiang0/gci：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/daixiang0/gci@v0.13.7`。基线 v0.13.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-danieljoos-wincred","Owner":"root","Path":"github.com/danieljoos/wincred","Version":"v1.2.3","Kind":"removal"} -->

root github.com/danieljoos/wincred：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/docker/docker-credential-helpers@v0.9.8 -> github.com/danieljoos/wincred@v1.2.3`。基线 v1.2.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-dave-dst","Owner":"root","Path":"github.com/dave/dst","Version":"v0.27.3","Kind":"removal"} -->

root github.com/dave/dst：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/dave/dst@v0.27.3`。基线 v0.27.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-dave-jennifer","Owner":"root","Path":"github.com/dave/jennifer","Version":"v1.7.1","Kind":"removal"} -->

root github.com/dave/jennifer：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/dave/dst@v0.27.3 -> github.com/dave/jennifer@v1.5.0`。基线 v1.7.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-dchest-siphash","Owner":"root","Path":"github.com/dchest/siphash","Version":"v1.2.3","Kind":"removal"} -->

root github.com/dchest/siphash：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ncruces/go-sqlite3@v0.32.0 -> github.com/dchest/siphash@v1.2.3`。基线 v1.2.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-denis-tingaikin-go-header","Owner":"root","Path":"github.com/denis-tingaikin/go-header","Version":"v0.5.0","Kind":"removal"} -->

root github.com/denis-tingaikin/go-header：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/denis-tingaikin/go-header@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-docker-cli","Owner":"root","Path":"github.com/docker/cli","Version":"v29.6.1+incompatible","Kind":"removal"} -->

root github.com/docker/cli：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/docker/cli@v29.6.1+incompatible`。基线 v29.6.1+incompatible，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-docker-docker-credential-helpers","Owner":"root","Path":"github.com/docker/docker-credential-helpers","Version":"v0.9.8","Kind":"removal"} -->

root github.com/docker/docker-credential-helpers：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/docker/docker-credential-helpers@v0.9.8`。基线 v0.9.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-dprotaso-go-yit","Owner":"root","Path":"github.com/dprotaso/go-yit","Version":"v0.0.0-20220510233725-9ba8df137936","Kind":"removal"} -->

root github.com/dprotaso/go-yit：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/dprotaso/go-yit@v0.0.0-20220510233725-9ba8df137936`。基线 v0.0.0-20220510233725-9ba8df137936，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-erikgeiser-coninput","Owner":"root","Path":"github.com/erikgeiser/coninput","Version":"v0.0.0-20211004153227-1c3628e74d0f","Kind":"removal"} -->

root github.com/erikgeiser/coninput：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/erikgeiser/coninput@v0.0.0-20211004153227-1c3628e74d0f`。基线 v0.0.0-20211004153227-1c3628e74d0f，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ettle-strcase","Owner":"root","Path":"github.com/ettle/strcase","Version":"v0.2.0","Kind":"removal"} -->

root github.com/ettle/strcase：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ettle/strcase@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-expr-lang-expr","Owner":"root","Path":"github.com/expr-lang/expr","Version":"v1.17.7","Kind":"removal"} -->

root github.com/expr-lang/expr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/expr-lang/expr@v1.17.7`。基线 v1.17.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-fatih-color","Owner":"root","Path":"github.com/fatih/color","Version":"v1.19.0","Kind":"removal"} -->

root github.com/fatih/color：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/fatih/color@v1.19.0`。基线 v1.19.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-fatih-structtag","Owner":"root","Path":"github.com/fatih/structtag","Version":"v1.2.0","Kind":"removal"} -->

root github.com/fatih/structtag：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/fatih/structtag@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-firefart-nonamedreturns","Owner":"root","Path":"github.com/firefart/nonamedreturns","Version":"v1.0.6","Kind":"removal"} -->

root github.com/firefart/nonamedreturns：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/firefart/nonamedreturns@v1.0.6`。基线 v1.0.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-frankban-quicktest","Owner":"root","Path":"github.com/frankban/quicktest","Version":"v1.14.3","Kind":"removal"} -->

root github.com/frankban/quicktest：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/cast@v1.5.0 -> github.com/frankban/quicktest@v1.14.3`。基线 v1.14.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-fsnotify-fsnotify","Owner":"root","Path":"github.com/fsnotify/fsnotify","Version":"v1.5.4","Kind":"removal"} -->

root github.com/fsnotify/fsnotify：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/fsnotify/fsnotify@v1.5.4`。基线 v1.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-fzipp-gocyclo","Owner":"root","Path":"github.com/fzipp/gocyclo","Version":"v0.6.0","Kind":"removal"} -->

root github.com/fzipp/gocyclo：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/fzipp/gocyclo@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ghostiam-protogetter","Owner":"root","Path":"github.com/ghostiam/protogetter","Version":"v0.3.20","Kind":"removal"} -->

root github.com/ghostiam/protogetter：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ghostiam/protogetter@v0.3.20`。基线 v0.3.20，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-critic-go-critic","Owner":"root","Path":"github.com/go-critic/go-critic","Version":"v0.14.3","Kind":"removal"} -->

root github.com/go-critic/go-critic：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-critic/go-critic@v0.14.3`。基线 v0.14.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-quicktest-qt","Owner":"root","Path":"github.com/go-quicktest/qt","Version":"v1.101.0","Kind":"removal"} -->

root github.com/go-quicktest/qt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `mvdan.cc/gofumpt@v0.9.2 -> github.com/go-quicktest/qt@v1.101.0`。基线 v1.101.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-task-slim-sprig","Owner":"root","Path":"github.com/go-task/slim-sprig","Version":"v0.0.0-20210107165309-348f09dbbbc0","Kind":"removal"} -->

root github.com/go-task/slim-sprig：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/onsi/ginkgo@v1.16.4 -> github.com/go-task/slim-sprig@v0.0.0-20210107165309-348f09dbbbc0`。基线 v0.0.0-20210107165309-348f09dbbbc0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-task-slim-sprig-v3","Owner":"root","Path":"github.com/go-task/slim-sprig/v3","Version":"v3.0.0","Kind":"removal"} -->

root github.com/go-task/slim-sprig/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/go-task/slim-sprig/v3@v3.0.0`。基线 v3.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-astcast","Owner":"root","Path":"github.com/go-toolsmith/astcast","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/astcast：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/astcast@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-astcopy","Owner":"root","Path":"github.com/go-toolsmith/astcopy","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/astcopy：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/astcopy@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-astequal","Owner":"root","Path":"github.com/go-toolsmith/astequal","Version":"v1.2.0","Kind":"removal"} -->

root github.com/go-toolsmith/astequal：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/astequal@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-astfmt","Owner":"root","Path":"github.com/go-toolsmith/astfmt","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/astfmt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/astfmt@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-astp","Owner":"root","Path":"github.com/go-toolsmith/astp","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/astp：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/astp@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-pkgload","Owner":"root","Path":"github.com/go-toolsmith/pkgload","Version":"v1.2.2","Kind":"removal"} -->

root github.com/go-toolsmith/pkgload：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/go-critic/go-critic@v0.14.3 -> github.com/go-toolsmith/pkgload@v1.2.2`。基线 v1.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-strparse","Owner":"root","Path":"github.com/go-toolsmith/strparse","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/strparse：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/strparse@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-toolsmith-typep","Owner":"root","Path":"github.com/go-toolsmith/typep","Version":"v1.1.0","Kind":"removal"} -->

root github.com/go-toolsmith/typep：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-toolsmith/typep@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-go-xmlfmt-xmlfmt","Owner":"root","Path":"github.com/go-xmlfmt/xmlfmt","Version":"v1.1.3","Kind":"removal"} -->

root github.com/go-xmlfmt/xmlfmt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/go-xmlfmt/xmlfmt@v1.1.3`。基线 v1.1.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gobwas-glob","Owner":"root","Path":"github.com/gobwas/glob","Version":"v0.2.3","Kind":"removal"} -->

root github.com/gobwas/glob：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gobwas/glob@v0.2.3`。基线 v0.2.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-godoc-lint-godoc-lint","Owner":"root","Path":"github.com/godoc-lint/godoc-lint","Version":"v0.11.2","Kind":"removal"} -->

root github.com/godoc-lint/godoc-lint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/godoc-lint/godoc-lint@v0.11.2`。基线 v0.11.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gofrs-flock","Owner":"root","Path":"github.com/gofrs/flock","Version":"v0.13.0","Kind":"removal"} -->

root github.com/gofrs/flock：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gofrs/flock@v0.13.0`。基线 v0.13.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golang-groupcache","Owner":"root","Path":"github.com/golang/groupcache","Version":"v0.0.0-20210331224755-41bb18bfe9da","Kind":"removal"} -->

root github.com/golang/groupcache：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/golang/groupcache@v0.0.0-20210331224755-41bb18bfe9da`。基线 v0.0.0-20210331224755-41bb18bfe9da，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-asciicheck","Owner":"root","Path":"github.com/golangci/asciicheck","Version":"v0.5.0","Kind":"removal"} -->

root github.com/golangci/asciicheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/asciicheck@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-dupl","Owner":"root","Path":"github.com/golangci/dupl","Version":"v0.0.0-20260401084720-c99c5cf5c202","Kind":"removal"} -->

root github.com/golangci/dupl：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/dupl@v0.0.0-20260401084720-c99c5cf5c202`。基线 v0.0.0-20260401084720-c99c5cf5c202，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-go-printf-func-name","Owner":"root","Path":"github.com/golangci/go-printf-func-name","Version":"v0.1.1","Kind":"removal"} -->

root github.com/golangci/go-printf-func-name：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/go-printf-func-name@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-gofmt","Owner":"root","Path":"github.com/golangci/gofmt","Version":"v0.0.0-20250106114630-d62b90e6713d","Kind":"removal"} -->

root github.com/golangci/gofmt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/gofmt@v0.0.0-20250106114630-d62b90e6713d`。基线 v0.0.0-20250106114630-d62b90e6713d，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-golangci-lint-v2","Owner":"root","Path":"github.com/golangci/golangci-lint/v2","Version":"v2.12.2","Kind":"removal"} -->

root github.com/golangci/golangci-lint/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/golangci-lint/v2@v2.12.2`。基线 v2.12.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-golines","Owner":"root","Path":"github.com/golangci/golines","Version":"v0.15.0","Kind":"removal"} -->

root github.com/golangci/golines：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/golines@v0.15.0`。基线 v0.15.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-misspell","Owner":"root","Path":"github.com/golangci/misspell","Version":"v0.8.0","Kind":"removal"} -->

root github.com/golangci/misspell：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/misspell@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-plugin-module-register","Owner":"root","Path":"github.com/golangci/plugin-module-register","Version":"v0.1.2","Kind":"removal"} -->

root github.com/golangci/plugin-module-register：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/plugin-module-register@v0.1.2`。基线 v0.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-revgrep","Owner":"root","Path":"github.com/golangci/revgrep","Version":"v0.8.0","Kind":"removal"} -->

root github.com/golangci/revgrep：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/revgrep@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-rowserrcheck","Owner":"root","Path":"github.com/golangci/rowserrcheck","Version":"v0.0.0-20260419091836-c5f79b8a11ba","Kind":"removal"} -->

root github.com/golangci/rowserrcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/rowserrcheck@v0.0.0-20260419091836-c5f79b8a11ba`。基线 v0.0.0-20260419091836-c5f79b8a11ba，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-swaggoswag","Owner":"root","Path":"github.com/golangci/swaggoswag","Version":"v0.0.0-20250504205917-77f2aca3143e","Kind":"removal"} -->

root github.com/golangci/swaggoswag：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/swaggoswag@v0.0.0-20250504205917-77f2aca3143e`。基线 v0.0.0-20250504205917-77f2aca3143e，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-golangci-unconvert","Owner":"root","Path":"github.com/golangci/unconvert","Version":"v0.0.0-20250410112200-a129a6e6413e","Kind":"removal"} -->

root github.com/golangci/unconvert：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/golangci/unconvert@v0.0.0-20250410112200-a129a6e6413e`。基线 v0.0.0-20250410112200-a129a6e6413e，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-google-cel-go","Owner":"root","Path":"github.com/google/cel-go","Version":"v0.29.2","Kind":"removal"} -->

root github.com/google/cel-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/google/cel-go@v0.29.2`。基线 v0.29.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-google-go-containerregistry","Owner":"root","Path":"github.com/google/go-containerregistry","Version":"v0.21.7","Kind":"removal"} -->

root github.com/google/go-containerregistry：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/google/go-containerregistry@v0.21.7`。基线 v0.21.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-google-s2a-go","Owner":"root","Path":"github.com/google/s2a-go","Version":"v0.1.9","Kind":"removal"} -->

root github.com/google/s2a-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/google/s2a-go@v0.1.9`。基线 v0.1.9，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-googleapis-enterprise-certificate-proxy","Owner":"root","Path":"github.com/googleapis/enterprise-certificate-proxy","Version":"v0.3.6","Kind":"removal"} -->

root github.com/googleapis/enterprise-certificate-proxy：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/googleapis/enterprise-certificate-proxy@v0.3.6`。基线 v0.3.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-googleapis-gax-go-v2","Owner":"root","Path":"github.com/googleapis/gax-go/v2","Version":"v2.15.0","Kind":"removal"} -->

root github.com/googleapis/gax-go/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/googleapis/gax-go/v2@v2.15.0`。基线 v2.15.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gookit-color","Owner":"root","Path":"github.com/gookit/color","Version":"v1.6.0","Kind":"removal"} -->

root github.com/gookit/color：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/gookit/color@v1.6.0`。基线 v1.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gordonklaus-ineffassign","Owner":"root","Path":"github.com/gordonklaus/ineffassign","Version":"v0.2.0","Kind":"removal"} -->

root github.com/gordonklaus/ineffassign：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gordonklaus/ineffassign@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gostaticanalysis-analysisutil","Owner":"root","Path":"github.com/gostaticanalysis/analysisutil","Version":"v0.7.1","Kind":"removal"} -->

root github.com/gostaticanalysis/analysisutil：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gostaticanalysis/analysisutil@v0.7.1`。基线 v0.7.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gostaticanalysis-comment","Owner":"root","Path":"github.com/gostaticanalysis/comment","Version":"v1.5.0","Kind":"removal"} -->

root github.com/gostaticanalysis/comment：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gostaticanalysis/comment@v1.5.0`。基线 v1.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gostaticanalysis-forcetypeassert","Owner":"root","Path":"github.com/gostaticanalysis/forcetypeassert","Version":"v0.2.0","Kind":"removal"} -->

root github.com/gostaticanalysis/forcetypeassert：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gostaticanalysis/forcetypeassert@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gostaticanalysis-nilerr","Owner":"root","Path":"github.com/gostaticanalysis/nilerr","Version":"v0.1.2","Kind":"removal"} -->

root github.com/gostaticanalysis/nilerr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/gostaticanalysis/nilerr@v0.1.2`。基线 v0.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-gostaticanalysis-testutil","Owner":"root","Path":"github.com/gostaticanalysis/testutil","Version":"v0.5.0","Kind":"removal"} -->

root github.com/gostaticanalysis/testutil：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ckaznocha/intrange@v0.3.1 -> github.com/gostaticanalysis/testutil@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-consul-api","Owner":"root","Path":"github.com/hashicorp/consul/api","Version":"v1.12.0","Kind":"removal"} -->

root github.com/hashicorp/consul/api：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/consul/api@v1.12.0`。基线 v1.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-go-cleanhttp","Owner":"root","Path":"github.com/hashicorp/go-cleanhttp","Version":"v0.5.2","Kind":"removal"} -->

root github.com/hashicorp/go-cleanhttp：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/go-cleanhttp@v0.5.2`。基线 v0.5.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-go-hclog","Owner":"root","Path":"github.com/hashicorp/go-hclog","Version":"v1.2.0","Kind":"removal"} -->

root github.com/hashicorp/go-hclog：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/go-hclog@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-go-immutable-radix","Owner":"root","Path":"github.com/hashicorp/go-immutable-radix","Version":"v1.3.1","Kind":"removal"} -->

root github.com/hashicorp/go-immutable-radix：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/go-immutable-radix@v1.3.1`。基线 v1.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-go-immutable-radix-v2","Owner":"root","Path":"github.com/hashicorp/go-immutable-radix/v2","Version":"v2.1.0","Kind":"removal"} -->

root github.com/hashicorp/go-immutable-radix/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/hashicorp/go-immutable-radix/v2@v2.1.0`。基线 v2.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-go-rootcerts","Owner":"root","Path":"github.com/hashicorp/go-rootcerts","Version":"v1.0.2","Kind":"removal"} -->

root github.com/hashicorp/go-rootcerts：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/go-rootcerts@v1.0.2`。基线 v1.0.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-golang-lru","Owner":"root","Path":"github.com/hashicorp/golang-lru","Version":"v0.5.4","Kind":"removal"} -->

root github.com/hashicorp/golang-lru：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/golang-lru@v0.5.4`。基线 v0.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-hcl","Owner":"root","Path":"github.com/hashicorp/hcl","Version":"v1.0.0","Kind":"removal"} -->

root github.com/hashicorp/hcl：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/hashicorp/hcl@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hashicorp-serf","Owner":"root","Path":"github.com/hashicorp/serf","Version":"v0.9.7","Kind":"removal"} -->

root github.com/hashicorp/serf：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/hashicorp/serf@v0.9.7`。基线 v0.9.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hexops-gotextdiff","Owner":"root","Path":"github.com/hexops/gotextdiff","Version":"v1.0.3","Kind":"removal"} -->

root github.com/hexops/gotextdiff：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/hexops/gotextdiff@v1.0.3`。基线 v1.0.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-hpcloud-tail","Owner":"root","Path":"github.com/hpcloud/tail","Version":"v1.0.0","Kind":"removal"} -->

root github.com/hpcloud/tail：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/onsi/gomega@v1.7.0 -> github.com/hpcloud/tail@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-inconshreveable-mousetrap","Owner":"root","Path":"github.com/inconshreveable/mousetrap","Version":"v1.1.0","Kind":"removal"} -->

root github.com/inconshreveable/mousetrap：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/inconshreveable/mousetrap@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-invopop-jsonschema","Owner":"root","Path":"github.com/invopop/jsonschema","Version":"v0.13.0","Kind":"removal"} -->

root github.com/invopop/jsonschema：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/invopop/jsonschema@v0.13.0`。基线 v0.13.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jdx-go-netrc","Owner":"root","Path":"github.com/jdx/go-netrc","Version":"v1.0.0","Kind":"removal"} -->

root github.com/jdx/go-netrc：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/jdx/go-netrc@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jgautheron-goconst","Owner":"root","Path":"github.com/jgautheron/goconst","Version":"v1.10.0","Kind":"removal"} -->

root github.com/jgautheron/goconst：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/jgautheron/goconst@v1.10.0`。基线 v1.10.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jhump-protoreflect-v2","Owner":"root","Path":"github.com/jhump/protoreflect/v2","Version":"v2.0.0-beta.2","Kind":"removal"} -->

root github.com/jhump/protoreflect/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/bufbuild/buf@v1.72.0 -> github.com/jhump/protoreflect/v2@v2.0.0-beta.2`。基线 v2.0.0-beta.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jinzhu-inflection","Owner":"root","Path":"github.com/jinzhu/inflection","Version":"v1.0.0","Kind":"removal"} -->

root github.com/jinzhu/inflection：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/jinzhu/inflection@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jjti-go-spancheck","Owner":"root","Path":"github.com/jjti/go-spancheck","Version":"v0.6.5","Kind":"removal"} -->

root github.com/jjti/go-spancheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/jjti/go-spancheck@v0.6.5`。基线 v0.6.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jmoiron-sqlx","Owner":"root","Path":"github.com/jmoiron/sqlx","Version":"v1.3.5","Kind":"removal"} -->

root github.com/jmoiron/sqlx：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ryanrolds/sqlclosecheck@v0.6.0 -> github.com/jmoiron/sqlx@v1.3.5`。基线 v1.3.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-jordanlewis-gcassert","Owner":"root","Path":"github.com/jordanlewis/gcassert","Version":"v0.0.0-20250430164644-389ef753e22e","Kind":"removal"} -->

root github.com/jordanlewis/gcassert：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/quic-go/quic-go@v0.60.0 -> github.com/jordanlewis/gcassert@v0.0.0-20250430164644-389ef753e22e`。基线 v0.0.0-20250430164644-389ef753e22e，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-julz-importas","Owner":"root","Path":"github.com/julz/importas","Version":"v0.2.0","Kind":"removal"} -->

root github.com/julz/importas：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/julz/importas@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-karamaru-alpha-copyloopvar","Owner":"root","Path":"github.com/karamaru-alpha/copyloopvar","Version":"v1.2.2","Kind":"removal"} -->

root github.com/karamaru-alpha/copyloopvar：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/karamaru-alpha/copyloopvar@v1.2.2`。基线 v1.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-keybase-go-keychain","Owner":"root","Path":"github.com/keybase/go-keychain","Version":"v0.0.1","Kind":"removal"} -->

root github.com/keybase/go-keychain：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/docker/docker-credential-helpers@v0.9.8 -> github.com/keybase/go-keychain@v0.0.1`。基线 v0.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-kisielk-errcheck","Owner":"root","Path":"github.com/kisielk/errcheck","Version":"v1.10.0","Kind":"removal"} -->

root github.com/kisielk/errcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/kisielk/errcheck@v1.10.0`。基线 v1.10.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-kkHAIKE-contextcheck","Owner":"root","Path":"github.com/kkHAIKE/contextcheck","Version":"v1.1.6","Kind":"removal"} -->

root github.com/kkHAIKE/contextcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/kkHAIKE/contextcheck@v1.1.6`。基线 v1.1.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-klauspost-pgzip","Owner":"root","Path":"github.com/klauspost/pgzip","Version":"v1.2.6","Kind":"removal"} -->

root github.com/klauspost/pgzip：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/klauspost/pgzip@v1.2.6`。基线 v1.2.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-kulti-thelper","Owner":"root","Path":"github.com/kulti/thelper","Version":"v0.7.1","Kind":"removal"} -->

root github.com/kulti/thelper：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/kulti/thelper@v0.7.1`。基线 v0.7.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-kunwardeep-paralleltest","Owner":"root","Path":"github.com/kunwardeep/paralleltest","Version":"v1.0.15","Kind":"removal"} -->

root github.com/kunwardeep/paralleltest：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/kunwardeep/paralleltest@v1.0.15`。基线 v1.0.15，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-lasiar-canonicalheader","Owner":"root","Path":"github.com/lasiar/canonicalheader","Version":"v1.1.2","Kind":"removal"} -->

root github.com/lasiar/canonicalheader：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/lasiar/canonicalheader@v1.1.2`。基线 v1.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-exptostd","Owner":"root","Path":"github.com/ldez/exptostd","Version":"v0.4.5","Kind":"removal"} -->

root github.com/ldez/exptostd：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/exptostd@v0.4.5`。基线 v0.4.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-gomoddirectives","Owner":"root","Path":"github.com/ldez/gomoddirectives","Version":"v0.8.0","Kind":"removal"} -->

root github.com/ldez/gomoddirectives：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/gomoddirectives@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-grignotin","Owner":"root","Path":"github.com/ldez/grignotin","Version":"v0.10.1","Kind":"removal"} -->

root github.com/ldez/grignotin：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/grignotin@v0.10.1`。基线 v0.10.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-structtags","Owner":"root","Path":"github.com/ldez/structtags","Version":"v0.6.1","Kind":"removal"} -->

root github.com/ldez/structtags：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/structtags@v0.6.1`。基线 v0.6.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-tagliatelle","Owner":"root","Path":"github.com/ldez/tagliatelle","Version":"v0.7.2","Kind":"removal"} -->

root github.com/ldez/tagliatelle：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/tagliatelle@v0.7.2`。基线 v0.7.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ldez-usetesting","Owner":"root","Path":"github.com/ldez/usetesting","Version":"v0.5.0","Kind":"removal"} -->

root github.com/ldez/usetesting：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ldez/usetesting@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-leonklingele-grouper","Owner":"root","Path":"github.com/leonklingele/grouper","Version":"v1.1.2","Kind":"removal"} -->

root github.com/leonklingele/grouper：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/leonklingele/grouper@v1.1.2`。基线 v1.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-lib-pq","Owner":"root","Path":"github.com/lib/pq","Version":"v1.12.3","Kind":"removal"} -->

root github.com/lib/pq：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/lib/pq@v1.12.3`。基线 v1.12.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-lucasb-eyer-go-colorful","Owner":"root","Path":"github.com/lucasb-eyer/go-colorful","Version":"v1.4.0","Kind":"removal"} -->

root github.com/lucasb-eyer/go-colorful：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/lucasb-eyer/go-colorful@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-macabu-inamedparam","Owner":"root","Path":"github.com/macabu/inamedparam","Version":"v0.2.0","Kind":"removal"} -->

root github.com/macabu/inamedparam：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/macabu/inamedparam@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-magefile-mage","Owner":"root","Path":"github.com/magefile/mage","Version":"v1.14.0","Kind":"removal"} -->

root github.com/magefile/mage：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/curioswitch/go-reassign@v0.3.0 -> github.com/magefile/mage@v1.14.0`。基线 v1.14.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-manuelarte-embeddedstructfieldcheck","Owner":"root","Path":"github.com/manuelarte/embeddedstructfieldcheck","Version":"v0.4.0","Kind":"removal"} -->

root github.com/manuelarte/embeddedstructfieldcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/manuelarte/embeddedstructfieldcheck@v0.4.0`。基线 v0.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-manuelarte-funcorder","Owner":"root","Path":"github.com/manuelarte/funcorder","Version":"v0.6.0","Kind":"removal"} -->

root github.com/manuelarte/funcorder：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/manuelarte/funcorder@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-maratori-testableexamples","Owner":"root","Path":"github.com/maratori/testableexamples","Version":"v1.0.1","Kind":"removal"} -->

root github.com/maratori/testableexamples：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/maratori/testableexamples@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-maratori-testpackage","Owner":"root","Path":"github.com/maratori/testpackage","Version":"v1.1.2","Kind":"removal"} -->

root github.com/maratori/testpackage：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/maratori/testpackage@v1.1.2`。基线 v1.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-matoous-godox","Owner":"root","Path":"github.com/matoous/godox","Version":"v1.1.0","Kind":"removal"} -->

root github.com/matoous/godox：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/matoous/godox@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-matryer-is","Owner":"root","Path":"github.com/matryer/is","Version":"v1.4.0","Kind":"removal"} -->

root github.com/matryer/is：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/blizzy78/varnamelen@v0.8.0 -> github.com/matryer/is@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mattn-go-localereader","Owner":"root","Path":"github.com/mattn/go-localereader","Version":"v0.0.1","Kind":"removal"} -->

root github.com/mattn/go-localereader：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/mattn/go-localereader@v0.0.1`。基线 v0.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mattn-go-runewidth","Owner":"root","Path":"github.com/mattn/go-runewidth","Version":"v0.0.23","Kind":"removal"} -->

root github.com/mattn/go-runewidth：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/mattn/go-runewidth@v0.0.23`。基线 v0.0.23，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-matttproud-golang-protobuf-extensions","Owner":"root","Path":"github.com/matttproud/golang_protobuf_extensions","Version":"v1.0.1","Kind":"removal"} -->

root github.com/matttproud/golang_protobuf_extensions：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/golangci/golangci-lint/v2@v2.12.2 -> github.com/matttproud/golang_protobuf_extensions@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mgechev-dots","Owner":"root","Path":"github.com/mgechev/dots","Version":"v1.0.0","Kind":"removal"} -->

root github.com/mgechev/dots：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/mgechev/revive@v1.15.0 -> github.com/mgechev/dots@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mgechev-revive","Owner":"root","Path":"github.com/mgechev/revive","Version":"v1.15.0","Kind":"removal"} -->

root github.com/mgechev/revive：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/mgechev/revive@v1.15.0`。基线 v1.15.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mitchellh-go-homedir","Owner":"root","Path":"github.com/mitchellh/go-homedir","Version":"v1.1.0","Kind":"removal"} -->

root github.com/mitchellh/go-homedir：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/mitchellh/go-homedir@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mitchellh-mapstructure","Owner":"root","Path":"github.com/mitchellh/mapstructure","Version":"v1.5.0","Kind":"removal"} -->

root github.com/mitchellh/mapstructure：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/mitchellh/mapstructure@v1.5.0`。基线 v1.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-moricho-tparallel","Owner":"root","Path":"github.com/moricho/tparallel","Version":"v0.3.2","Kind":"removal"} -->

root github.com/moricho/tparallel：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/moricho/tparallel@v0.3.2`。基线 v0.3.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-mozilla-tls-observatory","Owner":"root","Path":"github.com/mozilla/tls-observatory","Version":"v0.0.0-20250923143331-eef96233227e","Kind":"removal"} -->

root github.com/mozilla/tls-observatory：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/mozilla/tls-observatory@v0.0.0-20250923143331-eef96233227e`。基线 v0.0.0-20250923143331-eef96233227e，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-muesli-ansi","Owner":"root","Path":"github.com/muesli/ansi","Version":"v0.0.0-20230316100256-276c6243b2f6","Kind":"removal"} -->

root github.com/muesli/ansi：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/muesli/ansi@v0.0.0-20230316100256-276c6243b2f6`。基线 v0.0.0-20230316100256-276c6243b2f6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-muesli-cancelreader","Owner":"root","Path":"github.com/muesli/cancelreader","Version":"v0.2.2","Kind":"removal"} -->

root github.com/muesli/cancelreader：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/muesli/cancelreader@v0.2.2`。基线 v0.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-muesli-termenv","Owner":"root","Path":"github.com/muesli/termenv","Version":"v0.16.0","Kind":"removal"} -->

root github.com/muesli/termenv：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/MirrexOne/unqueryvet@v1.5.4 -> github.com/muesli/termenv@v0.16.0`。基线 v0.16.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-nakabonne-nestif","Owner":"root","Path":"github.com/nakabonne/nestif","Version":"v0.3.1","Kind":"removal"} -->

root github.com/nakabonne/nestif：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/nakabonne/nestif@v0.3.1`。基线 v0.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ncruces-go-sqlite3","Owner":"root","Path":"github.com/ncruces/go-sqlite3","Version":"v0.32.0","Kind":"removal"} -->

root github.com/ncruces/go-sqlite3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ncruces/go-sqlite3@v0.32.0`。基线 v0.32.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ncruces-julianday","Owner":"root","Path":"github.com/ncruces/julianday","Version":"v1.0.0","Kind":"removal"} -->

root github.com/ncruces/julianday：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ncruces/julianday@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ncruces-sort","Owner":"root","Path":"github.com/ncruces/sort","Version":"v0.1.6","Kind":"removal"} -->

root github.com/ncruces/sort：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ncruces/go-sqlite3@v0.32.0 -> github.com/ncruces/sort@v0.1.6`。基线 v0.1.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ncruces-wbt","Owner":"root","Path":"github.com/ncruces/wbt","Version":"v1.0.0","Kind":"removal"} -->

root github.com/ncruces/wbt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ncruces/go-sqlite3@v0.32.0 -> github.com/ncruces/wbt@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-nishanths-exhaustive","Owner":"root","Path":"github.com/nishanths/exhaustive","Version":"v0.12.0","Kind":"removal"} -->

root github.com/nishanths/exhaustive：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/nishanths/exhaustive@v0.12.0`。基线 v0.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-nishanths-predeclared","Owner":"root","Path":"github.com/nishanths/predeclared","Version":"v0.2.2","Kind":"removal"} -->

root github.com/nishanths/predeclared：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/nishanths/predeclared@v0.2.2`。基线 v0.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-nunnatsa-ginkgolinter","Owner":"root","Path":"github.com/nunnatsa/ginkgolinter","Version":"v0.23.0","Kind":"removal"} -->

root github.com/nunnatsa/ginkgolinter：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/nunnatsa/ginkgolinter@v0.23.0`。基线 v0.23.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-nxadm-tail","Owner":"root","Path":"github.com/nxadm/tail","Version":"v1.4.8","Kind":"removal"} -->

root github.com/nxadm/tail：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/onsi/ginkgo@v1.16.4 -> github.com/nxadm/tail@v1.4.8`。基线 v1.4.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-oapi-codegen-oapi-codegen-v2","Owner":"root","Path":"github.com/oapi-codegen/oapi-codegen/v2","Version":"v2.8.0","Kind":"removal"} -->

root github.com/oapi-codegen/oapi-codegen/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/oapi-codegen/oapi-codegen/v2@v2.8.0`。基线 v2.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-onsi-ginkgo","Owner":"root","Path":"github.com/onsi/ginkgo","Version":"v1.16.4","Kind":"removal"} -->

root github.com/onsi/ginkgo：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/dprotaso/go-yit@v0.0.0-20220510233725-9ba8df137936 -> github.com/onsi/ginkgo@v1.16.4`。基线 v1.16.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-onsi-ginkgo-v2","Owner":"root","Path":"github.com/onsi/ginkgo/v2","Version":"v2.28.2","Kind":"removal"} -->

root github.com/onsi/ginkgo/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/onsi/ginkgo/v2@v2.28.2`。基线 v2.28.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-onsi-gomega","Owner":"root","Path":"github.com/onsi/gomega","Version":"v1.39.1","Kind":"removal"} -->

root github.com/onsi/gomega：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/dprotaso/go-yit@v0.0.0-20220510233725-9ba8df137936 -> github.com/onsi/gomega@v1.19.0`。基线 v1.39.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-openai-openai-go-v3","Owner":"root","Path":"github.com/openai/openai-go/v3","Version":"v3.32.0","Kind":"removal"} -->

root github.com/openai/openai-go/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/openai/openai-go/v3@v3.32.0`。基线 v3.32.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-otiai10-copy","Owner":"root","Path":"github.com/otiai10/copy","Version":"v1.14.0","Kind":"removal"} -->

root github.com/otiai10/copy：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ckaznocha/intrange@v0.3.1 -> github.com/otiai10/copy@v1.14.0`。基线 v1.14.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-otiai10-curr","Owner":"root","Path":"github.com/otiai10/curr","Version":"v1.0.0","Kind":"removal"} -->

root github.com/otiai10/curr：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/otiai10/mint@v1.3.1 -> github.com/otiai10/curr@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-otiai10-mint","Owner":"root","Path":"github.com/otiai10/mint","Version":"v1.3.1","Kind":"removal"} -->

root github.com/otiai10/mint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/otiai10/copy@v1.2.0 -> github.com/otiai10/mint@v1.3.1`。基线 v1.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pelletier-go-toml","Owner":"root","Path":"github.com/pelletier/go-toml","Version":"v1.9.5","Kind":"removal"} -->

root github.com/pelletier/go-toml：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pelletier/go-toml@v1.9.5`。基线 v1.9.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-petermattis-goid","Owner":"root","Path":"github.com/petermattis/goid","Version":"v0.0.0-20260716134002-a9b348f0a2b9","Kind":"removal"} -->

root github.com/petermattis/goid：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/petermattis/goid@v0.0.0-20260716134002-a9b348f0a2b9`。基线 v0.0.0-20260716134002-a9b348f0a2b9，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pganalyze-pg-query-go-v6","Owner":"root","Path":"github.com/pganalyze/pg_query_go/v6","Version":"v6.2.2","Kind":"removal"} -->

root github.com/pganalyze/pg_query_go/v6：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pganalyze/pg_query_go/v6@v6.2.2`。基线 v6.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-phayes-checkstyle","Owner":"root","Path":"github.com/phayes/checkstyle","Version":"v0.0.0-20170904204023-bfd46e6a821d","Kind":"removal"} -->

root github.com/phayes/checkstyle：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ryancurrah/gomodguard@v1.4.1 -> github.com/phayes/checkstyle@v0.0.0-20170904204023-bfd46e6a821d`。基线 v0.0.0-20170904204023-bfd46e6a821d，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pingcap-errors","Owner":"root","Path":"github.com/pingcap/errors","Version":"v0.11.5-0.20250523034308-74f78ae071ee","Kind":"removal"} -->

root github.com/pingcap/errors：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pingcap/errors@v0.11.5-0.20250523034308-74f78ae071ee`。基线 v0.11.5-0.20250523034308-74f78ae071ee，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pingcap-failpoint","Owner":"root","Path":"github.com/pingcap/failpoint","Version":"v0.0.0-20240528011301-b51a646c7c86","Kind":"removal"} -->

root github.com/pingcap/failpoint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pingcap/failpoint@v0.0.0-20240528011301-b51a646c7c86`。基线 v0.0.0-20240528011301-b51a646c7c86，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pingcap-log","Owner":"root","Path":"github.com/pingcap/log","Version":"v1.1.0","Kind":"removal"} -->

root github.com/pingcap/log：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pingcap/log@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-pingcap-tidb-pkg-parser","Owner":"root","Path":"github.com/pingcap/tidb/pkg/parser","Version":"v0.0.0-20260418072757-ce92298d1124","Kind":"removal"} -->

root github.com/pingcap/tidb/pkg/parser：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124`。基线 v0.0.0-20260418072757-ce92298d1124，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-protocolbuffers-protoscope","Owner":"root","Path":"github.com/protocolbuffers/protoscope","Version":"v0.0.0-20221109213918-8e7a6aafa2c9","Kind":"removal"} -->

root github.com/protocolbuffers/protoscope：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/bufbuild/protocompile@v0.14.2-0.20260716165721-bb5762d29672 -> github.com/protocolbuffers/protoscope@v0.0.0-20221109213918-8e7a6aafa2c9`。基线 v0.0.0-20221109213918-8e7a6aafa2c9，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-psanford-httpreadat","Owner":"root","Path":"github.com/psanford/httpreadat","Version":"v0.1.0","Kind":"removal"} -->

root github.com/psanford/httpreadat：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ncruces/go-sqlite3@v0.32.0 -> github.com/psanford/httpreadat@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-go-ruleguard","Owner":"root","Path":"github.com/quasilyte/go-ruleguard","Version":"v0.4.5","Kind":"removal"} -->

root github.com/quasilyte/go-ruleguard：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quasilyte/go-ruleguard@v0.4.5`。基线 v0.4.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-go-ruleguard-dsl","Owner":"root","Path":"github.com/quasilyte/go-ruleguard/dsl","Version":"v0.3.23","Kind":"removal"} -->

root github.com/quasilyte/go-ruleguard/dsl：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quasilyte/go-ruleguard/dsl@v0.3.23`。基线 v0.3.23，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-go-ruleguard-rules","Owner":"root","Path":"github.com/quasilyte/go-ruleguard/rules","Version":"v0.0.0-20211022131956-028d6511ab71","Kind":"removal"} -->

root github.com/quasilyte/go-ruleguard/rules：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/quasilyte/go-ruleguard@v0.4.5 -> github.com/quasilyte/go-ruleguard/rules@v0.0.0-20211022131956-028d6511ab71`。基线 v0.0.0-20211022131956-028d6511ab71，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-gogrep","Owner":"root","Path":"github.com/quasilyte/gogrep","Version":"v0.5.0","Kind":"removal"} -->

root github.com/quasilyte/gogrep：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quasilyte/gogrep@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-regex-syntax","Owner":"root","Path":"github.com/quasilyte/regex/syntax","Version":"v0.0.0-20210819130434-b3f0c404a727","Kind":"removal"} -->

root github.com/quasilyte/regex/syntax：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quasilyte/regex/syntax@v0.0.0-20210819130434-b3f0c404a727`。基线 v0.0.0-20210819130434-b3f0c404a727，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quasilyte-stdinfo","Owner":"root","Path":"github.com/quasilyte/stdinfo","Version":"v0.0.0-20220114132959-f7386bf02567","Kind":"removal"} -->

root github.com/quasilyte/stdinfo：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quasilyte/stdinfo@v0.0.0-20220114132959-f7386bf02567`。基线 v0.0.0-20220114132959-f7386bf02567，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quic-go-go-ossfuzz-seeds","Owner":"root","Path":"github.com/quic-go/go-ossfuzz-seeds","Version":"v0.1.0","Kind":"removal"} -->

root github.com/quic-go/go-ossfuzz-seeds：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/quic-go/quic-go@v0.60.0 -> github.com/quic-go/go-ossfuzz-seeds@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quic-go-qpack","Owner":"root","Path":"github.com/quic-go/qpack","Version":"v0.6.0","Kind":"removal"} -->

root github.com/quic-go/qpack：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quic-go/qpack@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-quic-go-quic-go","Owner":"root","Path":"github.com/quic-go/quic-go","Version":"v0.60.0","Kind":"removal"} -->

root github.com/quic-go/quic-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/quic-go/quic-go@v0.60.0`。基线 v0.60.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-raeperd-recvcheck","Owner":"root","Path":"github.com/raeperd/recvcheck","Version":"v0.2.0","Kind":"removal"} -->

root github.com/raeperd/recvcheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/raeperd/recvcheck@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-rivo-uniseg","Owner":"root","Path":"github.com/rivo/uniseg","Version":"v0.4.7","Kind":"removal"} -->

root github.com/rivo/uniseg：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/rivo/uniseg@v0.4.7`。基线 v0.4.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-riza-io-grpc-go","Owner":"root","Path":"github.com/riza-io/grpc-go","Version":"v0.2.0","Kind":"removal"} -->

root github.com/riza-io/grpc-go：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/riza-io/grpc-go@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-rodaine-protogofakeit","Owner":"root","Path":"github.com/rodaine/protogofakeit","Version":"v0.1.1","Kind":"removal"} -->

root github.com/rodaine/protogofakeit：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/go/protovalidate@v1.2.0 -> github.com/rodaine/protogofakeit@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-rs-cors","Owner":"root","Path":"github.com/rs/cors","Version":"v1.11.1","Kind":"removal"} -->

root github.com/rs/cors：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/rs/cors@v1.11.1`。基线 v1.11.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-russross-blackfriday","Owner":"root","Path":"github.com/russross/blackfriday","Version":"v1.6.0","Kind":"removal"} -->

root github.com/russross/blackfriday：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/opencontainers/image-spec@v1.1.1 -> github.com/russross/blackfriday@v1.6.0`。基线 v1.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ryancurrah-gomodguard","Owner":"root","Path":"github.com/ryancurrah/gomodguard","Version":"v1.4.1","Kind":"removal"} -->

root github.com/ryancurrah/gomodguard：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ryancurrah/gomodguard@v1.4.1`。基线 v1.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ryancurrah-gomodguard-v2","Owner":"root","Path":"github.com/ryancurrah/gomodguard/v2","Version":"v2.1.3","Kind":"removal"} -->

root github.com/ryancurrah/gomodguard/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ryancurrah/gomodguard/v2@v2.1.3`。基线 v2.1.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ryanrolds-sqlclosecheck","Owner":"root","Path":"github.com/ryanrolds/sqlclosecheck","Version":"v0.6.0","Kind":"removal"} -->

root github.com/ryanrolds/sqlclosecheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ryanrolds/sqlclosecheck@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sagikazarmark-crypt","Owner":"root","Path":"github.com/sagikazarmark/crypt","Version":"v0.6.0","Kind":"removal"} -->

root github.com/sagikazarmark/crypt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> github.com/sagikazarmark/crypt@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sanposhiho-wastedassign-v2","Owner":"root","Path":"github.com/sanposhiho/wastedassign/v2","Version":"v2.1.0","Kind":"removal"} -->

root github.com/sanposhiho/wastedassign/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sanposhiho/wastedassign/v2@v2.1.0`。基线 v2.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-santhosh-tekuri-jsonschema-v5","Owner":"root","Path":"github.com/santhosh-tekuri/jsonschema/v5","Version":"v5.3.1","Kind":"removal"} -->

root github.com/santhosh-tekuri/jsonschema/v5：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/opencontainers/image-spec@v1.1.1 -> github.com/santhosh-tekuri/jsonschema/v5@v5.3.1`。基线 v5.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sashamelentyev-interfacebloat","Owner":"root","Path":"github.com/sashamelentyev/interfacebloat","Version":"v1.1.0","Kind":"removal"} -->

root github.com/sashamelentyev/interfacebloat：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sashamelentyev/interfacebloat@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sashamelentyev-usestdlibvars","Owner":"root","Path":"github.com/sashamelentyev/usestdlibvars","Version":"v1.29.0","Kind":"removal"} -->

root github.com/sashamelentyev/usestdlibvars：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sashamelentyev/usestdlibvars@v1.29.0`。基线 v1.29.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-securego-gosec-v2","Owner":"root","Path":"github.com/securego/gosec/v2","Version":"v2.26.1","Kind":"removal"} -->

root github.com/securego/gosec/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/securego/gosec/v2@v2.26.1`。基线 v2.26.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-segmentio-encoding","Owner":"root","Path":"github.com/segmentio/encoding","Version":"v0.5.4","Kind":"removal"} -->

root github.com/segmentio/encoding：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/segmentio/encoding@v0.5.4`。基线 v0.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sergi-go-diff","Owner":"root","Path":"github.com/sergi/go-diff","Version":"v1.2.0","Kind":"removal"} -->

root github.com/sergi/go-diff：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/dave/dst@v0.27.3 -> github.com/sergi/go-diff@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sivchari-containedctx","Owner":"root","Path":"github.com/sivchari/containedctx","Version":"v1.0.3","Kind":"removal"} -->

root github.com/sivchari/containedctx：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sivchari/containedctx@v1.0.3`。基线 v1.0.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sonatard-noctx","Owner":"root","Path":"github.com/sonatard/noctx","Version":"v0.5.1","Kind":"removal"} -->

root github.com/sonatard/noctx：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sonatard/noctx@v0.5.1`。基线 v0.5.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sourcegraph-go-diff","Owner":"root","Path":"github.com/sourcegraph/go-diff","Version":"v0.8.0","Kind":"removal"} -->

root github.com/sourcegraph/go-diff：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sourcegraph/go-diff@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-speakeasy-api-jsonpath","Owner":"root","Path":"github.com/speakeasy-api/jsonpath","Version":"v0.6.3","Kind":"removal"} -->

root github.com/speakeasy-api/jsonpath：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/speakeasy-api/jsonpath@v0.6.3`。基线 v0.6.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-speakeasy-api-openapi","Owner":"root","Path":"github.com/speakeasy-api/openapi","Version":"v1.24.0","Kind":"removal"} -->

root github.com/speakeasy-api/openapi：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/speakeasy-api/openapi@v1.24.0`。基线 v1.24.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-afero","Owner":"root","Path":"github.com/spf13/afero","Version":"v1.15.0","Kind":"removal"} -->

root github.com/spf13/afero：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/afero@v1.15.0`。基线 v1.15.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-cast","Owner":"root","Path":"github.com/spf13/cast","Version":"v1.5.0","Kind":"removal"} -->

root github.com/spf13/cast：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/cast@v1.5.0`。基线 v1.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-cobra","Owner":"root","Path":"github.com/spf13/cobra","Version":"v1.10.2","Kind":"removal"} -->

root github.com/spf13/cobra：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/cobra@v1.10.2`。基线 v1.10.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-jwalterweatherman","Owner":"root","Path":"github.com/spf13/jwalterweatherman","Version":"v1.1.0","Kind":"removal"} -->

root github.com/spf13/jwalterweatherman：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/jwalterweatherman@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-pflag","Owner":"root","Path":"github.com/spf13/pflag","Version":"v1.0.10","Kind":"removal"} -->

root github.com/spf13/pflag：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/pflag@v1.0.10`。基线 v1.0.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-spf13-viper","Owner":"root","Path":"github.com/spf13/viper","Version":"v1.12.0","Kind":"removal"} -->

root github.com/spf13/viper：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/spf13/viper@v1.12.0`。基线 v1.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sqlc-dev-doubleclick","Owner":"root","Path":"github.com/sqlc-dev/doubleclick","Version":"v1.0.0","Kind":"removal"} -->

root github.com/sqlc-dev/doubleclick：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sqlc-dev/doubleclick@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-sqlc-dev-sqlc","Owner":"root","Path":"github.com/sqlc-dev/sqlc","Version":"v1.31.1","Kind":"removal"} -->

root github.com/sqlc-dev/sqlc：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/sqlc-dev/sqlc@v1.31.1`。基线 v1.31.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ssgreg-nlreturn-v2","Owner":"root","Path":"github.com/ssgreg/nlreturn/v2","Version":"v2.2.1","Kind":"removal"} -->

root github.com/ssgreg/nlreturn/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ssgreg/nlreturn/v2@v2.2.1`。基线 v2.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-stbenjam-no-sprintf-host-port","Owner":"root","Path":"github.com/stbenjam/no-sprintf-host-port","Version":"v0.3.1","Kind":"removal"} -->

root github.com/stbenjam/no-sprintf-host-port：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/stbenjam/no-sprintf-host-port@v0.3.1`。基线 v0.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-stoewer-go-strcase","Owner":"root","Path":"github.com/stoewer/go-strcase","Version":"v1.3.0","Kind":"removal"} -->

root github.com/stoewer/go-strcase：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/go/protoyaml@v0.7.0 -> github.com/stoewer/go-strcase@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-subosito-gotenv","Owner":"root","Path":"github.com/subosito/gotenv","Version":"v1.4.1","Kind":"removal"} -->

root github.com/subosito/gotenv：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/subosito/gotenv@v1.4.1`。基线 v1.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tenntenn-modver","Owner":"root","Path":"github.com/tenntenn/modver","Version":"v1.0.1","Kind":"removal"} -->

root github.com/tenntenn/modver：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ckaznocha/intrange@v0.3.1 -> github.com/tenntenn/modver@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tenntenn-text-transform","Owner":"root","Path":"github.com/tenntenn/text/transform","Version":"v0.0.0-20200319021203-7eef512accb3","Kind":"removal"} -->

root github.com/tenntenn/text/transform：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ckaznocha/intrange@v0.3.1 -> github.com/tenntenn/text/transform@v0.0.0-20200319021203-7eef512accb3`。基线 v0.0.0-20200319021203-7eef512accb3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tetafro-godot","Owner":"root","Path":"github.com/tetafro/godot","Version":"v1.5.6","Kind":"removal"} -->

root github.com/tetafro/godot：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/tetafro/godot@v1.5.6`。基线 v1.5.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tetratelabs-wazero","Owner":"root","Path":"github.com/tetratelabs/wazero","Version":"v1.12.0","Kind":"removal"} -->

root github.com/tetratelabs/wazero：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/tetratelabs/wazero@v1.12.0`。基线 v1.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tidwall-btree","Owner":"root","Path":"github.com/tidwall/btree","Version":"v1.8.1","Kind":"removal"} -->

root github.com/tidwall/btree：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/tidwall/btree@v1.8.1`。基线 v1.8.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tidwall-gjson","Owner":"root","Path":"github.com/tidwall/gjson","Version":"v1.18.0","Kind":"removal"} -->

root github.com/tidwall/gjson：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/tidwall/gjson@v1.18.0`。基线 v1.18.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tidwall-match","Owner":"root","Path":"github.com/tidwall/match","Version":"v1.1.1","Kind":"removal"} -->

root github.com/tidwall/match：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/tidwall/match@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tidwall-pretty","Owner":"root","Path":"github.com/tidwall/pretty","Version":"v1.2.1","Kind":"removal"} -->

root github.com/tidwall/pretty：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/tidwall/pretty@v1.2.1`。基线 v1.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tidwall-sjson","Owner":"root","Path":"github.com/tidwall/sjson","Version":"v1.2.5","Kind":"removal"} -->

root github.com/tidwall/sjson：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/tidwall/sjson@v1.2.5`。基线 v1.2.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-timakin-bodyclose","Owner":"root","Path":"github.com/timakin/bodyclose","Version":"v0.0.0-20260129054331-73d1f95b84b4","Kind":"removal"} -->

root github.com/timakin/bodyclose：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/timakin/bodyclose@v0.0.0-20260129054331-73d1f95b84b4`。基线 v0.0.0-20260129054331-73d1f95b84b4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-timandy-routine","Owner":"root","Path":"github.com/timandy/routine","Version":"v1.1.6","Kind":"removal"} -->

root github.com/timandy/routine：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `buf.build/go/protovalidate@v1.2.0 -> github.com/timandy/routine@v1.1.6`。基线 v1.1.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-timonwong-loggercheck","Owner":"root","Path":"github.com/timonwong/loggercheck","Version":"v0.11.0","Kind":"removal"} -->

root github.com/timonwong/loggercheck：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/timonwong/loggercheck@v0.11.0`。基线 v0.11.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tomarrell-wrapcheck-v2","Owner":"root","Path":"github.com/tomarrell/wrapcheck/v2","Version":"v2.12.0","Kind":"removal"} -->

root github.com/tomarrell/wrapcheck/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/tomarrell/wrapcheck/v2@v2.12.0`。基线 v2.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-tommy-muehle-go-mnd-v2","Owner":"root","Path":"github.com/tommy-muehle/go-mnd/v2","Version":"v2.5.1","Kind":"removal"} -->

root github.com/tommy-muehle/go-mnd/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/tommy-muehle/go-mnd/v2@v2.5.1`。基线 v2.5.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ultraware-funlen","Owner":"root","Path":"github.com/ultraware/funlen","Version":"v0.2.0","Kind":"removal"} -->

root github.com/ultraware/funlen：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ultraware/funlen@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ultraware-whitespace","Owner":"root","Path":"github.com/ultraware/whitespace","Version":"v0.2.0","Kind":"removal"} -->

root github.com/ultraware/whitespace：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ultraware/whitespace@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-uudashr-gocognit","Owner":"root","Path":"github.com/uudashr/gocognit","Version":"v1.2.1","Kind":"removal"} -->

root github.com/uudashr/gocognit：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/uudashr/gocognit@v1.2.1`。基线 v1.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-uudashr-iface","Owner":"root","Path":"github.com/uudashr/iface","Version":"v1.4.2","Kind":"removal"} -->

root github.com/uudashr/iface：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/uudashr/iface@v1.4.2`。基线 v1.4.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-valyala-quicktemplate","Owner":"root","Path":"github.com/valyala/quicktemplate","Version":"v1.8.0","Kind":"removal"} -->

root github.com/valyala/quicktemplate：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/golangci/golangci-lint/v2@v2.12.2 -> github.com/valyala/quicktemplate@v1.8.0`。基线 v1.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-vmware-labs-yaml-jsonpath","Owner":"root","Path":"github.com/vmware-labs/yaml-jsonpath","Version":"v0.3.2","Kind":"removal"} -->

root github.com/vmware-labs/yaml-jsonpath：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/vmware-labs/yaml-jsonpath@v0.3.2`。基线 v0.3.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-wasilibs-go-pgquery","Owner":"root","Path":"github.com/wasilibs/go-pgquery","Version":"v0.0.0-20250409022910-10ac41983c07","Kind":"removal"} -->

root github.com/wasilibs/go-pgquery：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/wasilibs/go-pgquery@v0.0.0-20250409022910-10ac41983c07`。基线 v0.0.0-20250409022910-10ac41983c07，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-wasilibs-wazero-helpers","Owner":"root","Path":"github.com/wasilibs/wazero-helpers","Version":"v0.0.0-20240620070341-3dff1577cd52","Kind":"removal"} -->

root github.com/wasilibs/wazero-helpers：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/wasilibs/wazero-helpers@v0.0.0-20240620070341-3dff1577cd52`。基线 v0.0.0-20240620070341-3dff1577cd52，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-wk8-go-ordered-map-v2","Owner":"root","Path":"github.com/wk8/go-ordered-map/v2","Version":"v2.1.8","Kind":"removal"} -->

root github.com/wk8/go-ordered-map/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> github.com/wk8/go-ordered-map/v2@v2.1.8`。基线 v2.1.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-xeipuuv-gojsonpointer","Owner":"root","Path":"github.com/xeipuuv/gojsonpointer","Version":"v0.0.0-20180127040702-4e3ac2762d5f","Kind":"removal"} -->

root github.com/xeipuuv/gojsonpointer：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/sqlc-dev/sqlc@v1.31.1 -> github.com/xeipuuv/gojsonpointer@v0.0.0-20180127040702-4e3ac2762d5f`。基线 v0.0.0-20180127040702-4e3ac2762d5f，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-xeipuuv-gojsonreference","Owner":"root","Path":"github.com/xeipuuv/gojsonreference","Version":"v0.0.0-20180127040603-bd5ef7bd5415","Kind":"removal"} -->

root github.com/xeipuuv/gojsonreference：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/sqlc-dev/sqlc@v1.31.1 -> github.com/xeipuuv/gojsonreference@v0.0.0-20180127040603-bd5ef7bd5415`。基线 v0.0.0-20180127040603-bd5ef7bd5415，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-xeipuuv-gojsonschema","Owner":"root","Path":"github.com/xeipuuv/gojsonschema","Version":"v1.2.0","Kind":"removal"} -->

root github.com/xeipuuv/gojsonschema：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/sqlc-dev/sqlc@v1.31.1 -> github.com/xeipuuv/gojsonschema@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-xen0n-gosmopolitan","Owner":"root","Path":"github.com/xen0n/gosmopolitan","Version":"v1.3.0","Kind":"removal"} -->

root github.com/xen0n/gosmopolitan：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/xen0n/gosmopolitan@v1.3.0`。基线 v1.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-xo-terminfo","Owner":"root","Path":"github.com/xo/terminfo","Version":"v0.0.0-20220910002029-abceb7e1c41e","Kind":"removal"} -->

root github.com/xo/terminfo：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/xo/terminfo@v0.0.0-20220910002029-abceb7e1c41e`。基线 v0.0.0-20220910002029-abceb7e1c41e，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-yagipy-maintidx","Owner":"root","Path":"github.com/yagipy/maintidx","Version":"v1.0.0","Kind":"removal"} -->

root github.com/yagipy/maintidx：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/yagipy/maintidx@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-yeya24-promlinter","Owner":"root","Path":"github.com/yeya24/promlinter","Version":"v0.3.0","Kind":"removal"} -->

root github.com/yeya24/promlinter：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/yeya24/promlinter@v0.3.0`。基线 v0.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-ykadowak-zerologlint","Owner":"root","Path":"github.com/ykadowak/zerologlint","Version":"v0.1.5","Kind":"removal"} -->

root github.com/ykadowak/zerologlint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> github.com/ykadowak/zerologlint@v0.1.5`。基线 v0.1.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-github-com-yuin-goldmark","Owner":"root","Path":"github.com/yuin/goldmark","Version":"v1.4.13","Kind":"removal"} -->

root github.com/yuin/goldmark：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `go.uber.org/mock@v0.6.0 -> github.com/yuin/goldmark@v1.4.13`。基线 v1.4.13，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gitlab-com-bosi-decorder","Owner":"root","Path":"gitlab.com/bosi/decorder","Version":"v0.4.2","Kind":"removal"} -->

root gitlab.com/bosi/decorder：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> gitlab.com/bosi/decorder@v0.4.2`。基线 v0.4.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-simpler-org-assert","Owner":"root","Path":"go-simpler.org/assert","Version":"v0.9.0","Kind":"removal"} -->

root go-simpler.org/assert：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `go-simpler.org/musttag@v0.14.0 -> go-simpler.org/assert@v0.9.0`。基线 v0.9.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-simpler-org-musttag","Owner":"root","Path":"go-simpler.org/musttag","Version":"v0.14.0","Kind":"removal"} -->

root go-simpler.org/musttag：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go-simpler.org/musttag@v0.14.0`。基线 v0.14.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-simpler-org-sloglint","Owner":"root","Path":"go-simpler.org/sloglint","Version":"v0.12.0","Kind":"removal"} -->

root go-simpler.org/sloglint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go-simpler.org/sloglint@v0.12.0`。基线 v0.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-augendre-info-arangolint","Owner":"root","Path":"go.augendre.info/arangolint","Version":"v0.4.0","Kind":"removal"} -->

root go.augendre.info/arangolint：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.augendre.info/arangolint@v0.4.0`。基线 v0.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-augendre-info-fatcontext","Owner":"root","Path":"go.augendre.info/fatcontext","Version":"v0.9.0","Kind":"removal"} -->

root go.augendre.info/fatcontext：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.augendre.info/fatcontext@v0.9.0`。基线 v0.9.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-etcd-io-etcd-api-v3","Owner":"root","Path":"go.etcd.io/etcd/api/v3","Version":"v3.5.4","Kind":"removal"} -->

root go.etcd.io/etcd/api/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> go.etcd.io/etcd/api/v3@v3.5.4`。基线 v3.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-etcd-io-etcd-client-pkg-v3","Owner":"root","Path":"go.etcd.io/etcd/client/pkg/v3","Version":"v3.5.4","Kind":"removal"} -->

root go.etcd.io/etcd/client/pkg/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> go.etcd.io/etcd/client/pkg/v3@v3.5.4`。基线 v3.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-etcd-io-etcd-client-v2","Owner":"root","Path":"go.etcd.io/etcd/client/v2","Version":"v2.305.4","Kind":"removal"} -->

root go.etcd.io/etcd/client/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> go.etcd.io/etcd/client/v2@v2.305.4`。基线 v2.305.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-etcd-io-etcd-client-v3","Owner":"root","Path":"go.etcd.io/etcd/client/v3","Version":"v3.5.4","Kind":"removal"} -->

root go.etcd.io/etcd/client/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> go.etcd.io/etcd/client/v3@v3.5.4`。基线 v3.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-lsp-dev-jsonrpc2","Owner":"root","Path":"go.lsp.dev/jsonrpc2","Version":"v0.10.0","Kind":"removal"} -->

root go.lsp.dev/jsonrpc2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.lsp.dev/jsonrpc2@v0.10.0`。基线 v0.10.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-lsp-dev-pkg","Owner":"root","Path":"go.lsp.dev/pkg","Version":"v0.0.0-20210717090340-384b27a52fb2","Kind":"removal"} -->

root go.lsp.dev/pkg：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.lsp.dev/pkg@v0.0.0-20210717090340-384b27a52fb2`。基线 v0.0.0-20210717090340-384b27a52fb2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-lsp-dev-protocol","Owner":"root","Path":"go.lsp.dev/protocol","Version":"v0.12.0","Kind":"removal"} -->

root go.lsp.dev/protocol：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.lsp.dev/protocol@v0.12.0`。基线 v0.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-lsp-dev-uri","Owner":"root","Path":"go.lsp.dev/uri","Version":"v0.3.0","Kind":"removal"} -->

root go.lsp.dev/uri：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> go.lsp.dev/uri@v0.3.0`。基线 v0.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-go-opencensus-io","Owner":"root","Path":"go.opencensus.io","Version":"v0.23.0","Kind":"removal"} -->

root go.opencensus.io：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> go.opencensus.io@v0.23.0`。基线 v0.23.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-golang-org-x-exp-typeparams","Owner":"root","Path":"golang.org/x/exp/typeparams","Version":"v0.0.0-20260209203927-2842357ff358","Kind":"removal"} -->

root golang.org/x/exp/typeparams：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> golang.org/x/exp/typeparams@v0.0.0-20260209203927-2842357ff358`。基线 v0.0.0-20260209203927-2842357ff358，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-golang-org-x-telemetry","Owner":"root","Path":"golang.org/x/telemetry","Version":"v0.0.0-20260708182218-49f421fb7959","Kind":"removal"} -->

root golang.org/x/telemetry：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `golang.org/x/tools@v0.48.0 -> golang.org/x/telemetry@v0.0.0-20260708182218-49f421fb7959`。基线 v0.0.0-20260708182218-49f421fb7959，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-google-golang-org-api","Owner":"root","Path":"google.golang.org/api","Version":"v0.81.0","Kind":"removal"} -->

root google.golang.org/api：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/spf13/viper@v1.12.0 -> google.golang.org/api@v0.81.0`。基线 v0.81.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-google-golang-org-genai","Owner":"root","Path":"google.golang.org/genai","Version":"v1.54.0","Kind":"removal"} -->

root google.golang.org/genai：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/securego/gosec/v2@v2.26.1 -> google.golang.org/genai@v1.54.0`。基线 v1.54.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gopkg-in-alecthomas-kingpin-v2","Owner":"root","Path":"gopkg.in/alecthomas/kingpin.v2","Version":"v2.2.6","Kind":"removal"} -->

root gopkg.in/alecthomas/kingpin.v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/yeya24/promlinter@v0.3.0 -> gopkg.in/alecthomas/kingpin.v2@v2.2.6`。基线 v2.2.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gopkg-in-fsnotify-v1","Owner":"root","Path":"gopkg.in/fsnotify.v1","Version":"v1.4.7","Kind":"removal"} -->

root gopkg.in/fsnotify.v1：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/onsi/gomega@v1.7.0 -> gopkg.in/fsnotify.v1@v1.4.7`。基线 v1.4.7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gopkg-in-natefinch-lumberjack-v2","Owner":"root","Path":"gopkg.in/natefinch/lumberjack.v2","Version":"v2.2.1","Kind":"removal"} -->

root gopkg.in/natefinch/lumberjack.v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> gopkg.in/natefinch/lumberjack.v2@v2.2.1`。基线 v2.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gopkg-in-src-d-go-billy-v4","Owner":"root","Path":"gopkg.in/src-d/go-billy.v4","Version":"v4.3.2","Kind":"removal"} -->

root gopkg.in/src-d/go-billy.v4：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/dave/dst@v0.27.3 -> gopkg.in/src-d/go-billy.v4@v4.3.2`。基线 v4.3.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gopkg-in-tomb-v1","Owner":"root","Path":"gopkg.in/tomb.v1","Version":"v1.0.0-20141024135613-dd632973f1e7","Kind":"removal"} -->

root gopkg.in/tomb.v1：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/nxadm/tail@v1.4.8 -> gopkg.in/tomb.v1@v1.0.0-20141024135613-dd632973f1e7`。基线 v1.0.0-20141024135613-dd632973f1e7，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-gotest-tools-v3","Owner":"root","Path":"gotest.tools/v3","Version":"v3.5.2","Kind":"removal"} -->

root gotest.tools/v3：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/moby/moby/api@v1.55.0 -> gotest.tools/v3@v3.5.2`。基线 v3.5.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-lukechampine-com-adiantum","Owner":"root","Path":"lukechampine.com/adiantum","Version":"v1.1.1","Kind":"removal"} -->

root lukechampine.com/adiantum：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/ncruces/go-sqlite3@v0.32.0 -> lukechampine.com/adiantum@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-modernc-org-golex","Owner":"root","Path":"modernc.org/golex","Version":"v1.1.0","Kind":"removal"} -->

root modernc.org/golex：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/golex@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-modernc-org-parser","Owner":"root","Path":"modernc.org/parser","Version":"v1.1.0","Kind":"removal"} -->

root modernc.org/parser：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/parser@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-modernc-org-y","Owner":"root","Path":"modernc.org/y","Version":"v1.1.0","Kind":"removal"} -->

root modernc.org/y：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/pingcap/tidb/pkg/parser@v0.0.0-20260418072757-ce92298d1124 -> modernc.org/y@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-mvdan-cc-gofumpt","Owner":"root","Path":"mvdan.cc/gofumpt","Version":"v0.9.2","Kind":"removal"} -->

root mvdan.cc/gofumpt：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> mvdan.cc/gofumpt@v0.9.2`。基线 v0.9.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-mvdan-cc-unparam","Owner":"root","Path":"mvdan.cc/unparam","Version":"v0.0.0-20251027182757-5beb8c8f8f15","Kind":"removal"} -->

root mvdan.cc/unparam：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> mvdan.cc/unparam@v0.0.0-20251027182757-5beb8c8f8f15`。基线 v0.0.0-20251027182757-5beb8c8f8f15，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-mvdan-cc-xurls-v2","Owner":"root","Path":"mvdan.cc/xurls/v2","Version":"v2.6.0","Kind":"removal"} -->

root mvdan.cc/xurls/v2：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> mvdan.cc/xurls/v2@v2.6.0`。基线 v2.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-pgregory-net-rapid","Owner":"root","Path":"pgregory.net/rapid","Version":"v1.2.0","Kind":"removal"} -->

root pgregory.net/rapid：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `github.com/moby/moby/api@v1.55.0 -> pgregory.net/rapid@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-root-pluginrpc-com-pluginrpc","Owner":"root","Path":"pluginrpc.com/pluginrpc","Version":"v0.5.0","Kind":"removal"} -->

root pluginrpc.com/pluginrpc：本归属入口无可达链，保留在 tools；见路径证据同 owner/path 的完整链，末边 `talenro.local/devtools -> pluginrpc.com/pluginrpc@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-dario-cat-mergo","Owner":"tools","Path":"dario.cat/mergo","Version":"v1.0.2","Kind":"removal"} -->

tools dario.cat/mergo：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> dario.cat/mergo@v1.0.2`。基线 v1.0.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Azure-azure-sdk-for-go-sdk-azcore","Owner":"tools","Path":"github.com/Azure/azure-sdk-for-go/sdk/azcore","Version":"v1.21.0","Kind":"removal"} -->

tools github.com/Azure/azure-sdk-for-go/sdk/azcore：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.21.0`。基线 v1.21.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Azure-azure-sdk-for-go-sdk-azidentity","Owner":"tools","Path":"github.com/Azure/azure-sdk-for-go/sdk/azidentity","Version":"v1.13.1","Kind":"removal"} -->

tools github.com/Azure/azure-sdk-for-go/sdk/azidentity：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.13.1`。基线 v1.13.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Azure-azure-sdk-for-go-sdk-internal","Owner":"tools","Path":"github.com/Azure/azure-sdk-for-go/sdk/internal","Version":"v1.11.2","Kind":"removal"} -->

tools github.com/Azure/azure-sdk-for-go/sdk/internal：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/Azure/azure-sdk-for-go/sdk/internal@v1.11.2`。基线 v1.11.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Azure-azure-sdk-for-go-sdk-security-keyvault-azkeys","Owner":"tools","Path":"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys","Version":"v1.4.0","Kind":"removal"} -->

tools github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Azure-azure-sdk-for-go-sdk-security-keyvault-internal","Owner":"tools","Path":"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal","Version":"v1.2.0","Kind":"removal"} -->

tools github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-AzureAD-microsoft-authentication-library-for-go","Owner":"tools","Path":"github.com/AzureAD/microsoft-authentication-library-for-go","Version":"v1.6.0","Kind":"removal"} -->

tools github.com/AzureAD/microsoft-authentication-library-for-go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/AzureAD/microsoft-authentication-library-for-go@v1.6.0`。基线 v1.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ClickHouse-ch-go","Owner":"tools","Path":"github.com/ClickHouse/ch-go","Version":"v0.71.0","Kind":"removal"} -->

tools github.com/ClickHouse/ch-go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ClickHouse/ch-go@v0.71.0`。基线 v0.71.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ClickHouse-clickhouse-go-v2","Owner":"tools","Path":"github.com/ClickHouse/clickhouse-go/v2","Version":"v2.45.0","Kind":"removal"} -->

tools github.com/ClickHouse/clickhouse-go/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ClickHouse/clickhouse-go/v2@v2.45.0`。基线 v2.45.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-CloudyKit-fastprinter","Owner":"tools","Path":"github.com/CloudyKit/fastprinter","Version":"v0.0.0-20200109182630-33d98a066a53","Kind":"removal"} -->

tools github.com/CloudyKit/fastprinter：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/CloudyKit/fastprinter@v0.0.0-20200109182630-33d98a066a53`。基线 v0.0.0-20200109182630-33d98a066a53，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-CloudyKit-jet-v6","Owner":"tools","Path":"github.com/CloudyKit/jet/v6","Version":"v6.2.0","Kind":"removal"} -->

tools github.com/CloudyKit/jet/v6：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/CloudyKit/jet/v6@v6.2.0`。基线 v6.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Joker-jade","Owner":"tools","Path":"github.com/Joker/jade","Version":"v1.1.3","Kind":"removal"} -->

tools github.com/Joker/jade：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/Joker/jade@v1.1.3`。基线 v1.1.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-RaveNoX-go-jsoncommentstrip","Owner":"tools","Path":"github.com/RaveNoX/go-jsoncommentstrip","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/RaveNoX/go-jsoncommentstrip：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/apapsch/go-jsonmerge/v2@v2.0.0 -> github.com/RaveNoX/go-jsoncommentstrip@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-Shopify-goreferrer","Owner":"tools","Path":"github.com/Shopify/goreferrer","Version":"v0.0.0-20220729165902-8cddb4f5de06","Kind":"removal"} -->

tools github.com/Shopify/goreferrer：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/Shopify/goreferrer@v0.0.0-20220729165902-8cddb4f5de06`。基线 v0.0.0-20220729165902-8cddb4f5de06，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-andybalholm-brotli","Owner":"tools","Path":"github.com/andybalholm/brotli","Version":"v1.2.1","Kind":"removal"} -->

tools github.com/andybalholm/brotli：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/andybalholm/brotli@v1.2.1`。基线 v1.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-antihax-optional","Owner":"tools","Path":"github.com/antihax/optional","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/antihax/optional：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/grpc-ecosystem/grpc-gateway@v1.16.0 -> github.com/antihax/optional@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-apapsch-go-jsonmerge-v2","Owner":"tools","Path":"github.com/apapsch/go-jsonmerge/v2","Version":"v2.0.0","Kind":"removal"} -->

tools github.com/apapsch/go-jsonmerge/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/apapsch/go-jsonmerge/v2@v2.0.0`。基线 v2.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-aymerick-douceur","Owner":"tools","Path":"github.com/aymerick/douceur","Version":"v0.2.0","Kind":"removal"} -->

tools github.com/aymerick/douceur：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/aymerick/douceur@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-bmatcuk-doublestar","Owner":"tools","Path":"github.com/bmatcuk/doublestar","Version":"v1.1.1","Kind":"removal"} -->

tools github.com/bmatcuk/doublestar：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/apapsch/go-jsonmerge/v2@v2.0.0 -> github.com/bmatcuk/doublestar@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-boombuler-barcode","Owner":"tools","Path":"github.com/boombuler/barcode","Version":"v1.0.1-0.20190219062509-6c824513bacc","Kind":"removal"} -->

tools github.com/boombuler/barcode：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/boombuler/barcode@v1.0.1-0.20190219062509-6c824513bacc`。基线 v1.0.1-0.20190219062509-6c824513bacc，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-bsm-ginkgo-v2","Owner":"tools","Path":"github.com/bsm/ginkgo/v2","Version":"v2.12.0","Kind":"removal"} -->

tools github.com/bsm/ginkgo/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/redis/go-redis/v9@v9.22.0 -> github.com/bsm/ginkgo/v2@v2.12.0`。基线 v2.12.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-bsm-gomega","Owner":"tools","Path":"github.com/bsm/gomega","Version":"v1.27.10","Kind":"removal"} -->

tools github.com/bsm/gomega：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/redis/go-redis/v9@v9.22.0 -> github.com/bsm/gomega@v1.27.10`。基线 v1.27.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-bytedance-sonic","Owner":"tools","Path":"github.com/bytedance/sonic","Version":"v1.11.6","Kind":"removal"} -->

tools github.com/bytedance/sonic：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/bytedance/sonic@v1.11.6`。基线 v1.11.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-bytedance-sonic-loader","Owner":"tools","Path":"github.com/bytedance/sonic/loader","Version":"v0.1.1","Kind":"removal"} -->

tools github.com/bytedance/sonic/loader：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/bytedance/sonic/loader@v0.1.1`。基线 v0.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-cenkalti-backoff-v4","Owner":"tools","Path":"github.com/cenkalti/backoff/v4","Version":"v4.3.0","Kind":"removal"} -->

tools github.com/cenkalti/backoff/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/ch-go@v0.71.0 -> github.com/cenkalti/backoff/v4@v4.3.0`。基线 v4.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-census-instrumentation-opencensus-proto","Owner":"tools","Path":"github.com/census-instrumentation/opencensus-proto","Version":"v0.2.1","Kind":"removal"} -->

tools github.com/census-instrumentation/opencensus-proto：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/envoyproxy/go-control-plane@v0.10.2-0.20220325020618-49ff273808a1 -> github.com/census-instrumentation/opencensus-proto@v0.2.1`。基线 v0.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-client9-misspell","Owner":"tools","Path":"github.com/client9/misspell","Version":"v0.3.4","Kind":"removal"} -->

tools github.com/client9/misspell：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `google.golang.org/grpc@v1.23.0 -> github.com/client9/misspell@v0.3.4`。基线 v0.3.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-cloudwego-base64x","Owner":"tools","Path":"github.com/cloudwego/base64x","Version":"v0.1.4","Kind":"removal"} -->

tools github.com/cloudwego/base64x：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/cloudwego/base64x@v0.1.4`。基线 v0.1.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-cloudwego-iasm","Owner":"tools","Path":"github.com/cloudwego/iasm","Version":"v0.2.0","Kind":"removal"} -->

tools github.com/cloudwego/iasm：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/cloudwego/iasm@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-cncf-udpa-go","Owner":"tools","Path":"github.com/cncf/udpa/go","Version":"v0.0.0-20210930031921-04548b0d99d4","Kind":"removal"} -->

tools github.com/cncf/udpa/go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `google.golang.org/grpc@v1.47.0 -> github.com/cncf/udpa/go@v0.0.0-20210930031921-04548b0d99d4`。基线 v0.0.0-20210930031921-04548b0d99d4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-coder-websocket","Owner":"tools","Path":"github.com/coder/websocket","Version":"v1.8.14","Kind":"removal"} -->

tools github.com/coder/websocket：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/coder/websocket@v1.8.14`。基线 v1.8.14，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-containerd-log","Owner":"tools","Path":"github.com/containerd/log","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/containerd/log：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/containerd/log@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-containerd-platforms","Owner":"tools","Path":"github.com/containerd/platforms","Version":"v0.2.1","Kind":"removal"} -->

tools github.com/containerd/platforms：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/containerd/platforms@v0.2.1`。基线 v0.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-cpuguy83-dockercfg","Owner":"tools","Path":"github.com/cpuguy83/dockercfg","Version":"v0.3.2","Kind":"removal"} -->

tools github.com/cpuguy83/dockercfg：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/cpuguy83/dockercfg@v0.3.2`。基线 v0.3.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-dmarkham-enumer","Owner":"tools","Path":"github.com/dmarkham/enumer","Version":"v1.6.3","Kind":"removal"} -->

tools github.com/dmarkham/enumer：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/ch-go@v0.71.0 -> github.com/dmarkham/enumer@v1.6.3`。基线 v1.6.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-docker-docker","Owner":"tools","Path":"github.com/docker/docker","Version":"v28.5.2+incompatible","Kind":"removal"} -->

tools github.com/docker/docker：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/docker/docker@v28.5.2+incompatible`。基线 v28.5.2+incompatible，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-dustin-go-humanize","Owner":"tools","Path":"github.com/dustin/go-humanize","Version":"v1.0.1","Kind":"removal"} -->

tools github.com/dustin/go-humanize：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/dustin/go-humanize@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-elastic-go-sysinfo","Owner":"tools","Path":"github.com/elastic/go-sysinfo","Version":"v1.15.4","Kind":"removal"} -->

tools github.com/elastic/go-sysinfo：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/elastic/go-sysinfo@v1.15.4`。基线 v1.15.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-elastic-go-windows","Owner":"tools","Path":"github.com/elastic/go-windows","Version":"v1.0.2","Kind":"removal"} -->

tools github.com/elastic/go-windows：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/elastic/go-windows@v1.0.2`。基线 v1.0.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-fatih-structs","Owner":"tools","Path":"github.com/fatih/structs","Version":"v1.1.0","Kind":"removal"} -->

tools github.com/fatih/structs：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/fatih/structs@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-flosch-pongo2-v4","Owner":"tools","Path":"github.com/flosch/pongo2/v4","Version":"v4.0.2","Kind":"removal"} -->

tools github.com/flosch/pongo2/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/flosch/pongo2/v4@v4.0.2`。基线 v4.0.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-fxamacker-cbor-v2","Owner":"tools","Path":"github.com/fxamacker/cbor/v2","Version":"v2.9.2","Kind":"removal"} -->

tools github.com/fxamacker/cbor/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/fxamacker/cbor/v2@v2.9.2`。基线 v2.9.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gabriel-vasile-mimetype","Owner":"tools","Path":"github.com/gabriel-vasile/mimetype","Version":"v1.4.3","Kind":"removal"} -->

tools github.com/gabriel-vasile/mimetype：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/gabriel-vasile/mimetype@v1.4.3`。基线 v1.4.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ghodss-yaml","Owner":"tools","Path":"github.com/ghodss/yaml","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/ghodss/yaml：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/grpc-ecosystem/grpc-gateway@v1.16.0 -> github.com/ghodss/yaml@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gin-contrib-sse","Owner":"tools","Path":"github.com/gin-contrib/sse","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/gin-contrib/sse：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/gin-contrib/sse@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gin-gonic-gin","Owner":"tools","Path":"github.com/gin-gonic/gin","Version":"v1.10.1","Kind":"removal"} -->

tools github.com/gin-gonic/gin：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/gin-gonic/gin@v1.10.1`。基线 v1.10.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-faster-city","Owner":"tools","Path":"github.com/go-faster/city","Version":"v1.0.1","Kind":"removal"} -->

tools github.com/go-faster/city：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/go-faster/city@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-faster-errors","Owner":"tools","Path":"github.com/go-faster/errors","Version":"v0.7.1","Kind":"removal"} -->

tools github.com/go-faster/errors：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/go-faster/errors@v0.7.1`。基线 v0.7.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-playground-locales","Owner":"tools","Path":"github.com/go-playground/locales","Version":"v0.14.1","Kind":"removal"} -->

tools github.com/go-playground/locales：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/go-playground/locales@v0.14.1`。基线 v0.14.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-playground-universal-translator","Owner":"tools","Path":"github.com/go-playground/universal-translator","Version":"v0.18.1","Kind":"removal"} -->

tools github.com/go-playground/universal-translator：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/go-playground/universal-translator@v0.18.1`。基线 v0.18.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-playground-validator-v10","Owner":"tools","Path":"github.com/go-playground/validator/v10","Version":"v10.20.0","Kind":"removal"} -->

tools github.com/go-playground/validator/v10：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/go-playground/validator/v10@v10.20.0`。基线 v10.20.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-webauthn-webauthn","Owner":"tools","Path":"github.com/go-webauthn/webauthn","Version":"v0.17.4","Kind":"removal"} -->

tools github.com/go-webauthn/webauthn：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/go-webauthn/webauthn@v0.17.4`。基线 v0.17.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-go-webauthn-x","Owner":"tools","Path":"github.com/go-webauthn/x","Version":"v0.2.6","Kind":"removal"} -->

tools github.com/go-webauthn/x：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/go-webauthn/x@v0.2.6`。基线 v0.2.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-goccy-go-json","Owner":"tools","Path":"github.com/goccy/go-json","Version":"v0.10.2","Kind":"removal"} -->

tools github.com/goccy/go-json：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/goccy/go-json@v0.10.2`。基线 v0.10.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-jwt-jwt-v4","Owner":"tools","Path":"github.com/golang-jwt/jwt/v4","Version":"v4.5.2","Kind":"removal"} -->

tools github.com/golang-jwt/jwt/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/golang-jwt/jwt/v4@v4.5.2`。基线 v4.5.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-jwt-jwt-v5","Owner":"tools","Path":"github.com/golang-jwt/jwt/v5","Version":"v5.3.1","Kind":"removal"} -->

tools github.com/golang-jwt/jwt/v5：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/golang-jwt/jwt/v5@v5.3.1`。基线 v5.3.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-sql-civil","Owner":"tools","Path":"github.com/golang-sql/civil","Version":"v0.0.0-20220223132316-b832511892a9","Kind":"removal"} -->

tools github.com/golang-sql/civil：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/golang-sql/civil@v0.0.0-20220223132316-b832511892a9`。基线 v0.0.0-20220223132316-b832511892a9，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-sql-sqlexp","Owner":"tools","Path":"github.com/golang-sql/sqlexp","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/golang-sql/sqlexp：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/golang-sql/sqlexp@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-mock","Owner":"tools","Path":"github.com/golang/mock","Version":"v1.1.1","Kind":"removal"} -->

tools github.com/golang/mock：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `google.golang.org/grpc@v1.25.1 -> github.com/golang/mock@v1.1.1`。基线 v1.1.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-golang-snappy","Owner":"tools","Path":"github.com/golang/snappy","Version":"v0.0.4","Kind":"removal"} -->

tools github.com/golang/snappy：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/golang/snappy@v0.0.4`。基线 v0.0.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gomarkdown-markdown","Owner":"tools","Path":"github.com/gomarkdown/markdown","Version":"v0.0.0-20240328165702-4d01890c35c0","Kind":"removal"} -->

tools github.com/gomarkdown/markdown：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/gomarkdown/markdown@v0.0.0-20240328165702-4d01890c35c0`。基线 v0.0.0-20240328165702-4d01890c35c0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-google-go-tpm","Owner":"tools","Path":"github.com/google/go-tpm","Version":"v0.9.8","Kind":"removal"} -->

tools github.com/google/go-tpm：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/google/go-tpm@v0.9.8`。基线 v0.9.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-google-go-tpm-tools","Owner":"tools","Path":"github.com/google/go-tpm-tools","Version":"v0.3.13-0.20230620182252-4639ecce2aba","Kind":"removal"} -->

tools github.com/google/go-tpm-tools：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/google/go-tpm@v0.9.8 -> github.com/google/go-tpm-tools@v0.3.13-0.20230620182252-4639ecce2aba`。基线 v0.3.13-0.20230620182252-4639ecce2aba，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gorilla-css","Owner":"tools","Path":"github.com/gorilla/css","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/gorilla/css：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/gorilla/css@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-gowebpki-jcs","Owner":"tools","Path":"github.com/gowebpki/jcs","Version":"v1.0.1","Kind":"removal"} -->

tools github.com/gowebpki/jcs：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/gowebpki/jcs@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-grpc-ecosystem-grpc-gateway","Owner":"tools","Path":"github.com/grpc-ecosystem/grpc-gateway","Version":"v1.16.0","Kind":"removal"} -->

tools github.com/grpc-ecosystem/grpc-gateway：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `go.opentelemetry.io/proto/otlp@v0.7.0 -> github.com/grpc-ecosystem/grpc-gateway@v1.16.0`。基线 v1.16.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-iris-contrib-schema","Owner":"tools","Path":"github.com/iris-contrib/schema","Version":"v0.0.6","Kind":"removal"} -->

tools github.com/iris-contrib/schema：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/iris-contrib/schema@v0.0.6`。基线 v0.0.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-aescts-v2","Owner":"tools","Path":"github.com/jcmturner/aescts/v2","Version":"v2.0.0","Kind":"removal"} -->

tools github.com/jcmturner/aescts/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/aescts/v2@v2.0.0`。基线 v2.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-dnsutils-v2","Owner":"tools","Path":"github.com/jcmturner/dnsutils/v2","Version":"v2.0.0","Kind":"removal"} -->

tools github.com/jcmturner/dnsutils/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/dnsutils/v2@v2.0.0`。基线 v2.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-gofork","Owner":"tools","Path":"github.com/jcmturner/gofork","Version":"v1.7.6","Kind":"removal"} -->

tools github.com/jcmturner/gofork：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/gofork@v1.7.6`。基线 v1.7.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-goidentity-v6","Owner":"tools","Path":"github.com/jcmturner/goidentity/v6","Version":"v6.0.1","Kind":"removal"} -->

tools github.com/jcmturner/goidentity/v6：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/goidentity/v6@v6.0.1`。基线 v6.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-gokrb5-v8","Owner":"tools","Path":"github.com/jcmturner/gokrb5/v8","Version":"v8.4.4","Kind":"removal"} -->

tools github.com/jcmturner/gokrb5/v8：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/gokrb5/v8@v8.4.4`。基线 v8.4.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jcmturner-rpc-v2","Owner":"tools","Path":"github.com/jcmturner/rpc/v2","Version":"v2.0.3","Kind":"removal"} -->

tools github.com/jcmturner/rpc/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/jcmturner/rpc/v2@v2.0.3`。基线 v2.0.3，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jessevdk-go-flags","Owner":"tools","Path":"github.com/jessevdk/go-flags","Version":"v1.4.0","Kind":"removal"} -->

tools github.com/jessevdk/go-flags：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `howett.net/plist@v1.0.1 -> github.com/jessevdk/go-flags@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-joeshaw-multierror","Owner":"tools","Path":"github.com/joeshaw/multierror","Version":"v0.0.0-20140124173710-69b34d4ec901","Kind":"removal"} -->

tools github.com/joeshaw/multierror：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/elastic/go-sysinfo@v1.8.1 -> github.com/joeshaw/multierror@v0.0.0-20140124173710-69b34d4ec901`。基线 v0.0.0-20140124173710-69b34d4ec901，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-joho-godotenv","Owner":"tools","Path":"github.com/joho/godotenv","Version":"v1.5.1","Kind":"removal"} -->

tools github.com/joho/godotenv：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/joho/godotenv@v1.5.1`。基线 v1.5.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-jonboulle-clockwork","Owner":"tools","Path":"github.com/jonboulle/clockwork","Version":"v0.5.0","Kind":"removal"} -->

tools github.com/jonboulle/clockwork：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/jonboulle/clockwork@v0.5.0`。基线 v0.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-josharian-intern","Owner":"tools","Path":"github.com/josharian/intern","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/josharian/intern：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/josharian/intern@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-juju-gnuflag","Owner":"tools","Path":"github.com/juju/gnuflag","Version":"v0.0.0-20171113085948-2ce1bb71843d","Kind":"removal"} -->

tools github.com/juju/gnuflag：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/apapsch/go-jsonmerge/v2@v2.0.0 -> github.com/juju/gnuflag@v0.0.0-20171113085948-2ce1bb71843d`。基线 v0.0.0-20171113085948-2ce1bb71843d，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-blocks","Owner":"tools","Path":"github.com/kataras/blocks","Version":"v0.0.8","Kind":"removal"} -->

tools github.com/kataras/blocks：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/blocks@v0.0.8`。基线 v0.0.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-golog","Owner":"tools","Path":"github.com/kataras/golog","Version":"v0.1.11","Kind":"removal"} -->

tools github.com/kataras/golog：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/golog@v0.1.11`。基线 v0.1.11，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-iris-v12","Owner":"tools","Path":"github.com/kataras/iris/v12","Version":"v12.2.11","Kind":"removal"} -->

tools github.com/kataras/iris/v12：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/iris/v12@v12.2.11`。基线 v12.2.11，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-pio","Owner":"tools","Path":"github.com/kataras/pio","Version":"v0.0.13","Kind":"removal"} -->

tools github.com/kataras/pio：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/pio@v0.0.13`。基线 v0.0.13，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-sitemap","Owner":"tools","Path":"github.com/kataras/sitemap","Version":"v0.0.6","Kind":"removal"} -->

tools github.com/kataras/sitemap：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/sitemap@v0.0.6`。基线 v0.0.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-kataras-tunnel","Owner":"tools","Path":"github.com/kataras/tunnel","Version":"v0.0.4","Kind":"removal"} -->

tools github.com/kataras/tunnel：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/kataras/tunnel@v0.0.4`。基线 v0.0.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-klauspost-cpuid-v2","Owner":"tools","Path":"github.com/klauspost/cpuid/v2","Version":"v2.2.10","Kind":"removal"} -->

tools github.com/klauspost/cpuid/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/klauspost/cpuid/v2@v2.2.7`。基线 v2.2.10，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-labstack-echo-v4","Owner":"tools","Path":"github.com/labstack/echo/v4","Version":"v4.15.1","Kind":"removal"} -->

tools github.com/labstack/echo/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/labstack/echo/v4@v4.15.1`。基线 v4.15.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-labstack-gommon","Owner":"tools","Path":"github.com/labstack/gommon","Version":"v0.4.2","Kind":"removal"} -->

tools github.com/labstack/gommon：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/labstack/gommon@v0.4.2`。基线 v0.4.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-leodido-go-urn","Owner":"tools","Path":"github.com/leodido/go-urn","Version":"v1.4.0","Kind":"removal"} -->

tools github.com/leodido/go-urn：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/leodido/go-urn@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-mailgun-raymond-v2","Owner":"tools","Path":"github.com/mailgun/raymond/v2","Version":"v2.0.48","Kind":"removal"} -->

tools github.com/mailgun/raymond/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/mailgun/raymond/v2@v2.0.48`。基线 v2.0.48，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-mfridman-interpolate","Owner":"tools","Path":"github.com/mfridman/interpolate","Version":"v0.0.2","Kind":"removal"} -->

tools github.com/mfridman/interpolate：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/mfridman/interpolate@v0.0.2`。基线 v0.0.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-mfridman-xflag","Owner":"tools","Path":"github.com/mfridman/xflag","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/mfridman/xflag：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/mfridman/xflag@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-microcosm-cc-bluemonday","Owner":"tools","Path":"github.com/microcosm-cc/bluemonday","Version":"v1.0.26","Kind":"removal"} -->

tools github.com/microcosm-cc/bluemonday：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/microcosm-cc/bluemonday@v1.0.26`。基线 v1.0.26，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-microsoft-go-mssqldb","Owner":"tools","Path":"github.com/microsoft/go-mssqldb","Version":"v1.9.8","Kind":"removal"} -->

tools github.com/microsoft/go-mssqldb：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/microsoft/go-mssqldb@v1.9.8`。基线 v1.9.8，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-mkevac-debugcharts","Owner":"tools","Path":"github.com/mkevac/debugcharts","Version":"v0.0.0-20191222103121-ae1c48aa8615","Kind":"removal"} -->

tools github.com/mkevac/debugcharts：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/mkevac/debugcharts@v0.0.0-20191222103121-ae1c48aa8615`。基线 v0.0.0-20191222103121-ae1c48aa8615，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-moby-go-archive","Owner":"tools","Path":"github.com/moby/go-archive","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/moby/go-archive：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/go-archive@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-moby-patternmatcher","Owner":"tools","Path":"github.com/moby/patternmatcher","Version":"v0.6.0","Kind":"removal"} -->

tools github.com/moby/patternmatcher：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/patternmatcher@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-moby-sys-sequential","Owner":"tools","Path":"github.com/moby/sys/sequential","Version":"v0.6.0","Kind":"removal"} -->

tools github.com/moby/sys/sequential：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/sys/sequential@v0.6.0`。基线 v0.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-moby-sys-user","Owner":"tools","Path":"github.com/moby/sys/user","Version":"v0.4.0","Kind":"removal"} -->

tools github.com/moby/sys/user：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/sys/user@v0.4.0`。基线 v0.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-moby-sys-userns","Owner":"tools","Path":"github.com/moby/sys/userns","Version":"v0.1.0","Kind":"removal"} -->

tools github.com/moby/sys/userns：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/moby/sys/userns@v0.1.0`。基线 v0.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-morikuni-aec","Owner":"tools","Path":"github.com/morikuni/aec","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/morikuni/aec：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/morikuni/aec@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-nats-io-nats-go","Owner":"tools","Path":"github.com/nats-io/nats.go","Version":"v1.52.0","Kind":"removal"} -->

tools github.com/nats-io/nats.go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/nats-io/nats.go@v1.52.0`。基线 v1.52.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-nats-io-nkeys","Owner":"tools","Path":"github.com/nats-io/nkeys","Version":"v0.4.15","Kind":"removal"} -->

tools github.com/nats-io/nkeys：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/nats-io/nkeys@v0.4.15`。基线 v0.4.15，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-nats-io-nuid","Owner":"tools","Path":"github.com/nats-io/nuid","Version":"v1.0.1","Kind":"removal"} -->

tools github.com/nats-io/nuid：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/nats-io/nuid@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ncruces-go-strftime","Owner":"tools","Path":"github.com/ncruces/go-strftime","Version":"v1.0.0","Kind":"removal"} -->

tools github.com/ncruces/go-strftime：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ncruces/go-strftime@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-oapi-codegen-nullable","Owner":"tools","Path":"github.com/oapi-codegen/nullable","Version":"v1.1.0","Kind":"removal"} -->

tools github.com/oapi-codegen/nullable：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/oapi-codegen/nullable@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-oapi-codegen-runtime","Owner":"tools","Path":"github.com/oapi-codegen/runtime","Version":"v1.6.0","Kind":"removal"} -->

tools github.com/oapi-codegen/runtime：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/oapi-codegen/runtime@v1.6.0`。基线 v1.6.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-pascaldekloe-name","Owner":"tools","Path":"github.com/pascaldekloe/name","Version":"v1.0.1","Kind":"removal"} -->

tools github.com/pascaldekloe/name：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/ch-go@v0.71.0 -> github.com/pascaldekloe/name@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-paulmach-orb","Owner":"tools","Path":"github.com/paulmach/orb","Version":"v0.13.0","Kind":"removal"} -->

tools github.com/paulmach/orb：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/paulmach/orb@v0.13.0`。基线 v0.13.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-paulmach-protoscan","Owner":"tools","Path":"github.com/paulmach/protoscan","Version":"v0.2.1","Kind":"removal"} -->

tools github.com/paulmach/protoscan：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/paulmach/orb@v0.13.0 -> github.com/paulmach/protoscan@v0.2.1`。基线 v0.2.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-philhofer-fwd","Owner":"tools","Path":"github.com/philhofer/fwd","Version":"v1.2.0","Kind":"removal"} -->

tools github.com/philhofer/fwd：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/philhofer/fwd@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-pierrec-lz4-v4","Owner":"tools","Path":"github.com/pierrec/lz4/v4","Version":"v4.1.26","Kind":"removal"} -->

tools github.com/pierrec/lz4/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/pierrec/lz4/v4@v4.1.26`。基线 v4.1.26，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-pkg-browser","Owner":"tools","Path":"github.com/pkg/browser","Version":"v0.0.0-20240102092130-5ac0b6a4141c","Kind":"removal"} -->

tools github.com/pkg/browser：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/microsoft/go-mssqldb@v1.9.8 -> github.com/pkg/browser@v0.0.0-20240102092130-5ac0b6a4141c`。基线 v0.0.0-20240102092130-5ac0b6a4141c，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-pquerna-otp","Owner":"tools","Path":"github.com/pquerna/otp","Version":"v1.5.0","Kind":"removal"} -->

tools github.com/pquerna/otp：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/pquerna/otp@v1.5.0`。基线 v1.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-pressly-goose-v3","Owner":"tools","Path":"github.com/pressly/goose/v3","Version":"v3.27.1","Kind":"removal"} -->

tools github.com/pressly/goose/v3：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/pressly/goose/v3@v3.27.1`。基线 v3.27.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-redis-go-redis-v9","Owner":"tools","Path":"github.com/redis/go-redis/v9","Version":"v9.22.0","Kind":"removal"} -->

tools github.com/redis/go-redis/v9：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/redis/go-redis/v9@v9.22.0`。基线 v9.22.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-rekby-fixenv","Owner":"tools","Path":"github.com/rekby/fixenv","Version":"v0.6.1","Kind":"removal"} -->

tools github.com/rekby/fixenv：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ydb-platform/ydb-go-sdk/v3@v3.135.0 -> github.com/rekby/fixenv@v0.6.1`。基线 v0.6.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-rogpeppe-fastuuid","Owner":"tools","Path":"github.com/rogpeppe/fastuuid","Version":"v1.2.0","Kind":"removal"} -->

tools github.com/rogpeppe/fastuuid：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/grpc-ecosystem/grpc-gateway@v1.16.0 -> github.com/rogpeppe/fastuuid@v1.2.0`。基线 v1.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-schollz-closestmatch","Owner":"tools","Path":"github.com/schollz/closestmatch","Version":"v2.1.0+incompatible","Kind":"removal"} -->

tools github.com/schollz/closestmatch：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/schollz/closestmatch@v2.1.0+incompatible`。基线 v2.1.0+incompatible，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-sethvargo-go-retry","Owner":"tools","Path":"github.com/sethvargo/go-retry","Version":"v0.3.0","Kind":"removal"} -->

tools github.com/sethvargo/go-retry：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/sethvargo/go-retry@v0.3.0`。基线 v0.3.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-shirou-gopsutil","Owner":"tools","Path":"github.com/shirou/gopsutil","Version":"v3.21.11+incompatible","Kind":"removal"} -->

tools github.com/shirou/gopsutil：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/shirou/gopsutil@v3.21.11+incompatible`。基线 v3.21.11+incompatible，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-shopspring-decimal","Owner":"tools","Path":"github.com/shopspring/decimal","Version":"v1.4.0","Kind":"removal"} -->

tools github.com/shopspring/decimal：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/shopspring/decimal@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-spkg-bom","Owner":"tools","Path":"github.com/spkg/bom","Version":"v0.0.0-20160624110644-59b7046e48ad","Kind":"removal"} -->

tools github.com/spkg/bom：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/apapsch/go-jsonmerge/v2@v2.0.0 -> github.com/spkg/bom@v0.0.0-20160624110644-59b7046e48ad`。基线 v0.0.0-20160624110644-59b7046e48ad，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-tdewolff-minify-v2","Owner":"tools","Path":"github.com/tdewolff/minify/v2","Version":"v2.20.19","Kind":"removal"} -->

tools github.com/tdewolff/minify/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/tdewolff/minify/v2@v2.20.19`。基线 v2.20.19，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-tdewolff-parse-v2","Owner":"tools","Path":"github.com/tdewolff/parse/v2","Version":"v2.7.12","Kind":"removal"} -->

tools github.com/tdewolff/parse/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/tdewolff/parse/v2@v2.7.12`。基线 v2.7.12，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-testcontainers-testcontainers-go","Owner":"tools","Path":"github.com/testcontainers/testcontainers-go","Version":"v0.40.0","Kind":"removal"} -->

tools github.com/testcontainers/testcontainers-go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> github.com/testcontainers/testcontainers-go@v0.40.0`。基线 v0.40.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-tinylib-msgp","Owner":"tools","Path":"github.com/tinylib/msgp","Version":"v1.6.4","Kind":"removal"} -->

tools github.com/tinylib/msgp：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/tinylib/msgp@v1.6.4`。基线 v1.6.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-tursodatabase-libsql-client-go","Owner":"tools","Path":"github.com/tursodatabase/libsql-client-go","Version":"v0.0.0-20251219100830-236aa1ff8acc","Kind":"removal"} -->

tools github.com/tursodatabase/libsql-client-go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/tursodatabase/libsql-client-go@v0.0.0-20251219100830-236aa1ff8acc`。基线 v0.0.0-20251219100830-236aa1ff8acc，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-twitchyliquid64-golang-asm","Owner":"tools","Path":"github.com/twitchyliquid64/golang-asm","Version":"v0.15.1","Kind":"removal"} -->

tools github.com/twitchyliquid64/golang-asm：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/twitchyliquid64/golang-asm@v0.15.1`。基线 v0.15.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ugorji-go-codec","Owner":"tools","Path":"github.com/ugorji/go/codec","Version":"v1.2.12","Kind":"removal"} -->

tools github.com/ugorji/go/codec：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/ugorji/go/codec@v1.2.12`。基线 v1.2.12，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-valyala-fasttemplate","Owner":"tools","Path":"github.com/valyala/fasttemplate","Version":"v1.2.2","Kind":"removal"} -->

tools github.com/valyala/fasttemplate：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/valyala/fasttemplate@v1.2.2`。基线 v1.2.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-vertica-vertica-sql-go","Owner":"tools","Path":"github.com/vertica/vertica-sql-go","Version":"v1.3.6","Kind":"removal"} -->

tools github.com/vertica/vertica-sql-go：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/vertica/vertica-sql-go@v1.3.6`。基线 v1.3.6，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-vmihailenco-msgpack-v5","Owner":"tools","Path":"github.com/vmihailenco/msgpack/v5","Version":"v5.4.1","Kind":"removal"} -->

tools github.com/vmihailenco/msgpack/v5：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/vmihailenco/msgpack/v5@v5.4.1`。基线 v5.4.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-vmihailenco-tagparser-v2","Owner":"tools","Path":"github.com/vmihailenco/tagparser/v2","Version":"v2.0.0","Kind":"removal"} -->

tools github.com/vmihailenco/tagparser/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/vmihailenco/tagparser/v2@v2.0.0`。基线 v2.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-x448-float16","Owner":"tools","Path":"github.com/x448/float16","Version":"v0.8.4","Kind":"removal"} -->

tools github.com/x448/float16：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/x448/float16@v0.8.4`。基线 v0.8.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-xyproto-randomstring","Owner":"tools","Path":"github.com/xyproto/randomstring","Version":"v1.0.5","Kind":"removal"} -->

tools github.com/xyproto/randomstring：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/andybalholm/brotli@v1.2.1 -> github.com/xyproto/randomstring@v1.0.5`。基线 v1.0.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ydb-platform-ydb-go-genproto","Owner":"tools","Path":"github.com/ydb-platform/ydb-go-genproto","Version":"v0.0.0-20260311095541-ebbf792c1180","Kind":"removal"} -->

tools github.com/ydb-platform/ydb-go-genproto：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ydb-platform/ydb-go-genproto@v0.0.0-20260311095541-ebbf792c1180`。基线 v0.0.0-20260311095541-ebbf792c1180，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ydb-platform-ydb-go-sdk-v3","Owner":"tools","Path":"github.com/ydb-platform/ydb-go-sdk/v3","Version":"v3.135.0","Kind":"removal"} -->

tools github.com/ydb-platform/ydb-go-sdk/v3：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ydb-platform/ydb-go-sdk/v3@v3.135.0`。基线 v3.135.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-yosssi-ace","Owner":"tools","Path":"github.com/yosssi/ace","Version":"v0.0.5","Kind":"removal"} -->

tools github.com/yosssi/ace：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> github.com/yosssi/ace@v0.0.5`。基线 v0.0.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-zeebo-xxh3","Owner":"tools","Path":"github.com/zeebo/xxh3","Version":"v1.1.0","Kind":"removal"} -->

tools github.com/zeebo/xxh3：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/redis/go-redis/v9@v9.22.0 -> github.com/zeebo/xxh3@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-github-com-ziutek-mymysql","Owner":"tools","Path":"github.com/ziutek/mymysql","Version":"v1.5.4","Kind":"removal"} -->

tools github.com/ziutek/mymysql：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> github.com/ziutek/mymysql@v1.5.4`。基线 v1.5.4，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-go-mongodb-org-mongo-driver-v2","Owner":"tools","Path":"go.mongodb.org/mongo-driver/v2","Version":"v2.5.0","Kind":"removal"} -->

tools go.mongodb.org/mongo-driver/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/paulmach/orb@v0.13.0 -> go.mongodb.org/mongo-driver/v2@v2.5.0`。基线 v2.5.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-go-opentelemetry-io-otel-exporters-otlp-otlptrace","Owner":"tools","Path":"go.opentelemetry.io/otel/exporters/otlp/otlptrace","Version":"v1.19.0","Kind":"removal"} -->

tools go.opentelemetry.io/otel/exporters/otlp/otlptrace：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> go.opentelemetry.io/otel/exporters/otlp/otlptrace@v1.19.0`。基线 v1.19.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-go-opentelemetry-io-proto-otlp","Owner":"tools","Path":"go.opentelemetry.io/proto/otlp","Version":"v1.0.0","Kind":"removal"} -->

tools go.opentelemetry.io/proto/otlp：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/ClickHouse/clickhouse-go/v2@v2.45.0 -> go.opentelemetry.io/proto/otlp@v1.0.0`。基线 v1.0.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-golang-org-x-arch","Owner":"tools","Path":"golang.org/x/arch","Version":"v0.8.0","Kind":"removal"} -->

tools golang.org/x/arch：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `github.com/oapi-codegen/runtime@v1.6.0 -> golang.org/x/arch@v0.8.0`。基线 v0.8.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-gopkg-in-yaml-v1","Owner":"tools","Path":"gopkg.in/yaml.v1","Version":"v1.0.0-20140924161607-9f9df34309c0","Kind":"removal"} -->

tools gopkg.in/yaml.v1：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `howett.net/plist@v1.0.1 -> gopkg.in/yaml.v1@v1.0.0-20140924161607-9f9df34309c0`。基线 v1.0.0-20140924161607-9f9df34309c0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-howett-net-plist","Owner":"tools","Path":"howett.net/plist","Version":"v1.0.1","Kind":"removal"} -->

tools howett.net/plist：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> howett.net/plist@v1.0.1`。基线 v1.0.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-cc-v4","Owner":"tools","Path":"modernc.org/cc/v4","Version":"v4.28.1","Kind":"removal"} -->

tools modernc.org/cc/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/cc/v4@v4.28.1`。基线 v4.28.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-ccgo-v4","Owner":"tools","Path":"modernc.org/ccgo/v4","Version":"v4.33.0","Kind":"removal"} -->

tools modernc.org/ccgo/v4：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/ccgo/v4@v4.33.0`。基线 v4.33.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-fileutil","Owner":"tools","Path":"modernc.org/fileutil","Version":"v1.4.0","Kind":"removal"} -->

tools modernc.org/fileutil：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/fileutil@v1.4.0`。基线 v1.4.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-gc-v2","Owner":"tools","Path":"modernc.org/gc/v2","Version":"v2.6.5","Kind":"removal"} -->

tools modernc.org/gc/v2：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/gc/v2@v2.6.5`。基线 v2.6.5，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-gc-v3","Owner":"tools","Path":"modernc.org/gc/v3","Version":"v3.1.2","Kind":"removal"} -->

tools modernc.org/gc/v3：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/gc/v3@v3.1.2`。基线 v3.1.2，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-goabi0","Owner":"tools","Path":"modernc.org/goabi0","Version":"v0.2.0","Kind":"removal"} -->

tools modernc.org/goabi0：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/goabi0@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-libc","Owner":"tools","Path":"modernc.org/libc","Version":"v1.72.1","Kind":"removal"} -->

tools modernc.org/libc：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> modernc.org/libc@v1.72.1`。基线 v1.72.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-memory","Owner":"tools","Path":"modernc.org/memory","Version":"v1.11.0","Kind":"removal"} -->

tools modernc.org/memory：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> modernc.org/memory@v1.11.0`。基线 v1.11.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-opt","Owner":"tools","Path":"modernc.org/opt","Version":"v0.2.0","Kind":"removal"} -->

tools modernc.org/opt：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/opt@v0.2.0`。基线 v0.2.0，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-sqlite","Owner":"tools","Path":"modernc.org/sqlite","Version":"v1.49.1","Kind":"removal"} -->

tools modernc.org/sqlite：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `talenro.local/platform -> modernc.org/sqlite@v1.49.1`。基线 v1.49.1，非 Required，未从另一侧删除。

<!-- devtools-evidence {"ID":"membership-tools-modernc-org-token","Owner":"tools","Path":"modernc.org/token","Version":"v1.1.0","Kind":"removal"} -->

tools modernc.org/token：本归属入口无可达链，保留在 root；见路径证据同 owner/path 的完整链，末边 `modernc.org/libc@v1.72.1 -> modernc.org/token@v1.1.0`。基线 v1.1.0，非 Required，未从另一侧删除。
