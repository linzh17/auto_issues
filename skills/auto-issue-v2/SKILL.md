---
name: auto-issue-v2
description: "Auto processes GitHub/GitLab issues: classify → AI reply → code → PR. Also resolves PR conflicts. Trigger: 处理issue, 扫描冲突."
---

# Auto Issue Agent v2

自动处理 GitHub/GitLab issues 和 PR/MR 冲突。

## CLI 优先级

| 平台 | 优先使用 | 备选 | 仅限 git |
|------|---------|------|----------|
| GitHub | `gh` | API | 克隆、提交、分支 |
| GitLab | `glab` | API | 克隆、提交、分支 |

---

## 模式一：处理 Issue

触发词：`处理issue`

### 流程概览

```
获取 open issues
    ↓
遍历每个 issue
    ↓
检查是否有对应 MR
    ├── 有 MR → 检查冲突
    │           ├── 有冲突 → 调用冲突解决流程
    │           └── 无冲突 → 跳过
    │
    └── 无 MR → 正常处理
                ├── 1. 分析 + 打标签
                ├── 2. AI 回复
                ├── 3. 代码实现（如需要）
                ├── 4. 创建 MR/PR
                └── 5. 标记 + 记录
```

### Step 1: 配置检查

读取 `~/.config/agents/skills/auto-issue-v2/config.yaml`：

```yaml
github:
  enabled: false
  token: "${GH_TOKEN}"
  repositories:
    - "owner/repo1"

gitlab:
  enabled: true
  token: "${GITLAB_TOKEN}"
  base_url: "https://gitlab.com"
  repositories:
    - "group/project1"
```

### Step 2: 获取 Open Issues

```bash
# GitHub
gh issue list --repo owner/repo --state open --json number,title,body,labels

# GitLab
glab issue list --repo group/project --opened
```

### Step 3: 检查 MR 状态

```bash
# GitHub - 获取 issue 关联的 PR
gh pr list --repo owner/repo --head "owner:fix/issue-{number}" --json number,mergeableState

# GitLab - 获取 issue 关联的 MR
glab mr list --repo group/project --search "issue-{number}"
```

**判断逻辑：**
- 有对应 MR → 检查 `mergeableState` 或 `merge_status`
- 无对应 MR → 进入正常处理流程

### Step 4: 冲突解决（仅当有冲突时）

调用下方「模式二：扫描冲突」流程。

### Step 5: 正常处理（无 MR 或 MR 无冲突）

#### 5.1 分析 Issue

- 问题类型：bug / feature / question / documentation
- 是否需要代码实现
- 优先级

#### 5.2 打标签

```bash
# GitHub
gh issue edit {number} --repo owner/repo --add-label "bug,priority:high"

# GitLab
glab issue update {iid} --repo group/project --add-label "bug,priority:high"
```

**自动标签规则：**

| 关键词 | 标签 |
|--------|------|
| bug / 错误 / 修复 | bug |
| feature / 功能 / 新增 | enhancement |
| urgent / 紧急 / critical | priority:high |
| question / 怎么 / 如何 | question |
| 文档 / 文档 / 翻译 | documentation |

#### 5.3 AI 回复

```bash
amp "分析这个 Issue，生成一条友好的回复：
Title: {title}
Description: {description}

回复要包含：1. 问题摘要 2. 下一步或问题 3. 是否需要代码实现
保持简洁专业，markdown 格式。"
```

发布评论：

```bash
# GitHub
gh issue comment create {number} --repo owner/repo --body "{回复}"

# GitLab
glab issue note {iid} --repo group/project --message "{回复}"
```

#### 5.4 代码实现（如需要）

**⚠️ 安全约束：**
- 仅限项目目录内操作
- 禁止删除 /tmp、$HOME 等外部文件
- 临时文件使用 /tmp/auto-issue-*

**执行流程：**

