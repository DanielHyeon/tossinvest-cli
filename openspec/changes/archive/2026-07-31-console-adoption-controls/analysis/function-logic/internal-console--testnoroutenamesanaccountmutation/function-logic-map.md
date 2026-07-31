# Function Logic Map: `TestNoRouteNamesAnAccountMutation`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

가드 개정: actVerbs에 config-쓰기 어휘(save·include·enable·config) 추가(P2-7) + 전사 문장 갱신. 계좌 동사 목록 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B2 | range | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B3 | range | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B4 | if | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B5 | if | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B6 | range | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B7 | if | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B8 | range | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |
| B9 | if | 없음 | — | 자체 실행 green(신규 라우트는 허용 목록·기존 라우트는 어휘 비저촉 확인) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(테스트)

## Safety conclusion

- Safe edit boundary: 어휘·전사 확장만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
