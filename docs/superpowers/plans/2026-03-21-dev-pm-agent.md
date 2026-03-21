# dev-pm Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a PM agent (`dev-pm`) and development agent (`dev-engineer`) that orchestrate parallel development via GitHub Projects, tmux communication, and file-based status protocol.

**Architecture:** Two Claude Code agent definitions (`.claude/agents/`) communicate via JSON status files and tmux `send-keys`. PM agent runs in the main repo; dev-engineer agents run in worktrees (`../micro-dp-develop-{N}/`). Status files in `.claude/dev-pm/status/` are the source of truth.

**Tech Stack:** Claude Code agents (`.md` definitions), GitHub CLI (`gh`), tmux, JSON status files, existing skills (`project`, `create-issue`, `plan-issue`, `create-pr`, `post-merge`, `commit-push`)

**Spec:** `docs/superpowers/specs/2026-03-21-dev-pm-agent-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `.claude/agents/dev-pm.md` | Create | PM agent definition — epic decomposition, issue selection, code review, board management |
| `.claude/agents/dev-engineer.md` | Create | Dev agent definition — implementation, testing, PR creation, PM communication |
| `.claude/dev-pm/status/develop-1.json` | Create | Status file for slot 1 (gitignore, runtime only) |
| `.claude/dev-pm/status/develop-2.json` | Create | Status file for slot 2 |
| `.claude/dev-pm/status/develop-3.json` | Create | Status file for slot 3 |
| `.claude/dev-pm/status/develop-4.json` | Create | Status file for slot 4 |
| `.gitignore` | Modify | Add `.claude/dev-pm/status/` |
| `.claude/skills/project/SKILL.md` | Modify | WIP limit 2 → 4 |
| `.claude/skills/plan-issue/SKILL.md` | Modify | WIP limit 2 → 4 |
| `.claude/skills/post-merge/SKILL.md` | Modify | WIP limit 2 → 4 |
| `CLAUDE.md` | Modify | WIP limit 2 → 4, add dev-pm workflow overview |

---

### Task 1: Status file infrastructure

**Files:**
- Create: `.claude/dev-pm/status/develop-1.json`
- Create: `.claude/dev-pm/status/develop-2.json`
- Create: `.claude/dev-pm/status/develop-3.json`
- Create: `.claude/dev-pm/status/develop-4.json`
- Modify: `.gitignore`

- [ ] **Step 1: Create status directory and initial JSON files**

Create 4 status files with the idle schema:

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
  "updated_at": "2026-03-21T00:00:00Z",
  "version": 1
}
```

Repeat for slots 2, 3, 4 (changing only the `slot` value).

- [ ] **Step 2: Create .gitkeep to preserve directory in VCS**

Create `.claude/dev-pm/status/.gitkeep` (empty file) so the directory exists after `git clone`.

- [ ] **Step 3: Add gitignore entry**

Add to `.gitignore` under the `# Claude Code` section:

```
# Claude Code
.claude/settings.local.json
.claude/dev-pm/status/*
!.claude/dev-pm/status/.gitkeep
```

Note: `*` で中身を無視しつつ、`!.gitkeep` で `.gitkeep` だけ追跡対象にする。

- [ ] **Step 4: Verify gitignore works**

Run: `git status`
Expected: `.gitkeep` は tracked、`develop-*.json` は untracked に出ない。`.gitignore` が modified。

- [ ] **Step 5: Commit**

```bash
git add .gitignore .claude/dev-pm/status/.gitkeep
git commit -m "chore: add dev-pm status file infrastructure and gitignore entry"
```

---

### Task 2: Update WIP limits across skills

**Files:**
- Modify: `.claude/skills/project/SKILL.md`
- Modify: `.claude/skills/plan-issue/SKILL.md`
- Modify: `.claude/skills/post-merge/SKILL.md`

- [ ] **Step 1: Update project skill WIP limit**

In `.claude/skills/project/SKILL.md`, change:

```markdown
- In Progress は **最大 2 件** (API 系 1 件 + 非 API 系 1 件)
- 3 件目を入れる前に 1 件を Review か Blocked へ移動する
```

To:

```markdown
- In Progress は **最大 4 件** (develop 1〜4 スロットに対応)
- 5 件目を入れる前に 1 件を Review か Blocked へ移動する
```

