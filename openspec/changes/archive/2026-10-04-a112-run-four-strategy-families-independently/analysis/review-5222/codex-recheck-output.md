BLOCK — #1은 종결됐지만, #2의 계좌 무결성 오류 분류가 남았습니다. #3의 핵심 논증은 정정됐습니다.

읽기 전용 코드 판정입니다. 사본 생성이 `Read-only file system`으로 실패하여 시험·변이를 재실행하지 못했습니다. 제출된 검증 로그는 확인했으며, 생성된 사본·삭제 대상은 없습니다. 금지 경로는 읽지 않았습니다.

| # | 등급 | 발견 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | `accountScopeCause`가 ctx 이외의 모든 오류를 범위 국소로 분류합니다. **계좌 매니페스트 digest·서명·계좌 결속 검증 실패도 포함**됩니다. 따라서 무결성 오류 후 같은 주기의 다른 범위 발급 가능성이 남습니다. | [계좌 생산 적재:134–140](/tmp/claude-1000/a112-review-face8d0d/internal/strategyaccount/production.go:134)은 파일 부재와 무결성 실패를 같은 오류로 반환합니다. [범위 분류:54–65](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_owner_scope_authority.go:54)는 ctx가 아니면 모두 국소로 판단하고, [마지막 권한:331–337](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_account_first_leg_authority.go:331)이 범위 거절을 만듭니다. | 계좌 적재 원점에서 허용된 범위 국소 결손과 무결성·예상 밖 I/O 오류를 구분하십시오. 국소 원인만 명시적으로 허용하고 종단 시험을 추가하십시오. |
| 2 | **T** | CAS 전수표는 정정됐지만, 일부 주석에는 같은 파도 둘째 범위가 반드시 stale로 거절되는 듯한 설명이 남았습니다. BLOCK 사유는 아닙니다. | 정정된 [review:5681](/tmp/claude-1000/a112-review-face8d0d/openspec/changes/a112-run-four-strategy-families-independently/review.md:5681)과 달리 [시험 주석:306–310](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/a112_owner_scope_trading_test.go:306)은 일반적인 설계 설명으로 남아 있습니다. | “이 fixture에서는”이라는 전제를 맞춰 적으십시오. |

P1의 구체 경로는 **코드상 도출이며 미실행**입니다.

첫 범위 계좌 적재에서 파일 digest 불일치가 발생하고, 다음 범위 적재 전에 정상 파일로 복구되면 둘째 범위만 준비될 수 있습니다. 시장은 준비된 범위 하나로 Ready가 됩니다([시장 집계:203–222](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_account_first_leg_authority.go:203)). 첫 오류는 범위 거절로 바뀌고, `admit`과 `dispatch`가 이를 보존하여 전달 몸통이 다음 범위로 진행합니다([admit:80–84](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_first_leg_admission.go:80), [dispatch:151–156](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_dispatch_cycle.go:151), [전달:69–72](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_market_handoff_delivery.go:69)). 잘못된 계좌 권한 자체가 발급되는 문제는 아니지만, 요청한 **무결성 오류 뒤 같은 주기 발급 중단**은 충족하지 못합니다.

항목별 판정:

1. **OFF 수용 집합 — 기존 발견 종결.** `!Verified && len != 1` 관문이 복원됐고([권한:311–313](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/strategy_account_first_leg_authority.go:311)), 비활성 두 항목 입력을 직접 거절하는 [시험:350–363](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:350)이 추가됐습니다. 검토한 생산 조립 경로에서 새로 통과하는 입력은 찾지 못했습니다.
2. **오류 분류 — 위험 쪽 종결, 전체는 미종결.** 위험 sentinel 생성은 정책 밖 종목과 실제 scope latch 두 곳뿐입니다([302](/tmp/claude-1000/a112-review-face8d0d/internal/riskbucket/production_snapshot_authority.go:302), [380–385](/tmp/claude-1000/a112-review-face8d0d/internal/riskbucket/production_snapshot_authority.go:380)). 조회 결함에는 sentinel이 붙지 않고 원인이 운반됩니다. 정책 밖 종목·latch·계좌 부재의 역방향 오분류는 확인하지 못했으나, 계좌 무결성 실패는 위 P1이 남습니다. **riskbucket 자체의 수락·거절 집합은 변경되지 않았습니다**—오류 신원 보존과 기존 거절 조건 분리입니다.
3. **CAS 문구 — 핵심 정정 적합.** 전수표가 사용량 하한 비교와 중간 해제 가능성을 명시합니다. 노출 시험도 이제 첫 파도에서 `OPEN_EXPOSURE`와 `already held 801`을 함께 단언합니다([시험:367–377](/tmp/claude-1000/a112-review-face8d0d/internal/app/engine/a112_owner_scope_trading_test.go:367)). 남은 주석 정리는 T입니다.

Recommendation: BLOCK — 계좌 적재기의 무결성 오류를 범위 국소 거절에서 제외하고, 첫 범위 무결성 실패·둘째 범위 정상 조건의 종단 시험으로 #2를 닫으십시오.