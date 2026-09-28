# a090 · tasks

- **Change**: `a090-an-unobserved-position-is-counted`
- **위험 등급**: **High-risk** — 손절 관측 경로. Pre-Edit 선언과 proposal-freeze 리뷰(적대 보이스 1 + codex)가 구현 전에 필요하다.
- **이 change 는 판정·발의·주문을 바꾸지 않는다.** 더하는 것은 포지션 단위 미관측 기록·경보(·Q1 에 따라 모드 강화)뿐이다.

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정(`capture_change_base.py`)
- [x] 0.2 `openspec validate a090-an-unobserved-position-is-counted --strict --no-interactive`
- [x] 0.3 **AST 산출물이 문서보다 먼저** — `ExitObserver.ObserveOnce`(분기 8 · return 5), FLM · BTM · 진입 실측
      (`analysis/harness/observeonce_entry.sh` → `observeonce.blocks`, commit `b0a202b8`, 깨끗한 detached worktree)
- [ ] 0.4 `check_analysis.py --change a090-…` 통과
- [ ] 0.5 **proposal-freeze 리뷰 — 적대 보이스 1**(Claude, 구현과 분리된 컨텍스트) → `review.md`
- [ ] 0.6 **proposal-freeze 리뷰 — codex 교차 모델**(Manager 슬롯 대기열) → `review.md` · `analysis/freeze-review/`
- [x] 0.7 **Q1·Q3 결정 기록**(Manager 2026-09-29, 사용자행 아님) — Q1 = (a) 정본 준수(`exit-policy/spec.md:62`·`:65`), Q3 = 판정 진입 + 하류 5자리 명명 잔여. design 「Q — 결정 기록」
- [ ] 0.8 **Q2 실측 — 사전 승인됨(Manager 2026-09-29), 구현 로트가 장중 실행** — 정지·0가격 종목 포함 `/prices` 읽기 전용 GET 1회, 쓰기 0.
      결과로 design D6 의 두 [미측정] 행을 확정한다. 구현을 막지 않는다

## 1. Pre-Edit

- [ ] 1.0 **Pre-Edit 선언** — `internal/app/engine/exitloop.go` `ExitObserver.ObserveOnce`(기존) · 새 파일 `exit_unobserved.go` ·
      `ExitObserver` 필드 2개. 호출자: `Run`(`exitloop.go:354`) · tracer `Run`(`tracer.go:273`). 불변식: B1~B4·B8 무변화, 새 브로커 호출 0

## 2. RED (BTM 「필요한 RED」 — a092 R1~R6 승계 + R7~R9)

- [ ] 2.1 **R1** — 보유 2종목, 1종목만 `Last = 0`: `cycle.Err == nil`, 다른 종목 판정, 빠진 포지션의 미관측 시작 시각이 기록된다
- [ ] 2.2 **R2** — 1종목이 응답에 **부재**: R1 과 같다
- [ ] 2.3 **R3** — 같은 포지션이 `OutageAfter` 이상 연속 미관측: `EventExitObservationOutage` 가 **포지션 key**
      (`type|account|positionID`)로 **1회**, 다음 주기에 반복 없음. 필드에 `position_id`·`symbol`·`unobserved_seconds`·`cause`
- [ ] 2.3a **R3 (Q1=a 확정)** — 같은 순간 `EscalateOperatingMode(…ModeTriggerExitObservationOutage…)` 1회, 같은 연속에서 반복 없음
- [ ] 2.4 **R4** — 미관측 포지션이 다음 주기에 판정에 닿으면 기록 해제. 다시 빠지면 새 연속(새 시작 시각, 경보 래치 해제)
- [ ] 2.5 **R5** — 전 종목 미응답(B4): 계정 사다리 **무변화**, 그 주기에 포지션 경보 없음
- [ ] 2.6 **R6** — 양보(B1): **무변화**, `checkOutage` 계속
- [ ] 2.7 **R7** — B7(임대 만료)로 빠진 포지션도 기록된다(`cause=quote_expired`). 임계 전 한 번은 경보 아님
- [ ] 2.8 **R8** — 미관측 중 보유에서 사라진 포지션: 기록 정리, 경보 없음
- [ ] 2.9 **R9** — 기존 `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol`(:759) ·
      `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection`(:833) **무변화 통과**
