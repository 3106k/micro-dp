---
name: verify
description: Build, deploy, and run E2E tests to verify the current codebase
disable-model-invocation: true
allowed-tools: Bash, Read, Grep
argument-hint: "[--skip-e2e]"
---

# Verify Skill

実装後の一連の検証を実行する。引数に `--skip-e2e` を渡すと E2E テストをスキップする。

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`

## Steps

### 1. Go build

```bash
cd apps/golang/backend && CGO_ENABLED=1 go build ./...
```

ビルドが失敗した場合はエラー内容を報告して停止する。

### 2. Docker rebuild + health check

```bash
make down && make up
```

全コンテナが healthy になるまで待機してから:

```bash
make health
```

いずれかが unhealthy の場合は `docker logs` でエラーを確認して報告する。

### 3. Startup log verification

API と Worker のログから初期化メッセージを確認する:

```bash
docker logs $(docker ps -qf name=-api) 2>&1 | grep -E "feature flags|observability|api server"
docker logs $(docker ps -qf name=-worker) 2>&1 | grep -E "feature flags|observability|worker starting"
```

### 4. E2E tests

`$ARGUMENTS` に `--skip-e2e` が含まれていなければ実行:

```bash
make e2e-cli
```

### 5. Report

全ステップの結果を以下のフォーマットでサマリ報告する:

| Step | Result |
|------|--------|
| Go build | pass/fail |
| Health check | pass/fail |
| Startup logs | OK / issues found |
| E2E tests | N passed, N failed, N skipped / skipped |
