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
- [x] 0.5 **proposal-freeze 리뷰 — 적대 보이스 1**(Claude, 구현과 분리된 컨텍스트) → `review.md` 「1라운드」 — **REJECT**(P1 4 · P2 6 · P3 6)
- [x] 0.5a **2판 반영**(Manager 2026-09-29: F1~F4 전부, F2 enqueue-only) — design 전면 개정 · `workingSet` AST·FLM·BTM(분기 22) · spec delta 개정 ·
      tasks §2 개정 · 진입 하네스 두 번들. **판정 아님**
- [ ] 0.6 **proposal-freeze 리뷰 — codex 교차 모델**(Manager 슬롯 대기열) → `review.md` · `analysis/freeze-review/`
- [x] 0.7 **Q1·Q3 결정 기록**(Manager 2026-09-29, 사용자행 아님) — Q1 = (a) 정본 준수(`exit-policy/spec.md:62`·`:65`), Q3 = 판정 진입 + 하류 5자리 명명 잔여. design 「Q — 결정 기록」
- [ ] 0.8 **Q2 실측 — 사전 승인됨(Manager 2026-09-29), 구현 로트가 장중 실행** — 정지·0가격 종목 포함 `/prices` 읽기 전용 GET 1회, 쓰기 0.
      결과로 design D6 의 두 [미측정] 행을 확정한다. 구현을 막지 않는다

## 1. Pre-Edit

- [ ] 1.0 **Pre-Edit 선언** — `internal/app/engine/exitloop.go` `ExitObserver.ObserveOnce`·`ExitObserver.workingSet`(기존) · `ExitCycle`·`ExitObserver`
      구조체 필드 · 새 파일 `exit_unobserved.go`. 호출자: `Run`(`exitloop.go:354`) · tracer `Run`(`tracer.go:273`, B6·B7 도달 불가). 불변식: 두 함수의
      분기 조건·이탈 무변화, 새 브로커 호출 0, 새 배선 0(`Retrier.Gate`)

## 2. RED (BTM 「필요한 RED」 — a092 R1~R6 승계 + a090 R7~R14)

- [ ] 2.0 **RED 전 전수 검색(F9)** — 엔진 시험 중 보유 2종목 이상 + 한 종목 미응답 + 가짜 시계 전진 합이 60초 이상인 것을 센다(2판 작성 시 1차 표본:
      `h.entry` 2개 이상 + `Advance` 있는 시험 2 — `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` · `…SlowFirstPosition…`,
      둘 다 60초 미만). 걸리는 시험은 경보·모드 단언이 바뀌는지 먼저 적는다
- [ ] 2.1 **R1** — 보유 2종목, 1종목만 `Last = 0`: `cycle.Err == nil`, 다른 종목 판정, `cycle.Unobserved == 1`, 연속 시작 로그 1줄(`cause=no_quote`)
- [ ] 2.2 **R2** — 1종목이 응답에 **부재**: R1 과 같다
- [ ] 2.3 **R3** — 같은 포지션이 **마지막 판정 뒤** `OutageAfter` 이상 판정 안 됨: `EventExitObservationOutage` 가 `EnqueueAlert` 로 **1회**,
      key = `type|account|position|<연속 기점>`, 다음 주기 반복 없음. 필드 `position_id`·`symbol`·`unobserved_seconds`·`cause`, 본문이 포지션 명명
- [ ] 2.3a **R3 (Q1=a)** — 같은 순회 뒤 `EscalateOperatingMode(…ModeTriggerExitObservationOutage…)` 1회, 같은 연속 반복 없음
- [ ] 2.3b **R3 (F1 — 기점)** — 판정 → 양보 주기들(B1) 55초 → 부분 응답 주기: 기점이 **판정 시각**이라 60초 경과 즉시 경보(첫 미스부터 60초 더 기다리지 않는다).
      전 종목 실패(B4) 주기를 끼운 판본도 같다
- [ ] 2.3c **R3 (F2 — 순서)** — 막힌 알리미 스파이 + 같은 주기에 임계 넘는 포지션 A 와 손절 조건의 포지션 B: B 의 기록·제출이 A 의 적재보다 **먼저**,
      관측 루프의 `ExitAlerter.Notify` 호출 0(enqueue-only)
