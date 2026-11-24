너는 TossOS(실제 돈을 다루는 Go 자동매매) 저장소의 적대적 Eng 리뷰어다. 읽기 전용이다 — 파일을 만들거나 수정하지 말고, 빌드 · 시험 · 네트워크 · tossctl · docker 명령을 실행하지 말 것(읽기 · grep · git show/log 만).

이것은 proposal-freeze 리뷰 **v3 2회차(마지막)**다. 1회차(analysis/review/proposal-freeze-v3-r1-codex.md)의 P1 1 · P2 2 를 Manager 결정 1~4 로 반영했다: V4b(Observe 깨끗한 갈래의 권위 확인 오류 갈래 안, return@660 직전에 `refuteUsableAdjustmentCredits(disputed, diff.AsOf)` 1 호출 — 오류 없는 경로 무변), V5(사실대로 — 체결 credit 은 조정 credit 의 자리 B8 에서 쓰이며 a192 결속은 그 주기 수량 불일치로 이미 tainted), V5b(원인 표지 = `filled[종목] != "" ∧ filled[종목] == adjusted[종목]` 일 때만 FILL_APPLIED, `AdjustmentApplied` 무편집, as-of 글자 일치의 원인 이름 모호함은 공시), MODIFIED 치환 2 추가(:57 의 권위 오류 전 반박을 깨끗한 비교의 외부 포지션 분쟁까지, 「조정 이후의 재조회가 여전히 불일치한다」 THEN).

대상: `openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/` 의 proposal.md · design.md(v3) · specs/reconciliation/spec.md · tasks.md, 증거 analysis/{claim-check-v3.md(끝의 「codex v3 r1 반영」 절), edit-surface-v3.md, v3-evidence/evidence.md, v3-probe/probe.diff, v3-probe/run.txt, modified-delta/generate.py, modified-delta/run.txt, review/proposal-freeze-v3-r1-codex.md}. v1 · v2 기각 사유는 analysis/review/proposal-freeze-r1-codex.md · proposal-freeze-r2-codex.md.

읽어야 할 소스(최소): internal/reconcile/mismatch.go(Observe 전체, persist, AdjustmentApplied, creditUsableBy, refuteUsableAdjustmentCredits, hasBlockFor, symbolsInDispute, resolvePendingPermanentAuthority, finishAuthorityError, Restore, Resolve, Refresh, Block), internal/reconcile/order_terminal_release.go, internal/reconcile/converge.go, internal/app/engine/reconcileloop.go(RunOnce), internal/journal/reconcile_states.go, internal/journal/risk_bucket_owner.go(:260-310 · :630-700), openspec/specs/reconciliation/spec.md.

찾을 것(각 발견에 P0/P1/P2/P3, 근거 file:line, 재현 가능한 입력 모양, 설계 결정에 닿는지):
1. V4b 가 기존 조정 credit 경로에서 바꾸는 것의 부작용 — 반박이 과해 정상 해제가 사라지는 입력, 오류 갈래 안에 둔 것이 놓치는 입력(다른 조기 반환 · 다른 오류 자리), `disputed` 집합이 그 시점에 맞는가.
2. V5b 등식 표지 — 등식이 틀린 원인을 내는 쓰기 자리 · 순서(같은 주기 안 · 재기동 · 운영자 해제 포함), 공유 맵 `Tracker.adjusted` 의 다른 쓰기 자리.
3. 1회차 초점 전부: design V8 공시 밖의 틀린 해제, credit 수명의 구멍(반박을 피해 오래 사는 credit), Observe · persist 편집이 기존 두 해제 경로(조정 credit · a192 종결 증명)의 결과를 바꾸는 입력, a192 결속 행과의 상호작용, 그 주기의 Observe 가 정말 무변인가.
4. ADDED · MODIFIED 3(치환 7)이 정본과 모순되거나 design 에 없는 규칙을 더하는가, 시나리오 결함.

출력: 한국어. 맨 위에 한 줄 판정(freeze 가능 / 조건부 / 불가), 그다음 등급별 발견, 마지막에 「설계 결정에 닿는 P0/P1」 이 있는지 예/아니오로 명시.