- [ ] 2.10 **R10 (§0.4)** — 이 경로가 더하는 브로커 호출 0 — 가격 읽기 호출 수가 주기당 1 그대로
- [ ] 2.11 **R11 (재시작)** — 관측자를 새로 만들면 기록이 비고, 같은 포지션이 계속 미관측이면 임계 뒤 **다시** 경보한다(파일 머리 계약 `:67-68`)

## 3. GREEN

- [ ] 3.1 `exit_unobserved.go` — 지연 초기화 두 맵, `noteUnobserved`(B6·B7) · `clearUnobserved`(판정 도달) · `pruneUnobserved`(순회 뒤) ·
      임계 판정 → `o.alert`(기존) · (Q1=a) `EscalateOperatingMode`(기존)
- [ ] 3.2 `ObserveOnce` 편집 — 호출 4자리만. 분기 조건·이탈 무변화(편집 후 AST 로 B1~B8 조건 원문 동일 확인)
- [ ] 3.3 경보의 전달 형태는 a094 6판의 enqueue-only 요구가 정해지면 그 형태를 쓴다(design D7). 정해지기 전이면 오늘의 `o.alert`

## 4. 게이트

- [ ] 4.1 FLM·AST 재생성(구현 후) + BTM 재번호(difflib) + `check_analysis.py` 통과
- [ ] 4.2 `go test ./internal/app/engine/ -count=1` · `make test` · `make test-seams` · `make lint`
- [ ] 4.3 **§0.3 확인** — 판정·발의·주문 경로 무변화(`judge` 이하 편집 0, AST 비교)
- [ ] 4.4 **§0.4 확인** — 새 브로커 호출 0(2.10)
- [ ] 4.5 **토글** — 도입하지 않는다(무도입)
- [ ] 4.6 `make sdd-sync` → `make sdd-check`
- [ ] 4.7 격리 worktree 에서 `make gate CHANGE=a090-an-unobserved-position-is-counted`
- [ ] 4.8 독립 리뷰(구현과 분리된 컨텍스트, 교차 모델)
- [ ] 4.9 PM 동기화 → `openspec archive`

## 5. 배포 — 사람이 승인한다

- [ ] 5.1 배포 전 `main` 과 SchemaVersion 대조(이 change 는 스키마를 바꾸지 않는다)
- [ ] 5.2 엔진 재시작은 사람이 승인하고 두 시장이 닫힌 창에서만(엔진 정지 = 손절 없음)
- [ ] 5.3 배포 뒤 첫 포지션 단위 경보의 실물 확인 — 그 포지션이 왜 답하지 않았는지(Q2 와 같은 질문)

## 안전 불변식 확인

| 불변식 | 이 change 에서 |
| --- | --- |
| §1 사람 승인 없는 LIVE 주문 side effect 금지 | 주문 경로 편집 0. 시험은 fixture |
| §2 `mutating: true` 자동 실행 금지 | 준수 |
| §3 토글 OFF = upstream | 토글 없음 |
| §4 손절·비상 청산 즉시성 | 판정·발의·제출 경로 편집 0. 경보 전달의 동기 지연은 이름 붙인 잔여(design D7) |
| §5 High-risk | 손절 관측 경로 — Pre-Edit 1.0, freeze 0.5·0.6 |
| §6 보수 방향만 | 추가는 경보와(Q1) 진입 **조이기**뿐. 손절·익절·사이징 불변 |
| §7 운영 토글 flip·live 검증은 사람 | 5절 |
| §8 시크릿·계좌 정보 | 경보 필드는 계정 ref·종목·포지션 id·초·원인만 |
