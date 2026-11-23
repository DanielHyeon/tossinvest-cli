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

## gstack review r1 (팀메이트 실행, 한 번 — 상한, HEAD d939e8795 대 base 28680af3d)

판정: **blocker 0**(지금 입력으로 시험 실패 · 생산 고장 · 데이터 손실을 내는 발견 없음). 코드 편집 0 · 새 가드 0 — 코드를 바꿔야 하는 발견은 아래 「Manager 에게 묻는 것」. 원문 `analysis/review/gstack-r1/`(codex 프롬프트 · 출력, state matrix).

실행한 갈래(침묵 생략 없음): gate scan(`listed=9 eligible=6 inspected=6 unread=0 tagged={RH-1:2}` — RH-1 둘은 뒤집힌 시험 셋 삭제 · 무관리 W4/W5 이름 바꿈, 결정이 diff 에 보여 finding 아님) · banned terms(`none`) · specialist 6(testing 3 · maintainability 1 · security 0 · performance 0 · simplification 0 · state-machines 2, 별도 문맥, 읽기 전용) · red-team(diff > 200 줄) 2 + 인용 줄 1 · Claude adversarial(시험 파일 요약 모드) 막는 것 없음 · codex 외부 목소리(프롬프트 파일을 stdin 으로, `read-only` sandbox) `VERDICT: FINDINGS` High 1 · Medium 2 · smoke(`-run TestA200` 두 패키지 ok, 아래 검증). state matrix check rc=1 은 미시험 칸 4 를 INFORMATIONAL 로 적은 탓(도구는 CRITICAL 요구) — 그 넷은 a200 이 행동을 바꾸지 않는 칸.

**상태기계 — CLOSED + 매도 체결의 세 결과 × 네 조건**(판정은 `ProjectPosition` :104 에서 `enterReconcileInTx` 의 F10 활성 차단 확인 :406-411 보다 **앞**):

| 조건 | 활성 차단 없음 | 활성 차단 있음 |
|---|---|---|
| 넷 다 성립 | 인정(표지 1 · 차단 0) | 인정 — 표지 1, 기존 차단은 건드리지 않음(W6, 이번에 단언 추가) |
| (가)만 실패 | 도달 불가(CLOSED 의 매도 거절은 X58~X60 `SELL_ON_CLOSED` 뿐, 판정 시험만) | 도달 불가 |
| (나)만 실패 | `ATTRIBUTION_FAILED` | 새 행 없음(F10, a200 무변 · 무시험) |
| (다)만 실패 | `ATTRIBUTION_FAILED` | 새 행 없음(무관리 W6 시험) |
| (라)만 실패 | `ATTRIBUTION_FAILED` | 새 행 없음(a200 무변 · 무시험) |

표지 행은 어느 exit 상태기계 전이에도 들지 않음: `exit_events` 독자 전수 — `ExitEvents`(생산 호출 0) · `AccountExitEvents` → 콘솔 표시(원문 action 표시, 창 한 칸 차지) · provenance 표시(「judged …」, design §4) · `proposed_intent_id`/`decision_id`/`action = COMPLETED` 로 거르는 독자(account_views · a166 · a087 · trade_outcomes · backfill · exit_state:494 · provenance exit_intent)는 NULL/다른 값이라 제외. `ApplyExitFill` 은 `exit_states` 만 읽음(완료 플래그는 ADJUSTMENT_CLOSED 가 이미 세움 → COMPLETED · 성과 없음). `action` 에 CHECK · 트리거 없음.

