---
name: create-bug-report
description: Create a GitHub bug report issue following the project's bug-report template
allowed-tools: Bash, Read, Grep, Glob, AskUserQuestion, Agent
argument-hint: "[bug description]"
---

# Create Bug Report Skill

GitHub issue を `.github/ISSUE_TEMPLATE/bug-report.md` の構成に従って作成する。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Remote: !`git remote get-url origin`
- Container status: !`cd apps/docker && docker compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || echo "not running"`

## Steps

### 1. バグ情報の整理

`$ARGUMENTS` が与えられている場合はそれを起点にする。不足している情報があれば AskUserQuestion で確認する。

以下の項目を整理する:
- **概要**: 何が問題か
- **再現手順**: ステップバイステップ
- **期待動作**: 本来どう動くべきか
- **実際の動作**: 実際にどうなるか
- **環境**: OS, Browser, Branch/Commit, Docker 状態
- **ログ・スクリーンショット**: エラーログ、コンソール出力等

### 2. コードベース調査

Agent (Explore) で関連コードを調査し、原因の仮説を立てる:

- エラーメッセージやスタックトレースから関連ファイルを特定
- 最近の変更 (`git log`) で関連するコミットがないか確認
- 影響範囲を把握

### 3. Issue 本文の生成

テンプレート構成に従って Markdown 本文を生成する。
調査で得られた仮説がある場合は「関連」セクションに追記する。

### 4. ユーザー確認

生成した本文をユーザーに提示し、修正点がないか確認する。

### 5. Issue 作成

確認後、`gh issue create` で作成する:

```bash
gh issue create \
  --title "[Bug] タイトル" \
  --label "bug" \
  --body "$(cat <<'EOF'
本文
EOF
)"
```

作成された issue の URL を報告する。
