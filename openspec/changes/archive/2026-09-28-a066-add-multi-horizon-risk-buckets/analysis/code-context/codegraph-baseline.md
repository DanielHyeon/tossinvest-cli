# CodeGraph baseline — a066 Wave 2A (2026-09-25)

- Tool: CodeGraph **1.6.0**. Index after `make sdd-sync` (CodeGraph phase: "Already up to date"):
  2,162 files / 38,720 nodes / 147,446 edges at HEAD `d72bc401`; 2,169 / 38,796 / 147,695 when re-read at
  `0233d776` (other sessions committed in between; the a066 symbols below did not move).
- Change base (unchanged): `23794f86` (2026-08-04). Queries run at HEAD `d72bc401`…`648df8ef`.
- Callers/callees counted with `codegraph callers|callees <sym> --limit 500 --json` — the default limit of 20
  silently truncates (`RecordFill` reads "20" by default, 147 uncapped).

## Symbols (task 1.1: Guardian evaluation/issuance, reservation tx, Gateway revalidation, fill apply, mode/loss locks)

| Area | Symbol | Definition (HEAD) | Callers prod / test | Production callers | Callees | Impact |
|---|---|---|---|---|---|---|
| Guardian precheck | `RiskGuardian.PrecheckQFinalEntry` | `internal/execgw/riskguardian_qfinal.go:82` | 2 / 0 | `PrecheckQFinalCampaignFirstLeg`, `IssueQFinalEntry` | 20 | 4 |
| Guardian issuance | `RiskGuardian.IssuePrecheckedQFinalEntry` | `internal/execgw/riskguardian_qfinal.go:177` | 1 / 0 | `IssueQFinalEntry` | 25 | 2 |
| reservation tx | `Journal.RecordQFinalDecisionAndReserve` | `internal/journal/risk_bucket_issuance.go:70` | 1 / 5 | `RecordQFinalDecisionAndReserveWithRecollection` | 19 | 12 |
| reservation tx | `Journal.CommitRiskBucketAdmission` | `internal/journal/risk_bucket.go:88` | 0 / 22 | none (test-only entry) | 23 | 41 |
| Gateway revalidation | `Gateway.checkReservation` | `internal/execgw/gateway.go:889` | 1 / 0 | `Gateway.submit` | 8 | 5 |
| Gateway revalidation | `Journal.RevalidateQFinalAdmission` | `internal/journal/risk_bucket_issuance.go:558` | 1 / 2 | `Gateway.checkReservation` | 13 | 5 |
| broker boundary | `Gateway.submit` | `internal/execgw/gateway.go:466` | 3 / — (see reconciliation) | `place`, `Cancel`, `Amend` | 70 | 34 |
| fill apply | `Journal.RecordFill` | `internal/journal/fills.go:313` | 1 / 146 | `filldetect.Apply` (`internal/filldetect/ledger.go:36`) | 34 | 246 |
| fill apply | `Journal.runApplyHooks` | `internal/journal/apply_hook.go:264` | 2 / 2 | `RecordFill`, `backfillConfirmedStrategyFillTx` | 5 | 153 |
| fill accounting | `riskbucket.ApplyFill` | `internal/riskbucket/fill.go:96` | 2 / 13 | `applyRiskBucketFillInTx`, `completeRiskBucketFillActual` | 23 | 21 |
| owner lifecycle | `bindRiskBucketOwnerActualInTx` | `internal/journal/risk_bucket_owner.go:381` | 2 / 0 | `bindRiskBucketOwnerActual`, `applyRiskBucketOwnerBindingInTx` | 10 | 17 |
| owner lifecycle | `Journal.ReleaseRiskBucketOwner` | `internal/journal/risk_bucket_owner.go` | 0 / 8 | none — official broker-zero mint intentionally absent | 19 | 9 |
| entry latch | `riskbucket.EntryBlocked` | `internal/riskbucket/fill.go:50` | 0 / 3 | none | 2 | 4 |
| mode/entry gate | `Gateway.checkEntry` | `internal/execgw/gateway.go:855` | 1 / 0 | `Gateway.submit` | 3 | 5 |
| mode/entry gate | `EntryGate.CheckEntryFor` | `internal/execgw/symbolgate.go:229` | 4 / 57 | `EntryLatchFor`, `checkEntry`, `CheckEntry`, `ObserveStrategyEntryGate` | 5 | 120 |
| daily loss | `risk.checkDailyLoss` | `internal/risk/chain.go:518` | 1 / 0 | `entryChain` | 12 | 5 |
| horizon/market loss lock | — | **no symbol exists** (`grep -rniE 'loss_?lock' --include=*.go internal cmd`, excluding this lot's tagged RED test: 0 hits) | — | — | — | — |

## Affected tests

`codegraph affected internal/journal/fills.go internal/execgw/gateway.go internal/journal/apply_hook.go
internal/riskbucket/fill.go internal/journal/risk_bucket.go internal/journal/reconcile_states.go`:

- default matcher: **1** file (`auth-helper/tests/test_cli.py`) — the known Go `_test.go` blind spot (.claude/CLAUDE.md).
- `--filter '*_test.go'`: **880** files (117 `internal/journal`, 97 `internal/app`, 80 `cmd/tossctl`, 38 `internal/execgw`, …).
