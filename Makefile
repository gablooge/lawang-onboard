.DEFAULT_GOAL := check
.PHONY: check secrets corpus test vet hooks up down logs ps

# What CI runs.
check: secrets vet test

# Fails if any secret-looking string is in the working tree or history.
secrets:
	gitleaks detect --source . --redact --no-banner

# Rebuilds a corpus from a local clone (read-only). The defaults rebuild the Lawang demo corpus.
# Your own repository: make corpus REPO=../myrepo GITHUB=me/myrepo OUT=local/corpus
# GITHUB= (empty) exports files and commits only, with no network call.
REPO ?= ../lawang
GITHUB ?= gablooge/lawang
OUT ?= corpus
corpus:
	./scripts/export_corpus.py "$(REPO)" $(GITHUB) --out "$(OUT)"

vet:
	@if [ -f go.mod ]; then go vet ./...; else echo "no go.mod yet"; fi

test:
	@if [ -f go.mod ]; then go test -race -count=1 ./...; else echo "no go.mod yet"; fi

# Installs the pre-commit hook that runs `make secrets`.
hooks:
	git config core.hooksPath .githooks

# Runs the server and the Cloudflare Tunnel connector (see docs/deploy.md).
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs --tail=100 -f

ps:
	docker compose ps
