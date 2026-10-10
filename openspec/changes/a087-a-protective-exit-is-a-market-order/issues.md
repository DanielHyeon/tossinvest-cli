# a087 issues

## I-P1 — Phase 1(D2a) 의 대상 모집단이 생산에서 0 이다 (blocking · 2026-09-30 · Teammate → Manager)

**분류**: WORKFLOW 예외 경로 ① blocking(스펙 전제와 코드 사실의 모순). production 편집 전 중단.

### 주장

design D2 는 "관측가도 기준선도 없으면 청산을 거부한다. 시세를 못 읽은 순간 손절이 막힌다"고 쓰고, D2a 는 그 거부
(`sellIntent` B2, `exitloop.go:1576-1579`) 앞에 KR 하한가 단을 둔다. **현재 HEAD 에서 B2 는 — 그 앞의 B1(기준선 폴백)부터 —
생산 경로로 도달할 수 없다.** 따라서 D2a 를 구현하면 생산 동작 변화는 0 이고, 새 단과 `PriceLimits` 배선은 도달 불가 코드가 된다.

### 증거 (tasks P1.2 ① 모집단 열거)

1. **호출 사슬 (CodeGraph 1.6.0)** — `sellIntent` 호출자 = `submit` 1(`:1382`), `submit` 비시험 호출자 = `record` 1(`:1301`),
   `record` 호출자 = `judgeRatchet`·`judgeLadder`. 메서드 값 참조 0(`grep '\.sellIntent\b|\.submit\b'`, 비시험).
2. **값의 출처** — `submit` 은 `observed` 를 재대입 없이 넘기고, `record` 는 `judgement.ObservedPrice = snapshot.ObservedPrice`
   (`:1190`). 스냅숏의 값은 `EvaluateRatchetSnapshot`(`snapshot.go:227`) · `EvaluateLadderSnapshot`(`:155`)이 채운다.
3. **AST (tools/logic-map, `analysis/p1-population/*.ast.json`)** — 두 스냅숏 함수는 `EvaluateRatchet`(`:189`) ·
   `EvaluateLadder`(`:120`) 의 오류를 그대로 반환(`:191` · `:122`)한 뒤에만 스냅숏을 만든다. 두 평가기는 성공 반환 전에
   `positive("observed price", …)`(`ratchet.go:347` · `ladder.go:333`)를 무조건 지나며, 그 앞의 반환은 전부 오류 반환이다
   (ratchet 341·345 / ladder 309·314·319·322·326). `positive` 는 파싱 실패와 `Sign() <= 0` 을 거부(`ratchet.go:592-601`).
4. **커버리지 실측** — `go test ./internal/app/engine/ -count=1 -coverprofile` (157 s, 70.7%): B1 본문 `1573.17-1575.3` = 0,
   B2 본문 `1576.17-1579.3` = 0 (`analysis/p1-population/sellintent-coverage.txt`). 거부문을 단언하는 시험은 저장소에 0.

### "가격이 없어 손절이 막히는" 실제 자리

가격이 없으면 거부가 `sellIntent` 에서 일어나는 것이 아니라 **판정 자체가 일어나지 않는다** — `observe` 가 `Last <= 0` 을 거르고
(`:765`), `ObserveOnce` 가 응답에서 빠진 종목·만료된 시세를 무음 `continue` 한다. 이것은 a090(미관측 계수) 의 범위이고, 하한가 단은
그 자리를 돕지 못한다(이탈 여부를 판정할 가격이 없으므로 보호 제안 자체가 없다).

### Manager 결정 요청 (선택지)

1. **Phase 1 을 not-applicable 로 닫는다** — 모집단 0 을 근거로 P1.2~P1.4 를 `not-applicable`, D2a 를 "측정 결과 불요" 로 정정.
   B1·B2 는 오늘처럼 방어 코드로 남는다(도달 불가지만 fail-closed). 가장 작다(YAGNI).
2. **방어 심층으로 구현한다** — 미래 호출자가 빈 관측가를 넘길 때만 효력. 비용: 도달 불가 분기에 High-risk 코드 + 새 GET 배선,
   변이 시험은 직접 호출로만 가능(생산 경로 지도자 없음 — [[tests-that-always-drive-leave-production-untested]]).
3. **Phase 1 을 재조준한다** — 실제 가격 부재 경로(판정 불가)로. 다만 그것은 a090 과 표면이 겹친다.

Teammate 권고: 1. B2 의 도달 불가를 시험으로 못 박는 것(평가기가 빈/0 관측가를 거부 — `ratchet_test.go:383` 은 `-1` 만,
ladder 쪽 빈 값 단언은 없음)은 선택 사항으로 함께 제안한다.

