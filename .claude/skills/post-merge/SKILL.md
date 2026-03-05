---
name: post-merge
description: Run post-merge workflow — update project board, unblock issues, check docs, propose next work
allowed-tools: Bash, Read, Grep, Glob, AskUserQuestion, Agent
argument-hint: "[issue number]"
---

# Post-merge Skill

PR マージ後の一連のワークフローを実行する。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Recent merges: !`gh pr list --state merged --limit 3 --json number,title,mergedAt --jq '.[] | "#\(.number) \(.title)"'`

## Steps

### 1. マージされた issue の特定

`$ARGUMENTS` で issue 番号が指定されていればそれを使う。
指定がなければ、最近マージされた PR から `Closes #N` を抽出して特定する:

```bash
gh pr list --state merged --limit 1 --json body --jq '.[0].body' | grep -oP 'Closes #\K\d+'
```

### 2. Status → Done に更新

対象 issue の Project Board Status を **Done** に更新する:

```bash
# item_id を取得
gh project item-list 2 --owner 3106k --format json | python3 -c "
import json, sys
data = json.load(sys.stdin)
for item in data.get('items', []):
    if item.get('content', {}).get('number') == <issue_number>:
        print(item['id'])
"

# Status → Done
gh api graphql -f query='
mutation {
  updateProjectV2ItemFieldValue(input: {
    projectId: "PVT_kwHOAC1ux84BQwnR"
    itemId: "<item_id>"
    fieldId: "PVTSSF_lAHOAC1ux84BQwnRzg-yVxk"
    value: { singleSelectOptionId: "98236657" }
  }) {
    projectV2Item { id }
  }
}'
```

### 3. ブロック解除

Project Board で **Blocked** ステータスの issue を確認し、DependsOn にマージ済み issue を参照しているものがあれば:

1. DependsOn のブロック元が全て Done か確認
2. 全て解消していれば Status: Blocked → **Ready** に更新
3. 一部未解消ならそのまま Blocked を維持

### 4. ドキュメント更新

マージされた PR の差分を調査し、ドキュメントを更新する。
**「不要」と安易に判断しないこと。** Feature / Refactor PR では何かしら更新があるのが普通。

#### 4a. 差分の取得

```bash
# マージされた PR の変更ファイルを確認
gh pr diff <pr_number> --name-only

# 主要な変更内容を確認
gh pr diff <pr_number> | head -500
```

#### 4b. 各ドキュメントのチェックリスト

以下を **1 つずつ** 確認し、更新が必要なものを特定する:

**CLAUDE.md** — 以下のいずれかに該当するか:
- [ ] 新しい API エンドポイント追加 → API Endpoints テーブルに追加
- [ ] 既存エンドポイントの説明変更 (stub → 実装 等)
- [ ] 新しい環境変数追加
- [ ] アーキテクチャ変更 (新しいパイプライン、新しいサービス間通信等)
- [ ] 新しい make コマンド追加
- [ ] 運用ルール変更

**MEMORY.md** — 以下のいずれかに該当するか:
- [ ] 新しい usecase / handler パターン (DI、クエリ構成、エラーハンドリング等)
- [ ] 新しいフロントエンドパターン (コンポーネント設計、API 通信、状態管理等)
- [ ] 踏襲すべき設計判断 (なぜこの実装にしたか)
- [ ] 踏まえるべきハマりポイント (NULL 対応、型変換、ライブラリの癖等)
- [ ] 新しい依存関係やライブラリの使い方

**Skills** — 以下のいずれかに該当するか:
- [ ] 既存スキルの手順に影響する変更
- [ ] 新しいワークフローが追加された

#### 4c. 更新の実行

1. 既存の MEMORY.md を Read で読み、重複がないか確認
2. 更新内容のプランをユーザーに提示
3. 承認を得てから更新を実施

全チェックリストに該当なしの場合のみ「ドキュメント更新不要」と報告。

### 5. 次 issue の提案

Project Board を確認し、次の着手候補を提案する:

1. 現在の In Progress 件数を確認 (WIP 制限: 最大 2 件)
2. ブロック解除された issue があればそれを優先提示
3. なければ Ready の Priority 高い順に提案
4. 候補の issue 番号・タイトル・Priority・Area を一覧表示

```bash
gh project item-list 2 --owner 3106k --format json
```

### 6. 結果報告

| Step | Result |
|------|--------|
| Issue | #N タイトル |
| Status update | Done |
| Unblocked | #X, #Y / なし |
| Docs update | 必要 (内容) / 不要 |
| Next candidates | #A [P0], #B [P1] / なし |
