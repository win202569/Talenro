# 开发工具隔离修订：根模块 28 项图级版本的精确条件例外

日期：2026-09-28。
状态：用户已明确确认本文书面设计，并随后同时批准对应实施计划修订及 R3c-2 的 27 个精确版本准备范围；批准不等于验收通过。实际执行证据见版本锁证据文档。

## 1. 意图、适用范围与优先级

继续双模块隔离，保持业务、Goose 和五工具实际使用的包集合及版本不变；解决根图级节点原版本约束与正常 tidy 的冲突。选择“固定精确候选、逐项补证验收”，不把完整图比较退化为仅比较运行包，也不批量接受低版本。

本文书面批准并落实到获批实施计划后，仅替代以下冲突条款：

- [2026-09-27 锁版本修订](2026-09-27-c12-devtools-version-lock-revision-design.md)第 4/7 节的 root 28 项无条件精确恢复要求；
- [2026-09-28 四项 tools 修订](2026-09-28-c12-devtools-four-graph-exceptions-design.md)中“root 28 必须恢复、root 最多三个例外”的限制及相应回归描述。

除此以外，原设计、Windows-only 修订、证据绑定、两模块完整性和验收合同继续有效。历史设计与失败证据不改写。未批准本文及后续实施计划前，现行策略仍有效。

模块身份仍为 talenro.local/platform 与 talenro.local/devtools；Go 1.26.0 / go1.26.5、Goose 3.27.1、Buf 1.72.0、protoc-gen-go 1.36.11、oapi-codegen 2.8.0、sqlc 1.31.1、golangci-lint 2.12.2 不变。继续既有 Native inline、原 worktree、Task 2，不重做 Task 1 或重新选择执行方式。

## 2. 已知证据及不能推断的结论

依据[逐项评估](../../roadmap/2026-09-28-devtools-root28-assessment-proposal.md)和[稳定性调查](../../roadmap/2026-09-28-devtools-root-version-stability-investigation.md)：28 项在历史 root/integration/goose 包映射中均无包，本轮离线 why 也未发现包路径；原较高版本均能在迁出工具相关图中找到来源。

其中 22 项有保留模块版本的直接声明边，6 项涉及历史传递图。此分类不是风险等级；历史版本节点仍可能出现在完整图中，不能据父边误报当前选中版本。28 项严格版本元数据查询因缺缓存失败，候选未验收。只有 cel 的根 require 已实测被正常 tidy 删除，不声称其他 27 项已逐项做过相同实验或所有恢复方案均不可能。

不凭 Windows 包缺席、why 无路径或缓存存在声称低版本安全等价。第三方测试、其他平台、源码差异和漏洞状态未核实时明确保留不确定性。

## 3. 唯一新增候选集合

下表所有 owner 均为 root。原值是不可变基线和 Required 的首选；候选只是附条件允许值，不是新的基线。路径不得缩写、使用版本范围或换成其他版本。

| 路径 | 原值 | 唯一候选 |
| --- | --- | --- |
| cel.dev/expr | v0.25.2 | v0.25.1 |
| cloud.google.com/go | v0.121.2 | v0.34.0 |
| github.com/Azure/go-ansiterm | v0.0.0-20250102033503-faa5f7b0171c | v0.0.0-20210617225240-d185dfc1b5a1 |
| github.com/BurntSushi/toml | v1.6.0 | v1.3.2 |
| github.com/alecthomas/units | v0.0.0-20240927000941-0f3dac36c52b | v0.0.0-20211218093645-b94a6e3cc137 |
| github.com/ebitengine/purego | v0.10.0 | v0.8.4 |
| github.com/gorilla/websocket | v1.5.3 | v1.4.2 |
| github.com/hashicorp/go-version | v1.9.0 | v1.8.0 |
| github.com/mattn/go-colorable | v0.1.15 | v0.1.14 |
| github.com/moby/moby/api | v1.55.0 | v1.54.2 |
| github.com/moby/moby/client | v0.5.0 | v0.4.1 |
| github.com/moby/term | v0.5.2 | v0.5.0 |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.2.2 |
| github.com/power-devops/perfstat | v0.0.0-20240221224432-82ca36839d55 | v0.0.0-20210106213030-5aafc221ea8c |
| github.com/shirou/gopsutil/v4 | v4.26.4 | v4.25.6 |
| github.com/sirupsen/logrus | v1.9.4 | v1.9.3 |
| github.com/tklauser/go-sysconf | v0.3.16 | v0.3.12 |
| github.com/tklauser/numcpus | v0.11.0 | v0.6.1 |
| go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp | v0.69.0 | v0.68.0 |
| go.uber.org/zap | v1.28.0 | v1.27.1 |
| golang.org/x/lint | v0.0.0-20190930215403-16217165b5de | v0.0.0-20190313153728-d0100b6bd8b3 |
| golang.org/x/oauth2 | v0.36.0 | v0.34.0 |
| golang.org/x/xerrors | v0.0.0-20220517211312-f3a8303e98df | v0.0.0-20200804184101-5ec99f83aff1 |
| google.golang.org/appengine | v1.6.7 | v1.4.0 |
| google.golang.org/genproto | v0.0.0-20220519153652-3a47de7e79bd | v0.0.0-20200526211855-cb27e3aa2013 |
| google.golang.org/genproto/googleapis/api | v0.0.0-20260715232425-e75dac1f907d | v0.0.0-20260120221211-b8f7ae30c516 |
| gopkg.in/yaml.v2 | v2.4.0 | v2.2.3 |
| honnef.co/go/tools | v0.7.0 | v0.0.0-20190523083050-ea95bdfd59fc |

