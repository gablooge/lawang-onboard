# Problem and solution

Draft, Fri 25 Sep. Target 400 words, hard limit 500. Numbers in [brackets] are filled in from the
benchmark on Saturday.

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

We built and tested it on a real open-source codebase with [50] commits, [6] decision records and
[13] reviewed pull requests. On [10] real onboarding questions, Onboard answered in [seconds] what
took [minutes] by hand, with [n] citations per answer. A leak suite of [N] probes per role,
including prompt injection planted inside the corpus, returned [0] items outside a role's scopes,
and it runs in CI on every push.

## Why it matters

Onboarding becomes self-service without becoming a data leak. The same filter works for any model
and any repository, and it comes from Lawang, our permission-aware connector project, which
applies the same scope rule to Slack, Teams, Outlook, ClickUp and HubSpot.
