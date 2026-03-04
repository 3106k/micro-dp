---
name: Feature Implementation
about: 新機能実装・拡張のIssueテンプレート
title: "[Feature] "
labels: enhancement
assignees: ""
---

## 背景
- なぜこの機能が必要か
- 現状の課題

## 目的
- このIssueで達成したいこと

## スコープ
### In
- 

### Out
- 

## 仕様メモ（任意）
- 認証方式
- データモデル
- UI/UX 方針

## API / Contract 影響
- [ ] OpenAPI 変更なし
- [ ] OpenAPI 変更あり（`spec/openapi/v1.yaml` を更新）

OpenAPI変更がある場合:
- [ ] `make openapi-lint`
- [ ] `make openapi-generate`
- [ ] `make openapi-check`

## 実装タスク
- [ ] Backend
- [ ] Frontend
- [ ] Test（unit/integration/e2e）
- [ ] Docs（必要な場合）

## 受け入れ条件
- [ ] ユーザー操作で期待フローが完了する
- [ ] 主要エラーケースで原因が識別できる
- [ ] 既存機能の回帰がない

## 依存関係
- Depends on: #
- Blocks: #

## 補足
- 関連Issue / 参考リンク / メモ
