---
name: project
description: Manage the GitHub Project board — view, update status, set fields on issues
allowed-tools: Bash, Read, AskUserQuestion
argument-hint: "[status | update <issue#> | list]"
---

# Project Board Skill

GitHub Projects (micro-dp #2) のボード操作と運用ルールの参照。

## Project info

- Owner: `3106k`
- Project number: `2`
- URL: https://github.com/users/3106k/projects/2

## カラム (Status) 定義

| Status | 意味 | 入る条件 | 出る条件 |
|--------|------|---------|---------|
| **Todo** | 未着手、バックログ | issue 作成時のデフォルト | Ready に移動 |
| **Ready** | 着手可能、依存なし | 仕様・依存が明確になった | 作業開始時に In Progress へ |
| **In Progress** | 作業中 | ブランチを切って着手した | PR 作成時に Review へ |
| **Review** | PR レビュー待ち | PR が作成された | マージ後に Done へ |
| **Done** | 完了 | PR がマージされた | — |
| **Blocked** | ブロック中 | 依存 issue が未完了 / 外部要因で進行不可 | ブロック解消後に元の Status へ |

## フィールド定義

### Priority

| 値 | 意味 |
|----|------|
| **P0** | 最優先、他の作業を止めてでも対応 |
| **P1** | 次のイテレーションで対応 |
| **P2** | 余裕があれば対応 |

### Area

| 値 | 対象 |
|----|------|
| **API** | Go backend (`apps/golang/backend/`) — handler, usecase, domain, db, queue |
| **Web** | Next.js frontend (`apps/node/web/`) — ページ, コンポーネント, API route |
| **Worker** | Worker (`cmd/worker/`) — consumer, writer, aggregation |
| **E2E** | E2E テスト (`apps/golang/e2e-cli/`) |

複数 Area にまたがる場合は主要な変更先を選択する。

### DependsOn

ブロックしている issue 番号をテキストで記入 (例: `#120`)。
Blocked ステータスと併用する。

## 運用ルール

### issue 作成時
1. Status: **Todo**
2. Priority, Area を設定
3. 依存がある場合は DependsOn を記入

### 作業開始時
1. Status: **Todo** → **In Progress**
2. Assignee を自分に設定
3. ブランチを作成して作業開始

### PR 作成時
1. Status: **In Progress** → **Review**
2. PR 本文に `Closes #N` を含める (マージ時に自動クローズ)

### マージ時
1. Status: **Review** → **Done** (自動 or 手動)
2. DependsOn で本 issue を参照している issue の Blocked を解除

### ブロック発生時
1. Status → **Blocked**
2. DependsOn にブロック元の issue 番号を記入
3. ブロック元が Done になったら元の Status に戻す

## コマンド

### `$ARGUMENTS` = `list` or 空

プロジェクトの全アイテムを Status 別に一覧表示する:

```bash
gh project item-list 2 --owner 3106k --format json
```

### `$ARGUMENTS` = `status`

Status ごとの件数サマリを表示する。

### `$ARGUMENTS` = `update <issue#> <field> <value>`

issue のフィールドを更新する。例:
- `update 120 status "In progress"`
- `update 120 priority P0`
- `update 120 area API`

GraphQL mutation で更新:

```bash
# Status 更新
gh api graphql -f query='
mutation {
  updateProjectV2ItemFieldValue(input: {
    projectId: "PVT_kwHOAC1ux84BQwnR"
    itemId: "<item_id>"
    fieldId: "<field_id>"
    value: { singleSelectOptionId: "<option_id>" }
  }) {
    projectV2Item { id }
  }
}'
```

item_id は `gh project item-list` から取得する。

## フィールド ID リファレンス

### Status
| 値 | option_id |
|----|-----------|
| Todo | `f75ad846` |
| Ready | `6fd8f17d` |
| In progress | `47fc9ee4` |
| Review | `f99654d4` |
| Done | `98236657` |
| Blocked | `7f898f18` |

field_id: `PVTSSF_lAHOAC1ux84BQwnRzg-yVxk`

### Priority
| 値 | option_id |
|----|-----------|
| P0 | `97d6806c` |
| P1 | `5ba46900` |
| P2 | `932973d5` |

field_id: `PVTSSF_lAHOAC1ux84BQwnRzg-yV3s`

### Area
| 値 | option_id |
|----|-----------|
| API | `44003d7a` |
| Web | `9cbd3e2a` |
| Worker | `4ba1afa6` |
| E2E | `ed989640` |

field_id: `PVTSSF_lAHOAC1ux84BQwnRzg-yWpg`

### DependsOn
field_id: `PVTF_lAHOAC1ux84BQwnRzg-yWws` (text field)

### Project ID
`PVT_kwHOAC1ux84BQwnR`
