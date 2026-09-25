# Lawang Onboard: IBM Bob 2.0 Hackathon

This folder is the hackathon repository for team **Lawang Onboard** (Samsul Hadi, lead; Ryan
Rizki). It becomes the public GitHub repo `gablooge/lawang-onboard`. The full plan, schedule,
architecture and checklists are in `PLAN.md`, imported below. Read it before doing anything.

@PLAN.md

## Standing rules for Claude Code (these override anything else in this folder)

1. **Bob builds the product, Claude Code prepares and supports.** The judges require IBM Bob IDE
   to be the core of the solution, with exported Bob sessions as evidence. Claude Code does not
   write product code (the MCP server, the corpus/permission engine, the Bob "Onboard" mode, the
   demo web app). It does: repository setup, data export scripts, CI, secret scanning, deploy
   scripts, submission documents, checklists, reviews on request. If Samsul explicitly asks for a
   product piece, say once that it weakens the Bob evidence, then follow his decision and record
   it in `docs/submission/bob-usage.md` so the statement stays honest.
2. **This repository is public.** Nothing personal, no machine names, hostnames, IPs, tokens or
   emails in any file, commit, issue or PR. Real names allowed: only the two team members as
   listed on the lablab team page.
3. **Secrets:** never print, commit or put on a command line any IBM Cloud key, Bob credential,
   Cloudflare token or deploy key. Run `make secrets` before every push. A leaked IBM credential
   gets the team's account deactivated.
4. **`../lawang` is read-only.** Read files, run `git log` and `gh ... view/list` against it. Never
   branch, commit, push, edit or run its agents. Never read `~/.config/lawang/`.
5. **Data rules:** only our own repository data (Lawang), synthetic roles, no personal, client or
   social media data. Anything exported goes through the scrub step in `PLAN.md` section 5 (Block B).
6. **House style:** no em dashes anywhere. No tool attribution lines and no `Co-Authored-By`
   trailers in commits, docs or PRs.
7. **Never merge or force-push without Samsul saying so in the conversation.**
8. Keep `PLAN.md` current: tick checkboxes as work finishes, write decisions into its
   decision log (section 14), and record Bob config facts you verify in section 8.
