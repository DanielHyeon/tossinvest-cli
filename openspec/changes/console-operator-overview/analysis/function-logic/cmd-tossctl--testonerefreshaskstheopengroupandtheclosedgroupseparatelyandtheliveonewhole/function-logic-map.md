# Function Logic Map: `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L563–674, 분기 18개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

이 화면의 **wire**를 고정한다. 적대적 리뷰가 첫 구현에서 연 발견이다(design D2, 개정).

`status`는 GET /api/v1/orders에서 `required: true`이고 두 값이 서로 다른 모양의 답을 고른다 — OPEN은 "모든 대기 중 주문을 전량 반환, limit·cursor 무시", CLOSED는 "limit(기본 20, 최대 100)·cursor·from/to 적용". 첫 구현은 둘 다 보내지 않고 `limit=100`만 보냈다. 그러면 ① 요청 거절 → 미체결 건수 영구 미측정, 또는 ② 계좌 전 이력 위의 1페이지 → 100행 밖의 **살아있는 주문이 표에서도 건수에서도 사라지고 "0건 이상"으로 렌더**된다. 잔여물을 놓치는 것이 이 화면이 막으려는 실패 그 자체다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버의 `status` 라우팅 | OPEN/CLOSED/누락 | 이 테스트 | 누락 케이스는 400 — 실물이 하는 것 중 친절한 쪽 |
| `seen []wire` | 3건이어야 함 | 핸들러 기록 | 3이 아니면 예산 위반 |
| `consoleOrdersPageLimit` | 100 | console.go | CLOSED에만 실려야 함 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 토큰 엔드포인트 | 기록하지 않음 | 토큰 응답 | 이 테스트 |
| B2 | 요청 라우팅 switch | — | — | 동일 |
| B3 | `/api/v1/orders` + `status=OPEN` | 미체결 1건, `hasNext:false` | — | 동일 |
| B4 | `/api/v1/orders` + `status=CLOSED` | 종결 1건, `hasNext:true` | — | 동일 |
| B5 | `/api/v1/orders` (status 없음) | — | 400 — 이 테스트가 불가능하게 만들려는 모양 | 동일 |
| B6 | `/api/v1/conditional-orders` | 빈 목록 | — | 동일 |
| B7 | 그 외 | — | 404 | 동일 |
| B8 | `Orders` 에러 | — | `t.Fatalf` | 동일 |
| B9 | 콜 수 != 3 | — | `t.Fatalf` — 예산은 미체결1+종결1+조건1 | 동일 |
| B10 | 라이브 콜이 `status=OPEN`이 아님 | — | `t.Errorf` | 동일 |
| B11 | 라이브 콜이 limit/cursor를 실었음 | — | `t.Errorf` — 자를 수 없는 콜에 자른다는 표시 | 동일 |
| B12 | 두 번째 콜이 CLOSED가 아님 | — | `t.Errorf` | 동일 |
| B13 | CLOSED의 limit != 100 | — | `t.Errorf` | 동일 |
| B14 | 세 번째 콜이 조건주문 엔드포인트가 아님 | — | `t.Errorf` | 동일 |
| B15 | 미체결 목록이 그 1건이 아님 | — | `t.Errorf` | 동일 |
| B16 | `OpenTruncated`가 참 | — | `t.Error` — 미체결 건수는 하한이 아니라 수 | 동일 |
| B17 | 종결 목록이 그 1건이 아님 | — | `t.Errorf` | 동일 |
| B18 | `ClosedTruncated`가 거짓 | — | `t.Error` — hasNext 유실은 확정 건수로 렌더된다 | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleOrdersSeam(newConsoleBroker(&rootOptions{})).Orders(ctx)` | 실제 seam을 그대로 실행 | 3콜 | console.go L469 |
| `official.New(..., WithAccountSeq(7))` | 계좌 해석 없이 클라이언트 고정 | 테스트가 wire만 보게 한다 | L601 |

## State mutations and fallbacks

- 테스트 — httptest 서버, 실계좌 접촉 없음.

## Safety conclusion

- Safe edit boundary: 세 콜의 파라미터 기대값. `status`를 지우는 편집은 잔여물을 숨긴다.
- High-risk impact: yes (주문 경로) — 살아있는 주문이 화면에서 사라질 수 있는지를 결정하는 검사다. 실계좌 부작용은 없다.
