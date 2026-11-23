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
