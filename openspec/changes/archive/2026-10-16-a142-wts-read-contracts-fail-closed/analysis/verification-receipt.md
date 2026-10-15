# a142 focused verification receipt

## RED: baseline behavior

Command executed before the decoder implementation:

```text
go test ./internal/client -run 'TestListPendingOrders(Rejects|Accepts)|TestListCompletedOrdersRangeRejects|TestListWatchlist(Uses|Rejects)|TestWatchlistMutation' -count=1
```

Exit: `1`.

Observed failures included `TestListPendingOrdersRejectsAbsentOrNullResult` (absent/null returned nil error), `TestListCompletedOrdersRangeRejectsMissingCollections` and `TestListCompletedOrdersRangeRejectsPartialMarketResult` (nil collection returned nil error), and the dedicated watchlist tests because the baseline still searched the legacy dashboard section. This RED output was observed in the implementation terminal; it was not retained as a file.

## Retrospective baseline verification

To retain a reviewable current artifact without relabeling it as the historical TDD run, baseline `b36798f0` production files were checked out in a repository-external worktree and the current three affected client test files were overlaid. The focused command below exited `1`; raw output is `/tmp/tossos-a142-retrospective-baseline.log`.

```text
go test ./internal/client -run 'Test(ListPendingOrders|HasPendingOrder|FindOrderRejectsUnknownPendingCollection|ListCompletedOrdersRange|ListWatchlist)' -count=1
```

The log records the intended baseline failures: nil collections became successful empty reads, an accumulated US order was returned after a KR schema failure, the legacy watchlist adapter rejected the dedicated fixture, and `FindOrder` reached completed history after malformed pending data.

## GREEN: final focused fixtures

```text
go test ./internal/client -run 'Test(ListPendingOrders|HasPendingOrder|FindOrderRejectsUnknownPendingCollection|ListCompletedOrdersRange|ListWatchlist|WatchlistMutation)' -count=1
```

Exit: `0` (`ok github.com/JungHoonGhae/tossinvest-cli/internal/client`). This covers explicit empty arrays, missing/null/wrong-type failures, nonempty partial all-market failures, stable watchlist ordering/ties and retained fields, consumer error propagation, and unchanged watchlist mutation contracts. The command was last run after the equal-tie fixture repair.

The same final command was rerun after the retrospective check, exited `0`, and its raw output is `/tmp/tossos-a142-final-focused-green.log`.

```text
go test ./cmd/tossctl -run 'Watchlist|Order.*(List|Pending|Completed)|FindOrder' -count=1
```

Exit: `0` (`ok github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl`).

```text
python3 tools/logic-map/check_analysis.py --change a142-wts-read-contracts-fail-closed
```

Exit: `0` (`evidence complete or diff-proven exempt`).

## Independent guard mutants

Each guard was independently changed to `if false` in a repository-external temporary worktree. All corresponding fixture commands exited `1`:

| Guard | Raw log | Failure asserted |
|---|---|---|
| pending `result` | `/tmp/tossos-a142-pending.log` | absent/null no longer error |
| completed `result.body` | `/tmp/tossos-a142-completed.log` | missing/null and partial market no longer error |
| bulk `watchlists` | `/tmp/tossos-a142-watchlists.log` | missing/null no longer error |
| per-group `items` | `/tmp/tossos-a142-items.log` | missing items no longer error |

These logs are local temporary evidence, not tracked source inputs. No full aggregate suite was run at this stage; it remains the final-gate responsibility.

## Aggregate gate — 2026-10-16

Final `make gate CHANGE=a142-wts-read-contracts-fail-closed` PASS, all 11 stages including test/test-seams/test-race/vet/validate. Raw log `/tmp/a142-full-final/make-gate-rerun.log`, exit `/tmp/a142-full-final/rerun-exit-code.txt` = 0. Strict validation 49 passed/0 failed; engine race subset 270.331s. Terra observed execution; Manager independently read the final receipt.

The first gate stopped at hard CodeGraph freshness (exit 2), before full tests; see `/tmp/a142-full-final/make-gate.log`. An unfinished initial sync and a redundant remediation sync were both allowed to terminate before retry. Remediation sync exit 0 is retained in `/tmp/a142-full-final/sdd-sync-remediation.exit`; final gate hard freshness passed. No services/settings were changed. This does not assert full advisory knowledge-index coverage.
