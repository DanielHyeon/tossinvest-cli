# a066 5.5 relaxation — adversarial review brief (read-only)

Repository: /mnt/D/Axipient/workspace/TossOS (TossOS, a live-money auto-trading product). Branch feat/a112-four-family-runtime.
Review target: the three commits bce793a7, 75d9073b and 90e5170d (`git -C /mnt/D/Axipient/workspace/TossOS show <sha>`). The
lot's diff is `git -C /mnt/D/Axipient/workspace/TossOS diff 0c12844a 90e5170d -- internal cmd openspec/changes/a066-add-multi-horizon-risk-buckets`.
Other sessions commit on this branch concurrently: judge only these three commits and the files they touch.

## What the lot claims (design: openspec/changes/a066-add-multi-horizon-risk-buckets/design.md, section D8)

User-approved principles: automatic paths only tighten. A relaxation or release requires actor OPERATOR, a human approval
reference string, and an Auditor audit line written BEFORE the journal commit. When tightenings race, the conservative
side wins. The entry points are the journal API and tossctl `mutating: true` commands (never auto-run; no console button).
The lot is mechanism only: nothing is ever run on the operating journal, and nothing is activated or toggled.

1. Entry loss lock release (journal v35, `internal/journal/risk_bucket_relaxation_v35.sql`, `risk_bucket_relaxation.go`,
   `risk_bucket_entry_loss_lock.go`).
   - An activation onto an already-open lock writes a REAFFIRM event.
   - A release binds LockSeq and ExpectedLastEvent; any mismatch is refused as stale.
   - The "at most one open lock per scope" trigger replaces v33's first-cause-wins trigger via DROP/CREATE inside v35.
2. RISK_OVERAGE latch release.
   - It binds ExpectedStateDigest (the latest `risk_bucket_state_snapshots.state_digest`).
   - It clears only the owner and reservation `risk_overage_latched` flags. UNKNOWN_ACTUAL_RISK and overage amounts stay.
   - It reseals the state and writes an audit line before commit.
   - The next over-limit fill latches again.
   - Order: the latch release comes first, then the owner release.
3. Path (Manager ruling 2026-09-29, a092 "relaxation command family" contract).
   - tossctl `engine entry-lock-release` / `engine risk-latch-release` call the engine's authenticated loopback control
     endpoint: the position-policy server gains new routes through an optional capability, the same way as a079's
     quarantine release.
   - The engine process writes with its own journal handle and its own audit log. The journal is single-writer, and the
     first commit bce793a7 wrongly opened the journal from the CLI; 90e5170d fixes that.
   - When the engine is not running, the CLI refuses.
   - After commit, the engine enqueues an `engine.risk_relaxation` alert. If that fails, the release stands and the CLI
     reports 「완화됨·통지 실패」 with a non-zero exit.
   - `engine risk-latch-show` reads read-only (`journal.OpenReadOnly`).
   - No engine lock is taken, because stopping the engine removes stop-loss.
4. Hard safety invariant: a066 locks and latches may block only EXPOSURE_RAISING entries. Nothing in this lot may delay,
   refuse or abort a stop-loss, an emergency exit, reconciliation, or fill detection. Toggle OFF must equal upstream.

## Rules for you (non-negotiable)

- READ-ONLY on the repository. Never write, commit, stage, checkout, stash, or run `git config`, `make gate` or `make image`
  in /mnt/D/Axipient/workspace/TossOS. Never run any `mutating: true` tossctl command against a real journal. Never touch
  ~/.config/tossctl or ~/.local/share.
- Probes go only in a throwaway copy under your own temp directory. Start every probe script with `set -euo pipefail`. Use
  `git -C <absolute path>` only, and assert `git rev-parse --show-toplevel` equals your copy before any git write inside it.
  Delete the copy afterwards. At the end, run `git -C /mnt/D/Axipient/workspace/TossOS status --short -- internal cmd` and
  report it; it must be unchanged by you.
- Go: `go test` inside your copy with `GOFLAGS=-trimpath GOCACHE=<your temp>/gocache`.
- Report findings as: severity (P0 = can relax or open exposure without operator + approval + audit, or can delay or
  refuse a stop or exit; P1 = wrong result a real operator will hit; P2 = latent; P3 = note), file:line, the concrete
  scenario, and a probe (a command or test and its observed output) when you claim a behaviour. Say "not probed" when you
  did not probe. Do not propose rewrites beyond the fix direction.
