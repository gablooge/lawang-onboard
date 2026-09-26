# How IBM Bob is used

Skeleton, Fri 25 Sep. Fill in as the build goes; every claim here must match a screenshot in
`bob_sessions/` and a row in `bob_sessions/LEDGER.md`. Hard limit 500 words.

## Bob is the runtime

- **Onboard custom mode** (`.bob/custom_modes.yaml`, rules in `.bob/rules-onboard/`): the
  product's user interface. TODO: what it instructs.
- **Skills** (`.bob/skills/`): `tour`, `trace`, `first-week`. TODO: one line each.
- **Slash commands** (`.bob/commands/`): `/tour`, `/trace`, `/why`. TODO.
- **MCP** (`.bob/mcp.json`): Bob reaches the codebase and its history only through the Lawang
  Onboard server, which filters by role before Bob sees anything.
- **Subagents and parallel tasks**: the tour fans out one subagent per module the role can see,
  then merges. TODO: evidence (task number).

## Bob built it

| Stage | Bob mode or feature | Tasks | Coins |
|---|---|---|---|
| Design | Plan mode, document understanding of the corpus and docs | TODO | TODO |
| Permission filter and leak tests | Agent mode | TODO | TODO |
| MCP server and tools | Agent mode | TODO | TODO |
| Demo web page | Plan then Agent mode | TODO | TODO |
| Reviews | Bob code review on each pull request | TODO | TODO |
| Onboarding answers for the demo | Onboard mode | TODO | TODO |

Total: TODO of 80 Bobcoins across two team members.

## What was done outside Bob

Before kickoff, with Claude Code: repository setup, the data export script that turns the Lawang
repository into `corpus/`, CI and secret scanning, and drafts of these documents. During the
event, with Claude Code: the Docker and Cloudflare Tunnel deployment (Dockerfile, compose.yaml,
docs/deploy.md) and bookkeeping of `bob_sessions/`. Two one-line configuration edits were made
by hand after review (removing the `read` group from the Onboard mode so it can only see the
codebase through the MCP server, and the tour skill's `map_system` depth). Every line of product
code (filter, corpus loader, MCP server and tools, audit, web API and page, Onboard mode, rules,
skills and commands) was written by IBM Bob, as the task screenshots show.

## What we learned

TODO: two or three concrete observations (what Bob did well, where we had to steer it, coins per
task type).
