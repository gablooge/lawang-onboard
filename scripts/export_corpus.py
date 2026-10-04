#!/usr/bin/env python3
"""Export a repository and its GitHub history into corpus/*.jsonl.

Data preparation only: it records what exists and where (paths, numbers, links). It does not
assign scopes; the server does that from roles.yaml at load time.

Usage:
  scripts/export_corpus.py <path-to-local-clone> [<owner/repo>] [--out corpus] [--owner <login>]

Without <owner/repo> only files and commits are exported, and no network call is made.

GitHub access: public repositories work without a token (60 requests per hour). If GITHUB_TOKEN
is set in the environment it is used; it is never printed, logged or written anywhere.
Standard library only.
"""
import argparse
import collections
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request

# The only GitHub login kept as is; every other login is scrubbed. Set in main() from --owner,
# which defaults to the owner part of <owner/repo>.
OWNER_LOGIN = None

TEXT_EXT = {
    ".go", ".md", ".sql", ".yaml", ".yml", ".json", ".mod", ".toml", ".txt", ".sh",
    ".py", ".js", ".jsx", ".ts", ".tsx", ".rs", ".java", ".kt", ".rb", ".php", ".cs", ".swift",
    ".c", ".h", ".cc", ".cpp", ".hpp", ".html", ".css", ".proto", ".tf", ".ini", ".cfg", ".rst",
}
TEXT_NAMES = {"Makefile", "LICENSE", "Dockerfile", ".gitignore", ".golangci.yml"}
SKIP_PREFIXES = ("bin/", ".claude/worktrees/", ".git/")
# Third-party code is not the project's, and serving it as such would mislead. Any depth.
SKIP_DIRS = {"node_modules", "vendor", "third_party"}
SKIP_FILES = {"go.sum", "package-lock.json", "pnpm-lock.yaml", "npm-shrinkwrap.json", "composer.lock"}
SKIP_SUFFIXES = (".min.js", ".min.css")
# Configuration formats, where a credential is a plain value and not a fixture in a test. These
# get the assignment rule below, and a file of these types whose name says it holds secrets is
# left out altogether.
CONFIG_EXT = {".ini", ".cfg", ".tf", ".toml", ".yaml", ".yml", ".json"}
SECRET_NAME_EXT = CONFIG_EXT | {".txt"}
SECRET_NAME = re.compile(r"secret|credential|passw", re.I)
MAX_FILE_BYTES = 200_000

# ---------------------------------------------------------------- scrubbing