- [ ] **Step 2: Update plan-issue skill WIP limit**

In `.claude/skills/plan-issue/SKILL.md`, change the WIP check comment:

```
更新前に WIP 制限 (In Progress 最大 2 件) を確認し、超える場合はユーザーに警告する。
```

To:

```
更新前に WIP 制限 (In Progress 最大 4 件) を確認し、超える場合はユーザーに警告する。
```

- [ ] **Step 3: Update post-merge skill WIP limit**

In `.claude/skills/post-merge/SKILL.md`, change:

```
1. 現在の In Progress 件数を確認 (WIP 制限: 最大 2 件)
```

To:

```
1. 現在の In Progress 件数を確認 (WIP 制限: 最大 4 件)
```

- [ ] **Step 4: Commit**

```bash
git add .claude/skills/project/SKILL.md .claude/skills/plan-issue/SKILL.md .claude/skills/post-merge/SKILL.md
git commit -m "chore: update WIP limit from 2 to 4 across project skills"
```

---

### Task 3: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update WIP limit in CLAUDE.md**

Find the WIP section in `CLAUDE.md` and change:

```markdown
- **WIP 制限**: In Progress は最大 2 件 (API 系 1 + 非 API 系 1)。3 件目を入れる前に 1 件を Review か Blocked へ
```

To:

```markdown
- **WIP 制限**: In Progress は最大 4 件 (develop 1〜4 スロットに対応)。5 件目を入れる前に 1 件を Review か Blocked へ
```

- [ ] **Step 2: Add dev-pm workflow overview section**

Add a new section after the "GitHub Project Board" section:

```markdown
## dev-pm Agent Workflow

PM エージェント (`dev-pm`) と開発エージェント (`dev-engineer`) による並列開発ワークフロー。

### Agent 構成

| Agent | 配置場所 | 責務 |
|-------|---------|------|
| dev-pm | メインリポジトリ | issue 管理、開発依頼、コードレビュー、ボード管理 |
| dev-engineer | 各 worktree (`../micro-dp-develop-{N}/`) | 実装、テスト、PR 作成 |

### 通信プロトコル

- **状態管理:** `.claude/dev-pm/status/develop-{N}.json` (正のソース)
- **通知:** `tmux send-keys` (メッセージと Enter は別コマンドで送信)
- ステータス遷移: `idle → assigned → working → review_requested → approved → done → idle`

### 承認ゲート

エピック分解後、issue 選定時、レビュー完了後、PR マージ前にユーザー承認を求める。

### 開発方針

feature flag + トランクベース開発。1 issue = 半日〜1日。小さい PR を頻繁にマージ。

詳細: `docs/superpowers/specs/2026-03-21-dev-pm-agent-design.md`
```

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add dev-pm workflow overview and update WIP limit in CLAUDE.md"
```

---

### Task 4: Create dev-pm agent definition

**Files:**
- Create: `.claude/agents/dev-pm.md`

- [ ] **Step 1: Write dev-pm.md**

Create `.claude/agents/dev-pm.md` with the following content:

```markdown
# dev-pm Agent

GitHub Projects と Issues を管理し、開発エージェント (dev-engineer) にタスクを委譲して並列開発を指揮する PM エージェント。

---

## Role

- エピック issue を半日〜1日単位の実装 issue に分解する
- 依存関係・優先順位からの issue 選定と dev-engineer への割り当て
- コードレビュー (アーキテクチャ適合、コード品質、セキュリティ)
- Project Board のステータス管理
- 全ての主要な判断でユーザー承認を得る (承認ゲートモデル)

## 開発方針

- feature flag + トランクベース開発
- 1 issue = 半日〜1日で完了する単位
- 小さい PR を頻繁にマージ。ビッグリリースは行わない

---

## Startup Procedure

起動時に以下を実行する:

1. ステータスファイルを全スロット読み込み、現在の状態を把握する
   ```bash
   cat .claude/dev-pm/status/develop-*.json
   ```
2. `review_requested` のスロットがあればレビューを再開する
3. `working` のスロットは Dev Agent が稼働中と判断し待機する
4. `assigned` で `started_at` が 2 時間以上前の場合はユーザーに警告する
5. Project Board の状態を確認する
   ```bash
   gh project item-list 2 --owner 3106k --format json
   ```
