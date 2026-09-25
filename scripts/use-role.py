#!/usr/bin/env python3
"""Switch the role Bob connects as, for local demos.

Usage: scripts/use-role.py contractor|employee|maintainer

Rewrites ONBOARD_TOKEN in .bob/mcp.json (gitignored) with the value of ONBOARD_TOKEN_<ROLE> from
.env (gitignored). Prints the role name only, never a token. Reload MCP servers in Bob afterwards.
"""
import json
import re
import sys

ROLES = ("contractor", "employee", "maintainer")
if len(sys.argv) != 2 or sys.argv[1] not in ROLES:
    sys.exit("usage: scripts/use-role.py " + "|".join(ROLES))
role = sys.argv[1]
var = "ONBOARD_TOKEN_" + role.upper()
env = {}
with open(".env") as fh:
    for line in fh:
        m = re.match(r"^(ONBOARD_TOKEN_[A-Z]+)=(.*)$", line.strip())
        if m:
            env[m.group(1)] = m.group(2)
if not env.get(var):
    sys.exit(var + " is not set in .env")
with open(".bob/mcp.json") as fh:
    cfg = json.load(fh)
cfg["mcpServers"]["lawang-onboard"]["env"]["ONBOARD_TOKEN"] = env[var]
with open(".bob/mcp.json", "w") as fh:
    json.dump(cfg, fh, indent=2)
print("Bob will connect as " + role + ". Reload MCP servers in Bob.")
