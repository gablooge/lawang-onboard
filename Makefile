.DEFAULT_GOAL := check
.PHONY: check secrets corpus test vet hooks

# What CI runs.
check: secrets vet test

# Fails if any secret-looking string is in the working tree or history.
secrets:
	gitleaks detect --source . --redact --no-banner

# Rebuilds corpus/*.jsonl from a local clone of the Lawang repository (read-only).
corpus:
	./scripts/export-corpus.sh ../lawang

vet:
	@if [ -f go.mod ]; then go vet ./...; else echo "no go.mod yet"; fi

test:
	@if [ -f go.mod ]; then go test -race -count=1 ./...; else echo "no go.mod yet"; fi

# Installs the pre-commit hook that runs `make secrets`.
hooks:
	git config core.hooksPath .githooks
