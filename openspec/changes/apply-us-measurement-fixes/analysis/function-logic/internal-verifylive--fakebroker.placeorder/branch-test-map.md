# Branch Test Map: `fakeBroker.PlaceOrder`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if f.placeAlreadyProcessing > 0 {` (internal/verifylive/fake_broker_test.go:368) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if err := f.throttled("place"); err != nil {` (internal/verifylive/fake_broker_test.go:376) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if f.placeErr != nil {` (internal/verifylive/fake_broker_test.go:379) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if strings.EqualFold(intent.Side, "sell") && f.rejectOversell && intent.Quantity > f.sellable[intent.Symbol] {` (internal/verifylive/fake_broker_test.go:382) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if intent.ClientOrderID != "" {` (internal/verifylive/fake_broker_test.go:388) | 이 함수 자체가 테스트다 | yes | yes |
| B6 | `if seen {` (internal/verifylive/fake_broker_test.go:392) | 이 함수 자체가 테스트다 | yes | yes |
| B7 | `switch {` (internal/verifylive/fake_broker_test.go:393) | 이 함수 자체가 테스트다 | yes | yes |
| B8 | `case prior.body == body && f.honourIdempotency:` (internal/verifylive/fake_broker_test.go:394) | 이 함수 자체가 테스트다 | yes | yes |
| B9 | `case prior.body != body && f.conflictOnDifferentBody:` (internal/verifylive/fake_broker_test.go:396) | 이 함수 자체가 테스트다 | yes | yes |
| B10 | `if intent.ClientOrderID != "" {` (internal/verifylive/fake_broker_test.go:408) | 이 함수 자체가 테스트다 | yes | yes |
