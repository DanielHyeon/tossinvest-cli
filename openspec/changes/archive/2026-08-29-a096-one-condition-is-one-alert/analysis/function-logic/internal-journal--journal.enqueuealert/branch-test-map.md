# Branch Test Map: `Journal.EnqueueAlert`

측정: `go test -covermode=set ./internal/journal/...` — RED 74.9%(base `ec29dc72`), GREEN 75.0%(2판).
a096 이후 이 함수는 **분기가 없다**(위임 4줄, `outbox.go:115-118`). 분기 없는 함수의
happy-path 한 행을 적으며, GREEN 판정은 그 본체 블록의 count로 한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 유일 경로: 위임하고 id를 돌려준다 | `TestEnqueueAlertKeepsItsContract` (a096 신규) + `TestEnqueueAlertIsIdempotentOnTheEventKey` 등 기존 7개 (전부 무수정) | 진입 | 진입 (`outbox.go:115-118` 블록 count 1) |

## RED의 9분기는 어디로 갔나

base `ec29dc72`에서 이 함수는 B1@113 … B9@147의 9분기 본체였다. a096은 그 본체를
`ClaimAlertForDelivery`로 옮겼고, 분기 열거와 커버리지 판정은 그 함수의 BTM으로 이동했다.
같은 SQL, 같은 트랜잭션 경계, 같은 오류 문구다.

## 이 함수를 위임으로 남긴 것이 증거를 지켰다

서명이 그대로이므로 기존 호출자가 무수정이다. 이것은 편의가 아니라 증거의 문제다:
`internal/journal/outbox_test.go`의 7개 테스트를 arity 때문에 기계적으로 고치면 그 7개가
전부 "수정된 기존 함수"가 되어 각각 Function Logic Map을 요구한다. 그렇게 생산된 28개
파일은 각 테스트의 안전성이 아니라 **인자 개수가 바뀌었다는 사실**을 28번 적은 것이다.

실제로 한 번 그 길로 갔다. 서명을 `(int64, string, error)`로 넓히고 호출부를 기계적으로
고쳤더니 `check_analysis.py`가 증거 누락 8건을 보고했다 —
`internal/execgw/replay.go:Gateway.parkAlert`와 테스트 7개. 그 보고가 설계를 좁히게 했고,
좁힌 결과가 더 나은 API이기도 하다: "기록해라"와 "보내도 되나"는 다른 질문이다.
