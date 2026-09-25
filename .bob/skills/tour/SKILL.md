---
name: tour
description: Use when a new engineer wants a first-day tour of the codebase. Calls map_system to list visible modules, then uses parallel subagents to summarize each module with search and why, then merges the results into a structured first-day tour.
---

# First-Day Codebase Tour

Follow these steps to build a first-day tour of the codebase.

## Step 1 - Call whoami

Call `whoami` and state the role in one line before proceeding.

## Step 2 - Map the system

Call `map_system` with `depth: 0` to get every visible file, then group them by directory. Note every visible module or
directory returned.

## Step 3 - Summarize modules in parallel

For each visible top-level module from Step 2, spawn a subagent in parallel. Each subagent should:

1. Call `search` with a query describing the module name and its likely purpose.
2. Call `why` with the module path to retrieve commits, review comments, and ADRs that explain why
   it exists and how it has changed.
3. Return a short paragraph (3-5 sentences) summarizing: what the module does, why it exists
   (from `why` results), and any notable design decisions found in ADRs or reviews.

Cite every item ID used in the summary, for example: `[adr-0003]`, `[commit-abc1234]`.

## Step 4 - Merge into a tour

Collect the subagent summaries and assemble them into a single first-day tour document:

- Open with one paragraph describing the overall system purpose (from `map_system` and top-level
  search results).
- One section per module, using the subagent summary, with item ID citations preserved.
- Close with a "Where to start" section pointing to the 2-3 most accessible entry points for a
  new engineer.

## Step 5 - Withheld line

Call `withheld` and append: `Withheld: <scope> (<count>), ...` (or `Withheld: none`).
Note any modules that could not be summarized because items were withheld.
