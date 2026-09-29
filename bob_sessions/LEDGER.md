# Bobcoin ledger

Budget: 40 per person, 80 for the team. Samsul's tasks are here; Ryan's tasks from 17 on are in
[LEDGER-ryan.md](LEDGER-ryan.md), committed by Ryan. Stop starting new tasks at 80% spent; the rest is for the
recorded demo.

| Task | Time (WIB) | Who | Mode | Description | Coins | Running total | Screenshot |
|---|---|---|---|---|---|---|---|
| 01 | Fri 22:23 | Samsul | Agent | Read AGENTS.md and summarize the rules (setup check; includes one compaction) | 0.096 | 0.096 | lawangonboard_task01_setup_check_summary.png |
| 02 | Fri 22:33 | Samsul | Plan | Design: docs/design.md (package layout, filter, scope computation, leak tests, build order) | 0.468 | 0.564 | lawangonboard_task02_design_plan_summary.png |
| 03 | Fri 22:37 | Samsul | Agent | Steps 1 and 2: go.mod, cmd/onboard stub, internal/roles (glob matcher), internal/corpus, tests | 5.32 | 5.884 | lawangonboard_task03_corpus_roles_loader_summary.png |
| 04 | Fri 22:50 | Samsul | Agent | Step 3: internal/filter (Visible, Withheld), rule table tests, leak tests vs an independent oracle, injection items | 1.36 | 7.244 | lawangonboard_task04_filter_leak_tests_summary.png |
| 05 | Fri 23:05 | Samsul | Agent | Step 4 part 1: internal/tools (whoami, get, search, map_system, withheld), MCP wiring stdio + HTTP, auth tests | 10.24 | 17.484 | lawangonboard_task05_mcp_server_part1_summary.png |
| 06 | Fri 23:35 | Samsul | Agent | Step 4 part 2: trace_feature, why, setup_guide, starter_tasks; constant-time token compare; duplicate token_env refusal | 3.82 | 21.304 | lawangonboard_task06_mcp_server_part2_summary.png |
| 07 | Fri 23:52 | Samsul | Ask | MCP end-to-end check: whoami + get growth README (token was the maintainer one, so full access, as designed) | 0.053 | 21.357 | lawangonboard_task07_mcp_check_as_maintainer_summary.png |
| 08 | Fri 23:58 | Samsul | Ask | MCP end-to-end check as contractor: whoami = contractor (5 scopes), get growth README = withheld, no content | 0.052 | 21.409 | lawangonboard_task08_mcp_check_as_contractor_summary.png |
| 09 | Sat 00:45 | Samsul | Agent | Ambiguous-token denial; Onboard mode, rules, tour/trace/first-week skills, /tour /trace /why commands | 1.01 | 22.419 | lawangonboard_task09_onboard_mode_skills_summary.png |
| 10 | Sat 01:10 | Samsul | Onboard | /tour as contractor (first product run; found the withheld-count bug) | 0.370 | 22.789 | lawangonboard_task10_tour_as_contractor_summary.png |
| 11 | Sat 01:25 | Samsul | Agent | Fix: withheld counts only scopes the role lacks; withheld tool returns per-scope and total; rules: one withheld call, no em dashes | 4.19 | 26.979 | lawangonboard_task11_withheld_fix_summary.png |
| 12 | Sat 06:33 | Samsul | Onboard | /tour as contractor after the withheld fix: Withheld line matches the independent count exactly (406 hidden) | 0.276 | 27.255 | lawangonboard_task12_tour_fixed_summary.png |
| 13 | Sat 06:40 | Samsul | Agent | Demo page (embedded index.html), /api roles/tour/withheld/search/audit, internal/audit ring, -web-only flag | 2.33 | 29.585 | lawangonboard_task13_demo_page_api_summary.png |
| 14 | Sat 07:05 | Samsul | Agent | /api/audit filtered by role (test: no cross-role entries or hidden ids), centered layout, UTC labels, demo note | 0.938 | 30.523 | lawangonboard_task14_audit_filter_layout_summary.png |
| 15 | Sat 08:36 | Samsul | Onboard | /tour as contractor, final: provider files, fake, registry and ADRs covered; Withheld exact (406). Video take. Bob also mirrored /why as a skill | 0.480 | 31.003 | lawangonboard_task15_tour_final_summary.png |
| 16 | Sat 09:40 | Ryan | Onboard | whoami through the hosted server (https, Cloudflare Tunnel) with the contractor token: contractor, Withheld 406 | 0.032 | Ryan 0.032 | lawangonboard_task16_ryan_whoami_remote_summary.png |
| 32 | Sat 16:27 | Samsul | Agent | Fix from task 25: setup:public scope (Makefile, .github, go.mod, .golangci.yml, sqlc.yaml) for contractor and employee, rules for proposals and corpus-only citations, six filter test cases; go vet and go test -race pass | 0.772 | 31.775 | lawangonboard_task32_setup_scope_fix_summary.png |
| 36 | Sat 18:52 | Samsul | Onboard | /tour as contractor after the setup:public fix: the build and CI files now show up (Makefile, ci.yml, go.mod, .golangci.yml, sqlc.yaml); Withheld 399; slide screenshot | 0.461 | 32.236 | lawangonboard_task36_tour_after_fixes_summary.png |
| 38 | Tue 15:50 | Samsul | Agent | Demo page: "Source on GitHub" link in the header and a footer with a Privacy link to PRIVACY.md; go vet and go test -race pass; 15 s | 0.282 | 32.518 | lawangonboard_task38_github_link_footer_summary.png |
