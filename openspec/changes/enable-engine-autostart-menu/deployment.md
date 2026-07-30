# Deployment evidence — 2026-07-31

## Installed state

- Image: `tossos:local`
- Image ID: `sha256:65463e7d9c13d6a4927924c873f540a9d7caf48003456b24a1593b140ba9e8a1`
- Compose project/service: `tossos` / `tossos`
- URL: `https://127.0.0.1:37085`
- Host publish: `127.0.0.1:37085` only
- Config: `/home/daniel/.config/tossctl`
- Data: `/home/daniel/.local/share/tossos`
- Deployment secrets: `/home/daniel/.local/share/tossos-deploy/secrets` (mode 0600 files)

## Runtime proof

- HTTPS `/healthz`: HTTP 200 and body `ok`
- Remote token login with the canonical Origin: HTTP 303
- `/settings`: engine autostart section rendered OFF
- `/optimization`: `COMMON_LADDER_BALANCED`, `COMMON_LADDER_RUNNER`, and
  `COMMON_LADDER_HYBRID_50` rendered
- `docker top`: init plus one `tossctl console`; no `tossctl engine run`
- Container: non-root `1000:1000`, read-only root, all capabilities dropped,
  `no-new-privileges`, 128 PID limit
- Restart policy: `unless-stopped`
- Docker system service: enabled and active
- An actual `docker compose restart --timeout 75 tossos` returned healthy;
  `engine.autostart` remained false and no engine-start log appeared

## TLS

- Self-signed local certificate SAN: `IP:127.0.0.1`, `DNS:localhost`
- Valid: 2026-07-30 19:28:55Z through 2027-07-30 19:28:55Z
- SHA-256 fingerprint:
  `BC:8B:E2:B0:24:97:B0:1C:C4:B0:BA:8E:7C:FD:70:61:BB:E0:BD:B7:AE:5A:3B:00:85:B2:B7:70:26:48:F8:05`

## Remote-access blocker

No VPN interface/address exists on this host at deployment time. The service is
therefore deliberately loopback-only. Mobile access requires a real VPN bind IP,
client CIDR, mobile-resolvable VPN DNS name, and a certificate covering that
name. LAN or wildcard exposure was not substituted.
