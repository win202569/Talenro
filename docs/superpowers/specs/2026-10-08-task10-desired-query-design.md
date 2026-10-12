# Task 10：精确 desired 查询与固定根链历史设计

日期：2026-10-08

状态：用户已确认本书面设计；允许编写实施计划，计划审阅前不实施。不提交或推送。

## 目的与边界

将第一阶段私有 desiredProjection 接到有明确来源约束的数据库投影契约，
确保只读取同一节点当前指针指定的 desired、root、metadata 及完整轮换链。
本阶段产物为组合 SQL、对应生成接口及私有行转换，不是公开 DesiredStateFacts。
查询适配器、bound reader 接线、规范信封构造和生产入口留在后续阶段。

依据：原 B01 实施计划 Task 10 Step 2、C1.2 原设计 7.1，以及
2026-10-08-task10-desired-facts-design.md。既有证书查询保持不变。

用户已选择从固定初始根版本 1 开始、最多 32 次轮换的完整链策略。
另一选择是可移动历史窗口；它需要额外的可信边界、保留及旧节点恢复规则，
本阶段不引入。也不修改冻结的只接 node ID 的最终读取接口。

## 1. 固定历史策略

- ChainBaseRootVersion 固定为 1，不接受调用方、数据库最早记录或客户端
  high-water 作为可变起点。版本 1 的信任来自受信 host deployment，
  数据库记录不是 trust anchor。
- active root version 必须在 1..33。所需轮换版本为 2..active version，
  数量精确等于 active version - 1；第 33 次轮换即版本 34 必须拒绝读取。
  即使客户端已更新到较新版本也不绕过该限制。部署必须在达到上限前
  另行批准并实现窗口升级；本方案不会自动重置版本或删除历史。
- 根版本 1 时轮换链为空，但仍须有 active root 指针精确对应的真实已发布
  root 行与 fence。不得合成 0→1 轮换、空 envelope 或替代 publish ID。
  如果写入侧尚不能提供该行，读取失败；本阶段不增加 bootstrap 写入路径。
- 根版本大于 1 时只返回版本 2 起的真实轮换。无需把版本 1 的数据库行
  当作额外网络轮换或信任证据；首段 previous/base root version 必须为 1。
- 每段 base 精确等于前一版本，authority epoch 与 bound epoch 一致，
  sequence 严格递增，publish ID 不重复；末段必须精确等于 active root。
- 不按当前时间过滤历史、不截取最后 32 段、不使用 MAX/latest 或孤立行回退。
  保留过期历史不延长其有效期。B03 将来仅从完整事实选择连续后缀；
  客户端需要的某段已过期时，仍须走原规范的受信恢复流程。

## 2. 一条组合查询、精确指针

新增 ListExactDesiredProjection :many，输入为 node ID、已验证 bound epoch、
派生的 node scope digest 和固定 global-node-trust scope digest。
scope 派生复用现有 authority 规则，不接受任意 scope 字符串。
epoch 的最终来源必须是同一 bound DBTX 的 BoundReadEpoch；本阶段不接入口。

查询只从 node_inventory 的 exact node ID 开始。当前 desired generation、
root publish ID/version、metadata publish ID/version 必须取自同一 inventory 行。
不得从 desired 表或 publish 表寻找“最接近”的替代当前状态。

以下关系必须全部精确匹配：

1. desired 的 node/generation 等于 inventory 指针；其 root/metadata ID 与
   version 等于同一 inventory 行的对应指针。
2. signing intent 的 signing ID、node、reserved generation、operation ID、
   epoch、sequence 与 desired 相同，kind=desired、status=active；payload、
   payload digest、signature、expected key ID、root/metadata version、
   captured inventory version 分别等于 desired 中的对应值。
   signature_verified_at、terminal_at 必须存在且有限，failure_reason 为空。
3. current root/metadata 按 publish ID 与 version 精确连接，kind 分别为
   root/metadata，status=active，terminal_at 存在且有限、failure_reason 为空，
   payload 和已发布 envelope 及摘要完整。root base=version-1；metadata
   base metadata=version-1 且 base root=active root version。
4. 每个制品的 fence 都按 operation ID/epoch/sequence 精确连接，要求
   provider_status=committed、visibility_state=active。desired 的 effect
   为 desired_activate，scope 为 exact node；root/metadata 的 effect 分别
   为 root_publish/metadata_publish，scope 为固定 global_node_trust。
5. 所有返回制品的 epoch 等于输入 bound epoch。不能用另一个 epoch 的历史
   补齐链，也不能通过网络读取完成跨 epoch 信任迁移。

本阶段不把 desired.inventory_version 强制等于当前 inventory_version：
读取契约是精确 active desired 指针，而非重新执行激活流程。捕获的 identity、
security 版本和 expected public-key digest 也不能被误称为当前授权校验。
完整签名密钥、规范 body 和当前安全状态检查属于后续规范/服务门禁。

## 3. 历史行与返回形状

