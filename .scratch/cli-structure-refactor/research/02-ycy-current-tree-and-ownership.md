# ycy 当前树与 ownership 事实盘点

## 盘点范围

本记录基于工作区当前 checkout（`9bde380 feat(terminal-discovery): add terminal-discovery adapter for experience presentation`）的源码和构建文件。盘点目标是记录活动实现的路径、职责、直接依赖、测试/fixture 位置、平台文件和结构性风险；本文不定义目标目录，也不包含功能重构建议。

模块路径为 `github.com/hackycy/hackycy-cli`，`go list ./...` 当前列出 30 个活动 Go package：一个二进制入口、一个 CLI 组装包、15 个命令相关 package、7 个共享/基础 package、5 个工具 package 和 1 个 Web package。`legacy/bun`、`mock`、`.scratch`、`web/dist` 和依赖目录不属于活动 Go package 列表。

## 活动根目录

根目录中的活动区域及其当前事实如下：

| 路径 | 当前内容/角色 | 生成或冻结状态 |
| --- | --- | --- |
| `cmd/ycy` | 唯一活动 CLI 二进制入口，package `main` | 源码 |
| `internal/cliapp` | Cobra 根树、全局 flags、命令绑定和 typed handler contracts | 源码 |
| `internal/commands` | 按业务域划分的命令 Module；没有 `internal/commands` 父 package 文件 | 源码 |
| `internal/appconfig` | 用户配置、schema、加密、锁、机器 ID 和配置操作 | 源码 |
| `internal/filesession` | 文件会话状态、替换和权限/进程平台适配 | 源码 |
| `internal/logging` | 诊断配置、记录和敏感信息 redaction | 源码 |
| `internal/terminal` | Session 分类、交互、presentation、tracked operation 和 renderer lease | 源码 |
| `internal/terminaltest` | 终端事实、录制、语义回答、重定向 stream 和 PTY 测试工具 | 测试支持源码 |
| `internal/sevenzipmanifest` | 7-Zip 平台 artifact manifest | 源码；被工具和 7-Zip runtime 使用 |
| `internal/windowsacl` | Windows 私有路径 ACL 实现 | 平台源码 |
| `web` | Vite 源码、Go 嵌入/路由适配及 Web 测试 | 源码；`web/dist` 为生成输出 |
| `tools` | `check-no-bun`、hookctl、7-Zip 准备、release-artifacts、浏览器 harness | 构建/检查工具源码 |
| `scripts` | 安装脚本 `install.sh`/`install.ps1` | 脚本 |
| `legacy/bun` | 旧 Bun 实现和部署文件，供行为兼容查阅 | 冻结只读参考 |
| `mock/nginx-proxy-manager` | 模拟服务源码、schema、前后端和 Cypress 测试 | 测试/模拟应用 |
| `public` | `404.html` | 静态资源 |

## `cmd/ycy`：组合根和进程层

`cmd/ycy` 现有 77 个 Go 文件，其中 35 个生产文件、42 个测试文件，所有文件均为同一个 `main` package。其生产文件是平铺命名，而不是按命令目录分层。

### 入口和组合

- `main.go` 负责进程启动：解析 `os.Args`，分流隐藏升级和 `fs` thumbnail worker，创建终端体验，消费升级状态，调用 `web.Validate`，建立 signal context 和 logging runtime，逐一构造命令 handler，调用 `cliapp.New`/`Execute`，并由 `os.Exit` 结束进程。
- `main.go` 直接导入 `internal/cliapp`、`internal/commands/fs`、`internal/logging`、`internal/terminal` 和 `web`。
- 每个命令的具体 handler 在 `cmd/ycy` 创建 command Module、OS/network/config 依赖和 Terminal Adapter，然后返回 `internal/cliapp` 所需的 typed handler。handler 创建通常发生在闭包第一次运行时，部分长生命周期 Module（例如 `diff`、`fs`、`git heat`）在 handler 构造时创建。

### 命令组合文件

