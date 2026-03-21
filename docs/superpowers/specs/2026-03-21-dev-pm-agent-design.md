# dev-pm Agent Design Spec

## Overview

dev-pm は GitHub Projects と Issues を管理し、開発エージェント (dev-engineer) にタスクを委譲して並列開発を指揮する PM エージェント。

**目的:** エピック issue の分解 → 実装 issue 選定 → 開発依頼 → コードレビュー → マージのサイクルを、承認ゲート付きで自律的に回す。

**開発方針:** feature flag + トランクベース開発。小さい PR を頻繁にマージする。

## Architecture

### Agent 構成

| Agent        | ファイル                         | 配置場所                                 | 責務                                             |
| ------------ | -------------------------------- | ---------------------------------------- | ------------------------------------------------ |
| dev-pm       | `.claude/agents/dev-pm.md`       | メインリポジトリ (`micro-dp/`)           | issue 管理、開発依頼、コードレビュー、ボード管理 |
| dev-engineer | `.claude/agents/dev-engineer.md` | 各 worktree (`../micro-dp-develop-{N}/`) | 実装、テスト、PR 作成                            |

### tmux レイアウト (1 ウィンドウ・ペイン分割)

```
tmux session: dev-pm / window 1

┌──────────────┬──────────────┐
│ PM Agent     │ Dev Agent    │
│ (1.0)        │ develop-1    │
│              │ (1.1)        │
│              ├──────────────┤
│              │ Dev Agent    │
│              │ develop-2    │
│              │ (1.2)        │
│              ├──────────────┤
│              │ Dev Agent    │
│              │ develop-3    │
│              │ (1.3)        │
│              ├──────────────┤
│              │ Dev Agent    │
│              │ develop-4    │
│              │ (1.4)        │
└──────────────┴──────────────┘
```

- 全ペインが 1 ウィンドウ内に配置。タブ切り替え不要で全体を一覧できる
- PM ペイン (1.0) は待機時間が長いため小さめでも可
- **ペイン ID は動的:** ペインの追加・削除・並び替えにより ID が変わる可能性がある。メッセージ送信前に `tmux list-panes` で対象ペインを特定すること
- tmux ターゲット: `tmux send-keys -t dev-pm:1.{N}` でペイン指定 (N は `list-panes` で確認した値)

### 通信方式: ファイル + tmux ハイブリッド

- **状態管理:** `.claude/dev-pm/status/develop-{N}.json` (正のソース)
- **通知:** `tmux send-keys` (トリガー)
- ステータスファイルが信頼できるソース、tmux メッセージは通知のみ
- **重要:** ステータスファイルはメインリポジトリに配置される。Dev Agent は worktree から**絶対パス**でアクセスする (例: `/Users/.../micro-dp/.claude/dev-pm/status/develop-1.json`)。PM Agent が `/dev-assign` 時にメインリポジトリの絶対パスを `repo_root` フィールドとして伝達する

## Status File Protocol

### ファイル配置

```
.claude/dev-pm/status/
  develop-1.json
  develop-2.json
  develop-3.json
  develop-4.json
```

`.gitignore` 対象。ランタイム状態のため VCS に含めない。

### スキーマ

```json
{
  "slot": 1,
  "status": "idle",
  "issue_number": null,
  "branch": null,
  "pr_number": null,
  "message": null,
  "started_at": null,
  "error": null,
  "updated_at": "2026-03-21T10:00:00Z",
  "version": 1
}
```

- `started_at`: 現在のステータスに遷移した時刻。staleness 検知に使用
- `error`: 失敗時のエラー概要
- `version`: スキーマバージョン (現在 `1`)。将来の拡張時に互換性を判定

**Atomic Write:** ステータスファイルの書き込みは一時ファイルに書き出してからリネーム (`mv`) する。PM と Dev が同時にアクセスした際の truncated JSON を防ぐ

### ステータス遷移

```
idle → assigned → working → review_requested → approved → done → idle
                                    ↓
                            revision_requested → working → ...
```

| status               | 意味                   | 設定者 |
| -------------------- | ---------------------- | ------ |
| `idle`               | 空きスロット           | PM     |
| `assigned`           | issue 割り当て済み     | PM     |
| `working`            | 開発中                 | Dev    |
| `review_requested`   | 開発完了、レビュー待ち | Dev    |
| `revision_requested` | レビュー差し戻し       | PM     |
| `approved`           | レビュー通過           | PM     |
| `done`               | PR マージ完了          | PM     |

### tmux メッセージフォーマット

**Dev → PM (通知):**

```
/dev-report slot:{N} status:{status} issue:#{issue_number}
```

**PM → Dev (指示):**

```
/dev-assign issue:#{issue_number} branch:{branch_name} repo_root:{absolute_path}
/dev-revise issue:#{issue_number} feedback:"修正内容"
```

### tmux メッセージのセマンティクス

