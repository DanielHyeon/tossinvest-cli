# Branch Test Map: `aggregateClosedKRXFiveMinute`

- Source: `internal/strategymarket/bars.go`, post-edit SHA-256 `aa6c93cd24d872739e99ed9576fca8e80b3256c1b772f23737bc39fe39faab06` (pre-edit `a6f17d0cc8ede1b51d2e96344d124675d9022ed730142c50b9d06a6cc5cd9269`); branch IDs follow `ast.json`, regenerated 2026-08-18 after the edit.
- AST counts: branches 18, returns 13, calls 37, defers 0, go statements 0 — unchanged by the edit. Source range `131:1`-`228:2`.
- a117 edits B7, B12 and the two instants in the final return. Every other branch keeps its exact refusal kind, and the table records the test that holds it there.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fewer than five minutes | `TestAggregateClosedKRXFiveMinuteFailsClosed/missing` | not edited | green |
| B2 | zero `now` | reachable only through the unexported path; no direct case | not edited | green (unchanged) |
| B3 | `LoadLocation` failure | not-applicable: unreachable in a supported environment; no test drives it | not edited | not-applicable |
| B4 | per-candle loop | every case in the table below and above | not edited | green |
| B5 | naive timestamp literal | `TestAggregateClosedKRXFiveMinuteFailsClosed/naive` | not edited | green |
| B6 | offset present but unparseable | not-applicable: the regexp at B5 rejects every literal `time.Parse` would reject | not edited | not-applicable |
| B7 | the session window in label space, tested in both directions: label `09:00` covers the pre-open minute `08:59`-`09:00` and must be refused, label `15:30` is the session's last regular minute and must be admitted; the unedited disjuncts (non-zero seconds, wholly pre-session labels) keep their case | `TestABarLabelledAtTheOpenIsOutsideTheRegularSession`, `TestTheLastRegularMinuteLabelIsAdmitted`, `TestAggregateClosedKRXFiveMinuteFailsClosed` (`/outside`, labels 08:55-08:59) | FAIL twice — `개장 전 1분을 담은 봉이 통과했다: err=<nil>` (the pre-open bucket was accepted) and `strategy market: outside_regular_session: 2026-07-31T15:30:00+09:00` (the last regular bucket was refused) | PASS |
| B8 | first candle fixes the currency | every accepting case | not edited | green |
| B9 | non-KRW or mixed currency | not-applicable: no case in the suite feeds a non-KRW page through this function | not edited | not-applicable |
| B10 | five decimal fields per candle | every case | not edited | green |
| B11 | exponent form or negative decimal | `TestAggregateClosedKRXFiveMinuteFailsClosed/decimal` (`"1e2"`) | not edited | green |
| B12 | the genuine first bucket of the session (labels `09:01`-`09:05`) must align on the derived open | `TestTheBucketOpensOneMinuteBeforeItsFirstLabel` | FAIL — `strategy market: incomplete_bucket: not aligned to KRX five-minute boundary` | PASS |
| B13 | contiguity loop over minutes 1..4 | `TestAggregateClosedKRXFiveMinuteFailsClosed/gap` | not edited | green |
| B14 | a minute is missing or out of step | `TestAggregateClosedKRXFiveMinuteFailsClosed/gap` | not edited | green |
| B15 | bucket not yet closed | `TestAggregateClosedKRXFiveMinuteFailsClosed/open` (`closedAt − 1ns`) | not edited | green |
| B16 | high/low/volume fold over the five minutes | `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` | not edited | green |
| B17 | a higher high replaces the running high | `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` (`101.10` from the first row survives) | not edited | green |
| B18 | a lower low replaces the running low | `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` (`99.90`) | not edited | green |
| final return | the `OpenAt` / `ClosedAt` values | `TestTheBucketOpensOneMinuteBeforeItsFirstLabel` and `TestTheLastRegularMinuteLabelIsAdmitted` both assert both instants | FAIL (both, as above) | PASS |

## Mutation evidence (the tests were made to fail on purpose)

Run with `go test -overlay`, so the working-tree file was never written to. Command shape: `go test -overlay=<mutant>.json ./internal/strategymarket -run 'TestTheBucketOpens|TestABarLabelledAtTheOpen|TestTheLastRegularMinute' -count=1`.

| Mutant | Change | Result |
|---|---|---|
| `m1_window` | session window reverted to `< 9*60 \|\| >= 15*60+30` | KILLED — `TestABarLabelledAtTheOpenIsOutsideTheRegularSession` and `TestTheLastRegularMinuteLabelIsAdmitted` both fail |
| `m2_align` | alignment anchored one minute off (`openMinute+1`) | KILLED — `TestTheBucketOpensOneMinuteBeforeItsFirstLabel` and `TestTheLastRegularMinuteLabelIsAdmitted` both fail |
| `m3_openat` | open derivation reverted to the label (`Add(0)`) | KILLED — the same two tests fail |

Restoration proof: after the overlay runs, `grep -cF "openAt := minutes[0].local.Add(-time.Minute)"` and `grep -cF "minuteOfDay < 9*60+1 || minuteOfDay > 15*60+30"` each return `1` in `internal/strategymarket/bars.go`, and `go test ./internal/strategymarket -count=1` is green (24 tests). The mutants live only in the session scratchpad; `git checkout` was never used, so no GREEN work could be erased by the revert.

## Verification

`go test ./internal/strategymarket ./internal/strategycandle ./internal/strategyengine -count=1` green; full `go test ./... -count=1` green (9,425 tests, 102 packages, exit 0); `go vet` clean; `$(go env GOROOT)/bin/gofmt -l internal/ tools/ cmd/` empty. All on 2026-08-18.
