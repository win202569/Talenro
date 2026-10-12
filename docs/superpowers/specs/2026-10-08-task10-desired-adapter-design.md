# Task 10：私有 desired 适配器与受保护读取接线

日期：2026-10-08

状态：用户已确认本书面设计及实施计划；2026-10-09 本地实施与限定验证、独立审查完成。不提交或推送，不代表生产验收。

## 目标与选定方案

在既有精确 desired SQL、生成行转换和结构投影之上，完成一次查询从参数
绑定、解码暂存、清理到前后置检查的私有组合。只有后置检查通过才交付
独立拥有的 desiredProjection；它仍不是已验签的 DesiredStateFacts。

选定“适配器＋受保护读取接线”同一阶段，以便同时验证资源清理和失败时
不交付数据。另一方案只实现适配器，会把后置拒绝清理的组合证据延后。
不将证书与 desired 路径重构成通用查询框架，避免本阶段改变已验证路径。

依据：2026-10-08-task10-desired-query-design.md、第一阶段 desired facts
设计、既有 certificate_adapter.go、certificate_reader.go、guardedRead、
authority.BoundReadEpoch。原 Task 10 及 v7 生产门禁不因本设计被替代。

## 私有组件与接口

新增 serving/desired_adapter.go 与 desired_reader.go，以及对应测试文件。
复用现有 SQL、生成文件、desiredProjectionFromRows、discardDesiredProjection
与 guardedRead，不修改 schema，不重新生成 store，不导出新入口。

```go
func queryDesiredProjection(ctx context.Context, db store.DBTX, nodeID uuid.UUID, epoch uint64) (desiredProjection, error)
func readDesiredProjection(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID) (desiredProjection, error)
func readDesiredProjectionWithEpoch(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID, resolve func(store.DBTX) (uint64, error)) (desiredProjection, error)
```

最后一个函数仅为私有组合测试接缝；真实私有入口固定使用
authority.BoundReadEpoch。不得变成公开参数、配置或可替换全局变量。
queryDesiredProjection 的裸 epoch 仅供内部组合及隔离测试，不证明来源。

## 参数与执行顺序

1. reader/adapter 首先检查 context：已取消返回 ErrCanceled，nil 返回
   ErrAuthorityUnavailable。reader 然后验证 node UUID，无效返回
   ErrInvalidArgument 且不调用 source；adapter 先检查 DBTX 非 nil（含
   typed nil），再验证 node 与 epoch，epoch 范围 1..MaxInt64。
2. reader 经 guardedRead 进入同步 callback；在原始 callback DBTX 上先
   调用 BoundReadEpoch。只有取得合法 epoch 后才允许 adapter 包装 DBTX。
   不对扫描装饰器调用 epoch accessor，不再取连接或另开事务。
3. adapter 从 node UUID 的规范 16 字节派生 node scope SHA-256，使用固定
   `TALENRO-NODE-AUTHORITY-SCOPE-V1\x00` transcript；global scope 使用
   `TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust`。这里的
   `\x00` 表示单个 NUL 字节，不是四个可见字符。沿用 authority/证书规则。
4. 调用真实生成方法 ListExactDesiredProjection，绑定 node、epoch、两个
   scope digest。恰好一次查询，不重试、不回退、不接收任意 scope。
5. 成功解码后调用真实 desiredProjectionFromRows；保留固定根版本 1 起点、
   active 版本 1..33、最多 32 段链等全部原校验，不在 adapter 再定义业务规则。
6. adapter 清理查询暂存后将自有投影交给 guardedRead 暂存。source 后置
   检查成功后才交付；不保留 callback context、DBTX 或源对象在结果中。

context 取消仍优先于其他错误。无效 node 的 reader 前置校验是本地参数
拒绝，不启动读取；已开始的有效请求继续遵守后置检查与错误优先级。

## 查询生命周期与所有权

使用 desired 专用的私有 DBTX/Rows 包装器。真实生成代码的 Scan 在错误时
会丢弃局部结果，所以在调用底层 Scan **之前**保存所有目标指针，以覆盖
部分解码失败和 panic；不能只在生成方法成功返回后清理 rows 切片。

- Query 同时返回 rows 与 error 时，仍接管并关闭非 nil rows；nil/typed-nil
  rows 且无错误按不可用处理。不得调用 nil Rows 的方法。
- Close 包装器保证底层显式 Close 至多调用一次；清理本身可重复调用。
  先登记 closed 状态再调用底层 Close，防止异常退出后重复关闭。
- 清理所有已记录目标中的自有 []byte 内容，再清零目标值及目标指针容器；
  清理可访问的生成结果行容器、参数 digest 和暂存结构。无论成功、查询错误、
  部分 Scan 错误、Rows.Err、取消或 panic，都走 defer 清理路径。
- 清理代码必须在底层 Close 发生 panic 时仍执行内存清理。panic 不转换成
  成功或普通领域错误；正常情况下原 panic 向上传播，不声称可处理进程终止
  或保证多个同时发生的 panic 保留全部异常信息。
