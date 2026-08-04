# Talenro 品牌命名体系设计规格

- 状态：已完成方案评审，待用户审阅书面规格
- 日期：2026-08-04
- 主品牌：`Talenro`
- 官方中文名：`泰联诺`
- 消费者应用商店名：`Talenro VPN`
- 关联规格：[强网络限制环境 VPN 产品架构设计](./2026-08-04-resilient-vpn-architecture-design.md)
- 商业规格：[VPN 商业客户端、账号、计量与支付设计](./2026-08-04-vpn-commercial-client-design.md)
- 本地化规格：[VPN 五语本地化与 RTL 设计](./2026-08-04-vpn-localization-design.md)

## 1. 执行摘要

产品主品牌正式定为 `Talenro`，官方中文名为 `泰联诺`。品牌定位是克制、可信、国际化的私密连接基础设施，而不是强调“翻墙”“免费”或夸张匿名承诺的工具型品牌。

`Talenro` 是所有国家和语言中不变的拉丁字母主标识；`泰联诺` 是简体中文市场的官方本地化名称。俄语 `Таленро`、波斯语 `تالنرو` 和日语 `タレンロ` 仅用于发音提示、正文首次介绍或搜索辅助，不替代主 Logo，也不产生独立子品牌。

消费者应用在 Apple App Store、Google Play、官网和桌面安装包中统一命名为 `Talenro VPN`。安装后的短名称和品牌 Logo 使用 `Talenro`。品牌描述语为 `Private Connectivity`；首选英文传播标语为 `Private. Reliable. Connected.`，首选简体中文传播标语为“私密连接，始终可靠。”

本规格是品牌与产品命名决策，不是商标法律意见。公开发布、购买核心域名、冻结包标识符或使用注册商标符号前，必须完成目标市场的正式商标与权属核验。

## 2. 已确认的决策

1. 采用独立造词 `Talenro`，不把 `VPN`、`Proxy`、`Free` 写入主品牌。
2. 品牌人格是克制、可信、国际化，视觉和文案接近安全连接基础设施。
3. 官方中文名为 `泰联诺`；其含义联想可解释为稳定、连接和承诺，但不得声称它是 `Talenro` 的词源翻译。
4. 消费者应用和商店标题统一使用 `Talenro VPN`，安装后短名称使用 `Talenro`。
5. Basic、Plus、Pro、Enterprise 是商业套餐层级，不是独立应用或独立品牌。
6. 拉丁字母主标识在五种首发语言中保持不变；其他文字转写只作本地化辅助。
7. 品牌描述语使用 `Private Connectivity`；传播标语使用 `Private. Reliable. Connected.` 和“私密连接，始终可靠。”
8. 先取得核心域名和账号，再公开品牌；本规格不授权购买域名或创建外部账号。
9. 商标检索必须覆盖相同和近似的读音、外观、含义及相关商品或服务，不能只做精确字符串搜索。
10. 正式权属确认前不冻结不可轻易更改的应用包标识符，也不使用 `®`。

## 3. 品牌架构

### 3.1 主品牌

规范写法只有：

```text
Talenro
```

规则：

- 大写首字母 `T`，其余字母小写。
- 英文建议读音：`tuh-LEN-roh`，重音在第二音节。
- Logo、应用图标内字标、网站导航和产品启动页以 `Talenro` 为主。
- 首次在简体中文长文中出现时写作 `Talenro 泰联诺`；后文按版面选用 `Talenro` 或 `泰联诺`，同一页面保持一致。
- `TALENRO` 仅可在确有需要的等宽技术字段或法律表格中使用，不作为常规展示字标。

禁止的展示变体包括 `TalenRo`、`Talen RO`、`TalenroVPN`、`Talenro-VPN` 和将品牌拆成两个词。域名、社交账号和技术标识符可按平台规则使用全小写 `talenro`。

### 3.2 多语言名称

| 语言 | 主展示 | 本地转写或名称 | 使用规则 |
|---|---|---|---|
| 英文 | `Talenro` | — | 全球源名称 |
| 简体中文 | `Talenro` | `泰联诺` | 官方本地化名称，可用于中文正文与市场材料 |
| 俄语 | `Talenro` | `Таленро` | 只作首次发音提示、正文或搜索辅助 |
| 波斯语 | `Talenro` | `تالنرو` | 只作 RTL 正文中的发音提示或搜索辅助 |
| 日语 | `Talenro` | `タレンロ` | 只作首次发音提示、正文或搜索辅助 |

