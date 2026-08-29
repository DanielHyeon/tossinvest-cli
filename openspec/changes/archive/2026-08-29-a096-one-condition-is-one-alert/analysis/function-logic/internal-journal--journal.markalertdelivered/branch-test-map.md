# Branch Test Map: `Journal.MarkAlertDelivered`

측정: `go test -covermode=set ./internal/journal/...` — RED 74.9%(base `ec29dc72`), GREEN 75.0%(2판).
a096은 이 함수를 편집하지 않았다. 줄 번호가 밀렸다(`ClaimAlertForDelivery`가 앞에 생겼다).
줄 번호는 **GREEN 기준**이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `ExecContext` 오류 `:297` | 없음 — DB 실패 주입 하니스 없음 | 미진입 | 미진입 |
| — | 정상 전이 + `requireOneRow` | 기존 journal·obs 테스트 | 진입 | 진입 |

## `requireOneRow`의 0행 경로는 obs 쪽에서 진입했다

이 함수 자체의 유일한 분기는 양쪽 모두 미진입이지만, **0행 경로는 `internal/obs`에서
진입했다** — `deliver` B5가 그 오류를 받는 자리이고, `TestNotifierIsConcurrencySafe`가
그것을 만들었다.

```text
RED   notifier.go:258.88,260.5 1 1   ← 이 함수가 오류를 돌려준 결과
GREEN notifier.go:276.88,278.5 1 0
```

즉 "이미 전달된 행에 다시 표시" 조건은 **테스트되고 있었으나 다른 패키지에서, 단언 없이**
실행되고 있었다. a096은 그 조건이 애초에 발생하지 않게 만들었고, GREEN의 미진입이 그 측정값이다.

이 함수는 그래서 **덜 호출된다.** 계약도 SQL도 바뀌지 않았다.

## 미커버 분기에 대한 판단

B1은 DB 계층 실패 주입이 필요하고 하니스가 없다.
`not-applicable`: 실패 주입 하니스 부재, a096 편집 대상 아님.