原 root 三候选仍仅为 readline v1.5.1、demangle v0.0.0-20250417193237-f615e6bd150b、otel/metric/x v0.66.0，沿用原完整路径和必要性合同；tools 四候选仍仅为 x/time v0.11.0、mathutil v1.6.0、sortutil/strutil v1.2.0，沿用四项修订。总候选上限明确为 root 31 / tools 4，最多 35 个 owner/path/version 三元组，实际采用可少于此数。

原 Required 仍有 32 项（root 28/tools 4），原优先版本不改；这些目标必须存在且归属为 retained。目标消失不能标记 removed 规避验收，须重新审阅。原值自然稳定时不强行改候选；本表例外不得用于 tools 的同名节点。其他保留节点除既有精确例外外，仍严格等于原基线。

## 4. 逐项证据合同

每个实际采用的新增 root 候选须同时满足：

1. 两模块严格离线完整模块查询和完整图成功，无 Error、重复身份或缺失结果；禁止 -e 充当成功证据。记录真实选中版本，不从图中任一边猜测。B 类节点从保留入口追踪完整历史展开链，区分声明版本与选中版本。
2. 记录原值、候选、原高版本来源、当前活动父边、源码用途及保持隔离时选择该候选的理由。核对上游清单、包/测试导入与构建条件。允许以具体来源追踪说明约束为何不稳定，不要求为凑证据把 27 项全部重复 pin；已做实验必须如实保留结果，未做不得伪称失败。
3. 新鲜解析的 root/integration/goose 三组均不含本表任一路径的包，八组全部 ImportPath/Module.Path/Module.Version 与原基线相同。无论原值还是候选，一旦本表节点进入根侧这三组实际包，图级边界检查均失败，回到评审。tools 对同名节点的原有合法包使用不误拒绝；tools 四项原用途禁入规则不变。
4. 逐项检查相关 TestImports/XTestImports 和其他构建条件的潜在影响；静态审查不代替运行第三方测试。非 Windows 业务平台不因工具 Windows-only 自动排除。发现实际业务用途变化、无法解释的测试/平台影响或重复包，停止该候选验收；未运行范围在报告中明确列出，不宣称通过。
5. 来源校验和、必要性及用途审查正文保存到可提交证据文档。采用原独立行 `<!-- devtools-evidence {JSON} -->` 记录格式，使用 ID/Owner/Path/Version/Kind/Sum/GoModSum；Kind=exception。ID 唯一存在，绑定实际 root/path/version，两个校验和非空且与可信来源记录一致。机器绑定通过不代替正文审查。
6. 正常 tidy 在所属模块目录运行后，四锁字节及模块选中结果稳定；没有根工具专属依赖回流、新跨模块依赖或未解释差异。全部成员归属逐项审核，移出非 Required 节点仍需唯一 removal 证据绑定原版本；不能复制实际快照作为期望。

cloud/google genproto 聚合节点和独立 api/rpc 模块单列联合核对：不靠直接钉旧聚合模块、replace/exclude、忽略错误来掩盖重复包或裁剪变化。证据未闭合即保留未验收，不将 B 类自动降为较低标准。

## 5. 数据与检查器边界

