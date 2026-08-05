# Function Logic Map: `overviewPage.RefreshSeconds`

- Source: `internal/console/overview.go`
- AST evidence: `ast.json` (`source_sha256` 202505dbe12574e24960746df1d450ada672b37c7c4b17fb4e961a3a374fcc92, line 1164)
- Risk scan: `risk-pattern-report.md`
- 편집 전 작성 — 이 함수는 a080 task 4.3의 편집 대상이다.

## 현재 본문

```go
func (overviewPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `holdingsTTL` (패키지 상수) | `30 * time.Second` (`holdings.go:76`) | 현재 HEAD | 상수이므로 실패 없음 |
| 수신자 `overviewPage` | 값 무시 (빈 수신자) | — | 없음 |

**`positionsPage`와 근거가 다르다.** 이 화면은 `peek`만 쓰므로(`overview.go:655`,
`overview.go:1186`, `protection_liveness.go:104`) 리로드가 브로커 호출을 0회
발생시킨다. 주석이 그 사실을 이미 정확히 적고 있다.

```go
// Here every reload costs zero broker calls, so the TTL is a freshness bound
// rather than a budget one — but a second period to keep in step with the first
// is worth avoiding either way.
```

즉 여기서의 파생 근거는 예산이 아니라 **"두 번째 상수를 만들기 싫다"**이고, 그
주석은 틀리지 않았다(`positionsPage` 쪽과 달리 — issues.md I1). a080은 그
근거를 정면으로 뒤집는다: 두 번째 상수를 만들 이유가 이제 생겼다. 화면 갱신
주기는 엔진 관측 주기를 따라야 하고, 그것은 브로커 캐시 TTL과 다른 값이다.

**shadowing 해저드**: `positionsPage`와 동일 (`chrome.go:47-55`). `Refresh`는
이미 메서드에서 embedded chrome의 필드로 옮겨졌고(주석 1160-1163), a080은 그
방향을 되돌리지 않는다.

## Branches and early returns

AST `branches: null`. 분기·early return이 없다. 단일 return 1건(line 1164 col 44).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 없음 | 없음 | `int(holdingsTTL / time.Second)` = 30 (편집 전) → `lineRefreshSeconds()` = 5 | 3.1, 3.2 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `int(...)` 변환 | `time.Duration` → 초 정수 | 없음 (순수 변환) | ast.json `calls[0]` line 1164 col 51 |

live config binding 없음.

## State mutations and fallbacks

- mutation 없음 (`assignments: null`, `go_statements: null`, `defers: null`).
- fallback 없음. 순수 함수다.

## a080이 바꾸는 것

> **개정 2026-08-05 (review.md F5).** 앞 판은 철회된 fragment+스크립트 설계의
> "폴백 상수"를 서술했다. 아래는 정석 수정 뒤의 편집이다.

`positionsPage`와 동일하게 **출처가 바뀌고 값도 바뀐다** — 피연산자가
`holdingsTTL`(브로커 rate budget 상수)에서 `lineRefreshInterval`(엔진 관측 주기)로
교체되어 반환값이 30에서 5가 된다. 주석은 **정정**이다: 두 상수를 하나로 둔 것을
정당화하던 문장이 캐시 코드와 모순이었다 (issues.md I1).

`/dashboard`는 `peek` 계약을 유지해야 한다(design D2). 이 함수는 그 계약과
무관하지만, 같은 change가 두 화면을 건드리므로 여기 명시한다: 갱신 주기가
같아진다고 브로커 관계까지 같아지는 것이 아니다.

그리고 **브로커에 0원인 것이 엔진에도 0원은 아니다.** 이 화면의 렌더는
`decoratePositionRows`를 거치고 그것이 엔진 프로세스를 읽는다. 주기를 6배로
올리는 이 편집이 성립하려면 그 읽기가 주기와 분리되어 있어야 하고, 그것을
만드는 것은 a080이 아니라 **선행 change a081**이다 (review.md F1).

## Safety conclusion

- Safe edit boundary: 반환 표현식의 피연산자와 위 주석. 시그니처·수신자·분기 구조는 불변.
- High-risk impact: **no.** 표시 전용 순수 함수이며 브로커·원장·주문 경로에 닿지 않는다.
- 회귀 방지: `overview_test.go`가 이미 이 함수의 값을 상수에 결속한다. a080은 그
  테스트의 기대값 출처를 `lineRefreshInterval`로 옮기고, 새 상수가 `holdingsTTL`과
  같아지면 실패하는 절을 더한다 (task 3.1·5.1).
