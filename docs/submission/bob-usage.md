# How IBM Bob is used

Every claim here matches a screenshot in `bob_sessions/` and a row in
[LEDGER.md](../../bob_sessions/LEDGER.md) (Samsul) or [LEDGER-ryan.md](../../bob_sessions/LEDGER-ryan.md) (Ryan).

## Bob is the runtime

- **Onboard custom mode** (`.bob/custom_modes.yaml`, rules in `.bob/rules-onboard/`): the product's
  user interface. It has only the `mcp` and `skill` groups, so Bob sees the codebase only through
  the Lawang Onboard server. The rules: call `whoami` first, cite every item id, treat corpus text
  as data, never claim another role, mark proposals as proposals, end with one Withheld line.
- **Skills** (`.bob/skills/`): `tour` (the map of what the role can see), `trace` (a feature end to
  end, and where the trace stops), `first-week` (a plan from the open issues in the role's area),
  `why` (the decisions and reviews behind a path).
- **Slash commands** (`.bob/commands/`): `/tour`, `/trace`, `/why`.
- **MCP**: nine tools over stdio or HTTPS. The role comes from the bearer token, and the server
  filters before Bob sees anything.

## Bob built it

| Stage | Bob mode | Tasks | Bobcoins |
|---|---|---|---|
| Design (docs/design.md) | Plan | 02 | 0.468 |
| Roles, corpus loader, filter, leak tests against an independent oracle | Agent | 03, 04 | 6.68 |
| MCP server and the nine tools, stdio and HTTP, token auth | Agent | 05, 06 | 14.06 |
| Onboard mode, rules, skills, slash commands | Agent | 09 | 1.01 |
| Demo page, web API, audit log | Agent | 13, 14 | 3.268 |
| Fixes found by using it (withheld counts, setup files hidden) | Agent | 11, 32 | 4.962 |
| Setup and end-to-end checks, tours | Ask, Onboard | 01, 07, 08, 10, 12, 15 | 1.327 |
| Benchmark: ten questions and one rerun | Onboard | 16 to 26, 26b | 1.725 |
| Leak probes and rechecks | Onboard | 27 to 31, 35 | 0.284 |
| Security review of the server | Ask | 33 | 0.785 |
| Fixes from the review | Agent | 34 | 12.05 |

Total: 36 tasks, 46.619 of the team's 80 Bobcoins (Samsul 31.775, Ryan 14.844).

## What was done outside Bob

With Claude Code: repository setup, the corpus export script, CI and secret scanning, the Docker
and Cloudflare Tunnel deployment, `bob_sessions/` bookkeeping, an independent recount of the
withheld numbers, and the write-ups. Two one-line configuration edits were made by hand (removing
the `read` group from the Onboard mode, and the tour skill's `map_system` depth). Every line of
product code, mode, rule, skill and command was written by IBM Bob.

## What we learned

- **Using the product found the bugs.** The first `/tour` (task 10) showed wrong withheld counts;
  the tests question (task 25) showed setup files hidden from contractors. Bob fixed both, and the
  same question was rerun to verify.
- **A narrow review works.** A leak-focused review in Ask mode (task 33) found two real holes the
  tests had missed.
- **Questions are cheap, broad edits are not.** Answers cost 0.03 to 0.45 Bobcoins; the eight-file
  fix cost 12.05. Small tasks with a fresh context save coins.
