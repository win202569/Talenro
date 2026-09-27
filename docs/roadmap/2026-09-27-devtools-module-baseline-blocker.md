# 开发工具拆分：离线版本基线阻塞

日期：2026-09-27。状态：任务 1 专项完成；任务 2 停在基线准备，尚未改模块或消费者。

## 已完成与证据

旧 Unix 实现与失败测试归档提交：`03947e31536a803aaca19c295edd8d9eb4d258f1`。
Windows-only 独立验证入口与测试提交：`3a43a724de801bf24abdede71d806559016cb602`。
提交后任务 1 收尾复验退出 0 / 81.078 秒：PowerShell 29 项通过、Windows Bash
20 项通过、0 失败、1 项不适用 Skip。Linux 拒绝专项退出 0 / 0.153 秒，10 项通过、无 Skip。
这些不是全套回归、C11、C12 物理验收或最终独立审查结果。

## 任务 2 实际阻塞

固定 Go 1.26.5，GOENV/GOWORK 关闭、GOTOOLCHAIN=local、GOPROXY/GOSUMDB=off，
GOFLAGS=-mod=readonly，每条诊断有 90 秒上限；不执行被测产品或测试函数。
`go mod graph` 退出 0 / 0.132 秒。
`go list -m -json all` 退出 1 / 1.870 秒，报告 module lookup disabled by GOPROXY=off。
因此完整选中版本基线未取得，普通/integration/Goose/五工具包闭包尚未开始读取。

仅为列出错误模块，随后运行 `go list -m -e` 配合 Error 字段模板；207 项有错误。
该诊断退出 0 不代表基线通过，部分输出不能用于锁版本迁移。下表列出它报告的请求版本，
不是声称完成了整个模块图的版本验收，也不等同于 207 个源码包必然全部缺失。

根 go.mod/go.sum 未修改，归档时工作文件 SHA-256：
- go.mod：`37B8802CD19B0C232290CA4B68A7455CB26A8E1CBA01394306CB440899C3325C`
- go.sum：`B58809AC5E3092CA54C5124F96D76D0D9DC79AD15E547983A9AA99F9F9950A7A`

## 恢复条件

须单独批准联网准备当前锁定版本依赖，或提供等价离线缓存；不自动下载、不升级依赖。
准备结束后重新执行完整基线，核对文件未变，才继续原子模块/消费者迁移。
任务 3 未开始；Defender/I2/I3 及完整普通测试、C11、C12 的原有阻塞不变。
未推送或合并 main。

## 离线查询失败的 207 项