历史轮换必须 kind=root、reason=root_rotation、status=active，并有完整
已发布制品及 exact committed-active root_publish fence。这里 active 表示
该发布成功终结，不等于它仍被 inventory 当前指针选中。仅因出现更新版本
而失去指针引用，不构成接受 failed/superseded 历史的理由。
本设计不声称已有写入路径会保留这种状态；未来写入方必须满足该契约，
不能满足时读取拒绝，不放宽查询去接纳不确定的历史。

采用一个语句、同一语句快照：当前部分作为固定头，按版本升序返回所需
轮换行，每行附带相同的当前头。root=1 返回一行且明确标记无轮换。
候选版本范围先用 active version 1..33 限定，再生成 2..active version
的预期槽位并 LEFT JOIN 对应历史行。缺行/无效 fence 保留无效槽位，
不可经 INNER JOIN 静默消失。不得对未限制的 bigint 生成巨大序列。

当前部分同样保留 inventory 的存在性及指针完整性，使用显式有效性标记
或 nullable 字段表达缺失关系，而非把无效关系归并为“节点不存在”。
合法 root=1 一行，合法 root>1 恰为 active version-1 行；最多读取 33 行，
第 33 行只作异常数量哨兵，不作为截断后可成功的事实。
无效当前头只需返回一条诊断行，不返回制品；不在 SQL 中构造完整 signed
desired envelope，数据库现有的是 payload 与 signature 分列。

固定头重复带来的字节成本有明确上限，但可能超过最终网络响应 1 MiB；
数据库暂存上限不是放宽网络上限的依据。本阶段不创建网络响应。

## 4. 私有转换、错误与所有权

私有转换只接受生成行、预期 node ID 和 epoch，输出 desiredProjection。
先检查 nullable、时间有限性、正整数/长度及各有效性标记，再做类型转换。
拒绝头部字段不一致、重复/乱序/缺失槽位、非预期行数、错误 epoch/fence。
头部一致性比较包括所有复制的身份、坐标、摘要和 bytes，不能只比版本。
只有通过以上来源检查后才调用第一阶段构造函数；强制 chain base=1。

- 输入 node/epoch 无效：ErrInvalidArgument。
- inventory 不存在，或 active desired 未设置且其余指针组合法：ErrNotFound。
  root/metadata 可先于 desired 发布；不把这种正常前置状态当作缺链。
- 已设置 desired 但缺少对应行/指针/制品，部分指针组、越界或关系不符：
  ErrConflict，零投影。现有 SQL 故障不得伪装为 NotFound。
- 后续 adapter 的 Query/Scan/Rows 错误映射为 ErrAuthorityUnavailable；
  bound composition 保持 cancellation > authority > domain 的既有优先级。

成功值由现有构造函数深拷贝，行转换不修改调用方输入。后续 adapter 必须
负责关闭 rows、清理扫描暂存及失败投影，包括部分解码和 panic；本阶段
没有该 adapter，不能据此声明整个 SQL 读取生命周期已完成清理。
不承诺 driver 缓冲、不可变字符串或运行时副本的安全擦除。

## 5. 实施范围与验证标准

后续计划可涉及：nodecontrol_serving.sql、对应 sqlc 生成文件/Querier，
新增 serving/desired_row.go 与单元测试，以及范围内文档。无需 schema 迁移。
固定版本离线 sqlc 重生成两次并核对稳定性；不下载依赖或改锁文件。

测试使用彼此不同的 UUID，覆盖版本 1 空链、版本 2、版本 33 的 32 段链，
版本 34 拒绝；中间缺段、重复/乱序、尾部不符、每种 intent/fence 错误、
epoch 混用、NULL/无限时间、长度边界、同头不一致、零行及异常行数。
每项绑定至少有“合法但属于另一对象”的错误值，避免只测零值。
分别修改输入和输出的每一组 bytes 验证深拷贝，避免共享输入被重复异或
恢复原值的问题；这会吸收前次 desired ownership 测试审查的经验。

只运行既定允许范围：serving、contracts、readiness；authority 仅
TestBoundRead/TestCoordinator 组；store 仅编译。设计/生成检查和模拟行测试
不能证明真实 PostgreSQL 连接、快照、查询执行或 PITR 行为。
真实数据库、Docker、全套 C11/C12/bootstrap、race 与安全验收仍保留阻塞；
不更改 Defender、不加排除项、不绕过平台保护、不执行受阻命令。

即使上述限定测试通过，也仅完成私有 SQL/行投影阶段，不等于 Task 10、
B01、C12、生产或 main 验收。canonical 解析、签名验证、desired envelope
生成、可信时间、v7 proof/incarnation/lease 及生产 Reader 仍在后续门禁。

## 6. 审阅与后续门禁

本文件仅记录书面设计；默认提交步骤不执行，所有既有本地改动保留。
用户审阅通过后才编写实施计划；计划获批后才实施。实施继续采用当前会话
执行及完成后的独立审查方式，不据本设计扩大测试或远程写入授权。

自查：已区分初始根和轮换、32 次与版本 33、缺失与损坏、结构与授权、
历史保留与时间有效、模拟证据与数据库证据；无待填占位项。
