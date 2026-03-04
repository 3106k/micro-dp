---
name: create-pr
description: Create a GitHub pull request with auto-generated title, summary, and test plan
allowed-tools: Bash, Read, Grep, Glob, AskUserQuestion
argument-hint: "[base branch (default: main)]"
---

# Create PR Skill

現在のブランチの変更を分析し、PR を作成する。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Remote: !`git remote get-url origin`

## Steps

### 1. 変更の把握

以下を並列で確認する:

```bash
# 未コミットの変更
git status

# ベースブランチからの全コミット
git log <base>..HEAD --oneline

# ベースブランチからの差分サマリ
git diff <base>...HEAD --stat
```

ベースブランチは `$ARGUMENTS` で指定があればそれを使い、なければ `main` を使う。

未コミット・未プッシュの変更がある場合はユーザーに警告する。

### 2. 差分の分析

`git diff <base>...HEAD` の内容を分析し、以下を判定する:

- **変更の性質**: feature / fix / refactor / chore / docs
- **影響範囲**: backend / frontend / SDK / infra / skills
- **変更ファイル**: 主要な変更ファイルをリストアップ

### 3. PR タイトル・本文の生成

コミット履歴と差分から PR タイトルと本文を生成する。

**タイトル規則**:
- 70 文字以内
- prefix: `feat:` / `fix:` / `refactor:` / `chore:` / `docs:`
- 変更内容を簡潔に表現

**本文フォーマット**:
```markdown
## Summary
- 変更点を 1〜3 行の箇条書き

## Test plan
- [ ] 変更の種類に応じたテスト項目

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

**テストプラン自動推定**:
- Go backend 変更 → `go build` 確認、E2E テスト
- OpenAPI spec 変更 → `make openapi-check`
- Frontend 変更 → dev server 表示確認、UI レビュー
- Skills / templates 変更 → スキル一覧に表示されること
- Docker / infra 変更 → `make up && make health`

### 4. ユーザー確認

生成したタイトルと本文をユーザーに提示し、修正点がないか確認する。

### 5. PR 作成

確認後、リモートへのプッシュと PR 作成を行う:

```bash
# リモートにプッシュ (未プッシュの場合)
git push -u origin <branch>

# PR 作成
gh pr create \
  --title "タイトル" \
  --body "$(cat <<'EOF'
本文
EOF
)"
```

作成された PR の URL を報告する。