俄语、波斯语和日语转写不替代 `Talenro`，也不单独注册应用、域名、社交账号或套餐。若母语审校发现转写读音存在明显偏差，可修正转写；主品牌拼写不得随之改变。

### 3.3 产品与套餐命名

| 场景 | 规范名称 |
|---|---|
| Apple App Store / Google Play 标题 | `Talenro VPN` |
| 官网下载页产品名 | `Talenro VPN` |
| Windows / macOS 安装包产品名 | `Talenro VPN` |
| 安装后的短名称 | `Talenro` |
| 公司或产品总品牌 | `Talenro` |
| 套餐 | `Talenro Basic`、`Talenro Plus`、`Talenro Pro`、`Talenro Enterprise` |

应用商店可建立五种语言的本地化商品页，但标题默认均保留 `Talenro VPN`，不翻译 `Talenro`，也不在标题中加入价格、排名、促销、节点数量或绝对安全承诺。

未来产品只有在出现真实的独立用户群、独立价值主张和独立发布生命周期时才增加后缀名称。本阶段不预先创建 `Talenro Cloud`、`Talenro Business` 等未验证子品牌。

## 4. 描述语、标语与文案边界

### 4.1 规范文案

- 英文品牌描述语：`Private Connectivity`
- 英文传播标语：`Private. Reliable. Connected.`
- 简体中文传播标语：`私密连接，始终可靠。`

描述语用于解释产品类别，可出现在网站标题、品牌手册和企业介绍中。传播标语用于启动页、网站首屏和营销材料。应用商店副标题需结合各地区政策和本地化长度单独审核，不因本规格自动发布。

俄语、波斯语和日语标语必须由母语人员从英文源文案做创译与回译审核，本规格不以机器直译冻结最终文本。

### 4.2 禁止承诺

官方标题、标语和关键销售文案不得使用以下未经证实或容易误导的表述：

- “100% secure”“绝对匿名”“永不掉线”；
- “解锁一切”“绕过所有限制”；
- “永久免费”或与无永久免费套餐相冲突的表述；
- “No. 1”“Best VPN”等无可验证依据的排名；
- 把 `Talenro` 描述为某个自然语言单词的真实派生词。

产品可以准确说明功能、适用网络条件和测得的恢复指标，但必须保留测试口径与适用条件。

## 5. 商标和法律核验

### 5.1 正式检索范围

公开发布前由商标律师或合格服务商在实际销售、运营和用户触达地区完成清查，至少覆盖：

- `Talenro`、`泰联诺` 及读音、拼写、外观或含义近似的标识；
- `Таленро`、`تالنرو`、`タレンロ` 的相关文字检索；
- 官方商标数据库、公司名称、应用商店、域名、社交账号、软件仓库和普通法使用；
- 与可下载软件、网络通信/VPN 服务、SaaS 和安全技术相关的商品与服务。

尼斯分类第 9、38、42 类可作为初步讨论入口，最终申请类别和商品/服务描述由专业法律意见按实际经营模式确定。分类相同不自动构成冲突，分类不同也不自动排除混淆风险。

### 5.2 标记规则

- 未取得注册前禁止使用 `Talenro®` 或 `泰联诺®`。
- `™` 是否使用由目标地区法律顾问决定；默认产品 UI 和 Logo 不附加符号。
- 注册完成后只在注册有效的国家、文字标识和商品/服务范围内使用 `®`。
- 商标主体、域名主体、开发者账号主体和收款主体之间的权属关系必须留存书面记录。

## 6. 域名、账号与技术标识符

### 6.1 域名和账号优先级

在品牌公开前按以下顺序实时核验并取得可用资产：

1. `talenro.com`
2. `talenro.app`
3. `talenro.net`
4. `gettalenro.com`
5. 主要应用商店开发者名、社交平台、代码托管与支持渠道的 `talenro` 或 `@talenro`

首轮 DNS 查询未发现上述四个域名的有效 DNS 记录，但 DNS 不存在不代表域名尚未注册或可以购买。购买前必须在注册商或 RDAP 中实时确认注册状态、价格、历史与争议风险。

只选一个域名作为品牌主站，其余资产做安全重定向或防御性保留，不建设多个互相竞争的官网品牌。不得在未获用户明确授权时购买、竞价或联系域名持有人。

