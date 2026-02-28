.PHONY: up down build logs ps health clean setup-env dev-api dev-worker e2e-cli e2e-ci-template openapi-lint openapi-bundle openapi-generate-fe openapi-generate-be openapi-generate openapi-check worktree worktree-rm worktree-ls

COMPOSE_DIR := apps/docker
COMPOSE     := cd $(COMPOSE_DIR) && docker compose

# .env があれば読み込む（ポート変数を Makefile 内で参照するため）
-include .env
export

API_HOST_PORT         ?= 8080
WEB_HOST_PORT         ?= 3000
VALKEY_HOST_PORT      ?= 6379
MINIO_API_HOST_PORT   ?= 9000
MINIO_CONSOLE_HOST_PORT ?= 9001

# --- Docker Compose ---

setup-env:
	./apps/shell/setup-worktree-env.sh

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

health:
	@echo "=== API (localhost:$(API_HOST_PORT)) ===" && curl -sf http://localhost:$(API_HOST_PORT)/healthz && echo
	@echo "=== Web (localhost:$(WEB_HOST_PORT)) ===" && curl -sf http://localhost:$(WEB_HOST_PORT)/api/health && echo
	@echo "=== Valkey ===" && $(COMPOSE) exec -T valkey valkey-cli ping

clean:
	$(COMPOSE) down -v --rmi local

dev-api:
	cd apps/golang/backend && air -c .air.api.toml

dev-worker:
	cd apps/golang/backend && air -c .air.worker.toml

e2e-cli:
	cd apps/golang/e2e-cli && go run ./cmd/run --base-url=http://localhost:$(API_HOST_PORT)

e2e-ci-template:
	$(MAKE) up
	$(MAKE) e2e-cli || (ret=$$?; $(MAKE) down; exit $$ret)
	$(MAKE) down

openapi-lint:
	cd apps/node/web && npm run openapi:lint

openapi-bundle:
	cd apps/node/web && npm run openapi:bundle

openapi-generate-fe:
	cd apps/node/web && npm run openapi:generate

openapi-generate-be:
	cd apps/golang/backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 -config oapi-types.cfg.yaml ../../../spec/openapi/v1.yaml
	cd apps/golang/backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 -config oapi-server.cfg.yaml ../../../spec/openapi/v1.yaml

openapi-generate: openapi-generate-fe openapi-generate-be

openapi-check:
	$(MAKE) openapi-lint
	$(MAKE) openapi-generate
	git diff --exit-code

# --- Git Worktree ---
# 使い方:
#   make worktree BRANCH=feature-x        # ../micro-dp-feature-x/ に作成 + .env 生成
#   make worktree-rm BRANCH=feature-x     # ワークツリー削除
#   make worktree-ls                      # 一覧表示

worktree:
ifndef BRANCH
	$(error BRANCH is required. Usage: make worktree BRANCH=feature-x)
endif
	git worktree add -b $(BRANCH) ../micro-dp-$(BRANCH)
	cd ../micro-dp-$(BRANCH) && ./apps/shell/setup-worktree-env.sh
	@echo ""
	@echo "Worktree ready: ../micro-dp-$(BRANCH)"
	@echo "  cd ../micro-dp-$(BRANCH) && make up"

worktree-rm:
ifndef BRANCH
	$(error BRANCH is required. Usage: make worktree-rm BRANCH=feature-x)
endif
	cd ../micro-dp-$(BRANCH) && $(MAKE) down 2>/dev/null || true
	git worktree remove ../micro-dp-$(BRANCH)
	git branch -d $(BRANCH) 2>/dev/null || echo "INFO: ブランチ $(BRANCH) は手動で削除してください"

worktree-ls:
	@git worktree list
