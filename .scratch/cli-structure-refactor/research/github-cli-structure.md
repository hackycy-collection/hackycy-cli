# GitHub CLI 项目结构研究

## 研究范围

本记录基于 `cli/cli` 的 `trunk` 快照 `9b323de8005a9988f398ce547697bf43b944e505`，只引用 GitHub CLI 仓库自身的源码、维护文档和构建文件。源码链接均固定到该提交，避免后续分支变化造成证据漂移。

核心一手来源：

- [项目布局文档](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md)
- [维护者/贡献者约定](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/AGENTS.md)
- [贡献指南](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/.github/CONTRIBUTING.md)
- [Makefile](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/Makefile)
- [Windows/跨平台构建入口](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/script/build.go)

## 目录职责

项目布局文档明确区分了核心区域与历史遗留顶层包。职责如下：

| 目录 | `cli/cli` 的职责 | 证据 |
| --- | --- | --- |
| `cmd/` | 只放构建二进制所需的 `main` 包。`cmd/gh/main.go` 是极薄入口；`cmd/gen-docs` 是独立的文档生成工具。 | [project-layout.md#L3-L9](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L3-L9), [cmd/gh/main.go#L1-L12](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/cmd/gh/main.go#L1-L12) |
| `pkg/` | 大多数 Go 包，包含各个 `gh` 命令实现。 | [project-layout.md#L3-L9](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L3-L9) |
| `pkg/cmd/` | 面向命令树的包。命令按 CLI 语法分层：`pkg/cmd/<command>/<subcommand>/<subcommand>.go`。`root` 负责组装根命令，`factory` 负责默认依赖工厂，各业务组有自己的包和可复用的 `shared` 子包。 | [命令命名约定](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L18-L33), [pkg/cmd/root/root.go#L134-L179](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/root/root.go#L134-L179) |
| `internal/` | 仅项目自身使用的高特异性 Go 包。入口编排、配置、浏览器、遥测、特性探测等都放在这里，避免形成对外 API。 | [project-layout.md#L3-L9](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L3-L9), [AGENTS.md#L26-L37](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/AGENTS.md#L26-L37) |
| `api/` | 顶层历史包，提供 GitHub API 的主要请求工具。`api.Client` 以注入的 `*http.Client` 执行 GraphQL、Query、Mutation、REST，并统一转换 HTTP/GraphQL 错误。 | [project-layout.md#L11-L16](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L11-L16), [api/client.go#L20-L111](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/api/client.go#L20-L111) |
| `context/`、`git/` | 顶层历史包；前者只保留 git remote 引用（文档标注 deprecated），后者读取本地 git 仓库信息。 | [project-layout.md#L11-L16](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L11-L16) |
| `acceptance/` | 独立的真实 GitHub 实例验收层，使用 `go-internal/testscript` 和 `.txtar` 脚本；通过 `acceptance` build tag 与普通单测隔离。 | [acceptance/README.md#L1-L10](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/acceptance/README.md#L1-L10), [acceptance/acceptance_test.go#L1-L35](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/acceptance/acceptance_test.go#L1-L35) |
| `script/` | 构建、文档、许可证、签名和发布脚本。`script/build.go` 接收任务名和环境变量，默认构建 `bin/gh`，也处理 man pages 和 clean。 | [project-layout.md#L3-L9](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L3-L9), [script/build.go#L1-L72](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/script/build.go#L1-L72) |
| `build/` | 平台打包描述而非业务代码，例如 macOS distribution XML 与 Windows WiX 工程/WXS 文件。GoReleaser 仍在根目录 `.goreleaser.yml` 声明各平台构建和归档内容。 | [build 目录树](https://github.com/cli/cli/tree/9b323de8005a9988f398ce547697bf43b944e505/build), [.goreleaser.yml#L18-L113](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/.goreleaser.yml#L18-L113) |
| `docs/` | 维护者和贡献者文档、安装文档、项目布局、发布说明。命令帮助文本留在命令源码中，在发布阶段转换为手册页和网站文档。 | [project-layout.md#L3-L9](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L3-L9), [cmd/gen-docs/main.go#L29-L74](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/cmd/gen-docs/main.go#L29-L74) |
| `test/` | 顶层历史辅助包，项目文档明确标为 deprecated，不应新增使用。特定包的单测与 fixture 通常和被测包放在一起。 | [project-layout.md#L11-L16](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L11-L16), [test/](https://github.com/cli/cli/tree/9b323de8005a9988f398ce547697bf43b944e505/test) |
| `utils/` | 顶层历史辅助包，文档标为 deprecated，仅保留表格输出相关用途。 | [project-layout.md#L11-L16](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L11-L16), [utils/utils.go#L1-L36](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/utils/utils.go#L1-L36) |

这套布局的边界信号很清楚：`cmd` 是进程入口，`pkg/cmd` 是命令功能，`internal` 是项目私有基础设施，顶层历史包只在确有兼容需要时保留，`build`/`script`/`docs`/`acceptance` 不和业务包混放。

## 命令组织和依赖注入

### 调用链

维护文档把典型链路写成：`cmd/gh/main.go` -> `internal/ghcmd.Main()` -> `pkg/cmd/root.NewCmdRoot()` -> 具体命令包的 `RunE`。根命令负责 Cobra 命令树和全局行为，叶子命令负责业务。[AGENTS.md#L26-L41](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/AGENTS.md#L26-L41) 与 [project-layout.md#L35-L59](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L35-L59) 给出了入口和执行顺序。

`pkg/cmd/root/root.go` 显式注册一级命令并定义 Cobra 分组；需要仓库解析的命令先复制工厂，再替换 `BaseRepo` 为智能解析器，然后把同一个工厂传给 `issue`、`pr`、`repo` 等命令组。[root.go#L121-L179](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/root/root.go#L121-L179)

### 叶子包约定

以 `gh issue list` 为例：

- `pkg/cmd/issue/list/list.go` 定义 `ListOptions`，包括 IO、HTTP、配置、仓库解析函数和命令 flags；帮助、参数约束、flags 与 `RunE` 都在同一包。
- `NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error)` 从工厂复制依赖；`runF` 是构造器测试的注入点，省去真实 API/Git 副作用。
- `RunE` 内才绑定 `BaseRepo` 并执行校验，随后调用可独立测试的 `listRun(opts)`；这使得构造/解析与业务执行可以分别测试。
- `list.go` 同目录放 `http.go`、`*_test.go` 和 `fixtures/*.json`，而跨 issue/pr 使用的过滤逻辑放到 `pkg/cmd/*/shared`。

证据：[issue/list/list.go#L25-L123](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/issue/list/list.go#L25-L123)、[issue/list/list.go#L134-L210](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/issue/list/list.go#L134-L210)、[issue/list tree](https://github.com/cli/cli/tree/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/issue/list)。

### 共享工厂

`pkg/cmdutil.Factory` 是一个进程级依赖容器，字段包含 `IOStreams`、`GitClient`、`Browser`、`Prompter`、配置函数、认证 HTTP client、无认证 HTTP client、外部 HTTP client、`Remotes`、`BaseRepo` 和 `Branch`。[factory.go#L16-L47](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmdutil/factory.go#L16-L47)

`pkg/cmd/factory/default.go` 的 `New` 按依赖顺序组装这些字段：IO/配置先行，HTTP client、Git client、remote resolver、BaseRepo、Prompter、Browser、ExtensionManager、Branch 随后创建；字段多数是函数，因此配置、网络、仓库解析可以延迟到命令真正运行时。[default.go#L26-L47](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/factory/default.go#L26-L47)

维护约定要求命令中的 `BaseRepo`、`Remotes`、`Branch` 在 `RunE` 内惰性初始化，而不是构造命令时就访问环境或网络。[AGENTS.md#L49-L59](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/AGENTS.md#L49-L59)

## 测试和验收布局

### 单元测试

每个命令包旁边放 `foo.go`、`foo_test.go`，必要时放 `http.go`/`http_test.go`。维护约定建议至少分成两类测试：

1. 构造器表格测试，检查 flags 解析和 `Options` 内容；
2. `fooRun` 表格测试，检查业务、输出以及模拟 HTTP/Git 交互。

HTTP 使用 `pkg/httpmock.Registry` 注册 REST/GraphQL 响应并在测试结束时 `Verify`；IO 使用 `iostreams.Test()` 和 TTY 开关模拟终端；fixture 与命令包共置。证据：[AGENTS.md#L75-L108](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/AGENTS.md#L75-L108)、[issue/list/list_test.go#L24-L124](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/pkg/cmd/issue/list/list_test.go#L24-L124)。

项目布局文档还要求把真实 API 请求、真实 git shell-out、用户文件扫描等外部副作用 stub 掉，并把逻辑拆成可组合的小函数。[project-layout.md#L73-L84](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/docs/project-layout.md#L73-L84)

### Acceptance

`acceptance/acceptance_test.go` 使用 `//go:build acceptance`，通过 `testscript.RunMain` 把命令映射到 `ghcmd.Main`；脚本按命令组放在 `acceptance/testdata/<group>/*.txtar`。这层测试是真实 GitHub 实例上的黑盒式生命周期测试，要求 `GH_ACCEPTANCE_HOST`、`GH_ACCEPTANCE_ORG`、`GH_ACCEPTANCE_TOKEN`，脚本提供资源清理、随机命名和 JSON 断言等自定义命令。[acceptance/README.md#L1-L10](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/acceptance/README.md#L1-L10)、[acceptance/README.md#L53-L117](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/acceptance/README.md#L53-L117)

普通 `go test ./...` 不带 `acceptance` tag，因此不会意外触发真实外部资源；专门命令 `go test -tags acceptance ./acceptance` 才运行验收。[Makefile#L49-L57](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/Makefile#L49-L57)

### 顶层 `test/`

`test/` 被项目布局文档明确标为 deprecated；新测试应优先跟随具体包，跨进程/真实资源场景才进入 `acceptance/` 或已有 integration 目录。这是防止通用测试工具包继续膨胀的边界。

## 构建、文档和发布入口

- Unix 默认 `make` 目标构建 `bin/gh`；Makefile 的构建、clean、manpages、completion、test、acceptance 目标都委托给 `script/build.go` 或直接调用 Go 工具。[Makefile#L13-L57](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/Makefile#L13-L57)
- `script/build.go` 是跨平台任务入口：默认任务为 `bin/gh`，实际 `go build ... ./cmd/gh`，通过 linker flags 注入版本/日期；`manpages` 调用 `go run ./cmd/gen-docs`，`clean` 删除 `bin` 和 `share`。[build.go#L1-L72](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/script/build.go#L1-L72)
- `cmd/gen-docs` 构造一个无副作用的 root command，然后把命令源码中的 Cobra help 生成网站 Markdown 或 man pages；因此帮助文本不另建一套手工目录。[gen-docs/main.go#L29-L74](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/cmd/gen-docs/main.go#L29-L74)
- `.goreleaser.yml` 把 `./cmd/gh` 作为所有平台的 main，预先生成 manpages/completions，按 `build/` 平台资源和 `share/` 内容打包 Linux/macOS/Windows 归档及 deb/rpm。[.goreleaser.yml#L10-L113](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/.goreleaser.yml#L10-L113)
- 贡献指南把安装、构建、测试和 `docs/project-layout.md` 作为新贡献者的统一入口，避免目录规则只存在于代码习惯中。[CONTRIBUTING.md#L31-L45](https://github.com/cli/cli/blob/9b323de8005a9988f398ce547697bf43b944e505/.github/CONTRIBUTING.md#L31-L45)

## 对当前 `hackycy-cli` 的可迁移启示

以下观察来自当前工作区（生产代码未修改）：

- [cmd/ycy/main.go](../../../cmd/ycy/main.go) 已承担进程级启动、升级/thumbnail worker 分流、终端体验、信号、Web 资源校验、日志运行时和所有 handler 组装；它相当于 GitHub CLI 的 `cmd/gh` + `internal/ghcmd` 两层合并。
- [internal/cliapp/app.go](../../../internal/cliapp/app.go) 持有 Cobra 根树、全局 flags、依赖字段和退出结果；[internal/cliapp/*.go](../../../internal/cliapp) 以 `registerX` 方法分散注册各命令。
- [internal/commands](../../../internal/commands) 已按业务组拆包（`diff`、`fs`、`git/*`、`config/*`、`tunnel`、`upgrade` 等），每个组内同时有输入模型、业务模块、presentation 和大量共置测试；这比 `cmd/ycy` 中按文件堆叠命令更接近垂直切片，但还没有统一的“命令语法包”和“业务包”边界。
- `legacy/bun/` 是行为兼容参考，不应成为新的 Go 依赖方向；当前 Makefile 也有 `check-no-bun` 门禁。

在“不改功能”的前提下，建议把重构拆成可验证的结构迁移：

1. **先固定三层边界**：`cmd/ycy` 只保留 `main` 和进程级分流；`internal/cliapp` 只保留根命令、全局选项、退出码和组合；具体命令的 flags/usage/参数校验迁移到各自的命令包。不要把业务重新放回 `main`。
2. **以 CLI 语法建立叶子目录**：将 `config cm`、`config fork`、`git heat`、`git pulse`、`git fork`、`git cm`、`export env` 等逐步整理成 `<group>/<subcommand>` 目录，每个叶子拥有 `command.go`（Cobra 构造器）、`run.go`/现有模块适配器、`*_test.go` 和必要 fixture。当前 `internal/commands/<group>` 可先作为业务实现包，避免为了模仿 `pkg` 而扩大可见 API。
3. **统一工厂/依赖注入**：参考 `cmdutil.Factory`，为 ycy 保留一个组合根工厂或按命令组拆分的小工厂；把 IO、终端体验、配置、Web 资源、Git、网络、时钟等外部边界作为函数/接口注入，命令构造阶段不访问真实环境，`RunE` 中才惰性解析。现有 [cliapp.Dependencies](../../../internal/cliapp/app.go) 已是基础，可从“每个 handler 一个字段”逐步收敛为组装根负责的 typed factory。
4. **保持测试位置语义**：命令构造器测试和业务运行测试分开并共置；外部 HTTP/Git/文件系统继续使用 stub；需要真实浏览器/网络资源的流程另建 `acceptance/`，用 build tag 和 `.txtar` 按命令组组织，不能让 `go test ./...` 触发真实服务。
5. **把文档作为结构契约**：新增一份 ycy `docs/project-layout.md`（或在现有 `DEVELOPMENT.md` 中明确同等章节），列出入口链、命令目录规则、依赖边界、单测/验收命令和 deprecated 目录。命令帮助可继续嵌在 Cobra 构造器中，以便后续生成统一手册。
6. **区分业务、平台打包和脚本**：继续让 `web/` 保存前端源码/嵌入产物，`build/` 仅放平台打包描述，`scripts/` 或 `script/` 仅放构建/发布工具。可在不改行为的前提下统一命名为 `script/` 或保留 `scripts/`，但应把构建入口、版本注入、跨平台产物和清理策略写进一个稳定入口。
7. **控制历史包扩散**：像 GitHub CLI 对 `test/`、`utils/`、`context/` 的处理一样，给当前仅兼容用途的 `legacy/`、重复 helper 或旧入口标注“只读/弃用”，新代码禁止依赖；迁移完成后再单独决定删除或保留，不在结构重构中顺便改行为。

### 迁移验收信号

结构重构完成一阶段后，可以用以下不涉及功能变化的证据判断边界是否生效：

- `cmd/ycy` 中除 `main.go` 和极少数进程适配外，不再出现命令业务实现；
- 每个公开命令路径能从一个叶子目录定位到 constructor、run 逻辑和测试；
- 根命令只负责注册和全局策略，叶子构造器可在注入 fake dependencies 后独立执行；
- `go test ./...` 保持离线稳定，验收测试必须显式 tag 和环境变量；
- `make build`/`make check`/发布脚本的产物路径和 Web 嵌入行为不变；
- `legacy/bun`、旧 helper 和 deprecated 目录没有成为新包的依赖。