| 当前文件 | 组合对象和直接导入 | 当前事实 |
| --- | --- | --- |
| `configcm.go` | `appconfig`、`cliapp`、`commands/config/cm`、`terminal` | `config cm list/add/use/set/remove` 的 store/Module 构造、Terminal presentation；同文件包含多个 `newConfigCM*Handler` |
| `configcmadd.go` | `commands/config/cm`、`terminal` | CM add 的文本/密码/选择/终端交互 Adapter |
| `configcmremove.go` | `commands/config/cm`、`terminal` | CM remove 的确认 Adapter |
| `configcmtest.go` | `appconfig`、`cliapp`、`commands/config/cm`、`terminal` | CM provider test 的 store、HTTP transport 和 presentation |
| `configfork.go` | `appconfig`、`cliapp`、`commands/config/fork`、`terminal` | `config fork list` 的读取和 presentation，以及 handler 构造 |
| `configforkadd.go` | `commands/config/fork`、`terminal` | Fork add 的文本/选择 Adapter |
| `configforkremove.go` | `commands/config/fork`、`terminal` | Fork remove 的选择/确认 Adapter |
| `exportenv.go` | `cliapp`、`commands/exportenv`、`terminal` | `export env` Module 的 OS 文件 reader/writer、环境目录、选择/presentation Adapter |
| `diff.go` | `cliapp`、`commands/diff`、`terminal` | Diff Module 构造、网络接口注入、startup presentation、operation wait |
| `fs.go` | `cliapp`、`commands/fs`、`terminal` | FS Module 构造、网络接口注入、startup presentation、operation wait |
| `rm.go` | `cliapp`、`commands/rm`、`terminal` | RM Module 的工作目录、删除器、prompt/presentation Adapter |
| `run.go` | `cliapp`、`commands/run`、`terminal` | Run Module 的目录/filesystem 依赖、child runner、Terminal Adapter |
| `zip.go` | `cliapp`、`commands/zip`、`terminal` | Zip Module 的 prompt/presentation、远程名解析和 host reveal OS Adapter |
| `githeat.go` | `cliapp`、`commands/git/heat`、`terminal` | Git Heat Module 和 OS Git runner，报告 presentation |
| `gitpulse.go` | `cliapp`、`commands/git/pulse`、`terminal` | Git Pulse Module、OS filesystem/Git runner、prompt/presentation/tracking Adapter |
| `gitfork.go` | `appconfig`、`cliapp`、`logging`、`commands/git/fork`、`terminal` | Fork config/provider、OS clone runner、prompt/presentation/tracking Adapter |
| `gitcm.go` | `appconfig`、`cliapp`、`logging`、`commands/git/cm`、`terminal` | CM config、Git runner/filesystem、provider HTTP transport、prompt/presentation/tracking Adapter |
| `tunnel.go` | `appconfig`、`cliapp`、`logging`、`commands/tunnel`、`terminal` | Tunnel server/connect handlers、config resolution、logger、client connection selector/presentation |
| `upgrade.go` | `cliapp`、`commands/upgrade`、`logging`、`terminal` | 普通 upgrade handler、隐藏 internal updater、startup transaction consumption |

### 共享进程适配文件

- `git_process.go`、`git_process_unix.go`、`git_process_windows.go` 是 Heat/Pulse/CM/Fork 共同使用的外部 Git 进程生命周期和平台 signal 实现。`git_process.go` 定义 `gitProcessOutput`，Unix/Windows 文件定义 child group signal 行为。
- `git*_process.go` 将共同 Git process 输出映射成各命令 package 的 `GitOutput`/`CloneOutput` 接口；这些文件位于 `cmd/ycy`，不是 `internal/commands/git/*`。
- `run_process.go`、`run_process_unix.go`、`run_process_windows.go` 是 `run` 命令 child process 生命周期、stdin/stdout/stderr 透传和平台终止实现。
- `process_errors.go` 统一 `exec.ErrNotFound` 到 `fs.ErrNotExist` 的错误包装。
- `signals.go`、`signals_unix.go`、`signals_windows.go` 负责根进程 signal context；signal ownership 只在 `cmd/ycy`。

### `cmd/ycy` 测试

