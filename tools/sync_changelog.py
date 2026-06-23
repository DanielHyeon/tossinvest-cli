#!/usr/bin/env python3
"""Sync CHANGELOG.md → docs Changelog pages (single source of truth).

- content/docs/changelog.mdx     (ko): full customer-facing changelog body
- content/docs/changelog.en.mdx  (en): version index table linking to GitHub releases

Run on release (after finalizing CHANGELOG.md): python3 tools/sync_changelog.py
"""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "CHANGELOG.md"
KO = ROOT / "website-fumadocs" / "content" / "docs" / "changelog.mdx"
EN = ROOT / "website-fumadocs" / "content" / "docs" / "changelog.en.mdx"
RELEASES = "https://github.com/JungHoonGhae/tossinvest-cli/releases"

HEADING = re.compile(r"^## \[(?P<ver>[^\]]+)\](?:\s*-\s*(?P<date>.+))?\s*$")


def main() -> None:
    text = SRC.read_text()
    lines = text.splitlines()

    # Body = everything from the first "## [" heading onward.
    start = next((i for i, ln in enumerate(lines) if ln.startswith("## [")), len(lines))
    body = "\n".join(lines[start:]).strip()

    KO.write_text(
        "---\n"
        "title: 변경 이력\n"
        "description: tossctl 버전별 변경 사항 (사용자 관점)\n"
        "---\n\n"
        "버전별 변경 사항입니다. 각 릴리즈의 바이너리·릴리즈 노트는 "
        f"[GitHub Releases]({RELEASES})에서도 볼 수 있습니다.\n\n"
        f"{body}\n"
    )

    # English: a version index table (language-neutral) linking to each release.
    rows = []
    for ln in lines:
        m = HEADING.match(ln)
        if not m:
            continue
        ver = m.group("ver").strip()
        date = (m.group("date") or "").strip()
        if ver.lower() == "unreleased":
            rows.append(f"| Unreleased | — | [main]({RELEASES}) |")
        else:
            tag = f"v{ver}"
            rows.append(f"| {ver} | {date or '—'} | [{tag}]({RELEASES}/tag/{tag}) |")

    EN.write_text(
        "---\n"
        "title: Changelog\n"
        "description: tossctl version history\n"
        "---\n\n"
        "Version history. Full per-release notes (and binaries) are on "
        f"[GitHub Releases]({RELEASES}). The detailed changelog is written in Korean — "
        "see the [Korean changelog](/docs/changelog).\n\n"
        "| Version | Date | Release |\n"
        "| --- | --- | --- |\n"
        + "\n".join(rows)
        + "\n"
    )

    print(f"wrote {KO.relative_to(ROOT)} and {EN.relative_to(ROOT)} ({len(rows)} versions)")


if __name__ == "__main__":
    main()
