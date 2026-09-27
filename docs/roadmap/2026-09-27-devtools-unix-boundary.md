# Unix 开发工具验证器：首轮真实测试

日期：2026-09-27。状态：部分测试与一项修复完成，任务 1 未完成，禁止视为验收通过。

## 环境及测试范围

用户批准准备官方 Go 1.26.5 Linux 镜像；固定摘要为
`sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd`。
实际为 linux/amd64、Go 1.26.5、Bash 5.2.15。测试容器禁网、只读根文件系统、
仓库只读挂载，关闭额外 capabilities，使用 no-new-privileges；临时缓存与夹具只在
容器 /tmp，配置 --init 和 --rm。不挂载 Docker socket，不读取用户凭据。

本轮在 internal/e2e 目录按文件选择四份开发工具测试，隔离于其他 e2e 夹具：

```text
go test -v -tags=e2e devtools_verification_test.go devtools_runtime_test.go devtools_ownership_test.go devtools_unix_test.go -run '^TestDevtoolsVerificationBash$' -count=1 -timeout=3m
```

这是开发工具专项文件集，不是整个 e2e 包或项目全套测试。Linux 进程观察使用 /proc；
其他 Unix 系统尚未实测，不把 Linux 结果推广为所有 Unix 通过。

## 夹具修正与有效 RED

原完整性夹具在非 Windows 上直接 Skip，现改为真实本地 Go + 独占 file 模块代理和缓存。
缺失、源码篡改、zip 篡改、ziphash 缺失、混合大小写路径、继承模块覆盖项与齐备检查变异
均不接触共享模块缓存。进程用例单独使用自有 Go 可执行夹具，不冒充真实完整性验证。

第一次执行遇到临时模块目录只读导致清理失败，以及夹具编译尝试写只读 /root 缓存。
这不是产品 RED。仅将临时模块准备设为 -modcacherw，并给夹具编译独占可写 GOCACHE；
产品仍保持只读模块及离线设置。

修正夹具后的有效 RED：35.143 秒，13 PASS、5 FAIL、0 Skip，退出 1。
失败项为 windows-toolchain、parent-cancel、output-limit、stderr-limit、natural-exit-with-child。

## 已修复及最新 Linux 结果

原版本正则允许 go1.26.5 windows/amd64；负向夹具实测被错误接受。现对原生 Bash 路径
明确拒绝 Windows Go 平台；不改变版本、时间预算或 Windows Bash 提前拒绝路径。
修复后的相同专项测试：34.833 秒，14 PASS、4 FAIL、0 Skip，退出 1。
windows-toolchain 从失败变为通过，其余原通过项保持通过。

| 未解决用例 | 观察 | 源码对应边界 |
| --- | --- | --- |
| parent-cancel | 强杀实际 Bash 后，两个自有进程仍存活，临时捕获目录仍存在 | 清理依赖入口 finish/trap；强杀时不能执行该路径 |
| output-limit | 输出超过 4 MiB 后继续阻塞，10 秒独立测试保护到期；未返回预期失败 1 | 先等待 timeout/工具结束，再检查文件大小 |
| stderr-limit | 同上，发生在 stderr | 同一捕获顺序问题 |
| natural-exit-with-child | 主工具正常退出，子进程仍存活 | 正常完成只删除捕获文件，没有子进程清理 |

进程用例同时检查独立外部 canary；本轮没有发现其被终止。已取得的自有进程对象由夹具
负责清理，整个容器生命周期是最终隔离清理边界；不搜索或终止宿主无关进程。
测试的 2 秒／30 秒期限只替换私有脚本副本，产品 900 秒不变。超限用例的外层 10 秒
保护是失败证据，不能记为产品正常超时或输出限制成功。

## 接续边界

共享夹具调整后的 Windows 包级专项回归已完成：
`go test -v -tags=e2e ./internal/e2e -run '^TestDevtoolsVerification(PowerShell|Bash)$' -count=1 -timeout=5m`，
退出 0，125.046 秒，28 个子用例 PASS、0 FAIL、1 个 Git 包装器专属不适用用例 Skip。
这确认本轮 Windows 专项没有新增失败，不抵消 Linux 的四项失败，也不是完整项目回归。

当前 timeout/trap 组合不足以证明 Unix 强杀与正常退出的子进程清理；单纯延长等待或
修改期望不能修复。需要评估独立于入口强杀而仍能清理自有进程的 Unix 监督方式，
同时验证所有权、监督初始化窗口、自然退出、输出上限、捕获清理和外部 canary。
批准设计明确排除新的监督程序，不能把新原生帮助程序或新环境前提静默加入正式入口。
在这个边界明确前，不将当前红色专项标记完成，不进入任务 2 的模块迁移。

本轮未执行完整普通测试、完整 C11、C12/bootstrap 或物理 gate，没有改变防护。
现有 Defender/I2/I3 阻塞不变。诊断原始日志留在忽略工作目录；本说明仅保留脱敏结果。
