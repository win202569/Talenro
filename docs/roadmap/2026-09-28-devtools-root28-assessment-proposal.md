# 根模块 28 项图级依赖评估与修订建议

日期：2026-09-28。状态：待审评估建议，不是获批设计、实现计划或验收记录。

## 目的和授权边界

用户批准逐项评估并形成书面建议。目标是在保留双模块隔离、实际包集合和版本的前提下，解决“必须恢复原完整图节点版本”与正常 tidy 的冲突。本轮不修改 policy、模块锁文件、生产代码或测试，不准备额外依赖、不提交或推送。

本建议不宣称低版本安全等价，不把 Windows 工具入口限制扩展为业务平台限制，不解除 Defender、I2/I3 或完整普通/C11/C12 的既有阻塞。

## 证据强度

- 不可变 baseline.json：28 项在 root/integration/goose 的包模块映射中均为 0/0/0。八组总数已用 cel/mathutil 正对照核实，详见[稳定性调查](2026-09-28-devtools-root-version-stability-investigation.md)。这只是历史包基线。
- 本轮真实离线 `go mod why -m` 明确列出这 28 项，退出 0 / 0.273 秒，28 项均返回主模块不需要该模块。此命令检查包引入路径，不是第三方测试执行、跨平台验证或安全评估。
- 本轮真实离线 `go list -m -json` 明确列出这 28 项，退出 1 / 0.169 秒：模块查找被 GOPROXY=off 阻止，stdout 无有效结果。不以部分结果或 -e 补成“成功”，没有联网补齐。
- 前轮严格 root graph、tools graph 和原 baseline graph 用于逐项比对父边。图中包含历史较低版本的节点；存在一条边不等于父节点是当前选中版本，也不证明该节点提供实际运行包。
- 下表“观察候选”来自先前诊断差异与完整图声明边的交叉核对，不是本轮 28 项严格选中版本查询成功。其中 cel v0.25.1 已在上一轮单项严格查询核实。其他项最终必须取得严格完整查询和单独证据。
- 只有 cel 的根侧显式 require 已做正常 tidy 稳定性实验并失败。不能把该结果当成其他 27 项各自实验的结果。

所有诊断均使用固定 Go 1.26.5、根 cwd、GOENV/GOWORK=off、GOTOOLCHAIN=local、GOPROXY/GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly，单命令上限 90 秒。

## 来源简称

- Buf：github.com/bufbuild/buf v1.72.0；Lint：github.com/golangci/golangci-lint/v2 v2.12.2。
- Goose：github.com/pressly/goose/v3 v3.27.1；CH：github.com/ClickHouse/clickhouse-go/v2 v2.45.0；ch-go：github.com/ClickHouse/ch-go v0.71.0。
- runtime：github.com/oapi-codegen/runtime v1.6.0；grpc：google.golang.org/grpc v1.80.0；common：github.com/prometheus/common v0.66.1；YDB：github.com/ydb-platform/ydb-go-sdk/v3 v3.135.0。
- gosec v2.26.1、golines v0.15.0、viper v1.12.0 均有 Lint 父边；containerregistry v0.21.7 有 Buf 父边；failpoint v0.0.0-20240528011301-b51a646c7c86 有 sqlc/parser 父边。这里列举可核对的来源路径，不声称唯一来源。

## 逐项清单

A 类：当前根图中可直接核对到保留模块版本的声明边，共 22 项。B 类：当前依据含历史版本的传递图展开，共 6 项，需要额外追踪裁剪和聚合模块影响。分类是调查优先级，不是风险等级或自动放行规则。

