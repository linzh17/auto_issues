# auto_issues

基于 [Amp](https://ampcode.com) Agent 的自动化 Issue 处理 CLI 工具。定时调用 AI Agent 自动处理 GitLab 上的 open issues，包括分类、回复和代码实现。

## 功能特性

- ⏰ **定时执行** — 支持自定义执行间隔（默认 30 分钟）
- 🤖 **AI 驱动** — 集成 Amp Agent，AI 自动分析 issue 类型并生成回复
- 🏷️ **自动标签** — 根据关键词自动为 issue 打标签（bug/enhancement/question 等）
- 💻 **代码实现** — 对于需要代码修复的 issue，Agent 自动克隆仓库、创建分支、实现功能并提交 MR
- 🔧 **灵活配置** — 通过命令行参数轻松配置执行间隔和并发数
- 📦 **自动安装技能** — 首次运行时自动检测并安装 Agent 技能
- 🛠️ **多 Agent 支持** — 支持 Amp、Claude Code、Cursor Agent 等多种 CLI

## 安装

```bash
git clone https://github.com/linzh17/auto_issues.git
cd auto_issues
go mod tidy
go build -o auto_issues .
```

## 使用方法

```bash
# 默认每 30 分钟执行一次，使用当前目录作为工作目录
./auto_issues

# 自定义执行间隔
./auto_issues -interval 1h
./auto_issues -interval 10m

# 指定工作目录
./auto_issues -workdir /path/to/project

# 自定义 AI prompt
./auto_issues -prompt "你的自定义指令"

# 指定最大并发任务数（默认 5）
./auto_issues -concurrency 10

# 自动安装技能（无需询问）
./auto_issues -auto-install

# 指定技能安装路径
./auto_issues -install-path ~/.config/amp/skills/

# 指定使用的 coding agent CLI（默认 amp，可选 claude/cursor）
./auto_issues -agent claude
./auto_issues -agent cursor
```

## 支持的 Coding Agent CLI

| Agent | 命令 | 说明 |
|-------|------|------|
| amp | `-agent amp` | [Amp CLI](https://ampcode.com)，默认 |
| claude | `-agent claude` | [Claude Code](https://code.claude.com) |
| cursor | `-agent cursor` | [Cursor Agent](https://cursor.com/docs/cli/overview) |

## 配置

在 `main.go` 中修改以下配置：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-interval` | 执行间隔 | 30m |
| `-workdir` | Agent 工作目录 | 当前程序执行目录 |
| `-prompt` | 自定义 AI prompt | - |
| `-concurrency` | 最大并发任务数 | 5 |
| `-auto-install` | 自动安装技能（无需询问） | false |
| `-install-path` | 指定技能安装路径 | ~/.config/agents/skills/ |
| `-agent` | Coding Agent CLI | amp |
| `AMP_URL` | Amp 服务地址 | http://localhost:8317 |
| `AMP_API_KEY` | Amp API Key | your-api-key-1 |

## 项目结构

```
.
├── main.go              # 主程序入口
├── go.mod               # Go 依赖
├── README.md            # 本文件
└── skills/
    └── auto-issue-v2/   # Amp Agent skill
```

## 依赖

- Go 1.21+
- 至少安装以下其中一个 coding agent CLI：
  - [Amp CLI](https://ampcode.com)
  - [Claude Code](https://code.claude.com)
  - [Cursor Agent](https://cursor.com/docs/cli/overview)

## License

MIT

## Changelog

查看 [CHANGELOG.md](./CHANGELOG.md) 了解详细的更新历史。
