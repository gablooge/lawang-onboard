# Data sources

| Source | License | What we use | Collected how |
|---|---|---|---|
| [gablooge/lawang](https://github.com/gablooge/lawang), the team lead's own open-source project | Apache-2.0 | Source files, docs, architecture decision records, commit messages, pull request titles, bodies and review comments | `scripts/export_corpus.py`, read-only, from a local clone and the GitHub API |

## What is not used

- No client data, no company confidential data, no personal information, no social media data.
- No third-party websites are scraped.

## Scrubbing

Before an export is committed, `scripts/export_corpus.py` replaces, in commit messages, pull
requests, reviews and issues: email addresses, IP addresses, any GitHub login other than the
repository owner's, token-looking strings, local machine paths, and the owner's personal hostname.
In source files it replaces only token-looking strings and local paths, because the addresses in
the source are deliberately fake test fixtures (RFC 5737 ranges, `example` domains). Review
comments left with no content after scrubbing are dropped. The report is in
`corpus/manifest.json` under `scrubbed`.

## Roles

The roles (`maintainer`, `employee`, `contractor`) are synthetic and defined in `roles.yaml`. They
do not describe real people.

## Counts

From `corpus/manifest.json`, source commit `baa08cf` of `gablooge/lawang`:

| Kind | Items |
|---|---|
| Source files | 79 |
| Docs | 10 |
| Architecture decision records | 6 |
| Commits | 50 |
| Pull requests | 13 |
| Review bodies and review comments | 332 |
| Issues | 36 |
| Issue comments | 84 |

Items per scope are computed by the server from `roles.yaml` when it loads the corpus.
