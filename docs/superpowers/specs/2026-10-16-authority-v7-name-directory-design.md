# Task11 首切片范围修订：146项名称只读目录

状态：用户在7c6194bb后批准缩小首切片范围，并在a4af66ca后批准本书面修订。
[实施计划](../plans/2026-10-16-authority-v7-name-directory.md)已获批准，按Native/inline执行；
实现与验证状态以[结果记录](../../roadmap/2026-10-16-task11-name-directory-implementation.md)为准。
仅提供精确schema/object名称目录与独立目录向量，角色及证据策略整体后置。
本轮没有产品代码、向量、测试或负载执行，不修改原批准规格集。

## 1. 目的、依据和取舍

用已确定的名称映射交付一个小而可验证的基础组件，不以未决策略阻塞名称目录，
也不把名称存在冒称认证成功。[v7规格§4.1](2026-08-24-nodecontrol-authority-v7-upgrade-design.md)
固定146个V1 object/schema精确对应；[来源索引](../../roadmap/2026-10-16-task11-registry-reconciliation.md)
记录完整集合与此前误计表头的纠正。协议对象没有增加或删除。

相比完整元数据注册表，本方案少交付角色/证据资格，不能用于授权；相比先解决
所有策略例外，它能先验证名称、隔离和可恢复性。该取舍经用户明确选择，不是
把未知策略填成null/[]后偷偷交付。完整Task11与安全/生产门槛仍未完成。

## 2. 精确公开面

未来在internal/nodecontrol/contracts/schema_names.go实现以下独立名称API：

```go
type AuthorityV7SchemaName struct {
    Schema string
    ObjectName string
}
func LookupAuthorityV7SchemaName(schema string) (AuthorityV7SchemaName, error)
func AuthorityV7SchemaNames() []AuthorityV7SchemaName
```

未知输入统一返回有限sentinel ErrUnknownAuthorityV7SchemaName和零值结果；错误不
回显输入。精确区分大小写，不TrimSpace、不转换case、不接受别名或旧版本降级。
查询只按schema，不增加caller注册表、运行时register、body工厂或反射类型入口。
字符串标签不是已实现Go DTO或codec的证明。

列表按Schema的ASCII/bytewise升序，每次返回独立slice。条目只含string值，不含
共享可变集合；调用方修改slice、交换条目或替换字段不得改变后续lookup/list。
内部固定表不暴露，不读取Markdown/环境变量/可变配置，不允许caller覆盖。

没有SignerRoleOrNull、EvidenceKinds或其占位字段；没有Verify/Sign/ParseBody、
授权capability、可信key/root/policy、生成器注册、provider/readiness接线。
名称查找成功不允许caller签名、发布策略、解锁reader或把对象作为bundle证据。

## 3. 集合与未来文件边界

- internal/nodecontrol/contracts/schema_names.go：固定146项名称与上述API/sentinel。
- internal/nodecontrol/contracts/schema_names_test.go：独立集合、拒绝、排序/复制测试。
- testdata/c12/authority-v7/schema-name-registry.v1.json：完整两字段字面量映射。
- testdata/c12/authority-v7/schema-name-registry-golden.v1.json：目录正反向期望与拒绝向量。

这些是将来实施文件，本轮不创建。不得占用原schema-registry.v1.json或完整
canonical-envelope-golden.v1.json来表示部分元数据/codec已完成；两份目录fixture
只覆盖name-directory，不包含伪造的body、signature、认证策略或证据资格。

不得登记AuthorityEffectCommitmentV1、AuthorityEffectResolutionV1、
ActivationDecisionEvidenceV1为新v7项；原Abort/serving历史支持保持原样。
不改私有canonical/signature helper、SQL、生成物、部署配置、approved spec-set
或现有普通进程/安全门槛；不从contracts反向import authority。

## 4. 独立验收要求

未来测试先固定独立字面量oracle再实施：146项schema和object逐对比较、一对一
且与原§4.1集合相等，不能只检查数量，也不能从生产list自动生成expected。
正常lookup必须返回对应两字段；缺一加一抵消总数、错配、重复schema/object、
多项、漏项、空/大小写/空白/未知/历史名称/版本变体均有明确反例。

fixture检查拒绝缺失/额外/重复字段、错误JSON类型和尾随内容；这只是测试fixture
验证，不额外暴露运行时JSON注册/导入API。验证稳定排序及返回所有权，防止修改
一次返回影响另一次；结构/公开面检查确认无策略字段或验证capability。

不新增加密或摘要API，不测试JCS、domain framing或真验签来扩张名称切片。
已有talenro.c12.摘要前缀决定不变，相关codec向量留到相应后续交付。
具体离线RED/GREEN及回归命令、工具指纹和预算由后续实施计划冻结，本轮未运行。

## 5. 后置责任和交接

[元数据初审](../../roadmap/2026-10-16-task11-registry-metadata-audit.md)的93项作者分类
与53项未决完整保留；本修订不认证它们、不把例外改为通过。C1/C2局部批准、
canonical集合登记、部署兼容、动态trust policy、message exact-cover及物理验证
各自边界不变。后续策略层须引用同一名称集合，不能建立竞争名称标准；角色/证据
扩展须先关闭其契约审查，不能给本目录默认附加权限。

本修订仅替代[原注册表设计](2026-10-16-authority-v7-registry-design.md)中**首个可实施
交付**的范围、文件和API；原完整元数据目标作为后续门槛保留。原元数据计划的
“任何未决阻塞完整表”继续有效，但不阻塞这个完全不携带元数据的名称组件。

用户批准本文后才编写名称目录实施计划，再审阅计划并选择执行方式；这次范围
调整批准不是代码/测试执行授权。当前已有源码和进度从开发分支恢复，不假定
main含最新进展。原工具链/cache/DB状态仍须独立准备。

自检：未把后置等同解决、未创建策略占位、没有认证旁路、独立向量不自证，
名称查找/列表的类型、错误、排序、复制和文件边界均明确；无虚构通过结果。
