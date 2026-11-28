너는 TossOS(실제 돈을 다루는 Go 자동매매) 저장소의 적대적 Eng 리뷰어다. 읽기 전용이다 — 파일을 만들거나 수정하지 말고, 빌드 · 시험 · 네트워크 · tossctl · docker 명령을 실행하지 말 것(읽기 · grep · git show/log 만).

이것은 proposal-freeze 리뷰 **v3 1회차**다(회차는 v3 부터 새로 센다). 대상: `openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/` 의 proposal.md · design.md(v3) · specs/reconciliation/spec.md(ADDED 1 + MODIFIED 3) · tasks.md, 증거 analysis/{claim-check-v3.md, edit-surface-v3.md, v3-evidence/evidence.md, v3-evidence/probe-*.txt, v3-probe/probe.diff, v3-probe/run.txt, repro/README.md, modified-delta/generate.py}. 앞 설계 v1 · v2 는 기각됐다(analysis/review/proposal-freeze-r1-codex.md · proposal-freeze-r2-codex.md, design 「기각한 대안」) — 그 반례(R1-1 · R1-2 · R2-1 · R2-2)가 v3 에서 다시 열리는지도 보라.

v3 요지: 차단은 지금과 똑같이 세운다. RunOnce 가 로컬 읽기 전에 전역 최대 체결 id 를 읽고, 그 주기 Observe 가 성공해 돌아온 뒤 「수량 불일치 종목 ∧ 종목 워터마크 > 그 id ∧ 그 주기 수렴이 조정을 반영하지 않음(ConvergeReport.Converged 에 없음) ∧ hasBlockFor(QM ∧ adjusted_reconcile)」 인 종목에 **체결 credit** 을 diff.AsOf 로 준다. 체결 credit 은 **조정 credit 과 같은 맵 Tracker.adjusted 에 같은 값으로** 들어가 기존 사용(Observe B8 if@675 · B10 if@689) · 반박(call@716 · 끝 루프 B29 if@841) · 소멸(Restore · Resolve 의 adjusted=nil) 분기를 무편집으로 공유하고, 해제 원인만 곁 맵 Tracker.filled(같은 as-of 면 체결)로 골라 Observe 가 Cleared 직전에 Block 에 표지 → persist 가 Cause=FILL_APPLIED 로 ReleaseReconcile. 새 원인은 ValidReconcileReleaseCause 에 추가하고, 위험 버킷 broker-zero 허용 목록(risk_bucket_owner.go:280 · :677)은 무편집이라 거절. 편집 기존 함수: RunOnce · Observe · persist · ValidReconcileReleaseCause.

읽어야 할 소스(최소): internal/app/engine/reconcileloop.go(RunOnce), internal/reconcile/mismatch.go(Observe, persist, AdjustmentApplied, creditUsableBy, refuteUsableAdjustmentCredits, hasBlockFor, symbolsInDispute, Restore, Resolve, Refresh, Block), internal/reconcile/order_terminal_release.go, internal/reconcile/converge.go, internal/journal/reconcile_states.go, internal/journal/position_adjustments.go(FillWatermark), internal/journal/reconcile_order_terminal_release.go(MaxFillEventID), internal/journal/risk_bucket_owner.go(:260-310 · :630-700), internal/app/engine/reconcile_block_report.go, openspec/specs/reconciliation/spec.md.

찾을 것(각 발견에 P0/P1/P2/P3, 근거 file:line, 재현 가능한 입력 모양, 설계 결정에 닿는지):
1. design V8 의 공시(외부 분쟁 상쇄 · 다른 계좌 체결 · 수렴 거절 종목) **밖에서** 틀린 해제가 되는 입력.
2. credit 수명 규칙의 구멍 — 반박을 피해 오래 사는 체결 credit(예: 그 종목이 비교에서 사라졌다가 나타남, 외부 주문만 있는 비교, 계좌 범위 블록과 종목 블록, pending 블록), 맵 공유가 조정 credit 의 기존 해제 결과를 바꾸는 입력(덮어쓰기 · 원인 오표지 포함).
3. Observe · persist 편집이 기존 두 해제 경로(조정 credit · a192 종결 증명)의 결과를 바꾸는 입력, 그리고 V5(해제 순서) 를 맵 공유로 바꾼 것이 a192 결속 행과 상호작용하는 입력(claim-check (q7)).
4. 그 주기의 Observe 가 정말 무변인가(부여가 Observe 뒤라는 사실, Observe 오류 주기, 같은 as-of 규칙).
5. ADDED · MODIFIED 3 이 정본(:53 · :77 · :303 · :347 · :379-451)과 모순되거나 design 에 없는 규칙을 더하는가, 시나리오의 결함.

출력: 한국어. 맨 위에 한 줄 판정(freeze 가능 / 조건부 / 불가), 그다음 등급별 발견, 마지막에 「설계 결정에 닿는 P0/P1」 이 있는지 예/아니오로 명시.
