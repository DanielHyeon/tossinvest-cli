# tossinvest-cli

**An unofficial CLI for driving Toss Securities from AI agents and the terminal.**
The binary is `tossctl`.

It reuses your Toss Securities web session to read accounts, quotes, and transaction
history, and to place limited trades from the terminal. Claude Code · Codex · Cursor ·
OpenClaw · bash · HTTP — any tool or agent drives it through one command surface
(`tossctl`), and humans can use it directly too.

!!! warning "Unofficial project"
    This is **not** an official Toss Securities product. It uses the internal web API
    unofficially, which may violate the Terms of Service. The API can change without
    notice, and the author takes no responsibility for account restrictions, losses, or
    other consequences. Use at your own risk and discretion.

!!! danger "Trading is disabled by default"
    All trading actions are off out of the box. They run only after you explicitly
    enable them in `config.json`, and live orders require a two-step
    `--execute` + `--confirm <token>`. See the [Safety Model](safety.en.md).

## At a glance

- **Covers 100% of the official Open API's (upcoming) read & trade scope** — and goes beyond.
- Features the official API lacks: investor flows, market indices, detailed index quotes,
  Toss AI signals, news briefing, screener, dividend reports, sector movements, community
  rankings, watchlist management, transaction ledger, real-time push, fractional orders,
  dry-run preview, and more.
- Output formats: table · JSON · CSV · SSE — ready for automation and agent pipelines.

See the full comparison in [Support Scope](support-scope.en.md).

## 30-second start

```bash
# Install (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/JungHoonGhae/tossinvest-cli/main/install.sh | sh

# Check env → log in → first read
tossctl doctor
tossctl auth login
tossctl account summary --output json
```

More in [Install](install.en.md) and [Quick Start](quickstart.en.md).

## Using it from an AI agent

This project is **agent-friendly**: every command supports `--output json`, and trading
is guarded by an explicit two-step gate. See the [AI Agent Guide](agents.en.md).

- **Docs for LLMs**: [`/llms.txt`](https://junghoonghae.github.io/tossinvest-cli/en/llms.txt) (curated index) · [`/llms-full.txt`](https://junghoonghae.github.io/tossinvest-cli/en/llms-full.txt) (full text)