- 投影的成功交付必须晚于 adapter 清理完成；清理发生 panic 时，尚未交付
  的自有投影也要 discard。query 完成后观察到取消，同样 discard 后返回零值。
- 对 generated SQL 的 LIMIT 33 再加包装器的最多 33 次 Scan 防御上限。
  第 33 行是异常数量哨兵，完整读得 33 行由 converter 返回 Conflict；若
  底层仍提供第 34 行，停止继续解码并返回不可用，不把前缀作为成功结果。
  这只限制解码行数，不承诺限制 driver 预取或单次底层扫描的分配。

普通 pgx []byte 解码的拥有关系沿用已检查的现有路径。测试替身必须模拟
独立解码缓冲，不能让产品清理意外改写 fixture 源数据后再称为清理成功。
不承诺 driver 网络缓冲、不可达运行时副本、不可变字符串的安全擦除；
不支持调用方并发修改输入或异步保留 callback 能力。

## 错误与后置检查

直接 adapter 的 Query/Scan/Rows.Err 错误统一为 ErrAuthorityUnavailable，
即便底层伪装成 NotFound/Conflict；取消或 deadline 映射为 ErrCanceled。
只有成功查询后的真实 converter 可以产生 NotFound/Conflict。
不在返回值中泄漏驱动错误详情、SQL 或制品内容。

reader resolver 缺失、无效 epoch 或非取消错误均为 ErrAuthorityUnavailable，
不执行 SQL。真实 accessor 拒绝普通 DBTX、装饰器、仿造方法和已过期能力。
测试成功路径用私有 resolver 接缝，不扩大 authority 的工厂可见性。

guardedRead 保持 cancellation > source/authority failure > domain error。
NotFound/Conflict 不跳过后置检查；漏调/重复 callback 失败且不重复查询。
后置检查拒绝、外层取消或 source panic 都丢弃已暂存投影。
source 正常清理自己的 callback context 不应撤销已成功的读取结果，沿用
现有 guard 契约；本阶段不改它来支持异步 source。

## 测试与完成标准

测试保持真实 generated query、converter、projection 和 guard，仅替换
数据库传输和私有 source/resolver。UUID、scope 黄金值及各角色材料分离。

1. 参数绑定：精确 UUID/epoch，两种 scope 与独立字面黄金摘要一致；
   source 原始 DBTX 先解析 epoch，再由同一对象转发查询；一次调用。
2. 正常输出：空链、一段、32 段的完整独立快照；33 行拒绝、34 行限流拒绝；
   成功后驱动暂存清零但已交付投影和 fixture 不变。
3. 生命周期：Query 失败、rows+error、nil/typed nil rows、首行及中间行
   部分 Scan 失败、Rows.Err、Scan/Next/Close panic；保留目标和缓冲别名验证
   清零与关闭次数。panic 测试只在测试边界 recover，不在产品中吞掉异常。
4. 取消：查询前、查询/扫描后、转换后可观察取消、后置检查阶段；返回零值，
   不泄漏暂存；不得为可观测性增加生产全局 hook。
5. 组合：前后置顺序、领域错误仍后检、后检覆盖领域错误、取消覆盖两者，
   不调/重复 callback、无效/伪造 epoch、postcheck panic。通过真实 guard+
   adapter 的测试组合保留 staged aliases，证明 payload/envelope/签名及
   投影容器 discard；固定入口没有未绑定 DBTX 的成功路径。
6. 防回归：先 RED 再实现；深拷贝用逐次修改立即断言，不重复修改共享输入
   抵消变异。独立审查后修复重要问题，保留实际测试与审查范围记录。

仅运行已批准 serving/contracts/readiness、authority 的
`^(TestBoundRead|TestCoordinator)` 与 store `-run '^$'` 编译检查。
固定 Go 1.26.5、PowerShell、离线只读依赖；SQL/生成文件与四个锁文件应不变。
不运行数据库、Docker、完整 C11/C12/bootstrap/PITR/race，不更改安全设置或
绕过安全拒绝，不下载依赖，不暂存提交推送。

## 保留的门禁与审阅

本阶段只产生私有结构投影的读取组合，不是规范解析、验签、可信时间、
当前授权、完整 desired signed envelope、网络大小约束或生产 Reader。
真实物理连接/数据库行为与 PITR、writer 历史保留、跨 epoch 恢复、
v7 proof/incarnation/lease、B01/C12/main 及安全验收仍未关闭。

沿用已批准的当前会话实施与最终独立审查方式。只保留本地变更，不执行
技能默认的提交步骤；证据和审查结论见本阶段 execution ledger。

自查：已明确适配器与 reader 的参数校验顺序、装饰前取 epoch、33/34 行
区别、成功交付前清理、部分解码和 panic 覆盖，以及模拟证据不等于生产证明。
