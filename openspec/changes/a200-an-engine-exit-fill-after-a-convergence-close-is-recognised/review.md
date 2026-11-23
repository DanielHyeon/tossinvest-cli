# a200 review

## 1단계(등록 · 재현 · 증거) — 2026-11-24, Opus 팀메이트

### 첫 판: 멈춤

1단계 중 「멈추고 돌아올 조건」 둘이 걸려 정본 delta · FLM · 리뷰 전에 멈췄다.
1. 편집할 기존 함수가 둘을 넘음(조건부) — 편입 포지션 성과는 `adoptionEpoch` 를 고쳐야 함(재현 탐침 `post_adoption_disposal`).
2. design 결정과 사실 불일치 — (f) 체결 투영 · Exit 훅은 최신 인스턴스를 읽음, D3 의 수량을 실을 열이 스키마 36 에 없음.

문서에 바로 반영한 사실 정정: proposal 비목표 · design §4 의 「W2 · W3 재현 전」 → 「재현됨」.

### Manager 결정 (2026-11-24) — 범위 축소

| # | 결정 | 반영 자리 |
|---|---|---|
| 1 | 성과(F15)를 뺌 — `ApplyExitFill` 무편집, 인정된 청산은 W6 모양(`ADJUSTMENT_CLOSED` + `PROPOSAL_FILLED`, `COMPLETED` 0, 성과 0) | proposal 비목표 · design D5 · Story limits 첫 줄 · out-of-scope S3 |
| 2 | 수량 저장 안 함, (라)는 상태 없는 합 ≤ 감소분, 대가(조정을 사이에 둔 부분 체결)를 이름으로 | design D2 · D2a · §2 「이 change 가 거절하는 정상 입력」, 재현 시험 `…StraddlingTheConvergence…` |
| 3 | 표지 행 — 모든 소비자 「무시됨」 일 때만 | `analysis/edit-surface.md` 소비자 표 → 오분류 · 깨짐 0 → **표지를 씀**(design D3) |
| 4 | 운영 로그 뺌 | proposal 비목표 |
| 5 | (f) 범위 밖, 사실로 적고 별도 Story 후보 | design §2 「범위 밖 사실」 · Story limits · out-of-scope S1 |
| 6 | W2 · W3 범위 밖 · 재현됨 | proposal 비목표 · out-of-scope S2 |
| 7 | a137 측정 불가 limits | proposal 비목표 · Story limits |
| 8 | 계속 거절 6(kind) 유지 + 사실 | design D2 (나) · §2-6 |
| 9 | 편집할 기존 함수 하나 `ProjectPosition` | design §5 · edit-surface · FLM 번들 |
| 10 | base 재고정 승인 — 실행은 a198 착지 · merge 뒤 | 아래 「base 재고정」 · tasks 1.2 |
| 11 | γ 보류 확정 + 다시 여는 조건 | proposal 「γ」 절 |

reconciliation delta 는 결정 본문의 「귀속 실패 … Requirement 에 Scenario 추가」 대신 **ADDED Requirement** 로 만들었다. 그 Requirement(「귀속 실패 종목 차단은 … 운영자가 해제할 수 있다」)의 본문이 「이 요구는 … 차단을 거는 조건 … 을 바꾸지 않는다(SHALL NOT)」 이라, 차단을 거는 조건이 바뀌는 Scenario 를 그 안에 넣으면 같은 Requirement 안에서 모순이 된다. 내용(인정된 체결은 차단을 안 만듦 · 어긋나면 지금처럼)은 같다.

### base 재고정 (Manager 승인 2026-11-24 — 실행은 a198 착지 · merge 뒤)

- 요건 (1) 귀속 실측: base `884be2a74` 이후 이 change 의 자기 Go 커밋은 새 시험 파일 2 개(`internal/journal/a200_convergence_close_exit_fill_repro_test.go` · `internal/reconcile/a200_convergence_close_exit_fill_cycles_test.go`)뿐이고 고친 기존 함수 0 — 원문 `analysis/base-refix/attribution.txt`(이 커밋 시점에 잼 — 재고정 직전에 다시 잰다).
- 요건 (2) 승인 기록: 이 절과 tasks 1.2.
- 요건 (3) `base-commit.txt` 한 파일 단독 커밋 + 영수증(옛 → 새 sha, 실측, 옛 base 에서의 required 수).

