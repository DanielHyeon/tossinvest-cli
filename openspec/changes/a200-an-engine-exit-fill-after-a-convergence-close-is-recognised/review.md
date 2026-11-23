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

reconciliation delta 는 **ADDED Requirement** 로 만들었다(Manager 승인 2026-11-24). 이유: 기존 Requirement(「귀속 실패 종목 차단은 … 운영자가 해제할 수 있다」)의 「**이 요구는** … 차단을 거는 조건 … 을 바꾸지 않는다」 는 그 Requirement 자신의 범위를 말할 뿐 다른 Requirement 를 금하지 않는다 — 그래서 모순은 없고, 「차단을 거는 조건」 의 변화는 별도 ADDED 가 자기 이름으로 진다. ADDED 본문에 「귀속 실패 차단의 해제 규칙(운영자 해제)은 바꾸지 않는다」 를 넣었다.

### Manager 결정 2 (2026-11-24, 1단계 보고 2 뒤)

1. codex r1 P1-2(인과 미증명) — **(A) 받아들이고 공시**. design §2 「인과는 …」 · §4, Story limits. 뒷받침 측정 M29 · M33(미체결 지정가 매도가 매도가능수량 예약), 앱 매도 · 부분 체결 잔량은 미측정 가정.
2. reconciliation ADDED — 승인(위 이유).
3. 표지 행 — 승인, action `ADJUSTMENT_CLOSE_FILL_RECOGNISED`(design D3).
4. 누적 출처(주문별 원문 누적) — 승인, 정정 계보 이중 계산 없음을 시험 한 사례로(design §3-3).
5. codex P2 셋 반영 — 승인.

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
- 추천 (A) + 공시. → **Manager (A) 결정**(위 「Manager 결정 2」).

### 제안 동결 적대 리뷰 — codex r2(마지막), 2026-11-24

보이스 r1 과 같음. 프롬프트 `analysis/review/proposal-freeze-r2-prompt.md`, 원문 `analysis/review/proposal-freeze-r2-codex.md`. 대상 HEAD `b15a1f489` + 작업 트리(Manager 결정 2 반영판). codex 는 시험을 돌리지 않음.
결과: **P0 0 · P1 2 · P2 2.** P1 둘 다 D2a(Manager 가 승인한 누적 출처)에 닿는다 → 지시대로 **반영하지 않고 멈춤. freeze 안 함, merge · base 재고정 · 구현 착수 안 함.**

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| R2-1 | P1 | `fill_snapshots` 는 `order_id` 기본키 호환 표라 범위가 다른 같은 번호의 최신 누적이 없을 수 있다(다른 거래일 같은 번호 P 의 옛 0 행이 남고 오늘 P 의 10 은 `scoped_fill_snapshots` 에만, fills.go:730-799) → 부모 0 + 자식 21 = 21 로 인정, 실제 31 | **Manager 결정 필요** |
| R2-2 | P1 | 최신 원문 스냅숏도 `nearlyZero` 범위의 작은 **감소** 재관측으로 덮어써져(fills.go:413 · :437 · :742) 관측 이력의 최댓값보다 작아질 수 있다 — 「남는 오차는 더 세는 쪽뿐」 반증(합성 소수 경계) | **Manager 결정 필요** |
| R2-3 | P2 | 정정 계보 정상 시험 계획이 부모의 `REPLACED`(종결) 관측 순서를 빠뜨림 — 그 관측의 Exit 훅이 pending 을 비우므로(ApplyExitFill:1009, successor 검사 없음) 자식 체결은 (다)에서 거절 | 기록만(멈춤) — 반영안: 「부모 종결 관측이 자식 체결보다 앞서면 거절」 을 limits 로, 시험 사례 추가 |
| R2-4 | P2 | M29 · M33 은 보유 조회와 주문 서비스 사이의 일관성(지연된 0 응답)을 배제하지 않음 | 기록만 — 반영안: design §2 「인과는 …」 · Story limits 에 미측정 가정 한 줄 |

### Manager 에게 묻는 것 (R2-1 · R2-2)

「주문별 원문 누적의 현재 값」 은 두 길(호환 표의 범위 충돌 · 작은 감소 덮어쓰기)로 과소 계상될 수 있다. 관측 이력 최댓값은 원장에 보존되지 않는다(스냅숏은 덮어쓰고 `fill_events` 는 float 차를 0 으로 버림).
- (A) **수량 상한으로 바꾼다**: (라)를 「그 intent 의 주문 수량(원주문 수량) ≤ 조정 감소분」 으로. 누적 체결은 원주문 수량을 넘을 수 없으므로 과소 계상 경로가 없다. 출처 후보: `intents.quantity`(intent 하나에 하나, 엔진 손절은 보유 전량) — 단 정정이 수량을 **늘리면** 상한이 깨지므로 정정 계보가 있으면(그 intent 의 CONFIRMED AMEND attempt 존재) 인정하지 않는다. 대가: 조정을 사이에 둔 부분 체결은 조정 뒤 첫 체결부터 거절(지금 문서의 「넘는 체결부터」 보다 엄격), 정정 계보는 전부 거절, 부분 익절(주문 수량 < 보유)은 영향 없음.
- (B) 세 출처의 최댓값(`scoped_fill_snapshots` · `fill_snapshots` · `fill_events` 합 · 이번 관측)을 decimal 로 — R2-2(덮어쓴 감소)를 막지 못한다.
- (C) 인정 포기.
- 추천 (A). 결정되면 design D2 (라) · D2a · §2 · §3 · ledger delta · claim-check 를 고치고 codex 를 다시 돌릴지(3회차) 정해 주길 바란다 — 2회 상한을 넘으므로 Manager 승인이 필요하다.

