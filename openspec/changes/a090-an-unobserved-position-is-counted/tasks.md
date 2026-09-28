# a090 · tasks

- **Change**: `a090-an-unobserved-position-is-counted`
- **위험 등급**: **High-risk** — 손절 관측 경로. Pre-Edit 선언과 proposal-freeze 리뷰(적대 보이스 1 + codex)가 구현 전에 필요하다.
- **이 change 는 판정·발의·주문을 바꾸지 않는다.** 더하는 것은 포지션 단위 미관측 기록·경보(·Q1 에 따라 모드 강화)뿐이다.

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정(`capture_change_base.py`)
- [x] 0.2 `openspec validate a090-an-unobserved-position-is-counted --strict --no-interactive`
- [x] 0.3 **AST 산출물이 문서보다 먼저** — `ExitObserver.ObserveOnce`(분기 8 · return 5), FLM · BTM · 진입 실측
      (`analysis/harness/observeonce_entry.sh` → `observeonce.blocks`, commit `b0a202b8`, 깨끗한 detached worktree)
- [ ] 0.4 `check_analysis.py --change a090-…` 통과 — **설계 단계 rc 1(이웃 a066 v35 착지, 자기 Go 0)**. 구현 로트 첫 행위로 사람 승인 base 재고정(영수증) 뒤 통과(Manager 2026-09-29)
- [x] 0.5 **proposal-freeze 리뷰 — 적대 보이스 1**(Claude, 구현과 분리된 컨텍스트) → `review.md` 「1라운드」 — **REJECT**(P1 4 · P2 6 · P3 6)
- [x] 0.5a **2판 반영**(Manager 2026-09-29: F1~F4 전부, F2 enqueue-only) — design 전면 개정 · `workingSet` AST·FLM·BTM(분기 22) · spec delta 개정 ·
      tasks §2 개정 · 진입 하네스 두 번들. **판정 아님**
- [x] 0.6 **proposal-freeze 리뷰 — codex 교차 모델** — **REJECT**(P0 2 · P1 4 · P2 4), `review.md` 「codex 1라운드」 · `analysis/freeze-review/codex-r1-output.md`
- [x] 0.6a **3판 반영**(Manager 2026-09-29: N1~N6 · P2 넷) — design D1·D2·D3·D4·D5·D7·D8·D10·D11 · `AnnounceOperatingMode` AST·FLM · spec delta · tasks. **판정 아님**
- [x] 0.6b **codex 2라운드** — **REJECT**(P0 1 · P1 2 · P2 3), `review.md` 「codex 2라운드」 · `analysis/freeze-review/codex-r2-output.md`
- [x] 0.6c **4판 반영**(Manager 2026-09-29 R2 처분 + a092 교차) — design D1·D3·D4·D5·D7·D8·D10·D12 · spec delta · tasks · workingSet FLM. **판정 아님**
- [x] 0.6d **codex 3라운드** — **REJECT**(P0 0 · P1 4 · P2 3 · P3 1), `review.md` 「codex 3라운드」 · `analysis/freeze-review/codex-r3-output.md`
- [x] 0.6e **5판 반영**(Manager 2026-09-29) — R3-1 tasks 수리 · R3-2/R3-3 a090 정화 · R3-4 a092 입구 구현 하드 의존 · P2 셋 · P3. **판정 아님**
- [x] 0.6f **codex 4라운드(좁은 확인)** — **PASS**(P2 3), `review.md` 「codex 4라운드」 · `analysis/freeze-review/codex-r4-output.md`
- [x] 0.6g **freeze 전 P2 반영**(Manager 2026-09-29) — N1 옛 문구 정리(직접 적재 · 삭제된 대안 · 계좌 필드) · N2 실패 카나리 보강(주입 도달 · a090 소유 로그 · 계좌 섞인 오류). N3 는 a092 명확화 후보(기록 유지)
- [x] 0.6h **proposal-freeze 선언**(2026-09-29) — 적대 보이스 1(0.5) + codex 4라운드(0.6 · 0.6b · 0.6d · 0.6f, 마지막 PASS) 완료, 사용자·Manager 결정(Q1~Q3 · R2/R3 처분) 기록 완료. **구현은 0.8(Q2 실측)·0.9(a092 입구 착지 하드 조건)·0.4(base 재고정) 뒤** — Manager 확인 대기
- [ ] 0.9 **구현 하드 조건** — a092 `RecordAlert` 입구가 main 에 착지한 뒤에만 1.x 이후를 시작한다(설계 freeze 는 독립)
- [x] 0.7 **Q1·Q3 결정 기록**(Manager 2026-09-29, 사용자행 아님) — Q1 = (a) 정본 준수(`exit-policy/spec.md:62`·`:65`), Q3 = 판정 진입 + 하류 5자리 명명 잔여. design 「Q — 결정 기록」
- [ ] 0.8 **Q2 실측 — 사전 승인됨(Manager 2026-09-29), 구현 로트가 장중 실행** — 정지·0가격 종목 포함 `/prices` 읽기 전용 GET 1회, 쓰기 0.
      결과로 design D6 의 두 [미측정] 행을 확정한다. 구현을 막지 않는다

