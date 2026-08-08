# a091 tasks

> **High-risk.** 손절 경로의 함수를 편집한다. 다만 **제출 수량 계산은 건드리지 않는다** —
> 바꾸는 것은 보고(등급·종류·문구)뿐이다. proposal-freeze 리뷰(적대적 Eng 필수)가
> 구현 착수 전에 필요하다.

## 0. 게이트 선행

- [ ] 0.1 `capture_change_base.py --change a091-a-stop-that-sold-nothing-is-critical`
- [ ] 0.2 `openspec validate a091-a-stop-that-sold-nothing-is-critical --strict --no-interactive`
- [ ] 0.3 **proposal-freeze 리뷰** (적대적 Eng 필수) → `review.md`
- [ ] 0.4 `check_analysis.py --change a091-…` — FLM 산출물 완결 확인

## 1. 산출물 (완료)

- [x] 1.1 **Function Logic Map** — `ExitObserver.applyFloor` (branches 6, returns 7)
- [x] 1.2 **Function Logic Map** — `SeverityOf` (branches 1, returns 2)
- [x] 1.3 **Branch Test Map** — 위 두 함수. 미테스트 분기 B4·B6 식별
- [ ] 1.4 새 이벤트 종류의 **소비자 조사** — 콘솔 필터·로그 대시보드·`CriticalEvents()`
      호출자. 조사 결과를 `issues.md`에 기록

## 2. 이벤트 종류 신설 (D1)

- [ ] 2.0 **Pre-Edit 선언** — `internal/obs/event.go`
- [ ] 2.1 **RED** — 새 종류가 `SeverityOf`에서 critical
- [ ] 2.2 **RED** — 기존 18종의 등급 **무변화**
- [ ] 2.3 **RED** — 미등록 종류는 여전히 normal (기본값 보존)
- [ ] 2.4 **GREEN** — 종류 추가 + `criticalEvents` 등록.
      `EventExitProposalCapped`는 **부분 캡 전용**으로 좁아지고 등급 무변화

## 3. 0주 두 경로를 같게 보고 (D3)

- [ ] 3.0 **Pre-Edit 선언** — `ExitObserver.applyFloor`, `ExitObserver.submit`
- [ ] 3.1 **RED** — 보호 + `floor.Quantity == 0` (`:1446`) → 새 종류, critical, outbox 행
- [ ] 3.2 **RED** — 보호 + 하한 계산 실패 (B2 `:1408`) → **같은 종류·등급**, 원인이 detail에
- [ ] 3.3 **RED** — 보호 + **부분** 캡 → `EventExitProposalCapped` 유지, 등급·문구 무변화
- [ ] 3.4 **RED** — 익절 + 0주 → 종전 등급 무변화
- [ ] 3.5 **RED (§0.3·§0.9 회귀)** — 위 전부에서 `applyFloor`의 반환값
      `(수량, capped, err)`이 **무변화**. 이 테스트가 이 change의 안전 경계다
- [ ] 3.6 **GREEN** — `submit`이 `isProtective(proposal)`을 `applyFloor`에 전달.
      판정기 `exitpolicy`는 건드리지 않는다
- [ ] 3.7 B2의 `logErr` 유지 확인 — 오류 객체를 담는 유일한 자리

## 4. 문구 (D4)

- [ ] 4.1 **RED** — 0주일 때 제목·본문이 "일부만 나갔다"라고 말하지 않는다
- [ ] 4.2 **RED** — 부분 캡의 문구는 **무변화**
- [ ] 4.3 **GREEN**

## 5. 실측 재생

- [ ] 5.1 2026-08-02의 13회 시퀀스를 fixture로 재생 — outbox 행이 생기는지,
      반복이 어떻게 접히는지(`event_key` 중복 제거) 확인
- [ ] 5.2 재생 결과를 `issues.md`에 기록. **a089(outbox 재발 장부)와의 상호작용**을 명시 —
      13회가 한 행으로 접히면 `attempts`가 1에 멈춘다

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0, upstream 650 green
- [ ] 6.2 §0.3 확인 — 제출 수량·시점 무변화를 diff로 보인다
- [ ] 6.3 §0.4 확인 — `applyFloor`는 브로커에 닿지 않는다(FLM calls 표)
- [ ] 6.4 `make sdd-sync` 재실행 → `make sdd-check`
- [ ] 6.5 **격리 worktree에서** `make gate CHANGE=a091-a-stop-that-sold-nothing-is-critical`
- [ ] 6.6 독립 검증 (구현과 분리된 컨텍스트)
- [ ] 6.7 PM 동기화 → `openspec archive`

## 선후 관계

| change | 관계 |
| --- | --- |
| a089 (outbox 재발 장부) | **독립하지만 상호작용한다.** a091이 만드는 critical이 반복되면 a089가 고치려는 장부 결함을 그대로 밟는다 — 13회가 한 행으로 접히고 `attempts`가 1에 멈춘다. 5.2가 그 사실을 기록한다 |
| a090 (관측 누락 계측) | 독립 |
| a087 (보호 청산의 가격) | 독립. a087은 "무엇을 보내는가", a091은 "0주였을 때 누가 아는가" |
