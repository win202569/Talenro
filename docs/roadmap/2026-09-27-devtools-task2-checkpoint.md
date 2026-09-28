# 开发工具拆分：任务 2 未验收检查点

日期：2026-09-27。工作分支：`codex/c12-b01-task9-coordinator`。
状态：实现正在工作树中；未原子提交、未推送、未合并，任务 2 未完成。

## 已取得的证据

- 经授权准备原锁定依赖：207 项下载全部返回校验和，退出 0 / 338.972 秒；迁移前根锁文件未变。
- 迁移前完整模块元数据 666 条，普通包 516、integration 523、Goose 656；五工具包分别为 Buf 858、protoc 137、oapi 279、sqlc 575、lint 1187。所有接受的输出无 Error/Incomplete/DepsErrors。
- 默认 Goose JSON 查询及五工具合并查询各在原 90 秒诊断上限失败，保留原失败事实。后续只读取完整依赖字段、不计算 Stale，并拆分工具查询；不能据此宣称冷态性能改善。
- Windows 新路由专项先见 RED，消费者实现后退出 0 / 242.636 秒；验证前停止、固定宿主、模块 argv/cwd、嵌套插件失败、强杀入口后的自有子进程退出和 foreign canary 均通过。
- Linux 四入口拒绝测试先见 RED（1.268 秒），后退出 0 / 0.960 秒；50 个叶场景通过（独立验证器 10，加四入口 40）。这不是 Unix 正向工具支持。
- PowerShell 隐私契约保留脱敏、环境隔离、阶段顺序和退出码 73；测试批处理的带空格参数引用修正后退出 0 / 34.618 秒。生产隐私断言未放宽。
- 根普通/integration/Goose 包闭包重新查询成功，集合与实际选中模块版本均零变化（516/523/656）。五工具同样逐包比较，858/137/279/575/1187 条全部零新增、零删除、零版本变化、零错误；不以模块图诊断输出替代成功包查询。
- 完整计划脚本专项选择器退出 0 / 577.969 秒，原 10 分钟上限不变：256 个叶场景 PASS、0 FAIL、1 不适用 Skip（Windows 不适用的 launcher-cancel）。全部 20 个顶层测试实际运行并通过，包含三个迁移后的 Bash 拒绝测试、Bash smoke、隐私及清理契约，未通过漏选制造绿色。专项通过不代表工具完整性验证、真实生成、实际 lint、完整普通/C11/C12 或最终独立审查通过。

## 当前阻塞：完整模块图版本差异

离线 tidy 本身在根和工具模块均退出 0，但不代表完整选中版本保持不变。
严格的根 `go list -m -json all` 失败：24 个图节点的当前版本元数据未缓存。
仅为调查运行带 `-e` 的诊断：根 308 条中 31 处版本差异，工具 508 条中 4 处差异、4 个元数据错误。
这些部分结果绝不算验收成功，未下载这些新选中版本，也未批准其变化。

根 pprof 版本仍为 `v0.0.0-20260115054156-294ebfa9ad83`，tidy 后成为显式间接依赖。
该已锁定模块自己的 go.mod 要求 readline v1.5.1 和 demangle v0.0.0-20250417193237-f615e6bd150b，
而旧根图在裁剪状态下选中下面列出的更旧版本。因此不能把所有差异描述成简单降级，
也不能承诺加低版本 require 就能还原它们。尚需确定可保持原版本的约束方案；若无法保持，
按批准设计第 3 节单独报告和审批，不能自行添加 replace/exclude 或接受变化。

