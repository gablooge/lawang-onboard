# Data sources

| Source | License | What we use | Collected how |
|---|---|---|---|
| [gablooge/lawang](https://github.com/gablooge/lawang), the team lead's own open-source project | Apache-2.0 | Source files, docs, architecture decision records, commit messages, pull request titles, bodies and review comments | `scripts/export-corpus.sh`, read-only, from a local clone and the GitHub API |

## What is not used

- No client data, no company confidential data, no personal information, no social media data.
- No third-party websites are scraped.

## Scrubbing

Before an export is committed, `scripts/export-corpus.sh` replaces email addresses, any GitHub
login other than the repository owner's, hostnames, IP addresses and token-looking strings with
placeholders, and prints a report of what it replaced.

## Roles

The roles (`maintainer`, `employee`, `contractor`) are synthetic and defined in `roles.yaml`. They
do not describe real people.

## Counts

TODO: filled in by the export (items per kind, items per scope).
