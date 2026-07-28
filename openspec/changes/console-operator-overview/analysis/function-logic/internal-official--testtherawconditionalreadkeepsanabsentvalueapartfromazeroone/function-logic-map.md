# Function Logic Map: `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne`

- Source: `internal/official/conditional_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L94-148). `ConditionalOrdersRaw`가 "브로커가 보내지 않은 값"과
"0"을 구분해 옮기는지, 그리고 페이지 경계를 버리지 않는지를 잰다.

픽스처는 MARKET형 SINGLE 하나다: `orderPrice`가 null(지정가가 없으므로),
`targetProfitRate`가 null(이 다리가 STOP이므로). 기존 읽기의 `parseDecimal`은 둘 다
"트리거 0"으로 렌더한다 — 이 테스트가 그 구분이 살아 있음을 고정한다.

이 제품에서 조건주문이 코너 케이스가 아니라는 것이 근거다: M18이 조건주문이 등록한
프로세스보다 오래 사는 것을 측정했고, verifylive의 정리(cleanup)는 잔여 조건주문을
잔여 일반 주문과 **같은 것** — 노출 상한을 채워 다음 요청을 막는 것 — 으로 센다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | `/oauth2/token`, `/api/v1/conditional-orders` | 이 테스트 | 그 밖은 `t.Errorf` |
| 클라이언트 | `WithAccountSeq(1)` | `New` + Option | — |
| 호출 | `ConditionalOrdersRaw(ctx, "OPEN", "", "", 0)` | 이 테스트 | 그룹 미지정이면 거부(별도 테스트) |
| 기대 | `orderPrice` → `""`, `quantity` → `"10"`, `triggerPrice` → `"70000"`, `market` → `"KR"`, `status` → `"WATCHING"` | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, L96) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B2 (case, L97) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B3 (case, L99) | `/api/v1/conditional-orders` | null 포함 페이지 응답 | — | 자체 실행 |
| B4 (case, L110) | default | 없음 | `t.Errorf("unexpected path")` | 자체 실행 |
| B5 (if, L125) | 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 (if, L128) | 건수 ≠ 1 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 (range, L132) | 필드 기대표 순회(6건) | 없음 | — | 자체 실행 |
| B8 (if, L140) | 필드 값 불일치 | 없음 | `t.Errorf` — 부재가 숫자로 도착하면 화면에서 측정된 0과 구분할 수 없다 | 자체 실행 |
| B9 (if, L144) | `HasNext`/`NextCursor` 유실 | 없음 | `t.Errorf` — 잘린 조건주문 페이지도 같은 "자신 있게 짧은 건수"다 | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역 | `defer srv.Close()` | ast.json calls/defers |
| `New` + Option 3종 | 테스트 클라이언트 | — | ast.json calls |
| `c.ConditionalOrdersRaw` | 측정 대상 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉). 재는 대상은 High-risk이며,
  이 테스트가 약해지면 잔여 조건주문의 트리거가가 화면에서 0으로 보이게 된다.
