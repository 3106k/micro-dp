---
name: verify
description: Build, deploy, and run E2E tests to verify the current codebase
allowed-tools: Bash, Read, Grep
argument-hint: "[--skip-e2e] [--skip-docker]"
---

# Verify Skill

実装後の一連の検証を実行する。

- `--skip-e2e`: E2E テストをスキップ
- `--skip-docker`: Docker rebuild + health check をスキップ（ローカルビルド確認のみ）

## Current state

- Working directory: !`pwd`
- Git branch: !`git branch --show-current`

## Steps

### 1. OpenAPI check

spec と生成コードの同期を確認する:

```bash
make openapi-check
```

差分がある場合は `make openapi-generate` の実行を促して停止する。

### 2. Go build

backend と e2e-cli の両方をビルドする:

```bash
cd apps/golang/backend && CGO_ENABLED=1 go build ./...
cd apps/golang/e2e-cli && go build ./...
```

ビルドが失敗した場合はエラー内容を報告して停止する。

### 3. Frontend build

Next.js のビルドを実行する:

```bash
cd apps/node/web && npm run build
```

型エラーやビルドエラーがあれば報告して停止する。

### 4. Docker rebuild + health check

`$ARGUMENTS` に `--skip-docker` が含まれていなければ実行:

```bash
make down && make up
```

全コンテナが healthy になるまで待機してから:

```bash
make health
```

いずれかが unhealthy の場合は `docker logs` でエラーを確認して報告する。

### 5. Startup log verification

`--skip-docker` でなければ、API と Worker のログから初期化メッセージを確認する:

```bash
docker logs $(docker ps -qf name=-api) 2>&1 | grep -E "feature flags|observability|api server"
docker logs $(docker ps -qf name=-worker) 2>&1 | grep -E "feature flags|observability|worker starting"
```

### 6. E2E tests

`$ARGUMENTS` に `--skip-e2e` が含まれていなければ実行:

```bash
make e2e-cli
```

### 7. Report

全ステップの結果を以下のフォーマットでサマリ報告する:

| Step | Result |
|------|--------|
| OpenAPI check | pass/fail/skipped |
| Go build | pass/fail |
| Frontend build | pass/fail |
| Docker health | pass/fail/skipped |
| Startup logs | OK / issues found / skipped |
| E2E tests | N passed, N failed, N skipped / skipped |
