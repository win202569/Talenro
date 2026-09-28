# 开发工具隔离锁版本修订：精确恢复与最小例外

日期：2026-09-27。
状态：用户已明确确认本文书面设计及对应修订计划，并批准计划 R3 的精确准备范围。不是实施完成或验收通过记录。

## 1. 目的、适用范围与优先级

目标是继续完成已批准的双模块隔离，优先恢复迁移前依赖版本，只对明确、必要且验证合格的差异保留最小例外。
本修订只细化以下原设计的锁版本条款：

- `2026-09-25-c12-development-tool-isolation-design.md` 第 3 节及相关版本验收条款；
- `2026-09-25-c12-devtools-windows-powershell-addendum-design.md`；
- `2026-09-27-c12-devtools-windows-only-design.md` 的双模块版本约束。

本文书面获批后，仅在版本差异处理上优先；其他要求原样保留。原文和历史失败记录不改写。
模块身份仍为根 `talenro.local/platform` 与 `talenro.local/devtools`，不新建模块或工作区。
用户已选择 Native inline；后续修订原实施计划，不重复选择执行方式。

## 2. 不变项

Go 版本 1.26.0、工具链 go1.26.5，Goose 库与 CLI 3.27.1 留在根模块且驱动集合不变。
开发工具仍为 Buf 1.72.0、protoc-gen-go 1.36.11、oapi-codegen 2.8.0、sqlc 1.31.1、golangci-lint 2.12.2。
业务实际依赖、运行包集合及其选中版本不允许借本修订升级或降级。

不建立 go.work、不添加跨模块 require、不新增任何 replace/exclude、不修改第三方模块清单或篡改已有缓存内容。按第 6 节另获授权的精确版本准备不属于缓存篡改。
工具调用仍使用显式 -modfile，工作目录、Buf 嵌套插件、Goose 分派及 Windows 独立宿主/Job 归属保持原合同。
两个模块完整性验证、离线验收、脱敏输出和既有预算不变；不新增跳过验证标志。
不修改生产逻辑、SQL、C12 runner、smoke、生成内容或 Docker 资源协议。
不关闭防护、添加排除项、恢复隔离文件或把 Defender 事件视为误报。

## 3. 基线与已知证据的边界

以迁移前完整 666 条模块记录及八组已成功解析的包闭包为不可变基线；版本目标在本设计明确列出。
原证据位于本计划忽略目录，实施阶段必须将可复核的版本清单、差异和摘要保存为适合提交的非敏感证据，不能只依赖本机缓存或忽略文件。
原 35 处差异按“模块归属 + 路径”计数：根 28 处降级、2 处升级、1 个新增，工具模块 4 处降级。

已记录的 Windows 八组包闭包集合和实际模块版本相同、专项 256 PASS / 0 FAIL / 1 不适用 Skip、Linux 拒绝 50 PASS，均为先前证据。
它们不能证明完整图无差异、第三方测试依赖无变化或其他平台不受影响，也不能替代真实工具和完整性验收。
带 -e 的模块输出仅用于诊断；完整严格查询未成功前，不能把缺失节点当作已删除或已接受。

## 4. 32 项精确恢复目标

以下各项目标均为原选中版本。它们没有降级例外；“恢复目标”不表示当前已有可行或验证通过的约束方案。
先追踪保留依赖的真实引入链和上游约束，再选择最小的精确约束。不能按名字批量删依赖或为凑齐数字重新挂回工具专属依赖。