- [ ] 2.3d **R3 (적재 실패)** — `EnqueueAlert` 실패 fixture: `BlockUnlessClearedSince(ReasonAlertUndelivered, …)` 로 진입 잠금, 로그 1줄, 루프 계속
- [ ] 2.3e **R3 (에피소드)** — 연속 해제(판정) 뒤 새 연속 → 새 기점이라 새 행. 같은 연속의 재적재는 옛 행 재사용·재전송 없음
- [ ] 2.4 **R4** — 미관측 포지션이 다음 주기에 판정에 닿으면 연속 해제 로그 1줄(`unobserved_seconds`), 래치 해제
- [ ] 2.5 **R5** — 전 종목 미응답(B4): 계정 사다리 **무변화**, 그 주기에 포지션 경보 없음
- [ ] 2.6 **R6** — 양보(B1): **무변화**, `checkOutage` 계속
- [ ] 2.7 **R7** — B7(임대 만료)로 빠진 포지션도 기록(`cause=quote_expired`). 임계 전은 경보 아님
- [ ] 2.8 **R8** — 미관측 중 보유 대상에서 사라진 포지션: 기록 정리, 경보 없음. **정리 기준은 `states` 가 아니라 보유 대상 표시**(F4)
- [ ] 2.9 **R9** — 기존 `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol`(:759) ·
      `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection`(:833) **무변화 통과** + 2.0 의 전수 결과
- [ ] 2.10 **R10 (§0.4)** — 가격 읽기 호출 수가 주기당 1 그대로
- [ ] 2.11 **R11 (재시작)** — 관측자를 새로 만들면 기록이 비고 기점이 새로 찍힌다(D2 한계) — 그 뒤 60초 미관측이면 **새 에피소드**로 경보
- [ ] 2.12 **R12 (F4 — workingSet B8)** — 진입 결정에 손절이 없어 exit state 열기가 매 주기 실패하는 보유 포지션: 미관측으로 세어져 60초 뒤 경보.
      그 포지션 하나뿐이라 `states` 가 비는 경우(B3 조기 반환)에도 같다(F7)
- [ ] 2.13 **R13 (F4 — 격리 읽기·쓰기 실패)** — `ActiveExitSnapshotQuarantine`·`QuarantineExitSnapshot` 오류 fixture: 미관측으로 센다
- [ ] 2.14 **R14 (명명 잔여 고정)** — 미관리(B6 `:512`)·완료 정책(B10 `:533`) 포지션은 세지 않는다(범위 밖을 시험으로 못 박는다). 격리 포지션(`refused` →
      `alertRefused`)은 관측됨으로 친다(F10)

## 3. GREEN

- [ ] 3.1 `exit_unobserved.go` — 지연 초기화 기록 맵, 보유 대상 표시 · 원인 기록 · 판정 표시 · 순회 뒤 판정(기점 = 마지막 판정/처음 봄) ·
      `Journal.EnqueueAlert`(에피소드 key) · 적재 실패 시 `Retrier.Gate.BlockUnlessClearedSince` · (Q1=a) `EscalateOperatingMode` · 정리 · 로그
- [ ] 3.2 `workingSet` 편집 — 보유 대상 표시 1자리(`:520`). `ObserveOnce` 편집 — B6·B7 원인 기록 2 · 판정 표시 1 · B3 앞 처리 1 · 순회 뒤 처리 1.
      두 함수 분기 조건·이탈 무변화(편집 후 AST 로 조건 원문 동일 확인)
- [ ] 3.3 `ExitCycle.Unobserved` 필드

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
| §4 손절·비상 청산 즉시성 | 판정·발의·제출 경로 편집 0. **새 경보는 순회 뒤 · enqueue-only**(design D4) — 뒤 포지션의 손절 판정 앞에 서지 않는다(2.3c) |
| §5 High-risk | 손절 관측 경로 — Pre-Edit 1.0, freeze 0.5·0.6 |
| §6 보수 방향만 | 추가는 경보와(Q1) 진입 **조이기**뿐. 손절·익절·사이징 불변 |
| §7 운영 토글 flip·live 검증은 사람 | 5절 |
| §8 시크릿·계좌 정보 | 경보 필드는 계정 ref·종목·포지션 id·초·원인만 |