| # | 根目标 | 原版本 → 观察候选 | 根图来源示例 | 原较高版本来源示例 | 类别及重点 |
| --- | --- | --- | --- | --- | --- |
| 1 | cel.dev/expr | v0.25.2 → v0.25.1 | grpc | Buf | A；唯一已实测根 pin 被 tidy 删除项 |
| 2 | cloud.google.com/go | v0.121.2 → v0.34.0 | 历史 oauth2 bf48bf16ab8d | Lint → gosec | B；聚合模块、传递旧图 |
| 3 | github.com/Azure/go-ansiterm | v0.0.0-20250102033503-faa5f7b0171c → v0.0.0-20210617225240-d185dfc1b5a1 | CH | Buf → moby/client v0.5.0 | A；平台条件 |
| 4 | github.com/BurntSushi/toml | v1.6.0 → v1.3.2 | runtime | Lint、revive、gosec | A；配置相关包/测试 |
| 5 | github.com/alecthomas/units | v0.0.0-20240927000941-0f3dac36c52b → v0.0.0-20211218093645-b94a6e3cc137 | common | Lint → golines | A；测试/辅助包 |
| 6 | github.com/ebitengine/purego | v0.10.0 → v0.8.4 | CH | Lint | A；平台和构建条件 |
| 7 | github.com/gorilla/websocket | v1.5.3 → v1.4.2 | CH | Lint → gosec | A；网络包用途 |
| 8 | github.com/hashicorp/go-version | v1.9.0 → v1.8.0 | ch-go | Lint | A；包/测试用途 |
| 9 | github.com/mattn/go-colorable | v0.1.15 → v0.1.14 | runtime | Buf | A；终端平台条件 |
| 10 | github.com/moby/moby/api | v1.55.0 → v1.54.2 | Goose | Buf、moby/client v0.5.0 | A；容器测试依赖 |
| 11 | github.com/moby/moby/client | v0.5.0 → v0.4.1 | Goose | Buf | A；容器测试依赖 |
| 12 | github.com/moby/term | v0.5.2 → v0.5.0 | CH | Buf → moby/client v0.5.0 | A；终端平台条件 |
| 13 | github.com/pelletier/go-toml/v2 | v2.3.1 → v2.2.2 | runtime | Lint | A；配置相关包/测试 |
| 14 | github.com/power-devops/perfstat | v0.0.0-20240221224432-82ca36839d55 → v0.0.0-20210106213030-5aafc221ea8c | CH | Lint | A；非 Windows 平台条件 |
| 15 | github.com/shirou/gopsutil/v4 | v4.26.4 → v4.25.6 | CH | Lint | A；系统/平台条件 |
| 16 | github.com/sirupsen/logrus | v1.9.4 → v1.9.3 | CH；runtime 声明 v1.9.1 | Buf、Lint、containerregistry | A；多条版本边 |
| 17 | github.com/tklauser/go-sysconf | v0.3.16 → v0.3.12 | CH | Lint | A；系统/平台条件 |
| 18 | github.com/tklauser/numcpus | v0.11.0 → v0.6.1 | CH | Lint | A；系统/平台条件 |
| 19 | go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp | v0.69.0 → v0.68.0 | Goose；CH 声明 v0.49.0 | Buf | A；网络遥测、多条版本边 |
| 20 | go.uber.org/zap | v1.28.0 → v1.27.1 | ch-go；YDB 声明 v1.27.0 | Buf | A；多条版本边 |
| 21 | golang.org/x/lint | v0.0.0-20190930215403-16217165b5de → v0.0.0-20190313153728-d0100b6bd8b3 | 历史 genproto/grpc | sqlc → failpoint；历史 atomic/goleak | B；旧工具图展开 |
| 22 | golang.org/x/oauth2 | v0.36.0 → v0.34.0 | grpc；common/client_golang 声明 v0.30.0 | Buf → containerregistry | A；认证用途、历史边须区分 |
| 23 | golang.org/x/xerrors | v0.0.0-20220517211312-f3a8303e98df → v0.0.0-20200804184101-5ec99f83aff1 | YDB；历史 grpc-gateway | Lint → viper | A；包/测试用途 |
| 24 | google.golang.org/appengine | v1.6.7 → v1.4.0 | 历史 oauth2、x/tools | Lint → viper | B；历史传递图 |
| 25 | google.golang.org/genproto | v0.0.0-20220519153652-3a47de7e79bd → v0.0.0-20200526211855-cb27e3aa2013 | 历史 grpc v1.47.0、envoy 等 | Lint → viper | B；聚合/拆分重复包风险 |
| 26 | google.golang.org/genproto/googleapis/api | v0.0.0-20260715232425-e75dac1f907d → v0.0.0-20260120221211-b8f7ae30c516 | grpc | Buf | A；与聚合 genproto 联合核对 |
| 27 | gopkg.in/yaml.v2 | v2.4.0 → v2.2.3 | 历史 grpc-gateway v1.16.0、plist/testify | Lint、viper、其他工具依赖 | B；旧图多父边 |
| 28 | honnef.co/go/tools | v0.7.0 → v0.0.0-20190523083050-ea95bdfd59fc | 历史 genproto/grpc | Lint | B；旧静态检查工具图 |