# Rules that apply to everything, source files included.
ALWAYS_RULES = [
    ("local_path", re.compile(r"(?:/private)?/tmp/claude-\d+/[^\s\"'`)]+"), "<local-path>"),
    ("local_path", re.compile(r"/(?:Users|home)/[A-Za-z0-9._-]+(?:/[^\s\"'`)]*)?"), "<local-path>"),
    ("local_path", re.compile(r"-Users-[A-Za-z0-9._-]+"), "<local-path>"),
    # Anchored to the start of a run and bounded, and the email rule below is bounded too: left
    # open, both cost quadratic time on one long line of "a.a.a.", and a 200 kB file did not
    # finish. The email rule has no anchor, so an address glued to a long run is still found.
    ("personal_hostname", re.compile(r"(?<![A-Za-z0-9.-])[A-Za-z0-9.-]{0,200}samsulhadi\.com\b", re.I), "<tunnel-hostname>"),
    ("personal_name", re.compile(r"samsulhadi", re.I), "<owner>"),
]
# Rules for free text only (commit messages, PRs, reviews, issues). Source files keep their
# fixtures (RFC 5737 addresses, example.* emails), which are deliberately fake.
TEXT_RULES = [
    ("email", re.compile(r"[A-Za-z0-9._%+-]{1,64}@[A-Za-z0-9.-]{1,255}\.[A-Za-z]{2,}"), "<email>"),
    ("ipv4", re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"), "<ip>"),
    ("github_token", re.compile(r"\b(?:ghp|gho|ghu|ghs|ghr|github_pat)_[A-Za-z0-9_]{20,}\b"), "<token>"),
    ("slack_token", re.compile(r"\bxox[abprs]-[A-Za-z0-9-]{10,}\b"), "<token>"),
    ("aws_key", re.compile(r"\bAKIA[0-9A-Z]{16}\b"), "<token>"),
    ("openai_like_key", re.compile(r"\bsk-[A-Za-z0-9_-]{20,}\b"), "<token>"),
    ("bearer", re.compile(r"(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{20,}"), "Bearer <token>"),
    ("stripe_key", re.compile(r"\b[sr]k_(?:live|test)_[A-Za-z0-9]{16,}\b"), "<token>"),
    ("google_key", re.compile(r"\bAIza[0-9A-Za-z_-]{35}\b"), "<token>"),
    ("gitlab_token", re.compile(r"\bglpat-[A-Za-z0-9_-]{20,}"), "<token>"),
    ("npm_token", re.compile(r"\bnpm_[A-Za-z0-9]{36}\b"), "<token>"),
    ("jwt", re.compile(r"\beyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}"), "<token>"),
    ("slack_webhook", re.compile(r"https://hooks\.slack\.com/services/[A-Za-z0-9/_-]+"), "<webhook-url>"),
    # The scheme is bounded: unbounded, a long run of "a.a.a." costs quadratic time.
    ("url_password", re.compile(r"(\b[a-z][a-z0-9+.-]{1,20}://[^\s:/@\"'`]{0,200}):[^\s@/\"'`]{1,500}@"), r"\1:<password>@"),
    ("private_key", re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----"), "<private-key>"),
]
# Example addresses in docs and tests are not personal data; keep them readable.
EMAIL_ALLOW = re.compile(r"@(example\.(com|org|net)|users\.noreply\.github\.com)$", re.I)

report = collections.Counter()


TOKEN_RULE_NAMES = {
    "github_token", "slack_token", "aws_key", "openai_like_key", "bearer", "private_key",
    "stripe_key", "google_key", "gitlab_token", "npm_token", "jwt", "slack_webhook", "url_password",
}

# For configuration files only: a key whose name ENDS in a credential word (so token_env,
# tokenUrl, secretName and the plural "secrets" are not keys of this kind), with a literal value
# on the same line. Only spaces and tabs are allowed around the separator: with \s the rule ran
# over the line end and replaced the next line's key. A value that begins with $, < or { is a
# reference or a placeholder and is kept, and so is one that reads as a variable name or a dotted
# reference (CONFIG_KEEP). This does not see a secret under a neutral key (Terraform's
# `default = "..."`), and nothing here can.
CONFIG_RULE = re.compile(
    r"""(?im)^([ \t]*["']?[\w.-]*(?:password|passwd|secret|token|(?:secret|private|api|access|auth)[_-]?key)["']?[ \t]*[:=][ \t]*)(["']?)([^\s"'$<{][^\s"']{7,})\2""")
# Kept: an upper-case name WITH an underscore (ONBOARD_TOKEN_MAINTAINER; upper-case hex and
# base32 have none), and a dotted reference whose segments hold no digit outside an index
# (var.db_password, module.db.outputs[0]; SG.abc123.def456 and first.last2024 are replaced).
CONFIG_KEEP = re.compile(r"^(?:[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+|[A-Za-z_]+(?:\[\d+\])?(?:\.[A-Za-z_]+(?:\[\d+\])?)+)$")


def _config_sub(m):
    if CONFIG_KEEP.match(m.group(3)):
        return m.group(0)
    report["config_secret"] += 1
    return m.group(1) + m.group(2) + "<secret>" + m.group(2)


def scrub(text, source_file=False, config_file=False):
    if not text:
        return text
    rules = ALWAYS_RULES + [r for r in TEXT_RULES if not source_file or r[0] in TOKEN_RULE_NAMES]
    if config_file:
        text = CONFIG_RULE.sub(_config_sub, text)
    for name, rx, repl in rules:
        def sub(m, name=name, repl=repl):
            if name == "email" and EMAIL_ALLOW.search(m.group(0)):
                return m.group(0)
            if name == "ipv4" and m.group(0) in ("127.0.0.1", "0.0.0.0"):
                return m.group(0)
            report[name] += 1
            return m.expand(repl)
        text = rx.sub(sub, text)
    return text


def login(user):
    if not user:
        return None
    name = user.get("login") if isinstance(user, dict) else user
    if name == OWNER_LOGIN:
        return name
    if name and name.endswith("[bot]"):
        return name
    report["other_login"] += 1
    return "<user>"


# ---------------------------------------------------------------- local git

def git(repo, *args):
    return subprocess.run(["git", "-C", repo, *args], check=True, capture_output=True, text=True).stdout


def git_names(repo, *args):
    """Paths from a git command run with -z. Without -z git quotes a name that has a non-ASCII
    byte, and a name with a space cannot be told from two names. A mangled path matches no glob
    in roles.yaml, which would move a private item into whatever the catch-all rule says."""
    return [n for n in git(repo, *args, "-z").split("\0") if n]


def is_text_path(path):
    base = os.path.basename(path)
    return base in TEXT_NAMES or os.path.splitext(base)[1] in TEXT_EXT


def kind_for(path):
    if path.startswith("docs/adr/"):
        return "adr"
    if path.endswith(".md"):
        return "doc"
    return "file"


def export_files(repo):
    items = []
    for path in git_names(repo, "ls-files"):
        base = os.path.basename(path)
        ext = os.path.splitext(base)[1]
        if path.startswith(SKIP_PREFIXES) or base in SKIP_FILES or base.endswith(SKIP_SUFFIXES):
            continue
        if not is_text_path(path):
            continue
        if SKIP_DIRS.intersection(path.split("/")[:-1]):
            report["skipped_third_party_dir"] += 1
            continue
        if ext in SECRET_NAME_EXT and SECRET_NAME.search(base):
            report["skipped_secret_named_file"] += 1
            continue
        full = os.path.join(repo, path)
        if os.path.islink(full):
            # A tracked link can point outside the clone; its target is not the repository's.
            report["skipped_symlink"] += 1
            continue
        size = os.path.getsize(full)
        with open(full, encoding="utf-8", errors="replace") as fh:
            text = fh.read(MAX_FILE_BYTES)
        title = path
        if path.endswith(".md"):
            m = re.search(r"^#\s+(.+)$", text, re.M)
            if m:
                title = m.group(1).strip()
        items.append({
            "id": f"{kind_for(path)}:{path}",
            "kind": kind_for(path),
            "title": title,
            "paths": [path],
            "text": scrub(text, source_file=True, config_file=ext in CONFIG_EXT),
            "truncated": size > MAX_FILE_BYTES,
            "bytes": size,
        })
    return items


def export_commits(repo):
    sep, end = "\x1f", "\x1e"
    fmt = sep.join(["%H", "%h", "%aI", "%s", "%b"]) + end
    raw = git(repo, "log", "--no-merges", f"--format={fmt}")
    items = []
    for chunk in raw.split(end):
        chunk = chunk.strip("\n")
        if not chunk:
            continue
        parts = chunk.split(sep)
        if len(parts) < 5:
            continue
        full, short, date, subject, body = parts[0], parts[1], parts[2], parts[3], parts[4]
        items.append({
            "id": f"commit:{short}",
            "kind": "commit",
            "hash": full,
            "title": scrub(subject),
            "date": date,
            "author": OWNER_LOGIN,  # right only when one person wrote every commit; fixed below
            "text": scrub(body.strip()),
            "paths": [],
        })
    # Paths in a second walk, NUL-separated: "<mark><hash>\0\n<path>\0<path>\0" per commit.
    # Split on NUL first, so a control character inside a name stays inside the name, and take
    # exactly the one newline git puts before a commit's first path.
    mark = "\x01\x01commit:"
    by_hash = {i["hash"]: i for i in items}
    cur, first = None, False
    for field in git(repo, "log", "--no-merges", "--name-only", "-z", f"--format={mark}%H").split("\0"):
        if field.lstrip("\n").startswith(mark):
            cur, first = by_hash.get(field.lstrip("\n")[len(mark):]), True
            continue
        if first and field.startswith("\n"):
            field = field[1:]
        first = False
        if field and cur is not None:
            cur["paths"].append(field)
    authors = set(git(repo, "log", "--format=%ae").split())
    if len(authors) > 1 or OWNER_LOGIN is None:
        # A commit carries an email, not a login, and an email is not exported. With several
        # authors, or with no owner named, nobody can be told apart, so nobody is named.
        if len(authors) > 1:
            report["commit_authors_other_than_owner"] += len(authors) - 1
        for it in items:
            it["author"] = "<user>"
    return items


def merge_paths(repo):
    """PR number -> paths it changed, from local merge commits (no API calls)."""
    out = {}
    for line in git(repo, "log", "--merges", "--format=%H %s").splitlines():
        h, _, subject = line.partition(" ")
        m = re.match(r"Merge pull request #(\d+)", subject)
        if not m:
            continue
        out[int(m.group(1))] = git_names(repo, "diff", "--name-only", f"{h}^1", h)
    return out


# ---------------------------------------------------------------- GitHub API

class GitHub:
    def __init__(self, repo, cache_dir=None):
        self.cache_dir = cache_dir
        if cache_dir:
            os.makedirs(cache_dir, exist_ok=True)
            # The cache is what GitHub sent, before scrubbing. It ignores itself, wherever it is:
            # the file goes INSIDE the cache and never replaces one that is already there.
            ignore = os.path.join(cache_dir, ".gitignore")
            if not os.path.exists(ignore):
                with open(ignore, "w", encoding="utf-8") as fh:
                    fh.write("*\n")
        self.base = f"https://api.github.com/repos/{repo}"
        self.token = os.environ.get("GITHUB_TOKEN")
        self.calls = 0

    def get(self, path, params=None):
        results, page = [], 1
        while True:
            q = dict(params or {}, per_page=100, page=page)
            url = self.base + path + "?" + "&".join(f"{k}={v}" for k, v in q.items())
            cache = None
            if self.cache_dir:
                cache = os.path.join(self.cache_dir, re.sub(r"[^A-Za-z0-9]+", "_", path + f"_{q.get('state','')}_p{page}") + ".json")
                if os.path.exists(cache):
                    with open(cache, encoding="utf-8") as fh:
                        data = json.load(fh)
                    if isinstance(data, dict):
                        return data
                    results.extend(data)
                    if len(data) < 100:
                        return results
                    page += 1
                    continue
            req = urllib.request.Request(url, headers={
                "Accept": "application/vnd.github+json",
                "X-GitHub-Api-Version": "2022-11-28",
                "User-Agent": "lawang-onboard-export",
            })
            if self.token:
                req.add_header("Authorization", "Bearer " + self.token)
            try:
                with urllib.request.urlopen(req, timeout=30) as resp:
                    self.calls += 1
                    remaining = resp.headers.get("X-RateLimit-Remaining")
                    data = json.load(resp)
            except urllib.error.HTTPError as e:
                raise SystemExit(f"GitHub API {e.code} on {path}; set GITHUB_TOKEN if rate limited")
            if cache:
                with open(cache, "w", encoding="utf-8") as fh:
                    json.dump(data, fh)
            if isinstance(data, dict):
                return data
            results.extend(data)
            if remaining is not None and int(remaining) < 3:
                raise SystemExit("GitHub rate limit nearly exhausted; set GITHUB_TOKEN and rerun")
            if len(data) < 100:
                return results
            page += 1


def export_github(gh, local_merge_paths):
    prs, reviews, issues = [], [], []
    for pr in gh.get("/pulls", {"state": "all"}):
        n = pr["number"]
        paths = local_merge_paths.get(n)
        if paths is None:
            paths = [f["filename"] for f in gh.get(f"/pulls/{n}/files")]
        prs.append({
            "id": f"pr:{n}",
            "kind": "pr",
            "number": n,
            "title": scrub(pr["title"]),
            "state": "merged" if pr.get("merged_at") else pr["state"],
            "date": pr["created_at"],
            "merged_at": pr.get("merged_at"),
            "branch": pr["head"]["ref"],
            "author": login(pr["user"]),
            "labels": [l["name"] for l in pr.get("labels", [])],
            "text": scrub(pr.get("body") or ""),
            "paths": paths,
        })
        for r in gh.get(f"/pulls/{n}/reviews"):
            if not (r.get("body") or "").strip():
                continue
            reviews.append({
                "id": f"review:{n}:{r['id']}",
                "kind": "review",
                "pr": n,
                "title": f"Review on PR #{n} ({r['state'].lower()})",
                "date": r.get("submitted_at"),
                "author": login(r["user"]),
                "text": scrub(r["body"]),
                "paths": paths,
            })
    pr_paths = {p["number"]: p["paths"] for p in prs}
    for c in gh.get("/pulls/comments"):
        n = int(c["pull_request_url"].rsplit("/", 1)[1])
        reviews.append({
            "id": f"review_comment:{n}:{c['id']}",
            "kind": "review",
            "pr": n,
            "title": f"Review comment on PR #{n}, {c['path']}",
            "date": c["created_at"],
            "author": login(c["user"]),
            "text": scrub(c["body"]),
            "paths": [c["path"]],
        })
    for c in gh.get("/issues/comments"):
        n = int(c["issue_url"].rsplit("/", 1)[1])
        is_pr = n in pr_paths
        reviews.append({
            "id": f"{'pr' if is_pr else 'issue'}_comment:{n}:{c['id']}",
            "kind": "review" if is_pr else "issue_comment",
            ("pr" if is_pr else "issue"): n,
            "title": f"Comment on {'PR' if is_pr else 'issue'} #{n}",
            "date": c["created_at"],
            "author": login(c["user"]),
            "text": scrub(c["body"]),
            "paths": pr_paths.get(n, []),
        })
    for i in gh.get("/issues", {"state": "all"}):
        if "pull_request" in i:
            continue
        issues.append({
            "id": f"issue:{i['number']}",
            "kind": "issue",
            "number": i["number"],
            "title": scrub(i["title"]),
            "state": i["state"],
            "date": i["created_at"],
            "author": login(i["user"]),
            "labels": [l["name"] for l in i.get("labels", [])],
            "milestone": (i.get("milestone") or {}).get("title"),
            "text": scrub(i.get("body") or ""),
            "paths": [],
        })
    return prs, reviews, issues


# ---------------------------------------------------------------- main

PLACEHOLDER_ONLY = re.compile(r"^[@\s]*(<[a-z-]+>[\s]*)+$")


def write(out, name, items):
    before = len(items)
    items[:] = [i for i in items if i["kind"] not in ("review", "issue_comment") or not PLACEHOLDER_ONLY.match(i.get("text") or "")]
    if before != len(items):
        report["dropped_placeholder_only_" + name.split(".")[0]] += before - len(items)
    items.sort(key=lambda x: x["id"])
    with open(os.path.join(out, name), "w", encoding="utf-8") as fh:
        for it in items:
            fh.write(json.dumps(it, ensure_ascii=False, sort_keys=True) + "\n")
    return len(items)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("repo_path")
    ap.add_argument("github_repo", nargs="?",
                    help="owner/name on GitHub; leave out to export files and commits only")
    ap.add_argument("--owner", help="GitHub login kept as is in the export (default: the owner in owner/name)")
    ap.add_argument("--out", default="corpus")
    ap.add_argument("--cache", help="raw, UNSCRUBBED API responses; delete to refetch "
                                    "(default: .<out>-raw/github beside <out>, never inside it)")
    ap.add_argument("--rescrub-github", action="store_true",
                    help="do not call GitHub; re-apply scrubbing to existing prs/reviews/issues.jsonl")
    a = ap.parse_args()
    out_abs = os.path.abspath(a.out)
    a.cache = a.cache or os.path.join(os.path.dirname(out_abs), "." + os.path.basename(out_abs) + "-raw", "github")
    if a.rescrub_github and not a.github_repo:
        raise SystemExit("--rescrub-github needs <owner/repo>")
    if not a.github_repo:
        for name in ("prs.jsonl", "reviews.jsonl", "issues.jsonl"):
            path = os.path.join(a.out, name)
            if os.path.exists(path) and os.path.getsize(path) > 0:
                raise SystemExit(f"{path} holds an earlier GitHub export and no <owner/repo> was given; "
                                 "name the repository, or choose another --out")
    os.makedirs(a.out, exist_ok=True)
    global OWNER_LOGIN
    OWNER_LOGIN = a.owner or (a.github_repo.split("/", 1)[0] if a.github_repo else None)

    files = export_files(a.repo_path)
    commits = export_commits(a.repo_path)
    gh = GitHub(a.github_repo, a.cache) if a.github_repo else None
    if gh is None:
        print("no <owner/repo> given: pull requests, reviews and issues are not exported", file=sys.stderr)
        prs, reviews, issues = [], [], []
    elif a.rescrub_github:
        def load(name):
            with open(os.path.join(a.out, name), encoding="utf-8") as fh:
                items = [json.loads(l) for l in fh if l.strip()]
            for it in items:
                it["title"] = scrub(it.get("title"))
                it["text"] = scrub(it.get("text"))
            return items
        prs, reviews, issues = load("prs.jsonl"), load("reviews.jsonl"), load("issues.jsonl")
    else:
        prs, reviews, issues = export_github(gh, merge_paths(a.repo_path))

    write(a.out, "files.jsonl", files)
    write(a.out, "commits.jsonl", commits)
    write(a.out, "prs.jsonl", prs)
    write(a.out, "reviews.jsonl", reviews)
    write(a.out, "issues.jsonl", issues)

    counts = collections.Counter()
    for group in (files, commits, prs, reviews, issues):
        for it in group:
            counts[it["kind"]] += 1
    head = git(a.repo_path, "rev-parse", "--short", "HEAD").strip()
    manifest = {
        "source": a.github_repo or os.path.basename(os.path.abspath(a.repo_path)),
        "source_commit": head,
        "counts": dict(sorted(counts.items())),
        "scrubbed": dict(sorted(report.items())),
        "github_api_calls": gh.calls if gh else 0,
    }
    with open(os.path.join(a.out, "manifest.json"), "w", encoding="utf-8") as fh:
        json.dump(manifest, fh, indent=2, sort_keys=True)
        fh.write("\n")
    print(json.dumps(manifest, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
