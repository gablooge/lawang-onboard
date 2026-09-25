# Corpus samples

One item per kind, text cut to 300 characters. Bob reads this file instead of
`corpus/*.jsonl`, which `.bobignore` hides because the full files are about 2.6 MB and reading
them costs many Bobcoins. Tests still load the real files..

## adr

```json
{
  "bytes": 1824,
  "id": "adr:docs/adr/0001-queries-sqlc.md",
  "kind": "adr",
  "paths": [
    "docs/adr/0001-queries-sqlc.md"
  ],
  "text": "# 1. Typed queries with sqlc, infrastructure SQL by hand\n\nStatus: accepted, 2026-09-19 (backlog item B03)\n\n## Decision\n\nQueries against Lawang's own tables are written as SQL files and compiled to Go with\n[`sqlc`](https://sqlc.dev), targeting `pgx/v5`. The generated code is checked in, and CI fails ...",
  "title": "1. Typed queries with sqlc, infrastructure SQL by hand",
  "truncated": false
}
```

## commit

```json
{
  "author": "gablooge",
  "date": "2026-09-20T17:33:19+07:00",
  "hash": "README.md\ndocs/architecture.md\ndocs/backlog.md\ninternal/config/config.go\ninternal/config/config_test.go\ninternal/ingress/ingress.go\ninternal/ingress/ingress_test.go\ninternal/provider/fake/fake.go\ninternal/provider/fake/fake_test.go\ninternal/provider/provider.go\ninternal/provider/registry.go\ninternal/provider/registry_test.go\n04459d8e0dc0307e6154a008445adf0e38ed9c72",
  "id": "commit:04459d8",
  "kind": "commit",
  "paths": [
    "README.md",
    "cmd/lawang/serve.go",
    "cmd/lawang/serve_timeouts_test.go",
    "docs/architecture.md",
    "docs/backlog.md",
    "internal/config/config.go",
    "internal/ingress/ingress.go",
    "internal/ingress/ingress_test.go",
    "internal/provider/fake/fake.go",
    "internal/provider/fake/fake_test.go",
    "internal/provider/provider.go",
    "internal/provider/registry.go",
    "internal/provider/registry_test.go",
    "internal/record/record.go",
    "internal/record/scope.go",
    "internal/record/validate.go"
  ],
  "text": "The webhook edge faces the public internet, so its order is its security\nargument. The path segment is text a stranger sent: net/http decodes %00 in\nit to a NUL byte and hands over 500 bytes of anything just as happily. So\nthe segment is resolved to a registered provider before a byte of the body\nis...",
  "title": "Put a provider behind the ingress edge before anything is read"
}
```

## doc

```json
{
  "bytes": 11152,
  "id": "doc:.claude/agents/bizdev.md",
  "kind": "doc",
  "paths": [
    ".claude/agents/bizdev.md"
  ],
  "text": "---\nname: bizdev\ndescription: Business development and community growth for Lawang. Researches who needs this project and what else exists, finds what stops a stranger from adopting it, proposes backlog items that make it more useful, prepares the repository for a worldwide community, and drafts lau...",
  "title": ".claude/agents/bizdev.md",
  "truncated": false
}
```

## file

```json
{
  "bytes": 640,
  "id": "file:.github/workflows/ci.yml",
  "kind": "file",
  "paths": [
    ".github/workflows/ci.yml"
  ],
  "text": "name: ci\n\non:\n  push:\n    branches: [main]\n  pull_request:\n\npermissions:\n  contents: read\n\nconcurrency:\n  group: ci-${{ github.ref }}\n  cancel-in-progress: true\n\njobs:\n  check:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-go@v5\n        with:\n   ...",
  "title": ".github/workflows/ci.yml",
  "truncated": false
}
```

## issue

```json
{
  "author": "gablooge",
  "date": "2026-09-18T13:53:15Z",
  "id": "issue:1",
  "kind": "issue",
  "labels": [
    "backlog"
  ],
  "milestone": "M0 · Foundations",
  "number": 1,
  "paths": [],
  "state": "closed",
  "text": "`cmd/sluiceway` with `serve`, `worker`, `migrate`, `version` (stubs where the role does not exist yet); `internal/config` from the environment with fail-closed defaults; `Makefile` with `check`; `.golangci.yml`; GitHub Actions running vet, lint and `go test -race`.\n\n## Done when\n\n- [ ] `make check` ...",
  "title": "B01: Skeleton, config, CI"
}
```

## issue_comment

```json
{
  "author": "gablooge",
  "date": "2026-09-18T19:23:10Z",
  "id": "issue_comment:10:5735082126",
  "issue": 10,
  "kind": "issue_comment",
  "paths": [],
  "text": "Carried over from the round 1 review of #32 (B04): in the outbox, `attempts` is only consulted by `Fail`. A row whose processing kills or hangs its worker never reaches `Fail`, so its lease just expires and it is taken over again, forever, holding every later version of its entity behind it.\n\nThe dr...",
  "title": "Comment on issue #10"
}
```

## pr

```json
{
  "author": "gablooge",
  "branch": "b01-skeleton",
  "date": "2026-09-18T18:39:35Z",
  "id": "pr:30",
  "kind": "pr",
  "labels": [
    "backlog",
    "review:needs-maintainer"
  ],
  "merged_at": "2026-09-19T04:24:27Z",
  "number": 30,
  "paths": [
    ".github/workflows/ci.yml",
    ".golangci.yml",
    "CLAUDE.md",
    "Makefile",
    "README.md",
    "cmd/sluiceway/main.go",
    "cmd/sluiceway/main_test.go",
    "cmd/sluiceway/serve.go",
    "docs/architecture.md",
    "docs/backlog.md",
    "docs/roadmap.md",
    "go.mod",
    "go.sum",
    "internal/appversion/appversion.go",
    "internal/appversion/appversion_test.go",
    "internal/config/config.go",
    "internal/config/config_test.go",
    "internal/ids/ids.go",
    "internal/ids/ids_test.go",
    "internal/ids/testdata/golden.json"
  ],
  "state": "merged",
  "text": "The first two backlog items, plus the plan they come from. This description is current as of `dad126c`, after three review rounds and the delta review of `7b3a2ee`.\n\n## B01: skeleton, config, CI (closes #1)\n\n- `cmd/sluiceway` with `serve`, `worker`, `migrate`, `version` and `help` (`-h`, `--help`). ...",
  "title": "M0: CLI skeleton, config, CI, and internal/ids (B01, B02)"
}
```

## review

```json
{
  "author": "gablooge",
  "date": "2026-09-18T19:17:49Z",
  "id": "review_comment:30:4049956732",
  "kind": "review",
  "paths": [
    "internal/config/config_test.go"
  ],
  "pr": 30,
  "text": "**blocking** The test that exists to prove \"errors never echo the value\" only exercises a URL that parses. The branch where the value is most likely to leak, a `url.Parse` failure, has no coverage, and the leak mutation survives.\n\n**Fails when:** I changed `checkDatabaseURL` to `return err` on the p...",
  "title": "Review comment on PR #30, internal/config/config_test.go"
}
```