| # | 출처 | 요지 | 4분류 | 처분 |
|---|---|---|---|---|
| G1 | adversarial A1 + codex C1(두 출처 독립 일치) | (라)는 `intents.quantity` 만 감소분과 견줌 — 브로커가 다른 수량으로 접수 · 체결(어댑터 · 브로커 결함)하면 감소분을 넘는 체결도 인정. 제안: 관측의 `OrderedQuantity` · `CumulativeQuantity` ≤ 감소분(빈 값 · 파싱 실패면 거절) | should(전제 (i) — intent 행 불변 · 브로커 호출은 같은 intent 복사본 — 이 깨져야 도달) | **Manager 대기**(새 가드). design §4 공시 |
| G2 | red-team R1 | 인정이 「같은 주문의 늦은 체결 조각 사이」 진입 차단을 없앰 → 수량 불일치가 풀린 뒤 재매수가 끼면 늦은 조각이 새 인스턴스를 줄임. 오늘은 첫 늦은 체결의 차단이 진입을 닫아 둠 | should(재매수 + 다조각 지연이 겹쳐야, 다음 비교가 계좌 값으로 다시 맞춤) | **Manager 대기**(가드 「인정된 주문 종결 전 진입 닫기」 또는 「종결 체결만 인정」). design §2 「범위 밖 사실」 을 「a200 이 닿게 만듦」 으로 정정 |
| G3 | codex C2 | `Gateway.Amend` 는 새 intent id — 원 intent 에 AMEND attempt 가 없어 (라) 우회 | nit(생산 호출 0 확인 — `ops/write_operations.go` 는 trading.Service 직접) | design §4 공시(a167 착지 때 계보 · `target_order_id` 로 재검토) |
| G4 | codex C3 | `created_at` 벽시계 역행이면 닫는 조정이 마지막이 아니게 정렬 → 거절 | nit(보수 방향 = 오늘 동작) | design §4 공시. 정렬 자체는 시험 T1 · M19 로 고정 |
| G5 | adversarial A3 | (다)가 `PendingAction` 빈 값은 안 봄 | nit(쓰는 쪽 불변식상 도달 불가) | **Manager 대기**(코드), 지금은 기록만 |
| G6 | adversarial A2 | 표지에 주문 · intent · 수량 없음 | nit(D3 결정 — intent 키 소비자 회피) | design §4 공시, 바꾸려면 Manager |
| G7 | red-team R2 | 가정 ② 가 틀리면 남은 보유가 EXTERNAL 로 접혀 손절 · a194 보고 없음 | 문서 | design §4 에 구체 대가로 |
| G8 | adversarial A4 · testing T3 | 판정 읽기 DB 오류면 체결 tx 롤백, 오류 경로 무시험 | 문서 | design §4(오류 주입 seam 없음) |
| T1 | testing | 원장 시험이 전부 조정 행 하나 — 정렬 무시험 | 시험 보강 | `TestA200AFillAfterATwoStepConvergenceIsRecognised`(31→28 · 같은 초 28→0) + 변이 M19 CAUGHT |
| T2 | testing | 재기동 시험이 `cleanCycle` 오류를 단언 안 함 | 시험 보강 | `t.Fatalf` 추가, PASS |
| S1 | state-machines | W6 에서 표지가 남는 칸 미단언 | 시험 보강 | W6 관리 시험에 표지 1 · 활성은 수량 불일치 하나뿐 단언, PASS |
| S2 | state-machines | (나)·(라) × 활성 차단 칸 무시험 | nit | a200 무변 경로(같은 F10) — 시험 안 넣음 |
| M1 | maintainability | 재현 시험 doc 주석이 기각된 「누적 합」 을 말함 | 주석 | 「주문 수량」 으로(본문 무변, W0 번들 ast 재생성) |
| R3 | red-team | design `exitloop.go:1839` → 지금 1850 | 문서 | 정정 |

변이: 3 회차의 M18 은 빌드 실패(가짜 CAUGHT) → 다시 써서 4 회차 **19 / 19 CAUGHT**(`analysis/mutation/README.md`). 검증(HEAD 25d9aea2d): 선택 시험 4 패키지 ok · `-race` A200 ok · `make lint` rc 0 · strict valid · check_analysis evidence complete · tracker `--check` current — `analysis/verify/selected-tests-gstack-r1.txt`.

### Manager 에게 묻는 것 (gstack r1)

