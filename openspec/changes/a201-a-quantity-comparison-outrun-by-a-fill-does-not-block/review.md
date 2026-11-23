# a201 review

## 1단계(등록 · 재현 · 증거) — 2026-11-24, Opus 구현 팀메이트

- worktree `~/tossos-work/a201-impl`, 브랜치 `fix/a201-stale-quantity-comparison`(로컬 커밋만, push 없음). base `f9b818a0e`(origin/main `7fec501f1` + 등록 문서).
- 끝낸 것: 1.1 등록 · 1.2 base · 1.3 사전 증거 재확인(AST 20 바이트 동일 + 새 16) · 1.4 실제 `RunOnce` 재현(6 PASS) · 1.5 전제 · 1.6 CodeGraph 1.6.0 · 1.7 편집 표면 · 1.8 잔여 창 · 1.9 ADDED 1(strict valid) · 1.10 RunOnce pre-edit 번들(`check_analysis` rc 0) · 1.11 claim check.
- **멈춤**: claim check 에서 design 의 결정이 사실과 어긋남(C-D6 · C-D4 · C-R1 · C-D5 — `analysis/claim-check.md`). 지시대로 고치지 않고 올림. **1.12(codex 제안 동결 리뷰)는 실행하지 않음** — 곧 바뀔 design 을 리뷰하는 것이 되므로 Manager 결정 뒤 그 판본으로 돌리는 것이 맞다고 판단.
- 사실 정정만 문서에 반영(결정 무변경): design D3 (가) 근거 문장(C-D3a), §2 정상 입력 1(C-W3b), §4 첫 항목(C-S4 · C-R1 참조), proposal Why(C-P6). 모두 「(1단계 정정: …)」 표지.

### 재현(1.4) 요지

- 끼우는 방법: production 편집 없이 `Converger.Journal`(ConvergeStore) · `Tracker.Journal`(ReconcileStore) 인터페이스 필드를 감싸 정해진 지점에서 체결 커밋. 운영 리터럴과 같은 훅 · 필드.
- W2 · W3-a · tx 직전 · W3-b 모두 그 주기 QM 1 → 깨끗한 5주기 뒤 미해제 · 진입 거절, W2 뒤 재매수 편입 거절.
- 추가 발견 셋: (b) 단독 stale 은 수량 불변 fill_events 행이 필요(엔진 체결 자신은 (a)), 앞 주기 크레딧 위의 W2 도 고착(크레딧 반박), 판 중단 뒤 종목의 체결이 Observe 뒤면 고착(재판정이 못 잡음).

### 탐침(1.5 (e)(f)) 요지 — scratch 사본, 추적 파일 무편집

| 모양 | D5 끔 | D5 켬 |
|---|---|---|
| W2 · W3 넷 | 그 주기 missing 블록 · 2주기 뒤 ORDER_TERMINAL_CONFIRMED | 행 0 |
| 앞 크레딧 위 W2 | 고착 그대로 | 그 주기 ADJUSTMENT_APPLIED 해제 |
| 판 중단 뒤 종목 B | B 고착 | B 고착 |

### Manager 결정 요청

`analysis/claim-check.md` 의 「결정에 닿는 것」 네 줄과 정본 해석 셋(`analysis/spec-contradictions.md`). 질문 목록은 1단계 보고에 있음.

## Manager 결정 1~6 반영(2026-11-24)

- 결정 1~5 를 design · proposal · delta · Story 에 반영(커밋 `4c9d53138`). 결정 6 에 따라 codex 1회차를 그 판본으로 돌림.

## proposal-freeze r1 — codex(2026-11-24)

- 보이스: codex-cli 0.165.0 `codex exec -s read-only -C <worktree> </dev/null`(다른 계열, 적대적 Eng). 프롬프트 `analysis/review/proposal-freeze-r1-prompt.md`, 최종 답 `analysis/review/proposal-freeze-r1-codex.md`(전체 대화 기록은 커밋하지 않고 scratch 에 둠). 대상 HEAD `4c9d53138`, 정적 검토(실행 0).
- 판정: **「freeze 불가 — 설계 결정에 닿는 P1 2」.** P0 0.

