# a136 review

## 사람 결정 기록 (2026-10-14)

Manager 가 전달한 사용자 최종 결정 — 원문은 design.md 「사람 결정」. 요약: D1=D(KST 08:00 경계, `clock.AccountTradingDay` 한 곳 — a137 공유) ·
D2=c(예약 원장 집계, 0.7 실측으로 체결 없음 증명 가능 확인) · D4=≥ · D5=USD base 무값 기동 거부 · D9=b(수동 매수 합산). 2차 결정(같은 날): 「별도 change 로 나눌 필요 없음」 — 수동 매수 합산과 콘솔 칸 모두 a136 안. FX 조항은 a137 문구 그대로 a136 델타로.

## proposal-freeze 리뷰 (2026-10-14)

- 보이스: 분리된 컨텍스트의 적대 리뷰어 1(code-reviewer, 적대 Eng + CEO 범위 + 운영 DX 관점), 읽기 전용, HEAD f504cc52 + 워킹트리 WIP.
- 판정: **APPROVE-WITH-FIXES** — P1 3 · P2 10 · P3 6.
- 수용 · 수리: P1-1(Z3 범위 — design D10, 변이 M3 · M4), P1-2(통화 전환 지움 — `SaveEngineGateLimits`, 변이 M14), P2-1(발급일 ≠ 커밋일 거부, M7),
  P2-2(spec ① RELEASED 한정), P2-3(engine-safety MODIFIED), P2-4(`LimitsDigest` 결속, M15), P2-7(콘솔 칸 — 2차 사용자 결정으로 a136 안),
  P2-8(UNION 두 갈래), P2-9(`measure.gridPolicy`), P2-10(proposal 갱신), P3(기본값 대조 시험 · 인자 수 정리 · 빠진 시나리오 · FAILED_CONFIRMED 포함 사유 ·
  UNKNOWN_BROKER_STATE 매일 셈을 「거절되는 정상 입력」에).
- 잔재(기록만): P2-5(a137 과 일 키 적용 시각 — 같은 함수, 적용 시각은 각 change 정의), P2-6(계산 불가 거부의 관측 sub-reason 없음), P3 「오늘 합계/한도」 읽기 전용 표시.
- **P1-3 단계 순서 위반 — 사실로 인정.** `reserveRows` · `DefaultPolicy` · `Policy.Validate` · `mergeEngine` 은 FLM 전에 편집됐다. 수리: base 트리(f504cc52
  `git archive`)에서 편집 전 AST(`ast-base.json`)를 뽑아 번들을 만들고, 그 뒤의 편집(설정 쓰기 · 인터록 · 발급 문 셋 · digest · gridPolicy)은 번들 뒤에 했다.
  편집 전 AST 와 편집 뒤 AST 의 분기 차이는 각 번들 머리에 적었다. **Manager 지시(2026-10-14): 게이트 전 독립 리뷰가 이 4 함수를 base 리비전 FLM/BTM
  (`ast-base.json`) 기준으로 다시 대조한다.**

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a136-daily-buy-notional-limit / 1.1~1.4
- 대상 심볼(편집하는 기존 함수):
    internal/journal/reservations.go       reserveRows                 (판정 호출 한 줄)
    internal/execgw/riskguardian.go        RiskGuardian.IssueEntry      (collectIssue 클로저에 한도 한 칸)
    internal/execgw/riskguardian_qfinal.go RiskGuardian.IssuePrecheckedQFinalEntry (같음)
    internal/execgw/riskguardian_first_leg.go RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg (같음)
    internal/risk/input.go                 DefaultPolicy · Policy.Validate
    internal/app/engine/guardian_wiring.go riskPolicyFromAutomationGate
    internal/app/engine/interlock.go       newAutomationStatus · verifyGate · acceptanceDetail · recordGateSettings
    internal/config/engine.go              mergeEngine
  새 파일(새 leaf): internal/clock/account_day.go · internal/journal/daily_buy.go · internal/config/daily_buy.go ·
                   internal/execgw/daily_buy.go · internal/app/engine/daily_buy_interlock.go
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (reserveRows 호출자 5 = 모든 예약 tx)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/ (불일치 0)
- 변이 원장: `analysis/mutation/round1.log`(18 중 3 이 컴파일 실패를 CAUGHT 로 오분류 · 2 SURVIVED) → 하네스 · 시험 수리 →
  `round2.log` **18/18 CAUGHT**(무변이 대조군 GREEN 선행).
- 기존 동작 파악 근거: HEAD 읽기 + analysis/ast-evidence/ + 기존 시험(reservations_test · riskguardian_test · interlock 시험 · guardian_wiring_internal_test)
- Function Logic Map / Branch Test Map: analysis/function-logic/<pkg>--<func>/ (편집 뒤 재추출)
- upstream 상속 테스트 영향: no — 게이트 OFF 는 Guardian 을 만들지 않음(engine.go:536-549), 매도 경로 무편집
- 실패 테스트 선행 작성: yes (analysis/red/)
- 설정·DB·journal 변경과 rollback: 설정 키 1(선택, 없음 = 기본) · 원장 스키마 변경 0 · rollback = 이전 이미지(키는 옛 빌드가 무시)
- 안전 불변식 §0 위반 여부 검토: 통과 — 매도 · 손절 · 청산은 reserveRows 를 지나지 않음(RISK_REDUCING 은 예약 없음), 판정은 거부 방향만 추가
```
