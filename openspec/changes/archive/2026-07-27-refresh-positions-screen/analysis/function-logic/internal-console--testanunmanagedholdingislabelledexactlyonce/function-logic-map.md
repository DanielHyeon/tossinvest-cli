# Function Logic Map: `TestAnUnmanagedHoldingIsLabelledExactlyOnce`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

기존 테스트 함수 — 이 change는 기대 문자열 1건만 교체했다(구 행별 문장 → 새 페이지 공지 문장). 함수 구조·나머지 단언은 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 페이지 렌더 결과 | HTML string | newDashboardHarness + seedJournal | t.Error로 실패 보고 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if | 라벨 등장 횟수 < 3 | 없음 | t.Errorf | 자체 실행 |
| B2 range | 필수 설명 문자열 순회(교체된 공지 문장 포함) | 없음 | — | 자체 실행 |
| B3 if | 문자열 부재 | 없음 | t.Errorf | 자체 실행 |
| B4 range | 금지 문자열(편입 버튼류) 순회 | 없음 | — | 자체 실행 |
| B5 if | 금지 문자열 존재 | 없음 | t.Errorf | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `h.page` 등 하네스 | 페이지 렌더 | 실패 시 t.Fatal | HEAD |

## State mutations and fallbacks

- 없음(테스트). 교체된 기대 문자열은 spec delta '같은 사유의 수동 보유 다수'와 정합.

## Safety conclusion

- Safe edit boundary: 기대 문자열 1건 교체만 — 단언 구조 무변경
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)
