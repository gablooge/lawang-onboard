# Deploy

The hosted demo is at https://lawang-onboard.samsulhadi.com. It is served by two containers
defined in [`compose.yaml`](../compose.yaml):

| Service | What it runs | Network |
|---|---|---|
| `onboard` | The Go server (`/`, `/api`, `/mcp`) in a distroless image, as nonroot, with a read-only root filesystem, no capabilities and `no-new-privileges`. | Private compose network. Also bound to `127.0.0.1:8080` on the host, for local testing only. |
| `cloudflared` | The Cloudflare Tunnel connector. It dials out to Cloudflare, so no inbound port is opened anywhere. | Private compose network, no ports. |

## The tunnel

The tunnel is managed from the Cloudflare dashboard (Zero Trust > Networks > Tunnels), not from a
local config file. It has one public hostname:

| Public hostname | Service |
|---|---|
| `lawang-onboard.samsulhadi.com` | `http://onboard:8080` |

`onboard` is the compose service name, which resolves on the private network. Use a tunnel of
its own for this demo, never one that already serves another app: every connector on a tunnel
receives traffic for every hostname on it.

## Secrets

Two files, both gitignored, never committed and never printed:

- `.env`: the demo role tokens `ONBOARD_TOKEN_MAINTAINER`, `ONBOARD_TOKEN_EMPLOYEE`,
  `ONBOARD_TOKEN_CONTRACTOR` (see [`.env.example`](../.env.example)).
- `.env.tunnel`: `TUNNEL_TOKEN=<token from the dashboard>`.

Check them without showing values:

```sh
git check-ignore .env .env.tunnel
grep -c '^ONBOARD_TOKEN_' .env      # 3
grep -c '^TUNNEL_TOKEN=' .env.tunnel # 1
```

## Start and stop

```sh
make up     # docker compose up -d --build
make ps     # both services should be running
make logs   # docker compose logs --tail=100 -f
make down   # stop and remove both containers
```

Smoke test:

```sh
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/          # 200
curl -s http://127.0.0.1:8080/api/roles                                   # the three roles
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://127.0.0.1:8080/mcp # 401 without a token
```

Run the same three against `https://lawang-onboard.samsulhadi.com` once the tunnel shows healthy.

## The Mac must stay awake

While this Mac is the host, it must not sleep. Run this in a separate terminal and leave it open:

```sh
caffeinate -dimsu
```

## Moving to a Linux VPS

The same compose file runs unchanged on any Linux host with Docker:

1. Copy the repository to the VPS (`git clone`).
2. Create `.env` and `.env.tunnel` there by hand, with the same variable names. Do not copy them
   over chat or commit them.
3. `make up`, then the smoke test above on the VPS.
4. On the Mac, stop its connector so only one connector serves the tunnel:
   `docker compose stop cloudflared` (or `make down`).
5. Check the public URL again.

The VPS needs no inbound ports: the tunnel only makes outbound connections.

## Connecting IBM Bob to the hosted server

Copy [`.bob/mcp.remote.example.json`](../.bob/mcp.remote.example.json) to `.bob/mcp.json`
(gitignored) and replace the placeholder with the contractor token, which Samsul hands out
privately.
