---
name: first-week
description: Use when a new engineer wants a plan for their first week. Calls setup_guide for environment setup steps and starter_tasks for open issues scoped to the caller's role, then produces a short actionable first-week plan.
---

# First-Week Plan

Follow these steps to produce a first-week plan for a new engineer.

## Step 1 - Call whoami

Call `whoami` and state the role in one line before proceeding.

## Step 2 - Get the setup guide

Call `setup_guide` to retrieve the README, Makefile targets, and CI workflow details. Note the
commands needed to get the project running locally.

## Step 3 - Get starter tasks

Call `starter_tasks` to retrieve open issues whose label areas are within the caller's scopes.
Note the issue IDs, titles, and any relevant labels.

## Step 4 - Build the first-week plan

Produce a short plan structured as five days:

- **Day 1:** Environment setup using the steps from `setup_guide`. List the exact commands.
- **Day 2:** Read the top-level documentation (from `setup_guide` README output). Identify the
  main components.
- **Day 3:** Run the tour skill (or summarize the system from `map_system`) to understand
  the codebase layout.
- **Day 4:** Pick the first starter task from `starter_tasks` and read its context with `get`.
  Cite the issue ID.
- **Day 5:** Make a small contribution: fix the starter task or write a note on findings.
  Use `why` on any file you are unsure about before touching it.

Keep each day to 3-5 bullet points. Cite every item ID used.

## Step 5 - Withheld line

Call `withheld` and append: `Withheld: <scope> (<count>), ...` (or `Withheld: none`).