| 模块 | 依赖 | 必须恢复的选中版本 |
| --- | --- | --- |
| tools | `golang.org/x/time` | `v0.14.0` |
| tools | `modernc.org/mathutil` | `v1.7.1` |
| tools | `modernc.org/sortutil` | `v1.2.1` |
| tools | `modernc.org/strutil` | `v1.2.1` |
| root | `cel.dev/expr` | `v0.25.2` |
| root | `cloud.google.com/go` | `v0.121.2` |
| root | `github.com/Azure/go-ansiterm` | `v0.0.0-20250102033503-faa5f7b0171c` |
| root | `github.com/BurntSushi/toml` | `v1.6.0` |
| root | `github.com/alecthomas/units` | `v0.0.0-20240927000941-0f3dac36c52b` |
| root | `github.com/ebitengine/purego` | `v0.10.0` |
| root | `github.com/gorilla/websocket` | `v1.5.3` |
| root | `github.com/hashicorp/go-version` | `v1.9.0` |
| root | `github.com/mattn/go-colorable` | `v0.1.15` |
| root | `github.com/moby/moby/api` | `v1.55.0` |
| root | `github.com/moby/moby/client` | `v0.5.0` |
| root | `github.com/moby/term` | `v0.5.2` |
| root | `github.com/pelletier/go-toml/v2` | `v2.3.1` |
| root | `github.com/power-devops/perfstat` | `v0.0.0-20240221224432-82ca36839d55` |
| root | `github.com/shirou/gopsutil/v4` | `v4.26.4` |
| root | `github.com/sirupsen/logrus` | `v1.9.4` |
| root | `github.com/tklauser/go-sysconf` | `v0.3.16` |
| root | `github.com/tklauser/numcpus` | `v0.11.0` |
| root | `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | `v0.69.0` |
| root | `go.uber.org/zap` | `v1.28.0` |
| root | `golang.org/x/lint` | `v0.0.0-20190930215403-16217165b5de` |
| root | `golang.org/x/oauth2` | `v0.36.0` |
| root | `golang.org/x/xerrors` | `v0.0.0-20220517211312-f3a8303e98df` |
| root | `google.golang.org/appengine` | `v1.6.7` |
| root | `google.golang.org/genproto` | `v0.0.0-20220519153652-3a47de7e79bd` |
| root | `google.golang.org/genproto/googleapis/api` | `v0.0.0-20260715232425-e75dac1f907d` |
| root | `gopkg.in/yaml.v2` | `v2.4.0` |
| root | `honnef.co/go/tools` | `v0.7.0` |

恢复 cloud.google.com/go 和旧 google.golang.org/genproto 时，优先约束已有上游依赖，使原版本间接保持。
不得把旧聚合 genproto 无差别加入直接 require，从而与独立 api/rpc 模块形成重复包。
新约束必须说明它服务于哪条保留依赖链，证明没有把五工具专属闭包带回根模块。
若再次 tidy 删除某个图级约束，必须解释稳定性方案；不得通过反复补写约束、跳过 tidy 或隐藏差异制造成功。
若没有既稳定又符合模块边界的方案，该项保持阻塞，提交事实与可选修订，不自动扩大例外。

## 5. 三个精确候选例外

书面批准本文的含义：允许在修订实施计划获批并满足第 6–7 节后，将下表精确差异作为锁版本规则的条件性例外。
这不是确认当前实现正确、依赖安全或验收通过；不批准其余差异，也不允许这些例外迁移到工具模块。

| 模块 | 依赖 | 迁移前 | 唯一允许的候选结果 |
| --- | --- | --- | --- |
| root | `github.com/chzyer/readline` | `v0.0.0-20180603132655-2972be24d48e` | `v1.5.1` |
| root | `github.com/ianlancetaylor/demangle` | `v0.0.0-20200824232613-28f6c0f3b639` | `v0.0.0-20250417193237-f615e6bd150b` |
| root | `go.opentelemetry.io/otel/metric/x` | 基线无此模块 | `v0.66.0` |

候选来源：Goose 闭包中 modernc.org/sqlite 的 TestImports 引入 pprof/profile；已有 pprof `v0.0.0-20260115054156-294ebfa9ad83` 声明上表 readline/demangle 版本。
grpc/balancer/pickfirst 的 XTestImports 引入 sdk/metric；已有 sdk/metric `v1.44.0` 声明 metric/x `v0.66.0`。
这些清单和导入记录说明来源，不证明差异不可避免或测试已经执行。不同根约束和裁剪状态可使同一清单不展开同样节点。

三个候选必须逐项记录：来源链、为何不能在不改变保留语义的前提下维持旧结果、精确版本、来源校验、包影响与验证结果。
如果恢复上游后某候选自然消失，则保留原结果或原来的缺席，不人为引入例外。
不允许 latest、版本范围、自动升级或额外模块差异。出现第 4 个候选、不同版本、不同模块归属或实际包变化，均需重新审批。

## 6. 准备、验证与执行权限分开

本次书面设计批准只允许修订实施计划，不立即执行模块修改、下载、第三方工具或产品测试。
修订计划应明确准备与验收两阶段、精确版本清单、输出位置、时限及允许的执行范围；计划获批后才实施。

既有原锁版本补齐授权不自动覆盖新增或新选中的候选版本。若其缓存缺失，按三个精确候选中实际缺失的项目单列准备范围，取得明确准备授权；如计划确认时明确批准该范围，则无需重复确认同一范围。
准备只从批准的官方代理及校验和服务获取精确版本，不使用 direct/VCS/auth 回退、不关闭校验、不执行所下载内容。
正式验证必须禁网，缺失/篡改/查询错误都失败；准备成功不等于完整验证通过。

不运行被原安全事件阻塞的完整普通测试、完整 C11 或 C12/bootstrap，除非另行处理事件并取得明确恢复授权。
为了解析候选影响，可在批准计划范围内读取包与测试依赖元数据；这不授权执行第三方测试，也不得把静态影响检查表述为动态验证。

## 7. 验收合同

1. 两模块严格离线完整模块查询成功，无 Error；完整图及所需包闭包可解析，无 Incomplete/DepsErrors。不接受 -e 的部分成功作为验收。
2. 32 项恢复到第 4 节的精确版本；其余所有保留模块选中版本与基线一致。根模块最多允许第 5 节中实际证明必要且验证合格的精确差异，工具模块零例外；零未解释新增、丢失或版本变化。
3. 模块成员因隔离而移出某一模块，可以发生，但每项删除/迁移须有依赖归属证据；不得让本应保留的依赖消失来规避比较。第 4 节目标若发现不应保留，先重新报批，不擅自将“恢复”改为“删除”。
4. 原八组 Windows 包集合、实际使用版本保持一致；针对候选单独展开和审查 TestImports/XTestImports、声明边及其他构建条件的潜在影响。未验证的平台明确标为未验证，不因四个工具入口 Windows-only 推断业务部署平台也已缩减。
5. 最终锁文件再次整理后稳定，重新读取的模块选择不漂移；每个额外精确约束有保留来源。根模块不能重新引入五工具专属闭包，工具模块保留 sqlc/doubleclick 等真实依赖，不删除缓存来减少校验范围。
6. 保留模块完整性检查、脚本专项及失败路径验证；补齐根模块专属 lint fixture。真实工具版本、根代码 lint 和完整生成无差异均取得独立证据，遵守原时限，不借专项替代实际工具验证。
7. 模块、消费者和测试仍按原计划原子提交；最终独立审查保留。被安全事件阻塞的门槛和 I2/I3 未解决前，项目不标记整体验收通过，不具备 main 合并资格。

## 8. 失败、回退及交付

任何恢复激活重复包、扩大根工具依赖、产生新差异或改变实际包行为时，停止相应实施，保留诊断并报告，不修改其他约束来掩盖失败。
原 90 秒元数据诊断上限、900 秒工具验证上限、C11 嵌套阶段预算以及已批准测试预算不因本修订扩大；不做无变化的重复重试。
回退仅针对本修订产生的明确改动，保留原工作树和基线；不使用广泛重置、删除缓存或覆盖无关文件。

待本文书面确认后，修订原实施计划并再次请用户确认，然后从任务 2 未完成处继续 Native inline。
不重做任务 1 已完成内容，不新建 worktree，不把本次批准当作推送、合并、完整安全验收恢复或更改保护设置的授权。
本轮只保存待审设计；原未提交实现仍是未验收检查点。
