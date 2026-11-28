# a194 gstack review 1회차 (2026-11-23)

- 스킬: gstack `review`(Skill 도구, preamble 4e82fa4c839f, SESSION_ID 36613-1795443460-e179dc46). base `064f59fd9`(Manager 지정 — `origin/main` 이 아님), 범위 `internal/` 코드 diff 4파일 1,082줄(테스트 804 포함). OpenSpec 은 의도 참고.
- PR type: application. Scope Check: CLEAN — Intent: 대사 차단으로 편입되지 못한 보유·운영자 전용 차단 보고(design.md). Delivered: 새 파일 1 + judgeHoldings 2줄 + 구조체 필드 3 + 시험. Diff read: full(코드 4파일). Undeclared behavior changes: none(새 critical 의 진입 래치·모드 승격은 proposal Impact 에 공시).
- Gate scan: `GATE_SUMMARY: listed=1 eligible=0 inspected=0 unread=0 tagged={} unmatched=62` — Gate edits: none detected(1 listed — 새 시험 파일, 0 read). Banned terms: none(목록 없음). Approval classes: 확인 못 함(`APPROVAL_SOURCE_UNAVAILABLE` — origin 에 base 가 없음).
- 전문가: testing · maintainability · security · performance · simplification(5) + native adversarial(Claude 별도 컨텍스트) + red team(diff > 200). Codex 외부 패스: 이번 gstack 회차에서는 돌리지 않음 — 외부 보이스 coverage 없음(직전 구현 적대 리뷰 r1 에서 codex 0.165.0 이 둘째 보이스였음).
- 탐색 QA(Step 4.7): UI 없음 — 기능 smoke 는 선택 시험(`analysis/verify/selected-tests.txt`)과 TestA194* 로 갈음. 별도 QA 프로브 없음.

## 발견

| 출처 | 등급(리뷰어) | 요지 | 처분 |
|---|---|---|---|
| testing | — | NO FINDINGS | — |
| security | — | NO FINDINGS(시험 파일은 summary mode) — 본문·필드에 계좌·수량·가격 없음 확인 | — |
| simplification | — | NO FINDINGS — `Simplification: lean already — nothing to cut.` | — |
| performance | INFO 6 | 차단 보고(`reportOperatorOnlyBlocks`)에 주기당 Notify 상한 없음 — 블록 N 개면 전송 막힘 시 주기당 N × 약 34초 | **Manager 결정과 일치**(「차단 보고는 블록 수만큼(통상 1)」, R2-2 처분) — 상한 추가는 새 정책이라 **Manager 결정 대기** |
| maintainability | INFO 6 | 본문 「10분」 이 `blockedHoldingGrace` 와 따로 적힌 리터럴 | nit — 기록만(상수 변경 시 본문 동기화 필요) |
| adversarial #1 | Medium, no | 판정 전에 끝나는 주기(Observe·Refresh·안정화 실패)에는 두 보고 모두 침묵 | Manager 결정 Q7 로 이미 범위 밖 공시(design §6) — 기록만 |
| adversarial #2 | Low, no | pending(비영속) 계좌 영구 승격 블록도 `ReleaseOperatorOnly` 라 「운영자만 풀 수 있다」로 즉시 보고되지만 다음 깨끗한 비교에서 철회될 수 있음 | **Manager 결정 대기** — `Block` 에 pending 노출은 reconcile 패키지 편집(범위 확장) |
| adversarial #3 | Low, no | 새 critical 의 전달 실패가 계좌 전체 진입 래치·모드 승격으로 넓어짐 | proposal Impact 에 공시된 기존 성질 — 기록만 |
| adversarial #4 | Low, no | = performance 발견 | 위와 같음 |
| adversarial #5 | Low, FIXABLE | 차례 도중 남은 보유가 빠지면 다음 주기가 아무것도 시도하지 않음(약 62초 공백) | **수리함**(`e26831d5a`): 루프 전 남은 보유가 없으면 차례를 비움 + `TestA194ARoundWhoseRestLeftStartsAgainNextCycle`, 되돌림 변이 M24 |
| adversarial #6 | Low, no | 알림 꺼짐 일반 등급 래치는 편입 성공에만 지워져 해제 뒤 재차단이 침묵 | Manager 결정(꺼짐 = (포지션, 사실) 래치로 한 번, 기존 `alertUnmanaged` 와 같은 방식)과 일치 — 기록만 |
| adversarial #7 | Low | 시장 범위 블록은 key 가 겹치고 「계좌 전체」로 표시됨 | 현재 생산 경로 없음(BlockRules·blockFromReconcileState 가 만들지 않음) — 기록만 |
| adversarial #8 | Low | 원인 「미상」 갈래 본문의 단정 · 영값 Since 를 가장 이른 것으로 고름 | 기록만(D1 결정의 갈래, 영값 Since 는 생산 블록에 없음 — 미측정) |
| adversarial #9 | Low | 운영자 전용 판정이 `Block.Release` 에 걸려 있어 a195 가 Release 를 안 바꾸면 문구가 거짓이 됨 | a195 에 넘길 주의 — 기록만 |
| **red team #1** | **CRITICAL 9** | 운영자 전용 보유 보고와 차단 보고가 `tossctl engine reconcile-resolve`(엔진 정지 필요)를 안내하지만, 그 명령은 원인이 QUANTITY_MISMATCH 가 아닌 활성 블록이 하나라도 있으면 거절한다(cmd/tossctl/engine_reconcile.go:308-310 「active cause %s on %s requires cause-specific recovery evidence」 — 직접 확인). 이 change 의 동기 사고인 ATTRIBUTION_FAILED 가 바로 그 경우라, 안내를 따르면 엔진을 세워 모든 보호 포지션의 손절을 잃은 뒤 명령이 거절된다 | **Manager 결정 대기(최우선)** — 본문 문구는 Manager 결정 D6 이고, 원인별 안내로 바꾸는 것은 결정 변경. a195(귀속 실패 차단 해제)와 묶어 판단 필요 |
| red team #2 | INFO 6 | 세 본문의 「엔진 정지 필요」가 정지 = 다른 포지션 손절 없음과 안전 창(두 시장이 닫힌 KST 05:00~09:00)을 말하지 않음 | **Manager 결정 대기**(본문 문구 결정) |
| red team(비발견 메모) | — | 판정까지 가는 주기의 일시적 포지션 조회 실패가 유예 시계를 지움 | 구현 리뷰 P2-1 로 이미 잔여 공시(design §6) |

## 결과

- Pre-Landing Review: 미해결 결함 1(red team #1 — 결정 대기), 정보 다수(위 표). 수리 1(adversarial #5). 자동 수리 0. 사람에게 묻는 대신 Manager 결정 대기로 남김(지시: 새 정책·결정 변경은 구현 금지).
- Gate edits: none detected (1 listed, 0 read; patterns: RH-1,RH-3,RH-4,RH-12,RH-13,RH-14,RH-15,RH-16; unlisted idioms: none)
- 회차: 1(지시대로 1회차로 끝냄 — 수리 뒤 재리뷰는 하지 않음, 수리는 시험 + 변이 M24 로 잼).
