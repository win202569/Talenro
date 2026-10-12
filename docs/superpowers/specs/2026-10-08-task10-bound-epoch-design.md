# Task 10：同连接读取的 epoch 绑定设计附录

日期：2026-10-08。状态：用户已批准本书面附录及实施计划，选择 Native 执行。
本文件不是实施完成、生产授权或 Task 10 验收证明；实际结果见执行记录。
不暂存、提交或推送。

## 目标与既有约束

证书读取所用 authority epoch 必须来自执行这次查询的读取源，在同一
物理连接上的前置 readiness 校验结果，而不是调用方提供的数字、另一次
CheckReady 调用或查询结果中的最大 epoch。后置校验成功前不得发布 facts。

保留冻结签名：
`WithConsistentReadyRead(context.Context, func(context.Context, store.DBTX) error) error`。
保留已有错误优先级、连接身份检查、前后快照比较、WAL 不回退和失败清理。
不新增生产 reader/source 工厂、HTTP 路由、数据库迁移或依赖。
普通 readiness 仍不能替代 Task 15 的 v7 incarnation、activation 与 lease 证明。

## 选择与替代方案

采用 authority 包私有的连接包装对象，将已冻结的 epoch 与原连接放在同一
对象内，仍通过 store.DBTX 交给回调。serving 通过 authority 提供的只读
访问函数获取该对象的 epoch，不接受外部 epoch 参数。

不修改回调参数列表，以免更改冻结契约；不独立查询 epoch，以免来源与
前置校验脱离。也不采用仅凭公开 Epoch 方法的结构接口认证：其他包能
自行实现该方法，不能据此证明对象由 authority 创建。

## authority 包契约

新增私有包装类型，包含原 store.DBTX、epoch 数值与回调有效期标记。
字段不导出，不提供构造函数，不暴露底层连接解包接口。它将 DBTX 操作
转发到前置校验所用的原连接，不再获取连接或创建独立事务。
“只读”指 epoch 元数据访问，不表示 DBTX 的 Exec 能力被撤销，也不构成
数据库只读权限控制；本增量的证书读取路径仅执行既有生成查询。

包装对象仅在前置校验成功后创建。要求 ProviderHead.Epoch 与
DatabaseHead.Epoch 相等且位于 1..MaxInt64；不满足时返回
ErrAuthorityUnavailable，不调用业务回调。合法空数据库继续采用已有
readiness 层将 DatabaseHead.Epoch 归一到 provider epoch 的行为，包装层
不得另行推断或修正 epoch。epoch 从已冻结前置快照复制一次，不保留其指针。

新增跨包访问函数 `authority.BoundReadEpoch(db store.DBTX) (uint64, error)`：

- 只接受 authority 私有包装类型的确切非 nil 指针，并检查有效期及数值范围。
- 普通连接、typed nil、伪造同名方法对象、额外装饰器、失效包装均返回
  零值与 ErrAuthorityUnavailable；不采用结构接口断言或递归解包。
- 访问函数不查询数据库、不检查 provider、不接受 epoch 输入，不提供授权。
- 回调正常返回或 panic 时，立即失效包装，再执行原有后置校验/资源释放。
  失效时清零 epoch；后置检查继续使用原连接。所有清理仍必须覆盖 panic。

回调同步使用对象，不得保留、异步使用或跨次复用 DBTX；生命周期标记用
同步安全方式读取/失效，避免访问函数与失效之间出现数据竞争。本设计不
声称能撤销已被复制出的整数、回收恶意代码持有的原连接或隔离进程内攻击者。
类型限制只是防止意外组装，最终可信性仍依赖 Coordinator-only 创建路径。

## serving 包组合

新增私有证书读取函数，参数只有 context、BoundAuthorityReadSource 和
CertificateLookup，没有 epoch 参数。复用 guardedRead，在其回调内首先
调用 BoundReadEpoch；成功后将同一个 DBTX 与返回 epoch 交给
queryCertificateFacts。不得从 facts、context value、全局变量或另一个
连接补取 epoch。现有底层 adapter 保留显式 epoch，作为私有分层实现细节。

缺失或无效绑定不执行证书 SQL，返回有限的 authority 错误。查询结果仍暂存，
仅源返回成功后发布；取消优先，其次 authority/postcheck 失败，其次领域错误。
继续使用 discardCertificateFacts 清理未发布的序列号、DER 和 trust 数据。
不承诺清除不可变字符串或运行时副本。此次不新增公开 Reader 构造方法。

## 验证设计与证据边界

authority 隔离测试覆盖：

- 非默认 epoch 精确传递；全部 DBTX 转发使用原连接；无额外连接获取。
- epoch 为零、超界、provider/database 不一致时不进入回调。
- 普通/typed-nil/伪造对象不能取得绑定；有效期结束及 panic 后访问失败。
- 后置 epoch 漂移、连接替换、取消仍失败；原资源释放与错误优先级不退化。

serving 隔离测试覆盖：

- 缺失真实绑定时不执行查询；不能由任意 Epoch 方法或调用方数字绕过。
- 查询参数收到解析得到的精确 epoch；错误路径零 facts，成功路径保留副本。
- 后置拒绝和取消清理暂存数据，缺失/重复回调继续失败。

包隔离不能成为添加导出测试构造器的理由。serving 内可使用私有测试接缝
验证组合算法，但必须明确其证据仅覆盖算法；authority 内测试真实私有包装
及公开访问函数。生产组合入口固定使用真实 BoundReadEpoch，不允许运行时
配置或全局可替换解析器。不把分层测试称作真实 Coordinator 到数据库的端到端验证。

允许的回归范围：serving、contracts、readiness；authority 仅
`^(TestBoundRead|TestCoordinator)` 所选组；store 仅 `-run '^$'` 编译检查。
新增 source 测试沿用 TestBoundRead 前缀，保证被许可选择器覆盖。
使用 PowerShell 和现有离线固定 Go；锁文件不得改变。实施需记录先失败后
通过证据，并检查成功/失败两种清理对照，不以模拟测试代替物理数据库证明。

## 不在本次验收范围

真实 PostgreSQL、连接故障物理实验、PITR、完整 C11/C12/bootstrap、Docker
权威验证、Defender 历史告警处理、Task 15/18 门禁、main 合并和生产授权均
保持未完成，不改变安全设置或下载依赖。之前 row fixture 使用不同 UUID
以加强字段交换检测的可选改进也不被本附录自动标记完成。

## 预期修改面与完成标准

实施预计触及 authority/bound_read.go 及隔离测试、一个 authority epoch
绑定小文件、serving 私有证书读取组合及测试、必要契约注释和进度文档。
不改生成 SQL/store、schema 或业务端点。旧回调收到的动态对象从原连接变为
包装，但连接身份断言仍针对原物理连接；测试应验证转发行为而非接口对象地址。

本增量完成标准是：生产可达路径没有增加；epoch 不再由私有高层读取器
调用方传入；来源绑定、生命周期与失败清理有上述隔离证据；既有安全和
真实数据库验收阻塞完整保留。书面设计获批后编写计划，再由用户选择执行方式。