- 测试与 production adapter 同目录、同 `main` package；命令测试文件按当前平铺文件名对应，例如 `gitcm_test.go`、`gitcm_process_test.go`、`gitcm_integration_test.go`。
- `*_integration_test.go` 覆盖独立二进制、HTTP 服务、Unix signal 和长运行流程；`standalone_binary_test.go` 提供 Windows `.exe` 路径兼容测试。
- 多数测试使用 `t.TempDir()`、`os.CreateTemp`、写入临时 `config.json`、`exec.Command("go", "build", ..., "./cmd/ycy")`，没有仓库内 `testdata`/`fixtures` 目录。
- 终端交互测试使用 `internal/terminaltest.NewRecordingExperience`；PTY 场景通过 `terminaltest.StartPTY` 和 helper process 执行。
- standalone/integration 测试会依赖存在的 `web/dist`，构建真实 `./cmd/ycy`，并使用本地临时目录、端口和 Git/FRP 子进程。

## `internal/cliapp`：Cobra 树和 handler contracts

`internal/cliapp` 有 38 个 Go 文件，其中 19 个生产文件、19 个测试文件，均为 package `cliapp`。

- `app.go` 定义 `BuildInfo`、`Dependencies`、`Outcome`、`App`；保存所有命令 handler 字段，设置默认 IO/environment/logging，构造 root Cobra tree，处理 diagnostic aliases，捕获 panic/错误并返回 `Outcome`。它不调用 `os.Exit`。
- `config.go`、`configcm.go`、`configfork.go` 注册 `config`、`config cm`、`config fork` 及参数/flags；`configcm.go` 用 `internal/commands/config/cm` 的 typed Input/Request/Result；`configfork.go` 用 `internal/commands/config/fork` 类型。
- `exportenv.go`、`rm.go`、`run.go`、`zip.go`、`diff.go`、`fs.go`、`upgrade.go` 分别注册对应一级命令并声明 handler type；`githeat.go`、`gitpulse.go`、`gitfork.go`、`gitcm.go` 注册 `git` 及子命令；`tunnel.go` 注册 `tunnel server/connect`。
- `diagnostics.go` 处理 `--log-level`、`--log-format`、`--verbose`、`--quiet` 和环境解析；`discovery.go` 定义 Cobra discovery document/presenter contract；`errors.go` 定义 Cobra/help/exit 归一化。
- `internal/cliapp` 是唯一允许导入 `github.com/spf13/cobra` 和 `github.com/spf13/pflag` 的活动目录（由 `cmd/ycy/architecture_test.go` 强制）。
- `cliapp` 生产代码直接导入的项目包是 `internal/commands/*`（typed contracts）和 `internal/logging`；不导入 `cmd/ycy`、`web` 或 `legacy/bun`。
- 测试通过 fake handler 和 `logging.NewRuntime` 验证 flags、参数数目、诊断配置、help 和 exit-coded errors；`discovery_contract_test.go` 使用 `internal/terminaltest` 检查重定向输出。

## `internal/commands`：当前业务 package

`internal/commands` 没有父 package；当前 package 由命令域/子域目录组成。下表统计 Go 文件（生产/测试）并记录文件名所表达的职责。每个 package 的测试都与源码同目录，当前没有命令 package 自有 `testdata` 或 `fixtures` 子目录。

