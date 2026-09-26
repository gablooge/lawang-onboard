.DEFAULT_GOAL := check
.PHONY: check secrets corpus test vet hooks up down logs ps

# What CI runs.
check: secrets vet test

# Fails if any secret-looking string is in the working tree or history.
secrets:
	gitleaks detect --source . --redact --no-banner

# Rebuilds corpus/*.jsonl from a local clone of the Lawang repository (read-only).
corpus:
	./scripts/export_corpus.py ../lawang gablooge/lawang --out corpus

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