## 1. Pre-Edit

- [ ] 1.1 **FLM(편집 전, 3판)** — `engineRuntime`(`cmd/tossctl/engine.go`) 번들 · `AnnounceOperatingMode` BTM 진입 실측(설계 단계 번들은 있음)
- [ ] 1.0 **Pre-Edit 선언** — `internal/app/engine/exitloop.go` `ExitObserver.ObserveOnce`·`ExitObserver.workingSet`(기존) · `ExitCycle`·`ExitObserver`
      구조체 필드 · 새 파일 `exit_unobserved.go` · **3판**: `Notifier.AnnounceOperatingMode`(추출만) · `internal/obs/event.go` 상수 · `cmd/tossctl/engine.go`
      `engineRuntime`(로거 옵션 한 줄). 호출자: `Run`(`exitloop.go:354`) · tracer `Run`(`tracer.go:273`, B6 도달 불가 · **B7 도달 가능**(`:793` 원장 작업이 임대를 태울 수 있음)). 불변식: 두 함수의
      분기 조건·이탈 무변화, 새 브로커 호출 0, 게이트 직접 잠금 0(입구의 몫)

## 2. RED (BTM 「필요한 RED」 — a092 R1~R6 승계 + a090 R7~R14)

- [ ] 2.0 **RED 전 전수 검색(F9)** — 엔진 시험 중 보유 2종목 이상 + 한 종목 미응답 + 가짜 시계 전진 합이 60초 이상인 것을 센다(2판 작성 시 1차 표본:
      `h.entry` 2개 이상 + `Advance` 있는 시험 2 — `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` · `…SlowFirstPosition…`,
      둘 다 60초 미만). 걸리는 시험은 경보·모드 단언이 바뀌는지 먼저 적는다
- [ ] 2.1 **R1** — 보유 2종목, 1종목만 `Last = 0`: `cycle.Err == nil`, 다른 종목 판정, `cycle.Unobserved == 1`, 연속 시작 로그 1줄(`cause=no_quote`,
      이벤트 `exit.position_unobserved`, **severity normal**)
- [ ] 2.2 **R2** — 1종목이 응답에 **부재**: R1 과 같다
- [ ] 2.3 **R3** — 같은 포지션이 **마지막 판정 뒤** `OutageAfter` 이상 판정 안 됨: `EventExitObservationOutage` 가 **a092 입구(`RecordAlert`, 창 0)** 로 **1회**,
      key = `type|position|<연속 id>`(**계좌 없음** — 5판 R3-1 수리), 다음 주기 반복 없음. **원장의 outbox 행과 운영 모드 행을 단언한다**(알리미 스파이가 아니라). 필드 `position_id`·`symbol`·`unobserved_seconds`·`cause`, 본문이 포지션 명명
- [ ] 2.3a **R3 (Q1=a · 3판 N2)** — 같은 순회 뒤 모드 ENTRY_BLOCKED 가 원장에 기록되고, 공지는 **outbox 행**(key `operating_mode:<mode>:<전이 id>`, 필드에 계좌 없음 — 5판 R3-1 수리)으로
      적재만 된다 — 관측 루프의 `Notify` 호출 0. 같은 연속 반복 없음. 두 번째 연속의 강화(운영자 완화 뒤) 공지가 **새 행**으로 적재된다(전이 id)
- [ ] 2.3b **R3 (F1 — 기점)** — 판정 → 양보 주기들(B1) 55초 → 부분 응답 주기: 기점이 **판정 시각**이라 60초 경과 즉시 경보(첫 미스부터 60초 더 기다리지 않는다).
      전 종목 실패(B4) 주기를 끼운 판본도 같다
