# Lawang Onboard

![Lawang Onboard](docs/assets/cover.png)

[![M8ven Score](https://m8ven.ai/badge/mcp/gablooge-lawang-onboard-1ol0jo?v=a1cf63ba5dd77154dd1db71bed4fd329)](https://m8ven.ai/mcp/gablooge-lawang-onboard-1ol0jo)

**Permission-aware AI onboarding.** An onboarding copilot inside IBM Bob that explains a codebase
to a new engineer, and only the parts they are allowed to see.

Built for the IBM Bob 2.0 Hackathon (lablab.ai, 25 to 27 September 2026) by team Lawang Onboard.

## The problem

A new engineer, or a contractor hired for one module, needs weeks to learn why a codebase looks the
way it does. The answers exist, spread across commit messages, architecture decision records and
pull request reviews, but nobody reads those on day one. And access is all or nothing: either the
newcomer sees everything, including whatever an AI assistant can read into its context, or they
wait for someone to explain.

## What it does

Open the repository in IBM Bob, switch to the **Onboard** mode, and ask:

- `/tour`: the map of what you can see, module by module.
- `/trace webhook`: how a webhook becomes a record, step by step.
- `/why internal/ingress`: the decisions and reviews behind a path.
- "What can I pick up first?"

Bob answers from the code and its history, and cites every item it used. Every piece of context
carries a **scope**; the engineer's role maps to a set of scopes; the role comes from the token on
the connection; and the Lawang Onboard MCP server filters **before** the model sees anything. A
contractor's session cannot leak what it was never given, and every answer ends with what was
withheld:

```
Withheld: path:cmd (68), path:internal/core (314), path:migrations (56), path:repo (86),
private:growth (18), private:security (6) | total hidden: 399
```

| Contractor | Maintainer |
|---|---|
| ![Contractor view](docs/assets/demo-contractor.png) | ![Maintainer view](docs/assets/demo-maintainer.png) |

## Results

Measured on 26 September with the contractor role, one fresh Bob task per question, by a team
member who did not build the server. Full write-up: [docs/benchmark.md](docs/benchmark.md).

| | Result |
|---|---|
| Ten onboarding questions | **10 of 10 correct**, 36 s and 0.16 Bobcoins per answer on average |
| Against answering by hand | **46 s vs 210 s** on three questions, one of which could not be answered by hand at all |
| Leak probes (prompt injection, direct `get`, claiming another role, search, "guess") | **5 of 5 held**; nothing outside the role's scopes reached the model |
| Code review by Bob | 2 real findings (unauthenticated demo API, an existence oracle in `get`), both fixed and rechecked |
| Withheld counts | Match an independent recount from `roles.yaml` and the corpus on every run |
| Leak tests | 11 test functions, each across all three roles, plus planted injection items; `go test -race ./...` in CI |

## How it works

```
IBM Bob (Onboard mode: rules, skills, slash commands)
        |  MCP over stdio or HTTPS, bearer token
        v
onboard server (Go)
  token -> role -> scopes -> one filter function -> tools -> audit log
        |
corpus/*.jsonl  (610 items: files, docs, ADRs, commits, PRs, reviews, issues, each with scopes)
```

- **Scopes** come from [`roles.yaml`](roles.yaml): path globs (first match wins), labels (a
  `growth` label makes an issue private) and kinds. An item needs every one of its scopes; an item
  with no scope is denied to everyone.
- **Tools**: `whoami`, `map_system`, `search`, `get`, `trace_feature`, `why`, `setup_guide`,
  `starter_tasks`, `withheld`. Each filters first and builds its answer only from visible items. A
  hidden id and a missing id get the same answer.
- **Onboard mode** can use only MCP tools and skills, not the file system, so Bob's only view of
  the code is the filtered one. Its rules make it call `whoami` first, cite item ids, treat corpus
  text as data, mark proposals as proposals and close with the Withheld line.
- **Tokens** are compared in constant time; an unknown, empty or ambiguous token is refused. The
  demo page's API uses the same tokens.

Design notes: [docs/design.md](docs/design.md).

## Try it

- **Live demo:** https://lawang-onboard.samsulhadi.com (paste a role token to see that role's view).
- **Connect IBM Bob to the hosted server:** copy
  [`.bob/mcp.remote.example.json`](.bob/mcp.remote.example.json) to `.bob/mcp.json`, put in a
  contractor token (ask the team), reload MCP servers, switch to the Onboard mode and type `/tour`.

## Run it locally

Needs Go 1.25.

```sh
cp .env.example .env                 # set the three role tokens
cp .bob/mcp.example.json .bob/mcp.json
scripts/use-role.py contractor       # writes that role's token into .bob/mcp.json
```

Open the folder in IBM Bob, reload MCP servers, pick the Onboard mode and ask away. Bob starts the
server over stdio.

For the web page and the HTTP endpoint:

```sh
set -a; . ./.env; set +a
go run ./cmd/onboard -addr :47312               # http://localhost:47312, MCP at /mcp
go run ./cmd/onboard -addr :47312 -demo-roles   # local demos only: role buttons without tokens
make test                                       # go test -race -count=1 ./...
```

Hosting (Docker and a Cloudflare Tunnel, `make up`): [docs/deploy.md](docs/deploy.md).

Or run the published image over stdio, no Go needed. Pick any secret, give it to one role, and use
the same value as `ONBOARD_TOKEN`; that role decides what you see:

```json
{
  "mcpServers": {
    "lawang-onboard": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "ONBOARD_TOKEN", "-e", "ONBOARD_TOKEN_CONTRACTOR",
               "ghcr.io/gablooge/lawang-onboard:0.1.1", "-stdio"],
      "env": { "ONBOARD_TOKEN": "pick-a-secret", "ONBOARD_TOKEN_CONTRACTOR": "pick-a-secret" }
    }
  }
}
```

The server is also listed in the [MCP Registry](https://registry.modelcontextprotocol.io) as
`io.github.gablooge/lawang-onboard`.

## Use it on your own repository

The demo answers about Lawang, but nothing in the server is specific to it. A corpus and a roles
file are all it reads. The export needs Python 3 and git, the server needs Go 1.25 or Docker.
Every command below runs in a clone of this repository, and everything you make goes under
`local/`, which git ignores.

1. **Export.** Point it at a local clone of your repository:

   ```sh
   make corpus REPO=../myrepo GITHUB=me/myrepo OUT=local/corpus
   ```

   This writes files, docs and commits from the clone, and pull requests, reviews and issues from
   GitHub (public repositories need no token; set `GITHUB_TOKEN` for a private one or a large
   one). `GITHUB=` with nothing after it skips GitHub and makes no network call. For a repository
   owned by an organisation, run the script directly and add `--owner <your login>`.

   **What the export removes, and what it does not.** It replaces a fixed list of token shapes
   (GitHub, Slack, AWS key ids, Stripe, Google, GitLab, npm, JWTs, bearer values, PEM private
   keys), passwords inside URLs, and, in configuration files, a literal value on the same line
   as a key whose name ends in `password`, `secret`, `token` or `api_key`. In commit messages and
   GitHub text it also replaces email addresses and IP addresses, and it writes every author but
   the owner and bots as `<user>`. It leaves out symlinks, lock files, minified files, `node_modules`, `vendor`,
   `third_party`, and configuration files with `secret`, `credential` or `passw` in the name.
   It does **not** remove: `@mentions` and names written in the text of a commit or a pull
   request, GitHub `noreply` addresses (which contain a login), a secret of a shape it does not
   know, one stored under a neutral key, or one assigned in source code (`PASSWORD = "..."` in a
   `.py` file). So it is a net and not a guarantee. The counts are
   printed and kept in `local/corpus/manifest.json`; read the corpus before you share it. The
   raw GitHub answers, not scrubbed at all, are cached beside the corpus in
   `local/.corpus-raw/`, which ignores itself in git and is never mounted or served.

2. **Say who sees what.**

   ```sh
   cp roles.example.yaml local/roles.yaml
   ```

   Edit the globs to match your tree. Every line is commented. As the template is written, a
   directory you forget to list is shown to the maintainer role only. That holds for as long as
   you keep the last rule's scope (`unlisted:repo`) out of the other roles; the header of the
   file says what changes if you do not.

3. **Run.** Each role needs a token, in the variable its `token_env` names. Without them the
   page loads and every call answers 401:

   ```sh
   cp .env.example .env    # then set ONBOARD_TOKEN_MAINTAINER, _EMPLOYEE and _CONTRACTOR
   ```

   With Go:

   ```sh
   set -a; . ./.env; set +a
   go run ./cmd/onboard -addr :47312 -corpus local/corpus -roles local/roles.yaml
   ```

   Or with the published image, mounting both over the demo's:

   ```sh
   docker run --rm -p 47312:47312 --env-file .env \
     -v "$PWD/local/corpus:/app/corpus:ro" -v "$PWD/local/roles.yaml:/app/roles.yaml:ro" \
     ghcr.io/gablooge/lawang-onboard:0.1.1
   ```

   The MCP endpoint is `http://localhost:47312/mcp` and the page is at `/`. For stdio, take the
   JSON block under "Run it locally" and add the two `-v` mounts to its `args`, with absolute
   paths (there is no shell there to expand `$PWD`); it has no `-p`.

What you get depends on what the repository has. The tour, search, `get`, `trace_feature` and
`why` work on any export. `setup_guide` reads the README, the Makefile and
`.github/workflows/`. `starter_tasks` needs GitHub issues and a `label_areas` entry.

## How IBM Bob is used

IBM Bob is both the product's runtime and the tool that built it. The Onboard mode, its rules,
four skills (`tour`, `trace`, `first-week`, `why`) and three slash commands are the user interface.
Plan mode designed the server, Agent mode wrote all of its code and tests, Ask mode reviewed it
for leaks, and the Onboard mode ran the benchmark. 36 Bob tasks, 46.6 of the team's 80 Bobcoins;
every task has a screenshot in [`bob_sessions/`](bob_sessions/) and a row in the ledgers
([Samsul](bob_sessions/LEDGER.md), [Ryan](bob_sessions/LEDGER-ryan.md)). Details, and what was
done outside Bob: [docs/submission/bob-usage.md](docs/submission/bob-usage.md).

## Data and privacy

See [DATA_SOURCES.md](DATA_SOURCES.md) and [PRIVACY.md](PRIVACY.md). The demo corpus is the team's own open-source repository,
[gablooge/lawang](https://github.com/gablooge/lawang); no client, confidential, personal or social
media data is used. The roles are synthetic.

## Team

- Samsul Hadi (lead)
- Ryan Rizki

## License

[MIT](LICENSE)