| # | 등급 | 요지 | 닿는 결정 | 내 확인 |
|---|---|---|---|---|
| R1-1 | P1 | 조건 (다)는 **다시 읽은 로컬**을 **옛 계좌 스냅숏**과 비교함. 스냅숏 **뒤에** 실제로 체결된 엔진 매도가 로컬 읽기 뒤 · 조정 tx 전에 커밋되면 (가)(나)(다)가 모두 참인데 진짜 불일치가 남음. 예: 앞 주기 조정으로 로컬 · 계좌 20 + 크레딧 · QM 활성 → 사람이 5 매도(스냅숏 15, 로컬 읽기 20) → 스냅숏 뒤 엔진 매도 5 체결 · 커밋(실제 계좌 10, 로컬 15) → 재판정 15 = 15 로 뺌 → 깨끗해진 비교가 **앞 크레딧으로 그 종목 QM 을 `ADJUSTMENT_APPLIED` 해제**. 워터마크는 「장부 커밋이 늦었다」의 증거이지 「그 체결이 스냅숏에 이미 반영됐다」의 증거가 아님 | D3 (다) · D4 · D6 · 결정 1 | **타당하다고 봄.** 코드 경로 확인(Observe B2 if@650 → B8 if@675 → 해제, B29 if@841 반박 안 됨). 결과는 해제된 그 주기 동안 실제 5주 차이가 있는데 진입이 열림 — 다음 주기 비교(로컬 15 · 계좌 10)가 새 QM 을 세우므로 「한 주기의 잘못된 해제」. 지금 코드(a201 전)는 이 입력에서 차단을 유지함. 시험으로 재현하지는 않음 |
| R1-2 | P1 | D5 의 「다시 읽은 로컬에서 더는 추적 아님」은 정정 lineage 에서 틀림: `LocalStateFromJournal` 은 부모 · 자식을 같은 현재 신원 키로 넣고 뒤의 것이 덮어써 `OriginalOrderID` 가 자식이 됨 → 부모의 missing-order 를 지우는데 살아 있는 자식이 실제로 missing. 또 `fail_closed` 종결도 추적에서 빠짐 — 「추적 부재 = 깨끗한 종결」이 아님. 1단계 탐침은 `TrackedFillOrders` 를 직접 읽어 이 반례를 검증하지 않았음 | D5 | **타당하다고 봄.** compare.go:225-235(map 키 = 현재 신원, 값 덮어씀) 확인. 내가 구현하려던 판정(원 번호 또는 현재 번호 일치)도 부모 번호 · 부모의 현재 번호가 둘 다 사라지면 같은 구멍 |
| R1-3 | P2 | §2 「차단이 늦어지는 상한 한 주기」 불성립 — R1-1 을 매 주기 되풀이(스냅숏 뒤 부분 체결이 주기마다 커밋)하면 실제 1주 차이가 여러 주기 차단 없이 지속. 60초도 주기 전체 시간의 상한이 아님(Run 은 작업 뒤 period 대기) | §2 위험 수용 근거 | 타당. R1-1 과 같은 원인 |
| 기타 | — | 조건 (나): 부분 커밋 · Applied=false · 알림 실패에서 `Converged` 유지, ApplyPositionAdjustment 는 Commit 뒤 오류를 내는 후속 연산 없음(position_adjustments.go:366). 오류 처리: Refresh 실패는 재판정 전에 반환, `cycle.Err` 는 R1-1 의 안전장치가 아님. a192: 결속된 같은 종목은 뒤 체결 워터마크 검사로 해제 거절(order_terminal_release.go:258) — 크레딧 경로는 이 검사를 안 거침. D2 의 id = 커밋 순서 반례 없음 | — | — |

- 처분: **반영하지 않고 멈춤**(Manager 지시 — 결정급 P1). freeze 안 함, tasks 1.12 · 1.13 미체크, 구현(2.x) 미착수, origin/main merge 미실행.