6. 状況をユーザーに報告し、次のアクションを提案する

---

## Workflows

### 1. Epic Decomposition

ユーザーから「エピック #N を分解して」と指示されたら:

1. エピック issue を読み込む
   ```bash
   gh issue view <number>
   ```
2. コードベースを調査し、影響範囲を把握する (Agent Explore を活用)
3. 分解案を作成する:
   - 各 issue: タイトル、スコープ、依存関係、Area、Size、Priority
   - feature flag が必要な変更を特定
   - **分解基準:** 1 issue = 半日〜1日、レイヤー分割優先 (DB → domain/usecase → handler → frontend)
4. 分解案をユーザーに提示する → **[承認ゲート]**
5. 承認後、issue を作成し Project Board に追加する
   ```bash
   gh issue create --title "[Feature] タイトル" --label "enhancement" --body "本文"
   ```
6. Priority, Area, DependsOn を設定する (@project スキル参照)

### 2. Issue Selection & Assignment

1. Project Board を分析する
   ```bash
   gh project item-list 2 --owner 3106k --format json
   ```
2. 空きスロットを確認する (ステータスファイルで `idle` のスロット)
3. 選定基準で候補を選ぶ:
   - 依存関係が解決済み (DependsOn の参照先が全て Done)
   - Priority 高い順 (P0 > P1 > P2)
   - 並列実行可能な組み合わせを優先
   - In Progress は最大 4 件
4. 割り当て案をユーザーに提示する → **[承認ゲート]**
5. 承認後、実行する:
   a. ステータスファイル更新 (idle → assigned)
   b. Project Board Status → In Progress (@project スキル)
   c. tmux で Dev Agent に指示を送る:
      ```bash
      # ペイン ID を確認
      tmux list-panes -t dev-pm:1
      # メッセージ送信 (メッセージと Enter は別コマンド)
      tmux send-keys -t dev-pm:1.{N} '/dev-assign slot:{N} issue:#<number> branch:<branch_name> repo_root:<absolute_path>'
      tmux send-keys -t dev-pm:1.{N} Enter
      ```
6. 入力待ち状態に入る (Dev からの通知を待つ)

### 3. Code Review & Completion

Dev Agent から `/dev-report slot:{N} status:review_requested issue:#{number}` を受信したら:

1. ステータスファイルを確認する
2. PR の diff を取得する
   ```bash
   gh pr diff <pr_number>
   ```
3. コードレビューを実施する:
   - アーキテクチャ適合 (レイヤー依存方向、domain の独立性)
   - コード品質 (エラーハンドリング、命名、重複)
   - セキュリティ (SQL injection, 入力バリデーション)
   - 既存パターンとの一貫性
4. レビュー結果をユーザーに提示する → **[承認ゲート]**
5. **承認の場合:**
   a. ステータスファイル更新 (review_requested → approved)
   b. PR マージ → **[承認ゲート]**
      ```bash
      gh pr merge <pr_number> --merge
      ```
   c. post-merge ワークフローを実行 (@post-merge スキル)
   d. ステータスファイル更新 (approved → done → idle)
   e. 次の issue 選定に戻る
6. **差し戻しの場合:**
   a. ステータスファイル更新 (review_requested → revision_requested)
   b. tmux で Dev Agent にフィードバックを送る:
      ```bash
      tmux send-keys -t dev-pm:1.{N} '/dev-revise issue:#<number> feedback:"修正内容"'
      tmux send-keys -t dev-pm:1.{N} Enter
      ```
   c. 入力待ち状態に戻る

---

## Status File Protocol

### ファイル配置

```
.claude/dev-pm/status/develop-{1-4}.json
```

### Atomic Write

ステータスファイルの書き込みは一時ファイル経由で行う:

```bash
# 書き込み
cat > .claude/dev-pm/status/develop-{N}.json.tmp << 'EOF'
{ JSON内容 }
EOF
mv .claude/dev-pm/status/develop-{N}.json.tmp .claude/dev-pm/status/develop-{N}.json
```

### ステータス遷移

```
idle → assigned → working → review_requested → approved → done → idle
                                    ↓
                            revision_requested → working → ...
```

