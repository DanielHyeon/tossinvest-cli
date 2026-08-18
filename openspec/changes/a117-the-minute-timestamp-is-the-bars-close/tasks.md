## 1. Evidence

- [x] 1.1 Produce the Function Logic Map, Branch Test Map, AST and risk-pattern bundle for `aggregateClosedKRXFiveMinute` before editing it, and record its production reach from CodeGraph.
- [x] 1.2 Record the three surviving copies of the "timestamp is the bar's open" claim and the ownership decision for each.

## 2. RED Tests

- [x] 2.1 Add a failing test proving the genuine first regular bucket (labels `09:01`–`09:05`) is refused today and must yield `OpenAt 09:00` / `ClosedAt 09:05`.
- [x] 2.2 Add a failing test proving a bar labelled `09:00` covers the pre-open minute and must be refused with `RefusalOutsideRegularSession`.
- [x] 2.3 Add a failing test proving the session's last minute (label `15:30`) must be admitted.

## 3. GREEN Implementation

- [x] 3.1 Derive the bucket open from the first label, anchor grid alignment on it, and set `closedAt` from it.
- [x] 3.2 Move the session admission window into label space (`09:01`–`15:30`, inclusive) without changing any refusal kind.
- [x] 3.3 Correct the `candle_reads.go` schema comment and state why the broker document disagrees.
- [x] 3.4 Shift the three affected fixtures by one minute so each names the same bucket as before.

## 4. Verification

- [x] 4.1 Run the affected package suites, the full `go test ./...`, `go vet` and `gofmt -l`.
- [x] 4.2 Kill three mutants with `go test -overlay` (reverted window, reverted alignment anchor, reverted open derivation) and prove the source was never mutated in place.
- [x] 4.3 Run strict OpenSpec validation, the PM tracker check, `check_analysis.py`, `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a117-the-minute-timestamp-is-the-bars-close`.