### 처분 (2026-09-30, Manager)

**선택지 1 채택** — Phase 1 불구현 종결, D2a 전제 정정. Manager 가 모집단 열거를 독립 대조해 확인했다(positive 가드 두 자리,
snapshot 평가-성공-후-생성 두 자리, `:765` 무음 skip, `sellIntent` 생산 호출 1곳). 근거: 도달 불가 High-risk 코드 금지 ·
선택지 2 는 미검증 가상 경로에 주문을 내주므로 보수 방향이 아님(B2 거부가 이미 fail-closed) · 선택지 3 은 a090 표면(중복 금지).
대체 산출물: 평가기 핀 시험 `internal/exitpolicy/a087_observed_price_pin_test.go`(`7bd8f197`, 변이 4/4 CAUGHT).
a087 은 Phase 2 사람 게이트(§0.7) 대기로 전환.

---

## 2차 proposal-freeze 재리뷰 처분에서 남긴 미해결 (2026-10-10)

`review.md` 「2차 proposal-freeze 재리뷰」 처분 중 문서 수정·구현 task 어느 쪽으로도 닫히지 않은 것. 처분 대조표는 `review.md` 끝.

### I-R1 — 셋째 관문 개방과 불변식 3(토글 OFF = upstream 동작) (미해결 · 3차 Eng 재리뷰 판정 대상)

사용자 결정(2026-10-10)으로 `internal/trading` `placeIntentSupported` 를 sell+market·가격 없음에 연다. 이 술어는 엔진 토글과
무관하게 사람 CLI(`tossctl order place`)·ops 쓰기·`PreviewPlace` 를 지배하므로, **엔진 토글이 OFF 여도** 사람 경로의 동작이
upstream 과 달라진다(시장가 매도 가능). 사용자는 이 동반 변경을 알고 결정했고 방향은 청산 즉시성 강화(불변식 6 보수 방향)다.
남은 질문: 불변식 3 이 「토글로 가려지지 않는 upstream 동작 변경」을 허용하는가, 아니면 ③의 개방을 엔진 경로 한정(토글 또는
호출자 구분)으로 좁혀야 하는가. 이 문서는 판정하지 않는다 — 0.7 Eng 재리뷰가 판정하고 1.5.0 Pre-Edit 선언이 그 판정을 인용한다.

### I-R2 — `order-type-not-allowed` 422 분류 (미해결 · 무소유)

openapi `POST /api/v1/orders` 422 예시에 `order-type-not-allowed`(「현재 사용할 수 없는 호가 유형」)가 있다. 저장소 Go 코드
참조 0, `refusal_code.go` 목록에도 없다(재리뷰 P1-4). a087 이후 세션 밖 MARKET 보호 청산이 이 코드를 받으면 분류되지 않은
거부가 된다. 분류 필요 여부는 §5.2 결과 뒤 판정. 400/422 분류 전반(1차 리뷰 I2)과 함께 무소유.

### I-R3 — MARKET vs 하한가 지정가의 bps 재비교 (미해결 · 사람 결정)

2차 리뷰 「3차 교정」 4 는 두 방식을 청산 슬리피지 bps 데이터로 결정하라고 했다. 사용자 결정(2026-10-10)이 시장가 방향을
택했으므로 이 change 는 비교를 기각 기록(design D6)으로 남기고 계측(tasks 3.8)만 둔다. 착지 뒤 bps 데이터로 재비교할지와 그
임계는 정하지 않았다.

### I-R4 — 경계 조건의 미측정 두 줄 (미해결 · 범위 밖)

design D8: (가) VI 단일가 구간에서 MARKET 접수·체결 거동, (나) 1억 이상 주문에서 MARKET 의 금액 평가 기준
(`ConfirmHighValueOrder: false` → `400 confirm-high-value-required` 는 유형 무관). 둘 다 §5 측정 항목이 아니다.

### I-R5 — a100 쪽 반영 요청 (Manager — 이 디렉터리 범위 밖)

1. a100 design D8 계약 1·2 와 a087 MARKET 의 상호작용(design D9) — a100 착수 시 재대조 필요.
2. a100 `ma-runbook.md` R2 의 a087 5.1·5.2 실행 명령 공백(재리뷰 P1-1) — 실행 수단은 `tools/a087-market-sell-probe/`.
   이번 개정은 a087 디렉터리만 바꿨으므로 a100 문서는 손대지 않았다.
