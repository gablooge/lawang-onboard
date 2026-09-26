# Problem and solution

## The problem

A new engineer's first weeks go into reconstructing context. The code shows what the system does;
the reasons live elsewhere, in commit messages, architecture decision records and pull request
reviews that nobody reads on day one. So newcomers interrupt senior engineers with questions
whose answers are already written down, and senior engineers lose hours repeating them.

Contractors make it worse. A contractor hired for one module either gets access to everything,
which is a risk the moment an AI assistant reads the whole repository into its context, or gets
access to almost nothing and waits for someone to explain. Today's AI coding assistants do not help
here: they see whatever is in the workspace, so the only way to limit what they can reveal is to
limit what the person can open.

## The solution

Lawang Onboard is an onboarding copilot that lives inside IBM Bob. A new engineer opens the
repository in Bob, switches to the Onboard mode and asks for a tour, a feature traced end to end,
the reason behind a decision, how to set up, or what to pick up first. Bob answers from the code and
from its history, and cites every commit, decision record and review it used.

Every piece of context carries a scope: the part of the code it touches, or the kind of material it
is. The engineer's role maps to a set of scopes, and the role comes from the token on the
connection, never from anything the model or a document can say. Bob gets its context only through
the Lawang Onboard MCP server, which filters before the model sees a single line. A contractor's
session cannot leak what it was never given. And every answer says what was withheld, so the gap
is visible instead of papered over.

We built and tested it on a real open-source codebase: 610 items, including 50 commits, 6 decision
records, 13 pull requests and 332 review comments. Asked ten real onboarding questions as a
contractor, Onboard got all ten right, in 36 seconds and 0.16 Bobcoins per answer on average. On
three of them it took 46 seconds against 210 by hand, and answered one that the manual search
could not. Five attempts to talk it into leaking (prompt injection, calling the tool directly,
claiming another role, searching, asking it to guess) got nothing outside the contractor's scopes,
and the leak tests, with injection text planted in the corpus, run for every role in CI.

## Why it matters

Onboarding becomes self-service without becoming a data leak. The same filter works for any model
and any repository, and it comes from Lawang, our permission-aware connector project, which
applies the same scope rule to Slack, Teams, Outlook, ClickUp and HubSpot.