1. G1: (라)에 관측 수량 상한(`OrderedQuantity` · `CumulativeQuantity` ≤ 감소분)을 더할지 — 더 엄격해지기만 하고 정상 입력은 그대로 통과(두 리뷰어 판단, 미측정).
2. G2: 인정된 주문이 종결될 때까지 그 종목 진입을 닫을지, 종결 체결만 인정할지, 공시로 둘지.
3. G5: (다)에 `PendingAction != ""` 를 더할지(도달 불가 방어).

### gstack r1 처분 — G1 수리 · G2 · G5 공시(Manager 2026-11-24)

- **G1 수리**: (라)에 이 체결 관측이 들고 온 `CumulativeQuantity` ≤ 감소분, `OrderedQuantity` 가 비어 있지 않으면 그것도 ≤ 감소분(못 읽거나 빈 누적 · 못 읽는 주문 수량 → 아님). 첫 시도는 멈춤 조건(판정에 `fill` 이 닿지 않음)으로 멈췄고, Manager 가 (A) — 이미 편집한 `ProjectPosition` 호출 줄 :104 에 인자 `fill` 하나 — 를 승인. 원장 읽기 없음(구조 시험 `TestA200TheRecognitionIsCalledOnceAndReadsNoFillLedger` 는 표 이름 철자와 호출 자리만 보므로 인자 받기는 읽기로 세지 않음, PASS). `ProjectPosition` ast 재생성: 분기 6 · 호출 · 반환 · 부수효과가 앞 판과 같음.
  - `AppliedFill` 생산 자리 둘: fills.go:478-497(`OrderedQuantity` = 관측 `Quantity`, `orZero` 없음 → 빈 문자열 가능 → 누적만), strategy_dispatch_runtime.go:1319(매수 backfill → (가)에서 탈락). design D2 · D2a.
  - 시험: 판정 경계 8 사례(누적 = 감소분 인정 · 누적 < 감소분 인정 · 누적 > 감소분이고 intent ≤ 감소분 거절 · 누적 비었음 · 못 읽음 거절 · 관측 주문 수량 > 감소분 거절 · 실리지 않음은 누적만 · 못 읽음 거절), 원장 `TestA200ABrokerFillLargerThanTheIntentStaysAnAttributionFailure`(intent 31 · 조정 31→0 · 브로커 40/40 → `ATTRIBUTION_FAILED` X59 활성 1 · 표지 0). 기존 A200 시험 전부 PASS — 거절된 정상 입력 없음.
  - spec delta(ledger ADDED 조건 문장 · 거절 Scenario) 재생성, strict valid.
- **G2 공시**: design §4 · Story limits 에 「오늘은 첫 늦은 조각이 세우는 차단이 그 종목의 엔진 진입과 편입을 막았다 — 사람의 재매수 자체와 그 뒤 조각이 새 보유를 줄이는 것은 오늘도 못 막는다」. 진입 닫기 가드는 넣지 않음.
- **G5 공시**: design §4 한 줄(도달 불가).
- G3 · G4 · G6 · G7 · G8 공시 승인. 「(라) 실패 × 활성 차단 → 새 행 없음」 시험 `TestA200AnUnrecognisedW6FillWritesNoNewRow`(PASS — 같은 모양의 W6 관리 시험은 표지 1, 이 시험은 PLACE 둘이라 표지 0 · 활성 수량 불일치 하나뿐). (나) 쪽은 같은 F10 경로라 시험 없음. 판정 읽기 DB 오류 경로는 미실행 범위.
- 변이 5 회차(`4f6337a1e`) **22 / 22 CAUGHT**, 빌드 실패 0. 검증(HEAD `4f6337a1e`): 선택 시험 4 패키지 ok · `-race` A200 ok · `make lint` rc 0 · strict valid · check_analysis evidence complete · tracker current — `analysis/verify/selected-tests-g1.txt`.

## Manager 검증·종결 (2026-11-24)

### Manager 검증 (Fable — tasks 3.4)

