# P2-A 限定离线验收建议（待用户明确批准）

日期：2026-10-18。候选提交：cc764dba8617933cc1dda341cf46322f0135cf7f。
状态：READY_FOR_LIMITED_ACCEPTANCE_DECISION；LimitedAcceptanceApproved=false。
本文是对现有结果的验收建议，不是实施计划、运行许可或已批准的验收记录。
用户要求继续后，本轮仅核对要求、固定版本和保存证据，没有新测试或审查。

## 1. 建议决定

建议接受 **Windows 普通用户、PowerShell 7.6.5 Core x64 下，
P2-A 纯声明整数时基换算器的指定离线向量结果**。

接受的是固定版本在已测范围内满足数值/拒绝/独立返回合同；
不是所有输入、所有环境或真实运行能力的普遍保证。
完整 P2、完整 C12、PS5、真实来源验证、安全放行及 main 合并仍未获批准。
旧失败、未测项、已消费和冻结额度全部保留。

## 2. 固定版本与恢复入口

代码目录：docs/diagnostics/isolated-fuzz/p2a-type-receiver-repair/。

| 对象 | SHA256 |
| --- | --- |
| offline-timebase-plan.ps1（产品） | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| full.tests.ps1（修正后测试） | B5C4E0CE14F0B08C647A966884BE6496AA9C05A5D5BEA827B6F86F87EAFCAA65 |
| fixture-builder.ps1（独立输入构造） | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| expected-output.json（独立输出预期） | DA3CF1837213F9451A2CB2941DC248B663116CD52EEC199D1D01646A90CA0225 |

业务入口仍为 New-OfflineTimebasePlan([object]$Request)，没有新增真实模式。
产品与原实现字节一致。两次经另行批准的测试修正分别解决输入构造共享存储、
类型辅助函数未识别 Plan 容器的问题；没有修改产品算法或放宽原断言。
修正历史与原失败分开记录，不把新成功回写为旧失败通过。
恢复从现有 codex/c12-b01-task9-coordinator 开发分支开始，不从 main 推定进度。

## 3. 要求与证据对应

依据：[P2 设计 §5/§9](../superpowers/specs/2026-10-18-p2-adapter-boundaries-design.md)、
[P2-A 原计划](../superpowers/plans/2026-10-18-p2a-timebase-plan.md)及后续另获准的
[修正与完整结果](2026-10-18-p2a-type-receiver-repair.md)。
原设计/计划的起草或停止状态是历史，不修改其被冻结身份或旧预算。

| 要求类别 | 保存的指定证据 | 有限结论 |
| --- | --- | --- |
| 完整数值计划、非零起点、null Stop | NullStop/NormalStop/NonzeroOrigin 的完整 literal 与精确类型比较 | 指定向量通过 |
| 余数/大整数精度、保守舍入 | FractionalFrequency/FractionalPhaseRounding/StopCutoff/StageWindows | 指定向量通过 |
| 等值到期和迟到停止，不追加预算 | DeadlineEquality/LateStop | 指定向量通过 |
| Int64 边界及中间溢出不得被 min 掩盖 | OriginLimit/FrequencyLimit/IntermediateOverflow | 指定向量通过 |
| 精确形状、类型、被动描述符及装饰拒绝 | RequestShape/ExactTypes/PassiveDescriptors/DecoratedScalars/RangeRelationships | 指定可构造向量通过，非穷尽 |
| 输入修改、输出独立、连续调用隔离 | InputMutation/OutputOwnership/SequentialIsolation | 指定顺序突变向量通过 |
| 拒绝完整形状、事实及权限字段始终 false | RefusalShape/ConservativeFlags/Complete | 指定成功和拒绝向量通过 |
| 测试能检出指定舍入/到期错误 | floor→ceil 与 >=→> 两项单独 mutant 精确失败 | 两项负控制有效 |
| 固定产品还原后结果一致 | restored final 与 repaired full 均为 24PASS/27行 | 两次全套结果一致 |
| 独立只读审查 | 一次 fresh-context 审查，无 Critical/Important/Minor finding | 无本次范围内阻塞项 |

实际五轮：RED exit1/1193ms；full exit0/1782ms；phase-mutant exit1/814ms；
expiry-mutant exit1/819ms；final exit0/1829ms。负控制 exit1 是预期，
不代表把失败的全套测试算作成功。五轮均普通上下文、stderr 空、无超时、
owned exit、完整并发分流捕获、文件校验完整。
24 是组数，不是覆盖率或独立需求数量。

本轮只读复核839映射项、五次精确结果、十份原始流哈希及当前三份核心文件身份；
没有启动目标、导入产品、重跑测试、增加 reviewer 或改变源码。

## 4. 必须随验收保留的限制

- 仅限上述普通 Windows PS7.6.5 Core x64 与固定版本，不包含 PS5、其他引擎、
  操作系统或提权环境的可用性结论。
- 仅声明整数计算，不读取真实 QPC/UTC，不证明调度精度、硬墙钟截止、真实身份、
  会话所有权、停止/取消、双管道、清理、存储或发布。
- CodeProperty 拒绝及 recoverable internal-error 路径仍只有静态说明，不冒称动态覆盖；
  不保证灾难恢复、内存耗尽、并发调用者修改或敌对全局 type-data。
- 字符串 held-view 的元数据突变与字段再次读取的持久性不同；测试保留原区分，
  不据此承诺所有字符串重新读取都保留附加属性。
- 控制器不是安全边界；本次结果不证明恶意 ledger、并发控制器、后代进程、
  操作系统故障或超时清理分支的普遍正确性。
- 审查验证的是源码与保存证据的一致性，不是历史执行的独立认证。
- 旧停止路线及两次诊断不因本次成功改变；原2/5与后续3/6路线的余项继续冻结。
  最新5/5和1/1审查均已消费，批准验收不续期任何额度。
- SourceVerified、LiveActionsPerformed、RuntimeApproved 三字段仍为 false。
  用户的有限离线验收不应改写这些产品事实字段。
- P2-B/C、P3、Stage B、I2/I3、完整 C12、真实运行、安全 owner/vendor 结论、
  main 合并和发布不在这次决定内。

## 5. 用户批准后的唯一动作

若用户明确批准本文限定验收，只登记该固定版本与范围已获用户接受，
同步 current-status 和 repository-recovery，按长期授权提交推送必要文档。
不改产品、不重跑、不解冻、不提升其他入口、不合并 main，不自动开始 B/C/P3。
下一阶段仍须单独明确目标、设计和执行授权。

本文尚未批准；普通“继续整理”不记作批准以上验收。
请对这份具体范围作明确决定。

## 6. 证据与换机边界

完整有序输出、源码/原始记录身份、审查结论及逐项限制已保存在上述结果报告。
原始 packet 保留在本机 ignored 目录，不提交、不删除；clone 不恢复这些记录、
历史映射、引擎缓存或会话。换机可恢复必要源码、预期、报告和消费状态，
但不因此拥有自动重播许可。旧计划勾选不回写成原路线成功。
