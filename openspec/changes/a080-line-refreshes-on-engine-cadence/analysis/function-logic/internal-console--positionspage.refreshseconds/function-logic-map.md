# Function Logic Map: `positionsPage.RefreshSeconds`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (`source_sha256` 16975cd71904f2871baab9a0669b549e0b1959650e161408879519f0e7b7b52a, line 57)
- Risk scan: `risk-pattern-report.md`
- 편집 전 작성 — 이 함수는 a080 task 4.3의 편집 대상이다.

## 현재 본문

```go
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `holdingsTTL` (패키지 상수) | `30 * time.Second` (`holdings.go:76`) | 현재 HEAD | 상수이므로 실패 없음 |
| 수신자 `positionsPage` | 값 무시 (빈 수신자) | — | 없음 |

**불변식**: 반환값은 템플릿 두 곳이 소비한다 — `templates.go:321`의 meta 태그와
`templates.go:383`의 상태 스트립 reload 셀. `chrome.go:77-79`는 이 둘이 **한
사실의 복사본**이어야 한다고 고정한다. 값이 바뀌면 두 곳이 함께 바뀐다.

**shadowing 해저드**: `chrome.RefreshSeconds`가 `0`을 반환하고 embedded 승격으로
이 메서드를 대체할 수 있다(`chrome.go:47-55`). 메서드를 지우면 컴파일도 렌더도
성공하면서 화면만 조용히 갱신을 멈춘다. a080은 메서드를 **지우지 않는다**.

## Branches and early returns

AST `branches: null`. 분기·early return이 없다. 단일 return 1건(line 57 col 45).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 없음 | 없음 | `int(holdingsTTL / time.Second)` = 30 (편집 전) → `lineRefreshSeconds()` = 5 | 3.1, 3.2 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `int(...)` 변환 | `time.Duration` → 초 정수 | 없음 (순수 변환) | ast.json `calls[0]` line 57 col 52 |

live config binding 없음. 런타임 설정을 읽지 않는다.

## State mutations and fallbacks

- mutation 없음 (`assignments: null`, `go_statements: null`, `defers: null`).
- fallback 없음. 순수 함수다.

## a080이 바꾸는 것

> **개정 2026-08-05 (review.md F5).** 앞 판은 철회된 fragment+스크립트 설계의
> "폴백 상수 · 값 30 유지"를 서술했다. 아래는 정석 수정 뒤의 편집이다.

**출처가 바뀌고 값도 바뀐다.** 피연산자가 `holdingsTTL`(브로커 rate budget
상수)에서 `lineRefreshInterval`(엔진 관측 주기)로 교체되어 반환값이 30에서 5가
된다. 동시에 그 위의 주석을
정정한다 — 그 주석이 서술하는 위험(TTL 아래 주기가 호출을 늘린다)은
`holdings.go:179`의 TTL gate 때문에 현재 코드에 존재하지 않는다 (issues.md I1).

편집 후 분기 수는 그대로 0이다. 이 함수가 조건부가 되면 그 자체가 결함이다 —
주기는 화면의 성질이지 요청의 성질이 아니고, 조건이 붙는 순간 스트립의 reload
셀과 meta 태그가 서로 다른 값을 말할 수 있게 된다.

이 편집이 안전하려면 주기를 6배로 올리는 것이 브로커에도(C6) 엔진에도(C7)
비용을 늘리지 않아야 한다. 앞의 것은 `holdingsCache`가 이미 보장하고, 뒤의 것은
**선행 change a081**이 만든다 (review.md F1).

## Safety conclusion

- Safe edit boundary: 반환 표현식의 피연산자와 위 주석. 시그니처·수신자·분기 구조는 불변.
- High-risk impact: **no.** 주문·손절·사이징·Guardian·원장·reconciliation·retry
  matrix·인증·체결 감지 어디에도 속하지 않는다. 표시 전용 순수 함수다.
- 회귀 방지: `overview_test.go:1074`·`orders_test.go:949`와 같은 형태의 상수
  결속 테스트가 이 함수에도 필요하다 (task 3.3).