### 6.2 应用与代码标识符

在核心域名和商标路径确认后，应用标识符可采用以下命名空间：

```text
com.talenro.*
```

正式发布前再冻结 Android application ID、Apple bundle ID、Windows package identity、macOS bundle ID、签名证书主体和 API audience。不得因为本设计中的暂定字符串提前占用错误的不可变标识符。

内部服务、仓库和监控标签使用小写 ASCII `talenro`；显示给用户的产品名仍遵循 `Talenro` 的规范大小写。

## 7. 应用商店元数据约束

`Talenro VPN` 为 11 个 ASCII 字符，满足 Apple 和 Google Play 当前 30 字符的应用名称上限。上架时仍须执行平台当期规则复核。

每种语言的商品页遵循以下约束：

- 品牌拼写保持 `Talenro`，不把本地转写替换进标题；
- 标题、图标、截图和描述与真实功能一致；
- 标题不加入 emoji、重复特殊字符、价格、折扣、排名或促销语；
- 副标题和短描述在各自平台限制内由本地化流程维护；
- 商店关键词和描述不得制造无法证明的匿名性、安全性或解锁能力承诺；
- 官网 APK 与 Google Play 版保持同一品牌和产品名称，渠道差异只体现在签名、分发与功能合规范围。

## 8. 本地化与 RTL

- `Talenro` 是不可翻译术语，加入共享术语表并锁定大小写。
- `VPN` 作为通用技术缩写保留拉丁字母。
- 波斯语 RTL 页面中的 `Talenro`、域名、版本号和技术标识符作为 LTR 隔离片段处理，避免双向文本重排。
- 中文第一次正式提及时使用 `Talenro 泰联诺`；俄语、波斯语和日语第一次需要解释发音时可在括号中给出转写。
- Logo 资产只含 `Talenro` 主字标；本地化名称作为可访问文本或正文，不为五种语言制造五套 Logo。
- 搜索关键词可包含合规的本地转写，但不得使用户误以为存在不同发行方或产品。

## 9. 发布门与验收标准

品牌公开、商店建档和安装包签名前必须全部满足：

1. 目标市场完成正式商标清查并形成书面结论。
2. 商标申请主体、开发者主体、收款主体和域名权属方案已确认。
3. 核心域名及必要账号已实时核验并取得控制权。
4. 五种语言的主品牌拼写、中文名和转写通过母语审校。
5. `Talenro VPN` 在各应用商店完成相似名称与元数据政策复核。
6. 包标识符只在上述检查通过后冻结。
7. 所有公开文案通过禁止承诺、隐私、安全与本地化审查。
8. Logo 和字标在 LTR、RTL、小尺寸应用图标、深浅色背景和辅助功能场景中均可识别。
9. 发布资产中不存在 `TalenRo`、`TALENRO`、`TalenroVPN` 等非规范变体。
10. `®` 使用范围与实际注册状态一致；未注册时不显示。

## 10. 初步排查结论与局限

在候选阶段，`Avenro`、`Aveniq`、`Velanor`、`Senvora`、`Orvani`、`Averune`、`Elarivo`、`Selunor`、`Orelune` 等名称因发现活跃企业、产品、服务或商标冲突信号而淘汰。

对 `Talenro`、`泰联诺` 及三种转写所做的首轮公开网络检索未发现明显的同名商业产品；候选域名的首轮 DNS 查询也未返回有效记录。这些结果只能降低早期命名风险，不能证明商标可注册、名称可合法使用或域名可购买。最终发布决策以正式检索和实时权属核验为准。

## 11. 官方参考

- [USPTO：Likelihood of confusion](https://www.uspto.gov/trademarks/search/likelihood-confusion)
- [USPTO：Why search for similar trademarks?](https://www.uspto.gov/trademarks/basics/why-search-similar-trademarks)
- [USPTO：Federal trademark searching](https://www.uspto.gov/trademarks/search/federal-trademark-searching)
- [Apple：App Store product page](https://developer.apple.com/app-store/product-page/)
- [Apple：App information](https://developer.apple.com/help/app-store-connect/reference/app-information/app-information)
- [Google Play：Create and set up your app](https://support.google.com/googleplay/android-developer/answer/9859152?hl=en)
- [Google Play：Metadata policy](https://support.google.com/googleplay/android-developer/answer/9898842?hl=en)

