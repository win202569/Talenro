# 根图级版本稳定性调查

日期：2026-09-28。状态：调查完成，恢复仍阻塞；不是合同修订或验收通过。

## 范围与结论

用户批准继续寻找稳定恢复方案，不扩大版本例外。本轮仅检查既有证据、固定工具链源码、官方规则及离线包路径/选中版本；没有再次补写 require、下载新版本或修改生产锁文件。

对 cel.dev/expr，已核对的固定依赖图没有能在正常 tidy 后保留 v0.25.2 的合法来源：原 v0.25.2 来自迁出的 Buf，根侧仅有 grpc v1.80.0 声明 v0.25.1，且根模块没有需要 cel 包的引用路径。此前直接约束已被正常 tidy 删除。本结论针对当前固定依赖及合同，不声称穷尽所有可能架构或上游版本。

## 证据链

1. 原 baseline-graph：Buf v1.72.0 → cel.dev/expr v0.25.2。当前 tools 图仍有该边及显式 v0.25.2。
2. 当前 root 严格图：grpc v1.80.0 → cel.dev/expr v0.25.1；所有 cel 入边检索仅找到这一条。固定 Buf/grpc 的缓存 go.mod 与上述边一致。
3. 本轮固定 Go 1.26.5，根 cwd，GOENV/GOWORK=off、GOTOOLCHAIN=local、GOPROXY/GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly，单命令 90 秒上限：
   - `go mod why -m cel.dev/expr modernc.org/mathutil google.golang.org/grpc`：退出 0，0.179 秒。
   - `go list -m -json cel.dev/expr modernc.org/mathutil google.golang.org/grpc`：退出 0，0.119 秒，无 `-e`，三条结果无 Error。这不是全模块查询。
4. 实际选中 cel v0.25.1、mathutil v1.7.1、grpc v1.80.0。why 表明 cel 不被主模块需要；对照链为 Goose CLI → sqlite → libc → mathutil，以及 Goose CLI → ydb SDK → grpc。没有把声明版本误当选中版本。
5. 前轮单项 require cel v0.25.2 实验在正常根 cwd tidy 后消失，四锁回到实验前哈希，详见[版本证据](2026-09-27-devtools-version-lock-evidence.md)。本轮没有重复此失败实验。

## 整理机制

[Go 官方模块说明](https://go.dev/ref/mod#go-mod-tidy)规定，tidy 根据主模块、工具及递归包/测试导入整理要求，并移除不提供相关包的要求；不能把任意图节点的显式 require 当作永久版本锁。

固定本机 Go 1.26.5 源码 `src/cmd/go/internal/modload/buildlist.go` 的 `tidyPrunedRoots`（803 行起）与此一致：先以实际包所属模块建立根要求，再补足包/测试解析需要及消歧义的要求，不以完整旧模块节点集为保留目标。此源码检查用于解释已观察到的删除，不是新算法实现或第三方代码执行。

## 不可变包基线的对照

以下是历史 baseline.json 的统计，不是本轮实时八组解析结果。

| 组 | 总包数 | cel 包数 | mathutil 包数 |
| --- | ---: | ---: | ---: |
| root | 516 | 0 | 0 |
| integration | 523 | 0 | 0 |
| goose | 656 | 0 | 1 |
| buf | 858 | 1 | 0 |
| protoc | 137 | 0 | 0 |
| oapi | 279 | 0 | 0 |
| sqlc | 575 | 1 | 0 |
| lint | 1187 | 0 | 0 |

独立读取 policy.Required.root 的 28 项并逐项与 baseline 的 root/integration/goose Module 字段比较，28 项均为 0/0/0。这提示可能存在同类图级版本问题，但不能据此断言 28 项都无法恢复、都没有其他构建条件/第三方测试影响，或一次批准全部降级。

## 路径筛选与边界

- 重复补写 cel require：已被实验排除，且违反正常 tidy 稳定性要求。
- 将 Buf 或工具模块重新引入根：能解释原版本来源，但违反模块隔离/禁止跨模块 require，不实施。
- 升级 grpc 或其他上游：改变固定版本及可能实际包版本，超出现有合同，不实施。
- 添加人为包导入、工具锚点，或改 Go 版本、工作区、replace/exclude、第三方缓存：改变既定边界，不作为修复。
- 保持当前边界并逐项评估根图级节点的验收修订：可作为后续设计调查方向，但尚未授权规则变更，更不是安全等价或版本例外批准。

当前未找到满足全部不变项的 cel 稳定恢复路径。建议将问题作为合同冲突审阅：先分类并逐项评估根图级节点，仅讨论证据完整的精确差异；实际运行包集合/版本、正常 tidy、完整性与安全门禁仍不可放宽。不进行逐个失败 pin 的盲试。

## 工作树状态

四锁哈希仍为前轮记录的 A0883514… / 642D51BD… / 0A831254… / 85A30247…，本轮未改策略、测试或产品实现。Task 2、R5–R8 和 Task 3 未完成；无暂存、提交、推送、合并或受阻的普通/C11/C12 测试。原有未提交工作保留。
