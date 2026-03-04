---
name: sync-main
description: Fetch and merge origin/main into the current branch
disable-model-invocation: true
allowed-tools: Bash, Read
---

# Sync Main Skill

origin/main の最新を取得し、現在のブランチにマージする。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Uncommitted changes: !`git status --short`

## Steps

### 1. 未コミット変更の確認

```bash
git status --short
```

未コミットの変更がある場合は一覧を表示し、**マージ前にコミットまたは stash を推奨** する旨を報告して停止する。

### 2. main を fetch

```bash
git fetch origin main
```

### 3. 差分確認

```bash
git log HEAD..origin/main --oneline
```

取り込まれるコミットがなければ「Already up to date」と報告して終了する。

### 4. マージ

```bash
git merge origin/main --no-edit
```

### 5. コンフリクト処理

コンフリクトが発生した場合:

1. `git diff --name-only --diff-filter=U` でコンフリクトファイルを一覧表示
2. 各ファイルのコンフリクト箇所を報告
3. **自動解決は行わず**、ユーザーに判断を委ねる

### 6. 結果報告

| Item | Result |
|------|--------|
| Branch | 現在のブランチ名 |
| Merged commits | 取り込んだコミット数 |
| Conflicts | なし / ファイル一覧 |
| Status | success / conflict (要手動解決) |