| Package | Go 文件（生产/测试） | 当前职责和主要文件 |
| --- | ---: | --- |
| `config/cm` | 20 / 18 | CM profile list/add/use/set/remove/test；`input.go`、`*_input.go`、`*_presentation.go`、`*_run.go`、`*_write.go`、`read.go`、`test_transport.go`；直接依赖 `internal/appconfig` 和标准库/HTTP |
| `config/fork` | 12 / 11 | Fork provider 配置 list/add/remove；`input.go`、`add_*`、`remove_*`、`run.go`、`read.go`；直接依赖 `internal/appconfig` 和标准库 |
| `diff` | 26 / 24 | 目录 comparison workspace/discovery/snapshot、文本和 blob 内容、HTTP API、MCP、嵌入 Web server、listener 生命周期；`model.go`、`workspace.go`、`http*.go`、`mcp*.go`、`server*.go`、`run.go`；依赖 `web`、MCP SDK、gitignore/doublestar/jsonschema、`x/sys/unix` |
| `exportenv` | 8 / 8 | dotenv discovery/parse/select/read/write/JSON/presentation；依赖标准库，无其他活动内部 package |
| `fs` | 37 / 29 | 文件 workspace、只读和管理 API、auth/session、upload/download/edit、archive/extraction、thumbnail、HTTP server、运行时；依赖 `internal/commands/fs/sevenzipruntime`、`internal/filesession`、`web` 和图像/网络/加密库 |
| `fs/sevenzipruntime` | 8 / 2 | 7-Zip runtime 查找、校验、提取和平台 `go:embed` payload；依赖 `internal/sevenzipmanifest`。payload 下另有各平台二进制与 License 文件 |
| `git/cm` | 14 / 11 | Git 状态/快照/evidence、commit message provider、分类、stage/commit/push/mutation、tracking；`git.go`、`snapshot.go`、`evidence*.go`、`provider.go`、`run.go`；依赖 `internal/appconfig` 和标准库/HTTP |
| `git/fork` | 6 / 5 | provider URL/branch/archive、clone fallback、progress tracking；`input.go`、`provider.go`、`archive.go`、`clone.go`、`run.go`；依赖 `internal/appconfig` 和标准库/HTTP |
| `git/heat` | 6 / 6 | Git log 读取/解析、文件与目录 heat aggregation、报告和输入归一化；`git.go`、`log.go`、`aggregate.go`、`report.go`、`run.go` |
| `git/pulse` | 9 / 8 | repository scan、Git commit fetch、author/day selection、report、progress tracking；`discovery.go`、`git.go`、`author.go`、`input.go`、`report.go`、`run.go` |
| `rm` | 6 / 6 | explicit/smart deletion、prompt、presentation、run；依赖标准库 |
| `run` | 7 / 6 | `package.json` script discovery、prompt/presentation、child-run request；依赖标准库 |
| `tunnel` | 42 / 32 | client/server config、client reconciliation/agent、SQLite state/accounts/sessions/tunnels、FRP runtime/supervisor/TOML/protocol、server HTTP/control plane、Web handler；依赖 `internal/appconfig`、`internal/filesession`、`internal/logging`、`web` 和 websocket/sqlite/toml/crypto 网络库 |
| `upgrade` | 17 / 12 | release metadata/checksum/artifact resolution、download/verify、update transaction/state、replacement/retry、internal updater；含 Unix/Windows process/permissions 文件；依赖标准库 |
| `zip` | 5 / 4 | project/workspace discovery、source selection、planning session、archive build/write、presentation；依赖标准库 |

### 命令 package 的依赖方向

- `cmd/ycy` 直接导入所有活动 `internal/commands/*` package，用于构造 Module 和适配 OS/terminal/network dependencies。
- `internal/cliapp` 直接导入所有需要 typed handler contract 的命令 package；它还导入 `internal/logging` 和 Cobra/pflag。
- 命令 package 之间没有跨域导入。`architecture_test.go` 的 `commandOwner` 将 `internal/commands/config/*` 和 `internal/commands/git/*` 视为各自 owner；导入 sibling owner 会失败。当前唯一的嵌套命令依赖是 `internal/commands/fs` -> `internal/commands/fs/sevenzipruntime`。
- 活动代码没有导入 `legacy/bun`。`architecture_test.go` 还禁止非 `cmd/ycy` 的 `os.Exit`、非 `internal/appconfig` 的 `config.json`/`.config.lock` 字面量，以及非命令 owner 的 `web` 导入。

## 共享和基础 package

### `internal/appconfig`

18 个生产 Go 文件、7 个测试文件。`store.go`、`schema.go`、`operations.go`、`publish.go`、`lock.go`、`crypto.go`、`machineid*.go`、`process_*.go`、`replace_*.go` 实现用户目录配置、JSON schema、加密、锁、原子发布和平台进程/替换。`cm.go`、`fork.go`、`tunnel.go` 提供相应配置解析/操作。生产代码只依赖标准库；使用方包括 `cmd/ycy`、`internal/commands/config/cm`、`config/fork`、`git/cm`、`git/fork` 和 `tunnel`。架构测试把配置持久化文件名 ownership 固定在此目录。

