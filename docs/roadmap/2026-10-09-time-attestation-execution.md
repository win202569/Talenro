# Task 10：时间证据隔离验证执行记录

日期：2026-10-09。设计及实施计划均经本会话用户批准；Native 主代理实现。
当前：实现、修复后限定回归及一次独立审查完成；重要问题已修复，
两项轻微测试覆盖建议已在用户批准的后续补充中关闭，见
2026-10-09-time-attestation-test-followup.md。未提交或推送，不是生产验收。

## 范围

stateformat 新增 TimeAttestation：严格封装、32-byte nonce、独立 checkpoint、
最长 5 分钟窗口、独立 P/E 身份与签名消息。新增 statetime 纯模块，
针对显式当前 metadata 模型执行 Ed25519、请求绑定、时间/checkpoint
判定和完整快照重检。无生产消费者，无时钟/网络/存储访问。

metadata 原始 bytes 与观测精确绑定不是根信任认证；观测可由调用方伪造。
成功只表示模型条件成立。不会验证 TLS、安装根、消费请求、提交状态或
建立可信时间。重复输入同一未消费请求模型仍可成功；这不是防重放。

## 验证实证

所有测试由主代理运行，固定 Go 1.26.5，离线环境，-count=1。

| 阶段 | 观察结果 |
| --- | --- |
| stateformat 基线 | exit 0，1.866s |
| 格式 RED | 四项新测试因 Kind 不支持而失败，exit 1，3.227s |
| 格式 GREEN（含冻结向量） | exit 0，1.720s |
| 验签 RED → GREEN | 缺少模型/API 编译失败 → TestTimeVerify* exit 0，1.756s |
| 时间规则 RED → GREEN | 缺少 EvaluateModel 编译失败 → 全 statetime exit 0，1.651s |
| 重检及边界 RED → GREEN | 缺少 RecheckModel/测试 guard 编译失败 → 全 statetime exit 0，1.791s |

修复前限定回归（全部 exit 0）：

| 范围 | 结果 |
| --- | --- |
| statetime | 6.594s |
| stateformat | 8.400s |
| stateverify | 5.866s |
| statepolicy | 6.674s |
| serving | 7.220s |
| contracts | 6.302s |
| readiness | 3.741s |
| authority，仅 TestBoundRead/TestCoordinator | 9.455s |
| store，仅编译（-run '^$'） | 3.543s，no tests to run |

修复后同范围再次验证：statetime 6.744s、stateformat 8.630s、stateverify
6.063s、statepolicy 5.987s、serving 13.079s、contracts 9.850s、readiness
4.122s、限定 authority 5.296s；store 只编译 2.821s，no tests to run。
全部 exit 0；没有扩展测试范围。

格式检查无待格式化文件；tracked git diff --check 无诊断。
冻结格式向量由 PowerShell/.NET 独立计算 P/E/hash/transcript，再由测试消费。
接口不存在导致的早期编译失败仅作为新增 API RED，不能冒充行为反例。
格式拒绝的行为 RED 与后续完整规则表分别保留。

## 保护与裁定

实现前记录 289 个既有文件 hash；实现后仅 stateformat/artifact.go 和
envelope.go 两个允许修改的旧源码发生变化。最终核对 284/289 未变，
另有本增量设计、计划和 current-status 三份允许更新的文档；无其他
既有文件变化。新文件已执行格式/空白检查，所有测试会话已结束。
HEAD 保持 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee。
新代码不修改现有 stateverify/statepolicy 的产品源码或依赖边界。

1. PowerShell/apply_patch 账本替代 Bash 技能脚本，满足 Windows 约束；
   风险是人工记录遗漏，使用持久账本和 hash 基线补充。
2. 保留 scratch，不下载/提交/执行整仓测试，批准范围优先于技能模板；
   代价是跨机保存和广泛验证仍未完成。
3. 第 5 项边界测试在第 4 项收尾时提前声明；缺失 helper 的 RED 后才实现，
   两项共同通过全包 GREEN。代价是编译依赖导致两项测试验证时点相邻。

## 最终独立审查

fresh gpt-6-astra/high 只读源码审查：0 Critical、1 Important、2 Minor。
重要问题为隔离 guard 只记录最后一个 time 导入别名，较早别名的 Now
可漏检。没有发现当前产品源码读取真实时钟。主代理复现
TestTimeBoundaryMultipleAliases RED 1.610s/exit1，改为记录所有别名，
针对性 GREEN 1.556s/exit0。没有第二轮独立审查。

原始审查的两项轻微建议（现已补充测试，保留当时发现供追溯）：

- 构造真正同 P、不同 E 的两份 metadata 原始封装替换测试；当前测试
  修改观测 E 摘要，产品 P/E 比较正确，但未直接覆盖该完整情景。
- 对负向 policy 测试断言准确错误类别；当前多数只断言失败，未发现
  产品实现误分类。

审查者未运行测试。其未判断项由主代理逐项保留为限制：根安装与观测
真实性、TLS/响应绑定、nonce 随机性/原子消费/并发重放、真实时间与
平台回退、持久 checkpoint/恢复/硬件/隔离、并发输入/Clear、不可变及
分离/运行时副本清除、部署协议与签名服务兼容性、历史工作和整体
验收、测试运行结果、最终报告及后续修复。后两项由主代理自行验证，
不冒充独立验证。对应风险分别为伪造模型、重放、崩溃/重启安全未知、
竞态/副本残留、不兼容及范围外缺陷未排除。

## 尚未验收

真实平台来源认证、跨重启可信时间、防回退持久化、TLS、nonce 原子消费、
生产发布兼容性及整体 Task10/B01/C12/main。无真实 DB/Docker、整仓、
C11/C12、bootstrap、PITR 或 race 测试。并发修改/Clear 不支持。
这是局部开发结果，不是生产安全或合并验收。