- HEAD `687ee4d70`, GOFLAGS 비움. journal `-run 'A200|Project|Position|Adjustment|ExitFill|Guarded|A130|A066'` ok 65.4s · reconcile `-run 'A200|A192|A195|Converge|Tracker|Observe'` ok 11.7s · app/engine `-run 'A200|A194|Adoption|Reconcile'` ok 17.3s · accountbudgetruntime ok 25.4s · A200 시험 32 PASS / FAIL 0 · vet rc 0 · `openspec validate --strict` valid · `check_analysis` 「evidence complete」.
- 판정: **통과 2026-11-24**. 종결 진행.

### 사용자 공시 (Manager 결정 — 완료 보고에 그대로 싣는다)

- (가) 이 change 는 「인정」(차단을 만들지 않음)만 하고 성과 기록(F15)은 하지 않는다 — 운영의 경합 두 건(11-16 · 11-20)은 성과가 빈 채다.
- (나) 여전히 차단되는 정상 입력: 주문 수량이 조정 감소분보다 큰 주문(전량 청산 주문의 일부가 조정보다 먼저 반영된 경우) · 정정된 청산 주문 · flatten/운영자 매도.
- (다) 인과는 미증명 — 근거는 측정 M29 · M33 이고, 사람의 앱 매도 · 부분 체결 잔량 · 보유 조회 지연은 미측정 가정이다.
- (라) 재매수로 새 인스턴스가 생긴 뒤의 옛 체결은 새 인스턴스에 닿는다(별도 결함 후보, out-of-scope S1).
- (마) 체결이 먼저인 창의 수량 불일치 고착(W2 · W3)은 재현됐고 a201 이 보류 상태다.
- (바) γ(자동 해제)는 사용자 결정으로 보류.
- (사) 판정의 원장 읽기 실패 경로는 시험이 없다.

### 종결 (종결 에이전트 Opus — 사용자 요청 「archive(완료 게이트 포함)」)

- 착지 먼저: `687ee4d70`(base `28680af3d` 위 그대로)에서 `--record-landing` → `landed-commit.txt` = `4f6337a1e`(도구 계산 — 그 뒤 커밋은 `.go` 무편집), 단독 커밋 `efa1ec403`. 그 뒤 `check_analysis --change` 「evidence complete」 required 7(HEAD `efa1ec403`).
- main 따라가기: `git fetch origin` 뒤 `origin/main` = `7fec501f1`(a198 archive) — base 재고정 때 이미 merge 한 커밋이라 `git merge origin/main` 은 「Already up to date」, merge 커밋 없음.
- merge 뒤 정본 대조: `analysis/modified-delta/build_delta.py` 재실행 rc 0 — position-ledger 「포지션 상태기계」 치환 1 · 역치환 원문 일치 · Scenario 2, ADDED 본문 453자 · reconciliation ADDED 369자, 산출물 git 무변(멱등). base 이후 정본 `openspec/specs` 무변. `openspec validate --strict` valid.
- merge 뒤 선택 시험(`analysis/close/merge-selected-tests.txt`, HEAD `efa1ec403`, GOFLAGS 비움, `-count=1 -timeout 30m`, 순차): journal 위 묶음 ok 91.9s · reconcile ok 23.8s · app/engine ok 31.9s · accountbudgetruntime ok 29.3s · `--- PASS` 550 / `--- FAIL` 0(최상위 TestA200 32 PASS, SKIP 1 = `TestA130JournalCopyCensus` — 사람이 지정한 읽기 전용 사본이 있을 때만 도는 시험) · `go vet ./internal/journal ./internal/reconcile` rc 0 · `make lint` rc 0. 부하 load≈7~8(다른 세션 시험)로 Manager 수치보다 느림.
- tasks: 3.4 체크(위 Manager 검증), 4.1 **선체크(gate ② 선행 조건)** — 게이트·archive·Story·push 는 이 체크 뒤 수행. 「5. 종결 뒤(사람 승인)」 는 원래 체크박스 밖(무변). 운영 반영 안 함 — 이 change 는 미배포로 끝난다.
