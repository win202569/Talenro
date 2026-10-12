# 时间证据两项测试覆盖补充

范围：用户批准的有限测试改动；仅测试和进度记录，不改变产品行为。
这是上一增量两项 Minor 的后续，不重写原始独立审查结论。

## 覆盖

1. TestTimeVerifyMetadataEnvelopeSubstitution 构造两份格式合法、payload
   bytes/摘要完全相同而签名份额及 envelope bytes/摘要不同的 metadata。
   用原封装与观测先验证成功；替换原始封装而保留旧观测必须准确返回
   ErrBindingMismatch 且无结果；显式更新观测 E 后，条件性验证成功并
   返回正确的 P/E。测试不宣称根签名、可信观测或持久化已获认证。
2. TestTimePolicyWindows/TestTimePolicyCheckpoint 对无效时间、有效期外、
   密钥/metadata 窗口绑定和 checkpoint 回退分别以 errors.Is 检查
   ErrInvalidTime、ErrExpired、ErrBindingMismatch、ErrCheckpointRollback；
   成功边界仍保留。

采用真实解析器与 EvaluateModel，无验证结果 mock。期望错误类别来自
已批准规则，不从被测结果反推。错误消息措辞不是断言对象。

## 验证

全部 exit 0：定向测试 4.457s；完整 statetime 3.270s；stateformat
6.073s、stateverify 6.943s、statepolicy 7.198s、serving 10.777s、contracts
20.383s、readiness 5.548s；authority 仅 TestBoundRead/TestCoordinator
11.345s；store 仅编译 6.083s（no tests to run）。固定 Go 1.26.5，
离线环境，均使用 -count=1。没有真实 DB/Docker 或整仓测试。

219 个源码/数据库相关文件 hash 核对：只改 verify_test.go 和
policy_test.go，其他 217 个文件未变。gofmt 和空白检查无诊断，
HEAD 仍为 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee。
本增量两项 Minor 的测试覆盖已补齐，所有已知测试会话已结束。
本轮是既有正确行为的测试覆盖补充，没有产品修复或行为 RED→GREEN，
未人为破坏产品代码以制造失败；不能将通过记录冒充缺陷修复证明。
本轮未重新独立审查。未提交、推送、合并或接入生产。