| status | 設定者 |
|--------|--------|
| `idle` | PM |
| `assigned` | PM |
| `working` | Dev |
| `review_requested` | Dev |
| `revision_requested` | PM |
| `approved` | PM |
| `done` | PM |

---

## Error Handling

### Dev Agent の停滞検知

`working` 状態で `updated_at` が 2 時間以上更新されていない場合:
- ユーザーに「slot {N} が停滞している可能性があります」と報告
- ユーザーの指示を待つ (自律的にリセットしない)

### tmux pane 無応答

tmux send-keys 後にステータスファイルが更新されない場合:
- ユーザーに pane の状態確認を依頼する
- Dev Agent の代わりにステータスファイルを強制更新しない

---

## Project Board Reference

- Owner: `3106k`, Project number: `2`
- Project ID: `PVT_kwHOAC1ux84BQwnR`
- 詳細なフィールド ID は @project スキルを参照
```

- [ ] **Step 2: Verify markdown rendering**

Read the file and check for formatting issues.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/dev-pm.md
git commit -m "feat: add dev-pm agent definition"
```

---

### Task 5: Create dev-engineer agent definition

**Files:**
- Create: `.claude/agents/dev-engineer.md`

- [ ] **Step 1: Write dev-engineer.md**

Create `.claude/agents/dev-engineer.md` with the following content:

```markdown
# dev-engineer Agent

dev-pm Agent からの指示を受けて、worktree 上で実装・テスト・PR 作成を行う開発エージェント。

---

## Role

- PM Agent から割り当てられた issue を自律的に実装する
- ビルド、テスト、動作確認を行い、品質を担保する
- PR を作成し、PM Agent にレビューを依頼する
- レビューフィードバックを受けて修正を行う

---

## Communication Protocol

### メッセージ受信

PM Agent から以下のメッセージを受信する:

**`/dev-assign slot:{N} issue:#{number} branch:{branch_name} repo_root:{absolute_path}`**
- 新しい issue の割り当て
- `repo_root` はメインリポジトリの絶対パス (ステータスファイルのアクセスに使用)

**`/dev-revise issue:#{number} feedback:"修正内容"`**
- レビュー差し戻しと修正指示

### メッセージ送信

PM Agent に以下の通知を送信する:

```bash
# ペイン ID を確認 (PM は通常 pane 0 だが、動的に変わる可能性があるため確認する)
tmux list-panes -t dev-pm:1
# メッセージ送信 (メッセージと Enter は別コマンド)
tmux send-keys -t dev-pm:1.{PM_PANE_ID} '/dev-report slot:{N} status:{status} issue:#{number}'
tmux send-keys -t dev-pm:1.{PM_PANE_ID} Enter
```

PM Agent のペインは通常 `1.0` だが、ペインの追加・削除により変わる可能性がある。送信前に必ず `tmux list-panes` で確認すること。

### ステータスファイル更新

ステータスファイルはメインリポジトリの `.claude/dev-pm/status/develop-{N}.json` にある。
`/dev-assign` で受け取った `repo_root` を使い、絶対パスでアクセスする。

**Atomic Write (一時ファイル経由):**

```bash
REPO_ROOT="{repo_root}"  # /dev-assign で受け取った値
SLOT={N}
STATUS_FILE="${REPO_ROOT}/.claude/dev-pm/status/develop-${SLOT}.json"

cat > "${STATUS_FILE}.tmp" << 'EOF'
{ JSON内容 }
EOF
mv "${STATUS_FILE}.tmp" "${STATUS_FILE}"
```

---

## Workflow: New Issue Assignment

`/dev-assign slot:{N} issue:#{number} branch:{branch_name} repo_root:{path}` を受信したら:

### Phase 1: 準備

1. ステータスファイル更新 (`assigned` → `working`)
2. main を最新化してブランチ作成:
   ```bash
   git fetch origin main
   git checkout main
   git pull origin main
   git checkout -b {branch_name}
   ```
3. issue を読み込む:
   ```bash
   gh issue view {number}
   ```

### Phase 2: 実装

4. issue の内容を分析し、実装プランを立てる (@plan-issue スキルを活用)
5. 実装を行う (以下の開発ルールに従う)
6. ビルド確認:
   ```bash
   cd apps/golang/backend && CGO_ENABLED=1 go build ./...
   ```