### 제안 동결 적대 리뷰(1.12) — codex r1, 2026-11-24

보이스: codex-cli 0.165.0, `exec -s read-only`, `</dev/null`, 작성자(Claude)와 다른 계열. 프롬프트 `analysis/review/proposal-freeze-prompt.md`, 원문 `analysis/review/proposal-freeze-r1-codex.md`. 대상 HEAD `b62dd4cfd` + 작업 트리. codex 는 Go 시험을 돌리지 않았다고 적음.
결과: **P0 0 · P1 2 · P2 3.** 「멈출 조건」(설계 결정에 닿는 P1)에 걸려 **2회차는 돌리지 않았다** — Manager 결정 뒤 개정판으로 돌린다.

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| R1-1 | P1 | D2a 의 `fill_events.delta_quantity` 합은 실제 누적보다 작아질 수 있다 — delta 는 float64 차, `nearlyZero`(상대 1e-9)가 작은 차를 0 으로 만들어 행이 안 남는데 스냅숏은 전진(합성 대수량 경계, 계산 확인) | **반영**(D2a 출처는 Manager 가 위임한 선택): 주문별 브로커 원문 누적(`fill.CumulativeQuantity` · `fill_snapshots.filled_quantity` 최댓값)으로 바꿈. design D2a · §3 · ledger delta · claim-check C-D2a. 2회차에서 확인 필요 |
| R1-2 | P1 | D2 는 주문 소유를 증명하지만 그 체결이 조정의 원인임은 증명 못 함 — 엔진 매도 대기 중 외부 처분 → 조정 31→0 → 재매수가 **원장에 접히기 전** 옛 엔진 매도 체결이면 D2 전부 통과(설계 반례, 브로커 허용 여부 미확인) | **Manager 결정 필요** — 아래 질문 |
| R1-3 | P2 | 「조정을 사이에 둔 부분 체결은 거절」 은 중간 체결에서 거짓(조정 뒤 1 · 20 이면 첫 1 은 누적 11 ≤ 21 로 인정) | **반영**(사실 정정): design §2 · ledger Scenario 를 「누적이 감소분을 넘는 체결부터 거절(결국 AF)」 로, §3 에 나눈 판 시험 추가 |
| R1-4 | P2 | 재기동 뒤에는 인정돼도 QM 이 다음 일치에서 안 풀림(크레딧은 프로세스 기억, `Tracker.Restore`) | **반영**(공시): design D6 · reconciliation delta 에 「재기동이 끼면 지금처럼 조정 대기」, §3 공시 시험, Story limits |
| R1-5 | P2 | 「어긋나면 AF 를 만든다」 는 활성 행이 있으면 거짓(F10) | **반영**: reconciliation delta 를 「활성 차단이 없으면 AF, 있으면 새 행 안 씀 — 지금과 같이」 로 좁힘 |

표지 행(D3) 반례 없음, MODIFIED 의 의도 외 변경 없음(codex 확인).

### Manager 에게 묻는 것 (R1-2)

「조정 뒤 · 재매수 접기 전」 창에서 옛 엔진 매도의 체결이 재매수분을 판 것이어도 인정된다. 원장에 조정 스냅숏과 체결의 인과를 가를 열은 없다.
- (A) 받아들이고 근거를 적는다 — 미체결 엔진 매도가 묶은 수량은 브로커가 사람의 앱 매도에서 빼므로(매도가능수량) 「외부 처분이 엔진 주문의 수량을 판다」 는 일어나기 어렵다(미측정). 틀려도 수량은 계좌 값이라 맞고, 잃는 것은 그 종목의 차단 하나(귀속)다.
- (B) 「주문 접수가 조정보다 앞섬」 조건 추가 — 반례의 주문도 조정 전에 있었으므로 효과 없음.
- (C) 인정하지 않는다(β 포기).
- 추천 (A) + 공시.