- [ ] 2.3c **R3 (F2 · 3판 N2 — 순서)** — **실제 Notifier + 막힌 발행자** 구성: ① 같은 주기에 임계 넘는 포지션 A 와 손절 조건의 포지션 B → B 의 기록·제출이
      A 의 적재·강화보다 먼저 ② 강화 뒤 **다음 주기**에 손절 조건의 포지션 C 가 지연 없이 판정·제출된다(가짜 시계로 주기 간격 단언). 관측 루프의 `Notify` 호출 0
- [ ] 2.3d **R3 (적재 실패)** — a092 입구의 기록 실패 fixture: 진입 잠금은 **입구의 생산자 래치**가 한다(a090 은 게이트를 직접 잠그지 않음), 루프 계속
- [ ] 2.3e **R3 (에피소드)** — 연속 해제(판정) 뒤 새 연속 → 새 연속 id 라 새 행. 같은 연속의 재적재는 옛 행 재사용·재전송 없음
- [ ] 2.3f **R3 (3판 N4 · 4판 R2-5 · 5판 R3-7 — 시계)** — **앵커 인식 픽스처**(새로 작성 — a111 분리 픽스처는 역행 뒤 앵커를 틀리게 잰다)에서 벽시계를 뒤로 돌려도
      경과가 늘지 않는다; 앵커 생성 직후 경과 0, 이후 정확한 진행 —
      앵커를 역행 **전과 후** 모두에서 만든 두 연속으로(일반 `clock.Fake` 는 단조가 아니다 — 계약 명시). 같은 벽시계 기점을 갖는 두 연속이
      다른 key 를 받는다
- [ ] 2.3g **R3 (3판 N5 · 4판 R2-2 · R2-4 — 실패 전이)** — ① 적재 실패 → 잠금(입구) + 다음 주기 재시도 → 성공 ② 강화 커밋 실패 → 다음 주기 재시도
      ③ 커밋 성공 → 공지 적재 실패 → 저장소 복구 → **정확히 한 공지 행**(같은 전이 id), 전이 재시도 0, 그 사이 운영자 완화 뒤에도 재강화 0, **연속이 끝난 뒤에도**
      공지 재시도 ④ 강화 뒤 운영자 완화 → 같은 연속에서 재강화 없음 ⑤ 해제 세대: 기록 **중** 해제가 끼어도 잠근다(세대는 오류 반환 직후 — 입구의 몫, a090 경로에서 단언)
      ⑥ 실패 → 해제 → 재시도 → 실패는 다시 잠근다(새 증거)
      ⑦ 세대를 **읽은 뒤·적용 전** 해제가 끼면 잠그지 않는다(codex 3라운드 부수 — 입구가 정본을 지키는지 a090 경로에서도 단언)
      ⑧ 공지 재시도는 처리 주기에서만: 연속이 끝나고 포지션이 사라진 뒤(B3 포함), 그리고 B1/B4 주기가 끼어도 다음 처리 주기에 정확히 한 행(5판 R3-5)
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
- [ ] 2.14 **R14 (명명 잔여 고정 — 3판 N1 로 구현 가능)** — 미관리(B6 `:512`)·완료 정책(B10 `:533`) 포지션은 세지 않는다(범위 밖을 시험으로 못 박는다). 격리 포지션(`refused` →
      `alertRefused`)은 관측됨으로 친다(F10)
- [ ] 2.15 **R15 (4판 R2-3 — 이탈 전수 핀)** — **`go/parser` 로 현재 `exitloop.go` 를 직접 파싱**한다(저장된 `ast.json` 은 문장 순서·`continue` 목록이 없다).
      `workingSet` 의 포지션 순회 본문에서 **모든 `continue`·`return`(재귀)**을 세어 **얼린 목록**과 대조한다(다중집합 — 개수 유지, 짝 없는 이탈은 빨강).
      좌표는 편집에 안정된 **조상 경로**다(5판 R3-6 — B12·B21 이 둘 다 `if qerr != nil` 안이라 감싼 조건만으로는 유일하지 않다): 순회 본문에서 그 이탈까지의
      `if` 조건 원문 목록 + **같은 조건 원문의 출현 순번** + 이탈 종류(`continue`/`return`). 목록: B5(보유 아님) · B6(미관리, 표시 **앞**) · B8(열기 실패 — 그
      `continue` 노드 하나, 블록 전체 허용 아님) · B10(완료 — 해제 호출 **뒤**) · B12 · B14 · B21(격리 오류) · B11 · B17 · B20(refused).
      추가 단언: 표시 호출은 B6 블록 **바로 뒤 형제 문장**, 해제 호출은 B10 블록의 **첫 문장**. 변이 다섯이 각각 빨강 — ① 표시를 B10 뒤로 이동 ② 해제 삭제
      ③ 표시~B10 사이에 `continue` 삽입 ④ **B5~B6 사이(표시 앞)에 새 조기 탈락** ⑤ **B8 블록 안에 새 `return`**
