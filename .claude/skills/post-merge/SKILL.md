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

### 4. ドキュメント更新判断

マージされた変更内容を確認し、以下の更新が必要か判断する:

| 対象 | 更新が必要なケース |
|------|-------------------|
| `CLAUDE.md` | 新しい API エンドポイント、アーキテクチャ変更、環境変数追加、運用ルール変更 |
| Skills | 新しいワークフロー追加、既存スキルの手順変更 |
| `MEMORY.md` | 新しいコードパターン、設計判断、踏襲すべき規約 |

必要がある場合:
1. 更新内容のプランをユーザーに提示
2. 承認を得てから更新を実施

不要な場合は「ドキュメント更新不要」と報告してスキップ。

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
