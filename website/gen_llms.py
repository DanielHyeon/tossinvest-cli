#!/usr/bin/env python3
"""Generate LLM-friendly artifacts after `mkdocs build`.

For each language it writes, under the built site dir:
  - <stem>.md            raw markdown per page (clean source for LLMs to fetch)
  - llms.txt             curated index (llmstxt.org format): title, summary, sections
  - llms-full.txt        all pages concatenated into one file

We do this with a script (not mkdocs-llmstxt) because that plugin is incompatible
with mkdocs-static-i18n's URI rewriting. stdlib only.

Run from the website/ directory after `mkdocs build`:
    python3 gen_llms.py
"""
import os

DOCS = "docs"
SITE = "site"
BASE = "https://junghoonghae.github.io/tossinvest-cli"

# (section, [(stem, ko_title, en_title)]) — order matters.
STRUCTURE = [
    ("시작하기", "Getting started", [
        ("index", "홈", "Home"),
        ("install", "설치", "Install"),
        ("quickstart", "빠른 시작", "Quick Start"),
    ]),
    ("에이전트", "Agents", [
        ("agents", "AI 에이전트 가이드", "AI Agent Guide"),
    ]),
    ("레퍼런스", "Reference", [
        ("commands", "명령 레퍼런스", "Command Reference"),
        ("support-scope", "지원 범위", "Support Scope"),
        ("safety", "안전 모델", "Safety Model"),
        ("configuration", "설정", "Configuration"),
    ]),
]

SUMMARY = {
    "ko": "토스증권을 AI 에이전트·터미널에서 다루는 비공식 CLI (tossctl) — 공식 Open API(예정)보다 넓은 조회·거래 범위.",
    "en": "An unofficial CLI (tossctl) for driving Toss Securities from AI agents and the terminal — broader read & trade scope than the (upcoming) official Open API.",
}
DESC = {
    "ko": ("tossinvest-cli (binary: tossctl) — 토스증권 웹 세션을 재사용해 조회·제한된 거래를 "
           "터미널/AI 에이전트에서 다루는 비공식 CLI. 공식 Open API 의 조회·거래 범위를 100% 커버하고 "
           "그 너머(수급·시장지수·AI 시그널·배당·업종 등락·커뮤니티 랭킹·실시간 푸시·소수점 주문 등)까지 "
           "제공한다. 거래는 기본 비활성이며 config.json 에서 명시적 허용 필요."),
    "en": ("tossinvest-cli (binary: tossctl) — an unofficial CLI that reuses your Toss Securities web "
           "session to read and place limited trades from the terminal / AI agents. Covers 100% of the "
           "official Open API's read & trade scope and goes beyond (investor flows, indices, AI signals, "
           "dividends, sector moves, community rankings, real-time push, fractional orders, …). Trading is "
           "off by default and must be enabled in config.json."),
}


def src_path(stem, lang):
    return os.path.join(DOCS, f"{stem}.md" if lang == "ko" else f"{stem}.en.md")


def out_dir(lang):
    return SITE if lang == "ko" else os.path.join(SITE, "en")


def md_url(stem, lang):
    # raw .md endpoints are written flat at the (locale) site root
    base = BASE if lang == "ko" else BASE + "/en"
    return f"{base}/{stem}.md"


def gen(lang):
    odir = out_dir(lang)
    os.makedirs(odir, exist_ok=True)
    title_idx = 1 if lang == "ko" else 2

    full_parts = []
    index_lines = ["# tossinvest-cli", "", f"> {SUMMARY[lang]}", "", DESC[lang], ""]

    for sec_ko, sec_en, items in STRUCTURE:
        index_lines.append(f"## {sec_ko if lang == 'ko' else sec_en}")
        index_lines.append("")
        for stem, t_ko, t_en in items:
            sp = src_path(stem, lang)
            if not os.path.exists(sp):
                continue
            raw = open(sp, encoding="utf-8").read()
            # per-page raw .md endpoint
            with open(os.path.join(odir, f"{stem}.md"), "w", encoding="utf-8") as f:
                f.write(raw)
            title = (t_ko, t_en)[title_idx - 1]
            index_lines.append(f"- [{title}]({md_url(stem, lang)})")
            full_parts.append(raw.strip())
        index_lines.append("")

    with open(os.path.join(odir, "llms.txt"), "w", encoding="utf-8") as f:
        f.write("\n".join(index_lines).rstrip() + "\n")
    with open(os.path.join(odir, "llms-full.txt"), "w", encoding="utf-8") as f:
        f.write("\n\n---\n\n".join(full_parts) + "\n")
    print(f"[{lang}] wrote {odir}/llms.txt, llms-full.txt, and per-page .md")


def main():
    if not os.path.isdir(SITE):
        raise SystemExit("site/ not found — run `mkdocs build` first")
    for lang in ("ko", "en"):
        gen(lang)


if __name__ == "__main__":
    main()