```bash
# 1. 创建临时目录并克隆
WORK_DIR="/tmp/auto-issue-$(date +%s)"
mkdir -p "$WORK_DIR"
git clone git@host:owner/repo.git "$WORK_DIR"
cd "$WORK_DIR"

# 2. 创建分支
git checkout -b fix/issue-{number}

# 3. 运行 coding agent
amp "实现 Issue #{number}：{title} - {description}"

# 4. 更新 Changelog
CHANGELOG="CHANGELOG.md"
if [ -f "$CHANGELOG" ]; then
  sed -i '' "1s/^/## [$(date +%Y-%m-%d)] - Issue #{number}: {title}\n\n- {description}\n\n/" "$CHANGELOG"
else
  echo "# Changelog\n\n## [$(date +%Y-%m-%d)] - Issue #{number}: {title}\n\n- {description}\n" > "$CHANGELOG"
fi

# 5. 提交推送
git add path/to/file
git commit -m "Fix issue #{number}: {title}"
git push -u origin fix/issue-{number}

# 6. 清理
cd / && rm -rf "$WORK_DIR"
```

#### 5.5 创建 MR/PR

```bash
# GitHub
gh pr create --repo owner/repo \
  --title "Fix: {title}" \
  --body "解决 issue #{number}" \
  --base main

# GitLab
glab mr create --repo group/project \
  --title "Fix: {title}" \
  --description "解决 issue #{number}" \
  --target-branch main
```

#### 5.6 标记与记录

```bash
# 标记 agent 已处理（不关闭 issue）
gh issue edit {number} --repo owner/repo --add-label "agent_processed"
glab issue update {iid} --repo group/project --add-label "agent_processed"

# 评论通知
gh issue comment create {number} --repo owner/repo --body "已创建 PR: {url}，请审查"
glab issue note {iid} --repo group/project --message "已创建 MR: {url}，请审查"

# 记录已处理
echo "{issue_id}:$(date +%s)" >> ~/.config/agents/skills/auto-issue-v2/processed.log
```

### Step 6: 后续处理

MR 合并后手动关闭 issue：

```bash
# GitHub
gh issue close {number} --repo owner/repo

# GitLab
glab issue close {iid} --repo group/project
```

---

## 模式二：扫描冲突

触发词：`扫描冲突`、`check conflicts`

### 流程概览

```
获取 open MRs/PRs
    ↓
检查冲突状态
    ↓
对有冲突的 MR
    ├── 克隆仓库
    ├── Rebase 解决冲突
    ├── 强制推送
    └── 验证结果
```

### Step 1: 配置检查

同模式一 Step 1。

### Step 2: 获取有冲突的 MR/PR

```bash
# GitHub
gh pr list --repo owner/repo --state open --json number,title,mergeableState

# GitLab
glab mr list --repo group/project --state opened
```

**冲突判断：**
- GitHub: `mergeableState` 为 `dirty`
- GitLab: `merge_status` 不为 `can_merge`

### Step 3: 解决冲突

```bash
# 克隆仓库
WORK_DIR="/tmp/pr-conflict-$(date +%s)"
git clone git@host:owner/repo.git "$WORK_DIR"
cd "$WORK_DIR"

# 获取 MR/PR 分支
gh pr checkout {number} --repo owner/repo
# 或
glab mr checkout {mr_iid} --repo group/project

# Rebase 到 main
git fetch origin main
git rebase origin/main

# ⚠️ 解决冲突后继续（重要：避免编辑器卡住）
git add <resolved-files>
GIT_EDITOR="cat" git rebase --continue

# 强制推送
git push --force-with-lease origin fix/issue-{number}

# 清理
cd / && rm -rf "$WORK_DIR"
```

> **注意**：`git rebase --continue` 会打开编辑器等待确认 commit message。使用 `GIT_EDITOR="cat"` 可以跳过编辑器直接继续。

### Step 4: 验证

```bash
# GitHub
gh pr view {number} --repo owner/repo --json mergeableState

# GitLab
glab mr view {mr_iid} --repo group/project
```

### Step 5: 记录

```bash
echo "conflict_resolved:{pr_number}:$(date +%s)" >> ~/.config/agents/skills/auto-issue-v2/processed.log
```

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `GITHUB_TOKEN` / `GH_TOKEN` | GitHub 访问令牌 |
| `GITLAB_TOKEN` | GitLab 访问令牌 |
| `GITLAB_BASE_URL` | GitLab 实例 URL |

## 认证方式

1. `glab auth status` / `gh auth status` - CLI 已登录
2. 环境变量 `GITLAB_TOKEN` / `GITHUB_TOKEN`
