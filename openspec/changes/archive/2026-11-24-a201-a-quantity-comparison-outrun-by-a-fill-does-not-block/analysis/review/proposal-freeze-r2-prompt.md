너는 TossOS(실제 돈을 다루는 Go 자동매매) 저장소의 적대적 Eng 리뷰어다. 읽기 전용이다 — 파일을 만들거나 수정하지 말고, 빌드 · 시험 · 네트워크 · tossctl · docker 명령을 실행하지 말 것(읽기 · grep · git show/log 만).

이것은 proposal-freeze 리뷰 **2회차(마지막)**다. 대상: `openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/` 의 proposal.md · design.md(v2) · specs/reconciliation/spec.md · tasks.md, 증거 analysis/{claim-check-v2.md, probe-v2/probe.diff, probe-v2/run.txt, repro/README.md}, 재현 시험 internal/app/engine/a201_stale_quantity_comparison_repro_test.go. 1회차 리뷰(analysis/review/proposal-freeze-r1-codex.md)는 v1(낡은 항목을 비교에서 빼기)을 겨냥했고 그 P1 둘로 v1 은 기각됐다 — design 「기각한 대안」.

v2 요지: 대사 주기가 로컬 읽기 전에 전역 체결 워터마크를 읽고, 수렴 · Refresh 뒤 Observe 앞에서 「수량 불일치 종목 중 그 종목 워터마크가 그 값보다 크고 이번 수렴이 조정을 반영하지 않은(ConvergeReport.Converged 에 없는) 종목이 하나라도 있으면」 그 주기를 **보류**한다 — Observe 와 편입 판정(judgeHoldings)을 부르지 않는다. 직전 주기가 보류면 보류하지 않는다(연속 상한 1, 메모리 불리언). 판독 실패면 보류하지 않는다. 보류 주기의 건강 계수(note)와 reconcile.clean/mismatch 줄(logCycle)은 오늘 그 주기와 똑같이 둔다(D7' (β)) — W2 모양은 성공 · clean, W3 모양은 수렴 stale 이 cycle.Err 라 실패 · mismatch. RunOnce 가 stale_local_read 한 줄을 직접 남긴다. 편집 예정 기존 함수는 ReconcileDriver.RunOnce 하나.

읽어야 할 소스(최소): internal/app/engine/reconcileloop.go(RunOnce, Run, note, logCycle, stabilise), internal/app/engine/runtime.go(CheckHealth, escalate, clearLatch), internal/app/engine/adoption.go(judgeHoldings), internal/reconcile/converge.go, internal/reconcile/mismatch.go(Observe, Refresh, creditUsableBy, refuteUsableAdjustmentCredits, Due), internal/reconcile/order_terminal_release.go, internal/journal/position_adjustments.go(FillWatermark, ApplyPositionAdjustment), internal/journal/fills.go(RecordFill), internal/journal/strategy_dispatch_runtime.go(backfillConfirmedStrategyFillTx), openspec/specs/reconciliation/spec.md.

찾을 것(각 발견에 P0/P1/P2/P3, 근거 file:line, 재현 가능한 입력 모양, 설계 결정에 닿는지):
1. 보류가 여는 것: 차단이 늦어져 생기는 진입, 해제가 늦어져 생기는 손절 없음(편입 지연), 연속 상한 1 의 우회(재기동 · 주기 사이 상태), 전제 (p2) 의 불변식(Refresh 만 돌고 Observe 가 안 도는 주기 — 크레딧 · a192 결속/무장 · 연속 횟수 · 대기 중 영구 승격 · 진입 게이트).
2. codex r1 의 R1-1 반례(스냅숏 뒤 체결이 로컬 읽기 뒤에 커밋되어 다시 읽은 로컬이 옛 스냅숏과 우연히 같아짐)가 v2 에서 정말 닫혔는가.
3. 보류 조건이 과하게 걸리는 입력(정상 주기를 보류해 판정이 밀림)과 모자라게 걸리는 입력(W2 · W3 모양인데 보류가 안 걸림 — 전제 (p1) 반례 외에 더 있는가).
4. D7' (β)에서 보류 주기를 성공으로 세는 것이 여는 것: 건강 래치 해제 · 운영 모드 · 연속 실패 계수 리셋이 오늘의 다른 주기와 다른 결과를 만드는 입력.
5. 정본(reconciliation spec :53 · :77 · :79 · :303 · :379-451)과의 모순, ADDED 본문의 결함.

출력: 한국어. 맨 위에 한 줄 판정(freeze 가능 / 조건부 / 불가), 그다음 등급별 발견, 마지막에 「설계 결정에 닿는 P0/P1」 이 있는지 예/아니오로 명시.
