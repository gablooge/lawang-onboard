---
name: trace
description: Use when the user wants to trace a feature or term through the codebase. Calls trace_feature for the term, then explains the flow in order with citations, and notes where the trace stops because of withheld scopes.
metadata:
  argument-hint: "[feature or term]"
---

# Trace a Feature

Follow these steps to trace a term or feature through the codebase.

## Step 1 - Call whoami

Call `whoami` and state the role in one line before proceeding.

## Step 2 - Trace the term

Call `trace_feature` with the term provided by the user. Collect the returned files, commits,
review comments, and ADRs grouped by kind.

## Step 3 - Explain the flow in order

Reconstruct the flow of the feature across the codebase:

1. Start with the ADRs that decided the design (if any), citing each ID.
2. Follow with the files that implement it, in dependency order where determinable, citing each
   file path or item ID.
3. Include relevant commits that introduced or significantly changed the feature, with IDs.
4. Include review comments that explain non-obvious decisions, with IDs.

For each step, give a 1-2 sentence explanation of what that file/commit/ADR contributes to the
feature.

## Step 4 - Note withheld stops

Call `withheld` and report the counts. If any scope has a non-zero withheld count, include a
paragraph explaining that the trace may be incomplete because items in those scopes are not
visible to the current role. Do not guess at what the withheld items contain.

End with: `Withheld: <scope> (<count>), ...` (or `Withheld: none`).