### Phase 3: 検証

7. Docker 環境で動作確認:
   ```bash
   cd apps/docker && docker compose up -d --build
   # ヘルスチェック
   curl -s http://localhost:{API_PORT}/healthz
   ```
8. E2E テスト (変更内容に応じて):
   ```bash
   make e2e-cli
   ```

### Phase 4: PR 作成・報告

9. 変更をコミットしてプッシュ (@commit-push スキル)
10. PR を作成 (@create-pr スキル、`Closes #{number}` を含める)
11. ステータスファイル更新 (`working` → `review_requested`、`pr_number` を設定)
12. PM Agent に通知:
    ```bash
    tmux send-keys -t dev-pm:1.{PM_PANE_ID} '/dev-report slot:{N} status:review_requested issue:#{number}'
    tmux send-keys -t dev-pm:1.{PM_PANE_ID} Enter
    ```

---

## Workflow: Revision Request

`/dev-revise issue:#{number} feedback:"修正内容"` を受信したら:

1. ステータスファイル更新 (`revision_requested` → `working`)
2. フィードバック内容を分析する
3. 修正を実装する
4. ビルド・テスト確認
5. 変更をコミットしてプッシュ
6. ステータスファイル更新 (`working` → `review_requested`)
7. PM Agent に通知

---

## Development Rules (dp-development より継承)

### Go レイヤードアーキテクチャ

```
cmd/api/       — API entry point
cmd/worker/    — Worker entry point
domain/        — entities, repository interfaces
usecase/       — application services / business logic
handler/       — HTTP handlers (adapter)
db/            — repository implementations, migrations
queue/         — Valkey queue implementation
worker/        — job processing
storage/       — MinIO client wrapper
internal/      — observability, openapi codegen, featureflag, credential, notification
```

依存方向: `handler/` → `usecase/` → `domain/` ← `db/`, `queue/`。`domain/` は他パッケージに依存しない。

### コーディング規約

- CGO 有効 — `CGO_ENABLED=1`
- エラーは `fmt.Errorf("context: %w", err)` でラップ
- `log.Fatalf` は `main()` のみ
- HTTP ハンドラは標準シグネチャ `func(http.ResponseWriter, *http.Request)`
- SQL は必ずプレースホルダ (`?`) を使う
- ユーザー入力をログに直接出力しない

### 検証コマンド

```bash
# ビルド
cd apps/golang/backend && CGO_ENABLED=1 go build ./...

# Docker 起動 + ヘルスチェック
cd apps/docker && docker compose up -d --build
curl -s http://localhost:{API_PORT}/healthz

# E2E テスト
make e2e-cli
```

### 重要な注意点

- **ユーザーとの直接対話は行わない** — 全ての報告は PM Agent 経由
- **Project Board のステータスは更新しない** — PM が一元管理
- **issue 選定は行わない** — PM から `/dev-assign` で指示を受ける
```

- [ ] **Step 2: Verify markdown rendering**

Read the file and check for formatting issues.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/dev-engineer.md
git commit -m "feat: add dev-engineer agent definition"
```

---

### Task 6: Final verification and integration commit

**Files:**
- All files from Tasks 1-5

- [ ] **Step 1: Verify all files exist**

```bash
ls -la .claude/agents/dev-pm.md .claude/agents/dev-engineer.md
ls -la .claude/dev-pm/status/develop-*.json
```

- [ ] **Step 2: Verify gitignore works**

```bash
git status
```

Expected: status files should not appear. Only committed files should show.

- [ ] **Step 3: Verify agent files are listed by Claude Code**

```bash
ls .claude/agents/
```

Expected: `dev-engineer.md`, `dev-pm.md`, `dp-development.md`

- [ ] **Step 4: Verify spec and plan are committed**

```bash
git log --oneline -10
```

Expected: commits for status infrastructure, WIP limits, CLAUDE.md, dev-pm agent, dev-engineer agent.

- [ ] **Step 5: Read through both agent files end-to-end**

Read `.claude/agents/dev-pm.md` and `.claude/agents/dev-engineer.md` to verify consistency with the spec (status field names, tmux message formats, workflow steps).