- メッセージ送信と Enter は**別コマンド**で実行する (同時送信だと入力が欠落することがある):
  ```bash
  tmux send-keys -t dev-pm:1.{N} '/dev-report slot:1 status:review_requested issue:#160'
  tmux send-keys -t dev-pm:1.{N} Enter
  ```
- 送信前に `tmux list-panes -t dev-pm:1` で対象ペインの ID を確認する
- 受信側が処理中の場合、メッセージはプロンプトにキューされ、現在のターン完了後に処理される
- tmux 通知はベストエフォート。**重要な状態遷移はステータスファイルを正とする**
- PM が通知を見逃した場合でも、ステータスファイルを読めば正確な状態を把握できる

## Recovery & Error Handling

### PM 起動時のリカバリ

PM Agent 起動時は全ステータスファイルを読み込み、現在の状態を復元する:

1. 各 `develop-{N}.json` を読み込み
2. `review_requested` のスロットがあればレビューを再開
3. `working` のスロットは Dev Agent が稼働中と判断し待機
4. `assigned` で `started_at` が古い (2 時間以上) 場合はユーザーに警告

### Dev Agent の異常終了

- `working` 状態で `updated_at` が 2 時間以上更新されていない場合、PM はユーザーに「slot {N} が停滞している可能性があります」と報告
- ユーザーが判断して手動で Dev Agent を再起動するか、PM に `idle` リセットを指示する
- PM は自律的にスロットをリセットしない (進行中の作業を失う可能性があるため)

### tmux pane が無応答の場合

- PM が `tmux send-keys` 後にステータスファイルが更新されない場合、ユーザーに pane の状態確認を依頼
- PM が Dev Agent の代わりにステータスファイルを強制更新することはしない

## Worktree Lifecycle

### 前提

worktree は `make worktree BRANCH=develop-{N}` で事前に作成済みとする (develop-1〜4)。各 worktree は `apps/shell/setup-worktree-env.sh` により固有のポートオフセットを持つ。

### タスク間のブランチ管理

```
issue 完了 (done) → PM が次の issue を割り当てる前に:
  1. Dev Agent: 現在のブランチの作業が全て push 済みであることを確認
  2. Dev Agent: main を fetch & checkout
  3. Dev Agent: 新しい feature ブランチを main から作成
  4. Dev Agent: docker compose down → up で環境リセット
```

### worktree の作成・削除

- worktree の作成・削除は PM のスコープ外 (ユーザーが `make worktree` / `make worktree-rm` で管理)
- PM は空きスロットに issue を割り当てるだけ

## PM Agent Workflows

### 1. Epic Decomposition

```
ユーザー: 「エピック #155 を分解して」
  ↓
PM: エピック issue 読み込み + コードベース調査
  ↓
PM: 分解案を提示
  - 各 issue: タイトル、スコープ、依存関係、Area、Size
  - feature flag が必要な変更を特定
  ↓
[ユーザー承認]
  ↓
PM: gh issue create × N → Project Board 追加 → Priority/Area/DependsOn 設定
```

**分解基準:**

- 1 issue = 半日〜1日で完了する単位
- レイヤー分割を優先 (DB → domain/usecase → handler → frontend)
- 依存関係を明示 (直列 / 並列)
- feature flag で保護すべき変更を特定

### 2. Issue Selection & Assignment

```
PM: ボード分析 (Ready × 優先順位 × 依存関係 × 空きスロット)
  ↓
PM: 割り当て案を提示
  ↓
[ユーザー承認]
  ↓
PM: ステータスファイル更新 (idle → assigned)
PM: tmux send-keys で Dev Agent に指示
PM: Project Board Status → In Progress
  ↓
PM: 入力待ち (Dev からの通知を待つ)
```

**選定基準:**

1. 依存関係が解決済み (DependsOn の参照先が全て Done)
2. Priority 高い順 (P0 > P1 > P2)
3. 並列実行可能な組み合わせを優先
4. In Progress は最大 4 件 (develop 1〜4 スロットに対応)。空きスロット分だけ割り当てる

### 3. Code Review & Completion

```
Dev: tmux send-keys で PM に完了通知
  ↓
PM: ステータスファイル確認 → PR の diff 取得
  ↓
PM: コードレビュー実施
  - アーキテクチャ適合 (レイヤー依存方向、domain の独立性)
  - コード品質 (エラーハンドリング、命名、重複)
  - セキュリティ (SQL injection, 入力バリデーション)
  - 既存パターンとの一貫性
  ↓
PM: レビュー結果をユーザーに提示
  ↓
[ユーザー承認]
  ↓
(OK)  PM: approved → gh pr merge → post-merge スキル実行 → done → 次の issue 選定へ
(NG)  PM: revision_requested → tmux で Dev にフィードバック → 入力待ち
```

### Approval Gates

