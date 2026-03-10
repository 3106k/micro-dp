---
name: commit-push
description: Stage, commit, and push changes with auto-generated commit message
allowed-tools: Bash, Read, Grep, Glob, AskUserQuestion
argument-hint: "[commit message override]"
---

# Commit & Push Skill

変更をステージング・コミット・プッシュする。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`
- Uncommitted changes: !`git status --short`

## Steps

### 1. 変更の確認

以下を並列で確認する:

```bash
# 未ステージの変更
git status --short

# 差分の内容
git diff
git diff --staged
```

変更がない場合は「コミットする変更がありません」と報告して終了。

### 2. ステージング

変更ファイルを確認し、コミット対象を選定する:

- `.env`, credentials, secrets 系ファイルは除外し警告
- 画像・バイナリなど意図しないファイルは確認を取る
- 関連する変更をまとめてステージング (`git add <files>`)

### 3. コミットメッセージ生成

`$ARGUMENTS` でメッセージが指定されていればそれを使う。

指定がなければ、差分内容から自動生成する:

**メッセージ規則**:
- prefix: `feat:` / `fix:` / `refactor:` / `chore:` / `docs:` / `test:`
- 1行目: 70文字以内で変更内容を簡潔に
- 必要に応じて空行 + 詳細説明

**Co-Authored-By**: 末尾に以下を付与:
```
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

### 4. ユーザー確認

ステージ対象ファイルとコミットメッセージをユーザーに提示し、修正点がないか確認する。

### 5. コミット & プッシュ

```bash
# コミット
git commit -m "$(cat <<'EOF'
メッセージ
EOF
)"

# プッシュ
git push
```

### 6. 既存 PR の確認

現在のブランチに紐づく open PR があれば、その URL を報告する:

```bash
gh pr list --head <branch> --state open --json number,url --jq '.[0] | "#\(.number) \(.url)"'
```

### 7. 結果報告

| Item | Result |
|------|--------|
| Branch | ブランチ名 |
| Commit | ハッシュ (短縮) + メッセージ |
| Files | コミットしたファイル数 |
| Push | success / failed |
| PR | #N URL / なし |