B 类示例入口：保留 ydb-go-genproto v0.0.0-20260311095541-ebbf792c1180 声明 grpc v1.47.0，该历史 grpc 声明旧 oauth2 和旧 genproto；这能解释为何完整图含较旧边，不能把 grpc v1.47.0 误报成实际选中版本（根实际为 v1.80.0）。B 类仍需逐节点完整路径审核。

## 三种处理方向

1. **建议：固定精确候选、逐项条件验收。** 原版本优先自然保留；仅在完整证据满足时接受上表指定的 owner=root/path/version 候选，不接受范围、latest、其他版本或目标删除。比反复恢复图节点更符合当前拆分结果，但会改变原合同，必须另行批准正式设计和计划。
2. **保持现行合同，维持阻塞。** 不接受任何新增根例外。实现不能标记完成；在不改变边界的前提下，可继续调查具有具体来源证据的稳定方案，但不重复无变化试验。
3. **重新设计隔离或上游版本。** 将工具来源挂回根或调整实际使用上游可能改变选中版本，但损害既定隔离或实际包版本不变条件，影响面更大，不建议作为本次修订捷径。

## 建议方向 1 的必要条件（尚未生效）

这不是“28 项批量通过”。建议以本表作为候选上限，实施前仍需定义独立字面清单与对应测试：

- 原 28 个 root 目标不得自动移除；在基线版本与逐项精确候选之间条件选择。原 root 3/tools 4 规则保持各自作用域；若全部候选最终采用，例外上限需明确变为 root 31/tools 4，而不是仍声称总共 7 项。
- 每项候选必须绑定唯一证据 ID、owner、path、实际选中版本、校验和、活动父边、较高版本来源和包影响审查。既有两个模块中原有版本正常保留的地方不强行应用例外。
- 缺失元数据先做只读精确缓存清单；准备版本/来源/总预算须单列并另获授权。当前批准仅限评估，不能拿现有四工具或三根候选的下载授权覆盖这 28 项。
- 两模块严格完整模块查询和完整图均成功；任意查询错误阻塞。不能从本轮失败查询、历史 -e 结果或实际快照自动生成“通过”的期望。
- 实时复验八组包映射、52 个保护文件哈希及正常 tidy 后四锁稳定性。root/integration/goose 任何一组开始使用这 28 项的包，应拒绝图级例外，回到评审；同一路径在 tools 中的合法使用不能误拒绝。
- 逐项检查 TestImports/XTestImports、构建条件及涉及平台；静态检查不等同运行第三方测试，未验证范围要明示。cloud/genproto 和 API/RPC 拆分的重复包风险单列，不靠 replace/exclude 隐藏。
- 不更改完整性验证、专项预算、真实工具验证、独立审查或原安全门禁；没有完成这些条件就不原子提交 Task 2、不合并 main。

## 结论与下一阶段

本轮完成 28 项初步来源分类及建议，未完成候选安全/兼容性验收，未证明其他 27 项的所有稳定恢复路径均不可能。建议先审阅方向 1；同意方向仅允许据此撰写正式设计，不等于修改策略或准备下载。

按 brainstorming 的架构路径，本文件是调查及方案比较稿，不是已批准 spec；下一步为设计审阅。按 systematic-debugging，不继续以重复 pin 掩盖已识别的合同冲突。本轮无提交/推送；本机未提交报告仍不能从 GitHub 恢复。