保留 schema 1 的不可变 666 条模块、八组包映射、54 条原保护哈希归档；其中 52 个非根锁文件必须保持不变，原根 go.mod/go.sum 哈希作为历史基线，不要求拆分后的锁文件等于迁移前内容。本轮及候选验证的四锁稳定性单独比较。

实现只在原版本策略/测试层扩展精确候选和根侧用途检查，不更换基线，不扩展产品接口。Required 保留原 32 个字面三元组；Exceptions 用独立的原 7 加本文 28 字面三元组作期望，不由现场查询生成。Membership、Evidence 保持独立审核，版本符合不自动使其通过。

复用既有快照、证据解析/绑定和包用途检查分工。完整 Lock 必须联合执行严格模块读取、版本/归属比较、证据读取绑定、八组实时包比较和用途拒绝；任何一层失败不得被其他绿色结果抵消。JSON 字段、重复记录、未知字段和额外内容按原严格合同拒绝。

## 6. 准备权限与实施顺序

本文批准仅允许修订原实施计划，不立即修改策略或模块，不立即下载。计划先列只读精确缓存核查，再列缺失的具体 module@version、准备文件类型、来源、输出、总预算及批准范围。准备清单不得从已批准的旧 207 项、root 三项或 tools 四项授权推定。

准备范围需明确授权；可在审阅实施计划时一并明确批准。所需额外版本或传递元数据不在获批清单时停止报告，不使用 latest/direct/VCS/auth 回退。固定 Go、独立最小备用模块，仅官方 proxy.golang.org 和 sum.golang.org，不关闭校验、不执行下载代码、不修改生产锁文件。沿用原整批 900 秒准备上限，不因项数增加偷偷分批重置预算。

正式验证继续禁网，缺包或篡改必须失败。先完成新规则 RED→GREEN 及严格读取/反例覆盖，再在准备授权满足后填充真实证据和归属、复验完整图/八组包/稳定性，最后继续原脚本与真实工具验收。计划须接续现有工作，保留已完成部分，不重跑无变化失败试验。

## 7. 回归与完成标准

至少覆盖 28 个精确候选逐项正例、原值正例，以及每项错误版本、用于 tools、目标缺失/removed、缺失或错误绑定证据的反例；额外第 29 个新增 root 路径或第 5 个 tools 路径必须失败。原 root 三项、tools 四项及其用途边界回归保留。

根侧三个阶段逐一覆盖候选提供包的拒绝，即使版本符合也不能通过；在 tools 中本表同名合法包是正对照，root/Goose mathutil v1.7.1 是原规则正对照。八组缺组/空组/重复/Incomplete/Error/DepsErrors 均失败。保留历史 35 差异回放：区分条件齐备的精确候选与仅值相同但证据不足，不整体删除旧反例。

必须完成两模块完整性校验、正常重复 tidy 稳定性、严格完整图、八组包映射、52 个保护文件、适用脚本专项、根专属 lint fixture、真实工具版本、生成无差异、真实根 lint、最终独立审查。沿用原元数据单命令 90 秒、专项 5m/组合 10m/Linux 3m、工具验证 900 秒及 C11 嵌套预算；若增加覆盖导致超时，报告阻塞，不提高上限或缩减检查。

不新增 go.work、跨模块 require、replace/exclude、虚构 import/tool；不修改第三方缓存或清单，不改变生产逻辑、SQL、C12 runner、smoke、生成内容或 Docker 资源协议。不通过放弃正常 tidy、排除真实依赖或重写包基线达成绿色。

## 8. 审批、失败与交付

任何未知版本/新增、目标删除、实际包变化、用途/来源无法解释、查询错误或证据不完整均保持阻塞；未知新候选需要单独设计决定，不自动扩展到 36 项或其他归属。回退仅针对本次明确改动，保留历史 WIP、基线及失败证据，不进行广泛重置。

模块、消费者和实现测试仍按原 Task 2 原子提交，未满足合同不完成 Task 2，不开始冒称 Task 3 成功。设计保存不等于实现可换机恢复；本轮仅保存未提交设计，不暂存已有 WIP、不推送或合并。

Defender 事件、完整普通/C11/C12/bootstrap 与 I2/I3 的阻塞保持。不得关闭防护、添加排除或恢复隔离文件；本文不批准安全事件恢复、main 合并或推送。

下一步：使用 writing-plans 修订既有实施计划；用户再审阅计划及其中独立列出的精确准备权限，之后接续既有 Native inline。本文书面确认不追认为对尚未审阅的计划、准备清单或实现的批准。
