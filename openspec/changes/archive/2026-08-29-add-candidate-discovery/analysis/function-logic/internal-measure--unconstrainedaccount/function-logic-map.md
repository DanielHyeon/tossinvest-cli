# Function Logic Map: `unconstrainedAccount`

- Source: `internal/measure/counterfactual.go`
- AST evidence: `ast.json` (revision=current, L319–339, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: 주석 열 정렬만 (gofmt, 커밋 dcc2030) — 코드 토큰 동일 (revision=current)

**본문 코드는 base와 동일하다.** 이 branch range에서 바뀐 것은 주석 정렬뿐이다 — 커밋 dcc2030(발굴 §1)의 `gofmt`가 구조체 리터럴의 줄 끝 주석 열을 다시 맞췄다. 주석과 공백을 제거하고 비교하면 토큰 열이 완전히 일치한다(검증: base `583772c4`의 L319–339 대 현재 L319–339, 주석 제거 후 동일).

함수 자체는 counterfactual 그리드의 fixture다. Guardian 체인이 **비용 모델만** 남기고 판정하도록, 게이트 12개 중 산술과 무관한 것들을 전부 무해한 값으로 세운다: kill switch off, NORMAL, latch clear, 무제한 현금·자본, 노출 0, 일손실 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market costs.Market` | KR 또는 US | 그리드 스펙 | US가 아니면 KRW/기본 헤드룸 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == costs.MarketUS` | `currency="USD"`, `headroom` 교체 | AccountState | `internal/measure`의 US 그리드 테스트 |
| (else) | KR | 기본값 유지 | AccountState | KR 그리드 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `money` | 금액+통화 구성 | 순수 | ast.json calls |

## State mutations and fallbacks

- 순수 함수. 실계좌·원장·설정 어디에도 닿지 않는다.
- 호출자는 `internal/measure`의 counterfactual 그리드뿐이다.

## Safety conclusion

- Safe edit boundary: 주석 정렬만 — 코드 무변경.
- High-risk impact: no (측정 산출물 전용 fixture — 라이브 진입 판정 경로에서 도달 불가). 다만 이 값이 언젠가 `internal/measure` 밖으로 새면 **모든 Guardian 게이트가 꺼진 계좌 상태**가 되므로, 호출자가 측정 코드뿐이라는 사실이 이 함수의 안전 근거 전부다.
