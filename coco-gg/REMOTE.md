# Running `coco-gg` in remote mode

The same plugin binary can run in two deployment modes:

- **local** (default) — spawned by the Wails host's `longrunning.Registry`, loopback-bound, auth gated by the Wails host's `:2014` gateway. This is what `wails dev` and a packaged Wails app do automatically.
- **remote** — standalone process on a public-internet server, bound to a TCP port, with built-in admin-token auth and mobile-session minting. No Wails app required.

The game packages (`games/<id>/`) are identical in both modes — they don't know which deployment is running them.

## Required env vars (remote mode)

| Variable | Required | Example | Description |
|---|---|---|---|
| `MODE` | yes | `remote` | Selects remote-mode runtime. Default `local`. |
| `ADMIN_TOKEN` | yes | `$(openssl rand -hex 32)` | Long-lived operator secret. Gates room CRUD, kick, list, session-minting, and the `/admin/` UI. |
| `PUBLIC_URL` | yes | `https://game.example.com` | Public origin the frontend uses to build QR-code URLs. Must match what mobile players actually see in their browser address bar. (Validated at startup; the frontend derives mode + base URL client-side from `window.location.origin`, so this value is consumed only by the startup log line — but it's still required so the operator is forced to set it consciously.) |
| `PORT` | no | `8080` | TCP port to bind. Default `8080`. |
| `UI_DIR` | no | `./ui` | Directory containing the built React bundle. Default `./ui`. |

Hard-fail with a clear error message if `ADMIN_TOKEN` or `PUBLIC_URL` are unset.

## Endpoints

Public (no auth required):
- `GET /health` — liveness probe. Returns `{"ok":true,"pluginId":"coco-gg"}`.
- `GET /plugins/coco-gg/` — the player-facing React bundle. Phones load this via the QR.

Mobile-session gated (validates `?t=<token>`):
- `GET /ws/games/movement?room=<code>&t=<token>` — game WebSocket. Token comes from `POST /api/admin/sessions`.

Admin gated (`Authorization: Bearer <ADMIN_TOKEN>` or `?admin=<token>`):
- `GET /api/games` — list registered games.
- `POST /api/games/movement/rooms` — create a room. Body: none. Response: `{"code":"ABCDEF"}`.
- `GET /api/games/movement/rooms` — list active rooms and aggregate stats.
- `GET /api/games/movement/rooms/{code}` — single room status.
- `DELETE /api/games/movement/rooms/{code}` — destroy a room.
- `DELETE /api/games/movement/rooms/{code}/players/{playerId}` — kick a player.
- `POST /api/admin/sessions` — mint a mobile-session token. Body: `{"ttlSeconds":3600}` (default 3600).
- `GET /admin/` — operator dashboard (the same React bundle as `/plugins/coco-gg/`, gated by admin token at the network layer; the in-app login is the primary gate).

## TLS

The plugin speaks plain HTTP. Run a reverse proxy in front for TLS. Caddy example:

```caddy
game.example.com {
    reverse_proxy localhost:8080
}
```

Nginx example (note the `Connection: upgrade` headers for WebSocket):

```nginx
server {
    listen 443 ssl;
    server_name game.example.com;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }
}
```

## Operator workflow

1. Generate an admin token: `openssl rand -hex 32`.
2. Build the binary: `make backend` (from `plugins/coco-gg/app/`).
3. Build the frontend: `make frontend` (produces `plugins/coco-gg/ui/`).
4. Start the server (process supervisor or `nohup`):
   ```sh
   MODE=remote \
   PORT=8080 \
   ADMIN_TOKEN=<your-token> \
   PUBLIC_URL=https://game.example.com \
   UI_DIR=/path/to/plugins/coco-gg/ui \
   ./coco-gg
   ```
5. Open `https://game.example.com/admin/` in a browser. The React app asks for the admin token; paste it.
6. Click "Open" on the Movement Arena card → "Create Instance" → "QR" → players scan to join.

## Operational notes

- **Sessions are in-memory.** A restart wipes all mobile-session tokens. Existing QRs become invalid; reissue. Persistence (Redis / SQLite) is a future addition.
- **Rooms are in-memory too.** Same restart caveat.
- **No rate limiting.** Put it at the reverse proxy if you need it.
- **Latency.** The game is 20 Hz server-authoritative with no client-side prediction. Sub-50 ms RTT (LAN) is fine; 100+ ms (transcontinental) will feel sluggish. Client-side prediction is a separate piece of work.
- **Single instance.** The plugin doesn't share rooms across processes. To scale horizontally you'd need a shared room store and consistent room-to-instance routing — out of scope.
