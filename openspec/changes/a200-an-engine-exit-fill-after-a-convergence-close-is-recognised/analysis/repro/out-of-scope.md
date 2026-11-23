# a200 1.5 — 범위 밖 재현 시도(수리 안 함, 결과만)

시험 파일 `internal/reconcile/a200_convergence_close_exit_fill_cycles_test.go`(실제 원장 · 운영 훅 리터럴 · 실제 `Converger` · 실제 `Tracker`,
RunOnce 순서 수렴 → Refresh → Observe 를 손으로 밟음). 원문 `cycles.txt`. 픽스처는 합성 값.

## (i) W2 · W3 — 체결이 수렴보다 먼저 커밋될 때의 수량 불일치 고착: **재현됨**

| 창 | 만든 방법 | 관측(원문) |
|---|---|---|
| W2 | 비교(로컬 31 · 계좌 0)를 만든 뒤 · 수렴 전에 손절 체결 커밋 | 수렴 `refused` = 「the projection holds no live instance」, 조정 0 · 크레딧 0. Observe 가 `QUANTITY_MISMATCH` 행 삽입. **일치 비교 5 주기 모두 `awaiting=1`, 활성 행 QM 그대로** |
| W3 | `ConvergeStore` 를 감싸 `FillWatermark` 직후(조정 tx 전)에 체결 커밋 | 조정 tx 재검증이 「expected quantity 31 but found 0」 으로 stale → 수렴 오류(RunOnce 는 기록만 하고 계속). 이후 W2 와 같음 — **5 주기 뒤 QM 고착** |

- 근거 사슬은 사전 증거 Q1-1 의 AST 서술과 같음(크레딧 없음 → 깨끗한 비교의 해제 루프가 `AwaitingAdjustment`, a192 결속 없음).
- 시험은 이 고착을 단언으로 못 박음(`TestA200CurrentBehaviourW2…` · `…W3…` — 지금 동작 PASS).
- 운영 지금 상태: 미해제 QM 0(운영 집계). 고착되면 a194 가 `blocked_stale` 로 알리고 운영자 해제로 풀림(코드 서술, 이 1단계는 a194 경로를 돌리지 않음).
- 판정: β 와 별개의 결함(크레딧을 못 얻는 수량 불일치). **별도 change 후보**.

## (ii) 부분 체결 중간값을 본 수렴 뒤의 초과 체결: **재현됨 — 고착 없음**

- 계좌 11 · 로컬 31 → 수렴이 31 → 11 제자리 조정(인스턴스 열린 채, 크레딧 기록) → 그 매도의 누적 20 체결(delta 20 > 11) →
  투영 거절 `X49 (OVERSELL)` → `QUANTITY_MISMATCH` 행(체결 tx 가 씀) → 투영 11 그대로.
- 다음 일치 비교(계좌 11 = 로컬 11)에서 그 QM 행이 `ADJUSTMENT_APPLIED` 로 풀림 — 투영 거절이 만든 행이어도 cause 가 QM 이라 크레딧 해제 대상.
- 잔량 11 체결(누적 31, delta 11)은 11 → 0 으로 정상 청산.
- 판정: 사전 증거 「인접 모양」 서술(AF 아님 · 크레딧으로 풀림)과 같음. 범위 밖 유지, 추가 조치 불요. 단 그 사이 투영 수량(11)은 실제(엔진이 20 을 팔고 계좌는 11)와 맞고,
  체결 20 은 투영에 흔적이 없음 — 성과 계산에는 `fill_events` 로 들어감(이 1단계는 성과를 재지 않음).

## 덤 — 재매수 뒤 옛 손절 체결(전제 (f), journal 시험)

`internal/journal/a200_convergence_close_exit_fill_repro_test.go` 의 두 시험(`…AReBuyLandsOnTheNewInstance`, `…ReducesTheNewInstance`):
수렴이 닫은 뒤 사람의 재매수가 새 EXTERNAL 인스턴스로 접힌 상태에서 옛 손절 체결이 오면, 체결은 **새 인스턴스**에 닿음 —
전량이면 OVERSELL → QM, 부분이면 **차단 없이 사람의 새 보유 수량을 줄임**(5 → 2, 상태 CLOSING). 옛 인스턴스의 `pending_intent_id` 는
Exit 훅도 최신 인스턴스를 읽으므로 끝내 비워지지 않음. 이것은 β 범위 밖의 별도 결함(인스턴스 귀속) 후보이고, design §2 「계속 거절 5」 ·
「정상 입력 3」 의 전제를 무너뜨림(`claim-check.md` C-f).
