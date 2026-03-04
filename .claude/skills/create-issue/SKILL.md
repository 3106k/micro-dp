---
name: create-issue
description: Create a GitHub issue following the project's feature-implementation template
allowed-tools: Bash, Read, Grep, Glob, WebSearch, WebFetch, AskUserQuestion, Agent
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

### 2. コードベース調査 (プラン)

Agent (Explore) を使い、関連するコードを調査する。以下の観点で探索する:

- **既存パターン**: 類似機能がどう実装されているか（domain / usecase / handler / db の構成）
- **影響範囲**: 変更が波及するファイル・パッケージ
- **データモデル**: 既存テーブル・マイグレーションの現状
- **フロントエンド**: 関連する UI コンポーネント・API route の構成
- **テスト**: E2E シナリオへの影響

調査結果を実装方針セクションに反映する。

### 3. スコープ判定

issue のスコープが大きすぎないか判定する:

- **独立した関心事が 3 つ以上** ある場合 → 分割を提案
- **バックエンド + フロントエンド + Worker** が全て大幅変更になる場合 → 分割を提案
- 分割する場合は依存関係 (直列 / 並列) を明示する

AskUserQuestion で分割の要否をユーザーに確認する。

### 4. Issue 本文の生成

テンプレート構成に従って Markdown 本文を生成する。

分割 issue の場合:
- 各 issue に `(N/M)` のナンバリング
- 関連 issue セクションで依存関係を明示
- 並列開発可能かどうかを記載

### 5. ユーザー確認

生成した本文をユーザーに提示し、修正点がないか確認する。

### 6. Issue 作成

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

複数 issue の場合は順に作成し、issue 番号が確定次第クロスリファレンスを更新する。

作成された issue の URL を報告する。