```text
buf.build/gen/go/bufbuild/bufplugin/connectrpc/go@v1.20.0-20250718181942-e35f9b667443.1
buf.build/gen/go/bufbuild/protovalidate/connectrpc/go@v1.20.0-20250717185734-6c6e0d3c608e.1
buf.build/go/hyperpb@v0.1.3
cloud.google.com/go@v0.121.2
cloud.google.com/go/auth@v0.16.5
cloud.google.com/go/compute@v1.6.1
cloud.google.com/go/firestore@v1.6.1
connectrpc.com/grpcreflect@v1.3.0
dario.cat/mergo@v1.0.2
github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.21.0
github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.13.1
github.com/Azure/azure-sdk-for-go/sdk/internal@v1.11.2
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys@v1.4.0
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal@v1.2.0
github.com/Azure/go-ansiterm@v0.0.0-20250102033503-faa5f7b0171c
github.com/AzureAD/microsoft-authentication-library-for-go@v1.6.0
github.com/CloudyKit/fastprinter@v0.0.0-20200109182630-33d98a066a53
github.com/CloudyKit/jet/v6@v6.2.0
github.com/Joker/jade@v1.1.3
github.com/RaveNoX/go-jsoncommentstrip@v1.0.0
github.com/Shopify/goreferrer@v0.0.0-20220729165902-8cddb4f5de06
github.com/alecthomas/assert/v2@v2.11.0
github.com/alecthomas/kingpin/v2@v2.4.0
github.com/alecthomas/repr@v0.5.2
github.com/alecthomas/units@v0.0.0-20240927000941-0f3dac36c52b
github.com/anthropics/anthropic-sdk-go@v1.38.0
github.com/antihax/optional@v1.0.0
github.com/armon/go-metrics@v0.3.10
github.com/aymanbagabas/go-osc52/v2@v2.0.1
github.com/aymanbagabas/go-udiff@v0.4.1
github.com/aymerick/douceur@v0.2.0
github.com/benbjohnson/clock@v1.1.0
github.com/bits-and-blooms/bitset@v1.24.4
github.com/blackwell-systems/gcf-go@v1.2.2
github.com/bmatcuk/doublestar@v1.1.1
github.com/bmatcuk/doublestar/v4@v4.10.0
github.com/brianvoe/gofakeit/v6@v6.28.0
github.com/buger/jsonparser@v1.1.2
github.com/census-instrumentation/opencensus-proto@v0.2.1
github.com/charmbracelet/bubbles@v0.21.0
github.com/charmbracelet/bubbletea@v1.3.10
github.com/charmbracelet/lipgloss@v1.1.0
github.com/charmbracelet/x/cellbuf@v0.0.13-0.20250311204145-2c3ea96c31dd
github.com/charmbracelet/x/exp/golden@v0.0.0-20250806222409-83e3a29d542f
github.com/clipperhouse/stringish@v0.1.1
github.com/cncf/udpa/go@v0.0.0-20210930031921-04548b0d99d4
github.com/cncf/xds/go@v0.0.0-20251210132809-ee656c7534f5
github.com/containerd/log@v0.1.0
github.com/containerd/platforms@v0.2.1
github.com/containerd/typeurl/v2@v2.2.0
github.com/cpuguy83/dockercfg@v0.3.2
github.com/cristalhq/acmd@v0.12.0
github.com/danieljoos/wincred@v1.2.3
github.com/dave/jennifer@v1.7.1
github.com/dchest/siphash@v1.2.3
github.com/dmarkham/enumer@v1.6.3
github.com/docker/docker@v28.5.2+incompatible
github.com/ebitengine/purego@v0.10.0
github.com/envoyproxy/go-control-plane/envoy@v1.36.0
github.com/envoyproxy/protoc-gen-validate@v1.3.0
github.com/erikgeiser/coninput@v0.0.0-20211004153227-1c3628e74d0f
github.com/expr-lang/expr@v1.17.7
github.com/fatih/structs@v1.1.0
github.com/flosch/pongo2/v4@v4.0.2
github.com/frankban/quicktest@v1.14.3
github.com/ghodss/yaml@v1.0.0
github.com/go-jose/go-jose/v4@v4.1.3
github.com/go-openapi/testify/v2@v2.4.2
github.com/go-quicktest/qt@v1.101.0
github.com/go-task/slim-sprig@v0.0.0-20210107165309-348f09dbbbc0
github.com/go-task/slim-sprig/v3@v3.0.0
github.com/go-toolsmith/pkgload@v1.2.2
github.com/golang/mock@v1.1.1
github.com/gomarkdown/markdown@v0.0.0-20240328165702-4d01890c35c0
github.com/google/go-tpm-tools@v0.3.13-0.20230620182252-4639ecce2aba
github.com/google/pprof@v0.0.0-20260115054156-294ebfa9ad83
github.com/googleapis/enterprise-certificate-proxy@v0.3.6
github.com/googleapis/gax-go/v2@v2.15.0
github.com/gookit/color@v1.6.0
github.com/gorilla/css@v1.0.0
github.com/gostaticanalysis/testutil@v0.5.0
github.com/grpc-ecosystem/grpc-gateway@v1.16.0
github.com/hashicorp/consul/api@v1.12.0
github.com/hashicorp/go-hclog@v1.2.0
github.com/hashicorp/serf@v0.9.7
github.com/ianlancetaylor/demangle@v0.0.0-20200824232613-28f6c0f3b639
github.com/iris-contrib/schema@v0.0.6
github.com/jhump/protoreflect/v2@v2.0.0-beta.2
github.com/jmoiron/sqlx@v1.3.5
github.com/joeshaw/multierror@v0.0.0-20140124173710-69b34d4ec901
github.com/jordanlewis/gcassert@v0.0.0-20250430164644-389ef753e22e
github.com/jpillora/backoff@v1.0.0
github.com/juju/gnuflag@v0.0.0-20171113085948-2ce1bb71843d
github.com/julienschmidt/httprouter@v1.3.0
github.com/kataras/blocks@v0.0.8
github.com/kataras/golog@v0.1.11
github.com/kataras/iris/v12@v12.2.11
github.com/kataras/pio@v0.0.13
github.com/kataras/sitemap@v0.0.6
github.com/kataras/tunnel@v0.0.4
github.com/keybase/go-keychain@v0.0.1
github.com/labstack/echo/v4@v4.15.1
github.com/labstack/gommon@v0.4.2
github.com/lib/pq@v1.12.3
github.com/magefile/mage@v1.14.0
github.com/mailgun/raymond/v2@v2.0.48
github.com/mailru/easyjson@v0.7.7
github.com/matryer/is@v1.4.0
github.com/mattn/go-localereader@v0.0.1
github.com/mgechev/dots@v1.0.0
github.com/microcosm-cc/bluemonday@v1.0.26
github.com/mkevac/debugcharts@v0.0.0-20191222103121-ae1c48aa8615
github.com/moby/go-archive@v0.1.0
github.com/moby/patternmatcher@v0.6.0
github.com/moby/sys/sequential@v0.6.0
github.com/moby/sys/user@v0.4.0
github.com/moby/sys/userns@v0.1.0
github.com/moby/term@v0.5.2
github.com/morikuni/aec@v1.0.0
github.com/mozilla/tls-observatory@v0.0.0-20250923143331-eef96233227e
github.com/muesli/ansi@v0.0.0-20230316100256-276c6243b2f6
github.com/muesli/termenv@v0.16.0
github.com/mwitkow/go-conntrack@v0.0.0-20190716064945-2f068394615f
github.com/ncruces/sort@v0.1.6
github.com/ncruces/wbt@v1.0.0
github.com/oapi-codegen/nullable@v1.1.0
github.com/onsi/ginkgo@v1.16.4
github.com/onsi/ginkgo/v2@v2.28.2
github.com/openai/openai-go/v3@v3.32.0
github.com/otiai10/copy@v1.14.0
github.com/otiai10/curr@v1.0.0
github.com/otiai10/mint@v1.3.1
github.com/pascaldekloe/name@v1.0.1
github.com/paulmach/protoscan@v0.2.1
github.com/phayes/checkstyle@v0.0.0-20170904204023-bfd46e6a821d
github.com/protocolbuffers/protoscope@v0.0.0-20221109213918-8e7a6aafa2c9
github.com/psanford/httpreadat@v0.1.0
github.com/quasilyte/go-ruleguard/rules@v0.0.0-20211022131956-028d6511ab71
github.com/quic-go/go-ossfuzz-seeds@v0.1.0
github.com/rekby/fixenv@v0.6.1
github.com/rodaine/protogofakeit@v0.1.1
github.com/rogpeppe/fastuuid@v1.2.0
github.com/russross/blackfriday@v1.6.0
github.com/sagikazarmark/crypt@v0.6.0
github.com/santhosh-tekuri/jsonschema/v5@v5.3.1
github.com/schollz/closestmatch@v2.1.0+incompatible
github.com/sergi/go-diff@v1.2.0
github.com/shirou/gopsutil/v4@v4.26.4
github.com/spiffe/go-spiffe/v2@v2.6.0
github.com/spkg/bom@v0.0.0-20160624110644-59b7046e48ad
github.com/stoewer/go-strcase@v1.3.0
github.com/tdewolff/minify/v2@v2.20.19
github.com/tdewolff/parse/v2@v2.7.12
github.com/tenntenn/modver@v1.0.1
github.com/tenntenn/text/transform@v0.0.0-20200319021203-7eef512accb3
github.com/testcontainers/testcontainers-go@v0.40.0
github.com/tidwall/gjson@v1.18.0
github.com/tidwall/match@v1.1.1
github.com/tidwall/pretty@v1.2.1
github.com/tidwall/sjson@v1.2.5
github.com/timandy/routine@v1.1.6
github.com/tklauser/go-sysconf@v0.3.16
github.com/tklauser/numcpus@v0.11.0
github.com/valyala/bytebufferpool@v1.0.0
github.com/valyala/fasttemplate@v1.2.2
github.com/valyala/quicktemplate@v1.8.0
github.com/vmihailenco/msgpack/v5@v5.4.1
github.com/vmihailenco/tagparser/v2@v2.0.0
github.com/xeipuuv/gojsonpointer@v0.0.0-20180127040702-4e3ac2762d5f
github.com/xeipuuv/gojsonreference@v0.0.0-20180127040603-bd5ef7bd5415
github.com/xeipuuv/gojsonschema@v1.2.0
github.com/xhit/go-str2duration/v2@v2.1.0
github.com/yosssi/ace@v0.0.5
github.com/zeebo/xxh3@v1.1.0
go-simpler.org/assert@v0.9.0
go.etcd.io/etcd/api/v3@v3.5.4
go.etcd.io/etcd/client/pkg/v3@v3.5.4
go.etcd.io/etcd/client/v2@v2.305.4
go.etcd.io/etcd/client/v3@v3.5.4
go.mongodb.org/mongo-driver/v2@v2.5.0
go.opentelemetry.io/contrib/detectors/gcp@v1.39.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace@v1.19.0
go.opentelemetry.io/otel/sdk@v1.44.0
go.opentelemetry.io/otel/sdk/metric@v1.44.0
go.opentelemetry.io/proto/otlp@v1.0.0
golang.org/x/lint@v0.0.0-20190930215403-16217165b5de
golang.org/x/telemetry@v0.0.0-20260708182218-49f421fb7959
golang.org/x/xerrors@v0.0.0-20220517211312-f3a8303e98df
gonum.org/v1/gonum@v0.17.0
google.golang.org/api@v0.81.0
google.golang.org/appengine@v1.6.7
google.golang.org/genai@v1.54.0
google.golang.org/genproto@v0.0.0-20220519153652-3a47de7e79bd
gopkg.in/src-d/go-billy.v4@v4.3.2
gotest.tools/v3@v3.5.2
lukechampine.com/adiantum@v1.1.1
modernc.org/cc/v4@v4.28.1
modernc.org/ccgo/v4@v4.33.0
modernc.org/fileutil@v1.4.0
modernc.org/gc/v2@v2.6.5
modernc.org/gc/v3@v3.1.2
modernc.org/goabi0@v0.2.0
modernc.org/golex@v1.1.0
modernc.org/opt@v0.2.0
modernc.org/parser@v1.1.0
modernc.org/y@v1.1.0
pgregory.net/rapid@v1.2.0
```
