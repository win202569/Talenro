# 2026-09-28 Task 3 real verification and review checkpoint

**Historical checkpoint; directory diagnosis corrected:** after approved cache
preparation, all 147 physical directories exist, but readonly Go metadata still
omits their `Dir` because full-module checksums are absent from the tools go.sum.
The original inference that empty `Dir` proves a missing directory was too strong.
The consumer-binding repair, complete regression pass and remaining verifier
blocker are recorded in the [latest follow-up](2026-09-28-devtools-cache-binding-fix.md).

Status: **not accepted; no push or merge**. Implementation commit:
`a0ca18928f552752c6580929f9b6fff88cd8d58b`, branch
`codex/c12-b01-task9-coordinator`. Commands ran from that worktree root with
fixed Go 1.26.5 and stable PowerShell Core 7.6.5. No dependency download,
security-setting change, full ordinary/C11/C12 run or physical gate was performed.

## Actual results

| Phase | UTC start → finish | Limit | Exit / result |
| --- | --- | --- | --- |
| Root `go mod verify` | 12:36:03.4255199 → 12:50:39.3022453 | 900s | 0 / 875.905s; `all modules verified` |
| Public `scripts/verify-devtools.ps1` | 12:50:54.2443840 → 12:51:01.5290778 | 900s | 1 / 7.296s; `verify-devtools: dependencies failed with exit code 1.` |
| Private offline dependency query | 12:51:41.7747031 → 12:51:41.9049993 | 90s | 0 / 0.132s; module directories missing |
| Two-package compile-only | 12:51:56.5306810 → 12:52:56.2925204 | 300s | 0 / 59.767s; both packages report `[no tests to run]` |

The private diagnostic repeated only the underlying read-only metadata query,
not the failed public verification. It used the tools directory as its actual
working directory and the same fixed GOPATH/GOMODCACHE and offline settings:

```text
go list -m -f '{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}' all
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run ^$ -count=1 -timeout=5m
```

The second command is the separate compile-only phase, run from the root.
The diagnostic wrapper's initial Cwd field names its caller root; the actual
dependency process directory is `tools/devtools`. No application test ran.

Root integrity verifies downloaded content only: it does not prove a complete
cache. Offline tool metadata returned 147 records without extracted directories.
Read-only file inspection found 120 with existing zip/ziphash but no extracted
directory, and 27 missing all three. File existence is not integrity verification.
The verifier has not reached its tool-module integrity phase.

Tool versions, generation, generated-output comparison and real root lint were
**not run** after this failure. No identical verification retry or extended budget.
Full ordinary/C11/C12, 13 physical gates and I2/I3 remain blocked by the prior
security event. Historical readiness failures remain unexplained; later focused
passes do not prove cold-state stability.

## One independent read-only review

Reviewed plan range `f969c19af43bb917f1ca4c221893e0274e8dfa36..a0ca18928f552752c6580929f9b6fff88cd8d58b`
and current recovery documentation. No reviewer tests, writes or subagents.
Verdict: **not ready to merge**; no Critical findings.

1. **Important / open:** consumers verify a fixed cache in a child, then can run
   PATH-resolved Go with inherited cache/proxy settings. Verification and execution
   are not bound to the same cache. This needs a security-sensitive consumer
   environment contract repair with a differing-cache negative test and RED/GREEN.
   No production fix was made in this checkpoint; request explicit approval for
   fixed toolchain/cache/offline execution including nested protoc, without changing
   smoke/C12 hosts, root lint targeting or time limits.
2. **Important / documentation corrected, fresh-machine execution unverified:**
   installed Go plus module downloads did not ensure the verifier's exact
   toolchain-cache path. Recovery instructions now name that separately authorized
   prerequisite, pin preparation to the same cache and check path/version. No
   toolchain installation or download was performed here.
3. **Minor / deferred:** plan-top Revision status still describes the older Task 2
   stop point. Current results are recorded in this checkpoint and Task 3 evidence;
   that historical summary has not been polished during the single fix pass.

Reviewer exclusions are retained rather than silently treated as passes:

- Defender classification, I2/I3 and physical C12 behavior remain separately gated.
- Third-party vulnerability equivalence and untested platform/configuration behavior
  are not established by static package absence.
