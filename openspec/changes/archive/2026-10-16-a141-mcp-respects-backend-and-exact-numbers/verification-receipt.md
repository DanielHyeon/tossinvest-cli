# a141 full verification receipt

Verified at commit `c3e15cf2cdf276918a8906e24bcecb94fa646f76` in
`/tmp/tossos-compat-20261015`. This validation run changed no production or
test source, runtime/service configuration, order state, tokens, or ledgers.

| Command | Result | Raw log |
| --- | --- | --- |
| `make test` | exit 0 | `/tmp/a141-full-final/make-test.log` |
| `make test-seams` | exit 0 | `/tmp/a141-full-final/make-test-seams.log` |
| `make test-race` | exit 0 | `/tmp/a141-full-final/make-test-race.log` |
| `make sdd-check` | exit 0; CodeGraph hard-evidence index matched worktree | `/tmp/a141-full-final/make-sdd-check.log` |
| `make sdd-sync` | exit 0; CodeGraph full sync and CodeGraphContext completed | `/tmp/a141-full-final/make-sdd-sync.log` |

`make test` included the long-running `tools/a112-mb-us-source` package
(374.308s). `make test-race` passed its engine subset (323.476s). `make
test-seams` included `tools/a112-mb-us-source` (449.134s).

## Advisory indexing exception

The `make sdd-sync` process itself returned exit 0 and printed
`[sdd-sync] all indexes current`, but its GBrain advisory phase also reported
one open parse failure. It is not a complete-all-indexes result:

- `openspec/changes/a063-align-attestation-renewal-profile/analysis/execution-baseline-ledger.json`
- `FILE_TOO_LARGE`, 6,469,224 bytes
- recorded in `.sdd/gbrain-home/.gbrain/sync-failures.jsonl` at commit
  `c3e15cf2cdf276918a8906e24bcecb94fa646f76`, attempts 2

No `--skip-failed`, source change, or corrective reindex was used. A final
`make sdd-sync` is required after this receipt because the receipt changes the
worktree fingerprint.


## Aggregate gate — 2026-10-16

`make gate CHANGE=a141-mcp-respects-backend-and-exact-numbers` passed all 11 stages. Raw exit `/tmp/a141-full-final/make-gate-a141.exit` is 0; log `/tmp/a141-full-final/make-gate-a141.log` ends with GATE PASS. Strict validation: 49 passed, 0 failed. Manager inspected these raw records independently. Post-receipt sync ran before the aggregate gate. No product/test source changed during verification.