### `internal/filesession`

7 个生产 Go 文件、4 个测试文件。`store.go` 是 831 行的文件会话 store；`replace_*`、`process_*`、`permissions_*` 是文件替换/进程/权限平台实现。生产代码只依赖标准库；`internal/commands/fs` 和 `internal/commands/tunnel` 导入它。Windows implementation 还导入 `internal/windowsacl`。

### `internal/logging`

3 个生产 Go 文件、2 个测试文件。`configuration.go` 解析 diagnostic configuration；`logging.go` 实现 runtime/logger；`redaction.go` 实现敏感值 redaction。`cmd/ycy`、`internal/cliapp` 和 `internal/commands/tunnel` 生产代码使用它；测试还使用 `internal/terminaltest`。

### `internal/terminal`

7 个生产 Go 文件、8 个测试文件。`session.go`/`experience.go` 分类和创建 terminal experience；`interaction.go`/`presentation.go` 实现语义交互与输出；`tracked.go`/`lease.go` 实现 tracked operation 和 renderer lease；`semantic.go` 定义语义类型。生产代码依赖 Bubble Tea、Huh、Lipgloss 和 `golang.org/x/term`。生产使用方主要是 `cmd/ycy`；测试依赖 `internal/terminaltest`。

### `internal/terminaltest`

7 个生产 Go 文件、1 个 package 自测。`facts.go`、`streams.go`、`recording.go`、`semantic_recording.go` 和 `pty*.go` 提供测试 stream、terminal facts、semantic answer recording 和平台 PTY process。`cmd/ycy`、`internal/cliapp`、`internal/logging`、`internal/terminal` 的测试直接导入它。

### `internal/sevenzipmanifest` 和 `internal/windowsacl`

- `internal/sevenzipmanifest` 有 1 个生产文件和 1 个测试，manifest API 被 `tools/prepare-sevenzip` 和 `internal/commands/fs/sevenzipruntime` 使用。
- `internal/windowsacl` 有 Unix/Windows 实现和 Windows 测试；`internal/filesession/permissions_windows.go` 在 Windows build 中调用它。

## `web`：前端、嵌入和路由 ownership

`web` package 的 Go 部分有 3 个生产文件、4 个测试文件；前端 TypeScript/TSX 源码按应用目录组织：`web/diff`、`web/fs`、`web/tunnel-server`，共用组件位于 `web/shared`。主要前端入口是各应用的 `main.tsx`/`app.tsx`；大型文件包括 `web/fs/app.tsx`（1485 行）和 `web/tunnel-server/client-pages.tsx`（882 行）。

- `web/assets.go` 的 package 名为 `webassets`，通过 `//go:embed dist` 嵌入生成的 Vite 输出；固定 shell 为 `diff/index.html`、`fs/index.html`、`tunnel-server/index.html`。`Validate` 检查三个 shell，`Load` 选择一个 shell，`Site` 提供 asset/shell 服务。
- `web/readiness_handler.go` 提供 `NewFSProductionHandler`、`NewTunnelProductionHandler` 和 static readiness handlers；它把命令-owned API adapter 与对应浏览器 shell 组合起来，但不负责命令 lifecycle。
- `internal/commands/diff/server_handler.go` 直接导入 `web` 组合 Diff API 与 Diff shell；`internal/commands/fs/http.go` 组合 FS API 与 FS shell；`internal/commands/tunnel/server_runtime.go` 组合 Tunnel control-plane API 与 Tunnel shell；`cmd/ycy/main.go` 调用 `web.Validate`。
- `tools/web-browser-harness/main.go` 导入 `web`，构造 readiness handler；`architecture_test.go` 只允许 `cmd/ycy`、`internal/commands/diff`、`internal/commands/fs`、`internal/commands/tunnel` 作为 `web` consumer。
- `web` 的 Go route tests 使用 `httptest` 和嵌入 dist 资产。前端测试/构建由 `pnpm --dir web` 管理。`web/dist`、`web/node_modules` 是生成/依赖目录，不作为源码盘点。