- Review did not independently replay every checksum origin and all 1,330 membership
  judgments; the checked evidence machinery is not a provenance re-audit.
- Real Buf/plugin and lint behavior remain unverified because real tool acceptance
  stopped at dependency presence.
- Runtime timing/cleanup beyond recorded tests were not independently replayed;
  no cold/load stability claim.

Task 3 remains incomplete. Preserve the ignored local execution ledger/logs; do not
run task-done or delete them while acceptance and the Important finding remain open.
Any expanded cache preparation requires separate approval of the exact list below.

## Exact missing-cache inventory

Snapshot after the failed public verifier; this is a preparation request, not
authorization. Versions are the currently selected tools-module versions, not
new upgrades. `directory` means absent extraction; `zip` and `ziphash` refer to
the module cache archive and its hash sidecar.

| Module | Version | Missing |
| --- | --- | --- |
| buf.build/gen/go/bufbuild/bufplugin/connectrpc/go | v1.20.0-20250718181942-e35f9b667443.1 | directory |
| buf.build/gen/go/bufbuild/protovalidate/connectrpc/go | v1.20.0-20250717185734-6c6e0d3c608e.1 | directory |
| buf.build/go/hyperpb | v0.1.3 | directory |
| cloud.google.com/go | v0.121.2 | directory |
| cloud.google.com/go/auth | v0.16.5 | directory |
| cloud.google.com/go/compute | v1.6.1 | directory |
| cloud.google.com/go/compute/metadata | v0.9.0 | directory,zip,ziphash |
| cloud.google.com/go/firestore | v1.6.1 | directory |
| github.com/Azure/go-ansiterm | v0.0.0-20250102033503-faa5f7b0171c | directory |
| github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp | v1.31.0 | directory,zip,ziphash |
| github.com/alecthomas/kingpin/v2 | v2.4.0 | directory |
| github.com/alecthomas/template | v0.0.0-20190718012654-fb15b899a751 | directory,zip,ziphash |
| github.com/alecthomas/units | v0.0.0-20240927000941-0f3dac36c52b | directory |
| github.com/anthropics/anthropic-sdk-go | v1.38.0 | directory |
| github.com/armon/go-metrics | v0.3.10 | directory |
| github.com/aymanbagabas/go-osc52/v2 | v2.0.1 | directory |
| github.com/aymanbagabas/go-udiff | v0.4.1 | directory |
| github.com/bahlo/generic-list-go | v0.2.0 | directory |
| github.com/benbjohnson/clock | v1.1.0 | directory |
| github.com/bits-and-blooms/bitset | v1.24.4 | directory |
| github.com/blackwell-systems/gcf-go | v1.2.2 | directory |
| github.com/buger/jsonparser | v1.1.2 | directory |
| github.com/charmbracelet/bubbles | v0.21.0 | directory |
| github.com/charmbracelet/bubbletea | v1.3.10 | directory |
| github.com/charmbracelet/lipgloss | v1.1.0 | directory |
| github.com/charmbracelet/x/cellbuf | v0.0.13-0.20250311204145-2c3ea96c31dd | directory |
| github.com/charmbracelet/x/exp/golden | v0.0.0-20250806222409-83e3a29d542f | directory |
| github.com/chzyer/logex | v1.1.10 | directory,zip,ziphash |
| github.com/chzyer/readline | v0.0.0-20180603132655-2972be24d48e | directory,zip,ziphash |
| github.com/chzyer/test | v0.0.0-20180213035817-a1ea475d72b1 | directory,zip,ziphash |
| github.com/clipperhouse/stringish | v0.1.1 | directory |
| github.com/cncf/xds/go | v0.0.0-20251210132809-ee656c7534f5 | directory |
| github.com/containerd/typeurl/v2 | v2.2.0 | directory |
| github.com/coreos/go-systemd/v22 | v22.3.2 | directory,zip,ziphash |
| github.com/creack/pty | v1.1.24 | directory,zip,ziphash |
| github.com/cristalhq/acmd | v0.12.0 | directory |
| github.com/danieljoos/wincred | v1.2.3 | directory |
| github.com/dchest/siphash | v1.2.3 | directory |
| github.com/ebitengine/purego | v0.10.0 | directory |
| github.com/envoyproxy/go-control-plane | v0.14.0 | directory,zip,ziphash |
| github.com/envoyproxy/go-control-plane/envoy | v1.36.0 | directory |
| github.com/envoyproxy/go-control-plane/ratelimit | v0.1.0 | directory,zip,ziphash |
| github.com/envoyproxy/protoc-gen-validate | v1.3.0 | directory |
| github.com/erikgeiser/coninput | v0.0.0-20211004153227-1c3628e74d0f | directory |
| github.com/expr-lang/expr | v1.17.7 | directory |
| github.com/go-jose/go-jose/v4 | v4.1.3 | directory |
| github.com/go-ole/go-ole | v1.2.6 | directory |
| github.com/gogo/protobuf | v1.3.2 | directory,zip,ziphash |
| github.com/golang/glog | v1.2.5 | directory,zip,ziphash |
| github.com/golang/groupcache | v0.0.0-20210331224755-41bb18bfe9da | directory,zip,ziphash |
| github.com/google/s2a-go | v0.1.9 | directory,zip,ziphash |
| github.com/googleapis/enterprise-certificate-proxy | v0.3.6 | directory |
| github.com/googleapis/gax-go/v2 | v2.15.0 | directory |
| github.com/gookit/color | v1.6.0 | directory |
| github.com/gorilla/mux | v1.8.0 | directory |
| github.com/gorilla/websocket | v1.5.3 | directory |
| github.com/hashicorp/consul/api | v1.12.0 | directory |
| github.com/hashicorp/go-cleanhttp | v0.5.2 | directory,zip,ziphash |
| github.com/hashicorp/go-hclog | v1.2.0 | directory |
| github.com/hashicorp/go-immutable-radix | v1.3.1 | directory,zip,ziphash |
| github.com/hashicorp/go-rootcerts | v1.0.2 | directory,zip,ziphash |
| github.com/hashicorp/golang-lru | v0.5.4 | directory,zip,ziphash |
| github.com/hashicorp/serf | v0.9.7 | directory |
| github.com/hpcloud/tail | v1.0.0 | directory,zip,ziphash |
| github.com/ianlancetaylor/demangle | v0.0.0-20200824232613-28f6c0f3b639 | directory |
| github.com/invopop/jsonschema | v0.13.0 | directory |
| github.com/jmoiron/sqlx | v1.3.5 | directory |
| github.com/jordanlewis/gcassert | v0.0.0-20250430164644-389ef753e22e | directory |
| github.com/jpillora/backoff | v1.0.0 | directory |
| github.com/json-iterator/go | v1.1.12 | directory |
| github.com/julienschmidt/httprouter | v1.3.0 | directory |
| github.com/keybase/go-keychain | v0.0.1 | directory |
| github.com/kr/pty | v1.1.1 | directory,zip,ziphash |
| github.com/kylelemons/godebug | v1.1.0 | directory |
| github.com/lib/pq | v1.12.3 | directory |
| github.com/lufia/plan9stats | v0.0.0-20211012122336-39d0f177ccd0 | directory,zip,ziphash |
| github.com/magefile/mage | v1.14.0 | directory |
| github.com/mailru/easyjson | v0.7.7 | directory |
| github.com/mattn/go-localereader | v0.0.1 | directory |
| github.com/matttproud/golang_protobuf_extensions | v1.0.1 | directory |
| github.com/mgechev/dots | v1.0.0 | directory |
| github.com/moby/term | v0.5.2 | directory |
| github.com/modern-go/concurrent | v0.0.0-20180306012644-bacd9c7ef1dd | directory |
| github.com/modern-go/reflect2 | v1.0.2 | directory |
| github.com/mozilla/tls-observatory | v0.0.0-20250923143331-eef96233227e | directory |
| github.com/muesli/ansi | v0.0.0-20230316100256-276c6243b2f6 | directory |
| github.com/muesli/termenv | v0.16.0 | directory |
| github.com/mwitkow/go-conntrack | v0.0.0-20190716064945-2f068394615f | directory |
| github.com/ncruces/sort | v0.1.6 | directory |
| github.com/ncruces/wbt | v1.0.0 | directory |
| github.com/openai/openai-go/v3 | v3.32.0 | directory |
| github.com/otiai10/curr | v1.0.0 | directory |
| github.com/otiai10/mint | v1.3.1 | directory |
| github.com/phayes/checkstyle | v0.0.0-20170904204023-bfd46e6a821d | directory |
| github.com/pkg/errors | v0.9.1 | directory |
| github.com/planetscale/vtprotobuf | v0.6.1-0.20240319094008-0393e58bdf10 | directory,zip,ziphash |
| github.com/power-devops/perfstat | v0.0.0-20240221224432-82ca36839d55 | directory |
| github.com/psanford/httpreadat | v0.1.0 | directory |
| github.com/quasilyte/go-ruleguard/rules | v0.0.0-20211022131956-028d6511ab71 | directory |
| github.com/remyoudompheng/bigfft | v0.0.0-20230129092748-24d4a6f8daec | directory |
| github.com/russross/blackfriday | v1.6.0 | directory |
| github.com/sagikazarmark/crypt | v0.6.0 | directory |
| github.com/santhosh-tekuri/jsonschema/v5 | v5.3.1 | directory |
| github.com/shirou/gopsutil/v4 | v4.26.4 | directory |
| github.com/spiffe/go-spiffe/v2 | v2.6.0 | directory |
| github.com/stoewer/go-strcase | v1.3.0 | directory |
| github.com/tidwall/gjson | v1.18.0 | directory |
| github.com/tidwall/match | v1.1.1 | directory |
| github.com/tidwall/pretty | v1.2.1 | directory |
| github.com/tidwall/sjson | v1.2.5 | directory |
| github.com/timandy/routine | v1.1.6 | directory |
| github.com/tklauser/go-sysconf | v0.3.16 | directory |
| github.com/tklauser/numcpus | v0.11.0 | directory |
| github.com/valyala/bytebufferpool | v1.0.0 | directory |
| github.com/valyala/quicktemplate | v1.8.0 | directory |
| github.com/wk8/go-ordered-map/v2 | v2.1.8 | directory |
| github.com/xeipuuv/gojsonpointer | v0.0.0-20180127040702-4e3ac2762d5f | directory |
| github.com/xeipuuv/gojsonreference | v0.0.0-20180127040603-bd5ef7bd5415 | directory |
| github.com/xeipuuv/gojsonschema | v1.2.0 | directory |
| github.com/xhit/go-str2duration/v2 | v2.1.0 | directory |
| github.com/yuin/goldmark | v1.4.13 | directory,zip,ziphash |
| github.com/yusufpapurcu/wmi | v1.2.4 | directory |
| go.etcd.io/etcd/api/v3 | v3.5.4 | directory |
| go.etcd.io/etcd/client/pkg/v3 | v3.5.4 | directory |
| go.etcd.io/etcd/client/v2 | v2.305.4 | directory |
| go.etcd.io/etcd/client/v3 | v3.5.4 | directory |
| go.opencensus.io | v0.23.0 | directory,zip,ziphash |
| go.opentelemetry.io/contrib/detectors/gcp | v1.39.0 | directory |
| golang.org/x/lint | v0.0.0-20190930215403-16217165b5de | directory |
| golang.org/x/oauth2 | v0.36.0 | directory,zip,ziphash |
| golang.org/x/telemetry | v0.0.0-20260708182218-49f421fb7959 | directory |
| golang.org/x/time | v0.11.0 | directory |
| golang.org/x/xerrors | v0.0.0-20220517211312-f3a8303e98df | directory |
| google.golang.org/api | v0.81.0 | directory |
| google.golang.org/appengine | v1.6.7 | directory |
| google.golang.org/genai | v1.54.0 | directory |
| google.golang.org/genproto | v0.0.0-20220519153652-3a47de7e79bd | directory |
| gopkg.in/alecthomas/kingpin.v2 | v2.2.6 | directory,zip,ziphash |
| gopkg.in/fsnotify.v1 | v1.4.7 | directory,zip,ziphash |
| gopkg.in/src-d/go-billy.v4 | v4.3.2 | directory |
| lukechampine.com/adiantum | v1.1.1 | directory |
| modernc.org/golex | v1.1.0 | directory |
| modernc.org/mathutil | v1.6.0 | directory |
| modernc.org/parser | v1.1.0 | directory |
| modernc.org/sortutil | v1.2.0 | directory |
| modernc.org/strutil | v1.2.0 | directory |
| modernc.org/y | v1.1.0 | directory |
