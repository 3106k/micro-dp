---
name: plan-issue
description: Read a GitHub issue and create an implementation plan with codebase investigation
allowed-tools: Bash, Read, Grep, Glob, WebSearch, WebFetch, AskUserQuestion, Agent
argument-hint: "<issue number>"
---

# Plan Issue Skill

既存の GitHub issue を読み、コードベースを調査して実装プランを立てる。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Remote: !`git remote get-url origin`

## Steps

### 1. Issue の読み込み

`$ARGUMENTS` から issue 番号を取得し、内容を読み込む:

```bash
gh issue view <number>
```

issue の背景・スコープ・実装方針・受け入れ条件を把握する。

### 2. コードベース深掘り調査

Agent (Explore) で issue に関連するコードを徹底的に調査する:

- **既存パターン**: 類似機能の実装を特定し、踏襲すべきパターンを把握
- **変更対象ファイル**: 追加・変更が必要なファイルを具体的にリストアップ
- **データモデル**: 既存テーブル・マイグレーション番号の現状確認
- **依存関係**: import グラフ、DI の配線 (`cmd/api/main.go`, `cmd/worker/main.go`)
- **テスト**: 既存 E2E シナリオ、影響するテストケース
- **OpenAPI**: spec 変更が必要な場合、既存エンドポイントとの整合

### 3. 実装ステップの設計

調査結果をもとに、具体的な実装ステップを設計する:

```markdown
## 実装プラン

### Step 1: [タイトル]
- 対象ファイル: `path/to/file.go`
- 作業内容: ...
- 確認方法: ...

### Step 2: [タイトル]
...
```

各ステップは以下の基準で分割する:
- 1ステップ = 1つの論理的な変更単位
- ステップ間の依存関係を明示
- 各ステップ後の確認方法 (ビルド、テスト等) を記載

### 4. ユーザー確認

プランをユーザーに提示し、方針の確認を行う:
- 技術選定に迷う点があれば選択肢を提示
- スコープの調整が必要ならその旨を伝える

### 5. Issue コメントに投稿

確認後、実装プランを issue コメントとして投稿する:

```bash
gh issue comment <number> --body "$(cat <<'EOF'
## 実装プラン

...

---
🤖 Generated with Claude Code
EOF
)"
```

投稿した旨を報告する。