## `tools`、构建和生成输出

- `tools/check-no-bun` 扫描活动源码/配置，拒绝活动工具链引用旧 Bun token、`simple-git-hooks` 或 `lint-staged`，并拒绝若干过时的根入口。它跳过 `legacy`、`.scratch`、`mock`、`web/dist`、`build` 和 `tools/lefthook/bin`。
- `tools/hookctl` 管理仓库 hook 的 install/doctor/uninstall；有独立 `tools/lefthook/go.mod` 和生成的 `tools/lefthook/bin/lefthook`。
- `tools/prepare-sevenzip` 使用 `internal/sevenzipmanifest` 下载/校验各平台 7-Zip；生成 `internal/commands/fs/sevenzipruntime/payload/<target>` 下的二进制和 License。
- `tools/release-artifacts` 检查 release 二进制、嵌入 Web 资产、7-Zip payload 和 Tunnel FRP manifest；`tools/web-browser-harness` 启动 Web readiness harness。
- `Makefile` 的 `GO_FIND` 扫描 `cmd`、`internal`、指定 `tools` 子目录和 `web` 的 Go 文件；`make build` 与 `make cross-build` 始终构建 `./cmd/ycy`，并用 linker flags 注入 `main.version`。`make check-go` 运行 `go vet ./...` 和 `go test ./...`；`make check` 还包括 Web、lock 和 `check-no-bun`。
- `web/dist`、`build`、`release`、7-Zip payload 和 Lefthook binary 在 Makefile 中被视为生成候选输出；`release-untracked` 检查这些路径没有被 Git 跟踪。

## `legacy/bun` 与 `mock` 边界

- `legacy/bun` 当前有 224 个文件，包含 `src/cli.ts`、`src/commands/*`、`src/config/*`、`src/shared/*`、Bun package/lock/build/deploy 文件和 `ARCHIVE.md`。README、DEVELOPMENT 和 `tools/check-no-bun` 将它描述为行为兼容参考，不是当前启动入口。
- `cmd/ycy/architecture_test.go` 遍历活动树时跳过 `legacy`、`.scratch`、`mock`、`tools` 和生成目录；对活动 Go 文件检测 legacy import、Cobra ownership、command sibling imports、web ownership、config file literals 和 `os.Exit` ownership。
- `mock/nginx-proxy-manager` 是独立的 Node/TypeScript 模拟应用，包含 backend schema/lib/migrations/routes、frontend、Docker 和 Cypress tests。活动 Go 包不导入 mock；当前架构测试也跳过整个 `mock`。

## 文件名和路径可发现性事实

活动源码中 basename 最长的 Go/TS 文件包括：

| 文件 | basename 长度 | 事实 |
| --- | ---: | --- |
| `internal/commands/tunnel/client_forwarding_integration_test.go` | 37 | Tunnel client/server forwarding integration test |
| `internal/commands/tunnel/server_runtime_acquisition_test.go` | 34 | Tunnel server runtime acquisition test |
| `internal/commands/tunnel/server_http_tunnel_import_input.go` | 34 | Server HTTP tunnel import input |
| `internal/commands/tunnel/server_http_tunnel_patch_input.go` | 33 | Server HTTP tunnel patch input |
| `internal/commands/tunnel/server_domain_validation_test.go` | 32 | Server domain validation test |
| `internal/commands/tunnel/file_permissions_windows_test.go` | 32 | Windows file permissions test |
| `cmd/ycy/configcmset_presentation_test.go` | 32 | Config CM set presentation test |
| `internal/commands/tunnel/server_http_custom_404_input.go` | 31 | Server HTTP custom 404 input |
| `internal/commands/upgrade/standalone_integration_test.go` | 30 | Upgrade standalone integration test |
| `internal/commands/tunnel/file_permissions_other_test.go` | 30 | Non-Windows file permissions test |

`cmd/ycy` 生产文件的 command-prefixed平铺命名包括 `configcm.go`、`configcmadd.go`、`configcmremove.go`、`configcmtest.go`、`configfork.go`、`configforkadd.go`、`configforkremove.go`、`githeat.go`、`gitpulse.go`、`gitfork.go`、`gitcm.go` 以及对应的 `_process.go`/`_integration_test.go` 文件。`internal/commands/tunnel` 将多个 ownership 维度编码在 basename 中（`client_*`、`server_*`、`frp_*`、`database_*`、`file_permissions_*`）。

