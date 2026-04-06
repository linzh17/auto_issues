---
name: auto-issue-v2
description: "Automatically processes GitHub/GitLab issues: classifies with labels, generates AI replies, creates code implementations, and submits MRs. Use when asked to auto-manage issues, patrol issues, or handle GitHub/GitLab issue workflows."
---

# Auto Issue Agent v2

Automatically processes GitHub and GitLab issues using AI analysis.

## Workflow

当用户说 "处理 issue" 或 "start issue patrol" 时，执行以下流程：

### Step 1: 配置检查

读取配置文件 `~/.config/agents/skills/auto-issue-v2/config.yaml`：

```yaml
github:
  enabled: false
  token: "${GH_TOKEN}"
  repositories:
    - "owner/repo1"

gitlab:
  enabled: true
  token: "${GITLAB_TOKEN}"
  base_url: "https://{your-gitlab-host}.com"  # 替换为实际实例
  repositories:
    - "group/project1"
```

**常用 GitLab 实例：**
| 实例 | base_url |
|------|----------|
| GitLab 官方 | `https://gitlab.com` |
| 极狐GitLab | `https://jihulab.com` |
| 自定义私有 | `https://git.yourcompany.com` |

### Step 2: 获取 Issues

**GitHub:**
```bash
gh issue list --repo owner/repo --state open --json number,title,body,labels
```

**GitLab (通用):**
```bash
glab issue list --repo group/project --opened
# 或使用 API
curl -s --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "${GITLAB_BASE_URL}/api/v4/projects/$(echo group/project | sed 's/\//%2F/g')/issues?state=opened"
```

### Step 3: 处理每个 Issue

对每个新/未处理的 issue，执行：

#### 3.1 分析 Issue 内容

读取 issue 的 title、body、description，判断：
- 问题类型：bug / feature / question / documentation
- 是否需要代码实现
- 优先级

#### 3.2 添加标签

**GitHub:**
```bash
gh issue edit {number} --repo owner/repo --add-label "bug,priority:high"
```

**GitLab:**
```bash
glab issue update {iid} --repo group/project --add-label "bug,priority:high"
# 或 API
curl -s --request PUT \
  --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  --header "Content-Type: application/json" \
  --data '{"labels":["bug","priority:high"]}' \
  "${GITLAB_BASE_URL}/api/v4/projects/{project_id}/issues/{iid}"
```

**自动标签规则：**
| 关键词 | 标签 |
|--------|------|
| bug / 错误 / 修复 | bug |
| feature / 功能 / 新增 | enhancement |
| urgent / 紧急 / critical | priority:high |
| question / 怎么 / 如何 | question |
| 文档 / 文档 / 翻译 | documentation |

#### 3.3 生成 AI 回复

使用 `amp` 或 `claude-code` 生成回复：

```bash
amp "分析这个 Issue，生成一条友好的回复：
Title: {title}
Description: {description}

回复要包含：
1. 问题摘要
2. 下一步或问题
3. 是否需要代码实现

保持简洁专业，markdown 格式。"
```

然后发布评论：

**GitHub:**
```bash
gh issue comment create {number} --repo owner/repo --body "{回复内容}"
```

**GitLab:**
```bash
glab issue note {iid} --repo group/project --message "{回复内容}"
```

#### 3.4 代码实现（如需要）

判断 issue 是否需要代码实现。如果是：

1. **克隆仓库**
```bash
git clone ${GIT_BASE_URL}/group/project.git
# 例如:
# https://gitlab.com/owner/repo.git
# https://jihulab.com/owner/repo.git
# https://git.yourcompany.com/owner/repo.git
```

2. **创建分支**
```bash
git checkout -b fix/issue-{number}
```

3. **运行 coding agent**
```bash
amp "实现 Issue #{number} 的功能：
Title: {title}
Description: {description}

要求：
1. 实现所需功能
2. 代码清晰可维护
3. 包含测试（如适用）
4. 更新文档（如需要）"
```

4. **提交并推送**
```bash
git add -A
git commit -m "Fix issue #{number}: {title}"
git push -u origin fix/issue-{number}
```

5. **更新 Changelog**

每次代码改动必须更新对应项目的 changelog：
- 如果项目没有 changelog，则新建 `CHANGELOG.md`
- 如果已有 changelog，则追加新条目到顶部

```markdown
## [Unreleased] - {日期}

### Added/Changed/Fixed
- 描述本次改动内容
```

6. **创建 MR/PR**

**注意：** 提交代码改动和 changelog 更新后，再创建 MR/PR。

**GitHub:**
```bash
gh pr create --repo owner/repo \
  --title "Fix: {title}" \
  --body "解决 issue #{number}" \
  --base main
```

**GitLab:**
```bash
glab mr create --repo group/project \
  --title "Fix: {title}" \
  --description "解决 issue #{number}" \
  --target-branch main
```

7. **评论通知**
```bash
gh issue comment create {number} --repo owner/repo --body "已创建 PR: {url}"
glab issue note {iid} --repo group/project --message "已创建 MR: {url}"
```

### Step 4: 记录状态

记录已处理的 issue，避免重复处理：
```bash
echo "{issue_id}:$(date +%s)" >> ~/.config/agents/skills/auto-issue-v2/processed.log
```

## 命令示例

```
用户: 处理 group/project 的 issues
Agent:
  1. 检查配置文件获取 base_url
  2. 获取 issues
  3. 对每个 issue：添加标签 → AI 回复 → 代码实现 → 创建 MR
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `GITHUB_TOKEN` / `GH_TOKEN` | GitHub 访问令牌 |
| `GITLAB_TOKEN` | GitLab 访问令牌 |
| `GITLAB_BASE_URL` | GitLab 实例 URL（默认: https://gitlab.com）|
| `GIT_BASE_URL` | Git clone 用的基础 URL（通常与 GITLAB_BASE_URL 相同）|

## 认证方式

1. `glab auth status` / `gh auth status` - CLI 已登录
2. 环境变量 `GITLAB_TOKEN` / `GITHUB_TOKEN`
3. 配置文件中的 token

## 提示

- 使用 `glab` / `gh` CLI 优先，API 作为 fallback
- 代码生成推荐 `claude-code`，`amp` 需要交互模式
- 私有仓库需要配置 SSH key 或使用 token
- 处理前检查 issue 是否已处理（避免重复）
- 自定义 GitLab 实例需确保 token 有对应权限
