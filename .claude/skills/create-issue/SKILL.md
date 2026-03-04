---
name: create-issue
description: Create a GitHub issue following the project's feature-implementation template
allowed-tools: Bash, Read, Grep, Glob, WebSearch, WebFetch, AskUserQuestion
argument-hint: "[title or description]"
---

# Create Issue Skill

GitHub issue を `.github/ISSUE_TEMPLATE/feature-implementation.md` の構成に従って作成する。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Remote: !`git remote get-url origin`

## Steps

### 1. 要件の整理

`$ARGUMENTS` が与えられている場合はそれを起点にする。不足している情報があれば AskUserQuestion で確認する。

以下の項目を整理する:
- **背景・目的**: なぜ必要か、何を達成するか
- **スコープ (In/Out)**: やること・やらないこと
- **UI イメージ**: 画面変更がある場合（任意）
- **実装方針**: アーキテクチャ、データモデル、処理フローなど
- **API / Contract 影響**: OpenAPI 変更の有無
- **期待動作・受け入れ条件**: 完了条件のチェックリスト
- **依存**: 前提・後続 issue
- **参考リンク**: 関連ドキュメント、issue、PR

### 2. コードベース調査

必要に応じて Grep, Glob, Read で既存コードを確認し、実装方針の精度を上げる。

### 3. Issue 本文の生成

テンプレート構成に従って Markdown 本文を生成する。

### 4. ユーザー確認

生成した本文をユーザーに提示し、修正点がないか確認する。

### 5. Issue 作成

確認後、`gh issue create` で作成する:

```bash
gh issue create \
  --title "[Feature] タイトル" \
  --label "enhancement" \
  --body "$(cat <<'EOF'
本文
EOF
)"
```

作成された issue の URL を報告する。
