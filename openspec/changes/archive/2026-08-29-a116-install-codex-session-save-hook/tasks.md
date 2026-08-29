## 1. Contract Freeze

- [x] 1.1 Validate the a116 OpenSpec contract strictly and record the lightweight four-voice proposal review.
- [x] 1.2 Record CodeGraph hard evidence, the empty CodeGraphContext result, and the Function Logic Map not-applicable rationale.

## 2. RED Tests

- [x] 2.1 Add failing tests for Codex-only path isolation, bounded transcript extraction/redaction, malformed input, atomic backup retention, and non-blocking concurrency.
- [x] 2.2 Add a failing hook-configuration test that preserves the existing SDD handler and matches Codex Bash/apply_patch calls asynchronously.

## 3. GREEN Implementation

- [x] 3.1 Implement the standard-library-only Codex session saver under `.codex/hooks/`.
- [x] 3.2 Install the new PostToolUse matcher alongside the existing matcher and ignore `.codex-context/` generated data.

## 4. Verification

- [x] 4.1 Run focused unit tests, a synthetic end-to-end hook invocation, JSON/Python validation, and PM/OpenSpec validation.
- [x] 4.2 Verify `.claude/settings.json` and `save-session.sh` were not modified by a116 and document the required `/hooks` trust step.
- [x] 4.3 Run `make sdd-sync`, `make sdd-check`, `make gate CHANGE=a116-install-codex-session-save-hook`, and record independent review results.