| タイミング     | ユーザーに提示する内容                |
| -------------- | ------------------------------------- |
| エピック分解後 | issue 一覧 + 依存関係                 |
| issue 選定時   | スロット × issue の割り当て案         |
| レビュー完了後 | レビュー結果 + approve / 差し戻し判断 |
| PR マージ前    | PR リンク + マージ可否                |

## Dev Engineer Workflow

### 受信 → 実行 → 報告

```
PM から /dev-assign 受信
  ↓
ステータスファイル更新 (assigned → working)
  ↓
gh issue view で issue 読み込み
  ↓
plan-issue で実装プラン作成
  ↓
worktree 上で実装
  ↓
ビルド確認 (CGO_ENABLED=1 go build ./...)
  ↓
docker compose up → health check → E2E テスト
  ↓
commit-push で push
  ↓
create-pr で PR 作成 (Closes #N)
  ↓
ステータスファイル更新 (working → review_requested)
  ↓
tmux send-keys で PM に通知
```

### 差し戻し時

```
PM から /dev-revise 受信
  ↓
ステータスファイル更新 (revision_requested → working)
  ↓
フィードバック内容に基づき修正
  ↓
再 push → ステータス → review_requested → PM に通知
```

## File Structure

### New Files

```
.claude/
  agents/
    dev-pm.md              # PM エージェント定義
    dev-engineer.md        # 開発エージェント定義
  dev-pm/
    status/                # ステータスファイル (gitignore)
      develop-1.json
      develop-2.json
      develop-3.json
      develop-4.json

docs/superpowers/specs/
  2026-03-21-dev-pm-agent-design.md  # 本ドキュメント
```

### Modified Files

| ファイル                             | 変更内容                                    |
| ------------------------------------ | ------------------------------------------- |
| `.gitignore`                         | `.claude/dev-pm/status/` 追加               |
| `.claude/skills/project/SKILL.md`    | WIP 制限 2 → 4                              |
| `.claude/skills/plan-issue/SKILL.md` | WIP 制限 2 → 4                              |
| `.claude/skills/post-merge/SKILL.md` | WIP 制限 2 → 4                              |
| `CLAUDE.md`                          | WIP 制限 2 → 4、dev-pm ワークフロー概要追記 |

### Existing Files Relationship

| 既存ファイル                           | dev-pm での扱い                                       |
| -------------------------------------- | ----------------------------------------------------- |
| `.claude/agents/dp-development.md`     | 残す (手動開発用)。dev-engineer.md は下記の原則を継承 |
| `.claude/skills/project/SKILL.md`      | PM が内部で利用 (ボード操作)                          |
| `.claude/skills/create-issue/SKILL.md` | PM がエピック分解時に利用                             |
| `.claude/skills/plan-issue/SKILL.md`   | Dev が実装プラン作成時に利用                          |
| `.claude/skills/create-pr/SKILL.md`    | Dev が PR 作成時に利用                                |
| `.claude/skills/post-merge/SKILL.md`   | PM がレビュー承認後に利用                             |
| `.claude/skills/commit-push/SKILL.md`  | Dev が push 時に利用                                  |

## dev-engineer と dp-development の関係

### 継承する原則 (dp-development → dev-engineer)

- Go レイヤードアーキテクチャ (依存方向、パッケージ構成)
- Go コーディング規約 (CGO, エラーハンドリング, HTTP ハンドラ)
- SQLite 規約 (マイグレーション命名、PRAGMA)
- Pipeline パターン (events / uploads)
- 検証コマンド (go build, docker compose, E2E)
- セキュリティ規約

### 置き換える振る舞い

| dp-development                         | dev-engineer                                  |
| -------------------------------------- | --------------------------------------------- |
| ユーザーとの直接対話 (AskUserQuestion) | PM Agent への報告 (tmux + ステータスファイル) |
| ユーザーにプラン確認を求める           | プランは自律実行 (issue に方針が記載済み)     |
| Project Board の Status 更新           | PM が一元管理 (Dev は更新しない)              |
| 自律的な issue 選定                    | PM から `/dev-assign` で指示を受ける          |

## Project Board Updates

### WIP 制限変更

- 現行: In Progress 最大 2 件 (API 系 1 + 非 API 系 1)
- 変更: In Progress **最大 4 件** (develop 1〜4 に対応)

### Label 管理

既存ラベル (bug, enhancement, etc.) に加え、エピック分解で必要に応じてラベルを追加する。PM がラベルの妥当性を判断する。

## Future Extensions

以下は初期スコープ外。運用しながら段階的に追加する。

- **QA エージェント:** 受け入れテスト専任。ステータスに `qa_requested` / `qa_passed` を追加
- **待機中の有効活用:** ポーリング待機 (定期的にステータスファイル確認)、空き時間での別スロットレビューやボード整理
- **通知チャネル抽象化:** Discord / Slack 連携
- **自動マージ:** レビュー承認後のマージを承認ゲートなしで実行するオプション