| 模块 | 依赖 | 迁移前选中版本 | 当前诊断选中版本 |
| --- | --- | --- | --- |
| tools | `golang.org/x/time` | `v0.14.0` | `v0.11.0` |
| tools | `modernc.org/mathutil` | `v1.7.1` | `v1.6.0` |
| tools | `modernc.org/sortutil` | `v1.2.1` | `v1.2.0` |
| tools | `modernc.org/strutil` | `v1.2.1` | `v1.2.0` |
| root | `cel.dev/expr` | `v0.25.2` | `v0.25.1` |
| root | `cloud.google.com/go` | `v0.121.2` | `v0.34.0` |
| root | `github.com/Azure/go-ansiterm` | `v0.0.0-20250102033503-faa5f7b0171c` | `v0.0.0-20210617225240-d185dfc1b5a1` |
| root | `github.com/BurntSushi/toml` | `v1.6.0` | `v1.3.2` |
| root | `github.com/alecthomas/units` | `v0.0.0-20240927000941-0f3dac36c52b` | `v0.0.0-20211218093645-b94a6e3cc137` |
| root | `github.com/chzyer/readline` | `v0.0.0-20180603132655-2972be24d48e` | `v1.5.1` |
| root | `github.com/ebitengine/purego` | `v0.10.0` | `v0.8.4` |
| root | `github.com/gorilla/websocket` | `v1.5.3` | `v1.4.2` |
| root | `github.com/hashicorp/go-version` | `v1.9.0` | `v1.8.0` |
| root | `github.com/ianlancetaylor/demangle` | `v0.0.0-20200824232613-28f6c0f3b639` | `v0.0.0-20250417193237-f615e6bd150b` |
| root | `github.com/mattn/go-colorable` | `v0.1.15` | `v0.1.14` |
| root | `github.com/moby/moby/api` | `v1.55.0` | `v1.54.2` |
| root | `github.com/moby/moby/client` | `v0.5.0` | `v0.4.1` |
| root | `github.com/moby/term` | `v0.5.2` | `v0.5.0` |
| root | `github.com/pelletier/go-toml/v2` | `v2.3.1` | `v2.2.2` |
| root | `github.com/power-devops/perfstat` | `v0.0.0-20240221224432-82ca36839d55` | `v0.0.0-20210106213030-5aafc221ea8c` |
| root | `github.com/shirou/gopsutil/v4` | `v4.26.4` | `v4.25.6` |
| root | `github.com/sirupsen/logrus` | `v1.9.4` | `v1.9.3` |
| root | `github.com/tklauser/go-sysconf` | `v0.3.16` | `v0.3.12` |
| root | `github.com/tklauser/numcpus` | `v0.11.0` | `v0.6.1` |
| root | `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | `v0.69.0` | `v0.68.0` |
| root | `go.opentelemetry.io/otel/metric/x` | `（基线无此模块）` | `v0.66.0` |
| root | `go.uber.org/zap` | `v1.28.0` | `v1.27.1` |
| root | `golang.org/x/lint` | `v0.0.0-20190930215403-16217165b5de` | `v0.0.0-20190313153728-d0100b6bd8b3` |
| root | `golang.org/x/oauth2` | `v0.36.0` | `v0.34.0` |
| root | `golang.org/x/xerrors` | `v0.0.0-20220517211312-f3a8303e98df` | `v0.0.0-20200804184101-5ec99f83aff1` |
| root | `google.golang.org/appengine` | `v1.6.7` | `v1.4.0` |
| root | `google.golang.org/genproto` | `v0.0.0-20220519153652-3a47de7e79bd` | `v0.0.0-20200526211855-cb27e3aa2013` |
| root | `google.golang.org/genproto/googleapis/api` | `v0.0.0-20260715232425-e75dac1f907d` | `v0.0.0-20260120221211-b8f7ae30c516` |
| root | `gopkg.in/yaml.v2` | `v2.4.0` | `v2.2.3` |
| root | `honnef.co/go/tools` | `v0.7.0` | `v0.0.0-20190523083050-ea95bdfd59fc` |

## 实施中的裁定与保留边界

1. 网络准备使用私有备用模块，不改原基线锁文件；准备成功仍须另做离线验证。
2. 依赖元数据查询不计算 Stale，保留完整包/模块/导入边和错误字段；逐工具查询保持每条原 90 秒上限。代价是不能提供构建陈旧状态或冷态性能证据。
3. 将全图 665 个节点直接设为新模块约束会激活旧 genproto 并造成重复包；改用 312 个实际工具包来源约束。随后补回原 mock v0.6.0、procfs v0.20.1，并准备三个原基线已有测试源码（gomega/expect/packagestest）。代价是其他图节点的选中版本必须继续逐项核对，当前差异仍是阻塞。

52 个受保护基线文件哈希未变化；源码、SQL、C12 runner、smoke 文件和生成目录未修改。doubleclick 保留在工具模块，
其来源仍是原 sqlc v1.31.1，不能删除缓存或第三方 testdata 规避验证。
Defender 历史事件及 I2/I3 未解决；完整普通/C11/C12 执行授权仍未恢复。
在版本差异解决前，不做任务 2 原子提交、不开始真实工具验收、不推送或合并。

接续时仍需补足计划中的根模块专属源码 fixture（当前 lint 专项检查实际 argv/cwd/环境，真实 lint 仍未运行），
并在锁版本决策后完成模块图无错误查询、原子提交和 task-done；本次绿色结果不能替代这些条件。
