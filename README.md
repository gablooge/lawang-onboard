# Lawang Onboard

![Lawang Onboard](docs/assets/cover.png)

**An onboarding copilot inside IBM Bob that explains a codebase to a new engineer, and only the
parts they are allowed to see.**

Built for the IBM Bob 2.0 Hackathon (lablab.ai, 25 to 27 September 2026) by team Lawang Onboard.

> Status: in progress during the hackathon. Sections marked TODO are filled in as they are built.

## The problem

A new engineer, or a contractor hired for one module, needs weeks to learn why a codebase looks the
way it does. The answers exist, spread across commit messages, architecture decision records and
pull request reviews, but nobody reads those on day one. And access is all or nothing: either the
newcomer sees everything, or they wait for someone to explain.

## What it does

Open the repository in IBM Bob, switch to the **Onboard** mode, and ask:

- "Give me the tour."
- "Trace how a webhook becomes a record."
- "Why is there no message broker?"
- "What can I pick up first?"

Bob answers from the code and its history, with citations. Every piece of context carries a
**scope**; the engineer's role maps to a set of scopes; and the Lawang Onboard MCP server filters
**before** the model sees anything. A contractor's session cannot leak what it was never given, and
every answer says what was withheld.

## How it works

```
Bob IDE (Onboard mode, skills, slash commands)
        |  MCP, bearer token -> role
        v
onboard server (Go)  --  token -> role -> scopes -> one filter function -> audit log
        |
corpus/*.jsonl  (files, docs, ADRs, commits, PR reviews, each stamped with scopes)
```

TODO: diagram, tool list, demo link.

- **Live demo:** https://lawang-onboard.samsulhadi.com
- **Connect IBM Bob:** copy [`.bob/mcp.remote.example.json`](.bob/mcp.remote.example.json) to
  `.bob/mcp.json` and add the contractor token (ask Samsul). Hosting details in
  [docs/deploy.md](docs/deploy.md).

## Impact

TODO: time to answer by hand vs with Onboard, correctness, leak suite result, Bobcoins per answer.

## How IBM Bob is used

TODO: summary of `docs/submission/bob-usage.md`. Evidence: [`bob_sessions/`](bob_sessions/).

## Run it

TODO.

## Data

See [DATA_SOURCES.md](DATA_SOURCES.md). The demo corpus is the team's own open-source repository;
no client, confidential, personal or social media data is used.

## Team

- Samsul Hadi (lead)
- Ryan Rizki

## License

[MIT](LICENSE)
