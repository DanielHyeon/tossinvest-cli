# a112 7.3.1 SHADOW — Pre-Edit FLM (편집 전, HEAD 9e5f3ccf)

설계: `analysis/shadow-2026-10/design-brief.md` v3.3(freeze 종결 2026-10-05). 좌표 · 분기 · return · 호출은 `go run ./tools/logic-map` 산출물만(`harness/render_pre_edit.py`).

## High-risk Pre-Edit 선언

이 로트는 High-risk 경로(조정 · 제안 권한 · 조립 · 시장 주기 · supervisor 배선)의 기존 함수 본문을 바꾼다. 허용 범위는 브리프 v3.3 의 운반 · 보관 · 배선뿐이다:
주문 · 손절 · 익절 · 사이징 · Guardian · 원장 쓰기 · dispatch 원천 · 활성화 판정은 바꾸지 않는다. 생산 shadow 핀 0 → 생산 동작 변화는 「파도마다 미선언 판정(파일 I/O 0)」 하나.

## 편집 대상 번들(14) — 브리프 §4 목록 13 + FLM 이 새로 잡은 1

| 번들 | 분기 | return | 비고 |
|---|---:|---:|---|
| `internal-app-engine--coordinatemarketproposals` | 8 | 2 | §4 수집 helper · 부재 값 |
| `internal-app-engine--strategyproposalauthorityloader.collectmarket` | 15 | 15 | 아래 「운반 범위 AST 대조」 |
| `internal-app-engine--strategyproposalauthorityloader.collect` | 6 | 2 | shadow 짝 반환 |
| `internal-app-engine--context.newpairedstrategyentryproductionassembly` | 8 | 6 | 조립 별개 필드 |
| `internal-app-engine--context.runproductionstrategymarketcycle` | 4 | 5 | evaluate 인자 index 5 |
| `internal-app-engine--context.newrefreshingpairedstrategyentrysupervisor` | 4 | 5 | cycle 클로저 → helper |
| `internal-app-engine--context.productionstrategyworker` | 1 | 3 | cycle 클로저 → helper |
| `internal-app-engine--newstrategylaneruntime` | 1 | 2 | **목록 밖 — FLM 이 새로 잡음**(v3.3 「epoch 맵은 생성자에서」 가 기존 생성자 본문을 편집) |
| `internal-app-engine--strategylaneruntime.evaluate` | 8 | 4 | 인자 전달만 |
| `internal-app-engine--strategylaneruntime.record` | 4 | 1 | 칸 보관 대입 |
| `internal-app-engine--strategylaneruntime.runlane` | 2 | 2 | 편집 0 예상(목록 항목 — 기준만) |
| `internal-app-engine--strategylaneruntime.projection` | 2 | 2 | §5.1 판정 |
| `internal-app-engine--strategylaneprojection` | 5 | 2 | SHADOW 투영 |
| `internal-strategyprojection--validatelane` | 15 | 16 | 교차 규칙 |

## 운반 범위 AST 대조(`collectMarket` — 브리프 §4 의 「조정 앞 일곱 · 조정 뒤 다섯 + 충돌 + 성공」)

`ast.json` returns 15 = `fail` 클로저 본문 1(:312) + 조정 앞 7(:319 ROUTE_NOT_READY · :330 FX · :333 설정(INTERNAL) · :338 열쇠(INVALID) · :350 중복(INTERNAL) · :362 적재(INVALID) · :371 고장) +
조정 뒤 6(:407 B10 범위 지움 = FAMILY_GATE_CLOSED · :415 B11 계보 충돌 · :424 B12 overflow · :433 B13 중재 거절 · :442 B14 미해결 · :449 B15 NO_ACCEPTED_SCOPE) + 성공 :453.
브리프의 수와 일치. B10(:402) 이 B11(:409) 보다 **앞** 이므로 충돌 ∧ 범위 지움이면 FAMILY_GATE_CLOSED 로 닫히지만 그때 `coordinateMarketProposals` 가 돌려준 묶음은 이미 부재 값이라 그대로
싣는다(보이스 3 재검 4 판독과 일치 — 「조정 결과 값을 그대로 싣는다」 가 규칙).

## 무편집 기준(`no-edit/`, 15) — 착지 전 본문 digest 대조

`refreshPairedStrategyEntryProductionAssembly` · `strategyLaneInputs`(§4 무편집 사유) · `strategyProposalAuthorityPair.forMarket`(§5 shape 핀의 형제 선례) · 활성화 적재기와 §3 도우미
(`LoadProductionFamilyActivation` · `decode…` · `validate…` · `productionRouteDescriptors` · `…Time` · `…Identity` · `…DigestValid` · `…Digest` · `productionRouteOwnerUID` ·
`readProductionRouteFile`) · `invokeStrategyCycle` · `invokeBoundedStrategyCycle`(§5 defer · 마감 판정이 기대는 회복 경로). `body-digests.tsv` = 각 함수 줄 범위 본문의 sha256(편집 대상 14 포함 29행) —
무편집 15 는 착지 직전 같은 계산으로 동일해야 한다.

## RED 로 넘기는 것

브리프 §2 · §3 · §4 · §5 · §5.1 · §6 · §8 · §9 · §10 핀 + v3.3 핀 (i)~(v) · central-integrity 신원 · 의도된 간극 · 나이 상한 등식 · RED 후보 R1(두 시계). Branch Test Map 은 GREEN 뒤
편집 후 ast.json 과 difflib 정렬로(위치 분기 번호 교훈).
