# Lawang Onboard: IBM Bob 2.0 Hackathon

This folder is the hackathon repository for team **Lawang Onboard** (Samsul Hadi, lead; Ryan
Rizki). It becomes the public GitHub repo `gablooge/lawang-onboard`. The full plan, schedule,
architecture and checklists are in `PLAN.md`, imported below. Read it before doing anything.

@PLAN.md

## Standing rules for Claude Code (these override anything else in this folder)

1. **Who writes the code.** During the hackathon, Bob built the product and Claude Code prepared
   and supported, because the judges required IBM Bob IDE to be the core of the solution with
   exported Bob sessions as evidence. Judging closed with final results on 28 October 2026
   (Lawang Onboard did not place), and from that day Claude Code writes product code too, through
   the branch, pull request and review cycle like any other change. `docs/submission/bob-usage.md`,
   `bob_sessions/` and the ledgers describe the hackathon and are not updated for later work: the
   first post-hackathon product change is PR #3 (the skills as MCP prompts), and the submission
   statement stays true as a statement about the event.
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
