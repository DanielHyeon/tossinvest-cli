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
