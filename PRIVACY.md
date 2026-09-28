# Privacy

Lawang Onboard is a demo MCP server and web page built for the IBM Bob 2.0 Hackathon. This page
says what it handles and what it keeps.

## What it serves

The corpus in `corpus/` is exported from our own public repository,
[gablooge/lawang](https://github.com/gablooge/lawang): files, docs, decision records, commits, pull
requests, reviews and issues. Email addresses, IP addresses and local paths are scrubbed during the
export. It contains no client, confidential or personal data beyond the public authorship of that
repository. The roles (maintainer, employee, contractor) are synthetic.

## What it keeps

- **Audit log.** Every tool call and web API call is recorded in an in-memory ring of the last 200
  entries: time, role name, tool name, the ids of the items returned, and the per-scope withheld
  counts. Search queries are not recorded. The log is never written to disk and is gone when the
  server restarts. `/api/audit` shows a caller only the entries for their own role.
- **Tokens.** Role tokens are read from environment variables and compared in constant time. They
  are never logged, stored, returned in a response or written to the audit log.
- **Server logs.** Only startup and fatal errors are logged. Requests are not.

## What it does not do

- No accounts, no cookies, no analytics, no tracking.
- The demo page keeps a pasted token in a JavaScript variable only, never in local storage or
  cookies, so it is gone when the tab closes.
- The server makes no outbound network calls. (`scripts/export_corpus.py`, which builds the corpus
  offline, calls the GitHub API; it is not part of the server.)

## Hosting

The hosted demo at https://lawang-onboard.samsulhadi.com runs behind a Cloudflare Tunnel, so
requests pass through Cloudflare and are subject to
[Cloudflare's privacy policy](https://www.cloudflare.com/privacypolicy/). If you run the server
yourself, nothing leaves your machine.

## Contact

Open an issue on [gablooge/lawang-onboard](https://github.com/gablooge/lawang-onboard/issues).