### Manager 결정 3 (2026-11-24) — (라) 수량 상한, codex 3회차 없음

- (라) = 「그 intent 의 PLACE · AMEND attempt 가 정확히 하나(CONFIRMED PLACE) ∧ `intents.quantity` ≤ 조정 감소분」. 반영: design D2 · D2a · §2 · §3 · §4 · §5, proposal, ledger delta(조건 · Scenario 2 개 문장 — 생성 스크립트 재실행, 역치환 단언, strict valid), claim-check C-D2a · C-§2 대가 · C-종결 뒤 · 전제 (i)~(iii), Story acceptance · limits, edit-surface(a167 줄).
- R2-3 반영: 「종결 관측이 pending 을 비운 뒤 온 체결」 을 계속 거절 3 에. R2-4 반영: 보유 조회 지연을 미측정 가정으로 design §2 · Story limits.
- **3회차를 하지 않은 이유**: (라)가 원장 누적을 읽지 않는 정적 상한으로 바뀌어 1 · 2회차의 P1 넷이 지목한 출처(`fill_events` 합 · `fill_snapshots` · 원문 누적의 덮어쓰기) 자체가 판정에서 없어졌다. 단순해진 설계의 반례 탐색은 구현 적대 리뷰(tasks 3.1)가 맡는다.

## Manager freeze 2026-11-24

(라) 수량 상한. codex 2회차 P1 2건은 출처 교체로 해소. 전제 확인값(claim-check 「전제 (i)~(iii)」):
- (i) 주문 수량 = `intents.quantity`(schema.go:212) — `Gateway.prepareRequest` decimalString@1053 ← `Gateway.place` :374 `intent.Quantity`, 브로커 호출 `g.trading.Place@400` 은 같은 intent 복사본, intent 행 불변(`Journal.Prepare` B6 if@392 `ErrIntentMutated`).
- (ii) 청산 주문 수량 ≤ 인스턴스 수량: `ExitObserver.submit` `IssueReduction@1517`(`HeldQuantity: m.position.Quantity`) → `risk.Evaluate` → `evaluateReduction` B4(if@241) `SellExceedsHoldings`.
- (iii) 한 intent 에 attempt 여러 개 가능(`attempt_no` per intent, schema.go:229) → 「PLACE · AMEND attempt 상태 무관 정확히 하나 · CONFIRMED PLACE」 를 조건에 넣음.

## 구현 적대 리뷰 impl-r1 (Manager 실행, 별도 문맥 Opus + codex, HEAD 8ad6c3e79 대 base 28680af3d)

판정 조건부 통과 — P0 없음(base 와 표 전체 덤프 차등에서 달라진 것은 차단 행과 표지 행뿐). 원문 · 탐침 `analysis/review/impl-r1/`(탐침 시험은 `.go.txt` — 저장소 빌드에 들지 않게).

| # | 등급 | 요지 | 처분(Manager) · 반영 |
|---|---|---|---|
| P1-1 | 문서 | 구현은 주문 수량 ≤ 감소분이면 조정 전 일부 체결이 있어도 인정하는데 문서는 「조정 뒤 첫 체결부터 거절」 로 단정 | 문서를 구현에 맞춤: ledger Scenario WHEN 을 「주문 수량이 감소분보다 큰」 으로 좁히고 새 Scenario 「감소분 이하의 부분 청산 주문이 조정을 사이에 둔다 → 인정」, design §2 (가) · §4, Story limits. 계약 시험 `TestA200ASmallOrderStraddlingTheConvergenceIsRecognised` |
| P2-1 | 코드 | 보유 인스턴스의 모든 거절이 판정의 원장 읽기를 돌아 무관한 거절에 새 실패 경로 | `recogniseAdjustmentCloseFill` 첫머리에서 (가)를 먼저 — 정의는 `recognisableRefusal` 한 곳(순수 판정도 그것을 부름). 구조 시험 `TestA200OtherRefusalsReadNothingBeforeTheRefusalCheck`, 변이 M17. `ProjectPosition` 무편집 |
| P2-2 | 코드 | 0 · 음수 주문 수량이 참 | (라)에 「주문 수량 > 0」, 판정 시험 「0」 · 「-1」, 변이 M18 |
| P2-3 | 주석 · 이름 | reconcile 재현 파일 머리 · `CurrentBehaviour` 이름 · 없는 `…_windows_test.go` 참조 | 머리 주석 두 파일 정정, 무관리 W4 · W5 이름을 `TestA200AnUnmanagedW4/W5FillStaysAnAttributionFailure` 로(본문 무변, 옛 이름은 revision base 번들) |
| P2-4 | spec · 공시 | MODIFIED 의 「인정한 체결 하나뿐」, provenance 「judged …」 문구 | 「인정한 체결뿐이며(체결마다 인정한다)」 로, provenance 문구는 Story limits · design §4 에 한 줄(표시 문구) |
| 확인 | — | (가)는 도달 가능한 상태에서 (나)에 함의 · a199 청산 전용 기동에서도 작동 | design §5 · 변이 원장 주석 |