## 机械迁移的已观察风险边

以下是源码当前形状造成的可验证结构风险，不是目标设计判断：

1. `cmd/ycy/main.go` 同时拥有进程启动、根终端、signal、upgrade/worker 分流、Web 资产验证和所有 handler 组装；移动任何一个命令 adapter 都可能同时影响 `main.go` 的 `cliapp.Dependencies` 字段 wiring 和同 package 的测试辅助。
2. `internal/cliapp.Dependencies` 和 `App` 为每个命令 leaf 保存独立 handler 字段（包括 `ConfigCM*`、`ConfigFork*`、`Git*`、`Tunnel*`），命令注册文件通过这些字段决定是否向 root tree 添加子命令；路径移动会触及 typed contract 的导入路径和条件注册测试。
3. `cmd/ycy` 的 Git process helper 同时被四个 Git command adapter 使用；`run_process*` 同时承载 child process lifecycle 和平台 signal。它们与 command-specific `GitOutput`/`Result` 类型耦合，不能只按文件名移动而不更新 imports。
4. `internal/commands/diff`、`fs`、`tunnel` 直接组合 `webassets.Site` 和 HTTP routes；`web` 的 package name 是 `webassets`，但 import path 是根 `web`。Web shell、embedded dist 和 command handler 的组合关系跨目录存在。
5. `internal/commands/fs` 依赖嵌套 package `fs/sevenzipruntime`，后者依赖 `internal/sevenzipmanifest` 并包含 platform `go:embed` 文件；移动 `fs` 或 payload 目录会影响 `make prepare-7zip`、编译时 embed 和 release verifier。
6. `internal/appconfig` 的 `config.json`、`.config.lock` 字面量 ownership 被 AST 架构测试锁定；配置使用方在 `cmd/ycy` 和多个 command package，测试大量构造临时 HOME/config 文件。
7. `architecture_test.go` 当前把 `pkg` 列为 forbidden generic package segment（同 `utils`、`services`、`interfaces`、`adapters`、`common`）；活动目录若出现 `pkg/...` 会被当前门禁报告。该测试还跳过 `tools`，因此工具包不受同一 import 检查覆盖。
8. `architecture_test.go` 的 `commandOwner` 只识别 `internal/commands/config/*`、`internal/commands/git/*` 和其他一级 command package；改变 command package 层级会改变 sibling-import 判定范围。
9. `make`、测试和 README/DEVELOPMENT 中硬编码了 `./cmd/ycy`、`internal/commands/<domain>`、`internal/commands/fs/sevenzipruntime/payload`、`web/dist` 和测试命令路径；机械移动会同时影响构建、检查、文档和测试命令。
10. 当前活动 Go 测试没有仓库内通用 fixture 目录；fixture 多由 `t.TempDir()`、写临时 JSON/源文件、`httptest`、recording experience、PTY helper 或 standalone build 在测试中即时创建。仅移动源码目录不能假定存在可独立搬迁的 fixture 文件。
11. `legacy/bun` 被构建文档和 `check-no-bun` 明确作为冻结边界；活动代码不得导入它。`mock`、`legacy`、`.scratch` 和 generated output 被架构遍历跳过，导致这些目录的结构不会通过活动树 AST 门禁验证。

## 当前活动调用链

现有开发指南描述的主调用链与源码一致：

```text
命令行参数
  -> cmd/ycy
       进程分流、终端/日志/signal、OS 和网络依赖、handler 组装、退出码
  -> internal/cliapp
       Cobra root tree、flags、参数约束、diagnostic configuration、Outcome
  -> internal/commands/<domain>
       输入模型、Module、业务/协议、presentation contract、生命周期
  -> web/
       diff/fs/tunnel 的前端源码和嵌入 shell/asset serving
```

`legacy/bun` 不在活动调用链中；`mock` 只作为外部模拟服务目录存在。
