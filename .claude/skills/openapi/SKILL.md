---
name: openapi
description: Run the Contract-First OpenAPI workflow — lint, generate types, and verify build
disable-model-invocation: true
allowed-tools: Bash, Read, Edit, Grep, Glob
argument-hint: "[--check]"
---

# OpenAPI Skill

Contract-First OpenAPI ワークフローを実行する。

`$ARGUMENTS` に `--check` を含む場合は drift detection (git diff) も行う。

## Current state

- Spec: !`head -5 spec/openapi/v1.yaml`
- Git status: !`git diff --stat spec/openapi/`

## Steps

### 1. Lint

```bash
make openapi-lint
```

lint エラーがあれば報告して停止する。

### 2. Generate

FE (TypeScript) と BE (Go) の型を再生成する:

```bash
make openapi-generate
```

### 3. Go build

生成型が handler/usecase と整合するか確認:

```bash
cd apps/golang/backend && CGO_ENABLED=1 go build ./...
```

ビルドエラーがあれば、生成型と既存コードの不整合箇所を特定して報告する。

### 4. Drift check (--check のみ)

`$ARGUMENTS` に `--check` が含まれる場合:

```bash
git diff --exit-code
```

差分があれば「spec と生成コードが同期していない」旨を報告する。

### 5. Report

結果をサマリ報告する:

| Step | Result |
|------|--------|
| Lint | pass/fail |
| Generate | done |
| Go build | pass/fail |
| Drift check | clean / diff found / skipped |

生成されたファイル:
- `apps/node/web/src/lib/api/generated.ts`
- `apps/golang/backend/internal/openapi/*.gen.go`
