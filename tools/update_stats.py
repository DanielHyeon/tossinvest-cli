#!/usr/bin/env python3
"""Generate STATS.md — daily stars (backfilled from launch) + release downloads.

Stars are recomputed every run from each stargazer's `starredAt`, so the series
is always accurate back to launch. GitHub only exposes *cumulative* release
download totals (no historical dailies), so downloads are logged forward from
the day this script first runs, persisted in docs/.stats-downloads.json.

Usage: python3 tools/update_stats.py   (needs `gh` authenticated)
Run daily via CI to append a fresh row.
"""
from __future__ import annotations

import json
import subprocess
from datetime import date, datetime, timedelta, timezone
from pathlib import Path

REPO = "JungHoonGhae/tossinvest-cli"
ROOT = Path(__file__).resolve().parent.parent
STATS_FILE = ROOT / "STATS.md"
DL_LOG = ROOT / "docs" / ".stats-downloads.json"

STAR_QUERY = """
query($endCursor: String) {
  repository(owner: "JungHoonGhae", name: "tossinvest-cli") {
    stargazers(first: 100, after: $endCursor) {
      pageInfo { hasNextPage endCursor }
      edges { starredAt }
    }
  }
}
"""


def _gh_json(args: list[str]) -> dict:
    out = subprocess.run(["gh", *args], capture_output=True, text=True, check=True).stdout
    return json.loads(out)


def fetch_star_dates() -> list[str]:
    """All stargazer star dates (UTC, YYYY-MM-DD), oldest first."""
    dates: list[str] = []
    cursor: str | None = None
    while True:
        args = ["api", "graphql", "-f", f"query={STAR_QUERY}"]
        if cursor:
            args += ["-F", f"endCursor={cursor}"]
        conn = _gh_json(args)["data"]["repository"]["stargazers"]
        dates += [e["starredAt"][:10] for e in conn["edges"]]
        if conn["pageInfo"]["hasNextPage"]:
            cursor = conn["pageInfo"]["endCursor"]
        else:
            break
    return sorted(dates)


def fetch_download_total() -> int:
    releases = _gh_json(["api", f"repos/{REPO}/releases?per_page=100"])
    return sum(a.get("download_count", 0) for r in releases for a in r.get("assets", []))


def load_dl_log() -> dict[str, int]:
    if DL_LOG.exists():
        return json.loads(DL_LOG.read_text())
    return {}


def fmt(n: int) -> str:
    return f"{n:,}"


def main() -> None:
    today = datetime.now(timezone.utc).date()
    star_dates = fetch_star_dates()
    if not star_dates:
        raise SystemExit("no stargazers found")

    launch = date.fromisoformat(star_dates[0])

    # cumulative stars per calendar day
    per_day: dict[str, int] = {}
    for d in star_dates:
        per_day[d] = per_day.get(d, 0) + 1

    # download log: record today's cumulative total, keep history
    dl_log = load_dl_log()
    dl_log[today.isoformat()] = fetch_download_total()
    DL_LOG.parent.mkdir(parents=True, exist_ok=True)
    DL_LOG.write_text(json.dumps(dict(sorted(dl_log.items())), indent=2) + "\n")

    total_stars = len(star_dates)
    total_dl = dl_log[today.isoformat()]

    rows: list[str] = []
    cum = 0
    prev_dl: int | None = None
    d = launch
    while d <= today:
        iso = d.isoformat()
        cum += per_day.get(iso, 0)
        delta = per_day.get(iso, 0)
        if iso in dl_log:
            cur = dl_log[iso]
            dl_cell = f"{fmt(cur)} (+{fmt(cur - prev_dl)})" if prev_dl is not None else f"{fmt(cur)}"
            prev_dl = cur
        else:
            dl_cell = "—"
        rows.append(f"| {iso} | {fmt(cum)} (+{fmt(delta)}) | {dl_cell} |")
        d += timedelta(days=1)

    header = f"""# Stats

> Auto-updated daily. **Stars** are backfilled from launch via each stargazer's
> `starredAt`. **Release downloads** are GitHub's cumulative asset totals — GitHub
> exposes no historical dailies, so the download column is recorded forward from
> the day tracking started (earlier rows show `—`).

**⭐ {fmt(total_stars)} stars · ⬇️ {fmt(total_dl)} release downloads · since {launch.isoformat()}**

_Last updated: {today.isoformat()} (UTC)_

| Date | Stars | Release Downloads |
| ---------- | -------------------- | -------------------- |
"""
    STATS_FILE.write_text(header + "\n".join(rows) + "\n")
    print(f"wrote {STATS_FILE.relative_to(ROOT)}: {total_stars} stars, {total_dl} downloads, "
          f"{(today - launch).days + 1} daily rows since {launch}")


if __name__ == "__main__":
    main()
