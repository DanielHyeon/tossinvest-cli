# a200 1.3 — 사전 증거 재확인

- 사전 증거: Manager scratch `a200-evidence/evidence.md`(HEAD `9738a0c32`, 시험 0 · 일부 코드 추론).
- 재확인 위치: worktree `/home/daniel/tossos-work/a200-impl`, 브랜치 `fix/a200-engine-exit-fill-recognised`,
  HEAD `884be2a74`(= origin/main `9738a0c32` + 이 change 등록 커밋, Go 변경 0). a198 은 2026-11-24 마지막 fetch 까지 origin/main 에 없음.
- 방법: `analysis/cited-ast/extract.py`(커밋됨) 가 함수 44 개를 `go run ./tools/logic-map` 로 다시 뽑아 `analysis/cited-ast/*.ast.json` 에 쓰고
  사전 증거의 같은 이름 파일과 JSON 전체를 비교. 원문 표 `analysis/cited-ast/extract-result.md`.

## 결과

| 항목 | 결과 |
|---|---|
| 사전 증거 AST 34 개 | **34 개 전부 JSON 전체 동일**(source_sha256 · 줄 · 분기 · 호출 · 반환). 줄 번호 · `B<n>` 좌표 변화 0 |
| 1단계가 새로 뽑은 AST 10 개 | `currentInstance` · `exitContextForFill` · `computeTradeOutcome` · `roundTripLegs` · `adoptedRoundTripLegs` · `markExitCompletedTx` · `Journal.AttachExitIntent` · `Journal.ResolveExitProposal` · `clearExitProposalTx` · `ApplyTx.ResolvePending` |
| 달라진 곳 | AST 좌표는 없음. **서술**이 재현 · 추가 조사로 바뀐 곳은 아래 표 |

## 사전 증거 서술 중 1단계가 정정 · 보강한 곳

| 사전 증거 | 1단계 사실 | 근거 |
|---|---|---|
| Q1-1 W2 · W3 「코드상 QM 고착을 만든다 … 재현 시험은 돌리지 않음」 | **재현됨** — 일치 비교 5 주기 뒤에도 `QUANTITY_MISMATCH` 활성, 매 주기 `AwaitingAdjustment` | `analysis/repro/cycles.txt` W2 · W3, `analysis/repro/out-of-scope.md` |
| Q1-1 W4 · W5 · W6 표 | 셋 다 서술과 같게 재현(W4 AF 고착 · W5 `durable scope is already owned by cause ATTRIBUTION_FAILED` · W6 QM 이 다음 일치에서 `ADJUSTMENT_APPLIED`) | `analysis/repro/cycles.txt` |
| Q1-2 원장 표 | 재현과 일치. 추가: `exit_events` 는 `OPENED · BASELINE_BREACH_EXIT · ADJUSTMENT_CLOSED · PROPOSAL_FILLED` 순서(운영 집계 「ADJUSTMENT_CLOSED 뒤 PROPOSAL_FILLED 2」 와 같은 모양), 조정 `created_at` = 체결 `committed_at` 같은 초 | `analysis/repro/w4.txt` 첫 시험 |
| Q4 표 UNATTRIBUTED_SELL (b)·Q2(2) 「최신 인스턴스」 | **체결 투영과 Exit 훅 둘 다 주문이 아니라 (계좌 · 시장 · 종목)의 최신 인스턴스를 읽음**(`currentInstance` 270-301, `exitContextForFill` 939-942 SQL `ORDER BY instance_seq DESC LIMIT 1`). 재매수가 새 인스턴스로 접힌 뒤 옛 손절의 체결은 **새 인스턴스에 닿아** `SELL_ON_CLOSED` 가 아니라 OVERSELL(X50 → QM) 이 되거나, 부분 체결이면 **차단 없이 새 외부 보유를 줄임**(5 → 2, CLOSING) | `analysis/repro/w4.txt` 마지막 둘 앞 두 시험 |
| Q2(5) `runApplyHooks` 생산 호출 「RecordFill 1」 | 생산 호출 자리 2: `fills.go:551`(RecordFill) · `strategy_dispatch_runtime.go:1348`(`backfillConfirmedStrategyFillTx` — `side='BUY'` 만, 이 change 무관) | CodeGraph 1.6.0 callers + grep |
| Q2(4) 편입 성과 「`post_adoption_disposal` 거절」(추론) | **확인** — W4 원장에서 동결 탐침(`computeTradeOutcome`, 체결 tx 와 같은 nil 분봉)이 `post_adoption_disposal` | `analysis/repro/w4.txt` 편입 시험 |
| Q2(4) 엔진 진입 성과 | W4 원장 탐침이 **W0 동결 값과 같은 성과**를 냄(같은 초 — `closed_at`·`held_seconds` 가 조정 시각을 따르므로 다른 초면 그 둘은 다름, `computeTradeOutcome` 184 · 261) | 같은 파일 첫 두 시험 |

## 운영 집계 요약(Manager 제공 — `analysis/ops-aggregates.md`, 건수 · 종류 · 날짜만)

- `ATTRIBUTION_FAILED` 이력 1(2026-11-20, `X59`/`SELL_ON_CLOSED`, 미해제), 미해제 `QUANTITY_MISMATCH` 0.
- 수렴 조정으로 닫힌 포지션 34 중 닫힌 뒤 `PROPOSAL_FILLED` 가 온 것 2(11-16 · 11-20), 둘 다 성과 0.
- 1단계 재현과의 대조: 11-20 = W4 또는 W5 모양, 11-16 = W6 모양(서술과 같음 — 확정 질의는 하지 않음).
