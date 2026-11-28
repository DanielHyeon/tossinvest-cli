# CodeGraph hard evidence — pre-edit

`make sdd-sync` completed successfully on 2026-09-05 before these queries.

| Symbol | Definition and binding | Direct callers | Impact evidence |
|---|---|---|---|
| `resolveSoakRecord` | `cmd/tossctl/soak.go:568`; explicit `root.configDir` joins `soak.FileName`, otherwise `journal.DataDir()` | `runSoakRun`, `loadSoakSummary`, `runConsole` | profile mismatch affects survey, status, attestation, and console paths |
| `resolveSoakAttestationPath` | `cmd/tossctl/soak.go:585`; config setting wins, then explicit config directory, then default config directory | `runSoakAttest`, `runConsole` | a shared sibling diagnostic path can derive from this resolved attestation path without new config plumbing |
| `runSoakAttest` | `cmd/tossctl/soak.go:470` | Cobra `soak attest` command | builds and saves the interlock input only after existing qualification passes |
| `Console.readAttestation` | `internal/console/data.go:259` | `Console.snapshotFor` | reads local metadata for the existing capability-attestation section |

The tracked repository has no `tossos-attest.service` or `tossos-attest.timer` template at current HEAD. The installed user unit is external operational state and is not treated as source truth.

## Proposed path binding

For opt-in diagnostics, derive the path from the full `resolveSoakAttestationPath(root, "")` result and append `.renewal-status.json`. The empty override is intentional: recording mode rejects `--record` and `--out`, so both the writer and `runConsole` bind to the same resolver-selected profile, including a configured custom attestation location. Full-path derivation keeps two profile-specific attestation basenames in one directory from colliding. The console derives the same path from the `Attestation` option. No legacy data-profile record is read, copied, moved, or relabeled.
