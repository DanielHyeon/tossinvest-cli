# Function Logic Map: `ordersRawClient`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`, `branches: null` — 무분기)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트 헬퍼**다(HEAD L466-475). `httptest` 서버를 향한 `*Client` 하나를 만든다.
`WithAccountSeq(3)`이 요점이다 — 계좌 seq를 고정해 `/api/v1/accounts` 지연 해석 왕복을
없앤다. 그래서 `TestBothOrderReadsSendTheGroupTheyWereGiven`의 "요청이 정확히 2건"과
`TestTheRawReadsRefuseARequestWithNoStatusGroup`의 "요청이 0건" 단언이 **계좌 해석 요청에
오염되지 않고** 성립한다.

한 헬퍼가 클라이언트를 하나만 만든다는 사실이 additive 주장의 실측 조건이기도 하다:
두 읽기가 같은 클라이언트를 공유하므로 계좌 해석·토큰·rate 예산이 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `srv` | 호출 테스트가 세운 `httptest.Server` | 호출 테스트 | — |
| 토큰 캐시 | `t.TempDir()/t.json` | 이 헬퍼 | 테스트 종료 시 정리 |
| 계좌 seq | `3` 고정 | 이 헬퍼 | 지연 해석 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (무분기 happy path) | — | 없음 | `*Client` 하나 | `TestTheRawReadsRefuseARequestWithNoStatusGroup`, `TestBothOrderReadsSendTheGroupTheyWereGiven`, `TestTheAdaptedOrderReadCannotTellAnAbsentPriceFromAZeroOne`, `TestTheRawOrderRead*` 3건 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` / `t.TempDir` | 실패 위치·임시 캐시 | — | ast.json calls |
| `New` + `WithBaseURL`/`WithHTTPClient`/`WithAccountSeq` | 테스트 클라이언트 | — | ast.json calls |
| `filepath.Join` | 캐시 경로 | 순수 | ast.json calls |

## State mutations and fallbacks

- `t.TempDir()` 안에 토큰 캐시 파일 하나. 실계좌·실브로커 무접촉.

## Safety conclusion

- Safe edit boundary: 신규 테스트 헬퍼 가산. 무분기.
- High-risk impact: **no** — 테스트 전용. 다만 `WithAccountSeq(3)`이 없으면 위 테스트들의
  요청 수 단언이 계좌 해석 요청에 오염되어 조용히 무의미해진다는 점은 기록해 둔다.
