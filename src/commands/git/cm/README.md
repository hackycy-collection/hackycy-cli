# `git cm` 命令

`ycy git cm` 根据当前 Git 变更生成 Angular 风格的提交信息。它可以只生成信息，也可以在用户确认后暂存、提交和推送。

```bash
ycy git cm
ycy git cm --stage
ycy git cm --stage-all
ycy git cm --staged
```

这个模块的一个重要约束是：**Git 的提交范围不能因为 LLM 的上下文预算而被截断。** 文件很多时，`cm` 压缩的是发送给模型的描述，而不是可选择、可暂存或可提交的文件集合。

## 命令语义

| 用法 | Git 范围与行为 |
| --- | --- |
| `ycy git cm` | 读取全部未提交变更并只生成提交信息，不暂存也不提交。 |
| `--stage` | 展示全部未提交文件，让用户选择指定文件；默认全选只是初始选择，用户可取消任何文件。确认后只提交选中的文件。 |
| `--stage-all` | 执行 `git add -A`，确认后提交 index 中的全部变更。 |
| `--staged` | 只根据当前 index 中的全部变更生成信息并确认提交。 |
| `--stage-push` / `--push` | 在相应的暂存和提交流程成功后推送。 |
| `--dry-run` | 只生成并打印提交信息，不改变 Git 状态。 |

`--stage` 是“选择指定文件”的交互，不是小批量提交功能。它没有文件数量上限；一百个或更多文件仍会全部出现在候选列表中。

## 范围与上下文分离

`changes.ts` 将一次运行分成两个不同的概念：

```text
git status
  -> 完整 FileChange[]
       -> --stage 的全部候选项
       -> stageFiles() 的用户选择
       -> git add -A / git commit 的实际范围
       -> 受预算约束的 LLM promptText
```

`FileChange[]` 来自 `git status --porcelain=v1 -z --untracked-files=all`，过滤暂存范围和子模块后会完整保留。不存在按“前 N 个文件”截取的逻辑。

因此，LLM 没有看到某个文件的完整 diff，不表示该文件从提交中移除。最终 commit 仍完全由 Git index 决定。

## 上下文预算

当前实现只发起一次 LLM 请求，并使用固定的字符预算控制 `promptText`：

| 常量 | 值 | 作用 |
| --- | ---: | --- |
| `MAX_PROMPT_CHARS` | 24,000 | 变更清单和 diff 上下文的总字符预算。 |
| `MAX_MANIFEST_CHARS` | 8,000 | 文件状态与路径清单最多使用的字符数，剩余空间优先留给 diff。 |
| `LARGE_FILE_BYTES` | 200,000 | 超过此大小的文件不读取原始内容。 |

这里使用的是**字符预算**，不是精确 token 计数。不同模型和语言的分词结果不同，且系统提示词与提交历史也会额外占用请求输入。`cm` profile 中的 `maxOutputTokens` 只约束模型的输出长度，不控制这部分输入上下文。

## Prompt 构建流程

### 1. 文件清单

先构造完整的 `状态 + 路径` 清单，例如：

```text
Changed files (3 total):
M src/commands/git/cm/changes.ts
A src/commands/git/cm/changes.test.ts
M src/commands/git/cm/run.ts
```

清单未超过 8,000 字符时，所有路径都会发送给模型。若路径本身超过这部分预算，会改为按“变更状态 + 首级目录”聚合，并始终保留总文件数：

```text
Changed files (1,200 total; paths grouped):
A src (840 files)
M docs (210 files)
M tests (150 files)
```

这只是 prompt 的压缩形式，不影响完整的 `FileChange[]` 或可选择文件数。

### 2. 原始 diff

在文件清单之后，剩余预算用于原始 diff：

1. 跳过不应发送内容的文件，见下节。
2. 按路径的首级目录分组；根目录文件归入 `.`。
3. 每个仍有 diff 的目录先获得相同的可用预算份额。
4. 目录内按 diff 长度从大到小选择代表性内容；长度相同时按路径排序，保证输出稳定。
5. 某个目录用不完自己的份额时，未使用空间会被后续仍待处理的目录重新分配。
6. 单个 diff 超出目录额度时，只保留前缀并加上 `...[diff truncated for prompt budget]`。

这样不会因为 Git 返回顺序而让前若干个文件占满上下文，也不会以文件数量为代价节省请求内容。它不能让模型在固定预算内逐行阅读无限大的提交，但会让模型知道完整提交范围，并从每个变更区域获得代表性证据。

发生清单聚合或 diff 截断时，命令输出会说明：

```text
Commit scope: 101 changed files; raw diffs were compressed to fit the prompt budget.
This does not affect which files are committed.
```

## 不发送原始内容的文件

以下文件仍属于 Git 提交范围，也会体现在清单或统计中，但其原始内容不会发送给模型：

- 敏感路径：`.env`、`.env.*`、`id_rsa`、`id_ed25519`、`.pem`、`.key`、`.p12`、`.pfx`。
- 常见二进制扩展名，例如图片、压缩包、数据库、字体、音视频文件。
- 锁文件与典型构建产物：`bun.lock`、各类 npm 锁文件、`dist/`、`build/`、`coverage/`、`.min.js`、source map。
- 大于 200,000 字节的文件，或 Git 判定为二进制的 diff。

这条规则防止密钥内容泄露，也避免机器生成内容或大二进制文件消耗上下文预算。它不改变 `git add`、`git commit` 的结果。

## 取舍与维护

- 该策略优先保证单次 commit 的原子性和交互选择完整性。
- 单次请求的成本稳定，但超大提交不保证每个文件的完整代码都被模型阅读。
- 若将来需要模型逐文件理解全部超大 diff，应增加“分块总结，再汇总”的多次请求模式；这能突破单次上下文窗口，但总 token 和请求次数都会增加。
- 修改预算、清单或目录分配规则时，同时更新 `changes.test.ts`。测试覆盖超过 80 个文件的明确暂存、大 diff 的目录覆盖、路径清单聚合和总字符上限。

可用的本地检查：

```bash
bun test src/commands/git/cm
bun run typecheck
bun run lint
```