- [ ] 2.16 **R16 (3판 N10 — 「관측됨」 의 열거)** — 판정 진입 뒤 즉시 끝나는 경우(격리 → `alertRefused` · 선택자 스탬프 실패 `:876-879` · 정책 신원 오류)가
      관측됨으로 쳐지고 연속을 끝냄을 고정한다(정의가 바뀌면 빨강)
- [ ] 2.17 **R17 (4판 R2-1 — 전용 로거 배선 + 카나리)** — 생산 조립(`engineRuntime`)으로 만든 관측자의 `opts.UnobservedLog` 가 non-nil 이고 **`opts.Log` 는
      여전히 nil**(기존 줄의 생산 출력 무변화). 부분 미응답 주기에 `exit.position_unobserved` 줄이 실제 로그 출력에 나온다(배선 수준). **카나리**: 계좌 ref 로
      고유 문자열을 쓴 구성에서 그 문자열이 로그 출력 · 적재된 미관측 알림 · 강화 공지 행 어디에도 없다(행의 **key 와 페이로드 둘 다**).
      **5판 R3-3 — 실패 주입 카나리**: ① 미관측 알림 기록 실패 ② 모드 기록 실패 ③ 공지 기록 실패 각각에서 **a090 이 입구에 넘긴 값**(이벤트·key·필드)에 계좌
      문자열이 없다. 공유 경로가 자기 필드로 붙이는 계좌(`notifier.go:387`·`:394`)는 이 시험의 대상이 아니다(사용자 큐 항목).
      **freeze 전 N2 보강(codex 4라운드)**: ⓐ **주입 도달 단언** — 세 실패 각각이 실제로 일어났음을 단언한다(모드 기록 실패에서는 공지가 입구에 **가지 않으므로**
      "입구 인자 무계좌" 는 공허하다 — 그 경우는 공지 대기열에 레코드가 **없음**과 강화 재시도 상태를 단언한다) ⓑ **a090 소유 로그 검사** — 세 실패 주입에서
      `UnobservedLog` 로 나간 줄 전체(필드·메시지)를 읽어 계좌 문자열이 없음을 단언한다 ⓒ **계좌 섞인 오류 주입** — 주입하는 오류 문자열 자체에 계좌 카나리를
      넣어, a090 소유 로그가 원문 오류를 옮기지 않음(오류 **종류**만 싣는다)을 단언한다

## 3. GREEN

- [ ] 3.1 `exit_unobserved.go` — 지연 초기화 기록 맵, 보유 대상 표시 · 원인 기록 · 판정 표시 · 순회 뒤 판정(기점 = 마지막 판정/처음 봄) ·
      **a092 입구 `RecordAlert`(창 0, 에피소드 key)** — 직접 `EnqueueAlert`·게이트 잠금 없음(5판 R3-1) · (Q1=a) `EscalateOperatingMode`(a090 전용 정화 공지자) · 정리 · 로그
- [ ] 3.2 `workingSet` 편집 — 보유 대상 표시 1자리(`:520`). `ObserveOnce` 편집 — B6·B7 원인 기록 2 · 판정 표시 1 · B3 앞 처리 1 · 순회 뒤 처리 1.
      두 함수 분기 조건·이탈 무변화(편집 후 AST 로 조건 원문 동일 확인). **3판**: `workingSet` 은 표시 1 + 해제 1(B10 첫 문장)
- [ ] 3.4 `obs.OperatingModeEvent` 추출(`AnnounceOperatingMode` 동작 무변화 — **정확한 Event 동등 단언**) · enqueue-only 공지자(key `operating_mode:<mode>:<전이 id>`,
      `FieldAccount` 제거, **a090 전용 정화 어댑터**가 a092 입구 위에서 기록 — a092 announcer 재사용 안 함, 5판 R3-2) · `EventExitPositionUnobserved`(normal) ·
      `ExitObserverOptions.UnobservedLog` + `engineRuntime` 에 `UnobservedLog: logger`
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
| §8 시크릿·계좌 정보 | 새 경보·공지·로그의 필드와 key 에 **계좌 ref 없음**(종목·포지션 id·초·원인·전이 id 만, design D4·D5·D7·D12) — 카나리 2.17 |
