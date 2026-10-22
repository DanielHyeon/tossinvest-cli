# a171 — 진입 접수 결과 불명의 8s 창 재전송(전략 claimed 경로 포함)

> **정의만 — 착수 전.** a165 에서 이관된 범위의 골격이다(사용자 결정 2026-10-23, a165 review.md 「1.3 재리뷰(2판)」 F3 = 1).
> design·base·FLM·코드는 없다. 실계좌 주문·운영 토글 flip 은 어떤 단계도 자동으로 하지 않는다.

## Why

`verify-execution-capability` 1.3 확정(2026-10-22)의 「주문 접수 안전 재시도」 중 (b) — 결과 불명(타임아웃·5xx·전송 후 끊김)은 조회 후 부재일 때만
같은 `clientOrderId` 로 재전송, 최초 전송부터 8s 창 안·최대 2회 — 은 a165 가 설계했으나 **생산 도달 0** 이었다. 생산 매수 진입은 전략 claimed 경로
(`execgw.Gateway.PlaceClaimedStrategy`, strategy_dispatch_cycle.go:206)뿐인데 그 경로는 EntryGate 잠금 안에서 브로커를 부르고 SUBMITTING 리스·사전 전송
거절 정산이 얽혀 있어 a165 가 제외했다(a165 design 2판 D9). 그래서 진입 재전송을 그 경로와 함께 새 change 로 설계한다.

## What Changes (예정 — design 에서 확정)

- 노출 증가 지정가 접수의 결과 불명을 같은 attempt 안에서 8s 창·최대 2회 재전송(같은 키·같은 본문), 재전송 전 미체결·종결 조회.
- 재전송 ACK 의 정착 전 목록 대조(일치 정확히 1건·같은 orderId 일 때만 CONFIRMED) — a165 1차 리뷰 A1, 사용자 결정 7.
- 전략 claimed 경로: 재전송마다 `withStrategyEntryGateAuthority` 재통과, 대기는 잠금 밖, 리스·정산과의 관계.
- 429 갈래는 a166(계좌 공유 3 req/s) 착지 뒤 별도 결정(a165 사용자 결정 5). 전송 전 실패 갈래는 제외(결정 4).

## a165 에서 이관된 미해소 발견·잔여

| id | 내용 | 출처 |
|---|---|---|
| F2 | 조회·R2 를 1페이지로 판정할 때의 완전성 — CLOSED 정렬 방향 openapi 미정, `From` 미지정 시 전 기간. `From`=거래일, `HasNext=true`·일치 없음 → 「모름」 제안. CLOSED 정렬 [미측정] | a165 review.md 재리뷰 F2 |
| F4 | `err == official.ErrServer` 동일성·ctx 전체 tracker 로는 「주문 POST 가 5xx/끊김」을 판정 못 함(토큰 교환 non-200 bare `ErrServer`, token.go:152–156; accounts GET) — 요청별 신호 필요 | 재리뷰 F4 |
| F6 | R2 강도 한계 — 목록 지연이면 첫 주문을 못 보고 CONFIRMED 가능(R-a165-3 아래) | 재리뷰 F6 |
| A2 | 응답 받은 뒤의 실패(2xx 디코드·본문 읽기·echo 불일치)를 (b) 에서 거르기 | 1차 A2 |
| F8 | `GatewayAdapter.PlaceStrategyEntry`(strategy nil 인 `Gateway.Place`) 경계 | 재리뷰 F8 |
| R-a165-1 | 401 을 받은 POST 가 실제 미접수라는 실측 없음 | a165 issues.md |
| R-a165-2 | 429 를 받은 POST 가 실제 미접수라는 것은 사용자 결정의 전제 | a165 issues.md |
| R-a165-3 | 실패 응답 뒤 같은 키 재요청의 멱등 동작 미측정(사용자 수락, A1 수리 조건) | a165 issues.md |
| R-a165-4 | HTTP/2 transport 내부 재전송 가능성 | a165 issues.md |

a165 의 design 2판(D1–D6)·FLM 번들(조회 경로 newMatcher·parseOrderFacts·matcher.matches·roundTripFor·DispatchVerified 등)은 출발 증거로 쓸 수 있으나,
착수 시 base 를 새로 고정하고 재추출한다.

## Capabilities

### Modified Capabilities

- `order-execution`: 「Retry Matrix 산출물」 예외(진입 결과 불명 재전송), 「IN_DOUBT 해소」 와의 관계. 아래 delta 는 골격 초안(ADDED)이며 design 에서 MODIFIED 로 다시 쓴다.

## Impact

- 대상(예상): `internal/execgw/`(submit·strategy_gateway·strategy_entry_gate_authority), `internal/official/`(요청별 결과 신호), 정본 retry matrix.
- 위험도: **High**. NOT READY.
