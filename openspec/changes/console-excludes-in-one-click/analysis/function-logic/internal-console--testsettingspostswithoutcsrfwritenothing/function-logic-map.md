# Function Logic Map: `TestSettingsPostsWithoutCSRFWriteNothing`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 대상 경로 목록 | 세 config 편집 라우트 | `consoleStateChanging`와 같은 집합이어야 한다 | 빠뜨린 경로는 무게이트로 남아도 이 테스트가 침묵한다 |
| `fakeSettings.saves` | 0이어야 한다 | 세는 fake | 1 이상이면 CSRF 없는 요청이 seam에 닿았다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` 세 경로 | CSRF 없는 POST 발사 | — | 이 테스트 |
| B2 | 응답이 200이고 `/refused`가 아님 | 없음 — 응답 형태는 판정하지 않는다 | — | 이 테스트 |
| B3 | `saves != 0` | 없음 | 실패 | 이 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `harness.post` | CSRF 토큰 없는 POST | — | CodeGraph + AST |
| `fakeSettings.saved` | 저장 횟수 | — | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 경로 목록에 `/settings/exclude`를 **더한 것**뿐이다.
- 판정은 응답 코드가 아니라 **seam이 닿았는가**다 — 콘솔이 어떤 거부 페이지를 그리든 config가 바뀌지 않았다는 것이 주장이다.

## Safety conclusion

- Safe edit boundary: 리터럴 목록 1개 원소.
- High-risk impact: no — 테스트다. 다만 이 목록이 실제 라우트보다 짧으면 게이트 부재가 조용히 통과한다.
