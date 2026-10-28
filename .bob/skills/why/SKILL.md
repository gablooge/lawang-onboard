---
name: why
description: Use when the user asks why a file or directory looks the way it does. Calls why for the path, then explains the decisions behind it from commits, review comments, and ADRs and docs, with citations, and notes what was withheld.
metadata:
  argument-hint: "[repository path]"
---

# Why a Path Looks the Way It Does

Follow these steps to explain the history behind a repository path.

## Step 1 - Call whoami

Call `whoami` and state the role in one line before proceeding.

## Step 2 - Ask for the history

Call `why` with the path provided by the user. It returns three groups: commits, review
comments, and ADRs and docs. If the path is a directory, the result covers everything under it.
Each item comes back as an id, a kind and a title only, so call `get` on an item before you state
its reason, its date or a sentence from it.

## Step 3 - Explain the decisions

Write the explanation in time order:

1. Start with the ADRs and docs that bear on the path, citing each ID, and say what each one
   decided or describes.
2. Follow with the commits, oldest first by the date `get` returns, citing each ID. For each, say
   what changed and why, from the commit message; do not invent a reason the message does not
   give.
3. Add the review comments that explain a non-obvious choice, citing each ID.

Keep it to the decisions a newcomer would otherwise have to ask about. Quote a sentence from an
item when the wording matters; otherwise paraphrase and cite.

## Step 4 - Note withheld stops

Call `withheld` and report the counts. If any scope has a non-zero withheld count, say that the
history may be incomplete because items in those scopes are not visible to the current role. Do
not guess at what the withheld items contain.

End with: `Withheld: <scope> (<count>), ... | total hidden: <total>` (or `Withheld: none`).
