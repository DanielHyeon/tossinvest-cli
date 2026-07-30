# Function Logic Map: `TestTheUSAdvisorySaysItsClosedResponseIsUnmeasured`

- Source: `internal/verifylive/us_market_test.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 테스트 — **본문 무변경**이다. 앞선 커밋 `f62457c`(`apply-us-measurement-fixes`)가
이 함수 **뒤**(base L226 이후)에 `observationDetail`과 기록 테스트 둘을 삽입했고,
그 hunk가 함수 끝과 인접해 evidence가 요구됐다. base(`583772c`) L214-226과 HEAD L214-226은
**바이트 동일**(함수 구간 sha256 `3cd99306b88ec6b7…` 일치, 본 세션 확인). 줄 번호도 그대로다.

이 테스트가 지키는 규율은 이 change 구간 전체의 규율과 같다: **재지 않은 것을 잰 것처럼
말하지 않는다.** KR 권고문은 이 계좌가 실제로 돌려받은 코드(`order-hours-closed`, M1)를
인용하고, US에 대해서는 그에 준하는 관측이 없으므로 KR의 것을 빌려 쓰지 않고 "미측정"이라고
말한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `SessionAdvisoryFor(MarketUS, 08:00 ET)` | 미국 장 시작 전 | `hours.go` | `t.Errorf` |
| `SessionAdvisoryFor(MarketKR, 21:31 KST)` | 한국 장 마감 후 | 동상 | `t.Errorf` |

불변식: 양방향이다 — US 문구에 KR의 측정 코드가 **없어야** 하고, KR 문구에는 그것이
**있어야** 한다(측정을 잃지 않는다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, base L216) | US 권고문이 `order-hours-closed`를 인용 | 없음 | `t.Errorf` — 겪은 적 없는 시장에 대해 사실을 주장했다 | 자체 실행 |
| B2 (if, base L219) | US 권고문에 `미측정`이 없음 | 없음 | `t.Errorf` | 자체 실행 |
| B3 (if, base L223) | KR 권고문이 측정 코드를 잃음 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SessionAdvisoryFor` ×2 | 시장별 권고문 | 순수 | ast.json calls |
| `atET` / `atKST` | 시각 픽스처 | `t.Fatalf` on parse | ast.json calls |
| `strings.Contains` | 문구 단언 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음. 브로커·계좌·파일 무접촉. **LIVE 주문 side effect 0**.

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 인접 삽입만.
- High-risk impact: **no** — 테스트 전용이고 계좌에 닿지 않는다.
  다만 재는 규율("재지 않은 것을 잰 것처럼 말하지 않는다")은 이 change 구간에서 기록
  문장을 시장에 연동한 이유와 같은 것이며, 그 연속성을 남겨 둔다.
