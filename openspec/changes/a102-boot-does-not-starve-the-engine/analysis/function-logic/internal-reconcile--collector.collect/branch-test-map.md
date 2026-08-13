# Branch Test Map: `Collector.Collect`

Source: `internal/reconcile/snapshot.go` (245-295). AST 기준 branches **8** / returns 6.

## 커버리지는 주장이 아니라 측정값이다

편집 전(`:232-282`, 130건 통과)과 편집 후(`:245-295`, 143건 통과 · 86.6%)를 같은 명령
`go test ./internal/reconcile/ -count=1 -coverprofile`으로 재고 블록 카운트를 잘라 읽었다.

| Branch | 위치 | 조건 평가 | 본문 실행 | 근거 블록 (편집 후) |
|---|---|---|---|---|
| B1 | `:246` validate 실패 | yes | **yes** | `246.37,248.3` count=1 |
| B2 | `:251` `maxPages <= 0` | yes | **yes** | `251.19,253.3` count=1 |
| B3 | `:260` `ScanOrders` 오류 | yes | **yes** | `260.16,262.3` count=1 |
| B4 | `:263` `range raws` | yes | **yes** | `263.27,265.18` count=1 |
| B5 | `:265` `parseBrokerOrder` 오류 | yes | **yes** | `265.18,267.4` count=1 |
| B6 | `:275` `holdings` 오류 | yes | **yes** | `275.16,277.3` count=1 |
| B7 | `:281` `range Currencies` | yes | **yes** | `281.40,283.17` count=1 |
| B8 | `:283` `BuyingPower` 오류 | yes | **yes** | `283.17,286.4` count=1 |

여덟 분기 전부 편집 전에도 실행됐다. **그런데 그것을 실행시킨 기존 테스트는
`errors.Is(err, ErrPartialSnapshot)` 하나만 물었다.** 원인의 정체를 묻는 판정이 한 건도
없었기 때문에 `%v` wrap이 사슬을 끊는다는 사실이 커버리지 100%인 채로 숨어 있었다 —
**커버리지는 "실행됐다"이지 "검사됐다"가 아니다.**

| 분기 (재인용) | a102가 요구하는 성질 | 지는 테스트 |
|---|---|---|
| B3 `:260` | 429가 `errors.Is(err, official.ErrRateLimited)`로 보인다 | `TestCollectPreservesTheRateLimitIdentity` (open-order list) |
| B6 `:275` | 같음 (보유 읽기) | `TestCollectPreservesTheRateLimitIdentity` (holdings) |
| B8 `:283` | 같음 (현금 읽기) | `TestCollectPreservesTheRateLimitIdentity` (buying power) |
| B3 `:260` (둘째 성질) | 429가 아닌 원인도 사슬에 남는다 | `TestCollectStillReportsAPartialSnapshot` |
| B5 `:265` | `ErrPartialSnapshot` 판정 무회귀 | `TestCollectStillReportsAPartialSnapshot` · `TestUnreadableOrderDiscardsTheSnapshot` |
| B1·B2·B4 | 기존 계약 무회귀 | `TestCollectorRefusesIncompleteWiring` · `TestSnapshotWalksTheOpenListToTheLastPage` |
| B3·B6·B8 | 부분 스냅샷 폐기 무회귀 | `TestPartialSnapshotIsDiscardedWhole` |

## 뮤테이션 정산 (통과는 실패시켜 본 뒤에만 증거다)

뮤테이션 (c)는 자리마다 따로 가했다. 네 자리 중 **셋은 죽고 하나는 살아남는다.**

| 뮤테이션 (c) 자리 | 죽은 테스트 | 살아남았나 |
|---|---|---|
| `:261` `walking the open-order list: %w` → `%v` | `TestCollectPreservesTheRateLimitIdentity/open-order_list` · `TestCollectStillReportsAPartialSnapshot/an_ordinary_failure…` · `TestRateLimitedCollectDoesNotEndRecovery` · `TestRateLimitDoesNotConsumeAStabilisationAttempt` · `TestRateLimitDefaultsMatchTheSurveyDiscipline` · `TestRateLimitWaitBudgetExhaustionFailsClosed` | 아니오 |
| `:276` `sweeping the holdings: %w` → `%v` | `TestCollectPreservesTheRateLimitIdentity/holdings` | 아니오 |
| `:284` `reading the %s buying power: %w` → `%v` | `TestCollectPreservesTheRateLimitIdentity/buying_power` | 아니오 |
| `:266` `%w: %w` (파싱 오류) → `%v` | **없음 — 전 스위트 통과(ok, 15.9s)** | **예 ⚠** |

### ⚠ 살아남은 뮤테이션 하나를 조용히 넘기지 않는다

`:266`이 감싸는 원인은 `parseBrokerOrder`가 만든다. 그 함수는 자기 원인을 이미
`%v`로 지운다(`compare.go:640` — `"an open order could not be read: %v"`), 그리고
`:643`은 아예 문자열만 있는 오류다. **되살릴 정체가 애초에 없으므로 이 자리의 `%w`는
관측 가능한 차이를 만들지 않는다.** 그래서 그것을 재는 테스트도 쓸 수 없었다.

`compare.go`를 고치면 관측 가능해지지만 그것은 a102 §1의 편집 범위가 아니고
(design D1b는 `snapshot.go`의 네 줄을 지정한다) 새 FLM 대상을 하나 더 만든다.
**설계대로 네 곳을 다 바꾸되, 그중 하나가 미검증이라는 사실을 여기에 남긴다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 8, returns 6) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/reconcile/ -count=1 -coverprofile` exit 0 · 143건 통과 · 86.6%
- 소비자 전수: `rg -n '\.Collect\(' internal/ cmd/` → non-test 3건 (recovery·reconcileloop·flatten)
